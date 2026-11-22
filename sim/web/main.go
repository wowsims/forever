package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"runtime/debug"
	"runtime/pprof"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/browser"
	dist "github.com/wowsims/forever/binary_dist"
	"github.com/wowsims/forever/sim"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/bulk"
	proto "github.com/wowsims/forever/sim/core/proto"
	reforgeoptimizer "github.com/wowsims/forever/sim/core/reforge_optimizer"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/selfupdate"
	"github.com/wowsims/forever/sim/siml"

	googleProto "google.golang.org/protobuf/proto"
)

func init() {
	sim.RegisterAll()
}

var (
	Version string
	updater *selfupdate.Updater // nil when self-update cannot be set up.
)

const versionRecheckInterval = 6 * time.Hour

// What the last successful check found, as /version reports it.
var (
	versionMu sync.Mutex
	outdated  int                 // 0 unknown, 1 current, 2 a newer release exists.
	latest    *selfupdate.Release // The newer release, nil when there is none.
	canUpdate bool                // A newer release exists and the executable's folder is writable.
)

func newUpdater() (*selfupdate.Updater, error) {
	executable, err := os.Executable()
	if err != nil {
		return nil, err
	}
	source, err := selfupdate.GitHubSource()
	if err != nil {
		return nil, err
	}
	return selfupdate.New(selfupdate.Config{Name: "wowsimforever", Version: Version, Executable: executable, Source: source})
}

// Checks for a newer release and records the result; a failed check keeps the previous one.
func checkVersion(timeout time.Duration) *selfupdate.Release {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	rel, err := updater.Check(ctx)
	if err != nil {
		return nil
	}
	versionMu.Lock()
	defer versionMu.Unlock()
	latest = rel
	if rel == nil {
		outdated, canUpdate = 1, false
	} else {
		outdated, canUpdate = 2, updater.CanWrite()
	}
	return rel
}

// Before the server listens and the browser opens, updates to a newer release and restarts with the same flags. A
// failed update leaves this version running with a notice.
func updateOnStartup() {
	rel := checkVersion(10 * time.Second)
	if rel == nil {
		return
	}
	notice := fmt.Sprintf("New version of simulator available: %s\n\tDownload at: %s\n", rel.Tag, rel.URL)
	if selfupdate.Restarted() {
		fmt.Print(notice)
		return
	}
	fmt.Printf("Updating to %s...\n", rel.Tag)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := updater.Apply(ctx, rel); err != nil {
		fmt.Printf("Could not update: %s\n", err)
		fmt.Print(notice)
		return
	}
	fmt.Printf("Updated to %s, restarting.\n", rel.Tag)
	if err := updater.Restart(os.Args[1:]); err != nil {
		fmt.Printf("Could not restart: %s\n\tStart the simulator again to run %s.\n", err, rel.Tag)
	}
}

func main() {
	if Version == "" {
		Version = "development"
	}
	var useFS = flag.Bool("usefs", false, "Use local file system for client files. Set to true during development.")
	var wasm = flag.Bool("wasm", false, "Use wasm for sim instead of web server apis. Can only be used with usefs=true")
	var simName = flag.String("sim", "", "Name of simulator to launch (ex: balance_druid, elemental_shaman, etc)")
	var host = flag.String("host", "localhost:3333", "URL to host the interface on.")
	var launch = flag.Bool("launch", true, "auto launch browser")
	var skipVersionCheck = flag.Bool("nvc", false, "set true to skip version check")

	flag.Parse()

	fmt.Printf("Version: %s\n", Version)
	if u, err := newUpdater(); err != nil {
		log.Printf("Self-update is unavailable: %s", err)
	} else {
		updater = u
		go updater.RemoveOld()
		if !*skipVersionCheck && Version != "development" {
			updateOnStartup()
			go func() {
				for range time.Tick(versionRecheckInterval) {
					checkVersion(time.Minute)
				}
			}()
		}
	}

	// Warm up the reforge optimizer's HiGHS WASM runtime in the background so the one-time
	// module compile happens at startup instead of stalling the first /reforgeOptimizeAsync request.
	go func() {
		if err := reforgeoptimizer.WarmUp(); err != nil {
			fmt.Printf("reforge optimizer warm-up failed: %v\n", err)
		}
	}()

	s := &server{
		progMut:         sync.RWMutex{},
		asyncProgresses: map[string]*asyncProgress{},
	}
	s.runServer(*useFS, *host, *launch, *simName, *wasm, bufio.NewReader(os.Stdin))
}

// Handlers to decode and handle each proto function
var handlers = map[string]apiHandler{
	"/raidSim": {msg: func() googleProto.Message { return &proto.RaidSimRequest{} }, handle: func(msg googleProto.Message) googleProto.Message {
		return core.RunRaidSim(msg.(*proto.RaidSimRequest))
	}},
	"/statWeights": {msg: func() googleProto.Message { return &proto.StatWeightsRequest{} }, handle: func(msg googleProto.Message) googleProto.Message {
		return core.StatWeights(msg.(*proto.StatWeightsRequest))
	}},
	"/statWeightRequests": {msg: func() googleProto.Message { return &proto.StatWeightsRequest{} }, handle: func(msg googleProto.Message) googleProto.Message {
		return core.StatWeightRequests(msg.(*proto.StatWeightsRequest))
	}},
	"/statWeightCompute": {msg: func() googleProto.Message { return &proto.StatWeightsCalcRequest{} }, handle: func(msg googleProto.Message) googleProto.Message {
		return core.StatWeightCompute(msg.(*proto.StatWeightsCalcRequest))
	}},
	"/statWeightTopUpPlan": {msg: func() googleProto.Message { return &proto.StatWeightsTopUpRequest{} }, handle: func(msg googleProto.Message) googleProto.Message {
		return core.StatWeightTopUpPlan(msg.(*proto.StatWeightsTopUpRequest))
	}},
	"/computeStats": {msg: func() googleProto.Message { return &proto.ComputeStatsRequest{} }, handle: func(msg googleProto.Message) googleProto.Message {
		return core.ComputeStats(msg.(*proto.ComputeStatsRequest))
	}},
	"/bulkCombinationCount": {msg: func() googleProto.Message { return &proto.BulkCombinationCountRequest{} }, handle: func(msg googleProto.Message) googleProto.Message {
		return bulk.BulkCombinationCount(msg.(*proto.BulkCombinationCountRequest))
	}},
	"/bulkCandidates": {msg: func() googleProto.Message { return &proto.BulkCandidatesRequest{} }, handle: func(msg googleProto.Message) googleProto.Message {
		return bulk.BulkCandidates(msg.(*proto.BulkCandidatesRequest))
	}},
	"/abortById": {msg: func() googleProto.Message { return &proto.AbortRequest{} }, handle: func(msg googleProto.Message) googleProto.Message {
		requestId := msg.(*proto.AbortRequest).RequestId
		triggered := simsignals.AbortById(requestId)
		return &proto.AbortResponse{RequestId: requestId, WasTriggered: triggered}
	}},
	"/simlParse": simlHandler(func() *proto.SimlParseRequest { return &proto.SimlParseRequest{} }, siml.ParseRequest,
		func(err string) *proto.SimlParseResult { return &proto.SimlParseResult{Error: err} }),
	"/simlPrint": simlHandler(func() *proto.SimlPrintRequest { return &proto.SimlPrintRequest{} }, siml.PrintRequest,
		func(err string) *proto.SimlPrintResult { return &proto.SimlPrintResult{Error: err} }),
}

// The site the server serves, which holds the name table SIML requests read with.
var siteFS fs.FS

var (
	simlNamesMu    sync.Mutex
	simlNamesTable *siml.Names
)

// sim/siml/names.json, as the site serves it, read on the first SIML request that finds it: the server is built
// without the with_db tag, so it holds no table of its own.
func simlNames() (*siml.Names, error) {
	simlNamesMu.Lock()
	defer simlNamesMu.Unlock()
	if simlNamesTable != nil {
		return simlNamesTable, nil
	}
	data, err := fs.ReadFile(siteFS, "forever/siml_names.json")
	if err != nil {
		return nil, err
	}
	names, err := siml.ParseNames(data)
	if err == nil {
		simlNamesTable = names
	}
	return names, err
}

// The handler of a SIML request. A name table that does not read answers with the reason, from failed.
func simlHandler[Req, Res googleProto.Message](newReq func() Req, handle func(Req, *siml.Names) Res, failed func(string) Res) apiHandler {
	return apiHandler{msg: func() googleProto.Message { return newReq() }, handle: func(msg googleProto.Message) googleProto.Message {
		names, err := simlNames()
		if err != nil {
			return failed(fmt.Sprintf("the SIML name table does not read: %s", err))
		}
		return handle(msg.(Req), names)
	}}
}

var asyncAPIHandlers = map[string]asyncAPIHandler{
	"/raidSimAsync": {msg: func() googleProto.Message { return &proto.RaidSimRequest{} }, handle: func(msg googleProto.Message, reporter chan *proto.ProgressMetrics, requestId string) {
		core.RunRaidSimConcurrentAsync(msg.(*proto.RaidSimRequest), reporter, requestId)
	}, errorProgress: func(errorOutcome *proto.ErrorOutcome) *proto.ProgressMetrics {
		return &proto.ProgressMetrics{FinalRaidResult: &proto.RaidSimResult{Error: errorOutcome}}
	}},
	"/statWeightsAsync": {msg: func() googleProto.Message { return &proto.StatWeightsRequest{} }, handle: func(msg googleProto.Message, reporter chan *proto.ProgressMetrics, requestId string) {
		core.StatWeightsAsync(msg.(*proto.StatWeightsRequest), reporter, requestId)
	}, errorProgress: func(errorOutcome *proto.ErrorOutcome) *proto.ProgressMetrics {
		return &proto.ProgressMetrics{FinalWeightResult: &proto.StatWeightsResult{Error: errorOutcome}}
	}},
	"/bulkSimAsync": {msg: func() googleProto.Message { return &proto.BulkSimRequest{} }, handle: func(msg googleProto.Message, reporter chan *proto.ProgressMetrics, requestId string) {
		BulkSimAsync(msg.(*proto.BulkSimRequest), reporter, requestId)
	}, errorProgress: func(errorOutcome *proto.ErrorOutcome) *proto.ProgressMetrics {
		return &proto.ProgressMetrics{FinalBulkSimResult: &proto.BulkSimResult{Error: errorOutcome}}
	}},
	"/reforgeOptimizeAsync": {msg: func() googleProto.Message { return &proto.ReforgeOptimizeRequest{} }, handle: func(msg googleProto.Message, reporter chan *proto.ProgressMetrics, requestId string) {
		ReforgeOptimizeAsync(msg.(*proto.ReforgeOptimizeRequest), reporter, requestId)
	}, errorProgress: func(errorOutcome *proto.ErrorOutcome) *proto.ProgressMetrics {
		return &proto.ProgressMetrics{FinalReforgeResult: &proto.ReforgeOptimizeResult{Error: errorOutcome}}
	}},
}

const asyncProgressSilenceTimeout = 10 * time.Minute

func ReforgeOptimizeAsync(request *proto.ReforgeOptimizeRequest, progress chan *proto.ProgressMetrics, requestId string) {
	if requestId == "" {
		requestId = request.GetRequestId()
	}
	signals, err := simsignals.RegisterWithId(requestId)
	if err != nil {
		progress <- &proto.ProgressMetrics{
			FinalReforgeResult: &proto.ReforgeOptimizeResult{
				Error: &proto.ErrorOutcome{Message: "Couldn't register for signal API: " + err.Error()},
			},
		}
		close(progress)
		return
	}

	go func() {
		defer simsignals.UnregisterId(requestId)
		defer close(progress)
		defer func() {
			if err := recover(); err != nil {
				errStr := fmt.Sprint(err) + "\nStack Trace:\n" + string(debug.Stack())
				log.Printf("[ERROR] Reforge optimization panicked: %s", errStr)
				progress <- &proto.ProgressMetrics{
					FinalReforgeResult: &proto.ReforgeOptimizeResult{Error: &proto.ErrorOutcome{Message: errStr}},
				}
			}
		}()
		progress <- &proto.ProgressMetrics{
			FinalReforgeResult: reforgeoptimizer.OptimizeAsync(request, signals),
		}
	}()
}

type server struct {
	progMut         sync.RWMutex
	asyncProgresses map[string]*asyncProgress
}

type apiHandler struct {
	msg    func() googleProto.Message
	handle func(googleProto.Message) googleProto.Message
}
type asyncAPIHandler struct {
	msg    func() googleProto.Message
	handle func(googleProto.Message, chan *proto.ProgressMetrics, string)
	// errorProgress wraps a request-level failure (e.g. a parse error) in the endpoint's own
	// final result type, so the client's consumer for this endpoint sees the error.
	errorProgress func(*proto.ErrorOutcome) *proto.ProgressMetrics
}

func (s *server) handleAsyncAPI(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}
	endpoint := r.URL.Path
	handler, ok := asyncAPIHandlers[endpoint]
	if !ok {
		log.Printf("Invalid Endpoint: %s", endpoint)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	msg := handler.msg()
	if err := googleProto.Unmarshal(body, msg); err != nil {
		message := fmt.Sprintf("Failed to parse %s request: %s (%s)", endpoint, err.Error(), protoParsePreview(body))
		log.Print(message)
		simProgress := s.addNewSim()
		simProgress.latestProgress.Store(handler.errorProgress(&proto.ErrorOutcome{Message: message}))
		s.writeAsyncAPIResult(w, simProgress.id)
		return
	}

	// reporter channel is handed into the core simulation.
	//  as the simulation advances it will push changes to the channel
	//  these changes will be consumed by the goroutine below so the asyncProgress endpoint can fetch the results.
	reporter := make(chan *proto.ProgressMetrics, 100)
	requestId := r.URL.Query().Get("requestId")
	handler.handle(msg, reporter, requestId)

	// Generate a new async simulation
	simProgress := s.addNewSim()

	// Now launch a background process that pulls progress reports off the reporter channel
	// and pushes it into the async progress cache.
	go func() {
		for {
			select {
			case <-time.After(asyncProgressSilenceTimeout):
				// Nothing drains reporter after this, so the pipeline behind it would block
				// on a full buffer for the process lifetime.
				if requestId != "" {
					simsignals.AbortById(requestId)
				}
				s.progMut.Lock()
				delete(s.asyncProgresses, simProgress.id)
				s.progMut.Unlock()
				return
			case progMetric := <-reporter:
				if progMetric == nil {
					return
				}
				simProgress.latestProgress.Store(progMetric)
				if isFinalProgress(progMetric) {
					return
				}
			}
		}
	}()

	s.writeAsyncAPIResult(w, simProgress.id)
}

// Whether progress carries any endpoint's final result, i.e. the async run is finished. Keep in
// sync with the wasm copy in sim/wasm/main.go.
func isFinalProgress(progress *proto.ProgressMetrics) bool {
	return progress.FinalRaidResult != nil || progress.FinalWeightResult != nil || progress.FinalBulkSimResult != nil || progress.FinalReforgeResult != nil
}

func (s *server) writeAsyncAPIResult(w http.ResponseWriter, progressId string) {
	protoResult := &proto.AsyncAPIResult{ProgressId: progressId}
	outbytes, err := googleProto.Marshal(protoResult)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal result: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/x-protobuf")
	w.Write(outbytes)
}

func (s *server) setupAsyncServer() {
	// All async handlers here will call the addNewSim, generating a new UUID and cached progress state.
	for route := range asyncAPIHandlers {
		http.Handle(route, corsMiddleware(http.HandlerFunc(s.handleAsyncAPI)))
	}

	// asyncProgress will fetch the current progress of a simulation by its UUID.
	http.Handle("/asyncProgress", corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return
		}
		msg := &proto.AsyncAPIResult{}
		if err := googleProto.Unmarshal(body, msg); err != nil {
			log.Printf("Failed to parse /asyncProgress request: %s (%s)", err.Error(), protoParsePreview(body))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Read lock the map of all progress statuses, fetching current one.
		s.progMut.RLock()
		progress, ok := s.asyncProgresses[msg.ProgressId]
		s.progMut.RUnlock()
		if !ok {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		latest := progress.latestProgress.Load().(*proto.ProgressMetrics)
		outbytes, err := googleProto.Marshal(latest)
		if err != nil {
			log.Printf("[ERROR] Failed to marshal result: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/x-protobuf")
		if _, err := w.Write(outbytes); err != nil {
			// Nothing reached the client, so leave the cached progress in place for the next
			// poll rather than dropping it.
			log.Printf("[ERROR] Failed to write async progress response: %s", err.Error())
			return
		}

		// If this was the last result, delete the cache for this simulation.
		if isFinalProgress(latest) {
			s.progMut.Lock()
			delete(s.asyncProgresses, msg.ProgressId)
			s.progMut.Unlock()
		}
	})))
}

func protoParsePreview(body []byte) string {
	previewLen := min(len(body), 16)
	return fmt.Sprintf("len=%d first_bytes=% x ascii=%q", len(body), body[:previewLen], body[:previewLen])
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// Held from the start of an update until the restart, so a second request cannot run one alongside it.
var updating sync.Mutex

// Installs the newest release and restarts on it, answering 202 with the new version before the restart.
func handleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSONError(w, http.StatusMethodNotAllowed, "use POST")
		return
	}
	// The API answers any origin, but another site must not be able to restart the simulator.
	if origin := r.Header.Get("Origin"); origin != "" {
		if u, err := url.Parse(origin); err != nil || u.Host != r.Host {
			writeJSONError(w, http.StatusForbidden, "updates are only started from the simulator's own page")
			return
		}
	}
	if updater == nil {
		writeJSONError(w, http.StatusConflict, "self-update is unavailable")
		return
	}
	if !updating.TryLock() {
		writeJSONError(w, http.StatusConflict, "an update is already running")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()
	rel, err := updater.Check(ctx)
	if errors.Is(err, selfupdate.ErrDevelopmentBuild) || (err == nil && rel == nil) {
		updating.Unlock()
		message := "no newer release"
		if err != nil {
			message = err.Error()
		}
		writeJSONError(w, http.StatusConflict, message)
		return
	}
	if err == nil {
		err = updater.Apply(ctx, rel)
	}
	if err != nil {
		updating.Unlock()
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("Updated to %s, restarting.", rel.Tag)
	writeJSON(w, http.StatusAccepted, map[string]string{"version": rel.Tag})
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	go func() {
		time.Sleep(500 * time.Millisecond)
		if err := updater.Restart(restartArgs(os.Args[1:])); err != nil {
			log.Printf("Could not restart: %s. Start the simulator again to run %s.", err, rel.Tag)
			updating.Unlock()
		}
	}()
}

// The process's flags with --launch=false, so the restart opens no second browser tab. It goes first because flag
// parsing stops at the first argument that is not a flag.
func restartArgs(args []string) []string {
	kept := []string{"--launch=false"}
	for _, arg := range args {
		name, _, _ := strings.Cut(strings.TrimLeft(arg, "-"), "=")
		if strings.HasPrefix(arg, "-") && name == "launch" {
			continue
		}
		kept = append(kept, arg)
	}
	return kept
}

// Listens on host. A process restarted after an update retries for a few seconds, since on Windows the process it
// replaces may still hold the port.
func listen(host string) (net.Listener, error) {
	listener, err := net.Listen("tcp", host)
	for i := 0; err != nil && selfupdate.Restarted() && i < 20; i++ {
		time.Sleep(250 * time.Millisecond)
		listener, err = net.Listen("tcp", host)
	}
	return listener, err
}

func (s *server) runServer(useFS bool, host string, launchBrowser bool, simName string, wasm bool, inputReader *bufio.Reader) {
	s.setupAsyncServer()

	var fs http.Handler
	if useFS {
		log.Printf("Using local file system for development.")
		fs = http.FileServer(http.Dir("./dist"))
		siteFS = os.DirFS("./dist")
	} else {
		log.Printf("Embedded file server running.")
		fs = http.FileServer(http.FS(dist.FS))
		siteFS = dist.FS
	}

	for route := range handlers {
		http.Handle(route, corsMiddleware(http.HandlerFunc(handleAPI)))
	}

	http.HandleFunc("/version", func(resp http.ResponseWriter, req *http.Request) {
		versionMu.Lock()
		msg := struct {
			Version   string `json:"version"`
			Outdated  int    `json:"outdated"`
			Latest    string `json:"latest"`
			CanUpdate bool   `json:"canUpdate"`
		}{Version: Version, Outdated: outdated, CanUpdate: canUpdate}
		if latest != nil {
			msg.Latest = latest.Tag
		}
		versionMu.Unlock()
		writeJSON(resp, http.StatusOK, msg)
	})
	http.HandleFunc("/update", handleUpdate)
	http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" {
			http.Redirect(resp, req, "/forever/", http.StatusPermanentRedirect)
			return
		}
		resp.Header().Add("Cache-Control", "no-cache")
		if strings.HasSuffix(req.URL.Path, ".wasm") {
			resp.Header().Set("Content-Type", "application/wasm")
		}
		if strings.HasSuffix(req.URL.Path, ".js") {
			resp.Header().Set("Content-Type", "application/javascript")
		}
		if !useFS || (useFS && !wasm) {
			if strings.HasSuffix(req.URL.Path, "sim_worker.js") {
				req.URL.Path = strings.Replace(req.URL.Path, "sim_worker.js", "net_worker.js", 1)
			}
		}
		fs.ServeHTTP(resp, req)
	})

	if launchBrowser {
		if strings.HasPrefix(host, ":") {
			host = "localhost" + host
		}
		url := fmt.Sprintf("http://%s/forever/%s", host, simName)
		log.Printf("Launching interface on %s", url)
		go func() {
			err := browser.OpenURL(url)
			if err != nil {
				fmt.Printf("Error launching browser: %#v\n", err.Error())
				fmt.Printf("You will need to manually open your web browser to %s\n", url)
			}
		}()
	}

	go func() {
		// Launch server!
		listener, err := listen(host)
		if err != nil {
			log.Printf("Failed to listen on %s: %s", host, err)
			os.Exit(1)
		}
		if err := http.Serve(listener, nil); err != nil {
			log.Printf("Failed to shutdown server: %s", err)
			os.Exit(1)
		}
		log.Printf("Server shutdown successfully.")
		os.Exit(0)
	}()

	// used to read a CTRL+C
	c := make(chan os.Signal, 10)
	signal.Notify(c, syscall.SIGINT)

	go func() {
		<-c
		log.Printf("Shutting down")
		os.Exit(0)
	}()
	fmt.Printf("Enter Command... '?' for list\n")
	for {
		fmt.Printf("> ")
		text, err := inputReader.ReadString('\n')
		if err != nil {
			// block forever
			<-c
			os.Exit(-1)
		}
		if len(text) == 0 {
			continue
		}
		command := strings.TrimSpace(text)
		switch command {
		case "profile":
			filename := fmt.Sprintf("profile_%d.cpu", time.Now().Unix())
			fmt.Printf("Running profiling for 15 seconds, output to %s\n", filename)
			f, err := os.Create(filename)
			if err != nil {
				log.Fatal("could not create CPU profile: ", err)
			}
			if err := pprof.StartCPUProfile(f); err != nil {
				log.Fatal("could not start CPU profile: ", err)
			}
			go func() {
				time.Sleep(time.Second * 15)
				pprof.StopCPUProfile()
				f.Close()
				fmt.Printf("Profiling complete.\n> ")
			}()
		case "heap_profile":
			filename := fmt.Sprintf("profile_%d.heap", time.Now().Unix())
			fmt.Printf("Capturing heap snapshot, output to %s\n", filename)
			f, err := os.Create(filename)
			if err != nil {
				log.Fatal("could not create output file: ", err)
			}
			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Fatal("could not capture heap profile: ", err)
			}
			go func() {
				time.Sleep(time.Second)
				f.Close()
				fmt.Printf("Profiling complete.\n> ")
			}()
		case "sims":
			s.progMut.RLock()
			fmt.Printf("Total Sims Running: %d\n", len(s.asyncProgresses))
			for _, v := range s.asyncProgresses {
				latest := (v.latestProgress.Load()).(*proto.ProgressMetrics)
				fmt.Printf("Process: %s (%d sims)\n\t  Progress: %d/%d\n", v.id, latest.TotalSims, latest.CompletedIterations, latest.TotalIterations)
			}
			s.progMut.RUnlock()
		case "quit":
			os.Exit(1)
		case "?":
			fmt.Printf("Commands:\n\tsims - Lists all active async sims running currently.\n\tprofile - start a CPU profile for debugging performance\n\theap_profile - capture a memory snapshot for debugging performance\n\tquit - exits\n\n")
		case "":
			// nothing.
		default:
			fmt.Printf("Unknown command: '%s'", command)
		}
	}
}

// The generic handler for any api function using protos.
func handleAPI(w http.ResponseWriter, r *http.Request) {
	endpoint := r.URL.Path

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}
	handler, ok := handlers[endpoint]
	if !ok {
		log.Printf("Invalid Endpoint: %s", endpoint)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	msg := handler.msg()
	if err := googleProto.Unmarshal(body, msg); err != nil {
		log.Printf("Failed to parse request: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if googleProto.Equal(msg, msg.ProtoReflect().New().Interface()) {
		log.Printf("Request is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result := handler.handle(msg)

	outbytes, err := googleProto.Marshal(result)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal result: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/x-protobuf")
	w.Write(outbytes)
}

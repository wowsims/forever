// Package lsp is the SIML language server: diagnostics, completion, hover, navigation, rename, outline, folding,
// semantic tokens, formatting and code actions for every *.siml file, and the requests the rotation panel of the
// VS Code extension makes (siml/rotation, siml/pathAt, siml/registered, siml/splice, siml/database, siml/patterns, siml/insertPattern, siml/sim).
// It speaks JSON-RPC over a stream, and reads the name table and the item database from the repository each file
// sits in.
package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/wowsims/forever/sim/siml"
	"github.com/wowsims/forever/sim/siml/lsp/jsonrpc"
)

// How long the server waits for typing to pause before it checks a changed file, and before it checks the open files
// that use it.
const (
	changedDelay = 150 * time.Millisecond
	usersDelay   = 400 * time.Millisecond
)

var triggerCharacters = []string{" ", "(", ".", `"`, "$", "%", ",", "=", "[", "+"}

// What the server needs from the program that starts it.
type Options struct {
	// Sims a request of `wowsimcli siml sim` in the repository at root and gives its result; it stops when ctx is
	// done.
	Sim func(ctx context.Context, root string, req json.RawMessage) (json.RawMessage, error)
	// The game files' row of a spell, as `go run ./tools/spelldata <id>` prints it, for a hover.
	Row func(spellID int32) (string, bool)
}

// An open document.
type document struct {
	uri        string
	path       string // "" for an untitled document.
	languageID string
	version    int
	text       string
}

// A document read: the Doc, or why it cannot be read; and the files its use lines read, which a change to makes it
// stale.
type analysis struct {
	version int
	root    string
	names   *siml.Names
	doc     *siml.Doc
	err     error
	used    []string
	stale   bool
	// While set, the files the opener reads are the ones the use lines read; the services read files near it after.
	recording bool
	diags     []Diagnostic // Computed on first need.
}

type server struct {
	opts Options
	w    *jsonrpc.FrameWriter
	ws   *workspaces

	mu        sync.Mutex // Holds everything below and every Doc: one analysis runs at a time.
	docs      map[string]*document
	analyses  map[string]*analysis
	timers    map[string]*time.Timer
	pending   map[string]context.CancelFunc
	folders   []string // The workspace folders, which untitled documents are read in.
	watches   bool
	shutdown  bool
	requestID int

	simSettings simSettings
	simRecords  map[string]simRecord
	simming     bool
	clientCaps  struct{ progress, lensRefresh, hintRefresh bool }

	cards *cards
}

// Serves the protocol on in and out until the client sends exit or in ends. The result says whether shutdown came
// first, as the exit code should.
func Serve(in io.Reader, out io.Writer, opts Options) (shutdownRequested bool, err error) {
	s := &server{
		opts:     opts,
		w:        jsonrpc.NewFrameWriter(out),
		ws:       newWorkspaces(),
		docs:     map[string]*document{},
		analyses: map[string]*analysis{},
		timers:   map[string]*time.Timer{},
		pending:  map[string]context.CancelFunc{},
		cards:    newCards(opts.Row),

		simRecords: map[string]simRecord{},
	}
	defer s.stop()
	reader := jsonrpc.NewFrameReader(in)
	for {
		body, err := reader.Read()
		if errors.Is(err, io.EOF) {
			return s.isShutdown(), nil
		}
		if err != nil {
			return s.isShutdown(), err
		}
		var msg message
		if err := json.Unmarshal(body, &msg); err != nil {
			s.w.Respond(json.RawMessage("null"), nil, &ResponseError{Code: codeParseError, Message: err.Error()})
			continue
		}
		if msg.Method == "exit" {
			return s.isShutdown(), nil
		}
		s.dispatch(msg)
	}
}

func (s *server) isShutdown() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.shutdown
}

// Stops the pending checks and requests.
func (s *server) stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, t := range s.timers {
		t.Stop()
	}
	for _, cancel := range s.pending {
		cancel()
	}
}

func (s *server) dispatch(msg message) {
	if msg.Method == "" {
		return // A response to a request of the server's, such as client/registerCapability.
	}
	if len(msg.ID) == 0 {
		s.guard(msg.Method, func() { s.notification(msg.Method, msg.Params) })
		return
	}
	switch msg.Method {
	case "initialize", "shutdown":
		var result any
		var rpcErr *ResponseError
		s.guard(msg.Method, func() { result, rpcErr = s.lifecycle(msg.Method, msg.Params) })
		s.w.Respond(msg.ID, result, rpcErr)
		return
	}
	// Every other request runs on its own, so a slow one (a hover that reads the game files, a sim) does not hold up
	// the rest, and $/cancelRequest reaches it.
	ctx, cancel := context.WithCancel(context.Background())
	key := string(msg.ID)
	s.mu.Lock()
	s.pending[key] = cancel
	s.mu.Unlock()
	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.pending, key)
			s.mu.Unlock()
			cancel()
		}()
		var result any
		var rpcErr *ResponseError
		ran := s.guard(msg.Method, func() { result, rpcErr = s.request(ctx, msg.Method, msg.Params) })
		if !ran {
			rpcErr = &ResponseError{Code: codeInternalError, Message: msg.Method + " failed"}
		}
		if ctx.Err() != nil && rpcErr == nil {
			rpcErr = &ResponseError{Code: codeRequestCancelled, Message: "the request was cancelled"}
		}
		s.w.Respond(msg.ID, result, rpcErr)
	}()
}

// Runs f, and reports a panic in it to the client's log; false when it panicked.
func (s *server) guard(method string, f func()) (ran bool) {
	defer func() {
		if r := recover(); r != nil {
			s.log(fmt.Sprintf("%s panicked: %v\n%s", method, r, debug.Stack()))
			ran = false
		}
	}()
	f()
	return true
}

func (s *server) log(message string) {
	s.w.Notify("window/logMessage", map[string]any{"type": 4, "message": message})
}

func (s *server) lifecycle(method string, params json.RawMessage) (any, *ResponseError) {
	if method == "shutdown" {
		s.mu.Lock()
		s.shutdown = true
		s.mu.Unlock()
		return nil, nil
	}
	var p struct {
		RootURI          string `json:"rootUri"`
		WorkspaceFolders []struct {
			URI string `json:"uri"`
		} `json:"workspaceFolders"`
		Capabilities struct {
			Workspace struct {
				DidChangeWatchedFiles struct {
					DynamicRegistration bool `json:"dynamicRegistration"`
				} `json:"didChangeWatchedFiles"`
				CodeLens struct {
					RefreshSupport bool `json:"refreshSupport"`
				} `json:"codeLens"`
				InlayHint struct {
					RefreshSupport bool `json:"refreshSupport"`
				} `json:"inlayHint"`
			} `json:"workspace"`
			Window struct {
				WorkDoneProgress bool `json:"workDoneProgress"`
			} `json:"window"`
		} `json:"capabilities"`
		InitializationOptions json.RawMessage `json:"initializationOptions"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, invalidParams(err)
	}
	s.mu.Lock()
	for _, f := range p.WorkspaceFolders {
		if path, ok := uriPath(f.URI); ok {
			s.folders = append(s.folders, path)
		}
	}
	if path, ok := uriPath(p.RootURI); ok && len(s.folders) == 0 {
		s.folders = append(s.folders, path)
	}
	s.watches = p.Capabilities.Workspace.DidChangeWatchedFiles.DynamicRegistration
	s.clientCaps.progress = p.Capabilities.Window.WorkDoneProgress
	s.clientCaps.lensRefresh = p.Capabilities.Workspace.CodeLens.RefreshSupport
	s.clientCaps.hintRefresh = p.Capabilities.Workspace.InlayHint.RefreshSupport
	if len(p.InitializationOptions) > 0 && json.Unmarshal(p.InitializationOptions, &s.simSettings) != nil {
		s.simSettings = simSettings{}
	}
	s.mu.Unlock()
	capabilities := map[string]any{
		"textDocumentSync":       map[string]any{"openClose": true, "change": 1},
		"completionProvider":     map[string]any{"triggerCharacters": triggerCharacters},
		"hoverProvider":          true,
		"definitionProvider":     true,
		"referencesProvider":     true,
		"renameProvider":         map[string]any{"prepareProvider": true},
		"documentSymbolProvider": true,
		"foldingRangeProvider":   true,
		"semanticTokensProvider": map[string]any{
			"legend": map[string]any{"tokenTypes": nonNil(siml.TokenTypes()), "tokenModifiers": nonNil(siml.TokenModifiers())},
			"full":   true,
		},
		"documentFormattingProvider": true,
		"codeActionProvider":         map[string]any{"codeActionKinds": []string{"quickfix", "refactor.rewrite"}},
		"documentLinkProvider":       map[string]any{"resolveProvider": false},
	}
	s.simCapabilities(capabilities)
	return map[string]any{
		"capabilities": capabilities,
		"serverInfo":   map[string]any{"name": "wowsims-siml"},
	}, nil
}

func nonNil[T any](list []T) []T {
	if list == nil {
		return []T{}
	}
	return list
}

// The files the server watches: the name table and the database every check reads, and the .siml files open files
// use.
var watchedGlobs = []string{"**/sim/siml/names.json", "**/assets/database/db.json", "**/assets/database/leftover_db.json", "**/*.siml"}

func (s *server) notification(method string, params json.RawMessage) {
	switch method {
	case "initialized":
		s.mu.Lock()
		watches := s.watches
		s.requestID++
		id := s.requestID
		s.mu.Unlock()
		if !watches {
			return
		}
		watchers := make([]map[string]string, len(watchedGlobs))
		for i, glob := range watchedGlobs {
			watchers[i] = map[string]string{"globPattern": glob}
		}
		s.w.Request(id, "client/registerCapability", map[string]any{"registrations": []any{map[string]any{
			"id":              "siml-watched-files",
			"method":          "workspace/didChangeWatchedFiles",
			"registerOptions": map[string]any{"watchers": watchers},
		}}})
	case "$/cancelRequest":
		var p struct {
			ID json.RawMessage `json:"id"`
		}
		if json.Unmarshal(params, &p) == nil {
			s.mu.Lock()
			if cancel := s.pending[string(p.ID)]; cancel != nil {
				cancel()
			}
			s.mu.Unlock()
		}
	case "workspace/didChangeWatchedFiles":
		var p struct {
			Changes []struct {
				URI string `json:"uri"`
			} `json:"changes"`
		}
		if json.Unmarshal(params, &p) != nil {
			return
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		for _, change := range p.Changes {
			path, ok := uriPath(change.URI)
			if !ok {
				continue
			}
			if strings.HasSuffix(path, ".siml") {
				s.changedLocked(path, "")
				continue
			}
			// The table and the database are read again by the next check, which every open file gets.
			for uri := range s.docs {
				if a := s.analyses[uri]; a != nil {
					a.stale = true
				}
				s.scheduleLocked(uri, usersDelay)
			}
		}
	case "textDocument/didOpen", "textDocument/didChange", "textDocument/didClose":
		var p textDocumentParams
		if json.Unmarshal(params, &p) != nil {
			return
		}
		s.documentChanged(method, &p)
	}
}

// Whether the server reads documents of a URI: files and untitled documents, not a diff view's or another scheme's.
func served(uri string) bool {
	return strings.HasPrefix(uri, "file:") || strings.HasPrefix(uri, "untitled:")
}

func (s *server) documentChanged(method string, p *textDocumentParams) {
	uri := p.TextDocument.URI
	if !served(uri) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	switch method {
	case "textDocument/didOpen":
		path, _ := uriPath(uri)
		d := &document{uri: uri, path: path, languageID: p.TextDocument.LanguageID, version: p.TextDocument.Version, text: p.TextDocument.Text}
		if !strings.HasSuffix(uri, ".siml") && d.languageID != "siml" {
			return
		}
		s.docs[uri] = d
		s.changedLocked(path, uri)
	case "textDocument/didChange":
		d := s.docs[uri]
		if d == nil {
			return
		}
		for _, change := range p.ContentChanges {
			d.text = change.Text
		}
		d.version = p.TextDocument.Version
		s.changedLocked(d.path, uri)
	case "textDocument/didClose":
		d := s.docs[uri]
		if d == nil {
			return
		}
		delete(s.docs, uri)
		delete(s.analyses, uri)
		if t := s.timers[uri]; t != nil {
			t.Stop()
			delete(s.timers, uri)
		}
		s.w.Notify("textDocument/publishDiagnostics", map[string]any{"uri": uri, "diagnostics": []Diagnostic{}})
		// A file closed unsaved is read from disk again by the open files that use it.
		s.changedLocked(d.path, "")
	}
}

// Checks the document at uri soon, and the open files that use the file at path a little later, read again.
func (s *server) changedLocked(path, uri string) {
	if uri != "" {
		s.scheduleLocked(uri, changedDelay)
	}
	if path == "" {
		return
	}
	for other, a := range s.analyses {
		if other != uri && slices.ContainsFunc(a.used, func(used string) bool { return siml.SamePath(used, path) }) {
			a.stale = true
			s.scheduleLocked(other, usersDelay)
		}
	}
}

func (s *server) scheduleLocked(uri string, delay time.Duration) {
	if t := s.timers[uri]; t != nil {
		t.Stop()
	}
	s.timers[uri] = time.AfterFunc(delay, func() { s.guard("diagnostics", func() { s.publish(uri) }) })
}

// Publishes the diagnostics of an open document.
func (s *server) publish(uri string) {
	var catalog catalogRead
	s.mu.Lock()
	if a := s.analyzeLocked(uri); a == nil || a.diags == nil {
		// The diagnostics are not read yet, so the catalog they check against is read first, outside the lock.
		s.mu.Unlock()
		catalog = s.readCatalog(uri)
		s.mu.Lock()
	}
	defer s.mu.Unlock()
	a := s.analyzeLocked(uri)
	if a == nil {
		return
	}
	params := map[string]any{"uri": uri, "version": a.version, "diagnostics": s.diagnosticsLocked(a, catalog.of(a))}
	s.w.Notify("textDocument/publishDiagnostics", params)
}

// The open document at uri read, from the last read while its text, the files it uses and the name table hold. Nil
// when the document is not open.
func (s *server) analyzeLocked(uri string) *analysis {
	d := s.docs[uri]
	if d == nil {
		return nil
	}
	root, ok := s.rootLocked(d.path)
	if !ok {
		return &analysis{version: d.version, err: errors.New("the file is not in a WoWSims repository: no sim/siml/names.json above it")}
	}
	names, err := s.ws.names(root)
	if err != nil {
		return &analysis{version: d.version, root: root, err: err}
	}
	if a := s.analyses[uri]; a != nil && a.version == d.version && a.names == names && !a.stale {
		return a
	}
	a := &analysis{version: d.version, root: root, names: names, recording: true}
	a.doc = siml.Analyze(d.path, []byte(d.text), names, s.openerLocked(a))
	a.recording = false
	s.analyses[uri] = a
	return a
}

// Reads the files a document's use lines name, with the text of an open editor first, and notes each one the
// analysis reads.
func (s *server) openerLocked(a *analysis) siml.Opener {
	return func(path string) ([]byte, error) {
		if a != nil && a.recording && !slices.ContainsFunc(a.used, func(used string) bool { return siml.SamePath(used, path) }) {
			a.used = append(a.used, path)
		}
		if d := s.docByPathLocked(path); d != nil {
			return []byte(d.text), nil
		}
		return siml.ReadSource(path)
	}
}

func (s *server) docByPathLocked(path string) *document {
	if path == "" {
		return nil
	}
	for _, d := range s.docs {
		if d.path != "" && siml.SamePath(d.path, path) {
			return d
		}
	}
	return nil
}

// The repository a file is read in: the one above it, or for an untitled document the first workspace folder that
// is in one.
func (s *server) rootLocked(path string) (string, bool) {
	if path != "" {
		return siml.RepositoryRoot(path)
	}
	for _, folder := range s.folders {
		if root, ok := siml.RepositoryRoot(filepath.Join(folder, "x")); ok {
			return root, true
		}
	}
	return "", false
}

// The diagnostics of a document: the parser's and the checks' against c, which the caller reads before taking the
// lock (readCatalog).
func (s *server) diagnosticsLocked(a *analysis, c *siml.DatabaseCatalog) []Diagnostic {
	if a.diags != nil {
		return a.diags
	}
	if a.err != nil {
		a.diags = []Diagnostic{{Severity: severityError, Source: "siml", Message: a.err.Error()}}
		return a.diags
	}
	a.diags = []Diagnostic{}
	for _, d := range a.doc.EditorDiagnostics(c) {
		a.diags = append(a.diags, s.toDiagnosticLocked(d))
	}
	return a.diags
}

// A diagnostic as the protocol's; a used file's error reported at the use line points at its place in that file.
func (s *server) toDiagnosticLocked(d siml.Diagnostic) Diagnostic {
	severity := severityInformation
	switch d.Severity {
	case siml.Error:
		severity = severityError
	case siml.Warning:
		severity = severityWarning
	}
	out := Diagnostic{Range: toRange(siml.Range{Start: d.Start, End: d.End}), Severity: severity, Code: d.Code, Source: "siml", Message: d.Message}
	if d.Related != nil {
		out.RelatedInformation = []RelatedInformation{{
			Location: Location{URI: s.uriOfLocked(d.Related.Path), Range: toRange(d.Related.Range)},
			Message:  d.Message,
		}}
	}
	return out
}

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/siml"
)

type simlSimRequest struct {
	SIML          string    `json:"siml"`
	Path          string    `json:"path"`
	Compare       *compared `json:"compare,omitempty"`
	SettingsFile  string    `json:"settingsFile,omitempty"`
	Iterations    int32     `json:"iterations,omitempty"`
	Seed          int64     `json:"seed,omitempty"`
	Timeline      bool      `json:"timeline,omitempty"`
	TimelineLimit int       `json:"timelineLimit,omitempty"`
}

type compared struct {
	SIML string `json:"siml"`
}

type distribution struct {
	Mean  float64 `json:"mean"`
	Stdev float64 `json:"stdev"`
}

type rowStats struct {
	Executions float64 `json:"executions"`
	Hidden     bool    `json:"hidden,omitempty"`
}

type spellStats struct {
	ID    json.RawMessage `json:"id"`
	Name  string          `json:"name,omitempty"`
	DPS   float64         `json:"dps"`
	Casts float64         `json:"casts"`
	Pet   string          `json:"pet,omitempty"`
}

type timelineEvent struct {
	TimeMs int64  `json:"timeMs"`
	Path   string `json:"path"`
	Label  string `json:"label,omitempty"`
	Via    string `json:"via,omitempty"`
}

type simResult struct {
	DPS        distribution        `json:"dps"`
	Iterations int32               `json:"iterations"`
	Duration   float64             `json:"duration"`
	Rows       map[string]rowStats `json:"rows"`
	Spells     []spellStats        `json:"spells"`
	Timeline   []timelineEvent     `json:"timeline,omitempty"`
}

type simlSimResult struct {
	*simResult
	Character *characterInfo `json:"character,omitempty"`
	Compare   *simResult     `json:"compare,omitempty"`
	Error     string         `json:"error,omitempty"`
}

// Sims one spec's character with a given APL, for the APL editor: DPS, how often each APL row ran, the
// damage by spell, and the rows one iteration ran in order. Reads one JSON request on stdin and writes one
// JSON result on stdout.
//
//	go run --tags=with_db ./cmd/wowsimcli siml sim < request.json
//
// Request:
//
//	{
//	  "siml": "siml 1\nspec hunter\n...", // the text of an .apl.siml, parsed with the committed name table
//	  "path": "ui/specs/hunter/dps/apls/default.apl.siml",
//	                                     // its file, relative to the repository root or absolute
//	  "compare": {"siml": "..."},        // optional: another text of the file, simmed on the same character and seed
//	  "settingsFile": "path",            // optional: settings the site exports, in protojson, or an x.siml (a full
//	                                     // setup, or a preset build loaded over its spec's defaults.siml), for
//	                                     // everything but the APL
//	  "iterations": 1000,                // default 1000
//	  "seed": 1,                         // default 1, so the same request gives the same result
//	  "timeline": true,                  // optional: the rows one iteration (the first, with `seed`) ran
//	  "timelineLimit": 400               // default 400 events
//	}
//
// The text's spec line names the spec, and its use lines are read against path. Without a settings file the character
// is the spec's defaults.siml with the one of its builds/*.siml loaded over it whose rotation is the saved file's at
// path, as the site loads a preset build; a rotation no build has, or several have, is simmed on the defaults. The
// sim's request is the one the site sends. A request with any other field is refused.
//
// Result:
//
//	{
//	  "dps": {"mean": 812.3, "stdev": 60.1},
//	  "iterations": 1000,
//	  "duration": 180.2,              // average iteration length, in seconds
//	  "rows": {"priorityList/3": {"executions": 12.4}, "groups/1/actions/0": {"executions": 3.1, "hidden": true}},
//	  "spells": [{"id": {"spellId": 2643}, "name": "Multi-Shot", "dps": 120.4, "casts": 14.2, "pet": "Cat"}],
//	  "timeline": [{"timeMs": -1500, "path": "prepullActions/0", "label": "Cast Spell(Hunter's Mark)"},
//	               {"timeMs": 0, "path": "groups/0/actions/1", "label": "...", "via": "priorityList/2"}],
//	  "character": {"source": "ui/specs/hunter/dps/builds/mm.siml over ui/specs/hunter/dps/defaults.siml",
//	                "race": "RaceOrc", "talents": "...", "gear": "p1", "warnings": []},
//	  "compare": {...},               // the same fields for compare, without character
//	  "error": "..."                  // set instead of the results when the sim cannot run; a text that does not
//	                                  // parse gives its diagnostics, one "path:line:col code message" per line
//	}
//
// Rows are keyed by their JSON paths in the APLRotation the text parses to, the rows of the files it uses after its
// own. A row's executions are the times per iteration the rotation ran it; a cast that then fails still counts.
// Every row of the APL is listed, so a row that never ran has 0. A group reference row counts each time it
// runs a row of its group, and that group row counts too; a group the APL references twice sums both uses
// under the group's rows. A sequence row counts when the rotation runs it, not for each spell of the
// sequence. Hidden rows are listed with "hidden" and never run.
func simlSim(out, errOut io.Writer, args []string) error {
	if err := parseFlags(newFlags("sim < request.json", errOut), args, 0); err != nil {
		return err
	}
	var res simlSimResult
	if req, err := readRequest(os.Stdin); err != nil {
		res.Error = fmt.Sprintf("reading the request: %v", err)
	} else {
		res = runSimlSim(context.Background(), ".", req)
	}
	return json.NewEncoder(out).Encode(res)
}

// The request in r, which holds no field the request does not have.
func readRequest(r io.Reader) (simlSimRequest, error) {
	var req simlSimRequest
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&req)
	return req, err
}

// The sim of req in the repository at root; it stops with an error when ctx is done.
func runSimlSim(ctx context.Context, root string, req simlSimRequest) simlSimResult {
	if req.Path == "" {
		return simlSimResult{Error: "the request needs the siml text of an .apl.siml and its path"}
	}
	if req.Iterations <= 0 {
		req.Iterations = 1000
	}
	if req.Seed == 0 {
		req.Seed = 1
	}
	if req.TimelineLimit <= 0 {
		req.TimelineLimit = 400
	}
	names, err := siml.DefaultNames()
	if err != nil {
		return simlSimResult{Error: err.Error()}
	}
	rotation, spec, err := readRotation(root, req.Path, req.SIML, names)
	if err != nil {
		return simlSimResult{Error: err.Error()}
	}
	var compare *proto.APLRotation
	if req.Compare != nil {
		var compareSpec proto.Spec
		compare, compareSpec, err = readRotation(root, req.Path, req.Compare.SIML, names)
		if err == nil && compareSpec != spec {
			err = fmt.Errorf("its spec is %s, and the rotation's %s", compareSpec, spec)
		}
		if err != nil {
			return simlSimResult{Error: "compare: " + err.Error()}
		}
	}

	// A settings file supplies the whole character, so a spec with no defaults.siml sims with one too; the spec's
	// field then names it.
	folder, defaults, err := specDefaults(root, spec)
	var char *character
	switch {
	case req.SettingsFile != "":
		if err != nil {
			info, _ := siml.SpecOf(spec)
			folder = string(info.Field.Name())
		}
		char, err = characterFromSettings(root, req.SettingsFile, spec, folder)
	case err == nil:
		char, err = characterFromRepo(root, folder, defaults, absolute(root, req.Path))
	}
	if err != nil {
		return simlSimResult{Error: err.Error()}
	}
	char.info.Race = char.settings.GetPlayer().GetRace().String()
	char.info.Talents = char.settings.GetPlayer().GetTalentsString()
	base, warnings := core.IndividualSimRequest(char.settings, &proto.SimOptions{RandomSeed: req.Seed})
	char.info.Warnings = append(char.info.Warnings, warnings...)

	res := simlSimResult{Character: &char.info}
	res.simResult, err = simAPL(ctx, base, folder, rotation, req)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	if compare != nil {
		res.Compare, err = simAPL(ctx, base, folder, compare, req)
		if err != nil {
			res.Error = "compare: " + err.Error()
		}
	}
	return res
}

// path, read against root when it is relative.
func absolute(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}

// The rotation that text, the text of the .apl.siml at path, parses to, with the files it uses read against path,
// and its spec. A relative path is read against root.
func readRotation(root, path, text string, names *siml.Names) (*proto.APLRotation, proto.Spec, error) {
	if kind, err := siml.KindOf(path); err != nil || kind.Section != "apl" {
		return nil, 0, fmt.Errorf("%s is not an .apl.siml", path)
	}
	f, err := siml.Parse(absolute(root, path), []byte(text), names, siml.ReadSource)
	if err != nil {
		return nil, 0, err
	}
	return f.Rotation(), f.Spec, nil
}

// Rows the executions are counted for, each with the hidden flag of its list item.
func listRows(rotation *proto.APLRotation) map[string]rowStats {
	rows := map[string]rowStats{}
	for i, item := range rotation.PrepullActions {
		rows[core.PrepullRowPath(i)] = rowStats{Hidden: item.Hide}
	}
	for i, item := range rotation.PriorityList {
		rows[core.PriorityRowPath(i)] = rowStats{Hidden: item.Hide}
	}
	for g, group := range rotation.Groups {
		for i, item := range group.Actions {
			rows[core.GroupRowPath(g, i)] = rowStats{Hidden: item.Hide}
		}
	}
	return rows
}

// Counts the player's executed rows, and while record is set writes the first iteration's timeline. The rows and
// their paths come from the player's rotation on the first action the sim executes, and the player's next reset,
// the start of the second iteration, ends the recording.
type recorder struct {
	counts   map[string]int
	timeline []timelineEvent
	limit    int
	record   bool
	rows     map[*core.APLAction]string // Nil until the first action.
}

func (r *recorder) executed(sim *core.Simulation, action *core.APLAction) {
	if r.rows == nil {
		r.rows = map[*core.APLAction]string{}
		for _, unit := range sim.Raid.AllUnits {
			if unit.Type == core.PlayerUnit && unit.Rotation != nil {
				maps.Copy(r.rows, unit.Rotation.RowPaths())
				unit.RegisterResetEffect(func(*core.Simulation) { r.record = false })
			}
		}
	}
	path, ok := r.rows[action]
	if !ok {
		return
	}
	r.counts[path]++
	if !r.record || len(r.timeline) >= r.limit {
		return
	}
	event := timelineEvent{TimeMs: sim.CurrentTime.Milliseconds(), Path: path, Label: label(action)}
	// A group's row runs only from a reference, which is recorded just before it.
	if n := len(r.timeline); n > 0 && strings.HasPrefix(path, "groups/") {
		event.Via = r.timeline[n-1].Path
	}
	r.timeline = append(r.timeline, event)
}

var spellIDPattern = regexp.MustCompile(`\{SpellID: (\d+)[^}]*\}`)

// The action of a row as the sim prints it, with spell ids replaced by their names: "Cast Spell(Multi-Shot)".
func label(action *core.APLAction) string {
	text := action.String()
	if i := strings.LastIndex(text, "ACTION = "); i >= 0 {
		text = text[i+len("ACTION = "):]
	}
	return spellIDPattern.ReplaceAllStringFunc(text, func(m string) string {
		id, _ := strconv.Atoi(spellIDPattern.FindStringSubmatch(m)[1])
		if spell := spelldata.Find(int32(id)); spell != nil && spell.Name != "" {
			return spell.Name
		}
		return m
	})
}

// Sims rotation as the APL of the one player of base, an individual sim request for the spec folder under ui/specs.
func simAPL(ctx context.Context, base *proto.RaidSimRequest, folder string, rotation *proto.APLRotation, req simlSimRequest) (*simResult, error) {
	rotation = cloneProto(rotation)
	rotation.Type = proto.APLRotation_TypeAPL

	makeRequest := func(iterations int32) *proto.RaidSimRequest {
		request := cloneProto(base)
		request.Raid.Parties[0].Players[0].Rotation = rotation
		request.SimOptions.Iterations = iterations
		return request
	}

	rec := &recorder{counts: map[string]int{}, timeline: []timelineEvent{}, limit: req.TimelineLimit, record: req.Timeline}
	simRes := runRaidSim(ctx, makeRequest(req.Iterations), rec.executed)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if simRes.Error != nil {
		return nil, fmt.Errorf("the sim cannot run %s: %s", folder, simError(simRes.Error))
	}
	out := &simResult{
		DPS:        distribution{Mean: simRes.RaidMetrics.Dps.Avg, Stdev: simRes.RaidMetrics.Dps.Stdev},
		Iterations: simRes.IterationsDone,
		Duration:   simRes.AvgIterationDuration,
		Rows:       listRows(rotation),
		Spells:     spells(simRes),
	}
	for path, count := range rec.counts {
		row := out.Rows[path]
		row.Executions = float64(count) / float64(max(simRes.IterationsDone, 1))
		out.Rows[path] = row
	}

	if req.Timeline {
		out.Timeline = rec.timeline
	}
	return out, nil
}

// The sim of request, aborted when ctx is done.
func runRaidSim(ctx context.Context, request *proto.RaidSimRequest, executed core.APLActionHook) *proto.RaidSimResult {
	signals := simsignals.CreateSignals()
	stop := context.AfterFunc(ctx, signals.Abort.Trigger)
	defer stop()
	return core.RunSimWithAPLHook(request, signals, executed)
}

// The message of a sim error without the stack trace the sim appends to a panic.
func simError(e *proto.ErrorOutcome) string {
	message, _, _ := strings.Cut(e.Message, "\nStack Trace:")
	if message == "" {
		return e.Type.String()
	}
	return message
}

func spells(res *proto.RaidSimResult) []spellStats {
	iterations := float64(max(res.IterationsDone, 1))
	duration := res.AvgIterationDuration
	var list []spellStats
	add := func(unit *proto.UnitMetrics, pet string) {
		for _, action := range unit.Actions {
			var damage float64
			var casts int32
			for _, target := range action.Targets {
				damage += target.Damage
				casts += target.Casts
			}
			if damage == 0 && casts == 0 {
				continue
			}
			entry := spellStats{
				ID:    marshalID(action.Id),
				Casts: float64(casts) / iterations,
				Pet:   pet,
			}
			if duration > 0 {
				entry.DPS = damage / iterations / duration
			}
			if spell := spelldata.Find(action.Id.GetSpellId()); action.Id.GetSpellId() != 0 && spell != nil {
				entry.Name = spell.Name
			}
			list = append(list, entry)
		}
	}
	for _, party := range res.RaidMetrics.Parties {
		for _, player := range party.Players {
			add(player, "")
			for _, pet := range player.Pets {
				add(pet, pet.Name)
			}
		}
	}
	sort.SliceStable(list, func(i, j int) bool { return list[i].DPS > list[j].DPS })
	return list
}

func marshalID(id *proto.ActionID) json.RawMessage {
	data, err := protojson.Marshal(id)
	if err != nil {
		return json.RawMessage("{}")
	}
	return data
}

func cloneProto[T googleProto.Message](m T) T {
	return googleProto.Clone(m).(T)
}

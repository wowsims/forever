package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode/utf16"
)

// The sim as commands, for an editor with no panel. On an .apl.siml a code action and a code lens run siml.sim or
// siml.compare through workspace/executeCommand, which sims as siml/sim does; a message gives the result, the lens
// shows it, and an inlay hint at the end of each row's line says how often the row ran per fight while the text is the
// one simmed. The client turns them on with the initialization option simCommands, as the Zed extension does; VS Code's
// panel sims through siml/sim.

const (
	commandSim     = "siml.sim"
	commandCompare = "siml.compare"
)

// The initialization options of the sim commands.
type simSettings struct {
	Commands     bool   `json:"simCommands"`
	Iterations   int    `json:"simIterations"`   // 0 sims defaultIterations.
	SettingsFile string `json:"simSettingsFile"` // As siml/sim's settingsFile.
}

// A run of a sim result, as the commands read it.
type simRunSummary struct {
	DPS struct {
		Mean  float64 `json:"mean"`
		Stdev float64 `json:"stdev"`
	} `json:"dps"`
	Iterations int `json:"iterations"`
	Rows       map[string]struct {
		Executions float64 `json:"executions"`
		Hidden     bool    `json:"hidden"`
	} `json:"rows"` // Keyed by the row's JSON path joined with "/".
}

type simSummary struct {
	simRunSummary
	Compare    *simRunSummary `json:"compare"`
	Difference *dpsDifference `json:"difference"`
	Baseline   string         `json:"baseline"`
	Character  struct {
		Warnings []string `json:"warnings"`
	} `json:"character"`
}

// The last sim of a document and the text it simmed.
type simRecord struct {
	text    string
	summary simSummary
}

// The run of the text: a compare's second run, else the sim.
func (r simSummary) run() simRunSummary {
	if r.Compare != nil {
		return *r.Compare
	}
	return r.simRunSummary
}

// The result in one line, as VS Code's status bar gives it.
func (r simSummary) line() string {
	run := r.run()
	if r.Difference != nil {
		sign := ""
		if r.Difference.Delta >= 0 {
			sign = "+"
		}
		return fmt.Sprintf("%.1f DPS, %s%.1f (%.2f%%) against %s", run.DPS.Mean, sign, r.Difference.Delta, r.Difference.Percent, r.Baseline)
	}
	standardError := 0.0
	if run.Iterations > 0 {
		standardError = run.DPS.Stdev / math.Sqrt(float64(run.Iterations))
	}
	return fmt.Sprintf("%.1f DPS ± %.1f", run.DPS.Mean, standardError)
}

// A row's runs per fight: whole numbers from 10, one decimal from 1, two below.
func perFight(n float64) string {
	switch {
	case n >= 10:
		return strconv.FormatFloat(n, 'f', 0, 64)
	case n >= 1:
		return strconv.FormatFloat(n, 'f', 1, 64)
	}
	return strconv.FormatFloat(n, 'f', 2, 64)
}

// The document at uri when the commands sim it: an open .apl.siml file that reads.
func (s *server) simmableLocked(uri string) (*analysis, *document, bool) {
	d := s.docs[uri]
	if !s.simSettings.Commands || d == nil || d.path == "" {
		return nil, nil, false
	}
	a := s.analyzeLocked(uri)
	if a == nil || a.doc == nil || a.doc.Kind().Section != "apl" {
		return nil, nil, false
	}
	return a, d, true
}

func simCommands(uri string) (sim, compare Command) {
	return Command{Title: "▶ Sim", Command: commandSim, Arguments: []any{uri}},
		Command{Title: "⇄ Compare", Command: commandCompare, Arguments: []any{uri}}
}

// The commands as source actions, where the request takes them.
func (s *server) simActionsLocked(uri string, only []string) []CodeAction {
	if _, _, ok := s.simmableLocked(uri); !ok {
		return nil
	}
	if len(only) > 0 && !slices.ContainsFunc(only, func(kind string) bool { return kind == "source" }) {
		return nil
	}
	sim, compare := simCommands(uri)
	sim.Title, compare.Title = "Sim this rotation", "Compare this rotation with the saved file, or HEAD when nothing is unsaved"
	return []CodeAction{{Title: sim.Title, Kind: "source", Command: &sim}, {Title: compare.Title, Kind: "source", Command: &compare}}
}

// The commands as lenses on the first line, each with its last result while the text is the one it simmed.
func (s *server) simLensesLocked(uri string) []CodeLens {
	_, d, ok := s.simmableLocked(uri)
	if !ok {
		return []CodeLens{}
	}
	sim, compare := simCommands(uri)
	if r, ok := s.simRecords[uri]; ok && r.text == d.text {
		if r.summary.Difference != nil {
			compare.Title += " · " + r.summary.line()
		} else {
			sim.Title += " · " + r.summary.line()
		}
	}
	return []CodeLens{{Command: &sim}, {Command: &compare}}
}

// How often each row of the last sim ran, after the first line of the row, while the text is the one simmed. A hidden
// row, and one of a used file, has no line here.
func (s *server) simHintsLocked(uri string, in Range) []InlayHint {
	out := []InlayHint{}
	a, d, ok := s.simmableLocked(uri)
	r, simmed := s.simRecords[uri]
	if !ok || !simmed || r.text != d.text {
		return out
	}
	lines := map[string]int{}
	for _, span := range a.doc.Spans() {
		parts := make([]string, len(span.Path))
		for i, part := range span.Path {
			parts[i] = fmt.Sprint(part)
		}
		lines[strings.Join(parts, "/")] = toPosition(span.Range.Start).Line
	}
	text := strings.Split(d.text, "\n")
	for key, row := range r.summary.run().Rows {
		line, ok := lines[key]
		if row.Hidden || !ok || line < in.Start.Line || line > in.End.Line || line >= len(text) {
			continue
		}
		label := "never ran"
		if row.Executions > 0 {
			label = "▸ " + perFight(row.Executions) + " per fight"
		}
		end := len(utf16.Encode([]rune(strings.TrimRight(text[line], "\r"))))
		out = append(out, InlayHint{Position: Position{Line: line, Character: end}, Label: label, PaddingLeft: true})
	}
	slices.SortFunc(out, func(a, b InlayHint) int { return a.Position.Line - b.Position.Line })
	return out
}

// workspace/executeCommand: sims the rotation at the command's one argument, a document URI. A failure comes back as
// an error message, since a client shows a command's error response in its log at most.
func (s *server) executeCommand(ctx context.Context, params json.RawMessage) (any, *ResponseError) {
	var p struct {
		Command   string            `json:"command"`
		Arguments []json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, invalidParams(err)
	}
	var uri string
	if !s.simSettings.Commands || (p.Command != commandSim && p.Command != commandCompare) {
		return nil, &ResponseError{Code: codeInvalidParams, Message: "unknown command: " + p.Command}
	}
	if len(p.Arguments) != 1 || json.Unmarshal(p.Arguments[0], &uri) != nil {
		return nil, &ResponseError{Code: codeInvalidParams, Message: p.Command + " takes a document URI"}
	}
	compare := p.Command == commandCompare

	s.mu.Lock()
	if s.simming {
		s.mu.Unlock()
		s.showMessage(messageWarning, "A sim is running; start the next once it ends.")
		return nil, nil
	}
	s.simming = true
	text := ""
	if d := s.docs[uri]; d != nil {
		text = d.text
	}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.simming = false
		s.mu.Unlock()
	}()

	name := uri
	if path, ok := uriPath(uri); ok {
		name = filepath.Base(path)
	}
	title := "Simming " + name
	if compare {
		title = "Comparing " + name
	}
	end := s.beginProgress(title)
	result, rpcErr := s.sim(ctx, simParams{uri: uri, compare: compare, settingsFile: s.simSettings.SettingsFile, iterations: s.simSettings.Iterations})
	if rpcErr != nil {
		end("")
		if rpcErr.Code != codeRequestCancelled {
			s.showMessage(messageError, rpcErr.Message)
		}
		return nil, nil
	}
	var summary simSummary
	if raw, err := json.Marshal(result); err == nil {
		err = json.Unmarshal(raw, &summary)
		if err != nil {
			s.log("the sim's result does not read: " + err.Error())
		}
	}
	s.mu.Lock()
	s.simRecords[uri] = simRecord{text: text, summary: summary}
	s.mu.Unlock()
	end(summary.line())
	s.refreshSimViews()
	s.showMessage(messageInfo, name+": "+summary.line())
	if len(summary.Character.Warnings) > 0 {
		s.showMessage(messageWarning, "Sim character: "+strings.Join(summary.Character.Warnings, "; "))
	}
	return nil, nil
}

const (
	messageError   = 1
	messageWarning = 2
	messageInfo    = 3
)

func (s *server) showMessage(kind int, message string) {
	s.w.Notify("window/showMessage", map[string]any{"type": kind, "message": message})
}

// Starts a progress report, where the client takes them, and gives the function that ends it with a message.
func (s *server) beginProgress(title string) (end func(message string)) {
	s.mu.Lock()
	if !s.clientCaps.progress {
		s.mu.Unlock()
		return func(string) {}
	}
	s.requestID++
	id := s.requestID
	s.mu.Unlock()
	token := fmt.Sprintf("siml-sim-%d", id)
	s.w.Request(id, "window/workDoneProgress/create", map[string]any{"token": token})
	s.w.Notify("$/progress", map[string]any{"token": token, "value": map[string]any{"kind": "begin", "title": title}})
	return func(message string) {
		value := map[string]any{"kind": "end"}
		if message != "" {
			value["message"] = message
		}
		s.w.Notify("$/progress", map[string]any{"token": token, "value": value})
	}
}

// Asks the client for the lenses and hints again, which a sim changed.
func (s *server) refreshSimViews() {
	s.mu.Lock()
	var methods []string
	if s.clientCaps.lensRefresh {
		methods = append(methods, "workspace/codeLens/refresh")
	}
	if s.clientCaps.hintRefresh {
		methods = append(methods, "workspace/inlayHint/refresh")
	}
	ids := make([]int, len(methods))
	for i := range methods {
		s.requestID++
		ids[i] = s.requestID
	}
	s.mu.Unlock()
	for i, method := range methods {
		s.w.Request(ids[i], method, nil)
	}
}

// The capabilities the commands add to the server's: none unless the client turned them on.
func (s *server) simCapabilities(capabilities map[string]any) {
	if !s.simSettings.Commands {
		return
	}
	capabilities["codeActionProvider"] = map[string]any{"codeActionKinds": []string{"quickfix", "refactor.rewrite", "source"}}
	capabilities["codeLensProvider"] = map[string]any{"resolveProvider": false}
	capabilities["executeCommandProvider"] = map[string]any{"commands": []string{commandSim, commandCompare}}
	capabilities["inlayHintProvider"] = true
}

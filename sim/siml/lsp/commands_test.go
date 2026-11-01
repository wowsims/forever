package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// The sim commands: off unless the client turns them on, then a code action and a lens that sim, a message with the
// result, and a hint after each row of the text that was simmed.

var simCommandsPath = filepath.Join(repository, "ui", "specs", "hunter", "dps", "apls", "default.apl.siml")

func readSimCommandsText(t *testing.T) string {
	t.Helper()
	text, err := os.ReadFile(simCommandsPath)
	if err != nil {
		t.Fatal(err)
	}
	return string(text)
}

func initializeWith(c *client, options any, capabilities any) map[string]any {
	c.t.Helper()
	var result struct {
		Capabilities map[string]any `json:"capabilities"`
	}
	c.request("initialize", map[string]any{"rootUri": pathURI(repository), "capabilities": capabilities, "initializationOptions": options}, &result)
	c.notify("initialized", map[string]any{})
	return result.Capabilities
}

func actionTitles(c *client, uri string, only []string) []string {
	c.t.Helper()
	var actions []CodeAction
	context := map[string]any{"diagnostics": []any{}}
	if only != nil {
		context["only"] = only
	}
	c.request("textDocument/codeAction", map[string]any{"textDocument": textDocument(uri), "range": Range{}, "context": context}, &actions)
	var titles []string
	for _, a := range actions {
		if a.Command != nil {
			titles = append(titles, a.Command.Command)
		}
	}
	return titles
}

func TestSimCommandsAreOffByDefault(t *testing.T) {
	c := connect(t, Options{})
	capabilities := initializeWith(c, nil, map[string]any{})
	for _, name := range []string{"codeLensProvider", "executeCommandProvider", "inlayHintProvider"} {
		if _, ok := capabilities[name]; ok {
			t.Errorf("a client that did not turn the sim commands on gets %s", name)
		}
	}
	uri := pathURI(simCommandsPath)
	c.open(uri, readSimCommandsText(t))
	if titles := actionTitles(c, uri, nil); len(titles) > 0 {
		t.Errorf("a client that did not turn the sim commands on gets the actions %v", titles)
	}
	if r := c.send("workspace/executeCommand", map[string]any{"command": commandSim, "arguments": []string{uri}}); r.Error == nil {
		t.Error("siml.sim runs for a client that did not turn the sim commands on")
	}
}

func TestSimCommands(t *testing.T) {
	var got simRequest
	c := connect(t, Options{Sim: func(ctx context.Context, root string, req json.RawMessage) (json.RawMessage, error) {
		if err := json.Unmarshal(req, &got); err != nil {
			return nil, err
		}
		return json.RawMessage(`{"dps":{"mean":1234.56,"stdev":100},"iterations":400,"spells":[],"character":{"source":"x"},
			"rows":{"priorityList/0":{"executions":12.34},"priorityList/1":{"executions":0},"priorityList/2":{"executions":3,"hidden":true}}}`), nil
	}})
	capabilities := initializeWith(c, map[string]any{"simCommands": true, "simIterations": 250, "simSettingsFile": "settings.json"}, map[string]any{
		"window":    map[string]any{"workDoneProgress": true},
		"workspace": map[string]any{"codeLens": map[string]any{"refreshSupport": true}, "inlayHint": map[string]any{"refreshSupport": true}},
	})
	if provider, _ := capabilities["executeCommandProvider"].(map[string]any); provider == nil || len(provider["commands"].([]any)) != 2 {
		t.Errorf("executeCommandProvider = %v", capabilities["executeCommandProvider"])
	}
	if capabilities["inlayHintProvider"] != true || capabilities["codeLensProvider"] == nil {
		t.Errorf("capabilities = %v", capabilities)
	}

	uri := pathURI(simCommandsPath)
	text := readSimCommandsText(t)
	c.open(uri, text)
	if titles := actionTitles(c, uri, nil); !slices.Equal(titles, []string{commandSim, commandCompare}) {
		t.Errorf("code actions = %v", titles)
	}
	if titles := actionTitles(c, uri, []string{"quickfix"}); len(titles) > 0 {
		t.Errorf("a request for quick fixes only gets %v", titles)
	}

	lensTitles := func() []string {
		var lenses []CodeLens
		c.request("textDocument/codeLens", map[string]any{"textDocument": textDocument(uri)}, &lenses)
		var titles []string
		for _, l := range lenses {
			titles = append(titles, l.Command.Title)
		}
		return titles
	}
	hints := func() []InlayHint {
		var hints []InlayHint
		c.request("textDocument/inlayHint", map[string]any{"textDocument": textDocument(uri), "range": Range{End: Position{Line: 1000}}}, &hints)
		return hints
	}
	if titles := lensTitles(); !slices.Equal(titles, []string{"▶ Sim", "⇄ Compare"}) {
		t.Errorf("lenses before a sim = %v", titles)
	}
	if h := hints(); len(h) > 0 {
		t.Errorf("hints before a sim = %v", h)
	}

	c.request("workspace/executeCommand", map[string]any{"command": commandSim, "arguments": []string{uri}})
	if got.Iterations != 250 || got.SettingsFile != "settings.json" || got.SIML != text {
		t.Errorf("request = %+v", got)
	}
	c.notification("window/showMessage", func(params json.RawMessage) bool {
		return strings.Contains(string(params), "default.apl.siml: 1234.6 DPS ± 5.0")
	})
	c.notification("$/progress", func(params json.RawMessage) bool { return strings.Contains(string(params), `"end"`) })
	c.notification("workspace/codeLens/refresh", nil)
	c.notification("workspace/inlayHint/refresh", nil)
	if titles := lensTitles(); !slices.Equal(titles, []string{"▶ Sim · 1234.6 DPS ± 5.0", "⇄ Compare"}) {
		t.Errorf("lenses after a sim = %v", titles)
	}

	lines := strings.Split(text, "\n")
	first := slices.Index(lines, "priority:") + 1
	want := []InlayHint{
		{Position: Position{Line: first, Character: len(lines[first])}, Label: "▸ 12 per fight", PaddingLeft: true},
		{Position: Position{Line: first + 1, Character: len(lines[first+1])}, Label: "never ran", PaddingLeft: true},
	}
	if h := hints(); !slices.Equal(h, want) {
		t.Errorf("hints = %+v, expected %+v", h, want)
	}

	c.change(uri, 2, strings.Replace(text, "priority:\n", "priority:\n  cast multi_shot\n", 1))
	if h := hints(); len(h) > 0 {
		t.Errorf("hints after an edit = %v", h)
	}
	if titles := lensTitles(); !slices.Equal(titles, []string{"▶ Sim", "⇄ Compare"}) {
		t.Errorf("lenses after an edit = %v", titles)
	}
}

func TestSimCommandFailureIsAMessage(t *testing.T) {
	c := connect(t, Options{Sim: func(context.Context, string, json.RawMessage) (json.RawMessage, error) {
		return nil, errors.New("no character")
	}})
	initializeWith(c, map[string]any{"simCommands": true}, map[string]any{})
	uri := pathURI(simCommandsPath)
	c.open(uri, readSimCommandsText(t))
	if r := c.send("workspace/executeCommand", map[string]any{"command": commandCompare, "arguments": []string{uri}}); r.Error != nil {
		t.Errorf("a failed sim answers %v", r.Error)
	}
	c.notification("window/showMessage", func(params json.RawMessage) bool {
		var p struct {
			Type    int    `json:"type"`
			Message string `json:"message"`
		}
		return json.Unmarshal(params, &p) == nil && p.Type == messageError && strings.Contains(p.Message, "no character")
	})
}

package lsp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// siml/sim builds the request `siml sim` reads from the open text, and adds the difference to a compare's result.
func TestSimRequestAndResult(t *testing.T) {
	var got simRequest
	c := connect(t, Options{Sim: func(ctx context.Context, root string, req json.RawMessage) (json.RawMessage, error) {
		if err := json.Unmarshal(req, &got); err != nil {
			return nil, err
		}
		if root != repository {
			t.Errorf("root = %s", root)
		}
		return json.RawMessage(`{"dps":{"mean":100,"stdev":10},"iterations":100,"rows":{},"spells":[],"character":{"source":"x"},
			"compare":{"dps":{"mean":110,"stdev":10},"iterations":100,"rows":{},"spells":[]}}`), nil
	}})
	c.initialize(repository, map[string]any{})
	path := filepath.Join(repository, "ui", "specs", "hunter", "dps", "apls", "default.apl.siml")
	saved, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(string(saved), "priority:\n", "priority:\n  cast multi_shot\n", 1)
	c.open(pathURI(path), edited)
	var result struct {
		Baseline   string        `json:"baseline"`
		Difference dpsDifference `json:"difference"`
	}
	c.request("siml/sim", map[string]any{"uri": pathURI(path), "compare": true, "iterations": 500000, "timeline": true}, &result)
	if got.Path != "ui/specs/hunter/dps/apls/default.apl.siml" || got.Iterations != maxIterations || !got.Timeline {
		t.Errorf("request = %+v", got)
	}
	if got.SIML != string(saved) || got.Compare == nil || got.Compare.SIML != edited {
		t.Error("a compare of an edited text does not sim the saved file as the base and the text against it")
	}
	if result.Baseline != "saved" || result.Difference.Delta != 10 || result.Difference.Percent != 10 {
		t.Errorf("result = %+v", result)
	}

	c.change(pathURI(path), 2, "siml 1\nspec hunter\n\npriority:\n  cast hunters_mrk\n")
	r := c.send("siml/sim", map[string]any{"uri": pathURI(path)})
	if r.Error == nil || !strings.HasPrefix(r.Error.Message, "The rotation does not read (line 5:") {
		t.Errorf("a sim of a text with an error = %+v", r.Error)
	}
}

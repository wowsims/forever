//go:build with_db

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// The language server's sims run `siml sim`'s request in process, and stop when the request is cancelled.
func TestSimlSimJSON(t *testing.T) {
	path := filepath.Join("ui", "specs", "hunter", "dps", "apls", "default.apl.siml")
	text, err := os.ReadFile(filepath.Join(root, path))
	if err != nil {
		t.Fatal(err)
	}
	req, _ := json.Marshal(map[string]any{"siml": string(text), "path": path, "iterations": 5})
	raw, err := simlSimJSON(context.Background(), root, req)
	if err != nil {
		t.Fatal(err)
	}
	var res struct {
		DPS        distribution `json:"dps"`
		Iterations int32        `json:"iterations"`
		Error      string       `json:"error"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		t.Fatal(err)
	}
	if res.Error != "" || res.DPS.Mean <= 0 || res.Iterations != 5 {
		t.Errorf("result = %s", raw)
	}

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := simlSimJSON(cancelled, root, req); !errors.Is(err, context.Canceled) {
		t.Errorf("a cancelled sim gave %v", err)
	}
	if _, err := simlSimJSON(context.Background(), root, []byte(`{"siml": "", "unknown": 1}`)); err == nil {
		t.Error("a request with an unknown field is read")
	}

}

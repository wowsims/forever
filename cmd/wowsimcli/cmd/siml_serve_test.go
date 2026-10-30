package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Serves the requests, one a line, with the name table at names, and gives the responses.
func serveRequests(t *testing.T, names string, reqs ...simlServeRequest) []simlServeResponse {
	t.Helper()
	var in, out bytes.Buffer
	for _, req := range reqs {
		line, err := json.Marshal(req)
		if err != nil {
			t.Fatal(err)
		}
		in.Write(append(line, '\n'))
	}
	if err := serveSiml(&in, &out, names); err != nil {
		t.Fatal(err)
	}
	var responses []simlServeResponse
	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		var res simlServeResponse
		if err := json.Unmarshal([]byte(line), &res); err != nil {
			t.Fatalf("%v: %s", err, line)
		}
		responses = append(responses, res)
	}
	if len(responses) != len(reqs) {
		t.Fatalf("%d responses to %d requests", len(responses), len(reqs))
	}
	return responses
}

const committedNamesFile = "../../../sim/siml/names.json"

const serveRotation = "siml 1\nspec hunter\n\npriority:\n  cast multi_shot\n"

func TestServeReadsARotation(t *testing.T) {
	res := serveRequests(t, committedNamesFile, simlServeRequest{ID: 7, Path: "x.apl.siml", Text: serveRotation})[0]
	if res.ID != 7 || res.Error != "" || len(res.Errors) > 0 || len(res.Warnings) > 0 {
		t.Fatalf("%+v", res)
	}
	if want := `{"type":"TypeAPL","priorityList":[{"action":{"castSpell":{"spellId":{"spellId":2643}}}}]}`; !jsonEqual(t, res.JSON, want) {
		t.Errorf("json %s, want %s", res.JSON, want)
	}
}

func TestServeReadsWhatAFullSetupUses(t *testing.T) {
	dir := t.TempDir()
	used := filepath.Join(dir, "apls", "r.apl.siml")
	if err := os.MkdirAll(filepath.Dir(used), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(used, []byte(serveRotation), 0o644); err != nil {
		t.Fatal(err)
	}
	setup := "siml 1\nspec hunter\nuse \"apls/r.apl.siml\"\n\nsettings:\n  distance_from_target 11\n"
	res := serveRequests(t, committedNamesFile, simlServeRequest{ID: 1, Path: filepath.Join(dir, "setup.siml"), Text: setup})[0]
	if res.Error != "" || len(res.Errors) > 0 {
		t.Fatalf("%+v", res)
	}
	if len(res.Uses) != 1 || res.Uses[0] != used {
		t.Errorf("uses %q, want [%q]", res.Uses, used)
	}
	var settings struct {
		Player struct {
			DistanceFromTarget float64
			Rotation           struct{ PriorityList []any }
		}
	}
	if err := json.Unmarshal(res.JSON, &settings); err != nil {
		t.Fatal(err)
	}
	if settings.Player.DistanceFromTarget != 11 || len(settings.Player.Rotation.PriorityList) != 1 {
		t.Errorf("json %s holds neither the setting nor the used rotation", res.JSON)
	}
}

func TestServeFailsAnErrorAndAWarningUnderUISpecs(t *testing.T) {
	broken := "siml 1\nspec hunter\n\npriority:\n  cast hunters_mrak\n"
	undefinedGroup := "siml 1\nspec hunter\n\npriority:\n  group \"Nowhere\"\n"
	underSpecs := filepath.Join(repositoryFiles(t, nil), "ui/specs/hunter/dps/apls/nowhere.apl.siml")
	responses := serveRequests(t, committedNamesFile,
		simlServeRequest{ID: 1, Path: "broken.apl.siml", Text: broken},
		simlServeRequest{ID: 2, Path: "nowhere.apl.siml", Text: undefinedGroup},
		simlServeRequest{ID: 3, Path: underSpecs, Text: undefinedGroup},
	)
	want := serveDiagnostic{Line: 5, Col: 8, Severity: "error", Code: "unknown-name", Message: "no spell is named hunters_mrak; did you mean hunters_mark?"}
	if res := responses[0]; len(res.Errors) != 1 || res.Errors[0] != want || res.JSON != nil {
		t.Errorf("broken: %+v", res)
	}
	if res := responses[1]; len(res.Errors) != 0 || len(res.Warnings) != 1 || res.Warnings[0].Code != "undefined-group" || res.JSON == nil {
		t.Errorf("a warning outside ui/specs: %+v", res)
	}
	if res := responses[2]; len(res.Errors) != 1 || res.Errors[0].Severity != "warning" || len(res.Warnings) != 0 {
		t.Errorf("a warning under ui/specs: %+v", res)
	}
}

func TestServeReadsTheNameTableAgainWhenItChanges(t *testing.T) {
	names := filepath.Join(t.TempDir(), "names.json")
	write := func(table string) {
		if err := os.WriteFile(names, []byte(table), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(`{"classes":{},"enchants":{},"items":{},"shared":{},"suffixes":{}}`)
	inReader, inWriter := io.Pipe()
	outReader, outWriter := io.Pipe()
	go func() { outWriter.CloseWithError(serveSiml(inReader, outWriter, names)) }()
	encoder, responses := json.NewEncoder(inWriter), json.NewDecoder(outReader)
	// Each request waits for the response before it, so the table changes between the two.
	serve := func(id int) (res simlServeResponse) {
		if err := encoder.Encode(simlServeRequest{ID: id, Path: "x.apl.siml", Text: serveRotation}); err != nil {
			t.Fatal(err)
		}
		if err := responses.Decode(&res); err != nil {
			t.Fatal(err)
		}
		return res
	}
	before := serve(1)
	write(`{"classes":{"hunter":{"spell":{"multi_shot":{"spellId":2643,"tag":-1,"registeredBy":["hunter"]}}}},"enchants":{},"items":{},"shared":{},"suffixes":{}}`)
	after := serve(2)
	inWriter.Close()
	if len(before.Errors) != 1 || before.Errors[0].Code != "unknown-name" {
		t.Errorf("with no name: %+v", before)
	}
	if after.Error != "" || len(after.Errors) != 0 || after.JSON == nil {
		t.Errorf("with the name: %+v", after)
	}
}

func jsonEqual(t *testing.T, a json.RawMessage, b string) bool {
	t.Helper()
	var x, y any
	if err := json.Unmarshal(a, &x); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &y); err != nil {
		t.Fatal(err)
	}
	xs, _ := json.Marshal(x)
	ys, _ := json.Marshal(y)
	return bytes.Equal(xs, ys)
}

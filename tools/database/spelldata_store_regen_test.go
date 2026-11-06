package database

// Rebuilds sim/core/spelldata/spells_auto_gen.go from the game files rows committed in
// assets/db_inputs/spell_store_inputs.json and asserts the committed file is what comes out, so a
// hand-edited or stale store fails here. No game files database and no build tag: this is the one gate
// over the store that CI can run, since tools/database/wowsims.db is gitignored.
//
// It re-runs everything the generator derives - the closure over triggers, overrides and tooltip
// references, the hand links, the talent curves, the tooltip hints, the overrides and the emitter -
// against rows it did not derive. A change to any of those fails until the store is regenerated with
// `go run ./tools/database/gen_spelldata`, which rewrites both files together.

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
)

// The generator reads its paths from the repository root - the captured inputs, and the proc audit's
// list beside this file - and a test runs in its own package's directory.
const repositoryRoot = "../.."

func TestStoreRegeneratesFromTheCommittedInputs(t *testing.T) {
	inRepositoryRoot(t)

	rendered, err := renderStore(committedInputs(t), newRankEnumNamer())
	if err != nil {
		t.Fatalf("rendering the store: %v", err)
	}
	assertRendersCommitted(t, "sim/core/spelldata/spells_auto_gen.go", rendered)
}

func TestEveryClassSpellExportRendersTheGameFilesValues(t *testing.T) {
	inRepositoryRoot(t)

	const noHandTrigger = "var generatedHandTriggers = map[int32][]int32{}\n"
	sim, err := renderStore(committedInputs(t), newRankEnumNamer())
	if err != nil {
		t.Fatalf("rendering the store: %v", err)
	}
	if !bytes.Contains(sim, []byte("// override:")) || bytes.Contains(sim, []byte(noHandTrigger)) {
		t.Fatalf("the store the sim is built against renders no override or no hand trigger, so this test checks nothing")
	}

	EveryClassSpell = true
	t.Cleanup(func() { EveryClassSpell = false })
	export, err := renderStore(committedInputs(t), newRankEnumNamer())
	if err != nil {
		t.Fatalf("rendering the export: %v", err)
	}
	for i, line := range lines(export) {
		if strings.Contains(line, "// override:") {
			t.Errorf("the export renders an override at line %d: %s", i+1, line)
			break
		}
	}
	if !bytes.Contains(export, []byte(noHandTrigger)) {
		t.Errorf("the export renders a hand trigger")
	}
}

// The committed inputs, read once for every test that renders from them: nothing a rendering does
// writes through them. The first caller has to be in the repository root.
func committedInputs(t *testing.T) *storeInputs {
	t.Helper()

	inputs, err := readCommittedInputs()
	if err != nil {
		t.Fatalf("%v", err)
	}
	return inputs
}

var readCommittedInputs = sync.OnceValues(func() (*storeInputs, error) {
	return readStoreInputs(spellStoreInputsPath)
})

// The rendering counts its rows on stderr, which says nothing a passing gate needs, so it goes
// nowhere for the duration.
func inRepositoryRoot(t *testing.T) {
	t.Helper()

	t.Chdir(repositoryRoot)
	progress = io.Discard
	t.Cleanup(func() { progress = os.Stderr })
}

func assertRendersCommitted(t *testing.T, path string, rendered []byte) {
	t.Helper()

	committed, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	if bytes.Equal(committed, rendered) {
		return
	}

	line, want, got := firstDifference(committed, rendered)
	t.Errorf("%s is not what the committed inputs render, from line %d:\n  committed: %s\n  rendered:  %s\n"+
		"regenerate with `go run ./tools/database/gen_spelldata`",
		path, line, want, got)
}

// The first line the two differ on, so a diff of megabytes reports as one row.
func firstDifference(want, got []byte) (int, string, string) {
	a, b := lines(want), lines(got)
	for i := range a {
		if i >= len(b) {
			return i + 1, a[i], "(the rendered store ends here)"
		}
		if a[i] != b[i] {
			return i + 1, a[i], b[i]
		}
	}
	if len(b) > len(a) {
		return len(a) + 1, "(the committed store ends here)", b[len(a)]
	}
	return 0, "", ""
}

func lines(b []byte) []string {
	var out []string
	scanner := bufio.NewScanner(bytes.NewReader(b))
	scanner.Buffer(make([]byte, 0, 1<<20), 1<<20)
	for scanner.Scan() {
		out = append(out, strings.TrimRight(scanner.Text(), " \t"))
	}
	return out
}

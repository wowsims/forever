package siml

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func TestIdentifier(t *testing.T) {
	f, err := os.Open("testdata/ident.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	lines := bufio.NewScanner(f)
	n := 0
	for lines.Scan() {
		n++
		name, want, ok := strings.Cut(lines.Text(), "\t")
		if !ok {
			t.Fatalf("testdata/ident.txt:%d: want a game name, a tab and its identifier", n)
		}
		if got := Identifier(name); got != want {
			t.Errorf("testdata/ident.txt:%d: Identifier(%q) = %q, want %q", n, name, got, want)
		}
	}
	if err := lines.Err(); err != nil {
		t.Fatal(err)
	}
	if n == 0 {
		t.Fatal("testdata/ident.txt holds no pairs")
	}
}

func TestSnake(t *testing.T) {
	for in, want := range map[string]string{
		"FasterAttackII": "faster_attack_ii",
		"MainHandAuto":   "main_hand_auto",
		"E20":            "e20",
		"TypeAPL":        "type_apl",
		"APLValue":       "apl_value",
		"StatMP5":        "stat_mp5",
		"mhImbueId":      "mh_imbue_id",
		"conserveStart":  "conserve_start",
		"Finger1":        "finger1",
		"Slot1Name":      "slot1_name",
		"":               "",
	} {
		if got := snake(in); got != want {
			t.Errorf("snake(%q) = %q, want %q", in, got, want)
		}
	}
}

package database

// Every proc an item, an enchant or a set bonus can reach, and what the sim makes of it. A proc the
// rows state enough for is registered; one they do not is named in unsupported_procs.txt with the
// reason, so the list of what the sim cannot model is reviewed rather than discovered. That file is a
// local baseline, not tracked: a checkout without it writes it on the first run.
//
// Reads the committed capture and the committed store, so it needs no game files database: the
// roots are the ones the item, enchant and set tables gave, and the rows are the ones the sim
// itself reads.

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

const unsupportedProcsPath = "tools/database/unsupported_procs.txt"

func TestEveryReachableProcIsSupportedOrListed(t *testing.T) {
	inRepositoryRoot(t)

	inputs := committedInputs(t)

	var out strings.Builder
	out.WriteString("# Every proc reachable from an item, an enchant or a set bonus that the sim does not\n")
	out.WriteString("# register as the game files state it, as \"<spell id> <reason>\". Written by\n")
	out.WriteString("# TestEveryReachableProcIsSupportedOrListed in tools/database. Regenerate with\n")
	out.WriteString("#   UPDATE_UNSUPPORTED_PROCS=1 go test ./tools/database/ -run TestEveryReachableProcIsSupportedOrListed -count=1\n")
	out.WriteString("#\n")
	out.WriteString("# A row is judged in the kinder of the two shapes an item can put it in: an aura that hears\n")
	out.WriteString("# hits through its own proc mask, and a weapon proc whose hits the weapon counts.\n\n")

	supported := 0
	procs := reachableProcs(inputs.ItemRoots)
	for _, row := range procs {
		reasons := itemProcReasons(row)
		if len(reasons) == 0 {
			supported++
			continue
		}
		fmt.Fprintf(&out, "%d %s\n", row.ID, strings.Join(reasons, "; "))
	}

	t.Logf("%d roots, %d procs reachable from them, %d of which the rows state enough for",
		len(inputs.ItemRoots), len(procs), supported)

	if supported == 0 {
		t.Fatal("no reachable proc resolves at all, so this proves nothing")
	}

	committed, err := os.ReadFile(unsupportedProcsPath)
	if os.IsNotExist(err) || os.Getenv("UPDATE_UNSUPPORTED_PROCS") != "" {
		if err := os.WriteFile(unsupportedProcsPath, []byte(out.String()), 0644); err != nil {
			t.Fatalf("%v", err)
		}
		t.Logf("wrote the baseline %s", unsupportedProcsPath)
		return
	}
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(committed) == out.String() {
		return
	}

	for _, line := range changedProcLines(string(committed), out.String()) {
		t.Error(line)
	}
	t.Errorf("%s is not what the rows say; once the moves above are the intended ones, rewrite the baseline:\n"+
		"  UPDATE_UNSUPPORTED_PROCS=1 go test ./tools/database/ -run TestEveryReachableProcIsSupportedOrListed -count=1",
		unsupportedProcsPath)
}

// Why the sim cannot build the listener the game files describe, in the shape that refuses least. A
// weapon proc's listener is the item's rather than the row's, so a row that states no proc mask at
// all is not refused for it - Annihilator's 16928 is that shape.
func itemProcReasons(row *spelldata.Spell) []string {
	reasons := spelldata.ItemProcUnsupportedWithRate(row, false, 0)
	if asWeaponProc := spelldata.ItemProcUnsupportedWithRate(row, true, 0); len(asWeaponProc) < len(reasons) {
		return asWeaponProc
	}
	return reasons
}

// Every row carrying a proc trigger that an item, enchant or set spell reaches, in id order. The
// walk follows the game files' trigger edges, which is how a proc hangs below the spell an item names.
func reachableProcs(roots []int32) []*spelldata.Spell {
	seen := map[int32]bool{}
	var procs []*spelldata.Spell

	var walk func(id int32)
	walk = func(id int32) {
		if id == 0 || seen[id] {
			return
		}
		seen[id] = true

		row := spelldata.Find(id)
		if row == spelldata.Nil {
			return
		}

		if triggersAProcEffect(row) {
			procs = append(procs, row)
		}

		for i := range row.Effects {
			walk(row.Effects[i].TriggerID)
		}
	}

	for _, id := range roots {
		walk(id)
	}

	sort.Slice(procs, func(i, j int) bool { return procs[i].ID < procs[j].ID })
	return procs
}

// The three auras that fire a spell off a hit. A_PROC_TRIGGER_DAMAGE is in as well: the shield
// spikes state their retaliation that way rather than through a spell of their own.
func triggersAProcEffect(row *spelldata.Spell) bool {
	for i := range row.Effects {
		aura := row.Effects[i].Aura
		if aura.IsProcTrigger() || aura == dbcenums.A_PROC_TRIGGER_SPELL_COPY ||
			aura == dbcenums.A_PROC_TRIGGER_DAMAGE {
			return true
		}
	}
	return false
}

// A line one side has and the other does not: a proc that stopped resolving, one that started, and
// one whose reason changed, which shows up as both.
func changedProcLines(committed string, current string) []string {
	was := map[string]bool{}
	for line := range strings.SplitSeq(committed, "\n") {
		was[line] = true
	}

	var changed []string
	now := map[string]bool{}
	for line := range strings.SplitSeq(current, "\n") {
		now[line] = true
		if !was[line] && line != "" {
			changed = append(changed, "not listed: "+line)
		}
	}

	for line := range strings.SplitSeq(committed, "\n") {
		if !now[line] && line != "" && !strings.HasPrefix(line, "#") {
			changed = append(changed, "listed but supported now: "+line)
		}
	}

	return changed
}

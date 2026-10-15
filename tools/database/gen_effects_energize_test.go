package database

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/dbc"
)

// An on-use whose spell restores the wearer's mana, rage or energy registers through the energize
// constructor, in a group of its own.
func TestOnUseEnergizeRoutesFromTheSpellItCasts(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		itemID  int
		spellID int32
	}{
		{19954, 24532}, // Renataki's Charm of Trickery: E_ENERGIZE 60 energy
		{19951, 24571}, // Gri'lek's Charm of Might: E_ENERGIZE 300 rage-tenths
		{14152, 18385}, // Robe of the Archmage: E_ENERGIZE 500 mana, rolled
		{23027, 28760}, // Warmth of Forgiveness: E_ENERGIZE 500 mana
		{20525, 24884}, // Earthen Sigil: A_PERIODIC_ENERGIZE 40 mana every 1 s for 10 s
	} {
		item := instance.Items[tc.itemID]
		parsed := item.ToUIItem()
		effects := dbc.MergeItemEffectsForAllStates(parsed)
		i := slices.IndexFunc(effects, func(e *proto.ItemEffect) bool { return e.BuffId == tc.spellID && e.GetOnUse() != nil })
		if i < 0 {
			t.Fatalf("item %d carries no on-use effect on %d", tc.itemID, tc.spellID)
		}

		groups := map[string]Group{}
		if got := TryParseOnUseEffect(parsed, effects[i], instance, groups); got != EffectParseResultSuccess {
			t.Errorf("item %d parsed as %v, want it registered", tc.itemID, got)
			continue
		}

		entries := groups["Resources"].Entries
		if len(entries) != 1 {
			t.Errorf("item %d: %d entries in Resources, want 1", tc.itemID, len(entries))
			continue
		}
		r := entries[0].Proc
		if r == nil || r.TriggerSpellID != int(tc.spellID) || r.Shape != ShapeEnergize || !entries[0].Supported {
			t.Errorf("item %d: routing %+v, supported %v; want an energize on %d", tc.itemID, r, entries[0].Supported, tc.spellID)
			continue
		}
		if got := r.OnUseConstructor(); got != "NewSpellDataEnergizeOnUse" {
			t.Errorf("item %d registers through %s, want NewSpellDataEnergizeOnUse", tc.itemID, got)
		}
	}
}

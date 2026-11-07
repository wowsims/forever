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
		parsed.ItemEffects = dbc.MergeItemEffectsForAllStates(parsed)
		i := slices.IndexFunc(parsed.ItemEffects, func(e *proto.ItemEffect) bool { return e.BuffId == tc.spellID && e.GetOnUse() != nil })
		if i < 0 {
			t.Fatalf("item %d carries no on-use effect on %d", tc.itemID, tc.spellID)
		}

		groups := map[string]Group{}
		if got := TryParseOnUseEffect(parsed, parsed.ItemEffects[i], instance, groups); got != EffectParseResultSuccess {
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

// A relic's cast proc whose dummy trigger names the spell it fires registers the gain through the
// energize proc constructor, its class's mask read as the relic's lock. The same trigger on an item
// of no relic type keeps the class mask refusal.
func TestRelicCastProcRoutesToTheGainItFires(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		itemID         int
		trigger, fired int32
	}{
		{23004, 28847, 28848}, // Idol of Longevity: Healing Touch Refund, 25 mana
		{23005, 28849, 28850}, // Totem of Flowing Water: Lesser Healing Wave, 10 mana
	} {
		item := instance.Items[tc.itemID]
		parsed := item.ToUIItem()
		parsed.ItemEffects = dbc.MergeItemEffectsForAllStates(parsed)
		i := slices.IndexFunc(parsed.ItemEffects, func(e *proto.ItemEffect) bool { return e.BuffId == tc.trigger && e.GetProc() != nil })
		if i < 0 {
			t.Fatalf("item %d carries no proc effect on %d", tc.itemID, tc.trigger)
		}

		groups := map[string]Group{}
		if got := TryParseProcEffect(parsed, parsed.ItemEffects[i], instance, groups); got != EffectParseResultSuccess {
			t.Errorf("item %d parsed as %v, want it registered", tc.itemID, got)
			continue
		}
		entries := groups["Procs"].Entries
		if len(entries) != 1 {
			t.Errorf("item %d: %d entries in Procs, want 1", tc.itemID, len(entries))
			continue
		}
		r := entries[0].Proc
		if r == nil || r.Shape != ShapeEnergize || r.BuffSpellID != int(tc.fired) || !entries[0].Supported {
			t.Errorf("item %d: routing %+v, supported %v; want an energize firing %d", tc.itemID, r, entries[0].Supported, tc.fired)
			continue
		}
		if got := r.ProcConstructor(); got != "NewSpellDataEnergizeProc" {
			t.Errorf("item %d registers through %s, want NewSpellDataEnergizeProc", tc.itemID, got)
		}

		parsed.RangedWeaponType = proto.RangedWeaponType_RangedWeaponTypeUnknown
		if got := routeItemProc(parsed, parsed.ItemEffects[i]); !slices.Contains(got.Unsupported, "the trigger carries a class mask") {
			t.Errorf("item %d of no relic type is refused for %v, want its class mask named", tc.itemID, got.Unsupported)
		}
	}
}

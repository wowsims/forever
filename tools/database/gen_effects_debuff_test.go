package database

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/dbc"
)

// A chance-on-hit spell whose aura lands on the enemy routes to the debuff shape, never to a stat buff
// on the wearer, even where the effect entry resolves stats from it:
//   - Frostguard 12797's Chilled 16927 slows melee beside a movement slow.
//   - Dark Iron Sunderer 11607's Cleave Armor 15280 takes 300 armor, resolved as Armor -300.
//   - Bleakwood Hew 12769's Bleakwood Curse 16871 takes 25 of every magic resistance, stacking to 3.
//   - Howling Hide 284262's Howler's Cry 1315767 takes 60 attack power.
//
// The Lobotomizer 19324's Brain Damage 1290950 deals damage beside its cast slow and routes to the
// damage shape, whose spell carries the debuff. Its rate is the override's 0.4 PPM, so it registers;
// none of the others states a rate for its chance on hit, so they stay commented out. Mug O' Hurt
// 4090's Dazed 13496 slows movement alone and stays skipped.
func TestChanceOnHitDebuffsRoute(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	if ItemEffectIsSupported(instance, 13496) {
		t.Errorf("Dazed 13496, a movement slow alone, is not skipped")
	}

	for _, tc := range []struct {
		itemID  int
		spellID int32
		damage  bool
		rated   bool
		slows   []int32
		debuffs []int32
	}{
		{12797, 16927, false, false, []int32{1}, []int32{1}},
		{19324, 1290950, true, true, []int32{2}, []int32{2}},
		{11607, 15280, false, false, nil, []int32{1}},
		{12769, 16871, false, false, nil, []int32{1}},
		{284262, 1315767, false, false, nil, []int32{1, 2}},
	} {
		row := spelldata.Find(tc.spellID)
		if got := row.SlowEffects(); !slices.Equal(got, tc.slows) {
			t.Errorf("%d slows the target through effects %v, want %v", tc.spellID, got, tc.slows)
		}
		if got := row.DebuffEffects(); !slices.Equal(got, tc.debuffs) {
			t.Errorf("%d debuffs the target through effects %v, want %v", tc.spellID, got, tc.debuffs)
		}

		item := instance.Items[tc.itemID]
		parsed := item.ToUIItem()
		parsed.ItemEffects = dbc.MergeItemEffectsForAllStates(parsed)
		i := slices.IndexFunc(parsed.ItemEffects, func(e *proto.ItemEffect) bool { return e.BuffId == tc.spellID })
		if i < 0 {
			t.Fatalf("item %d carries no effect on %d", tc.itemID, tc.spellID)
		}

		groups := map[string]Group{}
		wantResult, wantUnsupported := EffectParseResultRefused, []string{spelldata.ReasonStatesNoRate}
		if tc.rated {
			wantResult, wantUnsupported = EffectParseResultSuccess, nil
		}
		if got := TryParseProcEffect(parsed, parsed.ItemEffects[i], instance, groups); got != wantResult {
			t.Errorf("item %d parsed as %v, want %v", tc.itemID, got, wantResult)
			continue
		}
		entry := groups["Procs"].Entries[0]
		r := entry.Proc
		wantShape := ShapeDebuff
		if tc.damage {
			wantShape = ShapeDamage
		}
		if r.Shape != wantShape || !r.IsWeaponProc || r.TriggerSpellID != int(tc.spellID) {
			t.Errorf("item %d: shape %v, weapon proc %v, trigger %d; want a weapon proc on %d, shape %v",
				tc.itemID, r.Shape, r.IsWeaponProc, r.TriggerSpellID, tc.spellID, wantShape)
		}
		if !slices.Equal(r.Unsupported, wantUnsupported) {
			t.Errorf("item %d refused for %q, want %q", tc.itemID, r.Unsupported, wantUnsupported)
		}
	}
}

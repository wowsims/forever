package database

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/dbc"
)

// A chance-on-hit spell whose aura lands on the enemy routes to the debuff shape, never to a stat buff
// on the wearer, even where the effect entry resolves stats from it: Dark Iron Sunderer 11607's Cleave
// Armor 15280 takes armor, resolved as negative Armor. The Lobotomizer 19324's Brain Damage 1290950
// deals damage beside its cast slow and routes to the damage shape, whose spell carries the debuff.
// Mug O' Hurt 4090's Dazed 13496 slows movement alone and stays skipped.
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
		slows   bool
	}{
		{12797, 16927, false, true},
		{19324, 1290950, true, true},
		{11607, 15280, false, false},
		{12769, 16871, false, false},
		{12798, 16928, false, false},
	} {
		row := spelldata.Find(tc.spellID)
		debuffs, slows := row.DebuffEffects(), row.SlowEffects()
		if len(debuffs) == 0 {
			t.Errorf("%d debuffs the target through no effect", tc.spellID)
		}
		if (len(slows) > 0) != tc.slows {
			t.Errorf("%d slows the target through effects %v, want a slow: %v", tc.spellID, slows, tc.slows)
		}
		for _, slow := range slows {
			if !slices.Contains(debuffs, slow) {
				t.Errorf("%d slows through effect %d, which is not among its debuffs %v", tc.spellID, slow, debuffs)
			}
		}

		item := instance.Items[tc.itemID]
		parsed := item.ToUIItem()
		parsed.ItemEffects = dbc.MergeItemEffectsForAllStates(parsed)
		i := slices.IndexFunc(parsed.ItemEffects, func(e *proto.ItemEffect) bool { return e.BuffId == tc.spellID })
		if i < 0 {
			t.Fatalf("item %d carries no effect on %d", tc.itemID, tc.spellID)
		}

		groups := map[string]Group{}
		wantUnsupported, ok := rateOutcome(t, tc.itemID, tc.spellID, TryParseProcEffect(parsed, parsed.ItemEffects[i], instance, groups))
		if !ok {
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

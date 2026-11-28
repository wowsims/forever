package shared

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// A proc's damage spell without a missile speed deals its damage without allocating.
func TestProcDamageWithoutAMissileSpeedDoesNotAllocate(t *testing.T) {
	for _, tc := range []struct {
		neckID, row int32
	}{
		{991130, 7712},
		{991131, 6297},
	} {
		neck, _ := listeningNeck(tc.neckID, core.CallbackOnSpellHitDealt, tc.row)
		sim, caster := newOnUseSimAgainst(t, NewSpellDataDamageOnUse, map[int32]*proto.ItemEffect{}, neck, 3)
		spell := caster.GetSpell(core.ActionID{SpellID: tc.row})
		target := sim.Encounter.ActiveTargetUnits[0]

		if allocs := testing.AllocsPerRun(100, func() { spell.Cast(sim, target) }); allocs != 0 {
			t.Errorf("a cast of %d allocates %v times, want 0", tc.row, allocs)
		}
		if spell.SpellMetrics[target.UnitIndex].TotalDamage == 0 {
			t.Errorf("the casts of %d dealt nothing", tc.row)
		}
	}
}

// The flight of a proc's missile, Linken's Boomerang 15712, allocates the pending arrival, not a copy
// of the results.
func TestProcDamageMissileAllocatesOnlyItsFlight(t *testing.T) {
	neck, _ := listeningNeck(991132, core.CallbackOnSpellHitDealt, 15712)
	sim, caster := newOnUseSimAgainst(t, NewSpellDataDamageOnUse, map[int32]*proto.ItemEffect{}, neck, 1)
	spell := caster.GetSpell(core.ActionID{SpellID: 15712})
	target := sim.Encounter.ActiveTargetUnits[0]

	if allocs := testing.AllocsPerRun(100, func() { spell.Cast(sim, target) }); allocs > 4 {
		t.Errorf("a missile cast allocates %v times, want at most the flight's 4", allocs)
	}
}

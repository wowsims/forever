package shared

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The row's damage spell used from a test trinket against the given number of targets, run until its
// cast has finished, and what it dealt to each target.
func dealtPerTarget(t *testing.T, itemID int32, spellID int32, targetCount int) []float64 {
	t.Helper()
	sim, caster := newOnUseSimAgainst(t, NewSpellDataDamageOnUse,
		map[int32]*proto.ItemEffect{itemID: testOnUse(spellID, 60000, 0, 0)}, &proto.ItemSpec{}, targetCount)
	spell := onUseSpell(t, caster, itemID)

	if !spell.Cast(sim, sim.Encounter.ActiveTargetUnits[0]) {
		t.Fatalf("spell %d could not be used", spellID)
	}
	stepPast(t, sim, sim.CurrentTime+5*time.Second)

	dealt := make([]float64, targetCount)
	for i, target := range sim.Encounter.ActiveTargetUnits {
		metrics := spell.SpellMetrics[target.UnitIndex]
		if metrics.Misses+metrics.ResistedHits > 0 {
			t.Fatalf("spell %d missed target %d", spellID, i)
		}
		dealt[i] = metrics.TotalDamage
	}
	return dealt
}

func sum(values []float64) float64 {
	var total float64
	for _, v := range values {
		total += v
	}
	return total
}

// Fiery Blaze 6297 hits TARGET_UNIT_DEST_AREA_ENEMY, every enemy within its radius, and the sim counts
// every enemy of the encounter as within it. Each target rolls its own damage.
func TestAreaDamageHitsEveryTarget(t *testing.T) {
	effect := spelldata.MustFind(6297).DamageEffect()
	low, high := effect.Min(), effect.Max()

	dealt := dealtPerTarget(t, 991101, 6297, 3)
	for i, d := range dealt {
		if d < low || d > high {
			t.Errorf("target %d took %v, want one roll of %v to %v", i, d, low, high)
		}
	}
	if dealt[0] == dealt[1] && dealt[1] == dealt[2] {
		t.Errorf("every target took %v, want a roll of its own each", dealt[0])
	}
}

// Shard of the Fallen Star 26789 burns every enemy in the area for "$s1 total Fire damage": one roll,
// divided evenly among the three targets.
func TestSplitDamageDividesOneRollAmongTheTargets(t *testing.T) {
	effect := spelldata.MustFind(26789).DamageEffect()
	low, high := effect.Min(), effect.Max()

	dealt := dealtPerTarget(t, 991102, 26789, 3)
	if dealt[0] != dealt[1] || dealt[1] != dealt[2] {
		t.Errorf("the targets took %v, want one equal share each", dealt)
	}
	if total := sum(dealt); total < low-1e-6 || total > high+1e-6 {
		t.Errorf("the shares add up to %v, want one roll of %v to %v", total, low, high)
	}
}

// Prototype Pathcarver 1295270 splits its damage "between up to $s3 nearby enemies", its MaxTargets,
// and does not roll. One target fewer than the cap share it among themselves; one more than the cap
// share it as many ways as the cap and the last takes none. A copy fixes the cap at 5.
func TestSplitDamageStopsAtTheRowsMaxTargets(t *testing.T) {
	editRow(t, 1295270, func(s *spelldata.Spell) { s.MaxTargets = 5 })
	row := spelldata.MustFind(1295270)
	amount := row.DamageEffect().Average()
	needAmount(t, "1295270's damage", amount)
	maxTargets := int(row.MaxTargets)

	for _, tc := range []struct {
		itemID       int32
		targets, hit int
	}{
		{991103, maxTargets - 1, maxTargets - 1},
		{991104, maxTargets + 1, maxTargets},
	} {
		dealt := dealtPerTarget(t, tc.itemID, 1295270, tc.targets)
		for i, d := range dealt {
			want := amount / float64(tc.hit)
			if i >= tc.hit {
				want = 0
			}
			if math.Abs(d-want) > 1e-6 {
				t.Errorf("%d targets: target %d took %v, want %v", tc.targets, i, d, want)
			}
		}
		if total := sum(dealt); math.Abs(total-amount) > 1e-6 {
			t.Errorf("%d targets: the shares add up to %v, want the one amount %v", tc.targets, total, amount)
		}
	}
}

// Fire Strike 7712 hits TARGET_UNIT_TARGET_ENEMY alone, however many enemies the encounter has.
func TestSingleTargetDamageHitsOnlyItsTarget(t *testing.T) {
	dealt := dealtPerTarget(t, 991110, 7712, 3)
	if dealt[0] == 0 {
		t.Errorf("the target took nothing")
	}
	if dealt[1] != 0 || dealt[2] != 0 {
		t.Errorf("the other targets took %v and %v, want nothing", dealt[1], dealt[2])
	}
}

// The ticks of the row's damage over time each target took once the use and the whole damage over
// time ran out, from a test trinket against the given number of targets.
func ticksPerTarget(t *testing.T, itemID int32, spellID int32, targetCount int) []int32 {
	t.Helper()
	row := spelldata.MustFind(spellID)
	sim, caster := newOnUseSimAgainst(t, NewSpellDataDamageOnUse,
		map[int32]*proto.ItemEffect{itemID: testOnUse(spellID, 60000, 0, 0)}, &proto.ItemSpec{}, targetCount)
	spell := onUseSpell(t, caster, itemID)

	if !spell.Cast(sim, sim.Encounter.ActiveTargetUnits[0]) {
		t.Fatalf("spell %d could not be used", spellID)
	}
	stepPast(t, sim, sim.CurrentTime+row.CastTime()+row.Duration()+time.Second)

	ticks := make([]int32, targetCount)
	for i, target := range sim.Encounter.ActiveTargetUnits {
		ticks[i] = spell.SpellMetrics[target.UnitIndex].Ticks
	}
	return ticks
}

// Ez-Thro Radiation Bomb's 435968 hits every enemy in its area and puts its damage over time on each
// one the hit lands on.
func TestAreaDamageOverTimeGoesOnEveryEnemyTheHitLandsOn(t *testing.T) {
	row := spelldata.MustFind(435968)
	want := int32(row.Duration() / row.PeriodicDamageEffect().Period())
	for i, got := range ticksPerTarget(t, 991112, row.ID, 3) {
		if got != want {
			t.Errorf("target %d took %d ticks, want %d", i, got, want)
		}
	}
}

// An area damage over time with no hit reaches every enemy in its area up to the row's MaxTargets:
// one target more than the cap takes none. A copy of Ez-Thro Radiation Bomb's 435968 that keeps only
// its damage over time and caps its targets stands in for the rows that state one.
func TestAreaDamageOverTimeWithoutAHitStopsAtTheRowsMaxTargets(t *testing.T) {
	editRow(t, 435968, func(s *spelldata.Spell) {
		s.Effects = []spelldata.Effect{*s.PeriodicDamageEffect()}
		s.MaxTargets = 3
	})
	row := spelldata.MustFind(435968)
	if row.DamageEffect() != spelldata.NilEffect || row.MaxTargets == 0 {
		t.Fatalf("435968's copy deals a hit or states no MaxTargets, nothing to check")
	}
	want := int32(row.Duration() / row.PeriodicDamageEffect().Period())
	for i, got := range ticksPerTarget(t, 991113, row.ID, int(row.MaxTargets)+1) {
		if i < int(row.MaxTargets) && got != want {
			t.Errorf("target %d took %d ticks, want %d", i, got, want)
		}
		if i >= int(row.MaxTargets) && got != 0 {
			t.Errorf("target %d, past the row's %d, took %d ticks, want none", i, row.MaxTargets, got)
		}
	}
}

// Chain Lightning 21179 jumps to its chain targets, each keeping the row's share of the damage the one
// before it took, and a target past the chain takes none. A copy without its variance and with three
// chain targets makes the jumps exact.
func TestChainDamageJumpsToItsChainTargets(t *testing.T) {
	editRow(t, 21179, func(s *spelldata.Spell) {
		s.Effects[0].Variance = 0
		s.Effects[0].ChainTargets = 3
	})
	effect := spelldata.MustFind(21179).DamageEffect()
	amount, keep, jumps := effect.Average(), float64(effect.ChainAmp), int(effect.ChainTargets)
	needAmount(t, "21179's damage", amount)
	needAmount(t, "21179's chain amplitude", keep)

	dealt := dealtPerTarget(t, 991111, 21179, jumps+1)
	for i := range dealt {
		want := 0.0
		if i < jumps {
			want = amount * math.Pow(keep, float64(i))
		}
		if math.Abs(dealt[i]-want) > 1e-3 {
			t.Errorf("target %d took %v, want %v", i, dealt[i], want)
		}
	}
}

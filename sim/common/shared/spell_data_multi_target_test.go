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

// Fiery Blaze 6297 hits TARGET_UNIT_DEST_AREA_ENEMY, every enemy within 3 yards, and the sim counts
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

// Prototype Pathcarver 1295270 splits its damage "between up to $s3 nearby enemies", MaxTargets 5, and
// does not roll. Three targets share it three ways; six share it five ways and the sixth takes none.
func TestSplitDamageStopsAtTheRowsMaxTargets(t *testing.T) {
	amount := spelldata.MustFind(1295270).DamageEffect().Average()

	for _, tc := range []struct {
		targets, hit int
	}{
		{3, 3},
		{6, 5},
	} {
		dealt := dealtPerTarget(t, 991103+int32(tc.targets), 1295270, tc.targets)
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

// Chain Lightning 21179 jumps to 3 targets, each keeping 0.7 of the damage the one before it took. A
// copy without its variance makes the jumps exact.
func TestChainDamageJumpsToItsChainTargets(t *testing.T) {
	editRow(t, 21179, func(s *spelldata.Spell) { s.Effects[0].Variance = 0 })
	amount := spelldata.MustFind(21179).DamageEffect().Average()

	dealt := dealtPerTarget(t, 991111, 21179, 4)
	for i, want := range []float64{amount, amount * 0.7, amount * 0.7 * 0.7, 0} {
		if math.Abs(dealt[i]-want) > 1e-3 {
			t.Errorf("target %d took %v, want %v", i, dealt[i], want)
		}
	}
}

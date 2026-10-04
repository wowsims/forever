package shared

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Linken's Boomerang 11905's on-use, 15712, flies at 20 yards a second. Thrown from 20 yards it hits
// 1 s after its 0.5 s cast completes, not when the cast does.
func TestMissileDamageLandsAfterItsTravelTime(t *testing.T) {
	const itemID int32 = 991120
	sim, caster := newOnUseSim(t, NewSpellDataDamageOnUse, map[int32]*proto.ItemEffect{itemID: testOnUse(15712, 180000, 0, 0)})
	target := sim.Encounter.ActiveTargetUnits[0]
	table := caster.AttackTables[target.UnitIndex]
	for _, hand := range []*core.AttackTableHand{&table.MH, &table.OH, &table.Ranged} {
		hand.BaseMissChance, hand.BaseDodgeChance, hand.BaseParryChance, hand.BaseBlockChance = 0, 0, 0, 0
	}

	boomerang := onUseSpell(t, caster, itemID)
	metrics := &boomerang.SpellMetrics[target.UnitIndex]
	caster.DistanceFromTarget = 20
	start := sim.CurrentTime
	if !boomerang.Cast(sim, target) {
		t.Fatalf("the boomerang could not be used from 20 yards")
	}

	stepPast(t, sim, start+500*time.Millisecond+time.Millisecond)
	if metrics.Casts != 1 || metrics.Hits+metrics.Crits != 1 || metrics.TotalDamage != 0 {
		t.Errorf("at the end of the cast: %d casts, %d hits, %v damage; want 1 cast whose hit is still in flight",
			metrics.Casts, metrics.Hits+metrics.Crits, metrics.TotalDamage)
	}

	stepPast(t, sim, start+1500*time.Millisecond-time.Millisecond)
	if metrics.TotalDamage != 0 {
		t.Errorf("the boomerang dealt %v before its 1 s flight ended", metrics.TotalDamage)
	}

	stepPast(t, sim, start+1500*time.Millisecond+time.Millisecond)
	if metrics.TotalDamage == 0 {
		t.Errorf("the boomerang had dealt nothing 1 s after its cast completed")
	}
}

// Fire Strike 7712 states no missile speed: from 20 yards it deals its damage the moment it is cast.
func TestDamageWithoutAMissileSpeedLandsAtCast(t *testing.T) {
	const itemID int32 = 991121
	sim, caster := newOnUseSim(t, NewSpellDataDamageOnUse, map[int32]*proto.ItemEffect{itemID: testOnUse(7712, 60000, 0, 0)})
	target := sim.Encounter.ActiveTargetUnits[0]

	caster.DistanceFromTarget = 20
	fireStrike := onUseSpell(t, caster, itemID)
	if !fireStrike.Cast(sim, target) {
		t.Fatalf("Fire Strike could not be used from 20 yards")
	}
	if fireStrike.SpellMetrics[target.UnitIndex].TotalDamage == 0 {
		t.Errorf("Fire Strike dealt nothing at cast")
	}
}

// A copy of Fiery Blaze 6297, whose area reaches every enemy, flying at 20 yards a second: from 20
// yards all three targets take their damage 1 s after its 3 s cast completes.
func TestAreaMissileDamageLandsOnEveryTargetAfterItsTravelTime(t *testing.T) {
	const itemID int32 = 991122
	editRow(t, 6297, func(s *spelldata.Spell) { s.Speed = 20 })
	sim, caster := newOnUseSimAgainst(t, NewSpellDataDamageOnUse,
		map[int32]*proto.ItemEffect{itemID: testOnUse(6297, 60000, 0, 0)}, &proto.ItemSpec{}, 3)
	targets := sim.Encounter.ActiveTargetUnits

	blaze := onUseSpell(t, caster, itemID)
	caster.DistanceFromTarget = 20
	start := sim.CurrentTime
	if !blaze.Cast(sim, targets[0]) {
		t.Fatalf("Fiery Blaze could not be used from 20 yards")
	}
	dealt := func() []float64 {
		values := make([]float64, len(targets))
		for i, target := range targets {
			values[i] = blaze.SpellMetrics[target.UnitIndex].TotalDamage
		}
		return values
	}

	stepPast(t, sim, start+3*time.Second+time.Millisecond)
	if blaze.SpellMetrics[targets[0].UnitIndex].Casts != 1 || sum(dealt()) != 0 {
		t.Errorf("at the end of the cast the targets took %v, want nothing yet", dealt())
	}

	stepPast(t, sim, start+4*time.Second-time.Millisecond)
	if sum(dealt()) != 0 {
		t.Errorf("the targets took %v before the 1 s flight ended", dealt())
	}

	stepPast(t, sim, start+4*time.Second+time.Millisecond)
	for i, d := range dealt() {
		if d == 0 {
			t.Errorf("target %d took nothing 1 s after the cast completed", i)
		}
	}
}

// A copy of the lasso's 443265 flying at 20 yards a second puts its damage over time on the target
// when it arrives from 20 yards, 1 s after it is used.
func TestMissileDotAppliesOnArrival(t *testing.T) {
	const itemID int32 = 991123
	editRow(t, lassoSpell, func(s *spelldata.Spell) { s.Speed = 20 })
	sim, caster := newOnUseSim(t, NewSpellDataDamageOnUse, map[int32]*proto.ItemEffect{itemID: lassoOnUse()})
	target := sim.Encounter.ActiveTargetUnits[0]

	lasso := onUseSpell(t, caster, itemID)
	caster.DistanceFromTarget = 20
	start := sim.CurrentTime
	if !lasso.Cast(sim, target) {
		t.Fatalf("the lasso could not be used from 20 yards")
	}

	stepPast(t, sim, start+time.Second-time.Millisecond)
	if lasso.Dot(target).IsActive() {
		t.Errorf("the lasso's damage over time was up before its 1 s flight ended")
	}

	stepPast(t, sim, start+time.Second+time.Millisecond)
	if !lasso.Dot(target).IsActive() {
		t.Errorf("the lasso's damage over time was not up 1 s after it was used")
	}
	if got := lasso.Dot(target).ExpiresAt(); got != start+time.Second+12*time.Second {
		t.Errorf("the lasso's damage over time runs until %v, want 12 s from its arrival at %v", got, start+time.Second)
	}
}

//go:build with_db

package forever_test

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Dragonbreath Chili 12217's 15852 rolls its chance on the eater's melee hits, at most once per its
// internal cooldown, and casts 15851, whose fire damage stays within the row's roll.
func TestDragonbreathChiliBreathesFireOnMeleeHits(t *testing.T) {
	const sword = 2078
	trigger, breath := spelldata.MustFind(15852), spelldata.MustFind(15851)
	if _, character := warriorWithConsumes(t, &proto.ConsumesSpec{}, sword, 0, proto.MobType_MobTypeHumanoid); character.GetSpell(core.ActionID{SpellID: breath.ID}) != nil {
		t.Fatal("15851 is registered without the chili")
	}

	sim, character := warriorWithConsumes(t, &proto.ConsumesSpec{ExtraIds: []int32{12217}}, sword, 0, proto.MobType_MobTypeHumanoid)
	aura := character.GetAura("Dragonbreath Chili")
	if aura == nil || !aura.IsActive() {
		t.Fatal("the chili's trigger is not up from the pull")
	}
	spell := character.GetSpell(core.ActionID{SpellID: breath.ID})
	if spell == nil {
		t.Fatal("12217 registers no 15851")
	}

	const fight = 170 * time.Second
	character.AutoAttacks.EnableAutoSwing(sim)
	// Longer than stepTo's step budget covers.
	for steps := 0; sim.CurrentTime < fight; steps++ {
		if steps > 1_000_000 {
			t.Fatalf("the sim did not reach %v", fight)
		}
		sim.Step()
	}

	target := character.CurrentTarget
	metrics := spell.SpellMetrics[target.UnitIndex]
	casts := metrics.Casts
	if most := int32(fight/trigger.ICD()) + 1; casts == 0 || casts > most {
		t.Errorf("15851 was cast %d times in %v, want at least once and at most %d on the row's %v cooldown", casts, fight, most, trigger.ICD())
	}
	// A partial resist can take a hit below the roll, never above it.
	hitDamage := metrics.TotalDamage - metrics.TotalCritDamage
	if high := float64(metrics.Hits) * (breath.DamageEffect().Max() + spell.SpellDamage(target)); hitDamage <= 0 || hitDamage > high {
		t.Errorf("%d hits of 15851 dealt %v, want more than none and at most %v", metrics.Hits, hitDamage, high)
	}
}

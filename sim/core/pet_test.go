package core_test

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// A dynamic pet inherits from its owner's totals: strength the owner gains reaches the pet at the
// next pet heartbeat, and comes back off at the heartbeat after the owner loses it.
func TestDynamicPetInheritsFromTheOwnersTotals(t *testing.T) {
	sim, owner := setupPetOwnerSim(&proto.RaidBuffs{}, &proto.PartyBuffs{})
	for _, unit := range slices.Concat(sim.Raid.AllUnits, sim.Encounter.AllTargetUnits) {
		unit.AutoAttacks.CancelAutoSwing(sim)
	}
	pet := owner.dynamic
	pet.Enable(sim, pet)

	inherited := func() float64 { return max(owner.GetStat(stats.Strength), owner.GetStat(stats.Agility)) }
	start := pet.GetStat(stats.AttackPower)
	if want := inherited(); start != want {
		t.Fatalf("the pet starts with %v attack power, want the owner's better of strength and agility, %v", start, want)
	}

	owner.AddStatsDynamic(sim, stats.Stats{stats.Strength: 100})
	stepTo(t, sim, sim.CurrentTime+core.PetUpdateInterval)
	if got, want := pet.GetStat(stats.AttackPower), inherited(); got != want {
		t.Errorf("with +100 strength on the owner the pet has %v attack power, want %v", got, want)
	}

	owner.AddStatsDynamic(sim, stats.Stats{stats.Strength: -100})
	stepTo(t, sim, sim.CurrentTime+core.PetUpdateInterval)
	if got := pet.GetStat(stats.AttackPower); got != start {
		t.Errorf("with the strength gone the pet has %v attack power, want the %v it started with", got, start)
	}
}

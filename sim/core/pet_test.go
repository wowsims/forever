package core_test

import (
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

// A pet set on the target before the pull swings from then on; one left alone waits for the pull.
func TestPetAttackStartsThePetBeforeThePull(t *testing.T) {
	prepullAt := func(doAt string, action *proto.APLAction) *proto.APLPrepullAction {
		return &proto.APLPrepullAction{
			DoAtValue: &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: doAt}}},
			Action:    action,
		}
	}
	// The owner's move at -5 s starts every run's prepull there.
	moveAway := prepullAt("-5s", &proto.APLAction{Action: &proto.APLAction_Move{Move: &proto.APLActionMove{
		RangeFromTarget: &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: "10"}}},
	}}})
	petAttack := prepullAt("-5s", &proto.APLAction{Action: &proto.APLAction_PetAttack{PetAttack: &proto.APLActionPetAttack{}}})

	swingsBeforePull := func(prepull ...*proto.APLPrepullAction) int32 {
		sim := core.NewSim(&proto.RaidSimRequest{
			SimOptions: &proto.SimOptions{RandomSeed: 100},
			Raid: &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{{
				Name:      "Owner",
				Class:     proto.Class_ClassHunter,
				Spec:      &proto.Player_Hunter{},
				Equipment: &proto.EquipmentSpec{},
				Rotation:  &proto.APLRotation{PrepullActions: prepull},
			}}}}},
			Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 180},
		}, simsignals.CreateSignals())
		sim.Reset()
		sim.PrePull()
		stepTo(t, sim, -time.Millisecond)
		return sim.Raid.Parties[0].Players[0].(*petOwner).pet.AutoAttacks.MHAuto().SpellMetrics[0].Casts
	}

	if got := swingsBeforePull(moveAway); got != 0 {
		t.Errorf("with no Pet Attack the pet swings %d times before the pull, want none", got)
	}
	// In reach from the start, with a 2 s swing: at -5 s, -3 s and -1 s.
	if got := swingsBeforePull(moveAway, petAttack); got != 3 {
		t.Errorf("set on the target at -5 s the pet swings %d times before the pull, want 3", got)
	}
}

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

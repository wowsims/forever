package hunter

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

// A dual wielding hunter at 20 yd, one second into the fight, with a rotation that does nothing on
// its own.
func newStartAttackSim(t *testing.T, rotation *proto.APLRotation) (*core.Simulation, *Hunter, func(time.Duration)) {
	t.Helper()
	player := &proto.Player{
		Name:               "Weaver",
		Race:               proto.Race_RaceOrc,
		Class:              proto.Class_ClassHunter,
		Equipment:          core.GetGearSet("../../ui/specs/hunter/dps/gear_sets", "p1_melee").GearSet,
		TalentsString:      DefaultSVTalents,
		Spec:               DefaultOptions,
		Rotation:           rotation,
		DistanceFromTarget: 20,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	hunter := sim.Raid.Parties[0].Players[0].(*Hunter)
	if !hunter.AutoAttacks.IsDualWielding {
		t.Fatal("the p1_melee hunter does not dual wield")
	}
	stepTo := func(at time.Duration) {
		t.Helper()
		sim.AddPendingAction(&core.PendingAction{NextActionAt: at, OnAction: func(*core.Simulation) {}})
		for i := 0; sim.CurrentTime < at; i++ {
			if i > 100000 {
				t.Fatalf("the sim did not reach %v", at)
			}
			sim.Step()
		}
	}
	stepTo(time.Second)
	return sim, hunter, stepTo
}

// Walks to the given range and returns when the walk crosses melee range, if it does.
func walk(sim *core.Simulation, hunter *Hunter, stepTo func(time.Duration), to float64) time.Duration {
	from := hunter.DistanceFromTarget
	hunter.MoveTo(to, sim)
	crossing := sim.CurrentTime + core.DurationFromSeconds((from-core.MaxMeleeRange)/hunter.GetMovementSpeed())
	stepTo(sim.CurrentTime + core.DurationFromSeconds((from-to)/hunter.GetMovementSpeed()) + time.Millisecond)
	return crossing
}

// The off hand's half swing runs from when the attack starts. Started by the rotation on the way in,
// at 7 yd, the off hand swings half a swing after that, before half a swing has passed since the
// arrival.
func TestStartAttackOnTheWayInStartsTheOffHandsHalfSwing(t *testing.T) {
	rotation := &proto.APLRotation{PriorityList: []*proto.APLListItem{{Action: &proto.APLAction{
		Action: &proto.APLAction_StartAttack{StartAttack: &proto.APLActionStartAttack{}},
		Condition: &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{
			Op:  proto.APLValueCompare_OpLt,
			Lhs: &proto.APLValue{Value: &proto.APLValue_UnitDistance{UnitDistance: &proto.APLValueUnitDistance{}}},
			Rhs: &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: "7"}}},
		}}},
	}}}}
	sim, hunter, stepTo := newStartAttackSim(t, rotation)
	aa := &hunter.AutoAttacks
	walk(sim, hunter, stepTo, 8) // into the deadzone, out of both ranges

	at7 := sim.CurrentTime + core.DurationFromSeconds(1/hunter.GetMovementSpeed())
	arrival := walk(sim, hunter, stepTo, core.MaxMeleeRange-0.01)
	stepTo(arrival + aa.OffhandSwingSpeed()/2)

	// The rotation looks every 10ms while it has nothing to do.
	if got, want := aa.OffhandPreviousSwing(), at7+aa.OffhandSwingSpeed()/2; got < want || got > want+20*time.Millisecond {
		t.Errorf("an attack started at 7 yd (%v) swings the off hand at %v, want half a swing later at %v", at7, got, want)
	}
}

// Walking in without starting the attack starts it on arrival, half a swing before the off hand.
func TestWalkingInStartsTheAttackOnArrival(t *testing.T) {
	sim, hunter, stepTo := newStartAttackSim(t, &proto.APLRotation{})
	aa := &hunter.AutoAttacks

	arrival := walk(sim, hunter, stepTo, core.MaxMeleeRange-0.01)
	if got, want := aa.OffhandSwingAt(), arrival+aa.OffhandSwingSpeed()/2; (got - want).Abs() > time.Millisecond {
		t.Errorf("arriving at %v puts the off hand's swing at %v, want half a swing later at %v", arrival, got, want)
	}
}

// Starting Auto Shot turns the attack off, so the next trip in starts it again with a fresh half
// swing. A trip out and back with no shot keeps it on: the off hand swings on arrival once ready.
func TestAutoShotTurnsTheAttackOff(t *testing.T) {
	t.Run("a shot between two trips", func(t *testing.T) {
		sim, hunter, stepTo := newStartAttackSim(t, &proto.APLRotation{})
		aa := &hunter.AutoAttacks
		walk(sim, hunter, stepTo, core.MaxMeleeRange-0.01)
		walk(sim, hunter, stepTo, 20)
		shots := aa.PreviousRangedAttack()
		stepTo(sim.CurrentTime + 4*time.Second)
		if aa.PreviousRangedAttack() == shots {
			t.Fatal("no Auto Shot went off at range")
		}

		arrival := walk(sim, hunter, stepTo, core.MaxMeleeRange-0.01)
		if got, want := aa.OffhandSwingAt(), arrival+aa.OffhandSwingSpeed()/2; (got - want).Abs() > time.Millisecond {
			t.Errorf("back in at %v after a shot, the off hand swings at %v, want half a swing later at %v", arrival, got, want)
		}
	})

	t.Run("no shot between two trips", func(t *testing.T) {
		sim, hunter, stepTo := newStartAttackSim(t, &proto.APLRotation{})
		aa := &hunter.AutoAttacks
		walk(sim, hunter, stepTo, core.MaxMeleeRange-0.01)
		walk(sim, hunter, stepTo, 8) // the deadzone, where no shot goes off
		stepTo(sim.CurrentTime + 4*time.Second)

		arrival := walk(sim, hunter, stepTo, core.MaxMeleeRange-0.01)
		if got := aa.OffhandPreviousSwing(); (got - arrival).Abs() > time.Millisecond {
			t.Errorf("back in at %v with the attack still on, the off hand swung at %v, want on arrival", arrival, got)
		}
	})
}

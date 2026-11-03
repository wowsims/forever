package holy

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/paladin"
)

// Consecrated Ground 2/2 in place of Infusion of Light.
var consecratedGroundTalents = "55320015020121251-503020005001"

// Consecration marks its bonus targets as it is cast, and the mark outlasts the ground by the mark
// row's duration. A Holy hit on a marked target takes Consecrated Ground's bonus, and a target past
// the bonus targets takes none.
func TestConsecratedGroundMarksTheBonusTargetsAtTheCast(t *testing.T) {
	ground := paladin.ConsecrationRankMap.Highest()
	bonusTargets := int(ground.Effect(dbcenums.A_PERIODIC_DUMMY, 0).Average())
	want := core.MultiplierFromPercent(spelldata.MustFind(1310905).EffectN(1).BasePoints)
	markEnds := ground.Duration() + spelldata.MustFind(1280353).Duration()

	targets := make([]*proto.Target, bonusTargets+1)
	for i := range targets {
		targets[i] = &proto.Target{Level: 63}
	}
	player := core.WithSpec(&proto.Player{
		Class:         proto.Class_ClassPaladin,
		Race:          proto.Race_RaceHuman,
		Equipment:     core.GetGearSet("../../../ui/specs/paladin/holy/gear_sets", "default").GearSet,
		TalentsString: consecratedGroundTalents,
		Rotation:      &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}, PlayerOptions)
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 180, Targets: targets},
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	pally := sim.Raid.Parties[0].Players[0].(paladin.PaladinAgent).GetPaladin()
	if pally.Talents.ConsecratedGround != 2 {
		t.Fatalf("the talents take Consecrated Ground %d/2, want 2/2", pally.Talents.ConsecratedGround)
	}
	pally.AutoAttacks.CancelAutoSwing(sim)

	consecration := pally.GetSpell(core.ActionID{SpellID: ground.ID})
	if consecration == nil {
		t.Fatal("Consecration is not registered")
	}
	units := append([]*core.Unit(nil), sim.Encounter.ActiveTargetUnits...)
	taken := func(target *core.Unit) float64 {
		return consecration.TargetDamageMultiplier(sim, pally.AttackTables[target.UnitIndex], true)
	}
	// The sampled multiplier on the first target at a time, which the sim reaches by an action of its own.
	sampled := map[time.Duration]float64{}
	sampleAt := func(at time.Duration) {
		sim.AddPendingAction(core.NewDelayedAction(core.DelayedActionOptions{
			DoAt:     at,
			OnAction: func(sim *core.Simulation) { sampled[at] = taken(units[0]) },
		}))
	}

	if got := taken(units[0]); got != 1 {
		t.Errorf("before the cast the first target takes %v, want 1", got)
	}
	consecration.SkipCastAndApplyEffects(sim, pally.CurrentTarget)
	for i, target := range units {
		expected := core.TernaryFloat64(i < bonusTargets, want, 1)
		if got := taken(target); math.Abs(got-expected) > 1e-9 {
			t.Errorf("at the cast target %d takes %v, want %v", i, got, expected)
		}
	}

	before, after := markEnds-time.Millisecond, markEnds+time.Millisecond
	sampleAt(before)
	sampleAt(after)
	for sim.CurrentTime <= after && !sim.Step() {
	}
	if got := sampled[before]; math.Abs(got-want) > 1e-9 {
		t.Errorf("just before the mark ends the first target takes %v, want %v", got, want)
	}
	if got := sampled[after]; got != 1 {
		t.Errorf("after the mark ends the first target takes %v, want 1", got)
	}
}

// The mark is Consecration's own, laid on the first bonus targets with or without the talent; without
// it the mark's damage from caster mod stays at the row's 0 points.
func TestConsecrationMarksTheBonusTargetsWithoutTheTalent(t *testing.T) {
	ground := paladin.ConsecrationRankMap.Highest()
	bonusTargets := int(ground.Effect(dbcenums.A_PERIODIC_DUMMY, 0).Average())

	targets := make([]*proto.Target, bonusTargets+1)
	for i := range targets {
		targets[i] = &proto.Target{Level: 63}
	}
	player := core.WithSpec(&proto.Player{
		Class:         proto.Class_ClassPaladin,
		Race:          proto.Race_RaceHuman,
		Equipment:     core.GetGearSet("../../../ui/specs/paladin/holy/gear_sets", "default").GearSet,
		TalentsString: StandardTalents,
		Rotation:      &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}, PlayerOptions)
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 180, Targets: targets},
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	pally := sim.Raid.Parties[0].Players[0].(paladin.PaladinAgent).GetPaladin()
	if pally.Talents.ConsecratedGround != 0 {
		t.Fatalf("the talents take Consecrated Ground %d/2, want none", pally.Talents.ConsecratedGround)
	}

	consecration := pally.GetSpell(core.ActionID{SpellID: ground.ID})
	if consecration == nil {
		t.Fatal("Consecration is not registered")
	}
	consecration.SkipCastAndApplyEffects(sim, pally.CurrentTarget)
	for i, target := range sim.Encounter.ActiveTargetUnits {
		mark := target.GetAuraByID(core.ActionID{SpellID: 1280353})
		if mark == nil {
			t.Fatalf("target %d has no Consecration mark registered", i)
		}
		if got, expected := mark.IsActive(), i < bonusTargets; got != expected {
			t.Errorf("at the cast target %d is marked: %v, want %v", i, got, expected)
		}
		if got := consecration.TargetDamageMultiplier(sim, pally.AttackTables[target.UnitIndex], true); got != 1 {
			t.Errorf("without the talent target %d takes %v, want 1", i, got)
		}
	}
}

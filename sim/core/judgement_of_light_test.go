package core_test

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

func TestJudgementOfLightHealsOnlyOnDamagingMeleeHits(t *testing.T) {
	setup := func(t *testing.T) (*core.Simulation, *core.FakeRageWarrior, *core.Unit) {
		t.Helper()
		sim := core.NewSim(&proto.RaidSimRequest{
			SimOptions: &proto.SimOptions{RandomSeed: 100},
			Raid: core.SinglePlayerRaidProto(&proto.Player{
				Name:      "Warrior",
				Class:     proto.Class_ClassWarrior,
				Race:      proto.Race_RaceHuman,
				Buffs:     &proto.IndividualBuffs{},
				Spec:      &proto.Player_DpsWarrior{},
				Equipment: &proto.EquipmentSpec{},
				Rotation:  &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
			}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{JudgementOfLight: true}),
			Encounter: &proto.Encounter{
				Targets:  []*proto.Target{{Name: "target", Level: 63, MobType: proto.MobType_MobTypeHumanoid}},
				Duration: 180,
			},
		}, simsignals.CreateSignals())
		sim.Reset()

		fw := sim.Raid.Parties[0].Players[0].(*core.FakeRageWarrior)
		target := sim.Encounter.AllTargetUnits[0]
		if aura := target.GetAura("Judgement of Light (External)"); aura == nil || !aura.IsActive() {
			t.Fatalf("the target should carry an active Judgement of Light")
		}
		fw.AutoAttacks.CancelAutoSwing(sim)
		fw.AddStatDynamic(sim, stats.Health, 10000)
		fw.RemoveHealth(sim, fw.CurrentHealth()/2)
		if fw.CurrentHealth() == 0 || fw.CurrentHealth() >= fw.MaxHealth() {
			t.Fatalf("the attacker has %v of %v health, want room to be healed", fw.CurrentHealth(), fw.MaxHealth())
		}
		return sim, fw, target
	}

	// The heal lands one spell batch after the hit, so the sim steps past it once every hit is in.
	healedWithin := func(sim *core.Simulation, fw *core.FakeRageWarrior, target *core.Unit, spell *core.Spell, tries int) bool {
		before := fw.CurrentHealth()
		for range tries {
			spell.SkipCastAndApplyEffects(sim, target)
		}
		for sim.CurrentTime <= core.SpellBatchWindow && !sim.Step() {
		}
		return fw.CurrentHealth() > before
	}

	t.Run("a damaging melee special heals", func(t *testing.T) {
		sim, fw, target := setup(t)
		strike := fw.RegisterSpell(core.SpellConfig{
			ActionID:         core.ActionID{SpellID: 78},
			SpellSchool:      core.SpellSchoolPhysical,
			DefenseType:      core.DefenseTypeMelee,
			ProcMask:         core.ProcMaskMeleeMHSpecial,
			Flags:            core.SpellFlagNoActiveDefense,
			DamageMultiplier: 1,
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeMeleeSpecialHit)
			},
		})
		if !healedWithin(sim, fw, target, strike, 200) {
			t.Errorf("200 damaging strikes on a judged target never healed the attacker")
		}
	})

	t.Run("a melee special dealing no damage does not", func(t *testing.T) {
		sim, fw, target := setup(t)
		sunder := fw.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 7386},
			SpellSchool: core.SpellSchoolPhysical,
			DefenseType: core.DefenseTypeMelee,
			ProcMask:    core.ProcMaskMeleeMHSpecial,
			Flags:       core.SpellFlagNoActiveDefense,
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
			},
		})
		if healedWithin(sim, fw, target, sunder, 200) {
			t.Errorf("a landed hit dealing no damage healed the attacker")
		}
	})
}

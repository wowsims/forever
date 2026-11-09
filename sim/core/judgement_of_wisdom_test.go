package core_test

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

func TestJudgementOfWisdomPaysOnlyOnLandedHits(t *testing.T) {
	setup := func(t *testing.T) (*core.Simulation, *core.Character, *core.Unit) {
		t.Helper()
		sim := core.NewSim(&proto.RaidSimRequest{
			SimOptions: &proto.SimOptions{RandomSeed: 100},
			Raid: core.SinglePlayerRaidProto(&proto.Player{
				Name:      "Warlock",
				Class:     proto.Class_ClassWarlock,
				Race:      proto.Race_RaceHuman,
				Buffs:     &proto.IndividualBuffs{},
				Spec:      &proto.Player_Warlock{},
				Equipment: &proto.EquipmentSpec{},
				Rotation:  &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
			}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{JudgementOfWisdom: true}),
			Encounter: &proto.Encounter{
				Targets:  []*proto.Target{{Name: "target", Level: 63, MobType: proto.MobType_MobTypeHumanoid}},
				Duration: 180,
			},
		}, simsignals.CreateSignals())
		sim.Reset()

		char := sim.Raid.Parties[0].Players[0].GetCharacter()
		target := sim.Encounter.AllTargetUnits[0]
		if aura := target.GetAura("Judgement of Wisdom (External)"); aura == nil || !aura.IsActive() {
			t.Fatalf("the target should carry an active Judgement of Wisdom")
		}
		if !char.HasManaBar() {
			t.Fatalf("the attacker needs a mana bar for the judgement to pay into")
		}
		return sim, char, target
	}

	strike := func(char *core.Character, procMask core.ProcMask, defense core.DefenseType, landed bool) *core.Spell {
		return char.RegisterSpell(core.SpellConfig{
			ActionID:         core.ActionID{SpellID: 100},
			SpellSchool:      core.SpellSchoolPhysical,
			DefenseType:      defense,
			ProcMask:         procMask,
			DamageMultiplier: 1,
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				if landed {
					spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeAlwaysHit)
				} else {
					spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeAlwaysMiss)
				}
			},
		})
	}

	// The mana arrives one spell batch after the hit, so the sim steps past it once every hit is in.
	// Regeneration ticks while it steps, so the judgement's own metrics tell whether it paid.
	paidWithin := func(sim *core.Simulation, char *core.Character, target *core.Unit, spell *core.Spell, tries int) bool {
		for range tries {
			spell.SkipCastAndApplyEffects(sim, target)
		}
		for sim.CurrentTime <= core.SpellBatchWindow && !sim.Step() {
		}
		return char.JowManaMetrics != nil && char.JowManaMetrics.Events > 0
	}

	cases := []struct {
		name     string
		procMask core.ProcMask
		defense  core.DefenseType
		landed   bool
		pays     bool
	}{
		{"a landed melee swing pays", core.ProcMaskMeleeMHAuto, core.DefenseTypeMelee, true, true},
		{"a missed melee swing does not", core.ProcMaskMeleeMHAuto, core.DefenseTypeMelee, false, false},
		{"a landed ranged shot pays", core.ProcMaskRangedAuto, core.DefenseTypeRanged, true, true},
		{"a missed ranged shot does not", core.ProcMaskRangedAuto, core.DefenseTypeRanged, false, false},
		{"a landed direct spell pays", core.ProcMaskSpellDamage, core.DefenseTypeMagic, true, true},
		{"a missed direct spell does not", core.ProcMaskSpellDamage, core.DefenseTypeMagic, false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sim, char, target := setup(t)
			spell := strike(char, c.procMask, c.defense, c.landed)
			if got := paidWithin(sim, char, target, spell, 200); got != c.pays {
				t.Errorf("200 tries on a judged target paid mana: %v, want %v", got, c.pays)
			}
		})
	}
}

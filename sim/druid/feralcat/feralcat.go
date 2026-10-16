package feralcat

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/druid"
)

func RegisterFeralCatDruid() {
	core.RegisterAgentFactory(
		proto.Player_FeralCatDruid{},
		proto.Spec_SpecFeralCatDruid,
		func(character *core.Character, options *proto.Player, _ *proto.Raid) core.Agent {
			return NewFeralCatDruid(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_FeralCatDruid)
			if !ok {
				panic("Invalid spec value for Feral Druid!")
			}
			player.Spec = playerSpec
		},
	)
}

func NewFeralCatDruid(character *core.Character, options *proto.Player) *FeralDruid {
	feralOptions := options.GetFeralCatDruid()
	selfBuffs := druid.SelfBuffs{}

	cat := &FeralDruid{
		Druid: druid.New(character, druid.Cat, selfBuffs, options.TalentsString),
	}

	cat.CannotShredTarget = feralOptions.Options.CannotShredTarget
	cat.AssumeBleedActive = feralOptions.Options.GetAssumeBleedActive()
	cat.prepullClearcasting = feralOptions.Options.GetPrepullClearcasting()

	cat.EnableEnergyBar(core.EnergyBarOptions{
		MaxComboPoints: 5,
		MaxEnergy:      100.0,
		UnitClass:      proto.Class_ClassDruid,
	})
	cat.EnableRageBar(core.RageBarOptions{BaseRageMultiplier: 1, TwoHandRage: druid.BearRageIsTwoHanded})

	cat.EnableAutoAttacks(cat, core.AutoAttackOptions{
		// Base paw weapon.
		MainHand:       cat.GetCatWeapon(),
		AutoSwingMelee: true,
		ReplaceMHSwing: cat.TryMaul,
	})

	cat.RegisterCatFormAura()
	cat.RegisterBearFormAura()

	return cat
}

type FeralDruid struct {
	*druid.Druid

	prepullClearcasting bool
}

func (cat *FeralDruid) GetDruid() *druid.Druid {
	return cat.Druid
}

// Leader of the Pack comes from the druid's own talent.
func (cat *FeralDruid) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {
	cat.Druid.AddRaidBuffs(raidBuffs)
}

func (cat *FeralDruid) AddPartyBuffs(_ *proto.PartyBuffs) {
}

func (cat *FeralDruid) Initialize() {
	if cat.prepullClearcasting {
		// Omen of Clarity procs off any druid spell, so a cat can fish for it out of combat; the
		// shift back to Cat Form takes a GCD before the pull. The proc started Omen of Clarity's ICD
		// then at the latest.
		cat.RegisterPrepullAction(-core.GCDDefault, func(sim *core.Simulation) {
			cat.ClearcastingAura.Activate(sim)
			cat.OmenOfClarityAura.Icd.Use(sim)
		})
	}
	cat.Druid.Initialize()
	cat.RegisterFeralCatSpells()
}

func (cat *FeralDruid) ApplyTalents() {
	cat.Druid.ApplyTalents()
}

func (cat *FeralDruid) Reset(sim *core.Simulation) {
	cat.Druid.Reset(sim)
	cat.Druid.ClearForm(sim)
	cat.CatFormAura.Activate(sim)
}

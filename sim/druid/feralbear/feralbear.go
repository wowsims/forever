package feralbear

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/druid"
)

func RegisterFeralBearDruid() {
	core.RegisterAgentFactory(
		proto.Player_FeralBearDruid{},
		proto.Spec_SpecFeralBearDruid,
		func(character *core.Character, options *proto.Player, _ *proto.Raid) core.Agent {
			return NewFeralBearDruid(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_FeralBearDruid)
			if !ok {
				panic("Invalid spec value for Guardian Druid!")
			}
			player.Spec = playerSpec
		},
	)
}

func NewFeralBearDruid(character *core.Character, options *proto.Player) *GuardianDruid {
	tankOptions := options.GetFeralBearDruid()
	selfBuffs := druid.SelfBuffs{}

	bear := &GuardianDruid{
		Druid:   druid.New(character, druid.DireBear, selfBuffs, options.TalentsString),
		Options: tankOptions.Options,
	}
	bear.AssumeBleedActive = tankOptions.Options.GetAssumeBleedActive()

	healingModel := options.HealingModel
	if healingModel != nil {
		if healingModel.InspirationUptime > 0.0 {
			buffs.ApplyInspiration(bear.GetCharacter(), healingModel.InspirationUptime)
		}
	}

	bear.EnableEnergyBar(core.EnergyBarOptions{
		MaxComboPoints: 5,
		MaxEnergy:      100,
		UnitClass:      proto.Class_ClassDruid,
	})
	bear.EnableRageBar(core.RageBarOptions{
		BaseRageMultiplier: 1,
		StartingRage:       tankOptions.Options.GetStartingRage(),
		TwoHandRage:        druid.BearRageIsTwoHanded,
	})
	bear.EnableAutoAttacks(bear, core.AutoAttackOptions{
		// Base paw weapon.
		MainHand:       bear.GetBearWeapon(),
		AutoSwingMelee: true,
		ReplaceMHSwing: bear.TryMaul,
	})

	bear.RegisterBearFormAura()
	bear.RegisterCatFormAura()

	return bear
}

type GuardianDruid struct {
	*druid.Druid

	Options *proto.FeralBearDruid_Options
}

func (bear *GuardianDruid) GetDruid() *druid.Druid {
	return bear.Druid
}

func (bear *GuardianDruid) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {
	bear.Druid.AddRaidBuffs(raidBuffs)
}

func (bear *GuardianDruid) ApplyTalents() {
	bear.Druid.ApplyTalents()
}

func (bear *GuardianDruid) AddPartyBuffs(_ *proto.PartyBuffs) {
}

func (bear *GuardianDruid) Initialize() {
	bear.Druid.Initialize()
	bear.RegisterFeralTankSpells()
}

func (bear *GuardianDruid) Reset(sim *core.Simulation) {
	bear.Druid.Reset(sim)
	bear.Druid.ClearForm(sim)
	bear.DireBearFormAura.Activate(sim)
	bear.Druid.PseudoStats.Stunned = false
}

func (bear *GuardianDruid) OnEncounterStart(sim *core.Simulation) {
	bear.Druid.OnEncounterStart(sim)
}

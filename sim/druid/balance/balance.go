package balance

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/druid"
)

func RegisterBalanceDruid() {
	core.RegisterAgentFactory(
		proto.Player_BalanceDruid{},
		proto.Spec_SpecBalanceDruid,
		func(character *core.Character, options *proto.Player, _ *proto.Raid) core.Agent {
			return NewBalanceDruid(character, options)
		},
		func(player *proto.Player, spec any) {
			playerSpec, ok := spec.(*proto.Player_BalanceDruid)
			if !ok {
				panic("Invalid spec value for Balance Druid!")
			}
			player.Spec = playerSpec
		},
	)
}

func NewBalanceDruid(character *core.Character, options *proto.Player) *BalanceDruid {
	balanceOptions := options.GetBalanceDruid()
	selfBuffs := druid.SelfBuffs{}

	moonkin := &BalanceDruid{
		Druid:   druid.New(character, druid.Moonkin, selfBuffs, options.TalentsString),
		Options: balanceOptions.Options,
	}
	if !moonkin.Talents.MoonkinForm {
		moonkin.StartingForm = druid.Humanoid
	}

	moonkin.SelfBuffs.InnervateTarget = &proto.UnitReference{}
	if balanceOptions.Options.ClassOptions.InnervateTarget != nil {
		moonkin.SelfBuffs.InnervateTarget = balanceOptions.Options.ClassOptions.InnervateTarget
	}

	moonkin.RegisterMoonkinFormAura()
	moonkin.RegisterMoonkinFormSpell()

	return moonkin
}

type BalanceDruid struct {
	*druid.Druid
	Options *proto.BalanceDruid_Options
}

func (moonkin *BalanceDruid) GetDruid() *druid.Druid {
	return moonkin.Druid
}

func (moonkin *BalanceDruid) Initialize() {
	moonkin.Druid.Initialize()
	moonkin.Druid.RegisterBalanceSpells()
}

func (moonkin *BalanceDruid) ApplyTalents() {
	moonkin.Druid.ApplyTalents()
}

func (moonkin *BalanceDruid) Reset(sim *core.Simulation) {
	moonkin.Druid.Reset(sim)
	if moonkin.MoonkinFormAura != nil {
		moonkin.MoonkinFormAura.Activate(sim)
	}
}

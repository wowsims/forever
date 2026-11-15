package elemental

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/shaman"
)

func RegisterElementalShaman() {
	core.RegisterAgentFactory(
		proto.Player_ElementalShaman{},
		proto.Spec_SpecElementalShaman,
		func(character *core.Character, options *proto.Player, _ *proto.Raid) core.Agent {
			return NewElementalShaman(character, options)
		},
		func(player *proto.Player, spec any) {
			playerSpec, ok := spec.(*proto.Player_ElementalShaman)
			if !ok {
				panic("Invalid spec value for Elemental Shaman!")
			}
			player.Spec = playerSpec
		},
	)
}

func NewElementalShaman(character *core.Character, options *proto.Player) *ElementalShaman {
	eleOptions := options.GetElementalShaman().Options

	selfBuffs := shaman.SelfBuffs{
		ShieldProcrate: eleOptions.ClassOptions.ShieldProcrate,
		ImbueMH:        eleOptions.ClassOptions.ImbueMh,
		ImbueMHSwap:    eleOptions.ClassOptions.ImbueMhSwap,
	}

	ele := &ElementalShaman{
		Shaman: shaman.NewShaman(character, options.TalentsString, selfBuffs),
	}

	//Some spells use weapon damage (Unleash Wind, ...)
	ele.EnableAutoAttacks(ele, core.AutoAttackOptions{
		MainHand:       ele.WeaponFromMainHand(),
		AutoSwingMelee: false,
	})

	return ele
}

func (eleShaman *ElementalShaman) Initialize() {
	eleShaman.Shaman.Initialize()
}

func (ele *ElementalShaman) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {
	ele.Shaman.AddRaidBuffs(raidBuffs)
}

func (ele *ElementalShaman) ApplyTalents() {
	ele.Shaman.ApplyTalents()
}

type ElementalShaman struct {
	*shaman.Shaman
}

func (eleShaman *ElementalShaman) GetShaman() *shaman.Shaman {
	return eleShaman.Shaman
}

func (eleShaman *ElementalShaman) Reset(sim *core.Simulation) {
	eleShaman.Shaman.Reset(sim)
}

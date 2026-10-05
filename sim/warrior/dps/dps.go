package dps

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/warrior"
)

func RegisterDpsWarrior() {
	core.RegisterAgentFactory(
		proto.Player_DpsWarrior{},
		proto.Spec_SpecDpsWarrior,
		func(character *core.Character, options *proto.Player, _ *proto.Raid) core.Agent {
			return NewDpsWarrior(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_DpsWarrior)
			if !ok {
				panic("Invalid spec value for Dps Warrior!")
			}
			player.Spec = playerSpec
		},
	)
}

type DpsWarrior struct {
	*warrior.Warrior

	Options *proto.DpsWarrior_Options

	BloodsurgeAura  *core.Aura
	MeatCleaverAura *core.Aura
}

func (war *DpsWarrior) ApplyTalents() {
	war.Warrior.ApplyTalents()
}

func NewDpsWarrior(character *core.Character, options *proto.Player) *DpsWarrior {
	dpsOptions := options.GetDpsWarrior().Options
	classOptions := dpsOptions.ClassOptions

	war := &DpsWarrior{
		Warrior: warrior.NewWarrior(character, dpsOptions.ClassOptions, options.TalentsString, warrior.WarriorInputs{
			UseBattleShout: classOptions.UseBattleShout,
			DefaultStance:  classOptions.DefaultStance,
			StartingRage:   classOptions.StartingRage,
			QueueDelay:     classOptions.QueueDelay,
			StanceSnapshot: classOptions.StanceSnapshot,
			HasBsT2:        classOptions.HasBsT2,
		}),
		Options: dpsOptions,
	}

	return war
}

func (war *DpsWarrior) GetWarrior() *warrior.Warrior {
	return war.Warrior
}

func (war *DpsWarrior) Initialize() {
	war.Warrior.Initialize()
}

func (war *DpsWarrior) Reset(sim *core.Simulation) {
	war.Warrior.Reset(sim)
}

func (war *DpsWarrior) OnEncounterStart(sim *core.Simulation) {
	war.Warrior.OnEncounterStart(sim)
}

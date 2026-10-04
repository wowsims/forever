// Package testagent is a character with no class code behind it, for the tests of code every class
// shares. A whole sim needs an agent for each player, and borrowing a real class's agent ties those
// tests to that class's spell data, so a class still to move onto a new game build would break them.
package testagent

import (
	"github.com/wowsims/forever/sim/core"
	_ "github.com/wowsims/forever/sim/core/buffs" // a whole sim needs the raid buffs registered, which a class package would import
	"github.com/wowsims/forever/sim/core/proto"
)

type Agent struct {
	core.Character
}

func (a *Agent) GetCharacter() *core.Character       { return &a.Character }
func (a *Agent) Initialize()                         {}
func (a *Agent) ApplyTalents()                       {}
func (a *Agent) Reset(_ *core.Simulation)            {}
func (a *Agent) OnGCDReady(_ *core.Simulation)       {}
func (a *Agent) OnEncounterStart(_ *core.Simulation) {}

// Swings the weapons the character wears, the way a melee class's constructor sets up its auto attacks.
func (a *Agent) EnableMelee() {
	a.EnableAutoAttacks(a, core.AutoAttackOptions{
		MainHand:       a.WeaponFromMainHand(),
		OffHand:        a.WeaponFromOffHand(),
		AutoSwingMelee: true,
	})
}

// Registers the agent under a spec, so a player of that spec in a test's raid is built as the agent.
// setup, when given, runs on each new agent: what the borrowed class would have enabled, such as melee
// swings or parrying. The spec must be one no class package in the same test binary registers.
func Register(emptyOptions any, spec proto.Spec, specSetter core.SpecSetter, setup func(*Agent)) {
	core.RegisterAgentFactory(emptyOptions, spec, func(character *core.Character, _ *proto.Player, _ *proto.Raid) core.Agent {
		agent := &Agent{Character: *character}
		if setup != nil {
			setup(agent)
		}
		return agent
	}, specSetter)
}

// Registers the agent as the DPS warrior's spec: a player built from a warrior's proto, gear and
// consumables, swinging its weapons, with none of the warrior's own code.
func RegisterAsDpsWarrior() {
	Register(proto.Player_DpsWarrior{}, proto.Spec_SpecDpsWarrior, func(player *proto.Player, spec any) {
		player.Spec = spec.(*proto.Player_DpsWarrior)
	}, (*Agent).EnableMelee)
}

// Registers the agent as the protection warrior's spec, the way RegisterAsDpsWarrior does.
func RegisterAsProtectionWarrior(setup func(*Agent)) {
	Register(proto.Player_ProtectionWarrior{}, proto.Spec_SpecProtectionWarrior, func(player *proto.Player, spec any) {
		player.Spec = spec.(*proto.Player_ProtectionWarrior)
	}, setup)
}

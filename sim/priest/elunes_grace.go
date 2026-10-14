package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerElunesGrace(_ int32, rank *spelldata.Spell) {
	aura := priest.RegisterAura(spelldata.AuraConfig(rank))
	spelldata.ParseEffects(&priest.Character, aura, rank)

	// Each melee or ranged attack that misses the priest spends a charge. A miss is an outcome no
	// proc mask states and deals no damage, so both are set here.
	trigger := spelldata.ProcTrigger(&priest.Character, rank,
		func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			aura.RemoveStack(sim)
		})
	trigger.Outcome = core.OutcomeMiss
	trigger.RequireDamageDealt = false
	aura.AttachProcTrigger(trigger)

	priest.registerSelfBuffSpell(rank, aura)
}

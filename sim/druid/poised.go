package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Poised 1328829, the Feral Combat passive. Its subtext reads Passive, not Rank N, so the store
// carries it as an extra spell.
const poisedSpellID = 1328829

// "While at maximum Energy, and out of combat, reduce the Energy cost of your next ability." The
// passive is a dummy, and the buff its tooltip names holds the cost mod, the charge and the cat
// abilities that spend it. The druid starts each iteration out of combat at full energy, and nothing
// takes it out of combat again, so the buff is up from the start until the first of those casts.
func (druid *Druid) registerPoised() {
	buffRow := spelldata.MustFind(poisedSpellID).Refs()[0]
	config := spelldata.AuraConfig(buffRow)
	config.OnReset = func(aura *core.Aura, sim *core.Simulation) {
		aura.Activate(sim)
		aura.SetStacks(sim, aura.MaxStacks)
	}
	buff := druid.RegisterAura(config)
	spelldata.ParseEffects(&druid.Character, buff, buffRow)
	// Cower carries no proc mask in the sim, so the decoded melee mask would never hear it and the
	// discount would carry over to the next ability. The buff's class flags already name the eight
	// abilities that spend it.
	spelldata.AttachChargeSpender(&druid.Character, buff, buffRow, "Poised - Spend",
		func(_ *core.Character, trigger *core.ProcTrigger) {
			trigger.ProcMask = core.ProcMaskUnknown
		})
}

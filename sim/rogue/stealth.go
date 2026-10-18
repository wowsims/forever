package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var stealthRank = spellData.Stealth.Highest()

func (rogue *Rogue) registerStealthAura() {
	config := spelldata.SpellConfig(&rogue.Unit, stealthRank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagPrepullOnly))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.StealthAura.Activate(sim)
	}

	rogue.StealthAura = rogue.RegisterAura(spelldata.AuraConfig(stealthRank)).
		ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) {
			rogue.ShapeshiftForm = stealthRank.ShapeshiftForm()
		}).
		ApplyOnExpire(func(_ *core.Aura, _ *core.Simulation) {
			rogue.ShapeshiftForm = 0
		})

	spelldata.ParseEffects(&rogue.Character, rogue.StealthAura, stealthRank)

	rogue.RegisterSpell(config)
}

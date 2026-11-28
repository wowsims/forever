package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var recklessnessRank = spellData.Recklessness.Highest()

func (warrior *Warrior) registerRecklessness() {
	aura := warrior.RegisterAura(spelldata.AuraConfig(recklessnessRank))
	spelldata.ParseEffects(&warrior.Character, aura, recklessnessRank)

	// Grants immunity to Fear effects, which the spell states as A_MECHANIC_IMMUNITY and the parse
	// skips.
	aura.AttachFearImmunity()

	config := spelldata.SpellConfig(&warrior.Unit, recklessnessRank,
		spelldata.Flags(core.SpellFlagAPL|core.SpellFlagCastWhileIncapacitated))

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		aura.Activate(sim)
	}

	config.RelatedSelfBuff = aura

	spell := warrior.RegisterSpell(config)

	warrior.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeDPS,
	})
}

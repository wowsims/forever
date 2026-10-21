package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var arcanePowerRank = spellData.ArcanePower.Highest()

// Arcane Power
// https://www.wowhead.com/forever/spell=12042
//
// "For the next 15 sec, your spells deal 30% more damage while costing 30% more mana to cast."
func (mage *Mage) registerArcanePowerSpell() {
	mage.ArcanePowerAura = mage.RegisterAura(spelldata.AuraConfig(arcanePowerRank))

	// Effect 1's SPELLMOD_DAMAGE raises the direct hits and effect 3's SPELLMOD_DOT the ticks.
	spelldata.ParseEffects(&mage.Character, mage.ArcanePowerAura, arcanePowerRank)

	config := spelldata.SpellConfig(&mage.Unit, arcanePowerRank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagNoOnCastComplete))
	config.ClassSpellMask = MageSpellArcanePower

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		mage.ArcanePowerAura.Activate(sim)
	}

	config.RelatedSelfBuff = mage.ArcanePowerAura

	spell := mage.RegisterSpell(config)

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeDPS,
	})
}

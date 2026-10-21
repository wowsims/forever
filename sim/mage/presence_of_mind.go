package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var presenceOfMindRank = spellData.PresenceOfMind.Highest()

// Presence of Mind
// https://www.wowhead.com/forever/spell=12043
//
// "When activated, your next Mage spell with a casting time less than 10 sec becomes an instant
// cast spell." The row's -100% cast time names the spells it reaches, and the one charge is
// spent by the first of them cast.
func (mage *Mage) registerPresenceOfMindSpell() {
	castTimeEffect := presenceOfMindRank.EffectN(1)

	var pomSpell *core.Spell

	config := spelldata.AuraConfig(presenceOfMindRank)

	// The row is Disabled While Active (Attr0 0x2000000): the cooldown starts when the charge is spent.
	config.OnExpire = func(aura *core.Aura, sim *core.Simulation) {
		pomSpell.CD.Use(sim)
		mage.UpdateMajorCooldowns()
	}

	config.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if spell.MatchesFlags(castTimeEffect.ClassFlags) && spell.DefaultCast.CastTime > 0 {
			aura.Deactivate(sim)
		}
	}

	mage.PresenceOfMindAura = mage.RegisterAura(config)
	spelldata.ParseEffects(&mage.Character, mage.PresenceOfMindAura, presenceOfMindRank, spelldata.IgnoreStacks())

	spellConfig := spelldata.SpellConfig(&mage.Unit, presenceOfMindRank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagNoOnCastComplete))
	spellConfig.ClassSpellMask = MageSpellPresenceOfMind

	spellConfig.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return !mage.PresenceOfMindAura.IsActive()
	}

	spellConfig.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		mage.PresenceOfMindAura.Activate(sim)
	}

	spellConfig.RelatedSelfBuff = mage.PresenceOfMindAura

	pomSpell = mage.RegisterSpell(spellConfig)

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: pomSpell,
		Type:  core.CooldownTypeDPS,
	})
}

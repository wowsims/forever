package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var iceBlockRank = spellData.IceBlock.Highest()

// Ice Block
// https://www.wowhead.com/forever/spell=11958
//
// "You become encased in a block of ice, protecting you from all physical attacks and spells for 10
// sec, but during that time you cannot attack, move, or cast spells."
//
// The row's stun (effect 1) is core's stun aura, which stops the mage casting; its two school
// immunities (effects 2 and 3, every school between them) zero the damage taken while it is up.
// Being core's stun, its duration also takes the mage's StunDurationMultiplier, which nothing a
// mage wears changes.
func (mage *Mage) registerIceBlockSpell() {
	iceBlockAura := mage.RegisterStunAura(iceBlockRank.Name, core.ActionID{SpellID: iceBlockRank.ID}, iceBlockRank.Duration())

	mage.AddDynamicDamageTakenModifier(func(_ *core.Simulation, _ *core.Spell, result *core.SpellResult, _ bool) {
		if iceBlockAura.IsActive() {
			result.Damage = 0
		}
	})

	config := spelldata.SpellConfig(&mage.Unit, iceBlockRank, spelldata.Flags(core.SpellFlagAPL))
	config.ClassSpellMask = MageSpellIceBlock
	config.RelatedSelfBuff = iceBlockAura

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		iceBlockAura.Activate(sim)
	}

	mage.RegisterSpell(config)
}

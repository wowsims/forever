package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Divine Favor (talent)
// https://www.wowhead.com/forever/spell=20216
//
// When activated, gives your next Flash of Light, Holy Light, or Holy Shock spell a 100% critical
// effect chance.
func (paladin *Paladin) registerDivineFavor() {
	rank := spellData.DivineFavor.Highest()

	// The crit and the one charge the next Holy Light, Flash of Light or Holy Shock spends are the
	// row's.
	divineFavorAura := paladin.RegisterAura(spelldata.AuraConfig(rank, spelldata.Label("Divine Favor"+paladin.Label)))
	spelldata.ParseEffects(&paladin.Character, divineFavorAura, rank)
	// The cast spends the charge whether or not it lands, so a resisted Holy Shock, or one a Light's
	// Vigil turns into its strike, spends it too.
	spelldata.AttachChargeSpender(&paladin.Character, divineFavorAura, rank, "Divine Favor Charge"+paladin.Label,
		func(_ *core.Character, trigger *core.ProcTrigger) {
			trigger.Callback = core.CallbackOnApplyEffects
			trigger.Outcome = core.OutcomeEmpty
		})

	config := spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Flags(core.SpellFlagAPL))
	config.ClassSpellMask = SpellMaskDivineFavor
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.RelatedSelfBuff.Activate(sim)
		spell.RelatedSelfBuff.SetStacks(sim, spell.RelatedSelfBuff.MaxStacks)
	}
	config.RelatedSelfBuff = divineFavorAura
	divineFavor := paladin.RegisterSpell(config)

	paladin.AddMajorCooldown(core.MajorCooldown{
		Spell: divineFavor,
		Type:  core.CooldownTypeDPS,
	})
}

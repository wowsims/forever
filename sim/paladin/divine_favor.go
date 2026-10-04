package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// Divine Favor (talent)
// https://www.wowhead.com/forever/spell=20216
//
// When activated, gives your next Flash of Light, Holy Light, or Holy Shock spell a 100% critical
// effect chance.
func (paladin *Paladin) registerDivineFavor() {
	rank := spellData.DivineFavor.Highest()
	actionID := core.ActionID{SpellID: rank.ID}

	var divineFavorAura *core.Aura
	divineFavorAura = paladin.RegisterAura(core.Aura{
		Label:    "Divine Favor" + paladin.Label,
		ActionID: actionID,
		Duration: core.NeverExpires,
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		ClassMask:  SpellMaskHealingSpells | SpellMaskHolyShock,
		FloatValue: rank.Effect(dbcenums.A_ADD_FLAT_MODIFIER, int32(dbcenums.SPELLMOD_CRITICAL_CHANCE)).Average(),
	}).AttachProcTrigger(core.ProcTrigger{
		CanProcFromProcs:   true, // 20216 carries the bit.
		Callback:           core.CallbackOnCastComplete,
		ClassSpellMask:     SpellMaskHealingSpells | SpellMaskHolyShock,
		TriggerImmediately: true,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			divineFavorAura.Deactivate(sim)
		},
	})

	divineFavor := paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    rank.SpellSchool(),
		DefenseType:    rank.DefenseTypeCore(),
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskDivineFavor,

		ManaCost: manaCost(rank),
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: cooldown(rank),
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.RelatedSelfBuff.Activate(sim)
		},

		RelatedSelfBuff: divineFavorAura,
	})

	paladin.AddMajorCooldown(core.MajorCooldown{
		Spell: divineFavor,
		Type:  core.CooldownTypeDPS,
	})
}

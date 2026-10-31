package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var HolyShieldRankMap = spellData.HolyShield

// Holy Shield (talent)
// https://www.wowhead.com/forever/spell=20928
//
// Increases chance to block by 30% for 10 sec, and deals 221 Holy damage for each attack blocked
// while active. Damage caused by Holy Shield causes 20% additional threat.
func (paladin *Paladin) registerHolyShield(_ int32, rank *spelldata.Spell) {
	actionID := core.ActionID{SpellID: rank.ID}
	damageEffect := rank.Effect(dbcenums.A_PROC_TRIGGER_DAMAGE, 0)
	damage := damageEffect.Mean()
	blockPercent := rank.Effect(dbcenums.A_MOD_BLOCK_PERCENT, 0).BaseValue()
	charges := int32(rank.ProcCharges)

	// The per-block damage has no row of its own: it is the shield's own proc-trigger-damage effect.
	procConfig := spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Magic(core.ProcMaskEmpty), spelldata.Proc(),
		spelldata.Tag(2), spelldata.Flags(core.SpellFlagBinary))
	procConfig.ClassSpellMask = SpellMaskHolyShieldProc
	procConfig.Flags &^= core.SpellFlagHelpful // The row's first effect is the block chance on the paladin.
	procConfig.BonusCoefficient = damageEffect.Coeff()
	procConfig.ThreatMultiplier = 1.2 // The tooltip's 20% additional threat.
	procConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMagicHit)
	}
	procSpell := paladin.RegisterSpell(procConfig)

	var holyShieldAura *core.Aura
	holyShieldAura = paladin.RegisterAura(core.Aura{
		Label:     rank.RankedLabel(&paladin.Unit),
		ActionID:  actionID,
		Duration:  rank.Duration(),
		MaxStacks: charges,
	}).AttachProcTrigger(core.ProcTrigger{
		Callback:           core.CallbackOnSpellHitTaken,
		Outcome:            core.OutcomeBlock,
		TriggerImmediately: true,
		Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			procSpell.Cast(sim, spell.Unit)
			holyShieldAura.RemoveStack(sim)
		},
	}).AttachStatBuff(stats.BlockPercent, blockPercent)

	// Shieldsworn: a Holy Strike that lands deals the shield's damage and spends no charge.
	if paladin.Talents.Shieldsworn {
		holyShieldAura.AttachProcTrigger(core.ProcTrigger{
			Callback:           core.CallbackOnSpellHitDealt,
			ClassSpellMask:     SpellMaskHolyStrike,
			Outcome:            core.OutcomeLanded,
			TriggerImmediately: true,
			Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
				procSpell.Cast(sim, result.Target)
			},
		})
	}

	// Steadfast Libram: more block value while the shield is up.
	if paladin.holyShieldBlockValueMultiplier != 1 {
		holyShieldAura.AttachMultiplicativePseudoStatBuff(&paladin.PseudoStats.BlockValueMultiplier, paladin.holyShieldBlockValueMultiplier)
	}

	// The row requires a shield, which the cast requirement checks.
	config := spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ClassSpellMask = SpellMaskHolyShield
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		holyShieldAura.Activate(sim)
		holyShieldAura.SetStacks(sim, charges)
	}
	config.RelatedSelfBuff = holyShieldAura
	paladin.RegisterSpell(config)
}

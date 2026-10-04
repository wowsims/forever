package paladin

import (
	"fmt"

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
// while active. Damage caused by Holy Shield causes 20% additional threat. Each block expends a
// charge. 4 charges.
//
// The per-block damage and its spell power share sit on the proc-trigger-damage aura effect,
// beside the block chance on an aura of its own.
func (paladin *Paladin) registerHolyShield(_ int32, rank *spelldata.Spell) {
	actionID := core.ActionID{SpellID: rank.ID}
	damageEffect := rank.Effect(dbcenums.A_PROC_TRIGGER_DAMAGE, 0)
	damage := damageEffect.Mean()
	blockPercent := rank.Effect(dbcenums.A_MOD_BLOCK_PERCENT, 0).BaseValue()
	charges := int32(rank.ProcCharges)

	procSpell := paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID.WithTag(2),
		SpellSchool:    core.SpellSchoolHoly,
		DefenseType:    core.DefenseTypeMagic,
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell | core.SpellFlagBinary,
		ClassSpellMask: SpellMaskHolyShieldProc,

		DamageMultiplier: 1,
		ThreatMultiplier: 1.2,
		BonusCoefficient: damageEffect.Coeff(),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMagicHit)
		},
	})

	var holyShieldAura *core.Aura
	holyShieldAura = paladin.RegisterAura(core.Aura{
		Label:     fmt.Sprintf("Holy Shield%s Rank %d", paladin.Label, rank.RankNumber()),
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

	// Steadfast Libram: more block value while the shield is up.
	if paladin.holyShieldBlockValueMultiplier != 1 {
		holyShieldAura.AttachMultiplicativePseudoStatBuff(&paladin.PseudoStats.BlockValueMultiplier, paladin.holyShieldBlockValueMultiplier)
	}

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    rank.SpellSchool(),
		DefenseType:    rank.DefenseTypeCore(),
		ProcMask:       core.ProcMaskEmpty,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskHolyShield,
		Rank:           rank.RankNumber(),

		ManaCost: manaCost(rank),
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: rank.GCD(),
			},
			CD: core.Cooldown{
				Timer:    paladin.sharedTimer(&paladin.holyShieldTimer),
				Duration: cooldown(rank),
			},
		},

		ExtraCastCondition: func(_ *core.Simulation, _ *core.Unit) bool {
			return paladin.PseudoStats.CanBlock
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			holyShieldAura.Activate(sim)
			holyShieldAura.SetStacks(sim, charges)
		},

		RelatedSelfBuff: holyShieldAura,
	})
}

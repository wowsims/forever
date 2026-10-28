package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var ambushRank = spellData.Ambush.Highest()

func (rogue *Rogue) registerAmbushSpell() {
	ambushBaseDamage := ambushRank.DamageEffect().Average()
	comboPointGen := ambushRank.EnergizeEffect().Average()
	config := spelldata.SpellConfig(&rogue.Unit, ambushRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.Flags |= SpellFlagBuilder | SpellFlagColdBlooded
	config.ClassSpellMask = RogueSpellAmbush
	config.DamageMultiplier = ambushRank.EffectN(2).Percent()
	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return rogue.HasDagger(core.MainHand) && !rogue.PseudoStats.InFrontOfTarget && (rogue.IsStealthed() || rogue.CutthroatAura.IsActive())
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		baseDamage := ambushBaseDamage +
			spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))

		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

		if result.Landed() {
			rogue.AddComboPoints(sim, int32(comboPointGen), spell.ComboPointMetrics())
		} else {
			spell.IssueRefund(sim)
		}
	}

	rogue.Ambush = rogue.RegisterSpell(config)
}

package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var garroteRank = spellData.Garrote.Highest()

func (rogue *Rogue) registerGarrote() {
	tick := garroteRank.PeriodicEffect()
	apScaling := 0.03 // Estimated from in-game tests.
	comboPointGen := garroteRank.EnergizeEffect().Average()

	config := spelldata.SpellConfig(&rogue.Unit, garroteRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.Flags |= SpellFlagBuilder
	config.ClassSpellMask = RogueSpellGarrote
	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return (!rogue.PseudoStats.InFrontOfTarget || rogue.Talents.DirtyDeeds > 0) &&
			(rogue.IsStealthed() || rogue.GhostlyStrikeAura != nil && rogue.GhostlyStrikeAura.IsActive())
	}

	config.Dot = spelldata.DotConfig(garroteRank, tick)
	if rogue.Talents.BloodSplatter > 0 {
		config.Dot.NumberOfTicks += 3
	}
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, tick.Average()+dot.Spell.MeleeAttackPower(target)*apScaling, garroteRank.TickOutcome(dot))
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		if rogue.GhostlyStrikeAura != nil {
			rogue.GhostlyStrikeAura.Deactivate(sim)
		}
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHitNoHitCounter)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
			rogue.AddComboPoints(sim, int32(comboPointGen), spell.ComboPointMetrics())
		} else {
			spell.IssueRefund(sim)
		}
	}

	rogue.Garrote = rogue.RegisterSpell(config)
}

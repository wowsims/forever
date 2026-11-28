package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var deathMarkRank = spellData.DeathMark.Highest()

// The mark's rates are stated only in the tooltip: description variables 1021 give $ratio and $maxAP, and the text of
// rank 4 (1325758), the rank the sim registers, caps the payout at
// ((156 + 245 per point + $maxAP*AP per point) * ($mcrit/100 + 1)) * $coup.
const (
	deathMarkStoredPerPoint = 0.12 // $ratio: 12% of the damage per combo point.
	deathMarkCapAPPerPoint  = 0.04 // $maxAP.
	deathMarkCapBase        = 156.0
	deathMarkCapPerPoint    = 245.0
)

// Stored damage in EV-0192. The stored damage already carries every modifier and the target's armor, so
// the payout takes neither again.
func (rogue *Rogue) registerDeathMark() {
	if !rogue.Talents.DeathMark {
		return
	}

	// Coup de Grace's second effect is a SPELLMOD_DAMAGE on Death Mark's class flags, which the sim applies to direct
	// hits only, and the payout is a tick, so it is read here. The tooltip's $coup reads the talent's first effect,
	// which states the same values.
	coupDeGrace := spellData.CoupDeGrace.EffectAt(2).MultiplierAt(rogue.Talents.CoupDeGrace)
	var comboPoints int32

	config := spelldata.SpellConfig(&rogue.Unit, deathMarkRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ExtraCastCondition = rogue.FinisherExtraCastCondition
	config.Flags |= core.SpellFlagIgnoreModifiers | core.SpellFlagIgnoreResists

	currentCap := func(target *core.Unit) float64 {
		points := float64(comboPoints)
		// $mcrit is the melee crit chance the character sheet shows, before the target's level takes any.
		critFactor := 1 + rogue.GetStat(stats.PhysicalCritPercent)/100
		return (deathMarkCapBase + deathMarkCapPerPoint*points + deathMarkCapAPPerPoint*points*rogue.DeathMark.MeleeAttackPower(target)) *
			critFactor * coupDeGrace
	}

	store := func(aura *core.Aura, _ *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		if spell.Unit != &rogue.Unit || spell == rogue.DeathMark || spell.SpellSchool != core.SpellSchoolPhysical {
			return
		}
		rate := deathMarkStoredPerPoint * float64(comboPoints) * coupDeGrace
		if spell.Matches(RogueSpellGarrote | RogueSpellRupture) {
			rate *= 2
		}
		dot := rogue.DeathMark.Dot(aura.Unit)
		limit := currentCap(aura.Unit)
		if dot.SnapshotBaseDamage < limit {
			dot.SnapshotBaseDamage = min(dot.SnapshotBaseDamage+result.Damage*rate, limit)
		}
	}

	config.Dot = spelldata.DotConfig(deathMarkRank, deathMarkRank.PeriodicEffect())
	config.Dot.Aura.OnSpellHitTaken = store
	config.Dot.Aura.OnPeriodicDamageTaken = store
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, dot.SnapshotBaseDamage, deathMarkRank.TickOutcome(dot))
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHitNoHitCounter)
		if result.Landed() {
			comboPoints = rogue.ComboPoints(target)
			dot := spell.Dot(target)
			dot.SnapshotBaseDamage = 0
			dot.Apply(sim)
			rogue.ApplyFinisher(sim, target, spell)
		} else {
			spell.IssueRefund(sim)
		}
		spell.DealOutcome(sim, result)
	}

	config.MetricSplits = 6
	config.Cast.ModifyCast = rogue.setFinisherMetricsSplit

	rogue.DeathMark = rogue.RegisterSpell(config)
}

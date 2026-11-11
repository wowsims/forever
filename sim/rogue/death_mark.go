package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var deathMarkRank = spellData.DeathMark.Highest()

// The mark's rates are stated only in the tooltip: description variables 1021 give $ratio and $maxAP, and the text of
// rank 4 (1325758), the rank the sim registers, caps the payout at
// ((156 + 245 per point + $maxAP*AP per point) * ($mcrit/100 + 1)) * $coup. The store carries neither.
const (
	deathMarkStoredPerPoint = 0.12 // $ratio: 12% of the damage per combo point.
	deathMarkCapAPPerPoint  = 0.04 // $maxAP.
	deathMarkCapBase        = 156.0
	deathMarkCapPerPoint    = 245.0
)

// The mark stores a share of the damage the rogue's attacks, poisons and ticks deal its target, twice the share for
// Garrote and Rupture, and deals what it stored when it ends, up to a cap that its combo points, attack power and
// melee crit chance raise. The stored damage already carries every modifier, so the payout takes none again.
func (rogue *Rogue) registerDeathMark() {
	if !rogue.Talents.DeathMark {
		return
	}

	// Coup de Grace raises the storage rate and the cap, each on its own: below the cap the stored amount carries it,
	// and the cap is raised by it too. Its second effect is a SPELLMOD_DAMAGE on Death Mark's class flags, which the
	// sim applies to direct hits only, and the payout is a tick, so it is read here. The tooltip's $coup reads the
	// talent's first effect, which states the same values.
	coupDeGrace := spellData.CoupDeGrace.EffectAt(2).MultiplierAt(rogue.Talents.CoupDeGrace)
	var comboPoints int32

	config := spelldata.SpellConfig(&rogue.Unit, deathMarkRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ExtraCastCondition = rogue.FinisherExtraCastCondition
	config.Flags |= core.SpellFlagIgnoreModifiers

	store := func(aura *core.Aura, _ *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		if spell.Unit != &rogue.Unit || spell == rogue.DeathMark {
			return
		}
		rate := deathMarkStoredPerPoint * float64(comboPoints) * coupDeGrace
		if spell.Matches(RogueSpellGarrote | RogueSpellRupture) {
			rate *= 2
		}
		rogue.DeathMark.Dot(aura.Unit).SnapshotBaseDamage += result.Damage * rate
	}

	config.Dot = spelldata.DotConfig(deathMarkRank, deathMarkRank.PeriodicEffect())
	config.Dot.Aura.OnSpellHitTaken = store
	config.Dot.Aura.OnPeriodicDamageTaken = store
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		points := float64(comboPoints)
		// $mcrit is the melee crit chance the character sheet shows, before the target's level takes any.
		critFactor := 1 + rogue.GetStat(stats.PhysicalCritPercent)/100
		limit := (deathMarkCapBase + deathMarkCapPerPoint*points + deathMarkCapAPPerPoint*points*dot.Spell.MeleeAttackPower(target)) *
			critFactor * coupDeGrace
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, min(dot.SnapshotBaseDamage, limit), deathMarkRank.TickOutcome(dot))
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

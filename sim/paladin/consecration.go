package paladin

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var ConsecrationRankMap = spellData.Consecration

// The mark the ground lays on the first enemies inside it, one aura per enemy: the bonus tick's dummy,
// and Consecrated Ground's Holy damage from caster mod, whose points the server sets from the talent.
// The area trigger keeps it up while the ground stands, so it outlives the last tick by the row's
// 0.6 sec.
func (paladin *Paladin) registerConsecrationMark() {
	mark := spellData.ConsecratedGroundTriggered.Highest()

	paladin.consecrationMarks = paladin.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		aura := target.GetOrRegisterAura(spelldata.AuraConfig(mark, spelldata.Label("Consecration Mark"+paladin.Label)))
		if paladin.consecratedGroundPoints != nil {
			spelldata.ParseEffects(nil, aura, mark, spelldata.FromCaster(&paladin.Unit, 0, 1), spelldata.ScaledBy(paladin.consecratedGroundPoints))
		}
		return aura
	})
}

// Consecration
// https://www.wowhead.com/forever/spell=20924
//
// Consecrates the land beneath the Paladin, doing 96 Holy damage over 8 sec to enemies who enter the
// area. The first 4 enemies who enter the area will take an additional 216 damage over 8 sec.
//
// The ground ticks once a period after the cast, none at the cast. Each tick rolls spell hit and crit
// on its own, though no row in the ladder states Periodic Can Crit.
func (paladin *Paladin) registerConsecration(n int32, rank *spelldata.Spell) {
	tickSpell := spellData.ConsecrationTriggered.Rank(n)
	tick := tickSpell.EffectN(1)
	bonus := tickSpell.EffectN(2)

	dummy := rank.Effect(dbcenums.A_PERIODIC_DUMMY, 0)
	bonusTargets := int(dummy.Average())
	tickLength := dummy.Period()
	numberOfTicks := int32(rank.Duration() / tickLength)
	markTail := spellData.ConsecratedGroundTriggered.Highest().Duration()

	// The ground marks the enemies inside it, up to the bonus targets, and keeps every mark up for the
	// time it has left and the mark's tail: at the cast, and at each tick for an enemy that came later.
	markTargets := func(sim *core.Simulation, remaining time.Duration) {
		marked := 0
		for _, target := range sim.Encounter.ActiveTargetUnits {
			mark := paladin.consecrationMarks.Get(target)
			if !mark.IsActive() && marked >= bonusTargets {
				continue
			}
			mark.Duration = remaining + markTail
			mark.Activate(sim)
			marked++
		}
	}

	config := spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = SpellMaskConsecration
	// The row states no range: the area is the dummy's 8 yd radius around the paladin, so the target
	// has to stand inside it.
	config.MaxRange = float64(rank.EffectN(1).RadiusMax)

	config.Dot = core.DotConfig{
		IsAOE: true,
		Aura: core.Aura{
			ActionID: core.ActionID{SpellID: rank.ID},
			Label:    rank.RankedLabel(&paladin.Unit),
		},
		NumberOfTicks:    numberOfTicks,
		TickLength:       tickLength,
		BonusCoefficient: tick.Coeff(),
		// The bonus scales on its own coefficient, so it is added to the base damage here rather than
		// through the dot's, which is the base tick's and states none: the spell power and the Holy
		// damage the target takes extra both sit inside the tick's base. The marked enemies take it.
		OnTick: func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
			markTargets(sim, dot.RemainingDuration(sim))
			for _, target := range sim.Encounter.ActiveTargetUnits {
				damage := tick.Average()
				if paladin.consecrationMarks.Get(target).IsActive() {
					damage += bonus.Average() + bonus.Coeff()*(dot.Spell.BonusDamage(dot.Spell.Unit.AttackTables[target.UnitIndex])+
						target.PseudoStats.SchoolBonusSpellDamage[stats.SchoolIndexHoly])
				}
				dot.Spell.CalcAndDealPeriodicDamage(sim, target, damage, dot.Spell.OutcomeTickMagicHitAndCrit)
			}
		},
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		// The cast rolls its own hit only so that procs on a resist can fire; the ground is laid whether
		// it lands or not. The roll counts nothing, so the row's Miss % is the ticks' own.
		spell.CalcAndDealOutcome(sim, target, spell.OutcomeTickMagicHitNoHitCounter)

		spell.AOEDot().Apply(sim)
		markTargets(sim, rank.Duration())
	}

	paladin.RegisterSpell(config)
}

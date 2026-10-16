package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var ConsecrationRankMap = spellData.Consecration

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
		// through the dot's, which is the base tick's. The periodic dummy picks the bonus targets from its
		// first tick on, and Consecrated Ground marks them on that first tick.
		OnTick: func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
			for i, target := range sim.Encounter.ActiveTargetUnits {
				damage := tick.Average()
				if i < bonusTargets {
					damage += bonus.Average() + bonus.Coeff()*dot.Spell.BonusDamage(dot.Spell.Unit.AttackTables[target.UnitIndex])
					if dot.TickCount() == 1 && paladin.consecratedGroundAuras != nil {
						paladin.consecratedGroundAuras.Get(target).Activate(sim)
					}
				}
				dot.Spell.CalcAndDealPeriodicDamage(sim, target, damage, dot.Spell.OutcomeTickMagicHitAndCrit)
			}
		},
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		// The cast rolls its own hit only so that procs on a resist can fire; the ground is laid whether
		// it lands or not.
		spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)

		spell.AOEDot().Apply(sim)
	}

	paladin.RegisterSpell(config)
}

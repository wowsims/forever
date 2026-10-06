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
// The ticks crit in game, though no row in the ladder states Periodic Can Crit.
func (paladin *Paladin) registerConsecration(n int32, rank *spelldata.Spell) {
	tickSpell := spellData.ConsecrationTriggered.Rank(n)
	tick := tickSpell.EffectN(1)
	bonus := tickSpell.EffectN(2)

	dummy := rank.Effect(dbcenums.A_PERIODIC_DUMMY, 0)
	bonusTargets := int(dummy.Average())
	tickLength := dummy.Period()
	numberOfTicks := int32(rank.Duration() / tickLength)

	// The bonus scales on its own coefficient, so it is added to the base damage here rather than
	// through the dot's, which is the base tick's. Consecrated Ground marks the same targets.
	dealTick := func(sim *core.Simulation, dot *core.Dot, firstTick bool) {
		for i, target := range sim.Encounter.ActiveTargetUnits {
			damage := tick.Average()
			if i < bonusTargets {
				damage += bonus.Average() + bonus.Coeff()*dot.Spell.BonusDamage(dot.Spell.Unit.AttackTables[target.UnitIndex])
				if firstTick && paladin.consecratedGroundAuras != nil {
					paladin.consecratedGroundAuras.Get(target).Activate(sim)
				}
			}
			dot.Spell.CalcAndDealPeriodicDamage(sim, target, damage, dot.Spell.OutcomeTickMagicHitAndCrit)
		}
	}

	config := spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Flags(core.SpellFlagPureDot))
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
		NumberOfTicks:    numberOfTicks - 1, // the sim adds an immediate tick below
		TickLength:       tickLength,
		BonusCoefficient: tick.Coeff(),
		OnTick: func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
			dealTick(sim, dot, false)
		},
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		// Consecration does one hit check on cast but the ground effect will still be applied
		// meaning it's only needed to proc things like Eye of Magtheridon (procs on resist)
		spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)

		dot := spell.AOEDot()
		dot.Apply(sim)
		dealTick(sim, dot, true)
	}

	paladin.RegisterSpell(config)
}

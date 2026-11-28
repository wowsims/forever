package paladin

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var ConsecrationRankMap = spellData.Consecration

// The mark the ground lays on the first enemies inside it: the bonus tick's dummy, and Consecrated
// Ground's Holy damage from caster mod, whose points the server sets from the talent.
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
//
// A tick is a proc and periodic damage, so neither a listener on spell hits nor one on periodic damage
// that cannot proc from procs hears it (EV-0226).
// TODO: Ingame test needed -- whether a periodic listener that can proc from procs hears a tick. The sim lets it.
func (paladin *Paladin) registerConsecration(n int32, rank *spelldata.Spell) {
	tickSpell := spellData.ConsecrationTriggered.Rank(n)
	tick := tickSpell.EffectN(1)
	bonus := tickSpell.EffectN(2)

	dummy := rank.Effect(dbcenums.A_PERIODIC_DUMMY, 0)
	bonusTargets := int(dummy.Average())
	tickLength := dummy.Period()
	numberOfTicks := int32(rank.Duration() / tickLength)
	markTail := spellData.ConsecratedGroundTriggered.Highest().Duration()

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

	pulseConfig := spelldata.SpellConfig(&paladin.Unit, tickSpell, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc(), spelldata.Triggered())
	pulseConfig.ClassSpellMask = SpellMaskConsecration

	pulseConfig.Dot = core.DotConfig{
		IsAOE: true,
		Aura: core.Aura{
			ActionID: core.ActionID{SpellID: rank.ID},
			Label:    rank.RankedLabel(&paladin.Unit),
		},
		NumberOfTicks:    numberOfTicks,
		TickLength:       tickLength,
		BonusCoefficient: tick.Coeff(),
		// Bonus on marked enemies in EV-0143.
		OnTick: func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
			markTargets(sim, dot.RemainingDuration(sim))
			coefficient := dot.BonusCoefficient
			for _, target := range sim.Encounter.ActiveTargetUnits {
				damage := tick.Average()
				if paladin.consecrationMarks.Get(target).IsActive() {
					damage += bonus.Average()
					dot.BonusCoefficient += bonus.Coeff()
				}
				result := dot.Spell.CalcPeriodicDamage(sim, target, damage, dot.Spell.OutcomeTickMagicHitAndCrit)
				dot.BonusCoefficient = coefficient
				dot.Spell.DealPeriodicDamage(sim, result)
			}
		},
	}
	pulse := paladin.RegisterSpell(pulseConfig)

	config := spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = SpellMaskConsecration
	config.RelatedDotSpell = pulse

	// The cast lands on each enemy in the area as a spell of its own. Proc rolls per enemy in EV-0115.
	// The landing rolls hit, but the ground is laid whether it lands or not, and the roll counts
	// nothing, so the row's Miss % is the ticks' own.
	landing := paladin.RegisterSpell(spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage),
		spelldata.Proc(), spelldata.Tag(1), spelldata.Flags(core.SpellFlagNoMetrics|core.SpellFlagNoLogs)))

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		for _, target := range sim.Encounter.ActiveTargetUnits {
			landing.CalcAndDealOutcome(sim, target, landing.OutcomeTickMagicHitNoHitCounter)
		}

		pulse.AOEDot().Apply(sim)
		markTargets(sim, rank.Duration())
	}

	paladin.RegisterSpell(config)
}

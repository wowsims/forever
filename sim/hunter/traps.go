package hunter

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// TODO: In-game testing required. The rows state no attack power share; the Forever notes give the
// damage traps 10% of ranged attack power over their whole duration, split evenly over the ticks and
// read at each tick. Explosive Trap's initial hit takes the same 10% on top.
const trapTotalAPShare = 0.10

// TODO: In-game testing required. The game files carry no trap object, so its trigger radius and
// arming delay are not in the store. Forever lets traps be placed in combat; a trap lies where the
// hunter stood and springs on a target within this radius once armed.
const trapTriggerRadius = 3
const trapArmingDelay = 500 * time.Millisecond

var immolationTrapRank = spellData.ImmolationTrap.Highest()
var immolationTrapEffectRank = spellData.ImmolationTrapTriggered.Highest()
var explosiveTrapRank = spellData.ExplosiveTrap.Highest()
var explosiveTrapEffectRank = spellData.ExplosiveTrapTriggered.Highest()

func (hunter *Hunter) registerTraps() {
	hunter.registerImmolationTrap()
	hunter.registerExplosiveTrap()
}

// The placement spell: shared trap cooldown, mana and GCD from its row. The trap lies at the hunter's
// distance from the target when placed, so moving away afterwards does not stop it; a trap placed
// out of reach of the standing target never springs.
func (hunter *Hunter) trapConfig(rank *spelldata.Spell, trigger *core.Spell) core.SpellConfig {
	config := spelldata.SpellConfig(&hunter.Unit, rank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.CastRequirement.NotInCombat = false
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
		if hunter.DistanceFromTarget > trapTriggerRadius {
			return
		}
		sim.AddPendingAction(&core.PendingAction{
			NextActionAt: sim.CurrentTime + trapArmingDelay,
			OnAction: func(sim *core.Simulation) {
				trigger.Cast(sim, target)
			},
		})
	}
	return config
}

func (hunter *Hunter) registerImmolationTrap() {
	tick := immolationTrapEffectRank.PeriodicEffect()

	config := spelldata.SpellConfig(&hunter.Unit, immolationTrapEffectRank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Flags(core.SpellFlagNoOnCastComplete|core.SpellFlagPassiveSpell))
	config.DamageMultiplier = 1
	config.ThreatMultiplier = 1

	config.Dot = spelldata.DotConfig(immolationTrapEffectRank, tick)
	tickAPShare := trapTotalAPShare / float64(config.Dot.NumberOfTicks)
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		baseDamage := tick.Average() + tickAPShare*dot.Spell.RangedAttackPower(target)
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, baseDamage, immolationTrapEffectRank.TickOutcome(dot))
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
	}

	hunter.RegisterSpell(hunter.trapConfig(immolationTrapRank, hunter.RegisterSpell(config)))
}

func (hunter *Hunter) registerExplosiveTrap() {
	hit := explosiveTrapEffectRank.EffectN(1)
	// Effect 2 is a persistent area aura the resolvers do not read, so its dot is built here.
	area := explosiveTrapEffectRank.EffectN(2)
	numberOfTicks := int32(explosiveTrapEffectRank.Duration() / area.Period())
	tickAPShare := trapTotalAPShare / float64(numberOfTicks)

	config := spelldata.SpellConfig(&hunter.Unit, explosiveTrapEffectRank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Flags(core.SpellFlagNoOnCastComplete|core.SpellFlagPassiveSpell))
	config.DamageMultiplier = 1
	config.ThreatMultiplier = 1

	config.Dot = core.DotConfig{
		IsAOE: true,
		Aura: core.Aura{
			ActionID: core.ActionID{SpellID: explosiveTrapEffectRank.ID},
			Label:    explosiveTrapEffectRank.RankedLabel(&hunter.Unit),
		},
		NumberOfTicks: numberOfTicks,
		TickLength:    area.Period(),
		OnTick: func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
			for _, target := range sim.Encounter.ActiveTargetUnits {
				baseDamage := area.Average() + tickAPShare*dot.Spell.RangedAttackPower(target)
				dot.Spell.CalcAndDealPeriodicDamage(sim, target, baseDamage, dot.Spell.OutcomeTickMagicHitAndCrit)
			}
		},
	}

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		for _, target := range sim.Encounter.ActiveTargetUnits {
			spell.CalcAndDealDamage(sim, target, hit.Roll(sim)+trapTotalAPShare*spell.RangedAttackPower(target), spell.OutcomeMagicHitAndCrit)
		}
		spell.AOEDot().Apply(sim)
	}

	hunter.RegisterSpell(hunter.trapConfig(explosiveTrapRank, hunter.RegisterSpell(config)))
}

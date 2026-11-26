package hunter

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The rows state no attack power share. In game Explosive Trap's initial hit takes 10% of ranged
// attack power and its ticks take none (EV-0068). TODO: In-game testing required for Immolation Trap, which the
// Forever notes give the same 10% over its whole duration, split evenly over the ticks and read at
// each tick.
const trapTotalAPShare = 0.10

// TODO: In-game testing required. The game files carry no trap object, so its trigger radius is not
// in the store. Forever lets traps be placed in combat; a trap lies where the hunter stood and springs
// on a target within this radius once armed.
const trapTriggerRadius = 5

// The game files carry no arming delay either. A trap arms 0.8 s after it is placed and looks for a
// target every 0.4 s from a point that falls evenly in its first 0.4 s; about one trap in four springs
// a check later (EV-0218).
const trapArmingDelay = 800 * time.Millisecond
const trapCheckInterval = 400 * time.Millisecond
const trapLateCheckChance = 0.24

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
		// A trap placed mid-walk lies where the hunter is now, not where the walk started.
		hunter.UpdatePosition(sim, false)
		if hunter.DistanceFromTarget > trapTriggerRadius+1e-6 {
			return
		}
		delay := trapArmingDelay + time.Duration(sim.RandomFloat("Trap Check")*float64(trapCheckInterval))
		if sim.Proc(trapLateCheckChance, "Trap Check") {
			delay += trapCheckInterval
		}
		sim.AddPendingAction(&core.PendingAction{
			NextActionAt: sim.CurrentTime + delay,
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
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.RemoveFlags(core.SpellFlagAPL),
		spelldata.Flags(core.SpellFlagNoOnCastComplete|core.SpellFlagPassiveSpell))
	if immolationTrapEffectRank.IsAProc() {
		config.Flags |= core.SpellFlagProc
	}
	config.DamageMultiplier = 1
	config.ThreatMultiplier = 1

	// The trap casts the debuff where it lies, so the row's 10 yd range from the hunter does not
	// limit it.
	config.MaxRange = 0

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
	tickDamage := area.Average()

	config := spelldata.SpellConfig(&hunter.Unit, explosiveTrapEffectRank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.RemoveFlags(core.SpellFlagAPL),
		spelldata.Flags(core.SpellFlagNoOnCastComplete|core.SpellFlagPassiveSpell))
	config.DamageMultiplier = 1
	config.ThreatMultiplier = 1
	// The hit rolls its base once, below, and the attack power share not at all (EV-0134).
	config.DirectVariance = 0

	config.Dot = core.DotConfig{
		IsAOE: true,
		Aura: core.Aura{
			ActionID: core.ActionID{SpellID: explosiveTrapEffectRank.ID},
			Label:    explosiveTrapEffectRank.RankedLabel(&hunter.Unit),
		},
		NumberOfTicks: int32(explosiveTrapEffectRank.Duration() / area.Period()),
		TickLength:    area.Period(),
		OnTick: func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
			outcome := explosiveTrapEffectRank.TickOutcome(dot)
			for _, target := range sim.Encounter.ActiveTargetUnits {
				dot.Spell.CalcAndDealPeriodicDamage(sim, target, tickDamage, outcome)
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

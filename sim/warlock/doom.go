package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// One tick at the end of its minute: the row's period is its duration.
// TODO: Bane of Doom Effect (18662) summons a Doomguard (11859) for 15s. The tooltip has it come only when
// Doom's tick is the killing blow, at a chance no row states, and a target dies on the encounter's schedule
// rather than from a damage event; the sim has no Doomguard either.
func (warlock *Warlock) registerBaneOfDoom() {
	rankedSpell := spellData.BaneOfDoom.ByID(603)
	doomTick := rankedSpell.PeriodicEffect()

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellBaneOfDoom

	// Tick the row's average, rolling only the crit: the hit was rolled when the dot landed
	config.Dot = spelldata.DotConfig(rankedSpell, doomTick)
	config.Dot.Aura.Tag = "Affliction"
	config.Dot.OnTick = spelldata.PeriodicDamageTick(doomTick, rankedSpell.TickOutcome)

	// Roll to hit and apply as dot in place of any other Bane
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
		if result.Landed() {
			warlock.deactivateOtherBanes(sim, spell, target)
			spell.Dot(target).Apply(sim)
		}
		spell.DealOutcome(sim, result)
	}

	// Calculate expected damage from the row's average tick
	config.ExpectedTickDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcPeriodicDamage(sim, target, doomTick.Average(), spell.OutcomeExpectedMagicAlwaysHit)
	}

	warlock.BaneOfDoom = warlock.RegisterSpell(config)
	warlock.Banes = append(warlock.Banes, warlock.BaneOfDoom)
}

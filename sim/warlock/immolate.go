package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Tags every warlock's Immolate dot, so Incinerate can see any Immolate on the target.
const immolateTag = "Immolate"

// A hit, then a dot. Effect 3 is a server script (E_SCRIPT_EFFECT) the game files do not describe.
func (warlock *Warlock) registerImmolate() {
	rankedSpell := spellData.Immolate.ByID(25309)
	immolateHit := rankedSpell.DamageEffect()
	immolateTick := rankedSpell.PeriodicEffect()

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellImmolate

	// Tick the row's average, rolling only the crit: the hit was rolled when the dot landed
	config.Dot = spelldata.DotConfig(rankedSpell, immolateTick)
	config.Dot.Aura.Tag = immolateTag
	config.Dot.OnTick = spelldata.PeriodicDamageTick(immolateTick, rankedSpell.TickOutcome)

	// Roll the hit, and apply the dot if it landed
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, immolateHit.Average(), spell.OutcomeMagicHitAndCrit)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
	}

	// Calculate expected damage from the row's average hit and tick
	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, immolateHit.Average(), spell.OutcomeExpectedMagicCrit)
	}
	config.ExpectedTickDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcPeriodicDamage(sim, target, immolateTick.Average(), spell.OutcomeExpectedMagicAlwaysHit)
	}

	warlock.Immolate = warlock.RegisterSpell(config)
}

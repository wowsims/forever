package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Granted by the Conflagrate talent: ignites the warlock's own Immolate and consumes it, unless Shadow and
// Flame keeps it burning. The 10 sec cooldown is the row's category cooldown.
func (warlock *Warlock) registerConflagrate() {
	rankedSpell := spellData.Conflagrate.ByID(18932)
	conflagrateHit := rankedSpell.DamageEffect()

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellConflagrate

	// Needs this warlock's Immolate on the target. Immolate registers after the talents, so it is read here.
	config.ExtraCastCondition = func(_ *core.Simulation, target *core.Unit) bool {
		return warlock.Immolate != nil && warlock.Immolate.Dot(target).IsActive()
	}

	// Roll the damage and to hit, then consume the Immolate unless Shadow and Flame keeps it
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, conflagrateHit.Average(), spell.OutcomeMagicHitAndCrit)
		if result.Landed() && !sim.Proc(warlock.conflagrateKeepsImmolate, "Shadow and Flame") {
			warlock.Immolate.Dot(target).Deactivate(sim)
		}
	}

	// Calculate expected damage from the row's average hit
	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, conflagrateHit.Average(), spell.OutcomeExpectedMagicHitAndCrit)
	}

	warlock.Conflagrate = warlock.RegisterSpell(config)
}

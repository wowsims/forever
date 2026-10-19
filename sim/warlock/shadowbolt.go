package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Nightfall's Shadow Trance makes it instant through the cast time modifier parsed onto that aura.
func (warlock *Warlock) registerShadowBolt() {
	rankedSpell := spellData.ShadowBolt.ByID(25307)
	shadowBoltHit := rankedSpell.DamageEffect()

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellShadowBolt

	// Roll the damage and to hit, then land it after the bolt's travel time
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, shadowBoltHit.Average(), spell.OutcomeMagicHitAndCrit)
		spell.DealDamageAfterTravel(sim, result)
	}

	// Calculate expected damage from the row's average hit
	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, shadowBoltHit.Average(), spell.OutcomeExpectedMagicCrit)
	}

	warlock.ShadowBolt = warlock.RegisterSpell(config)
}

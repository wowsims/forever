package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Rank 9 is the last rank the trainer teaches; rank 10 comes from the Grimoire of Shadow Bolt X book.
func (warlock *Warlock) registerShadowBolt() {
	warlock.ShadowBolt = warlock.registerShadowBoltRank(spellData.ShadowBolt.ByID(11661))
	warlock.registerShadowBoltRank(spellData.ShadowBolt.ByID(25307))
}

// Nightfall's Shadow Trance makes it instant through the cast time modifier parsed onto that aura.
func (warlock *Warlock) registerShadowBoltRank(rankedSpell *spelldata.Spell) *core.Spell {
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

	return warlock.RegisterSpell(config)
}

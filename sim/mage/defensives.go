package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var counterspellRank = spellData.Counterspell.Highest()

func (mage *Mage) registerDefensiveSpells() {
	// Only the highest rank is registered: the mage casts the rank configured here, not every rank
	// spellData.FireWard lists.
	mage.registerWardSpell(spellData.FireWard.Highest(), nil)
	// Only the highest rank is registered: the mage casts the rank configured here, not every rank
	// spellData.FrostWard lists.
	mage.registerWardSpell(spellData.FrostWard.Highest(), nil)
	// Only the highest rank is registered: the mage casts the rank configured here, not every rank
	// spellData.ManaShield lists.
	mage.registerManaShieldSpell(spellData.ManaShield.Len(), spellData.ManaShield.Highest())
	mage.registerCounterspellSpell()
}

// An absorb shield on the row's effect 1, which takes only the school its misc names. Fire Ward and
// Frost Ward ("Absorbs $s1 Fire damage. Lasts 30 sec.") share a category cooldown; their effect 2 is
// the reflect Frost Warding and Improved Fire Ward give a chance to, which needs an enemy casting at
// the mage, so it is not modelled.
func (mage *Mage) registerWardSpell(rank *spelldata.Spell, onAbsorbed core.OnDamageAbsorbedCallback) *core.DamageAbsorptionAura {
	absorbEffect := rank.EffectN(1)
	school := core.SpellSchool(absorbEffect.Misc)
	strength := absorbEffect.Average()

	shield := mage.NewDamageAbsorptionAura(core.AbsorptionAuraConfig{
		Aura: spelldata.AuraConfig(rank, spelldata.RankLabel(rank, &mage.Unit)),
		ShieldStrengthCalculator: func(_ *core.Unit) float64 {
			return strength
		},
		ShouldApplyToResult: func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult, _ bool) bool {
			return spell.SpellSchool.Matches(school)
		},
		OnDamageAbsorbed: onAbsorbed,
	})

	config := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagHelpful))
	config.RelatedSelfBuff = shield.Aura

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		shield.Activate(sim)
	}

	mage.RegisterSpell(config)
	return shield
}

// Mana Shield: "Absorbs $s1 physical damage, at the cost of 2 mana per damage absorbed. Lasts 1 min."
// The school is effect 1's misc (1, physical) and the mana per point its amplitude. Arcane Shielding
// lowers the mana per point (its effect 1). The shield breaks once the mana runs out; the hit that
// empties it is still absorbed whole.
func (mage *Mage) registerManaShieldSpell(_ int32, rank *spelldata.Spell) {
	manaPerPoint := float64(rank.EffectN(1).Amplitude) * (1 + spellData.ArcaneShielding.EffectAt(1).FractionAt(mage.Talents.ArcaneShielding))
	manaMetrics := mage.NewManaMetrics(core.ActionID{SpellID: rank.ID})

	mage.registerWardSpell(rank,
		func(sim *core.Simulation, shield *core.DamageAbsorptionAura, _ *core.SpellResult, absorbed float64) {
			cost := absorbed * manaPerPoint
			if cost >= mage.CurrentMana() {
				mage.SpendMana(sim, mage.CurrentMana(), manaMetrics)
				shield.Deactivate(sim)
				return
			}
			mage.SpendMana(sim, cost, manaMetrics)
		})
}

// Counterspell interrupts the target's cast and locks it out of that school for 10 sec. The sim has
// no enemy cast to stop, so the spell only takes its mana, global cooldown and cooldown.
func (mage *Mage) registerCounterspellSpell() {
	config := spelldata.SpellConfig(&mage.Unit, counterspellRank, spelldata.Flags(core.SpellFlagAPL))

	config.ApplyEffects = func(_ *core.Simulation, _ *core.Unit, _ *core.Spell) {}

	mage.RegisterSpell(config)
}

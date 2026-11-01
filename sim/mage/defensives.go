package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var counterspellRank = spellData.Counterspell.Highest()

// Rank 4 is the last rank the trainer teaches; rank 5 comes from the Tome of Frost Ward V book.
var frostWardTrainerRank = spellData.FrostWard.ByID(10177)
var frostWardBookRank = spellData.FrostWard.ByID(28609)

func (mage *Mage) registerDefensiveSpells() {
	// Only the highest rank is registered: the mage casts the rank configured here, not every rank
	// spellData.FireWard lists.
	mage.registerWardSpell(spellData.FireWard.Highest(), nil, nil)
	mage.registerFrostWardSpell(frostWardTrainerRank)
	mage.registerFrostWardSpell(frostWardBookRank)
	// Only the highest rank is registered: the mage casts the rank configured here, not every rank
	// spellData.ManaShield lists.
	mage.registerManaShieldSpell(spellData.ManaShield.Len(), spellData.ManaShield.Highest())
	mage.registerCounterspellSpell()
}

// "Absorbs ${$s1+($SPFR*0.1)*($11189s3/100+1)} Frost damage". The spell power share is the
// tooltip's, since the row states no coefficient, and Frost Warding's effect 3 raises the whole absorb.
func (mage *Mage) registerFrostWardSpell(rank *spelldata.Spell) {
	warding := 1 + spellData.FrostWarding.EffectAt(3).FractionAt(mage.Talents.FrostWarding)
	base := rank.EffectN(1).Average()

	mage.registerWardSpell(rank, func() float64 {
		return warding * (base + 0.1*mage.frostSpellPower())
	}, nil)
}

// An absorb shield on the row's effect 1, which takes only the school its misc names. Fire Ward and
// Frost Ward share a category cooldown; their effect 2 is the reflect or negate Improved Fire Ward and
// Frost Warding give a chance to, which needs an enemy casting at the mage, so it is not modelled.
// The shield absorbs effect 1's value, or what absorbAt answers when the ward is cast.
func (mage *Mage) registerWardSpell(rank *spelldata.Spell, absorbAt func() float64, onAbsorbed core.OnDamageAbsorbedCallback) *core.DamageAbsorptionAura {
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
		if absorbAt != nil {
			strength = absorbAt()
		}
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

	mage.registerWardSpell(rank, nil,
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

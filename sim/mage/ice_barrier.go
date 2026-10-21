package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Frigid Power, the Frost damage buff every Ice Barrier cast applies. No effect of Ice Barrier names
// it; its tooltip does ("$1326891s1%").
var frigidPowerRank = spellData.IceBarrierTriggered.ByID(1326891)

// Only the highest rank is registered: the mage casts the rank configured here, not every rank
// spellData.IceBarrier lists.
func (mage *Mage) registerIceBarrierSpell() {
	mage.registerIceBarrierRank(spellData.IceBarrier.Len(), spellData.IceBarrier.Highest())
}

// Ice Barrier
// https://www.wowhead.com/forever/spell=13033
//
// "Shield yourself, absorbing ${$s1+($SPFR*0.1)*($11189s3/100+1)} damage and increasing your Frost
// damage by $1326891s1% for $d. While shielded, your Frost damage is increased by an additional $s2%
// and you do not lose casting time when you take damage."
//
// The shield's effects 2 and 3 raise Frost damage while it holds, and Frigid Power raises it again for
// the full minute, however soon the shield breaks. The spell power share is the tooltip's, since the
// row states no coefficient, and Frost Warding's effect 3 raises the whole absorb. The sim models no
// pushback, so the casting time clause does nothing.
func (mage *Mage) registerIceBarrierRank(_ int32, rank *spelldata.Spell) {
	frigidPower := mage.RegisterAura(spelldata.AuraConfig(frigidPowerRank))
	spelldata.ParseEffects(&mage.Character, frigidPower, frigidPowerRank)

	warding := 1 + spellData.FrostWarding.EffectAt(3).FractionAt(mage.Talents.FrostWarding)
	base := rank.EffectN(1).Average()
	var absorb float64

	shield := mage.NewDamageAbsorptionAura(core.AbsorptionAuraConfig{
		Aura: spelldata.AuraConfig(rank, spelldata.RankLabel(rank, &mage.Unit)),
		ShieldStrengthCalculator: func(_ *core.Unit) float64 {
			return absorb
		},
	})
	spelldata.ParseEffects(&mage.Character, shield.Aura, rank, spelldata.Effects(2, 3))

	config := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Flags(core.SpellFlagAPL))
	config.ClassSpellMask = MageSpellIceBarrier
	config.RelatedSelfBuff = shield.Aura

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		absorb = warding * (base + 0.1*mage.frostSpellPower())
		shield.Activate(sim)
		frigidPower.Activate(sim)
	}

	mage.RegisterSpell(config)
}

// The spell power a Frost spell reads ($SPFR in a tooltip): the general bonus plus the Frost one.
func (mage *Mage) frostSpellPower() float64 {
	return mage.GetStat(stats.SpellDamage) + mage.GetStat(stats.FrostDamage)
}

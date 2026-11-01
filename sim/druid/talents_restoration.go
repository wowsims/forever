package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/stats"
)

func (druid *Druid) registerRestorationTalents() {
	// Tier 1
	druid.applyNaturesFocus()
	druid.applyHeartOfTheWild()

	// Tier 2
	druid.applyNaturalist()
	druid.applyReflection()
	druid.applyNaturalShapeshifter()

	// Tier 3
	druid.applySubtlety()
	druid.applyGiftOfNature()
	druid.applyBloom()

	// Tier 4
	druid.applyTranquilSpirit()
	druid.applyImprovedRejuvenation()
	druid.applySwiftmend()

	// Tier 5
	druid.applyNaturesSwiftness()
	druid.applyLivingSpirit()
	druid.applyImprovedTranquility()

	// Tier 6
	druid.applyImprovedRegrowth()

	// Tier 7
	druid.applyWildGrowth()
}

func (druid *Druid) applyHeartOfTheWild() {
	if druid.Talents.HeartOfTheWild == 0 {
		return
	}
	// Effect 1 is the Intellect ($s1), which the parse reads from its stat mask. The bear Stamina ($s2)
	// and cat Strength ($s3) are dummies the forms carry.
	druid.parseTalent(druid.Talents.HeartOfTheWild, spellData.HeartOfTheWild)
	row := spellData.HeartOfTheWild.Rank(druid.Talents.HeartOfTheWild)
	strength := druid.NewDynamicMultiplyStat(stats.Strength, 1+row.EffectN(3).Percent())
	if druid.CatFormAura != nil {
		druid.CatFormAura.AttachStatDependency(strength)
	}
	stamina := druid.NewDynamicMultiplyStat(stats.Stamina, 1+row.EffectN(2).Percent())
	for _, aura := range []*core.Aura{druid.BearFormAura, druid.DireBearFormAura} {
		if aura != nil {
			aura.AttachStatDependency(stamina)
		}
	}
}

func (druid *Druid) applyNaturalShapeshifter() {
	druid.parseTalent(druid.Talents.NaturalShapeshifter, spellData.NaturalShapeshifter)
}

func (druid *Druid) applyNaturalist() {
	druid.parseTalent(druid.Talents.Naturalist, spellData.Naturalist)
}

func (druid *Druid) applySubtlety() {
	druid.parseTalent(druid.Talents.Subtlety, spellData.Subtlety)
}

func (druid *Druid) applyLivingSpirit() {
	druid.parseTalent(druid.Talents.LivingSpirit, spellData.LivingSpirit)
}

// The chance a hit does not push back an Arcane or Nature cast. Every spell the druid hardcasts is
// one, so the unit's own pushback chance carries it.
func (druid *Druid) applyNaturesFocus() {
	if druid.Talents.NaturesFocus == 0 {
		return
	}
	druid.PseudoStats.PushbackChance -= spellData.NaturesFocus.FractionAt(druid.Talents.NaturesFocus)
}

func (druid *Druid) applyReflection() {
	druid.parseTalent(druid.Talents.Reflection, spellData.Reflection)
}

// applyGiftOfNature implements Gift of Nature, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyGiftOfNature() {
	if druid.Talents.GiftOfNature == 0 {
		return
	}
}

// TODO: To be implemented. The heal on a Rejuvenation or Regrowth expiry needs those heals, which
// the sim does not register.
func (druid *Druid) applyBloom() {
	if !druid.Talents.Bloom {
		return
	}
}

// applyTranquilSpirit implements Tranquil Spirit, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyTranquilSpirit() {
	if druid.Talents.TranquilSpirit == 0 {
		return
	}
}

// applyImprovedRejuvenation implements Improved Rejuvenation, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyImprovedRejuvenation() {
	if druid.Talents.ImprovedRejuvenation == 0 {
		return
	}
}

// applySwiftmend implements Swiftmend, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applySwiftmend() {
	if !druid.Talents.Swiftmend {
		return
	}
}

// applyImprovedTranquility implements Improved Tranquility, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyImprovedTranquility() {
	if druid.Talents.ImprovedTranquility == 0 {
		return
	}
}

// applyImprovedRegrowth implements Improved Regrowth, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyImprovedRegrowth() {
	if druid.Talents.ImprovedRegrowth == 0 {
		return
	}
}

// applyWildGrowth implements Wild Growth, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyWildGrowth() {
	if !druid.Talents.WildGrowth {
		return
	}
}

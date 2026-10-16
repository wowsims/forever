package druid

func (druid *Druid) registerRestorationTalents() {
	// Tier 1
	druid.applyNaturesFocus()
	// Furor is read on entering Cat Form and the bear forms (forms.go).

	// Tier 2
	druid.applyNaturalist()
	druid.applySubtlety()
	druid.applyNaturalShapeshifter()

	// Tier 3
	druid.applyReflection()
	druid.applyGiftOfNature()
	druid.applyGiftOfTheEarthmother()

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

// applyGiftOfTheEarthmother implements Gift of the Earthmother, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyGiftOfTheEarthmother() {
	if !druid.Talents.GiftOfTheEarthmother {
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

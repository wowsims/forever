package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var coldSnapRank = spellData.ColdSnap.Highest()

// Cold Snap
// https://www.wowhead.com/forever/spell=12472
//
// "Finishes the remaining cooldown on all your other Frost spells." The row's effect is a dummy, so
// "Frost spells" is read as the school: Frost Nova, Cone of Cold, Ice Block and Ice Barrier. A
// cooldown a spell of another school also runs on is left alone, so Frost Ward, which shares
// category 56 with Fire Ward, is not reset.
func (mage *Mage) registerColdSnapSpell() {
	var resets []*core.Spell

	config := spelldata.SpellConfig(&mage.Unit, coldSnapRank, spelldata.Flags(core.SpellFlagAPL))
	config.ClassSpellMask = MageSpellColdSnap

	config.ApplyEffects = func(_ *core.Simulation, _ *core.Unit, _ *core.Spell) {
		for _, spell := range resets {
			spell.CD.Reset()
		}
	}

	coldSnap := mage.RegisterSpell(config)

	mage.Env.RegisterPreFinalizeEffect(func() {
		otherSchools := make(map[*core.Timer]bool)
		for _, spell := range mage.Spellbook {
			if !spell.SpellSchool.Matches(core.SpellSchoolFrost) {
				otherSchools[spell.CD.Timer] = true
				otherSchools[spell.SharedCD.Timer] = true
			}
		}

		for _, spell := range mage.Spellbook {
			if spell != coldSnap && spell.SpellSchool.Matches(core.SpellSchoolFrost) && spell.CD.Timer != nil && !otherSchools[spell.CD.Timer] {
				resets = append(resets, spell)
			}
		}
	})
}

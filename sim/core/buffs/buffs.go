// Package buffs holds the raid buffs and debuffs: the constructors tools/database/gen_spelldata
// generates from tools/database/buffmanifest, the drivers that add what the game files do not state,
// and the hand-written auras that share categories with them. They sit above sim/core so that they
// can read the spell store, which imports sim/core, so core cannot call them: init registers them
// as core's buff hooks, and sim/common imports the package so that every sim links it.
package buffs

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func init() {
	core.RegisterBuffHooks(core.BuffHooks{
		ApplyBuffs:   applyGeneratedBuffs,
		ApplyDebuffs: applyDebuffs,
	})
}

// A warlock's own Blood Pact, one per warlock, bidding in the party buff's category.
func NewSpellDataBloodPactAuraFor(unit *core.Unit, warlock *core.Unit) *core.Aura {
	return newCasterBuff(unit, bloodPactMeta, warlock, 0)
}

// A priest's own Dark Inspiration at the talent's rank, one per priest, bidding in the party buff's
// category. It yields ties, so a copy no stronger than the party's leaves the party's holding the
// Spirit while it is up.
func NewSpellDataDarkInspirationAuraFor(unit *core.Unit, priest *core.Unit, ranks int32) *core.Aura {
	aura := newCasterBuff(unit, darkInspirationMeta, priest, ranks)
	for _, effect := range aura.ExclusiveEffects {
		effect.YieldsTies = true
	}
	return aura
}

// The share of an Innervate a druid gains on casting it on another player (Dreamstate): every amount
// of the spell at percent of its value, bidding in Innervate's category.
func NewSpellDataInnervatesAuraShare(druid *core.Unit, percent int32) *core.Aura {
	opts := append(innervatesMeta.Options(0), spelldata.ScaledByRanks(percent, 100))
	return innervatesMeta.buff(druid, true, innervatesMeta.Label+" (Dreamstate)", 0, opts)
}

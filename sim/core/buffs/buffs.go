// Package buffs holds the raid buffs and debuffs: the constructors tools/database/gen_spelldata
// generates from tools/database/buffmanifest, the drivers that add what the game files do not state,
// and the hand-written auras that share categories with them. They sit above sim/core so that they
// can read the spell store, which imports sim/core, so core cannot call them: init registers them
// as core's buff hooks, and sim/common imports the package so that every sim links it.
package buffs

import (
	"github.com/wowsims/forever/sim/core"
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

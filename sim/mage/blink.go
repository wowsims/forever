package mage

import (
	"math"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var blinkRank = spellData.Blink.Highest()

// Blink
// https://www.wowhead.com/forever/spell=1953
//
// The leap's radius is how far the mage moves toward the target. The move takes the row's duration,
// which the aura of its mechanic immunities (effects 1 and 2) lasts too; the immunities do nothing
// against a boss the sim fights. From closer than the radius the mage passes through the target and
// lands the rest of the way past it. The sim has no facing, so that move is a straight walk to where
// the mage lands, and only its end point is right.
func (mage *Mage) registerBlinkSpell() {
	leap := float64(blinkRank.FindEffect(dbcenums.E_LEAP, 0, 0).RadiusMax)
	aura := mage.RegisterAura(spelldata.AuraConfig(blinkRank))

	config := spelldata.SpellConfig(&mage.Unit, blinkRank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagMovesToTarget))

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		aura.Activate(sim)
		mage.MoveToOver(math.Abs(mage.DistanceFromTarget-leap), aura.Duration, sim)
	}

	mage.RegisterSpell(config)
}

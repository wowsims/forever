package forever

import (
	"fmt"

	"github.com/wowsims/forever/sim/common/shared"
)

// The Arcane Elemental 277729 that Violet Sorcerer's Vestments' 5-piece bonus summons: a guardian that only
// melees, a physical swing every 2.0 s. The game files carry no creature row for it. Its normal hit before
// armor is 0.62 of 2 s of the creature dps for its owner's level, ±10%, with no spell power share (EV-0169);
// the target's armor, the attack table and the crit multiplier apply as for any pet's melee.
var arcaneElemental = shared.GuardianCreature{
	Name:  "Arcane Elemental",
	Swing: shared.GuardianSwing{Share: 0.62, Speed: 2.0, Variance: 0.1},
}

// Two elementals can be up at once, since one lasts 20 s and the summon rests 15 s (EV-0169). Two pieces carry
// one each, the robes "Arcane Elemental 1" and the mantle "Arcane Elemental 2", since every piece is worn when the
// 5-piece bonus is.
func init() {
	for i, itemID := range []int32{273044, 273051} { // Violet Sorcerer's Robes, Violet Sorcerer's Mantle
		elemental := arcaneElemental
		elemental.Name = fmt.Sprintf("%s %d", arcaneElemental.Name, i+1)
		shared.RegisterItemGuardian(itemID, arcaneElementalCall, elemental)
	}
}

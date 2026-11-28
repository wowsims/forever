package forever

import (
	"fmt"

	"github.com/wowsims/forever/sim/common/shared"
)

// The Arcane Elemental 277729 that Violet Sorcerer's Vestments' 5-piece bonus summons. The game files carry no
// creature row for it.
// Normal hit in EV-0169.
var arcaneElemental = shared.GuardianCreature{
	Name:  "Arcane Elemental",
	Swing: shared.GuardianSwing{Share: 0.62, Speed: 2.0, Variance: 0.1},
}

// Elementals up at once in EV-0169.
// Two pieces carry one each, since every piece is worn when the 5-piece bonus is.
func init() {
	for i, itemID := range []int32{273044, 273051} { // Violet Sorcerer's Robes, Violet Sorcerer's Mantle
		elemental := arcaneElemental
		elemental.Name = fmt.Sprintf("%s %d", arcaneElemental.Name, i+1)
		shared.RegisterItemGuardian(itemID, arcaneElementalCall, elemental)
	}
}

package forever

import (
	"github.com/wowsims/forever/sim/common/shared"
)

// The Arcane Elemental 277729 that Violet Sorcerer's Vestments' 5-piece bonus summons: a guardian that only
// melees, a physical swing every 2.0 s. The game files carry no creature row for it. Its normal hit before
// armor is 0.62 of 2 s of the creature dps for its owner's level, ±10%, with no spell power share (EV-0169);
// the target's armor, the attack table and the crit multiplier apply as for any pet's melee.
var arcaneElemental = shared.GuardianConfig{
	Summon: arcaneElementalCall,
	Swing:  shared.GuardianSwing{Share: 0.62, Speed: 2.0, Variance: 0.1},
}

// Two elementals can be up at once, since one lasts 20 s and the summon rests 15 s (EV-0169). Two pieces carry
// one each, since every piece is worn when the 5-piece bonus is.
func init() {
	shared.RegisterItemGuardian(273044, "Arcane Elemental 1", arcaneElemental) // Violet Sorcerer's Robes
	shared.RegisterItemGuardian(273051, "Arcane Elemental 2", arcaneElemental) // Violet Sorcerer's Mantle
}

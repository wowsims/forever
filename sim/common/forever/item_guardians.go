package forever

import "github.com/wowsims/forever/sim/common/shared"

// The guardians item summons put up, by the creature id the summon's E_SUMMON names: what the game files do not
// state, measured in logs. A use item whose creature has no entry here keeps only its stats. A summon on a proc or a
// set bonus is wired by hand; its creature is listed here as well.
var GuardianCreatures = map[int32]shared.GuardianCreature{
	8776:   emeraldDragonWhelp, // Dragon's Call, EV-0021
	11200:  summonedSkeleton,   // Book of the Dead, EV-0217
	277729: arcaneElemental,    // Violet Sorcerer's Vestments, EV-0169
}

// Book of the Dead 13353: "Summons a Skeleton that will protect you for 1min."
// Health, armor and swing in EV-0217.
var summonedSkeleton = shared.GuardianCreature{
	Name:   "Summoned Skeleton",
	Health: 1.0,
	Armor:  1.0,
	Swing:  shared.GuardianSwing{Share: 1.0, Speed: 2.0, Variance: 0.14},
}

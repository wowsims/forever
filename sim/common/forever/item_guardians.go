package forever

import "github.com/wowsims/forever/sim/common/shared"

// The guardians item summons put up, by the creature id the summon's E_SUMMON names: what the game files do not
// state, measured in logs. item_guardians_auto_gen.go registers every use item whose spell summons a creature, and an
// item whose creature has no entry here keeps only its stats; the generator lists it as missing its effect until an
// entry lands, and drops it from that list on the next db run. A summon on a proc or a set bonus is wired by hand; its
// creature is listed here as well, so the table holds every creature the sim builds.
var GuardianCreatures = map[int32]shared.GuardianCreature{
	8776:   emeraldDragonWhelp, // Dragon's Call, EV-0021
	277729: arcaneElemental,    // Violet Sorcerer's Vestments, EV-0169
}

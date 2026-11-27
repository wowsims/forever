package core

import (
	"slices"

	"github.com/wowsims/forever/sim/core/proto"
)

func (encounter *Encounter) InArea(areaType proto.AreaType) bool {
	return slices.Contains(encounter.AreaTypes, areaType)
}

// The areas an effect names: any of the area types.
type AreaTargets struct {
	Areas []proto.AreaType
}

// Whether the encounter is in any of the areas.
func (areas AreaTargets) In(encounter *Encounter) bool {
	return slices.ContainsFunc(areas.Areas, encounter.InArea)
}

// A factor an effect takes while the encounter is in any of the areas, once however many of them it is in.
type AreaBonus struct {
	AreaTargets
	Multiplier float64
}

// The factor in the encounter, 1 where it is in none of the areas.
func (bonus AreaBonus) Factor(encounter *Encounter) float64 {
	if bonus.In(encounter) {
		return bonus.Multiplier
	}
	return 1
}

// The factor a source's effect takes in the encounter, kept for the stats sheet under the source as
// RegisterVersus keeps a versus factor, whether or not the encounter is in the areas. The areas are fixed
// for the run, so the effect takes the factor when it is built. A factor of 1 or in no area changes
// nothing and is not kept.
func (character *Character) RegisterArea(source ActionID, kind proto.VersusEffectKind, bonus AreaBonus) float64 {
	if len(bonus.Areas) == 0 || bonus.Multiplier == 1 {
		return 1
	}
	character.versusRegistrations = append(character.versusRegistrations,
		versusRegistration{source: source, kind: kind, bonus: &VersusBonus{Multiplier: bonus.Multiplier}, areas: bonus.Areas})
	return bonus.Factor(&character.Env.Encounter)
}

// Whether the encounter is in any of the areas, where a source's effect is a flat amount, kept for the stats
// sheet under the source in every encounter as RegisterArea keeps a factor.
func (character *Character) RegisterAreaAmount(source ActionID, kind proto.VersusEffectKind, areas AreaTargets, amount float64) bool {
	character.versusRegistrations = append(character.versusRegistrations,
		versusRegistration{source: source, kind: kind, bonus: &VersusBonus{}, areas: areas.Areas, amount: amount})
	return areas.In(&character.Env.Encounter)
}

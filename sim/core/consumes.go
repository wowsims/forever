package core

import (
	"fmt"
	"slices"

	"github.com/wowsims/forever/sim/core/proto"
)

// Registers all consume-related effects to the Agent.
func applyConsumeEffects(agent Agent) {
	character := agent.GetCharacter()
	if character.Consumables == nil {
		return
	}

	for _, id := range character.SelectedConsumableIDs() {
		if apply, ok := consumableEffect(id); ok {
			apply(agent)
		}
	}
}

var consumableEffects = map[int32]ApplyEffect{}

// The effects built from the consumables' rows. A hand-written effect for the same id wins.
var dataDrivenConsumableEffects = map[int32]ApplyEffect{}

func consumableEffect(id int32) (ApplyEffect, bool) {
	if apply, ok := consumableEffects[id]; ok {
		return apply, true
	}
	apply, ok := dataDrivenConsumableEffects[id]
	return apply, ok
}

func HasConsumableEffect(id int32) bool {
	_, ok := consumableEffect(id)
	return ok
}

// Registers a hand-written ApplyEffect function which will be called while the consumes are applied,
// for any Agent whose ConsumesSpec names the id. It takes the place of a data-driven registration for
// the id, whichever of the two registers first.
func NewConsumableEffect(id int32, consumableEffect ApplyEffect) {
	if !consumableInDatabase(id) {
		return
	}

	if _, ok := consumableEffects[id]; ok {
		panic(fmt.Sprintf("Cannot add multiple effects for one consumable: %d, %#v", id, consumableEffect))
	}

	consumableEffects[id] = consumableEffect
}

// Registers an ApplyEffect function built from the consumable's rows, which applies unless a
// hand-written one is registered for the id.
func NewSpellDataConsumableEffect(id int32, consumableEffect ApplyEffect) {
	if !consumableInDatabase(id) {
		return
	}

	if _, ok := dataDrivenConsumableEffects[id]; ok {
		panic(fmt.Sprintf("Cannot add multiple data-driven effects for one consumable: %d, %#v", id, consumableEffect))
	}

	dataDrivenConsumableEffects[id] = consumableEffect
}

func consumableInDatabase(id int32) bool {
	return !WITH_DB || GetConsumableByID(id).Id != 0
}

// The character's own consumables the spec names, each once, in the order they apply.
func (character *Character) SelectedConsumableIDs() []int32 {
	if character.Consumables == nil {
		return nil
	}
	return selectedConsumableIDs(character.Consumables)
}

// The agent's own consumables the spec names, each once, and of each exclusive group only the first.
func selectedConsumableIDs(consumables *proto.ConsumesSpec) []int32 {
	ids := append([]int32{consumables.PotId, consumables.FlaskId}, consumables.ElixirIds...)
	ids = append(ids, consumables.FoodId, consumables.ConjuredId, consumables.ExplosiveId,
		consumables.OhImbueId, consumables.OhWeaponImbueId, consumables.MhImbueId, consumables.MhWeaponImbueId)
	ids = append(ids, consumables.ExtraIds...)
	ids = append(ids, consumables.Potions...)
	ids = append(ids, consumables.ConjuredItems...)

	var selected []int32
	var groups []string
	for _, id := range ids {
		if id == 0 || slices.Contains(selected, id) {
			continue
		}
		if group := GetConsumableByID(id).GroupKey; group != "" {
			if slices.Contains(groups, group) {
				continue
			}
			groups = append(groups, group)
		}
		selected = append(selected, id)
	}
	return selected
}

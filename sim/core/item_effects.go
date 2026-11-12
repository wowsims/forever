package core

import (
	"fmt"
)

// Function for applying permanent effects to an Agent.
//
// Passing Character instead of Agent would work for almost all cases,
// but there are occasionally class-specific item effects.
type ApplyEffect func(Agent)

var itemEffects = map[int32]ApplyEffect{}
var enchantEffects = map[int32]ApplyEffect{}

// The lowest item level whose effects the database generator writes, and where the AllItems tests start:
// an effect hotfix on lower gear moves the .results files with no level 60 gear at stake.
const MinEffectItemLevel = 40

func HasItemEffect(id int32) bool {
	_, ok := itemEffects[id]
	return ok
}

func HasEnchantEffect(id int32) bool {
	_, ok := enchantEffects[id]
	return ok
}

// Registers an ApplyEffect function which will be called before the Sim
// starts, for any Agent that is wearing the item.
// missingItemEffects collects the items an effect was registered for that the game files
// do not ship, so the count can be reported rather than silently swallowed.
var missingItemEffects []int32

var missingEnchantEffects []int32

// MissingItemEffects reports the item ids whose effects were skipped because the loaded
// database has no such item.
func MissingItemEffects() []int32 { return missingItemEffects }

// MissingEnchantEffects reports the enchant effect ids skipped for the same reason.
func MissingEnchantEffects() []int32 { return missingEnchantEffects }

func NewItemEffect(id int32, itemEffect ApplyEffect) {
	if WITH_DB {
		if GetItemByID(id) == nil {
			// Forever's item set is a subset of TBC's -- 81 items these effects were
			// written for are simply not in the game files any more, and item 1168 is not
			// in the DBC at all. Skipping keeps the implementation around for the day
			// an item comes back, where panicking would stop the sim from starting and
			// deleting the code would lose work that is still correct.
			missingItemEffects = append(missingItemEffects, id)
			return
		}
	}

	if HasItemEffect(id) {
		panic(fmt.Sprintf("Cannot add multiple effects for one item: %d, %#v", id, itemEffect))
	}

	itemEffects[id] = itemEffect
}

func NewEnchantEffect(id int32, enchantEffect ApplyEffect) {
	if WITH_DB {
		if GetEnchantByEffectID(id) == nil {
			// Same as NewItemEffect: enchants the game files do not ship are skipped rather
			// than aborting the process.
			missingEnchantEffects = append(missingEnchantEffects, id)
			return
		}
	}

	if HasEnchantEffect(id) {
		panic(fmt.Sprintf("Cannot add multiple effects for one enchant: %d, %#v", id, enchantEffect))
	}

	enchantEffects[id] = enchantEffect
}

var itemPets = map[int32]func(owner *Character) PetAgent{}

// Registers the pet an item's effect summons, such as a creature a proc calls forth. A pet must be
// added while its owner is constructed, before the item effects run, so the owner builds it then:
// once for an item it wears or carries in its item swap set, if its class can use the item. The
// item's effect finds the pet among the owner's PetAgents.
func NewItemPet(id int32, newPet func(owner *Character) PetAgent) {
	if _, ok := itemPets[id]; ok {
		panic(fmt.Sprintf("Cannot add multiple pets for one item: %d", id))
	}
	itemPets[id] = newPet
}

func (character *Character) addItemPets() {
	added := make(map[int32]bool)
	for _, equipment := range []*Equipment{&character.Equipment, &character.ItemSwap.unEquippedItems} {
		for _, item := range equipment {
			if newPet, ok := itemPets[item.ID]; ok && !added[item.ID] && item.UsableBy(character.Class) {
				character.AddPet(newPet(character))
				added[item.ID] = true
			}
		}
	}
}

func (equipment *Equipment) applyItemEffects(agent Agent, registeredItemEffects map[int32]bool, registeredItemEnchantEffects map[int32]bool) {
	class := agent.GetCharacter().Class
	for _, eq := range equipment {
		if applyItemEffect, ok := itemEffects[eq.ID]; ok && !registeredItemEffects[eq.ID] && eq.UsableBy(class) {
			applyItemEffect(agent)
			registeredItemEffects[eq.ID] = true
		}

		if applyEnchantEffect, ok := enchantEffects[eq.Enchant.EffectID]; ok && !registeredItemEnchantEffects[eq.Enchant.EffectID] {
			applyEnchantEffect(agent)
			registeredItemEnchantEffects[eq.Enchant.EffectID] = true
		}
	}
}

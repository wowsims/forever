package core

import "github.com/wowsims/forever/sim/core/proto"

// Where an aura or spell was registered from, which files it under a submenu in the APL pickers.
type ActionSource int32

const (
	SourceUnset ActionSource = iota // Registration fills in the phase it runs in.
	SourceNone                      // The class's own, whatever phase registers it.
	SourceItem
	SourceTalent
	SourceConsumable
	SourceRacial
	SourceProfession
	SourceExternalBuff
	SourceExternalDebuff
)

func (env *Environment) withSource(source ActionSource, register func()) {
	previous := env.actionSource
	env.actionSource = source
	defer func() { env.actionSource = previous }()
	register()
}

// Registers the auras and spells in register with this source, unless they set their own.
func (character *Character) WithSource(source ActionSource, register func()) {
	character.Env.withSource(source, register)
}

func (unit *Unit) currentSource() ActionSource {
	if unit.Env == nil {
		return SourceUnset
	}
	return unit.Env.actionSource
}

var sourceCategories = map[ActionSource]proto.ActionCategory{
	SourceItem:           proto.ActionCategory_ActionCategoryItem,
	SourceTalent:         proto.ActionCategory_ActionCategoryTalent,
	SourceConsumable:     proto.ActionCategory_ActionCategoryConsumable,
	SourceRacial:         proto.ActionCategory_ActionCategoryRacial,
	SourceProfession:     proto.ActionCategory_ActionCategoryProfession,
	SourceExternalBuff:   proto.ActionCategory_ActionCategoryExternalBuff,
	SourceExternalDebuff: proto.ActionCategory_ActionCategoryExternalDebuff,
}

// A consumable wins over everything, then anything switched on in the settings, then a major
// cooldown, then where it was registered from.
func actionCategory(source ActionSource, isConsumable bool, isMajorCooldown bool) proto.ActionCategory {
	switch {
	case isConsumable || source == SourceConsumable:
		return proto.ActionCategory_ActionCategoryConsumable
	case source == SourceExternalBuff && isMajorCooldown:
		return proto.ActionCategory_ActionCategoryExternalCooldown
	case source == SourceExternalBuff || source == SourceExternalDebuff:
		return sourceCategories[source]
	case isMajorCooldown:
		return proto.ActionCategory_ActionCategoryCooldown
	}
	return sourceCategories[source]
}

// A spell a talent grants is listed with the class's own spells; only talent auras get their own
// submenu.
func (spell *Spell) category() proto.ActionCategory {
	source := spell.Source
	if source == SourceTalent {
		source = SourceNone
	}
	return actionCategory(source, spell.Flags.Matches(SpellFlagPotion|SpellFlagConjured), spell.Flags.Matches(SpellFlagMCD))
}

// An aura is a major cooldown's when a major cooldown casts it: the spell names it as its buff or
// shares its id. The spell can be another unit's, such as an owner's cooldown that buffs its pet.
func auraCategory(aura *Aura, casters []*Unit) proto.ActionCategory {
	isConsumable, isMajorCooldown := false, false
	for _, caster := range casters {
		for _, spell := range caster.Spellbook {
			if !castsAura(spell, aura) {
				continue
			}
			isConsumable = isConsumable || spell.Flags.Matches(SpellFlagPotion|SpellFlagConjured)
			isMajorCooldown = isMajorCooldown || spell.Flags.Matches(SpellFlagMCD)
		}
	}
	return actionCategory(aura.Source, isConsumable, isMajorCooldown)
}

func (pet *Pet) GetMetadata() *proto.UnitMetadata {
	return pet.metadata(&pet.Unit, &pet.Owner.Unit)
}

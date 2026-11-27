package core

import (
	"slices"

	"github.com/wowsims/forever/sim/core/proto"
)

// The targets a versus effect names: any of the mob types or creature families.
type VersusTargets struct {
	MobTypes []proto.MobType
	Families []proto.CreatureFamily
}

// Whether the unit is of one of the mob types or families.
func (targets VersusTargets) Matches(unit *Unit) bool {
	return slices.Contains(targets.MobTypes, unit.MobType) || slices.Contains(targets.Families, unit.CreatureFamily)
}

// Whether the targets name no mob type and no family.
func (targets VersusTargets) IsEmpty() bool {
	return len(targets.MobTypes) == 0 && len(targets.Families) == 0
}

// Whether any of the units is one of the targets.
func (targets VersusTargets) AnyIn(units []*Unit) bool {
	return slices.ContainsFunc(units, targets.Matches)
}

// A factor a spell deals or restores against a target of any of the mob types or creature families,
// once however many of them the target is. The game files state the factor on a row and name the
// targets only in its tooltip.
type VersusBonus struct {
	VersusTargets
	Multiplier float64
}

// The factor against the target, 1 where it is none of the mob types and families or there is no bonus.
func (bonus *VersusBonus) Against(target *Unit) float64 {
	if bonus != nil && bonus.Matches(target) {
		return bonus.Multiplier
	}
	return 1
}

func (bonus *VersusBonus) namesFamilyOf(target *Unit) bool {
	return bonus != nil && slices.Contains(bonus.Families, target.CreatureFamily)
}

// The proc that casts a spell, where it fires only against the targets the bonus names or more often
// against them, for the combat log to name on the spell's hits on them. The bonus's Multiplier is how
// many times likelier the proc is against them, 0 where it fires against no other target.
type VersusProc struct {
	Source ActionID
	VersusBonus
}

type itemVersusBonus struct {
	itemID int32
	kind   proto.VersusEffectKind
	bonus  *VersusBonus
}

// The factor an item's spell deals or restores against the mob types and families, kept for the stats
// sheet. A damage spell applies it as its SpellConfig.VersusBonus; any other amount is multiplied by
// Against at the call site.
func (character *Character) NewItemVersusBonus(itemID int32, kind proto.VersusEffectKind, bonus VersusBonus) *VersusBonus {
	character.itemVersusBonuses = append(character.itemVersusBonuses, itemVersusBonus{itemID: itemID, kind: kind, bonus: &bonus})
	return &bonus
}

// The proc of an item's spell, with the item's factor kept for the stats sheet.
func (character *Character) NewItemVersusProc(itemID int32, kind proto.VersusEffectKind, bonus VersusBonus) *VersusProc {
	return &VersusProc{Source: ActionID{ItemID: itemID}, VersusBonus: *character.NewItemVersusBonus(itemID, kind, bonus)}
}

// The worn items' factors. An item swap registers the effects of the items it can swap in as well,
// so an item the character does not wear is left out.
func (character *Character) versusEffectsProto() []*proto.VersusEffect {
	var effects []*proto.VersusEffect
	for _, versus := range character.itemVersusBonuses {
		slot := slices.IndexFunc(character.Equipment[:], func(item Item) bool { return item.ID == versus.itemID })
		if slot < 0 {
			continue
		}
		effects = append(effects, &proto.VersusEffect{
			ItemId:           versus.itemID,
			ItemName:         character.Equipment[slot].Name,
			MobTypes:         versus.bonus.MobTypes,
			CreatureFamilies: versus.bonus.Families,
			Multiplier:       versus.bonus.Multiplier,
			Kind:             versus.kind,
		})
	}
	return effects
}

package core

import (
	"slices"

	"github.com/wowsims/forever/sim/core/proto"
)

// A factor a spell deals or restores against a target of any of the mob types or creature families,
// once however many of them the target is. The game files state the factor on a row and name the
// targets only in its tooltip.
type VersusBonus struct {
	Multiplier float64
	MobTypes   []proto.MobType
	Families   []proto.CreatureFamily
}

// The factor against the target, 1 where it is none of the mob types and families or there is no bonus.
func (bonus *VersusBonus) Against(target *Unit) float64 {
	if bonus != nil && (slices.Contains(bonus.MobTypes, target.MobType) || slices.Contains(bonus.Families, target.CreatureFamily)) {
		return bonus.Multiplier
	}
	return 1
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

// The worn items' factors. An item swap registers the effects of the items it can swap in as well,
// so an item the character does not wear is left out.
func (character *Character) versusEffectsProto() []*proto.VersusEffect {
	var effects []*proto.VersusEffect
	for _, versus := range character.itemVersusBonuses {
		if !slices.ContainsFunc(character.Equipment[:], func(item Item) bool { return item.ID == versus.itemID }) {
			continue
		}
		effects = append(effects, &proto.VersusEffect{
			ItemId:           versus.itemID,
			ItemName:         GetItemByID(versus.itemID).Name,
			MobTypes:         versus.bonus.MobTypes,
			CreatureFamilies: versus.bonus.Families,
			Multiplier:       versus.bonus.Multiplier,
			Kind:             versus.kind,
		})
	}
	return effects
}

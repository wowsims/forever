package shared

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

type versusFactor struct {
	index    uint8 // The EffectIndex whose BasePoints is the factor: the tooltip's $s<index+1>.
	mobTypes []proto.MobType
	families []proto.CreatureFamily
}

// The rows that state a factor against some targets, keyed by the row. No row names the targets, so
// they are typed here as each row's tooltip names them. The tooltips of 12731, 1302279 and 1294794
// also name Golems and Titan-forged, which no CreatureFamily holds, so those count Earth Elementals.
var versusFactors = map[int32]versusFactor{
	12731:   {index: 1, families: []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyEarthElemental}}, // Strength of Stone
	1302279: {index: 1, families: []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyEarthElemental}}, // Stoneshatter
	1294794: {index: 0, families: []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyEarthElemental}}, // Fist of Stone
	1293568: {index: 1, families: []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyPlant}},          // Grass Killer
	12257:   {index: 1, families: []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyPlant}},          // Breath of Fire
	17283:   {index: 1, families: []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyPlant}},          // Smokey's Lighter
	1295270: {index: 1, families: []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyPlant}},          // Prototype Pathcarver
	1295039: {index: 2, mobTypes: []proto.MobType{proto.MobType_MobTypeBeast}, // Serpentbane
		families: []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilySerpentine}},
	1318159: {index: 0, families: []proto.CreatureFamily{ // Ancient Power
		proto.CreatureFamily_CreatureFamilyNaga, proto.CreatureFamily_CreatureFamilySatyr}},
	1311505: {index: 1, families: []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyDinosaur}}, // Grounded Electroshocks
	14118:   {index: 1, mobTypes: []proto.MobType{proto.MobType_MobTypeDragonkin}},                     // Rend
	1299631: {index: 1, mobTypes: []proto.MobType{proto.MobType_MobTypeUndead}},                        // Smite
}

// The factor the row deals or restores against its targets, kept for the item on the stats sheet, or
// nil where the row states none.
func ItemVersusBonus(character *core.Character, itemID int32, kind proto.VersusEffectKind, row *spelldata.Spell) *core.VersusBonus {
	factor, ok := versusFactors[row.ID]
	if !ok {
		return nil
	}
	return character.NewItemVersusBonus(itemID, kind, core.VersusBonus{
		Multiplier: row.EffectAtIndex(factor.index).BaseValue(),
		MobTypes:   factor.mobTypes,
		Families:   factor.families,
	})
}

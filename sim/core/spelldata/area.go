package spelldata

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// Terrain area groups, each citing a tooltip; any other group is a single zone.
var areaTypeByGroup = map[int32]proto.AreaType{
	9161: proto.AreaType_AreaTypeForestGrassland,   // "in Forest and Grassland areas", Grovewalker 1306076
	9165: proto.AreaType_AreaTypeMountainous,       // "in Mountainous areas", Might of Bedrock 1308764
	9164: proto.AreaType_AreaTypeSnowy,             // "in Snowy areas", Bitter Cold 1308511
	9097: proto.AreaType_AreaTypeDesert,            // "in Desert areas", The Desert Rose 1294646
	9162: proto.AreaType_AreaTypeSwamp,             // "in Swamp areas", Sporebloom 1306672
	9163: proto.AreaType_AreaTypeWasteland,         // "in Wasteland areas", Fallen Grove Walker 1306632
	9202: proto.AreaType_AreaTypeHaunted,           // "in Haunted areas", Phantomspeaker Mask 1307334
	9324: proto.AreaType_AreaTypeCavernous,         // "in Cavernous or Underground areas", Royal Seal of Eldre'Thalas 1318480
	9326: proto.AreaType_AreaTypeStrongholdsCities, // "in Strongholds and Cities", Black Horn Necklace 1320332
	// Inferred: no spell requires 9203, but it holds the Blackrock zones, MC, BWL and Onyxia's Lair.
	9203: proto.AreaType_AreaTypeVolcanic,
}

func AreaTypeOfGroup(group int32) proto.AreaType {
	return areaTypeByGroup[group]
}

func (s *Spell) AreaType() proto.AreaType {
	return AreaTypeOfGroup(s.RequiredAreas)
}

// The factors the row's area bonus puts on its amount and on its duration, which an override supplies;
// neither names an area where the row has no bonus.
func (s *Spell) AreaBonus() (amount, duration core.AreaBonus) {
	var areas core.AreaTargets
	for _, group := range s.AreaBonusGroups {
		areas.Areas = append(areas.Areas, AreaTypeOfGroup(group))
	}
	return core.AreaBonus{AreaTargets: areas, Multiplier: float64(s.AreaMultiplier)},
		core.AreaBonus{AreaTargets: areas, Multiplier: float64(s.AreaDurationMultiplier)}
}

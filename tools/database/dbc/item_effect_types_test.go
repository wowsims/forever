package dbc

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
)

func TestItemEffectTypesReadEveryEffectAndItsTriggerChain(t *testing.T) {
	const (
		forestGroup = 9161
		desertGroup = 9097
		zoneGroup   = 6589
		undeadMask  = 1 << 5
		demonMask   = 1 << 2
		beastMask   = 1 << 0

		equipSpell, procSpell, triggeredSpell, loopSpell, zoneSpell = 9000001, 9000002, 9000003, 9000004, 9000005
		taggedItem, zoneItem, plainItem                             = 9100001, 9100002, 9100003
	)

	previous := dbcInstance
	dbcInstance = NewDBC()
	t.Cleanup(func() { dbcInstance = previous })

	dbcInstance.Spells = map[int]Spell{
		equipSpell:     {RequiredAreasID: forestGroup},
		procSpell:      {Description: "Deals 3 times as much damage to Naga and Satyrs."},
		triggeredSpell: {RequiredAreasID: desertGroup},
		loopSpell:      {},
		zoneSpell:      {RequiredAreasID: zoneGroup},
	}
	dbcInstance.SpellEffects = map[int]map[int]SpellEffect{
		equipSpell: {
			0: {EffectAura: dbcenums.A_MOD_MELEE_ATTACK_POWER_VERSUS, EffectMiscValues: []int{undeadMask}},
			1: {EffectAura: dbcenums.A_MOD_DAMAGE_DONE_CREATURE, EffectMiscValues: []int{beastMask}},
		},
		procSpell: {0: {EffectTriggerSpell: triggeredSpell}},
		triggeredSpell: {
			0: {EffectAura: dbcenums.A_MOD_DAMAGE_DONE_VERSUS, EffectMiscValues: []int{demonMask | undeadMask}},
			1: {EffectTriggerSpell: loopSpell},
		},
		loopSpell: {0: {EffectTriggerSpell: procSpell}},
	}
	dbcInstance.indexSpellEffects()
	dbcInstance.ItemEffectsByParentID = map[int][]ItemEffect{
		taggedItem: {{SpellID: equipSpell}, {SpellID: procSpell}},
		zoneItem:   {{SpellID: zoneSpell}},
	}

	areaTypes, mobTypes, families := ItemEffectTypes(taggedItem, false, noSetBonus)
	if want := []proto.AreaType{proto.AreaType_AreaTypeForestGrassland, proto.AreaType_AreaTypeDesert}; !slices.Equal(areaTypes, want) {
		t.Errorf("area types = %v, want %v: the equip spell's area and the triggered spell's", areaTypes, want)
	}
	if want := []proto.MobType{proto.MobType_MobTypeBeast, proto.MobType_MobTypeDemon, proto.MobType_MobTypeUndead}; !slices.Equal(mobTypes, want) {
		t.Errorf("mob types = %v, want %v: each once, from the three versus auras", mobTypes, want)
	}

	if want := []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyNaga, proto.CreatureFamily_CreatureFamilySatyr}; !slices.Equal(families, want) {
		t.Errorf("creature groups = %v, want %v: the ones the proc spell's tooltip names", families, want)
	}

	for name, itemID := range map[string]int32{"an item whose spell requires a single zone": zoneItem, "an item with no effects": plainItem} {
		if areaTypes, mobTypes, families := ItemEffectTypes(itemID, false, noSetBonus); len(areaTypes)+len(mobTypes)+len(families) != 0 {
			t.Errorf("%s got %v, %v and %v, want none", name, areaTypes, mobTypes, families)
		}
	}

	dbcInstance.Items[int(plainItem)] = Item{BonusStat: []int{ITEM_MOD_ATTACK_POWER_VS_UNDEAD, ITEM_MOD_SPELL_DAMAGE_VS_DEMON, ITEM_MOD_STAMINA}}
	if _, mobTypes, _ := ItemEffectTypes(plainItem, false, noSetBonus); !slices.Equal(mobTypes, []proto.MobType{proto.MobType_MobTypeDemon, proto.MobType_MobTypeUndead}) {
		t.Errorf("an item with no effect row and creature type stats got %v, want Demon and Undead", mobTypes)
	}
}

func TestTooltipTargetsReadsTheNamesAfterALead(t *testing.T) {
	const (
		undead    = proto.MobType_MobTypeUndead
		demon     = proto.MobType_MobTypeDemon
		beast     = proto.MobType_MobTypeBeast
		elemental = proto.MobType_MobTypeElemental
	)
	for _, tc := range []struct {
		tooltip  string
		mobTypes []proto.MobType
		families []proto.CreatureFamily
	}{
		{tooltip: "Smite your target for 112 Holy damage. Damage is tripled against Undead.", mobTypes: []proto.MobType{undead}},
		{tooltip: "Increases attack power by 150 when fighting Undead and Demons.", mobTypes: []proto.MobType{undead, demon}},
		{tooltip: "If there is an elemental nearby, gain an additional 20.", mobTypes: []proto.MobType{elemental}},
		{
			tooltip:  "Deals 2 times as much damage to Beast and Serpentine targets.",
			mobTypes: []proto.MobType{beast},
			families: []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilySerpentine},
		},
		{
			tooltip:  "Deals 3 times as much damage to Earth Elementals, Fire Elementals and Mountain Giants.",
			families: []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyEarthElemental, proto.CreatureFamily_CreatureFamilyFireElemental, proto.CreatureFamily_CreatureFamilyMountainGiant},
		},
		{tooltip: "Attacks against Dwarves are 2 times as likely to activate this effect.", families: []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyDwarf}},
		{tooltip: "Attacks against Undead have a chance to summon a Fire Elemental.", mobTypes: []proto.MobType{undead}},
		{tooltip: "Deals 40 Fire damage to all enemies. Transforms you into a Furbolg against your will."},
		{tooltip: "Increases damage against Demonic foes."},
		{tooltip: "Trolls and Orcs take 2 times as much damage.", families: []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyTroll, proto.CreatureFamily_CreatureFamilyOrc}},
		{tooltip: "Chance to strike back. Undead that strike you are 3 times as likely to trigger this effect.", mobTypes: []proto.MobType{undead}},
		{tooltip: "Deals 30 damage, tripled if attacking an Earth Elemental.", families: []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyEarthElemental}},
		{tooltip: "Mechanical targets take 2x damage.", mobTypes: []proto.MobType{proto.MobType_MobTypeMechanical}},
		{tooltip: "Tracks nearby creatures. Only works against Beasts."},
		{tooltip: "Undead are everywhere in the Plaguelands."},
	} {
		mobTypes, families := tooltipTargets(tc.tooltip)
		if !slices.Equal(mobTypes, tc.mobTypes) || !slices.Equal(families, tc.families) {
			t.Errorf("%q names %v and %v, want %v and %v", tc.tooltip, mobTypes, families, tc.mobTypes, tc.families)
		}
	}
}

func TestTooltipAreasReadsTheAreaNamesAfterIn(t *testing.T) {
	const (
		forest   = proto.AreaType_AreaTypeForestGrassland
		swamp    = proto.AreaType_AreaTypeSwamp
		cities   = proto.AreaType_AreaTypeStrongholdsCities
		cavern   = proto.AreaType_AreaTypeCavernous
		haunted  = proto.AreaType_AreaTypeHaunted
		wastes   = proto.AreaType_AreaTypeWasteland
		mountain = proto.AreaType_AreaTypeMountainous
	)
	for _, tc := range []struct {
		tooltip string
		areas   []proto.AreaType
	}{
		{tooltip: "Melee attacks have a chance to increase your Critical Strike chance by 2%. This effect is doubled in swamps.", areas: []proto.AreaType{swamp}},
		{tooltip: "This effect is doubled in Marsh and Swamp areas.", areas: []proto.AreaType{swamp, swamp}},
		{tooltip: "Increases Spirit by 10 in Cavernous or Underground areas.", areas: []proto.AreaType{cavern, cavern}},
		{tooltip: "This chance is doubled in Wasteland and Haunted areas.", areas: []proto.AreaType{wastes, haunted}},
		{tooltip: "Healing is doubled in Strongholds and Cities.", areas: []proto.AreaType{cities, cities}},
		{tooltip: "Gain an additional 5 in Forest and Grassland areas. This effect is doubled in Mountainous areas.", areas: []proto.AreaType{forest, forest, mountain}},
		{tooltip: "2% chance when struck in melee to gain a holy shield."},
		{tooltip: "Increases movement speed. While out of combat, gain an additional 10%."},
	} {
		if areas := tooltipAreas(tc.tooltip); !slices.Equal(areas, tc.areas) {
			t.Errorf("%q names %v, want %v", tc.tooltip, areas, tc.areas)
		}
	}
}

// The areas a tooltip names count only for an effect the sim models by hand: an item's own effects when
// the sim models the item, and each set bonus it models, by its piece count. A set bonus names no mob type.
func TestItemEffectTypesReadTooltipAreasOfModelledEffects(t *testing.T) {
	const (
		setID                                           = 9200001
		citiesBonus, versusBonus, swampBonus, itemSpell = 9000011, 9000012, 9000013, 9000014
		piece, item                                     = 9100011, 9100012
	)

	previous := dbcInstance
	dbcInstance = NewDBC()
	t.Cleanup(func() { dbcInstance = previous })

	dbcInstance.Spells = map[int]Spell{
		citiesBonus: {Description: "Chance on block to heal for 18. Healing is doubled in Strongholds and Cities."},
		versusBonus: {Description: "Deals 2 times as much damage to Undead."},
		swampBonus:  {Description: "Damage is doubled in swamps."},
		itemSpell:   {Description: "This effect is doubled in Snowy areas."},
	}
	dbcInstance.indexSpellEffects()
	dbcInstance.Items[piece] = Item{ItemSetId: setID}
	dbcInstance.ItemSetSpells[setID] = []ItemSetSpell{
		{ItemSetID: setID, Threshold: 3, SpellID: citiesBonus},
		{ItemSetID: setID, Threshold: 4, SpellID: versusBonus},
		{ItemSetID: setID, Threshold: 5, SpellID: swampBonus},
	}
	dbcInstance.ItemEffectsByParentID = map[int][]ItemEffect{item: {{SpellID: itemSpell}}}

	threeAndFour := func(pieces int32) bool { return pieces == 3 || pieces == 4 }
	areaTypes, mobTypes, families := ItemEffectTypes(piece, false, threeAndFour)
	if want := []proto.AreaType{proto.AreaType_AreaTypeStrongholdsCities}; !slices.Equal(areaTypes, want) || len(mobTypes)+len(families) != 0 {
		t.Errorf("the set piece names %v, %v and %v, want %v alone: the modelled bonuses' area", areaTypes, mobTypes, families, want)
	}
	if want := []proto.AreaType{proto.AreaType_AreaTypeSnowy}; !slices.Equal(firstOf(ItemEffectTypes(item, true, noSetBonus)), want) {
		t.Errorf("a modelled item names %v, want %v", firstOf(ItemEffectTypes(item, true, noSetBonus)), want)
	}
	if areaTypes := firstOf(ItemEffectTypes(item, false, noSetBonus)); len(areaTypes) != 0 {
		t.Errorf("an item the sim does not model names %v, want no area", areaTypes)
	}
}

func noSetBonus(int32) bool { return false }

func firstOf(areaTypes []proto.AreaType, _ []proto.MobType, _ []proto.CreatureFamily) []proto.AreaType {
	return areaTypes
}

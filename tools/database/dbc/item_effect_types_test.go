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
		equipSpell: {0: {EffectAura: dbcenums.A_MOD_MELEE_ATTACK_POWER_VERSUS, EffectMiscValues: []int{undeadMask}}},
		procSpell:  {0: {EffectTriggerSpell: triggeredSpell}},
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

	areaTypes, mobTypes, families := ItemEffectTypes(taggedItem)
	if want := []proto.AreaType{proto.AreaType_AreaTypeForestGrassland, proto.AreaType_AreaTypeDesert}; !slices.Equal(areaTypes, want) {
		t.Errorf("area types = %v, want %v: the equip spell's area and the triggered spell's", areaTypes, want)
	}
	if want := []proto.MobType{proto.MobType_MobTypeDemon, proto.MobType_MobTypeUndead}; !slices.Equal(mobTypes, want) {
		t.Errorf("mob types = %v, want %v: each once, from both versus auras", mobTypes, want)
	}

	if want := []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyNaga, proto.CreatureFamily_CreatureFamilySatyr}; !slices.Equal(families, want) {
		t.Errorf("creature groups = %v, want %v: the ones the proc spell's tooltip names", families, want)
	}

	for name, itemID := range map[string]int32{"an item whose spell requires a single zone": zoneItem, "an item with no effects": plainItem} {
		if areaTypes, mobTypes, families := ItemEffectTypes(itemID); len(areaTypes)+len(mobTypes)+len(families) != 0 {
			t.Errorf("%s got %v, %v and %v, want none", name, areaTypes, mobTypes, families)
		}
	}

	if _, mobTypes, _ := ItemEffectTypes(23206); !slices.Equal(mobTypes, tooltipMobTypes[23206]) {
		t.Errorf("an item with no effect row got %v, want %v", mobTypes, tooltipMobTypes[23206])
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
	} {
		mobTypes, families := tooltipTargets(tc.tooltip)
		if !slices.Equal(mobTypes, tc.mobTypes) || !slices.Equal(families, tc.families) {
			t.Errorf("%q names %v and %v, want %v and %v", tc.tooltip, mobTypes, families, tc.mobTypes, tc.families)
		}
	}
}

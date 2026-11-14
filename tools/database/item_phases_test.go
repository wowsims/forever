package database

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestInferPhases(t *testing.T) {
	var openRaid, openRaidFaction, unknownFaction int32
	for _, raid := range raids {
		if raid.open && raid.mapID != 0 {
			openRaid = raid.mapID
		}
		if len(raid.factions) > 0 {
			if raid.open {
				openRaidFaction = raid.factions[0]
			} else {
				unknownFaction = raid.factions[0]
			}
		}
	}
	if openRaid == 0 || openRaidFaction == 0 || unknownFaction == 0 {
		t.Fatal("raids lists no open raid with a map, no open raid with a faction, or no closed raid with a faction")
	}
	var unknownQuestZone, unknownItem int32
	for zone := range unknownQuestZones {
		unknownQuestZone = zone
	}
	for id := range unknownItems {
		unknownItem = id
	}
	const (
		openWorld, openRaidZone, unknownRaidZone, unknownRaidSubzone int32 = 10, 11, 12, 13
		unknownRaid                                                  int32 = 1 << 20
	)
	drop := func(zone int32) *proto.UIItemSource {
		return &proto.UIItemSource{Source: &proto.UIItemSource_Drop{Drop: &proto.DropSource{ZoneId: zone}}}
	}
	craft := func(spell int32) *proto.UIItemSource {
		return &proto.UIItemSource{Source: &proto.UIItemSource_Crafted{Crafted: &proto.CraftedSource{SpellId: spell}}}
	}
	quest := func(id int32) *proto.UIItemSource {
		return &proto.UIItemSource{Source: &proto.UIItemSource_Quest{Quest: &proto.QuestSource{Id: id}}}
	}
	rep := func(faction int32) *proto.UIItemSource {
		return &proto.UIItemSource{Source: &proto.UIItemSource_Rep{Rep: &proto.RepSource{RepFactionId: faction}}}
	}
	item := func(id, ilvl, set int32, sources ...*proto.UIItemSource) *proto.UIItem {
		return &proto.UIItem{Id: id, SetId: set, Sources: sources, ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {Ilvl: ilvl}}}
	}

	db := NewWowDatabase()
	for _, item := range []*proto.UIItem{
		item(100, phase1MaxIlvl, 0, drop(openWorld)),
		item(101, 1, 1, drop(unknownRaidZone)),
		item(102, 1, 0, drop(unknownRaidZone), drop(openWorld)),
		item(103, 1, 0, craft(500)),
		item(104, 1, 0, craft(501)),
		item(105, 1, 0, quest(600)),
		item(106, 1, 0, quest(601)),
		item(107, phase1MaxIlvl+1, 0, drop(openWorld)),
		item(108, phase1MaxIlvl+1, 0, drop(openRaidZone)),
		item(109, 1, 1),
		item(110, 1, 2),
		item(111, 1, 2, drop(openWorld)),
		item(112, 1, 2, drop(unknownRaidZone)),
		item(113, 1, 0, craft(502)),
		item(114, 1, 0, drop(openWorld), rep(unknownFaction)),
		item(115, 1, 0, craft(504)),
		item(116, 1, 0, drop(openRaidZone)),
		item(117, phase1MaxIlvl+1, 0, rep(openRaidFaction)),
		item(118, 1, 0, quest(602)),
		item(119, 1, 0, quest(603)),
		item(120, 1, 0, drop(unknownQuestZone)),
		item(121, phase1MaxIlvl+1, 3, drop(openWorld)),
		item(122, 1, 3, drop(openWorld)),
		item(123, phase1MaxIlvl+1, 4, drop(openWorld)),
		item(124, phase1MaxIlvl+1, 4, drop(openWorld)),
		item(125, 1, 0, drop(unknownRaidZone)),
	} {
		db.Items[item.Id] = item
	}
	enchants := map[int32]*proto.UIEnchant{}
	for effect, enchantItem := range map[int32]int32{1: 101, 2: 106, 3: 100, 4: 0, 5: 0} {
		enchants[effect] = &proto.UIEnchant{EffectId: effect, ItemId: enchantItem}
		db.Enchants[EnchantToDBKey(enchants[effect])] = enchants[effect]
	}

	InferPhases(db, PhaseInputs{
		Sources: map[int32][]*proto.UIItemSource{
			125: {drop(openWorld)},
			900: {drop(unknownRaidSubzone)},
			901: {drop(openWorld), rep(unknownFaction)},
			902: {craft(503)},
			903: {drop(unknownRaidZone)},
			904: {craft(505)},
			905: {craft(504)},
			906: {drop(unknownRaidZone)},
		},
		Reagents:       map[int32][]int32{500: {900}, 502: {902}, 503: {903}, 504: {904}, 505: {905}},
		Recipes:        map[int32][]int32{501: {901}},
		EnchantRecipes: map[EnchantDBKey][]int{EnchantToDBKey(enchants[5]): {906}},
		Quests: map[int32]*QuestieQuest{
			600: {Zone: unknownRaidZone},
			601: {Zone: openWorld, Items: []int32{101}},
			602: {Zone: unknownQuestZone},
			603: {Zone: openWorld, Items: []int32{unknownItem}},
		},
		RaidMaps: map[int32]int32{openRaidZone: openRaid, unknownRaidZone: unknownRaid, unknownRaidSubzone: unknownRaid},
	})

	for id, want := range map[int32]struct {
		phase int32
		why   string
	}{
		100: {phase1, "an open-world drop at the item level cap is phase 1"},
		101: {unknownPhase, "a drop in a raid phase 1 does not open is Unknown"},
		102: {phase1, "one open source is enough for phase 1"},
		103: {unknownPhase, "a craft with a reagent from an Unknown raid's subzone is Unknown"},
		104: {unknownPhase, "a craft whose only recipe needs an Unknown faction's reputation is Unknown"},
		105: {unknownPhase, "a quest filed under an Unknown raid is Unknown"},
		106: {unknownPhase, "a quest asking for an Unknown item is Unknown"},
		107: {unknownPhase, "an item above the item level cap is Unknown"},
		108: {phase1, "an open raid's drop stays phase 1 above the item level cap"},
		109: {unknownPhase, "an item with no source takes Unknown from its set when every sourced piece is Unknown"},
		110: {phase1, "an item with no source stays phase 1 when a sourced piece of its set is phase 1"},
		113: {unknownPhase, "a craft is Unknown through a crafted reagent whose own reagent is Unknown"},
		114: {unknownPhase, "an Unknown faction's reputation makes the item Unknown beside an open vendor or drop"},
		115: {phase1, "reagents that are only crafted from each other are not Unknown"},
		116: {phase1, "a drop in an open raid is phase 1"},
		117: {phase1, "an open raid's reputation reward stays phase 1 above the item level cap"},
		118: {unknownPhase, "a quest filed under a zone whose quests phase 1 does not hold is Unknown"},
		119: {unknownPhase, "a quest asking for an item of an event phase 1 does not hold is Unknown"},
		120: {phase1, "a drop in that zone keeps its own phase"},
		121: {phase1, "a piece above the item level cap follows its set when its other pieces are phase 1"},
		123: {unknownPhase, "a set whose every piece is above the item level cap stays Unknown"},
		125: {phase1, "an open source the inputs give and the database item does not show makes it phase 1"},
	} {
		if got := db.Items[id].Phase; got != want.phase {
			t.Errorf("item %d has phase %d, want %d: %s", id, got, want.phase, want.why)
		}
	}
	for effect, want := range map[int32]struct {
		phase int32
		why   string
	}{
		1: {unknownPhase, "an enchant whose item drops in an Unknown raid is Unknown"},
		2: {unknownPhase, "an enchant whose item is an Unknown quest's reward is Unknown"},
		3: {phase1, "an enchant whose item has an open source is phase 1"},
		4: {phase1, "an enchant with no item and no recipe is phase 1"},
		5: {unknownPhase, "an enchant whose only recipe is Unknown is Unknown"},
	} {
		if got := enchants[effect].Phase; got != want.phase {
			t.Errorf("enchant %d has phase %d, want %d: %s", effect, got, want.phase, want.why)
		}
	}
}

package database

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestInferPhases(t *testing.T) {
	var openRaid, openRaidFaction, futureFaction int32
	for _, raid := range raids {
		if raid.open && raid.mapID != 0 {
			openRaid = raid.mapID
		}
		if len(raid.factions) > 0 {
			if raid.open {
				openRaidFaction = raid.factions[0]
			} else {
				futureFaction = raid.factions[0]
			}
		}
	}
	if openRaid == 0 || openRaidFaction == 0 || futureFaction == 0 {
		t.Fatal("raids lists no open raid with a map, no open raid with a faction, or no closed raid with a faction")
	}
	var futureQuestZone, futureItem int32
	for zone := range futureQuestZones {
		futureQuestZone = zone
	}
	for id := range futureItems {
		futureItem = id
	}
	const (
		openWorld, openRaidZone, futureRaidZone, futureRaidSubzone int32 = 10, 11, 12, 13
		futureRaid                                                 int32 = 1 << 20
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
		item(101, 1, 1, drop(futureRaidZone)),
		item(102, 1, 0, drop(futureRaidZone), drop(openWorld)),
		item(103, 1, 0, craft(500)),
		item(104, 1, 0, craft(501)),
		item(105, 1, 0, quest(600)),
		item(106, 1, 0, quest(601)),
		item(107, phase1MaxIlvl+1, 0, drop(openWorld)),
		item(108, phase1MaxIlvl+1, 0, drop(openRaidZone)),
		item(109, 1, 1),
		item(110, 1, 2),
		item(111, 1, 2, drop(openWorld)),
		item(112, 1, 2, drop(futureRaidZone)),
		item(113, 1, 0, craft(502)),
		item(114, 1, 0, drop(openWorld), rep(futureFaction)),
		item(115, 1, 0, craft(504)),
		item(116, 1, 0, drop(openRaidZone)),
		item(117, phase1MaxIlvl+1, 0, rep(openRaidFaction)),
		item(118, 1, 0, quest(602)),
		item(119, 1, 0, quest(603)),
		item(120, 1, 0, drop(futureQuestZone)),
		item(121, phase1MaxIlvl+1, 3, drop(openWorld)),
		item(122, 1, 3, drop(openWorld)),
		item(123, phase1MaxIlvl+1, 4, drop(openWorld)),
		item(124, phase1MaxIlvl+1, 4, drop(openWorld)),
		item(125, 1, 0, drop(futureRaidZone)),
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
			900: {drop(futureRaidSubzone)},
			901: {drop(openWorld), rep(futureFaction)},
			902: {craft(503)},
			903: {drop(futureRaidZone)},
			904: {craft(505)},
			905: {craft(504)},
			906: {drop(futureRaidZone)},
		},
		Reagents:       map[int32][]int32{500: {900}, 502: {902}, 503: {903}, 504: {904}, 505: {905}},
		Recipes:        map[int32][]int32{501: {901}},
		EnchantRecipes: map[EnchantDBKey][]int{EnchantToDBKey(enchants[5]): {906}},
		Quests: map[int32]*QuestieQuest{
			600: {Zone: futureRaidZone},
			601: {Zone: openWorld, Items: []int32{101}},
			602: {Zone: futureQuestZone},
			603: {Zone: openWorld, Items: []int32{futureItem}},
		},
		RaidMaps: map[int32]int32{openRaidZone: openRaid, futureRaidZone: futureRaid, futureRaidSubzone: futureRaid},
	})

	for id, want := range map[int32]struct {
		phase int32
		why   string
	}{
		100: {phase1, "an open-world drop at the item level cap is phase 1"},
		101: {futurePhase, "a drop in a raid phase 1 does not open is Future"},
		102: {phase1, "one open source is enough for phase 1"},
		103: {futurePhase, "a craft with a reagent from a Future raid's subzone is Future"},
		104: {futurePhase, "a craft whose only recipe needs a Future faction's reputation is Future"},
		105: {futurePhase, "a quest filed under a Future raid is Future"},
		106: {futurePhase, "a quest asking for a Future item is Future"},
		107: {futurePhase, "an item above the item level cap is Future"},
		108: {phase1, "an open raid's drop stays phase 1 above the item level cap"},
		109: {futurePhase, "an item with no source takes Future from its set when every sourced piece is Future"},
		110: {phase1, "an item with no source stays phase 1 when a sourced piece of its set is phase 1"},
		113: {futurePhase, "a craft is Future through a crafted reagent whose own reagent is Future"},
		114: {futurePhase, "a Future faction's reputation makes the item Future beside an open vendor or drop"},
		115: {phase1, "reagents that are only crafted from each other are not Future"},
		116: {phase1, "a drop in an open raid is phase 1"},
		117: {phase1, "an open raid's reputation reward stays phase 1 above the item level cap"},
		118: {futurePhase, "a quest filed under a zone whose quests phase 1 does not hold is Future"},
		119: {futurePhase, "a quest asking for an item of an event phase 1 does not hold is Future"},
		120: {phase1, "a drop in that zone keeps its own phase"},
		121: {phase1, "a piece above the item level cap follows its set when its other pieces are phase 1"},
		123: {futurePhase, "a set whose every piece is above the item level cap stays Future"},
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
		1: {futurePhase, "an enchant whose item drops in a Future raid is Future"},
		2: {futurePhase, "an enchant whose item is a Future quest's reward is Future"},
		3: {phase1, "an enchant whose item has an open source is phase 1"},
		4: {phase1, "an enchant with no item and no recipe is phase 1"},
		5: {futurePhase, "an enchant whose only recipe is Future is Future"},
	} {
		if got := enchants[effect].Phase; got != want.phase {
			t.Errorf("enchant %d has phase %d, want %d: %s", effect, got, want.phase, want.why)
		}
	}
}

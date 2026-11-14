package database

import (
	"maps"
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestInferPhases(t *testing.T) {
	openRaid := slices.Collect(maps.Keys(phase1RaidMaps))[0]
	futureFaction := slices.Collect(maps.Keys(futureRepFactions))[0]
	openRaidFaction := slices.Collect(maps.Keys(phase1RaidFactions))[0]
	futureQuestZone := slices.Collect(maps.Keys(futureQuestZones))[0]
	futureItem := slices.Collect(maps.Keys(futureItems))[0]
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
	} {
		db.Items[item.Id] = item
	}
	enchant := func(id int32, sources ...*proto.UIItemSource) *proto.UIEnchant {
		return &proto.UIEnchant{EffectId: id, Sources: sources}
	}
	for _, enchant := range []*proto.UIEnchant{enchant(1, drop(futureRaidSubzone)), enchant(2, quest(601)), enchant(3, drop(openWorld)), enchant(4)} {
		db.Enchants[EnchantToDBKey(enchant)] = enchant
	}

	InferPhases(db, PhaseInputs{
		Sources: map[int32][]*proto.UIItemSource{
			900: {drop(futureRaidSubzone)},
			901: {drop(openWorld), rep(futureFaction)},
			902: {craft(503)},
			903: {drop(futureRaidZone)},
			904: {craft(505)},
			905: {craft(504)},
		},
		Reagents: map[int32][]int32{500: {900}, 502: {902}, 503: {903}, 504: {905}, 505: {904}},
		Recipes:  map[int32][]int32{501: {901}},
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
		101: {FuturePhase, "a drop in a raid phase 1 does not open is Future"},
		102: {phase1, "one open source is enough for phase 1"},
		103: {FuturePhase, "a craft with a reagent from a Future raid's subzone is Future"},
		104: {FuturePhase, "a craft whose only recipe needs a Future faction's reputation is Future"},
		105: {FuturePhase, "a quest filed under a Future raid is Future"},
		106: {FuturePhase, "a quest asking for a Future item is Future"},
		107: {FuturePhase, "an item above the item level cap is Future"},
		108: {phase1, "a phase 1 raid's drop stays phase 1 above the item level cap"},
		109: {FuturePhase, "an item with no source takes Future from its set when every sourced piece is Future"},
		110: {phase1, "an item with no source stays phase 1 when a sourced piece of its set is phase 1"},
		113: {FuturePhase, "a craft is Future through a crafted reagent whose own reagent is Future"},
		114: {FuturePhase, "a Future faction's reputation makes the item Future beside an open vendor or drop"},
		115: {phase1, "reagents that are only crafted from each other are not Future"},
		116: {phase1, "a drop in a phase 1 raid is phase 1"},
		117: {phase1, "a phase 1 raid's reputation reward stays phase 1 above the item level cap"},
		118: {FuturePhase, "a quest filed under a zone whose quests phase 1 does not hold is Future"},
		119: {FuturePhase, "a quest asking for an item of an event phase 1 does not hold is Future"},
		120: {phase1, "a drop in that zone keeps its own phase"},
	} {
		if got := db.Items[id].Phase; got != want.phase {
			t.Errorf("item %d has phase %d, want %d: %s", id, got, want.phase, want.why)
		}
	}
	for id, want := range map[int32]int32{1: FuturePhase, 2: FuturePhase, 3: phase1, 4: phase1} {
		if got := db.Enchants[EnchantToDBKey(&proto.UIEnchant{EffectId: id})].Phase; got != want {
			t.Errorf("enchant %d has phase %d, want %d: an enchant follows its sources as an item does, and none is phase 1", id, got, want)
		}
	}
}

package database

import (
	"maps"
	"slices"

	"github.com/wowsims/forever/sim/core/proto"
)

// Every item and enchant is phase 1, what can be had now, or Unknown, which holds the rest until the later phases are
// known. A source is Unknown when it lies in a raid phase 1 does not open, belongs to an event it does not hold, needs
// reputation with a faction such a raid rewards, or needs an Unknown item: a quest's objective, a craft's reagent, or
// every recipe that teaches the craft. An item is Unknown when every source it has is, or when it needs such
// reputation.
const (
	phase1       = int32(proto.ContentPhase_ContentPhase1)
	unknownPhase = int32(proto.ContentPhase_ContentPhaseUnknown)
)

// The raids phase 1 opens and the ones whose reputation it holds back, each with the factions that reputation comes
// from; a raid the game files hold no map for yet has map 0. Every raid not listed is Unknown, and one opens when it is
// listed with open set.
var raids = []struct {
	mapID    int32
	factions []int32
	open     bool
}{
	{249, nil, true},           // Onyxia's Lair.
	{0, []int32{2819}, true},   // Mount Hyjal, with The Watchers. Its map joins once the game files hold it.
	{309, []int32{270}, false}, // Zul'Gurub, with the Zandalar Tribe.
	{531, []int32{910}, false}, // Ahn'Qiraj Temple, with the Brood of Nozdormu.
}

// Phase 1 opens neither the Ahn'Qiraj war effort nor the Scourge Invasion. The quests filed under Silithus are the war
// effort's Cenarion Hold quests. The zone's drops and vendors keep their own phase, since Valor's Rest has sold its
// patterns since launch. The invasion's rewards cost the runes and insignias its mobs drop everywhere while it runs, so
// those items are Unknown whatever their sources say.
var unknownQuestZones = map[int32]bool{
	1377: true, // Silithus.
}
var unknownItems = map[int32]bool{
	22484: true, // Necrotic Rune.
	22523: true, // Insignia of the Dawn.
	22524: true, // Insignia of the Crusade.
}

// The highest item level phase 1 rewards outside its raids: its dungeons, Onyxia's quest rewards and Forever's own
// crafted epics. An item above it is Unknown unless an open raid drops it, an open raid's faction sells it, or every
// piece of its set the item level does not make Unknown is phase 1.
const phase1MaxIlvl = 65

// What InferPhases reads besides the database's items.
type PhaseInputs struct {
	Sources        map[int32][]*proto.UIItemSource // Every source the inputs give each item, the ones the database does not show included.
	Reagents       map[int32][]int32               // The reagent items of each spell.
	Recipes        map[int32][]int32               // The items that teach each spell.
	EnchantRecipes map[EnchantDBKey][]int          // The items that teach each enchant.
	Quests         map[int32]*QuestieQuest         // The reward quests by id.
	RaidMaps       map[int32]int32                 // The raid map of each raid zone.
}

// Sets the phase of every item and enchant in db, from their sources.
func InferPhases(db *WowDatabase, inputs PhaseInputs) {
	p := newPhaser(db, inputs)
	p.findUnknownItems()
	for _, item := range db.Items {
		item.Phase = phase1
		if p.unknown[item.Id] || item.ScalingOptions[0].Ilvl > phase1MaxIlvl && !p.openRaidSource(item.Id) {
			item.Phase = unknownPhase
		}
	}
	// The item level 66 legs of the dungeon set 0.5 sets follow their phase 1 sets, and the Edge of Madness charms,
	// which have no source, their Unknown Zul'Gurub sets.
	followSet(db, phase1, func(item *proto.UIItem) bool { return item.Phase == unknownPhase && !p.unknown[item.Id] })
	followSet(db, unknownPhase, func(item *proto.UIItem) bool { return len(item.Sources) == 0 })
	for key, enchant := range db.Enchants {
		var sources []*proto.UIItemSource
		for _, id := range append([]int{int(enchant.ItemId)}, inputs.EnchantRecipes[key]...) {
			sources = append(sources, p.sources[int32(id)]...)
		}
		enchant.Phase = phase1
		if p.allUnknown(sources) {
			enchant.Phase = unknownPhase
		}
	}
}

// Gives phase to each item follows picks when the pieces of its set follows does not pick, one at least, all have it.
func followSet(db *WowDatabase, phase int32, follows func(*proto.UIItem) bool) {
	setHasPhase := map[int32]bool{}
	for _, item := range db.Items {
		if all, seen := setHasPhase[item.SetId]; item.SetId != 0 && !follows(item) && (all || !seen) {
			setHasPhase[item.SetId] = item.Phase == phase
		}
	}
	for _, item := range db.Items {
		if follows(item) && setHasPhase[item.SetId] {
			item.Phase = phase
		}
	}
}

type phaser struct {
	inputs          PhaseInputs
	sources         map[int32][]*proto.UIItemSource // Each item's sources: the database's own, then every one the inputs give.
	unknown         map[int32]bool                  // The Unknown items: unknownItems, and the items allUnknown judges Unknown.
	openMaps        map[int32]bool
	openFactions    map[int32]bool
	unknownFactions map[int32]bool
}

func newPhaser(db *WowDatabase, inputs PhaseInputs) *phaser {
	p := &phaser{
		inputs:          inputs,
		sources:         maps.Clone(inputs.Sources),
		unknown:         maps.Clone(unknownItems),
		openMaps:        map[int32]bool{},
		openFactions:    map[int32]bool{},
		unknownFactions: map[int32]bool{},
	}
	if p.sources == nil {
		p.sources = map[int32][]*proto.UIItemSource{}
	}
	for id, item := range db.Items {
		p.sources[id] = append(slices.Clip(item.Sources), inputs.Sources[id]...)
	}
	for _, raid := range raids {
		if raid.open && raid.mapID != 0 {
			p.openMaps[raid.mapID] = true
		}
		for _, faction := range raid.factions {
			p.openFactions[faction] = raid.open
			p.unknownFactions[faction] = !raid.open
		}
	}
	return p
}

// Grows the Unknown items until none joins them, so that each item is judged on the others' final answers whatever
// the order. Items that are only to be had through each other stay out.
func (p *phaser) findUnknownItems() {
	for changed := true; changed; {
		changed = false
		for id, sources := range p.sources {
			if !p.unknown[id] && p.allUnknown(sources) {
				p.unknown[id], changed = true, true
			}
		}
	}
}

func (p *phaser) itemUnknown(id int32) bool {
	return p.unknown[id]
}

// A reputation source is the standing a vendor asks for on top of the vendor, so an Unknown faction's alone decides.
// No source at all is not Unknown.
func (p *phaser) allUnknown(sources []*proto.UIItemSource) bool {
	return slices.ContainsFunc(sources, func(source *proto.UIItemSource) bool { return p.unknownFactions[source.GetRep().GetRepFactionId()] }) ||
		every(sources, p.sourceUnknown)
}

func (p *phaser) sourceUnknown(source *proto.UIItemSource) bool {
	switch source := source.Source.(type) {
	case *proto.UIItemSource_Drop:
		return p.unknownZone(source.Drop.ZoneId)
	case *proto.UIItemSource_SoldBy:
		return p.unknownZone(source.SoldBy.ZoneId)
	case *proto.UIItemSource_Quest:
		quest := p.inputs.Quests[source.Quest.Id]
		return quest != nil && (p.unknownZone(quest.Zone) || unknownQuestZones[quest.Zone] || slices.ContainsFunc(quest.Items, p.itemUnknown))
	case *proto.UIItemSource_Crafted:
		return slices.ContainsFunc(p.inputs.Reagents[source.Crafted.SpellId], p.itemUnknown) || every(p.inputs.Recipes[source.Crafted.SpellId], p.itemUnknown)
	}
	// An open faction's reputation, and PvP.
	return false
}

func (p *phaser) unknownZone(zone int32) bool {
	raid, ok := p.inputs.RaidMaps[zone]
	return ok && !p.openMaps[raid]
}

func (p *phaser) openRaidSource(id int32) bool {
	return slices.ContainsFunc(p.sources[id], func(source *proto.UIItemSource) bool {
		return p.openMaps[p.inputs.RaidMaps[source.GetDrop().GetZoneId()]] || p.openFactions[source.GetRep().GetRepFactionId()]
	})
}

// Whether items holds one and f holds for each.
func every[T any](items []T, f func(T) bool) bool {
	return len(items) > 0 && !slices.ContainsFunc(items, func(item T) bool { return !f(item) })
}

// Every source the game files, AtlasLoot and QuestieDB give each item. QuestieDB's drops beside an AtlasLoot drop count
// here: Blood of the Mountain, a reagent the database does not hold, is a Molten Core drop in AtlasLoot and a world drop
// in QuestieDB.
func PhaseSources(inputs ItemSourceInputs) map[int32][]*proto.UIItemSource {
	ids := map[int32]bool{}
	for id := range inputs.Crafted {
		ids[int32(id)] = true
	}
	for id := range inputs.Rep {
		ids[int32(id)] = true
	}
	for id := range inputs.Atlas.Items {
		ids[id] = true
	}
	for id := range inputs.Questie.Items {
		ids[id] = true
	}
	sources := make(map[int32][]*proto.UIItemSource, len(ids))
	for id := range ids {
		sources[id] = slices.Concat(inputs.gameFilesSources(int(id)), inputs.Atlas.Items[id].GetSources(), inputs.Questie.Items[id].GetSources())
	}
	return sources
}

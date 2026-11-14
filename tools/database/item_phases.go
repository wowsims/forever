package database

import (
	"maps"
	"slices"

	"github.com/wowsims/forever/sim/core/proto"
)

// Every item and enchant is phase 1, what can be had now, or FuturePhase, which holds the rest until the later phases
// are known. A source is Future when it lies in a raid phase 1 does not open, belongs to an event it does not hold,
// needs reputation with a faction such a raid rewards, or needs a Future item: a quest's objective, a craft's reagent,
// or every recipe that teaches the craft. An item is Future when every source it has is, or when it needs such
// reputation.
const (
	phase1      int32 = 1
	FuturePhase int32 = 2
)

var phase1RaidMaps = map[int32]bool{
	249: true, // Onyxia's Lair. Mount Hyjal's raid joins it once the game files hold its map.
}
var phase1RaidFactions = map[int32]bool{
	2819: true, // The Watchers, of Mount Hyjal.
}
var futureRepFactions = map[int32]bool{
	270: true, // Zandalar Tribe, of Zul'Gurub.
	910: true, // Brood of Nozdormu, of Ahn'Qiraj.
}

// Phase 1 opens neither the Ahn'Qiraj war effort nor the Scourge Invasion. The quests filed under Silithus are the war
// effort's Cenarion Hold quests. The zone's drops and vendors keep their own phase, since Valor's Rest has sold its
// patterns since launch. The invasion's rewards cost the runes and insignias its mobs drop everywhere while it runs, so
// those items are Future whatever their sources say.
var futureQuestZones = map[int32]bool{
	1377: true, // Silithus.
}
var futureItems = map[int32]bool{
	22484: true, // Necrotic Rune.
	22523: true, // Insignia of the Dawn.
	22524: true, // Insignia of the Crusade.
}

// The highest item level phase 1 rewards outside its raids: its dungeons, Onyxia's quest rewards and Forever's own
// crafted epics. An item above it is Future unless a phase 1 raid drops it, its faction's reputation buys it, or every
// other piece of its set is phase 1.
const phase1MaxIlvl = 65

// What InferPhases reads besides the database's items.
type PhaseInputs struct {
	Sources  map[int32][]*proto.UIItemSource // Every item's sources in the inputs, the database's own and the items it leaves out.
	Reagents map[int32][]int32               // The reagent items of each spell.
	Recipes  map[int32][]int32               // The items that teach each spell.
	Quests   map[int32]*QuestieQuest         // The reward quests by id.
	RaidMaps map[int32]int32                 // The raid map of each raid zone.
}

// Sets the phase of every item and enchant in db, from their sources.
func InferPhases(db *WowDatabase, inputs PhaseInputs) {
	p := &phaser{db: db, inputs: inputs, future: map[int32]bool{}}
	p.findFutureItems()
	capped := map[int32]bool{}
	for _, item := range db.Items {
		item.Phase = phase1
		if p.future[item.Id] {
			item.Phase = FuturePhase
		} else if item.ScalingOptions[0].Ilvl > phase1MaxIlvl && !p.phase1RaidSource(item) {
			item.Phase = FuturePhase
			capped[item.Id] = true
		}
	}
	// A piece the item level alone makes Future follows its set when every other piece is phase 1, as the item level
	// 66 legs of the dungeon set 0.5 sets do.
	setPhase1 := map[int32]bool{}
	for _, item := range db.Items {
		if open, seen := setPhase1[item.SetId]; item.SetId != 0 && !capped[item.Id] && (open || !seen) {
			setPhase1[item.SetId] = item.Phase == phase1
		}
	}
	for id := range capped {
		if setPhase1[db.Items[id].SetId] {
			db.Items[id].Phase = phase1
		}
	}
	// An item with no source is Future when every piece of its set that has one is, as the Edge of Madness charms of
	// the Zul'Gurub sets.
	setFuture := map[int32]bool{}
	for _, item := range db.Items {
		if future, seen := setFuture[item.SetId]; item.SetId != 0 && len(item.Sources) > 0 && (future || !seen) {
			setFuture[item.SetId] = item.Phase == FuturePhase
		}
	}
	for _, item := range db.Items {
		if len(item.Sources) == 0 && setFuture[item.SetId] {
			item.Phase = FuturePhase
		}
	}
	for _, enchant := range db.Enchants {
		enchant.Phase = phase1
		if p.allFuture(enchant.Sources) {
			enchant.Phase = FuturePhase
		}
	}
}

type phaser struct {
	db     *WowDatabase
	inputs PhaseInputs
	future map[int32]bool // The items whose every source is Future.
}

// A database item's sources with the inputs', since the item shows only some of them: an AtlasLoot drop hides
// QuestieDB's, which can name an open zone (Blood of the Mountain drops in Molten Core and Blackrock Depths).
func (p *phaser) sources(id int32) []*proto.UIItemSource {
	if item, ok := p.db.Items[id]; ok {
		return append(slices.Clip(item.Sources), p.inputs.Sources[id]...)
	}
	return p.inputs.Sources[id]
}

// Grows the Future items until none joins them, so that each item is judged on the others' final answers whatever
// the order. Items that are only to be had through each other stay out.
func (p *phaser) findFutureItems() {
	maps.Copy(p.future, futureItems)
	ids := slices.Collect(maps.Keys(p.db.Items))
	for id := range p.inputs.Sources {
		if _, ok := p.db.Items[id]; !ok {
			ids = append(ids, id)
		}
	}
	for changed := true; changed; {
		changed = false
		for _, id := range ids {
			if !p.future[id] && p.allFuture(p.sources(id)) {
				p.future[id] = true
				changed = true
			}
		}
	}
}

func (p *phaser) itemFuture(id int32) bool {
	return p.future[id]
}

// A reputation source is the standing a vendor asks for on top of the vendor, so a Future faction's alone decides.
// No source at all is not Future.
func (p *phaser) allFuture(sources []*proto.UIItemSource) bool {
	if slices.ContainsFunc(sources, func(source *proto.UIItemSource) bool { return futureRepFactions[source.GetRep().GetRepFactionId()] }) {
		return true
	}
	return len(sources) > 0 && !slices.ContainsFunc(sources, func(source *proto.UIItemSource) bool { return !p.sourceFuture(source) })
}

func (p *phaser) sourceFuture(source *proto.UIItemSource) bool {
	switch source := source.Source.(type) {
	case *proto.UIItemSource_Drop:
		return p.futureZone(source.Drop.ZoneId)
	case *proto.UIItemSource_SoldBy:
		return p.futureZone(source.SoldBy.ZoneId)
	case *proto.UIItemSource_Quest:
		quest := p.inputs.Quests[source.Quest.Id]
		return quest != nil && (p.futureZone(quest.Zone) || futureQuestZones[quest.Zone] || slices.ContainsFunc(quest.Items, p.itemFuture))
	case *proto.UIItemSource_Crafted:
		recipes := p.inputs.Recipes[source.Crafted.SpellId]
		return slices.ContainsFunc(p.inputs.Reagents[source.Crafted.SpellId], p.itemFuture) ||
			len(recipes) > 0 && !slices.ContainsFunc(recipes, func(id int32) bool { return !p.itemFuture(id) })
	}
	// An open faction's reputation, and PvP.
	return false
}

func (p *phaser) futureZone(zone int32) bool {
	raid, ok := p.inputs.RaidMaps[zone]
	return ok && !phase1RaidMaps[raid]
}

func (p *phaser) phase1RaidSource(item *proto.UIItem) bool {
	return slices.ContainsFunc(item.Sources, func(source *proto.UIItemSource) bool {
		drop := source.GetDrop()
		return drop != nil && phase1RaidMaps[p.inputs.RaidMaps[drop.ZoneId]] || phase1RaidFactions[source.GetRep().GetRepFactionId()]
	})
}

// Every source the game files, AtlasLoot and QuestieDB give each item.
func PhaseSources(crafted map[int][]*proto.CraftedSource, rep map[int][]*proto.RepSource, atlasLoot, questie *WowDatabase) map[int32][]*proto.UIItemSource {
	sources := map[int32][]*proto.UIItemSource{}
	for id, item := range atlasLoot.Items {
		sources[id] = append(sources[id], item.Sources...)
	}
	for id, item := range questie.Items {
		sources[id] = append(sources[id], item.Sources...)
	}
	for id, list := range crafted {
		for _, craft := range list {
			sources[int32(id)] = append(sources[int32(id)], &proto.UIItemSource{Source: &proto.UIItemSource_Crafted{Crafted: craft}})
		}
	}
	for id, list := range rep {
		for _, standing := range list {
			sources[int32(id)] = append(sources[int32(id)], &proto.UIItemSource{Source: &proto.UIItemSource_Rep{Rep: standing}})
		}
	}
	return sources
}

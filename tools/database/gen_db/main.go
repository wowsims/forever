package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"maps"
	"os"
	"runtime/debug"
	"slices"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/wowsims/forever/sim"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	_ "github.com/wowsims/forever/sim/encounters" // Needed for preset encounters.
	"github.com/wowsims/forever/tools"
	"github.com/wowsims/forever/tools/database"
	"github.com/wowsims/forever/tools/database/dbc"
	googleProto "google.golang.org/protobuf/proto"
)

// To do a full re-scrape, delete the previous output file first.
// go run ./tools/database/gen_db -outDir=assets -gen=atlasloot
// go run ./tools/database/gen_db -outDir=assets -gen=questie
// go run ./tools/database/gen_db -outDir=assets -gen=db

var outDir = flag.String("outDir", "assets", "Path to output directory for writing generated .go files.")
var genAsset = flag.String("gen", "", "Asset to generate. Valid values are 'db', 'atlasloot', 'questie' and 'go-to-ts'")
var dbPath = flag.String("dbPath", "./tools/database/wowsims.db", "Location of the wowsims.db file produced by tools/db2tool")
var classes = flag.String("classes", "", "Write the talent files of these classes only (comma-separated package names, all or none, as gen_spelldata -classes); the item database is always written")

func main() {
	flag.Parse()

	database.DatabasePath = *dbPath

	if *outDir == "" {
		panic("outDir flag is required!")
	}

	dbDir := fmt.Sprintf("%s/database", *outDir)
	inputsDir := fmt.Sprintf("%s/db_inputs", *outDir)
	talentTreesDir = fmt.Sprintf("%s/../../ui/sim/talents/trees", inputsDir)

	if *genAsset == "go-to-ts" {
		if err := database.GenerateCharacterConstantsTSFile(); err != nil {
			log.Fatalf("failed to generate capabilities TS file: %v", err)
		}
		if err := database.GenerateBulkSimConstantsTSFile(); err != nil {
			log.Fatalf("failed to generate bulk sim constants TS file: %v", err)
		}
		if err := database.GenerateBulkSimTuningConstantsTSFile(); err != nil {
			log.Fatalf("failed to generate bulk sim tuning constants TS file: %v", err)
		}
		if err := database.GenerateSimTablesTSFile(); err != nil {
			log.Fatalf("failed to generate sim tables TS file: %v", err)
		}
		if err := database.GenerateStatWeightConstantsTSFile(); err != nil {
			log.Fatalf("failed to generate stat weight constants TS file: %v", err)
		}
		return
	} else if *genAsset == "atlasloot" {
		helper, err := database.NewDBHelper()
		if err != nil {
			log.Fatalf("failed to initialize database: %v", err)
		}
		defer helper.Close()

		db := database.ReadAtlasLootData(helper)
		db.WriteJson(fmt.Sprintf("%s/atlasloot_db.json", inputsDir))
		return
	} else if *genAsset == "questie" {
		database.ReadQuestieData().WriteJson(fmt.Sprintf("%s/questie_db.json", inputsDir))
		return
	} else if *genAsset != "db" {
		panic("Invalid gen value")
	}
	helper, err := database.NewDBHelper()
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer helper.Close()

	only, behind, err := database.ResolveSpellDataGroups(helper, *classes, false, true)
	if err != nil {
		log.Fatalf("spell data groups: %v", err)
	}
	for _, line := range behind {
		fmt.Fprintf(os.Stderr, "talent files kept, not ported: %s\n", line)
	}

	// The item database and the talent files of the groups taken are written from the database, so one
	// behind the committed inputs in hotfixes would drop the items and talents only a hotfix adds.
	if err := database.RequireCommittedHotfixes(helper, only); err != nil {
		log.Fatalf("hotfixes: %v", err)
	}

	if err := database.RunOverrides(helper, "tools/database/overrides"); err != nil {
		log.Fatalf("failed to run overrides: %v", err)
	}

	// The extractions below only read the (post-override) database and write
	// disjoint files, so they run concurrently; everything after the Wait
	// consumes their results.
	var (
		consumables     []dbc.Consumable
		dropSources     map[int][]*proto.DropSource
		names           map[int]string
		craftingSources map[int][]*proto.CraftedSource
		repSources      map[int][]*proto.RepSource
		factionNames    map[int32]string
		atlaslootDB     *database.WowDatabase
		questieDB       *database.WowDatabase
		raidZones       map[int32]bool
		areaTableIDs    map[int32]bool
		iconsMap        map[int]string
		icons           map[int]database.SpellIcon
	)
	var g errgroup.Group
	g.Go(func() error { _, err := database.LoadAndWriteRawRandomSuffixes(helper, inputsDir); return err })
	g.Go(func() error {
		_, err := database.LoadAndWriteRawItems(helper, database.SimItemFilter(core.CharacterLevel), inputsDir)
		return err
	})
	g.Go(func() error { _, err := database.LoadAndWriteRandomPropAllocations(helper, inputsDir); return err })
	g.Go(func() error { _, err := database.LoadAndWriteRawEnchants(helper, inputsDir); return err })
	g.Go(func() error { _, err := database.LoadAndWriteRawSpellEffects(helper, inputsDir); return err })
	g.Go(func() error { _, err := database.LoadAndWriteItemDamageTables(helper, inputsDir); return err })
	g.Go(func() error { _, err := database.LoadAndWriteItemArmorTotal(helper, inputsDir); return err })
	g.Go(func() error { _, err := database.LoadAndWriteItemArmorQuality(helper, inputsDir); return err })
	g.Go(func() error { _, err := database.LoadAndWriteItemArmorShield(helper, inputsDir); return err })
	g.Go(func() error { _, err := database.LoadAndWriteArmorLocation(helper, inputsDir); return err })
	g.Go(func() error { _, err := database.LoadAndWriteItemEffects(helper, inputsDir); return err })
	g.Go(func() error { _, err := database.LoadAndWriteItemSetSpells(helper, inputsDir); return err })
	g.Go(func() error { _, err := database.LoadAndWriteSpells(helper, inputsDir); return err })
	g.Go(func() (err error) { consumables, err = database.LoadAndWriteConsumables(helper, inputsDir); return })
	g.Go(func() (err error) {
		dropSources, names, err = database.LoadAndWriteDropSources(helper, inputsDir)
		return
	})
	g.Go(func() (err error) { icons, err = database.LoadSpellIcons(helper); return })
	g.Go(func() error { craftingSources = database.LoadCraftedItems(helper); return nil })
	g.Go(func() (err error) { repSources, factionNames, err = database.LoadRepItems(helper); return })
	g.Go(func() error {
		//Todo: See if we cant get rid of these as well
		atlaslootDB = database.ReadDatabaseFromJson(tools.ReadFile(fmt.Sprintf("%s/atlasloot_db.json", inputsDir)))
		return nil
	})
	g.Go(func() error {
		questieDB = database.ReadDatabaseFromJson(tools.ReadFile(fmt.Sprintf("%s/questie_db.json", inputsDir)))
		return nil
	})
	g.Go(func() (err error) { raidZones, err = database.LoadRaidZoneIDs(helper); return })
	g.Go(func() (err error) { areaTableIDs, err = database.LoadAreaTableIDs(helper); return })
	g.Go(func() (err error) { iconsMap, err = database.LoadArtTexturePaths(database.ListfilePath); return })
	if err := g.Wait(); err != nil {
		panic(fmt.Sprintf("Error loading DBC data %v", err))
	}

	pruneAtlasLootDrops(atlaslootDB, raidZones, areaTableIDs)
	pruneQuestieVendors(questieDB, areaTableIDs)

	db := database.NewWowDatabase()
	db.Encounters = core.PresetEncounters

	var instance = dbc.GetDBC()
	instance.LoadSpellScaling()
	instance.LoadShieldBlockValues()
	database.GenerateProtos(instance, db, only)

	processItems(instance, iconsMap, names, dropSources, craftingSources, repSources, db)

	for _, enchant := range instance.Enchants {
		parsed := enchant.ToProto()

		if parsed.Icon == "" {
			parsed.Icon = strings.ToLower(database.GetIconName(iconsMap, enchant.FDID))
		}
		if !enchant.IsLive {
			nonLiveEnchants[database.EnchantToDBKey(parsed)] = struct{}{}
		}
		db.MergeEnchant(parsed)
	}

	for _, atlaslootItem := range atlaslootDB.Items {
		if dbItem, ok := db.Items[atlaslootItem.Id]; ok {
			firstDBRepFaction := int32(-1)

			for _, dbItemSource := range dbItem.Sources {
				if dbRepSource := dbItemSource.GetRep(); dbRepSource != nil {
					firstDBRepFaction = dbRepSource.RepFactionId
					break
				}
			}

			atlaslootItem.Sources = core.FilterSlice(atlaslootItem.Sources, func(atlaslootItemSource *proto.UIItemSource) bool {
				atlaslootRepSource := atlaslootItemSource.GetRep()
				return (atlaslootRepSource == nil) || (atlaslootRepSource.RepFactionId != firstDBRepFaction)
			})

			db.MergeItem(atlaslootItem)
		}
	}
	appendQuestieSources(db, questieDB)

	consumableRoutes := map[int]*database.ConsumableRouting{}
	for _, consumable := range consumables {
		routing := database.RouteConsumable(&consumable, instance)
		consumableRoutes[consumable.Id] = routing
		if !database.ShipsConsumable(routing) {
			continue
		}
		db.MergeConsumable(database.ShippedConsumable(&consumable, routing, instance))
	}

	for _, consumable := range database.ConsumableOverrides {
		db.MergeConsumable(consumable)
	}
	if err := database.AddUnparsedStats(db); err != nil {
		log.Fatalf("failed to add the stats the parse leaves out: %v", err)
	}
	if err := database.AddPotionClasses(db); err != nil {
		log.Fatalf("failed to classify the potions: %v", err)
	}

	db.MergeItems(database.ItemOverrides)
	database.ApplyItemSetClasses(db)
	db.MergeEnchants(database.EnchantOverrides)
	ApplyGlobalFilters(db)
	database.AddConsumableIcons(db, consumables, instance, iconsMap)
	// The PvP source follows from the item's name, so the names must be final here.
	for _, item := range db.Items {
		database.InferPvpSource(item)
	}
	leftovers := db.Clone()
	ApplyNonSimmableFilters(leftovers)
	applyRepSources(leftovers, factionNames)
	leftovers.WriteBinaryAndJson(fmt.Sprintf("%s/leftover_db.bin", dbDir), fmt.Sprintf("%s/leftover_db.json", dbDir))
	ApplySimmableFilters(db)

	database.GenerateItemEffects(instance, db)
	database.GenerateEnchantEffects(instance, db)
	if err := database.GenerateConsumableRegistrations(instance, db, consumableRoutes); err != nil {
		log.Fatalf("failed to generate the consumable registrations: %v", err)
	}
	if err := database.GenerateConsumablesTSFile(db); err != nil {
		log.Fatalf("failed to generate the consumables TS file: %v", err)
	}
	if err := database.GenerateMissingEffectsFile(instance, db); err != nil {
		log.Fatalf("failed to generate the missing effects TS file: %v", err)
	}
	database.GenerateItemEffectRandomPropPoints(instance, db)

	// Disbabled due to duplicate SpellIDs
	// Our enchants already have the icon saved so this is redundant
	// for _, enchant := range db.Enchants {
	// 	if enchant.ItemId != 0 {
	// 		db.AddItemIcon(enchant.ItemId, enchant.Icon, enchant.Name)
	// 	}
	// 	if enchant.SpellId != 0 {
	// 		db.AddSpellIcon(enchant.SpellId, enchant.Icon, enchant.Name)
	// 	}
	// }

	for _, consume := range db.Consumables {
		if len(consume.EffectIds) > 0 {
			for _, se := range consume.EffectIds {
				effect := instance.SpellEffectsById[int(se)]
				db.MergeEffect(effect.ToProto())
			}
		}
	}

	for _, item := range db.Items {
		for _, id := range item.RandomSuffixOptions {
			if _, exists := db.RandomSuffixes[id]; !exists {
				db.RandomSuffixes[id] = instance.RandomSuffix[int(id)].ToProto()
			}
		}
	}

	addSpellIcons(db, database.SharedSpellsIcons, icons, iconsMap)

	for _, group := range GetAllTalentSpellIds(&inputsDir) {
		addSpellIcons(db, group, icons, iconsMap)
	}

	for _, group := range GetAllRotationSpellIds() {
		addSpellIcons(db, group, icons, iconsMap)
	}

	craftedSpellIds := []int32{}
	for _, item := range db.Items {
		for _, source := range item.Sources {
			if crafted := source.GetCrafted(); crafted != nil {
				craftedSpellIds = append(craftedSpellIds, crafted.SpellId)

				// Manual override for epic crafted items with crafting requirement
				if instance.Items[int(item.Id)].Bonding == dbc.BIND_ON_ACQUIRE && item.Quality == proto.ItemQuality_ItemQualityEpic && item.RequiredProfession == proto.Profession_ProfessionUnknown {
					item.RequiredProfession = crafted.Profession
				}
			}
		}

		// Forever's content phases are not known yet, so every item is phase 1.
		item.Phase = 1
	}
	addSpellIcons(db, craftedSpellIds, icons, iconsMap)

	database.LoadAndWriteEnchantDescriptions("assets/enchants/descriptions.json", db, instance)

	atlasDBProto := atlaslootDB.ToUIProto()
	db.MergeZones(atlasDBProto.Zones)
	db.MergeNpcs(atlasDBProto.Npcs)
	if err := database.LoadZoneAreas(db, helper); err != nil {
		log.Fatal(err)
	}
	applyRepSources(db, factionNames)
	db.WriteBinaryAndJson(fmt.Sprintf("%s/db.bin", dbDir), fmt.Sprintf("%s/db.json", dbDir))
}

// Drop sources from NPCs that are not in Forever although their zone is.
var dropNpcsNotInForever = map[int32]bool{
	36296: true, // Apothecary Hummel, Shadowfang Keep.
}

// pruneAtlasLootDrops removes the zones, NPCs and drop sources of AtlasLoot that are not in the game files or
// are in dropNpcsNotInForever, sets the remaining drops to Raid in a raid zone and Dungeon elsewhere, and
// removes the exact duplicates this leaves.
func pruneAtlasLootDrops(atlas *database.WowDatabase, raidZones, areaTableIDs map[int32]bool) {
	maps.DeleteFunc(atlas.Zones, func(id int32, _ *proto.UIZone) bool { return !areaTableIDs[id] })
	maps.DeleteFunc(atlas.Npcs, func(id int32, npc *proto.UINPC) bool {
		return !areaTableIDs[npc.ZoneId] || dropNpcsNotInForever[id]
	})
	for _, item := range atlas.Items {
		var drops []*proto.DropSource
		item.Sources = core.FilterSlice(item.Sources, func(source *proto.UIItemSource) bool {
			drop := source.GetDrop()
			if drop == nil {
				return true
			}
			if !areaTableIDs[drop.ZoneId] || dropNpcsNotInForever[drop.NpcId] {
				return false
			}
			drop.Difficulty = database.ParseDungeonDifficulty(raidZones[drop.ZoneId])
			if slices.ContainsFunc(drops, func(seen *proto.DropSource) bool { return googleProto.Equal(seen, drop) }) {
				return false
			}
			drops = append(drops, drop)
			return true
		})
	}
}

// pruneQuestieVendors removes the vendors with no zone or a zone that is not in the game files.
func pruneQuestieVendors(questie *database.WowDatabase, areaTableIDs map[int32]bool) {
	for _, item := range questie.Items {
		item.Sources = core.FilterSlice(item.Sources, func(source *proto.UIItemSource) bool {
			soldBy := source.GetSoldBy()
			return soldBy == nil || areaTableIDs[soldBy.ZoneId]
		})
	}
}

// appendQuestieSources adds the quest and vendor sources to the items the database has.
func appendQuestieSources(db, questie *database.WowDatabase) {
	for id, questieItem := range questie.Items {
		if item, ok := db.Items[id]; ok {
			item.Sources = append(slices.Clip(item.Sources), questieItem.Sources...)
		}
	}
}

func applyRepSources(db *database.WowDatabase, factionNames map[int32]string) {
	for _, item := range db.Items {
		sides := make(map[proto.Faction]bool)
		for _, source := range item.Sources {
			if rep := source.GetRep(); rep != nil {
				db.MergeFaction(&proto.UIFaction{Id: rep.RepFactionId, Name: factionNames[rep.RepFactionId]})
				sides[rep.FactionId] = true
			}
		}
		// A side's restriction holds only when every rep source names that side.
		if len(sides) == 1 && !sides[proto.Faction_Unknown] {
			for side := range sides {
				item.FactionRestriction = database.AtlasLootFactionRestrictions[side]
			}
		}
	}
}

func processItems(instance *dbc.DBC,
	iconsMap map[int]string,
	names map[int]string,
	dropSources map[int][]*proto.DropSource,
	craftingSources map[int][]*proto.CraftedSource,
	repSources map[int][]*proto.RepSource,
	db *database.WowDatabase) {
	sourceMap := make(map[string][]*proto.UIItemSource, len(instance.Items))
	parsedItems := make([]*proto.UIItem, 0, len(instance.Items))

	sortedItemKeys := slices.Sorted(maps.Keys(instance.Items))
	sortedItems := make([]dbc.Item, len(instance.Items))
	for i, k := range sortedItemKeys {
		sortedItems[i] = instance.Items[k]
	}

	for _, item := range sortedItems {
		if item.Flags2&0x10 != 0 && (item.StatAlloc[0] > 0 && item.StatAlloc[0] < 600) {
			continue
		}
		if _, allowed := database.ItemAllowList[int32(item.Id)]; item.Flags0.Has(dbc.DEPRECATED) && !allowed {
			continue
		}
		parsed := item.ToUIItem()
		if parsed.Icon == "" {
			parsed.Icon = strings.ToLower(database.GetIconName(iconsMap, item.FDID))
		}

		drops := dropSources[int(item.Id)]
		if drops != nil {
			sources := make([]*proto.UIItemSource, 0, len(drops))
			for _, drop := range drops {
				sources = append(sources, &proto.UIItemSource{
					Source: &proto.UIItemSource_Drop{Drop: drop},
				})
				db.MergeZone(&proto.UIZone{Id: drop.ZoneId, Name: names[int(drop.ZoneId)]})
				db.MergeNpc(&proto.UINPC{Id: drop.NpcId, Name: drop.OtherName, ZoneId: drop.ZoneId})
			}
			parsed.Sources = sources
			sourceMap[parsed.Name] = sources
		}

		crafted := craftingSources[int(item.Id)]
		if crafted != nil {
			sources := make([]*proto.UIItemSource, 0, len(crafted))
			for _, craft := range crafted {
				sources = append(sources, &proto.UIItemSource{
					Source: &proto.UIItemSource_Crafted{Crafted: craft},
				})
			}
			parsed.Sources = sources
			sourceMap[parsed.Name] = sources
		}

		rep := repSources[int(item.Id)]
		if rep != nil {
			sources := make([]*proto.UIItemSource, 0, len(rep))
			for _, repItem := range rep {
				sources = append(sources, &proto.UIItemSource{
					Source: &proto.UIItemSource_Rep{Rep: repItem},
				})
			}
			parsed.Sources = append(parsed.Sources, sources...)
			sourceMap[parsed.Name] = parsed.Sources
		}

		parsedItems = append(parsedItems, parsed)
	}

	for _, parsed := range parsedItems {
		if len(parsed.Sources) == 0 {
			if fallbacks, ok := sourceMap[parsed.Name]; ok {
				parsed.Sources = fallbacks
			}
		}
	}
	db.MergeItems(parsedItems)
}

// Filters out entities which shouldn't be included anywhere.
func ApplyGlobalFilters(db *database.WowDatabase) {
	db.Items = core.FilterMap(db.Items, func(_ int32, item *proto.UIItem) bool {
		if _, ok := database.ItemAllowList[item.Id]; ok {
			return true
		}
		if _, ok := database.ItemDenyList[item.Id]; ok {
			return false
		}
		if len(item.ScalingOptions) <= 0 {
			return false
		}
		if item.ScalingOptions[0].Ilvl > core.MaxIlvl {
			return false
		}
		for _, pattern := range database.DenyListNameRegexes {
			if pattern.MatchString(item.Name) {
				return false
			}
		}
		return true
	})

	// There is an 'unavailable' version of every naxx set, e.g. https://www.wowhead.com/mop-classic/item=43728/bonescythe-gauntlets
	heroesItems := core.FilterMap(db.Items, func(_ int32, item *proto.UIItem) bool {
		return strings.HasPrefix(item.Name, "Heroes' ")
	})
	db.Items = core.FilterMap(db.Items, func(_ int32, item *proto.UIItem) bool {
		nameToMatch := "Heroes' " + item.Name
		for _, heroItem := range heroesItems {
			if heroItem.Name == nameToMatch {
				return false
			}
		}
		return true
	})

	// There is an 'unavailable' version of many t8 set pieces, e.g. https://www.wowhead.com/mop-classic/item=46235/darkruned-gauntlets
	valorousItems := core.FilterMap(db.Items, func(_ int32, item *proto.UIItem) bool {
		return strings.HasPrefix(item.Name, "Valorous ")
	})
	db.Items = core.FilterMap(db.Items, func(_ int32, item *proto.UIItem) bool {
		nameToMatch := "Valorous " + item.Name
		for _, item := range valorousItems {
			if item.Name == nameToMatch {
				return false
			}
		}
		return true
	})

	// There is an 'unavailable' version of many t9 set pieces, e.g. https://www.wowhead.com/mop-classic/item=48842/thralls-hauberk
	triumphItems := core.FilterMap(db.Items, func(_ int32, item *proto.UIItem) bool {
		return strings.HasSuffix(item.Name, "of Triumph")
	})
	db.Items = core.FilterMap(db.Items, func(_ int32, item *proto.UIItem) bool {
		nameToMatch := item.Name + " of Triumph"
		for _, item := range triumphItems {
			if item.Name == nameToMatch {
				return false
			}
		}
		return true
	})

	database.DropLowerSameNameSets(db)

	db.ItemIcons = core.FilterMap(db.ItemIcons, func(_ int32, icon *proto.IconData) bool {
		return icon.Name != "" && icon.Icon != ""
	})
	db.SpellIcons = core.FilterMap(db.SpellIcons, func(_ int32, icon *proto.IconData) bool {
		return icon.Name != "" && icon.Icon != "" && icon.Id != 0
	})

	db.Enchants = core.FilterMap(db.Enchants, func(_ database.EnchantDBKey, enchant *proto.UIEnchant) bool {
		return database.KeepEnchant(enchant)
	})

	db.Consumables = core.FilterMap(db.Consumables, func(_ int32, consumable *proto.Consumable) bool {
		return database.KeepConsumable(consumable)
	})
}

// Filters out entities which shouldn't be included in the sim.
func ApplySimmableFilters(db *database.WowDatabase) {
	db.Items = core.FilterMap(db.Items, simmableItemFilter)
	db.Enchants = core.FilterMap(db.Enchants, simmableEnchantFilter)
}

func ApplyNonSimmableFilters(db *database.WowDatabase) {
	db.Items = core.FilterMap(db.Items, func(id int32, item *proto.UIItem) bool {
		return !simmableItemFilter(id, item)
	})
	db.Enchants = core.FilterMap(db.Enchants, func(id database.EnchantDBKey, enchant *proto.UIEnchant) bool {
		return !simmableEnchantFilter(id, enchant)
	})
	// db.json ships every consumable and its icon.
	for id := range db.Consumables {
		delete(db.ItemIcons, id)
	}
	db.Consumables = map[int32]*proto.Consumable{}
}
func simmableItemFilter(_ int32, item *proto.UIItem) bool {
	if _, ok := database.ItemAllowList[item.Id]; ok {
		return true
	}

	if item.Quality < proto.ItemQuality_ItemQualityUncommon {
		return false
	} else if item.Quality == proto.ItemQuality_ItemQualityArtifact {
		return false
	} else if item.Quality >= proto.ItemQuality_ItemQualityHeirloom {
		return false
	} else if item.Quality <= proto.ItemQuality_ItemQualityEpic {
		// TODO: Figure out the appropriate filter
		// if item.ScalingOptions[0].Ilvl < 60 {
		// 	return false
		// }
	}
	if item.ScalingOptions[0].Ilvl == 0 {
		fmt.Printf("Missing ilvl: %s\n", item.Name)
	}

	return true
}

var nonLiveEnchants = map[database.EnchantDBKey]struct{}{}

func simmableEnchantFilter(key database.EnchantDBKey, enchant *proto.UIEnchant) bool {
	if slices.Contains(database.EnchantAllowList, enchant.EffectId) {
		return true
	}
	if _, ok := database.EnchantDenyList[enchant.EffectId]; ok {
		return false
	}
	if _, ok := nonLiveEnchants[key]; ok {
		return false
	}
	// TODO: Refine this filter to better capture simmable enchants based on effect ID and item ID ranges.
	// return enchant.EffectId > 1000 && (enchant.ItemId == 0 || enchant.ItemId > 20000 || enchant.ItemId == 18283)
	return true
}

func loadTalentTrees(path string) []database.TalentTabConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to load talent json file: %s", err)
	}
	var trees []database.TalentTabConfig
	if err := json.Unmarshal(data, &trees); err != nil {
		log.Fatalf("failed to parse talent json %s: %s", path, err)
	}
	return trees
}

func getSpellIdsFromTalentJson(infile *string) []int32 {
	spellIds := make([]int32, 0)
	for _, tree := range loadTalentTrees(*infile) {
		for _, talent := range tree.Talents {
			spellIds = append(spellIds, int32(talent.SpellID))
		}
	}
	return spellIds
}

var talentTreesDir string

// rotationTalentsString is the talent string GetAllRotationSpellIds registers each spec
// with, purely to reach the rotation spells whose icons the database needs: every talent of
// the class at its maximum, built from the same tree JSON the UI reads. An unmodelled talent
// returns rather than panics, so a maxed build registers every ability a talent grants; an
// ability that is itself still a stub panics, which CreateTempAgent reports and skips.
func rotationTalentsString(class string) string {
	return database.MaxedTalentsString(loadTalentTrees(fmt.Sprintf("%s/%s.json", talentTreesDir, class)))
}

func GetAllTalentSpellIds(inputsDir *string) map[string][]int32 {
	talentsDir := talentTreesDir
	specFiles := []string{
		"druid.json",
		"hunter.json",
		"mage.json",
		"paladin.json",
		"priest.json",
		"rogue.json",
		"shaman.json",
		"warlock.json",
		"warrior.json",
	}

	ret_db := make(map[string][]int32, 0)

	for _, specFile := range specFiles {
		specPath := fmt.Sprintf("%s/%s", talentsDir, specFile)
		ret_db[specFile[:len(specFile)-5]] = getSpellIdsFromTalentJson(&specPath)
	}

	return ret_db

}

// Returns nil rather than dying when the spec cannot be built. Every class has abilities
// still stubbed as panic("To be implemented"), and building an agent runs its registrars,
// so a stubbed spec takes the whole database build down with it. The rotation spell ids
// this feeds are a UI convenience; a spec that cannot start simply contributes none.
func CreateTempAgent(r *proto.Raid) (agent core.Agent) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintf(os.Stderr, "[rotation] spec skipped: %v\n%s\n", err, debug.Stack())
			agent = nil
		}
	}()

	encounter := core.MakeSingleTargetEncounter(0.0)
	env, _, _ := core.NewEnvironment(r, encounter, false, false)
	return env.Raid.Parties[0].Players[0]
}

type RotContainer struct {
	Name string
	Raid *proto.Raid
}

func GetAllRotationSpellIds() map[string][]int32 {
	sim.RegisterAll()

	rotMapping := []RotContainer{
		// Druid
		{Name: "balanceDruid", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassDruid,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: rotationTalentsString("druid"),
		}, &proto.Player_BalanceDruid{BalanceDruid: &proto.BalanceDruid{Options: &proto.BalanceDruid_Options{ClassOptions: &proto.DruidOptions{}}}}), nil, nil, nil)},
		{Name: "feralCatDruid", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassDruid,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: rotationTalentsString("druid"),
		}, &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{Options: &proto.FeralCatDruid_Options{ClassOptions: &proto.DruidOptions{}}}}), nil, nil, nil)},
		{Name: "feralBearDruid", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassDruid,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: rotationTalentsString("druid"),
		}, &proto.Player_FeralBearDruid{FeralBearDruid: &proto.FeralBearDruid{Options: &proto.FeralBearDruid_Options{ClassOptions: &proto.DruidOptions{}}}}), nil, nil, nil)},
		{Name: "restorationDruid", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassDruid,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: rotationTalentsString("druid"),
		}, &proto.Player_RestorationDruid{RestorationDruid: &proto.RestorationDruid{Options: &proto.RestorationDruid_Options{ClassOptions: &proto.DruidOptions{}}}}), nil, nil, nil)},

		// Hunter
		{Name: "hunter", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassHunter,
			Race:          proto.Race_RaceTroll,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: rotationTalentsString("hunter"),
		}, &proto.Player_Hunter{Hunter: &proto.Hunter{Options: &proto.Hunter_Options{ClassOptions: &proto.HunterOptions{}}}}), nil, nil, nil)},

		// Mage
		{Name: "mage", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassMage,
			Race:          proto.Race_RaceTroll,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: rotationTalentsString("mage"),
		}, &proto.Player_Mage{Mage: &proto.Mage{Options: &proto.Mage_Options{ClassOptions: &proto.MageOptions{}}}}), nil, nil, nil)},

		// Paladin
		{Name: "holyPaladin", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassPaladin,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: rotationTalentsString("paladin"),
		}, &proto.Player_HolyPaladin{HolyPaladin: &proto.HolyPaladin{Options: &proto.HolyPaladin_Options{ClassOptions: &proto.PaladinOptions{}}}}), nil, nil, nil)},
		{Name: "protPaladin", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassPaladin,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: rotationTalentsString("paladin"),
		}, &proto.Player_ProtectionPaladin{ProtectionPaladin: &proto.ProtectionPaladin{Options: &proto.ProtectionPaladin_Options{ClassOptions: &proto.PaladinOptions{}}}}), nil, nil, nil)},
		{Name: "retPaladin", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassPaladin,
			Race:          proto.Race_RaceUndead,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: rotationTalentsString("paladin"),
		}, &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{Options: &proto.RetributionPaladin_Options{ClassOptions: &proto.PaladinOptions{}}}}), nil, nil, nil)},

		// Priest
		{Name: "priest", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassPriest,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: rotationTalentsString("priest"),
		}, &proto.Player_DpsPriest{DpsPriest: &proto.DpsPriest{Options: &proto.DpsPriest_Options{ClassOptions: &proto.PriestOptions{}}}}), nil, nil, nil)},
		{Name: "healerPriest", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassPriest,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: rotationTalentsString("priest"),
		}, &proto.Player_HealerPriest{HealerPriest: &proto.HealerPriest{Options: &proto.HealerPriest_Options{ClassOptions: &proto.PriestOptions{}}}}), nil, nil, nil)},

		// Rogue
		{Name: "rogue", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassRogue,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: rotationTalentsString("rogue"),
		}, &proto.Player_Rogue{Rogue: &proto.Rogue{Options: &proto.Rogue_Options{ClassOptions: &proto.RogueOptions{}}}}), nil, nil, nil)},

		// Shaman
		{Name: "elementalShaman", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassShaman,
			Race:          proto.Race_RaceTroll,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: rotationTalentsString("shaman"),
		}, &proto.Player_ElementalShaman{ElementalShaman: &proto.ElementalShaman{Options: &proto.ElementalShaman_Options{ClassOptions: &proto.ShamanOptions{}}}}), nil, nil, nil)},
		{Name: "enhancementShaman", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassShaman,
			Race:          proto.Race_RaceTroll,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: rotationTalentsString("shaman"),
		}, &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{Options: &proto.EnhancementShaman_Options{ClassOptions: &proto.ShamanOptions{}}}}), nil, nil, nil)},
		{Name: "restorationShaman", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassShaman,
			Race:          proto.Race_RaceTroll,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: rotationTalentsString("shaman"),
		}, &proto.Player_RestorationShaman{RestorationShaman: &proto.RestorationShaman{Options: &proto.RestorationShaman_Options{ClassOptions: &proto.ShamanOptions{}}}}), nil, nil, nil)},

		// Warlock
		{Name: "warlock", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassWarlock,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: rotationTalentsString("warlock"),
			Profession1:   proto.Profession_Herbalism,
		}, &proto.Player_Warlock{Warlock: &proto.Warlock{Options: &proto.Warlock_Options{ClassOptions: &proto.WarlockOptions{}}}}), nil, nil, nil)},

		// Warrior
		{Name: "DpsWarrior", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassWarrior,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: rotationTalentsString("warrior"),
		}, &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{Options: &proto.DpsWarrior_Options{ClassOptions: &proto.WarriorOptions{}}}}), nil, nil, nil)},
		{Name: "protectionWarrior", Raid: core.SinglePlayerRaidProto(core.WithSpec(&proto.Player{
			Class:         proto.Class_ClassWarrior,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: rotationTalentsString("warrior"),
		}, &proto.Player_ProtectionWarrior{ProtectionWarrior: &proto.ProtectionWarrior{Options: &proto.ProtectionWarrior_Options{ClassOptions: &proto.WarriorOptions{}}}}), nil, nil, nil)},
	}

	ret_db := make(map[string][]int32, 0)

	stubbed := 0
	for _, r := range rotMapping {
		agent := CreateTempAgent(r.Raid)
		if agent == nil {
			stubbed++
			continue
		}
		f := agent.GetCharacter()

		spells := make([]int32, 0, len(f.Spellbook))

		for _, s := range f.Spellbook {
			if s.SpellID != 0 {
				spells = append(spells, s.SpellID)
			}
		}

		for _, s := range f.GetAuras() {
			if s.ActionID.SpellID != 0 {
				spells = append(spells, s.ActionID.SpellID)
			}
		}

		ret_db[r.Name] = spells
	}
	if stubbed > 0 {
		fmt.Printf("Rotation spell ids: %d of %d specs skipped, their abilities are still stubbed\n", stubbed, len(rotMapping))
	}

	return ret_db
}

func addSpellIcons(db *database.WowDatabase, spellIds []int32, icons map[int]database.SpellIcon, iconsMap map[int]string) {
	for _, spellId := range spellIds {
		iconEntry := icons[int(spellId)]
		if iconEntry.Name == "" {
			continue
		}
		db.SpellIcons[spellId] = &proto.IconData{
			Id:      int32(iconEntry.SpellID),
			Name:    iconEntry.Name,
			Icon:    strings.ToLower(database.GetIconName(iconsMap, iconEntry.FDID)),
			HasBuff: iconEntry.HasBuff,
			Rank:    int32(iconEntry.Rank),
		}
	}
}

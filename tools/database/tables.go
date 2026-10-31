package database

import (
	"cmp"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/dbc"
	"github.com/wowsims/forever/tools/tooltip"
)

// Loading tables
// Below is the definition and loading of tables
//

// Raw Item Data

func ScanRawItemData(rows *sql.Rows) (dbc.Item, error) {
	var raw dbc.Item
	var randomSuffixOptions sql.NullString
	var bonusAmountCalculated string
	var bonusStatString string
	var statValue string
	var maxCount int
	err := rows.Scan(
		&raw.Id,
		&raw.Name,
		&raw.InventoryType,
		&raw.ItemDelay,
		&raw.OverallQuality,
		&raw.DmgVariance,
		&raw.ItemLevel,
		&raw.RequiredLevel,
		&statValue,
		&bonusStatString,
		&raw.ArmorValue,
		&raw.FireResistance,
		&raw.NatureResistance,
		&raw.FrostResistance,
		&raw.ShadowResistance,
		&raw.ArcaneResistance,
		&raw.Flags0,
		&raw.Flags1,
		&raw.Flags2,
		&raw.Flags4,
		&raw.FDID,
		&raw.ItemSetName,
		&raw.ItemSetId,
		&raw.ClassMask,
		&raw.RaceMask,
		&raw.QualityModifier,
		&raw.DamageType,
		&randomSuffixOptions,
		&bonusAmountCalculated,
		&raw.ItemClass,
		&raw.ItemSubClass,
		&raw.NameDescription,
		&raw.LimitCategory,
		&raw.Bonding,
		&maxCount,
	)
	if err != nil {
		panic(err)
	}
	var parseErr error
	raw.RandomSuffixOptions, parseErr = ParseRandomSuffixOptions(randomSuffixOptions)
	if parseErr != nil {
		return raw, fmt.Errorf("failed to parse RandomSuffixOptions: %w", parseErr)
	}
	slices.Sort(raw.RandomSuffixOptions)

	raw.StatAlloc, err = parseFloatArrayField(statValue, 10)
	if err != nil {
		return raw, fmt.Errorf("failed to parse StatAlloc: %w", err)
	}
	raw.BonusAmountCalculated, err = parseFloatArrayField(bonusAmountCalculated, 10)
	if err != nil {
		return raw, fmt.Errorf("failed to parse BonusAmountCalculated: %w", err)
	}
	raw.BonusStat, err = parseIntArrayField(bonusStatString, 10)
	if err != nil {
		return raw, fmt.Errorf("failed to parse BonusStat: %w", err)
	}
	// MaxCount identifies that an item is unique
	// this is not the same as unique-equipped but
	// we can use this to identify items that are unique in the database
	if maxCount > 0 {
		raw.Flags0 |= dbc.UNIQUE_EQUIPPABLE
	}
	return raw, err
}

// The gear the sim ships: weapons, armour and relics of a quality the game hands out, at or below
// the level the sim plays. RequiredLevel 0 means "no level requirement", so <= keeps those. Written
// here rather than at the call site because the spell data store selects the item procs it carries
// through the same predicate.
func SimItemFilter(maxLevel int32) string {
	return fmt.Sprintf("s.OverallQualityId != 7 AND s.OverallQualityId != 0 AND (i.ClassID = 2 OR i.ClassID = 4 OR (i.ClassID = 7 AND i.InventoryType = 12)) AND s.Display_lang != '' AND s.RequiredLevel <= %d AND (s.ID != 34219 AND s.Display_lang NOT LIKE '%%Test%%' AND s.Display_lang NOT LIKE 'QA%%' AND s.Display_lang != 'unused')", maxLevel)
}

func LoadAndWriteRawItems(dbHelper *DBHelper, filter string, inputsDir string) ([]dbc.Item, error) {
	baseQuery := `
		SELECT
			i.ID,
			s.Display_lang AS Name,
			i.InventoryType,
			s.ItemDelay,
			s.OverallQualityID,
			s.DmgVariance,
			s.ItemLevel,
			s.RequiredLevel,
			s.StatPercentEditor as StatValue,
			s.StatModifier_bonusStat as bonusStat,
			0 as ArmorValue, -- Item.Resistances_* is absent from the game files: GetArmorValue computes
			0 as FireResistance, -- armor from the item level tables, and resistances arrive as stat-array entries
			0 as NatureResistance,
			0 as FrostResistance,
			0 as ShadowResistance,
			0 as ArcaneResistance,
			s.Flags_0 as Flags_0,
			s.Flags_1 as Flags_1,
			s.Flags_2 as Flags_2,
			s.Flags_4 as Flags_4,
			i.IconFileDataId as FDID,
			COALESCE(itemset.Name_lang, '') as ItemSetName,
			COALESCE(itemset.ID, 0) as ItemSetID,
			s.AllowableClass as ClassMask,
			(s.AllowableRace_1 << 32) | (s.AllowableRace_0 & 0xFFFFFFFF) as RaceMask,
			s.QualityModifier,
			s.DamageType,
			(SELECT group_concat(DISTINCT n.ChildItemBonusListID) ` + randomSuffixNodes + ` AND x.ItemID = i.ID) AS RandomSuffixOptions,
			 '[0,0,0,0,0,0,0,0,0,0]', -- StatModifier_bonusAmount: dropped; derived from StatAlloc
			 i.ClassID,
			 i.SubClassID,
			 COALESCE(ind.Description_lang, ''),
			s.LimitCategory,
			s.Bonding,
			COALESCE(s.MaxCount, 0) as MaxCount
		FROM Item i
		JOIN ItemSparse s ON i.ID = s.ID
		JOIN ItemClass ic ON i.ClassID = ic.ClassID
		JOIN ItemSubClass isc ON i.ClassID = isc.ClassID AND i.SubClassID = isc.SubClassID
		JOIN RandPropPoints rpp ON s.ItemLevel = rpp.ID
		LEFT JOIN ItemArmorShield ias ON s.ItemLevel = ias.ItemLevel
		LEFT JOIN ItemSet itemset ON s.ItemSet = itemset.ID
		LEFT JOIN ItemArmorQuality iaq ON s.ItemLevel = iaq.ID
		LEFT JOIN ItemNameDescription as ind ON s.ItemNameDescriptionID = ind.ID
		JOIN ItemArmorTotal at ON s.ItemLevel = at.ItemLevel
		`

	if strings.TrimSpace(filter) != "" {
		// Items in the allowlist always bypass the filter.
		allowListIDs := make([]string, 0, len(ItemAllowList))
		for id := range ItemAllowList {
			allowListIDs = append(allowListIDs, strconv.Itoa(int(id)))
		}
		slices.Sort(allowListIDs)
		baseQuery += " WHERE (" + filter + ")"
		if len(allowListIDs) > 0 {
			baseQuery += " OR i.ID IN (" + strings.Join(allowListIDs, ",") + ")"
		}
	}

	items, err := LoadRows(dbHelper.db, baseQuery, ScanRawItemData)
	fmt.Println("Loaded Items:", len(items))
	if err != nil {
		fmt.Println("Error loading items:", err.Error())
		return nil, err
	}
	if err := applyCreatorClassMasks(dbHelper, items, func(item *dbc.Item) (int, *int) { return item.Id, &item.ClassMask }); err != nil {
		return nil, err
	}
	json, _ := json.Marshal(items)
	if err := dbc.WriteGzipFile(fmt.Sprintf("%s/dbc/items.json", inputsDir), json); err != nil {
		panic(fmt.Sprintf("Error loading DBC data %v", err))
	}

	return items, nil
}

// The class mask of each soulbound item the game files leave open to every class while only class
// spells create it, such as the warlock's Firestone and the mage's mana gems: the creating spells'
// SkillLineAbility class bits, ORed. A soulbound item never leaves the character who made it. A
// tradable one (Healthstone, Conjured Water, Blinding Powder) stays open to every class.
func LoadCreatorClassMasks(dbHelper *DBHelper) (map[int]int, error) {
	type creator struct{ itemID, classMask int }
	creators, err := LoadRows(dbHelper.db, fmt.Sprintf(`
		SELECT DISTINCT e.EffectItemType, sla.ClassMask
		FROM SpellEffect e
		JOIN SkillLineAbility sla ON sla.Spell = e.SpellID AND sla.ClassMask != 0
		JOIN ItemSparse s ON s.ID = e.EffectItemType
		WHERE e.Effect = %d AND s.Bonding = %d AND s.AllowableClass IN (-1, 0)`,
		dbcenums.E_CREATE_ITEM, dbc.BIND_ON_ACQUIRE),
		func(rows *sql.Rows) (creator, error) {
			var c creator
			return c, rows.Scan(&c.itemID, &c.classMask)
		})
	if err != nil {
		return nil, fmt.Errorf("error loading creator class masks: %w", err)
	}
	masks := make(map[int]int, len(creators))
	for _, c := range creators {
		masks[c.itemID] |= c.classMask
	}
	return masks, nil
}

// Each row's class mask becomes the one LoadCreatorClassMasks derives for its item, where it derives one.
func applyCreatorClassMasks[T any](dbHelper *DBHelper, rows []T, classMaskOf func(*T) (int, *int)) error {
	masks, err := LoadCreatorClassMasks(dbHelper)
	if err != nil {
		return err
	}
	for i := range rows {
		if id, classMask := classMaskOf(&rows[i]); masks[id] != 0 {
			*classMask = masks[id]
		}
	}
	return nil
}

// ItemDamage tables
func ScanItemDamageTable(rows *sql.Rows) (dbc.ItemDamageTable, error) {
	var raw dbc.ItemDamageTable
	var qualityString string
	err := rows.Scan(&raw.ItemLevel, &qualityString)
	if err != nil {
		return raw, fmt.Errorf("scanning item damage table: %w", err)
	}

	raw.Quality, err = parseFloatArrayField(qualityString, 7)
	if err != nil {
		return raw, fmt.Errorf("parsing quality string '%s': %w", qualityString, err)
	}

	return raw, nil
}

var ItemDamageByTableAndItemLevel = make(map[string]map[int]dbc.ItemDamageTable)
var itemDamageTableNames = []string{
	"ItemDamageAmmo",
	"ItemDamageOneHand",
	"ItemDamageOneHandCaster",
	"ItemDamageRanged",
	"ItemDamageTwoHand",
	"ItemDamageTwoHandCaster",
	"ItemDamageThrown",
	"ItemDamageWand",
}

func LoadAndWriteItemDamageTables(dbHelper *DBHelper, inputsDir string) (map[string]map[int]dbc.ItemDamageTable, error) {
	for _, tableName := range itemDamageTableNames {
		query := fmt.Sprintf("SELECT ItemLevel, Quality FROM %s", tableName)
		items, err := LoadRows(dbHelper.db, query, ScanItemDamageTable)
		if err != nil {
			return nil, fmt.Errorf("error loading items for table %s: %w", tableName, err)
		}

		// Cache the slice of ItemDamageTable into a map keyed by ItemLevel.
		ItemDamageByTableAndItemLevel[tableName] = CacheBy(items, func(table dbc.ItemDamageTable) int {
			return table.ItemLevel
		})
	}
	json, _ := json.Marshal(ItemDamageByTableAndItemLevel)
	if err := dbc.WriteGzipFile(fmt.Sprintf("%s/dbc/item_damage_tables.json", inputsDir), json); err != nil {
		panic(fmt.Sprintf("Error loading DBC data %v", err))
	}
	return ItemDamageByTableAndItemLevel, nil
}

func LoadAndWriteItemArmorQuality(dbHelper *DBHelper, inputsDir string) (map[int]dbc.ItemArmorQuality, error) {
	query := "SELECT ID, Qualitymod FROM ItemArmorQuality"
	result, err := LoadRows(dbHelper.db, query, ScanItemArmorQualityTable)

	cache := CacheBy(result, func(table dbc.ItemArmorQuality) int {
		return table.ItemLevel
	})
	json, _ := json.Marshal(cache)
	if err := dbc.WriteGzipFile(fmt.Sprintf("%s/dbc/item_armor_quality.json", inputsDir), json); err != nil {
		panic(fmt.Sprintf("Error loading DBC data %v", err))
	}
	return cache, err
}

func ScanItemArmorQualityTable(rows *sql.Rows) (dbc.ItemArmorQuality, error) {
	var raw dbc.ItemArmorQuality
	var qualityString string
	err := rows.Scan(&raw.ItemLevel, &qualityString)
	if err != nil {
		return raw, fmt.Errorf("scanning item armor quality: %w", err)
	}

	raw.Quality, err = parseFloatArrayField(qualityString, 7)
	if err != nil {
		return raw, fmt.Errorf("parsing quality string '%s': %w", qualityString, err)
	}

	return raw, nil
}

func LoadAndWriteItemArmorShield(dbHelper *DBHelper, inputsDir string) (map[int]dbc.ItemArmorShield, error) {
	query := "SELECT ItemLevel, Quality FROM ItemArmorShield"
	result, err := LoadRows(dbHelper.db, query, ScanItemArmorShieldTable)

	cache := CacheBy(result, func(table dbc.ItemArmorShield) int {
		return table.ItemLevel
	})
	json, _ := json.Marshal(cache)
	if err := dbc.WriteGzipFile(fmt.Sprintf("%s/dbc/item_armor_shield.json", inputsDir), json); err != nil {
		panic(fmt.Sprintf("Error loading DBC data %v", err))
	}
	return cache, err
}

func ScanItemArmorShieldTable(rows *sql.Rows) (dbc.ItemArmorShield, error) {
	var raw dbc.ItemArmorShield
	var qualityString string
	err := rows.Scan(&raw.ItemLevel, &qualityString)
	if err != nil {
		return raw, fmt.Errorf("scanning item armor shield: %w", err)
	}

	raw.Quality, err = parseFloatArrayField(qualityString, 7)
	if err != nil {
		return raw, fmt.Errorf("parsing quality string '%s': %w", qualityString, err)
	}

	return raw, nil
}
func LoadAndWriteItemArmorTotal(dbHelper *DBHelper, inputsDir string) (map[int]dbc.ItemArmorTotal, error) {
	query := "SELECT ItemLevel, Cloth, Leather, Mail, Plate FROM ItemArmorTotal"
	result, err := LoadRows(dbHelper.db, query, ScanItemArmorTotalTable)
	cached := CacheBy(result, func(table dbc.ItemArmorTotal) int {
		return table.ItemLevel
	})
	json, _ := json.Marshal(cached)
	if err := dbc.WriteGzipFile(fmt.Sprintf("%s/dbc/item_armor_total.json", inputsDir), json); err != nil {
		panic(fmt.Sprintf("Error loading DBC data %v", err))
	}

	return cached, err
}

func ScanItemArmorTotalTable(rows *sql.Rows) (dbc.ItemArmorTotal, error) {
	var raw dbc.ItemArmorTotal
	var qualityString string
	err := rows.Scan(&raw.ItemLevel, &raw.Cloth, &raw.Leather, &raw.Mail, &raw.Plate)
	if err != nil {
		fmt.Println(err.Error(), 3, qualityString)
		return raw, fmt.Errorf("error loading ScanItemArmorTotalTable")
	}
	return raw, err
}

func LoadAndWriteArmorLocation(dbHelper *DBHelper, inputsDir string) (map[int]dbc.ArmorLocation, error) {
	query := "SELECT ID, Clothmodifier, Leathermodifier, Chainmodifier, Platemodifier, Modifier FROM ArmorLocation"
	result, err := LoadRows(dbHelper.db, query, ScanArmorLocation)
	cache := CacheBy(result, func(table dbc.ArmorLocation) int {
		return table.Id
	})
	json, _ := json.Marshal(cache)
	if err := dbc.WriteGzipFile(fmt.Sprintf("%s/dbc/armor_location.json", inputsDir), json); err != nil {
		panic(fmt.Sprintf("Error loading DBC data %v", err))
	}
	return cache, err
}

func ScanArmorLocation(rows *sql.Rows) (dbc.ArmorLocation, error) {
	var raw dbc.ArmorLocation
	raw.Modifier = [5]float64{}
	err := rows.Scan(&raw.Id, &raw.Modifier[0], &raw.Modifier[1], &raw.Modifier[2], &raw.Modifier[3], &raw.Modifier[4])
	return raw, err
}

func ScanEnchantsTable(rows *sql.Rows) (dbc.Enchant, error) {
	var raw dbc.Enchant
	var effectsString string
	var effectPointsString string
	var effectArgsString string
	var spellItemEnchantmentFlags dbc.SpellItemEnchantmentFlags
	var spellItemLevel int
	var itemClassMasks string
	err := rows.Scan(
		&raw.EffectId,
		&raw.Name,
		&raw.SpellId,
		&raw.ItemId,
		&raw.ProfessionId,
		&effectsString,
		&effectPointsString,
		&effectArgsString,
		&raw.IsWeaponEnchant,
		&raw.InventoryType,
		&raw.SubClassMask,
		&raw.ClassMask,
		&itemClassMasks,
		&raw.FDID,
		&raw.Quality,
		&raw.RequiredProfession,
		&raw.EffectName,
		&spellItemEnchantmentFlags,
		&raw.IsLive,
		&spellItemLevel,
	)
	if err != nil {
		return raw, fmt.Errorf("scanning enchant data for effect ID %d: %w", raw.EffectId, err)
	}
	raw.Lvl = enchantLvl(spellItemLevel)

	// A class may use the enchant when it may use any item that applies it; AllowableClass -1 is every class.
	if itemClassMasks != "" {
		for _, field := range strings.Split(itemClassMasks, ",") {
			mask, err := strconv.Atoi(field)
			if err != nil {
				return raw, fmt.Errorf("parsing item class masks for enchant %d (%s): %w", raw.EffectId, itemClassMasks, err)
			}
			if mask <= 0 {
				mask = -1
			}
			raw.ClassMask |= mask
		}
	}

	raw.Effects, err = parseIntArrayField(effectsString, 3)
	if err != nil {
		return raw, fmt.Errorf("parsing effects for enchant %d (%s): %w", raw.EffectId, effectsString, err)
	}

	raw.EffectPoints, err = parseIntArrayField(effectPointsString, 3)
	if err != nil {
		return raw, fmt.Errorf("parsing effect points for enchant %d (%s): %w", raw.EffectId, effectPointsString, err)
	}

	raw.EffectArgs, err = parseIntArrayField(effectArgsString, 3)
	if err != nil {
		return raw, fmt.Errorf("parsing effect args for enchant %d (%s): %w", raw.EffectId, effectArgsString, err)
	}

	// Almost all enchants have Enchanting as requirement,
	// however Enchanting profession requirement is defined by the SOULBOUND flag
	if raw.RequiredProfession == 333 && !spellItemEnchantmentFlags.Has(dbc.SOULBOUND) {
		raw.ProfessionId = 0
		raw.RequiredProfession = 0
	}

	return raw, nil
}

// The highest item level of an item outside the Season of Discovery range that applies or teaches the
// enchant's spell; 0 where every such item is level 1, as the Forever enchant scrolls are.
func enchantLvl(spellItemLevel int) int {
	if spellItemLevel > 1 {
		return spellItemLevel
	}
	return 0
}

// Season of Discovery items occupy this id range in the game files; an enchant only they carry is left
// out, and none of them gives an enchant its level.
const (
	sodItemIDMin = 202251
	sodItemIDMax = 244460
)

// The SQL condition that the item id in column is in the Season of Discovery range.
func sodItemID(column string) string {
	return column + ` BETWEEN ` + strconv.Itoa(sodItemIDMin) + ` AND ` + strconv.Itoa(sodItemIDMax)
}

func LoadAndWriteRawEnchants(dbHelper *DBHelper, inputsDir string) ([]dbc.Enchant, error) {
	query := `
		SELECT DISTINCT
			sie.ID as effectId,
			CASE
			WHEN s.NameSubtext_lang IS NOT NULL
				AND TRIM(s.NameSubtext_lang) <> ''
			THEN
				(CASE
				WHEN sn.Name_lang LIKE '%+%' THEN COALESCE(isp.Display_lang, sn.Name_lang)
				ELSE sn.Name_lang
				END)
				|| ' (' || s.NameSubtext_lang || ')'
			WHEN sn.Name_lang LIKE '%+%' THEN COALESCE(isp.Display_lang, sn.Name_lang)
			ELSE sn.Name_lang
			END AS name,
			CASE
				WHEN sie.Effect_0 IN (1, 3) THEN sie.EffectArg_0
				WHEN sie.Effect_1 IN (1, 3) THEN sie.EffectArg_1
				WHEN sie.Effect_2 IN (1, 3) THEN sie.EffectArg_2
				ELSE se.SpellID
			END AS spellId,
			-- The lowest item the game files ship (it has an ItemSparse row), else the lowest item, an item in
			-- the Season of Discovery range last. This is the query's only aggregate, so the bare columns (icon
			-- and quality among them) come from its row.
			COALESCE(MIN(
				CASE WHEN ` + sodItemID("ixie.ItemID") + ` THEN 2000000000 ELSE 0 END
				+ CASE WHEN isp.ID IS NULL THEN 1000000000 ELSE 0 END
				+ ixie.ItemID
			) % 1000000000, 0) as ItemId,
			sie.RequiredSkillID as professionId,
			sie.Effect as Effect,
			sie.EffectPointsMin as EffectPoints,
			sie.EffectArg as EffectArgs,
			CASE
				WHEN sei.EquippedItemClass = 4 THEN false
				ELSE true
			END AS isWeaponEnchant,
			COALESCE(sei.EquippedItemInvTypes, 0) as InvTypes,
			COALESCE(sei.EquippedItemSubclass, 0),
			COALESCE(sla.ClassMask, 0),
			-- The AllowableClass of every item that applies an enchant no skill line teaches.
			CASE WHEN sla.ID IS NULL THEN COALESCE((
				SELECT group_concat(DISTINCT lisp.AllowableClass) FROM ItemEffect lie
					JOIN ItemXItemEffect lixie ON lixie.ItemEffectID = lie.ID
					JOIN ItemSparse lisp ON lisp.ID = lixie.ItemID
				WHERE lie.SpellID = se.SpellID AND NOT ` + sodItemID("lixie.ItemID") + `
			), '') ELSE '' END AS ItemClassMasks,
			COALESCE(it.IconFileDataID, 0),
			COALESCE(isp.OverallQualityID, 1),
			COALESCE(sla.SkillLine, 0) as RequiredProfession,
			COALESCE(sie.Name_lang, ""),
			COALESCE(sie.Flags, 0) AS SpellItemEnchantmentFlags,
			-- A grant is live when a profession teaches it or an item the game files ship uses it, a
			-- placeholder ([PH]) item aside.
			EXISTS (SELECT 1 FROM SkillLineAbility WHERE Spell = se.SpellID)
				OR EXISTS (
					SELECT 1 FROM ItemEffect lie
						JOIN ItemXItemEffect lixie ON lixie.ItemEffectID = lie.ID
						JOIN ItemSparse lisp ON lisp.ID = lixie.ItemID
					WHERE lie.SpellID = se.SpellID AND ` + notPlaceholderItemName("lisp.Display_lang") + `
				) AS IsLive,
			COALESCE((
				SELECT MAX(lisp.ItemLevel) FROM ItemEffect lie
					JOIN ItemXItemEffect lixie ON lixie.ItemEffectID = lie.ID
					JOIN ItemSparse lisp ON lisp.ID = lixie.ItemID
				WHERE lie.SpellID = se.SpellID AND NOT ` + sodItemID("lixie.ItemID") + `
			), 0) AS SpellItemLevel
		FROM SpellEffect se
			JOIN Spell s ON se.SpellID = s.ID
			JOIN SpellName sn ON se.SpellID = sn.ID
			JOIN SpellItemEnchantment sie ON se.EffectMiscValue_0 = sie.ID
			LEFT JOIN ItemEffect ie ON se.SpellID = ie.SpellID
			LEFT JOIN ItemXItemEffect ixie ON ixie.ItemEffectID = ie.ID
			LEFT JOIN SpellEquippedItems sei ON se.SpellId = sei.SpellID
			LEFT JOIN SkillLineAbility sla ON se.SpellID = sla.Spell
			LEFT JOIN Item it ON ixie.ItemID = it.ID
			LEFT JOIN ItemSparse isp ON ixie.ItemID = isp.ID
			WHERE se.Effect = 53
				AND (
					(
						sie.RequiredSkillID > 0
						OR sla.ID               IS NOT NULL
						OR sie.RequiredSkillRank IS NOT NULL
					)
					OR
					sie.RequiredSkillID = 0
					OR
					sie.RequiredSkillID = 773
				)
		GROUP BY sie.ID, name `
	items, err := LoadRows(dbHelper.db, query, ScanEnchantsTable)
	if err != nil {
		return nil, fmt.Errorf("error loading items for EnchantTables: %w", err)
	}
	json, _ := json.Marshal(items)
	if err := dbc.WriteGzipFile(fmt.Sprintf("%s/dbc/enchants.json", inputsDir), json); err != nil {
		panic(fmt.Sprintf("Error loading DBC data %v", err))
	}
	return items, nil
}

// RandPropPoints
// TBC ANNI: This table exists and contains data, so removing it seems bad, BUT item stats do not match MoP+.
// Stats come from two places - ItemSparse inside of StatModifier, and ItemEffect as Static (effectDuration = -1) and Proc/On-Use (effectDuration > 0)
type RandPropAllocationRow struct {
	Ilvl       int32
	Allocation dbc.RandomPropAllocation
}

func ScanRandPropAllocationRow(rows *sql.Rows) (RandPropAllocationRow, error) {
	var row RandPropAllocationRow
	var damageReplaceStat int
	err := rows.Scan(
		&row.Ilvl,
		&damageReplaceStat,
		&row.Allocation.Epic0,
		&row.Allocation.Epic1,
		&row.Allocation.Epic2,
		&row.Allocation.Epic3,
		&row.Allocation.Epic4,
		&row.Allocation.Superior0,
		&row.Allocation.Superior1,
		&row.Allocation.Superior2,
		&row.Allocation.Superior3,
		&row.Allocation.Superior4,
		&row.Allocation.Good0,
		&row.Allocation.Good1,
		&row.Allocation.Good2,
		&row.Allocation.Good3,
		&row.Allocation.Good4,
	)
	return row, err
}

func LoadAndWriteRandomPropAllocations(dbHelper *DBHelper, inputsDir string) (map[int32]RandPropAllocationRow, error) {
	query := `SELECT ID, DamageReplaceStat, Epic_0, Epic_1, Epic_2, Epic_3, Epic_4, Superior_0, Superior_1, Superior_2, Superior_3, Superior_4, Good_0, Good_1, Good_2 ,Good_3, Good_4 FROM RandPropPoints`
	rowsData, err := LoadRows(dbHelper.db, query, ScanRandPropAllocationRow)
	if err != nil {
		return nil, fmt.Errorf("error loading random property allocations: %w", err)
	}

	processed := make(map[int32]RandPropAllocationRow)
	for _, r := range rowsData {
		processed[r.Ilvl] = r
	}

	randProps := make(dbc.RandomPropAllocationsByIlvl)
	for _, r := range processed {
		randProps[int(r.Ilvl)] = dbc.RandomPropAllocationMap{
			proto.ItemQuality_ItemQualityEpic:     [5]int32{r.Allocation.Epic0, r.Allocation.Epic1, r.Allocation.Epic2, r.Allocation.Epic3, r.Allocation.Epic4},
			proto.ItemQuality_ItemQualityRare:     [5]int32{r.Allocation.Superior0, r.Allocation.Superior1, r.Allocation.Superior2, r.Allocation.Superior3, r.Allocation.Superior4},
			proto.ItemQuality_ItemQualityUncommon: [5]int32{r.Allocation.Good0, r.Allocation.Good1, r.Allocation.Good2, r.Allocation.Good3, r.Allocation.Good4},
		}
	}
	json, _ := json.Marshal(randProps)
	if err := dbc.WriteGzipFile(fmt.Sprintf("%s/dbc/rand_prop_points.json", inputsDir), json); err != nil {
		log.Fatalf("Error writing file: %v", err)
	}
	return processed, nil
}

func ScanSpellEffect(rows *sql.Rows) (dbc.SpellEffect, error) {
	var raw dbc.SpellEffect
	raw.EffectMinRange = []float64{0, 0}
	raw.EffectMaxRange = []float64{0, 0}

	// Temporary strings to hold the concatenated JSON for grouped fields.
	var miscValuesStr, spellClassMasksStr, implicitTargetsStr string
	err := rows.Scan(
		&raw.ID,
		&raw.DifficultyID,
		&raw.EffectIndex,
		&raw.EffectType,
		&raw.EffectAmplitude,
		&raw.EffectAttributes,
		&raw.EffectAura,
		&raw.EffectAuraPeriod,
		&raw.EffectBasePoints,
		&raw.EffectBonusCoefficient,
		&raw.EffectChainAmplitude,
		&raw.EffectChainTargets,
		&raw.EffectDieSides,
		&raw.EffectItemType,
		&raw.EffectMechanic,
		&raw.EffectPointsPerResource,
		&raw.EffectPosFacing,
		&raw.EffectRealPointsPerLevel,
		&raw.EffectTriggerSpell,
		&raw.BonusCoefficientFromAP,
		&raw.PvpMultiplier,
		&raw.Coefficient,
		&raw.Variance,
		&raw.ResourceCoefficient,
		&raw.GroupSizeBasePointsCoefficient,
		&miscValuesStr,
		&spellClassMasksStr,
		&implicitTargetsStr,
		&raw.SpellID,
		&raw.ScalingType,
		&raw.EffectMinRange[0],
		&raw.EffectMaxRange[0],
		&raw.EffectMinRange[1],
		&raw.EffectMaxRange[1],
	)
	if err != nil {
		return raw, err
	}

	raw.EffectMiscValues, err = parseIntArrayField(miscValuesStr, 2)
	if err != nil {
		return raw, fmt.Errorf("error parsing EffectMiscValues: %w", err)
	}
	raw.EffectSpellClassMasks, err = parseIntArrayField(spellClassMasksStr, 4)
	if err != nil {
		return raw, fmt.Errorf("error parsing EffectSpellClassMasks: %w", err)
	}
	raw.ImplicitTargets, err = parseIntArrayField(implicitTargetsStr, 2)
	if err != nil {
		return raw, fmt.Errorf("error parsing ImplicitTargets: %w", err)
	}

	return raw, nil
}

func LoadAndWriteRawSpellEffects(dbHelper *DBHelper, inputsDir string) (map[int]map[int]dbc.SpellEffect, error) {
	query := `
	SELECT
		se.ID,
		se.DifficultyID,
		se.EffectIndex,
		se.Effect,
		se.EffectAmplitude,
		se.EffectAttributes,
		se.EffectAura,
		se.EffectAuraPeriod,
		se.EffectBasePointsF,
		se.EffectBonusCoefficient,
		se.EffectChainAmplitude,
		se.EffectChainTargets,
		0, -- EffectDieSides: dropped from the game files; Variance carries the spread
		se.EffectItemType,
		se.EffectMechanic,
		se.EffectPointsPerResource,
		se.EffectPos_facing,
		se.EffectRealPointsPerLevel,
		se.EffectTriggerSpell,
		se.BonusCoefficientFromAP,
		se.PvpMultiplier,
		se.Coefficient,
		se.Variance,
		se.ResourceCoefficient,
		se.GroupSizeBasePointsCoefficient,
		se.EffectMiscValue,
		se.EffectSpellClassMask,
		se.ImplicitTarget,
		se.SpellID,
		0, -- SpellScaling.Class: dropped from the game files
		COALESCE(sr1.RadiusMin, 0),
		COALESCE(sr1.RadiusMax, 0),
		COALESCE(sr2.RadiusMin, 0),
		COALESCE(sr2.RadiusMax, 0)
	FROM SpellEffect se
	LEFT JOIN SpellScaling ss ON se.SpellID = ss.SpellID
	LEFT JOIN SpellRadius sr1 ON sr1.ID = se.EffectRadiusIndex_0
	LEFT JOIN SpellRadius sr2 ON sr2.ID = se.EffectRadiusIndex_1
	`
	items, err := LoadRows(dbHelper.db, query, ScanSpellEffect)
	if err != nil {
		return nil, fmt.Errorf("error loading SpellEffects: %w", err)
	}
	groupedBySpellID := make(map[int][]dbc.SpellEffect)
	for _, effect := range items {
		groupedBySpellID[effect.SpellID] = append(groupedBySpellID[effect.SpellID], effect)
	}

	RawSpellEffectBySpellIdAndIndex := make(map[int]map[int]dbc.SpellEffect)
	for spellID, effects := range groupedBySpellID {
		RawSpellEffectBySpellIdAndIndex[spellID] = CacheBy(effects, func(e dbc.SpellEffect) int {
			return e.EffectIndex
		})
	}
	json, _ := json.Marshal(RawSpellEffectBySpellIdAndIndex)
	if err := dbc.WriteGzipFile(fmt.Sprintf("%s/dbc/spell_effects.json", inputsDir), json); err != nil {
		panic(fmt.Sprintf("Error loading DBC data %v", err))
	}
	return RawSpellEffectBySpellIdAndIndex, nil
}

// The bonus tree nodes that hand an item one of its random suffixes: a bonus list of the default item
// context with stat rows and a suffix name. The few other nodes (trees 6050, 6130, 6138 and 6143) name
// a child tree, another item context or a list without stat rows. Selects from ItemXBonusTree x and
// ItemBonusTreeNode n.
var randomSuffixNodes = fmt.Sprintf(`
	FROM ItemXBonusTree x
	JOIN ItemBonusTreeNode n ON n.ParentItemBonusTreeID = x.ItemBonusTreeID
	WHERE n.ItemContext = 0 AND n.ChildItemBonusListID > 0
		AND EXISTS (SELECT 1 FROM ItemBonus sb WHERE sb.ParentItemBonusListID = n.ChildItemBonusListID AND sb.Type = %d)
		AND EXISTS (SELECT 1 FROM ItemBonus nb WHERE nb.ParentItemBonusListID = n.ChildItemBonusListID AND nb.Type = %d)`,
	dbc.ITEM_BONUS_STAT, dbc.ITEM_BONUS_SUFFIX)

type randomSuffixRow struct {
	listID, bonusType, value0, value1 int
	name                              string
	spellID                           int
}

func LoadAndWriteRawRandomSuffixes(dbHelper *DBHelper, inputsDir string) ([]dbc.RandomSuffix, error) {
	query := fmt.Sprintf(`
	SELECT b.ParentItemBonusListID, b.Type, b.Value_0, b.Value_1, COALESCE(d.Description_lang, ''), COALESCE(e.SpellID, 0)
	FROM ItemBonus b
	LEFT JOIN ItemNameDescription d ON b.Type = %[1]d AND d.ID = b.Value_0
	LEFT JOIN ItemEffect e ON b.Type = %[2]d AND e.ID = b.Value_0
	WHERE b.Type IN (%[1]d, %[2]d, %[3]d, %[5]d)
		AND b.ParentItemBonusListID IN (SELECT n.ChildItemBonusListID %[4]s)
	ORDER BY b.ParentItemBonusListID, b.OrderIndex
`, dbc.ITEM_BONUS_SUFFIX, dbc.ITEM_BONUS_ITEM_EFFECT, dbc.ITEM_BONUS_STAT, randomSuffixNodes, dbc.ITEM_BONUS_REDUCED_CASTER_DPS)
	rows, err := LoadRows(dbHelper.db, query, func(rows *sql.Rows) (randomSuffixRow, error) {
		var row randomSuffixRow
		err := rows.Scan(&row.listID, &row.bonusType, &row.value0, &row.value1, &row.name, &row.spellID)
		return row, err
	})
	if err != nil {
		return nil, fmt.Errorf("error loading random suffixes: %w", err)
	}

	var suffixes []dbc.RandomSuffix
	for _, row := range rows {
		if len(suffixes) == 0 || suffixes[len(suffixes)-1].ID != row.listID {
			suffixes = append(suffixes, dbc.RandomSuffix{ID: row.listID})
		}
		suffix := &suffixes[len(suffixes)-1]
		switch row.bonusType {
		case dbc.ITEM_BONUS_STAT:
			suffix.Stats = append(suffix.Stats, dbc.RandomSuffixStat{ItemMod: row.value0, Allocation: row.value1})
		case dbc.ITEM_BONUS_SUFFIX:
			suffix.Name = row.name
		case dbc.ITEM_BONUS_ITEM_EFFECT:
			suffix.EquipSpells = append(suffix.EquipSpells, row.spellID)
		case dbc.ITEM_BONUS_REDUCED_CASTER_DPS:
			suffix.ReducedCasterDps = true
		}
	}

	json, _ := json.Marshal(suffixes)
	if err := dbc.WriteGzipFile(fmt.Sprintf("%s/dbc/random_suffix.json", inputsDir), json); err != nil {
		panic(fmt.Sprintf("Error loading DBC data %v", err))
	}
	return suffixes, nil
}

func ScanConsumable(rows *sql.Rows) (dbc.Consumable, error) {
	var consumable dbc.Consumable
	var itemEffectsStr sql.NullString

	err := rows.Scan(
		&consumable.Id,
		&consumable.Name,
		&consumable.ItemLevel,
		&consumable.RequiredLevel,
		&consumable.ClassId,
		&consumable.SubClassId,
		&consumable.IconFileDataID,
		&consumable.SpellCategoryID,
		&consumable.SpellCategoryFlags,
		&itemEffectsStr,
		&consumable.ClassMask,
		&consumable.RaceMask,
		&consumable.RequiredSkill,
		&consumable.RequiredSkillRank,
		&consumable.RequiredHoliday,
		&consumable.TemporaryEnchant.ID,
		&consumable.TemporaryEnchant.Effects[0],
		&consumable.TemporaryEnchant.Effects[1],
		&consumable.TemporaryEnchant.Effects[2],
		&consumable.TemporaryEnchant.Points[0],
		&consumable.TemporaryEnchant.Points[1],
		&consumable.TemporaryEnchant.Points[2],
		&consumable.TemporaryEnchant.Args[0],
		&consumable.TemporaryEnchant.Args[1],
		&consumable.TemporaryEnchant.Args[2],
		&consumable.TemporaryEnchant.RequiredAreasID,
	)
	if err != nil {
		return consumable, fmt.Errorf("scanning consumable data: %w", err)
	}

	if !itemEffectsStr.Valid || itemEffectsStr.String == "" || itemEffectsStr.String == "null" {
		consumable.ItemEffects = []int{}
	} else {
		parts := strings.Split(itemEffectsStr.String, ",")
		effects := make([]int, 0, len(parts))
		for _, part := range parts {
			token := strings.TrimSpace(part)
			num, err := strconv.Atoi(token)
			if err != nil {
				fmt.Printf("Warning: parsing item effects for consumable %d (%s): %v\n", consumable.Id, token, err)
				effects = []int{}
				break
			}
			effects = append(effects, num)
		}
		consumable.ItemEffects = effects
	}

	return consumable, nil
}

// The SQL condition that the item name in column is not a placeholder: [PH] anywhere in it.
func notPlaceholderItemName(column string) string {
	return column + ` NOT LIKE '%[PH]%'`
}

// The items the consumable catalogue is built from, over Item as i and ItemSparse as s. Every
// Consumable-class item but the permanent enhancements of subclass 6, which the enchant table
// carries; the Trade Goods explosives; and any item whose on-use shares the explosives' category or
// casts a temporary enchant (Effect 54, or the imbue Effect 360, as IsTemporaryEnchantEffect reads
// it), whatever class the game files put it under: Crystal Charge 11566 is a Quest item and Lesser
// Mana Oil 20747 an Enchanting trade good. The item has to state an on-use spell, and one that also
// teaches a spell is an Engrave rune rather than a consumable. Test, QA, UNUSED, placeholder ([PH])
// and deprecated ([DEP]) names are left out: UNUSED Flask of Merciless Pursuit 215126.
func ConsumableItemFilter() string {
	onUse := func(condition string) string {
		return fmt.Sprintf(`EXISTS (SELECT 1 FROM ItemXItemEffect cx JOIN ItemEffect ce ON ce.ID = cx.ItemEffectID
			WHERE cx.ItemID = i.ID AND ce.TriggerType = %d AND ce.SpellID > 0 AND %s)`, dbc.ITEM_SPELLTRIGGER_ON_USE, condition)
	}
	return fmt.Sprintf(`s.Display_lang != '' AND s.Display_lang NOT LIKE '%%Test%%' AND s.Display_lang NOT LIKE 'QA%%'
		AND s.Display_lang NOT LIKE 'UNUSED%%' AND %s AND s.Display_lang NOT LIKE '%%[DEP]%%'
		AND %s
		AND NOT EXISTS (SELECT 1 FROM ItemXItemEffect lx JOIN ItemEffect le ON le.ID = lx.ItemEffectID
			WHERE lx.ItemID = i.ID AND le.TriggerType = %d)
		AND ((i.ClassID = 0 AND i.SubclassID != 6) OR (i.ClassID = 7 AND i.SubclassID = 2)
			OR %s
			OR %s)`,
		notPlaceholderItemName("s.Display_lang"),
		onUse("1"),
		dbc.ITEM_SPELLTRIGGER_LEARN_SPELL_ID,
		onUse(fmt.Sprintf("ce.SpellCategoryID = %d", dbc.EXPLOSIVES_SPELL_CATEGORY)),
		onUse(fmt.Sprintf("EXISTS (SELECT 1 FROM SpellEffect cse WHERE cse.SpellID = ce.SpellID AND cse.Effect IN (%d, %d))",
			dbcenums.E_ENCHANT_ITEM_TEMPORARY, dbcenums.E_360)))
}

func LoadAndWriteConsumables(dbHelper *DBHelper, inputsDir string) ([]dbc.Consumable, error) {
	cataloguedByID := catalogueConsumableIDs()
	allowListStrs := make([]string, 0, len(cataloguedByID))
	for _, id := range cataloguedByID {
		allowListStrs = append(allowListStrs, strconv.FormatInt(int64(id), 10))
	}
	allowList := strings.Join(allowListStrs, ", ")

	query := `
		SELECT
				i.ID,
				s.Display_lang AS Name,
				s.ItemLevel,
				s.RequiredLevel,
				i.ClassID,
				i.SubClassID,
				i.IconFileDataID,
				COALESCE(ie.SpellCategoryID, 0) AS SpellCategoryID,
				COALESCE(sc.Flags, 0) AS SpellCategoryFlags,
				(
					SELECT group_concat(ie2.ID, ',')
					FROM ItemEffect ie2
					JOIN ItemXItemEffect ixie2 ON ixie2.ItemEffectID = ie2.ID
					WHERE ixie2.ItemID = i.ID
				) AS ItemEffects,
				s.AllowableClass,
				(s.AllowableRace_1 << 32) | (s.AllowableRace_0 & 0xFFFFFFFF) AS RaceMask,
				s.RequiredSkill,
				s.RequiredSkillRank,
				s.RequiredHoliday,
				COALESCE(tsie.ID, 0),
				COALESCE(tsie.Effect_0, 0), COALESCE(tsie.Effect_1, 0), COALESCE(tsie.Effect_2, 0),
				COALESCE(tsie.EffectPointsMin_0, 0), COALESCE(tsie.EffectPointsMin_1, 0), COALESCE(tsie.EffectPointsMin_2, 0),
				COALESCE(tsie.EffectArg_0, 0), COALESCE(tsie.EffectArg_1, 0), COALESCE(tsie.EffectArg_2, 0),
				COALESCE((
					SELECT MAX(scr.RequiredAreasID) FROM SpellCastingRequirements scr
					WHERE scr.SpellID > 0 AND scr.SpellID IN (tsie.EffectArg_0, tsie.EffectArg_1, tsie.EffectArg_2)
				), 0)
			FROM Item i
			JOIN ItemSparse s ON i.ID = s.ID
			LEFT JOIN ItemXItemEffect ixie ON ixie.ItemID = i.ID
			LEFT JOIN ItemEffect ie ON ie.ID = ixie.ItemEffectID
			LEFT JOIN SpellCategory sc ON ie.SpellCategoryID = sc.ID
			LEFT JOIN SpellItemEnchantment tsie ON tsie.ID = (
				SELECT tse.EffectMiscValue_0
				FROM ItemXItemEffect tx
				JOIN ItemEffect te ON te.ID = tx.ItemEffectID
				JOIN SpellEffect tse ON tse.SpellID = te.SpellID
				WHERE tx.ItemID = i.ID AND te.TriggerType = ` + strconv.Itoa(dbc.ITEM_SPELLTRIGGER_ON_USE) + `
					AND tse.Effect IN (` + strconv.Itoa(int(dbcenums.E_ENCHANT_ITEM_TEMPORARY)) + `, ` + strconv.Itoa(int(dbcenums.E_360)) + `)
				LIMIT 1)
			WHERE (` + ConsumableItemFilter() + `) OR i.ID IN (` + allowList + `)
			GROUP BY i.ID
	`

	consumables, err := LoadRows(dbHelper.db, query, ScanConsumable)
	if err != nil {
		return nil, fmt.Errorf("error loading consumables: %w", err)
	}
	if err := applyCreatorClassMasks(dbHelper, consumables, func(c *dbc.Consumable) (int, *int) { return c.Id, &c.ClassMask }); err != nil {
		return nil, err
	}
	if consumables, err = mergeIdenticalConsumables(dbHelper, consumables, cataloguedByID); err != nil {
		return nil, err
	}

	fmt.Println("Loaded Consumables:", len(consumables))
	json, _ := json.Marshal(consumables)
	if err := dbc.WriteGzipFile(fmt.Sprintf("%s/dbc/consumables.json", inputsDir), json); err != nil {
		panic(fmt.Sprintf("Error loading DBC data %v", err))
	}
	return consumables, nil
}

// Items that share a display name, who may use them (class, race, skill and world event) and their item
// effect rows are one consumable, whatever their level, binding, price or flags: Minor Healing Potion
// 287966 is 118. An id the catalogue or ConsumableOverrides names stays and its unnamed twins go; among
// unnamed twins the lowest id stays. An item naming an item effect row the game files lack has no twin.
func mergeIdenticalConsumables(dbHelper *DBHelper, consumables []dbc.Consumable, catalogued []int32) ([]dbc.Consumable, error) {
	effects, err := LoadRows(dbHelper.db, itemEffectsQuery, ScanItemEffect)
	if err != nil {
		return nil, fmt.Errorf("error loading the consumables' item effects: %w", err)
	}
	effectsByID := CacheBy(effects, func(e dbc.ItemEffect) int { return e.ID })

	named := map[int]bool{}
	for _, id := range catalogued {
		named[int(id)] = true
	}
	for _, c := range ConsumableOverrides {
		named[int(c.Id)] = true
	}

	key := func(c dbc.Consumable) (string, bool) {
		rows := make([]string, 0, len(c.ItemEffects))
		for _, id := range c.ItemEffects {
			e, ok := effectsByID[id]
			if !ok {
				return "", false
			}
			rows = append(rows, fmt.Sprint(e.TriggerType, e.SpellID, e.Charges, e.CoolDownMSec, e.CategoryCoolDownMSec,
				e.SpellCategoryID, e.ChrSpecializationID))
		}
		slices.Sort(rows)
		return fmt.Sprintf("%s|%d|%d|%d|%d|%d|%s", c.Name, c.ClassMask, c.RaceMask, c.RequiredSkill, c.RequiredSkillRank,
			c.RequiredHoliday, strings.Join(rows, ";")), true
	}
	keys := make(map[int]string, len(consumables))
	keeper := map[string]int{}
	for _, c := range consumables {
		k, ok := key(c)
		if !ok {
			continue
		}
		keys[c.Id] = k
		if id, seen := keeper[k]; !seen || (named[c.Id] && !named[id]) || (named[c.Id] == named[id] && c.Id < id) {
			keeper[k] = c.Id
		}
	}
	return slices.DeleteFunc(consumables, func(c dbc.Consumable) bool {
		k, ok := keys[c.Id]
		return ok && !named[c.Id] && keeper[k] != c.Id
	}), nil
}

func ScanItemEffect(rows *sql.Rows) (dbc.ItemEffect, error) {
	var effect dbc.ItemEffect

	err := rows.Scan(
		&effect.ID,
		&effect.LegacySlotIndex,
		&effect.TriggerType,
		&effect.Charges,
		&effect.CoolDownMSec,
		&effect.CategoryCoolDownMSec,
		&effect.SpellCategoryID,
		&effect.SpellID,
		&effect.ChrSpecializationID,
		&effect.ParentItemID,
	)
	if err != nil {
		return effect, fmt.Errorf("scanning item effect data: %w", err)
	}

	return effect, nil
}

const itemEffectsQuery = `
	SELECT
		ie.ID,
		ie.LegacySlotIndex,
		ie.TriggerType,
		ie.Charges,
		ie.CoolDownMSec,
		ie.CategoryCoolDownMSec,
		ie.SpellCategoryID,
		ie.SpellID,
		ie.ChrSpecializationID,
		COALESCE(ixie.ItemID, 0) AS ParentItemID
	FROM ItemEffect ie
	LEFT JOIN ItemXItemEffect ixie ON ixie.ItemEffectID = ie.ID
	`

func LoadAndWriteItemEffects(dbHelper *DBHelper, inputsDir string) ([]dbc.ItemEffect, error) {
	effects, err := LoadRows(dbHelper.db, itemEffectsQuery, ScanItemEffect)
	if err != nil {
		return nil, fmt.Errorf("error loading item effects: %w", err)
	}

	fmt.Println("Loaded ItemEffects:", len(effects))
	json, _ := json.Marshal(effects)
	if err := dbc.WriteGzipFile(fmt.Sprintf("%s/dbc/item_effects.json", inputsDir), json); err != nil {
		panic(fmt.Sprintf("Error loading DBC data %v", err))
	}
	return effects, nil
}

func ScanItemSetSpell(rows *sql.Rows) (dbc.ItemSetSpell, error) {
	var spell dbc.ItemSetSpell

	err := rows.Scan(
		&spell.ItemSetID,
		&spell.Threshold,
		&spell.SpellID,
	)
	if err != nil {
		return spell, fmt.Errorf("scanning item set spell data: %w", err)
	}

	return spell, nil
}

const itemSetSpellsQuery = `
	SELECT
		iss.ItemSetID,
		iss.Threshold,
		iss.SpellID
	FROM ItemSetSpell iss
	ORDER BY iss.ID
	`

func LoadAndWriteItemSetSpells(dbHelper *DBHelper, inputsDir string) ([]dbc.ItemSetSpell, error) {
	spells, err := LoadRows(dbHelper.db, itemSetSpellsQuery, ScanItemSetSpell)
	if err != nil {
		return nil, fmt.Errorf("error loading item set spells: %w", err)
	}

	fmt.Println("Loaded ItemSetSpells:", len(spells))
	json, _ := json.Marshal(spells)
	if err := dbc.WriteGzipFile(fmt.Sprintf("%s/dbc/item_set_spells.json", inputsDir), json); err != nil {
		panic(fmt.Sprintf("Error loading DBC data %v", err))
	}
	return spells, nil
}

type RawTalent struct {
	DefinitionID   int
	TierID         int
	TalentName     string
	ColumnIndex    int
	ClassMask      int
	SpellID        int
	MaxRanks       int
	SpellRank      string
	PrereqRank     string
	PrereqTalent   string
	TabName        string
	BackgroundFile string
	PrereqRow      sql.NullInt64
	PrereqCol      sql.NullInt64
}

func ScanTalent(rows *sql.Rows) (RawTalent, error) {
	var talent RawTalent

	err := rows.Scan(
		&talent.TierID,
		&talent.TalentName,
		&talent.ColumnIndex,
		&talent.ClassMask,
		&talent.SpellRank,
		&talent.PrereqRank,
		&talent.PrereqTalent,
		&talent.TabName,
		&talent.BackgroundFile,
		&talent.PrereqRow,
		&talent.PrereqCol,
	)
	if err != nil {
		return talent, fmt.Errorf("scanning talent data: %w", err)
	}

	return talent, nil
}

func LoadTalents(dbHelper *DBHelper) ([]RawTalent, error) {
	query := `
SELECT
  t.TierID,
  COALESCE(sn.Name_lang, '') AS Name_lang,
  t.ColumnIndex,
  tb.ClassMask,
  t.SpellRank,
  t.PrereqRank,
  t.PrereqTalent,
  tb.Name_lang AS TabName,
  tb.ID as BackgroundFile,
  (SELECT t2.TierID
     FROM Talent t2
     WHERE t2.ID = (
         SELECT value
         FROM json_each(t.PrereqTalent)
         WHERE value <> 0
         LIMIT 1
     )
  ) AS PrereqRow,
  (SELECT t2.ColumnIndex
     FROM Talent t2
     WHERE t2.ID = (
         SELECT value
         FROM json_each(t.PrereqTalent)
         WHERE value <> 0
         LIMIT 1
     )
  ) AS PrereqCol
FROM Talent t
JOIN TalentTab tb ON t.TabID = tb.ID
JOIN SpellName sn ON sn.ID = t.SpellRank_0
ORDER BY tb.Name_lang;
`

	talents, err := LoadRows(dbHelper.db, query, ScanTalent)
	if err != nil {
		return nil, fmt.Errorf("error loading talents: %w", err)
	}

	fmt.Println("Loaded talents:", len(talents))
	return talents, nil
}

// Trait-driven talent loading.
//
// The Forever game files keep the talent trees in the Trait* tables instead of
// the legacy Talent/TalentTab pair. A tree is per class and holds all three
// specs side by side on a pixel lattice, so the generator has to recover the
// class, the tab and the grid coordinates that the picker and the protos still
// expect. LoadTraitTalents does that and hands back the same RawTalent rows the
// legacy loader produced, leaving everything downstream untouched.

// traitGridPitch is the spacing between two neighbouring nodes in TraitNode
// pixel coordinates. Both axes use it.
const traitGridPitch = 600

// traitTabGap is the smallest horizontal distance that separates two spec
// blocks inside a tree. Blocks sit at least 2200 apart while nodes inside a
// block are 600 apart, so any gap above this splits tabs.
const traitTabGap = 2 * traitGridPitch

type traitNodeRow struct {
	TreeID    int
	NodeID    int
	XrefIndex int
	XrefID    int
	PosX      int
	PosY      int
	MaxRanks  int
	SpellID   int
	DefID     int
	Name      string
}

func scanTraitNode(rows *sql.Rows) (traitNodeRow, error) {
	var n traitNodeRow
	err := rows.Scan(
		&n.TreeID,
		&n.NodeID,
		&n.XrefIndex,
		&n.XrefID,
		&n.PosX,
		&n.PosY,
		&n.MaxRanks,
		&n.SpellID,
		&n.DefID,
		&n.Name,
	)
	if err != nil {
		return n, fmt.Errorf("scanning trait node: %w", err)
	}
	return n, nil
}

type traitClassVote struct {
	TreeID    int
	ClassMask int
	Nodes     int
}

type traitSkillLine struct {
	NodeID int
	Name   string
}

type traitEdge struct {
	LeftNodeID  int
	RightNodeID int
}

type traitTab struct {
	ID         int
	ClassMask  int
	OrderIndex int
	Name       string
}

// traitGridPos is a node's place in the picker grid: which tab it belongs to
// and its cell inside that tab.
type traitGridPos struct {
	TabIdx int
	Row    int
	Col    int
}

// traitPlacedNode is a node with its grid cell and the pixel distance between
// the node and the centre of that cell.
type traitPlacedNode struct {
	Node     traitNodeRow
	Pos      traitGridPos
	Residual int
}

// filterTraitCandidates keeps a single node per key, preferring the one closest
// to its grid cell, and reports every node it drops.
func filterTraitCandidates(candidates []traitPlacedNode, key func(traitPlacedNode) any, onDrop func(loser, winner traitPlacedNode)) []traitPlacedNode {
	winners := map[any]traitPlacedNode{}
	for _, candidate := range candidates {
		other, seen := winners[key(candidate)]
		if !seen {
			winners[key(candidate)] = candidate
			continue
		}
		winner, loser := other, candidate
		if candidate.Residual < other.Residual {
			winner, loser = candidate, other
		}
		winners[key(candidate)] = winner
		onDrop(loser, winner)
	}

	kept := []traitPlacedNode{}
	for _, candidate := range candidates {
		if winners[key(candidate)].Node.NodeID == candidate.Node.NodeID {
			kept = append(kept, candidate)
		}
	}
	return kept
}

// selectTraitTrees picks one tree per class. The class comes from the class
// mask on the SkillLineAbility rows of the tree's spells (the mask is 0 for
// spells shared across classes, so the non-zero majority wins). Several trees
// can carry the same class - the game files keep stale drafts around - so the tree
// with the most nodes is taken as the live one.
func selectTraitTrees(dbHelper *DBHelper) (map[int]int, error) {
	votes, err := LoadRows(dbHelper.db, `
SELECT tn.TraitTreeID, sla.ClassMask, COUNT(DISTINCT tn.ID)
FROM TraitNode tn
JOIN TraitNodeXTraitNodeEntry x ON x.TraitNodeID = tn.ID
JOIN TraitNodeEntry e ON e.ID = x.TraitNodeEntryID
JOIN TraitDefinition d ON d.ID = e.TraitDefinitionID
JOIN SkillLineAbility sla ON sla.Spell = d.SpellID
WHERE sla.ClassMask <> 0
GROUP BY tn.TraitTreeID, sla.ClassMask
ORDER BY tn.TraitTreeID, sla.ClassMask
`, func(rows *sql.Rows) (traitClassVote, error) {
		var v traitClassVote
		err := rows.Scan(&v.TreeID, &v.ClassMask, &v.Nodes)
		return v, err
	})
	if err != nil {
		return nil, fmt.Errorf("error loading trait class masks: %w", err)
	}

	bestMask := map[int]traitClassVote{}
	for _, v := range votes {
		if cur, ok := bestMask[v.TreeID]; !ok || v.Nodes > cur.Nodes {
			bestMask[v.TreeID] = v
		}
	}

	counts, err := LoadRows(dbHelper.db, `SELECT TraitTreeID, COUNT(*) FROM TraitNode GROUP BY TraitTreeID`,
		func(rows *sql.Rows) (traitClassVote, error) {
			var v traitClassVote
			err := rows.Scan(&v.TreeID, &v.Nodes)
			return v, err
		})
	if err != nil {
		return nil, fmt.Errorf("error counting trait nodes: %w", err)
	}
	sizes := map[int]int{}
	for _, c := range counts {
		sizes[c.TreeID] = c.Nodes
	}

	live := map[int]int{}
	for treeID, vote := range bestMask {
		if cur, ok := live[vote.ClassMask]; !ok || sizes[treeID] > sizes[cur] {
			live[vote.ClassMask] = treeID
		}
	}
	return live, nil
}

// parkedTraitNodes names the nodes the game files park off their tree's canvas. A retired talent
// stays in its tree with one coordinate ten times a lattice one - Lightning Reflexes 1091/104982 at
// PosX 102800, Improved Serpent Sting 1091/105003 at PosY 39300, Holy Specialization 1114/105865 at
// PosY 21300 - while the talent that replaced it, if any, sits on the lattice. A node is parked when a
// coordinate is off the lattice and a tenth of it lands back on it, so a node placed by hand is never
// one.
func parkedTraitNodes(db *sql.DB) (map[int]bool, error) {
	nodes, err := LoadRows(db, `SELECT TraitTreeID, ID, PosX, PosY FROM TraitNode`,
		func(rows *sql.Rows) (traitNodeRow, error) {
			var n traitNodeRow
			err := rows.Scan(&n.TreeID, &n.NodeID, &n.PosX, &n.PosY)
			return n, err
		})
	if err != nil {
		return nil, fmt.Errorf("error loading trait node positions: %w", err)
	}

	xs, ys := map[int][]int{}, map[int][]int{}
	for _, n := range nodes {
		xs[n.TreeID] = append(xs[n.TreeID], n.PosX)
		ys[n.TreeID] = append(ys[n.TreeID], n.PosY)
	}

	parked := map[int]bool{}
	for _, n := range nodes {
		if parkedTraitCoord(n.PosX, xs[n.TreeID]) || parkedTraitCoord(n.PosY, ys[n.TreeID]) {
			parked[n.NodeID] = true
		}
	}
	return parked, nil
}

// A condition on TraitNode tn that leaves the parked nodes out.
func unparkedTraitNodes(db *sql.DB) (string, error) {
	parked, err := parkedTraitNodes(db)
	if err != nil || len(parked) == 0 {
		return "1 = 1", err
	}
	ids := []string{}
	for _, id := range slices.Sorted(maps.Keys(parked)) {
		ids = append(ids, strconv.Itoa(id))
	}
	return "tn.ID NOT IN (" + strings.Join(ids, ", ") + ")", nil
}

func parkedTraitCoord(value int, others []int) bool {
	nearLattice := func(v int) bool {
		for _, o := range others {
			if o != v && o-v <= traitGridPitch && v-o <= traitGridPitch {
				return true
			}
		}
		return false
	}
	return !nearLattice(value) && value%10 == 0 && nearLattice(value/10)
}

// traitGridIndex maps a pixel coordinate onto the lattice index.
func traitGridIndex(value, origin int) int {
	return (value - origin + traitGridPitch/2) / traitGridPitch
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

// LoadTraitTalents reads the class trees out of the Trait* tables and returns
// them as RawTalent rows, so the proto/json generators keep working unchanged.
func LoadTraitTalents(dbHelper *DBHelper) ([]RawTalent, error) {
	live, err := selectTraitTrees(dbHelper)
	if err != nil {
		return nil, err
	}
	parked, err := parkedTraitNodes(dbHelper.db)
	if err != nil {
		return nil, err
	}

	treeClass := map[int]int{}
	treeIDs := []int{}
	for mask, treeID := range live {
		treeClass[treeID] = mask
		treeIDs = append(treeIDs, treeID)
	}
	slices.Sort(treeIDs)

	nodes, err := LoadRows(dbHelper.db, `
SELECT
  tn.TraitTreeID,
  tn.ID,
  x.[Index],
  x.ID,
  tn.PosX,
  tn.PosY,
  e.MaxRanks,
  COALESCE(d.SpellID, 0),
  d.ID,
  COALESCE(NULLIF(sn.Name_lang, ''), d.OverrideName_lang, '') AS Name_lang
FROM TraitNode tn
JOIN TraitNodeXTraitNodeEntry x ON x.TraitNodeID = tn.ID
JOIN TraitNodeEntry e ON e.ID = x.TraitNodeEntryID
JOIN TraitDefinition d ON d.ID = e.TraitDefinitionID
LEFT JOIN SpellName sn ON sn.ID = d.SpellID
ORDER BY tn.TraitTreeID, tn.ID, x.[Index], x.ID
`, scanTraitNode)
	if err != nil {
		return nil, fmt.Errorf("error loading trait nodes: %w", err)
	}

	skillLines, err := LoadRows(dbHelper.db, `
SELECT DISTINCT x.TraitNodeID, sl.DisplayName_lang
FROM TraitNodeXTraitNodeEntry x
JOIN TraitNodeEntry e ON e.ID = x.TraitNodeEntryID
JOIN TraitDefinition d ON d.ID = e.TraitDefinitionID
JOIN SkillLineAbility sla ON sla.Spell = d.SpellID
JOIN SkillLine sl ON sl.ID = sla.SkillLine
WHERE sl.DisplayName_lang <> ''
ORDER BY x.TraitNodeID, sl.DisplayName_lang
`, func(rows *sql.Rows) (traitSkillLine, error) {
		var s traitSkillLine
		err := rows.Scan(&s.NodeID, &s.Name)
		return s, err
	})
	if err != nil {
		return nil, fmt.Errorf("error loading trait skill lines: %w", err)
	}
	nodeSkills := map[int][]string{}
	for _, s := range skillLines {
		nodeSkills[s.NodeID] = append(nodeSkills[s.NodeID], s.Name)
	}

	// TraitEdge.Type is the game files' gating kind: 2 makes the left node sufficient for the right,
	// 3 makes it required, and 0 only draws a line. Reading every row as a prerequisite turned
	// those drawn-only links into hard gates -- six node pairs carry one in each direction, and two
	// nodes cannot each gate the other.
	edges, err := LoadRows(dbHelper.db, `SELECT LeftTraitNodeID, RightTraitNodeID FROM TraitEdge WHERE Type IN (2, 3) ORDER BY RightTraitNodeID, LeftTraitNodeID`,
		func(rows *sql.Rows) (traitEdge, error) {
			var e traitEdge
			err := rows.Scan(&e.LeftNodeID, &e.RightNodeID)
			return e, err
		})
	if err != nil {
		return nil, fmt.Errorf("error loading trait edges: %w", err)
	}

	tabs, err := LoadRows(dbHelper.db, `SELECT ID, ClassMask, OrderIndex, Name_lang FROM TalentTab ORDER BY ClassMask, OrderIndex`,
		func(rows *sql.Rows) (traitTab, error) {
			var t traitTab
			err := rows.Scan(&t.ID, &t.ClassMask, &t.OrderIndex, &t.Name)
			return t, err
		})
	if err != nil {
		return nil, fmt.Errorf("error loading talent tabs: %w", err)
	}
	tabBackground := map[[2]int]int{}
	// TalentTab.Name_lang is the label the talent frame shows ("Elemental", "Shadow").
	// SkillLine.DisplayName_lang, which the block's majority vote yields, is the skill
	// line behind it ("Elemental Combat", "Shadow Magic"). Prefer the frame label and
	// fall back to the skill line where no tab row matches.
	tabLabel := map[[2]int]string{}
	for _, t := range tabs {
		tabBackground[[2]int{t.ClassMask, t.OrderIndex}] = t.ID
		tabLabel[[2]int{t.ClassMask, t.OrderIndex}] = t.Name
	}

	// One node can point at several entries (a choice node). Keep the lowest
	// Index, which is the entry shown first in game.
	choiceDiscards := 0
	byNode := map[int]traitNodeRow{}
	nodeOrder := map[int][]int{}
	for _, n := range nodes {
		if _, ok := treeClass[n.TreeID]; !ok {
			continue
		}
		if parked[n.NodeID] {
			fmt.Fprintf(os.Stderr, "[traits] tree %d node %d (%s at %d/%d): skipped, parked off the canvas\n", n.TreeID, n.NodeID, n.Name, n.PosX, n.PosY)
			continue
		}
		if prev, ok := byNode[n.NodeID]; ok {
			choiceDiscards++
			if prev.XrefIndex < n.XrefIndex || (prev.XrefIndex == n.XrefIndex && prev.XrefID <= n.XrefID) {
				continue
			}
		} else {
			nodeOrder[n.TreeID] = append(nodeOrder[n.TreeID], n.NodeID)
		}
		byNode[n.NodeID] = n
	}
	if choiceDiscards > 0 {
		fmt.Fprintf(os.Stderr, "[traits] dropped %d extra choice-node entries (kept the lowest Index)\n", choiceDiscards)
	}

	// Three pairs are joined by a sufficient edge in each direction, which the single-parent model
	// below reads as each node requiring the other -- a cycle leaving both unreachable. Collapse
	// such a pair onto the half that runs down the tree.
	hasEdge := map[[2]int]bool{}
	for _, e := range edges {
		hasEdge[[2]int{e.LeftNodeID, e.RightNodeID}] = true
	}
	isMirror := func(e traitEdge) bool {
		if !hasEdge[[2]int{e.RightNodeID, e.LeftNodeID}] {
			return false
		}
		left, right := byNode[e.LeftNodeID], byNode[e.RightNodeID]
		if left.PosY != right.PosY {
			return left.PosY > right.PosY
		}
		return e.LeftNodeID > e.RightNodeID
	}

	// A node can be the right side of several edges. Keep the lowest left node
	// id so the generated prereq is stable.
	prereqOf := map[int]int{}
	edgeDiscards := 0
	mirrorDiscards := 0
	for _, e := range edges {
		if _, ok := byNode[e.RightNodeID]; !ok {
			continue
		}
		if _, ok := byNode[e.LeftNodeID]; !ok {
			continue
		}
		if isMirror(e) {
			mirrorDiscards++
			continue
		}
		if cur, ok := prereqOf[e.RightNodeID]; ok {
			edgeDiscards++
			if cur <= e.LeftNodeID {
				continue
			}
		}
		prereqOf[e.RightNodeID] = e.LeftNodeID
	}
	if mirrorDiscards > 0 {
		fmt.Fprintf(os.Stderr, "[traits] dropped %d trait edges that mirror one running down the tree\n", mirrorDiscards)
	}
	if edgeDiscards > 0 {
		fmt.Fprintf(os.Stderr, "[traits] dropped %d extra incoming trait edges (kept the lowest left node id)\n", edgeDiscards)
	}

	var talents []RawTalent
	for _, treeID := range treeIDs {
		classMask := treeClass[treeID]
		treeNodes := []traitNodeRow{}
		for _, nodeID := range nodeOrder[treeID] {
			treeNodes = append(treeNodes, byNode[nodeID])
		}
		if len(treeNodes) == 0 {
			fmt.Fprintf(os.Stderr, "[traits] tree %d: no nodes\n", treeID)
			continue
		}

		// Split the tree into the spec blocks that sit next to each other.
		distinctX := []int{}
		for i := range treeNodes {
			if !slices.Contains(distinctX, treeNodes[i].PosX) {
				distinctX = append(distinctX, treeNodes[i].PosX)
			}
		}
		slices.Sort(distinctX)
		origins := []int{}
		for i, x := range distinctX {
			if i == 0 || x-distinctX[i-1] > traitTabGap {
				origins = append(origins, x)
			}
		}
		if len(origins) != 3 {
			fmt.Fprintf(os.Stderr, "[traits] tree %d: found %d tabs instead of 3\n", treeID, len(origins))
		}
		tabOf := func(x int) int {
			tab := 0
			for i, origin := range origins {
				if x >= origin {
					tab = i
				}
			}
			return tab
		}

		minY := treeNodes[0].PosY
		for i := range treeNodes {
			if treeNodes[i].PosY < minY {
				minY = treeNodes[i].PosY
			}
		}

		candidates := []traitPlacedNode{}
		for _, node := range treeNodes {
			if node.SpellID == 0 || node.MaxRanks == 0 || node.Name == "" {
				fmt.Fprintf(os.Stderr, "[traits] tree %d node %d: skipped, spell %d / %d ranks / name %q\n", treeID, node.NodeID, node.SpellID, node.MaxRanks, node.Name)
				continue
			}
			tabIdx := tabOf(node.PosX)
			candidate := traitPlacedNode{
				Node: node,
				Pos: traitGridPos{
					TabIdx: tabIdx,
					Row:    traitGridIndex(node.PosY, minY),
					Col:    traitGridIndex(node.PosX, origins[tabIdx]),
				},
			}
			candidate.Residual = abs(node.PosX-(origins[tabIdx]+candidate.Pos.Col*traitGridPitch)) +
				abs(node.PosY-(minY+candidate.Pos.Row*traitGridPitch))
			candidates = append(candidates, candidate)
		}

		// A tree can hold a talent twice, and two nodes can land on the same
		// cell when one of them is authored off the lattice. The first breaks protoc, which rejects the
		// repeated field name, the second breaks the picker, which rejects
		// duplicate locations, so the node sitting closest to its cell wins and
		// the other one is dropped.
		candidates = filterTraitCandidates(candidates, func(p traitPlacedNode) any { return p.Node.SpellID },
			func(loser, winner traitPlacedNode) {
				fmt.Fprintf(os.Stderr, "[traits] tree %d node %d (%s at %d/%d): skipped, spell %d is already used by node %d\n",
					treeID, loser.Node.NodeID, loser.Node.Name, loser.Node.PosX, loser.Node.PosY, loser.Node.SpellID, winner.Node.NodeID)
			})
		candidates = filterTraitCandidates(candidates, func(p traitPlacedNode) any { return p.Pos },
			func(loser, winner traitPlacedNode) {
				fmt.Fprintf(os.Stderr, "[traits] tree %d node %d (%s at %d/%d): skipped, tab %d row %d col %d is held by node %d\n",
					treeID, loser.Node.NodeID, loser.Node.Name, loser.Node.PosX, loser.Node.PosY,
					loser.Pos.TabIdx, loser.Pos.Row, loser.Pos.Col, winner.Node.NodeID)
			})

		positions := map[int]traitGridPos{}
		kept := []traitNodeRow{}
		for _, candidate := range candidates {
			positions[candidate.Node.NodeID] = candidate.Pos
			kept = append(kept, candidate.Node)
		}

		// The tab name is the skill line most of its nodes belong to; single
		// nodes can be listed under a second skill line (pet abilities, spells
		// shared between specs) and must not rename the tab.
		tabNames := make([]string, len(origins))
		for tabIdx := range origins {
			votes := map[string]int{}
			for _, node := range kept {
				if positions[node.NodeID].TabIdx != tabIdx {
					continue
				}
				for _, name := range nodeSkills[node.NodeID] {
					votes[name]++
				}
			}
			best := ""
			for name, count := range votes {
				if count > votes[best] || (count == votes[best] && (best == "" || name < best)) {
					best = name
				}
			}
			if best == "" {
				fmt.Fprintf(os.Stderr, "[traits] tree %d tab %d: no skill line found\n", treeID, tabIdx)
			}
			tabNames[tabIdx] = best
		}

		for _, node := range kept {
			pos := positions[node.NodeID]
			if node.SpellID == 0 {
				return nil, fmt.Errorf("trait node %d (%s) has no spell", node.NodeID, node.Name)
			}
			talent := RawTalent{
				DefinitionID:   node.DefID,
				TierID:         pos.Row,
				TalentName:     node.Name,
				ColumnIndex:    pos.Col,
				ClassMask:      classMask,
				SpellID:        node.SpellID,
				MaxRanks:       node.MaxRanks,
				TabName:        tabDisplayName(tabLabel, classMask, pos.TabIdx, tabNames[pos.TabIdx]),
				BackgroundFile: strconv.Itoa(tabBackground[[2]int{classMask, pos.TabIdx}]),
			}
			if prereq, ok := prereqOf[node.NodeID]; ok {
				prereqPos, mapped := positions[prereq]
				switch {
				case !mapped:
					fmt.Fprintf(os.Stderr, "[traits] tree %d node %d: prereq node %d was not mapped\n", treeID, node.NodeID, prereq)
				case prereqPos.TabIdx != pos.TabIdx:
					fmt.Fprintf(os.Stderr, "[traits] tree %d node %d: prereq node %d sits in another tab\n", treeID, node.NodeID, prereq)
				default:
					talent.PrereqRow = sql.NullInt64{Int64: int64(prereqPos.Row), Valid: true}
					talent.PrereqCol = sql.NullInt64{Int64: int64(prereqPos.Col), Valid: true}
				}
			}
			talents = append(talents, talent)
		}

		fmt.Fprintf(os.Stderr, "[traits] class mask %d: tree %d, %d talents (%s)\n", classMask, treeID, len(kept), strings.Join(tabNames, ", "))
	}

	slices.SortStableFunc(talents, func(a, b RawTalent) int {
		return cmp.Or(
			cmp.Compare(a.ClassMask, b.ClassMask),
			cmp.Compare(a.TabName, b.TabName),
			cmp.Compare(a.TierID, b.TierID),
			cmp.Compare(a.ColumnIndex, b.ColumnIndex),
		)
	})

	fmt.Println("Loaded trait talents:", len(talents))
	return talents, nil
}

type SpellIcon struct {
	SpellID int
	FDID    int
	HasBuff bool
	Name    string
	Rank    int
}

var spellRankRegex = regexp.MustCompile(`Rank ([0-9]+)`)

func ScanSpellIcon(rows *sql.Rows) (SpellIcon, error) {
	var icon SpellIcon
	var nameSubtext string
	err := rows.Scan(
		&icon.SpellID,
		&icon.FDID,
		&icon.HasBuff,
		&icon.Name,
		&nameSubtext,
	)
	if err != nil {
		fmt.Println(icon.Name, err)
		return icon, fmt.Errorf("scanning talent data: %w", err)
	}

	if spellRankRegex.MatchString(nameSubtext) {
		rank, _ := strconv.Atoi(spellRankRegex.FindStringSubmatch(nameSubtext)[1])
		if rank != 0 {
			icon.Rank = rank
		}
	}

	return icon, nil
}

func LoadSpellIcons(dbHelper *DBHelper) (map[int]SpellIcon, error) {
	query := `
SELECT
	sm.SpellID,
	sm.SpellIconFileDataID,
	(
		(ss.AuraDescription_lang != '' and ss.AuraDescription_lang is not null)
	) AS HasBuff,
	COALESCE(sn.Name_lang, '') AS Name_lang,
	COALESCE(ss.NameSubtext_lang, "")
FROM SpellMisc sm
LEFT JOIN Spell ss ON ss.ID = sm.SpellID
LEFT JOIN SpellName sn ON sn.ID = sm.SpellID
`

	talents, err := LoadRows(dbHelper.db, query, ScanSpellIcon)
	if err != nil {
		return nil, fmt.Errorf("error loading spellicons: %w", err)
	}
	iconsByID := make(map[int]SpellIcon, len(talents))
	for _, icon := range talents {
		iconsByID[icon.SpellID] = icon
	}
	fmt.Println("Loaded spellicons:", len(talents))
	return iconsByID, nil
}

var iconsMap, _ = LoadArtTexturePaths(ListfilePath)

func ScanSpells(rows *sql.Rows) (dbc.Spell, error) {
	var spell dbc.Spell

	var stringAttr string
	var stringClassMask string
	var stringProcType string              //2
	var stringAuraIFlags string            //2
	var stringChannelInterruptFlags string // 2
	var stringShapeShift string            //2
	var iconId int                         //
	err := rows.Scan(
		&spell.NameLang,
		&spell.ID,
		&spell.SchoolMask,
		&spell.Speed,
		&spell.LaunchDelay,
		&spell.MinDuration,
		&spell.MaxScalingLevel,
		&spell.MinScalingLevel,
		&spell.ScalesFromItemLevel,
		&spell.SpellLevel,
		&spell.BaseLevel,
		&spell.MaxLevel,
		&spell.MaxPassiveAuraLevel,
		&spell.Cooldown,
		&spell.CategoryRecoveryTime,
		&spell.GCD,
		&spell.MinRange,
		&spell.MaxRange,
		&stringAttr,
		&spell.CategoryFlags,
		&spell.MaxCharges,
		&spell.ChargeRecoveryTime,
		&spell.CategoryTypeMask,
		&spell.Category,
		&spell.DefenseType,
		&spell.Duration,
		&spell.ProcChance,
		&spell.ProcCharges,
		&stringProcType,
		&spell.ProcCategoryRecovery,
		&spell.EquippedItemClass,
		&spell.EquippedItemInvTypes,
		&spell.EquippedItemSubclass,
		&spell.CastTimeMin,
		&stringClassMask,
		&spell.SpellClassSet,
		&stringAuraIFlags,
		&stringChannelInterruptFlags,
		&stringShapeShift,
		&spell.Description,
		&spell.Variables,
		&spell.MaxCumulativeStacks,
		&spell.MaxTargets,
		&spell.RequiredAreasID,
		&iconId,
	)
	if err != nil {
		return spell, fmt.Errorf("scanning spell data: %w", err)
	}

	spell.Attributes, err = parseIntArrayField(stringAttr, 17)
	if err != nil {
		return spell, fmt.Errorf("parsing attributes args for spell %d (%s): %w", spell.ID, stringAttr, err)
	}
	spell.SpellClassMask, err = parseIntArrayField(stringClassMask, 4)
	if err != nil {
		return spell, fmt.Errorf("parsing classmask args for spell %d (%s): %w", spell.ID, stringClassMask, err)
	}

	spell.ProcTypeMask, err = parseIntArrayField(stringProcType, 2)
	if err != nil {
		return spell, fmt.Errorf("parsing ProcTypeMask args for spell %d (%s): %w", spell.ID, stringProcType, err)
	}
	spell.AuraInterruptFlags, err = parseIntArrayField(stringAuraIFlags, 2)
	if err != nil {
		return spell, fmt.Errorf("parsing stringAuraIFlags args for spell %d (%s): %w", spell.ID, stringAuraIFlags, err)
	}
	spell.ChannelInterruptFlags, err = parseIntArrayField(stringChannelInterruptFlags, 2)
	if err != nil {
		return spell, fmt.Errorf("parsing stringChannelInterruptFlags args for spell %d (%s): %w", spell.ID, stringChannelInterruptFlags, err)
	}
	spell.ShapeshiftMask, err = parseIntArrayField(stringShapeShift, 2)
	if err != nil {
		return spell, fmt.Errorf("parsing stringShapeShift args for spell %d (%s): %w", spell.ID, stringShapeShift, err)
	}
	spell.IconPath = iconsMap[iconId]
	return spell, nil
}

func LoadAndWriteSpells(dbHelper *DBHelper, inputsDir string) ([]dbc.Spell, error) {
	query := `
	SELECT DISTINCT
	sn.Name_lang,
	sn.ID,
	COALESCE(sm.SchoolMask, 0),
	COALESCE(sm.Speed, 0),
	COALESCE(sm.LaunchDelay, 0),
	COALESCE(sm.MinDuration, 0),
	COALESCE(ss.MaxScalingLevel, 0),
	COALESCE(ss.MinScalingLevel, 0),
	0, -- SpellScaling.ScalesFromItemLevel: dropped from the game files
	COALESCE(sl.SpellLevel, 0),
	COALESCE(sl.BaseLevel, 0),
	COALESCE(sl.MaxLevel, 0),
	COALESCE(sl.MaxPassiveAuraLevel, 0),
	COALESCE(sc.RecoveryTime, 0),
	COALESCE(sc.CategoryRecoveryTime, 0),
	COALESCE(sc.StartRecoveryTime, 0),
	COALESCE(sr.RangeMin_0, 0.0),
	COALESCE(sr.RangeMax_0, 0.0),
	COALESCE(sm."Attributes", ""),
	COALESCE(ssc.Flags, 0),
	COALESCE(ssc.MaxCharges, 0),
	COALESCE(ssc.ChargeRecoveryTime, 0),
	COALESCE(ssc.TypeMask, 0),
	COALESCE(scs.Category, 0),
	COALESCE(scs.DefenseType, 0),
	COALESCE(sd.Duration, 0),
	COALESCE(sao.ProcChance, 0),
	COALESCE(sao.ProcCharges, 0),
	COALESCE(sao.ProcTypeMask, ""),
	COALESCE(sao.ProcCategoryRecovery, 0),
	COALESCE(sei.EquippedItemClass, 0),
	COALESCE(sei.EquippedItemInvTypes, 0),
	COALESCE(sei.EquippedItemSubclass, 0),
	0, -- SpellScaling.CastTimeMin: dropped from the game files
	COALESCE(sco.SpellClassMask, ""),
	COALESCE(sco.SpellClassSet, 0),
	COALESCE(si.AuraInterruptFlags, ""),
	COALESCE(si.ChannelInterruptFlags, ""),
	COALESCE(ssp.ShapeshiftMask, ""),
	COALESCE(s.Description_lang, ""),
	COALESCE(sdv.Variables, ""),
	COALESCE(sao.CumulativeAura, 0),
	COALESCE(str.MaxTargets, 0),
	COALESCE(scr.RequiredAreasID, 0),
	COALESCE(sm.SpellIconFileDataID, 0)
FROM
    Spell as s
	LEFT JOIN SpellName sn ON s.ID = sn.ID
	LEFT JOIN SpellEffect se ON s.ID = se.SpellID
	LEFT JOIN (
		SELECT
			*
		FROM
			SpellMisc
		GROUP BY
			SpellID
	) sm ON s.ID = sm.SpellID
	LEFT JOIN (
		SELECT
			*
		FROM
			SpellLevels
		GROUP BY
			SpellID
	) sl ON s.ID = sl.SpellID
	LEFT JOIN (
		SELECT
			*
		FROM
			SpellCooldowns
		GROUP BY
			SpellID
	) sc ON s.ID = sc.SpellID
	LEFT JOIN (
		SELECT
			*
		FROM
			SpellScaling
		GROUP BY
			SpellID
	) ss ON s.ID = ss.SpellID
	LEFT JOIN SpellLabel slb ON s.ID = slb.SpellID
	LEFT JOIN (
		SELECT
			*
		FROM
			SpellCategories
		GROUP BY
			SpellID
	) scs ON s.ID = scs.SpellID
	LEFT JOIN SpellCategory ssc ON ssc.ID = scs.Category
	LEFT JOIN SpellDuration sd ON sm.DurationIndex = sd.ID
	LEFT JOIN SpellPower sp ON sp.SpellID = s.ID
	LEFT JOIN SpellInterrupts si ON si.SpellID = s.ID
	LEFT JOIN SpellEquippedItems sei ON sei.SpellID = s.ID
	LEFT JOIN (
		SELECT
			*
		FROM
			SpellAuraOptions
		GROUP BY
			SpellID
	) sao ON sao.SpellID = s.ID
	LEFT JOIN SpellClassOptions sco ON s.ID = sco.SpellID
	LEFT JOIN SpellShapeshift ssp ON ssp.SpellID = s.ID
	LEFT JOIN SpellXDescriptionVariables sxd ON s.ID = sxd.SpellID
	LEFT JOIN SpellDescriptionVariables sdv ON sdv.ID = sxd.SpellDescriptionVariablesID
	LEFT JOIN (
		SELECT
			*
		FROM
			SpellTargetRestrictions
		GROUP BY
			SpellID
	) str ON s.ID = str.SpellID
	LEFT JOIN (
		SELECT
			*
		FROM
			SpellCastingRequirements
		GROUP BY
			SpellID
	) scr ON s.ID = scr.SpellID
	LEFT JOIN SpellRange sr ON sr.ID = sm.RangeIndex
	GROUP BY s.ID
	ORDER BY s.ID asc
`

	spells, err := LoadRows(dbHelper.db, query, ScanSpells)
	if err != nil {
		return nil, fmt.Errorf("error loading spells: %w", err)
	}

	fmt.Println("Loaded spells:", len(spells))
	json, _ := json.Marshal(spells)
	if err := dbc.WriteGzipFile(fmt.Sprintf("%s/dbc/spells.json", inputsDir), json); err != nil {
		panic(fmt.Sprintf("Error loading DBC data %v", err))
	}
	return spells, nil
}

func LoadAndWriteEnchantDescriptions(outputPath string, db *WowDatabase, instance *dbc.DBC) error {
	descriptions := make(map[int32]string)

	dataProvider := tooltip.DBCTooltipDataProvider{DBC: instance}
	for _, enchant := range db.Enchants {
		dbcEnch := instance.EnchantsByEffectId[int(enchant.EffectId)]
		tooltip, err := tooltip.ParseTooltip(dbcEnch.EffectName, dataProvider, int64(enchant.EffectId))
		if err != nil {
			fmt.Printf("Could not parse enchant (%d), '%s'\n", enchant.EffectId, dbcEnch.EffectName)
		} else {
			descriptions[enchant.EffectId] = tooltip.String()
		}
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(descriptions); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

func ScanDropRow(rows *sql.Rows) (itemID int, ds *proto.DropSource, instanceName string, err error) {
	var (
		dropSource   proto.DropSource
		jiName       string
		instanceType int
	)
	err = rows.Scan(
		&itemID,
		&dropSource.NpcId,
		&dropSource.ZoneId,
		&dropSource.OtherName,
		&jiName,
		&instanceType,
	)
	if err != nil {
		return 0, nil, "", fmt.Errorf("scanning drop row: %w", err)
	}

	dropSource.Difficulty = ParseDungeonDifficulty(instanceType == 2)
	return itemID, &dropSource, jiName, nil
}

func LoadAndWriteDropSources(dbHelper *DBHelper, inputsDir string) (
	sourcesByItem map[int][]*proto.DropSource,
	namesByZone map[int]string,
	err error,
) {
	const query = `
		SELECT DISTINCT
		jei.ItemID,
		je.ID                               AS NpcId,
		COALESCE(COALESCE(
			NULLIF(ji.AreaID, 0),
			at.ID
		), 0)                                AS ZoneId,
		je.Name_lang                     AS OtherName,
		ji.Name_lang,
		Map.InstanceType
		FROM JournalEncounterItem AS jei
		INNER JOIN JournalEncounter AS je
		ON je.ID = jei.JournalEncounterID
		ON ji.ID = je.JournalInstanceID
		LEFT JOIN AreaTable AS at
		ON (
			at.ZoneName       = ji.Name_lang
			OR at.AreaName_lang  = ji.Name_lang
		)
		LEFT JOIN Map ON Map.ID = ji.MapID
		GROUP BY jei.ItemID
    `

	//rows, err := dbHelper.db.Query(query)
	// if err != nil {
	// 	return nil, nil, fmt.Errorf("querying drop sources: %w", err)
	// }
	// defer rows.Close()

	sourcesByItem = make(map[int][]*proto.DropSource)
	namesByZone = make(map[int]string)

	// for rows.Next() {
	// 	itemID, ds, jiName, scanErr := ScanDropRow(rows)
	// 	if scanErr != nil {
	// 		return nil, nil, scanErr
	// 	}
	// 	sourcesByItem[itemID] = append(sourcesByItem[itemID], ds)
	// 	namesByZone[int(ds.ZoneId)] = jiName
	// }
	// if err = rows.Err(); err != nil {
	// 	return nil, nil, fmt.Errorf("iterating drop rows: %w", err)
	// }
	json, _ := json.Marshal(sourcesByItem)
	if err := dbc.WriteGzipFile(fmt.Sprintf("%s/dbc/dropSources.json", inputsDir), json); err != nil {
		log.Fatalf("Error writing file: %v", err)
	}
	return sourcesByItem, namesByZone, nil
}

func ScanCraftedItems(rows *sql.Rows) (itemID int, ds *proto.CraftedSource, err error) {
	var (
		profId  int
		crafted proto.CraftedSource
	)

	err = rows.Scan(
		&itemID,
		&profId,
		&crafted.SpellId,
	)
	crafted.Profession = dbc.GetProfession(profId)
	if err != nil {
		return 0, nil, fmt.Errorf("scanning drop row: %w", err)
	}
	return itemID, &crafted, nil
}

func LoadCraftedItems(dbHelper *DBHelper) (
	sourcesByItem map[int][]*proto.CraftedSource,
) {
	const query = `
		SELECT se.EffectItemType, sla.SkillLine, sla.Spell FROM SkillLineAbility sla
		LEFT JOIN SpellEffect se ON sla.Spell == se.SpellID
		WHERE se.Effect = 24
    `

	rows, err := dbHelper.db.Query(query)
	if err != nil {
		return nil
	}
	defer rows.Close()

	sourcesByItem = make(map[int][]*proto.CraftedSource)

	for rows.Next() {
		itemID, ds, scanErr := ScanCraftedItems(rows)
		if scanErr != nil {
			return nil
		}
		sourcesByItem[itemID] = append(sourcesByItem[itemID], ds)
	}
	if err = rows.Err(); err != nil {
		return nil
	}

	return sourcesByItem
}

func ScanRepItems(rows *sql.Rows) (itemID int, factionName string, ds *proto.RepSource, err error) {
	var (
		minReputation      int
		rep                proto.RepSource
		reputationRaceMask dbc.Race
	)

	err = rows.Scan(
		&itemID,
		&rep.RepFactionId,
		&factionName,
		&reputationRaceMask,
		&minReputation,
	)
	if err != nil {
		return 0, "", nil, fmt.Errorf("scanning rep row: %w", err)
	}

	// A side's faction names that side's races in its first race mask (the other side starts Hated in the
	// second); a first mask naming races of both sides is every race's faction.
	alliance := dbc.AlliedRaces.Matches(reputationRaceMask)
	horde := dbc.HordeRaces.Matches(reputationRaceMask)
	if alliance && !horde {
		rep.FactionId = proto.Faction_Alliance
	} else if horde && !alliance {
		rep.FactionId = proto.Faction_Horde
	}

	rep.RepLevel = dbc.GetRepLevel(minReputation)
	return itemID, factionName, &rep, nil
}

// The items whose ItemSparse row asks for a standing with a faction, and the names of those factions.
func LoadRepItems(dbHelper *DBHelper) (
	sourcesByItem map[int][]*proto.RepSource,
	factionNames map[int32]string,
	err error,
) {
	const query = `
		SELECT isp.ID, fa.ID, fa.Name_lang,
			(COALESCE(fa.ReputationRaceMasks0_1, 0) << 32) | (COALESCE(fa.ReputationRaceMasks0_0, 0) & 0xFFFFFFFF) AS ReputationRaceMask,
			isp.MinReputation FROM ItemSparse isp
		JOIN Faction fa on fa.ID = isp.MinFactionID
		WHERE isp.MinFactionID > 0
    `

	rows, err := dbHelper.db.Query(query)
	if err != nil {
		return nil, nil, fmt.Errorf("querying rep items: %w", err)
	}
	defer rows.Close()

	sourcesByItem = make(map[int][]*proto.RepSource)
	factionNames = make(map[int32]string)

	for rows.Next() {
		itemID, factionName, ds, scanErr := ScanRepItems(rows)
		if scanErr != nil {
			return nil, nil, scanErr
		}
		sourcesByItem[itemID] = append(sourcesByItem[itemID], ds)
		factionNames[ds.RepFactionId] = factionName
	}
	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("reading rep items: %w", err)
	}
	return sourcesByItem, factionNames, nil
}

func ScanItemUpgradePath(rows *sql.Rows) (UpgradeID int, upgradePathID int, ilvl int, err error) {
	err = rows.Scan(
		&UpgradeID,
		&upgradePathID,
		&ilvl,
	)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("scanning rep row: %w", err)
	}
	return UpgradeID, upgradePathID, ilvl, nil
}

func LoadItemUpgradePath(dbHelper *DBHelper) (upgradePath map[int][]int, err error) {
	const query = `SELECT
		iu.ID,
		iu.ItemUpgradePathID,
		iu.ItemLevelIncrement
		FROM ItemUpgrade iu
    `
	rows, err := dbHelper.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	upgradeIDToPathID := make(map[int]int)
	pathIDToIlvls := make(map[int][]int)

	for rows.Next() {
		ID, pathID, ilvl, scanErr := ScanItemUpgradePath(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		upgradeIDToPathID[ID] = pathID
		pathIDToIlvls[pathID] = append(pathIDToIlvls[pathID], ilvl)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	upgradePath = make(map[int][]int)
	for id, upgrade := range upgradeIDToPathID {
		upgradePath[id] = pathIDToIlvls[upgrade]
	}
	for _, ilvls := range upgradePath {
		slices.Sort(ilvls)
	}
	fmt.Println("Loaded Upgrade Path", len(upgradePath))
	return upgradePath, nil
}

// tabDisplayName prefers the talent frame's own label for a tree, falling back to the
// skill-line name derived from the block's spells when no TalentTab row matches.
func tabDisplayName(labels map[[2]int]string, classMask, tabIdx int, fallback string) string {
	if name, ok := labels[[2]int{classMask, tabIdx}]; ok && name != "" {
		return name
	}
	return fallback
}

// LoadRaidZoneIDs returns the AreaTable ids of the raid zones: the zone a Map with InstanceType 2 names in
// AreaTableID, and every area whose ContinentID is such a Map. The first path finds each vanilla raid,
// Onyxia's Lair included, whose ContinentID is Kalimdor. The second adds the raids' sub-areas, such as
// Zul'Gurub's and Ruins of Ahn'Qiraj's, and the raid maps that leave AreaTableID 0.
func LoadRaidZoneIDs(dbHelper *DBHelper) (map[int32]bool, error) {
	return loadIDSet(dbHelper, `
		SELECT ID FROM AreaTable
		WHERE ID IN (SELECT AreaTableID FROM Map WHERE InstanceType = 2)
		OR ContinentID IN (SELECT ID FROM Map WHERE InstanceType = 2)
	`, "raid zones")
}

// LoadAreaTableIDs returns every AreaTable id in the game files.
func LoadAreaTableIDs(dbHelper *DBHelper) (map[int32]bool, error) {
	return loadIDSet(dbHelper, `SELECT ID FROM AreaTable`, "AreaTable ids")
}

func loadIDSet(dbHelper *DBHelper, query, what string) (map[int32]bool, error) {
	ids, err := LoadRows(dbHelper.db, query, func(rows *sql.Rows) (int32, error) {
		var id int32
		return id, rows.Scan(&id)
	})
	if err != nil {
		return nil, fmt.Errorf("loading %s: %w", what, err)
	}
	set := make(map[int32]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set, nil
}

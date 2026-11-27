package database

import (
	"cmp"
	"database/sql"
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/dbc"
	"github.com/wowsims/forever/tools/tooltip"
)

// The site's tooltips (ui/sim/proto/tooltips.ts): spells by id, and the parts of an item's tooltip the item
// database does not hold, by item id, with its sets by set id and the enchants' names by effect id.
type tooltipFile struct {
	Spells   map[int32]spellTooltip `json:"spells"`
	Items    map[int32]itemTooltip  `json:"items"`
	Sets     map[int32]setTooltip   `json:"sets"`
	Enchants map[int32]string       `json:"enchants"` // The name the game writes on an enchanted item (SpellItemEnchantment.Name_lang).
}

// One spell's tooltip: its name, rank and icon, the header lines above its description as [left, right] pairs
// (cost, range, cast time, cooldown, requirements, tools and reagents; headerLines), and its description and buff
// text compiled by tooltip.Compile. A side of a line is a string, or the segments of one whose number the player's
// talents can change.
type spellTooltip struct {
	Name  string            `json:"n"`
	Rank  string            `json:"r,omitempty"`
	Icon  string            `json:"i,omitempty"`
	Lines [][2]any          `json:"h,omitempty"`
	Text  []tooltip.Segment `json:"d,omitempty"`
	Aura  []tooltip.Segment `json:"a,omitempty"`
	// A talent's points and, where its curve gives each rank its own amounts, the description at each rank
	// from rank 1; the description above is the top rank's, as the row states it.
	MaxPoints int32               `json:"m,omitempty"`
	RankTexts [][]tooltip.Segment `json:"x,omitempty"`
}

// An item's binding (1 when picked up, 2 when equipped, 3 when used), its flavour text and its spells as
// [trigger, spell id, cooldown in ms], with the charges as a fourth where the game writes them: trigger 0 is a use,
// 1 an equip effect and 2 a chance on hit. An item the item database does not hold, a consumable or an enchant's
// item, carries its name, quality, icon, levels, slot and classes too.
type itemTooltip struct {
	Binding int       `json:"b,omitempty"`
	Flavor  string    `json:"f,omitempty"`
	Spells  [][]int32 `json:"e,omitempty"`
	itemFacts
	Name          string        `json:"n,omitempty"`
	Quality       int32         `json:"q,omitempty"`
	Icon          string        `json:"i,omitempty"`
	ItemLevel     int32         `json:"l,omitempty"`
	RequiredLevel int32         `json:"r,omitempty"`
	InventoryType int32         `json:"o,omitempty"` // Such as 1 for the head or 24 for ammo.
	Type          string        `json:"t,omitempty"` // The subclass the game names beside the slot or a bag's slots, such as Arrow or Quiver.
	Slots         int32         `json:"c,omitempty"` // A bag's or a quiver's slots.
	AmmoDPS       float64       `json:"d,omitempty"` // The damage per second an arrow or bullet adds.
	Classes       []proto.Class `json:"a,omitempty"` // The classes that may use it; none for every class.
}

// The lines of an item's tooltip read off its ItemSparse row that the item database does not carry.
type itemFacts struct {
	SellPrice  int32            `json:"s,omitempty"` // In copper.
	Unique     int32            `json:"u,omitempty"` // The most a player may carry ("Unique" at 1, "Unique (N)" above), or -1 for an item only one of which may be worn ("Unique-Equipped").
	Skill      *itemRequirement `json:"k,omitempty"` // The profession and the skill in it the item asks for.
	Reputation *itemRequirement `json:"g,omitempty"` // The faction and the standing the item asks for, from 0 for Hated to 7 for Exalted.
}

type itemRequirement struct {
	Name string `json:"n"`
	Rank int32  `json:"r,omitempty"`
}

// A set's name, its pieces the database holds in the order the game lists them, and its bonuses as [pieces worn,
// spell id].
type setTooltip struct {
	Name    string     `json:"n"`
	Pieces  []int32    `json:"p"`
	Bonuses [][2]int32 `json:"b"`
}

// The item spell triggers a tooltip shows: use, equip and chance on hit.
var shownItemTriggers = map[int]bool{0: true, 1: true, 2: true}

// Writes the tooltips of every spell and item in the game files, in two files. path holds what the site shows on
// its pages: the spells the sim's store holds, every talent in the trees (their spell ids and points), and every
// item, consumable, enchant and spell the given databases name, with the spells of those items and of their sets.
// The file beside it, _extra.json, holds every other item and named spell, which the site loads the first time it
// shows one, so any id a page shows has a tooltip. A condition on a rank of a talent the trees offer is left to the
// player's points, one on a talent they do not offer is false; the game files rows give the header lines.
func WriteTooltips(path string, helper *DBHelper, instance *dbc.DBC, treeTalents map[int32]int32, rawTalents []RawTalent, databases ...*proto.UIDatabase) error {
	ids := map[int32]bool{}
	inDatabase := map[int32]bool{}
	var itemIDs, otherIDs []int32
	for _, db := range databases {
		for _, item := range db.Items {
			itemIDs = append(itemIDs, item.Id)
			inDatabase[item.Id] = true
		}
	}
	other := func(id int32) {
		if id != 0 && !inDatabase[id] {
			otherIDs = append(otherIDs, id)
		}
	}
	for _, db := range databases {
		for _, data := range db.SpellIcons {
			ids[data.Id] = true
		}
		for _, enchant := range db.Enchants {
			if enchant.SpellId != 0 {
				ids[enchant.SpellId] = true
			}
			other(enchant.ItemId)
		}
		for _, consumable := range db.Consumables {
			other(consumable.Id)
		}
		for _, data := range db.ItemIcons {
			other(data.Id)
		}
	}
	every, err := wholeItemTooltips(helper, instance)
	if err != nil {
		return err
	}
	setPieces, err := setPieceOrder(helper)
	if err != nil {
		return err
	}
	items, sets := itemTooltips(instance, itemIDs, every, setPieces)
	enchants, err := enchantNames(helper, databases)
	if err != nil {
		return err
	}
	for _, id := range otherIDs {
		if item, ok := every[id]; ok {
			items[id] = item
		}
	}
	for _, item := range items {
		for _, spell := range item.Spells {
			ids[spell[1]] = true
		}
	}
	for _, set := range sets {
		for _, bonus := range set.Bonuses {
			ids[bonus[1]] = true
		}
	}
	talents := map[int64]tooltip.Talent{}
	for id := range treeTalents {
		ids[id] = true
		talents[int64(id)] = tooltip.Talent{Spell: int64(id), Rank: 1}
	}
	for _, raw := range rawTalents {
		ranks, err := parseIntArrayField(raw.SpellRank, 9)
		if err != nil || len(ranks) == 0 || ranks[0] == 0 {
			continue
		}
		first := int64(ranks[0])
		offered := ids[int32(first)]
		for i, id := range ranks {
			if id == 0 {
				continue
			}
			talent := tooltip.Talent{Rank: i + 1}
			if offered {
				talent.Spell = first
			}
			talents[int64(id)] = talent
		}
	}
	for _, s := range spelldata.All() {
		ids[s.ID] = true
	}

	conditions, err := tooltipConditions(helper, instance, treeTalents)
	if err != nil {
		return err
	}
	provider := tooltip.DBCTooltipDataProvider{DBC: instance, Modifiers: tooltip.NewModifiers(instance, treeTalents), Conditions: conditions}
	headers, err := loadTooltipHeaders(helper, every)
	if err != nil {
		return err
	}
	spells, failed := spellTooltips(instance, provider, headers, ids, treeTalents, talents)
	if err := writeTooltipFile(path, tooltipFile{Spells: spells, Items: items, Sets: sets, Enchants: enchants}, len(failed)); err != nil {
		return err
	}
	named := itemDatabaseSpells(instance, databases...)
	for _, f := range failed {
		if named[f.Spell] {
			fmt.Printf("No tooltip text for %s, which the item database names (TestItemSpellTextsCompile fails on it)\n", f)
		}
	}

	extraIDs := map[int32]bool{}
	for id, row := range instance.Spells {
		if row.NameLang != "" && !ids[int32(id)] {
			extraIDs[int32(id)] = true
		}
	}
	extraSpells, extraFailed := spellTooltips(instance, provider, headers, extraIDs, treeTalents, talents)
	extraItems := map[int32]itemTooltip{}
	for id, item := range every {
		if _, ok := items[id]; !ok {
			extraItems[id] = item
		}
	}
	return writeTooltipFile(strings.TrimSuffix(path, ".json")+"_extra.json", tooltipFile{Spells: extraSpells, Items: extraItems, Sets: map[int32]setTooltip{}, Enchants: map[int32]string{}}, len(extraFailed))
}

// The spells the item database names, which the site's pages show: the spells it lists icons for, its enchants'
// spells, the use, equip and chance on hit spells of its items, consumables, enchants' items and the items it
// lists icons for, and its items' set bonuses.
func itemDatabaseSpells(instance *dbc.DBC, databases ...*proto.UIDatabase) map[int32]bool {
	spells := map[int32]bool{}
	ofItem := func(id int32) {
		for _, spell := range itemSpells(instance, id) {
			spells[spell[1]] = true
		}
	}
	for _, db := range databases {
		for _, data := range db.SpellIcons {
			spells[data.Id] = true
		}
		for _, item := range db.Items {
			ofItem(item.Id)
			for _, bonus := range instance.ItemSetSpells[instance.Items[int(item.Id)].ItemSetId] {
				spells[int32(bonus.SpellID)] = true
			}
		}
		for _, consumable := range db.Consumables {
			ofItem(consumable.Id)
		}
		for _, data := range db.ItemIcons {
			ofItem(data.Id)
		}
		for _, enchant := range db.Enchants {
			ofItem(enchant.ItemId)
			if enchant.SpellId != 0 {
				spells[enchant.SpellId] = true
			}
		}
	}
	return spells
}

// A description or buff text that did not compile, which the tooltip lacks, or keeps only the text around an
// include that did not.
type uncompiledText struct {
	Spell int32
	Name  string
	Aura  bool
	Err   error
}

func (u uncompiledText) String() string {
	text := "description"
	if u.Aura {
		text = "buff text"
	}
	return fmt.Sprintf("%d %s, its %s: %v", u.Spell, u.Name, text, u.Err)
}

func writeTooltipFile(path string, file tooltipFile, failed int) error {
	data, err := json.Marshal(file)
	if err != nil {
		return err
	}
	fmt.Printf("Wrote %d spell, %d item and %d set tooltips to %s (%d KB); %d texts did not compile\n",
		len(file.Spells), len(file.Items), len(file.Sets), path, len(data)/1024, failed)
	return os.WriteFile(path, data, 0o644)
}

// Each spell's tooltip; with no headers, the tooltips go without their header lines.
func spellTooltips(instance *dbc.DBC, provider tooltip.DBCTooltipDataProvider, headers *tooltipHeaders, ids map[int32]bool, treeTalents map[int32]int32, talents map[int64]tooltip.Talent) (map[int32]spellTooltip, []uncompiledText) {
	tooltips := map[int32]spellTooltip{}
	var failed []uncompiledText
	for _, id := range slices.Sorted(maps.Keys(ids)) {
		row, ok := instance.Spells[int(id)]
		if !ok || row.NameLang == "" {
			continue
		}
		t := spellTooltip{Name: row.NameLang, Rank: row.Subtext, Icon: iconName(row.IconPath), Lines: headers.lines(id, provider.Modifiers), MaxPoints: treeTalents[id]}
		t.RankTexts = rankTexts(provider, id, row.Description, treeTalents[id], talents)
		var err error
		if row.Description != "" {
			if t.Text, err = tooltip.Compile(row.Description, provider, int64(id), talents); err != nil {
				failed = append(failed, uncompiledText{Spell: id, Name: row.NameLang, Err: err})
			}
		}
		if row.AuraDescription != "" {
			if t.Aura, err = tooltip.Compile(row.AuraDescription, provider, int64(id), talents); err != nil {
				failed = append(failed, uncompiledText{Spell: id, Name: row.NameLang, Aura: true, Err: err})
			}
		}
		tooltips[id] = t
	}
	return tooltips, failed
}

// every holds the whole tooltips, which give an item's lines from its ItemSparse row; setPieces the order the
// game lists a set's pieces in.
func itemTooltips(instance *dbc.DBC, itemIDs []int32, every map[int32]itemTooltip, setPieces map[int32][]int32) (map[int32]itemTooltip, map[int32]setTooltip) {
	items := map[int32]itemTooltip{}
	sets := map[int32]setTooltip{}
	for _, id := range itemIDs {
		row, ok := instance.Items[int(id)]
		if !ok {
			continue
		}
		items[id] = itemTooltip{Binding: row.Bonding, Flavor: row.Flavor, Spells: itemSpells(instance, id), itemFacts: every[id].itemFacts}

		if row.ItemSetId == 0 {
			continue
		}
		set, ok := sets[int32(row.ItemSetId)]
		if !ok {
			set = setTooltip{Name: row.ItemSetName}
			for _, bonus := range instance.ItemSetSpells[row.ItemSetId] {
				set.Bonuses = append(set.Bonuses, [2]int32{int32(bonus.Threshold), int32(bonus.SpellID)})
			}
			slices.SortFunc(set.Bonuses, func(a, b [2]int32) int { return int(a[0] - b[0]) })
		}
		set.Pieces = append(set.Pieces, id)
		sets[int32(row.ItemSetId)] = set
	}
	for id, set := range sets {
		order := setPieces[id]
		slices.SortFunc(set.Pieces, func(a, b int32) int {
			return cmp.Or(cmp.Compare(pieceIndex(order, a), pieceIndex(order, b)), cmp.Compare(a, b))
		})
		sets[id] = set
	}
	return items, sets
}

// A piece's place in its set's ItemSet row, which the game lists the pieces in; one the row does not name goes last.
func pieceIndex(order []int32, id int32) int {
	if i := slices.Index(order, id); i >= 0 {
		return i
	}
	return len(order)
}

// Each set's pieces in the order of its ItemSet row.
func setPieceOrder(helper *DBHelper) (map[int32][]int32, error) {
	type row struct {
		id     int32
		pieces string
	}
	rows, err := LoadRows(helper.db, `SELECT ID, ItemID FROM ItemSet`, func(rows *sql.Rows) (row, error) {
		var r row
		return r, rows.Scan(&r.id, &r.pieces)
	})
	if err != nil {
		return nil, fmt.Errorf("loading the item sets for tooltips: %w", err)
	}
	order := map[int32][]int32{}
	for _, r := range rows {
		var pieces []int32
		if err := json.Unmarshal([]byte(r.pieces), &pieces); err != nil {
			return nil, fmt.Errorf("item set %d: %w", r.id, err)
		}
		order[r.id] = slices.DeleteFunc(pieces, func(id int32) bool { return id == 0 })
	}
	return order, nil
}

// The name the game writes on an item for each enchant the databases list.
func enchantNames(helper *DBHelper, databases []*proto.UIDatabase) (map[int32]string, error) {
	type row struct {
		id     int32
		name   string
		points string
	}
	rows, err := LoadRows(helper.db, `SELECT ID, COALESCE(Name_lang, ''), COALESCE(EffectPointsMin, '[]') FROM SpellItemEnchantment`, func(rows *sql.Rows) (row, error) {
		var r row
		return r, rows.Scan(&r.id, &r.name, &r.points)
	})
	if err != nil {
		return nil, fmt.Errorf("loading the enchants for tooltips: %w", err)
	}
	listed := map[int32]bool{}
	for _, db := range databases {
		for _, enchant := range db.Enchants {
			listed[enchant.EffectId] = true
		}
	}
	names := map[int32]string{}
	for _, r := range rows {
		if !listed[r.id] || r.name == "" {
			continue
		}
		var points []int
		if err := json.Unmarshal([]byte(r.points), &points); err != nil {
			return nil, fmt.Errorf("reading enchant %d's points: %w", r.id, err)
		}
		if names[r.id], err = enchantName(helper, r.name, points); err != nil {
			return nil, fmt.Errorf("writing enchant %d's name: %w", r.id, err)
		}
	}
	return names, nil
}

// $k1 to $k3 in an enchant's name, or $<spell>s1 and $<spell>M1.
var enchantNameValue = regexp.MustCompile(`\$(?:k(\d)|(\d+)[sSmM](\d))`)

// An enchant's name with its numbers in: Agility +$k1 is Agility +7, $k1 being the enchant's first effect's points.
func enchantName(helper *DBHelper, name string, points []int) (string, error) {
	var err error
	return enchantNameValue.ReplaceAllStringFunc(name, func(match string) string {
		sub := enchantNameValue.FindStringSubmatch(match)
		if sub[1] != "" {
			if i, _ := strconv.Atoi(sub[1]); i >= 1 && i <= len(points) {
				return strconv.Itoa(points[i-1])
			}
			err = fmt.Errorf("%s names no effect of the enchant", match)
			return match
		}
		var value float64
		if scanErr := helper.db.QueryRow(`SELECT EffectBasePointsF FROM SpellEffect WHERE SpellID = ? AND EffectIndex = ? AND DifficultyID = 0`,
			sub[2], sub[3][0]-'1').Scan(&value); scanErr != nil {
			err = fmt.Errorf("%s: %w", match, scanErr)
			return match
		}
		return strconv.FormatFloat(value, 'f', -1, 64)
	}), err
}

const wholeItemsQuery = `
	SELECT i.ID, s.Display_lang, s.OverallQualityID, s.ItemLevel, s.RequiredLevel, s.Bonding, COALESCE(s.Description_lang, ''),
		i.IconFileDataID, i.ClassID, i.SubclassID, COALESCE(isc.DisplayName_lang, ''), COALESCE(s.ContainerSlots, 0), COALESCE(a.Quality, ''),
		s.InventoryType, s.AllowableClass, s.SellPrice, s.MaxCount, s.Flags_0,
		s.RequiredSkillRank, COALESCE(skill.DisplayName_lang, ''), s.MinReputation, COALESCE(f.Name_lang, '')
	FROM Item i
	JOIN ItemSparse s ON s.ID = i.ID
	LEFT JOIN ItemSubClass isc ON isc.ClassID = i.ClassID AND isc.SubClassID = i.SubclassID
	LEFT JOIN ItemDamageAmmo a ON a.ItemLevel = s.ItemLevel
	LEFT JOIN SkillLine skill ON skill.ID = s.RequiredSkill AND s.RequiredSkill != 0
	LEFT JOIN Faction f ON f.ID = s.MinFactionID AND s.MinFactionID != 0`

const (
	itemClassArmor          = 4
	itemClassProjectileAmmo = 6
)

// Every item's tooltip, whole: the parts any item has, and its name, quality, icon, levels, slot and type, with a
// bag's slots and the damage per second an arrow adds, which the game reads off ItemDamageAmmo by the item level and
// quality and rounds down to the half point (gen_ammo.go). The type is the subclass the game writes beside the slot,
// so an item with no slot has none, and neither has Miscellaneous armor, such as a ring or a trinket.
func wholeItemTooltips(helper *DBHelper, instance *dbc.DBC) (map[int32]itemTooltip, error) {
	type row struct {
		id         int32
		item       itemTooltip
		iconID     int
		class      int
		subclass   int
		ammoRows   string
		classMask  int
		maxCount   int32
		flags      int64
		skillRank  int32
		skill      string
		standing   int32
		reputation string
	}
	rows, err := LoadRows(helper.db, wholeItemsQuery, func(rows *sql.Rows) (row, error) {
		var r row
		err := rows.Scan(&r.id, &r.item.Name, &r.item.Quality, &r.item.ItemLevel, &r.item.RequiredLevel, &r.item.Binding, &r.item.Flavor,
			&r.iconID, &r.class, &r.subclass, &r.item.Type, &r.item.Slots, &r.ammoRows,
			&r.item.InventoryType, &r.classMask, &r.item.SellPrice, &r.maxCount, &r.flags,
			&r.skillRank, &r.skill, &r.standing, &r.reputation)
		return r, err
	})
	if err != nil {
		return nil, fmt.Errorf("loading the items for tooltips: %w", err)
	}
	items := map[int32]itemTooltip{}
	for _, r := range rows {
		r.item.Icon = iconName(iconsMap[r.iconID])
		r.item.Spells = itemSpells(instance, r.id)
		if strings.Contains(r.item.Type, "OBSOLETE") || (r.item.InventoryType == 0 && r.item.Slots == 0) || (r.class == itemClassArmor && r.subclass == 0) {
			r.item.Type = ""
		}
		var dps []float64
		if r.class == itemClassProjectileAmmo && json.Unmarshal([]byte(r.ammoRows), &dps) == nil && int(r.item.Quality) < len(dps) {
			r.item.AmmoDPS = math.Floor(dps[r.item.Quality]*2) / 2
		}
		r.item.Classes = dbc.GetClassesFromClassMask(r.classMask)
		switch {
		case r.maxCount > 0:
			r.item.Unique = r.maxCount
		case dbc.ItemStaticFlags0(uint32(r.flags)).Has(dbc.UNIQUE_EQUIPPABLE):
			r.item.Unique = -1
		}
		if r.skill != "" {
			r.item.Skill = &itemRequirement{Name: r.skill, Rank: r.skillRank}
		}
		if r.reputation != "" {
			r.item.Reputation = &itemRequirement{Name: r.reputation, Rank: r.standing}
		}
		items[r.id] = r.item
	}
	return items, nil
}

// An item's use, equip and chance on hit spells. The cooldown is the effect's own, or its category's for one with none
// of its own, as a potion's; the game writes none as short as the 1.5 sec global cooldown (a scroll's 1 sec, an
// elixir's own 1 sec beside its category's 3), nor charges of one, so those are left out.
func itemSpells(instance *dbc.DBC, id int32) (spells [][]int32) {
	for _, effect := range instance.ItemEffectsByParentID[int(id)] {
		if !shownItemTriggers[effect.TriggerType] || effect.SpellID == 0 {
			continue
		}
		cooldown := effect.CoolDownMSec
		if cooldown <= 0 {
			cooldown = effect.CategoryCoolDownMSec
		}
		if cooldown <= 1500 {
			cooldown = 0
		}
		spell := []int32{int32(effect.TriggerType), int32(effect.SpellID), int32(cooldown)}
		if charges := int32(effect.Charges); charges > 1 || charges < -1 {
			spell = append(spell, max(charges, -charges))
		}
		spells = append(spells, spell)
	}
	return spells
}

// The icon name the item database and the site's icon folder use for a listfile path: the file name without its
// extension, lowercased, with its spaces as dashes (GetIconName).
func iconName(path string) string {
	if path == "" {
		return ""
	}
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}

// A talent's description at each rank, read off the store's ladder for it; none when every rank reads alike.
func rankTexts(provider tooltip.DBCTooltipDataProvider, id int32, description string, maxPoints int32, talents map[int64]tooltip.Talent) (texts [][]tooltip.Segment) {
	if maxPoints < 2 || description == "" || spelldata.Find(id) == spelldata.Nil {
		return nil
	}
	// A curve whose length disagrees with the tree's points panics in the ladder; that talent keeps the row's text.
	defer func() {
		if recover() != nil {
			texts = nil
		}
	}()
	ladder := spelldata.Talent(id, maxPoints)
	differ := false
	for n := int32(1); n <= maxPoints; n++ {
		rank := tooltip.RankProvider{DBCTooltipDataProvider: provider, SpellID: int64(id), Rank: ladder.Rank(n)}
		text, err := tooltip.Compile(description, rank, int64(id), talents)
		if err != nil {
			return nil
		}
		differ = differ || (len(texts) > 0 && fmt.Sprint(text) != fmt.Sprint(texts[0]))
		texts = append(texts, text)
	}
	if !differ {
		return nil
	}
	return texts
}

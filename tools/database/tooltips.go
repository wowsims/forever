package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/dbc"
	"github.com/wowsims/forever/tools/tooltip"
)

// The site's tooltips (ui/sim/proto/tooltips.ts): spells by id, and the parts of an item's tooltip the item
// database does not hold, by item id, with its sets by set id.
type tooltipFile struct {
	Spells map[int32]spellTooltip `json:"spells"`
	Items  map[int32]itemTooltip  `json:"items"`
	Sets   map[int32]setTooltip   `json:"sets"`
}

// One spell's tooltip: its name, rank and icon, the cost, range, cast and cooldown lines as [left, right]
// pairs, and its description and buff text compiled by tooltip.Compile.
type spellTooltip struct {
	Name  string            `json:"n"`
	Rank  string            `json:"r,omitempty"`
	Icon  string            `json:"i,omitempty"`
	Lines [][2]string       `json:"h,omitempty"`
	Text  []tooltip.Segment `json:"d,omitempty"`
	Aura  []tooltip.Segment `json:"a,omitempty"`
	// A talent's points and, where its curve gives each rank its own amounts, the description at each rank
	// from rank 1; the description above is the top rank's, as the row states it.
	MaxPoints int32               `json:"m,omitempty"`
	RankTexts [][]tooltip.Segment `json:"x,omitempty"`
}

// An item's binding (1 when picked up, 2 when equipped, 3 when used), its flavour text and its spells as
// [trigger, spell id, cooldown in ms]: trigger 0 is a use, 1 an equip effect and 2 a chance on hit. An item the item
// database does not hold, a consumable or an enchant's item, carries its name, quality, icon and levels too.
type itemTooltip struct {
	Binding       int        `json:"b,omitempty"`
	Flavor        string     `json:"f,omitempty"`
	Spells        [][3]int32 `json:"e,omitempty"`
	Name          string     `json:"n,omitempty"`
	Quality       int32      `json:"q,omitempty"`
	Icon          string     `json:"i,omitempty"`
	ItemLevel     int32      `json:"l,omitempty"`
	RequiredLevel int32      `json:"r,omitempty"`
}

// A set's name, its pieces the database holds, and its bonuses as [pieces worn, spell id].
type setTooltip struct {
	Name    string     `json:"n"`
	Pieces  []int32    `json:"p"`
	Bonuses [][2]int32 `json:"b"`
}

// The item spell triggers a tooltip shows: use, equip and chance on hit.
var shownItemTriggers = map[int]bool{0: true, 1: true, 2: true}

// Writes the tooltip of every spell and item the site can show to path: the spells the sim's store holds, every
// talent in the trees (their spell ids and points), and every item, consumable, enchant and spell the databases
// given name, with the spells of those items and of their sets. A condition on a rank of a talent the trees offer
// is left to the player's points, one on a talent they do not offer is false; the store's rows give the header lines.
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
	items, sets := itemTooltips(instance, itemIDs)
	others, err := otherItemTooltips(helper, instance, otherIDs)
	if err != nil {
		return err
	}
	maps.Copy(items, others)
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

	provider := tooltip.DBCTooltipDataProvider{DBC: instance}
	tooltips := map[int32]spellTooltip{}
	failed := 0
	for _, id := range slices.Sorted(maps.Keys(ids)) {
		row, ok := instance.Spells[int(id)]
		if !ok || row.NameLang == "" {
			continue
		}
		t := spellTooltip{Name: row.NameLang, Rank: row.Subtext, Icon: iconName(row.IconPath), Lines: headerLines(spelldata.Find(id)), MaxPoints: treeTalents[id]}
		t.RankTexts = rankTexts(provider, id, row.Description, treeTalents[id], talents)
		var err error
		if row.Description != "" {
			if t.Text, err = tooltip.Compile(row.Description, provider, int64(id), talents); err != nil {
				failed++
			}
		}
		if row.AuraDescription != "" {
			if t.Aura, err = tooltip.Compile(row.AuraDescription, provider, int64(id), talents); err != nil {
				failed++
			}
		}
		tooltips[id] = t
	}

	data, err := json.Marshal(tooltipFile{Spells: tooltips, Items: items, Sets: sets})
	if err != nil {
		return err
	}
	fmt.Printf("Wrote %d spell, %d item and %d set tooltips to %s (%d KB); %d texts did not compile\n",
		len(tooltips), len(items), len(sets), path, len(data)/1024, failed)
	return os.WriteFile(path, data, 0o644)
}

func itemTooltips(instance *dbc.DBC, itemIDs []int32) (map[int32]itemTooltip, map[int32]setTooltip) {
	items := map[int32]itemTooltip{}
	sets := map[int32]setTooltip{}
	for _, id := range itemIDs {
		row, ok := instance.Items[int(id)]
		if !ok {
			continue
		}
		items[id] = itemTooltip{Binding: row.Bonding, Flavor: row.Flavor, Spells: itemSpells(instance, id)}

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
		slices.Sort(set.Pieces)
		sets[id] = set
	}
	return items, sets
}

const otherItemsQuery = `
	SELECT i.ID, s.Display_lang, s.OverallQualityID, s.ItemLevel, s.RequiredLevel, s.Bonding, COALESCE(s.Description_lang, ''), i.IconFileDataID
	FROM Item i
	JOIN ItemSparse s ON s.ID = i.ID`

// The tooltips of items the item database does not hold, each whole: the parts any item has, and its name, quality,
// icon and levels.
func otherItemTooltips(helper *DBHelper, instance *dbc.DBC, ids []int32) (map[int32]itemTooltip, error) {
	wanted := map[int32]bool{}
	for _, id := range ids {
		wanted[id] = true
	}
	type row struct {
		id     int32
		item   itemTooltip
		iconID int
	}
	rows, err := LoadRows(helper.db, otherItemsQuery, func(rows *sql.Rows) (row, error) {
		var r row
		err := rows.Scan(&r.id, &r.item.Name, &r.item.Quality, &r.item.ItemLevel, &r.item.RequiredLevel, &r.item.Binding, &r.item.Flavor, &r.iconID)
		return r, err
	})
	if err != nil {
		return nil, fmt.Errorf("loading the items for tooltips: %w", err)
	}
	items := map[int32]itemTooltip{}
	for _, r := range rows {
		if !wanted[r.id] {
			continue
		}
		r.item.Icon = iconName(iconsMap[r.iconID])
		r.item.Spells = itemSpells(instance, r.id)
		items[r.id] = r.item
	}
	return items, nil
}

func itemSpells(instance *dbc.DBC, id int32) (spells [][3]int32) {
	for _, effect := range instance.ItemEffectsByParentID[int(id)] {
		if shownItemTriggers[effect.TriggerType] && effect.SpellID != 0 {
			spells = append(spells, [3]int32{int32(effect.TriggerType), int32(effect.SpellID), int32(effect.CoolDownMSec)})
		}
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

// The lines under a spell's name as the game shows them: cost and range, then cast time and cooldown.
func headerLines(s *spelldata.Spell) [][2]string {
	if s == nil || s.IsPassive() {
		return nil
	}
	var cost string
	for i := range s.Powers {
		p := &s.Powers[i]
		switch {
		case p.CostPct != 0:
			cost = strconv.FormatFloat(float64(p.CostPct), 'f', -1, 32) + "% of base mana"
		case p.Cost != 0:
			cost = strconv.FormatFloat(s.PowerCost(p.Type), 'f', -1, 64) + " " + powerLabel(p.Type)
		}
	}
	var reach string
	switch {
	case s.MaxRange == 0:
	case s.MaxRange <= 5:
		reach = "Melee Range"
	case s.MinRange != 0:
		reach = fmt.Sprintf("%s-%s yd range", trim(float64(s.MinRange)), trim(float64(s.MaxRange)))
	default:
		reach = trim(float64(s.MaxRange)) + " yd range"
	}
	cast := "Instant"
	switch {
	case s.IsChanneled():
		cast = "Channeled"
	case s.CastTimeMs > 0:
		cast = trim(float64(s.CastTimeMs)/1000) + " sec cast"
	}
	var cooldown string
	if ms := max(s.CooldownMs, s.CategoryCooldownMs); ms > 0 {
		if ms >= 60000 && ms%60000 == 0 {
			cooldown = strconv.Itoa(int(ms/60000)) + " min cooldown"
		} else {
			cooldown = trim(float64(ms)/1000) + " sec cooldown"
		}
	}
	lines := [][2]string{{cast, cooldown}}
	if cost != "" || reach != "" {
		lines = append([][2]string{{cost, reach}}, lines...)
	}
	return lines
}

func trim(v float64) string {
	return strconv.FormatFloat(math.Round(v*100)/100, 'f', -1, 64)
}

func powerLabel(t dbcenums.PowerType) string {
	switch t {
	case dbcenums.POWER_RAGE:
		return "Rage"
	case dbcenums.POWER_FOCUS:
		return "Focus"
	case dbcenums.POWER_ENERGY:
		return "Energy"
	case dbcenums.POWER_HEALTH:
		return "Health"
	}
	return "Mana"
}

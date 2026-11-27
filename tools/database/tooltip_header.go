package database

import (
	"database/sql"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/tooltip"
)

// A spell's header, the lines above its description, uses the game's own strings from the game files' GlobalStrings
// table, which wowsims.db does not carry. Each line names its string (MANA_COST, SPELL_RANGE, SPELL_CAST_TIME_SEC
// and the rest), or says the game files hold none for it.

// The rows a spell's header reads: every spell's row as the store builds it, the spells the store leaves out
// included, and what the store's row does not carry.
type tooltipHeaders struct {
	tables     *spellTables
	ranges     map[int32]headerRange
	optional   map[int32][]int32    // SpellPower.OptionalCost, in the order of the spell's powers.
	forms      map[int32]string     // SpellShapeshiftForm.Name_lang by form id.
	items      map[int32]string     // An item's name by its id.
	masks      []itemSubclassMask   // ItemSubClassMask, by id.
	kinds      map[[2]int32]string  // ItemSubClass.VerboseName_lang, or DisplayName_lang where it has none, by class and subclass.
	tools      map[int32]spellTools // SpellTotems by spell; none where the database lacks the table.
	categories map[int32]string     // TotemCategory.Name_lang by id.
}

// SpellRange.Flags and the range on a friendly target, RangeMin_1 and RangeMax_1, beside the hostile range the
// store's row holds.
type headerRange struct {
	Flags                    int32
	MinFriendly, MaxFriendly float64
}

const rangeFlagMelee = 1

// The second bit the game files mark a next-swing ability with, beside dbcenums.ATTR_ON_NEXT_SWING: Maul carries
// both, a few creature attacks only this one.
const attrOnNextSwing2 uint32 = 0x400

type itemSubclassMask struct {
	Name  string
	Class int32
	Mask  uint32
}

// The totem categories (RequiredTotemCategoryID) and items (Totem) a spell needs as tools.
type spellTools struct {
	Categories []int32
	Items      []int32
}

type headerQuery struct {
	query string
	scan  func(*sql.Rows) error
}

// The header rows of every spell in the game files database, with the names of the items given. A wowsims.db
// extracted before generator-settings.json listed SpellTotems and TotemCategory goes without them, and its
// tooltips without Tools lines.
func loadTooltipHeaders(helper *DBHelper, items map[int32]itemTooltip) (*tooltipHeaders, error) {
	tables, err := loadSpellTables(helper.db)
	if err != nil {
		return nil, fmt.Errorf("loading the spells for tooltip headers: %w", err)
	}
	h := &tooltipHeaders{
		tables:     tables,
		ranges:     map[int32]headerRange{},
		optional:   map[int32][]int32{},
		forms:      map[int32]string{},
		items:      map[int32]string{},
		kinds:      map[[2]int32]string{},
		tools:      map[int32]spellTools{},
		categories: map[int32]string{},
	}
	for id, item := range items {
		h.items[id] = item.Name
	}
	queries := []headerQuery{
		{`SELECT m.SpellID, r.Flags, COALESCE(r.RangeMin_1, 0), COALESCE(r.RangeMax_1, 0)
			FROM SpellMisc m JOIN SpellRange r ON r.ID = m.RangeIndex WHERE m.DifficultyID = 0`, func(rows *sql.Rows) error {
			var id int32
			var r headerRange
			if err := rows.Scan(&id, &r.Flags, &r.MinFriendly, &r.MaxFriendly); err != nil {
				return err
			}
			h.ranges[id] = r
			return nil
		}},
		// In loadPowers' order, so each entry stands beside the power it belongs to.
		{`SELECT SpellID, COALESCE(OptionalCost, 0) FROM SpellPower ORDER BY SpellID, OrderIndex`, func(rows *sql.Rows) error {
			var id, optional int32
			if err := rows.Scan(&id, &optional); err != nil {
				return err
			}
			h.optional[id] = append(h.optional[id], optional)
			return nil
		}},
		{`SELECT ID, COALESCE(Name_lang, '') FROM SpellShapeshiftForm`, func(rows *sql.Rows) error {
			var id int32
			var name string
			if err := rows.Scan(&id, &name); err != nil {
				return err
			}
			h.forms[id] = name
			return nil
		}},
		{`SELECT COALESCE(Name_lang, ''), ClassID, Mask FROM ItemSubClassMask ORDER BY ID`, func(rows *sql.Rows) error {
			var m itemSubclassMask
			var mask int64
			if err := rows.Scan(&m.Name, &m.Class, &mask); err != nil {
				return err
			}
			m.Mask = uint32(mask)
			h.masks = append(h.masks, m)
			return nil
		}},
		{`SELECT ClassID, SubClassID, COALESCE(NULLIF(VerboseName_lang, ''), DisplayName_lang, '') FROM ItemSubClass`, func(rows *sql.Rows) error {
			var class, subclass int32
			var name string
			if err := rows.Scan(&class, &subclass, &name); err != nil {
				return err
			}
			h.kinds[[2]int32{class, subclass}] = name
			return nil
		}},
	}
	if helper.tableExists("SpellTotems") && helper.tableExists("TotemCategory") {
		queries = append(queries, headerQuery{`SELECT SpellID, RequiredTotemCategoryID_0, RequiredTotemCategoryID_1, Totem_0, Totem_1 FROM SpellTotems`, func(rows *sql.Rows) error {
			var id int32
			var categories, items [2]int32
			if err := rows.Scan(&id, &categories[0], &categories[1], &items[0], &items[1]); err != nil {
				return err
			}
			var tools spellTools
			for i := range 2 {
				if categories[i] != 0 {
					tools.Categories = append(tools.Categories, categories[i])
				}
				if items[i] != 0 {
					tools.Items = append(tools.Items, items[i])
				}
			}
			h.tools[id] = tools
			return nil
		}}, headerQuery{`SELECT ID, COALESCE(Name_lang, '') FROM TotemCategory`, func(rows *sql.Rows) error {
			var id int32
			var name string
			if err := rows.Scan(&id, &name); err != nil {
				return err
			}
			h.categories[id] = name
			return nil
		}})
	} else {
		fmt.Println("wowsims.db has no SpellTotems or TotemCategory table, so the spell tooltips leave out their Tools lines until `make db` extracts them")
	}
	for _, q := range queries {
		if err := eachRow(helper.db, q.query, q.scan); err != nil {
			return nil, fmt.Errorf("loading the tooltip headers: %w", err)
		}
	}
	return h, nil
}

// The header of a spell the game files hold, none for one they do not.
func (h *tooltipHeaders) lines(id int32, mods tooltip.Modifiers) [][2]any {
	if h == nil {
		return nil
	}
	if _, ok := h.tables.Misc[id]; !ok {
		return nil
	}
	return headerLines(storeRowSpell(h.tables.row(id)), h, mods)
}

// The lines under a spell's name as the game shows them: cost and range, a friendly range where it differs, then
// cast time and cooldown, what the spell requires, its tools and its reagents. A passive shows none. A part the
// talents' modifiers reach is the segments the UI works the player's points into.
func headerLines(s *spelldata.Spell, h *tooltipHeaders, mods tooltip.Modifiers) [][2]any {
	if s.IsPassive() {
		return nil
	}
	id := int64(s.ID)
	cost := costSegments(s, h.optional[s.ID], mods.For(id, dbcenums.SPELLMOD_COST))

	r := h.ranges[s.ID]
	rangeMods := mods.For(id, dbcenums.SPELLMOD_RANGE)
	reach := reachSegments(float64(s.MinRange), float64(s.MaxRange), r.Flags, rangeMods)
	var friendly []tooltip.Segment
	if r.MaxFriendly != float64(s.MaxRange) && reach != nil {
		// SPELL_RANGE_DUAL, "%1$s: %2$s yd range", with ENEMY and FRIENDLY; the friendly range goes on the line below.
		reach = append([]tooltip.Segment{"Enemy: "}, reach...)
		friendly = append([]tooltip.Segment{"Friendly: "}, reachSegments(r.MinFriendly, r.MaxFriendly, r.Flags, rangeMods)...)
	}

	var cast []tooltip.Segment
	switch m := mods.For(id, dbcenums.SPELLMOD_CASTING_TIME); {
	case s.IsChanneled():
		cast = []tooltip.Segment{"Channeled"} // SPELL_CAST_CHANNELED
	case s.HasAttr(dbcenums.ATTR_INDEX_BASE, dbcenums.ATTR_ON_NEXT_SWING|attrOnNextSwing2):
		cast = []tooltip.Segment{"Next melee"} // SPELL_ON_NEXT_SWING
	case s.CastTimeMs < 0:
		// SpellCastTimes row 18's -1000000 ms, on Arcane Shot and the stings, has no game string, so no cast line.
	case s.CastTimeMs > 0 && len(m) > 0:
		cast = tooltip.ModifiedCastTime(float64(s.CastTimeMs), m)
	default:
		cast = []tooltip.Segment{tooltip.CastTimeText(float64(s.CastTimeMs))}
	}
	var cooldown []tooltip.Segment
	if ms := max(s.CooldownMs, s.CategoryCooldownMs); ms > 0 {
		cooldown = []tooltip.Segment{tooltip.CooldownText(float64(ms))}
		if m := mods.For(id, dbcenums.SPELLMOD_COOLDOWN); len(m) > 0 {
			cooldown = tooltip.ModifiedCooldown(float64(ms), m)
		}
	}

	var lines [][2]any
	if cost != nil || reach != nil {
		lines = append(lines, [2]any{side(cost), side(reach)})
	}
	if friendly != nil {
		lines = append(lines, [2]any{"", side(friendly)})
	}
	if cast != nil || cooldown != nil {
		lines = append(lines, [2]any{side(cast), side(cooldown)})
	}
	for _, requirement := range h.requirements(s) {
		lines = append(lines, [2]any{requirement, ""})
	}
	if tools := h.toolNames(s.ID); len(tools) > 0 {
		lines = append(lines, [2]any{"Tools: " + strings.Join(tools, ", "), ""}) // SPELL_TOTEMS
	}
	if reagents := h.reagentNames(s); len(reagents) > 0 {
		// SPELL_REAGENTS, "Reagents: |n", puts the reagents on the line below it.
		lines = append(lines, [2]any{"Reagents:", ""}, [2]any{strings.Join(reagents, ", "), ""})
	}
	return lines
}

// One side of a header line: its text where no talent reaches it, else its segments.
func side(segments []tooltip.Segment) any {
	var text strings.Builder
	for _, segment := range segments {
		s, ok := segment.(string)
		if !ok {
			return segments
		}
		text.WriteString(s)
	}
	return text.String()
}

// What each bar the spell spends costs, in the game files' order. The cost modifiers reach the first; a cost
// stated as a share of base mana takes only their percentages. The game files hold no string joining two costs,
// so they are joined with a slash.
func costSegments(s *spelldata.Spell, optional []int32, mods []tooltip.Modifier) []tooltip.Segment {
	var out []tooltip.Segment
	for i, p := range s.Powers {
		var first []tooltip.Modifier
		if i == 0 {
			first = mods
		}
		var extra int32
		if i < len(optional) {
			extra = optional[i]
		}
		part := powerCost(p, extra, first)
		if part == nil {
			continue
		}
		if out != nil {
			out = append(out, " / ")
		}
		out = append(out, part...)
	}
	return out
}

// The names the game's strings give each bar: MANA_COST, RAGE_COST, FOCUS_COST, ENERGY_COST and HEALTH_COST.
var powerNames = map[dbcenums.PowerType]string{
	dbcenums.POWER_MANA:   "Mana",
	dbcenums.POWER_RAGE:   "Rage",
	dbcenums.POWER_FOCUS:  "Focus",
	dbcenums.POWER_ENERGY: "Energy",
	dbcenums.POWER_HEALTH: "Health",
}

// One bar's cost: an amount ("%s Mana"), a share of the caster's base or maximum ("%s%% of Maximum Health",
// HEALTH_COST_PCT; the game files hold no string for a share of mana), and what a channel drains each second
// ("%s Mana, plus %s per sec" or "%s Mana per sec", MANA_COST_PER_TIME and _NO_BASE). Combo points are
// COMBO_POINTS_COST, "%s Combo |4Point:Points;", with the least a finisher spends to the most as "%s to %s".
func powerCost(p spelldata.Power, optional int32, mods []tooltip.Modifier) []tooltip.Segment {
	if p.Type == dbcenums.POWER_COMBO_POINTS {
		if p.Cost == 0 {
			return nil
		}
		amount, points := strconv.Itoa(int(p.Cost)), "Combo Points"
		if optional > 0 {
			amount += " to " + strconv.Itoa(int(p.Cost+optional))
		} else if p.Cost == 1 {
			points = "Combo Point"
		}
		return []tooltip.Segment{amount + " " + points}
	}
	name, ok := powerNames[p.Type]
	if !ok {
		return nil
	}
	scale := 1.0
	if p.Type.InTenths() {
		scale = 10
	}
	var out []tooltip.Segment
	switch {
	case p.CostPct != 0:
		share := "% of base mana"
		if p.Type == dbcenums.POWER_HEALTH {
			share = "% of Maximum Health"
		}
		out = []tooltip.Segment{trim(float64(p.CostPct)) + share}
		if pct := percentOnly(mods); len(pct) > 0 {
			out = tooltip.ModifiedAmount(float64(p.CostPct), pct, share, false)
		}
	case p.Cost != 0:
		cost := float64(p.Cost) / scale
		out = []tooltip.Segment{trim(cost) + " " + name}
		if len(mods) > 0 {
			out = tooltip.ModifiedAmount(cost, mods, " "+name, true)
		}
	}
	if p.PerSecond != 0 {
		perSecond := trim(float64(p.PerSecond) / scale)
		if out == nil {
			return []tooltip.Segment{perSecond + " " + name + " per sec"}
		}
		out = append(out, ", plus "+perSecond+" per sec")
	}
	return out
}

// A cost the game states as a share of base mana takes only the percentage cost modifiers.
func percentOnly(mods []tooltip.Modifier) []tooltip.Modifier {
	var out []tooltip.Modifier
	for _, mod := range mods {
		if !mod.Flat {
			out = append(out, mod)
		}
	}
	return out
}

// A range as the game writes it: MELEE_RANGE on a range row flagged melee, SPELL_RANGE_UNLIMITED from 50000
// yards, and SPELL_RANGE, "%s yd range", otherwise. The game files hold no string for a range with a minimum, so
// it reads "8 - 35 yd range".
func reachSegments(least, most float64, flags int32, mods []tooltip.Modifier) []tooltip.Segment {
	switch {
	case most == 0:
		return nil
	case flags&rangeFlagMelee != 0:
		return []tooltip.Segment{"Melee Range"}
	case most >= 50000:
		return []tooltip.Segment{"Unlimited range"}
	}
	out := []tooltip.Segment{trim(most) + " yd range"}
	if len(mods) > 0 {
		out = tooltip.ModifiedAmount(most, mods, " yd range", false)
	}
	if least != 0 {
		out = append([]tooltip.Segment{trim(least) + " - "}, out...)
	}
	return out
}

// The weapon or shield the spell needs, then the forms it is cast in, as SPELL_EQUIPPED_ITEM and
// SPELL_REQUIRED_FORM write them: "Requires %s". A spell castable out of its forms as well requires none of them.
// An enchant's row names the item it goes on, and a row naming inventory types the slots, so neither gets an item
// line.
func (h *tooltipHeaders) requirements(s *spelldata.Spell) []string {
	var out []string
	enchants := slices.ContainsFunc(s.Effects, func(e spelldata.Effect) bool {
		return e.Type == dbcenums.E_ENCHANT_ITEM || e.Type == dbcenums.E_ENCHANT_ITEM_TEMPORARY
	})
	if s.EquipSubclass != 0 && s.EquipInvType == 0 && !enchants && (s.EquipClass == itemClassWeapon || s.EquipClass == itemClassArmor) {
		if kinds := h.itemKinds(int32(s.EquipClass), uint32(s.EquipSubclass)); len(kinds) > 0 {
			out = append(out, "Requires "+strings.Join(kinds, ", "))
		}
	}
	var forms []string
	for bit := range 64 {
		if s.StanceMask&(1<<bit) != 0 && !s.CastableInCasterForm() && h.forms[int32(bit+1)] != "" {
			forms = append(forms, h.forms[int32(bit+1)])
		}
	}
	if len(forms) > 0 {
		out = append(out, "Requires "+strings.Join(forms, ", "))
	}
	return out
}

const itemClassWeapon = 2

// The item kinds a subclass mask names: a group the game files name whole (ItemSubClassMask's Melee Weapon or
// Ranged Weapon), then each subclass the groups leave, by its plural name, such as Daggers.
func (h *tooltipHeaders) itemKinds(class int32, mask uint32) []string {
	var out []string
	for _, group := range h.masks {
		if group.Class == class && group.Mask != 0 && mask&group.Mask == group.Mask && group.Name != "" {
			out = append(out, group.Name)
			mask &^= group.Mask
		}
	}
	for bit := range 32 {
		if mask&(1<<bit) != 0 && h.kinds[[2]int32{class, int32(bit)}] != "" {
			out = append(out, h.kinds[[2]int32{class, int32(bit)}])
		}
	}
	return out
}

// The totem categories and items the spell needs, as SpellTotems lists them.
func (h *tooltipHeaders) toolNames(id int32) []string {
	var out []string
	tools := h.tools[id]
	for _, category := range tools.Categories {
		if name := h.categories[category]; name != "" {
			out = append(out, name)
		}
	}
	for _, item := range tools.Items {
		if name := h.items[item]; name != "" {
			out = append(out, name)
		}
	}
	return out
}

// Each reagent by name, with the count after it where the cast takes more than one: REAGENT_COST_CONSUME_ITEMS,
// "%s (%d)".
func (h *tooltipHeaders) reagentNames(s *spelldata.Spell) []string {
	var out []string
	for _, reagent := range s.Reagents {
		name := h.items[reagent.Item]
		if name == "" {
			continue
		}
		if reagent.Count > 1 {
			name += " (" + strconv.Itoa(int(reagent.Count)) + ")"
		}
		out = append(out, name)
	}
	return out
}

func trim(v float64) string {
	return strconv.FormatFloat(math.Round(v*100)/100, 'f', -1, 64)
}

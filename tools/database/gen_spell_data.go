package database

import (
	"database/sql"
	"errors"
	"fmt"
	"go/format"
	"io"
	"maps"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/tools/database/dbc"
)

var rankSubtext = regexp.MustCompile(`^Rank (\d+)$`)

// SkillLineAbility.AcquireMethod: 0 trainer, 1 with the skill, 2 on level, 3 granted by another spell.
const (
	acquireOnLevel = 2
	acquireGranted = 3
)

type rankCandidate struct {
	SpellID       int32
	Rank          int32
	ClassMask     int
	SkillLine     int32
	AcquireMethod int32
}

type rankLadder struct {
	Name  string
	Field string
	Ranks map[int32]int32

	// What each rank's effects are worth, keyed by rank and then by the game files' EffectIndex. Set
	// only on a ladder the talent tree supplied, where every rank is the same spell and the numbers
	// sit on the tree's curves rather than on a spell per rank. Nil means the rank's own spell states
	// its numbers, which is what every "Rank N" family does.
	Points map[int32]map[int32]float64
}

// One talent as the tree states it: a single spell, the rank cap the game enforces, and the value
// each rank gives. The rank spells the legacy Talent table still lists are gone from the game files -
// Anticipation's 12750-12753 have no row in SpellName, SpellEffect or SpellMisc - so the tree is the
// only place a talent's ranks are described.
type traitLadder struct {
	SpellID  int32
	MaxRanks int32
	Points   map[int32]map[int32]float64
}

// SkillLineAbility.ClassMask is a bitmask over the class index dbc.Classes already carries, so the bit
// is that index shifted rather than a second table to keep in step with it.
func classMaskOf(class dbc.DbcClass) int {
	return 1 << (class.ID - 1)
}

// The Go identifier a family is reached by: "Shadow Word: Pain" -> ShadowWordPain. A "(DND)" suffix is
// the game files' do-not-display marker, not part of the name.
func fieldNameOf(spellName string) string {
	spellName = strings.TrimSuffix(spellName, " (DND)")
	var b strings.Builder
	upper := true
	for _, r := range spellName {
		switch {
		case r == '\'' || r == '\u2019':
			// Dropped without breaking the word, so "Avenger's Shield" is AvengersShield rather than
			// AvengerSShield.
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if upper {
				b.WriteRune(unicode.ToUpper(r))
				upper = false
			} else {
				b.WriteRune(r)
			}
		default:
			upper = true
		}
	}

	name := b.String()
	if name == "" || unicode.IsDigit(rune(name[0])) {
		return ""
	}
	return name
}

// The spells that belong to a Season of Discovery rune and to nothing a Forever character can learn.
// A spell is one when an "Engrave <slot> - <rune>" spell reaches it and no SkillLineAbility row
// outside the rune lines 2851 and 2853 teaches it on a trainer (AcquireMethod 0) or on level (2):
// Shadow Cleave 403835 is reached through Engrave Gloves - Metamorphosis 403938, its enchant's
// Metamorphosis 403789 and that aura's action bar override, and its only rows are grants. A rune
// ability a trainer also teaches - Victory Rush 402927, Incinerate 412758 - is Forever's, and so is
// anything a talent reaches, which keeps Primal Bite 407995.
//
// The spell an override on a rune ability replaces is the rune's own when it does nothing itself: a
// dummy effect and no description. Metamorphosis : Demon Charge 412783 is the one such spell with a
// level row, on Demonology with no class bit, and it only stands on the action bar for Metamorphosis
// to swap into Demon Charge 412788.
func loadRuneOnlySpells(db *sql.DB) (map[int32]bool, error) {
	fromRune, err := grantClosure(db, `SELECT ID FROM SpellName WHERE Name_lang LIKE 'Engrave % - %'`)
	if err != nil {
		return nil, err
	}
	fromTalent, err := grantClosure(db, `SELECT SpellID FROM TraitDefinition WHERE SpellID > 0`)
	if err != nil {
		return nil, err
	}
	trained, err := LoadRows(db, `
		SELECT DISTINCT Spell FROM SkillLineAbility
		WHERE AcquireMethod IN (0, ?) AND SkillLine NOT IN (2851, 2853)`,
		scanSpellID, acquireOnLevel)
	if err != nil {
		return nil, err
	}

	runeOnly := map[int32]bool{}
	for id := range fromRune {
		runeOnly[id] = true
	}
	for _, id := range trained {
		delete(runeOnly, id)
	}

	placeholders, err := LoadRows(db, fmt.Sprintf(`
		SELECT o.SpellID, s.ID
		FROM SpellEffect o
		JOIN Spell s ON s.ID = o.EffectMiscValue_0
		WHERE o.EffectAura = %d AND COALESCE(s.Description_lang, '') = ''
		AND NOT EXISTS (SELECT 1 FROM SpellEffect e WHERE e.SpellID = s.ID AND e.Effect != %d)`,
		dbcenums.A_OVERRIDE_ACTIONBAR_SPELLS, dbcenums.E_DUMMY),
		func(r *sql.Rows) ([2]int32, error) {
			var pair [2]int32
			return pair, r.Scan(&pair[0], &pair[1])
		})
	if err != nil {
		return nil, err
	}
	for _, pair := range placeholders {
		if fromRune[pair[0]] {
			runeOnly[pair[1]] = true
		}
	}

	for id := range fromTalent {
		delete(runeOnly, id)
	}
	return runeOnly, nil
}

// The seed spells and every spell they reach within six steps: an item enchant (Effect 53 or 54)
// reaches the spells its SpellItemEnchantment names, an action bar override the ability it puts on the
// bar, and a trigger its target.
func grantClosure(db *sql.DB, seedQuery string) (map[int32]bool, error) {
	ids, err := LoadRows(db, fmt.Sprintf(`
		WITH RECURSIVE
		seed(sid) AS (%s),
		ench(eid, arg) AS (
			SELECT ID, EffectArg_0 FROM SpellItemEnchantment WHERE EffectArg_0 > 0
			UNION SELECT ID, EffectArg_1 FROM SpellItemEnchantment WHERE EffectArg_1 > 0
			UNION SELECT ID, EffectArg_2 FROM SpellItemEnchantment WHERE EffectArg_2 > 0),
		reach(sid, depth) AS (
			SELECT sid, 0 FROM seed
			UNION
			SELECT ench.arg, r.depth + 1 FROM reach r
			JOIN SpellEffect se ON se.SpellID = r.sid AND se.Effect IN (%d, %d)
			JOIN ench ON ench.eid = se.EffectMiscValue_0
			WHERE r.depth < 6
			UNION
			SELECT CAST(se.EffectBasePointsF AS INTEGER), r.depth + 1 FROM reach r
			JOIN SpellEffect se ON se.SpellID = r.sid AND se.EffectAura = %d AND se.EffectBasePointsF > 0
			WHERE r.depth < 6
			UNION
			SELECT se.EffectTriggerSpell, r.depth + 1 FROM reach r
			JOIN SpellEffect se ON se.SpellID = r.sid AND se.EffectTriggerSpell > 0
			WHERE r.depth < 6)
		SELECT DISTINCT sid FROM reach`,
		seedQuery, dbcenums.E_ENCHANT_ITEM, dbcenums.E_ENCHANT_ITEM_TEMPORARY, dbcenums.A_OVERRIDE_ACTIONBAR_SPELLS),
		scanSpellID)
	if err != nil {
		return nil, err
	}
	set := make(map[int32]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set, nil
}

func scanSpellID(r *sql.Rows) (int32, error) {
	var id int32
	return id, r.Scan(&id)
}

// A family is a rune's when every spell it could resolve to is.
func runeOnlyFamily(byRank map[int32][]rankCandidate, runeOnly map[int32]bool) bool {
	if len(byRank) == 0 {
		return false
	}
	for _, cands := range byRank {
		for _, c := range cands {
			if !runeOnly[c.SpellID] {
				return false
			}
		}
	}
	return true
}

// Every family a class can learn, from two sources: every spell in one of the class's skill lines
// whose subtext reads "Rank N", "Shapeshift", "Summon" or is empty (a single-rank ability like
// Whirlwind, Cat Form or Summon Imp is its own rank 1), and every talent in the class's tree. Nothing
// hand-maintained. A family every spell of which is in runeOnly is a Season of Discovery rune ability
// and is left out.
func discoverLadders(db *sql.DB, class dbc.DbcClass, treeID int, runeOnly map[int32]bool) ([]rankLadder, []string, []string, error) {
	mask := classMaskOf(class)

	exclusive, err := exclusiveSkillLines(db, mask)
	if err != nil {
		return nil, nil, nil, err
	}
	scaling := petScalingSpells[dbc.ClassNameFromDBC(class)]
	petLines, err := petSkillLines(db, scaling)
	if err != nil {
		return nil, nil, nil, err
	}
	owned := maps.Clone(exclusive)
	maps.Copy(owned, petLines)

	// The Defense skill line carries the one-class passives that open a counterattack window - Offensive
	// State (DND) fires the 5 s Overpower aura 1282733 on a melee hit, Defensive State (DND) the Revenge
	// one - beside Parry and Block, which have several classes' bits.
	rows, err := db.Query(`
		SELECT n.Name_lang, sla.Spell, s.NameSubtext_lang, sla.ClassMask, sla.SkillLine, sla.AcquireMethod,
		       COALESCE(lv.BaseLevel, 0), (COALESCE(json_extract(sm.Attributes, '$[0]'), 0) & ?) != 0
		FROM SkillLineAbility sla
		JOIN SpellName n ON n.ID = sla.Spell
		JOIN Spell s ON s.ID = sla.Spell
		LEFT JOIN SpellLevels lv ON lv.SpellID = sla.Spell AND lv.DifficultyID = 0
		LEFT JOIN SpellMisc sm ON sm.SpellID = sla.Spell AND sm.DifficultyID = 0
		WHERE (sla.SkillLine IN (
			SELECT DISTINCT sla2.SkillLine
			FROM SkillLineAbility sla2
			JOIN SkillLine sl2 ON sl2.ID = sla2.SkillLine AND sl2.CategoryID = 7
			WHERE (sla2.ClassMask & ?) != 0
		) OR (sla.SkillLine = ? AND sla.ClassMask = ? AND sla.AcquireMethod = ?)
		OR sla.SkillLine IN (`+petSkillLinesQuery+`))
		AND (s.NameSubtext_lang LIKE 'Rank %' OR s.NameSubtext_lang IN ('', 'Shapeshift', 'Summon'))
		AND sla.SkillLine NOT IN (2851, 2853)
		AND NOT EXISTS (SELECT 1 FROM SpellEffect se WHERE se.SpellID = sla.Spell AND se.EffectAura = ?)
		ORDER BY n.Name_lang, sla.Spell`, dbcenums.ATTR_PASSIVE, mask, dbc.SkillLineDefense, mask, acquireOnLevel, scaling, dbcenums.A_MOUNTED)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	treeSpells, err := treeSpellIDs(db, treeID)
	if err != nil {
		return nil, nil, nil, err
	}
	legacyTalents, err := legacyTalentNames(db)
	if err != nil {
		return nil, nil, nil, err
	}

	byName := map[string]map[int32][]rankCandidate{}
	// A spell with no subtext is a single-rank family of its own only where no "Rank N" row carries
	// its name - Execute's granted 20647 is a sub-spell of the ranked Execute, not a rank - and only
	// when a trainer teaches it (Death Wish, Sweeping Strikes and the stance passives all have a
	// trainer row) or a talent the tree could not describe shares its name. A grant-only one is a
	// Season of Discovery rune ability or an internal (Quick Strike, Meathook). Skill lines 2851 and
	// 2853 are the runes themselves, and the Mounted aura keeps the paladin Mounts line out, which
	// is a class line too.
	unranked := map[string][]rankCandidate{}
	trainedUnranked := map[string]bool{}
	for rows.Next() {
		var name, subtext string
		var c rankCandidate
		var baseLevel int32
		var passive bool
		if err := rows.Scan(&name, &c.SpellID, &subtext, &c.ClassMask, &c.SkillLine, &c.AcquireMethod, &baseLevel, &passive); err != nil {
			return nil, nil, nil, err
		}
		m := rankSubtext.FindStringSubmatch(subtext)
		if m == nil {
			c.Rank = 1
			unranked[name] = append(unranked[name], c)
			// A passive the legacy Talent table names, with a trainer row at level 1 and no place in
			// the Forever tree, is a talent that no longer exists (Emberstorm, Deadliness); Anger
			// Management has the same rows but sits in the tree.
			legacy := passive && baseLevel <= 1 && legacyTalents[name] && !treeSpells[c.SpellID]
			if c.AcquireMethod != acquireGranted && !legacy {
				trainedUnranked[name] = true
			}
			continue
		}
		rank, _ := strconv.Atoi(m[1])
		c.Rank = int32(rank)
		if byName[name] == nil {
			byName[name] = map[int32][]rankCandidate{}
		}
		byName[name][c.Rank] = append(byName[name][c.Rank], c)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, nil, err
	}

	imbues, err := discoverImbueFamilies(db, exclusive)
	if err != nil {
		return nil, nil, nil, err
	}
	for name, byRank := range imbues {
		if byName[name] != nil {
			return nil, nil, nil, fmt.Errorf("%s: both a class ability and a crafted imbue carry the name", name)
		}
		byName[name] = byRank
	}

	traits, traitSkipped, traitPartial, err := discoverTraitLadders(db, treeID)
	if err != nil {
		return nil, nil, nil, err
	}
	for name, cands := range unranked {
		if byName[name] != nil {
			continue
		}
		if _, talent := traits[name]; talent {
			continue // the tree describes it; its own spell and legacy ranks would only collide
		}
		_, talentSkipped := traitSkipped[name]
		if trainedUnranked[name] || talentSkipped {
			byName[name] = map[int32][]rankCandidate{1: cands}
		}
	}

	seen := map[string]bool{}
	var names []string
	for _, source := range []map[string]bool{keysOf(byName), keysOf(traits), keysOf(traitSkipped)} {
		for name := range source {
			if !seen[name] {
				seen[name] = true
				names = append(names, name)
			}
		}
	}
	sort.Strings(names)

	var ladders []rankLadder
	var skipped, partial []string
	taken := map[string]string{}

	for _, name := range names {
		_, talent := traits[name]
		_, talentSkipped := traitSkipped[name]
		if !talent && !talentSkipped && runeOnlyFamily(byName[name], runeOnly) {
			skipped = append(skipped, fmt.Sprintf("%s: Season of Discovery rune ability", name))
			continue
		}
		field := fieldNameOf(name)
		if field == "" {
			skipped = append(skipped, fmt.Sprintf("%s: no usable Go identifier", name))
			continue
		}
		if owner, clash := taken[field]; clash {
			skipped = append(skipped, fmt.Sprintf("%s: identifier %s already taken by %s", name, field, owner))
			continue
		}

		var ladder map[int32]int32
		var points map[int32]map[int32]float64
		var resolveErr error
		if claimedByClass(byName[name], mask, owned) {
			ladder, resolveErr = resolveLadder(db, name, byName[name], mask)
		}

		// The tree wins where it states more ranks than the "Rank N" spells resolved to, because the
		// rank cap it carries is the one the talent proto and the UI trees are built from: Precision
		// kept its subtext on rank 1 alone, and leaving it as that one rank would let a 3-point
		// talent be read past its own ladder. Below that the "Rank N" spells win - they state a cost
		// and a cast time per rank, which the tree does not - and a family the resolver refused
		// stays refused, since a tree node that grants an ability does not describe its ranks. At an
		// equal count the tree wins too when no "Rank N" spell is in it: those are a legacy talent's
		// ranks (Shield Specialization 20148 and 20149 beside the tree's 20150), which state
		// none of what the tree's spell does.
		if t, ok := traits[name]; ok && resolveErr == nil &&
			(int(t.MaxRanks) > len(ladder) || int(t.MaxRanks) == len(ladder) && !anyInTree(ladder, treeSpells)) {
			ladder, points = map[int32]int32{}, t.Points
			for rank := int32(1); rank <= t.MaxRanks; rank++ {
				ladder[rank] = t.SpellID
			}
			if note := traitPartial[name]; note != "" {
				partial = append(partial, fmt.Sprintf("%s: %s", name, note))
			}
		}

		if resolveErr != nil {
			skipped = append(skipped, fmt.Sprintf("%s: %s", name, resolveErr))
			continue
		}
		if len(ladder) == 0 {
			if reason := traitSkipped[name]; reason != "" {
				skipped = append(skipped, fmt.Sprintf("%s: %s", name, reason))
			}
			continue
		}

		taken[field] = name
		ladders = append(ladders, rankLadder{Name: name, Field: field, Ranks: ladder, Points: points})
	}

	return ladders, skipped, partial, nil
}

// Whether any rank of the ladder is a spell the talent tree defines.
func anyInTree(ladder map[int32]int32, treeSpells map[int32]bool) bool {
	for _, id := range ladder {
		if treeSpells[id] {
			return true
		}
	}
	return false
}

// A trailing roman numeral on a crafted imbue's name: "Deadly Poison IV", "Occult Poison I".
var imbueRankNumeral = regexp.MustCompile(` (?:I|II|III|IV|V|VI|VII|VIII|IX|X)$`)

// The weapon imbues a class crafts, by family and rank, each rank the spell its enchant casts off the
// weapon's hits. Poisons sit in skill line 40, outside the class category, and its rows are recipes:
// Instant Poison VI 11343 creates item 8928, whose use spell puts enchant 625 on the weapon, and the
// enchant's combat spell slot casts 11337 - the spell the sim registers. The use spell's effect is
// E_360 on the trained poisons and E_ENCHANT_ITEM_TEMPORARY on the newer ones. The numeral in the
// recipe's name is its rank, so the family is the name without it; a recipe with no rank subtext is
// rank 1. Only lines no other class appears in are read, and a recipe whose chain ends in no combat
// spell - Blinding Powder is a reagent - is no imbue.
func discoverImbueFamilies(db *sql.DB, exclusive map[int32]bool) (map[string]map[int32][]rankCandidate, error) {
	families := map[string]map[int32][]rankCandidate{}
	for line := range exclusive {
		rows, err := db.Query(enchantSlotsCTE+fmt.Sprintf(`
			SELECT DISTINCT n.Name_lang, s.NameSubtext_lang, sla.ClassMask, sla.AcquireMethod, slot.SpellID
			FROM SkillLineAbility sla
			JOIN SpellName n ON n.ID = sla.Spell
			JOIN Spell s ON s.ID = sla.Spell
			JOIN SpellEffect recipe ON recipe.SpellID = sla.Spell AND recipe.Effect = %d
			JOIN ItemXItemEffect x ON x.ItemID = recipe.EffectItemType
			JOIN ItemEffect ie ON ie.ID = x.ItemEffectID
			JOIN SpellEffect imbue ON imbue.SpellID = ie.SpellID AND imbue.Effect IN (%d, %d)
			JOIN slots slot ON slot.ID = imbue.EffectMiscValue_0 AND slot.Effect = %d AND slot.SpellID > 0
			WHERE sla.SkillLine = ?`,
			dbcenums.E_CREATE_ITEM, dbcenums.E_ENCHANT_ITEM_TEMPORARY, dbcenums.E_360, dbc.ITEM_ENCHANTMENT_COMBAT_SPELL), line)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var name, subtext string
			c := rankCandidate{SkillLine: line, Rank: 1}
			if err := rows.Scan(&name, &subtext, &c.ClassMask, &c.AcquireMethod, &c.SpellID); err != nil {
				rows.Close()
				return nil, err
			}
			if m := rankSubtext.FindStringSubmatch(subtext); m != nil {
				rank, _ := strconv.Atoi(m[1])
				c.Rank = int32(rank)
			}
			family := imbueRankNumeral.ReplaceAllString(name, "")
			if families[family] == nil {
				families[family] = map[int32][]rankCandidate{}
			}
			families[family][c.Rank] = append(families[family][c.Rank], c)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}
	return families, nil
}

// A $<spellID><token> in a description: "$12880d" on Enrage, "$12976s1" on Last Stand.
var descriptionSpellRef = regexp.MustCompile(`\$(?:/\d+;)?(\d{4,7})[a-z]`)

// The spells a rank triggers or reads its tooltip from: an EffectTriggerSpell edge (Intercept's
// stun 20615, Intimidating Shout's fear 20511), a $<id> token (Enrage's buff 12880, Flurry's
// 12966, Last Stand's 12976) or an overrides.HandTriggers entry, which drops the $<id> it replaces. In id
// order, without the rank itself and without ids the game files have no spell for.
func triggeredSpells(db *sql.DB, spellID int32) ([]int32, error) {
	seen := map[int32]bool{}
	for _, id := range handTriggered(spellID) {
		seen[id] = true
	}
	rows, err := db.Query(`SELECT EffectTriggerSpell FROM SpellEffect WHERE SpellID = ? AND EffectTriggerSpell > 0`, spellID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int32
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		seen[id] = true
	}
	rows.Close()

	var desc string
	if err := scanOptional(db, `SELECT COALESCE(Description_lang, '') FROM Spell WHERE ID = ?`, spellID, &desc); err != nil {
		return nil, err
	}
	for _, m := range descriptionSpellRef.FindAllStringSubmatch(desc, -1) {
		id, _ := strconv.Atoi(m[1])
		seen[int32(id)] = true
	}
	for _, id := range handReplaced(spellID) {
		delete(seen, id)
	}
	delete(seen, spellID)

	var ids []int32
	for id := range seen {
		var name string
		if err := scanOptional(db, `SELECT Name_lang FROM SpellName WHERE ID = ?`, id, &name); err != nil {
			return nil, err
		}
		if name != "" {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids, nil
}

// One spell of a family's triggered ladder, and the rank the ladder files it under.
type triggeredRow struct {
	Rank    int32
	SpellID int32
}

// The spells a family's ranks trigger, numbered for its triggered ladder. Where every rank triggers
// the same spell there is one row, rank 1; where each rank triggers its own, the row takes the
// rank's number; anything else is numbered in id order.
func triggeredRows(db *sql.DB, l rankLadder) ([]triggeredRow, error) {
	ranks := make([]int32, 0, len(l.Ranks))
	for rank := range l.Ranks {
		ranks = append(ranks, rank)
	}
	sort.Slice(ranks, func(i, j int) bool { return ranks[i] < ranks[j] })

	perRank := map[int32][]int32{}
	distinct := map[int32]bool{}
	var order []int32
	for _, rank := range ranks {
		ids, err := triggeredSpells(db, l.Ranks[rank])
		if err != nil {
			return nil, err
		}
		perRank[rank] = ids
		for _, id := range ids {
			if !distinct[id] {
				distinct[id] = true
				order = append(order, id)
			}
		}
	}
	if len(order) == 0 {
		return nil, nil
	}

	numbered := map[int32]int32{}
	oneEach := true
	for _, rank := range ranks {
		if len(perRank[rank]) != 1 {
			oneEach = false
			break
		}
		numbered[perRank[rank][0]] = rank
	}
	switch {
	case len(order) == 1:
		numbered[order[0]] = 1
	case oneEach && len(numbered) == len(order):
	default:
		sort.Slice(order, func(i, j int) bool { return order[i] < order[j] })
		for i, id := range order {
			numbered[id] = int32(i + 1)
		}
	}

	out := make([]triggeredRow, 0, len(order))
	for _, id := range order {
		out = append(out, triggeredRow{Rank: numbered[id], SpellID: id})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Rank < out[j].Rank })
	return out, nil
}

// Every spell a node of the class's Forever tree defines.
func treeSpellIDs(db *sql.DB, treeID int) (map[int32]bool, error) {
	unparked, err := unparkedTraitNodes(db)
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(`
		SELECT DISTINCT COALESCE(d.SpellID, 0)
		FROM TraitNode tn
		JOIN TraitNodeXTraitNodeEntry x ON x.TraitNodeID = tn.ID
		JOIN TraitNodeEntry e ON e.ID = x.TraitNodeEntryID
		JOIN TraitDefinition d ON d.ID = e.TraitDefinitionID
		WHERE tn.TraitTreeID = ? AND `+unparked, treeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int32]bool{}
	for rows.Next() {
		var id int32
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		if id != 0 {
			out[id] = true
		}
	}
	return out, rows.Err()
}

// The names the game files' legacy Talent table still carries, tree or no tree.
func legacyTalentNames(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query(`SELECT DISTINCT n.Name_lang FROM Talent t JOIN SpellName n ON n.ID = t.SpellRank_0`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out[name] = true
	}
	return out, rows.Err()
}

func keysOf[V any](m map[string]V) map[string]bool {
	keys := make(map[string]bool, len(m))
	for k := range m {
		keys[k] = true
	}
	return keys
}

// Every talent in one class's tree, by name, plus the ones that had to be left out and the ones that
// came out short an effect. The tree states one spell per talent and a curve per effect, where the
// curve reads rank -> value, so a rank's numbers are the curve's and never the spell's own: those
// hold the top rank on some talents and a stale number on others, and 240 of the 640 curves in this
// build disagree with them.
func discoverTraitLadders(db *sql.DB, treeID int) (map[string]traitLadder, map[string]string, map[string]string, error) {
	unparked, err := unparkedTraitNodes(db)
	if err != nil {
		return nil, nil, nil, err
	}
	rows, err := db.Query(`
		SELECT DISTINCT d.ID, COALESCE(d.SpellID, 0), e.MaxRanks,
		       COALESCE(NULLIF(sn.Name_lang, ''), d.OverrideName_lang, '')
		FROM TraitNode tn
		JOIN TraitNodeXTraitNodeEntry x ON x.TraitNodeID = tn.ID
		JOIN TraitNodeEntry e ON e.ID = x.TraitNodeEntryID
		JOIN TraitDefinition d ON d.ID = e.TraitDefinitionID
		LEFT JOIN SpellName sn ON sn.ID = d.SpellID
		WHERE tn.TraitTreeID = ? AND `+unparked+`
		ORDER BY d.ID`, treeID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	type traitDef struct {
		ID       int32
		SpellID  int32
		MaxRanks int32
		Name     string
	}
	var defs []traitDef
	for rows.Next() {
		var d traitDef
		if err := rows.Scan(&d.ID, &d.SpellID, &d.MaxRanks, &d.Name); err != nil {
			return nil, nil, nil, err
		}
		defs = append(defs, d)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, nil, err
	}

	ladders := map[string]traitLadder{}
	skipped := map[string]string{}
	partial := map[string]string{}
	spellOf := map[string]int32{}

	for _, d := range defs {
		if d.Name == "" || d.SpellID == 0 {
			continue
		}
		// A node appears once per entry, and the game files keep retired twins around, so a second
		// definition on the same spell is the same talent and a second one on another spell is two
		// talents the generated file could only reach by one name.
		if first, ok := spellOf[d.Name]; ok {
			if first != d.SpellID {
				delete(ladders, d.Name)
				skipped[d.Name] = fmt.Sprintf(
					"two talent tree nodes carry the name, on spells %d and %d", first, d.SpellID)
			}
			continue
		}
		spellOf[d.Name] = d.SpellID

		// A one-rank node the game teaches grants an ability rather than describing a ladder. Hemorrhage
		// is a node on a spell the trainer lists, and the one row it would yield names the node's
		// spell, which is not the spell the ability's own ranks are keyed on. A one-rank node nothing
		// teaches is the talent itself, whether or not the spell is passive: Raging Blows, Water
		// Shield, and Berserk, whose spell is active and is not on a skill line. It yields its one
		// row at the spell's base points.
		oneRankUntaught := false
		if d.MaxRanks <= 1 {
			taught, err := taughtBySkillLine(db, d.SpellID)
			if err != nil {
				return nil, nil, nil, err
			}
			if taught {
				continue
			}
			oneRankUntaught = true
		}

		effects, err := effectIndicesOf(db, d.SpellID)
		if err != nil {
			return nil, nil, nil, err
		}
		if len(effects) == 0 {
			skipped[d.Name] = fmt.Sprintf("the game files state no effects for spell %d", d.SpellID)
			continue
		}

		curves, err := traitCurves(db, d.ID)
		if err != nil {
			return nil, nil, nil, err
		}

		points := map[int32]map[int32]float64{}
		var uncovered []string
		for _, index := range effects {
			values, ok := curveRanks(curves[index], d.MaxRanks)
			if !ok {
				uncovered = append(uncovered, strconv.Itoa(int(index)))
				continue
			}
			for rank, value := range values {
				if points[rank] == nil {
					points[rank] = map[int32]float64{}
				}
				points[rank][index] = value
			}
		}

		if len(points) == 0 {
			if !oneRankUntaught {
				skipped[d.Name] = fmt.Sprintf("the talent tree states no per-rank value for spell %d", d.SpellID)
				continue
			}
			points[1] = map[int32]float64{}
		}
		if len(uncovered) > 0 {
			partial[d.Name] = fmt.Sprintf("effect %s of spell %d has no rank curve and is held at its base points",
				strings.Join(uncovered, ", "), d.SpellID)
		}
		ladders[d.Name] = traitLadder{SpellID: d.SpellID, MaxRanks: d.MaxRanks, Points: points}
	}

	return ladders, skipped, partial, nil
}

// Whether a trainer, the skill itself or a level grants the spell (AcquireMethod 0, 1 or 2).
func taughtBySkillLine(db *sql.DB, spellID int32) (bool, error) {
	var taught bool
	err := scanOptional(db, fmt.Sprintf(`SELECT COUNT(*) > 0 FROM SkillLineAbility WHERE Spell = ? AND AcquireMethod != %d`, acquireGranted), spellID, &taught)
	return taught, err
}

func effectIndicesOf(db *sql.DB, spellID int32) ([]int32, error) {
	rows, err := db.Query(`SELECT EffectIndex FROM SpellEffect WHERE SpellID = ? ORDER BY EffectIndex`, spellID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indices []int32
	for rows.Next() {
		var index int32
		if err := rows.Scan(&index); err != nil {
			return nil, err
		}
		indices = append(indices, index)
	}
	return indices, rows.Err()
}

// Each curve as rank -> value. OperationType is 0 on all 640 rows in this build, meaning the point is
// the value the rank is set to; anything else would modify the spell's own number instead, and is
// dropped rather than guessed at - the effect then reads as having no curve.
func traitCurves(db *sql.DB, defID int32) (map[int32]map[int32]float64, error) {
	rows, err := db.Query(`
		SELECT p.EffectIndex, p.OperationType, cp.Pos_0, cp.Pos_1
		FROM TraitDefinitionEffectPoints p
		JOIN CurvePoint cp ON cp.CurveID = p.CurveID
		WHERE p.TraitDefinitionID = ?
		ORDER BY p.EffectIndex, cp.Pos_0`, defID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	curves := map[int32]map[int32]float64{}
	for rows.Next() {
		var index, operation int32
		var at, value float64
		if err := rows.Scan(&index, &operation, &at, &value); err != nil {
			return nil, err
		}
		rank := int32(at)
		if operation != 0 || float64(rank) != at {
			continue
		}
		if curves[index] == nil {
			curves[index] = map[int32]float64{}
		}
		curves[index][rank] = value
	}
	return curves, rows.Err()
}

func curveRanks(curve map[int32]float64, maxRanks int32) (map[int32]float64, bool) {
	values := make(map[int32]float64, maxRanks)
	for rank := int32(1); rank <= maxRanks; rank++ {
		value, ok := curve[rank]
		if !ok {
			return nil, false
		}
		values[rank] = value
	}
	return values, true
}

// The pet scaling spell that marks a class's own pet skill lines. The warlock's demons learn their
// abilities on those lines alone, with no class bit and no teacher row on a class line. The hunter's
// pet abilities are taught on Beast Training and the families are read by discoverPetFamilies, so
// the hunter has no entry. A line with no trained ability belongs to a pet no trainer or grimoire
// teaches: the Infernal and Doomguard carry only passives, and the Felguard is a Season of Discovery
// rune whose one leveled passive, Avoidance 427743, would otherwise reach the class file.
var petScalingSpells = map[string]string{"Warlock": "Warlock Pet Scaling"}

const petSkillLinesQuery = `
	SELECT h.SkillLine FROM SkillLineAbility h
	JOIN SpellName hn ON hn.ID = h.Spell
	JOIN SkillLine hl ON hl.ID = h.SkillLine AND hl.CategoryID = 7 AND hl.DisplayName_lang LIKE 'Pet - %'
	WHERE hn.Name_lang = ?
	AND EXISTS (SELECT 1 FROM SkillLineAbility t WHERE t.SkillLine = h.SkillLine AND t.AcquireMethod = 0)`

// The pet skill lines that carry the scaling spell; none for an empty name.
func petSkillLines(db *sql.DB, scaling string) (map[int32]bool, error) {
	lines := map[int32]bool{}
	if scaling == "" {
		return lines, nil
	}
	rows, err := db.Query(petSkillLinesQuery, scaling)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var line int32
		if err := rows.Scan(&line); err != nil {
			return nil, err
		}
		lines[line] = true
	}
	return lines, rows.Err()
}

// The skill lines only this class appears in. 11 are shared: "Holy" carries both the paladin and the
// priest bit, which is how paladin Holy Shock reached the priest file. Inside an exclusive line a
// ClassMask-0 spell must be this class's, which is what makes Ignite attributable.
func exclusiveSkillLines(db *sql.DB, mask int) (map[int32]bool, error) {
	rows, err := db.Query(`
		SELECT SkillLine, group_concat(DISTINCT ClassMask)
		FROM SkillLineAbility WHERE ClassMask != 0 GROUP BY SkillLine`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := map[int32]bool{}
	for rows.Next() {
		var line int32
		var masks string
		if err := rows.Scan(&line, &masks); err != nil {
			return nil, err
		}
		union := 0
		for _, part := range strings.Split(masks, ",") {
			m, err := strconv.Atoi(part)
			if err != nil {
				continue
			}
			union |= m
		}
		if union == mask {
			lines[line] = true
		}
	}
	return lines, rows.Err()
}

// A family is this class's if one of its ranks carries the class bit, or if its ranks sit in a skill
// line no other class appears in.
func claimedByClass(byRank map[int32][]rankCandidate, mask int, exclusive map[int32]bool) bool {
	for _, cands := range byRank {
		for _, c := range cands {
			if c.ClassMask&mask != 0 || exclusive[c.SkillLine] {
				return true
			}
		}
	}
	return false
}

// Picks one spell per rank. Lightning Bolt's Elemental Overload twins (45284-45293) share the name,
// skill line and class set, differing only by ClassMask 0. So the class bit wins where it exists, and
// a ClassMask-0 candidate is taken only when no real one was found - that is how Holy Shield 1-3
// resolve.
// Where no rule can separate two candidates and none should try: Seal of Righteousness rank 1 is
// 20154 and 21084, same name, rank, effect shape and 20 mana. The sim has always used 21084.
var ladderPins = map[string]map[int32]int32{
	"Seal of Righteousness": {1: 21084},
}

// Forever re-issues a classic ability as a second spell in the 4xxxxx range, same name, same rank
// subtext, same class: Exorcism rank 1 is both 879 and 415068, and Raptor Strike rank 1 is 2973,
// 409691 and 415335. The copies are granted (AcquireMethod 3) where the original is trained (0) or
// learned (2), so the lowest method is the ability the player actually casts. Narrowing is relative
// and never empties the set, which leaves the 44 accepted ladder spells that are only ever granted
// exactly where they were.
func lowestAcquireMethod(cands []rankCandidate) []rankCandidate {
	perSpell := map[int32]int32{}
	for _, c := range cands {
		if best, seen := perSpell[c.SpellID]; !seen || c.AcquireMethod < best {
			perSpell[c.SpellID] = c.AcquireMethod
		}
	}

	best := int32(math.MaxInt32)
	for _, method := range perSpell {
		if method < best {
			best = method
		}
	}

	var kept []rankCandidate
	for _, c := range cands {
		if perSpell[c.SpellID] == best {
			kept = append(kept, c)
		}
	}
	return kept
}

func resolveLadder(db *sql.DB, name string, byRank map[int32][]rankCandidate, mask int) (map[int32]int32, error) {
	ladder := map[int32]int32{}

	// Sorted, so that when several ranks are ambiguous the one named in the skipped list is always the
	// lowest rather than whichever the map handed over first.
	rankNums := make([]int32, 0, len(byRank))
	for rank := range byRank {
		rankNums = append(rankNums, rank)
	}
	sort.Slice(rankNums, func(i, j int) bool { return rankNums[i] < rankNums[j] })

	var ambiguous []string
	for _, rank := range rankNums {
		cands := byRank[rank]
		var chosen []rankCandidate
		for _, c := range cands {
			if c.ClassMask&mask != 0 {
				chosen = append(chosen, c)
			}
		}
		if len(chosen) == 0 {
			chosen = cands
		}
		chosen = lowestAcquireMethod(chosen)

		ids := map[int32]bool{}
		for _, c := range chosen {
			ids[c.SpellID] = true
		}

		if len(ids) == 1 {
			ladder[rank] = chosen[0].SpellID
			continue
		}

		// Exorcism is one name over two full ladders: the trainer ranks 879-10314, and 415068-415073,
		// which the Season of Discovery passive Exorcist (415076) swaps onto the action bar so the
		// spell can hit any target. The stand-in copies the rank's name, subtext, class bit and mana,
		// so nothing above separates them, but the override aura names both sides. The overridden
		// spell is the one a trainer teaches and the one the ladder wants; the stand-in is only ever
		// reachable through its owner. Fire Blast, Drain Life, Renew and Raptor Strike carry the same
		// shape under Overheat, Master Channeler, Empowered Renew and Melee Specialist.
		replaced, err := overrideReplacements(db, ids)
		if err != nil {
			return nil, err
		}
		if len(replaced) > 0 && len(replaced) < len(ids) {
			for id := range replaced {
				delete(ids, id)
			}
			if len(ids) == 1 {
				for id := range ids {
					ladder[rank] = id
				}
				continue
			}
		}

		// Holy Shock is one name over three spells per rank: a dummy the player casts, plus a damage
		// and a heal spell never exposed in game. Only the castable one carries a SpellPower row,
		// and it is the one the sim registers, so that is the tie-break.
		castable, err := castableOf(db, ids)
		if err != nil {
			return nil, err
		}
		if castable == 0 && len(chosen) < len(cands) {
			// Holy Shock's ranks 2-4 put the class bit on the damage and heal spells and none on the
			// castable one, so the class filter above keeps only the two never exposed. The castable
			// one is still the rank's when it is the only one of all the rank's spells.
			all := map[int32]bool{}
			for _, c := range lowestAcquireMethod(cands) {
				all[c.SpellID] = true
			}
			castable, err = castableOf(db, all)
			if err != nil {
				return nil, err
			}
		}
		if castable == 0 {
			// Judgement of Command is the same dispatcher shape with no mana on either half - the
			// parent Judgement pays - so SpellPower cannot separate them. One E_DUMMY against one
			// E_SCHOOL_DAMAGE can: the dummy is what the seal triggers and what the sim registers,
			// and the damage spell rides along as a sibling, exactly as Holy Shock does.
			castable, err = dispatcherOf(db, ids)
			if err != nil {
				return nil, err
			}
		}
		if castable == 0 {
			// Natural Armor is one name over two full ladders of the hunter's, both on class skill
			// lines: the rank the trainer teaches the hunter, which is an E_LEARN_SPELL aimed at
			// the pet, and the passive the pet then has. The teaching rank is the one the ladder
			// wants, and the passive is its triggered spell, as every other Beast Training row
			// already reads; the passive is only ever reachable through its teacher.
			castable, err = teacherOf(db, ids)
			if err != nil {
				return nil, err
			}
		}
		if pinned, ok := ladderPins[name][rank]; ok && ids[pinned] {
			ladder[rank] = pinned
			continue
		}
		if castable == 0 {
			// Every ambiguous rank is collected rather than returning on the first, so the skipped
			// list in the class file names all of them. Judgement of Command is ambiguous on all six.
			var list []string
			for id := range ids {
				list = append(list, strconv.Itoa(int(id)))
			}
			sort.Strings(list)
			ambiguous = append(ambiguous, fmt.Sprintf("rank %d between spells %s", rank, strings.Join(list, ", ")))
			continue
		}
		ladder[rank] = castable
	}
	if len(ambiguous) > 0 {
		return nil, fmt.Errorf("ambiguous %s", strings.Join(ambiguous, "; "))
	}

	maxRank := int32(0)
	for rank := range ladder {
		if rank > maxRank {
			maxRank = rank
		}
	}
	for rank := int32(1); rank <= maxRank; rank++ {
		if _, ok := ladder[rank]; !ok {
			return nil, fmt.Errorf("missing rank %d of %d", rank, maxRank)
		}
	}
	return ladder, nil
}

// The one spell among these whose first effect is a dummy, when exactly one other is a direct
// damage effect. That pairing is the game files' dispatcher shape: the dummy is the spell the game
// grants and triggers, the damage spell is never exposed. Anything else returns 0 rather than
// guessing - two talent auras that merely differ somewhere are not a dispatcher.
func dispatcherOf(db *sql.DB, ids map[int32]bool) (int32, error) {
	var dummy, damage int32
	for id := range ids {
		var effect dbc.SpellEffectType
		err := db.QueryRow(
			`SELECT Effect FROM SpellEffect WHERE SpellID = ? ORDER BY EffectIndex LIMIT 1`, id).Scan(&effect)
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		if err != nil {
			return 0, err
		}
		switch effect {
		case dbcenums.E_DUMMY:
			if dummy != 0 {
				return 0, nil
			}
			dummy = id
		case dbcenums.E_SCHOOL_DAMAGE:
			if damage != 0 {
				return 0, nil
			}
			damage = id
		default:
			return 0, nil
		}
	}
	if dummy == 0 || damage == 0 {
		return 0, nil
	}
	return dummy, nil
}

// The one spell among these whose trigger edges reach every other, or 0 when none does.
func teacherOf(db *sql.DB, ids map[int32]bool) (int32, error) {
	var found int32
	for id := range ids {
		taught, err := triggeredSpells(db, id)
		if err != nil {
			return 0, err
		}
		reaches := map[int32]bool{}
		for _, t := range taught {
			reaches[t] = true
		}
		all := true
		for other := range ids {
			if other != id && !reaches[other] {
				all = false
				break
			}
		}
		if !all {
			continue
		}
		if found != 0 {
			return 0, nil
		}
		found = id
	}
	return found, nil
}

// The one spell among these that has a mana cost, or 0 when that does not single one out.
func castableOf(db *sql.DB, ids map[int32]bool) (int32, error) {
	var found int32
	for id := range ids {
		var n int
		if err := db.QueryRow(`SELECT count(*) FROM SpellPower WHERE SpellID = ?`, id).Scan(&n); err != nil {
			return 0, err
		}
		if n == 0 {
			continue
		}
		if found != 0 {
			return 0, nil
		}
		found = id
	}
	return found, nil
}

// The candidates that another candidate is swapped for by an A_OVERRIDE_ACTIONBAR_SPELLS aura. On
// that aura the misc value is the spell being overridden and the base points is its replacement; a
// candidate is dropped only when the spell it stands in for is itself in the running, so a stand-in
// whose base rank was never a candidate is left alone.
func overrideReplacements(db *sql.DB, ids map[int32]bool) (map[int32]bool, error) {
	replaced := map[int32]bool{}
	for id := range ids {
		rows, err := db.Query(`
			SELECT CAST(EffectBasePointsF AS INTEGER)
			FROM SpellEffect
			WHERE EffectAura = ? AND EffectMiscValue_0 = ?`, int(dbcenums.A_OVERRIDE_ACTIONBAR_SPELLS), id)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var replacement int32
			if err := rows.Scan(&replacement); err != nil {
				rows.Close()
				return nil, err
			}
			if ids[replacement] {
				replaced[replacement] = true
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}
	return replaced, nil
}

// Every file the generator writes, by the path it is written to, rendered and none written: what
// happens to them is writeSpellDataFiles' business, and -check's business is that nothing does.
//
// A spell data group only does not take keeps the committed inputs' rows, and a class it does not take
// keeps its committed class file, which is therefore not among the files returned. committed is nil
// where no inputs are committed yet.
func renderSpellDataFiles(helper *DBHelper, only SpellDataGroups, committed *storeInputs) (map[string][]byte, *storeInputs, error) {
	if err := RequireSpellCastTimes(helper); err != nil {
		return nil, nil, err
	}
	version, err := buildVersion(helper)
	if err != nil {
		return nil, nil, err
	}
	if committed == nil && !only.all() {
		return nil, nil, fmt.Errorf("%s is missing, so there is nothing to keep the groups not taken from", spellStoreInputsPath)
	}
	push, err := hotfixPush(helper)
	if err != nil {
		return nil, nil, err
	}
	if committed != nil {
		if later := groupsOnALaterBuild(committed, only, version); len(later) > 0 {
			return nil, nil, fmt.Errorf("the game files database is %s, and %s already record a later build: "+
				"extract the later build, or leave those groups out", version, strings.Join(later, ", "))
		}
		if err := checkHotfixes(committed.Builds, committed.Hotfixes, only, version, push); err != nil {
			return nil, nil, err
		}
	}

	// Rendered in full before anything is written, so a class that fails validation cannot leave half
	// the packages regenerated and half stale.
	namer := newRankEnumNamer()

	// One tree per class, picked the same way the talent protos pick it, so the rank caps here and the
	// ones the sim's Talents message carries are the same numbers.
	trees, err := selectTraitTrees(helper)
	if err != nil {
		return nil, nil, err
	}

	rendered := map[string][]byte{}
	extra := map[string][]byte{}

	// The store carries every spell the class files are built from, so the discovery runs here and
	// both consumers read the one result rather than each rediscovering the ladders.
	classRoots := map[string][]int32{}
	runeOnly, err := loadRuneOnlySpells(helper.db)
	if err != nil {
		return nil, nil, err
	}
	for _, class := range dbc.Classes {
		pkg := classGroup(class)
		out, files, roots, err := discoverClass(helper.db, class, pkg, trees, runeOnly)
		if err != nil {
			// A class the run does not take keeps its committed rows and files. When the new build
			// breaks its discovery, the class's own port fixes that.
			if only.Takes(pkg) {
				return nil, nil, err
			}
			fmt.Fprintf(progress, "spelldata: %s is not taken and keeps its committed rows: %v\n", pkg, err)
			continue
		}
		classRoots[pkg] = roots
		if only.Takes(pkg) {
			rendered[pkg] = out
			maps.Copy(extra, files)
		}
	}

	// The game files rows the store is built from are captured on the way through and handed back, so
	// the caller can write them next to the store and the regeneration can be checked without the
	// database - see spelldata_inputs.go.
	inputs, err := loadStoreInputs(helper.db, classRoots, trees, version, push)
	if err != nil {
		return nil, nil, err
	}
	if !only.all() {
		var notes []string
		inputs, notes, err = mergeStoreInputs(committed, inputs, only, version)
		if err != nil {
			return nil, nil, err
		}
		for _, note := range notes {
			fmt.Fprintf(progress, "spelldata: %s\n", note)
		}
	}
	if committed != nil && progress != io.Discard {
		for _, line := range changeReport(committed, inputs, only, version) {
			fmt.Fprintf(progress, "spelldata: %s\n", line)
		}
	}
	store, err := renderStore(inputs, namer)
	if err != nil {
		return nil, nil, err
	}

	formsFile, err := renderFormsFile(inputs.Forms)
	if err != nil {
		return nil, nil, err
	}

	// The buffs read the store's rows, so they are rendered from the same inputs, and checked and
	// type-checked with the rest.
	files, err := renderBuffOutputs(inputs)
	if err != nil {
		return nil, nil, fmt.Errorf("buffs: %w", err)
	}
	files["sim/core/dbcenums/forms_auto_gen.go"] = formsFile
	files["sim/core/spelldata/spells_auto_gen.go"] = store
	for pkg, out := range rendered {
		files[fmt.Sprintf("sim/%s/spell_data_auto_gen.go", pkg)] = out
	}
	for path, out := range extra {
		files[path] = out
	}
	return files, inputs, nil
}

// One class's file, the hunter's extra files by path, and the roots the class gives the store.
func discoverClass(db *sql.DB, class dbc.DbcClass, pkg string, trees map[int]int,
	runeOnly map[int32]bool) (out []byte, files map[string][]byte, roots []int32, err error) {
	tree, ok := trees[classMaskOf(class)]
	if !ok {
		return nil, nil, nil, fmt.Errorf("%s: the game files hold no talent tree for this class", pkg)
	}
	ladders, skipped, partial, err := discoverLadders(db, class, tree, runeOnly)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s: %w", pkg, err)
	}
	triggered, err := triggeredByField(db, ladders)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s: %w", pkg, err)
	}
	out, err = renderClassFile(pkg, ladders, triggered, skipped, partial)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s: %w", pkg, err)
	}
	files = map[string][]byte{}

	// The hunter's pet families read the class's ladders, so they are discovered here, and their
	// passives are roots the ladders do not reach.
	if pkg == "hunter" {
		families, familiesSkipped, err := discoverPetFamilies(db, ladders, triggered)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("%s pet families: %w", pkg, err)
		}
		petFile, err := renderPetFamiliesFile(families, familiesSkipped)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("%s pet families: %w", pkg, err)
		}
		files[fmt.Sprintf("sim/%s/pet_families_auto_gen.go", pkg)] = petFile
		roots = append(roots, petFamilyRoots(families)...)

		arrows, arrowsSkipped, err := discoverArrows(db)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("%s arrows: %w", pkg, err)
		}
		quivers, quiversSkipped, err := discoverQuivers(db)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("%s quivers: %w", pkg, err)
		}
		ammoFile, err := renderAmmoFile(arrows, quivers, append(arrowsSkipped, quiversSkipped...))
		if err != nil {
			return nil, nil, nil, fmt.Errorf("%s ammo: %w", pkg, err)
		}
		files[fmt.Sprintf("sim/%s/ammo_auto_gen.go", pkg)] = ammoFile
	}

	for _, l := range ladders {
		for _, id := range l.Ranks {
			roots = append(roots, id)
		}
	}

	// A node the ladders do not describe still grants a spell the sim registers - Hemorrhage is
	// one - so every spell the tree defines is a root of the store.
	treeSpells, err := treeSpellIDs(db, tree)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s: %w", pkg, err)
	}
	for id := range treeSpells {
		if id != 0 {
			roots = append(roots, id)
		}
	}
	return out, files, roots, nil
}

// The triggered table of every family that has one, by the family's field.
func triggeredByField(db *sql.DB, ladders []rankLadder) (map[string][]triggeredRow, error) {
	triggered := map[string][]triggeredRow{}
	for _, l := range ladders {
		rows, err := triggeredRows(db, l)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", l.Name, err)
		}
		if len(rows) > 0 {
			triggered[l.Field] = rows
		}
	}
	return triggered, nil
}

func renderClassFile(pkg string, ladders []rankLadder, triggered map[string][]triggeredRow, skipped, partial []string) ([]byte, error) {
	// Named rather than dropped silently, so a family the resolver could not make sense of is visible
	// here instead of merely absent.
	var notGenerated strings.Builder
	for _, block := range []struct {
		header string
		lines  []string
	}{
		{"// Not generated:\n", skipped},
		{"// Generated with an effect held at its base points, the talent tree stating no rank value for it:\n", partial},
	} {
		if len(block.lines) == 0 {
			continue
		}
		notGenerated.WriteString(block.header)
		for _, s := range block.lines {
			fmt.Fprintf(&notGenerated, "//   %s\n", s)
		}
		notGenerated.WriteString("\n")
	}

	return renderLadderClassFile(pkg, ladders, triggered, notGenerated.String())
}

// A class file is references into the store: one ladder per family, and no data of its own. Every
// number comes out of sim/core/spelldata.
//
// A family the talent tree prices is one spell whose per-rank numbers live in a curve, which is what
// Talent reads; every other family is one spell per rank, which is Ranked. Both index by position,
// so a ladder whose ranks are not 1..n would silently misnumber and is refused instead.
func renderLadderClassFile(pkg string, ladders []rankLadder, triggered map[string][]triggeredRow, notGenerated string) ([]byte, error) {
	var b strings.Builder
	b.WriteString("type generatedSpellData struct {\n")
	for _, l := range ladders {
		fmt.Fprintf(&b, "\t%s spelldata.Ladder\n", l.Field)
		if triggered[l.Field] != nil {
			fmt.Fprintf(&b, "\t%sTriggered spelldata.Ladder\n", l.Field)
		}
	}
	b.WriteString("}\n\nvar spellData = generatedSpellData{\n")

	for _, l := range ladders {
		ranks := make([]int32, 0, len(l.Ranks))
		for rank := range l.Ranks {
			ranks = append(ranks, rank)
		}
		sort.Slice(ranks, func(i, j int) bool { return ranks[i] < ranks[j] })

		ids := make([]int32, 0, len(ranks))
		for i, rank := range ranks {
			if rank != int32(i+1) {
				return nil, fmt.Errorf("%s: rank %d where rank %d was expected, and a ladder indexes by position",
					l.Name, rank, i+1)
			}
			ids = append(ids, l.Ranks[rank])
		}

		if l.Points != nil {
			fmt.Fprintf(&b, "\t%s: spelldata.Talent(%d, %d),\n", l.Field, ids[0], len(ids))
		} else {
			fmt.Fprintf(&b, "\t%s: spelldata.Ranked(%s),\n", l.Field, joinIDs(ids))
		}

		rows := triggered[l.Field]
		if rows == nil {
			continue
		}
		triggeredIDs := make([]int32, 0, len(rows))
		for i, row := range rows {
			if row.Rank != int32(i+1) {
				return nil, fmt.Errorf("%s triggered: rank %d where rank %d was expected, and a ladder indexes by position",
					l.Name, row.Rank, i+1)
			}
			triggeredIDs = append(triggeredIDs, row.SpellID)
		}
		fmt.Fprintf(&b, "\t%sTriggered: spelldata.Ranked(%s),\n", l.Field, joinIDs(triggeredIDs))
	}
	b.WriteString("}\n")

	var head strings.Builder
	fmt.Fprintf(&head, "// Code generated by tools/database/gen_spelldata. DO NOT EDIT.\n\n")
	fmt.Fprintf(&head, "package %s\n\n", pkg)
	head.WriteString("import (\n\t\"github.com/wowsims/forever/sim/core/spelldata\"\n)\n\n")

	out, err := format.Source([]byte(head.String() + notGenerated + b.String()))
	if err != nil {
		return nil, fmt.Errorf("generated %s file does not parse, refusing to write it: %w", pkg, err)
	}
	return out, nil
}

func joinIDs(ids []int32) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.Itoa(int(id))
	}
	return strings.Join(parts, ", ")
}

func num(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

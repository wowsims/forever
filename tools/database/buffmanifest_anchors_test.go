package database

// The manifest pins the spell each buff reads, so generation needs no game files database. What pinned
// it is the rank resolution below, which reads SkillLineAbility, SkillLine and the rune enchantments -
// tables the store does not capture - so this is where a new game build is checked against the
// pins, and against the names and owners the generator reads off them. Skips without
// tools/database/wowsims.db.

import (
	"database/sql"
	"math"
	"os"
	"regexp"
	"slices"
	"strconv"
	"testing"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/dbc"
)

func openBuffTestDB(t *testing.T) *DBHelper {
	t.Helper()

	DatabasePath = "wowsims.db"
	if _, err := os.Stat(DatabasePath); err != nil {
		t.Skipf("no game files database at %s - run `make db` from a local WoW install to enable this gate", DatabasePath)
	}

	helper, err := NewDBHelper()
	if err != nil {
		t.Fatalf("opening %s: %v", DatabasePath, err)
	}
	t.Cleanup(func() { helper.Close() })
	return helper
}

// The rows of a class family whose spell no skill line grants, which are pinned by hand, each with
// what grants it instead.
var unanchoredBuffs = map[string]string{
	"dark_inspiration": "the shadow talent 1326159 triggers it, and the trait tree's node grants the talent",
}

// The rows that pin the trainer's top rank, each with the rank above it that only a class book teaches,
// which is the one the game files resolve.
var trainerRankBuffs = map[string]int32{
	"battle_shout":            25289, // Manual of Battle Shout VII
	"strength_of_earth_totem": 25362, // Tablet of Strength of Earth Totem V
	"grace_of_air_totem":      25360, // Tablet of Grace of Air Totem III
}

func TestManifestAnchorsMatchTheGameFiles(t *testing.T) {
	db := openBuffTestDB(t).db

	runes, err := loadRuneGrantedSpells(db)
	if err != nil {
		t.Fatalf("%v", err)
	}

	fields := map[string]bool{}
	for _, row := range resolveCommittedBuffs(t) {
		fields[row.Field] = true
		// A row without a castable family - an item's aura, an elixir's debuff - is pinned by hand.
		if row.Name == "" {
			continue
		}
		if _, unanchored := unanchoredBuffs[row.Field]; unanchored {
			if runes[row.SpellID] {
				t.Errorf("%s: spell %d is granted by a rune, which is a class rune and not a raid buff", row.Field, row.SpellID)
			}
			continue
		}

		// A totem or dummy whose cast the manifest does not name reads the aura's name, which no
		// skill line grants, so it fails here.
		classMask := ownerClassMask(row.Owner)
		cands, err := anchorCandidates(db, row.Name, classMask)
		if err != nil {
			t.Fatalf("%s: %v", row.Field, err)
		}
		if len(cands) == 0 {
			t.Errorf("%s: no SkillLineAbility row grants %q to %s", row.Field, row.Name, row.Owner)
			continue
		}

		cast := topRank(cands, classMask)
		aura := cast.SpellID
		if row.AuraName != "" {
			if aura, err = auraFamilyMember(db, row.AuraName, cast.Subtext); err != nil {
				t.Fatalf("%s: %v", row.Field, err)
			}
		}

		resolved := row.SpellID
		book, trainerRank := trainerRankBuffs[row.Field]
		if trainerRank {
			resolved = book
		}
		if resolved != aura {
			t.Errorf("%s: the manifest pins spell %d, which stands for %d; the game files resolve %d", row.Field, row.SpellID, resolved, aura)
		}
		if !trainerRank && row.CastID != 0 && row.CastID != cast.SpellID {
			t.Errorf("%s: the manifest pins cast %d, the game files resolve %d", row.Field, row.CastID, cast.SpellID)
		}
		if runes[aura] {
			t.Errorf("%s: spell %d is granted by a rune, which is a class rune and not a raid buff", row.Field, aura)
		}
	}
	for field := range unanchoredBuffs {
		if !fields[field] {
			t.Errorf("unanchoredBuffs names %q, which is no manifest field", field)
		}
	}
	for field := range trainerRankBuffs {
		if !fields[field] {
			t.Errorf("trainerRankBuffs names %q, which is no manifest field", field)
		}
	}
}

// The two skill lines that grant runes rather than class abilities. Both are CategoryID 7, so the
// anchor query has to name them.
const (
	skillLineEngraving = 2851
	skillLineRunes     = 2853
)

var buffRankSubtext = regexp.MustCompile(`^Rank (\d+)$`)

type buffCandidate struct {
	SpellID       int32
	Subtext       string
	Rank          int32
	ClassMask     int32
	AcquireMethod int32
	Supercedes    int32
}

func queryEach(db *sql.DB, scan func(*sql.Rows) error, query string, args ...any) error {
	rows, err := db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := scan(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}

// The spells a rune grants. An `Engrave <slot> - <name>` spell enchants the item with a
// SpellItemEnchantment whose EffectArg columns hold the spells the rune teaches.
func loadRuneGrantedSpells(db *sql.DB) (map[int32]bool, error) {
	granted := map[int32]bool{}
	err := queryEach(db, func(rows *sql.Rows) error {
		var a, b, c int32
		if err := rows.Scan(&a, &b, &c); err != nil {
			return err
		}
		for _, id := range []int32{a, b, c} {
			if id != 0 {
				granted[id] = true
			}
		}
		return nil
	}, `
		SELECT COALESCE(sie.EffectArg_0, 0), COALESCE(sie.EffectArg_1, 0), COALESCE(sie.EffectArg_2, 0)
		FROM SpellEffect se
		JOIN SpellName n ON n.ID = se.SpellID
		JOIN SpellItemEnchantment sie ON sie.ID = se.EffectMiscValue_0
		WHERE se.Effect = 54 AND n.Name_lang LIKE 'Engrave % - %'`)
	return granted, err
}

// Every spell of the named family the owner can learn. CategoryID 7 is the class and profession
// skill lines.
func anchorCandidates(db *sql.DB, name string, classMask int32) ([]buffCandidate, error) {
	var out []buffCandidate
	err := queryEach(db, func(rows *sql.Rows) error {
		var c buffCandidate
		if err := rows.Scan(&c.SpellID, &c.Subtext, &c.ClassMask, &c.AcquireMethod, &c.Supercedes); err != nil {
			return err
		}
		if m := buffRankSubtext.FindStringSubmatch(c.Subtext); m != nil {
			rank, _ := strconv.Atoi(m[1])
			c.Rank = int32(rank)
		}
		out = append(out, c)
		return nil
	}, `
		SELECT sla.Spell, COALESCE(s.NameSubtext_lang, ''), sla.ClassMask,
		       sla.AcquireMethod, COALESCE(sla.SupercedesSpell, 0)
		FROM SkillLineAbility sla
		JOIN SpellName n ON n.ID = sla.Spell
		JOIN Spell s ON s.ID = sla.Spell
		JOIN SkillLine sl ON sl.ID = sla.SkillLine AND sl.CategoryID = 7
		  AND sl.ID NOT IN (?, ?)
		WHERE n.Name_lang = ? AND ((sla.ClassMask & ?) != 0 OR sla.ClassMask = 0)
		ORDER BY sla.Spell`, skillLineEngraving, skillLineRunes, name, classMask)
	return out, err
}

// The rank the player ends up with. A family with rank subtexts is ordered by them, each rank
// narrowed to its lowest acquire method, since Forever re-issues single ranks as granted copies of a
// trained ability. A family without one - Blessing of Kings, Innervate - is ordered by
// SupercedesSpell, where the spell no other row supersedes is the last one learnt.
func topRank(cands []buffCandidate, classMask int32) buffCandidate {
	var owned []buffCandidate
	for _, c := range cands {
		if classMask != 0 && c.ClassMask&classMask != 0 {
			owned = append(owned, c)
		}
	}
	if len(owned) == 0 {
		owned = cands
	}

	top := int32(0)
	for _, c := range owned {
		top = max(top, c.Rank)
	}
	if top > 0 {
		var ranked []buffCandidate
		best := int32(math.MaxInt32)
		for _, c := range owned {
			if c.Rank == top {
				ranked = append(ranked, c)
				best = min(best, c.AcquireMethod)
			}
		}
		return ranked[slices.IndexFunc(ranked, func(c buffCandidate) bool { return c.AcquireMethod == best })]
	}

	superseded := map[int32]bool{}
	for _, c := range owned {
		superseded[c.Supercedes] = true
	}
	for _, c := range owned {
		if !superseded[c.SpellID] {
			return c
		}
	}
	return owned[0]
}

// The aura of a totem or of a dummy passive, which the game files only tie to the cast by name. The
// rank subtext of the cast picks the matching rank; where neither carries one - Leader of the Pack
// is 17007 and 24932, both rankless - the spell that applies a party or raid aura is the one other
// players see, unless it only procs the buff, the way Windfury Totem's 10612 procs 10610.
func auraFamilyMember(db *sql.DB, name string, subtext string) (int32, error) {
	type auraCandidate struct {
		SpellID int32
		Subtext string
		Shared  bool
	}
	seen := map[int32]int{}
	var cands []auraCandidate
	err := queryEach(db, func(rows *sql.Rows) error {
		var id, target int32
		var sub string
		var effect dbc.SpellEffectType
		if err := rows.Scan(&id, &sub, &effect, &target); err != nil {
			return err
		}
		shared := effect != dbcenums.E_APPLY_AURA || isSharedTarget(dbc.ImplicitTarget(target))
		if index, ok := seen[id]; ok {
			cands[index].Shared = cands[index].Shared || shared
			return nil
		}
		seen[id] = len(cands)
		cands = append(cands, auraCandidate{SpellID: id, Subtext: sub, Shared: shared})
		return nil
	}, `
		SELECT n.ID, COALESCE(s.NameSubtext_lang, ''), e.Effect, COALESCE(e.ImplicitTarget_0, 0)
		FROM SpellName n
		JOIN Spell s ON s.ID = n.ID
		JOIN SpellEffect e ON e.SpellID = n.ID
		WHERE n.Name_lang = ? AND e.Effect IN (?, ?, ?)
		ORDER BY n.ID, e.EffectIndex`,
		name, dbcenums.E_APPLY_AURA, dbcenums.E_APPLY_AREA_AURA_PARTY, dbcenums.E_APPLY_AREA_AURA_RAID)
	if err != nil || len(cands) == 0 {
		return 0, err
	}

	var matching []auraCandidate
	for _, c := range cands {
		if c.Subtext == subtext {
			matching = append(matching, c)
		}
	}
	if len(matching) == 0 {
		matching = cands
	}
	if len(matching) > 1 {
		for _, c := range matching {
			if c.Shared {
				return procTarget(db, c.SpellID)
			}
		}
	}
	return matching[0].SpellID, nil
}

// The spell a proc-trigger aura fires, named by EffectTriggerSpell or, where that is 0, by the base
// points; any other aura is its own buff.
func procTarget(db *sql.DB, aura int32) (int32, error) {
	var target int32
	err := db.QueryRow(`
		SELECT CASE WHEN EffectTriggerSpell != 0 THEN EffectTriggerSpell ELSE CAST(EffectBasePointsF AS INTEGER) END
		FROM SpellEffect WHERE SpellID = ? AND EffectAura IN (?, ?)`,
		aura, dbcenums.A_PROC_TRIGGER_SPELL, dbcenums.A_PROC_TRIGGER_SPELL_WITH_VALUE).Scan(&target)
	if err == sql.ErrNoRows {
		return aura, nil
	}
	return target, err
}

func isSharedTarget(target dbc.ImplicitTarget) bool {
	switch target {
	case dbc.TARGET_UNIT_CASTER_AREA_PARTY, dbc.TARGET_UNIT_CASTER_AREA_RAID,
		dbc.TARGET_UNIT_TARGET_ALLY, dbc.TARGET_UNIT_TARGET_RAID, dbc.TARGET_153:
		return true
	}
	return false
}

// SkillLineAbility.ClassMask is a bitmask over the game files' class ids, which is not the order
// proto.Class uses.
func ownerClassMask(class proto.Class) int32 {
	for _, c := range dbc.Classes {
		if c.ProtoClass == class {
			return int32(1) << (c.ID - 1)
		}
	}
	return 0
}

package database

import (
	"database/sql"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/tools/database/buffmanifest"
	"github.com/wowsims/forever/tools/database/dbc"
	"github.com/wowsims/forever/tools/database/overrides"
)

// The spells the sim can reach from gear without anything naming them first: every spell an item, an
// enchant or a set bonus casts. With the spells the nine class files are built from, they are the
// store's roots, and what those spells trigger is reached from there by reachableSpells. They are kept
// apart from the class roots: they are the items spell data group and the universe the proc audit
// asks its question over, and which query a root came from cannot be recovered from a merged list.
//
// The hand-kept extra spells are not here: they are added while the store is rendered, so that adding
// one and forgetting to regenerate fails the regeneration check instead of passing it - see
// withExtraIDs.
func gearRoots(db *sql.DB, t *spellTables) (gear []int32, err error) {
	for _, source := range []struct {
		name  string
		query string
	}{
		{"item effects", itemEffectSpellQuery()},
		{"enchant effects", enchantSpellQuery},
		{"set bonuses", `SELECT DISTINCT SpellID FROM ItemSetSpell WHERE SpellID > 0`},
		{"every class spell", everyClassSpellQuery()},
	} {
		if err := eachRow(db, source.query, func(rows *sql.Rows) error {
			var id int32
			if err := rows.Scan(&id); err != nil {
				return err
			}
			gear = append(gear, id)
			return nil
		}); err != nil {
			return nil, fmt.Errorf("%s: %w", source.name, err)
		}
	}

	return namedIDs(t, gear), nil
}

// Set by gen_spelldata -everyClassSpell: roots the store at every spell of a class family as well,
// for an export of the game files' class spells rather than a store the sim is built against.
var EveryClassSpell bool

// Selects nothing unless EveryClassSpell is set.
func everyClassSpellQuery() string {
	if !EveryClassSpell {
		return `SELECT SpellID FROM SpellClassOptions WHERE 0`
	}
	return `SELECT DISTINCT SpellID FROM SpellClassOptions WHERE SpellClassSet BETWEEN 3 AND 11`
}

// The spells a rank reads its numbers off by name. A rank that states no amount of its own keeps it
// on a same-name sibling - Frenzied Regeneration's heal is on 22845 - so the store carries those
// too: same name, same rank subtext, a shared class bit, and taught by a skill line rather than
// merely existing.
//
// One pass over everything already reached is enough: the relation is the name and subtext, so a
// sibling's siblings are the ones already in hand. The siblings come back by the spell they are
// siblings of, so one query serves every spell data group's reach.
func siblingSpells(db *sql.DB, ids []int32) (map[int32][]int32, error) {
	list := make([]string, len(ids))
	for i, id := range ids {
		list[i] = strconv.Itoa(int(id))
	}

	siblings := map[int32][]int32{}
	err := eachRow(db, `
		SELECT DISTINCT a.Spell, b.Spell
		FROM SkillLineAbility a
		JOIN SkillLineAbility b ON b.Spell != a.Spell AND (a.ClassMask & b.ClassMask) != 0
		JOIN SpellName na ON na.ID = a.Spell
		JOIN SpellName nb ON nb.ID = b.Spell AND nb.Name_lang = na.Name_lang
		JOIN Spell sa ON sa.ID = a.Spell
		JOIN Spell sb ON sb.ID = b.Spell AND sb.NameSubtext_lang = sa.NameSubtext_lang
		WHERE a.Spell IN (`+strings.Join(list, ", ")+`)
		ORDER BY b.Spell`, func(rows *sql.Rows) error {
		var of, id int32
		if err := rows.Scan(&of, &id); err != nil {
			return err
		}
		siblings[of] = append(siblings[of], id)
		return nil
	})
	return siblings, err
}

// The hand-kept extra spells, added to the captured roots while the store is rendered rather than
// while the game files tables are read, so an entry added without a regeneration shows up as a row the
// committed store lacks.
func withExtraIDs(t *spellTables, roots []int32) ([]int32, error) {
	extras := make([]int32, 0, len(overrides.ExtraSpells))
	for _, extra := range overrides.ExtraSpells {
		if extra.Reason == "" {
			return nil, fmt.Errorf("the extra spell %d states no reason", extra.SpellID)
		}
		if extra.SpellDataGroup != "" && extra.SpellDataGroup != groupRacials {
			return nil, fmt.Errorf("the extra spell %d names spell data group %q, and only %q is named by hand",
				extra.SpellID, extra.SpellDataGroup, groupRacials)
		}
		// Dropped like any other unnamed id, but said out loud: it was written down on purpose.
		if _, named := t.Names[extra.SpellID]; !named {
			fmt.Fprintf(progress, "spelldata: extra spell %d is no spell in the game files, so the store does not carry it\n",
				extra.SpellID)
		}
		extras = append(extras, extra.SpellID)
	}

	return namedIDs(t, roots, extras, buffManifestIDs()), nil
}

// Every spell the buff manifest names. The generated buffs read their numbers off these rows, so they
// are roots the way the extra spells are, and added at the same point for the same reason: a
// manifest row added without a regeneration fails the check.
func buffManifestIDs() []int32 {
	var ids []int32
	for _, spec := range buffmanifest.All() {
		ids = append(ids, buffSpecIDs(spec)...)
	}
	return ids
}

func buffSpecIDs(spec buffmanifest.Row) []int32 {
	ids := []int32{spec.SpellID, spec.CastID, spec.Talent}
	if spec.ImpAction != nil {
		ids = append(ids, spec.ImpAction.SpellID)
	}
	for _, rank := range spec.Ranks {
		ids = append(ids, rank.SpellID, rank.CastID)
	}
	return ids
}

// The ids of every list that this build names as a spell, deduped and in search order. An id no
// SpellName row carries is not a spell here: ItemEffect keeps rows for spells the game files dropped,
// and following one would put an empty row in the store.
func namedIDs(t *spellTables, lists ...[]int32) []int32 {
	set := map[int32]bool{}
	for _, list := range lists {
		for _, id := range list {
			if _, named := t.Names[id]; named {
				set[id] = true
			}
		}
	}
	return slices.Sorted(maps.Keys(set))
}

// The gear and consumables gen_db ships, which is where the item procs the sim registers come from.
// Each half is the predicate gen_db itself selects with: SimItemFilter for LoadAndWriteRawItems and
// ConsumableItemFilter for LoadAndWriteConsumables. Both halves let the allowlists through, as gen_db
// does: Hand of Justice and the raid consumables are shipped by id rather than by predicate.
//
// The predicate runs over Item left-joined to ItemSparse alone, without the inner joins to
// ItemClass, RandPropPoints and the armour tables that gen_db's own query carries, so this selects
// at least the items gen_db ships and possibly a few more. Extra rows only add spells to the store,
// which is the safe direction: a missing one is a spell the sim cannot read.
func itemEffectSpellQuery() string {
	return `
		SELECT DISTINCT ie.SpellID
		FROM ItemEffect ie
		JOIN ItemXItemEffect ixie ON ixie.ItemEffectID = ie.ID
		JOIN Item i ON i.ID = ixie.ItemID
		LEFT JOIN ItemSparse s ON s.ID = i.ID
		WHERE ie.SpellID > 0 AND (
			(` + SimItemFilter(core.CharacterLevel) + `)
			OR (` + ConsumableItemFilter() + `)
			OR i.ID IN (` + allowListedItemIDs() + `))`
}

// The items gen_db ships by id: ItemAllowList bypasses the gear filter and catalogueConsumableIDs
// are added to the consumable query. Hand of Justice and Skullflame Shield have no ItemSparse row in
// this build, which is why the join to it above is a left one: an allowlisted item is taken by id,
// whatever the sparse table says about it.
func allowListedItemIDs() string {
	ids := append(slices.Collect(maps.Keys(ItemAllowList)), catalogueConsumableIDs()...)
	slices.Sort(ids)

	list := make([]string, len(ids))
	for i, id := range ids {
		list[i] = strconv.Itoa(int(id))
	}
	return strings.Join(list, ", ")
}

// An enchant states the spell it applies in the EffectArg of the effect that applies it, which is
// what LoadAndWriteRawEnchants reads as its spell id: effect 1 and effect 3 are the two that name a
// spell.
var enchantSpellQuery = enchantSlotsCTE + fmt.Sprintf(`
	SELECT DISTINCT SpellID FROM slots WHERE Effect IN (%d, %d) AND SpellID > 0`,
	dbc.ITEM_ENCHANTMENT_COMBAT_SPELL, dbc.ITEM_ENCHANTMENT_EQUIP_SPELL)

// Every spell the roots reach: what an effect triggers, what an actionbar override swaps in, what a
// hand link names, and what a tooltip's $<id> token points at. A spell the sim registers reads its
// numbers off those, so the store carries them all rather than the roots alone.
func reachableSpells(t *spellTables, roots []int32) []int32 {
	seen := map[int32]bool{}
	queue := make([]int32, 0, len(roots))
	for _, id := range roots {
		if !seen[id] {
			seen[id] = true
			queue = append(queue, id)
		}
	}

	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		for _, next := range spellEdges(t, id) {
			if seen[next] {
				continue
			}
			if _, named := t.Names[next]; !named {
				continue
			}
			seen[next] = true
			queue = append(queue, next)
		}
	}
	return slices.Sorted(maps.Keys(seen))
}

// A_OVERRIDE_ACTIONBAR_SPELLS states the spell it swaps in as its base points, the way
// overrideReplacements reads it.
func spellEdges(t *spellTables, id int32) []int32 {
	var next []int32
	for _, e := range t.Effects[id] {
		if e.TriggerID > 0 {
			next = append(next, e.TriggerID)
		}
		if e.Aura == dbcenums.A_OVERRIDE_ACTIONBAR_SPELLS && e.BasePoints > 0 {
			next = append(next, int32(e.BasePoints))
		}
	}
	next = append(next, handTriggered(id)...)
	return append(next, t.referencedIDs(id)...)
}

// The spells a server-side handler casts off this one, which no game files row states.
func handTriggered(id int32) []int32 {
	var ids []int32
	for _, link := range overrides.HandTriggers {
		if link.Spell == id {
			ids = append(ids, link.Triggers)
		}
	}
	return ids
}

// The ids this one's tooltip names where its handler casts another spell.
func handReplaced(id int32) []int32 {
	var ids []int32
	for _, link := range overrides.HandTriggers {
		if link.Spell == id && link.Replaces != 0 {
			ids = append(ids, link.Replaces)
		}
	}
	return ids
}

func parseSpellID(s string) int32 {
	id, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return int32(id)
}

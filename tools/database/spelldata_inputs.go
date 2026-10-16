package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/wowsims/forever/tools/database/dbc"
)

// The game files rows sim/core/spelldata is built from, captured next to the generated store so
// that the store can be rebuilt - and checked - without the game files database, which is
// gitignored and comes from a local WoW install.
//
// Raw game files rows, copied as the loaders in spelldata_store.go read them:
//
//   - every SpellName row, whole. It is the universe the closure tests an edge against, so cutting it
//     to the store's own ids would make a spell that ought to be reached look like an id this build
//     does not name. The other tables are cut to the store's ids, which loses nothing: the closure
//     only ever reads the rows of a spell it has already reached.
//   - the talent tree's nodes, and the points each definition states.
//   - the spell granting the enchant of each enchant equip spell in the store, with that spell's
//     description: the equip spell ships none of its own.
//   - the chance the enchantments state for each enchant combat spell in the store.
//   - every SpellShapeshiftForm row, whole, since renderFormsFile reads it on its own rather than
//     through the store's ids.
//
// Derived, and captured all the same because re-deriving it needs tables the store does not otherwise
// read: the root ids. They come from the item, enchant and set-bonus tables, from the ladder and tree
// spells the class files are built from, and from the sibling-by-name rule, and they are captured
// wholesale - so a change to how a root is found shows up only once the store is regenerated.
//
// Everything else is left to be re-derived from the above: the closure, the hand links, the talent
// curves, the tooltip hints, the overrides and the rows themselves. That is what makes the
// regeneration a check on the generator rather than a copy of its answer. The hand-kept extra spells
// are added while rendering for the same reason - adding one without regenerating renders a row the
// committed store does not carry, and the check fails.

// Beside the game files extraction rather than in it: assets/db_inputs/dbc is gitignored, being
// tens of megabytes rebuilt by `make db`, and this file has to be committed for the check to run
// without a database. Gzipped like the extraction's files, and named the same way, since it is read
// back the same way.
const spellStoreInputsPath = "assets/db_inputs/spell_store_inputs.json"

type storeInputs struct {
	// The game files rows, whose fields encode first and under their own names.
	spellTables

	Roots []int32

	// The roots that came from the item, enchant and set-bonus tables alone. The store is rendered
	// from Roots, which holds these and the class ladder both; this subset is what the proc audit
	// asks its question over, and no reading of a row says which query reached it.
	ItemRoots []int32

	// The talent nodes of every class tree in the order storeCurves reads them, since the first
	// definition to price a spell is the one kept, and the points each definition states by the
	// game files' EffectIndex and rank.
	TraitNodes  []traitNode
	TraitPoints map[int32]map[int32]map[int32]float64

	// Every SpellShapeshiftForm row, whole: renderFormsFile reads it directly rather than through the
	// store's own ids, since a form names no spell for the closure to reach it by.
	Forms []formRow

	// Roots by spell data group, whose union is Roots, and the game build and highest hotfix push each
	// group's rows were taken from. A partial update takes some groups from the database and keeps the
	// others as they are here (see spelldata_groups.go).
	SpellDataGroupRoots map[string][]int32
	Builds              map[string]string
	Hotfixes            map[string]int32

	owners *rowOwners // ownersOf's answer, worked out on first use; never encoded
}

// The tables the generator reads, from the captured rows.
//
// The effects are cloned because storeCurves bakes a one-rank talent's value into its base points:
// sharing the slices would put that back into what is written out, and the regeneration would then
// read an answer it is supposed to re-derive. Nothing else is written through.
func (in *storeInputs) tables() *spellTables {
	t := in.spellTables
	t.Effects = make(map[int32][]storeEffect, len(in.Effects))
	for id, rows := range in.Effects {
		t.Effects[id] = slices.Clone(rows)
	}
	return &t
}

// The rows of the store's own ids, and every name in the build. Restricting the rest to those ids is
// what keeps the file to the size of the store rather than the size of the game files.
func captureStoreInputs(t *spellTables, roots []int32, ids []int32,
	nodes []traitNode, points map[int32]map[int32]map[int32]float64) *storeInputs {
	in := &storeInputs{
		spellTables: *newSpellTables(),
		Roots:       roots,
		TraitNodes:  nodes,
		TraitPoints: points,
	}
	in.Names = t.Names

	for _, id := range ids {
		keepString(in.Subtexts, id, t.Subtexts[id])
		keepString(in.Descriptions, id, t.Descriptions[id])

		keepValue(in.Misc, id, t.Misc)
		keepValue(in.Levels, id, t.Levels)
		keepValue(in.Cooldowns, id, t.Cooldowns)
		keepValue(in.Categories, id, t.Categories)
		keepValue(in.AuraOptions, id, t.AuraOptions)
		keepValue(in.ClassOptions, id, t.ClassOptions)
		keepValue(in.Interrupts, id, t.Interrupts)
		keepValue(in.Shapeshift, id, t.Shapeshift)
		keepValue(in.AuraRestrictions, id, t.AuraRestrictions)
		keepValue(in.Targets, id, t.Targets)
		keepValue(in.CreatureType, id, t.CreatureType)
		keepValue(in.Requirements, id, t.Requirements)
		keepValue(in.Equipped, id, t.Equipped)

		keepSlice(in.Labels, id, t.Labels)
		keepSlice(in.Powers, id, t.Powers)
		keepSlice(in.Reagents, id, t.Reagents)
		keepSlice(in.Effects, id, t.Effects)

		keepValue(in.EnchantGrants, id, t.EnchantGrants)
		keepValue(in.EnchantChances, id, t.EnchantChances)
		if grant, ok := in.EnchantGrants[id]; ok {
			keepString(in.Descriptions, grant, t.Descriptions[grant])
		}
	}
	return in
}

// A row is kept only where the game files state one: the absence of a SpellLevels row is itself a
// reading - no level scaling - so writing a zero row in its place would change what the store says.
func keepValue[V any](into map[int32]V, id int32, from map[int32]V) {
	if v, ok := from[id]; ok {
		into[id] = v
	}
}

func keepSlice[V any](into map[int32][]V, id int32, from map[int32][]V) {
	if v := from[id]; len(v) > 0 {
		into[id] = v
	}
}

func keepString(into map[int32]string, id int32, s string) {
	if s != "" {
		into[id] = s
	}
}

func writeStoreInputs(in *storeInputs) error {
	out, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("encoding the store's inputs: %w", err)
	}
	if err := dbc.WriteGzipFile(spellStoreInputsPath, out); err != nil {
		return fmt.Errorf("writing %s: %w", spellStoreInputsPath, err)
	}
	fmt.Fprintf(progress, "spelldata: wrote %s, %d names and %d roots\n",
		spellStoreInputsPath, len(in.Names), len(in.Roots))
	return nil
}

// Whether the committed capture is the one the database gives today. Compared as the encoded bytes
// rather than through the decoded values, since a map is written in key order and a slice the game
// files state as empty decodes as nil - neither of which the rendered store can tell apart, but
// both of which a comparison of the values would trip over.
func storeInputsAreCurrent(fresh *storeInputs) (bool, error) {
	committed, err := dbc.ReadGzipFile(spellStoreInputsPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("reading %s: %w", spellStoreInputsPath, err)
	}

	out, err := json.Marshal(fresh)
	if err != nil {
		return false, fmt.Errorf("encoding the store's inputs: %w", err)
	}
	return bytes.Equal(committed, out), nil
}

// The committed inputs, or nil where none are committed yet.
func loadCommittedInputs() (*storeInputs, error) {
	if _, err := os.Stat(spellStoreInputsPath); os.IsNotExist(err) {
		return nil, nil
	}
	return readStoreInputs(spellStoreInputsPath)
}

// The game build and hotfix push each spell data group of the committed inputs was taken from,
// decoded alone: the rows are not needed to choose a run's groups or to check its database.
func readCommittedBuilds() (builds map[string]string, hotfixes map[string]int32, err error) {
	if _, err := os.Stat(spellStoreInputsPath); os.IsNotExist(err) {
		return nil, nil, nil
	}
	raw, err := dbc.ReadGzipFile(spellStoreInputsPath)
	if err != nil {
		return nil, nil, fmt.Errorf("reading %s: %w", spellStoreInputsPath, err)
	}
	var stamps struct {
		Builds   map[string]string
		Hotfixes map[string]int32
	}
	if err := json.Unmarshal(raw, &stamps); err != nil {
		return nil, nil, fmt.Errorf("decoding %s: %w", spellStoreInputsPath, err)
	}
	return stamps.Builds, stamps.Hotfixes, nil
}

// The committed inputs, for a caller with no game files database. The path is relative to the
// repository root, the way every other path the generator reads is.
func readStoreInputs(path string) (*storeInputs, error) {
	raw, err := dbc.ReadGzipFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	in := &storeInputs{}
	if err := json.Unmarshal(raw, in); err != nil {
		return nil, fmt.Errorf("decoding %s: %w", path, err)
	}
	return in, nil
}

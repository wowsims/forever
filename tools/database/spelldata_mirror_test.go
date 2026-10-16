package database

// The generator builds the store's rows in its own structs in spelldata_store.go, which carry its
// bookkeeping beside the store's fields. The cost of that is a mirror which can drift: a field added
// to the store would simply never be written, and the row would read as a zero the game files never
// stated. This asserts the two sides name the same fields.

import (
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/spelldata"
)

var mirroredTypes = []struct {
	store  any
	mirror any

	// Mirror fields the store has no field for, because they are the generator's own bookkeeping.
	generatorOnly []string
}{
	{
		store:  spelldata.Spell{},
		mirror: storeSpell{},
	},
	{
		store:  spelldata.Effect{},
		mirror: storeEffect{},
	},
	{
		store:  spelldata.Power{},
		mirror: storePower{},
	},
	{
		store:  spelldata.Reagent{},
		mirror: storeReagent{},
	},
}

func TestSpellStoreMirrorNamesEveryField(t *testing.T) {
	for _, m := range mirroredTypes {
		fields := exportedFields(m.store)
		mirror := exportedFields(m.mirror)

		var missing []string
		for _, field := range fields {
			if !slices.Contains(mirror, field) {
				missing = append(missing, field)
			}
		}
		if len(missing) > 0 {
			t.Errorf("%T: the generator's mirror has no field for %s - add it to %T and emit it",
				m.store, strings.Join(missing, ", "), m.mirror)
		}

		var stale []string
		for _, field := range mirror {
			if !slices.Contains(fields, field) && !slices.Contains(m.generatorOnly, field) {
				stale = append(stale, field)
			}
		}
		sort.Strings(stale)
		if len(stale) > 0 {
			t.Errorf("%T: %s is not a field of %T - the store dropped or renamed it",
				m.mirror, strings.Join(stale, ", "), m.store)
		}
	}
}

func exportedFields(v any) []string {
	var fields []string
	typ := reflect.TypeOf(v)
	for i := 0; i < typ.NumField(); i++ {
		if field := typ.Field(i); field.IsExported() {
			fields = append(fields, field.Name)
		}
	}
	return fields
}

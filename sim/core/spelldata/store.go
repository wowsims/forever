package spelldata

import (
	"fmt"
	"sort"

	"github.com/wowsims/forever/sim/core/dbcenums"
)

// What the generated half of the package has to provide. tools/database writes spells_auto_gen.go;
// nothing else in the package may assume it exists, so the store starts empty and answers Nil for
// every id until it is installed.
//
// spells_auto_gen.go declares:
//
//	var generatedSpells = []Spell{...}                  // every row, by strictly increasing ID
//	var generatedCurves = map[int32][][]float64{...}     // talent curve values, [effect position][rank - 1]
//	var generatedHandTriggers = map[int32][]int32{...}   // spells a server-side handler casts, by caster
//	func init() { install(generatedSpells, generatedCurves, generatedHandTriggers) }
//
// Calling install from the generated file's own init is what keeps the order right: package-level
// variables are initialised before any init runs, so the data is complete when install indexes it,
// and no file-name ordering is relied on.
//
// Two of the row's fields are named for the id they hold, because a Go field and a method cannot
// share a name: Effect.TriggerID feeds Trigger(), and Spell.RefIDs feeds Refs(). The emitter writes
// those two names.
var spells = []Spell{}

// Talent curve values by spell id, as [effect position][rank - 1]. The effect is at the position
// EffectN counts by, not at the game files' EffectIndex, which has gaps. A trait talent states one
// spell and its per-rank numbers live here rather than on ranks of their own.
var curves = map[int32][][]float64{}

// Every row as a pointer, in the same order as spells, so All hands out a view without rebuilding it.
var all []*Spell

// The spells a server-side handler casts off a given spell, which no effect of it states. The
// game files' own columns are left as they state them.
var handTriggers = map[int32][]int32{}

// Which spells reach a given spell: every effect that triggers it, every handler that casts it, and
// every actionbar override that replaces a spell with it.
var drivers = map[int32][]int32{}

func install(rows []Spell, rowCurves map[int32][][]float64, rowHandTriggers map[int32][]int32) {
	curves = rowCurves
	handTriggers = rowHandTriggers
	setSpells(rows)
}

// Tests build their rows by hand and swap them in here too. The indices are rebuilt, the curves and
// the hand triggers are left alone so a test can state its own.
func setSpells(rows []Spell) {
	spells = rows

	all = make([]*Spell, len(spells))
	drivers = map[int32][]int32{}

	for i := range spells {
		s := &spells[i]
		all[i] = s

		// Find binary searches, so an unsorted or duplicated row would silently answer the wrong
		// spell.
		if i > 0 && s.ID <= spells[i-1].ID {
			panic(fmt.Sprintf("spelldata: spell ids are out of order at %d: %d after %d",
				i, s.ID, spells[i-1].ID))
		}

		for _, e := range s.Effects {
			if e.TriggerID != 0 {
				addDriver(e.TriggerID, s.ID)
			}
			// A_OVERRIDE_ACTIONBAR_SPELLS states the replacing spell in its base points, which
			// makes the overriding spell a driver of it the same way a trigger effect is.
			if e.Aura == dbcenums.A_OVERRIDE_ACTIONBAR_SPELLS && e.BasePoints > 0 {
				addDriver(int32(e.BasePoints), s.ID)
			}
		}
		for _, triggered := range handTriggers[s.ID] {
			addDriver(triggered, s.ID)
		}
	}
}

// Two effects of the same spell can name the same trigger, and the caller wants the spell once.
func addDriver(triggered int32, driver int32) {
	ids := drivers[triggered]
	if len(ids) > 0 && ids[len(ids)-1] == driver {
		return
	}
	drivers[triggered] = append(ids, driver)
}

// The row for an id, or Nil when the store does not carry it. The row is the store's own, not a
// copy: callers must not write through it.
func Find(id int32) *Spell {
	i := sort.Search(len(spells), func(i int) bool { return spells[i].ID >= id })
	if i < len(spells) && spells[i].ID == id {
		return &spells[i]
	}
	return Nil
}

// For a spell a caller depends on: a missing row is a generator problem, not a runtime condition. The
// row is the store's own, not a copy.
func MustFind(id int32) *Spell {
	s := Find(id)
	if s == Nil {
		panic(fmt.Sprintf("spelldata: spell %d is not in the store (regenerate or add it to tools/database/overrides/extra_spells.go)", id))
	}
	return s
}

// Every row in id order. The rows are shared, not copied, so callers must not write through them.
func All() []*Spell {
	return all
}

// Every row with this exact name, by linear scan over the whole store. A sim should name the spell
// by id.
func ByName(name string) []*Spell {
	var out []*Spell
	for i := range spells {
		if spells[i].Name == name {
			out = append(out, &spells[i])
		}
	}
	return out
}

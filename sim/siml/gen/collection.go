// Package gen writes the SIML name table, sim/siml/names.json (SPEC.md §7.4): it asks the sim what every spec
// registers, adds a name for each spell, aura, item, enchant and random suffix the table does not name yet, and
// keeps every name the table already holds.
package gen

import (
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/siml"
)

// What the sim and the database hold, which the names are generated from.
type Collection struct {
	// Every spell and aura any build registers, by namespace.
	Registered map[siml.Namespace]map[Identity]Registration
	// What the builds of a class's specs register with no gear, consumables or preset encounter.
	Classes map[proto.Class]map[siml.Namespace]map[Identity]bool
	// What gear, consumables and preset encounters add to a spec's build beyond what its class registers.
	Shared map[siml.Namespace]map[Identity]bool
	// The specs of each class, by Player spec field name.
	Specs map[proto.Class][]string
	// The items whose effect a spec's build lists only as an internal cooldown trigger, by item id, with those
	// specs in byte order.
	ICDItems map[int32][]string

	SpellNames map[int32]string // Game names by spell id.
	Items      []Item           // In the database's order: its items, then its consumables and item icons.
	Enchants   []EnchantRow     // In the database's order.
	Suffixes   []Suffix

	Skipped []string // What could not be built, for the report.
}

// A spell id and tag, what an ActionID means to the sim: it finds a spell or an aura by them whatever the rank.
type Identity struct{ SpellID, Tag int32 }

type Registration struct {
	Rank     int32    // The highest rank a build registers it with.
	Castable bool     // Whether a build registers it castable from the APL (SpellFlagAPL), which SpellStats.is_castable reports.
	Specs    []string // The specs whose builds register it, in byte order.
	ICDOnly  bool     // In one build: an aura it lists only as an internal cooldown trigger.
	ICDSpecs []string // The specs whose builds list the aura only as an internal cooldown trigger, in byte order.
}

type Item struct {
	ID    int32
	Name  string
	Level int32 // The item level; 0 when the database states none.
}

// One enchant row: an effect id, its name, and the slots of the item types it fits.
type EnchantRow struct {
	EffectID int32
	Name     string
	Slots    []proto.ItemSlot // In ItemSlot order.
}

type Suffix struct {
	ID   int32
	Name string
}

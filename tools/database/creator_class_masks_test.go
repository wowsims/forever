package database

// A soulbound item only class spells create belongs to that class, though the game files leave it
// open to every class; a tradable one stays open. Reads SpellEffect and SkillLineAbility, so it
// skips without tools/database/wowsims.db.

import (
	"maps"
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/dbc"
)

func TestCreatorClassMasks(t *testing.T) {
	masks, err := LoadCreatorClassMasks(openBuffTestDB(t))
	if err != nil {
		t.Fatalf("%v", err)
	}

	mage, warlock := proto.Class_ClassMage, proto.Class_ClassWarlock
	want := map[int]proto.Class{
		1254:   warlock, // Lesser Firestone
		13699:  warlock, // Firestone
		13700:  warlock, // Greater Firestone
		13701:  warlock, // Major Firestone
		5522:   warlock, // Spellstone
		13602:  warlock, // Greater Spellstone
		13603:  warlock, // Major Spellstone
		5232:   warlock, // Minor Soulstone
		16892:  warlock, // Lesser Soulstone
		16893:  warlock, // Soulstone
		16895:  warlock, // Greater Soulstone
		16896:  warlock, // Major Soulstone
		5514:   mage,    // Mana Agate
		5513:   mage,    // Mana Jade
		8007:   mage,    // Mana Citrine
		8008:   mage,    // Mana Ruby
		277651: mage,    // Bundle of Scrolls
	}

	for id, class := range want {
		if got := dbc.GetClassesFromClassMask(masks[id]); !slices.Equal(got, []proto.Class{class}) {
			t.Errorf("item %d: classes %v, want [%v]", id, got, class)
		}
	}
	for _, id := range slices.Sorted(maps.Keys(masks)) {
		if _, ok := want[id]; !ok {
			t.Errorf("item %d: class mask %d derived, want it open to every class", id, masks[id])
		}
	}
}

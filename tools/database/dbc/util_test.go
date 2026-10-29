package dbc

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestGetRacesFromRaceMask(t *testing.T) {
	for _, tc := range []struct {
		mask int64
		want []proto.Race
	}{
		{1<<1 | 1<<9, []proto.Race{proto.Race_RaceBloodElf, proto.Race_RaceOrc}},
		{1<<32 | 1<<0, []proto.Race{proto.Race_RaceHuman, proto.Race_RaceHighOrderSkyborne}},
		{1<<33 | 1<<10, []proto.Race{proto.Race_RaceWindshaperSkyborne}},
		{-1, nil},
		{0b11<<32 | 0b10_1111_1111, nil},
	} {
		if got := GetRacesFromRaceMask(tc.mask); !slices.Equal(got, tc.want) {
			t.Errorf("mask %#x: %v, want %v", tc.mask, got, tc.want)
		}
	}
}

// Goblin and Draenei are not playable races, so their bits restrict nothing.
func TestGetFactionRestrictionFromRaceMask(t *testing.T) {
	for _, tc := range []struct {
		mask int64
		want proto.UIItem_FactionRestriction
	}{
		{int64(Tauren | Goblin), proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY},
		{int64(Human | Dwarf | NightElf | Gnome | Draenei), proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY},
		{int64(HighOrderSkyborne), proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY},
		{int64(WindshaperSkyborne), proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY},
		{int64(Human | Orc), proto.UIItem_FACTION_RESTRICTION_UNSPECIFIED},
		{int64(Goblin | Draenei), proto.UIItem_FACTION_RESTRICTION_UNSPECIFIED},
		{-1, proto.UIItem_FACTION_RESTRICTION_UNSPECIFIED},
		{0, proto.UIItem_FACTION_RESTRICTION_UNSPECIFIED},
	} {
		if got := GetFactionRestrictionFromRaceMask(tc.mask); got != tc.want {
			t.Errorf("mask %#x: %v, want %v", tc.mask, got, tc.want)
		}
	}
}

// The faction flags decide before the race mask does.
func TestParseItemFlagsFactionRestriction(t *testing.T) {
	for _, tc := range []struct {
		name string
		item Item
		want proto.UIItem_FactionRestriction
	}{
		{"horde flag over an alliance mask", Item{Flags1: HORDE_SPECIFIC, RaceMask: int64(Human)}, proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY},
		{"alliance mask without a flag", Item{RaceMask: int64(Human | Gnome)}, proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY},
		{"every race without a flag", Item{RaceMask: -1}, proto.UIItem_FACTION_RESTRICTION_UNSPECIFIED},
	} {
		uiItem := &proto.UIItem{}
		tc.item.ParseItemFlags(uiItem)
		if uiItem.FactionRestriction != tc.want {
			t.Errorf("%s: %v, want %v", tc.name, uiItem.FactionRestriction, tc.want)
		}
	}
}

// Bit n-1 is class n; bits of classes no player can be (6, 10, 12-15) drop out.
func TestGetClassesFromClassMask(t *testing.T) {
	for _, tc := range []struct {
		mask int
		want []proto.Class
	}{
		{1 << 1, []proto.Class{proto.Class_ClassPaladin}},
		{1<<3 | 1<<0, []proto.Class{proto.Class_ClassWarrior, proto.Class_ClassRogue}},
		{1<<10 | 1<<6, []proto.Class{proto.Class_ClassShaman, proto.Class_ClassDruid}},
		{1<<7 | 1<<9 | 1<<11 | 1<<12 | 1<<13 | 1<<14, []proto.Class{proto.Class_ClassMage}},
		{1<<5 | 1<<9, nil},
		{0, nil},
		{-1, nil},
	} {
		if got := GetClassesFromClassMask(tc.mask); !slices.Equal(got, tc.want) {
			t.Errorf("mask %#x: %v, want %v", tc.mask, got, tc.want)
		}
	}
}

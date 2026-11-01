package database

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestPvpSourceFollowsTheItemName(t *testing.T) {
	cases := map[string]bool{
		"Knight-Captain's Plate Helm":                  true,
		"Premier Warlord's Linked Helm":                true,
		"Premier Chain Boots":                          true,
		"Highlander's Plate Girdle":                    true,
		"Insignia of the Horde":                        true,
		"Thunderfury, Blessed Blade of the Windseeker": false,
		"Commander's Boots":                            false,
		"Knight's Lance":                               false,
		"Sentinel's Guard":                             false,
		"":                                             false,
	}
	for name, want := range cases {
		item := &proto.UIItem{Name: name}
		InferPvpSource(item)
		if got := len(item.Sources) == 1 && item.Sources[0].GetPvp() != nil; got != want {
			t.Errorf("%q: pvp source = %v, want %v", name, got, want)
		}
	}
}

func TestPvpSourceComesAfterTheItemsOtherSources(t *testing.T) {
	item := &proto.UIItem{Name: "Stormpike Battle Charger", Sources: []*proto.UIItemSource{
		{Source: &proto.UIItemSource_Rep{Rep: &proto.RepSource{RepFactionId: 730}}},
	}}
	InferPvpSource(item)
	if len(item.Sources) != 2 || item.Sources[0].GetRep() == nil || item.Sources[1].GetPvp() == nil {
		t.Fatalf("sources = %v, want the rep source then the pvp source", item.Sources)
	}
}

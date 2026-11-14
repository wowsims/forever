package database

import (
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
)

// Honor rank and battleground item name prefixes. "Premier " items are Forever's reissued PvP items and all count.
var pvpNamePrefixes = []string{
	"Premier ",
	"Private's ",
	"Corporal's ",
	"Sergeant's ",
	"Master Sergeant's ",
	"Sergeant Major's ",
	"Knight-Lieutenant's ",
	"Knight-Captain's ",
	"Knight-Champion's ",
	"Lieutenant Commander's ",
	"Marshal's ",
	"Field Marshal's ",
	"Grand Marshal's ",
	"Scout's ",
	"Grunt's ",
	"Senior Sergeant's ",
	"First Sergeant's ",
	"Stone Guard's ",
	"Blood Guard's ",
	"Legionnaire's ",
	"Centurion's ",
	"Champion's ",
	"Lieutenant General's ",
	"General's ",
	"Warlord's ",
	"High Warlord's ",
	"Highlander's ",
	"Defiler's ",
	"Sentinel's ",
	"Outrider's ",
	"Stormpike ",
	"Frostwolf ",
	"Insignia of the ",
}

// PvE items whose names start with a PvP prefix.
var pveItemNames = map[string]bool{
	"Legionnaire's Leggings": true,
	"Sentinel's Guard":       true,
	"Sentinel's Belt":        true,
	"Sentinel's Gauntlet":    true,
	"Sentinel's Plate":       true,
	"Sentinel's Boots":       true,
	"Sentinel's Gloves":      true,
	"Sentinel's Sabatons":    true,
	"Sentinel's Waistguard":  true,
	"Sentinel's Pauldrons":   true,
	"Sentinel's Wristguards": true,
	"Scout's Dagger":         true,
	"Champion's Legplates":   true,
}

func isPvpItemName(name string) bool {
	if pveItemNames[name] {
		return false
	}
	for _, prefix := range pvpNamePrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

// Appends a PvpSource after the item's other sources when its name is a PvP name. Items of the same name can share
// one Sources slice, so the append copies it.
func InferPvpSource(item *proto.UIItem) {
	if isPvpItemName(item.Name) {
		item.Sources = append(slices.Clip(item.Sources), &proto.UIItemSource{Source: &proto.UIItemSource_Pvp{Pvp: &proto.PvpSource{}}})
	}
}

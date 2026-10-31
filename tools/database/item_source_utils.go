package database

import (
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
)

type phaseByName struct {
	fragment string
	phase    int32
}

// The phase ladders InferPhase reads, empty until Forever's phases are known.
var (
	phaseBySetIlvl     = map[int32]int32{}
	phaseByCraftedIlvl = map[int32]int32{}
	phaseByDropZone    = map[int32]int32{}
	phaseByRepFaction  = map[int32]int32{}
	phaseByItemName    = []phaseByName{}
	phaseByItemID      = map[int32]int32{}
)

// Content phase from item level, drop zone, faction and name, for a phase ladder the game files do
// not state. Not called: Forever's phases are not known yet, so gen_db stamps every item phase 1.
func InferPhase(item *proto.UIItem) int32 {
	ilvl := item.ScalingOptions[int32(0)].Ilvl

	if phase, ok := phaseByItemID[item.Id]; ok {
		return phase
	}
	if item.SetId > 0 {
		if phase, ok := phaseBySetIlvl[ilvl]; ok {
			return phase
		}
	}
	for _, src := range item.Sources {
		if src.GetCrafted() != nil {
			if phase, ok := phaseByCraftedIlvl[ilvl]; ok {
				return phase
			}
		}
		if drop := src.GetDrop(); drop != nil {
			if phase, ok := phaseByDropZone[drop.ZoneId]; ok {
				return phase
			}
		}
		if rep := src.GetRep(); rep != nil {
			if phase, ok := phaseByRepFaction[rep.RepFactionId]; ok {
				return phase
			}
		}
	}
	for _, entry := range phaseByItemName {
		if strings.Contains(item.Name, entry.fragment) {
			return entry.phase
		}
	}
	return 0
}

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
	"Scout's Dagger":         true,
	"Champion's Legplates":   true,
}

func isPvpItemName(name string) bool {
	if name == "" || pveItemNames[name] {
		return false
	}
	for _, prefix := range pvpNamePrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

// Appends a PvpSource after the item's other sources when its name is a PvP name.
func InferPvpSource(item *proto.UIItem) {
	if isPvpItemName(item.Name) {
		item.Sources = append(item.Sources, &proto.UIItemSource{Source: &proto.UIItemSource_Pvp{Pvp: &proto.PvpSource{}}})
	}
}

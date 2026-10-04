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

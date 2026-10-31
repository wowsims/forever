package druid

import (
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Files the feral kit in the APL picker's submenus once it is registered; Feral Charge files itself. A spell
// or aura left out stays under its category, Cooldowns for the major cooldowns.
func (druid *Druid) registerFeralMenuGroups() {
	groups := []struct {
		path    string
		ladders []spelldata.Ladder
	}{
		{"spell.forms", []spelldata.Ladder{spellData.CatForm, spellData.DireBearForm}},
		{"spell.cat", []spelldata.Ladder{spellData.Claw, spellData.Shred, spellData.Rake, spellData.Rip, spellData.FerociousBite, spellData.ShiftingPower}},
		{"spell.cat.stealth", []spelldata.Ladder{spellData.Prowl, spellData.Ravage, spellData.Pounce}},
		{"spell.bear", []spelldata.Ladder{spellData.Maul, spellData.Swipe, spellData.Lacerate, spellData.PrimalBite, spellData.Enrage, spellData.DemoralizingRoar}},
		{"spell.utility", []spelldata.Ladder{spellData.FaerieFire, spellData.Dash, spellData.Cower, spellData.Growl, spellData.ChallengingRoar, spellData.Bash}},
	}
	druid.OnSpellRegistered(func(spell *core.Spell) {
		for _, group := range groups {
			if slices.ContainsFunc(group.ladders, func(ladder spelldata.Ladder) bool { return ladder.Contains(spell) }) {
				spell.MenuGroup = group.path
				break
			}
		}
	})

	for path, auras := range map[string][]*core.Aura{
		"druid.forms": {druid.CatFormAura, druid.DireBearFormAura},
		"druid":       {druid.ClearcastingAura, druid.DashAura, druid.EnrageAura, druid.maulQueueAura, druid.ProwlAura},
	} {
		for _, aura := range auras {
			aura.MenuGroup = path
		}
	}
}

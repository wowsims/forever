package druid

import (
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Files the feral kit in the APL picker's submenus once it is registered, and orders the submenus. A spell or aura
// left out stays under its category, Cooldowns for Innervate and the racials.
func (druid *Druid) registerFeralMenuGroups() {
	groups := []struct {
		path    string
		ladders []spelldata.Ladder
	}{
		{"spell.forms", []spelldata.Ladder{spellData.CatForm, spellData.DireBearForm}},
		{"spell.cat", []spelldata.Ladder{spellData.Claw, spellData.Shred, spellData.Rake, spellData.Rip, spellData.FerociousBite, spellData.ShiftingPower}},
		{"spell.cat.stealth", []spelldata.Ladder{spellData.Prowl, spellData.Ravage, spellData.Pounce}},
		{"spell.bear", []spelldata.Ladder{spellData.Maul, spellData.Swipe, spellData.Lacerate, spellData.PrimalBite, spellData.Enrage, spellData.DemoralizingRoar}},
		{"spell.utility", []spelldata.Ladder{spellData.FaerieFire, spellData.Dash, spellData.Cower, spellData.Growl, spellData.ChallengingRoar, spellData.Bash, feralCharges}},
		// Filed beside the major cooldowns; the bear's Barkskin is one.
		{"cooldowns", []spelldata.Ladder{spellData.Barkskin, spellData.FrenziedRegeneration}},
	}
	druid.OnSpellRegistered(func(spell *core.Spell) {
		for _, group := range groups {
			if slices.ContainsFunc(group.ladders, func(ladder spelldata.Ladder) bool { return ladder.Contains(spell) }) {
				spell.MenuGroup = group.path
				break
			}
		}
	})

	for _, aura := range []*core.Aura{druid.CatFormAura, druid.DireBearFormAura} {
		aura.MenuGroup = "druid.forms"
	}
	for _, aura := range []*core.Aura{druid.ClearcastingAura, druid.DashAura, druid.EnrageAura, druid.FeralChargeMovementAura, druid.maulQueueAura, druid.ProwlAura} {
		aura.MenuGroup = "druid"
	}
	for _, aura := range []*core.Aura{druid.Barkskin.RelatedSelfBuff, druid.FrenziedRegenerationAura} {
		aura.MenuGroup = "cooldowns"
	}

	druid.MenuGroupOrder = []string{"spell.forms", "spell.cat", "spell.cat.stealth", "spell.bear", "spell.utility", "cooldowns", "druid.forms", "druid"}
}

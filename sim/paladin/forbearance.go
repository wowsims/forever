package paladin

import (
	"github.com/wowsims/forever/sim/core"
)

// Forbearance
// https://www.wowhead.com/forever/spell=25771
//
// Cannot be made invulnerable by Divine Shield, Divine Protection, Blessing of Protection, or
// shielded by Templar's Bulwark.
func (paladin *Paladin) registerForbearance() {
	rank := spellData.TemplarsBulwarkTriggered.Highest()
	paladin.Forbearance = paladin.RegisterAura(core.Aura{
		Label:     "Forbearance",
		ActionID:  core.ActionID{SpellID: rank.ID},
		Duration:  rank.Duration() - paladin.forbearanceReduction,
		MenuGroup: menuPaladin,
	})
}

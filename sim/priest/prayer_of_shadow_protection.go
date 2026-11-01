package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
)

func (priest *Priest) registerPrayerOfShadowProtection() {
	prayer := spellData.PrayerOfShadowProtection.Highest()
	priest.registerPrayer(prayer, func(unit *core.Unit, isPlayer bool, _ int32) *core.Aura {
		return buffs.NewSpellDataPrayerOfShadowProtectionAura(unit, isPlayer, prayer)
	})
}

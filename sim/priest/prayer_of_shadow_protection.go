package priest

import (
	"github.com/wowsims/forever/sim/core/buffs"
)

func (priest *Priest) registerPrayerOfShadowProtection() {
	priest.registerPrayer(spellData.PrayerOfShadowProtection.Highest(), buffs.NewSpellDataPrayerOfShadowProtectionAura)
}

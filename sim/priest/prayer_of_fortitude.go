package priest

import (
	"github.com/wowsims/forever/sim/core/buffs"
)

func (priest *Priest) registerPrayerOfFortitude() {
	priest.registerPrayer(spellData.PrayerOfFortitude.Highest(), buffs.NewSpellDataPrayerOfFortitudeAura)
}

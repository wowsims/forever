package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
)

func (priest *Priest) registerPrayerOfFortitude() {
	prayer := spellData.PrayerOfFortitude.Highest()
	priest.registerPrayer(prayer, func(unit *core.Unit, isPlayer bool, _ int32) *core.Aura {
		return buffs.NewSpellDataPrayerOfFortitudeAura(unit, isPlayer, prayer)
	})
}

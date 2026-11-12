package common

// Just import other directories, so importing common from elsewhere is enough.
import (
	_ "github.com/wowsims/forever/sim/common/classic"
	"github.com/wowsims/forever/sim/common/forever"
	_ "github.com/wowsims/forever/sim/core/buffs"
)

func RegisterAllEffects() {
	forever.RegisterAllOnUseCds()
	forever.RegisterAllProcs()
	forever.RegisterAllSetBonuses()
	forever.RegisterAllEnchants()
	forever.RegisterAllExplosives()
	forever.RegisterPotionsAndConjured()
	forever.RegisterAllTemporaryEnchantConsumables()
	forever.RegisterAllStatConsumables()
}

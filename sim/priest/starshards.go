package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var StarshardsRankMap = spellData.Starshards

func (priest *Priest) registerStarshardsSpell(_ int32, rank *spelldata.Spell) {
	config := priest.periodicDamageSpellConfig(rank, 0)
	// A tick meets the target's full resistance, where other channels' ticks meet a tenth (EV-0142).
	config.Flags |= core.SpellFlagFullResistanceTicks
	priest.RegisterSpell(config)
}

package priest

import (
	"github.com/wowsims/forever/sim/core/spelldata"
)

var StarshardsRankMap = spellData.Starshards

func (priest *Priest) registerStarshardsSpell(_ int32, rank *spelldata.Spell) {
	priest.RegisterSpell(priest.periodicDamageSpellConfig(rank, 0))
}

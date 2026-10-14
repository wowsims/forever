package priest

import (
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerChastise(_ int32, rank *spelldata.Spell) {
	priest.RegisterSpell(priest.directDamageSpellConfig(rank, 0))
}

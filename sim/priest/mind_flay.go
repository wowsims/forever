package priest

import (
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerMindFlaySpell(_ int32, rank *spelldata.Spell) {
	priest.RegisterSpell(priest.periodicDamageSpellConfig(rank, PriestSpellMindFlay))
}

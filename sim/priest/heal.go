package priest

import (
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerHeal(_ int32, rank *spelldata.Spell) {
	priest.RegisterSpell(priest.directHealSpellConfig(rank, PriestSpellHeal))
}

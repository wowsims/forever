package priest

import (
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerLesserHeal(_ int32, rank *spelldata.Spell) {
	priest.RegisterSpell(priest.directHealSpellConfig(rank, PriestSpellLesserHeal))
}

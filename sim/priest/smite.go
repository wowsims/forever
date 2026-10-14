package priest

import (
	"github.com/wowsims/forever/sim/core/spelldata"
)

var SmiteRankMap = spellData.Smite

func (priest *Priest) registerSmiteSpell(_ int32, rank *spelldata.Spell) {
	priest.RegisterSpell(priest.directDamageSpellConfig(rank, PriestSpellSmite))
}

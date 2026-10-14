package priest

import (
	"github.com/wowsims/forever/sim/core/spelldata"
)

var ShadowWordPainRankMap = spellData.ShadowWordPain

func (priest *Priest) registerShadowWordPainSpell(_ int32, rank *spelldata.Spell) {
	priest.RegisterSpell(priest.periodicDamageSpellConfig(rank, PriestSpellShadowWordPain))
}

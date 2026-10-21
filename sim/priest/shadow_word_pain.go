package priest

import (
	"github.com/wowsims/forever/sim/core/spelldata"
)

var ShadowWordPainRankMap = spellData.ShadowWordPain

func (priest *Priest) registerShadowWordPainSpell(_ int32, rank *spelldata.Spell) {
	config := priest.periodicDamageSpellConfig(rank, PriestSpellShadowWordPain)
	priest.attachPowerInLightMarker(&config)
	priest.RegisterSpell(config)
}

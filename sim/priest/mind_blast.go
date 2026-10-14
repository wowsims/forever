package priest

import (
	"github.com/wowsims/forever/sim/core/spelldata"
)

var MindBlastRankMap = spellData.MindBlast

func (priest *Priest) registerMindBlastSpell(_ int32, rank *spelldata.Spell) {
	priest.RegisterSpell(priest.directDamageSpellConfig(rank, PriestSpellMindBlast))
}

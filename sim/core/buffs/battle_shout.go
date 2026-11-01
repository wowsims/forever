package buffs

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Rank 7 comes from Manual of Battle Shout VII; the trainer's top rank is the generated row's Rank 6.
// The raid's shout is either rank, and a warrior registers both.
var battleShoutBookRank = spelldata.MustFind(25289)

// The attack power one rank grants, before the tier 2 set's bonus.
func BattleShoutRankValue(rank *spelldata.Spell) float64 {
	return battleShoutMeta.atRank(rank).Value(0)
}

func NewSpellDataBattleShoutRankAura(unit *core.Unit, isPlayer bool, rank *spelldata.Spell) *core.Aura {
	return newBuff(unit, battleShoutMeta.atRank(rank), isPlayer, 0)
}

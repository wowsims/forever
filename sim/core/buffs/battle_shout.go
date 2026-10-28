package buffs

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Rank 7 comes from Manual of Battle Shout VII; the trainer's top rank is the generated row's Rank 6.
// The raid's shout is either rank, and a warrior registers both.
var battleShoutBookRank = spelldata.MustFind(25289)

// A rank of the shout is the generated row on the rank's own spell, labelled with the rank so that two
// ranks on one unit never share an aura. Both bid in the row's category.
func battleShoutRankMeta(rank *spelldata.Spell) *Meta {
	m := *battleShoutMeta
	m.Label = rank.RankedName()
	m.Spell = rank
	return &m
}

// The attack power one rank grants, before the tier 2 set's bonus.
func BattleShoutRankValue(rank *spelldata.Spell) float64 {
	return battleShoutRankMeta(rank).Value(0)
}

func NewSpellDataBattleShoutRankAura(unit *core.Unit, isPlayer bool, rank *spelldata.Spell) *core.Aura {
	return newBuff(unit, battleShoutRankMeta(rank), isPlayer, 0)
}

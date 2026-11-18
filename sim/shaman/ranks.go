package shaman

import (
	"github.com/wowsims/forever/sim/core/spelldata"
)

// In game a rank deals less damage once the caster is far enough past the level it is taught at; at
// level 60 a rank taught at 44 or later is clear of that.
const minRankLevel = 44 // TODO: Ingame test needed: the cutoff is believed to be 42 or 43 at level 60.

// The ranks of a damage spell the shaman registers: those taught at minRankLevel or later.
func castableRanks(ladder spelldata.Ladder) spelldata.Ladder {
	var ids []int32
	ladder.Each(func(_ int32, rank *spelldata.Spell) {
		if rank.SpellLevel >= minRankLevel {
			ids = append(ids, rank.ID)
		}
	})
	return spelldata.Ranked(ids...)
}

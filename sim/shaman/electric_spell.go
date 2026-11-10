package shaman

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The rows Lightning Overload casts, one for each rank of Lightning Bolt and of Chain Lightning, in rank
// order. No effect of the game files names them.
var lightningBoltOverloadRanks = spelldata.Ranked(408439, 408440, 408441, 408442, 408443, 408472, 408473, 408474, 408475, 408477)
var chainLightningOverloadRanks = spelldata.Ranked(408479, 408481, 408482, 408484)

// The tag the results label an overload by.
const CastTagLightningOverload int32 = 6

// Each rank, with the overload row of the same rank number as the copy Lightning Overload casts. The
// copy reports under the spell id of the rank it copies, tagged CastTagLightningOverload.
func (shaman *Shaman) registerWithOverloads(ranks, overloads spelldata.Ladder, newConfig func(row *spelldata.Spell, overload *core.Spell) core.SpellConfig) {
	ranks.Each(func(_ int32, rank *spelldata.Spell) {
		overloadConfig := newConfig(overloads.Rank(rank.RankNumber()), nil)
		overloadConfig.ActionID = core.ActionID{SpellID: rank.ID, Tag: CastTagLightningOverload}
		overload := shaman.RegisterSpell(overloadConfig)
		shaman.RegisterSpell(newConfig(rank, overload))
	})
}

// Lightning Bolt and Chain Lightning, and the copy of either that Lightning Overload casts, with the
// average damage of one hit. The copy is registered from its own row, whose damage, coefficient and
// roll it deals, cast with no cast time or cost. Its row states no threat and lacks Not a Proc, so
// only listeners that can proc from procs hear its hits. As the tooltip of 408438 says, the copy's row
// states half the base points, points per level and coefficient of the rank it copies.
func (shaman *Shaman) newElectricSpellConfig(row *spelldata.Spell, isOverload bool) (core.SpellConfig, float64) {
	opts := []spelldata.SpellOpt{
		spelldata.Magic(core.ProcMaskSpellDamage),
		spelldata.Flags(SpellFlagShamanSpell | SpellFlagFocusable),
	}
	if isOverload {
		opts = append(opts, spelldata.Proc(), spelldata.Flags(core.SpellFlagProc))
	}

	return spelldata.SpellConfig(&shaman.Unit, row, opts...), row.DamageEffect().Average()
}

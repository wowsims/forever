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

// An overload deals half of the damage and the spell power coefficient of the rank it copies, as the
// tooltip of 408438 states.
//
// TODO: the overload rows (408439-408477, 408479-408484) have not moved with Lightning Bolt and Chain
// Lightning as of hotfix push 113596: Lightning Bolt rank 10's still states 98 base, half of the old
// rank's 196 against the new 353. So an overload reads the damage effect of the rank it copies, at
// this share.
const overloadShare = 0.5

// Each rank, with the overload row of the same rank number as the copy Lightning Overload casts. The
// copy reports under the spell id of the rank it copies, tagged CastTagLightningOverload.
func (shaman *Shaman) registerWithOverloads(ranks, overloads spelldata.Ladder, newConfig func(row, rank *spelldata.Spell, overload *core.Spell) core.SpellConfig) {
	ranks.Each(func(_ int32, rank *spelldata.Spell) {
		overloadConfig := newConfig(overloads.Rank(rank.RankNumber()), rank, nil)
		overloadConfig.ActionID = core.ActionID{SpellID: rank.ID, Tag: CastTagLightningOverload}
		overload := shaman.RegisterSpell(overloadConfig)
		shaman.RegisterSpell(newConfig(rank, rank, overload))
	})
}

// Lightning Bolt and Chain Lightning, and the copy of either that Lightning Overload casts, with the
// average damage of one hit. The copy is registered from its own row, cast with no cast time or cost.
// Its row states no threat and lacks Not a Proc, so only listeners that can proc from procs hear its
// hits. Its damage and coefficient are those of the rank it copies at overloadShare, and its roll is
// that rank's.
func (shaman *Shaman) newElectricSpellConfig(row, rank *spelldata.Spell) (core.SpellConfig, float64) {
	isOverload := row != rank
	opts := []spelldata.SpellOpt{
		spelldata.Magic(core.ProcMaskSpellDamage),
		spelldata.Flags(SpellFlagShamanSpell | SpellFlagFocusable),
	}
	if isOverload {
		opts = append(opts, spelldata.Proc(), spelldata.Flags(core.SpellFlagProc))
	}

	config := spelldata.SpellConfig(&shaman.Unit, row, opts...)
	damage := rank.DamageEffect()
	if !isOverload {
		return config, damage.Average()
	}
	config.BonusCoefficient = damage.Coeff() * overloadShare
	config.DirectVariance = damage.DirectVariance()
	return config, damage.Average() * overloadShare
}

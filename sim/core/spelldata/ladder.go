package spelldata

import (
	"fmt"
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// A spell's ranks in rank order, whether the game files state them as one spell per rank or as one
// spell with a curve. Rank 0 is untaken, so the readers below answer 0 there instead of panicking.
type Ladder struct {
	ranks []*Spell
}

// A ladder of one spell per rank, lowest first. Every id has to be in the store: a rank named by typo
// has to fail loudly rather than register nothing.
func Ranked(ids ...int32) Ladder {
	ranks := make([]*Spell, len(ids))
	for i, id := range ids {
		ranks[i] = MustFind(id)
	}
	return Ladder{ranks: ranks}
}

// A ladder of a trait-tree talent, which is one spell whose per-rank numbers live in a curve. Each
// rank is a copy of the spell with the curve's value on the effects the curve covers; an effect the
// curve has no row for keeps the spell's own base value.
func Talent(spellID int32, maxRanks int32) Ladder {
	base := MustFind(spellID)
	curve := curves[spellID]

	// A curve row states one value per rank. A row of another length means the curve and the talent
	// disagree about how many ranks there are, and the ranks past the shorter of the two would
	// silently keep the base value.
	for i, row := range curve {
		if int32(len(row)) != maxRanks {
			panic(fmt.Sprintf("spelldata: spell %d curve row %d has %d ranks, want %d",
				spellID, i, len(row), maxRanks))
		}
	}

	ranks := make([]*Spell, 0, maxRanks)
	for n := int32(1); n <= maxRanks; n++ {
		// Only the effects are copied, since only their base points differ by rank: the powers, the
		// labels and the tooltip references stay the store's own, and no rank writes through them.
		rank := *base
		rank.Effects = slices.Clone(base.Effects)
		for i := range rank.Effects {
			if i < len(curve) {
				rank.Effects[i].BasePoints = curve[i][n-1]
			}
		}
		ranks = append(ranks, &rank)
	}
	return Ladder{ranks: ranks}
}

// The spell at a rank. Rank 0 is untaken and answers Nil, as does a rank the ladder does not have.
func (l Ladder) Rank(n int32) *Spell {
	if n <= 0 || n > l.Len() {
		return Nil
	}
	return l.ranks[n-1]
}

func (l Ladder) Highest() *Spell {
	if len(l.ranks) == 0 {
		return Nil
	}
	return l.ranks[len(l.ranks)-1]
}

// The highest rank that is no class book rank, or Nil when the ladder has none.
func (l Ladder) HighestTrained() *Spell {
	for _, s := range slices.Backward(l.ranks) {
		if !IsBookRank(s.ID) {
			return s
		}
	}
	return Nil
}

// The rank with this spell id. Panics on an id the ladder does not carry, the way MustFind does on a
// spell the store does not carry.
func (l Ladder) ByID(id int32) *Spell {
	for _, s := range l.ranks {
		if s.ID == id {
			return s
		}
	}
	panic(fmt.Sprintf("spelldata: spell %d is not one of the %d ranks of this ladder", id, len(l.ranks)))
}

func (l Ladder) Len() int32 {
	return int32(len(l.ranks))
}

// Whether the spell was registered from one of the ladder's ranks.
func (l Ladder) Contains(spell *core.Spell) bool {
	return slices.ContainsFunc(l.ranks, func(s *Spell) bool { return s.ID == spell.SpellID })
}

func (l Ladder) Each(fn func(rank int32, s *Spell)) {
	for i, s := range l.ranks {
		fn(int32(i+1), s)
	}
}

// The rank's only effect, in the game files' units. Rank 0 is untaken and answers 0.
func (l Ladder) ValueAt(rank int32) float64 {
	return l.only().ValueAt(rank)
}

// The game files state a percentage as an integer: 16, not 0.16.
func (l Ladder) FractionAt(rank int32) float64 {
	return l.only().FractionAt(rank)
}

// The sign comes from the data: Improved Righteous Fury states -2/-4/-6, so rank 3 gives 0.94.
func (l Ladder) MultiplierAt(rank int32) float64 {
	return l.only().MultiplierAt(rank)
}

// The game files state rage on a 0-1000 bar.
func (l Ladder) TenthsAt(rank int32) float64 {
	return l.only().TenthsAt(rank)
}

// The effect at a position, counted from 1, for the ranks that carry more than one.
func (l Ladder) EffectAt(index int32) LadderEffect {
	return LadderEffect{ladder: l, pick: func(s *Spell, _ int32) *Effect {
		e := s.EffectN(int(index))
		if e == NilEffect {
			panic(fmt.Sprintf("spell %d has no effect at position %d, in %d effects",
				s.ID, index, len(s.Effects)))
		}
		return e
	}}
}

// The effect with this aura and misc value, which panics when the rank has two of them.
func (l Ladder) Effect(aura dbcenums.EffectAuraType, misc int32) LadderEffect {
	return LadderEffect{ladder: l, pick: func(s *Spell, _ int32) *Effect { return s.Effect(aura, misc) }}
}

// Nothing named means nothing to choose between - reading the first of several silently is the bug
// this shape exists to prevent.
func (l Ladder) only() LadderEffect {
	return LadderEffect{ladder: l, pick: func(s *Spell, rank int32) *Effect {
		if len(s.Effects) != 1 {
			panic(fmt.Sprintf("spell %d rank %d has %d effects - name the one you mean with Effect(aura, misc)",
				s.ID, rank, len(s.Effects)))
		}
		return &s.Effects[0]
	}}
}

// One named effect across the ladder's ranks.
type LadderEffect struct {
	ladder Ladder
	pick   func(s *Spell, rank int32) *Effect
}

func (e LadderEffect) ValueAt(rank int32) float64 {
	return e.at(rank).BasePoints
}

func (e LadderEffect) FractionAt(rank int32) float64 {
	return e.at(rank).Percent()
}

func (e LadderEffect) MultiplierAt(rank int32) float64 {
	return 1 + e.FractionAt(rank)
}

func (e LadderEffect) TenthsAt(rank int32) float64 {
	return e.at(rank).Tenths()
}

// The effect at a rank. Rank 0 is untaken and answers NilEffect, which reads as zero.
func (e LadderEffect) at(rank int32) *Effect {
	if rank <= 0 {
		return NilEffect
	}
	if rank > e.ladder.Len() {
		panic(fmt.Sprintf("rank %d in a ladder of %d ranks", rank, e.ladder.Len()))
	}
	return e.pick(e.ladder.Rank(rank), rank)
}

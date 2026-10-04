package spelldata

import (
	"fmt"

	"github.com/wowsims/forever/sim/core"
)

// The snippets docs/spell_data/ shows a porter, as runnable code: a doc that names an accessor the
// package no longer has, or a value it no longer answers, fails here.
//
// They read the generated store rather than the hand-built rows the rest of the package's tests run
// against, so the numbers are the game files' own. TestMain installs those rows for the whole package,
// so every example puts them back on its way out; the values themselves are the ones snapshot_test.go
// pins.
func generatedStore() func() {
	install(generatedSpells, generatedCurves, generatedHandTriggers)
	return func() {
		setSpells(fixture())
		curves = map[int32][][]float64{}
		handTriggers = map[int32][]int32{}
	}
}

// A row the store does not carry reads as zeroes rather than crashing the registration that asked
// for it.
func ExampleFind() {
	defer generatedStore()()

	frostbolt := Find(116)
	fmt.Println(frostbolt.Name, frostbolt.Rank)
	fmt.Println(Find(0) == Nil, Find(0).CastTime())
	// Output:
	// Frostbolt Rank 1
	// true 0s
}

// Effects are counted from 1 by position, not by the game files' EffectIndex, and the values come
// in the units the game files state them in.
func ExampleSpell_EffectN() {
	defer generatedStore()()

	frostbolt := MustFind(116)
	damage := frostbolt.EffectN(2)

	fmt.Println(damage.Index)
	fmt.Printf("%.0f %.0f %.3f\n", damage.BaseValue(), damage.AverageAt(60), damage.Coeff())
	fmt.Println(frostbolt.EffectN(3) == NilEffect)
	// Output:
	// 1
	// 19 21 0.407
	// true
}

// A cost is answered in the units the sim spends, and the school mask is core's own.
func ExampleSpell_PowerCost() {
	defer generatedStore()()

	frostbolt := MustFind(116)
	fmt.Println(frostbolt.PowerCost(0), frostbolt.SpellSchool() == core.SpellSchoolFrost, frostbolt.CastTime())
	// Output:
	// 25 true 1.5s
}

// What the row fills in a spell config, before the caller adds the effects and the flags the game
// files do not state.
func ExampleSpellConfig() {
	defer generatedStore()()

	config := SpellConfig(&core.Unit{}, MustFind(116), Magic(core.ProcMaskSpellDamage))

	fmt.Println(config.ActionID.SpellID, config.Rank, config.ManaCost.FlatCost)
	fmt.Println(config.Cast.DefaultCast.CastTime, config.Cast.DefaultCast.GCD)
	fmt.Printf("%.3f\n", config.BonusCoefficient)
	// Output:
	// 116 1 25
	// 1.5s 1.5s
	// 0.407
}

// The aura's duration and stack count are the row's. Lightning Shield states charges rather than
// cumulative stacks, and the sim keeps both in one field.
func ExampleAuraConfig() {
	defer generatedStore()()

	aura := AuraConfig(MustFind(324))
	fmt.Println(aura.Label, aura.Duration, aura.MaxStacks)
	// Output:
	// Lightning Shield 10m0s 3
}

// A ladder is the ranks of one family in rank order. Rank 0 is untaken and answers Nil, so no
// "if rank > 0" guard is needed.
func ExampleLadder() {
	defer generatedStore()()

	frostbolt := Ranked(116, 205, 837)
	fmt.Println(frostbolt.Len(), frostbolt.Rank(1).Rank, frostbolt.Highest().Rank)
	fmt.Println(frostbolt.Rank(0) == Nil)
	// Output:
	// 3 Rank 1 Rank 3
	// true
}

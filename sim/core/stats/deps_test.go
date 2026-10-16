package stats

import (
	"testing"
)

func TestStatDependencies(t *testing.T) {
	baseStat := Stats{
		Stamina:   1,
		Intellect: 1,
		Agility:   2,
		Spirit:    1,
	}

	sdm := NewStatDependencyManager()

	sdm.MultiplyStat(Intellect, 2)
	sdm.AddStatDependency(Stamina, Intellect, 1)
	sdm.MultiplyStat(Stamina, 2)
	sdm.AddStatDependency(Agility, Stamina, 1)

	dep1 := sdm.NewDynamicMultiplyStat(Agility, 2)
	dep2 := sdm.NewDynamicMultiplyStat(Spirit, 3)
	dep3 := sdm.NewDynamicStatDependency(Agility, Spirit, 0.75)

	sdm.FinalizeStatDeps()

	result := sdm.ApplyStatDependencies(baseStat)
	expectedResult := Stats{
		Stamina:   6,
		Intellect: 14,
		Agility:   2,
		Spirit:    1,
	}
	if !result.Equals(expectedResult) {
		t.Fatalf("Stats do not match:\nActual: %s\nExpected: %s", result, expectedResult)
	}

	sdm.EnableDynamicStatDep(dep1)
	result2 := sdm.ApplyStatDependencies(baseStat)
	expectedResult2 := Stats{
		Stamina:   10,
		Intellect: 22,
		Agility:   4,
		Spirit:    1,
	}
	if !result2.Equals(expectedResult2) {
		t.Fatalf("Updated stats do not match:\nActual: %s\nExpected: %s", result2, expectedResult2)
	}

	sdm.EnableDynamicStatDep(dep2)
	result3 := sdm.ApplyStatDependencies(baseStat)
	expectedResult3 := Stats{
		Stamina:   10,
		Intellect: 22,
		Agility:   4,
		Spirit:    3,
	}
	if !result3.Equals(expectedResult3) {
		t.Fatalf("Updated stats do not match:\nActual: %s\nExpected: %s", result3, expectedResult3)
	}

	sdm.DisableDynamicStatDep(dep2)
	result4 := sdm.ApplyStatDependencies(baseStat)
	if !result4.Equals(expectedResult2) {
		t.Fatalf("Updated stats do not match:\nActual: %s\nExpected: %s", result4, expectedResult2)
	}

	sdm.EnableDynamicStatDep(dep3)
	result5 := sdm.ApplyStatDependencies(baseStat)
	expectedResult5 := Stats{
		Stamina:   10,
		Intellect: 22,
		Agility:   4,
		Spirit:    4,
	}
	if !result5.Equals(expectedResult5) {
		t.Fatalf("Updated stats do not match:\nActual: %s\nExpected: %s", result5, expectedResult5)
	}

	sdm.DisableDynamicStatDep(dep3)
	result6 := sdm.ApplyStatDependencies(baseStat)
	if !result6.Equals(expectedResult2) {
		t.Fatalf("Updated stats do not match:\nActual: %s\nExpected: %s", result6, expectedResult2)
	}
}

func TestFlooredStatDepInPlace(t *testing.T) {
	sdm := NewStatDependencyManager()
	sdm.AddFlooredStatDependency(DefenseRating, ReducedCritTakenPercent, 2.5, 0.1)
	sdm.FinalizeStatDeps()

	baseStat := Stats{DefenseRating: 9}
	expectedResult := Stats{DefenseRating: 9, ReducedCritTakenPercent: 7.5 * 0.1}

	result := sdm.ApplyStatDependencies(baseStat)
	if !result.Equals(expectedResult) {
		t.Fatalf("Stats do not match:\nActual: %s\nExpected: %s", result, expectedResult)
	}

	inPlace := baseStat
	sdm.ApplyStatDependenciesInPlace(&inPlace)
	if !inPlace.Equals(expectedResult) {
		t.Fatalf("In-place stats do not match:\nActual: %s\nExpected: %s", inPlace, expectedResult)
	}
}

func TestMultipleStatDep(t *testing.T) {
	sdm := NewStatDependencyManager()

	baseStat := Stats{
		Intellect:   100,
		SpellDamage: 100,
	}

	sdm.AddStatDependency(Intellect, SpellDamage, 0.2)
	sdm.AddStatDependency(Intellect, SpellDamage, 0.2)
	sdm.MultiplyStat(Intellect, 1.2)
	sdm.FinalizeStatDeps()
	result := sdm.ApplyStatDependencies(baseStat)

	expectedResult := Stats{
		Intellect:   100 * 1.2,
		SpellDamage: 100 + (100*1.2)*(0.2+0.2),
	}

	if !result.Equals(expectedResult) {
		t.Fatalf("Stats do not match:\nActual: %s\nExpected: %s", result, expectedResult)
	}
}

// The doubling scales the 200 armor and adds the 100 ArmorNoMods after it, whichever was registered
// first, and a dependency reading armor counts all of it.
func TestUnmultipliedStatDependencyStaysOutOfTheSameStatMultiplier(t *testing.T) {
	for _, unmultipliedFirst := range []bool{true, false} {
		sdm := NewStatDependencyManager()
		if unmultipliedFirst {
			sdm.AddUnmultipliedStatDependency(ArmorNoMods, Armor)
			sdm.MultiplyStat(Armor, 2)
		} else {
			sdm.MultiplyStat(Armor, 2)
			sdm.AddUnmultipliedStatDependency(ArmorNoMods, Armor)
		}
		sdm.AddStatDependency(Armor, AttackPower, 0.1)
		sdm.FinalizeStatDeps()

		got := sdm.ApplyStatDependencies(Stats{Armor: 200, ArmorNoMods: 100})
		if got[Armor] != 500 {
			t.Errorf("ArmorNoMods registered first %v: armor %v, want 200*2+100 = 500", unmultipliedFirst, got[Armor])
		}
		if got[AttackPower] != 50 {
			t.Errorf("ArmorNoMods registered first %v: attack power from armor %v, want 500*0.1 = 50", unmultipliedFirst, got[AttackPower])
		}
	}
}

// The dependencies give a change of stats as they give whole stats: under the doubling, a change of
// 300 armor is 600 and a change of 100 ArmorNoMods is 100 armor.
func TestUnmultipliedStatDependencyAppliesToAChange(t *testing.T) {
	sdm := NewStatDependencyManager()
	sdm.MultiplyStat(Armor, 2)
	sdm.AddUnmultipliedStatDependency(ArmorNoMods, Armor)
	sdm.FinalizeStatDeps()

	if got := sdm.ApplyStatDependencies(Stats{Armor: 300})[Armor]; got != 600 {
		t.Errorf("a change of 300 armor: %v, want 300*2 = 600", got)
	}
	if got := sdm.ApplyStatDependencies(Stats{ArmorNoMods: 100})[Armor]; got != 100 {
		t.Errorf("a change of 100 ArmorNoMods: %v armor, want all 100", got)
	}
}

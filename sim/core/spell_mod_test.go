package core

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/stats"
)

const fakeThreatSpellClassMask int64 = 1 << 0
const fakeGCDSpellClassMask int64 = 1 << 1

var fakeThreatSpellActionID = ActionID{SpellID: 11597}
var fakeGCDSpellActionID = ActionID{SpellID: 16886}

func (fw *FakeRageWarrior) registerFakeThreatSpell() {
	fw.RegisterSpell(SpellConfig{
		ActionID:        fakeThreatSpellActionID,
		ClassSpellMask:  fakeThreatSpellClassMask,
		FlatThreatBonus: 100,
	})
}

func (fw *FakeRageWarrior) registerFakeGCDSpell() {
	fw.RegisterSpell(SpellConfig{
		ActionID:       fakeGCDSpellActionID,
		ClassSpellMask: fakeGCDSpellClassMask,
		Cast: CastConfig{
			DefaultCast: Cast{GCD: GCDDefault},
		},
	})
}

// Casts the spell with the GCD free and answers the GCD the cast started, before and after the
// GCD minimum.
func castGCD(t *testing.T, sim *Simulation, fw *FakeRageWarrior, spell *Spell) (time.Duration, time.Duration) {
	fw.GCD.Reset()
	if !spell.Cast(sim, sim.Encounter.ActiveTargetUnits[0]) {
		t.Fatal("the spell did not cast")
	}
	return spell.CurCast.GCD, spell.CurCast.GCDTime()
}

func TestGlobalCooldownPctMod(t *testing.T) {
	sim := SetupFakeRageSim()
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	spell := fw.GetSpell(fakeGCDSpellActionID)

	pct := fw.AddDynamicMod(SpellModConfig{
		ClassMask:  fakeGCDSpellClassMask,
		Kind:       SpellMod_GlobalCooldown_Pct,
		FloatValue: -0.1,
	})
	flat := fw.AddDynamicMod(SpellModConfig{
		ClassMask: fakeGCDSpellClassMask,
		Kind:      SpellMod_GlobalCooldown_Flat,
		TimeValue: -400 * time.Millisecond,
	})

	steps := []struct {
		name    string
		toggle  func()
		gcd     time.Duration
		gcdTime time.Duration
	}{
		{"without a mod", func() {}, 1500 * time.Millisecond, 1500 * time.Millisecond},
		{"with the percent mod", pct.Activate, 1350 * time.Millisecond, 1350 * time.Millisecond},
		// (1500 - 400) x 0.9 = 990, which the GCD minimum floors at a second.
		{"with both mods", flat.Activate, 990 * time.Millisecond, time.Second},
		{"with the flat mod alone", pct.Deactivate, 1100 * time.Millisecond, 1100 * time.Millisecond},
		{"after both were removed", flat.Deactivate, 1500 * time.Millisecond, 1500 * time.Millisecond},
	}

	for _, step := range steps {
		step.toggle()
		gcd, gcdTime := castGCD(t, sim, fw, spell)
		if gcd != step.gcd || gcdTime != step.gcdTime {
			t.Errorf("%s: GCD %v floored to %v, want %v floored to %v", step.name, gcd, gcdTime, step.gcd, step.gcdTime)
		}
	}

	if spell.GCDMultiplier != 1 {
		t.Errorf("GCD multiplier after both mods were removed: %v, want 1", spell.GCDMultiplier)
	}
}

// A -100% cooldown mod, such as Berserk's on Primal Bite, removes the cooldown while it is up and gives it
// back once it is gone, beside a mod that only shortens it.
func TestCooldownMultiplierModOfZero(t *testing.T) {
	sim := SetupFakeRageSim()
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	spell := fw.GetSpell(fakeGCDSpellActionID)

	removed := fw.AddDynamicMod(SpellModConfig{
		ClassMask:  fakeGCDSpellClassMask,
		Kind:       SpellMod_Cooldown_Multiplier,
		FloatValue: 0,
	})
	halved := fw.AddDynamicMod(SpellModConfig{
		ClassMask:  fakeGCDSpellClassMask,
		Kind:       SpellMod_Cooldown_Multiplier,
		FloatValue: 0.5,
	})

	steps := []struct {
		name   string
		toggle func()
		want   float64
	}{
		{"without a mod", func() {}, 1},
		{"with the halving mod", halved.Activate, 0.5},
		{"with both mods", removed.Activate, 0},
		{"after the removing mod is gone", removed.Deactivate, 0.5},
		{"after both are gone", halved.Deactivate, 1},
	}
	for _, step := range steps {
		step.toggle()
		if got := spell.CooldownMultiplier(); got != step.want {
			t.Errorf("%s: cooldown multiplier %v, want %v", step.name, got, step.want)
		}
	}
}

// A cooldown mod below -100% removes the cooldown like one of -100%, and gives it back once it is gone.
func TestCooldownMultiplierModBelowZero(t *testing.T) {
	sim := SetupFakeRageSim()
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	spell := fw.GetSpell(fakeGCDSpellActionID)

	removed := fw.AddDynamicMod(SpellModConfig{
		ClassMask:  fakeGCDSpellClassMask,
		Kind:       SpellMod_Cooldown_Multiplier,
		FloatValue: -0.5,
	})
	removed.Activate()
	if got := spell.CooldownMultiplier(); got != 0 {
		t.Errorf("with a -150%% mod: cooldown multiplier %v, want 0", got)
	}
	removed.Deactivate()
	if got := spell.CooldownMultiplier(); got != 1 {
		t.Errorf("after the -150%% mod is gone: cooldown multiplier %v, want 1", got)
	}
}

// A flat reduction longer than the cooldown, such as Raiments of Conviction's 60 s off Devouring
// Plague's 24 s, stops the cooldown at 0, and removing the mods gives back exactly what they took.
func TestCooldownFlatModStopsAtZero(t *testing.T) {
	sim := SetupFakeRageSim()
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	spell := fw.GetSpell(fakeGCDSpellActionID)
	spell.CD.Duration = 24 * time.Second

	removed := fw.AddDynamicMod(SpellModConfig{
		ClassMask: fakeGCDSpellClassMask,
		Kind:      SpellMod_Cooldown_Flat,
		TimeValue: -60 * time.Second,
	})
	shortened := fw.AddDynamicMod(SpellModConfig{
		ClassMask: fakeGCDSpellClassMask,
		Kind:      SpellMod_Cooldown_Flat,
		TimeValue: -10 * time.Second,
	})

	steps := []struct {
		name   string
		toggle func()
		want   time.Duration
	}{
		{"with the 60 s mod", removed.Activate, 0},
		{"with both mods", shortened.Activate, 0},
		{"after the 10 s mod is gone", shortened.Deactivate, 0},
		{"after both are gone", removed.Deactivate, 24 * time.Second},
		{"with the 10 s mod", shortened.Activate, 14 * time.Second},
		{"with both mods again", removed.Activate, 0},
		{"after the 60 s mod is gone", removed.Deactivate, 14 * time.Second},
		{"after both are gone again", shortened.Deactivate, 24 * time.Second},
	}
	for _, step := range steps {
		step.toggle()
		if got := spell.CD.Duration; got != step.want {
			t.Errorf("%s: cooldown %v, want %v", step.name, got, step.want)
		}
	}
}

func TestFlatThreatBonusPctMod(t *testing.T) {
	sim := SetupFakeRageSim()
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	spell := fw.GetSpell(fakeThreatSpellActionID)
	attackTable := fw.AttackTables[sim.Encounter.ActiveTargetUnits[0].UnitIndex]

	// 100 flat + 0.05 x the 1000 attack power the fake warrior holds.
	spell.ThreatAPCoefficient = 0.05
	fw.AddStatDynamic(sim, stats.AttackPower, 1000)

	threat := spell.ThreatFromDamage(sim, OutcomeHit, 0, attackTable)
	if !WithinToleranceFloat64(150, threat, 0.001) {
		t.Fatalf("Incorrect threat without a mod: Expected: %0.3f, Actual: %0.3f", 150.0, threat)
	}

	mod := fw.AddDynamicMod(SpellModConfig{
		ClassMask:  fakeThreatSpellClassMask,
		Kind:       SpellMod_FlatThreatBonus_Pct,
		FloatValue: 0.15,
	})

	mod.Activate()
	threat = spell.ThreatFromDamage(sim, OutcomeHit, 0, attackTable)
	if !WithinToleranceFloat64(172.5, threat, 0.001) {
		t.Fatalf("Incorrect threat with the mod active: Expected: %0.3f, Actual: %0.3f", 172.5, threat)
	}

	mod.Deactivate()
	threat = spell.ThreatFromDamage(sim, OutcomeHit, 0, attackTable)
	if !WithinToleranceFloat64(150, threat, 0.001) {
		t.Fatalf("Incorrect threat after the mod was removed: Expected: %0.3f, Actual: %0.3f", 150.0, threat)
	}
}

// A mod that names its spells by the game files' class flags reaches exactly those spells: the
// family has to agree and the masks have to overlap.
func TestClassFlagsModMatching(t *testing.T) {
	mod := &SpellMod{ClassFlags: ClassFlags{Family: 4, Mask: [4]uint32{0b0110}}}

	tests := []struct {
		name  string
		spell *Spell
		want  bool
	}{
		{"overlapping mask", &Spell{ClassFlags: ClassFlags{Family: 4, Mask: [4]uint32{0b0100}}}, true},
		{"disjoint mask in the same family", &Spell{ClassFlags: ClassFlags{Family: 4, Mask: [4]uint32{0b1000}}}, false},
		{"same mask in another family", &Spell{ClassFlags: ClassFlags{Family: 9, Mask: [4]uint32{0b0110}}}, false},
		{"overlap in a later word", &Spell{ClassFlags: ClassFlags{Family: 4, Mask: [4]uint32{0, 1}}}, false},
		{"no class flags at all", &Spell{}, false},
	}

	for _, test := range tests {
		if got := shouldApply(test.spell, mod); got != test.want {
			t.Errorf("%s: shouldApply = %v, want %v", test.name, got, test.want)
		}
	}
}

// The two masks are separate filters, so a mod that sets both applies only where both name the
// spell.
func TestModWithBothMasksRequiresBoth(t *testing.T) {
	mod := &SpellMod{
		ClassMask:  1 << 3,
		ClassFlags: ClassFlags{Family: 4, Mask: [4]uint32{0b0001}},
	}

	tests := []struct {
		name  string
		spell *Spell
		want  bool
	}{
		{"both", &Spell{ClassSpellMask: 1 << 3, ClassFlags: ClassFlags{Family: 4, Mask: [4]uint32{0b0001}}}, true},
		{"only the sim tag", &Spell{ClassSpellMask: 1 << 3}, false},
		{"only the class flags", &Spell{ClassFlags: ClassFlags{Family: 4, Mask: [4]uint32{0b0001}}}, false},
	}

	for _, test := range tests {
		if got := shouldApply(test.spell, mod); got != test.want {
			t.Errorf("%s: shouldApply = %v, want %v", test.name, got, test.want)
		}
	}
}

// A mod that names no spells at all still applies to every spell, so adding the class flags
// filter cannot change any existing mod.
func TestModWithoutMasksAppliesToEverything(t *testing.T) {
	mod := &SpellMod{}
	if !shouldApply(&Spell{ClassFlags: ClassFlags{Family: 4, Mask: [4]uint32{1}}}, mod) {
		t.Error("an unfiltered mod should apply to a spell that carries class flags")
	}
}

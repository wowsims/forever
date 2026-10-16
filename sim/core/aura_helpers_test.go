package core

import (
	"testing"
	"time"
)

// A proc trigger that names its spells by the game files' class flags fires on exactly those spells.
func TestProcTriggerClassFlags(t *testing.T) {
	config := ProcTrigger{ClassFlags: ClassFlags{Family: 8, Mask: [4]uint32{0b0011}}}

	tests := []struct {
		name  string
		spell *Spell
		want  bool
	}{
		{"overlapping mask", &Spell{ClassFlags: ClassFlags{Family: 8, Mask: [4]uint32{0b0010}}}, true},
		{"disjoint mask in the same family", &Spell{ClassFlags: ClassFlags{Family: 8, Mask: [4]uint32{0b0100}}}, false},
		{"same mask in another family", &Spell{ClassFlags: ClassFlags{Family: 3, Mask: [4]uint32{0b0011}}}, false},
		{"no class flags at all", &Spell{}, false},
	}

	for _, test := range tests {
		if got := config.matchesSpell(test.spell); got != test.want {
			t.Errorf("%s: matchesSpell = %v, want %v", test.name, got, test.want)
		}
	}
}

// A trigger that sets both masks fires only on the spells both name.
func TestProcTriggerWithBothMasksRequiresBoth(t *testing.T) {
	config := ProcTrigger{
		ClassSpellMask: 1 << 2,
		ClassFlags:     ClassFlags{Family: 8, Mask: [4]uint32{0b0001}},
	}

	tests := []struct {
		name  string
		spell *Spell
		want  bool
	}{
		{"both", &Spell{ClassSpellMask: 1 << 2, ClassFlags: ClassFlags{Family: 8, Mask: [4]uint32{0b0001}}}, true},
		{"only the sim tag", &Spell{ClassSpellMask: 1 << 2}, false},
		{"only the class flags", &Spell{ClassFlags: ClassFlags{Family: 8, Mask: [4]uint32{0b0001}}}, false},
	}

	for _, test := range tests {
		if got := config.matchesSpell(test.spell); got != test.want {
			t.Errorf("%s: matchesSpell = %v, want %v", test.name, got, test.want)
		}
	}
}

// Only Proc From Class Abilities asks whether the spell is a class ability at all, which either
// mask answers.
func TestProcTriggerClassSpellsOnly(t *testing.T) {
	config := ProcTrigger{ClassSpellsOnly: true}

	tests := []struct {
		name  string
		spell *Spell
		want  bool
	}{
		{"tagged by the sim", &Spell{ClassSpellMask: 1 << 5}, true},
		{"named by the game files' class flags", &Spell{ClassFlags: ClassFlags{Family: 8, Mask: [4]uint32{1}}}, true},
		{"neither", &Spell{}, false},
	}

	for _, test := range tests {
		if got := config.matchesSpell(test.spell); got != test.want {
			t.Errorf("%s: matchesSpell = %v, want %v", test.name, got, test.want)
		}
	}
}

// A proc cooldown mod scales the internal cooldown a proc starts on the triggers whose own class flags
// it names, and leaves every other trigger and every proc chance alone.
func TestProcCooldownModScalesTheICDsItNames(t *testing.T) {
	omen := ClassFlags{Family: 7, Mask: [4]uint32{3: 0x40000}}
	other := ClassFlags{Family: 7, Mask: [4]uint32{0: 0x4}}
	const icd = 10 * time.Second

	startedICD := func(unit *Unit, flags ClassFlags) time.Duration {
		sim := &Simulation{rand: NewSplitMix(1)}
		aura := &Aura{Unit: unit}
		aura.AttachProcTriggerCallback(unit, ProcTrigger{
			Callback:            CallbackOnSpellHitDealt,
			TriggerImmediately:  true,
			ICD:                 icd,
			ProcSpellClassFlags: flags,
			Handler:             func(_ *Simulation, _ *Spell, _ *SpellResult) {},
		})
		aura.OnSpellHitDealt(aura, sim, &Spell{ProcMask: ProcMaskMeleeMH}, &SpellResult{Outcome: OutcomeHit})
		return aura.Icd.TimeToReady(sim)
	}

	unit := &Unit{}
	mod := unit.AddDynamicMod(SpellModConfig{Kind: SpellMod_ProcCooldown_Pct, ClassFlags: omen, FloatValue: -0.5})
	if got := startedICD(unit, omen); got != icd {
		t.Errorf("the mod turned off: ICD %v, want %v", got, icd)
	}

	mod.Activate()
	for _, test := range []struct {
		name  string
		unit  *Unit
		flags ClassFlags
		want  time.Duration
	}{
		{"a trigger the mod names", unit, omen, icd / 2},
		{"a pet's trigger", &Unit{procChanceOwner: unit}, omen, icd / 2},
		{"a trigger of other class flags", unit, other, icd},
		{"a trigger stating no class flags", unit, ClassFlags{}, icd},
	} {
		if got := startedICD(test.unit, test.flags); got != test.want {
			t.Errorf("%s: ICD %v, want %v", test.name, got, test.want)
		}
	}
	if got := unit.ModifiedProcChance(omen, 0.2); got != 0.2 {
		t.Errorf("a proc cooldown mod changed the proc chance: %v, want 0.2", got)
	}
}

// A proc chance mod raises the roll of the triggers whose own class flags it names, on the chance
// the trigger states and on the one its manager states, and leaves every other trigger alone.
func TestProcChanceModReachesTheTriggersItNames(t *testing.T) {
	poison := ClassFlags{Family: 8, Mask: [4]uint32{0: 0x2000}}
	other := ClassFlags{Family: 8, Mask: [4]uint32{0: 0x4}}
	const almostNever = 1e-6

	fires := func(unit *Unit, config ProcTrigger) int {
		sim := &Simulation{rand: NewSplitMix(1)}
		aura := &Aura{Unit: unit}
		count := 0
		config.Callback = CallbackOnSpellHitDealt
		config.TriggerImmediately = true
		config.Handler = func(_ *Simulation, _ *Spell, _ *SpellResult) { count++ }
		aura.AttachProcTriggerCallback(unit, config)
		for range 1000 {
			aura.OnSpellHitDealt(aura, sim, &Spell{ProcMask: ProcMaskMeleeMH}, &SpellResult{Outcome: OutcomeHit})
		}
		return count
	}
	manager := func(chance float64) *DynamicProcManager {
		return &DynamicProcManager{procMasks: []ProcMask{ProcMaskMelee}, procChances: []DynamicProc{staticProc{chance: chance}}}
	}

	unit := &Unit{}
	flat := unit.AddDynamicMod(SpellModConfig{Kind: SpellMod_ProcChance_Flat, ClassFlags: poison, FloatValue: 100})

	if got := fires(unit, ProcTrigger{ProcChance: almostNever, ProcSpellClassFlags: poison}); got != 0 {
		t.Errorf("the mod turned off: fired %d times, want 0", got)
	}

	flat.Activate()
	for _, test := range []struct {
		name   string
		config ProcTrigger
		want   int
	}{
		{"a stated chance the mod names", ProcTrigger{ProcChance: almostNever, ProcSpellClassFlags: poison}, 1000},
		{"a manager's chance the mod names", ProcTrigger{DPM: manager(almostNever), ProcSpellClassFlags: poison}, 1000},
		{"a trigger of other class flags", ProcTrigger{ProcChance: almostNever, ProcSpellClassFlags: other}, 0},
		{"a trigger stating no class flags", ProcTrigger{ProcChance: almostNever}, 0},
	} {
		if got := fires(unit, test.config); got != test.want {
			t.Errorf("%s: fired %d times, want %d", test.name, got, test.want)
		}
	}

	// A pet's trigger reads its owner's mods, and a trigger on the target reads the ones of the unit
	// it names.
	pet := &Unit{procChanceOwner: unit}
	target := &Unit{}
	for _, test := range []struct {
		name   string
		unit   *Unit
		config ProcTrigger
		want   int
	}{
		{"a pet's trigger", pet, ProcTrigger{ProcChance: almostNever, ProcSpellClassFlags: poison}, 1000},
		{"a trigger on the target naming the caster", target,
			ProcTrigger{ProcChance: almostNever, ProcSpellClassFlags: poison, Caster: unit}, 1000},
		{"a trigger on the target naming no one", target, ProcTrigger{ProcChance: almostNever, ProcSpellClassFlags: poison}, 0},
	} {
		if got := fires(test.unit, test.config); got != test.want {
			t.Errorf("%s: fired %d times, want %d", test.name, got, test.want)
		}
	}

	// The percentage multiplies the chance once the flat points are in.
	flat.UpdateFloatValue(10)
	pct := unit.AddDynamicMod(SpellModConfig{Kind: SpellMod_ProcChance_Pct, ClassFlags: poison, FloatValue: 1})
	pct.Activate()
	if got := unit.ModifiedProcChance(poison, 0.2); got < 0.6-1e-9 || got > 0.6+1e-9 {
		t.Errorf("20%% with +10 points and +100%%: %v, want 0.6", got)
	}
}

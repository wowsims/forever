package core

import (
	"testing"
	"time"
)

// A mod built straight from the kind's registered functions, so Activate, Deactivate and the
// UpdateXValue paths run exactly as they do on a unit's mod.
func newTestMod(config SpellModConfig, spells ...*Spell) *SpellMod {
	functions := spellModMap[config.Kind]

	return &SpellMod{
		Kind:           config.Kind,
		floatValue:     config.FloatValue,
		intValue:       config.IntValue,
		timeValue:      config.TimeValue,
		keyValue:       config.KeyValue,
		Apply:          functions.Apply,
		Remove:         functions.Remove,
		OnReset:        functions.OnReset,
		AffectedSpells: spells,
	}
}

// A spell carrying one of each duration target: a dot, an aoe dot, a self-buff and an aura array.
func newDurationTargetSpell() *Spell {
	spell := &Spell{
		RelatedSelfBuff:   &Aura{Duration: time.Second * 10, MaxStacks: 3},
		RelatedAuraArrays: LabeledAuraArrays{"debuff": AuraArray{{Duration: time.Second * 12, MaxStacks: 2}, nil}},
	}
	spell.dots = DotArray{{Aura: &Aura{}, BaseTickCount: 6, BaseTickLength: time.Second * 3, BaseDurationMultiplier: 1}, nil}
	spell.aoeDot = &Dot{Aura: &Aura{}, BaseTickCount: 4, BaseTickLength: time.Second * 2, BaseDurationMultiplier: 1}

	return spell
}

func TestDurationFlatMod(t *testing.T) {
	spell := newDurationTargetSpell()
	mod := newTestMod(SpellModConfig{Kind: SpellMod_Duration_Flat, TimeValue: time.Second * 6}, spell)

	mod.Activate()
	if got := spell.dots[0].BaseDuration(); got != time.Second*24 {
		t.Errorf("dot duration with the mod active: %v, want %v", got, time.Second*24)
	}
	if got := spell.aoeDot.BaseDuration(); got != time.Second*14 {
		t.Errorf("aoe dot duration with the mod active: %v, want %v", got, time.Second*14)
	}
	if got := spell.RelatedSelfBuff.Duration; got != time.Second*16 {
		t.Errorf("self-buff duration with the mod active: %v, want %v", got, time.Second*16)
	}
	if got := spell.RelatedAuraArrays["debuff"][0].Duration; got != time.Second*18 {
		t.Errorf("aura array duration with the mod active: %v, want %v", got, time.Second*18)
	}

	mod.Deactivate()
	if got := spell.dots[0].BaseDuration(); got != time.Second*18 {
		t.Errorf("dot duration after removing the mod: %v, want %v", got, time.Second*18)
	}
	if got := spell.aoeDot.BaseDuration(); got != time.Second*8 {
		t.Errorf("aoe dot duration after removing the mod: %v, want %v", got, time.Second*8)
	}
	if got := spell.RelatedSelfBuff.Duration; got != time.Second*10 {
		t.Errorf("self-buff duration after removing the mod: %v, want %v", got, time.Second*10)
	}
	if got := spell.RelatedAuraArrays["debuff"][0].Duration; got != time.Second*12 {
		t.Errorf("aura array duration after removing the mod: %v, want %v", got, time.Second*12)
	}
}

// The extra duration turns into whole extra ticks when the dot is next applied.
func TestDurationFlatModAddsTicks(t *testing.T) {
	sim := SetupFakeSim()
	fa := sim.Raid.Parties[0].Players[0].(*FakeAgent)

	fa.Dot.Apply(sim)
	if got := fa.Dot.RemainingTicks(); got != 6 {
		t.Fatalf("ticks without the mod: %d, want 6", got)
	}

	mod := newTestMod(SpellModConfig{Kind: SpellMod_Duration_Flat, TimeValue: time.Second * 6}, fa.Spell)
	mod.Activate()

	fa.Dot.Deactivate(sim)
	fa.Dot.Apply(sim)
	if got := fa.Dot.RemainingTicks(); got != 8 {
		t.Errorf("ticks with the mod active: %d, want 8", got)
	}
}

// A spell with none of the duration targets is left alone rather than panicking.
func TestDurationFlatModOnSpellWithoutTargets(t *testing.T) {
	spell := &Spell{}
	mod := newTestMod(SpellModConfig{Kind: SpellMod_Duration_Flat, TimeValue: time.Second * 6}, spell)

	mod.Activate()
	mod.Deactivate()
}

func TestBuffMaxStacksFlatMod(t *testing.T) {
	spell := newDurationTargetSpell()
	// An aura the game files never give stacks, which the mod has to leave non-stacking.
	spell.RelatedAuraArrays["debuff"] = append(spell.RelatedAuraArrays["debuff"], &Aura{})

	mod := newTestMod(SpellModConfig{Kind: SpellMod_BuffMaxStacks_Flat, IntValue: 2}, spell)

	mod.Activate()
	if got := spell.RelatedSelfBuff.MaxStacks; got != 5 {
		t.Errorf("self-buff max stacks with the mod active: %d, want 5", got)
	}
	if got := spell.RelatedAuraArrays["debuff"][0].MaxStacks; got != 4 {
		t.Errorf("aura array max stacks with the mod active: %d, want 4", got)
	}
	if got := spell.RelatedAuraArrays["debuff"][2].MaxStacks; got != 0 {
		t.Errorf("non-stacking aura max stacks with the mod active: %d, want 0", got)
	}

	mod.UpdateIntValue(4)
	if got := spell.RelatedSelfBuff.MaxStacks; got != 7 {
		t.Errorf("self-buff max stacks after updating the mod: %d, want 7", got)
	}

	mod.Deactivate()
	if got := spell.RelatedSelfBuff.MaxStacks; got != 3 {
		t.Errorf("self-buff max stacks after removing the mod: %d, want 3", got)
	}
	if got := spell.RelatedAuraArrays["debuff"][0].MaxStacks; got != 2 {
		t.Errorf("aura array max stacks after removing the mod: %d, want 2", got)
	}
}

func TestRangeFlatMod(t *testing.T) {
	spell := &Spell{MaxRange: 30}
	mod := newTestMod(SpellModConfig{Kind: SpellMod_Range_Flat, FloatValue: 5}, spell)

	mod.Activate()
	if spell.MaxRange != 35 {
		t.Errorf("max range with the mod active: %v, want 35", spell.MaxRange)
	}

	mod.UpdateFloatValue(8)
	if spell.MaxRange != 38 {
		t.Errorf("max range after updating the mod: %v, want 38", spell.MaxRange)
	}

	mod.Deactivate()
	if spell.MaxRange != 30 {
		t.Errorf("max range after removing the mod: %v, want 30", spell.MaxRange)
	}
}

func TestRangePctMod(t *testing.T) {
	spell := &Spell{MaxRange: 30, RangeMultiplier: 1}
	mod := newTestMod(SpellModConfig{Kind: SpellMod_Range_Pct, FloatValue: 0.1}, spell)

	mod.Activate()
	if !WithinToleranceFloat64(1.1, spell.RangeMultiplier, 1e-9) {
		t.Errorf("range multiplier with the mod active: %v, want 1.1", spell.RangeMultiplier)
	}

	mod.UpdateFloatValue(0.2)
	if !WithinToleranceFloat64(1.2, spell.RangeMultiplier, 1e-9) {
		t.Errorf("range multiplier after updating the mod: %v, want 1.2", spell.RangeMultiplier)
	}

	mod.Deactivate()
	if !WithinToleranceFloat64(1, spell.RangeMultiplier, 1e-9) {
		t.Errorf("range multiplier after removing the mod: %v, want 1", spell.RangeMultiplier)
	}
}

// The flat mod moves MaxRange and the percentage one the multiplier, so the two leave the spell as
// they found it whatever order they come and go in.
func TestRangeFlatAndPctModsRemoveInAnyOrder(t *testing.T) {
	spell := &Spell{MaxRange: 30, RangeMultiplier: 1}
	pct := newTestMod(SpellModConfig{Kind: SpellMod_Range_Pct, FloatValue: 0.1}, spell)
	flat := newTestMod(SpellModConfig{Kind: SpellMod_Range_Flat, FloatValue: 5}, spell)

	pct.Activate()
	flat.Activate()
	if got := spell.MaxRange * spell.RangeMultiplier; !WithinToleranceFloat64(38.5, got, 1e-9) {
		t.Errorf("range with both mods active: %v, want 38.5", got)
	}

	pct.Deactivate()
	flat.Deactivate()
	if spell.MaxRange != 30 || !WithinToleranceFloat64(1, spell.RangeMultiplier, 1e-9) {
		t.Errorf("range after removing both mods: %v x %v, want 30 x 1", spell.MaxRange, spell.RangeMultiplier)
	}
}

// The direct-only bucket reaches direct hits and leaves ticks at the shared additive bucket.
func TestDirectDamageDoneFlatMod(t *testing.T) {
	sim := SetupFakeSim()
	fa := sim.Raid.Parties[0].Players[0].(*FakeAgent)
	attackTable := fa.AttackTables[sim.Encounter.ActiveTargetUnits[0].UnitIndex]

	mod := newTestMod(SpellModConfig{Kind: SpellMod_DirectDamageDone_Flat, FloatValue: 0.2}, fa.Spell)
	mod.Activate()

	if got := fa.Spell.AttackerDamageMultiplier(attackTable, false); !WithinToleranceFloat64(1.8, got, 0.001) {
		t.Errorf("direct multiplier with the mod active: %0.3f, want 1.800", got)
	}
	if got := fa.Spell.AttackerDamageMultiplier(attackTable, true); !WithinToleranceFloat64(1.5, got, 0.001) {
		t.Errorf("tick multiplier with the mod active: %0.3f, want 1.500", got)
	}

	mod.UpdateFloatValue(0.4)
	if got := fa.Spell.AttackerDamageMultiplier(attackTable, false); !WithinToleranceFloat64(2.1, got, 0.001) {
		t.Errorf("direct multiplier after updating the mod: %0.3f, want 2.100", got)
	}

	mod.Deactivate()
	if got := fa.Spell.AttackerDamageMultiplier(attackTable, false); !WithinToleranceFloat64(1.5, got, 0.001) {
		t.Errorf("direct multiplier after removing the mod: %0.3f, want 1.500", got)
	}
}

// SPELLMOD_DAMAGE raises healing done as well, so a direct heal reads the bucket and a hot does not.
func TestDirectDamageDoneFlatModReachesDirectHeals(t *testing.T) {
	sim := SetupFakeSim()
	fa := sim.Raid.Parties[0].Players[0].(*FakeAgent)

	mod := newTestMod(SpellModConfig{Kind: SpellMod_DirectDamageDone_Flat, FloatValue: 0.2}, fa.Spell)
	mod.Activate()

	if got := fa.Spell.CasterHealingMultiplier(); !WithinToleranceFloat64(1.8, got, 0.001) {
		t.Errorf("direct heal multiplier with the mod active: %0.3f, want 1.800", got)
	}
	if got := fa.Spell.casterHealingMultiplier(true); !WithinToleranceFloat64(1.5, got, 0.001) {
		t.Errorf("hot multiplier with the mod active: %0.3f, want 1.500", got)
	}
}

// Stacks the mod took to 0 could not come back: the aura would be skipped as one that never stacked.
func TestBuffMaxStacksFlatModRefusesToEmptyTheStacks(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("a mod taking the stacks to 0 did not panic")
		}
	}()

	spell := &Spell{RelatedSelfBuff: &Aura{Label: "Test Buff", MaxStacks: 2}}
	newTestMod(SpellModConfig{Kind: SpellMod_BuffMaxStacks_Flat, IntValue: -2}, spell).Activate()
}

// A range of 0 is no range check at all, so a mod may not take the range there.
func TestRangeFlatModRefusesToEmptyTheRange(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("a mod taking the range to 0 did not panic")
		}
	}()

	spell := &Spell{MaxRange: 5}
	newTestMod(SpellModConfig{Kind: SpellMod_Range_Flat, FloatValue: -5}, spell).Activate()
}

func TestRangePctModRefusesToEmptyTheRange(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("a mod taking the range multiplier to 0 did not panic")
		}
	}()

	spell := &Spell{MaxRange: 30, RangeMultiplier: 1}
	newTestMod(SpellModConfig{Kind: SpellMod_Range_Pct, FloatValue: -1}, spell).Activate()
}

// Every rank of a family points at one aura and a class mask names every rank, so the value lands on
// that aura once however many ranks the mod reaches.
func TestAuraMutatingModsReachAnAuraOnce(t *testing.T) {
	buff := &Aura{Label: "Shared Buff", Duration: time.Second * 10, MaxStacks: 2}
	ranks := []*Spell{{RelatedSelfBuff: buff}, {RelatedSelfBuff: buff}, {RelatedSelfBuff: buff}}

	duration := newTestMod(SpellModConfig{Kind: SpellMod_Duration_Flat, TimeValue: time.Second * 6}, ranks...)
	stacks := newTestMod(SpellModConfig{Kind: SpellMod_BuffMaxStacks_Flat, IntValue: 2}, ranks...)

	duration.Activate()
	stacks.Activate()
	if got := buff.Duration; got != time.Second*16 {
		t.Errorf("the shared aura's duration with the mod active: %v, want 16s", got)
	}
	if got := buff.MaxStacks; got != 4 {
		t.Errorf("the shared aura's max stacks with the mod active: %d, want 4", got)
	}

	duration.Deactivate()
	stacks.Deactivate()
	if got := buff.Duration; got != time.Second*10 {
		t.Errorf("the shared aura's duration after removing the mod: %v, want 10s", got)
	}
	if got := buff.MaxStacks; got != 2 {
		t.Errorf("the shared aura's max stacks after removing the mod: %d, want 2", got)
	}

	duration.Activate()
	if got := buff.Duration; got != time.Second*16 {
		t.Errorf("the shared aura's duration on turning the mod on again: %v, want 16s", got)
	}
}

// Two mods of the same kind each reach the aura, since the dedupe is one mod's own bookkeeping.
func TestTwoAuraMutatingModsBothReachOneAura(t *testing.T) {
	buff := &Aura{Label: "Shared Buff", Duration: time.Second * 10}
	spell := &Spell{RelatedSelfBuff: buff}

	newTestMod(SpellModConfig{Kind: SpellMod_Duration_Flat, TimeValue: time.Second * 6}, spell).Activate()
	newTestMod(SpellModConfig{Kind: SpellMod_Duration_Flat, TimeValue: time.Second * 3}, spell).Activate()

	if got := buff.Duration; got != time.Second*19 {
		t.Errorf("the aura's duration under two mods: %v, want 19s", got)
	}
}

// Three ranks pointing at one self-buff, which the buff duration mod has to reach once.
func TestBuffDurationFlatModReachesASharedAuraOnce(t *testing.T) {
	buff := &Aura{Label: "Shared Buff", Duration: time.Second * 10}
	ranks := []*Spell{{RelatedSelfBuff: buff}, {RelatedSelfBuff: buff}, {RelatedSelfBuff: buff}}

	mod := newTestMod(SpellModConfig{Kind: SpellMod_BuffDuration_Flat, TimeValue: time.Second * 4}, ranks...)

	mod.Activate()
	if got := buff.Duration; got != time.Second*14 {
		t.Errorf("the shared buff's duration with the mod active: %v, want 14s", got)
	}

	mod.Deactivate()
	if got := buff.Duration; got != time.Second*10 {
		t.Errorf("the shared buff's duration after removing the mod: %v, want 10s", got)
	}
}

// The same for the debuff array every rank of a family hands out.
func TestDebuffDurationFlatModReachesASharedAuraOnce(t *testing.T) {
	debuff := &Aura{Label: "Shared Debuff", Duration: time.Second * 12}
	arrays := LabeledAuraArrays{"debuff": AuraArray{debuff, nil}}
	ranks := []*Spell{{RelatedAuraArrays: arrays}, {RelatedAuraArrays: arrays}, {RelatedAuraArrays: arrays}}

	mod := newTestMod(SpellModConfig{Kind: SpellMod_DebuffDuration_Flat, KeyValue: "debuff",
		TimeValue: time.Second * 3}, ranks...)

	mod.Activate()
	if got := debuff.Duration; got != time.Second*15 {
		t.Errorf("the shared debuff's duration with the mod active: %v, want 15s", got)
	}

	mod.Deactivate()
	if got := debuff.Duration; got != time.Second*12 {
		t.Errorf("the shared debuff's duration after removing the mod: %v, want 12s", got)
	}
}

// A spell registered after the mod was turned on has the mod applied to it, which is what
// OnSpellRegistered does on a unit's mod, and the aura it shares must not take the value again.
func TestAuraMutatingModReachesALateSpellOnce(t *testing.T) {
	buff := &Aura{Label: "Shared Buff", Duration: time.Second * 10}
	first := &Spell{RelatedSelfBuff: buff}

	mod := newTestMod(SpellModConfig{Kind: SpellMod_Duration_Flat, TimeValue: time.Second * 6}, first)
	mod.Activate()

	late := &Spell{RelatedSelfBuff: buff}
	mod.AffectedSpells = append(mod.AffectedSpells, late)
	mod.Apply(mod, late)

	if got := buff.Duration; got != time.Second*16 {
		t.Errorf("the shared aura's duration after a late rank: %v, want 16s", got)
	}

	mod.Deactivate()
	if got := buff.Duration; got != time.Second*10 {
		t.Errorf("the shared aura's duration after removing the mod: %v, want 10s", got)
	}
}

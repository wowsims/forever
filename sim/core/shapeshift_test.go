package core

import (
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
)

var (
	wrathRequirement    = CastRequirement{Forms: 0x40000000, ExcludedForms: 0x2, CasterForm: true, NotShapeshifted: true}
	regrowthRequirement = CastRequirement{Forms: 0x2, ExcludedForms: 0x40000000, CasterForm: true, NotShapeshifted: true}
	chargeRequirement   = CastRequirement{Forms: 0x10000}
)

func TestShapeshiftFormRule(t *testing.T) {

	tests := []struct {
		name    string
		req     CastRequirement
		allowed []dbcenums.ShapeshiftForm
		refused []dbcenums.ShapeshiftForm
	}{
		{
			name:    "Wrath",
			req:     wrathRequirement,
			allowed: []dbcenums.ShapeshiftForm{0, dbcenums.FORM_MOONKIN_FORM},
			refused: []dbcenums.ShapeshiftForm{dbcenums.FORM_CAT_FORM, dbcenums.FORM_DIRE_BEAR_FORM, dbcenums.FORM_TREE_FORM},
		},
		{
			name:    "Regrowth",
			req:     regrowthRequirement,
			allowed: []dbcenums.ShapeshiftForm{0, dbcenums.FORM_TREE_FORM},
			refused: []dbcenums.ShapeshiftForm{dbcenums.FORM_MOONKIN_FORM, dbcenums.FORM_CAT_FORM},
		},
		{
			name:    "Charge",
			req:     chargeRequirement,
			allowed: []dbcenums.ShapeshiftForm{dbcenums.FORM_BATTLE_STANCE},
			refused: []dbcenums.ShapeshiftForm{dbcenums.FORM_DEFENSIVE_STANCE, dbcenums.FORM_BERSERKER_STANCE, 0},
		},
		{
			name: "Heroic Strike",
			allowed: []dbcenums.ShapeshiftForm{0, dbcenums.FORM_CAT_FORM, dbcenums.FORM_BEAR_FORM, dbcenums.FORM_DIRE_BEAR_FORM,
				dbcenums.FORM_TREE_FORM, dbcenums.FORM_BATTLE_STANCE, dbcenums.FORM_DEFENSIVE_STANCE, dbcenums.FORM_BERSERKER_STANCE,
				dbcenums.FORM_MOONKIN_FORM},
		},
	}

	for _, test := range tests {
		for _, form := range test.allowed {
			if !test.req.allowsForm(form) {
				t.Errorf("%s should be castable in form %d", test.name, form)
			}
		}
		for _, form := range test.refused {
			if test.req.allowsForm(form) {
				t.Errorf("%s should not be castable in form %d", test.name, form)
			}
		}
	}
}

func TestCastRequirementAutoUnshift(t *testing.T) {
	unit := &Unit{ShapeshiftForm: dbcenums.FORM_CAT_FORM}
	spell := &Spell{Unit: unit, CastRequirement: wrathRequirement}

	if reason, _ := spell.castRequirementFailure(nil, nil); reason != "wrong form" {
		t.Errorf("Wrath in Cat Form without AutoUnshift: %q, want \"wrong form\"", reason)
	}

	unit.AutoUnshift = func(*Simulation) {}
	if reason, unshift := spell.castRequirementFailure(nil, nil); reason != "" || !unshift {
		t.Errorf("Wrath in Cat Form with AutoUnshift: (%q, %v), want (\"\", true)", reason, unshift)
	}

	unit.ShapeshiftForm = dbcenums.FORM_TREE_FORM
	if reason, unshift := spell.castRequirementFailure(nil, nil); reason != "" || !unshift {
		t.Errorf("Wrath in Tree of Life with AutoUnshift: (%q, %v), want (\"\", true)", reason, unshift)
	}

	unit.ShapeshiftForm = dbcenums.FORM_MOONKIN_FORM
	spell.CastRequirement = regrowthRequirement
	if reason, unshift := spell.castRequirementFailure(nil, nil); reason != "" || !unshift {
		t.Errorf("Regrowth in Moonkin Form with AutoUnshift: (%q, %v), want (\"\", true)", reason, unshift)
	}

	unit.ShapeshiftForm = dbcenums.FORM_DEFENSIVE_STANCE
	spell.CastRequirement = chargeRequirement
	if reason, _ := spell.castRequirementFailure(nil, nil); reason != "wrong form" {
		t.Errorf("Charge in Defensive Stance: %q, want \"wrong form\"", reason)
	}
}

func TestCastRequirementCasterAura(t *testing.T) {
	unit := &Unit{auraTracker: newAuraTracker()}
	catForm := unit.RegisterAura(Aura{Label: "Cat Form", ActionID: ActionID{SpellID: 768}, Duration: NeverExpires})
	catFormSpell := &Spell{Unit: unit, CastRequirement: CastRequirement{CasterAura: 768}}
	catFormSpell.resolveCasterAuras()

	if reason, _ := catFormSpell.castRequirementFailure(nil, nil); reason != "missing caster aura" {
		t.Errorf("Cat Form spell without Cat Form: %q, want \"missing caster aura\"", reason)
	}

	catForm.active = true
	if reason, _ := catFormSpell.castRequirementFailure(nil, nil); reason != "" {
		t.Errorf("Cat Form spell in Cat Form: %q, want castable", reason)
	}

	notInCat := &Spell{Unit: unit, CastRequirement: CastRequirement{ExcludeCasterAura: 768}}
	notInCat.resolveCasterAuras()
	if reason, _ := notInCat.castRequirementFailure(nil, nil); reason != "excluded caster aura" {
		t.Errorf("a spell excluding Cat Form while in it: %q, want \"excluded caster aura\"", reason)
	}
}

// A hand the requirement names must hold a matching weapon, both where it names both; with no hand
// named any weapon slot will do. A shield is the off hand's, and nil equipment meets nothing.
func TestCastRequirementEquipment(t *testing.T) {
	twoHandedAxe := Item{WeaponType: proto.WeaponType_WeaponTypeAxe, HandType: proto.HandType_HandTypeTwoHand}
	oneHandedAxe := Item{WeaponType: proto.WeaponType_WeaponTypeAxe, HandType: proto.HandType_HandTypeOneHand}
	dagger := Item{WeaponType: proto.WeaponType_WeaponTypeDagger, HandType: proto.HandType_HandTypeOneHand}
	shield := Item{WeaponType: proto.WeaponType_WeaponTypeShield}
	bow := Item{RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeBow}

	spearingStrike := CastRequirement{Weapon: EquipRequirement{SubclassMask: 0x562, Hand: ProcMaskMeleeMH}}
	mutilate := CastRequirement{Weapon: EquipRequirement{SubclassMask: dbcenums.ITEM_SUBCLASS_BIT_WEAPON_DAGGERS, Hand: ProcMaskMeleeMH | ProcMaskMeleeOH}}
	aimedShot := CastRequirement{Weapon: EquipRequirement{SubclassMask: dbcenums.ITEM_SUBCLASS_MASK_WEAPON_SHOOTING}}
	shieldBash := CastRequirement{Shield: true}

	gear := func(mh, oh, ranged Item) *Equipment {
		var equipment Equipment
		equipment[proto.ItemSlot_ItemSlotMainHand] = mh
		equipment[proto.ItemSlot_ItemSlotOffHand] = oh
		equipment[proto.ItemSlot_ItemSlotRanged] = ranged
		return &equipment
	}

	for _, c := range []struct {
		name      string
		req       CastRequirement
		equipment *Equipment
		want      bool
	}{
		{"Spearing Strike with a two-handed axe", spearingStrike, gear(twoHandedAxe, Item{}, Item{}), true},
		{"Spearing Strike with a one-handed axe", spearingStrike, gear(oneHandedAxe, oneHandedAxe, Item{}), false},
		{"Spearing Strike unarmed", spearingStrike, gear(Item{}, Item{}, Item{}), false},
		{"Mutilate with two daggers", mutilate, gear(dagger, dagger, Item{}), true},
		{"Mutilate with a main-hand dagger alone", mutilate, gear(dagger, oneHandedAxe, Item{}), false},
		{"Aimed Shot with a bow", aimedShot, gear(oneHandedAxe, Item{}, bow), true},
		{"Aimed Shot with no ranged weapon", aimedShot, gear(oneHandedAxe, Item{}, Item{}), false},
		{"Shield Bash with a shield", shieldBash, gear(oneHandedAxe, shield, Item{}), true},
		{"Shield Bash with an off-hand axe", shieldBash, gear(oneHandedAxe, oneHandedAxe, Item{}), false},
		{"Shield Bash on a unit that is not a player", shieldBash, nil, false},
		{"no requirement on a unit that is not a player", CastRequirement{}, nil, true},
	} {
		spell := &Spell{Unit: &Unit{}, CastRequirement: c.req, equipment: c.equipment}
		reason, _ := spell.castRequirementFailure(nil, nil)
		if got := reason == ""; got != c.want {
			t.Errorf("%s: castable = %v (%q), want %v", c.name, got, reason, c.want)
		} else if !got && reason != "missing equipped item" {
			t.Errorf("%s: %q, want \"missing equipped item\"", c.name, reason)
		}
	}

	equipment := gear(twoHandedAxe, Item{}, Item{})
	spell := &Spell{Unit: &Unit{}, CastRequirement: spearingStrike, equipment: equipment}
	equipment[proto.ItemSlot_ItemSlotMainHand] = oneHandedAxe
	if reason, _ := spell.castRequirementFailure(nil, nil); reason != "missing equipped item" {
		t.Errorf("Spearing Strike after swapping to a one-handed axe: %q, want \"missing equipped item\"", reason)
	}
}

// A spell castable only out of combat is castable before the encounter starts, and not from its start.
func TestCastRequirementNotInCombat(t *testing.T) {
	sim := &Simulation{}
	charge := &Spell{Unit: &Unit{}, CastRequirement: CastRequirement{NotInCombat: true}}

	sim.CurrentTime = -time.Second
	if reason, _ := charge.castRequirementFailure(sim, nil); reason != "" {
		t.Errorf("Charge before the encounter: %q, want castable", reason)
	}
	sim.CurrentTime = 0
	if reason, _ := charge.castRequirementFailure(sim, nil); reason != "in combat" {
		t.Errorf("Charge at the encounter's start: %q, want \"in combat\"", reason)
	}
}

func TestCastRequirementCastPath(t *testing.T) {
	sim := SetupFakeRageSim()
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	target := sim.Encounter.ActiveTargetUnits[0]
	sim.CurrentTime = 0

	var logs []string
	sim.Log = func(message string, vals ...any) {
		logs = append(logs, message)
	}

	charge := fw.RegisterSpell(SpellConfig{
		ActionID:        ActionID{SpellID: 100},
		Cast:            CastConfig{DefaultCast: Cast{GCD: time.Second}},
		CastRequirement: chargeRequirement,
	})
	fw.ShapeshiftForm = dbcenums.FORM_DEFENSIVE_STANCE
	if charge.CanCast(sim, target) {
		t.Error("Charge should not be castable in Defensive Stance")
	}
	if charge.Cast(sim, target) {
		t.Error("Charge should not cast in Defensive Stance")
	}
	if len(logs) == 0 || !strings.Contains(logs[len(logs)-1], "wrong form") {
		t.Errorf("Charge's failure should log \"wrong form\", logs: %v", logs)
	}

	wrath := fw.RegisterSpell(SpellConfig{
		ActionID:        ActionID{SpellID: 5176},
		Cast:            CastConfig{DefaultCast: Cast{GCD: time.Second}},
		CastRequirement: wrathRequirement,
	})
	unshifts := 0
	fw.AutoUnshift = func(*Simulation) {
		unshifts++
		fw.ShapeshiftForm = 0
	}
	fw.ShapeshiftForm = dbcenums.FORM_CAT_FORM
	if !wrath.CanCast(sim, target) {
		t.Error("Wrath should be castable in Cat Form when the cast leaves it")
	}
	if unshifts != 0 {
		t.Errorf("CanCast left the form %d times, want 0", unshifts)
	}
	if !wrath.Cast(sim, target) {
		t.Error("Wrath should cast in Cat Form when the cast leaves it")
	}
	if unshifts != 1 || fw.ShapeshiftForm != 0 {
		t.Errorf("Wrath's cast left the form %d times, now in form %d; want once, no form", unshifts, fw.ShapeshiftForm)
	}
}

func TestMajorCooldownKeepsForm(t *testing.T) {
	sim := SetupFakeRageSim()
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	sim.CurrentTime = 0

	potion := fw.RegisterSpell(SpellConfig{
		ActionID:        ActionID{ItemID: 13442},
		CastRequirement: CastRequirement{NotShapeshifted: true},
	})
	fw.AutoUnshift = func(*Simulation) {
		fw.ShapeshiftForm = 0
	}
	mcd := &MajorCooldown{
		Spell:          potion,
		ShouldActivate: func(*Simulation, *Character) bool { return true },
	}

	fw.ShapeshiftForm = dbcenums.FORM_CAT_FORM
	if !potion.CanCast(sim, fw.CurrentTarget) {
		t.Fatal("the potion should be castable in Cat Form when the cast leaves it")
	}
	if mcd.shouldActivateHelper(sim, &fw.Character) {
		t.Error("a major cooldown should not activate when its cast leaves Cat Form")
	}
	mcd.timings = []time.Duration{0}
	if mcd.shouldActivateHelper(sim, &fw.Character) {
		t.Error("a timed major cooldown should not activate when its cast leaves Cat Form")
	}

	fw.ShapeshiftForm = 0
	if !mcd.shouldActivateHelper(sim, &fw.Character) {
		t.Error("a timed major cooldown should activate with no form")
	}
	mcd.timings = nil
	if !mcd.shouldActivateHelper(sim, &fw.Character) {
		t.Error("a major cooldown should activate with no form")
	}
}

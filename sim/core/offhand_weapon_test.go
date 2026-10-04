package core

import (
	"math"
	"testing"
)

// A modifier on the off-hand's hits alone reaches an off-hand hit only while an off-hand weapon is
// behind it. A modifier on both hands still reaches a shield's hit.
func TestOffHandModsNeedAnOffHandWeapon(t *testing.T) {
	offHandHit := func(dualWielding bool) *Spell {
		return &Spell{ProcMask: ProcMaskMeleeOHSpecial, Unit: &Unit{AutoAttacks: AutoAttacks{IsDualWielding: dualWielding}}}
	}

	tests := []struct {
		name  string
		mod   *SpellMod
		spell *Spell
		want  bool
	}{
		{"an off-hand mod on an off-hand weapon's hit", &SpellMod{ProcMask: ProcMaskMeleeOH}, offHandHit(true), true},
		{"an off-hand mod on a shield's hit", &SpellMod{ProcMask: ProcMaskMeleeOH}, offHandHit(false), false},
		{"an off-hand special mod on a shield's hit", &SpellMod{ProcMask: ProcMaskMeleeOHSpecial}, offHandHit(false), false},
		{"a melee mod on a shield's hit", &SpellMod{ProcMask: ProcMaskMelee}, offHandHit(false), true},
	}

	for _, test := range tests {
		if got := shouldApply(test.spell, test.mod); got != test.want {
			t.Errorf("%s: shouldApply = %v, want %v", test.name, got, test.want)
		}
	}
}

// A procs-per-minute rate on an off-hand hit is measured against the off-hand weapon, and against
// the main hand where there is none behind the hit.
func TestOffHandProcsPerMinuteFallBackToTheMainHand(t *testing.T) {
	const ppm = 6.0

	for _, row := range []struct {
		name      string
		offHand   float64
		wantSpeed float64
	}{
		{"an off-hand weapon", fakeOHSwingSpeed, fakeOHSwingSpeed},
		{"a shield", 0, fakeMHSwingSpeed},
	} {
		character := &Character{}
		character.AutoAttacks.AutoSwingMelee = true
		character.AutoAttacks.mh.Weapon = Weapon{SwingSpeed: fakeMHSwingSpeed}
		character.AutoAttacks.oh.Weapon = Weapon{SwingSpeed: row.offHand}

		dpm := character.newDynamicWeaponProcManager(ppm, 0, ProcMaskMelee)
		want := ppm * row.wantSpeed / 60
		if got, _ := dpm.chanceFor(ProcMaskMeleeOHSpecial, nil); math.Abs(got-want) > 1e-12 {
			t.Errorf("%s: an off-hand special rolls %v, want %v", row.name, got, want)
		}
		wantMH := ppm * fakeMHSwingSpeed / 60
		if got, _ := dpm.chanceFor(ProcMaskMeleeMHSpecial, nil); math.Abs(got-wantMH) > 1e-12 {
			t.Errorf("%s: a main-hand special rolls %v, want %v", row.name, got, wantMH)
		}
		if got := character.AutoAttacks.offHandProcSpeed(); got != row.wantSpeed {
			t.Errorf("%s: a one-off rate is measured against %v, want %v", row.name, got, row.wantSpeed)
		}
	}
}

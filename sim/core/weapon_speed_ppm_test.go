package core

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/stats"
)

// A rate on the weapon's speed follows Seal of the Crusader's faster, lighter swings and a weapon
// swap, and stays put under haste.
func TestWeaponSpeedPPMReadsTheWeaponNotHaste(t *testing.T) {
	const ppm = 5.0

	character := &Character{}
	character.PseudoStats = stats.NewPseudoStats()
	character.AutoAttacks.mh.Weapon = Weapon{SwingSpeed: 3.6}
	dpm := character.NewWeaponSpeedPPMManager(ppm, ProcMaskMelee)

	check := func(state string, want float64) {
		t.Helper()
		if got := dpm.Chance(ProcMaskMeleeMHAuto, nil); math.Abs(got-want) > 1e-12 {
			t.Errorf("%s: a swing rolls %v, want %v", state, got, want)
		}
	}

	check("a 3.6 speed weapon", ppm*3.6/60)

	character.MultiplyMeleeSpeed(nil, 1.3)
	check("under 30% haste", ppm*3.6/60)

	character.MultiplyWeaponSpeed(nil, 1.4)
	check("under Seal of the Crusader", ppm*3.6/60/1.4)

	character.AutoAttacks.mh.Weapon = Weapon{SwingSpeed: 1.6}
	check("after a swap to a 1.6 speed weapon", ppm*1.6/60/1.4)

	character.MultiplyWeaponSpeed(nil, 1/1.4)
	check("after the seal drops", ppm*1.6/60)
}

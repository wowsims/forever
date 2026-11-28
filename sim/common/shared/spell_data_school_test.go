package shared

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
)

// The generator emits a school by name and the game files state it as a bit; nothing else would notice
// if core's SpellSchool constants drifted off the game files' bits.
func TestSpellSchoolBitsMatchTheGameFiles(t *testing.T) {
	// SpellMisc.SchoolMask, as the game files number it.
	for _, c := range []struct {
		clientBit byte
		school    core.SpellSchool
		name      string
	}{
		{1, core.SpellSchoolPhysical, "Physical"},
		{2, core.SpellSchoolHoly, "Holy"},
		{4, core.SpellSchoolFire, "Fire"},
		{8, core.SpellSchoolNature, "Nature"},
		{16, core.SpellSchoolFrost, "Frost"},
		{32, core.SpellSchoolShadow, "Shadow"},
		{64, core.SpellSchoolArcane, "Arcane"},
	} {
		if byte(c.school) != c.clientBit {
			t.Errorf("%s is bit %d in the game files and %d in core", c.name, c.clientBit, byte(c.school))
		}
	}
}

package core_test

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func TestSpiritRegenIsTheSameForEveryClass(t *testing.T) {
	for _, tc := range []struct {
		class     proto.Class
		race      proto.Race
		spirit    float64
		perSecond float64
	}{
		{proto.Class_ClassPriest, proto.Race_RaceNightElf, 125, 6.25 + 125.0/8},
		{proto.Class_ClassMage, proto.Race_RaceGnome, 120, 6.25 + 120.0/8},
		{proto.Class_ClassPaladin, proto.Race_RaceDwarf, 75, 6.25 + 75.0/8},
		{proto.Class_ClassDruid, proto.Race_RaceNightElf, 110, 6.25 + 110.0/8},
	} {
		t.Run(tc.class.String(), func(t *testing.T) {
			character, _ := nakedSheet(t, tc.class, tc.race)
			if spirit := character.GetStat(stats.Spirit); spirit != tc.spirit {
				t.Fatalf("spirit %v, want %v", spirit, tc.spirit)
			}
			if got := character.SpiritManaRegenPerSecond(); math.Abs(got-tc.perSecond) > 1e-9 {
				t.Errorf("spirit regen %v per second, want %v", got, tc.perSecond)
			}
		})
	}
}

func TestSetSpiritManaRegenReplacesTheDefault(t *testing.T) {
	owner := core.NewCharacter(&core.Party{Raid: &core.Raid{}}, 0, nakedPlayer(proto.Class_ClassWarlock, proto.Race_RaceHuman))
	newPet := func(name string) *core.Pet {
		pet := core.NewPet(core.PetConfig{Name: name, Owner: &owner, BaseStats: stats.Stats{stats.Spirit: 100}})
		pet.EnableManaBar()
		return &pet
	}

	unchanged := newPet("Default Pet")
	if got := unchanged.SpiritManaRegenPerSecond(); math.Abs(got-(6.25+100.0/8)) > 1e-9 {
		t.Errorf("default pet spirit regen %v per second, want %v", got, 6.25+100.0/8)
	}

	imp := newPet("Imp")
	imp.SetSpiritManaRegen(0, 1.0/5)
	if got := imp.SpiritManaRegenPerSecond(); math.Abs(got-100.0/5) > 1e-9 {
		t.Errorf("spirit regen %v per second, want %v", got, 100.0/5)
	}
	if got := unchanged.SpiritManaRegenPerSecond(); math.Abs(got-(6.25+100.0/8)) > 1e-9 {
		t.Errorf("setting one pet's spirit regen changed another's to %v per second", got)
	}
}

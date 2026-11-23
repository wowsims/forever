package shared

import (
	"fmt"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// A test trinket that summons as Barov Peasant Caller's 18307 does: three servants 11637 for 20 s.
const servantTrinket int32 = 993101

var servant = GuardianConfig{
	Summon: spelldata.MustFind(18307),
	Health: 0.5,
	Armor:  0.25,
	Swing:  GuardianSwing{Share: 0.5, Speed: 2.0, Variance: 0.1},
}

func init() {
	RegisterItemGuardian(servantTrinket, "Test Servant", servant)
}

// One item whose summon brings three builds three guardians, numbered, at the shares of the creature numbers.
func TestRegisterItemGuardianBuildsOneGuardianPerCreatureTheSummonBrings(t *testing.T) {
	_, wearer := newConsumableSim(t, rageUser(), &proto.ConsumesSpec{}, servantTrinket)

	guardians := Guardians(&wearer.Character, servant.Summon)
	if want := int(servant.Summon.SummonEffect().BasePoints); want < 2 || len(guardians) != want {
		t.Fatalf("the wearer owns %d guardians for 18307, want the %d its base points bring, more than one", len(guardians), want)
	}
	for i, guardian := range guardians {
		if want := fmt.Sprintf("Test Servant %d", i+1); guardian.Name != want {
			t.Errorf("guardian %d is named %q, want %q", i, guardian.Name, want)
		}
		if !guardian.IsGuardian() {
			t.Errorf("%s is not a guardian", guardian.Name)
		}
		base := guardian.GetBaseStats()
		if base[stats.Health] != 0.5*core.CreatureHealthMaxLevel || base[stats.Armor] != 0.25*core.CreatureArmorMaxLevel {
			t.Errorf("%s has %.1f health and %.1f armor, want half the creature health and a quarter of its armor", guardian.Name,
				base[stats.Health], base[stats.Armor])
		}
	}
}

// A summon puts up every creature its row brings on the target, each swinging at it, until the row's duration runs out.
func TestSummonGuardiansPutsUpTheRowsCountForTheRowsDuration(t *testing.T) {
	sim, wearer := newConsumableSim(t, rageUser(), &proto.ConsumesSpec{}, servantTrinket)
	wearer.AutoAttacks.CancelAutoSwing(sim)
	guardians := Guardians(&wearer.Character, servant.Summon)
	target := wearer.CurrentTarget

	start := sim.CurrentTime
	SummonGuardians(sim, guardians, target)
	duration := servant.Summon.Duration()
	stepPast(t, sim, start+duration-time.Millisecond)
	for _, guardian := range guardians {
		if !guardian.IsEnabled() || guardian.CurrentTarget != target {
			t.Fatalf("%s is not up on the target 1 ms before the row's %v run out", guardian.Name, duration)
		}
		// 20 s of a 2.0 s swing, the first as it arrives.
		if swings := guardian.AutoAttacks.MHAuto().SpellMetrics[target.UnitIndex].Casts; swings < 9 || swings > 11 {
			t.Errorf("%s swung %d times in %v, want one every 2.0 s", guardian.Name, swings, duration)
		}
	}
	stepPast(t, sim, start+duration)
	for _, guardian := range guardians {
		if guardian.IsEnabled() {
			t.Errorf("%s is still up as the row's %v run out", guardian.Name, duration)
		}
	}
}

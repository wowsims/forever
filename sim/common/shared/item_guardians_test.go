package shared

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Barov Peasant Caller's 18307 summons servants 11637. The tests pin it at three servants 11637 for 20 s, so a game
// build that changes the row leaves them as they are.
const (
	servantSummon   int32 = 18307
	servantCreature int32 = 11637
)

var servant = GuardianCreature{
	Name:   "Test Servant",
	Health: 0.5,
	Armor:  0.25,
	Swing:  GuardianSwing{Share: 0.5, Speed: 2.0, Variance: 0.1},
}

// A caster wearing a test trinket whose use is the pinned servant summon, on a 5 min cooldown, registered through
// NewGuardianOnUse with the given creatures.
func newServantSim(t *testing.T, itemID int32, creatures map[int32]GuardianCreature) (*core.Simulation, *testCaster) {
	t.Helper()
	editRow(t, servantSummon, func(row *spelldata.Spell) {
		row.Effects[0].Misc, row.Effects[0].BasePoints, row.DurationMs = servantCreature, 3, 20000
	})
	return newOnUseSim(t, func(id int32) { NewGuardianOnUse(id, servantSummon, creatures) },
		map[int32]*proto.ItemEffect{itemID: testOnUse(servantSummon, 300000, 0, 0)})
}

// A use item builds one guardian for each creature its summon brings, numbered, at the stated shares of the creature
// numbers.
func TestNewGuardianOnUseBuildsOneGuardianPerCreatureTheSummonBrings(t *testing.T) {
	_, wearer := newServantSim(t, 993101, map[int32]GuardianCreature{servantCreature: servant})

	guardians := Guardians(&wearer.Character, spelldata.MustFind(servantSummon))
	if len(guardians) != 3 {
		t.Fatalf("the wearer owns %d guardians, want the 3 its summon brings", len(guardians))
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

// The use puts every guardian up on the wearer's target, each swinging at it, until the summon's duration runs out,
// and then rests the item's cooldown.
func TestNewGuardianOnUsePutsTheGuardiansUpOnTheTargetForTheSummonsDuration(t *testing.T) {
	sim, wearer := newServantSim(t, 993102, map[int32]GuardianCreature{servantCreature: servant})
	wearer.AutoAttacks.CancelAutoSwing(sim)
	guardians := Guardians(&wearer.Character, spelldata.MustFind(servantSummon))
	target := wearer.CurrentTarget
	use := onUseSpell(t, wearer, 993102)

	start := sim.CurrentTime
	if !use.Cast(sim, target) {
		t.Fatal("the use did not cast")
	}
	if got := use.CD.TimeToReady(sim); got != 5*time.Minute {
		t.Errorf("the use rests %v, want the item's 5 min", got)
	}
	stepPast(t, sim, start+20*time.Second-time.Millisecond)
	for _, guardian := range guardians {
		if !guardian.IsEnabled() || guardian.CurrentTarget != target {
			t.Fatalf("%s is not up on the target 1 ms before the summon's 20 s run out", guardian.Name)
		}
		// 20 s of a 2.0 s swing, the first as it arrives.
		if swings := guardian.AutoAttacks.MHAuto().SpellMetrics[target.UnitIndex].Casts; swings < 9 || swings > 11 {
			t.Errorf("%s swung %d times in 20 s, want one every 2.0 s", guardian.Name, swings)
		}
	}
	stepPast(t, sim, start+20*time.Second)
	for _, guardian := range guardians {
		if guardian.IsEnabled() {
			t.Errorf("%s is still up as the summon's 20 s run out", guardian.Name)
		}
	}
}

// A use item whose creature has no measurement registers neither the use nor a guardian, and keeps only its stats.
func TestNewGuardianOnUseRegistersNothingForAnUnmeasuredCreature(t *testing.T) {
	_, wearer := newServantSim(t, 993103, map[int32]GuardianCreature{})

	if spell := wearer.GetSpell(core.ActionID{ItemID: 993103}); spell != nil {
		t.Error("an unmeasured creature's item registered its use")
	}
	if guardians := Guardians(&wearer.Character, spelldata.MustFind(servantSummon)); len(guardians) != 0 {
		t.Errorf("an unmeasured creature's item built %d guardians, want none", len(guardians))
	}
}

// A guardian whose summon puts it beside its owner comes up as far from the target as the owner stands and walks
// in to melee range at its run speed, swinging only once it is there.
func TestNewGuardianOnUsePutsTheGuardiansUpBesideTheOwnerAndWalksThemIn(t *testing.T) {
	sim, wearer := newServantSim(t, 993104, map[int32]GuardianCreature{servantCreature: servant})
	wearer.AutoAttacks.CancelAutoSwing(sim)
	wearer.DistanceFromTarget = 30
	guardian := Guardians(&wearer.Character, spelldata.MustFind(servantSummon))[0]
	target := wearer.CurrentTarget

	start := sim.CurrentTime
	onUseSpell(t, wearer, 993104).Cast(sim, target)
	if guardian.DistanceFromTarget != 30 {
		t.Fatalf("%s comes up %.1f yd from the target, want the owner's 30", guardian.Name, guardian.DistanceFromTarget)
	}
	arrival := start + core.DurationFromSeconds((30-core.MaxMeleeRange)/guardian.GetMovementSpeed())
	stepPast(t, sim, arrival-100*time.Millisecond)
	if swings := guardian.AutoAttacks.MHAuto().SpellMetrics[target.UnitIndex].Casts; swings != 0 {
		t.Errorf("%s swung %d times before it reached melee range", guardian.Name, swings)
	}
	stepPast(t, sim, arrival+time.Second)
	if guardian.DistanceFromTarget > core.MaxMeleeRange {
		t.Errorf("%s stands %.1f yd from the target a second after it should have arrived", guardian.Name, guardian.DistanceFromTarget)
	}
	if swings := guardian.AutoAttacks.MHAuto().SpellMetrics[target.UnitIndex].Casts; swings == 0 {
		t.Errorf("%s has not swung a second after it reached melee range", guardian.Name)
	}
}

// A guardian whose summon puts it at the target comes up in melee range however far its owner stands.
func TestNewGuardianOnUsePutsAGuardianSummonedAtTheTargetInMeleeRange(t *testing.T) {
	editRow(t, servantSummon, func(row *spelldata.Spell) {
		row.Effects[0].Target[0] = dbcenums.TARGET_DEST_TARGET_RANDOM
	})
	sim, wearer := newServantSim(t, 993105, map[int32]GuardianCreature{servantCreature: servant})
	wearer.AutoAttacks.CancelAutoSwing(sim)
	wearer.DistanceFromTarget = 30
	guardian := Guardians(&wearer.Character, spelldata.MustFind(servantSummon))[0]

	onUseSpell(t, wearer, 993105).Cast(sim, wearer.CurrentTarget)
	if guardian.DistanceFromTarget > core.MaxMeleeRange {
		t.Errorf("%s comes up %.1f yd from the target, want melee range", guardian.Name, guardian.DistanceFromTarget)
	}
}

// A creature with a level of its own takes the creature numbers for that level, whatever its owner's.
func TestGuardianCreatureWithALevelTakesThatLevelsCreatureNumbers(t *testing.T) {
	leveled := servant
	leveled.Level = 57
	_, wearer := newServantSim(t, 993106, map[int32]GuardianCreature{servantCreature: leveled})
	guardian := Guardians(&wearer.Character, spelldata.MustFind(servantSummon))[0]
	level57 := core.CreatureStatsByLevel[57]

	base := guardian.GetBaseStats()
	if base[stats.Health] != 0.5*level57.Health || base[stats.Armor] != 0.25*level57.Armor {
		t.Errorf("%s has %.1f health and %.1f armor, want half the level 57 creature health and a quarter of its armor",
			guardian.Name, base[stats.Health], base[stats.Armor])
	}
	if want := 0.5 * level57.AutoAttackDps * 2.0 * 1.1; math.Abs(guardian.AutoAttacks.MH().BaseDamageMax-want) > 1e-9 {
		t.Errorf("%s swings up to %.2f, want half of 2 s of the level 57 creature dps plus 10%%, %.2f",
			guardian.Name, guardian.AutoAttacks.MH().BaseDamageMax, want)
	}
}

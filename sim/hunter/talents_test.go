package hunter

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// A reset sim of one hunter with the talents and, unless PetNone, a pet of that family.
func newTalentSim(talents string, pet proto.HunterOptions_PetType) (*core.Simulation, *Hunter) {
	sim, _, _ := newPetSimWithTalents(talents, pet, &proto.Debuffs{})
	return sim, sim.Raid.Parties[0].Players[0].(*Hunter)
}

func assertSpeed(t *testing.T, what string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("%s moves at %.4f yd/s, want %.4f", what, got, want)
	}
}

// Aspect of the Cheetah's 30% with Pathfinding's 6% on top, and Dazed's 70% slow on what is left.
func TestAspectOfTheCheetahSpeedWithPathfindingAndDazed(t *testing.T) {
	for _, tc := range []struct {
		talents string
		speed   float64
	}{
		{"", 0.30},
		{"00002", 0.36},
	} {
		sim, hunter := newTalentSim(tc.talents, proto.HunterOptions_PetNone)
		base := hunter.GetMovementSpeed()

		hunter.GetSpell(core.ActionID{SpellID: aspectOfTheCheetahRank.ID}).SkipCastAndApplyEffects(sim, hunter.CurrentTarget)
		assertSpeed(t, "a hunter in Cheetah with talents "+tc.talents, hunter.GetMovementSpeed(), base*(1+tc.speed))

		dazed := hunter.GetAura(dazedRank.Name)
		dazed.Activate(sim)
		assertSpeed(t, "a dazed hunter in Cheetah", hunter.GetMovementSpeed(), base*(1+tc.speed)*0.3)
		dazed.Deactivate(sim)

		hunter.GetSpell(core.ActionID{SpellID: aspectOfTheHawkRank.ID}).SkipCastAndApplyEffects(sim, hunter.CurrentTarget)
		if hunter.GetAura(aspectOfTheCheetahRank.Name).IsActive() {
			t.Error("Aspect of the Hawk left Aspect of the Cheetah up, want one aspect at a time")
		}
		assertSpeed(t, "a hunter back in Hawk", hunter.GetMovementSpeed(), base)
	}
}

// Survival Tactics' 10% hit on the two damage traps, on the effect spells that roll the hit.
func TestSurvivalTacticsRaisesTrapHit(t *testing.T) {
	for _, tc := range []struct {
		talents string
		hit     float64
	}{
		{"", 0},
		{"--0000000002", 10},
	} {
		_, hunter := newTalentSim(tc.talents, proto.HunterOptions_PetNone)
		for _, rank := range []int32{immolationTrapEffectRank.ID, explosiveTrapEffectRank.ID} {
			spell := hunter.GetSpell(core.ActionID{SpellID: rank})
			if spell == nil {
				t.Fatalf("the hunter has no trap effect %d", rank)
			}
			if spell.BonusHitPercent != tc.hit {
				t.Errorf("trap effect %d with talents %q has %v%% bonus hit, want %v%%", rank, tc.talents, spell.BonusHitPercent, tc.hit)
			}
		}
	}
}

// Bestial Swiftness' 30% on the pet, which Dash's larger speed replaces while it is up.
func TestBestialSwiftnessSpeedsThePetUnderDash(t *testing.T) {
	sim, hunter := newTalentSim("000000100", proto.HunterOptions_Wolf)
	pet := hunter.Pet
	assertSpeed(t, "a Bestial Swiftness pet", pet.GetMovementSpeed(), 8*1.3)

	dash := pet.closers[0]
	if dash.ActionID.SpellID != spellData.DashTriggered.Highest().ID {
		t.Fatalf("the wolf's closer is %v, want Dash", dash.ActionID)
	}
	dash.SkipCastAndApplyEffects(sim, pet.CurrentTarget)
	assertSpeed(t, "a dashing Bestial Swiftness pet", pet.GetMovementSpeed(), 8*(1+spellData.DashTriggered.Highest().EffectN(1).Percent()))
}

// Intimidation's 100% crit stays on the pet until an attack of its spends the charge.
func TestIntimidationCritsThePetsNextAttack(t *testing.T) {
	sim, hunter := newTalentSim("0000000000001", proto.HunterOptions_Wolf)
	pet := hunter.Pet
	crit := pet.GetStats()[stats.PhysicalCritPercent]

	hunter.GetSpell(core.ActionID{SpellID: intimidationRank.ID}).SkipCastAndApplyEffects(sim, hunter.CurrentTarget)
	aura := pet.GetAura(intimidationRank.Name)
	if got := pet.GetStats()[stats.PhysicalCritPercent]; math.Abs(got-crit-100) > 1e-9 {
		t.Errorf("the intimidating pet has %v%% crit, want %v%%", got, crit+100)
	}

	for range 1000 {
		if !aura.IsActive() {
			break
		}
		sim.Step()
	}
	if aura.IsActive() || sim.CurrentTime >= intimidationRank.Duration() {
		t.Fatalf("Intimidation was still up or ran out at %v, want its charge spent by the pet's first attack", sim.CurrentTime)
	}
	if got := pet.GetStats()[stats.PhysicalCritPercent]; math.Abs(got-crit) > 1e-9 {
		t.Errorf("the pet has %v%% crit after Intimidation, want %v%%", got, crit)
	}
}

// Rapid Killing comes with a killing blow or the death of an enemy with the hunter's Serpent Sting on
// it, once per death, and not with an unstung death the raid gets no kill for.
func TestRapidKillingComesWithAKillingBlowOrAStungDeath(t *testing.T) {
	sim, hunter := newTalentSim("-0000002", proto.HunterOptions_PetNone)
	target := sim.Encounter.AllTargets[0]
	buff := hunter.GetAuraByID(core.ActionID{SpellID: rapidKillingBuff.ID})
	sting := hunter.GetSpell(core.ActionID{SpellID: serpentStingRank.ID})
	grants := 0
	sim.Log = func(message string, vals ...any) {
		line := fmt.Sprintf(message, vals...)
		if strings.Contains(line, buff.ActionID.String()) && (strings.Contains(line, "Aura gained") || strings.Contains(line, "Aura refreshed")) {
			grants++
		}
	}
	stepPastBatch := func() {
		at := sim.CurrentTime + core.SpellBatchWindow
		for sim.CurrentTime <= at {
			sim.AddPendingAction(&core.PendingAction{NextActionAt: at + 1, OnAction: func(*core.Simulation) {}})
			sim.Step()
		}
	}

	for _, c := range []struct {
		name      string
		stung     bool
		killShare float64
		grants    int
	}{
		{"an unstung death with no kill credit", false, 0, 0},
		{"an unstung killing blow", false, 1, 1},
		{"a stung death with no kill credit", true, 0, 1},
		{"a stung killing blow", true, 1, 1},
	} {
		buff.Deactivate(sim)
		target.Enable(sim)
		if c.stung {
			sting.Dot(&target.Unit).Apply(sim)
		}
		grants = 0

		target.Die(sim, c.killShare)
		stepPastBatch()
		if grants != c.grants || buff.IsActive() != (c.grants > 0) {
			t.Errorf("%s granted Rapid Killing %d times and left it up %v, want %d", c.name, grants, buff.IsActive(), c.grants)
		}
		if buff.IsActive() && buff.GetStacks() != buff.MaxStacks {
			t.Errorf("%s left Rapid Killing at %d charges, want %d", c.name, buff.GetStacks(), buff.MaxStacks)
		}
	}
}

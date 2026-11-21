package core

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

const planBossSwingSpeed = 2 * time.Second

// A request whose boss swings at the fake Protection Warrior, its tank.
func planTankRequest(plan *proto.EncounterPlan) *proto.RaidSimRequest {
	request := planTestRequest(1, plan)
	request.Raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	boss := request.Encounter.Targets[0]
	boss.SwingSpeed = planBossSwingSpeed.Seconds()
	boss.MinBaseDamage = 1000
	return request
}

type planHitTaken struct {
	at       time.Duration
	actionID ActionID
	damage   float64
	outcome  HitOutcome
}

func recordPlanHitsTaken(unit *Unit, hits *[]planHitTaken) {
	MakePermanent(unit.RegisterAura(Aura{
		Label: "Hits Taken",
		OnSpellHitTaken: func(_ *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			*hits = append(*hits, planHitTaken{at: sim.CurrentTime, actionID: spell.ActionID, damage: result.Damage, outcome: result.Outcome})
		},
	}))
}

// A sim at the pull that records every hit the player takes; register runs as setupPlanSim's does.
func setupPlanHitsTaken(request *proto.RaidSimRequest, register func(char *Character)) (*Simulation, *[]planHitTaken) {
	hits := &[]planHitTaken{}
	sim, _ := setupPlanSim(request, func(char *Character) {
		recordPlanHitsTaken(&char.Unit, hits)
		if register != nil {
			register(char)
		}
	})
	return sim, hits
}

// Every hit the player takes over one iteration.
func runPlanHitsTaken(t *testing.T, request *proto.RaidSimRequest) (*Simulation, []planHitTaken) {
	t.Helper()
	sim, hits := setupPlanHitsTaken(request, nil)
	runLifecycleUntil(sim, sim.Duration)
	return sim, *hits
}

func bossSwingsBetween(hits []planHitTaken, from, to time.Duration) int {
	count := 0
	for _, hit := range hits {
		if isBossMeleeSwing(hit.actionID) && hit.at > from && hit.at < to {
			count++
		}
	}
	return count
}

func planHitsOf(hits []planHitTaken, matches func(ActionID) bool) []planHitTaken {
	var matched []planHitTaken
	for _, hit := range hits {
		if matches(hit.actionID) {
			matched = append(matched, hit)
		}
	}
	return matched
}

func isBossMeleeSwing(id ActionID) bool {
	return id.OtherID == proto.OtherAction_OtherActionAttack && id.Tag == int32(proto.ItemSlot_ItemSlotMainHand)
}

func planWindow(start, end float64) (*proto.PlanTime, *proto.PlanTime) {
	return &proto.PlanTime{At: start}, &proto.PlanTime{At: end}
}

func TestPlanSpellHitLandsAtEachOccurrenceAfterItsHardcast(t *testing.T) {
	const minDamage, maxDamage = 1000.0, 2000.0
	const castTime = 2 * time.Second
	spellHit := func(start, end float64, repeat *proto.PlanRepeat) *proto.PlanBossAbility {
		startAt, endAt := planWindow(start, end)
		return &proto.PlanBossAbility{
			Kind:      proto.PlanBossAbility_SpellHit,
			Start:     startAt,
			End:       endAt,
			Repeat:    repeat,
			School:    proto.SpellSchool_SpellSchoolFire,
			MinDamage: minDamage,
			MaxDamage: maxDamage,
			CastTime:  castTime.Seconds(),
		}
	}
	plan := &proto.EncounterPlan{BossAbilities: []*proto.PlanBossAbility{
		spellHit(0.1, 0.35, &proto.PlanRepeat{Interval: 0.1, Count: 2}),
		spellHit(0.55, 1, &proto.PlanRepeat{Interval: 0.1, Count: 1}),
	}}
	_, hits := runPlanHitsTaken(t, planTestRequest(1, plan))

	for index, want := range [][]time.Duration{
		{planAt(0.1) + castTime, planAt(0.2) + castTime, planAt(0.3) + castTime},
		{planAt(0.55) + castTime, planAt(0.65) + castTime},
	} {
		tag := planBossAbilityTag + int32(index)
		landed := planHitsOf(hits, func(id ActionID) bool { return id.SpellID == planSpellHitSpellID && id.Tag == tag })
		if len(landed) != len(want) {
			t.Fatalf("spell hit %d: expected %d hits on the player, got %d: %v", index+1, len(want), len(landed), landed)
		}
		for i, hit := range landed {
			if hit.at != want[i] {
				t.Errorf("spell hit %d, hit %d: expected at %s, got %s", index+1, i+1, want[i], hit.at)
			}
			if hit.damage < minDamage || hit.damage > maxDamage {
				t.Errorf("spell hit %d, hit %d: damage %.1f outside [%.0f, %.0f]", index+1, i+1, hit.damage, minDamage, maxDamage)
			}
		}
	}
}

func TestPlanBossAbilityOfADisabledTargetDoesNothing(t *testing.T) {
	startAt, endAt := planWindow(0.1, 0.35)
	plan := &proto.EncounterPlan{
		Downtimes: []*proto.PlanDowntime{{TargetIndex: 0, Start: &proto.PlanTime{At: 0.15}, End: &proto.PlanTime{At: 0.25}}},
		BossAbilities: []*proto.PlanBossAbility{{
			Kind:      proto.PlanBossAbility_SpellHit,
			Start:     startAt,
			End:       endAt,
			Repeat:    &proto.PlanRepeat{Interval: 0.1, Count: 2},
			MinDamage: 100,
			MaxDamage: 100,
		}},
	}
	_, hits := runPlanHitsTaken(t, planTestRequest(1, plan))

	landed := planHitsOf(hits, func(id ActionID) bool { return id.SpellID == planSpellHitSpellID })
	want := []time.Duration{planAt(0.1), planAt(0.3)}
	if len(landed) != len(want) {
		t.Fatalf("expected the hits at the occurrences outside the downtime, %v; got %v", want, landed)
	}
	for i, hit := range landed {
		if hit.at != want[i] {
			t.Errorf("hit %d: expected at %s, got %s", i+1, want[i], hit.at)
		}
	}
}

// A downtime or a death that takes the boss away mid-hardcast stops the cast, so the spell hit does not land
// once the boss is back; a later occurrence still does.
func TestPlanSpellHitStopsWhenItsTargetGoesAway(t *testing.T) {
	const castTime = 4 * time.Second
	plan := func(downtimes []*proto.PlanDowntime) *proto.EncounterPlan {
		startAt, endAt := planWindow(0.1, 0.2)
		return &proto.EncounterPlan{
			Downtimes: downtimes,
			BossAbilities: []*proto.PlanBossAbility{{
				Kind:      proto.PlanBossAbility_SpellHit,
				Start:     startAt,
				End:       endAt,
				Repeat:    &proto.PlanRepeat{Interval: 0.1, Count: 1},
				MinDamage: 100,
				MaxDamage: 100,
				CastTime:  castTime.Seconds(),
			}},
		}
	}
	spellHits := func(hits []planHitTaken) []planHitTaken {
		return planHitsOf(hits, func(id ActionID) bool { return id.SpellID == planSpellHitSpellID })
	}

	t.Run("downtime", func(t *testing.T) {
		downtime := []*proto.PlanDowntime{{TargetIndex: 0, Start: &proto.PlanTime{At: 0.11}, End: &proto.PlanTime{At: 0.12}}}
		_, hits := runPlanHitsTaken(t, planTestRequest(1, plan(downtime)))
		landed := spellHits(hits)
		if len(landed) != 1 || landed[0].at != planAt(0.2)+castTime {
			t.Errorf("expected only the hit of the occurrence after the downtime, at %s; got %v", planAt(0.2)+castTime, landed)
		}
	})

	t.Run("death", func(t *testing.T) {
		sim, hits := setupPlanHitsTaken(planTestRequest(1, plan(nil)), nil)
		boss := sim.Encounter.AllTargets[0]
		runLifecycleUntil(sim, planAt(0.11))
		if boss.Hardcast.Expires <= sim.CurrentTime {
			t.Fatal("test setup: expected the boss mid-hardcast")
		}
		boss.Die(sim, 0)
		if boss.Hardcast.Expires > sim.CurrentTime {
			t.Errorf("the death left the hardcast running until %s", boss.Hardcast.Expires)
		}
		runLifecycleUntil(sim, sim.Duration)
		if landed := spellHits(*hits); len(landed) != 0 {
			t.Errorf("expected no hit from the dead boss, got %v", landed)
		}
	})
}

func TestPlanStrikeReplacesOneSwingPerOccurrence(t *testing.T) {
	startAt, endAt := planWindow(0.1, 0.35)
	plan := &proto.EncounterPlan{BossAbilities: []*proto.PlanBossAbility{{
		Kind:          proto.PlanBossAbility_Strike,
		Start:         startAt,
		End:           endAt,
		Repeat:        &proto.PlanRepeat{Interval: 0.1, Count: 2},
		WeaponPercent: 150,
	}}}
	_, hits := runPlanHitsTaken(t, planTankRequest(plan))

	isStrike := func(id ActionID) bool { return id.SpellID == planStrikeSpellID }
	attacks := planHitsOf(hits, func(id ActionID) bool { return isBossMeleeSwing(id) || isStrike(id) })
	if len(planHitsOf(attacks, isStrike)) != 3 {
		t.Fatalf("expected 3 strikes, got %d", len(planHitsOf(attacks, isStrike)))
	}
	for i := 1; i < len(attacks); i++ {
		if gap := attacks[i].at - attacks[i-1].at; gap != planBossSwingSpeed {
			t.Fatalf("attack %d comes %s after the one before, expected one swing of %s", i+1, gap, planBossSwingSpeed)
		}
	}
	for _, occurrence := range []float64{0.1, 0.2, 0.3} {
		next := 0
		for next < len(attacks) && attacks[next].at < planAt(occurrence) {
			next++
		}
		if next+1 >= len(attacks) {
			t.Fatalf("no two attacks after the occurrence at %s", planAt(occurrence))
		}
		if !isStrike(attacks[next].actionID) {
			t.Errorf("the first swing after the occurrence at %s is %s, expected the strike", planAt(occurrence), attacks[next].actionID)
		}
		if isStrike(attacks[next+1].actionID) {
			t.Errorf("the second swing after the occurrence at %s is a strike too", planAt(occurrence))
		}
	}
}

func TestPlanThrashSwingsExtraOnlyInsideItsWindow(t *testing.T) {
	startAt, endAt := planWindow(0.2, 0.4)
	plan := &proto.EncounterPlan{BossAbilities: []*proto.PlanBossAbility{{
		Kind:         proto.PlanBossAbility_Thrash,
		Start:        startAt,
		End:          endAt,
		ProcChance:   100,
		ExtraAttacks: 1,
	}}}
	_, hits := runPlanHitsTaken(t, planTankRequest(plan))

	extra := planHitsOf(hits, func(id ActionID) bool {
		return id.OtherID == proto.OtherAction_OtherActionAttack && id.Tag == planBossAbilityTag
	})
	if len(extra) == 0 {
		t.Fatal("expected extra attacks inside the window, got none")
	}
	for _, hit := range extra {
		if hit.at < planAt(0.2) || hit.at >= planAt(0.4) {
			t.Errorf("extra attack at %s, outside the window [%s, %s)", hit.at, planAt(0.2), planAt(0.4))
		}
	}
}

func planBossWindow(kind proto.PlanBossAbility_Kind, start, end float64) *proto.PlanBossAbility {
	startAt, endAt := planWindow(start, end)
	return &proto.PlanBossAbility{Kind: kind, Start: startAt, End: endAt}
}

func planSwingTiming(start, end, multiplier float64) *proto.PlanBossAbility {
	ability := planBossWindow(proto.PlanBossAbility_SwingTiming, start, end)
	ability.SwingSpeedMultiplier = multiplier
	return ability
}

func planThrash(start, end float64) *proto.PlanBossAbility {
	ability := planBossWindow(proto.PlanBossAbility_Thrash, start, end)
	ability.ProcChance = 100
	ability.ExtraAttacks = 1
	return ability
}

func TestPlanSwingTimingScalesOrPausesTheSwing(t *testing.T) {
	swingTiming := func(multiplier float64) *proto.EncounterPlan {
		return &proto.EncounterPlan{BossAbilities: []*proto.PlanBossAbility{planSwingTiming(0.2, 0.4, multiplier)}}
	}

	t.Run("doubled", func(t *testing.T) {
		_, hits := runPlanHitsTaken(t, planTankRequest(swingTiming(2)))
		inside := bossSwingsBetween(hits, planAt(0.2), planAt(0.4))
		outside := bossSwingsBetween(hits, planAt(0.5), planAt(0.7))
		if outside == 0 || inside < 2*outside-1 || inside > 2*outside+1 {
			t.Errorf("expected twice the %d swings of a window of the same length outside, got %d", outside, inside)
		}
	})

	t.Run("paused", func(t *testing.T) {
		_, hits := runPlanHitsTaken(t, planTankRequest(swingTiming(0)))
		if got := bossSwingsBetween(hits, planAt(0.2), planAt(0.4)); got != 0 {
			t.Errorf("expected no swing inside the paused window, got %d", got)
		}
		if got := bossSwingsBetween(hits, planAt(0.4)-time.Nanosecond, planAt(0.4)+2*planBossSwingSpeed); got == 0 {
			t.Error("expected the swing to resume at the window's end")
		}
	})
}

// A pause ending inside another ability's pause leaves the swing stopped until the last one ends.
func TestPlanOverlappingPausesHoldTheSwing(t *testing.T) {
	plan := &proto.EncounterPlan{BossAbilities: []*proto.PlanBossAbility{planSwingTiming(0.2, 0.6, 0), planSwingTiming(0.3, 0.4, 0)}}
	_, hits := runPlanHitsTaken(t, planTankRequest(plan))
	if got := bossSwingsBetween(hits, planAt(0.2), planAt(0.6)); got != 0 {
		t.Errorf("expected no swing while either pause runs, got %d", got)
	}
	if got := bossSwingsBetween(hits, planAt(0.6)-time.Nanosecond, planAt(0.6)+2*planBossSwingSpeed); got == 0 {
		t.Error("expected the swing to resume at the last pause's end")
	}
}

// A downtime that clears debuffs leaves the window auras, which are permanent, and the boss coming back
// inside a pause does not swing until the pause ends.
func TestPlanBossWindowsHoldThroughADowntime(t *testing.T) {
	downtime := []*proto.PlanDowntime{{TargetIndex: 0, Start: &proto.PlanTime{At: 0.3}, End: &proto.PlanTime{At: 0.4}, ClearDebuffs: true}}

	t.Run("paused", func(t *testing.T) {
		plan := &proto.EncounterPlan{Downtimes: downtime, BossAbilities: []*proto.PlanBossAbility{planSwingTiming(0.2, 0.6, 0)}}
		sim, hits := setupPlanHitsTaken(planTankRequest(plan), nil)
		pause := sim.Encounter.plan.bossAbilities[0].aura

		runLifecycleUntil(sim, planAt(0.45))
		if !sim.Encounter.AllTargets[0].IsEnabled() {
			t.Fatal("test setup: expected the boss back after its downtime")
		}
		if !pause.IsActive() {
			t.Error("the downtime that clears debuffs took the pause's aura off the boss")
		}
		runLifecycleUntil(sim, sim.Duration)
		if got := bossSwingsBetween(*hits, planAt(0.2), planAt(0.6)); got != 0 {
			t.Errorf("expected no swing inside the paused window, the downtime's end included; got %d", got)
		}
		if got := bossSwingsBetween(*hits, planAt(0.6)-time.Nanosecond, planAt(0.6)+2*planBossSwingSpeed); got == 0 {
			t.Error("expected the swing to resume at the pause's end")
		}
	})

	t.Run("thrash", func(t *testing.T) {
		plan := &proto.EncounterPlan{Downtimes: downtime, BossAbilities: []*proto.PlanBossAbility{planThrash(0.2, 0.6)}}
		sim, hits := setupPlanHitsTaken(planTankRequest(plan), nil)
		thrash := sim.Encounter.plan.bossAbilities[0].aura

		runLifecycleUntil(sim, planAt(0.45))
		if !thrash.IsActive() {
			t.Error("the downtime that clears debuffs took the thrash's aura off the boss")
		}
		runLifecycleUntil(sim, sim.Duration)
		extra := planHitsOf(*hits, func(id ActionID) bool {
			return id.OtherID == proto.OtherAction_OtherActionAttack && id.Tag == planBossAbilityTag
		})
		after := 0
		for _, hit := range extra {
			if hit.at > planAt(0.4) && hit.at < planAt(0.6) {
				after++
			}
		}
		if after == 0 {
			t.Error("expected extra attacks after the downtime, inside the window")
		}
	})
}

// A strike armed at the fight's end and a window still open there are gone by the next iteration, which
// plays out as the first did, and the boss's own swing replacement comes back each time no strike is armed.
func TestPlanBossAbilitiesStartEachIterationClean(t *testing.T) {
	strike := func(start float64) *proto.PlanBossAbility {
		ability := planBossWindow(proto.PlanBossAbility_Strike, start, 1)
		ability.WeaponPercent = 150
		return ability
	}
	plan := &proto.EncounterPlan{BossAbilities: []*proto.PlanBossAbility{
		strike(0.1),
		strike(0.9999),
		planSwingTiming(0.9, 1, 2),
	}}
	ownReplacements := 0
	sim, hits := setupPlanHitsTaken(planTankRequest(plan), func(char *Character) {
		char.Env.Encounter.AllTargets[0].AutoAttacks.SetReplaceMHSwing(func(_ *Simulation, spell *Spell) *Spell {
			ownReplacements++
			return spell
		})
	})
	boss := sim.Encounter.AllTargets[0]
	abilities := sim.Encounter.plan.bossAbilities
	lateStrike, window := abilities[1], abilities[2]
	isStrike := func(id ActionID) bool { return id.SpellID == planStrikeSpellID }

	var strikes [2]int
	for iteration := range 2 {
		*hits = (*hits)[:0]
		ownReplacements = 0
		runLifecycleUntil(sim, sim.Duration)
		strikes[iteration] = len(planHitsOf(*hits, isStrike))
		if !lateStrike.armed || !lateStrike.striker.installed {
			t.Fatalf("iteration %d, test setup: expected the last strike armed at the fight's end", iteration+1)
		}
		if !window.aura.IsActive() || window.running != 1 {
			t.Fatalf("iteration %d, test setup: expected the swing timing window open at the fight's end", iteration+1)
		}
		if ownReplacements == 0 {
			t.Errorf("iteration %d: the boss's own swing replacement never ran", iteration+1)
		}

		sim.Cleanup()
		if window.aura.IsActive() {
			t.Errorf("iteration %d: the window open at the fight's end is still up after it", iteration+1)
		}
		if got := boss.PseudoStats.MeleeSpeedMultiplier; got != 1 {
			t.Errorf("iteration %d: melee speed multiplier after the fight %.2f, expected 1", iteration+1, got)
		}
		sim.Reset()
		sim.PrePull()
		if lateStrike.armed || lateStrike.striker.installed || window.running != 0 {
			t.Errorf("iteration %d: reset left armed %t, installed %t, running %d", iteration+1,
				lateStrike.armed, lateStrike.striker.installed, window.running)
		}
		if boss.AutoAttacks.mh.replaceSwing == nil {
			t.Errorf("iteration %d: reset took the boss's own swing replacement off", iteration+1)
		}
	}
	if strikes[0] != 1 || strikes[1] != strikes[0] {
		t.Errorf("strikes per iteration: expected 1 and 1, got %d and %d", strikes[0], strikes[1])
	}
}

// Repeating strikes, swing timing windows, pulses and area hits allocate nothing per iteration beyond what the
// plan without them does. A spell hit's hardcast and a thrash's extra attack allocate in the cast and extra
// attack code they go through, so they stay out of the count. The area damage is counted on a player who does not
// tank, since every hit a tank takes goes into a list that grows from empty each iteration, and deals 1, since a
// death allocates.
func TestPlanBossAbilitiesAllocateNothingPerIteration(t *testing.T) {
	allocsPerIteration := func(request func(*proto.EncounterPlan) *proto.RaidSimRequest, abilities []*proto.PlanBossAbility) float64 {
		plan := &proto.EncounterPlan{
			Downtimes:     []*proto.PlanDowntime{{TargetIndex: 0, Start: &proto.PlanTime{At: 0.95}, End: &proto.PlanTime{At: 0.97}}},
			BossAbilities: abilities,
		}
		sim, _ := setupPlanSim(request(plan), nil)
		return testing.AllocsPerRun(20, func() {
			runLifecycleUntil(sim, sim.Duration)
			sim.Cleanup()
			sim.Reset()
			sim.PrePull()
		})
	}
	strike := planBossWindow(proto.PlanBossAbility_Strike, 0.1, 1)
	strike.WeaponPercent = 150
	strike.Repeat = &proto.PlanRepeat{Interval: 0.05}
	speed := planSwingTiming(0.3, 0.35, 2)
	speed.Repeat = &proto.PlanRepeat{Interval: 0.2}
	pause := planSwingTiming(0.37, 0.4, 0)
	pause.Repeat = &proto.PlanRepeat{Interval: 0.2}

	baseline := allocsPerIteration(planTankRequest, nil)
	if got := allocsPerIteration(planTankRequest, []*proto.PlanBossAbility{strike, speed, pause}); got > baseline {
		t.Errorf("allocations per iteration: %.1f with the melee abilities, %.1f without", got, baseline)
	}

	pulse := planPulse(0.5, 0.6, 1)
	pulse.TickJitter = 0.2
	pulse.Repeat = &proto.PlanRepeat{Interval: 0.2}
	areaHit := planAreaHit(0.45)
	areaHit.Repeat = &proto.PlanRepeat{Interval: 0.1}
	for _, area := range []*proto.PlanBossAbility{pulse, areaHit} {
		area.MinDamage = 1
		area.MaxDamage = 1
	}
	untanked := func(plan *proto.EncounterPlan) *proto.RaidSimRequest { return planTestRequest(1, plan) }
	baseline = allocsPerIteration(untanked, nil)
	if got := allocsPerIteration(untanked, []*proto.PlanBossAbility{pulse, areaHit}); got > baseline {
		t.Errorf("allocations per iteration: %.1f with the area damage, %.1f without", got, baseline)
	}
}

func init() {
	RegisterAgentFactory(
		proto.Player_RestorationShaman{},
		proto.Spec_SpecRestorationShaman,
		newPlanPetOwner,
		func(player *proto.Player, spec interface{}) {
			player.Spec = spec.(*proto.Player_RestorationShaman)
		},
	)
}

// A player with one pet out from the start, which stands still.
type planPetOwner struct {
	Character
	pet *planPet
}

type planPet struct {
	Pet
}

func newPlanPetOwner(char *Character, _ *proto.Player, _ *proto.Raid) Agent {
	owner := &planPetOwner{Character: *char}
	owner.pet = &planPet{Pet: NewPet(PetConfig{
		Name:            "Plan Pet",
		Owner:           &owner.Character,
		StatInheritance: func(stats.Stats) stats.Stats { return stats.Stats{} },
		EnabledOnStart:  true,
	})}
	owner.AddPet(owner.pet)
	return owner
}

func (owner *planPetOwner) GetCharacter() *Character       { return &owner.Character }
func (owner *planPetOwner) ApplyTalents()                  {}
func (owner *planPetOwner) Initialize()                    {}
func (owner *planPetOwner) Reset(_ *Simulation)            {}
func (owner *planPetOwner) OnGCDReady(_ *Simulation)       {}
func (owner *planPetOwner) OnEncounterStart(_ *Simulation) {}

func (pet *planPet) GetPet() *Pet                        { return &pet.Pet }
func (pet *planPet) Reset(_ *Simulation)                 {}
func (pet *planPet) OnEncounterStart(_ *Simulation)      {}
func (pet *planPet) ExecuteCustomRotation(_ *Simulation) {}

// A sim of any agent, reset and through the prepull; register runs as setupPlanSim's does.
func setupPlanAgentSim(request *proto.RaidSimRequest, register func(char *Character)) (*Simulation, *Character) {
	sim := NewSim(request, simsignals.CreateSignals())
	char := sim.Raid.Parties[0].Players[0].GetCharacter()
	if register != nil {
		sim.Environment.State = Constructed
		register(char)
		sim.Environment.State = Finalized
	}
	sim.Reset()
	sim.PrePull()
	return sim, char
}

func planPulse(start, end, interval float64) *proto.PlanBossAbility {
	ability := planBossWindow(proto.PlanBossAbility_Pulse, start, end)
	ability.School = proto.SpellSchool_SpellSchoolFire
	ability.MinDamage = 500
	ability.MaxDamage = 700
	ability.TickInterval = interval
	return ability
}

func planAreaHit(start float64) *proto.PlanBossAbility {
	ability := &proto.PlanBossAbility{Kind: proto.PlanBossAbility_AreaHit, Start: &proto.PlanTime{At: start}}
	ability.School = proto.SpellSchool_SpellSchoolFire
	ability.MinDamage = 500
	ability.MaxDamage = 700
	return ability
}

func planAreaHitsOf(hits []planHitTaken, index int) []planHitTaken {
	tag := planBossAbilityTag + int32(index)
	return planHitsOf(hits, func(id ActionID) bool { return id.SpellID == planSpellHitSpellID && id.Tag == tag })
}

func expectPlanHitTimes(t *testing.T, name string, hits []planHitTaken, want []time.Duration) {
	t.Helper()
	if len(hits) != len(want) {
		t.Fatalf("%s: expected %d hits at %v, got %d: %v", name, len(want), want, len(hits), hits)
	}
	for i, hit := range hits {
		if hit.at != want[i] {
			t.Errorf("%s, hit %d: expected at %s, got %s", name, i+1, want[i], hit.at)
		}
	}
}

func planTicks(start time.Duration, interval time.Duration, count int) []time.Duration {
	ticks := make([]time.Duration, count)
	for k := range ticks {
		ticks[k] = start + time.Duration(k)*interval
	}
	return ticks
}

// The first tick lands at the window's start and the last one before its end, each repeat opening a window of
// the same length; a window shorter than one interval ticks once.
func TestPlanPulseTicksFromItsWindowStartUntilItsEnd(t *testing.T) {
	repeating := planPulse(0.1, 0.2, 3)
	repeating.Repeat = &proto.PlanRepeat{Interval: 0.5, Count: 1}
	plan := &proto.EncounterPlan{BossAbilities: []*proto.PlanBossAbility{repeating, planPulse(0.3, 0.301, 2)}}
	_, hits := runPlanHitsTaken(t, planTestRequest(1, plan))

	ticks := planAreaHitsOf(hits, 0)
	expectPlanHitTimes(t, "repeating pulse", ticks,
		append(planTicks(planAt(0.1), 3*time.Second, 4), planTicks(planAt(0.6), 3*time.Second, 4)...))
	for i, tick := range ticks {
		if tick.damage < 500 || tick.damage > 700 {
			t.Errorf("tick %d: damage %.1f outside [500, 700]", i+1, tick.damage)
		}
	}
	expectPlanHitTimes(t, "short pulse", planAreaHitsOf(hits, 1), []time.Duration{planAt(0.3)})
}

// Environment area damage comes from no target: it lands while the boss is in a downtime.
func TestPlanAreaDamageLandsWhileTheBossIsDown(t *testing.T) {
	areaHit := planAreaHit(0.2)
	areaHit.Repeat = &proto.PlanRepeat{Interval: 0.1, Count: 2}
	plan := &proto.EncounterPlan{
		Downtimes:     []*proto.PlanDowntime{{TargetIndex: 0, Start: &proto.PlanTime{At: 0.05}, End: &proto.PlanTime{At: 0.5}}},
		BossAbilities: []*proto.PlanBossAbility{planPulse(0.1, 0.15, 1), areaHit},
	}
	sim, hits := setupPlanHitsTaken(planTestRequest(1, plan), nil)
	runLifecycleUntil(sim, planAt(0.1))
	if sim.Encounter.AllTargets[0].IsEnabled() || len(sim.Encounter.ActiveTargets) != 0 {
		t.Fatal("test setup: expected no active target at the pulse's start")
	}
	runLifecycleUntil(sim, sim.Duration)

	expectPlanHitTimes(t, "pulse", planAreaHitsOf(*hits, 0), planTicks(planAt(0.1), time.Second, 5))
	expectPlanHitTimes(t, "area hit", planAreaHitsOf(*hits, 1), []time.Duration{planAt(0.2), planAt(0.3), planAt(0.4)})
}

// A spell hit, a strike and an area hit whose end comes before their first repeat still repeat: count times, or
// with a count of 0 until the fight ends.
func TestPlanInstantAbilitiesRepeatPastTheirEnd(t *testing.T) {
	for _, c := range []struct {
		count int32
		want  int
	}{{count: 2, want: 3}, {count: 0, want: 5}} {
		instant := func(kind proto.PlanBossAbility_Kind) *proto.PlanBossAbility {
			startAt, endAt := planWindow(0.1, 0.15)
			return &proto.PlanBossAbility{
				Kind:          kind,
				Start:         startAt,
				End:           endAt,
				Repeat:        &proto.PlanRepeat{Interval: 0.2, Count: c.count},
				School:        proto.SpellSchool_SpellSchoolFire,
				MinDamage:     100,
				MaxDamage:     100,
				WeaponPercent: 150,
			}
		}
		plan := &proto.EncounterPlan{BossAbilities: []*proto.PlanBossAbility{
			instant(proto.PlanBossAbility_SpellHit),
			instant(proto.PlanBossAbility_AreaHit),
			instant(proto.PlanBossAbility_Strike),
		}}
		_, hits := runPlanHitsTaken(t, planTankRequest(plan))

		want := make([]time.Duration, c.want)
		for k := range want {
			want[k] = planAt(0.1 + float64(k)*0.2)
		}
		expectPlanHitTimes(t, fmt.Sprintf("count %d, spell hit", c.count), planAreaHitsOf(hits, 0), want)
		expectPlanHitTimes(t, fmt.Sprintf("count %d, area hit", c.count), planAreaHitsOf(hits, 1), want)
		if strikes := planHitsOf(hits, func(id ActionID) bool { return id.SpellID == planStrikeSpellID }); len(strikes) != c.want {
			t.Errorf("count %d: expected %d strikes, got %d", c.count, c.want, len(strikes))
		}
	}
}

// A tick is damage taken from an enemy spell: a warrior gains rage from it with no target to swing at, and it
// counts in the damage taken metrics.
func TestPlanPulseTickGrantsRageAndCountsAsDamageTaken(t *testing.T) {
	request := planTestRequest(1, &proto.EncounterPlan{
		Downtimes:     []*proto.PlanDowntime{{TargetIndex: 0, Start: &proto.PlanTime{At: 0.05}, End: &proto.PlanTime{At: 0.5}}},
		BossAbilities: []*proto.PlanBossAbility{planPulse(0.1, 0.2, 2)},
	})
	request.Raid.Parties[0].Players[0].Spec = &proto.Player_DpsWarrior{}
	sim, char := setupPlanAgentSim(request, nil)

	runLifecycleUntil(sim, planAt(0.1)-time.Millisecond)
	if len(sim.Encounter.ActiveTargets) != 0 {
		t.Fatal("test setup: expected no active target before the first tick")
	}
	before := char.CurrentRage()
	if before >= char.MaximumRage() {
		t.Fatalf("test setup: expected room for rage before the first tick, got %.2f", before)
	}
	runLifecycleUntil(sim, planAt(0.1))
	if got := char.CurrentRage(); got <= before {
		t.Errorf("expected rage from the first tick: %.2f before it, %.2f after", before, got)
	}
	runLifecycleUntil(sim, sim.Duration)
	sim.Cleanup()
	if got := char.Metrics.ToProto().Dtps.Avg; got <= 0 {
		t.Errorf("expected the ticks in the damage taken metrics, got %.2f DTPS", got)
	}
}

// Against a high fire resistance a resistible fire hit loses damage to partial resists; one that is not
// resistible lands every time for its full damage.
func TestPlanAreaDamageResistsOnlyWhenResistible(t *testing.T) {
	areaHit := func(resistible bool) *proto.PlanBossAbility {
		ability := planAreaHit(0.01)
		ability.Repeat = &proto.PlanRepeat{Interval: 0.01}
		ability.MinDamage = 1000
		ability.MaxDamage = 1000
		ability.Resistible = resistible
		return ability
	}
	plan := &proto.EncounterPlan{BossAbilities: []*proto.PlanBossAbility{areaHit(true), areaHit(false)}}
	sim, hits := setupPlanHitsTaken(planTestRequest(1, plan), func(char *Character) {
		MakePermanent(char.RegisterAura(Aura{
			Label: "Fire Resistance",
			OnGain: func(aura *Aura, sim *Simulation) {
				aura.Unit.AddStatDynamic(sim, stats.FireResistance, 300)
			},
			OnExpire: func(aura *Aura, sim *Simulation) {
				aura.Unit.AddStatDynamic(sim, stats.FireResistance, -300)
			},
		}))
	})
	runLifecycleUntil(sim, sim.Duration)

	resistible, plain := planAreaHitsOf(*hits, 0), planAreaHitsOf(*hits, 1)
	if len(resistible) < 90 || len(plain) != len(resistible) {
		t.Fatalf("test setup: expected about 100 hits of each, got %d resistible and %d not", len(resistible), len(plain))
	}
	resisted, total := 0, 0.0
	for _, hit := range resistible {
		total += hit.damage
		if hit.outcome.Matches(OutcomePartial) {
			resisted++
		}
	}
	if resisted == 0 || total >= 1000*float64(len(resistible)) {
		t.Errorf("resistible: expected partial resists, got %d resisted of %d for %.0f damage", resisted, len(resistible), total)
	}
	for i, hit := range plain {
		if hit.damage != 1000 || !hit.outcome.Matches(OutcomeHit) || hit.outcome.Matches(OutcomePartial) {
			t.Errorf("not resistible, hit %d: expected a full hit of 1000, got %.1f (%s)", i+1, hit.damage, hit.outcome)
		}
	}
}

// An avoidable tick skips the player from the start of a plan's move to its end and still hits the pet; one that
// is not avoidable hits both throughout.
func TestPlanAvoidableAreaDamageSkipsAWalkingPlayerOnly(t *testing.T) {
	avoidable := planPulse(0.15, 0.3, 1)
	avoidable.Avoidable = true
	avoidableHit := planAreaHit(0.2)
	avoidableHit.Avoidable = true
	// The move's walk at 0.2 is its repeat, queued once the walk at 0.1 runs, after the area hit at 0.2 was.
	plan := &proto.EncounterPlan{
		Moves:         []*proto.PlanMove{{At: &proto.PlanTime{At: 0.1}, Seconds: 5, Repeat: &proto.PlanRepeat{Interval: 0.1, Count: 1}}},
		BossAbilities: []*proto.PlanBossAbility{avoidable, planPulse(0.15, 0.3, 1), avoidableHit},
	}
	request := planTestRequest(1, plan)
	request.Raid.Parties[0].Players[0].Class = proto.Class_ClassShaman
	request.Raid.Parties[0].Players[0].Spec = &proto.Player_RestorationShaman{}
	var playerHits, petHits []planHitTaken
	sim, char := setupPlanAgentSim(request, func(char *Character) {
		recordPlanHitsTaken(&char.Unit, &playerHits)
		recordPlanHitsTaken(&char.Pets[0].Unit, &petHits)
	})
	if !char.Pets[0].IsEnabled() {
		t.Fatal("test setup: expected the pet out")
	}
	runLifecycleUntil(sim, sim.Duration)

	all := planTicks(planAt(0.15), time.Second, 15)
	expectPlanHitTimes(t, "avoidable, player", planAreaHitsOf(playerHits, 0), append(slices.Clone(all[:5]), all[10:]...))
	expectPlanHitTimes(t, "avoidable, pet", planAreaHitsOf(petHits, 0), all)
	expectPlanHitTimes(t, "not avoidable, player", planAreaHitsOf(playerHits, 1), all)
	expectPlanHitTimes(t, "not avoidable, pet", planAreaHitsOf(petHits, 1), all)
	expectPlanHitTimes(t, "avoidable area hit at the move's start, player", planAreaHitsOf(playerHits, 2), nil)
	expectPlanHitTimes(t, "avoidable area hit at the move's start, pet", planAreaHitsOf(petHits, 2), []time.Duration{planAt(0.2)})
}

// Tick jitter is rolled from the plan rand: each tick after the first lands within the jitter of its place
// on the interval, the same for a seed and for other gear, and moves between iterations.
func TestPlanPulseTickJitterIsDeterministicAndIndependentOfGear(t *testing.T) {
	const jitter = 500 * time.Millisecond
	pulse := planPulse(0.1, 0.3, 2)
	pulse.TickJitter = jitter.Seconds()
	plan := &proto.EncounterPlan{BossAbilities: []*proto.PlanBossAbility{pulse}}

	run := func(bonus stats.Stats) ([][]time.Duration, float64) {
		request := planTankRequest(plan)
		request.Raid.Parties[0].Players[0].BonusStats = &proto.UnitStats{Stats: bonus.ToProtoArray()}
		sim, hits := setupPlanHitsTaken(request, nil)
		sim.Cleanup()
		// One seed for every iteration, so the combat rolls of each one, which the gear changes, run on into
		// the next one's reset.
		sim.reseedRands(0)
		var iterations [][]time.Duration
		for range 10 {
			*hits = (*hits)[:0]
			sim.reset()
			sim.PrePull()
			sim.runPendingActions()
			sim.Cleanup()
			var ticks []time.Duration
			for _, hit := range planAreaHitsOf(*hits, 0) {
				ticks = append(ticks, hit.at)
			}
			iterations = append(iterations, ticks)
		}
		return iterations, sim.Raid.Parties[0].Players[0].GetCharacter().Metrics.ToProto().Dps.Avg
	}

	base, baseDamage := run(stats.Stats{})
	again, _ := run(stats.Stats{})
	geared, gearedDamage := run(stats.Stats{stats.AttackPower: 1000, stats.HasteRating: 500})
	if baseDamage == gearedDamage {
		t.Fatal("test setup: the bonus stats deal the same damage, so the gear change moves nothing")
	}

	nominal := planTicks(planAt(0.1), 2*time.Second, 10)
	moved := false
	for i, ticks := range base {
		if !slices.Equal(ticks, again[i]) {
			t.Errorf("iteration %d: the same seed rolled other ticks", i)
		}
		if !slices.Equal(ticks, geared[i]) {
			t.Errorf("iteration %d: other stats rolled other ticks: %v against %v", i, ticks, geared[i])
		}
		if len(ticks) != len(nominal) {
			t.Fatalf("iteration %d: expected %d ticks, got %d: %v", i, len(nominal), len(ticks), ticks)
		}
		if ticks[0] != nominal[0] {
			t.Errorf("iteration %d: the first tick at %s, expected the window's start %s", i, ticks[0], nominal[0])
		}
		for k := 1; k < len(ticks); k++ {
			if ticks[k] < nominal[k]-jitter || ticks[k] > nominal[k]+jitter {
				t.Errorf("iteration %d, tick %d: at %s, outside %s ± %s", i, k+1, ticks[k], nominal[k], jitter)
			}
		}
		if i > 0 && !slices.Equal(ticks, base[0]) {
			moved = true
		}
	}
	if !moved {
		t.Error("expected the ticks to move between iterations")
	}
}

// Area damage needs damage and a pulse an interval; the target index plays no part in an Environment one, and a
// Boss one naming a target the encounter lacks is skipped.
func TestPlanAreaDamageSkipsWithoutDamageOrInterval(t *testing.T) {
	noInterval := planPulse(0.1, 0.2, 0)
	noDamage := planAreaHit(0.1)
	noDamage.MaxDamage = 0
	anyTarget := planAreaHit(0.1)
	anyTarget.TargetIndex = 3
	missingCaster := planBossCast(planAreaHit(0.1), 0)
	missingCaster.TargetIndex = 3
	plan := &proto.EncounterPlan{BossAbilities: []*proto.PlanBossAbility{noInterval, noDamage, anyTarget, missingCaster}}
	sim, hits := setupPlanHitsTaken(planTestRequest(1, plan), nil)
	runLifecycleUntil(sim, sim.Duration)

	abilities := sim.Encounter.plan.bossAbilities
	if !abilities[0].skip || !abilities[1].skip || abilities[2].skip || !abilities[3].skip {
		t.Errorf("skipped: expected true, true, false, true; got %t, %t, %t, %t", abilities[0].skip, abilities[1].skip, abilities[2].skip, abilities[3].skip)
	}
	notes := strings.Join(sim.Encounter.plan.constructNotes, " ")
	for _, want := range []string{"boss ability 1 ticks at no interval", "boss ability 2 deals no damage", "boss ability 4 names target 3"} {
		if !strings.Contains(notes, want) {
			t.Errorf("construct notes %q: expected %q", notes, want)
		}
	}
	expectPlanHitTimes(t, "area hit naming target 3", planAreaHitsOf(*hits, 2), []time.Duration{planAt(0.1)})
}

// The ability cast by its target, the boss, with the cast time.
func planBossCast(ability *proto.PlanBossAbility, castTime time.Duration) *proto.PlanBossAbility {
	ability.Source = proto.PlanBossAbility_Boss
	ability.CastTime = castTime.Seconds()
	return ability
}

// A spell hit and a Boss area hit land when the boss's hardcast completes. The boss does not swing during the cast
// and starts a full swing when it completes, unless the ability keeps the swing timer, when it swings through it.
func TestPlanBossHardcastHoldsTheSwingUnlessItKeepsTheTimer(t *testing.T) {
	const castTime = 5 * time.Second
	castEnd := planAt(0.2) + castTime
	for _, c := range []struct {
		name    string
		ability func() *proto.PlanBossAbility
	}{
		{"spell hit", func() *proto.PlanBossAbility { return planNamedSpellHit("", 0, 0.2, castTime) }},
		{"Boss area hit", func() *proto.PlanBossAbility { return planBossCast(planAreaHit(0.2), castTime) }},
	} {
		for _, keeps := range []bool{false, true} {
			name := fmt.Sprintf("%s, keeps the swing timer %t", c.name, keeps)
			ability := c.ability()
			ability.KeepsSwingTimer = keeps
			_, hits := runPlanHitsTaken(t, planTankRequest(&proto.EncounterPlan{BossAbilities: []*proto.PlanBossAbility{ability}}))

			expectPlanHitTimes(t, name, planAreaHitsOf(hits, 0), []time.Duration{castEnd})
			during := bossSwingsBetween(hits, planAt(0.2), castEnd)
			if keeps {
				if during < 2 {
					t.Errorf("%s: expected the boss to swing through the cast, got %d swings", name, during)
				}
				continue
			}
			if during != 0 {
				t.Errorf("%s: expected no swing during the cast, got %d", name, during)
			}
			swings := planHitsOf(hits, isBossMeleeSwing)
			next := slices.IndexFunc(swings, func(hit planHitTaken) bool { return hit.at >= castEnd })
			if next < 0 || swings[next].at != castEnd+planBossSwingSpeed {
				t.Errorf("%s: expected the first swing after the cast one full swing after it, at %s; got %v", name, castEnd+planBossSwingSpeed, swings)
			}
		}
	}
}

// A Boss pulse's first tick lands when its cast completes, the others every interval after it to the window's end,
// and the boss swings again once the cast is done. A cast that would end past the window's end ticks nothing, and a
// window opening while the boss still casts the one before ticks nothing while the one before ticks on.
func TestPlanBossPulseTicksFromItsCastsEnd(t *testing.T) {
	const castTime = 2 * time.Second
	repeating := planBossCast(planPulse(0.1, 0.2, 3), castTime)
	repeating.Repeat = &proto.PlanRepeat{Interval: 0.5, Count: 1}
	pastEnd := planBossCast(planPulse(0.3, 0.315, 1), castTime)
	overlapping := planBossCast(planPulse(0.4, 0.6, 3), castTime)
	overlapping.Repeat = &proto.PlanRepeat{Interval: 0.01, Count: 1}
	plan := &proto.EncounterPlan{BossAbilities: []*proto.PlanBossAbility{repeating, pastEnd, overlapping}}
	_, hits := runPlanHitsTaken(t, planTankRequest(plan))

	expectPlanHitTimes(t, "repeating pulse", planAreaHitsOf(hits, 0),
		append(planTicks(planAt(0.1)+castTime, 3*time.Second, 3), planTicks(planAt(0.6)+castTime, 3*time.Second, 3)...))
	expectPlanHitTimes(t, "pulse whose cast ends past its window", planAreaHitsOf(hits, 1), nil)
	expectPlanHitTimes(t, "pulse whose second window opens mid-cast", planAreaHitsOf(hits, 2), planTicks(planAt(0.4)+castTime, 3*time.Second, 6))
	if got := bossSwingsBetween(hits, planAt(0.1), planAt(0.1)+castTime); got != 0 {
		t.Errorf("expected no swing during the cast, got %d", got)
	}
	if got := bossSwingsBetween(hits, planAt(0.1)+castTime, planAt(0.2)); got == 0 {
		t.Error("expected the boss to swing in the window after its cast")
	}
}

// A Boss area hit or pulse window does nothing while its target is down, where an Environment one lands. A downtime
// that starts mid-cast stops the cast, so its hit, or its window's ticks, never land, and a Boss pulse's tick due
// while its target is down does not land.
func TestPlanBossAreaDamageNeedsItsTargetUp(t *testing.T) {
	const castTime = 2 * time.Second
	repeat := &proto.PlanRepeat{Interval: 0.2, Count: 2}
	environment := planAreaHit(0.2)
	environment.Repeat = repeat
	boss := planBossCast(planAreaHit(0.2), 0)
	boss.Repeat = repeat
	plan := &proto.EncounterPlan{
		Downtimes: []*proto.PlanDowntime{
			{TargetIndex: 0, Start: &proto.PlanTime{At: 0.05}, End: &proto.PlanTime{At: 0.5}},
			{TargetIndex: 0, Start: &proto.PlanTime{At: 0.705}, End: &proto.PlanTime{At: 0.71}},
			{TargetIndex: 0, Start: &proto.PlanTime{At: 0.805}, End: &proto.PlanTime{At: 0.81}},
		},
		BossAbilities: []*proto.PlanBossAbility{
			environment,
			boss,
			planBossCast(planPulse(0.1, 0.15, 1), 0),
			planBossCast(planPulse(0.02, 0.08, 0.7), 0),
			planBossCast(planAreaHit(0.7), castTime),
			planBossCast(planPulse(0.8, 0.9, 1), castTime),
		},
	}
	_, hits := runPlanHitsTaken(t, planTestRequest(1, plan))

	expectPlanHitTimes(t, "Environment area hit", planAreaHitsOf(hits, 0), []time.Duration{planAt(0.2), planAt(0.4), planAt(0.6)})
	expectPlanHitTimes(t, "Boss area hit", planAreaHitsOf(hits, 1), []time.Duration{planAt(0.6)})
	expectPlanHitTimes(t, "Boss pulse starting in the downtime", planAreaHitsOf(hits, 2), nil)
	expectPlanHitTimes(t, "Boss pulse running into the downtime", planAreaHitsOf(hits, 3), planTicks(planAt(0.02), 700*time.Millisecond, 5))
	expectPlanHitTimes(t, "Boss area hit whose cast a downtime stops", planAreaHitsOf(hits, 4), nil)
	expectPlanHitTimes(t, "Boss pulse whose cast a downtime stops", planAreaHitsOf(hits, 5), nil)
}

package core

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

const designBossSwingSpeed = 2 * time.Second

// A request whose boss swings at the fake Protection Warrior, its tank.
func designTankRequest(design *proto.EncounterDesign) *proto.RaidSimRequest {
	request := designTestRequest(1, design)
	request.Raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	boss := request.Encounter.Targets[0]
	boss.SwingSpeed = designBossSwingSpeed.Seconds()
	boss.MinBaseDamage = 1000
	return request
}

type designHitTaken struct {
	at       time.Duration
	actionID ActionID
	damage   float64
}

// A sim at the pull that records every hit the player takes; register runs as setupDesignSim's does.
func setupDesignHitsTaken(request *proto.RaidSimRequest, register func(char *Character)) (*Simulation, *[]designHitTaken) {
	hits := &[]designHitTaken{}
	sim, _ := setupDesignSim(request, func(char *Character) {
		MakePermanent(char.RegisterAura(Aura{
			Label: "Hits Taken",
			OnSpellHitTaken: func(_ *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
				*hits = append(*hits, designHitTaken{at: sim.CurrentTime, actionID: spell.ActionID, damage: result.Damage})
			},
		}))
		if register != nil {
			register(char)
		}
	})
	return sim, hits
}

// Every hit the player takes over one iteration.
func runDesignHitsTaken(t *testing.T, request *proto.RaidSimRequest) (*Simulation, []designHitTaken) {
	t.Helper()
	sim, hits := setupDesignHitsTaken(request, nil)
	runLifecycleUntil(sim, sim.Duration)
	return sim, *hits
}

func bossSwingsBetween(hits []designHitTaken, from, to time.Duration) int {
	count := 0
	for _, hit := range hits {
		if isBossMeleeSwing(hit.actionID) && hit.at > from && hit.at < to {
			count++
		}
	}
	return count
}

func designHitsOf(hits []designHitTaken, matches func(ActionID) bool) []designHitTaken {
	var matched []designHitTaken
	for _, hit := range hits {
		if matches(hit.actionID) {
			matched = append(matched, hit)
		}
	}
	return matched
}

func isBossMeleeSwing(id ActionID) bool {
	return id.OtherID == proto.OtherAction_OtherActionAttack && id.Tag == 1
}

func designWindow(start, end float64) (*proto.DesignTime, *proto.DesignTime) {
	return &proto.DesignTime{At: start}, &proto.DesignTime{At: end}
}

func TestDesignSpellHitLandsAtEachOccurrenceAfterItsHardcast(t *testing.T) {
	const minDamage, maxDamage = 1000.0, 2000.0
	const castTime = 2 * time.Second
	spellHit := func(start, end float64, repeat *proto.DesignRepeat) *proto.DesignBossAbility {
		startAt, endAt := designWindow(start, end)
		return &proto.DesignBossAbility{
			Kind:      proto.DesignBossAbility_SpellHit,
			Start:     startAt,
			End:       endAt,
			Repeat:    repeat,
			School:    proto.SpellSchool_SpellSchoolFire,
			MinDamage: minDamage,
			MaxDamage: maxDamage,
			CastTime:  castTime.Seconds(),
		}
	}
	design := &proto.EncounterDesign{BossAbilities: []*proto.DesignBossAbility{
		spellHit(0.1, 0.35, &proto.DesignRepeat{Interval: 0.1}),
		spellHit(0.55, 1, &proto.DesignRepeat{Interval: 0.1, Count: 1}),
	}}
	_, hits := runDesignHitsTaken(t, designTestRequest(1, design))

	for index, want := range [][]time.Duration{
		{designAt(0.1) + castTime, designAt(0.2) + castTime, designAt(0.3) + castTime},
		{designAt(0.55) + castTime, designAt(0.65) + castTime},
	} {
		tag := designBossAbilityTag + int32(index)
		landed := designHitsOf(hits, func(id ActionID) bool { return id.SpellID == designSpellHitSpellID && id.Tag == tag })
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

func TestDesignBossAbilityOfADisabledTargetDoesNothing(t *testing.T) {
	startAt, endAt := designWindow(0.1, 0.35)
	design := &proto.EncounterDesign{
		Downtimes: []*proto.DesignDowntime{{TargetIndex: 0, Start: &proto.DesignTime{At: 0.15}, End: &proto.DesignTime{At: 0.25}}},
		BossAbilities: []*proto.DesignBossAbility{{
			Kind:      proto.DesignBossAbility_SpellHit,
			Start:     startAt,
			End:       endAt,
			Repeat:    &proto.DesignRepeat{Interval: 0.1},
			MinDamage: 100,
			MaxDamage: 100,
		}},
	}
	_, hits := runDesignHitsTaken(t, designTestRequest(1, design))

	landed := designHitsOf(hits, func(id ActionID) bool { return id.SpellID == designSpellHitSpellID })
	want := []time.Duration{designAt(0.1), designAt(0.3)}
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
func TestDesignSpellHitStopsWhenItsTargetGoesAway(t *testing.T) {
	const castTime = 4 * time.Second
	design := func(downtimes []*proto.DesignDowntime) *proto.EncounterDesign {
		startAt, endAt := designWindow(0.1, 0.2)
		return &proto.EncounterDesign{
			Downtimes: downtimes,
			BossAbilities: []*proto.DesignBossAbility{{
				Kind:      proto.DesignBossAbility_SpellHit,
				Start:     startAt,
				End:       endAt,
				Repeat:    &proto.DesignRepeat{Interval: 0.1},
				MinDamage: 100,
				MaxDamage: 100,
				CastTime:  castTime.Seconds(),
			}},
		}
	}
	spellHits := func(hits []designHitTaken) []designHitTaken {
		return designHitsOf(hits, func(id ActionID) bool { return id.SpellID == designSpellHitSpellID })
	}

	t.Run("downtime", func(t *testing.T) {
		downtime := []*proto.DesignDowntime{{TargetIndex: 0, Start: &proto.DesignTime{At: 0.11}, End: &proto.DesignTime{At: 0.12}}}
		_, hits := runDesignHitsTaken(t, designTestRequest(1, design(downtime)))
		landed := spellHits(hits)
		if len(landed) != 1 || landed[0].at != designAt(0.2)+castTime {
			t.Errorf("expected only the hit of the occurrence after the downtime, at %s; got %v", designAt(0.2)+castTime, landed)
		}
	})

	t.Run("death", func(t *testing.T) {
		sim, hits := setupDesignHitsTaken(designTestRequest(1, design(nil)), nil)
		boss := sim.Encounter.AllTargets[0]
		runLifecycleUntil(sim, designAt(0.11))
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

func TestDesignStrikeReplacesOneSwingPerOccurrence(t *testing.T) {
	startAt, endAt := designWindow(0.1, 0.35)
	design := &proto.EncounterDesign{BossAbilities: []*proto.DesignBossAbility{{
		Kind:          proto.DesignBossAbility_Strike,
		Start:         startAt,
		End:           endAt,
		Repeat:        &proto.DesignRepeat{Interval: 0.1},
		WeaponPercent: 150,
	}}}
	_, hits := runDesignHitsTaken(t, designTankRequest(design))

	isStrike := func(id ActionID) bool { return id.SpellID == designStrikeSpellID }
	attacks := designHitsOf(hits, func(id ActionID) bool { return isBossMeleeSwing(id) || isStrike(id) })
	if len(designHitsOf(attacks, isStrike)) != 3 {
		t.Fatalf("expected 3 strikes, got %d", len(designHitsOf(attacks, isStrike)))
	}
	for i := 1; i < len(attacks); i++ {
		if gap := attacks[i].at - attacks[i-1].at; gap != designBossSwingSpeed {
			t.Fatalf("attack %d comes %s after the one before, expected one swing of %s", i+1, gap, designBossSwingSpeed)
		}
	}
	for _, occurrence := range []float64{0.1, 0.2, 0.3} {
		next := 0
		for next < len(attacks) && attacks[next].at < designAt(occurrence) {
			next++
		}
		if next+1 >= len(attacks) {
			t.Fatalf("no two attacks after the occurrence at %s", designAt(occurrence))
		}
		if !isStrike(attacks[next].actionID) {
			t.Errorf("the first swing after the occurrence at %s is %s, expected the strike", designAt(occurrence), attacks[next].actionID)
		}
		if isStrike(attacks[next+1].actionID) {
			t.Errorf("the second swing after the occurrence at %s is a strike too", designAt(occurrence))
		}
	}
}

func TestDesignThrashSwingsExtraOnlyInsideItsWindow(t *testing.T) {
	startAt, endAt := designWindow(0.2, 0.4)
	design := &proto.EncounterDesign{BossAbilities: []*proto.DesignBossAbility{{
		Kind:         proto.DesignBossAbility_Thrash,
		Start:        startAt,
		End:          endAt,
		ProcChance:   100,
		ExtraAttacks: 1,
	}}}
	_, hits := runDesignHitsTaken(t, designTankRequest(design))

	extra := designHitsOf(hits, func(id ActionID) bool {
		return id.OtherID == proto.OtherAction_OtherActionAttack && id.Tag == designBossAbilityTag
	})
	if len(extra) == 0 {
		t.Fatal("expected extra attacks inside the window, got none")
	}
	for _, hit := range extra {
		if hit.at < designAt(0.2) || hit.at >= designAt(0.4) {
			t.Errorf("extra attack at %s, outside the window [%s, %s)", hit.at, designAt(0.2), designAt(0.4))
		}
	}
}

func designBossWindow(kind proto.DesignBossAbility_Kind, start, end float64) *proto.DesignBossAbility {
	startAt, endAt := designWindow(start, end)
	return &proto.DesignBossAbility{Kind: kind, Start: startAt, End: endAt}
}

func designSwingTiming(start, end, multiplier float64) *proto.DesignBossAbility {
	ability := designBossWindow(proto.DesignBossAbility_SwingTiming, start, end)
	ability.SwingSpeedMultiplier = multiplier
	return ability
}

func designThrash(start, end float64) *proto.DesignBossAbility {
	ability := designBossWindow(proto.DesignBossAbility_Thrash, start, end)
	ability.ProcChance = 100
	ability.ExtraAttacks = 1
	return ability
}

func TestDesignSwingTimingScalesOrPausesTheSwing(t *testing.T) {
	swingTiming := func(multiplier float64) *proto.EncounterDesign {
		return &proto.EncounterDesign{BossAbilities: []*proto.DesignBossAbility{designSwingTiming(0.2, 0.4, multiplier)}}
	}

	t.Run("doubled", func(t *testing.T) {
		_, hits := runDesignHitsTaken(t, designTankRequest(swingTiming(2)))
		inside := bossSwingsBetween(hits, designAt(0.2), designAt(0.4))
		outside := bossSwingsBetween(hits, designAt(0.5), designAt(0.7))
		if outside == 0 || inside < 2*outside-1 || inside > 2*outside+1 {
			t.Errorf("expected twice the %d swings of a window of the same length outside, got %d", outside, inside)
		}
	})

	t.Run("paused", func(t *testing.T) {
		_, hits := runDesignHitsTaken(t, designTankRequest(swingTiming(0)))
		if got := bossSwingsBetween(hits, designAt(0.2), designAt(0.4)); got != 0 {
			t.Errorf("expected no swing inside the paused window, got %d", got)
		}
		if got := bossSwingsBetween(hits, designAt(0.4)-time.Nanosecond, designAt(0.4)+2*designBossSwingSpeed); got == 0 {
			t.Error("expected the swing to resume at the window's end")
		}
	})
}

// A downtime that clears debuffs leaves the window auras, which are permanent, and the boss coming back
// inside a pause does not swing until the pause ends.
func TestDesignBossWindowsHoldThroughADowntime(t *testing.T) {
	downtime := []*proto.DesignDowntime{{TargetIndex: 0, Start: &proto.DesignTime{At: 0.3}, End: &proto.DesignTime{At: 0.4}, ClearDebuffs: true}}

	t.Run("paused", func(t *testing.T) {
		design := &proto.EncounterDesign{Downtimes: downtime, BossAbilities: []*proto.DesignBossAbility{designSwingTiming(0.2, 0.6, 0)}}
		sim, hits := setupDesignHitsTaken(designTankRequest(design), nil)
		pause := sim.Encounter.design.bossAbilities[0].aura

		runLifecycleUntil(sim, designAt(0.45))
		if !sim.Encounter.AllTargets[0].IsEnabled() {
			t.Fatal("test setup: expected the boss back after its downtime")
		}
		if !pause.IsActive() {
			t.Error("the downtime that clears debuffs took the pause's aura off the boss")
		}
		runLifecycleUntil(sim, sim.Duration)
		if got := bossSwingsBetween(*hits, designAt(0.2), designAt(0.6)); got != 0 {
			t.Errorf("expected no swing inside the paused window, the downtime's end included; got %d", got)
		}
		if got := bossSwingsBetween(*hits, designAt(0.6)-time.Nanosecond, designAt(0.6)+2*designBossSwingSpeed); got == 0 {
			t.Error("expected the swing to resume at the pause's end")
		}
	})

	t.Run("thrash", func(t *testing.T) {
		design := &proto.EncounterDesign{Downtimes: downtime, BossAbilities: []*proto.DesignBossAbility{designThrash(0.2, 0.6)}}
		sim, hits := setupDesignHitsTaken(designTankRequest(design), nil)
		thrash := sim.Encounter.design.bossAbilities[0].aura

		runLifecycleUntil(sim, designAt(0.45))
		if !thrash.IsActive() {
			t.Error("the downtime that clears debuffs took the thrash's aura off the boss")
		}
		runLifecycleUntil(sim, sim.Duration)
		extra := designHitsOf(*hits, func(id ActionID) bool {
			return id.OtherID == proto.OtherAction_OtherActionAttack && id.Tag == designBossAbilityTag
		})
		after := 0
		for _, hit := range extra {
			if hit.at > designAt(0.4) && hit.at < designAt(0.6) {
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
func TestDesignBossAbilitiesStartEachIterationClean(t *testing.T) {
	strike := func(start float64) *proto.DesignBossAbility {
		ability := designBossWindow(proto.DesignBossAbility_Strike, start, 1)
		ability.WeaponPercent = 150
		return ability
	}
	design := &proto.EncounterDesign{BossAbilities: []*proto.DesignBossAbility{
		strike(0.1),
		strike(0.9999),
		designSwingTiming(0.9, 1, 2),
	}}
	ownReplacements := 0
	sim, hits := setupDesignHitsTaken(designTankRequest(design), func(char *Character) {
		char.Env.Encounter.AllTargets[0].AutoAttacks.SetReplaceMHSwing(func(_ *Simulation, spell *Spell) *Spell {
			ownReplacements++
			return spell
		})
	})
	boss := sim.Encounter.AllTargets[0]
	abilities := sim.Encounter.design.bossAbilities
	lateStrike, window := abilities[1], abilities[2]
	isStrike := func(id ActionID) bool { return id.SpellID == designStrikeSpellID }

	var strikes [2]int
	for iteration := range 2 {
		*hits = (*hits)[:0]
		ownReplacements = 0
		runLifecycleUntil(sim, sim.Duration)
		strikes[iteration] = len(designHitsOf(*hits, isStrike))
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

// Repeating strikes and swing timing windows allocate nothing per iteration beyond what the design without
// them does. A spell hit's hardcast and a thrash's extra attack allocate in the cast and extra attack code
// they go through, so they stay out of the count.
func TestDesignBossAbilitiesAllocateNothingPerIteration(t *testing.T) {
	allocsPerIteration := func(abilities []*proto.DesignBossAbility) float64 {
		design := &proto.EncounterDesign{
			Downtimes:     []*proto.DesignDowntime{{TargetIndex: 0, Start: &proto.DesignTime{At: 0.95}, End: &proto.DesignTime{At: 0.97}}},
			BossAbilities: abilities,
		}
		sim, _ := setupDesignSim(designTankRequest(design), nil)
		return testing.AllocsPerRun(20, func() {
			runLifecycleUntil(sim, sim.Duration)
			sim.Cleanup()
			sim.Reset()
			sim.PrePull()
		})
	}
	strike := designBossWindow(proto.DesignBossAbility_Strike, 0.1, 1)
	strike.WeaponPercent = 150
	strike.Repeat = &proto.DesignRepeat{Interval: 0.05}
	speed := designSwingTiming(0.3, 0.35, 2)
	speed.Repeat = &proto.DesignRepeat{Interval: 0.2}
	pause := designSwingTiming(0.37, 0.4, 0)
	pause.Repeat = &proto.DesignRepeat{Interval: 0.2}

	baseline := allocsPerIteration(nil)
	if got := allocsPerIteration([]*proto.DesignBossAbility{strike, speed, pause}); got > baseline {
		t.Errorf("allocations per iteration: %.1f with the abilities, %.1f without", got, baseline)
	}
}

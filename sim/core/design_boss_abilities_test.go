package core

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
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
	outcome  HitOutcome
}

func recordDesignHitsTaken(unit *Unit, hits *[]designHitTaken) {
	MakePermanent(unit.RegisterAura(Aura{
		Label: "Hits Taken",
		OnSpellHitTaken: func(_ *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			*hits = append(*hits, designHitTaken{at: sim.CurrentTime, actionID: spell.ActionID, damage: result.Damage, outcome: result.Outcome})
		},
	}))
}

// A sim at the pull that records every hit the player takes; register runs as setupDesignSim's does.
func setupDesignHitsTaken(request *proto.RaidSimRequest, register func(char *Character)) (*Simulation, *[]designHitTaken) {
	hits := &[]designHitTaken{}
	sim, _ := setupDesignSim(request, func(char *Character) {
		recordDesignHitsTaken(&char.Unit, hits)
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
	return id.OtherID == proto.OtherAction_OtherActionAttack && id.Tag == int32(proto.ItemSlot_ItemSlotMainHand)
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

// A pause ending inside another ability's pause leaves the swing stopped until the last one ends.
func TestDesignOverlappingPausesHoldTheSwing(t *testing.T) {
	design := &proto.EncounterDesign{BossAbilities: []*proto.DesignBossAbility{designSwingTiming(0.2, 0.6, 0), designSwingTiming(0.3, 0.4, 0)}}
	_, hits := runDesignHitsTaken(t, designTankRequest(design))
	if got := bossSwingsBetween(hits, designAt(0.2), designAt(0.6)); got != 0 {
		t.Errorf("expected no swing while either pause runs, got %d", got)
	}
	if got := bossSwingsBetween(hits, designAt(0.6)-time.Nanosecond, designAt(0.6)+2*designBossSwingSpeed); got == 0 {
		t.Error("expected the swing to resume at the last pause's end")
	}
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

// Repeating strikes, swing timing windows, pulses and area hits allocate nothing per iteration beyond what the
// design without them does. A spell hit's hardcast and a thrash's extra attack allocate in the cast and extra
// attack code they go through, so they stay out of the count. The area damage is counted on a player who does not
// tank, since every hit a tank takes goes into a list that grows from empty each iteration, and deals 1, since a
// death allocates.
func TestDesignBossAbilitiesAllocateNothingPerIteration(t *testing.T) {
	allocsPerIteration := func(request func(*proto.EncounterDesign) *proto.RaidSimRequest, abilities []*proto.DesignBossAbility) float64 {
		design := &proto.EncounterDesign{
			Downtimes:     []*proto.DesignDowntime{{TargetIndex: 0, Start: &proto.DesignTime{At: 0.95}, End: &proto.DesignTime{At: 0.97}}},
			BossAbilities: abilities,
		}
		sim, _ := setupDesignSim(request(design), nil)
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

	baseline := allocsPerIteration(designTankRequest, nil)
	if got := allocsPerIteration(designTankRequest, []*proto.DesignBossAbility{strike, speed, pause}); got > baseline {
		t.Errorf("allocations per iteration: %.1f with the melee abilities, %.1f without", got, baseline)
	}

	pulse := designPulse(0.5, 0.6, 1)
	pulse.TickJitter = 0.2
	pulse.Repeat = &proto.DesignRepeat{Interval: 0.2}
	areaHit := designAreaHit(0.45)
	areaHit.Repeat = &proto.DesignRepeat{Interval: 0.1}
	for _, area := range []*proto.DesignBossAbility{pulse, areaHit} {
		area.MinDamage = 1
		area.MaxDamage = 1
	}
	untanked := func(design *proto.EncounterDesign) *proto.RaidSimRequest { return designTestRequest(1, design) }
	baseline = allocsPerIteration(untanked, nil)
	if got := allocsPerIteration(untanked, []*proto.DesignBossAbility{pulse, areaHit}); got > baseline {
		t.Errorf("allocations per iteration: %.1f with the area damage, %.1f without", got, baseline)
	}
}

func init() {
	RegisterAgentFactory(
		proto.Player_RestorationShaman{},
		proto.Spec_SpecRestorationShaman,
		newDesignPetOwner,
		func(player *proto.Player, spec interface{}) {
			player.Spec = spec.(*proto.Player_RestorationShaman)
		},
	)
}

// A player with one pet out from the start, which stands still.
type designPetOwner struct {
	Character
	pet *designPet
}

type designPet struct {
	Pet
}

func newDesignPetOwner(char *Character, _ *proto.Player, _ *proto.Raid) Agent {
	owner := &designPetOwner{Character: *char}
	owner.pet = &designPet{Pet: NewPet(PetConfig{
		Name:            "Design Pet",
		Owner:           &owner.Character,
		StatInheritance: func(stats.Stats) stats.Stats { return stats.Stats{} },
		EnabledOnStart:  true,
	})}
	owner.AddPet(owner.pet)
	return owner
}

func (owner *designPetOwner) GetCharacter() *Character       { return &owner.Character }
func (owner *designPetOwner) ApplyTalents()                  {}
func (owner *designPetOwner) Initialize()                    {}
func (owner *designPetOwner) Reset(_ *Simulation)            {}
func (owner *designPetOwner) OnGCDReady(_ *Simulation)       {}
func (owner *designPetOwner) OnEncounterStart(_ *Simulation) {}

func (pet *designPet) GetPet() *Pet                        { return &pet.Pet }
func (pet *designPet) Reset(_ *Simulation)                 {}
func (pet *designPet) OnEncounterStart(_ *Simulation)      {}
func (pet *designPet) ExecuteCustomRotation(_ *Simulation) {}

// A sim of any agent, reset and through the prepull; register runs as setupDesignSim's does.
func setupDesignAgentSim(request *proto.RaidSimRequest, register func(char *Character)) (*Simulation, *Character) {
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

func designPulse(start, end, interval float64) *proto.DesignBossAbility {
	ability := designBossWindow(proto.DesignBossAbility_Pulse, start, end)
	ability.School = proto.SpellSchool_SpellSchoolFire
	ability.MinDamage = 500
	ability.MaxDamage = 700
	ability.TickInterval = interval
	return ability
}

// An area hit whose repeats run until the fight ends.
func designAreaHit(start float64) *proto.DesignBossAbility {
	ability := designBossWindow(proto.DesignBossAbility_AreaHit, start, 1)
	ability.School = proto.SpellSchool_SpellSchoolFire
	ability.MinDamage = 500
	ability.MaxDamage = 700
	return ability
}

func designAreaHitsOf(hits []designHitTaken, index int) []designHitTaken {
	tag := designBossAbilityTag + int32(index)
	return designHitsOf(hits, func(id ActionID) bool { return id.SpellID == designSpellHitSpellID && id.Tag == tag })
}

func expectDesignHitTimes(t *testing.T, name string, hits []designHitTaken, want []time.Duration) {
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

func designTicks(start time.Duration, interval time.Duration, count int) []time.Duration {
	ticks := make([]time.Duration, count)
	for k := range ticks {
		ticks[k] = start + time.Duration(k)*interval
	}
	return ticks
}

// The first tick lands at the window's start and the last one before its end, each repeat opening a window of
// the same length; a window shorter than one interval ticks once.
func TestDesignPulseTicksFromItsWindowStartUntilItsEnd(t *testing.T) {
	repeating := designPulse(0.1, 0.2, 3)
	repeating.Repeat = &proto.DesignRepeat{Interval: 0.5, Count: 1}
	design := &proto.EncounterDesign{BossAbilities: []*proto.DesignBossAbility{repeating, designPulse(0.3, 0.301, 2)}}
	_, hits := runDesignHitsTaken(t, designTestRequest(1, design))

	ticks := designAreaHitsOf(hits, 0)
	expectDesignHitTimes(t, "repeating pulse", ticks,
		append(designTicks(designAt(0.1), 3*time.Second, 4), designTicks(designAt(0.6), 3*time.Second, 4)...))
	for i, tick := range ticks {
		if tick.damage < 500 || tick.damage > 700 {
			t.Errorf("tick %d: damage %.1f outside [500, 700]", i+1, tick.damage)
		}
	}
	expectDesignHitTimes(t, "short pulse", designAreaHitsOf(hits, 1), []time.Duration{designAt(0.3)})
}

// Area damage comes from no target: it lands while the boss is in a downtime, and an area hit's end bounds its
// repeats as a spell hit's does.
func TestDesignAreaDamageLandsWhileTheBossIsDown(t *testing.T) {
	areaHit := designAreaHit(0.2)
	areaHit.End = &proto.DesignTime{At: 0.45}
	areaHit.Repeat = &proto.DesignRepeat{Interval: 0.1}
	design := &proto.EncounterDesign{
		Downtimes:     []*proto.DesignDowntime{{TargetIndex: 0, Start: &proto.DesignTime{At: 0.05}, End: &proto.DesignTime{At: 0.5}}},
		BossAbilities: []*proto.DesignBossAbility{designPulse(0.1, 0.15, 1), areaHit},
	}
	sim, hits := setupDesignHitsTaken(designTestRequest(1, design), nil)
	runLifecycleUntil(sim, designAt(0.1))
	if sim.Encounter.AllTargets[0].IsEnabled() || len(sim.Encounter.ActiveTargets) != 0 {
		t.Fatal("test setup: expected no active target at the pulse's start")
	}
	runLifecycleUntil(sim, sim.Duration)

	expectDesignHitTimes(t, "pulse", designAreaHitsOf(*hits, 0), designTicks(designAt(0.1), time.Second, 5))
	expectDesignHitTimes(t, "area hit", designAreaHitsOf(*hits, 1), []time.Duration{designAt(0.2), designAt(0.3), designAt(0.4)})
}

// A tick is damage taken from an enemy spell: a warrior gains rage from it with no target to swing at, and it
// counts in the damage taken metrics.
func TestDesignPulseTickGrantsRageAndCountsAsDamageTaken(t *testing.T) {
	request := designTestRequest(1, &proto.EncounterDesign{
		Downtimes:     []*proto.DesignDowntime{{TargetIndex: 0, Start: &proto.DesignTime{At: 0.05}, End: &proto.DesignTime{At: 0.5}}},
		BossAbilities: []*proto.DesignBossAbility{designPulse(0.1, 0.2, 2)},
	})
	request.Raid.Parties[0].Players[0].Spec = &proto.Player_DpsWarrior{}
	sim, char := setupDesignAgentSim(request, nil)

	runLifecycleUntil(sim, designAt(0.1)-time.Millisecond)
	if len(sim.Encounter.ActiveTargets) != 0 {
		t.Fatal("test setup: expected no active target before the first tick")
	}
	before := char.CurrentRage()
	if before >= char.MaximumRage() {
		t.Fatalf("test setup: expected room for rage before the first tick, got %.2f", before)
	}
	runLifecycleUntil(sim, designAt(0.1))
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
func TestDesignAreaDamageResistsOnlyWhenResistible(t *testing.T) {
	areaHit := func(resistible bool) *proto.DesignBossAbility {
		ability := designAreaHit(0.01)
		ability.Repeat = &proto.DesignRepeat{Interval: 0.01}
		ability.MinDamage = 1000
		ability.MaxDamage = 1000
		ability.Resistible = resistible
		return ability
	}
	design := &proto.EncounterDesign{BossAbilities: []*proto.DesignBossAbility{areaHit(true), areaHit(false)}}
	sim, hits := setupDesignHitsTaken(designTestRequest(1, design), func(char *Character) {
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

	resistible, plain := designAreaHitsOf(*hits, 0), designAreaHitsOf(*hits, 1)
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

// An avoidable tick skips the player from a design move's start to its end and still hits the pet; one that
// is not avoidable hits both throughout.
func TestDesignAvoidableAreaDamageSkipsAWalkingPlayerOnly(t *testing.T) {
	avoidable := designPulse(0.15, 0.3, 1)
	avoidable.Avoidable = true
	avoidableHit := designAreaHit(0.2)
	avoidableHit.Avoidable = true
	// The move's walk at 0.2 is its repeat, queued once the walk at 0.1 runs, after the area hit at 0.2 was.
	design := &proto.EncounterDesign{
		Moves:         []*proto.DesignMove{{At: &proto.DesignTime{At: 0.1}, Seconds: 5, Repeat: &proto.DesignRepeat{Interval: 0.1, Count: 1}}},
		BossAbilities: []*proto.DesignBossAbility{avoidable, designPulse(0.15, 0.3, 1), avoidableHit},
	}
	request := designTestRequest(1, design)
	request.Raid.Parties[0].Players[0].Class = proto.Class_ClassShaman
	request.Raid.Parties[0].Players[0].Spec = &proto.Player_RestorationShaman{}
	var playerHits, petHits []designHitTaken
	sim, char := setupDesignAgentSim(request, func(char *Character) {
		recordDesignHitsTaken(&char.Unit, &playerHits)
		recordDesignHitsTaken(&char.Pets[0].Unit, &petHits)
	})
	if !char.Pets[0].IsEnabled() {
		t.Fatal("test setup: expected the pet out")
	}
	runLifecycleUntil(sim, sim.Duration)

	all := designTicks(designAt(0.15), time.Second, 15)
	expectDesignHitTimes(t, "avoidable, player", designAreaHitsOf(playerHits, 0), append(slices.Clone(all[:5]), all[10:]...))
	expectDesignHitTimes(t, "avoidable, pet", designAreaHitsOf(petHits, 0), all)
	expectDesignHitTimes(t, "not avoidable, player", designAreaHitsOf(playerHits, 1), all)
	expectDesignHitTimes(t, "not avoidable, pet", designAreaHitsOf(petHits, 1), all)
	expectDesignHitTimes(t, "avoidable area hit at the move's start, player", designAreaHitsOf(playerHits, 2), nil)
	expectDesignHitTimes(t, "avoidable area hit at the move's start, pet", designAreaHitsOf(petHits, 2), []time.Duration{designAt(0.2)})
}

// Tick jitter is rolled from the design rand: each tick after the first lands within the jitter of its place
// on the interval, the same for a seed and for other gear, and moves between iterations.
func TestDesignPulseTickJitterIsDeterministicAndIndependentOfGear(t *testing.T) {
	const jitter = 500 * time.Millisecond
	pulse := designPulse(0.1, 0.3, 2)
	pulse.TickJitter = jitter.Seconds()
	design := &proto.EncounterDesign{BossAbilities: []*proto.DesignBossAbility{pulse}}

	run := func(bonus stats.Stats) ([][]time.Duration, float64) {
		request := designTankRequest(design)
		request.Raid.Parties[0].Players[0].BonusStats = &proto.UnitStats{Stats: bonus.ToProtoArray()}
		sim, hits := setupDesignHitsTaken(request, nil)
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
			for _, hit := range designAreaHitsOf(*hits, 0) {
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

	nominal := designTicks(designAt(0.1), 2*time.Second, 10)
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

// Area damage needs damage and a pulse an interval; the target index plays no part in either kind.
func TestDesignAreaDamageSkipsWithoutDamageOrInterval(t *testing.T) {
	noInterval := designPulse(0.1, 0.2, 0)
	noDamage := designAreaHit(0.1)
	noDamage.MaxDamage = 0
	anyTarget := designAreaHit(0.1)
	anyTarget.TargetIndex = 3
	design := &proto.EncounterDesign{BossAbilities: []*proto.DesignBossAbility{noInterval, noDamage, anyTarget}}
	sim, hits := setupDesignHitsTaken(designTestRequest(1, design), nil)
	runLifecycleUntil(sim, sim.Duration)

	abilities := sim.Encounter.design.bossAbilities
	if !abilities[0].skip || !abilities[1].skip || abilities[2].skip {
		t.Errorf("skipped: expected true, true, false; got %t, %t, %t", abilities[0].skip, abilities[1].skip, abilities[2].skip)
	}
	notes := strings.Join(sim.Encounter.design.constructNotes, " ")
	for _, want := range []string{"boss ability 1 ticks at no interval", "boss ability 2 deals no damage"} {
		if !strings.Contains(notes, want) {
			t.Errorf("construct notes %q: expected %q", notes, want)
		}
	}
	expectDesignHitTimes(t, "area hit naming target 3", designAreaHitsOf(*hits, 2), []time.Duration{designAt(0.1)})
}

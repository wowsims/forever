//go:build with_db

package feralbear

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/druid"
)

func newBear(t *testing.T) (*core.Simulation, *GuardianDruid) {
	return newBearWithTalents(t, "0000000000000000-00000000000000000000-0000000000000000")
}

func newBearWithTalents(t *testing.T, talents string) (*core.Simulation, *GuardianDruid) {
	t.Helper()
	return newBearAgainst(t, talents, 1)
}

func newBearAgainst(t *testing.T, talents string, numTargets int) (*core.Simulation, *GuardianDruid) {
	t.Helper()
	targets := make([]*proto.Target, numTargets)
	for i := range targets {
		targets[i] = &proto.Target{Name: "target", Level: 60}
	}
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 866
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Bear", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
				TalentsString: talents,
				Spec: &proto.Player_FeralBearDruid{FeralBearDruid: &proto.FeralBearDruid{
					Options: &proto.FeralBearDruid_Options{StartingRage: 100},
				}},
				Equipment:   &proto.EquipmentSpec{Items: items},
				Consumables: &proto.ConsumesSpec{},
				Buffs:       &proto.IndividualBuffs{},
				Rotation:    &proto.APLRotation{},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: targets, Duration: 60},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*GuardianDruid)
}

func TestDemoralizingRoarReachesEveryTarget(t *testing.T) {
	sim, bear := newBearAgainst(t, "0000000000000000-00000000000000000000-0000000000000000", 3)
	targets := sim.Encounter.ActiveTargetUnits
	for range 20 {
		bear.AddRage(sim, 50, bear.NewRageMetrics(core.ActionID{SpellID: 1}))
		bear.DemoralizingRoar.Cast(sim, bear.CurrentTarget)
		sim.CurrentTime += 2 * time.Second
		debuffed := 0
		for _, target := range targets {
			if bear.DemoralizingRoarAuras.Get(target).IsActive() {
				debuffed++
			}
		}
		if debuffed == len(targets) {
			return
		}
	}
	t.Error("demoralizing roar never debuffed every target")
}

func TestDemoralizingRoarIgnoresDodge(t *testing.T) {
	sim, bear := newBear(t)
	target := sim.Encounter.ActiveTargetUnits[0]
	target.PseudoStats.BaseDodgeChance = 1
	for range 20 {
		bear.AddRage(sim, 50, bear.NewRageMetrics(core.ActionID{SpellID: 1}))
		bear.DemoralizingRoar.Cast(sim, bear.CurrentTarget)
		sim.CurrentTime += 2 * time.Second
		if m := bear.DemoralizingRoar.SpellMetrics[target.UnitIndex]; m.Dodges != 0 {
			t.Fatalf("demoralizing roar was dodged %d times", m.Dodges)
		}
		if bear.DemoralizingRoarAuras.Get(target).IsActive() {
			return
		}
	}
	t.Error("demoralizing roar never landed against a target that dodges every melee attack")
}

func TestBerserkPrimalBiteStrikesThreeTargets(t *testing.T) {
	sim, bear := newBearAgainst(t, "0000000000000000-00000000000100000001-0000000000000000", 5)
	metrics := bear.NewRageMetrics(core.ActionID{SpellID: 1})
	bear.BerserkAura.Activate(sim)
	bear.AddRage(sim, 100, metrics)
	if !bear.PrimalBite.Cast(sim, bear.CurrentTarget) {
		t.Fatal("primal bite did not cast under berserk")
	}
	struck := 0
	for _, target := range sim.Encounter.ActiveTargetUnits {
		if m := bear.PrimalBite.SpellMetrics[target.UnitIndex]; m.Hits+m.Crits+m.Misses+m.Dodges+m.Parries+m.Glances+m.Blocks > 0 {
			struck++
		}
	}
	if struck != 3 {
		t.Errorf("primal bite under berserk struck %d targets, want 3", struck)
	}
}

// Bash 8983 states a melee defense type and a miss refund: a dodged Bash stuns nothing and gives rage back.
func TestDodgedBashStunsNothingAndRefunds(t *testing.T) {
	sim, bear := newBear(t)
	target := sim.Encounter.ActiveTargetUnits[0]
	target.PseudoStats.BaseDodgeChance = 1
	before := bear.CurrentRage()
	if !bear.Bash.Cast(sim, bear.CurrentTarget) {
		t.Fatal("bash did not cast")
	}
	if m := bear.Bash.SpellMetrics[target.UnitIndex]; m.Dodges != 1 {
		t.Fatalf("bash dodges = %d, want 1", m.Dodges)
	}
	for _, aura := range bear.Bash.RelatedAuraArrays {
		if aura.Get(target).IsActive() {
			t.Error("a dodged bash stunned the target")
		}
	}
	if spent := before - bear.CurrentRage(); spent >= bear.Bash.DefaultCast.Cost {
		t.Errorf("a dodged bash spent %v of its %v rage", spent, bear.Bash.DefaultCast.Cost)
	}
}

// Berserk effect 2 is -100% on Primal Bite's 6 s cooldown, while Berserk is up.
func TestBerserkLiftsPrimalBiteCooldownOnlyWhileUp(t *testing.T) {
	sim, bear := newBearWithTalents(t, "0000000000000000-00000000000100000001-0000000000000000")
	metrics := bear.NewRageMetrics(core.ActionID{SpellID: 1})
	cast := func(when string) {
		t.Helper()
		bear.AddRage(sim, 100, metrics)
		if !bear.PrimalBite.Cast(sim, bear.CurrentTarget) {
			t.Fatalf("primal bite did not cast %s", when)
		}
		sim.CurrentTime += 2 * time.Second
	}

	bear.BerserkAura.Activate(sim)
	cast("under berserk")
	if !bear.PrimalBite.IsReady(sim) {
		t.Errorf("primal bite is on cooldown until %v under berserk", bear.PrimalBite.ReadyAt())
	}
	bear.BerserkAura.Deactivate(sim)
	cast("after berserk")
	if got, want := bear.PrimalBite.ReadyAt(), sim.CurrentTime+4*time.Second; got != want {
		t.Errorf("primal bite is ready at %v after berserk, want %v", got, want)
	}
}

// Berserk's tooltip "removes its cooldown": a Primal Bite cooldown already running when Berserk starts
// ends with it.
func TestBerserkClearsARunningPrimalBiteCooldown(t *testing.T) {
	sim, bear := newBearWithTalents(t, "0000000000000000-00000000000100000001-0000000000000000")
	bear.AddRage(sim, 100, bear.NewRageMetrics(core.ActionID{SpellID: 1}))
	if !bear.PrimalBite.Cast(sim, bear.CurrentTarget) {
		t.Fatal("primal bite did not cast")
	}
	sim.CurrentTime += 2 * time.Second
	if bear.PrimalBite.IsReady(sim) {
		t.Fatal("primal bite is ready 2 s after a cast without berserk")
	}
	bear.BerserkAura.Activate(sim)
	if !bear.PrimalBite.IsReady(sim) {
		t.Errorf("primal bite is on cooldown until %v once berserk starts", bear.PrimalBite.ReadyAt())
	}
}

func TestFrenziedRegenerationSpendsRageForHealth(t *testing.T) {
	sim, bear := newBear(t)
	bear.AddRage(sim, 100, bear.NewRageMetrics(core.ActionID{SpellID: 1}))
	before := bear.CurrentRage()
	if !bear.FrenziedRegeneration.Cast(sim, bear.CurrentTarget) {
		t.Fatal("frenzied regeneration did not cast")
	}
	for sim.CurrentTime < 1100*time.Millisecond {
		if sim.Step() {
			break
		}
	}
	if spent := before - bear.CurrentRage(); spent < 10 {
		t.Errorf("frenzied regeneration spent %v rage in its first tick, want 10", spent)
	}
	healed := 0.0
	for _, m := range bear.FrenziedRegeneration.SpellMetrics {
		healed += m.TotalHealing
	}
	if healed <= 0 {
		t.Error("frenzied regeneration healed nothing")
	}
}

func TestFurorPreservesEnergyWhenStartingInABearForm(t *testing.T) {
	sim, bear := newBearWithTalents(t, "0000000000000000-00000000000000000000-0500000000000000")
	if !bear.InForm(druid.DireBear) {
		t.Fatal("the bear did not start in dire bear form")
	}
	sim.CurrentTime += 3 * time.Second
	bear.DireBearFormAura.Deactivate(sim)
	sim.CurrentTime += 2 * time.Second
	bear.AddEnergy(sim, bear.MaximumEnergy(), bear.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	if !bear.CatForm.Cast(sim, bear.CurrentTarget) {
		t.Fatal("cat form did not cast")
	}
	// The pull starts at full energy, which remains capped after the caster interlude.
	if got := bear.CurrentEnergy(); got != 100 {
		t.Errorf("energy in cat form after starting in dire bear form = %v, want 100", got)
	}
}

func TestFirstLacerateHitsForNothing(t *testing.T) {
	sim, bear := newBear(t)
	target := bear.CurrentTarget
	metrics := bear.NewRageMetrics(core.ActionID{SpellID: 1})
	for range 20 {
		bear.AddRage(sim, 50, metrics)
		bear.Lacerate.Cast(sim, target)
		if bear.Lacerate.Dot(target).IsActive() {
			break
		}
		sim.CurrentTime += 2 * time.Second
	}
	if !bear.Lacerate.Dot(target).IsActive() {
		t.Fatal("lacerate never landed")
	}
	if got := bear.Lacerate.SpellMetrics[target.UnitIndex].TotalDamage; got != 0 {
		t.Errorf("the first lacerate dealt %v, want 0 with no stack already on the target", got)
	}
}

func TestLacerateStacksToFive(t *testing.T) {
	sim, bear := newBear(t)
	if !bear.InForm(druid.Bear) {
		t.Fatal("the bear did not start in bear form")
	}
	target := bear.CurrentTarget
	metrics := bear.NewRageMetrics(core.ActionID{SpellID: 1})
	for range 40 {
		bear.AddRage(sim, 50, metrics)
		bear.Lacerate.Cast(sim, target)
		sim.CurrentTime += 2 * time.Second
		if bear.Lacerate.Dot(target).GetStacks() >= 5 {
			break
		}
	}
	if got := bear.Lacerate.Dot(target).GetStacks(); got != 5 {
		t.Errorf("lacerate stacks = %d, want 5", got)
	}
}

func TestMaulReplacesTheNextSwing(t *testing.T) {
	sim, bear := newBear(t)
	if bear.TryMaul(sim, nil) != nil {
		t.Fatal("maul replaced a swing before it was queued")
	}
	if !bear.Maul.Cast(sim, bear.CurrentTarget) {
		t.Fatal("maul did not queue")
	}
	for sim.CurrentTime < 100*time.Millisecond {
		if sim.Step() {
			break
		}
	}
	replaced := bear.TryMaul(sim, nil)
	if replaced == nil || replaced.ActionID.SpellID != 9881 {
		t.Fatalf("the queued swing was %v, want maul 9881", replaced)
	}
}

func TestFeralChargeCastsOutsideMelee(t *testing.T) {
	sim, bear := newBearWithTalents(t, "0000000000000000-00000000100000000000-0000000000000000")
	bear.DistanceFromTarget = 10
	if !bear.FeralCharge.Cast(sim, bear.CurrentTarget) {
		t.Fatal("feral charge did not cast at 10 yards")
	}
	if bear.FeralCharge.CD.Timer != nil {
		bear.FeralCharge.CD.Reset()
	}
	bear.DistanceFromTarget = 0
	if bear.FeralCharge.CanCast(sim, bear.CurrentTarget) {
		t.Error("feral charge could be cast in melee")
	}
}

func TestSwipeDoesNotScaleWithAttackPower(t *testing.T) {
	sim, bear := newBear(t)
	target := bear.CurrentTarget
	metrics := bear.NewRageMetrics(core.ActionID{SpellID: 1})
	for range 20 {
		bear.AddRage(sim, 50, metrics)
		bear.Swipe.Cast(sim, target)
		sim.CurrentTime += 2 * time.Second
		if bear.Swipe.SpellMetrics[target.UnitIndex].Hits > 0 {
			break
		}
	}
	hits := bear.Swipe.SpellMetrics[target.UnitIndex]
	if hits.Hits == 0 {
		t.Fatal("swipe never hit")
	}
	perHit := hits.TotalDamage / float64(hits.Hits)
	// The row's damage is 83. Attack power on a bear is far larger than the gap this allows.
	if perHit < 40 || perHit > 160 {
		t.Errorf("swipe hit for %.1f, want the row's damage without attack power", perHit)
	}
}

// A bear's swing takes the two-hander rage factor: 4.5 rage per second of its 2.5 second swing is 11.25,
// and a crit adds the swing's one-hand rage, 3.46 * 2.5 = 8.65, for 19.9.
func TestBearSwingRageTakesTheTwoHandFactor(t *testing.T) {
	for _, c := range []struct {
		outcome core.HitOutcome
		want    float64
	}{{core.OutcomeHit, 11.25}, {core.OutcomeCrit, 19.9}} {
		sim, bear := newBear(t)
		bear.SpendRage(sim, bear.CurrentRage(), bear.NewRageMetrics(core.ActionID{SpellID: 1}))
		result := &core.SpellResult{Target: bear.CurrentTarget, Outcome: c.outcome, Damage: 100}
		before := bear.CurrentRage()
		bear.Unit.OnSpellHitDealt(sim, bear.AutoAttacks.MHAuto(), result)
		if got := bear.CurrentRage() - before; math.Abs(got-c.want) > 1e-9 {
			t.Errorf("%v swing grants %g rage, want %g", c.outcome, got, c.want)
		}
	}
}

// Bear Form (Passive2) 21178 raises threat by 50% while a bear form is up.
func TestBearFormsRaiseThreatByHalf(t *testing.T) {
	sim, bear := newBear(t)
	if got := bear.PseudoStats.ThreatMultiplier; math.Abs(got-1.5) > 1e-9 {
		t.Errorf("threat multiplier in Dire Bear Form is %g, want 1.5", got)
	}
	bear.DireBearFormAura.Deactivate(sim)
	if got := bear.PseudoStats.ThreatMultiplier; math.Abs(got-1) > 1e-9 {
		t.Errorf("threat multiplier out of form is %g, want 1", got)
	}
	bear.BearFormAura.Activate(sim)
	if got := bear.PseudoStats.ThreatMultiplier; math.Abs(got-1.5) > 1e-9 {
		t.Errorf("threat multiplier in Bear Form is %g, want 1.5", got)
	}
}

// Natural Reaction 5/5 (417051): every dodge in a bear form grants 417053's 5 rage.
func TestNaturalReactionGivesRageOnDodges(t *testing.T) {
	req := bearRequest(bearAPL(bestBearAPL), 20)
	result := core.RunSim(req, nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}
	for _, r := range result.RaidMetrics.Parties[0].Players[0].Resources {
		if r.Id.GetSpellId() == 417053 {
			if r.Events == 0 || math.Abs(r.Gain/float64(r.Events)-5) > 1e-9 {
				t.Fatalf("Natural Reaction gave %g rage over %d events, want 5 each", r.Gain, r.Events)
			}
			return
		}
	}
	t.Fatal("Natural Reaction never gave rage")
}

// Swipe 9908 adds 3% of attack power to each hit, an override on the row.
func TestSwipeGainsThreePercentOfAttackPower(t *testing.T) {
	sim, bear := newBear(t)
	bear.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
	bear.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
	bear.CurrentTarget.PseudoStats.BaseDodgeChance = 0
	hit := func() float64 {
		bear.AddRage(sim, 100, bear.NewRageMetrics(core.ActionID{SpellID: 1}))
		before := bear.Swipe.SpellMetrics[bear.CurrentTarget.UnitIndex].TotalDamage
		if !bear.Swipe.Cast(sim, bear.CurrentTarget) {
			t.Fatal("swipe did not cast")
		}
		sim.CurrentTime += 2 * time.Second
		return bear.Swipe.SpellMetrics[bear.CurrentTarget.UnitIndex].TotalDamage - before
	}
	low := hit()
	bear.AddStatDynamic(sim, stats.AttackPower, 1000)
	high := hit()
	armor, _ := bear.Swipe.ResistanceMultiplier(sim, false, bear.AttackTables[bear.CurrentTarget.UnitIndex])
	if got, want := high-low, 30*armor; math.Abs(got-want) > 1e-6 {
		t.Errorf("1000 attack power adds %g to a Swipe, want %g", got, want)
	}
}

// The bear forms scale the items' bonus armor by their 180% or 360% (A_MOD_BONUS_ARMOR_PCT) as well as
// their base armor, and Thick Hide adds the level times 3 plus $m1/100 = 2 base armor per point of bonus
// defense, which the form scales too. The UI's bear has Thick Hide 3/3 and gear with bonus armor and
// defense.
func TestBearFormsScaleBonusArmorAndThickHideDefense(t *testing.T) {
	sim := core.NewSim(bearRequest(bearAPL(bestBearAPL), 1), simsignals.CreateSignals())
	sim.Reset()
	bear := sim.Raid.Parties[0].Players[0].(*GuardianDruid)
	inBear := bear.GetStat(stats.Armor)
	bear.DireBearFormAura.Deactivate(sim)
	caster := bear.GetStat(stats.Armor)
	equip := bear.EquipStats()
	if equip[stats.BonusArmor] == 0 || bear.GetStat(stats.DefenseRating) == 0 {
		t.Fatal("the bear gear has no bonus armor or defense")
	}
	thickHide := 60*3 + 2*bear.GetStat(stats.DefenseRating)
	want := 3.6*(equip[stats.Armor]+equip[stats.BonusArmor]) + 4.6*thickHide
	if got := inBear - caster; math.Abs(got-want) > 1e-6 {
		t.Errorf("Dire Bear adds %g armor, want %g", got, want)
	}
}

// Thick Hide's level part is the level times the rank: a level 30 druid measured 30, 60 and 90 armor from
// ranks 1 to 3 in Cat Form and 84, 168 and 252 in Bear Form. The game files' per-level formula (178 and 179
// at level 60) is not what the game does. Dire Bear Form scales it by 360%, 4.6 times in all.
func TestThickHideArmorIsLevelTimesRank(t *testing.T) {
	armorAt := func(rank int) float64 {
		talents := "0000000000000000-00000" + string(rune('0'+rank)) + "00000000000000-0000000000000000"
		_, bear := newBearWithTalents(t, talents)
		return bear.GetStat(stats.Armor)
	}
	base := armorAt(0)
	for rank := 1; rank <= 3; rank++ {
		want := 4.6 * 60 * float64(rank)
		if got := armorAt(rank) - base; math.Abs(got-want) > 1e-6 {
			t.Errorf("Thick Hide %d/3 adds %g armor in Dire Bear Form, want %g", rank, got, want)
		}
	}
}

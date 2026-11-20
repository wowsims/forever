package rogue

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common"
	_ "github.com/wowsims/forever/sim/common" // imported to get item effects included.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func init() {
	RegisterRogue()
	common.RegisterAllEffects()
}

// Each book rank registers beside the trainer's top rank. Backstab builds a combo point at either rank, Eviscerate
// spends them all at either rank and hits harder at the book rank, and both ranks of Feint run one category cooldown.
func TestBookRanksBesideTrainerRanks(t *testing.T) {
	newRogue := func(rank *spelldata.Spell) (*core.Simulation, *Rogue, *core.Spell) {
		player := core.WithSpec(&proto.Player{
			Class:     proto.Class_ClassRogue,
			Race:      proto.Race_RaceHuman,
			Equipment: core.GetGearSet("../../ui/specs/rogue/dps/gear_sets", "muti").GearSet,
			Rotation:  &proto.APLRotation{},
		}, DefaultOptions)
		sim := core.NewSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  core.MakeSingleTargetEncounter(0),
			SimOptions: &proto.SimOptions{RandomSeed: 1},
		}, simsignals.CreateSignals())
		sim.Reset()
		rogue := sim.Raid.Parties[0].Players[0].(*Rogue)
		rogue.CurrentTarget.PseudoStats.BaseDodgeChance = 0
		rogue.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
		rogue.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
		spell := rogue.GetSpell(core.ActionID{SpellID: rank.ID})
		if spell == nil {
			t.Fatalf("%s is not registered", rank.RankedName())
		}
		return sim, rogue, spell
	}

	for _, rank := range []*spelldata.Spell{backstabTrainerRank, backstabBookRank} {
		sim, rogue, backstab := newRogue(rank)
		if !backstab.Cast(sim, rogue.CurrentTarget) {
			t.Fatalf("%s did not cast", rank.RankedName())
		}
		if rogue.ComboPoints(rogue.CurrentTarget) != 1 || !backstab.Matches(RogueSpellBackstab) {
			t.Errorf("%s leaves %d combo points, or lost the Backstab mask; want 1", rank.RankedName(), rogue.ComboPoints(rogue.CurrentTarget))
		}
	}

	eviscerate := func(rank *spelldata.Spell) float64 {
		sim, rogue, spell := newRogue(rank)
		target := rogue.CurrentTarget
		rogue.AddComboPoints(sim, target, 5, rogue.NewComboPointMetrics(core.ActionID{SpellID: 1}))
		if !spell.Cast(sim, target) {
			t.Fatalf("%s did not cast", rank.RankedName())
		}
		if rogue.ComboPoints(target) != 0 {
			t.Errorf("%s leaves %d combo points, want 0", rank.RankedName(), rogue.ComboPoints(target))
		}
		return spell.SpellMetrics[target.UnitIndex].TotalDamage
	}
	if trainer, book := eviscerate(eviscerateTrainerRank), eviscerate(eviscerateBookRank); book <= trainer {
		t.Errorf("the book rank of Eviscerate deals %.0f at 5 combo points, want more than the trainer rank's %.0f", book, trainer)
	}

	sim, rogue, trainerFeint := newRogue(feintTrainerRank)
	bookFeint := rogue.GetSpell(core.ActionID{SpellID: feintBookRank.ID})
	if bookFeint == nil {
		t.Fatalf("%s is not registered", feintBookRank.RankedName())
	}
	if !trainerFeint.Cast(sim, rogue.CurrentTarget) {
		t.Fatal("the trainer rank of Feint did not cast")
	}
	if bookFeint.CD.IsReady(sim) {
		t.Error("the trainer rank of Feint left the book rank ready, want one category cooldown")
	}
}

func TestRogue(t *testing.T) {
	rotation := core.GetAplRotation("../../ui/specs/rogue/dps/apls", "default")
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:         proto.Class_ClassRogue,
			Race:          proto.Race_RaceHuman,
			OtherRaces:    []proto.Race{proto.Race_RaceOrc},
			GearSet:       core.GetGearSet("../../ui/specs/rogue/dps/gear_sets", "muti"),
			OtherGearSets: []core.GearSetCombo{
				//core.GetGearSet("../../ui/specs/rogue/dps/gear_sets", "p1"),
			},
			Talents: DefaultTalents,
			OtherTalentSets: []core.TalentsCombo{
				{
					Label:   "Combat SS",
					Talents: "00530320201-3200132202012321551",
				},
				{
					Label:   "Mutilate",
					Talents: "00530311231251031-32030302020103-001",
				},
			},
			Consumables: DefaultConsumables,
			SpecOptions: core.SpecOptionsCombo{Label: "Rogue", SpecOptions: DefaultOptions},

			Rotation: rotation,
			OtherRotations: []core.RotationCombo{
				core.RotationWithSpellIds("default_book_ranks", rotation, map[int32]int32{11300: 31016, 11281: 25300}),
			},
			ItemFilter: core.ItemFilter{
				ArmorType: proto.ArmorType_ArmorTypeLeather,
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeDagger,
					proto.WeaponType_WeaponTypeFist,
					proto.WeaponType_WeaponTypeMace,
					proto.WeaponType_WeaponTypeSword,
					proto.WeaponType_WeaponTypeAxe,
				},
				HandTypes: []proto.HandType{
					proto.HandType_HandTypeMainHand,
					proto.HandType_HandTypeOffHand,
					proto.HandType_HandTypeOneHand,
				},
			},
		},
	}))
}

var DefaultOptions = &proto.Player_Rogue{
	Rogue: &proto.Rogue{
		Options: &proto.Rogue_Options{
			ClassOptions: &proto.RogueOptions{},
		},
	},
}

var DefaultTalents = "00530301-320003-305230031321001151"

var DefaultConsumables = &proto.ConsumesSpec{
	FlaskId:         274274, // Flask of Natural Aggression
	FoodId:          250072, // Swiftstrike Steak
	OffensivePotId:  250943, // Major Frenzy Potion
	ConjuredId:      7676,
	MhWeaponImbueId: 8928, // Instant Poison VI
	OhWeaponImbueId: 8985, // Deadly Poison IV
}

// A reset sim of one rogue with the talents, no gear and no rotation, against two targets.
func newTwoTargetSim(talents string) (*core.Simulation, *Rogue) {
	encounter := core.MakeSingleTargetEncounter(0)
	encounter.Targets = append(encounter.Targets, core.NewDefaultTarget())
	player := &proto.Player{
		Name:          "Rogue",
		Race:          proto.Race_RaceHuman,
		Class:         proto.Class_ClassRogue,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: talents,
		Spec:          DefaultOptions,
		Rotation:      &proto.APLRotation{},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*Rogue)
}

// Runs the sim for wait and strictly past one spell batch after it, so every deferred proc due in it has
// fired whatever the order of actions at the same time.
func stepPastBatch(sim *core.Simulation, wait time.Duration) {
	at := sim.CurrentTime + wait + core.SpellBatchWindow
	for sim.CurrentTime <= at {
		sim.AddPendingAction(&core.PendingAction{NextActionAt: at + 1, OnAction: func(*core.Simulation) {}})
		sim.Step()
	}
}

// Remorseless Attacks hears the killing blow: a credited kill gives Sinister Strike the crit of the
// talent's rank, a death the raid gets no kill for does not, and the strike spends it.
func TestRemorselessAttacksCritsTheStrikeAfterAKill(t *testing.T) {
	sim, rogue := newTwoTargetSim("01")
	dying, other := sim.Encounter.AllTargets[0], sim.Encounter.AllTargetUnits[1]
	buff := rogue.GetAuraByID(core.ActionID{SpellID: remorselessRank.ID})
	strike := rogue.SinisterStrike
	base := strike.BonusCritPercent
	bonus := spellData.RemorselessAttacks.Rank(1).EffectN(1).BaseValue()

	dying.Die(sim, 0)
	stepPastBatch(sim, 0)
	if buff.IsActive() {
		t.Error("a death with no kill credit put up Remorseless")
	}

	dying.Enable(sim)
	dying.Die(sim, 1)
	stepPastBatch(sim, 0)
	if got := strike.BonusCritPercent - base; !buff.IsActive() || math.Abs(got-bonus) > 1e-9 {
		t.Fatalf("after a killing blow Remorseless is up %v and adds %v crit to Sinister Strike, want up and %v", buff.IsActive(), got, bonus)
	}

	strike.SkipCastAndApplyEffects(sim, other)
	stepPastBatch(sim, 0)
	if buff.IsActive() || strike.BonusCritPercent != base {
		t.Errorf("after Sinister Strike Remorseless is up %v at %v crit, want gone and %v", buff.IsActive(), strike.BonusCritPercent, base)
	}
}

// A miss, dodge or parry of the strike spends Remorseless as a landed one does, and the landed strike
// still rolls its crit with the bonus.
func TestRemorselessAttacksSpendsOnAvoidedStrikesAndCritsLandedOnes(t *testing.T) {
	sim, rogue := newTwoTargetSim("01")
	target := sim.Encounter.AllTargetUnits[1]
	buff := rogue.GetAuraByID(core.ActionID{SpellID: remorselessRank.ID})
	strike := rogue.SinisterStrike
	table := rogue.AttackTables[target.UnitIndex].Hand(strike)
	bonus := spellData.RemorselessAttacks.Rank(1).EffectN(1).BaseValue()
	table.BaseMissChance, table.BaseDodgeChance, table.BaseParryChance, table.BaseBlockChance = 0, 0, 0, 0

	for _, c := range []struct {
		name string
		set  func()
	}{
		{"a miss", func() { table.BaseMissChance = 1 }},
		{"a dodge", func() { table.BaseDodgeChance = 1 }},
		{"a parry", func() { table.BaseParryChance = 1 }},
	} {
		c.set()
		buff.Activate(sim)
		strike.SkipCastAndApplyEffects(sim, target)
		if buff.IsActive() || strike.BonusCritPercent != 0 {
			t.Errorf("after %s Remorseless is up %v at %v crit, want gone and 0", c.name, buff.IsActive(), strike.BonusCritPercent)
		}
		table.BaseMissChance, table.BaseDodgeChance, table.BaseParryChance = 0, 0, 0
	}

	const casts = 2000
	critRate := func(withBuff bool) float64 {
		before := strike.SpellMetrics[target.UnitIndex].Crits
		for range casts {
			buff.Deactivate(sim)
			if withBuff {
				buff.Activate(sim)
			}
			strike.SkipCastAndApplyEffects(sim, target)
		}
		return float64(strike.SpellMetrics[target.UnitIndex].Crits-before) / casts
	}
	with, without := critRate(true), critRate(false)
	if with-without < bonus/200 {
		t.Errorf("landed strikes crit %v with Remorseless and %v without, want the bonus of %v points on top", with, without, bonus)
	}
}

// Lightning Reflexes 5/5, Deflection 3/3, Riposte, Setup 3/3 and Ghostly Strike.
const tankTalents = "-005030001-300000000001"

// A rogue with the talents given, facing two targets.
func newTankSim(t *testing.T, talents string) (*core.Simulation, *Rogue, []*core.Unit) {
	t.Helper()

	player := &proto.Player{
		Name:          "Tank",
		Race:          proto.Race_RaceHuman,
		Class:         proto.Class_ClassRogue,
		Equipment:     core.GetGearSet("../../ui/specs/rogue/dps/gear_sets", "muti").GearSet,
		TalentsString: talents,
		Spec:          DefaultOptions,
		Rotation:      &proto.APLRotation{},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter: &proto.Encounter{
			Duration: core.LongDuration,
			Targets:  []*proto.Target{core.NewDefaultTarget(), core.NewDefaultTarget()},
		},
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()

	return sim, sim.Raid.Parties[0].Players[0].(RogueAgent).GetRogue(), sim.Encounter.AllTargetUnits
}

// Steps the sim until the rogue's GCD is ready.
func waitForGCD(t *testing.T, sim *core.Simulation, rogue *Rogue) {
	t.Helper()
	for !rogue.GCD.IsReady(sim) {
		if sim.Step() {
			t.Fatal("the sim ended while the GCD was running")
		}
	}
}

// A trigger without TriggerImmediately runs its handler one batch window after the hit.
func passBatchWindow(t *testing.T, sim *core.Simulation) {
	t.Helper()
	until := sim.CurrentTime + core.SpellBatchWindow
	for sim.CurrentTime <= until {
		if sim.Step() {
			t.Fatal("the sim ended inside the batch window")
		}
	}
}

func auraOf(t *testing.T, rogue *Rogue, id int32) *core.Aura {
	t.Helper()
	aura := rogue.GetAuraByID(core.ActionID{SpellID: id})
	if aura == nil {
		t.Fatalf("the rogue has no aura %d", id)
	}
	return aura
}

func swingAt(sim *core.Simulation, rogue *Rogue, from *core.Unit, outcome core.HitOutcome) {
	rogue.OnSpellHitTaken(sim, &core.Spell{ProcMask: core.ProcMaskMeleeMHAuto, Unit: from},
		&core.SpellResult{Target: &rogue.Unit, Outcome: outcome})
}

// Defensive State (16092) hears melee hits taken, and Riposte's window opens on a parry only.
func TestRiposteWindowOpensOnParriedSwings(t *testing.T) {
	for _, c := range []struct {
		name    string
		outcome core.HitOutcome
		opens   bool
	}{
		{"a parried swing opens the window", core.OutcomeParry, true},
		{"a dodged swing does not open the window", core.OutcomeDodge, false},
		{"a landed swing does not open the window", core.OutcomeHit, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			sim, rogue, targets := newTankSim(t, tankTalents)
			window := auraOf(t, rogue, riposteWindow.ID)

			swingAt(sim, rogue, targets[0], c.outcome)
			passBatchWindow(t, sim)

			if got := window.IsActive(); got != c.opens {
				t.Errorf("the window is up: %v, want %v", got, c.opens)
			}
		})
	}
}

// A finisher spends the combo points on its cast target and is refused at a unit that holds none.
func TestFinisherNeedsComboPointsOnItsTarget(t *testing.T) {
	sim, rogue, targets := newTankSim(t, "")
	holder, other := targets[0], targets[1]
	eviscerate := rogue.GetSpell(core.ActionID{SpellID: eviscerateTrainerRank.ID})
	rogue.AddComboPoints(sim, holder, 3, rogue.NewComboPointMetrics(core.ActionID{SpellID: 1}))

	if eviscerate.CanCast(sim, other) {
		t.Error("Eviscerate can be cast at a target that holds no combo points")
	}
	if !eviscerate.CanCast(sim, holder) {
		t.Fatal("Eviscerate cannot be cast at the target that holds the combo points")
	}
	holder.PseudoStats.BaseDodgeChance = 0
	rogue.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
	if !eviscerate.Cast(sim, holder) {
		t.Fatal("Eviscerate did not cast at the target that holds the combo points")
	}
	if got := rogue.ComboPoints(holder); got != 0 {
		t.Errorf("Eviscerate leaves %d combo points on its target, want 0", got)
	}
}

// A Rupture on a second target leaves the first target's ticks at the points the first cast spent.
func TestRuptureTicksKeepTheirOwnCastsPoints(t *testing.T) {
	sim, rogue, targets := newTankSim(t, "")
	first, second := targets[0], targets[1]
	rogue.AutoAttacks.CancelAutoSwing(sim)
	rogue.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
	rupture := func(target *core.Unit, points int32) {
		t.Helper()
		for !rogue.GCD.IsReady(sim) {
			if sim.Step() {
				t.Fatal("the sim ended while the GCD was running")
			}
		}
		target.PseudoStats.BaseDodgeChance = 0
		rogue.AddEnergy(sim, 100, rogue.EnergyRefundMetrics)
		rogue.AddComboPoints(sim, target, points, rogue.Rupture.ComboPointMetrics())
		if !rogue.Rupture.Cast(sim, target) {
			t.Fatalf("Rupture did not cast at %s", target.Label)
		}
	}
	tick := func() float64 {
		before := rogue.Rupture.SpellMetrics[first.UnitIndex].TotalDamage
		rogue.Rupture.Dot(first).TickOnce(sim)
		return rogue.Rupture.SpellMetrics[first.UnitIndex].TotalDamage - before
	}

	rupture(first, 5)
	atFive := tick()
	rupture(second, 1)
	if got := tick(); got != atFive {
		t.Errorf("a 5-point Rupture ticks for %.4f after a 1-point Rupture on another target, want %.4f", got, atFive)
	}
}

// Riposte is off the GCD, and only its own cast spends the window: a Sinister Strike leaves it up.
func TestRiposteCastsOffTheGCDAndSpendsItsWindow(t *testing.T) {
	sim, rogue, targets := newTankSim(t, tankTalents)
	window := auraOf(t, rogue, riposteWindow.ID)

	swingAt(sim, rogue, targets[0], core.OutcomeParry)
	passBatchWindow(t, sim)

	if !rogue.SinisterStrike.Cast(sim, targets[0]) {
		t.Fatal("Sinister Strike did not cast")
	}
	if !window.IsActive() {
		t.Fatal("Sinister Strike spent the Riposte window")
	}
	if rogue.GCD.IsReady(sim) {
		t.Fatal("Sinister Strike left the GCD ready")
	}
	if !rogue.Riposte.Cast(sim, targets[0]) {
		t.Fatal("Riposte did not cast while the GCD was up")
	}
	if window.IsActive() {
		t.Error("Riposte's cast left its window up")
	}
}

// Setup hears the current target's attacks only, and a dodge of one adds the row's point at the rank's chance.
func TestSetupAddsAPointOnTheCurrentTargetsDodgedSwings(t *testing.T) {
	const dodges = 150
	perProc := int32(setupTriggered.ProcEnergizeEffect().BaseValue())
	chance := spellData.Setup.Rank(3).EffectN(1).Percent()
	for _, c := range []struct {
		name      string
		fromIndex int
		rate      float64
		swapAway  bool
	}{
		{"a dodge of the current target's swing adds a point at the rank's chance", 0, chance, false},
		{"a dodge of another target's swing adds none", 1, 0, false},
		{"a swap before the deferred proc still adds the point on the attacker", 0, chance, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			sim, rogue, targets := newTankSim(t, tankTalents)
			if rogue.CurrentTarget != targets[0] {
				t.Fatal("the rogue's current target is not the first target")
			}
			icd := spellData.Setup.Rank(3).ICD()

			var gained int32
			for range dodges {
				rogue.CurrentTarget = targets[0]
				rogue.ResetComboPoints(sim, targets[0], 0)
				swingAt(sim, rogue, targets[c.fromIndex], core.OutcomeDodge)
				if c.swapAway {
					rogue.CurrentTarget = targets[1]
				}
				passBatchWindow(t, sim)
				gained += rogue.ComboPoints(targets[0])
				if got := rogue.ComboPoints(targets[1]); got != 0 {
					t.Fatalf("%s holds %d combo points after a dodge of %s's swing, want 0", targets[1].Label, got, targets[c.fromIndex].Label)
				}

				until := sim.CurrentTime + icd
				for sim.CurrentTime <= until {
					if sim.Step() {
						t.Fatal("the sim ended between two dodges")
					}
				}
			}

			if got := float64(gained) / float64(perProc*dodges); math.Abs(got-c.rate) > 0.12 {
				t.Errorf("a dodge adds a point %.2f of the time, want %.2f", got, c.rate)
			}
		})
	}
}

// Lightning Reflexes 5/5 and Deflection 3/3 add their rank's dodge and parry, and Ghostly Strike's buff its
// second effect's dodge.
func TestAvoidanceTalentsAddTheirRowsAmounts(t *testing.T) {
	_, bare, _ := newTankSim(t, "")
	sim, rogue, _ := newTankSim(t, tankTalents)

	wantDodge := spellData.LightningReflexes.Rank(5).EffectN(1).BaseValue()
	if got := rogue.GetStat(stats.DodgePercent) - bare.GetStat(stats.DodgePercent); got != wantDodge {
		t.Errorf("Lightning Reflexes adds %.2f%% dodge, want %.2f%%", got, wantDodge)
	}
	wantParry := spellData.Deflection.Rank(3).EffectN(1).BaseValue()
	if got := rogue.GetStat(stats.ParryPercent) - bare.GetStat(stats.ParryPercent); got != wantParry {
		t.Errorf("Deflection adds %.2f%% parry, want %.2f%%", got, wantParry)
	}

	buff := auraOf(t, rogue, ghostlyStrikeRank.ID)
	before := rogue.GetStat(stats.DodgePercent)
	buff.Activate(sim)
	if got, want := rogue.GetStat(stats.DodgePercent)-before, ghostlyStrikeRank.EffectN(2).BaseValue(); math.Abs(got-want) > 1e-9 {
		t.Errorf("Ghostly Strike's buff adds %.2f%% dodge, want %.2f%%", got, want)
	}
}

// Ghostly Strike puts its dodge buff up when it lands, and an avoided strike grants none.
func TestGhostlyStrikeBuffsOnlyWhenItLands(t *testing.T) {
	sim, rogue, targets := newTankSim(t, tankTalents)
	target := targets[0]
	buff := auraOf(t, rogue, ghostlyStrikeRank.ID)
	metrics := &rogue.GhostlyStrike.SpellMetrics[target.UnitIndex]
	avoided := func() int32 { return metrics.Misses + metrics.Dodges + metrics.Parries }

	var landedSeen, avoidedSeen bool
	for tries := 0; tries < 500 && !(landedSeen && avoidedSeen); tries++ {
		for rogue.CurrentEnergy() < rogue.GhostlyStrike.DefaultCast.Cost || !rogue.GCD.IsReady(sim) {
			if sim.Step() {
				t.Fatal("the sim ended before Ghostly Strike was castable")
			}
		}
		buff.Deactivate(sim)
		rogue.GhostlyStrike.CD.Reset()
		before := avoided()
		if !rogue.GhostlyStrike.Cast(sim, target) {
			t.Fatal("Ghostly Strike did not cast")
		}

		if avoided() > before {
			avoidedSeen = true
			if buff.IsActive() {
				t.Fatal("an avoided Ghostly Strike put its buff up")
			}
		} else {
			landedSeen = true
			if !buff.IsActive() {
				t.Fatal("a landed Ghostly Strike left its buff down")
			}
		}
	}
	if !landedSeen || !avoidedSeen {
		t.Fatalf("saw a landed strike: %v, an avoided one: %v", landedSeen, avoidedSeen)
	}
}

// Venom and Flash of Steel 5/5, under 21 Assassination and 30 Combat points.
const venomFlashOfSteelTalents = "00530321201301-323303000201232105"

// Boot Knife 5/5, under 30 Subtlety points.
const bootKnifeTalents = "--32520303130100115"

// Shank is never avoided, restores its row's energy plus Flash of Steel's, cools down for its row's time less Flash
// of Steel's, puts Venom up with that talent, and adds Boot Knife's points at 5/5, where the rank's chance is 100%.
func TestShankFollowsItsTalents(t *testing.T) {
	for _, c := range []struct {
		name         string
		talents      string
		flashOfSteel int32
		venom        bool
		bootKnife    bool
	}{
		{"untalented", "", 0, false, false},
		{"Venom and Flash of Steel 5/5", venomFlashOfSteelTalents, 5, true, false},
		{"Boot Knife 5/5", bootKnifeTalents, 0, false, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			sim, rogue, targets := newTankSim(t, c.talents)
			target := targets[0]

			// Room under the energy cap for what Shank restores.
			for rogue.CurrentEnergy() >= 40 {
				waitForGCD(t, sim, rogue)
				if !rogue.SinisterStrike.Cast(sim, target) {
					t.Fatal("Sinister Strike did not cast")
				}
			}
			waitForGCD(t, sim, rogue)
			before := rogue.CurrentEnergy()
			if !rogue.Shank.Cast(sim, target) {
				t.Fatal("Shank did not cast")
			}

			wantEnergy := shankRank.EffectN(2).BaseValue() + spellData.FlashOfSteel.EffectAt(1).ValueAt(c.flashOfSteel) -
				rogue.Shank.DefaultCast.Cost
			if got := rogue.CurrentEnergy() - before; got != wantEnergy {
				t.Errorf("Shank changes energy by %.0f, want %.0f", got, wantEnergy)
			}
			wantCooldown := shankRank.Cooldown() + spellData.FlashOfSteel.Rank(c.flashOfSteel).EffectN(2).TimeValue()
			if got := rogue.Shank.CD.Duration; got != wantCooldown {
				t.Errorf("Shank cools down for %v, want %v", got, wantCooldown)
			}
			if venom := rogue.GetAuraByID(core.ActionID{SpellID: venomBuff.ID}); (venom != nil && venom.IsActive()) != c.venom {
				t.Errorf("Venom is up after Shank: %v, want %v", venom != nil && venom.IsActive(), c.venom)
			}

			wantPoints := int32(0)
			if c.bootKnife {
				wantPoints = int32(bootKnifePoints.ProcEnergizeEffect().BaseValue())
			}
			for range 100 {
				waitForGCD(t, sim, rogue)
				rogue.Shank.CD.Reset()
				rogue.ResetComboPoints(sim, target, 0)
				if !rogue.Shank.Cast(sim, target) {
					t.Fatal("Shank did not cast")
				}
				if got := rogue.ComboPoints(target); got != wantPoints {
					t.Fatalf("Shank adds %d combo points, want %d", got, wantPoints)
				}
			}
			if m := rogue.Shank.SpellMetrics[target.UnitIndex]; m.Misses+m.Dodges+m.Parries+m.Blocks > 0 {
				t.Errorf("Shank was avoided: %d misses, %d dodges, %d parries, %d blocks", m.Misses, m.Dodges, m.Parries, m.Blocks)
			}
		})
	}
}

// Shank strikes with the off hand, so the off hand's poison hears it: over 100 Shanks, Deadly Poison casts off some.
func TestShankRollsTheOffHandPoison(t *testing.T) {
	player := &proto.Player{
		Name:        "Rogue",
		Race:        proto.Race_RaceHuman,
		Class:       proto.Class_ClassRogue,
		Equipment:   core.GetGearSet("../../ui/specs/rogue/dps/gear_sets", "muti").GearSet,
		Consumables: &proto.ConsumesSpec{OhWeaponImbueId: 8985}, // Deadly Poison IV
		Spec:        DefaultOptions,
		Rotation:    &proto.APLRotation{},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: core.LongDuration, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	rogue := sim.Raid.Parties[0].Players[0].(RogueAgent).GetRogue()
	target := sim.Encounter.AllTargetUnits[0]
	if rogue.DeadlyPoison == nil {
		t.Fatal("the off hand's imbue registered no Deadly Poison")
	}

	// Auto attacks also cast the poison between Shanks, so only the casts inside each Shank are counted.
	var poisons int32
	for range 100 {
		waitForGCD(t, sim, rogue)
		rogue.Shank.CD.Reset()
		before := rogue.DeadlyPoison.SpellMetrics[target.UnitIndex].Casts
		if !rogue.Shank.Cast(sim, target) {
			t.Fatal("Shank did not cast")
		}
		poisons += rogue.DeadlyPoison.SpellMetrics[target.UnitIndex].Casts - before
	}
	if poisons == 0 {
		t.Error("Deadly Poison cast off none of 100 Shanks")
	}
}

// Death Mark under 31 Subtlety points, with and without Coup de Grace 3/3.
const deathMarkTalents = "--335223031321001101"
const deathMarkCoupDeGraceTalents = "-3-335223031321001101"

// Death Mark spends its combo points, stores its share of the rogue's damage per point, twice for Rupture and none of
// another unit's, pays out what it stored up to the cap, and its payout never crits. Coup de Grace raises the stored
// share below the cap and the cap itself, each on its own.
func TestDeathMarkStoresAndPaysOut(t *testing.T) {
	for _, c := range []struct {
		name        string
		talents     string
		coupDeGrace int32
	}{
		{"without Coup de Grace", deathMarkTalents, 0},
		{"Coup de Grace 3/3", deathMarkCoupDeGraceTalents, 3},
	} {
		t.Run(c.name, func(t *testing.T) {
			sim, rogue, targets := newTankSim(t, c.talents)
			target := targets[0]
			rogue.AutoAttacks.CancelAutoSwing(sim)
			target.PseudoStats.BaseDodgeChance = 0
			rogue.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
			rogue.AddStatDynamic(sim, stats.PhysicalCritPercent, 100)
			coup := spellData.CoupDeGrace.EffectAt(2).MultiplierAt(c.coupDeGrace)
			deathMark := rogue.DeathMark
			dot := deathMark.Dot(target)

			mark := func(points int32) {
				t.Helper()
				for !rogue.GCD.IsReady(sim) {
					if sim.Step() {
						t.Fatal("the sim ended while the GCD was running")
					}
				}
				deathMark.CD.Reset()
				rogue.AddComboPoints(sim, target, points, deathMark.ComboPointMetrics())
				if !deathMark.Cast(sim, target) {
					t.Fatal("Death Mark did not cast")
				}
				if rogue.ComboPoints(target) != 0 || !dot.IsActive() {
					t.Fatalf("Death Mark leaves %d combo points and its mark up: %v; want 0 and up", rogue.ComboPoints(target), dot.IsActive())
				}
			}
			payout := func() float64 {
				before := deathMark.SpellMetrics[target.UnitIndex].TotalDamage
				dot.TickOnce(sim)
				return deathMark.SpellMetrics[target.UnitIndex].TotalDamage - before
			}

			const hit = 100.0
			mark(3)
			target.OnSpellHitTaken(sim, rogue.SinisterStrike, &core.SpellResult{Target: target, Damage: hit, Outcome: core.OutcomeHit})
			target.OnPeriodicDamageTaken(sim, rogue.Rupture, &core.SpellResult{Target: target, Damage: hit, Outcome: core.OutcomeHit})
			target.OnSpellHitTaken(sim, &core.Spell{Unit: targets[1]}, &core.SpellResult{Target: target, Damage: hit, Outcome: core.OutcomeHit})
			wantStored := 3 * deathMarkStoredPerPoint * coup * (hit + 2*hit)
			if got := payout(); math.Abs(got-wantStored) > 1e-6 {
				t.Errorf("Death Mark at 3 points pays out %.4f, want %.4f, the stored share with Coup de Grace on it", got, wantStored)
			}

			mark(5)
			target.OnSpellHitTaken(sim, rogue.SinisterStrike, &core.SpellResult{Target: target, Damage: 1e7, Outcome: core.OutcomeHit})
			wantCap := (deathMarkCapBase + 5*deathMarkCapPerPoint + 5*deathMarkCapAPPerPoint*deathMark.MeleeAttackPower(target)) *
				(1 + rogue.GetStat(stats.PhysicalCritPercent)/100) * coup
			if got := payout(); math.Abs(got-wantCap) > 1e-6 {
				t.Errorf("Death Mark at 5 points pays out %.4f, want its cap %.4f, raised by Coup de Grace", got, wantCap)
			}

			if crits := deathMark.SpellMetrics[target.UnitIndex].Crits; crits > 0 {
				t.Errorf("Death Mark's payout crit %d times at +100%% crit, want never", crits)
			}
		})
	}
}

// Bonescythe Armor's 4-piece bonus: a crit of Sinister Strike or Backstab regains the energy of the spell its row
// casts, while a Sinister Strike that does not crit and a crit of Shank, which the row's class flags also name, regain
// none.
func TestBonescytheHeadRushRegainsEnergyOnTheTooltipsCrits(t *testing.T) {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	for slot, id := range map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotHead:     22478, // Bonescythe Helmet
		proto.ItemSlot_ItemSlotShoulder: 22479, // Bonescythe Pauldrons
		proto.ItemSlot_ItemSlotChest:    22476, // Bonescythe Breastplate
		proto.ItemSlot_ItemSlotLegs:     22477, // Bonescythe Legplates
		proto.ItemSlot_ItemSlotMainHand: 2632,  // Curved Dagger
		proto.ItemSlot_ItemSlotOffHand:  18392, // Distracting Dagger
	} {
		items[slot].Id = id
	}
	player := &proto.Player{
		Name:      "Rogue",
		Race:      proto.Race_RaceHuman,
		Class:     proto.Class_ClassRogue,
		Equipment: &proto.EquipmentSpec{Items: items},
		Spec:      DefaultOptions,
		Rotation:  &proto.APLRotation{},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: core.LongDuration, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	rogue := sim.Raid.Parties[0].Players[0].(RogueAgent).GetRogue()
	target := sim.Encounter.AllTargetUnits[0]
	rogue.AutoAttacks.CancelAutoSwing(sim)
	target.PseudoStats.BaseDodgeChance = 0
	rogue.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)

	energy := bonescytheHeadRush.FiredSpell()
	perProc := energy.EnergizeEffect().BaseValue()
	regained := func() *proto.ResourceMetrics {
		for _, r := range rogue.Metrics.ToProto().Resources {
			if r.Id.GetSpellId() == energy.ID && r.Type == proto.ResourceType_ResourceTypeEnergy {
				return r
			}
		}
		return &proto.ResourceMetrics{}
	}

	// Strikes once and runs the sim past the row's cooldown and the spell batch window, so the next strike can proc.
	strike := func(spell *core.Spell) (crit bool, procs int32) {
		t.Helper()
		crits, before := spell.SpellMetrics[target.UnitIndex].Crits, regained().Events
		spell.SkipCastAndApplyEffects(sim, target)
		stepPastBatch(sim, bonescytheHeadRush.ICD())
		return spell.SpellMetrics[target.UnitIndex].Crits > crits, regained().Events - before
	}

	backstab := rogue.GetSpell(core.ActionID{SpellID: backstabBookRank.ID})
	rogue.AddStatDynamic(sim, stats.PhysicalCritPercent, 200)
	for _, c := range []struct {
		name  string
		spell *core.Spell
		want  int32
	}{
		{"Sinister Strike", rogue.SinisterStrike, 1},
		{"Backstab", backstab, 1},
		{"Shank", rogue.Shank, 0},
	} {
		crit, procs := strike(c.spell)
		if !crit {
			t.Fatalf("%s did not crit at +200%% crit", c.name)
		}
		if procs != c.want {
			t.Errorf("a %s crit regained energy %d times, want %d", c.name, procs, c.want)
		}
	}

	rogue.AddStatDynamic(sim, stats.PhysicalCritPercent, -400)
	if crit, procs := strike(rogue.SinisterStrike); crit || procs != 0 {
		t.Errorf("a Sinister Strike at -200%% crit crit %v and regained energy %d times, want no crit and none", crit, procs)
	}

	if got := regained(); got.Gain != float64(got.Events)*perProc {
		t.Errorf("the 4-piece regained %v energy over %d procs, want the row's %v each", got.Gain, got.Events, perProc)
	}
}

// Bonescythe Armor's 8-piece bonus: a landed Eviscerate grants Revealed Flaw 28815 at 28814's chance per combo point
// it spent, and never at none. Its charge raises the crit of Sinister Strike and Backstab by 28815's percent and
// leaves Shank's, which the row's class flags also name, alone; a Sinister Strike crit spends it. A second grant
// waits out 28815's cooldown.
func TestBonescytheRevealedFlawFollowsEviscerate(t *testing.T) {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	for slot, id := range map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotHead:     22478, // Bonescythe Helmet
		proto.ItemSlot_ItemSlotShoulder: 22479, // Bonescythe Pauldrons
		proto.ItemSlot_ItemSlotChest:    22476, // Bonescythe Breastplate
		proto.ItemSlot_ItemSlotLegs:     22477, // Bonescythe Legplates
		proto.ItemSlot_ItemSlotFeet:     22480, // Bonescythe Sabatons
		proto.ItemSlot_ItemSlotHands:    22481, // Bonescythe Gauntlets
		proto.ItemSlot_ItemSlotWaist:    22482, // Bonescythe Waistguard
		proto.ItemSlot_ItemSlotWrist:    22483, // Bonescythe Bracers
		proto.ItemSlot_ItemSlotMainHand: 2632,  // Curved Dagger
		proto.ItemSlot_ItemSlotOffHand:  18392, // Distracting Dagger
	} {
		items[slot].Id = id
	}
	player := &proto.Player{
		Name:      "Rogue",
		Race:      proto.Race_RaceHuman,
		Class:     proto.Class_ClassRogue,
		Equipment: &proto.EquipmentSpec{Items: items},
		Spec:      DefaultOptions,
		Rotation:  &proto.APLRotation{},
	}
	// Long enough to wait out the cooldown once.
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 600, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	rogue := sim.Raid.Parties[0].Players[0].(RogueAgent).GetRogue()
	target := sim.Encounter.AllTargetUnits[0]
	rogue.AutoAttacks.CancelAutoSwing(sim)
	target.PseudoStats.BaseDodgeChance = 0
	rogue.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)

	buff := rogue.GetAuraByID(core.ActionID{SpellID: revealedFlaw.ID})
	if buff == nil || bonescytheFlaw.EffectN(1).PointsPerResource <= 0 {
		t.Fatal("eight pieces register no 28815, or 28814 states no chance per combo point")
	}
	eviscerate := rogue.GetSpell(core.ActionID{SpellID: eviscerateBookRank.ID})
	eviscerateWith := func(points int32) {
		t.Helper()
		if points > 0 {
			rogue.AddComboPoints(sim, target, points, eviscerate.ComboPointMetrics())
		}
		eviscerate.SkipCastAndApplyEffects(sim, target)
		stepPastBatch(sim, 0)
	}
	grant := func() {
		t.Helper()
		for tries := 0; !buff.IsActive(); tries++ {
			if tries > 200 {
				t.Fatal("200 Eviscerates at 5 combo points granted no Revealed Flaw")
			}
			eviscerateWith(5)
		}
	}

	for range 200 {
		eviscerateWith(0)
	}
	if buff.IsActive() {
		t.Fatal("Eviscerates at no combo points granted Revealed Flaw, want none")
	}

	backstab := rogue.GetSpell(core.ActionID{SpellID: backstabBookRank.ID})
	strikes := []*core.Spell{rogue.SinisterStrike, backstab, rogue.Shank}
	before := make([]float64, len(strikes))
	for i, spell := range strikes {
		before[i] = spell.BonusCritPercent
	}
	grant()
	if buff.GetStacks() != 1 {
		t.Errorf("Revealed Flaw holds %d charges, want the row's 1", buff.GetStacks())
	}
	bonus := revealedFlaw.EffectN(1).BaseValue()
	for i, want := range []float64{bonus, bonus, 0} {
		if got := strikes[i].BonusCritPercent - before[i]; got != want {
			t.Errorf("Revealed Flaw raises %v's crit by %v%%, want %v%%", strikes[i].ActionID, got, want)
		}
	}

	// The rogue's own crit covers what the target's level takes off the charge's 100%.
	rogue.AddStatDynamic(sim, stats.PhysicalCritPercent, 10)
	crits := rogue.SinisterStrike.SpellMetrics[target.UnitIndex].Crits
	rogue.SinisterStrike.SkipCastAndApplyEffects(sim, target)
	stepPastBatch(sim, 0)
	if rogue.SinisterStrike.SpellMetrics[target.UnitIndex].Crits == crits || buff.IsActive() {
		t.Fatal("a Sinister Strike under Revealed Flaw did not crit and spend the charge")
	}

	// 28815's cooldown does not hold the next grant back.
	grant()
}

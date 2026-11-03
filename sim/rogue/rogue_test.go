package rogue

import (
	"testing"

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
			Equipment: core.GetGearSet("../../ui/specs/rogue/dps/gear_sets", "p1_muti").GearSet,
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
		if rogue.ComboPoints() != 1 || !backstab.Matches(RogueSpellBackstab) {
			t.Errorf("%s leaves %d combo points, or lost the Backstab mask; want 1", rank.RankedName(), rogue.ComboPoints())
		}
	}

	eviscerate := func(rank *spelldata.Spell) float64 {
		sim, rogue, spell := newRogue(rank)
		target := rogue.CurrentTarget
		rogue.AddComboPoints(sim, 5, rogue.NewComboPointMetrics(core.ActionID{SpellID: 1}))
		if !spell.Cast(sim, target) {
			t.Fatalf("%s did not cast", rank.RankedName())
		}
		if rogue.ComboPoints() != 0 {
			t.Errorf("%s leaves %d combo points, want 0", rank.RankedName(), rogue.ComboPoints())
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
			GearSet:       core.GetGearSet("../../ui/specs/rogue/dps/gear_sets", "p1_muti"),
			OtherGearSets: []core.GearSetCombo{
				//core.GetGearSet("../../ui/specs/rogue/dps/gear_sets", "p1"),
			},
			Talents: DefaultTalents,
			OtherTalentSets: []core.TalentsCombo{
				{
					Label:   "Subtlety BS",
					Talents: "0053031-3203-4320203300013211051",
				},
				{
					Label:   "Asn Muti",
					Talents: "00532310151501031-302303202014",
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

var DefaultTalents = "005323101012-32030320201515231"

var DefaultConsumables = &proto.ConsumesSpec{
	FlaskId:         274274, // Flask of Natural Aggression
	FoodId:          250072, // Swiftstrike Steak
	OffensivePotId:  250943, // Major Frenzy Potion
	ConjuredId:      7676,
	MhWeaponImbueId: 8928, // Instant Poison VI
	OhWeaponImbueId: 8985, // Deadly Poison IV
}

// Lightning Reflexes 5/5, Deflection 3/3, Riposte, Setup 3/3 and Ghostly Strike.
const tankTalents = "-00503001-000300001"

// A rogue with the talents given, facing two targets.
func newTankSim(t *testing.T, talents string) (*core.Simulation, *Rogue, []*core.Unit) {
	t.Helper()

	player := &proto.Player{
		Name:          "Tank",
		Race:          proto.Race_RaceHuman,
		Class:         proto.Class_ClassRogue,
		Equipment:     core.GetGearSet("../../ui/specs/rogue/dps/gear_sets", "p1_muti").GearSet,
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

// Setup hears the current target's attacks only, and at 3/3 every dodge of one adds the row's point.
func TestSetupAddsAPointOnTheCurrentTargetsDodgedSwings(t *testing.T) {
	want := int32(setupTriggered.ProcEnergizeEffect().BaseValue())
	for _, c := range []struct {
		name      string
		fromIndex int
		points    int32
	}{
		{"a dodge of the current target's swing adds a point", 0, want},
		{"a dodge of another target's swing adds none", 1, 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			sim, rogue, targets := newTankSim(t, tankTalents)
			if rogue.CurrentTarget != targets[0] {
				t.Fatal("the rogue's current target is not the first target")
			}
			before := rogue.ComboPoints()

			swingAt(sim, rogue, targets[c.fromIndex], core.OutcomeDodge)
			passBatchWindow(t, sim)

			if got := rogue.ComboPoints() - before; got != c.points {
				t.Errorf("combo points gained = %d, want %d", got, c.points)
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
	if got, want := rogue.GetStat(stats.DodgePercent)-before, ghostlyStrikeRank.EffectN(2).BaseValue(); got != want {
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

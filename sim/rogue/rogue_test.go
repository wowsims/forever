package rogue

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/common"
	_ "github.com/wowsims/forever/sim/common" // imported to get item effects included.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

func init() {
	RegisterRogue()
	common.RegisterAllEffects()
}

func TestRogue(t *testing.T) {
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

			Rotation: core.GetAplRotation("../../ui/specs/rogue/dps/apls", "default"),
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

// Runs the sim strictly past one spell batch, so every deferred proc due in it has fired whatever the
// order of actions at the same time.
func stepPastBatch(sim *core.Simulation) {
	at := sim.CurrentTime + core.SpellBatchWindow
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
	stepPastBatch(sim)
	if buff.IsActive() {
		t.Error("a death with no kill credit put up Remorseless")
	}

	dying.Enable(sim)
	dying.Die(sim, 1)
	stepPastBatch(sim)
	if got := strike.BonusCritPercent - base; !buff.IsActive() || math.Abs(got-bonus) > 1e-9 {
		t.Fatalf("after a killing blow Remorseless is up %v and adds %v crit to Sinister Strike, want up and %v", buff.IsActive(), got, bonus)
	}

	strike.SkipCastAndApplyEffects(sim, other)
	stepPastBatch(sim)
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

// Quietus reads the health of each strike's own target: the bonus is on for a strike at a target
// below its threshold and off for one at a target above it.
func TestQuietusFollowsTheStrikesTarget(t *testing.T) {
	sim, rogue := newTwoTargetSim("--00000000000000005")
	full, low := sim.Encounter.AllTargetUnits[0], sim.Encounter.AllTargets[1]
	low.SetHealthCurve(sim, 0, sim.Duration, core.HealthKnot{At: 0, Percent: 30}, core.HealthKnot{At: 1, Percent: 0})
	strike := rogue.SinisterStrike
	base := strike.DamageMultiplierAdditive
	bonus := spellData.Quietus.EffectAt(1).FractionAt(5)

	for _, c := range []struct {
		target *core.Unit
		want   float64
	}{
		{full, 0},
		{&low.Unit, bonus},
		{full, 0},
	} {
		strike.SkipCastAndApplyEffects(sim, c.target)
		if got := strike.DamageMultiplierAdditive - base; math.Abs(got-c.want) > 1e-9 {
			t.Errorf("Sinister Strike at %s adds %v to its damage, want %v", c.target.Label, got, c.want)
		}
	}
}

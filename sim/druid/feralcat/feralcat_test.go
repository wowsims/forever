package feralcat

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	goproto "google.golang.org/protobuf/proto"
)

func init() {
	RegisterFeralCatDruid()
	common.RegisterAllEffects()
}

func TestFeralCat(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{{
		Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
		GearSet:     core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "default"),
		Talents:     DefaultTalents,
		SpecOptions: core.SpecOptionsCombo{Label: "Standard", SpecOptions: DefaultSpecOptions},
		OtherSpecOptions: []core.SpecOptionsCombo{
			{Label: "BleedAssumed", SpecOptions: &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{
				Options: &proto.FeralCatDruid_Options{AssumeBleedActive: true},
			}}},
			{Label: "PrepullClearcasting", SpecOptions: &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{
				Options: &proto.FeralCatDruid_Options{PrepullClearcasting: true},
			}}},
		},
		OtherTalentSets: []core.TalentsCombo{
			{Label: "Hybrid", Talents: HybridTalents},
		},
		Rotation:    core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "adaptive"),
		Consumables: &proto.ConsumesSpec{},
		ItemFilter: core.ItemFilter{
			ArmorType: proto.ArmorType_ArmorTypeLeather,
			WeaponTypes: []proto.WeaponType{
				proto.WeaponType_WeaponTypeDagger, proto.WeaponType_WeaponTypeFist,
				proto.WeaponType_WeaponTypeMace, proto.WeaponType_WeaponTypeStaff,
			},
			RangedWeaponTypes: []proto.RangedWeaponType{proto.RangedWeaponType_RangedWeaponTypeIdol},
		},
	}}))
}
func BenchmarkSimulate(b *testing.B) {
	rsr := &proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Class:         proto.Class_ClassDruid,
				Race:          proto.Race_RaceNightElf,
				TalentsString: DefaultTalents,
				Equipment:     core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "default").GearSet,
				Consumables:   DefaultConsumables,
				Spec:          DefaultSpecOptions,
				Rotation:      core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "adaptive").Rotation,
			},
			nil, nil, nil,
		),
		Encounter: &proto.Encounter{
			Duration: 300,
			Targets:  []*proto.Target{core.NewDefaultTarget()},
		},
		SimOptions: core.AverageDefaultSimTestOptions,
	}

	core.RaidBenchmark(b, rsr)
}

// The cat UI's Cat and Hybrid presets (presets.ts CatTalents and HybridTalents).
const DefaultTalents = "050022-35200032121130212051-053"
const HybridTalents = "050022-35230332121132212051-"

var DefaultSpecOptions = &proto.Player_FeralCatDruid{
	FeralCatDruid: &proto.FeralCatDruid{
		Options: &proto.FeralCatDruid_Options{},
	},
}

var DefaultConsumables = &proto.ConsumesSpec{
	ResourcePotId: 13444,                 // Major Mana Potion
	ElixirIds:     []int32{13452, 13447}, // Elixir of the Mongoose, Elixir of the Sages
	FoodId:        250072,                // Swiftstrike Steak
	MhImbueId:     12643,                 // Dense Weightstone
	ConjuredId:    7676,                  // Thistle Tea
	// Goblin Sapper Charge, Scroll of Strength IV. Scroll of Agility IV is left out, as in the UI:
	// it shares the Agility group with Elixir of the Mongoose, which comes first.
	ExtraIds: []int32{10646, 10310},
}

func TestAvailableThistleTeaRequiresSelection(t *testing.T) {
	if !core.WITH_DB {
		t.Skip("requires the committed game database")
	}
	for _, selected := range []int32{0, 7676} {
		t.Run(fmt.Sprintf("selected=%d", selected), func(t *testing.T) {
			consumables := goproto.Clone(DefaultConsumables).(*proto.ConsumesSpec)
			consumables.ConjuredId = selected
			consumables.ConjuredItems = []int32{7676}
			request := &proto.RaidSimRequest{
				Raid: core.SinglePlayerRaidProto(&proto.Player{
					Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
					TalentsString: DefaultTalents, Spec: DefaultSpecOptions, Consumables: consumables,
					Equipment: core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "default").GearSet,
					Rotation:  core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "adaptive").Rotation,
				}, nil, nil, nil),
				Encounter:  &proto.Encounter{Duration: 180, Targets: []*proto.Target{core.NewDefaultTarget()}},
				SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 101, IsTest: true},
			}
			result := core.RunRaidSim(request)
			if result.Error != nil {
				t.Fatal(result.Error.Message)
			}
			casts := int32(0)
			for _, action := range result.RaidMetrics.Parties[0].Players[0].Actions {
				if action.Id.GetItemId() == 7676 {
					for _, target := range action.Targets {
						casts += target.Casts
					}
				}
			}
			t.Logf("default selected=%d available=[7676] Tea casts=%d", selected, casts)
			if selected == 0 && casts != 0 {
				t.Fatalf("unselected Tea casts = %d, want 0", casts)
			}
			if selected == 7676 && casts == 0 {
				t.Fatal("selected Tea was never used")
			}
		})
	}
}

func newFormScenario(t *testing.T) (*core.Simulation, *FeralDruid) {
	t.Helper()
	if !core.WITH_DB {
		t.Skip("requires the committed game database")
	}
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 1613
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1},
		Raid: core.SinglePlayerRaidProto(&proto.Player{
			Name: "Cat", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
			Spec: DefaultSpecOptions, Equipment: &proto.EquipmentSpec{Items: items},
			Consumables: &proto.ConsumesSpec{}, Buffs: &proto.IndividualBuffs{}, Rotation: &proto.APLRotation{},
		}, &proto.PartyBuffs{}, nil, nil),
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 60},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*FeralDruid)
}

func fillFormScenarioResources(sim *core.Simulation, cat *FeralDruid) {
	cat.AddEnergy(sim, cat.MaximumEnergy(), cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	cat.AddMana(sim, cat.MaxMana(), cat.NewManaMetrics(core.ActionID{SpellID: 1}))
	cat.AddRage(sim, cat.MaximumRage(), cat.NewRageMetrics(core.ActionID{SpellID: 1}))
}

func TestFeralFormsUseFeralSkillAndRestoreEquippedSkill(t *testing.T) {
	sim, cat := newFormScenario(t)
	cat.ClearForm(sim)
	cat.PseudoStats.WeaponSkills[proto.WeaponSkill_WeaponSkillStaves] = -20
	cat.SetFeralCombat(false)
	table := cat.AttackTables[cat.CurrentTarget.UnitIndex]
	if got := table.MH.BaseMissChance; got != 0.09 {
		t.Fatalf("humanoid staff miss = %v, want 0.09", got)
	}

	fillFormScenarioResources(sim, cat)
	if !cat.CatForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Cat Form did not cast")
	}
	if got := table.MH.BaseMissChance; got != 0.05 {
		t.Fatalf("Cat Form miss = %v, want 0.05", got)
	}

	cat.ClearForm(sim)
	sim.CurrentTime += 2 * time.Second
	fillFormScenarioResources(sim, cat)
	if !cat.BearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Bear Form did not cast")
	}
	if got := table.MH.BaseMissChance; got != 0.05 {
		t.Fatalf("Bear Form miss = %v, want 0.05", got)
	}

	cat.ClearForm(sim)
	if got := table.MH.BaseMissChance; got != 0.09 {
		t.Fatalf("restored humanoid staff miss = %v, want 0.09", got)
	}
}

func TestFeralAttackPowerConvertsOneForOneOnlyInFeralForms(t *testing.T) {
	// Given a cat with 77 Feral Attack Power added after initialization.
	sim, cat := newFormScenario(t)
	catAP := cat.GetStat(stats.AttackPower)
	cat.AddStatDynamic(sim, stats.FeralAttackPower, 77)

	// When the druid moves through Cat, humanoid, Bear, and humanoid forms.
	if got := cat.GetStat(stats.AttackPower) - catAP; got != 77 {
		t.Errorf("Cat Form gained %v attack power, want 77", got)
	}
	cat.ClearForm(sim)
	humanoidAP := cat.GetStat(stats.AttackPower)
	cat.AddStatDynamic(sim, stats.FeralAttackPower, -77)
	if got := cat.GetStat(stats.AttackPower); got != humanoidAP {
		t.Errorf("removing Feral Attack Power in humanoid form changed attack power from %v to %v", humanoidAP, got)
	}
	cat.AddStatDynamic(sim, stats.FeralAttackPower, 77)
	fillFormScenarioResources(sim, cat)
	if !cat.BearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Bear Form did not cast")
	}
	bearAP := cat.GetStat(stats.AttackPower)
	cat.AddStatDynamic(sim, stats.FeralAttackPower, -77)
	if got := bearAP - cat.GetStat(stats.AttackPower); got != 77 {
		t.Errorf("Bear Form gained %v attack power, want 77", got)
	}
	cat.AddStatDynamic(sim, stats.FeralAttackPower, 77)
	cat.ClearForm(sim)

	// Then the form-only dependency is removed again on exit.
	secondHumanoidAP := cat.GetStat(stats.AttackPower)
	cat.AddStatDynamic(sim, stats.FeralAttackPower, -77)
	if got := cat.GetStat(stats.AttackPower); got != secondHumanoidAP {
		t.Errorf("removing Feral Attack Power after Bear changed humanoid attack power from %v to %v", secondHumanoidAP, got)
	}
}

func TestDashSpeedExpiresAfterItsDuration(t *testing.T) {
	sim, cat := newFormScenario(t)
	wantExpires := sim.CurrentTime + time.Duration(spelldata.MustFind(cat.Dash.SpellID).DurationMs)*time.Millisecond
	before := cat.GetMovementSpeed()
	if !cat.Dash.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Dash did not cast in Cat Form")
	}
	if got, want := cat.GetMovementSpeed(), before*1.6; math.Abs(got-want) > 1e-9 {
		t.Fatalf("Dash movement speed = %g, want %g", got, want)
	}
	for cat.DashAura.IsActive() {
		if sim.Step() {
			t.Fatal("simulation ended before Dash expired")
		}
	}
	if sim.CurrentTime != wantExpires {
		t.Fatalf("Dash expired at %s, want %s from its row", sim.CurrentTime, wantExpires)
	}
	if got := cat.GetMovementSpeed(); math.Abs(got-before) > 1e-9 {
		t.Fatalf("movement after Dash = %g, want %g", got, before)
	}
	t.Logf("Dash speed %g -> %g -> %g; expired at %s", before, before*1.6, cat.GetMovementSpeed(), sim.CurrentTime)
}

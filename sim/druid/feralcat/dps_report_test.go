//go:build with_db

package feralcat

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

// The UI's default cat (spec.ts and presets.ts): a Night Elf in the Pre-raid set with its raid buffs,
// debuffs and consumables, Engineering and Enchanting, a bleed assumed on the target, against a level
// 63 target.
func uiDefaultRaidSim(talents string, rotation *proto.APLRotation, duration, variation float64, iterations int32) *proto.RaidSimRequest {
	target := make([]float64, stats.SimStatsLen)
	target[stats.AttackPower] = 320
	target[stats.BlockValue] = 54
	target[stats.Armor] = 3731
	target[stats.Health] = 1299000
	return &proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 101, Iterations: iterations},
		Raid: &proto.Raid{
			Parties: []*proto.Party{{
				Players: []*proto.Player{{
					Name:           "Cat",
					Class:          proto.Class_ClassDruid,
					Race:           proto.Race_RaceNightElf,
					TalentsString:  talents,
					ReactionTimeMs: 250,
					Profession1:    proto.Profession_Engineering,
					Profession2:    proto.Profession_Enchanting,
					Equipment:      core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "pre_raid").GearSet,
					Consumables: &proto.ConsumesSpec{
						ResourcePotId: 13442,  // Mighty Rage Potion
						FlaskId:       274275, // Flask of Natural Precision
						ElixirIds:     []int32{13452},
						FoodId:        250072, // Swiftstrike Steak
						MhImbueId:     18262,
						ConjuredId:    7676,
						ExtraIds:      []int32{12451, 12460, 8412},
					},
					Buffs:    &proto.IndividualBuffs{GreaterBlessingOfKings: true, GreaterBlessingOfMight: true},
					Spec:     &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{Options: &proto.FeralCatDruid_Options{AssumeBleedActive: true}}},
					Rotation: rotation,
				}},
				Buffs: &proto.PartyBuffs{
					ManaSpringTotem: proto.TristateEffect_TristateEffectRegular,
				},
			}},
			Buffs: &proto.RaidBuffs{BattleShout: 1, ArcaneBrilliance: true, PrayerOfSpirit: true, GiftOfTheWild: true, PrayerOfFortitude: true,
				StrengthOfEarthTotem: true, WindfuryTotem: true},
			Debuffs: &proto.Debuffs{JudgementOfWisdom: true, FaerieFire: true, SunderArmor: true, GiftOfArthas: true},
		},
		Encounter: &proto.Encounter{Duration: duration, DurationVariation: variation, Targets: []*proto.Target{{
			Id: 31146, Name: "Raid Target", Level: 63, MobType: proto.MobType_MobTypeMechanical, Stats: target,
			MinBaseDamage: 15113, DamageSpread: 0.5, SwingSpeed: 2, ParryHaste: true, CanCrush: true,
		}}},
	}
}

// DPS_REPORT=1 prints the cat preset and the bearweave preset on the UI default over the fight lengths
// the presets are judged on, as the markdown tables a change's pull request comment carries.
// DPS_REPORT_ITERATIONS sets the iterations of each fixed-length row (10000); the 180s ± 5 row runs
// twice that.
func TestDpsReport(t *testing.T) {
	if os.Getenv("DPS_REPORT") == "" {
		t.Skip("set DPS_REPORT=1 to print the preset DPS report")
	}
	iterations := int32(10000)
	if n, err := strconv.Atoi(os.Getenv("DPS_REPORT_ITERATIONS")); err == nil && n > 0 {
		iterations = int32(n)
	}
	presets := []struct {
		name, talents, apl string
	}{
		{"Cat", catTalents, "default"},
		{"Bearweave", catTalents, "bearweave"},
	}
	run := func(talents, apl string, duration, variation float64, iterations int32) (*proto.UnitMetrics, float64) {
		rotation := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", apl).Rotation
		result := core.RunSim(uiDefaultRaidSim(talents, rotation, duration, variation, iterations), nil, simsignals.CreateSignals())
		if result.Error != nil {
			t.Fatal(result.Error.Message)
		}
		return result.RaidMetrics.Parties[0].Players[0], float64(result.IterationsDone)
	}

	fmt.Printf("\n| Length | %s | %s | Difference | Trips | Maul | Innervate (cat) | Shifting Power (cat) |\n", presets[0].name, presets[1].name)
	fmt.Println("| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |")
	for _, duration := range []float64{30, 45, 60, 90, 120, 180, 300} {
		cat, n := run(presets[0].talents, presets[0].apl, duration, 0, iterations)
		weave, m := run(presets[1].talents, presets[1].apl, duration, 0, iterations)
		fmt.Printf("| %.0fs | %.2f | %.2f | %+.1f | %.2f | %.2f | %.2f | %.2f |\n", duration, cat.Dps.Avg, weave.Dps.Avg, weave.Dps.Avg-cat.Dps.Avg,
			float64(castsOf(weave.Actions, spellDireBear))/m, float64(castsOf(weave.Actions, spellMaul))/m,
			float64(castsOf(cat.Actions, spellInnervate))/n, float64(castsOf(cat.Actions, spellSP))/n)
	}
	cat, _ := run(presets[0].talents, presets[0].apl, 180, 5, 2*iterations)
	weave, _ := run(presets[1].talents, presets[1].apl, 180, 5, 2*iterations)
	fmt.Printf("| 180s ± 5 | %.2f | %.2f | %+.1f | | | | |\n\n", cat.Dps.Avg, weave.Dps.Avg, weave.Dps.Avg-cat.Dps.Avg)
}

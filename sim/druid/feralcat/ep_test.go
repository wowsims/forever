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

// Prints DPS EP for the UI's default cat: the Pre-raid set, Cat talents and Monocat APL, its buffs,
// debuffs and consumables, a level 63 target, 180s, normalized to agility. The UI default is over the
// special-attack hit cap, so hit is weighed again with the hit taken 2% below the cap.
// Pass the UI's iteration count; stat weights then split it above and below the baseline.
// EP_GEAR names another of the UI's gear sets (default is the Level 60 set).
func TestDefaultEp(t *testing.T) {
	if os.Getenv("EP_SWEEP") == "" {
		t.Skip("set EP_SWEEP=1 (and EP_SWEEP_ITERATIONS) to compute default EP")
	}
	iterations := int32(12500)
	if n, err := strconv.Atoi(os.Getenv("EP_SWEEP_ITERATIONS")); err == nil && n > 0 {
		iterations = int32(n)
	}
	weigh := []proto.Stat{
		proto.Stat_StatStrength,
		proto.Stat_StatAgility,
		proto.Stat_StatAttackPower,
		proto.Stat_StatFeralAttackPower,
		proto.Stat_StatPhysicalDamage,
		proto.Stat_StatHitRating,
		proto.Stat_StatExpertiseRating,
		proto.Stat_StatCritRating,
		proto.Stat_StatHasteRating,
		proto.Stat_StatArmorPenetration,
	}
	request := func(iterations int32) *proto.RaidSimRequest {
		req := catEPRequest(catTalents, defaultCatRotation(), 180, 0, iterations)
		if gear := os.Getenv("EP_GEAR"); gear != "" {
			req.Raid.Parties[0].Players[0].Equipment = core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", gear).GearSet
		}
		return req
	}
	run := func(base *proto.RaidSimRequest) *proto.StatWeightsResult {
		// IsTest caps the sim at 3 workers. Labeled RNG is turned on inside stat weights.
		base.SimOptions.IsTest = false
		result := core.StatWeights(&proto.StatWeightsRequest{
			Player:          base.Raid.Parties[0].Players[0],
			RaidBuffs:       base.Raid.Buffs,
			PartyBuffs:      base.Raid.Parties[0].Buffs,
			Debuffs:         base.Raid.Debuffs,
			Encounter:       base.Encounter,
			SimOptions:      base.SimOptions,
			StatsToWeigh:    weigh,
			EpReferenceStat: proto.Stat_StatAgility,
		})
		if result.Error != nil {
			t.Fatal(result.Error.Message)
		}
		return result
	}
	result := run(request(iterations))

	sim := core.NewSim(request(1), simsignals.CreateSignals())
	sim.Reset()
	hit := sim.Raid.Parties[0].Players[0].GetCharacter().GetStat(stats.PhysicalHitPercent)
	below := request(iterations)
	bonus := make([]float64, stats.SimStatsLen)
	bonus[stats.HitRating] = (6 - hit) * core.HitRatingPerHitPercent
	below.Raid.Parties[0].Players[0].BonusStats = &proto.UnitStats{Stats: bonus}
	hitResult := run(below)

	fmt.Printf("hit %.2f%%; hit weighed at 6%%\n", hit)
	fmt.Printf("%-22s %12s %12s %10s %10s\n", "stat", "dps/point", "stdev", "ep", "ep stdev")
	for _, s := range weigh {
		i := int(s)
		r := result
		if s == proto.Stat_StatHitRating {
			r = hitResult
		}
		fmt.Printf("%-22s %12.4f %12.4f %10.4f %10.4f\n",
			stats.Stat(s).StatName(),
			r.Dps.Weights.Stats[i], r.Dps.WeightsStdev.Stats[i],
			r.Dps.EpValues.Stats[i], r.Dps.EpValuesStdev.Stats[i])
	}
}

// Pre-raid Cat measurement fixture with fixed buffs, consumables and a level 63 target.
func catEPRequest(talents string, rotation *proto.APLRotation, duration, variation float64, iterations int32) *proto.RaidSimRequest {
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
					Buffs:    &proto.IndividualBuffs{GreaterBlessingOfKings: true, GreaterBlessingOfMight: 1},
					Spec:     &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{Options: &proto.FeralCatDruid_Options{AssumeBleedActive: true}}},
					Rotation: rotation,
				}},
				Buffs: &proto.PartyBuffs{
					ManaSpringTotem: proto.TristateEffect_TristateEffectRegular,
				},
			}},
			Buffs: &proto.RaidBuffs{BattleShout: 1, ArcaneIntellect: true, PrayerOfSpirit: true, MarkOfTheWild: true, PowerWordFortitude: true,
				StrengthOfEarthTotem: 1, WindfuryTotem: true},
			Debuffs: &proto.Debuffs{JudgementOfWisdom: true, FaerieFire: true, SunderArmor: true, GiftOfArthas: true},
		},
		Encounter: &proto.Encounter{Duration: duration, DurationVariation: variation, Targets: []*proto.Target{{
			Id: 31146, Name: "Raid Target", Level: 63, MobType: proto.MobType_MobTypeMechanical, Stats: target,
			MinBaseDamage: 15113, DamageSpread: 0.5, SwingSpeed: 2, ParryHaste: true, CanCrush: true,
		}}},
	}
}

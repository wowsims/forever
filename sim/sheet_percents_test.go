//go:build with_db

package sim

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func warriorSheetStats(t *testing.T, talents string) *proto.PlayerStats {
	t.Helper()
	gear := core.GetGearSet("../ui/specs/warrior/dps/gear_sets", "p1_fury").GearSet
	if len(gear.Items) == 0 {
		t.Fatal("the fury gear set did not load")
	}
	result := core.ComputeStats(&proto.ComputeStatsRequest{
		Raid: &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{{
			Name:          "Cruelty Tester",
			Class:         proto.Class_ClassWarrior,
			Race:          proto.Race_RaceOrc,
			Equipment:     gear,
			TalentsString: talents,
			Buffs:         &proto.IndividualBuffs{},
			Spec: &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{Options: &proto.DpsWarrior_Options{
				ClassOptions: &proto.WarriorOptions{DefaultStance: proto.WarriorStance_WarriorStanceBerserker}}}},
		}}, Buffs: &proto.PartyBuffs{}}}, Buffs: &proto.RaidBuffs{}},
		Encounter: STEncounter,
	})
	if result.ErrorResult != "" {
		t.Fatal(result.ErrorResult)
	}
	return result.RaidStats.Parties[0].Players[0]
}

func TestSheetShowsCrueltyFromTheTalentsStage(t *testing.T) {
	without := warriorSheetStats(t, "")
	with := warriorSheetStats(t, "-05")

	stages := []struct {
		name         string
		stage        func(*proto.PlayerStats) *proto.UnitStats
		crueltyShown float64
	}{
		{"base", (*proto.PlayerStats).GetBaseStats, 0},
		{"gear", (*proto.PlayerStats).GetGearStats, 0},
		{"talents", (*proto.PlayerStats).GetTalentsStats, 5},
		{"buffs", (*proto.PlayerStats).GetBuffsStats, 5},
		{"consumes", (*proto.PlayerStats).GetConsumesStats, 5},
		{"final", (*proto.PlayerStats).GetFinalStats, 5},
	}
	for _, stage := range stages {
		for _, tc := range []struct {
			pseudoStat proto.PseudoStat
			want       float64
		}{
			{proto.PseudoStat_PseudoStatMeleeCritPercent, stage.crueltyShown},
			{proto.PseudoStat_PseudoStatOffHandCritPercent, stage.crueltyShown},
			{proto.PseudoStat_PseudoStatMainHandCritRating, 0},
			{proto.PseudoStat_PseudoStatRangedCritPercent, 0},
		} {
			got := stats.PseudoStatValue(stage.stage(with).PseudoStats, tc.pseudoStat) - stats.PseudoStatValue(stage.stage(without).PseudoStats, tc.pseudoStat)
			if !core.WithinToleranceFloat64(tc.want, got, 1e-9) {
				t.Errorf("%s stage: Cruelty moves %s by %v, want %v", stage.name, tc.pseudoStat, got, tc.want)
			}
		}
	}
}

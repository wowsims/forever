//go:build with_db

package reforgeoptimizer

import (
	"math"
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/core/testagent"
	protopkg "google.golang.org/protobuf/proto"
)

// A protection warrior's gear on a character that can parry. The sheet reads nothing else off the
// warrior, and blocking comes with the shield.
func init() {
	testagent.RegisterAsProtectionWarrior(func(agent *testagent.Agent) {
		agent.PseudoStats.CanParry = true
	})
}

// subtractUnitStats diffs a sheet against the optimizer's base stats.
func subtractUnitStats(a core.UnitStats, b core.UnitStats) core.UnitStats {
	result := a
	result.Stats = a.Stats.Subtract(b.Stats)
	maxLen := max(len(a.PseudoStats), len(b.PseudoStats))
	result.PseudoStats = make([]float64, maxLen)
	copy(result.PseudoStats, a.PseudoStats)
	for idx, value := range b.PseudoStats {
		result.PseudoStats[idx] -= value
	}
	return result
}

func protectionWarriorRaid(gearSet string) *proto.Raid {
	return core.SinglePlayerRaidProto(&proto.Player{
		Class:     proto.Class_ClassWarrior,
		Race:      proto.Race_RaceOrc,
		Equipment: core.GetGearSet("../../../ui/specs/warrior/protection/gear_sets", gearSet).GearSet,
		Spec: &proto.Player_ProtectionWarrior{ProtectionWarrior: &proto.ProtectionWarrior{
			Options: &proto.ProtectionWarrior_Options{ClassOptions: &proto.WarriorOptions{}},
		}},
	}, nil, nil, nil)
}

func sheetStats(t *testing.T, raid *proto.Raid) core.UnitStats {
	t.Helper()
	result := computeReforgeStats(&proto.ComputeStatsRequest{Raid: raid})
	if result.ErrorResult != "" {
		t.Fatalf("ComputeStats: %s", result.ErrorResult)
	}
	return protoToCoreUnitStats(result.RaidStats.Parties[0].Players[0].FinalStats)
}

func TestCapSpaceDeltaMatchesTheSheet(t *testing.T) {
	unshielded := protectionWarriorRaid("preraid")
	unshielded.Parties[0].Players[0].Equipment.Items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{}

	block := proto.PseudoStat_PseudoStatBlockPercent
	dodge := proto.PseudoStat_PseudoStatDodgePercent
	parry := proto.PseudoStat_PseudoStatParryPercent
	critTaken := proto.PseudoStat_PseudoStatReducedCritTakenPercent
	for _, character := range []struct {
		name  string
		raid  *proto.Raid
		sheet core.SheetAvoidance
	}{
		{"warrior with a shield", protectionWarriorRaid("preraid"), core.SheetAvoidance{CanBlock: true, CanParry: true}},
		{"warrior without a shield", unshielded, core.SheetAvoidance{CanParry: true}},
	} {
		t.Run(character.name, func(t *testing.T) {
			baseResult, sdm, sheet := computeReforgeStatsAndDeps(&proto.ComputeStatsRequest{Raid: protopkg.Clone(character.raid).(*proto.Raid)})
			if baseResult.ErrorResult != "" {
				t.Fatalf("ComputeStats: %s", baseResult.ErrorResult)
			}
			if sheet != character.sheet {
				t.Fatalf("sheet shows %+v, want %+v", sheet, character.sheet)
			}
			o := &reforgeOptimizer{
				statDeps:  sdm,
				sheet:     sheet,
				baseStats: protoToCoreUnitStats(baseResult.RaidStats.Parties[0].Players[0].FinalStats),
			}

			for _, test := range []struct {
				stat  stats.Stat
				moves []proto.PseudoStat
			}{
				{stats.BlockRating, []proto.PseudoStat{block}},
				{stats.DefenseRating, []proto.PseudoStat{block, dodge, parry, critTaken}},
				{stats.DodgeRating, []proto.PseudoStat{dodge}},
				{stats.ParryRating, []proto.PseudoStat{parry}},
			} {
				var delta stats.Stats
				delta[test.stat] = 25

				bonusRaid := protopkg.Clone(character.raid).(*proto.Raid)
				bonusRaid.Parties[0].Players[0].BonusStats = &proto.UnitStats{Stats: delta[:stats.ProtoStatsLen]}
				sheetDelta := subtractUnitStats(sheetStats(t, bonusRaid), o.baseStats)
				capCoeffs := o.resolveCapCoeffs(delta)

				for _, pseudoStat := range []proto.PseudoStat{block, dodge, parry, critTaken} {
					moves := slices.Contains(test.moves, pseudoStat) &&
						(pseudoStat != block || sheet.CanBlock) &&
						(pseudoStat != parry || sheet.CanParry)
					want := getUnitStat(sheetDelta, stats.UnitStatFromPseudoStat(pseudoStat))
					if (want != 0) != moves {
						t.Errorf("%s moves the sheet's %s by %v", test.stat.StatName(), pseudoStat, want)
					}
					if got := capCoeffs[pseudoStatCoeffKey(pseudoStat)]; math.Abs(got-want) > 1e-9 {
						t.Errorf("%s moves the cap space's %s by %v, the sheet's by %v", test.stat.StatName(), pseudoStat, got, want)
					}
				}
			}
		})
	}
}

//go:build with_db

package feralbear

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// Prints DPS, threat and survival EP for the default bear: Level 60 gear, the APL preset, the UI's
// consumables and healing model, and the Level 60 boss. DPS and threat EP have agility as 1. Survival EP is
// TMI with stamina as 1, the way the EP dialog scales it with the bear's tank reference stat.
// Pass the UI's iteration count; stat weights then split it above and below the baseline.
func TestDefaultEp(t *testing.T) {
	if os.Getenv("EP_SWEEP") == "" {
		t.Skip("set EP_SWEEP=1 (and EP_SWEEP_ITERATIONS) to compute default EP")
	}
	iterations := int32(12500)
	if n, err := strconv.Atoi(os.Getenv("EP_SWEEP_ITERATIONS")); err == nil && n > 0 {
		iterations = int32(n)
	}
	base := bearRequest(bearAPL(bestBearAPL), iterations)
	// The UI default assumes a raid bleed for Rend and Tear.
	base.Raid.Parties[0].Players[0].GetFeralBearDruid().Options.AssumeBleedActive = true
	// IsTest caps the sim at 3 workers. Labeled RNG is turned on inside stat weights.
	base.SimOptions.IsTest = false
	// The spec's epStats.
	weigh := []proto.Stat{
		proto.Stat_StatStamina,
		proto.Stat_StatAgility,
		proto.Stat_StatStrength,
		proto.Stat_StatAttackPower,
		proto.Stat_StatFeralAttackPower,
		proto.Stat_StatArmor,
		proto.Stat_StatBonusArmor,
		proto.Stat_StatArmorNoMods,
		proto.Stat_StatDodgeRating,
		proto.Stat_StatDefenseRating,
		proto.Stat_StatHitRating,
		proto.Stat_StatCritRating,
		proto.Stat_StatHasteRating,
		proto.Stat_StatExpertiseRating,
		proto.Stat_StatPhysicalDamage,
		proto.Stat_StatArmorPenetration,
	}
	result := core.StatWeights(&proto.StatWeightsRequest{
		Player:             base.Raid.Parties[0].Players[0],
		RaidBuffs:          base.Raid.Buffs,
		PartyBuffs:         base.Raid.Parties[0].Buffs,
		Debuffs:            base.Raid.Debuffs,
		Encounter:          base.Encounter,
		SimOptions:         base.SimOptions,
		Tanks:              base.Raid.Tanks,
		StatsToWeigh:       weigh,
		PseudoStatsToWeigh: []proto.PseudoStat{proto.PseudoStat_PseudoStatMainHandDps},
		EpReferenceStat:    proto.Stat_StatAgility,
	})
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}
	dpsRef := result.Dps.Weights.Stats[proto.Stat_StatAgility]
	tpsRef := result.Tps.Weights.Stats[proto.Stat_StatAgility]
	tmiRef := result.Tmi.Weights.Stats[proto.Stat_StatStamina]
	deathRef := result.PDeath.Weights.Stats[proto.Stat_StatStamina]
	fmt.Printf("%-22s %10s %9s %9s | %10s %9s %9s | %10s %9s %9s | %12s %9s\n",
		"stat", "dps/pt", "dps ep", "ep stdev", "tps/pt", "tps ep", "ep stdev", "tmi/pt", "surv ep", "ep stdev", "death/pt", "death ep")
	row := func(name string, dps, dpsSd, tps, tpsSd, tmi, tmiSd, death float64) {
		fmt.Printf("%-22s %10.4f %9.4f %9.4f | %10.4f %9.4f %9.4f | %10.5f %9.4f %9.4f | %12.3e %9.4f\n",
			name, dps, dps/dpsRef, dpsSd/abs(dpsRef), tps, tps/tpsRef, tpsSd/abs(tpsRef),
			tmi, tmi/tmiRef, tmiSd/abs(tmiRef), death, death/deathRef)
	}
	for _, s := range weigh {
		i := int(s)
		row(stats.Stat(s).StatName(),
			result.Dps.Weights.Stats[i], result.Dps.WeightsStdev.Stats[i],
			result.Tps.Weights.Stats[i], result.Tps.WeightsStdev.Stats[i],
			result.Tmi.Weights.Stats[i], result.Tmi.WeightsStdev.Stats[i],
			result.PDeath.Weights.Stats[i])
	}
	i := int(proto.PseudoStat_PseudoStatMainHandDps)
	row("MainHandDps",
		result.Dps.Weights.PseudoStats[i], result.Dps.WeightsStdev.PseudoStats[i],
		result.Tps.Weights.PseudoStats[i], result.Tps.WeightsStdev.PseudoStats[i],
		result.Tmi.Weights.PseudoStats[i], result.Tmi.WeightsStdev.PseudoStats[i],
		result.PDeath.Weights.PseudoStats[i])
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

//go:build with_db

package reforgeoptimizer

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	protopkg "google.golang.org/protobuf/proto"
)

func TestOptimizeReturnsTheGearUnchanged(t *testing.T) {
	raid := protectionWarriorRaid("preraid")
	result := OptimizeAsync(&proto.ReforgeOptimizeRequest{Raid: raid}, simsignals.CreateSignals())
	if err := result.GetError(); err != nil {
		t.Fatalf("OptimizeAsync: %s", err.GetMessage())
	}

	if gear := result.GetOptimizedGear(); !protopkg.Equal(gear, raid.Parties[0].Players[0].Equipment) {
		t.Errorf("optimized gear %v, want the input gear", gear)
	}
	if result.GetScore() != 0 {
		t.Errorf("score %v, want 0", result.GetScore())
	}
	want := computeReforgeStats(&proto.ComputeStatsRequest{Raid: raid})
	if want.ErrorResult != "" {
		t.Fatalf("ComputeStats: %s", want.ErrorResult)
	}
	if !protopkg.Equal(result.GetOptimizedPlayerStats().GetFinalStats(), want.RaidStats.Parties[0].Players[0].GetFinalStats()) {
		t.Error("optimized stats differ from the input gear's stats")
	}
}

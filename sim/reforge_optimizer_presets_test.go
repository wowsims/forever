//go:build with_db

package sim

import (
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	assetsdb "github.com/wowsims/forever/assets/database"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	reforgeoptimizer "github.com/wowsims/forever/sim/core/reforge_optimizer"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
	protopkg "google.golang.org/protobuf/proto"
)

// The committed coverage fixtures (<name>.test.json) the optimizer is run against. Each is a real spec's
// preset, named after the optimizer behaviour it guards. Running a real spec's preset needs every
// class registered, so the test lives here and the optimizer's own package stays free of classes.
const reforgeFixturesDir = "core/reforge_optimizer/test-fixtures"

func TestReforgerOptimizer(t *testing.T) {
	if !slices.ContainsFunc(assetsdb.Load().Items, func(item *proto.UIItem) bool { return len(item.GemSockets) > 0 }) {
		t.Skip("no socketed items in the database - Forever encrypts item stats until discovery")
	}
	RegisterAll()

	paths, err := filepath.Glob(filepath.Join(reforgeFixturesDir, "*.test.json"))
	if err != nil {
		t.Fatalf("globbing fixtures: %v", err)
	}
	if len(paths) == 0 {
		t.Skip("no fixtures in " + reforgeFixturesDir)
	}
	sort.Strings(paths)

	for _, path := range paths {
		fileName := filepath.Base(path)
		t.Run(strings.TrimSuffix(fileName, ".test.json"), func(t *testing.T) {
			request := loadReforgePreset(t, fileName)
			expectedGear := request.GetRaid().GetParties()[0].GetPlayers()[0].GetEquipment()
			if expectedGear == nil {
				t.Fatal("preset has no player equipment to compare against")
			}

			result := reforgeoptimizer.OptimizeAsync(request, simsignals.CreateSignals())
			if err := result.GetError(); err != nil {
				t.Fatalf("Optimize returned error: %s", err.GetMessage())
			}
			optimizedGear := result.GetOptimizedGear()
			if optimizedGear == nil {
				t.Fatal("Optimize returned no optimized gear")
			}

			expectedRaid := protopkg.Clone(request.Raid).(*proto.Raid)
			expectedRaid.Parties[0].Players[0].Equipment = expectedGear
			expectedResult := core.ComputeStats(&proto.ComputeStatsRequest{Raid: expectedRaid, SkipRotation: true})
			if expectedResult.ErrorResult != "" {
				t.Fatalf("ComputeStats on expected gear failed: %s", expectedResult.ErrorResult)
			}
			expected := expectedResult.RaidStats.Parties[0].Players[0].FinalStats
			optimized := result.GetOptimizedPlayerStats().GetFinalStats()
			if protopkg.Equal(expected, optimized) {
				return
			}

			for i, expItem := range expectedGear.GetItems() {
				var optItem *proto.ItemSpec
				if i < len(optimizedGear.GetItems()) {
					optItem = optimizedGear.GetItems()[i]
				}
				if !protopkg.Equal(expItem, optItem) {
					expJSON, _ := protojson.Marshal(expItem)
					optJSON, _ := protojson.Marshal(optItem)
					t.Logf("slot %d: expected %s", i, expJSON)
					t.Logf("slot %d: got      %s", i, optJSON)
				}
			}
			if optimized == nil {
				t.Fatal("Optimize returned no optimized player stats")
			}
			for i, want := range expected.GetStats() {
				if got := statAt(optimized.GetStats(), i); got != want {
					t.Logf("stat %-24s expected=%8.2f got=%8.2f", stats.Stat(i).StatName(), want, got)
				}
			}
			for i, want := range expected.GetPseudoStats() {
				if got := statAt(optimized.GetPseudoStats(), i); got != want {
					t.Logf("stat %-24s expected=%8.4f got=%8.4f", proto.PseudoStat(i).String(), want, got)
				}
			}
			t.Fatal("optimized stats do not match expected stats")
		})
	}
}

// The gem list the UI hands the reforger: every non-meta socket colour from the embedded asset
// database, with no lower quality bound. Quality is gated inside the optimizer by the request's
// max_gem_quality (an upper bound), so filtering out Uncommon gems here would starve the solver of
// candidates the reference reforger considers and change which equally-scoring solution it settles on.
func reforgeGemOptionsFromDB() []*proto.ReforgeGemOption {
	uiDB := assetsdb.Load()
	seen := make(map[int32]struct{})
	var options []*proto.ReforgeGemOption
	for _, gem := range uiDB.GetGems() {
		if gem.GetId() == 0 {
			continue
		}
		c := gem.GetColor()
		if c == proto.GemColor_GemColorUnknown || c == proto.GemColor_GemColorMeta {
			continue
		}
		if _, dup := seen[gem.GetId()]; dup {
			continue
		}
		seen[gem.GetId()] = struct{}{}
		options = append(options, &proto.ReforgeGemOption{
			Id:                 gem.GetId(),
			Name:               gem.GetName(),
			Color:              gem.GetColor(),
			Stats:              gem.GetStats(),
			Unique:             gem.GetUnique(),
			RequiredProfession: gem.GetRequiredProfession(),
			Icon:               gem.GetIcon(),
			Phase:              gem.GetPhase(),
			Quality:            gem.GetQuality(),
		})
	}
	return options
}

func statAt(values []float64, i int) float64 {
	if i < len(values) {
		return values[i]
	}
	return 0
}

func loadReforgePreset(t *testing.T, fileName string) *proto.ReforgeOptimizeRequest {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(reforgeFixturesDir, fileName))
	if err != nil {
		t.Fatalf("failed reading preset %s: %v", fileName, err)
	}

	request := &proto.ReforgeOptimizeRequest{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(data, request); err != nil {
		t.Fatalf("failed unmarshalling fixture %s: %v", fileName, err)
	}

	// A fixture with no gem options but gem optimization on (maxGemPhase > 0) takes them from the
	// embedded database.
	if len(request.GemOptions) == 0 && request.GetSettings().GetMaxGemPhase() > 0 {
		request.GemOptions = reforgeGemOptionsFromDB()
	}

	return request
}

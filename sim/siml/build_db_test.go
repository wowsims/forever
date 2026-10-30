//go:build with_db

package siml

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"

	"github.com/wowsims/forever/sim/core/proto"
)

// Each committed build loaded over its spec's defaults, as testdata/builds holds it. The site's applyBuild is
// checked against the same files by ui/app/apply_build_files.test.ts. UPDATE_BUILDS=1
// rewrites them, and `npx oxfmt sim/siml/testdata/builds` lays them out for the format check.
func TestBuildsOverDefaults(t *testing.T) {
	builds, err := filepath.Glob(filepath.Join(langRoot, "ui/specs/*/*/builds/*.siml"))
	if err != nil || len(builds) == 0 {
		t.Fatalf("no builds under ui/specs: %v", err)
	}
	for _, path := range builds {
		folder := filepath.Dir(filepath.Dir(path))
		name := filepath.Base(filepath.Dir(folder)) + "_" + filepath.Base(folder) + "_" + strings.TrimSuffix(filepath.Base(path), ".siml") + ".json"
		t.Run(name, func(t *testing.T) {
			defaults, err := ReadFile(filepath.Join(folder, "defaults.siml"))
			if err != nil {
				t.Fatal(err)
			}
			build, err := ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			merged := googleProto.Clone(defaults.Settings).(*proto.IndividualSimSettings)
			LoadBuild(merged, build.Settings)
			got, err := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(merged)
			if err != nil {
				t.Fatal(err)
			}
			got = append(got, '\n')
			fixture := filepath.Join("testdata", "builds", name)
			if os.Getenv("UPDATE_BUILDS") == "1" {
				if err := os.WriteFile(fixture, got, 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(fixture)
			if err != nil {
				t.Fatalf("%v; UPDATE_BUILDS=1 writes it", err)
			}
			var w proto.IndividualSimSettings
			if err := protojson.Unmarshal(want, &w); err != nil {
				t.Fatal(err)
			}
			if !googleProto.Equal(&w, merged) {
				t.Errorf("%s over its defaults is not %s; UPDATE_BUILDS=1 rewrites it", path, fixture)
			}
		})
	}
}

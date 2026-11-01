//go:build with_db

package cmd

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/siml"
)

const root = "../../.."

// The rotation of a spec folder's apls/<name> as the sim's tests load it, and its print for spec, an .apl.siml.
func printedAPL(t *testing.T, folder, name string, spec proto.Spec) (*proto.APLRotation, string) {
	t.Helper()
	rotation := core.GetAplRotation(filepath.Join(root, "ui/specs", folder, "apls"), name).Rotation
	if rotation == nil {
		t.Fatalf("ui/specs/%s/apls holds no %s rotation", folder, name)
	}
	names, err := siml.DefaultNames()
	if err != nil {
		t.Fatal(err)
	}
	return rotation, printRotation(spec, rotation, names)
}

// The text of rotation as an .apl.siml for spec.
func printRotation(spec proto.Spec, rotation *proto.APLRotation, names *siml.Names) string {
	settings := &proto.IndividualSimSettings{Player: &proto.Player{Rotation: rotation}}
	return string(siml.Print(siml.Kind{Section: "apl"}, spec, settings, nil, names))
}

// The path of a file named name in a new temporary folder.
func tempPath(t *testing.T, name string) string {
	return filepath.Join(t.TempDir(), name)
}

// The character of the hunter rotation at path, from the repository.
func hunterCharacter(t *testing.T, path string) *character {
	t.Helper()
	folder, defaults, err := specDefaults(root, proto.Spec_SpecHunter)
	if err != nil {
		t.Fatal(err)
	}
	if folder != "hunter/dps" {
		t.Fatalf("folder %q, want hunter/dps", folder)
	}
	c, err := characterFromRepo(root, folder, defaults, filepath.Join(root, path))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCharacterFromRepoReadsTheSpecDefaults(t *testing.T) {
	// Two builds use the turret rotation, so it is simmed on the defaults, which the warning says.
	c := hunterCharacter(t, "ui/specs/hunter/dps/apls/turret.apl.siml")
	p := c.settings.Player
	if c.info.Source != filepath.Join(root, "ui/specs/hunter/dps/defaults.siml") || !slices.Equal(c.info.Warnings, []string{"builds mm.siml, sv.siml all use this file; simmed on the defaults"}) {
		t.Errorf("source %q, warnings %q", c.info.Source, c.info.Warnings)
	}
	if p.TalentsString != "-00530501114-550200030050220151" || p.Race != proto.Race_RaceOrc || c.info.Gear != "p1" {
		t.Errorf("got talents %q, race %v and gear %q, want the SV Weave talents, an orc and the P1 set", p.TalentsString, p.Race, c.info.Gear)
	}
	if p.Consumables.GetResourcePotId() == 0 || p.GetHunter().GetOptions().GetClassOptions().GetAmmo() == proto.HunterOptions_AmmoNone {
		t.Errorf("consumables %v, options %v", p.Consumables, p.GetHunter().GetOptions())
	}
	if c.settings.RaidBuffs.GetGiftOfTheWild() == 0 || !c.settings.Debuffs.GetSunderArmor() || !p.Buffs.GetGreaterBlessingOfKings() {
		t.Errorf("buffs not read: raid %v, debuffs %v, individual %v", c.settings.RaidBuffs, c.settings.Debuffs, p.Buffs)
	}
	// The defaults write the site's default fight, so the request adds no default target and no warning.
	if c.settings.Encounter.GetDuration() != 180 || len(c.settings.Encounter.GetTargets()) != 1 {
		t.Errorf("encounter %v, want the defaults' fight", c.settings.Encounter)
	}
}

func TestCharacterFromRepoLoadsTheBuildOfTheRotation(t *testing.T) {
	c := hunterCharacter(t, "ui/specs/hunter/dps/apls/melee.apl.siml")
	p := c.settings.Player
	if want := filepath.Join(root, "ui/specs/hunter/dps/builds/sv_melee.siml") + " over " + filepath.Join(root, "ui/specs/hunter/dps/defaults.siml"); c.info.Source != want {
		t.Errorf("source %q", c.info.Source)
	}
	if p.TalentsString != "53200005021-005003-5502000300500221" || c.info.Gear != "p1_melee" || p.DistanceFromTarget != 5 {
		t.Errorf("got talents %q, gear %q and distance %v, want the SV Melee build's", p.TalentsString, c.info.Gear, p.DistanceFromTarget)
	}
	if p.GetHunter().GetOptions().GetClassOptions().GetPetType() != proto.HunterOptions_Raptor {
		t.Errorf("options %v, want the build's raptor", p.GetHunter().GetOptions())
	}
	// The build sets no race, professions or consumables, so the defaults' stay.
	if p.Race != proto.Race_RaceOrc || p.Profession1 != proto.Profession_Engineering || p.Consumables.GetResourcePotId() == 0 {
		t.Errorf("race %v, profession %v, consumables %v, want the defaults'", p.Race, p.Profession1, p.Consumables)
	}
}

func TestSpecDefaultsSkipsAnotherSpecsBrokenDefaults(t *testing.T) {
	dir := t.TempDir()
	write := func(folder, text string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Join(dir, "ui/specs/hunter", folder), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "ui/specs/hunter", folder, "defaults.siml"), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("a", "siml 1\nspec hunter\n\nbroken:\n")
	write("b", "siml 1\nspec hunter\ntalents \"-005-55\"\n")

	folder, defaults, err := specDefaults(dir, proto.Spec_SpecHunter)
	if err != nil || folder != "hunter/b" || defaults.Settings.GetPlayer().GetTalentsString() != "-005-55" {
		t.Errorf("got folder %q (%v), want hunter/b past the broken hunter/a", folder, err)
	}

	write("b", "siml 1\nspec hunter\n\nbroken:\n")
	_, _, err = specDefaults(dir, proto.Spec_SpecHunter)
	if err == nil || !strings.Contains(err.Error(), "holds no defaults.siml for SpecHunter") ||
		!strings.Contains(err.Error(), filepath.Join("hunter", "a", "defaults.siml")) || !strings.Contains(err.Error(), filepath.Join("hunter", "b", "defaults.siml")) {
		t.Errorf("error %v, want it to name both broken files", err)
	}
}

func TestLoadBuildKeepsTheRotationOverAnAutoOne(t *testing.T) {
	apl := &proto.APLRotation{Type: proto.APLRotation_TypeAPL, PriorityList: []*proto.APLListItem{{}}}
	s := &proto.IndividualSimSettings{Player: &proto.Player{Rotation: apl}}
	siml.LoadBuild(s, &proto.IndividualSimSettings{Player: &proto.Player{Rotation: &proto.APLRotation{Type: proto.APLRotation_TypeAuto}}})
	if s.Player.Rotation != apl {
		t.Errorf("rotation %v, want the defaults' APL kept under an Auto build", s.Player.Rotation)
	}

	simple := &proto.APLRotation{Type: proto.APLRotation_TypeSimple}
	siml.LoadBuild(s, &proto.IndividualSimSettings{Player: &proto.Player{Rotation: simple}})
	if s.Player.Rotation != simple {
		t.Errorf("rotation %v, want the build's own", s.Player.Rotation)
	}
}

func TestSimlSimCountsEveryRowAndKeepsPathsAcrossHiddenRows(t *testing.T) {
	rotation, text := printedAPL(t, "hunter/dps", "default", proto.Spec_SpecHunter)
	path := tempPath(t, "default.apl.siml")
	res := runSimlSim(context.Background(), root, simlSimRequest{SIML: text, Path: path, Iterations: 20, Timeline: true, TimelineLimit: 25})
	if res.Error != "" {
		t.Fatal(res.Error)
	}
	if want := len(listRows(rotation)); len(res.Rows) != want {
		t.Errorf("%d rows, want %d", len(res.Rows), want)
	}
	for i := range rotation.PrepullActions {
		if got := res.Rows["prepullActions/"+strconv.Itoa(i)].Executions; got != 1 {
			t.Errorf("prepullActions/%d ran %v times per iteration, want 1", i, got)
		}
	}
	if len(res.Timeline) != 25 || res.Timeline[0].Path != "prepullActions/0" {
		t.Errorf("timeline of %d events starting %v, want 25 starting with the first prepull row", len(res.Timeline), res.Timeline[0])
	}
	if res.DPS.Mean <= 0 || len(res.Spells) == 0 || res.Iterations != 20 {
		t.Errorf("dps %v, %d spells, %d iterations", res.DPS, len(res.Spells), res.Iterations)
	}

	// Hiding the first row that runs leaves every later row under its own index.
	first := -1
	for i := range rotation.PriorityList {
		if res.Rows["priorityList/"+strconv.Itoa(i)].Executions > 0 {
			first = i
			break
		}
	}
	rotation.PriorityList[first].Hide = true
	names, _ := siml.DefaultNames()
	hidden := runSimlSim(context.Background(), root, simlSimRequest{SIML: printRotation(proto.Spec_SpecHunter, rotation, names), Path: path, Iterations: 20})
	if row := hidden.Rows["priorityList/"+strconv.Itoa(first)]; !row.Hidden || row.Executions != 0 {
		t.Errorf("hidden row %d: %+v", first, row)
	}
	for path, row := range hidden.Rows {
		if strings.HasPrefix(path, "prepullActions/") && row.Executions != 1 {
			t.Errorf("%s ran %v times with priorityList/%d hidden", path, row.Executions, first)
		}
	}
}

func TestSIMLSimsWhatTheRotationItWasPrintedFromSims(t *testing.T) {
	rotation, text := printedAPL(t, "hunter/dps", "default", proto.Spec_SpecHunter)
	req := simlSimRequest{SIML: text, Path: tempPath(t, "default.apl.siml"), Iterations: 50, Seed: 3}
	fromSIML := runSimlSim(context.Background(), root, req)
	if fromSIML.Error != "" {
		t.Fatal(fromSIML.Error)
	}
	char := hunterCharacter(t, "ui/specs/hunter/dps/apls/default.apl.siml")
	base, _ := individualSimRequest(char.settings, &proto.SimOptions{RandomSeed: req.Seed})
	direct, err := simAPL(context.Background(), base, "hunter/dps", rotation, req)
	if err != nil {
		t.Fatal(err)
	}
	if fromSIML.DPS != direct.DPS || !reflect.DeepEqual(fromSIML.Rows, direct.Rows) {
		t.Errorf("the .apl.siml sims %v, the rotation it was printed from %v", fromSIML.DPS, direct.DPS)
	}
}

func TestSimlSimComparesOnTheSameSeed(t *testing.T) {
	_, text := printedAPL(t, "hunter/dps", "default", proto.Spec_SpecHunter)
	path := tempPath(t, "default.apl.siml")
	res := runSimlSim(context.Background(), root, simlSimRequest{SIML: text, Path: path, Compare: &compared{SIML: text}, Iterations: 10, Seed: 7})
	if res.Error != "" || res.Compare == nil {
		t.Fatalf("error %q, compare %v", res.Error, res.Compare)
	}
	if res.DPS != res.Compare.DPS {
		t.Errorf("the same APL twice: %v and %v", res.DPS, res.Compare.DPS)
	}
	_, warrior := printedAPL(t, "warrior/dps", "arms", proto.Spec_SpecDpsWarrior)
	if res := runSimlSim(context.Background(), root, simlSimRequest{SIML: text, Path: path, Compare: &compared{SIML: warrior}, Iterations: 2}); !strings.Contains(res.Error, "compare: its spec is SpecDpsWarrior") {
		t.Errorf("a warrior rotation to compare with a hunter one: %q", res.Error)
	}
}

func TestUseLinesAreReadAgainstThePath(t *testing.T) {
	_, text := printedAPL(t, "hunter/dps", "default", proto.Spec_SpecHunter)
	path := tempPath(t, "default.apl.siml")
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	using := "siml 1\nspec hunter\nuse \"default.apl.siml\" with $use_multi_shot = false\n"
	direct := runSimlSim(context.Background(), root, simlSimRequest{SIML: text, Path: path, Iterations: 10})
	used := runSimlSim(context.Background(), root, simlSimRequest{SIML: using, Path: filepath.Join(filepath.Dir(path), "no_multi.apl.siml"), Iterations: 10})
	if used.Error != "" || direct.Error != "" {
		t.Fatalf("errors %q and %q", used.Error, direct.Error)
	}
	if len(used.Rows) != len(direct.Rows) || used.DPS == direct.DPS {
		t.Errorf("with Multi-Shot off: %d rows and %v, the file it uses: %d rows and %v", len(used.Rows), used.DPS, len(direct.Rows), direct.DPS)
	}
}

func TestSimlSimFindsTheBuildOfARotationFromAnyFolder(t *testing.T) {
	// The test runs in cmd/wowsimcli/cmd, so the repository root is root, and the path names the file absolute and
	// against the root.
	const rotation = "ui/specs/hunter/dps/apls/melee.apl.siml"
	abs, err := filepath.Abs(filepath.Join(root, rotation))
	if err != nil {
		t.Fatal(err)
	}
	text, err := os.ReadFile(abs)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, "ui/specs/hunter/dps/builds/sv_melee.siml") + " over " + filepath.Join(root, "ui/specs/hunter/dps/defaults.siml")
	for _, path := range []string{abs, rotation} {
		res := runSimlSim(context.Background(), repositoryRootOf(path), simlSimRequest{SIML: string(text), Path: path, Iterations: 2})
		if res.Error != "" || res.Character == nil || res.Character.Source != want {
			t.Errorf("%s: error %q, character %+v, want it simmed on %s", path, res.Error, res.Character, want)
		}
	}
}

func TestSettingsFileSuppliesTheCharacter(t *testing.T) {
	c := hunterCharacter(t, "ui/specs/hunter/dps/apls/default.apl.siml")
	data, _ := protojson.Marshal(c.settings)
	path := tempPath(t, "settings.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	_, text := printedAPL(t, "hunter/dps", "default", proto.Spec_SpecHunter)
	apl := tempPath(t, "default.apl.siml")
	fromRepo := runSimlSim(context.Background(), root, simlSimRequest{SIML: text, Path: apl, Iterations: 10})
	fromFile := runSimlSim(context.Background(), root, simlSimRequest{SIML: text, Path: apl, Iterations: 10, SettingsFile: path})
	if fromFile.Error != "" || fromFile.DPS != fromRepo.DPS || fromFile.Character.Source != path {
		t.Errorf("from the file: %v (%s), from the repo: %v", fromFile.DPS, fromFile.Error, fromRepo.DPS)
	}
	// defaults.siml is a full setup with its own encounter, so it sims as the repo does, with no warning.
	defaults := filepath.Join(root, "ui/specs/hunter/dps/defaults.siml")
	fromSIML := runSimlSim(context.Background(), root, simlSimRequest{SIML: text, Path: apl, Iterations: 10, SettingsFile: defaults})
	if fromSIML.Error != "" || fromSIML.DPS != fromRepo.DPS || len(fromSIML.Character.Warnings) != 0 {
		t.Errorf("from defaults.siml: %v (%s, warnings %q), from the repo: %v", fromSIML.DPS, fromSIML.Error, fromSIML.Character.Warnings, fromRepo.DPS)
	}
	// A preset build is loaded over its spec's defaults, as with no settings file.
	build := filepath.Join(root, "ui/specs/hunter/dps/builds/sv.siml")
	fromBuild := runSimlSim(context.Background(), root, simlSimRequest{SIML: text, Path: apl, Iterations: 10, SettingsFile: build})
	if want := build + " over " + defaults; fromBuild.Error != "" || fromBuild.Character.Source != want || fromBuild.Character.Race != "RaceOrc" ||
		fromBuild.Character.Talents != "1-00531500114-52225303000021005" || !slices.Contains(fromBuild.Character.Warnings, build+" is a preset build: simmed over "+defaults) {
		t.Errorf("from a preset build: %s, character %+v", fromBuild.Error, fromBuild.Character)
	}
	_, warrior := printedAPL(t, "warrior/dps", "arms", proto.Spec_SpecDpsWarrior)
	// The warrior's presets are JSON, so the spec has no defaults.siml and its field names it.
	if res := runSimlSim(context.Background(), root, simlSimRequest{SIML: warrior, Path: tempPath(t, "arms.apl.siml"), SettingsFile: path}); !strings.Contains(res.Error, "not a dps_warrior character") {
		t.Errorf("a hunter file for a warrior rotation: %q", res.Error)
	}
}

func TestSimlSimReportsWhatCannotBeSimmed(t *testing.T) {
	path := tempPath(t, "default.apl.siml")
	for text, want := range map[string]string{
		"siml 1\nspec nope\n":                             path + ":2:6 unknown-spec",
		"siml 1\nspec hunter\n\npriority:\n  cast nope\n": path + ":5:8 unknown-name no spell is named nope",
		"siml 1\nspec hunter\n":                           "",
	} {
		res := runSimlSim(context.Background(), root, simlSimRequest{SIML: text, Path: path, Iterations: 2})
		if want == "" && res.Error != "" || want != "" && !strings.Contains(res.Error, want) {
			t.Errorf("%q: error %q, want %q", text, res.Error, want)
		}
	}
	if res := runSimlSim(context.Background(), root, simlSimRequest{SIML: "siml 1\nspec hunter\n", Path: tempPath(t, "default.gear.siml")}); !strings.Contains(res.Error, "is not an .apl.siml") {
		t.Errorf("a gear file: %q", res.Error)
	}
	if res := runSimlSim(context.Background(), root, simlSimRequest{SIML: "siml 1\nspec hunter\n"}); !strings.Contains(res.Error, "needs the siml text of an .apl.siml and its path") {
		t.Errorf("no path: %q", res.Error)
	}
}

func TestSimlSimRefusesOtherFields(t *testing.T) {
	req, err := readRequest(strings.NewReader(`{"siml": "siml 1", "path": "x.apl.siml", "compare": {"siml": "y"}, "iterations": 5}`))
	if err != nil || req.SIML != "siml 1" || req.Path != "x.apl.siml" || req.Compare.SIML != "y" || req.Iterations != 5 {
		t.Errorf("got %+v, %v", req, err)
	}
	if _, err := readRequest(strings.NewReader(`{"spec": "hunter/dps", "apl": {}}`)); err == nil || !strings.Contains(err.Error(), `unknown field "spec"`) {
		t.Errorf("a request with spec and apl: %v", err)
	}
}

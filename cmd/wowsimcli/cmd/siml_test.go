//go:build with_db

package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/siml"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	sim.RegisterAll()
}

const corpus = "../../../sim/siml/testdata"

// Runs a siml subcommand, and gives what it wrote to standard output and standard error.
func runSiml(run func(out, errOut io.Writer, args []string) error, args ...string) (string, string, error) {
	var out, errOut bytes.Buffer
	err := run(&out, &errOut, args)
	return out.String(), errOut.String(), err
}

// Writes files, by path relative to a new temporary folder, and gives the folder.
func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, text := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// A folder that reads as a repository, holding files: the name table marks its root, and its database files link to
// this repository's, which the checks of files under ui/specs read.
func repositoryFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := writeFiles(t, files)
	names, err := filepath.Abs("../../../sim/siml/names.json")
	if err != nil {
		t.Fatal(err)
	}
	links := map[string]string{filepath.Join("sim", "siml", "names.json"): names}
	for _, file := range siml.CatalogFiles {
		links[file] = filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(names))), file)
	}
	for link, target := range links {
		if err := os.MkdirAll(filepath.Join(dir, filepath.Dir(link)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, filepath.Join(dir, link)); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestPrintReadsBackAsItsJSON(t *testing.T) {
	for _, c := range []struct{ json, kind, spec, siml string }{
		{"parse/hunter_default.apl.json", "apl", "hunter", "default.apl.siml"},
		{"parse/hunter_p1.gear.json", "gear", "", "p1.gear.siml"},
		{"parse/hunter_defaults.json", "siml", "", "defaults.siml"},
		{"parse/hunter_defaults.json", "siml", "hunter", "defaults.siml"},
	} {
		t.Run(c.json+"/"+c.spec, func(t *testing.T) {
			args := []string{"-kind", c.kind}
			if c.spec != "" {
				args = append(args, "-spec", c.spec)
			}
			text, notes, err := runSiml(simlPrint, append(args, filepath.Join(corpus, c.json))...)
			if err != nil {
				t.Fatal(err)
			}
			if notes != "" {
				t.Errorf("print wrote %q to standard error", notes)
			}
			path := filepath.Join(t.TempDir(), c.siml)
			if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
				t.Fatal(err)
			}
			if out, _, err := runSiml(simlVerify, filepath.Join(corpus, c.json), path); err != nil || !strings.Contains(out, "reads back as") {
				t.Errorf("verify: %v\n%s\n%s", err, out, text)
			}
		})
	}
}

func TestPrintWritesTheCategoriesGiven(t *testing.T) {
	text, _, err := runSiml(simlPrint, "-kind", "siml", "-categories", "gr", filepath.Join(corpus, "parse/hunter_defaults.json"))
	if err != nil {
		t.Fatal(err)
	}
	names, _ := siml.DefaultNames()
	f, err := siml.Parse("defaults.siml", []byte(text), names, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(f.Categories, []string{"g", "r"}) {
		t.Errorf("the print holds the categories %v, want g and r\n%s", f.Categories, text)
	}
}

func TestPrintSaysWhatItCannotPrint(t *testing.T) {
	rotation := filepath.Join(corpus, "parse/hunter_default.apl.json")
	settings := filepath.Join(corpus, "parse/hunter_defaults.json")
	for want, args := range map[string][]string{
		"-kind is apl, gear or siml":      {rotation},
		"-kind apl needs -spec":           {"-kind", "apl", rotation},
		`no spec is named "hunters"`:      {"-kind", "apl", "-spec", "hunters", rotation},
		"-categories is for -kind siml":   {"-kind", "apl", "-spec", "hunter", "-categories", "r", rotation},
		"-categories takes each of":       {"-kind", "siml", "-categories", "gg", settings},
		"-categories needs":               {"-kind", "siml", "-categories=", settings},
		"does not read as EquipmentSpec":  {"-kind", "gear", rotation},
		"holds a hunter player, not mage": {"-kind", "siml", "-spec", "mage", settings},
		"wrong number of arguments":       {"-kind", "gear"},
		"no such file or directory":       {"-kind", "gear", "missing.gear.json"},
	} {
		if _, _, err := runSiml(simlPrint, args...); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%v: error %v, want %q", args, err, want)
		}
	}
}

func TestPrintNamesTheRawRefsItWrites(t *testing.T) {
	dir := writeFiles(t, map[string]string{"x.apl.json": `{"type": "TypeAPL", "priorityList": [{"action": {"castSpell": {"spellId": {"spellId": 1}}}}]}`})
	text, notes, err := runSiml(simlPrint, "-kind", "apl", "-spec", "hunter", filepath.Join(dir, "x.apl.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "cast `{\"spellId\":1}`") || !strings.Contains(notes, "x.apl.json: the print writes raw JSON at 5:8") {
		t.Errorf("print:\n%s\nstandard error: %s", text, notes)
	}
}

func TestVerifyNamesWhereTheFilesDiffer(t *testing.T) {
	original := filepath.Join(corpus, "parse/hunter_default.apl.json")
	text, _, err := runSiml(simlPrint, "-kind", "apl", "-spec", "hunter", original)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	rotation := &proto.APLRotation{}
	if err := protojson.Unmarshal(data, rotation); err != nil {
		t.Fatal(err)
	}
	rotation.PriorityList[1].Hide = !rotation.PriorityList[1].Hide
	rotation.PrepullActions = rotation.PrepullActions[1:]
	changed, _ := protojson.Marshal(rotation)
	dir := writeFiles(t, map[string]string{"x.apl.json": string(changed), "x.apl.siml": text})

	out, _, err := runSiml(simlVerify, filepath.Join(dir, "x.apl.json"), filepath.Join(dir, "x.apl.siml"))
	if err == nil {
		t.Fatal("verify passes a file that differs")
	}
	for _, want := range []string{"does not read back as", "  prepullActions[0].", "  priorityList[1].hide: "} {
		if !strings.Contains(out, want) {
			t.Errorf("verify does not say %q:\n%s", want, out)
		}
	}
}

func TestVerifyListsTheDiagnosticsOfAFileThatDoesNotLoad(t *testing.T) {
	dir := writeFiles(t, map[string]string{"x.apl.json": `{}`, "x.apl.siml": "siml 1\nspec hunter\n\npriority:\n  cast nope\n"})
	_, notes, err := runSiml(simlVerify, filepath.Join(dir, "x.apl.json"), filepath.Join(dir, "x.apl.siml"))
	if err == nil || !strings.Contains(err.Error(), "does not load") || !strings.Contains(notes, "x.apl.siml:5:8 unknown-name no spell is named nope") {
		t.Errorf("error %v, standard error %q", err, notes)
	}
}

func TestParse(t *testing.T) {
	dir := repositoryFiles(t, map[string]string{
		"x.apl.siml":      "siml 1\nspec hunter\n\npriority:\n  cast multi_shot if $nope\n",
		"broken.siml":     "siml 1\nspec hunter\n\nnope:\n",
		"ui/specs/x.siml": "siml 1\nspec hunter\n\npriority:\n  cast multi_shot if $nope\n",
	})
	out, notes, err := runSiml(simlParse, filepath.Join(dir, "x.apl.siml"))
	if err != nil || !strings.HasSuffix(out, "x.apl.siml: loads as a rotation for hunter\n") || !strings.Contains(notes, "undefined-variable no variable here is named $nope (warning)") {
		t.Errorf("error %v, output %q, standard error %q", err, out, notes)
	}
	out, _, err = runSiml(simlParse, "-json", filepath.Join(dir, "x.apl.siml"))
	var rotation map[string]any
	if err != nil || json.Unmarshal([]byte(out), &rotation) != nil || rotation["priorityList"] == nil {
		t.Errorf("-json: error %v, output %q", err, out)
	}
	if _, notes, err := runSiml(simlParse, filepath.Join(dir, "broken.siml")); err == nil || !strings.Contains(notes, "broken.siml:4:1 unknown-section") {
		t.Errorf("a file with an error: %v, standard error %q", err, notes)
	}
	if _, _, err := runSiml(simlParse, filepath.Join(dir, "ui/specs/x.siml")); err == nil || !strings.Contains(err.Error(), "fails the check of the files under ui/specs") {
		t.Errorf("a warning under ui/specs: %v", err)
	}
}

func TestFmt(t *testing.T) {
	canonical := "siml 1\nspec hunter\n\npriority:\n  cast multi_shot\n"
	warning := "siml 1\nspec hunter\n\npriority:\n  cast multi_shot if $nope\n"
	dir := repositoryFiles(t, map[string]string{
		"a.apl.siml":                         canonical,
		"b.apl.siml":                         "siml 1\nspec hunter\npriority:\n    cast   multi_shot\n",
		"c.apl.siml":                         "siml 1\nspec hunter\n\npriority:\n  cast nope\n",
		"d.apl.siml":                         warning,
		"ui/specs/e.apl.siml":                warning,
		"ui/specs/f.apl.siml":                "siml 1\nspec hunter\n\npriority:\n  cast `{\"spellId\":1}`\n",
		"node_modules/g.apl.siml":            "not read",
		".hidden/h.apl.siml":                 "not read",
		"ui/specs/README.md":                 "not read",
		"ui/specs/hunter/canonical.apl.siml": canonical,
	})
	at := func(name string) string { return filepath.Join(dir, name) }
	if err := os.Mkdir(at("empty"), 0o755); err != nil {
		t.Fatal(err)
	}

	out, notes, err := runSiml(simlFmt, "-check", dir)
	if err == nil || err.Error() != "4 of 7 .siml files are not in canonical form or fail the check" {
		t.Errorf("-check: %v", err)
	}
	if out != at("b.apl.siml")+": not in canonical form\n" {
		t.Errorf("-check lists %q, want only b.apl.siml", out)
	}
	for _, want := range []string{
		at("c.apl.siml") + ":5:8 unknown-name",
		at("ui/specs/e.apl.siml") + ":5:22 undefined-variable no variable here is named $nope (warning)",
		at("ui/specs/f.apl.siml") + ":5:8 raw-ref",
	} {
		if !strings.Contains(notes, want) {
			t.Errorf("-check does not say %q:\n%s", want, notes)
		}
	}
	if strings.Contains(notes, "d.apl.siml") {
		t.Errorf("-check fails a warning outside ui/specs:\n%s", notes)
	}

	if out, _, err := runSiml(simlFmt, at("b.apl.siml"), at("a.apl.siml")); err != nil || out != canonical+canonical {
		t.Errorf("fmt prints %q (%v), want the canonical form of both", out, err)
	}
	if _, _, err := runSiml(simlFmt, "-w", at("b.apl.siml"), at("c.apl.siml")); err == nil || err.Error() != "1 of 2 .siml files fail the check" {
		t.Errorf("-w: %v", err)
	}
	if text, _ := os.ReadFile(at("b.apl.siml")); string(text) != canonical {
		t.Errorf("-w writes %q", text)
	}
	if _, _, err := runSiml(simlFmt, "-check", at("a.apl.siml"), at("b.apl.siml"), at("ui/specs/hunter/canonical.apl.siml")); err != nil {
		t.Errorf("-check on canonical files: %v", err)
	}

	for want, args := range map[string][]string{
		"-check and -w do not go together": {"-check", "-w", dir},
		"is not a .siml file":              {at("ui/specs/README.md")},
		"no .siml file in":                 {at("empty")},
		"wrong number of arguments":        {"-check"},
	} {
		if _, _, err := runSiml(simlFmt, args...); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%v: error %v, want %q", args, err, want)
		}
	}
}

func TestSimReadsAFullSetup(t *testing.T) {
	request, warnings, err := readSimRequest(filepath.Join(corpus, "parse/hunter_defaults.siml"))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(warnings, []string{
		"the rotation is TypeSimple, which the site builds into an APL before simming; the file holds no APL row, so this sims no rotation",
		"the settings hold no encounter; simmed against one default target",
	}) {
		t.Errorf("warnings %q", warnings)
	}
	player := request.Raid.Parties[0].Players[0]
	if request.SimOptions.Iterations != 25000 || player.GetHunter() == nil || player.Race != proto.Race_RaceOrc || len(request.Encounter.Targets) != 1 {
		t.Errorf("iterations %d, player %v", request.SimOptions.Iterations, player)
	}
	if !request.Raid.Debuffs.SunderArmor || !request.Raid.Buffs.GiftOfTheWild || request.Raid.Buffs.BattleShout == proto.TristateEffect_TristateEffectMissing {
		t.Errorf("buffs not read: raid %v, party %v, debuffs %v", request.Raid.Buffs, request.Raid.Parties[0].Buffs, request.Raid.Debuffs)
	}
	request.SimOptions.Iterations = 20
	if res := core.RunRaidSim(request); res.Error != nil || res.RaidMetrics.Dps.Avg <= 0 {
		t.Errorf("the sim gives %v (%v)", res.GetRaidMetrics().GetDps(), res.Error)
	}

	const orc = "siml 1\nspec hunter\n\nsettings: race orc\n\ngear:\n\nencounter:\n  duration 60\n  targets(level=63)\n"
	dir := writeFiles(t, map[string]string{
		"request.json":      `{"raid": {}, "simOptions": {"iterations": 3}}`,
		"options.json":      `{"raid": {}}`,
		"rotation.apl.siml": "siml 1\nspec hunter\n",
		"broken.siml":       "siml 1\nspec hunter\n\nnope:\n",
		"build.siml":        "siml 1\nspec hunter\ntalents \"-0053\"\n\nsettings: distance_from_target 11\n",
		"none.siml":         strings.Replace(orc, "hunter", "dps_warrior", 1),
		"auto.siml":         orc + "\napl: type auto\n",
		"simple_rows.siml":  orc + "\nsimple: use_multi\n\npriority:\n  cast arcane_shot\n",
		"apl.siml":          orc + "\npriority:\n  cast arcane_shot\n",
		"dummies.siml":      orc + "\npriority:\n  cast arcane_shot\n\nraid_buffs: target_dummies 9\n",
	})
	if request, _, err := readSimRequest(filepath.Join(dir, "request.json")); err != nil || request.SimOptions.Iterations != 3 {
		t.Errorf("a RaidSimRequest: %v, %v", request, err)
	}
	for file, want := range map[string]string{
		"none.siml":        "the settings hold no rotation; simmed with none",
		"auto.siml":        "the rotation is TypeAuto, which the site builds into an APL before simming; the file holds no APL row, so this sims no rotation",
		"simple_rows.siml": "the rotation is TypeSimple, which the site builds into an APL before simming; this sims the APL rows the file holds instead",
		"apl.siml":         "",
		"dummies.siml":     "",
	} {
		request, warnings, err := readSimRequest(filepath.Join(dir, file))
		if err != nil {
			t.Fatal(err)
		}
		if want != "" && !slices.Equal(warnings, []string{want}) || want == "" && len(warnings) != 0 {
			t.Errorf("%s: warnings %q, want %q", file, warnings, want)
		}
		request.SimOptions.Iterations = 5
		if res := core.RunRaidSim(request); res.Error != nil {
			t.Errorf("%s: the sim fails: %v", file, res.Error)
		}
	}
	for file, want := range map[string]string{
		"options.json":      "a RaidSimRequest needs simOptions",
		"rotation.apl.siml": "holds only its apl section, and a setup is an x.siml",
		"broken.siml":       "broken.siml:4:1 unknown-section",
		// The test runs where ui/specs holds no defaults.siml to load a build over.
		"build.siml": "build.siml is a preset build, and ui/specs holds no defaults.siml for SpecHunter",
	} {
		if _, _, err := readSimRequest(filepath.Join(dir, file)); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: error %v, want %q", file, err, want)
		}
	}
}

func TestSimLoadsAPresetBuildOverItsSpecDefaults(t *testing.T) {
	t.Chdir(root)
	const build, defaults = "ui/specs/hunter/dps/builds/sv.siml", "ui/specs/hunter/dps/defaults.siml"
	request, warnings, err := readSimRequest(build)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(warnings, build+" is a preset build: simmed over "+defaults) {
		t.Errorf("warnings %q", warnings)
	}
	player := request.Raid.Parties[0].Players[0]
	if player.Race != proto.Race_RaceOrc || player.TalentsString != "1-00531500114-52225303000021005" || player.DistanceFromTarget != 12 ||
		player.Consumables.GetResourcePotId() == 0 || len(player.Rotation.GetPriorityList()) == 0 {
		t.Errorf("race %v, talents %q, distance %v, consumables %v, %d priority rows, want the defaults' race and consumables under the build's talents, distance and rotation",
			player.Race, player.TalentsString, player.DistanceFromTarget, player.Consumables, len(player.Rotation.GetPriorityList()))
	}
	request.SimOptions.Iterations = 20
	if res := core.RunRaidSim(request); res.Error != nil || res.RaidMetrics.Dps.Avg <= 0 {
		t.Errorf("the sim gives %v (%v)", res.GetRaidMetrics().GetDps(), res.Error)
	}
}

func TestSettingsRequestSimsAsTheSiteDoes(t *testing.T) {
	ring := func() *proto.EquipmentSpec {
		items := make([]*proto.ItemSpec, 17)
		for i := range items {
			items[i] = &proto.ItemSpec{}
		}
		items[proto.ItemSlot_ItemSlotFinger1] = &proto.ItemSpec{Id: 1, Enchant: 2}
		items[proto.ItemSlot_ItemSlotHead] = &proto.ItemSpec{Id: 3, Enchant: 4}
		return &proto.EquipmentSpec{Items: items}
	}
	settings := &proto.IndividualSimSettings{
		Player:    &proto.Player{Equipment: ring()},
		Encounter: &proto.Encounter{Duration: 60},
		Settings:  &proto.SimSettings{Iterations: 7, FixedRngSeed: 9},
	}
	request, warnings := settingsRequest(settings)
	items := request.Raid.Parties[0].Players[0].Equipment.Items
	if items[proto.ItemSlot_ItemSlotFinger1].Enchant != 0 || items[proto.ItemSlot_ItemSlotHead].Enchant != 4 {
		t.Errorf("without Enchanting: ring enchant %d, head enchant %d", items[proto.ItemSlot_ItemSlotFinger1].Enchant, items[proto.ItemSlot_ItemSlotHead].Enchant)
	}
	if settings.Player.Equipment.Items[proto.ItemSlot_ItemSlotFinger1].Enchant != 2 {
		t.Error("the settings' own ring enchant is cleared")
	}
	if request.Encounter.Duration != 60 || len(request.Encounter.Targets) != 1 || len(warnings) != 1 || settings.Encounter.Targets != nil {
		t.Errorf("an encounter with no target: %v, warnings %q", request.Encounter, warnings)
	}
	if request.SimOptions.Iterations != 7 || request.SimOptions.RandomSeed != 9 || request.Type != proto.SimType_SimTypeIndividual {
		t.Errorf("options %v, type %v", request.SimOptions, request.Type)
	}
	if request.Raid.Buffs == nil || request.Raid.Debuffs == nil || request.Raid.Parties[0].Buffs == nil {
		t.Errorf("absent buffs are not empty messages: %v", request.Raid)
	}

	settings.Player.Profession2 = proto.Profession_Enchanting
	settings.Settings = nil
	request, _ = settingsRequest(settings)
	if request.Raid.Parties[0].Players[0].Equipment.Items[proto.ItemSlot_ItemSlotFinger1].Enchant != 2 {
		t.Error("an enchanter loses the ring enchant")
	}
	if request.SimOptions.Iterations != defaultIterations || request.SimOptions.RandomSeed == 0 {
		t.Errorf("with no sim settings: %v", request.SimOptions)
	}
}

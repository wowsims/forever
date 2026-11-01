//go:build with_db

package siml

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

// The language services on the committed name table and the item database.

func langNames(t *testing.T) *Names {
	t.Helper()
	names, err := DefaultNames()
	if err != nil {
		t.Fatal(err)
	}
	return names
}

// The spell id a hunter file's spell name reads as, from the name table.
func tableSpellID(t *testing.T, name string) int32 {
	t.Helper()
	id, ok := langNames(t).Spell(proto.Class_ClassHunter, NamespaceSpell, name)
	if !ok {
		t.Fatalf("the name table has no hunter spell %s", name)
	}
	return id.GetSpellId()
}

func langCatalog() *DatabaseCatalog { return CatalogAt(langRoot) }

func analyzeText(t *testing.T, path, text string) *Doc {
	t.Helper()
	return Analyze(path, []byte(text), langNames(t), langOpen)
}

func complete(t *testing.T, source string, path ...string) []string {
	t.Helper()
	at := hunterPath
	if len(path) > 0 {
		at = path[0]
	}
	return labelsOf(completionsIn(t, langNames(t), langCatalog(), at, source))
}

func hasAll(t *testing.T, what string, labels []string, want ...string) {
	t.Helper()
	for _, w := range want {
		if !slices.Contains(labels, w) {
			t.Errorf("%s: %q is not offered; offered %q", what, w, labels)
		}
	}
}

func hasNone(t *testing.T, what string, labels []string, unwanted ...string) {
	t.Helper()
	for _, w := range unwanted {
		if slices.Contains(labels, w) {
			t.Errorf("%s: %q is offered", what, w)
		}
	}
}

func TestLangWorkspace(t *testing.T) {
	if _, ok := langNames(t).spellEntry(proto.Class_ClassHunter, NamespaceSpell, "hunters_mark"); !ok {
		t.Error("the table has no hunter spell hunters_mark")
	}
}

func TestCompletePrepull(t *testing.T) {
	start := complete(t, langPreface+"prepull:\n  |\n")
	hasAll(t, "row start", start, "-1s", "${Melee weave}")
	hasNone(t, "row start", start, "cast")
	hasAll(t, "after the time", complete(t, langPreface+"prepull:\n  -1s |\n"), "cast", "group")
	hasAll(t, "after cast", complete(t, langPreface+"prepull:\n  -1s cast |\n"), "hunters_mark")
}

func TestCompleteRowsAndConditions(t *testing.T) {
	hasAll(t, "row", complete(t, langPreface+"priority:\n  |\n"), "cast", "channel", "wait until", "move for", "sequence", "autocast_other_cooldowns")
	hasAll(t, "condition", complete(t, langPreface+"priority:\n  cast hunters_mark if |\n"), "aura", "dot", "current_time", "target.", "${Melee weave}", "not")
	if got := complete(t, langPreface+"priority:\n  cast hunters_mark |\n"); !slices.Equal(got, []string{"on", "if"}) {
		t.Errorf("after the spell: %q", got)
	}
}

func TestCompleteAccessors(t *testing.T) {
	hasAll(t, "props", complete(t, langPreface+"priority:\n  cast x if target.aura(hunters_mark).|\n"), "is_active", "remaining_time", "num_stacks", "should_refresh")
	if got := complete(t, langPreface+"priority:\n  cast x if target.aura(hunters_mark).num_stacks(|\n"); !slices.Equal(got, []string{"include_reaction_time", "include_reaction_time="}) {
		t.Errorf("args: %q", got)
	}
	hasAll(t, "unprefixed args", complete(t, langPreface+"priority:\n  cast x if aura(hunters_mark).should_refresh(|\n"), "source_unit=", "max_overlap=")
}

func TestCompleteAuraNamespace(t *testing.T) {
	target := complete(t, langPreface+"priority:\n  cast x if target.aura(|\n")
	hasAll(t, "target", target, "hunters_mark")
	hasNone(t, "target", target, "aspect_of_the_hawk")
	player := complete(t, langPreface+"priority:\n  cast x if aura(|\n")
	hasAll(t, "player", player, "aspect_of_the_hawk")
	hasNone(t, "player", player, "hunters_mark")
	argument := complete(t, langPreface+"priority:\n  cast x if aura(|).should_refresh(source_unit=target)\n")
	hasAll(t, "unit argument", argument, "hunters_mark")
	hasNone(t, "unit argument", argument, "aspect_of_the_hawk")
	hasAll(t, "pet", complete(t, langPreface+"priority:\n  cast x if pet1.aura(|\n"), "frenzy")
}

func TestCompleteNamesAndDefinitions(t *testing.T) {
	hasAll(t, "item.", complete(t, langPreface+"priority:\n  cast item.|\n"), "goblin_sapper_charge")
	hasAll(t, "other.", complete(t, langPreface+"priority:\n  cast other.|\n"), "offensive_potion")
	hasAll(t, "on", complete(t, langPreface+"priority:\n  cast hunters_mark on |\n"), "previous_target")
	hasAll(t, "enum", complete(t, langPreface+"priority:\n  cast x if is_execute_phase(|\n"), "e20")
	text, at := marked(langPreface + "priority:\n  group \"|\"\n\ngroup \"Weave\":\n  none\n")
	groups := analyzeText(t, hunterPath, text).Complete(at, langCatalog())
	if got := labelsOf(groups); !slices.Equal(got, []string{`"Weave"`}) {
		t.Fatalf("groups: %q", got)
	}
	if got, want := groups[0].Replace, (Range{Position{6, 9}, Position{6, 11}}); got != want {
		t.Errorf("the group replaces %v, want the whole string %v", got, want)
	}
	if got := complete(t, langPreface+"priority:\n  reset_sequence(|)\n  sequence \"HHA\" [none]\n"); !slices.Equal(got, []string{"sequence_name=", `"HHA"`}) {
		t.Errorf("sequences: %q", got)
	}
	if got := complete(t, langPreface+"priority:\n  cast x if $|\n"); !slices.Equal(got, []string{"${Melee weave}"}) {
		t.Errorf("variables: %q", got)
	}
	if got := complete(t, langPreface+"priority:\n  group \"G\" with %|\n\ngroup \"G\":\n  cast x if %ready\n"); !slices.Equal(got, []string{"%ready"}) {
		t.Errorf("placeholders: %q", got)
	}
}

func TestCompleteBelowOpenBracket(t *testing.T) {
	below := "\n  sequence \"S\" [none]\n\ngroup \"G\":\n  none\n"
	if got := complete(t, langPreface+"priority:\n  reset_sequence(\"|"+below); !slices.Equal(got, []string{`"S"`}) {
		t.Errorf("reset_sequence: %q", got)
	}
	if got := complete(t, langPreface+"priority:\n  cast x if sequence(|"+below); !slices.Equal(got, []string{`"S"`}) {
		t.Errorf("sequence(: %q", got)
	}
	hasAll(t, "action_group_used", complete(t, langPreface+"priority:\n  cast x if action_group_used(|"+below), `"G"`)
	continued := langPreface + "priority:\n  cast x if and(\n      gcd_is_ready,\n      action_group_used(\"|" + below
	if got := complete(t, continued); !slices.Equal(got, []string{`"G"`}) {
		t.Errorf("opened again on a continuation line: %q", got)
	}
}

func TestCompleteBelowManyOpenBrackets(t *testing.T) {
	open := langPreface + "priority:\n  cast x if action_group_used(|\n  cast x if and(\n  cast x if [\n  cast x if or(gcd_is_ready,\n"
	hasAll(t, "below four open brackets", complete(t, open+"\ngroup \"G\":\n  none\n"), `"G"`)
}

func TestCompleteBracedVariableReplacesName(t *testing.T) {
	text, at := marked(langPreface + "priority:\n  cast x if ${Melee we|ave}\n")
	items := analyzeText(t, hunterPath, text).Complete(at, langCatalog())
	if len(items) != 1 || items[0].Replace != (Range{Position{6, 13}, Position{6, 27}}) {
		t.Errorf("got %+v, want ${Melee weave} replacing 6:13 to 6:27", items)
	}
}

func TestCompleteNothingInTrivia(t *testing.T) {
	if got := complete(t, langPreface+"priority:\n  // cast |\n"); len(got) > 0 {
		t.Errorf("comment: %q", got)
	}
	if got := complete(t, langPreface+"priority:\n  /// cast |\n  none\n"); len(got) > 0 {
		t.Errorf("note: %q", got)
	}
}

func TestCompleteLines(t *testing.T) {
	hasAll(t, "line", complete(t, "siml 1\n|\n", setupPath), "settings:", "gear:", "consumables:", "priority:", `group "Name":`, "spec", "use")
	hasAll(t, "spec", complete(t, "siml 1\nspec |\n", setupPath), "dps_warrior")
}

func TestCompleteEntries(t *testing.T) {
	hasAll(t, "settings", complete(t, "siml 1\nspec hunter\n\nsettings:\n  |\n", setupPath), "race", "distance_from_target", "quiver_bonus", "cobra_reflexes")
	hasAll(t, "race", complete(t, "siml 1\nspec hunter\n\nsettings:\n  race |\n", setupPath), "orc")
	hasAll(t, "quiver", complete(t, "siml 1\nspec hunter\n\nsettings:\n  quiver_bonus |\n", setupPath), "ancient_sinew_wrapped_lamina")
	hasAll(t, "buffs", complete(t, "siml 1\n\nbuffs: |\n", setupPath), "greater_blessing_of_kings")
	hasAll(t, "flag value", complete(t, "siml 1\n\nparty_buffs: mana_spring_totem=|\n", setupPath), "improved")
}

func TestCompleteConsumables(t *testing.T) {
	pots := complete(t, "siml 1\n\nconsumables:\n  offensive_pot |\n", setupPath)
	hasAll(t, "offensive_pot", pots, "major_frenzy_potion")
	hasNone(t, "offensive_pot", pots, "swiftstrike_steak")
	food := complete(t, "siml 1\n\nconsumables:\n  food |\n", setupPath)
	hasAll(t, "food", food, "swiftstrike_steak")
	hasNone(t, "food", food, "major_frenzy_potion")
}

func TestCompleteSlotRows(t *testing.T) {
	hasAll(t, "slot", complete(t, "siml 1\n\n|\n", gearPath), "main_hand")
	heads := complete(t, "siml 1\n\nhead |\n", gearPath)
	hasAll(t, "head", heads, "lieutenant_commanders_chain_helmet")
	hasNone(t, "head", heads, "mongoose_boots")
	offHand := complete(t, "siml 1\n\noff_hand |\n", gearPath)
	hasAll(t, "off hand", offHand, "thunderfury_blessed_blade_of_the_windseeker")
	hasNone(t, "a two-hander goes in the main hand only", offHand, "legionite_glaive")
	hasAll(t, "suffixes", complete(t, "siml 1\n\nmain_hand northern_shortsword |\n", gearPath), "of_the_eagle", "+")
	enchants := complete(t, "siml 1\n\nhead lieutenant_commanders_chain_helmet + |\n", gearPath)
	hasAll(t, "enchants", enchants, "falcons_call")
	hasNone(t, "enchants", enchants, "cloak_agility")
	hasAll(t, "two-hand", complete(t, "siml 1\n\nmain_hand legionite_glaive + |\n", gearPath), "enchant_2h_weapon_agility")
	hasNone(t, "one-hand", complete(t, "siml 1\n\nmain_hand thunderfury_blessed_blade_of_the_windseeker + |\n", gearPath), "enchant_2h_weapon_agility")
}

func hoverText(t *testing.T, d *Doc, at Position) string {
	t.Helper()
	h, ok := d.Hover(at, langCatalog())
	if !ok {
		t.Fatalf("no hover at %v", at)
	}
	return h.Markdown
}

func matchesText(t *testing.T, what, text, pattern string) {
	t.Helper()
	if !regexp.MustCompile(pattern).MatchString(text) {
		t.Errorf("%s: %q does not match %s", what, text, pattern)
	}
}

func hunterDefault(t *testing.T) (string, *Doc) {
	t.Helper()
	src, err := os.ReadFile(filepath.Join("testdata", "parse", "hunter_default.apl.siml"))
	if err != nil {
		t.Fatal(err)
	}
	return string(src), analyzeText(t, hunterPath, string(src))
}

func TestHoverSpell(t *testing.T) {
	text, d := hunterDefault(t)
	h, ok := d.Hover(positionOf(text, "hunters_mark", 1), langCatalog())
	if !ok {
		t.Fatal("no hover")
	}
	id := tableSpellID(t, "hunters_mark")
	matchesText(t, "spell", h.Markdown, fmt.Sprintf(`\*\*Hunter's Mark\*\* · a spell · spell %d`, id))
	matchesText(t, "spell", h.Markdown, `registered by hunter`)
	if h.SpellID != id {
		t.Errorf("SpellID %d, want %d, whose game files row the server adds", h.SpellID, id)
	}
}

func TestHoverWords(t *testing.T) {
	text, d := hunterDefault(t)
	matchesText(t, "aura on a target", hoverText(t, d, positionOf(text, "aura(hunters_mark", 6)), `an aura on a target`)
	matchesText(t, "unit", hoverText(t, d, positionOf(text, "target.aura", 1)), `"CurrentTarget"`)
	matchesText(t, "kind", hoverText(t, d, positionOf(text, "auto_time_to_next", 1)), `kind \*\*auto_time_to_next\*\*`)
	matchesText(t, "variable", hoverText(t, d, positionOf(text, "$melee_weave", 2)), `var \$\{Melee weave\} = true`)
}

func TestHoverGear(t *testing.T) {
	gear := "siml 1\n\nhead lieutenant_commanders_chain_helmet + falcons_call\nmain_hand northern_shortsword of_the_eagle\n"
	d := analyzeText(t, gearPath, gear)
	matchesText(t, "item", hoverText(t, d, positionOf(gear, "lieutenant", 2)), `Lieutenant Commander's Chain Helm.* · item \d+`)
	falcons := langNames(t).Enchants["falcons_call"]
	fits := regexp.QuoteMeta("fits " + strings.Join(d.slotWords(falcons.Slots), ", "))
	matchesText(t, "enchant", hoverText(t, d, positionOf(gear, "falcons_call", 2)), fmt.Sprintf(`enchant effect %d[\s\S]*%s`, falcons.EffectID, fits))
	matchesText(t, "suffix", hoverText(t, d, positionOf(gear, "of_the_eagle", 2)), `random suffix`)
}

func editorDiagnostics(t *testing.T, path, text string) Diagnostics {
	t.Helper()
	return analyzeText(t, path, text).EditorDiagnostics(langCatalog())
}

func TestChecksDidYouMean(t *testing.T) {
	found := editorDiagnostics(t, hunterPath, plainPreface+"priority:\n  cast hunters_mrk\n")
	if got, want := diagRows(found), []diagRow{{4, "unknown-name", 1}}; !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	matchesText(t, "message", found[0].Message, `did you mean hunters_mark`)
}

func TestChecksLints(t *testing.T) {
	var got [][3]int
	for _, x := range editorDiagnostics(t, hunterPath, plainPreface+"prepull:\n  1s cast hunters_mark\n") {
		if x.Code == "prepull-positive-time" {
			got = append(got, [3]int{x.Start.Line - 1, x.Start.Col - 1, x.End.Col - 1})
		}
	}
	if want := [][3]int{{4, 2, 4}}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestChecksUnderSpecs(t *testing.T) {
	text := plainPreface + "priority:\n  cast `{\"spellId\":14325}`\n  group \"Nowhere\"\n"
	codes := func(diags Diagnostics) []string {
		var out []string
		for _, x := range diags {
			out = append(out, x.Code+" "+x.Severity.String())
		}
		slices.Sort(out)
		return out
	}
	if got, want := codes(editorDiagnostics(t, hunterPath, text)), []string{"undefined-group warning"}; !slices.Equal(got, want) {
		t.Errorf("outside: %q, want %q", got, want)
	}
	if got, want := codes(editorDiagnostics(t, specsPath, text)), []string{"raw-ref error", "undefined-group error"}; !slices.Equal(got, want) {
		t.Errorf("under ui/specs: %q, want %q", got, want)
	}
}

func TestChecksGear(t *testing.T) {
	gear := textLines(
		"siml 1",
		"",
		"head fury_of_the_forgotten_swarm + cloak_agility",
		"main_hand thunderfury_blessed_blade_of_the_windseeker + enchant_2h_weapon_agility",
		"neck fury_of_the_forgotten_swarm of_the_eagle",
		"off_hand legionite_glaive",
	)
	var got []diagRow
	for _, r := range diagRows(editorDiagnostics(t, gearPath, gear)) {
		got = append(got, diagRow{r.line, r.code, 0})
	}
	want := []diagRow{{2, "gear-slot", 0}, {2, "gear-enchant", 0}, {3, "gear-enchant", 0}, {4, "gear-suffix", 0}, {5, "gear-slot", 0}}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// --- Variables, groups, sequences and placeholders.

var symbolsText = textLines(
	"siml 1",
	"spec hunter",
	"",
	"var ${Melee weave} = true",
	"",
	"priority:",
	`  group "Weave" with %on = $melee_weave if ${Melee weave}`,
	`  reset_sequence("S") if sequence("S").is_complete`,
	`  sequence "S" [none]`,
	`  cast hunters_mark if action_group_used("Weave")`,
	"",
	`group "Weave":`,
	"  cast hunters_mark if %on",
)

// Each location as "line:character", both counted from 0 as an editor counts them.
func lineCols(locations []Location) []string {
	var out []string
	for _, l := range locations {
		out = append(out, fmt.Sprintf("%d:%d", l.Range.Start.Line-1, l.Range.Start.Col-1))
	}
	return out
}

func TestDefinition(t *testing.T) {
	d := analyzeText(t, hunterPath, symbolsText)
	for _, c := range []struct {
		needle string
		k      int
		want   []string
	}{
		{"$melee_weave", 1, []string{"3:4"}},
		{`action_group_used("Weave`, 20, []string{"11:6"}},
		{`sequence("S"`, 10, []string{"8:11"}},
	} {
		if got := lineCols(d.Definition(positionOf(symbolsText, c.needle, c.k))); !slices.Equal(got, c.want) {
			t.Errorf("%s: %v, want %v", c.needle, got, c.want)
		}
	}
}

func TestReferences(t *testing.T) {
	d := analyzeText(t, hunterPath, symbolsText)
	if got, want := lineCols(d.References(positionOf(symbolsText, "${Melee weave} = true", 2), true)), []string{"3:4", "6:27", "6:43"}; !slices.Equal(got, want) {
		t.Errorf("variable: %v, want %v", got, want)
	}
	if got, want := lineCols(d.References(positionOf(symbolsText, `"Weave":`, 1), false)), []string{"6:8", "9:41"}; !slices.Equal(got, want) {
		t.Errorf("group: %v, want %v", got, want)
	}
	if got, want := lineCols(d.References(positionOf(symbolsText, "%on", 1), true)), []string{"6:21", "12:23"}; !slices.Equal(got, want) {
		t.Errorf("placeholder: %v, want %v", got, want)
	}
}

func TestGroupVariableShadows(t *testing.T) {
	text := textLines(
		"siml 1", "spec hunter", "", "var $x = true", "", "priority:", `  group "G"`, "  cast hunters_mark if $x", "",
		`group "G":`, "  var $x = false", "  cast arcane_shot if $x",
	)
	d := analyzeText(t, hunterPath, text)
	if got, want := lineCols(d.Definition(positionOf(text, "arcane_shot if $x", 16))), []string{"10:6"}; !slices.Equal(got, want) {
		t.Errorf("the group's $x: %v, want %v", got, want)
	}
	if got, want := lineCols(d.References(positionOf(text, "$x = true", 1), true)), []string{"3:4", "7:23"}; !slices.Equal(got, want) {
		t.Errorf("the file's $x: %v, want %v", got, want)
	}
}

func TestRenameThroughUse(t *testing.T) {
	dir := t.TempDir()
	usedPath, userPath := filepath.Join(dir, "a.apl.siml"), filepath.Join(dir, "b.siml")
	used := textLines("siml 1", "spec hunter", "", "var $x = true", "", "priority:", `  group "G" if $x`, "", `group "G":`, "  cast arcane_shot")
	user := textLines("siml 1", "spec hunter", `use "a.apl.siml"`, "", "prepull:", `  -1s group "G" if $x`)
	for path, text := range map[string]string{usedPath: used, userPath: user} {
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	d := analyzeText(t, userPath, user)
	for _, c := range []struct{ needle, newName, want string }{{"$x", "y", "$y"}, {`"G"`, "H", `"H"`}} {
		at := positionOf(user, c.needle, 1)
		if got := whereIs(d.Definition(at)); len(got) != 1 || !strings.HasPrefix(got[0], "a.apl.siml ") {
			t.Errorf("%s: definition %q, want in a.apl.siml", c.needle, got)
		}
		edits, err := d.Rename(at, c.newName)
		if err != nil {
			t.Fatal(err)
		}
		var got []string
		for _, f := range edits {
			got = append(got, fmt.Sprintf("%s %d %s", filepath.Base(f.Path), len(f.Edits), f.Edits[0].NewText))
		}
		if want := []string{"a.apl.siml 2 " + c.want, "b.siml 1 " + c.want}; !slices.Equal(got, want) {
			t.Errorf("%s: %q, want %q", c.needle, got, want)
		}
	}
}

func renamedTexts(t *testing.T, d *Doc, at Position, newName string) []string {
	t.Helper()
	edits, err := d.Rename(at, newName)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, f := range edits {
		for _, e := range f.Edits {
			out = append(out, e.NewText)
		}
	}
	return out
}

func TestRenameSpelling(t *testing.T) {
	d := analyzeText(t, hunterPath, symbolsText)
	if got, want := renamedTexts(t, d, positionOf(symbolsText, "$melee_weave", 1), "weave"), []string{"$weave", "$weave", "$weave"}; !slices.Equal(got, want) {
		t.Errorf("variable: %q", got)
	}
	if got, want := renamedTexts(t, d, positionOf(symbolsText, `"S"`, 1), "Two words"), []string{`"Two words"`, `"Two words"`, `"Two words"`}; !slices.Equal(got, want) {
		t.Errorf("sequence: %q", got)
	}
	if got, want := renamedTexts(t, d, positionOf(symbolsText, `"Weave":`, 1), "A & <B>"), `"A & <B>"`; len(got) != 3 || got[0] != want {
		t.Errorf("group: %q, want %q", got, want)
	}
	if _, _, err := d.PrepareRename(positionOf(symbolsText, "hunters_mark", 1)); err == nil {
		t.Error("a spell name renames")
	}
}

func TestRenameRefusals(t *testing.T) {
	d := analyzeText(t, hunterPath, symbolsText)
	refused := func(d *Doc, text, needle, newName string) string {
		_, err := d.Rename(positionOf(text, needle, 1), newName)
		if err == nil {
			return ""
		}
		return err.Error()
	}
	for _, c := range []struct{ needle, newName, want string }{
		{"$melee_weave", "a}b", "A variable's name cannot hold } or a line break."},
		{"%on", "a\nb", "A placeholder's name cannot hold } or a line break."},
		{`"S" [`, "", "A name cannot be empty."},
	} {
		if got := refused(d, symbolsText, c.needle, c.newName); got != c.want {
			t.Errorf("%s → %q: %q, want %q", c.needle, c.newName, got, c.want)
		}
	}
	twoText := strings.Replace(symbolsText, "var ${Melee weave} = true", "var ${Melee weave} = true\nvar $other = false", 1)
	two := analyzeText(t, hunterPath, twoText)
	for _, newName := range []string{"other", "Other"} {
		if got, want := refused(two, twoText, "$melee_weave", newName), "x.apl.siml has a variable named other already."; got != want {
			t.Errorf("%s: %q, want %q", newName, got, want)
		}
	}
	groups := analyzeText(t, hunterPath, symbolsText+"\ngroup \"Other\":\n  none\n")
	if got, want := refused(groups, symbolsText, `"Weave":`, "Other"), "x.apl.siml has a group named Other already."; got != want {
		t.Errorf("group: %q, want %q", got, want)
	}
	if got := renamedTexts(t, d, positionOf(symbolsText, "$melee_weave", 1), "Melee Weave"); len(got) != 3 {
		t.Errorf("its own name in another spelling: %q", got)
	}
}

func TestOutlineAndFolding(t *testing.T) {
	d := analyzeText(t, hunterPath, symbolsText)
	symbols := d.Symbols()
	var names []string
	for _, s := range symbols {
		names = append(names, s.Name)
	}
	if want := []string{"Melee weave", "priority", `group "Weave"`}; !slices.Equal(names, want) {
		t.Fatalf("symbols %q, want %q", names, want)
	}
	if len(symbols[1].Children) != 1 || symbols[1].Children[0].Name != "S" {
		t.Errorf("priority holds %v", symbols[1].Children)
	}
	if !slices.ContainsFunc(d.Folds(), func(f Fold) bool { return f.StartLine == 6 && f.EndLine == 10 }) {
		t.Errorf("no fold over the priority section: %v", d.Folds())
	}
}

func TestUseLinks(t *testing.T) {
	path := filepath.Join(langRoot, "ui", "specs", "hunter", "dps", "builds", "mm.siml")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(src)
	d := analyzeText(t, path, text)
	type link struct {
		line int
		path string
	}
	var got []link
	for _, l := range d.Links() {
		rel, _ := filepath.Rel(langRoot, l.Path)
		got = append(got, link{l.Range.Start.Line - 1, filepath.ToSlash(rel)})
	}
	want := []link{{3, "ui/specs/hunter/dps/gear_sets/p1.gear.siml"}, {4, "ui/specs/hunter/dps/apls/turret.apl.siml"}}
	if !slices.Equal(got, want) {
		t.Fatalf("links %v, want %v", got, want)
	}
	if defs := d.Definition(positionOf(text, "apls/turret", 3)); len(defs) != 1 || !SamePath(defs[0].Path, d.Links()[1].Path) {
		t.Errorf("definition of the use path: %v", defs)
	}
	missing := analyzeText(t, path, strings.Replace(text, "../apls/turret.apl.siml", "../apls/missing.apl.siml", 1))
	if links := missing.Links(); len(links) != 1 || links[0].Range.Start.Line != 4 {
		t.Errorf("a link to a file that does not exist: %v", links)
	}
}

// default.apl.siml declares ${Use Summon Hawk}, which hawks.apl.siml beside it sets.
func fixtureDoc(t *testing.T, name string) (string, *Doc) {
	t.Helper()
	path := filepath.Join(langRoot, "sim", "siml", "lsp", "testdata", name)
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(src), analyzeText(t, path, string(src))
}

// Each location as "file line:character", counted from 0.
func whereIs(locations []Location) []string {
	var out []string
	for i, l := range lineCols(locations) {
		out = append(out, filepath.Base(locations[i].Path)+" "+l)
	}
	return out
}

// A def a used file declares: its calls here go to its line there, hover shows that line, completion offers it, and a
// rename reaches both files.
func TestImportedDef(t *testing.T) {
	dir := t.TempDir()
	usedPath, userPath := filepath.Join(dir, "a.apl.siml"), filepath.Join(dir, "b.apl.siml")
	used := textLines("siml 1", "spec hunter", "", "def far(x) = spell(x).is_ready and self.unit_distance > 5")
	user := textLines("siml 1", "spec hunter", `use "a.apl.siml"`, "", "priority:", "  cast multi_shot if far(multi_shot)", "  cast arcane_shot if ")
	for path, text := range map[string]string{usedPath: used, userPath: user} {
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	d := analyzeText(t, userPath, user)
	at := positionOf(user, "far(", 1)
	if got := whereIs(d.Definition(at)); len(got) != 1 || got[0] != "a.apl.siml 3:4" {
		t.Errorf("definition %q, want a.apl.siml 3:4", got)
	}
	if h, ok := d.Hover(at, langCatalog()); !ok || !strings.Contains(h.Markdown, "def far(x) = spell(x).is_ready") {
		t.Errorf("hover %q", h.Markdown)
	}
	if labels := labelsOf(d.Complete(Position{7, 23}, langCatalog())); !slices.Contains(labels, "far") {
		t.Errorf("completion offers %q, without far", labels)
	}
	edits, err := d.Rename(at, "beyond")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, f := range edits {
		got = append(got, fmt.Sprintf("%s %d %s", filepath.Base(f.Path), len(f.Edits), f.Edits[0].NewText))
	}
	if want := []string{"a.apl.siml 1 beyond", "b.apl.siml 1 beyond"}; !slices.Equal(got, want) {
		t.Errorf("rename %q, want %q", got, want)
	}
}

func TestWithBindDefinition(t *testing.T) {
	settingText, setting := fixtureDoc(t, "hawks.apl.siml")
	bind := positionOf(settingText, "$use_summon_hawk", 1)
	if got, want := whereIs(setting.Definition(bind)), []string{"default.apl.siml 15:4"}; !slices.Equal(got, want) {
		t.Errorf("definition %q, want %q", got, want)
	}
	if _, name, err := setting.PrepareRename(bind); err != nil || name != "Use Summon Hawk" {
		t.Errorf("prepare rename: %q, %v", name, err)
	}
}

func TestWithBindReferences(t *testing.T) {
	declaredText, declared := fixtureDoc(t, "default.apl.siml")
	settingText, setting := fixtureDoc(t, "hawks.apl.siml")
	all := []string{
		"default.apl.siml 15:4", "default.apl.siml 29:28", "default.apl.siml 31:9",
		"default.apl.siml 33:22", "default.apl.siml 34:47", "hawks.apl.siml 3:28",
	}
	if got := whereIs(setting.References(positionOf(settingText, "$use_summon_hawk", 1), true)); !slices.Equal(got, all) {
		t.Errorf("from the with list: %q", got)
	}
	if got := whereIs(declared.References(positionOf(declaredText, "${Use Summon Hawk}", 2), true)); !slices.Equal(got, all) {
		t.Errorf("from the declaration: %q", got)
	}
}

func TestWithBindRename(t *testing.T) {
	declaredText, declared := fixtureDoc(t, "default.apl.siml")
	settingText, setting := fixtureDoc(t, "hawks.apl.siml")
	for _, c := range []struct {
		d  *Doc
		at Position
	}{
		{declared, positionOf(declaredText, "${Use Summon Hawk}", 2)},
		{setting, positionOf(settingText, "$use_summon_hawk", 1)},
	} {
		edits, err := c.d.Rename(c.at, "Summon hawks")
		if err != nil {
			t.Fatal(err)
		}
		var got []string
		for _, f := range edits {
			got = append(got, fmt.Sprintf("%s %d %s", filepath.Base(f.Path), len(f.Edits), f.Edits[0].NewText))
		}
		if want := []string{"default.apl.siml 5 ${Summon hawks}", "hawks.apl.siml 1 ${Summon hawks}"}; !slices.Equal(got, want) {
			t.Errorf("from %s: %q, want %q", filepath.Base(c.d.Path()), got, want)
		}
	}
}

func TestWithBindRenameRefused(t *testing.T) {
	settingText, setting := fixtureDoc(t, "hawks.apl.siml")
	_, err := setting.Rename(positionOf(settingText, "$use_summon_hawk", 1), "Use Aimed Shot")
	if err == nil || err.Error() != "default.apl.siml has a variable named Use Aimed Shot already." {
		t.Errorf("got %v", err)
	}
}

func TestTokens(t *testing.T) {
	text := langPreface + "priority:\n  cast hunters_mark if target.aura(hunters_mark).is_active and $melee_weave\n"
	d := analyzeText(t, hunterPath, text)
	var got [][2]string
	for _, tok := range d.Tokens() {
		got = append(got, [2]string{text[d.Offset(tok.Range.Start):d.Offset(tok.Range.End)], tok.Type})
	}
	want := [][2]string{
		{"hunter", "class"},
		{"${Melee weave}", "variable"},
		{"priority", "namespace"},
		{"hunters_mark", "type"},
		{"target", "variable"},
		{"aura", "function"},
		{"hunters_mark", "type"},
		{"is_active", "method"},
		{"$melee_weave", "variable"},
	}
	if !slices.Equal(got, want) {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestQuickFixDidYouMean(t *testing.T) {
	text := langPreface + "priority:\n  cast mongoos_bite if target.aura(hunters_mrk).is_active\n"
	d := analyzeText(t, hunterPath, text)
	fixes := d.QuickFixes(Range{Position{6, 1}, Position{6, 61}})
	type row struct {
		title, kind string
		preferred   bool
		col         int
	}
	var got []row
	for _, f := range fixes {
		got = append(got, row{f.Title, f.Kind, f.Preferred, f.Diagnostic.Start.Col - 1})
	}
	want := []row{{"Change to mongoose_bite", "quickfix", true, 7}, {"Change to hunters_mark", "quickfix", true, 35}}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	fixed := text
	for _, fixe := range slices.Backward(fixes) {
		e := fixe.Edits[0]
		fixed = fixed[:d.Offset(e.Range.Start)] + e.NewText + fixed[d.Offset(e.Range.End):]
	}
	var codes []string
	for _, x := range editorDiagnostics(t, hunterPath, fixed) {
		codes = append(codes, x.Code)
	}
	if want := []string{"unused-variable"}; !slices.Equal(codes, want) {
		t.Errorf("the fixed file: %q, want %q", codes, want)
	}
}

// The diagnostics of an opened file and of a file that uses another, and the completions and hover the server gives
// from them.
func TestServerContentCases(t *testing.T) {
	text := textLines("siml 1", "spec hunter", "", "priority:", "  cast hunters_mrk", "  ")
	d := analyzeText(t, specsPath, text)
	diags := d.EditorDiagnostics(langCatalog())
	if len(diags) != 1 || diags[0].Code != "unknown-name" || diags[0].Start != (Position{5, 8}) {
		t.Fatalf("diagnostics %v", diags)
	}
	fixes := d.QuickFixes(Range{diags[0].Start, diags[0].End})
	if len(fixes) == 0 || fixes[0].Title != "Change to hunters_mark" || fixes[0].Edits[0].Range != (Range{diags[0].Start, diags[0].End}) {
		t.Errorf("quick fix %v", fixes)
	}
	labels := labelsOf(d.Complete(Position{6, 3}, langCatalog()))
	hasAll(t, "row start", labels, "cast", "wait until")
	hasAll(t, "after cast", labelsOf(d.Complete(Position{5, 13}, langCatalog())), "hunters_mark")
	fixedText := textLines("siml 1", "spec hunter", "", "priority:", "  cast hunters_mark")
	h, ok := analyzeText(t, specsPath, fixedText).Hover(Position{5, 10}, langCatalog())
	if !ok {
		t.Fatal("no hover")
	}
	matchesText(t, "hover", h.Markdown, `Hunter's Mark`)
	matchesText(t, "hover", h.Markdown, fmt.Sprintf(`spell %d`, tableSpellID(t, "hunters_mark")))

	// A file that uses one which does not declare the variable its with list sets.
	usedText, _ := fixtureDoc(t, "default.apl.siml")
	settingText, _ := fixtureDoc(t, "hawks.apl.siml")
	usedPath := filepath.Join(langRoot, "sim", "siml", "lsp", "testdata", "default.apl.siml")
	unsaved := strings.Replace(usedText, "var ${Use Summon Hawk} = false\n", "", 1)
	open := func(path string) ([]byte, error) {
		if SamePath(filepath.Clean(path), usedPath) {
			return []byte(unsaved), nil
		}
		return os.ReadFile(path)
	}
	settingPath := filepath.Join(filepath.Dir(usedPath), "hawks.apl.siml")
	before := Analyze(settingPath, []byte(settingText), langNames(t), open).EditorDiagnostics(langCatalog())
	if len(before) != 1 || before[0].Code != "with-unknown-variable" {
		t.Errorf("with the unsaved text: %v", before)
	}
	if after := Analyze(settingPath, []byte(settingText), langNames(t), langOpen).EditorDiagnostics(langCatalog()); len(after) != 0 {
		t.Errorf("read from disk: %v", after)
	}
}

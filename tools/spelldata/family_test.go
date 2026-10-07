package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestFamilyIndex(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"-family", "warrior/Execute"}, &out); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(out.String(), "\n")
	want := []string{
		"warrior spellData.Execute",
		"5308     Execute  Rank 1   Rank(1)   effect 1 = 125",
		"20658    Execute  Rank 2   Rank(2)   effect 1 = 200",
		"20660    Execute  Rank 3   Rank(3)   effect 1 = 325",
		"20661    Execute  Rank 4   Rank(4)   effect 1 = 450",
		"20662    Execute  Rank 5   Highest() effect 1 = 600",
		"",
		"20662 Execute (Rank 5) warrior spellData.Execute.Highest()",
	}
	for i, line := range want {
		if i >= len(lines) || lines[i] != line {
			t.Fatalf("line %d is %q, want %q", i+1, lines[i], line)
		}
	}
}

// A talent is one spell whose ranks are a curve, so the index is one line per rank of the ladder
// Talent builds, with the value the curve gives effect 1 at that rank.
func TestFamilyIndexTalent(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"-family", "warrior/Cruelty"}, &out); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(out.String(), "\n")
	want := []string{
		"warrior spellData.Cruelty",
		"12320    Cruelty  rank 1 of 5 Rank(1)   effect 1 = 1",
		"12320    Cruelty  rank 2 of 5 Rank(2)   effect 1 = 2",
		"12320    Cruelty  rank 3 of 5 Rank(3)   effect 1 = 3",
		"12320    Cruelty  rank 4 of 5 Rank(4)   effect 1 = 4",
		"12320    Cruelty  rank 5 of 5 Highest() effect 1 = 5",
		"",
		"12320 Cruelty (rank 5 of 5) warrior spellData.Cruelty.Rank(n), n up to 5",
	}
	for i, line := range want {
		if i >= len(lines) || lines[i] != line {
			t.Fatalf("line %d is %q, want %q", i+1, lines[i], line)
		}
	}
}

func TestFamilyTalentJSON(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"-family", "warrior/ImprovedRend", "-json"}, &out); err != nil {
		t.Fatal(err)
	}

	var got familyJSON
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	values := make([]string, 0, len(got.Ranks))
	for _, rank := range got.Ranks {
		values = append(values, rank.Value)
	}
	if strings.Join(values, ", ") != "effect 1 = 12, effect 1 = 23, effect 1 = 35" {
		t.Errorf("the ranks read %v", values)
	}
	if got.Highest.heading() != "12286 Improved Rend (rank 3 of 3)" {
		t.Errorf("the highest rank is titled %q", got.Highest.heading())
	}
}

func TestFamilyWithoutAPackage(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"-family", "Execute"}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out.String(), "warrior spellData.Execute\n") {
		t.Fatalf("the heading is %q", strings.SplitN(out.String(), "\n", 2)[0])
	}
}

func TestFamilyJSON(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"-family", "warrior/Execute", "-json"}, &out); err != nil {
		t.Fatal(err)
	}

	var got familyJSON
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Family != "warrior/Execute" {
		t.Errorf("family %q", got.Family)
	}
	if len(got.Ranks) != 5 {
		t.Fatalf("%d ranks", len(got.Ranks))
	}
	if got.Ranks[0] != (familyRow{ID: 5308, Name: "Execute", Rank: "Rank 1", Accessor: "Rank(1)", Value: "effect 1 = 125"}) {
		t.Errorf("rank 1 is %+v", got.Ranks[0])
	}
	if got.Ranks[4].Accessor != "Highest()" {
		t.Errorf("the top rank is reached by %q", got.Ranks[4].Accessor)
	}
	if got.Highest.ID != 20662 || len(got.Highest.Effects) == 0 {
		t.Errorf("the highest rank is %+v", got.Highest)
	}
}

func TestExpr(t *testing.T) {
	cases := []struct {
		expr string
		id   int32
	}{
		{"spellData.Execute.Highest()", 20662},
		{"spellData.Execute.Rank(3)", 20660},
		{"spellData.Execute.ByID(20658)", 20658},
		{"Execute.Rank(1)", 5308},
		// Every rank of a talent is the one spell the curve belongs to.
		{"spellData.Cruelty.Rank(2)", 12320},
		{"spellData.Cruelty.Highest()", 12320},
	}

	for _, c := range cases {
		var out bytes.Buffer
		if err := run([]string{"-expr", c.expr, "-package", "warrior", "-json"}, &out); err != nil {
			t.Errorf("%s: %v", c.expr, err)
			continue
		}

		var got exprJSON
		if err := json.Unmarshal(out.Bytes(), &got); err != nil {
			t.Errorf("%s: %v", c.expr, err)
			continue
		}
		if got.ID != c.id {
			t.Errorf("%s resolves to %d, want %d", c.expr, got.ID, c.id)
		}
	}
}

// A talent rank is the rank Talent builds from the curve, so every accessor after it reads what the sim
// reads at that rank rather than the store's base row. Improved Rend's curve is not linear, and its base
// row states 15, which no rank has.
func TestExprTalentRank(t *testing.T) {
	cases := map[string]string{
		"spellData.Cruelty.Rank(1).EffectN(1).BaseValue()":                 "1",
		"spellData.Cruelty.Rank(3).EffectN(1).BaseValue()":                 "3",
		"spellData.Cruelty.Rank(5).EffectN(1).BaseValue()":                 "5",
		"spellData.Cruelty.ByID(12320).EffectN(1).BaseValue()":             "1",
		"spellData.ImprovedRend.Rank(1).EffectN(1).BaseValue()":            "12",
		"spellData.ImprovedRend.Rank(2).EffectN(1).BaseValue()":            "23",
		"spellData.ImprovedRend.Highest().EffectN(1).BaseValue()":          "35",
		"spellData.DualWieldSpecialization.Rank(2).EffectN(2).BaseValue()": "40",
	}
	for expr, want := range cases {
		if got := evalJSON(t, expr); got.Value != want {
			t.Errorf("%s = %s, want %s", expr, got.Value, want)
		}
	}

	if got := evalJSON(t, "spellData.Cruelty.Rank(3)"); got.Rank != "rank 3 of 5" {
		t.Errorf("the rank is %q", got.Rank)
	}

	var out bytes.Buffer
	if err := run([]string{"-expr", "spellData.Cruelty.Rank(3).EffectN(1)", "-package", "warrior"}, &out); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"spellData.Cruelty.Rank(3).EffectN(1) = effect 1 of 12320 Cruelty (rank 3 of 5)",
		"12320 Cruelty (rank 3 of 5) warrior spellData.Cruelty.Rank(n), n up to 5",
		"E_APPLY_AURA A_MOD_WEAPON_CRIT_PERCENT base=3 ",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("the text lacks %q:\n%s", want, out.String())
		}
	}
}

func TestExprText(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"-expr", "spellData.Execute.Highest()", "-package", "warrior"}, &out); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(out.String(), "\n")
	if lines[0] != "spellData.Execute.Highest() = 20662" {
		t.Fatalf("the first line is %q", lines[0])
	}
	if !strings.HasPrefix(lines[2], "20662 Execute (Rank 5)") {
		t.Fatalf("the card is %q", lines[2])
	}
}

func TestExprRefused(t *testing.T) {
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"-expr", "spellData.Execute.Rank(9)", "-package", "warrior"}, "has 5 ranks, not rank 9"},
		{[]string{"-expr", "spellData.Execute.Rank(9.0)", "-package", "warrior"}, "has 5 ranks, not rank 9"},
		{[]string{"-expr", "spellData.Execute.ByID(25236)", "-package", "warrior"}, "no rank with id 25236"},
		{[]string{"-expr", "spellData.Nope.Highest()", "-package", "warrior"}, `warrior states no ladder named "Nope"`},
		{[]string{"-expr", "executeRank.EffectN(1)", "-package", "warrior"}, "no ladder-shaped declaration of executeRank"},
		{[]string{"-family", "warrior/Nope"}, `warrior states no ladder named "Nope"`},
		{[]string{"-family", "Execute", "-expr", "spellData.Execute.Highest()"}, "ask for one thing"},
		{[]string{"-family"}, "-family takes a value"},
	}

	for _, c := range cases {
		var out bytes.Buffer
		err := run(c.args, &out)
		if err == nil {
			t.Errorf("%v answered %q", c.args, out.String())
			continue
		}
		if !strings.Contains(err.Error(), c.want) {
			t.Errorf("%v failed with %q, want %q in it", c.args, err, c.want)
		}
	}
}

// The class files a session is editing while this runs are not a fixed index, so the tiebreak between
// classes is tested against one that is.
func TestFamilyAcrossClasses(t *testing.T) {
	index := map[string]*ladderFamily{
		"warrior/Execute": {pkg: "warrior", field: "Execute"},
		"paladin/Execute": {pkg: "paladin", field: "Execute"},
		"rogue/Rupture":   {pkg: "rogue", field: "Rupture"},
	}

	if _, err := findFamily(index, "Execute", ""); err == nil ||
		!strings.Contains(err.Error(), "Execute is a ladder in paladin, warrior") {
		t.Errorf("a name two classes state answered %v", err)
	}
	if family, err := findFamily(index, "Execute", "paladin"); err != nil || family.pkg != "paladin" {
		t.Errorf("the package did not choose: %v", err)
	}
	if family, err := findFamily(index, "warrior/Execute", "paladin"); err != nil || family.pkg != "warrior" {
		t.Errorf("the spec's own class did not win: %v", err)
	}
	if family, err := findFamily(index, "Rupture", ""); err != nil || family.pkg != "rogue" {
		t.Errorf("a name one class states did not answer: %v", err)
	}
	if _, err := findFamily(index, "Whirlwind", ""); err == nil ||
		!strings.Contains(err.Error(), `no class file states a ladder named "Whirlwind"`) {
		t.Errorf("an unknown name answered %v", err)
	}
}

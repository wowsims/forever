package main

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

func evalJSON(t *testing.T, expr string) exprJSON {
	t.Helper()

	var out bytes.Buffer
	if err := run([]string{"-expr", expr, "-package", "warrior", "-json"}, &out); err != nil {
		t.Fatalf("%s: %v", expr, err)
	}

	var got exprJSON
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("%s: %v", expr, err)
	}
	return got
}

// The effect the chain read, counted from 1, or 0 where it read none.
func readEffect(got exprJSON) int {
	for i, line := range got.Effects {
		if line.Read {
			return i + 1
		}
	}
	return 0
}

func TestExprValue(t *testing.T) {
	cases := []struct {
		expr  string
		kind  string
		trail string
		value string
		read  int
		id    int32
	}{
		{
			expr:  "spellData.Execute.Highest().EffectN(1).AverageAt(60)",
			kind:  kindValue,
			trail: "spellData.Execute.Highest().EffectN(1).AverageAt(60)",
			value: "600",
			read:  1,
			id:    20662,
		},
		{
			// The name a class file writes is substituted, and the trail states what it stood for.
			expr:  "spellData.Bloodthirst.Highest().DamageEffect().Average()",
			kind:  kindValue,
			trail: "spellData.Bloodthirst.Highest().DamageEffect().Average()",
			value: "48",
			read:  1,
			id:    23894,
		},
		{
			expr:  "spellData.Rend.Highest().EffectN(1).Period()",
			kind:  kindValue,
			trail: "spellData.Rend.Highest().EffectN(1).Period()",
			value: "3s",
			read:  1,
			id:    11574,
		},
		{
			// An effect the row does not carry answers the zero the store answers, and reads no effect.
			expr:  "spellData.Execute.Highest().EffectN(9).AverageAt(60)",
			kind:  kindValue,
			trail: "spellData.Execute.Highest().EffectN(9).AverageAt(60)",
			value: "0",
			read:  0,
			id:    20662,
		},
		{
			expr:  "spellData.Execute.Highest().EffectN(1)",
			kind:  kindEffect,
			trail: "spellData.Execute.Highest().EffectN(1)",
			value: "effect 1",
			read:  1,
			id:    20662,
		},
		{
			expr:  "spellData.Execute.Highest()",
			kind:  kindSpell,
			trail: "spellData.Execute.Highest()",
			value: "20662",
			read:  0,
			id:    20662,
		},
		{
			// core.SpellSchool is a byte.
			expr:  "spelldata.MustFind(11574).SpellSchool()",
			kind:  kindValue,
			trail: "spelldata.MustFind(11574).SpellSchool()",
			value: "SpellSchoolPhysical",
			read:  0,
			id:    11574,
		},
	}

	for _, c := range cases {
		got := evalJSON(t, c.expr)
		if got.Kind != c.kind || got.Trail != c.trail || got.Value != c.value ||
			readEffect(got) != c.read || got.ID != c.id {
			t.Errorf("%s answered kind %q trail %q value %q effect %d of %d",
				c.expr, got.Kind, got.Trail, got.Value, readEffect(got), got.ID)
		}
	}
}

// The accessors that read something off the effect the chain stopped on, which is what a reader holding
// an effect wants next.
func TestExprEffectAccessors(t *testing.T) {
	got := evalJSON(t, "spellData.Execute.Highest().EffectN(1)")

	for _, want := range []string{"Average() = 600", "AverageAt(60) = 600", "BaseValue() = 600", "Tenths() = 60"} {
		if !slices.Contains(got.Accessors, want) {
			t.Errorf("the accessors are %v, want %q in them", got.Accessors, want)
		}
	}
	// Period is 0 on an effect that does not tick, and an accessor that reads nothing is left out.
	for _, unwanted := range []string{"Period() = 0s", "Trigger()"} {
		if slices.ContainsFunc(got.Accessors, func(line string) bool { return strings.HasPrefix(line, unwanted) }) {
			t.Errorf("the accessors are %v, want no %q", got.Accessors, unwanted)
		}
	}
}

// The doc comment sim/core/spelldata writes on the accessor that answered, read out of its source.
func TestExprDoc(t *testing.T) {
	got := evalJSON(t, "spellData.Execute.Highest().EffectN(1).AverageAt(60)")
	if !strings.Contains(got.Doc, "level") || !strings.Contains(got.Doc, "scaling") {
		t.Errorf("Average's doc is %q", got.Doc)
	}

	if doc := methodDoc("Spell", "EffectN"); !strings.Contains(doc, "counted from 1") {
		t.Errorf("EffectN's doc is %q", doc)
	}
	if doc := methodDoc("Effect", "Nope"); doc != "" {
		t.Errorf("an accessor that does not exist has the doc %q", doc)
	}
	if doc := evalJSON(t, "spellData.Execute.Highest()").Doc; doc != "" {
		t.Errorf("a pick states the doc %q", doc)
	}
}

// The text form states the value, the accessor's own words and the row, with the effect the chain read
// marked in it.
func TestExprValueText(t *testing.T) {
	var out bytes.Buffer
	args := []string{"-expr", "spellData.Execute.Highest().EffectN(1).AverageAt(core.CharacterLevel)", "-package", "warrior"}
	if err := run(args, &out); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(out.String(), "\n")
	if lines[0] != "spellData.Execute.Highest().EffectN(1).AverageAt(60) = 600" {
		t.Fatalf("the first line is %q", lines[0])
	}
	if !strings.Contains(out.String(), "effect 1 (read) a dummy effect holding 600 to the enemy") {
		t.Errorf("the effect the chain read is not marked:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "    The amount at a caster level") {
		t.Errorf("the accessor's doc is not stated:\n%s", out.String())
	}
}

func TestExprChainRefused(t *testing.T) {
	cases := []struct {
		expr string
		want string
	}{
		{"spellData.Execute.Highest().EffectN(1).Nope()", `"Nope" is not a method of *spelldata.Effect`},
		{"spellData.Execute.Highest().AverageAt(60)", `"AverageAt" is not a method of *spelldata.Spell`},
		{"spellData.Execute.Highest().EffectN(1).AverageAt()", "*spelldata.Effect.AverageAt takes 1 argument, not 0"},
		{"spellData.Execute.Highest().EffectN(1, 2)", "*spelldata.Spell.EffectN takes 1 argument, not 2"},
		{"spellData.Execute.Highest().EffectN(1).AverageAt(level)", "level is not a literal"},
		{"spellData.Execute.Highest().EffectN(1).AverageAt(60.5)", "60.5 is not the int32 it takes"},
		{"spellData.Execute.Highest().EffectN(1).AverageAt(4294967356)", "4294967356 is not the int32 it takes"},
		{"spellData.Execute.Highest().EffectN([]int{1}[0])", "[]int{…}[0] is not a literal"},
		{"spelldata.MustFind(4294978870)", "4294978870 is not the int32 it takes"},
		{"spellData.Execute.Highest().Refs()", "a []*spelldata.Spell is not a value to read"},
		{"spellData.Execute.Highest().ChainAmp", `"ChainAmp" is not a field of *spelldata.Spell`},
		{"spellData.Execute.Highest() + 1", "is not a chain of accessor calls"},
		{"spellData.Execute", "names a ladder, not a rank"},
		// The store panics where the row states no effect the finder names, and the panic is the answer.
		{"spellData.Execute.Highest().Effect(6, 0)", "has no effect with aura 6 misc 0"},
	}

	for _, c := range cases {
		var out bytes.Buffer
		err := run([]string{"-expr", c.expr, "-package", "warrior"}, &out)
		if err == nil {
			t.Errorf("%s answered %q", c.expr, out.String())
			continue
		}
		if !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s failed with %q, want %q in it", c.expr, err, c.want)
		}
	}
}

// A value read off the ladder itself reads the rank its ...At(rank) names, with the effect it read.
func TestExprLadderAccessors(t *testing.T) {
	cases := []struct {
		expr  string
		trail string
		value string
		title string
		read  int
	}{
		{
			expr:  "spellData.Bloodrage.EffectAt(1).TenthsAt(1)",
			trail: "spellData.Bloodrage.EffectAt(1).TenthsAt(1)",
			value: "10",
			title: "2687 Bloodrage",
			read:  1,
		},
		{
			expr:  "spellData.ImprovedBloodrage.MultiplierAt(2)",
			trail: "spellData.ImprovedBloodrage.MultiplierAt(2)",
			value: "1.5",
			title: "12301 Improved Bloodrage (rank 2 of 2)",
			read:  1,
		},
		{
			expr:  "spellData.BattleShout.Highest().Effect(dbcenums.A_MOD_ATTACK_POWER, 0).AverageAt(core.CharacterLevel)",
			trail: "spellData.BattleShout.Highest().Effect(99, 0).AverageAt(60)",
			value: "139",
			title: "25289 Battle Shout (Rank 7)",
			read:  1,
		},
	}
	for _, c := range cases {
		got := evalJSON(t, c.expr)
		if got.Kind != kindValue || got.Trail != c.trail || got.Value != c.value || got.heading() != c.title || readEffect(got) != c.read {
			t.Errorf("%s answered kind %q trail %q value %q on %q effect %d",
				c.expr, got.Kind, got.Trail, got.Value, got.heading(), readEffect(got))
		}
	}

	var out bytes.Buffer
	err := run([]string{"-expr", "spellData.Bloodrage.EffectAt(1)", "-package", "warrior"}, &out)
	if err == nil || !strings.Contains(err.Error(), "names a ladder, not a rank") {
		t.Errorf("a chain ending on the ladder answered %q, %v", out.String(), err)
	}
}

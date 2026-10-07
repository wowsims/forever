package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf16"
)

const executeGo = `package warrior

var executeRank = spellData.Execute.Highest()

var executeBaseDamage = executeRank.EffectN(1).AverageAt(core.CharacterLevel)

var executeDamagePerRage = float64(executeRank.EffectN(1).ChainAmp) * 10

func (warrior *Warrior) registerExecute() {
	baseDamage := executeBaseDamage + executeDamagePerRage*extraRage
}
`

func warriorURI(t *testing.T, name string) string {
	t.Helper()
	root, err := moduleRoot()
	if err != nil {
		t.Fatal(err)
	}
	return pathURI(filepath.Join(root, "sim", "warrior", name))
}

func positionOf(t *testing.T, text, needle string, offset int) (int, int) {
	t.Helper()
	for i, line := range strings.Split(text, "\n") {
		if at := strings.Index(line, needle); at >= 0 {
			return i, at + offset
		}
	}
	t.Fatalf("%q is not in the text", needle)
	return 0, 0
}

func hoverOn(t *testing.T, ws *workspace, text, uri, needle string, offset int) (string, []string, bool) {
	t.Helper()
	line, col := positionOf(t, text, needle, offset)
	return ws.hover(text, line, col, uri, nil)
}

func wantHover(t *testing.T, text, needle string, offset int, wants ...string) string {
	t.Helper()
	markdown, trace, ok := hoverOn(t, newWorkspace(), text, warriorURI(t, "execute.go"), needle, offset)
	if !ok {
		t.Fatalf("no hover on %q:\n%s", needle, strings.Join(trace, "\n"))
	}
	for _, want := range wants {
		if !strings.Contains(markdown, want) {
			t.Errorf("the hover on %q lacks %q:\n%s", needle, want, markdown)
		}
	}
	return markdown
}

func TestHoverSpellIDShapes(t *testing.T) {
	cases := []struct {
		line, needle, uri string
		id                int
	}{
		{"\tvar rend = spelldata.MustFind(11574)", "MustFind(11574", "a.go", 11574},
		{"\tspelldata.Find(116)", "Find(116", "a.go", 116},
		{"\tspellData.Rend.ByID(11574)", "ByID(11574", "a.go", 11574},
		{"\tActionID: core.ActionID{SpellID: 11574},", "SpellID: 11574", "a.go", 11574},
		{"\t\t\t\"spellId\": 11574", "\"spellId\": 11574", "apl.json", 11574},
		{"\tActionId.fromSpellId(23563)", "fromSpellId(23563", "a.ts", 23563},
		{"\t\tspellId: 23563,", "spellId: 23563", "a.tsx", 23563},
	}
	for _, c := range cases {
		markdown, trace, ok := hoverOn(t, newWorkspace(), c.line, "file:///tmp/"+c.uri, c.needle, len(c.needle))
		if !ok || !strings.HasPrefix(markdown, fmt.Sprintf("### %d ", c.id)) {
			t.Errorf("%s answered %v %q:\n%s", c.line, ok, markdown, strings.Join(trace, "\n"))
		}
		if trace[1] != fmt.Sprintf("id %d", c.id) {
			t.Errorf("%s traced %v", c.line, trace)
		}
	}

	if _, _, ok := newWorkspace().hover("\tspell.ApplyEffects = nil", 0, 12, "file:///tmp/a.ts", nil); ok {
		t.Error("a TS line with no id answered a hover")
	}
	if _, _, ok := newWorkspace().hover("\tspelldata.Find(116) // the rank 1", 0, 30, "file:///tmp/a.ts", nil); ok {
		t.Error("a column outside the match answered a hover")
	}
}

// Columns are UTF-16 code units: the dash before the id is one unit and three bytes.
func TestHoverColumnIsUTF16(t *testing.T) {
	line := "// — spelldata.MustFind(11574)"
	col := len([]rune(line[:strings.Index(line, "11574")]))
	if markdown, _, ok := newWorkspace().hover(line, 0, col, "file:///tmp/a.go", nil); !ok || !strings.HasPrefix(markdown, "### 11574 ") {
		t.Errorf("the id after a dash answered %v %q", ok, markdown)
	}
}

func TestHoverFamily(t *testing.T) {
	markdown := wantHover(t, executeGo, "Execute.Highest", 3,
		"### warrior/Execute\n\n| id | name | rank | call | value |\n|--:|:--|:--|:--|:--|\n",
		"| 20662 | Execute | Rank 5 | `Highest()` | effect 1 = 600 |",
		"### 20662 Execute · Rank 5\n")
	if !strings.Contains(markdown, "| **school** | physical |") {
		t.Errorf("the family's highest rank states no header table:\n%s", markdown)
	}
}

// The card a rank pick hovers as: a heading, the ladder call, the row's columns under its name, and the
// effects below a rule with each wording over its game files row.
const executeCard = "### 20662 Execute · Rank 5\n" +
	"`warrior spellData.Execute.Highest()`  \n" +
	"\n" +
	"| | Execute (Rank 5) |\n" +
	"|--:|:--|\n" +
	"| **school** | physical |\n" +
	"| **defense** | melee |\n" +
	"| **gcd** | 1.5 s |\n" +
	"| **cost** | 15 rage |\n" +
	"| **range** | 5 yd |\n" +
	"| **stance** | Battle, Berserker |\n" +
	"| **equip** | needs a melee weapon in the main hand |\n" +
	"| **attrs** | refund on miss |\n" +
	"| **labels** | 25 |\n" +
	"\n" +
	"---\n" +
	"| # | effect |\n" +
	"|--:|:--|\n" +
	"| 1 | a dummy effect holding 600 to the enemy<br>`E_DUMMY base=600 sp=1 target=[6,0]` |\n" +
	"| 2 | casts 26651 Execute to the caster<br>`E_TRIGGER_SPELL base=0 trigger=26651 target=[1,0]` |\n" +
	"\n" +
	"[Wowhead](https://www.wowhead.com/forever/spell=20662)\n"

func TestHoverRankPick(t *testing.T) {
	if got := wantHover(t, executeGo, "executeRank =", 4); got != executeCard {
		t.Errorf("the rank pick hovers as\n%s\nwant\n%s", got, executeCard)
	}
}

func TestHoverValue(t *testing.T) {
	markdown := wantHover(t, executeGo, "executeBaseDamage =", 4,
		"`executeBaseDamage` = **600**\n\n`spellData.Execute.Highest().EffectN(1).AverageAt(60)`",
		"`spellData.Execute.Highest().EffectN(1).AverageAt(60)`\n\n### 20662 Execute · Rank 5\n",
		"| 1 ▶ | a dummy effect holding 600 to the enemy<br>",
		"| 2 | casts 26651 Execute to the caster<br>")
	if strings.Contains(markdown, "The amount at a caster level") {
		t.Errorf("the name's hover states the accessor's doc:\n%s", markdown)
	}

	wantHover(t, executeGo, "executeBaseDamage +", 4, "`executeBaseDamage` = **600**")
}

const crueltyGo = `package warrior

func (warrior *Warrior) applyCruelty() {
	crueltyRank := spellData.Cruelty.Rank(3)
	crueltyCrit := crueltyRank.EffectN(1).BaseValue()
}
`

// A talent rank hovers as the rank Talent builds from the curve, titled with its rank.
func TestHoverTalentRank(t *testing.T) {
	wantHover(t, crueltyGo, "crueltyRank :=", 4,
		"### 12320 Cruelty · rank 3 of 5\n",
		"| | Cruelty (rank 3 of 5) |",
		"| 1 | +3% physical crit, per weapon: only attacks made with a melee weapon<br>`E_APPLY_AURA A_MOD_WEAPON_CRIT_PERCENT base=3 target=[1,0]` |")
	wantHover(t, crueltyGo, "crueltyCrit :=", 4,
		"`crueltyCrit` = **3**\n\n`spellData.Cruelty.Rank(3).EffectN(1).BaseValue()`",
		"### 12320 Cruelty · rank 3 of 5\n",
		"| 1 ▶ | +3% physical crit, per weapon: only attacks made with a melee weapon<br>")
	wantHover(t, crueltyGo, "Cruelty.Rank", 3,
		"### warrior/Cruelty\n",
		"| 12320 | Cruelty | rank 3 of 5 | `Rank(3)` | effect 1 = 3 |",
		"### 12320 Cruelty · rank 5 of 5\n")
}

const bloodrageGo = `package warrior

var instantRage = spellData.Bloodrage.EffectAt(1).TenthsAt(1)
`

// A value read off a ladder hovers as the rank it read, like one read off a rank.
func TestHoverLadderAccessor(t *testing.T) {
	wantHover(t, bloodrageGo, "instantRage =", 4,
		"`instantRage` = **10**\n\n`spellData.Bloodrage.EffectAt(1).TenthsAt(1)`",
		"### 2687 Bloodrage\n",
		"| 1 ▶ | restores 10 rage to the caster<br>")
	wantHover(t, bloodrageGo, "TenthsAt(1)", 2, "`TenthsAt(1)` = **10**")
}

// Two lines of sim/warrior/battle_shout.go, whose finder names its aura by a dbcenums constant.
const battleShoutGo = `package warrior

var battleShoutRank = spellData.BattleShout.Highest()

func (warrior *Warrior) registerBattleShout() {
	baseAttackPower := battleShoutRank.Effect(dbcenums.A_MOD_ATTACK_POWER, 0).AverageAt(core.CharacterLevel)
}
`

func TestHoverEnumArgument(t *testing.T) {
	wantHover(t, battleShoutGo, "baseAttackPower :=", 4,
		"`baseAttackPower` = **139**\n\n`spellData.BattleShout.Highest().Effect(99, 0).AverageAt(60)`",
		"| 1 ▶ | +139 attack power to the party around the caster")
}

// A line of sim/warrior/stances.go with two ladder reads in one expression, then a declaration split
// over lines.
const stancesGo = `package warrior

func (warrior *Warrior) makeStanceSpell() {
	maxRetainedRage := spellData.TacticalMastery.ValueAt(1) + spellData.ImprovedTacticalMastery.ValueAt(warrior.Talents.ImprovedTacticalMastery)
	split := spellData.Execute.
		Highest().
		EffectN(1)
}
`

func TestHoverInsideExpression(t *testing.T) {
	wantHover(t, stancesGo, "ValueAt(1)", 2,
		"`ValueAt(1)` = **10**\n\n`spellData.TacticalMastery.ValueAt(1)`",
		"| 1 ▶ | a dummy aura holding 10<br>")
	wantHover(t, stancesGo, "split :=", 2, "`EffectN(1)` = **effect 1** of 20662 Execute (Rank 5)")
	wantHover(t, stancesGo, "\t\tEffectN(1)", 4, "`EffectN(1)` = **effect 1** of 20662 Execute (Rank 5)")
}

func TestHoverSegments(t *testing.T) {
	wantHover(t, executeGo, "EffectN(1).Average", 2,
		"`EffectN(1)` = **effect 1** of 20662 Execute (Rank 5)",
		"`spellData.Execute.Highest().EffectN(1)`",
		"---\n| # | effect |\n|--:|:--|\n| 1 ▶ | a dummy effect holding 600 to the enemy<br>`E_DUMMY base=600 sp=1 target=[6,0]` |\n",
		"`AverageAt(60) = 600`  \n")

	markdown := wantHover(t, executeGo, "AverageAt(core", 2,
		"`AverageAt(60)` = **600**",
		"`spellData.Execute.Highest().EffectN(1).AverageAt(60)`",
		"| 1 ▶ |")
	// The editor's Go server shows the accessor's doc beside this hover already.
	if strings.Contains(markdown, "The amount at a caster level") {
		t.Errorf("the call's hover repeats the accessor's doc:\n%s", markdown)
	}
	if again := wantHover(t, executeGo, "core.CharacterLevel", 2); again != markdown {
		t.Errorf("the argument hovers differently from its call:\n%s", again)
	}

	wantHover(t, executeGo, "executeRank.EffectN(1).Average", 4, executeCard)
	wantHover(t, executeGo, "executeRank.EffectN(1).ChainAmp", 4, executeCard)
	wantHover(t, executeGo, "ChainAmp)", 2, "`ChainAmp` = **1.5**", "| 1 ▶ |")
}

func TestHoverTrace(t *testing.T) {
	_, trace, _ := hoverOn(t, newWorkspace(), executeGo, warriorURI(t, "execute.go"), "executeBaseDamage =", 4)
	want := []string{
		"sim/warrior/execute.go:5:9",
		"declarations sim/warrior: cache miss",
		"ident executeBaseDamage",
		"  executeBaseDamage = executeRank.EffectN(1).AverageAt(core.CharacterLevel)  (sim/warrior/execute.go:5)",
		"  executeRank = spellData.Execute.Highest()  (sim/warrior/execute.go:3)",
		"expr spellData.Execute.Highest().EffectN(1).AverageAt(core.CharacterLevel)",
		`render hover: name="executeBaseDamage" rank=false segments=3`,
		"✓ 20662 Execute (Rank 5) → effect 1 → 600",
	}
	if len(trace) != len(want) {
		t.Fatalf("the trace is\n%s", strings.Join(trace, "\n"))
	}
	for i := range want {
		if !strings.HasPrefix(trace[i], want[i]) {
			t.Errorf("trace line %d is %q, want %q", i, trace[i], want[i])
		}
	}
}

func TestHoverSilent(t *testing.T) {
	cases := []struct {
		needle string
		reason string
	}{
		{"extraRage", "✗ no ladder-shaped declaration of extraRage in sim/warrior"},
		{"executeDamagePerRage =", "✗ executeDamagePerRage is bound to no chain the evaluator reads"},
		{"registerExecute", "✗ no ladder-shaped declaration of registerExecute"},
		{"func", "✗ no spell id, family or name under the cursor"},
	}
	for _, c := range cases {
		markdown, trace, ok := hoverOn(t, newWorkspace(), executeGo, warriorURI(t, "execute.go"), c.needle, 1)
		if ok {
			t.Errorf("%s answered\n%s", c.needle, markdown)
		}
		if last := trace[len(trace)-1]; !strings.HasPrefix(last, c.reason) {
			t.Errorf("%s stopped with %q, want %q", c.needle, last, c.reason)
		}
	}

	if _, _, ok := hoverOn(t, newWorkspace(), executeGo, "file:///tmp/execute.ts", "executeRank =", 2); ok {
		t.Error("a TS file answered a Go name")
	}
}

func TestDeclarations(t *testing.T) {
	file := strings.Join([]string{
		"package warrior",
		"",
		"func f() {",
		"\tsecondRank := spellData.Execute.Rank(2)",
		"\tnamed := spellData.Execute.ByID(20658)",
		"\tbare := Cruelty.Rank(3)",
		"\tif executeRank == spellData.Execute.Highest() {",
		"\t}",
		"// var stubbedRank = spellData.MangleBear.ByID(33987)",
		"\tconfig := spelldata.SpellConfig(&warrior.Unit, executeRank, spelldata.Melee(cost))",
		"\tmaxRage := warrior.MaximumRage() - spell.Cost.GetCurrentCost()",
		"\tticks := executeRank.EffectN(1).AverageAt(level)",
		"\tsplit := spellData.Execute.",
		"\t\tHighest().",
		"\t\tEffectN(1)",
		"}",
		"",
		"var executeRank = spellData.Execute.Highest() // the top rank",
	}, "\n")

	found := map[string]string{}
	for _, d := range parseGo("a.go", file).decls {
		found[d.name] = d.chain.text(false)
	}
	want := map[string]string{
		"secondRank":  "spellData.Execute.Rank(2)",
		"named":       "spellData.Execute.ByID(20658)",
		"bare":        "Cruelty.Rank(3)",
		"split":       "spellData.Execute.Highest().EffectN(1)",
		"executeRank": "spellData.Execute.Highest()",
	}
	if len(found) != len(want) {
		t.Errorf("declared %v, want %v", found, want)
	}
	for name, chain := range want {
		if found[name] != chain {
			t.Errorf("%s is declared as %q, want %q", name, found[name], chain)
		}
	}
}

func TestEachParameterDeclarations(t *testing.T) {
	file := strings.Join([]string{
		"package priest",
		"",
		"func registerHolyFireSpell(rank *spelldata.Spell) {",
		"\trank.DamageEffect()",
		"}",
		"",
		"func registerVampiricEmbrace(rank *spelldata.Spell, timer int) {",
		"\trank.DamageEffect()",
		"}",
		"",
		"func register() {",
		"\tspellData.HolyFire.Each(func(_ int32, rank *spelldata.Spell) {",
		"\t\tregisterHolyFireSpell(rank)",
		"\t})",
		"\tspellData.VampiricEmbrace.Each(func(_ int32, rank *spelldata.Spell) {",
		"\t\tregisterVampiricEmbrace(rank, 1)",
		"\t})",
		"}",
	}, "\n")

	parsed := parseGo("priest.go", file)

	if len(parsed.parameterSources) != 2 {
		t.Fatalf("got %d parameter sources, want 2", len(parsed.parameterSources))
	}

	for _, source := range parsed.parameterSources {
		if got := source.chain.text(false); got != "spellData.HolyFire" &&
			got != "spellData.VampiricEmbrace" {
			t.Errorf("unexpected source chain %q", got)
		}
	}
}

func TestVirtualRankPreservesAccessorChain(t *testing.T) {
	file := strings.Join([]string{
		"package priest",
		"",
		"func registerHolyFireSpell(rank *spelldata.Spell) {",
		"\trank.EffectN(1).Average()",
		"}",
		"",
		"func register() {",
		"\tspellData.HolyFire.Each(func(_ int32, rank *spelldata.Spell) {",
		"\t\tregisterHolyFireSpell(rank)",
		"\t})",
		"}",
	}, "\n")

	parsed := parseGo("priest.go", file)
	if len(parsed.parameterSources) != 1 {
		t.Fatalf("got %d parameter sources, want 1", len(parsed.parameterSources))
	}

	source := parsed.parameterSources[0]
	sourceDeclaration, ok := parameterDeclaration(parsed, source)
	if !ok {
		t.Fatal("rank parameter was not declared")
	}

	declarations := map[string]declaration{
		sourceDeclaration.name: sourceDeclaration,
	}

	chain, err := parseChain("rank.EffectN(1).Average()")
	if err != nil {
		t.Fatal(err)
	}

	hover, ok := resolved(chain, "rank", declarations, &tracer{})
	if !ok {
		t.Fatal("rank chain did not resolve")
	}

	if hover.parameter == nil {
		t.Fatal("rank chain lost its virtual parameter")
	}

	if hover.rank == nil {
		t.Fatal("rank chain did not produce a virtual rank")
	}

	if got := hover.chain.text(false); got != "rank.EffectN(1).Average()" {
		t.Errorf("resolved chain became %q, want rank.EffectN(1).Average()", got)
	}
}

func declared(chains map[string]string) map[string]declaration {
	out := map[string]declaration{}
	for name, text := range chains {
		c, err := parseChain(text)
		if err != nil {
			panic(err)
		}
		out[name] = declaration{name: name, chain: c, file: "a.go", line: 1}
	}
	return out
}

func resolveText(text string, decls map[string]declaration) (string, error) {
	c, err := parseChain(text)
	if err != nil {
		return "", err
	}
	resolved, err := resolveChain(c, decls, &tracer{})
	if err != nil {
		return "", err
	}
	return resolved.text(false), nil
}

func TestResolveChain(t *testing.T) {
	decls := declared(map[string]string{
		"executeRank":       "spellData.Execute.Highest()",
		"executeBaseDamage": "executeRank.EffectN(1).AverageAt(core.CharacterLevel)",
	})
	resolves := map[string]string{
		"executeBaseDamage":      "spellData.Execute.Highest().EffectN(1).AverageAt(core.CharacterLevel)",
		"executeRank":            "spellData.Execute.Highest()",
		"executeRank.EffectN(1)": "spellData.Execute.Highest().EffectN(1)",
		"Cruelty.Rank(3)":        "spellData.Cruelty.Rank(3)",
		"spellData.Rend.Highest().EffectN(1).Period()": "spellData.Rend.Highest().EffectN(1).Period()",
	}
	for chain, want := range resolves {
		got, err := resolveText(chain, decls)
		if err != nil || got != want {
			t.Errorf("%s resolved to %q, %v; want %q", chain, got, err, want)
		}
	}

	refuses := map[string]string{
		"warrior.MaximumRage()":  "no ladder-shaped declaration of warrior",
		"executeRank.EffectN(n)": "n is not a literal",
	}
	for chain, want := range refuses {
		if got, err := resolveText(chain, decls); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s resolved to %q, %v; want %q", chain, got, err, want)
		}
	}
}

func TestResolveChainGuards(t *testing.T) {
	deep := declared(map[string]string{
		"a": "spellData.Execute.Highest()",
		"b": "a.EffectN(1)",
		"c": "b.AverageAt(60)",
		"d": "c.MinAt(60)",
		"e": "d.MinAt(60)",
	})
	if got, err := resolveText("c", deep); got != "spellData.Execute.Highest().EffectN(1).AverageAt(60)" {
		t.Errorf("c resolved to %q, %v", got, err)
	}
	if _, err := resolveText("e", deep); err == nil || !strings.Contains(err.Error(), "more than 4 names") {
		t.Errorf("e resolved past the depth limit: %v", err)
	}

	cyclic := declared(map[string]string{
		"loop":  "other.EffectN(1)",
		"other": "loop.EffectN(2)",
		"self":  "self.EffectN(1)",
	})
	for _, name := range []string{"loop", "self"} {
		if _, err := resolveText(name, cyclic); err == nil || !strings.Contains(err.Error(), "stands on itself") {
			t.Errorf("%s resolved through a cycle: %v", name, err)
		}
	}
}

func TestWorkspaceBuffers(t *testing.T) {
	ws := newWorkspace()
	other := warriorURI(t, "zz_hover_test_buffer.go")
	ws.update(other, "package warrior\n\nvar hoverTestRank = spellData.Rend.Highest()\n", true)

	use := "package warrior\n\nvar x = hoverTestRank\n"
	uri := warriorURI(t, "execute.go")
	markdown, trace, ok := hoverOn(t, ws, use, uri, "hoverTestRank", 2)
	if !ok || !strings.HasPrefix(markdown, "### 11574 Rend · Rank 7\n") {
		t.Fatalf("the buffer's declaration was not read:\n%s\n%s", markdown, strings.Join(trace, "\n"))
	}

	_, trace, _ = hoverOn(t, ws, use, uri, "hoverTestRank", 2)
	if !strings.HasSuffix(trace[1], "cache hit") {
		t.Errorf("the second hover missed the cache: %v", trace)
	}

	ws.update(other, "package warrior\n\nvar hoverTestRank = spellData.Execute.Highest()\n", true)
	markdown, trace, _ = hoverOn(t, ws, use, uri, "hoverTestRank", 2)
	if !strings.HasPrefix(trace[1], "declarations sim/warrior: cache miss") || !strings.HasPrefix(markdown, "### 20662 Execute") {
		t.Errorf("the edit did not reach the hover:\n%s\n%s", markdown, strings.Join(trace, "\n"))
	}

	ws.update(other, "", false)
	if _, _, ok := hoverOn(t, ws, use, uri, "hoverTestRank", 2); ok {
		t.Error("a closed buffer's declaration is still read")
	}
}

const scopedGo = `package warrior

func (warrior *Warrior) registerRend() {
	tick := spellData.Rend.Highest().PeriodicEffect()
	rendUse(tick)
}

func (warrior *Warrior) registerBloodCraze() {
	earlyUse(tick)
	tick := spellData.BloodCrazeTriggered.Highest().PeriodicEffect()
	lateUse(tick)
}
`

// A name bound inside a function answers only there and only after its binding, whatever another
// function or file binds to the same name.
func TestHoverLocalScope(t *testing.T) {
	uri := warriorURI(t, "zz_scope_test.go")
	wantHover(t, scopedGo, "rendUse(tick", 8, "of 11574 Rend")
	wantHover(t, scopedGo, "lateUse(tick", 8, "of 16488 Blood Craze")
	if markdown, _, ok := hoverOn(t, newWorkspace(), scopedGo, uri, "earlyUse(tick", 9); ok {
		t.Errorf("a name used before its binding answered:\n%s", markdown)
	}
}

func TestHoverWithoutPackageClause(t *testing.T) {
	if _, _, ok := newWorkspace().hover("x := spellData.Execute.Highest()\n", 0, 2, "file:///tmp/a.go", nil); ok {
		t.Error("a file with no package clause answered a hover")
	}
}

// A file the editor does not hold is read again when it changes on disk, and dropped when it is gone.
func TestWorkspaceDiskChanges(t *testing.T) {
	folder := filepath.Join(t.TempDir(), "warrior")
	if err := os.Mkdir(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(folder, "other.go")
	write := func(family string, stamp time.Time) {
		t.Helper()
		if err := os.WriteFile(other, []byte("package warrior\n\nvar diskRank = spellData."+family+".Highest()\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(other, stamp, stamp); err != nil {
			t.Fatal(err)
		}
	}

	ws := newWorkspace()
	use := "package warrior\n\nvar x = diskRank\n"
	uri := pathURI(filepath.Join(folder, "use.go"))
	write("Rend", time.Unix(1, 0))
	if markdown, trace, _ := hoverOn(t, ws, use, uri, "diskRank", 2); !strings.HasPrefix(markdown, "### 11574 Rend") {
		t.Fatalf("the file on disk was not read:\n%s\n%s", markdown, strings.Join(trace, "\n"))
	}

	write("Execute", time.Unix(2, 0))
	markdown, trace, _ := hoverOn(t, ws, use, uri, "diskRank", 2)
	if !strings.HasPrefix(markdown, "### 20662 Execute") || !strings.HasPrefix(trace[1], "declarations") || !strings.Contains(trace[1], "cache miss, read 1 files") {
		t.Errorf("the change on disk did not reach the hover:\n%s\n%s", markdown, strings.Join(trace, "\n"))
	}

	if err := os.Remove(other); err != nil {
		t.Fatal(err)
	}
	if _, _, ok := hoverOn(t, ws, use, uri, "diskRank", 2); ok {
		t.Error("a file removed from disk is still read")
	}
}

func frame(t *testing.T, msg map[string]any) string {
	t.Helper()
	body, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), body)
}

func readFrames(t *testing.T, out io.Reader) []rpcMessage {
	t.Helper()
	reader := textproto.NewReader(bufio.NewReader(out))
	var msgs []rpcMessage
	for {
		body, err := readFrame(reader)
		if err == io.EOF {
			return msgs
		}
		if err != nil {
			t.Fatal(err)
		}
		var msg rpcMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			t.Fatal(err)
		}
		msgs = append(msgs, msg)
	}
}

func TestLSPRoundTrip(t *testing.T) {
	uri := warriorURI(t, "execute.go")
	line, col := positionOf(t, executeGo, "executeBaseDamage =", 4)
	position := map[string]any{"textDocument": map[string]any{"uri": uri}, "position": map[string]any{"line": line, "character": col}}

	var in strings.Builder
	for _, msg := range []map[string]any{
		{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{"initializationOptions": map[string]any{"trace": "on"}}},
		{"jsonrpc": "2.0", "method": "initialized", "params": map[string]any{}},
		{"jsonrpc": "2.0", "method": "textDocument/didOpen", "params": map[string]any{
			"textDocument": map[string]any{"uri": uri, "languageId": "go", "version": 1, "text": executeGo},
		}},
		{"jsonrpc": "2.0", "id": 2, "method": "textDocument/hover", "params": position},
		{"jsonrpc": "2.0", "id": 3, "method": "textDocument/definition", "params": position},
		{"jsonrpc": "2.0", "method": "$/cancelRequest", "params": map[string]any{"id": 9}},
		{"jsonrpc": "2.0", "id": 4, "method": "shutdown"},
		{"jsonrpc": "2.0", "method": "exit"},
	} {
		in.WriteString(frame(t, msg))
	}

	var out bytes.Buffer
	shutdown, err := serveLSP(strings.NewReader(in.String()), &out)
	if err != nil || !shutdown {
		t.Fatalf("the server stopped with %v, shutdown %v", err, shutdown)
	}

	responses := map[string]rpcMessage{}
	var logs []string
	for _, msg := range readFrames(t, &out) {
		if msg.Method == "window/logMessage" {
			var params struct {
				Type    int    `json:"type"`
				Message string `json:"message"`
			}
			_ = json.Unmarshal(msg.Params, &params)
			if params.Type != messageTypeLog {
				t.Errorf("a log message of type %d", params.Type)
			}
			logs = append(logs, params.Message)
			continue
		}
		responses[string(msg.ID)] = msg
	}

	var initialized struct {
		Capabilities struct {
			HoverProvider    bool `json:"hoverProvider"`
			TextDocumentSync struct {
				Change int `json:"change"`
			} `json:"textDocumentSync"`
		} `json:"capabilities"`
	}
	if err := json.Unmarshal(responses["1"].Result, &initialized); err != nil || !initialized.Capabilities.HoverProvider || initialized.Capabilities.TextDocumentSync.Change != 2 {
		t.Errorf("initialize answered %s", responses["1"].Result)
	}

	var hover struct {
		Contents struct {
			Kind  string `json:"kind"`
			Value string `json:"value"`
		} `json:"contents"`
	}
	if err := json.Unmarshal(responses["2"].Result, &hover); err != nil {
		t.Fatalf("hover answered %s", responses["2"].Result)
	}
	if hover.Contents.Kind != "markdown" || !strings.HasPrefix(hover.Contents.Value, "`executeBaseDamage` = **600**") ||
		!strings.Contains(hover.Contents.Value, "| 1 ▶ |") {
		t.Errorf("hover answered %s:\n%s", hover.Contents.Kind, hover.Contents.Value)
	}

	if responses["3"].Error == nil || responses["3"].Error.Code != codeMethodNotFound {
		t.Errorf("an unknown method answered %+v", responses["3"])
	}
	if string(responses["4"].Result) != "null" || responses["4"].Error != nil {
		t.Errorf("shutdown answered %+v", responses["4"])
	}
	if len(logs) != 1 || !strings.Contains(logs[0], "✓ 20662 Execute (Rank 5) → effect 1 → 600") {
		t.Errorf("the trace logged %q", logs)
	}
}

// Ranged edits land where their UTF-16 positions say, the dash before the id being one unit and three
// bytes.
func TestLSPIncrementalSync(t *testing.T) {
	uri := warriorURI(t, "execute.go")
	text := "package warrior\n\n// — spelldata.MustFind(11574)\nvar executeRank = spellData.Execute.Highest()\n"
	edit := func(line int, lineText, old, replacement string) map[string]any {
		at := len(utf16.Encode([]rune(lineText[:strings.Index(lineText, old)])))
		return map[string]any{
			"range": map[string]any{
				"start": map[string]any{"line": line, "character": at},
				"end":   map[string]any{"line": line, "character": at + len(old)},
			},
			"text": replacement,
		}
	}
	hover := func(id, line, col int) map[string]any {
		return map[string]any{"jsonrpc": "2.0", "id": id, "method": "textDocument/hover", "params": map[string]any{
			"textDocument": map[string]any{"uri": uri}, "position": map[string]any{"line": line, "character": col},
		}}
	}

	var in strings.Builder
	for _, msg := range []map[string]any{
		{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{"initializationOptions": map[string]any{"trace": "off"}}},
		{"jsonrpc": "2.0", "method": "textDocument/didOpen", "params": map[string]any{
			"textDocument": map[string]any{"uri": uri, "languageId": "go", "version": 1, "text": text},
		}},
		{"jsonrpc": "2.0", "method": "textDocument/didChange", "params": map[string]any{
			"textDocument": map[string]any{"uri": uri, "version": 2},
			"contentChanges": []any{
				edit(2, "// — spelldata.MustFind(11574)", "11574", "11572"),
				edit(3, "var executeRank = spellData.Execute.Highest()", "Highest()", "Rank(2)"),
			},
		}},
		hover(2, 2, 26),
		hover(3, 3, 6),
		{"jsonrpc": "2.0", "id": 4, "method": "shutdown"},
		{"jsonrpc": "2.0", "method": "exit"},
	} {
		in.WriteString(frame(t, msg))
	}

	var out bytes.Buffer
	if _, err := serveLSP(strings.NewReader(in.String()), &out); err != nil {
		t.Fatal(err)
	}
	wants := map[string]string{"2": "### 11572 Rend · Rank 5\n", "3": "### 20658 Execute · Rank 2\n"}
	for _, msg := range readFrames(t, &out) {
		want, ok := wants[string(msg.ID)]
		if !ok {
			continue
		}
		var hover struct {
			Contents struct {
				Value string `json:"value"`
			} `json:"contents"`
		}
		if err := json.Unmarshal(msg.Result, &hover); err != nil || !strings.HasPrefix(hover.Contents.Value, want) {
			t.Errorf("hover %s answered %s, want %q", msg.ID, msg.Result, want)
		}
		delete(wants, string(msg.ID))
	}
	if len(wants) > 0 {
		t.Errorf("no answer to hovers %v", wants)
	}
}

func TestLSPTraceOff(t *testing.T) {
	uri := warriorURI(t, "execute.go")
	var in strings.Builder
	in.WriteString(frame(t, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{"initializationOptions": map[string]any{"trace": "off"}}}))
	in.WriteString(frame(t, map[string]any{"jsonrpc": "2.0", "id": 2, "method": "textDocument/hover", "params": map[string]any{
		"textDocument": map[string]any{"uri": uri}, "position": map[string]any{"line": 7, "character": 6},
	}}))

	var out bytes.Buffer
	if shutdown, err := serveLSP(strings.NewReader(in.String()), &out); err != nil || shutdown {
		t.Fatalf("the server stopped with %v, shutdown %v", err, shutdown)
	}
	for _, msg := range readFrames(t, &out) {
		if msg.Method == "window/logMessage" {
			t.Errorf("trace off still logged %s", msg.Params)
		}
	}
}

func TestLSPInvocation(t *testing.T) {
	for _, args := range [][]string{{"-lsp"}, {"--lsp"}, {"-lsp", "--stdio"}} {
		if opts, err := parseArgs(args); err != nil || opts.mode != "lsp" {
			t.Errorf("%q should start the language server: %v", args, err)
		}
	}
	for _, args := range [][]string{{}, {"11574"}, {"-lsp", "11574"}, {"--stdio", "-lsp"}} {
		if opts, err := parseArgs(args); err == nil && opts.mode == "lsp" {
			t.Errorf("%q should not start the language server", args)
		}
	}
}

// Item and set-bonus spells are named by id rather than through a ladder.
const itemsGo = `package warrior

var mightBlockValue = spelldata.MustFind(23562)

func apply() {
	setBonusAura.AttachStatBuff(stats.BlockValue, mightBlockValue.EffectN(1).BaseValue())
}
`

func TestHoverRowByID(t *testing.T) {
	wantHover(t, itemsGo, "mightBlockValue =", 4, "### 23562 Block Value 30\n")
	wantHover(t, itemsGo, "EffectN(1).Base", 2,
		"`EffectN(1)` = **effect 1** of 23562 Block Value 30",
		"`spelldata.MustFind(23562).EffectN(1)`")
	wantHover(t, itemsGo, "BaseValue()", 2,
		"`BaseValue()` = **30**",
		"`spelldata.MustFind(23562).EffectN(1).BaseValue()`",
		"| 1 ▶ |")
}

func TestVirtualRankAccessorRendering(t *testing.T) {
	file := strings.Join([]string{
		"package priest",
		"",
		"func registerDarkSacrificeSpell(rank *spelldata.Spell) {",
		"\trank.Duration()",
		"\trank.EnergizeEffect()",
		"\trank.EnergizeEffect().Average()",
		"}",
		"",
		"func register() {",
		"\tspellData.DarkSacrifice.Each(func(_ int32, rank *spelldata.Spell) {",
		"\t\tregisterDarkSacrificeSpell(rank)",
		"\t})",
		"}",
	}, "\n")

	parsed := parseGo("dark_sacrifice.go", file)

	if len(parsed.parameterSources) != 1 {
		t.Fatalf("got %d parameter sources, want 1", len(parsed.parameterSources))
	}

	source := parsed.parameterSources[0]
	sourceDeclaration, ok := parameterDeclaration(parsed, source)
	if !ok {
		t.Fatal("rank parameter was not declared")
	}

	declarations := map[string]declaration{
		sourceDeclaration.name: sourceDeclaration,
	}

	tests := []struct {
		name string
		expr string
		want string
	}{
		{
			name: "duration",
			expr: "rank.Duration()",
			want: "`Duration()` = **15s**",
		},
		{
			name: "energize effect",
			expr: "rank.EnergizeEffect()",
			want: "`EnergizeEffect()` = **no effect**",
		},
		{
			name: "average",
			expr: "rank.EnergizeEffect().Average()",
			want: "`Average()` = **0**",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain, err := parseChain(tt.expr)
			if err != nil {
				t.Fatal(err)
			}

			hover, ok := resolved(chain, "rank", declarations, &tracer{})
			if !ok {
				t.Fatal("rank chain did not resolve")
			}

			if hover.rank == nil {
				t.Fatal("rank chain did not produce a virtual rank")
			}

			hover.rank.rank = 5

			result, err := evalExpr(ladderFamilies(), hover.chain, "priest", hover.rank)
			if err != nil {
				t.Fatal(err)
			}

			markdown := exprMarkdown(result, hover)

			if !strings.Contains(markdown, tt.want) {
				t.Errorf("hover does not contain %q:\n%s", tt.want, markdown)
			}

			if !strings.Contains(markdown, "### 1277328 Dark Sacrifice (Rank 5)") {
				t.Errorf("hover does not contain resolved spell/rank heading:\n%s", markdown)
			}

			if strings.Contains(markdown, "| **school** |") {
				t.Errorf("virtual-rank accessor hover rendered the full spell card:\n%s", markdown)
			}

			if strings.Contains(markdown, "`rank.EnergizeEffect()`") {
				t.Errorf("virtual-rank accessor hover rendered the evaluator trail:\n%s", markdown)
			}

			if strings.Contains(markdown, "`rank.EnergizeEffect().Average()`") {
				t.Errorf("virtual-rank accessor hover rendered the full expression:\n%s", markdown)
			}
		})
	}
}

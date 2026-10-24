package printer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf16"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/siml/lsp/jsonrpc"
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
		"| 20662 | Execute | Rank 5 | `Highest()` | "+rankValues(t, "warrior/Execute")(5)+" |",
		"### 20662 Execute · Rank 5\n")
	if !strings.Contains(markdown, "| **school** | physical |") {
		t.Errorf("the family's highest rank states no header table:\n%s", markdown)
	}
}

// Execute's damage effect, the one executeGo reads: what its AverageAt(60) answers, and the amount
// its wording holds.
func executeDamage(t *testing.T) (average, holding string) {
	t.Helper()
	damage := ladderOf(t, "warrior/Execute").Highest().EffectN(1)
	return column(damage.AverageAt(60)), column(damage.Mean())
}

// The card a rank pick hovers as: a heading, the ladder call, the row's columns under its name, and the
// effects below a rule with each wording over its game files row.
func executeCard(t *testing.T) string {
	t.Helper()
	execute := ladderOf(t, "warrior/Execute").Highest()
	return "### 20662 Execute · Rank 5\n" +
		"`warrior spellData.Execute.Highest()`  \n" +
		"\n" +
		"| | Execute (Rank 5) |\n" +
		"|--:|:--|\n" +
		"| **school** | physical |\n" +
		"| **defense** | melee |\n" +
		"| **gcd** | " + column(float64(execute.GCDMs)/1000) + " s |\n" +
		"| **cost** | " + column(execute.PowerCost(dbcenums.POWER_RAGE)) + " rage |\n" +
		"| **range** | " + column(float64(execute.MaxRange)) + " yd |\n" +
		"| **stance** | Battle, Berserker |\n" +
		"| **equip** | needs a melee weapon in the main hand |\n" +
		"| **attrs** | refund on miss |\n" +
		"| **labels** | 25 |\n" +
		"\n" +
		"---\n" +
		"\n**1** · a dummy effect holding " + column(execute.EffectN(1).Mean()) + " to the enemy  \n" +
		"`E_DUMMY` `base=" + column(execute.EffectN(1).BasePoints) + "` `sp=1` `target=[6,0]`\n" +
		"\n**2** · casts 26651 Execute to the caster  \n`E_TRIGGER_SPELL` `base=0` `trigger=26651` `target=[1,0]`\n" +
		"\n" +
		"[Wowhead](https://www.wowhead.com/forever/spell=20662)\n"
}

func TestHoverRankPick(t *testing.T) {
	if got, want := wantHover(t, executeGo, "executeRank =", 4), executeCard(t); got != want {
		t.Errorf("the rank pick hovers as\n%s\nwant\n%s", got, want)
	}
}

func TestHoverValue(t *testing.T) {
	average, holding := executeDamage(t)
	markdown := wantHover(t, executeGo, "executeBaseDamage =", 4,
		"`executeBaseDamage` = **"+average+"**\n\n`spellData.Execute.Highest().EffectN(1).AverageAt(60)`",
		"`spellData.Execute.Highest().EffectN(1).AverageAt(60)`\n\n### 20662 Execute · Rank 5\n",
		"**1 ▶** · a dummy effect holding "+holding+" to the enemy  \n",
		"**2** · casts 26651 Execute to the caster  \n")
	if strings.Contains(markdown, "The amount at a caster level") {
		t.Errorf("the name's hover states the accessor's doc:\n%s", markdown)
	}

	wantHover(t, executeGo, "executeBaseDamage +", 4, "`executeBaseDamage` = **"+average+"**")
}

const crueltyGo = `package warrior

func (warrior *Warrior) applyCruelty() {
	crueltyRank := spellData.Cruelty.Rank(3)
	crueltyCrit := crueltyRank.EffectN(1).BaseValue()
}
`

// A talent rank hovers as the rank Talent builds from the curve, titled with its rank.
func TestHoverTalentRank(t *testing.T) {
	crit := ladderOf(t, "warrior/Cruelty").Rank(3).EffectN(1)
	wording := signedColumn(crit.Mean()) + "% physical crit, per weapon: only attacks made with a melee weapon  \n"
	wantHover(t, crueltyGo, "crueltyRank :=", 4,
		"### 12320 Cruelty · rank 3 of 5\n",
		"| | Cruelty (rank 3 of 5) |",
		"**1** · "+wording+"`E_APPLY_AURA` `A_MOD_WEAPON_CRIT_PERCENT` `base="+column(crit.BasePoints)+"` `target=[1,0]`\n")
	wantHover(t, crueltyGo, "crueltyCrit :=", 4,
		"`crueltyCrit` = **"+column(crit.BaseValue())+"**\n\n`spellData.Cruelty.Rank(3).EffectN(1).BaseValue()`",
		"### 12320 Cruelty · rank 3 of 5\n",
		"**1 ▶** · "+wording)
	wantHover(t, crueltyGo, "Cruelty.Rank", 3,
		"### warrior/Cruelty\n",
		"| 12320 | Cruelty | rank 3 of 5 | `Rank(3)` | "+rankValues(t, "warrior/Cruelty")(3)+" |",
		"### 12320 Cruelty · rank 5 of 5\n")
}

const bloodrageGo = `package warrior

var instantRage = spellData.Bloodrage.EffectAt(1).TenthsAt(1)
`

// A value read off a ladder hovers as the rank it read, like one read off a rank.
func TestHoverLadderAccessor(t *testing.T) {
	bloodrage := ladderOf(t, "warrior/Bloodrage")
	rage := column(bloodrage.EffectAt(1).TenthsAt(1))
	wantHover(t, bloodrageGo, "instantRage =", 4,
		"`instantRage` = **"+rage+"**\n\n`spellData.Bloodrage.EffectAt(1).TenthsAt(1)`",
		"### 2687 Bloodrage\n",
		"**1 ▶** · restores "+column(bloodrage.Rank(1).EffectN(1).Mean()/10)+" rage to the caster  \n")
	wantHover(t, bloodrageGo, "TenthsAt(1)", 2, "`TenthsAt(1)` = **"+rage+"**")
}

// Two lines of sim/warrior/battle_shout.go, whose finder names its aura by a dbcenums constant.
const battleShoutGo = `package warrior

var battleShoutRank = spellData.BattleShout.Highest()

func (warrior *Warrior) registerBattleShout() {
	baseAttackPower := battleShoutRank.Effect(dbcenums.A_MOD_ATTACK_POWER, 0).AverageAt(core.CharacterLevel)
}
`

func TestHoverEnumArgument(t *testing.T) {
	attackPower := ladderOf(t, "warrior/BattleShout").Highest().Effect(dbcenums.A_MOD_ATTACK_POWER, 0)
	wantHover(t, battleShoutGo, "baseAttackPower :=", 4,
		"`baseAttackPower` = **"+column(attackPower.AverageAt(60))+"**\n\n`spellData.BattleShout.Highest().Effect(99, 0).AverageAt(60)`",
		"**1 ▶** · "+signedColumn(attackPower.Mean())+" attack power to the raid around the caster")
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
	tacticalMastery := ladderOf(t, "warrior/TacticalMastery")
	wantHover(t, stancesGo, "ValueAt(1)", 2,
		"`ValueAt(1)` = **"+column(tacticalMastery.ValueAt(1))+"**\n\n`spellData.TacticalMastery.ValueAt(1)`",
		"**1 ▶** · a dummy aura holding "+column(tacticalMastery.Rank(1).EffectN(1).Mean())+"  \n")
	wantHover(t, stancesGo, "split :=", 2, "`EffectN(1)` = **effect 1** of 20662 Execute (Rank 5)")
	wantHover(t, stancesGo, "\t\tEffectN(1)", 4, "`EffectN(1)` = **effect 1** of 20662 Execute (Rank 5)")
}

func TestHoverSegments(t *testing.T) {
	damage := ladderOf(t, "warrior/Execute").Highest().EffectN(1)
	average, holding := executeDamage(t)
	wantHover(t, executeGo, "EffectN(1).Average", 2,
		"`EffectN(1)` = **effect 1** of 20662 Execute (Rank 5)",
		"`spellData.Execute.Highest().EffectN(1)`",
		"---\n\n**1 ▶** · a dummy effect holding "+holding+" to the enemy  \n`E_DUMMY` `base="+column(damage.BasePoints)+"` `sp=1` `target=[6,0]`\n",
		"`AverageAt(60) = "+average+"`  \n")

	markdown := wantHover(t, executeGo, "AverageAt(core", 2,
		"`AverageAt(60)` = **"+average+"**",
		"`spellData.Execute.Highest().EffectN(1).AverageAt(60)`",
		"**1 ▶**")
	// The editor's Go server shows the accessor's doc beside this hover already.
	if strings.Contains(markdown, "The amount at a caster level") {
		t.Errorf("the call's hover repeats the accessor's doc:\n%s", markdown)
	}
	if again := wantHover(t, executeGo, "core.CharacterLevel", 2); again != markdown {
		t.Errorf("the argument hovers differently from its call:\n%s", again)
	}

	wantHover(t, executeGo, "executeRank.EffectN(1).Average", 4, executeCard(t))
	wantHover(t, executeGo, "executeRank.EffectN(1).ChainAmp", 4, executeCard(t))
	wantHover(t, executeGo, "ChainAmp)", 2, "`ChainAmp` = **"+column(float64(damage.ChainAmp))+"**", "**1 ▶**")
}

func TestHoverTrace(t *testing.T) {
	_, trace, _ := hoverOn(t, newWorkspace(), executeGo, warriorURI(t, "execute.go"), "executeBaseDamage =", 4)
	average, _ := executeDamage(t)
	want := []string{
		"sim/warrior/execute.go:5:9",
		"declarations sim/warrior: cache miss",
		"ident executeBaseDamage",
		"  executeBaseDamage = executeRank.EffectN(1).AverageAt(core.CharacterLevel)  (sim/warrior/execute.go:5)",
		"  executeRank = spellData.Execute.Highest()  (sim/warrior/execute.go:3)",
		"expr spellData.Execute.Highest().EffectN(1).AverageAt(core.CharacterLevel)",
		"✓ 20662 Execute (Rank 5) → effect 1 → " + average,
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
	reader := jsonrpc.NewFrameReader(out)
	var msgs []rpcMessage
	for {
		body, err := reader.Read()
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
	average, _ := executeDamage(t)
	if hover.Contents.Kind != "markdown" || !strings.HasPrefix(hover.Contents.Value, "`executeBaseDamage` = **"+average+"**") ||
		!strings.Contains(hover.Contents.Value, "**1 ▶**") {
		t.Errorf("hover answered %s:\n%s", hover.Contents.Kind, hover.Contents.Value)
	}

	if responses["3"].Error == nil || responses["3"].Error.Code != codeMethodNotFound {
		t.Errorf("an unknown method answered %+v", responses["3"])
	}
	if string(responses["4"].Result) != "null" || responses["4"].Error != nil {
		t.Errorf("shutdown answered %+v", responses["4"])
	}
	if len(logs) != 1 || !strings.Contains(logs[0], "✓ 20662 Execute (Rank 5) → effect 1 → "+average) {
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

const darkSacrificeGo = `package priest

func registerDarkSacrificeSpell(rank *spelldata.Spell) {
	config := spelldata.SpellConfig(&priest.Unit, rank)
	rank.EffectN(1).Average()
}

func register() {
	spellData.DarkSacrifice.Each(func(_ int32, rank *spelldata.Spell) {
		registerDarkSacrificeSpell(rank)
	})
}
`

// The answers of a session that opens darkSacrificeGo and sends each message in turn, by request id.
func rankSession(t *testing.T, rankLinks bool, msgs ...map[string]any) (string, map[string]rpcMessage) {
	t.Helper()
	root, err := moduleRoot()
	if err != nil {
		t.Fatal(err)
	}
	uri := pathURI(filepath.Join(root, "sim", "priest", "dark_sacrifice.go"))

	var in strings.Builder
	in.WriteString(frame(t, map[string]any{"jsonrpc": "2.0", "id": 0, "method": "initialize", "params": map[string]any{
		"initializationOptions": map[string]any{"trace": "off", "rankLinks": rankLinks},
	}}))
	in.WriteString(frame(t, map[string]any{"jsonrpc": "2.0", "method": "textDocument/didOpen", "params": map[string]any{
		"textDocument": map[string]any{"uri": uri, "languageId": "go", "version": 1, "text": darkSacrificeGo},
	}}))
	for _, msg := range msgs {
		body, err := json.Marshal(msg)
		if err != nil {
			t.Fatal(err)
		}
		var withURI map[string]any
		_ = json.Unmarshal([]byte(strings.ReplaceAll(string(body), "$URI", uri)), &withURI)
		in.WriteString(frame(t, withURI))
	}

	var out bytes.Buffer
	if _, err := serveLSP(strings.NewReader(in.String()), &out); err != nil {
		t.Fatal(err)
	}
	responses := map[string]rpcMessage{}
	for _, msg := range readFrames(t, &out) {
		responses[string(msg.ID)] = msg
	}
	return uri, responses
}

func rankRequest(t *testing.T, id int, method, needle string, offset int, extra map[string]any) map[string]any {
	t.Helper()
	line, col := positionOf(t, darkSacrificeGo, needle, offset)
	at := map[string]any{"line": line, "character": col}
	params := map[string]any{"textDocument": map[string]any{"uri": "$URI"}, "position": at, "range": map[string]any{"start": at, "end": at}}
	for key, value := range extra {
		params[key] = value
	}
	return map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params}
}

func setRank(id int, rank int) map[string]any {
	return map[string]any{"jsonrpc": "2.0", "id": id, "method": "workspace/executeCommand", "params": map[string]any{
		"command":   setRankCommand,
		"arguments": []any{map[string]any{"uri": "$URI", "function": "registerDarkSacrificeSpell", "parameter": "rank", "rank": rank}},
	}}
}

func hoverValue(t *testing.T, msg rpcMessage) string {
	t.Helper()
	var hover struct {
		Contents struct {
			Value string `json:"value"`
		} `json:"contents"`
	}
	if err := json.Unmarshal(msg.Result, &hover); err != nil || msg.Error != nil {
		t.Fatalf("hover answered %s, %+v", msg.Result, msg.Error)
	}
	return hover.Contents.Value
}

// An editor that cannot run a hover's command links picks the rank a Ladder.Each callback parameter
// hovers as through the code actions.
func TestLSPRankPick(t *testing.T) {
	uri, responses := rankSession(t, false,
		rankRequest(t, 1, "textDocument/codeAction", "Spell(rank", 6, map[string]any{"context": map[string]any{"diagnostics": []any{}}}),
		rankRequest(t, 2, "textDocument/codeAction", "Spell(rank", 6, map[string]any{"context": map[string]any{"only": []string{"source.organizeImports"}}}),
		rankRequest(t, 3, "textDocument/codeAction", "package", 2, nil),
		setRank(4, 2),
		rankRequest(t, 5, "textDocument/hover", "Spell(rank", 6, nil),
		rankRequest(t, 6, "textDocument/hover", "EffectN", 2, nil),
		rankRequest(t, 7, "textDocument/hover", "config", 2, nil),
		rankRequest(t, 8, "textDocument/codeAction", "Average", 2, nil),
		map[string]any{"jsonrpc": "2.0", "id": 9, "method": "workspace/executeCommand", "params": map[string]any{"command": "wowsims-spelldata.other"}},
	)

	var initialized struct {
		Capabilities struct {
			CodeActionProvider     bool `json:"codeActionProvider"`
			ExecuteCommandProvider struct {
				Commands []string `json:"commands"`
			} `json:"executeCommandProvider"`
		} `json:"capabilities"`
	}
	if err := json.Unmarshal(responses["0"].Result, &initialized); err != nil || !initialized.Capabilities.CodeActionProvider ||
		fmt.Sprint(initialized.Capabilities.ExecuteCommandProvider.Commands) != "["+setRankCommand+"]" {
		t.Errorf("initialize answered %s", responses["0"].Result)
	}

	type command struct {
		Title     string     `json:"title"`
		Command   string     `json:"command"`
		Arguments []rankPick `json:"arguments"`
	}
	actions := func(id string) []command {
		var commands []command
		if err := json.Unmarshal(responses[id].Result, &commands); err != nil || responses[id].Error != nil {
			t.Fatalf("code actions %s answered %s, %+v", id, responses[id].Result, responses[id].Error)
		}
		return commands
	}
	highest := actions("1")
	if len(highest) != 4 || highest[0].Title != "Hover rank as 1277324 Dark Sacrifice (Rank 1)" || highest[0].Command != setRankCommand ||
		fmt.Sprint(highest[3].Arguments) != fmt.Sprint([]rankPick{{URI: uri, Parameter: "rank", Function: "registerDarkSacrificeSpell", Rank: 4}}) {
		t.Errorf("the code actions on the parameter are %+v", highest)
	}
	if got := actions("2"); len(got) != 0 {
		t.Errorf("a request for kinds got %+v", got)
	}
	if got := actions("3"); len(got) != 0 {
		t.Errorf("a position outside any callback got %+v", got)
	}
	if string(responses["4"].Result) != "null" || responses["4"].Error != nil {
		t.Errorf("%s answered %+v", setRankCommand, responses["4"])
	}

	if value := hoverValue(t, responses["5"]); !strings.HasPrefix(value, "**Rank:** **2** / 5\n\nPick another rank from the code actions.\n\n") ||
		strings.Contains(value, "command:") || !strings.Contains(value, "1277325 Dark Sacrifice") {
		t.Errorf("the hover on the parameter after the pick:\n%s", value)
	}
	if value := hoverValue(t, responses["6"]); !strings.Contains(value, "### 1277325 Dark Sacrifice (Rank 2)") {
		t.Errorf("the hover on an accessor after the pick:\n%s", value)
	}
	if value := hoverValue(t, responses["7"]); !strings.Contains(value, "Dark Sacrifice") {
		t.Errorf("the hover on a SpellConfig declaration read through the parameter:\n%s", value)
	}
	if got := actions("8"); len(got) != 4 || strings.Contains(fmt.Sprint(got), "(Rank 2)") {
		t.Errorf("the code actions after the pick are %+v", got)
	}

	if responses["9"].Error == nil || responses["9"].Error.Code != codeInvalidParams {
		t.Errorf("an unknown command answered %+v", responses["9"])
	}
}

// An editor that runs command:wowsims-spelldata.changeRank links gets one per other rank in the hover.
func TestLSPRankLinks(t *testing.T) {
	_, responses := rankSession(t, true, rankRequest(t, 1, "textDocument/hover", "Spell(rank", 6, nil))
	value := hoverValue(t, responses["1"])
	if !strings.HasPrefix(value, "**Rank:** **5** / 5\n\n[1](command:wowsims-spelldata.changeRank?") || !strings.Contains(value, " **[5]**\n\n") ||
		strings.Contains(value, "code actions") {
		t.Errorf("the hover with rank links:\n%s", value)
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
		"`BaseValue()` = **"+column(spelldata.MustFind(23562).EffectN(1).BaseValue())+"**",
		"`spelldata.MustFind(23562).EffectN(1).BaseValue()`",
		"**1 ▶**")
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
			want: "`Duration()` = **" + ladderOf(t, "priest/DarkSacrifice").Rank(5).Duration().String() + "**",
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

			markdown := exprMarkdown(result, hover, false)

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

const sealOfFuryGo = `package paladin

func (paladin *Paladin) registerSealOfFury(_ int32, rank *spelldata.Spell) {
	judgementRank := spellData.SealOfFuryTriggered.ByID(int32(rank.EffectN(3).BaseValue()))
	judgementDamage := judgementRank.DamageEffect()
	procRank := rank.Refs()[0]
}

func (paladin *Paladin) registerSeals() {
	spellData.SealOfFury.Each(paladin.registerSealOfFury)
}
`

// A chain read through a Ladder.Each callback parameter inside another call's argument, and an index
// into the rows Refs answers.
func TestHoverArgumentChainAndIndex(t *testing.T) {
	root, err := moduleRoot()
	if err != nil {
		t.Fatal(err)
	}
	uri := pathURI(filepath.Join(root, "sim", "paladin", "seal_of_fury.go"))
	hover := func(needle string, offset int, wants ...string) {
		t.Helper()
		markdown, trace, ok := hoverOn(t, newWorkspace(), sealOfFuryGo, uri, needle, offset)
		if !ok {
			t.Fatalf("no hover on %q:\n%s", needle, strings.Join(trace, "\n"))
		}
		for _, want := range wants {
			if !strings.Contains(markdown, want) {
				t.Errorf("the hover on %q lacks %q:\n%s", needle, want, markdown)
			}
		}
	}

	seal := ladderOf(t, "paladin/SealOfFury").Highest()
	judgementID := int32(seal.EffectN(3).BaseValue())
	judgement := spelldata.MustFind(judgementID)
	proc := seal.Refs()[0]

	judgementHeading := fmt.Sprintf("### %d %s · %s\n", judgement.ID, judgement.Name, judgement.Rank)
	hover("judgementRank :=", 2, judgementHeading, fmt.Sprintf("`paladin spellData.SealOfFuryTriggered.ByID(%d)`", judgementID), "| **school** |")
	hover("judgementRank.DamageEffect", 2, judgementHeading)
	hover("judgementRank.DamageEffect", 16, "`DamageEffect()` = **effect 1**", fmt.Sprintf("### %s\n", title(judgement)))
	hover("ByID", 2, fmt.Sprintf("`ByID(int32(rank.EffectN(3).BaseValue()))` = **%d**", judgementID))
	// The argument answers for itself, not for the call it sits in.
	hover("BaseValue", 2, fmt.Sprintf("`BaseValue()` = **%d**", judgementID), fmt.Sprintf("### %s\n", title(seal)))

	hover("procRank :=", 2, fmt.Sprintf("### %d %s · %s\n", proc.ID, proc.Name, proc.Rank), "| **school** |")
	hover("Refs()", 2, "`Refs()` = **[0] "+title(proc)+", [1] ")
	hover("[0]", 1, fmt.Sprintf("`Refs()[0]` = **%d**", proc.ID), fmt.Sprintf("### %s\n", title(proc)))
}

// A name bound through its own argument, which an unsaved buffer can state, ends in an answer.
func TestHoverSelfReferentialArgument(t *testing.T) {
	text := "package paladin\n\nvar loop = spellData.SealOfFuryTriggered.ByID(int32(loop.EffectN(3).BaseValue()))\n" +
		"var ping = pong.EffectN(1)\nvar pong = ping.EffectN(1)\n"
	for _, needle := range []string{"loop =", "ping =", "pong ="} {
		hoverOn(t, newWorkspace(), text, warriorURI(t, "loop.go"), needle, 2)
	}
}

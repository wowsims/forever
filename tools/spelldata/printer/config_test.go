package printer

import (
	"bytes"
	"strconv"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// Every name the printer states for a constant: the stringer's for a typed enum, the declaration's for
// the PROC_FLAG_ bits and SPELLMOD_ ops, and a number where neither has one.
func TestConstantNames(t *testing.T) {
	spellFlag := func(bit uint64) (string, bool) { return dbcenums.Named(core.SpellFlag(bit)) }
	if got := strings.Join(setBits(uint64(core.SpellFlagAPL|core.SpellFlagMeleeMetrics), spellFlag), " | "); got != "SpellFlagMeleeMetrics | SpellFlagAPL" {
		t.Errorf("the flag bits read %q", got)
	}
	if got := setBits(1<<63, spellFlag); len(got) != 1 || got[0] != "bit 0x8000000000000000" {
		t.Errorf("an unnamed flag bit reads %q", got)
	}

	flags := procFlagNames([2]uint32{dbcenums.PROC_FLAG_KILL, dbcenums.PROC_FLAG_2_KNOCKBACK | 1<<3})
	if strings.Join(flags, ", ") != "PROC_FLAG_KILL, PROC_FLAG_2_KNOCKBACK, bit 0x8" {
		t.Errorf("the proc flags read %q", flags)
	}

	for got, want := range map[string]string{
		namedOr(dbcenums.SPELLMOD_COST, "op %d"):        "SPELLMOD_COST",
		namedOr(dbcenums.SpellModOp(99), "op %d"):       "op 99",
		namedOr(dbcenums.E_SCHOOL_DAMAGE, "E_%d"):       "E_SCHOOL_DAMAGE",
		namedOr(dbcenums.SpellEffectType(9999), "E_%d"): "E_9999",
		namedOr(dbcenums.A_DUMMY, "A_%d"):               "A_DUMMY",
		namedOr(dbcenums.EffectAuraType(9999), "A_%d"):  "A_9999",
	} {
		if got != want {
			t.Errorf("named %q, want %q", got, want)
		}
	}
}

const executeConfigGo = `package warrior

var executeRank = spellData.Execute.Highest()

func (warrior *Warrior) registerExecute() {
	config := spelldata.SpellConfig(&warrior.Unit, executeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.Flags(core.SpellFlagNoOnCastComplete))
	config.ClassSpellMask = SpellMaskExecute

	otherConfig := spelldata.SpellConfig(&warrior.Unit, executeRank)
	otherConfig.ClassSpellMask = SpellMaskExecute

	ww := spelldata.SpellConfig(&warrior.Unit, spellData.Whirlwind.Highest(),
		spelldata.Melee(core.ProcMaskMeleeOHSpecial), // the off hand
		spelldata.Proc(), spelldata.Tag(2))

	odd := spelldata.SpellConfig(&warrior.Unit, executeRank, spelldata.Flags(flags), spelldata.Label(3), spelldata.Flags(core.SpellFlagAPL|core.SpellFlagHelpful), spelldata.Tag(1<<40))
}
`

const spellConfigAssignmentsGo = `package priest

var holyFireRank = spellData.HolyFire.Rank(8)

func (priest *Priest) registerHolyFire() {
	config := spelldata.SpellConfig(&priest.Unit, holyFireRank, spelldata.Magic(core.ProcMaskSpellDamage))

	config.Dot = spelldata.DotConfig(holyFireRank, holyFireRank)
	config.Dot.Aura.Tag = "Holy Fire"
	config.Dot.OnTick = spelldata.PeriodicDamageTick(holyFireRank, holyFireRank.TickOutcome)

	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, 100, spell.OutcomeExpectedMagicHitAndCrit)
	}

	config.ExpectedTickDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
		return spell.CalcPeriodicDamage(sim, target, 50, spell.OutcomeExpectedMagicAlwaysHit)
	}
}`

func TestHoverSpellConfig(t *testing.T) {
	markdown, trace, ok := hoverOn(t, newWorkspace(), executeConfigGo, warriorURI(t, "execute.go"), "SpellConfig(&warrior.Unit, executeRank, spelldata.Melee", 3)
	if !ok {
		t.Fatalf("no hover:\n%s", strings.Join(trace, "\n"))
	}
	execute := ladderOf(t, "warrior/Execute").Highest()
	for _, want := range []string{
		"`SpellConfig` of 20662 Execute (Rank 5)",
		"`20662 Execute (Rank 5)`",
		"| **ActionID** | `SpellID 20662` | SpellData · 20662 Execute (Rank 5) |",
		"| **Rank** | `5` | SpellData · 20662 Execute (Rank 5) |",
		"| **SpellSchool** | `SpellSchoolPhysical` | SpellData · 20662 Execute (Rank 5) |",
		"| **DefenseType** | `DefenseTypeMelee` | SpellData · 20662 Execute (Rank 5) |",
		"| **Flags** | `SpellFlagMeleeMetrics \\| SpellFlagNoOnCastComplete \\| SpellFlagAPL` | Melee(ProcMaskMeleeMHSpecial), Flags(SpellFlagNoOnCastComplete) |",
		"| **ProcMask** | `ProcMaskMeleeMHSpecial` | Melee(ProcMaskMeleeMHSpecial) |",
		"| **Cast.DefaultCast.GCD** | `" + execute.GCD().String() + "` | SpellData · 20662 Execute (Rank 5) |",
		"| **Cast.IgnoreHaste** | `true` | SpellData · 20662 Execute (Rank 5) |",
		"| **RageCost.Cost** | `" + strconv.Itoa(int(execute.PowerCost(dbcenums.POWER_RAGE))) + "` | SpellData · 20662 Execute (Rank 5) |",
		"| **RageCost.Refund** | `0.8` | SpellData · 20662 Execute (Rank 5) |",
		"| **DamageMultiplier** | `1` | Melee(ProcMaskMeleeMHSpecial) |",
		"| **ThreatMultiplier** | `1` | Melee(ProcMaskMeleeMHSpecial) |",
		"| **MaxRange** | `" + strconv.FormatFloat(float64(execute.MaxRange), 'f', -1, 64) + "` | SpellData · 20662 Execute (Rank 5) |",
		"| **ClassFlags** | `family 4 mask 0x20000000` | SpellData · 20662 Execute (Rank 5) |",
		"| **CastRequirement.Forms** | `0x50000 (Battle, Berserker)` | SpellData · 20662 Execute (Rank 5) |",
		"| **CastRequirement.NotShapeshifted** | `true` | SpellData · 20662 Execute (Rank 5) |",
		"| **CastRequirement.Weapon.SubclassMask** | `173555 (melee weapon)` | SpellData · 20662 Execute (Rank 5) |",
		"| **CastRequirement.Weapon.Hand** | `ProcMaskMeleeMHAuto \\| ProcMaskMeleeMHSpecial` | SpellData · 20662 Execute (Rank 5) |",
	} {
		if !strings.Contains(markdown, want) {
			t.Errorf("the hover lacks %q:\n%s", want, markdown)
		}
	}

	flags := ""
	for line := range strings.SplitSeq(markdown, "\n") {
		if strings.HasPrefix(line, "| **Flags** |") {
			flags = line
		}
	}
	for _, want := range []string{"SpellFlagMeleeMetrics", "SpellFlagAPL", "SpellFlagNoOnCastComplete", "Melee(ProcMaskMeleeMHSpecial), Flags(SpellFlagNoOnCastComplete)"} {
		if !strings.Contains(flags, want) {
			t.Errorf("the Flags row %q lacks %q", flags, want)
		}
	}

	wantTrace := []string{"SpellConfig", "  row executeRank", "  executeRank = spellData.Execute.Highest()", "  option Melee(ProcMaskMeleeMHSpecial)", "✓ 20662 Execute (Rank 5) → "}
	joined := strings.Join(trace, "\n")
	for _, want := range wantTrace {
		if !strings.Contains(joined, want) {
			t.Errorf("the trace lacks %q:\n%s", want, joined)
		}
	}
}

func TestHoverSpellConfigLocalVariable(t *testing.T) {
	markdown, trace, ok := hoverOn(
		t,
		newWorkspace(),
		executeConfigGo,
		warriorURI(t, "execute.go"),
		"otherConfig := spelldata.SpellConfig",
		len("otherConfig"),
	)
	if !ok {
		t.Fatalf("no hover:\n%s", strings.Join(trace, "\n"))
	}

	for _, want := range []string{
		"`SpellConfig` of 20662 Execute (Rank 5)",
		"| **ActionID** | `SpellID 20662` | SpellData · 20662 Execute (Rank 5) |",
		"| **Rank** | `5` | SpellData · 20662 Execute (Rank 5) |",
		"### Assigned later",
		"| field | value | source |",
		"| **ClassSpellMask** | `SpellMaskExecute` | [10](",
	} {
		if !strings.Contains(markdown, want) {
			t.Errorf("the hover lacks %q:\n%s", want, markdown)
		}
	}
}

func TestHoverSpellConfigAcrossLines(t *testing.T) {
	markdown, trace, ok := hoverOn(t, newWorkspace(), executeConfigGo, warriorURI(t, "execute.go"), "SpellConfig(&warrior.Unit, spellData.Whirlwind", 3)
	if !ok {
		t.Fatalf("no hover:\n%s", strings.Join(trace, "\n"))
	}
	for _, want := range []string{
		"| **ActionID** | `SpellID 1680, Tag 2` | SpellData · 1680 Whirlwind, Tag(2) |",
		"| **ProcMask** | `ProcMaskMeleeOHSpecial` | Melee(ProcMaskMeleeOHSpecial) |",
		"SpellFlagPassiveSpell",
		"Proc()",
	} {
		if !strings.Contains(markdown, want) {
			t.Errorf("the hover lacks %q:\n%s", want, markdown)
		}
	}
	if strings.Contains(markdown, "RageCost") || strings.Contains(markdown, "Cast.DefaultCast.GCD") {
		t.Errorf("Proc() left a cost or a cast behind:\n%s", markdown)
	}
}

func TestHoverSpellConfigUnevaluated(t *testing.T) {
	markdown, trace, ok := hoverOn(t, newWorkspace(), executeConfigGo, warriorURI(t, "execute.go"), "SpellConfig(&warrior.Unit, executeRank, spelldata.Flags(flags)", 3)
	if !ok {
		t.Fatalf("no hover:\n%s", strings.Join(trace, "\n"))
	}
	for _, want := range []string{
		"unevaluated: `spelldata.Flags(flags)`",
		"unevaluated: `spelldata.Label(3)` (spelldata.Label is not an option this reads)",
		"unevaluated: `spelldata.Tag(1 << 40)` (1 << 40 is not the int32 it takes)",
		"Flags(SpellFlagAPL \\| SpellFlagHelpful)",
	} {
		if !strings.Contains(markdown, want) {
			t.Errorf("the hover lacks %q:\n%s", want, markdown)
		}
	}
}

func TestConfigText(t *testing.T) {
	var out bytes.Buffer
	call := "spelldata.SpellConfig(&warrior.Unit, executeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))"
	if err := Run([]string{"-config", call, "-package", "warrior"}, &out); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{"SpellConfig of 20662 Execute (Rank 5)\n", "ProcMaskMeleeMHSpecial", "Melee(ProcMaskMeleeMHSpecial)"} {
		if !strings.Contains(text, want) {
			t.Errorf("-config lacks %q:\n%s", want, text)
		}
	}
}

func TestConfigCastRequirement(t *testing.T) {
	for _, c := range []struct {
		pkg, family string
		want        []string
	}{
		{"warrior", "ShieldBash", []string{"CastRequirement.Forms  ", "0x30000 (Battle, Defensive)", "CastRequirement.NotShapeshifted", "CastRequirement.Shield  "}},
		{"warrior", "Charge", []string{"CastRequirement.Forms  ", "0x10000 (Battle)", "CastRequirement.NotInCombat  "}},
		{"warrior", "SpearingStrike", []string{"173555 (melee weapon)", "ProcMaskMeleeMHAuto | ProcMaskMeleeMHSpecial"}},
		{"hunter", "AimedShot", []string{"CastRequirement.Weapon.SubclassMask", "262156 (bow, gun or crossbow)"}},
		{"druid", "Wrath", []string{"0x40000000 (Moonkin)", "CastRequirement.ExcludedForms", "0x2 (Tree of Life)", "CastRequirement.CasterForm  "}},
	} {
		var out bytes.Buffer
		call := "spelldata.SpellConfig(&" + c.pkg + ".Unit, spellData." + c.family + ".Highest())"
		if err := Run([]string{"-config", call, "-package", c.pkg}, &out); err != nil {
			t.Fatal(err)
		}
		for _, want := range c.want {
			if !strings.Contains(out.String(), want) {
				t.Errorf("-config %s lacks %q:\n%s", c.family, want, out.String())
			}
		}
	}
}

func TestHoverSpellConfigIncludesLaterAssignments(t *testing.T) {
	markdown, trace, ok := hoverOn(
		t,
		newWorkspace(),
		executeConfigGo,
		warriorURI(t, "execute.go"),
		"config := spelldata.SpellConfig",
		3,
	)
	if !ok {
		t.Fatalf("no hover:\n%s", strings.Join(trace, "\n"))
	}

	for _, want := range []string{
		"### Assigned later",
		"| field | value | source |",
		"| **ClassSpellMask** | `SpellMaskExecute` | [7](",
	} {
		if !strings.Contains(markdown, want) {
			t.Errorf("the hover lacks %q:\n%s", want, markdown)
		}
	}

	if strings.Contains(markdown, "Assignments to the config after the call are not folded in.") {
		t.Errorf("old assignment footnote is still present:\n%s", markdown)
	}
}

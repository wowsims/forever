package database

// Pins what the generator emits for each shape a row can take, from synthetic
// rows that cover every branch of the templates, and builds the result against
// sim/core/buffs so that a generated Meta or constructor call that does not
// compile against sim/core/buffs/meta.go is noticed here.
//
// Needs no game files database. Set UPDATE_BUFF_FIXTURES=1 to rewrite the fixtures
// after a deliberate change.

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/buffmanifest"
)

// One row per shape the templates have a branch for. The proto field of each row
// is a real one whose compiled type matches the row's declared type, so the
// rendered apply blocks type-check against the sim. Each row states its spell
// and goes through the same parse a resolved row does.
func syntheticBuffRows() []ResolvedBuff {
	aura := func(a dbcenums.EffectAuraType, misc int32, points float64) spelldata.Effect {
		return spelldata.Effect{Type: dbcenums.E_APPLY_AURA, Aura: a, Misc: misc, BasePoints: points}
	}
	spell := func(id int32, school core.SpellSchool, durationMs int32, effects ...spelldata.Effect) *spelldata.Spell {
		for i := range effects {
			effects[i].SpellID, effects[i].Index = id, uint8(i)
		}
		return &spelldata.Spell{ID: id, School: school, DurationMs: durationMs, Effects: effects}
	}
	mana := spelldata.Effect{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_PERIODIC_ENERGIZE,
		Misc: int32(dbcenums.POWER_MANA), BasePoints: 10, PeriodMs: 2000}

	rows := []ResolvedBuff{
		{
			BuffSpec: buffmanifest.BuffSpec{
				Field: "mana_spring_totem", Category: "ManaSpringTotem",
			},
			Scope: buffmanifest.ScopeParty, Proto: buffmanifest.ProtoTristate,
			Go: "SynthManaSpring", Name: "Mana Spring Totem",
			SpellID: 10494, CastSpellID: 10494, Totem: true, Supported: true,
			Spell:         spell(10494, core.SpellSchoolNature, 0, mana),
			TalentRanks:   5,
			TalentSpellID: 16187, TalentPosition: 1,
		},
		{
			BuffSpec: buffmanifest.BuffSpec{
				Field: "mana_tide_totems", Category: "ManaTideTotem",
			},
			Scope: buffmanifest.ScopeParty, Proto: buffmanifest.ProtoInt32, Kind: buffmanifest.KindExternalCD,
			Go: "SynthManaTide", Name: "Mana Tide Totem",
			SpellID: 17360, CastSpellID: 17359, DurationMs: 12000, CooldownMs: 300000, Supported: true,
			Spell: spell(17360, core.SpellSchoolNature, 0, mana),
		},
		{
			BuffSpec: buffmanifest.BuffSpec{Field: "greater_blessing_of_kings"},
			Scope:    buffmanifest.ScopeIndividual, Proto: buffmanifest.ProtoBool,
			Go: "SynthBlessingOfKings", Name: "Blessing of Kings",
			SpellID: 20217, CastSpellID: 20217, DurationMs: 3600000, Supported: true,
			Spell: spell(20217, core.SpellSchoolHoly, 3600000,
				aura(dbcenums.A_MOD_TOTAL_STAT_PERCENTAGE, 0, 10),
				aura(dbcenums.A_MOD_TOTAL_STAT_PERCENTAGE, 1, 10)),
		},
		{
			BuffSpec: buffmanifest.BuffSpec{
				Field: "battle_shout", Category: "SynthBattleShout",
				SingleAura: true, Driver: true,
			},
			Scope: buffmanifest.ScopeRaid, Proto: buffmanifest.ProtoTristate,
			Go: "SynthBattleShout", Name: "Battle Shout",
			SpellID: 25289, CastSpellID: 25289, DurationMs: 180000, Supported: true,
			Spell: spell(25289, core.SpellSchoolPhysical, 180000, aura(dbcenums.A_MOD_ATTACK_POWER, 0, 139)),
		},
		{
			BuffSpec: buffmanifest.BuffSpec{
				Field: "devotion_aura", Category: "DevotionAura",
				SharedCategory: "SynthPaladinAura", SingleAura: true,
			},
			Scope: buffmanifest.ScopeParty, Proto: buffmanifest.ProtoBool,
			Go: "SynthDevotionAura", Name: "Devotion Aura",
			SpellID: 10293, CastSpellID: 10293, DurationMs: 600000, Supported: true,
			Spell: spell(10293, core.SpellSchoolHoly, 600000,
				aura(dbcenums.A_MOD_ARMOR_NO_MODS, 1, 735),
				aura(dbcenums.A_MOD_HEALING_PCT, 0, 0)),
			TalentRanks:          2,
			TalentScalesDuration: true,
			TalentSpellID:        20140, TalentPosition: 2,
		},
		{
			BuffSpec: buffmanifest.BuffSpec{
				Field: "frost_resistance_aura", Category: "FrostResistanceAura", SharedCategory: "SynthPaladinAura", SingleAura: true,
			},
			Scope: buffmanifest.ScopeRaid, Proto: buffmanifest.ProtoBool,
			Go: "SynthFrostResistanceAura", Name: "Frost Resistance Aura",
			SpellID: 19898, CastSpellID: 19898, Supported: true,
			Spell: spell(19898, core.SpellSchoolHoly, -1, aura(dbcenums.A_MOD_RESISTANCE, 16, 60)),
		},
		{
			BuffSpec: buffmanifest.BuffSpec{Field: "frost_resistance_totem"},
			Scope:    buffmanifest.ScopeRaid, Proto: buffmanifest.ProtoBool,
			Go: "SynthFrostResistanceTotem", Name: "Frost Resistance Totem",
			SpellID: 10477, CastSpellID: 10477, Totem: true, Supported: true,
			Spell: spell(10477, core.SpellSchoolNature, 0, aura(dbcenums.A_MOD_RESISTANCE, 16, 60)),
		},
		{
			BuffSpec: buffmanifest.BuffSpec{
				Field: "thunder_clap", Category: "AtkSpdReduction",
			},
			Scope: buffmanifest.ScopeDebuff, Proto: buffmanifest.ProtoBool,
			Go: "SynthThunderClap", Name: "Thunder Clap",
			SpellID: 11581, CastSpellID: 11581, DurationMs: 30000, Supported: true,
			Spell:         spell(11581, core.SpellSchoolPhysical, 30000, aura(dbcenums.A_MOD_MELEE_HASTE_3, 0, -20)),
			TalentRanks:   2,
			TalentSpellID: 12287, TalentPosition: 1,
		},
		{
			BuffSpec: buffmanifest.BuffSpec{
				Field: "innervates", Label: "Innervates",
				Category: "Innervate",
			},
			Scope: buffmanifest.ScopeIndividual, Proto: buffmanifest.ProtoInt32, Kind: buffmanifest.KindExternalCD,
			Go: "SynthInnervates", Name: "Innervate",
			SpellID: 29166, CastSpellID: 29166, DurationMs: 20000, CooldownMs: 360000, Supported: true,
			Spell: spell(29166, core.SpellSchoolNature, 20000,
				aura(dbcenums.A_MOD_MANA_REGEN_INTERRUPT, 0, 100),
				aura(dbcenums.A_MOD_POWER_REGEN_PERCENT, 0, 400)),
		},
		{
			BuffSpec: buffmanifest.BuffSpec{
				Field: "power_infusions", Label: "Power Infusions",
				Category: "PowerInfusion",
			},
			Scope: buffmanifest.ScopeIndividual, Proto: buffmanifest.ProtoInt32, Kind: buffmanifest.KindExternalCD,
			Go: "SynthPowerInfusions", Name: "Power Infusion",
			SpellID: 10060, CastSpellID: 10060, DurationMs: 15000, CooldownMs: 180000, Supported: true,
			Spell: spell(10060, core.SpellSchoolHoly, 15000,
				aura(dbcenums.A_MOD_DAMAGE_PERCENT_DONE, 126, 20),
				aura(dbcenums.A_MOD_HEALING_DONE_PERCENT, 0, 20)),
		},
		{
			BuffSpec: buffmanifest.BuffSpec{
				Field: "totem_twisting", Category: "Absent",
			},
			Scope: buffmanifest.ScopeParty, Proto: buffmanifest.ProtoBool, Kind: buffmanifest.KindFlag,
			Go:     "SynthAbsent",
			Reason: "the game files have no row for it",
		},
		{
			BuffSpec: buffmanifest.BuffSpec{
				Field: "atiesh_mage", Label: "Atiesh - Mage",
			},
			Scope: buffmanifest.ScopeParty, Proto: buffmanifest.ProtoInt32, Kind: buffmanifest.KindItemCount,
			Go:      "SynthAtieshMage",
			SpellID: 28142, CastSpellID: 28142, Supported: true,
			Spell: spell(28142, core.SpellSchoolPhysical, -1, aura(dbcenums.A_MOD_SPELL_CRIT_CHANCE, 0, 2)),
		},
		{
			BuffSpec: buffmanifest.BuffSpec{
				Field: "thorns", Category: "Thorns",
			},
			Scope: buffmanifest.ScopeIndividual, Proto: buffmanifest.ProtoBool, Kind: buffmanifest.KindDamageShield,
			Go: "SynthThorns", Name: "Thorns",
			SpellID: 9910, CastSpellID: 9910, DurationMs: 600000, Supported: true,
			Spell:         spell(9910, core.SpellSchoolNature, 600000, aura(dbcenums.A_DAMAGE_SHIELD, 0, 18)),
			TalentRanks:   2,
			TalentSpellID: 16836, TalentPosition: 1,
		},
		{
			BuffSpec: buffmanifest.BuffSpec{
				Field: "sunder_armor", Category: "MajorArmorReduction",
				SingleAura: true, Driver: true,
			},
			Scope: buffmanifest.ScopeDebuff, Proto: buffmanifest.ProtoBool,
			Go: "SynthSunderArmor", Name: "Sunder Armor",
			SpellID: 11597, CastSpellID: 11597, DurationMs: 30000, Supported: true,
			Spell: func() *spelldata.Spell {
				s := spell(11597, core.SpellSchoolPhysical, 30000, aura(dbcenums.A_MOD_ARMOR_NO_MODS, 1, -450))
				s.MaxStack = 5
				return s
			}(),
		},
		{
			BuffSpec: buffmanifest.BuffSpec{
				Field: "expose_armor", Category: "MajorArmorReduction",
				SingleAura: true,
			},
			Scope: buffmanifest.ScopeDebuff, Proto: buffmanifest.ProtoBool,
			Go: "SynthExposeArmor", Name: "Expose Armor",
			SpellID: 11198, CastSpellID: 11198, DurationMs: 30000, Supported: true,
			Spell: spell(11198, core.SpellSchoolPhysical, 30000, spelldata.Effect{Type: dbcenums.E_APPLY_AURA,
				Aura: dbcenums.A_MOD_ARMOR_NO_MODS, Misc: 1, PointsPerResource: -450}),
		},
		{
			BuffSpec: buffmanifest.BuffSpec{
				Field: "curse_of_elements", Category: "CurseOfElements",
				SingleAura: true,
			},
			Scope: buffmanifest.ScopeDebuff, Proto: buffmanifest.ProtoBool,
			Go: "SynthCurseOfElements", Name: "Curse of the Elements",
			SpellID: 1311680, CastSpellID: 1311680, DurationMs: 300000, Supported: true,
			Spell: spell(1311680, core.SpellSchoolShadow, 300000,
				aura(dbcenums.A_MOD_RESISTANCE, 124, -75),
				aura(dbcenums.A_MOD_DAMAGE_PERCENT_TAKEN, 126, 10)),
		},
		{
			BuffSpec: buffmanifest.BuffSpec{
				Field: "judgement_of_the_crusader", Category: "Judgement of the Crusader", SingleAura: true,
			},
			Scope: buffmanifest.ScopeDebuff, Proto: buffmanifest.ProtoBool,
			Go: "SynthJudgementOfTheCrusader", Name: "Judgement of the Crusader",
			SpellID: 20303, CastSpellID: 20303, DurationMs: 10000, Supported: true,
			Spell: spell(20303, core.SpellSchoolHoly, 10000, aura(dbcenums.A_MOD_DAMAGE_TAKEN, 2, 140)),
		},
	}
	for i := range rows {
		if rows[i].Spell == nil {
			continue
		}
		parseBuff(&rows[i])
	}
	return rows
}

var syntheticFixtures = map[string]string{
	buffsGenFile:   filepath.Join("testdata", "buffs_synthetic_auto_gen.go"),
	debuffsGenFile: filepath.Join("testdata", "debuffs_synthetic_auto_gen.go"),
}

func renderSyntheticBuffFiles(t *testing.T) map[string][]byte {
	t.Helper()

	files, err := renderBuffFiles(syntheticBuffRows())
	if err != nil {
		t.Fatalf("rendering the synthetic rows: %v", err)
	}
	return files
}

// A cast is a totem when it summons into a totem slot; any other summon, or a slot named by
// another effect, is not.
func TestTotemIsASummonIntoATotemSlot(t *testing.T) {
	for _, tc := range []struct {
		name    string
		effects []storeEffect
		want    bool
	}{
		{"totem slot", []storeEffect{{Type: dbcenums.E_SUMMON, Misc2: totemSlotSummonProperties[0]}}, true},
		{"other summon", []storeEffect{{Type: dbcenums.E_SUMMON, Misc2: 1}}, false},
		{"slot on another effect", []storeEffect{{Type: dbcenums.E_APPLY_AURA, Misc2: totemSlotSummonProperties[0]}}, false},
		{"no effects", nil, false},
	} {
		if got := summonsTotem(storeSpell{Effects: tc.effects}); got != tc.want {
			t.Errorf("%s: a totem %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestRenderedBuffFilesMatchTheFixtures(t *testing.T) {
	files := renderSyntheticBuffFiles(t)

	for name, rendered := range files {
		fixture := syntheticFixtures[name]
		if os.Getenv("UPDATE_BUFF_FIXTURES") != "" {
			if err := os.WriteFile(fixture, rendered, 0644); err != nil {
				t.Fatalf("writing %s: %v", fixture, err)
			}
			continue
		}

		committed, err := os.ReadFile(fixture)
		if err != nil {
			t.Fatalf("reading %s: %v", fixture, err)
		}
		if string(committed) != string(rendered) {
			line, want, got := firstDifference(committed, rendered)
			t.Errorf("%s differs from what the generator emits, from line %d:\n  committed: %s\n  rendered:  %s\n"+
				"rewrite it with `UPDATE_BUFF_FIXTURES=1 go test ./tools/database/`",
				fixture, line, want, got)
		}
	}
}

// Builds sim/core/buffs with the rendered files overlaid onto it, which is the
// only check that a generated constructor still names an identifier the support
// API declares. The apply functions are renamed because the real generated files
// already declare them, and the drivers the rows call are stubbed here the way
// sim/core/buffs/drivers.go would declare them.
func TestRenderedBuffFilesCompile(t *testing.T) {
	goTool := findGoTool(t)
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("finding the repository root: %v", err)
	}

	dir := t.TempDir()
	overlay := map[string]string{}
	for name, rendered := range renderSyntheticBuffFiles(t) {
		body := strings.ReplaceAll(string(rendered), "func applyGenerated", "func synthApplyGenerated")
		body = strings.ReplaceAll(body, "var debuffMetas", "var synthDebuffMetas")
		path := filepath.Join(dir, filepath.Base(name))
		if err := os.WriteFile(path, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
		overlay[filepath.Join(root, "sim", "core", "buffs", "zz_synthetic_"+filepath.Base(name))] = path
	}

	drivers := filepath.Join(dir, "drivers.go")
	if err := os.WriteFile(drivers, []byte("package buffs\n\n"+
		"import (\n\t\"github.com/wowsims/forever/sim/core\"\n\t\"github.com/wowsims/forever/sim/core/proto\"\n)\n\n"+
		"func driveSynthInnervates(char *core.Character, individual *proto.IndividualBuffs) {\n"+
		"\tnewGeneratedExternalCD(char, NewSpellDataSynthInnervatesAura(&char.Unit, false, 0),"+
		" generatedExternalCD{NumSources: individual.Innervates,"+
		" Cooldown: SynthInnervatesCooldown(), Type: core.CooldownTypeMana})\n}\n\n"+
		"func driveSynthBattleShout(char *core.Character, _ *proto.RaidBuffs) {\n"+
		"\tapplyFixedShoutAura(char, NewSpellDataSynthBattleShoutAura(&char.Unit, false, 0),"+
		" SynthBattleShoutCategory)\n}\n\n"+
		"func driveSynthSunderArmor(target *core.Unit, _ *proto.Debuffs, _ *proto.Raid) {\n"+
		"\tcore.MakePermanent(NewSpellDataSynthSunderArmorAura(target, false, 0))\n}\n\n"+
		"func driveSynthAtieshMage(char *core.Character, party *proto.PartyBuffs) {\n"+
		"\tcore.MakePermanent(NewSpellDataSynthAtieshMageAura(&char.Unit, false, 0,"+
		" float64(party.AtieshMage)))\n}\n\n"+
		"func driveSynthPowerInfusions(char *core.Character, individual *proto.IndividualBuffs) {\n"+
		"\tnewGeneratedExternalCD(char, NewSpellDataSynthPowerInfusionsAura(&char.Unit, false, 0),"+
		" generatedExternalCD{NumSources: individual.PowerInfusions,"+
		" Cooldown: SynthPowerInfusionsCooldown(), Type: core.CooldownTypeDPS})\n}\n\n"+
		"func driveSynthManaTide(char *core.Character, party *proto.PartyBuffs) {\n"+
		"\tnewGeneratedExternalCD(char, NewSpellDataSynthManaTideAura(&char.Unit, false, 0),"+
		" generatedExternalCD{NumSources: party.ManaTideTotems,"+
		" Cooldown: SynthManaTideCooldown(), Type: core.CooldownTypeMana})\n}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	overlay[filepath.Join(root, "sim", "core", "buffs", "zz_synthetic_drivers.go")] = drivers

	overlayPath := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(overlayPath, []byte(overlayJSON(overlay)), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(goTool, "build", "-overlay", overlayPath, "./sim/core/buffs/")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Errorf("the generated constructors do not compile against sim/core/buffs:\n%s", out)
	}
}

func overlayJSON(replace map[string]string) string {
	var b strings.Builder
	b.WriteString(`{"Replace":{`)
	first := true
	for from, to := range replace {
		if !first {
			b.WriteString(",")
		}
		first = false
		b.WriteString(`"` + filepath.ToSlash(from) + `":"` + filepath.ToSlash(to) + `"`)
	}
	b.WriteString("}}")
	return b.String()
}

func findGoTool(t *testing.T) string {
	t.Helper()

	if path := filepath.Join(runtime.GOROOT(), "bin", "go"); runtime.GOROOT() != "" {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	path, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go tool on PATH, so the generated files cannot be type-checked")
	}
	return path
}

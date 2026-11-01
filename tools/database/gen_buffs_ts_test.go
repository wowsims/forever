package database

// The shapes the settings-input renderer chooses per proto type and scope. That
// the committed file is what the manifest renders is buffs_regen_test.go's.

import (
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/buffmanifest"
)

func TestRenderBuffsDebuffsTSShapes(t *testing.T) {
	rows := []ResolvedBuff{
		{
			BuffSpec: buffmanifest.BuffSpec{
				Field: "prayer_of_shadow_protection",
				Stats: []proto.Stat{proto.Stat_StatShadowResistance, proto.Stat_StatStamina},
			},
			Scope: buffmanifest.ScopeRaid, Proto: buffmanifest.ProtoBool, Go: "PrayerOfShadowProtection",
			Owner:   proto.Class_ClassPriest,
			SpellID: 27683, CastSpellID: 27683, DBName: "Prayer of Shadow Protection",
		},
		{
			BuffSpec: buffmanifest.BuffSpec{
				Field: "hunters_mark", Stats: []proto.Stat{proto.Stat_StatRangedAttackPower},
			},
			Scope: buffmanifest.ScopeDebuff, Proto: buffmanifest.ProtoTristate, Go: "HuntersMark",
			Owner:   proto.Class_ClassHunter,
			SpellID: 14325, CastSpellID: 14325, DBName: "Hunter's Mark",
			TalentRanks: 1, TalentSpellID: 19425,
		},
		{
			BuffSpec: buffmanifest.BuffSpec{
				Field: "mana_tide_totems", Stats: []proto.Stat{proto.Stat_StatMP5}, Label: "Mana Tide Totem",
			},
			Scope: buffmanifest.ScopeParty, Proto: buffmanifest.ProtoInt32, Kind: buffmanifest.KindExternalCD,
			Go: "ManaTideTotems", Owner: proto.Class_ClassShaman,
			SpellID: 17359, CastSpellID: 17359, DBName: "Mana Tide Totem",
		},
		{
			BuffSpec: buffmanifest.BuffSpec{
				Field: "greater_blessing_of_might", Stats: []proto.Stat{proto.Stat_StatAttackPower},
			},
			Scope: buffmanifest.ScopeIndividual, Proto: buffmanifest.ProtoInt32, Go: "GreaterBlessingOfMight",
			SpellID: 25916, CastSpellID: 25916, Name: "Greater Blessing of Might",
			RankStates: []rankState{
				{SpellID: 25782, CastSpellID: 25782, Label: "Greater Blessing of Might Rank 1"},
				{SpellID: 25916, CastSpellID: 25916, Label: "Greater Blessing of Might Rank 2"},
			},
		},
	}

	rendered, err := RenderBuffsDebuffsTS(rows)
	if err != nil {
		t.Fatalf("rendering four rows: %v", err)
	}
	out := string(rendered)

	want := []string{
		`export const GreaterBlessingOfMight = makeRankedIndividualBuffInput({
	ranks: [
		{ actionId: ActionId.fromSpellId(25782), label: 'Greater Blessing of Might Rank 1' },
		{ actionId: ActionId.fromSpellId(25916), label: 'Greater Blessing of Might Rank 2' },
	],
	fieldName: 'greaterBlessingOfMight',
	label: 'Greater Blessing of Might',
});`,
		`export const PrayerOfShadowProtection = makeBooleanRaidBuffInput({
	actionId: ActionId.fromSpellId(27683),
	fieldName: 'prayerOfShadowProtection',
	label: 'Prayer of Shadow Protection',
});`,
		`export const HuntersMark = makeTristateDebuffInput({
	actionId: ActionId.fromSpellId(14325),
	impId: ActionId.fromSpellId(19425),
	fieldName: 'huntersMark',
	label: "Hunter's Mark",
});`,
		`export const ManaTideTotems = makeMultistatePartyBuffInput({
	actionId: ActionId.fromSpellId(17359),
	numStates: 5,
	fieldName: 'manaTideTotems',
	label: 'Mana Tide Totem',
});`,
		`export const GENERATED_RAID_BUFFS_CONFIG: RenderableStatOptions[] = [
	{
		config: PrayerOfShadowProtection,
		stats: [Stat.StatShadowResistance, Stat.StatStamina],
		ownerClass: Class.ClassPriest,
	},
];`,
		`export const GENERATED_PARTY_BUFFS_CONFIG: RenderableStatOptions[] = [
	{
		config: ManaTideTotems,
		stats: [Stat.StatMP5],
		ownerClass: Class.ClassShaman,
	},
];`,
		`export const GENERATED_INDIVIDUAL_BUFFS_CONFIG: RenderableStatOptions[] = [
	{
		config: GreaterBlessingOfMight,
		stats: [Stat.StatAttackPower],
	},
];`,
		`export const GENERATED_DEBUFFS_CONFIG: RenderableStatOptions[] = [
	{
		config: HuntersMark,
		stats: [Stat.StatRangedAttackPower],
		ownerClass: Class.ClassHunter,
	},
];`,
	}
	for _, block := range want {
		if !strings.Contains(out, block) {
			t.Errorf("the rendered file is missing:\n%s\ngot:\n%s", block, out)
		}
	}
}

func TestRenderBuffsDebuffsTSSkips(t *testing.T) {
	rows := []ResolvedBuff{
		{
			BuffSpec: buffmanifest.BuffSpec{Field: "misery"},
			Scope:    buffmanifest.ScopeDebuff, Proto: buffmanifest.ProtoBool, Kind: buffmanifest.KindFlag,
			Go: "Misery", Owner: proto.Class_ClassPriest,
			SpellID: 33195, DBName: "Misery",
			Reason: "no SpellName row for Misery.",
		},
		{
			BuffSpec: buffmanifest.BuffSpec{Field: "greater_blessing_of_salvation", Kind: buffmanifest.KindFlag},
			Scope:    buffmanifest.ScopeIndividual, Proto: buffmanifest.ProtoBool,
			Go: "GreaterBlessingOfSalvation", Owner: proto.Class_ClassPaladin,
			Reason: "the spell is gone from the game files.",
		},
	}

	rendered, err := RenderBuffsDebuffsTS(rows)
	if err != nil {
		t.Fatalf("rendering two skipped rows: %v", err)
	}
	out := string(rendered)

	for _, name := range []string{"Misery", "GreaterBlessingOfSalvation"} {
		if strings.Contains(out, "export const "+name+" ") {
			t.Errorf("%s has no settings input but rendered one:\n%s", name, out)
		}
	}
	for _, comment := range []string{
		"// misery: no SpellName row for Misery.",
		"// greater_blessing_of_salvation: the spell is gone from the game files.",
	} {
		if !strings.Contains(out, comment) {
			t.Errorf("the rendered file is missing %q:\n%s", comment, out)
		}
	}
}

func TestRenderBuffsDebuffsTSRejectsAnUncountedInt32(t *testing.T) {
	rows := []ResolvedBuff{{
		BuffSpec: buffmanifest.BuffSpec{Field: "totem_of_wrath"},
		Scope:    buffmanifest.ScopeParty, Proto: buffmanifest.ProtoInt32, Kind: buffmanifest.KindItemCount,
		Go: "TotemOfWrath", Owner: proto.Class_ClassShaman,
		SpellID: 30706, DBName: "Totem of Wrath",
	}}

	if _, err := RenderBuffsDebuffsTS(rows); err == nil {
		t.Error("an int32 row with no numStates rendered without an error")
	}
}

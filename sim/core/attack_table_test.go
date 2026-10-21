package core

import (
	"fmt"
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func attackTableUnits(attackerType UnitType, attackerLevel int32, defenderType UnitType, defenderLevel int32) (*Unit, *Unit) {
	attacker := &Unit{Type: attackerType, Level: attackerLevel, PseudoStats: stats.NewPseudoStats()}
	defender := &Unit{Type: defenderType, Level: defenderLevel, PseudoStats: stats.NewPseudoStats()}
	return attacker, defender
}

type attackTableValue struct {
	name       string
	got, value float64
}

func checkAttackTableValues(t *testing.T, when string, values []attackTableValue) {
	t.Helper()
	for _, v := range values {
		if math.Abs(v.got-v.value) > 1e-9 {
			t.Errorf("%s: %s = %v, want %v", when, v.name, v.got, v.value)
		}
	}
}

func TestPlayerAttackTableAgainstEachTargetLevel(t *testing.T) {
	for _, tc := range []struct {
		level                                   int32
		miss, dodge, parry, block               float64
		glance, glanceMin, glanceMax, critSuppr float64
		auraCritSuppr, spellMiss                float64
	}{
		{55, 0.025, 0.025, 0.025, 0.025, 0, 0.91, 0.99, -0.01, 0, 0},
		{56, 0.03, 0.03, 0.03, 0.03, 0, 0.91, 0.99, -0.008, 0, 0},
		{57, 0.035, 0.035, 0.035, 0.035, 0, 0.91, 0.99, -0.006, 0, 0.01},
		{58, 0.04, 0.04, 0.04, 0.04, 0, 0.91, 0.99, -0.004, 0, 0.02},
		{59, 0.045, 0.045, 0.045, 0.045, 0, 0.91, 0.99, -0.002, 0, 0.03},
		{60, 0.05, 0.05, 0.05, 0.05, 0.10, 0.91, 0.99, 0, 0, 0.04},
		{61, 0.055, 0.055, 0.055, 0.05, 0.20, 0.91, 0.99, 0.01, 0, 0.05},
		{62, 0.06, 0.06, 0.06, 0.05, 0.30, 0.80, 0.90, 0.02, 0, 0.06},
		{63, 0.08, 0.065, 0.14, 0.05, 0.40, 0.55, 0.75, 0.03, 0.018, 0.17},
		{64, 0.08, 0.065, 0.14, 0.05, 0.40, 0.55, 0.75, 0.03, 0.018, 0.17},
	} {
		attacker, defender := attackTableUnits(PlayerUnit, 60, EnemyUnit, tc.level)
		table := NewAttackTable(attacker, defender)
		want := AttackTableHand{
			BaseMissChance:       tc.miss,
			BaseDodgeChance:      tc.dodge,
			BaseParryChance:      tc.parry,
			BaseBlockChance:      tc.block,
			BaseGlanceChance:     tc.glance,
			GlanceMultiplierMin:  tc.glanceMin,
			GlanceMultiplierMax:  tc.glanceMax,
			MeleeCritSuppression: tc.critSuppr,
		}
		values := []attackTableValue{
			{"aura crit suppression", table.AuraCritSuppression, tc.auraCritSuppr},
			{"spell miss", table.BaseSpellMissChance, tc.spellMiss},
		}
		for hand, entry := range map[string]AttackTableHand{"main hand": table.MH, "off hand": table.OH, "ranged": table.Ranged} {
			values = append(values, handValues(hand, entry, want)...)
		}
		checkAttackTableValues(t, fmt.Sprintf("level 60 against %d", tc.level), values)
	}
}

func handValues(hand string, got, want AttackTableHand) []attackTableValue {
	return []attackTableValue{
		{hand + " miss", got.BaseMissChance, want.BaseMissChance},
		{hand + " dodge", got.BaseDodgeChance, want.BaseDodgeChance},
		{hand + " parry", got.BaseParryChance, want.BaseParryChance},
		{hand + " block", got.BaseBlockChance, want.BaseBlockChance},
		{hand + " glance chance", got.BaseGlanceChance, want.BaseGlanceChance},
		{hand + " glance multiplier min", got.GlanceMultiplierMin, want.GlanceMultiplierMin},
		{hand + " glance multiplier max", got.GlanceMultiplierMax, want.GlanceMultiplierMax},
		{hand + " melee crit suppression", got.MeleeCritSuppression, want.MeleeCritSuppression},
	}
}

// Skill below the level-based maximum widens the sword hand's table except block. Skill above it narrows miss,
// dodge and glancing damage, while parry, glance chance and crit suppression stay on the level-based skill.
func TestWeaponSkillSettingSetsTheSwordHandsTableAgainstALevel63Target(t *testing.T) {
	sword := &Item{WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand}
	mace := &Item{WeaponType: proto.WeaponType_WeaponTypeMace, HandType: proto.HandType_HandTypeOneHand}
	level63 := AttackTableHand{
		BaseMissChance:       0.08,
		BaseDodgeChance:      0.065,
		BaseParryChance:      0.14,
		BaseBlockChance:      0.05,
		BaseGlanceChance:     0.40,
		GlanceMultiplierMin:  0.55,
		GlanceMultiplierMax:  0.75,
		MeleeCritSuppression: 0.03,
	}

	for _, tc := range []struct {
		skill int32
		sword AttackTableHand
	}{
		{285, AttackTableHand{
			BaseMissChance:       0.11,
			BaseDodgeChance:      0.08,
			BaseParryChance:      0.23,
			BaseBlockChance:      0.05,
			BaseGlanceChance:     0.70,
			GlanceMultiplierMin:  0.01,
			GlanceMultiplierMax:  0.30,
			MeleeCritSuppression: 0.06,
		}},
		{300, level63},
		{305, AttackTableHand{
			BaseMissChance:       0.06,
			BaseDodgeChance:      0.06,
			BaseParryChance:      0.14,
			BaseBlockChance:      0.05,
			BaseGlanceChance:     0.40,
			GlanceMultiplierMin:  0.80,
			GlanceMultiplierMax:  0.90,
			MeleeCritSuppression: 0.03,
		}},
		{315, AttackTableHand{
			BaseMissChance:       0.05,
			BaseDodgeChance:      0.05,
			BaseParryChance:      0.14,
			BaseBlockChance:      0.05,
			BaseGlanceChance:     0.40,
			GlanceMultiplierMin:  0.91,
			GlanceMultiplierMax:  0.99,
			MeleeCritSuppression: 0.03,
		}},
	} {
		character := NewCharacter(&Party{}, 0, &proto.Player{
			Name:         "Skill Tester",
			Class:        proto.Class_ClassWarrior,
			Race:         proto.Race_RaceUndead,
			Spec:         &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{}},
			Equipment:    &proto.EquipmentSpec{},
			WeaponSkills: []*proto.WeaponSkillValue{{Skill: proto.WeaponSkill_WeaponSkillSwords, Value: tc.skill}},
		})
		defender := &Unit{Type: EnemyUnit, Level: 63, PseudoStats: stats.NewPseudoStats()}
		table := NewAttackTable(&character.Unit, defender)
		table.setWeapons(sword, mace, nil)

		values := handValues("sword hand", table.MH, tc.sword)
		values = append(values, handValues("mace hand", table.OH, level63)...)
		values = append(values, handValues("ranged", table.Ranged, level63)...)
		values = append(values, attackTableValue{"aura crit suppression", table.AuraCritSuppression, 0.018})
		checkAttackTableValues(t, fmt.Sprintf("%d sword skill", tc.skill), values)
	}
}

// Weapon skill above the target's defense lowers its block by 0.1% a point; skill below it leaves block at 5%.
func TestWeaponSkillAboveTheTargetsDefenseLowersItsBlock(t *testing.T) {
	sword := &Item{WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand}
	for _, tc := range []struct {
		skill, targetLevel int32
		block              float64
	}{
		{320, 63, 0.045},
		{305, 60, 0.045},
		{300, 63, 0.05},
		{285, 63, 0.05},
	} {
		character := NewCharacter(&Party{}, 0, &proto.Player{
			Name:         "Block Tester",
			Class:        proto.Class_ClassWarrior,
			Race:         proto.Race_RaceUndead,
			Spec:         &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{}},
			Equipment:    &proto.EquipmentSpec{},
			WeaponSkills: []*proto.WeaponSkillValue{{Skill: proto.WeaponSkill_WeaponSkillSwords, Value: tc.skill}},
		})
		defender := &Unit{Type: EnemyUnit, Level: tc.targetLevel, PseudoStats: stats.NewPseudoStats()}
		table := NewAttackTable(&character.Unit, defender)
		table.setWeapons(sword, nil, nil)
		spell := &Spell{Unit: &character.Unit, ProcMask: ProcMaskMeleeMHSpecial}
		checkAttackTableValues(t, fmt.Sprintf("%d sword skill against level %d", tc.skill, tc.targetLevel), []attackTableValue{
			{"main hand block", table.MH.BaseBlockChance, tc.block},
			{"total block", defender.GetTotalBlockChanceAsDefender(spell, table), tc.block},
		})
	}
}

// The wand hand's spell miss is the spell table's at a level difference of (capped level difference × 5 −
// wand skill above the maximum) / 5, truncated toward zero over the whole expression. The level difference
// is capped at +3 like a spell's, so 303 skill against a level 61 target is +0 and 310 against level 65 is +1.
func TestWandSkillSetsTheRangedHandsSpellMiss(t *testing.T) {
	wand := &Item{RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand}
	for _, tc := range []struct {
		skill, targetLevel int32
		spellMiss          float64
	}{
		{300, 60, 0.04},
		{300, 61, 0.05},
		{300, 63, 0.17},
		{297, 60, 0.04},
		{295, 60, 0.05},
		{290, 60, 0.06},
		{285, 60, 0.17},
		{275, 60, 0.17},
		{305, 63, 0.06},
		{310, 60, 0.02},
		{310, 65, 0.05},
		{315, 65, 0.04},
		{303, 61, 0.04},
		{303, 62, 0.05},
		{308, 62, 0.04},
		{312, 61, 0.03},
	} {
		character := NewCharacter(&Party{}, 0, &proto.Player{
			Name:         "Wand Tester",
			Class:        proto.Class_ClassPriest,
			Race:         proto.Race_RaceHuman,
			Spec:         &proto.Player_DpsPriest{DpsPriest: &proto.DpsPriest{}},
			Equipment:    &proto.EquipmentSpec{},
			WeaponSkills: []*proto.WeaponSkillValue{{Skill: proto.WeaponSkill_WeaponSkillWands, Value: tc.skill}},
		})
		defender := &Unit{Type: EnemyUnit, Level: tc.targetLevel, PseudoStats: stats.NewPseudoStats()}
		table := NewAttackTable(&character.Unit, defender)
		table.setWeapons(nil, nil, wand)
		checkAttackTableValues(t, fmt.Sprintf("%d wand skill against level %d", tc.skill, tc.targetLevel), []attackTableValue{
			{"ranged spell miss", table.Ranged.BaseSpellMissChance, tc.spellMiss},
		})
	}
}

func TestFeralCombatSkillReplacesTheWeaponsSkillInAForm(t *testing.T) {
	staff := &Item{ID: 1, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand}
	for _, tc := range []struct {
		name        string
		inForm      bool
		feralSkill  float64
		weapon      *Item
		weaponSkill float64
	}{
		{"feral skill at the maximum", true, 0, staff, 0},
		{"no weapon in a form", true, -10, nil, -10},
		{"out of a form", false, -10, staff, -20},
	} {
		unit := &Unit{PseudoStats: stats.NewPseudoStats()}
		unit.PseudoStats.WeaponSkills[proto.WeaponSkill_WeaponSkillStaves] = -20
		unit.PseudoStats.FeralCombatEnabled = tc.inForm
		unit.PseudoStats.FeralCombatSkill = tc.feralSkill
		if got := GetWeaponSkill(unit, tc.weapon); got != tc.weaponSkill {
			t.Errorf("%s: weapon skill %v, want %v", tc.name, got, tc.weaponSkill)
		}
	}
}

func TestFeralCombatRefreshesExistingAttackTables(t *testing.T) {
	character := NewCharacter(&Party{}, 0, &proto.Player{
		Class:     proto.Class_ClassDruid,
		Race:      proto.Race_RaceNightElf,
		Spec:      &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{}},
		Equipment: &proto.EquipmentSpec{},
		WeaponSkills: []*proto.WeaponSkillValue{{
			Skill: proto.WeaponSkill_WeaponSkillStaves,
			Value: 280,
		}},
	})
	staff := &Item{ID: 1, WeaponType: proto.WeaponType_WeaponTypeStaff, HandType: proto.HandType_HandTypeTwoHand}
	defender := &Unit{Type: EnemyUnit, Level: 63, PseudoStats: stats.NewPseudoStats()}
	character.AttackTables = []*AttackTable{NewAttackTable(&character.Unit, defender)}
	character.Equipment[proto.ItemSlot_ItemSlotMainHand] = *staff
	character.UpdateAttackTableWeapons()
	if got := character.AttackTables[0].MH.BaseMissChance; math.Abs(got-0.12) > 1e-9 {
		t.Fatalf("280 staff skill miss = %v, want 0.12", got)
	}

	character.SetFeralCombat(true)
	if got := character.AttackTables[0].MH.BaseMissChance; math.Abs(got-0.08) > 1e-9 {
		t.Fatalf("feral combat miss = %v, want 0.08", got)
	}

	character.SetFeralCombat(false)
	if got := character.AttackTables[0].MH.BaseMissChance; math.Abs(got-0.12) > 1e-9 {
		t.Fatalf("restored staff skill miss = %v, want 0.12", got)
	}
}

func TestEnemyAttackTableAgainstALevel60Player(t *testing.T) {
	for _, tc := range []struct {
		level                                       int32
		miss, dodge, parry, block, crush, spellMiss float64
	}{
		{58, 0.054, 0.004, 0.054, 0.054, 0, 0.06},
		{59, 0.052, 0.002, 0.052, 0.052, 0, 0.05},
		{60, 0.05, 0, 0.05, 0.05, 0, 0.04},
		{61, 0.048, -0.002, 0.048, 0.048, 0, 0.03},
		{62, 0.046, -0.004, 0.046, 0.046, 0, 0.02},
		{63, 0.044, -0.006, 0.044, 0.044, 0.15, 0.01},
	} {
		attacker, defender := attackTableUnits(EnemyUnit, tc.level, PlayerUnit, 60)
		table := NewAttackTable(attacker, defender)
		want := AttackTableHand{
			BaseMissChance:  tc.miss,
			BaseDodgeChance: tc.dodge,
			BaseParryChance: tc.parry,
			BaseBlockChance: tc.block,
		}
		values := []attackTableValue{
			{"crush", table.BaseCrushChance, tc.crush},
			{"spell miss", table.BaseSpellMissChance, tc.spellMiss},
		}
		for hand, entry := range map[string]AttackTableHand{"main hand": table.MH, "off hand": table.OH, "ranged": table.Ranged} {
			values = append(values, handValues(hand, entry, want)...)
		}
		checkAttackTableValues(t, fmt.Sprintf("level %d against 60", tc.level), values)
	}
}

func TestTargetBaseCritAndBlockValue(t *testing.T) {
	for _, tc := range []struct {
		level int32
		crit  float64
	}{
		{58, 5.0},
		{60, 5.0},
		{61, 5.2},
		{62, 5.4},
		{63, 5.6},
	} {
		target := NewTarget(&proto.Target{Level: tc.level}, 0)
		if got := target.GetStat(stats.PhysicalCritPercent); math.Abs(got-tc.crit) > 1e-9 {
			t.Errorf("level %d target crit = %v, want %v", tc.level, got, tc.crit)
		}
		if got := target.GetStat(stats.BlockValue); got != 0 {
			t.Errorf("level %d target block value = %v, want 0", tc.level, got)
		}
	}
}

func TestItemSwapRecomputesTheHandEntries(t *testing.T) {
	sim, character := NewWeaponTestSim(t, proto.Race_RaceUndead,
		WeaponTestSlots(WeaponTestSwordID, 0, 0),
		WeaponTestSlots(WeaponTestDaggerID, 0, 0),
		func(character *Character) {
			character.PseudoStats.WeaponSkills[proto.WeaponSkill_WeaponSkillSwords] = 5
		})
	table := character.AttackTables[character.CurrentTarget.UnitIndex]

	for _, step := range []struct {
		when string
		swap func()
		miss float64
	}{
		{"sword in the main hand", func() {}, 0.06},
		{"dagger swapped into the main hand", func() { character.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false) }, 0.08},
		{"reset to the sword", func() { character.ItemSwap.reset(sim) }, 0.06},
	} {
		step.swap()
		checkAttackTableValues(t, step.when, []attackTableValue{
			{"main hand miss", table.MH.BaseMissChance, step.miss},
		})
	}
}

func TestAttackTableHandFollowsTheProcMask(t *testing.T) {
	attacker, defender := attackTableUnits(PlayerUnit, 60, EnemyUnit, 63)
	table := NewAttackTable(attacker, defender)

	for _, tc := range []struct {
		name     string
		procMask ProcMask
		want     *AttackTableHand
	}{
		{"main hand auto", ProcMaskMeleeMHAuto, &table.MH},
		{"main hand special", ProcMaskMeleeMHSpecial, &table.MH},
		{"off hand auto", ProcMaskMeleeOHAuto, &table.OH},
		{"off hand special", ProcMaskMeleeOHSpecial, &table.OH},
		{"both hands", ProcMaskMeleeMHSpecial | ProcMaskMeleeOHSpecial, &table.MH},
		{"ranged auto", ProcMaskRangedAuto, &table.Ranged},
		{"ranged special", ProcMaskRangedSpecial, &table.Ranged},
		{"spell", ProcMaskSpellDamage, &table.MH},
	} {
		if got := table.Hand(&Spell{ProcMask: tc.procMask}); got != tc.want {
			t.Errorf("%s picks the wrong hand entry", tc.name)
		}
	}
}

func TestGlanceDamageRollsWithinTheHandsRange(t *testing.T) {
	sim := &Simulation{rand: NewSplitMix(42)}
	table := &AttackTable{
		MH: AttackTableHand{BaseGlanceChance: 1, GlanceMultiplierMin: 0.55, GlanceMultiplierMax: 0.75},
		OH: AttackTableHand{BaseGlanceChance: 1, GlanceMultiplierMin: 0.80, GlanceMultiplierMax: 0.90},
	}
	target := &Unit{}

	for _, tc := range []struct {
		name     string
		procMask ProcMask
		min, max float64
	}{
		{"main hand", ProcMaskMeleeMHAuto, 55, 75},
		{"off hand", ProcMaskMeleeOHAuto, 80, 90},
	} {
		spell := &Spell{ProcMask: tc.procMask, SpellMetrics: make([]SpellMetrics, 1)}
		lowest, highest := math.Inf(1), math.Inf(-1)
		for range 1000 {
			result := &SpellResult{Target: target, Damage: 100}
			chance := 0.0
			if !result.applyAttackTableGlance(sim, spell, table, 0, &chance) || result.Outcome != OutcomeGlance {
				t.Fatalf("%s: a roll of 0 against a 100%% glance chance did not glance", tc.name)
			}
			lowest, highest = min(lowest, result.Damage), max(highest, result.Damage)
		}
		if lowest < tc.min || highest > tc.max {
			t.Errorf("%s: glancing damage ranged %v to %v, want within %v to %v", tc.name, lowest, highest, tc.min, tc.max)
		}
		if spread := tc.max - tc.min; lowest > tc.min+spread/10 || highest < tc.max-spread/10 {
			t.Errorf("%s: glancing damage ranged %v to %v, want it to cover %v to %v", tc.name, lowest, highest, tc.min, tc.max)
		}
	}
}

// In front of the target, a weapon damage special rolls block on the attack table before its crit roll, so a
// blocked attack never crits and blocks come at the full block chance.
func TestWeaponSpecialBlockPreventsTheCrit(t *testing.T) {
	attacker, defender := attackTableUnits(PlayerUnit, 60, EnemyUnit, 63)
	attacker.PseudoStats.InFrontOfTarget = true
	attacker.stats[stats.PhysicalHitPercent] = 20
	attacker.stats[stats.PhysicalCritPercent] = 50
	table := NewAttackTable(attacker, defender)
	crit := (&Spell{Unit: attacker, ProcMask: ProcMaskMeleeMHSpecial}).PhysicalCritChance(table)
	landed := 1 - table.MH.BaseDodgeChance - table.MH.BaseParryChance - table.MH.BaseBlockChance

	for _, tc := range []struct {
		name      string
		outcome   func(*Spell) OutcomeApplier
		wantCrits float64
	}{
		{"weapon special", func(spell *Spell) OutcomeApplier { return spell.OutcomeMeleeWeaponSpecialHitAndCrit }, landed * crit},
		{"block and crit only", func(spell *Spell) OutcomeApplier { return spell.OutcomeMeleeSpecialBlockAndCrit }, (1 - table.MH.BaseBlockChance) * crit},
	} {
		sim := &Simulation{rand: NewSplitMix(42)}
		spell := &Spell{Unit: attacker, ProcMask: ProcMaskMeleeMHSpecial, DefenseType: DefenseTypeMelee, CritMultiplierPct: 1, SpellMetrics: make([]SpellMetrics, 1)}
		outcome := tc.outcome(spell)

		const n = 100000
		for range n {
			outcome(sim, &SpellResult{Target: defender, Damage: 100}, table)
		}

		metrics := spell.SpellMetrics[0]
		if metrics.BlockedCrits != 0 {
			t.Errorf("%s: %d blocked crits, want none", tc.name, metrics.BlockedCrits)
		}
		for _, check := range []struct {
			name      string
			got, want float64
		}{
			{"blocks", float64(metrics.Blocks) / n, table.MH.BaseBlockChance},
			{"crits", float64(metrics.Crits) / n, tc.wantCrits},
		} {
			if math.Abs(check.got-check.want) > 0.005 {
				t.Errorf("%s: %s %v of attacks, want %v", tc.name, check.name, check.got, check.want)
			}
		}
	}
}

func TestExpectedWhiteDamageUsesTheGlanceMidpoint(t *testing.T) {
	attacker := &Unit{PseudoStats: stats.NewPseudoStats()}
	target := &Unit{PseudoStats: stats.NewPseudoStats()}
	table := &AttackTable{
		Attacker:       attacker,
		Defender:       target,
		CritMultiplier: 1,
		MH:             AttackTableHand{BaseGlanceChance: 0.4, GlanceMultiplierMin: 0.55, GlanceMultiplierMax: 0.75},
	}
	spell := &Spell{Unit: attacker, ProcMask: ProcMaskMeleeMHAuto, DefenseType: DefenseTypeMelee, CritMultiplierPct: 1}

	result := &SpellResult{Target: target, Damage: 100}
	spell.OutcomeExpectedMeleeWhite(nil, result, table)
	if want := 100 * (1 - 0.4*(1-0.65)); math.Abs(result.Damage-want) > 1e-9 {
		t.Errorf("expected white damage %v, want %v from the 0.65 midpoint", result.Damage, want)
	}
}

func TestExpectedMeleeDamageBlocksNoMoreThanTheDamage(t *testing.T) {
	attacker := &Unit{PseudoStats: stats.NewPseudoStats()}
	attacker.PseudoStats.InFrontOfTarget = true
	target := &Unit{PseudoStats: stats.NewPseudoStats()}
	table := &AttackTable{
		Attacker:       attacker,
		Defender:       target,
		CritMultiplier: 1,
		MH:             AttackTableHand{BaseBlockChance: 0.05},
	}
	spell := &Spell{Unit: attacker, ProcMask: ProcMaskMeleeMHSpecial, DefenseType: DefenseTypeMelee, CritMultiplierPct: 1}

	for name, outcome := range map[string]OutcomeApplier{
		"white":          spell.OutcomeExpectedMeleeWhite,
		"weapon special": spell.OutcomeExpectedMeleeWeaponSpecialHitAndCrit,
	} {
		for _, tc := range []struct {
			blockValue, damage, want float64
		}{
			{20, 100, 99},
			{200, 100, 95},
			{200, 0, 0},
			{0, 0, 0},
		} {
			target.stats[stats.BlockValue] = tc.blockValue
			result := &SpellResult{Target: target, Damage: tc.damage}
			outcome(nil, result, table)
			if !(math.Abs(result.Damage-tc.want) <= 1e-9) {
				t.Errorf("%s, %v block value against %v damage: expected damage %v, want %v", name, tc.blockValue, tc.damage, result.Damage, tc.want)
			}
		}
	}
}

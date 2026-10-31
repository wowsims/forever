package spelldata

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func weaponDamageDoneRow(misc int32, subclass int32) *Spell {
	e := casterAura(dbcenums.A_MOD_DAMAGE_PERCENT_DONE, 10)
	e.Misc = misc
	return &Spell{ID: 1500, Name: "Weapon Specialization", DurationMs: -1,
		EquipClass: dbcenums.ITEM_CLASS_WEAPON, EquipSubclass: subclass, Effects: []Effect{e}}
}

func swordWarrior(hand proto.HandType) *core.Character {
	character := parseWarrior()
	character.Equipment[proto.ItemSlot_ItemSlotMainHand] = core.Item{
		ID: 1, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: hand}
	return character
}

func defenseSpell(character *core.Character, id int32, school core.SpellSchool, defense core.DefenseType) *core.Spell {
	spell := parseSpell(character, id, school, core.ProcMaskMeleeMHSpecial, core.ClassFlags{})
	spell.DefenseType = defense
	return spell
}

func physicalMultiplier(character *core.Character) float64 {
	return character.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical]
}

// The warrior's Two-Handed Weapon Specialization 12163 and the paladin's 20111 at 3/3 give their
// percent on Physical and on the melee-defense Holy spells such as the Seal of Righteousness hit, and
// nothing on a magic-defense Holy spell such as Consecration. A physical spell takes it through the
// school multiplier alone.
func TestWeaponSpecializationsReachMeleeDefenseHolySpells(t *testing.T) {
	defer generatedStore()()

	for _, tc := range []struct {
		row  *Spell
		hand proto.HandType
	}{
		{Talent(12163, 3).Rank(3), proto.HandType_HandTypeTwoHand},
		{Talent(20111, 3).Rank(3), proto.HandType_HandTypeTwoHand},
	} {
		want := 1 + tc.row.EffectN(1).Percent()
		character := swordWarrior(tc.hand)
		holyMelee := defenseSpell(character, 1, core.SpellSchoolHoly, core.DefenseTypeMelee)
		holyMagic := defenseSpell(character, 2, core.SpellSchoolHoly, core.DefenseTypeMagic)
		physical := defenseSpell(character, 3, core.SpellSchoolPhysical, core.DefenseTypeMelee)

		ParseStatic(character, tc.row)

		if got := holyMelee.DamageMultiplier; math.Abs(got-want) > 1e-9 {
			t.Errorf("%s: a melee-defense holy spell: %v, want %v", tc.row.Name, got, want)
		}
		if got := holyMagic.DamageMultiplier; got != 1 {
			t.Errorf("%s: a magic-defense holy spell: %v, want 1", tc.row.Name, got)
		}
		if got := physical.DamageMultiplier; got != 1 {
			t.Errorf("%s: a physical spell's own multiplier: %v, want 1", tc.row.Name, got)
		}
		if got := physicalMultiplier(character); math.Abs(got-want) > 1e-9 {
			t.Errorf("%s: the physical school multiplier: %v, want %v", tc.row.Name, got, want)
		}
	}
}

// A row that names every school adds nothing past its school multiplier, and a row that names no
// weapon holds whatever the character wears.
func TestOtherDamageDoneRowsStayOnTheirSchools(t *testing.T) {
	character := swordWarrior(proto.HandType_HandTypeOneHand)
	holyMelee := defenseSpell(character, 1, core.SpellSchoolHoly, core.DefenseTypeMelee)
	ParseStatic(character, weaponDamageDoneRow(127, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_1H))
	if got := holyMelee.DamageMultiplier; got != 1 {
		t.Errorf("every school: a melee-defense holy spell's own multiplier: %v, want 1", got)
	}

	noWeapon := weaponDamageDoneRow(1, 0)
	noWeapon.EquipClass = 0
	character = parseWarrior()
	ParseStatic(character, noWeapon)
	if got := physicalMultiplier(character); math.Abs(got-1.1) > 1e-9 {
		t.Errorf("no weapon, bare hands: the physical school multiplier: %v, want 1.1", got)
	}
}

// A row that names a weapon holds only while a matching weapon is equipped: One-Handed Weapon
// Specialization gives nothing with bare hands or a two-hander.
func TestWeaponDamageDoneRowNeedsItsWeapon(t *testing.T) {
	for name, character := range map[string]*core.Character{
		"bare hands": parseWarrior(),
		"two-hander": swordWarrior(proto.HandType_HandTypeTwoHand),
	} {
		holyMelee := defenseSpell(character, 1, core.SpellSchoolHoly, core.DefenseTypeMelee)

		ParseStatic(character, weaponDamageDoneRow(1, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_1H))

		if got := physicalMultiplier(character); got != 1 {
			t.Errorf("%s: the physical school multiplier: %v, want 1", name, got)
		}
		if got := holyMelee.DamageMultiplier; got != 1 {
			t.Errorf("%s: a melee-defense holy spell: %v, want 1", name, got)
		}
	}
}

// The row follows the weapon through item swaps: nothing with a bow alone, the bonus once a sword is
// swapped in, and nothing again once it is swapped out.
func TestWeaponDamageDoneRowFollowsItemSwaps(t *testing.T) {
	row := weaponDamageDoneRow(1, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_1H)
	r := parseScopedWith(t, row, core.WeaponTestSlots(0, 0, core.WeaponTestBowID),
		core.WeaponTestSlots(core.WeaponTestSwordID, 0, core.WeaponTestBowID), false)

	for _, step := range []struct {
		when string
		swap proto.APLActionItemSwap_SwapSet
		want float64
	}{
		{"bow alone", proto.APLActionItemSwap_Main, 1},
		{"sword swapped in", proto.APLActionItemSwap_Swap1, 1.1},
		{"sword swapped out", proto.APLActionItemSwap_Main, 1},
	} {
		r.character.ItemSwap.SwapItems(r.sim, step.swap, false)
		if got := physicalMultiplier(r.character); math.Abs(got-step.want) > 1e-9 {
			t.Errorf("%s: the physical school multiplier: %v, want %v", step.when, got, step.want)
		}
	}
}

// Ranged Weapon Specialization 19507 names every school and ranged weapons alone: a mod on the ranged
// attacks, which the melee attacks do not take.
func TestRangedWeaponDamageDoneRowIsARangedMod(t *testing.T) {
	r := parseScoped(t, weaponDamageDoneRow(127, dbcenums.ITEM_SUBCLASS_MASK_WEAPON_SHOOTING))

	for _, c := range []struct {
		name  string
		spell *core.Spell
		want  float64
	}{
		{"ranged auto", r.character.AutoAttacks.RangedAuto(), 1.1},
		{"main-hand auto", r.character.AutoAttacks.MHAuto(), 1},
	} {
		if got := c.spell.DamageMultiplier; math.Abs(got-c.want) > 1e-9 {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
		}
	}
	if got := r.character.PseudoStats.DamageDealtMultiplier; got != 1 {
		t.Errorf("the every-school multiplier: %v, want 1", got)
	}
}

// Applied names the melee-defense part beside the school multiplier.
func TestWeaponDamageDoneRowReportsItsMeleeDefensePart(t *testing.T) {
	defer generatedStore()()

	row := Talent(12163, 3).Rank(3)
	applied := DryRun(row).Applied
	if len(applied) != 2 {
		t.Fatalf("applied %+v, want the school multiplier and the melee-defense part", applied)
	}
	part := row.EffectN(1).Percent()
	for i, want := range []struct {
		kind  string
		value float64
	}{{"damage-dealt-by-school", 1 + part}, {"damage-dealt-melee", part}} {
		if applied[i].Kind != want.kind || math.Abs(applied[i].Value-want.value) > 1e-9 {
			t.Errorf("applied[%d] = %s %v, want %s %v", i, applied[i].Kind, applied[i].Value, want.kind, want.value)
		}
	}
}

// A melee-defense spell of several schools takes the part by the school its SchoolIndex reads. Under a
// Physical row, Physical and Holy, such as Infernal Strike 412790, reads Physical and takes none, its
// multiplier carrying the bonus; Holy and Shadow reads Holy and takes it, under a Shadow row as well.
func TestWeaponDamageDoneRowCountsOnceOnSeveralSchools(t *testing.T) {
	for _, tc := range []struct {
		row    core.SpellSchool
		school core.SpellSchool
		want   float64
	}{
		{core.SpellSchoolPhysical, core.SpellSchoolPhysical | core.SpellSchoolHoly, 1},
		{core.SpellSchoolPhysical, core.SpellSchoolHoly | core.SpellSchoolShadow, 1.1},
		{core.SpellSchoolShadow, core.SpellSchoolHoly | core.SpellSchoolShadow, 1.1},
	} {
		character := swordWarrior(proto.HandType_HandTypeOneHand)
		spell := defenseSpell(character, 1, tc.school, core.DefenseTypeMelee)

		ParseStatic(character, weaponDamageDoneRow(int32(tc.row), dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_1H))

		if got := spell.DamageMultiplier; math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("a %s spell under a %s row: %v, want %v", tc.school, tc.row, got, tc.want)
		}
	}
}

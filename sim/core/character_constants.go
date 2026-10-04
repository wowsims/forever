package core

import (
	"slices"

	"github.com/wowsims/forever/sim/core/proto"
)

// EligibleWeaponType describes whether a weapon type can be used as a two-hand weapon.
type EligibleWeaponType struct {
	CanUseTwoHand bool
}

var ClassArmorTypeCapabilities = map[proto.Class][]proto.ArmorType{
	proto.Class_ClassWarrior: {proto.ArmorType_ArmorTypePlate, proto.ArmorType_ArmorTypeMail, proto.ArmorType_ArmorTypeLeather, proto.ArmorType_ArmorTypeCloth},
	proto.Class_ClassPaladin: {proto.ArmorType_ArmorTypePlate, proto.ArmorType_ArmorTypeMail, proto.ArmorType_ArmorTypeLeather, proto.ArmorType_ArmorTypeCloth},
	proto.Class_ClassHunter:  {proto.ArmorType_ArmorTypeMail, proto.ArmorType_ArmorTypeLeather, proto.ArmorType_ArmorTypeCloth},
	proto.Class_ClassRogue:   {proto.ArmorType_ArmorTypeLeather, proto.ArmorType_ArmorTypeCloth},
	proto.Class_ClassPriest:  {proto.ArmorType_ArmorTypeCloth},
	proto.Class_ClassShaman:  {proto.ArmorType_ArmorTypeMail, proto.ArmorType_ArmorTypeLeather, proto.ArmorType_ArmorTypeCloth},
	proto.Class_ClassMage:    {proto.ArmorType_ArmorTypeCloth},
	proto.Class_ClassWarlock: {proto.ArmorType_ArmorTypeCloth},
	proto.Class_ClassDruid:   {proto.ArmorType_ArmorTypeLeather, proto.ArmorType_ArmorTypeCloth},
}

var ClassWeaponTypeCapabilities = map[proto.Class]map[proto.WeaponType]EligibleWeaponType{
	proto.Class_ClassWarrior: {
		proto.WeaponType_WeaponTypeAxe:     {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeDagger:  {},
		proto.WeaponType_WeaponTypeFist:    {},
		proto.WeaponType_WeaponTypeMace:    {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeOffHand: {},
		proto.WeaponType_WeaponTypePolearm: {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeShield:  {},
		proto.WeaponType_WeaponTypeStaff:   {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeSword:   {CanUseTwoHand: true},
	},
	proto.Class_ClassPaladin: {
		proto.WeaponType_WeaponTypeAxe:     {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeMace:    {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypePolearm: {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeShield:  {},
		proto.WeaponType_WeaponTypeSword:   {CanUseTwoHand: true},
	},
	proto.Class_ClassHunter: {
		proto.WeaponType_WeaponTypeAxe:     {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeDagger:  {},
		proto.WeaponType_WeaponTypeFist:    {},
		proto.WeaponType_WeaponTypePolearm: {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeStaff:   {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeSword:   {CanUseTwoHand: true},
	},
	proto.Class_ClassRogue: {
		proto.WeaponType_WeaponTypeAxe:     {},
		proto.WeaponType_WeaponTypeDagger:  {},
		proto.WeaponType_WeaponTypeFist:    {},
		proto.WeaponType_WeaponTypeMace:    {},
		proto.WeaponType_WeaponTypeOffHand: {},
		proto.WeaponType_WeaponTypeSword:   {},
	},
	proto.Class_ClassDruid: {
		proto.WeaponType_WeaponTypeDagger:  {},
		proto.WeaponType_WeaponTypeFist:    {},
		proto.WeaponType_WeaponTypeMace:    {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeOffHand: {},
		proto.WeaponType_WeaponTypePolearm: {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeStaff:   {CanUseTwoHand: true},
	},
	proto.Class_ClassShaman: {
		proto.WeaponType_WeaponTypeAxe:     {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeDagger:  {},
		proto.WeaponType_WeaponTypeFist:    {},
		proto.WeaponType_WeaponTypeMace:    {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeOffHand: {},
		proto.WeaponType_WeaponTypeShield:  {},
		proto.WeaponType_WeaponTypeStaff:   {CanUseTwoHand: true},
	},
	proto.Class_ClassMage: {
		proto.WeaponType_WeaponTypeDagger:  {},
		proto.WeaponType_WeaponTypeOffHand: {},
		proto.WeaponType_WeaponTypeStaff:   {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeSword:   {},
	},
	proto.Class_ClassPriest: {
		proto.WeaponType_WeaponTypeDagger:  {},
		proto.WeaponType_WeaponTypeMace:    {},
		proto.WeaponType_WeaponTypeOffHand: {},
		proto.WeaponType_WeaponTypeStaff:   {CanUseTwoHand: true},
	},
	proto.Class_ClassWarlock: {
		proto.WeaponType_WeaponTypeDagger:  {},
		proto.WeaponType_WeaponTypeOffHand: {},
		proto.WeaponType_WeaponTypeStaff:   {CanUseTwoHand: true},
		proto.WeaponType_WeaponTypeSword:   {},
	},
}

var ClassRangedWeaponTypeCapabilities = map[proto.Class][]proto.RangedWeaponType{
	proto.Class_ClassWarrior: {proto.RangedWeaponType_RangedWeaponTypeBow, proto.RangedWeaponType_RangedWeaponTypeCrossbow, proto.RangedWeaponType_RangedWeaponTypeGun, proto.RangedWeaponType_RangedWeaponTypeThrown},
	proto.Class_ClassRogue:   {proto.RangedWeaponType_RangedWeaponTypeBow, proto.RangedWeaponType_RangedWeaponTypeCrossbow, proto.RangedWeaponType_RangedWeaponTypeGun, proto.RangedWeaponType_RangedWeaponTypeThrown},
	proto.Class_ClassHunter:  {proto.RangedWeaponType_RangedWeaponTypeBow, proto.RangedWeaponType_RangedWeaponTypeCrossbow, proto.RangedWeaponType_RangedWeaponTypeGun},
	proto.Class_ClassMage:    {proto.RangedWeaponType_RangedWeaponTypeWand},
	proto.Class_ClassPriest:  {proto.RangedWeaponType_RangedWeaponTypeWand},
	proto.Class_ClassWarlock: {proto.RangedWeaponType_RangedWeaponTypeWand},
	proto.Class_ClassDruid:   {proto.RangedWeaponType_RangedWeaponTypeIdol},
	proto.Class_ClassPaladin: {proto.RangedWeaponType_RangedWeaponTypeLibram},
	proto.Class_ClassShaman:  {proto.RangedWeaponType_RangedWeaponTypeTotem},
}

var ClassRaceCapabilities = map[proto.Class][]proto.Race{
	proto.Class_ClassWarrior: {
		proto.Race_RaceHuman,
		proto.Race_RaceDwarf,
		proto.Race_RaceNightElf,
		proto.Race_RaceGnome,
		proto.Race_RaceHighOrderSkyborne,
		proto.Race_RaceOrc,
		proto.Race_RaceUndead,
		proto.Race_RaceTauren,
		proto.Race_RaceTroll,
		proto.Race_RaceWindshaperSkyborne,
	},
	proto.Class_ClassPaladin: {
		proto.Race_RaceHuman,
		proto.Race_RaceDwarf,
		proto.Race_RaceUndead,
	},
	proto.Class_ClassHunter: {
		proto.Race_RaceHuman,
		proto.Race_RaceDwarf,
		proto.Race_RaceNightElf,
		proto.Race_RaceHighOrderSkyborne,
		proto.Race_RaceOrc,
		proto.Race_RaceTauren,
		proto.Race_RaceTroll,
		proto.Race_RaceWindshaperSkyborne,
	},
	proto.Class_ClassRogue: {
		proto.Race_RaceHuman,
		proto.Race_RaceDwarf,
		proto.Race_RaceNightElf,
		proto.Race_RaceGnome,
		proto.Race_RaceHighOrderSkyborne,
		proto.Race_RaceOrc,
		proto.Race_RaceUndead,
		proto.Race_RaceTroll,
		proto.Race_RaceWindshaperSkyborne,
	},
	proto.Class_ClassPriest: {
		proto.Race_RaceHuman,
		proto.Race_RaceDwarf,
		proto.Race_RaceNightElf,
		proto.Race_RaceGnome,
		proto.Race_RaceUndead,
		proto.Race_RaceTroll,
	},
	proto.Class_ClassShaman: {
		proto.Race_RaceDwarf,
		proto.Race_RaceOrc,
		proto.Race_RaceTauren,
		proto.Race_RaceTroll,
		proto.Race_RaceWindshaperSkyborne,
	},
	proto.Class_ClassMage: {
		proto.Race_RaceHuman,
		proto.Race_RaceGnome,
		proto.Race_RaceHighOrderSkyborne,
		proto.Race_RaceOrc,
		proto.Race_RaceUndead,
		proto.Race_RaceTroll,
	},
	proto.Class_ClassWarlock: {
		proto.Race_RaceHuman,
		proto.Race_RaceGnome,
		proto.Race_RaceOrc,
		proto.Race_RaceUndead,
		proto.Race_RaceTroll,
	},
	proto.Class_ClassDruid: {
		proto.Race_RaceNightElf,
		proto.Race_RaceHighOrderSkyborne,
		proto.Race_RaceTauren,
		proto.Race_RaceWindshaperSkyborne,
	},
}

var SpecCanDualWieldCapabilities = map[proto.Spec]bool{
	proto.Spec_SpecHunter:            true,
	proto.Spec_SpecRogue:             true,
	proto.Spec_SpecEnhancementShaman: true,
	proto.Spec_SpecDpsWarrior:        true,
	proto.Spec_SpecProtectionWarrior: true,
}

var oneHandWeaponSkills = map[proto.WeaponType]proto.WeaponSkill{
	proto.WeaponType_WeaponTypeAxe:    proto.WeaponSkill_WeaponSkillAxes,
	proto.WeaponType_WeaponTypeDagger: proto.WeaponSkill_WeaponSkillDaggers,
	proto.WeaponType_WeaponTypeFist:   proto.WeaponSkill_WeaponSkillFistWeapons,
	proto.WeaponType_WeaponTypeMace:   proto.WeaponSkill_WeaponSkillMaces,
	proto.WeaponType_WeaponTypeSword:  proto.WeaponSkill_WeaponSkillSwords,
}

var twoHandWeaponSkills = map[proto.WeaponType]proto.WeaponSkill{
	proto.WeaponType_WeaponTypeAxe:     proto.WeaponSkill_WeaponSkillTwoHandedAxes,
	proto.WeaponType_WeaponTypeMace:    proto.WeaponSkill_WeaponSkillTwoHandedMaces,
	proto.WeaponType_WeaponTypePolearm: proto.WeaponSkill_WeaponSkillPolearms,
	proto.WeaponType_WeaponTypeStaff:   proto.WeaponSkill_WeaponSkillStaves,
	proto.WeaponType_WeaponTypeSword:   proto.WeaponSkill_WeaponSkillTwoHandedSwords,
}

var rangedWeaponSkills = map[proto.RangedWeaponType]proto.WeaponSkill{
	proto.RangedWeaponType_RangedWeaponTypeBow:      proto.WeaponSkill_WeaponSkillBows,
	proto.RangedWeaponType_RangedWeaponTypeCrossbow: proto.WeaponSkill_WeaponSkillCrossbows,
	proto.RangedWeaponType_RangedWeaponTypeGun:      proto.WeaponSkill_WeaponSkillGuns,
	proto.RangedWeaponType_RangedWeaponTypeThrown:   proto.WeaponSkill_WeaponSkillThrown,
}

// ClassWeaponSkills lists the weapon skills of the weapons the class can use, sorted. Race plays no part: in the
// game files (SkillRaceClassInfo) race only decides which weapon skills a character starts with.
func ClassWeaponSkills(class proto.Class) []proto.WeaponSkill {
	var skills []proto.WeaponSkill
	for weaponType, eligible := range ClassWeaponTypeCapabilities[class] {
		if skill, ok := oneHandWeaponSkills[weaponType]; ok {
			skills = append(skills, skill)
		}
		if skill, ok := twoHandWeaponSkills[weaponType]; ok && eligible.CanUseTwoHand {
			skills = append(skills, skill)
		}
	}
	for _, rangedWeaponType := range ClassRangedWeaponTypeCapabilities[class] {
		if skill, ok := rangedWeaponSkills[rangedWeaponType]; ok {
			skills = append(skills, skill)
		}
	}
	slices.Sort(skills)
	return skills
}

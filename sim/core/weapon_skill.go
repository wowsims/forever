package core

import (
	"math"

	"github.com/wowsims/forever/sim/core/proto"
)

// GetWeaponSkill returns the unit's bonus weapon skill for the weapon, on top of its level-based skill.
func GetWeaponSkill(unit *Unit, weapon *Item) float64 {
	if unit.PseudoStats.FeralCombatEnabled {
		return unit.PseudoStats.FeralCombatSkill
	}

	if weapon == nil {
		return 0
	}

	return unit.PseudoStats.WeaponSkills[weaponSkillOf(weapon)]
}

func weaponSkillOf(weapon *Item) proto.WeaponSkill {
	if weapon.HandType == proto.HandType_HandTypeTwoHand {
		return twoHandWeaponSkills[weapon.WeaponType]
	}
	if weapon.RangedWeaponType != proto.RangedWeaponType_RangedWeaponTypeUnknown {
		return rangedWeaponSkills[weapon.RangedWeaponType]
	}
	return oneHandWeaponSkills[weapon.WeaponType]
}

// The full skill per weapon skill: the level-based maximum and every bonus on top of it.
func (character *Character) weaponSkillsProto() []*proto.WeaponSkillValue {
	var skills []*proto.WeaponSkillValue
	for skill := proto.WeaponSkill_WeaponSkillAxes; int(skill) < len(character.PseudoStats.WeaponSkills); skill++ {
		skills = append(skills, &proto.WeaponSkillValue{
			Skill: skill,
			Value: character.Level*5 + int32(math.Round(character.PseudoStats.WeaponSkills[skill])),
		})
	}
	return skills
}

// Sets each listed skill's bonus over the level-based skill; a skill the list leaves out stays at the maximum.
func (character *Character) applyWeaponSkills(weaponSkills []*proto.WeaponSkillValue) {
	for _, weaponSkill := range weaponSkills {
		if weaponSkill.Skill > proto.WeaponSkill_WeaponSkillUnknown && int(weaponSkill.Skill) < len(character.PseudoStats.WeaponSkills) {
			character.PseudoStats.WeaponSkills[weaponSkill.Skill] = float64(weaponSkill.Value - character.Level*5)
		}
	}
}

// Rebuilds the hand entries of the attack tables against every enemy, which read the weapons and their
// weapon skill.
func (character *Character) UpdateAttackTableWeapons() {
	mh, oh, ranged := character.GetMHWeapon(), character.GetOHWeapon(), character.GetRangedWeapon()
	for _, table := range character.AttackTables {
		if table.Defender.Type == EnemyUnit {
			table.setWeapons(mh, oh, ranged)
		}
	}
}

// Feral forms replace equipped weapon skill and refresh existing attack tables.
func (character *Character) SetFeralCombat(enabled bool) {
	character.PseudoStats.FeralCombatEnabled = enabled
	character.UpdateAttackTableWeapons()
}

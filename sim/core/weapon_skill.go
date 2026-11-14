package core

import (
	"math"

	"github.com/wowsims/forever/sim/core/proto"
)

// The unit's bonus weapon skill for the weapon held in the slot, on top of its level-based skill. A
// main-hand or off-hand weapon adds that hand's bonus; an empty hand and a feral form take none.
func GetWeaponSkill(unit *Unit, weapon *Item, slot proto.ItemSlot) float64 {
	if unit.PseudoStats.FeralCombatEnabled {
		return unit.PseudoStats.FeralCombatSkill
	}

	if weapon == nil {
		return 0
	}

	skill := unit.PseudoStats.WeaponSkills[weaponSkillOf(weapon)]
	switch slot {
	case proto.ItemSlot_ItemSlotMainHand:
		skill += unit.PseudoStats.MHWeaponSkill
	case proto.ItemSlot_ItemSlotOffHand:
		skill += unit.PseudoStats.OHWeaponSkill
	}
	return skill
}

func weaponSkillOf(weapon *Item) proto.WeaponSkill {
	if weapon.HandType == proto.HandType_HandTypeTwoHand {
		return TwoHandWeaponSkills[weapon.WeaponType]
	}
	if weapon.RangedWeaponType != proto.RangedWeaponType_RangedWeaponTypeUnknown {
		return rangedWeaponSkills[weapon.RangedWeaponType]
	}
	return OneHandWeaponSkills[weapon.WeaponType]
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
		if character.keepsWeaponSkill(weaponSkill.Skill) {
			character.PseudoStats.WeaponSkills[weaponSkill.Skill] = float64(weaponSkill.Value - character.Level*5)
		}
	}
}

// An item's weapon skill goes on top of the stated skill, and every copy worn counts; sign -1 takes it off.
func (character *Character) addItemWeaponSkills(item *Item, sign float64) {
	for _, weaponSkill := range item.ScalingOptions[0].GetWeaponSkills() {
		if character.keepsWeaponSkill(weaponSkill.Skill) {
			character.PseudoStats.WeaponSkills[weaponSkill.Skill] += sign * float64(weaponSkill.Value)
		}
	}
}

// Whether the skill has a slot in the unit's weapon skill bonuses.
func (character *Character) keepsWeaponSkill(skill proto.WeaponSkill) bool {
	return skill > proto.WeaponSkill_WeaponSkillUnknown && int(skill) < len(character.PseudoStats.WeaponSkills)
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

package spelldata

import (
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// The stats each versus aura grants against the creature types its misc value masks. Flat damage done
// against them reaches every school: Beastslayer 7784 adds its 2 to weapon hits and spells alike.
var versusStats = map[dbcenums.EffectAuraType][]stats.Stat{
	dbcenums.A_MOD_MELEE_ATTACK_POWER_VERSUS:  {stats.AttackPower},
	dbcenums.A_MOD_RANGED_ATTACK_POWER_VERSUS: {stats.RangedAttackPower},
	dbcenums.A_MOD_FLAT_SPELL_DAMAGE_VERSUS:   {stats.SpellDamage},
	dbcenums.A_MOD_DAMAGE_DONE_CREATURE:       {stats.PhysicalDamage, stats.SpellDamage},
}

// The stats a versus aura grants; none when the aura is no versus aura.
func VersusStats(aura dbcenums.EffectAuraType) []stats.Stat {
	return versusStats[aura]
}

// The mob type of each creature type a mask names by the bit 1 << (CreatureType - 1). Critter, 8, has
// no mob type.
var creatureTypeMobTypes = [...]proto.MobType{
	1: proto.MobType_MobTypeBeast,
	2: proto.MobType_MobTypeDragonkin,
	3: proto.MobType_MobTypeDemon,
	4: proto.MobType_MobTypeElemental,
	5: proto.MobType_MobTypeGiant,
	6: proto.MobType_MobTypeUndead,
	7: proto.MobType_MobTypeHumanoid,
	9: proto.MobType_MobTypeMechanical,
}

// The mob types a creature type mask names, as the versus auras' misc value and
// TargetCreatureType state it: 4 is Demon, 32 Undead.
func MobTypes(mask int32) []proto.MobType {
	var out []proto.MobType
	for creatureType, mobType := range creatureTypeMobTypes {
		if mobType != proto.MobType_MobTypeUnknown && mask&(1<<(creatureType-1)) != 0 {
			out = append(out, mobType)
		}
	}
	return out
}

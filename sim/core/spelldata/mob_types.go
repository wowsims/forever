package spelldata

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// What a versus aura gives against the creature types its misc value masks: stats at the effect's
// amount, or a multiplier of one plus the amount in percent.
type VersusAuraEffect struct {
	Stats      []stats.Stat
	Multiplier VersusMultiplier
}

// What a versus aura's multiplier scales.
type VersusMultiplier int

const (
	VersusMultiplierDamage     VersusMultiplier = iota + 1 // The damage dealt.
	VersusMultiplierCritDamage                             // The crit damage dealt, on top of the damage multiplier.
)

// Every versus aura and what it gives. Flat damage done against creature types reaches every school:
// Beastslayer 7784 adds its 2 to weapon hits and spells alike.
var versusAuras = map[dbcenums.EffectAuraType]VersusAuraEffect{
	dbcenums.A_MOD_MELEE_ATTACK_POWER_VERSUS:  {Stats: []stats.Stat{stats.AttackPower}},
	dbcenums.A_MOD_RANGED_ATTACK_POWER_VERSUS: {Stats: []stats.Stat{stats.RangedAttackPower}},
	dbcenums.A_MOD_FLAT_SPELL_DAMAGE_VERSUS:   {Stats: []stats.Stat{stats.SpellDamage}},
	dbcenums.A_MOD_DAMAGE_DONE_CREATURE:       {Stats: []stats.Stat{stats.PhysicalDamage, stats.SpellDamage}},
	dbcenums.A_MOD_DAMAGE_DONE_VERSUS:         {Multiplier: VersusMultiplierDamage},
	dbcenums.A_MOD_CRIT_PERCENT_VERSUS:        {Multiplier: VersusMultiplierCritDamage},
}

// What the aura gives against the creature types its misc value masks, and false when it is no versus
// aura.
func VersusAura(aura dbcenums.EffectAuraType) (VersusAuraEffect, bool) {
	versus, ok := versusAuras[aura]
	return versus, ok
}

// Each stat the aura grants, at the amount; none for a multiplier.
func (v VersusAuraEffect) Bonus(amount float64) stats.Stats {
	var bonus stats.Stats
	for _, stat := range v.Stats {
		bonus[stat] = amount
	}
	return bonus
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

// The targets a creature type mask names, as MobTypes decodes it.
func TargetCreatures(mask int32) core.VersusTargets {
	return core.VersusTargets{MobTypes: MobTypes(mask)}
}

// The targets the row's TargetCreatureType limits it to, none where it states no limit.
func (s *Spell) TargetCreatureTypes() core.VersusTargets {
	return TargetCreatures(s.TargetCreatureType)
}

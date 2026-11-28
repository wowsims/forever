package spelldata

import (
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// The stats a versus aura gives at the effect's amount against the creature types its misc value masks.
type VersusAuraEffect struct {
	Stats []stats.Stat
}

// Every versus aura and the stats it gives; the damage and crit damage multipliers give none. Flat
// damage done against creature types reaches every school.
var versusAuras = map[dbcenums.EffectAuraType]VersusAuraEffect{
	dbcenums.A_MOD_MELEE_ATTACK_POWER_VERSUS:  {Stats: []stats.Stat{stats.AttackPower}},
	dbcenums.A_MOD_RANGED_ATTACK_POWER_VERSUS: {Stats: []stats.Stat{stats.RangedAttackPower}},
	dbcenums.A_MOD_FLAT_SPELL_DAMAGE_VERSUS:   {Stats: []stats.Stat{stats.SpellDamage}},
	dbcenums.A_MOD_DAMAGE_DONE_CREATURE:       {Stats: []stats.Stat{stats.PhysicalDamage, stats.SpellDamage}},
	dbcenums.A_MOD_DAMAGE_DONE_VERSUS:         {},
	dbcenums.A_MOD_CRIT_PERCENT_VERSUS:        {},
}

// The stats the aura gives against the creature types its misc value masks; none for a multiplier or
// an aura that is no versus aura.
func VersusAura(aura dbcenums.EffectAuraType) VersusAuraEffect {
	return versusAuras[aura]
}

// Whether the aura gives anything against the creature types its misc value masks.
func IsVersusAura(aura dbcenums.EffectAuraType) bool {
	_, ok := versusAuras[aura]
	return ok
}

// The versus auras that grant stats, in aura order.
func VersusStatAuras() []dbcenums.EffectAuraType {
	var auras []dbcenums.EffectAuraType
	for aura, versus := range versusAuras {
		if len(versus.Stats) > 0 {
			auras = append(auras, aura)
		}
	}
	slices.Sort(auras)
	return auras
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
// TargetCreatureType state it.
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

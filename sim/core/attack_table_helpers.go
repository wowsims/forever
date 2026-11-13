package core

import (
	"maps"
	"math"
	"slices"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// The attack tables exist once the environment is set up, so the Unit helpers act on the tables of
// that moment: a caller runs them from a post-finalize effect or an aura's gain and expire.

// What the unit has against one mob type. The attack tables hold it only for the defenders the
// encounter has, so the helpers below keep it per mob type as well, from the moment the tables exist,
// for the stats sheet to list every mob type whatever the targets are. Nothing in combat reads it.
type mobTypeBonus struct {
	stats            stats.Stats
	damageMultiplier float64
	critMultiplier   float64
}

// Changes the unit's bonus against each of the mob types once, however often the list names one, as
// the attack tables count it.
func (unit *Unit) changeMobTypeBonuses(mobTypes []proto.MobType, change func(bonus *mobTypeBonus)) {
	if unit.AttackTables == nil {
		return
	}
	if unit.mobTypeBonuses == nil {
		unit.mobTypeBonuses = make(map[proto.MobType]*mobTypeBonus)
	}
	for i, mobType := range mobTypes {
		if slices.Index(mobTypes, mobType) < i {
			continue
		}
		bonus := unit.mobTypeBonuses[mobType]
		if bonus == nil {
			bonus = &mobTypeBonus{damageMultiplier: 1, critMultiplier: 1}
			unit.mobTypeBonuses[mobType] = bonus
		}
		change(bonus)
	}
}

// Adds stats the unit has against every defender of the mob types; a negative bonus takes them away.
func (unit *Unit) AddMobTypeBonusStats(bonus stats.Stats, mobTypes ...proto.MobType) {
	for _, at := range unit.AttackTables {
		if mobType := at.Defender.MobType; slices.Contains(mobTypes, mobType) {
			at.MobTypeBonusStats[mobType] = at.MobTypeBonusStats[mobType].Add(bonus)
		}
	}
	unit.changeMobTypeBonuses(mobTypes, func(versus *mobTypeBonus) { versus.stats = versus.stats.Add(bonus) })
}

// An item's stats against mob types, every copy worn counting; sign -1 takes them off.
func (character *Character) addItemMobTypeStats(item *Item, sign float64) {
	for _, versus := range item.ScalingOptions[0].GetMobTypeStats() {
		character.AddMobTypeBonusStats(stats.FromProtoMap(versus.Stats).Multiply(sign), versus.MobTypes...)
	}
}

// Multiplies the damage the unit deals to every defender of the mob types; 1/mult takes the
// multiplier back off.
func (unit *Unit) MultiplyDamageVersus(mult float64, mobTypes ...proto.MobType) {
	for _, at := range unit.AttackTables {
		if slices.Contains(mobTypes, at.Defender.MobType) {
			at.DamageDealtMultiplier *= mult
		}
	}
	unit.changeMobTypeBonuses(mobTypes, func(versus *mobTypeBonus) { versus.damageMultiplier *= mult })
}

// Multiplies the crit damage the unit deals to every defender of the mob types on top of the damage
// multiplier; 1/mult takes it back off.
func (unit *Unit) MultiplyCritDamageVersus(mult float64, mobTypes ...proto.MobType) {
	for _, at := range unit.AttackTables {
		if slices.Contains(mobTypes, at.Defender.MobType) {
			at.CritMultiplier *= mult
		}
	}
	unit.changeMobTypeBonuses(mobTypes, func(versus *mobTypeBonus) { versus.critMultiplier *= mult })
}

// Adds to the share of every defender's armor the unit's attacks ignore; a negative delta takes it away.
func (unit *Unit) AddArmorIgnoreFactor(delta float64) {
	for _, at := range unit.AttackTables {
		at.ArmorIgnoreFactor += delta
	}
}

// The stats against the mob types while the aura is up.
// Returns parent aura for chaining
func (parentAura *Aura) AttachMobTypeBonusStats(bonus stats.Stats, mobTypes ...proto.MobType) *Aura {
	parentAura.ApplyOnGain(func(aura *Aura, _ *Simulation) {
		aura.Unit.AddMobTypeBonusStats(bonus, mobTypes...)
	})

	parentAura.ApplyOnExpire(func(aura *Aura, _ *Simulation) {
		aura.Unit.AddMobTypeBonusStats(bonus.Invert(), mobTypes...)
	})

	if parentAura.IsActive() {
		parentAura.Unit.AddMobTypeBonusStats(bonus, mobTypes...)
	}

	return parentAura
}

// The unit's bonuses against each mob type, leaving out a mob type with none. A bonus taken back off
// can leave a rounding error, which counts as none.
func (unit *Unit) mobTypeBonusesProto() []*proto.MobTypeBonus {
	const epsilon = 1e-9
	var bonuses []*proto.MobTypeBonus
	for _, mobType := range slices.Sorted(maps.Keys(unit.mobTypeBonuses)) {
		bonus := unit.mobTypeBonuses[mobType]
		statMap := map[int32]float64{}
		for stat, value := range bonus.stats[:stats.ProtoStatsLen] {
			if math.Abs(value) > epsilon {
				statMap[int32(stat)] = value
			}
		}
		if len(statMap) > 0 || math.Abs(bonus.damageMultiplier-1) > epsilon || math.Abs(bonus.critMultiplier-1) > epsilon {
			bonuses = append(bonuses, &proto.MobTypeBonus{
				MobType:              mobType,
				Stats:                statMap,
				DamageMultiplier:     bonus.damageMultiplier,
				CritDamageMultiplier: bonus.critMultiplier,
			})
		}
	}
	return bonuses
}

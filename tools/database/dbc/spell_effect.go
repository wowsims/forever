package dbc

import (
	"math"
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

const MAX_SCALING_LEVEL = 100
const BASE_LEVEL = 60

type SpellEffect struct {
	ID                             int
	DifficultyID                   int
	EffectIndex                    int
	EffectType                     SpellEffectType
	EffectAmplitude                float64
	EffectAttributes               int
	EffectAura                     EffectAuraType
	EffectAuraPeriod               int
	EffectBasePoints               float64
	EffectBonusCoefficient         float64
	EffectChainAmplitude           float64
	EffectChainTargets             int
	EffectDieSides                 float64
	EffectItemType                 int
	EffectMechanic                 int
	EffectPointsPerResource        float64
	EffectPosFacing                float64
	EffectRealPointsPerLevel       float64
	EffectTriggerSpell             int
	BonusCoefficientFromAP         float64
	PvpMultiplier                  float64
	Coefficient                    float64
	Variance                       float64
	ResourceCoefficient            float64
	GroupSizeBasePointsCoefficient float64
	// Grouped properties parsed from JSON strings:
	EffectMiscValues      []int     // from EffectMiscValue_0, EffectMiscValue_1
	EffectMinRange        []float64 // from EffectRadiusIndex_0, EffectRadiusIndex_1
	EffectMaxRange        []float64
	EffectSpellClassMasks []int // from EffectSpellClassMask_0, EffectSpellClassMask_1, EffectSpellClassMask_2, EffectSpellClassMask_3
	ImplicitTargets       []int // from ImplicitTarget_0, ImplicitTarget_1
	SpellID               int
	ScalingType           int
}

func (se *SpellEffect) ToProto() *proto.SpellEffect {
	spellEffect := &proto.SpellEffect{
		Id:           int32(se.ID),
		SpellId:      int32(se.SpellID),
		Index:        int32(se.EffectIndex),
		Type:         proto.EffectType(se.EffectType),
		AuraPeriodMs: int32(se.EffectAuraPeriod),
		DurationMs:   int32(dbcInstance.Spells[se.SpellID].Duration),
	}
	if effectStats, ok := se.ParseStatEffect(false, 0); ok {
		spellEffect.Stats = effectStats.ToProtoArray()
	}
	switch se.EffectType {
	case dbcenums.E_ENERGIZE, dbcenums.E_ENERGIZE_PCT:
		spellEffect.MiscValue0 = &proto.SpellEffect_ResourceType{ResourceType: MapPowerTypeEnumToResourceType[int32(se.EffectMiscValues[0])]}
	case dbcenums.E_HEAL, dbcenums.E_HEAL_PCT:
		spellEffect.MiscValue0 = &proto.SpellEffect_ResourceType{ResourceType: proto.ResourceType_ResourceTypeHealth}
	case dbcenums.E_APPLY_AURA:
		if se.EffectAura == dbcenums.A_PERIODIC_ENERGIZE {
			spellEffect.MiscValue0 = &proto.SpellEffect_ResourceType{ResourceType: MapPowerTypeEnumToResourceType[int32(se.EffectMiscValues[0])]}
		} else if se.EffectAura == dbcenums.A_PERIODIC_HEAL {
			spellEffect.MiscValue0 = &proto.SpellEffect_ResourceType{ResourceType: proto.ResourceType_ResourceTypeHealth}
		}
	}

	return spellEffect
}

func (s *SpellEffect) GetRadiusMax() float64 {
	return math.Max(s.EffectMaxRange[0], s.EffectMaxRange[1])
}

func (s *SpellEffect) GetRadiusMin() float64 {
	return math.Min(s.EffectMinRange[0], s.EffectMinRange[1])
}

func (s *SpellEffect) ScalingClass() proto.Class {
	switch s.ScalingType {
	case 1:
		return proto.Class_ClassWarrior
	case 2:
		return proto.Class_ClassPaladin
	case 3:
		return proto.Class_ClassHunter
	case 4:
		return proto.Class_ClassRogue
	case 5:
		return proto.Class_ClassPriest
	case 7:
		return proto.Class_ClassShaman
	case 8:
		return proto.Class_ClassMage
	case 9:
		return proto.Class_ClassWarlock
	case 11:
		return proto.Class_ClassDruid
	case -1:
		return proto.Class_ClassExtra1
	case -2:
		return proto.Class_ClassExtra2
	case -3:
		return proto.Class_ClassExtra3
	case -4:
		return proto.Class_ClassExtra4
	case -5:
		return proto.Class_ClassExtra5
	case -6:
		return proto.Class_ClassExtra6
	default:
		return proto.Class_ClassUnknown
	}
}

// Reports whether the aura fires another spell when its owner procs. Both forms are the same edge
// for the purposes of walking a trigger chain; only the amount differs.
func (effect *SpellEffect) IsProcTrigger() bool {
	return effect.EffectAura.IsProcTrigger()
}

// Reports whether the aura resolves to nothing on its own. Either it applies in game as a
// marker, or its real behaviour is server-scripted and not present in DBC at all.
func (effect *SpellEffect) IsDummy() bool {
	return effect.EffectAura == dbcenums.A_DUMMY || effect.EffectAura == dbcenums.A_PERIODIC_DUMMY
}

// The aura types capable of resolving to stats. resolveStatsSpell walks a trigger chain looking for
// one of these to decide which spell an item's stats are read off, so anything ParseStatEffect
// learns to resolve has to be listed here as well or the chain walk stops short of it. The reverse
// need not hold: A_MOD_TARGET_RESISTANCE ends the walk on a spell ParseStatEffect then declines,
// which is the existing behaviour.
var statAuraTypes = map[EffectAuraType]bool{
	dbcenums.A_MOD_STAT:                true,
	dbcenums.A_MOD_RATING:              true,
	dbcenums.A_MOD_RANGED_ATTACK_POWER: true,
	dbcenums.A_MOD_ATTACK_POWER:        true,
	dbcenums.A_MOD_DAMAGE_DONE:         true,
	dbcenums.A_MOD_TARGET_RESISTANCE:   true,
	dbcenums.A_MOD_RESISTANCE:          true,
	dbcenums.A_MOD_ARMOR_NO_MODS:       true,
	dbcenums.A_MOD_INCREASE_ENERGY:     true,
	dbcenums.A_MOD_MAX_POWER:           true,
	dbcenums.A_MOD_INCREASE_HEALTH_2:   true,
	dbcenums.A_MOD_BLOCK_VALUE_FLAT:    true,
	dbcenums.A_PERIODIC_TRIGGER_SPELL:  true,
}

// Whether either implicit target is an enemy. What such an aura changes is the enemy's, never a stat
// of the wearer.
func (effect *SpellEffect) HitsAnEnemy() bool {
	return slices.ContainsFunc(effect.ImplicitTargets, func(target int) bool {
		return spelldata.TargetsAnEnemy(dbcenums.ImplicitTarget(target))
	})
}

// Reports whether the effect's aura is one of the types that can carry stats.
func (effect *SpellEffect) GrantsStats() bool {
	return statAuraTypes[effect.EffectAura]
}

func (effect *SpellEffect) IsDirectDamageEffect() bool {
	types := []SpellEffectType{
		dbcenums.E_HEAL, dbcenums.E_SCHOOL_DAMAGE, dbcenums.E_HEALTH_LEECH,
		dbcenums.E_NORMALIZED_WEAPON_DMG, dbcenums.E_WEAPON_DAMAGE, dbcenums.E_WEAPON_PERCENT_DAMAGE,
	}
	return slices.Contains(types, effect.EffectType)
}

func (effect *SpellEffect) IsPeriodicDamageEffect() bool {
	subtypes := []EffectAuraType{
		dbcenums.A_PERIODIC_DAMAGE, dbcenums.A_PERIODIC_LEECH, dbcenums.A_PERIODIC_HEAL,
	}
	if effect.EffectType == dbcenums.E_APPLY_AURA {
		return slices.Contains(subtypes, effect.EffectAura)
	}
	return false
}

func (data *SpellEffect) ClassFlag(index uint) uint32 {
	return uint32(data.EffectSpellClassMasks[index/32]) & (1 << (index % 32))
}
func (effect *SpellEffect) CalcCoefficientStatValue(ilvl int) float64 {
	propPoints := effect.GetScalingValue(ilvl)
	return math.Round(float64(propPoints) * effect.Coefficient)
}
func (effect *SpellEffect) GetScalingValue(ilvl int) float64 {
	if ilvl > 0 {
		// If item we get rand prop points
		return float64(dbcInstance.RandomPropertiesByIlvl[ilvl][proto.ItemQuality_ItemQualityEpic][0])
	}
	spell := dbcInstance.Spells[effect.SpellID]
	if spell.ScalesFromItemLevel > 0 {
		// If item scales from a fixed ilvl we get rand prop points
		return float64(dbcInstance.RandomPropertiesByIlvl[int(spell.ScalesFromItemLevel)][proto.ItemQuality_ItemQualityEpic][0])
	}

	// if not we get class scaling based on the spell
	scale := effect.ScalingClass()
	return dbcInstance.SpellScalings[min(spell.MaxScalingLevel, BASE_LEVEL)].Values[scale]
}

// The stats the effect grants, and whether it granted any. The caller needs the two apart: an
// effect that resolves to nothing is the signal to keep walking down the trigger chain, and "all
// zeroes" is a legitimate resolution for an aura this does not model.
func (effect *SpellEffect) ParseStatEffect(scalesWithIlvl bool, ilvl int) (stats.Stats, bool) {
	effectStats := stats.Stats{}
	stat, _ := MapMainStatToStat(effect.EffectMiscValues[0])

	switch {
	case effect.EffectAura == dbcenums.A_MOD_RANGED_ATTACK_POWER:
		if effect.Coefficient != 0 && scalesWithIlvl {
			effectStats[proto.Stat_StatRangedAttackPower] = effect.CalcCoefficientStatValue(ilvl)
			break
		}
		effectStats[proto.Stat_StatRangedAttackPower] = float64(effect.EffectBasePoints + effect.EffectDieSides)
	case effect.EffectAura == dbcenums.A_MOD_ATTACK_POWER:
		if effect.Coefficient != 0 && scalesWithIlvl {
			effectStats[proto.Stat_StatAttackPower] = effect.CalcCoefficientStatValue(ilvl)
			break
		}
		effectStats[proto.Stat_StatAttackPower] = float64(effect.EffectBasePoints + effect.EffectDieSides)
	case effect.EffectMiscValues[0] == -1 && effect.EffectAura == dbcenums.A_MOD_STAT && effect.EffectType == dbcenums.E_APPLY_AURA:
		// -1 represents ALL STATS if present in MiscValue 0
		for _, s := range []proto.Stat{
			proto.Stat_StatAgility, proto.Stat_StatIntellect, proto.Stat_StatSpirit,
			proto.Stat_StatStamina, proto.Stat_StatStrength,
		} {
			if effect.Coefficient != 0 && effect.ScalingType != 0 {
				effectStats[s] = effect.CalcCoefficientStatValue(core.TernaryInt(scalesWithIlvl, ilvl, 0))
				continue
			}
			effectStats[s] = float64(effect.EffectBasePoints + effect.EffectDieSides)
		}
	case effect.EffectAura == dbcenums.A_MOD_STAT && effect.EffectType == dbcenums.E_APPLY_AURA:
		if effect.Coefficient != 0 && effect.ScalingType != 0 {
			effectStats[stat] = effect.CalcCoefficientStatValue(core.TernaryInt(scalesWithIlvl, ilvl, 0))
			break
		}

		// if Coefficient is not set, we fall back to EffectBasePoints
		effectStats[stat] = float64(effect.EffectBasePoints + effect.EffectDieSides)
	case effect.EffectAura == dbcenums.A_MOD_DAMAGE_DONE && effect.EffectType == dbcenums.E_APPLY_AURA:
		school := SpellSchool(effect.EffectMiscValues[0])
		if school == ALL_SPELL_DAMAGE {
			stat := ConvertEffectAuraToStatIndex(effect.EffectAura, effect.EffectMiscValues[0])
			if effect.Coefficient != 0 && effect.ScalingType != 0 {
				effectStats[stat] = effect.CalcCoefficientStatValue(core.TernaryInt(scalesWithIlvl, ilvl, 0))
				break
			}
			// Apply spell power, A_MOD_HEALING_DONE is also a possibility for healing power
			effectStats[stat] += float64(effect.EffectBasePoints + effect.EffectDieSides)
		} else {
			for schoolType, stat := range SpellSchoolToSpellDamageStat {
				if school.Has(schoolType) && stat > -1 {
					if effect.Coefficient != 0 && effect.ScalingType != 0 {
						effectStats[stat] = effect.CalcCoefficientStatValue(core.TernaryInt(scalesWithIlvl, ilvl, 0))
						continue
					}
					// Apply spell power, A_MOD_HEALING_DONE is also a possibility for healing power
					effectStats[stat] += float64(effect.EffectBasePoints + effect.EffectDieSides)
				}
			}
		}
	case effect.EffectAura == dbcenums.A_MOD_HEALING_DONE && effect.EffectType == dbcenums.E_APPLY_AURA:
		if effect.Coefficient != 0 && effect.ScalingType != 0 {
			effectStats[proto.Stat_StatHealingPower] = effect.CalcCoefficientStatValue(core.TernaryInt(scalesWithIlvl, ilvl, 0))
			break
		}
		// Apply spell power, A_MOD_HEALING_DONE is also a possibility for healing power
		effectStats[proto.Stat_StatHealingPower] = float64(effect.EffectBasePoints + effect.EffectDieSides)
	case effect.EffectAura == dbcenums.A_MOD_ARMOR_NO_MODS:
		for _, stat := range spelldata.ArmorNoModsStats(int32(effect.EffectMiscValues[0])) {
			if effect.Coefficient != 0 && scalesWithIlvl {
				effectStats[stat] = effect.CalcCoefficientStatValue(ilvl)
				continue
			}
			effectStats[stat] += float64(effect.EffectBasePoints + effect.EffectDieSides)
		}
	case effect.EffectAura == dbcenums.A_MOD_RESISTANCE:
		school := SpellSchool(effect.EffectMiscValues[0])
		for schoolType, stat := range SpellSchoolToResistanceStat {
			if school.Has(schoolType) && stat > -1 {
				if effect.Coefficient != 0 && scalesWithIlvl {
					effectStats[stat] = effect.CalcCoefficientStatValue(ilvl)
					continue
				}
				// Armor/Spell Pen is negative in the DB, so Subtract it
				effectStats[stat] += float64(effect.EffectBasePoints + effect.EffectDieSides)
			}
		}

	case effect.EffectAura == dbcenums.A_MOD_RATING:
		scaled := effect.Coefficient != 0 && scalesWithIlvl
		var scaledValue float64
		if scaled {
			scaledValue = effect.CalcCoefficientStatValue(ilvl)
		}

		for _, statMod := range spelldata.RatingStats(int32(effect.EffectMiscValues[0])) {
			if scaled {
				effectStats[statMod] = scaledValue
			} else {
				effectStats[statMod] = float64(effect.EffectBasePoints + effect.EffectDieSides)
			}
		}
	case effect.EffectAura == dbcenums.A_MOD_INCREASE_ENERGY,
		effect.EffectAura == dbcenums.A_MOD_MAX_POWER,
		effect.EffectAura == dbcenums.A_MOD_INCREASE_HEALTH,
		effect.EffectAura == dbcenums.A_MOD_INCREASE_HEALTH_2,
		effect.EffectAura == dbcenums.A_MOD_SHIELD_BLOCKVALUE:
		if stat := ConvertEffectAuraToStatIndex(effect.EffectAura, effect.EffectMiscValues[0]); stat >= 0 {
			effectStats[stat] = float64(effect.EffectBasePoints + effect.EffectDieSides)
		}
	// Written here rather than through ConvertEffectAuraToStatIndex, whose stat an equip spell's 274
	// would add on top of the one AddEquipSpellStats adds for it.
	case effect.EffectAura == dbcenums.A_MOD_BLOCK_VALUE_FLAT:
		effectStats[proto.Stat_StatBlockValue] = float64(effect.EffectBasePoints + effect.EffectDieSides)
	case effect.EffectAura == dbcenums.A_PERIODIC_TRIGGER_SPELL && effect.EffectAuraPeriod == 10000:
		for _, sub := range dbcInstance.SpellEffectsInOrder(effect.EffectTriggerSpell) {
			if subStats, ok := sub.ParseStatEffect(false, 0); ok {
				effectStats.AddInplace(&subStats)
			}
		}
	case effect.EffectAura == dbcenums.A_MOD_TARGET_RESISTANCE:
		if resist := ConvertTargetResistanceFlagToPenetrationStat(effect.EffectMiscValues[0]); resist >= 0 {
			effectStats[resist] = math.Abs(float64(effect.EffectBasePoints + effect.EffectDieSides))
		}
	case effect.EffectAura == dbcenums.A_MOD_POWER_REGEN && effect.EffectType == dbcenums.E_APPLY_AURA:
		if effect.Coefficient != 0 && effect.ScalingType != 0 {
			effectStats[proto.Stat_StatMP5] = effect.CalcCoefficientStatValue(core.TernaryInt(scalesWithIlvl, ilvl, 0))
			break
		}
		effectStats[proto.Stat_StatMP5] = float64(effect.EffectBasePoints + effect.EffectDieSides)
	}

	return effectStats, effectStats != stats.Stats{}
}

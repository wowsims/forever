package dbc

import (
	"compress/gzip"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func GetProfession(id int) proto.Profession {
	if profession, ok := MapProfessionIdToProfession[id]; ok {
		return profession
	}
	return 0
}

func GetRepLevel(minReputation int) proto.RepLevel {
	if repLevel, ok := MapMinReputationToRepLevel[minReputation]; ok {
		return repLevel
	}
	return proto.RepLevel_RepLevelUnknown
}
func NullFloat(arr []float64) []float64 {
	for _, v := range arr {
		if v > 0 {
			return arr
		}
	}

	return nil
}

// The playable races by ChrRaces.PlayableRaceBit, their bit in ItemSparse.AllowableRace: ChrRaces.ID - 1
// for the classic races, 32 and 33 for High Order (ChrRaces 95) and Windshaper (96) Skyborne.
var playableRaces = []struct {
	race proto.Race
	bit  int
}{
	{proto.Race_RaceHuman, 0}, {proto.Race_RaceOrc, 1}, {proto.Race_RaceDwarf, 2}, {proto.Race_RaceNightElf, 3},
	{proto.Race_RaceUndead, 4}, {proto.Race_RaceTauren, 5}, {proto.Race_RaceGnome, 6}, {proto.Race_RaceTroll, 7},
	{proto.Race_RaceBloodElf, 9}, {proto.Race_RaceHighOrderSkyborne, 32}, {proto.Race_RaceWindshaperSkyborne, 33},
}

// The playable races an AllowableRace mask names, or none where it names every one of them.
func GetRacesFromRaceMask(mask int64) []proto.Race {
	var result []proto.Race
	for _, r := range playableRaces {
		if mask&(1<<r.bit) != 0 {
			result = append(result, r.race)
		}
	}
	if len(result) == len(playableRaces) {
		return nil
	}
	slices.Sort(result)
	return result
}

func GetClassesFromClassMask(mask int) []proto.Class {
	var result []proto.Class

	allClasses := (1 << len(Classes)) - 1
	if mask&allClasses == allClasses {
		return result
	}

	for _, class := range Classes {
		if mask&(1<<(class.ID-1)) != 0 {
			result = append(result, class.ProtoClass)
		}
	}
	slices.Sort(result)
	return result
}

func WriteGzipFile(filePath string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directories for %s: %w", filePath, err)
	}
	// Create the file
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Create a gzip writer on top of the file writer
	gw := gzip.NewWriter(f)
	defer gw.Close()

	// Write the data to the gzip writer
	_, err = gw.Write(data)
	return err
}
func ReadGzipFile(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, DataLoadError{
			Source:   filename,
			DataType: "gzip file",
			Reason:   err.Error(),
		}
	}
	defer f.Close()

	gzReader, err := gzip.NewReader(f)
	if err != nil {
		return nil, DataLoadError{
			Source:   filename,
			DataType: "gzip",
			Reason:   err.Error(),
		}
	}
	defer gzReader.Close()

	data, err := io.ReadAll(gzReader)
	if err != nil {
		return nil, DataLoadError{
			Source:   filename,
			DataType: "decompression",
			Reason:   err.Error(),
		}
	}

	return data, nil
}

func processEnchantmentEffects(
	effects []int,
	effectArgs []int,
	effectPoints []int,
	outStats *stats.Stats,
	outPseudoStats []float64,
	addRanged bool,
) {
	for i, effect := range effects {
		switch effect {
		case ITEM_ENCHANTMENT_RESISTANCE:
			stat, match := MapResistanceToStat(effectArgs[i])
			if !match {
				continue
			}
			outStats[stat] = float64(effectPoints[i])
		case ITEM_ENCHANTMENT_STAT:
			mapped, success := MapBonusStatIndexToStats(effectArgs[i])
			if !success {
				continue
			}
			for _, stat := range mapped {
				outStats[stat] = float64(effectPoints[i])

				// If the bonus stat is attack power, copy it to ranged attack power
				if addRanged && stat == proto.Stat_StatAttackPower {
					outStats[proto.Stat_StatRangedAttackPower] = float64(effectPoints[i])
				}
			}
		case ITEM_ENCHANTMENT_EQUIP_SPELL: //Buff
			if !EquipSpellFolds(effectArgs[i]) {
				continue
			}
			AddEquipSpellStats(outStats, outPseudoStats, effectArgs[i])
			for _, spellEffect := range dbcInstance.SpellEffects[effectArgs[i]] {
				addEquipEffectStat(outStats, spellEffect, spellEffect.EffectBasePoints+spellEffect.EffectDieSides)
			}
		case ITEM_ENCHANTMENT_COMBAT_SPELL:
			// Not processed (chance on hit, ignore for now)
		case ITEM_ENCHANTMENT_USE_SPELL:
			// Not processed
		}
	}
}

// Adds the stat one effect of an equip spell gives to out, value points of it: a main stat aura's
// stat, or all five main stats for a misc value of -1, and otherwise the stat addEquipAuraStat adds.
func addEquipEffectStat(out *stats.Stats, effect SpellEffect, value float64) {
	if effect.EffectType == dbcenums.E_APPLY_AURA && effect.EffectAura == dbcenums.A_MOD_STAT {
		if effect.EffectMiscValues[0] == -1 {
			out[proto.Stat_StatAgility] += value
			out[proto.Stat_StatIntellect] += value
			out[proto.Stat_StatSpirit] += value
			out[proto.Stat_StatStamina] += value
			out[proto.Stat_StatStrength] += value
		} else if stat, ok := MapMainStatToStat(effect.EffectMiscValues[0]); ok {
			out[stat] += value
		}
		return
	}
	addEquipAuraStat(out, effect, value)
}

const SkillLineDefense = 95

// Where one effect of an equip spell goes in the data of the item or enchant carrying it.
type EquipFold uint8

const (
	// The effect's stat, on every channel it names.
	FoldStats EquipFold = iota
	// Ranged attacks alone: a hit or crit aura bound to ranged weapons alone, or a rating mask of
	// ranged bits alone, as the ranged percents.
	FoldRanged
	// Applied at run time, where the attack's own weapon decides: any other hit or crit aura the parse
	// binds to a weapon, a rating aura bound to any weapon, whose mask the fold would read as a rating
	// on every channel, and a damage done aura bound to any weapon.
	FoldRuntime
)

// The fold of an effect, from the rule spelldata.WeaponScope gives it.
func equipFold(s *spelldata.Spell, e *spelldata.Effect) EquipFold {
	switch rule, scope := spelldata.WeaponScope(s, e); {
	case rule == spelldata.WeaponRuleNone:
		return FoldStats
	case rule == spelldata.WeaponRuleRangedRatings:
		return FoldRanged
	case e.Aura == dbcenums.A_MOD_RATING, e.Aura == dbcenums.A_MOD_DAMAGE_PERCENT_DONE, scope.ProcMask != core.ProcMaskRanged:
		return FoldRuntime
	}
	return FoldRanged
}

// The fold of a game files effect, read off its row in the spell store. An effect of a spell the store
// does not carry names no weapon.
func EquipEffectFold(effect SpellEffect) EquipFold {
	s := spelldata.Find(int32(effect.SpellID))
	if e := s.EffectAtIndex(uint8(effect.EffectIndex)); e != spelldata.NilEffect {
		return equipFold(s, e)
	}
	return equipFold(s, &spelldata.Effect{Aura: effect.EffectAura, Misc: int32(effect.EffectMiscValues[0])})
}

// Whether an equip spell folds into the stats of the item or enchant carrying it: it does unless an
// effect of it is FoldRuntime. Every such spell on an item or enchant is a single hit, crit or rating
// aura, so the whole spell goes one way.
func EquipSpellFolds(spellID int) bool {
	s := spelldata.Find(int32(spellID))
	for i := range s.Effects {
		if equipFold(s, &s.Effects[i]) == FoldRuntime {
			return false
		}
	}
	return true
}

// Adds the stat ConvertEffectAuraToStatIndex names for an equip spell's aura to out, value points of
// it, and reports whether it named one. A FoldStats effect alone adds. A resistance to every magic
// school is each of the five resistances, a spell whose description says "forms only" gives feral
// attack power, and armor penetration and spell piercing count as positive.
func addEquipAuraStat(out *stats.Stats, effect SpellEffect, value float64) bool {
	if EquipEffectFold(effect) != FoldStats {
		return false
	}
	stat := ConvertEffectAuraToStatIndex(effect.EffectAura, effect.EffectMiscValues[0])
	if stat < 0 && stat != -2 {
		return false
	}
	if strings.Contains(dbcInstance.Spells[effect.SpellID].Description, "forms only") {
		stat = proto.Stat_StatFeralAttackPower
	}
	switch stat {
	case -2:
		out[proto.Stat_StatArcaneResistance] += value
		out[proto.Stat_StatFireResistance] += value
		out[proto.Stat_StatFrostResistance] += value
		out[proto.Stat_StatNatureResistance] += value
		out[proto.Stat_StatShadowResistance] += value
	case proto.Stat_StatArmorPenetration, proto.Stat_StatSpellPiercing:
		out[stat] += math.Abs(value)
	default:
		out[stat] += value
	}
	return true
}

// The ranged offset percents a combat rating mask of ranged bits alone grants, from rating points.
func RangedRatingPercents(mask int, rating float64) map[proto.PseudoStat]float64 {
	percents := map[proto.PseudoStat]float64{}
	for _, scoped := range spelldata.ScopedRatingPercents(int32(mask), rating) {
		if channels, ok := core.RatingChannelsOf(scoped.Stat); ok {
			percents[channels.Ranged] = scoped.Percent
		}
	}
	return percents
}

// Adds what an equip spell states that no stat index carries, and reports whether it added to each of
// the two: to outStats the defense skill and flat block value - 24148 (Presence of Might) states
// A_MOD_SKILL 7 on the Defense skill line and A_MOD_BLOCK_VALUE_FLAT 15 - and to pseudoStats, indexed by
// proto.PseudoStat and skipped where nil, the percents for hit, crit, block, dodge, parry and attack and
// cast speed. Only auras on the wearer count: Atiesh's 28142 is a party aura and reaches the sim as a
// raid buff. A spell EquipSpellFolds leaves to run time adds nothing. A FoldRanged effect adds to the
// ranged percents alone, as the total the character sheet shows.
func AddEquipSpellStats(outStats *stats.Stats, pseudoStats []float64, spellID int) (addedStats bool, addedPseudoStats bool) {
	if !EquipSpellFolds(spellID) {
		return false, false
	}
	add := func(pseudoStat proto.PseudoStat, value float64) {
		pseudoStats[pseudoStat] += value
		addedPseudoStats = true
	}

	for _, effect := range dbcInstance.SpellEffectsInOrder(spellID) {
		if effect.EffectType != dbcenums.E_APPLY_AURA {
			continue
		}
		value := effect.EffectBasePoints + effect.EffectDieSides

		switch {
		case effect.EffectAura == dbcenums.A_MOD_SKILL && effect.EffectMiscValues[0] == SkillLineDefense:
			outStats[proto.Stat_StatDefenseRating] += value * core.DefenseRatingPerDefenseLevel
			addedStats = true
			continue
		case effect.EffectAura == dbcenums.A_MOD_BLOCK_VALUE_FLAT:
			outStats[proto.Stat_StatBlockValue] += value
			addedStats = true
			continue
		}

		// 1293881 (Pendulum of Doom) states a zero melee haste and changes the speed through its procs.
		if pseudoStats == nil || value == 0 {
			continue
		}

		ranged := EquipEffectFold(effect) == FoldRanged
		addWeapon := func(meleeStat, rangedStat proto.PseudoStat, value float64) {
			if !ranged {
				add(meleeStat, value)
			}
			add(rangedStat, value)
		}

		switch effect.EffectAura {
		case dbcenums.A_MOD_HIT_CHANCE:
			addWeapon(proto.PseudoStat_PseudoStatMeleeHitPercent, proto.PseudoStat_PseudoStatRangedHitPercent, value)
		case dbcenums.A_MOD_WEAPON_CRIT_PERCENT:
			addWeapon(proto.PseudoStat_PseudoStatMeleeCritPercent, proto.PseudoStat_PseudoStatRangedCritPercent, value)
		case dbcenums.A_MOD_CRIT_PCT:
			addWeapon(proto.PseudoStat_PseudoStatMeleeCritPercent, proto.PseudoStat_PseudoStatRangedCritPercent, value)
			if !ranged {
				add(proto.PseudoStat_PseudoStatSpellCritPercent, value)
			}
		case dbcenums.A_MOD_SPELL_HIT_CHANCE:
			add(proto.PseudoStat_PseudoStatSpellHitPercent, value)
		case dbcenums.A_MOD_SPELL_CRIT_CHANCE:
			add(proto.PseudoStat_PseudoStatSpellCritPercent, value)
		case dbcenums.A_MOD_BLOCK_PERCENT:
			add(proto.PseudoStat_PseudoStatBlockPercent, value)
		case dbcenums.A_MOD_DODGE_PERCENT:
			add(proto.PseudoStat_PseudoStatDodgePercent, value)
		case dbcenums.A_MOD_PARRY_PERCENT:
			add(proto.PseudoStat_PseudoStatParryPercent, value)
		case dbcenums.A_MOD_RATING:
			if !ranged {
				break
			}
			for pseudoStat, percent := range RangedRatingPercents(effect.EffectMiscValues[0], value) {
				add(pseudoStat, percent)
			}
		}

		for _, pseudoStat := range spelldata.SpeedAuraPseudoStats[effect.EffectAura] {
			add(pseudoStat, value)
		}
	}

	return addedStats, addedPseudoStats
}

func ConvertEffectAuraToStatIndex(effectAura EffectAuraType, effectMisc int) proto.Stat {
	switch effectAura {
	case dbcenums.A_MOD_ATTACK_POWER:
		return proto.Stat_StatAttackPower
	case dbcenums.A_MOD_RANGED_ATTACK_POWER:
		return proto.Stat_StatRangedAttackPower
	case dbcenums.A_MOD_DAMAGE_DONE:
		return ConvertSpellDamageFlagToSchoolDamageStat(effectMisc)
	case dbcenums.A_MOD_HEALING_DONE:
		return proto.Stat_StatHealingPower
	case dbcenums.A_MOD_INCREASE_HEALTH, dbcenums.A_MOD_INCREASE_HEALTH_2:
		return proto.Stat_StatHealth
	case dbcenums.A_MOD_INCREASE_ENERGY, dbcenums.A_MOD_MAX_POWER:
		// MiscValue 0 is the power type. Only mana has a matching stat; rage, focus, energy
		// and the rest are resources the sim tracks per spec, not stats.
		if effectMisc == POWER_TYPE_MANA {
			return proto.Stat_StatMana
		}
		return -1
	case dbcenums.A_MOD_TARGET_RESISTANCE:
		return ConvertTargetResistanceFlagToPenetrationStat(effectMisc)
	case dbcenums.A_MOD_RESISTANCE:
		return ConvertResistanceFlagToResistanceStat(effectMisc)
	case dbcenums.A_MOD_ARMOR_NO_MODS:
		if sts := spelldata.ArmorNoModsStats(int32(effectMisc)); len(sts) == 1 {
			return proto.Stat(sts[0])
		}
		return -1
	case dbcenums.A_MOD_RATING:
		return ConvertModRatingFlagToRatingStat(effectMisc)
	case dbcenums.A_MOD_SHIELD_BLOCKVALUE:
		return proto.Stat_StatBlockValue
	case dbcenums.A_MOD_POWER_REGEN:
		return proto.Stat_StatMP5
	default:
		return -1
	}
}

func ConvertResistanceFlagToResistanceStat(flag int) proto.Stat {
	school := SpellSchool(flag)
	if school == ALL_SPELL_DAMAGE || school == SPELL_PENETRATION {
		// All 5 Magic School resist; return -2 to be handled as special case
		// 124 excludes "Holy" which isn't a resist anyways
		return -2
	}
	for schoolType, stat := range SpellSchoolToResistanceStat {
		if school.Has(schoolType) {
			return stat
		}
	}
	return -1
}

// See core.SpellSchool.PiercingStat; a mask without a stat returns -1.
func ConvertTargetResistanceFlagToPenetrationStat(flag int) proto.Stat {
	if stat, ok := core.SpellSchool(flag).PiercingStat(); ok {
		return proto.Stat(stat)
	}
	return -1
}

func ConvertSpellDamageFlagToSchoolDamageStat(flag int) proto.Stat {
	switch flag {
	case 1:
		return proto.Stat_StatPhysicalDamage
	case 2:
		return proto.Stat_StatHolyDamage
	case 4:
		return proto.Stat_StatFireDamage
	case 8:
		return proto.Stat_StatNatureDamage
	case 16:
		return proto.Stat_StatFrostDamage
	case 32:
		return proto.Stat_StatShadowDamage
	case 64:
		return proto.Stat_StatArcaneDamage
	default:
		return proto.Stat_StatSpellDamage
	}
}

// The rating stat a combat rating mask names, read bit by bit through spelldata.RatingStats, so a
// mask of only ranged bits names none. The callers read one stat per aura, so a mask naming two
// different stats names none either; no game files aura carries one.
func ConvertModRatingFlagToRatingStat(flag int) proto.Stat {
	ratings := spelldata.RatingStats(int32(flag))
	if len(ratings) == 1 {
		return proto.Stat(ratings[0])
	}
	if flag != 0 && !spelldata.RangedBitsOnly(int32(flag)) {
		println("UNHANDLED RATING FLAG: " + strconv.Itoa(flag))
	}
	return -1
}

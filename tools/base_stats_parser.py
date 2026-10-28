#!/usr/bin/python

import csv
import os
import re
import sqlite3

# Generates sim/core/base_stats_auto_gen.go from the data extracted from the game files.
#
# Two sources, both produced by `make db` from the Forever beta game files:
#
#   tools/database/wowsims.db     PlayerExpectedStat -- base mana and the
#                                 crit-per-agility / crit-per-intellect slopes;
#                                 ExpectedStat -- the creature health, armor
#                                 and auto attack dps of the level
#   assets/db_inputs/basestats/   the CombatRatings gametable, plus the two
#                                 base-crit tables described below
#
# The per-level crit slope tables (chancetomeleecrit / chancetospellcrit) and
# octbasempbyclass used to supply the first three quantities, but the beta
# game files do not ship those gametables -- PlayerExpectedStat replaced them,
# and it disagrees: Forever retuned crit-per-agility for seven of the nine
# classes (hunter 0.0189 where the old tables said 0.0301).

BASE_DIR = ""

DIR_PATH = "assets/db_inputs/basestats/"
OUTPUT_PATH = "sim/core/"
TS_OUTPUT_PATH = "ui/sim/constants/mechanics.ts"
CONSTANTS_PATH = "sim/core/constants.go"
DB_PATH = "tools/database/wowsims.db"

# Neither the Forever nor the Era game files ship the two base-crit gametables,
# so these stay checked in by hand. They hold the vanilla per-class crit a
# character has before any agility/intellect contribution; warrior, hunter and
# rogue have no base melee crit. The Death Knight and Monk columns are unused.
MELEE_CRIT_BASE = "chancetomeleecritbase.txt"
SPELL_CRIT_BASE = "chancetospellcritbase.txt"
COMBAT_RATINGS = "combatratings.txt"

# Read out of sim/core/constants.go rather than repeated here, so there is one place
# that says what level this sim runs at.
def GoConstant(name: str):
    text = open(BASE_DIR + CONSTANTS_PATH).read()
    return float(re.search(rf"^const {name} = ([\d.]+)$", text, re.M).group(1))


BASE_LEVEL = int(GoConstant("CharacterLevel"))

# The boss is three levels above the player; see defaultRaidBossLevel in sim/core/target.go.
BOSS_LEVEL_OFFSET = 3

# ChrClasses.ID -> the proto.Class suffix. Hardcoded rather than read from the
# game files so the generated identifiers cannot silently follow a renumbering.
CLASS_IDS = {
    1: "Warrior",
    2: "Paladin",
    3: "Hunter",
    4: "Rogue",
    5: "Priest",
    7: "Shaman",
    8: "Mage",
    9: "Warlock",
    11: "Druid",
}

CLASSES = ["Warrior", "Paladin", "Hunter", "Rogue", "Priest", "Shaman", "Mage", "Warlock", "Druid"]


# Returns {row key: {column name: value}}. Indexing by header name rather than
# by position survives the column order changes the game files make between
# expansions -- the beta's SpellScaling grew Evoker/Adventurer/Traveler columns.
def GenNameIndexedDb(file: str):
    db = {}
    with open(file) as tsv:
        rows = list(csv.reader(tsv, delimiter="\t"))
    header = rows[0]
    for row in rows[1:]:
        db[row[0]] = dict(zip(header[1:], row[1:]))
    return db


def GenExpectedStats(dbPath: str, level: int):
    conn = sqlite3.connect(dbPath)
    rows = conn.execute(
        """SELECT ClassID, BaseMana, CritPerAgility, SpellCritPerIntellect
           FROM PlayerExpectedStat WHERE Level = ? AND ContentSetID = 0""",
        (level,),
    ).fetchall()
    conn.close()

    stats = {}
    for classID, baseMana, critPerAgi, critPerInt in rows:
        if classID not in CLASS_IDS:
            continue
        stats[CLASS_IDS[classID]] = {
            "Mana": baseMana,
            # The game files store these as a fraction per point (0.0005 = one
            # percent per 20 agility); the sim's stat dependencies are in
            # percent per point.
            "CritPerAgi": critPerAgi * 100,
            "CritPerInt": critPerInt * 100,
        }

    missing = [c for c in CLASSES if c not in stats]
    if missing:
        raise SystemExit(f"PlayerExpectedStat has no level {level} row for: {', '.join(missing)}")
    return stats


def GenCreatureExpectedStats(dbPath: str, level: int):
    conn = sqlite3.connect(dbPath)
    rows = conn.execute(
        """SELECT CreatureHealth, CreatureArmor, CreatureAutoAttackDps
           FROM ExpectedStat WHERE Lvl = ? AND ContentSetID = 0""",
        (level,),
    ).fetchall()
    conn.close()

    if len(rows) != 1:
        raise SystemExit(f"ExpectedStat has {len(rows)} level {level} rows, expected one")
    health, armor, autoAttackDps = rows[0]
    return {"Health": float(health), "Armor": float(armor), "AutoAttackDps": float(autoAttackDps)}


def GenCreatureArmorByLevel(dbPath: str):
    conn = sqlite3.connect(dbPath)
    rows = conn.execute("SELECT Lvl, CreatureArmor FROM ExpectedStat WHERE ContentSetID = 0 ORDER BY Lvl").fetchall()
    conn.close()

    if not rows:
        raise SystemExit("ExpectedStat has no creature rows")
    return {int(level): float(armor) for level, armor in rows}


class ClassStats:
    Expected: dict
    Creature: dict
    CreatureArmorByLevel: dict
    MCritBase: dict
    SCritBase: dict
    CombatRatings: dict


# The sim has one hit, crit and haste rating that feeds melee, ranged and spells alike, so the
# gametable's three columns for each must convert at the same rate.
def SharedRating(ratings, kind: str) -> float:
    columns = {f"{kind} - {channel}": float(ratings[f"{kind} - {channel}"]) for channel in ("Melee", "Ranged", "Spell")}
    if len(set(columns.values())) != 1:
        raise SystemExit(f"{kind} rating converts differently per channel at level {BASE_LEVEL}: {columns}")
    return next(iter(columns.values()))


def GenExtraStatsGoFile(cs: ClassStats):
    header = '''
package core

// **************************************
// AUTO GENERATED BY BASE_STATS_PARSER.PY
// **************************************

import (
    "github.com/wowsims/forever/sim/core/proto"
    "github.com/wowsims/forever/sim/core/stats"
)

'''
    output = header
    ratings = cs.CombatRatings[str(BASE_LEVEL)]

    # Rating per whole percent, like the hit and crit constants below. ExpertiseRating feeds
    # ExpertisePercent through a stat dependency; there is no quarter-point rounding in Forever.
    output += f"const ExpertiseRatingPerExpertisePercent = {float(ratings['Expertise']):f}\n"
    output += f"const DefenseRatingPerDefenseLevel = {float(ratings['Defense Skill']):f}\n"
    output += f"const DodgeRatingPerDodgePercent = {float(ratings['Dodge']):f}\n"
    output += f"const ParryRatingPerParryPercent = {float(ratings['Parry']):f}\n"
    output += f"const BlockRatingPerBlockPercent = {float(ratings['Block']):f}\n"
    output += f"const HitRatingPerHitPercent = {SharedRating(ratings, 'Hit'):f}\n"
    output += f"const CritRatingPerCritPercent = {SharedRating(ratings, 'Crit'):f}\n"
    output += f"const HasteRatingPerHastePercent = {SharedRating(ratings, 'Haste'):f}\n"

    output += "// ExpectedStat's creature values at the character level\n"
    output += f"const CreatureHealthMaxLevel = {cs.Creature['Health']!r}\n"
    output += f"const CreatureArmorMaxLevel = {cs.Creature['Armor']!r}\n"
    output += f"const CreatureAutoAttackDpsMaxLevel = {cs.Creature['AutoAttackDps']!r}\n"

    output += "// ExpectedStat's creature armor at each level\n"
    output += "var CreatureArmorByLevel = map[int32]float64{"
    for level, armor in cs.CreatureArmorByLevel.items():
        output += f"\n{level}: {armor!r},"
    output += "\n}\n"

    output += '''var CritPerAgiMaxLevel = map[proto.Class]float64{
proto.Class_ClassUnknown: 0.0,'''
    for c in CLASSES:
        output += f"\nproto.Class_Class{c}: {cs.Expected[c]['CritPerAgi']:.8f},"
    output += "\n}\n"

    output += '''var CritPerIntMaxLevel = map[proto.Class]float64{
proto.Class_ClassUnknown: 0.0,'''
    for c in CLASSES:
        output += f"\nproto.Class_Class{c}: {cs.Expected[c]['CritPerInt']:.8f},"
    output += "\n}\n"

    output += '''var ExtraClassBaseStats = map[proto.Class]stats.Stats{
proto.Class_ClassUnknown: {},'''
    for c in CLASSES:
        output += f"\nproto.Class_Class{c}: {{"
        output += f"\n stats.Mana: {float(cs.Expected[c]['Mana']):.4f},"
        output += f"\n stats.SpellCritPercent: {float(cs.SCritBase['1'][c]) * 100:.4f},"
        output += f"\n stats.PhysicalCritPercent: {float(cs.MCritBase['1'][c]) * 100:.4f},"
        output += "\n},"
    output += "\n}\n"
    return output


# ui/sim/constants/mechanics.ts used to be a hand-kept copy of these numbers, and it drifted:
# it was still converting spell hit at 8 and defense at 1.5 after the game files moved both. It is
# generated from the same values as the Go file now, so the two cannot disagree.
def GenMechanicsTsFile(cs: ClassStats):
    ratings = cs.CombatRatings[str(BASE_LEVEL)]

    def num(value):
        return f"{float(value):g}"

    lines = [
        "// **************************************",
        "// AUTO GENERATED BY BASE_STATS_PARSER.PY",
        "//",
        "// Do not edit. These mirror sim/core, and a Go test fails if they drift:",
        "// TestMechanicsConstantsMatchTheUI in sim/core/mechanics_mirror_test.go.",
        "// **************************************",
        "",
        f"export const CHARACTER_LEVEL = {BASE_LEVEL};",
        f"export const BOSS_LEVEL = CHARACTER_LEVEL + {BOSS_LEVEL_OFFSET};",
        "",
        f"export const EXPERTISE_RATING_PER_EXPERTISE_PERCENT = {num(ratings['Expertise'])};",
        f"export const HASTE_RATING_PER_HASTE_PERCENT = {num(SharedRating(ratings, 'Haste'))};",
        f"export const CRIT_RATING_PER_CRIT_PERCENT = {num(SharedRating(ratings, 'Crit'))};",
        f"export const HIT_RATING_PER_HIT_PERCENT = {num(SharedRating(ratings, 'Hit'))};",
        f"export const DODGE_RATING_PER_DODGE_PERCENT = {num(ratings['Dodge'])};",
        f"export const PARRY_RATING_PER_PARRY_PERCENT = {num(ratings['Parry'])};",
        f"export const DEFENSE_RATING_PER_DEFENSE_LEVEL = {num(ratings['Defense Skill'])};",
        f"export const BLOCK_RATING_PER_BLOCK_PERCENT = {num(ratings['Block'])};",
        "",
        f"export const MISS_DODGE_PARRY_BLOCK_CRIT_CHANCE_PER_DEFENSE = {num(GoConstant('MissDodgeParryBlockCritChancePerDefense'))};",
        "",
    ]
    return "\n".join(lines)


if __name__ == "__main__":
    args = ClassStats()
    args.Expected = GenExpectedStats(BASE_DIR + DB_PATH, BASE_LEVEL)
    args.Creature = GenCreatureExpectedStats(BASE_DIR + DB_PATH, BASE_LEVEL)
    args.CreatureArmorByLevel = GenCreatureArmorByLevel(BASE_DIR + DB_PATH)
    args.MCritBase = GenNameIndexedDb(BASE_DIR + DIR_PATH + MELEE_CRIT_BASE)
    args.SCritBase = GenNameIndexedDb(BASE_DIR + DIR_PATH + SPELL_CRIT_BASE)
    args.CombatRatings = GenNameIndexedDb(BASE_DIR + DIR_PATH + COMBAT_RATINGS)

    output = GenExtraStatsGoFile(args)
    fname = BASE_DIR + OUTPUT_PATH + "base_stats_auto_gen.go"
    print(f"Writing stats to: {fname}")
    f = open(fname, "w")
    f.write(output)
    f.close()
    os.system(f"gofmt -w {fname}")

    tsname = BASE_DIR + TS_OUTPUT_PATH
    print(f"Writing mechanics to: {tsname}")
    with open(tsname, "w") as tsfile:
        tsfile.write(GenMechanicsTsFile(args))

package spelldata

import (
	"fmt"
	"math"
	"strings"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// The misc values the stat and school auras key on.
const (
	miscAllSchools  int32 = 127 // every school bit, which the sim states as one multiplier
	miscMagicSchool int32 = 126 // every school but physical, which is what spell damage covers
	miscArmor       int32 = 1   // A_MOD_RESISTANCE, A_MOD_ARMOR_NO_MODS and the two resistance percentages state armor as school 1
	miscAllStats    int32 = -1  // A_MOD_STAT, A_MOD_PERCENT_STAT and A_MOD_TOTAL_STAT_PERCENTAGE: all five at once
	miscDefense     int32 = 95  // A_MOD_SKILL on the Defense skill line
)

// The weapon skill each of the game files' weapon skill lines is. Unarmed (162) is fist weapons: Devilsaur
// Claws' "Increased Fist Weapons" names it.
var weaponSkillLines = map[int32]proto.WeaponSkill{
	43:  proto.WeaponSkill_WeaponSkillSwords,
	44:  proto.WeaponSkill_WeaponSkillAxes,
	45:  proto.WeaponSkill_WeaponSkillBows,
	46:  proto.WeaponSkill_WeaponSkillGuns,
	54:  proto.WeaponSkill_WeaponSkillMaces,
	55:  proto.WeaponSkill_WeaponSkillTwoHandedSwords,
	136: proto.WeaponSkill_WeaponSkillStaves,
	160: proto.WeaponSkill_WeaponSkillTwoHandedMaces,
	162: proto.WeaponSkill_WeaponSkillFistWeapons,
	172: proto.WeaponSkill_WeaponSkillTwoHandedAxes,
	173: proto.WeaponSkill_WeaponSkillDaggers,
	176: proto.WeaponSkill_WeaponSkillThrown,
	226: proto.WeaponSkill_WeaponSkillCrossbows,
	228: proto.WeaponSkill_WeaponSkillWands,
	229: proto.WeaponSkill_WeaponSkillPolearms,
	473: proto.WeaponSkill_WeaponSkillFistWeapons,
}

// The five stats the game files count in their own order, which is what A_MOD_STAT, A_MOD_PERCENT_STAT
// and A_MOD_TOTAL_STAT_PERCENTAGE index by.
var clientStats = [5]stats.Stat{stats.Strength, stats.Agility, stats.Stamina, stats.Intellect, stats.Spirit}

// The state one ParseEffects or ParseStatic call shares with the table.
type parser struct {
	unit  *core.Unit
	spell *Spell

	// The character the unit belongs to, for the rows whose helper is a character's rather than a
	// unit's. An aura on a unit with no character of its own - an enemy's debuff - leaves it nil,
	// and those rows are skipped there.
	character *core.Character

	// ParseStatic: there is no aura whose gain hands a Simulation to the attachment, so the rows
	// that need one to act are skipped instead of attached.
	static bool

	// The caller gates every mod through a closure, so a row that cannot be turned back off is
	// skipped instead of attached.
	conditional bool

	// The row states CumulativeAura, or the caller counts the aura more than once, so a value follows
	// a level other than 1. The rows whose value cannot be scaled are skipped while this is set;
	// IgnoreStacks clears the stacks' share of it.
	stacking bool

	// The aura the attachments follow, which the unit's temporary stats listeners are told about.
	// Nil on the static path.
	aura *core.Aura

	// DryRun: the rows are read for what they would attach, and nothing is registered on the unit.
	dry bool

	// OnHand: the weapon the row sits on.
	hand core.ProcMask

	// The slot FromCaster names, which the rows that reach only the caster's attacks share. Nil
	// without a caster.
	casterSlot *casterSlot

	// ShowAsSchoolHit: the schools the hit mods are shown under, and whether a row stamped them.
	showAsSchoolHit        core.SpellSchool
	showAsSchoolHitStamped bool
}

// One attachment made for one effect: the sim kind it became, the value in the sim's own units, and
// the single operation that turns it on, off or up. The level is 0 for off, 1 for the plain value and
// N for an aura sitting at N stacks.
type attachment struct {
	kind  string
	value float64
	set   func(sim *core.Simulation, level float64)

	// The operation reads the Simulation it is handed, so it cannot act on an aura that was already
	// up when the parse ran.
	needsSim bool

	// The value multiplies rather than adds, so a category prices it by how far it is from 1.
	multiplicative bool

	// The name a per-stat category files the attachment under, which is the one core's own exclusive
	// stat buffs use: the stat's name, or the name of the PseudoStats field the row writes.
	key string

	// One attachment per stat, for a row that writes several: a per-stat category and a school's
	// resistance category each hold one stat. A part of a stat row bids its own value, the scale core's
	// exclusive stat buffs bid on.
	parts   []*attachment
	stat    stats.Stat
	statBid bool

	// The stats the attachment adds to or multiplies.
	stats []stats.Stat

	skipReason string
}

func skipped(reason string) *attachment {
	return &attachment{skipReason: reason}
}

// What the parser does with one effect, by the aura it applies. A row answers nil for an effect it
// cannot take, or skipped with the reason, which sends the effect to Skipped and to the report.
type row func(p *parser, e *Effect, v float64) *attachment

// The auras the parser knows, keyed the way the game files store them. A_ADD_FLAT_MODIFIER and
// A_ADD_PCT_MODIFIER name a spell property in their misc value and read the two tables below it;
// the percentage table is handed its value as a fraction. A percentage crit chance is the same
// percentage points the flat one adds, so it reads the flat row with the game files' own number.
var auraTable = map[dbcenums.EffectAuraType]row{
	dbcenums.A_ADD_FLAT_MODIFIER: func(p *parser, e *Effect, v float64) *attachment {
		return lookup(flatModTable, dbcenums.SpellModOp(e.Misc))(p, e, v)
	},
	dbcenums.A_ADD_PCT_MODIFIER: func(p *parser, e *Effect, v float64) *attachment {
		op := dbcenums.SpellModOp(e.Misc)
		if op == dbcenums.SPELLMOD_CRITICAL_CHANCE {
			return lookup(flatModTable, op)(p, e, v)
		}
		return lookup(pctModTable, op)(p, e, v/100)
	},

	// The damage the unit deals and takes, by school mask. A row that names ranged weapons alone is a
	// mod on the attacks made with them; one that names a melee weapon holds while it is equipped, and
	// also reaches the melee-defense spells of the schools its mask leaves out.
	dbcenums.A_MOD_DAMAGE_PERCENT_DONE: func(p *parser, e *Effect, v float64) *attachment {
		rule, scope, ok := p.weaponScope(e)
		switch {
		case !ok, rule == WeaponRuleWhileEquipped && p.character == nil && !p.dry:
			return nil
		case rule == WeaponRulePerWeapon:
			scope.School = core.SpellSchool(e.Misc)
			return p.scopedMod("SpellMod_DamageDone_Pct", core.SpellMod_DamageDone_Pct, scope, v/100)
		}

		a := p.schoolMultiplier("damage-dealt", "DamageDealtMultiplier", e.Misc, core.MultiplierFromPercent(v),
			&p.unit.PseudoStats.DamageDealtMultiplier, &p.unit.PseudoStats.SchoolDamageDealtMultiplier)
		if a == nil || rule != WeaponRuleWhileEquipped {
			return a
		}
		if !scope.ProcMask.Matches(core.ProcMaskMelee) || e.Misc&miscAllSchools == miscAllSchools {
			return p.whileEquipped(scope.Equip(), a)
		}
		melee := p.modFloat("damage-dealt-melee", core.SpellModConfig{
			Kind:             core.SpellMod_DamageDone_Pct,
			DefenseType:      core.DefenseTypeMelee,
			MultiplierSchool: core.SpellSchool(miscAllSchools &^ e.Misc),
		}, v/100)
		return p.whileEquipped(scope.Equip(), combined(a.kind+"+"+melee.kind, a, melee))
	},
	dbcenums.A_MOD_DAMAGE_PERCENT_TAKEN: func(p *parser, e *Effect, v float64) *attachment {
		return p.schoolMultiplier("damage-taken", "DamageTakenMultiplier", e.Misc, core.MultiplierFromPercent(v),
			&p.unit.PseudoStats.DamageTakenMultiplier, &p.unit.PseudoStats.SchoolDamageTakenMultiplier)
	},

	// Flat damage done, which the sim keeps as a stat per school.
	dbcenums.A_MOD_DAMAGE_DONE: func(p *parser, e *Effect, v float64) *attachment {
		return p.statsBuff(DamageDoneStats(e.Misc), v)
	},
	// The cost of the spells of the schools the mask names, in core's multiplicative cost bucket.
	// Whether a second stack multiplies again or adds again is not stated, so a stacking row is
	// skipped.
	dbcenums.A_MOD_POWER_COST_SCHOOL_PCT: func(p *parser, e *Effect, v float64) *attachment {
		if p.stacking || e.Misc == 0 {
			return nil
		}
		return p.modFloat("SpellMod_PowerCost_Pct", core.SpellModConfig{
			Kind:   core.SpellMod_PowerCost_Pct,
			School: core.SpellSchool(e.Misc),
		}, v/100)
	},

	// Flat damage taken, which the sim keeps once for physical and once for every spell school
	// together: every school is both, and a mask that names some spell schools and not others has no
	// field to land on.
	dbcenums.A_MOD_DAMAGE_TAKEN: func(p *parser, e *Effect, v float64) *attachment {
		physical := func() *attachment {
			return p.pseudoAdd("physical-damage-taken-flat", "BonusPhysicalDamageTaken",
				[]*float64{&p.unit.PseudoStats.BonusPhysicalDamageTaken}, v)
		}
		spell := func() *attachment {
			return p.pseudoAdd("spell-damage-taken-flat", "BonusSpellDamageTaken",
				[]*float64{&p.unit.PseudoStats.BonusSpellDamageTaken}, v)
		}
		switch {
		case e.Misc&miscAllSchools == miscAllSchools:
			return combined("damage-taken-flat", physical(), spell())
		case e.Misc&int32(core.SpellSchoolPhysical) != 0:
			return physical()
		case e.Misc&miscMagicSchool == miscMagicSchool:
			return spell()
		}
		return nil
	},

	// Melee and ranged attacks alone. The melee multiplier reaches a melee attack of any school; the
	// flat ranged amount reaches physical ranged attacks, and a mask without physical has no field.
	dbcenums.A_MOD_MELEE_DAMAGE_TAKEN_PCT: func(p *parser, e *Effect, v float64) *attachment {
		return p.pseudoMultiplier("melee-damage-taken", "MeleeDamageTakenMultiplier",
			[]*float64{&p.unit.PseudoStats.MeleeDamageTakenMultiplier}, core.MultiplierFromPercent(v))
	},
	dbcenums.A_MOD_RANGED_DAMAGE_TAKEN: func(p *parser, e *Effect, v float64) *attachment {
		if e.Misc&int32(core.SpellSchoolPhysical) == 0 {
			return nil
		}
		return p.pseudoAdd("ranged-damage-taken-flat", "BonusRangedDamageTaken",
			[]*float64{&p.unit.PseudoStats.BonusRangedDamageTaken}, v)
	},

	dbcenums.A_MOD_THREAT: func(p *parser, e *Effect, v float64) *attachment {
		return p.schoolMultiplier("threat", "ThreatMultiplier", e.Misc, core.MultiplierFromPercent(v),
			&p.unit.PseudoStats.ThreatMultiplier, &p.unit.PseudoStats.SchoolThreatMultiplier)
	},

	dbcenums.A_MOD_ATTACK_POWER: func(p *parser, e *Effect, v float64) *attachment {
		return p.statBuff(stats.AttackPower, v)
	},

	// A stat, the damage dealt or the crit damage dealt, against the creature types the misc value
	// masks.
	dbcenums.A_MOD_MELEE_ATTACK_POWER_VERSUS:  mobTypeStatRow,
	dbcenums.A_MOD_RANGED_ATTACK_POWER_VERSUS: mobTypeStatRow,
	dbcenums.A_MOD_FLAT_SPELL_DAMAGE_VERSUS:   mobTypeStatRow,
	dbcenums.A_MOD_DAMAGE_DONE_VERSUS:         damageVersusRow,
	dbcenums.A_MOD_CRIT_PERCENT_VERSUS:        critDamageVersusRow,

	dbcenums.A_MOD_RANGED_ATTACK_POWER: func(p *parser, e *Effect, v float64) *attachment {
		return p.statBuff(stats.RangedAttackPower, v)
	},

	// Ranged attack power for whoever shoots the unit, which the sim keeps on the unit shot.
	dbcenums.A_RANGED_ATTACK_POWER_ATTACKER_BONUS: func(p *parser, e *Effect, v float64) *attachment {
		return p.pseudoAdd("ranged-attack-power-attacker-bonus", "BonusRangedAttackPower",
			[]*float64{&p.unit.PseudoStats.BonusRangedAttackPower}, v)
	},

	// The ratings the misc value names as a mask of the game files' combat ratings, by the combinations
	// the sim keeps one stat for. A row that names the weapon it needs, and a mask of ranged bits
	// alone, hold for some attacks only, and become the percent the rating is worth on them.
	dbcenums.A_MOD_RATING: func(p *parser, e *Effect, v float64) *attachment {
		rule, scope, ok := p.weaponScope(e)
		if !ok {
			return nil
		}
		switch rule {
		case WeaponRulePerWeapon, WeaponRuleRangedRatings:
			return p.scopedRatings(e.Misc, scope, v)
		}
		return p.statsBuff(RatingStats(e.Misc), v)
	},

	dbcenums.A_MOD_STAT: func(p *parser, e *Effect, v float64) *attachment {
		return p.statsBuff(clientStatList(e.Misc), v)
	},
	dbcenums.A_MOD_PERCENT_STAT:          percentStatRow,
	dbcenums.A_MOD_TOTAL_STAT_PERCENTAGE: totalStatPercentRow,

	// Hit and crit, which the sim states in percentage points the way the game files do. A row that
	// names the weapon it needs is a mod on the attacks made with that weapon, except an A_MOD_CRIT_PCT
	// whose weapons are not ranged alone: that is crit on every spell and attack while one is equipped.
	dbcenums.A_MOD_HIT_CHANCE: func(p *parser, e *Effect, v float64) *attachment {
		if rule, scope, ok := p.weaponScope(e); !ok {
			return nil
		} else if rule == WeaponRulePerWeapon {
			return p.scopedMod("SpellMod_BonusHit_Percent", core.SpellMod_BonusHit_Percent, scope, v)
		}
		return p.statBuff(stats.PhysicalHitPercent, v)
	},
	dbcenums.A_MOD_SPELL_HIT_CHANCE: func(p *parser, e *Effect, v float64) *attachment {
		return p.statBuff(stats.SpellHitPercent, v)
	},
	dbcenums.A_MOD_WEAPON_CRIT_PERCENT: func(p *parser, e *Effect, v float64) *attachment {
		if rule, scope, ok := p.weaponScope(e); !ok {
			return nil
		} else if rule == WeaponRulePerWeapon {
			return p.scopedMod("SpellMod_BonusCrit_Percent", core.SpellMod_BonusCrit_Percent, scope, v)
		}
		return p.statBuff(stats.PhysicalCritPercent, v)
	},
	dbcenums.A_MOD_SPELL_CRIT_CHANCE: func(p *parser, e *Effect, v float64) *attachment {
		return p.statBuff(stats.SpellCritPercent, v)
	},
	dbcenums.A_MOD_CRIT_PCT: func(p *parser, e *Effect, v float64) *attachment {
		crit := []stats.Stat{stats.PhysicalCritPercent, stats.SpellCritPercent}
		rule, scope, ok := p.weaponScope(e)
		if !ok {
			return nil
		}
		switch rule {
		case WeaponRulePerWeapon:
			return p.scopedMod("SpellMod_BonusCrit_Percent", core.SpellMod_BonusCrit_Percent, scope, v)
		case WeaponRuleWhileEquipped:
			return p.statsWhileEquipped(crit, v, scope.Equip())
		}
		return p.statsBuff(crit, v)
	},

	// The haste auras. Each multiplies a speed the unit recomputes, so each needs the Simulation an
	// aura's gain hands over. Melee and ranged speed together is the sim's attack speed.
	dbcenums.A_MOD_CASTING_SPEED_NOT_STACK: func(p *parser, e *Effect, v float64) *attachment {
		return p.speed("cast-speed", "CastSpeedMultiplier", (*core.Unit).MultiplyCastSpeed, core.SymmetricMultiplierFromPercent(v))
	},
	dbcenums.A_MOD_MELEE_HASTE_3: func(p *parser, e *Effect, v float64) *attachment {
		return p.speed("melee-speed", "MeleeSpeedMultiplier", (*core.Unit).MultiplyMeleeSpeed, core.SymmetricMultiplierFromPercent(v))
	},
	dbcenums.A_MOD_ATTACKSPEED:          attackSpeedRow,
	dbcenums.A_MOD_MELEE_RANGED_HASTE_2: attackSpeedRow,
	dbcenums.A_MOD_RANGED_HASTE: func(p *parser, e *Effect, v float64) *attachment {
		return p.speed("ranged-speed", "RangedSpeedMultiplier", (*core.Unit).MultiplyRangedSpeed, core.SymmetricMultiplierFromPercent(v))
	},

	// Healing. 135 is the flat bonus to healing done, 136 the multiplier on it, and 118 the
	// multiplier on healing taken; 115 is the flat bonus to healing taken.
	dbcenums.A_MOD_HEALING_DONE: func(p *parser, e *Effect, v float64) *attachment {
		return p.statBuff(stats.HealingPower, v)
	},
	dbcenums.A_MOD_HEALING_DONE_PERCENT: func(p *parser, e *Effect, v float64) *attachment {
		return p.pseudoMultiplier("healing-dealt", "HealingDealtMultiplier",
			[]*float64{&p.unit.PseudoStats.HealingDealtMultiplier}, core.MultiplierFromPercent(v))
	},
	dbcenums.A_MOD_HEALING_PCT: func(p *parser, e *Effect, v float64) *attachment {
		return p.pseudoMultiplier("healing-taken", "HealingTakenMultiplier",
			[]*float64{&p.unit.PseudoStats.HealingTakenMultiplier}, core.MultiplierFromPercent(v))
	},
	dbcenums.A_MOD_HEALING: func(p *parser, e *Effect, v float64) *attachment {
		return p.pseudoAdd("healing-taken-flat", "BonusHealingTaken", []*float64{&p.unit.PseudoStats.BonusHealingTaken}, v)
	},

	// The chance a hit pushes a cast back, which the sim keeps as a chance that starts at 1: the
	// game files' 35% less pushback is -0.35 there.
	dbcenums.A_REDUCE_PUSHBACK: func(p *parser, e *Effect, v float64) *attachment {
		return p.pseudoAdd("pushback", "PushbackChance", []*float64{&p.unit.PseudoStats.PushbackChance}, -v/100)
	},

	// Mana regen, which the game files state per five seconds on the mana bar.
	dbcenums.A_MOD_POWER_REGEN: func(p *parser, e *Effect, v float64) *attachment {
		if e.Misc != int32(dbcenums.POWER_MANA) {
			return nil
		}
		return p.statBuff(stats.MP5, v)
	},
	// The share of spirit regen that continues while casting, which core caps at all of it, and a
	// multiplier on spirit regen. The energy bar's regen is a speed, which needs the Simulation an
	// aura's gain hands over.
	dbcenums.A_MOD_MANA_REGEN_INTERRUPT: func(p *parser, e *Effect, v float64) *attachment {
		return p.refreshingManaRegen(p.pseudoAdd("spirit-regen-casting", "SpiritRegenRateCasting",
			[]*float64{&p.unit.PseudoStats.SpiritRegenRateCasting}, v/100))
	},
	dbcenums.A_MOD_POWER_REGEN_PERCENT: func(p *parser, e *Effect, v float64) *attachment {
		switch e.Misc {
		case int32(dbcenums.POWER_MANA):
			return p.refreshingManaRegen(p.pseudoMultiplier("spirit-regen", "SpiritRegenMultiplier",
				[]*float64{&p.unit.PseudoStats.SpiritRegenMultiplier}, core.MultiplierFromPercent(v)))
		case int32(dbcenums.POWER_ENERGY):
			if !p.unit.HasEnergyBar() {
				return nil
			}
			return p.speed("energy-regen", "EnergyRegenMultiplier", (*core.Unit).MultiplyEnergyRegenSpeed,
				core.MultiplierFromPercent(v))
		}
		return nil
	},

	dbcenums.A_MOD_MAX_POWER:       maxPowerRow,
	dbcenums.A_MOD_INCREASE_ENERGY: maxPowerRow,

	dbcenums.A_MOD_INCREASE_HEALTH: flatHealthRow,
	dbcenums.A_MOD_MAX_HEALTH:      flatHealthRow,
	dbcenums.A_MOD_INCREASE_HEALTH_PERCENT: func(p *parser, e *Effect, v float64) *attachment {
		return p.statMultiplier([]stats.Stat{stats.Health}, core.MultiplierFromPercent(v))
	},

	// Armor is school 1 of the resistance aura; the other bits are the five magic resistances.
	dbcenums.A_MOD_RESISTANCE: func(p *parser, e *Effect, v float64) *attachment {
		return p.statsBuff(resistanceStats(e.Misc), v)
	},
	// Flat armor on the same armor bit, which the class armor buffs and debuffs state, as ArmorNoMods.
	// A mask naming any other school is skipped whole.
	dbcenums.A_MOD_ARMOR_NO_MODS: func(p *parser, e *Effect, v float64) *attachment {
		sts := ArmorNoModsStats(e.Misc)
		if sts == nil {
			return skipped(fmt.Sprintf("misc %d is not armor alone, the only school the row takes", e.Misc))
		}
		return p.statsBuff(sts, v)
	},

	// The game files state the armor ladder twice, once on base armor and once on bonus armor, and the
	// tooltip states armor from items, which is the equipment share the sim scales.
	dbcenums.A_MOD_BASE_RESISTANCE_PCT: func(p *parser, e *Effect, v float64) *attachment {
		if e.Misc&miscArmor == 0 {
			return nil
		}
		return p.equipScaling(stats.Armor, core.MultiplierFromPercent(v))
	},
	// The resistance percentage multiplies the whole of the stat but the A_MOD_ARMOR_NO_MODS armor,
	// which it leaves alone. Armor alone is taken: no row states it for a magic resistance.
	dbcenums.A_MOD_RESISTANCE_PCT: func(p *parser, e *Effect, v float64) *attachment {
		if e.Misc != miscArmor {
			return nil
		}
		return p.statMultiplier([]stats.Stat{stats.Armor}, core.MultiplierFromPercent(v))
	},

	// Armor, attack power, spell damage and healing worth a share of a stat. The spell damage aura names
	// its schools in the first misc value and the stat in the second, and only the mask of every magic
	// school has a stat to land on.
	dbcenums.A_MOD_ARMOR_PCT_FROM_STAT:                 statShareRow,
	dbcenums.A_MOD_ATTACK_POWER_OF_STAT_PERCENT:        statShareRow,
	dbcenums.A_MOD_RANGED_ATTACK_POWER_OF_STAT_PERCENT: statShareRow,
	dbcenums.A_MOD_SPELL_DAMAGE_OF_STAT_PERCENT: func(p *parser, e *Effect, v float64) *attachment {
		if e.Misc != miscMagicSchool {
			return nil
		}
		return p.statShare(clientStatList(e.Misc2), stats.SpellDamage, v/100)
	},
	dbcenums.A_MOD_SPELL_HEALING_OF_STAT_PERCENT: statShareRow,

	dbcenums.A_MOD_MAX_POWER_PCT: func(p *parser, e *Effect, v float64) *attachment {
		if e.Misc != int32(dbcenums.POWER_MANA) {
			return nil
		}
		return p.statMultiplier([]stats.Stat{stats.Mana}, core.MultiplierFromPercent(v))
	},

	dbcenums.A_MOD_BLOCK_VALUE_PCT: func(p *parser, e *Effect, v float64) *attachment {
		return p.pseudoMultiplier("block-value", "BlockValueMultiplier",
			[]*float64{&p.unit.PseudoStats.BlockValueMultiplier}, core.MultiplierFromPercent(v))
	},

	// The chance to be hit by melee, which the sim keeps in percentage points as a reduction that
	// ranged attacks read too. The ranged aura has no field of its own, and a row stating both would
	// count the one field twice.
	dbcenums.A_MOD_ATTACKER_MELEE_HIT_CHANCE: func(p *parser, e *Effect, v float64) *attachment {
		return p.pseudoAdd("physical-hit-taken", "ReducedPhysicalHitTakenChance",
			[]*float64{&p.unit.PseudoStats.ReducedPhysicalHitTakenChance}, -v)
	},

	// The target's resistance the unit's attacks ignore, by core.SpellSchool.PiercingStat, and a mask
	// without a stat as the per-school piercing of each school it names. The sign differs between rows
	// while the tooltip states the amount pierced either way, so the amount is its size.
	dbcenums.A_MOD_TARGET_RESISTANCE: func(p *parser, e *Effect, v float64) *attachment {
		v = math.Abs(v)
		if stat, ok := core.SpellSchool(e.Misc).PiercingStat(); ok {
			return p.statBuff(stat, v)
		}
		if e.Misc&miscArmor != 0 {
			return nil
		}

		var fields []*float64
		for _, index := range schoolIndexes(e.Misc) {
			fields = append(fields, &p.unit.PseudoStats.SchoolSpellPiercing[index])
		}
		if len(fields) == 0 {
			return nil
		}
		// Piercing for other schools does not compete with this one.
		return p.pseudoAdd("school-spell-piercing", fmt.Sprintf("SchoolSpellPiercing%d", e.Misc), fields, v)
	},

	// Avoidance, in percentage points.
	dbcenums.A_MOD_BLOCK_PERCENT: func(p *parser, e *Effect, v float64) *attachment {
		return p.statBuff(stats.BlockPercent, v)
	},
	dbcenums.A_MOD_DODGE_PERCENT: func(p *parser, e *Effect, v float64) *attachment {
		return p.statBuff(stats.DodgePercent, v)
	},
	dbcenums.A_MOD_PARRY_PERCENT: func(p *parser, e *Effect, v float64) *attachment {
		return p.statBuff(stats.ParryPercent, v)
	},

	// Skill points: Defense as defense rating, and a weapon skill on top of the skill the player states.
	// A profession skill is skipped.
	dbcenums.A_MOD_SKILL: func(p *parser, e *Effect, v float64) *attachment {
		if e.Misc == miscDefense {
			return p.statBuff(stats.DefenseRating, v*core.DefenseRatingPerDefenseLevel)
		}
		skill, ok := weaponSkillLines[e.Misc]
		if !ok || p.character == nil {
			return nil
		}
		return p.weaponSkill(skill, v)
	},

	// The off-hand bonus is a damage modifier on the off-hand's own hits rather than on a family of
	// spells, so it carries a proc mask instead of the effect's class flags.
	dbcenums.A_MOD_OFFHAND_DAMAGE_PCT: func(p *parser, e *Effect, v float64) *attachment {
		return p.modFloat("SpellMod_DamageDone_Pct", core.SpellModConfig{
			Kind:     core.SpellMod_DamageDone_Pct,
			ProcMask: core.ProcMaskMeleeOH,
		}, v/100)
	},

	// The share of every defender's armor the unit's attacks ignore, which the sim keeps on the attack
	// tables as a fraction.
	dbcenums.A_SPELL_AURA_MOD_ARMOR_PENETRATION_PCT: func(p *parser, e *Effect, v float64) *attachment {
		a := additive("armor-ignore", v/100, p.onAttackTables(func(_ *core.Simulation, delta float64) {
			p.unit.AddArmorIgnoreFactor(delta)
		}))
		a.key = "ArmorIgnoreFactor"
		return a
	},

	// The game files state expertise as the percent it takes off dodge and parry.
	dbcenums.A_MOD_EXPERTISE: func(p *parser, e *Effect, v float64) *attachment {
		return p.statBuff(stats.ExpertisePercent, v)
	},

	// The damage the unit takes from the attacks of the aura's caster, and the caster's crit chance
	// against it: see FromCaster.
	dbcenums.A_MOD_SCHOOL_MASK_DAMAGE_FROM_CASTER: func(p *parser, e *Effect, v float64) *attachment {
		if e.Misc <= 0 {
			return nil
		}
		school := core.SpellSchool(e.Misc)
		return p.casterDamage("caster-damage-by-school", v, func(spell *core.Spell) bool {
			return spell.SpellSchool.Matches(school)
		})
	},
	dbcenums.A_MOD_SPELL_DAMAGE_FROM_CASTER: func(p *parser, e *Effect, v float64) *attachment {
		return p.casterDamage("caster-damage-by-spell", v, p.namesSpell(e))
	},
	dbcenums.A_MOD_CRIT_CHANCE_FOR_CASTER_WITH_ABILITIES: func(p *parser, e *Effect, v float64) *attachment {
		return p.casterCrit("caster-crit-by-spell", v, p.namesSpell(e))
	},

	// How long a crowd control effect lasts on the unit, by the mechanic the misc value names.
	dbcenums.A_MECHANIC_DURATION_MOD: func(p *parser, e *Effect, v float64) *attachment {
		switch e.Misc {
		case int32(dbcenums.MECHANIC_FEAR):
			return p.pseudoMultiplier("fear-duration", "FearDurationMultiplier",
				[]*float64{&p.unit.PseudoStats.FearDurationMultiplier}, core.MultiplierFromPercent(v))
		case int32(dbcenums.MECHANIC_STUN):
			return p.pseudoMultiplier("stun-duration", "StunDurationMultiplier",
				[]*float64{&p.unit.PseudoStats.StunDurationMultiplier}, core.MultiplierFromPercent(v))
		}
		return nil
	},
}

// A_ADD_FLAT_MODIFIER: the game files' own amount, in the property's units.
var flatModTable = map[dbcenums.SpellModOp]row{
	dbcenums.SPELLMOD_COST: func(p *parser, e *Effect, v float64) *attachment {
		cfg := p.modConfig(e, core.SpellMod_Custom)
		cfg.ApplyCustom = func(mod *core.SpellMod, spell *core.Spell) { addFlatCost(spell, mod.GetIntValue()) }
		cfg.RemoveCustom = func(mod *core.SpellMod, spell *core.Spell) { addFlatCost(spell, -mod.GetIntValue()) }
		return p.modInt("SpellMod_PowerCost_Flat", cfg, v)
	},
	dbcenums.SPELLMOD_CASTING_TIME: func(p *parser, e *Effect, v float64) *attachment {
		return p.modTime("SpellMod_CastTime_Flat", p.modConfig(e, core.SpellMod_CastTime_Flat), v)
	},
	dbcenums.SPELLMOD_COOLDOWN: func(p *parser, e *Effect, v float64) *attachment {
		return p.modTime("SpellMod_Cooldown_Flat", p.modConfig(e, core.SpellMod_Cooldown_Flat), v)
	},
	dbcenums.SPELLMOD_GLOBAL_COOLDOWN: func(p *parser, e *Effect, v float64) *attachment {
		return p.modTime("SpellMod_GlobalCooldown_Flat", p.modConfig(e, core.SpellMod_GlobalCooldown_Flat), v)
	},
	dbcenums.SPELLMOD_ACTIVATION_TIME: func(p *parser, e *Effect, v float64) *attachment {
		return p.modTime("SpellMod_DotTickLength_Flat", p.modConfig(e, core.SpellMod_DotTickLength_Flat), v)
	},
	dbcenums.SPELLMOD_CRITICAL_CHANCE: func(p *parser, e *Effect, v float64) *attachment {
		return p.modFloat("SpellMod_BonusCrit_Percent", p.modConfig(e, core.SpellMod_BonusCrit_Percent), v)
	},
	dbcenums.SPELLMOD_RESIST_MISS_CHANCE: func(p *parser, e *Effect, v float64) *attachment {
		cfg := p.modConfig(e, core.SpellMod_BonusHit_Percent)
		cfg.ShowAsSchoolHit = p.showAsSchoolHit
		p.showAsSchoolHitStamped = true
		return p.modFloat("SpellMod_BonusHit_Percent", cfg, v)
	},
	dbcenums.SPELLMOD_DURATION: func(p *parser, e *Effect, v float64) *attachment {
		return p.modTime("SpellMod_Duration_Flat", p.modConfig(e, core.SpellMod_Duration_Flat), v)
	},
	dbcenums.SPELLMOD_CHARGES: func(p *parser, e *Effect, v float64) *attachment {
		return p.modInt("SpellMod_BuffMaxStacks_Flat", p.modConfig(e, core.SpellMod_BuffMaxStacks_Flat), v)
	},
	dbcenums.SPELLMOD_RANGE: func(p *parser, e *Effect, v float64) *attachment {
		return p.rangeMod("SpellMod_Range_Flat", p.modConfig(e, core.SpellMod_Range_Flat), v)
	},
	// The chance of the procs the class mask names, in percentage points: the rows the mask names are
	// the proc auras themselves, and a trigger built from one rolls at the raised chance, wherever it
	// sits: on the caster, on a pet or on the target.
	dbcenums.SPELLMOD_CHANCE_OF_SUCCESS: func(p *parser, e *Effect, v float64) *attachment {
		return p.modFloat("SpellMod_ProcChance_Flat", p.modConfig(e, core.SpellMod_ProcChance_Flat), v)
	},
	dbcenums.SPELLMOD_EFFECT1: flatEffectAmount,
	dbcenums.SPELLMOD_EFFECT2: flatEffectAmount,
	dbcenums.SPELLMOD_EFFECT3: flatEffectAmount,
}

// A_ADD_PCT_MODIFIER: a percentage the game files state as an integer, handed over as a fraction. Two such
// rows on one spell multiply, for cost, direct damage and periodic damage alike, and a flat
// A_ADD_FLAT_MODIFIER on the same spell applies before any percent. The aura type decides which table a row
// takes, whatever its tooltip reads: Unleashed Fury 19616 describes a percent and its row is a flat 107
// modifier, so the flat value is what the sim applies and the tooltip's number is flagged, not adopted.
var pctModTable = map[dbcenums.SpellModOp]row{
	dbcenums.SPELLMOD_DAMAGE: func(p *parser, e *Effect, v float64) *attachment {
		// Direct hits only: ticks take SPELLMOD_DOT, which a talent states beside it when it raises both.
		return p.modFloat("SpellMod_DirectDamageDone_Pct", p.modConfig(e, core.SpellMod_DirectDamageDone_Pct), v)
	},
	dbcenums.SPELLMOD_ALL_EFFECTS: pctDamageDone,
	dbcenums.SPELLMOD_DOT: func(p *parser, e *Effect, v float64) *attachment {
		return p.modFloat("SpellMod_DotDamageDone_Pct", p.modConfig(e, core.SpellMod_DotDamageDone_Pct), v)
	},
	dbcenums.SPELLMOD_COST: func(p *parser, e *Effect, v float64) *attachment {
		return p.modFloat("SpellMod_PowerCost_Pct", p.modConfig(e, core.SpellMod_PowerCost_Pct), v)
	},
	dbcenums.SPELLMOD_CASTING_TIME: func(p *parser, e *Effect, v float64) *attachment {
		return p.modFloat("SpellMod_CastTime_Pct", p.modConfig(e, core.SpellMod_CastTime_Pct), v)
	},
	dbcenums.SPELLMOD_COOLDOWN: func(p *parser, e *Effect, v float64) *attachment {
		// The field is the multiplier itself, not a bonus on top of one.
		return p.modMultiplier("SpellMod_Cooldown_Multiplier", p.modConfig(e, core.SpellMod_Cooldown_Multiplier), 1+v)
	},
	dbcenums.SPELLMOD_CRIT_DAMAGE_BONUS: func(p *parser, e *Effect, v float64) *attachment {
		return p.modFloat("SpellMod_CritMultiplier_Flat", p.modConfig(e, core.SpellMod_CritMultiplier_Flat), v)
	},
	dbcenums.SPELLMOD_THREAT: func(p *parser, e *Effect, v float64) *attachment {
		return p.modFloat("SpellMod_ThreatMultiplier_Pct", p.modConfig(e, core.SpellMod_ThreatMultiplier_Pct), v)
	},
	dbcenums.SPELLMOD_DURATION: func(p *parser, e *Effect, v float64) *attachment {
		return p.modFloat("SpellMod_DotBaseDuration_Pct", p.modConfig(e, core.SpellMod_DotBaseDuration_Pct), v)
	},
	dbcenums.SPELLMOD_CHANCE_OF_SUCCESS: func(p *parser, e *Effect, v float64) *attachment {
		return p.modFloat("SpellMod_ProcChance_Pct", p.modConfig(e, core.SpellMod_ProcChance_Pct), v)
	},
	dbcenums.SPELLMOD_PROC_COOLDOWN: func(p *parser, e *Effect, v float64) *attachment {
		return p.modFloat("SpellMod_ProcCooldown_Pct", p.modConfig(e, core.SpellMod_ProcCooldown_Pct), v)
	},
	dbcenums.SPELLMOD_GLOBAL_COOLDOWN: func(p *parser, e *Effect, v float64) *attachment {
		return p.modFloat("SpellMod_GlobalCooldown_Pct", p.modConfig(e, core.SpellMod_GlobalCooldown_Pct), v)
	},
	dbcenums.SPELLMOD_RANGE: func(p *parser, e *Effect, v float64) *attachment {
		return p.rangeMod("SpellMod_Range_Pct", p.modConfig(e, core.SpellMod_Range_Pct), v)
	},
	dbcenums.SPELLMOD_EFFECT1: pctEffectAmount,
	dbcenums.SPELLMOD_EFFECT2: pctEffectAmount,
	dbcenums.SPELLMOD_EFFECT3: pctEffectAmount,
}

// SPELLMOD_ALL_EFFECTS raises every hit the spell deals, direct and periodic, which the sim sums into
// spell.DamageMultiplierAdditive.
func pctDamageDone(p *parser, e *Effect, v float64) *attachment {
	return p.modFloat("SpellMod_DamageDone_Flat", p.modConfig(e, core.SpellMod_DamageDone_Flat), v)
}

// SPELLMOD_EFFECT1/2/3 name one effect of the target spell, and which effect carries the damage is a
// property of that spell rather than of this one. Reading it would mean resolving every spell the
// class mask names, so the parser assumes the named effect is the damage and says so in the kind: a
// caller whose spell states something else there has to wire that effect by hand.
func flatEffectAmount(p *parser, e *Effect, v float64) *attachment {
	return p.modFloat(effectAssumedKind(e.Misc, "SpellMod_BaseDamage_Flat"),
		p.modConfig(e, core.SpellMod_BaseDamage_Flat), v)
}

func pctEffectAmount(p *parser, e *Effect, v float64) *attachment {
	return p.modFloat(effectAssumedKind(e.Misc, "SpellMod_DamageDone_Flat"),
		p.modConfig(e, core.SpellMod_DamageDone_Flat), v)
}

func effectAssumedKind(misc int32, kind string) string {
	n := 1
	switch dbcenums.SpellModOp(misc) {
	case dbcenums.SPELLMOD_EFFECT2:
		n = 2
	case dbcenums.SPELLMOD_EFFECT3:
		n = 3
	}
	return "effect" + string(rune('0'+n)) + "-assumed-damage " + kind
}

// The row for a misc value, or one that skips: a modifier op the table does not know is reported the
// same way an unknown aura is.
func lookup(table map[dbcenums.SpellModOp]row, op dbcenums.SpellModOp) row {
	if r, ok := table[op]; ok {
		return r
	}
	return skipRow
}

func skipRow(_ *parser, _ *Effect, _ float64) *attachment {
	return nil
}

// The most power the unit's bar holds, by the power type the misc value names. The game files state
// rage on a 0-1000 bar and energy as it is; mana is the Mana stat, which the sim keeps as the maximum.
// A unit without the rage or energy bar the row names has no maximum to raise, so the row is skipped
// there. A dry run's unit has no bar at all, so it answers what a unit with the bar would attach.
func maxPowerRow(p *parser, e *Effect, v float64) *attachment {
	switch dbcenums.PowerType(e.Misc) {
	case dbcenums.POWER_MANA:
		return p.statBuff(stats.Mana, v)
	case dbcenums.POWER_RAGE:
		if !p.dry && !p.unit.HasRageBar() {
			return nil
		}
		metrics := p.powerMetrics(p.unit.NewRageMetrics)
		return maxPower("max-rage", "MaxRage", v/10, func(sim *core.Simulation, delta float64) {
			p.unit.SetMaximumRage(sim, p.unit.MaximumRage()+delta, metrics)
		})
	case dbcenums.POWER_ENERGY:
		if !p.dry && !p.unit.HasEnergyBar() {
			return nil
		}
		metrics := p.powerMetrics(p.unit.NewEnergyMetrics)
		return maxPower("max-energy", "MaxEnergy", v, func(sim *core.Simulation, delta float64) {
			p.unit.SetMaximumEnergy(sim, p.unit.MaximumEnergy()+delta, metrics)
		})
	}
	return nil
}

// The metrics a lowered maximum counts the lost power against. Only a parse a Simulation reaches -
// an aura's, or a conditional one's Refresh - can lose any, and metrics registered on a unit are
// reported, so the static path leaves them out.
func (p *parser) powerMetrics(newMetrics func(core.ActionID) *core.ResourceMetrics) *core.ResourceMetrics {
	if p.dry || (p.static && !p.conditional) {
		return nil
	}
	actionID := core.ActionID{SpellID: p.spell.ID}
	if p.aura != nil {
		actionID = p.aura.ActionID
	}
	return newMetrics(actionID)
}

func maxPower(kind string, key string, v float64, apply func(sim *core.Simulation, delta float64)) *attachment {
	a := additive(kind, v, apply)
	a.key = key
	return a
}

func percentStatRow(p *parser, e *Effect, v float64) *attachment {
	return p.statMultiplier(clientStatList(e.Misc), core.MultiplierFromPercent(v))
}

func totalStatPercentRow(p *parser, e *Effect, v float64) *attachment {
	return p.statMultiplier(e.TotalStatPercentStats(), core.MultiplierFromPercent(v))
}

// The spells a modifier effect names. The game files leave the mask empty on an effect that means the
// caster's whole family, and a spell with no family of its own leaves the mod unrestricted.
func (p *parser) modConfig(e *Effect, kind core.SpellModType) core.SpellModConfig {
	return core.SpellModConfig{Kind: kind, ClassFlags: p.classFlags(e)}
}

func (p *parser) classFlags(e *Effect) core.ClassFlags {
	if e.ClassFlags.IsZero() && p.spell.ClassFlags.Family != 0 {
		return core.ClassFlags{
			Family: p.spell.ClassFlags.Family,
			Mask:   [4]uint32{^uint32(0), ^uint32(0), ^uint32(0), ^uint32(0)},
		}
	}
	return e.ClassFlags
}

// Whether a spell is one the effect names, the way a modifier's class flags reach it.
func (p *parser) namesSpell(e *Effect) func(*core.Spell) bool {
	flags := p.classFlags(e)
	if flags.IsZero() {
		return func(*core.Spell) bool { return true }
	}
	return func(spell *core.Spell) bool {
		return spell.MatchesFlags(flags)
	}
}

// A flat cost modifier on one spell, in the units of the power that spell costs. The effect states no
// power type and the game files state rage on a 0-1000 bar, so a -30 is 3 rage off a rage cost and 30
// off any other (Shredding Attacks takes 18 Energy off Shred and 30 off Lacerate's 150 Rage).
func addFlatCost(spell *core.Spell, v int32) {
	if spell.Cost == nil {
		return
	}
	if _, ok := spell.Cost.ResourceCostImpl.(*core.RageCost); ok {
		v /= 10
	}
	spell.Cost.FlatModifier += v
}

// WeaponScope, narrowed to the hand OnHand names for a row bound to a weapon. Answers false for a row
// that hand cannot carry.
func (p *parser) weaponScope(e *Effect) (WeaponRule, core.SpellModConfig, bool) {
	rule, scope := WeaponScope(p.spell, e)
	if p.hand == 0 || (rule != WeaponRulePerWeapon && rule != WeaponRuleWhileEquipped) {
		return rule, scope, true
	}

	hand := p.hand
	if scope.EquipHand != 0 {
		hand &= scope.EquipHand
	}
	scope.EquipHand = hand
	if rule == WeaponRulePerWeapon {
		scope.ProcMask &= hand
	}
	return rule, scope, scope.EquipHand != 0 && (rule != WeaponRulePerWeapon || scope.ProcMask != 0)
}

func (p *parser) scopedMod(kind string, modKind core.SpellModType, scope core.SpellModConfig, v float64) *attachment {
	scope.Kind = modKind
	return p.modFloat(kind, scope, v)
}

// The hit and crit a weapon-bound or ranged-only rating mask names, as mods on the attacks the scope
// reaches. Its haste has no mod that reaches some attacks and not others, so the parse leaves it
// unread, and the other ratings too.
func (p *parser) scopedRatings(mask int32, scope core.SpellModConfig, v float64) *attachment {
	scope.FromRating = true
	var parts []*attachment
	for _, rating := range ScopedRatingPercents(mask, v) {
		switch rating.Stat {
		case stats.HitRating:
			parts = append(parts, p.scopedMod("SpellMod_BonusHit_Percent", core.SpellMod_BonusHit_Percent, scope, rating.Percent))
		case stats.CritRating:
			parts = append(parts, p.scopedMod("SpellMod_BonusCrit_Percent", core.SpellMod_BonusCrit_Percent, scope, rating.Percent))
		}
	}

	switch len(parts) {
	case 0:
		return nil
	case 1:
		return parts[0]
	}
	kinds := make([]string, len(parts))
	for i, part := range parts {
		kinds[i] = part.kind
	}
	return combined(strings.Join(kinds, "+"), parts...)
}

// One attachment for parts that apply together, its value the first part's.
func combined(kind string, parts ...*attachment) *attachment {
	a := &attachment{kind: kind, value: parts[0].value, parts: parts}
	a.set = func(sim *core.Simulation, level float64) {
		for _, part := range a.parts {
			part.set(sim, level)
		}
	}
	return a
}

// A spell mod the parse turns on and off. AddDynamicMod rather than AddStaticMod even on a passive:
// a mod that can be switched off is what lets an expiring aura and a conditional share one path.
func (p *parser) mod(kind string, cfg core.SpellModConfig, value float64, scale func(*core.SpellMod, float64)) *attachment {
	a := &attachment{kind: kind, value: value}
	if p.dry {
		return a
	}

	mod := p.unit.AddDynamicMod(cfg)
	a.set = func(_ *core.Simulation, level float64) {
		if level <= 0 {
			mod.Deactivate()
			return
		}
		scale(mod, level)
		mod.Activate()
	}
	return a
}

func (p *parser) modFloat(kind string, cfg core.SpellModConfig, v float64) *attachment {
	cfg.FloatValue = v
	return p.mod(kind, cfg, v, func(mod *core.SpellMod, level float64) {
		mod.UpdateFloatValue(v * level)
	})
}

// A range mod moves the ranged auto attack's reach with it, so the auto attack checks its range again
// whenever the mod turns on or off.
func (p *parser) rangeMod(kind string, cfg core.SpellModConfig, v float64) *attachment {
	a := p.modFloat(kind, cfg, v)
	if a.set == nil {
		return a
	}
	set := a.set
	a.set = func(sim *core.Simulation, level float64) {
		set(sim, level)
		if sim != nil {
			p.unit.AutoAttacks.UpdateRangedRange(sim)
		}
	}
	return a
}

func (p *parser) modInt(kind string, cfg core.SpellModConfig, v float64) *attachment {
	cfg.IntValue = int32(v)
	return p.mod(kind, cfg, v, func(mod *core.SpellMod, level float64) {
		mod.UpdateIntValue(int32(v * level))
	})
}

// The game files state a time modifier in milliseconds, which is also what Value reports.
func (p *parser) modTime(kind string, cfg core.SpellModConfig, ms float64) *attachment {
	cfg.TimeValue = core.DurationFromMillis(ms)
	return p.mod(kind, cfg, ms, func(mod *core.SpellMod, level float64) {
		mod.UpdateTimeValue(core.DurationFromMillis(ms * level))
	})
}

// A mod whose field is the multiplier itself rather than a bonus on top of one. Whether a second
// stack multiplies again or adds again is not stated anywhere, so a stacking row is skipped the way
// the other multiplier rows are.
func (p *parser) modMultiplier(kind string, cfg core.SpellModConfig, mult float64) *attachment {
	if p.stacking {
		return nil
	}

	cfg.FloatValue = mult
	a := p.mod(kind, cfg, mult, func(mod *core.SpellMod, _ float64) {
		mod.UpdateFloatValue(mult)
	})
	a.multiplicative = true
	return a
}

func (p *parser) statBuff(stat stats.Stat, v float64) *attachment {
	return p.statsBuff([]stats.Stat{stat}, v)
}

// Flat stats, which follow the stacks: the attachment keeps what it has handed out and adds the
// difference, so a level of 2 is worth twice the game files' amount.
func (p *parser) statsBuff(sts []stats.Stat, v float64) *attachment {
	if len(sts) == 0 {
		return nil
	}

	a := additive("stat "+statNames(sts), v, p.addStats(sts))
	a.stats = sts
	for _, stat := range sts {
		part := additive("stat "+stat.StatName(), v, p.addStats([]stats.Stat{stat}))
		part.key, part.stat, part.statBid = stat.StatName(), stat, true
		a.parts = append(a.parts, part)
	}
	return a
}

// Flat stats that hold only while a weapon meeting the requirement is equipped, as well as while the
// parse is on. They follow item swaps. A unit with no character wears no weapon, so the row is skipped
// there.
func (p *parser) statsWhileEquipped(sts []stats.Stat, v float64, req core.EquipRequirement) *attachment {
	if p.character == nil && !p.dry {
		return nil
	}

	a := &attachment{kind: "stat-while-equipped " + statNames(sts), value: v, stats: sts}
	var gated *core.EquipGatedStats
	if !p.dry {
		bonus := stats.Stats{}
		for _, stat := range sts {
			bonus[stat] = v
		}
		gated = p.character.NewEquipGatedStats(req, bonus)
		a.set = gated.SetLevel
	}
	for _, stat := range sts {
		part := &attachment{kind: "stat-while-equipped " + stat.StatName(), value: v, stats: []stats.Stat{stat},
			key: stat.StatName(), stat: stat, statBid: true}
		if gated != nil {
			part.set = func(sim *core.Simulation, level float64) { gated.SetStatLevel(sim, stat, level) }
		}
		a.parts = append(a.parts, part)
	}
	return a
}

func attackSpeedRow(p *parser, _ *Effect, v float64) *attachment {
	return p.speed("attack-speed", "AttackSpeedMultiplier", (*core.Unit).MultiplyAttackSpeed, core.SymmetricMultiplierFromPercent(v))
}

// Flat health. The bear forms state it on A_MOD_MAX_HEALTH; Last Stand states a percentage on
// A_MOD_INCREASE_HEALTH, which the warrior reads by hand rather than parsing.
func flatHealthRow(p *parser, _ *Effect, v float64) *attachment {
	return p.statBuff(stats.Health, v)
}

// The stat each stat-share aura adds to, a share of the stat its misc value names.
var statShareTargets = map[dbcenums.EffectAuraType]stats.Stat{
	dbcenums.A_MOD_ARMOR_PCT_FROM_STAT:                 stats.Armor,
	dbcenums.A_MOD_ATTACK_POWER_OF_STAT_PERCENT:        stats.AttackPower,
	dbcenums.A_MOD_RANGED_ATTACK_POWER_OF_STAT_PERCENT: stats.RangedAttackPower,
	dbcenums.A_MOD_SPELL_HEALING_OF_STAT_PERCENT:       stats.HealingPower,
}

func statShareRow(p *parser, e *Effect, v float64) *attachment {
	return p.statShare(clientStatList(e.Misc), statShareTargets[e.Aura], v/100)
}

// A stat the unit has against the mob types a creature type mask names, which the sim keeps on the
// attack tables.
func mobTypeStatRow(p *parser, e *Effect, v float64) *attachment {
	stat := versusStats[e.Aura]
	mobTypes := MobTypes(e.Misc)
	if len(mobTypes) == 0 {
		return nil
	}

	a := additive("mob-type-stat "+stat.StatName(), v, p.onAttackTables(func(_ *core.Simulation, delta float64) {
		var bonus stats.Stats
		bonus[stat] = delta
		p.unit.AddMobTypeBonusStats(bonus, mobTypes...)
	}))
	// A bonus against other creature types does not compete with this one.
	a.key = fmt.Sprintf("%sVersus%d", stat.StatName(), e.Misc)
	return a
}

// A multiplier on the damage the unit deals to the mob types a creature type mask names, which the
// sim keeps on the attack tables.
func damageVersusRow(p *parser, e *Effect, v float64) *attachment {
	return p.versusMultiplier("damage-dealt-versus", "DamageDealtVersus", e, v, p.unit.MultiplyDamageVersus)
}

// A multiplier on the crit damage the unit deals to the mob types a creature type mask names, on
// top of the damage multiplier: a crit against them deals that share more than it would.
func critDamageVersusRow(p *parser, e *Effect, v float64) *attachment {
	return p.versusMultiplier("crit-damage-versus", "CritDamageVersus", e, v, p.unit.MultiplyCritDamageVersus)
}

// A multiplier against the mob types a creature type mask names, on the attack tables. Whether a
// second stack multiplies again is not stated, so a stacking row is skipped.
func (p *parser) versusMultiplier(kind string, key string, e *Effect, v float64, multiply func(mult float64, mobTypes ...proto.MobType)) *attachment {
	mobTypes := MobTypes(e.Misc)
	mult := core.MultiplierFromPercent(v)
	if p.stacking || len(mobTypes) == 0 || mult <= 0 {
		return nil
	}

	a := multiplier(kind, mult, p.onAttackTables(func(_ *core.Simulation, factor float64) {
		multiply(factor, mobTypes...)
	}))
	a.key = fmt.Sprintf("%s%d", key, e.Misc)
	return a
}

// The attack tables exist once the environment is finalized, so the static path, which runs while
// the unit is being built, acts on them after that; an aura's gain acts at once.
func (p *parser) onAttackTables(apply func(sim *core.Simulation, v float64)) func(sim *core.Simulation, v float64) {
	if !p.static {
		return apply
	}
	return func(sim *core.Simulation, v float64) {
		if env := p.unit.Env; env != nil && !env.IsFinalized() {
			env.RegisterPostFinalizeEffect(func() { apply(sim, v) })
			return
		}
		apply(sim, v)
	}
}

func (p *parser) addStats(sts []stats.Stat) func(sim *core.Simulation, delta float64) {
	return func(sim *core.Simulation, delta float64) {
		bonus := stats.Stats{}
		for _, stat := range sts {
			bonus[stat] = delta
		}
		if sim == nil {
			p.unit.AddStats(bonus)
		} else {
			p.unit.AddStatsDynamic(sim, bonus)
		}
	}
}

// A multiplier on a stat, which the sim states as a dependency.
func (p *parser) statMultiplier(sts []stats.Stat, mult float64) *attachment {
	if len(sts) == 0 {
		return nil
	}

	a := &attachment{kind: "multiply-stat " + statNames(sts), value: mult, multiplicative: true, stats: sts}
	for _, stat := range sts {
		a.parts = append(a.parts, &attachment{kind: "multiply-stat " + stat.StatName(), value: mult, multiplicative: true,
			key: stat.StatName(), stat: stat, statBid: true})
	}
	return p.dependencies(a, func() {
		for _, stat := range sts {
			p.unit.MultiplyStat(stat, mult)
		}
	}, func() (toggles []func(sim *core.Simulation, level float64)) {
		for _, part := range a.parts {
			part.set = p.statDependency(p.unit.NewDynamicMultiplyStat(part.stat, mult))
			toggles = append(toggles, part.set)
		}
		return toggles
	})
}

// A share of each source stat added to another stat, which the sim states as a dependency.
func (p *parser) statShare(sources []stats.Stat, stat stats.Stat, share float64) *attachment {
	if len(sources) == 0 {
		return nil
	}

	a := &attachment{kind: "stat-share " + statNames(sources) + ">" + stat.StatName(), value: share,
		stats: []stats.Stat{stat}}
	return p.dependencies(a, func() {
		for _, source := range sources {
			p.unit.AddStatDependency(source, stat, share)
		}
	}, func() (toggles []func(sim *core.Simulation, level float64)) {
		for _, source := range sources {
			toggles = append(toggles, p.statDependency(p.unit.NewDynamicStatDependency(source, stat, share)))
		}
		return toggles
	})
}

// Turns an attachment on through stat dependencies. A dependency cannot carry a per-stack value, so a
// stacking row is skipped rather than attached at one stack. The static path has no aura to follow and
// no Simulation to answer a later Refresh with, so it applies once, and a conditional row is skipped
// there. Off the static path the toggles are built unless the parse is dry, and the unit's temporary
// stats listeners hear the stats they add or remove, measured at that moment.
func (p *parser) dependencies(a *attachment, applyStatic func(), newToggles func() []func(sim *core.Simulation, level float64)) *attachment {
	if p.stacking || (p.static && p.conditional) {
		return nil
	}

	if p.static {
		applied := false
		a.set = func(_ *core.Simulation, level float64) {
			if level <= 0 || applied {
				return
			}
			applied = true
			applyStatic()
		}
		return a
	}

	if p.dry {
		return a
	}
	toggles := newToggles()
	active := false
	a.set = func(sim *core.Simulation, level float64) {
		if (level > 0) == active {
			return
		}
		active = level > 0
		p.announceStats(sim, func() {
			for _, toggle := range toggles {
				toggle(sim, level)
			}
		})
	}
	return a
}

// Runs a change to the unit's stats and hands the unit's temporary stats listeners the difference it
// made, measured at that moment.
func (p *parser) announceStats(sim *core.Simulation, change func()) {
	if sim == nil || len(p.unit.OnTemporaryStatsChanges) == 0 {
		change()
		return
	}

	before := p.unit.GetStats()
	change()
	delta := p.unit.GetStats().Subtract(before)
	for _, onChange := range p.unit.OnTemporaryStatsChanges {
		onChange(sim, p.aura, delta)
	}
}

func (p *parser) statDependency(dep *stats.StatDependency) func(sim *core.Simulation, level float64) {
	active := false
	return func(sim *core.Simulation, level float64) {
		if (level > 0) == active {
			return
		}
		active = level > 0
		if active {
			p.unit.EnableBuildPhaseStatDep(sim, dep)
		} else {
			p.unit.DisableBuildPhaseStatDep(sim, dep)
		}
	}
}

// A multiplier on the equipment share of a stat, which is what the game files' base-resistance
// modifier states. The character keeps one multiplier per stat, so the value cannot follow stacks,
// and the static path has no Simulation for a later Refresh to hand over.
//
// Expiry undoes the multiplier by dividing by it, so a row that states -100% or worse has no way
// back and is reported rather than applied. A pet wears no equipment, so there is no share to scale.
func (p *parser) equipScaling(stat stats.Stat, mult float64) *attachment {
	if p.character == nil || p.unit.Type == core.PetUnit || p.stacking || (p.static && p.conditional) || mult <= 0 {
		return nil
	}

	return multiplier("equip-scaling "+stat.StatName(), mult, func(sim *core.Simulation, factor float64) {
		if sim == nil {
			p.character.ApplyEquipScaling(stat, factor)
		} else {
			p.character.ApplyDynamicEquipScaling(sim, stat, factor)
		}
	})
}

// A pseudo-stat the sim multiplies rather than adds. A stack multiplies again, which the parser
// refuses to assume: a stacking row is skipped unless the caller says the value does not follow the
// stacks.
//
// Expiry divides the multiplier back out, so a row that states -100% or worse leaves the field at
// zero or at an infinity and is reported rather than applied.
func (p *parser) pseudoMultiplier(kind string, key string, fields []*float64, mult float64) *attachment {
	if p.stacking || mult <= 0 {
		return nil
	}

	a := multiplier(kind, mult, func(_ *core.Simulation, factor float64) {
		for _, field := range fields {
			*field *= factor
		}
	})
	a.key = key
	return a
}

// A pseudo-stat the sim adds to, which follows the stacks the way a flat stat does.
func (p *parser) pseudoAdd(kind string, key string, fields []*float64, v float64) *attachment {
	a := additive(kind, v, func(_ *core.Simulation, delta float64) {
		for _, field := range fields {
			*field += delta
		}
	})
	a.key = key
	return a
}

// Weapon skill, which the attack tables read when they are built, so a change rebuilds them.
func (p *parser) weaponSkill(skill proto.WeaponSkill, v float64) *attachment {
	a := p.pseudoAdd("weapon-skill", fmt.Sprintf("WeaponSkills%d", skill), []*float64{&p.unit.PseudoStats.WeaponSkills[skill]}, v)
	set := a.set
	character := p.character
	a.set = func(sim *core.Simulation, level float64) {
		set(sim, level)
		character.UpdateAttackTableWeapons()
	}
	return a
}

// A pseudo-stat the mana bar's regen reads, which the unit recomputes after every change.
func (p *parser) refreshingManaRegen(a *attachment) *attachment {
	if a == nil {
		return nil
	}
	set := a.set
	a.set = func(sim *core.Simulation, level float64) {
		set(sim, level)
		p.unit.UpdateManaRegenRates(sim)
	}
	return a
}

// One of the speeds the unit recomputes on every change. They take the Simulation the aura's gain
// hands over, so the static path skips them; a stack multiplies the speed again, which the parser
// does not assume.
func (p *parser) speed(kind string, key string, apply func(*core.Unit, *core.Simulation, float64), mult float64) *attachment {
	if p.static || p.stacking {
		return nil
	}

	a := multiplier(kind, mult, func(sim *core.Simulation, factor float64) {
		apply(p.unit, sim, factor)
	})
	a.key = key
	a.needsSim = true
	return a
}

// A value added once per level: the attachment keeps what it has handed out and passes on the
// difference.
func additive(kind string, v float64, apply func(sim *core.Simulation, delta float64)) *attachment {
	current := 0.0
	return &attachment{kind: kind, value: v, set: func(sim *core.Simulation, level float64) {
		delta := v*level - current
		if delta == 0 {
			return
		}
		current = v * level
		apply(sim, delta)
	}}
}

// A multiplier applied while the level is up and divided back out when it drops. Only a row that does
// not stack reaches one, so the level is 0 or 1.
func multiplier(kind string, mult float64, apply func(sim *core.Simulation, factor float64)) *attachment {
	active := false
	return &attachment{kind: kind, value: mult, multiplicative: true, set: func(sim *core.Simulation, level float64) {
		if (level > 0) == active {
			return
		}
		active = level > 0

		factor := mult
		if !active {
			factor = 1 / mult
		}
		apply(sim, factor)
	}}
}

// The attachment, or each of its parts, holds only while a weapon meeting the requirement is equipped,
// as well as while the parse is on, and follows item swaps. The row's caller checks that there is a
// character to wear the weapon.
func (p *parser) whileEquipped(req core.EquipRequirement, a *attachment) *attachment {
	if p.dry {
		return a
	}

	gated := a.parts
	if len(gated) == 0 {
		gated = []*attachment{a}
	}
	for _, g := range gated {
		g.set = p.character.NewEquipGate(req, g.set).SetLevel
	}
	return a
}

// The damage multipliers the sim keeps once for everything and once per school: the game files'
// mask of every school is the first, and a narrower mask is the second, once per school bit it
// names.
func (p *parser) schoolMultiplier(kind string, key string, mask int32, mult float64, all *float64, perSchool *[stats.SchoolLen]float64) *attachment {
	if mask == miscAllSchools {
		return p.pseudoMultiplier(kind, key, []*float64{all}, mult)
	}

	var fields []*float64
	for _, index := range schoolIndexes(mask) {
		fields = append(fields, &perSchool[index])
	}
	if len(fields) == 0 {
		return nil
	}
	return p.pseudoMultiplier(kind+"-by-school", "School"+key, fields, mult)
}

// The stats a game files stat index names: -1 is all five, and 0 through 4 are one of them.
func clientStatList(misc int32) []stats.Stat {
	if misc == miscAllStats {
		return clientStats[:]
	}
	if misc < 0 || misc >= int32(len(clientStats)) {
		return nil
	}
	return []stats.Stat{clientStats[misc]}
}

// The stats a game files stat mask names, one bit per stat index.
func statMaskList(mask int32) []stats.Stat {
	var statList []stats.Stat
	for i, stat := range clientStats {
		if mask&(1<<i) != 0 {
			statList = append(statList, stat)
		}
	}
	return statList
}

// The stats an A_MOD_TOTAL_STAT_PERCENTAGE effect multiplies. A mask in the second misc value, where
// there is one, names them, and the first then does not: Heart of the Wild 17003 states 0 and 8,
// Intellect, and Vitality 29140 states 2 and 5, Strength and Stamina.
func (e *Effect) TotalStatPercentStats() []stats.Stat {
	if e.Misc2 != 0 {
		return statMaskList(e.Misc2)
	}
	return clientStatList(e.Misc)
}

// The stats a flat damage mask names: the sim keeps one stat for physical damage and one for magic,
// plus a stat per magic school, so a mask of every school is the first two together, a mask of every
// magic school is the second alone, and any other mask names a stat for each school it states: the
// Spellstones' 36 is fire and shadow damage. The item generator reads the aura's stats by the same rule.
func DamageDoneStats(mask int32) []stats.Stat {
	switch mask {
	case miscAllSchools:
		return []stats.Stat{stats.PhysicalDamage, stats.SpellDamage}
	case miscMagicSchool:
		return []stats.Stat{stats.SpellDamage}
	}

	var out []stats.Stat
	if mask&int32(core.SpellSchoolPhysical) != 0 {
		out = append(out, stats.PhysicalDamage)
	}
	for bit := int32(core.SpellSchoolHoly); bit <= int32(core.SpellSchoolArcane); bit <<= 1 {
		if mask&bit != 0 {
			out = append(out, core.SpellSchool(bit).SchoolDamage())
		}
	}
	return out
}

// The stats a resistance mask names: school 1 is armor, and holy has no resistance stat to take.
func resistanceStats(mask int32) []stats.Stat {
	var out []stats.Stat
	if mask&miscArmor != 0 {
		out = append(out, stats.Armor)
	}

	for bit := int32(core.SpellSchoolHoly); bit <= int32(core.SpellSchoolArcane); bit <<= 1 {
		if mask&bit == 0 {
			continue
		}
		if stat, ok := core.SpellSchool(bit).ResistanceStat(); ok {
			out = append(out, stat)
		}
	}
	return out
}

// The stats an A_MOD_ARMOR_NO_MODS mask names: ArmorNoMods for school 1 alone, nothing for any other
// mask. The item generator reads the aura's armor by the same rule.
func ArmorNoModsStats(mask int32) []stats.Stat {
	if mask == miscArmor {
		return []stats.Stat{stats.ArmorNoMods}
	}
	return nil
}

// Whether the effect states armor alone with A_MOD_ARMOR_NO_MODS.
func (e *Effect) ArmorNoMods() bool {
	return e.Aura == dbcenums.A_MOD_ARMOR_NO_MODS && len(ArmorNoModsStats(e.Misc)) > 0
}

// The ratings a combat rating mask names, read bit by bit, so a mask naming several ratings gives each
// of them. One rating feeds melee, ranged and spells alike, so its melee and spell bits name the same
// stat. The ranged bits are not read: beside the melee or spell bit of the same rating they add
// nothing, and a mask of them alone is a ranged row, which ScopedRatingStats reads.
func RatingStats(mask int32) []stats.Stat {
	var out []stats.Stat
	for _, rating := range []struct {
		bits int32
		stat stats.Stat
	}{
		{dbcenums.RATING_MOD_DEFENSE, stats.DefenseRating},
		{dbcenums.RATING_MOD_DODGE, stats.DodgeRating},
		{dbcenums.RATING_MOD_PARRY, stats.ParryRating},
		{dbcenums.RATING_MOD_BLOCK, stats.BlockRating},
		{dbcenums.RATING_MOD_HIT_MELEE | dbcenums.RATING_MOD_HIT_SPELL, stats.HitRating},
		{dbcenums.RATING_MOD_CRIT_MELEE | dbcenums.RATING_MOD_CRIT_SPELL, stats.CritRating},
		{dbcenums.RATING_MOD_HASTE_MELEE | dbcenums.RATING_MOD_HASTE_SPELL, stats.HasteRating},
		{dbcenums.RATING_MOD_EXPERTISE, stats.ExpertiseRating},
	} {
		if mask&rating.bits != 0 {
			out = append(out, rating.stat)
		}
	}
	return out
}

// The school bits a mask names, as the positions the sim's per-school arrays index by.
func schoolIndexes(mask int32) []stats.SchoolIndex {
	var out []stats.SchoolIndex
	for bit := int32(1); bit <= int32(core.SpellSchoolArcane); bit <<= 1 {
		if mask&bit != 0 {
			out = append(out, core.SpellSchool(bit).SchoolIndex())
		}
	}
	return out
}

func isOneSchool(mask int32) bool {
	return mask > 0 && mask&(mask-1) == 0 && mask <= int32(core.SpellSchoolArcane)
}

func statNames(sts []stats.Stat) string {
	out := ""
	for i, stat := range sts {
		if i > 0 {
			out += "+"
		}
		out += stat.StatName()
	}
	return out
}

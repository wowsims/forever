package core

import (
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

type sheetAttackStats struct {
	HitPercent, CritPercent, HitRating, CritRating float64
	CritSuppressionPercent                         float64
}

// The hit and crit the character sheet shows for the auto attack with this proc mask: the physical
// percent, the ranged percent for a ranged probe, and the hit or crit mods that reach the attack. The
// dual-wield miss penalty is not in them.
//
// The rating is the one hit or crit rating, plus the ranged percent and the mods converted from a
// rating, each at the rating per percent. A mod stated as a percent is in the percent alone.
func (character *Character) sheetAttackStats(probe ProcMask) sheetAttackStats {
	var channelHit, channelCrit float64
	if probe.Matches(ProcMaskRanged) {
		channelHit, channelCrit = character.GetStat(stats.RangedHitPercent), character.GetStat(stats.RangedCritPercent)
	}
	scoped := character.ScopedPercents(probe)
	attack := sheetAttackStats{
		HitPercent:  character.GetStat(stats.PhysicalHitPercent) + channelHit + scoped.Hit,
		CritPercent: character.GetStat(stats.PhysicalCritPercent) + channelCrit + scoped.Crit,
		HitRating:   character.GetStat(stats.HitRating) + (channelHit+scoped.HitFromRating)*HitRatingPerHitPercent,
		CritRating:  character.GetStat(stats.CritRating) + (channelCrit+scoped.CritFromRating)*CritRatingPerCritPercent,
	}
	attack.CritSuppressionPercent = character.sheetCritSuppressionPercent(probe, scoped, attack.CritPercent)
	return attack
}

// The crit the sim takes off the auto attack against the current target, in percent. The sheet's
// crit percents keep it in; before the attack tables exist it is 0.
func (character *Character) sheetCritSuppressionPercent(probe ProcMask, scoped ScopedPercents, critPercent float64) float64 {
	table := character.sheetAttackTable()
	if table == nil {
		return 0
	}
	probeSpell := &Spell{Unit: &character.Unit, ProcMask: probe, BonusCritPercent: scoped.Crit}
	return probeSpell.physicalCritSuppression(table, critPercent) * 100
}

// The attack table against the current target, nil before the attack tables exist.
func (character *Character) sheetAttackTable() *AttackTable {
	if character.CurrentTarget == nil || int(character.CurrentTarget.UnitIndex) >= len(character.AttackTables) {
		return nil
	}
	return character.AttackTables[character.CurrentTarget.UnitIndex]
}

// An agent that moves the melee crit cap the character sheet shows. Suppression and the cap are worked out
// after the hook, the cap from the info fields, so a change to any of them reaches it.
type MeleeCritCapModifier interface {
	ModifyMeleeCritCap(info *proto.MeleeCritCapInfo)
}

// The melee crit cap of the main-hand auto against the current target; nil before the attack tables exist and
// without a main-hand auto. sheetCritPercent is the melee crit the sheet shows.
func (character *Character) meleeCritCapProto(sheetCritPercent float64) *proto.MeleeCritCapInfo {
	table, auto := character.sheetAttackTable(), character.AutoAttacks.MHAuto()
	if table == nil || auto == nil {
		return nil
	}
	white := auto.whiteHitTable(table.Defender, table)
	info := &proto.MeleeCritCapInfo{
		Glancing:           white.glance * 100,
		RemainingHit:       white.miss * 100,
		RemainingExpertise: (white.dodge + white.parry) * 100,
		Block:              white.block * 100,
	}
	if modifier, ok := character.Env.GetAgentFromUnit(&character.Unit).(MeleeCritCapModifier); ok {
		modifier.ModifyMeleeCritCap(info)
	}
	info.Suppression = sheetCritPercent + info.SpecOffset - auto.physicalCritChanceUnclamped(table, info.SpecOffset)*100
	info.CritCap = 100 - info.Glancing - info.RemainingHit - info.RemainingExpertise - info.Block - info.SpecOffset + info.Suppression
	info.Delta = sheetCritPercent - info.CritCap
	return info
}

// The hit the active hit mods with a ShowAsSchoolHit give each of those schools. A talent such as Arcane
// Focus states its hit on its spells' class flags, which name no school the sheet could show it under.
func (character *Character) schoolHitMods() [stats.SchoolLen]float64 {
	var hit [stats.SchoolLen]float64
	for _, mod := range character.spellMods {
		if !mod.IsActive || mod.ShowAsSchoolHit == 0 {
			continue
		}
		for _, school := range singleSchools {
			if mod.ShowAsSchoolHit.Matches(school) {
				hit[school.SchoolIndex()] += mod.floatValue
			}
		}
	}
	return hit
}

// Which of block and parry the character sheet shows: a character that cannot block or parry shows
// no Block% or Parry%, and no change in its stats moves either.
type SheetAvoidance struct{ CanBlock, CanParry bool }

func (sheet SheetAvoidance) Gate(pseudoStats []float64) {
	if !sheet.CanBlock {
		pseudoStats[proto.PseudoStat_PseudoStatBlockPercent] = 0
	}
	if !sheet.CanParry {
		pseudoStats[proto.PseudoStat_PseudoStatParryPercent] = 0
	}
}

func (character *Character) sheetAvoidance() SheetAvoidance {
	return SheetAvoidance{CanBlock: character.PseudoStats.CanBlock, CanParry: character.PseudoStats.CanParry}
}

func (character *Character) GetPseudoStatsProto() []float64 {
	mainHand, ranged := character.sheetAttackStats(ProcMaskMeleeMHAuto), character.sheetAttackStats(ProcMaskRangedAuto)
	var offHand sheetAttackStats
	if character.AutoAttacks.IsDualWielding {
		offHand = character.sheetAttackStats(ProcMaskMeleeOHAuto)
	}
	pseudoStats := []float64{
		proto.PseudoStat_PseudoStatMainHandDps: character.AutoAttacks.MH().DPS(),
		proto.PseudoStat_PseudoStatOffHandDps:  character.AutoAttacks.OH().DPS(),
		proto.PseudoStat_PseudoStatRangedDps:   character.AutoAttacks.Ranged().DPS(),

		// Avoidance against an attacker of the character's own level, whose attack table adds 5% parry and 5% block.
		proto.PseudoStat_PseudoStatDodgePercent:            (character.PseudoStats.BaseDodgeChance + character.GetDodgeFromRating()) * 100,
		proto.PseudoStat_PseudoStatParryPercent:            (TernaryFloat64(character.PseudoStats.CanParry, 0.05, 0) + character.PseudoStats.BaseParryChance + character.GetParryFromRating()) * 100,
		proto.PseudoStat_PseudoStatBlockPercent:            (TernaryFloat64(character.PseudoStats.CanBlock, 0.05, 0) + character.PseudoStats.BaseBlockChance + character.GetBlockFromRating()) * 100,
		proto.PseudoStat_PseudoStatBlockValueMultiplier:    character.PseudoStats.BlockValueMultiplier,
		proto.PseudoStat_PseudoStatBlockValuePerStrength:   character.PseudoStats.BlockValuePerStrength,
		proto.PseudoStat_PseudoStatReducedCritTakenPercent: character.PseudoStats.ReducedCritTakenPercent * 100,

		// Used by UI to incorporate multiplicative Haste buffs into final character stats display.
		proto.PseudoStat_PseudoStatRangedSpeedMultiplier: character.PseudoStats.RangedSpeedMultiplier * character.PseudoStats.AttackSpeedMultiplier,
		proto.PseudoStat_PseudoStatMeleeSpeedMultiplier:  character.PseudoStats.MeleeSpeedMultiplier * character.PseudoStats.AttackSpeedMultiplier,
		proto.PseudoStat_PseudoStatCastSpeedMultiplier:   character.PseudoStats.CastSpeedMultiplier,
		proto.PseudoStat_PseudoStatMeleeHastePercent:     (character.TotalMeleeHasteMultiplier() - 1) * 100,
		proto.PseudoStat_PseudoStatRangedHastePercent:    (character.TotalRangedHasteMultiplier() - 1) * 100,
		proto.PseudoStat_PseudoStatSpellHastePercent:     (character.TotalSpellHasteMultiplier() - 1) * 100,

		// School-specific fully buffed Hit/Crit are represented as proper Stats in the back-end so
		// that stat dependencies will work correctly, but are stored as PseudoStats in proto
		// messages. This is done so that the stats arrays embedded in database files and saved
		// Encounter settings can omit these extraneous fields.
		proto.PseudoStat_PseudoStatSpellHitPercent:  character.GetStat(stats.SpellHitPercent),
		proto.PseudoStat_PseudoStatSpellCritPercent: character.GetStat(stats.SpellCritPercent),
		proto.PseudoStat_PseudoStatExpertisePercent: character.GetStat(stats.ExpertisePercent),

		proto.PseudoStat_PseudoStatMeleeHitPercent:    mainHand.HitPercent,
		proto.PseudoStat_PseudoStatMeleeCritPercent:   mainHand.CritPercent,
		proto.PseudoStat_PseudoStatMainHandHitRating:  mainHand.HitRating,
		proto.PseudoStat_PseudoStatMainHandCritRating: mainHand.CritRating,
		proto.PseudoStat_PseudoStatOffHandHitPercent:  offHand.HitPercent,
		proto.PseudoStat_PseudoStatOffHandCritPercent: offHand.CritPercent,
		proto.PseudoStat_PseudoStatOffHandHitRating:   offHand.HitRating,
		proto.PseudoStat_PseudoStatOffHandCritRating:  offHand.CritRating,
		proto.PseudoStat_PseudoStatRangedHitPercent:   ranged.HitPercent,
		proto.PseudoStat_PseudoStatRangedCritPercent:  ranged.CritPercent,
		proto.PseudoStat_PseudoStatRangedHitRating:    ranged.HitRating,
		proto.PseudoStat_PseudoStatRangedCritRating:   ranged.CritRating,

		proto.PseudoStat_PseudoStatRangedCritSuppressionPercent: ranged.CritSuppressionPercent,
	}
	modHit := character.schoolHitMods()
	for school := stats.SchoolIndexArcane; school < stats.SchoolLen; school++ {
		pseudoStats[SchoolHitPseudoStats[school]] = character.GetStat(stats.SpellHitPercent) + character.PseudoStats.SchoolBonusHitChance[school] + modHit[school]
	}
	character.sheetAvoidance().Gate(pseudoStats)
	return pseudoStats
}

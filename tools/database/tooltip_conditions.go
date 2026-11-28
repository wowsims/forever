package database

import (
	"database/sql"
	"fmt"
	"maps"
	"math/bits"
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/tools/database/dbc"
	"github.com/wowsims/forever/tools/tooltip"
)

// The spells the classes' spec skill lines teach (SkillLine category 7 with flag 0x400; Engraving, Beast Training
// and the pets' lines have it clear), with each row's classes, the rank the spell replaces and the level it is
// learned at.
const classSpellsQuery = `
	SELECT sla.Spell, sla.ClassMask, sla.SupercedesSpell, COALESCE((SELECT MIN(lv.SpellLevel) FROM SpellLevels lv WHERE lv.SpellID = sla.Spell), 0)
	FROM SkillLineAbility sla
	JOIN SkillLine line ON line.ID = sla.SkillLine
	WHERE line.CategoryID = 7 AND line.Flags & 1024 != 0`

// What the tooltips' conditions on spells, auras and passives read off the player, beside the talent ranks:
//   - a spell a class skill line teaches by the character's level is known to every player of the classes its rows
//     name, and a later rank of a talent's spell to a player with the talent;
//   - an aura that shifts the player into a form or stance is up in it, a class's passive bound to forms is up in
//     them, and a class's other passives are always up.
//
// Any other spell or aura, a rune, a set bonus, a buff the sim does not model, reads false.
func tooltipConditions(helper *DBHelper, instance *dbc.DBC, treeTalents map[int32]int32) (*tooltip.Conditions, error) {
	type row struct {
		spell, classMask, supersedes, level int32
	}
	rows, err := LoadRows(helper.db, classSpellsQuery, func(rows *sql.Rows) (row, error) {
		var r row
		return r, rows.Scan(&r.spell, &r.classMask, &r.supersedes, &r.level)
	})
	if err != nil {
		return nil, fmt.Errorf("loading the class spells for tooltips: %w", err)
	}
	classMasks := map[int32]int32{}
	previous := map[int32]int32{}
	for _, r := range rows {
		if r.level <= core.CharacterLevel {
			classMasks[r.spell] |= r.classMask
		}
		if r.supersedes != 0 {
			previous[r.spell] = r.supersedes
		}
	}

	conditions := &tooltip.Conditions{Known: map[int64]tooltip.Expr{}, Auras: map[int64]tooltip.Expr{}}
	for _, spell := range slices.Sorted(maps.Keys(classMasks)) {
		classes := classesOf(classMasks[spell])
		if len(classes) == 0 {
			continue
		}
		known := tooltip.PlayerOfClass(classes...)
		talent, fromTalent := talentRoot(spell, previous, treeTalents)
		if fromTalent {
			known = tooltip.TalentTaken(int64(talent))
		}
		conditions.Known[int64(spell)] = known

		row := instance.Spells[int(spell)]
		if !row.HasAttributeAt(dbcenums.ATTR_INDEX_BASE, dbcenums.ATTR_PASSIVE) {
			continue
		}
		if forms := formsOf(row.ShapeshiftMask); len(forms) > 0 && !fromTalent {
			conditions.Auras[int64(spell)] = tooltip.PlayerInForm(forms...)
		} else {
			conditions.Auras[int64(spell)] = known
		}
	}
	for _, id := range slices.Sorted(maps.Keys(instance.Spells)) {
		for _, effect := range instance.SpellEffectsInOrder(id) {
			if effect.EffectAura == dbcenums.A_MOD_SHAPESHIFT && len(effect.EffectMiscValues) > 0 && effect.EffectMiscValues[0] != 0 {
				conditions.Auras[int64(id)] = tooltip.PlayerInForm(int32(effect.EffectMiscValues[0]))
			}
		}
	}
	return conditions, nil
}

// The class ids of a ClassMask, whose bit class-1 names each.
func classesOf(mask int32) []int32 {
	var classes []int32
	for m := uint32(mask); m != 0; m &= m - 1 {
		classes = append(classes, int32(bits.TrailingZeros32(m))+1)
	}
	return classes
}

// The forms of a ShapeshiftMask pair, whose bit form-1 names each.
func formsOf(mask []int) []int32 {
	var forms []int32
	for word, value := range mask {
		for m := uint32(value); m != 0; m &= m - 1 {
			forms = append(forms, int32(word*32+bits.TrailingZeros32(m))+1)
		}
	}
	return forms
}

// The tree talent whose spell a later rank goes back to, rank by rank.
func talentRoot(spell int32, previous map[int32]int32, treeTalents map[int32]int32) (int32, bool) {
	for range 20 {
		if _, ok := treeTalents[spell]; ok {
			return spell, true
		}
		next, ok := previous[spell]
		if !ok {
			return 0, false
		}
		spell = next
	}
	return 0, false
}

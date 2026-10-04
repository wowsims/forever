package spelldata

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/stats"
)

func TestWeaponScopeRules(t *testing.T) {
	enemyAura := func(aura dbcenums.EffectAuraType) Effect {
		return Effect{Type: dbcenums.E_APPLY_AURA, Aura: aura,
			Target: [2]dbcenums.ImplicitTarget{dbcenums.TARGET_UNIT_TARGET_ENEMY}}
	}
	rating := func(mask int32) Effect {
		e := casterAura(dbcenums.A_MOD_RATING, 10)
		e.Misc = mask
		return e
	}
	offHand := [17]uint32{3: dbcenums.ATTR_EX_3_REQ_OFF_HAND}

	cases := []struct {
		name  string
		row   Spell
		e     Effect
		rule  WeaponRule
		scope core.SpellModConfig
	}{
		{"weapon crit with melee weapons", Spell{EquipClass: 2, EquipSubclass: 173555},
			casterAura(dbcenums.A_MOD_WEAPON_CRIT_PERCENT, 5), WeaponRulePerWeapon,
			core.SpellModConfig{ProcMask: core.ProcMaskMelee, EquipSubclass: 173555}},
		{"hit with one-handers in the off hand", Spell{EquipClass: 2, EquipSubclass: 41105, Attr: offHand},
			casterAura(dbcenums.A_MOD_HIT_CHANCE, 10), WeaponRulePerWeapon,
			core.SpellModConfig{ProcMask: core.ProcMaskMelee, EquipSubclass: 41105, EquipHand: core.ProcMaskMeleeOH}},
		{"crit on everything with melee weapons", Spell{EquipClass: 2, EquipSubclass: 173555},
			casterAura(dbcenums.A_MOD_CRIT_PCT, 1), WeaponRuleWhileEquipped,
			core.SpellModConfig{ProcMask: core.ProcMaskMelee, EquipSubclass: 173555}},
		{"crit on everything with ranged weapons", Spell{EquipClass: 2, EquipSubclass: 262156},
			casterAura(dbcenums.A_MOD_CRIT_PCT, 2), WeaponRulePerWeapon,
			core.SpellModConfig{ProcMask: core.ProcMaskRanged, EquipSubclass: 262156}},
		{"crit on everything with a ranged inventory type", Spell{EquipClass: 2, EquipSubclass: 262156, EquipInvType: dbcenums.INVTYPE_BIT_RANGED},
			casterAura(dbcenums.A_MOD_CRIT_PCT, 2), WeaponRuleWhileEquipped,
			core.SpellModConfig{ProcMask: core.ProcMaskRanged, EquipSubclass: 262156, EquipInvType: dbcenums.INVTYPE_BIT_RANGED}},
		{"hit taken off an enemy", Spell{EquipClass: 2, EquipSubclass: 262156},
			enemyAura(dbcenums.A_MOD_HIT_CHANCE), WeaponRuleNone, core.SpellModConfig{}},
		{"hit with a shield", Spell{EquipClass: 4, EquipSubclass: 64},
			casterAura(dbcenums.A_MOD_HIT_CHANCE, 1), WeaponRuleNone, core.SpellModConfig{}},
		{"spell hit with melee weapons", Spell{EquipClass: 2, EquipSubclass: 173555},
			casterAura(dbcenums.A_MOD_SPELL_HIT_CHANCE, 1), WeaponRuleNone, core.SpellModConfig{}},
		{"hit rating with melee weapons", Spell{EquipClass: 2, EquipSubclass: 173555},
			rating(32 | 64 | 128), WeaponRulePerWeapon,
			core.SpellModConfig{ProcMask: core.ProcMaskMelee, EquipSubclass: 173555}},
		{"ranged hit rating", Spell{}, rating(64), WeaponRuleRangedRatings,
			core.SpellModConfig{ProcMask: core.ProcMaskRanged, DefenseType: core.DefenseTypeRanged}},
		{"hit rating on every channel", Spell{}, rating(32 | 64 | 128), WeaponRuleNone, core.SpellModConfig{}},
		{"expertise rating", Spell{}, rating(8388608), WeaponRuleNone, core.SpellModConfig{}},
	}
	for _, c := range cases {
		rule, scope := WeaponScope(&c.row, &c.e)
		if rule != c.rule || scope.ProcMask != c.scope.ProcMask || scope.EquipSubclass != c.scope.EquipSubclass ||
			scope.EquipInvType != c.scope.EquipInvType || scope.EquipHand != c.scope.EquipHand {
			t.Errorf("%s: rule %d scope %+v, want rule %d scope %+v", c.name, rule, scope, c.rule, c.scope)
		}
	}
}

func TestScopedRatingStatsReadsTheRangedBits(t *testing.T) {
	for mask, want := range map[int32][]stats.Stat{
		64:                {stats.HitRating},
		512 | 262144:      {stats.CritRating, stats.HasteRating},
		32 | 8388608:      {stats.HitRating, stats.ExpertiseRating},
		64 | 512 | 262144: {stats.HitRating, stats.CritRating, stats.HasteRating},
	} {
		if got := ScopedRatingStats(mask); !slices.Equal(got, want) {
			t.Errorf("mask %d reads %v, want %v", mask, got, want)
		}
	}
	if got := RatingStats(64); len(got) != 0 {
		t.Errorf("the unscoped read of a ranged bit alone gives %v, want nothing", got)
	}
}

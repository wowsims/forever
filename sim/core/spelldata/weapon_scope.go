package spelldata

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/stats"
)

// How the parse binds a hit, crit or rating aura to the weapon its row requires in
// SpellEquippedItems.
type WeaponRule uint8

const (
	// The effect gives its stat: the row names no weapon, names a shield, the aura reaches another unit,
	// or the effect is not a hit, crit, rating or damage done aura.
	WeaponRuleNone WeaponRule = iota
	// A mod on the attacks made with a matching weapon.
	WeaponRulePerWeapon
	// The stat on every spell and attack while a matching weapon is in a weapon slot.
	WeaponRuleWhileEquipped
	// A rating mask of ranged bits alone, on a row that names no weapon: the percent the rating is
	// worth, on ranged-defense attacks. A magic-defense spell on the ranged proc mask, such as Thunder
	// Clap, rolls on the spell tables and takes none of it.
	WeaponRuleRangedRatings
)

const rangedInvTypes int32 = dbcenums.INVTYPE_BIT_RANGED | dbcenums.INVTYPE_BIT_THROWN | dbcenums.INVTYPE_BIT_RANGEDRIGHT

// The rule the parse applies to a hit, crit, rating or damage done effect, and the scope a mod under it
// takes: the proc mask and the Equip fields, with no class flags, so the mod reaches auto attacks
// whatever family the spell belongs to. The scope is zero under WeaponRuleNone.
//
// A_MOD_HIT_CHANCE and A_MOD_WEAPON_CRIT_PERCENT are per weapon. A_MOD_CRIT_PCT and
// A_MOD_DAMAGE_PERCENT_DONE are per weapon where their weapons are ranged weapons alone, named by
// subclass only, and hold while equipped otherwise. A_MOD_RATING is per weapon where the row names
// one, and on ranged attacks where its mask holds ranged bits alone.
func WeaponScope(s *Spell, e *Effect) (WeaponRule, core.SpellModConfig) {
	if !WeaponBindable(e.Aura) {
		return WeaponRuleNone, core.SpellModConfig{}
	}

	scope, scoped := weaponScope(s, e)
	switch {
	case scoped && (e.Aura == dbcenums.A_MOD_CRIT_PCT || e.Aura == dbcenums.A_MOD_DAMAGE_PERCENT_DONE) &&
		(s.EquipSubclass == 0 || s.EquipSubclass&^dbcenums.ITEM_SUBCLASS_MASK_WEAPON_RANGED != 0 || scope.EquipInvType != 0 || scope.EquipHand != 0):
		return WeaponRuleWhileEquipped, scope
	case scoped:
		return WeaponRulePerWeapon, scope
	case e.Aura == dbcenums.A_MOD_RATING && RangedBitsOnly(e.Misc):
		return WeaponRuleRangedRatings, core.SpellModConfig{ProcMask: core.ProcMaskRanged, DefenseType: core.DefenseTypeRanged}
	}
	return WeaponRuleNone, core.SpellModConfig{}
}

// Whether the aura is a hit, crit, rating or damage done aura, the ones WeaponScope can bind to a
// weapon.
func WeaponBindable(aura dbcenums.EffectAuraType) bool {
	switch aura {
	case dbcenums.A_MOD_HIT_CHANCE, dbcenums.A_MOD_WEAPON_CRIT_PERCENT, dbcenums.A_MOD_CRIT_PCT, dbcenums.A_MOD_RATING,
		dbcenums.A_MOD_DAMAGE_PERCENT_DONE:
		return true
	}
	return false
}

// Whether a combat rating mask holds ranged bits and no other.
func RangedBitsOnly(mask int32) bool {
	return mask&dbcenums.RATING_MOD_MASK_RANGED != 0 && mask&^dbcenums.RATING_MOD_MASK_RANGED == 0
}

// The ratings a scoped A_MOD_RATING mask names, its ranged bits read as the rating they belong to. The
// aura parse turns the hit and crit among them into mods, and the item fold turns the hit, crit and
// haste of a mask of ranged bits alone into the ranged percents.
func ScopedRatingStats(mask int32) []stats.Stat {
	return RatingStats(mask | (mask&dbcenums.RATING_MOD_MASK_RANGED)>>1)
}

type RatingPercent struct {
	Stat    stats.Stat
	Percent float64
}

// The ratings ScopedRatingStats reads, each with the percent v points of it are worth. A rating with
// no percent of its own is worth 0.
func ScopedRatingPercents(mask int32, v float64) []RatingPercent {
	ratings := ScopedRatingStats(mask)
	out := make([]RatingPercent, 0, len(ratings))
	for _, stat := range ratings {
		percent := 0.0
		if channels, ok := core.RatingChannelsOf(stat); ok {
			percent = v / channels.RatingPerPercent
		}
		out = append(out, RatingPercent{stat, percent})
	}
	return out
}

// The hand ATTR_EX_3_REQ_MAIN_HAND and ATTR_EX_3_REQ_OFF_HAND name, as a melee proc mask.
func RequiredHand(s *Spell) core.ProcMask {
	var hand core.ProcMask
	if s.HasAttr(dbcenums.ATTR_INDEX_EX_3, dbcenums.ATTR_EX_3_REQ_MAIN_HAND) {
		hand |= core.ProcMaskMeleeMH
	}
	if s.HasAttr(dbcenums.ATTR_INDEX_EX_3, dbcenums.ATTR_EX_3_REQ_OFF_HAND) {
		hand |= core.ProcMaskMeleeOH
	}
	return hand
}

// An aura cast at someone else is not bound to the weapon: the requirement is the caster's, for the
// cast.
func weaponScope(s *Spell, e *Effect) (core.SpellModConfig, bool) {
	if s.EquipClass != dbcenums.ITEM_CLASS_WEAPON || e.Target[0] != dbcenums.TARGET_UNIT_CASTER {
		return core.SpellModConfig{}, false
	}

	scope := core.SpellModConfig{EquipSubclass: s.EquipSubclass, EquipInvType: s.EquipInvType, EquipHand: RequiredHand(s)}
	if scope.Equip().IsZero() {
		return core.SpellModConfig{}, false
	}

	if scope.EquipSubclass&dbcenums.ITEM_SUBCLASS_MASK_WEAPON_RANGED != 0 || scope.EquipInvType&rangedInvTypes != 0 {
		scope.ProcMask |= core.ProcMaskRanged
	}
	if scope.EquipSubclass&^dbcenums.ITEM_SUBCLASS_MASK_WEAPON_RANGED != 0 || scope.EquipInvType&^rangedInvTypes != 0 || scope.EquipHand != 0 {
		scope.ProcMask |= core.ProcMaskMelee
	}
	return scope, true
}

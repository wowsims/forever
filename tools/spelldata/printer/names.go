package printer

import (
	"fmt"
	"math/bits"
	"os"
	"path/filepath"
	"strings"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The name dbcenums gives a value, or its number in format where it gives none.
func namedOr[T fmt.Stringer](v T, format string) string {
	if name, ok := dbcenums.Named(v); ok {
		return name
	}
	return fmt.Sprintf(format, v)
}

// The bits of SpellAuraOptions.ProcTypeMask by name, word 0 then word 1.
func procFlagNames(flags [2]uint32) []string {
	out := setBits(uint64(flags[0]), func(bit uint64) (string, bool) {
		return constName(procFlagConsts, int64(bit))
	})
	return append(out, setBits(uint64(flags[1]), func(bit uint64) (string, bool) {
		return constName(procFlag2Consts, int64(bit))
	})...)
}

// The bits a mask sets, lowest first, each by its name or as `bit 0x…` where it has none.
func setBits(mask uint64, name func(bit uint64) (string, bool)) []string {
	var out []string
	for mask != 0 {
		bit := mask & -mask
		mask &^= bit
		if named, ok := name(bit); ok {
			out = append(out, named)
		} else {
			out = append(out, fmt.Sprintf("bit %#x", bit))
		}
	}
	return out
}

// The repository the tool is reading, found by walking up from the working directory. The extension
// spawns the tool with the folder holding go.mod as its working directory, and `go run` is normally
// typed there too.
func moduleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no go.mod above %s", dir)
		}
		dir = parent
	}
}

// One ImplicitTarget value: what the effect reaches, and whether the value names a place rather than
// a unit. A pair of a place and a unit - Whirlwind's [22,15] - is worded by the unit.
type implicitTarget struct {
	phrase string
	place  bool
}

// The ImplicitTarget values the store's rows carry, under the game files' own numbering. The ones this
// table leaves out read as "target N": naming a value the store carries only a handful of times, on
// no evidence beyond its number, would put a guess where the literal already states the fact.
var implicitTargets = map[dbcenums.ImplicitTarget]implicitTarget{
	1:  {phrase: "to the caster"},
	2:  {phrase: "to a nearby enemy"},
	3:  {phrase: "to a nearby party member"},
	4:  {phrase: "to a nearby ally"},
	5:  {phrase: "to the caster's pet"},
	6:  {phrase: "to the enemy"},
	7:  {phrase: "to every enemy in the area"},
	15: {phrase: "to every enemy around the caster"},
	16: {phrase: "to every enemy in the target area"},
	18: {place: true},
	20: {phrase: "to the party around the caster"},
	21: {phrase: "to the friendly target"},
	22: {place: true},
	24: {phrase: "to a cone of enemies in front of the caster"},
	25: {phrase: "to the target"},
	27: {phrase: "to the caster's master"},
	30: {phrase: "to every ally around the caster"},
	31: {phrase: "to every ally in the target area"},
	33: {phrase: "to the party around the caster"},
	34: {phrase: "to the party in the target area"},
	35: {phrase: "to the target's party"},
	45: {phrase: "to the chain-heal target"},
	47: {place: true},
	53: {place: true},
	55: {place: true},
	56: {phrase: "to the raid around the caster"},
	57: {phrase: "to the target's raid"},
}

// SpellShapeshift.ShapeshiftMask names a form as bit form-1. Only the forms the store's masks
// actually set are named.
var stanceNames = map[int]string{
	1:  "Cat",
	2:  "Tree of Life",
	3:  "Travel",
	5:  "Bear",
	8:  "Dire Bear",
	17: "Battle",
	18: "Defensive",
	19: "Berserker",
	28: "Shadowform",
	30: "Stealth",
	31: "Moonkin",
	32: "Spirit of Redemption",
}

func stanceList(mask uint64) string {
	var out []string
	for form := 1; form <= 64; form++ {
		if mask&(1<<uint(form-1)) != 0 {
			out = append(out, formName(form))
		}
	}
	return strings.Join(out, ", ")
}

func formName(form int) string {
	if name, ok := stanceNames[form]; ok {
		return name
	}
	return fmt.Sprintf("form %d", form)
}

// ItemSubClassWeapon by its bit in the mask SpellEquippedItems states.
var weaponSubclasses = map[int32]string{
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_1H_AXE:       "one-handed axe",
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_2H_AXE:       "two-handed axe",
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_BOW:          "bow",
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_GUNS:         "gun",
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_MACE_1H:      "one-handed mace",
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_MACE_2H:      "two-handed mace",
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_POLEARM:      "polearm",
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_1H:     "one-handed sword",
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_2H:     "two-handed sword",
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_STAFF:        "staff",
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_1H_EXOTIC:    "exotic",
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_FIST:         "fist weapon",
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_MISC:         "miscellaneous weapon",
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_DAGGERS:      "dagger",
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_THROWN:       "thrown",
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SPEAR:        "spear",
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_CROSSBOW:     "crossbow",
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_WAND:         "wand",
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_FISHING_POLE: "fishing pole",
}

var armorSubclasses = map[int]string{
	0: "miscellaneous",
	1: "cloth",
	2: "leather",
	3: "mail",
	4: "plate",
	6: "shield",
	7: "libram",
	8: "idol",
	9: "totem",
}

// InventoryType, which SpellEquippedItems states as a mask of slots.
var inventoryTypes = map[int]string{
	1:  "head",
	2:  "neck",
	3:  "shoulders",
	5:  "chest",
	6:  "waist",
	7:  "legs",
	8:  "feet",
	9:  "wrists",
	10: "hands",
	11: "finger",
	12: "trinket",
	13: "weapon",
	14: "shield",
	15: "ranged",
	16: "cloak",
	17: "two-handed weapon",
	20: "robe",
	21: "main hand",
	22: "off hand",
	23: "held in off hand",
	25: "thrown",
	26: "ranged right",
	28: "relic",
}

// Every melee weapon, a mask a row naming the whole family reads better as.
const meleeWeapons = dbcenums.ITEM_SUBCLASS_BIT_WEAPON_1H_AXE | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_2H_AXE |
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_MACE_1H | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_MACE_2H |
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_POLEARM | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_1H |
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_2H | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_STAFF |
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_FIST | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_DAGGERS |
	dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SPEAR

// What the spell states it needs equipped: the item class and subclass mask a weapon-specific
// ability or proc names, the inventory slots an enchant names, and the hand an ATTR_EX_3 bit names.
func equipRequirement(s *spelldata.Spell) string {
	var item string
	switch s.EquipClass {
	case 0:
		// Nothing stated, which is also what a row carrying only an inventory slot leaves here.
		if s.EquipInvType == 0 {
			return ""
		}
		item = "an item"
	case dbcenums.ITEM_CLASS_WEAPON:
		return "needs " + weaponsPhrase(s.EquipSubclass, s.EquipInvType, spelldata.RequiredHand(s))
	case 4:
		item = armorRequirement(s.EquipSubclass)
	default:
		item = fmt.Sprintf("an item of class %d", s.EquipClass)
	}
	if s.EquipInvType != 0 {
		item += " in the " + maskNames(s.EquipInvType, inventoryTypes)
	}
	return "needs " + item
}

// A weapon requirement in words, from the Equip fields a row or a SpellModConfig states: the weapons,
// then the slots and the hand that hold them.
func weaponsPhrase(subclass, invType int32, hand core.ProcMask) string {
	weapons := "a weapon"
	if subclass != 0 {
		weapons = "a " + weaponRequirement(subclass)
	}

	var slots []string
	if invType != 0 {
		slots = append(slots, maskNames(invType, inventoryTypes))
	}
	if hand.Matches(core.ProcMaskMeleeMH) && invType&dbcenums.INVTYPE_BIT_WEAPONMAINHAND == 0 {
		slots = append(slots, inventoryTypes[21])
	}
	if hand.Matches(core.ProcMaskMeleeOH) && invType&dbcenums.INVTYPE_BIT_WEAPONOFFHAND == 0 {
		slots = append(slots, inventoryTypes[22])
	}
	if len(slots) == 0 {
		return weapons
	}
	return weapons + " in the " + orList(slots)
}

func armorRequirement(mask int32) string {
	if mask == 1<<shieldSubclass {
		return "a shield"
	}
	return maskNames(mask, armorSubclasses)
}

const shieldSubclass = 6

// A weapon mask, summarised where it names a whole family: the melee set and the shooting set are
// what the game files state on a stance ability and on a shot.
func weaponRequirement(mask int32) string {
	switch mask {
	case dbcenums.ITEM_SUBCLASS_MASK_WEAPON_SHOOTING:
		return "bow, gun or crossbow"
	case meleeWeapons:
		return "melee weapon"
	}
	return orList(setBits(uint64(uint32(mask)), func(bit uint64) (string, bool) {
		name, ok := weaponSubclasses[int32(bit)]
		return name, ok
	}))
}

// A mask of bit positions as the names it sets.
func maskNames(mask int32, names map[int]string) string {
	return orList(setBits(uint64(uint32(mask)), func(bit uint64) (string, bool) {
		name, ok := names[bits.TrailingZeros64(bit)]
		return name, ok
	}))
}

func orList(names []string) string {
	switch len(names) {
	case 0:
		return ""
	case 1:
		return names[0]
	}
	return strings.Join(names[:len(names)-1], ", ") + " or " + names[len(names)-1]
}

// SpellClassOptions as a talent reads it: the family the spells are filed under and the mask of
// spells inside it. The words past the last set one are left off.
func classFlagsPhrase(flags core.ClassFlags) string {
	if flags.IsZero() {
		return ""
	}

	mask := maskWords(flags)
	if mask == "" {
		return fmt.Sprintf("family %d", flags.Family)
	}
	return fmt.Sprintf("family %d mask %s", flags.Family, mask)
}

// EffectSpellClassMask_0..3, up to the last word the effect sets: one word is the common case and
// reads as a bare hex, and a mask reaching further states every word up to it in order.
func maskWords(flags core.ClassFlags) string {
	last := -1
	for i, word := range flags.Mask {
		if word != 0 {
			last = i
		}
	}
	if last < 0 {
		return ""
	}

	words := make([]string, 0, last+1)
	for _, word := range flags.Mask[:last+1] {
		words = append(words, fmt.Sprintf("%#x", word))
	}
	return strings.Join(words, ",")
}

// The ProcHint bits the generator baked onto the row, which is what the tooltip says and the mask
// cannot.
func procHintNames(hint core.ProcHint) []string {
	var out []string
	for _, h := range []struct {
		bit  core.ProcHint
		name string
	}{
		{core.ProcHintCastTrigger, "cast trigger"},
		{core.ProcHintCrit, "on a crit"},
		{core.ProcHintHeals, "heals count"},
		{core.ProcHintPureHeal, "heals only"},
		{core.ProcHintNamedAbility, "one named ability"},
		{core.ProcHintOutcomeTaken, "an outcome the mask has no bit for"},
		{core.ProcHintAttackDodged, "own attack dodged"},
		{core.ProcHintAttackParried, "own attack parried"},
		{core.ProcHintDamage, "damaging spells only"},
	} {
		if hint.Matches(h.bit) {
			out = append(out, h.name)
		}
	}
	return out
}

// SpellPower.PowerType, of which the store's rows carry six.
func powerName(t dbcenums.PowerType) string {
	switch t {
	case dbcenums.POWER_HEALTH:
		return "health"
	case dbcenums.POWER_MANA:
		return "mana"
	case dbcenums.POWER_RAGE:
		return "rage"
	case dbcenums.POWER_FOCUS:
		return "focus"
	case dbcenums.POWER_ENERGY:
		return "energy"
	case dbcenums.POWER_COMBO_POINTS:
		return "combo points"
	}
	return fmt.Sprintf("power %d", t)
}

// A_MOD_STAT states the stat in its misc value, and -1 is the game files' "every stat".
func statName(misc int32) string {
	switch misc {
	case -1:
		return "to all stats"
	case 0:
		return "strength"
	case 1:
		return "agility"
	case 2:
		return "stamina"
	case 3:
		return "intellect"
	case 4:
		return "spirit"
	}
	return fmt.Sprintf("stat %d", misc)
}

// A_MOD_TOTAL_STAT_PERCENTAGE states its stats as a mask of stat indexes in its second misc value,
// where it has one; 31 is all five.
func statMaskName(mask int32) string {
	if mask == 31 {
		return statName(-1)
	}
	var names []string
	for i := int32(0); i < 32; i++ {
		if uint32(mask)&(1<<i) != 0 {
			names = append(names, statName(i))
		}
	}
	return strings.Join(names, " and ")
}

// SpellCategories.DispelType, which is also what an E_DISPEL effect names in its misc value.
func dispelName(t int32) string {
	switch t {
	case 1:
		return "magic"
	case 2:
		return "curse"
	case 3:
		return "disease"
	case 4:
		return "poison"
	case 5:
		return "stealth"
	case 6:
		return "invisibility"
	case 8:
		return "enrage"
	}
	return fmt.Sprintf("dispel type %d", t)
}

// The school mask as the schools it holds, since a row can carry more than one bit.
func schoolName(mask core.SpellSchool) string {
	if mask == allSchools {
		return "every school"
	}
	return strings.Join(setBits(uint64(mask), func(bit uint64) (string, bool) {
		name, ok := schoolWords[core.SpellSchool(bit)]
		return name, ok
	}), "+")
}

var schoolWords = map[core.SpellSchool]string{
	core.SpellSchoolPhysical: "physical",
	core.SpellSchoolHoly:     "holy",
	core.SpellSchoolFire:     "fire",
	core.SpellSchoolNature:   "nature",
	core.SpellSchoolFrost:    "frost",
	core.SpellSchoolShadow:   "shadow",
	core.SpellSchoolArcane:   "arcane",
}

// Every school bit, which a damage aura states as one mask rather than as seven.
const allSchools core.SpellSchool = 127

func defenseName(t core.DefenseType) string {
	switch t {
	case core.DefenseTypeMagic:
		return "magic"
	case core.DefenseTypeMelee:
		return "melee"
	case core.DefenseTypeRanged:
		return "ranged"
	}
	return ""
}

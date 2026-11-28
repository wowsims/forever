package dbc

import "github.com/wowsims/forever/sim/core/dbcenums"

type Spell struct {
	NameLang              string
	ID                    int32
	SchoolMask            int32
	Speed                 float32
	LaunchDelay           float32
	MinDuration           float32
	MaxScalingLevel       int
	MinScalingLevel       int32
	ScalesFromItemLevel   int32
	SpellLevel            int
	BaseLevel             int32
	MaxLevel              int
	MaxPassiveAuraLevel   int32
	Cooldown              int32
	GCD                   int32
	MinRange              float32
	MaxRange              float32
	Attributes            []int
	CategoryFlags         int32
	MaxCharges            int32
	ChargeRecoveryTime    int32
	CategoryTypeMask      int32
	Category              int32
	DefenseType           int32 // SpellCategories.DefenseType: 0 none, 1 magic, 2 melee, 3 ranged
	Duration              int32
	ProcChance            float32
	ProcCharges           int32
	ProcTypeMask          []int
	ProcCategoryRecovery  int32
	CategoryRecoveryTime  int32
	EquippedItemClass     int32
	EquippedItemInvTypes  int32
	EquippedItemSubclass  int32
	CastTimeMin           float32
	SpellClassMask        []int
	SpellClassSet         int32
	AuraInterruptFlags    []int
	ChannelInterruptFlags []int
	ShapeshiftMask        []int
	Description           string
	Variables             string
	MaxCumulativeStacks   int32
	MaxTargets            int32
	RequiredAreasID       int32
	IconPath              string
}

func (s *Spell) HasAttributeAt(index int, flag uint32) bool {
	if index < 0 || index >= len(s.Attributes) {
		return false
	}
	return (uint32(s.Attributes[index]) & flag) != 0
}

// Reports whether the spell's effect amounts come from the item level of the item carrying it
// rather than from a class curve. On such an effect EffectBasePoints is a stale snapshot and
// only the scaling coefficient resolves the real amount.
func (s *Spell) ScalesWithItemLevel() bool {
	return s.HasAttributeAt(dbcenums.ATTR_INDEX_EX_11, dbcenums.ATTR_EX_11_SCALES_WITH_ITEM_LEVEL)
}

// Reports whether the listener aura also fires from hits of triggered spells that lack
// ATTR_EX_3_NOT_A_PROC. Emitted as ProcTrigger.CanProcFromProcs.
func (s *Spell) CanProcFromProcs() bool {
	return s.HasAttributeAt(dbcenums.ATTR_INDEX_EX_3, dbcenums.ATTR_EX_3_CAN_PROC_FROM_PROCS)
}

// Reports whether a triggered spell's hits count as ability hits to aura listeners rather than
// as proc hits. A damage proc built from a spell without it gets SpellFlagProc.
func (s *Spell) NotAProc() bool {
	return s.HasAttributeAt(dbcenums.ATTR_INDEX_EX_3, dbcenums.ATTR_EX_3_NOT_A_PROC)
}

// Reports whether weapon procs ignore this spell's hits.
func (s *Spell) SuppressesWeaponProcs() bool {
	return s.HasAttributeAt(dbcenums.ATTR_INDEX_EX_4, dbcenums.ATTR_EX_4_SUPPRESS_WEAPON_PROCS)
}

// Reports whether the listener aura, although not a weapon proc, ignores hits that suppress
// weapon procs. Emitted as SpellFlagsExclude: SpellFlagSuppressWeaponProcs.
func (s *Spell) IsWeaponProcAura() bool {
	return s.HasAttributeAt(dbcenums.ATTR_INDEX_EX_6, dbcenums.ATTR_EX_6_AURA_IS_WEAPON_PROC)
}

// Reports whether the spell is barred from critically striking. Seal of Light is the shape: effects
// that resolve at a flat amount in game. A damage proc built from one of these has to roll an outcome
// that cannot crit.
func (s *Spell) CannotCrit() bool {
	return s.HasAttributeAt(dbcenums.ATTR_INDEX_EX_2, dbcenums.ATTR_EX_2_CANT_CRIT)
}

// Reports whether the spell only procs from class abilities rather than from any hit.
func (s *Spell) OnlyProcsFromClassAbilities() bool {
	return s.HasAttributeAt(dbcenums.ATTR_INDEX_EX_12, dbcenums.ATTR_EX_12_ONLY_PROC_FROM_CLASS_ABILITIES)
}

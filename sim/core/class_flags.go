package core

import "github.com/wowsims/forever/sim/core/proto"

// The game files' SpellClassSet per class: the family every one of that class's spells files its class
// mask under. Verified row by row against the spell store in spelldata's TestClassSpellFamilies.
var ClassSpellFamilies = map[proto.Class]int32{
	proto.Class_ClassMage:    3,
	proto.Class_ClassWarrior: 4,
	proto.Class_ClassWarlock: 5,
	proto.Class_ClassPriest:  6,
	proto.Class_ClassDruid:   7,
	proto.Class_ClassRogue:   8,
	proto.Class_ClassHunter:  9,
	proto.Class_ClassPaladin: 10,
	proto.Class_ClassShaman:  11,
}

// The game files' SpellClassOptions: the family (SpellClassSet) and the four
// 32-bit words of SpellClassMask. A talent effect carries the same shape (EffectSpellClassMask)
// naming the spells it modifies.
type ClassFlags struct {
	Family int32
	Mask   [4]uint32
}

// Whether the spell carries no class options at all. The game files leave both the family and every
// mask word at zero on the spells no talent addresses by family.
func (f ClassFlags) IsZero() bool {
	return f.Family == 0 && f.Mask == [4]uint32{}
}

// Whether the two sets name at least one spell in common. Families are separate namespaces, so
// the same bit means a different spell in each and a cross-family overlap is never a match.
func (f ClassFlags) Matches(o ClassFlags) bool {
	if f.Family != o.Family {
		return false
	}

	for i := range f.Mask {
		if f.Mask[i]&o.Mask[i] != 0 {
			return true
		}
	}

	return false
}

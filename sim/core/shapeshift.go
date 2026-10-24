package core

import (
	"cmp"
	"slices"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
)

// What the game files require of the caster's form, auras and equipment, of the target's creature
// type, and of the ranks of the periodic the cast applies.
type CastRequirement struct {
	Forms, ExcludedForms          uint64           // SpellShapeshift masks, bit form-1
	CasterForm                    bool             // castable with no form even though Forms names some
	NotShapeshifted               bool             // not castable in a shapeshift (a form that is not a stance)
	CasterAura, ExcludeCasterAura int32            // aura spell id that must / must not be active on the caster
	Weapon                        EquipRequirement // SpellEquippedItems of a weapon: each hand named holds a matching weapon, else any weapon slot does
	Shield                        bool             // SpellEquippedItems of armor naming the shield subclass: the off hand holds a shield
	NotInCombat                   bool             // castable only before the encounter starts
	TargetMobTypes                uint32           // mob types the target must be one of, bit 1 << proto.MobType; zero allows any
	NotUnderHigherRank            bool             // a ranked row whose only effect besides a dummy or script effect is the periodic: refused while a higher rank of its dot (DotConfig.RankGroup) from the caster is up on the unit
}

func (r CastRequirement) allowsForm(form dbcenums.ShapeshiftForm) bool {
	bit := form.Mask()
	if bit&r.ExcludedForms != 0 {
		return false
	}
	if bit&r.Forms != 0 {
		return true
	}
	if form != 0 && !form.IsStance() {
		return !r.NotShapeshifted && r.Forms == 0
	}
	return r.Forms == 0 || r.CasterForm
}

func (unit *Unit) aurasBySpellID(spellID int32) []*Aura {
	if spellID == 0 {
		return nil
	}
	var auras []*Aura
	for _, aura := range unit.auras {
		if aura.ActionID.SpellID == spellID {
			auras = append(auras, aura)
		}
	}
	return auras
}

func (spell *Spell) resolveCasterAuras() {
	spell.casterAuras = spell.Unit.aurasBySpellID(spell.CastRequirement.CasterAura)
	spell.excludeCasterAuras = spell.Unit.aurasBySpellID(spell.CastRequirement.ExcludeCasterAura)
}

func anyAuraActive(auras []*Aura) bool {
	for _, aura := range auras {
		if aura.IsActive() {
			return true
		}
	}
	return false
}

// Whether the equipment meets the weapon and shield the requirement names, read at cast time so it
// follows item swaps. Nil equipment, a unit that is not a player, meets neither.
func (r *CastRequirement) equipped(equipment *Equipment) bool {
	if r.Shield && (equipment == nil || equipment[proto.ItemSlot_ItemSlotOffHand].WeaponType != proto.WeaponType_WeaponTypeShield) {
		return false
	}
	req := r.Weapon
	if req.IsZero() {
		return true
	}
	if equipment == nil {
		return false
	}
	if req.Hand == 0 {
		return req.matchesItem(proto.ItemSlot_ItemSlotMainHand, &equipment[proto.ItemSlot_ItemSlotMainHand]) ||
			req.matchesItem(proto.ItemSlot_ItemSlotOffHand, &equipment[proto.ItemSlot_ItemSlotOffHand]) ||
			req.matchesItem(proto.ItemSlot_ItemSlotRanged, &equipment[proto.ItemSlot_ItemSlotRanged])
	}
	if req.Hand.Matches(ProcMaskMeleeMH) && !req.matchesItem(proto.ItemSlot_ItemSlotMainHand, &equipment[proto.ItemSlot_ItemSlotMainHand]) {
		return false
	}
	if req.Hand.Matches(ProcMaskMeleeOH) && !req.matchesItem(proto.ItemSlot_ItemSlotOffHand, &equipment[proto.ItemSlot_ItemSlotOffHand]) {
		return false
	}
	return true
}

// unshift: castable only once Unit.AutoUnshift has left the current form. A nil target, a weapon
// swing's, skips the target's mob type and periodic ranks; a target of unknown mob type is in no
// mask and is refused.
func (spell *Spell) castRequirementFailure(sim *Simulation, target *Unit) (reason string, unshift bool) {
	r := &spell.CastRequirement
	unit := spell.Unit
	if !r.allowsForm(unit.ShapeshiftForm) {
		if unit.AutoUnshift == nil || !r.allowsForm(0) {
			return "wrong form", false
		}
		unshift = true
	}
	if r.CasterAura != 0 && !anyAuraActive(spell.casterAuras) {
		return "missing caster aura", false
	}
	if r.ExcludeCasterAura != 0 && anyAuraActive(spell.excludeCasterAuras) {
		return "excluded caster aura", false
	}
	if !r.equipped(spell.equipment) {
		return "missing equipped item", false
	}
	if r.NotInCombat && sim.CurrentTime >= 0 {
		return "in combat", false
	}
	if r.TargetMobTypes != 0 && target != nil && r.TargetMobTypes&(1<<target.MobType) == 0 {
		return "wrong target creature type", false
	}
	if r.NotUnderHigherRank && target != nil && spell.higherRankBlocksCast(target) {
		return "a more powerful spell is active", false
	}
	return "", unshift
}

// The unit's shift spells the APL can cast, one per form, in form id order.
func (unit *Unit) shiftForms() []*Spell {
	var shifts []*Spell
	for _, spell := range unit.Spellbook {
		if spell.ShiftsTo == 0 || !spell.Flags.Matches(SpellFlagAPL) || slices.ContainsFunc(shifts, func(s *Spell) bool { return s.ShiftsTo == spell.ShiftsTo }) {
			continue
		}
		shifts = append(shifts, spell)
	}
	slices.SortFunc(shifts, func(a, b *Spell) int { return cmp.Compare(a.ShiftsTo, b.ShiftsTo) })
	return shifts
}

// Whether the spell casts the aura: it names the aura as its buff or shares its id.
func castsAura(spell *Spell, aura *Aura) bool {
	return spell.RelatedSelfBuff == aura || spell.ActionID == aura.ActionID
}

// The spell of the unit that casts the aura.
func (unit *Unit) auraSpell(aura *Aura) *Spell {
	for _, spell := range unit.Spellbook {
		if castsAura(spell, aura) {
			return spell
		}
	}
	return nil
}

package core

import (
	"fmt"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

type APLValueBossSpellIsCasting struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueBossSpellIsCasting(config *proto.APLValueBossSpellIsCasting, _ *proto.UUID) APLValue {
	spell := rot.GetTargetAPLSpell(config.SpellId, rot.GetTargetUnit(config.TargetUnit))
	if spell == nil {
		return nil
	}
	return &APLValueBossSpellIsCasting{
		spell: spell,
	}
}
func (value *APLValueBossSpellIsCasting) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueBossSpellIsCasting) GetBool(sim *Simulation) bool {
	return bossCastTimeRemaining(sim, value.spell) > 0
}
func (value *APLValueBossSpellIsCasting) String() string {
	return fmt.Sprintf("Boss is Casting(%s)", value.spell.ActionID)
}

func bossCastTimeRemaining(sim *Simulation, spell *Spell) time.Duration {
	hc := &spell.Unit.Hardcast
	if hc.ActionID != spell.ActionID {
		return 0
	}
	return max(0, hc.Expires-sim.CurrentTime)
}

func (rot *APLRotation) newValueBossSpellIsKnown(config *proto.APLValueBossSpellIsKnown, uuid *proto.UUID) APLValue {
	spell := rot.GetTargetAPLSpell(config.SpellId, rot.GetTargetUnit(config.TargetUnit))
	return rot.newValueConst(&proto.APLValueConst{
		Val: Ternary(spell != nil, "true", "false"),
	}, uuid)
}

type APLValueBossSpellCastTimeRemaining struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueBossSpellCastTimeRemaining(config *proto.APLValueBossSpellCastTimeRemaining, _ *proto.UUID) APLValue {
	spell := rot.GetTargetAPLSpell(config.SpellId, rot.GetTargetUnit(config.TargetUnit))
	if spell == nil {
		return nil
	}
	if spell.DefaultCast.CastTime <= 0 {
		rot.ValidationMessage(proto.LogLevel_Warning, "%s has no cast time, so Boss Spell Cast Time Remaining always reads 0", spell.ActionID)
		return nil
	}
	return &APLValueBossSpellCastTimeRemaining{
		spell: spell,
	}
}
func (value *APLValueBossSpellCastTimeRemaining) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueBossSpellCastTimeRemaining) GetDuration(sim *Simulation) time.Duration {
	return bossCastTimeRemaining(sim, value.spell)
}
func (value *APLValueBossSpellCastTimeRemaining) String() string {
	return fmt.Sprintf("Boss Spell Cast Time Remaining(%s)", value.spell.ActionID)
}

type APLValueBossSpellTimeToReady struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueBossSpellTimeToReady(config *proto.APLValueBossSpellTimeToReady, _ *proto.UUID) APLValue {
	spell := rot.GetTargetAPLSpell(config.SpellId, rot.GetTargetUnit(config.TargetUnit))
	if spell == nil {
		return nil
	}
	return &APLValueBossSpellTimeToReady{
		spell: spell,
	}
}
func (value *APLValueBossSpellTimeToReady) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueBossSpellTimeToReady) GetDuration(sim *Simulation) time.Duration {
	return value.spell.TimeToReady(sim)
}
func (value *APLValueBossSpellTimeToReady) String() string {
	return fmt.Sprintf("Boss Spell Time to Ready(%s)", value.spell.ActionID)
}

type APLValueBossCurrentTarget struct {
	DefaultAPLValueImpl
	player *Unit
	target UnitReference
}

func (rot *APLRotation) newValueBossCurrentTarget(config *proto.APLValueBossCurrentTarget, _ *proto.UUID) APLValue {
	unit := rot.GetSourceUnit(config.TargetUnit)
	if unit.Get() == nil {
		return nil
	}
	return &APLValueBossCurrentTarget{
		player: rot.unit,
		target: unit,
	}
}
func (value *APLValueBossCurrentTarget) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueBossCurrentTarget) GetBool(sim *Simulation) bool {
	return value.target.Get().CurrentTarget == value.player
}
func (value *APLValueBossCurrentTarget) String() string {
	return fmt.Sprintf("IsTanking(%s)", value.target.Get().Label)
}

type APLValueTargetIsMobType struct {
	DefaultAPLValueImpl
	target  UnitReference
	mobType proto.MobType
}

func (rot *APLRotation) newValueTargetIsMobType(config *proto.APLValueTargetIsMobType, _ *proto.UUID) APLValue {
	target := rot.GetTargetUnit(config.TargetUnit)
	if target.Get() == nil {
		return nil
	}
	return &APLValueTargetIsMobType{
		target:  target,
		mobType: config.MobType,
	}
}
func (value *APLValueTargetIsMobType) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueTargetIsMobType) GetBool(sim *Simulation) bool {
	return value.target.Get().MobType == value.mobType
}
func (value *APLValueTargetIsMobType) String() string {
	return fmt.Sprintf("Is Mob Type(%s, %s)", value.target.Get().Label, value.mobType)
}

type APLValueTargetIsBleedImmune struct {
	DefaultAPLValueImpl
	target UnitReference
}

func (rot *APLRotation) newValueTargetIsBleedImmune(config *proto.APLValueTargetIsBleedImmune, _ *proto.UUID) APLValue {
	target := rot.GetTargetUnit(config.TargetUnit)
	if target.Get() == nil {
		return nil
	}
	return &APLValueTargetIsBleedImmune{target: target}
}
func (value *APLValueTargetIsBleedImmune) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueTargetIsBleedImmune) GetBool(_ *Simulation) bool {
	return value.target.Get().BleedImmune
}
func (value *APLValueTargetIsBleedImmune) String() string {
	return fmt.Sprintf("Is Bleed Immune(%s)", value.target.Get().Label)
}

package core

import (
	"fmt"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

func (rot *APLRotation) newValueSpellIsKnown(config *proto.APLValueSpellIsKnown, uuid *proto.UUID) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	return rot.newValueConst(&proto.APLValueConst{
		Val: Ternary(spell != nil, "true", "false"),
	}, uuid)
}

type APLValueSpellCanCast struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellCanCast(config *proto.APLValueSpellCanCast, _ *proto.UUID) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellCanCast{
		spell: spell,
	}
}
func (value *APLValueSpellCanCast) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueSpellCanCast) GetBool(sim *Simulation) bool {
	return value.spell.CanCastOrQueue(sim, value.spell.Unit.CurrentTarget)
}
func (value *APLValueSpellCanCast) String() string {
	return fmt.Sprintf("Can Cast(%s)", value.spell.ActionID)
}

type APLValueSpellIsReady struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellIsReady(config *proto.APLValueSpellIsReady, _ *proto.UUID) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellIsReady{
		spell: spell,
	}
}
func (value *APLValueSpellIsReady) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueSpellIsReady) GetBool(sim *Simulation) bool {
	return value.spell.IsReady(sim) || (value.spell.TimeToReady(sim) <= MaxSpellQueueWindow)
}
func (value *APLValueSpellIsReady) String() string {
	return fmt.Sprintf("Is Ready(%s)", value.spell.ActionID)
}

type APLValueSpellTimeToReady struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellTimeToReady(config *proto.APLValueSpellTimeToReady, _ *proto.UUID) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellTimeToReady{
		spell: spell,
	}
}
func (value *APLValueSpellTimeToReady) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueSpellTimeToReady) GetDuration(sim *Simulation) time.Duration {
	return value.spell.TimeToReady(sim)
}
func (value *APLValueSpellTimeToReady) GetFloat(sim *Simulation) float64 {
	return value.spell.TimeToReady(sim).Seconds()
}
func (value *APLValueSpellTimeToReady) String() string {
	return fmt.Sprintf("Time To Ready(%s)", value.spell.ActionID)
}

type APLValueSpellCastTime struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellCastTime(config *proto.APLValueSpellCastTime, _ *proto.UUID) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellCastTime{
		spell: spell,
	}
}
func (value *APLValueSpellCastTime) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueSpellCastTime) GetDuration(_ *Simulation) time.Duration {
	return value.spell.CastTime()
}
func (value *APLValueSpellCastTime) String() string {
	return fmt.Sprintf("Cast Time(%s)", value.spell.ActionID)
}

type APLValueSpellTravelTime struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellTravelTime(config *proto.APLValueSpellTravelTime, _ *proto.UUID) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellTravelTime{
		spell: spell,
	}
}
func (value *APLValueSpellTravelTime) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueSpellTravelTime) GetDuration(_ *Simulation) time.Duration {
	return value.spell.TravelTime()
}
func (value *APLValueSpellTravelTime) String() string {
	return fmt.Sprintf("Travel Time(%s)", value.spell.ActionID)
}

type APLValueSpellCPM struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellCPM(config *proto.APLValueSpellCPM, _ *proto.UUID) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellCPM{
		spell: spell,
	}
}
func (value *APLValueSpellCPM) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueSpellCPM) GetFloat(sim *Simulation) float64 {
	return value.spell.CurCPM(sim)
}
func (value *APLValueSpellCPM) String() string {
	return fmt.Sprintf("CPM(%s)", value.spell.ActionID)
}

type APLValueSpellIsChanneling struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellIsChanneling(config *proto.APLValueSpellIsChanneling, _ *proto.UUID) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellIsChanneling{
		spell: spell,
	}
}
func (value *APLValueSpellIsChanneling) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueSpellIsChanneling) GetBool(_ *Simulation) bool {
	return value.spell.Unit.ChanneledDot != nil && value.spell.Unit.ChanneledDot.Spell == value.spell
}
func (value *APLValueSpellIsChanneling) String() string {
	return fmt.Sprintf("IsChanneling(%s)", value.spell.ActionID)
}

type APLValueSpellChanneledTicks struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellChanneledTicks(config *proto.APLValueSpellChanneledTicks, _ *proto.UUID) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellChanneledTicks{
		spell: spell,
	}
}
func (value *APLValueSpellChanneledTicks) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeInt
}
func (value *APLValueSpellChanneledTicks) GetInt(_ *Simulation) int32 {
	channeledDot := value.spell.Unit.ChanneledDot
	if channeledDot == nil {
		return 0
	} else {
		return channeledDot.TickCount()
	}
}
func (value *APLValueSpellChanneledTicks) String() string {
	return fmt.Sprintf("ChanneledTicks(%s)", value.spell.ActionID)
}

type APLValueSpellCurrentCost struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellCurrentCost(config *proto.APLValueSpellCurrentCost, _ *proto.UUID) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellCurrentCost{
		spell: spell,
	}
}
func (value *APLValueSpellCurrentCost) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueSpellCurrentCost) GetFloat(_ *Simulation) float64 {
	spell := value.spell
	if spell.Cost == nil {
		return 0
	}
	return spell.Cost.GetCurrentCost()
}
func (value *APLValueSpellCurrentCost) String() string {
	return fmt.Sprintf("CurrentCost(%s)", value.spell.ActionID)
}

// Spell Charges
type APLValueSpellNumCharges struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellNumCharges(config *proto.APLValueSpellNumCharges, _ *proto.UUID) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellNumCharges{
		spell: spell,
	}
}

func (value *APLValueSpellNumCharges) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeInt
}

func (value *APLValueSpellNumCharges) GetInt(_ *Simulation) int32 {
	return int32(value.spell.GetNumCharges())
}

func (value *APLValueSpellNumCharges) String() string {
	return fmt.Sprintf("SpellNumCharges(%s)", value.spell.ActionID)
}

type APLValueSpellTimeToCharge struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellTimeToCharge(config *proto.APLValueSpellTimeToCharge, _ *proto.UUID) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellTimeToCharge{
		spell: spell,
	}
}

func (value *APLValueSpellTimeToCharge) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}

func (value *APLValueSpellTimeToCharge) GetDuration(sim *Simulation) time.Duration {
	return value.spell.NextChargeIn(sim)
}

func (value *APLValueSpellTimeToCharge) GetFloat(sim *Simulation) float64 {
	return value.GetDuration(sim).Seconds()
}

func (value *APLValueSpellTimeToCharge) String() string {
	return fmt.Sprintf("SpellTimeToCharge(%s)", value.spell.ActionID)
}

// GCD duration
type APLValueSpellGCDHastedDuration struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellGCDHastedDuration(config *proto.APLValueSpellGCDHastedDuration, _ *proto.UUID) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellGCDHastedDuration{
		spell: spell,
	}
}

func (value *APLValueSpellGCDHastedDuration) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}

func (value *APLValueSpellGCDHastedDuration) GetDuration(_ *Simulation) time.Duration {
	defaultCast := value.spell.DefaultCast
	if value.spell.IgnoreHaste {
		return defaultCast.GCDTime()
	}
	gcdMin := TernaryDuration(defaultCast.GCDMin != 0, defaultCast.GCDMin, GCDMin)
	return max(gcdMin, value.spell.hastedGCD(defaultCast.GCD))
}

func (value *APLValueSpellGCDHastedDuration) GetFloat(sim *Simulation) float64 {
	return value.GetDuration(sim).Seconds()
}

func (value *APLValueSpellGCDHastedDuration) String() string {
	return fmt.Sprintf("SpellGCDHastedDuration(%s)", value.spell.ActionID)
}

// Full Cooldown duration
type APLValueSpellFullCooldown struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueSpellFullCooldown(config *proto.APLValueSpellFullCooldown, _ *proto.UUID) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellFullCooldown{
		spell: spell,
	}
}

func (value *APLValueSpellFullCooldown) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}

func (value *APLValueSpellFullCooldown) GetDuration(sim *Simulation) time.Duration {
	return value.spell.CD.Duration
}

func (value *APLValueSpellFullCooldown) String() string {
	return fmt.Sprintf("SpellFullCooldown(%s)", value.spell.ActionID)
}

// Spell In Flight
type APLValueSpellInFlight struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) NewValueSpellInFlight(config *proto.APLValueSpellInFlight, uuid *proto.UUID) APLValue {
	return rot.newValueSpellInFlight(config, uuid)
}

func (rot *APLRotation) newValueSpellInFlight(config *proto.APLValueSpellInFlight, _ *proto.UUID) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellInFlight{
		spell: spell,
	}
}

func (value *APLValueSpellInFlight) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueSpellInFlight) GetBool(sim *Simulation) bool {
	return value.spell.Unit.SpellInFlight(value.spell)
}
func (value *APLValueSpellInFlight) String() string {
	return fmt.Sprintf("SpellInFlight(%s)", value.spell.ActionID)
}

// The expected damage of one full cast if cast now, its hit and its ticks, over the time the cast
// costs the caster: the cast time or channel, or the GCD when that is longer. Only what lands before the
// fight ends counts, and a channel costs only the ticks it gets in.
type APLValueSpellExpectedDpet struct {
	DefaultAPLValueImpl
	spell     *Spell
	targetRef UnitReference
}

func (rot *APLRotation) newValueSpellExpectedDpet(config *proto.APLValueSpellExpectedDpet, _ *proto.UUID) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil || (spell.expectedInitialDamageInternal == nil && spell.expectedTickDamageInternal == nil) {
		return nil
	}
	return &APLValueSpellExpectedDpet{
		spell:     spell,
		targetRef: rot.GetTargetUnit(config.TargetUnit),
	}
}

func (value *APLValueSpellExpectedDpet) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}

func (value *APLValueSpellExpectedDpet) GetFloat(sim *Simulation) float64 {
	spell := value.spell
	target := value.targetRef.Get()

	cast := spell.DefaultCast
	if !spell.IgnoreHaste {
		cast.GCD = spell.hastedGCD(cast.GCD)
		cast.CastTime = spell.CastTime().Round(time.Millisecond)
	}

	remaining := sim.GetRemainingDuration()
	landsAt := cast.CastTime + spell.TravelTime()
	if landsAt > remaining {
		return 0
	}

	var damage float64
	if spell.expectedInitialDamageInternal != nil {
		damage += spell.ExpectedInitialDamage(sim, target)
	}
	if dot := spell.Dot(target); dot != nil {
		channeled := spell.Flags.Matches(SpellFlagChanneled)
		tickPeriod := dot.CalcTickPeriod()
		// A channel ticks from the start of the cast, a DoT from when it lands.
		firstTickFrom := landsAt
		if channeled {
			firstTickFrom = 0
		}
		ticks := int32(min(time.Duration(dot.ExpectedTickCount()), (remaining-firstTickFrom)/tickPeriod))
		if spell.expectedTickDamageInternal != nil {
			damage += spell.ExpectedTickDamage(sim, target) * float64(ticks)
		}
		if channeled {
			cast.CastTime = tickPeriod * time.Duration(ticks)
		}
	}

	// A spell with no GCD and no cast time (an auto attack, an off-GCD instant) scores 0.
	executeTime := cast.EffectiveTime()
	if executeTime == 0 {
		return 0
	}
	return damage / executeTime.Seconds()
}

func (value *APLValueSpellExpectedDpet) String() string {
	return fmt.Sprintf("Spell Expected DPET(%s)", value.spell.ActionID)
}

// The table's base chances are float sums, so a stat equal to the cap can leave a remainder of about 1e-17.
const capTolerance = 1e-9

// Whether the spell cannot miss the current target: the spell table's miss for magic, the weapon miss of a
// special for melee and ranged, which leaves out the dual-wield penalty of auto attacks. It reads the table
// alone, so a spell that always hits or rolls its miss through another spell reads as its table does.
type APLValueSpellIsHitCapped struct {
	DefaultAPLValueImpl
	spell     *Spell
	targetRef UnitReference
	magic     bool
}

func (rot *APLRotation) newValueSpellIsHitCapped(config *proto.APLValueSpellIsHitCapped, _ *proto.UUID) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellIsHitCapped{
		spell:     spell,
		targetRef: rot.GetTargetUnit(nil),
		magic:     spell.rolledDefenseType() == DefenseTypeMagic,
	}
}
func (value *APLValueSpellIsHitCapped) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueSpellIsHitCapped) GetBool(_ *Simulation) bool {
	spell := value.spell
	attackTable := spell.Unit.AttackTables[value.targetRef.Get().UnitIndex]
	if value.magic {
		return spell.SpellChanceToMiss(attackTable) <= capTolerance
	}
	return spell.physicalMissChanceNoDWPenalty(attackTable) <= capTolerance
}
func (value *APLValueSpellIsHitCapped) String() string {
	return fmt.Sprintf("Is Hit Capped(%s)", value.spell.ActionID)
}

// Whether the current target can neither dodge the spell nor, while the unit stands in front of it, parry it.
// Only the melee table has dodge and parry, so anything else is capped. It reads the table alone, so a
// special whose outcome never rolls dodge or parry still counts them.
type APLValueSpellIsExpertiseCapped struct {
	DefaultAPLValueImpl
	spell     *Spell
	targetRef UnitReference
	melee     bool
}

func (rot *APLRotation) newValueSpellIsExpertiseCapped(config *proto.APLValueSpellIsExpertiseCapped, _ *proto.UUID) APLValue {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	return &APLValueSpellIsExpertiseCapped{
		spell:     spell,
		targetRef: rot.GetTargetUnit(nil),
		melee:     spell.rolledDefenseType() == DefenseTypeMelee,
	}
}
func (value *APLValueSpellIsExpertiseCapped) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueSpellIsExpertiseCapped) GetBool(_ *Simulation) bool {
	if !value.melee {
		return true
	}
	spell := value.spell
	target := value.targetRef.Get()
	attackTable := spell.Unit.AttackTables[target.UnitIndex]
	if !spell.Flags.Matches(SpellFlagCannotBeDodged) && target.GetTotalDodgeChanceAsDefender(spell, attackTable) > capTolerance {
		return false
	}
	return !spell.Unit.PseudoStats.InFrontOfTarget || target.GetTotalParryChanceAsDefender(spell, attackTable) <= capTolerance
}
func (value *APLValueSpellIsExpertiseCapped) String() string {
	return fmt.Sprintf("Is Expertise Capped(%s)", value.spell.ActionID)
}

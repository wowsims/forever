package core

import (
	"fmt"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

type APLValueCurrentHealth struct {
	DefaultAPLValueImpl
	unit UnitReference
}

func (rot *APLRotation) newValueCurrentHealth(config *proto.APLValueCurrentHealth, uuid *proto.UUID) APLValue {
	unit := rot.GetSourceUnit(config.SourceUnit)
	resolvedUnit := unit.Get()
	if resolvedUnit == nil {
		return nil
	}
	if !rot.enemyHasHealth(resolvedUnit, uuid) {
		return nil
	}
	return &APLValueCurrentHealth{
		unit: unit,
	}
}

// Whether the unit has health to read: an enemy has its target curve, others need a health bar. Warns when not,
// and for an enemy without a Health stat, which reads 0.
func (rot *APLRotation) enemyHasHealth(unit *Unit, uuid *proto.UUID) bool {
	if unit.Type != EnemyUnit {
		if !unit.HasHealthBar() {
			rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s does not use Health", unit.Label)
			return false
		}
		return true
	}
	if unit.stats[stats.Health] == 0 {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s has no Health stat", unit.Label)
	}
	return true
}
func (value *APLValueCurrentHealth) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueCurrentHealth) GetFloat(sim *Simulation) float64 {
	unit := value.unit.Get()
	if target := enemyTargetOf(unit); target != nil {
		return target.HealthPercent(sim) / 100 * unit.stats[stats.Health]
	}
	return unit.CurrentHealth()
}
func (value *APLValueCurrentHealth) String() string {
	return "Current Health"
}

type APLValueCurrentHealthPercent struct {
	DefaultAPLValueImpl
	unit UnitReference
}

func (rot *APLRotation) newValueCurrentHealthPercent(config *proto.APLValueCurrentHealthPercent, uuid *proto.UUID) APLValue {
	unit := rot.GetSourceUnit(config.SourceUnit)
	resolvedUnit := unit.Get()

	if resolvedUnit == nil {
		return nil
	}
	if !rot.enemyHasHealth(resolvedUnit, uuid) {
		return nil
	}
	return &APLValueCurrentHealthPercent{
		unit: unit,
	}
}
func (value *APLValueCurrentHealthPercent) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueCurrentHealthPercent) GetFloat(sim *Simulation) float64 {
	unit := value.unit.Get()
	if target := enemyTargetOf(unit); target != nil {
		return target.HealthPercent(sim) / 100
	}
	if !unit.HasHealthBar() {
		return 0
	}
	return unit.CurrentHealthPercent()
}
func (value *APLValueCurrentHealthPercent) String() string {
	return fmt.Sprintf("Current Health %%")
}

type APLValueMaxHealth struct {
	DefaultAPLValueImpl
	unit *Unit
}

func (rot *APLRotation) newValueMaxHealth(_ *proto.APLValueMaxHealth, uuid *proto.UUID) APLValue {
	unit := rot.unit
	if !unit.HasHealthBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s does not use Health", unit.Label)
		return nil
	}
	return &APLValueMaxHealth{
		unit: unit,
	}
}
func (value *APLValueMaxHealth) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueMaxHealth) GetFloat(sim *Simulation) float64 {
	return value.unit.MaxHealth()
}
func (value *APLValueMaxHealth) String() string {
	return "Max Health"
}

type APLValueCurrentMana struct {
	DefaultAPLValueImpl
	unit UnitReference
}

func (rot *APLRotation) newValueCurrentMana(config *proto.APLValueCurrentMana, uuid *proto.UUID) APLValue {
	unit := rot.GetSourceUnit(config.SourceUnit)
	resolvedUnit := unit.Get()

	if resolvedUnit == nil {
		return nil
	}
	if !resolvedUnit.HasManaBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s does not use Mana", resolvedUnit.Label)
		return nil
	}
	return &APLValueCurrentMana{
		unit: unit,
	}
}
func (value *APLValueCurrentMana) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueCurrentMana) GetFloat(sim *Simulation) float64 {
	// During a hardcast, the mana left once the cast pays its cost, as the cost checks read it.
	return value.unit.Get().manaAfterHardcast(sim)
}
func (value *APLValueCurrentMana) String() string {
	return "Current Mana"
}

type APLValueCurrentManaPercent struct {
	DefaultAPLValueImpl
	unit UnitReference
}

func (rot *APLRotation) newValueCurrentManaPercent(config *proto.APLValueCurrentManaPercent, uuid *proto.UUID) APLValue {
	unit := rot.GetSourceUnit(config.SourceUnit)
	resolvedUnit := unit.Get()

	if resolvedUnit == nil {
		return nil
	}
	if !resolvedUnit.HasManaBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s does not use Mana", resolvedUnit.Label)
		return nil
	}
	return &APLValueCurrentManaPercent{
		unit: unit,
	}
}
func (value *APLValueCurrentManaPercent) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueCurrentManaPercent) GetFloat(sim *Simulation) float64 {
	unit := value.unit.Get()
	return unit.manaAfterHardcast(sim) / unit.MaxMana()
}
func (value *APLValueCurrentManaPercent) String() string {
	return fmt.Sprintf("Current Mana %%")
}

type APLValueMaxMana struct {
	DefaultAPLValueImpl
	unit *Unit
}

func (rot *APLRotation) newValueMaxMana(_ *proto.APLValueMaxMana, uuid *proto.UUID) APLValue {
	unit := rot.unit
	if !unit.HasManaBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s does not use Mana", unit.Label)
		return nil
	}
	return &APLValueMaxMana{
		unit: unit,
	}
}
func (value *APLValueMaxMana) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueMaxMana) GetFloat(sim *Simulation) float64 {
	return value.unit.MaxMana()
}
func (value *APLValueMaxMana) String() string {
	return "Max Mana"
}

type APLValueCurrentManaRegen struct {
	DefaultAPLValueImpl
	unit *Unit
}

func (rot *APLRotation) newValueCurrentManaRegen(_ *proto.APLValueCurrentManaRegen, uuid *proto.UUID) APLValue {
	unit := rot.unit
	if !unit.HasManaBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s does not use Mana", unit.Label)
		return nil
	}
	return &APLValueCurrentManaRegen{
		unit: unit,
	}
}
func (value *APLValueCurrentManaRegen) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueCurrentManaRegen) GetFloat(sim *Simulation) float64 {
	return value.unit.manaRegenWhileCasting
}
func (value *APLValueCurrentManaRegen) String() string {
	return "Mana Regen while casting"
}

type APLValueCurrentRage struct {
	DefaultAPLValueImpl
	unit *Unit
}

func (rot *APLRotation) newValueCurrentRage(_ *proto.APLValueCurrentRage, uuid *proto.UUID) APLValue {
	unit := rot.unit
	if !unit.HasRageBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s does not use Rage", unit.Label)
		return nil
	}
	return &APLValueCurrentRage{
		unit: unit,
	}
}
func (value *APLValueCurrentRage) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueCurrentRage) GetFloat(sim *Simulation) float64 {
	return value.unit.CurrentRage()
}
func (value *APLValueCurrentRage) String() string {
	return "Current Rage"
}

type APLValueMaxRage struct {
	DefaultAPLValueImpl
	unit *Unit
}

func (rot *APLRotation) newValueMaxRage(_ *proto.APLValueMaxRage, uuid *proto.UUID) APLValue {
	unit := rot.unit
	if !unit.HasRageBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Error, "%s does not use Rage", unit.Label)
		return nil
	}
	return &APLValueMaxRage{
		unit: unit,
	}
}
func (value *APLValueMaxRage) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueMaxRage) GetFloat(sim *Simulation) float64 {
	return value.unit.MaximumRage()
}
func (value *APLValueMaxRage) String() string {
	return "Max Rage"
}

type APLValueCurrentEnergy struct {
	DefaultAPLValueImpl
	unit *Unit
}

func (rot *APLRotation) newValueCurrentEnergy(_ *proto.APLValueCurrentEnergy, uuid *proto.UUID) APLValue {
	unit := rot.unit
	if !unit.HasEnergyBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s does not use Energy", unit.Label)
		return nil
	}
	return &APLValueCurrentEnergy{
		unit: unit,
	}
}
func (value *APLValueCurrentEnergy) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueCurrentEnergy) GetFloat(sim *Simulation) float64 {
	return value.unit.CurrentEnergy()
}
func (value *APLValueCurrentEnergy) String() string {
	return "Current Energy"
}

type APLValueMaxEnergy struct {
	DefaultAPLValueImpl
	unit *Unit
}

func (rot *APLRotation) newValueMaxEnergy(_ *proto.APLValueMaxEnergy, uuid *proto.UUID) APLValue {
	unit := rot.unit
	if !unit.HasEnergyBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Error, "%s does not use Energy", unit.Label)
		return nil
	}
	return &APLValueMaxEnergy{
		unit: unit,
	}
}
func (value *APLValueMaxEnergy) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueMaxEnergy) GetFloat(sim *Simulation) float64 {
	return value.unit.MaximumEnergy()
}
func (value *APLValueMaxEnergy) String() string {
	return "Max Energy"
}

type APLValueTimeToNextEnergyTick struct {
	DefaultAPLValueImpl
	unit *Unit
}

func (rot *APLRotation) newValueTimeToNextEnergyTick(_ *proto.APLValueTimeToNextEnergyTick, uuid *proto.UUID) APLValue {
	unit := rot.unit
	if !unit.HasEnergyBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s does not use Energy", unit.Label)
		return nil
	}
	return &APLValueTimeToNextEnergyTick{
		unit: unit,
	}
}
func (value *APLValueTimeToNextEnergyTick) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueTimeToNextEnergyTick) GetDuration(sim *Simulation) time.Duration {
	return value.unit.TimeToNextEnergyTick(sim)
}
func (value *APLValueTimeToNextEnergyTick) String() string {
	return "Time To Next Energy Tick"
}

type APLValueEnergyTimeToTarget struct {
	DefaultAPLValueImpl
	unit         *Unit
	targetEnergy APLValue
}

func (rot *APLRotation) newValueEnergyTimeToTarget(config *proto.APLValueEnergyTimeToTarget, uuid *proto.UUID) APLValue {
	unit := rot.unit
	if !unit.HasEnergyBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s does not use Energy", unit.Label)
		return nil
	}

	targetEnergy := rot.coerceTo(rot.newAPLValue(config.TargetEnergy), proto.APLValueType_ValueTypeFloat)
	if targetEnergy == nil {
		return nil
	}

	return &APLValueEnergyTimeToTarget{
		unit:         unit,
		targetEnergy: targetEnergy,
	}
}
func (value *APLValueEnergyTimeToTarget) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueEnergyTimeToTarget) GetDuration(sim *Simulation) time.Duration {
	return value.unit.TimeToTargetEnergy(sim, value.targetEnergy.GetFloat(sim))
}
func (value *APLValueEnergyTimeToTarget) String() string {
	return "Estimated Time To Target Energy"
}

type APLValueCurrentComboPoints struct {
	DefaultAPLValueImpl
	unit *Unit
}

func (rot *APLRotation) newValueCurrentComboPoints(_ *proto.APLValueCurrentComboPoints, uuid *proto.UUID) APLValue {
	unit := rot.unit
	if !unit.HasEnergyBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s does not use Combo Points", unit.Label)
		return nil
	}
	return &APLValueCurrentComboPoints{
		unit: unit,
	}
}
func (value *APLValueCurrentComboPoints) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeInt
}
func (value *APLValueCurrentComboPoints) GetInt(sim *Simulation) int32 {
	return value.unit.ComboPoints()
}
func (value *APLValueCurrentComboPoints) String() string {
	return "Current Combo Points"
}

type APLValueMaxComboPoints struct {
	DefaultAPLValueImpl
	maxComboPoints int32
}

func (rot *APLRotation) newValueMaxComboPoints(_ *proto.APLValueMaxComboPoints, uuid *proto.UUID) APLValue {
	unit := rot.unit
	if !unit.HasEnergyBar() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Error, "%s does not use Combo Points", unit.Label)
		return nil
	}
	return &APLValueMaxComboPoints{
		maxComboPoints: unit.MaxComboPoints(),
	}
}
func (value *APLValueMaxComboPoints) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeInt
}
func (value *APLValueMaxComboPoints) GetInt(sim *Simulation) int32 {
	return value.maxComboPoints
}
func (value *APLValueMaxComboPoints) String() string {
	return fmt.Sprintf("Max Combo Points(%d)", value.maxComboPoints)
}

type APLValueCurrentGenericResource struct {
	DefaultAPLValueImpl
	unit *Unit
}

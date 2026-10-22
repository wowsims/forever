package druid

import (
	"fmt"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func (druid *Druid) NewAPLValue(rot *core.APLRotation, config *proto.APLValue) core.APLValue {
	switch config.Value.(type) {
	case *proto.APLValue_DruidBleedIsActive:
		return druid.newValueDruidBleedIsActive(rot, config.GetDruidBleedIsActive())
	case *proto.APLValue_DruidExcessEnergy:
		return &APLValueDruidExcessEnergy{druid: druid}
	default:
		return nil
	}
}

type APLValueDruidBleedIsActive struct {
	core.DefaultAPLValueImpl
	druid  *Druid
	target core.UnitReference
}

func (druid *Druid) newValueDruidBleedIsActive(rot *core.APLRotation, config *proto.APLValueDruidBleedIsActive) core.APLValue {
	target := rot.GetTargetUnit(config.TargetUnit)
	if target.Get() == nil {
		return nil
	}
	return &APLValueDruidBleedIsActive{
		druid:  druid,
		target: target,
	}
}

func (value *APLValueDruidBleedIsActive) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}

func (value *APLValueDruidBleedIsActive) GetBool(_ *core.Simulation) bool {
	target := value.target.Get()
	return target != nil && value.druid.targetIsBleeding(target)
}

func (value *APLValueDruidBleedIsActive) String() string {
	return fmt.Sprintf("Bleed Is Active(%s)", value.target.Get().Label)
}

type APLValueDruidExcessEnergy struct {
	core.DefaultAPLValueImpl
	druid *Druid
}

func (value *APLValueDruidExcessEnergy) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}

func (value *APLValueDruidExcessEnergy) GetFloat(sim *core.Simulation) float64 {
	return value.druid.excessEnergy(sim)
}

func (value *APLValueDruidExcessEnergy) String() string {
	return "Excess Energy()"
}

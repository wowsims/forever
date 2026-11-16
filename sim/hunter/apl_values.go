package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func (hunter *Hunter) NewAPLValue(rot *core.APLRotation, config *proto.APLValue) core.APLValue {
	switch config.Value.(type) {
	case *proto.APLValue_HunterActiveHawks:
		return &APLValueHunterActiveHawks{hunter: hunter}
	default:
		return nil
	}
}

// A hawk counts from the moment the dive lands until it leaves or the next one replaces it.
type APLValueHunterActiveHawks struct {
	core.DefaultAPLValueImpl
	hunter *Hunter
}

func (value *APLValueHunterActiveHawks) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeInt
}

func (value *APLValueHunterActiveHawks) GetInt(_ *core.Simulation) int32 {
	active := int32(0)
	for _, hawk := range value.hunter.Hawks {
		if hawk.IsEnabled() {
			active++
		}
	}
	return active
}

func (value *APLValueHunterActiveHawks) String() string {
	return "Active Hawks()"
}

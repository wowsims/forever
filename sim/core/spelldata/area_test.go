package spelldata

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func TestAreaBonusReadsAnyOfItsGroups(t *testing.T) {
	bonus := &Spell{AreaBonusGroups: []int32{9163, 9202}, AreaMultiplier: 2, AreaDurationMultiplier: 1}
	plain := &Spell{}

	for _, tc := range []struct {
		name           string
		areas          []proto.AreaType
		amount, length float64
	}{
		{"no area", nil, 1, 1},
		{"another area", []proto.AreaType{proto.AreaType_AreaTypeForestGrassland}, 1, 1},
		{"the second group", []proto.AreaType{proto.AreaType_AreaTypeSnowy, proto.AreaType_AreaTypeHaunted}, 2, 1},
		{"the first group", []proto.AreaType{proto.AreaType_AreaTypeWasteland}, 2, 1},
	} {
		encounter := &core.Encounter{AreaTypes: tc.areas}
		if amount, length := bonus.AreaBonus(); amount.Factor(encounter) != tc.amount || length.Factor(encounter) != tc.length {
			t.Errorf("%s: bonus reads x%v (duration x%v), want x%v (duration x%v)", tc.name, amount.Factor(encounter), length.Factor(encounter), tc.amount, tc.length)
		}
		if amount, length := plain.AreaBonus(); amount.Factor(encounter) != 1 || length.Factor(encounter) != 1 {
			t.Errorf("%s: a row without a bonus reads x%v (duration x%v)", tc.name, amount.Factor(encounter), length.Factor(encounter))
		}
	}

	volcanic := &core.Encounter{AreaTypes: []proto.AreaType{proto.AreaType_AreaTypeVolcanic}}
	if amount, length := Nil.AreaBonus(); amount.Factor(volcanic) != 1 || length.Factor(volcanic) != 1 {
		t.Errorf("the nil row reads x%v (duration x%v)", amount.Factor(volcanic), length.Factor(volcanic))
	}
}

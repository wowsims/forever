package shared

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// Both buffs carry area bonuses from the override table.
func TestOnUseStatBuffScalesByTheEncounterArea(t *testing.T) {
	ap := int32(proto.Stat_StatAttackPower)
	moltenFury := &proto.ItemEffect{
		BuffId:           1249113,
		EffectDurationMs: 15000,
		ScalingOptions:   map[int32]*proto.ScalingItemEffectProperties{0: {Stats: map[int32]float64{ap: 55}}},
	}
	monkeyBusiness := &proto.ItemEffect{
		BuffId:           1287571,
		EffectDurationMs: 15000,
		ScalingOptions:   map[int32]*proto.ScalingItemEffectProperties{0: {Stats: map[int32]float64{ap: 16}}},
	}

	for _, tc := range []struct {
		name     string
		areas    []proto.AreaType
		effect   *proto.ItemEffect
		wantAP   float64
		wantTime time.Duration
	}{
		{"Molten Fury outside", nil, moltenFury, 55, 15 * time.Second},
		{"Molten Fury in Forest", []proto.AreaType{proto.AreaType_AreaTypeForestGrassland}, moltenFury, 55, 15 * time.Second},
		{"Molten Fury in Blackrock", []proto.AreaType{proto.AreaType_AreaTypeMountainous, proto.AreaType_AreaTypeVolcanic}, moltenFury, 110, 15 * time.Second},
		{"Monkey Business in Forest", []proto.AreaType{proto.AreaType_AreaTypeForestGrassland}, monkeyBusiness, 32, 30 * time.Second},
	} {
		agent := newTestAgent()
		agent.character.Env = &core.Environment{Encounter: core.Encounter{AreaTypes: tc.areas}}

		buffStats, duration := onUseStatBuff(agent.character, 0, tc.effect)
		if got := buffStats[stats.AttackPower]; got != tc.wantAP {
			t.Errorf("%s: attack power = %v, want %v", tc.name, got, tc.wantAP)
		}
		if duration != tc.wantTime {
			t.Errorf("%s: duration = %v, want %v", tc.name, duration, tc.wantTime)
		}
	}
}

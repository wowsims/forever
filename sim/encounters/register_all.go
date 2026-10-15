package encounters

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func init() {
	AddDefaultPresetEncounter()
	addMovementAI()
	addDynamicAddsAI()
	addCustomBossAI()
}

func AddDefaultPresetEncounter() {
	core.AddPresetTarget(&core.PresetTarget{
		PathPrefix: "Default",
		Config: &proto.Target{
			Id:        31146,
			Name:      "Raid Target",
			Level:     63,
			MobType:   proto.MobType_MobTypeMechanical,
			TankIndex: 0,

			Stats: stats.Stats{
				stats.Health:      1_428_900,
				stats.Armor:       3731,
				stats.AttackPower: 320,
			}.ToProtoArray(),

			SpellSchool:      proto.SpellSchool_SpellSchoolPhysical,
			SwingSpeed:       2,
			MinBaseDamage:    2601,
			DamageSpread:     0.5,
			SuppressDodge:    false,
			ParryHaste:       true,
			CanCrush:         true,
			DualWield:        false,
			DualWieldPenalty: false,
			TargetInputs:     []*proto.TargetInput{},
		},
		AI: nil,
	})
	core.AddPresetEncounter("Raid Target", []string{
		"Default/Raid Target",
	})
}

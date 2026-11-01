package encounters

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// The fight length the designs' second-based numbers are fractions of; the UI's default duration.
const designBaseSeconds = 180.0

const (
	movementTargetID   int32 = 31147
	dynamicBossID      int32 = 99999
	dynamicAddID       int32 = 99998
	customBossTargetID int32 = 100000
)

func init() {
	AddDefaultPresetEncounter()
	addMovementEncounter()
	addDynamicAddsEncounter()
	addCustomBossEncounter()
}

func fractionOfFight(seconds float64) float64 {
	return seconds / designBaseSeconds
}

// A level-63 boss with a level-63 creature's armor and health from ExpectedStat, the health times 300 for a boss
// (4763 x 300). The swings hit with Chromaggus's MinBaseDamage.
func bossTarget(id int32, name string, damageSpread float64) *proto.Target {
	return &proto.Target{
		Id:        id,
		Name:      name,
		Level:     63,
		MobType:   proto.MobType_MobTypeMechanical,
		TankIndex: 0,

		Stats: stats.Stats{
			stats.Health:      1_428_900,
			stats.Armor:       3731,
			stats.AttackPower: 320,
		}.ToProtoArray(),

		SpellSchool:   proto.SpellSchool_SpellSchoolPhysical,
		SwingSpeed:    2,
		MinBaseDamage: 2601,
		DamageSpread:  damageSpread,
		ParryHaste:    true,
		CanCrush:      true,
	}
}

func addMovementEncounter() {
	core.AddPresetTarget(&core.PresetTarget{
		PathPrefix: "Default",
		Config:     bossTarget(movementTargetID, "Movement", 0.5),
	})
	core.AddPresetEncounter("Movement", []string{
		"Default/Movement",
	}, &proto.EncounterDesign{
		Moves: []*proto.DesignMove{{
			At:     &proto.DesignTime{},
			Yards:  5,
			Repeat: &proto.DesignRepeat{Interval: fractionOfFight(10)},
		}},
	})
}

func dynamicAddTarget() *proto.Target {
	const level = 62
	return &proto.Target{
		Id:        dynamicAddID,
		Name:      "Dynamic Add",
		Level:     level,
		MobType:   proto.MobType_MobTypeMechanical,
		TankIndex: 1,

		Stats: stats.Stats{
			stats.Armor:       core.CreatureArmorByLevel[level],
			stats.AttackPower: 320,
		}.ToProtoArray(),

		SpellSchool:   proto.SpellSchool_SpellSchoolPhysical,
		SwingSpeed:    1.5,
		MinBaseDamage: 5240,
		DamageSpread:  0.4,
	}
}

func addDynamicAddsEncounter() {
	core.AddPresetTarget(&core.PresetTarget{
		PathPrefix: "Default",
		Config:     bossTarget(dynamicBossID, "Dynamic Boss", 0.5),
	})
	core.AddPresetTarget(&core.PresetTarget{
		PathPrefix: "Default",
		Config:     dynamicAddTarget(),
	})
	core.AddPresetEncounter("Dynamic Adds", []string{
		"Default/Dynamic Boss",
	}, &proto.EncounterDesign{
		AddWaves: []*proto.AddWave{{
			Name:                 "Add",
			Adds:                 []*proto.Target{dynamicAddTarget()},
			Spawn:                &proto.DesignTime{At: fractionOfFight(10)},
			Lifetime:             fractionOfFight(20),
			Repeat:               &proto.DesignRepeat{Interval: fractionOfFight(30)},
			ExecuteProportion_20: 0.2,
			ExecuteProportion_35: 0.35,
			KillShare:            0.2,
		}},
	})
}

// A tank example target. It ships with no boss abilities; the player adds them in the encounter designer.
func addCustomBossEncounter() {
	core.AddPresetTarget(&core.PresetTarget{
		PathPrefix: "Default",
		Config:     bossTarget(customBossTargetID, "Boss", 0.413),
	})
	core.AddPresetEncounter("Custom Boss", []string{
		"Default/Boss",
	}, nil)
}

func AddDefaultPresetEncounter() {
	core.AddPresetTarget(&core.PresetTarget{
		PathPrefix: "Default",
		Config:     bossTarget(31146, "Raid Target", 0.5),
	})
	core.AddPresetEncounter("Raid Target", []string{
		"Default/Raid Target",
	}, nil)
}

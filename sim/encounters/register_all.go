package encounters

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// The fight length the plans' second-based numbers are fractions of; the UI's default duration.
const planBaseSeconds = 180.0

const dynamicAddID int32 = 99998

// Every preset encounter fights the Raid Target; what sets them apart is their plan.
func init() {
	AddDefaultPresetEncounter()
	addMovementEncounter()
	addDynamicAddsEncounter()
}

func fractionOfFight(seconds float64) float64 {
	return seconds / planBaseSeconds
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
	core.AddPresetEncounter("Movement", []string{
		"Default/Raid Target",
	}, &proto.EncounterPlan{
		Moves: []*proto.PlanMove{{
			At:     &proto.PlanTime{},
			Yards:  5,
			Repeat: &proto.PlanRepeat{Interval: fractionOfFight(10)},
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
	core.AddPresetEncounter("Dynamic Adds", []string{
		"Default/Raid Target",
	}, &proto.EncounterPlan{
		AddWaves: []*proto.AddWave{{
			Name:                 "Add",
			Adds:                 []*proto.Target{dynamicAddTarget()},
			Spawn:                &proto.PlanTime{At: fractionOfFight(10)},
			Lifetime:             fractionOfFight(20),
			Repeat:               &proto.PlanRepeat{Interval: fractionOfFight(30)},
			ExecuteProportion_20: 0.2,
			ExecuteProportion_35: 0.35,
			KillShare:            0.2,
		}},
	})
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

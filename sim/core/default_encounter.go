package core

import (
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// The fight the site starts with (Encounter.defaultEncounterProto in ui/sim/raid/encounter.ts): 180 s varied by 5 s,
// against one default target. Each call returns a new message.
func SiteDefaultEncounter() *proto.Encounter {
	return &proto.Encounter{
		Duration:             180,
		DurationVariation:    5,
		ExecuteProportion_20: 0.2,
		ExecuteProportion_35: 0.35,
		Targets:              []*proto.Target{SiteDefaultTarget()},
	}
}

// The site's default target (Encounter.defaultTargetProto), the raid target the encounters register. Each call
// returns a new message.
func SiteDefaultTarget() *proto.Target {
	return &proto.Target{
		Id:      31146,
		Name:    "Raid Target",
		Level:   CharacterLevel + 3,
		MobType: proto.MobType_MobTypeMechanical,
		Stats: stats.Stats{
			stats.Armor:       3731,
			stats.AttackPower: 320,
			stats.BlockValue:  54,
			stats.Health:      1428900,
		}.ToProtoArray(),
		MinBaseDamage: 2601,
		DamageSpread:  0.5,
		SwingSpeed:    2,
		ParryHaste:    true,
		CanCrush:      true,
	}
}

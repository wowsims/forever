//go:build with_db

package feralcat

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func spellReady(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_SpellIsReady{SpellIsReady: &proto.APLValueSpellIsReady{
		SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}},
	}}}
}

func spellKnown(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_SpellIsKnown{SpellIsKnown: &proto.APLValueSpellIsKnown{
		SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}},
	}}}
}

func currentEnergy() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_CurrentEnergy{CurrentEnergy: &proto.APLValueCurrentEnergy{}}}
}

// Sets one feral-tree digit. Index 8 is Feral Charge, 12 Predatory Discipline, 19 Shifting Power.
func withFeralDigit(talents string, index int, digit byte) string {
	parts := []byte(talents)
	start := 0
	for i, c := range parts {
		if c == '-' {
			start = i + 1
			break
		}
	}
	parts[start+index] = digit
	return string(parts)
}

// A 180-second fight against a level-63 demon with the cat UI's default raid buffs and debuffs,
// the Level 60 gear preset (ui/specs/druid/feralcat/gear_sets/default.gear.json), and no consumables.
// Weapon is Fist of Cenarius. Howling Idol shortens Shifting Power by 1s.
func sweepRequest(talents string, rotation *proto.APLRotation, cannotShred bool, iterations int32) *proto.RaidSimRequest {
	return &proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 101, Iterations: iterations, IsTest: true},
		Raid: &proto.Raid{
			Parties: []*proto.Party{{
				Players: []*proto.Player{{
					Name: "Cat", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
					TalentsString:  talents,
					ReactionTimeMs: 250,
					Equipment:      core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "default").GearSet,
					Consumables:    &proto.ConsumesSpec{},
					Buffs: &proto.IndividualBuffs{
						GreaterBlessingOfKings: true,
						GreaterBlessingOfMight: true,
					},
					Spec: &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{
						Options: &proto.FeralCatDruid_Options{CannotShredTarget: cannotShred},
					}},
					Rotation: rotation,
				}},
				Buffs: &proto.PartyBuffs{
					ManaSpringTotem: proto.TristateEffect_TristateEffectRegular,
				},
			}},
			Buffs: &proto.RaidBuffs{
				BattleShout:              1, // Rank 6
				ArcaneBrilliance:         true,
				PrayerOfSpirit:           true,
				GiftOfTheWild:            true,
				PrayerOfFortitude:        true,
				PrayerOfShadowProtection: true,
				WindfuryTotem:            true,
				StrengthOfEarthTotem:     true,
			},
			Debuffs: &proto.Debuffs{
				ExposeArmor:         true,
				HuntersMark:         true,
				JudgementOfWisdom:   true,
				CurseOfRecklessness: true,
				FaerieFire:          true,
				SunderArmor:         true,
				GiftOfArthas:        true,
			},
		},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{
			Name: "target", Level: 63, MobType: proto.MobType_MobTypeDemon,
		}}, Duration: 180},
	}
}

package sim

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// A character with no talents, gear or buffs deals every school's damage and its healing at a factor of 1,
// whatever its class, and only a warrior starts in a form, its stance.
func TestDoneMultipliersAreOneOnAPlainCharacter(t *testing.T) {
	players := []*proto.Player{
		{Class: proto.Class_ClassWarrior, Race: proto.Race_RaceHuman, Spec: &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{Options: &proto.DpsWarrior_Options{
			ClassOptions: &proto.WarriorOptions{DefaultStance: proto.WarriorStance_WarriorStanceBattle}}}}},
		{Class: proto.Class_ClassMage, Race: proto.Race_RaceHuman, Spec: &proto.Player_Mage{Mage: &proto.Mage{Options: &proto.Mage_Options{ClassOptions: &proto.MageOptions{}}}}},
		{Class: proto.Class_ClassPriest, Race: proto.Race_RaceHuman, Spec: &proto.Player_HealerPriest{HealerPriest: &proto.HealerPriest{Options: &proto.HealerPriest_Options{ClassOptions: &proto.PriestOptions{}}}}},
		{Class: proto.Class_ClassRogue, Race: proto.Race_RaceHuman, Spec: &proto.Player_Rogue{Rogue: &proto.Rogue{Options: &proto.Rogue_Options{ClassOptions: &proto.RogueOptions{}}}}},
	}
	for _, player := range players {
		player.Name = player.Class.String()
		player.Equipment = &proto.EquipmentSpec{}
		player.Buffs = &proto.IndividualBuffs{}
		player.Rotation = &proto.APLRotation{}
		result := core.ComputeStats(&proto.ComputeStatsRequest{
			Raid:      core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter: STEncounter,
		})
		if result.ErrorResult != "" {
			t.Fatalf("%s: %s", player.Name, result.ErrorResult)
		}
		done := result.RaidStats.Parties[0].Players[0].GetDoneMultipliers()
		if got := len(done.GetDamage()); got != len(proto.SpellSchool_name) {
			t.Fatalf("%s states %d school damage multipliers, want one for each of the %d schools", player.Name, got, len(proto.SpellSchool_name))
		}
		for school, got := range done.GetDamage() {
			if got != 1 {
				t.Errorf("%s deals %s damage at %v, want 1", player.Name, proto.SpellSchool(school), got)
			}
		}
		if got := done.GetHealing(); got != 1 {
			t.Errorf("%s heals at %v, want 1", player.Name, got)
		}
		if got := result.RaidStats.Parties[0].Players[0].GetShapeshiftForm(); player.Class != proto.Class_ClassWarrior && got != 0 {
			t.Errorf("%s starts in form %d, want none", player.Name, got)
		}
	}
}

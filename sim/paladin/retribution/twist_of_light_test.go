package retribution

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/paladin"
)

// Twist of Light keeps one Echo up at a time: replacing a second seal drops the first seal's Echo
// for the second's.
func TestTwistOfLightKeepsOneEcho(t *testing.T) {
	player := &proto.Player{
		Name:          "Twist of Light",
		Race:          proto.Race_RaceHuman,
		Class:         proto.Class_ClassPaladin,
		Equipment:     core.GetGearSet("../../../ui/specs/paladin/retribution/gear_sets", "default").GearSet,
		TalentsString: TwistOfLightTalents,
		Spec: &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{
			Options: &proto.RetributionPaladin_Options{ClassOptions: &proto.PaladinOptions{}},
		}},
		Rotation: core.GetAplRotation("../../../ui/specs/paladin/retribution/apls", "default").Rotation,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	pal := sim.Raid.Parties[0].Players[0].(paladin.PaladinAgent).GetPaladin()

	if !pal.Talents.TwistOfLight {
		t.Fatal("the talents do not take Twist of Light")
	}

	seal := func(rankMap spelldata.Ladder) *core.Spell {
		spell := pal.GetSpell(core.ActionID{SpellID: rankMap.Highest().ID})
		if spell == nil {
			t.Fatalf("seal %d is not registered", rankMap.Highest().ID)
		}
		return spell
	}
	command := seal(paladin.SealOfCommandRankMap)
	righteousness := seal(paladin.SealOfRighteousnessRankMap)
	fury := seal(paladin.SealOfFuryRankMap)

	echoOfCommand := pal.GetAuraByID(core.ActionID{SpellID: 1311703})
	echoOfRighteousness := pal.GetAuraByID(core.ActionID{SpellID: 1311704})
	if echoOfCommand == nil || echoOfRighteousness == nil {
		t.Fatal("the Echo auras are not registered")
	}

	target := pal.CurrentTarget
	command.SkipCastAndApplyEffects(sim, target)
	righteousness.SkipCastAndApplyEffects(sim, target)
	if !echoOfCommand.IsActive() {
		t.Fatal("replacing Seal of Command leaves no Echo of Command")
	}

	fury.SkipCastAndApplyEffects(sim, target)
	if !echoOfRighteousness.IsActive() {
		t.Error("replacing Seal of Righteousness leaves no Echo of Righteousness")
	}
	if echoOfCommand.IsActive() {
		t.Error("Echo of Command stays up beside Echo of Righteousness")
	}
}

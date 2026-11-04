package healer

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/priest"
)

func init() {
	RegisterHealerPriest()
	common.RegisterAllEffects()
}

// Pins the final character stats for each gear preset so the passives stay covered, and the healing
// of the Disc and Holy preset rotations on a raid-healing encounter. The stats rows' empty APL
// rotation and fake prepull (no SkipRotation) make them exercise a full environment reset, the path
// the UI's stats request takes.
func TestHealerPriest(t *testing.T) {
	var generators []core.TestGenerator
	for _, gearSet := range []string{"preraid"} {
		player := core.WithSpec(
			&proto.Player{
				Class:         proto.Class_ClassPriest,
				Race:          proto.Race_RaceDwarf,
				Equipment:     core.GetGearSet("../../../ui/specs/priest/healer/gear_sets", gearSet).GearSet,
				Consumables:   FullConsumes,
				Buffs:         core.FullIndividualBuffs,
				TalentsString: HolyTalents,
				Profession1:   proto.Profession_Tailoring,
				Profession2:   proto.Profession_Enchanting,
				Rotation:      &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
			},
			PlayerOptions,
		)
		generators = append(generators, &core.SingleCharacterStatsTestGenerator{
			Name: gearSet,
			Request: &proto.ComputeStatsRequest{
				Raid: core.SinglePlayerRaidProto(player, core.FullPartyBuffs, core.FullRaidBuffs, core.FullDebuffs),
			},
		})
	}
	for _, preset := range []struct{ name, build, talents string }{
		{"Disc", "disc_testing", DiscTalents},
		{"Holy", "holy_testing", HolyTalents},
	} {
		generators = append(generators, &core.SingleDpsTestGenerator{
			Name:    preset.name + "-RaidHealing10",
			Request: raidHealingRequest(preset.build, preset.talents),
		})
	}
	core.RunTestSuite(t, t.Name(), generators)
}

// A preset build's rotation, options, consumables and buffs on its encounter, in the pre-raid gear,
// with the Raid healing assignment of a 10-player group: nine target dummies, four in the priest's
// party and five in a second one.
func raidHealingRequest(buildFile string, talents string) *proto.RaidSimRequest {
	build := core.GetTestBuildFromJSON(proto.Class_ClassPriest, "../../../ui/specs/priest/healer/builds", buildFile,
		core.ItemFilter{}, nil, nil)
	player := core.WithSpec(
		&proto.Player{
			Class:              proto.Class_ClassPriest,
			Race:               build.Race,
			Equipment:          core.GetGearSet("../../../ui/specs/priest/healer/gear_sets", "preraid").GearSet,
			Consumables:        build.Consumables,
			Buffs:              build.IndividualBuffs,
			TalentsString:      talents,
			Profession1:        build.Profession1,
			Profession2:        build.Profession2,
			Rotation:           build.Rotation.Rotation,
			HealingModel:       build.HealingModel,
			DistanceFromTarget: build.StartingDistance,
			ReactionTimeMs:     build.ReactionTimeMs,
		},
		build.SpecOptions.SpecOptions,
	)
	raid := core.SinglePlayerRaidProto(player, build.PartyBuffs, build.RaidBuffs, build.Debuffs)
	raid.Parties = append(raid.Parties, &proto.Party{})
	raid.TargetDummies = 9
	return &proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  build.Encounter.Encounter,
		SimOptions: core.DefaultSimTestOptions,
	}
}

// The Holy and Disc presets from ui/specs/priest/healer/presets.ts.
var HolyTalents = "00500502103-33505003030121531"
var DiscTalents = "005005031330011531-335020030301"

// Holy Binding heals the priest for its share of a Flash Heal on another player, with no threat. A
// Flash Heal on the priest itself and the priest's other heals give nothing back. The test sits with
// the healer spec because only the healer registers Flash Heal.
func TestHolyBindingHealsThePriestForFlashHealOnOthers(t *testing.T) {
	share := spelldata.MustFind(1328569).EffectN(1).Percent()
	if share == 0 {
		t.Fatal("Holy Binding's row states no share")
	}
	selfHealID := core.ActionID{SpellID: spelldata.MustFind(1328570).ID}

	healerPlayer := func(name string) *proto.Player {
		return core.WithSpec(&proto.Player{
			Name:          name,
			Class:         proto.Class_ClassPriest,
			Race:          proto.Race_RaceDwarf,
			Equipment:     core.GetGearSet("../../../ui/specs/priest/healer/gear_sets", "preraid").GearSet,
			TalentsString: HolyBindingTalents,
			Rotation:      &proto.APLRotation{},
		}, PlayerOptions)
	}
	raid := core.SinglePlayerRaidProto(healerPlayer("Priest"), &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	raid.Parties[0].Players = append(raid.Parties[0].Players, healerPlayer("Ally"))
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	healer := sim.Raid.Parties[0].Players[0].(*HealerPriest)
	ally := sim.Raid.Parties[0].Players[1].(*HealerPriest)
	if !healer.Talents.HolyBinding {
		t.Fatal("the Holy preset takes no Holy Binding")
	}

	selfHeal := healer.GetSpell(selfHealID)
	if selfHeal == nil {
		t.Fatalf("the priest has no Holy Binding heal %v", selfHealID)
	}
	lastRankOf := func(mask int64) *core.Spell {
		var found *core.Spell
		for _, spell := range healer.Spellbook {
			if spell.ClassSpellMask == mask {
				found = spell
			}
		}
		if found == nil {
			t.Fatalf("the priest has no spell in class mask %d", mask)
		}
		return found
	}
	flashHeal, greaterHeal := lastRankOf(priest.PriestSpellFlashHeal), lastRankOf(priest.PriestSpellGreaterHeal)

	// Heal threat counts only the health a heal restores, so both players start missing health.
	for _, player := range []*HealerPriest{healer, ally} {
		player.RemoveHealth(sim, player.CurrentHealth()-1)
	}
	// The trigger heals one batch window after the Flash Heal lands.
	heal := func(spell *core.Spell, target *core.Unit) {
		spell.SkipCastAndApplyEffects(sim, target)
		until := sim.CurrentTime + core.SpellBatchWindow
		for sim.CurrentTime <= until {
			if sim.Step() {
				t.Fatal("the sim ended inside the batch window")
			}
		}
		if spell.SpellMetrics[target.UnitIndex].TotalHealing == 0 {
			t.Fatalf("%v on %s healed nothing", spell.ActionID, target.Label)
		}
	}
	threatOf := func(spell *core.Spell) float64 {
		var threat float64
		for _, enemy := range sim.Encounter.ActiveTargetUnits {
			threat += spell.SpellMetrics[enemy.UnitIndex].TotalThreat
		}
		return threat
	}
	selfHealing := func() float64 { return selfHeal.SpellMetrics[healer.UnitIndex].TotalHealing }

	heal(flashHeal, &ally.Unit)
	healed := flashHeal.SpellMetrics[ally.UnitIndex].TotalHealing
	if got, want := selfHealing(), share*healed; !core.WithinToleranceFloat64(want, got, 1e-6) {
		t.Errorf("a Flash Heal of %.2f on the ally healed the priest for %.2f, want the row's share %.2f", healed, got, want)
	}
	if threatOf(flashHeal) == 0 {
		t.Error("Flash Heal on the ally made no threat, so the threat check below proves nothing")
	}
	if got := threatOf(selfHeal); got != 0 {
		t.Errorf("Holy Binding's heal made %.2f threat, want none", got)
	}

	before := selfHealing()
	heal(flashHeal, &healer.Unit)
	heal(greaterHeal, &ally.Unit)
	if got := selfHealing(); got != before {
		t.Errorf("Flash Heal on the priest and Greater Heal on the ally healed the priest for %.2f through Holy Binding, want nothing", got-before)
	}
}

var HolyBindingTalents = "00520101133-33505003030121531"

var FullConsumes = &proto.ConsumesSpec{
	FoodId:        250066, // Soaring Pamplona
	ResourcePotId: 13444,  // Major Mana Potion
}

var PlayerOptions = &proto.Player_HealerPriest{
	HealerPriest: &proto.HealerPriest{
		Options: &proto.HealerPriest_Options{
			ClassOptions: &proto.PriestOptions{},
		},
	},
}

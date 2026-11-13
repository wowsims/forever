package holy

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/paladin"
)

// Enchant Weapon - Revelation 8217 gives a paladin buff 1323419, whose crit modifier and charge name
// Holy Light by its class flags (family 10) and not Lay on Hands. Holy Light with the buff up takes
// the crit and spends the charge; Lay on Hands leaves the charge up.
func TestRevelationReachesHolyLight(t *testing.T) {
	gear := core.GetGearSet("../../../ui/specs/paladin/holy/gear_sets", "default").GearSet
	gear.Items[proto.ItemSlot_ItemSlotMainHand].Enchant = 8217
	player := core.WithSpec(&proto.Player{
		Class:         proto.Class_ClassPaladin,
		Race:          proto.Race_RaceHuman,
		Equipment:     gear,
		TalentsString: StandardTalents,
		Rotation:      &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}, PlayerOptions)
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	pally := sim.Raid.Parties[0].Players[0].GetCharacter()

	buff := pally.GetAuraByID(core.ActionID{SpellID: 1323419})
	if buff == nil {
		t.Fatal("no Revelation buff 1323419 on the paladin")
	}
	holyLight := pally.GetSpell(core.ActionID{SpellID: paladin.HolyLightRankMap.Highest().ID})
	layOnHands := pally.GetSpell(core.ActionID{SpellID: paladin.LayOnHandsRankMap.Highest().ID})
	if holyLight == nil || layOnHands == nil {
		t.Fatal("Holy Light or Lay on Hands is not registered")
	}

	baseCrit := holyLight.HealingCritChance(&pally.Unit)
	buff.Activate(sim)
	buff.SetStacks(sim, buff.MaxStacks)
	if got := holyLight.HealingCritChance(&pally.Unit); got <= baseCrit {
		t.Errorf("Holy Light's crit chance is %v with Revelation up and %v without", got, baseCrit)
	}

	// A proc's handler runs one spell batch window after the heal, so each cast is followed by
	// stepping past that window.
	castAndStep := func(spell *core.Spell) {
		spell.SkipCastAndApplyEffects(sim, &pally.Unit)
		for at := sim.CurrentTime; sim.CurrentTime <= at+core.SpellBatchWindow; {
			sim.Step()
		}
	}

	castAndStep(layOnHands)
	if !buff.IsActive() {
		t.Fatal("Lay on Hands, which the buff's mask does not name, spent its charge")
	}

	castAndStep(holyLight)
	if buff.IsActive() {
		t.Error("Holy Light did not spend Revelation's charge")
	}
}

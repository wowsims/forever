package protection

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/paladin"
)

// A Holy Strike that lands puts up Shieldsworn's 6 sec spell block. While it is up a Fire swing
// rolls the paladin's block chance, and a blocked Fire swing is a block to Holy Shield, which spends
// a charge, and to Shield Specialization, which restores mana.
func TestShieldswornBlocksNonPhysicalSwings(t *testing.T) {
	encounter := core.MakeSingleTargetEncounter(0)
	encounter.Targets[0].SpellSchool = proto.SpellSchool_SpellSchoolFire

	player := core.WithSpec(&proto.Player{
		Class:         proto.Class_ClassPaladin,
		Race:          proto.Race_RaceHuman,
		Equipment:     core.GetGearSet("../../../ui/specs/paladin/protection/gear_sets", "default").GearSet,
		TalentsString: DefaultProtectionTalents,
		Rotation:      &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}, DefaultOptions)
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	pally := sim.Raid.Parties[0].Players[0].(paladin.PaladinAgent).GetPaladin()
	target := sim.Encounter.ActiveTargetUnits[0]
	if !pally.Talents.Shieldsworn || !pally.Talents.HolyShield || pally.Talents.ShieldSpecialization != 2 {
		t.Fatal("the default protection talents lack Shieldsworn, Holy Shield or Shield Specialization 2/2")
	}

	swing := target.AutoAttacks.MHAuto()
	spellBlock := pally.GetAuraByID(core.ActionID{SpellID: 1328551})
	holyShield := pally.GetAuraByID(core.ActionID{SpellID: paladin.HolyShieldRankMap.Highest().ID})
	holyStrike := pally.GetSpell(core.ActionID{SpellID: paladin.HolyStrikeRankMap.Highest().ID})
	if swing == nil || spellBlock == nil || holyShield == nil || holyStrike == nil {
		t.Fatal("the boss's swing, Shieldsworn's spell block, Holy Shield or Holy Strike is not registered")
	}
	if swing.SpellSchool != core.SpellSchoolFire {
		t.Fatalf("the boss swings for school %d, not Fire", swing.SpellSchool)
	}
	blocks := func() int32 {
		return swing.SpellMetrics[pally.UnitIndex].Blocks
	}

	holyShield.Activate(sim)
	holyShield.SetStacks(sim, holyShield.MaxStacks)

	for range 50 {
		swing.SkipCastAndApplyEffects(sim, &pally.Unit)
	}
	if blocks() != 0 {
		t.Fatalf("%d of 50 Fire swings were blocked without Shieldsworn's spell block up", blocks())
	}

	for range 50 {
		holyStrike.SkipCastAndApplyEffects(sim, target)
		if spellBlock.IsActive() {
			break
		}
	}
	if !spellBlock.IsActive() {
		t.Fatal("no Holy Strike in 50 put up Shieldsworn's spell block")
	}
	if holyShield.GetStacks() != holyShield.MaxStacks {
		t.Fatalf("a Holy Strike left Holy Shield at %d of %d charges", holyShield.GetStacks(), holyShield.MaxStacks)
	}

	// Mana is drained first so Shield Specialization's restore has room to land.
	pally.SpendMana(sim, pally.CurrentMana()/2, pally.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionNone}))
	manaBefore := pally.CurrentMana()

	for range 50 {
		swing.SkipCastAndApplyEffects(sim, &pally.Unit)
		if blocks() > 0 {
			break
		}
	}
	if blocks() == 0 {
		t.Fatal("none of 50 Fire swings were blocked with Shieldsworn's spell block up")
	}
	if holyShield.GetStacks() != holyShield.MaxStacks-1 {
		t.Fatalf("a blocked Fire swing left Holy Shield at %d of %d charges", holyShield.GetStacks(), holyShield.MaxStacks)
	}
	restore := pally.MaxMana() * spelldata.MustFind(1310925).EffectN(1).Percent()
	if got := pally.CurrentMana() - manaBefore; math.Abs(got-restore) > 1e-6 {
		t.Fatalf("a blocked Fire swing restored %.2f mana, want Shield Specialization's %.2f", got, restore)
	}
}

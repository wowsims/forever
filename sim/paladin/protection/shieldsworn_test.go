package protection

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/paladin"
)

// A Holy Strike puts up Shieldsworn's 6 sec spell block, whether or not it lands. While it is up a Fire swing
// rolls the paladin's block chance, and a blocked Fire swing is a block to Holy Shield, which spends
// a charge, and to Shield Specialization, which restores mana.
func TestShieldswornBlocksNonPhysicalSwings(t *testing.T) {
	encounter := core.MakeSingleTargetEncounter(0)
	encounter.Targets[0].SpellSchool = proto.SpellSchool_SpellSchoolFire
	gear := core.GetGearSet("../../../ui/specs/paladin/protection/gear_sets", "default").GearSet
	sim, pally, target := shieldswornPaladin(gear, encounter)
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

	// A Holy Strike's shield damage reports under its own spell 456544, and a block's under Holy Shield's
	// tag 2.
	blockDamage := pally.GetSpell(holyShield.ActionID.WithTag(2))
	strikeDamage := pally.GetSpell(core.ActionID{SpellID: 456544})
	if blockDamage == nil || strikeDamage == nil {
		t.Fatal("Holy Shield's block damage or its Shieldsworn copy is not registered")
	}

	// Every strike puts the spell block up, a dodged or missed one too, and only one that lands deals the
	// shield's damage.
	landed, avoided := 0, 0
	for casts := 0; casts < 5000 && (landed == 0 || avoided == 0); casts++ {
		spellBlock.Deactivate(sim)
		strikes, shieldCasts := holyStrike.SpellMetrics[target.UnitIndex], strikeDamage.SpellMetrics[target.UnitIndex].Casts
		holyStrike.SkipCastAndApplyEffects(sim, target)
		after := holyStrike.SpellMetrics[target.UnitIndex]
		wasAvoided := after.Misses+after.Dodges+after.Parries > strikes.Misses+strikes.Dodges+strikes.Parries
		if !spellBlock.IsActive() {
			t.Fatalf("a Holy Strike put up no spell block (avoided: %t)", wasAvoided)
		}
		if dealt := strikeDamage.SpellMetrics[target.UnitIndex].Casts > shieldCasts; dealt == wasAvoided {
			t.Fatalf("a Holy Strike dealt the shield's damage: %t (avoided: %t)", dealt, wasAvoided)
		}
		if wasAvoided {
			avoided++
		} else {
			landed++
		}
	}
	if landed == 0 || avoided == 0 {
		t.Fatalf("%d Holy Strikes landed and %d were avoided, want both", landed, avoided)
	}
	if holyShield.GetStacks() != holyShield.MaxStacks {
		t.Fatalf("a Holy Strike left Holy Shield at %d of %d charges", holyShield.GetStacks(), holyShield.MaxStacks)
	}
	if blockDamage.SpellMetrics[target.UnitIndex].Casts != 0 {
		t.Fatalf("a Holy Strike's shield damage was cast %d times under the block's tag", blockDamage.SpellMetrics[target.UnitIndex].Casts)
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

// Shieldsworn's strike 456544 rolls Enchant Weapon - Revelation 8217 at the share of spell level 30 under a
// level 60 wearer, a tenth of the chance, and Holy Shield's damage on a block rolls nothing (EV-0161).
func TestShieldswornStrikeRollsRevelationAtItsSpellLevel(t *testing.T) {
	gear := core.GetGearSet("../../../ui/specs/paladin/protection/gear_sets", "default").GearSet
	gear.Items[proto.ItemSlot_ItemSlotMainHand].Enchant = 8217
	sim, pally, target := shieldswornPaladin(gear, core.MakeSingleTargetEncounter(0))

	revelation := pally.GetAuraByID(core.ActionID{SpellID: 1323419})
	holyShield := pally.GetAuraByID(core.ActionID{SpellID: paladin.HolyShieldRankMap.Highest().ID})
	holyStrike := pally.GetSpell(core.ActionID{SpellID: paladin.HolyStrikeRankMap.Highest().ID})
	strike := pally.GetSpell(core.ActionID{SpellID: 456544})
	if revelation == nil || holyShield == nil || holyStrike == nil || strike == nil {
		t.Fatal("Revelation's buff, Holy Shield, Holy Strike or Shieldsworn's strike is not registered")
	}
	if level := spelldata.MustFind(456544).SpellLevel; level != 30 || pally.Level != 60 {
		t.Fatalf("the strike reads spell level %d under a level %d paladin, want 30 under 60", level, pally.Level)
	}

	// A proc's handler runs one spell batch window after the hit.
	stepPast := func() {
		reached := false
		sim.AddPendingAction(core.NewDelayedAction(core.DelayedActionOptions{
			DoAt:     sim.CurrentTime + core.SpellBatchWindow + 1,
			OnAction: func(*core.Simulation) { reached = true },
		}))
		for steps := 0; !reached; steps++ {
			if steps > 10000 {
				t.Fatal("the sim did not reach the end of the spell batch window")
			}
			sim.Step()
		}
	}
	rolls := func(hit func()) int {
		procs := 0
		for range 6000 {
			revelation.Deactivate(sim)
			holyShield.Activate(sim)
			holyShield.SetStacks(sim, holyShield.MaxStacks)
			hit()
			stepPast()
			if revelation.IsActive() {
				procs++
			}
		}
		return procs
	}

	strikeProcs := rolls(func() { holyStrike.SkipCastAndApplyEffects(sim, target) })
	landed := strike.SpellMetrics[target.UnitIndex].Hits
	if landed < 4000 || strikeProcs == 0 || float64(strikeProcs)/float64(landed) > 0.02 {
		t.Errorf("%d Shieldsworn strikes landed and procced Revelation %d times, want some procs at a tenth of its chance", landed, strikeProcs)
	}

	swing := target.AutoAttacks.MHAuto()
	if blockProcs := rolls(func() {
		pally.OnSpellHitTaken(sim, swing, &core.SpellResult{Target: &pally.Unit, Outcome: core.OutcomeBlock, Damage: 100})
	}); blockProcs != 0 {
		t.Errorf("Holy Shield's damage on 6000 blocks procced Revelation %d times, want none", blockProcs)
	}
}

// Shieldsworn's strike 456544 crits for 1.5 times, and Holy Shield's damage on a block never crits
// (EV-0189).
func TestShieldswornStrikeCritsAndBlockDamageDoesNot(t *testing.T) {
	gear := core.GetGearSet("../../../ui/specs/paladin/protection/gear_sets", "default").GearSet
	sim, pally, target := shieldswornPaladin(gear, core.MakeSingleTargetEncounter(0))

	holyShield := pally.GetAuraByID(core.ActionID{SpellID: paladin.HolyShieldRankMap.Highest().ID})
	holyStrike := pally.GetSpell(core.ActionID{SpellID: paladin.HolyStrikeRankMap.Highest().ID})
	strike := pally.GetSpell(core.ActionID{SpellID: 456544})
	if holyShield == nil || holyStrike == nil || strike == nil {
		t.Fatal("Holy Shield, Holy Strike or Shieldsworn's strike is not registered")
	}
	blockDamage := pally.GetSpell(holyShield.ActionID.WithTag(2))
	if blockDamage == nil {
		t.Fatal("Holy Shield's block damage is not registered")
	}
	pally.AddStatDynamic(sim, stats.SpellCritPercent, 50)

	swing := target.AutoAttacks.MHAuto()
	for range 2000 {
		holyShield.Activate(sim)
		holyShield.SetStacks(sim, holyShield.MaxStacks)
		holyStrike.SkipCastAndApplyEffects(sim, target)
		pally.OnSpellHitTaken(sim, swing, &core.SpellResult{Target: &pally.Unit, Outcome: core.OutcomeBlock, Damage: 100})
	}

	strikes := strike.SpellMetrics[target.UnitIndex]
	if strikes.Hits == 0 || strikes.Crits == 0 {
		t.Fatalf("Shieldsworn's strike landed %d hits and %d crits, want both", strikes.Hits, strikes.Crits)
	}
	hit := (strikes.TotalDamage - strikes.TotalCritDamage) / float64(strikes.Hits)
	if got := strikes.TotalCritDamage / float64(strikes.Crits) / hit; math.Abs(got-1.5) > 1e-6 {
		t.Errorf("a crit of Shieldsworn's strike deals %v times a hit, want 1.5", got)
	}
	if blocks := blockDamage.SpellMetrics[target.UnitIndex]; blocks.Hits == 0 || blocks.Crits != 0 {
		t.Errorf("Holy Shield's damage on a block landed %d hits and %d crits, want hits and no crits", blocks.Hits, blocks.Crits)
	}
}

// A protection paladin with the default talents in the given gear, ready to act by hand, and its target.
func shieldswornPaladin(gear *proto.EquipmentSpec, encounter *proto.Encounter) (*core.Simulation, *paladin.Paladin, *core.Unit) {
	player := core.WithSpec(&proto.Player{
		Class:         proto.Class_ClassPaladin,
		Race:          proto.Race_RaceHuman,
		Equipment:     gear,
		TalentsString: DefaultProtectionTalents,
		Rotation:      &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}, DefaultOptions)
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(paladin.PaladinAgent).GetPaladin(), sim.Encounter.ActiveTargetUnits[0]
}

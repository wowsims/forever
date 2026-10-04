package core

import (
	"math"
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

func setupRacialSim(race proto.Race) (*Simulation, *FakeRageWarrior) {
	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{
			RandomSeed: 100,
		},
		Raid: &proto.Raid{
			Parties: []*proto.Party{
				{
					Players: []*proto.Player{
						{
							Name:      "Warrior",
							Class:     proto.Class_ClassWarrior,
							Race:      race,
							Buffs:     &proto.IndividualBuffs{},
							Spec:      &proto.Player_DpsWarrior{},
							Equipment: &proto.EquipmentSpec{},
							Rotation:  &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
						},
					},
					Buffs: &proto.PartyBuffs{},
				},
			},
		},
		Encounter: &proto.Encounter{
			Targets: []*proto.Target{
				{Name: "target", Level: 63, MobType: proto.MobType_MobTypeElemental},
			},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()

	return sim, sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
}

func castRacial(t *testing.T, sim *Simulation, fw *FakeRageWarrior, spellID int32) {
	t.Helper()
	spell := fw.GetSpell(ActionID{SpellID: spellID})
	if spell == nil {
		t.Fatalf("Racial %d is not registered", spellID)
	}
	if !spell.Cast(sim, fw.CurrentTarget) {
		t.Fatalf("Racial %d could not be cast", spellID)
	}
}

func TestPlayableRaces(t *testing.T) {
	for class, races := range ClassRaceCapabilities {
		for _, race := range races {
			if race == proto.Race_RaceBloodElf {
				t.Errorf("%s is not playable, but %s lists it", race, class)
			}
			if _, ok := BaseStats[BaseStatsKey{Race: race, Class: class}]; !ok {
				t.Errorf("No base stats for %s %s", race, class)
			}
		}
	}

	for class, race := range map[proto.Class]proto.Race{
		proto.Class_ClassPaladin: proto.Race_RaceUndead,
		proto.Class_ClassShaman:  proto.Race_RaceDwarf,
		proto.Class_ClassHunter:  proto.Race_RaceHuman,
		proto.Class_ClassMage:    proto.Race_RaceOrc,
		proto.Class_ClassPriest:  proto.Race_RaceGnome,
		proto.Class_ClassWarlock: proto.Race_RaceTroll,
		proto.Class_ClassDruid:   proto.Race_RaceHighOrderSkyborne,
	} {
		if !slices.Contains(ClassRaceCapabilities[class], race) {
			t.Errorf("%s should be playable as %s", class, race)
		}
	}
	if slices.Contains(ClassRaceCapabilities[proto.Class_ClassShaman], proto.Race_RaceHighOrderSkyborne) {
		t.Errorf("Only the Windshaper Skyborne can be shamans")
	}
	if slices.Contains(ClassRaceCapabilities[proto.Class_ClassMage], proto.Race_RaceWindshaperSkyborne) {
		t.Errorf("Only the High Order Skyborne can be mages")
	}
}

func TestBloodFuryIsPercentBased(t *testing.T) {
	sim, fw := setupRacialSim(proto.Race_RaceOrc)

	apBefore := fw.GetStat(stats.AttackPower)
	castRacial(t, sim, fw, 20572)

	if !WithinToleranceFloat64(apBefore*1.1, fw.GetStat(stats.AttackPower), 0.01) {
		t.Fatalf("Blood Fury should grant 10%% Attack Power: before %0.2f, after %0.2f", apBefore, fw.GetStat(stats.AttackPower))
	}
}

func TestBerserkingIsFlatTenPercent(t *testing.T) {
	sim, fw := setupRacialSim(proto.Race_RaceTroll)

	speedBefore := fw.PseudoStats.AttackSpeedMultiplier
	castRacial(t, sim, fw, 20554)

	if !WithinToleranceFloat64(speedBefore*1.1, fw.PseudoStats.AttackSpeedMultiplier, 0.0001) {
		t.Fatalf("Berserking should grant 10%% attack speed, got x%0.4f", fw.PseudoStats.AttackSpeedMultiplier/speedBefore)
	}
	if fw.CurrentRage() != 0 || fw.GetSpell(ActionID{SpellID: 20554}).Cost != nil {
		t.Fatalf("Berserking should have no cost")
	}
}

func TestElunesLight(t *testing.T) {
	sim, fw := setupRacialSim(proto.Race_RaceNightElf)

	critBefore := fw.GetStat(stats.PhysicalCritPercent)
	castRacial(t, sim, fw, 1259799)

	if !WithinToleranceFloat64(critBefore+10, fw.GetStat(stats.PhysicalCritPercent), 0.0001) {
		t.Fatalf("Elune's Light should grant 10%% crit: before %0.2f, after %0.2f", critBefore, fw.GetStat(stats.PhysicalCritPercent))
	}
}

// A warrior ability as spelldata builds one: the game files' family 4 flags and no ClassSpellMask.
func fakeWarriorAbility(id int32, bit uint, damage float64) SpellConfig {
	return SpellConfig{
		ActionID:         ActionID{SpellID: id},
		ClassFlags:       ClassFlags{Family: 4, Mask: [4]uint32{1 << bit}},
		ProcMask:         ProcMaskMeleeMHSpecial,
		SpellSchool:      SpellSchoolPhysical,
		RageCost:         RageCostOptions{Cost: 30},
		DamageMultiplier: 1,
		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeAlwaysHit)
		},
	}
}

var fakeRendActionID = ActionID{SpellID: 772}

// A DoT's aura has to be registered before the environment is finalized.
func (fw *FakeRageWarrior) registerFakeRend() {
	config := fakeWarriorAbility(fakeRendActionID.SpellID, 5, 0)
	config.Dot = DotConfig{
		Aura:          Aura{Label: "Rend"},
		NumberOfTicks: 5,
		TickLength:    time.Second * 3,
		OnTick: func(sim *Simulation, target *Unit, dot *Dot) {
			dot.Spell.CalcAndDealPeriodicDamage(sim, target, 50, dot.OutcomeTick)
		},
	}
	config.ApplyEffects = func(sim *Simulation, target *Unit, spell *Spell) {
		if spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit).Landed() {
			spell.Dot(target).Apply(sim)
		}
	}
	fw.RegisterSpell(config)
}

func TestGnomeWarrior(t *testing.T) {
	sim, fw := setupRacialSim(proto.Race_RaceGnome)
	fw.AutoAttacks.CancelAutoSwing(sim)

	if fw.MaximumRage() != 105 {
		t.Fatalf("Expansive Mind should raise maximum Rage to 105, got %0.1f", fw.MaximumRage())
	}

	castRacial(t, sim, fw, 1259813)
	eureka := fw.GetAura("Eureka!")
	if !eureka.IsActive() || eureka.GetStacks() != 3 {
		t.Fatalf("Eureka! should start at 3 charges, got %d", eureka.GetStacks())
	}

	mortalStrike := fw.RegisterSpell(fakeWarriorAbility(12294, 25, 100))
	hamstring := fw.RegisterSpell(fakeWarriorAbility(1715, 1, 0))
	sunderArmor := fw.RegisterSpell(fakeWarriorAbility(7386, 14, 0))
	rend := fw.GetSpell(fakeRendActionID)
	itemSpell := fw.RegisterSpell(SpellConfig{
		ActionID:         ActionID{SpellID: 2},
		Flags:            SpellFlagAPL,
		ProcMask:         ProcMaskMeleeMHSpecial,
		SpellSchool:      SpellSchoolPhysical,
		DamageMultiplier: 1,
		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeAlwaysHit)
		},
	})

	if !WithinToleranceFloat64(0.1, mortalStrike.DirectDamageMultiplierAdditive, 0.0001) {
		t.Fatalf("Eureka! should raise Mortal Strike's direct damage by 10%%, got +%0.2f", mortalStrike.DirectDamageMultiplierAdditive)
	}
	if !WithinToleranceFloat64(27, mortalStrike.Cost.GetCurrentCost(), 0.0001) {
		t.Fatalf("Eureka! should cut Mortal Strike's 30 Rage by 10%%, got %0.1f", mortalStrike.Cost.GetCurrentCost())
	}
	if sunderArmor.DirectDamageMultiplierAdditive != 0 || sunderArmor.Cost.GetCurrentCost() != 30 {
		t.Fatalf("Eureka! should leave Sunder Armor alone, which its class mask does not name")
	}

	// The spend lands one spell batch after the hit.
	stacksAfter := func(spell *Spell) int32 {
		spell.SkipCastAndApplyEffects(sim, fw.CurrentTarget)
		for end := sim.CurrentTime + SpellBatchWindow; sim.CurrentTime <= end && !sim.Step(); {
		}
		return eureka.GetStacks()
	}

	if stacksAfter(itemSpell) != 3 || itemSpell.DirectDamageMultiplierAdditive != 0 {
		t.Fatalf("Eureka! should ignore spells that are not class abilities")
	}
	if stacksAfter(sunderArmor) != 3 {
		t.Fatalf("Sunder Armor should not spend a charge")
	}
	if stacksAfter(hamstring) != 3 {
		t.Fatalf("an ability the mask names should not spend a charge when it deals no damage")
	}
	if stacksAfter(rend) != 3 {
		t.Fatalf("Rend should not spend a charge: it is periodic, and the class mask does not name it")
	}
	if stacksAfter(mortalStrike) != 2 {
		t.Fatalf("Mortal Strike should spend a charge")
	}
	// An off-hand strike on the same spell id, as Whirlwind's is, lands with the main-hand one and
	// spends no second charge.
	offHandConfig := fakeWarriorAbility(12294, 25, 100)
	offHandConfig.ActionID = offHandConfig.ActionID.WithTag(2)
	offHandConfig.ProcMask = ProcMaskMeleeOHSpecial
	offHandStrike := fw.RegisterSpell(offHandConfig)
	mortalStrike.SkipCastAndApplyEffects(sim, fw.CurrentTarget)
	if stacksAfter(offHandStrike) != 1 {
		t.Fatalf("a main- and off-hand strike of one cast should spend one charge, %d left", eureka.GetStacks())
	}
	stacksAfter(mortalStrike)
	if eureka.IsActive() {
		t.Fatalf("Eureka! should end after 3 abilities, %d left at %v", eureka.GetStacks(), sim.CurrentTime)
	}
	if !WithinToleranceFloat64(0, mortalStrike.DirectDamageMultiplierAdditive, 0.0001) {
		t.Fatalf("Eureka! should drop its damage bonus when it ends, got +%0.2f", mortalStrike.DirectDamageMultiplierAdditive)
	}
}

func TestGnomeEurekaReadsEachClassRow(t *testing.T) {
	for spellID, want := range map[int32]EurekaMods{
		1259812: {
			CostFlags:   ClassFlags{Family: 8, Mask: [4]uint32{100794910, 2097408}},
			DamageFlags: ClassFlags{Family: 8, Mask: [4]uint32{100794910, 2097408}},
			Cost:        -0.1, Damage: 0.1,
		},
		1259813: {
			CostFlags:   ClassFlags{Family: 4, Mask: [4]uint32{1784679630, 5, 1}},
			DamageFlags: ClassFlags{Family: 4, Mask: [4]uint32{1784679630, 5, 1}},
			Cost:        -0.1, Damage: 0.1,
		},
		1259817: {
			CostFlags:   ClassFlags{Family: 3, Mask: [4]uint32{549591799, 4096}},
			DamageFlags: ClassFlags{Family: 3, Mask: [4]uint32{551688947, 4096, 0, 8}},
			Cost:        -0.1, Damage: 0.1, Channel: 0.1,
		},
		1259821: {
			CostFlags:    ClassFlags{Family: 5, Mask: [4]uint32{541165, 8650880}},
			DamageFlags:  ClassFlags{Family: 5, Mask: [4]uint32{524773, 8388736}},
			ChannelFlags: ClassFlags{Family: 5, Mask: [4]uint32{16392, 262144}},
			Cost:         -0.1, Damage: 0.1, Channel: 0.1,
		},
		1259823: {
			CostFlags:    ClassFlags{Family: 6, Mask: [4]uint32{150224528, 8388614}},
			DamageFlags:  ClassFlags{Family: 6, Mask: [4]uint32{150224528, 8486918}},
			ChannelFlags: ClassFlags{Family: 6, Mask: [4]uint32{10485760, 98304}},
			Cost:         -0.1, Damage: 0.1, Channel: 0.1,
		},
	} {
		if got := eurekaMods(spellID); got != want {
			t.Errorf("Eureka! %d reads %+v, want %+v", spellID, got, want)
		}
	}
}

// Eureka!'s idx2 raises the ticks of a warlock channel its mask names that spends a stack, on all of
// its ticks, also after the last stack is gone. A channel cast with no stack up takes no bonus, nor
// does a periodic spell that is not a channel, on the mask or off it.
func TestGnomeWarlockEurekaChannel(t *testing.T) {
	sim, fw := setupFakeWarlockSim(proto.Race_RaceGnome)
	target := fw.CurrentTarget

	eurekaSpell := fw.GetSpell(ActionID{SpellID: 1259821})
	if eurekaSpell == nil || !eurekaSpell.Cast(sim, target) {
		t.Fatalf("a gnome warlock should cast Eureka! 1259821")
	}
	eureka := fw.GetAura("Eureka!")

	multiplier := func(spell *Spell) float64 { return spell.Dot(target).PeriodicDamageMultiplier }
	drainLife := fw.DrainLife.Dot(target)
	var tickMultipliers []float64
	onTick := drainLife.onTick
	drainLife.onTick = func(sim *Simulation, target *Unit, dot *Dot) {
		tickMultipliers = append(tickMultipliers, dot.PeriodicDamageMultiplier)
		onTick(sim, target, dot)
	}
	castAndSpend := func(spell *Spell) {
		tickMultipliers = nil
		spell.SkipCastAndApplyEffects(sim, target)
		for end := sim.CurrentTime + SpellBatchWindow; sim.CurrentTime <= end && !sim.Step(); {
		}
	}
	channelTicks := func() []float64 {
		for drainLife.IsActive() && !sim.Step() {
		}
		return tickMultipliers
	}

	if got := multiplier(fw.DrainLife); got != 1 {
		t.Fatalf("Eureka! should leave Drain Life alone until a cast spends a stack, got x%0.2f", got)
	}
	castAndSpend(fw.DrainLife)
	if eureka.GetStacks() != 2 {
		t.Fatalf("Drain Life's landed cast should spend a stack, %d left", eureka.GetStacks())
	}
	if got := multiplier(fw.DrainLife); !WithinToleranceFloat64(1.1, got, 0.0001) {
		t.Fatalf("Eureka! should raise the ticks of the Drain Life that spent a stack by 10%%, got x%0.2f", got)
	}
	castAndSpend(fw.Haunt)
	if got := multiplier(fw.Haunt); got != 1 {
		t.Fatalf("Eureka! should leave a periodic spell that is not a channel alone, even on its mask, got x%0.2f", got)
	}
	if got := multiplier(fw.Corruption); got != 1 {
		t.Fatalf("Eureka! should leave Corruption's ticks alone, got x%0.2f", got)
	}

	castAndSpend(fw.DrainLife)
	if eureka.IsActive() {
		t.Fatalf("Drain Life should spend the last stack, %d left", eureka.GetStacks())
	}
	ticks := channelTicks()
	if len(ticks) != 5 {
		t.Fatalf("Drain Life should tick 5 times, got %d", len(ticks))
	}
	for i, got := range ticks {
		if !WithinToleranceFloat64(1.1, got, 0.0001) {
			t.Fatalf("the Drain Life that spent the last stack should keep x1.1 on every tick, tick %d got x%0.2f", i+1, got)
		}
	}
	if got := multiplier(fw.DrainLife); !WithinToleranceFloat64(1, got, 0.0001) {
		t.Fatalf("Drain Life should drop the bonus when its channel ends, got x%0.2f", got)
	}

	castAndSpend(fw.DrainLife)
	ticks = channelTicks()
	if len(ticks) != 5 {
		t.Fatalf("Drain Life should tick 5 times, got %d", len(ticks))
	}
	for i, got := range ticks {
		if !WithinToleranceFloat64(1, got, 0.0001) {
			t.Fatalf("a Drain Life cast after Eureka! ended should take no bonus, tick %d got x%0.2f", i+1, got)
		}
	}
}

func TestTaurenEndurance(t *testing.T) {
	_, human := setupRacialSim(proto.Race_RaceHuman)
	_, tauren := setupRacialSim(proto.Race_RaceTauren)

	if !WithinToleranceFloat64(human.GetStat(stats.PhysicalHitPercent)+1, tauren.GetStat(stats.PhysicalHitPercent), 0.0001) {
		t.Fatalf("Endurance should grant 1%% hit")
	}
}

func TestTouchOfTheGraveUsesTheMeleeVariant(t *testing.T) {
	_, fw := setupRacialSim(proto.Race_RaceUndead)

	if aura := fw.GetAura("Touch of the Grave"); aura == nil || aura.ActionIDForProc.SpellID != 1260189 {
		t.Fatalf("An undead warrior should carry the 5%% Touch of the Grave")
	} else if aura.OnSpellHitDealt == nil {
		t.Fatalf("Touch of the Grave listens for no hit, so it never procs")
	}
	if fw.GetSpell(ActionID{SpellID: 1260198}) == nil {
		t.Fatalf("Touch of the Grave's drain is not registered")
	}
}

func TestSkyborneRacials(t *testing.T) {
	_, human := setupRacialSim(proto.Race_RaceHuman)

	for _, race := range []proto.Race{proto.Race_RaceHighOrderSkyborne, proto.Race_RaceWindshaperSkyborne} {
		_, fw := setupRacialSim(race)

		swingRatio := human.AutoAttacks.MainhandSwingSpeed().Seconds() / fw.AutoAttacks.MainhandSwingSpeed().Seconds()
		if !WithinToleranceFloat64(1.01, swingRatio, 0.001) {
			t.Errorf("%s: Wind Blessed should swing 1%% faster, got x%0.4f", race, swingRatio)
		}
		if at := fw.AttackTables[fw.CurrentTarget.UnitIndex]; !WithinToleranceFloat64(1.05, at.DamageDealtMultiplier, 0.0001) {
			t.Errorf("%s: Elemental Insight should grant 5%% damage against elementals, got x%0.4f", race, at.DamageDealtMultiplier)
		}

		hasReadLeyLine := fw.GetSpell(ActionID{SpellID: 1259705}) != nil
		hasSkysight := fw.GetSpell(ActionID{SpellID: 1259686}) != nil
		if hasReadLeyLine != (race == proto.Race_RaceHighOrderSkyborne) || hasSkysight != (race == proto.Race_RaceWindshaperSkyborne) {
			t.Errorf("%s: Read Ley Line is the High Order's and Skysight the Windshapers', got %t / %t", race, hasReadLeyLine, hasSkysight)
		}
	}
}

// Blood Fury's multipliers report the stats they add on gain and the stats they remove on expire.
func TestBloodFuryReportsItsTemporaryStats(t *testing.T) {
	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name:      "Orc",
				Class:     proto.Class_ClassWarrior,
				Race:      proto.Race_RaceOrc,
				Buffs:     &proto.IndividualBuffs{},
				Spec:      &proto.Player_DpsWarrior{},
				Equipment: &proto.EquipmentSpec{},
				Rotation:  &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{
			Targets:  []*proto.Target{{Name: "target", Level: 63, MobType: proto.MobType_MobTypeHumanoid}},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()

	character := sim.Raid.Parties[0].Players[0].GetCharacter()
	bloodFury := character.GetAura("Blood Fury")
	if bloodFury == nil {
		t.Fatal("no Blood Fury aura on an Orc")
	}

	var heard []stats.Stats
	character.OnTemporaryStatsChanges = append(character.OnTemporaryStatsChanges, func(_ *Simulation, aura *Aura, change stats.Stats) {
		if aura != bloodFury {
			t.Errorf("callback names aura %s, want Blood Fury", aura.Label)
		}
		heard = append(heard, change)
	})

	before := character.GetStats()
	bloodFury.Activate(sim)
	during := character.GetStats()
	bloodFury.Deactivate(sim)
	after := character.GetStats()

	if len(heard) != 2 {
		t.Fatalf("callback heard %d changes, want one on gain and one on expire", len(heard))
	}
	if heard[0] != during.Subtract(before) {
		t.Errorf("gain reported %s, want the stats the multipliers added: %s",
			heard[0].FlatString(), during.Subtract(before).FlatString())
	}
	if heard[1] != after.Subtract(during) {
		t.Errorf("expire reported %s, want the stats the multipliers removed: %s",
			heard[1].FlatString(), after.Subtract(during).FlatString())
	}
	if after != before {
		t.Errorf("stats after expire %s, want the ones before gain %s", after.FlatString(), before.FlatString())
	}

	attackPower := before[stats.AttackPower]
	if attackPower <= 0 {
		t.Fatalf("attack power is %v before Blood Fury, nothing for 10%% to show on", attackPower)
	}
	if got := heard[0][stats.AttackPower]; math.Abs(got-attackPower*0.1) > 1 {
		t.Errorf("gain reported %v attack power, want 10%% of %v", got, attackPower)
	}
	if got := heard[1][stats.AttackPower]; got != -heard[0][stats.AttackPower] {
		t.Errorf("expire reported %v attack power, want the reverse of the gain's %v", got, heard[0][stats.AttackPower])
	}
}

// A weapon specialization follows the weapons through an item swap in combat: a human's Sword
// Specialization and an orc's Axe Specialization move the crit of the sheet, of spells and of every
// auto when the main hand changes between a sword and an axe, and move it back on the swap back.
func TestWeaponSpecializationFollowsAnItemSwapInCombat(t *testing.T) {
	for _, tc := range []struct {
		race      proto.Race
		swordCrit float64
		axeCrit   float64
	}{
		{race: proto.Race_RaceHuman, swordCrit: 2},
		{race: proto.Race_RaceOrc, axeCrit: 1},
	} {
		t.Run(tc.race.String(), func(t *testing.T) {
			withSword := WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, WeaponTestBowID)
			withAxe := WeaponTestSlots(weaponTestAxeID, WeaponTestDaggerID, WeaponTestBowID)

			finalStats := func(equipped, swapped []*proto.ItemSpec) stats.Stats {
				result := ComputeStats(&proto.ComputeStatsRequest{Raid: weaponTestRaid(tc.race, equipped, swapped), Encounter: weaponTestEncounter()})
				return stats.FromUnitStatsProto(result.RaidStats.Parties[0].Players[0].FinalStats)
			}
			startSword, startAxe := finalStats(withSword, withAxe), finalStats(withAxe, withSword)
			for _, stat := range []stats.Stat{stats.PhysicalCritPercent, stats.SpellCritPercent} {
				if got := startSword[stat] - startAxe[stat]; !WithinToleranceFloat64(tc.swordCrit-tc.axeCrit, got, 1e-9) {
					t.Errorf("starting with the sword over the axe moves the sheet's %s by %v, want %v", stat.StatName(), got, tc.swordCrit-tc.axeCrit)
				}
			}

			sim, character := NewWeaponTestSim(t, tc.race, withSword, withAxe, nil)
			startCrit := stats.Stats{
				stats.PhysicalCritPercent: character.GetStat(stats.PhysicalCritPercent),
				stats.SpellCritPercent:    character.GetStat(stats.SpellCritPercent),
			}
			sim.PrePull()
			for sim.CurrentTime < 3*time.Second {
				sim.Step()
			}

			target := sim.Encounter.ActiveTargetUnits[0]
			attackTable := character.AttackTables[target.UnitIndex]
			spell := &Spell{Unit: &character.Unit, ProcMask: ProcMaskSpellDamage}
			// Lifts every chance clear of the level 63 target's crit suppression, which floors it at 0.
			character.AddStatsDynamic(sim, stats.Stats{stats.PhysicalCritPercent: 20, stats.SpellCritPercent: 20})

			critChances := func() map[string]float64 {
				chances := map[string]float64{
					"physical crit stat": character.GetStat(stats.PhysicalCritPercent) / 100,
					"spell crit stat":    character.GetStat(stats.SpellCritPercent) / 100,
					"spell":              spell.SpellCritChance(target),
				}
				for name, auto := range autos(character) {
					if auto.BonusCritPercent != 0 {
						t.Errorf("the %s carries %v%% crit of its own, want none", name, auto.BonusCritPercent)
					}
					chances[name] = auto.PhysicalCritChance(attackTable)
				}
				return chances
			}
			check := func(when string, weaponType proto.WeaponType, before, after map[string]float64, want float64) {
				t.Helper()
				if got := character.MainHand().WeaponType; got != weaponType {
					t.Fatalf("%s: main hand holds a %s, want a %s", when, got, weaponType)
				}
				for name := range before {
					if got := after[name] - before[name]; !WithinToleranceFloat64(want/100, got, 1e-9) {
						t.Errorf("%s at %v: the %s moved by %v, want %v", when, sim.CurrentTime, name, got, want/100)
					}
				}
			}

			sword := critChances()
			if sim.CurrentTime <= 0 {
				t.Fatalf("the swap is at %v, want it in combat", sim.CurrentTime)
			}
			character.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)
			axe := critChances()
			check("sword swapped for the axe", proto.WeaponType_WeaponTypeAxe, sword, axe, tc.axeCrit-tc.swordCrit)

			for sim.CurrentTime < 6*time.Second {
				sim.Step()
			}
			character.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Main, false)
			check("axe swapped back for the sword", proto.WeaponType_WeaponTypeSword, axe, critChances(), tc.swordCrit-tc.axeCrit)

			// An iteration that ends on the axe starts the next one on the sword with the crit it began with.
			character.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)
			sim.Cleanup()
			sim.Reset()
			for _, stat := range []stats.Stat{stats.PhysicalCritPercent, stats.SpellCritPercent} {
				if got := character.GetStat(stat); !WithinToleranceFloat64(startCrit[stat], got, 1e-9) {
					t.Errorf("the next iteration starts at %v %s, want %v", got, stat.StatName(), startCrit[stat])
				}
			}
		})
	}
}

func init() {
	RegisterAgentFactory(
		proto.Player_Warlock{},
		proto.Spec_SpecWarlock,
		newFakeDotWarlock,
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_Warlock)
			if !ok {
				panic("Invalid spec value for Warlock!")
			}
			player.Spec = playerSpec
		},
	)
}

type fakeDotWarlock struct {
	Character

	Corruption    *Spell
	Curse         *Spell
	DrainLife     *Spell
	Haunt         *Spell
	damagingTicks int
}

func newFakeDotWarlock(char *Character, _ *proto.Player, _ *proto.Raid) Agent {
	return &fakeDotWarlock{Character: *char}
}

func (fw *fakeDotWarlock) GetCharacter() *Character { return &fw.Character }

func (fw *fakeDotWarlock) Initialize() {
	fw.Corruption = fw.RegisterSpell(SpellConfig{
		ActionID:         ActionID{SpellID: 172},
		SpellSchool:      SpellSchoolShadow,
		DefenseType:      DefenseTypeMagic,
		ProcMask:         ProcMaskSpellDamage,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: DotConfig{
			Aura:          Aura{Label: "Corruption"},
			NumberOfTicks: 6,
			TickLength:    time.Second * 3,
			OnTick: func(sim *Simulation, target *Unit, dot *Dot) {
				if dot.Spell.CalcAndDealPeriodicDamage(sim, target, 100, dot.OutcomeTick).Damage > 0 {
					fw.damagingTicks++
				}
			},
		},

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
			}
			spell.DealOutcome(sim, result)
		},
	})

	fw.Curse = fw.RegisterSpell(SpellConfig{
		ActionID:         ActionID{SpellID: 702},
		SpellSchool:      SpellSchoolShadow,
		DefenseType:      DefenseTypeMagic,
		ProcMask:         ProcMaskSpellDamage,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
		},
	})

	fw.DrainLife = fw.RegisterSpell(fakeWarlockDot(689, "Drain Life", ClassFlags{Family: 5, Mask: [4]uint32{8}}, SpellFlagChanneled))
	fw.Haunt = fw.RegisterSpell(fakeWarlockDot(1293694, "Haunt", ClassFlags{Family: 5, Mask: [4]uint32{0, 262144}}, 0))
}

// A periodic warlock spell with the game files' family 5 flags, which always lands.
func fakeWarlockDot(id int32, label string, flags ClassFlags, spellFlags SpellFlag) SpellConfig {
	return SpellConfig{
		ActionID:         ActionID{SpellID: id},
		ClassFlags:       flags,
		Flags:            spellFlags,
		SpellSchool:      SpellSchoolShadow,
		DefenseType:      DefenseTypeMagic,
		ProcMask:         ProcMaskSpellDamage,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		Dot: DotConfig{
			Aura:          Aura{Label: label},
			NumberOfTicks: 5,
			TickLength:    time.Second,
			OnTick: func(sim *Simulation, target *Unit, dot *Dot) {
				dot.Spell.CalcAndDealPeriodicDamage(sim, target, 100, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeAlwaysHit)
			spell.Dot(target).Apply(sim)
			spell.DealOutcome(sim, result)
		},
	}
}

func (fw *fakeDotWarlock) ApplyTalents()                  {}
func (fw *fakeDotWarlock) Reset(_ *Simulation)            {}
func (fw *fakeDotWarlock) OnGCDReady(_ *Simulation)       {}
func (fw *fakeDotWarlock) OnEncounterStart(_ *Simulation) {}

func setupTouchOfTheGraveWarlock(t *testing.T) (*Simulation, *fakeDotWarlock, *Aura) {
	t.Helper()
	sim, fw := setupFakeWarlockSim(proto.Race_RaceUndead)
	aura := fw.GetAura("Touch of the Grave")
	if aura == nil || !aura.IsActive() || aura.Icd == nil {
		t.Fatalf("An undead warlock should carry an active Touch of the Grave with its cooldown")
	}
	return sim, fw, aura
}

func setupFakeWarlockSim(race proto.Race) (*Simulation, *fakeDotWarlock) {
	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 100},
		Raid: SinglePlayerRaidProto(&proto.Player{
			Name:      "Warlock",
			Class:     proto.Class_ClassWarlock,
			Race:      race,
			Buffs:     &proto.IndividualBuffs{},
			Spec:      &proto.Player_Warlock{},
			Equipment: &proto.EquipmentSpec{},
			Rotation:  &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
		}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter: &proto.Encounter{
			Targets:  []*proto.Target{{Name: "target", Level: 63, MobType: proto.MobType_MobTypeHumanoid}},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()

	return sim, sim.Raid.Parties[0].Players[0].(*fakeDotWarlock)
}

// Whether Touch of the Grave procs within the given number of hits. Time never advances, so the
// first proc leaves its cooldown running for good.
func touchOfTheGraveProcsWithin(sim *Simulation, aura *Aura, tries int, hit func()) bool {
	for range tries {
		hit()
		if !aura.Icd.IsReady(sim) {
			return true
		}
	}
	return false
}

func TestTouchOfTheGraveProcsOnTheLandedCastOfADot(t *testing.T) {
	sim, fw, aura := setupTouchOfTheGraveWarlock(t)

	if !touchOfTheGraveProcsWithin(sim, aura, 200, func() { fw.Corruption.SkipCastAndApplyEffects(sim, fw.CurrentTarget) }) {
		t.Fatalf("A DoT's landed cast deals no damage but should still trigger Touch of the Grave")
	}
}

func TestTouchOfTheGraveIgnoresSpellsWithoutDamage(t *testing.T) {
	sim, fw, aura := setupTouchOfTheGraveWarlock(t)

	if touchOfTheGraveProcsWithin(sim, aura, 200, func() { fw.Curse.SkipCastAndApplyEffects(sim, fw.CurrentTarget) }) {
		t.Fatalf("A debuff with no damage should not trigger Touch of the Grave")
	}
}

func TestTouchOfTheGraveIgnoresPeriodicTicks(t *testing.T) {
	sim, fw, aura := setupTouchOfTheGraveWarlock(t)

	dot := fw.Corruption.Dot(fw.CurrentTarget)
	dot.Apply(sim)
	if touchOfTheGraveProcsWithin(sim, aura, 200, func() { dot.TickOnce(sim) }) {
		t.Fatalf("A DoT's ticks should not trigger Touch of the Grave")
	}
	if fw.damagingTicks == 0 {
		t.Fatalf("The DoT's ticks dealt no damage, so the test proves nothing")
	}
}

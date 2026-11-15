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
	if got := fw.GetSpell(ActionID{SpellID: 20572}).TimeToReady(sim); got != 2*time.Minute {
		t.Fatalf("Blood Fury should be ready again 2 minutes after its cast, got %v", got)
	}
}

func TestShatterCurseCutsMagicDamageTaken(t *testing.T) {
	sim, fw := setupRacialSim(proto.Race_RaceOrc)

	before := fw.PseudoStats.SchoolDamageTakenMultiplier
	castRacial(t, sim, fw, 1299026)

	for school := stats.SchoolIndexPhysical; school < stats.SchoolLen; school++ {
		want := before[school] * 0.95
		if school == stats.SchoolIndexPhysical {
			want = before[school]
		}
		if got := fw.PseudoStats.SchoolDamageTakenMultiplier[school]; !WithinToleranceFloat64(want, got, 1e-9) {
			t.Errorf("Shatter Curse should leave school %d taking x%0.2f, got x%0.2f", school, want, got)
		}
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

// One-bit class flags in family, on the lowest bit that every mask in named sets and no mask in
// unnamed sets; zero flags where no bit does.
func oneClassFlagBit(family int32, named, unnamed []ClassFlags) ClassFlags {
	for word := range 4 {
		for bit := range 32 {
			flags := ClassFlags{Family: family}
			flags.Mask[word] = 1 << bit
			if !slices.ContainsFunc(named, func(f ClassFlags) bool { return !f.Matches(flags) }) &&
				!slices.ContainsFunc(unnamed, flags.Matches) {
				return flags
			}
		}
	}
	return ClassFlags{}
}

func warriorEurekaFlags() []ClassFlags {
	mods := eurekaMods(1259813)
	return []ClassFlags{mods.CostFlags, mods.DamageFlags}
}

// A warrior ability as spelldata builds one: the game files' class flags and no ClassSpellMask.
func fakeWarriorAbility(id int32, flags ClassFlags, damage float64) SpellConfig {
	return SpellConfig{
		ActionID:         ActionID{SpellID: id},
		ClassFlags:       flags,
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
	unnamed := oneClassFlagBit(ClassSpellFamilies[proto.Class_ClassWarrior], nil, warriorEurekaFlags())
	config := fakeWarriorAbility(fakeRendActionID.SpellID, unnamed, 0)
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
	castAt := sim.CurrentTime
	eureka := fw.GetAura("Eureka!")
	if !eureka.IsActive() || eureka.MaxStacks != 0 {
		t.Fatalf("Eureka! should come up with no charges, got %d of %d", eureka.GetStacks(), eureka.MaxStacks)
	}

	mods := eurekaMods(1259813)
	family := ClassSpellFamilies[proto.Class_ClassWarrior]
	named := oneClassFlagBit(family, warriorEurekaFlags(), nil)
	unnamed := oneClassFlagBit(family, nil, warriorEurekaFlags())
	if named.IsZero() || unnamed.IsZero() {
		t.Fatalf("the warrior's Eureka! masks %+v leave no class flag bit named by both (%+v) or by neither (%+v)",
			warriorEurekaFlags(), named, unnamed)
	}
	mortalStrike := fw.RegisterSpell(fakeWarriorAbility(12294, named, 100))
	sunderArmor := fw.RegisterSpell(fakeWarriorAbility(7386, unnamed, 0))
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

	if !WithinToleranceFloat64(1+mods.Damage, mortalStrike.DirectDamageMultiplier, 0.0001) {
		t.Fatalf("Eureka! should multiply Mortal Strike's direct damage by %0.2f, got %0.2f", 1+mods.Damage, mortalStrike.DirectDamageMultiplier)
	}
	if want := 30 * (1 + mods.Cost); !WithinToleranceFloat64(want, mortalStrike.Cost.GetCurrentCost(), 0.0001) {
		t.Fatalf("Eureka! should cut Mortal Strike's 30 Rage to %0.1f, got %0.1f", want, mortalStrike.Cost.GetCurrentCost())
	}
	if sunderArmor.DirectDamageMultiplier != 1 || sunderArmor.Cost.GetCurrentCost() != 30 {
		t.Fatalf("Eureka! should leave Sunder Armor alone, which its class mask does not name")
	}
	if itemSpell.DirectDamageMultiplier != 1 {
		t.Fatalf("Eureka! should ignore spells that are not class abilities")
	}

	// However many abilities it raises, Eureka! holds to the last moment of its duration.
	for range 5 {
		mortalStrike.SkipCastAndApplyEffects(sim, fw.CurrentTarget)
	}
	var upAtTheEnd, upAfter, done bool
	var multiplierAtTheEnd float64
	at := func(when time.Duration, action func(sim *Simulation)) {
		sim.AddPendingAction(NewDelayedAction(DelayedActionOptions{DoAt: when, OnAction: action}))
	}
	at(castAt+mods.Duration-time.Millisecond, func(_ *Simulation) {
		upAtTheEnd, multiplierAtTheEnd = eureka.IsActive(), mortalStrike.DirectDamageMultiplier
	})
	// An aura runs out when the sim advances to its end, which takes a pending action at that time.
	at(castAt+mods.Duration, func(_ *Simulation) { upAfter, done = eureka.IsActive(), true })
	for !done && !sim.Step() {
	}
	if !upAtTheEnd || !WithinToleranceFloat64(1+mods.Damage, multiplierAtTheEnd, 0.0001) {
		t.Fatalf("Eureka! should keep Mortal Strike at x%0.2f to the end of its %v, got up %t at x%0.2f",
			1+mods.Damage, mods.Duration, upAtTheEnd, multiplierAtTheEnd)
	}
	if upAfter {
		t.Fatalf("Eureka! should end %v after its cast", mods.Duration)
	}
	if mortalStrike.DirectDamageMultiplier != 1 || mortalStrike.Cost.GetCurrentCost() != 30 {
		t.Fatalf("Eureka! should drop its damage bonus and cost cut when it ends, got x%0.2f and %0.1f Rage",
			mortalStrike.DirectDamageMultiplier, mortalStrike.Cost.GetCurrentCost())
	}
}

func TestGnomeEurekaReadsEachClassRow(t *testing.T) {
	for spellID, class := range map[int32]proto.Class{
		1259812: proto.Class_ClassRogue,
		1259813: proto.Class_ClassWarrior,
		1259817: proto.Class_ClassMage,
		1259821: proto.Class_ClassWarlock,
		1259823: proto.Class_ClassPriest,
	} {
		mods := eurekaMods(spellID)
		family := ClassSpellFamilies[class]
		if !(mods.Cost < 0 && mods.Damage > 0) {
			t.Errorf("Eureka! %d reads cost %v and damage %v, want a cut and a raise", spellID, mods.Cost, mods.Damage)
		}
		if mods.Duration <= 0 || mods.Cooldown <= 0 {
			t.Errorf("Eureka! %d reads duration %v and cooldown %v, want both", spellID, mods.Duration, mods.Cooldown)
		}
		if mods.CostFlags.Mask == [4]uint32{} || mods.CostFlags.Family != family ||
			mods.DamageFlags.Mask == [4]uint32{} || mods.DamageFlags.Family != family {
			t.Errorf("Eureka! %d reads cost flags %+v and damage flags %+v, want masks in %s's family %d",
				spellID, mods.CostFlags, mods.DamageFlags, class, family)
		}
		if !mods.ChannelFlags.IsZero() && (mods.ChannelFlags.Family != family || mods.Channel <= 0) {
			t.Errorf("Eureka! %d reads channel flags %+v with %v, want a raise on channels in %s's family %d",
				spellID, mods.ChannelFlags, mods.Channel, class, family)
		}
	}
}

// Eureka!'s idx2 raises the ticks of a warlock channel its mask names that land while the buff is up,
// on a channel started before the buff and on one that runs on past its end. A periodic spell that is
// not a channel takes no bonus, on the mask or off it.
func TestGnomeWarlockEurekaChannel(t *testing.T) {
	sim, fw := setupFakeWarlockSim(proto.Race_RaceGnome, &proto.IndividualBuffs{})
	target := fw.CurrentTarget

	eurekaSpell := fw.GetSpell(ActionID{SpellID: 1259821})
	if eurekaSpell == nil {
		t.Fatalf("a gnome warlock should carry Eureka! 1259821")
	}
	mods := eurekaMods(1259821)
	if mods.ChannelFlags.IsZero() || mods.Channel <= 0 {
		t.Fatalf("the warlock's Eureka! reads channel flags %+v with %v, want a raise on its channels", mods.ChannelFlags, mods.Channel)
	}
	if fw.DrainLife.ClassFlags.IsZero() {
		t.Fatalf("no class flag bit sits on both the warlock's Eureka! channel mask %+v and its cost mask %+v",
			mods.ChannelFlags, mods.CostFlags)
	}
	boosted := 1 + mods.Channel

	at := func(when time.Duration, action func(sim *Simulation)) {
		sim.AddPendingAction(NewDelayedAction(DelayedActionOptions{DoAt: when, OnAction: action}))
	}
	multiplier := func(spell *Spell) float64 { return spell.Dot(target).PeriodicDamageMultiplier }
	drainLife := fw.DrainLife.Dot(target)
	var tickMultipliers []float64
	onTick := drainLife.onTick
	drainLife.onTick = func(sim *Simulation, target *Unit, dot *Dot) {
		tickMultipliers = append(tickMultipliers, dot.PeriodicDamageMultiplier)
		onTick(sim, target, dot)
	}
	channelFrom := func(start time.Duration, what string, want ...float64) {
		t.Helper()
		tickMultipliers = nil
		at(start, func(sim *Simulation) { fw.DrainLife.SkipCastAndApplyEffects(sim, target) })
		for !drainLife.IsActive() && !sim.Step() {
		}
		for drainLife.IsActive() && !sim.Step() {
		}
		if len(tickMultipliers) != len(want) {
			t.Fatalf("%s should tick %d times, got %d", what, len(want), len(tickMultipliers))
		}
		for i, got := range tickMultipliers {
			if !WithinToleranceFloat64(want[i], got, 0.0001) {
				t.Fatalf("%s: tick %d should take x%0.2f, got x%0.2f", what, i+1, want[i], got)
			}
		}
	}

	start := sim.CurrentTime
	eurekaCast := false
	at(start+2500*time.Millisecond, func(sim *Simulation) { eurekaCast = eurekaSpell.Cast(sim, target) })
	channelFrom(start, "a Drain Life with Eureka! cast after its second tick", 1, 1, boosted, boosted, boosted)
	if !eurekaCast {
		t.Fatalf("a gnome warlock should cast Eureka! during a channel")
	}
	if got := multiplier(fw.Haunt); got != 1 {
		t.Fatalf("Eureka! should leave a periodic spell that is not a channel alone, even on its mask, got x%0.2f", got)
	}
	if got := multiplier(fw.Corruption); got != 1 {
		t.Fatalf("Eureka! should leave Corruption's ticks alone, got x%0.2f", got)
	}

	// Eureka! runs out 2.5 seconds into this channel.
	channelFrom(start+mods.Duration, "a Drain Life running past the end of Eureka!", boosted, boosted, 1, 1, 1)
}

// Endurance carries no hit and no crit damage, Artisan's Precision 1% physical and spell hit.
func TestHitAndCritDamageRacials(t *testing.T) {
	_, human := setupRacialSim(proto.Race_RaceHuman)
	for _, tc := range []struct {
		race proto.Race
		hit  float64
	}{
		{proto.Race_RaceTauren, 0},
		{proto.Race_RaceDwarf, 1},
	} {
		_, character := setupRacialSim(tc.race)
		for _, stat := range []stats.Stat{stats.PhysicalHitPercent, stats.SpellHitPercent} {
			if got, want := character.GetStat(stat), human.GetStat(stat)+tc.hit; got != want {
				t.Errorf("a %s should hold %s %v, got %v", tc.race, stat.StatName(), want, got)
			}
		}
		if got := character.PseudoStats.CritDamageMultiplier; got != human.PseudoStats.CritDamageMultiplier {
			t.Errorf("a %s should crit for a human's x%0.4f, got x%0.4f", tc.race, human.PseudoStats.CritDamageMultiplier, got)
		}
	}
}

// Ruthless raises the crit of every spell and attack whatever the character holds, bare hands
// included. An undead and an orc have no crit racial, so at the same attributes they hold the crit
// without it.
func TestRuthlessNeedsNoWeapon(t *testing.T) {
	for _, tc := range []struct {
		race proto.Race
		crit float64
	}{
		{proto.Race_RaceHuman, 1},
		{proto.Race_RaceOrc, 0},
	} {
		for name, weapons := range map[string][]*proto.ItemSpec{
			"bare hands": WeaponTestSlots(0, 0, 0),
			"daggers":    WeaponTestSlots(WeaponTestDaggerID, WeaponTestDaggerID, WeaponTestBowID),
			"a sword":    WeaponTestSlots(WeaponTestSwordID, 0, 0),
			"an axe":     WeaponTestSlots(weaponTestAxeID, 0, 0),
		} {
			_, character := NewWeaponTestSim(t, tc.race, weapons, nil, nil)
			sim, undead := NewWeaponTestSim(t, proto.Race_RaceUndead, weapons, nil, nil)
			undead.AddStatsDynamic(sim, stats.Stats{
				stats.Agility:   character.GetStat(stats.Agility) - undead.GetStat(stats.Agility),
				stats.Intellect: character.GetStat(stats.Intellect) - undead.GetStat(stats.Intellect),
			})
			for _, stat := range []stats.Stat{stats.PhysicalCritPercent, stats.SpellCritPercent} {
				if got := character.GetStat(stat) - undead.GetStat(stat); !WithinToleranceFloat64(tc.crit, got, 1e-6) {
					t.Errorf("%s with %s: %s +%v over an undead, want +%v", tc.race, name, stat.StatName(), got, tc.crit)
				}
			}
		}
	}
}

func TestTouchOfTheGraveUsesTheMeleeVariant(t *testing.T) {
	_, fw := setupRacialSim(proto.Race_RaceUndead)

	if aura := fw.GetAura("Touch of the Grave"); aura == nil || aura.ActionIDForProc.SpellID != 1260189 {
		t.Fatalf("An undead warrior should carry the 20%% Touch of the Grave")
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
		if !WithinToleranceFloat64(1.02, swingRatio, 0.001) {
			t.Errorf("%s: Wind Blessed should swing 2%% faster, got x%0.4f", race, swingRatio)
		}
		if at := fw.AttackTables[fw.CurrentTarget.UnitIndex]; !WithinToleranceFloat64(1.05, at.DamageDealtMultiplier, 0.0001) {
			t.Errorf("%s: Elemental Insight should grant 5%% damage against elementals, got x%0.4f", race, at.DamageDealtMultiplier)
		}
		if at := fw.AttackTables[fw.CurrentTarget.UnitIndex]; at.CritMultiplier != 1 {
			t.Errorf("%s: Elemental Insight states no critical damage against elementals, got x%0.4f", race, at.CritMultiplier)
		}

		hasReadLeyLine := fw.GetSpell(ActionID{SpellID: 1259705}) != nil
		hasSkysight := fw.GetSpell(ActionID{SpellID: 1259686}) != nil
		if hasReadLeyLine != (race == proto.Race_RaceHighOrderSkyborne) || hasSkysight != (race == proto.Race_RaceWindshaperSkyborne) {
			t.Errorf("%s: Read Leyline is the High Order's and Skysight the Windshapers', got %t / %t", race, hasReadLeyLine, hasSkysight)
		}
	}
}

// Read Leyline and Skysight each hold a plain cast and a tag 1 cast for an Attuned or Charged
// character, on one cooldown: the plain one lasts 15 s at half the tagged one's bonus, and the aura
// takes its bonus back off when it expires. An activation with no cast takes the plain bonus, also
// after a tagged cast.
func TestSkyborneCastsShareACooldown(t *testing.T) {
	for _, tc := range []struct {
		race    proto.Race
		spellID int32
		aura    string
		value   func(*fakeDotWarlock) float64
		plain   float64
		buffed  float64
	}{
		{proto.Race_RaceHighOrderSkyborne, 1259705, "Energized", func(fw *fakeDotWarlock) float64 { return fw.manaRegenMultiplier }, 1.5, 2},
		{proto.Race_RaceWindshaperSkyborne, 1259686, "Elemental Blessing", func(fw *fakeDotWarlock) float64 { return fw.PseudoStats.MovementSpeedMultiplier }, 1.15, 1.3},
	} {
		for tag, want := range []float64{tc.plain, tc.buffed} {
			sim, fw := setupFakeWarlockSim(tc.race, &proto.IndividualBuffs{})
			before := tc.value(fw)
			cast := fw.GetSpell(ActionID{SpellID: tc.spellID, Tag: int32(tag)})
			other := fw.GetSpell(ActionID{SpellID: tc.spellID, Tag: int32(1 - tag)})
			if cast == nil || other == nil {
				t.Fatalf("%s: spell %d should register tags 0 and 1", tc.race, tc.spellID)
			}
			cast.SkipCastAndApplyEffects(sim, fw.CurrentTarget)
			cast.CD.Use(sim)

			aura := fw.GetAura(tc.aura)
			if !aura.IsActive() || aura.Duration != 15*time.Second {
				t.Errorf("%s tag %d: %s should be up for 15s, got active %t for %v", tc.race, tag, tc.aura, aura.IsActive(), aura.Duration)
			}
			if got := tc.value(fw) / before; !WithinToleranceFloat64(want, got, 1e-9) {
				t.Errorf("%s tag %d: %s should multiply by %v, got x%0.4f", tc.race, tag, tc.aura, want, got)
			}
			if other.CanCast(sim, fw.CurrentTarget) {
				t.Errorf("%s tag %d: the other cast should share the cooldown", tc.race, tag)
			}
			aura.Deactivate(sim)
			if got := tc.value(fw); !WithinToleranceFloat64(before, got, 1e-9) {
				t.Errorf("%s tag %d: after %s ends the multiplier should be back at %v, got %v", tc.race, tag, tc.aura, before, got)
			}

			aura.Activate(sim)
			if got := tc.value(fw) / before; !WithinToleranceFloat64(tc.plain, got, 1e-9) {
				t.Errorf("%s tag %d: %s raised with no cast should multiply by %v, got x%0.4f", tc.race, tag, tc.aura, tc.plain, got)
			}
			aura.Deactivate(sim)
		}
	}
}

// Read Leyline cannot be cast under Innervate, and an Innervate that lands ends Energized.
func TestInnervateReplacesEnergized(t *testing.T) {
	sim, fw := setupFakeWarlockSim(proto.Race_RaceHighOrderSkyborne, &proto.IndividualBuffs{Innervates: 1})

	innervates := fw.GetAurasWithTag(innervateTag)
	if len(innervates) == 0 {
		t.Fatalf("an Innervates buff should register an aura tagged %q", innervateTag)
	}
	readLeyline := fw.GetSpell(ActionID{SpellID: 1259705})
	energized := fw.GetAura("Energized")

	if !readLeyline.CanCast(sim, fw.CurrentTarget) {
		t.Fatalf("Read Leyline should cast with no Innervate up")
	}
	readLeyline.SkipCastAndApplyEffects(sim, fw.CurrentTarget)
	if !energized.IsActive() {
		t.Fatalf("Read Leyline should raise Energized")
	}
	innervates[0].Activate(sim)
	if energized.IsActive() {
		t.Errorf("Innervate landing should end Energized")
	}

	for _, tag := range []int32{0, 1} {
		if fw.GetSpell(ActionID{SpellID: 1259705, Tag: tag}).CanCast(sim, fw.CurrentTarget) {
			t.Errorf("Read Leyline tag %d should not cast under Innervate", tag)
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

func init() {
	RegisterAgentFactory(
		proto.Player_Warlock{},
		proto.Spec_SpecWarlock,
		newFakeDotWarlock,
		func(player *proto.Player, spec any) {
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
	fw.EnableManaBar()
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

	// Both on the Eureka! channel mask and on its cost mask; only Drain Life is a channel.
	mods := eurekaMods(1259821)
	onEurekaMasks := oneClassFlagBit(ClassSpellFamilies[proto.Class_ClassWarlock], []ClassFlags{mods.ChannelFlags, mods.CostFlags}, nil)
	fw.DrainLife = fw.RegisterSpell(fakeWarlockDot(689, "Drain Life", onEurekaMasks, SpellFlagChanneled))
	fw.Haunt = fw.RegisterSpell(fakeWarlockDot(1293694, "Haunt", onEurekaMasks, 0))
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
	sim, fw := setupFakeWarlockSim(proto.Race_RaceUndead, &proto.IndividualBuffs{})
	aura := fw.GetAura("Touch of the Grave")
	if aura == nil || !aura.IsActive() || aura.Icd == nil {
		t.Fatalf("An undead warlock should carry an active Touch of the Grave with its cooldown")
	}
	return sim, fw, aura
}

func setupFakeWarlockSim(race proto.Race, buffs *proto.IndividualBuffs) (*Simulation, *fakeDotWarlock) {
	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 100},
		Raid: SinglePlayerRaidProto(&proto.Player{
			Name:      "Warlock",
			Class:     proto.Class_ClassWarlock,
			Race:      race,
			Buffs:     buffs,
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

func TestTouchOfTheGraveProcsOnceIn25Seconds(t *testing.T) {
	sim, fw, aura := setupTouchOfTheGraveWarlock(t)
	target := fw.CurrentTarget
	drain := fw.GetSpell(ActionID{SpellID: 1260198})
	drains := func() int32 { return drain.SpellMetrics[target.UnitIndex].Casts }
	hit := func() { fw.Corruption.SkipCastAndApplyEffects(sim, target) }

	if !touchOfTheGraveProcsWithin(sim, aura, 200, hit) {
		t.Fatalf("Touch of the Grave should proc within 200 landed casts")
	}
	procAt := sim.CurrentTime

	procsAfter, done := false, false
	at := func(when time.Duration, action func(sim *Simulation)) {
		sim.AddPendingAction(NewDelayedAction(DelayedActionOptions{DoAt: when, OnAction: action}))
	}
	// A proc from these hits would restart the cooldown, and the hits at 15 seconds could not proc.
	at(procAt+15*time.Second-time.Millisecond, func(_ *Simulation) {
		for range 200 {
			hit()
		}
	})
	at(procAt+15*time.Second, func(sim *Simulation) {
		procsAfter = touchOfTheGraveProcsWithin(sim, aura, 200, hit)
	})
	// Each drain lands one spell batch after its proc.
	at(procAt+15*time.Second+2*SpellBatchWindow, func(_ *Simulation) { done = true })
	for !done && !sim.Step() {
	}

	if !procsAfter {
		t.Errorf("Touch of the Grave should proc again on the first hits once 15 seconds have passed, and on none before")
	}
	if got := drains(); got != 2 {
		t.Errorf("Touch of the Grave should drain twice, at its first proc and 15 seconds later, got %d", got)
	}
}

// The drain is a share of the caster's health, which no Shadow damage taken on the target or Shadow
// damage done by the caster scales.
func TestTouchOfTheGraveIgnoresShadowModifiers(t *testing.T) {
	sim, fw, _ := setupTouchOfTheGraveWarlock(t)
	target := fw.CurrentTarget
	attackTable := fw.AttackTables[target.UnitIndex]
	drain := fw.GetSpell(ActionID{SpellID: 1260198})

	target.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexShadow] *= 2
	fw.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexShadow] *= 2

	if got := fw.Corruption.TargetDamageMultiplier(sim, attackTable, false); got != 2 {
		t.Fatalf("a Shadow spell should take the target's x2 Shadow damage taken, got x%0.2f", got)
	}
	if got := drain.TargetDamageMultiplier(sim, attackTable, false); got != 1 {
		t.Errorf("the drain should take no damage taken multiplier on the target, got x%0.2f", got)
	}
	if got := drain.AttackerDamageMultiplier(attackTable, false); got != 1 {
		t.Errorf("the drain should take no damage done multiplier of the caster, got x%0.2f", got)
	}
}

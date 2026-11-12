package warlock

import (
	"math"
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

func init() {
	RegisterWarlock()
	common.RegisterAllEffects()
}

// Each suite runs a build preset from ui/specs/warlock/dps/builds, the file the UI's build of that name reads:
// its talents, demons, consumables and rotation, on the gear set named.
func TestDemonologyFireBrand(t *testing.T) {
	runBuildSuite(t, "Fire Brand", "demo_fire_brand", "preraid", proto.Race_RaceOrc)
}

func TestDemonologyShadowBrand(t *testing.T) {
	runBuildSuite(t, "Shadow Brand", "demo_shadow_brand", "preraid", proto.Race_RaceOrc)
}

func TestDestruction(t *testing.T) {
	runBuildSuite(t, "Destruction", "destro", "raid", proto.Race_RaceGnome)
}

func TestAffliction(t *testing.T) {
	runBuildSuite(t, "Affliction", "sm_ruin", "preraid", proto.Race_RaceOrc)
}

func TestAfflictionWrack(t *testing.T) {
	runBuildSuite(t, "Wrack", "aff_wrack", "raid", proto.Race_RaceHuman)
}

func runBuildSuite(t *testing.T, label string, buildFile string, gearSet string, otherRace proto.Race) {
	build := core.GetTestBuildFromJSON(proto.Class_ClassWarlock, "../../ui/specs/warlock/dps/builds", buildFile, core.ItemFilter{}, nil, nil)
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:       proto.Class_ClassWarlock,
			Race:        build.Race,
			OtherRaces:  []proto.Race{otherRace},
			SpecOptions: core.SpecOptionsCombo{Label: label, SpecOptions: build.SpecOptions.SpecOptions},
			GearSet:     core.GetGearSet("../../ui/specs/warlock/dps/gear_sets", gearSet),
			Talents:     build.Talents,
			Consumables: build.Consumables,
			Profession1: build.Profession1,
			Profession2: build.Profession2,
			Rotation:    core.RotationCombo{Label: buildFile, Rotation: build.Rotation.Rotation},
			// The UI preset's distance, from which the demon walks in
			StartingDistance: 20,
			ItemFilter: core.ItemFilter{
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeDagger,
					proto.WeaponType_WeaponTypeStaff,
					proto.WeaponType_WeaponTypeSword,
				},
				ArmorType: proto.ArmorType_ArmorTypeCloth,
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeWand,
				},
				EnchantBlacklist: []int32{2673, 3225, 3273},
				IDBlacklist:      []int32{28556},
			},
		},
	}))
}

// A reset sim of one ungeared warlock with no rotation.
func newWarlockSim(talents string, options *proto.WarlockOptions, encounter *proto.Encounter) (*core.Simulation, *Warlock) {
	raid := core.SinglePlayerRaidProto(warlockPlayer("Warlock", talents, options), &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	sim := newWarlockRaidSim(raid, encounter)
	return sim, sim.Raid.Parties[0].Players[0].(*Warlock)
}

// A reset sim of this raid.
func newWarlockRaidSim(raid *proto.Raid, encounter *proto.Encounter) *core.Simulation {
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim
}

// An ungeared warlock with no rotation.
func warlockPlayer(name string, talents string, options *proto.WarlockOptions) *proto.Player {
	return &proto.Player{
		Name:          name,
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarlock,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: talents,
		Spec:          &proto.Player_Warlock{Warlock: &proto.Warlock{Options: &proto.Warlock_Options{ClassOptions: options}}},
		Rotation:      &proto.APLRotation{},
	}
}

func stepFor(sim *core.Simulation, d time.Duration) {
	for end := sim.CurrentTime + d; sim.CurrentTime < end && !sim.Step(); {
	}
}

// Soul Fire takes the Soul Shard its row states, and none while Decimation's buff (440873) is up.
func TestSoulFireShardAndDecimation(t *testing.T) {
	sim, warlock := newWarlockSim("-000000000002", &proto.WarlockOptions{SoulShards: 1}, core.MakeSingleTargetEncounter(0))
	target := sim.Encounter.AllTargetUnits[0]
	soulFire := warlock.Soulfire
	decimation := warlock.GetAuraByID(core.ActionID{SpellID: 440873})
	if decimation == nil {
		t.Fatalf("Decimation 2/2 registers no buff 440873")
	}

	if !soulFire.ExtraCastCondition(sim, target) {
		t.Fatalf("Soul Fire cannot be cast with 1 Soul Shard")
	}
	soulFire.SkipCastAndApplyEffects(sim, target)
	if got := warlock.SoulShardsAura.GetStacks(); got != 0 {
		t.Fatalf("%d Soul Shards left after Soul Fire, want 0", got)
	}
	if soulFire.ExtraCastCondition(sim, target) {
		t.Errorf("Soul Fire can be cast with no Soul Shard and no Decimation")
	}

	decimation.Activate(sim)
	if !soulFire.ExtraCastCondition(sim, target) {
		t.Errorf("Soul Fire cannot be cast with no Soul Shard under Decimation")
	}
	soulFire.SkipCastAndApplyEffects(sim, target)
	if got := warlock.SoulShardsAura.GetStacks(); got != 0 {
		t.Errorf("%d Soul Shards after Soul Fire under Decimation, want 0 and none spent", got)
	}
}

// A demon's melee and spell crit are the warlock's spell crit, and follow it when it changes in combat.
func TestDemonCritIsTheWarlocksSpellCrit(t *testing.T) {
	sim, warlock := newWarlockSim("", &proto.WarlockOptions{Summon: proto.WarlockOptions_Succubus}, core.MakeSingleTargetEncounter(0))
	sayaad := warlock.Succubus
	if !sayaad.IsEnabled() {
		t.Fatalf("Sayaad not summoned at the start")
	}

	check := func(when string) {
		spellCrit := warlock.GetStat(stats.SpellCritPercent)
		if spellCrit <= 0 {
			t.Fatalf("%s: warlock has %.2f%% spell crit, want some to inherit", when, spellCrit)
		}
		for _, stat := range []stats.Stat{stats.PhysicalCritPercent, stats.SpellCritPercent} {
			if got := sayaad.GetStat(stat); got != spellCrit {
				t.Errorf("%s: Sayaad %s %.2f%%, want the warlock's spell crit %.2f%%", when, stat.StatName(), got, spellCrit)
			}
		}
	}
	check("at the start")
	warlock.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: 5})
	stepFor(sim, core.PetUpdateInterval)
	check("a pet update after +5% spell crit")
}

// A demon takes the warlock's spell hit whole as spell hit, and 8/17 of it as melee hit.
func TestDemonHitIsTheWarlocksSpellHit(t *testing.T) {
	_, warlock := newWarlockSim("", &proto.WarlockOptions{}, core.MakeSingleTargetEncounter(0))
	inherited := warlock.petStatInheritance(0)(stats.Stats{stats.SpellHitPercent: 17})
	if got := inherited[stats.SpellHitPercent]; got != 17 {
		t.Errorf("demon spell hit %.4f%% from 17%% spell hit, want 17%%", got)
	}
	if got := inherited[stats.PhysicalHitPercent]; got != 8 {
		t.Errorf("demon melee hit %.4f%% from 17%% spell hit, want 8%%", got)
	}
}

// In execute, Decimation's buff (440873) comes with the cast of a Shadow Bolt or Searing Pain whose hit lands,
// before a Shadow Bolt reaches the target; a miss or a full resist grants nothing, and the bolt's impact
// grants nothing either.
func TestDecimationGrantedOnLandedCast(t *testing.T) {
	newSim := func(resistance float64) (*core.Simulation, *Warlock, *core.Unit) {
		encounter := core.MakeSingleTargetEncounter(0)
		encounter.Duration = 3600
		encounter.ExecuteProportion_35 = 1
		// A fresh target, since the default one is shared by every test in the package.
		encounter.Targets[0] = core.FreshDefaultTargetConfig()
		encounter.Targets[0].Stats[stats.FireResistance] = resistance
		encounter.Targets[0].Stats[stats.ShadowResistance] = resistance
		sim, warlock := newWarlockSim("-000000000002", &proto.WarlockOptions{}, encounter)
		for !sim.Encounter.AllTargetUnits[0].IsExecutePhase35() && !sim.Step() {
		}
		warlock.DistanceFromTarget = 30
		return sim, warlock, sim.Encounter.AllTargetUnits[0]
	}

	sim, warlock, target := newSim(0)
	decimation := warlock.GetAuraByID(core.ActionID{SpellID: 440873})
	if warlock.ShadowBolt.TravelTime() <= 2*core.SpellBatchWindow {
		t.Fatalf("Shadow Bolt travels %v, too short to tell its cast from its impact", warlock.ShadowBolt.TravelTime())
	}
	granted, missed := 0, 0
	for i := range 400 {
		spell := warlock.SearingPain
		if i%2 == 1 {
			spell = warlock.ShadowBolt
		}
		decimation.Deactivate(sim)
		misses := spell.SpellMetrics[target.UnitIndex].Misses
		spell.SkipCastAndApplyEffects(sim, target)
		landed := spell.SpellMetrics[target.UnitIndex].Misses == misses
		if decimation.IsActive() != landed {
			t.Fatalf("%s at %v landed %v and left Decimation up %v, want the buff on a landed cast only",
				spell.ActionID, sim.CurrentTime, landed, decimation.IsActive())
		}
		if landed {
			granted++
		} else {
			missed++
		}
		decimation.Deactivate(sim)
		stepFor(sim, warlock.ShadowBolt.TravelTime()+core.SpellBatchWindow)
		if decimation.IsActive() {
			t.Fatalf("%s's impact at %v granted Decimation, want it granted at the cast only", spell.ActionID, sim.CurrentTime)
		}
	}
	if granted == 0 || missed == 0 {
		t.Fatalf("%d casts granted Decimation and %d missed, want both", granted, missed)
	}

	sim, warlock, target = newSim(1e12)
	decimation = warlock.GetAuraByID(core.ActionID{SpellID: 440873})
	for i := range 100 {
		spell := warlock.SearingPain
		if i%2 == 1 {
			spell = warlock.ShadowBolt
		}
		spell.SkipCastAndApplyEffects(sim, target)
		stepFor(sim, warlock.ShadowBolt.TravelTime()+core.SpellBatchWindow)
		if decimation.IsActive() {
			t.Fatalf("%s at %v granted Decimation against a target that resists it in full", spell.ActionID, sim.CurrentTime)
		}
	}
	if misses := warlock.SearingPain.SpellMetrics[target.UnitIndex].Misses; misses != 50 {
		t.Fatalf("%d of 50 Searing Pains against the fully resisting target missed, want all resisted in full", misses)
	}
}

// Bane of Agony's base ramps in thirds while each tick reads the spell power of the moment, with its share
// unramped, and Amplify Curse (talent 9) raises the whole tick for the whole of the Bane it landed with, spent or
// not. Each tick leaves its base and share in the dot's snapshot base, which the last tick clears as the dot ends,
// so the test reads the first 11.
func TestBaneOfAgonyTicksFollowSpellPower(t *testing.T) {
	encounter := core.MakeSingleTargetEncounter(0)
	encounter.Duration = 600
	sim, warlock := newWarlockSim("000000001", &proto.WarlockOptions{}, encounter)
	target := sim.Encounter.AllTargetUnits[0]
	agony := warlock.BaneOfAgony
	dot := agony.Dot(target)
	warlock.AddStatsDynamic(sim, stats.Stats{stats.SpellDamage: 300, stats.SpellHitPercent: 100})

	// The snapshot base of the first 11 ticks of one cast, with spellPower more spell power from the third on.
	castTicks := func(amplified bool, spellPower float64) []float64 {
		if amplified {
			warlock.AmplifyCurseAura.Activate(sim)
		}
		agony.SkipCastAndApplyEffects(sim, target)
		if !dot.IsActive() {
			t.Fatalf("Bane of Agony did not land")
		}
		warlock.AmplifyCurseAura.Deactivate(sim)
		var ticks []float64
		for len(ticks) < 11 {
			before := dot.TickCount()
			for dot.TickCount() == before && !sim.Step() {
			}
			ticks = append(ticks, dot.SnapshotBaseDamage)
			if len(ticks) == 2 {
				warlock.AddStatsDynamic(sim, stats.Stats{stats.SpellDamage: spellPower})
			}
		}
		return ticks
	}

	plain := castTicks(false, 500)
	coefficient := dot.BonusCoefficient
	if got, want := plain[2]-plain[1], 500*coefficient; math.Abs(got-want) > 1e-6 {
		t.Fatalf("500 spell power gained mid-dot moved the next tick by %.3f, want %.3f", got, want)
	}
	base := 2 * (plain[4] - plain[3])
	if step := plain[8] - plain[7]; base <= 0 || math.Abs(step-base/2) > 1e-6 {
		t.Fatalf("thirds %.3f, %.3f, %.3f do not ramp at half, one and one and a half of a base", plain[3], plain[4], plain[8])
	}
	share := plain[3] - base/2

	amplified := castTicks(true, 0)
	for i, ramp := range []float64{0.5, 1, 1.5} {
		want := (base*ramp + share) * warlock.AmplifyCurseMultiplier
		if got := amplified[4*i]; math.Abs(got-want) > 1e-6 {
			t.Errorf("amplified third %d tick %.3f, want %.3f: the base %.3f plus the share of now %.3f, amplified",
				i+1, got, want, base*ramp, share)
		}
	}
}

// Casting Bane of Agony spends Amplify Curse. The tooltip names it, but neither of Amplify Curse's class masks in
// the game files does.
func TestBaneOfAgonySpendsAmplifyCurse(t *testing.T) {
	sim, warlock := newWarlockSim("000000001", &proto.WarlockOptions{}, core.MakeSingleTargetEncounter(0))
	target := sim.Encounter.AllTargetUnits[0]

	warlock.AmplifyCurseAura.Activate(sim)
	if !warlock.BaneOfAgony.Cast(sim, target) {
		t.Fatalf("Bane of Agony did not cast")
	}
	if warlock.AmplifyCurseAura.IsActive() {
		t.Fatalf("Amplify Curse is still up after a Bane of Agony cast")
	}
}

// A Bane of Agony recast at an instant its tick is due, before the tick's action runs, deals that tick as the old
// Bane ends at the ramp of its time since the old Bane landed: at its first tick, the first third's share, the same
// as the new Bane's first tick.
func TestBaneOfAgonyRecastAtADueTickKeepsItsRamp(t *testing.T) {
	sim, warlock := newWarlockSim("", &proto.WarlockOptions{}, core.MakeSingleTargetEncounter(0))
	target := sim.Encounter.AllTargetUnits[0]
	warlock.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: -100})

	dot := warlock.BaneOfAgony.Dot(target)
	dot.Apply(sim)
	start, period := sim.CurrentTime, dot.TickPeriod()
	sim.AddPendingAction(&core.PendingAction{NextActionAt: start + period, Priority: core.ActionPriorityDOT, OnAction: dot.Apply})

	var ticks []dotTick
	metrics := &warlock.BaneOfAgony.SpellMetrics[target.UnitIndex]
	for len(ticks) < 2 {
		damage := metrics.TotalDamage
		if sim.Step() {
			break
		}
		if metrics.TotalDamage > damage {
			ticks = append(ticks, dotTick{sim.CurrentTime, metrics.TotalDamage - damage})
		}
	}
	if got, want := tickTimes(ticks, start), []time.Duration{period, 2 * period}; !slices.Equal(got, want) {
		t.Fatalf("Bane of Agony ticked at %v, want %v", got, want)
	}
	if got := ticks[0].damage / ticks[1].damage; math.Abs(got-1) > 1e-9 {
		t.Errorf("the tick dealt as the recast ended the old Bane is %.4f times the new Bane's first tick, want 1", got)
	}
}

// The Imp's Blood Pact is up on the party's players while the Imp is out, counts in the stats with the Imp out at
// the start, and drops when the Imp is dismissed or sacrificed. A Blood Pact ticked in the party buffs adds nothing
// while the Imp's is up and takes over once it drops. With two Imps in the party one Blood Pact counts, and it
// stays up while either Imp is out.
func TestBloodPactLastsWhileTheImpIsOut(t *testing.T) {
	bloodPact := buffs.BloodPactValue(0)
	encounter := core.MakeSingleTargetEncounter(0)
	imp := &proto.WarlockOptions{Summon: proto.WarlockOptions_Imp}
	voidwalker := &proto.WarlockOptions{Summon: proto.WarlockOptions_Voidwalker}
	bloodPactOf := func(member *Warlock, warlock *Warlock) *core.Aura {
		return member.GetAura("Blood Pact-" + warlock.Label)
	}
	party := func(players ...*proto.Player) *proto.Raid {
		return &proto.Raid{Parties: []*proto.Party{{Players: players, Buffs: &proto.PartyBuffs{}}}, Buffs: &proto.RaidBuffs{}, Debuffs: &proto.Debuffs{}}
	}
	_, alone := newWarlockSim("", imp, encounter)
	withImp := alone.GetStat(stats.Stamina)

	t.Run("summon and dismiss", func(t *testing.T) {
		sim, warlock := newWarlockSim("", imp, encounter)
		aura := bloodPactOf(warlock, warlock)
		if aura == nil {
			t.Fatalf("A warlock with an Imp registers no Blood Pact of its own")
		}
		if !aura.IsActive() {
			t.Fatalf("Blood Pact is not up with the Imp out at the start")
		}

		warlock.summonDemon(sim, warlock.Voidwalker)
		if aura.IsActive() {
			t.Errorf("Blood Pact is still up after the Voidwalker replaces the Imp")
		}
		if got := withImp - warlock.GetStat(stats.Stamina); math.Abs(got-bloodPact) > 1e-9 {
			t.Errorf("dismissing the Imp lowers Stamina by %.2f, want Blood Pact's %.2f", got, bloodPact)
		}

		warlock.summonDemon(sim, warlock.Imp)
		if !aura.IsActive() {
			t.Errorf("Blood Pact is not up after the Imp is summoned again")
		}
	})

	t.Run("ticked takes over", func(t *testing.T) {
		raid := core.SinglePlayerRaidProto(warlockPlayer("Warlock", "", imp), &proto.PartyBuffs{BloodPact: true}, &proto.RaidBuffs{}, &proto.Debuffs{})
		sim := newWarlockRaidSim(raid, encounter)
		ticked := sim.Raid.Parties[0].Players[0].(*Warlock)
		if got := ticked.GetStat(stats.Stamina); math.Abs(got-withImp) > 1e-9 {
			t.Errorf("Stamina %.2f with the Imp and a ticked Blood Pact, want the Imp's alone %.2f", got, withImp)
		}
		ticked.summonDemon(sim, ticked.Voidwalker)
		if got := ticked.GetStat(stats.Stamina); math.Abs(got-withImp) > 1e-9 {
			t.Errorf("Stamina %.2f after the Imp leaves with a ticked Blood Pact, want the ticked one's %.2f", got, withImp)
		}
	})

	t.Run("sacrifice", func(t *testing.T) {
		sim, sacrificer := newWarlockSim("-0000000001", imp, encounter)
		if !sacrificer.Talents.DemonicSacrifice {
			t.Fatalf("the talents string takes no Demonic Sacrifice")
		}
		if !sacrificer.DemonicSacrifice.Cast(sim, &sacrificer.Unit) {
			t.Fatalf("Demonic Sacrifice cannot be cast with the Imp out")
		}
		if bloodPactOf(sacrificer, sacrificer).IsActive() {
			t.Errorf("Blood Pact is still up after the Imp is sacrificed")
		}
	})

	t.Run("sheet", func(t *testing.T) {
		sheetStamina := func(options *proto.WarlockOptions) float64 {
			raid := core.SinglePlayerRaidProto(warlockPlayer("Warlock", "", options), &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
			result := core.ComputeStats(&proto.ComputeStatsRequest{Raid: raid, Encounter: encounter})
			return result.RaidStats.Parties[0].Players[0].FinalStats.Stats[stats.Stamina]
		}
		if got := sheetStamina(imp) - sheetStamina(voidwalker); math.Abs(got-bloodPact) > 1e-9 {
			t.Errorf("the Imp out at the start adds %.2f Stamina to the stats, want Blood Pact's %.2f", got, bloodPact)
		}
	})

	t.Run("party member", func(t *testing.T) {
		sim := newWarlockRaidSim(party(warlockPlayer("Warlock", "", imp), warlockPlayer("Second Warlock", "", voidwalker)), encounter)
		owner, member := sim.Raid.Parties[0].Players[0].(*Warlock), sim.Raid.Parties[0].Players[1].(*Warlock)
		if !bloodPactOf(member, owner).IsActive() {
			t.Errorf("Blood Pact is not up on a party member with the Imp out")
		}
		owner.summonDemon(sim, owner.Voidwalker)
		if bloodPactOf(member, owner).IsActive() {
			t.Errorf("Blood Pact is still up on a party member after the Imp is dismissed")
		}
	})

	t.Run("two warlocks", func(t *testing.T) {
		sim := newWarlockRaidSim(party(warlockPlayer("Warlock", "", imp), warlockPlayer("Second Warlock", "", imp)), encounter)
		first, second := sim.Raid.Parties[0].Players[0].(*Warlock), sim.Raid.Parties[0].Players[1].(*Warlock)
		before := []float64{first.GetStat(stats.Stamina), second.GetStat(stats.Stamina)}

		first.summonDemon(sim, first.Voidwalker)
		for i, player := range []*Warlock{first, second} {
			if !bloodPactOf(player, second).IsActive() {
				t.Errorf("the second warlock's Blood Pact is not up on %s after the first dismisses its Imp", player.Label)
			}
			if got := player.GetStat(stats.Stamina); math.Abs(got-before[i]) > 1e-9 {
				t.Errorf("%s has %.2f Stamina after the first Imp leaves, want %.2f from the second's Blood Pact", player.Label, got, before[i])
			}
		}

		second.summonDemon(sim, second.Voidwalker)
		for i, player := range []*Warlock{first, second} {
			if got := before[i] - player.GetStat(stats.Stamina); math.Abs(got-bloodPact) > 1e-9 {
				t.Errorf("%s loses %.2f Stamina once both Imps are gone, want Blood Pact's %.2f", player.Label, got, bloodPact)
			}
		}
	})
}

// Bane Is Active On All Targets counts any of the warlock's Banes on each target that is up, and Assigned Curse
// Is Active On All Targets the assigned curse. A target that is down is skipped, and with no curse assigned the
// curse value reads false.
func TestBaneAndAssignedCurseOnAllTargets(t *testing.T) {
	down := core.FreshDefaultTargetConfig()
	down.DisabledAtStart = true
	encounter := &proto.Encounter{Duration: core.LongDuration, Targets: []*proto.Target{core.NewDefaultTarget(), core.NewDefaultTarget(), down}}
	baneValue := &proto.APLValue{Value: &proto.APLValue_WarlockBaneIsActiveOnAllTargets{WarlockBaneIsActiveOnAllTargets: &proto.APLValueWarlockBaneIsActiveOnAllTargets{}}}
	curseValue := &proto.APLValue{Value: &proto.APLValue_WarlockAssignedCurseIsActiveOnAllTargets{WarlockAssignedCurseIsActiveOnAllTargets: &proto.APLValueWarlockAssignedCurseIsActiveOnAllTargets{}}}

	// Bane of Havoc is the 13th destruction talent.
	sim, warlock := newWarlockSim("--0000000000001", &proto.WarlockOptions{CurseOptions: proto.WarlockOptions_Elements}, encounter)
	first, second := sim.Encounter.AllTargetUnits[0], sim.Encounter.AllTargetUnits[1]
	bane, curse := warlock.NewAPLValue(warlock.Rotation, baneValue), warlock.NewAPLValue(warlock.Rotation, curseValue)

	warlock.BaneOfDoom.Dot(first).Apply(sim)
	if bane.GetBool(sim) {
		t.Errorf("Bane Is Active On All Targets is true with Doom on the first target alone")
	}
	warlock.BaneOfAgony.Dot(second).Apply(sim)
	if !bane.GetBool(sim) {
		t.Errorf("Bane Is Active On All Targets is false with Doom on the first target, Agony on the second and the third down")
	}
	warlock.BaneOfAgony.Dot(second).Deactivate(sim)
	warlock.BaneOfHavocAuras.Get(second).Activate(sim)
	if !bane.GetBool(sim) {
		t.Errorf("Bane Is Active On All Targets is false with Havoc in place of Agony on the second target")
	}

	warlock.CurseOfElementsAuras.Get(first).Activate(sim)
	if curse.GetBool(sim) {
		t.Errorf("Assigned Curse Is Active On All Targets is true with Curse of the Elements on the first target alone")
	}
	warlock.CurseOfElementsAuras.Get(second).Activate(sim)
	if !curse.GetBool(sim) {
		t.Errorf("Assigned Curse Is Active On All Targets is false with Curse of the Elements on both targets that are up")
	}

	sim, warlock = newWarlockSim("", &proto.WarlockOptions{CurseOptions: proto.WarlockOptions_NoCurse}, encounter)
	if warlock.NewAPLValue(warlock.Rotation, curseValue).GetBool(sim) {
		t.Errorf("Assigned Curse Is Active On All Targets is true with no curse assigned")
	}
}

// A reset sim of one warlock against two targets, so the first can die while the second stays up.
func newTwoTargetWarlockSim(talents string, options *proto.WarlockOptions) (*core.Simulation, *Warlock, *core.Target, *core.Target) {
	encounter := &proto.Encounter{Duration: core.LongDuration, Targets: []*proto.Target{core.FreshDefaultTargetConfig(), core.FreshDefaultTargetConfig()}}
	sim, warlock := newWarlockSim(talents, options, encounter)
	return sim, warlock, sim.Encounter.AllTargets[0], sim.Encounter.AllTargets[1]
}

// A target that dies with Drain Soul's channel on it returns a Soul Shard, whoever gets the kill; a death
// with the channel on another target returns none.
func TestDrainSoulShardWhenTheChannelledTargetDies(t *testing.T) {
	sim, warlock, first, _ := newTwoTargetWarlockSim("", &proto.WarlockOptions{})
	warlock.DrainSoul.Dot(&first.Unit).Apply(sim)
	first.Die(sim, 0)
	stepFor(sim, 2*core.SpellBatchWindow)
	if got := warlock.SoulShardsAura.GetStacks(); got != 1 {
		t.Errorf("%d Soul Shards after the channelled target died with no kill credit, want 1", got)
	}

	sim, warlock, first, second := newTwoTargetWarlockSim("", &proto.WarlockOptions{})
	warlock.DrainSoul.Dot(&second.Unit).Apply(sim)
	first.Die(sim, 1)
	stepFor(sim, 2*core.SpellBatchWindow)
	if got := warlock.SoulShardsAura.GetStacks(); got != 0 {
		t.Errorf("%d Soul Shards after a kill with the channel on the other target, want 0", got)
	}
}

// A warlock carries at most 32 Soul Shards: more carried into the fight count as 32, and a shard returned
// at 32 is lost.
func TestSoulShardsCapAt32(t *testing.T) {
	sim, warlock, first, _ := newTwoTargetWarlockSim("", &proto.WarlockOptions{SoulShards: 40})
	if got := warlock.SoulShardsAura.GetStacks(); got != 32 {
		t.Fatalf("%d Soul Shards carried in from 40, want 32", got)
	}
	warlock.DrainSoul.Dot(&first.Unit).Apply(sim)
	first.Die(sim, 0)
	stepFor(sim, 2*core.SpellBatchWindow)
	if got := warlock.SoulShardsAura.GetStacks(); got != 32 {
		t.Errorf("%d Soul Shards after a shard returned at 32, want 32", got)
	}
}

// A target that dies within Shadowburn's debuff window returns a Soul Shard, whoever gets the kill; once the
// window has passed, or for the same target enabled again and dying with no Shadowburn of its own, its death
// returns none. Shadowburn is the 8th destruction talent.
func TestShadowburnShardWhenTheTargetDiesInTheWindow(t *testing.T) {
	shadowburnOn := func() (*core.Simulation, *Warlock, *core.Target) {
		sim, warlock, first, _ := newTwoTargetWarlockSim("--00000001", &proto.WarlockOptions{SoulShards: 1})
		warlock.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100})
		warlock.Shadowburn.SkipCastAndApplyEffects(sim, &first.Unit)
		if got := warlock.SoulShardsAura.GetStacks(); got != 0 {
			t.Fatalf("%d Soul Shards after Shadowburn, want its reagent spent", got)
		}
		return sim, warlock, first
	}

	sim, warlock, first := shadowburnOn()
	stepFor(sim, time.Second)
	first.Die(sim, 0)
	stepFor(sim, 2*core.SpellBatchWindow)
	if got := warlock.SoulShardsAura.GetStacks(); got != 1 {
		t.Errorf("%d Soul Shards after the target died 1 sec after Shadowburn with no kill credit, want 1", got)
	}

	// The target comes back inside the first window and dies with no Shadowburn of its own.
	first.Enable(sim)
	stepFor(sim, time.Second)
	first.Die(sim, 0)
	stepFor(sim, 2*core.SpellBatchWindow)
	if got := warlock.SoulShardsAura.GetStacks(); got != 1 {
		t.Errorf("%d Soul Shards after the target enabled again died within the window of its first life's Shadowburn, want 1", got)
	}

	sim, warlock, first = shadowburnOn()
	stepFor(sim, spellData.Shadowburn.ByID(18871).Duration()+time.Second)
	first.Die(sim, 1)
	stepFor(sim, 2*core.SpellBatchWindow)
	if got := warlock.SoulShardsAura.GetStacks(); got != 0 {
		t.Errorf("%d Soul Shards after a kill past Shadowburn's window, want 0", got)
	}
}

// A credited kill of a target with Drain Soul on it grants Soul Harvest (1242853) and its mana regen; a death
// without the kill, or a kill without Drain Soul on the target, grants nothing. Soul Harvest 2/2 is the 5th
// affliction talent.
func TestSoulHarvestOnADrainSoulKill(t *testing.T) {
	killWith := func(drainSoul bool, killShare float64) (*Warlock, *core.Aura) {
		sim, warlock, first, _ := newTwoTargetWarlockSim("00002", &proto.WarlockOptions{})
		if drainSoul {
			warlock.DrainSoul.Dot(&first.Unit).Apply(sim)
		}
		first.Die(sim, killShare)
		stepFor(sim, 2*core.SpellBatchWindow)
		return warlock, warlock.GetAuraByID(core.ActionID{SpellID: 1242853})
	}

	if _, buff := killWith(true, 0); buff.IsActive() {
		t.Errorf("Soul Harvest is up after the Drain Soul target died with no kill credit")
	}
	if _, buff := killWith(false, 1); buff.IsActive() {
		t.Errorf("Soul Harvest is up after a kill with no Drain Soul on the target")
	}

	warlock, buff := killWith(true, 1)
	if !buff.IsActive() {
		t.Fatalf("Soul Harvest is not up after a credited kill of the Drain Soul target")
	}
	talent := spellData.SoulHarvest.Rank(2)
	if got, want := warlock.PseudoStats.SpiritRegenMultiplier, 1+talent.EffectN(1).Percent(); math.Abs(got-want) > 1e-9 {
		t.Errorf("mana regen multiplier %.3f under Soul Harvest 2/2, want %.3f", got, want)
	}
	if got, want := warlock.PseudoStats.SpiritRegenRateCasting, talent.EffectN(2).Percent(); math.Abs(got-want) > 1e-9 {
		t.Errorf("mana regen while casting %.3f under Soul Harvest 2/2, want %.3f", got, want)
	}
}

// Decimation's damage bonus counts against the target of each Shadow Bolt once that target is below 35% health,
// and not while it is above.
func TestDecimationBonusAgainstTheTargetBelow35(t *testing.T) {
	sim, warlock := newWarlockSim("-000000000002", &proto.WarlockOptions{}, core.MakeSingleTargetEncounter(0))
	target := sim.Encounter.AllTargetUnits[0]
	above := warlock.ShadowBolt.ExpectedInitialDamage(sim, target)
	for !target.IsExecutePhase35() && !sim.Step() {
	}
	want := 1 + spellData.Decimation.Rank(2).EffectN(4).Percent()
	if got := warlock.ShadowBolt.ExpectedInitialDamage(sim, target) / above; math.Abs(got-want) > 1e-9 {
		t.Errorf("Shadow Bolt against the target below 35%% is %.4f times that while it was above, want %.4f", got, want)
	}
}

// The trainer's top ranks of Shadow Bolt, Immolate and Corruption register beside the book ranks, and
// Conflagrate consumes an Immolate of the trainer's rank.
func TestTrainerRanksBesideBookRanks(t *testing.T) {
	sim, warlock := newWarlockSim("--0050355103101351", &proto.WarlockOptions{}, core.MakeSingleTargetEncounter(0))
	target := sim.Encounter.AllTargetUnits[0]
	for _, id := range []int32{11661, 25307, 11668, 25309, 11672, 25311} {
		if warlock.GetSpell(core.ActionID{SpellID: id}) == nil {
			t.Errorf("spell %d is not registered", id)
		}
	}

	// The dot is applied directly and Conflagrate retried, since a warlock with no hit rating can miss.
	warlock.conflagrateKeepsImmolate = 0
	immolate := warlock.GetSpell(core.ActionID{SpellID: 11668}).Dot(target)
	immolate.Apply(sim)
	if !warlock.Conflagrate.ExtraCastCondition(sim, target) {
		t.Fatalf("Conflagrate cannot be cast on Immolate rank 7")
	}
	for range 20 {
		if !immolate.IsActive() {
			break
		}
		warlock.Conflagrate.SkipCastAndApplyEffects(sim, target)
	}
	if immolate.IsActive() {
		t.Errorf("Immolate rank 7 still burns after 20 Conflagrates")
	}
}

// A dot whose row marks Periodic Can Crit expects its tick with the crit share in it, as the tick
// itself rolls it: more spell crit raises what Corruption's tick is expected to deal.
func TestExpectedDotTickCountsItsCrits(t *testing.T) {
	sim, warlock := newWarlockSim("", &proto.WarlockOptions{}, core.MakeSingleTargetEncounter(0))
	target := sim.Encounter.AllTargetUnits[0]
	if !spellData.Corruption.ByID(warlock.Corruption.SpellID).PeriodicCanCrit() {
		t.Fatal("Corruption's row no longer marks Periodic Can Crit, so the check below proves nothing")
	}

	plain := warlock.Corruption.ExpectedTickDamage(sim, target)
	warlock.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: 50})
	if raised := warlock.Corruption.ExpectedTickDamage(sim, target); raised <= plain {
		t.Errorf("Corruption's expected tick is %.3f with more spell crit and %.3f without, want it higher", raised, plain)
	}
}

// While Wrack is channelled on a target, this warlock's Corruption, Siphon Life and Bane of Agony there also tick on
// the half points of their period from their application: each half point from the instant the channel starts up
// to the instant it ends, both included. Their own ticks and their expiry stay, and every tick deals in full;
// Bane of Agony's ramp follows the time since it landed, a tick on a third's end taking the lower third, and each
// Siphon Life tick heals the warlock for its damage. The custom metric wrack holds the half-point ticks' damage,
// each tick's in full. The warlock's effect on another target and a second warlock's on the channelled target tick
// as they do without Wrack. The windows are in seconds from the effect's application, shaped like channels logged
// in game, and one with no end runs the channel out: one started on a half point ticks it at once, and a recast at
// the instant one ends ticks that point once. Siphon Life is the 14th affliction talent, Wrack the 17th.
func TestWrackTicksTheHalfPointsOfTheWarlocksDots(t *testing.T) {
	type window struct{ from, to float64 }
	corruption := func(w *Warlock) *core.Spell { return w.Corruption }
	siphonLife := func(w *Warlock) *core.Spell { return w.SiphonLife }
	agony := func(w *Warlock) *core.Spell { return w.BaneOfAgony }
	cases := []struct {
		name    string
		spell   func(*Warlock) *core.Spell
		period  float64 // The effect's period and duration in seconds, which the ticks are written for.
		dur     float64
		windows []window
		want    []float64
	}{
		{name: "Corruption, Wrack from 3.2 s to 5.94 s", spell: corruption, period: 3, dur: 18,
			windows: []window{{3.2, 5.94}}, want: []float64{3, 4.5, 6, 9, 12, 15, 18}},
		{name: "Corruption, Wrack from 3.36 s to 7.84 s", spell: corruption, period: 3, dur: 18,
			windows: []window{{3.36, 7.84}}, want: []float64{3, 4.5, 6, 7.5, 9, 12, 15, 18}},
		{name: "Corruption, Wrack from 3.43 s to 9.26 s and 9.55 s to 9.75 s", spell: corruption, period: 3, dur: 18,
			windows: []window{{3.43, 9.26}, {9.55, 9.75}}, want: []float64{3, 4.5, 6, 7.5, 9, 12, 15, 18}},
		{name: "Corruption, Wrack run out from its tick at 3 s", spell: corruption, period: 3, dur: 18,
			windows: []window{{3, 0}}, want: []float64{3, 4.5, 6, 7.5, 9, 12, 15, 18}},
		{name: "Siphon Life, Wrack from 1.66 s to 7.5 s", spell: siphonLife, period: 3, dur: 30,
			windows: []window{{1.66, 7.5}}, want: []float64{3, 4.5, 6, 7.5, 9, 12, 15, 18, 21, 24, 27, 30}},
		{name: "Siphon Life, Wrack one global cooldown after it, run out and recast as it ends", spell: siphonLife, period: 3, dur: 30,
			windows: []window{{1.5, 0}, {7.5, 0}}, want: []float64{1.5, 3, 4.5, 6, 7.5, 9, 10.5, 12, 13.5, 15, 18, 21, 24, 27, 30}},
		{name: "Bane of Agony, Wrack from 4.84 s to 9.32 s", spell: agony, period: 2, dur: 24,
			windows: []window{{4.84, 9.32}}, want: []float64{2, 4, 5, 6, 7, 8, 9, 10, 12, 14, 16, 18, 20, 22, 24}},
		{name: "Bane of Agony, Wrack from 14.6 s to 17.4 s", spell: agony, period: 2, dur: 24,
			windows: []window{{14.6, 17.4}}, want: []float64{2, 4, 6, 8, 10, 12, 14, 15, 16, 17, 18, 20, 22, 24}},
		{name: "Bane of Agony, Wrack run out from 3 s, an odd second after it", spell: agony, period: 2, dur: 24,
			windows: []window{{3, 0}}, want: []float64{2, 3, 4, 5, 6, 7, 8, 9, 10, 12, 14, 16, 18, 20, 22, 24}},
	}
	seconds := func(s float64) time.Duration { return time.Duration(math.Round(s*1000)) * time.Millisecond }
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			encounter := &proto.Encounter{Duration: core.LongDuration, Targets: []*proto.Target{core.FreshDefaultTargetConfig(), core.FreshDefaultTargetConfig()}}
			raid := &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{
				warlockPlayer("Warlock", "00000000000001001", &proto.WarlockOptions{}),
				warlockPlayer("Second Warlock", "00000000000001001", &proto.WarlockOptions{}),
			}, Buffs: &proto.PartyBuffs{}}}, Buffs: &proto.RaidBuffs{}, Debuffs: &proto.Debuffs{}}
			sim := newWarlockRaidSim(raid, encounter)
			warlock, second := sim.Raid.Parties[0].Players[0].(*Warlock), sim.Raid.Parties[0].Players[1].(*Warlock)
			target, otherTarget := sim.Encounter.AllTargetUnits[0], sim.Encounter.AllTargetUnits[1]
			// A crit would hide whether a tick is a full one.
			warlock.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: -100})

			spell := c.spell(warlock)
			dots := []*core.Dot{spell.Dot(target), spell.Dot(otherTarget), c.spell(second).Dot(target)}
			for _, dot := range dots {
				dot.Apply(sim)
			}
			start, period := sim.CurrentTime, dots[0].TickPeriod()
			if period != seconds(c.period) || dots[0].ExpiresAt()-start != seconds(c.dur) {
				t.Fatalf("%s ticks every %v for %v, the case is written for %vs and %vs", c.name, period, dots[0].ExpiresAt()-start, c.period, c.dur)
			}
			if fast := wrackPeriod(period); fast != period/2 {
				t.Fatalf("Wrack ticks the effect every %v, the case is written for half its period", fast)
			}
			wrack := warlock.Wrack.Dot(target)
			for _, w := range c.windows {
				sim.AddPendingAction(&core.PendingAction{NextActionAt: start + seconds(w.from), OnAction: wrack.Apply})
				if w.to > 0 {
					sim.AddPendingAction(&core.PendingAction{NextActionAt: start + seconds(w.to), OnAction: wrack.Deactivate})
				}
			}
			ticks := recordDotTicks(sim, dots...)
			sim.Cleanup()

			sped := ticks[0]
			if len(sped) != len(c.want) {
				t.Fatalf("ticked at %v, want %v", tickTimes(sped, start), c.want)
			}
			var halves int32
			var halvesDamage float64
			for i, tick := range sped {
				elapsed := tick.at - start
				if elapsed != seconds(c.want[i]) {
					t.Errorf("tick %d at %v, want %vs", i+1, elapsed, c.want[i])
				}
				if elapsed%period != 0 {
					halves++
					halvesDamage += tick.damage
				}
				want := 1.0
				if c.spell(warlock) == warlock.BaneOfAgony {
					// The ramp's thirds from the first tick's: once, twice or three times it.
					switch duration := seconds(c.dur); {
					case 3*elapsed > 2*duration:
						want = 3
					case 3*elapsed > duration:
						want = 2
					}
				}
				if got := tick.damage / sped[0].damage; math.Abs(got-want) > 1e-9 {
					t.Errorf("tick %d at %v dealt %.4f times the first tick, want %.0f", i+1, elapsed, got, want)
				}
			}

			got := wrackMetric(warlock, spell, target)
			if got == nil {
				t.Fatalf("recorded no Wrack contribution, want %.3f over %d ticks", halvesDamage, halves)
			}
			if total := got.Avg * float64(got.AggregatorData.N); got.AggregatorData.N != halves || math.Abs(total-halvesDamage) > 1e-6 {
				t.Errorf("Wrack contributed %.3f over %d ticks, want the half-point ticks' %.3f over %d", total, got.AggregatorData.N, halvesDamage, halves)
			}

			for i, name := range []string{"the warlock's effect on the other target", "the second warlock's effect on the channelled target"} {
				var want []time.Duration
				for at := period; at <= seconds(c.dur); at += period {
					want = append(want, at)
				}
				if got := tickTimes(ticks[i+1], start); !slices.Equal(got, want) {
					t.Errorf("%s ticked at %v, want %v", name, got, want)
				}
			}
			if got := wrackMetric(warlock, spell, otherTarget); got != nil {
				t.Errorf("the warlock's effect on the other target recorded a Wrack contribution over %d ticks, want none", got.AggregatorData.N)
			}
			if spell == warlock.SiphonLife {
				var heals int32
				var healed, dealt float64
				for _, resource := range warlock.GetMetricsProto().Resources {
					if resource.Type == proto.ResourceType_ResourceTypeHealth && resource.Id.GetSpellId() == spell.SpellID {
						heals, healed = heals+resource.Events, healed+resource.Gain
					}
				}
				for _, tick := range append(slices.Clone(ticks[0]), ticks[1]...) {
					dealt += tick.damage
				}
				if want := int32(len(ticks[0]) + len(ticks[1])); heals != want || math.Abs(healed-dealt) > 1e-6 {
					t.Errorf("Siphon Life healed %.3f over %d heals, want the %.3f its %d ticks dealt", healed, heals, dealt, want)
				}
			}
			if got := wrackMetric(second, c.spell(second), target); got != nil {
				t.Errorf("the second warlock's effect recorded a Wrack contribution over %d ticks, want none", got.AggregatorData.N)
			}
		})
	}
}

// The ticks' times from the effect's application.
func tickTimes(ticks []dotTick, start time.Duration) []time.Duration {
	return core.MapSlice(ticks, func(tick dotTick) time.Duration { return tick.at - start })
}

type dotTick struct {
	at     time.Duration
	damage float64
}

// Steps the sim until every dot has ended, recording each dot's ticks and the damage its spell dealt the
// dot's target in the step.
func recordDotTicks(sim *core.Simulation, dots ...*core.Dot) [][]dotTick {
	ticks := make([][]dotTick, len(dots))
	for running := true; running; {
		counts, damages := make([]int32, len(dots)), make([]float64, len(dots))
		for i, dot := range dots {
			counts[i], damages[i] = dot.TickCount(), dot.Spell.SpellMetrics[dot.Unit.UnitIndex].TotalDamage
		}
		if sim.Step() {
			break
		}
		running = false
		for i, dot := range dots {
			if dot.TickCount() > counts[i] {
				ticks[i] = append(ticks[i], dotTick{sim.CurrentTime, dot.Spell.SpellMetrics[dot.Unit.UnitIndex].TotalDamage - damages[i]})
			}
			running = running || dot.IsActive()
		}
	}
	return ticks
}

// The period Wrack gives a dot that otherwise ticks every period.
func wrackPeriod(period time.Duration) time.Duration {
	return time.Duration(float64(period) / (1 + spellData.Wrack.ByID(1316697).EffectN(2).Percent())).Round(time.Millisecond)
}

// Wrack's channel starting or ending at a time in seconds from Corruption's application, queued to run before
// Corruption's own tick due at that instant (first) or after it.
type wrackEvent struct {
	at    float64
	start bool
	first bool
}

// A warlock with Wrack and no crit, the warlock's Corruption on the target, Wrack's channel there, the time
// Corruption was applied, and a check to run on Corruption's ticks once it has ended. Each start and end of the
// channel with Corruption up keeps its application, duration and expiry, and the last leaves it the ticks it deals
// later.
func newWrackOnCorruption(t *testing.T) (*core.Simulation, *Warlock, *core.Dot, *core.Dot, time.Duration, func([]dotTick)) {
	t.Helper()
	sim, warlock := newWarlockSim("00000000000001001", &proto.WarlockOptions{}, core.MakeSingleTargetEncounter(0))
	target := sim.Encounter.AllTargetUnits[0]
	warlock.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: -100})
	corruption, wrack := warlock.Corruption.Dot(target), warlock.Wrack.Dot(target)
	// Applied off the warlock's mana ticks, every 0.1 s, so at Corruption's ticks nothing else is due.
	sim.AddPendingAction(&core.PendingAction{NextActionAt: 137 * time.Millisecond, OnAction: corruption.Apply})
	stepFor(sim, 137*time.Millisecond)
	start := sim.CurrentTime
	if corruption.TickPeriod() != 3*time.Second || corruption.ExpiresAt()-start != 18*time.Second || wrack.Duration != 6*time.Second {
		t.Fatalf("Corruption ticks every %v for %v and Wrack runs %v, the cases are written for 3 s, 18 s and 6 s",
			corruption.TickPeriod(), corruption.ExpiresAt()-start, wrack.Duration)
	}

	retimed, lastAt, remaining := false, time.Duration(0), int32(0)
	watch := func(sim *core.Simulation, retime func()) {
		if !corruption.IsActive() {
			retime()
			if corruption.IsActive() {
				t.Errorf("at %v the channel put Corruption up, which was down", sim.CurrentTime-start)
			}
			return
		}
		startedAt, duration, expiresAt := corruption.StartedAt(), corruption.Duration, corruption.ExpiresAt()
		retime()
		if corruption.StartedAt() != startedAt || corruption.Duration != duration || corruption.ExpiresAt() != expiresAt {
			t.Errorf("at %v Corruption started at %v for %v to %v, want it unchanged from %v for %v to %v", sim.CurrentTime-start,
				corruption.StartedAt()-start, corruption.Duration, corruption.ExpiresAt()-start, startedAt-start, duration, expiresAt-start)
		}
		retimed, lastAt, remaining = true, sim.CurrentTime, corruption.RemainingTicks()
	}
	onGain, onExpire := wrack.OnGain, wrack.OnExpire
	wrack.OnGain = func(aura *core.Aura, sim *core.Simulation) { watch(sim, func() { onGain(aura, sim) }) }
	wrack.OnExpire = func(aura *core.Aura, sim *core.Simulation) { watch(sim, func() { onExpire(aura, sim) }) }

	return sim, warlock, corruption, wrack, start, func(ticks []dotTick) {
		t.Helper()
		later := int32(0)
		for _, tick := range ticks {
			if tick.at > lastAt {
				later++
			}
		}
		if retimed && remaining != later {
			t.Errorf("Corruption had %d ticks remaining after the channel's last start or end at %v, want the %d it dealt later", remaining, lastAt-start, later)
		}
	}
}

// Queues the events and returns, for each, whether Corruption's tick at its instant was still to deal when it ran.
func queueWrackEvents(sim *core.Simulation, corruption, wrack *core.Dot, start time.Duration, events []wrackEvent) []bool {
	due := make([]bool, len(events))
	for i, e := range events {
		priority := core.ActionPriorityLow
		if e.first {
			priority = core.ActionPriorityDOT
		}
		at := start + time.Duration(math.Round(e.at*1000))*time.Millisecond
		sim.AddPendingAction(&core.PendingAction{NextActionAt: at, Priority: priority, OnAction: func(sim *core.Simulation) {
			due[i] = corruption.NextTickAt() == sim.CurrentTime
			if e.start {
				wrack.Apply(sim)
			} else {
				wrack.Deactivate(sim)
			}
		}})
	}
	return due
}

// Checks the ticks against their times in seconds from the application, each a full tick.
func checkFullTicks(t *testing.T, ticks []dotTick, start time.Duration, want []float64) {
	t.Helper()
	seconds := func(s float64) time.Duration { return time.Duration(math.Round(s*1000)) * time.Millisecond }
	if got := tickTimes(ticks, start); !slices.Equal(got, core.MapSlice(want, seconds)) {
		t.Fatalf("Corruption ticked at %v, want %v", got, want)
	}
	for i, tick := range ticks {
		if got := tick.damage / ticks[0].damage; math.Abs(got-1) > 1e-9 {
			t.Errorf("tick %d at %v dealt %.4f times the first tick, want a full tick", i+1, tick.at-start, got)
		}
	}
}

// Wrack's channel starting or ending at an instant Corruption has a tick due deals that tick once, whether the event
// runs before the tick's own action or after it, and the next tick falls on the grid of the period the channel
// leaves. A channel recast at the half point another one ended on ticks that point once, also when the end runs
// before the tick and the recast after it.
func TestWrackOnADueTickDealsItOnce(t *testing.T) {
	cases := []struct {
		name   string
		events func(first bool) []wrackEvent
		at     int // The event at the instant of the tick.
		want   []float64
	}{
		{
			name:   "starts on the tick at 3 s",
			events: func(first bool) []wrackEvent { return []wrackEvent{{at: 3, start: true, first: first}} },
			want:   []float64{3, 4.5, 6, 7.5, 9, 12, 15, 18},
		},
		{
			name: "ends on the half point 7.5 s",
			events: func(first bool) []wrackEvent {
				return []wrackEvent{{at: 3.2, start: true}, {at: 7.5, first: first}}
			},
			at:   1,
			want: []float64{3, 4.5, 6, 7.5, 9, 12, 15, 18},
		},
		{
			name: "ends on the half point 7.5 s and is recast there",
			events: func(first bool) []wrackEvent {
				return []wrackEvent{{at: 3.2, start: true}, {at: 7.5, first: first}, {at: 7.5, start: true, first: first}}
			},
			at:   1,
			want: []float64{3, 4.5, 6, 7.5, 9, 10.5, 12, 13.5, 15, 18},
		},
		{
			name: "ends on the half point 7.5 s before the tick and is recast there after it",
			events: func(first bool) []wrackEvent {
				return []wrackEvent{{at: 3.2, start: true}, {at: 7.5, first: first}, {at: 7.5, start: true}}
			},
			at:   1,
			want: []float64{3, 4.5, 6, 7.5, 9, 10.5, 12, 13.5, 15, 18},
		},
	}
	for _, c := range cases {
		for _, first := range []bool{true, false} {
			if c.name == cases[3].name && !first {
				continue
			}
			t.Run(c.name+map[bool]string{true: ", before the tick", false: ", after the tick"}[first], func(t *testing.T) {
				sim, _, corruption, wrack, start, checkRemaining := newWrackOnCorruption(t)
				due := queueWrackEvents(sim, corruption, wrack, start, c.events(first))
				ticks := recordDotTicks(sim, corruption)[0]
				sim.Cleanup()

				if due[c.at] != first {
					t.Fatalf("Corruption's tick was still to deal at the event: %v, want %v", due[c.at], first)
				}
				checkFullTicks(t, ticks, start, c.want)
				checkRemaining(ticks)
			})
		}
	}
}

// Wrack's channel running out on Corruption's half point 7.5 s deals that tick once, also when a channel is recast
// there: with nothing else due at that instant, so the sim takes the tick's action off its queue before it handles
// the channel's expiry, and with another action due first, so the tick's action is still queued then.
func TestWrackRunningOutOnATickDealsItOnce(t *testing.T) {
	for _, recast := range []bool{false, true} {
		for _, otherFirst := range []bool{false, true} {
			name := map[bool]string{false: "run out", true: "run out and recast"}[recast] +
				map[bool]string{false: ", nothing else due", true: ", another action due first"}[otherFirst]
			t.Run(name, func(t *testing.T) {
				sim, _, corruption, wrack, start, checkRemaining := newWrackOnCorruption(t)
				events := []wrackEvent{{at: 1.5, start: true}}
				want := []float64{1.5, 3, 4.5, 6, 7.5, 9, 12, 15, 18}
				if recast {
					events = append(events, wrackEvent{at: 7.5, start: true, first: otherFirst})
					want = []float64{1.5, 3, 4.5, 6, 7.5, 9, 10.5, 12, 13.5, 15, 18}
				} else if otherFirst {
					sim.AddPendingAction(&core.PendingAction{NextActionAt: start + 7500*time.Millisecond, Priority: core.ActionPriorityDOT, OnAction: func(_ *core.Simulation) {}})
				}
				queueWrackEvents(sim, corruption, wrack, start, events)
				var dueAtEnd []bool
				onExpire := wrack.OnExpire
				wrack.OnExpire = func(aura *core.Aura, sim *core.Simulation) {
					dueAtEnd = append(dueAtEnd, corruption.NextTickAt() == sim.CurrentTime)
					onExpire(aura, sim)
				}
				ticks := recordDotTicks(sim, corruption)[0]
				sim.Cleanup()

				if len(dueAtEnd) == 0 || !dueAtEnd[0] {
					t.Fatalf("Corruption's tick at 7.5 s was not still to deal as the channel ran out (%v), so the case is not reached", dueAtEnd)
				}
				checkFullTicks(t, ticks, start, want)
				checkRemaining(ticks)
			})
		}
	}
}

// Wrack's channel running out at the instant Corruption expires, with its expiry handled while Corruption is still
// up, leaves Corruption's last tick at the expiry dealt once, in full.
func TestWrackEndingAtTheExpiryDealsTheLastTickOnce(t *testing.T) {
	sim, warlock, corruption, wrack, start, checkRemaining := newWrackOnCorruption(t)
	// The channel's expiry is handled before Corruption's at an instant both expire when the channel sits before
	// Corruption on the target: the curse taken off swaps the channel into its place.
	curse := warlock.CurseOfRecklessnessAuras.Get(corruption.Unit)
	corruption.Deactivate(sim)
	curse.Activate(sim)
	corruption.Apply(sim)
	sim.AddPendingAction(&core.PendingAction{NextActionAt: start + 12*time.Second, OnAction: func(sim *core.Simulation) {
		wrack.Apply(sim)
		curse.Deactivate(sim)
	}})
	upAtEnd := false
	onExpire := wrack.OnExpire
	wrack.OnExpire = func(aura *core.Aura, sim *core.Simulation) {
		upAtEnd = corruption.IsActive() && corruption.ExpiresAt() == sim.CurrentTime
		onExpire(aura, sim)
	}
	ticks := recordDotTicks(sim, corruption)[0]
	sim.Cleanup()

	if !upAtEnd {
		t.Fatal("Corruption was not up at its expiry as the channel ran out, so the case is not reached")
	}
	checkFullTicks(t, ticks, start, []float64{3, 6, 9, 12, 13.5, 15, 16.5, 18})
	checkRemaining(ticks)
}

// Wrack's channel leaves an effect that is not up alone: Siphon Life and Bane of Agony, never applied, stay down
// through a channel, and so does Corruption once it has expired.
func TestWrackLeavesAnEffectThatIsNotUpAlone(t *testing.T) {
	sim, warlock, corruption, wrack, start, checkRemaining := newWrackOnCorruption(t)
	target := corruption.Unit
	queueWrackEvents(sim, corruption, wrack, start, []wrackEvent{{at: 3.2, start: true}, {at: 5}, {at: 20, start: true}, {at: 23}})
	ticks := recordDotTicks(sim, corruption)[0]
	stepFor(sim, 10*time.Second)
	sim.Cleanup()

	checkFullTicks(t, ticks, start, []float64{3, 4.5, 6, 9, 12, 15, 18})
	checkRemaining(ticks)
	if corruption.IsActive() {
		t.Error("Corruption is up again after a channel started once it had expired")
	}
	for _, spell := range []*core.Spell{warlock.SiphonLife, warlock.BaneOfAgony} {
		if spell.Dot(target).IsActive() || spell.SpellMetrics[target.UnitIndex].TotalDamage != 0 {
			t.Errorf("%s is up or dealt damage after channels with it never applied", spell.ActionID)
		}
	}
}

// With Wrack learnt and not channelled, the warlock's Corruption, Siphon Life and Bane of Agony record no Wrack
// contribution. Siphon Life is the 14th affliction talent, Wrack the 17th.
func TestNoWrackContributionWithoutTheChannel(t *testing.T) {
	sim, warlock := newWarlockSim("00000000000001001", &proto.WarlockOptions{}, core.MakeSingleTargetEncounter(0))
	target := sim.Encounter.AllTargetUnits[0]
	spells := []*core.Spell{warlock.Corruption, warlock.SiphonLife, warlock.BaneOfAgony}
	dots := make([]*core.Dot, len(spells))
	for i, spell := range spells {
		dots[i] = spell.Dot(target)
		dots[i].Apply(sim)
	}
	recordDotTicks(sim, dots...)
	sim.Cleanup()

	for _, spell := range spells {
		if spell.SpellMetrics[target.UnitIndex].TotalDamage <= 0 {
			t.Fatalf("%s dealt no damage, so the check below proves nothing", spell.ActionID)
		}
		if got := wrackMetric(warlock, spell, target); got != nil {
			t.Errorf("%s recorded a Wrack contribution over %d ticks with no Wrack channelled, want none", spell.ActionID, got.AggregatorData.N)
		}
	}
}

// A Corruption ticking under Wrack before the pull records no Wrack contribution, as its damage is not counted.
func TestNoWrackContributionBeforeThePull(t *testing.T) {
	sim, warlock := newWarlockSim("00000000000000001", &proto.WarlockOptions{}, core.MakeSingleTargetEncounter(0))
	target := sim.Encounter.AllTargetUnits[0]
	sim.CurrentTime = -20 * time.Second

	corruption := warlock.Corruption.Dot(target)
	corruption.Apply(sim)
	warlock.Wrack.Dot(target).Apply(sim)
	ticks := recordDotTicks(sim, corruption)[0]
	sim.Cleanup()

	if len(ticks) <= int(corruption.BaseTickCount) || ticks[len(ticks)-1].at >= 0 {
		t.Fatalf("Corruption ticked %d times, the last at %v, want more than its %d under Wrack, all before the pull", len(ticks), ticks[len(ticks)-1].at, corruption.BaseTickCount)
	}
	if got := wrackMetric(warlock, warlock.Corruption, target); got != nil {
		t.Errorf("Corruption recorded a Wrack contribution over %d ticks before the pull, want none", got.AggregatorData.N)
	}
}

// The Wrack contribution the spell recorded on the target, read from the warlock's metrics once the iteration is
// done (sim.Cleanup), or nil when it recorded none.
func wrackMetric(warlock *Warlock, spell *core.Spell, target *core.Unit) *proto.DistributionMetrics {
	for _, action := range warlock.GetMetricsProto().Actions {
		if action.Id.GetSpellId() != spell.SpellID || action.Id.GetTag() != spell.Tag {
			continue
		}
		for _, targetMetrics := range action.Targets {
			if targetMetrics.UnitIndex != target.UnitIndex {
				continue
			}
			for _, custom := range targetMetrics.CustomMetrics {
				if custom.Key == "wrack" {
					return custom.Range
				}
			}
		}
	}
	return nil
}

// Soul Siphon raises Drain Life by its share per Corruption, Siphon Life and Bane of Agony of the warlock's on the
// target, and not for Bane of Doom or a curse. Siphon Life is the 14th affliction talent and Soul Siphon 3/3 the 15th.
func TestSoulSiphonCountsCorruptionSiphonLifeAndAgony(t *testing.T) {
	sim, warlock := newWarlockSim("000000000000013", &proto.WarlockOptions{CurseOptions: proto.WarlockOptions_Recklessness}, core.MakeSingleTargetEncounter(0))
	target := sim.Encounter.AllTargetUnits[0]
	plain := warlock.DrainLife.ExpectedTickDamage(sim, target)

	warlock.BaneOfDoom.Dot(target).Apply(sim)
	warlock.CurseOfRecklessnessAuras.Get(target).Activate(sim)
	if got := warlock.DrainLife.ExpectedTickDamage(sim, target) / plain; math.Abs(got-1) > 1e-9 {
		t.Errorf("Drain Life is %.4f times its plain tick with Bane of Doom and Curse of Recklessness on the target, want 1", got)
	}

	warlock.BaneOfDoom.Dot(target).Deactivate(sim)
	for _, spell := range []*core.Spell{warlock.Corruption, warlock.SiphonLife, warlock.BaneOfAgony} {
		spell.Dot(target).Apply(sim)
	}
	want := 1 + 3*spellData.SoulSiphon.FractionAt(3)
	if got := warlock.DrainLife.ExpectedTickDamage(sim, target) / plain; math.Abs(got-want) > 1e-9 {
		t.Errorf("Drain Life is %.4f times its plain tick with Corruption, Siphon Life and Bane of Agony on the target, want %.4f", got, want)
	}
}

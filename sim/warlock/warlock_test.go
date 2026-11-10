package warlock

import (
	"math"
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

// While Wrack is channelled on a target, the warlock's Corruption there ticks at its period divided by one plus
// the percent of Wrack's dummy effect, keeps its expiry, and ticks at its period again once the channel ends.
// The warlock's Corruption on another target and a second warlock's Corruption on the channelled target keep
// their period. Wrack is the 17th affliction talent.
func TestWrackSpeedsUpTheWarlocksDotsOnItsTarget(t *testing.T) {
	encounter := &proto.Encounter{Duration: core.LongDuration, Targets: []*proto.Target{core.FreshDefaultTargetConfig(), core.FreshDefaultTargetConfig()}}
	raid := &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{
		warlockPlayer("Warlock", "00000000000000001", &proto.WarlockOptions{}),
		warlockPlayer("Second Warlock", "", &proto.WarlockOptions{}),
	}, Buffs: &proto.PartyBuffs{}}}, Buffs: &proto.RaidBuffs{}, Debuffs: &proto.Debuffs{}}
	sim := newWarlockRaidSim(raid, encounter)
	warlock, second := sim.Raid.Parties[0].Players[0].(*Warlock), sim.Raid.Parties[0].Players[1].(*Warlock)
	target, otherTarget := sim.Encounter.AllTargetUnits[0], sim.Encounter.AllTargetUnits[1]

	corruptions := []struct {
		name    string
		dot     *core.Dot
		expires time.Duration
		ticks   []time.Duration
	}{
		{name: "the warlock's Corruption on the channelled target", dot: warlock.Corruption.Dot(target)},
		{name: "the warlock's Corruption on the other target", dot: warlock.Corruption.Dot(otherTarget)},
		{name: "the second warlock's Corruption on the channelled target", dot: second.Corruption.Dot(target)},
	}
	dots := make([]*core.Dot, len(corruptions))
	for i := range corruptions {
		corruptions[i].dot.Apply(sim)
		corruptions[i].expires = corruptions[i].dot.ExpiresAt()
		dots[i] = corruptions[i].dot
	}
	sped := &corruptions[0]
	start, period := sim.CurrentTime, sped.dot.TickPeriod()
	speedup, fast := wrackSpeedup(period)

	// The channel starts a sixth of a period after the first tick, so neither of its ends meets a tick.
	wrackStart := start + period + period/6
	var wrackEnd time.Duration
	sim.AddPendingAction(&core.PendingAction{NextActionAt: wrackStart, OnAction: func(sim *core.Simulation) {
		wrack := warlock.Wrack.Dot(target)
		wrack.Apply(sim)
		wrackEnd = wrack.ExpiresAt()
	}})
	for i, dotTicks := range recordDotTicks(sim, dots...) {
		for _, tick := range dotTicks {
			corruptions[i].ticks = append(corruptions[i].ticks, tick.at)
		}
	}
	if wrackEnd <= wrackStart {
		t.Fatalf("Wrack's channel did not run")
	}

	during, after := 0, 0
	for i := 1; i < len(sped.ticks); i++ {
		from, to := sped.ticks[i-1], sped.ticks[i]
		switch {
		case from > wrackStart && to <= wrackEnd:
			during++
			if to-from != fast {
				t.Errorf("%s ticked %v after the last tick at %v while Wrack was channelled, want %v", sped.name, to-from, from, fast)
			}
		case from > wrackEnd:
			after++
			if to-from != period {
				t.Errorf("%s ticked %v after the last tick at %v once Wrack ended, want %v", sped.name, to-from, from, period)
			}
		}
	}
	if during < 2 || after < 1 {
		t.Fatalf("%s has %d gaps between ticks within the channel and %d after it, too few to tell the periods", sped.name, during, after)
	}

	// The channel runs the dot's clock faster, so it adds the ticks its share of a period fits.
	added := int(math.Floor(float64(wrackEnd-wrackStart)*(speedup-1)/float64(period) + 1e-9))
	for i := range corruptions {
		c := &corruptions[i]
		want := int(c.dot.BaseTickCount)
		if c == sped {
			want += added
		} else {
			for j := 1; j < len(c.ticks); j++ {
				if gap := c.ticks[j] - c.ticks[j-1]; gap != period {
					t.Errorf("%s ticked %v after the last tick at %v, want %v", c.name, gap, c.ticks[j-1], period)
				}
			}
		}
		if len(c.ticks) != want {
			t.Errorf("%s ticked %d times, want %d", c.name, len(c.ticks), want)
		}
		if last := c.ticks[len(c.ticks)-1]; last != c.expires {
			t.Errorf("%s last ticked at %v, want its expiry %v", c.name, last, c.expires)
		}
	}
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

// Wrack's speedup, and the period it gives a dot that otherwise ticks every period.
func wrackSpeedup(period time.Duration) (float64, time.Duration) {
	speedup := 1 + spellData.Wrack.ByID(1316697).EffectN(2).Percent()
	return speedup, time.Duration(float64(period) / speedup).Round(time.Millisecond)
}

// Wrack ended part-way through its channel keeps the share of progress the warlock's Corruption made towards
// its next tick at both ends: the time to that tick shrinks by the speedup as the channel starts and grows back
// as it ends. The Corruption then expires part-way to a tick and deals that tick at its expiry, the share of a
// full tick it ran. The warlock's Corruption on another target, never re-timed, ends on a whole tick.
func TestWrackEndedEarlyLeavesCorruptionATruncatedLastTick(t *testing.T) {
	sim, warlock, first, second := newTwoTargetWarlockSim("00000000000000001", &proto.WarlockOptions{})
	target, otherTarget := &first.Unit, &second.Unit
	// A crit would hide a tick's share of a full one.
	warlock.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: -100})

	sped, other := warlock.Corruption.Dot(target), warlock.Corruption.Dot(otherTarget)
	sped.Apply(sim)
	other.Apply(sim)
	start, period := sim.CurrentTime, sped.TickPeriod()
	_, fast := wrackSpeedup(period)

	// The channel starts a third of the way to the second tick and ends a sixth of a period later.
	wrackStart, wrackEnd := start+period+period/3, start+period+period/2
	wantAtStart := (period - period/3) * fast / period
	wantAtEnd := (wantAtStart - period/6) * period / fast
	var atStart, atEnd time.Duration
	sim.AddPendingAction(&core.PendingAction{NextActionAt: wrackStart, OnAction: func(sim *core.Simulation) {
		warlock.Wrack.Dot(target).Apply(sim)
		atStart = sped.TimeUntilNextTick(sim)
	}})
	sim.AddPendingAction(&core.PendingAction{NextActionAt: wrackEnd, OnAction: func(sim *core.Simulation) {
		warlock.Wrack.Dot(target).Deactivate(sim)
		atEnd = sped.TimeUntilNextTick(sim)
	}})

	expires := sped.ExpiresAt()
	ticks := recordDotTicks(sim, sped, other)
	if atStart != wantAtStart || atEnd != wantAtEnd {
		t.Errorf("next tick %v away as Wrack started and %v as it ended, want %v and %v", atStart, atEnd, wantAtStart, wantAtEnd)
	}

	full := ticks[0][0].damage
	for i, dotTicks := range ticks {
		name := []string{"the Corruption Wrack sped up", "the Corruption on the other target"}[i]
		want := int(sped.BaseTickCount)
		if i == 0 {
			want++
		}
		if len(dotTicks) != want {
			t.Fatalf("%s ticked %d times, want %d", name, len(dotTicks), want)
		}
		last := dotTicks[len(dotTicks)-1]
		if last.at != expires {
			t.Errorf("%s last ticked at %v, want its expiry %v", name, last.at, expires)
		}
		for j, tick := range dotTicks {
			share := 1.0
			if i == 0 && j == len(dotTicks)-1 {
				share = float64(expires-dotTicks[j-1].at) / float64(period)
			}
			if math.Abs(tick.damage-share*full) > 1e-6 {
				t.Errorf("%s ticked for %.3f at %v, want %.3f, %.4f of a full tick", name, tick.damage, tick.at, share*full, share)
			}
		}
	}
	if share := float64(expires-ticks[0][len(ticks[0])-2].at) / float64(period); share <= 0 || share >= 1 {
		t.Errorf("the sped-up Corruption's last tick is %.4f of a full tick, want part of one", share)
	}
}

// Bane of Agony's ramp advances by tick number, so under Wrack its faster ticks reach the stronger thirds
// sooner: each tick deals once, twice or three times the first tick's damage by the third its number falls in
// among the base duration's ticks, and the ticks Wrack adds deal in the last third.
func TestWrackAgonyRampFollowsTheTickNumber(t *testing.T) {
	sim, warlock := newWarlockSim("00000000000000001", &proto.WarlockOptions{}, core.MakeSingleTargetEncounter(0))
	target := sim.Encounter.AllTargetUnits[0]
	warlock.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: -100})

	agony := warlock.BaneOfAgony.Dot(target)
	agony.Apply(sim)
	wrack := warlock.Wrack.Dot(target)
	wrack.Apply(sim)
	expires, wrackEnd := agony.ExpiresAt(), wrack.ExpiresAt()

	ticks := recordDotTicks(sim, agony)[0]
	third := int(agony.BaseTickCount / 3)
	if len(ticks) <= 3*third {
		t.Fatalf("Bane of Agony ticked %d times under Wrack, want more than its %d", len(ticks), 3*third)
	}
	if ticks[third].at > wrackEnd {
		t.Errorf("Bane of Agony's first tick of the second third came at %v, after Wrack ended at %v", ticks[third].at, wrackEnd)
	}
	if last := ticks[len(ticks)-1].at; last != expires {
		t.Errorf("Bane of Agony last ticked at %v, want its expiry %v", last, expires)
	}
	for i, tick := range ticks {
		want := float64(min(i/third, 2) + 1)
		if got := tick.damage / ticks[0].damage; math.Abs(got-want) > 1e-9 {
			t.Errorf("Bane of Agony's tick %d at %v is %.4f times the first, want %.0f", i+1, tick.at, got, want)
		}
	}
}

// Wrack's contribution to a tick of the warlock's Corruption is the share of the tick's progress the channel added,
// times the damage the tick dealt, recorded on Corruption's metrics under the key wrack. Corruption ticks every 3 s
// for 18 s and Wrack doubles its speed, so each second Wrack runs on it adds a third of a tick, and the contributions
// sum to that. The warlock's Corruption on another target records none.
func TestWrackContributionIsTheTickProgressItAdded(t *testing.T) {
	cases := []struct {
		name               string
		from, to           time.Duration // Wrack's channel from Corruption's application; to 0 lets it run its 6 s.
		ticks              int32         // The ticks Wrack added progress to.
		units, least, most float64       // The contribution in full ticks: all of it, and the least and most of one tick.
	}{
		// Ticks at 3, 5, 6.5, 8, 9.5, 12, 15, 18 hold 0, 1/3, 1/2, 1/2, 1/2, 1/6, 0, 0 of Wrack's.
		{name: "Wrack from 4 s to 10 s", from: 4 * time.Second, ticks: 5, units: 2, least: 1.0 / 6, most: 0.5},
		// Ticks at 3, 5, 7, 10, 13, 16 and two thirds of one at 18 hold 0, 1/3, 1/3, 0, 0, 0, 0.
		{name: "Wrack from 4 s to 6 s", from: 4 * time.Second, to: 6 * time.Second, ticks: 2, units: 2.0 / 3, least: 1.0 / 3, most: 1.0 / 3},
		// Ticks at 12, 14.5, 16, 17.5 and a third of one at 18 hold 0, 1/6, 1/2, 1/2 and half the third.
		{name: "Wrack from 14 s past Corruption's expiry", from: 14 * time.Second, ticks: 4, units: 4.0 / 3, least: 1.0 / 6, most: 0.5},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sim, warlock, first, second := newTwoTargetWarlockSim("00000000000000001", &proto.WarlockOptions{})
			target, otherTarget := &first.Unit, &second.Unit
			// A crit would hide a tick's share of a full one.
			warlock.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: -100})

			sped, other := warlock.Corruption.Dot(target), warlock.Corruption.Dot(otherTarget)
			sped.Apply(sim)
			other.Apply(sim)
			start, period := sim.CurrentTime, sped.TickPeriod()
			if _, fast := wrackSpeedup(period); period != 3*time.Second || sped.BaseTickCount != 6 || fast != period/2 {
				t.Fatalf("Corruption ticks %d times every %v and Wrack every %v, which the cases do not work for", sped.BaseTickCount, period, fast)
			}

			sim.AddPendingAction(&core.PendingAction{NextActionAt: start + c.from, OnAction: func(sim *core.Simulation) {
				warlock.Wrack.Dot(target).Apply(sim)
			}})
			if c.to > 0 {
				sim.AddPendingAction(&core.PendingAction{NextActionAt: start + c.to, OnAction: func(sim *core.Simulation) {
					warlock.Wrack.Dot(target).Deactivate(sim)
				}})
			}
			full := recordDotTicks(sim, sped, other)[0][0].damage
			sim.Cleanup()

			got := wrackMetric(warlock, warlock.Corruption, target)
			if got == nil {
				t.Fatalf("Corruption recorded no Wrack contribution, want %.4f full ticks", c.units)
			}
			n := got.AggregatorData.N
			if total := got.Avg * float64(n) / full; n != c.ticks || math.Abs(total-c.units) > 1e-6 {
				t.Errorf("Wrack contributed %.4f full ticks over %d ticks, want %.4f over %d", total, n, c.units, c.ticks)
			}
			if least, most := got.Min/full, got.Max/full; math.Abs(least-c.least) > 1e-6 || math.Abs(most-c.most) > 1e-6 {
				t.Errorf("Wrack contributed %.4f to %.4f of a full tick to one tick, want %.4f to %.4f", least, most, c.least, c.most)
			}
			if got := wrackMetric(warlock, warlock.Corruption, otherTarget); got != nil {
				t.Errorf("the Corruption on the other target recorded a Wrack contribution over %d ticks, want none", got.AggregatorData.N)
			}
		})
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

// Wrack's contribution to Bane of Agony and Siphon Life is each tick's share of progress the channel added times the
// damage the tick dealt. For Bane of Agony that holds whatever third of the ramp the tick deals in: reaching a
// stronger third sooner counts for nothing extra. What a Siphon Life tick heals counts for nothing. The shares are
// worked here from the time each tick's interval ran under the channel. Siphon Life is the 14th affliction talent,
// Wrack the 17th.
func TestWrackContributionIsTheShareOfEachTicksDamage(t *testing.T) {
	cases := []struct {
		name    string
		talents string
		spell   func(*Warlock) *core.Spell
	}{
		{name: "Bane of Agony", talents: "00000000000000001", spell: func(w *Warlock) *core.Spell { return w.BaneOfAgony }},
		{name: "Siphon Life", talents: "00000000000001001", spell: func(w *Warlock) *core.Spell { return w.SiphonLife }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sim, warlock := newWarlockSim(c.talents, &proto.WarlockOptions{}, core.MakeSingleTargetEncounter(0))
			target := sim.Encounter.AllTargetUnits[0]
			warlock.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: -100})

			spell := c.spell(warlock)
			dot := spell.Dot(target)
			dot.Apply(sim)
			start, period := sim.CurrentTime, dot.TickPeriod()
			_, fast := wrackSpeedup(period)

			// The channel starts a third of the way to the second tick.
			wrackStart := start + period + period/3
			var wrackEnd time.Duration
			sim.AddPendingAction(&core.PendingAction{NextActionAt: wrackStart, OnAction: func(sim *core.Simulation) {
				wrack := warlock.Wrack.Dot(target)
				wrack.Apply(sim)
				wrackEnd = wrack.ExpiresAt()
			}})
			ticks := recordDotTicks(sim, dot)[0]
			sim.Cleanup()

			want, shared := wantWrackContribution(ticks, start, wrackStart, wrackEnd, period, fast)
			if shared < 3 {
				t.Fatalf("Wrack added progress to %d of %s's ticks, too few to tell", shared, c.name)
			}

			got := wrackMetric(warlock, spell, target)
			if got == nil {
				t.Fatalf("%s recorded no Wrack contribution, want %.3f", c.name, want)
			}
			if total := got.Avg * float64(got.AggregatorData.N); got.AggregatorData.N != shared || math.Abs(total-want) > 1e-6 {
				t.Errorf("Wrack contributed %.3f to %s over %d ticks, want %.3f over %d", total, c.name, got.AggregatorData.N, want, shared)
			}
		})
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

// The Wrack contribution to a dot's ticks, worked from the time each tick's interval ran under the channel, and the
// number of ticks it added progress to. The dot ticks every period, and every fast one while the channel runs.
func wantWrackContribution(ticks []dotTick, start, wrackStart, wrackEnd, period, fast time.Duration) (float64, int32) {
	var want float64
	var shared int32
	last := start
	for _, tick := range ticks {
		under := max(0, min(tick.at, wrackEnd)-max(last, wrackStart))
		progress := float64(tick.at-last-under)/float64(period) + float64(under)/float64(fast)
		if added := float64(under)/float64(fast) - float64(under)/float64(period); added > 0 {
			want += added / progress * tick.damage
			shared++
		}
		last = tick.at
	}
	return want, shared
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

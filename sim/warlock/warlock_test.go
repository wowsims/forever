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
// its talents, demons, consumables and rotation, on the preraid gear set.
func TestDemonologyFireBrand(t *testing.T) {
	runBuildSuite(t, "Fire Brand", "demo_fire_brand", proto.Race_RaceOrc)
}

func TestDemonologyShadowBrand(t *testing.T) {
	runBuildSuite(t, "Shadow Brand", "demo_shadow_brand", proto.Race_RaceOrc)
}

func TestDestruction(t *testing.T) {
	runBuildSuite(t, "Destruction", "destro", proto.Race_RaceGnome)
}

func TestAffliction(t *testing.T) {
	runBuildSuite(t, "Affliction", "sm_ruin", proto.Race_RaceOrc)
}

func TestAfflictionWrack(t *testing.T) {
	runBuildSuite(t, "Wrack", "aff_wrack", proto.Race_RaceHuman)
}

func runBuildSuite(t *testing.T, label string, buildFile string, otherRace proto.Race) {
	build := core.GetTestBuildFromJSON(proto.Class_ClassWarlock, "../../ui/specs/warlock/dps/builds", buildFile, core.ItemFilter{}, nil, nil)
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:       proto.Class_ClassWarlock,
			Race:        build.Race,
			OtherRaces:  []proto.Race{otherRace},
			SpecOptions: core.SpecOptionsCombo{Label: label, SpecOptions: build.SpecOptions.SpecOptions},
			GearSet:     core.GetGearSet("../../ui/specs/warlock/dps/gear_sets", "preraid"),
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
	for i := 0; i < 400; i++ {
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
	for i := 0; i < 100; i++ {
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

// Bane of Agony's ticks are set at cast: the base ramps in thirds, the spell power share is the spell power at
// cast on every tick alike, and Amplify Curse (talent 9) raises both. Each tick leaves its base and share in
// the dot's snapshot base, which the last tick clears as the dot ends, so the test reads the first 11.
func TestBaneOfAgonyTicksSetAtCast(t *testing.T) {
	encounter := core.MakeSingleTargetEncounter(0)
	encounter.Duration = 600
	sim, warlock := newWarlockSim("000000001", &proto.WarlockOptions{}, encounter)
	target := sim.Encounter.AllTargetUnits[0]
	agony := warlock.BaneOfAgony
	dot := agony.Dot(target)
	warlock.AddStatsDynamic(sim, stats.Stats{stats.SpellDamage: 300, stats.SpellHitPercent: 100})

	// The snapshot base of the first 11 ticks of one cast, with spellPower more spell power from the fifth on.
	castTicks := func(amplified bool, spellPower float64) []float64 {
		if amplified {
			warlock.AmplifyCurseAura.Activate(sim)
		}
		agony.SkipCastAndApplyEffects(sim, target)
		if !dot.IsActive() {
			t.Fatalf("Bane of Agony did not land")
		}
		var ticks []float64
		for len(ticks) < 11 {
			before := dot.TickCount()
			for dot.TickCount() == before && !sim.Step() {
			}
			ticks = append(ticks, dot.SnapshotBaseDamage)
			if len(ticks) == 4 {
				warlock.AddStatsDynamic(sim, stats.Stats{stats.SpellDamage: spellPower})
			}
		}
		return ticks
	}

	plain := castTicks(false, 500)
	for i, want := range []float64{plain[0], plain[4], plain[8]} {
		for _, got := range plain[4*i : min(4*i+4, len(plain))] {
			if math.Abs(got-want) > 1e-6 {
				t.Fatalf("third %d ticks %v, want each the same with 500 spell power gained mid-dot", i+1, plain)
			}
		}
	}
	base := 2 * (plain[4] - plain[0])
	if step := plain[8] - plain[4]; base <= 0 || math.Abs(step-base/2) > 1e-6 {
		t.Fatalf("thirds %.3f, %.3f, %.3f do not ramp at half, one and one and a half of a base", plain[0], plain[4], plain[8])
	}
	share := plain[0] - base/2

	amplified := castTicks(true, 0)
	coefficient := dot.BonusCoefficient
	ampShare := share + 500*coefficient
	for i, ramp := range []float64{0.5, 1, 1.5} {
		want := (base*ramp + ampShare) * warlock.AmplifyCurseMultiplier
		if got := amplified[4*i]; math.Abs(got-want) > 1e-6 {
			t.Errorf("amplified third %d tick %.3f, want %.3f: the base %.3f plus the share at cast %.3f, amplified",
				i+1, got, want, base*ramp, ampShare)
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

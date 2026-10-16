package warlock

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

func init() {
	RegisterWarlock()
	common.RegisterAllEffects()
}

func TestAffliction(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassWarlock,
			Race:       proto.Race_RaceOrc,
			OtherRaces: []proto.Race{proto.Race_RaceHuman},
			SpecOptions: core.SpecOptionsCombo{Label: "Affliction", SpecOptions: &proto.Player_Warlock{
				Warlock: &proto.Warlock{
					Options: &proto.Warlock_Options{
						ClassOptions: &proto.WarlockOptions{
							Summon:       proto.WarlockOptions_Imp,
							Armor:        proto.WarlockOptions_DemonArmor,
							CurseOptions: proto.WarlockOptions_Elements,
							SoulShards:   32,
						},
					},
				},
			}},
			GearSet:  core.GetGearSet("../../ui/specs/warlock/dps/gear_sets", "preraid"),
			Talents:  "25550320135201351--053",
			Rotation: core.GetAplRotation("../../ui/specs/warlock/dps/apls", "affliction"),
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

func TestDestruction(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassWarlock,
			Race:       proto.Race_RaceOrc,
			OtherRaces: []proto.Race{proto.Race_RaceGnome},
			SpecOptions: core.SpecOptionsCombo{Label: "Destruction", SpecOptions: &proto.Player_Warlock{
				Warlock: &proto.Warlock{
					Options: &proto.Warlock_Options{
						ClassOptions: &proto.WarlockOptions{
							SacrificedDemon: proto.WarlockOptions_Imp,
							Armor:           proto.WarlockOptions_DemonArmor,
							CurseOptions:    proto.WarlockOptions_Elements,
							SoulShards:      32,
						},
					},
				},
			}},
			GearSet:  core.GetGearSet("../../ui/specs/warlock/dps/gear_sets", "preraid"),
			Talents:  "03-0050203001-0550355103100351",
			Rotation: core.GetAplRotation("../../ui/specs/warlock/dps/apls", "destruction"),
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
	player := &proto.Player{
		Name:          "Warlock",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarlock,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: talents,
		Spec:          &proto.Player_Warlock{Warlock: &proto.Warlock{Options: &proto.Warlock_Options{ClassOptions: options}}},
		Rotation:      &proto.APLRotation{},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*Warlock)
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

// In execute, Decimation's buff (440873) comes with the cast of a Shadow Bolt or Searing Pain whose hit lands,
// before a Shadow Bolt reaches the target; a miss or a full resist grants nothing, and the bolt's impact
// grants nothing either.
func TestDecimationGrantedOnLandedCast(t *testing.T) {
	newSim := func(resistance float64) (*core.Simulation, *Warlock, *core.Unit) {
		encounter := core.MakeSingleTargetEncounter(0)
		encounter.Duration = 3600
		encounter.ExecuteProportion_90, encounter.ExecuteProportion_45, encounter.ExecuteProportion_35 = 1, 1, 1
		encounter.Targets[0].Stats[stats.FireResistance] = resistance
		encounter.Targets[0].Stats[stats.ShadowResistance] = resistance
		sim, warlock := newWarlockSim("-000000000002", &proto.WarlockOptions{}, encounter)
		for !sim.IsExecutePhase35() && !sim.Step() {
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

// Bane of Agony's ticks are set at cast: Amplify Curse (talent 9) raises the base that ramps in thirds, and the
// spell power share is the spell power at cast, on every tick alike. Each tick leaves its base and share in
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
		want := base*ramp*warlock.AmplifyCurseMultiplier + ampShare
		if got := amplified[4*i]; math.Abs(got-want) > 1e-6 {
			t.Errorf("amplified third %d tick %.3f, want %.3f: the base %.3f amplified, plus the share at cast %.3f",
				i+1, got, want, base*ramp, ampShare)
		}
	}
}

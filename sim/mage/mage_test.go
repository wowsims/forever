package mage

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common" // imported to get item effects included.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func init() {
	RegisterMage()
	common.RegisterAllEffects()
}

func TestArcane(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{arcaneSuite()}))
}

func TestFFB(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{ffbSuite()}))
}

func TestFire(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{fireSuite()}))
}

func TestFrost(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{frostSuite()}))
}

func TestFingersOfFrostGivesBlizzardTicksNoShatterBonus(t *testing.T) {
	const noShatter, shatter = "--000000001000000200", "--000000001000300200"
	ticksWithout, directWithout, _ := shatterCrits(noShatter)
	ticksWith, directWith, uptimeWith := shatterCrits(shatter)

	if uptimeWith == 0 {
		t.Fatal("Fingers of Frost was never up")
	}
	if ticksWithout == 0 {
		t.Fatal("no Blizzard tick crit")
	}
	if directWith <= directWithout {
		t.Fatalf("Fire Blast gained no crits with Shatter: %d without, %d with", directWithout, directWith)
	}
	if ticksWith != ticksWithout {
		t.Errorf("Blizzard tick crits moved with Shatter: %d without, %d with", ticksWithout, ticksWith)
	}
}

// Each book rank registers beside the trainer's top rank, Arcane Missiles with its own missile. Both ranks of Frost
// Ward run Fire Ward's one category cooldown and keep a shield each, and the higher rank's Fireball burn holds a target.
func TestBookRanksBesideTrainerRanks(t *testing.T) {
	sim, mage := newMageSim("")
	target := mage.CurrentTarget

	spellOf := func(id int32, tag int32) *core.Spell {
		spell := mage.GetSpell(core.ActionID{SpellID: id, Tag: tag})
		if spell == nil {
			t.Fatalf("spell %d tag %d is not registered", id, tag)
		}
		return spell
	}
	for _, rank := range []*spelldata.Spell{frostboltTrainerRank, frostboltBookRank, arcaneMissilesTrainerRank, arcaneMissilesBookRank} {
		spellOf(rank.ID, 0)
	}
	spellOf(arcaneMissilesTrainerRank.PeriodicEffect().TriggerID, 0)
	spellOf(arcaneMissilesBookRank.PeriodicEffect().TriggerID, 0)

	trainerWard, bookWard := spellOf(frostWardTrainerRank.ID, 0), spellOf(frostWardBookRank.ID, 0)
	if !bookWard.Cast(sim, target) {
		t.Fatal("the book rank of Frost Ward did not cast")
	}
	if trainerWard.CD.IsReady(sim) || spellOf(spellData.FireWard.Highest().ID, 0).CD.IsReady(sim) {
		t.Error("the book rank of Frost Ward left the trainer's rank or Fire Ward ready, want one category cooldown")
	}
	if !bookWard.RelatedSelfBuff.IsActive() || trainerWard.RelatedSelfBuff.IsActive() {
		t.Error("the book rank of Frost Ward shares its shield with the trainer's rank, want a shield each")
	}

	spellOf(fireballTrainerRank.ID, 0)
	spellOf(fireballBookRank.ID, 0)
	trainerBurn := spellOf(fireballTrainerRank.ID, 1).Dot(target)
	bookBurn := spellOf(fireballBookRank.ID, 1).Dot(target)
	trainerBurn.Apply(sim)
	bookBurn.Apply(sim)
	if trainerBurn.IsActive() || !bookBurn.IsActive() {
		t.Error("the book rank's Fireball burn left the trainer's up, want it to replace the trainer's")
	}
	trainerBurn.Apply(sim)
	if trainerBurn.IsActive() || !bookBurn.IsActive() {
		t.Error("the trainer's Fireball burn landed under the book rank, want it refused")
	}
}

// An Arcane Missiles that Missile Barrage pays for keeps Clearcasting up; one it does not pay for spends it.
func TestMissileBarrageArcaneMissilesKeepsClearcasting(t *testing.T) {
	for _, withBarrage := range []bool{true, false} {
		sim, mage := newMageSim(ArcaneTalents)
		mage.ClearCasting.Activate(sim)
		if withBarrage {
			mage.MissileBarrageAura.Activate(sim)
		}
		if !mage.GetSpell(core.ActionID{SpellID: arcaneMissilesBookRank.ID}).Cast(sim, mage.CurrentTarget) {
			t.Fatalf("Arcane Missiles did not cast, Missile Barrage up = %v", withBarrage)
		}
		if mage.MissileBarrageAura.IsActive() {
			t.Error("Arcane Missiles left Missile Barrage up")
		}
		if mage.ClearCasting.IsActive() != withBarrage {
			t.Errorf("Clearcasting up after Arcane Missiles = %v with Missile Barrage up = %v, want them equal",
				mage.ClearCasting.IsActive(), withBarrage)
		}
	}
}

// A rotation of only Arcane Missiles procs Clearcasting at about the talent's chance per channel: the
// channel start and its five missiles share that chance between them.
func TestArcaneMissilesSplitsClearcastingChance(t *testing.T) {
	_, mage := newMageSim(ArcaneTalents)
	chance := arcaneConcentrationChance(mage)
	metrics := runMageRotation(arcaneSuite(), ArcaneTalents, 300, arcaneMissilesBookRank.ID)
	var channels, procs float64
	for _, action := range metrics.Actions {
		if action.Id.GetSpellId() == arcaneMissilesBookRank.ID {
			for _, target := range action.Targets {
				channels += float64(target.Casts) / float64(core.DefaultSimTestOptions.Iterations)
			}
		}
	}
	for _, aura := range metrics.Auras {
		if aura.Id.GetSpellId() == spellData.Clearcasting.Highest().ID {
			procs = aura.ProcsAvg
		}
	}
	if channels == 0 || procs < chance*channels/2 || procs > chance*channels*3/2 {
		t.Fatalf("Clearcasting procced %.2f times over %.2f channels, want about the talent's %.3f a channel", procs, channels, chance)
	}
}

// The Arcane Missiles channel rolls Arcane Concentration once as it starts, before its first missile flies,
// at a sixth of the talent's chance.
func TestArcaneMissilesChannelStartProcsClearcasting(t *testing.T) {
	sim, mage := newMageSimAgainst(ArcaneTalents, longEncounter(1))
	target := mage.CurrentTarget
	channel := mage.GetSpell(core.ActionID{SpellID: arcaneMissilesBookRank.ID})
	chance := arcaneConcentrationChance(mage)
	manaMetrics := mage.NewManaMetrics(core.ActionID{SpellID: arcaneMissilesBookRank.ID})
	gains := 0
	mage.ClearCasting.ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) { gains++ })

	const channels = 2300
	for range channels {
		mage.AddMana(sim, mage.MaxMana(), manaMetrics)
		if !channel.Cast(sim, target) {
			t.Fatalf("Arcane Missiles did not cast at %v", sim.CurrentTime)
		}
		stepFor(sim, 100*time.Millisecond)
		dot := channel.Dot(target)
		dot.Deactivate(sim)
		if dot.TickCount() != 0 {
			t.Fatal("a missile flew before the channel was ended")
		}
		mage.ClearCasting.Deactivate(sim)
		stepFor(sim, 1400*time.Millisecond)
	}

	if rate, want := float64(gains)/channels, chance/6; rate < want/2 || rate > want*3/2 {
		t.Errorf("Clearcasting procced on %d of %d channel starts (%.4f), want about a sixth of the talent's chance, %.4f", gains, channels, rate, want)
	}
}

// One Flamestrike ground tick rolls Arcane Concentration once, however many enemies stand in it.
func TestFlamestrikeGroundTickProcsClearcastingOnce(t *testing.T) {
	sim, mage := newMageSimAgainst(ArcaneTalents, longEncounter(3))
	chance := arcaneConcentrationChance(mage)
	burn := flamestrikeBurn(t, mage)
	gains := 0
	mage.ClearCasting.ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) { gains++ })

	const ticks = 2000
	for range ticks {
		burn.AOEDot().TickOnce(sim)
		stepFor(sim, 1100*time.Millisecond)
		mage.ClearCasting.Deactivate(sim)
	}

	if rate := float64(gains) / ticks; rate < chance/2 || rate > chance*3/2 {
		t.Errorf("Clearcasting procced on %d of %d ticks on 3 targets (%.3f), want about the talent's %.3f once per tick", gains, ticks, rate, chance)
	}
}

// A Flamestrike ground tick leaves Clearcasting up for the next cast.
func TestFlamestrikeGroundTickKeepsClearcasting(t *testing.T) {
	sim, mage := newMageSimAgainst(ArcaneTalents, longEncounter(3))
	burn := flamestrikeBurn(t, mage)
	mage.ClearCasting.Activate(sim)
	if !burn.Cast(sim, mage.CurrentTarget) {
		t.Fatal("the Flamestrike burn did not cast")
	}
	for tick := range int32(4) {
		stepFor(sim, 2100*time.Millisecond)
		if ticks := burn.AOEDot().TickCount(); ticks <= tick {
			t.Fatalf("the burn ticked %d times by %v", ticks, sim.CurrentTime)
		}
		if !mage.ClearCasting.IsActive() {
			t.Fatalf("Clearcasting fell off at %v while only Flamestrike's ground ticks dealt damage", sim.CurrentTime)
		}
	}
}

func shatterCrits(talents string) (tickCrits, directCrits int32, fingersOfFrostUptime float64) {
	tickID := spellData.BlizzardTriggered.Highest().ID
	directID := spellData.FireBlast.Highest().ID
	metrics := runMageRotation(frostSuite(), talents, 40, directID, spellData.Blizzard.Highest().ID)
	for _, action := range metrics.Actions {
		for _, target := range action.Targets {
			switch action.Id.GetSpellId() {
			case tickID:
				tickCrits += target.Crits
			case directID:
				directCrits += target.Crits
			}
		}
	}
	for _, aura := range metrics.Auras {
		if aura.Id.GetSpellId() == spellData.FingersOfFrostTriggered.Highest().ID {
			fingersOfFrostUptime = aura.UptimeSecondsAvg
		}
	}
	return tickCrits, directCrits, fingersOfFrostUptime
}

// A mage with no gear and no rotation, ready to cast by hand.
func newMageSim(talents string) (*core.Simulation, *Mage) {
	return newMageSimAgainst(talents, core.MakeSingleTargetEncounter(0))
}

func newMageSimAgainst(talents string, encounter *proto.Encounter) (*core.Simulation, *Mage) {
	suite := arcaneSuite()
	player := core.WithSpec(&proto.Player{
		Class:         suite.Class,
		Race:          suite.Race,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: talents,
		Rotation:      &proto.APLRotation{},
	}, suite.SpecOptions.SpecOptions)
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*Mage)
}

// An hour against the given number of default targets, long enough for a test to cast by hand a thousand times.
func longEncounter(targets int) *proto.Encounter {
	encounter := core.MakeSingleTargetEncounter(0)
	encounter.Duration = 3600
	for range targets - 1 {
		encounter.Targets = append(encounter.Targets, core.NewDefaultTarget())
	}
	return encounter
}

// The burn Flamestrike's cast leaves on the ground, registered under the cast's id with tag 1.
func flamestrikeBurn(t *testing.T, mage *Mage) *core.Spell {
	burn := mage.GetSpell(core.ActionID{SpellID: spellData.Flamestrike.Highest().ID, Tag: 1})
	if burn == nil {
		t.Fatal("the mage has no Flamestrike burn spell")
	}
	return burn
}

func arcaneConcentrationChance(mage *Mage) float64 {
	return spellData.ArcaneConcentration.Rank(mage.Talents.ArcaneConcentration).StatedChance()
}

func stepFor(sim *core.Simulation, d time.Duration) {
	for end := sim.CurrentTime + d; sim.CurrentTime < end && !sim.Step(); {
	}
}

// The suite's mage in its gear, casting the spells in priority order against one target.
func runMageRotation(suite core.CharacterSuiteConfig, talents string, duration float64, spellIDs ...int32) *proto.UnitMetrics {
	actions := make([]string, len(spellIDs))
	for i, id := range spellIDs {
		actions[i] = fmt.Sprintf(`{"action":{"castSpell":{"spellId":{"spellId":%d}}}}`, id)
	}
	rotation := fmt.Sprintf(`{"type":"TypeAPL","priorityList":[%s]}`, strings.Join(actions, ","))
	player := core.WithSpec(&proto.Player{
		Class:         suite.Class,
		Race:          suite.Race,
		Equipment:     suite.GearSet.GearSet,
		TalentsString: talents,
		Rotation:      core.APLRotationFromJsonString(rotation),
	}, suite.SpecOptions.SpecOptions)
	result := core.RunRaidSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: duration, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: core.DefaultSimTestOptions,
	})
	return result.RaidMetrics.Parties[0].Players[0]
}

func BenchmarkMage(b *testing.B) {
	b.Run("Arcane", func(b *testing.B) { core.CharacterBenchmark(b, arcaneSuite()) })
	b.Run("FFB", func(b *testing.B) { core.CharacterBenchmark(b, ffbSuite()) })
	b.Run("Fire", func(b *testing.B) { core.CharacterBenchmark(b, fireSuite()) })
	b.Run("Frost", func(b *testing.B) { core.CharacterBenchmark(b, frostSuite()) })
}

func arcaneSuite() core.CharacterSuiteConfig {
	config := mageSuite(ArcaneTalents, "arcane")
	config.OtherRotations = []core.RotationCombo{
		core.RotationWithSpellIds("arcane_book_ranks", config.Rotation, map[int32]int32{10212: 25345}),
	}
	return config
}

func ffbSuite() core.CharacterSuiteConfig {
	return mageSuite(FFBTalents, "ffb")
}

func fireSuite() core.CharacterSuiteConfig {
	config := mageSuite(FireTalents, "deepFire")
	config.OtherRotations = []core.RotationCombo{
		core.RotationWithSpellIds("deepFire_book_ranks", config.Rotation, map[int32]int32{10151: 25306, 10212: 25345}),
	}
	return config
}

func frostSuite() core.CharacterSuiteConfig {
	config := mageSuite(FrostTalents, "frost")
	config.StartingDistance = 30
	config.OtherRotations = []core.RotationCombo{
		core.RotationWithSpellIds("frost_book_ranks", config.Rotation, map[int32]int32{10181: 25304}),
	}
	return config
}

func mageSuite(talents string, apl string) core.CharacterSuiteConfig {
	return core.CharacterSuiteConfig{
		Class:      proto.Class_ClassMage,
		Race:       proto.Race_RaceHighOrderSkyborne,
		OtherRaces: []proto.Race{proto.Race_RaceGnome},
		SpecOptions: core.SpecOptionsCombo{Label: "Default", SpecOptions: &proto.Player_Mage{
			Mage: &proto.Mage{
				Options: &proto.Mage_Options{
					ClassOptions: &proto.MageOptions{
						DefaultMageArmor: proto.MageArmor_MageArmorMageArmor,
					},
				},
			},
		}},
		Consumables: DefaultConsumables,
		Profession1: proto.Profession_Engineering,
		Profession2: proto.Profession_Tailoring,
		GearSet:     core.GetGearSet("../../ui/specs/mage/dps/gear_sets", "p1Arcane"),
		Talents:     talents,
		Rotation:    core.GetAplRotation("../../ui/specs/mage/dps/apls", apl),
		ItemFilter: core.ItemFilter{
			WeaponTypes: []proto.WeaponType{
				proto.WeaponType_WeaponTypeDagger,
				proto.WeaponType_WeaponTypeSword,
				proto.WeaponType_WeaponTypeOffHand,
				proto.WeaponType_WeaponTypeStaff,
			},
			ArmorType: proto.ArmorType_ArmorTypeCloth,
			RangedWeaponTypes: []proto.RangedWeaponType{
				proto.RangedWeaponType_RangedWeaponTypeWand,
			},
			EnchantBlacklist: []int32{2673, 3225, 3273},
		},
	}
}

var ArcaneTalents = "050205223101311531-5500003-003"
var FFBTalents = "-5550000123011303-2035010311013002"
var FireTalents = "0502052001001-55500001230133031-003"
var FrostTalents = "203205-5-053500011121331251"

var DefaultConsumables = &proto.ConsumesSpec{
	FlaskId: 13512, // Flask of Supreme Power
	FoodId:  13931, // Nightfin Soup
	ElixirIds: []int32{
		13454,  // Greater Arcane Elixir
		13447,  // Elixir of the Sages
		250348, // Elixir of the Whale
		250341, // Greater Mageblood Elixir
	},
	MhImbueId:       20749,  // Brilliant Wizard Oil
	MhWeaponImbueId: 277502, // Scroll of Imbue Precision
	ResourcePotId:   13444,  // Major Mana Potion
}

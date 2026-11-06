package mage

import (
	"fmt"
	"strings"
	"testing"

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

// A rotation of only Arcane Missiles still procs Clearcasting: the channel never rolls, its missiles do.
func TestArcaneMissilesMissilesProcClearcasting(t *testing.T) {
	metrics := runMageRotation(arcaneSuite(), ArcaneTalents, 60, arcaneMissilesBookRank.ID)
	for _, aura := range metrics.Auras {
		if aura.Id.GetSpellId() == spellData.Clearcasting.Highest().ID && aura.ProcsAvg > 0 {
			return
		}
	}
	t.Fatal("Clearcasting never procced from Arcane Missiles' missiles")
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
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*Mage)
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

package core_test

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func statsFromHundredAgility(t *testing.T, class proto.Class, race proto.Race) stats.Stats {
	t.Helper()
	character, _ := nakedSheet(t, class, race)
	base := character.GetStats()
	more := base
	more[stats.Agility] += 100
	return character.ApplyStatDependencies(more).Subtract(character.ApplyStatDependencies(base))
}

func TestAgilityPerClass(t *testing.T) {
	for _, tc := range []struct {
		class  proto.Class
		race   proto.Race
		gained map[stats.Stat]float64
	}{
		{proto.Class_ClassWarrior, proto.Race_RaceHuman, map[stats.Stat]float64{stats.DodgePercent: 5, stats.PhysicalCritPercent: 5}},
		{proto.Class_ClassPaladin, proto.Race_RaceHuman, map[stats.Stat]float64{stats.DodgePercent: 5.06}},
		{proto.Class_ClassHunter, proto.Race_RaceNightElf, map[stats.Stat]float64{stats.DodgePercent: 3.78, stats.RangedAttackPower: 200}},
		{proto.Class_ClassRogue, proto.Race_RaceHuman, map[stats.Stat]float64{stats.DodgePercent: 6.9, stats.AttackPower: 100}},
		{proto.Class_ClassPriest, proto.Race_RaceHuman, map[stats.Stat]float64{stats.DodgePercent: 5}},
		{proto.Class_ClassShaman, proto.Race_RaceOrc, map[stats.Stat]float64{stats.DodgePercent: 5.08}},
		{proto.Class_ClassMage, proto.Race_RaceHuman, map[stats.Stat]float64{stats.DodgePercent: 5.14}},
		{proto.Class_ClassWarlock, proto.Race_RaceHuman, map[stats.Stat]float64{stats.DodgePercent: 5}},
		{proto.Class_ClassDruid, proto.Race_RaceTauren, map[stats.Stat]float64{stats.DodgePercent: 5}},
	} {
		t.Run(tc.class.String(), func(t *testing.T) {
			gained := statsFromHundredAgility(t, tc.class, tc.race)
			for stat, want := range tc.gained {
				if got := gained[stat]; math.Abs(got-want) > 1e-9 {
					t.Errorf("100 agility gives %v %s, want %v", got, stat.StatName(), want)
				}
			}
		})
	}
}

func TestDefenseRaisesDodgeParryAndBlock(t *testing.T) {
	character, _ := nakedSheet(t, proto.Class_ClassWarrior, proto.Race_RaceHuman)
	base := character.GetStats()
	more := base
	more[stats.DefenseRating] += 25 * core.DefenseRatingPerDefenseLevel
	gained := character.ApplyStatDependencies(more).Subtract(character.ApplyStatDependencies(base))
	for _, stat := range []stats.Stat{stats.DodgePercent, stats.ParryPercent, stats.BlockPercent, stats.ReducedCritTakenPercent} {
		if got := gained[stat]; math.Abs(got-1) > 1e-9 {
			t.Errorf("25 defense gives %v%% %s, want 1%%", got, stat.StatName())
		}
	}
}

func TestClassBaseAvoidance(t *testing.T) {
	for _, tc := range []struct {
		class        proto.Class
		race         proto.Race
		dodge        float64
		agilityDodge float64
	}{
		{proto.Class_ClassWarrior, proto.Race_RaceHuman, 0, 4},
		{proto.Class_ClassPaladin, proto.Race_RaceHuman, 0.007, 3.289},
		{proto.Class_ClassHunter, proto.Race_RaceNightElf, 0, 4.914},
		{proto.Class_ClassRogue, proto.Race_RaceHuman, 0, 8.97},
		{proto.Class_ClassPriest, proto.Race_RaceHuman, 0.03, 2},
		{proto.Class_ClassShaman, proto.Race_RaceOrc, 0.017, 2.6416},
		{proto.Class_ClassMage, proto.Race_RaceHuman, 0.032, 1.799},
		{proto.Class_ClassWarlock, proto.Race_RaceHuman, 0.02, 2.5},
		{proto.Class_ClassDruid, proto.Race_RaceTauren, 0.009, 2.75},
	} {
		t.Run(tc.class.String(), func(t *testing.T) {
			character, sheet := nakedSheet(t, tc.class, tc.race)
			if got := character.PseudoStats.BaseDodgeChance; math.Abs(got-tc.dodge) > 1e-12 {
				t.Errorf("base dodge %v, want %v", got, tc.dodge)
			}
			if got := sheet[stats.DodgePercent]; math.Abs(got-tc.agilityDodge) > 1e-9 {
				t.Errorf("dodge from %v agility %v%%, want %v%%", sheet[stats.Agility], got, tc.agilityDodge)
			}
			if character.PseudoStats.BaseParryChance != 0 || character.PseudoStats.BaseBlockChance != 0 {
				t.Errorf("class adds parry %v and block %v, want both from the attack table alone",
					character.PseudoStats.BaseParryChance, character.PseudoStats.BaseBlockChance)
			}
		})
	}
}

func TestSheetParryAndBlockAgainstAnAttackerOfTheSameLevel(t *testing.T) {
	paladin, _ := nakedSheet(t, proto.Class_ClassPaladin, proto.Race_RaceHuman)
	if got := paladin.GetPseudoStatsProto()[proto.PseudoStat_PseudoStatParryPercent]; math.Abs(got-5) > 1e-9 {
		t.Errorf("naked paladin parry %v%%, want 5%%", got)
	}

	warrior, _ := nakedSheet(t, proto.Class_ClassWarrior, proto.Race_RaceHuman)
	sheet := warrior.GetPseudoStatsProto()
	if got := sheet[proto.PseudoStat_PseudoStatParryPercent]; math.Abs(got-5) > 1e-9 {
		t.Errorf("naked warrior parry %v%%, want 5%%", got)
	}
	if got := sheet[proto.PseudoStat_PseudoStatBlockPercent]; got != 0 {
		t.Errorf("warrior without a shield blocks %v%%, want 0%%", got)
	}

	warrior.PseudoStats.CanBlock = true
	if got := warrior.GetPseudoStatsProto()[proto.PseudoStat_PseudoStatBlockPercent]; math.Abs(got-5) > 1e-9 {
		t.Errorf("naked warrior with a shield blocks %v%%, want 5%%", got)
	}

	mage, _ := nakedSheet(t, proto.Class_ClassMage, proto.Race_RaceHuman)
	if got := mage.GetPseudoStatsProto()[proto.PseudoStat_PseudoStatParryPercent]; got != 0 {
		t.Errorf("mage parry %v%%, want 0%%", got)
	}
}

func TestBlockValueFromStrength(t *testing.T) {
	for _, tc := range []struct {
		class       proto.Class
		race        proto.Race
		perStrength float64
	}{
		{proto.Class_ClassWarrior, proto.Race_RaceHuman, 0.05},
		{proto.Class_ClassPaladin, proto.Race_RaceHuman, 0.05},
		{proto.Class_ClassShaman, proto.Race_RaceOrc, 0.05},
		{proto.Class_ClassDruid, proto.Race_RaceTauren, 0},
	} {
		t.Run(tc.class.String(), func(t *testing.T) {
			character, sheet := nakedSheet(t, tc.class, tc.race)
			if got := character.PseudoStats.BlockValuePerStrength; got != tc.perStrength {
				t.Errorf("block value per strength %v, want %v", got, tc.perStrength)
			}
			if got := sheet[stats.BlockValue]; got != 0 {
				t.Errorf("block value stat %v, want 0: strength adds to the block reduction only", got)
			}
			if got := character.GetPseudoStatsProto()[proto.PseudoStat_PseudoStatBlockValuePerStrength]; got != tc.perStrength {
				t.Errorf("sheet block value per strength %v, want %v", got, tc.perStrength)
			}
		})
	}
}

package core_test

import (
	"maps"
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
		class        proto.Class
		race         proto.Race
		dodgePerCrit float64
		crit         bool
		gained       map[stats.Stat]float64
	}{
		{proto.Class_ClassWarrior, proto.Race_RaceHuman, 1, true, nil},
		{proto.Class_ClassPaladin, proto.Race_RaceHuman, 1, false, nil},
		{proto.Class_ClassHunter, proto.Race_RaceNightElf, 2, false, map[stats.Stat]float64{stats.RangedAttackPower: 200}},
		{proto.Class_ClassRogue, proto.Race_RaceHuman, 2, false, map[stats.Stat]float64{stats.AttackPower: 100}},
		{proto.Class_ClassPriest, proto.Race_RaceHuman, 1, false, nil},
		{proto.Class_ClassShaman, proto.Race_RaceOrc, 1, false, nil},
		{proto.Class_ClassMage, proto.Race_RaceHuman, 1, false, nil},
		{proto.Class_ClassWarlock, proto.Race_RaceHuman, 1, false, nil},
		{proto.Class_ClassDruid, proto.Race_RaceTauren, 1, false, nil},
	} {
		t.Run(tc.class.String(), func(t *testing.T) {
			gained := statsFromHundredAgility(t, tc.class, tc.race)
			critPerHundred := 100 * core.CritPerAgiMaxLevel[tc.class]
			wanted := map[stats.Stat]float64{stats.DodgePercent: tc.dodgePerCrit * critPerHundred}
			if tc.crit {
				wanted[stats.PhysicalCritPercent] = critPerHundred
			}
			maps.Copy(wanted, tc.gained)
			for stat, want := range wanted {
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
		dodgePerCrit float64
	}{
		{proto.Class_ClassWarrior, proto.Race_RaceHuman, 0, 1},
		{proto.Class_ClassPaladin, proto.Race_RaceHuman, 0.007, 1},
		{proto.Class_ClassHunter, proto.Race_RaceNightElf, 0, 2},
		{proto.Class_ClassRogue, proto.Race_RaceHuman, 0, 2},
		{proto.Class_ClassPriest, proto.Race_RaceHuman, 0.03, 1},
		{proto.Class_ClassShaman, proto.Race_RaceOrc, 0.017, 1},
		{proto.Class_ClassMage, proto.Race_RaceHuman, 0.032, 1},
		{proto.Class_ClassWarlock, proto.Race_RaceHuman, 0.02, 1},
		{proto.Class_ClassDruid, proto.Race_RaceTauren, 0.009, 1},
	} {
		t.Run(tc.class.String(), func(t *testing.T) {
			character, sheet := nakedSheet(t, tc.class, tc.race)
			if got := character.PseudoStats.BaseDodgeChance; math.Abs(got-tc.dodge) > 1e-12 {
				t.Errorf("base dodge %v, want %v", got, tc.dodge)
			}
			agilityDodge := sheet[stats.Agility] * tc.dodgePerCrit * core.CritPerAgiMaxLevel[tc.class]
			if got := sheet[stats.DodgePercent]; math.Abs(got-agilityDodge) > 1e-9 {
				t.Errorf("dodge from %v agility %v%%, want %v%%", sheet[stats.Agility], got, agilityDodge)
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

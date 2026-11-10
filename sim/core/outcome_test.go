package core

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func TestCritChanceSubtractsCritTakenReduction(t *testing.T) {
	tests := []struct {
		name             string
		rawChance        float64
		reducedCritTaken float64
		want             float64
	}{
		{name: "defense immunity", rawChance: 0.05, reducedCritTaken: 0.05, want: 0},
		{name: "partial reduction", rawChance: 0.05, reducedCritTaken: 0.03, want: 0.02},
		{name: "reduction beyond the raw chance", rawChance: 0.05, reducedCritTaken: 0.06, want: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			target := &Unit{PseudoStats: stats.NewPseudoStats()}
			target.PseudoStats.ReducedCritTakenPercent = test.reducedCritTaken

			if got := getCritChance(test.rawChance, target); math.Abs(got-test.want) > 1e-9 {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}
}

// Reduced crit taken, from Defense or bonus stats, lowers physical crit only.
func TestCritTakenReductionLowersPhysicalCritOnly(t *testing.T) {
	attacker := &Unit{Type: EnemyUnit}
	attacker.stats[stats.PhysicalCritPercent] = 20
	attacker.stats[stats.SpellCritPercent] = 20
	defender := &Unit{Type: PlayerUnit, PseudoStats: stats.NewPseudoStats()}
	defender.stats[stats.ReducedCritTakenPercent] = 4
	defender.AddReducedCritTakenPercent(0.01)
	attacker.AttackTables = []*AttackTable{NewAttackTable(attacker, defender)}
	spell := &Spell{Unit: attacker, ProcMask: ProcMaskMeleeMHSpecial}

	if got := spell.PhysicalCritChance(attacker.AttackTables[0]); math.Abs(got-0.15) > 1e-9 {
		t.Errorf("physical crit chance %v, want 0.15", got)
	}
	if got := spell.SpellCritChance(defender); math.Abs(got-0.2) > 1e-9 {
		t.Errorf("spell crit chance %v, want 0.2", got)
	}
}

// A level 60 attacker with 0.05 crit per Agility against a level 63 target loses 3% to the level and
// up to 1.8% more of the crit beyond its base and Agility crit, from any source.
func TestPhysicalCritSuppressionCountsOnlyAuraCrit(t *testing.T) {
	for _, tc := range []struct {
		name            string
		targetLevel     int32
		procMask        ProcMask
		agility         float64
		baseCrit        float64
		auraCrit        float64
		rangedCrit      float64
		bonusCrit       float64
		critByCaster    float64
		wantCrit        float64
		wantSuppression float64
	}{
		{name: "no aura crit", agility: 100, wantCrit: 2.0, wantSuppression: 3.0},
		{name: "1% aura crit", agility: 100, auraCrit: 1, wantCrit: 2.0, wantSuppression: 4.0},
		{name: "5% aura crit", agility: 100, auraCrit: 5, wantCrit: 5.2, wantSuppression: 4.8},
		{name: "double Agility, no aura crit", agility: 200, wantCrit: 7.0, wantSuppression: 3.0},
		{name: "base crit and 1% aura crit", agility: 100, baseCrit: 3, auraCrit: 1, wantCrit: 5.0, wantSuppression: 4.0},
		{name: "3% ranged crit on a ranged auto", procMask: ProcMaskRangedAuto, agility: 100, rangedCrit: 3, wantCrit: 3.2, wantSuppression: 4.8},
		{name: "3% ranged crit on a melee auto", agility: 100, rangedCrit: 3, wantCrit: 2.0, wantSuppression: 3.0},
		{name: "1% crit of the ability's own is aura crit", agility: 100, bonusCrit: 1, wantCrit: 2.0, wantSuppression: 4.0},
		{name: "5% crit of the ability's own is aura crit", agility: 100, bonusCrit: 5, wantCrit: 5.2, wantSuppression: 4.8},
		{name: "6% crit of the ability's own caps at 1.8%", agility: 100, bonusCrit: 6, wantCrit: 6.2, wantSuppression: 4.8},
		{name: "2% crit from a target aura for the caster", agility: 100, critByCaster: 2, wantCrit: 2.2, wantSuppression: 4.8},
		{name: "level 62 target, 5% aura crit", targetLevel: 62, agility: 100, auraCrit: 5, wantCrit: 8.0, wantSuppression: 2.0},
		{name: "level 61 target, 5% aura crit", targetLevel: 61, agility: 100, auraCrit: 5, wantCrit: 9.0, wantSuppression: 1.0},
		{name: "negative aura crit adds no crit back", agility: 200, auraCrit: -2, wantCrit: 5.0, wantSuppression: 3.0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			targetLevel := tc.targetLevel
			if targetLevel == 0 {
				targetLevel = 63
			}
			procMask := tc.procMask
			if procMask == ProcMaskUnknown {
				procMask = ProcMaskMeleeMHAuto
			}

			attacker, defender := attackTableUnits(PlayerUnit, 60, EnemyUnit, targetLevel)
			attacker.basePhysicalCritPercent = tc.baseCrit
			attacker.critPerAgility = 0.05
			attacker.stats[stats.Agility] = tc.agility
			attacker.stats[stats.PhysicalCritPercent] = tc.baseCrit + tc.agility*0.05 + tc.auraCrit
			attacker.stats[stats.RangedCritPercent] = tc.rangedCrit
			table := NewAttackTable(attacker, defender)
			if tc.critByCaster != 0 {
				EnableCritChanceByCaster(0, 1, table, func(*Spell) float64 { return tc.critByCaster })
			}
			spell := &Spell{Unit: attacker, ProcMask: procMask, BonusCritPercent: tc.bonusCrit}

			totalCrit := attacker.stats[stats.PhysicalCritPercent] + tc.bonusCrit + tc.critByCaster
			if procMask.Matches(ProcMaskRanged) {
				totalCrit += tc.rangedCrit
			}

			got := spell.PhysicalCritChance(table) * 100
			if math.Abs(got-tc.wantCrit) > 1e-9 {
				t.Errorf("crit chance %v%%, want %v%%", got, tc.wantCrit)
			}
			if suppression := totalCrit - got; math.Abs(suppression-tc.wantSuppression) > 1e-9 {
				t.Errorf("%v%% of %v%% crit suppressed, want %v%%", suppression, totalCrit, tc.wantSuppression)
			}
		})
	}
}

// A 5% crit mod on the weapon tester, which has no crit beyond its base crit, against the level 63
// target: every crit mod is aura crit, whatever its scope.
func TestEveryCritModCountsAsAuraCrit(t *testing.T) {
	for _, tc := range []struct {
		name            string
		mod             SpellModConfig
		ranged          bool
		wantSuppression float64
	}{
		{"a sword-bound mod", SpellModConfig{ProcMask: ProcMaskMelee, EquipSubclass: WeaponSubclassMask(proto.WeaponType_WeaponTypeSword)},
			false, 4.8},
		{"a ranged rating mod", SpellModConfig{ProcMask: ProcMaskRanged, DefenseType: DefenseTypeRanged, FromRating: true},
			true, 4.8},
		{"a mod on the main-hand auto alone", SpellModConfig{ProcMask: ProcMaskMeleeMHAuto},
			false, 4.8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, character := NewWeaponTestSim(t, proto.Race_RaceUndead, WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, WeaponTestBowID), nil, nil)
			tc.mod.Kind, tc.mod.FloatValue = SpellMod_BonusCrit_Percent, 5
			character.AddStaticMod(tc.mod)

			auto := character.AutoAttacks.MHAuto()
			critPercent := character.GetStat(stats.PhysicalCritPercent)
			if tc.ranged {
				auto = character.AutoAttacks.RangedAuto()
				critPercent += character.GetStat(stats.RangedCritPercent)
			}
			if auto.BonusCritPercent != 5 {
				t.Fatalf("the auto carries %v%% crit of its own, want the mod's 5%%", auto.BonusCritPercent)
			}
			table := character.AttackTables[character.CurrentTarget.UnitIndex]

			if got := auto.physicalCritSuppression(table, critPercent+auto.BonusCritPercent) * 100; !WithinToleranceFloat64(tc.wantSuppression, got, 1e-9) {
				t.Errorf("the auto loses %v%% crit, want %v%%", got, tc.wantSuppression)
			}
			pseudoStats := character.GetPseudoStatsProto()
			sheetSuppression := character.meleeCritCapProto(pseudoStats[proto.PseudoStat_PseudoStatMeleeCritPercent]).Suppression
			if tc.ranged {
				sheetSuppression = pseudoStats[proto.PseudoStat_PseudoStatRangedCritSuppressionPercent]
			}
			if got := sheetSuppression; !WithinToleranceFloat64(tc.wantSuppression, got, 1e-9) {
				t.Errorf("the sheet shows %v%% suppression, want %v%%", got, tc.wantSuppression)
			}
		})
	}
}

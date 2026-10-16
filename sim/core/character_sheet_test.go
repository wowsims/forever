package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

var sheetAttackPseudoStats = []proto.PseudoStat{
	proto.PseudoStat_PseudoStatMeleeHitPercent,
	proto.PseudoStat_PseudoStatMeleeCritPercent,
	proto.PseudoStat_PseudoStatMainHandHitRating,
	proto.PseudoStat_PseudoStatMainHandCritRating,
	proto.PseudoStat_PseudoStatOffHandHitPercent,
	proto.PseudoStat_PseudoStatOffHandCritPercent,
	proto.PseudoStat_PseudoStatOffHandHitRating,
	proto.PseudoStat_PseudoStatOffHandCritRating,
	proto.PseudoStat_PseudoStatRangedHitPercent,
	proto.PseudoStat_PseudoStatRangedCritPercent,
	proto.PseudoStat_PseudoStatRangedHitRating,
	proto.PseudoStat_PseudoStatRangedCritRating,
}

func sheetDeltaOf(character *Character, change func()) map[proto.PseudoStat]float64 {
	before := character.GetPseudoStatsProto()
	change()
	after := character.GetPseudoStatsProto()
	delta := make(map[proto.PseudoStat]float64)
	for _, pseudoStat := range sheetAttackPseudoStats {
		delta[pseudoStat] = after[pseudoStat] - before[pseudoStat]
	}
	return delta
}

func checkSheetDelta(t *testing.T, delta map[proto.PseudoStat]float64, want map[proto.PseudoStat]float64) {
	t.Helper()
	for _, pseudoStat := range sheetAttackPseudoStats {
		if !WithinToleranceFloat64(want[pseudoStat], delta[pseudoStat], 1e-9) {
			t.Errorf("%s moved by %v, want %v", pseudoStat, delta[pseudoStat], want[pseudoStat])
		}
	}
}

func TestSheetAttackPercentsMatchTheAutos(t *testing.T) {
	_, character := NewWeaponTestSim(t, proto.Race_RaceUndead, WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, WeaponTestBowID), nil, nil)
	for _, config := range []SpellModConfig{
		{Kind: SpellMod_BonusCrit_Percent, FloatValue: 5, ProcMask: ProcMaskMelee, EquipSubclass: WeaponSubclassMask(proto.WeaponType_WeaponTypeSword)},
		{Kind: SpellMod_BonusCrit_Percent, FloatValue: 3, EquipSubclass: dbcenums.ITEM_SUBCLASS_MASK_WEAPON_SHOOTING},
		{Kind: SpellMod_BonusCrit_Percent, FloatValue: 2, ProcMask: ProcMaskMeleeOH},
		{Kind: SpellMod_BonusCrit_Percent, FloatValue: 1},
		{Kind: SpellMod_BonusCrit_Percent, FloatValue: 7, ClassMask: 1 << 40},
		{Kind: SpellMod_BonusCrit_Percent, FloatValue: 7, School: SpellSchoolFire},
		{Kind: SpellMod_BonusHit_Percent, FloatValue: 2, ProcMask: ProcMaskMeleeOH, EquipHand: ProcMaskMeleeOH},
		{Kind: SpellMod_BonusHit_Percent, FloatValue: 3, ProcMask: ProcMaskRanged, EquipSubclass: dbcenums.ITEM_SUBCLASS_MASK_WEAPON_SHOOTING},
		{Kind: SpellMod_BonusHit_Percent, FloatValue: 11},
	} {
		character.AddStaticMod(config)
	}
	character.AddDynamicMod(SpellModConfig{Kind: SpellMod_BonusCrit_Percent, FloatValue: 13})

	pseudoStats := character.GetPseudoStatsProto()
	hit, crit := character.GetStat(stats.PhysicalHitPercent), character.GetStat(stats.PhysicalCritPercent)
	for _, tc := range []struct {
		name                string
		probe               ProcMask
		auto                *Spell
		hitShown, critShown proto.PseudoStat
		hitBase, critBase   float64
		wantHit, wantCrit   float64
	}{
		{"main hand", ProcMaskMeleeMHAuto, character.AutoAttacks.MHAuto(),
			proto.PseudoStat_PseudoStatMeleeHitPercent, proto.PseudoStat_PseudoStatMeleeCritPercent, hit, crit, 11, 6},
		{"off hand", ProcMaskMeleeOHAuto, character.AutoAttacks.OHAuto(),
			proto.PseudoStat_PseudoStatOffHandHitPercent, proto.PseudoStat_PseudoStatOffHandCritPercent, hit, crit, 13, 3},
		{"ranged", ProcMaskRangedAuto, character.AutoAttacks.RangedAuto(),
			proto.PseudoStat_PseudoStatRangedHitPercent, proto.PseudoStat_PseudoStatRangedCritPercent,
			hit + character.GetStat(stats.RangedHitPercent), crit + character.GetStat(stats.RangedCritPercent), 14, 4},
	} {
		probed := character.ScopedPercents(tc.probe)
		probedHit, probedCrit := probed.Hit, probed.Crit
		if probedHit != tc.wantHit || probedCrit != tc.wantCrit {
			t.Errorf("%s: probed %v hit and %v crit, want %v and %v", tc.name, probedHit, probedCrit, tc.wantHit, tc.wantCrit)
		}
		if probedHit != tc.auto.BonusHitPercent || probedCrit != tc.auto.BonusCritPercent {
			t.Errorf("%s: probed %v hit and %v crit, the auto attack carries %v and %v",
				tc.name, probedHit, probedCrit, tc.auto.BonusHitPercent, tc.auto.BonusCritPercent)
		}
		if got, want := pseudoStats[tc.hitShown], tc.hitBase+tc.auto.BonusHitPercent; !WithinToleranceFloat64(want, got, 1e-9) {
			t.Errorf("%s: sheet shows %v hit, want %v", tc.name, got, want)
		}
		if got, want := pseudoStats[tc.critShown], tc.critBase+tc.auto.BonusCritPercent; !WithinToleranceFloat64(want, got, 1e-9) {
			t.Errorf("%s: sheet shows %v crit, want %v", tc.name, got, want)
		}
	}
}

func TestSheetOffHandIsZeroWithoutDualWielding(t *testing.T) {
	_, character := NewWeaponTestSim(t, proto.Race_RaceUndead, WeaponTestSlots(WeaponTestSwordID, WeaponTestShieldID, WeaponTestBowID), nil, nil)
	character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusCrit_Percent, FloatValue: 4, ProcMask: ProcMaskMelee})
	character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusHit_Percent, FloatValue: 2, ProcMask: ProcMaskMeleeOH})

	pseudoStats := character.GetPseudoStatsProto()
	for _, pseudoStat := range []proto.PseudoStat{
		proto.PseudoStat_PseudoStatOffHandHitPercent,
		proto.PseudoStat_PseudoStatOffHandCritPercent,
		proto.PseudoStat_PseudoStatOffHandHitRating,
		proto.PseudoStat_PseudoStatOffHandCritRating,
	} {
		if pseudoStats[pseudoStat] != 0 {
			t.Errorf("sword and shield show %s %v, want 0", pseudoStat, pseudoStats[pseudoStat])
		}
	}
}

func TestSheetRangedHitModRaisesOnlyRanged(t *testing.T) {
	_, character := NewWeaponTestSim(t, proto.Race_RaceUndead, WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, WeaponTestBowID), nil, nil)
	delta := sheetDeltaOf(character, func() {
		character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusHit_Percent, FloatValue: 3, ProcMask: ProcMaskRanged, EquipSubclass: dbcenums.ITEM_SUBCLASS_MASK_WEAPON_SHOOTING})
	})
	checkSheetDelta(t, delta, map[proto.PseudoStat]float64{
		proto.PseudoStat_PseudoStatRangedHitPercent: 3,
	})
}

func TestSheetSwordCritReachesOnlyTheSwordHand(t *testing.T) {
	_, character := NewWeaponTestSim(t, proto.Race_RaceUndead, WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, WeaponTestBowID), nil, nil)
	delta := sheetDeltaOf(character, func() {
		character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusCrit_Percent, FloatValue: 5, ProcMask: ProcMaskMelee,
			EquipSubclass: WeaponSubclassMask(proto.WeaponType_WeaponTypeSword)})
	})
	checkSheetDelta(t, delta, map[proto.PseudoStat]float64{
		proto.PseudoStat_PseudoStatMeleeCritPercent: 5,
	})
}

func TestSheetSwordGatedCritShowsInEveryCritRow(t *testing.T) {
	sim, character := NewWeaponTestSim(t, proto.Race_RaceUndead,
		WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, WeaponTestBowID),
		WeaponTestSlots(WeaponTestDaggerID, WeaponTestDaggerID, WeaponTestBowID), func(character *Character) {
			character.NewEquipGatedStats(EquipRequirement{SubclassMask: WeaponSubclassMask(proto.WeaponType_WeaponTypeSword)},
				stats.Stats{stats.PhysicalCritPercent: 2, stats.SpellCritPercent: 2}).SetLevel(nil, 1)
		})
	spellCrit := character.GetStat(stats.SpellCritPercent)
	delta := sheetDeltaOf(character, func() {
		character.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)
	})
	checkSheetDelta(t, delta, map[proto.PseudoStat]float64{
		proto.PseudoStat_PseudoStatMeleeCritPercent:   -2,
		proto.PseudoStat_PseudoStatOffHandCritPercent: -2,
		proto.PseudoStat_PseudoStatRangedCritPercent:  -2,
	})
	if got := character.GetStat(stats.SpellCritPercent) - spellCrit; !WithinToleranceFloat64(-2, got, 1e-9) {
		t.Errorf("spell crit moved by %v, want -2", got)
	}
}

func TestSheetRatingShowsOnlyTheModsFromARating(t *testing.T) {
	const meleeWeapons = 173555
	for _, tc := range []struct {
		name       string
		fromRating bool
		want       map[proto.PseudoStat]float64
	}{
		{"a percent talent", false, map[proto.PseudoStat]float64{
			proto.PseudoStat_PseudoStatMeleeCritPercent:   5,
			proto.PseudoStat_PseudoStatOffHandCritPercent: 5,
		}},
		{"a rating", true, map[proto.PseudoStat]float64{
			proto.PseudoStat_PseudoStatMeleeCritPercent:   5,
			proto.PseudoStat_PseudoStatOffHandCritPercent: 5,
			proto.PseudoStat_PseudoStatMainHandCritRating: 5 * CritRatingPerCritPercent,
			proto.PseudoStat_PseudoStatOffHandCritRating:  5 * CritRatingPerCritPercent,
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, character := NewWeaponTestSim(t, proto.Race_RaceUndead, WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, WeaponTestBowID), nil, nil)
			delta := sheetDeltaOf(character, func() {
				character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusCrit_Percent, FloatValue: 5, ProcMask: ProcMaskMelee,
					EquipSubclass: meleeWeapons, FromRating: tc.fromRating})
			})
			checkSheetDelta(t, delta, tc.want)
		})
	}
}

func TestSheetRatingsAddTheOneRating(t *testing.T) {
	_, character := NewWeaponTestSim(t, proto.Race_RaceUndead, WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, WeaponTestBowID), nil, nil)
	pseudoStats := character.GetPseudoStatsProto()
	hitRating, critRating := character.GetStat(stats.HitRating), character.GetStat(stats.CritRating)
	for _, tc := range []struct {
		shown proto.PseudoStat
		want  float64
	}{
		{proto.PseudoStat_PseudoStatMainHandHitRating, hitRating + character.ScopedPercents(ProcMaskMeleeMHAuto).HitFromRating*HitRatingPerHitPercent},
		{proto.PseudoStat_PseudoStatOffHandCritRating, critRating + character.ScopedPercents(ProcMaskMeleeOHAuto).CritFromRating*CritRatingPerCritPercent},
		{proto.PseudoStat_PseudoStatRangedHitRating, hitRating + (character.GetStat(stats.RangedHitPercent)+
			character.ScopedPercents(ProcMaskRangedAuto).HitFromRating)*HitRatingPerHitPercent},
	} {
		if !WithinToleranceFloat64(tc.want, pseudoStats[tc.shown], 1e-9) {
			t.Errorf("%s = %v, want %v", tc.shown, pseudoStats[tc.shown], tc.want)
		}
	}
}

func TestSheetRangedRatingShowsTheRangedOnlyPercent(t *testing.T) {
	sim, character := NewWeaponTestSim(t, proto.Race_RaceUndead, WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, WeaponTestBowID), nil, nil)
	delta := sheetDeltaOf(character, func() {
		character.AddStatDynamic(sim, stats.RangedHitPercent, 2)
		character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusHit_Percent, FloatValue: 1, ProcMask: ProcMaskRanged, EquipSubclass: dbcenums.ITEM_SUBCLASS_MASK_WEAPON_SHOOTING,
			FromRating: true})
	})
	checkSheetDelta(t, delta, map[proto.PseudoStat]float64{
		proto.PseudoStat_PseudoStatRangedHitPercent: 3,
		proto.PseudoStat_PseudoStatRangedHitRating:  3 * HitRatingPerHitPercent,
	})
}

// The ranged offsets the test suite's stats snapshot records: the sheet's ranged row less the physical
// percent.
func sheetRangedOffsets(character *Character) (hit, crit float64) {
	pseudoStats := character.GetPseudoStatsProto()
	return pseudoStats[proto.PseudoStat_PseudoStatRangedHitPercent] - character.GetStat(stats.PhysicalHitPercent),
		pseudoStats[proto.PseudoStat_PseudoStatRangedCritPercent] - character.GetStat(stats.PhysicalCritPercent)
}

func TestSheetRangedOffsetsLeaveOutTheMainHandsMods(t *testing.T) {
	sim, character := NewWeaponTestSim(t, proto.Race_RaceUndead, WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, WeaponTestBowID), nil, nil)
	hitBefore, critBefore := sheetRangedOffsets(character)
	character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusCrit_Percent, FloatValue: 5, ProcMask: ProcMaskMelee,
		EquipSubclass: WeaponSubclassMask(proto.WeaponType_WeaponTypeSword)})
	character.AddStatDynamic(sim, stats.RangedHitPercent, 2)
	character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusHit_Percent, FloatValue: 1, ProcMask: ProcMaskRanged, EquipSubclass: dbcenums.ITEM_SUBCLASS_MASK_WEAPON_SHOOTING})
	hitAfter, critAfter := sheetRangedOffsets(character)

	if got := critAfter - critBefore; got != 0 {
		t.Errorf("a sword's crit moved the ranged crit offset by %v, want 0", got)
	}
	if got := hitAfter - hitBefore; !WithinToleranceFloat64(3, got, 1e-9) {
		t.Errorf("ranged hit and a ranged hit mod moved the ranged hit offset by %v, want 3", got)
	}
}

// The weapon tester has no crit beyond its base crit, so the level 63 target takes 3% off its autos
// and up to 1.8% more of any crit a buff adds. The sheet's crit less the suppression is the autos' crit.
func TestSheetCritSuppressionIsTheAutosAgainstTheTarget(t *testing.T) {
	sim, character := NewWeaponTestSim(t, proto.Race_RaceUndead, WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, WeaponTestBowID), nil, nil)
	table := character.AttackTables[character.CurrentTarget.UnitIndex]
	for _, tc := range []struct {
		buffCrit, want float64
	}{
		{0, 3.0},
		{1, 4.0},
		{2, 4.8},
		{20, 4.8},
	} {
		character.AddStatDynamic(sim, stats.PhysicalCritPercent, tc.buffCrit)
		pseudoStats := character.GetPseudoStatsProto()
		for _, row := range []struct {
			name              string
			suppression, crit proto.PseudoStat
			auto              *Spell
		}{
			{"melee", proto.PseudoStat_PseudoStatMeleeCritSuppressionPercent, proto.PseudoStat_PseudoStatMeleeCritPercent, character.AutoAttacks.MHAuto()},
			{"ranged", proto.PseudoStat_PseudoStatRangedCritSuppressionPercent, proto.PseudoStat_PseudoStatRangedCritPercent, character.AutoAttacks.RangedAuto()},
		} {
			if got := pseudoStats[row.suppression]; !WithinToleranceFloat64(tc.want, got, 1e-9) {
				t.Errorf("%v%% crit buff: the %s row shows %v%% suppression, want %v%%", tc.buffCrit, row.name, got, tc.want)
			}
			shown := max(pseudoStats[row.crit]-pseudoStats[row.suppression]-table.Defender.PseudoStats.ReducedCritTakenPercent*100, 0)
			if rolled := row.auto.PhysicalCritChance(table) * 100; !WithinToleranceFloat64(rolled, shown, 1e-9) {
				t.Errorf("%v%% crit buff: the %s row shows %v%% crit after suppression, the auto rolls %v%%", tc.buffCrit, row.name, shown, rolled)
			}
		}
		character.AddStatDynamic(sim, stats.PhysicalCritPercent, -tc.buffCrit)
	}
}

// A hit mod with ShowAsSchoolHit shows in those schools from the talents stage on, with none of the
// spells it names registered; a class-masked hit mod without it shows nowhere.
func TestSheetShowsSchoolHitFromTheTalentsStage(t *testing.T) {
	weaponTestTalents = func(character *Character) {
		character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusHit_Percent, FloatValue: 3, ClassMask: 1 << 40, ShowAsSchoolHit: SpellSchoolFrostfire})
		character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusHit_Percent, FloatValue: 5, ClassMask: 1 << 41})
	}
	defer func() { weaponTestTalents = nil }()
	raid := weaponTestRaid(proto.Race_RaceUndead, WeaponTestSlots(WeaponTestSwordID, 0, 0), nil)
	playerStats := ComputeStats(&proto.ComputeStatsRequest{Raid: raid, Encounter: weaponTestEncounter()}).RaidStats.Parties[0].Players[0]

	for name, stage := range map[string]*proto.UnitStats{"talents": playerStats.TalentsStats, "final": playerStats.FinalStats} {
		for pseudoStat, want := range map[proto.PseudoStat]float64{
			proto.PseudoStat_PseudoStatSchoolHitPercentArcane: 0,
			proto.PseudoStat_PseudoStatSchoolHitPercentFire:   3,
			proto.PseudoStat_PseudoStatSchoolHitPercentFrost:  3,
			proto.PseudoStat_PseudoStatSpellHitPercent:        0,
			proto.PseudoStat_PseudoStatMeleeHitPercent:        0,
			proto.PseudoStat_PseudoStatRangedHitPercent:       0,
		} {
			if got := stage.PseudoStats[pseudoStat] - playerStats.GearStats.PseudoStats[pseudoStat]; !WithinToleranceFloat64(want, got, 1e-9) {
				t.Errorf("the %s stage shows %v more %s than the gear stage, want %v", name, got, pseudoStat, want)
			}
		}
	}
}

func TestSheetMainHandTableIsTheMainHandAutoAgainstTheTarget(t *testing.T) {
	for _, tc := range []struct {
		swordSkill                 int32
		miss, dodge, parry, glance float64
	}{
		{300, 8, 6.5, 14, 40},
		{285, 11, 8, 23, 70},
	} {
		raid := weaponTestRaid(proto.Race_RaceUndead, WeaponTestSlots(WeaponTestSwordID, 0, 0), nil)
		raid.Parties[0].Players[0].WeaponSkills = []*proto.WeaponSkillValue{{Skill: proto.WeaponSkill_WeaponSkillSwords, Value: tc.swordSkill}}
		sim := NewSim(&proto.RaidSimRequest{SimOptions: &proto.SimOptions{RandomSeed: 1}, Raid: raid, Encounter: weaponTestEncounter()}, simsignals.CreateSignals())
		pseudoStats := sim.Raid.Parties[0].Players[0].GetCharacter().GetPseudoStatsProto()
		for _, row := range []struct {
			pseudoStat proto.PseudoStat
			want       float64
		}{
			{proto.PseudoStat_PseudoStatMainHandMissPercent, tc.miss},
			{proto.PseudoStat_PseudoStatMainHandDodgePercent, tc.dodge},
			{proto.PseudoStat_PseudoStatMainHandParryPercent, tc.parry},
			{proto.PseudoStat_PseudoStatMainHandGlancePercent, tc.glance},
		} {
			if got := pseudoStats[row.pseudoStat]; !WithinToleranceFloat64(row.want, got, 1e-9) {
				t.Errorf("%d sword skill against a level 63 target: the sheet shows %v for %s, want %v", tc.swordSkill, got, row.pseudoStat, row.want)
			}
		}
	}
}

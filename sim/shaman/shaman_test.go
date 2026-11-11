package shaman_test

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/shaman"
	"github.com/wowsims/forever/sim/shaman/enhancement"
)

func init() {
	enhancement.RegisterEnhancementShaman()
	common.RegisterAllEffects()
}

// The Enhancement suite's talents, which take Elemental Devastation.
const enhancementTalents = "053503102-053030310051012251"

// The raid of one Enhancement shaman with the given talents.
func enhancementRaid(talents string) *proto.Raid {
	player := &proto.Player{
		Name:          "Enhancement",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassShaman,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: talents,
		Spec:          &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{Options: &proto.EnhancementShaman_Options{ClassOptions: &proto.ShamanOptions{}}}},
		Rotation:      &proto.APLRotation{},
	}
	return core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
}

// A totem's trainer rank grants its own row's amount and takes the book rank's slot, and the book rank takes it back.
func TestTrainerTotemRanksBesideBookRanks(t *testing.T) {
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       enhancementRaid(enhancementTalents),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sham := sim.Raid.Parties[0].Players[0].(shaman.ShamanAgent).GetShaman()

	for _, tc := range []struct {
		name                                           string
		stat                                           stats.Stat
		trainerTotem, trainerBuff, bookTotem, bookBuff int32
	}{
		{"Strength of Earth Totem", stats.Strength, 10442, 10441, 25361, 25362},
		{"Grace of Air Totem", stats.Agility, 10627, 10626, 25359, 25360},
	} {
		cast := func(totem, buff, otherBuff int32, wantGain float64) {
			t.Helper()
			before := sham.GetStat(tc.stat)
			sham.GetSpell(core.ActionID{SpellID: totem}).SkipCastAndApplyEffects(sim, &sham.Unit)
			if got := sham.GetStat(tc.stat) - before; got != wantGain {
				t.Errorf("%s: casting %d changes the stat by %v, want %v", tc.name, totem, got, wantGain)
			}
			if !sham.GetAuraByID(core.ActionID{SpellID: buff}).IsActive() || sham.GetAuraByID(core.ActionID{SpellID: otherBuff}).IsActive() {
				t.Errorf("%s: casting %d does not leave buff %d alone in the slot", tc.name, totem, buff)
			}
		}

		trainerAmount := spelldata.MustFind(tc.trainerBuff).EffectN(1).Average()
		bookAmount := spelldata.MustFind(tc.bookBuff).EffectN(1).Average()
		cast(tc.trainerTotem, tc.trainerBuff, tc.bookBuff, trainerAmount)
		cast(tc.bookTotem, tc.bookBuff, tc.trainerBuff, bookAmount-trainerAmount)
		cast(tc.trainerTotem, tc.trainerBuff, tc.bookBuff, trainerAmount-bookAmount)
	}
}

// The melee crit cap counts the crit Elemental Devastation's buff grants as taken, and a shaman
// without the talent counts none.
func TestMeleeCritCapCountsElementalDevastation(t *testing.T) {
	critCap := func(talents string) *proto.MeleeCritCapInfo {
		t.Helper()
		info := core.ComputeStats(&proto.ComputeStatsRequest{Raid: enhancementRaid(talents), Encounter: core.MakeSingleTargetEncounter(0)}).RaidStats.Parties[0].Players[0].MeleeCritCap
		if info == nil {
			t.Fatalf("talents %q: the sheet has no melee crit cap", talents)
		}
		return info
	}

	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       enhancementRaid(enhancementTalents),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	sham := sim.Raid.Parties[0].Players[0].(shaman.ShamanAgent).GetShaman()
	devastation := sham.GetAuraByID(core.ActionID{SpellID: 30165})
	if devastation == nil {
		t.Fatal("Elemental Devastation's buff is not registered")
	}
	before := sham.GetStat(stats.PhysicalCritPercent)
	devastation.Activate(sim)
	granted := sham.GetStat(stats.PhysicalCritPercent) - before
	if granted <= 0 {
		t.Fatalf("Elemental Devastation's buff grants %v crit, want above 0", granted)
	}

	if sham.PseudoStats.InFrontOfTarget {
		t.Fatal("the shaman stands in front of the target, and the room below leaves out parry and block")
	}
	auto := sham.AutoAttacks.MHAuto()
	table := sham.AttackTables[sham.CurrentTarget.UnitIndex]
	critWithBuff := auto.PhysicalCritChance(table)
	room := 1 - auto.GetPhysicalMissChance(table) - sham.CurrentTarget.GetTotalDodgeChanceAsDefender(auto, table) - table.MH.BaseGlanceChance

	with := critCap(enhancementTalents)
	if !core.WithinToleranceFloat64(granted, with.SpecOffset, 1e-9) {
		t.Errorf("spec offset %v, want the buff's %v crit", with.SpecOffset, granted)
	}
	if want := 100*critWithBuff - 100*room; !core.WithinToleranceFloat64(want, with.Delta, 1e-9) {
		t.Errorf("delta %v, want the crit the auto rolls with the buff up less the white-hit table's room for it, %v", with.Delta, want)
	}
	if without := critCap("-053030310051012251"); without.SpecOffset != 0 {
		t.Errorf("without the talent the spec offset is %v, want 0", without.SpecOffset)
	}

}

// Fire Nova's hits are a spell the cast triggers, and the row lacks Not a Proc, so Elemental
// Devastation, which cannot proc from procs, never hears their crits. It hears an Earth Shock's.
func TestFireNovaHitsAreProcs(t *testing.T) {
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       enhancementRaid(enhancementTalents),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()

	sham := sim.Raid.Parties[0].Players[0].(shaman.ShamanAgent).GetShaman()
	target := sham.CurrentTarget
	devastation := sham.GetAuraByID(core.ActionID{SpellID: 30165})
	novaHits := sham.GetSpell(core.ActionID{SpellID: sham.FireNova.ActionID.SpellID, Tag: 1})
	earthShock := sham.EarthShock[len(sham.EarthShock)-1]
	if devastation == nil || novaHits == nil {
		t.Fatalf("Elemental Devastation %v, Fire Nova's hits %v; want both registered", devastation, novaHits)
	}

	devastatedByCrit := func(cast *core.Spell, hits *core.Spell) bool {
		t.Helper()
		devastation.Deactivate(sim)
		ended := false
		casts := 0
		for casts < 1000 && !ended {
			crits := hits.SpellMetrics[target.UnitIndex].Crits
			cast.SkipCastAndApplyEffects(sim, target)
			casts++
			for end := sim.CurrentTime + core.SpellBatchWindow; sim.CurrentTime <= end && !ended; {
				ended = sim.Step()
			}
			if hits.SpellMetrics[target.UnitIndex].Crits > crits {
				return devastation.IsActive()
			}
		}
		t.Fatalf("%d casts of %v never crit", casts, cast.ActionID)
		return false
	}

	if devastatedByCrit(sham.FireNova, novaHits) {
		t.Error("a Fire Nova crit procced Elemental Devastation")
	}
	if !devastatedByCrit(earthShock, earthShock) {
		t.Error("an Earth Shock crit did not proc Elemental Devastation, so the check above proves nothing")
	}
}

// The cast deals no damage, so the results show Fire Nova as its hit's row, which holds every cast
// and its GCD; the cast's row holds neither, so no table counts a Fire Nova twice.
func TestFireNovaHitsHoldTheCastsAndCastTime(t *testing.T) {
	const fireNova, searingTotem = 408345, 10438
	castSpell := func(id int32) *proto.APLListItem {
		return &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{
			SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}},
		}}}}
	}
	player := &proto.Player{
		Name:          "Fire Nova",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassShaman,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: enhancementTalents,
		Spec:          &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{Options: &proto.EnhancementShaman_Options{ClassOptions: &proto.ShamanOptions{}}}},
		Rotation:      &proto.APLRotation{PriorityList: []*proto.APLListItem{castSpell(fireNova), castSpell(searingTotem)}},
	}
	result := core.RunRaidSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{Iterations: 5, RandomSeed: 1},
	})
	if result.Error != nil {
		t.Fatalf("the sim failed: %s", result.Error.Message)
	}

	var cast, hits *proto.ActionMetrics
	for _, action := range result.RaidMetrics.Parties[0].Players[0].Actions {
		if action.Id.GetSpellId() != fireNova {
			continue
		}
		switch action.Id.Tag {
		case 0:
			cast = action
		case 1:
			hits = action
		}
	}
	if cast == nil || hits == nil {
		t.Fatalf("Fire Nova's cast %v, its hits %v; want both in the results", cast, hits)
	}
	if hits.IsPassive {
		t.Error("Fire Nova's hits report as passive, so the results show no casts for them")
	}
	if casts, castTime := cast.Targets[0].Casts, cast.Targets[0].CastTimeMs; casts != 0 || castTime != 0 {
		t.Errorf("Fire Nova's cast row holds %d casts and %v ms of cast time; want none", casts, castTime)
	}
	// Each Fire Nova is instant and takes a 1.5 s GCD.
	hitCasts, castTime := hits.Targets[0].Casts, hits.Targets[0].CastTimeMs
	if hitCasts == 0 || castTime != float64(hitCasts)*1500 {
		t.Errorf("Fire Nova's hits hold %d casts and %v ms of cast time; want 1500 ms a cast, above 0 casts", hitCasts, castTime)
	}
}

// A Dwarf's beast bonus reaches the shaman's own Fire Nova but not a Magma Totem pulse, which deals
// the same damage to a beast as to any other creature.
func TestTotemAttacksTakeNoCreatureTypeBonus(t *testing.T) {
	raid := enhancementRaid(enhancementTalents)
	raid.Parties[0].Players[0].Race = proto.Race_RaceDwarf
	beast := core.FreshDefaultTargetConfig()
	beast.MobType = proto.MobType_MobTypeBeast
	encounter := core.MakeSingleTargetEncounter(0)
	encounter.Targets = append(encounter.Targets, beast)
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()

	sham := sim.Raid.Parties[0].Players[0].(shaman.ShamanAgent).GetShaman()
	other := sham.AttackTables[sim.Encounter.AllTargetUnits[0].UnitIndex]
	beastTable := sham.AttackTables[sim.Encounter.AllTargetUnits[1].UnitIndex]
	if beastTable.Defender.MobType != proto.MobType_MobTypeBeast || other.Defender.MobType == proto.MobType_MobTypeBeast {
		t.Fatalf("the targets are %v and %v; want a beast and another creature", beastTable.Defender.MobType, other.Defender.MobType)
	}
	beastBonus := func(spell *core.Spell) float64 {
		t.Helper()
		if spell == nil {
			t.Fatal("the spell is not registered")
		}
		return spell.AttackerDamageMultiplier(beastTable, false) / spell.AttackerDamageMultiplier(other, false)
	}

	if bonus := beastBonus(sham.GetSpell(core.ActionID{SpellID: sham.FireNova.ActionID.SpellID, Tag: 1})); bonus <= 1 {
		t.Errorf("Fire Nova's hits deal %v times the damage to a beast; want the racial's bonus above 1", bonus)
	}
	if bonus := beastBonus(sham.GetSpell(core.ActionID{SpellID: sham.MagmaTotem.ActionID.SpellID, Tag: 1})); bonus != 1 {
		t.Errorf("a Magma Totem pulse deals %v times the damage to a beast; want 1", bonus)
	}
}

// Water Shield's tooltip: a hit taken or one of the shaman's spell crits spends a globe for 2% of
// maximum mana, one globe every few seconds; a spell that does not crit spends none.
func TestWaterShieldSpendsAGlobeOnAHitTakenOrASpellCrit(t *testing.T) {
	player := &proto.Player{
		Name:          "Water Shield",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassShaman,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: enhancementTalents + "-000000001",
		Spec:          &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{Options: &proto.EnhancementShaman_Options{ClassOptions: &proto.ShamanOptions{}}}},
		Rotation:      &proto.APLRotation{},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()

	sham := sim.Raid.Parties[0].Players[0].(shaman.ShamanAgent).GetShaman()
	target := sham.CurrentTarget
	cast := sham.GetSpell(core.ActionID{SpellID: 408510})
	if cast == nil || sham.WaterShieldAura == nil {
		t.Fatalf("Water Shield %v, aura %v; want both registered under the talent", cast, sham.WaterShieldAura)
	}
	cast.SkipCastAndApplyEffects(sim, &sham.Unit)
	shield := sham.WaterShieldAura
	if got := shield.GetStacks(); got != 3 {
		t.Fatalf("the shield holds %d globes, want 3", got)
	}
	sham.SpendMana(sim, sham.MaxMana()/2, sham.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionManaRegen, Tag: 99}))
	globe := sham.MaxMana() * 0.02

	// A trigger's handler runs a spell batch window after the hit.
	settle := func() {
		for end := sim.CurrentTime + core.SpellBatchWindow; sim.CurrentTime <= end; {
			if sim.Step() {
				t.Fatal("the encounter ended")
			}
		}
	}
	// Regen over a batch window is a fraction of a globe, so a spent globe is told apart by the
	// mana it adds.
	expect := func(when string, hit func(), globes int32, spends bool) {
		t.Helper()
		before := sham.CurrentMana()
		hit()
		settle()
		if got := shield.GetStacks(); got != globes {
			t.Errorf("%s: %d globes, want %d", when, got, globes)
		}
		if gained := sham.CurrentMana() - before; (gained >= globe) != spends {
			t.Errorf("%s: gained %v mana, want a globe's %v: %v", when, gained, globe, spends)
		}
	}

	spell := &core.Spell{Unit: &sham.Unit, ProcMask: core.ProcMaskSpellDamage}
	spellHit := func(outcome core.HitOutcome) func() {
		return func() {
			shield.OnSpellHitDealt(shield, sim, spell, &core.SpellResult{Target: target, Outcome: outcome, Damage: 100})
		}
	}
	expect("a spell hit", spellHit(core.OutcomeHit), 3, false)
	expect("a spell crit", spellHit(core.OutcomeCrit), 2, true)
	expect("a second crit within the cooldown", spellHit(core.OutcomeCrit), 2, false)

	sim.CurrentTime += 4 * time.Second
	settle()
	attack := &core.Spell{Unit: target, ProcMask: core.ProcMaskMeleeMHAuto}
	expect("a hit taken", func() {
		shield.OnSpellHitTaken(shield, sim, attack, &core.SpellResult{Target: &sham.Unit, Outcome: core.OutcomeHit, Damage: 100})
	}, 1, true)
}

// Totemic Recall takes down the shaman's standing totems and returns the row's share of their base
// mana costs.
func TestTotemicRecallReturnsAShareOfTheStandingTotemsBaseCosts(t *testing.T) {
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       enhancementRaid(enhancementTalents),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()

	sham := sim.Raid.Parties[0].Players[0].(shaman.ShamanAgent).GetShaman()
	recallRow := spelldata.MustFind(36936)
	recall := sham.GetSpell(core.ActionID{SpellID: recallRow.ID})
	if recall == nil {
		t.Fatal("Totemic Recall is not registered")
	}

	totems := []*core.Spell{sham.SearingTotem, sham.WindfuryTotem}
	var baseCosts float64
	for _, totem := range totems {
		totem.SkipCastAndApplyEffects(sim, sham.CurrentTarget)
		baseCosts += float64(totem.Cost.BaseCost)
	}
	sham.SpendMana(sim, sham.MaxMana()/2, sham.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionManaRegen, Tag: 99}))

	before := sham.CurrentMana()
	recall.SkipCastAndApplyEffects(sim, &sham.Unit)
	if got, want := sham.CurrentMana()-before, baseCosts*recallRow.EffectN(1).Percent(); got != want || want == 0 {
		t.Errorf("Totemic Recall returned %v mana, want %v above 0", got, want)
	}
	for _, totem := range totems {
		if totem.RelatedSelfBuff.IsActive() {
			t.Errorf("%s still stands after Totemic Recall", totem.ActionID)
		}
	}
}

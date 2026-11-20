package shaman_test

import (
	"maps"
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/shaman"
	"github.com/wowsims/forever/sim/shaman/elemental"
	"github.com/wowsims/forever/sim/shaman/enhancement"
)

func init() {
	elemental.RegisterElementalShaman()
	enhancement.RegisterEnhancementShaman()
	common.RegisterAllEffects()
}

// The Enhancement suite's talents, which take Elemental Devastation.
const enhancementTalents = "053503102-053030310051012251"

// The Elemental suite's talents, which take Elemental Focus and Lightning Overload.
const elementalTalents = "5500331023001351-05500001-053"

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

// A Dwarf's beast bonus reaches the shaman's own Fire Nova but not a fire totem's attack, which deals
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
	beastBonus := func(spell *core.Spell) float64 {
		return spell.AttackerDamageMultiplier(beastTable, false) / spell.AttackerDamageMultiplier(other, false)
	}

	if bonus := beastBonus(sham.GetSpell(core.ActionID{SpellID: sham.FireNova.ActionID.SpellID, Tag: 1})); bonus <= 1 {
		t.Errorf("Fire Nova's hits deal %v times the damage to a beast; want the racial's bonus above 1", bonus)
	}
	for _, totem := range []*core.Spell{sham.MagmaTotem, sham.SearingTotem} {
		if bonus := beastBonus(sham.GetSpell(core.ActionID{SpellID: totem.ActionID.SpellID, Tag: 1})); bonus != 1 {
			t.Errorf("%s's attack deals %v times the damage to a beast; want 1", totem.ActionID, bonus)
		}
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

// Maelstrom Weapon procs off main-hand swings and Windfury Weapon's strike, never off Windfury
// Totem's strike.
func TestMaelstromWeaponSkipsTheWindfuryTotemStrike(t *testing.T) {
	player := windfuryShaman(&proto.APLRotation{})
	player.TalentsString = enhancementTalents
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{WindfuryTotem: true}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()

	sham := sim.Raid.Parties[0].Players[0].(shaman.ShamanAgent).GetShaman()
	maelstrom := sham.GetAuraByID(core.ActionID{SpellID: spelldata.MustFind(408505).ID})
	trigger := sham.GetAura("Maelstrom Weapon - Trigger")
	strike := sham.GetSpell(windfuryWeaponStrike)
	totemStrike := sham.GetSpell(windfuryTotemStrike)
	if maelstrom == nil || trigger == nil || strike == nil || totemStrike == nil {
		t.Fatalf("Maelstrom Weapon %v, its trigger %v, Windfury Weapon's strike %v, Windfury Totem's strike %v; want all registered",
			maelstrom, trigger, strike, totemStrike)
	}

	sham.AutoAttacks.CancelAutoSwing(sim)

	// The stack lands a spell batch window after its hit; a no-op action just past the window stops
	// each step there.
	procs := func(spell *core.Spell) int {
		gains := 0
		for range 300 {
			maelstrom.Deactivate(sim)
			trigger.OnSpellHitDealt(trigger, sim, spell, &core.SpellResult{Target: sham.CurrentTarget, Outcome: core.OutcomeHit, Damage: 100})
			for at := sim.CurrentTime + core.SpellBatchWindow; sim.CurrentTime <= at; {
				sim.AddPendingAction(&core.PendingAction{NextActionAt: at + 1, OnAction: func(*core.Simulation) {}})
				sim.Step()
			}
			if maelstrom.IsActive() {
				gains++
			}
		}
		return gains
	}
	if got := procs(totemStrike); got != 0 {
		t.Errorf("Windfury Totem's strike procced Maelstrom Weapon %d times in 300 hits, want none", got)
	}
	if procs(sham.AutoAttacks.MHAuto()) == 0 || procs(strike) == 0 {
		t.Error("a main-hand swing or Windfury Weapon's strike never procced Maelstrom Weapon in 300 hits, so the check above proves nothing")
	}
}

// A Lightning Overload's launch rolls Elemental Focus once, a Chain Lightning overload's included
// however many targets it hits. The Chain Lightning that launched the overload keeps the charge the
// overload procced as it completes, and an Earth Shock in that instant spends it. A Lightning Bolt
// overload launches after its bolt's cast has completed, so a Lightning Bolt of the same rank in that
// instant spends the charge.
func TestLightningOverloadRollsElementalFocusOnce(t *testing.T) {
	encounter := core.MakeSingleTargetEncounter(0)
	encounter.Targets = append(encounter.Targets, core.FreshDefaultTargetConfig(), core.FreshDefaultTargetConfig())
	player := &proto.Player{
		Name:          "Elemental",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassShaman,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: elementalTalents,
		Spec:          &proto.Player_ElementalShaman{ElementalShaman: &proto.ElementalShaman{Options: &proto.ElementalShaman_Options{ClassOptions: &proto.ShamanOptions{}}}},
		Rotation:      &proto.APLRotation{},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()

	sham := sim.Raid.Parties[0].Players[0].(shaman.ShamanAgent).GetShaman()
	target := sham.CurrentTarget
	const boltRank, chainRank = 15208, 10605
	clearcasting := sham.GetAuraByID(core.ActionID{SpellID: 16246})
	bolt := sham.GetSpell(core.ActionID{SpellID: boltRank})
	chainLightning := sham.GetSpell(core.ActionID{SpellID: chainRank})
	boltOverload := sham.GetSpell(core.ActionID{SpellID: boltRank, Tag: shaman.CastTagLightningOverload})
	chainOverload := sham.GetSpell(core.ActionID{SpellID: chainRank, Tag: shaman.CastTagLightningOverload})
	if clearcasting == nil || bolt == nil || chainLightning == nil || boltOverload == nil || chainOverload == nil {
		t.Fatalf("Clearcasting %v, Lightning Bolt %v, Chain Lightning %v, the overloads %v and %v; want all registered",
			clearcasting, bolt, chainLightning, boltOverload, chainOverload)
	}

	// A roll per hit would gain about three times as often off a Chain Lightning overload that hits
	// three targets.
	chance := spelldata.MustFind(16164).StatedChance()
	const launches = 2000
	tolerance := 4 * math.Sqrt(chance*(1-chance)/launches)
	for _, overload := range []*core.Spell{boltOverload, chainOverload} {
		gains := 0
		for range launches {
			clearcasting.Deactivate(sim)
			overload.Proc(sim, target)
			if clearcasting.IsActive() {
				gains++
			}
		}
		if rate := float64(gains) / launches; math.Abs(rate-chance) > tolerance {
			t.Errorf("%v gained Clearcasting on %v of its launches, want the row's %v within %v", overload.ActionID, rate, chance, tolerance)
		}
	}

	// The overload launches inside the cast's effects, before the cast completes.
	gainedAtLaunch := false
	launch := chainOverload.ApplyEffects
	chainOverload.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		launch(sim, target, spell)
		gainedAtLaunch = clearcasting.IsActive()
	}
	// Earth Shock's own roll can refresh the charge it spends, so one spend over the loop is enough.
	earthShock := sham.EarthShock[len(sham.EarthShock)-1]
	kept, spent := 0, false
	for range launches {
		clearcasting.Deactivate(sim)
		gainedAtLaunch = false
		chainLightning.Proc(sim, target)
		if !gainedAtLaunch {
			continue
		}
		kept++
		if !clearcasting.IsActive() {
			t.Fatal("the Chain Lightning that launched an overload spent the charge the overload procced")
		}
		earthShock.Proc(sim, target)
		spent = spent || !clearcasting.IsActive()
	}
	if kept == 0 {
		t.Errorf("no Chain Lightning overload procced Clearcasting in %d casts, so the check above proves nothing", launches)
	}
	if !spent {
		t.Error("an Earth Shock in the instant a Chain Lightning overload procced Clearcasting never spent the charge")
	}

	// The bolt's own roll can refresh the charge too.
	boltSpent := false
	for range launches {
		clearcasting.Deactivate(sim)
		boltOverload.Proc(sim, target)
		if !clearcasting.IsActive() {
			continue
		}
		bolt.Proc(sim, target)
		boltSpent = boltSpent || !clearcasting.IsActive()
	}
	if !boltSpent {
		t.Error("a Lightning Bolt in the instant a Lightning Bolt overload procced Clearcasting never spent the charge")
	}
}

// An untalented Elemental shaman wearing the items, at the start of the encounter.
func shamanWearing(t *testing.T, encounter *proto.Encounter, worn map[proto.ItemSlot]int32) (*core.Simulation, *shaman.Shaman) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{Id: worn[proto.ItemSlot(i)]}
	}
	player := &proto.Player{
		Name:      "Wearer",
		Race:      proto.Race_RaceOrc,
		Class:     proto.Class_ClassShaman,
		Equipment: &proto.EquipmentSpec{Items: items},
		Spec:      &proto.Player_ElementalShaman{ElementalShaman: &proto.ElementalShaman{Options: &proto.ElementalShaman_Options{ClassOptions: &proto.ShamanOptions{}}}},
		Rotation:  &proto.APLRotation{},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(shaman.ShamanAgent).GetShaman()
}

// Gift of the Gathering Storm's 3-piece bonus multiplies the share of the damage each Chain Lightning
// jump keeps by its row's percent: the first target takes the same, the second that factor more and
// the third its square.
func TestGiftOfTheGatheringStormRaisesEachChainLightningJump(t *testing.T) {
	const bonusRow, chainRank = 26123, 10605
	multiplier := 1 + spelldata.MustFind(bonusRow).EffectN(1).Percent()
	if multiplier <= 1 {
		t.Fatalf("the bonus multiplies the jumps' share by %v, want above 1", multiplier)
	}
	worn := map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotBack:     21400,
		proto.ItemSlot_ItemSlotFinger1:  21399,
		proto.ItemSlot_ItemSlotMainHand: 21398,
	}

	// What each of three targets takes from the same casts, with the bonus up or taken down. Both
	// sims draw the same rolls, so the jumps' share alone sets them apart.
	dealt := func(bonusUp bool) [3]float64 {
		t.Helper()
		encounter := core.MakeSingleTargetEncounter(0)
		encounter.Targets = append(encounter.Targets, core.FreshDefaultTargetConfig(), core.FreshDefaultTargetConfig())
		sim, sham := shamanWearing(t, encounter, worn)
		bonus := sham.GetAuraByID(core.ActionID{SpellID: bonusRow})
		chainLightning := sham.GetSpell(core.ActionID{SpellID: chainRank})
		if bonus == nil || !bonus.IsActive() || chainLightning == nil {
			t.Fatalf("the set bonus %v, Chain Lightning %v; want both registered and the bonus up", bonus, chainLightning)
		}
		if !bonusUp {
			bonus.Deactivate(sim)
		}
		for range 50 {
			chainLightning.Proc(sim, sham.CurrentTarget)
		}
		var damage [3]float64
		for i, target := range sim.Encounter.ActiveTargetUnits {
			damage[i] = chainLightning.SpellMetrics[target.UnitIndex].TotalDamage
		}
		return damage
	}

	with, without := dealt(true), dealt(false)
	for i, want := range []float64{1, multiplier, multiplier * multiplier} {
		if without[i] <= 0 {
			t.Fatalf("target %d takes no damage without the bonus", i)
		}
		if got := with[i] / without[i]; math.Abs(got-want) > 1e-9 {
			t.Errorf("target %d takes %v times its damage with the bonus, want %v", i, got, want)
		}
	}
}

// Stormcaller's Garb's 3-piece bonus grants Stormcaller's Wrath off Lightning Bolt, Chain Lightning and
// the shocks, the spells its tooltip names, and never off Fire Nova, another harmful shaman spell the
// row's listener would hear on its own.
func TestStormcallersGarbProcsOnlyOffTheSpellsItsTooltipNames(t *testing.T) {
	const boltRank, chainRank, wrathRow = 15208, 10605, 26121
	sim, sham := shamanWearing(t, core.MakeSingleTargetEncounter(0), map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotHead:  21372,
		proto.ItemSlot_ItemSlotFeet:  21373,
		proto.ItemSlot_ItemSlotChest: 21374,
	})
	wrath := sham.GetAuraByID(core.ActionID{SpellID: wrathRow})
	if wrath == nil {
		t.Fatal("Stormcaller's Wrath is not registered with three pieces worn")
	}

	// A trigger's handler runs a spell batch window after the cast.
	const casts = 200
	grants := func(spell *core.Spell) bool {
		t.Helper()
		for range casts {
			wrath.Deactivate(sim)
			spell.Proc(sim, sham.CurrentTarget)
			for end := sim.CurrentTime + core.SpellBatchWindow; sim.CurrentTime <= end; {
				if sim.Step() {
					t.Fatal("the encounter ended")
				}
			}
			if wrath.IsActive() {
				return true
			}
		}
		return false
	}

	named := []*core.Spell{
		sham.GetSpell(core.ActionID{SpellID: boltRank}),
		sham.GetSpell(core.ActionID{SpellID: chainRank}),
		sham.EarthShock[len(sham.EarthShock)-1],
		sham.FlameShock[len(sham.FlameShock)-1],
		sham.FrostShock[len(sham.FrostShock)-1],
	}
	for _, spell := range named {
		if spell == nil {
			t.Fatal("a spell the tooltip names is not registered")
		}
		if !grants(spell) {
			t.Errorf("%v never granted Stormcaller's Wrath in %d casts", spell.ActionID, casts)
		}
	}
	if grants(sham.FireNova) {
		t.Error("a Fire Nova granted Stormcaller's Wrath")
	}
}

// The Earthshatterer's 8-piece bonus gives the shaman its row's mana per 5 sec while Lightning Shield is
// up, and none once it falls or with seven pieces worn.
func TestEarthshattererGivesManaWhileLightningShieldIsUp(t *testing.T) {
	want := spelldata.MustFind(28820).EffectN(1).BaseValue()
	if want <= 0 {
		t.Fatalf("the bonus states %v mana per 5 sec, want above 0", want)
	}
	eight := map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotHead:     22466,
		proto.ItemSlot_ItemSlotShoulder: 22467,
		proto.ItemSlot_ItemSlotChest:    22464,
		proto.ItemSlot_ItemSlotWrist:    22471,
		proto.ItemSlot_ItemSlotHands:    22469,
		proto.ItemSlot_ItemSlotWaist:    22470,
		proto.ItemSlot_ItemSlotLegs:     22465,
		proto.ItemSlot_ItemSlotFeet:     22468,
	}

	// The mana per 5 sec the shaman gains with the shield up, and still has once it falls.
	shieldMP5 := func(worn map[proto.ItemSlot]int32) (up, fallen float64) {
		t.Helper()
		sim, sham := shamanWearing(t, core.MakeSingleTargetEncounter(0), worn)
		before := sham.GetStat(stats.MP5)
		sham.LightningShield.SkipCastAndApplyEffects(sim, &sham.Unit)
		up = sham.GetStat(stats.MP5) - before
		sham.LightningShieldAura.Deactivate(sim)
		return up, sham.GetStat(stats.MP5) - before
	}

	if up, fallen := shieldMP5(eight); up != want || fallen != 0 {
		t.Errorf("with eight pieces, Lightning Shield gives %v mana per 5 sec and leaves %v when it falls, want %v and 0", up, fallen, want)
	}
	seven := maps.Clone(eight)
	delete(seven, proto.ItemSlot_ItemSlotFeet)
	if up, _ := shieldMP5(seven); up != 0 {
		t.Errorf("with seven pieces, Lightning Shield gives %v mana per 5 sec, want 0", up)
	}
}

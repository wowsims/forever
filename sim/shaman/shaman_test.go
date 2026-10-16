package shaman_test

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/shaman"
	"github.com/wowsims/forever/sim/shaman/enhancement"
)

func init() {
	enhancement.RegisterEnhancementShaman()
	common.RegisterAllEffects()
}

// The Enhancement suite's talents, which take Elemental Devastation.
const enhancementTalents = "053503102-053030031005112251"

// Fire Nova's hits are a spell the cast triggers, and the row lacks Not a Proc, so Elemental
// Devastation, which cannot proc from procs, never hears their crits. It hears an Earth Shock's.
func TestFireNovaHitsAreProcs(t *testing.T) {
	player := &proto.Player{
		Name:          "Fire Nova",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassShaman,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: enhancementTalents,
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

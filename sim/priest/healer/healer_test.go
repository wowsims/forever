package healer

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	priestpkg "github.com/wowsims/forever/sim/priest"
)

func init() {
	RegisterHealerPriest()
	common.RegisterAllEffects()
}

// Stats-only suite: this spec is a gear planner, it has no healing rotation.
// Pins the final character stats for each gear preset so the passives stay covered. The empty APL
// rotation and the fake prepull (no SkipRotation) make it exercise a full environment reset, the
// path the UI's stats request takes.
func TestHealerPriest(t *testing.T) {
	var generators []core.TestGenerator
	for _, gearSet := range []string{"preraid"} {
		player := core.WithSpec(
			&proto.Player{
				Class:         proto.Class_ClassPriest,
				Race:          proto.Race_RaceDwarf,
				Equipment:     core.GetGearSet("../../../ui/specs/priest/healer/gear_sets", gearSet).GearSet,
				Consumables:   FullConsumes,
				Buffs:         core.FullIndividualBuffs,
				TalentsString: StandardTalents,
				Profession1:   proto.Profession_Tailoring,
				Profession2:   proto.Profession_Enchanting,
				Rotation:      &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
			},
			PlayerOptions,
		)
		generators = append(generators, &core.SingleCharacterStatsTestGenerator{
			Name: gearSet,
			Request: &proto.ComputeStatsRequest{
				Raid: core.SinglePlayerRaidProto(player, core.FullPartyBuffs, core.FullRaidBuffs, core.FullDebuffs),
			},
		})
	}
	core.RunTestSuite(t, t.Name(), generators)
}

func newHealingMetricsTestSim(t *testing.T) (*core.Simulation, *HealerPriest) {
	t.Helper()
	player := &proto.Player{
		Name:          "Healing Metrics",
		Race:          proto.Race_RaceTroll,
		Class:         proto.Class_ClassPriest,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: "005025021330011531-335001030301",
		Rotation:      &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
		Spec: &proto.Player_HealerPriest{
			HealerPriest: &proto.HealerPriest{
				Options: &proto.HealerPriest_Options{ClassOptions: &proto.PriestOptions{}},
			},
		},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(10),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()

	return sim, sim.Raid.Parties[0].Players[0].(*HealerPriest)
}

func TestDivineAegisMetricsOnlyRecordDirectCriticalHeals(t *testing.T) {
	sim, healer := newHealingMetricsTestSim(t)
	priest := healer.GetPriest()
	if priest.Talents.DivineAegis == 0 {
		t.Fatal("test talents do not enable Divine Aegis")
	}
	target := &priest.Unit
	spellForMask := func(mask int64) *core.Spell {
		for _, spell := range priest.Spellbook {
			if spell.Matches(mask) {
				return spell
			}
		}
		return nil
	}
	const divineAegisTriggeredID = 431624
	bindingHeal := spellForMask(priestpkg.PriestSpellBindingHeal)
	if bindingHeal == nil {
		t.Fatal("Binding Heal was not registered")
	}
	divineAegis := priest.GetSpell(core.ActionID{SpellID: divineAegisTriggeredID, Tag: bindingHeal.ActionID.SpellID})
	if divineAegis == nil {
		t.Fatal("Divine Aegis metrics were not registered for Binding Heal")
	}

	priest.OnHealDealt(sim, bindingHeal, &core.SpellResult{Target: target, Outcome: core.OutcomeCrit, Damage: 100})
	for sim.CurrentTime <= core.SpellBatchWindow && !sim.Step() {
	}
	if got := divineAegis.SpellMetrics[target.UnitIndex].TotalShielding; got <= 0 {
		t.Fatalf("Divine Aegis recorded %v shielding after a direct critical heal, want a positive absorb", got)
	}
	if got := divineAegis.SpellMetrics[target.UnitIndex].Hits; got != 1 {
		t.Fatalf("Divine Aegis recorded %d hits after one critical heal, want 1", got)
	}
	fillShield := func() {
		priest.OnHealDealt(sim, bindingHeal, &core.SpellResult{Target: target, Outcome: core.OutcomeCrit, Damage: 1e9})
		until := sim.CurrentTime + core.SpellBatchWindow
		for sim.CurrentTime <= until && !sim.Step() {
		}
	}
	fillShield()
	capacity := divineAegis.SpellMetrics[target.UnitIndex].TotalShielding
	fillShield()
	if got := divineAegis.SpellMetrics[target.UnitIndex].TotalShielding; got != capacity {
		t.Errorf("refreshing a full Divine Aegis shield changed credited capacity from %v to %v", capacity, got)
	}

	periodicSim, periodicHealer := newHealingMetricsTestSim(t)
	periodicPriest := periodicHealer.GetPriest()
	renew := func() *core.Spell {
		for _, spell := range periodicPriest.Spellbook {
			if spell.Matches(priestpkg.PriestSpellRenew) {
				return spell
			}
		}
		return nil
	}()
	if renew == nil {
		t.Fatal("Renew was not registered")
	}
	periodicTarget := &periodicPriest.Unit
	periodicAegis := periodicPriest.GetSpell(core.ActionID{SpellID: divineAegisTriggeredID, Tag: renew.ActionID.SpellID})
	if periodicAegis == nil {
		t.Fatal("Divine Aegis metrics were not registered for Renew")
	}
	periodicPriest.OnPeriodicHealDealt(periodicSim, renew, &core.SpellResult{Target: periodicTarget, Outcome: core.OutcomeCrit, Damage: 100})
	for periodicSim.CurrentTime <= core.SpellBatchWindow && !periodicSim.Step() {
	}
	if got := periodicAegis.SpellMetrics[periodicTarget.UnitIndex].TotalShielding; got != 0 {
		t.Errorf("Divine Aegis recorded %v shielding after a periodic critical heal, want 0", got)
	}
}

func TestPowerWordShieldRecordsAppliedAbsorbOnce(t *testing.T) {
	sim, healer := newHealingMetricsTestSim(t)
	priest := healer.GetPriest()
	target := &priest.Unit
	var spell *core.Spell
	for _, candidate := range priest.Spellbook {
		if candidate.Matches(priestpkg.PriestSpellPowerWordShield) {
			spell = candidate
		}
	}
	if spell == nil {
		t.Fatal("Power Word: Shield was not registered")
	}
	if !spell.Cast(sim, target) {
		t.Fatal("Power Word: Shield could not be cast on the priest")
	}
	metrics := spell.SpellMetrics[target.UnitIndex]
	if metrics.TotalShielding <= 0 {
		t.Fatalf("Power Word: Shield recorded %v shielding, want a positive applied absorb", metrics.TotalShielding)
	}
	if metrics.Hits != 1 {
		t.Errorf("Power Word: Shield recorded %d hits for one applied shield, want 1", metrics.Hits)
	}

	spell.SkipCastAndApplyEffects(sim, target)
	if got := spell.SpellMetrics[target.UnitIndex].TotalShielding; got != metrics.TotalShielding {
		t.Errorf("replacing an unused shield changed credited capacity from %v to %v", metrics.TotalShielding, got)
	}

	var lowerRank *core.Spell
	for _, candidate := range priest.Spellbook {
		if candidate.Matches(priestpkg.PriestSpellPowerWordShield) && candidate.Rank < spell.Rank {
			lowerRank = candidate
			break
		}
	}
	if lowerRank == nil {
		t.Fatal("no lower rank of Power Word: Shield was registered")
	}
	lowerRank.SkipCastAndApplyEffects(sim, target)
	if got := lowerRank.SpellMetrics[target.UnitIndex].TotalShielding; got != 0 {
		t.Errorf("replacing an unused shield with a weaker one credited %v capacity, want 0", got)
	}
}

// The Holy preset from ui/specs/priest/healer/presets.ts.
var StandardTalents = "00520101133-33505003030121531"

var FullConsumes = &proto.ConsumesSpec{
	FoodId:        250066, // Soaring Pamplona
	ResourcePotId: 13444,  // Major Mana Potion
}

var PlayerOptions = &proto.Player_HealerPriest{
	HealerPriest: &proto.HealerPriest{
		Options: &proto.HealerPriest_Options{
			ClassOptions: &proto.PriestOptions{},
		},
	},
}

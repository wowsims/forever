package core

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

func init() {
	RegisterAgentFactory(
		proto.Player_ElementalShaman{},
		proto.Spec_SpecElementalShaman,
		NewFakeElementalShaman,
		func(player *proto.Player, spec any) {
			playerSpec, ok := spec.(*proto.Player_ElementalShaman)
			if !ok {
				panic("Invalid spec value for Elemental Shaman!")
			}
			player.Spec = playerSpec
		},
	)
}

type FakeAgent struct {
	Spell *Spell
	Dot   *Dot
	Character
	Init func()
}

func (fa *FakeAgent) GetCharacter() *Character {
	return &fa.Character
}

func (fa *FakeAgent) Initialize() {
	if fa.Init != nil {
		fa.Init()
	}
}

func (fa *FakeAgent) ApplyTalents()                  {}
func (fa *FakeAgent) Reset(_ *Simulation)            {}
func (fa *FakeAgent) OnGCDReady(_ *Simulation)       {}
func (fa *FakeAgent) OnEncounterStart(_ *Simulation) {}

func NewFakeElementalShaman(char *Character, _ *proto.Player, _ *proto.Raid) Agent {
	fa := &FakeAgent{
		Character: *char,
	}

	fa.Init = func() {
		fa.Spell = fa.RegisterSpell(SpellConfig{
			ActionID:    ActionID{SpellID: 42},
			SpellSchool: SpellSchoolShadow,
			ProcMask:    ProcMaskSpellDamage,
			Flags:       SpellFlagIgnoreResists,
			Cast:        CastConfig{},

			BonusCritPercent: 3,
			DamageMultiplier: 1.5,
			ThreatMultiplier: 1,

			Dot: DotConfig{
				Aura: Aura{
					Label: "fakedot",
				},
				NumberOfTicks:       6,
				TickLength:          time.Second * 3,
				AffectedByCastSpeed: false,
				BonusCoefficient:    1,

				OnSnapshot: func(sim *Simulation, target *Unit, dot *Dot) {
					dot.Snapshot(target, 100)
				},
				OnTick: func(sim *Simulation, target *Unit, dot *Dot) {
					dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
				},
			},

			ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
				result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
				if result.Landed() {
					spell.Dot(target).Apply(sim)
				}
				spell.DealOutcome(sim, result)
			},
		})
		fa.Dot = fa.Spell.CurDot()
	}

	return fa
}

func SetupFakeSim() *Simulation {
	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{
			RandomSeed: 100,
		},
		Raid: &proto.Raid{
			Parties: []*proto.Party{
				{
					Players: []*proto.Player{
						{
							Name:      "Caster",
							Class:     proto.Class_ClassShaman,
							Buffs:     &proto.IndividualBuffs{},
							Spec:      &proto.Player_ElementalShaman{},
							Equipment: &proto.EquipmentSpec{},
						},
					},
					Buffs: &proto.PartyBuffs{},
				},
			},
		},
		Encounter: &proto.Encounter{
			Targets: []*proto.Target{
				{Name: "target", Level: 60, MobType: proto.MobType_MobTypeDemon},
			},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()

	return sim
}

func expectDotTickDamage(t *testing.T, sim *Simulation, dot *Dot, expectedDamage float64) {
	damageBefore := dot.Spell.SpellMetrics[0].TotalDamage
	dot.TickOnce(sim)
	damageAfter := dot.Spell.SpellMetrics[0].TotalDamage
	delta := damageAfter - damageBefore

	if !WithinToleranceFloat64(expectedDamage, delta, 0.01) {
		t.Fatalf("Incorrect tick damage applied: Expected: %0.3f, Actual: %0.3f", expectedDamage, delta)
	}
}

func TestBleedDoesNotLandOnAnImmuneTarget(t *testing.T) {
	sim := SetupFakeSim()
	fa := sim.Raid.Parties[0].Players[0].(*FakeAgent)
	fa.Dot.Unit.BleedImmune = true

	fa.Dot.Apply(sim)
	if !fa.Dot.IsActive() {
		t.Fatal("a Dot that is not a bleed did not land on a bleed-immune target")
	}
	fa.Dot.Deactivate(sim)

	fa.Spell.Flags |= SpellFlagBleed
	fa.Dot.Apply(sim)
	if fa.Dot.IsActive() {
		t.Fatal("a bleed landed on a bleed-immune target")
	}

	fa.Dot.Unit.BleedImmune = false
	fa.Dot.Apply(sim)
	if !fa.Dot.IsActive() {
		t.Fatal("a bleed did not land on a target that takes bleeds")
	}
}

func TestBleedHitMissesAnImmuneTarget(t *testing.T) {
	sim := SetupFakeSim()
	fa := sim.Raid.Parties[0].Players[0].(*FakeAgent)
	target := sim.Encounter.AllTargetUnits[0]
	target.BleedImmune = true
	always := func(_ *Simulation, result *SpellResult, _ *AttackTable) { result.Outcome = OutcomeHit }

	if result := fa.Spell.CalcDamage(sim, target, 100, always); !result.Landed() {
		t.Fatal("a spell with no bleed hit missed a bleed-immune target")
	}
	fa.Spell.Flags |= SpellFlagBleed
	if result := fa.Spell.CalcDamage(sim, target, 100, always); result.Landed() || result.Damage != 0 {
		t.Fatalf("a bleed hit landed for %v on a bleed-immune target", result.Damage)
	}
	if result := fa.Spell.CalcOutcome(sim, target, always); result.Landed() {
		t.Fatal("a bleed outcome landed on a bleed-immune target")
	}
	target.BleedImmune = false
	if result := fa.Spell.CalcDamage(sim, target, 100, always); !result.Landed() {
		t.Fatal("a bleed hit missed a target that takes bleeds")
	}
}

func TestTargetReadsBleedImmunity(t *testing.T) {
	if !NewTarget(&proto.Target{BleedImmune: true}, 0).BleedImmune {
		t.Fatal("the target setting did not reach the unit")
	}
	if NewTarget(&proto.Target{}, 0).BleedImmune {
		t.Fatal("a target is bleed-immune by default")
	}
}

func TestDotSnapshot(t *testing.T) {
	sim := SetupFakeSim()
	fa := sim.Raid.Parties[0].Players[0].(*FakeAgent)

	fa.Dot.Apply(sim)
	expectDotTickDamage(t, sim, fa.Dot, 150) // (100) * 1.5
}

func TestDotSnapshotSpellDamage(t *testing.T) {
	sim := SetupFakeSim()
	fa := sim.Raid.Parties[0].Players[0].(*FakeAgent)

	fa.Dot.Apply(sim)
	expectDotTickDamage(t, sim, fa.Dot, 150) // (100) * 1.5

	// Spell power shouldn't get applied because dot was already snapshot.
	fa.GetCharacter().AddStatDynamic(sim, stats.SpellDamage, 100)
	expectDotTickDamage(t, sim, fa.Dot, 150) // (100) * 1.5

	fa.Dot.Deactivate(sim)
	fa.Dot.Apply(sim)
	expectDotTickDamage(t, sim, fa.Dot, 300) // (100 + 100) * 1.5
}

func TestDotSnapshotSpellMultiplier(t *testing.T) {
	sim := SetupFakeSim()
	fa := sim.Raid.Parties[0].Players[0].(*FakeAgent)
	spell := fa.GetCharacter().Spellbook[0]
	spell.DamageMultiplier *= 2

	fa.Dot.Apply(sim)
	expectDotTickDamage(t, sim, fa.Dot, 300) // (100) * 1.5 * 2
}

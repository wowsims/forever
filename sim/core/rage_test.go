package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

func init() {
	RegisterAgentFactory(
		proto.Player_DpsWarrior{},
		proto.Spec_SpecDpsWarrior,
		NewFakeRageWarrior,
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_DpsWarrior)
			if !ok {
				panic("Invalid spec value for Dps Warrior!")
			}
			player.Spec = playerSpec
		},
	)
}

const (
	fakeMHSwingSpeed = 2.6
	fakeOHSwingSpeed = 1.8
)

type FakeRageWarrior struct {
	Character
}

func (fw *FakeRageWarrior) GetCharacter() *Character { return &fw.Character }

func (fw *FakeRageWarrior) Initialize() {
	fw.registerFakeThreatSpell()
	fw.registerFakeGCDSpell()
	fw.registerFakeCategorySpells()
	fw.registerFakeRend()
	fw.registerFakeHeal()
}
func (fw *FakeRageWarrior) ApplyTalents()                  {}
func (fw *FakeRageWarrior) Reset(_ *Simulation)            {}
func (fw *FakeRageWarrior) OnGCDReady(_ *Simulation)       {}
func (fw *FakeRageWarrior) OnEncounterStart(_ *Simulation) {}

func NewFakeRageWarrior(char *Character, _ *proto.Player, _ *proto.Raid) Agent {
	fw := &FakeRageWarrior{
		Character: *char,
	}

	fw.EnableRageBar(RageBarOptions{
		MaxRage:            100,
		BaseRageMultiplier: 1,
		StartingRage:       0,
	})

	fw.EnableAutoAttacks(fw, AutoAttackOptions{
		MainHand:       Weapon{SwingSpeed: fakeMHSwingSpeed},
		OffHand:        Weapon{SwingSpeed: fakeOHSwingSpeed},
		AutoSwingMelee: true,
	})

	return fw
}

func SetupFakeRageSim() *Simulation {
	return setupFakeRageSimFrom(fakeRageSimRequest())
}

func setupFakeRageSimFrom(request *proto.RaidSimRequest) *Simulation {
	sim := NewSim(request, simsignals.CreateSignals())
	sim.Reset()

	return sim
}

func fakeRageSimRequest() *proto.RaidSimRequest {
	return &proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{
			RandomSeed: 100,
		},
		Raid: &proto.Raid{
			Parties: []*proto.Party{
				{
					Players: []*proto.Player{
						{
							Name:      "Warrior",
							Class:     proto.Class_ClassWarrior,
							Buffs:     &proto.IndividualBuffs{},
							Spec:      &proto.Player_DpsWarrior{},
							Equipment: &proto.EquipmentSpec{},
							// AddRage() pokes the rotation, so it needs to be non-nil.
							Rotation: &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
						},
					},
					Buffs: &proto.PartyBuffs{},
				},
			},
		},
		Encounter: &proto.Encounter{
			Targets: []*proto.Target{
				{Name: "target", Level: 63, MobType: proto.MobType_MobTypeHumanoid, SwingSpeed: 2},
			},
			Duration: 180,
		},
	}
}

// Feeds a hand-crafted auto attack result to the Rage bar and returns the Rage gained from it.
func rageFromAutoAttack(sim *Simulation, fw *FakeRageWarrior, spell *Spell, outcome HitOutcome, preOutcomeDamage float64) float64 {
	result := &SpellResult{
		Target:  sim.Encounter.ActiveTargetUnits[0],
		Outcome: outcome,
	}
	if outcome.Matches(OutcomeLanded) {
		result.Damage = preOutcomeDamage
	}

	rageBefore := fw.CurrentRage()
	fw.Unit.OnSpellHitDealt(sim, spell, result)

	return fw.CurrentRage() - rageBefore
}

func TestAutoAttackRageGeneration(t *testing.T) {
	// A one-hand MH swing at 2.6 speed: 2.6 * 3.46 = 8.996, whatever it dealt; a crit adds the same again, 17.992.
	const swingDamage = 500.0

	tests := []struct {
		name     string
		outcome  HitOutcome
		wantRage float64
	}{
		{
			name:     "hit",
			outcome:  OutcomeHit,
			wantRage: 8.996,
		},
		{
			name:     "crit",
			outcome:  OutcomeCrit,
			wantRage: 17.992,
		},
		{
			name:     "glance",
			outcome:  OutcomeGlance,
			wantRage: 8.996,
		},
		{
			name:     "dodge",
			outcome:  OutcomeDodge,
			wantRage: 0,
		},
		{
			name:     "parry",
			outcome:  OutcomeParry,
			wantRage: 0,
		},
		{
			name:     "miss",
			outcome:  OutcomeMiss,
			wantRage: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sim := SetupFakeRageSim()
			fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)

			gained := rageFromAutoAttack(sim, fw, fw.AutoAttacks.MHAuto(), test.outcome, swingDamage)

			if !WithinToleranceFloat64(test.wantRage, gained, 0.01) {
				t.Fatalf("Incorrect Rage generated on %s: Expected: %0.3f, Actual: %0.3f", test.name, test.wantRage, gained)
			}
		})
	}
}

func TestTwoHandAutoAttackRageGeneration(t *testing.T) {
	// A two-hand MH swing at 2.6 speed: 2.6 * 4.5 = 11.7. A crit adds the swing's rage before the two-hand
	// multiplier, 2.6 * 3.46 = 8.996, for 20.696.
	const swingDamage = 500.0

	sim := SetupFakeRageSim()
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	fw.AutoAttacks.MH().NormalizedSwingSpeed = TwoHandNormalizedSwingSpeed

	hitRage := rageFromAutoAttack(sim, fw, fw.AutoAttacks.MHAuto(), OutcomeHit, swingDamage)
	if !WithinToleranceFloat64(11.7, hitRage, 0.001) {
		t.Fatalf("Incorrect Rage generated on 2H hit: Expected: %0.3f, Actual: %0.3f", 11.7, hitRage)
	}

	critRage := rageFromAutoAttack(sim, fw, fw.AutoAttacks.MHAuto(), OutcomeCrit, swingDamage)
	if !WithinToleranceFloat64(20.696, critRage, 0.001) {
		t.Fatalf("Incorrect Rage generated on 2H crit: Expected: %0.3f, Actual: %0.3f", 20.696, critRage)
	}
}

func TestOffHandAutoAttackRageGeneration(t *testing.T) {
	// An OH swing at 1.8 speed: 1.8 * 3.46 * 0.5 = 3.114.
	const swingDamage = 500.0

	sim := SetupFakeRageSim()
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	ohAuto := fw.AutoAttacks.OHAuto()

	hitRage := rageFromAutoAttack(sim, fw, ohAuto, OutcomeHit, swingDamage)
	if !WithinToleranceFloat64(3.114, hitRage, 0.01) {
		t.Fatalf("Incorrect Rage generated on OH hit: Expected: %0.3f, Actual: %0.3f", 3.114, hitRage)
	}

	dodgeRage := rageFromAutoAttack(sim, fw, ohAuto, OutcomeDodge, swingDamage)
	if dodgeRage != 0 {
		t.Fatalf("Dodged OH swing generated %0.3f Rage, expected none", dodgeRage)
	}
}

func TestDamageTakenRageGeneration(t *testing.T) {
	// 1000 damage before armor from a level 63 attacker, which assumes 40% armor: 2.5 * 0.6 * 1000 / 243.64
	// = 6.157, whatever the school, and a periodic tick alike. From a level 30 attacker (20%): 2.5 * 0.8 * 1000 /
	// 109.23 = 18.310; level 44 (30%): 2.5 * 0.7 * 1000 / 163.83 = 10.682; level 45 sits halfway between the 44 and
	// 46 rows (33%): 2.5 * 0.67 * 1000 / 167.87 = 9.978.
	const preArmorDamage = 1000.0

	tests := []struct {
		name          string
		school        SpellSchool
		outcome       HitOutcome
		periodic      bool
		attackerLevel int32
		wantRage      float64
	}{
		{name: "physical hit", school: SpellSchoolPhysical, outcome: OutcomeHit, wantRage: 6.157},
		{name: "fire hit", school: SpellSchoolFire, outcome: OutcomeHit, wantRage: 6.157},
		{name: "dodge", school: SpellSchoolPhysical, outcome: OutcomeDodge, wantRage: 0},
		{name: "nature tick", school: SpellSchoolNature, outcome: OutcomeHit, periodic: true, wantRage: 6.157},
		{name: "physical tick", school: SpellSchoolPhysical, outcome: OutcomeHit, periodic: true, wantRage: 6.157},
		{name: "lower level attacker", school: SpellSchoolPhysical, outcome: OutcomeHit, attackerLevel: 30, wantRage: 18.310},
		{name: "attacker on a row", school: SpellSchoolPhysical, outcome: OutcomeHit, attackerLevel: 44, wantRage: 10.682},
		{name: "attacker between rows", school: SpellSchoolPhysical, outcome: OutcomeHit, attackerLevel: 45, wantRage: 9.978},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sim := SetupFakeRageSim()
			fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
			attacker := sim.Encounter.ActiveTargetUnits[0]
			if test.attackerLevel != 0 {
				attacker.Level = test.attackerLevel
			}

			result := &SpellResult{
				Target:         &fw.Unit,
				Outcome:        test.outcome,
				PreArmorDamage: preArmorDamage,
			}

			spell := &Spell{Unit: attacker, SpellSchool: test.school}
			rageBefore := fw.CurrentRage()
			if test.periodic {
				fw.Unit.OnPeriodicDamageTaken(sim, spell, result)
			} else {
				fw.Unit.OnSpellHitTaken(sim, spell, result)
			}
			gained := fw.CurrentRage() - rageBefore

			if !WithinToleranceFloat64(test.wantRage, gained, 0.001) {
				t.Fatalf("Incorrect Rage from %s: Expected: %0.3f, Actual: %0.3f", test.name, test.wantRage, gained)
			}
		})
	}
}

func TestPreArmorDamage(t *testing.T) {
	// A 1000 hit through 5755 armor against a level 63 attacker (half), 50 flat physical damage taken,
	// a 10% damage taken reduction, a crit, a 100 block and a 300 absorb:
	// (1000 * 0.5 + 50) * 0.9 * 2 - 100 - 300 = 590 taken. The damage rage reads keeps the crit and the
	// 10% reduction and drops the rest: 1000 * 0.9 * 2 = 1800.
	sim := SetupFakeRageSim()
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	attacker := sim.Encounter.ActiveTargetUnits[0]
	spell := attacker.AutoAttacks.MHAuto()

	fw.stats[stats.Armor] = 5755
	fw.stats[stats.BlockValue] = 100
	fw.PseudoStats.CanBlock = true
	fw.PseudoStats.BonusPhysicalDamageTaken = 50
	fw.PseudoStats.DamageTakenMultiplier = 0.9
	fw.DynamicDamageTakenModifiers = append(fw.DynamicDamageTakenModifiers, func(_ *Simulation, _ *Spell, result *SpellResult, _ bool) {
		result.Damage -= 300
	})

	critThenBlock := func(sim *Simulation, result *SpellResult, attackTable *AttackTable) {
		result.Outcome = OutcomeCrit
		result.Damage *= 2
		chance := 0.0
		result.applyEnemyAttackTableBlock(sim, spell, attackTable, -1, &chance)
	}

	rageBefore := fw.CurrentRage()
	result := spell.calcDamageInternal(sim, &fw.Unit, 1000, 1, false, critThenBlock)
	spell.DealDamage(sim, result)
	gained := fw.CurrentRage() - rageBefore

	if !WithinToleranceFloat64(590, result.Damage, 0.001) {
		t.Fatalf("Damage taken: Expected: %0.3f, Actual: %0.3f", 590.0, result.Damage)
	}
	if !WithinToleranceFloat64(1800, result.PreArmorDamage, 0.001) {
		t.Fatalf("Damage before armor: Expected: %0.3f, Actual: %0.3f", 1800.0, result.PreArmorDamage)
	}
	// 2.5 * 0.6 * 1800 / 243.64 = 11.082.
	if !WithinToleranceFloat64(11.082, gained, 0.001) {
		t.Fatalf("Rage from the hit: Expected: %0.3f, Actual: %0.3f", 11.082, gained)
	}
}

func TestOffHandRageMultiplier(t *testing.T) {
	const swingDamage = 500.0

	sim := SetupFakeRageSim()
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)

	unmodified := rageFromAutoAttack(sim, fw, fw.AutoAttacks.OHAuto(), OutcomeHit, swingDamage)
	if !WithinToleranceFloat64(3.114, unmodified, 0.01) {
		t.Fatalf("Incorrect Rage generated on OH hit: Expected: %0.3f, Actual: %0.3f", 3.114, unmodified)
	}

	fw.SetOffHandRageMultiplier(1.5)

	modified := rageFromAutoAttack(sim, fw, fw.AutoAttacks.OHAuto(), OutcomeHit, swingDamage)
	if !WithinToleranceFloat64(4.671, modified, 0.01) {
		t.Fatalf("Incorrect Rage generated on multiplied OH hit: Expected: %0.3f, Actual: %0.3f", 4.671, modified)
	}

	mhRage := rageFromAutoAttack(sim, fw, fw.AutoAttacks.MHAuto(), OutcomeHit, swingDamage)
	if !WithinToleranceFloat64(8.996, mhRage, 0.01) {
		t.Fatalf("Incorrect Rage generated on MH hit: Expected: %0.3f, Actual: %0.3f", 8.996, mhRage)
	}
}

func newPowerBarCharacter() *Character {
	character := NewCharacter(&Party{}, 0, &proto.Player{
		Name:      "Power Bar Tester",
		Class:     proto.Class_ClassWarrior,
		Race:      proto.Race_RaceHuman,
		Spec:      &proto.Player_ProtectionWarrior{ProtectionWarrior: &proto.ProtectionWarrior{}},
		Equipment: &proto.EquipmentSpec{},
	})
	return &character
}

func TestSetMaximumRage(t *testing.T) {
	sim := SetupFakeRageSim()
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	metrics := fw.NewRageMetrics(ActionID{SpellID: 1310236})

	fw.SetMaximumRage(sim, 130, metrics)
	if fw.MaximumRage() != 130 || fw.CurrentRage() != 0 {
		t.Fatalf("raised to 130: max %v with %v rage, want 130 with none granted", fw.MaximumRage(), fw.CurrentRage())
	}

	fw.AddRage(sim, 120, metrics)
	if fw.CurrentRage() != 120 {
		t.Fatalf("rage under the raised maximum: %v, want 120", fw.CurrentRage())
	}

	fw.SetMaximumRage(sim, 100, metrics)
	if fw.MaximumRage() != 100 || fw.CurrentRage() != 100 {
		t.Fatalf("lowered to 100: max %v with %v rage, want 100 with 100", fw.MaximumRage(), fw.CurrentRage())
	}
	if metrics.ActualGain != 100 || metrics.Events != 2 {
		t.Errorf("metrics after gaining 120 and losing 20: actual gain %v over %d events, want 100 over 2",
			metrics.ActualGain, metrics.Events)
	}
}

func TestSetMaximumRageClampsTheStartingRage(t *testing.T) {
	character := newPowerBarCharacter()
	character.EnableRageBar(RageBarOptions{MaxRage: 100, StartingRage: 120, BaseRageMultiplier: 1})

	steps := []struct {
		maxRage float64
		want    float64
	}{
		{100, 100},
		{130, 120},
		{110, 110},
	}
	for _, step := range steps {
		character.SetMaximumRage(nil, step.maxRage, nil)
		character.rageBar.reset(nil)
		if got := character.CurrentRage(); got != step.want {
			t.Errorf("starting rage 120 under a maximum of %v: %v, want %v", step.maxRage, got, step.want)
		}
	}
}

func TestSetMaximumEnergy(t *testing.T) {
	character := newPowerBarCharacter()
	character.EnableEnergyBar(EnergyBarOptions{MaxEnergy: 100, MaxComboPoints: 5, HasNoRegen: true})
	metrics := character.NewEnergyMetrics(ActionID{SpellID: 14983})
	sim := &Simulation{}

	character.energyBar.currentEnergy = 100
	character.SetMaximumEnergy(sim, 110, metrics)
	if character.MaximumEnergy() != 110 || character.CurrentEnergy() != 100 {
		t.Fatalf("raised to 110: max %v with %v energy, want 110 with 100", character.MaximumEnergy(), character.CurrentEnergy())
	}

	character.energyBar.currentEnergy = 110
	character.SetMaximumEnergy(sim, 100, metrics)
	if character.MaximumEnergy() != 100 || character.CurrentEnergy() != 100 {
		t.Fatalf("lowered to 100: max %v with %v energy, want 100 with 100", character.MaximumEnergy(), character.CurrentEnergy())
	}
	if metrics.ActualGain != -10 || metrics.Events != 1 {
		t.Errorf("metrics after losing 10: actual gain %v over %d events, want -10 over 1", metrics.ActualGain, metrics.Events)
	}
}

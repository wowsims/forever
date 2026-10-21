package core

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

const (
	fakeDefensiveID int32 = 993001
	fakeOffensiveID int32 = 993002
	fakeHardcastID  int32 = 993003
	fakeExplosiveID int32 = 993004
	fakeInstantID   int32 = 993005
)

func init() {
	RegisterAgentFactory(
		proto.Player_Mage{},
		proto.Spec_SpecMage,
		func(char *Character, _ *proto.Player, _ *proto.Raid) Agent {
			return &fakeCooldownMage{Character: *char, casts: map[int32]int{}}
		},
		func(player *proto.Player, spec interface{}) {
			player.Spec = spec.(*proto.Player_Mage)
		},
	)
}

// A stand-in with one survival and one DPS major cooldown, each off the GCD on a 10 s cooldown, an
// explosive with a 1 s cast on a 20.8 s cooldown, a spell whose cast time is longer than the GCD and an
// instant spell on the GCD, which counts how often each is used. Thrown as a GCD starts, the explosive
// comes ready 0.8 s into a later one.
type fakeCooldownMage struct {
	Character
	casts map[int32]int

	// When set, the rotation is also evaluated this often, the way auto attacks evaluate it between GCDs.
	reactEvery time.Duration
}

func (m *fakeCooldownMage) GetCharacter() *Character { return &m.Character }
func (m *fakeCooldownMage) ApplyTalents()            {}
func (m *fakeCooldownMage) Reset(_ *Simulation)      {}
func (m *fakeCooldownMage) OnGCDReady(_ *Simulation) {}

func (m *fakeCooldownMage) OnEncounterStart(sim *Simulation) {
	if m.reactEvery > 0 {
		StartPeriodicAction(sim, PeriodicActionOptions{Period: m.reactEvery, OnAction: func(sim *Simulation) {
			m.ReactToEvent(sim, false, false)
		}})
	}
}

func (m *fakeCooldownMage) Initialize() {
	m.registerCooldown(fakeDefensiveID, CooldownTypeSurvival)
	m.registerCooldown(fakeOffensiveID, CooldownTypeDPS)
	m.AddMajorCooldown(MajorCooldown{
		Spell: m.RegisterSpell(SpellConfig{
			ActionID: ActionID{SpellID: fakeExplosiveID},
			Flags:    SpellFlagNoOnCastComplete,
			Cast: CastConfig{
				DefaultCast: Cast{CastTime: time.Second},
				CD:          Cooldown{Timer: m.NewTimer(), Duration: 20800 * time.Millisecond},
			},
			ApplyEffects: func(_ *Simulation, _ *Unit, _ *Spell) {
				m.casts[fakeExplosiveID]++
			},
		}),
		Type:     CooldownTypeDPS | CooldownTypeExplosive,
		Priority: CooldownPriorityLow,
	})
	m.registerSpell(fakeHardcastID, 2*time.Second)
	m.registerSpell(fakeInstantID, 0)
}

func (m *fakeCooldownMage) registerCooldown(id int32, cdType CooldownType) {
	spell := m.RegisterSpell(SpellConfig{
		ActionID: ActionID{SpellID: id},
		Flags:    SpellFlagNoOnCastComplete,
		Cast:     CastConfig{CD: Cooldown{Timer: m.NewTimer(), Duration: 10 * time.Second}},
		ApplyEffects: func(_ *Simulation, _ *Unit, _ *Spell) {
			m.casts[id]++
		},
	})
	m.AddMajorCooldown(MajorCooldown{Spell: spell, Type: cdType})
}

func (m *fakeCooldownMage) registerSpell(id int32, castTime time.Duration) {
	m.RegisterSpell(SpellConfig{
		ActionID: ActionID{SpellID: id},
		Flags:    SpellFlagAPL,
		Cast:     CastConfig{DefaultCast: Cast{CastTime: castTime, GCD: GCDDefault}},
		ApplyEffects: func(_ *Simulation, _ *Unit, _ *Spell) {
			m.casts[id]++
		},
	})
}

func autocastRotation() *proto.APLRotation {
	return &proto.APLRotation{Type: proto.APLRotation_TypeAPL, PriorityList: []*proto.APLListItem{{
		Action: &proto.APLAction{Action: &proto.APLAction_AutocastOtherCooldowns{AutocastOtherCooldowns: &proto.APLActionAutocastOtherCooldowns{}}},
	}}}
}

func castRotation(id int32) *proto.APLRotation {
	return &proto.APLRotation{Type: proto.APLRotation_TypeAPL, PriorityList: []*proto.APLListItem{{
		Action: &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: ActionID{SpellID: id}.ToProto()}}},
	}}}
}

// One rotation that runs the given ones' actions in order.
func joinRotations(rotations ...*proto.APLRotation) *proto.APLRotation {
	joined := &proto.APLRotation{Type: proto.APLRotation_TypeAPL}
	for _, rotation := range rotations {
		joined.PriorityList = append(joined.PriorityList, rotation.PriorityList...)
	}
	return joined
}

const cooldownMagesDuration = 55 * time.Second

// Runs one 55 s iteration with a stand-in per rotation, the raid listing the given tanks, and returns
// the stand-ins.
func runCooldownMages(t *testing.T, tanks []*proto.UnitReference, rotations ...*proto.APLRotation) []*fakeCooldownMage {
	t.Helper()
	sim, mages := newCooldownMages(tanks, rotations...)
	sim.runOnce()
	return mages
}

// A 55 s sim with a stand-in per rotation and the raid listing the given tanks, not yet run, and its
// stand-ins.
func newCooldownMages(tanks []*proto.UnitReference, rotations ...*proto.APLRotation) (*Simulation, []*fakeCooldownMage) {
	var players []*proto.Player
	for _, rotation := range rotations {
		players = append(players, &proto.Player{
			Name:      "Mage",
			Class:     proto.Class_ClassMage,
			Race:      proto.Race_RaceHuman,
			Buffs:     &proto.IndividualBuffs{},
			Spec:      &proto.Player_Mage{},
			Equipment: &proto.EquipmentSpec{},
			Rotation:  rotation,
		})
	}
	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{
			Parties: []*proto.Party{{Players: players, Buffs: &proto.PartyBuffs{}}},
			Tanks:   tanks,
		},
		Encounter: &proto.Encounter{
			Targets:  []*proto.Target{{Name: "target", Level: 63, MobType: proto.MobType_MobTypeHumanoid}},
			Duration: cooldownMagesDuration.Seconds(),
		},
	}, simsignals.CreateSignals())

	var mages []*fakeCooldownMage
	for _, player := range sim.Raid.Parties[0].Players {
		mages = append(mages, player.(*fakeCooldownMage))
	}
	return sim, mages
}

func playerRef(index int32) *proto.UnitReference {
	return &proto.UnitReference{Type: proto.UnitReference_Player, Index: index}
}

// The manager uses a survival cooldown on its own only for a player in a tank slot, any slot, whether
// or not a target is assigned to it, and a DPS cooldown for everyone.
func TestSurvivalCooldownsAutocastOnlyForTanks(t *testing.T) {
	for _, tc := range []struct {
		name  string
		tanks []*proto.UnitReference
		want  []bool
	}{
		{"no tanks", nil, []bool{false, false}},
		{"the second player tanks", []*proto.UnitReference{playerRef(1)}, []bool{false, true}},
		{"the second player in the second slot", []*proto.UnitReference{playerRef(0), playerRef(1)}, []bool{true, true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mages := runCooldownMages(t, tc.tanks, autocastRotation(), autocastRotation())
			for i, mage := range mages {
				if used := mage.casts[fakeDefensiveID] > 0; used != tc.want[i] {
					t.Errorf("player %d used the survival cooldown %d times, want used %v", i, mage.casts[fakeDefensiveID], tc.want[i])
				}
				if mage.casts[fakeOffensiveID] == 0 {
					t.Errorf("player %d never used the DPS cooldown", i)
				}
			}
		})
	}
}

// An APL action that names a survival cooldown uses it for a player no tank slot lists.
func TestSurvivalCooldownCastByTheRotationForANonTank(t *testing.T) {
	mage := runCooldownMages(t, nil, castRotation(fakeDefensiveID))[0]
	if mage.casts[fakeDefensiveID] == 0 {
		t.Errorf("the rotation never cast the survival cooldown")
	}
}

// How often a cooldown is used when it goes off each time it is ready.
func usesWhenAlwaysReady(caster *fakeCooldownMage, id int32) int {
	return int(cooldownMagesDuration/caster.GetMajorCooldown(ActionID{SpellID: id}).Spell.CD.Duration) + 1
}

// A rotation of hardcasts never leaves an explosive the GCD it waits for, and the cooldowns sorted after
// it still go off each time they are ready.
func TestExplosiveDoesNotHoldBackTheOtherCooldowns(t *testing.T) {
	caster := runCooldownMages(t, nil, joinRotations(autocastRotation(), castRotation(fakeHardcastID)))[0]
	if caster.casts[fakeHardcastID] == 0 {
		t.Fatalf("the rotation never cast its spell")
	}
	if n := caster.casts[fakeExplosiveID]; n != 0 {
		t.Fatalf("the explosive was used %d times, want 0: a rotation of hardcasts leaves it no GCD", n)
	}
	if got, want := caster.casts[fakeOffensiveID], usesWhenAlwaysReady(caster, fakeOffensiveID); got != want {
		t.Errorf("the DPS cooldown was used %d times, want %d: once each time it was ready", got, want)
	}
}

// Autocast checked only while the GCD runs never fires the DPS cooldown sorted first, and the explosive
// after it still goes off each time it is ready.
func TestCooldownWaitingForTheGCDDoesNotHoldBackAnExplosive(t *testing.T) {
	caster := runCooldownMages(t, nil, joinRotations(castRotation(fakeInstantID), autocastRotation()))[0]
	if caster.casts[fakeInstantID] == 0 {
		t.Fatalf("the rotation never cast its spell")
	}
	if n := caster.casts[fakeOffensiveID]; n != 0 {
		t.Fatalf("the DPS cooldown was used %d times, want 0: autocast only ever runs during the GCD", n)
	}
	if got, want := caster.casts[fakeExplosiveID], usesWhenAlwaysReady(caster, fakeExplosiveID); got != want {
		t.Errorf("the explosive was used %d times, want %d: once each time it was ready", got, want)
	}
}

// With the rotation also evaluated between GCDs, an explosive that comes ready with less of the GCD left
// than its cast waits for the next GCD to start, so it never delays the instant spell cast on each GCD,
// and it still goes off each time it is ready.
func TestExplosiveWaitsForAGCDThatCoversItsCast(t *testing.T) {
	sim, mages := newCooldownMages(nil, joinRotations(castRotation(fakeInstantID), autocastRotation()))
	caster := mages[0]
	caster.reactEvery = 100 * time.Millisecond
	sim.runOnce()
	if got, want := caster.casts[fakeExplosiveID], usesWhenAlwaysReady(caster, fakeExplosiveID); got != want {
		t.Fatalf("the explosive was used %d times, want %d: once each time it was ready", got, want)
	}
	if got, want := caster.casts[fakeInstantID], int(cooldownMagesDuration/GCDDefault)+1; got != want {
		t.Errorf("the instant spell was cast %d times, want %d: once each GCD", got, want)
	}
}

// An explosive whose cast is longer than any GCD still goes off each time it is ready.
func TestExplosiveLongerThanAnyGCDStillGoes(t *testing.T) {
	sim, mages := newCooldownMages(nil, joinRotations(castRotation(fakeInstantID), autocastRotation()))
	caster := mages[0]
	explosive := caster.GetSpell(ActionID{SpellID: fakeExplosiveID})
	explosive.DefaultCast.CastTime = 3 * time.Second
	if castTime := explosive.CastTime(); castTime <= GCDMin {
		t.Fatalf("the explosive casts in %v, want longer than the shortest GCD", castTime)
	}
	sim.runOnce()
	if got, want := caster.casts[fakeExplosiveID], usesWhenAlwaysReady(caster, fakeExplosiveID); got != want {
		t.Errorf("the explosive was used %d times, want %d: once each time it was ready", got, want)
	}
}

// A stat buff's cooldown type: any defensive stat or resistance makes it survival, mana regen and
// spirit alone mana, and anything else DPS.
func TestInferCDTypeFromTheBuffedStats(t *testing.T) {
	for _, tc := range []struct {
		stats []stats.Stat
		want  CooldownType
	}{
		{[]stats.Stat{stats.Armor}, CooldownTypeSurvival},
		{[]stats.Stat{stats.BlockValue}, CooldownTypeSurvival},
		{[]stats.Stat{stats.BlockPercent}, CooldownTypeSurvival},
		{[]stats.Stat{stats.DodgePercent}, CooldownTypeSurvival},
		{[]stats.Stat{stats.ParryPercent}, CooldownTypeSurvival},
		{[]stats.Stat{stats.ArcaneResistance}, CooldownTypeSurvival},
		{[]stats.Stat{stats.FireResistance}, CooldownTypeSurvival},
		{[]stats.Stat{stats.FrostResistance}, CooldownTypeSurvival},
		{[]stats.Stat{stats.NatureResistance}, CooldownTypeSurvival},
		{[]stats.Stat{stats.ShadowResistance}, CooldownTypeSurvival},
		{[]stats.Stat{stats.Spirit, stats.Armor}, CooldownTypeSurvival},
		{[]stats.Stat{stats.MP5}, CooldownTypeMana},
		{[]stats.Stat{stats.Spirit}, CooldownTypeMana},
		{[]stats.Stat{stats.MP5, stats.Spirit}, CooldownTypeMana},
		{[]stats.Stat{stats.Spirit, stats.Strength}, CooldownTypeDPS},
		{[]stats.Stat{stats.AttackPower}, CooldownTypeDPS},
		{[]stats.Stat{stats.SpellCritPercent}, CooldownTypeDPS},
		{nil, CooldownTypeDPS},
	} {
		if got := (&StatBuffAura{BuffedStatTypes: tc.stats}).InferCDType(); got != tc.want {
			t.Errorf("%v: type %v, want %v", tc.stats, got, tc.want)
		}
	}
}

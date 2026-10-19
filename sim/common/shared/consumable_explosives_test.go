package shared

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

const (
	explosivesCategory   int32 = 24
	explosivesCategoryMs int32 = 60000
)

// A test consumable under Goblin Sapper Charge's id 10646 (13241, a 5 min cooldown of its own, needs
// Engineering), which the spec selects by its flag, and test explosives standing in for Ez-Thro
// Dynamite II 18588 (23000) and Crystal Charge 11566 (15239).
const (
	goblinSapperID     int32 = 10646
	testDynamiteID     int32 = 991202
	testCrystalID      int32 = 991203
	testUnskilledSapID int32 = 991204
)

func init() {
	explosive := func(id int32, spellID int32, cooldownMs int32, profession proto.Profession) *proto.Consumable {
		return &proto.Consumable{Id: id, Name: "Test Explosive", Type: proto.ConsumableType_ConsumableTypeExplosive,
			ItemEffects:        []*proto.ItemEffect{testOnUse(spellID, cooldownMs, explosivesCategory, explosivesCategoryMs)},
			RequiredProfession: profession}
	}
	core.AddToDatabase(&proto.SimDatabase{Consumables: []*proto.Consumable{
		explosive(goblinSapperID, 13241, 300000, proto.Profession_Engineering),
		explosive(testDynamiteID, 23000, 0, proto.Profession_ProfessionUnknown),
		explosive(testCrystalID, 15239, 0, proto.Profession_ProfessionUnknown),
		explosive(testUnskilledSapID, 13241, 300000, proto.Profession_Engineering),
	}})
	for _, id := range []int32{goblinSapperID, testDynamiteID, testCrystalID, testUnskilledSapID} {
		NewSpellDataExplosive(id)
	}
}

// A caster with the given profession carrying the consumables, against the given number of targets,
// its spells certain to hit and never to crit.
func newExplosiveSim(t *testing.T, consumes *proto.ConsumesSpec, profession proto.Profession, targetCount int) (*core.Simulation, *testCaster) {
	t.Helper()
	targets := make([]*proto.Target, targetCount)
	for i := range targets {
		targets[i] = &proto.Target{Name: "target", Level: 60, MobType: proto.MobType_MobTypeDemon}
	}

	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name:        "Caster",
				Class:       proto.Class_ClassShaman,
				Race:        proto.Race_RaceTroll,
				Buffs:       &proto.IndividualBuffs{},
				Spec:        &proto.Player_ElementalShaman{},
				Equipment:   &proto.EquipmentSpec{},
				Consumables: consumes,
				Profession1: profession,
				Rotation:    &proto.APLRotation{},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: targets, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()

	caster := sim.Raid.Parties[0].Players[0].(*testCaster)
	caster.AddStatsDynamic(sim, stats.Stats{
		stats.SpellHitPercent:  100,
		stats.SpellCritPercent: -caster.GetStat(stats.SpellCritPercent),
	})
	return sim, caster
}

// 13241 rolls its damage on every enemy around the thrower, each target rolling its own, and the
// thrower's own damage on the thrower; the thrower's half is a spell of its own tagged 1, a harmful
// spell hit.
func TestSapperDamagesEveryEnemyAndTheThrower(t *testing.T) {
	sim, caster := newExplosiveSim(t, &proto.ConsumesSpec{ExtraIds: []int32{goblinSapperID}}, proto.Profession_Engineering, 3)
	sapper := onUseSpell(t, caster, goblinSapperID)
	self := caster.GetSpell(core.ActionID{ItemID: goblinSapperID}.WithTag(1))
	if self == nil {
		t.Fatalf("the sapper registered no spell for its damage to the thrower")
	}
	if self.ProcMask != core.ProcMaskSpellDamage {
		t.Errorf("the thrower's half carries proc mask %v, want spell damage", self.ProcMask)
	}
	if sapper.BonusHitPercent != 0 || self.BonusHitPercent != 0 {
		t.Errorf("the sapper carries %v%% bonus hit and its thrower's half %v%%, want none", sapper.BonusHitPercent, self.BonusHitPercent)
	}

	row := spelldata.MustFind(13241)
	enemy, thrower := row.DamageEffect(), row.CasterDamageEffect()
	health := caster.CurrentHealth()
	if !sapper.Cast(sim, sim.Encounter.ActiveTargetUnits[0]) {
		t.Fatalf("the sapper could not be used")
	}

	var dealt []float64
	for i, target := range sim.Encounter.ActiveTargetUnits {
		d := sapper.SpellMetrics[target.UnitIndex].TotalDamage
		dealt = append(dealt, d)
		if low, high := enemy.Min(), enemy.Max(); d < low || d > high {
			t.Errorf("target %d took %v, want one roll of %v to %v", i, d, low, high)
		}
	}
	if dealt[0] == dealt[1] && dealt[1] == dealt[2] {
		t.Errorf("every target took %v, want a roll of its own each", dealt[0])
	}

	took := health - caster.CurrentHealth()
	if low, high := thrower.Min(), thrower.Max(); took < low || took > high {
		t.Errorf("the thrower took %v, want one roll of %v to %v", took, low, high)
	}
}

// An explosive whose damage lands on an area around the thrower and whose row states no range reaches
// as far as the area's radius: 13241 is usable from its 10 yd and not a yard past it.
func TestAnExplosiveAroundTheThrowerReachesOnlyItsRadius(t *testing.T) {
	row := spelldata.MustFind(13241)
	effect := row.DamageEffect()
	if !effect.HitsAroundTheCaster() || row.MaxRange != 0 {
		t.Fatalf("13241 states range %v and area around the thrower %v, nothing to check", row.MaxRange, effect.HitsAroundTheCaster())
	}
	sim, caster := newExplosiveSim(t, &proto.ConsumesSpec{ExtraIds: []int32{goblinSapperID}}, proto.Profession_Engineering, 1)
	sapper := onUseSpell(t, caster, goblinSapperID)
	target := sim.Encounter.ActiveTargetUnits[0]

	caster.DistanceFromTarget = float64(effect.RadiusMax)
	if !sapper.CanCast(sim, target) {
		t.Errorf("the sapper cannot be used from its %v yd radius", effect.RadiusMax)
	}
	caster.DistanceFromTarget = float64(effect.RadiusMax) + 1
	if sapper.CanCast(sim, target) {
		t.Errorf("the sapper can be used from %v yd, past its %v yd radius", caster.DistanceFromTarget, effect.RadiusMax)
	}
}

// Every test explosive names category 24: using one puts the others out of reach for its 60 s, and
// the test sapper's own 5 min cooldown runs beside it.
func TestExplosivesShareTheirCategoryCooldown(t *testing.T) {
	sim, caster := newExplosiveSim(t, &proto.ConsumesSpec{ExplosiveId: testDynamiteID, ExtraIds: []int32{goblinSapperID}}, proto.Profession_Engineering, 1)
	sapper, dynamite := onUseSpell(t, caster, goblinSapperID), onUseSpell(t, caster, testDynamiteID)
	target := sim.Encounter.ActiveTargetUnits[0]

	if !sapper.Cast(sim, target) {
		t.Fatalf("the sapper could not be used")
	}
	start := sim.CurrentTime
	if got := dynamite.ReadyAt(); got != start+time.Minute {
		t.Errorf("the dynamite is ready at %v after the sapper was used at %v, want 60 s later", got, start)
	}
	if got := sapper.ReadyAt(); got != start+5*time.Minute {
		t.Errorf("the sapper is ready again at %v after it was used at %v, want 5 min later", got, start)
	}
}

// A copy of 23000 flying at 14 yards a second: from 14 yards the dynamite lands 1 s after the cast its
// row states completes.
func TestDynamiteLandsAfterItsCastAndFlight(t *testing.T) {
	editRow(t, 23000, func(s *spelldata.Spell) { s.Speed = 14 })
	cast := spelldata.MustFind(23000).CastTime()
	sim, caster := newExplosiveSim(t, &proto.ConsumesSpec{ExplosiveId: testDynamiteID}, proto.Profession_ProfessionUnknown, 1)
	dynamite := onUseSpell(t, caster, testDynamiteID)
	target := sim.Encounter.ActiveTargetUnits[0]
	metrics := &dynamite.SpellMetrics[target.UnitIndex]
	if got := dynamite.DefaultCast.CastTime; got != cast {
		t.Errorf("the dynamite casts in %v, want the row's %v", got, cast)
	}

	caster.DistanceFromTarget = 14
	start := sim.CurrentTime
	if !dynamite.Cast(sim, target) {
		t.Fatalf("the dynamite could not be used from 14 yards")
	}
	stepPast(t, sim, start+cast+time.Second-time.Millisecond)
	if metrics.TotalDamage != 0 {
		t.Errorf("the dynamite dealt %v before its cast and 1 s flight ended", metrics.TotalDamage)
	}
	stepPast(t, sim, start+cast+time.Second+time.Millisecond)
	if metrics.TotalDamage == 0 {
		t.Errorf("the dynamite had dealt nothing 1 s after its cast completed")
	}
}

// An instant copy of 15239 flying at 14 yards a second: from 14 yards the charge lands 1 s after it is
// used.
func TestCrystalChargeLandsAfterItsFlight(t *testing.T) {
	editRow(t, 15239, func(s *spelldata.Spell) { s.Speed, s.CastTimeMs = 14, 0 })
	sim, caster := newExplosiveSim(t, &proto.ConsumesSpec{ExplosiveId: testCrystalID}, proto.Profession_ProfessionUnknown, 1)
	charge := onUseSpell(t, caster, testCrystalID)
	target := sim.Encounter.ActiveTargetUnits[0]

	caster.DistanceFromTarget = 14
	start := sim.CurrentTime
	if !charge.Cast(sim, target) {
		t.Fatalf("the charge could not be used from 14 yards")
	}
	if got := charge.SpellMetrics[target.UnitIndex].TotalDamage; got != 0 {
		t.Errorf("the charge dealt %v when it was used, want nothing until it lands", got)
	}
	stepPast(t, sim, start+time.Second+time.Millisecond)
	if charge.SpellMetrics[target.UnitIndex].TotalDamage == 0 {
		t.Errorf("the charge had dealt nothing 1 s after it was used")
	}
}

// An explosive that requires Engineering cannot be used in a shapeshift form, which the game files do
// not mark on its spell; one anyone can throw, such as the dynamite, keeps no such requirement.
func TestEngineeringExplosivesCannotBeUsedShapeshifted(t *testing.T) {
	_, caster := newExplosiveSim(t, &proto.ConsumesSpec{ExplosiveId: testDynamiteID, ExtraIds: []int32{goblinSapperID}}, proto.Profession_Engineering, 1)
	if sapper := onUseSpell(t, caster, goblinSapperID); !sapper.CastRequirement.NotShapeshifted {
		t.Error("the Engineering sapper can be used in a shapeshift form")
	}
	if dynamite := onUseSpell(t, caster, testDynamiteID); dynamite.CastRequirement.NotShapeshifted {
		t.Error("the dynamite, which needs no profession, cannot be used in a shapeshift form")
	}
}

// The sapper requires Engineering: a thrower without it has no sapper to use.
func TestExplosiveNeedsTheProfessionItRequires(t *testing.T) {
	_, caster := newExplosiveSim(t, &proto.ConsumesSpec{ExplosiveId: testUnskilledSapID}, proto.Profession_Blacksmithing, 1)
	if spell := caster.GetSpell(core.ActionID{ItemID: testUnskilledSapID}); spell != nil {
		t.Errorf("a thrower without Engineering has the sapper %v", spell.ActionID)
	}
}

// 13241's half on the thrower lands however poor the thrower's hit, while the half on the enemies
// rolls the hit table and, with no hit at all, misses.
func TestSapperAlwaysHitsTheThrower(t *testing.T) {
	sim, caster := newExplosiveSim(t, &proto.ConsumesSpec{ExtraIds: []int32{goblinSapperID}}, proto.Profession_Engineering, 1)
	caster.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: -1000})
	sapper := onUseSpell(t, caster, goblinSapperID)
	self := caster.GetSpell(core.ActionID{ItemID: goblinSapperID}.WithTag(1))
	target := sim.Encounter.ActiveTargetUnits[0]
	thrower := spelldata.MustFind(13241).CasterDamageEffect()

	const uses = 20
	for range uses {
		sapper.CD.Reset()
		sapper.SharedCD.Reset()
		health := caster.CurrentHealth()
		if !sapper.Cast(sim, target) {
			t.Fatalf("the sapper could not be used")
		}
		if took := health - caster.CurrentHealth(); took < thrower.Min() {
			t.Errorf("the thrower took %v, want at least %v", took, thrower.Min())
		}
		caster.GainHealth(sim, caster.MaxHealth(), self.HealthMetrics(&caster.Unit))
	}

	if metrics := sapper.SpellMetrics[target.UnitIndex]; metrics.Misses != uses {
		t.Errorf("the enemy half missed %d of %d uses with no hit, want every one", metrics.Misses, uses)
	}
}

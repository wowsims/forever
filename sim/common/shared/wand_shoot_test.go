package shared

import (
	"cmp"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

const (
	shadowWandID   int32 = 991801
	physicalWandID int32 = 991802
	markerSpellID  int32 = 990002

	shadowWandSpeed   = 1800 * time.Millisecond
	physicalWandSpeed = 1700 * time.Millisecond
	shotWindup        = 500 * time.Millisecond // Shoot's wind-up before its first shot, the sim's own.
)

// A wand of the given speed and school rolling 21 to 40.
func testWand(id int32, speed time.Duration, school proto.SpellSchool) *proto.SimItem {
	return &proto.SimItem{
		Id:               id,
		Name:             "Test Wand",
		Type:             proto.ItemType_ItemTypeRanged,
		RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeWand,
		WeaponSpeed:      speed.Seconds(),
		SpellSchool:      school,
		ScalingOptions:   map[int32]*proto.ScalingItemProperties{0: {WeaponDamageMin: 21, WeaponDamageMax: 40}},
	}
}

func init() {
	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{
		testWand(shadowWandID, shadowWandSpeed, proto.SpellSchool_SpellSchoolShadow),
		testWand(physicalWandID, physicalWandSpeed, proto.SpellSchool_SpellSchoolPhysical),
	}})
	core.RegisterAgentFactory(
		proto.Player_DpsPriest{},
		proto.Spec_SpecDpsPriest,
		func(character *core.Character, _ *proto.Player, _ *proto.Raid) core.Agent {
			return &wandCaster{testCaster: testCaster{Character: *character}}
		},
		func(player *proto.Player, spec any) {
			player.Spec = spec.(*proto.Player_DpsPriest)
		},
	)
}

// Registers Shoot where the classes do, while the character is built, a three-charge buff that any
// ranged attack hit spends, and a marker spell for a rotation to cast when its condition holds.
type wandCaster struct {
	testCaster
	shoot   *core.Spell
	shot    *core.Spell
	charges *core.Aura
	marker  *core.Spell
}

func (c *wandCaster) Initialize() {
	RegisterShoot(&c.Character)
	c.shoot = c.GetSpell(core.ActionID{SpellID: shootRow.ID})
	c.shot = c.GetSpell(core.ActionID{SpellID: shootRow.ID, Tag: 1})
	c.marker = c.RegisterSpell(core.SpellConfig{
		ActionID: core.ActionID{SpellID: markerSpellID},
		Flags:    core.SpellFlagAPL,
		Cast:     core.CastConfig{DefaultCast: core.Cast{GCD: core.GCDDefault}},
	})
	row := &spelldata.Spell{ID: 990001, Name: "Test Charges", ProcFlags: [2]uint32{dbcenums.PROC_FLAG_DEAL_RANGED_ATTACK, 0}}
	c.charges = c.RegisterAura(core.Aura{Label: "Test Charges", ActionID: core.ActionID{SpellID: row.ID}, Duration: time.Minute, MaxStacks: 3})
	spelldata.AttachChargeSpender(&c.Character, c.charges, row, "Test Charges - Spend", spelldata.Chance(100))
}

// A priest with the given ranged slot and rotation, at the given distance from a level 60 target, with
// its spells and shots certain to hit and never to crit.
func newWandSim(t *testing.T, ranged *proto.ItemSpec, rotation *proto.APLRotation, distance float64) (*core.Simulation, *wandCaster) {
	t.Helper()
	return newSwapWandSim(t, wandSimOptions{ranged: ranged, rotation: rotation, distance: distance})
}

type wandSimOptions struct {
	ranged, swapped *proto.ItemSpec
	rotation        *proto.APLRotation
	distance        float64
	targetLevel     int32
	weaponSkills    []*proto.WeaponSkillValue
}

// newWandSim with a swap set that holds the swapped ranged slot, when it is set, and the given weapon
// skills. The rotation defaults to an empty one and the target level to 60.
func newSwapWandSim(t *testing.T, o wandSimOptions) (*core.Simulation, *wandCaster) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotRanged] = o.ranged
	var swapItems []*proto.ItemSpec
	if o.swapped != nil {
		swapItems = make([]*proto.ItemSpec, len(items))
		for i := range swapItems {
			swapItems[i] = &proto.ItemSpec{}
		}
		swapItems[proto.ItemSlot_ItemSlotRanged] = o.swapped
	}

	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, IsTest: true},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name:               "Wander",
				Class:              proto.Class_ClassPriest,
				Race:               proto.Race_RaceHuman,
				Buffs:              &proto.IndividualBuffs{},
				Spec:               &proto.Player_DpsPriest{},
				Equipment:          &proto.EquipmentSpec{Items: items},
				Rotation:           cmp.Or(o.rotation, &proto.APLRotation{}),
				DistanceFromTarget: o.distance,
				EnableItemSwap:     o.swapped != nil,
				ItemSwap:           &proto.ItemSwap{Items: swapItems},
				WeaponSkills:       o.weaponSkills,
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{
			Targets:  []*proto.Target{{Name: "target", Level: cmp.Or(o.targetLevel, 60), MobType: proto.MobType_MobTypeDemon}},
			Duration: 600,
		},
	}, simsignals.CreateSignals())
	sim.Reset()

	caster := sim.Raid.Parties[0].Players[0].(*wandCaster)
	caster.AddStatsDynamic(sim, stats.Stats{
		stats.SpellHitPercent:     100,
		stats.PhysicalHitPercent:  100,
		stats.SpellCritPercent:    -caster.GetStat(stats.SpellCritPercent),
		stats.PhysicalCritPercent: -caster.GetStat(stats.PhysicalCritPercent),
		stats.RangedCritPercent:   -caster.GetStat(stats.RangedCritPercent),
	})
	return sim, caster
}

func newShadowWandSim(t *testing.T) (*core.Simulation, *wandCaster) {
	t.Helper()
	return newWandSim(t, &proto.ItemSpec{Id: shadowWandID}, &proto.APLRotation{}, 0)
}

// Casts Shoot, whose plain cast fires one shot, n times in a row and returns what each shot dealt.
func shots(t *testing.T, sim *core.Simulation, caster *wandCaster, n int) []float64 {
	t.Helper()
	target := sim.Encounter.ActiveTargetUnits[0]
	metrics := &caster.shot.SpellMetrics[target.UnitIndex]
	dealt := make([]float64, n)
	for i := range dealt {
		before := metrics.TotalDamage
		start := sim.CurrentTime
		if !caster.shoot.Cast(sim, target) {
			t.Fatalf("Shoot %d could not be cast at %v", i+1, start)
		}
		stepPast(t, sim, start+max(shadowWandSpeed, physicalWandSpeed)+time.Millisecond)
		dealt[i] = metrics.TotalDamage - before
	}
	return dealt
}

func landed(metrics *core.SpellMetrics) int32 {
	return metrics.Hits + metrics.Crits
}

func TestWandShoot(t *testing.T) {
	t.Run("a shot rolls the wand's damage in its school, with no power", func(t *testing.T) {
		sim, caster := newShadowWandSim(t)
		if caster.shoot.SpellSchool != core.SpellSchoolShadow || caster.shot.SpellSchool != core.SpellSchoolShadow {
			t.Fatalf("Shoot and its shot with the shadow wand are %v and %v, want shadow", caster.shoot.SpellSchool, caster.shot.SpellSchool)
		}
		wand := caster.WeaponFromRanged()
		plain := shots(t, sim, caster, 200)
		for i, damage := range plain {
			if damage < wand.BaseDamageMin || damage > wand.BaseDamageMax {
				t.Fatalf("shot %d dealt %v, want %v to %v", i+1, damage, wand.BaseDamageMin, wand.BaseDamageMax)
			}
		}

		sim, caster = newShadowWandSim(t)
		caster.AddStatsDynamic(sim, stats.Stats{stats.SpellDamage: 500, stats.AttackPower: 500, stats.RangedAttackPower: 500})
		powered := shots(t, sim, caster, 200)
		for i := range powered {
			if powered[i] != plain[i] {
				t.Fatalf("with 500 spell power and attack power shot %d dealt %v, want the same %v as without", i+1, powered[i], plain[i])
			}
		}
	})

	t.Run("a shot is never partially resisted", func(t *testing.T) {
		sim, caster := newShadowWandSim(t)
		plain := shots(t, sim, caster, 200)

		sim, caster = newShadowWandSim(t)
		sim.Encounter.ActiveTargetUnits[0].AddStatsDynamic(sim, stats.Stats{stats.ShadowResistance: 150})
		resisted := shots(t, sim, caster, 200)
		for i := range resisted {
			if resisted[i] != plain[i] {
				t.Fatalf("against 150 shadow resistance shot %d dealt %v, want the same %v as against none", i+1, resisted[i], plain[i])
			}
		}
	})

	t.Run("a shot crits on spell crit for 1.5 times its roll", func(t *testing.T) {
		sim, caster := newShadowWandSim(t)
		if caster.shot.DefenseType != core.DefenseTypeMagic {
			t.Fatalf("a shot's defense type is %v, want magic", caster.shot.DefenseType)
		}
		plain := shots(t, sim, caster, 50)

		sim, caster = newShadowWandSim(t)
		caster.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: 100})
		crits := shots(t, sim, caster, 50)
		for i := range crits {
			if !core.WithinToleranceFloat64(1.5*plain[i], crits[i], 1e-9) {
				t.Fatalf("shot %d crit for %v, want 1.5 × %v", i+1, crits[i], plain[i])
			}
		}
		metrics := caster.shot.SpellMetrics[sim.Encounter.ActiveTargetUnits[0].UnitIndex]
		if metrics.Crits != 50 || metrics.Hits != 0 {
			t.Fatalf("with 100%% spell crit Shoot landed %d crits and %d hits, want 50 and 0", metrics.Crits, metrics.Hits)
		}
	})

	t.Run("ranged crit does not crit a shot", func(t *testing.T) {
		sim, caster := newShadowWandSim(t)
		caster.AddStatsDynamic(sim, stats.Stats{stats.PhysicalCritPercent: 100, stats.RangedCritPercent: 100})
		shots(t, sim, caster, 50)
		metrics := caster.shot.SpellMetrics[sim.Encounter.ActiveTargetUnits[0].UnitIndex]
		if metrics.Crits != 0 || metrics.Hits != 50 {
			t.Fatalf("with 100%% ranged crit and no spell crit Shoot landed %d crits and %d hits, want 0 and 50", metrics.Crits, metrics.Hits)
		}
	})

	// Shoots the shadow wand 300 times at a target of the given level, with the given weapon skills and the
	// stats bonus reads off the caster, and returns Shoot's metrics.
	shoot300 := func(t *testing.T, targetLevel int32, bonus func(*wandCaster) stats.Stats, weaponSkills ...*proto.WeaponSkillValue) core.SpellMetrics {
		t.Helper()
		sim, caster := newSwapWandSim(t, wandSimOptions{ranged: &proto.ItemSpec{Id: shadowWandID}, targetLevel: targetLevel, weaponSkills: weaponSkills})
		caster.AddStatsDynamic(sim, bonus(caster))
		shots(t, sim, caster, 300)
		return caster.shot.SpellMetrics[sim.Encounter.ActiveTargetUnits[0].UnitIndex]
	}

	// The misses with the given physical hit, no ranged hit and full spell hit.
	misses := func(t *testing.T, targetLevel int32, physicalHit float64, weaponSkills ...*proto.WeaponSkillValue) int32 {
		t.Helper()
		return shoot300(t, targetLevel, func(caster *wandCaster) stats.Stats {
			return stats.Stats{
				stats.PhysicalHitPercent: physicalHit - caster.GetStat(stats.PhysicalHitPercent),
				stats.RangedHitPercent:   -caster.GetStat(stats.RangedHitPercent),
			}
		}, weaponSkills...).Misses
	}

	t.Run("at full wand skill a shot misses on the ranged table, which spell hit does not lower", func(t *testing.T) {
		for _, tc := range []struct {
			targetLevel, low, expected, high int32
		}{{60, 5, 15, 27}, {63, 11, 24, 39}} {
			if got := misses(t, tc.targetLevel, 0); got < tc.low || got > tc.high {
				t.Errorf("with no physical hit and full spell hit Shoot missed %d of 300 shots on a level %d target, want about %d", got, tc.targetLevel, tc.expected)
			}
		}
	})

	t.Run("physical hit lowers a shot's miss", func(t *testing.T) {
		if got := misses(t, 63, 9); got != 0 {
			t.Errorf("with 9%% physical hit Shoot missed %d of 300 shots on a level 63 target, want none", got)
		}
	})

	t.Run("10 wand skill short of the maximum misses like 2 more target levels", func(t *testing.T) {
		short := misses(t, 60, 0, &proto.WeaponSkillValue{Skill: proto.WeaponSkill_WeaponSkillWands, Value: 290})
		capped := misses(t, 62, 0)
		if short != capped || short < 6 || short > 36 {
			t.Errorf("Shoot missed %d of 300 shots at 290 wand skill on a level 60 target and %d at 300 on a level 62 target, want the same, about 18", short, capped)
		}
	})

	t.Run("a game files row on the Wands skill line raises the wand hand's skill", func(t *testing.T) {
		sim, caster := newSwapWandSim(t, wandSimOptions{
			ranged:       &proto.ItemSpec{Id: shadowWandID},
			weaponSkills: []*proto.WeaponSkillValue{{Skill: proto.WeaponSkill_WeaponSkillWands, Value: 290}},
		})
		table := caster.AttackTables[sim.Encounter.ActiveTargetUnits[0].UnitIndex]
		short := table.Ranged.BaseMissChance
		spelldata.ParseStatic(&caster.Character, &spelldata.Spell{ID: 990003, Name: "Test Wand Skill", Effects: []spelldata.Effect{
			{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_SKILL, Misc: 228, BasePoints: 10},
		}})
		if got := caster.PseudoStats.WeaponSkills[proto.WeaponSkill_WeaponSkillWands]; got != 0 {
			t.Errorf("wand skill after a 10 point Wands skill row at 290: %v over the maximum, want 0", got)
		}
		if got := table.Ranged.BaseMissChance; math.Abs(short-0.06) > 1e-9 || math.Abs(got-0.05) > 1e-9 {
			t.Errorf("the wand hand's miss on a level 60 target was %v at 290 wand skill and %v after a 10 point Wands skill row, want 0.06 and 0.05", short, got)
		}
	})

	t.Run("wand skill short of the maximum leaves a shot's crit as it is", func(t *testing.T) {
		crits := func(skill int32) int32 {
			return shoot300(t, 60, func(*wandCaster) stats.Stats {
				return stats.Stats{stats.SpellCritPercent: 20}
			}, &proto.WeaponSkillValue{Skill: proto.WeaponSkill_WeaponSkillWands, Value: skill}).Crits
		}
		full, short := crits(300), crits(275)
		if full < 45 || full > 75 || short != full {
			t.Errorf("Shoot crit %d of 300 shots at 300 wand skill and %d at 275 on a level 60 target with 20%% spell crit, want the same, about 60", full, short)
		}
	})

	t.Run("a shot is never dodged, parried or blocked", func(t *testing.T) {
		sim, caster := newShadowWandSim(t)
		target := sim.Encounter.ActiveTargetUnits[0]
		target.PseudoStats.BaseDodgeChance = 1
		target.PseudoStats.BaseParryChance = 1
		target.PseudoStats.BaseBlockChance = 1
		caster.PseudoStats.InFrontOfTarget = true
		shots(t, sim, caster, 50)
		metrics := caster.shot.SpellMetrics[target.UnitIndex]
		if metrics.Dodges != 0 || metrics.Parries != 0 || metrics.Blocks != 0 || metrics.BlockedCrits != 0 || landed(&metrics) != 50 {
			t.Fatalf("in front of a target certain to dodge, parry and block Shoot landed %d of 50 shots, %d dodged, %d parried, %d blocked, want all landed", landed(&metrics), metrics.Dodges, metrics.Parries, metrics.Blocks+metrics.BlockedCrits)
		}
	})

	t.Run("a first shot fires after the wind-up and flies at 20 yards a second", func(t *testing.T) {
		editRow(t, shootRow.ID, func(s *spelldata.Spell) { s.Speed = 20 })
		sim, caster := newWandSim(t, &proto.ItemSpec{Id: shadowWandID}, &proto.APLRotation{}, 20)
		target := sim.Encounter.ActiveTargetUnits[0]
		metrics := &caster.shot.SpellMetrics[target.UnitIndex]
		if !caster.shoot.Cast(sim, target) {
			t.Fatalf("Shoot could not be cast from 20 yards")
		}

		stepPast(t, sim, shotWindup-time.Millisecond)
		if landed(metrics) != 0 {
			t.Fatalf("Shoot fired before its 0.5 s wind-up ended")
		}
		stepPast(t, sim, shotWindup+time.Millisecond)
		if landed(metrics) != 1 || metrics.TotalDamage != 0 {
			t.Fatalf("at 0.5 s: %d shots, %v damage; want 1 shot still in flight", landed(metrics), metrics.TotalDamage)
		}
		stepPast(t, sim, shotWindup+time.Second-time.Millisecond)
		if metrics.TotalDamage != 0 {
			t.Fatalf("the shot landed %v before its 1 s flight ended", metrics.TotalDamage)
		}
		stepPast(t, sim, shotWindup+time.Second+time.Millisecond)
		if metrics.TotalDamage == 0 {
			t.Fatalf("the shot had dealt nothing at 1.50 s")
		}
	})

	t.Run("a shot logs as a cast of its own and its damage after the flight", func(t *testing.T) {
		editRow(t, shootRow.ID, func(s *spelldata.Spell) { s.Speed = 20 })
		sim, caster := newWandSim(t, &proto.ItemSpec{Id: shadowWandID}, &proto.APLRotation{}, 20)
		var lines []string
		sim.Log = func(message string, vals ...any) {
			lines = append(lines, fmt.Sprintf("[%0.2f] "+message, append([]any{sim.CurrentTime.Seconds()}, vals...)...))
		}
		if !caster.shoot.Cast(sim, sim.Encounter.ActiveTargetUnits[0]) {
			t.Fatalf("Shoot could not be cast from 20 yards")
		}
		stepPast(t, sim, shotWindup+time.Second+time.Millisecond)

		shot := core.ActionID{SpellID: shootRow.ID, Tag: 1}.String()
		want := []struct{ at, text string }{
			{"[0.50]", "Casting " + shot + " (Cost = 0.000, Cast Time = 0s, GCD = 0s"},
			{"[0.50]", "Completed cast " + shot},
			{"[1.50]", "] " + shot + " Hit for "},
		}
		var got []string
		for _, line := range lines {
			if strings.Contains(line, shot) && !strings.Contains(line, "[DEBUG]") {
				got = append(got, line)
			}
		}
		matched := len(got) == len(want)
		for i := 0; matched && i < len(want); i++ {
			matched = strings.HasPrefix(got[i], want[i].at) && strings.Contains(got[i], want[i].text)
		}
		if !matched {
			t.Fatalf("the shot's lines are\n%s\nwant its instant cast at 0.50 s and its hit at 1.50 s", strings.Join(got, "\n"))
		}
	})

	t.Run("a caster registered without a wand reads zero and cannot shoot", func(t *testing.T) {
		for _, tc := range []struct {
			ranged *proto.ItemSpec
			casts  bool
		}{{&proto.ItemSpec{}, true}, {&proto.ItemSpec{Id: shadowWandID}, false}} {
			sim, caster := newWandSim(t, tc.ranged, wandValuesRotation(), 0)
			sim.PrePull()
			stepPast(t, sim, 100*time.Millisecond)
			if casts := caster.marker.SpellMetrics[sim.Encounter.ActiveTargetUnits[0].UnitIndex].Casts > 0; casts != tc.casts {
				t.Fatalf("with ranged item %d the wand's swing time and time to the next shot read 0 and its time since the last the current time: %t, want %t", tc.ranged.Id, casts, tc.casts)
			}
		}

		sim, caster := newWandSim(t, &proto.ItemSpec{}, &proto.APLRotation{}, 0)
		target := sim.Encounter.ActiveTargetUnits[0]
		if caster.shoot.CanCast(sim, target) {
			t.Fatalf("Shoot can be cast with no wand")
		}
		caster.shoot.Cast(sim, target)
		if caster.ChanneledDot != nil {
			t.Fatalf("Shoot is channeling with no wand")
		}
	})

	t.Run("a channel from the pull fires after the wind-up and then every wand speed", func(t *testing.T) {
		sim, caster := newWandSim(t, &proto.ItemSpec{Id: shadowWandID}, channelShootRotation(), 0)
		metrics := &caster.shot.SpellMetrics[sim.Encounter.ActiveTargetUnits[0].UnitIndex]
		sim.PrePull()

		stepPast(t, sim, shotWindup-time.Millisecond)
		if !caster.wanding() || metrics.TotalDamage != 0 {
			t.Fatalf("just before the wind-up ends the pull's Shoot should be channeling with nothing dealt, wanding %t, %v dealt", caster.wanding(), metrics.TotalDamage)
		}
		stepPast(t, sim, shotWindup+time.Millisecond)
		if landed(metrics) != 1 || metrics.TotalDamage == 0 {
			t.Fatalf("at 0.5 s the pull's Shoot should have landed its first shot, %d shots for %v", landed(metrics), metrics.TotalDamage)
		}
		second := shotWindup + shadowWandSpeed
		stepPast(t, sim, second-time.Millisecond)
		if landed(metrics) != 1 {
			t.Fatalf("the second shot should wait a wand speed after the first, %d shots just before %v", landed(metrics), second)
		}
		stepPast(t, sim, second+time.Millisecond)
		casts := caster.shoot.SpellMetrics[sim.Encounter.ActiveTargetUnits[0].UnitIndex].Casts
		if casts != 1 || landed(metrics) != 2 {
			t.Fatalf("at %v the pull's channel should have fired its second shot from one cast, %d casts, %d shots", second, casts, landed(metrics))
		}
	})

	t.Run("cast speed at the start sets the shot interval", func(t *testing.T) {
		castSpeed := 1.1
		sim, caster := newWandSim(t, &proto.ItemSpec{Id: shadowWandID}, channelShootRotation(), 0)
		caster.MultiplyCastSpeed(sim, castSpeed)
		metrics := &caster.shot.SpellMetrics[sim.Encounter.ActiveTargetUnits[0].UnitIndex]
		sim.PrePull()
		interval := time.Duration(float64(shadowWandSpeed) / castSpeed).Round(time.Millisecond)

		for _, step := range []struct {
			at    time.Duration
			shots int32
		}{
			{shotWindup + time.Millisecond, 1},
			{shotWindup + interval - time.Millisecond, 1},
			{shotWindup + interval + time.Millisecond, 2},
			{shotWindup + 2*interval + time.Millisecond, 3},
		} {
			stepPast(t, sim, step.at)
			if got := landed(metrics); got != step.shots {
				t.Fatalf("with 10%% cast speed from the pull Shoot should fire at 0.5 s and then every %v, %d shots at %v, want %d", interval, got, step.at, step.shots)
			}
		}
	})

	t.Run("moving ends Shoot and keeps it from starting", func(t *testing.T) {
		sim, caster := newWandSim(t, &proto.ItemSpec{Id: shadowWandID}, channelShootRotation(), 0)
		target := sim.Encounter.ActiveTargetUnits[0]
		sim.PrePull()

		stepPast(t, sim, time.Second)
		caster.MoveDuration(2*time.Second, sim)
		if caster.wanding() || caster.shoot.CanCast(sim, target) {
			t.Fatalf("moving at 1 s should end Shoot and keep it from starting, wanding %t", caster.wanding())
		}
	})

	t.Run("each shot of a channel spends a charge as it lands", func(t *testing.T) {
		sim, caster := newWandSim(t, &proto.ItemSpec{Id: shadowWandID}, channelShootRotation(), 0)
		sim.PrePull()

		stepPast(t, sim, 100*time.Millisecond)
		caster.charges.Activate(sim)
		caster.charges.SetStacks(sim, 3)
		stepPast(t, sim, shotWindup+shadowWandSpeed+100*time.Millisecond)
		if !caster.wanding() || caster.charges.GetStacks() != 1 {
			t.Fatalf("after the shots at %v and %v the channel should still run with one charge left, wanding %t, %d charges",
				shotWindup, shotWindup+shadowWandSpeed, caster.wanding(), caster.charges.GetStacks())
		}
	})

	t.Run("Wand Specialization's school mask leaves out a physical wand and follows a swapped-in wand's school", func(t *testing.T) {
		for _, tc := range []struct {
			ranged, swapped int32 // 0 swaps nothing in.
			multiplier      float64
		}{
			{shadowWandID, 0, 1.25},
			{physicalWandID, 0, 1},
			{physicalWandID, shadowWandID, 1.25},
			{shadowWandID, physicalWandID, 1},
		} {
			wandShots := func(specialized bool) []float64 {
				o := wandSimOptions{ranged: &proto.ItemSpec{Id: tc.ranged}}
				if tc.swapped != 0 {
					o.swapped = &proto.ItemSpec{Id: tc.swapped}
				}
				sim, caster := newSwapWandSim(t, o)
				if specialized {
					caster.AddStaticMod(wandSpecializationMod())
				}
				if tc.swapped != 0 {
					caster.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)
					stepPast(t, sim, 5*time.Second)
				}
				return shots(t, sim, caster, 20)
			}
			plain, specialized := wandShots(false), wandShots(true)
			for i := range specialized {
				if !core.WithinToleranceFloat64(tc.multiplier*plain[i], specialized[i], 1e-9) {
					t.Fatalf("wand %d, %d swapped in: shot %d dealt %v with the talent's mod, want %v × %v", tc.ranged, tc.swapped, i+1, specialized[i], tc.multiplier, plain[i])
				}
			}
		}
	})
}

// Wand Specialization's mod at its second rank.
func wandSpecializationMod() core.SpellModConfig {
	return core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Pct,
		ProcMask:   core.ProcMaskRangedAuto,
		School:     core.SpellSchool(126),
		FloatValue: 0.25,
	}
}

func TestWandShootReadsEquippedWand(t *testing.T) {
	// Swaps the second ranged slot in and casts Shoot twice, the second time right after the first shot,
	// then checks the shots against the wand now equipped: the second comes one wand speed after the first.
	checkSwappedShoot := func(t *testing.T, ranged, swapped *proto.ItemSpec, school core.SpellSchool, speed time.Duration) {
		t.Helper()
		sim, caster := newSwapWandSim(t, wandSimOptions{ranged: ranged, swapped: swapped})
		caster.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)
		wand := caster.WeaponFromRanged()
		target := sim.Encounter.ActiveTargetUnits[0]
		metrics := &caster.shot.SpellMetrics[target.UnitIndex]
		stepPast(t, sim, 5*time.Second)

		firstShotAt := sim.CurrentTime + shotWindup
		if !caster.shoot.Cast(sim, target) {
			t.Fatalf("Shoot could not be cast with the swapped-in wand")
		}
		if caster.shoot.SpellSchool != school {
			t.Fatalf("Shoot is %v, want the swapped-in wand's %v", caster.shoot.SpellSchool, school)
		}
		stepPast(t, sim, firstShotAt+time.Millisecond)
		if landed(metrics) != 1 {
			t.Fatalf("Shoot fired %d shots by the end of its wind-up, want 1", landed(metrics))
		}
		if metrics.TotalDamage < wand.BaseDamageMin || metrics.TotalDamage > wand.BaseDamageMax {
			t.Fatalf("the shot dealt %v, want %v to %v", metrics.TotalDamage, wand.BaseDamageMin, wand.BaseDamageMax)
		}

		if !caster.shoot.Cast(sim, target) {
			t.Fatalf("Shoot could not be cast again after its first shot")
		}
		stepPast(t, sim, firstShotAt+speed-time.Millisecond)
		if landed(metrics) != 1 {
			t.Fatalf("the second shot came within one wand speed of the first, want it %s after", speed)
		}
		stepPast(t, sim, firstShotAt+speed+time.Millisecond)
		if landed(metrics) != 2 {
			t.Fatalf("Shoot fired %d shots one wand speed after the first, want 2 at the swapped-in wand's %s", landed(metrics), speed)
		}
	}

	t.Run("a wand swapped for another shoots at the new wand's speed, damage and school", func(t *testing.T) {
		checkSwappedShoot(t, &proto.ItemSpec{Id: shadowWandID}, &proto.ItemSpec{Id: physicalWandID}, core.SpellSchoolPhysical, physicalWandSpeed)
	})

	t.Run("a wand swapped in after registering with none shoots at its speed, damage and school", func(t *testing.T) {
		checkSwappedShoot(t, &proto.ItemSpec{}, &proto.ItemSpec{Id: shadowWandID}, core.SpellSchoolShadow, shadowWandSpeed)
	})
}

func (c *wandCaster) wanding() bool {
	return c.ChanneledDot != nil && c.ChanneledDot.Spell == c.shoot
}

// The marker spell, cast while the wand's swing time and time to the next shot read 0 and its time since
// the last shot reads the current time.
func wandValuesRotation() *proto.APLRotation {
	cmp := func(lhs, rhs *proto.APLValue) *proto.APLValue {
		return &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{Op: proto.APLValueCompare_OpEq, Lhs: lhs, Rhs: rhs}}}
	}
	zero := &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: "0s"}}}
	return &proto.APLRotation{
		Type: proto.APLRotation_TypeAPL,
		PriorityList: []*proto.APLListItem{{Action: &proto.APLAction{
			Condition: &proto.APLValue{Value: &proto.APLValue_And{And: &proto.APLValueAnd{Vals: []*proto.APLValue{
				cmp(&proto.APLValue{Value: &proto.APLValue_AutoSwingTime{AutoSwingTime: &proto.APLValueAutoSwingTime{AutoType: proto.APLValueAutoSwingTime_Wand}}}, zero),
				cmp(&proto.APLValue{Value: &proto.APLValue_AutoTimeToNext{AutoTimeToNext: &proto.APLValueAutoTimeToNext{AutoType: proto.APLValueAutoAttackType_WandAuto}}}, zero),
				cmp(&proto.APLValue{Value: &proto.APLValue_AutoTimeSinceLast{AutoTimeSinceLast: &proto.APLValueAutoTimeSinceLast{AutoType: proto.APLValueAutoAttackType_WandAuto}}},
					&proto.APLValue{Value: &proto.APLValue_CurrentTime{CurrentTime: &proto.APLValueCurrentTime{}}}),
			}}}},
			Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: core.ActionID{SpellID: markerSpellID}.ToProto()}},
		}}},
	}
}

// Shoot as a channel that any other ready action interrupts.
func channelShootRotation() *proto.APLRotation {
	return &proto.APLRotation{
		Type: proto.APLRotation_TypeAPL,
		PriorityList: []*proto.APLListItem{{Action: &proto.APLAction{Action: &proto.APLAction_ChannelSpell{ChannelSpell: &proto.APLActionChannelSpell{
			SpellId:     core.ActionID{SpellID: 5019}.ToProto(),
			InterruptIf: &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: "true"}}},
		}}}}},
	}
}

package shared

import (
	"math"
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

const (
	frostguardChilled         int32 = 16927
	lobotomizerBrainDamage    int32 = 1290950
	speedTolerance                  = 1e-9
	frostguardChilledDuration       = 5 * time.Second
	brainDamageDuration             = 30 * time.Second
)

// Neither row states a rate for its chance on hit: Frostguard stays unregistered, and The Lobotomizer
// registers on the 0.4 PPM its item effect carries. The tests below give each a 100% chance of their
// own.
func TestSlowProcRowsStateNoRate(t *testing.T) {
	for _, id := range []int32{frostguardChilled, lobotomizerBrainDamage} {
		got := spelldata.ItemProcUnsupportedWithRate(spelldata.MustFind(id), true, 0)
		if !slices.Equal(got, []string{spelldata.ReasonStatesNoRate}) {
			t.Errorf("%d: unsupported = %v, want only the missing rate", id, got)
		}
	}
	if got := spelldata.ItemProcUnsupportedWithRate(spelldata.MustFind(lobotomizerBrainDamage), true, 0.4); len(got) != 0 {
		t.Errorf("%d with its item's rate: unsupported = %v, want none", lobotomizerBrainDamage, got)
	}
}

func alwaysProcs(s *spelldata.Spell) {
	s.ProcChance = 100
	s.ProcChanceSource = spelldata.ProcChanceColumn
}

func permanentSlowNeck(neckID int32, slow func(*core.Unit) *core.Aura) *proto.ItemSpec {
	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{{Id: neckID, Name: "Test Slow Neck",
		Type: proto.ItemType_ItemTypeNeck, ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}}}})
	core.NewItemEffect(neckID, func(agent core.Agent) {
		agent.GetCharacter().NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
			return core.MakePermanent(slow(target))
		})
	})
	return &proto.ItemSpec{Id: neckID}
}

func newSlowProcSim(t *testing.T, weaponID int32, register func(SpellDataProc), row int32, neck *proto.ItemSpec) (*core.Simulation, *testCaster, *core.Unit) {
	t.Helper()
	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{testOneHander(weaponID)}})
	register(SpellDataProc{Name: "Test Slow Weapon", ItemID: weaponID, TriggerSpellID: row, BuffSpellID: row,
		IsWeaponProc: true})

	items := testHands(&proto.ItemSpec{Id: weaponID}, &proto.ItemSpec{})
	if neck != nil {
		items[proto.ItemSlot_ItemSlotNeck] = neck
	}
	sim := newTestCasterSim(items, nil)
	caster := sim.Raid.Parties[0].Players[0].(*testCaster)
	caster.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100})
	return sim, caster, sim.Encounter.ActiveTargetUnits[0]
}

func landMainHandSwing(sim *core.Simulation, caster *testCaster, target *core.Unit) {
	caster.OnSpellHitDealt(sim, &core.Spell{ProcMask: core.ProcMaskMeleeMHAuto},
		&core.SpellResult{Target: target, Outcome: core.OutcomeHit, Damage: 100})
}

func expectSpeeds(t *testing.T, when string, target *core.Unit, melee float64, cast float64) {
	t.Helper()
	if got := target.TotalMeleeHasteMultiplier(); math.Abs(got-melee) > speedTolerance {
		t.Errorf("%s: the target's melee speed is %v×, want %v×", when, got, melee)
	}
	if got := target.TotalSpellHasteMultiplier(); math.Abs(got-cast) > speedTolerance {
		t.Errorf("%s: the target's cast speed is %v×, want %v×", when, got, cast)
	}
}

// Frostguard 12797's Chilled 16927 states A_MOD_MELEE_HASTE_3 -25 on the target for 5 s, "increasing
// the time between attacks by 25%". A hit slows the target's attacks to 1/1.25×; a second hit 3 s
// later runs the slow to 5 s from then rather than slowing it twice, and it lifts when that runs out.
func TestFrostguardSlowsTheTargetsMeleeAttacks(t *testing.T) {
	editRow(t, frostguardChilled, alwaysProcs)
	sim, caster, target := newSlowProcSim(t, 991400, registerSpellDataDebuffProc, frostguardChilled, nil)
	expectSpeeds(t, "before the proc", target, 1, 1)

	start := sim.CurrentTime
	landMainHandSwing(sim, caster, target)
	expectSpeeds(t, "after the proc", target, 1/1.25, 1)

	stepPast(t, sim, start+3*time.Second)
	landMainHandSwing(sim, caster, target)
	expectSpeeds(t, "after the second proc", target, 1/1.25, 1)

	refreshed := sim.CurrentTime
	stepPast(t, sim, start+frostguardChilledDuration+time.Millisecond)
	expectSpeeds(t, "past the first proc's 5 s", target, 1/1.25, 1)

	stepPast(t, sim, refreshed+frostguardChilledDuration+time.Millisecond)
	expectSpeeds(t, "past the second proc's 5 s", target, 1, 1)
}

// Thunder Clap 11581's -20 and Frostguard's -25 share the attack speed slot: with Thunder Clap up the
// target attacks at 1/1.2×, Frostguard's stronger slow replaces it at 1/1.25× rather than multiplying
// on top, and Thunder Clap's returns when Frostguard's runs out.
func TestFrostguardsSlowSharesThunderClapsSlot(t *testing.T) {
	editRow(t, frostguardChilled, alwaysProcs)
	neck := permanentSlowNeck(991402, func(target *core.Unit) *core.Aura { return buffs.NewSpellDataThunderClapAura(target, true, 0) })
	sim, caster, target := newSlowProcSim(t, 991403, registerSpellDataDebuffProc, frostguardChilled, neck)
	expectSpeeds(t, "under Thunder Clap", target, 1/1.2, 1)

	start := sim.CurrentTime
	landMainHandSwing(sim, caster, target)
	expectSpeeds(t, "after the proc", target, 1/1.25, 1)

	stepPast(t, sim, start+frostguardChilledDuration+time.Millisecond)
	expectSpeeds(t, "past the proc's 5 s", target, 1/1.2, 1)
}

// The Lobotomizer 19324's Brain Damage 1290950 deals its damage and states A_MOD_CASTING_SPEED_NOT_STACK
// -25 on the target for 30 s, "increases the casting time of all spells by 25%". A landed hit slows
// the target's casts to 1/1.25× and deals the damage.
func TestLobotomizerDamagesAndSlowsTheTargetsCasts(t *testing.T) {
	editRow(t, lobotomizerBrainDamage, alwaysProcs)
	sim, caster, target := newSlowProcSim(t, 991401, registerSpellDataDamageProc, lobotomizerBrainDamage, nil)

	start := sim.CurrentTime
	landMainHandSwing(sim, caster, target)
	expectSpeeds(t, "after the proc", target, 1, 1/1.25)

	brainDamage := caster.GetSpell(core.ActionID{SpellID: lobotomizerBrainDamage})
	if brainDamage == nil || brainDamage.SpellMetrics[target.UnitIndex].TotalDamage == 0 {
		t.Errorf("the proc dealt no damage alongside the slow")
	}

	stepPast(t, sim, start+brainDamageDuration-time.Millisecond)
	expectSpeeds(t, "inside the 30 s", target, 1, 1/1.25)

	stepPast(t, sim, start+brainDamageDuration+time.Millisecond)
	expectSpeeds(t, "past the 30 s", target, 1, 1)
}

// Slow 31589's 50% longer casts and Brain Damage's 25% share the cast speed slot: the target under
// Slow keeps casting at 1/1.5× when Brain Damage lands.
func TestLobotomizersSlowSharesSlowsSlot(t *testing.T) {
	editRow(t, lobotomizerBrainDamage, alwaysProcs)
	neck := permanentSlowNeck(991404, buffs.SlowAura)
	sim, caster, target := newSlowProcSim(t, 991405, registerSpellDataDamageProc, lobotomizerBrainDamage, neck)
	expectSpeeds(t, "under Slow", target, 1, 1/1.5)

	landMainHandSwing(sim, caster, target)
	expectSpeeds(t, "after the proc", target, 1, 1/1.5)
}

// Annihilator 12798's Armor Shatter 16928 states A_MOD_RESISTANCE -165 on armor, on the enemy, for
// 45 s, stacking to 3. Three procs leave the target 495 armor down and the wearer's armor untouched;
// a fourth keeps 3 stacks and runs the 45 s from itself; the armor comes back when that runs out.
// The row's 1 PPM is set aside for a chance of 100%.
func TestAnnihilatorShattersTheTargetsArmor(t *testing.T) {
	const armorShatter int32 = 16928
	editRow(t, armorShatter, func(s *spelldata.Spell) {
		alwaysProcs(s)
		s.RPPM = 0
	})
	sim, caster, target := newSlowProcSim(t, 991406, registerSpellDataDebuffProc, armorShatter, nil)
	wearerArmor, targetArmor := caster.GetStat(stats.Armor), target.GetStat(stats.Armor)
	expectArmor := func(when string, want float64) {
		t.Helper()
		if got := target.GetStat(stats.Armor); got != want {
			t.Errorf("%s: the target's armor is %v, want %v", when, got, want)
		}
		if got := caster.GetStat(stats.Armor); got != wearerArmor {
			t.Errorf("%s: the wearer's armor is %v, want its own %v", when, got, wearerArmor)
		}
	}

	for i := 1; i <= 3; i++ {
		landMainHandSwing(sim, caster, target)
		expectArmor("after a proc", targetArmor-165*float64(i))
	}

	stepPast(t, sim, sim.CurrentTime+10*time.Second)
	landMainHandSwing(sim, caster, target)
	expectArmor("after a fourth proc", targetArmor-495)
	refreshed := sim.CurrentTime

	stepPast(t, sim, refreshed+45*time.Second-time.Millisecond)
	expectArmor("inside the refreshed 45 s", targetArmor-495)

	stepPast(t, sim, refreshed+45*time.Second+time.Millisecond)
	expectArmor("past the refreshed 45 s", targetArmor)
}

// The raid's Armor Shatter holds its three stacks from the start, and the wielder's own procs land
// Annihilator's copy beside it in the same category: both are on the target, and its armor drops
// by the stronger of the two once.
func TestAnnihilatorBidsBesideTheRaidsArmorShatter(t *testing.T) {
	const armorShatter, weaponID int32 = 16928, 991409
	editRow(t, armorShatter, func(s *spelldata.Spell) {
		alwaysProcs(s)
		s.RPPM = 0
	})
	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{testOneHander(weaponID)}})
	registerSpellDataDebuffProc(SpellDataProc{Name: "Test Annihilator", ItemID: weaponID,
		TriggerSpellID: armorShatter, BuffSpellID: armorShatter, IsWeaponProc: true})

	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{
			Parties: []*proto.Party{{
				Players: []*proto.Player{{
					Name:      "Caster",
					Class:     proto.Class_ClassShaman,
					Race:      proto.Race_RaceTroll,
					Buffs:     &proto.IndividualBuffs{},
					Spec:      &proto.Player_ElementalShaman{},
					Equipment: &proto.EquipmentSpec{Items: testHands(&proto.ItemSpec{Id: weaponID}, &proto.ItemSpec{})},
				}},
				Buffs: &proto.PartyBuffs{},
			}},
			Debuffs: &proto.Debuffs{ArmorShatter: true},
		},
		Encounter: &proto.Encounter{
			Targets:  []*proto.Target{{Name: "target", Level: 60, MobType: proto.MobType_MobTypeDemon}},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()
	caster := sim.Raid.Parties[0].Players[0].(*testCaster)
	target := sim.Encounter.ActiveTargetUnits[0]

	raids := target.GetAura("Armor Shatter (External)")
	if raids == nil || !raids.IsActive() || raids.GetStacks() != 3 {
		t.Fatalf("the raid's Armor Shatter is %v, want it on the target at three stacks", raids)
	}
	shattered := target.GetStat(stats.Armor)
	raids.Deactivate(sim)
	base := target.GetStat(stats.Armor)
	raids.Activate(sim)
	if shattered != base-495 {
		t.Fatalf("the raid's three stacks leave the target at %v armor, want %v", shattered, base-495)
	}

	for range 3 {
		landMainHandSwing(sim, caster, target)
	}

	if got := target.GetStat(stats.Armor); got != shattered {
		t.Errorf("the target's armor is %v, want the %v one Armor Shatter at three stacks leaves", got, shattered)
	}
	if !raids.IsActive() {
		t.Error("the wielder's procs pushed the raid's Armor Shatter off the target, want both on it")
	}
	var wielders *core.Aura
	for _, aura := range target.GetAuras() {
		if aura != raids && aura.ActionID.SpellID == armorShatter && aura.IsActive() {
			wielders = aura
		}
	}
	if wielders == nil || wielders.GetStacks() != 3 || wielders.Tag != buffs.ArmorShatterCategory {
		t.Errorf("the wielder's Armor Shatter is %v, want it on the target at three stacks, tagged %q",
			wielders, buffs.ArmorShatterCategory)
	}
}

// Depleted Eye of Influence 275630's Eye of Influence 1297082 states A_MOD_ATTACK_POWER and
// A_MOD_RANGED_ATTACK_POWER -60 on the enemy for 30 s, off the harmful spells its wearer casts. Its
// trigger 1297085 states no rate, so the test hands it a 100% chance.
func TestEyeOfInfluenceLowersTheTargetsAttackPower(t *testing.T) {
	const itemID, trigger, debuff int32 = 991408, 1297085, 1297082
	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{{Id: itemID, Name: "Test Eye",
		Type: proto.ItemType_ItemTypeNeck, ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}}}})
	everyHit(t, trigger)
	registerSpellDataDebuffProc(SpellDataProc{Name: "Test Eye", ItemID: itemID, TriggerSpellID: trigger,
		BuffSpellID: debuff})

	items := testHands(&proto.ItemSpec{}, &proto.ItemSpec{})
	items[proto.ItemSlot_ItemSlotNeck] = &proto.ItemSpec{Id: itemID}
	sim := newTestCasterSim(items, nil)
	caster := sim.Raid.Parties[0].Players[0].(*testCaster)
	target := sim.Encounter.ActiveTargetUnits[0]
	ap, rap, wearerAP := target.GetStat(stats.AttackPower), target.GetStat(stats.RangedAttackPower), caster.GetStat(stats.AttackPower)

	caster.bolt.Cast(sim, target)
	if got := target.GetStat(stats.AttackPower); got != ap-60 {
		t.Errorf("the target's attack power is %v, want %v", got, ap-60)
	}
	if got := target.GetStat(stats.RangedAttackPower); got != rap-60 {
		t.Errorf("the target's ranged attack power is %v, want %v", got, rap-60)
	}
	if got := caster.GetStat(stats.AttackPower); got != wearerAP {
		t.Errorf("the wearer's attack power is %v, want its own %v", got, wearerAP)
	}
}

// A row whose aura lands on an enemy is never registered as a buff on the wearer.
func TestAnEnemyAuraIsNeverAWearerBuff(t *testing.T) {
	withTestItems(991407)
	registerSpellDataProc(SpellDataProc{Name: "Test Annihilator", ItemID: 991407, TriggerSpellID: 16928,
		IsWeaponProc: true})
	if core.HasItemEffect(991407) {
		t.Error("Armor Shatter 16928 registered as a stat buff on the wearer")
	}
}

package shared

import (
	"math"
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

const (
	frostguardChilled      int32 = 16927
	lobotomizerBrainDamage int32 = 1290950
	speedTolerance               = 1e-9
)

// The speed a slow effect leaves its target at: a negative percent makes each attack or cast take that
// much longer. A slow of 0 leaves the test nothing to check, and fails it.
func slowedSpeed(t *testing.T, e *spelldata.Effect) float64 {
	t.Helper()
	needAmount(t, "the slow", e.Average())
	return 1 / (1 - e.Average()/100)
}

// A weapon proc whose row states no rate of its own, whether by the 101 that fires on its own
// condition or by naming procs per minute it does not carry, is refused for the missing rate until
// the item's rate answers it. The tests below give each slow row a 100% chance of their own.
func TestSlowProcRowsStateNoRate(t *testing.T) {
	for _, row := range []*spelldata.Spell{
		{ID: 991410, Name: "Test Always Slow", ProcChance: 101, ProcChanceSource: spelldata.ProcChanceAlways},
		{ID: 991411, Name: "Test Rate-less Slow", ProcChance: 100, ProcChanceSource: spelldata.ProcChancePPM},
	} {
		if got := spelldata.ItemProcUnsupportedWithRate(row, true, 0); !slices.Equal(got, []string{spelldata.ReasonStatesNoRate}) {
			t.Errorf("%s: unsupported = %v, want only the missing rate", row.Name, got)
		}
		if got := spelldata.ItemProcUnsupportedWithRate(row, true, 0.4); len(got) != 0 {
			t.Errorf("%s with its item's rate: unsupported = %v, want none", row.Name, got)
		}
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
	sim, caster := newWielderSim(items)
	caster.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100})
	return sim, caster, sim.Encounter.ActiveTargetUnits[0]
}

// A shaman wielding the test's weapons with auto attacks enabled, which a weapon proc's manager
// measures the hands against, and its swings cancelled, so only the test's own swings land.
func newWielderSim(items []*proto.ItemSpec) (*core.Simulation, *testCaster) {
	sim := newTestCasterSimAgainst(items, nil, 1, func(player *proto.Player) { player.Spec = &proto.Player_EnhancementShaman{} })
	caster := sim.Raid.Parties[0].Players[0].(*testCaster)
	caster.AutoAttacks.CancelAutoSwing(sim)
	return sim, caster
}

func landMainHandSwing(sim *core.Simulation, caster *testCaster, target *core.Unit) {
	landSwing(sim, caster, target, core.ProcMaskMeleeMHAuto)
}

func landSwing(sim *core.Simulation, caster *testCaster, target *core.Unit, hand core.ProcMask) {
	caster.OnSpellHitDealt(sim, &core.Spell{ProcMask: hand},
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

// Frostguard 12797's Chilled 16927 states A_MOD_MELEE_HASTE_3 on the target, "increasing the time
// between attacks". A hit slows the target's attacks by the row's share; a second hit halfway through
// runs the slow for the row's duration from then rather than slowing it twice, and it lifts when that
// runs out.
func TestFrostguardSlowsTheTargetsMeleeAttacks(t *testing.T) {
	editRow(t, frostguardChilled, alwaysProcs)
	row := spelldata.MustFind(frostguardChilled)
	slowed, duration := slowedSpeed(t, row.FirstAura(dbcenums.A_MOD_MELEE_HASTE_3)), row.Duration()
	sim, caster, target := newSlowProcSim(t, 991400, registerSpellDataDebuffProc, frostguardChilled, nil)
	expectSpeeds(t, "before the proc", target, 1, 1)

	start := sim.CurrentTime
	landMainHandSwing(sim, caster, target)
	expectSpeeds(t, "after the proc", target, slowed, 1)

	stepPast(t, sim, start+duration/2)
	landMainHandSwing(sim, caster, target)
	expectSpeeds(t, "after the second proc", target, slowed, 1)

	refreshed := sim.CurrentTime
	stepPast(t, sim, start+duration+time.Millisecond)
	expectSpeeds(t, "past the first proc's duration", target, slowed, 1)

	stepPast(t, sim, refreshed+duration+time.Millisecond)
	expectSpeeds(t, "past the second proc's duration", target, 1, 1)
}

// Thunder Clap 11581's slow and Frostguard's share the attack speed slot: with Thunder Clap up the
// target attacks at Thunder Clap's speed, Frostguard's proc leaves it at the stronger of the two
// rather than multiplying them, and Thunder Clap's returns when Frostguard's runs out.
func TestFrostguardsSlowSharesThunderClapsSlot(t *testing.T) {
	editRow(t, frostguardChilled, alwaysProcs)
	row := spelldata.MustFind(frostguardChilled)
	frostguard := slowedSpeed(t, row.FirstAura(dbcenums.A_MOD_MELEE_HASTE_3))
	thunderClap := slowedSpeed(t, spelldata.MustFind(11581).FirstAura(dbcenums.A_MOD_MELEE_HASTE_3))
	neck := permanentSlowNeck(991402, func(target *core.Unit) *core.Aura { return buffs.NewSpellDataThunderClapAura(target, true, 0) })
	sim, caster, target := newSlowProcSim(t, 991403, registerSpellDataDebuffProc, frostguardChilled, neck)
	expectSpeeds(t, "under Thunder Clap", target, thunderClap, 1)

	start := sim.CurrentTime
	landMainHandSwing(sim, caster, target)
	expectSpeeds(t, "after the proc", target, min(thunderClap, frostguard), 1)

	stepPast(t, sim, start+row.Duration()+time.Millisecond)
	expectSpeeds(t, "past the proc's duration", target, thunderClap, 1)
}

// The Lobotomizer 19324's Brain Damage 1290950 deals its damage and states A_MOD_CASTING_SPEED_NOT_STACK
// on the target, "increases the casting time of all spells". A landed hit slows the target's casts
// by the row's share for the row's duration and deals the damage.
func TestLobotomizerDamagesAndSlowsTheTargetsCasts(t *testing.T) {
	editRow(t, lobotomizerBrainDamage, alwaysProcs)
	row := spelldata.MustFind(lobotomizerBrainDamage)
	slowed, duration := slowedSpeed(t, row.FirstAura(dbcenums.A_MOD_CASTING_SPEED_NOT_STACK)), row.Duration()
	sim, caster, target := newSlowProcSim(t, 991401, registerSpellDataDamageProc, lobotomizerBrainDamage, nil)

	start := sim.CurrentTime
	landMainHandSwing(sim, caster, target)
	expectSpeeds(t, "after the proc", target, 1, slowed)

	brainDamage := caster.GetSpell(core.ActionID{SpellID: lobotomizerBrainDamage})
	if brainDamage == nil || brainDamage.SpellMetrics[target.UnitIndex].TotalDamage == 0 {
		t.Errorf("the proc dealt no damage alongside the slow")
	}

	stepPast(t, sim, start+duration-time.Millisecond)
	expectSpeeds(t, "inside the duration", target, 1, slowed)

	stepPast(t, sim, start+duration+time.Millisecond)
	expectSpeeds(t, "past the duration", target, 1, 1)
}

// Slow 31589's 50% longer casts and Brain Damage's slow share the cast speed slot: the target under
// Slow casts at the stronger of the two when Brain Damage lands.
func TestLobotomizersSlowSharesSlowsSlot(t *testing.T) {
	editRow(t, lobotomizerBrainDamage, alwaysProcs)
	brainDamage := slowedSpeed(t, spelldata.MustFind(lobotomizerBrainDamage).FirstAura(dbcenums.A_MOD_CASTING_SPEED_NOT_STACK))
	neck := permanentSlowNeck(991404, buffs.SlowAura)
	sim, caster, target := newSlowProcSim(t, 991405, registerSpellDataDamageProc, lobotomizerBrainDamage, neck)
	expectSpeeds(t, "under Slow", target, 1, 1/1.5)

	landMainHandSwing(sim, caster, target)
	expectSpeeds(t, "after the proc", target, 1, min(1/1.5, brainDamage))
}

// Annihilator 12798's Armor Shatter 16928 states A_MOD_ARMOR_NO_MODS on armor, on the enemy, stacking.
// Each proc takes the row's armor off the target up to its stack count and leaves the wearer's armor
// untouched; a further proc keeps the full stacks and runs the row's duration from itself; the armor
// comes back when that runs out. The row's PPM is set aside for a chance of 100%, and a copy stacks to
// three.
func TestAnnihilatorShattersTheTargetsArmor(t *testing.T) {
	const armorShatter int32 = 16928
	editRow(t, armorShatter, func(s *spelldata.Spell) {
		alwaysProcs(s)
		s.RPPM = 0
		s.MaxStack = 3
	})
	row := spelldata.MustFind(armorShatter)
	stacks, duration := int(row.MaxStack), row.Duration()
	sim, caster, target := newSlowProcSim(t, 991406, registerSpellDataDebuffProc, armorShatter, nil)
	perStack := -row.FirstAura(dbcenums.A_MOD_ARMOR_NO_MODS).MeanAt(target.Level)
	needAmount(t, "16928's armor", perStack)
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

	for i := 1; i <= stacks; i++ {
		landMainHandSwing(sim, caster, target)
		expectArmor("after a proc", targetArmor-perStack*float64(i))
	}
	shattered := targetArmor - perStack*float64(stacks)

	stepPast(t, sim, sim.CurrentTime+duration/4)
	landMainHandSwing(sim, caster, target)
	expectArmor("after a proc past the stack count", shattered)
	refreshed := sim.CurrentTime

	stepPast(t, sim, refreshed+duration-time.Millisecond)
	expectArmor("inside the refreshed duration", shattered)

	stepPast(t, sim, refreshed+duration+time.Millisecond)
	expectArmor("past the refreshed duration", targetArmor)
}

// The raid's Armor Shatter holds its full stacks from the start, and the wielder's own procs land
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
					Spec:      &proto.Player_EnhancementShaman{},
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
	caster.AutoAttacks.CancelAutoSwing(sim)
	target := sim.Encounter.ActiveTargetUnits[0]

	row := spelldata.MustFind(armorShatter)
	stacks := int32(row.MaxStack)
	perStack := -row.FirstAura(dbcenums.A_MOD_ARMOR_NO_MODS).MeanAt(target.Level)
	needAmount(t, "16928's armor", perStack)
	if stacks < 2 {
		t.Fatalf("16928 stacks to %d, nothing to bid at full stacks", stacks)
	}
	raids := target.GetAura("Armor Shatter (External)")
	if raids == nil || !raids.IsActive() || raids.GetStacks() != stacks {
		t.Fatalf("the raid's Armor Shatter is %v, want it on the target at %d stacks", raids, stacks)
	}
	shattered := target.GetStat(stats.Armor)
	raids.Deactivate(sim)
	base := target.GetStat(stats.Armor)
	raids.Activate(sim)
	if want := base - perStack*float64(stacks); shattered != want {
		t.Fatalf("the raid's %d stacks leave the target at %v armor, want %v", stacks, shattered, want)
	}

	for range stacks {
		landMainHandSwing(sim, caster, target)
	}

	if got := target.GetStat(stats.Armor); got != shattered {
		t.Errorf("the target's armor is %v, want the %v one Armor Shatter at full stacks leaves", got, shattered)
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
	if wielders == nil || wielders.GetStacks() != stacks || wielders.Tag != buffs.ArmorShatterCategory {
		t.Errorf("the wielder's Armor Shatter is %v, want it on the target at %d stacks, tagged %q",
			wielders, stacks, buffs.ArmorShatterCategory)
	}
}

// Depleted Eye of Influence 275630's Eye of Influence 1297082 states A_MOD_ATTACK_POWER and
// A_MOD_RANGED_ATTACK_POWER on the enemy, off the harmful spells its wearer casts. Its trigger 1297085
// states no rate, so the test hands it a 100% chance.
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
	row := spelldata.MustFind(debuff)
	apChange := row.FirstAura(dbcenums.A_MOD_ATTACK_POWER).MeanAt(target.Level)
	rapChange := row.FirstAura(dbcenums.A_MOD_RANGED_ATTACK_POWER).MeanAt(target.Level)
	needAmount(t, "1297082's attack power", apChange)
	needAmount(t, "1297082's ranged attack power", rapChange)

	caster.bolt.Cast(sim, target)
	if got := target.GetStat(stats.AttackPower); got != ap+apChange {
		t.Errorf("the target's attack power is %v, want %v", got, ap+apChange)
	}
	if got := target.GetStat(stats.RangedAttackPower); got != rap+rapChange {
		t.Errorf("the target's ranged attack power is %v, want %v", got, rap+rapChange)
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

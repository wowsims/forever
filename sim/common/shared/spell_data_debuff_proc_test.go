package shared

import (
	"fmt"
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
// much longer.
func slowedSpeed(t *testing.T, e *spelldata.Effect) float64 {
	t.Helper()
	needAmount(t, "the slow", e.Average())
	return 1 / (1 - e.Average()/100)
}

// A weapon proc whose row states no rate of its own is refused for the missing rate until the item's rate
// answers it.
func TestSlowProcRowsStateNoRate(t *testing.T) {
	for _, row := range []*spelldata.Spell{
		{ID: 991410, Name: "Test Always Slow", ProcChance: 101, ProcChanceSource: spelldata.ProcChanceAlways},
		{ID: 991411, Name: "Test Rate-less Slow", ProcChance: 100, ProcChanceSource: spelldata.ProcChancePPM},
	} {
		if got := spelldata.ItemProcUnsupportedWithRate(row, true, 0, 0); !slices.Equal(got, []string{spelldata.ReasonStatesNoRate}) {
			t.Errorf("%s: unsupported = %v, want only the missing rate", row.Name, got)
		}
		if got := spelldata.ItemProcUnsupportedWithRate(row, true, 0.4, 0); len(got) != 0 {
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

// Frostguard's Chilled 16927: a hit slows the target's attacks by the row's share; a second hit halfway
// through runs the slow for the row's duration from then rather than slowing it twice.
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

// Thunder Clap 11581's slow and Frostguard's share the attack speed slot: Frostguard's proc leaves the
// target at the stronger of the two rather than multiplying them.
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

// The Lobotomizer's Brain Damage 1290950: a landed hit slows the target's casts by the row's share for
// the row's duration and deals the damage.
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

// Annihilator's Armor Shatter 16928: each proc takes the row's armor off the target up to its stack count
// and leaves the wearer's armor untouched; a further proc keeps the full stacks and runs the row's
// duration from itself.
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

// A sim whose one player wields the real weapon, under the raid's debuffs given. The raid's driver finds
// the wielder by the weapon's id.
func newRaidWielderSim(t *testing.T, weaponID, debuff int32, debuffs *proto.Debuffs) (*core.Simulation, *testCaster, *core.Unit) {
	t.Helper()
	editRow(t, debuff, func(s *spelldata.Spell) {
		alwaysProcs(s)
		s.RPPM = 0
	})
	if !core.HasItemEffect(weaponID) {
		// The item's own procs per minute would replace the row's chance the edit gives it.
		registerSpellDataDebuffProc(SpellDataProc{Name: "Test Wielder", ItemID: weaponID, TriggerSpellID: debuff,
			BuffSpellID: debuff, IsWeaponProc: true, Rate: func(*core.Character, *core.ProcTrigger) {}})
	}
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
			Debuffs: debuffs,
		},
		Encounter: &proto.Encounter{
			Targets:  []*proto.Target{{Name: "target", Level: 60, MobType: proto.MobType_MobTypeDemon}},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()
	caster := sim.Raid.Parties[0].Players[0].(*testCaster)
	caster.AutoAttacks.CancelAutoSwing(sim)
	return sim, caster, sim.Encounter.ActiveTargetUnits[0]
}

// The target's auras that read the spell.
func aurasOfSpell(target *core.Unit, spellID int32) []*core.Aura {
	var auras []*core.Aura
	for _, aura := range target.GetAuras() {
		if aura.ActionID.SpellID == spellID {
			auras = append(auras, aura)
		}
	}
	return auras
}

func auraLabels(auras []*core.Aura) []string {
	labels := make([]string, len(auras))
	for i, aura := range auras {
		labels[i] = aura.Label
	}
	return labels
}

// With the player wielding the weapon, the raid's copy of the weapon's debuff is the player's own.
// Bashguuder's proc stacks the raid's Rivenspike, since both land 17315.
func TestWielderStacksTheRaidsWeaponDebuff(t *testing.T) {
	for _, tc := range []struct {
		name               string
		weaponID, debuff   int32
		debuffs            *proto.Debuffs
		label, raidsWeapon string
	}{
		{"Annihilator", 12798, 16928, &proto.Debuffs{ArmorShatter: true}, "Armor Shatter (Player)", "Annihilator"},
		{"Bashguuder", 13204, 17315, &proto.Debuffs{PunctureArmor: 1}, "Puncture Armor (Player)", "Bashguuder"},
		{"Bashguuder under Rivenspike", 13204, 17315, &proto.Debuffs{PunctureArmor: 2}, "Puncture Armor (Player)", "Rivenspike"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if stacks := spelldata.MustFind(tc.debuff).MaxStack; stacks < 2 {
				t.Fatalf("%d stacks to %d, nothing to share", tc.debuff, stacks)
			}
			sim, caster, target := newRaidWielderSim(t, tc.weaponID, tc.debuff, tc.debuffs)
			auras := aurasOfSpell(target, tc.debuff)
			if len(auras) != 1 || auras[0].Label != tc.label {
				t.Fatalf("the target carries %q reading %d, want the raid's %s as %q alone", auraLabels(auras), tc.debuff, tc.raidsWeapon, tc.label)
			}
			shared := auras[0]
			shared.Activate(sim)
			shared.SetStacks(sim, 1)
			perStack := -spelldata.MustFind(tc.debuff).FirstAura(dbcenums.A_MOD_ARMOR_NO_MODS).MeanAt(target.Level)
			needAmount(t, fmt.Sprintf("%d's armor", tc.debuff), perStack)
			oneStack := target.GetStat(stats.Armor)

			landMainHandSwing(sim, caster, target)

			if got, want := target.GetStat(stats.Armor), oneStack-perStack; got != want {
				t.Errorf("the target's armor is %v, want %v from the second stack", got, want)
			}
			if got := aurasOfSpell(target, tc.debuff); len(got) != 1 {
				t.Errorf("after the wielder's proc the target carries %q reading %d, want one aura", auraLabels(got), tc.debuff)
			}
			if got := shared.GetStacks(); !shared.IsActive() || got != 2 {
				t.Errorf("%q is at %d stacks (active %v), want 2 from the raid's stack and the wielder's", shared.Label, got, shared.IsActive())
			}
		})
	}
}

// Vibroblade's Puncture Armor 1293005 is another row than the 17315 Bashguuder's proc lands, so the
// raid's copy stays the raid's and the wielder's procs land their own beside it.
func TestBashguuderBidsBesideTheRaidsVibroblade(t *testing.T) {
	sim, caster, target := newRaidWielderSim(t, 13204, 17315, &proto.Debuffs{PunctureArmor: 3})
	raids := target.GetAura("Puncture Armor (External)")
	if raids == nil || raids.ActionID.SpellID != 1293005 {
		t.Fatalf("the raid's Puncture Armor is %v, want Vibroblade's 1293005 as the raid's copy", raids)
	}
	raids.Activate(sim)
	raidsStacks := raids.GetStacks()

	landMainHandSwing(sim, caster, target)

	wielders := aurasOfSpell(target, 17315)
	if len(wielders) != 1 || wielders[0] == raids || !wielders[0].IsActive() || wielders[0].Label != "Puncture Armor (Player)" {
		t.Fatalf("the wielder's 17315 is %q, want its own active \"Puncture Armor (Player)\"", auraLabels(wielders))
	}
	if !raids.IsActive() || raids.GetStacks() != raidsStacks {
		t.Errorf("the raid's Vibroblade copy is at %d stacks (active %v), want it untouched at %d", raids.GetStacks(), raids.IsActive(), raidsStacks)
	}
}

// Puncture Armor bids in Faerie Fire's slot: under the raid's Faerie Fire the target keeps Faerie Fire's
// armor alone, and with Faerie Fire gone Puncture Armor's stacks count. Under Sunder Armor both count.
func TestPunctureArmorSharesFaerieFiresSlot(t *testing.T) {
	for _, weapon := range []struct {
		name                    string
		punctureArmor, weaponID int32
	}{
		{"Bashguuder", 17315, 991410},
		{"Vibroblade", 1293005, 991411},
	} {
		t.Run(weapon.name, func(t *testing.T) {
			punctureArmor, weaponID := weapon.punctureArmor, weapon.weaponID
			editRow(t, punctureArmor, func(s *spelldata.Spell) {
				alwaysProcs(s)
				s.RPPM = 0
			})
			core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{testOneHander(weaponID)}})
			registerSpellDataDebuffProc(SpellDataProc{Name: "Test " + weapon.name, ItemID: weaponID,
				TriggerSpellID: punctureArmor, BuffSpellID: punctureArmor, IsWeaponProc: true})

			run := func(debuffs *proto.Debuffs, raidLabel string) (*core.Simulation, *testCaster, *core.Unit, *core.Aura) {
				t.Helper()
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
						Debuffs: debuffs,
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
				// The raid's Sunder Armor builds its stacks a global cooldown apart.
				stepPast(t, sim, 10*time.Second)
				raids := target.GetAura(raidLabel)
				if raids == nil || !raids.IsActive() {
					t.Fatalf("the raid's %s is %v, want it on the target", raidLabel, raids)
				}
				return sim, caster, target, raids
			}

			row := spelldata.MustFind(punctureArmor)
			stacks := max(int32(row.MaxStack), 1)
			perStack := -row.FirstAura(dbcenums.A_MOD_ARMOR_NO_MODS).MeanAt(60)
			needAmount(t, fmt.Sprintf("%d's armor", punctureArmor), perStack)
			punctured := perStack * float64(stacks)

			t.Run("Faerie Fire", func(t *testing.T) {
				sim, caster, target, faerieFire := run(&proto.Debuffs{FaerieFire: true}, "Faerie Fire (External)")
				withFaerieFire := target.GetStat(stats.Armor)
				for range stacks {
					landMainHandSwing(sim, caster, target)
				}
				if got := target.GetStat(stats.Armor); got != withFaerieFire {
					t.Errorf("under Faerie Fire the target's armor is %v, want Faerie Fire's %v alone", got, withFaerieFire)
				}
				faerieFire.Deactivate(sim)
				base := withFaerieFire + -spelldata.MustFind(9907).FirstAura(dbcenums.A_MOD_ARMOR_NO_MODS).MeanAt(target.Level)
				if got, want := target.GetStat(stats.Armor), base-punctured; got != want {
					t.Errorf("with Faerie Fire gone the target's armor is %v, want %v from %d stacks of Puncture Armor", got, want, stacks)
				}
			})

			t.Run("Sunder Armor", func(t *testing.T) {
				sim, caster, target, _ := run(&proto.Debuffs{SunderArmor: true}, "Sunder Armor (External)")
				sundered := target.GetStat(stats.Armor)
				for range stacks {
					landMainHandSwing(sim, caster, target)
				}
				if got, want := target.GetStat(stats.Armor), sundered-punctured; got != want {
					t.Errorf("under Sunder Armor the target's armor is %v, want %v with Puncture Armor's stacks beside it", got, want)
				}
			})
		})
	}
}

// Eye of Influence 1297082 states A_MOD_ATTACK_POWER and A_MOD_RANGED_ATTACK_POWER on the enemy, off the
// harmful spells its wearer casts. Its trigger states no rate, so the test hands it a 100% chance.
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

// Lacerations 10370 and Puncture Armor 17315 each cast their row as a main-hand strike that deals no
// damage, and the debuff goes on with a strike that lands (EV-0145).
func TestMeleeDebuffWeaponProcRollsItsOwnHit(t *testing.T) {
	const punctureArmor int32 = 17315
	for _, tc := range []struct {
		name     string
		row      int32
		weaponID int32
		register func(SpellDataProc)
		applied  func(target *core.Unit) float64
	}{
		{"Lacerations", lacerations, 991710, registerSpellDataAuraProc,
			func(target *core.Unit) float64 { return target.PseudoStats.BonusPhysicalDamageTaken }},
		{"Puncture Armor", punctureArmor, 991712, registerSpellDataDebuffProc,
			func(target *core.Unit) float64 { return -target.GetStat(stats.Armor) }},
	} {
		for i, landing := range []struct {
			name  string
			lands bool
		}{
			{"a strike that lands", true},
			{"a strike that misses", false},
		} {
			t.Run(tc.name+", "+landing.name, func(t *testing.T) {
				editRow(t, tc.row, func(s *spelldata.Spell) {
					alwaysProcs(s)
					s.RPPM = 0
				})
				sim, caster, target := newSlowProcSim(t, tc.weaponID+int32(i), tc.register, tc.row, nil)
				if landing.lands {
					caster.AddStatsDynamic(sim, stats.Stats{stats.PhysicalHitPercent: 100, stats.ExpertisePercent: 100})
				} else {
					caster.AttackTables[target.UnitIndex].MH.BaseMissChance = 1
				}
				strike := mustGetSpell(t, caster, tc.row)
				if !strike.ProcMask.Matches(core.ProcMaskMeleeMHSpecial) || !strike.Flags.Matches(core.SpellFlagProc) {
					t.Fatalf("the strike's mask is %v and its flags %v, want a proc with a main-hand special's mask", strike.ProcMask, strike.Flags)
				}
				before := tc.applied(target)

				landMainHandSwing(sim, caster, target)
				metrics := strike.SpellMetrics[target.UnitIndex]
				if metrics.Casts != 1 {
					t.Fatalf("the swing cast %d strikes, want 1", metrics.Casts)
				}
				if metrics.TotalDamage != 0 {
					t.Errorf("the strike dealt %v damage, want none", metrics.TotalDamage)
				}
				avoided := metrics.Misses + metrics.Dodges + metrics.Parries
				gained := tc.applied(target) - before
				switch {
				case landing.lands && (avoided != 0 || gained <= 0):
					t.Errorf("%d strikes were avoided and the debuff added %v, want the strike to land and the debuff on", avoided, gained)
				case !landing.lands && (avoided != 1 || gained != 0):
					t.Errorf("%d strikes were avoided and the debuff added %v, want the strike avoided and no debuff", avoided, gained)
				}
			})
		}
	}
}

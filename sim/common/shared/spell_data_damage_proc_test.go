package shared

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Vengeance 20049 states Can Proc From Procs, and its harmful spell bit hears a crit of any school.
const vengeanceTalent int32 = 20049

// Storm Gauntlets 12632's 16615 casts Lightning Strike 16614 off each melee hit, and Blazefury
// Medallion 17111's 7711 casts Fire Strike 7712 the same way. Their hits are procs, so an aura hears
// them only where it carries Can Proc From Procs: Vengeance's listener, built from its row as the
// paladin builds it, procs off their crits, and the same listener without the attribute does not.
func TestItemDamageProcCritsReachListenersThatHearProcs(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		itemID, trigger, damage int32
		listenerItemID          int32
		itemType                proto.ItemType
		itemSlot                proto.ItemSlot
	}{
		{"Storm Gauntlets", 991601, 16615, 16614, 991602, proto.ItemType_ItemTypeHands, proto.ItemSlot_ItemSlotHands},
		{"Blazefury Medallion", 991603, 7711, 7712, 991604, proto.ItemType_ItemTypeTrinket, proto.ItemSlot_ItemSlotTrinket1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{
				{Id: tc.itemID, Name: "Test " + tc.name, Type: tc.itemType,
					ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}},
				{Id: tc.listenerItemID, Name: "Test Vengeance", Type: proto.ItemType_ItemTypeNeck,
					ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}},
			}})
			registerSpellDataDamageProc(SpellDataProc{Name: "Test " + tc.name, ItemID: tc.itemID,
				TriggerSpellID: tc.trigger, BuffSpellID: tc.damage})

			heard := map[bool]int{}
			core.NewItemEffect(tc.listenerItemID, func(agent core.Agent) {
				character := agent.GetCharacter()
				for _, canProcFromProcs := range []bool{true, false} {
					trigger := spelldata.ProcTrigger(character, spelldata.MustFind(vengeanceTalent),
						func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
							if spell.ActionID.SpellID == tc.damage {
								heard[canProcFromProcs]++
							}
						})
					trigger.Name = "Test Vengeance"
					if !canProcFromProcs {
						trigger.Name += " Without Procs"
					}
					trigger.ProcMask = core.ProcMaskUnknown
					trigger.CanProcFromProcs = canProcFromProcs
					trigger.TriggerImmediately = true
					character.MakeProcTriggerAura(trigger)
				}
			})

			items := testHands(&proto.ItemSpec{}, &proto.ItemSpec{})
			items[tc.itemSlot] = &proto.ItemSpec{Id: tc.itemID}
			items[proto.ItemSlot_ItemSlotNeck] = &proto.ItemSpec{Id: tc.listenerItemID}
			sim := newTestCasterSim(items, nil)
			caster := sim.Raid.Parties[0].Players[0].(*testCaster)
			target := sim.Encounter.ActiveTargetUnits[0]
			caster.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100, stats.SpellCritPercent: 100})

			landMainHandSwing(sim, caster, target)

			damageSpell := caster.GetSpell(core.ActionID{SpellID: tc.damage})
			if damageSpell == nil || damageSpell.SpellMetrics[target.UnitIndex].Crits != 1 {
				t.Fatalf("the swing did not land one %d crit", tc.damage)
			}
			if heard[true] != 1 {
				t.Errorf("the listener with Can Proc From Procs heard %d crits, want 1", heard[true])
			}
			if heard[false] != 0 {
				t.Errorf("the listener without it heard %d crits, want none", heard[false])
			}
		})
	}
}

// Windfury Weapon's passive 439431 hears melee auto attacks and abilities, without Can Proc From Procs.
const windfuryWeaponPassive int32 = 439431

// The deferred re-rolls of the strikes the swings so far cast, which a chance below 1 ends.
func settleChains(t *testing.T, sim *core.Simulation) {
	t.Helper()
	stepPast(t, sim, sim.CurrentTime+time.Second)
}

func mustGetSpell(t *testing.T, caster *testCaster, id int32) *core.Spell {
	t.Helper()
	spell := caster.GetSpell(core.ActionID{SpellID: id})
	if spell == nil {
		t.Fatalf("no spell %d is registered", id)
	}
	return spell
}

func weaponProcListener(name string, mask core.ProcMask, handler core.ProcHandler) core.ProcTrigger {
	return core.ProcTrigger{
		Name:               name,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           mask,
		Outcome:            core.OutcomeLanded,
		IsWeaponProc:       true,
		ProcChance:         1,
		TriggerImmediately: true,
		Handler:            handler,
	}
}

// Bloodpike 13057's Rend 18202 (a bleed, no hit) and Hanzo Sword 8190's Wound 16405 (a hit) at a chance
// below 1, on both hands. A strike a main-hand swing cast is a main-hand special: the main hand's
// weapon procs hear it, immediate and deferred alike, and its own weapon rolls again. A strike an
// off-hand swing cast is heard by no weapon proc on either hand, its own weapon included. Auto
// attacks' and Windfury Weapon's (Wound only, it needs damage) listeners hear neither. Swings of the
// two hands in turn, unsettled, leave the main-hand strikes heard.
func TestWeaponProcStrikeIsHeardOnlyFromTheMainHand(t *testing.T) {
	for _, tc := range []struct {
		name                     string
		weaponID, neckID, strike int32
		windfury                 bool
	}{
		{"Rend", 991606, 991607, 18202, false},
		{"Wound", 991608, 991609, 16405, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const swings = 20
			editRow(t, tc.strike, func(s *spelldata.Spell) {
				s.ProcChance, s.ProcChanceSource = 90, spelldata.ProcChanceColumn
			})
			core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{testOneHander(tc.weaponID)}})
			withAuraItem(tc.neckID, proto.ItemType_ItemTypeNeck)
			registerSpellDataDamageProc(SpellDataProc{Name: "Test " + tc.name + " Weapon", ItemID: tc.weaponID,
				TriggerSpellID: tc.strike, BuffSpellID: tc.strike, IsWeaponProc: true})

			heard := map[string]map[int32]int{}
			listen := func(name string) core.ProcHandler {
				heard[name] = map[int32]int{}
				return func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
					heard[name][spell.ActionID.SpellID]++
				}
			}
			core.NewItemEffect(tc.neckID, func(agent core.Agent) {
				character := agent.GetCharacter()
				for name, mask := range map[string]core.ProcMask{
					"main-hand oil": core.ProcMaskMeleeMH,
					"off-hand oil":  core.ProcMaskMeleeOH,
					"auto attacks":  core.ProcMaskMeleeWhiteHit,
				} {
					character.MakeProcTriggerAura(weaponProcListener("Test "+name, mask, listen(name)))
				}
				enchant := weaponProcListener("Test main-hand enchant", core.ProcMaskMeleeMH, listen("main-hand enchant"))
				enchant.TriggerImmediately = false
				character.MakeProcTriggerAura(enchant)
				if tc.windfury {
					windfury := spelldata.ProcTrigger(character, spelldata.MustFind(windfuryWeaponPassive),
						listen("Windfury Weapon"), spelldata.Chance(1))
					windfury.Name = "Test Windfury Weapon"
					windfury.ICD = 0
					windfury.TriggerImmediately = true
					character.MakeProcTriggerAura(windfury)
				}
			})

			items := testHands(&proto.ItemSpec{Id: tc.weaponID}, &proto.ItemSpec{Id: tc.weaponID})
			items[proto.ItemSlot_ItemSlotNeck] = &proto.ItemSpec{Id: tc.neckID}
			sim, caster := newWielderSim(items)
			target := sim.Encounter.ActiveTargetUnits[0]

			for range swings {
				landSwing(sim, caster, target, core.ProcMaskMeleeMHAuto)
			}
			settleChains(t, sim)
			strike := mustGetSpell(t, caster, tc.strike)
			if casts := strike.SpellMetrics[target.UnitIndex].Casts; casts <= swings {
				t.Errorf("%d main-hand swings cast %d strikes, want more: a strike re-rolls the main-hand weapon", swings, casts)
			}
			if heard["main-hand oil"][tc.strike] == 0 {
				t.Errorf("the main-hand weapon proc heard no strike a main-hand swing cast")
			}

			if n, want := heard["main-hand enchant"][tc.strike], heard["main-hand oil"][tc.strike]; n != want {
				t.Errorf("the deferred main-hand weapon proc heard %d strikes, want the immediate one's %d", n, want)
			}

			before := map[string]int{}
			for name := range heard {
				before[name] = heard[name][tc.strike]
			}
			casts := strike.SpellMetrics[target.UnitIndex].Casts
			for range swings {
				landSwing(sim, caster, target, core.ProcMaskMeleeOHAuto)
			}
			settleChains(t, sim)
			if n := strike.SpellMetrics[target.UnitIndex].Casts - casts; n == 0 || n > swings {
				t.Errorf("%d off-hand swings cast %d strikes, want some and at most one each: no weapon rolls again", swings, n)
			}
			for name := range heard {
				if n := heard[name][tc.strike] - before[name]; n != 0 {
					t.Errorf("the %s listener heard %d strikes off-hand swings cast, want none", name, n)
				}
			}

			mainHandHeard := heard["main-hand oil"][tc.strike]
			for range swings {
				landSwing(sim, caster, target, core.ProcMaskMeleeOHAuto)
				landSwing(sim, caster, target, core.ProcMaskMeleeMHAuto)
			}
			settleChains(t, sim)
			if heard["main-hand oil"][tc.strike] == mainHandHeard {
				t.Errorf("the main-hand weapon proc heard no strike swings of both hands in turn cast")
			}
			if n, want := heard["main-hand enchant"][tc.strike], heard["main-hand oil"][tc.strike]; n != want {
				t.Errorf("after swings of both hands the deferred main-hand weapon proc heard %d strikes, want the immediate one's %d", n, want)
			}
			if n := heard["off-hand oil"][tc.strike]; n != 0 {
				t.Errorf("the off-hand weapon proc heard %d strikes, want none", n)
			}

			names := []string{"auto attacks"}
			if tc.windfury {
				names = append(names, "Windfury Weapon")
			}
			for _, name := range names {
				if n := heard[name][tc.strike]; n != 0 {
					t.Errorf("the %s listener heard %d strikes, want none", name, n)
				}
				if n := heard[name][0]; n != 4*swings {
					t.Errorf("the %s listener heard %d swings, want %d", name, n, 4*swings)
				}
			}
		})
	}
}

// Two test one-handers, each with its own row and rate: Wound 16405 in the main hand and Rend 18202
// in the off hand, each measured on the weapon carrying it, so a wielder that dual wields carries
// them. A strike a main-hand swing cast rolls the main-hand weapon's proc and never the off-hand
// one's, and one an off-hand swing cast rolls neither: off-hand swings cast Rends that roll nothing,
// and main-hand swings cast Wounds that roll Wound again and never Rend. The off-hand weapon's rate is
// procs per minute in one case and a flat chance in its row's column in the other; both roll on the
// off hand's hits only. The swings are the test's own, so the wielder's auto attacks are off.
func TestWeaponProcStrikeRollsOnlyTheMainHandWeapon(t *testing.T) {
	const wound, rend int32 = 16405, 18202
	const swings = 20
	for _, tc := range []struct {
		name                  string
		mainHandID, offHandID int32
		offHandRate           func(*spelldata.Spell)
	}{
		{"procs per minute", 991614, 991615, func(s *spelldata.Spell) { s.RPPM = 12 }},
		{"flat chance", 991620, 991621, func(s *spelldata.Spell) {
			s.RPPM, s.ProcChance, s.ProcChanceSource = 0, 50, spelldata.ProcChanceColumn
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			editRow(t, wound, func(s *spelldata.Spell) { s.RPPM = 18 })
			editRow(t, rend, tc.offHandRate)
			core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{testOneHander(tc.mainHandID), testOneHander(tc.offHandID)}})
			registerSpellDataDamageProc(SpellDataProc{Name: "Test Wound Weapon", ItemID: tc.mainHandID,
				TriggerSpellID: wound, BuffSpellID: wound, IsWeaponProc: true})
			registerSpellDataDamageProc(SpellDataProc{Name: "Test Rend Weapon", ItemID: tc.offHandID,
				TriggerSpellID: rend, BuffSpellID: rend, IsWeaponProc: true})

			sim := newTestCasterSimAgainst(testHands(&proto.ItemSpec{Id: tc.mainHandID}, &proto.ItemSpec{Id: tc.offHandID}), nil, 1,
				func(player *proto.Player) { player.Spec = &proto.Player_EnhancementShaman{} })
			caster := sim.Raid.Parties[0].Players[0].(*testCaster)
			caster.AutoAttacks.CancelAutoSwing(sim)
			target := sim.Encounter.ActiveTargetUnits[0]
			woundSpell, rendSpell := mustGetSpell(t, caster, wound), mustGetSpell(t, caster, rend)
			casts := func(spell *core.Spell) int32 { return spell.SpellMetrics[target.UnitIndex].Casts }

			for range swings {
				landSwing(sim, caster, target, core.ProcMaskMeleeOHAuto)
			}
			settleChains(t, sim)
			if n := casts(rendSpell); n == 0 || n > swings {
				t.Errorf("%d off-hand swings cast %d Rends, want some and at most one each: a Rend never rolls the off-hand weapon", swings, n)
			}
			if n := casts(woundSpell); n != 0 {
				t.Errorf("Rends off-hand swings cast rolled the main-hand weapon's Wound %d times, want none", n)
			}

			rends, wounds := casts(rendSpell), casts(woundSpell)
			for range swings {
				landSwing(sim, caster, target, core.ProcMaskMeleeMHAuto)
			}
			settleChains(t, sim)
			if n := casts(rendSpell) - rends; n != 0 {
				t.Errorf("%d main-hand swings and the Wounds they cast cast %d Rends, want none", swings, n)
			}
			if n := casts(woundSpell) - wounds; n <= swings {
				t.Errorf("%d main-hand swings cast %d Wounds, want more: a Wound rolls the main-hand weapon again", swings, n)
			}
		})
	}
}

// A strike whose own weapon procs on every hit and that always lands chains for as long as it is
// rolled, each re-roll one spell batch after the strike that asked for it rather than inside it.
func TestWeaponProcStrikeThatAlwaysProcsChainsOneBatchApart(t *testing.T) {
	const weaponID, wound int32 = 991616, 16405
	const batches = 5
	editRow(t, wound, alwaysProcs)
	sim, caster, target := newSlowProcSim(t, weaponID, registerSpellDataDamageProc, wound, nil)
	caster.AddStatsDynamic(sim, stats.Stats{stats.PhysicalHitPercent: 100, stats.ExpertisePercent: 100})
	strike := mustGetSpell(t, caster, wound)
	metrics := &strike.SpellMetrics[target.UnitIndex]

	landMainHandSwing(sim, caster, target)
	if metrics.Casts != 1 {
		t.Fatalf("the swing cast %d strikes at once, want 1: a re-roll waits a batch", metrics.Casts)
	}
	stepPast(t, sim, sim.CurrentTime+batches*core.SpellBatchWindow+time.Millisecond)
	if metrics.Casts != 1+batches {
		t.Errorf("%d batches after the swing %d strikes were cast, want %d", batches, metrics.Casts, 1+batches)
	}
	if avoided := metrics.Misses + metrics.Dodges + metrics.Parries; avoided != 0 {
		t.Errorf("%d strikes were avoided, want every one to land", avoided)
	}
}

// Every landed Rend 18202 puts its bleed on the target, a Rend that its own strike chained included,
// and no weapon proc hears the bleed's ticks. A test one-hander carries the row at a chance below 1,
// and each swing lands once the last bleed has run out.
func TestWeaponProcStrikeChainAppliesEachLandedBleed(t *testing.T) {
	const weaponID, neckID, rend int32 = 991612, 991613, 18202
	const swings = 10
	editRow(t, rend, func(s *spelldata.Spell) {
		s.ProcChance, s.ProcChanceSource = 90, spelldata.ProcChanceColumn
	})
	duration := spelldata.MustFind(rend).Duration()
	withAuraItem(neckID, proto.ItemType_ItemTypeNeck)
	landed := 0
	core.NewItemEffect(neckID, func(agent core.Agent) {
		character := agent.GetCharacter()
		oil := spelldata.ProcTrigger(character, spelldata.MustFind(rend),
			func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
				if spell.ActionID.SpellID == rend {
					landed++
				}
			}, spelldata.WeaponProc(), spelldata.Chance(1))
		oil.Name = "Test Oil"
		oil.TriggerImmediately = true
		character.MakeProcTriggerAura(oil)
	})
	sim, caster, target := newSlowProcSim(t, weaponID, registerSpellDataDamageProc, rend, &proto.ItemSpec{Id: neckID})
	strike := mustGetSpell(t, caster, rend)

	for range swings {
		stepPast(t, sim, sim.CurrentTime+duration+time.Millisecond)
		before := landed
		landMainHandSwing(sim, caster, target)
		settleChains(t, sim)
		if landed > before && !strike.Dot(target).IsActive() {
			t.Fatalf("%d Rends landed off one swing and none left its bleed on the target", landed-before)
		}
	}
	stepPast(t, sim, sim.CurrentTime+duration+time.Millisecond)
	if landed == 0 {
		t.Fatalf("no Rend landed in %d swings", swings)
	}
	metrics := strike.SpellMetrics[target.UnitIndex]
	if metrics.Ticks == 0 {
		t.Fatalf("no bleed ticked")
	}
	if landed > int(metrics.Casts) {
		t.Errorf("the weapon proc heard %d Rends of %d cast: it heard the bleed's %d ticks", landed, metrics.Casts, metrics.Ticks)
	}
}

// A test one-hander carrying the row in the off hand, a plain one in the main hand, and a wielder whose
// strikes always land. The plain one keeps the test's in the off hand: a one-hander equipped with
// the main hand empty goes to the main hand.
func newOffHandProcSim(t *testing.T, weaponID, row int32, neck *proto.ItemSpec) (*core.Simulation, *testCaster, *core.Unit) {
	t.Helper()
	const plainMainHandID int32 = 991622
	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{testOneHander(weaponID), testOneHander(plainMainHandID)}})
	registerSpellDataDamageProc(SpellDataProc{Name: "Test Off-Hand Weapon", ItemID: weaponID, TriggerSpellID: row,
		BuffSpellID: row, IsWeaponProc: true})
	items := testHands(&proto.ItemSpec{Id: plainMainHandID}, &proto.ItemSpec{Id: weaponID})
	if neck != nil {
		items[proto.ItemSlot_ItemSlotNeck] = neck
	}
	sim, caster := newWielderSim(items)
	caster.AddStatsDynamic(sim, stats.Stats{stats.PhysicalHitPercent: 100, stats.ExpertisePercent: 100})
	return sim, caster, sim.Encounter.ActiveTargetUnits[0]
}

// Rend 18202 rolled at 100% in the off hand: the strike an off-hand swing casts puts its bleed on
// the target and rolls its own weapon no more, where a roll that cannot fail would chain.
func TestOffHandWeaponProcStrikeAppliesItsBleed(t *testing.T) {
	const weaponID, rend int32 = 991617, 18202
	editRow(t, rend, alwaysProcs)
	sim, caster, target := newOffHandProcSim(t, weaponID, rend, nil)
	strike := mustGetSpell(t, caster, rend)

	landSwing(sim, caster, target, core.ProcMaskMeleeOHAuto)
	settleChains(t, sim)
	if n := strike.SpellMetrics[target.UnitIndex].Casts; n != 1 {
		t.Errorf("the off-hand swing cast %d Rends, want 1", n)
	}
	if !strike.Dot(target).IsActive() {
		t.Errorf("the Rend the off-hand swing cast left no bleed on the target")
	}
}

// Wound 16405 rolled at 100% in the off hand, given a missile speed: the strike an off-hand swing
// casts lands after its flight, and its own weapon and a main-hand weapon proc hear the landing no
// more than they would an instant strike.
func TestOffHandWeaponProcStrikeInFlightIsHeardByNoWeaponProc(t *testing.T) {
	const weaponID, neckID, wound int32 = 991618, 991619, 16405
	editRow(t, wound, func(s *spelldata.Spell) {
		alwaysProcs(s)
		s.Speed = 20
	})
	withAuraItem(neckID, proto.ItemType_ItemTypeNeck)
	heard := 0
	core.NewItemEffect(neckID, func(agent core.Agent) {
		agent.GetCharacter().MakeProcTriggerAura(weaponProcListener("Test Oil", core.ProcMaskMeleeMH,
			func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
				if spell.ActionID.SpellID == wound {
					heard++
				}
			}))
	})
	sim, caster, target := newOffHandProcSim(t, weaponID, wound, &proto.ItemSpec{Id: neckID})
	strike := mustGetSpell(t, caster, wound)
	caster.DistanceFromTarget = 20
	metrics := &strike.SpellMetrics[target.UnitIndex]

	landSwing(sim, caster, target, core.ProcMaskMeleeOHAuto)
	if caster.SpellsInFlight[strike] != 1 || metrics.TotalDamage != 0 {
		t.Fatalf("the Wound is not in flight after the swing")
	}
	settleChains(t, sim)
	if caster.SpellsInFlight[strike] != 0 || metrics.TotalDamage == 0 {
		t.Fatalf("the Wound did not land")
	}
	if metrics.Casts != 1 {
		t.Errorf("the off-hand swing cast %d Wounds, want 1: its landing rolled its own weapon", metrics.Casts)
	}
	if heard != 0 {
		t.Errorf("the main-hand weapon proc heard %d Wounds, want none", heard)
	}
}

// A weapon proc of magic defense is no weapon's strike: Blaze 16898's hit is heard by no weapon proc.
func TestMagicWeaponProcIsHeardByNoWeaponProc(t *testing.T) {
	const weaponID, neckID, blaze int32 = 991610, 991611, 16898
	editRow(t, blaze, alwaysProcs)
	withAuraItem(neckID, proto.ItemType_ItemTypeNeck)
	heard := map[int32]int{}
	core.NewItemEffect(neckID, func(agent core.Agent) {
		agent.GetCharacter().MakeProcTriggerAura(weaponProcListener("Test Oil", core.ProcMaskUnknown,
			func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
				heard[spell.ActionID.SpellID]++
			}))
	})
	sim, caster, target := newSlowProcSim(t, weaponID, registerSpellDataDamageProc, blaze, &proto.ItemSpec{Id: neckID})

	landMainHandSwing(sim, caster, target)
	damage := mustGetSpell(t, caster, blaze)
	if damage.SpellMetrics[target.UnitIndex].Casts != 1 {
		t.Fatalf("the swing did not cast %d once", blaze)
	}
	if heard[0] != 1 {
		t.Errorf("the weapon proc heard %d swings, want 1", heard[0])
	}
	if heard[blaze] != 0 {
		t.Errorf("the weapon proc heard %d of %d's hits, want none", heard[blaze], blaze)
	}
}

// A damage proc whose trigger carries the kill bit, Blazefury Medallion 17111's 7711 edited to hear a
// killing blow alone. The kill passes the proc no spell and names the dying target, so its damage goes
// to the next target still up, and with none left it casts nothing.
func TestKillingBlowDamageProcHitsTheNextTarget(t *testing.T) {
	for _, tc := range []struct {
		name        string
		trinketID   int32
		targetCount int
		want        int32
	}{
		{"two targets", 991630, 2, 1},
		{"one target", 991631, 1, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			editRow(t, 7711, func(s *spelldata.Spell) {
				s.ProcFlags = [2]uint32{0x2, 0}
				s.ProcHint = 0
			})
			core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{{Id: tc.trinketID, Name: "Test Killing Blow",
				Type: proto.ItemType_ItemTypeTrinket, ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}}}})
			registerSpellDataDamageProc(SpellDataProc{Name: "Test Killing Blow", ItemID: tc.trinketID,
				TriggerSpellID: 7711, BuffSpellID: 7712})

			items := testHands(&proto.ItemSpec{}, &proto.ItemSpec{})
			items[proto.ItemSlot_ItemSlotTrinket1] = &proto.ItemSpec{Id: tc.trinketID}
			sim := newTestCasterSimAgainst(items, nil, tc.targetCount)
			caster := sim.Raid.Parties[0].Players[0].(*testCaster)
			dying := sim.Encounter.AllTargets[0]

			dying.Die(sim, 1)

			damage := mustGetSpell(t, caster, 7712)
			if casts := damage.SpellMetrics[dying.UnitIndex].Casts; casts != 0 {
				t.Errorf("the proc cast %d times on the dying target, want none", casts)
			}
			cast := int32(0)
			for _, target := range sim.Encounter.AllTargets[1:] {
				cast += damage.SpellMetrics[target.UnitIndex].Casts
			}
			if cast != tc.want {
				t.Errorf("the proc cast %d times on the targets still up, want %d", cast, tc.want)
			}
		})
	}
}

// Blazing Rapier 12777's chance-on-hit Blaze 16898 deals Fire every period for its duration and
// nothing up front, and states no rate. Rolled at 100% for the test, the damage proc puts the damage
// over time on the target the swing hit, and it ticks the row's duration over its period.
func TestWeaponProcDealingDamageOverTimeAloneAppliesIt(t *testing.T) {
	const weaponID, blaze int32 = 991605, 16898
	editRow(t, blaze, alwaysProcs)
	row := spelldata.MustFind(blaze)
	duration := row.Duration()
	wantTicks := int32(duration / row.PeriodicDamageEffect().Period())
	sim, caster, target := newSlowProcSim(t, weaponID, registerSpellDataDamageProc, blaze, nil)

	start := sim.CurrentTime
	landMainHandSwing(sim, caster, target)
	damage := caster.GetSpell(core.ActionID{SpellID: blaze})
	if damage == nil || !damage.Dot(target).IsActive() {
		t.Fatalf("the swing put no %d damage over time on the target", blaze)
	}

	stepPast(t, sim, start+duration+time.Millisecond)
	if damage.Dot(target).IsActive() {
		t.Errorf("the damage over time is still up after the row's %v", duration)
	}
	metrics := damage.SpellMetrics[target.UnitIndex]
	if ticks := metrics.Ticks + metrics.ResistedTicks + metrics.CritTicks + metrics.ResistedCritTicks; wantTicks < 2 || ticks != wantTicks {
		t.Errorf("the damage over time ticked %d times, want the row's %d", ticks, wantTicks)
	}
}

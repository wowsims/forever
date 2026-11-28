package shared

import (
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Vengeance 20049 states Can Proc From Procs, and its harmful spell bit hears a crit of any school.
const vengeanceTalent int32 = 20049

// Storm Gauntlets' 16615 casts Lightning Strike 16614 off each melee hit, and Blazefury Medallion's 7711
// casts Fire Strike 7712 the same way. Their hits are procs, so an aura hears them only where it carries
// Can Proc From Procs.
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

// An enchant's damage proc whose trigger states a factor against some targets, The Chief's Enforcer's
// 1294339 against Trolls, lists the factor on the stats sheet under the trigger's spell: the enchant's
// own id names no spell. With an item swap set it lists the factor while the worn weapon carries the
// enchant, and nothing while only the item swap set's weapon does.
func TestEnchantDamageProcListsItsVersusFactorByItsSpell(t *testing.T) {
	const swordID, enchantID, strike int32 = 991690, 991691, 1294339
	want, ok := spelldata.MustFind(strike).VersusBonus()
	if !ok {
		t.Fatalf("%d states no versus factor", strike)
	}
	core.AddToDatabase(&proto.SimDatabase{
		Items:    []*proto.SimItem{testOneHander(swordID)},
		Enchants: []*proto.SimEnchant{{EffectId: enchantID, Name: "Test Strike Enchant", Type: proto.ItemType_ItemTypeWeapon}},
	})
	registerSpellDataDamageProc(SpellDataProc{Name: "Test Strike Enchant", EnchantID: enchantID, TriggerSpellID: strike,
		BuffSpellID: strike, IsWeaponProc: true, Rate: spelldata.Chance(1)})

	_, caster := newWielderSim(testHands(&proto.ItemSpec{Id: swordID, Enchant: enchantID}, &proto.ItemSpec{}))
	playerStats := &proto.PlayerStats{}
	caster.FillPlayerStats(playerStats)
	if sheet := playerStats.VersusEffects; len(sheet) != 1 || core.ProtoToActionID(sheet[0].Source) != (core.ActionID{SpellID: strike}) ||
		sheet[0].ItemId != 0 || sheet[0].Multiplier != want.Multiplier || !slices.Equal(sheet[0].CreatureFamilies, want.Families) {
		t.Errorf("the stats sheet lists %v, want %d's factor of %v against %v", sheet, strike, want.Multiplier, want.Families)
	}

	plain, enchanted := testHands(&proto.ItemSpec{Id: swordID}, &proto.ItemSpec{}), testHands(&proto.ItemSpec{Id: swordID, Enchant: enchantID}, &proto.ItemSpec{})
	swapSheet := func(worn []*proto.ItemSpec) []*proto.VersusEffect {
		playerStats := &proto.PlayerStats{}
		newTestCasterSim(worn, enchanted).Raid.Parties[0].Players[0].GetCharacter().FillPlayerStats(playerStats)
		return playerStats.VersusEffects
	}
	if sheet := swapSheet(enchanted); len(sheet) != 1 || core.ProtoToActionID(sheet[0].Source) != (core.ActionID{SpellID: strike}) {
		t.Errorf("with the enchant on the worn weapon and the item swap set's the stats sheet lists %v, want %d's factor", sheet, strike)
	}
	if sheet := swapSheet(plain); len(sheet) != 0 {
		t.Errorf("with the enchant only on the item swap set's weapon the stats sheet lists %v, want none", sheet)
	}
}

// Windfury Weapon's passive 439431 hears melee auto attacks and abilities, without Can Proc From Procs.
const windfuryWeaponPassive int32 = 439431

// The deferred second roll of each strike the swings so far cast.
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

// Weapon procs' strikes on the melee table: Hameya's Slayer 15814's Rend 16406, a bleed with no hit,
// Hammer of Expertise 8708's Forging Strike 1317432, a hit, and Mutilator 12469's Lacerations 10370,
// a debuff alone.
const (
	rendStrike    int32 = 16406
	forgingStrike int32 = 1317432
	lacerations   int32 = 10370
)

// A neck whose listener is built as a weapon enchant's or an oil's, counting the spell's strikes it hears.
func strikeListenerNeck(neckID int32, name string, row int32, heard *int) *proto.ItemSpec {
	withAuraItem(neckID, proto.ItemType_ItemTypeNeck)
	core.NewItemEffect(neckID, func(agent core.Agent) {
		character := agent.GetCharacter()
		listener := spelldata.ProcTrigger(character, spelldata.MustFind(row),
			func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
				if spell.ActionID.SpellID == row {
					*heard++
				}
			}, spelldata.WeaponProc(), spelldata.Chance(1))
		listener.Name = name
		listener.TriggerImmediately = true
		character.MakeProcTriggerAura(listener)
	})
	return &proto.ItemSpec{Id: neckID}
}

// A strike a main-hand swing cast is a main-hand special: the main hand's weapon procs hear it and its own
// weapon rolls again. A strike an off-hand swing cast is heard by no weapon proc on either hand. Auto
// attacks' and Windfury Weapon's (Forging Strike only, it needs damage) listeners hear neither.
func TestWeaponProcStrikeIsHeardOnlyFromTheMainHand(t *testing.T) {
	for _, tc := range []struct {
		name                     string
		weaponID, neckID, strike int32
		windfury                 bool
	}{
		{"Rend", 991606, 991607, rendStrike, false},
		{"Forging Strike", 991608, 991609, forgingStrike, true},
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

// Forging Strike in the main hand and Rend in the off hand, on a wielder that dual wields: a strike a
// main-hand swing cast rolls the main-hand weapon's proc and never the off-hand one's, and one an off-hand
// swing cast rolls neither.
func TestWeaponProcStrikeRollsOnlyTheMainHandWeapon(t *testing.T) {
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
			editRow(t, forgingStrike, func(s *spelldata.Spell) { s.RPPM = 18 })
			editRow(t, rendStrike, tc.offHandRate)
			core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{testOneHander(tc.mainHandID), testOneHander(tc.offHandID)}})
			registerSpellDataDamageProc(SpellDataProc{Name: "Test Forging Strike Weapon", ItemID: tc.mainHandID,
				TriggerSpellID: forgingStrike, BuffSpellID: forgingStrike, IsWeaponProc: true})
			registerSpellDataDamageProc(SpellDataProc{Name: "Test Rend Weapon", ItemID: tc.offHandID,
				TriggerSpellID: rendStrike, BuffSpellID: rendStrike, IsWeaponProc: true})

			sim := newTestCasterSimAgainst(testHands(&proto.ItemSpec{Id: tc.mainHandID}, &proto.ItemSpec{Id: tc.offHandID}), nil, 1,
				func(player *proto.Player) { player.Spec = &proto.Player_EnhancementShaman{} })
			caster := sim.Raid.Parties[0].Players[0].(*testCaster)
			caster.AutoAttacks.CancelAutoSwing(sim)
			target := sim.Encounter.ActiveTargetUnits[0]
			strikeSpell, rendSpell := mustGetSpell(t, caster, forgingStrike), mustGetSpell(t, caster, rendStrike)
			casts := func(spell *core.Spell) int32 { return spell.SpellMetrics[target.UnitIndex].Casts }

			for range swings {
				landSwing(sim, caster, target, core.ProcMaskMeleeOHAuto)
			}
			settleChains(t, sim)
			if n := casts(rendSpell); n == 0 || n > swings {
				t.Errorf("%d off-hand swings cast %d Rends, want some and at most one each: a Rend never rolls the off-hand weapon", swings, n)
			}
			if n := casts(strikeSpell); n != 0 {
				t.Errorf("Rends off-hand swings cast rolled the main-hand weapon's Forging Strike %d times, want none", n)
			}

			rends, strikes := casts(rendSpell), casts(strikeSpell)
			for range swings {
				landSwing(sim, caster, target, core.ProcMaskMeleeMHAuto)
			}
			settleChains(t, sim)
			if n := casts(rendSpell) - rends; n != 0 {
				t.Errorf("%d main-hand swings and the Forging Strikes they cast cast %d Rends, want none", swings, n)
			}
			if n := casts(strikeSpell) - strikes; n <= swings {
				t.Errorf("%d main-hand swings cast %d Forging Strikes, want more: a Forging Strike rolls the main-hand weapon again", swings, n)
			}
		})
	}
}

// Weapon rolls in EV-0145.
func TestWeaponProcStrikeAWhiteSwingCastRollsItsWeaponOnce(t *testing.T) {
	const weaponID int32 = 991616
	const batches = 5
	editRow(t, forgingStrike, alwaysProcs)
	sim, caster, target := newSlowProcSim(t, weaponID, registerSpellDataDamageProc, forgingStrike, nil)
	caster.AddStatsDynamic(sim, stats.Stats{stats.PhysicalHitPercent: 100, stats.ExpertisePercent: 100})
	strike := mustGetSpell(t, caster, forgingStrike)
	metrics := &strike.SpellMetrics[target.UnitIndex]

	landMainHandSwing(sim, caster, target)
	if metrics.Casts != 1 {
		t.Fatalf("the swing cast %d strikes at once, want 1: the second roll waits a batch", metrics.Casts)
	}
	stepPast(t, sim, sim.CurrentTime+core.SpellBatchWindow+time.Millisecond)
	if metrics.Casts != 2 {
		t.Fatalf("one batch after the swing %d strikes were cast, want 2", metrics.Casts)
	}
	stepPast(t, sim, sim.CurrentTime+batches*core.SpellBatchWindow+time.Millisecond)
	if metrics.Casts != 2 {
		t.Errorf("%d batches later %d strikes were cast, want 2: the second strike rolls nothing", batches, metrics.Casts)
	}
	if avoided := metrics.Misses + metrics.Dodges + metrics.Parries; avoided != 0 {
		t.Errorf("%d strikes were avoided, want every one to land", avoided)
	}
}

// Weapon rolls in EV-0145.
func TestWeaponProcStrikeRollsItsWeaponOnlyOffAMainHandWhiteSwing(t *testing.T) {
	for _, tc := range []struct {
		name     string
		row      int32
		weaponID int32
		neckID   int32
		register func(SpellDataProc)
	}{
		{"Forging Strike", forgingStrike, 991714, 991716, registerSpellDataDamageProc},
		{"Lacerations", lacerations, 991715, 991717, registerSpellDataAuraProc},
	} {
		t.Run(tc.name, func(t *testing.T) {
			editRow(t, tc.row, func(s *spelldata.Spell) {
				alwaysProcs(s)
				s.RPPM = 0
			})
			heard := 0
			neck := strikeListenerNeck(tc.neckID, "Test Enchant", tc.row, &heard)
			sim, caster, target := newSlowProcSim(t, tc.weaponID, tc.register, tc.row, neck)
			caster.AddStatsDynamic(sim, stats.Stats{stats.PhysicalHitPercent: 100, stats.ExpertisePercent: 100})
			strike := mustGetSpell(t, caster, tc.row)
			metrics := &strike.SpellMetrics[target.UnitIndex]

			whiteSwing := &core.Spell{ProcMask: core.ProcMaskMeleeMHAuto}
			for _, hit := range []struct {
				name      string
				cast      *core.Spell
				paw       bool
				want      int32
				wantHeard int
			}{
				{"a main-hand white swing", whiteSwing, false, 2, 1},
				{"a main-hand ability", &core.Spell{ProcMask: core.ProcMaskMeleeMHSpecial, DefenseType: core.DefenseTypeMelee}, false, 1, 0},
				{"a hit that counts as a swing", &core.Spell{ProcMask: core.ProcMaskMeleeMHSpecial, DefenseType: core.DefenseTypeMelee,
					Flags: core.SpellFlagCountsAsSwing}, false, 1, 0},
				{"a paw's white swing", whiteSwing, true, 1, 0},
			} {
				before, heardBefore := metrics.Casts, heard
				caster.PseudoStats.FeralCombatEnabled = hit.paw
				caster.OnSpellHitDealt(sim, hit.cast, &core.SpellResult{Target: target, Outcome: core.OutcomeHit, Damage: 100})
				settleChains(t, sim)
				if got := metrics.Casts - before; got != hit.want {
					t.Errorf("%s cast %d strikes, want %d", hit.name, got, hit.want)
				}
				if got := heard - heardBefore; got != hit.wantHeard {
					t.Errorf("a weapon enchant heard %d of the strikes %s cast, want %d", got, hit.name, hit.wantHeard)
				}
			}
		})
	}
}

// Strike cooldown in EV-0145.
func TestWeaponProcStrikeOnItsOwnCooldownIsNotCastAgain(t *testing.T) {
	const weaponID int32 = 991718
	const cooldown = 3 * time.Second
	editRow(t, forgingStrike, func(s *spelldata.Spell) {
		alwaysProcs(s)
		s.CooldownMs = int32(cooldown / time.Millisecond)
	})
	sim, caster, target := newSlowProcSim(t, weaponID, registerSpellDataDamageProc, forgingStrike, nil)
	metrics := &mustGetSpell(t, caster, forgingStrike).SpellMetrics[target.UnitIndex]
	start := sim.CurrentTime

	for _, swing := range []struct {
		name string
		at   time.Duration
		want int32
	}{
		{"the first swing", 0, 1},
		{"a swing inside the cooldown", cooldown / 2, 0},
		{"the first swing after the cooldown", cooldown + time.Millisecond, 1},
	} {
		stepPast(t, sim, start+swing.at)
		before := metrics.Casts
		landMainHandSwing(sim, caster, target)
		settleChains(t, sim)
		if got := metrics.Casts - before; got != swing.want {
			t.Errorf("%s cast %d strikes, want %d", swing.name, got, swing.want)
		}
	}
}

// Every landed Rend puts its bleed on the target, a Rend that its own strike chained included, and
// no weapon proc hears the bleed's ticks. The bleed is short enough for every swing to fit the fight.
func TestWeaponProcStrikeChainAppliesEachLandedBleed(t *testing.T) {
	const weaponID, neckID, rend int32 = 991612, 991613, rendStrike
	const swings = 10
	editRow(t, rend, func(s *spelldata.Spell) {
		s.ProcChance, s.ProcChanceSource = 90, spelldata.ProcChanceColumn
		s.DurationMs = 12000
	})
	duration := spelldata.MustFind(rend).Duration()
	landed := 0
	neck := strikeListenerNeck(neckID, "Test Oil", rend, &landed)
	sim, caster, target := newSlowProcSim(t, weaponID, registerSpellDataDamageProc, rend, neck)
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

// A test one-hander carrying the spell in the off hand. The plain one in the main hand keeps it there: a
// one-hander equipped with the main hand empty goes to the main hand.
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

// The Rend an off-hand swing casts puts its bleed on the target and rolls its own weapon no more.
func TestOffHandWeaponProcStrikeAppliesItsBleed(t *testing.T) {
	const weaponID, rend int32 = 991617, rendStrike
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

// Forging Strike in the off hand, given a missile speed: its own weapon and a main-hand weapon proc hear
// the strike's landing after its flight no more than they would an instant strike.
func TestOffHandWeaponProcStrikeInFlightIsHeardByNoWeaponProc(t *testing.T) {
	const weaponID, neckID int32 = 991618, 991619
	editRow(t, forgingStrike, func(s *spelldata.Spell) {
		alwaysProcs(s)
		s.Speed = 20
	})
	withAuraItem(neckID, proto.ItemType_ItemTypeNeck)
	heard := 0
	core.NewItemEffect(neckID, func(agent core.Agent) {
		agent.GetCharacter().MakeProcTriggerAura(weaponProcListener("Test Oil", core.ProcMaskMeleeMH,
			func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
				if spell.ActionID.SpellID == forgingStrike {
					heard++
				}
			}))
	})
	sim, caster, target := newOffHandProcSim(t, weaponID, forgingStrike, &proto.ItemSpec{Id: neckID})
	strike := mustGetSpell(t, caster, forgingStrike)
	caster.DistanceFromTarget = 20
	metrics := &strike.SpellMetrics[target.UnitIndex]

	landSwing(sim, caster, target, core.ProcMaskMeleeOHAuto)
	if caster.SpellsInFlight[strike] != 1 || metrics.TotalDamage != 0 {
		t.Fatalf("the Forging Strike is not in flight after the swing")
	}
	settleChains(t, sim)
	if caster.SpellsInFlight[strike] != 0 || metrics.TotalDamage == 0 {
		t.Fatalf("the Forging Strike did not land")
	}
	if metrics.Casts != 1 {
		t.Errorf("the off-hand swing cast %d Forging Strikes, want 1: its landing rolled its own weapon", metrics.Casts)
	}
	if heard != 0 {
		t.Errorf("the main-hand weapon proc heard %d Forging Strikes, want none", heard)
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

// A damage proc whose trigger hears a killing blow alone: the kill passes the proc no spell and names the
// dying target, so its damage goes to the next target still up, and with none left it casts nothing.
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

// Blaze 16898 deals Fire every period and nothing up front: the damage proc puts the damage over time on
// the target the swing hit, and it ticks the spell's duration over its period.
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

// A damage proc's restore takes its trigger's factor against the hit's target, kept for the stats sheet
// as a restore under the item: 27860's mana off a 7711 that states three times as much against Plants,
// an override no row carries, set on the store's row for the test.
func TestDamageProcRestoresTheTriggersFactorAgainstItsTargets(t *testing.T) {
	const trinketID, engulfingShadows int32 = 991633, 27860
	trigger := spelldata.MustFind(7711)
	if len(trigger.VersusMobTypes) > 0 || len(trigger.VersusFamilies) > 0 {
		t.Fatalf("7711 states a factor against %v and %v already, nothing to set", trigger.VersusMobTypes, trigger.VersusFamilies)
	}
	trigger.VersusFamilies, trigger.VersusMultiplier = []int32{int32(proto.CreatureFamily_CreatureFamilyPlant)}, 3
	t.Cleanup(func() { trigger.VersusFamilies, trigger.VersusMultiplier = nil, 0 })
	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{{Id: trinketID, Name: "Test Plant Shadows",
		Type: proto.ItemType_ItemTypeTrinket, ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}}}})
	registerSpellDataDamageProc(SpellDataProc{Name: "Test Plant Shadows", ItemID: trinketID,
		TriggerSpellID: 7711, BuffSpellID: engulfingShadows})
	mana := spelldata.MustFind(engulfingShadows).EnergizeEffect().Average()
	needAmount(t, "27860's mana", mana)

	for _, tc := range []struct {
		family proto.CreatureFamily
		want   float64
	}{
		{proto.CreatureFamily_CreatureFamilyPlant, 3 * mana},
		{proto.CreatureFamily_CreatureFamilyUnknown, mana},
	} {
		t.Run(tc.family.String(), func(t *testing.T) {
			items := testHands(&proto.ItemSpec{}, &proto.ItemSpec{})
			items[proto.ItemSlot_ItemSlotTrinket1] = &proto.ItemSpec{Id: trinketID}
			sim := newTestCasterSimAgainst(items, nil, 1, func(player *proto.Player) {
				player.Class, player.Race, player.Spec = proto.Class_ClassMage, proto.Race_RaceGnome, &proto.Player_Mage{}
			})
			caster := sim.Raid.Parties[0].Players[0].(*testCaster)
			target := sim.Encounter.ActiveTargetUnits[0]
			target.CreatureFamily = tc.family

			landMainHandSwing(sim, caster, target)
			settleChains(t, sim)

			var gain float64
			for _, r := range caster.Metrics.ToProto().Resources {
				if r.Id.GetSpellId() == engulfingShadows {
					gain = r.Gain
				}
			}
			if gain != tc.want {
				t.Errorf("the proc restored %v mana, want %v", gain, tc.want)
			}
			playerStats := &proto.PlayerStats{}
			caster.FillPlayerStats(playerStats)
			if !slices.ContainsFunc(playerStats.VersusEffects, func(effect *proto.VersusEffect) bool {
				return effect.ItemId == trinketID && effect.Kind == proto.VersusEffectKind_VersusEffectKindRestore && effect.Multiplier == 3
			}) {
				t.Errorf("the stats sheet lists %v, want the item's x3 restore", playerStats.VersusEffects)
			}
		})
	}
}

// Engulfing Shadows 27860 deals Shadow damage to the enemy and restores mana to its caster, cast here
// off Blazefury Medallion 17111's 7711 on each melee hit. The wearer gains the row's mana with the
// proc, whether the damage lands or misses.
func TestDamageProcRestoresTheWearersManaLandedOrNot(t *testing.T) {
	const trinketID, engulfingShadows int32 = 991632, 27860
	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{{Id: trinketID, Name: "Test Engulfing Shadows",
		Type: proto.ItemType_ItemTypeTrinket, ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}}}})
	registerSpellDataDamageProc(SpellDataProc{Name: "Test Engulfing Shadows", ItemID: trinketID,
		TriggerSpellID: 7711, BuffSpellID: engulfingShadows})
	want := spelldata.MustFind(engulfingShadows).EnergizeEffect().Average()
	needAmount(t, "27860's mana", want)

	for _, tc := range []struct {
		name   string
		hit    float64
		landed bool
	}{
		{"landed", 100, true},
		{"missed", -100, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			items := testHands(&proto.ItemSpec{}, &proto.ItemSpec{})
			items[proto.ItemSlot_ItemSlotTrinket1] = &proto.ItemSpec{Id: trinketID}
			sim := newTestCasterSimAgainst(items, nil, 1, func(player *proto.Player) {
				player.Class, player.Race, player.Spec = proto.Class_ClassMage, proto.Race_RaceGnome, &proto.Player_Mage{}
			})
			caster := sim.Raid.Parties[0].Players[0].(*testCaster)
			target := sim.Encounter.ActiveTargetUnits[0]
			caster.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: tc.hit})

			landMainHandSwing(sim, caster, target)
			settleChains(t, sim)

			metrics := mustGetSpell(t, caster, engulfingShadows).SpellMetrics[target.UnitIndex]
			if landed := metrics.Misses == 0; metrics.Casts != 1 || landed != tc.landed {
				t.Fatalf("the swing cast %d with %d misses, want one cast that landed: %v", metrics.Casts, metrics.Misses, tc.landed)
			}
			var gain float64
			var events int32
			for _, r := range caster.Metrics.ToProto().Resources {
				if r.Id.GetSpellId() == engulfingShadows {
					gain, events = r.Gain, r.Events
				}
			}
			if events != 1 || gain != want {
				t.Errorf("the proc restored %v mana over %d gains, want %v once", gain, events, want)
			}
		})
	}
}

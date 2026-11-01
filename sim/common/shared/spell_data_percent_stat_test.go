package shared

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

const (
	insightWeaponID  int32 = 990951
	insightEnchantID int32 = 990952
	testBoltID       int32 = 990953
)

func init() {
	core.RegisterAgentFactory(
		proto.Player_ElementalShaman{},
		proto.Spec_SpecElementalShaman,
		func(character *core.Character, _ *proto.Player, _ *proto.Raid) core.Agent {
			return &testCaster{Character: *character}
		},
		func(player *proto.Player, spec any) {
			player.Spec = spec.(*proto.Player_ElementalShaman)
		},
	)
}

type testCaster struct {
	core.Character
	bolt *core.Spell
}

func (c *testCaster) GetCharacter() *core.Character       { return &c.Character }
func (c *testCaster) ApplyTalents()                       {}
func (c *testCaster) Reset(_ *core.Simulation)            {}
func (c *testCaster) OnGCDReady(_ *core.Simulation)       {}
func (c *testCaster) OnEncounterStart(_ *core.Simulation) {}

func (c *testCaster) Initialize() {
	c.bolt = c.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: testBoltID},
		SpellSchool:      core.SpellSchoolArcane,
		ProcMask:         core.ProcMaskSpellDamage,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeAlwaysHit)
		},
	})
}

func testHands(mainHand, offHand *proto.ItemSpec) []*proto.ItemSpec {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotOffHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = mainHand
	items[proto.ItemSlot_ItemSlotOffHand] = offHand
	return items
}

func newTestCasterSim(equipped, swapped []*proto.ItemSpec) *core.Simulation {
	return newTestCasterSimAgainst(equipped, swapped, 1)
}

func newTestCasterSimAgainst(equipped, swapped []*proto.ItemSpec, targetCount int, playerOpts ...func(*proto.Player)) *core.Simulation {
	targets := make([]*proto.Target, targetCount)
	for i := range targets {
		targets[i] = &proto.Target{Name: "target", Level: 60, MobType: proto.MobType_MobTypeDemon}
	}

	player := &proto.Player{
		Name:      "Caster",
		Class:     proto.Class_ClassShaman,
		Race:      proto.Race_RaceTroll,
		Buffs:     &proto.IndividualBuffs{},
		Spec:      &proto.Player_ElementalShaman{},
		Equipment: &proto.EquipmentSpec{Items: equipped},
	}
	if swapped != nil {
		player.EnableItemSwap = true
		player.ItemSwap = &proto.ItemSwap{Items: swapped}
	}
	for _, opt := range playerOpts {
		opt(player)
	}

	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{player},
			Buffs:   &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{
			Targets:  targets,
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim
}

// A shield enchant's buff drops when an item swap takes the shield out of the off hand, and stays up
// when the swap moves only the main hand, whose weapon enchant's buff drops either way.
func TestShieldEnchantBuffDropsWithItsShield(t *testing.T) {
	const enchantedWeaponID, plainWeaponID, enchantedShieldID, plainShieldID int32 = 990961, 990962, 990963, 990964
	const weaponEnchantID, shieldEnchantID int32 = 990965, 990966

	shield := func(id int32) *proto.SimItem {
		return &proto.SimItem{Id: id, Name: "Test Shield", Type: proto.ItemType_ItemTypeWeapon,
			WeaponType: proto.WeaponType_WeaponTypeShield, HandType: proto.HandType_HandTypeOffHand,
			ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}}
	}
	core.AddToDatabase(&proto.SimDatabase{
		Items: []*proto.SimItem{testOneHander(enchantedWeaponID), testOneHander(plainWeaponID),
			shield(enchantedShieldID), shield(plainShieldID)},
		Enchants: []*proto.SimEnchant{
			{EffectId: weaponEnchantID, Name: "Test Weapon Enchant", Type: proto.ItemType_ItemTypeWeapon},
			{EffectId: shieldEnchantID, Name: "Test Shield Enchant", Type: proto.ItemType_ItemTypeWeapon,
				EnchantType: proto.EnchantType_EnchantTypeShield},
		},
	})
	for name, id := range map[string]int32{"Test Weapon Enchant": weaponEnchantID, "Test Shield Enchant": shieldEnchantID} {
		registerSpellDataAuraProc(SpellDataProc{Name: name, EnchantID: id, TriggerSpellID: 1248758, BuffSpellID: 1299796})
	}

	for _, tc := range []struct {
		name         string
		swapOffHand  *proto.ItemSpec
		wantOffHand  int32
		wantShieldUp bool
	}{
		{"both hands swapped", &proto.ItemSpec{Id: plainShieldID}, plainShieldID, false},
		{"only the main hand swapped", &proto.ItemSpec{}, enchantedShieldID, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim := newTestCasterSim(
				testHands(&proto.ItemSpec{Id: enchantedWeaponID, Enchant: weaponEnchantID},
					&proto.ItemSpec{Id: enchantedShieldID, Enchant: shieldEnchantID}),
				testHands(&proto.ItemSpec{Id: plainWeaponID}, tc.swapOffHand))
			caster := sim.Raid.Parties[0].Players[0].(*testCaster)

			weaponBuff := caster.GetAura("Test Weapon Enchant Proc")
			shieldBuff := caster.GetAura("Test Shield Enchant Proc")
			if weaponBuff == nil || shieldBuff == nil {
				t.Fatalf("buffs registered: weapon %v, shield %v", weaponBuff != nil, shieldBuff != nil)
			}
			weaponBuff.Activate(sim)
			shieldBuff.Activate(sim)

			caster.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)
			if caster.MainHand().ID != plainWeaponID || caster.OffHand().ID != tc.wantOffHand {
				t.Fatalf("hands after the swap = %d / %d, want %d / %d",
					caster.MainHand().ID, caster.OffHand().ID, plainWeaponID, tc.wantOffHand)
			}
			if weaponBuff.IsActive() {
				t.Error("the weapon enchant's buff is still up after its weapon was swapped out")
			}
			if shieldBuff.IsActive() != tc.wantShieldUp {
				t.Errorf("shield enchant's buff up = %v after the swap, want %v", shieldBuff.IsActive(), tc.wantShieldUp)
			}
		})
	}
}

// Insight 8216's rows on a weapon enchant: trigger 1248758 hears spell damage and heals at its stated
// chance with a lockout, and its buff 1299796 multiplies Spirit for its duration, on top of whatever
// Spirit the caster gains while it is up.
func TestInsightMultipliesSpirit(t *testing.T) {
	trigger, buff := spelldata.MustFind(1248758), spelldata.MustFind(1299796)
	percent := buff.FirstAura(dbcenums.A_MOD_TOTAL_STAT_PERCENTAGE).Mean()
	needAmount(t, "1299796's Spirit share", percent)
	multiplier := 1 + percent/100
	weapon := testOneHander(insightWeaponID)
	core.AddToDatabase(&proto.SimDatabase{
		Items: []*proto.SimItem{weapon},
		Enchants: []*proto.SimEnchant{{EffectId: insightEnchantID, Name: "Test Insight",
			Type: proto.ItemType_ItemTypeWeapon}},
	})
	registerSpellDataAuraProc(SpellDataProc{Name: "Test Insight", EnchantID: insightEnchantID,
		TriggerSpellID: 1248758, BuffSpellID: 1299796})

	cfg := SpellDataProc{Name: "Test Insight", EnchantID: insightEnchantID, TriggerSpellID: 1248758,
		BuffSpellID: 1299796}
	listener := spellDataProcListener(newTestAgent().character, cfg, cfg.effectSource(), trigger)
	if want := core.CallbackOnSpellHitDealt | core.CallbackOnHealDealt; listener.Callback != want {
		t.Errorf("callback = %v, want %v", listener.Callback, want)
	}
	if want := core.ProcMaskSpellDamage | core.ProcMaskSpellHealing; listener.ProcMask != want {
		t.Errorf("proc mask = %v, want %v", listener.ProcMask, want)
	}
	if want := trigger.StatedChance(); listener.ProcChance != want || listener.DPM != nil {
		t.Errorf("chance = %v with manager %v, want the stated %v and no rate", listener.ProcChance, listener.DPM, want)
	}

	sim := newTestCasterSim(testHands(&proto.ItemSpec{Id: insightWeaponID, Enchant: insightEnchantID}, &proto.ItemSpec{}), nil)
	caster := sim.Raid.Parties[0].Players[0].(*testCaster)
	target := sim.Encounter.ActiveTargetUnits[0]
	insight := caster.GetAura("Test Insight Proc")
	if insight == nil {
		t.Fatal("no Insight buff on the caster")
	}
	if insight.Duration != buff.Duration() {
		t.Errorf("buff duration = %v, want 1299796's %v", insight.Duration, buff.Duration())
	}
	if procBuffs := caster.GetMatchingItemProcAuras([]stats.Stat{stats.Spirit}, 0); len(procBuffs) != 1 || procBuffs[0].Aura != insight {
		t.Errorf("Spirit proc buffs %v, want Insight's", procBuffs)
	}

	var heard []stats.Stats
	caster.OnTemporaryStatsChanges = append(caster.OnTemporaryStatsChanges, func(_ *core.Simulation, aura *core.Aura, change stats.Stats) {
		if aura == insight {
			heard = append(heard, change)
		}
	})

	spirit := caster.GetStat(stats.Spirit)
	if spirit <= 0 {
		t.Fatalf("Spirit is %v before Insight, nothing for a multiplier to show on", spirit)
	}

	// A proc's handler runs one spell batch window after the hit, so each cast is followed by a step.
	castAndStep := func() {
		caster.bolt.Cast(sim, target)
		sim.Step()
	}

	casts := 0
	for !insight.IsActive() {
		if casts++; casts > 100 {
			t.Fatal("100 spell hits and Insight never procced")
		}
		castAndStep()
	}
	procTime := sim.CurrentTime
	if got := caster.GetStat(stats.Spirit); !core.WithinToleranceFloat64(multiplier*spirit, got, 1e-6) {
		t.Errorf("Spirit with Insight = %v, want %v times %v", got, multiplier, spirit)
	}

	caster.AddStatsDynamic(sim, stats.Stats{stats.Spirit: 50})
	if got := caster.GetStat(stats.Spirit); !core.WithinToleranceFloat64(multiplier*(spirit+50), got, 1e-6) {
		t.Errorf("Spirit with Insight after 50 more = %v, want %v times %v", got, multiplier, spirit+50)
	}
	caster.AddStatsDynamic(sim, stats.Stats{stats.Spirit: -50})

	insight.Deactivate(sim)
	if got := caster.GetStat(stats.Spirit); !core.WithinToleranceFloat64(spirit, got, 1e-6) {
		t.Errorf("Spirit after Insight = %v, want %v", got, spirit)
	}
	gained := (multiplier - 1) * spirit
	if len(heard) != 2 || !core.WithinToleranceFloat64(gained, heard[0][stats.Spirit], 1e-6) ||
		!core.WithinToleranceFloat64(-gained, heard[1][stats.Spirit], 1e-6) {
		t.Errorf("temporary stats heard %v, want +%v Spirit on gain and -%v on expire", heard, gained, gained)
	}

	for range 100 {
		castAndStep()
	}
	if sim.CurrentTime-procTime >= trigger.ICD() {
		t.Fatalf("the casts ran %v past the proc, not inside the lockout", sim.CurrentTime-procTime)
	}
	if insight.IsActive() {
		t.Errorf("Insight procced again inside its %v lockout", trigger.ICD())
	}
}

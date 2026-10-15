package shared

import (
	"math"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Eternal Power, the Dormant Heart of the Mountain proc: 1249118 carries the proc flags and 1249119
// is the buff it applies. The item IDs below are the test's own, so that registering them cannot
// collide with a real effect.
const (
	healProcTrigger int32 = 1249118
	healProcBuff    int32 = 1249119
)

// core skips an effect whose item the game files do not ship, which every id below is: the test's own
// ids are chosen to be absent from the shipped database so that registering them cannot collide with
// a real effect. Putting them in the database first is what lets the registration run under the
// with_db build tag as it does without it.
func withTestItems(ids ...int32) {
	items := make([]*proto.SimItem, 0, len(ids))
	for _, id := range ids {
		items = append(items, &proto.SimItem{Id: id, Name: "Test Trinket"})
	}
	core.AddToDatabase(&proto.SimDatabase{Items: items})
}

func withTestEnchant(effectID int32) {
	core.AddToDatabase(&proto.SimDatabase{
		Enchants: []*proto.SimEnchant{{EffectId: effectID, Name: "Test Enchant"}},
	})
}

func TestSpellDataProcRegistersEveryVariant(t *testing.T) {
	withTestItems(990001, 990002)
	NewSpellDataProc(SpellDataProc{TriggerSpellID: healProcTrigger, BuffSpellID: healProcBuff},
		[]ItemVariant{
			{ItemID: 990001, ItemName: "Reissued Test Trinket"},
			{ItemID: 990002, ItemName: "Test Trinket"},
		})

	if !core.HasItemEffect(990001) || !core.HasItemEffect(990002) {
		t.Error("a variant was left unregistered")
	}
}

// A row that names no hits the sim hears would register a listener that can never fire. 1249119 is
// the buff half of the pair above, here stripped of any proc flags and hint.
func TestSpellDataProcSkipsARowWithNoListener(t *testing.T) {
	withTestItems(990003)
	editRow(t, healProcBuff, func(s *spelldata.Spell) { s.ProcFlags, s.ProcHint = [2]uint32{}, 0 })
	NewSpellDataProc(SpellDataProc{TriggerSpellID: healProcBuff},
		[]ItemVariant{{ItemID: 990003, ItemName: "Listenerless Test Trinket"}})

	if core.HasItemEffect(990003) {
		t.Error("an effect whose row states no listener was registered anyway")
	}
}

// An enchant states its own name and registers through the enchant registry instead.
func TestSpellDataProcRegistersAnEnchant(t *testing.T) {
	withTestEnchant(990004)
	NewSpellDataProc(SpellDataProc{
		Name:           "Test Enchant",
		EnchantID:      990004,
		TriggerSpellID: healProcTrigger,
		BuffSpellID:    healProcBuff,
	}, nil)

	if !core.HasEnchantEffect(990004) {
		t.Error("an enchant proc with no variants was left unregistered")
	}
	if core.HasItemEffect(990004) {
		t.Error("an enchant proc was registered as an item effect")
	}
}

// A row pair in the shape the charge rules read: a trigger that grants the buff and a buff that
// states its charges and what spends them.
func chargeRows() (*spelldata.Spell, *spelldata.Spell) {
	trigger := &spelldata.Spell{
		ID: 990100, Name: "Charge Trigger", DurationMs: 12000, ProcChance: 100,
		ProcChanceSource: spelldata.ProcChanceAlways,
		ProcFlags:        [2]uint32{0: dbcenums.PROC_FLAG_DEAL_MELEE_ABILITY},
	}
	buff := &spelldata.Spell{
		ID: 990101, Name: "Charge Buff", DurationMs: 12000, ProcChance: 100, ProcCharges: 2,
		ProcChanceSource: spelldata.ProcChanceAlways,
		ProcFlags:        [2]uint32{0: dbcenums.PROC_FLAG_DEAL_MELEE_SWING},
	}

	return trigger, buff
}

func chargeAura() *core.StatBuffAura {
	return &core.StatBuffAura{Aura: &core.Aura{Label: "Charge Buff Proc"}}
}

// A buff that is its own trigger hears the hits that granted it, so a spender read off that row
// would take a charge back on the hit that handed it over.
func TestSpellDataProcChargeSpenderNeedsABuffOfItsOwn(t *testing.T) {
	trigger, buff := chargeRows()

	sameRow := chargeAura()
	attachChargeSpender(&core.Character{}, SpellDataProc{Name: "Test"}, trigger, trigger, sameRow)
	if sameRow.OnSpellHitDealt != nil {
		t.Error("a buff that is its own trigger got a charge spender, which would spend the charge it was just given")
	}

	ownRow := chargeAura()
	attachChargeSpender(&core.Character{}, SpellDataProc{Name: "Test"}, trigger, buff, ownRow)
	if ownRow.OnSpellHitDealt == nil {
		t.Error("a buff stating its own charges and flags got no spender")
	}
}

// The rows a trigger cannot be built from are the rows spelldata.ProcTrigger panics on, and asking
// first is the only thing between a charge spender and that panic.
func TestSpellDataProcStatesATrigger(t *testing.T) {
	_, buff := chargeRows()
	if !statesATrigger(buff) {
		t.Error("a row stating both a listener and a chance was refused")
	}

	// Damage of any kind names a listener without naming a hit kind, so this row hears something and
	// still has no mask a procs-per-minute rate could be measured against.
	maskless := *buff
	maskless.ProcFlags = [2]uint32{0: dbcenums.PROC_FLAG_TAKE_ANY_DAMAGE}
	maskless.ProcChanceSource = spelldata.ProcChancePPM
	maskless.RPPM = 2
	if statesATrigger(&maskless) {
		t.Error("a procs-per-minute row with no mask to measure it on was accepted")
	}

	pastTheEffects := *buff
	pastTheEffects.ProcChanceSource = spelldata.ProcChanceEffectN
	pastTheEffects.ProcChanceEffect = 3
	if statesATrigger(&pastTheEffects) {
		t.Error("a row naming an effect it does not carry as its chance was accepted")
	}
}

// The game files leave the duration off the buff on a fair few procs and state it on the trigger, and
// an aura of no duration is one core refuses to activate.
func TestProcBuffDurationFallsBackToTheTrigger(t *testing.T) {
	trigger, buff := chargeRows()
	cfg := SpellDataProc{Name: "Test", ItemID: 990102}

	buff.DurationMs = 0
	if got := procBuffDuration(cfg, trigger, buff); got != 12*time.Second {
		t.Errorf("duration = %v, want the trigger's 12s", got)
	}

	trigger.DurationMs = 0
	requirePanicContaining(t, "states a duration", func() {
		procBuffDuration(cfg, trigger, buff)
	})
}

func requirePanicContaining(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("no panic, want one mentioning %q", want)
		}
		if msg, ok := r.(string); !ok || !strings.Contains(msg, want) {
			t.Fatalf("panic %v, want one mentioning %q", r, want)
		}
	}()
	fn()
}

// A character bare enough to register an effect on: no environment, no raid, no equipment, which is
// all the registration reads off it.
type testAgent struct{ character *core.Character }

func (a *testAgent) GetCharacter() *core.Character                                    { return a.character }
func (a *testAgent) Initialize()                                                      {}
func (a *testAgent) AddRaidBuffs(_ *proto.RaidBuffs)                                  {}
func (a *testAgent) AddPartyBuffs(_ *proto.PartyBuffs)                                {}
func (a *testAgent) ApplyTalents()                                                    {}
func (a *testAgent) Reset(_ *core.Simulation)                                         {}
func (a *testAgent) OnManaTick(_ *core.Simulation)                                    {}
func (a *testAgent) OnEncounterStart(_ *core.Simulation)                              {}
func (a *testAgent) ExecuteCustomRotation(_ *core.Simulation)                         {}
func (a *testAgent) NewAPLValue(_ *core.APLRotation, _ *proto.APLValue) core.APLValue { return nil }
func (a *testAgent) NewAPLAction(_ *core.APLRotation, _ *proto.APLAction) core.APLActionImpl {
	return nil
}

func newTestAgent() *testAgent {
	character := core.NewCharacter(&core.Party{}, 0, &proto.Player{
		Name:      "Proc Tester",
		Class:     proto.Class_ClassWarrior,
		Race:      proto.Race_RaceHuman,
		Spec:      &proto.Player_ProtectionWarrior{ProtectionWarrior: &proto.ProtectionWarrior{}},
		Equipment: &proto.EquipmentSpec{Items: []*proto.ItemSpec{}},
	})

	return &testAgent{character: &character}
}

func auraByLabel(character *core.Character, label string) *core.Aura {
	for _, aura := range character.GetAuras() {
		if aura.Label == label {
			return aura
		}
	}

	return nil
}

// The whole registration on a character: the buff as the rows describe it, the listener that applies
// it, and the internal cooldown handed to the buff for the APL values that ask after it.
func TestSpellDataProcAppliesToACharacter(t *testing.T) {
	const itemID int32 = 990201

	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{{Id: itemID, Name: "Test Trinket"}}})
	RegisterItemEffects(map[int32][]*proto.ItemEffect{itemID: {{
		BuffId:           990111,
		EffectDurationMs: 15000,
		ScalingOptions: map[int32]*proto.ScalingItemEffectProperties{
			0: {Stats: map[int32]float64{int32(proto.Stat_StatSpellDamage): 40}},
		},
		Effect: &proto.ItemEffect_Proc{Proc: &proto.ProcEffect{}},
	}}})

	trigger := &spelldata.Spell{
		ID: 990110, Name: "Test Trinket Trigger", ProcChance: 15, ICDMs: 45000,
		ProcChanceSource: spelldata.ProcChanceColumn,
		ProcFlags:        [2]uint32{0: dbcenums.PROC_FLAG_DEAL_MELEE_ABILITY},
	}
	buff := &spelldata.Spell{ID: 990111, Name: "Test Trinket Buff", DurationMs: 15000}

	cfg := SpellDataProc{Name: "Test Trinket", ItemID: itemID,
		TriggerSpellID: trigger.ID, BuffSpellID: buff.ID}

	agent := newTestAgent()
	applySpellDataProc(agent, cfg, cfg.effectSource(), trigger, buff)

	procAura := auraByLabel(agent.character, "Test Trinket Proc")
	if procAura == nil {
		t.Fatal("the buff the proc applies was not registered")
	}
	if procAura.Duration != 15*time.Second {
		t.Errorf("buff duration = %v, want the buff row's 15s", procAura.Duration)
	}

	triggerAura := auraByLabel(agent.character, cfg.Name)
	if triggerAura == nil {
		t.Fatal("the listener was not registered")
	}
	// MakeProcTriggerAura keeps the action a proc is keyed by apart from the one its metrics use.
	if triggerAura.ActionIDForProc != (core.ActionID{ItemID: itemID}) {
		t.Errorf("listener action = %v, want the item's", triggerAura.ActionIDForProc)
	}
	if triggerAura.Icd == nil || triggerAura.Icd.Duration != 45*time.Second {
		t.Fatalf("listener ICD = %v, want the trigger row's 45s", triggerAura.Icd)
	}
	if procAura.Icd != triggerAura.Icd {
		t.Error("the buff was not handed the listener's cooldown, which is what the ICD-aware APL values read")
	}
}

// A damage proc has no buff at all: the row it names is a spell that deals damage, and what the
// game files state about that spell is its school, its hit table and the amount it rolls. What it does
// not state is that the game casts it off a hit, which is the GCD, the cost and the cast time this
// has to strip.
func TestSpellDataDamageProcAppliesToACharacter(t *testing.T) {
	const itemID int32 = 990301

	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{{Id: itemID, Name: "Test Spike"}}})

	trigger := &spelldata.Spell{
		ID: 990310, Name: "Test Spike Trigger", ProcChance: 5,
		ProcChanceSource: spelldata.ProcChanceColumn,
		ProcFlags:        [2]uint32{0: dbcenums.PROC_FLAG_TAKE_MELEE_SWING},
	}
	damage := &spelldata.Spell{
		ID: 990311, Name: "Test Spike Bolt", School: 8, DefenseType: 1,
		GCDMs: 1500, StartRecoveryCategory: 133,
		Powers:  []spelldata.Power{{Type: 0, Cost: 100}},
		Effects: []spelldata.Effect{{SpellID: 990311, Type: dbcenums.E_SCHOOL_DAMAGE, BasePoints: 50}},
	}

	cfg := SpellDataProc{Name: "Test Spike", ItemID: itemID,
		TriggerSpellID: trigger.ID, BuffSpellID: damage.ID}

	agent := newTestAgent()
	applySpellDataDamageProc(agent, cfg, cfg.effectSource(), trigger, damage)

	spell := agent.character.GetSpell(core.ActionID{SpellID: damage.ID})
	if spell == nil {
		t.Fatal("the spell the proc casts was not registered")
	}
	if spell.SpellSchool != core.SpellSchoolNature {
		t.Errorf("school = %v, want the row's Nature", spell.SpellSchool)
	}
	if spell.DefenseType != core.DefenseTypeMagic {
		t.Errorf("defense type = %v, want the row's magic", spell.DefenseType)
	}
	for _, flag := range []core.SpellFlag{core.SpellFlagPassiveSpell, core.SpellFlagNoOnCastComplete,
		core.SpellFlagProc} {
		if !spell.Flags.Matches(flag) {
			t.Errorf("flags = %b, want %b set", spell.Flags, flag)
		}
	}
	if spell.Flags.Matches(core.SpellFlagNoOnDamageDealt) {
		t.Error("the proc's hits skip the damage listeners, want them heard by the ones that hear procs")
	}
	if spell.ProcMask != core.ProcMaskEmpty {
		t.Errorf("proc mask = %v, want none: a proc's damage is heard through its flags", spell.ProcMask)
	}
	if spell.DefaultCast.GCD != 0 || spell.DefaultCast.CastTime != 0 || spell.DefaultCast.Cost != 0 {
		t.Errorf("cast = %+v, want no global cooldown, no cast time and no cost", spell.DefaultCast)
	}

	triggerAura := auraByLabel(agent.character, cfg.Name)
	if triggerAura == nil {
		t.Fatal("the listener was not registered")
	}
	if triggerAura.ActionIDForProc != (core.ActionID{ItemID: itemID}) {
		t.Errorf("listener action = %v, want the item's", triggerAura.ActionIDForProc)
	}
}

// Which unit a proc's damage lands on, per callback. The self-hit case is the one the sapper charges
// created: a character deals that hit to itself, and a proc reading the result answered it by
// hitting its own wearer.
func TestProcDamageTarget(t *testing.T) {
	character := &core.Character{Unit: core.Unit{Label: "Wearer"}}
	enemy := &core.Unit{Label: "Enemy"}
	attacker := &core.Unit{Label: "Attacker"}
	character.CurrentTarget = enemy

	for _, tc := range []struct {
		callback core.AuraCallback
		spell    *core.Spell
		result   *core.SpellResult
		want     *core.Unit
		why      string
	}{
		{core.CallbackOnSpellHitTaken, &core.Spell{Unit: attacker}, &core.SpellResult{Target: &character.Unit},
			attacker, "a hit taken answers the attacker, not the wearer the result names"},
		{core.CallbackOnSpellHitDealt, &core.Spell{Unit: &character.Unit}, &core.SpellResult{Target: attacker},
			attacker, "a hit dealt answers whatever was hit"},
		{core.CallbackOnSpellHitDealt, &core.Spell{Unit: &character.Unit}, &core.SpellResult{Target: &character.Unit},
			enemy, "a hit the character dealt to itself answers the current target"},
		{core.CallbackOnPeriodicDamageDealt, &core.Spell{Unit: &character.Unit}, &core.SpellResult{Target: &character.Unit},
			enemy, "the same for a tick"},
		{core.CallbackOnHealDealt, &core.Spell{Unit: &character.Unit}, &core.SpellResult{Target: attacker},
			enemy, "a heal names an ally, so nothing there says what to damage"},
		{core.CallbackOnCastComplete, &core.Spell{Unit: &character.Unit}, nil,
			enemy, "a cast carries no result at all"},
	} {
		if got := procDamageTarget(character, tc.callback, tc.spell, tc.result); got != tc.want {
			t.Errorf("target = %v, want %v: %s", got.Label, tc.want.Label, tc.why)
		}
	}
}

func testOneHander(id int32) *proto.SimItem {
	return &proto.SimItem{
		Id:             id,
		Name:           "Test Sword",
		Type:           proto.ItemType_ItemTypeWeapon,
		WeaponType:     proto.WeaponType_WeaponTypeSword,
		HandType:       proto.HandType_HandTypeOneHand,
		WeaponSpeed:    2.6,
		ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {WeaponDamageMin: 50, WeaponDamageMax: 90}},
	}
}

// Fiery Blaze 36's shape: a combat enchant on the main hand only, whose enchantment states its chance,
// the ProcChance the store writes onto its spell's row; the test's row states 15.
func TestStatedEnchantChanceRollsOnTheEnchantedWeaponOnly(t *testing.T) {
	const mainHandID, offHandID, enchantID int32 = 990401, 990402, 990403
	core.AddToDatabase(&proto.SimDatabase{
		Items:    []*proto.SimItem{testOneHander(mainHandID), testOneHander(offHandID)},
		Enchants: []*proto.SimEnchant{{EffectId: enchantID, Name: "Test Fiery Blaze"}},
	})

	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotOffHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: mainHandID, Enchant: enchantID}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: offHandID}

	agent := newTestAgent()
	character := agent.character
	character.Equipment = core.ProtoToEquipment(&proto.EquipmentSpec{Items: items})
	character.EnableAutoAttacks(agent, core.AutoAttackOptions{
		MainHand:       character.WeaponFromMainHand(),
		OffHand:        character.WeaponFromOffHand(),
		AutoSwingMelee: true,
	})

	trigger := &spelldata.Spell{ID: 990410, Name: "Test Fiery Blaze", ProcChance: 15}
	cfg := SpellDataProc{Name: "Test Fiery Blaze", EnchantID: enchantID, TriggerSpellID: trigger.ID,
		BuffSpellID: trigger.ID, IsWeaponProc: true}
	config := spellDataProcListener(character, cfg, cfg.effectSource(), trigger)

	if config.ProcChance != 0 {
		t.Errorf("flat chance = %v, want none: it would roll on the off hand's hits too", config.ProcChance)
	}
	if config.DPM == nil {
		t.Fatal("no manager bound to the enchanted weapon")
	}
	if got := config.DPM.Chance(core.ProcMaskMeleeMHAuto, nil); got != 0.15 {
		t.Errorf("main hand chance = %v, want 0.15", got)
	}
	if got := config.DPM.Chance(core.ProcMaskMeleeOHAuto, nil); got != 0 {
		t.Errorf("off hand chance = %v, want 0", got)
	}
}

// A PPM override on a rate-less enchant proc's row clears its refusal, and the registration measures
// the rate on the enchanted weapon's hits. An equip aura that hears only spells never rolls, which is
// the refusal its rate gets. Each row's copy is stripped of any rate of its own first.
func TestSpellDataProcTakesAPPMOverride(t *testing.T) {
	const mainHandID, offHandID, enchantID int32 = 990501, 990502, 990503
	const ppm = 2.0
	core.AddToDatabase(&proto.SimDatabase{
		Items: []*proto.SimItem{testOneHander(mainHandID), testOneHander(offHandID)},
		Enchants: []*proto.SimEnchant{{EffectId: enchantID, Name: "Test Rate-less Enchant",
			Type: proto.ItemType_ItemTypeWeapon}},
	})

	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotOffHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: mainHandID, Enchant: enchantID}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: offHandID}

	agent := newTestAgent()
	character := agent.character
	character.Equipment = core.ProtoToEquipment(&proto.EquipmentSpec{Items: items})
	character.EnableAutoAttacks(agent, core.AutoAttackOptions{
		MainHand:       character.WeaponFromMainHand(),
		OffHand:        character.WeaponFromOffHand(),
		AutoSwingMelee: true,
	})
	mainHandSpeed := testOneHander(mainHandID).WeaponSpeed
	mainHandChance := mainHandSpeed * (ppm / 60)

	for _, tc := range []struct {
		name         string
		spellID      int32
		isWeaponProc bool
		heard        core.ProcMask
		want         float64
		refusedFor   string
	}{
		{"Fiery Weapon's combat spell 13897", 13897, true, core.ProcMaskMeleeMHAuto, mainHandChance, ""},
		{"Unholy Weapon's combat spell 20006", 20006, true, core.ProcMaskMeleeMHAuto, mainHandChance, ""},
		{"Revelation's equip aura 1248806", 1248806, false, core.ProcMaskSpellDamage, 0,
			spelldata.ReasonPPMHearsNoWeaponHits},
	} {
		t.Run(tc.name, func(t *testing.T) {
			row := *spelldata.MustFind(tc.spellID)
			row.ProcChance, row.RPPM = 0, 0
			row.ProcChanceSource, row.ProcChanceEffect = spelldata.ProcChancePPM, 0
			unsupported := func() []string {
				if tc.isWeaponProc {
					return spelldata.ItemProcUnsupportedWithRate(&row, true, 0)
				}
				return spelldata.EnchantAuraUnsupported(&row)
			}

			if !slices.Contains(unsupported(), spelldata.ReasonStatesNoRate) {
				t.Fatalf("unsupported = %v before the override, want the missing rate named", unsupported())
			}

			row.RPPM = ppm
			row.ProcChanceSource, row.ProcChanceEffect = spelldata.ProcChancePPM, 0
			var want []string
			if tc.refusedFor != "" {
				want = []string{tc.refusedFor}
			}
			if got := unsupported(); !slices.Equal(got, want) {
				t.Errorf("unsupported = %v with the override, want %v", got, want)
			}

			cfg := SpellDataProc{Name: "Test " + tc.name, EnchantID: enchantID, TriggerSpellID: row.ID,
				BuffSpellID: row.ID, IsWeaponProc: tc.isWeaponProc}
			config := spellDataProcListener(character, cfg, cfg.effectSource(), &row)

			if config.ProcChance != 0 {
				t.Errorf("flat chance = %v, want none beside the rate", config.ProcChance)
			}
			if config.DPM == nil {
				t.Fatal("no procs-per-minute manager")
			}
			if got := config.DPM.Chance(tc.heard, nil); math.Abs(got-tc.want) > 1e-12 {
				t.Errorf("chance = %v, want %v (%v ppm, main hand %v s)", got, tc.want, ppm, mainHandSpeed)
			}
			if got := config.DPM.Chance(core.ProcMaskMeleeOHAuto, nil); got != 0 {
				t.Errorf("off hand chance = %v, want 0", got)
			}
		})
	}
}

// An enchant aura's procs-per-minute rate follows its weapon: a weapon enchant's melee hits roll only
// on the hand carrying it, while an enchant on no weapon keeps the main hand's pricing, and no enchant's
// spell or heal hits roll. Recovery's 1248761 hears melee hits and Revelation's 1248806 spells.
func TestEnchantAuraPPMFollowsItsWeapon(t *testing.T) {
	const slowID, fastID int32 = 990801, 990802
	const weaponEnchantID, cloakEnchantID, shieldEnchantID int32 = 990803, 990804, 990805
	const slow, fast, ppm = 2.6, 1.8, 2.0
	fastOneHander := testOneHander(fastID)
	fastOneHander.WeaponSpeed = fast
	core.AddToDatabase(&proto.SimDatabase{
		Items: []*proto.SimItem{testOneHander(slowID), fastOneHander},
		Enchants: []*proto.SimEnchant{
			{EffectId: weaponEnchantID, Name: "Test Weapon Enchant", Type: proto.ItemType_ItemTypeWeapon},
			{EffectId: cloakEnchantID, Name: "Test Cloak Enchant", Type: proto.ItemType_ItemTypeBack},
			{EffectId: shieldEnchantID, Name: "Test Shield Enchant", Type: proto.ItemType_ItemTypeWeapon,
				EnchantType: proto.EnchantType_EnchantTypeShield},
		},
	})
	chance := func(speed float64) float64 { return speed * (ppm / 60) }

	for _, tc := range []struct {
		name                            string
		enchantID                       int32
		mainHandEnchant, offHandEnchant int32
		spellID                         int32
		mainHand, offHand, spell        float64
	}{
		{"weapon enchant on the main hand, melee aura", weaponEnchantID, weaponEnchantID, 0, 1248761,
			chance(slow), 0, 0},
		{"weapon enchant on the off hand, melee aura", weaponEnchantID, 0, weaponEnchantID, 1248761,
			0, chance(fast), 0},
		{"weapon enchant on both hands, melee aura", weaponEnchantID, weaponEnchantID, weaponEnchantID, 1248761,
			chance(slow), chance(fast), 0},
		{"weapon enchant on the off hand, spell aura", weaponEnchantID, 0, weaponEnchantID, 1248806,
			0, 0, 0},
		{"cloak enchant, melee aura", cloakEnchantID, 0, 0, 1248761,
			chance(slow), chance(fast), 0},
		{"cloak enchant, spell aura", cloakEnchantID, 0, 0, 1248806,
			0, 0, 0},
		{"shield enchant, spell aura", shieldEnchantID, 0, 0, 1248806,
			0, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotOffHand+1)
			for i := range items {
				items[i] = &proto.ItemSpec{}
			}
			items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: slowID, Enchant: tc.mainHandEnchant}
			items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: fastID, Enchant: tc.offHandEnchant}

			agent := newTestAgent()
			character := agent.character
			character.Equipment = core.ProtoToEquipment(&proto.EquipmentSpec{Items: items})
			character.EnableAutoAttacks(agent, core.AutoAttackOptions{
				MainHand:       character.WeaponFromMainHand(),
				OffHand:        character.WeaponFromOffHand(),
				AutoSwingMelee: true,
			})

			row := *spelldata.MustFind(tc.spellID)
			row.RPPM = ppm
			row.ProcChanceSource, row.ProcChanceEffect = spelldata.ProcChancePPM, 0
			cfg := SpellDataProc{Name: "Test " + tc.name, EnchantID: tc.enchantID, TriggerSpellID: row.ID,
				BuffSpellID: row.ID}
			config := spellDataProcListener(character, cfg, cfg.effectSource(), &row)

			if config.DPM == nil {
				t.Fatal("no procs-per-minute manager")
			}
			for _, hit := range []struct {
				name string
				mask core.ProcMask
				want float64
			}{
				{"main hand", core.ProcMaskMeleeMHAuto, tc.mainHand},
				{"off hand", core.ProcMaskMeleeOHAuto, tc.offHand},
				{"spell", core.ProcMaskSpellDamage, tc.spell},
				{"heal", core.ProcMaskSpellHealing, tc.spell},
			} {
				if got := config.DPM.Chance(hit.mask, nil); got != hit.want {
					t.Errorf("%s chance = %v, want %v", hit.name, got, hit.want)
				}
			}
		})
	}
}

// A stacking trigger's procs-per-minute rate is measured the way its opener's is: a weapon enchant's
// on the hand carrying it and never on spells, an item's off the main hand for everything but the
// off hand's and the ranged hits.
func TestStackTriggerPPMFollowsTheEffect(t *testing.T) {
	const slowID, fastID, itemID, weaponEnchantID int32 = 990901, 990902, 990903, 990904
	const slow, fast, ppm = 2.6, 1.8, 2.0
	fastOneHander := testOneHander(fastID)
	fastOneHander.WeaponSpeed = fast
	core.AddToDatabase(&proto.SimDatabase{
		Items: []*proto.SimItem{testOneHander(slowID), fastOneHander},
		Enchants: []*proto.SimEnchant{
			{EffectId: weaponEnchantID, Name: "Test Weapon Enchant", Type: proto.ItemType_ItemTypeWeapon},
		},
	})
	chance := func(speed float64) float64 { return speed * (ppm / 60) }

	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotOffHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: slowID}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: fastID, Enchant: weaponEnchantID}

	agent := newTestAgent()
	character := agent.character
	character.Equipment = core.ProtoToEquipment(&proto.EquipmentSpec{Items: items})
	character.EnableAutoAttacks(agent, core.AutoAttackOptions{
		MainHand:       character.WeaponFromMainHand(),
		OffHand:        character.WeaponFromOffHand(),
		AutoSwingMelee: true,
	})

	stackProc := &proto.ProcEffect{ProcRate: &proto.ProcEffect_Ppm{Ppm: ppm}}
	for _, tc := range []struct {
		name                     string
		source                   effectSource
		mainHand, offHand, spell float64
	}{
		{"weapon enchant on the off hand", effectSource{id: weaponEnchantID, isEnchant: true}, 0, chance(fast), 0},
		{"item", effectSource{id: itemID}, chance(slow), chance(fast), chance(slow)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dpm := dpmForMask(character, tc.source, stackProc.GetPpm(), core.ProcMaskMelee|core.ProcMaskSpellDamage)
			if dpm == nil {
				t.Fatal("no procs-per-minute manager")
			}
			for _, hit := range []struct {
				name string
				mask core.ProcMask
				want float64
			}{
				{"main hand", core.ProcMaskMeleeMHAuto, tc.mainHand},
				{"off hand", core.ProcMaskMeleeOHAuto, tc.offHand},
				{"spell", core.ProcMaskSpellDamage, tc.spell},
			} {
				if got := dpm.Chance(hit.mask, nil); got != hit.want {
					t.Errorf("%s chance = %v, want %v", hit.name, got, hit.want)
				}
			}
		})
	}
}

package shared

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

const alchemistsStoneID int32 = 13503

func init() {
	registerConsumableTester(proto.Player_Warlock{}, proto.Spec_SpecWarlock, func(c *testCaster) {
		c.EnableManaBar()
	}, func(player *proto.Player, spec interface{}) { player.Spec = spec.(*proto.Player_Warlock) })
	registerConsumableTester(proto.Player_ProtectionWarrior{}, proto.Spec_SpecProtectionWarrior, func(c *testCaster) {
		c.EnableRageBar(core.RageBarOptions{MaxRage: 100, BaseRageMultiplier: 1})
	}, func(player *proto.Player, spec interface{}) { player.Spec = spec.(*proto.Player_ProtectionWarrior) })
	registerConsumableTester(proto.Player_FeralCatDruid{}, proto.Spec_SpecFeralCatDruid, func(c *testCaster) {
		c.EnableEnergyBar(core.EnergyBarOptions{MaxEnergy: 100, MaxComboPoints: 5, UnitClass: proto.Class_ClassDruid})
	}, func(player *proto.Player, spec interface{}) { player.Spec = spec.(*proto.Player_FeralCatDruid) })
}

func registerConsumableTester(options interface{}, spec proto.Spec, enable func(*testCaster), setSpec func(*proto.Player, interface{})) {
	core.RegisterAgentFactory(options, spec, func(character *core.Character, _ *proto.Player, _ *proto.Raid) core.Agent {
		c := &testCaster{Character: *character}
		enable(c)
		return c
	}, setSpec)
}

func manaUser() *proto.Player {
	return &proto.Player{Class: proto.Class_ClassWarlock, Race: proto.Race_RaceHuman, Spec: &proto.Player_Warlock{}}
}

func rageUser() *proto.Player {
	return &proto.Player{Class: proto.Class_ClassWarrior, Race: proto.Race_RaceHuman, Spec: &proto.Player_ProtectionWarrior{}}
}

func energyUser() *proto.Player {
	return &proto.Player{Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf, Spec: &proto.Player_FeralCatDruid{}}
}

type testConsumable struct {
	id             int32
	consumableType proto.ConsumableType
	effect         *proto.ItemEffect
}

const (
	potion   = proto.ConsumableType_ConsumableTypePotion
	conjured = proto.ConsumableType_ConsumableTypeConjured
	other    = proto.ConsumableType_ConsumableTypeOther
)

// The player with the consumables in its spec, each of its type and registered through the potion
// builder, at half health, with its spells certain to hit and never to crit.
func newConsumableSim(t *testing.T, player *proto.Player, consumes *proto.ConsumesSpec, trinket int32, consumables ...testConsumable) (*core.Simulation, *testCaster) {
	t.Helper()
	for _, c := range consumables {
		core.AddToDatabase(&proto.SimDatabase{Consumables: []*proto.Consumable{{Id: c.id, Name: "Test Consumable",
			Type: c.consumableType, ItemEffects: []*proto.ItemEffect{c.effect}}}})
		NewSpellDataOnUseConsumable(c.id)
	}

	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotTrinket2+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	if trinket != 0 {
		core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{{Id: trinket, Name: "Test Trinket",
			Type: proto.ItemType_ItemTypeTrinket, ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}}}})
		items[proto.ItemSlot_ItemSlotTrinket1] = &proto.ItemSpec{Id: trinket}
	}

	player.Name = "Drinker"
	player.Buffs = &proto.IndividualBuffs{}
	player.Equipment = &proto.EquipmentSpec{Items: items}
	player.Consumables = consumes
	player.Profession1 = proto.Profession_Alchemy
	player.Rotation = &proto.APLRotation{}

	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{player},
			Buffs:   &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{
			Targets:  []*proto.Target{{Name: "target", Level: 60}},
			Duration: 600,
		},
	}, simsignals.CreateSignals())
	sim.Reset()

	caster := sim.Raid.Parties[0].Players[0].(*testCaster)
	caster.AddStatsDynamic(sim, stats.Stats{
		stats.SpellHitPercent:  100,
		stats.SpellCritPercent: -caster.GetStat(stats.SpellCritPercent),
	})
	caster.RemoveHealth(sim, caster.MaxHealth()/2)
	return sim, caster
}

func consumableSpell(t *testing.T, caster *testCaster, itemID int32) *core.Spell {
	t.Helper()
	spell := caster.GetSpell(core.ActionID{ItemID: itemID})
	if spell == nil {
		t.Fatalf("consumable %d registered no spell", itemID)
	}
	return spell
}

func gainedBy(caster *testCaster, itemID int32) float64 {
	for _, r := range caster.Metrics.ToProto().Resources {
		if r.Id.GetItemId() == itemID {
			return r.ActualGain
		}
	}
	return 0
}

// Major Mana Potion 13444 casts 17531, E_ENERGIZE 1800 mana with a 0.5 spread, on potion category 4
// for 120 s. The spec's combat potion is a major cooldown the manager holds until the whole gain fits;
// another potion the spec lists is only a spell, on the same category.
func TestPotionRestoresManaOnThePotionCategory(t *testing.T) {
	const combatID, otherID int32 = 992001, 992002
	sim, drinker := newConsumableSim(t, manaUser(), &proto.ConsumesSpec{PotId: combatID, Potions: []int32{otherID}}, 0,
		testConsumable{combatID, potion, testOnUse(17531, 0, 4, 120000)},
		testConsumable{otherID, potion, testOnUse(17531, 0, 4, 120000)})
	potion, other := consumableSpell(t, drinker, combatID), consumableSpell(t, drinker, otherID)

	mcd := drinker.GetInitialMajorCooldown(potion.ActionID)
	if mcd.Spell != potion || !mcd.Type.Matches(core.CooldownTypeMana) {
		t.Fatalf("the combat potion is not a mana major cooldown: %v", mcd.Type)
	}
	if !potion.Flags.Matches(core.SpellFlagCombatPotion|core.SpellFlagPotion|core.SpellFlagEncounterOnly) || other.Flags.Matches(core.SpellFlagCombatPotion) {
		t.Errorf("flags: combat potion %v, other %v", potion.Flags, other.Flags)
	}
	if drinker.GetInitialMajorCooldown(other.ActionID).Spell != nil {
		t.Errorf("the other potion is a major cooldown")
	}
	if mcd.ShouldActivate(sim, &drinker.Character) {
		t.Errorf("the potion would be drunk on a full bar")
	}

	drinker.SpendMana(sim, drinker.CurrentMana()-100, drinker.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionManaRegen, Tag: 99}))
	if !mcd.ShouldActivate(sim, &drinker.Character) {
		t.Errorf("the potion would not be drunk at 100 mana")
	}
	potion.Cast(sim, &drinker.Unit)
	start := sim.CurrentTime
	if got := drinker.CurrentMana() - 100; got < 1350 || got > 2250 {
		t.Errorf("the potion restored %v, want 1350 to 2250", got)
	}
	if offCooldown(sim, other) {
		t.Errorf("the other potion is usable while category 4 runs")
	}
	stepPast(t, sim, start+120*time.Second)
	if !offCooldown(sim, other) || !offCooldown(sim, potion) {
		t.Errorf("the potions are not usable once category 4's 120 s ran out")
	}
}

// Mighty Rage Potion 13442 casts 17528: E_ENERGIZE 600 into rage, stated in tenths, with a 0.5
// spread, and A_MOD_STAT 60 Strength for 20 s.
func TestPotionGivesRageAndItsStrength(t *testing.T) {
	const potionID int32 = 992003
	sim, drinker := newConsumableSim(t, rageUser(), &proto.ConsumesSpec{PotId: potionID}, 0,
		testConsumable{potionID, potion, testOnUse(17528, 0, 4, 120000)})
	potion := consumableSpell(t, drinker, potionID)
	strength := drinker.GetStat(stats.Strength)

	potion.Cast(sim, &drinker.Unit)
	if got := drinker.CurrentRage(); got < 45 || got > 75 {
		t.Errorf("rage after the potion is %v, want 45 to 75", got)
	}
	if got := drinker.GetStat(stats.Strength) - strength; got != 60 {
		t.Errorf("the potion gave %v Strength, want 60", got)
	}
	if potion.RelatedSelfBuff == nil || potion.RelatedSelfBuff.Duration != 20*time.Second {
		t.Fatalf("the potion's buff is not the row's 20 s aura")
	}
	stepPast(t, sim, sim.CurrentTime+20*time.Second+time.Millisecond)
	if got := drinker.GetStat(stats.Strength); got != strength {
		t.Errorf("Strength after the buff ran out is %v, want %v", got, strength)
	}
}

// Demonic Rune 12662 casts 16666: E_ENERGIZE 1200 mana and E_SCHOOL_DAMAGE 800 Shadow to the user,
// both with a 0.5 spread, on conjured category 1153 for 120 s. The conjured item is a major cooldown
// only while it is the spec's chosen one.
func TestConjuredRuneRestoresManaAtTheCostOfLife(t *testing.T) {
	const runeID, otherID int32 = 992004, 992005
	sim, drinker := newConsumableSim(t, manaUser(), &proto.ConsumesSpec{ConjuredId: otherID, ConjuredItems: []int32{runeID, otherID}}, 0,
		testConsumable{runeID, conjured, testOnUse(16666, -1, 1153, 120000)},
		testConsumable{otherID, conjured, testOnUse(10058, 0, 1153, 120000)})
	rune, other := consumableSpell(t, drinker, runeID), consumableSpell(t, drinker, otherID)

	drinker.SpendMana(sim, drinker.CurrentMana()-100, drinker.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionManaRegen, Tag: 99}))
	if mcd := drinker.GetInitialMajorCooldown(rune.ActionID); mcd.ShouldActivate(sim, &drinker.Character) {
		t.Errorf("the rune would be used while another conjured item is the spec's")
	}
	if !rune.Flags.Matches(core.SpellFlagConjured) {
		t.Errorf("the rune is not flagged conjured")
	}

	health := drinker.CurrentHealth()
	rune.Cast(sim, &drinker.Unit)
	if got := drinker.CurrentMana() - 100; got < 900 || got > 1500 {
		t.Errorf("the rune restored %v, want 900 to 1500", got)
	}
	if lost := health - drinker.CurrentHealth(); lost < 400 || lost > 1200 {
		t.Errorf("the rune cost %v health, want 400 to 1200", lost)
	}
	if offCooldown(sim, other) {
		t.Errorf("the mana gem is usable while category 1153 runs")
	}
}

// 16666's damage to the user lands however poor the user's spell hit.
func TestConjuredRuneAlwaysHitsTheUser(t *testing.T) {
	const runeID int32 = 992031
	sim, drinker := newConsumableSim(t, manaUser(), &proto.ConsumesSpec{ConjuredId: runeID, ConjuredItems: []int32{runeID}}, 0,
		testConsumable{runeID, conjured, testOnUse(16666, -1, 1153, 120000)})
	drinker.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: -1000})
	rune := consumableSpell(t, drinker, runeID)
	self := drinker.GetSpell(core.ActionID{ItemID: runeID}.WithTag(1))
	damage := spelldata.MustFind(16666).CasterDamageEffect()

	for range 20 {
		rune.SharedCD.Reset()
		health := drinker.CurrentHealth()
		rune.Cast(sim, &drinker.Unit)
		if lost := health - drinker.CurrentHealth(); lost < damage.Min() {
			t.Errorf("the rune cost %v health, want at least %v", lost, damage.Min())
		}
		drinker.GainHealth(sim, drinker.MaxHealth(), self.HealthMetrics(&drinker.Unit))
	}
}

// Thistle Tea 7676 casts 9512, E_ENERGIZE 100 into energy with no spread, on a 300 s cooldown of its
// own and conjured category 1153 for 120 s.
func TestConjuredTeaRestoresAFlatHundredEnergy(t *testing.T) {
	const teaID int32 = 992006
	sim, drinker := newConsumableSim(t, energyUser(), &proto.ConsumesSpec{ConjuredId: teaID, ConjuredItems: []int32{teaID}}, 0,
		testConsumable{teaID, conjured, testOnUse(9512, 300000, 1153, 120000)})
	tea := consumableSpell(t, drinker, teaID)
	mcd := drinker.GetInitialMajorCooldown(tea.ActionID)

	drinker.SpendEnergy(sim, drinker.CurrentEnergy(), drinker.EnergyRefundMetrics)
	if !mcd.ShouldActivate(sim, &drinker.Character) || !mcd.Type.Matches(core.CooldownTypeDPS) {
		t.Errorf("the tea is not a DPS cooldown used on an empty bar: %v", mcd.Type)
	}
	tea.Cast(sim, &drinker.Unit)
	if got := drinker.CurrentEnergy(); got != 100 {
		t.Errorf("energy after the tea is %v, want 100", got)
	}
	if got := tea.CD.ReadyAt(); got != sim.CurrentTime+300*time.Second {
		t.Errorf("the tea is ready again at %v, want 300 s on", got)
	}
}

// Razorlash Root 17747 casts 21955: A_PERIODIC_HEAL 60 and A_PERIODIC_ENERGIZE 60 mana every 1 s for
// 10 s.
func TestConjuredRootHealsAndRestoresManaOverTime(t *testing.T) {
	const rootID int32 = 992007
	sim, drinker := newConsumableSim(t, manaUser(), &proto.ConsumesSpec{ConjuredItems: []int32{rootID}}, 0,
		testConsumable{rootID, conjured, testOnUse(21955, 0, 1153, 120000)})
	root := consumableSpell(t, drinker, rootID)

	drinker.SpendMana(sim, 2000, drinker.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionManaRegen, Tag: 99}))
	root.Cast(sim, &drinker.Unit)
	stepPast(t, sim, sim.CurrentTime+10*time.Second+time.Millisecond)

	if got := gainedBy(drinker, rootID); got != 600 {
		t.Errorf("the root restored %v mana, want 10 ticks of 60", got)
	}
	if got := root.SpellMetrics[drinker.UnitIndex].TotalHealing; got != 600 {
		t.Errorf("the root healed %v, want 10 ticks of 60", got)
	}
}

// Restored Mana Potion 282013 casts 1312702, E_ENERGIZE_PCT 20: a fifth of the maximum.
func TestPotionRestoresAShareOfTheBar(t *testing.T) {
	const potionID int32 = 992008
	sim, drinker := newConsumableSim(t, manaUser(), &proto.ConsumesSpec{Potions: []int32{potionID}}, 0,
		testConsumable{potionID, potion, testOnUse(1312702, 0, 4, 120000)})

	drinker.SpendMana(sim, drinker.CurrentMana(), drinker.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionManaRegen, Tag: 99}))
	consumableSpell(t, drinker, potionID).Cast(sim, &drinker.Unit)
	if got, want := drinker.CurrentMana(), drinker.MaxMana()*0.2; math.Abs(got-want) > 1e-6 {
		t.Errorf("mana after the potion is %v, want %v", got, want)
	}
}

// Minor Spellblasting Potion 250932's item row states cooldown -1, category 0 and category cooldown
// -1; its spell 1251919 states potion category 4 for 120 s, which it runs on.
func TestPotionDefersToItsSpellsCategory(t *testing.T) {
	const blastingID, manaID int32 = 992009, 992010
	sim, drinker := newConsumableSim(t, manaUser(), &proto.ConsumesSpec{Potions: []int32{blastingID, manaID}}, 0,
		testConsumable{blastingID, potion, testOnUse(1251919, -1, 0, -1)},
		testConsumable{manaID, potion, testOnUse(17531, 0, 4, 120000)})
	blasting, mana := consumableSpell(t, drinker, blastingID), consumableSpell(t, drinker, manaID)

	blasting.Cast(sim, &drinker.Unit)
	if !blasting.RelatedSelfBuff.IsActive() {
		t.Errorf("the potion's buff is not up")
	}
	if offCooldown(sim, mana) || offCooldown(sim, blasting) {
		t.Errorf("a potion is usable while the spell's category 4 runs")
	}
	stepPast(t, sim, sim.CurrentTime+120*time.Second)
	if !offCooldown(sim, mana) {
		t.Errorf("the mana potion is not usable once category 4's 120 s ran out")
	}
}

// Alchemists' Stone 13503's 17619 raises the healing and mana potions give by its effect's 20%:
// Major Healing Potion 13446's 17534 and Major Mana Potion's 17531, not a healthstone's heal.
func TestAlchemistsStoneRaisesHealingAndManaPotions(t *testing.T) {
	stone := spelldata.MustFind(alchemistsStoneSpellID).EffectN(1).Percent()
	if stone != 0.2 {
		t.Fatalf("17619 states %v, want 20%%", stone)
	}

	gains := func(trinket int32, heal, mana, healthstone int32) (float64, float64, float64) {
		sim, drinker := newConsumableSim(t, manaUser(), &proto.ConsumesSpec{Potions: []int32{heal, mana}, ConjuredItems: []int32{healthstone}}, trinket,
			testConsumable{heal, potion, testOnUse(17534, 0, 4, 120000)},
			testConsumable{mana, potion, testOnUse(17531, 0, 4, 120000)},
			testConsumable{healthstone, conjured, testOnUse(11732, 0, 1153, 120000)})
		drinker.AddStatsDynamic(sim, stats.Stats{stats.Mana: 10000, stats.Health: 10000})
		drinker.SpendMana(sim, drinker.CurrentMana(), drinker.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionManaRegen, Tag: 99}))
		drinker.RemoveHealth(sim, drinker.CurrentHealth()-1)

		consumableSpell(t, drinker, heal).Cast(sim, &drinker.Unit)
		healed := consumableSpell(t, drinker, heal).SpellMetrics[drinker.UnitIndex].TotalHealing
		stepPast(t, sim, sim.CurrentTime+120*time.Second)
		consumableSpell(t, drinker, mana).Cast(sim, &drinker.Unit)
		drinker.RemoveHealth(sim, drinker.CurrentHealth()-1)
		consumableSpell(t, drinker, healthstone).Cast(sim, &drinker.Unit)
		return healed, gainedBy(drinker, mana), consumableSpell(t, drinker, healthstone).SpellMetrics[drinker.UnitIndex].TotalHealing
	}

	healed, restored, stoned := gains(0, 992011, 992012, 992013)
	healedWith, restoredWith, stonedWith := gains(alchemistsStoneID, 992014, 992015, 992016)
	if math.Abs(healedWith/healed-1.2) > 1e-9 || math.Abs(restoredWith/restored-1.2) > 1e-9 {
		t.Errorf("with the stone the potions gave %v and %v, want 1.2 times %v and %v", healedWith, restoredWith, healed, restored)
	}
	if stonedWith != stoned || stoned != 1440 {
		t.Errorf("the healthstone healed %v with the stone and %v without, want 1440 both", stonedWith, stoned)
	}
}

// Greater Fire Protection Potion 13457 casts 17543: A_SCHOOL_ABSORB 2600 with Misc 4 (Fire) and a
// 0.5 spread on the drinker for 120 s, on potion category 4 for 120 s. The combat potion is a survival
// major cooldown whose shield takes fire damage only; Alchemists' Stone does not raise it.
func TestProtectionPotionShieldsAgainstItsSchool(t *testing.T) {
	const potionID int32 = 992021
	const fireProtection int32 = 17543
	withoutSpread(t, fireProtection)
	sim, drinker := newConsumableSim(t, manaUser(), &proto.ConsumesSpec{PotId: potionID}, alchemistsStoneID,
		testConsumable{potionID, potion, testOnUse(fireProtection, 0, 4, 120000)})
	potion := consumableSpell(t, drinker, potionID)
	shield := absorbAura(t, drinker, fireProtection, potionID)

	if mcd := drinker.GetInitialMajorCooldown(potion.ActionID); mcd.Spell != potion || !mcd.Type.Matches(core.CooldownTypeSurvival) {
		t.Fatalf("the protection potion is not a survival major cooldown: %v", mcd.Type)
	}

	potion.Cast(sim, &drinker.Unit)
	if !shield.IsActive() || shield.Duration != 120*time.Second {
		t.Fatalf("the potion put up no 120 s shield: active %v, duration %v", shield.IsActive(), shield.Duration)
	}
	if got := takenAfterAbsorbs(sim, drinker, core.SpellSchoolFrost, 1000); got != 1000 {
		t.Errorf("a frost hit of 1000 landed for %v, want all of it", got)
	}
	if got := takenAfterAbsorbs(sim, drinker, core.SpellSchoolFire, 3000); got != 400 {
		t.Errorf("a fire hit of 3000 landed for %v, want 400 past the 2600 shield", got)
	}
	if offCooldown(sim, potion) {
		t.Errorf("the potion is usable again while category 4 runs")
	}
}

// Flame Deflector 4376 casts 4057, A_SCHOOL_ABSORB 500 Fire for 60 s, on the item's own 900 s
// cooldown with no category. Outside the potion and conjured categories it is a major cooldown
// whenever the spec names it.
func TestOnUseConsumableShieldsOnTheItemsCooldown(t *testing.T) {
	const deflectorID int32 = 992022
	sim, user := newConsumableSim(t, manaUser(), &proto.ConsumesSpec{ExtraIds: []int32{deflectorID}}, 0,
		testConsumable{deflectorID, other, testOnUse(4057, 900000, 0, 0)})
	deflector := consumableSpell(t, user, deflectorID)

	mcd := user.GetInitialMajorCooldown(deflector.ActionID)
	if mcd.Spell != deflector || !mcd.Type.Matches(core.CooldownTypeSurvival) {
		t.Fatalf("the deflector is not a survival major cooldown: %v", mcd.Type)
	}
	if mcd.ShouldActivate != nil && !mcd.ShouldActivate(sim, &user.Character) {
		t.Errorf("the manager holds the deflector back")
	}
	if deflector.Flags.Matches(core.SpellFlagPotion) || deflector.Flags.Matches(core.SpellFlagConjured) {
		t.Errorf("the deflector is flagged a potion or a conjured item: %v", deflector.Flags)
	}

	deflector.Cast(sim, &user.Unit)
	if got := takenAfterAbsorbs(sim, user, core.SpellSchoolFire, 800); got != 300 {
		t.Errorf("a fire hit of 800 landed for %v, want 300 past the 500 shield", got)
	}
	if deflector.CD.Duration != 900*time.Second || deflector.SharedCD.Timer != nil {
		t.Errorf("the deflector runs on a %v cooldown and shared timer %v, want its own 900 s alone",
			deflector.CD.Duration, deflector.SharedCD.Timer)
	}
}

// The spells the builder cannot run, with what stops it.
func TestConsumableUnsupportedNamesWhatStopsIt(t *testing.T) {
	for _, tc := range []struct {
		spellID int32
		want    string
	}{
		{1244863, "E_DUMMY"},         // Minor Discolored Healing Potion: the self-damage share lives in a dummy
		{1252554, "not the user"},    // Dragonfire Potion: a cone in front of the drinker
		{15822, "A_MOD_STUN"},        // Dreamless Sleep Potion
		{806, "A_PERIODIC_DAMAGE"},   // Potion of Fervor's damage to the drinker every 3 s
		{1297671, "E_TRIGGER_SPELL"}, // Power Recycler
		{833, "sits"},                // Green Tea Leaf: its mana regen drops when the user stands
	} {
		got := strings.Join(ConsumableUnsupported(spelldata.MustFind(tc.spellID)), "; ")
		if !strings.Contains(got, tc.want) {
			t.Errorf("%d: unsupported %q, want it to name %s", tc.spellID, got, tc.want)
		}
	}
	for _, id := range []int32{17531, 17528, 16666, 9512, 22729, 21955, 1312698, 1312702, 1251919, 11732, 7242} {
		if got := ConsumableUnsupported(spelldata.MustFind(id)); got != nil {
			t.Errorf("%d: unsupported %v, want none", id, got)
		}
	}
}

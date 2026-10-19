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
	return newConsumableSimAgainst(t, player, consumes, trinket, 1, consumables...)
}

// The same against the given number of targets.
func newConsumableSimAgainst(t *testing.T, player *proto.Player, consumes *proto.ConsumesSpec, trinket int32, targetCount int, consumables ...testConsumable) (*core.Simulation, *testCaster) {
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

	targets := make([]*proto.Target, targetCount)
	for i := range targets {
		targets[i] = &proto.Target{Name: "target", Level: 60}
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{player},
			Buffs:   &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{
			Targets:  targets,
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

// Major Mana Potion 13444 casts 17531, E_ENERGIZE mana with a spread; the test's on-use runs on
// potion category 4 for 120 s. The spec's resource potion is a major cooldown the manager holds until
// the whole gain fits; another potion the spec lists is only a spell, on the same category.
func TestPotionRestoresManaOnThePotionCategory(t *testing.T) {
	const resourceID, otherID int32 = 992001, 992002
	sim, drinker := newConsumableSim(t, manaUser(), &proto.ConsumesSpec{ResourcePotId: resourceID, Potions: []int32{otherID}}, 0,
		testConsumable{resourceID, potion, testOnUse(17531, 0, 4, 120000)},
		testConsumable{otherID, potion, testOnUse(17531, 0, 4, 120000)})
	potion, other := consumableSpell(t, drinker, resourceID), consumableSpell(t, drinker, otherID)

	mcd := drinker.GetInitialMajorCooldown(potion.ActionID)
	if mcd.Spell != potion || !mcd.Type.Matches(core.CooldownTypeMana) {
		t.Fatalf("the resource potion is not a mana major cooldown: %v", mcd.Type)
	}
	if !potion.Flags.Matches(core.SpellFlagPotion | core.SpellFlagEncounterOnly) {
		t.Errorf("flags: resource potion %v", potion.Flags)
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
	gain := spelldata.MustFind(17531).ProcEnergizeEffect()
	needAmount(t, "17531's mana", gain.Max())
	if got := drinker.CurrentMana() - 100; got < gain.Min() || got > gain.Max() {
		t.Errorf("the potion restored %v, want %v to %v", got, gain.Min(), gain.Max())
	}
	if offCooldown(sim, other) {
		t.Errorf("the other potion is usable while category 4 runs")
	}
	stepPast(t, sim, start+120*time.Second)
	if !offCooldown(sim, other) || !offCooldown(sim, potion) {
		t.Errorf("the potions are not usable once category 4's 120 s ran out")
	}
}

// Mighty Rage Potion 13442 casts 17528: E_ENERGIZE into rage, stated in tenths, with a spread, and
// A_MOD_STAT Strength for the row's duration.
func TestPotionGivesRageAndItsStrength(t *testing.T) {
	const potionID int32 = 992003
	sim, drinker := newConsumableSim(t, rageUser(), &proto.ConsumesSpec{ResourcePotId: potionID}, 0,
		testConsumable{potionID, potion, testOnUse(17528, 0, 4, 120000)})
	potion := consumableSpell(t, drinker, potionID)
	strength := drinker.GetStat(stats.Strength)
	row := spelldata.MustFind(17528)
	rage, buff := row.ProcEnergizeEffect(), row.FirstAura(dbcenums.A_MOD_STAT).Mean()
	needAmount(t, "17528's rage", rage.Max())
	needAmount(t, "17528's Strength", buff)

	potion.Cast(sim, &drinker.Unit)
	if got, low, high := drinker.CurrentRage(), rage.Min()/10, rage.Max()/10; got < low || got > high {
		t.Errorf("rage after the potion is %v, want %v to %v, a tenth of the row's roll", got, low, high)
	}
	if got := drinker.GetStat(stats.Strength) - strength; got != buff {
		t.Errorf("the potion gave %v Strength, want %v", got, buff)
	}
	if potion.RelatedSelfBuff == nil || potion.RelatedSelfBuff.Duration != row.Duration() {
		t.Fatalf("the potion's buff is not the row's %v aura", row.Duration())
	}
	stepPast(t, sim, sim.CurrentTime+row.Duration()+time.Millisecond)
	if got := drinker.GetStat(stats.Strength); got != strength {
		t.Errorf("Strength after the buff ran out is %v, want %v", got, strength)
	}
}

// Demonic Rune 12662 casts 16666: E_ENERGIZE mana and E_SCHOOL_DAMAGE Shadow to the user, both with a
// spread; the test's on-use runs on conjured category 1153 for 120 s. The conjured item is a major
// cooldown only while it is the spec's chosen one.
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

	gain, damage := spelldata.MustFind(16666).ProcEnergizeEffect(), spelldata.MustFind(16666).CasterDamageEffect()
	needAmount(t, "16666's mana", gain.Max())
	needAmount(t, "16666's damage to the user", damage.Max())
	health := drinker.CurrentHealth()
	rune.Cast(sim, &drinker.Unit)
	if got := drinker.CurrentMana() - 100; got < gain.Min() || got > gain.Max() {
		t.Errorf("the rune restored %v, want %v to %v", got, gain.Min(), gain.Max())
	}
	if lost := health - drinker.CurrentHealth(); lost < damage.Min() || lost > damage.Max() {
		t.Errorf("the rune cost %v health, want %v to %v", lost, damage.Min(), damage.Max())
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

// Thistle Tea 7676 casts 9512, E_ENERGIZE into energy, here a copy with no spread. The test's on-use
// runs on a 300 s cooldown of its own and conjured category 1153 for 120 s.
func TestConjuredTeaRestoresAFlatAmountOfEnergy(t *testing.T) {
	const teaID int32 = 992006
	withoutSpread(t, 9512)
	gain := spelldata.MustFind(9512).ProcEnergizeEffect().Average()
	needAmount(t, "9512's energy", gain)
	sim, drinker := newConsumableSim(t, energyUser(), &proto.ConsumesSpec{ConjuredId: teaID, ConjuredItems: []int32{teaID}}, 0,
		testConsumable{teaID, conjured, testOnUse(9512, 300000, 1153, 120000)})
	tea := consumableSpell(t, drinker, teaID)
	mcd := drinker.GetInitialMajorCooldown(tea.ActionID)

	drinker.SpendEnergy(sim, drinker.CurrentEnergy(), drinker.EnergyRefundMetrics)
	if !mcd.ShouldActivate(sim, &drinker.Character) || !mcd.Type.Matches(core.CooldownTypeDPS) {
		t.Errorf("the tea is not a DPS cooldown used on an empty bar: %v", mcd.Type)
	}
	tea.Cast(sim, &drinker.Unit)
	if got, want := drinker.CurrentEnergy(), min(gain, drinker.MaximumEnergy()); got != want {
		t.Errorf("energy after the tea is %v, want %v", got, want)
	}
	if got := tea.CD.ReadyAt(); got != sim.CurrentTime+300*time.Second {
		t.Errorf("the tea is ready again at %v, want 300 s on", got)
	}
}

// Razorlash Root 17747 casts 21955: A_PERIODIC_HEAL and A_PERIODIC_ENERGIZE mana every period for the
// row's duration.
func TestConjuredRootHealsAndRestoresManaOverTime(t *testing.T) {
	const rootID int32 = 992007
	sim, drinker := newConsumableSim(t, manaUser(), &proto.ConsumesSpec{ConjuredItems: []int32{rootID}}, 0,
		testConsumable{rootID, conjured, testOnUse(21955, 0, 1153, 120000)})
	root := consumableSpell(t, drinker, rootID)
	row := spelldata.MustFind(21955)
	overTime := func(e *spelldata.Effect) (float64, int) {
		if e.Period() <= 0 || row.Duration() < e.Period() || e.Average() == 0 {
			t.Fatalf("21955 ticks %v every %v for %v, nothing to check", e.Average(), e.Period(), row.Duration())
		}
		return e.Average(), int(row.Duration() / e.Period())
	}
	healPer, healTicks := overTime(row.FirstAura(dbcenums.A_PERIODIC_HEAL))
	manaPer, manaTicks := overTime(row.FirstAura(dbcenums.A_PERIODIC_ENERGIZE))

	drinker.SpendMana(sim, 2000, drinker.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionManaRegen, Tag: 99}))
	root.Cast(sim, &drinker.Unit)
	stepPast(t, sim, sim.CurrentTime+row.Duration()+time.Millisecond)

	if got, want := gainedBy(drinker, rootID), manaPer*float64(manaTicks); got != want {
		t.Errorf("the root restored %v mana, want %d ticks of %v", got, manaTicks, manaPer)
	}
	if got, want := root.SpellMetrics[drinker.UnitIndex].TotalHealing, healPer*float64(healTicks); got != want {
		t.Errorf("the root healed %v, want %d ticks of %v", got, healTicks, healPer)
	}
}

// Restored Mana Potion 282013 casts 1312702, E_ENERGIZE_PCT: a share of the maximum.
func TestPotionRestoresAShareOfTheBar(t *testing.T) {
	const potionID int32 = 992008
	sim, drinker := newConsumableSim(t, manaUser(), &proto.ConsumesSpec{Potions: []int32{potionID}}, 0,
		testConsumable{potionID, potion, testOnUse(1312702, 0, 4, 120000)})
	share := spelldata.MustFind(1312702).FindEffect(dbcenums.E_ENERGIZE_PCT, 0, int32(dbcenums.POWER_MANA)).Percent()

	drinker.SpendMana(sim, drinker.CurrentMana(), drinker.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionManaRegen, Tag: 99}))
	consumableSpell(t, drinker, potionID).Cast(sim, &drinker.Unit)
	if got, want := drinker.CurrentMana(), drinker.MaxMana()*share; share <= 0 || math.Abs(got-want) > 1e-6 {
		t.Errorf("mana after the potion is %v, want %v, the row's %v of the maximum", got, want, share)
	}
}

// A test potion whose on-use states cooldown -1, category 0 and category cooldown -1, as Minor
// Spellblasting Potion 250932's does, runs on its spell's category: a copy of 1251919 states potion
// category 4 for 120 s.
func TestPotionDefersToItsSpellsCategory(t *testing.T) {
	const blastingID, manaID int32 = 992009, 992010
	editRow(t, 1251919, func(s *spelldata.Spell) { s.Category, s.CategoryCooldownMs = 4, 120000 })
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
	categoryCooldown := spelldata.MustFind(1251919).CategoryCooldown()
	stepPast(t, sim, sim.CurrentTime+categoryCooldown)
	if !offCooldown(sim, mana) {
		t.Errorf("the mana potion is not usable once category 4's %v ran out", categoryCooldown)
	}
}

// Alchemists' Stone 13503's 17619 raises the healing and mana potions give by its effect's share:
// Major Healing Potion 13446's 17534 and Major Mana Potion's 17531, not a healthstone's heal.
func TestAlchemistsStoneRaisesHealingAndManaPotions(t *testing.T) {
	stone := spelldata.MustFind(alchemistsStoneSpellID).FirstAura(dbcenums.A_DUMMY).BasePoints / 100
	if stone <= 0 {
		t.Fatalf("17619 states %v, want a share that raises potions", stone)
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
	if raised := 1 + stone; math.Abs(healedWith/healed-raised) > 1e-9 || math.Abs(restoredWith/restored-raised) > 1e-9 {
		t.Errorf("with the stone the potions gave %v and %v, want %v times %v and %v", healedWith, restoredWith, raised, healed, restored)
	}
	if stoned <= 0 || stonedWith != stoned {
		t.Errorf("the healthstone healed %v with the stone and %v without, want the same both", stonedWith, stoned)
	}
}

// Greater Fire Protection Potion 13457 casts 17543: A_SCHOOL_ABSORB of Fire on the drinker for the
// row's duration, its spread removed in a copy; the test's on-use runs on potion category 4 for 120 s.
// The defensive potion is a survival major cooldown whose shield takes fire damage only; Alchemists'
// Stone does not raise it.
func TestProtectionPotionShieldsAgainstItsSchool(t *testing.T) {
	const potionID int32 = 992021
	const fireProtection int32 = 17543
	withoutSpread(t, fireProtection)
	row := spelldata.MustFind(fireProtection)
	absorbed := row.AbsorbEffect().Average()
	needAmount(t, "17543's shield", absorbed)
	sim, drinker := newConsumableSim(t, manaUser(), &proto.ConsumesSpec{DefensivePotId: potionID}, alchemistsStoneID,
		testConsumable{potionID, potion, testOnUse(fireProtection, 0, 4, 120000)})
	potion := consumableSpell(t, drinker, potionID)
	shield := absorbAura(t, drinker, fireProtection, potionID)

	if mcd := drinker.GetInitialMajorCooldown(potion.ActionID); mcd.Spell != potion || !mcd.Type.Matches(core.CooldownTypeSurvival) {
		t.Fatalf("the protection potion is not a survival major cooldown: %v", mcd.Type)
	}

	potion.Cast(sim, &drinker.Unit)
	if !shield.IsActive() || shield.Duration != row.Duration() {
		t.Fatalf("the potion put up no shield for the row's %v: active %v, duration %v", row.Duration(), shield.IsActive(), shield.Duration)
	}
	if got := takenAfterAbsorbs(sim, drinker, core.SpellSchoolFrost, 1000); got != 1000 {
		t.Errorf("a frost hit of 1000 landed for %v, want all of it", got)
	}
	if hit := absorbed + 400; takenAfterAbsorbs(sim, drinker, core.SpellSchoolFire, hit) != 400 {
		t.Errorf("a fire hit of %v did not land for the 400 past the %v shield", hit, absorbed)
	}
	if offCooldown(sim, potion) {
		t.Errorf("the potion is usable again while category 4 runs")
	}
}

// Flame Deflector 4376 casts 4057, A_SCHOOL_ABSORB Fire; the test's on-use runs on a 900 s cooldown
// of its own with no category. Outside the potion and conjured categories it is a major cooldown
// whenever the spec names it, which the manager holds back for a player no tank slot lists.
func TestOnUseConsumableShieldsOnTheItemsCooldown(t *testing.T) {
	const deflectorID int32 = 992022
	sim, user := newConsumableSim(t, manaUser(), &proto.ConsumesSpec{ExtraIds: []int32{deflectorID}}, 0,
		testConsumable{deflectorID, other, testOnUse(4057, 900000, 0, 0)})
	deflector := consumableSpell(t, user, deflectorID)

	mcd := user.GetInitialMajorCooldown(deflector.ActionID)
	if mcd.Spell != deflector || !mcd.Type.Matches(core.CooldownTypeSurvival) {
		t.Fatalf("the deflector is not a survival major cooldown: %v", mcd.Type)
	}
	if mcd.ShouldActivate(sim, &user.Character) {
		t.Errorf("the manager uses the deflector for a player no tank slot lists")
	}
	if deflector.Flags.Matches(core.SpellFlagPotion) || deflector.Flags.Matches(core.SpellFlagConjured) {
		t.Errorf("the deflector is flagged a potion or a conjured item: %v", deflector.Flags)
	}

	deflector.Cast(sim, &user.Unit)
	absorb := spelldata.MustFind(4057).AbsorbEffect()
	needAmount(t, "4057's shield", absorb.Min())
	hit := absorb.Max() + 300
	if got, low, high := takenAfterAbsorbs(sim, user, core.SpellSchoolFire, hit), hit-absorb.Max(), hit-absorb.Min(); got < low || got > high {
		t.Errorf("a fire hit of %v landed for %v, want %v to %v past the shield", hit, got, low, high)
	}
	if deflector.CD.Duration != 900*time.Second || deflector.SharedCD.Timer != nil {
		t.Errorf("the deflector runs on a %v cooldown and shared timer %v, want its own 900 s alone",
			deflector.CD.Duration, deflector.SharedCD.Timer)
	}
}

// The spell shapes the builder cannot run, with what stops it, and the shapes it runs.
func TestConsumableUnsupportedNamesWhatStopsIt(t *testing.T) {
	onUser := [2]dbcenums.ImplicitTarget{dbcenums.TARGET_UNIT_CASTER}
	mana, rage := int32(dbcenums.POWER_MANA), int32(dbcenums.POWER_RAGE)
	effect := func(typ dbcenums.SpellEffectType, misc int32) spelldata.Effect {
		return spelldata.Effect{Type: typ, Misc: misc, BasePoints: 100, Target: onUser}
	}
	aura := func(a dbcenums.EffectAuraType, misc int32, periodMs int32) spelldata.Effect {
		return spelldata.Effect{Type: dbcenums.E_APPLY_AURA, Aura: a, Misc: misc, BasePoints: 100, PeriodMs: periodMs, Target: onUser}
	}
	onEnemy := effect(dbcenums.E_ENERGIZE, mana)
	onEnemy.Target[0] = dbcenums.TARGET_UNIT_TARGET_ENEMY
	hitOnEnemy := effect(dbcenums.E_SCHOOL_DAMAGE, 0)
	hitOnEnemy.Target[0] = dbcenums.TARGET_UNIT_TARGET_ENEMY

	for _, tc := range []struct {
		name string
		row  spelldata.Spell
		want string
	}{
		{"a dummy effect", spelldata.Spell{Effects: []spelldata.Effect{effect(dbcenums.E_DUMMY, 0)}}, "E_DUMMY"},
		{"a gain beside damage to an enemy", spelldata.Spell{Effects: []spelldata.Effect{hitOnEnemy, effect(dbcenums.E_ENERGIZE, mana)}},
			"E_ENERGIZE lands on the user beside the damage"},
		{"an effect on an enemy", spelldata.Spell{Effects: []spelldata.Effect{onEnemy}}, "not the user"},
		{"a stun", spelldata.Spell{DurationMs: 5000, Effects: []spelldata.Effect{aura(dbcenums.A_MOD_STUN, 0, 0)}}, "A_MOD_STUN"},
		{"damage over time on the user", spelldata.Spell{DurationMs: 15000,
			Effects: []spelldata.Effect{aura(dbcenums.A_PERIODIC_DAMAGE, 0, 3000)}}, "A_PERIODIC_DAMAGE"},
		{"a triggered spell", spelldata.Spell{Effects: []spelldata.Effect{effect(dbcenums.E_TRIGGER_SPELL, 0)}}, "E_TRIGGER_SPELL"},
		// 0x40000 is the game files' interrupt-on-standing flag in AuraInterruptFlags_0.
		{"a gain only while seated", spelldata.Spell{AuraInterrupt: [2]uint32{0x40000},
			Effects: []spelldata.Effect{effect(dbcenums.E_ENERGIZE, mana)}}, "sits"},
	} {
		got := strings.Join(ConsumableUnsupported(&tc.row), "; ")
		if !strings.Contains(got, tc.want) {
			t.Errorf("%s: unsupported %q, want it to name %s", tc.name, got, tc.want)
		}
	}

	for _, tc := range []struct {
		name string
		row  spelldata.Spell
	}{
		{"a mana gain", spelldata.Spell{Effects: []spelldata.Effect{effect(dbcenums.E_ENERGIZE, mana)}}},
		{"rage and a stat for a while", spelldata.Spell{DurationMs: 20000,
			Effects: []spelldata.Effect{effect(dbcenums.E_ENERGIZE, rage), aura(dbcenums.A_MOD_STAT, 0, 0)}}},
		{"mana at the cost of life", spelldata.Spell{
			Effects: []spelldata.Effect{effect(dbcenums.E_ENERGIZE, mana), effect(dbcenums.E_SCHOOL_DAMAGE, 0)}}},
		{"a heal", spelldata.Spell{Effects: []spelldata.Effect{effect(dbcenums.E_HEAL, 0)}}},
		{"a heal and mana over time", spelldata.Spell{DurationMs: 10000,
			Effects: []spelldata.Effect{aura(dbcenums.A_PERIODIC_HEAL, 0, 1000), aura(dbcenums.A_PERIODIC_ENERGIZE, mana, 1000)}}},
		{"a share of the bar", spelldata.Spell{Effects: []spelldata.Effect{effect(dbcenums.E_ENERGIZE_PCT, mana)}}},
		{"a shield", spelldata.Spell{DurationMs: 60000, Effects: []spelldata.Effect{aura(dbcenums.A_SCHOOL_ABSORB, 4, 0)}}},
		{"damage to an enemy", spelldata.Spell{Effects: []spelldata.Effect{hitOnEnemy}}},
	} {
		if got := ConsumableUnsupported(&tc.row); got != nil {
			t.Errorf("%s: unsupported %v, want none", tc.name, got)
		}
	}
}

// Caustic Smog Potion's 1252525 leaves the cloud 1252526, which puts 1252543 on the enemies inside
// it. The cloud has to state how long it lasts, and its damage has to state Always Hit, since the
// cloud rolls none. Disorienting Smog Potion's cloud deals no damage.
func TestADamageCloudNamesWhatStopsIt(t *testing.T) {
	cloud, damage := DamageCloudOf(spelldata.MustFind(1252525))
	if cloud.ID != 1252526 || damage.ID != 1252543 {
		t.Fatalf("1252525 leaves cloud %d dealing %d, want 1252526 dealing 1252543", cloud.ID, damage.ID)
	}
	if !slices.Contains(spelldata.MustFind(1252526).Triggered(), damage) {
		t.Errorf("1252526 triggers %v, want 1252543 among them", spelldata.MustFind(1252526).Triggered())
	}
	if cloud, _ := DamageCloudOf(spelldata.MustFind(1252544)); cloud != spelldata.Nil || ConsumableDamages(spelldata.MustFind(1252544)) {
		t.Errorf("Disorienting Smog Potion's 1252544 leaves damage cloud %d, want none", cloud.ID)
	}
	for _, id := range []int32{1252525, 1252554} {
		if got := ConsumableUnsupported(spelldata.MustFind(id)); got != nil || !ConsumableDealsDamage(spelldata.MustFind(id)) {
			t.Errorf("%d: unsupported %v, deals damage %v; want the damage, supported", id, got, ConsumableDealsDamage(spelldata.MustFind(id)))
		}
	}

	for _, tc := range []struct {
		name string
		id   int32
		edit func(*spelldata.Spell)
		want string
	}{
		{"a cloud of no duration", 1252526, func(s *spelldata.Spell) { s.DurationMs = 0 }, "states no duration"},
		{"damage that rolls a hit", 1252543, func(s *spelldata.Spell) { s.Attr[dbcenums.ATTR_INDEX_EX_3] &^= dbcenums.ATTR_EX_3_ALWAYS_HIT }, "rolls a hit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			editRow(t, tc.id, tc.edit)
			if got := strings.Join(ConsumableUnsupported(spelldata.MustFind(1252525)), "; "); !strings.Contains(got, tc.want) {
				t.Errorf("unsupported %q, want it to name %s", got, tc.want)
			}
		})
	}
}

// A damage potion is an offensive potion: a potion of category 4, never an explosive, and a DPS major
// cooldown only as the spec's offensive potion pick.
func TestADamagePotionIsAnOffensivePotion(t *testing.T) {
	for i, row := range []int32{1252525, 1252554} {
		pickedID, otherID := 993001+2*int32(i), 993002+2*int32(i)
		sim, drinker := newConsumableSim(t, manaUser(), &proto.ConsumesSpec{OffensivePotId: pickedID, Potions: []int32{otherID}}, 0,
			testConsumable{pickedID, potion, testOnUse(row, 0, 4, 120000)},
			testConsumable{otherID, potion, testOnUse(row, 0, 4, 120000)})
		picked, other := consumableSpell(t, drinker, pickedID), consumableSpell(t, drinker, otherID)

		for _, spell := range []*core.Spell{picked, other} {
			if !spell.Flags.Matches(core.SpellFlagPotion|core.SpellFlagEncounterOnly) || spell.Flags.Matches(core.SpellFlagExplosive) {
				t.Errorf("%d: flags %v, want a potion used in the encounter, not an explosive", row, spell.Flags)
			}
		}
		mcd := drinker.GetInitialMajorCooldown(picked.ActionID)
		if mcd.Spell != picked || !mcd.Type.Matches(core.CooldownTypeDPS) || mcd.Type.Matches(core.CooldownTypeExplosive) {
			t.Errorf("%d: the picked potion is a major cooldown of type %v, want a DPS one that is no explosive", row, mcd.Type)
		}
		if drinker.GetInitialMajorCooldown(other.ActionID).Spell != nil {
			t.Errorf("%d: a potion the spec does not pick is a major cooldown", row)
		}
		if !picked.Cast(sim, sim.Encounter.ActiveTargetUnits[0]) {
			t.Errorf("%d: the picked potion could not be used", row)
		}
	}
}

// Caustic Smog Potion's 1252525 flies at its Speed for no less than its MinDuration, and the cloud
// 1252526 leaves puts 1252543 on every enemy for the cloud's duration. 1252543 states Always Hit, so
// it lands on each enemy even for a drinker who cannot hit anything, and every tick rolls its own
// amount. The potion shares category 4 with a mana potion.
func TestCausticSmogPotionPoisonsEveryEnemyInTheCloud(t *testing.T) {
	row := spelldata.MustFind(1252525)
	cloud, damage := DamageCloudOf(row)
	tick := damage.PeriodicDamageEffect()
	ticks := int(cloud.Duration() / tick.Period())
	travel := func(distance float64) time.Duration {
		return max(row.MissileMinDuration(), time.Duration(distance/float64(row.Speed)*float64(time.Second)))
	}

	for i, distance := range []float64{0, 5, float64(row.MaxRange)} {
		causticID, manaID := 993010+2*int32(i), 993011+2*int32(i)
		sim, drinker := newConsumableSimAgainst(t, manaUser(), &proto.ConsumesSpec{OffensivePotId: causticID, Potions: []int32{manaID}}, 0, 3,
			testConsumable{causticID, potion, testOnUse(row.ID, 0, 4, 120000)},
			testConsumable{manaID, potion, testOnUse(17531, 0, 4, 120000)})
		drinker.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: -1000})
		caustic, mana := consumableSpell(t, drinker, causticID), consumableSpell(t, drinker, manaID)
		poison := drinker.GetSpell(core.ActionID{SpellID: damage.ID})
		targets := sim.Encounter.ActiveTargetUnits

		drinker.DistanceFromTarget = distance
		start := sim.CurrentTime
		if !caustic.Cast(sim, targets[0]) {
			t.Fatalf("at %v yd: the potion could not be used", distance)
		}
		if offCooldown(sim, mana) {
			t.Errorf("at %v yd: the mana potion is usable while category 4 runs", distance)
		}
		lands := start + travel(distance)
		stepPast(t, sim, lands-time.Millisecond)
		for i, target := range targets {
			if poison.Dot(target).IsActive() {
				t.Errorf("at %v yd: target %d is poisoned before the missile's %v flight ended", distance, i, travel(distance))
			}
		}
		stepPast(t, sim, lands+time.Millisecond)
		for i, target := range targets {
			if !poison.Dot(target).IsActive() {
				t.Errorf("at %v yd: target %d is not poisoned once the missile landed after %v", distance, i, travel(distance))
			}
		}

		var dealt []float64
		for n := 1; n <= ticks; n++ {
			dealt = append(dealt, dealtUntil(t, sim, poison, lands, time.Duration(n)*tick.Period()))
		}
		stepPast(t, sim, lands+cloud.Duration()+tick.Period())
		for i, d := range dealt {
			if d < tick.Min() || d > tick.Max() {
				t.Errorf("at %v yd: tick %d dealt %v, want a roll of %v to %v", distance, i+1, d, tick.Min(), tick.Max())
			}
		}
		if slices.Min(dealt) == slices.Max(dealt) {
			t.Errorf("at %v yd: every tick dealt %v, want a roll of its own each", distance, dealt[0])
		}
		for i, target := range targets {
			metrics := poison.SpellMetrics[target.UnitIndex]
			if metrics.Ticks != int32(ticks) || metrics.Misses != 0 {
				t.Errorf("at %v yd: target %d took %d ticks and %d misses, want the cloud's %d ticks", distance, i, metrics.Ticks, metrics.Misses, ticks)
			}
		}

		stepPast(t, sim, start+row.CategoryCooldown())
		if !offCooldown(sim, mana) {
			t.Errorf("at %v yd: the mana potion is not usable once category 4's %v ran out", distance, row.CategoryCooldown())
		}
	}
}

// Dragonfire Potion's 1252554 burns every enemy in its cone with a hit and a damage over time,
// landing at the row's Speed, and reaches as far as the cone's radius.
func TestDragonfirePotionBurnsEveryEnemyInItsCone(t *testing.T) {
	const dragonfireID int32 = 993020
	row := spelldata.MustFind(1252554)
	direct, burn := row.DamageEffect(), row.PeriodicDamageEffect()
	ticks := int(row.Duration() / burn.Period())
	sim, drinker := newConsumableSimAgainst(t, manaUser(), &proto.ConsumesSpec{OffensivePotId: dragonfireID}, 0, 3,
		testConsumable{dragonfireID, potion, testOnUse(row.ID, 0, 4, 120000)})
	dragonfire := consumableSpell(t, drinker, dragonfireID)
	targets := sim.Encounter.ActiveTargetUnits

	drinker.DistanceFromTarget = float64(direct.RadiusMax) + 1
	if dragonfire.CanCast(sim, targets[0]) {
		t.Errorf("the potion can be used from %v yd, past its cone's %v", drinker.DistanceFromTarget, direct.RadiusMax)
	}
	drinker.DistanceFromTarget = float64(direct.RadiusMax)
	if !dragonfire.CanCast(sim, targets[0]) {
		t.Fatalf("the potion cannot be used from its cone's %v yd", direct.RadiusMax)
	}

	start := sim.CurrentTime
	dragonfire.Cast(sim, targets[0])
	lands := start + time.Duration(drinker.DistanceFromTarget/float64(row.Speed)*float64(time.Second))
	stepPast(t, sim, lands-time.Millisecond)
	for i, target := range targets {
		if got := dragonfire.SpellMetrics[target.UnitIndex].TotalDamage; got != 0 || dragonfire.Dot(target).IsActive() {
			t.Errorf("target %d took %v before the fire's flight ended at %v", i, got, lands-start)
		}
	}

	stepPast(t, sim, lands+row.Duration()+time.Millisecond)
	for i, target := range targets {
		metrics := dragonfire.SpellMetrics[target.UnitIndex]
		if metrics.Hits != 1 || metrics.Misses != 0 || metrics.Ticks != int32(ticks) {
			t.Errorf("target %d took %d hits, %d misses and %d ticks; want the hit and %d ticks", i, metrics.Hits, metrics.Misses, metrics.Ticks, ticks)
		}
		if metrics.TotalDamage <= metrics.TotalTickDamage {
			t.Errorf("target %d took no damage from the hit", i)
		}
		if metrics.TotalTickDamage < float64(ticks)*burn.Min() || metrics.TotalTickDamage > float64(ticks)*burn.Max() {
			t.Errorf("target %d took %v over time, want %d ticks of %v to %v", i, metrics.TotalTickDamage, ticks, burn.Min(), burn.Max())
		}
	}
}

package shared

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The items' on-use spells, on cooldowns and categories of the test's own.
func renatakisOnUse() *proto.ItemEffect { return testOnUse(24532, 180000, 1141, 10000) } // Renataki's Charm of Trickery 19954
func grileksOnUse() *proto.ItemEffect   { return testOnUse(24571, 180000, 1141, 10000) } // Gri'lek's Charm of Might 19951
func robeOnUse() *proto.ItemEffect      { return testOnUse(18385, 300000, 0, 0) }        // Robe of the Archmage 14152
func warmthOnUse() *proto.ItemEffect    { return testOnUse(28760, 180000, 0, 0) }        // Warmth of Forgiveness 23027
func sigilOnUse() *proto.ItemEffect     { return testOnUse(24884, 180000, 0, 0) }        // Earthen Sigil 20525

func init() {
	registerBarTester(proto.Player_Rogue{}, proto.Spec_SpecRogue, func(c *testCaster) {
		c.EnableEnergyBar(core.EnergyBarOptions{MaxEnergy: 100, MaxComboPoints: 5, UnitClass: proto.Class_ClassRogue})
	}, func(player *proto.Player, spec interface{}) { player.Spec = spec.(*proto.Player_Rogue) })
	registerBarTester(proto.Player_DpsWarrior{}, proto.Spec_SpecDpsWarrior, func(c *testCaster) {
		c.EnableRageBar(core.RageBarOptions{MaxRage: 100, BaseRageMultiplier: 1})
	}, func(player *proto.Player, spec interface{}) { player.Spec = spec.(*proto.Player_DpsWarrior) })
	registerBarTester(proto.Player_Mage{}, proto.Spec_SpecMage, func(c *testCaster) {
		c.EnableManaBar()
	}, func(player *proto.Player, spec interface{}) { player.Spec = spec.(*proto.Player_Mage) })
}

func registerBarTester(options interface{}, spec proto.Spec, enable func(*testCaster), setSpec func(*proto.Player, interface{})) {
	core.RegisterAgentFactory(options, spec, func(character *core.Character, _ *proto.Player, _ *proto.Raid) core.Agent {
		c := &testCaster{Character: *character}
		enable(c)
		return c
	}, setSpec)
}

func testRogue() *proto.Player {
	return &proto.Player{Class: proto.Class_ClassRogue, Race: proto.Race_RaceHuman, Spec: &proto.Player_Rogue{}}
}

func testWarrior() *proto.Player {
	return &proto.Player{Class: proto.Class_ClassWarrior, Race: proto.Race_RaceHuman, Spec: &proto.Player_DpsWarrior{}}
}

func testMage() *proto.Player {
	return &proto.Player{Class: proto.Class_ClassMage, Race: proto.Race_RaceGnome, Spec: &proto.Player_Mage{}}
}

type testTrinket struct {
	id       int32
	register func(int32)
	effect   *proto.ItemEffect
}

// The player wearing the trinkets, each registered through its own constructor.
func newBarSim(t *testing.T, player *proto.Player, trinkets ...testTrinket) (*core.Simulation, *testCaster) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotTrinket2+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	for i, trinket := range trinkets {
		core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{{Id: trinket.id, Name: "Test Trinket",
			Type: proto.ItemType_ItemTypeTrinket, ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}}}})
		RegisterItemEffects(map[int32][]*proto.ItemEffect{trinket.id: {trinket.effect}})
		trinket.register(trinket.id)
		items[int(proto.ItemSlot_ItemSlotTrinket1)+i] = &proto.ItemSpec{Id: trinket.id}
	}

	player.Name = "Wearer"
	player.Buffs = &proto.IndividualBuffs{}
	player.Equipment = &proto.EquipmentSpec{Items: items}
	player.Rotation = &proto.APLRotation{}

	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{player},
			Buffs:   &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{
			Targets:  []*proto.Target{{Name: "target", Level: 60, MobType: proto.MobType_MobTypeDemon}},
			Duration: 600,
		},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*testCaster)
}

// What the item's gains added up to, and how many there were.
func gainedFrom(caster *testCaster, itemID int32) (float64, int32) {
	for _, r := range caster.Metrics.ToProto().Resources {
		if r.Id.GetItemId() == itemID {
			return r.ActualGain, r.Events
		}
	}
	return 0, 0
}

// Renataki's Charm of Trickery 19954 casts 24532, E_ENERGIZE into energy (misc 3). The test's on-use
// runs on a 3 min cooldown in category 1141, which the test's lasso shares.
func TestOnUseEnergizeRestoresARoguesEnergy(t *testing.T) {
	const charmID, lassoID int32 = 991301, 991302
	sim, rogue := newBarSim(t, testRogue(),
		testTrinket{charmID, NewSpellDataEnergizeOnUse, renatakisOnUse()},
		testTrinket{lassoID, NewSpellDataDamageOnUse, lassoOnUse()})
	charm, lasso := onUseSpell(t, rogue, charmID), onUseSpell(t, rogue, lassoID)
	gain := spelldata.MustFind(24532).ProcEnergizeEffect().Average()
	needAmount(t, "24532's energy", gain)

	mcd := rogue.GetInitialMajorCooldown(charm.ActionID)
	if !mcd.Type.Matches(core.CooldownTypeDPS) {
		t.Errorf("the charm is a major cooldown of type %v, want a DPS one", mcd.Type)
	}
	if mcd.ShouldActivate(sim, &rogue.Character) {
		t.Errorf("the charm would be used on a full bar")
	}

	const first = 20.0
	rogue.SpendEnergy(sim, rogue.CurrentEnergy()-first, rogue.EnergyRefundMetrics)
	if got, fits := mcd.ShouldActivate(sim, &rogue.Character), rogue.MaximumEnergy()-first >= gain; got != fits {
		t.Errorf("with %v energy missing, using the charm for %v is %v, want %v",
			rogue.MaximumEnergy()-first, gain, got, fits)
	}
	if !charm.Cast(sim, &rogue.Unit) {
		t.Fatalf("the charm could not be used")
	}
	start := sim.CurrentTime
	firstGain := min(first+gain, rogue.MaximumEnergy()) - first
	if got := rogue.CurrentEnergy(); got != first+firstGain {
		t.Errorf("energy after the charm is %v, want %v + %v", got, first, firstGain)
	}
	if offCooldown(sim, lasso) {
		t.Errorf("the lasso could be used while category 1141's cooldown runs")
	}

	stepPast(t, sim, start+10*time.Second)
	if !offCooldown(sim, lasso) {
		t.Errorf("the lasso could not be used once category 1141's 10 s ran out")
	}
	if offCooldown(sim, charm) {
		t.Errorf("the charm could be used again inside its 3 min cooldown")
	}
	if got := charm.CD.ReadyAt(); got != start+3*time.Minute {
		t.Errorf("the charm is ready again at %v, want 3 min after its use at %v", got, start)
	}

	stepPast(t, sim, start+3*time.Minute)
	const second = 70.0
	rogue.SpendEnergy(sim, rogue.CurrentEnergy()-second, rogue.EnergyRefundMetrics)
	charm.Cast(sim, &rogue.Unit)
	secondGain := min(second+gain, rogue.MaximumEnergy()) - second
	if got := rogue.CurrentEnergy(); got != second+secondGain {
		t.Errorf("energy after a second use from %v is %v, want %v, capped at the %v maximum",
			second, got, second+secondGain, rogue.MaximumEnergy())
	}
	if gained, events := gainedFrom(rogue, charmID); gained != firstGain+secondGain || events != 2 {
		t.Errorf("the charm's metrics read %v energy over %d uses, want %v + %v over 2", gained, events, firstGain, secondGain)
	}
}

// Gri'lek's Charm of Might 19951 casts 24571, E_ENERGIZE into rage (misc 1): the game files state
// rage in tenths, and the tooltip reads a tenth of the row's amount.
func TestOnUseEnergizeGivesAWarriorTheTooltipsRage(t *testing.T) {
	const charmID int32 = 991303
	sim, warrior := newBarSim(t, testWarrior(), testTrinket{charmID, NewSpellDataEnergizeOnUse, grileksOnUse()})
	charm := onUseSpell(t, warrior, charmID)

	if got := warrior.GetInitialMajorCooldown(charm.ActionID); !got.Type.Matches(core.CooldownTypeDPS) {
		t.Errorf("the charm is a major cooldown of type %v, want a DPS one", got.Type)
	}
	want := spelldata.MustFind(24571).ProcEnergizeEffect().Average() / 10
	needAmount(t, "24571's rage", want)
	charm.Cast(sim, &warrior.Unit)
	if got := warrior.CurrentRage(); got != want {
		t.Errorf("rage after the charm is %v, want %v, a tenth of the row's amount", got, want)
	}
}

// Warmth of Forgiveness 23027 casts 28760, E_ENERGIZE mana; Robe of the Archmage 14152 casts 18385,
// a mana gain rolled over a spread.
func TestOnUseEnergizeRestoresAManaUsersMana(t *testing.T) {
	const warmthID, robeID int32 = 991304, 991305
	sim, mage := newBarSim(t, testMage(),
		testTrinket{warmthID, NewSpellDataEnergizeOnUse, warmthOnUse()},
		testTrinket{robeID, NewSpellDataEnergizeOnUse, robeOnUse()})
	warmth, robe := onUseSpell(t, mage, warmthID), onUseSpell(t, mage, robeID)
	spent := mage.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionManaRegen, Tag: 99})

	for _, spell := range []*core.Spell{warmth, robe} {
		mcd := mage.GetInitialMajorCooldown(spell.ActionID)
		if !mcd.Type.Matches(core.CooldownTypeMana) || !spell.Flags.Matches(core.SpellFlagHelpful) {
			t.Errorf("%v is a major cooldown of type %v, helpful %v; want a mana one cast on the wearer",
				spell.ActionID, mcd.Type, spell.Flags.Matches(core.SpellFlagHelpful))
		}
		if mcd.ShouldActivate(sim, &mage.Character) {
			t.Errorf("%v would be used on a full bar", spell.ActionID)
		}
	}

	want := spelldata.MustFind(28760).ProcEnergizeEffect().Average()
	needAmount(t, "28760's mana", want)
	mage.SpendMana(sim, 2000, spent)
	before := mage.CurrentMana()
	warmth.Cast(sim, &mage.Unit)
	if got := mage.CurrentMana() - before; got != want {
		t.Errorf("Warmth of Forgiveness restored %v, want %v", got, want)
	}

	effect := spelldata.MustFind(18385).ProcEnergizeEffect()
	before = mage.CurrentMana()
	robe.Cast(sim, &mage.Unit)
	if got, low, high := mage.CurrentMana()-before, effect.Min(), effect.Max(); got < low || got > high {
		t.Errorf("the robe restored %v, want %v to %v", got, low, high)
	}
}

// Earthen Sigil 20525 casts 24884, A_PERIODIC_ENERGIZE mana every period for the row's duration.
func TestOnUseEnergizeRestoresManaOverTime(t *testing.T) {
	const sigilID int32 = 991306
	sim, mage := newBarSim(t, testMage(), testTrinket{sigilID, NewSpellDataEnergizeOnUse, sigilOnUse()})
	sigil := onUseSpell(t, mage, sigilID)
	row := spelldata.MustFind(24884)
	per, _, ticks := ticksOf(t, row)

	mage.SpendMana(sim, 2000, mage.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionManaRegen, Tag: 99}))
	sigil.Cast(sim, &mage.Unit)
	start := sim.CurrentTime
	stepPast(t, sim, start+row.Duration())

	if sigil.SelfHot().IsActive() {
		t.Errorf("the gain over time is still up after its %v", row.Duration())
	}
	if gained, events := gainedFrom(mage, sigilID); gained != per*float64(ticks) || events != int32(ticks) {
		t.Errorf("the sigil restored %v over %d ticks, want %v over %d", gained, events, per*float64(ticks), ticks)
	}
}

// A character without the bar the gain fills does not register it: a rogue has no mana, a warrior no
// energy, a mage no rage.
func TestOnUseEnergizeSkipsACharacterWithoutTheBar(t *testing.T) {
	for _, tc := range []struct {
		name   string
		player *proto.Player
		itemID int32
		effect *proto.ItemEffect
	}{
		{"rogue wearing Warmth of Forgiveness", testRogue(), 991307, warmthOnUse()},
		{"warrior wearing Renataki's Charm", testWarrior(), 991308, renatakisOnUse()},
		{"mage wearing Gri'lek's Charm", testMage(), 991309, grileksOnUse()},
	} {
		_, wearer := newBarSim(t, tc.player, testTrinket{tc.itemID, NewSpellDataEnergizeOnUse, tc.effect})
		actionID := core.ActionID{ItemID: tc.itemID}
		if wearer.GetSpell(actionID) != nil || wearer.GetInitialMajorCooldown(actionID).Spell != nil {
			t.Errorf("%s: the item registered an on-use", tc.name)
		}
	}
}

package shared

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

const (
	smokedDesertDumplingsID int32 = 20452
	elixirOfTheMongooseID   int32 = 13452
	scrollOfIntellectIVID   int32 = 10308
	swiftfeatherTokenID     int32 = 1400
)

func registerStatConsumable(t *testing.T, id int32) {
	t.Helper()
	if core.GetConsumableByID(id).Id == 0 {
		t.Skipf("consumable %d is not in the loaded database", id)
	}
	if !core.HasConsumableEffect(id) {
		NewSpellDataStatConsumable(id)
	}
}

func newConsumablesSim(consumes *proto.ConsumesSpec, raidBuffs *proto.RaidBuffs) (*core.Simulation, *core.Character) {
	player := &proto.Player{
		Name:        "Caster",
		Class:       proto.Class_ClassShaman,
		Race:        proto.Race_RaceTroll,
		Buffs:       &proto.IndividualBuffs{},
		Spec:        &proto.Player_ElementalShaman{},
		Equipment:   &proto.EquipmentSpec{},
		Consumables: consumes,
	}

	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{
			Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}},
			Buffs:   raidBuffs,
		},
		Encounter: &proto.Encounter{
			Targets:  []*proto.Target{{Name: "target", Level: 60}},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

func statGain(consumes *proto.ConsumesSpec, raidBuffs *proto.RaidBuffs) stats.Stats {
	_, without := newConsumablesSim(&proto.ConsumesSpec{}, raidBuffs)
	_, with := newConsumablesSim(consumes, raidBuffs)
	return with.GetStats().Subtract(without.GetStats())
}

// Nutritious Food 1248401 hands its 20 to Well Fed 1248422, whose own row states 1 Strength. Well
// Fed persists through death, so it is up for the whole iteration.
func TestWellFedTakesTheValueItsTriggerHandsOver(t *testing.T) {
	registerStatConsumable(t, smokedDesertDumplingsID)

	gain := statGain(&proto.ConsumesSpec{FoodId: smokedDesertDumplingsID}, &proto.RaidBuffs{})
	if gain[stats.Strength] != 20 {
		t.Errorf("Smoked Desert Dumplings add %v Strength, want 20", gain[stats.Strength])
	}

	_, character := newConsumablesSim(&proto.ConsumesSpec{FoodId: smokedDesertDumplingsID}, &proto.RaidBuffs{})
	aura := character.GetAuraByID(core.ActionID{ItemID: smokedDesertDumplingsID})
	if aura == nil || !aura.IsActive() || aura.Duration != core.NeverExpires {
		t.Errorf("Well Fed is %+v, want up for the whole iteration", aura)
	}
}

// 17538 states 25 Agility and 2% critical strike chance for 30 minutes; the flat stats the catalogue
// carries name only the Agility.
func TestElixirTakesEveryEffectAndTheDurationOfItsRow(t *testing.T) {
	registerStatConsumable(t, elixirOfTheMongooseID)

	gain := statGain(&proto.ConsumesSpec{ElixirIds: []int32{elixirOfTheMongooseID}}, &proto.RaidBuffs{})
	if gain[stats.Agility] != 25 || gain[stats.PhysicalCritPercent] < 2-1e-9 {
		t.Errorf("Elixir of the Mongoose adds %v Agility and %v%% crit, want 25 and at least 2",
			gain[stats.Agility], gain[stats.PhysicalCritPercent])
	}

	_, character := newConsumablesSim(&proto.ConsumesSpec{ElixirIds: []int32{elixirOfTheMongooseID}}, &proto.RaidBuffs{})
	aura := character.GetAuraByID(core.ActionID{ItemID: elixirOfTheMongooseID})
	if aura == nil || !aura.IsActive() || aura.Duration != 30*time.Minute {
		t.Errorf("Elixir of the Mongoose is %+v, want up from the pull for 30 minutes", aura)
	}
}

// Elixir of the Grizzly 250351 (1250986) states 25 Strength and the same 2% crit as Elixir of the
// Mongoose: both stats apply, the crit once.
func TestOnlyOneElixirsCritApplies(t *testing.T) {
	const grizzly int32 = 250351
	registerStatConsumable(t, elixirOfTheMongooseID)
	registerStatConsumable(t, grizzly)

	mongoose := statGain(&proto.ConsumesSpec{ElixirIds: []int32{elixirOfTheMongooseID}}, &proto.RaidBuffs{})
	both := statGain(&proto.ConsumesSpec{ElixirIds: []int32{elixirOfTheMongooseID, grizzly}}, &proto.RaidBuffs{})
	if both[stats.Agility] != 25 || both[stats.Strength] != 25 {
		t.Errorf("Mongoose and Grizzly add %v Agility and %v Strength, want 25 and 25", both[stats.Agility], both[stats.Strength])
	}
	if both[stats.PhysicalCritPercent] != mongoose[stats.PhysicalCritPercent] || both[stats.SpellCritPercent] != mongoose[stats.SpellCritPercent] {
		t.Errorf("Mongoose and Grizzly add %v%% physical and %v%% spell crit, want Mongoose's alone: %v%% and %v%%",
			both[stats.PhysicalCritPercent], both[stats.SpellCritPercent], mongoose[stats.PhysicalCritPercent], mongoose[stats.SpellCritPercent])
	}
}

// Scroll of Intellect IV states 16 Intellect, which Arcane Brilliance's 40 overrides rather than adds
// to.
func TestScrollIsExclusiveWithTheRaidBuffOfItsStat(t *testing.T) {
	registerStatConsumable(t, scrollOfIntellectIVID)

	if gain := statGain(&proto.ConsumesSpec{ExtraIds: []int32{scrollOfIntellectIVID}}, &proto.RaidBuffs{}); gain[stats.Intellect] != 16 {
		t.Errorf("Scroll of Intellect IV adds %v Intellect, want 16", gain[stats.Intellect])
	}

	brilliance := &proto.RaidBuffs{ArcaneBrilliance: true}
	if gain := statGain(&proto.ConsumesSpec{ExtraIds: []int32{scrollOfIntellectIVID}}, brilliance); gain[stats.Intellect] != 0 {
		t.Errorf("Scroll of Intellect IV adds %v Intellect under Arcane Brilliance, want 0", gain[stats.Intellect])
	}
}

// Swiftfeather Token's buff lasts 30 s against the 5 min category cooldown the item shares, so it is
// used in combat rather than put up before the pull.
func TestShortBuffIsUsedOnItsCooldown(t *testing.T) {
	registerStatConsumable(t, swiftfeatherTokenID)

	_, character := newConsumablesSim(&proto.ConsumesSpec{FoodId: swiftfeatherTokenID}, &proto.RaidBuffs{})
	mcd := character.GetMajorCooldown(core.ActionID{ItemID: swiftfeatherTokenID})
	if mcd == nil {
		t.Fatal("Swiftfeather Token registers no major cooldown")
	}
	if mcd.Spell.SharedCD.Duration != 5*time.Minute {
		t.Errorf("Swiftfeather Token shares a %v cooldown, want 5m", mcd.Spell.SharedCD.Duration)
	}

	aura := character.GetAuraByID(core.ActionID{ItemID: swiftfeatherTokenID})
	if aura == nil || aura.IsActive() || aura.Duration != 30*time.Second {
		t.Errorf("Swiftfeather Token's buff is %+v, want down at the pull and lasting 30s", aura)
	}
}

// The Zanza group holds Spirit of Zanza 20079 (24382, 50 Spirit and 50 Stamina) and the Blasted Lands
// buffs, R.O.I.D.S. 8410 (10667, 25 Strength) and Ground Scorpok Assay 8412 (10669, 25 Agility) among
// them: only the first the spec names is up.
func TestOnlyTheFirstOfTheZanzaGroupApplies(t *testing.T) {
	const roids, scorpok, zanza int32 = 8410, 8412, 20079
	for _, id := range []int32{roids, scorpok, zanza} {
		registerStatConsumable(t, id)
	}

	for _, tc := range []struct {
		name  string
		ids   []int32
		stat  stats.Stat
		gain  float64
		other stats.Stat
	}{
		{"R.O.I.D.S. then Ground Scorpok Assay", []int32{roids, scorpok}, stats.Strength, 25, stats.Agility},
		{"Ground Scorpok Assay then R.O.I.D.S.", []int32{scorpok, roids}, stats.Agility, 25, stats.Strength},
		{"Spirit of Zanza then R.O.I.D.S.", []int32{zanza, roids}, stats.Spirit, 50, stats.Strength},
		{"R.O.I.D.S. then Spirit of Zanza", []int32{roids, zanza}, stats.Strength, 25, stats.Spirit},
	} {
		gain := statGain(&proto.ConsumesSpec{ExtraIds: tc.ids}, &proto.RaidBuffs{})
		if gain[tc.stat] != tc.gain || gain[tc.other] != 0 {
			t.Errorf("%s: %s rose by %v and %s by %v, want %v and 0", tc.name, tc.stat.StatName(), gain[tc.stat],
				tc.other.StatName(), gain[tc.other], tc.gain)
		}
	}
}

// Juju Power 12451 (16323, 30 Strength) joins the Strength elixirs, Elixir of Greater Strength 9206
// (11405, 25 Strength) among them. An elixir the spec names comes before its extra ids.
func TestAJujuAndAnElixirOfItsGroupDoNotStack(t *testing.T) {
	const greaterStrength, jujuPower int32 = 9206, 12451
	registerStatConsumable(t, greaterStrength)
	registerStatConsumable(t, jujuPower)

	if gain := statGain(&proto.ConsumesSpec{ExtraIds: []int32{jujuPower}}, &proto.RaidBuffs{}); gain[stats.Strength] != 30 {
		t.Errorf("Juju Power adds %v Strength, want 30", gain[stats.Strength])
	}
	both := &proto.ConsumesSpec{ElixirIds: []int32{greaterStrength}, ExtraIds: []int32{jujuPower}}
	if gain := statGain(both, &proto.RaidBuffs{}); gain[stats.Strength] != 25 {
		t.Errorf("Elixir of Greater Strength and Juju Power add %v Strength, want the elixir's 25", gain[stats.Strength])
	}
}

// A scroll joins the elixir group of its stat: Scroll of Agility IV 10309 (12174, 17 Agility) and
// Elixir of the Mongoose 13452 (17538, 25 Agility) are one Agility group, and the elixir the spec
// names comes before its extra ids.
func TestAScrollAndAnElixirOfItsStatDoNotStack(t *testing.T) {
	const scrollOfAgilityIV int32 = 10309
	registerStatConsumable(t, scrollOfAgilityIV)
	registerStatConsumable(t, elixirOfTheMongooseID)

	if gain := statGain(&proto.ConsumesSpec{ExtraIds: []int32{scrollOfAgilityIV}}, &proto.RaidBuffs{}); gain[stats.Agility] != 17 {
		t.Errorf("Scroll of Agility IV adds %v Agility, want 17", gain[stats.Agility])
	}
	both := &proto.ConsumesSpec{ElixirIds: []int32{elixirOfTheMongooseID}, ExtraIds: []int32{scrollOfAgilityIV}}
	if gain := statGain(both, &proto.RaidBuffs{}); gain[stats.Agility] != 25 {
		t.Errorf("Scroll of Agility IV and Elixir of the Mongoose add %v Agility, want the elixir's 25", gain[stats.Agility])
	}
}

// Scroll of Strength IV 10310 (12179, 17 Strength), Juju Power 12451 (16323, 30) and Elixir of Greater
// Strength 9206 (11405, 25) are one Strength group: only the first the spec names applies.
func TestOnlyOneOfAStrengthScrollJujuAndElixirApplies(t *testing.T) {
	const scrollOfStrengthIV, jujuPower, greaterStrength int32 = 10310, 12451, 9206
	for _, id := range []int32{scrollOfStrengthIV, jujuPower, greaterStrength} {
		registerStatConsumable(t, id)
	}

	for _, tc := range []struct {
		name     string
		consumes *proto.ConsumesSpec
		gain     float64
	}{
		{"the scroll then Juju Power", &proto.ConsumesSpec{ExtraIds: []int32{scrollOfStrengthIV, jujuPower}}, 17},
		{"Juju Power then the scroll", &proto.ConsumesSpec{ExtraIds: []int32{jujuPower, scrollOfStrengthIV}}, 30},
		{"the elixir, the scroll and Juju Power", &proto.ConsumesSpec{ElixirIds: []int32{greaterStrength}, ExtraIds: []int32{scrollOfStrengthIV, jujuPower}}, 25},
	} {
		if gain := statGain(tc.consumes, &proto.RaidBuffs{}); gain[stats.Strength] != tc.gain {
			t.Errorf("%s: %v Strength, want %v", tc.name, gain[stats.Strength], tc.gain)
		}
	}
}

// No elixir opens on A_MOD_STAT Stamina, so Scroll of Stamina IV 10307 (12178, 16 Stamina) keeps a group
// of its own and stacks with Elixir of the Mongoose 13452.
func TestAScrollNoElixirSharesStacksWithTheElixirs(t *testing.T) {
	const scrollOfStaminaIV int32 = 10307
	registerStatConsumable(t, scrollOfStaminaIV)
	registerStatConsumable(t, elixirOfTheMongooseID)

	scroll := statGain(&proto.ConsumesSpec{ExtraIds: []int32{scrollOfStaminaIV}}, &proto.RaidBuffs{})
	elixir := statGain(&proto.ConsumesSpec{ElixirIds: []int32{elixirOfTheMongooseID}}, &proto.RaidBuffs{})
	both := statGain(&proto.ConsumesSpec{ElixirIds: []int32{elixirOfTheMongooseID}, ExtraIds: []int32{scrollOfStaminaIV}}, &proto.RaidBuffs{})
	if scroll[stats.Stamina] != 16 || both != scroll.Add(elixir) {
		t.Errorf("Scroll of Stamina IV adds %v Stamina, and with Elixir of the Mongoose %v, want 16 and each one's in full: %v and %v",
			scroll[stats.Stamina], both, scroll, elixir)
	}
}

// A flask has no group: Flask of Supreme Power 13512 (17628, 150 spell damage) stacks with Arcane
// Elixir 9155 (11390, 20), whose aura is the same A_MOD_DAMAGE_DONE every school.
func TestAFlaskStacksWithAnElixir(t *testing.T) {
	const supremePower, arcane int32 = 13512, 9155
	registerStatConsumable(t, supremePower)
	registerStatConsumable(t, arcane)

	flask := statGain(&proto.ConsumesSpec{FlaskId: supremePower}, &proto.RaidBuffs{})
	elixir := statGain(&proto.ConsumesSpec{ElixirIds: []int32{arcane}}, &proto.RaidBuffs{})
	if both := statGain(&proto.ConsumesSpec{FlaskId: supremePower, ElixirIds: []int32{arcane}}, &proto.RaidBuffs{}); both != flask.Add(elixir) {
		t.Errorf("Flask of Supreme Power and Arcane Elixir add %v, want each one's in full: %v and %v", both, flask, elixir)
	}
}

// Flask of Distilled Wisdom 13511 (17627) states its 2000 as A_MOD_MAX_POWER on mana, which is the
// Mana stat, and nothing else.
func TestDistilledWisdomAddsTwoThousandMana(t *testing.T) {
	const distilledWisdom int32 = 13511
	registerStatConsumable(t, distilledWisdom)

	gain := statGain(&proto.ConsumesSpec{FlaskId: distilledWisdom}, &proto.RaidBuffs{})
	if want := (stats.Stats{stats.Mana: 2000}); gain != want {
		t.Errorf("Flask of Distilled Wisdom adds %v, want 2000 Mana alone", gain)
	}

	_, without := newConsumablesSim(&proto.ConsumesSpec{}, &proto.RaidBuffs{})
	_, with := newConsumablesSim(&proto.ConsumesSpec{FlaskId: distilledWisdom}, &proto.RaidBuffs{})
	if got := with.MaxMana() - without.MaxMana(); got != 2000 {
		t.Errorf("Flask of Distilled Wisdom raises maximum mana by %v, want 2000", got)
	}
}

// Fire-toasted Bun 23327 (29332) and Dark Desire 22237 (27723) each grant 2% hit, one Hit food group:
// only the first applies.
func TestOnlyOneHitFoodApplies(t *testing.T) {
	const bun, darkDesire int32 = 23327, 22237
	registerStatConsumable(t, bun)
	registerStatConsumable(t, darkDesire)

	one := statGain(&proto.ConsumesSpec{ExtraIds: []int32{bun}}, &proto.RaidBuffs{})
	if one == (stats.Stats{}) {
		t.Fatal("Fire-toasted Bun adds nothing")
	}
	if both := statGain(&proto.ConsumesSpec{ExtraIds: []int32{bun, darkDesire}}, &proto.RaidBuffs{}); both != one {
		t.Errorf("Fire-toasted Bun and Dark Desire add %v, want the Bun's alone: %v", both, one)
	}
}

// Elixir of Agility 8949 (11328, 15 Agility) and Elixir of the Mongoose 13452 (17538, 25 Agility and
// 2% crit) both open on A_MOD_STAT Agility (29/1), one group: only the one the spec names first
// applies, in full.
func TestOnlyTheFirstElixirOfAGroupApplies(t *testing.T) {
	const elixirOfAgilityID int32 = 8949
	registerStatConsumable(t, elixirOfAgilityID)
	registerStatConsumable(t, elixirOfTheMongooseID)

	gain := statGain(&proto.ConsumesSpec{ElixirIds: []int32{elixirOfAgilityID, elixirOfTheMongooseID}}, &proto.RaidBuffs{})
	if gain[stats.Agility] != 15 || gain[stats.PhysicalCritPercent] != 0 {
		t.Errorf("Agility then Mongoose add %v Agility and %v%% crit, want 15 and 0", gain[stats.Agility], gain[stats.PhysicalCritPercent])
	}
	gain = statGain(&proto.ConsumesSpec{ElixirIds: []int32{elixirOfTheMongooseID, elixirOfAgilityID}}, &proto.RaidBuffs{})
	if gain[stats.Agility] != 25 || gain[stats.PhysicalCritPercent] < 2-1e-9 {
		t.Errorf("Mongoose then Agility add %v Agility and %v%% crit, want 25 and 2", gain[stats.Agility], gain[stats.PhysicalCritPercent])
	}
}

// Elixirs of different groups stack in full. Elixir of the Sages 13447 (17535, 18 Intellect then 18
// Spirit) is an Intellect elixir (29/3) and Elixir of the Whale 250348 (1250981, 25 Spirit) a Spirit
// one (29/4): 43 Spirit. Arcane Elixir 9155 (11390, 20 spell damage, 13/126) and Elixir of Fire Power
// 6373 (7844, 10 fire damage, 13/4) are two groups as well.
func TestElixirsOfDifferentGroupsStack(t *testing.T) {
	const sagesID, whaleID, arcaneID, firePowerID int32 = 13447, 250348, 9155, 6373
	for _, id := range []int32{sagesID, whaleID, arcaneID, firePowerID} {
		registerStatConsumable(t, id)
	}

	gain := statGain(&proto.ConsumesSpec{ElixirIds: []int32{sagesID, whaleID}}, &proto.RaidBuffs{})
	if gain[stats.Intellect] != 18 || gain[stats.Spirit] != 43 {
		t.Errorf("Sages and Whale add %v Intellect and %v Spirit, want 18 and 43", gain[stats.Intellect], gain[stats.Spirit])
	}
	arcane := statGain(&proto.ConsumesSpec{ElixirIds: []int32{arcaneID}}, &proto.RaidBuffs{})
	fire := statGain(&proto.ConsumesSpec{ElixirIds: []int32{firePowerID}}, &proto.RaidBuffs{})
	if both := statGain(&proto.ConsumesSpec{ElixirIds: []int32{arcaneID, firePowerID}}, &proto.RaidBuffs{}); both != arcane.Add(fire) {
		t.Errorf("Arcane Elixir and Elixir of Fire Power add %v, want each one's in full: %v and %v", both, arcane, fire)
	}
}

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

// What the consumable adds used on its own, as an extra id.
func soloGain(id int32) stats.Stats {
	return statGain(&proto.ConsumesSpec{ExtraIds: []int32{id}}, &proto.RaidBuffs{})
}

// The spell the consumable's on-use casts.
func consumableRow(id int32) *spelldata.Spell {
	return spelldata.MustFind(core.GetConsumableByID(id).OnUse().BuffId)
}

// Nutritious Food 1248401 hands Well Fed 1248422 a value of its own, which is not the Strength Well
// Fed's own row states; a copy of the food hands over one more than that, so the two always differ.
// Well Fed persists through death, so it is up for the whole iteration.
func TestWellFedTakesTheValueItsTriggerHandsOver(t *testing.T) {
	registerStatConsumable(t, smokedDesertDumplingsID)
	food := consumableRow(smokedDesertDumplingsID)
	wellFed := food.FirstAura(dbcenums.A_PERIODIC_TRIGGER_SPELL_WITH_VALUE).Trigger()
	if wellFed == spelldata.Nil {
		t.Fatalf("%d hands no value to a spell it triggers", food.ID)
	}
	own := wellFed.FirstAura(dbcenums.A_MOD_STAT).BasePoints
	editRow(t, food.ID, func(s *spelldata.Spell) {
		for i := range s.Effects {
			if s.Effects[i].Aura == dbcenums.A_PERIODIC_TRIGGER_SPELL_WITH_VALUE {
				s.Effects[i].BasePoints = own + 1
			}
		}
	})
	handed := own + 1

	gain := statGain(&proto.ConsumesSpec{FoodId: smokedDesertDumplingsID}, &proto.RaidBuffs{})
	if gain[stats.Strength] != handed {
		t.Errorf("Smoked Desert Dumplings add %v Strength, want the %v the trigger hands over, not Well Fed's own %v",
			gain[stats.Strength], handed, own)
	}

	_, character := newConsumablesSim(&proto.ConsumesSpec{FoodId: smokedDesertDumplingsID}, &proto.RaidBuffs{})
	aura := character.GetAuraByID(core.ActionID{ItemID: smokedDesertDumplingsID})
	if aura == nil || !aura.IsActive() || aura.Duration != core.NeverExpires {
		t.Errorf("Well Fed is %+v, want up for the whole iteration", aura)
	}
}

// 17538 states Agility and critical strike chance for its duration; the flat stats the catalogue
// carries name only the Agility.
func TestElixirTakesEveryEffectAndTheDurationOfItsRow(t *testing.T) {
	registerStatConsumable(t, elixirOfTheMongooseID)
	row := consumableRow(elixirOfTheMongooseID)
	agility, crit := row.FirstAura(dbcenums.A_MOD_STAT).Mean(), row.FirstAura(dbcenums.A_MOD_CRIT_PCT).Mean()
	needAmount(t, "17538's Agility", agility)
	needAmount(t, "17538's crit", crit)

	gain := statGain(&proto.ConsumesSpec{ElixirIds: []int32{elixirOfTheMongooseID}}, &proto.RaidBuffs{})
	if gain[stats.Agility] != agility || gain[stats.PhysicalCritPercent] < crit-1e-9 {
		t.Errorf("Elixir of the Mongoose adds %v Agility and %v%% crit, want %v and at least %v",
			gain[stats.Agility], gain[stats.PhysicalCritPercent], agility, crit)
	}

	_, character := newConsumablesSim(&proto.ConsumesSpec{ElixirIds: []int32{elixirOfTheMongooseID}}, &proto.RaidBuffs{})
	aura := character.GetAuraByID(core.ActionID{ItemID: elixirOfTheMongooseID})
	if aura == nil || !aura.IsActive() || aura.Duration != row.Duration() {
		t.Errorf("Elixir of the Mongoose is %+v, want up from the pull for the row's %v", aura, row.Duration())
	}
}

// Elixir of the Grizzly 250351 (1250986) states Strength and the same crit as Elixir of the Mongoose:
// both stats apply, the crit once.
func TestOnlyOneElixirsCritApplies(t *testing.T) {
	const grizzly int32 = 250351
	registerStatConsumable(t, elixirOfTheMongooseID)
	registerStatConsumable(t, grizzly)

	mongoose := statGain(&proto.ConsumesSpec{ElixirIds: []int32{elixirOfTheMongooseID}}, &proto.RaidBuffs{})
	grizzlys := statGain(&proto.ConsumesSpec{ElixirIds: []int32{grizzly}}, &proto.RaidBuffs{})
	both := statGain(&proto.ConsumesSpec{ElixirIds: []int32{elixirOfTheMongooseID, grizzly}}, &proto.RaidBuffs{})
	if both[stats.Agility] != mongoose[stats.Agility] || both[stats.Strength] != grizzlys[stats.Strength] {
		t.Errorf("Mongoose and Grizzly add %v Agility and %v Strength, want each one's own: %v and %v",
			both[stats.Agility], both[stats.Strength], mongoose[stats.Agility], grizzlys[stats.Strength])
	}
	if both[stats.PhysicalCritPercent] != mongoose[stats.PhysicalCritPercent] || both[stats.SpellCritPercent] != mongoose[stats.SpellCritPercent] {
		t.Errorf("Mongoose and Grizzly add %v%% physical and %v%% spell crit, want Mongoose's alone: %v%% and %v%%",
			both[stats.PhysicalCritPercent], both[stats.SpellCritPercent], mongoose[stats.PhysicalCritPercent], mongoose[stats.SpellCritPercent])
	}
}

// Scroll of Intellect IV states Intellect, which Arcane Brilliance's larger amount overrides rather
// than adds to.
func TestScrollIsExclusiveWithTheRaidBuffOfItsStat(t *testing.T) {
	registerStatConsumable(t, scrollOfIntellectIVID)

	if gain := soloGain(scrollOfIntellectIVID); gain[stats.Intellect] <= 0 {
		t.Errorf("Scroll of Intellect IV adds %v Intellect, want some", gain[stats.Intellect])
	}

	brilliance := &proto.RaidBuffs{ArcaneBrilliance: 1}
	if gain := statGain(&proto.ConsumesSpec{ExtraIds: []int32{scrollOfIntellectIVID}}, brilliance); gain[stats.Intellect] != 0 {
		t.Errorf("Scroll of Intellect IV adds %v Intellect under Arcane Brilliance, want 0", gain[stats.Intellect])
	}
}

// Swiftfeather Token's buff runs out well inside the category cooldown the item shares, so it is used
// in combat rather than put up before the pull.
func TestShortBuffIsUsedOnItsCooldown(t *testing.T) {
	registerStatConsumable(t, swiftfeatherTokenID)

	_, character := newConsumablesSim(&proto.ConsumesSpec{FoodId: swiftfeatherTokenID}, &proto.RaidBuffs{})
	mcd := character.GetMajorCooldown(core.ActionID{ItemID: swiftfeatherTokenID})
	if mcd == nil {
		t.Fatal("Swiftfeather Token registers no major cooldown")
	}
	cooldown := mcd.Spell.SharedCD.Duration

	aura := character.GetAuraByID(core.ActionID{ItemID: swiftfeatherTokenID})
	if aura == nil || aura.IsActive() || aura.Duration <= 0 || aura.Duration >= cooldown {
		t.Errorf("Swiftfeather Token's buff is %+v, want down at the pull and lasting less than the %v cooldown it shares",
			aura, cooldown)
	}
}

// The Zanza group holds Spirit of Zanza 20079 (24382, Spirit and Stamina) and the Blasted Lands buffs,
// R.O.I.D.S. 8410 (10667, Strength) and Ground Scorpok Assay 8412 (10669, Agility) among them: only
// the first the spec names is up.
func TestOnlyTheFirstOfTheZanzaGroupApplies(t *testing.T) {
	const roids, scorpok, zanza int32 = 8410, 8412, 20079
	for _, id := range []int32{roids, scorpok, zanza} {
		registerStatConsumable(t, id)
	}

	for _, tc := range []struct {
		name string
		ids  []int32
	}{
		{"R.O.I.D.S. then Ground Scorpok Assay", []int32{roids, scorpok}},
		{"Ground Scorpok Assay then R.O.I.D.S.", []int32{scorpok, roids}},
		{"Spirit of Zanza then R.O.I.D.S.", []int32{zanza, roids}},
		{"R.O.I.D.S. then Spirit of Zanza", []int32{roids, zanza}},
	} {
		gain := statGain(&proto.ConsumesSpec{ExtraIds: tc.ids}, &proto.RaidBuffs{})
		if first := soloGain(tc.ids[0]); !gain.EqualsWithTolerance(first, 1e-9) {
			t.Errorf("%s: %v, want the first one's alone: %v", tc.name, gain, first)
		}
	}
}

// The Rum group holds Rumsey Rum Black Label 21151 (25804, Stamina) and Kreeg's Stout Beatdown 18284
// (22790, Spirit and less Intellect): only the first the spec names is up, and it stacks with the food,
// Runn Tum Tuber Surprise 18254 (Intellect).
func TestOnlyTheFirstOfTheRumGroupAppliesBesideTheFood(t *testing.T) {
	const blackLabel, kreegs, tuber int32 = 21151, 18284, 18254
	for _, id := range []int32{blackLabel, kreegs, tuber} {
		registerStatConsumable(t, id)
	}

	for _, tc := range []struct {
		name string
		ids  []int32
	}{
		{"Rumsey Rum Black Label then Kreeg's Stout Beatdown", []int32{blackLabel, kreegs}},
		{"Kreeg's Stout Beatdown then Rumsey Rum Black Label", []int32{kreegs, blackLabel}},
	} {
		gain := statGain(&proto.ConsumesSpec{ExtraIds: tc.ids}, &proto.RaidBuffs{})
		if first := soloGain(tc.ids[0]); !gain.EqualsWithTolerance(first, 1e-9) {
			t.Errorf("%s: %v, want the first one's alone: %v", tc.name, gain, first)
		}
	}

	food := statGain(&proto.ConsumesSpec{FoodId: tuber}, &proto.RaidBuffs{})
	both := statGain(&proto.ConsumesSpec{FoodId: tuber, ExtraIds: []int32{blackLabel}}, &proto.RaidBuffs{})
	if want := food.Add(soloGain(blackLabel)); !both.EqualsWithTolerance(want, 1e-9) {
		t.Errorf("Runn Tum Tuber Surprise and Rumsey Rum Black Label add %v, want both: %v", both, want)
	}
}

// Juju Power 12451 (16323) joins the Strength elixirs, Elixir of Greater Strength 9206 (11405) among
// them. An elixir the spec names comes before its extra ids.
func TestAJujuAndAnElixirOfItsGroupDoNotStack(t *testing.T) {
	const greaterStrength, jujuPower int32 = 9206, 12451
	registerStatConsumable(t, greaterStrength)
	registerStatConsumable(t, jujuPower)

	if gain := soloGain(jujuPower); gain[stats.Strength] <= 0 {
		t.Errorf("Juju Power adds %v Strength, want some", gain[stats.Strength])
	}
	elixir := statGain(&proto.ConsumesSpec{ElixirIds: []int32{greaterStrength}}, &proto.RaidBuffs{})
	both := &proto.ConsumesSpec{ElixirIds: []int32{greaterStrength}, ExtraIds: []int32{jujuPower}}
	if gain := statGain(both, &proto.RaidBuffs{}); gain[stats.Strength] != elixir[stats.Strength] {
		t.Errorf("Elixir of Greater Strength and Juju Power add %v Strength, want the elixir's %v",
			gain[stats.Strength], elixir[stats.Strength])
	}
}

// A scroll joins the elixir group of its stat: Scroll of Agility IV 10309 (12174) and Elixir of the
// Mongoose 13452 (17538) are one Agility group, and the elixir the spec names comes before its extra
// ids.
func TestAScrollAndAnElixirOfItsStatDoNotStack(t *testing.T) {
	const scrollOfAgilityIV int32 = 10309
	registerStatConsumable(t, scrollOfAgilityIV)
	registerStatConsumable(t, elixirOfTheMongooseID)

	if gain := soloGain(scrollOfAgilityIV); gain[stats.Agility] <= 0 {
		t.Errorf("Scroll of Agility IV adds %v Agility, want some", gain[stats.Agility])
	}
	elixir := statGain(&proto.ConsumesSpec{ElixirIds: []int32{elixirOfTheMongooseID}}, &proto.RaidBuffs{})
	both := &proto.ConsumesSpec{ElixirIds: []int32{elixirOfTheMongooseID}, ExtraIds: []int32{scrollOfAgilityIV}}
	if gain := statGain(both, &proto.RaidBuffs{}); gain[stats.Agility] != elixir[stats.Agility] {
		t.Errorf("Scroll of Agility IV and Elixir of the Mongoose add %v Agility, want the elixir's %v",
			gain[stats.Agility], elixir[stats.Agility])
	}
}

// Scroll of Strength IV 10310 (12179), Juju Power 12451 (16323) and Elixir of Greater Strength 9206
// (11405) are one Strength group: only the first the spec names applies.
func TestOnlyOneOfAStrengthScrollJujuAndElixirApplies(t *testing.T) {
	const scrollOfStrengthIV, jujuPower, greaterStrength int32 = 10310, 12451, 9206
	for _, id := range []int32{scrollOfStrengthIV, jujuPower, greaterStrength} {
		registerStatConsumable(t, id)
	}
	elixir := statGain(&proto.ConsumesSpec{ElixirIds: []int32{greaterStrength}}, &proto.RaidBuffs{})

	for _, tc := range []struct {
		name     string
		consumes *proto.ConsumesSpec
		first    stats.Stats
	}{
		{"the scroll then Juju Power", &proto.ConsumesSpec{ExtraIds: []int32{scrollOfStrengthIV, jujuPower}}, soloGain(scrollOfStrengthIV)},
		{"Juju Power then the scroll", &proto.ConsumesSpec{ExtraIds: []int32{jujuPower, scrollOfStrengthIV}}, soloGain(jujuPower)},
		{"the elixir, the scroll and Juju Power", &proto.ConsumesSpec{ElixirIds: []int32{greaterStrength}, ExtraIds: []int32{scrollOfStrengthIV, jujuPower}}, elixir},
	} {
		if gain := statGain(tc.consumes, &proto.RaidBuffs{}); gain[stats.Strength] != tc.first[stats.Strength] {
			t.Errorf("%s: %v Strength, want the first one's %v", tc.name, gain[stats.Strength], tc.first[stats.Strength])
		}
	}
}

// No elixir opens on A_MOD_STAT Stamina, so Scroll of Stamina IV 10307 (12178) keeps a group of its own
// and stacks with Elixir of the Mongoose 13452.
func TestAScrollNoElixirSharesStacksWithTheElixirs(t *testing.T) {
	const scrollOfStaminaIV int32 = 10307
	registerStatConsumable(t, scrollOfStaminaIV)
	registerStatConsumable(t, elixirOfTheMongooseID)

	scroll := soloGain(scrollOfStaminaIV)
	elixir := statGain(&proto.ConsumesSpec{ElixirIds: []int32{elixirOfTheMongooseID}}, &proto.RaidBuffs{})
	both := statGain(&proto.ConsumesSpec{ElixirIds: []int32{elixirOfTheMongooseID}, ExtraIds: []int32{scrollOfStaminaIV}}, &proto.RaidBuffs{})
	if scroll[stats.Stamina] <= 0 || both != scroll.Add(elixir) {
		t.Errorf("Scroll of Stamina IV adds %v Stamina, and with Elixir of the Mongoose %v, want some and each one's in full: %v and %v",
			scroll[stats.Stamina], both, scroll, elixir)
	}
}

// A flask has no group: Flask of Supreme Power 13512 (17628, spell damage) stacks with Arcane Elixir
// 9155 (11390), whose aura is the same A_MOD_DAMAGE_DONE every school.
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

// Flask of Distilled Wisdom 13511 (17627) states its amount as A_MOD_MAX_POWER on mana, which is the
// Mana stat, and nothing else.
func TestDistilledWisdomAddsMaximumManaAlone(t *testing.T) {
	const distilledWisdom int32 = 13511
	registerStatConsumable(t, distilledWisdom)
	mana := consumableRow(distilledWisdom).FirstAura(dbcenums.A_MOD_MAX_POWER).Mean()
	needAmount(t, "17627's mana", mana)

	gain := statGain(&proto.ConsumesSpec{FlaskId: distilledWisdom}, &proto.RaidBuffs{})
	if want := (stats.Stats{stats.Mana: mana}); gain != want {
		t.Errorf("Flask of Distilled Wisdom adds %v, want %v Mana alone", gain, mana)
	}

	_, without := newConsumablesSim(&proto.ConsumesSpec{}, &proto.RaidBuffs{})
	_, with := newConsumablesSim(&proto.ConsumesSpec{FlaskId: distilledWisdom}, &proto.RaidBuffs{})
	if got := with.MaxMana() - without.MaxMana(); got != mana {
		t.Errorf("Flask of Distilled Wisdom raises maximum mana by %v, want %v", got, mana)
	}
}

// Fire-toasted Bun 23327 (29332) and Dark Desire 22237 (27723) each grant hit, one Hit food group:
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

// Elixir of Agility 8949 (11328, Agility) and Elixir of the Mongoose 13452 (17538, Agility and crit)
// both open on A_MOD_STAT Agility (29/1), one group: only the one the spec names first applies, in
// full.
func TestOnlyTheFirstElixirOfAGroupApplies(t *testing.T) {
	const elixirOfAgilityID int32 = 8949
	registerStatConsumable(t, elixirOfAgilityID)
	registerStatConsumable(t, elixirOfTheMongooseID)
	alone := func(id int32) stats.Stats {
		return statGain(&proto.ConsumesSpec{ElixirIds: []int32{id}}, &proto.RaidBuffs{})
	}

	for _, tc := range []struct {
		name string
		ids  []int32
	}{
		{"Agility then Mongoose", []int32{elixirOfAgilityID, elixirOfTheMongooseID}},
		{"Mongoose then Agility", []int32{elixirOfTheMongooseID, elixirOfAgilityID}},
	} {
		gain := statGain(&proto.ConsumesSpec{ElixirIds: tc.ids}, &proto.RaidBuffs{})
		if first := alone(tc.ids[0]); !gain.EqualsWithTolerance(first, 1e-9) {
			t.Errorf("%s add %v, want the first one's in full and nothing of the second: %v", tc.name, gain, first)
		}
	}
}

// Elixirs of different groups stack in full. Elixir of the Sages 13447 (17535, Intellect then Spirit)
// is an Intellect elixir (29/3) and Elixir of the Whale 250348 (1250981, Spirit) a Spirit one (29/4).
// Arcane Elixir 9155 (11390, spell damage, 13/126) and Elixir of Fire Power 6373 (7844, fire damage,
// 13/4) are two groups as well.
func TestElixirsOfDifferentGroupsStack(t *testing.T) {
	const sagesID, whaleID, arcaneID, firePowerID int32 = 13447, 250348, 9155, 6373
	for _, id := range []int32{sagesID, whaleID, arcaneID, firePowerID} {
		registerStatConsumable(t, id)
	}

	sages := statGain(&proto.ConsumesSpec{ElixirIds: []int32{sagesID}}, &proto.RaidBuffs{})
	whale := statGain(&proto.ConsumesSpec{ElixirIds: []int32{whaleID}}, &proto.RaidBuffs{})
	if both := statGain(&proto.ConsumesSpec{ElixirIds: []int32{sagesID, whaleID}}, &proto.RaidBuffs{}); !both.EqualsWithTolerance(sages.Add(whale), 1e-9) {
		t.Errorf("Sages and Whale add %v, want each one's in full: %v and %v", both, sages, whale)
	}
	arcane := statGain(&proto.ConsumesSpec{ElixirIds: []int32{arcaneID}}, &proto.RaidBuffs{})
	fire := statGain(&proto.ConsumesSpec{ElixirIds: []int32{firePowerID}}, &proto.RaidBuffs{})
	if both := statGain(&proto.ConsumesSpec{ElixirIds: []int32{arcaneID, firePowerID}}, &proto.RaidBuffs{}); both != arcane.Add(fire) {
		t.Errorf("Arcane Elixir and Elixir of Fire Power add %v, want each one's in full: %v and %v", both, arcane, fire)
	}
}

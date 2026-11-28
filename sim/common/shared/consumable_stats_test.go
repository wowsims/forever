package shared

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/assets/database"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
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

// Nutritious Food 1248401 hands Well Fed 1248422 a value of its own. Well Fed persists through death,
// so it is up for the whole iteration.
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

// Elixir of the Grizzly 250351 states the same crit as Elixir of the Mongoose. The two share the crit
// group, so only the first the spec names applies.
func TestOnlyTheFirstCritElixirApplies(t *testing.T) {
	const grizzly int32 = 250351
	registerStatConsumable(t, elixirOfTheMongooseID)
	registerStatConsumable(t, grizzly)

	for _, ids := range [][]int32{{elixirOfTheMongooseID, grizzly}, {grizzly, elixirOfTheMongooseID}} {
		first := statGain(&proto.ConsumesSpec{ElixirIds: ids[:1]}, &proto.RaidBuffs{})
		if both := statGain(&proto.ConsumesSpec{ElixirIds: ids}, &proto.RaidBuffs{}); first[stats.PhysicalCritPercent] <= 0 || both != first {
			t.Errorf("elixirs %v add %v, want %d's alone: %v", ids, both, ids[0], first)
		}
	}
}

// Swiftfeather Token's buff runs out well inside the category cooldown the item shares, so it is used
// in combat.
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

// The Zanza group holds Spirit of Zanza 20079 and the Blasted Lands buffs: only the first the spec
// names is up.
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

// The Rum group holds Rumsey Rum Black Label 21151 and Kreeg's Stout Beatdown 18284: only the first the
// spec names is up, and it stacks with the food.
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

// Each stat the elixirs the spec names grant counts only the strongest one's amount, in whatever
// order they come, where the elixirs sit in different groups.
func TestEachStatCountsOnlyItsStrongestElixir(t *testing.T) {
	const ferocity, mongoose, greaterAgility, grizzly, jujuPower int32 = 250350, 13452, 9187, 250351, 12451
	const cunning, greaterIntellect, sages, whale, greaterDefense, phalanx int32 = 250328, 9179, 13447, 250348, 13445, 250329
	for _, tc := range []struct {
		name    string
		elixirs []int32
		extras  []int32
		checked []stats.Stat
	}{
		{"Ferocity and Mongoose", []int32{ferocity, mongoose}, nil, []stats.Stat{stats.Strength, stats.Agility}},
		{"Mongoose and Ferocity", []int32{mongoose, ferocity}, nil, []stats.Stat{stats.Strength, stats.Agility}},
		{"Ferocity, Greater Agility and Grizzly", []int32{ferocity, greaterAgility, grizzly}, nil, []stats.Stat{stats.Strength, stats.Agility}},
		{"Grizzly and Juju Power", []int32{grizzly}, []int32{jujuPower}, []stats.Stat{stats.Strength}},
		{"Cunning and Greater Intellect", []int32{cunning, greaterIntellect}, nil, []stats.Stat{stats.Agility, stats.Intellect}},
		{"the Sages and the Whale", []int32{sages, whale}, nil, []stats.Stat{stats.Intellect, stats.Spirit}},
		{"Greater Defense and Phalanx", []int32{greaterDefense, phalanx}, nil, []stats.Stat{stats.Armor}},
	} {
		ids := append(slices.Clone(tc.elixirs), tc.extras...)
		for _, id := range ids {
			registerStatConsumable(t, id)
		}

		gain := statGain(&proto.ConsumesSpec{ElixirIds: tc.elixirs, ExtraIds: tc.extras}, &proto.RaidBuffs{})
		shared := false
		for _, stat := range tc.checked {
			strongest, granting := 0.0, 0
			for _, id := range ids {
				if solo := soloGain(id)[stat]; solo > 0 {
					strongest, granting = max(strongest, solo), granting+1
				}
			}
			shared = shared || granting > 1
			if gain[stat] != strongest {
				t.Errorf("%s add %v %v, want the strongest one's %v", tc.name, gain[stat], stat.StatName(), strongest)
			}
		}
		if !shared {
			t.Fatalf("%s: no two of them grant one of %v", tc.name, tc.checked)
		}
	}

	both := statGain(&proto.ConsumesSpec{ElixirIds: []int32{ferocity, mongoose}}, &proto.RaidBuffs{})
	if crit := soloGain(mongoose)[stats.PhysicalCritPercent]; crit <= 0 || both[stats.PhysicalCritPercent] != crit {
		t.Errorf("Ferocity and Mongoose add %v%% crit, want Mongoose's %v%%", both[stats.PhysicalCritPercent], crit)
	}
}

// Elixir of the Grizzly 250351 (Strength and crit) and Elixir of Greater Strength 9206 state the same
// Strength, which counts once beside Grizzly's crit.
func TestTwoElixirsOfEqualStrengthCountItOnce(t *testing.T) {
	const grizzly, greaterStrength int32 = 250351, 9206
	registerStatConsumable(t, grizzly)
	registerStatConsumable(t, greaterStrength)

	grizzlys := statGain(&proto.ConsumesSpec{ElixirIds: []int32{grizzly}}, &proto.RaidBuffs{})
	greaters := statGain(&proto.ConsumesSpec{ElixirIds: []int32{greaterStrength}}, &proto.RaidBuffs{})
	if grizzlys[stats.Strength] <= 0 || grizzlys[stats.Strength] != greaters[stats.Strength] {
		t.Fatalf("Grizzly adds %v Strength and Greater Strength %v, want the same amount", grizzlys[stats.Strength], greaters[stats.Strength])
	}

	for _, ids := range [][]int32{{grizzly, greaterStrength}, {greaterStrength, grizzly}} {
		if both := statGain(&proto.ConsumesSpec{ElixirIds: ids}, &proto.RaidBuffs{}); !both.EqualsWithTolerance(grizzlys, 1e-9) {
			t.Errorf("elixirs %v add %v, want Grizzly's alone: %v", ids, both, grizzlys)
		}
	}
}

// A scroll counts with the elixirs of its stat, the stronger alone, whichever comes first.
func TestAScrollCountsWithTheElixirsOfItsStat(t *testing.T) {
	const scrollOfStrengthIV, jujuPower, cunning, scrollOfAgilityIV, lesserAgility, scrollOfProtectionIV, defense int32 = 10310, 12451,
		250328, 10309, 3390, 10305, 8951
	for _, tc := range []struct {
		name           string
		stat           stats.Stat
		scroll, elixir int32
	}{
		{"Scroll of Strength IV and Juju Power", stats.Strength, scrollOfStrengthIV, jujuPower},
		{"Scroll of Intellect IV and Elixir of Cunning", stats.Intellect, scrollOfIntellectIVID, cunning},
		{"Scroll of Agility IV and Elixir of Lesser Agility", stats.Agility, scrollOfAgilityIV, lesserAgility},
		{"Scroll of Protection IV and Elixir of Defense", stats.Armor, scrollOfProtectionIV, defense},
	} {
		registerStatConsumable(t, tc.scroll)
		registerStatConsumable(t, tc.elixir)
		scroll, elixir := soloGain(tc.scroll)[tc.stat], soloGain(tc.elixir)[tc.stat]
		if scroll <= 0 || elixir <= 0 || scroll == elixir {
			t.Fatalf("%s: the scroll adds %v %v and the elixir %v, want both some and one more", tc.name, scroll, tc.stat.StatName(), elixir)
		}
		for _, ids := range [][]int32{{tc.scroll, tc.elixir}, {tc.elixir, tc.scroll}} {
			if gain := statGain(&proto.ConsumesSpec{ExtraIds: ids}, &proto.RaidBuffs{}); gain[tc.stat] != max(scroll, elixir) {
				t.Errorf("%s as %v add %v %v, want the stronger one's %v", tc.name, ids, gain[tc.stat], tc.stat.StatName(), max(scroll, elixir))
			}
		}
	}
}

// Arcane Intellect overwrites the Intellect of the elixirs and of Scroll of Intellect IV 10308, Prayer
// of Spirit the Spirit of Elixir of Sages 250338, and Power Word: Fortitude the Stamina of Elixir of
// Brute Force 13453. Elixir of Cunning 250328 keeps its Agility, Sages its crit and Brute Force its
// Strength.
func TestTheRaidBuffsOverwriteTheIntellectSpiritAndStaminaOfTheElixirsAndScrolls(t *testing.T) {
	const cunning, greaterIntellect, sages, scrollOfSpiritIV, bruteForce, scrollOfStaminaIV int32 = 250328, 9179, 250338, 10306, 13453, 10307
	for _, tc := range []struct {
		name    string
		buffs   *proto.RaidBuffs
		stat    stats.Stat
		elixirs []int32
		scroll  int32
		kept    stats.Stat
	}{
		{"Arcane Intellect", &proto.RaidBuffs{ArcaneIntellect: true}, stats.Intellect, []int32{cunning, greaterIntellect},
			scrollOfIntellectIVID, stats.Agility},
		{"Prayer of Spirit", &proto.RaidBuffs{PrayerOfSpirit: true}, stats.Spirit, []int32{sages}, scrollOfSpiritIV, stats.SpellCritPercent},
		{"Power Word: Fortitude", &proto.RaidBuffs{PowerWordFortitude: true}, stats.Stamina, []int32{bruteForce}, scrollOfStaminaIV, stats.Strength},
	} {
		for _, id := range append(slices.Clone(tc.elixirs), tc.scroll) {
			registerStatConsumable(t, id)
		}
		gain := statGain(&proto.ConsumesSpec{ElixirIds: tc.elixirs, ExtraIds: []int32{tc.scroll}}, tc.buffs)
		if gain[tc.stat] != 0 {
			t.Errorf("under %s, %v and %d add %v %v, want none", tc.name, tc.elixirs, tc.scroll, gain[tc.stat], tc.stat.StatName())
		}
		if kept := soloGain(tc.elixirs[0])[tc.kept]; kept <= 0 || gain[tc.kept] != kept {
			t.Errorf("under %s, %d adds %v %v, want its own %v", tc.name, tc.elixirs[0], gain[tc.kept], tc.kept.StatName(), kept)
		}
	}
}

// Elixir of the Mongoose's 13452 crit stacks with Moonkin Aura's and Leader of the Pack's.
func TestTheElixirCritStacksWithTheRaidCrit(t *testing.T) {
	registerStatConsumable(t, elixirOfTheMongooseID)
	mongoose := soloGain(elixirOfTheMongooseID)
	_, none := newConsumablesSim(&proto.ConsumesSpec{}, &proto.RaidBuffs{})
	for _, tc := range []struct {
		name string
		raid *proto.RaidBuffs
		crit float64
	}{
		{"Moonkin Aura", &proto.RaidBuffs{MoonkinAura: true}, buffs.MoonkinAuraValue(0)},
		{"Leader of the Pack", &proto.RaidBuffs{LeaderOfThePack: true}, buffs.LeaderOfThePackValue(0)},
	} {
		_, with := newConsumablesSim(&proto.ConsumesSpec{ElixirIds: []int32{elixirOfTheMongooseID}}, tc.raid)
		gain := with.GetStats().Subtract(none.GetStats())
		for _, stat := range []stats.Stat{stats.PhysicalCritPercent, stats.SpellCritPercent} {
			if want := tc.crit + mongoose[stat]; mongoose[stat] <= 0 || gain[stat] != want {
				t.Errorf("%s and Elixir of the Mongoose add %v%% %v, want both: %v%%", tc.name, gain[stat], stat.StatName(), want)
			}
		}
	}
}

// Mark of the Wild stacks with the elixirs and the scrolls.
func TestMarkOfTheWildStacksWithTheElixirs(t *testing.T) {
	for _, id := range []int32{10310, 12451} {
		registerStatConsumable(t, id)
		strength := soloGain(id)[stats.Strength]
		if gain := statGain(&proto.ConsumesSpec{ExtraIds: []int32{id}}, &proto.RaidBuffs{MarkOfTheWild: true}); strength <= 0 || gain[stats.Strength] != strength {
			t.Errorf("%d under Mark of the Wild adds %v Strength to the Mark's, want its own %v", id, gain[stats.Strength], strength)
		}
	}
}

// A raid buff that overwrites the elixirs and the scrolls counts by its own category, so an elixir or a
// scroll granting as much as it or more could count beside it.
func TestNoElixirOrScrollOutbidsTheRaidBuffOfItsStat(t *testing.T) {
	ceilings := map[string]float64{
		"stat Intellect": buffs.ArcaneIntellectValue(0),
		"stat Spirit":    buffs.PrayerOfSpiritValue(0),
		"stat Stamina":   buffs.PowerWordFortitudeValue(0),
	}
	checked := 0
	for _, consumable := range database.Load().Consumables {
		if consumable.Type != proto.ConsumableType_ConsumableTypeElixir && consumable.Type != proto.ConsumableType_ConsumableTypeScroll &&
			!consumable.CountsWithElixirs || len(consumable.ItemEffects) == 0 {
			continue
		}
		for _, buff := range consumableBuffs(spelldata.Find(consumable.ItemEffects[0].BuffId)) {
			for _, applied := range spelldata.DryRun(buff).Applied {
				if ceiling, ok := ceilings[applied.Kind]; ok {
					checked++
					if applied.Value >= ceiling {
						t.Errorf("%s (%d) grants %v %s, at least the raid buff's %v", consumable.Name, consumable.Id, applied.Value, applied.Kind, ceiling)
					}
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no elixir or scroll grants Intellect, Spirit or Stamina")
	}
}

// Scroll of Stamina IV 10307 and Elixir of the Mongoose 13452 grant different stats, so both count in
// full.
func TestAScrollAndAnElixirOfDifferentStatsStack(t *testing.T) {
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

// Elixirs of different stats stack in full: Arcane Elixir 9155 states spell damage of every school
// and Elixir of Fire Power 6373 fire damage alone, two stats.
func TestElixirsOfDifferentStatsStack(t *testing.T) {
	const arcaneID, firePowerID int32 = 9155, 6373
	registerStatConsumable(t, arcaneID)
	registerStatConsumable(t, firePowerID)

	arcane := statGain(&proto.ConsumesSpec{ElixirIds: []int32{arcaneID}}, &proto.RaidBuffs{})
	fire := statGain(&proto.ConsumesSpec{ElixirIds: []int32{firePowerID}}, &proto.RaidBuffs{})
	if both := statGain(&proto.ConsumesSpec{ElixirIds: []int32{arcaneID, firePowerID}}, &proto.RaidBuffs{}); both != arcane.Add(fire) {
		t.Errorf("Arcane Elixir and Elixir of Fire Power add %v, want each one's in full: %v and %v", both, arcane, fire)
	}
}

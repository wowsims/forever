package buffs

import (
	"fmt"
	"math"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// What applying an aura does to a fresh unit, in every respect a raid buff is compared on: how it is
// registered, what it bids and at what, and what its stats and pseudo-stats are worth at each stack.
type auraShape struct {
	Label      string
	ActionID   core.ActionID
	Tag        string
	Duration   time.Duration
	MaxStacks  int32
	BuildPhase core.CharacterBuildPhase
	Bids       []string
	Stacks     []unitDelta
	Shield     string
}

type unitDelta struct {
	Stats  map[string]float64
	Pseudo map[string]float64
}

// Every stat starts above zero, so that a multiplier has something to move.
const shapeBaseStat = 1000.0

func shapeCharacter() *core.Character {
	character := core.NewCharacter(&core.Party{}, 0, &proto.Player{
		Name:      "Buff Tester",
		Race:      proto.Race_RaceOrc,
		Class:     proto.Class_ClassWarrior,
		Spec:      &proto.Player_ProtectionWarrior{ProtectionWarrior: &proto.ProtectionWarrior{}},
		Equipment: &proto.EquipmentSpec{Items: []*proto.ItemSpec{}},
	})
	character.Env = &core.Environment{MeasuringStats: true}
	character.AddStats(shapeBaseStats())
	return &character
}

func shapeTarget() *core.Unit {
	target := core.NewTarget(&proto.Target{Level: 63}, 0)
	target.Env = &core.Environment{MeasuringStats: true}
	target.AddStats(shapeBaseStats())
	return &target.Unit
}

func shapeBaseStats() stats.Stats {
	var s stats.Stats
	for i := range s {
		s[i] = shapeBaseStat
	}
	return s
}

// Registers the aura on a fresh unit, applies it at every stack it can hold, and checks that
// expiring it hands every stat and pseudo-stat back.
func shapeOf(t *testing.T, onTarget bool, build func(*core.Unit) *core.Aura) auraShape {
	t.Helper()
	unit := &shapeCharacter().Unit
	if onTarget {
		unit = shapeTarget()
	}
	sim := &core.Simulation{}

	measure := func() (stats.Stats, stats.PseudoStats) {
		return unit.SortAndApplyStatDependencies(unit.GetStats()), unit.PseudoStats
	}
	beforeStats, beforePseudo := measure()

	aura := build(unit)
	shape := auraShape{
		Label:      aura.Label,
		ActionID:   aura.ActionID,
		Tag:        aura.Tag,
		Duration:   aura.Duration,
		MaxStacks:  aura.MaxStacks,
		BuildPhase: aura.BuildPhase,
	}

	aura.Activate(sim)
	for stacks := int32(1); ; stacks++ {
		if aura.MaxStacks > 0 {
			aura.SetStacks(sim, stacks)
		}
		afterStats, afterPseudo := measure()
		shape.Stacks = append(shape.Stacks, unitDelta{statsDiff(beforeStats, afterStats), pseudoDiff(beforePseudo, afterPseudo)})
		if stacks >= aura.MaxStacks {
			break
		}
	}

	for _, ee := range aura.ExclusiveEffects {
		shape.Bids = append(shape.Bids, fmt.Sprintf("%s single=%v priority=%v", ee.Category.Name, ee.Category.SingleAura, round6(ee.Priority)))
	}
	slices.Sort(shape.Bids)

	for _, spell := range unit.Spellbook {
		if spell.ActionID.SpellID == aura.ActionID.SpellID {
			shape.Shield = fmt.Sprintf("%v school %v", spell.ActionID, spell.SpellSchool)
		}
	}

	aura.Deactivate(sim)
	afterStats, afterPseudo := measure()
	if d := statsDiff(beforeStats, afterStats); d != nil {
		t.Errorf("%s leaves %v behind on expiry", aura.Label, d)
	}
	if d := pseudoDiff(beforePseudo, afterPseudo); d != nil {
		t.Errorf("%s leaves %v behind on expiry", aura.Label, d)
	}
	return shape
}

func round6(f float64) float64 {
	return math.Round(f*1e6) / 1e6
}

// A bid as shapeOf writes it.
func bid(category string, singleAura bool, priority float64) string {
	return fmt.Sprintf("%s single=%v priority=%v", category, singleAura, round6(priority))
}

// What adding these stats moves on the unit, the stats they derive included.
func statsAdded(unit *core.Unit, add map[stats.Stat]float64) map[string]float64 {
	var added stats.Stats
	for stat, v := range add {
		added[stat] = v
	}
	before := unit.SortAndApplyStatDependencies(unit.GetStats())
	return statsDiff(before, unit.SortAndApplyStatDependencies(unit.GetStats().Add(added)))
}

// The amount the row's effect states, read past the parse so that a parse that drops the effect or
// flips its sign shows. A row worth 0 would let either pass, so the test stops there.
func rowAmount(t *testing.T, e *spelldata.Effect) float64 {
	t.Helper()
	v := e.Average()
	if v == 0 {
		t.Fatalf("spell %d states 0 on aura %v, want an amount to check", e.SpellID, e.Aura)
	}
	return v
}

// The row's armor effect: a resistance on the physical school, or an armor aura no armor modifier scales.
func armorEffect(t *testing.T, s *spelldata.Spell) *spelldata.Effect {
	t.Helper()
	for i := range s.Effects {
		e := &s.Effects[i]
		if e.Aura == dbcenums.A_MOD_ARMOR_NO_MODS || e.Aura == dbcenums.A_MOD_RESISTANCE && e.Misc == 1 {
			return e
		}
	}
	t.Fatalf("spell %d states no armor effect", s.ID)
	return nil
}

func statsDiff(before, after stats.Stats) map[string]float64 {
	diff := map[string]float64{}
	for i := range before {
		if d := round6(after[i] - before[i]); d != 0 {
			diff[stats.Stat(i).StatName()] = d
		}
	}
	if len(diff) == 0 {
		return nil
	}
	return diff
}

// Every float field of PseudoStats, the per-school arrays included.
func pseudoDiff(before, after stats.PseudoStats) map[string]float64 {
	diff := map[string]float64{}
	bv, av := reflect.ValueOf(before), reflect.ValueOf(after)
	for i := 0; i < bv.NumField(); i++ {
		name := bv.Type().Field(i).Name
		bf, af := bv.Field(i), av.Field(i)
		switch bf.Kind() {
		case reflect.Float64:
			if d := round6(af.Float() - bf.Float()); d != 0 {
				diff[name] = d
			}
		case reflect.Array:
			for j := 0; j < bf.Len(); j++ {
				if bf.Index(j).Kind() != reflect.Float64 {
					continue
				}
				if d := round6(af.Index(j).Float() - bf.Index(j).Float()); d != 0 {
					diff[fmt.Sprintf("%s[%d]", name, j)] = d
				}
			}
		}
	}
	if len(diff) == 0 {
		return nil
	}
	return diff
}

// One row of each shape a generated constructor builds, as the external caster's copy unless the
// row says otherwise. The rows a test beside the drivers already pins are left to it.
func TestGeneratedRowsBuildTheirShapes(t *testing.T) {
	type ctor func(unit *core.Unit, isPlayer bool, talentPoints int32) *core.Aura
	statsOnly := func(s map[string]float64) []unitDelta { return []unitDelta{{Stats: s}} }
	pseudoOnly := func(p map[string]float64) []unitDelta { return []unitDelta{{Pseudo: p}} }
	external := func(spellID int32) core.ActionID { return core.ActionID{SpellID: spellID, Tag: -1} }
	onCharacter := func(add map[stats.Stat]float64) map[string]float64 { return statsAdded(&shapeCharacter().Unit, add) }
	sameOnEach := func(keys []string, v float64) map[string]float64 {
		m := map[string]float64{}
		for _, k := range keys {
			m[k] = round6(v)
		}
		return m
	}
	magicSchools := func(field string) []string {
		var keys []string
		for school := 2; school <= 7; school++ {
			keys = append(keys, fmt.Sprintf("%s[%d]", field, school))
		}
		return keys
	}

	row := func(s *spelldata.Spell, aura dbcenums.EffectAuraType) float64 {
		return rowAmount(t, firstAura(t, s, aura))
	}
	bloodPact := row(bloodPactSpell, dbcenums.A_MOD_STAT)
	darkspear := row(darkspearSpell, dbcenums.A_MOD_CRIT_PCT)
	admiralsHat := row(admiralsHatSpell, dbcenums.A_MOD_STAT)
	arcaneBrilliance := row(arcaneBrillianceSpell, dbcenums.A_MOD_STAT)
	fireResistanceAura := row(fireResistanceAuraSpell, dbcenums.A_MOD_RESISTANCE)
	concentration := -row(concentrationAuraSpell, dbcenums.A_REDUCE_PUSHBACK) / 100
	thorns := row(thornsSpell, dbcenums.A_DAMAGE_SHIELD)
	thunderClap := core.SymmetricMultiplierFromPercent(row(thunderClapSpell, dbcenums.A_MOD_MELEE_HASTE_3))
	insectSwarm := row(insectSwarmSpell, dbcenums.A_MOD_HIT_CHANCE)
	piHealing := 1 + row(powerInfusionsSpell, dbcenums.A_MOD_HEALING_DONE_PERCENT)/100
	piDamage := 1 + row(powerInfusionsSpell, dbcenums.A_MOD_DAMAGE_PERCENT_DONE)/100
	coeResistance := row(curseOfElementsSpell, dbcenums.A_MOD_RESISTANCE)
	coeTaken := row(curseOfElementsSpell, dbcenums.A_MOD_DAMAGE_PERCENT_TAKEN) / 100

	// Gift of the Wild states its armor, its five stats and each school's resistance as effects of
	// their own, the resistances keyed by the game files' school mask.
	gotwStat := rowAmount(t, giftOfTheWildSpell.Effect(dbcenums.A_MOD_STAT, -1))
	gotwAdded := map[stats.Stat]float64{stats.ArmorNoMods: rowAmount(t, armorEffect(t, giftOfTheWildSpell))}
	for _, stat := range []stats.Stat{stats.Strength, stats.Agility, stats.Stamina, stats.Intellect, stats.Spirit} {
		gotwAdded[stat] = gotwStat
	}
	var gotwBids []string
	for _, school := range []struct {
		stat stats.Stat
		mask int32
		bid  string
	}{
		{stats.FireResistance, 4, "ResistanceFireFireResistanceAdd"},
		{stats.NatureResistance, 8, "ResistanceNatureNatureResistanceAdd"},
		{stats.FrostResistance, 16, "ResistanceFrostFrostResistanceAdd"},
		{stats.ShadowResistance, 32, "ResistanceShadowShadowResistanceAdd"},
		{stats.ArcaneResistance, 64, "ResistanceArcaneArcaneResistanceAdd"},
	} {
		resistance := rowAmount(t, giftOfTheWildSpell.Effect(dbcenums.A_MOD_RESISTANCE, school.mask))
		gotwAdded[school.stat] = resistance
		gotwBids = append(gotwBids, bid(school.bid, false, resistance))
	}

	resistances := []string{"ArcaneResistance", "FireResistance", "FrostResistance", "NatureResistance", "ShadowResistance"}
	var coeBids []string
	for _, r := range resistances {
		coeBids = append(coeBids, bid("CurseOfElements"+r+"Add", false, math.Abs(coeResistance)))
	}
	coeBids = append(coeBids, bid("CurseOfElementsSchoolDamageTakenMultiplierMul", false, math.Abs(coeTaken)))

	piPseudo := sameOnEach(magicSchools("SchoolDamageDealtMultiplier"), piDamage-1)
	piPseudo["HealingDealtMultiplier"] = round6(piHealing - 1)

	cases := []struct {
		name     string
		onTarget bool
		isPlayer bool
		build    ctor
		want     auraShape
	}{
		{"BloodPact", false, false, NewSpellDataBloodPactAura, auraShape{
			Label: "Blood Pact (External)", ActionID: external(11767), Duration: BloodPactDuration(0),
			BuildPhase: core.CharacterBuildPhaseBuffs,
			Stacks:     statsOnly(onCharacter(map[stats.Stat]float64{stats.Stamina: bloodPact})),
		}},
		{"Darkspear", false, true, NewSpellDataDarkspearAura, auraShape{
			Label: "Darkspear (Player)", ActionID: core.ActionID{SpellID: 16939}, Tag: "Darkspear",
			Duration: DarkspearDuration(0),
			Bids:     []string{bid("Darkspear", true, math.Abs(darkspear))},
			Stacks: statsOnly(onCharacter(map[stats.Stat]float64{
				stats.PhysicalCritPercent: darkspear, stats.SpellCritPercent: darkspear,
			})),
		}},
		{"AdmiralsHat", false, false, NewSpellDataAdmiralsHatAura, auraShape{
			Label: "Admiral's Hat (External)", ActionID: external(12022), Tag: "AdmiralsHat",
			Duration: AdmiralsHatDuration(0), BuildPhase: core.CharacterBuildPhaseBuffs,
			Bids:   []string{bid("AdmiralsHat", true, math.Abs(admiralsHat))},
			Stacks: statsOnly(onCharacter(map[stats.Stat]float64{stats.Stamina: admiralsHat})),
		}},
		{"ArcaneBrilliance", false, false, NewSpellDataArcaneBrillianceAura, auraShape{
			Label: "Arcane Brilliance (External)", ActionID: external(23028), Tag: "StatBuff", Duration: ArcaneBrillianceDuration(0),
			BuildPhase: core.CharacterBuildPhaseBuffs,
			Bids:       []string{bid("StatBuffIntellectAdd", false, arcaneBrilliance)},
			Stacks:     statsOnly(onCharacter(map[stats.Stat]float64{stats.Intellect: arcaneBrilliance})),
		}},
		{"GiftOfTheWild", false, false, NewSpellDataGiftOfTheWildAura, auraShape{
			Label: "Gift of the Wild (External)", ActionID: external(21850), Duration: GiftOfTheWildDuration(0),
			BuildPhase: core.CharacterBuildPhaseBuffs,
			Bids:       gotwBids,
			Stacks:     statsOnly(onCharacter(gotwAdded)),
		}},
		{"FireResistanceAura", false, true, NewSpellDataFireResistanceAura, auraShape{
			Label: "Fire Resistance Aura (Player)", ActionID: core.ActionID{SpellID: 19900}, Tag: "FireResistanceAura",
			Duration: FireResistanceAuraDuration(0),
			Bids: []string{
				bid("FireResistanceAura", true, math.Abs(fireResistanceAura)),
				bid("PaladinAura", true, 0),
				bid("ResistanceFireFireResistanceAdd", false, fireResistanceAura),
			},
			Stacks: statsOnly(onCharacter(map[stats.Stat]float64{stats.FireResistance: fireResistanceAura})),
		}},
		{"ConcentrationAura", false, false, NewSpellDataConcentrationAura, auraShape{
			Label: "Concentration Aura (External)", ActionID: external(19746), Tag: "ConcentrationAura",
			Duration: ConcentrationAuraDuration(0), BuildPhase: core.CharacterBuildPhaseBuffs,
			Bids:   []string{bid("ConcentrationAura", true, math.Abs(concentration))},
			Stacks: pseudoOnly(map[string]float64{"PushbackChance": round6(concentration)}),
		}},
		{"PowerInfusions", false, false, NewSpellDataPowerInfusionsAura, auraShape{
			Label: "Power Infusions (External)", ActionID: external(10060), Tag: "PowerInfusion",
			Duration: PowerInfusionsDuration(0), BuildPhase: core.CharacterBuildPhaseBuffs,
			Bids: []string{
				bid("PowerInfusionHealingDealtMultiplierMul", false, math.Abs(piHealing-1)),
				bid("PowerInfusionSchoolDamageDealtMultiplierMul", false, math.Abs(piDamage-1)),
			},
			Stacks: pseudoOnly(piPseudo),
		}},
		{"Thorns", false, false, NewSpellDataThornsAura, auraShape{
			Label: "Thorns (External)", ActionID: external(9910), Duration: ThornsDuration(0),
			BuildPhase: core.CharacterBuildPhaseBuffs,
			Bids:       []string{bid("Thorns", true, thorns)},
			Stacks:     []unitDelta{{}},
			Shield:     "{SpellID: 9910, Tag: 1} school SpellSchoolNature",
		}},
		{"CurseOfElements", true, false, NewSpellDataCurseOfElementsAura, auraShape{
			Label: "Curse of the Elements (External)", ActionID: external(1311680), Tag: "CurseOfElements",
			Duration: CurseOfElementsDuration(0),
			Bids:     coeBids,
			Stacks: []unitDelta{{
				Stats:  sameOnEach(resistances, coeResistance),
				Pseudo: sameOnEach(magicSchools("SchoolDamageTakenMultiplier"), coeTaken),
			}},
		}},
		{"ThunderClap", true, false, NewSpellDataThunderClapAura, auraShape{
			Label: "Thunder Clap (External)", ActionID: external(11581), Tag: "AtkSpdReduction",
			Duration: ThunderClapDuration(0),
			Bids:     []string{bid(AtkSpeedReductionCategory, false, math.Abs(thunderClap-1))},
			Stacks:   pseudoOnly(map[string]float64{"MeleeSpeedMultiplier": round6(thunderClap - 1)}),
		}},
		{"InsectSwarm", true, false, NewSpellDataInsectSwarmAura, auraShape{
			Label: "Insect Swarm (External)", ActionID: external(24977), Tag: "HitChanceReduction",
			Duration: InsectSwarmDuration(0),
			Bids:     []string{bid("HitChanceReductionPhysicalHitPercentAdd", false, math.Abs(insectSwarm))},
			Stacks:   statsOnly(map[string]float64{"PhysicalHitPercent": round6(insectSwarm)}),
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := shapeOf(t, c.onTarget, func(u *core.Unit) *core.Aura { return c.build(u, c.isPlayer, 0) })
			slices.Sort(c.want.Bids)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("\n got %+v\nwant %+v", got, c.want)
			}
		})
	}
}

// A single-aura buff bids its whole value under its category, which is what AddGeneratedFlatBonus
// raises: the improved copy grants the bonus while it holds the category.
func TestMetaBuffTakesAFlatBonus(t *testing.T) {
	meta := &Meta{Label: "Battle Shout", Spell: spelldata.MustFind(25289), Category: "BattleShout", SingleAura: true}
	character := shapeCharacter()
	sim := &core.Simulation{}
	before := character.GetStat(stats.AttackPower)

	aura := newBuff(&character.Unit, meta, false, 0)
	AddGeneratedFlatBonus(aura, stats.AttackPower, meta.Value(0), BattleShoutT2Bonus)
	aura.Activate(sim)

	if got, want := character.GetStat(stats.AttackPower)-before, meta.Value(0)+BattleShoutT2Bonus; got != want {
		t.Errorf("improved Battle Shout grants %v attack power, want %v", got, want)
	}
}

func TestInspirationReadsTheStore(t *testing.T) {
	unit := &shapeCharacter().Unit
	applied := unit.SortAndApplyStatDependencies(unit.GetStats())
	baseArmor := applied[stats.Armor] - applied[stats.ArmorNoMods]

	if inspirationTalent.Len() == 0 {
		t.Fatal("Inspiration has no ranks")
	}
	previous := 0.0
	for points := int32(1); points <= inspirationTalent.Len(); points++ {
		shape := shapeOf(t, false, func(u *core.Unit) *core.Aura { return inspirationAura(u, points) })

		if shape.ActionID.SpellID != 14893 || shape.Duration != inspirationBuff.Duration() {
			t.Errorf("Inspiration at %d points is %v for %v, want 14893 for %v", points, shape.ActionID, shape.Duration, inspirationBuff.Duration())
		}
		share := inspirationTalent.MultiplierAt(points) - 1
		if share <= previous {
			t.Errorf("Inspiration at %d points shares %v of the armor, want more than the %v a rank below", points, share, previous)
		}
		previous = share
		if got, want := shape.Stacks[0], (unitDelta{Stats: map[string]float64{"Armor": round6(baseArmor * share)}}); !reflect.DeepEqual(got, want) {
			t.Errorf("Inspiration at %d points on %v armor moves %+v, want only %+v", points, baseArmor, got, want)
		}
	}
}

// A flat bonus the game files' buff data does not state raises two numbers that a
// non-stacking buff keeps apart: what the aura applies and what it bids for its
// category. A stacking buff prices the one off the other, so it refuses one.
func TestAddGeneratedFlatBonusRaisesTheBidAndTheAmount(t *testing.T) {
	meta := &Meta{Label: "Battle Shout", Spell: spelldata.MustFind(25289), Category: "BattleShout", SingleAura: true}
	character := shapeCharacter()
	sim := &core.Simulation{}
	before := character.GetStat(stats.AttackPower)

	aura := newBuff(&character.Unit, meta, false, 0)

	AddGeneratedFlatBonus(aura, stats.AttackPower, meta.Value(0), 30)
	// A second party member wearing the same set asks for the same total.
	AddGeneratedFlatBonus(aura, stats.AttackPower, meta.Value(0), 30)

	if priority, want := aura.ExclusiveEffects[0].Priority, meta.Value(0)+30; priority != want {
		t.Errorf("the category effect bids %v, want the game files' %v plus the set's 30", priority, meta.Value(0))
	}

	aura.Activate(sim)

	if got, want := character.GetStat(stats.AttackPower)-before, meta.Value(0)+30; got != want {
		t.Errorf("the aura applied %v attack power, want the game files' %v plus the set's 30", got, meta.Value(0))
	}
}

func TestAddGeneratedFlatBonusRefusesAnAuraThatCannotCarryOne(t *testing.T) {
	char := shapeCharacter()

	stacking := char.GetOrRegisterAura(core.Aura{
		Label:     "Generated Sunder Armor",
		Tag:       "GeneratedArmorReduction",
		ActionID:  core.ActionID{SpellID: 25225},
		Duration:  core.NeverExpires,
		MaxStacks: 5,
	})
	stacking.NewExclusiveEffect("GeneratedArmorReduction", true, core.ExclusiveEffect{})
	assertPanics(t, "a stacking aura", func() {
		AddGeneratedFlatBonus(stacking, stats.Armor, -2600, -30)
	})

	uncontested := char.GetOrRegisterAura(core.Aura{
		Label:    "Generated Blood Pact",
		ActionID: core.ActionID{SpellID: 11767}.WithTag(-1),
		Duration: core.NeverExpires,
	})
	assertPanics(t, "an aura with no category", func() {
		AddGeneratedFlatBonus(uncontested, stats.Stamina, 54, 30)
	})

	contestedMeta := &Meta{Label: "Battle Shout", Spell: spelldata.MustFind(25289), Category: "BattleShout", SingleAura: true}
	contested := newBuff(&char.Unit, contestedMeta, false, 0)
	assertPanics(t, "a base that is not what the buff bids", func() {
		AddGeneratedFlatBonus(contested, stats.AttackPower, contestedMeta.Value(0)+1, 30)
	})

	// A category that holds several auras at once never deactivates the one it
	// outbids, so an amount attached to the aura would apply while the effect
	// holding it did not. No generated row has that shape; the guard is for a
	// hand-written one that does.
	shared := char.GetOrRegisterAura(core.Aura{
		Label:    "Hand-Written Shout",
		Tag:      "GeneratedSharedCategory",
		ActionID: core.ActionID{SpellID: 25289}.WithTag(-2),
		Duration: core.NeverExpires,
	})
	shared.NewExclusiveEffect("GeneratedSharedCategory", false, core.ExclusiveEffect{Priority: 139})
	assertPanics(t, "a category that holds more than one aura", func() {
		AddGeneratedFlatBonus(shared, stats.AttackPower, 139, 30)
	})
}

func assertPanics(t *testing.T, what string, call func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("%s took a flat bonus, which it cannot carry", what)
		}
	}()
	call()
}

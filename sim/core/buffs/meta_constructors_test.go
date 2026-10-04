package buffs

import (
	"fmt"
	"math"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
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

	cases := []struct {
		name     string
		onTarget bool
		isPlayer bool
		build    ctor
		want     auraShape
	}{
		{"BloodPact", false, false, NewSpellDataBloodPactAura, auraShape{
			Label: "Blood Pact (External)", ActionID: external(11767), Duration: core.NeverExpires,
			BuildPhase: core.CharacterBuildPhaseBuffs,
			Stacks:     statsOnly(map[string]float64{"Stamina": 54, "Health": 540}),
		}},
		{"ArcaneBrilliance", false, false, NewSpellDataArcaneBrillianceAura, auraShape{
			Label: "Arcane Brilliance (External)", ActionID: external(23028), Tag: "StatBuff", Duration: time.Hour,
			BuildPhase: core.CharacterBuildPhaseBuffs,
			Bids:       []string{"StatBuffIntellectAdd single=false priority=31"},
			Stacks:     statsOnly(map[string]float64{"Intellect": 31}),
		}},
		{"GiftOfTheWild", false, false, NewSpellDataGiftOfTheWildAura, auraShape{
			Label: "Gift of the Wild (External)", ActionID: external(21850), Duration: time.Hour,
			BuildPhase: core.CharacterBuildPhaseBuffs,
			Bids: []string{
				"ResistanceArcaneArcaneResistanceAdd single=false priority=27",
				"ResistanceFireFireResistanceAdd single=false priority=27",
				"ResistanceFrostFrostResistanceAdd single=false priority=27",
				"ResistanceNatureNatureResistanceAdd single=false priority=27",
				"ResistanceShadowShadowResistanceAdd single=false priority=27",
			},
			Stacks: statsOnly(map[string]float64{
				"Strength": 16, "Agility": 16, "Stamina": 16, "Intellect": 16, "Spirit": 16,
				"Armor": 417, "Health": 160, "DodgePercent": 0.8, "ArcaneResistance": 27, "FireResistance": 27,
				"FrostResistance": 27, "NatureResistance": 27, "ShadowResistance": 27,
			}),
		}},
		{"FireResistanceAura", false, true, NewSpellDataFireResistanceAura, auraShape{
			Label: "Fire Resistance Aura (Player)", ActionID: core.ActionID{SpellID: 19900}, Tag: "FireResistanceAura",
			Duration: core.NeverExpires,
			Bids: []string{
				"FireResistanceAura single=true priority=60",
				"PaladinAura single=true priority=0",
				"ResistanceFireFireResistanceAdd single=false priority=60",
			},
			Stacks: statsOnly(map[string]float64{"FireResistance": 60}),
		}},
		{"ConcentrationAura", false, false, NewSpellDataConcentrationAura, auraShape{
			Label: "Concentration Aura (External)", ActionID: external(19746), Tag: "ConcentrationAura",
			Duration: core.NeverExpires, BuildPhase: core.CharacterBuildPhaseBuffs,
			Bids:   []string{"ConcentrationAura single=true priority=0.35"},
			Stacks: pseudoOnly(map[string]float64{"PushbackChance": -0.35}),
		}},
		{"PowerInfusions", false, false, NewSpellDataPowerInfusionsAura, auraShape{
			Label: "Power Infusions (External)", ActionID: external(10060), Tag: "PowerInfusion",
			Duration: 15 * time.Second, BuildPhase: core.CharacterBuildPhaseBuffs,
			Bids: []string{
				"PowerInfusionHealingDealtMultiplierMul single=false priority=0.2",
				"PowerInfusionSchoolDamageDealtMultiplierMul single=false priority=0.2",
			},
			Stacks: pseudoOnly(map[string]float64{
				"HealingDealtMultiplier":         0.2,
				"SchoolDamageDealtMultiplier[2]": 0.2, "SchoolDamageDealtMultiplier[3]": 0.2,
				"SchoolDamageDealtMultiplier[4]": 0.2, "SchoolDamageDealtMultiplier[5]": 0.2,
				"SchoolDamageDealtMultiplier[6]": 0.2, "SchoolDamageDealtMultiplier[7]": 0.2,
			}),
		}},
		{"Thorns", false, false, NewSpellDataThornsAura, auraShape{
			Label: "Thorns (External)", ActionID: external(9910), Duration: 10 * time.Minute,
			BuildPhase: core.CharacterBuildPhaseBuffs,
			Bids:       []string{"Thorns single=true priority=22"},
			Stacks:     []unitDelta{{}},
			Shield:     "{SpellID: 9910, Tag: 1} school SpellSchoolNature",
		}},
		{"CurseOfElements", true, false, NewSpellDataCurseOfElementsAura, auraShape{
			Label: "Curse of the Elements (External)", ActionID: external(1311680), Tag: "CurseOfElements",
			Duration: 5 * time.Minute,
			Bids: []string{
				"CurseOfElementsArcaneResistanceAdd single=false priority=75",
				"CurseOfElementsFireResistanceAdd single=false priority=75",
				"CurseOfElementsFrostResistanceAdd single=false priority=75",
				"CurseOfElementsNatureResistanceAdd single=false priority=75",
				"CurseOfElementsSchoolDamageTakenMultiplierMul single=false priority=0.1",
				"CurseOfElementsShadowResistanceAdd single=false priority=75",
			},
			Stacks: []unitDelta{{
				Stats: map[string]float64{
					"ArcaneResistance": -75, "FireResistance": -75, "FrostResistance": -75,
					"NatureResistance": -75, "ShadowResistance": -75,
				},
				Pseudo: map[string]float64{
					"SchoolDamageTakenMultiplier[2]": 0.1, "SchoolDamageTakenMultiplier[3]": 0.1,
					"SchoolDamageTakenMultiplier[4]": 0.1, "SchoolDamageTakenMultiplier[5]": 0.1,
					"SchoolDamageTakenMultiplier[6]": 0.1, "SchoolDamageTakenMultiplier[7]": 0.1,
				},
			}},
		}},
		{"ThunderClap", true, false, NewSpellDataThunderClapAura, auraShape{
			Label: "Thunder Clap (External)", ActionID: external(11581), Tag: "AtkSpdReduction",
			Duration: 30 * time.Second,
			Bids:     []string{AtkSpeedReductionCategory + " single=false priority=0.166667"},
			Stacks:   pseudoOnly(map[string]float64{"MeleeSpeedMultiplier": -0.166667}),
		}},
		{"InsectSwarm", true, false, NewSpellDataInsectSwarmAura, auraShape{
			Label: "Insect Swarm (External)", ActionID: external(24977), Tag: "HitChanceReduction",
			Duration: 12 * time.Second,
			Bids:     []string{"HitChanceReductionPhysicalHitPercentAdd single=false priority=2"},
			Stacks:   statsOnly(map[string]float64{"PhysicalHitPercent": -2}),
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := shapeOf(t, c.onTarget, func(u *core.Unit) *core.Aura { return c.build(u, c.isPlayer, 0) })
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
	baseArmor := unit.SortAndApplyStatDependencies(unit.GetStats())[stats.Armor]

	for points, share := range map[int32]float64{1: 0.08, 2: 0.17, 3: 0.25} {
		shape := shapeOf(t, false, func(u *core.Unit) *core.Aura { return inspirationAura(u, points) })

		if shape.ActionID.SpellID != 14893 || shape.Duration != 15*time.Second {
			t.Errorf("Inspiration at %d points is %v for %v, want 14893 for 15s", points, shape.ActionID, shape.Duration)
		}
		if got, want := round6(shape.Stacks[0].Stats["Armor"]), round6(baseArmor*share); got != want {
			t.Errorf("Inspiration at %d points adds %v armor to %v, want %v", points, got, baseArmor, want)
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
		t.Errorf("the category effect bids %v, want the game files' 139 plus the set's 30", priority)
	}

	aura.Activate(sim)

	if got, want := character.GetStat(stats.AttackPower)-before, meta.Value(0)+30; got != want {
		t.Errorf("the aura applied %v attack power, want the game files' 139 plus the set's 30", got)
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

	contested := newBuff(&char.Unit, &Meta{Label: "Battle Shout", Spell: spelldata.MustFind(25289), Category: "BattleShout", SingleAura: true}, false, 0)
	assertPanics(t, "a base that is not what the buff bids", func() {
		AddGeneratedFlatBonus(contested, stats.AttackPower, 620, 30)
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

package tooltip

import (
	"encoding/json"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/tools/database/dbc"
)

// The game files texts below are copied from wowsims.db with the rows they read, so the cases run where the
// database is absent.

func TestCompileDamageRangeWithSpellPowerShare(t *testing.T) {
	// Shadow Bolt rank 10: 268 ± 14.7 at level 60, with 0.857 of shadow spell power on each end.
	provider := fixtureProvider(fixtureRow{
		spell: dbc.Spell{ID: 25307, SchoolMask: 32, BaseLevel: 60, MaxLevel: 65},
		effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_SCHOOL_DAMAGE, EffectBasePoints: 268, Variance: 0.10980392,
			EffectRealPointsPerLevel: 1.9, EffectBonusCoefficient: 0.857}},
	})
	text := "Sends a shadowy bolt at the enemy, causing $s1 Shadow damage."
	compileCase(t, provider, 25307, text, nil, player{}, "Sends a shadowy bolt at the enemy, causing 253 to 283 Shadow damage.")
	compileCase(t, provider, 25307, text, nil, player{values: map[string]float64{"SPS": 100}}, "Sends a shadowy bolt at the enemy, causing 339 to 369 Shadow damage.")
}

func TestCompileAttackPowerShare(t *testing.T) {
	// Infernal Strike: 100 Shadow damage plus the attack power.
	provider := fixtureProvider(fixtureRow{
		spell: dbc.Spell{ID: 412790, SchoolMask: 33, Duration: 2000},
		effects: []dbc.SpellEffect{
			{EffectIndex: 0, EffectType: 213},
			{EffectIndex: 1, EffectType: dbcenums.E_SCHOOL_DAMAGE, EffectBasePoints: 100, BonusCoefficientFromAP: 1},
		},
	})
	text := "Slam down on all enemies of the target area, causing $s2 Shadow damage and Stunning them for $d."
	compileCase(t, provider, 412790, text, nil, player{}, "Slam down on all enemies of the target area, causing 100 Shadow damage and Stunning them for 2 sec.")
	compileCase(t, provider, 412790, text, nil, player{values: map[string]float64{"AP": 500}},
		"Slam down on all enemies of the target area, causing 600 Shadow damage and Stunning them for 2 sec.")
}

func TestCompileTalentCondition(t *testing.T) {
	// Tactical Mastery adds its rank 1 amount to the rage kept, for a player with a point in it.
	provider := fixtureProvider(
		fixtureRow{spell: dbc.Spell{ID: 1310185}, effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_APPLY_AURA, EffectAura: dbcenums.A_DUMMY, EffectBasePoints: 10}}},
		fixtureRow{spell: dbc.Spell{ID: 12295}, effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_APPLY_AURA, EffectAura: dbcenums.A_DUMMY, EffectBasePoints: 15}}},
	)
	talents := map[int64]Talent{12295: {Spell: 12295, Rank: 1}}
	text := "You retain up to $?s12295[${$m1+$12295m1}][$m1] Rage when you change Stances."
	compileCase(t, provider, 1310185, text, talents, player{}, "You retain up to 10 Rage when you change Stances.")
	compileCase(t, provider, 1310185, text, talents, player{points: map[int64]int{12295: 1}}, "You retain up to 25 Rage when you change Stances.")
}

func TestCompileDuration(t *testing.T) {
	for _, c := range []struct {
		ms   int32
		want string
	}{
		{3000, "Stuns the target for 3 sec."},
		{1500, "Stuns the target for 1.5 sec."},
		{600000, "Stuns the target for 10 min."},
		{7200000, "Stuns the target for 2 hrs."},
	} {
		provider := fixtureProvider(fixtureRow{spell: dbc.Spell{ID: 853, Duration: c.ms}})
		compileCase(t, provider, 853, "Stuns the target for $d.", nil, player{}, c.want)
	}
}

// A spell with no duration and a permanent aura's -1 both read as the game writes them, until canceled.
func TestCompileDurationOfNone(t *testing.T) {
	for _, ms := range []int32{0, -1} {
		provider := fixtureProvider(fixtureRow{spell: dbc.Spell{ID: 101, Duration: ms}})
		compileCase(t, provider, 101, "Lasts $d.", nil, player{}, "Lasts until canceled.")
	}
}

// A roll's ends round a half up from the base points spread by the variance, plus the per-level gain up to the
// spell's MaxLevel: Pyroblast rank 6 is 191 ± 23.6 plus 6 levels of 2.1, and its dot 19 per tick for 4 ticks.
func TestCompileRollEndsAndLevelGain(t *testing.T) {
	provider := fixtureProvider(fixtureRow{
		spell: dbc.Spell{ID: 12522, SchoolMask: 4, BaseLevel: 30, MaxLevel: 36, Duration: 12000},
		effects: []dbc.SpellEffect{
			{EffectIndex: 0, EffectType: dbcenums.E_SCHOOL_DAMAGE, EffectBasePoints: 191, Variance: 0.24742268, EffectRealPointsPerLevel: 2.1},
			{EffectIndex: 1, EffectType: dbcenums.E_APPLY_AURA, EffectAura: dbcenums.A_PERIODIC_DAMAGE, EffectBasePoints: 19, EffectAuraPeriod: 3000},
		},
	})
	compileCase(t, provider, 12522, "Hurls an immense fiery boulder that causes $s1 Fire damage and an additional $o2 Fire damage over $d.", nil, player{},
		"Hurls an immense fiery boulder that causes 180 to 227 Fire damage and an additional 76 Fire damage over 12 sec.")
	// The equal ends of the dot write once.
	compileCase(t, provider, 12522, "$s2 Fire damage every $t2 seconds.", nil, player{}, "19 Fire damage every 3 seconds.")

	// Wild Growth's $o1 and $w1 carry the gain over 8 levels, and its radius keeps the half.
	growth := fixtureProvider(fixtureRow{
		spell: dbc.Spell{ID: 408120, BaseLevel: 40, MaxLevel: 48, Duration: 7000},
		effects: []dbc.SpellEffect{
			{EffectIndex: 0, EffectType: dbcenums.E_APPLY_AURA, EffectAura: dbcenums.A_PERIODIC_HEAL, EffectBasePoints: 40, EffectRealPointsPerLevel: 1, EffectAuraPeriod: 1000, EffectMaxRange: []float64{43.5, 0}},
		},
	})
	compileCase(t, growth, 408120, "Heals the target and their party for $o1 over $d. Party members must be within $a1 yards of target.", nil, player{},
		"Heals the target and their party for 336 over 7 sec. Party members must be within 43.5 yards of target.")
	compileCase(t, growth, 408120, "Heals $w1 damage every $t1 second.", nil, player{}, "Heals 48 damage every 1 second.")
}

func TestCompilePerComboPointList(t *testing.T) {
	// Expose Armor rank 5: $b is the armor the effect takes per combo point.
	provider := fixtureProvider(fixtureRow{
		spell:   dbc.Spell{ID: 11198, Duration: 30000},
		effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_APPLY_AURA, EffectAura: dbcenums.A_MOD_RESISTANCE, EffectPointsPerResource: -450}},
	})
	compileCase(t, provider, 11198,
		"Finishing move that exposes the target for $d, reducing Armor per combo point:\r\n   1 point  : ${$b*-1} Armor\r\n   2 points: ${$b*-2} Armor\r\n   5 points: ${$b*-5} Armor",
		nil, player{}, "Finishing move that exposes the target for 30 sec, reducing Armor per combo point:\n 1 point : 450 Armor\n 2 points: 900 Armor\n 5 points: 2250 Armor")
}

// Prayer of Mending's $<mult> comes from its description variables, whose conditions are talents; $bh is the
// healing power and $bc the coefficient its dummy effect carries.
func TestCompileNamedVariables(t *testing.T) {
	provider := fixtureProvider(
		fixtureRow{
			spell: dbc.Spell{ID: 1240826, Variables: "$twinD=$?s1225132[${$1225132m1/100+1)}][${1}]\r\n$spiritualH=$?s14898[${$14898m1/100+1)}][${1}]\r\n$mult=${$<twinD>*$<spiritualH>}"},
			effects: []dbc.SpellEffect{
				{EffectIndex: 0, EffectType: dbcenums.E_DUMMY, EffectBasePoints: 298, EffectBonusCoefficient: 0.429},
				{EffectIndex: 1, EffectType: dbcenums.E_DUMMY, EffectBasePoints: 5},
			},
		},
		fixtureRow{spell: dbc.Spell{ID: 1225132}, effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_APPLY_AURA, EffectAura: dbcenums.A_DUMMY, EffectBasePoints: 5}}},
		fixtureRow{spell: dbc.Spell{ID: 14898}, effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_APPLY_AURA, EffectAura: dbcenums.A_DUMMY, EffectBasePoints: 2}}},
	)
	talents := map[int64]Talent{1225132: {Spell: 1225132, Rank: 1}, 14898: {Spell: 14898, Rank: 1}}
	text := "Places a spell on the target that heals them for ${($m1+($bh*$bc))*$<mult>} the next time they take damage. Jumps up to $s2 times."
	compileCase(t, provider, 1240826, text, talents, player{}, "Places a spell on the target that heals them for 298 the next time they take damage. Jumps up to 5 times.")
	compileCase(t, provider, 1240826, text, talents, player{values: map[string]float64{"HEAL": 100}, points: map[int64]int{1225132: 1}},
		"Places a spell on the target that heals them for 358 the next time they take damage. Jumps up to 5 times.")
}

func TestCompileLevelConditions(t *testing.T) {
	// Greater Cryoblast reads its row at level 55 and up, and a formula on the level below it.
	cryoblast := fixtureProvider(fixtureRow{
		spell:   dbc.Spell{ID: 1302309, SchoolMask: 16, BaseLevel: 46, MaxLevel: 55},
		effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_SCHOOL_DAMAGE, EffectBasePoints: 400, Variance: 0.29767442, EffectRealPointsPerLevel: 12}},
	})
	text := "Inflicts $?$pl>=55[$s1][${340+($pl-46)*12} to ${459+($pl-46)*12}] Frost damage."
	compileCase(t, cryoblast, 1302309, text, nil, player{values: map[string]float64{"PL": 60}}, "Inflicts 448 to 568 Frost damage.")
	compileCase(t, cryoblast, 1302309, text, nil, player{values: map[string]float64{"PL": 50}}, "Inflicts 388 to 507 Frost damage.")

	// The Faction Banner's chain of $PL conditions, and the spaces an empty branch and its neighbours keep.
	banner := fixtureProvider(
		fixtureRow{spell: dbc.Spell{ID: 1307239}},
		fixtureRow{spell: dbc.Spell{ID: 1229718}, effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_APPLY_AURA, EffectAura: dbcenums.A_MOD_STAT, EffectBasePoints: 32}}},
	)
	chain := "gain $?$PL<40[14]?$PL<50[19]?$PL<60[27][$1229718s1] increased Spirit$?$PL<10[]?$PL<30[ and ][, ]rested."
	for level, want := range map[float64]string{30: "gain 14 increased Spirit, rested.", 45: "gain 19 increased Spirit, rested.", 60: "gain 32 increased Spirit, rested.", 20: "gain 14 increased Spirit and rested."} {
		compileCase(t, banner, 1307239, chain, nil, player{values: map[string]float64{"PL": level}}, want)
	}
}

func TestCompileIncludes(t *testing.T) {
	provider := fixtureProvider(
		fixtureRow{spell: dbc.Spell{ID: 1243064}},
		fixtureRow{spell: dbc.Spell{ID: 1243081, Description: "Inflicts $S1 Shadow damage to the current enemy target."},
			effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_SCHOOL_DAMAGE, EffectBasePoints: 30}}},
		fixtureRow{spell: dbc.Spell{ID: 1290054, Description: "Choking spores fill the air.", AuraDescription: "Inflicts $s1 Nature damage every $t1 sec."},
			effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_APPLY_AURA, EffectAura: dbcenums.A_PERIODIC_DAMAGE, EffectBasePoints: 10, EffectAuraPeriod: 1000}}},
		fixtureRow{spell: dbc.Spell{ID: 1225960, Description: "Teleport forward $M2 yds $?$broken."}},
	)
	compileCase(t, provider, 1243064, "$@spelldesc<1243081>\r\n", nil, player{}, "Inflicts 30 Shadow damage to the current enemy target.")
	compileCase(t, provider, 1243064, "$@spellaura<1290054>", nil, player{}, "Inflicts 10 Nature damage every 1 sec.")
	compileCase(t, provider, 1243064, "$@spellaura", nil, player{}, "")

	// An include that does not compile fails the text that includes it, which keeps the rest.
	segments, err := Compile("Augment your belt:\r\n\r\n$@spelldesc1225960", provider, 1243064, nil)
	if err == nil {
		t.Fatalf("an include that does not compile compiles to %v with no error", segments)
	}
	if got := renderJSON(t, segments, player{}); got != "Augment your belt:" {
		t.Errorf("the text around the failed include renders %q", got)
	}
}

// A PlayerCondition the extracted tables do not hold reads as not met.
func TestCompilePlayerCondition(t *testing.T) {
	provider := fixtureProvider(fixtureRow{spell: dbc.Spell{ID: 1312670}}, fixtureRow{spell: dbc.Spell{ID: 2484, NameLang: "Earthbind Totem"}})
	compileCase(t, provider, 1312670, "Returns you to $?pc923[Orgrimmar][Stormwind]. Single-use.", nil, player{}, "Returns you to Stormwind. Single-use.")
	compileCase(t, provider, 1312670, "You currently have access to:\r\n\r\n$?pc166245\r\n[$@spellicon2484 $@spellname2484][]\r\n", nil, player{}, "You currently have access to:")
}

func TestCompileEffectValues(t *testing.T) {
	provider := fixtureProvider(
		// Just a Flesh Wound reads its tenth to thirteenth effects.
		fixtureRow{spell: dbc.Spell{ID: 400014}, effects: []dbc.SpellEffect{
			{EffectIndex: 9, EffectType: dbcenums.E_APPLY_AURA, EffectBasePoints: 20},
			{EffectIndex: 11, EffectType: dbcenums.E_APPLY_AURA, EffectBasePoints: 12},
		}},
		// Resurrection's mana is its effect's misc value; Mind Soothe's level its target restriction.
		fixtureRow{spell: dbc.Spell{ID: 10880, MaxTargetLevel: 40}, effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_DUMMY, EffectBasePoints: 300, EffectMiscValues: []int{520, 0}}}},
		// Reckoning's aura stacks to 4 with one charge.
		fixtureRow{spell: dbc.Spell{ID: 1326084, ProcCharges: 1, MaxCumulativeStacks: 4}},
		// Shifting Power names a second effect it does not have.
		fixtureRow{spell: dbc.Spell{ID: 1322605}, effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_ENERGIZE, EffectBasePoints: 40}}},
	)
	compileCase(t, provider, 400014, "reduced by 1% for every $s12 of your Defense skill beyond ${$PL*5}, you deal $s10% less damage", nil, player{values: map[string]float64{"PL": 60}},
		"reduced by 1% for every 12 of your Defense skill beyond 300, you deal 20% less damage")
	compileCase(t, provider, 10880, "Brings a dead player back to life with $s1 health and $q1 mana. Affects targets level $v or lower.", nil, player{},
		"Brings a dead player back to life with 300 health and 520 mana. Affects targets level 40 or lower.")
	compileCase(t, provider, 10880, "Stores up to $1326084u extra attacks.", nil, player{}, "Stores up to 4 extra attacks.")
	compileCase(t, provider, 1322605, "Instantly convert $c Mana into $s2 Energy.", nil, player{}, "Instantly convert 0 Mana into 0 Energy.")
	compileCase(t, provider, 1322605, "Gain $s2 extra $lattack:attacks;.", nil, player{}, "Gain 0 extra attacks.")
}

func TestCompileComputedValues(t *testing.T) {
	provider := fixtureProvider(
		fixtureRow{spell: dbc.Spell{ID: 26111}, effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_APPLY_AURA, EffectAura: dbcenums.A_ADD_FLAT_MODIFIER, EffectBasePoints: -20}}},
		fixtureRow{spell: dbc.Spell{ID: 24491}, effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_APPLY_AURA, EffectAura: dbcenums.A_ADD_FLAT_MODIFIER, EffectBasePoints: -500}}},
		fixtureRow{spell: dbc.Spell{ID: 20467, SchoolMask: 2, BaseLevel: 20, MaxLevel: 28},
			effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_SCHOOL_DAMAGE, EffectBasePoints: 97, Variance: 0.08247422, EffectRealPointsPerLevel: 5.6}}},
		fixtureRow{spell: dbc.Spell{ID: 20660}, effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_DUMMY, EffectBasePoints: 325, EffectChainAmplitude: 0.9}}},
		fixtureRow{spell: dbc.Spell{ID: 405092}, effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_APPLY_AURA, EffectBasePoints: -75}}},
		fixtureRow{spell: dbc.Spell{ID: 25695}, effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_APPLY_AURA, EffectBasePoints: 358}}},
	)
	for _, c := range []struct {
		id         int64
		text, want string
	}{
		// The game writes the minus before an unsigned value.
		{26111, "-$/10;s1 rage cost to Intercept.", "-2 rage cost to Intercept."},
		// $10/;s1 for $/10;s1.
		{26111, "Restore $10/;s1 Rage.", "Restore 2 Rage."},
		// One decimal, none on a whole number, and both ends of a roll.
		{24491, "Reduces the casting time of your Flamestrike spell by $/1000;S1 sec.", "Reduces the casting time of your Flamestrike spell by 0.5 sec."},
		{20467, "instantly causing $/2;20467s1 Holy damage, $20467s1 if the target is stunned", "instantly causing 69 to 73 Holy damage, 138 to 146 if the target is stunned"},
		{20660, "converting each extra point of rage into $*10;F1 additional damage.", "converting each extra point of rage into 9 additional damage."},
		// $25695/5;s1 for $/5;25695s1.
		{25695, "Restores $25695/5;s1 health per second.", "Restores 71.6 health per second."},
		{405092, "Underwater breath is reduced by ${$abs($s1)}%.", "Underwater breath is reduced by 75%."},
		// A $ before the %.
		{405092, "increasing attack and movement speed by $s$%", "increasing attack and movement speed by 75%"},
	} {
		compileCase(t, provider, c.id, c.text, nil, player{}, c.want)
	}
}

func TestCompileTextForms(t *testing.T) {
	provider := fixtureProvider(
		fixtureRow{spell: dbc.Spell{ID: 364208}, effects: []dbc.SpellEffect{
			{EffectIndex: 0, EffectType: dbcenums.E_APPLY_AURA},
			{EffectIndex: 1, EffectType: dbcenums.E_APPLY_AURA, EffectBasePoints: 1},
			{EffectIndex: 2, EffectType: dbcenums.E_APPLY_AURA, EffectBasePoints: 0},
		}},
		fixtureRow{spell: dbc.Spell{ID: 17752}, effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_APPLY_AURA, EffectBasePoints: 3}}},
	)
	for _, c := range []struct {
		text, want string
	}{
		{"$?$W2==1[Immune to snare and root effects][Immune to snare effects].\r\n$?$W3==1[Movement speed increased by 50%.][]", "Immune to snare and root effects."},
		// A mount's speed is in tables the generator does not extract, so its branch is left out.
		{"$?j1g[Increases ground speed by $j1g%. ][]", ""},
		{"$@null\r\n$@null", ""},
		{"Controlled by $@auracaster.", "Controlled by the caster."},
		{"|CFFFFFFFFAxe/Sword:|R Your successful melee attacks", "Axe/Sword: Your successful melee attacks"},
		{"for now.\r\n\r\n<Combat between players is currently forbidden.>", "for now.\n\n<Combat between players is currently forbidden.>"},
		{"???", "???"},
		{"Deals $s2~ damage to surrounding allies on death.", "Deals 1~ damage to surrounding allies on death."},
		// The spaces the text has, and none where it has none.
		{"More effective than Suffering (Rank $s2).", "More effective than Suffering (Rank 1)."},
		{"$?a1[Heals for $s2.]\r\n\r\n[Fuses magic.]", "Heals for 1."},
	} {
		compileCase(t, provider, 364208, c.text, nil, player{}, c.want)
	}
}

// An empty branch is an empty list, which the UI renders; a null one it cannot.
func TestCompileWritesAnEmptyBranchAsAList(t *testing.T) {
	provider := fixtureProvider(fixtureRow{spell: dbc.Spell{ID: 1307234}})
	segments, err := Compile("Armor$?$PL<10[]?$PL<30[ and ][, ]all stats", provider, 1307234, nil)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(segments)
	if strings.Contains(string(data), "null") {
		t.Errorf("compiles to %s, with a null branch", data)
	}
}

// What a tooltip's expressions read off the player: values by the names the UI gives them, and talent points.
type player struct {
	values map[string]float64
	points map[int64]int
}

func compileCase(t *testing.T, provider TooltipDataProvider, spellId int64, text string, talents map[int64]Talent, p player, want string) {
	t.Helper()
	segments, err := Compile(text, provider, spellId, talents)
	if err != nil {
		t.Fatalf("%q: %v", text, err)
	}
	if got := renderJSON(t, segments, p); got != want {
		t.Errorf("%q renders %q, want %q", text, got, want)
	}
}

var doubleSpaces = regexp.MustCompile(` {2,}`)

// The text renderTooltipText (ui/sim/proto/tooltips.ts) writes for the segments as tooltips.json carries them.
func renderJSON(t *testing.T, segments []Segment, p player) string {
	t.Helper()
	data, err := json.Marshal(segments)
	if err != nil {
		t.Fatal(err)
	}
	var decoded []any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	return doubleSpaces.ReplaceAllString(renderSegments(t, decoded, p), " ")
}

func renderSegments(t *testing.T, segments []any, p player) string {
	var out strings.Builder
	for _, segment := range segments {
		switch s := segment.(type) {
		case string:
			out.WriteString(s)
		case map[string]any:
			if cond, ok := s["if"]; ok {
				side := s["n"]
				if evaluate(t, cond, p) != 0 {
					side = s["y"]
				}
				branch, ok := side.([]any)
				if !ok {
					t.Fatalf("a branch side is %v, not a list", side)
				}
				out.WriteString(renderSegments(t, branch, p))
				continue
			}
			value := evaluate(t, s["n"], p)
			if s["abs"] == true {
				value = math.Abs(value)
			}
			decimals := int(s["p"].(float64))
			scale := math.Pow(10, float64(decimals))
			if s["t"] == true {
				// At most p decimals, the trailing zeros dropped.
				decimals = -1
			}
			out.WriteString(strconv.FormatFloat(math.Round(value*scale)/scale, 'f', decimals, 64))
		}
	}
	return out.String()
}

func evaluate(t *testing.T, expr any, p player) float64 {
	switch e := expr.(type) {
	case float64:
		return e
	case map[string]any:
		if name, ok := e["v"]; ok {
			return p.values[name.(string)]
		}
		if spell, ok := e["t"]; ok {
			rank := 1.0
			if r, ok := e["r"]; ok {
				rank = r.(float64)
			}
			return boolValue(float64(p.points[int64(spell.(float64))]) >= rank)
		}
		if fx, ok := e["fx"].(map[string]any); ok {
			return evaluateModified(t, fx, p)
		}
		args := e["a"].([]any)
		values := make([]float64, len(args))
		for i, a := range args {
			values[i] = evaluate(t, a, p)
		}
		return folds[e["o"].(string)](values...)
	}
	t.Fatalf("%v is not an expression", expr)
	return 0
}

// A value the talents can change, as evaluateModifiedValue (ui/sim/proto/tooltips.ts) reads it with done multipliers of 1.
func evaluateModified(t *testing.T, fx map[string]any, p player) float64 {
	flat, percent, wholeFlat, wholePercent := 0.0, 1.0, 0.0, 1.0
	mods, _ := fx["m"].([]any)
	for _, m := range mods {
		mod := m.(map[string]any)
		values, _ := mod["v"].([]any)
		points := p.points[int64(mod["t"].(float64))]
		if points <= 0 || len(values) == 0 {
			continue
		}
		amount := values[min(points, len(values))-1].(float64)
		switch {
		case mod["w"] == true && mod["f"] == true:
			wholeFlat += amount
		case mod["w"] == true:
			wholePercent *= 1 + amount/100
		case mod["f"] == true:
			flat += amount
		default:
			percent *= 1 + amount/100
		}
	}
	total := (evaluate(t, fx["b"], p) + flat) * percent
	shares, _ := fx["s"].([]any)
	for _, share := range shares {
		total += evaluate(t, share, p)
	}
	return (total + wholeFlat) * wholePercent
}

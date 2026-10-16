//go:build with_db

package feralbear

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
	gproto "google.golang.org/protobuf/proto"
)

const (
	spellDireBear   int32 = 9634
	spellPrimalBite int32 = 1238073
	spellMaul       int32 = 9881
	spellLacerate   int32 = 1235827
	spellSwipe      int32 = 9908
	spellEnrage     int32 = 5229
	spellDemoRoar   int32 = 9898
	spellFF         int32 = 9907
	spellFrenzied   int32 = 22842
	spellBarkskin   int32 = 22812
	spellBerserk    int32 = 417141
	itemSapper      int32 = 10646
	itemStoneshield int32 = 13455
	itemHealthstone int32 = 9421
)

// The consumables the bear UI starts with (presets.ts DefaultConsumables).
var uiBearConsumables = &proto.ConsumesSpec{
	DefensivePotId: itemStoneshield,
	ElixirIds:      []int32{13452, 9088},
	FoodId:         21023,
	MhImbueId:      12643,
	ConjuredId:     itemHealthstone,
	ExtraIds:       []int32{itemSapper, 10310, 10305},
}

// The bear UI's defaults: the Level 60 gear preset, its buffs, debuffs, healing model and
// consumables, and the default raid target swinging at the bear.
func bearRequest(rotation *proto.APLRotation, iterations int32) *proto.RaidSimRequest {
	return &proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 101, Iterations: iterations, IsTest: true},
		Raid: &proto.Raid{
			Parties: []*proto.Party{{
				Players: []*proto.Player{{
					Name: "Bear", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
					TalentsString:   DefaultTalents,
					ReactionTimeMs:  250,
					InFrontOfTarget: true,
					Profession1:     proto.Profession_Engineering,
					Profession2:     proto.Profession_Enchanting,
					Equipment:       core.GetGearSet("../../../ui/specs/druid/feralbear/gear_sets", "default").GearSet,
					Consumables:     uiBearConsumables,
					Cooldowns:       &proto.Cooldowns{HpPercentForDefensives: 0.4},
					Buffs: &proto.IndividualBuffs{
						GreaterBlessingOfKings: true,
						GreaterBlessingOfMight: true,
						Thorns:                 true,
					},
					HealingModel: &proto.HealingModel{
						Hps:               2200,
						CadenceSeconds:    0.4,
						CadenceVariation:  1.2,
						AbsorbFrac:        0.02,
						BurstWindow:       6,
						InspirationUptime: 0.25,
					},
					Spec: &proto.Player_FeralBearDruid{FeralBearDruid: &proto.FeralBearDruid{
						Options: &proto.FeralBearDruid_Options{},
					}},
					Rotation: rotation,
				}},
				Buffs: &proto.PartyBuffs{
					ManaSpringTotem: proto.TristateEffect_TristateEffectRegular,
				},
			}},
			Buffs: &proto.RaidBuffs{
				BattleShout:              proto.TristateEffect_TristateEffectRegular,
				ArcaneBrilliance:         true,
				GiftOfTheWild:            true,
				PrayerOfFortitude:        true,
				PrayerOfShadowProtection: true,
				PrayerOfSpirit:           true,
				WindfuryTotem:            true,
				StrengthOfEarthTotem:     true,
			},
			Debuffs: &proto.Debuffs{
				ExposeArmor:         true,
				FaerieFire:          true,
				CurseOfRecklessness: true,
				InsectSwarm:         true,
				JudgementOfWisdom:   true,
				SunderArmor:         true,
			},
			Tanks: []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}},
		},
		Encounter: &proto.Encounter{
			Duration: 180,
			Targets:  []*proto.Target{bearTarget()},
		},
	}
}

// The bear's default encounter (presets.ts Level60Encounter): the UI's raid target, hitting for
// 6000 before armor instead of 15113.
const bearBossDamage = 6000

func bearTarget() *proto.Target {
	target := core.FreshDefaultTargetConfig()
	target.Id = 31146
	target.Name = "Raid Target"
	target.Stats[stats.BlockValue] = 54
	target.Stats[stats.Health] = 1299000
	target.MinBaseDamage = bearBossDamage
	return target
}

// bearTune is one candidate list.
type bearTune struct {
	sapper     bool    // throw the sapper, then shift back
	berserk    int     // rage to hold before Berserk; -1 leaves it to autocast
	elune      string  // "berserk": with Berserk; "auto": left to autocast
	enrage     string  // "never", "cd": on cooldown, "safe": on cooldown above 70% health, "berserk": only into Berserk, "hold": on cooldown unless Berserk is within 60s
	enrageRage int     // Enrage only below this rage; 0 means 100
	demo       bool    // keep Demoralizing Roar up
	ff         bool    // keep the druid's own Faerie Fire up
	swipe      bool    // Swipe before the Lacerate filler while 5 stacks have more than 3s left
	filler     int     // rage the Lacerate filler leaves for Maul; -1 drops the filler
	pbHold     float64 // no filler when Primal Bite is ready within this many seconds
	maul       int     // rage at which Maul queues
	frenzied   float64 // Frenzied Regeneration below this health fraction; 0 never
	bleedFirst bool    // Berserk only once Lacerate is on the target, for Rend and Tear
	opener     string  // "roar": Berserk once Demoralizing Roar is up; "roar_pb": and Primal Bite is on cooldown
}

// bestBearAPL is the list ui/specs/druid/feralbear/apls/default.apl.json holds. The druid's own
// Faerie Fire is left out: the raid's is the same armor reduction, so the cast only spends a GCD.
// Enrage is left out: Dire Bear loses 16% of its gear armor for 10s, which costs more deaths than
// its 30 rage buys damage. Berserk waits for Demoralizing Roar, so the roar does not take a GCD
// from Berserk's Primal Bites.
var bestBearAPL = bearTune{elune: "berserk", enrage: "never", demo: true, maul: 20, frenzied: 0.2, opener: "roar"}

// The preset files and the tunes that build them.
func bearAPLPresets() map[string]*proto.APLRotation {
	return map[string]*proto.APLRotation{
		"default": bearAPL(bestBearAPL),
	}
}

// APL_WRITE=1 rewrites the presets from their builders.
func TestWriteBearAPL(t *testing.T) {
	if os.Getenv("APL_WRITE") == "" {
		t.Skip("set APL_WRITE=1 to rewrite the APL presets")
	}
	for name, rotation := range bearAPLPresets() {
		data, err := protojson.Marshal(rotation)
		if err != nil {
			t.Fatal(err)
		}
		// protojson spaces its output at random; Compact then Indent makes the file stable.
		var compact, out bytes.Buffer
		if err := json.Compact(&compact, data); err != nil {
			t.Fatal(err)
		}
		if err := json.Indent(&out, compact.Bytes(), "", "\t"); err != nil {
			t.Fatal(err)
		}
		out.WriteByte('\n')
		if err := os.WriteFile("../../../ui/specs/druid/feralbear/apls/"+name+".apl.json", out.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// The presets on disk are the lists bearAPL builds, so a sweep of a tune measures what ships.
func TestBearAPLPresetMatchesItsBuilder(t *testing.T) {
	for name, built := range bearAPLPresets() {
		if preset := core.GetAplRotation("../../../ui/specs/druid/feralbear/apls", name).Rotation; !gproto.Equal(preset, built) {
			t.Errorf("%s.apl.json differs from its builder; rerun with APL_WRITE=1", name)
		}
	}
}

// bearAPLWarnings is what the APL editor marks with a warning or an error, by priority list line;
// group i's line j is 100*(i+1)+j. Prepull lines are -1 and validations by UUID -2. Information (the blue icon) is left out.
func bearAPLWarnings(t *testing.T, req *proto.RaidSimRequest) map[int][]string {
	t.Helper()
	stats := core.ComputeStats(&proto.ComputeStatsRequest{Raid: req.Raid, Encounter: req.Encounter}).
		RaidStats.Parties[0].Players[0].RotationStats
	warnings := map[int][]string{}
	add := func(line int, validations []*proto.APLValidation) {
		for _, v := range validations {
			if v.LogLevel >= proto.LogLevel_Warning {
				warnings[line] = append(warnings[line], v.Validation)
			}
		}
	}
	for i, action := range stats.PriorityList {
		add(i, action.Validations)
	}
	// Group lines count from 100 per group.
	for i, group := range stats.Groups {
		for j, action := range group.Actions {
			add(100*(i+1)+j, action.Validations)
		}
	}
	for _, action := range stats.PrepullActions {
		add(-1, action.Validations)
	}
	for _, uuid := range stats.UuidValidations {
		add(-2, uuid.Validations)
	}
	return warnings
}

// Every line names a spell the UI's default bear knows, with or without consumables, and every
// spell the list names is cast.
func TestBearAPLPresetHasNoWarnings(t *testing.T) {
	preset := core.GetAplRotation("../../../ui/specs/druid/feralbear/apls", "default").Rotation
	for _, consumes := range []bool{true, false} {
		req := bearRequest(preset, 200)
		if !consumes {
			req.Raid.Parties[0].Players[0].Consumables = &proto.ConsumesSpec{}
			req.Raid.Parties[0].Players[0].Profession1 = proto.Profession_ProfessionUnknown
		}
		if warnings := bearAPLWarnings(t, req); len(warnings) != 0 {
			t.Errorf("consumes %v: preset warnings by line: %v", consumes, warnings)
		}
		result := core.RunSim(req, nil, simsignals.CreateSignals())
		if result.Error != nil {
			t.Fatal(result.Error.Message)
		}
		player := result.RaidMetrics.Parties[0].Players[0]
		items := append([]*proto.APLListItem{}, preset.PriorityList...)
		for _, group := range preset.Groups {
			items = append(items, group.Actions...)
		}
		for _, item := range items {
			if id := item.Action.GetCastSpell().GetSpellId().GetSpellId(); id != 0 && bearCastsOf(player.Actions, id) == 0 {
				t.Errorf("consumes %v: the preset never casts %d", consumes, id)
			}
		}
	}
}

func bearAPL(tune bearTune) *proto.APLRotation {
	vars := &bearVars{}
	groups := &bearGroups{}
	rage := currentRage()
	rageAtLeast := func(n int) *proto.APLValue {
		return cmp(proto.APLValueCompare_OpGe, rage, konst(strconv.Itoa(n)))
	}
	stacks := auraStacksOnTarget(spellLacerate)
	fiveStacks := vars.def("Lacerate at 5 stacks", and(dotUp(spellLacerate), cmp(proto.APLValueCompare_OpGe, stacks, konst("5"))))
	roarUp := auraOnTarget(spellDemoRoar)
	if tune.demo || tune.opener == "roar" || tune.opener == "roar_pb" {
		roarUp = vars.def("Demoralizing Roar is up", roarUp)
	}

	var list []*proto.APLListItem
	if tune.sapper {
		list = append(list,
			&proto.APLListItem{Action: bearItemAction(itemSapper)},
			bearCastWhen(spellDireBear, not(auraUp(spellDireBear))))
	}

	var cooldowns []*proto.APLListItem
	var berserkWhen []*proto.APLValue
	if tune.berserk > 0 {
		berserkWhen = append(berserkWhen, rageAtLeast(tune.berserk))
	}
	if tune.bleedFirst {
		berserkWhen = append(berserkWhen, dotUp(spellLacerate))
	}
	if tune.opener == "roar" || tune.opener == "roar_pb" {
		berserkWhen = append(berserkWhen, roarUp)
	}
	if tune.opener == "roar_pb" {
		berserkWhen = append(berserkWhen, not(spellReadyBear(spellPrimalBite)))
	}
	if tune.berserk >= 0 {
		var condition *proto.APLValue
		switch len(berserkWhen) {
		case 0:
		case 1:
			condition = berserkWhen[0]
		default:
			condition = and(berserkWhen...)
		}
		cooldowns = append(cooldowns, bearCastWhen(spellBerserk, condition))
	}
	if tune.elune == "berserk" {
		cooldowns = append(cooldowns, bearCastWhen(1259799, or(auraUp(spellBerserk), not(spellKnown(spellBerserk)))))
	}
	enrageRage := tune.enrageRage
	if enrageRage == 0 {
		enrageRage = 100
	}
	enrageLow := cmp(proto.APLValueCompare_OpLt, rage, konst(strconv.Itoa(enrageRage)))
	switch tune.enrage {
	case "cd":
		cooldowns = append(cooldowns, bearCastWhen(spellEnrage, enrageLow))
	case "berserk":
		cooldowns = append(cooldowns, bearCastWhen(spellEnrage, and(enrageLow, auraUp(spellBerserk))))
	case "safe":
		cooldowns = append(cooldowns, bearCastWhen(spellEnrage, and(enrageLow,
			cmp(proto.APLValueCompare_OpGe, healthPercent(), konst("70%")))))
	case "hold":
		cooldowns = append(cooldowns, bearCastWhen(spellEnrage, and(enrageLow,
			or(auraUp(spellBerserk), cmp(proto.APLValueCompare_OpGt, spellLeft(spellBerserk), konst("60s"))))))
	}
	cooldowns = append(cooldowns, &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_AutocastOtherCooldowns{
		AutocastOtherCooldowns: &proto.APLActionAutocastOtherCooldowns{},
	}}})
	list = append(list, groups.def("Cooldowns", cooldowns...))
	if tune.frenzied > 0 {
		list = append(list, bearCastWhen(spellFrenzied,
			cmp(proto.APLValueCompare_OpLt, healthPercent(), konst(strconv.FormatFloat(tune.frenzied*100, 'f', -1, 64)+"%"))))
	}

	// Five stacks about to fall off are worth more than anything else the GCD could buy.
	keepUp := []*proto.APLListItem{bearCastWhen(spellLacerate, vars.def("Lacerate falls off",
		and(fiveStacks, cmp(proto.APLValueCompare_OpLe, dotLeft(spellLacerate), konst("1.5s")))))}
	if tune.demo {
		keepUp = append(keepUp, bearCastWhen(spellDemoRoar, not(roarUp)))
	}
	if tune.ff {
		keepUp = append(keepUp, bearCastWhen(spellFF, not(auraOnTarget(spellFF))))
	}
	list = append(list, groups.def("Keep up", keepUp...))

	damage := []*proto.APLListItem{bearCastWhen(spellPrimalBite, nil)}
	var filler []*proto.APLListItem
	if tune.filler >= 0 {
		// Primal Bite and Lacerate cost the same, so a ready Primal Bite the line above skipped is
		// unaffordable for the filler too; the hold only matters above 0s.
		var gates []*proto.APLValue
		if tune.pbHold > 0 {
			gates = append(gates, cmp(proto.APLValueCompare_OpGt, spellLeft(spellPrimalBite), konst(strconv.FormatFloat(tune.pbHold, 'f', -1, 64)+"s")))
		}
		if tune.filler > 0 {
			gates = append(gates, cmp(proto.APLValueCompare_OpGe, mathOp(proto.APLValueMath_OpSub, rage, spellCost(spellLacerate)), konst(strconv.Itoa(tune.filler))))
		}
		var gate *proto.APLValue
		if len(gates) > 0 {
			gate = and(gates...)
		}
		if tune.swipe {
			filler = append(filler, bearCastWhen(spellSwipe, and(append(gates, cmp(proto.APLValueCompare_OpGt, dotLeft(spellLacerate), konst("3s")))...)))
		}
		filler = append(filler, bearCastWhen(spellLacerate, gate))
	}
	// An ungated Lacerate filler already stacks Lacerate to 5.
	if len(filler) != 1 || filler[0].Action.Condition != nil {
		damage = append(damage, bearCastWhen(spellLacerate, not(fiveStacks)))
	}
	list = append(list, groups.def("Damage", append(damage, filler...)...))
	list = append(list, bearCastWhen(spellMaul, rageAtLeast(tune.maul)))
	return &proto.APLRotation{Type: proto.APLRotation_TypeAPL, PriorityList: list, Groups: groups.list, ValueVariables: vars.list}
}

// bearVars holds the named conditions of a list, which the APL editor shows as variables.
type bearVars struct {
	list []*proto.APLValueVariable
}

// def names value and returns a reference to it.
func (v *bearVars) def(name string, value *proto.APLValue) *proto.APLValue {
	v.list = append(v.list, &proto.APLValueVariable{Name: name, Value: value})
	return &proto.APLValue{Value: &proto.APLValue_VariableRef{VariableRef: &proto.APLValueVariableRef{Name: name}}}
}

// bearGroups holds the action groups of a list, which the APL editor shows by name.
type bearGroups struct {
	list []*proto.APLGroup
}

// def adds the group and returns the line that runs it: the first of its actions that is ready.
func (g *bearGroups) def(name string, actions ...*proto.APLListItem) *proto.APLListItem {
	g.list = append(g.list, &proto.APLGroup{Name: name, Actions: actions})
	return &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_GroupReference{
		GroupReference: &proto.APLActionGroupReference{GroupName: name},
	}}}
}

func bearCastWhen(id int32, condition *proto.APLValue) *proto.APLListItem {
	action := spellAction(id)
	action.Condition = condition
	return &proto.APLListItem{Action: action}
}

func bearItemAction(id int32) *proto.APLAction {
	return &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{
		SpellId: &proto.ActionID{RawId: &proto.ActionID_ItemId{ItemId: id}},
	}}}
}

func bearSpellID(id int32) *proto.ActionID {
	return &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}}
}

func currentTarget() *proto.UnitReference {
	return &proto.UnitReference{Type: proto.UnitReference_CurrentTarget}
}

func auraUp(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_AuraIsActive{AuraIsActive: &proto.APLValueAuraIsActive{AuraId: bearSpellID(id)}}}
}

func auraOnTarget(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_AuraIsActive{AuraIsActive: &proto.APLValueAuraIsActive{
		SourceUnit: currentTarget(), AuraId: bearSpellID(id),
	}}}
}

func auraStacksOnTarget(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_AuraNumStacks{AuraNumStacks: &proto.APLValueAuraNumStacks{
		SourceUnit: currentTarget(), AuraId: bearSpellID(id),
	}}}
}

func dotUp(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_DotIsActive{DotIsActive: &proto.APLValueDotIsActive{SpellId: bearSpellID(id)}}}
}

func dotLeft(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_DotRemainingTime{DotRemainingTime: &proto.APLValueDotRemainingTime{SpellId: bearSpellID(id)}}}
}

func spellLeft(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_SpellTimeToReady{SpellTimeToReady: &proto.APLValueSpellTimeToReady{SpellId: bearSpellID(id)}}}
}

func spellCost(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_SpellCurrentCost{SpellCurrentCost: &proto.APLValueSpellCurrentCost{SpellId: bearSpellID(id)}}}
}

func spellReadyBear(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_SpellIsReady{SpellIsReady: &proto.APLValueSpellIsReady{SpellId: bearSpellID(id)}}}
}

func spellKnown(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_SpellIsKnown{SpellIsKnown: &proto.APLValueSpellIsKnown{SpellId: bearSpellID(id)}}}
}

func currentRage() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_CurrentRage{CurrentRage: &proto.APLValueCurrentRage{}}}
}

func healthPercent() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_CurrentHealthPercent{CurrentHealthPercent: &proto.APLValueCurrentHealthPercent{}}}
}

func not(value *proto.APLValue) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Not{Not: &proto.APLValueNot{Val: value}}}
}

func and(values ...*proto.APLValue) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_And{And: &proto.APLValueAnd{Vals: values}}}
}

func or(values ...*proto.APLValue) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Or{Or: &proto.APLValueOr{Vals: values}}}
}

func konst(val string) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: val}}}
}

func cmp(op proto.APLValueCompare_ComparisonOperator, lhs, rhs *proto.APLValue) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{Op: op, Lhs: lhs, Rhs: rhs}}}
}

func mathOp(op proto.APLValueMath_MathOperator, lhs, rhs *proto.APLValue) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Math{Math: &proto.APLValueMath{Op: op, Lhs: lhs, Rhs: rhs}}}
}

type bearRow struct {
	name string
	rot  *proto.APLRotation
}

type bearScenario struct {
	name       string
	duration   float64
	consumes   bool
	targets    int
	bossDamage float64 // 0 keeps the UI's default target
	bleed      bool    // Assume Bleed Always Active
}

func runBearRows(t *testing.T, rows []bearRow, scenarios []bearScenario, iterations int32) {
	for _, scenario := range scenarios {
		fmt.Printf("\n== %s\n", scenario.name)
		fmt.Printf("%-30s %7s %7s %7s %7s %6s %6s %5s %5s %5s %5s %5s %5s %5s %5s %5s %5s %5s\n",
			"rotation", "dps", "tps", "dtps", "tmi", "death", "pbite", "maul", "lacer", "swipe", "enrag", "berk", "demo", "ff", "frenz", "bark", "sap", "form")
		for _, row := range rows {
			req := bearRequest(row.rot, iterations)
			req.Encounter.Duration = scenario.duration
			if !scenario.consumes {
				req.Raid.Parties[0].Players[0].Consumables = &proto.ConsumesSpec{}
			}
			for len(req.Encounter.Targets) < max(1, scenario.targets) {
				req.Encounter.Targets = append(req.Encounter.Targets, bearTarget())
			}
			if scenario.bossDamage > 0 {
				req.Encounter.Targets[0].MinBaseDamage = scenario.bossDamage
			}
			req.Raid.Parties[0].Players[0].GetFeralBearDruid().Options.AssumeBleedActive = scenario.bleed
			if os.Getenv("BEAR_SWEEP_DEBUG") != "" {
				req.SimOptions.Debug = true
				req.SimOptions.Iterations = 1
			}
			result := core.RunSim(req, nil, simsignals.CreateSignals())
			if result.Error != nil {
				t.Fatalf("%s / %s: %s", scenario.name, row.name, result.Error.Message)
			}
			m := result.RaidMetrics.Parties[0].Players[0]
			n := float64(result.IterationsDone)
			per := func(id int32) float64 { return float64(bearCastsOf(m.Actions, id)) / n }
			fmt.Printf("%-30s %7.1f %7.1f %7.1f %7.1f %6.2f %6.1f %5.1f %5.1f %5.1f %5.2f %5.2f %5.1f %5.1f %5.2f %5.2f %5.2f %5.2f\n",
				row.name, m.Dps.Avg, m.Threat.Avg, m.Dtps.Avg, m.Tmi.Avg, m.ChanceOfDeath*100,
				per(spellPrimalBite), per(spellMaul), per(spellLacerate), per(spellSwipe), per(spellEnrage), per(spellBerserk),
				per(spellDemoRoar), per(spellFF), per(spellFrenzied), per(spellBarkskin),
				float64(itemCastsOfBear(m.Actions, itemSapper))/n, per(spellDireBear))
			if debug := os.Getenv("BEAR_SWEEP_DEBUG"); debug != "" {
				for _, line := range strings.Split(result.Logs, "\n") {
					if debug == "all" || strings.Contains(line, "Casting") {
						fmt.Println(line)
					}
				}
			}
		}
	}
}

func sweepIterations() int32 {
	if n, err := strconv.Atoi(os.Getenv("BEAR_SWEEP_ITERATIONS")); err == nil && n > 0 {
		return int32(n)
	}
	return 2000
}

func TestBearAPLSweep(t *testing.T) {
	if os.Getenv("BEAR_SWEEP") == "" {
		t.Skip("set BEAR_SWEEP=1 to compare bear APL variants")
	}
	mirror := bearTune{berserk: -1, elune: "auto", enrage: "never", demo: true, ff: true, maul: 50}
	with := func(edit func(*bearTune)) *proto.APLRotation {
		tune := mirror
		edit(&tune)
		return bearAPL(tune)
	}
	rows := []bearRow{
		{"apl mirror", bearAPL(mirror)},
	}
	mirror = bestBearAPL
	rows = append(rows,
		bearRow{"best", bearAPL(bestBearAPL)},
		bearRow{"best, no frenzied", with(func(b *bearTune) { b.frenzied = 0 })},
		bearRow{"best, frenzied 30%", with(func(b *bearTune) { b.frenzied = 0.3 })},
		bearRow{"best, enrage safe <20", with(func(b *bearTune) { b.enrage = "safe"; b.enrageRage = 20 })},
		bearRow{"best, enrage berserk", with(func(b *bearTune) { b.enrage = "berserk" })},
		bearRow{"best, enrage cd <20", with(func(b *bearTune) { b.enrage = "cd"; b.enrageRage = 20 })},
		bearRow{"best, maul 40", with(func(b *bearTune) { b.maul = 40 })},
		bearRow{"best, berserk autocast", with(func(b *bearTune) { b.berserk = -1; b.elune = "auto" })},
		bearRow{"best, berserk at the pull", with(func(b *bearTune) { b.opener = "" })},
		bearRow{"best, roar, bite then berserk", with(func(b *bearTune) { b.opener = "roar_pb" })},
		bearRow{"best, roar, lacerate, berserk", with(func(b *bearTune) { b.bleedFirst = true })},
	)
	scenarios := []bearScenario{
		{name: "60s, consumes, bleed assumed", duration: 60, consumes: true, bleed: true},
		{name: "180s, consumes, bleed assumed", duration: 180, consumes: true, bleed: true},
		{name: "180s, consumes", duration: 180, consumes: true},
		{name: "180s, no consumes", duration: 180},
		{name: "60s, consumes", duration: 60, consumes: true},
		{name: "300s, consumes", duration: 300, consumes: true},
		{name: "180s, consumes, boss 4000", duration: 180, consumes: true, bossDamage: 4000},
		{name: "180s, consumes, boss 15113", duration: 180, consumes: true, bossDamage: 15113},
	}
	if s := os.Getenv("BEAR_SWEEP_BOSS"); s != "" {
		scenarios = nil
		for _, v := range strings.Split(s, ",") {
			dmg, _ := strconv.ParseFloat(v, 64)
			scenarios = append(scenarios, bearScenario{name: "180s, consumes, boss " + v, duration: 180, consumes: true, bossDamage: dmg})
		}
	}
	runBearRows(t, rows, scenarios, sweepIterations())
}

func itemCastsOfBear(actions []*proto.ActionMetrics, itemID int32) int32 {
	var n int32
	for _, action := range actions {
		if action.GetId().GetItemId() != itemID {
			continue
		}
		for _, target := range action.Targets {
			n += target.Casts
		}
	}
	return n
}

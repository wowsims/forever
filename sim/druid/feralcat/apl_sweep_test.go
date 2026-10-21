//go:build with_db

package feralcat

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
	"google.golang.org/protobuf/encoding/protojson"
	gproto "google.golang.org/protobuf/proto"
)

const (
	spellCatForm int32 = 768
	spellShred   int32 = 9830
	spellClaw    int32 = 9850
	spellRip     int32 = 9896
	spellRake    int32 = 9904
	spellFF      int32 = 9907
	spellSP      int32 = 1322605
	spellBerserk int32 = 417141
	spellElune   int32 = 1259799
	spellClear   int32 = 16870
	itemSapper   int32 = 10646
	itemTea      int32 = 7676

	spellBite       int32 = 22829
	spellDireBear   int32 = 9634
	spellPrimalBite int32 = 1238073
	spellEnrage     int32 = 5229
	spellMaul       int32 = 9881
	spellInnervate  int32 = 29166
	// Read Ley Line, the High Order Skyborne's racial: Energized, twice the mana regeneration for 15s.
	spellEnergized int32 = 1259705
)

// aplTune is one candidate. elune is "after" (once Berserk's crit has fallen), "now" (as soon as it
// is ready), or "builder" (after the window, and only into Rake or Shred). ceiling is the energy
// at which Shifting Power still fits; 0 means the row's 50. omenMax, when set, also requires energy
// under that number before Faerie Fire fishes. innervate leaves Cat Form for a self-Innervate when
// the mana would run short.
type aplTune struct {
	elune     string
	omen      bool
	ceiling   int
	omenMax   string
	innervate bool
	// rakePool, when set, is how far ahead of Rake falling off a builder that would leave the refresh
	// unpaid waits; Rip always pools 2s ahead.
	rakePool string
	// biteEnd spends 5 combo points on Ferocious Bite once a Rip would not tick out; biteLow, when
	// set, also bites at 3 or more inside that much of the end. rakeCut, when set, casts no Rake with
	// less than that left.
	biteEnd bool
	biteLow string
	rakeCut string
	// shadowmeld casts Shadowmeld into Ravage, a Night Elf's, on cooldown once Ravage is affordable.
	shadowmeld bool
	// energized casts Read Ley Line, a High Order Skyborne's, in a free GCD once the mana is under
	// energizedMana; "cd" casts it on cooldown.
	energized     bool
	energizedMana string
	// ccShred spends Clearcasting on Shred (Claw where Shred cannot land) ahead of the bleeds,
	// Shifting Power and Ferocious Bite: Shred is the most energy Clearcasting can save.
	ccShred bool
}

// aplVars holds the named conditions of a list, which the APL editor shows as variables so a reader
// sees "Rip now" rather than the comparisons behind it.
type aplVars struct {
	list []*proto.APLValueVariable
}

// def names value and returns a reference to it; a nil value stays nil.
func (v *aplVars) def(name string, value *proto.APLValue) *proto.APLValue {
	if value == nil {
		return nil
	}
	v.list = append(v.list, &proto.APLValueVariable{Name: name, Value: value})
	return &proto.APLValue{Value: &proto.APLValue_VariableRef{VariableRef: &proto.APLValueVariableRef{Name: name}}}
}

func catAPL(tune aplTune) *proto.APLRotation {
	if tune.ceiling == 0 {
		tune.ceiling = 50
	}
	vars := &aplVars{}
	inCat := auraUp(spellCatForm)
	clearcasting := auraUp(spellClear)
	ripUp := dotUp(spellRip)
	rakeUp := dotUp(spellRake)
	cp5 := cmp(proto.APLValueCompare_OpGe, comboPoints(), konst("5"))
	// Long enough for a new Rip to tick out.
	fightLong := vars.def("Rip has time", cmp(proto.APLValueCompare_OpGe, remainingTime(), konst("10s")))
	ripNow := vars.def("Rip now", and(cp5, not(ripUp), fightLong, spellCanCast(spellRip)))
	rakeNow := and(not(rakeUp), spellCanCast(spellRake))
	if tune.rakeCut != "" {
		rakeNow = and(rakeNow, cmp(proto.APLValueCompare_OpGe, remainingTime(), konst(tune.rakeCut)))
	}
	rakeNow = vars.def("Rake now", rakeNow)
	energyLow := func(ceiling int) *proto.APLValue {
		return cmp(proto.APLValueCompare_OpLt, currentEnergy(), konst(strconv.Itoa(ceiling)))
	}
	// 100 - 40 energy - 10 regen over the 1s GCD. Clearcasting keeps filling through a second GCD,
	// so its ceiling is another 10 lower.
	underCap := or(and(clearcasting, energyLow(tune.ceiling-10)), and(not(clearcasting), energyLow(tune.ceiling)))
	wantShift := vars.def("Shifting Power now", and(spellCanCast(spellSP), underCap))
	// A builder would leave the refresh short: energy - Shred + 10 regen < Rip.
	afterBuilder := mathOp(proto.APLValueMath_OpAdd, mathOp(proto.APLValueMath_OpSub, currentEnergy(), spellCost(spellShred)), konst("10"))
	pooling := and(cp5, ripUp, fightLong,
		cmp(proto.APLValueCompare_OpGt, dotLeft(spellRip), konst("0s")),
		cmp(proto.APLValueCompare_OpLe, dotLeft(spellRip), konst("2s")),
		cmp(proto.APLValueCompare_OpLt, afterBuilder, spellCost(spellRip)))
	if tune.rakePool != "" {
		pooling = or(pooling, and(rakeUp, cmp(proto.APLValueCompare_OpGe, remainingTime(), konst("9s")),
			cmp(proto.APLValueCompare_OpGt, dotLeft(spellRake), konst("0s")),
			cmp(proto.APLValueCompare_OpLe, dotLeft(spellRake), konst(tune.rakePool)),
			cmp(proto.APLValueCompare_OpLt, afterBuilder, spellCost(spellRake))))
	}
	pooling = vars.def("Pool for a refresh", pooling)
	// Berserk is a major cooldown, so a plain cast waits for a GCD that is already idle while the
	// builder queues into the window and takes that GCD. A sequence casts it with the spell.
	// The sapper line sits above these, so a throw and the Cat Form after it always go first.
	openRake := and(inCat, not(ripNow), rakeNow)
	openShred := and(inCat, not(ripNow), not(rakeNow), not(pooling), not(wantShift), spellCanCast(spellShred))
	// Berserk is +100% crit for 15s on the combo point builders, so Elune's 10% is lost on them in that
	// window; auto attacks and Ferocious Bite still gain it.
	elune := and(not(auraUp(spellBerserk)), cmp(proto.APLValueCompare_OpGt, spellLeft(spellBerserk), konst("15s")))
	switch tune.elune {
	case "now":
		elune = nil
	case "builder":
		elune = and(elune, or(spellCanCast(spellShred), spellCanCast(spellRake)))
	}
	if !tune.energized {
		elune = vars.def("Elune's Light now", elune)
	}
	// Faerie Fire only while nothing the rotation wants is affordable before its GCD ends.
	regenEnergy := mathOp(proto.APLValueMath_OpAdd, currentEnergy(), konst("10"))
	below := func(cost *proto.APLValue) *proto.APLValue {
		return cmp(proto.APLValueCompare_OpLt, regenEnergy, cost)
	}
	rakeSoon := or(not(rakeUp), cmp(proto.APLValueCompare_OpLt, dotLeft(spellRake), konst("1s")))
	ripSoon := and(cp5, fightLong, or(not(ripUp), cmp(proto.APLValueCompare_OpLt, dotLeft(spellRip), konst("1s"))))
	omenFish := and(not(clearcasting), auraIcdReady(16864),
		cmp(proto.APLValueCompare_OpGe, spellLeft(spellSP), konst("1s")),
		below(spellCost(spellShred)),
		or(not(rakeSoon), below(spellCost(spellRake))),
		or(not(ripSoon), below(spellCost(spellRip))),
		spellCanCast(spellFF))
	if tune.omenMax != "" {
		omenFish = and(omenFish, cmp(proto.APLValueCompare_OpLt, currentEnergy(), konst(tune.omenMax)))
	}
	omenFish = vars.def("Faerie Fire for Omen", omenFish)

	groups := &aplGroups{}
	builders := []*proto.APLListItem{
		castWhen(spellShred, not(pooling)),
		// Claw is 2 cheaper than Shred, so it only goes when Shred is affordable and still did not
		// land: in front of the target, or a target that cannot be shredded.
		castWhen(spellClaw, and(not(pooling), cmp(proto.APLValueCompare_OpGe, currentEnergy(), spellCost(spellShred)))),
	}
	if tune.shadowmeld {
		// Ahead of Shred: Ravage is 350% weapon damage plus 98 for 60 energy, Shred 155% plus 80 for 42.
		ravage := and(inCat, spellReady(spellShadowmeld), not(pooling), or(clearcasting,
			cmp(proto.APLValueCompare_OpGe, currentEnergy(), spellCost(spellRavage))))
		builders = append([]*proto.APLListItem{sequenceWhen(vars.def("Shadowmeld into Ravage", ravage),
			spellAction(spellShadowmeld), spellAction(spellRavage))}, builders...)
	}
	// Elune's Light is a Night Elf's; the Energized list is a High Order Skyborne's.
	var cooldownLines []*proto.APLListItem
	if !tune.energized {
		cooldownLines = append(cooldownLines, castWhen(spellElune, elune))
	}
	cooldownLines = append(cooldownLines,
		castItemWhen(itemTea, cmp(proto.APLValueCompare_OpLe, currentEnergy(), konst("20"))),
		// The potion's own rule: the full gain still fits, with five seconds of regen left over.
		castOtherWhen(proto.OtherAction_OtherActionResourcePotion, potionFits()),
		// Trinkets and anything else the list does not name.
		&proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_AutocastOtherCooldowns{
			AutocastOtherCooldowns: &proto.APLActionAutocastOtherCooldowns{},
		}}})
	list := []*proto.APLListItem{
		castItem(itemSapper),
		castWhen(spellCatForm, not(inCat)),
		groups.def("Berserk",
			sequenceWhen(openRake, spellAction(spellBerserk), spellAction(spellRake)),
			sequenceWhen(openShred, spellAction(spellBerserk), spellAction(spellShred))),
		groups.def("Cooldowns", cooldownLines...),
	}
	if tune.ccShred {
		spend := and(inCat, clearcasting)
		if tune.shadowmeld {
			// Ravage saves more than Shred, so a ready Shadowmeld keeps Clearcasting for it.
			spend = and(spend, not(spellReady(spellShadowmeld)))
		}
		spend = vars.def("Clearcasting into Shred", spend)
		list = append(list, castWhen(spellShred, spend), castWhen(spellClaw, and(spend, not(spellCanCast(spellShred)))))
	}
	list = append(list,
		groups.def("Bleeds", castWhen(spellRip, ripNow), castWhen(spellRake, rakeNow)))
	if tune.biteEnd {
		bite := []*proto.APLValue{and(cp5, not(fightLong))}
		if tune.biteLow != "" {
			bite = append(bite, and(cmp(proto.APLValueCompare_OpGe, comboPoints(), konst("3")),
				cmp(proto.APLValueCompare_OpLt, remainingTime(), konst(tune.biteLow))))
		}
		// Ahead of Shifting Power and Shred.
		list = append(list, castWhen(spellBite, vars.def("Bite at the end", or(bite...))))
	}
	list = append(list,
		castWhen(spellSP, or(wantShift, and(pooling, spellCanCast(spellSP)))),
		groups.def("Builders", builders...))
	if tune.energized {
		ready := []*proto.APLValue{inCat, spellReady(spellEnergized), not(clearcasting), not(auraUp(spellBerserk)),
			cmp(proto.APLValueCompare_OpLe, currentEnergy(), konst("30"))}
		if tune.energizedMana != "cd" {
			ready = append(ready, cmp(proto.APLValueCompare_OpLt, currentMana(), konst(tune.energizedMana)))
		}
		// Casting it leaves no GCD for a builder, so it waits for an idle one, as Innervate does.
		list = append(list, castWhen(spellEnergized, vars.def("Energized now", and(ready...))))
	}
	if tune.innervate {
		// Casting Innervate leaves Cat Form, so it waits for a free GCD with nothing to spend energy on;
		// Cat Form follows.
		list = append(list, castWhen(spellInnervate, and(inCat, vars.def("Mana short", manaShort()),
			cmp(proto.APLValueCompare_OpLe, currentEnergy(), konst("30")),
			not(clearcasting), not(auraUp(spellBerserk)),
			cmp(proto.APLValueCompare_OpGe, remainingTime(), konst("15s")))))
	}
	if tune.omen {
		list = append(list, castWhen(spellFF, omenFish))
	}
	return &proto.APLRotation{Type: proto.APLRotation_TypeAPL, PriorityList: list, Groups: groups.list, ValueVariables: vars.list}
}

// aplGroups holds the action groups of a list, which the APL editor shows by name.
type aplGroups struct {
	list []*proto.APLGroup
}

// def adds the group and returns the line that runs it: the first of its actions that is ready.
func (g *aplGroups) def(name string, actions ...*proto.APLListItem) *proto.APLListItem {
	g.list = append(g.list, &proto.APLGroup{Name: name, Actions: actions})
	return &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_GroupReference{
		GroupReference: &proto.APLActionGroupReference{GroupName: name},
	}}}
}

// Innervate's 20s of five times the spirit regen, which goes on in the five-second rule, is about
// 3900 mana on the Pre-raid set. It is cast under 1400 mana, so the gain fits under the bar, and only
// while the mana is under 50 per second of fight left: Shifting Power spends 684 every 7s or so, so a
// pool above that lasts to the end without it.
func manaShort() *proto.APLValue {
	return and(cmp(proto.APLValueCompare_OpLt, currentMana(), konst("1400")),
		cmp(proto.APLValueCompare_OpLt, currentMana(), mathOp(proto.APLValueMath_OpMul, remainingTime(), konst("50"))))
}

// BestCatAPL is the list ui/specs/druid/feralcat/apls/default.apl.json holds.
var bestCatAPL = aplTune{elune: "after", omen: true, innervate: true, rakePool: "1s", biteEnd: true, biteLow: "3s", rakeCut: "6s", ccShred: true}

// The cat UI's default talents (presets.ts CatTalents), which the Bearweave build plays as well.
const catTalents = DefaultTalents

// weaveTune is one bear trip. energyMax is the most energy the cat goes into bear with, spLeft the
// least Shifting Power cooldown left, manaPerSecond the mana kept back per second of fight left on
// top of the two forms. leave cancels Dire Bear Form after the bite, so the rest of its GCD is spent
// out of form, rather than shifting to cat from bear once the GCD ends. maul is false for a trip with
// Primal Bite alone.
type weaveTune struct {
	energyMax     int
	spLeft        string
	manaPerSecond int
	leave         bool
	maul          bool
}

// The trip ui/specs/druid/feralcat/apls/bearweave.apl.json holds.
var bestWeave = weaveTune{energyMax: 20, spLeft: "3s", manaPerSecond: 40, leave: true, maul: true}

// bearweaveAPL is catAPL(bestCatAPL) with a Dire Bear trip for Primal Bite and Maul, the list
// ui/specs/druid/feralcat/apls/bearweave.apl.json holds.
//
// The trip starts in cat with no Clearcasting, Berserk down, Primal Bite ready and 10s of fight
// left. The trip also needs the mana for Dire Bear Form and Cat Form spare over the tune's mana per
// second of fight left; a trip is worth less than the Shifting Powers it would cost. In bear: Enrage under 20 rage (Furor gives 10 on the shift, Enrage with Wolfshead 15),
// Primal Bite, and Maul with the bite's 25 rage kept back or once the bite is on cooldown. Leaving
// bear stays in it while a bear swing with rage for Maul lands before the bite's GCD ends.
func bearweaveAPL() *proto.APLRotation {
	return weaveAPL(bestWeave)
}

func weaveAPL(tune weaveTune) *proto.APLRotation {
	rot := catAPL(bestCatAPL)
	vars := &aplVars{list: rot.ValueVariables}
	groups := &aplGroups{list: rot.Groups}
	inCat := auraUp(spellCatForm)
	inBear := auraUp(spellDireBear)
	list := rot.PriorityList
	if list[1].Action.GetCastSpell().GetSpellId().GetSpellId() != spellCatForm {
		panic("weaveAPL expects Cat Form as the second line")
	}
	biteDown := not(spellReady(spellPrimalBite))
	if tune.leave {
		list[1] = castWhen(spellCatForm, and(not(inCat), not(inBear)))
	} else {
		list[1] = castWhen(spellCatForm, and(not(inCat), or(not(inBear), biteDown)))
	}

	bear := []*proto.APLListItem{
		castWhen(spellEnrage, cmp(proto.APLValueCompare_OpLt, currentRage(), konst("20"))),
		castWhen(spellPrimalBite, nil),
	}
	if tune.maul {
		bear = append(bear, castWhen(spellMaul, or(biteDown, cmp(proto.APLValueCompare_OpGe, currentRage(), konst("25")))))
	}
	if tune.leave {
		leaveNow := []*proto.APLValue{biteDown}
		if tune.maul {
			swingFirst := and(cmp(proto.APLValueCompare_OpLe, autoTimeToNext(), gcdLeft()),
				cmp(proto.APLValueCompare_OpGe, currentRage(), konst("10")))
			leaveNow = append(leaveNow, not(swingFirst))
		}
		bear = append(bear, cancelAuraWhen(spellDireBear, vars.def("Leave bear now", and(leaveNow...))))
	}
	inBearLine := groups.def("Bear trip", bear...)
	inBearLine.Action.Condition = inBear
	window := []*proto.APLValue{spellKnown(spellPrimalBite), inCat,
		cmp(proto.APLValueCompare_OpLe, currentEnergy(), konst(strconv.Itoa(tune.energyMax))),
		spellReady(spellPrimalBite),
		not(auraUp(spellClear)),
		not(auraUp(spellBerserk)),
		cmp(proto.APLValueCompare_OpGe, remainingTime(), konst("10s")),
	}
	if tune.spLeft != "" {
		window = append(window, cmp(proto.APLValueCompare_OpGe, spellLeft(spellSP), konst(tune.spLeft)))
	}
	trip := mathOp(proto.APLValueMath_OpAdd, spellCost(spellDireBear), spellCost(spellCatForm))
	spare := mathOp(proto.APLValueMath_OpSub, currentMana(), trip)
	window = append(window, cmp(proto.APLValueCompare_OpGe, spare,
		mathOp(proto.APLValueMath_OpMul, remainingTime(), konst(strconv.Itoa(tune.manaPerSecond)))))
	enter := castWhen(spellDireBear, vars.def("Bear trip now", and(window...)))

	// The bear's GCDs go right after Cat Form; the trip starts just ahead of Faerie Fire.
	out := append([]*proto.APLListItem{}, list[:2]...)
	out = append(out, inBearLine)
	out = append(out, list[2:len(list)-1]...)
	out = append(out, enter, list[len(list)-1])
	rot.PriorityList = out
	rot.Groups = groups.list
	rot.ValueVariables = vars.list
	return rot
}

const (
	spellShadowmeld int32 = 20580
	spellProwl      int32 = 9913
	spellRavage     int32 = 9867
	spellPounce     int32 = 9827
)

// shadowmeldAPL is Monocat with Shadowmeld into Ravage on cooldown, for a Night Elf.
func shadowmeldAPL() *proto.APLRotation {
	tune := bestCatAPL
	tune.shadowmeld = true
	return catAPL(tune)
}

// energizedAPL is Monocat with Read Ley Line, for a High Order Skyborne.
func energizedAPL() *proto.APLRotation {
	tune := bestCatAPL
	tune.energized = true
	tune.energizedMana = "1400"
	return catAPL(tune)
}

// openerAPL is the Monocat list behind a Prowl one second before the pull and the opener while
// Prowl is up, ahead of the sapper, which would end it.
func openerAPL(opener int32) *proto.APLRotation {
	rot := catAPL(bestCatAPL)
	rot.PriorityList = append([]*proto.APLListItem{castWhen(opener, auraUp(spellProwl))}, rot.PriorityList...)
	rot.PrepullActions = []*proto.APLPrepullAction{{
		DoAtValue: konst("-1s"),
		Action:    spellAction(spellProwl),
	}}
	return rot
}

// The preset files and the builders that make them.
func catAPLPresets() map[string]*proto.APLRotation {
	defaultRotation := catAPL(bestCatAPL)
	for _, group := range defaultRotation.Groups {
		for _, item := range group.Actions {
			if item.Action.GetCastSpell().GetSpellId().GetItemId() == itemTea {
				item.Action.Condition = and(
					&proto.APLValue{Value: &proto.APLValue_SelectedConjured{SelectedConjured: &proto.APLValueSelectedConjured{
						ConjuredId: &proto.ActionID{RawId: &proto.ActionID_ItemId{ItemId: itemTea}},
					}}},
					item.Action.Condition)
			}
		}
	}
	return map[string]*proto.APLRotation{
		"default":    defaultRotation,
		"shadowmeld": shadowmeldAPL(),
		"energized":  energizedAPL(),
		"bearweave":  bearweaveAPL(),
		"ravage":     openerAPL(spellRavage),
		"pounce":     openerAPL(spellPounce),
	}
}

// APL_WRITE=1 rewrites the presets from their builders.
func TestWriteCatAPL(t *testing.T) {
	if os.Getenv("APL_WRITE") == "" {
		t.Skip("set APL_WRITE=1 to rewrite the APL presets")
	}
	for name, rotation := range catAPLPresets() {
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
		if err := os.WriteFile("../../../ui/specs/druid/feralcat/apls/"+name+".apl.json", out.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// The presets on disk are the lists the builders make, so a sweep of a builder measures what ships.
func TestCatAPLPresetMatchesItsBuilder(t *testing.T) {
	for name, built := range catAPLPresets() {
		if preset := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", name).Rotation; !gproto.Equal(preset, built) {
			t.Errorf("%s.apl.json differs from its builder; rerun with APL_WRITE=1", name)
		}
	}
}

// The consumables, professions and race the cat UI starts with (presets.ts DefaultConsumables
// and OtherDefaults).
func uiDefaultCat(rotation *proto.APLRotation) *proto.ComputeStatsRequest {
	req := sweepRequest(powershiftTalents, rotation, false, 1)
	player := req.Raid.Parties[0].Players[0]
	player.Profession1 = proto.Profession_Engineering
	player.Profession2 = proto.Profession_Enchanting
	player.Consumables = &proto.ConsumesSpec{
		ResourcePotId: 13444,
		ElixirIds:     []int32{13452, 13447},
		FoodId:        250072,
		MhImbueId:     12643,
		ConjuredId:    7676,
		ExtraIds:      []int32{10646, 10310},
	}
	return &proto.ComputeStatsRequest{Raid: req.Raid, Encounter: req.Encounter}
}

// aplWarnings is what the APL editor marks with a warning or an error, by line: the priority list
// line's index, "prepull", "uuid", or a group's name and line, such as "Cooldowns 1". Information
// (the blue icon) is left out.
func aplWarnings(t *testing.T, preset *proto.APLRotation, req *proto.ComputeStatsRequest) map[string][]string {
	t.Helper()
	stats := core.ComputeStats(req).RaidStats.Parties[0].Players[0].RotationStats
	warnings := map[string][]string{}
	add := func(line string, validations []*proto.APLValidation) {
		for _, v := range validations {
			if v.LogLevel >= proto.LogLevel_Warning {
				warnings[line] = append(warnings[line], v.Validation)
			}
		}
	}
	for i, action := range stats.PriorityList {
		add(strconv.Itoa(i), action.Validations)
	}
	for i, group := range stats.Groups {
		for j, action := range group.Actions {
			add(preset.Groups[i%len(preset.Groups)].Name+" "+strconv.Itoa(j), action.Validations)
		}
	}
	for _, action := range stats.PrepullActions {
		add("prepull", action.Validations)
	}
	for _, uuid := range stats.UuidValidations {
		add("uuid", uuid.Validations)
	}
	return warnings
}

// Every line of the preset, its groups' lines included, keyed as aplWarnings keys them.
func aplLines(preset *proto.APLRotation) map[string]*proto.APLListItem {
	lines := map[string]*proto.APLListItem{}
	for i, item := range preset.PriorityList {
		lines[strconv.Itoa(i)] = item
	}
	for _, group := range preset.Groups {
		for j, item := range group.Actions {
			lines[group.Name+" "+strconv.Itoa(j)] = item
		}
	}
	return lines
}

// The autocast line is there for an on-use trinket the list does not name. With Innervate named,
// a set with none leaves it nothing to cast, and the editor says so on that line.
func withoutIdleAutocast(preset *proto.APLRotation, warnings map[string][]string) map[string][]string {
	for key, item := range aplLines(preset) {
		if item.Action.GetAutocastOtherCooldowns() == nil {
			continue
		}
		var kept []string
		for _, w := range warnings[key] {
			if !strings.HasPrefix(w, "Autocast Other Cooldowns will not cast any spells") {
				kept = append(kept, w)
			}
		}
		if len(kept) == 0 {
			delete(warnings, key)
		} else {
			warnings[key] = kept
		}
	}
	return warnings
}

// Every line of the preset names a spell or item the UI's default cat can cast.
func TestCatAPLPresetHasNoWarningsOnTheDefaultCat(t *testing.T) {
	preset := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "default").Rotation
	if warnings := withoutIdleAutocast(preset, aplWarnings(t, preset, uiDefaultCat(preset))); len(warnings) != 0 {
		t.Errorf("preset warnings by line: %v", warnings)
	}
}

// The bearweave preset casts everything it names on the talents it is played with.
func TestBearweaveAPLPresetHasNoWarningsOnItsTalents(t *testing.T) {
	preset := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "bearweave").Rotation
	req := uiDefaultCat(preset)
	req.Raid.Parties[0].Players[0].TalentsString = catTalents
	if warnings := withoutIdleAutocast(preset, aplWarnings(t, preset, req)); len(warnings) != 0 {
		t.Errorf("preset warnings by line: %v", warnings)
	}
}

// Without the consumables or Engineering, only the lines that cast them warn, and the sim still runs.
func TestCatAPLPresetWarnsOnlyOnMissingConsumables(t *testing.T) {
	preset := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "default").Rotation
	req := uiDefaultCat(preset)
	player := req.Raid.Parties[0].Players[0]
	player.Profession1 = proto.Profession_ProfessionUnknown
	player.Consumables = &proto.ConsumesSpec{}

	allowed := map[string]bool{}
	for key, item := range aplLines(preset) {
		id := item.Action.GetCastSpell().GetSpellId()
		if id.GetItemId() != 0 || id.GetOtherId() == proto.OtherAction_OtherActionResourcePotion {
			allowed[key] = true
		}
	}
	for line, messages := range withoutIdleAutocast(preset, aplWarnings(t, preset, req)) {
		if !allowed[line] {
			t.Errorf("line %s warns without consumables: %v", line, messages)
		}
	}

	result := core.RunSim(&proto.RaidSimRequest{
		Raid: req.Raid, Encounter: req.Encounter,
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 20, IsTest: true},
	}, nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}
	if dps := result.RaidMetrics.Parties[0].Players[0].Dps.Avg; dps < 500 {
		t.Errorf("bare cat on the preset does %.1f DPS", dps)
	}
}

type aplScenario struct {
	name     string
	duration float64
	consumes bool
	front    bool
	wolfhead bool
}

// withTune is bestCatAPL with one change, so a sweep row differs from the preset in that change only.
func withTune(edit func(*aplTune)) *proto.APLRotation {
	tune := bestCatAPL
	edit(&tune)
	return catAPL(tune)
}

func TestAPLSweep(t *testing.T) {
	if os.Getenv("ROTATION_SWEEP") == "" {
		t.Skip("set ROTATION_SWEEP=1 to compare cat APL variants")
	}
	iterations := int32(2000)
	if n, err := strconv.Atoi(os.Getenv("ROTATION_SWEEP_ITERATIONS")); err == nil && n > 0 {
		iterations = int32(n)
	}
	rows := []struct {
		name string
		rot  *proto.APLRotation
	}{
		{"apl preset (default.apl.json)", core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "default").Rotation},
		{"apl, elune at once", withTune(func(t *aplTune) { t.elune = "now" })},
		{"apl, elune into a builder", withTune(func(t *aplTune) { t.elune = "builder" })},
		{"apl, shift cap 45", withTune(func(t *aplTune) { t.ceiling = 45 })},
		{"apl, shift cap 55", withTune(func(t *aplTune) { t.ceiling = 55 })},
		{"apl, omen only under 20", withTune(func(t *aplTune) { t.omenMax = "20" })},
		{"apl, no omen fishing", withTune(func(t *aplTune) { t.omen = false })},
	}
	scenarios := []aplScenario{
		{name: "180s, consumes", duration: 180, consumes: true},
		{name: "180s, no consumes", duration: 180},
		{name: "60s, consumes", duration: 60, consumes: true},
		{name: "300s, consumes", duration: 300, consumes: true},
		{name: "180s, consumes, front (claw)", duration: 180, consumes: true, front: true},
		{name: "180s, consumes, wolfshead", duration: 180, consumes: true, wolfhead: true},
	}
	for _, scenario := range scenarios {
		fmt.Printf("\n== %s\n", scenario.name)
		fmt.Printf("%-34s %8s %7s %6s %6s %6s %6s %6s %6s %6s %6s %6s %5s\n",
			"rotation", "dps", "stdev", "sp", "shred", "claw", "rip", "rake", "berk", "elune", "tea", "ff", "pot")
		for _, row := range rows {
			req := sweepRequest(powershiftTalents, row.rot, scenario.front, iterations)
			req.Encounter.Duration = scenario.duration
			player := req.Raid.Parties[0].Players[0]
			player.Profession1 = proto.Profession_Engineering
			if scenario.consumes {
				player.Consumables = DefaultConsumables
			}
			if scenario.wolfhead {
				player.Equipment = gproto.Clone(player.Equipment).(*proto.EquipmentSpec)
				player.Equipment.Items[proto.ItemSlot_ItemSlotHead] = &proto.ItemSpec{Id: 8345}
			}
			if os.Getenv("ROTATION_SWEEP_DEBUG") != "" {
				req.SimOptions.Debug = true
				req.SimOptions.Iterations = 1
			}
			result := core.RunSim(req, nil, simsignals.CreateSignals())
			if result.Error != nil {
				t.Fatalf("%s / %s: %s", scenario.name, row.name, result.Error.Message)
			}
			metrics := result.RaidMetrics.Parties[0].Players[0]
			n := float64(result.IterationsDone)
			per := func(id int32) float64 { return float64(castsOf(metrics.Actions, id)) / n }
			fmt.Printf("%-34s %8.2f %7.2f %6.1f %6.1f %6.1f %6.1f %6.1f %6.2f %6.2f %6.2f %6.1f %5.2f\n",
				row.name, metrics.Dps.Avg, metrics.Dps.Stdev,
				per(spellSP), per(spellShred), per(spellClaw), per(spellRip), per(spellRake), per(spellBerserk), per(spellElune),
				float64(itemCastsOf(metrics.Actions, itemTea))/n, per(spellFF),
				float64(itemCastsOf(metrics.Actions, 13444))/n)
			if os.Getenv("ROTATION_SWEEP_DEBUG") != "" {
				for _, line := range strings.Split(result.Logs, "\n") {
					if strings.Contains(line, "Casting") && !strings.Contains(line, "OtherID") {
						fmt.Println(line)
					}
				}
			}
		}
	}
}

func castItem(id int32) *proto.APLListItem {
	return &proto.APLListItem{Action: itemAction(id)}
}

func castItemWhen(id int32, condition *proto.APLValue) *proto.APLListItem {
	action := itemAction(id)
	action.Condition = condition
	return &proto.APLListItem{Action: action}
}

func itemAction(id int32) *proto.APLAction {
	return &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{
		SpellId: &proto.ActionID{RawId: &proto.ActionID_ItemId{ItemId: id}},
	}}}
}

func castOtherWhen(id proto.OtherAction, condition *proto.APLValue) *proto.APLListItem {
	return &proto.APLListItem{Action: &proto.APLAction{
		Condition: condition,
		Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{
			SpellId: &proto.ActionID{RawId: &proto.ActionID_OtherId{OtherId: id}},
		}},
	}}
}

// Major Mana Potion restores 1350 to 2250. Autocast waits until the larger roll fits under the
// mana bar with five seconds of regen still to come.
func potionFits() *proto.APLValue {
	regen := mathOp(proto.APLValueMath_OpMul, manaRegen(), konst("5"))
	used := mathOp(proto.APLValueMath_OpAdd, currentMana(), mathOp(proto.APLValueMath_OpAdd, regen, konst("2250")))
	return cmp(proto.APLValueCompare_OpLe, used, maxMana())
}

func sequenceWhen(condition *proto.APLValue, actions ...*proto.APLAction) *proto.APLListItem {
	return &proto.APLListItem{Action: &proto.APLAction{
		Condition: condition,
		Action:    &proto.APLAction_StrictSequence{StrictSequence: &proto.APLActionStrictSequence{Actions: actions}},
	}}
}

func currentMana() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_CurrentMana{CurrentMana: &proto.APLValueCurrentMana{}}}
}

func maxMana() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_MaxMana{MaxMana: &proto.APLValueMaxMana{}}}
}

func manaRegen() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_CurrentManaRegen{CurrentManaRegen: &proto.APLValueCurrentManaRegen{}}}
}

func spellID(id int32) *proto.ActionID {
	return &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}}
}

func auraIcdReady(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_AuraIcdIsReady{AuraIcdIsReady: &proto.APLValueAuraICDIsReady{
		AuraId: spellID(id),
	}}}
}

func spellCanCast(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_SpellCanCast{SpellCanCast: &proto.APLValueSpellCanCast{SpellId: spellID(id)}}}
}

func spellLeft(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_SpellTimeToReady{SpellTimeToReady: &proto.APLValueSpellTimeToReady{SpellId: spellID(id)}}}
}

func spellCost(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_SpellCurrentCost{SpellCurrentCost: &proto.APLValueSpellCurrentCost{SpellId: spellID(id)}}}
}

func dotUp(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_DotIsActive{DotIsActive: &proto.APLValueDotIsActive{SpellId: spellID(id)}}}
}

func dotLeft(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_DotRemainingTime{DotRemainingTime: &proto.APLValueDotRemainingTime{SpellId: spellID(id)}}}
}

func comboPoints() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_CurrentComboPoints{CurrentComboPoints: &proto.APLValueCurrentComboPoints{}}}
}

func remainingTime() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_RemainingTime{RemainingTime: &proto.APLValueRemainingTime{}}}
}

func mathOp(op proto.APLValueMath_MathOperator, lhs, rhs *proto.APLValue) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Math{Math: &proto.APLValueMath{Op: op, Lhs: lhs, Rhs: rhs}}}
}

func currentRage() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_CurrentRage{CurrentRage: &proto.APLValueCurrentRage{}}}
}

func autoTimeToNext() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_AutoTimeToNext{AutoTimeToNext: &proto.APLValueAutoTimeToNext{}}}
}

func gcdLeft() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_GcdTimeToReady{GcdTimeToReady: &proto.APLValueGCDTimeToReady{}}}
}

func cancelAuraWhen(id int32, condition *proto.APLValue) *proto.APLListItem {
	return &proto.APLListItem{Action: &proto.APLAction{Condition: condition, Action: &proto.APLAction_CancelAura{
		CancelAura: &proto.APLActionCancelAura{AuraId: spellID(id)},
	}}}
}

// The Energized preset is a High Order Skyborne's: every line names something that race can cast, and
// Read Ley Line goes out on a long fight.
func TestEnergizedAPLPresetCastsReadLeyLine(t *testing.T) {
	preset := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "energized").Rotation
	req := uiDefaultCat(preset)
	req.Raid.Parties[0].Players[0].Race = proto.Race_RaceHighOrderSkyborne
	if warnings := withoutIdleAutocast(preset, aplWarnings(t, preset, req)); len(warnings) != 0 {
		t.Errorf("preset warnings by line: %v", warnings)
	}
	sim := uiDefaultRaidSim(catTalents, preset, 300, 0, 200)
	sim.Raid.Parties[0].Players[0].Race = proto.Race_RaceHighOrderSkyborne
	if castsOf(runCat(t, sim).Actions, spellEnergized) == 0 {
		t.Error("a 300s fight never cast Read Ley Line")
	}
}

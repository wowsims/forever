//go:build with_db

package feralcat

import (
	"strconv"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

// Same tree as the cat UI default. Shifting Power is the 11th feral talent.
const powershiftTalents = "05002-55000032021032212051-055"

// The Monocat APL the cat UI sends by default.
func defaultCatRotation() *proto.APLRotation {
	return catAPL(bestCatAPL)
}

func TestMonocatStaysInCatForm(t *testing.T) {
	player := runCat(t, catPowershiftRequest(powershiftTalents, defaultCatRotation(), false))
	for _, c := range []struct {
		name string
		id   int32
	}{{"Shifting Power", 1322605}, {"Shred", 9830}, {"Rip", 9896}, {"Rake", 9904}, {"Berserk", 417141}, {"Faerie Fire", 9907}} {
		if castsOf(player.Actions, c.id) == 0 {
			t.Errorf("%s was never cast", c.name)
		}
	}
	// The only way out of Cat Form is the self-Innervate.
	if forms, innervates := castsOf(player.Actions, 768), castsOf(player.Actions, spellInnervate); forms != innervates {
		t.Errorf("Cat Form casts = %d, Innervate casts = %d", forms, innervates)
	}
}

// Ferocious Bite spends the combo points only once a Rip would not tick out, in the last 10 seconds.
func TestMonocatBitesOnlyAtTheEnd(t *testing.T) {
	request := catPowershiftRequest(powershiftTalents, defaultCatRotation(), false)
	request.SimOptions.Debug = true
	result := core.RunSim(request, nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}
	bites := 0
	for _, line := range strings.Split(result.Logs, "\n") {
		if !strings.Contains(line, "Casting {SpellID: 22829}") {
			continue
		}
		bites++
		at, err := strconv.ParseFloat(line[1:strings.Index(line, "]")], 64)
		if err != nil {
			t.Fatal(err)
		}
		if at < request.Encounter.Duration-10 {
			t.Errorf("Ferocious Bite at %.2fs of %.0fs", at, request.Encounter.Duration)
		}
	}
	if bites == 0 {
		t.Error("Ferocious Bite was never cast at the end")
	}
}

func TestMonocatKeepsRakeUp(t *testing.T) {
	player := runCat(t, catPowershiftRequest(powershiftTalents, defaultCatRotation(), false))
	// Rake lasts 9 s, so a 90 s fight that keeps it up casts it about ten times.
	if got := castsOf(player.Actions, 9904); got < 8 {
		t.Errorf("rake casts = %d in 90 s, want it kept up", got)
	}
}

func TestMonocatWithoutShiftingPowerTalent(t *testing.T) {
	player := runCat(t, catPowershiftRequest(withFeralDigit(powershiftTalents, 10, '0'), defaultCatRotation(), false))
	if got := castsOf(player.Actions, 1322605); got != 0 {
		t.Fatalf("shifting power casts = %d, the talent is not taken", got)
	}
	if castsOf(player.Actions, 9896) == 0 {
		t.Fatal("rip was never cast")
	}
}

func TestMonocatClawsWhenShredCannotLand(t *testing.T) {
	player := runCat(t, catPowershiftRequest(powershiftTalents, defaultCatRotation(), true))
	if got := castsOf(player.Actions, 9830); got != 0 {
		t.Errorf("shred casts = %d against a target that cannot be shredded", got)
	}
	if castsOf(player.Actions, 9850) == 0 || castsOf(player.Actions, 9896) == 0 {
		t.Error("claw did not build combo points into rip")
	}
}

// Goblin Sapper Charge 10646 needs Engineering and cannot be used in Cat Form: the APL throws it and
// casts Cat Form again. EZ-Thro Dark Bomb 260817 needs no profession and never leaves the form.
func TestMonocatLeavesCatFormOnlyForEngineeringExplosives(t *testing.T) {
	for _, c := range []struct {
		name       string
		itemID     int32
		leavesForm bool
	}{{"Goblin Sapper Charge", 10646, true}, {"EZ-Thro Dark Bomb", 260817, false}} {
		request := catPowershiftRequest(powershiftTalents, defaultCatRotation(), false)
		cat := request.Raid.Parties[0].Players[0]
		cat.Profession1 = proto.Profession_Engineering
		cat.Consumables = &proto.ConsumesSpec{ExtraIds: []int32{c.itemID}}
		player := runCat(t, request)

		thrown := itemCastsOf(player.Actions, c.itemID)
		if thrown == 0 {
			t.Errorf("%s was never thrown", c.name)
		}
		want := castsOf(player.Actions, spellInnervate)
		if c.leavesForm {
			want += thrown
		}
		if got := castsOf(player.Actions, 768); got != want {
			t.Errorf("%s: cat form casts = %d after %d throws, want %d", c.name, got, thrown, want)
		}
	}
}

// Berserk's crit covers Shred and Rake, so the APL casts it in a sequence with one of them, at the
// same moment.
func TestMonocatOpensBerserkWithABuilder(t *testing.T) {
	request := catPowershiftRequest(powershiftTalents, defaultCatRotation(), false)
	request.SimOptions.Debug = true
	result := core.RunSim(request, nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}
	lines := strings.Split(result.Logs, "\n")
	found := false
	for i, line := range lines {
		if !strings.Contains(line, "Casting {SpellID: 417141}") {
			continue
		}
		found = true
		at := line[:strings.Index(line, "]")+1]
		builder := false
		for _, next := range lines[i+1:] {
			if !strings.HasPrefix(next, at) {
				break
			}
			if strings.Contains(next, "Casting {SpellID: 9830}") || strings.Contains(next, "Casting {SpellID: 9904}") {
				builder = true
				break
			}
		}
		if !builder {
			t.Errorf("Berserk at %s is not followed by Shred or Rake", at)
		}
	}
	if !found {
		t.Fatal("Berserk was never cast")
	}
}

func itemCastsOf(actions []*proto.ActionMetrics, itemID int32) int32 {
	var n int32
	for _, action := range actions {
		if action.GetId().GetItemId() != itemID || action.GetId().GetTag() != 0 {
			continue
		}
		for _, target := range action.Targets {
			n += target.Casts
		}
	}
	return n
}

func runCat(t *testing.T, request *proto.RaidSimRequest) *proto.UnitMetrics {
	t.Helper()
	result := core.RunSim(request, nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatalf("sim error: %s", result.Error.Message)
	}
	player := result.RaidMetrics.Parties[0].Players[0]
	if player.Dps.Avg <= 0 {
		t.Fatalf("dps = %v, want a fight that deals damage", player.Dps.Avg)
	}
	return player
}

func catPowershiftRequest(talents string, rotation *proto.APLRotation, cannotShred bool) *proto.RaidSimRequest {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 1613
	return &proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1, IsTest: true},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Cat", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
				TalentsString:  talents,
				ReactionTimeMs: 100,
				Equipment:      &proto.EquipmentSpec{Items: items},
				Consumables:    &proto.ConsumesSpec{},
				Buffs:          &proto.IndividualBuffs{},
				Spec: &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{
					Options: &proto.FeralCatDruid_Options{CannotShredTarget: cannotShred},
				}},
				Rotation: rotation,
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 90},
	}
}

// Bleed Is Active sees the cat's own Rake and the assumed raid bleed. The APL claws while the target
// bleeds and rakes while it does not: without the assumption Rake opens and Claw follows while it
// ticks; with it the target bleeds from the pull and Rake is never cast.
func TestBleedIsActiveSeesRakeAndTheAssumedBleed(t *testing.T) {
	const rake, claw = 9904, 9850
	bleeding := &proto.APLValue{Value: &proto.APLValue_DruidBleedIsActive{DruidBleedIsActive: &proto.APLValueDruidBleedIsActive{}}}
	rotation := &proto.APLRotation{PriorityList: []*proto.APLListItem{castWhen(claw, bleeding), castWhen(rake, not(bleeding))}}

	player := runCat(t, catPowershiftRequest("", rotation, false))
	if rakes, claws := castsOf(player.Actions, rake), castsOf(player.Actions, claw); rakes == 0 || claws == 0 {
		t.Errorf("Rake %d and Claw %d casts, want Rake to open and Claw while it ticks", rakes, claws)
	}

	assumed := catPowershiftRequest("", rotation, false)
	assumed.Raid.Parties[0].Players[0].GetFeralCatDruid().Options.AssumeBleedActive = true
	player = runCat(t, assumed)
	if got := castsOf(player.Actions, rake); got != 0 {
		t.Errorf("Rake casts = %d with a bleed assumed, want 0", got)
	}
	if castsOf(player.Actions, claw) == 0 {
		t.Error("Claw was never cast with a bleed assumed")
	}
}

func TestManaPotionAndThistleTeaAreUsed(t *testing.T) {
	const iterations int32 = 100
	req := sweepRequest(powershiftTalents, defaultCatRotation(), false, iterations)
	req.Raid.Parties[0].Players[0].Consumables = &proto.ConsumesSpec{ResourcePotId: 13444, ConjuredId: 7676}
	result := core.RunSim(req, nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}
	actions := result.RaidMetrics.Parties[0].Players[0].Actions
	per := func(id int32) float64 { return float64(usesOfID(actions, id)) / float64(result.IterationsDone) }
	if tea, pot := per(7676), per(13444); tea < 0.8 || pot < 0.8 {
		t.Fatalf("thistle tea %.2f, mana potion %.2f per fight, want both near 1", tea, pot)
	}
}

func usesOfID(actions []*proto.ActionMetrics, id int32) int32 {
	var n int32
	for _, action := range actions {
		got := action.GetId()
		if got.GetSpellId() != id && got.GetItemId() != id {
			continue
		}
		for _, target := range action.Targets {
			n += target.Casts
		}
	}
	return n
}

func castsOf(actions []*proto.ActionMetrics, spellID int32) int32 {
	var n int32
	for _, action := range actions {
		if action.GetId().GetSpellId() != spellID {
			continue
		}
		for _, target := range action.Targets {
			n += target.Casts
		}
	}
	return n
}

func TestJVMESSapperOnlyWhenSelected(t *testing.T) {
	for _, selected := range []bool{false, true} {
		t.Run(strconv.FormatBool(selected), func(t *testing.T) {
			rotation := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "jvmes").Rotation
			request := sweepRequest(powershiftTalents, rotation, false, 1)
			cat := request.Raid.Parties[0].Players[0]
			cat.Profession1 = proto.Profession_Engineering
			cat.Consumables = &proto.ConsumesSpec{ConjuredId: 7676}
			if selected {
				cat.Consumables.ExtraIds = []int32{10646}
			}
			player := runCat(t, request)
			thrown := itemCastsOf(player.Actions, 10646)
			if selected && thrown != 1 {
				t.Errorf("Sapper casts = %d, want 1 when selected", thrown)
			}
			if !selected && thrown != 0 {
				t.Errorf("Sapper casts = %d without the consumable selected", thrown)
			}
			if got, want := castsOf(player.Actions, 768), thrown+castsOf(player.Actions, spellInnervate); got != want {
				t.Errorf("Cat Form casts = %d, want %d for Sapper and Innervate", got, want)
			}
			if castsOf(player.Actions, 9907) != 0 {
				t.Error("Faerie Fire was cast despite the supplied armor debuff")
			}
		})
	}
}

func TestJVMESUsesAvailableRacials(t *testing.T) {
	for _, race := range []proto.Race{proto.Race_RaceNightElf, proto.Race_RaceHighOrderSkyborne, proto.Race_RaceWindshaperSkyborne, proto.Race_RaceTauren} {
		t.Run(race.String(), func(t *testing.T) {
			rotation := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "jvmes").Rotation
			request := sweepRequest(powershiftTalents, rotation, false, 10)
			request.Encounter.Duration = 300
			cat := request.Raid.Parties[0].Players[0]
			cat.Race = race
			cat.Consumables = &proto.ConsumesSpec{ConjuredId: 7676}
			player := runCat(t, request)
			for _, racial := range []struct {
				id   int32
				race proto.Race
			}{{1259799, proto.Race_RaceNightElf}, {20580, proto.Race_RaceNightElf}, {1259705, proto.Race_RaceHighOrderSkyborne}} {
				casts := castsOf(player.Actions, racial.id)
				if (casts > 0) != (race == racial.race) {
					t.Errorf("racial %d casts = %d for %s", racial.id, casts, race)
				}
			}
			if melds, ravages := castsOf(player.Actions, 20580), castsOf(player.Actions, 9867); ravages != melds+request.SimOptions.Iterations {
				t.Errorf("Shadowmeld casts = %d, Ravage casts = %d, want one opener per iteration plus each Shadowmeld", melds, ravages)
			}
			if forms, innervates := castsOf(player.Actions, 768), castsOf(player.Actions, spellInnervate); forms != innervates {
				t.Errorf("Cat Form casts = %d, Innervate casts = %d without Sapper", forms, innervates)
			}
		})
	}
}

func TestJVMESShadowmeldUsesClearcasting(t *testing.T) {
	for _, prepull := range []bool{false, true} {
		t.Run(strconv.FormatBool(prepull), func(t *testing.T) {
			rotation := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "jvmes").Rotation
			request := sweepRequest(powershiftTalents, rotation, false, 1)
			request.Encounter.Duration = 300
			request.SimOptions.Debug = true
			player := request.Raid.Parties[0].Players[0]
			player.Race = proto.Race_RaceNightElf
			player.GetFeralCatDruid().Options.PrepullClearcasting = prepull
			result := core.RunRaidSim(request)
			if result.Error != nil {
				t.Fatal(result.Error.Message)
			}
			active, combatProc, pendingRavage := false, false, false
			melds := 0
			for _, line := range strings.Split(result.Logs, "\n") {
				if strings.Contains(line, "Aura gained: {SpellID: 16870}") {
					active = true
					combatProc = combatProc || !strings.HasPrefix(line, "[-")
				}
				if strings.Contains(line, "Casting {SpellID: 20580}") {
					if !active || !combatProc {
						t.Fatalf("Shadowmeld without an in-combat Clearcasting proc: %s", line)
					}
					melds++
					pendingRavage = true
				}
				if pendingRavage && strings.Contains(line, "Casting {SpellID: 9867}") {
					if !strings.Contains(line, "Cost = 0.000") {
						t.Fatalf("Ravage after Shadowmeld spent energy: %s", line)
					}
					pendingRavage = false
				}
				if strings.Contains(line, "Aura faded: {SpellID: 16870}") {
					if combatProc && melds == 0 {
						t.Fatal("the first in-combat Clearcasting proc was spent before Shadowmeld")
					}
					active = false
				}
			}
			if melds == 0 || pendingRavage {
				t.Fatalf("Shadowmeld casts = %d, unfinished Ravage = %v", melds, pendingRavage)
			}
		})
	}
}

func TestJVMESRavageOpener(t *testing.T) {
	for _, tc := range []struct {
		name         string
		talents      string
		clearcasting bool
		berserk      bool
	}{
		{"paid Ravage", powershiftTalents, false, true},
		{"Clearcasting Ravage", powershiftTalents, true, true},
		{"without Berserk", "05002-55000032021032212050-055", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rotation := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "jvmes").Rotation
			request := sweepRequest(tc.talents, rotation, false, 1)
			request.Raid.Parties[0].Players[0].GetFeralCatDruid().Options.PrepullClearcasting = tc.clearcasting
			request.SimOptions.Debug = true
			result := core.RunSim(request, nil, simsignals.CreateSignals())
			if result.Error != nil {
				t.Fatal(result.Error.Message)
			}
			berserk := strings.Index(result.Logs, "[0.00] [Cat (#1)] Casting {SpellID: 417141}")
			ravage := strings.Index(result.Logs, "[0.00] [Cat (#1)] Casting {SpellID: 9867}")
			if ravage < 0 {
				t.Fatal("the opener did not cast Ravage at pull")
			}
			if tc.berserk && (berserk < 0 || ravage < berserk) {
				t.Fatal("the opener did not cast Berserk before Ravage")
			}
			if !tc.berserk && berserk >= 0 {
				t.Fatal("Berserk was cast without the talent")
			}
		})
	}
}

func TestJVMESResourcePotionOnlyWhenSelected(t *testing.T) {
	for _, potion := range []int32{0, 13444, 13442} {
		t.Run(strconv.Itoa(int(potion)), func(t *testing.T) {
			rotation := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "jvmes").Rotation
			request := sweepRequest(powershiftTalents, rotation, false, 10)
			request.Encounter.Duration = 300
			request.Raid.Debuffs.JudgementOfWisdom = false
			cat := request.Raid.Parties[0].Players[0]
			cat.Consumables = &proto.ConsumesSpec{OffensivePotId: 13442, ConjuredId: 7676}
			cat.Consumables.ResourcePotId = potion
			if potion == 13442 {
				cat.Consumables.OffensivePotId = 0
			}
			player := runCat(t, request)
			used := itemCastsOf(player.Actions, 13444)
			if potion == 13444 && used == 0 {
				t.Error("The selected mana potion was never used without Judgement of Wisdom")
			}
			if potion != 13444 && used != 0 {
				t.Errorf("Mana potion casts = %d without a resource potion selected", used)
			}
			if itemCastsOf(player.Actions, 13442) == 0 {
				t.Error("The offensive potion was never used")
			}
			for _, resource := range player.Resources {
				if resource.Id.GetItemId() == 13444 && resource.Type == proto.ResourceType_ResourceTypeMana && resource.Gain != resource.ActualGain {
					t.Errorf("Mana potion restored %v of %v mana", resource.ActualGain, resource.Gain)
				}
			}
		})
	}
}

func TestJVMESManaBudgetAccountsForFightEndAndWisdom(t *testing.T) {
	for _, tc := range []struct {
		name     string
		duration float64
		wisdom   bool
		want     int32
	}{
		{"short fight with Wisdom", 30, true, 9830},
		{"long fight with Wisdom", 180, true, 9850},
		{"short fight without Wisdom", 30, false, 9850},
	} {
		t.Run(tc.name, func(t *testing.T) {
			saved := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "jvmes").Rotation
			condition := &proto.APLValue{Value: &proto.APLValue_VariableRef{VariableRef: &proto.APLValueVariableRef{Name: "Current mana may not last until the next potion opportunity"}}}
			rotation := &proto.APLRotation{ValueVariables: saved.ValueVariables, PriorityList: []*proto.APLListItem{castWhen(9850, condition), castWhen(9830, nil)}}
			request := sweepRequest(powershiftTalents, rotation, false, 1)
			request.Encounter.Duration = tc.duration
			request.Raid.Debuffs.JudgementOfWisdom = tc.wisdom
			computed := core.ComputeStats(&proto.ComputeStatsRequest{Raid: request.Raid, Encounter: request.Encounter})
			if computed.ErrorResult != "" {
				t.Fatal(computed.ErrorResult)
			}
			mana := computed.RaidStats.Parties[0].Players[0].FinalStats.Stats[stats.Mana]
			request.Raid.Parties[0].Players[0].BonusStats = &proto.UnitStats{Stats: stats.Stats{stats.Mana: 1500 - mana}.ToProtoArray()}
			request.SimOptions.Debug = true
			result := core.RunSim(request, nil, simsignals.CreateSignals())
			if result.Error != nil {
				t.Fatal(result.Error.Message)
			}
			for _, line := range strings.Split(result.Logs, "\n") {
				if strings.Contains(line, "Casting {SpellID: 9850}") || strings.Contains(line, "Casting {SpellID: 9830}") {
					if !strings.Contains(line, "Casting {SpellID: "+strconv.Itoa(int(tc.want))+"}") {
						t.Errorf("First builder = %s, want spell %d", line, tc.want)
					}
					return
				}
			}
			t.Fatal("No builder was cast")
		})
	}
}

func TestJVMESBerserkImmediatelyPrecedesACatAttack(t *testing.T) {
	for _, race := range []proto.Race{proto.Race_RaceTauren, proto.Race_RaceNightElf, proto.Race_RaceHighOrderSkyborne, proto.Race_RaceWindshaperSkyborne} {
		t.Run(race.String(), func(t *testing.T) {
			rotation := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "jvmes").Rotation
			request := sweepRequest(powershiftTalents, rotation, false, 1)
			request.Encounter.Duration = 210
			request.Raid.Parties[0].Players[0].Race = race
			request.SimOptions.Debug = true
			result := core.RunSim(request, nil, simsignals.CreateSignals())
			if result.Error != nil {
				t.Fatal(result.Error.Message)
			}
			lines := strings.Split(result.Logs, "\n")
			uses := 0
			for i, line := range lines {
				if !strings.Contains(line, "Casting {SpellID: 417141}") {
					continue
				}
				uses++
				at := line[:strings.Index(line, "]")+1]
				found := false
				for _, next := range lines[i+1:] {
					if !strings.Contains(next, "Casting ") || strings.Contains(next, "GCD = 0s") {
						continue
					}
					found = strings.HasPrefix(next, at) && (strings.Contains(next, "Casting {SpellID: 9867}") || strings.Contains(next, "Casting {SpellID: 9904}") || strings.Contains(next, "Casting {SpellID: 9830}"))
					if !found {
						t.Errorf("Berserk at %s was followed by %s", at, next)
					}
					break
				}
				if !found {
					t.Error("Berserk was not immediately followed by a Cat builder")
				}
			}
			if uses < 2 {
				t.Fatal("the fixture did not exercise a later Berserk activation")
			}
		})
	}
}

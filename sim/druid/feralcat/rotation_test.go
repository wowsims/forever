//go:build with_db

package feralcat

import (
	"math"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	googleProto "google.golang.org/protobuf/proto"
)

const powershiftTalents = DefaultTalents
const catTalents = DefaultTalents
const spellProwl int32 = 9913

// Loads the same rotation offered by the Cat UI.
func defaultCatRotation() *proto.APLRotation {
	return core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "adaptive").Rotation
}

func TestMonocatStaysInCatForm(t *testing.T) {
	player := runCat(t, catPowershiftRequest(powershiftTalents, defaultCatRotation(), false))
	for _, c := range []struct {
		name string
		id   int32
	}{{"Shifting Power", 1322605}, {"Shred", 9830}, {"Rip", 9896}, {"Rake", 9904}, {"Faerie Fire", 9907}} {
		if castsOf(player.Actions, c.id) == 0 {
			t.Errorf("%s was never cast", c.name)
		}
	}
	// The only way out of Cat Form is the self-Innervate.
	if forms, innervates := castsOf(player.Actions, 768), castsOf(player.Actions, spellInnervate); forms != innervates {
		t.Errorf("Cat Form casts = %d, Innervate casts = %d", forms, innervates)
	}
}

// Without Predatory Discipline, Ferocious Bite spends the combo points only once a Rip would not tick out.
func TestMonocatBitesOnlyAtTheEnd(t *testing.T) {
	request := catPowershiftRequest(withFeralDigit(powershiftTalents, 12, '0'), defaultCatRotation(), false)
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

// With Predatory Discipline, the extra combo points two points of Blood Frenzy give go to Ferocious Bite while Rip ticks.
func TestMonocatBitesSpareComboPointsWithBloodFrenzy(t *testing.T) {
	endOnly := defaultCatRotation()
	endOnly.PriorityList = slices.DeleteFunc(endOnly.PriorityList, func(item *proto.APLListItem) bool {
		return item.Action.GetCondition().GetVariableRef().GetName() == "Midfight Bite is worthwhile"
	})
	spare := runCat(t, sweepRequest(catTalents, defaultCatRotation(), false, 100))
	without := runCat(t, sweepRequest(catTalents, endOnly, false, 100))
	if got, want := castsOf(spare.Actions, 22829), castsOf(without.Actions, 22829); got <= want {
		t.Errorf("Ferocious Bite casts = %d with the midfight line, %d without it, want more with it", got, want)
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
	player := runCat(t, catPowershiftRequest(withFeralDigit(powershiftTalents, 19, '0'), defaultCatRotation(), false))
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
			rotation := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "adaptive").Rotation
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
			if got, want := castsOf(player.Actions, 768), castsOf(player.Actions, spellInnervate); got != want {
				t.Errorf("Cat Form casts = %d, want %d for Innervate", got, want)
			}
			if castsOf(player.Actions, 9907) != 0 {
				t.Error("Faerie Fire was cast despite the supplied armor debuff")
			}

		})
	}
}

func TestAdaptiveCatRacialSelection(t *testing.T) {
	for _, race := range []proto.Race{proto.Race_RaceNightElf, proto.Race_RaceHighOrderSkyborne, proto.Race_RaceWindshaperSkyborne, proto.Race_RaceTauren} {
		t.Run(race.String(), func(t *testing.T) {
			rotation := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "adaptive").Rotation
			request := sweepRequest(powershiftTalents, rotation, false, 10)
			request.Encounter.Duration = 300
			cat := request.Raid.Parties[0].Players[0]
			cat.Race = race
			cat.Consumables = &proto.ConsumesSpec{ConjuredId: 7676}
			player := runCat(t, request)
			if casts := castsOf(player.Actions, 1259799); (casts > 0) != (race == proto.Race_RaceNightElf) {
				t.Errorf("Elune's Light casts = %d for %s", casts, race)
			}
			for _, spellID := range []int32{1259705, 20580} {
				if casts := castsOf(player.Actions, spellID); casts != 0 {
					t.Errorf("unwanted racial %d casts = %d", spellID, casts)
				}
			}
			if got := castsOf(player.Actions, 9867); got != request.SimOptions.Iterations {
				t.Errorf("Ravage casts = %d, want one Prowl opener per iteration", got)
			}

			if forms, innervates := castsOf(player.Actions, 768), castsOf(player.Actions, spellInnervate); forms != innervates {
				t.Errorf("Cat Form casts = %d, Innervate casts = %d without Sapper", forms, innervates)
			}
		})
	}
}

func TestJVMESDoesNotUseShadowmeld(t *testing.T) {
	rotation := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "adaptive").Rotation
	request := sweepRequest(powershiftTalents, rotation, false, 1)
	request.Encounter.Duration = 300
	request.Raid.Parties[0].Players[0].Race = proto.Race_RaceNightElf
	player := runCat(t, request)
	if castsOf(player.Actions, 20580) != 0 {
		t.Fatal("Shadowmeld was cast")
	}
}

func TestAdaptiveUnprowledOpenerUsesClearcastingForShredOtherwiseRake(t *testing.T) {
	for _, tc := range []struct {
		name         string
		clearcasting bool
	}{
		{"paid Rake", false},
		{"Clearcasting Shred", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rotation := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "adaptive").Rotation
			for _, variable := range rotation.ValueVariables {
				if variable.Name == "Ravage approach without Clearcasting" || variable.Name == "Ravage approach with Clearcasting" {
					variable.Value.GetConst().Val = "-1"
				}
			}
			request := sweepRequest(powershiftTalents, rotation, false, 1)
			request.Raid.Parties[0].Players[0].GetFeralCatDruid().Options.PrepullClearcasting = tc.clearcasting
			request.SimOptions.Debug = true
			result := core.RunSim(request, nil, simsignals.CreateSignals())
			if result.Error != nil {
				t.Fatal(result.Error.Message)
			}
			builderID := "9904"
			if tc.clearcasting {
				builderID = "9830"
			}
			if !strings.Contains(result.Logs, "[0.00] [Cat (#1)] Casting {SpellID: "+builderID+"}") {
				t.Fatal("the opener did not cast the expected builder at pull")
			}
		})
	}
}

func TestAdaptiveCatResourcePotionSelection(t *testing.T) {
	for _, potion := range []int32{0, 13444, 13443} {
		t.Run(strconv.Itoa(int(potion)), func(t *testing.T) {
			rotation := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "adaptive").Rotation
			request := sweepRequest(powershiftTalents, rotation, false, 10)
			request.Encounter.Duration = 300
			request.Raid.Debuffs.JudgementOfWisdom = false
			cat := request.Raid.Parties[0].Players[0]
			cat.Consumables = &proto.ConsumesSpec{OffensivePotId: 13442, ConjuredId: 7676}
			cat.Consumables.ResourcePotId = potion
			player := runCat(t, request)
			if potion == 13444 {
				current := googleProto.Clone(request).(*proto.RaidSimRequest)
				current.SimOptions.Iterations = 1
				current.SimOptions.Debug = true
				previous := googleProto.Clone(current).(*proto.RaidSimRequest)
				for _, variable := range previous.Raid.Parties[0].Players[0].Rotation.ValueVariables {
					if variable.Name == "Mana potion fits and is needed" {
						selected := &proto.APLValue{Value: &proto.APLValue_SelectedPotion{SelectedPotion: &proto.APLValueSelectedPotion{PotionId: &proto.ActionID{RawId: &proto.ActionID_ItemId{ItemId: 13444}}}}}
						variable.Value = and(selected, variable.Value)
					}
				}
				before := core.RunRaidSim(previous)
				after := core.RunRaidSim(current)
				if before.Error != nil || after.Error != nil {
					t.Fatalf("comparison errors: before=%v after=%v", before.Error, after.Error)
				}
				if before.Logs != after.Logs {
					t.Error("Major Mana Potion combat log differs from the item-specific condition")
				}
			}
			used := itemCastsOf(player.Actions, 13444) + itemCastsOf(player.Actions, 13443)
			if potion != 0 {
				used = itemCastsOf(player.Actions, potion)
			}
			if potion != 0 && used == 0 {
				t.Error("The selected mana potion was never used without Judgement of Wisdom")
			}
			if potion == 0 && used != 0 {
				t.Errorf("Mana potion casts = %d without a resource potion selected", used)
			}
			if itemCastsOf(player.Actions, 13442) == 0 {
				t.Error("The offensive potion was never used")
			}
			for _, resource := range player.Resources {
				if potion != 0 && resource.Id.GetItemId() == potion && resource.Type == proto.ResourceType_ResourceTypeMana && math.Abs(resource.Gain-resource.ActualGain) > 1e-8 {
					t.Errorf("Mana potion restored %v of %v mana", resource.ActualGain, resource.Gain)
				}
			}
		})
	}
}

func TestMonocatInstantExplosivesStayInCatForm(t *testing.T) {
	for _, c := range []struct {
		name       string
		itemID     int32
		leavesForm bool
	}{{"Goblin Sapper Charge", 10646, false}, {"EZ-Thro Dark Bomb", 260817, false}} {
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

func TestJVMESTeaRuneSelection(t *testing.T) {
	for _, selected := range []int32{0, 7676, 12662} {
		t.Run(strconv.Itoa(int(selected)), func(t *testing.T) {
			rotation := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "adaptive").Rotation
			req := sweepRequest(powershiftTalents, rotation, false, 1)
			req.Encounter.Duration = 300
			p := req.Raid.Parties[0].Players[0]
			p.Consumables = &proto.ConsumesSpec{ConjuredId: selected, ConjuredItems: []int32{7676, 12662}}
			player := runCat(t, req)
			tea, rune := itemCastsOf(player.Actions, 7676), itemCastsOf(player.Actions, 12662)
			t.Logf("Tea=%d Rune=%d", tea, rune)
			if (selected == 0 && (tea != 0 || rune != 0)) || (selected == 7676 && (tea == 0 || rune != 0)) || (selected == 12662 && (tea == 0 || rune == 0)) {
				t.Fatalf("selection %d: Tea=%d Rune=%d", selected, tea, rune)
			}
		})
	}
}

func TestJVMESMaintainsFaerieFire(t *testing.T) {
	for _, supplied := range []bool{false, true} {
		t.Run(strconv.FormatBool(supplied), func(t *testing.T) {
			rotation := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "adaptive").Rotation
			request := sweepRequest(powershiftTalents, rotation, false, 1)
			request.Raid.Debuffs.FaerieFire = supplied
			request.Raid.Debuffs.CurseOfRecklessness = false
			request.Encounter.Duration = 90
			request.SimOptions.Debug = true
			result := core.RunSim(request, nil, simsignals.CreateSignals())
			if result.Error != nil {
				t.Fatal(result.Error.Message)
			}
			casts := castsOf(result.RaidMetrics.Parties[0].Players[0].Actions, 9907)
			if supplied && casts != 0 {
				t.Fatalf("Faerie Fire casts = %d with raid coverage", casts)
			}
			if !supplied && casts < 2 {
				t.Fatalf("Faerie Fire casts = %d, want application and refresh", casts)
			}
			if !supplied {
				for _, line := range strings.Split(result.Logs, "\n") {
					if strings.HasPrefix(line, "[-") || !strings.Contains(line, "Casting {") {
						continue
					}
					if !strings.Contains(line, "SpellID: 9907") {
						t.Errorf("first combat cast: %s", line)
					}
					break
				}
			}
		})
	}
}

func TestAdaptiveMcpReturnsToMainWeapon(t *testing.T) {
	for _, selected := range []bool{false, true} {
		req := sweepRequest(powershiftTalents, core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "adaptive").Rotation, false, 1)
		req.Encounter.Duration = 60
		req.Encounter.DurationVariation = 0
		p := req.Raid.Parties[0].Players[0]
		p.DistanceFromTarget = 25
		mainID := p.Equipment.Items[proto.ItemSlot_ItemSlotMainHand].Id
		p.EnableItemSwap = selected
		if selected {
			items := make([]*proto.ItemSpec, 17)
			for i := range items {
				items[i] = &proto.ItemSpec{}
			}
			items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 9449, Enchant: 1896}
			p.ItemSwap = &proto.ItemSwap{Items: items}
			p.Consumables = googleProto.Clone(p.Consumables).(*proto.ConsumesSpec)
			p.Consumables.SwapMhImbueId = 12643
		}
		sim := core.NewSim(req, simsignals.CreateSignals())
		for iteration := 0; iteration < 2; iteration++ {
			sim.Reset()
			sim.PrePull()
			for sim.CurrentTime < 0 {
				if sim.Step() {
					t.Fatal("simulation ended before pull")
				}
			}
			c := sim.Raid.Parties[0].Players[0].GetCharacter()
			mcp := c.GetSpell(core.ActionID{ItemID: 9449})
			if selected && c.Equipment[proto.ItemSlot_ItemSlotMainHand].ID != 9449 {
				t.Fatal("MCP was not equipped before the pull")
			}
			for !sim.Step() {
			}
			if c.Equipment[proto.ItemSlot_ItemSlotMainHand].ID != mainID {
				t.Fatal("main weapon was not restored")
			}
			if selected {
				if mcp == nil {
					t.Fatal("MCP was not registered")
				}
				casts := int32(0)
				for _, metrics := range mcp.SpellMetrics {
					casts += metrics.Casts
				}
				if casts != 1 {
					t.Fatalf("MCP casts = %d, want one opener", casts)
				}
			}
			sim.Cleanup()
		}
	}
}

func TestAdaptiveClawsWhenShredCannotLand(t *testing.T) {
	for _, cannotShred := range []bool{false, true} {
		t.Run(strconv.FormatBool(cannotShred), func(t *testing.T) {
			rotation := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "adaptive").Rotation
			req := sweepRequest("050022-35200032123011125011-054", rotation, cannotShred, 10)
			player := runCat(t, req)
			claws, shreds := castsOf(player.Actions, 9850), castsOf(player.Actions, 9830)
			if cannotShred {
				if claws == 0 || shreds != 0 || castsOf(player.Actions, 9896) == 0 {
					t.Fatalf("Claw=%d Shred=%d Rip=%d when Shred cannot land", claws, shreds, castsOf(player.Actions, 9896))
				}
			} else if claws != 0 || shreds == 0 {
				t.Fatalf("Claw=%d Shred=%d when Shred can land", claws, shreds)
			}
		})
	}
}

func TestAdaptiveBleedsRespectExpiryReaction(t *testing.T) {
	for _, id := range []int32{9904, 9896} {
		t.Run(strconv.Itoa(int(id)), func(t *testing.T) {
			rotation := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "adaptive").Rotation
			rotation.PrepullActions = nil
			for _, item := range rotation.PriorityList {
				if item.Action.GetCastSpell().GetSpellId().GetSpellId() == id {
					rotation.PriorityList = []*proto.APLListItem{item}
					break
				}
			}
			req := sweepRequest(powershiftTalents, rotation, false, 1)
			sim := core.NewSim(req, simsignals.CreateSignals())
			sim.Reset()
			cat := sim.Raid.Parties[0].Players[0].(*FeralDruid)
			cat.AutoAttacks.CancelAutoSwing(sim)
			spell := cat.GetSpell(core.ActionID{SpellID: id})
			sim.CurrentTime = time.Second
			spell.Dot(cat.CurrentTarget).Apply(sim)
			sim.CurrentTime = 20 * time.Second
			cat.AddComboPoints(sim, 5, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
			fillResource(sim, cat)
			cat.SetGCDTimer(sim, 0)
			before := spell.SpellMetrics[cat.CurrentTarget.UnitIndex].Casts
			spell.Dot(cat.CurrentTarget).Deactivate(sim)
			for _, at := range []time.Duration{20 * time.Second, 20*time.Second + cat.ReactionTime - time.Nanosecond} {
				sim.CurrentTime = at
				cat.Rotation.DoNextAction(sim)
				if got := spell.SpellMetrics[cat.CurrentTarget.UnitIndex].Casts - before; got != 0 {
					t.Fatalf("bleed refreshed before the expiry reaction elapsed: %d casts", got)
				}
			}
			sim.CurrentTime = 20*time.Second + cat.ReactionTime
			cat.Rotation.DoNextAction(sim)
			if got := spell.SpellMetrics[cat.CurrentTarget.UnitIndex].Casts - before; got != 1 {
				t.Fatalf("bleed casts after the expiry reaction = %d, want one", got)
			}
		})
	}
}

func TestAdaptiveChoosesTheApproachBeforeMoving(t *testing.T) {
	for _, tc := range []struct {
		name                                          string
		distance                                      float64
		clearcasting, externalFF, cannotShred, ravage bool
	}{
		{"paid Ravage nearby", 10, false, true, false, true},
		{"paid approach too long", 18, false, true, false, false},
		{"Clearcasting Ravage", 25, true, true, false, true},
		{"Clearcasting approach too long", 50, true, true, false, false},
		{"supply Faerie Fire first", 25, true, false, false, false},
		{"target prevents Ravage", 0, true, true, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rotation := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "adaptive").Rotation
			req := sweepRequest(powershiftTalents, rotation, tc.cannotShred, 1)
			req.Encounter.Duration = 15
			req.Encounter.DurationVariation = 0
			req.Raid.Debuffs = googleProto.Clone(req.Raid.Debuffs).(*proto.Debuffs)
			req.Raid.Debuffs.FaerieFire = tc.externalFF
			pl := req.Raid.Parties[0].Players[0]
			pl.DistanceFromTarget = tc.distance
			pl.GetFeralCatDruid().Options.PrepullClearcasting = tc.clearcasting
			m := runCat(t, req)
			got := castsOf(m.Actions, 9867)
			if (got > 0) != tc.ravage {
				t.Fatalf("Ravage casts = %d, want opener %v", got, tc.ravage)
			}
			if castsOf(m.Actions, 9913) != 1 {
				t.Fatal("Prowl was not prepared before the pull")
			}
			if castsOf(m.Actions, 9830)+castsOf(m.Actions, 9850)+castsOf(m.Actions, 9904) == 0 {
				t.Fatal("rotation did not continue after the approach")
			}
		})
	}
}

func TestAdaptiveSkyborneApproachUsesClearcasting(t *testing.T) {
	for _, race := range []proto.Race{proto.Race_RaceHighOrderSkyborne, proto.Race_RaceWindshaperSkyborne} {
		for _, distance := range []float64{10, 25, 35} {
			for _, clearcasting := range []bool{false, true} {
				rotation := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "adaptive").Rotation
				req := sweepRequest(powershiftTalents, rotation, false, 1)
				req.Encounter.Duration = 15
				req.Encounter.DurationVariation = 0
				req.Raid.Debuffs = googleProto.Clone(req.Raid.Debuffs).(*proto.Debuffs)
				req.Raid.Debuffs.FaerieFire = true
				pl := req.Raid.Parties[0].Players[0]
				pl.Race = race
				pl.DistanceFromTarget = distance
				pl.GetFeralCatDruid().Options.PrepullClearcasting = clearcasting
				m := runCat(t, req)
				got := castsOf(m.Actions, 9867)
				want := clearcasting && distance <= 25
				if (got > 0) != want {
					t.Errorf("%s distance=%g Clearcasting=%v: Ravage casts=%d, want opener %v", race, distance, clearcasting, got, want)
				}
			}
		}
	}
}

const (
	spellCatForm int32 = 768
	spellShred   int32 = 9830
	spellClaw    int32 = 9850
	spellRip     int32 = 9896
	spellRake    int32 = 9904
	spellFF      int32 = 9907
	spellSP      int32 = 1322605
	spellElune   int32 = 1259799
	spellClear   int32 = 16870
	itemSapper   int32 = 10646
	itemTea      int32 = 7676

	spellBite       int32 = 22829
	spellDireBear   int32 = 9634
	spellPrimalBite int32 = 1238070
	spellEnrage     int32 = 5229
	spellMaul       int32 = 9881
	spellInnervate  int32 = 29166
	// Read Ley Line, the High Order Skyborne's racial: Energized, twice the mana regeneration for 15s.
	spellEnergized int32 = 1259705
)

// Fixed Cat consumables and professions for APL validation.
func catValidationRequest(rotation *proto.APLRotation) *proto.ComputeStatsRequest {
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

// Availability guards can report omitted optional spells; malformed conditions must still fail.
func TestAdaptiveCatAvailabilityWarnings(t *testing.T) {
	for _, consumes := range []bool{true, false} {
		preset := defaultCatRotation()
		req := catValidationRequest(preset)
		if !consumes {
			req.Raid.Parties[0].Players[0].Profession1 = proto.Profession_ProfessionUnknown
			req.Raid.Parties[0].Players[0].Consumables = &proto.ConsumesSpec{}
		}
		expected := map[string]bool{
			"Cat (#1) does not know spell {SpellID: 1259705}":   true,
			"No aura found on Cat (#1) for: {SpellID: 1270842}": true,
			"Cat (#1) does not know spell {SpellID: 1259686}":   true,
			"Cat (#1) does not know spell {ItemID: 9449}":       true,
			"Cat (#1) does not know spell {ItemID: 12662}":      true,
			"Cat (#1) picks no potion for {OtherID: 26}":        true,
			"No swap set configured in Settings.":               true,
		}
		if !consumes {
			expected["Cat (#1) does not know spell {ItemID: 7676}"] = true
			expected["Cat (#1) picks no potion for {OtherID: 24}"] = true
		}
		for line, messages := range withoutIdleAutocast(preset, aplWarnings(t, preset, req)) {
			for _, message := range messages {
				if !expected[message] {
					t.Errorf("consumes=%t line %s: %s", consumes, line, message)
				}
			}
		}
		result := core.RunSim(&proto.RaidSimRequest{Raid: req.Raid, Encounter: req.Encounter, SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 20, IsTest: true}}, nil, simsignals.CreateSignals())
		if result.Error != nil {
			t.Fatal(result.Error.Message)
		}
		metrics := result.RaidMetrics.Parties[0].Players[0]
		for _, id := range []int32{spellShred, spellRake, spellRip, spellSP} {
			if castsOf(metrics.Actions, id) == 0 {
				t.Errorf("consumes=%t spell %d was never cast", consumes, id)
			}
		}
	}
}

func spellID(id int32) *proto.ActionID {
	return &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}}
}

func comboPoints() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_CurrentComboPoints{CurrentComboPoints: &proto.APLValueCurrentComboPoints{}}}
}

func TestAdaptiveInnervateWaitsForEnergized(t *testing.T) {
	for _, race := range []proto.Race{proto.Race_RaceTauren, proto.Race_RaceHighOrderSkyborne} {
		for _, active := range []bool{false, true} {
			if race == proto.Race_RaceTauren && active {
				continue
			}
			rotation := defaultCatRotation()
			rotation.PrepullActions = nil
			for _, item := range rotation.PriorityList {
				if item.Action.GetCastSpell().GetSpellId().GetSpellId() == spellInnervate {
					rotation.PriorityList = []*proto.APLListItem{item}
					break
				}
			}
			req := sweepRequest(powershiftTalents, rotation, false, 1)
			req.Raid.Parties[0].Players[0].Race = race
			sim := core.NewSim(req, simsignals.CreateSignals())
			sim.Reset()
			cat := sim.Raid.Parties[0].Players[0].(*FeralDruid)
			cat.AutoAttacks.CancelAutoSwing(sim)
			sim.CurrentTime = time.Second
			cat.SetGCDTimer(sim, 0)
			cat.SpendEnergy(sim, cat.CurrentEnergy(), cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
			cat.SpendMana(sim, cat.CurrentMana()-1000, cat.NewManaMetrics(core.ActionID{SpellID: 1}))
			aura := cat.GetAuraByID(core.ActionID{SpellID: 1270842})
			if active {
				aura.Activate(sim)
			}
			cat.Rotation.DoNextAction(sim)
			spell := cat.GetSpell(core.ActionID{SpellID: spellInnervate})
			casts := int32(0)
			for _, metrics := range spell.SpellMetrics {
				casts += metrics.Casts
			}
			want := int32(1)
			if active {
				want = 0
			}
			if casts != want {
				t.Errorf("race=%s Energized=%v: Innervate casts=%d, want %d", race, active, casts, want)
			}
		}
	}
}

func TestAdaptiveBitePoolsByComboPointsAndClearcasting(t *testing.T) {
	for _, tc := range []struct {
		name      string
		cp        int32
		energy    float64
		clearcast bool
		ripLeft   time.Duration
		want      int32
	}{
		{"four points below pool", 4, 84, false, 3 * time.Second, 0},
		{"four points funded", 4, 85, false, 3 * time.Second, 1},
		{"five points below pool", 5, 69, false, 3 * time.Second, 0},
		{"five points funded", 5, 70, false, 3 * time.Second, 1},
		{"four points clearcast below pool", 4, 59, true, 3 * time.Second, 0},
		{"four points clearcast funded", 4, 60, true, 3 * time.Second, 1},
		{"five points clearcast below pool", 5, 49, true, 3 * time.Second, 0},
		{"five points clearcast funded", 5, 50, true, 3 * time.Second, 1},
		{"Rip window too short", 5, 100, false, 1500 * time.Millisecond, 0},
		{"Rip window funded", 5, 100, false, 2 * time.Second, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rotation := defaultCatRotation()
			rotation.PrepullActions = nil
			rotation.PriorityList = slices.DeleteFunc(rotation.PriorityList, func(item *proto.APLListItem) bool {
				return item.Action.GetCondition().GetVariableRef().GetName() != "Midfight Bite is worthwhile"
			})
			if len(rotation.PriorityList) != 1 {
				t.Fatal("expected one midfight Bite action")
			}
			sim := core.NewSim(sweepRequest(catTalents, rotation, false, 1), simsignals.CreateSignals())
			sim.Reset()
			cat := sim.Raid.Parties[0].Players[0].(*FeralDruid)
			cat.AutoAttacks.CancelAutoSwing(sim)
			sim.CurrentTime = 10 * time.Second
			fillResource(sim, cat)
			cat.SpendEnergy(sim, cat.CurrentEnergy()-tc.energy, cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
			cat.AddComboPoints(sim, tc.cp, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
			dot := cat.GetSpell(core.ActionID{SpellID: spellRip}).Dot(cat.CurrentTarget)
			dot.Apply(sim)
			dot.UpdateExpires(sim.CurrentTime + tc.ripLeft)
			if tc.clearcast {
				cat.GetAuraByID(core.ActionID{SpellID: spellClear}).Activate(sim)
			}
			cat.SetGCDTimer(sim, sim.CurrentTime)
			cat.Rotation.DoNextAction(sim)
			bite := cat.GetSpell(core.ActionID{SpellID: spellBite})
			if got := bite.SpellMetrics[cat.CurrentTarget.UnitIndex].Casts; got != tc.want {
				t.Fatalf("Bite casts = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestAdaptiveRipBuildLeavesTimeForNextGlobal(t *testing.T) {
	for _, tc := range []struct {
		name string
		left time.Duration
		want int32
	}{
		{"too late for Rip", 7500 * time.Millisecond, 0},
		{"Rip now before the window closes", 8500 * time.Millisecond, 1},
		{"time to build then Rip", 9500 * time.Millisecond, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rotation := defaultCatRotation()
			rotation.PrepullActions = nil
			rotation.PriorityList = slices.DeleteFunc(rotation.PriorityList, func(item *proto.APLListItem) bool {
				return item.Action.GetCastSpell().GetSpellId().GetSpellId() != spellRip
			})
			if len(rotation.PriorityList) != 1 {
				t.Fatal("expected one Rip action")
			}
			sim := core.NewSim(sweepRequest(catTalents, rotation, false, 1), simsignals.CreateSignals())
			sim.Reset()
			cat := sim.Raid.Parties[0].Players[0].(*FeralDruid)
			cat.AutoAttacks.CancelAutoSwing(sim)
			sim.CurrentTime = sim.Duration - tc.left
			fillResource(sim, cat)
			cat.AddComboPoints(sim, 4, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
			cat.SetGCDTimer(sim, sim.CurrentTime)
			cat.Rotation.DoNextAction(sim)
			rip := cat.GetSpell(core.ActionID{SpellID: spellRip})
			if got := rip.SpellMetrics[cat.CurrentTarget.UnitIndex].Casts; got != tc.want {
				t.Fatalf("Rip casts = %d with %s left, want %d", got, tc.left, tc.want)
			}
		})
	}
}

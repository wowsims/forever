//go:build with_db

package feralbear

import (
	"fmt"
	"github.com/wowsims/forever/sim/core/stats"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	googleProto "google.golang.org/protobuf/proto"
)

const bearTalents = DefaultTalents

func TestBearShippedAplRuns(t *testing.T) {
	req := bearNakedRequest()
	req.Raid.Parties[0].Players[0].GetFeralBearDruid().Options.StartingRage = 25
	result := core.RunSim(req, nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatalf("sim error: %s", result.Error.Message)
	}
	player := result.RaidMetrics.Parties[0].Players[0]
	for _, id := range []int32{1238070, 9881} {
		if bearCastsOf(player.Actions, id) == 0 {
			t.Fatalf("shipped bear APL never casts %d", id)
		}
	}
}

func bearNakedRequest() *proto.RaidSimRequest {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 1613
	return &proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1, IsTest: true},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Bear", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
				TalentsString: bearTalents,
				Equipment:     &proto.EquipmentSpec{Items: items},
				Consumables:   &proto.ConsumesSpec{},
				Buffs:         &proto.IndividualBuffs{},
				Spec: &proto.Player_FeralBearDruid{FeralBearDruid: &proto.FeralBearDruid{
					Options: &proto.FeralBearDruid_Options{},
				}},
				Rotation: core.GetAplRotation("../../../ui/specs/druid/feralbear/apls", "adaptive").Rotation,
			}},
			Buffs: &proto.PartyBuffs{},
		}},
			Tanks: []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}},
		},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 60},
	}
}

func bearCastsOf(actions []*proto.ActionMetrics, spellID int32) int32 {
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

func TestBearFormRecastGrantsFuror(t *testing.T) {
	for _, rank := range []int{0, 5} {
		for _, at := range []time.Duration{-10 * time.Second, 0, 2 * time.Second} {
			for _, id := range []int32{5487, spellDireBear} {
				t.Run(fmt.Sprintf("rank%d/time%s/form%d", rank, at, id), func(t *testing.T) {
					req := bearNakedRequest()
					player := req.Raid.Parties[0].Players[0]
					player.Rotation = &proto.APLRotation{}
					trees := strings.Split(player.TalentsString, "-")
					feral := []byte(trees[1])
					feral[1] = byte('0' + rank)
					trees[1] = string(feral)
					player.TalentsString = strings.Join(trees, "-")
					sim := core.NewSim(req, simsignals.CreateSignals())
					sim.Reset()
					sim.CurrentTime = at
					bear := sim.Raid.Parties[0].Players[0].(*GuardianDruid)
					form := bear.GetSpell(core.ActionID{SpellID: id})
					wantRage := core.Ternary(rank == 0, 0.0, 10.0)
					for cast := 0; cast < 2; cast++ {
						bear.SetGCDTimer(sim, at)
						bear.AddRage(sim, 40, bear.NewRageMetrics(core.ActionID{SpellID: 1}))
						mana := bear.CurrentMana()
						armor, health := bear.GetStat(stats.Armor), bear.MaxHealth()
						if !form.Cast(sim, bear.CurrentTarget) {
							t.Fatal("could not cast Bear form")
						}
						if cast == 1 && (bear.GetStat(stats.Armor) != armor || bear.MaxHealth() != health) {
							t.Fatal("recasting the form changed its passive bonuses")
						}
						if bear.CurrentRage() != wantRage || bear.CurrentMana() != mana-form.CurCast.Cost {
							t.Fatalf("cast %d: rage %g, want %g; mana spent %g", cast, bear.CurrentRage(), wantRage, mana-bear.CurrentMana())
						}
					}
				})
			}
		}
	}
}

func TestBearStartingRageDoesNotTriggerFuror(t *testing.T) {
	for _, rage := range []float64{0, 25} {
		req := bearNakedRequest()
		player := req.Raid.Parties[0].Players[0]
		player.GetFeralBearDruid().Options.StartingRage = rage
		player.Rotation = selectedAdaptiveBearAPL()
		sim := core.NewSim(req, simsignals.CreateSignals())
		for iteration := 0; iteration < 2; iteration++ {
			sim.Reset()
			bear := sim.Raid.Parties[0].Players[0].(*GuardianDruid)
			if bear.CurrentRage() != rage {
				t.Fatalf("iteration %d: starting rage = %g, want %g", iteration, bear.CurrentRage(), rage)
			}
			if bear.PrimalBite.SpellMetrics[bear.CurrentTarget.UnitIndex].Casts != 0 {
				t.Fatal("starting form executed the rotation during reset")
			}
			sim.Cleanup()
		}
	}
}

// Five points of Furor make its rage certain.
func withFurorFive(talents string) string {
	trees := strings.Split(talents, "-")
	feral := []byte(trees[1])
	feral[1] = '5'
	trees[1] = string(feral)
	return strings.Join(trees, "-")
}

func TestBearPrepullEnrage(t *testing.T) {
	rotation := selectedAdaptiveBearAPL()
	rotation.PriorityList = nil
	req := bearNakedRequest()
	req.Raid.Parties[0].Players[0].Rotation = rotation
	req.Raid.Parties[0].Players[0].TalentsString = withFurorFive(bearTalents)
	req.Raid.Tanks = nil
	sim := core.NewSim(req, simsignals.CreateSignals())
	for iteration := 0; iteration < 2; iteration++ {
		sim.Reset()
		bear := sim.Raid.Parties[0].Players[0].(*GuardianDruid)
		bear.AutoAttacks.CancelAutoSwing(sim)
		observed := false
		sim.AddPendingAction(&core.PendingAction{NextActionAt: 0, Priority: core.ActionPriorityLow, OnAction: func(sim *core.Simulation) {
			observed = true
			t.Logf("iteration %d: rage %g, Enrage active %v, cooldown %s", iteration, bear.CurrentRage(), bear.EnrageAura.IsActive(), bear.Enrage.CD.TimeToReady(sim))
			if bear.CurrentRage() != 40 || bear.EnrageAura.IsActive() || bear.Enrage.CD.TimeToReady(sim) != 50*time.Second || !bear.DireBearFormAura.IsActive() {
				t.Fatal("prepull powershift and Enrage did not finish before the pull")
			}
		}})
		sim.PrePull()
		for !observed {
			if sim.Step() {
				t.Fatal("pull observation did not run")
			}
		}
		sim.Cleanup()
	}
}

func TestBearPrepullEnrageWithFullRotation(t *testing.T) {
	req := bearRequest(selectedAdaptiveBearAPL(), 2)
	req.SimOptions.Debug = true
	sim := core.NewSim(req, simsignals.CreateSignals())
	sim.Reset()
	bear := sim.Raid.Parties[0].Players[0].(*GuardianDruid)
	if bear.PrimalBite.SpellMetrics[bear.CurrentTarget.UnitIndex].Casts != 0 {
		t.Fatal("starting Furor rage ran the rotation during reset")
	}
	result := core.RunSim(req, nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}
	for _, id := range []int32{spellDireBear, spellEnrage} {
		count := 0
		for _, line := range strings.Split(result.Logs, "\n") {
			if strings.Contains(line, fmt.Sprintf("Casting {SpellID: %d}", id)) {
				if id == spellEnrage && !strings.HasPrefix(line, "[-10.00]") {
					continue
				}
				count++
				if !strings.HasPrefix(line, "[-10.00]") {
					t.Fatalf("prepull action cast at the wrong time: %s", line)
				}
			}
		}
		if count != 2 {
			t.Fatalf("prepull spell %d cast %d times over two iterations", id, count)
		}
	}
}

func TestAdaptiveBearHoldsEnrageForRageShortage(t *testing.T) {
	for _, check := range []struct {
		name                string
		biteReady, gcdReady bool
		rage                float64
		want                bool
	}{
		{"bite-starved", true, true, 15, true},
		{"lacerate-starved", false, true, 5, true},
		{"bite-funded", true, true, 20, false},
		{"lacerate-funded", false, true, 15, false},
		{"global-busy", true, false, 0, false},
	} {
		t.Run(check.name, func(t *testing.T) {
			rot := selectedAdaptiveBearAPL()
			for _, item := range rot.PriorityList {
				if item.Action.GetCastSpell().GetSpellId().GetSpellId() == spellEnrage {
					rot.PriorityList = []*proto.APLListItem{item}
					break
				}
			}
			rot.PrepullActions = nil
			req := bearRequest(rot, 1)
			req.Raid.Parties[0].Players[0].Consumables = &proto.ConsumesSpec{}
			sim := core.NewSim(req, simsignals.CreateSignals())
			sim.Reset()
			sim.CurrentTime = time.Second
			bear := sim.Raid.Parties[0].Players[0].(*GuardianDruid)
			bear.AutoAttacks.CancelAutoSwing(sim)
			bear.SetGCDTimer(sim, 0)
			if !check.gcdReady {
				bear.SetGCDTimer(sim, 2*time.Second)
			}
			if !check.biteReady {
				bear.PrimalBite.CD.Set(4 * time.Second)
			}
			bear.SpendRage(sim, bear.CurrentRage(), bear.NewRageMetrics(core.ActionID{SpellID: 1}))
			bear.AddRage(sim, check.rage, bear.NewRageMetrics(core.ActionID{SpellID: 1}))
			bear.Rotation.DoNextAction(sim)
			if bear.EnrageAura.IsActive() != check.want {
				t.Fatalf("Enrage active=%v, wanted %v", bear.EnrageAura.IsActive(), check.want)
			}
		})
	}
}

func TestAdaptiveBearPotionSelection(t *testing.T) {
	for _, check := range []struct {
		name                 string
		stoneshield, ragePot bool
		rage                 float64
		want                 int32
	}{
		{"stoneshield-at-full-health", true, false, 100, itemStoneshield},
		{"stoneshield-before-rage", true, true, 0, itemStoneshield},
		{"rage-below-threshold", false, true, 19, 13442},
		{"hold-rage-at-threshold", false, true, 20, 0},
		{"hold-rage-at-cap", false, true, 100, 0},
		{"no-potion", false, false, 0, 0},
	} {
		t.Run(check.name, func(t *testing.T) {
			rot := selectedAdaptiveBearAPL()
			rot.PrepullActions = nil
			rot.PriorityList = slices.DeleteFunc(rot.PriorityList, func(row *proto.APLListItem) bool {
				id := row.Action.GetCastSpell().GetSpellId().GetOtherId()
				return id != proto.OtherAction_OtherActionDefensivePotion && id != proto.OtherAction_OtherActionOffensivePotion && row.Action.GetAutocastOtherCooldowns() == nil
			})
			req := bearRequest(rot, 1)
			player := req.Raid.Parties[0].Players[0]
			// The browser registers relevant potions independently of the selected slots.
			player.Consumables = &proto.ConsumesSpec{Potions: []int32{itemStoneshield, 13442}}
			if check.stoneshield {
				player.Consumables.DefensivePotId = itemStoneshield
			}
			if check.ragePot {
				player.Consumables.OffensivePotId = 13442
			}
			sim := core.NewSim(req, simsignals.CreateSignals())
			sim.Reset()
			sim.CurrentTime = 0
			bear := sim.Raid.Parties[0].Players[0].(*GuardianDruid)
			bear.AutoAttacks.CancelAutoSwing(sim)
			bear.SetGCDTimer(sim, 0)
			bear.SpendRage(sim, bear.CurrentRage(), bear.NewRageMetrics(core.ActionID{SpellID: 1}))
			bear.AddRage(sim, check.rage, bear.NewRageMetrics(core.ActionID{SpellID: 1}))
			bear.Rotation.DoNextAction(sim)
			for _, id := range []int32{itemStoneshield, 13442} {
				spell := bear.GetSpell(core.ActionID{ItemID: id})
				if spell != nil && spell.RelatedSelfBuff.IsActive() != (check.want == id) {
					t.Fatalf("item %d active=%v, wanted item %d", id, spell.RelatedSelfBuff.IsActive(), check.want)
				}
			}
			if check.want != 0 && !bear.GetSpell(core.ActionID{ItemID: check.want}).RelatedSelfBuff.IsActive() {
				t.Fatal("selected potion did not activate")
			}
		})
	}
}

func selectedAdaptiveBearAPL() *proto.APLRotation {
	return core.GetAplRotation("../../../ui/specs/druid/feralbear/apls", "adaptive").Rotation
}

func TestAdaptiveMcpReturnsToMainWeapon(t *testing.T) {
	for _, selected := range []bool{false, true} {
		req := bearRequest(selectedAdaptiveBearAPL(), 1)
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

func TestAdaptiveBearPrepullPreservesStartingRage(t *testing.T) {
	for _, rage := range []float64{0, 9, 10, 25, 100} {
		t.Run(fmt.Sprintf("rage=%g", rage), func(t *testing.T) {
			rotation := selectedAdaptiveBearAPL()
			rotation.PrepullActions = rotation.PrepullActions[:1]
			rotation.PriorityList = nil
			req := bearNakedRequest()
			req.Raid.Tanks = nil
			player := req.Raid.Parties[0].Players[0]
			player.GetFeralBearDruid().Options.StartingRage = rage
			player.Rotation = rotation
			sim := core.NewSim(req, simsignals.CreateSignals())
			for iteration := 0; iteration < 2; iteration++ {
				sim.Reset()
				bear := sim.Raid.Parties[0].Players[0].(*GuardianDruid)
				bear.AutoAttacks.CancelAutoSwing(sim)
				observed := false
				sim.AddPendingAction(&core.PendingAction{NextActionAt: -9 * time.Second, OnAction: func(sim *core.Simulation) {
					observed = true
					wantCasts := int32(0)
					if rage < 10 {
						wantCasts = 1
					}
					gotCasts := bear.DireBearForm.SpellMetrics[bear.CurrentTarget.UnitIndex].Casts
					if gotCasts != wantCasts {
						t.Fatalf("prepull shifts = %d, want %d", gotCasts, wantCasts)
					}
					if rage >= 10 && bear.CurrentRage() != rage {
						t.Fatalf("starting rage %g became %g", rage, bear.CurrentRage())
					}
					t.Logf("iteration %d: initial rage %g, observed rage %g, shifts %d", iteration, rage, bear.CurrentRage(), gotCasts)
				}})
				sim.PrePull()
				for !observed {
					if sim.Step() {
						t.Fatal("simulation ended before prepull observation")
					}
				}
				sim.Cleanup()
			}
		})
	}
}

func TestAdaptiveBearPrepullDoesNotShiftWithoutFuror(t *testing.T) {
	rotation := selectedAdaptiveBearAPL()
	rotation.PrepullActions = rotation.PrepullActions[:1]
	rotation.PriorityList = nil
	req := bearNakedRequest()
	req.Raid.Tanks = nil
	player := req.Raid.Parties[0].Players[0]
	player.TalentsString = "050022-20200302123102125010-055"
	player.Rotation = rotation
	sim := core.NewSim(req, simsignals.CreateSignals())
	sim.Reset()
	bear := sim.Raid.Parties[0].Players[0].(*GuardianDruid)
	if bear.Talents.Furor != 0 {
		t.Fatal("fixture has Furor talented")
	}
	bear.AutoAttacks.CancelAutoSwing(sim)
	observed := false
	sim.AddPendingAction(&core.PendingAction{NextActionAt: -9 * time.Second, OnAction: func(sim *core.Simulation) {
		observed = true
		if casts := bear.DireBearForm.SpellMetrics[bear.CurrentTarget.UnitIndex].Casts; casts != 0 {
			t.Fatalf("prepull shifts without Furor = %d", casts)
		}
	}})
	sim.PrePull()
	for !observed {
		if sim.Step() {
			t.Fatal("simulation ended before prepull observation")
		}
	}
	sim.Cleanup()
}

const (
	spellDireBear   int32 = 9634
	spellPrimalBite int32 = 1238070
	spellMaul       int32 = 9881
	spellLacerate   int32 = 1235827
	spellSwipe      int32 = 9908
	spellEnrage     int32 = 5229
	spellDemoRoar   int32 = 9898
	spellFF         int32 = 9907
	spellFrenzied   int32 = 22842
	spellBarkskin   int32 = 22812
	itemSapper      int32 = 10646
	itemStoneshield int32 = 13455
	itemHealthstone int32 = 9421
)

// Fixed consumables for Bear rotation and mechanics tests.
var bearTestConsumables = &proto.ConsumesSpec{
	DefensivePotId: itemStoneshield,
	ElixirIds:      []int32{13452, 9088},
	FoodId:         21023,
	MhImbueId:      12643,
	ConjuredId:     itemHealthstone,
	ExtraIds:       []int32{itemSapper, 10310, 10305},
}

// Level 60 tank fixture with fixed raid buffs, consumables, healing and incoming damage.
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
					Consumables:     bearTestConsumables,
					Cooldowns:       &proto.Cooldowns{HpPercentForDefensives: 0.4},
					Buffs: &proto.IndividualBuffs{
						GreaterBlessingOfKings: true,
						GreaterBlessingOfMight: 1,
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
				BattleShout:              1,
				ArcaneBrilliance:         1,
				GiftOfTheWild:            1,
				PrayerOfFortitude:        1,
				PrayerOfShadowProtection: 1,
				PrayerOfSpirit:           true,
				WindfuryTotem:            true,
				StrengthOfEarthTotem:     1,
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

// The tank fixture's damage before armor.
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

// One target keeps the Lacerate stacks; more than one takes Swipe in its place once one bleeds.
func TestAdaptiveBearSwipesOnlyWithMoreThanOneTarget(t *testing.T) {
	for _, check := range []struct {
		targets int
		swipes  bool
	}{{1, false}, {2, true}} {
		req := bearRequest(selectedAdaptiveBearAPL(), 20)
		req.Encounter.Targets = make([]*proto.Target, check.targets)
		for i := range req.Encounter.Targets {
			req.Encounter.Targets[i] = core.NewDefaultTarget()
		}
		result := core.RunSim(req, nil, simsignals.CreateSignals())
		if result.Error != nil {
			t.Fatal(result.Error.Message)
		}
		actions := result.RaidMetrics.Parties[0].Players[0].Actions
		if got := bearCastsOf(actions, spellSwipe) > 0; got != check.swipes {
			t.Errorf("%d targets: Swipe cast = %t, want %t", check.targets, got, check.swipes)
		}
		if bearCastsOf(actions, spellLacerate) == 0 {
			t.Errorf("%d targets: Lacerate was never cast, so nothing bleeds", check.targets)
		}
	}
}

// Omitted utility items and debuffs produce availability warnings, not malformed rotation branches.
func TestAdaptiveBearAvailabilityWarnings(t *testing.T) {
	for _, consumes := range []bool{true, false} {
		req := bearRequest(selectedAdaptiveBearAPL(), 20)
		if !consumes {
			req.Raid.Parties[0].Players[0].Consumables = &proto.ConsumesSpec{}
			req.Raid.Parties[0].Players[0].Profession1 = proto.Profession_ProfessionUnknown
		}
		expected := map[string]bool{
			"No swap set configured in Settings.":                       true,
			"Bear (#1) does not know spell {ItemID: 9449}":              true,
			"Bear (#1) picks no potion for {OtherID: 26}":               true,
			"No aura found on Target 1 for: {SpellID: 440667, Tag: -1}": true,
		}
		if !consumes {
			expected["Bear (#1) does not know spell {ItemID: 10646}"] = true
			expected["Bear (#1) does not know spell {ItemID: 9421}"] = true
			expected["Bear (#1) picks no potion for {OtherID: 25}"] = true
		}
		for line, messages := range bearAPLWarnings(t, req) {
			missingVindication := false
			for _, message := range messages {
				if message == "No aura found on Target 1 for: {SpellID: 440667, Tag: -1}" {
					missingVindication = true
				}
			}
			for _, message := range messages {
				if missingVindication && message == "A condition branch in this 'And' check could not be resolved and was dropped." {
					continue
				}
				if !expected[message] {
					t.Errorf("consumes=%t line %d: %s", consumes, line, message)
				}
			}
		}
		result := core.RunSim(req, nil, simsignals.CreateSignals())
		if result.Error != nil {
			t.Fatal(result.Error.Message)
		}
		metrics := result.RaidMetrics.Parties[0].Players[0]
		for _, id := range []int32{spellPrimalBite, spellLacerate, spellMaul} {
			if bearCastsOf(metrics.Actions, id) == 0 {
				t.Errorf("consumes=%t spell %d was never cast", consumes, id)
			}
		}
	}
}

func bearSpellID(id int32) *proto.ActionID {
	return &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}}
}

func spellCost(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_SpellCurrentCost{SpellCurrentCost: &proto.APLValueSpellCurrentCost{SpellId: bearSpellID(id)}}}
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

func cmp(op proto.APLValueCompare_ComparisonOperator, lhs, rhs *proto.APLValue) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{Op: op, Lhs: lhs, Rhs: rhs}}}
}

func mathOp(op proto.APLValueMath_MathOperator, lhs, rhs *proto.APLValue) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Math{Math: &proto.APLValueMath{Op: op, Lhs: lhs, Rhs: rhs}}}
}

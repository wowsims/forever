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
	for _, id := range []int32{1238073, 9881, 22812} {
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
	items[proto.ItemSlot_ItemSlotMainHand].Id = 866
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
				Rotation: core.GetAplRotation("../../../ui/specs/druid/feralbear/apls", "default").Rotation,
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
					trees[2] = fmt.Sprintf("0%d", rank)
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

func TestBearPrepullEnrage(t *testing.T) {
	rotation := selectedAdaptiveBearAPL()
	rotation.PriorityList = nil
	req := bearNakedRequest()
	req.Raid.Parties[0].Players[0].Rotation = rotation
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
		{"lacerate-funded", false, true, 12, false},
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
				id := row.Action.GetCastSpell().GetSpellId().GetItemId()
				return id != itemStoneshield && id != 13442 && row.Action.GetAutocastOtherCooldowns() == nil
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

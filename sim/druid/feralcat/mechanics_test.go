//go:build with_db

package feralcat

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	assetsdb "github.com/wowsims/forever/assets/database"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/druid"
)

// 16 balance, 20 feral, 16 restoration. Furor is the second feral talent.
const furorTalents = "0000000000000000-05000000000000000000-0000000000000000"

// Shifting Power is the 20th feral talent. Natural Shapeshifter is the 5th restoration talent.
const (
	shiftingPowerTalents = "0000000000000000-00000000000000000001-0000000000000000"
	shapeshifterTalents  = "0000000000000000-00000000000000000001-0000300000000000"
)

func newCat(t *testing.T, talents string, consumes *proto.ConsumesSpec) (*core.Simulation, *FeralDruid) {
	t.Helper()
	return newCatWearing(t, talents, consumes, nil)
}

func newCatWearing(t *testing.T, talents string, consumes *proto.ConsumesSpec, worn map[proto.ItemSlot]int32) (*core.Simulation, *FeralDruid) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 1613
	for slot, id := range worn {
		items[slot].Id = id
	}
	if consumes == nil {
		consumes = &proto.ConsumesSpec{}
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Cat", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
				TalentsString: talents,
				Spec: &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{
					Options: &proto.FeralCatDruid_Options{},
				}},
				Equipment:   &proto.EquipmentSpec{Items: items},
				Consumables: consumes,
				Buffs:       &proto.IndividualBuffs{},
				Rotation:    &proto.APLRotation{},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 60},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*FeralDruid)
}

func fillResource(sim *core.Simulation, cat *FeralDruid) {
	cat.AddEnergy(sim, cat.MaximumEnergy(), cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	cat.AddMana(sim, cat.MaxMana(), cat.NewManaMetrics(core.ActionID{SpellID: 1}))
	cat.AddRage(sim, cat.MaximumRage(), cat.NewRageMetrics(core.ActionID{SpellID: 1}))
}

// Ferocity 5/5 and Shredding Attacks 3/3: tooltips 16934 ("5 Rage or Energy") and 16966 ("Shred by 18",
// "Lacerate by 3").
func TestFeralCostTalentsLowerEnergyAndRageInTheirOwnUnits(t *testing.T) {
	_, base := newCat(t, "0000000000000000-00000000000100000010-0000000000000000", nil)
	_, talented := newCat(t, "0000000000000000-50000030000100000010-0000000000000000", nil)
	for _, c := range []struct {
		name       string
		base, with *druid.DruidSpell
		less       float64
	}{
		{"Shred", base.Shred, talented.Shred, 18},
		{"Rake", base.Rake, talented.Rake, 5},
		{"Claw", base.Claw, talented.Claw, 5},
		{"Lacerate", base.Lacerate, talented.Lacerate, 3},
		{"Swipe", base.Swipe, talented.Swipe, 5},
		{"Primal Bite", base.PrimalBite, talented.PrimalBite, 5},
	} {
		if got := c.base.Cost.GetCurrentCost() - c.with.Cost.GetCurrentCost(); got != c.less {
			t.Errorf("%s costs %v less with the talents, want %v", c.name, got, c.less)
		}
	}
}

func TestFurorRestoresEnergyAfterLeavingCat(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	if !cat.InForm(druid.Cat) {
		t.Fatal("the cat did not start in cat form")
	}
	fillResource(sim, cat)
	cat.SpendEnergy(sim, 60, cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	cat.CatFormAura.Deactivate(sim)
	sim.CurrentTime += 3 * time.Second
	fillResource(sim, cat)
	if !cat.CatForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("cat form did not cast on the way back")
	}
	// 40 stored plus 10 for each of the 3 seconds outside the forms, all of it at 5/5.
	if got := cat.CurrentEnergy(); got != 70 {
		t.Errorf("energy back in cat form = %v, want 70", got)
	}
}

func TestFurorPreservesEnergyInABearForm(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	fillResource(sim, cat)
	cat.SpendEnergy(sim, 60, cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	if !cat.DireBearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("dire bear form did not cast")
	}
	if cat.InForm(druid.Cat) {
		t.Fatal("dire bear form left the cat in cat form")
	}
	sim.CurrentTime += 3 * time.Second
	fillResource(sim, cat)
	if !cat.CatForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("cat form did not cast on the way back")
	}
	if got := cat.CurrentEnergy(); got != 40 {
		t.Errorf("energy back in cat form after dire bear form = %v, want 40", got)
	}
}

func TestWithoutFurorCatFormComesBackEmpty(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	fillResource(sim, cat)
	if !cat.DireBearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("dire bear form did not cast")
	}
	sim.CurrentTime += 3 * time.Second
	fillResource(sim, cat)
	if !cat.CatForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("cat form did not cast on the way back")
	}
	if got := cat.CurrentEnergy(); got != 0 {
		t.Errorf("energy back in cat form without Furor = %v, want 0", got)
	}
}

func TestProwlHoldsAutoAttacksUntilTheOpener(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	sim.CurrentTime = -time.Second
	if !cat.Prowl.Cast(sim, cat.CurrentTarget) {
		t.Fatal("prowl did not cast before the pull")
	}
	mh, _ := cat.AutoAttacks.SwingAt()
	if mh != core.NeverExpires-time.Second {
		t.Errorf("main-hand swing while prowling at %v, want held", mh)
	}
	sim.CurrentTime = 0
	fillResource(sim, cat)
	if cat.Prowl.CanCast(sim, cat.CurrentTarget) {
		t.Error("prowl could be cast in combat")
	}
	if !cat.Pounce.Cast(sim, cat.CurrentTarget) {
		t.Fatal("pounce did not cast from prowl")
	}
	if cat.ProwlAura.IsActive() {
		t.Error("pounce left the druid prowling")
	}
	if mh, _ := cat.AutoAttacks.SwingAt(); mh != 0 {
		t.Errorf("main-hand swing after the opener at %v, want 0", mh)
	}
}

func TestCatQueuesMaulOntoTheBearSwing(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	fillResource(sim, cat)
	if !cat.DireBearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("dire bear form did not cast")
	}
	fillResource(sim, cat)
	if !cat.Maul.Cast(sim, cat.CurrentTarget) {
		t.Fatal("maul did not queue")
	}
	for sim.CurrentTime < 100*time.Millisecond {
		if sim.Step() {
			break
		}
	}
	if got := cat.AutoAttacks.MaybeReplaceMHSwing(sim, nil); got == nil || got.ActionID.SpellID != 9881 {
		t.Fatalf("the queued swing was %v, want maul 9881", got)
	}
}

func TestCatSwingDoesNotRequestAMaulQueueCheck(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	cat.ReactionTime = 250 * time.Millisecond
	cat.SetRotationTimer(sim, 5*time.Second)
	seen := false
	swing := cat.AutoAttacks.MHAuto()
	apply := swing.ApplyEffects
	swing.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		seen = true
		if at := cat.NextRotationActionAt(); at != 5*time.Second {
			t.Fatalf("rotation before the Cat hit = %s, want 5s", at)
		}
		apply(sim, target, spell)
	}
	cat.AutoAttacks.EnableMeleeSwing(sim)
	for !seen && !sim.Step() {
	}
	if !seen {
		t.Fatal("Cat never swung")
	}
	if at := cat.NextRotationActionAt(); at != sim.CurrentTime+cat.ReactionTime {
		t.Fatalf("rotation after the Cat hit = %s, want %s", at, sim.CurrentTime+cat.ReactionTime)
	}
	t.Logf("Cat swing at %s preserves the scheduled check before impact and schedules the post-hit check at %s", sim.CurrentTime, cat.NextRotationActionAt())
}

func TestCatSwingClearcastingStillReachesTheRotation(t *testing.T) {
	rotation := core.APLRotationFromJsonString(`{"type":"TypeAPL","priorityList":[{"action":{"condition":{"auraIsActive":{"auraId":{"spellId":16870}}},"castSpell":{"spellId":{"spellId":9830}}}}]}`)
	request := catPowershiftRequest(furorTalents, rotation, false)
	request.Raid.Parties[0].Players[0].ReactionTimeMs = 250
	sim := core.NewSim(request, simsignals.CreateSignals())
	sim.Reset()
	cat := sim.Raid.Parties[0].Players[0].(*FeralDruid)
	cat.SpendEnergy(sim, cat.CurrentEnergy(), cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	cat.AddDynamicMod(core.SpellModConfig{Kind: core.SpellMod_ProcChance_Flat, ClassFlags: spelldata.MustFind(16864).ClassFlags, FloatValue: 100}).Activate()
	cat.AddStatDynamic(sim, stats.HitRating, 100*core.HitRatingPerHitPercent)
	cat.AddStatDynamic(sim, stats.ExpertiseRating, 100*core.ExpertiseRatingPerExpertisePercent)
	cat.SetRotationTimer(sim, 5*time.Second)
	procAt, castAt := core.NeverExpires, core.NeverExpires
	cat.ClearcastingAura.ApplyOnGain(func(_ *core.Aura, sim *core.Simulation) { procAt = sim.CurrentTime })
	apply := cat.Shred.ApplyEffects
	cat.Shred.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		castAt = sim.CurrentTime
		if spell.CurCast.Cost != 0 {
			t.Errorf("Clearcasting Shred costs %g", spell.CurCast.Cost)
		}
		apply(sim, target, spell)
	}
	cat.AutoAttacks.EnableMeleeSwing(sim)
	for castAt == core.NeverExpires && sim.CurrentTime < time.Second && !sim.Step() {
	}
	if procAt != core.SpellBatchWindow || castAt != 250*time.Millisecond || cat.ClearcastingAura.IsActive() {
		t.Fatalf("Clearcasting at %s, Shred at %s, proc active after Shred = %v", procAt, castAt, cat.ClearcastingAura.IsActive())
	}
	t.Logf("auto-triggered Clearcasting at %s is consumed by free Shred at %s", procAt, castAt)
}

func TestMaulQueueEndsWhenLeavingBear(t *testing.T) {
	for _, dire := range []bool{false, true} {
		for _, leaveAt := range []time.Duration{25 * time.Millisecond, 75 * time.Millisecond} {
			t.Run(fmt.Sprintf("dire=%v/leave=%s", dire, leaveAt), func(t *testing.T) {
				sim, cat := newCat(t, furorTalents, nil)
				for iteration := range 2 {
					if iteration != 0 {
						sim.Cleanup()
						sim.Reset()
					}
					cat.AutoAttacks.CancelAutoSwing(sim)
					cat.ClearForm(sim)
					form := cat.BearFormAura
					if dire {
						form = cat.DireBearFormAura
					}
					form.Activate(sim)
					fillResource(sim, cat)
					if !cat.Maul.Cast(sim, cat.CurrentTarget) {
						t.Fatal("Maul did not queue")
					}
					sim.AddPendingAction(&core.PendingAction{NextActionAt: leaveAt, OnAction: func(sim *core.Simulation) {
						cat.ClearForm(sim)
						cat.CatFormAura.Activate(sim)
					}})
					for sim.CurrentTime < 100*time.Millisecond && !sim.Step() {
					}
					if cat.GetAura("Maul Queue Aura").IsActive() {
						t.Fatal("Maul stayed queued in Cat Form")
					}
					cat.ClearForm(sim)
					form.Activate(sim)
					fillResource(sim, cat)
					if got := cat.AutoAttacks.MaybeReplaceMHSwing(sim, nil); got != nil {
						t.Fatal("a stale Maul queue survived Bear re-entry")
					}
					if !cat.Maul.Cast(sim, cat.CurrentTarget) {
						t.Fatal("Maul could not queue after Bear re-entry")
					}
					until := sim.CurrentTime + 100*time.Millisecond
					for sim.CurrentTime < until && !sim.Step() {
					}
					if got := cat.AutoAttacks.MaybeReplaceMHSwing(sim, nil); got == nil || got.ActionID.SpellID != 9881 {
						t.Fatal("Maul did not replace the Bear swing after re-entry")
					}
				}
			})
		}
	}
}

func TestCowerAndFaerieFireEndProwl(t *testing.T) {
	for _, spell := range []struct {
		name string
		cast func(*core.Simulation, *FeralDruid) bool
	}{
		{"cower", func(sim *core.Simulation, cat *FeralDruid) bool {
			return cat.Cower.Cast(sim, cat.CurrentTarget)
		}},
		{"faerie fire", func(sim *core.Simulation, cat *FeralDruid) bool {
			return cat.FaerieFire.Cast(sim, cat.CurrentTarget)
		}},
	} {
		t.Run(spell.name, func(t *testing.T) {
			sim, cat := newCat(t, furorTalents, nil)
			sim.CurrentTime = -time.Second
			if !cat.Prowl.Cast(sim, cat.CurrentTarget) {
				t.Fatal("prowl did not cast before the pull")
			}
			sim.CurrentTime = 3 * time.Second
			fillResource(sim, cat)
			if !spell.cast(sim, cat) {
				t.Fatalf("%s did not cast from prowl", spell.name)
			}
			if cat.ProwlAura.IsActive() {
				t.Errorf("%s left the druid prowling", spell.name)
			}
			if mh, _ := cat.AutoAttacks.SwingAt(); mh != 3*time.Second {
				t.Errorf("main-hand swing after %s at %v, want 3s", spell.name, mh)
			}
		})
	}
}

func TestFaerieFireFormCooldown(t *testing.T) {
	for _, form := range []string{"Cat", "Bear"} {
		t.Run(form, func(t *testing.T) {
			sim, cat := newCat(t, furorTalents, nil)
			fillResource(sim, cat)
			shift, aura := cat.CatForm, cat.CatFormAura
			if form == "Bear" {
				shift, aura = cat.BearForm, cat.BearFormAura
				cat.CatFormAura.Deactivate(sim)
				if !shift.Cast(sim, cat.CurrentTarget) {
					t.Fatal("initial form cast failed")
				}
				sim.CurrentTime = cat.GCD.ReadyAt()
			}
			if !cat.FaerieFire.Cast(sim, cat.CurrentTarget) {
				t.Fatal("in-form Faerie Fire cast failed")
			}
			expires := sim.CurrentTime + 6*time.Second
			if got := cat.FaerieFire.CD.ReadyAt(); got != expires {
				t.Fatalf("cooldown expires at %v, want %v", got, expires)
			}
			sim.CurrentTime = cat.GCD.ReadyAt()
			if cat.FaerieFire.CanCast(sim, cat.CurrentTarget) {
				t.Fatal("Faerie Fire available in form before its cooldown expires")
			}
			aura.Deactivate(sim)
			if cat.FaerieFire.CD.Duration != 0 {
				t.Fatalf("humanoid cooldown duration = %v, want 0", cat.FaerieFire.CD.Duration)
			}
			if got := cat.FaerieFire.CD.ReadyAt(); got != expires {
				t.Fatalf("leaving form changed the cooldown expiry to %v, want %v", got, expires)
			}
			if cat.FaerieFire.CanCast(sim, cat.CurrentTarget) {
				t.Fatal("leaving form cleared the active Faerie Fire cooldown")
			}
			sim.CurrentTime = expires
			for cast := range 2 {
				fillResource(sim, cat)
				if !cat.FaerieFire.Cast(sim, cat.CurrentTarget) {
					t.Fatalf("humanoid Faerie Fire cast %d failed", cast+1)
				}
				if !cat.FaerieFire.CD.IsReady(sim) {
					t.Fatal("humanoid Faerie Fire started a cooldown")
				}
				sim.CurrentTime = cat.GCD.ReadyAt()
			}
			fillResource(sim, cat)
			if !shift.Cast(sim, cat.CurrentTarget) {
				t.Fatal("return to form failed")
			}
			if !cat.FaerieFire.CD.IsReady(sim) {
				t.Fatal("entering form added a cooldown to the humanoid cast")
			}
			sim.CurrentTime = cat.GCD.ReadyAt()
			if !cat.FaerieFire.Cast(sim, cat.CurrentTarget) {
				t.Fatal("Faerie Fire unavailable after entering form and waiting for the GCD")
			}
			if got := cat.FaerieFire.CD.TimeToReady(sim); got != 6*time.Second {
				t.Fatalf("new in-form cooldown = %v, want 6s", got)
			}
		})
	}
}

func TestFaerieFireRefreshesWhileUp(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	fillResource(sim, cat)
	target := cat.CurrentTarget
	aura := cat.FaerieFireAuras.Get(target)
	aura.Activate(sim)
	sim.CurrentTime += 35 * time.Second
	if !cat.FaerieFire.CanCast(sim, target) {
		t.Fatal("faerie fire could not be recast with 5 s of its debuff left")
	}
}

func emptyEnergy(sim *core.Simulation, cat *FeralDruid) {
	cat.SpendEnergy(sim, cat.CurrentEnergy(), cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
}

func TestShiftingPowerTradesManaForEnergy(t *testing.T) {
	sim, cat := newCat(t, shiftingPowerTalents, nil)
	fillResource(sim, cat)
	emptyEnergy(sim, cat)
	mana := cat.CurrentMana()
	if !cat.ShiftingPower.Cast(sim, cat.CurrentTarget) {
		t.Fatal("shifting power did not cast in cat form")
	}
	if got := cat.CurrentEnergy(); got != 40 {
		t.Errorf("energy after shifting power = %v, want 40", got)
	}
	// 55% of base mana, the shapeshift's own price.
	if got, want := mana-cat.CurrentMana(), float64(int32(55*cat.BaseMana)/100); got != want {
		t.Errorf("mana spent = %v, want %v", got, want)
	}
}

func TestShiftingPowerNeedsTheTalentAndCatForm(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	fillResource(sim, cat)
	emptyEnergy(sim, cat)
	if cat.ShiftingPower.CanCast(sim, cat.CurrentTarget) {
		t.Error("shifting power could be cast without the talent")
	}
	sim, cat = newCat(t, shiftingPowerTalents, nil)
	fillResource(sim, cat)
	if !cat.DireBearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("dire bear form did not cast")
	}
	fillResource(sim, cat)
	emptyEnergy(sim, cat)
	if cat.ShiftingPower.CanCast(sim, cat.CurrentTarget) {
		t.Error("shifting power could be cast in dire bear form")
	}
}

func TestNaturalShapeshifterCheapensShiftingPower(t *testing.T) {
	_, cat := newCat(t, shiftingPowerTalents, nil)
	full := cat.ShiftingPower.Cost.GetCurrentCost()
	_, cat = newCat(t, shapeshifterTalents, nil)
	// 3/3 is 30% off, the same as the form spells.
	if got, want := cat.ShiftingPower.Cost.GetCurrentCost(), full*0.7; math.Abs(got-want) > 1 {
		t.Errorf("shifting power cost with natural shapeshifter = %v, want %v", got, want)
	}
}

func TestWolfsheadAddsShiftingPowerEnergyAndEnrageRage(t *testing.T) {
	sim, cat := newCatWearing(t, shiftingPowerTalents, nil, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotHead: 8345})
	fillResource(sim, cat)
	emptyEnergy(sim, cat)
	if !cat.ShiftingPower.Cast(sim, cat.CurrentTarget) {
		t.Fatal("shifting power did not cast")
	}
	if got := cat.CurrentEnergy(); got != 45 {
		t.Errorf("energy after shifting power with wolfshead = %v, want 45", got)
	}
	sim.CurrentTime += 2 * time.Second
	if !cat.BearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("bear form did not cast")
	}
	cat.SpendRage(sim, cat.CurrentRage(), cat.NewRageMetrics(core.ActionID{SpellID: 1}))
	if !cat.Enrage.Cast(sim, cat.CurrentTarget) {
		t.Fatal("enrage did not cast")
	}
	// The row's 10 instant rage, raised by half.
	if got := cat.CurrentRage(); got != 15 {
		t.Errorf("rage after enrage with wolfshead = %v, want 15", got)
	}
}

func TestHowlingIdolShortensShiftingPower(t *testing.T) {
	sim, cat := newCatWearing(t, shiftingPowerTalents, nil, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotRanged: 272427})
	fillResource(sim, cat)
	emptyEnergy(sim, cat)
	if !cat.ShiftingPower.Cast(sim, cat.CurrentTarget) {
		t.Fatal("shifting power did not cast")
	}
	// 1291059's effect 1 states the reduction as negative milliseconds.
	want := spelldata.MustFind(cat.ShiftingPower.SpellID).Cooldown() + spelldata.MustFind(1291059).EffectN(1).TimeValue()
	if got := cat.ShiftingPower.ReadyAt(); got != want {
		t.Errorf("shifting power ready at %v with the idol, want %v", got, want)
	}
}

// The October 8 hotfixes: "Shifting Power can no longer be used while already at Full Energy."
func TestShiftingPowerIsRefusedAtFullEnergy(t *testing.T) {
	sim, cat := newCat(t, shiftingPowerTalents, nil)
	fillResource(sim, cat)
	if cat.ShiftingPower.Cast(sim, cat.CurrentTarget) {
		t.Fatal("shifting power cast at full energy")
	}
	cat.SpendEnergy(sim, 1, cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	if !cat.ShiftingPower.Cast(sim, cat.CurrentTarget) {
		t.Fatal("shifting power did not cast one energy under full")
	}
}

func TestFurorGrantsRageOnEnteringBear(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	cat.AddMana(sim, cat.MaxMana(), cat.NewManaMetrics(core.ActionID{SpellID: 1}))
	if !cat.BearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("bear form did not cast")
	}
	if !cat.InForm(druid.Bear) {
		t.Fatal("bear form did not set the bear form")
	}
	// 5/5 is a guaranteed 10 rage. The bar starts empty.
	if got := cat.CurrentRage(); got != 10 {
		t.Errorf("rage on entering bear = %v, want 10", got)
	}
}

func TestClawAwardsComboPointsAndRipSpendsThem(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	fillResource(sim, cat)
	var landed bool
	for range 15 {
		if cat.Claw.Cast(sim, cat.CurrentTarget) && cat.ComboPoints() > 0 {
			landed = true
			break
		}
		fillResource(sim, cat)
		sim.CurrentTime += 2 * time.Second
	}
	if !landed {
		t.Fatal("claw never awarded a combo point")
	}
	points := cat.ComboPoints()
	fillResource(sim, cat)
	sim.CurrentTime += 2 * time.Second
	if !cat.Rip.Cast(sim, cat.CurrentTarget) {
		t.Fatal("rip did not cast")
	}
	if got := cat.ComboPoints(); got != 0 {
		t.Errorf("combo points after rip = %d, want 0 (spent %d)", got, points)
	}
	if !cat.Rip.Dot(cat.CurrentTarget).IsActive() {
		t.Error("rip left no bleed")
	}
}

func TestShredRequiresATargetThatCanBeShredded(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	fillResource(sim, cat)
	cat.PseudoStats.InFrontOfTarget = true
	if cat.Shred.CanCast(sim, cat.CurrentTarget) {
		t.Error("shred could be cast while in front of the target")
	}
	cat.PseudoStats.InFrontOfTarget = false
	if !cat.Shred.CanCast(sim, cat.CurrentTarget) {
		t.Error("shred could not be cast from behind")
	}
}

func TestConsumablesFollowTheNotShapeshiftedRow(t *testing.T) {
	interesting := map[proto.ConsumableType]bool{
		proto.ConsumableType_ConsumableTypePotion:    true,
		proto.ConsumableType_ConsumableTypeExplosive: true,
		proto.ConsumableType_ConsumableTypeDevice:    true,
		proto.ConsumableType_ConsumableTypeConjured:  true,
	}
	var breakingID, safeID int32
	var breakingSpell, safeSpell int32
	for _, consumable := range assetsdb.Load().Consumables {
		id := consumable.Id
		item := core.GetConsumableByID(id)
		if !interesting[item.Type] || len(item.OnUseEffects()) == 0 {
			continue
		}
		row := spelldata.Find(item.OnUse().GetBuffId())
		if row == spelldata.Nil {
			continue
		}
		breaks := row.NotShapeshifted()
		if breaks {
			t.Logf("breaks form: %d %s spell %d", id, item.Name, row.ID)
		}
		// The cast below uses a breaker and a potion that need no profession. The log above is every row.
		if item.RequiredProfession != 0 {
			continue
		}
		if breaks && breakingID == 0 {
			breakingID, breakingSpell = id, row.ID
		}
		if !breaks && safeID == 0 && item.Type == proto.ConsumableType_ConsumableTypePotion && id == 13444 {
			safeID, safeSpell = id, row.ID
		}
	}
	if breakingID == 0 || safeID == 0 {
		t.Fatalf("missing a consumable pair, breaking %d safe %d", breakingID, safeID)
	}

	sim, cat := newCat(t, furorTalents, &proto.ConsumesSpec{ResourcePotId: safeID, ExtraIds: []int32{breakingID}})
	breaking := cat.GetSpell(core.ActionID{ItemID: breakingID})
	safe := cat.GetSpell(core.ActionID{ItemID: safeID})
	if breaking == nil || safe == nil {
		t.Fatalf("consumable spells did not register, item %d spell %d and item %d spell %d", breakingID, breakingSpell, safeID, safeSpell)
	}
	if !breaking.CastRequirement.NotShapeshifted || safe.CastRequirement.NotShapeshifted {
		t.Fatalf("cast requirement breaking=%v safe=%v", breaking.CastRequirement.NotShapeshifted, safe.CastRequirement.NotShapeshifted)
	}

	sawBreaker := false
	for _, mcd := range cat.GetMajorCooldowns() {
		if mcd.Spell != breaking {
			continue
		}
		sawBreaker = true
		if mcd.TryActivate(sim, cat.GetCharacter()) || !cat.InForm(druid.Cat) {
			t.Error("a cat automatically used a consumable that leaves cat form")
		}
	}
	if !sawBreaker {
		t.Fatal("the form-breaking consumable was not a major cooldown")
	}

	castConsumable := func(form druid.DruidForm, spell *core.Spell, wantShift bool) {
		t.Helper()
		fillResource(sim, cat)
		waitReady(sim, cat)
		if !spell.Cast(sim, cat.CurrentTarget) {
			t.Fatalf("%s did not cast in %v", spell.ActionID, form)
		}
		waitReady(sim, cat)
		shifted := cat.InForm(druid.Cat) || cat.InForm(druid.Bear)
		if wantShift && shifted {
			t.Errorf("%s left the druid in a form", spell.ActionID)
		}
		if !wantShift && !cat.InForm(form) {
			t.Errorf("%s left %v", spell.ActionID, form)
		}
	}

	castConsumable(druid.Cat, breaking, true)
	fillResource(sim, cat)
	waitReady(sim, cat)
	if !cat.CatForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("could not return to cat form")
	}
	resetCooldown(safe)
	castConsumable(druid.Cat, safe, false)

	fillResource(sim, cat)
	waitReady(sim, cat)
	if !cat.BearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("could not enter bear form")
	}
	resetCooldown(breaking)
	castConsumable(druid.Bear, breaking, true)
	fillResource(sim, cat)
	waitReady(sim, cat)
	if !cat.BearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("could not return to bear form")
	}
	resetCooldown(safe)
	castConsumable(druid.Bear, safe, false)
}

func waitReady(sim *core.Simulation, cat *FeralDruid) {
	deadline := sim.CurrentTime + 10*time.Second
	for sim.CurrentTime < deadline {
		if cat.GCD.IsReady(sim) && cat.Hardcast.Expires <= sim.CurrentTime {
			return
		}
		if sim.Step() {
			return
		}
	}
}

// Subtlety 3/3 is the 6th restoration talent. Its mask is Nature and Arcane, so physical threat stays put.
func TestSubtletyThreatIsNatureAndArcane(t *testing.T) {
	_, plain := newCat(t, "", nil)
	_, cat := newCat(t, "0000000000000000-00000000000000000000-0000030000000000", nil)
	got := cat.PseudoStats.SchoolThreatMultiplier
	base := plain.PseudoStats.SchoolThreatMultiplier
	if got[stats.SchoolIndexNature] != 0.7 || got[stats.SchoolIndexArcane] != 0.7 {
		t.Fatalf("Subtlety 3/3 threat: nature %.3f arcane %.3f, want 0.70", got[stats.SchoolIndexNature], got[stats.SchoolIndexArcane])
	}
	if got[stats.SchoolIndexPhysical] != base[stats.SchoolIndexPhysical] || cat.PseudoStats.ThreatMultiplier != plain.PseudoStats.ThreatMultiplier {
		t.Fatalf("physical %.3f (base %.3f), global %.3f (base %.3f)",
			got[stats.SchoolIndexPhysical], base[stats.SchoolIndexPhysical],
			cat.PseudoStats.ThreatMultiplier, plain.PseudoStats.ThreatMultiplier)
	}
}

func resetCooldown(spell *core.Spell) {
	if spell.CD.Timer != nil {
		spell.CD.Reset()
	}
	if spell.SharedCD.Timer != nil {
		spell.SharedCD.Reset()
	}
}

func TestProwlAttacksRejectShadowmeld(t *testing.T) {
	for _, ability := range []struct {
		name  string
		spell func(*FeralDruid) *druid.DruidSpell
	}{
		{"Ravage", func(cat *FeralDruid) *druid.DruidSpell { return cat.Ravage }},
		{"Pounce", func(cat *FeralDruid) *druid.DruidSpell { return cat.Pounce }},
	} {
		for _, at := range []time.Duration{-time.Second, 10 * time.Second} {
			t.Run(fmt.Sprintf("%s/%s", ability.name, at), func(t *testing.T) {
				sim, cat := newCat(t, "", nil)
				sim.CurrentTime = at
				fillResource(sim, cat)
				attack := ability.spell(cat)
				if attack.CanCast(sim, cat.CurrentTarget) {
					t.Fatal("attack was available without Prowl")
				}
				if cat.Shadowmeld.Flags.Matches(core.SpellFlagAPL) {
					t.Fatal("Shadowmeld was exposed in the APL spell picker")
				}
				if !cat.Shadowmeld.Cast(sim, cat.CurrentTarget) {
					t.Fatal("Shadowmeld did not cast")
				}
				energy, gcd := cat.CurrentEnergy(), cat.GCD.ReadyAt()
				if attack.CanCast(sim, cat.CurrentTarget) || attack.Cast(sim, cat.CurrentTarget) {
					t.Fatal("Shadowmeld enabled an attack requiring Prowl")
				}
				if cat.CurrentEnergy() != energy || cat.GCD.ReadyAt() != gcd || !cat.ShadowmeldAura.IsActive() {
					t.Fatal("rejected attack changed energy, the GCD or Shadowmeld")
				}
				if got := cat.Shadowmeld.ReadyAt() - sim.CurrentTime; got != 2*time.Minute {
					t.Errorf("Shadowmeld is ready in %v, want 2m", got)
				}
				t.Logf("%s at %s: Shadowmeld active, attack rejected, energy and GCD unchanged", ability.name, at)
			})
		}
		t.Run(ability.name+"/Prowl", func(t *testing.T) {
			sim, cat := newCat(t, "", nil)
			sim.CurrentTime = -time.Second
			if !cat.Prowl.Cast(sim, cat.CurrentTarget) {
				t.Fatal("Prowl did not cast before combat")
			}
			sim.CurrentTime = time.Second
			fillResource(sim, cat)
			if !ability.spell(cat).Cast(sim, cat.CurrentTarget) {
				t.Fatal("attack did not cast from Prowl")
			}
			if cat.ProwlAura.IsActive() {
				t.Fatal("attack left Prowl active")
			}
			t.Logf("%s cast from Prowl and ended it", ability.name)
		})
	}
}

// With Clearcasting at the Pull, the cat starts the fight with Clearcasting up.
func TestPrepullClearcastingStartsTheFightClearcasting(t *testing.T) {
	for _, on := range []bool{false, true} {
		req := sweepRequest(catTalents, defaultCatRotation(), false, 1)
		req.SimOptions.Debug = true
		req.Raid.Parties[0].Players[0].GetFeralCatDruid().Options.PrepullClearcasting = on
		result := core.RunSim(req, nil, simsignals.CreateSignals())
		if result.Error != nil {
			t.Fatal(result.Error.Message)
		}
		got := false
		for line := range strings.SplitSeq(result.Logs, "\n") {
			got = got || strings.HasPrefix(line, "[-1.50]") && strings.Contains(line, "Aura gained: {SpellID: 16870}")
		}
		if got != on {
			t.Errorf("option %v: Clearcasting gained before the pull = %v", on, got)
		}
	}
}

func TestFurorPausesEnergyInBear(t *testing.T) {
	for rank := 0; rank <= 5; rank++ {
		for _, dire := range []bool{false, true} {
			for _, elapsed := range []time.Duration{0, 3 * time.Second, 20 * time.Second} {
				t.Run(fmt.Sprintf("rank%d/dire%t/%v", rank, dire, elapsed), func(t *testing.T) {
					sim, cat := newCat(t, fmt.Sprintf("0000000000000000-0%d000000000000000000-0000000000000000", rank), nil)
					cat.SpendEnergy(sim, 60, cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
					cat.CatFormAura.Deactivate(sim)
					sim.CurrentTime += 2 * time.Second
					bear := cat.BearFormAura
					if dire {
						bear = cat.DireBearFormAura
					}
					bear.Activate(sim)
					sim.CurrentTime += elapsed
					bear.Deactivate(sim)
					sim.CurrentTime += time.Second
					cat.CatFormAura.Activate(sim)
					want := float64(rank) / 5 * 70
					if math.Abs(cat.CurrentEnergy()-want) > 1e-9 {
						t.Fatalf("energy = %v, want %v", cat.CurrentEnergy(), want)
					}
					cat.CatFormAura.Deactivate(sim)
					bear.Activate(sim)
					sim.CurrentTime += 20 * time.Second
					bear.Deactivate(sim)
					cat.CatFormAura.Activate(sim)
					want *= float64(rank) / 5
					if math.Abs(cat.CurrentEnergy()-want) > 1e-9 {
						t.Fatalf("repeated shift energy = %v, want %v", cat.CurrentEnergy(), want)
					}
					cat.CatFormAura.Deactivate(sim)
					bear.Activate(sim)
					sim.Cleanup()
					sim.Reset()
					if cat.CurrentEnergy() != 100 || cat.EnergyPerTick != 1 || !cat.InForm(druid.Cat) {
						t.Fatalf("reset: energy %v, tick %v, form %v", cat.CurrentEnergy(), cat.EnergyPerTick, cat.InForm(druid.Cat))
					}
				})
			}
		}
	}
}

func TestFurorBearTicksAndDirectTransitions(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	cat.SpendEnergy(sim, 70, cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	if !cat.BearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Bear entry failed")
	}
	for sim.CurrentTime < 2*time.Second {
		if sim.Step() {
			t.Fatal("simulation ended in Bear")
		}
	}
	if cat.CurrentEnergy() != 30 {
		t.Fatalf("Bear regenerated energy: %v", cat.CurrentEnergy())
	}
	cat.AddMana(sim, cat.MaxMana(), cat.NewManaMetrics(core.ActionID{SpellID: 1}))
	if !cat.DireBearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Dire Bear entry failed")
	}
	for sim.CurrentTime < 4*time.Second {
		if sim.Step() {
			t.Fatal("simulation ended in Dire Bear")
		}
	}
	if cat.CurrentEnergy() != 30 {
		t.Fatalf("Dire Bear regenerated energy: %v", cat.CurrentEnergy())
	}
	cat.AddMana(sim, cat.MaxMana(), cat.NewManaMetrics(core.ActionID{SpellID: 1}))
	if !cat.CatForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Cat return failed")
	}
	if cat.CurrentEnergy() != 30 {
		t.Fatalf("Cat energy = %v, want 30", cat.CurrentEnergy())
	}
	returned := sim.CurrentTime
	for sim.CurrentTime < returned+time.Second {
		if sim.Step() {
			t.Fatal("simulation ended in Cat")
		}
	}
	if math.Abs(cat.CurrentEnergy()-40) > 1e-9 {
		t.Fatalf("Cat regeneration = %v, want 40", cat.CurrentEnergy())
	}
}

func TestFaerieFireOmenEligibilityByForm(t *testing.T) {
	for _, form := range []druid.DruidForm{druid.Humanoid, druid.Cat, druid.Bear, druid.DireBear} {
		t.Run(fmt.Sprint(form), func(t *testing.T) {
			sim, cat := newCat(t, furorTalents, nil)
			cat.ClearForm(sim)
			switch form {
			case druid.Cat:
				cat.CatFormAura.Activate(sim)
			case druid.Bear:
				cat.BearFormAura.Activate(sim)
			case druid.DireBear:
				cat.DireBearFormAura.Activate(sim)
			}
			procs := 0
			cat.AutoAttacks.CancelAutoSwing(sim)
			cat.ClearcastingAura.ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) { procs++ })
			for i := range 400 {
				sim.AddPendingAction(&core.PendingAction{NextActionAt: time.Duration(i+1) * 100 * time.Millisecond, OnAction: func(sim *core.Simulation) {
					cat.ClearcastingAura.Deactivate(sim)
					cat.OmenOfClarityAura.Icd.Set(0)
					cat.FaerieFire.CalcAndDealOutcome(sim, cat.CurrentTarget, cat.FaerieFire.OutcomeAlwaysHit)
				}})
			}
			for !sim.Step() {
			}

			if form == druid.Humanoid && procs == 0 {
				t.Fatal("caster Faerie Fire never triggered Omen")
			}
			if form != druid.Humanoid && procs != 0 {
				t.Fatalf("in-form Faerie Fire triggered Omen %d times", procs)
			}
		})
	}
}

func TestIdolOfTheDreamExtendsRipWithoutChangingItsTicks(t *testing.T) {
	row := spelldata.MustFind(9896)
	period := row.PeriodicDamageEffect().Period()
	extension := spelldata.MustFind(446212).EffectN(1).TimeValue()
	for _, item := range []int32{0, 220606} {
		for _, points := range []int32{4, 5} {
			t.Run(fmt.Sprintf("item=%d/points=%d", item, points), func(t *testing.T) {
				sim, cat := newCatWearing(t, furorTalents, nil, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotRanged: item})
				for iteration := range 2 {
					if iteration != 0 {
						sim.Cleanup()
						sim.Reset()
					}
					cat.AddComboPoints(sim, points, cat.NewComboPointMetrics(core.ActionID{SpellID: 9896}))
					dot := cat.Rip.Dot(cat.CurrentTarget)
					dot.Apply(sim)
					want := row.Duration()
					if item != 0 {
						want += extension
					}
					if dot.Duration != want || dot.TickPeriod() != period {
						t.Fatalf("Rip duration/period = %v/%v, want %v/%v", dot.Duration, dot.TickPeriod(), want, period)
					}
					base := row.PeriodicDamageEffect().AverageAt(60) + float64(row.PeriodicDamageEffect().PointsPerResource)*float64(points)
					if dot.SnapshotBaseDamage != base {
						t.Fatalf("Rip base tick damage = %v, want %v", dot.SnapshotBaseDamage, base)
					}
					if cat.Rake.Dot(cat.CurrentTarget).BaseDuration() != spelldata.MustFind(9904).Duration() {
						t.Fatal("Rip idol changed Rake duration")
					}
					wantTicks := int32(want / period)
					if iteration != 0 {
						for dot.TickCount() == 0 {
							if sim.Step() {
								t.Fatal("simulation ended before Rip ticked")
							}
						}
						dot.Apply(sim)
						if dot.Duration != want {
							t.Fatal("refresh changed Rip duration")
						}
						wantTicks++
					}
					for dot.IsActive() {
						if sim.Step() {
							t.Fatal("simulation ended before Rip expired")
						}
					}
					metrics := cat.Rip.SpellMetrics[cat.CurrentTarget.UnitIndex]
					ticks := metrics.Ticks + metrics.CritTicks
					if ticks != wantTicks {
						t.Fatalf("Rip dealt %d ticks, want %d", ticks, wantTicks)
					}
				}
			})
		}
	}
}

func TestCatChargeRunIn(t *testing.T) {
	sim, cat := newCat(t, withFeralDigit(powershiftTalents, 8, '1'), nil)
	for i := range 2 {
		if i > 0 {
			sim.Cleanup()
			sim.Reset()
		}
		cat.DistanceFromTarget = 25
		energy := cat.CurrentEnergy()
		cat.Prowl.Cast(sim, cat.CurrentTarget)
		cat.ClearcastingAura.Activate(sim)
		if !cat.FeralCharge.Cast(sim, cat.CurrentTarget) {
			t.Fatal("Cat Charge did not cast at 25 yards")
		}
		if !cat.ClearcastingAura.IsActive() {
			t.Error("Cat Charge consumed Clearcasting")
		}
		if cat.ProwlAura.IsActive() {
			t.Error("Charge did not break Prowl")
		}
		if cat.CurrentEnergy() != energy || cat.CurrentRage() != 0 {
			t.Error("Cat Charge spent resources")
		}
		if cat.FeralCharge.TimeToReady(sim) != 30*time.Second {
			t.Errorf("Cat Charge cooldown = %s, want 30s", cat.FeralCharge.TimeToReady(sim))
		}
		t.Logf("Cat Charge: GCD=%s cooldown=%s energy=%g Clearcasting=%t", cat.NextGCDAt()-sim.CurrentTime, cat.FeralCharge.TimeToReady(sim), cat.CurrentEnergy(), cat.ClearcastingAura.IsActive())
		if cat.NextGCDAt() != 1500*time.Millisecond {
			t.Errorf("GCD ends at %v, want 1.5s", cat.NextGCDAt())
		}
		for sim.CurrentTime < 312500*time.Microsecond {
			if sim.Step() {
				break
			}
		}
		if cat.DistanceFromTarget != 0 || cat.Moving {
			t.Errorf("landing: distance=%v moving=%t", cat.DistanceFromTarget, cat.Moving)
		}
		for cat.AutoAttacks.MHAuto().SpellMetrics[cat.CurrentTarget.UnitIndex].Casts == 0 {
			if sim.Step() {
				t.Fatal("simulation ended before the first auto-attack")
			}
		}
		if sim.CurrentTime < 462500*time.Microsecond || sim.CurrentTime > 1062500*time.Microsecond {
			t.Errorf("first auto-attack at %v, want travel plus 150–750ms", sim.CurrentTime)
		}
		for !sim.Step() {
		}
	}
	sim.Cleanup()
	sim.Reset()
	cat.DistanceFromTarget = 25
	cat.MoveTo(5, sim)
	sim.AddPendingAction(&core.PendingAction{NextActionAt: 100 * time.Millisecond, OnAction: func(*core.Simulation) {}})
	for sim.CurrentTime < 100*time.Millisecond {
		if sim.Step() {
			t.Fatal("simulation ended during the approach")
		}
	}
	if !cat.FeralCharge.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Cat Charge did not cast while running")
	}
	for cat.Moving {
		if sim.Step() {
			t.Fatal("simulation ended during Charge")
		}
	}
	if cat.DistanceFromTarget != 0 {
		t.Fatalf("Charge left the running Cat at %g yards", cat.DistanceFromTarget)
	}
}

func TestProwlSpeedChangePreservesMovementDestination(t *testing.T) {
	sim, cat := newCat(t, "0000000000000000-00200000000000000000-0000000000000000", nil)
	cat.DistanceFromTarget = 25
	sim.CurrentTime = -time.Second
	if !cat.Prowl.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Prowl did not cast")
	}
	sim.CurrentTime = 0
	if !cat.Dash.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Dash did not cast")
	}
	if !cat.ProwlAura.IsActive() || !cat.CatFormAura.IsActive() {
		t.Fatal("Dash removed Prowl or Cat Form")
	}
	cat.MoveTo(4, sim)
	for sim.CurrentTime < time.Second {
		if sim.Step() {
			t.Fatal("simulation ended before Prowl expired")
		}
	}
	cat.ProwlAura.Deactivate(sim)
	for cat.Moving {
		if sim.Step() {
			t.Fatal("simulation ended before arrival")
		}
	}
	if math.Abs(cat.DistanceFromTarget-4) > 1e-6 {
		t.Fatalf("arrived at %g yards, want 4", cat.DistanceFromTarget)
	}
}

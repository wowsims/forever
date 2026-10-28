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

// 16 balance, 20 feral, 16 restoration. Furor is the second restoration talent.
const furorTalents = "0000000000000000-00000000000000000000-0500000000000000"

// Shifting Power is the 11th feral talent and Improved Shifting Power the 15th. Natural Shapeshifter is
// the 5th restoration talent.
const (
	shiftingPowerTalents         = "0000000000000000-00000000001000000000-0000000000000000"
	improvedShiftingPowerTalents = "0000000000000000-00000000001000200000-0000000000000000"
	shapeshifterTalents          = "0000000000000000-00000000001000000000-0000300000000000"
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
	_, base := newCat(t, "0000000000000000-00000000000100000000-0000000000000000", nil)
	_, talented := newCat(t, "0000000000000000-50000030000100000000-0000000000000000", nil)
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
	mh, _, _ := cat.AutoAttacks.SwingAt()
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
	if mh, _, _ := cat.AutoAttacks.SwingAt(); mh != 0 {
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
			if mh, _, _ := cat.AutoAttacks.SwingAt(); mh != 3*time.Second {
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
			for cast := 0; cast < 2; cast++ {
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
	if got := cat.ShiftingPower.ReadyAt(); got != 16*time.Second {
		t.Errorf("shifting power ready at %v, want 16s", got)
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

func TestImprovedShiftingPowerTakesTheCooldownToEightSeconds(t *testing.T) {
	sim, cat := newCat(t, improvedShiftingPowerTalents, nil)
	fillResource(sim, cat)
	emptyEnergy(sim, cat)
	if !cat.ShiftingPower.Cast(sim, cat.CurrentTarget) {
		t.Fatal("shifting power did not cast")
	}
	if got := cat.ShiftingPower.ReadyAt(); got != 8*time.Second {
		t.Errorf("shifting power ready at %v, want 8s", got)
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

// Subtlety 3/3 is the 4th restoration talent. Its mask is Nature and Arcane, so physical threat stays put.
func TestSubtletyThreatIsNatureAndArcane(t *testing.T) {
	_, plain := newCat(t, "", nil)
	_, cat := newCat(t, "0000000000000000-00000000000000000000-0003000000000000", nil)
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

// Shadowmeld 20580 lets a Night Elf Ravage in combat; the Ravage ends it, and in combat its cooldown is
// 2 minutes.
func TestShadowmeldAllowsRavageInCombat(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	sim.CurrentTime = 10 * time.Second
	fillResource(sim, cat)
	if cat.Ravage.CanCast(sim, cat.CurrentTarget) {
		t.Fatal("ravage could be cast without stealth")
	}
	if !cat.Shadowmeld.Cast(sim, cat.CurrentTarget) {
		t.Fatal("shadowmeld did not cast in combat")
	}
	if !cat.Ravage.Cast(sim, cat.CurrentTarget) {
		t.Fatal("ravage did not cast from shadowmeld")
	}
	if cat.ShadowmeldAura.IsActive() {
		t.Error("ravage left the druid in shadowmeld")
	}
	if got := cat.Shadowmeld.ReadyAt() - sim.CurrentTime; got != 2*time.Minute {
		t.Errorf("shadowmeld is ready in %v, want 2m", got)
	}
}

// With Clearcasting at the Pull, the cat starts the fight with Clearcasting up.
func TestPrepullClearcastingStartsTheFightClearcasting(t *testing.T) {
	for _, on := range []bool{false, true} {
		req := sweepRequest(catTalents, catAPL(bestCatAPL), false, 1)
		req.SimOptions.Debug = true
		req.Raid.Parties[0].Players[0].GetFeralCatDruid().Options.PrepullClearcasting = on
		result := core.RunSim(req, nil, simsignals.CreateSignals())
		if result.Error != nil {
			t.Fatal(result.Error.Message)
		}
		got := false
		for _, line := range strings.Split(result.Logs, "\n") {
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
					sim, cat := newCat(t, fmt.Sprintf("0000000000000000-00000000000000000000-0%d00000000000000", rank), nil)
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
			for i := 0; i < 400; i++ {
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

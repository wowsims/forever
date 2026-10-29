//go:build with_db

package feralcat

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func reviewProwl(t *testing.T, sim *core.Simulation, cat *FeralDruid) {
	t.Helper()
	sim.CurrentTime = -time.Second
	if !cat.Prowl.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Prowl failed before combat")
	}
	sim.CurrentTime = time.Second
	fillResource(sim, cat)
}

func TestReviewProwlAndFeralSwiftnessSpeed(t *testing.T) {
	_, base := newCat(t, "", nil)
	baseSpeed := base.GetMovementSpeed()
	sim, cat := newCat(t, "0000000000000000-00200000000000000000-0000000000000000", nil)
	check := func(want float64) {
		t.Helper()
		if got := cat.GetMovementSpeed(); math.Abs(got-want) > 1e-9 {
			t.Fatalf("movement speed %g, want %g", got, want)
		}
	}
	check(baseSpeed * 1.3)
	reviewProwl(t, sim, cat)
	check(baseSpeed * 1.3 * 0.7)
	if !cat.Dash.Cast(sim, cat.CurrentTarget) || !cat.ProwlAura.IsActive() {
		t.Fatal("Dash failed or broke Prowl")
	}
	check(baseSpeed * 1.6 * 0.7)
	sim.CurrentTime += 2 * time.Second
	if !cat.Claw.Cast(sim, cat.CurrentTarget) || cat.ProwlAura.IsActive() {
		t.Fatal("Claw failed or retained Prowl")
	}
	check(baseSpeed * 1.6)
	cat.ClearForm(sim)
	check(baseSpeed)
	sim.CurrentTime += 2 * time.Second
	if !cat.CatForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("return to Cat Form failed")
	}
	check(baseSpeed * 1.6)
	cat.DashAura.Deactivate(sim)
	check(baseSpeed * 1.3)
	t.Logf("cat=%g yd/s, prowling cat=%g yd/s, humanoid=%g yd/s; re-entry restored cat speed", baseSpeed*1.3, baseSpeed*1.3*0.7, baseSpeed)
}

func TestReviewCatChargeBreaksProwlWithoutDebuff(t *testing.T) {
	sim, cat := newCat(t, "0000000000000000-00000000100000000000-0000000000000000", nil)
	reviewProwl(t, sim, cat)
	cat.DistanceFromTarget = 10
	if !cat.FeralCharge.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Cat Charge failed")
	}
	if cat.ProwlAura.IsActive() {
		t.Fatal("Cat Charge retained Prowl")
	}
	for _, array := range cat.FeralCharge.RelatedAuraArrays {
		if array.Get(cat.CurrentTarget).IsActive() {
			t.Fatal("Cat Charge applied a bear debuff")
		}
	}
	if mh, _, _ := cat.AutoAttacks.SwingAt(); mh != sim.CurrentTime {
		t.Fatalf("melee resumes at %v, want %v", mh, sim.CurrentTime)
	}
	t.Log("Cat Charge ended Prowl, resumed melee, and applied no root or daze")
}

func TestReviewProwlIgnoresFailedCastsAndTriggeredSpells(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	reviewProwl(t, sim, cat)
	cat.SpendEnergy(sim, cat.CurrentEnergy(), cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	if cat.Claw.Cast(sim, cat.CurrentTarget) || !cat.ProwlAura.IsActive() {
		t.Fatal("unaffordable Claw cast or broke Prowl")
	}
	if !cat.PounceBleed.Cast(sim, cat.CurrentTarget) || !cat.ProwlAura.IsActive() {
		t.Fatal("triggered Pounce bleed failed or broke Prowl")
	}
	if !cat.AutoAttacks.MHAuto().Cast(sim, cat.CurrentTarget) || !cat.ProwlAura.IsActive() {
		t.Fatal("automatic swing failed or was treated as a Prowl-breaking cast")
	}
	t.Log("failed cast and triggered bleed retained Prowl")
}

func TestReviewConsumableBreaksProwl(t *testing.T) {
	sim, cat := newCat(t, "", &proto.ConsumesSpec{ResourcePotId: 13444})
	reviewProwl(t, sim, cat)
	potion := cat.GetSpell(core.ActionID{ItemID: 13444})
	if potion == nil || !potion.Cast(sim, &cat.Unit) {
		t.Fatal("Major Mana Potion failed")
	}
	if cat.ProwlAura.IsActive() {
		t.Fatal("Major Mana Potion retained Prowl")
	}
	t.Log("item-ID consumable ended Prowl")
}

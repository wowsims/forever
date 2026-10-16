//go:build with_db

package feralcat

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/druid"
)

func TestBiteDodgeConsumesEnergyAndKeepsComboPoints(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	target := cat.CurrentTarget
	target.PseudoStats.BaseDodgeChance = 1
	cat.AddComboPoints(sim, 5, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
	if !cat.FerociousBite.Cast(sim, target) {
		t.Fatal("bite did not cast")
	}
	if cat.FerociousBite.SpellMetrics[target.UnitIndex].Dodges != 1 {
		t.Fatal("bite was not dodged")
	}
	if cat.CurrentEnergy() != 0 || cat.ComboPoints() != 5 {
		t.Fatalf("dodged bite leaves energy=%g CP=%d, want 0 and 5", cat.CurrentEnergy(), cat.ComboPoints())
	}
	t.Logf("dodged bite: energy=%g CP=%d", cat.CurrentEnergy(), cat.ComboPoints())
}

// Ravage 9867 and Pounce 9827 state No Active Defense: a target that dodges everything does not dodge them.
func TestRavageAndPounceCannotBeDodged(t *testing.T) {
	for _, opener := range []string{"Ravage", "Pounce"} {
		sim, cat := newCat(t, "", nil)
		target := cat.CurrentTarget
		target.PseudoStats.BaseDodgeChance = 1
		cat.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
		cat.ProwlAura.Activate(sim)
		spell := cat.Ravage
		if opener == "Pounce" {
			spell = cat.Pounce
		}
		if !spell.Cast(sim, target) {
			t.Fatalf("%s did not cast", opener)
		}
		if m := spell.SpellMetrics[target.UnitIndex]; m.Dodges != 0 || m.Hits+m.Crits != 1 {
			t.Errorf("%s against a target with 100%% dodge: %d dodges, %d hits and crits", opener, m.Dodges, m.Hits+m.Crits)
		}
	}
}

func TestRakeTickScalesWithAttackPower(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	cat.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
	dot := cat.Rake.Dot(cat.CurrentTarget)
	dot.Apply(sim)
	metric := &cat.Rake.SpellMetrics[cat.CurrentTarget.UnitIndex]
	before := metric.TotalDamage
	dot.TickOnce(sim)
	low := metric.TotalDamage - before
	cat.AddStatDynamic(sim, stats.AttackPower, 100)
	before = metric.TotalDamage
	dot.TickOnce(sim)
	high := metric.TotalDamage - before
	if math.Abs((high-low)-5.5) > 1e-5 {
		t.Fatalf("100 AP tick gain=%g, want 5.5", high-low)
	}
	t.Logf("Rake 100 AP gain per tick=%g", high-low)
}

// Neither Rake's initial hit nor its ticks are reduced by armor in game.
func TestRakeIgnoresArmor(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	target := cat.CurrentTarget
	hit := cat.Rake.CalcDamage(sim, target, 100, cat.Rake.OutcomeAlwaysHit).Damage
	tick := cat.Rake.CalcPeriodicDamage(sim, target, 100, cat.Rake.OutcomeAlwaysHit).Damage
	target.AddStatDynamic(sim, stats.Armor, 10000)
	armoredHit := cat.Rake.CalcDamage(sim, target, 100, cat.Rake.OutcomeAlwaysHit).Damage
	armoredTick := cat.Rake.CalcPeriodicDamage(sim, target, 100, cat.Rake.OutcomeAlwaysHit).Damage
	if math.Abs(hit-armoredHit) > 1e-6 || math.Abs(tick-armoredTick) > 1e-6 {
		t.Errorf("Rake hit %g -> %g and tick %g -> %g after 10000 armor, want unchanged", hit, armoredHit, tick, armoredTick)
	}
}

func TestFurorLeavingAtZeroTimeWithZeroEnergy(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	cat.SpendEnergy(sim, cat.CurrentEnergy(), cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	if !cat.DireBearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("bear cast failed")
	}
	sim.CurrentTime = 3 * time.Second
	fillResource(sim, cat)
	if !cat.CatForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("cat cast failed")
	}
	if cat.CurrentEnergy() != 0 {
		t.Fatalf("zero energy stored at t=0 restored %g", cat.CurrentEnergy())
	}
}

func TestRipAttackPowerCapsAtFourComboPoints(t *testing.T) {
	for cp := int32(1); cp <= 5; cp++ {
		sim, cat := newCat(t, "", nil)
		cat.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
		cat.AddComboPoints(sim, cp, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
		dot := cat.Rip.Dot(cat.CurrentTarget)
		dot.Apply(sim)
		metric := &cat.Rip.SpellMetrics[cat.CurrentTarget.UnitIndex]
		before := metric.TotalDamage
		dot.TickOnce(sim)
		low := metric.TotalDamage - before
		cat.AddStatDynamic(sim, stats.AttackPower, 100)
		before = metric.TotalDamage
		dot.TickOnce(sim)
		gain := metric.TotalDamage - before - low
		if math.Abs(gain-float64(min(cp, 4))) > 1e-5 {
			t.Fatalf("cp=%d tick gain=%g", cp, gain)
		}
		t.Logf("Rip %d CP: 100 AP gains %g per tick / %g total", cp, gain, gain*6)
	}
}

func TestCatShippedPresetRuns(t *testing.T) {
	req := catPowershiftRequest(powershiftTalents, core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "default").Rotation, false)
	req.Raid.Parties[0].Players[0].Equipment = core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "default").GearSet
	player := runCat(t, req)
	for _, id := range []int32{9830, 9904, 9896, 1322605} {
		if castsOf(player.Actions, id) == 0 {
			t.Fatalf("shipped preset never casts %d", id)
		}
	}
	if n := castsOf(player.Actions, 22812); n != 0 {
		t.Fatalf("shipped preset casts Barkskin %d times", n)
	}
}

func TestCatStealthPresetsRun(t *testing.T) {
	for name, opener := range map[string]int32{"ravage": 9867, "pounce": 9827} {
		t.Run(name, func(t *testing.T) {
			req := catPowershiftRequest(powershiftTalents, core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", name).Rotation, false)
			req.Raid.Parties[0].Players[0].Equipment = core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "default").GearSet
			player := runCat(t, req)
			if castsOf(player.Actions, opener) != 1 || castsOf(player.Actions, 9830) == 0 {
				t.Fatalf("%s must open once then continue with Shred", name)
			}
		})
	}
}

// Rend and Tear raises Shred on a bleeding target, by the same amount under each of the cat's own
// bleeds. With the bleed assumed, a target with no bleed of the cat's own takes the same raised Shred.
func TestAssumedBleedRaisesShredLikeOwnBleed(t *testing.T) {
	const rendAndTear5 = "0000000000000000-00000000000000000050-0000000000000000"
	shred := func(assume bool, bleed func(*FeralDruid) *druid.DruidSpell) float64 {
		sim, cat := newCat(t, rendAndTear5, nil)
		cat.AssumeBleedActive = assume
		target := cat.CurrentTarget
		cat.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
		cat.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
		target.PseudoStats.BaseDodgeChance = 0
		if bleed != nil {
			bleed(cat).Dot(target).Apply(sim)
		}
		fillResource(sim, cat)
		if !cat.Shred.Cast(sim, target) {
			t.Fatal("shred did not cast")
		}
		return cat.Shred.SpellMetrics[target.UnitIndex].TotalDamage
	}
	plain := shred(false, nil)
	bleeding := shred(false, func(cat *FeralDruid) *druid.DruidSpell { return cat.Rake })
	if plain <= 0 || bleeding <= plain*1.01 {
		t.Fatalf("shred %g without a bleed, %g with Rake up: Rend and Tear did not apply", plain, bleeding)
	}
	for name, bleed := range map[string]func(*FeralDruid) *druid.DruidSpell{
		"Rip":          func(cat *FeralDruid) *druid.DruidSpell { return cat.Rip },
		"Lacerate":     func(cat *FeralDruid) *druid.DruidSpell { return cat.Lacerate },
		"Pounce Bleed": func(cat *FeralDruid) *druid.DruidSpell { return cat.PounceBleed },
	} {
		if got := shred(false, bleed); math.Abs(got-bleeding) > 1e-6 {
			t.Errorf("shred with %s up = %g, with Rake up = %g", name, got, bleeding)
		}
	}
	if assumed := shred(true, nil); math.Abs(assumed-bleeding) > 1e-6 {
		t.Fatalf("shred with the bleed assumed = %g, with Rake up = %g", assumed, bleeding)
	}
}

// Rend and Tear raises bleed ticks. A Rake or Rip tick counts its own bleed as the bleed on the target,
// and the talent is the only difference between the two builds.
func TestRendAndTearRaisesBleedTicks(t *testing.T) {
	const rendAndTear5 = "0000000000000000-00000000000000000050-0000000000000000"
	tick := func(talents string, rip bool) float64 {
		sim, cat := newCat(t, talents, nil)
		cat.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
		spell := cat.Rake
		if rip {
			cat.AddComboPoints(sim, 5, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
			spell = cat.Rip
		}
		dot := spell.Dot(cat.CurrentTarget)
		dot.Apply(sim)
		metric := &spell.SpellMetrics[cat.CurrentTarget.UnitIndex]
		before := metric.TotalDamage
		dot.TickOnce(sim)
		return metric.TotalDamage - before
	}
	for _, rip := range []bool{false, true} {
		plain, raised := tick("", rip), tick(rendAndTear5, rip)
		if plain <= 0 || math.Abs(raised/plain-1.10) > 1e-6 {
			t.Errorf("rip=%v: tick %g without the talent, %g with 5/5, want x1.10", rip, plain, raised)
		}
	}
}

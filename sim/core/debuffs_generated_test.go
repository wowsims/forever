package core_test

import (
	"math"
	"os"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// The raid config registers its debuffs on every enemy and each is permanent,
// so the fight starts with all of them activated; the reset that does it in the
// sim needs a whole environment, which this stands in for.
func applyGeneratedTestDebuffs(target *core.Unit, debuffs *proto.Debuffs) *core.Simulation {
	sim := &core.Simulation{}
	core.ApplyDebuffEffects(target, debuffs, &proto.Raid{})
	for _, aura := range target.GetAuras() {
		aura.Activate(sim)
	}
	return sim
}

// The six schools the game files' mask 126 names: every school but physical.
var magicSchools = []stats.SchoolIndex{
	stats.SchoolIndexArcane, stats.SchoolIndexFire, stats.SchoolIndexFrost,
	stats.SchoolIndexHoly, stats.SchoolIndexNature, stats.SchoolIndexShadow,
}

func TestGeneratedCurseOfTheElementsHitsEverySchoolAndResistance(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()

	applyGeneratedTestDebuffs(target, &proto.Debuffs{CurseOfElements: true})

	for _, school := range magicSchools {
		if got := target.PseudoStats.SchoolDamageTakenMultiplier[school]; got != 1.1 {
			t.Errorf("school %d takes %v times the damage, want the game files' 1.1", school, got)
		}
	}
	if got := target.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexPhysical]; got != 1 {
		t.Errorf("physical damage taken is %v, want the game files' mask to have left it alone", got)
	}

	for _, stat := range []stats.Stat{
		stats.ArcaneResistance, stats.FireResistance, stats.FrostResistance,
		stats.NatureResistance, stats.ShadowResistance,
	} {
		if got := target.GetStats()[stat]; got != -75 {
			t.Errorf("%s is %v, want the game files' -75", stat.StatName(), got)
		}
	}

	aura := target.GetAura("Curse of the Elements (External)")
	if aura == nil {
		t.Fatalf("no aura is labelled %q; the target has %v", "Curse of the Elements (External)", targetAuraLabels(target))
	}
	if want := (core.ActionID{SpellID: 1311680, Tag: -1}); aura.ActionID != want {
		t.Errorf("the curse is %v, want %v", aura.ActionID, want)
	}
	if aura.Duration != core.NeverExpires {
		t.Errorf("the curse lasts %v, want the raid config's copy to be up all fight", aura.Duration)
	}
	if got := buffs.CurseOfElementsDuration(0); got != 5*time.Minute {
		t.Errorf("the curse is %v long before it is made permanent, want the game files' 5 minutes", got)
	}
}

// Expose Armor states its armor per combo point, and the raid config's debuff
// is the five-point finisher.
func TestGeneratedExposeArmorIsTheFivePointFinisher(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()

	applyGeneratedTestDebuffs(target, &proto.Debuffs{ExposeArmor: true})

	if got := target.GetStats()[stats.Armor]; got != -2250 {
		t.Errorf("armor is %v, want five times the game files' -450", got)
	}

	category := target.ExclusiveEffectManager.GetExclusiveEffectCategory(
		core.ExclusiveStatCategory(buffs.ExposeArmorCategory, stats.Armor.StatName(), false))
	if category.SingleAura {
		t.Error("the armor category is single-aura, so a stronger armor reduction would push this one off the target")
	}
	if len(category.Effects()) != 1 || category.Effects()[0].Priority != 2250 {
		t.Errorf("the category holds %d effects, first bid %v; want one bidding the magnitude 2250",
			len(category.Effects()), category.Effects()[0].Priority)
	}
}

// A stacking debuff is worth its amount once per stack, and it bids the same
// way, so that a debuff worth more than five stacks of it takes the slot.
func TestGeneratedSunderArmorPricesEveryStack(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()

	sim := applyGeneratedTestDebuffs(target, &proto.Debuffs{SunderArmor: true})

	sunder := target.GetAura("Sunder Armor (External)")
	if sunder == nil {
		t.Fatalf("no aura is labelled %q; the target has %v", "Sunder Armor (External)", targetAuraLabels(target))
	}
	if got := sunder.ActionID.Tag; got != -1 {
		t.Errorf("with no warrior in the raid, the raid's Sunder is tagged %d, want the external copy's -1", got)
	}
	if got := target.GetStats()[stats.Armor]; got != 0 {
		t.Errorf("a sunder with no stacks reduced armor by %v, want nothing", got)
	}

	sunder.SetStacks(sim, 5)

	if got := target.GetStats()[stats.Armor]; got != -2250 {
		t.Errorf("armor is %v, want five stacks of the game files' -450", got)
	}
	category := target.ExclusiveEffectManager.GetExclusiveEffectCategory(
		core.ExclusiveStatCategory(buffs.SunderArmorCategory, stats.Armor.StatName(), false))
	if len(category.Effects()) != 1 || category.Effects()[0].Priority != 2250 {
		t.Errorf("the category holds %d effects, first bid %v; want one bidding the magnitude 2250",
			len(category.Effects()), category.Effects()[0].Priority)
	}
}

// The strongest armor reduction applies and the weaker one stays on the target
// beside it; when the strongest ends, the weaker one's armor applies.
func TestGeneratedExposeArmorOutbidsAWeakerSunder(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()

	sim := applyGeneratedTestDebuffs(target, &proto.Debuffs{SunderArmor: true})
	sunder := target.GetAura("Sunder Armor (External)")
	sunder.SetStacks(sim, 1)
	if got := target.GetStats()[stats.Armor]; got != -450 {
		t.Fatalf("one stack of sunder reduced armor by %v, want the game files' -450", got)
	}

	expose := core.MakePermanent(buffs.NewSpellDataExposeArmorAura(target, false, 0))
	expose.Activate(sim)

	if got := target.GetStats()[stats.Armor]; got != -2250 {
		t.Errorf("armor is %v, want only Expose Armor's -2250", got)
	}
	if !sunder.IsActive() {
		t.Error("the weaker armor reduction was pushed off the target, want it beside Expose Armor")
	}

	expose.Deactivate(sim)
	if got := target.GetStats()[stats.Armor]; got != -450 {
		t.Errorf("with Expose Armor off, armor is %v, want the sunder's -450", got)
	}
}

// With both in the raid config, Expose Armor is worth its full -2250 from the
// first moment and the raid's Sunder Armor lands beside it worth nothing. The
// sunder's stacks build while Expose Armor holds the category, and apply once
// Expose Armor ends.
func TestGeneratedRaidsSunderStacksUnderExposeArmor(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()

	sim := applyGeneratedTestDebuffs(target, &proto.Debuffs{ExposeArmor: true, SunderArmor: true})
	expose := target.GetAura("Expose Armor (External)")
	sunder := target.GetAura("Sunder Armor (External)")
	if !expose.IsActive() || !sunder.IsActive() {
		t.Fatalf("expose active %v, sunder active %v; want both on the target", expose.IsActive(), sunder.IsActive())
	}

	sunder.SetStacks(sim, 3)
	if got := target.GetStats()[stats.Armor]; got != -2250 {
		t.Errorf("armor is %v, want Expose Armor's -2250 once", got)
	}

	expose.Deactivate(sim)
	if got := target.GetStats()[stats.Armor]; got != -1350 {
		t.Errorf("with Expose Armor off, armor is %v, want the sunder's three stacks of -450", got)
	}
}

// Every debuff in a shared category is on the target, but only the one holding the stat's slot in it
// applies its reduction; taking that one off hands the reduction to the next.
func checkOneReductionApplies(t *testing.T, debuffs *proto.Debuffs, category string, stat stats.Stat, want float64, labels ...string) {
	t.Helper()
	target := core.NewGeneratedDebuffTestTarget()
	base := target.GetStats()[stat]

	sim := applyGeneratedTestDebuffs(target, debuffs)

	for _, label := range labels {
		aura := target.GetAura(label)
		if aura == nil {
			t.Fatalf("no aura is labelled %q; the target has %v", label, targetAuraLabels(target))
		}
		if !aura.IsActive() {
			t.Errorf("%s is not on the target, want every debuff in the category applied", label)
		}
	}
	if got := target.GetStats()[stat] - base; got != want {
		t.Errorf("%s moved by %v, want one reduction of the game files' %v", stat.StatName(), got, want)
	}

	slot := target.ExclusiveEffectManager.GetExclusiveEffectCategory(core.ExclusiveStatCategory(category, stat.StatName(), false))
	holder := slot.GetActiveAura()
	if holder == nil {
		t.Fatalf("no debuff holds the %s category", category)
	}
	holder.Deactivate(sim)
	if next := slot.GetActiveAura(); next == nil || next == holder {
		t.Fatalf("with %s off, %v holds the %s category, want the other debuff", holder.Label, next, category)
	}
	if got := target.GetStats()[stat] - base; got != want {
		t.Errorf("with %s off, %s moved by %v, want the other debuff's %v", holder.Label, stat.StatName(), got, want)
	}
}

// Curse of Recklessness and Faerie Fire share one armor reduction slot: both are on the target, and
// it takes one -505, not both.
func TestGeneratedCurseOfRecklessnessAndFaerieFireDoNotStack(t *testing.T) {
	checkOneReductionApplies(t, &proto.Debuffs{CurseOfRecklessness: true, FaerieFire: true},
		buffs.CurseOfRecklessnessCategory, stats.Armor, -505,
		"Curse of Recklessness (External)", "Faerie Fire (External)")
}

func TestGeneratedThunderClapSlowsTheTarget(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()

	applyGeneratedTestDebuffs(target, &proto.Debuffs{ThunderClap: true})

	// The game files' -20 is 20% more time between attacks, so the speed is divided by 1.2.
	want := core.SymmetricMultiplierFromPercent(-20)
	if got := buffs.ThunderClapValue(0); got != want {
		t.Errorf("the clap is worth %v, want the game files' -20%% as %v", got, want)
	}
	if got := target.PseudoStats.MeleeSpeedMultiplier; got != want {
		t.Errorf("the melee speed multiplier is %v, want %v", got, want)
	}
	if got := target.TotalMeleeHasteMultiplier(); got != want {
		t.Errorf("the target swings at %v times its speed, want the slow to have reached the swing timers", got)
	}
}

// Every member of the attack-speed category bids how far from 1 its multiplier
// is, so that the strongest slow on the target is the one that applies whatever
// form its source states it in.
func TestGeneratedThunderClapBidsAgainstTheHandWrittenSlows(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()
	sim := applyGeneratedTestDebuffs(target, &proto.Debuffs{ThunderClap: true})
	clap := target.GetAura("Thunder Clap (External)")

	// A hand-written slow states the multiplier the target's speed is divided
	// by, so 1.5 is a third slower and outbids the clap's sixth.
	slower := target.GetOrRegisterAura(core.Aura{
		Label:    "Hand-written Slow",
		ActionID: core.ActionID{SpellID: 27648},
		Duration: time.Second * 12,
	})
	// Through a variable, so that the test does the same float64 arithmetic the
	// helper does rather than Go's exact constant arithmetic.
	divisor := 1.5
	stronger := buffs.AtkSpeedReductionEffect(slower, divisor)
	if want := 1 - 1/divisor; stronger.Priority != want {
		t.Errorf("a slow that divides by %v bids %v, want the magnitude %v", divisor, stronger.Priority, want)
	}

	slower.Activate(sim)

	if got := target.PseudoStats.MeleeSpeedMultiplier; got != 1 {
		t.Errorf("the melee speed multiplier is %v, want the weaker slow taken back", got)
	}
	if got, want := target.PseudoStats.AttackSpeedMultiplier, 1/divisor; got != want {
		t.Errorf("the attack speed multiplier is %v, want the stronger slow's %v", got, want)
	}
	if !clap.IsActive() {
		t.Error("losing the category pushed the weaker aura off, which only a single-aura category may do")
	}
}

// Thunderfury's Cyclone 27648 and Thunder Clap 11581 both state -20, the same slow, so they bid
// the same. The raid's clap outlasts the proc's 12 seconds, so it keeps the category when the
// proc lands.
func TestGeneratedThunderClapOutbidsThunderfurysCyclone(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()
	sim := applyGeneratedTestDebuffs(target, &proto.Debuffs{ThunderClap: true})
	clap := target.GetAura("Thunder Clap (External)")

	cyclone := target.GetOrRegisterAura(core.Aura{
		Label:    "Cyclone",
		ActionID: core.ActionID{SpellID: 27648},
		Duration: time.Second * 12,
	})
	tied := buffs.AtkSpeedReductionEffect(cyclone, core.MultiplierFromPercent(20))
	if clapBid := clap.ExclusiveEffects[0].Priority; tied.Priority != clapBid {
		t.Fatalf("the proc bids %v against the clap's %v, want the same -20%% slow to tie", tied.Priority, clapBid)
	}

	cyclone.Activate(sim)

	want := core.SymmetricMultiplierFromPercent(-20)
	if got := target.PseudoStats.MeleeSpeedMultiplier; got != want {
		t.Errorf("the melee speed multiplier is %v, want the clap's %v to have stayed", got, want)
	}
	if got := target.PseudoStats.AttackSpeedMultiplier; got != 1 {
		t.Errorf("the attack speed multiplier is %v, want the tied proc to have applied nothing", got)
	}
	if got := target.TotalMeleeHasteMultiplier(); got != want {
		t.Errorf("the target swings at %v times its speed, want the clap's %v", got, want)
	}
}

// The raid's Thunderfury takes 25 nature resistance off the target and slows it.
func TestGeneratedThunderfuryLowersNatureResistanceAndSlows(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()
	base := target.GetStats()[stats.NatureResistance]
	applyGeneratedTestDebuffs(target, &proto.Debuffs{Thunderfury: true})
	if got := target.GetStats()[stats.NatureResistance] - base; got != -25 {
		t.Errorf("nature resistance moved by %v, want -25", got)
	}
	if got, want := target.TotalMeleeHasteMultiplier(), core.SymmetricMultiplierFromPercent(-20); got != want {
		t.Errorf("the target swings at %v, want the -20%% slow's %v", got, want)
	}
}

// Thunder Clap ties the slow, which must not take the resistance with it.
func TestGeneratedThunderfuryKeepsItsResistanceBesideThunderClap(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()
	base := target.GetStats()[stats.NatureResistance]
	applyGeneratedTestDebuffs(target, &proto.Debuffs{Thunderfury: true, ThunderClap: true})
	if got := target.GetStats()[stats.NatureResistance] - base; got != -25 {
		t.Errorf("nature resistance moved by %v, want -25 whatever the slow tie decides", got)
	}
	if got, want := target.TotalMeleeHasteMultiplier(), core.SymmetricMultiplierFromPercent(-20); got != want {
		t.Errorf("the target swings at %v, want one -20%% slow, %v", got, want)
	}
}

// Demoralizing Roar and Demoralizing Shout are the same category, so the attack
// power comes off the target once however many of them the raid brings.
func TestGeneratedDemoralizingDebuffsApplyOnce(t *testing.T) {
	checkOneReductionApplies(t, &proto.Debuffs{DemoralizingRoar: true, DemoralizingShout: true},
		buffs.DemoralizingShoutCategory, stats.AttackPower, -205,
		"Demoralizing Roar (External)", "Demoralizing Shout (External)")
}

// The two judgements are bare 40-second auras in the game files; what they do to
// whoever strikes the target is the driver's proc trigger.
func TestGeneratedJudgementsCarryTheDriversProcTriggers(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()

	applyGeneratedTestDebuffs(target, &proto.Debuffs{JudgementOfLight: true, JudgementOfWisdom: true})

	for label, want := range map[string]core.ActionID{
		"Judgement of Light (External)":  {SpellID: 20346, Tag: -1},
		"Judgement of Wisdom (External)": {SpellID: 20355, Tag: -1},
	} {
		aura := target.GetAura(label)
		if aura == nil {
			t.Fatalf("the target has %v, want an aura labelled %q", targetAuraLabels(target), label)
		}
		if aura.ActionID != want {
			t.Errorf("%s is %v, want %v", label, aura.ActionID, want)
		}
		if !aura.IsActive() {
			t.Errorf("%s is not up, want the raid config's copy to be permanent", label)
		}
		if aura.OnSpellHitTaken == nil {
			t.Errorf("%s has no proc trigger, so nothing happens to whoever strikes the target", label)
		}
	}

	if buffs.JudgementOfLightDuration(0) != time.Second*40 || buffs.JudgementOfWisdomDuration(0) != time.Second*40 {
		t.Errorf("the judgements last %v and %v before they are made permanent, want the game files' 40 seconds",
			buffs.JudgementOfLightDuration(0), buffs.JudgementOfWisdomDuration(0))
	}
}

// Every flat attack-power reduction bids in one category, so the target loses the strongest once.
//
// The game files' numbers: Hex of Weakness 19285 states -204 with no per-level gain past its BaseLevel
// 60. Demoralizing Screech 24579 states -198 with -1.6 per level past its BaseLevel 56, capped at
// MaxLevel 60, so a level 60 target reads -198 + 4*-1.6 = -204.4, floored to -205 - the same -205
// Demoralizing Shout already states.
func TestGeneratedAttackPowerReductionsApplyTheStrongestOnce(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()
	applyGeneratedTestDebuffs(target, &proto.Debuffs{
		DemoralizingShout: true, HexOfWeakness: true, DemoralizingScreech: true,
	})
	if got := buffs.HexOfWeaknessValue(0); got != -204 {
		t.Errorf("Hex of Weakness is worth %v, want the game files' -204", got)
	}
	if got := buffs.DemoralizingScreechValue(0); got != -205 {
		t.Errorf("Demoralizing Screech is worth %v, want the game files' -198 plus 4 levels of -1.6, floored to -205", got)
	}
	want := min(buffs.DemoralizingShoutValue(0), buffs.HexOfWeaknessValue(0), buffs.DemoralizingScreechValue(0))
	if got := target.GetStats()[stats.AttackPower]; got != want {
		t.Errorf("attack power is %v, want the strongest reduction %v once", got, want)
	}
}

// Without the shout, Demoralizing Screech's -205 outbids Hex of Weakness's -204, and the hex's
// applies once the screech ends.
func TestGeneratedDemoralizingScreechOutbidsHexOfWeakness(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()
	sim := applyGeneratedTestDebuffs(target, &proto.Debuffs{HexOfWeakness: true, DemoralizingScreech: true})
	if got := target.GetStats()[stats.AttackPower]; got != -205 {
		t.Errorf("attack power is %v, want the screech's -205 once", got)
	}

	hex := target.GetAura("Hex of Weakness (External)")
	screech := target.GetAura("Demoralizing Screech (External)")
	if hex == nil || screech == nil {
		t.Fatalf("the target has %v, want both attack power reductions", targetAuraLabels(target))
	}
	if !screech.IsActive() || !hex.IsActive() {
		t.Errorf("screech active %v, hex active %v; want both on the target", screech.IsActive(), hex.IsActive())
	}

	screech.Deactivate(sim)
	if got := target.GetStats()[stats.AttackPower]; got != -204 {
		t.Errorf("with the screech off, attack power is %v, want the hex's -204", got)
	}
}

// Hex of Weakness 19285 states -204 attack power and -20% healing taken, and each bids alone:
// Demoralizing Shout's -205 holds the attack power, and the healing taken, which the shout does not
// state, stays the hex's.
func TestGeneratedHexOfWeaknessKeepsItsHealingReductionUnderDemoralizingShout(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()
	sim := applyGeneratedTestDebuffs(target, &proto.Debuffs{DemoralizingShout: true, HexOfWeakness: true})

	hex := target.GetAura("Hex of Weakness (External)")
	shout := target.GetAura("Demoralizing Shout (External)")
	if hex == nil || shout == nil || !hex.IsActive() || !shout.IsActive() {
		t.Fatalf("the target has %v, want the hex and the shout both on it", targetAuraLabels(target))
	}
	if got := target.GetStats()[stats.AttackPower]; got != -205 {
		t.Errorf("attack power is %v, want the shout's -205 once", got)
	}
	if got := target.PseudoStats.HealingTakenMultiplier; got != 0.8 {
		t.Errorf("healing taken is %v times, want the hex's -20%% beside the shout", got)
	}

	shout.Deactivate(sim)
	if got := target.GetStats()[stats.AttackPower]; got != -204 {
		t.Errorf("with the shout off, attack power is %v, want the hex's -204", got)
	}
	if got := target.PseudoStats.HealingTakenMultiplier; got != 0.8 {
		t.Errorf("with the shout off, healing taken is %v times, want the hex's -20%% once", got)
	}
}

// Vindication 440667 states -204 attack power and -204 ranged attack power: Demoralizing Shout's -205
// holds the attack power, and the ranged attack power, which the shout does not state, stays
// Vindication's.
func TestGeneratedVindicationKeepsItsRangedAttackPowerUnderDemoralizingShout(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()
	applyGeneratedTestDebuffs(target, &proto.Debuffs{DemoralizingShout: true, Vindication: true})

	vindication := target.GetAura("Vindication (External)")
	shout := target.GetAura("Demoralizing Shout (External)")
	if vindication == nil || shout == nil || !vindication.IsActive() || !shout.IsActive() {
		t.Fatalf("the target has %v, want Vindication and the shout both on it", targetAuraLabels(target))
	}
	if got := target.GetStats()[stats.AttackPower]; got != -205 {
		t.Errorf("attack power is %v, want the shout's -205 once", got)
	}
	if got := target.GetStats()[stats.RangedAttackPower]; got != -204 {
		t.Errorf("ranged attack power is %v, want Vindication's -204 beside the shout", got)
	}
}

// Curse of Weakness lowers the damage each hit does, not attack power, so it stacks with the shouts.
func TestGeneratedCurseOfWeaknessStacksWithDemoralizingShout(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()
	applyGeneratedTestDebuffs(target, &proto.Debuffs{DemoralizingShout: true, CurseOfWeakness: true})
	if got := target.GetStats()[stats.AttackPower]; got != buffs.DemoralizingShoutValue(0) {
		t.Errorf("attack power is %v, want the shout's %v alone", got, buffs.DemoralizingShoutValue(0))
	}
	if got := target.GetStats()[stats.PhysicalDamage]; got != -37 {
		t.Errorf("physical damage is %v, want the game files' -37", got)
	}
}

// The game files state Dust Cloud at one stack, which the store reads as a stacking row, so the raid's
// copy takes its stack from a driver: without one, the amount would follow zero stacks and apply
// nothing.
func TestGeneratedDustCloudTakesItsStackFromTheDriver(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()
	base := target.GetStats()[stats.Armor]
	applyGeneratedTestDebuffs(target, &proto.Debuffs{DustCloud: true})
	if got := buffs.DustCloudValue(0); got != -505 {
		t.Errorf("Dust Cloud is worth %v, want the game files' -505", got)
	}
	if got := target.GetStats()[stats.Armor] - base; got != -505 {
		t.Errorf("armor moved by %v, want the game files' -505", got)
	}
	dust := target.GetAura("Dust Cloud (External)")
	if dust == nil {
		t.Fatalf("no aura is labelled %q; the target has %v", "Dust Cloud (External)", targetAuraLabels(target))
	}
	if got := dust.GetStacks(); got != 1 {
		t.Errorf("Dust Cloud sits at %d stacks, want the driver to have taken its one stack", got)
	}
}

// Dust Cloud shares the armor reduction slot of Faerie Fire and Curse of Recklessness: both are on the
// target, and it takes one -505, not both.
func TestGeneratedDustCloudAndFaerieFireDoNotStack(t *testing.T) {
	checkOneReductionApplies(t, &proto.Debuffs{FaerieFire: true, DustCloud: true},
		buffs.DustCloudCategory, stats.Armor, -505,
		"Dust Cloud (External)", "Faerie Fire (External)")
}

// Scorpid Sting and Insect Swarm share one slot: both are on the target, and its hit chance drops by
// one -2%, not both.
func TestGeneratedScorpidStingAndInsectSwarmDoNotStack(t *testing.T) {
	checkOneReductionApplies(t, &proto.Debuffs{ScorpidSting: true, InsectSwarm: true},
		buffs.ScorpidStingCategory, stats.PhysicalHitPercent, -2,
		"Scorpid Sting (External)", "Insect Swarm (External)")
}

// The raid's Annihilator keeps all three stacks up. Armor Shatter 16928 states -165 per stack and a
// MaxStack of 3.
func TestGeneratedArmorShatterHoldsThreeStacks(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()
	base := target.GetStats()[stats.Armor]
	applyGeneratedTestDebuffs(target, &proto.Debuffs{ArmorShatter: true})
	if got := buffs.ArmorShatterValue(0); got != -165 {
		t.Errorf("Armor Shatter is worth %v, want the game files' -165", got)
	}
	if got := target.GetStats()[stats.Armor] - base; got != -495 {
		t.Errorf("armor moved by %v, want 3 x -165", got)
	}
}

func targetAuraLabels(target *core.Unit) []string {
	labels := make([]string, 0, len(target.GetAuras()))
	for _, aura := range target.GetAuras() {
		labels = append(labels, aura.Label)
	}
	return labels
}

// The addStat call under the comment that names HuntersMarkValue.
var huntersMarkSheetRE = regexp.MustCompile(`(?s)HuntersMarkValue.*?addStat\(\s*Stat\.StatRangedAttackPower,\s*([0-9.]+)\s*\)`)

// Nothing exports a generated debuff's value to TypeScript, so the character sheet repeats
// Hunter's Mark's ranged attack power as a literal. The sheet and the sim state one number.
func TestHuntersMarkOnTheCharacterSheetMatchesTheSim(t *testing.T) {
	raw, err := os.ReadFile("../../ui/sim/player/player.ts")
	if err != nil {
		t.Fatalf("read player.ts: %v", err)
	}

	match := huntersMarkSheetRE.FindStringSubmatch(string(raw))
	if match == nil {
		t.Fatalf("player.ts credits no ranged attack power under the comment naming HuntersMarkValue; the pattern is %v", huntersMarkSheetRE)
	}

	sheet, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		t.Fatalf("player.ts credits %q ranged attack power: %v", match[1], err)
	}
	if want := buffs.HuntersMarkValue(0); sheet != want {
		t.Errorf("the sheet credits Hunter's Mark with %v ranged attack power, want the %v HuntersMarkValue states", sheet, want)
	}
}

// Vindication's talent rank prices it; the raid's copy is the full 3/3 amount.
func TestGeneratedVindicationScalesWithTheTalentRank(t *testing.T) {
	full := buffs.VindicationValue(0)
	if full >= 0 {
		t.Fatalf("the raid's Vindication is worth %v, want an attack power reduction", full)
	}
	if got := buffs.VindicationValue(3); got != full {
		t.Errorf("3/3 is worth %v, want the full %v", got, full)
	}
	if got, want := buffs.VindicationValue(1), math.Round(full/3); got != want {
		t.Errorf("1/3 is worth %v, want a third of %v rounded, %v", got, full, want)
	}
	target := core.NewGeneratedDebuffTestTarget()
	applyGeneratedTestDebuffs(target, &proto.Debuffs{Vindication: true})
	if got := target.GetStats()[stats.AttackPower]; got != full {
		t.Errorf("attack power is %v, want the raid's %v", got, full)
	}
	if got := target.GetStats()[stats.RangedAttackPower]; got != full {
		t.Errorf("ranged attack power is %v, want the raid's %v", got, full)
	}
}

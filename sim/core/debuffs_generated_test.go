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
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func rowEffect(t *testing.T, spellID int32, aura dbcenums.EffectAuraType) *spelldata.Effect {
	t.Helper()
	row := spelldata.MustFind(spellID)
	for i := range row.Effects {
		if row.Effects[i].Aura == aura {
			return &row.Effects[i]
		}
	}
	t.Fatalf("%s %d states no %v effect", row.Name, spellID, aura)
	return nil
}

// The amount the row's first effect applying aura states at the sim's character level. A test that
// reads one checks an amount, so a row stating none fails it.
func rowAmount(t *testing.T, spellID int32, aura dbcenums.EffectAuraType) float64 {
	t.Helper()
	amount := rowEffect(t, spellID, aura).Average()
	if amount == 0 {
		t.Fatalf("spell %d states no amount on its %v effect", spellID, aura)
	}
	return amount
}

// The amount the row's armor effect states at the sim's character level. The game files state armor
// either as a resistance on the physical school or as an armor aura that no armor modifier scales.
func armorAmount(t *testing.T, spellID int32) float64 {
	t.Helper()
	row := spelldata.MustFind(spellID)
	for i := range row.Effects {
		e := &row.Effects[i]
		if e.Aura == dbcenums.A_MOD_ARMOR_NO_MODS || e.Aura == dbcenums.A_MOD_RESISTANCE && e.Misc&1 != 0 {
			if amount := e.Average(); amount != 0 {
				return amount
			}
		}
	}
	t.Fatalf("%s %d states no armor amount", row.Name, spellID)
	return 0
}

// How long the row states its aura lasts. A test that reads one checks a duration, so a row stating
// none fails it.
func rowDuration(t *testing.T, spellID int32) time.Duration {
	t.Helper()
	duration := spelldata.MustFind(spellID).Duration()
	if duration <= 0 {
		t.Fatalf("spell %d states no duration", spellID)
	}
	return duration
}

// A stat amount at the sim's character level: the base plus the per-level gain past the BaseLevel, up to
// the MaxLevel (0 for none), in float32 the way the game files state the gain, floored to a whole number.
func levelScaledFloor(e *spelldata.Effect) float64 {
	level := int32(core.CharacterLevel)
	if e.MaxLevel > 0 {
		level = min(level, int32(e.MaxLevel))
	}
	levels := max(0, level-int32(e.BaseLevel))
	return math.Floor(float64(float32(e.BasePoints) + float32(levels)*float32(e.PPL)))
}

// Thunder Clap's slow as the game files state it, a percentage of attack speed.
func thunderClapSlowPercent(t *testing.T) float64 {
	t.Helper()
	return rowAmount(t, 11581, dbcenums.A_MOD_MELEE_HASTE_3)
}

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

	taken := core.MultiplierFromPercent(rowAmount(t, 1311680, dbcenums.A_MOD_DAMAGE_PERCENT_TAKEN))
	for _, school := range magicSchools {
		if got := target.PseudoStats.SchoolDamageTakenMultiplier[school]; got != taken {
			t.Errorf("school %d takes %v times the damage, want the game files' %v", school, got, taken)
		}
	}
	if got := target.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexPhysical]; got != 1 {
		t.Errorf("physical damage taken is %v, want the game files' mask to have left it alone", got)
	}

	resistance := rowAmount(t, 1311680, dbcenums.A_MOD_RESISTANCE)
	for _, stat := range []stats.Stat{
		stats.ArcaneResistance, stats.FireResistance, stats.FrostResistance,
		stats.NatureResistance, stats.ShadowResistance,
	} {
		if got, want := target.GetStats()[stat], resistance; got != want {
			t.Errorf("%s is %v, want the game files' %v", stat.StatName(), got, want)
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
}

// Expose Armor at the five combo points the raid config's debuff is cast with, from the armor per
// point the row states.
func exposeArmorAtFivePoints(t *testing.T) float64 {
	t.Helper()
	perPoint := float64(rowEffect(t, 11198, dbcenums.A_MOD_ARMOR_NO_MODS).PointsPerResource)
	if perPoint == 0 {
		t.Fatal("Expose Armor states no armor per combo point")
	}
	return 5 * perPoint
}

// Expose Armor states its armor per combo point, and the raid config's debuff
// is the five-point finisher.
func TestGeneratedExposeArmorIsTheFivePointFinisher(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()

	applyGeneratedTestDebuffs(target, &proto.Debuffs{ExposeArmor: true})

	expose := exposeArmorAtFivePoints(t)
	if got := buffs.ExposeArmorValue(0); got != expose {
		t.Errorf("Expose Armor is worth %v, want five combo points of the game files' armor, %v", got, expose)
	}
	if got := target.GetStats()[stats.ArmorNoMods]; got != expose {
		t.Errorf("armor is %v, want the five-point %v", got, expose)
	}

	category := target.ExclusiveEffectManager.GetExclusiveEffectCategory(
		core.ExclusiveStatCategory(buffs.ExposeArmorCategory, stats.ArmorNoMods.StatName(), false))
	if category.SingleAura {
		t.Error("the armor category is single-aura, so a stronger armor reduction would push this one off the target")
	}
	if bid := -expose; len(category.Effects()) != 1 || category.Effects()[0].Priority != bid {
		t.Errorf("the category holds %d effects, first bid %v; want one bidding the magnitude %v",
			len(category.Effects()), category.Effects()[0].Priority, bid)
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
	if got := target.GetStats()[stats.ArmorNoMods]; got != 0 {
		t.Errorf("a sunder with no stacks reduced armor by %v, want nothing", got)
	}

	sunder.SetStacks(sim, 5)

	perStack := armorAmount(t, 11597)
	fiveStacks := 5 * perStack
	if got := target.GetStats()[stats.ArmorNoMods]; got != fiveStacks {
		t.Errorf("armor is %v, want five stacks of the game files' %v", got, perStack)
	}
	category := target.ExclusiveEffectManager.GetExclusiveEffectCategory(
		core.ExclusiveStatCategory(buffs.SunderArmorCategory, stats.ArmorNoMods.StatName(), false))
	if len(category.Effects()) != 1 || category.Effects()[0].Priority != -fiveStacks {
		t.Errorf("the category holds %d effects, first bid %v; want one bidding the magnitude %v",
			len(category.Effects()), category.Effects()[0].Priority, -fiveStacks)
	}
}

// Five stacks of Sunder Armor take their full armor off a target whose armor a percent armor
// modifier scales, which scales only the target's own armor.
func TestGeneratedSunderArmorLandsOutsideTheTargetsArmorMultiplier(t *testing.T) {
	const armor, multiplier = 3731.0, 1.5
	target := core.NewTarget(&proto.Target{Level: 63, Stats: stats.Stats{stats.Armor: armor}.ToProtoArray()}, 0)
	target.Env = &core.Environment{MeasuringStats: true}
	target.MultiplyStat(stats.Armor, multiplier)

	sim := applyGeneratedTestDebuffs(&target.Unit, &proto.Debuffs{SunderArmor: true})
	target.GetAura("Sunder Armor (External)").SetStacks(sim, 5)

	fiveStacks := 5 * armorAmount(t, 11597)
	if got, want := target.SortAndApplyStatDependencies(target.GetStats())[stats.Armor], armor*multiplier+fiveStacks; got != want {
		t.Errorf("armor is %v, want the scaled %v and five whole stacks of %v, %v", got, armor*multiplier, fiveStacks/5, want)
	}
}

// The strongest armor reduction applies and the weaker one stays on the target
// beside it; when the strongest ends, the weaker one's armor applies.
func TestGeneratedExposeArmorOutbidsAWeakerSunder(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()

	sim := applyGeneratedTestDebuffs(target, &proto.Debuffs{SunderArmor: true})
	sunder := target.GetAura("Sunder Armor (External)")
	sunder.SetStacks(sim, 1)
	oneStack := armorAmount(t, 11597)
	if got := target.GetStats()[stats.ArmorNoMods]; got != oneStack {
		t.Fatalf("one stack of sunder reduced armor by %v, want the game files' %v", got, oneStack)
	}

	expose := core.MakePermanent(buffs.NewSpellDataExposeArmorAura(target, false, 0))
	expose.Activate(sim)

	if got, want := target.GetStats()[stats.ArmorNoMods], min(exposeArmorAtFivePoints(t), oneStack); got != want {
		t.Errorf("armor is %v, want only the stronger reduction's %v", got, want)
	}
	if !sunder.IsActive() {
		t.Error("the weaker armor reduction was pushed off the target, want it beside Expose Armor")
	}

	expose.Deactivate(sim)
	if got := target.GetStats()[stats.ArmorNoMods]; got != oneStack {
		t.Errorf("with Expose Armor off, armor is %v, want the sunder's %v", got, oneStack)
	}
}

// With both in the raid config, Expose Armor is worth its full five points from
// the first moment and the raid's Sunder Armor lands beside it worth nothing. The
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
	threeStacks := 3 * armorAmount(t, 11597)
	if got, want := target.GetStats()[stats.ArmorNoMods], min(exposeArmorAtFivePoints(t), threeStacks); got != want {
		t.Errorf("armor is %v, want the stronger reduction's %v once", got, want)
	}

	expose.Deactivate(sim)
	if got := target.GetStats()[stats.ArmorNoMods]; got != threeStacks {
		t.Errorf("with Expose Armor off, armor is %v, want the sunder's three stacks, %v", got, threeStacks)
	}
}

// Every debuff in a shared category is on the target, but only the one holding the stat's slot in it
// applies its reduction, the strongest; taking that one off hands the reduction to the next. values
// maps each debuff's label to its reduction.
func checkOneReductionApplies(t *testing.T, debuffs *proto.Debuffs, category string, stat stats.Stat, values map[string]float64) {
	t.Helper()
	target := core.NewGeneratedDebuffTestTarget()
	base := target.GetStats()[stat]

	sim := applyGeneratedTestDebuffs(target, debuffs)

	want := math.Inf(1)
	for label, value := range values {
		want = min(want, value)
		aura := target.GetAura(label)
		if aura == nil {
			t.Fatalf("no aura is labelled %q; the target has %v", label, targetAuraLabels(target))
		}
		if !aura.IsActive() {
			t.Errorf("%s is not on the target, want every debuff in the category applied", label)
		}
	}
	if got := target.GetStats()[stat] - base; got != want {
		t.Errorf("%s moved by %v, want one reduction of the strongest, %v", stat.StatName(), got, want)
	}

	slot := target.ExclusiveEffectManager.GetExclusiveEffectCategory(core.ExclusiveStatCategory(category, stat.StatName(), false))
	holder := slot.GetActiveAura()
	if holder == nil {
		t.Fatalf("no debuff holds the %s category", category)
	}
	holder.Deactivate(sim)
	next := slot.GetActiveAura()
	if next == nil || next == holder {
		t.Fatalf("with %s off, %v holds the %s category, want the other debuff", holder.Label, next, category)
	}
	if got, want := target.GetStats()[stat]-base, values[next.Label]; got != want {
		t.Errorf("with %s off, %s moved by %v, want %s's %v", holder.Label, stat.StatName(), got, next.Label, want)
	}
}

// Curse of Recklessness and Faerie Fire share one armor reduction slot: both are on the target, and
// it takes one reduction, not both.
func TestGeneratedCurseOfRecklessnessAndFaerieFireDoNotStack(t *testing.T) {
	checkOneReductionApplies(t, &proto.Debuffs{CurseOfRecklessness: true, FaerieFire: true},
		buffs.CurseOfRecklessnessCategory, stats.ArmorNoMods, map[string]float64{
			"Curse of Recklessness (External)": armorAmount(t, 11717),
			"Faerie Fire (External)":           armorAmount(t, 9907),
		})
}

func TestGeneratedThunderClapSlowsTheTarget(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()

	applyGeneratedTestDebuffs(target, &proto.Debuffs{ThunderClap: true})

	// The game files' negative percentage is that much more time between attacks: a percentage p
	// below 0 divides the speed by 1 - p/100.
	want := core.SymmetricMultiplierFromPercent(thunderClapSlowPercent(t))
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
	// by; this one divides by 0.3 more than the clap, so it outbids it.
	slower := target.GetOrRegisterAura(core.Aura{
		Label:    "Hand-written Slow",
		ActionID: core.ActionID{SpellID: 27648},
		Duration: time.Second * 12,
	})
	divisor := core.MultiplierFromPercent(-thunderClapSlowPercent(t)) + 0.3
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

// A Thunderfury Cyclone proc stating the clap's own slow bids the same as the clap. The raid's clap
// outlasts the proc's 12 seconds, so it keeps the category when the proc lands.
func TestGeneratedThunderClapOutbidsThunderfurysCyclone(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()
	sim := applyGeneratedTestDebuffs(target, &proto.Debuffs{ThunderClap: true})
	clap := target.GetAura("Thunder Clap (External)")

	cyclone := target.GetOrRegisterAura(core.Aura{
		Label:    "Cyclone",
		ActionID: core.ActionID{SpellID: 27648},
		Duration: time.Second * 12,
	})
	tied := buffs.AtkSpeedReductionEffect(cyclone, core.MultiplierFromPercent(-thunderClapSlowPercent(t)))
	if clapBid := clap.ExclusiveEffects[0].Priority; tied.Priority != clapBid {
		t.Fatalf("the proc bids %v against the clap's %v, want the same slow to tie", tied.Priority, clapBid)
	}

	cyclone.Activate(sim)

	want := core.SymmetricMultiplierFromPercent(thunderClapSlowPercent(t))
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

func thunderfurySlow(t *testing.T) float64 {
	t.Helper()
	return core.SymmetricMultiplierFromPercent(rowAmount(t, 27648, dbcenums.A_MOD_MELEE_HASTE_3))
}

// The raid's Thunderfury takes nature resistance off the target and slows it.
func TestGeneratedThunderfuryLowersNatureResistanceAndSlows(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()
	base := target.GetStats()[stats.NatureResistance]
	applyGeneratedTestDebuffs(target, &proto.Debuffs{Thunderfury: true})
	if got, want := target.GetStats()[stats.NatureResistance]-base, rowAmount(t, 21992, dbcenums.A_MOD_RESISTANCE); got != want {
		t.Errorf("nature resistance moved by %v, want %v", got, want)
	}
	if got, want := target.TotalMeleeHasteMultiplier(), thunderfurySlow(t); got != want {
		t.Errorf("the target swings at %v, want the proc's slow, %v", got, want)
	}
}

// Thunder Clap bids for the slow, which must not take the resistance with it.
func TestGeneratedThunderfuryKeepsItsResistanceBesideThunderClap(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()
	base := target.GetStats()[stats.NatureResistance]
	applyGeneratedTestDebuffs(target, &proto.Debuffs{Thunderfury: true, ThunderClap: true})
	if got, want := target.GetStats()[stats.NatureResistance]-base, rowAmount(t, 21992, dbcenums.A_MOD_RESISTANCE); got != want {
		t.Errorf("nature resistance moved by %v, want %v whatever the slow bid decides", got, want)
	}
	if got, want := target.TotalMeleeHasteMultiplier(), min(thunderfurySlow(t), core.SymmetricMultiplierFromPercent(thunderClapSlowPercent(t))); got != want {
		t.Errorf("the target swings at %v, want the stronger slow once, %v", got, want)
	}
}

// Demoralizing Roar and Demoralizing Shout are the same category, so the attack
// power comes off the target once however many of them the raid brings.
func TestGeneratedDemoralizingDebuffsApplyOnce(t *testing.T) {
	checkOneReductionApplies(t, &proto.Debuffs{DemoralizingRoar: true, DemoralizingShout: true},
		buffs.DemoralizingShoutCategory, stats.AttackPower, map[string]float64{
			"Demoralizing Roar (External)":  rowAmount(t, 9898, dbcenums.A_MOD_ATTACK_POWER),
			"Demoralizing Shout (External)": demoralizingShout(t),
		})
}

func demoralizingShout(t *testing.T) float64 {
	t.Helper()
	return rowAmount(t, 11556, dbcenums.A_MOD_ATTACK_POWER)
}

func hexOfWeakness(t *testing.T) float64 {
	t.Helper()
	return rowAmount(t, 19285, dbcenums.A_MOD_ATTACK_POWER)
}

// The two judgements are bare auras in the game files; what they do to whoever
// strikes the target is the driver's proc trigger.
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
}

// Every flat attack-power reduction bids in one category, so the target loses the strongest once.
// Demoralizing Screech states its amount with a per-level gain past its BaseLevel, up to its
// MaxLevel, and the stat it lands as is floored to a whole number.
func TestGeneratedAttackPowerReductionsApplyTheStrongestOnce(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()
	applyGeneratedTestDebuffs(target, &proto.Debuffs{
		DemoralizingShout: true, HexOfWeakness: true, DemoralizingScreech: true,
	})
	screech := demoralizingScreech(t)
	if got := buffs.DemoralizingScreechValue(0); got != screech {
		t.Errorf("Demoralizing Screech is worth %v, want the game files' amount at the character's level, floored to %v",
			got, screech)
	}
	want := min(demoralizingShout(t), hexOfWeakness(t), screech)
	if got := target.GetStats()[stats.AttackPower]; got != want {
		t.Errorf("attack power is %v, want the strongest reduction %v once", got, want)
	}
}

func demoralizingScreech(t *testing.T) float64 {
	t.Helper()
	screech := levelScaledFloor(rowEffect(t, 24579, dbcenums.A_MOD_ATTACK_POWER))
	if screech == 0 {
		t.Fatal("Demoralizing Screech states no attack power")
	}
	return screech
}

// Without the shout, Demoralizing Screech and Hex of Weakness share the attack power slot: both are
// on the target, the stronger applies, and the other's applies once that one ends.
func TestGeneratedDemoralizingScreechAndHexOfWeaknessDoNotStack(t *testing.T) {
	checkOneReductionApplies(t, &proto.Debuffs{HexOfWeakness: true, DemoralizingScreech: true},
		buffs.DemoralizingScreechCategory, stats.AttackPower, map[string]float64{
			"Hex of Weakness (External)":      hexOfWeakness(t),
			"Demoralizing Screech (External)": demoralizingScreech(t),
		})
}

// Hex of Weakness states attack power and healing taken, and each bids alone: the stronger of it and
// Demoralizing Shout holds the attack power, and the healing taken, which the shout does not state,
// stays the hex's.
func TestGeneratedHexOfWeaknessKeepsItsHealingReductionUnderDemoralizingShout(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()
	sim := applyGeneratedTestDebuffs(target, &proto.Debuffs{DemoralizingShout: true, HexOfWeakness: true})

	hex := target.GetAura("Hex of Weakness (External)")
	shout := target.GetAura("Demoralizing Shout (External)")
	if hex == nil || shout == nil || !hex.IsActive() || !shout.IsActive() {
		t.Fatalf("the target has %v, want the hex and the shout both on it", targetAuraLabels(target))
	}
	if got, want := target.GetStats()[stats.AttackPower], min(demoralizingShout(t), hexOfWeakness(t)); got != want {
		t.Errorf("attack power is %v, want the stronger reduction's %v once", got, want)
	}
	healing := core.MultiplierFromPercent(rowAmount(t, 19285, dbcenums.A_MOD_HEALING_PCT))
	if got := target.PseudoStats.HealingTakenMultiplier; got != healing {
		t.Errorf("healing taken is %v times, want the hex's %v beside the shout", got, healing)
	}

	shout.Deactivate(sim)
	if got, want := target.GetStats()[stats.AttackPower], hexOfWeakness(t); got != want {
		t.Errorf("with the shout off, attack power is %v, want the hex's %v", got, want)
	}
	if got := target.PseudoStats.HealingTakenMultiplier; got != healing {
		t.Errorf("with the shout off, healing taken is %v times, want the hex's %v once", got, healing)
	}
}

// Vindication states attack power and ranged attack power: the stronger of it and Demoralizing Shout
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
	// The talent prices Vindication, so its amount is the generated reader's and not a row's.
	full := buffs.VindicationValue(0)
	if full >= 0 {
		t.Fatalf("the raid's Vindication is worth %v, want an attack power reduction", full)
	}
	if got, want := target.GetStats()[stats.AttackPower], min(demoralizingShout(t), full); got != want {
		t.Errorf("attack power is %v, want the stronger reduction's %v once", got, want)
	}
	if got, want := target.GetStats()[stats.RangedAttackPower], full; got != want {
		t.Errorf("ranged attack power is %v, want Vindication's %v beside the shout", got, want)
	}
}

// Curse of Weakness lowers the damage each hit does, not attack power, so it stacks with the shouts.
func TestGeneratedCurseOfWeaknessStacksWithDemoralizingShout(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()
	applyGeneratedTestDebuffs(target, &proto.Debuffs{DemoralizingShout: true, CurseOfWeakness: true})
	if got, want := target.GetStats()[stats.AttackPower], demoralizingShout(t); got != want {
		t.Errorf("attack power is %v, want the shout's %v alone", got, want)
	}
	if got, want := target.GetStats()[stats.PhysicalDamage], rowAmount(t, 11708, dbcenums.A_MOD_DAMAGE_DONE); got != want {
		t.Errorf("physical damage is %v, want the game files' %v", got, want)
	}
}

// The game files state Dust Cloud as a stacking row, so the raid's copy takes its stacks from a
// driver: without one, the amount would follow zero stacks and apply nothing.
func TestGeneratedDustCloudTakesItsStackFromTheDriver(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()
	base := target.GetStats()[stats.ArmorNoMods]
	applyGeneratedTestDebuffs(target, &proto.Debuffs{DustCloud: true})
	dust := target.GetAura("Dust Cloud (External)")
	if dust == nil {
		t.Fatalf("no aura is labelled %q; the target has %v", "Dust Cloud (External)", targetAuraLabels(target))
	}
	if dust.MaxStacks < 1 {
		t.Fatalf("Dust Cloud holds %d stacks, want a stacking row", dust.MaxStacks)
	}
	if got := dust.GetStacks(); got != dust.MaxStacks {
		t.Errorf("Dust Cloud sits at %d stacks, want the driver to have taken all %d", got, dust.MaxStacks)
	}
	perStack := armorAmount(t, 1265904)
	if got, want := target.GetStats()[stats.ArmorNoMods]-base, float64(dust.MaxStacks)*perStack; got != want {
		t.Errorf("armor moved by %v, want %d stacks of the game files' %v", got, dust.MaxStacks, perStack)
	}
}

// Dust Cloud shares the armor reduction slot of Faerie Fire and Curse of Recklessness: both are on the
// target, and it takes one reduction, not both.
func TestGeneratedDustCloudAndFaerieFireDoNotStack(t *testing.T) {
	checkOneReductionApplies(t, &proto.Debuffs{FaerieFire: true, DustCloud: true},
		buffs.DustCloudCategory, stats.ArmorNoMods, map[string]float64{
			"Dust Cloud (External)":  armorAmount(t, 1265904),
			"Faerie Fire (External)": armorAmount(t, 9907),
		})
}

// Scorpid Sting and Insect Swarm share one slot: both are on the target, and its hit chance drops by
// one reduction, not both.
func TestGeneratedScorpidStingAndInsectSwarmDoNotStack(t *testing.T) {
	checkOneReductionApplies(t, &proto.Debuffs{ScorpidSting: true, InsectSwarm: true},
		buffs.ScorpidStingCategory, stats.PhysicalHitPercent, map[string]float64{
			"Scorpid Sting (External)": rowAmount(t, 3043, dbcenums.A_MOD_HIT_CHANCE),
			"Insect Swarm (External)":  rowAmount(t, 24977, dbcenums.A_MOD_HIT_CHANCE),
		})
}

// The raid's Annihilator keeps every stack the row allows up.
func TestGeneratedArmorShatterHoldsEveryStack(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()
	base := target.GetStats()[stats.ArmorNoMods]
	applyGeneratedTestDebuffs(target, &proto.Debuffs{ArmorShatter: true})
	maxStacks := spelldata.MustFind(16928).MaxStack
	if maxStacks < 1 {
		t.Fatalf("Armor Shatter states %d stacks, want a stacking row", maxStacks)
	}
	perStack := armorAmount(t, 16928)
	if got, want := target.GetStats()[stats.ArmorNoMods]-base, float64(maxStacks)*perStack; got != want {
		t.Errorf("armor moved by %v, want %d stacks of the game files' %v", got, maxStacks, perStack)
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

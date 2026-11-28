package core_test

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
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

// The amount the row's first effect applying aura states at the sim's character level.
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

// How long the row states its aura lasts.
func rowDuration(t *testing.T, spellID int32) time.Duration {
	t.Helper()
	duration := spelldata.MustFind(spellID).Duration()
	if duration <= 0 {
		t.Fatalf("spell %d states no duration", spellID)
	}
	return duration
}

// A stat amount at the sim's character level, in float32 the way the game files state the gain.
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

// The raid config's debuffs are each permanent, so the fight starts with all of them activated; the
// reset that does it in the sim needs a whole environment, which this stands in for.
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

// Expose Armor at the five combo points the raid config's debuff is cast with.
func exposeArmorAtFivePoints(t *testing.T) float64 {
	t.Helper()
	perPoint := float64(rowEffect(t, 11198, dbcenums.A_MOD_ARMOR_NO_MODS).PointsPerResource)
	if perPoint == 0 {
		t.Fatal("Expose Armor states no armor per combo point")
	}
	return 5 * perPoint
}

// The raid config's debuff is the five-point finisher.
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

// With both in the raid config, the sunder's stacks build while Expose Armor holds the category, and
// apply once Expose Armor ends.
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
// applies its reduction, the strongest; taking that one off hands the reduction to the next.
func checkOneReductionApplies(t *testing.T, debuffs *proto.Debuffs, category string, stat stats.Stat, values map[string]float64,
	prepare ...func(*core.Simulation, *core.Unit)) {
	t.Helper()
	target := core.NewGeneratedDebuffTestTarget()
	base := target.GetStats()[stat]

	sim := applyGeneratedTestDebuffs(target, debuffs)
	for _, p := range prepare {
		p(sim, target)
	}

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

// Curse of Recklessness and Faerie Fire share one armor reduction slot.
func TestGeneratedCurseOfRecklessnessAndFaerieFireDoNotStack(t *testing.T) {
	checkOneReductionApplies(t, &proto.Debuffs{CurseOfRecklessness: true, FaerieFire: true},
		buffs.CurseOfRecklessnessCategory, stats.ArmorNoMods, map[string]float64{
			"Curse of Recklessness (External)": armorAmount(t, 11717),
			"Faerie Fire (External)":           armorAmount(t, 9907),
		})
}

// Armor reduction slot in EV-0147.
func TestGeneratedSwingProcDebuffsShareFaerieFiresSlot(t *testing.T) {
	for _, debuff := range swingProcDebuffs {
		t.Run(debuff.name, func(t *testing.T) {
			debuffs := proto.Debuffs{ArmorShatter: debuff.debuffs.ArmorShatter, PunctureArmor: debuff.debuffs.PunctureArmor, FaerieFire: true}
			label := spelldata.MustFind(debuff.spellID).Name + " (External)"
			stacks := max(spelldata.MustFind(debuff.spellID).MaxStack, 1)
			checkOneReductionApplies(t, &debuffs, buffs.FaerieFireCategory, stats.ArmorNoMods, map[string]float64{
				label:                    armorAmount(t, debuff.spellID) * float64(stacks),
				"Faerie Fire (External)": armorAmount(t, 9907),
			}, func(sim *core.Simulation, target *core.Unit) {
				if aura := target.GetAura(label); aura.MaxStacks > 0 {
					aura.SetStacks(sim, aura.MaxStacks)
				}
			})
		})
	}
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

	// A hand-written slow states the multiplier the target's speed is divided by.
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

// Demoralizing Roar and Demoralizing Shout are the same category.
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

// The raid's Judgement of the Crusader in its improved state adds what Libram of Fervor's equip spell
// 28852 adds to every effect of the judgement.
func TestJudgementOfTheCrusaderTakesLibramOfFervorOnItsImprovedState(t *testing.T) {
	judgement := rowAmount(t, 20303, dbcenums.A_MOD_DAMAGE_TAKEN)
	libram := libramOfFervorJudgementBonus(t)

	for _, row := range []struct {
		state proto.TristateEffect
		want  float64
	}{
		{proto.TristateEffect_TristateEffectMissing, 0},
		{proto.TristateEffect_TristateEffectRegular, judgement},
		{proto.TristateEffect_TristateEffectImproved, judgement + libram},
	} {
		t.Run(row.state.String(), func(t *testing.T) {
			target := core.NewGeneratedDebuffTestTarget()
			applyGeneratedTestDebuffs(target, &proto.Debuffs{JudgementOfTheCrusader: row.state})

			if got := target.PseudoStats.SchoolBonusSpellDamage[stats.SchoolIndexHoly]; got != row.want {
				t.Errorf("the target takes %v more Holy damage, want %v", got, row.want)
			}
		})
	}
}

// A paladin's own judgement without the libram bids in the same category as the improved raid copy,
// so the target takes the stronger of the two once.
func TestImprovedJudgementOfTheCrusaderOutbidsAPaladinsOwnWithoutTheLibram(t *testing.T) {
	judgement := rowAmount(t, 20303, dbcenums.A_MOD_DAMAGE_TAKEN)
	target := core.NewGeneratedDebuffTestTarget()
	sim := applyGeneratedTestDebuffs(target, &proto.Debuffs{JudgementOfTheCrusader: proto.TristateEffect_TristateEffectImproved})

	buffs.NewSpellDataJudgementOfTheCrusaderAuraRank(target, buffs.JudgementRank{SpellID: 20303, Rank: 6, Value: judgement}).Activate(sim)

	if got, want := target.PseudoStats.SchoolBonusSpellDamage[stats.SchoolIndexHoly], judgement+libramOfFervorJudgementBonus(t); got != want {
		t.Errorf("the target takes %v more Holy damage, want the improved copy's %v alone", got, want)
	}
}

// The flat amount spell 28852 adds to every effect of Judgement of the Crusader.
func libramOfFervorJudgementBonus(t *testing.T) float64 {
	t.Helper()
	amount := spelldata.MustFind(28852).Effect(dbcenums.A_ADD_FLAT_MODIFIER, int32(dbcenums.SPELLMOD_ALL_EFFECTS)).BaseValue()
	if amount == 0 {
		t.Fatalf("spell 28852 states no flat amount on every effect")
	}
	return amount
}

// Every flat attack-power reduction bids in one category, so the target loses the strongest once.
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

// Without the shout, Demoralizing Screech and Hex of Weakness share the attack power slot.
func TestGeneratedDemoralizingScreechAndHexOfWeaknessDoNotStack(t *testing.T) {
	checkOneReductionApplies(t, &proto.Debuffs{HexOfWeakness: true, DemoralizingScreech: true},
		buffs.DemoralizingScreechCategory, stats.AttackPower, map[string]float64{
			"Hex of Weakness (External)":      hexOfWeakness(t),
			"Demoralizing Screech (External)": demoralizingScreech(t),
		})
}

// Hex of Weakness states attack power and healing taken, and each bids alone: the healing taken,
// which the shout does not state, stays the hex's.
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

// Vindication states attack power and ranged attack power: the ranged attack power, which the shout
// does not state, stays Vindication's.
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

// Dust Cloud shares the armor reduction slot of Faerie Fire and Curse of Recklessness.
func TestGeneratedDustCloudAndFaerieFireDoNotStack(t *testing.T) {
	checkOneReductionApplies(t, &proto.Debuffs{FaerieFire: true, DustCloud: true},
		buffs.DustCloudCategory, stats.ArmorNoMods, map[string]float64{
			"Dust Cloud (External)":  armorAmount(t, 1265904),
			"Faerie Fire (External)": armorAmount(t, 9907),
		})
}

// Scorpid Sting and Insect Swarm share one slot.
func TestGeneratedScorpidStingAndInsectSwarmDoNotStack(t *testing.T) {
	checkOneReductionApplies(t, &proto.Debuffs{ScorpidSting: true, InsectSwarm: true},
		buffs.ScorpidStingCategory, stats.PhysicalHitPercent, map[string]float64{
			"Scorpid Sting (External)": rowAmount(t, 3043, dbcenums.A_MOD_HIT_CHANCE),
			"Insect Swarm (External)":  rowAmount(t, 24977, dbcenums.A_MOD_HIT_CHANCE),
		})
}

// The raid's Armor Shatter is worth the row's armor once per stack, up to the stacks the row allows.
func TestGeneratedArmorShatterPricesEveryStack(t *testing.T) {
	target := core.NewGeneratedDebuffTestTarget()
	base := target.GetStats()[stats.ArmorNoMods]
	sim := applyGeneratedTestDebuffs(target, &proto.Debuffs{ArmorShatter: true})
	maxStacks := int32(spelldata.MustFind(16928).MaxStack)
	if maxStacks < 2 {
		t.Fatalf("Armor Shatter states %d stacks, want a stacking row", maxStacks)
	}
	shatter := target.GetAura("Armor Shatter (External)")
	perStack := armorAmount(t, 16928)
	for _, stacks := range []int32{1, maxStacks, maxStacks + 1} {
		shatter.SetStacks(sim, stacks)
		want := float64(min(stacks, maxStacks)) * perStack
		if got := target.GetStats()[stats.ArmorNoMods] - base; got != want {
			t.Errorf("at %d stacks armor moved by %v, want %v", stacks, got, want)
		}
	}
}

// A raid's weapon proc debuff and the weapon one wielder procs it with.
type swingProcDebuff struct {
	name    string
	debuffs *proto.Debuffs
	itemID  int32
	spellID int32
}

var swingProcDebuffs = []swingProcDebuff{
	{"Armor Shatter", &proto.Debuffs{ArmorShatter: true}, 12798, 16928},
	{"Puncture Armor 1", &proto.Debuffs{PunctureArmor: 1}, 13204, 17315},
	{"Puncture Armor 2", &proto.Debuffs{PunctureArmor: 2}, 13286, 17315},
	{"Puncture Armor 3", &proto.Debuffs{PunctureArmor: 3}, 9485, 1293005},
}

// The weapon's procs per minute, as the item database states them.
func swingProcPPM(t *testing.T, debuff swingProcDebuff) float64 {
	t.Helper()
	item := core.GetItemByID(debuff.itemID)
	if item == nil {
		t.Fatalf("the item database has no item %d", debuff.itemID)
	}
	for _, effect := range item.ItemEffects {
		if ppm := effect.GetProc().GetPpm(); effect.BuffId == debuff.spellID && ppm > 0 {
			return ppm
		}
	}
	t.Fatalf("item %d states no procs per minute for spell %d", debuff.itemID, debuff.spellID)
	return 0
}

// A raid of one with the raid's weapon proc debuff on, its database the one the request carries.
func swingProcDebuffRaid(debuff swingProcDebuff, db *proto.SimDatabase) *proto.Raid {
	return core.SinglePlayerRaidProto(&proto.Player{
		Name:      "Caster",
		Class:     proto.Class_ClassShaman,
		Buffs:     &proto.IndividualBuffs{},
		Spec:      &proto.Player_ElementalShaman{},
		Equipment: &proto.EquipmentSpec{},
		Database:  db,
	}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, debuff.debuffs)
}

// Wielder hits in EV-0147. Strike dodges in EV-0152.
const wielderOtherHitsPerMinute = 48
const wielderStrikeDodgeChance = 0.065

// The raid's copy of a weapon proc debuff is one main-hand wielder's proc, rolled on its swings and its
// other landed hits. A row on the melee hit table is a strike: the target dodges its share of them,
// and one that lands off a swing rolls once more.
func TestGeneratedSwingProcDebuffsLandAtTheWeaponsRate(t *testing.T) {
	const iterations, seconds = 10000, 300
	for _, debuff := range swingProcDebuffs {
		t.Run(debuff.name, func(t *testing.T) {
			speed := core.GetItemByID(debuff.itemID).SwingSpeed
			chance := swingProcPPM(t, debuff) * speed / 60
			swings := 60 / speed
			hits := swings + wielderOtherHitsPerMinute
			perMinute := hits * chance
			if spelldata.MustFind(debuff.spellID).DefenseTypeCore() == core.DefenseTypeMelee {
				landed := chance * (1 - wielderStrikeDodgeChance)
				perMinute = landed * (hits + swings*landed)
			}
			result := core.RunRaidSim(&proto.RaidSimRequest{
				SimOptions: &proto.SimOptions{RandomSeed: 7, Iterations: iterations},
				Raid:       swingProcDebuffRaid(debuff, nil),
				Encounter:  &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: seconds},
			})
			if result.Error != nil {
				t.Fatalf("the sim failed: %s", result.Error.Message)
			}
			var procs float64
			found := false
			for _, aura := range result.EncounterMetrics.Targets[0].Auras {
				if aura.Id.GetSpellId() == debuff.spellID {
					procs, found = aura.ProcsAvg, true
				}
			}
			if !found {
				t.Fatalf("the target carries no aura of spell %d", debuff.spellID)
			}
			want := perMinute * seconds / 60
			if math.Abs(procs-want) > 0.015*want {
				t.Errorf("the raid's copy landed %.2f times a fight, want %.2f within 1.5%% (%.2f landings a minute over %d s)", procs, want, perMinute, seconds)
			}
		})
	}
}

// Steps the sim to the time given, or to the end of the fight, checking each step.
func stepChecking(sim *core.Simulation, until time.Duration, check func()) {
	for sim.CurrentTime < until && !sim.Step() {
		check()
	}
}

// The raid's Puncture Armor adds one stack a landing, up to the row's max, and falls off when no
// swing lands it for its duration. While the raid's debuffs are disabled no swing lands it.
func TestGeneratedPunctureArmorStacksAndFallsOff(t *testing.T) {
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 7},
		Raid:       swingProcDebuffRaid(swingProcDebuffs[2], nil),
		Encounter:  &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 1500},
	}, simsignals.CreateSignals())
	sim.Reset()
	target := sim.Encounter.AllTargets[0]
	aura := target.GetAura("Puncture Armor (External)")
	if aura == nil || aura.MaxStacks < 2 {
		t.Fatalf("the raid's Puncture Armor is %v, want a stacking aura on the target", aura)
	}
	base := target.GetStats()[stats.ArmorNoMods]
	perStack := armorAmount(t, 17315)

	landed, fellOff, wasActive := false, false, false
	stepChecking(sim, 1100*time.Second, func() {
		if aura.IsActive() && !landed {
			landed = true
			if got, want := target.GetStats()[stats.ArmorNoMods]-base, perStack; aura.GetStacks() != 1 || got != want {
				t.Errorf("the first landing left %d stacks and moved armor by %v, want 1 stack worth %v", aura.GetStacks(), got, want)
			}
		}
		if aura.GetStacks() > aura.MaxStacks {
			t.Fatalf("Puncture Armor holds %d stacks, past the row's %d", aura.GetStacks(), aura.MaxStacks)
		}
		fellOff = fellOff || wasActive && !aura.IsActive()
		wasActive = aura.IsActive()
	})
	if !landed || !fellOff {
		t.Fatalf("over the fight Puncture Armor landed: %v, fell off: %v; want both", landed, fellOff)
	}

	target.DisableRaidDebuffs(sim)
	stepChecking(sim, sim.CurrentTime+180*time.Second, func() {
		if aura.IsActive() {
			t.Fatalf("Puncture Armor landed at %v with the raid's debuffs disabled", sim.CurrentTime)
		}
	})
	target.ApplyRaidDebuffsAt(sim, sim.CurrentTime)
	relanded := false
	stepChecking(sim, sim.CurrentTime+180*time.Second, func() {
		relanded = relanded || aura.IsActive()
	})
	if !relanded {
		t.Error("Puncture Armor never landed in the 180 s after the raid's debuffs landed again, want it back")
	}
}

// Without the full item database the sim knows only the items the request carries. The raid's
// weapon proc debuff reads its weapon from there, and stops on a message naming the item when the
// request lacks it.
func TestSwingProcDebuffsReadTheWeaponTheRequestCarries(t *testing.T) {
	for _, debuff := range swingProcDebuffs {
		t.Run(debuff.name, func(t *testing.T) {
			item := requestItem(t, debuff.itemID)
			restore := core.KeepOnlyItems(nil)
			defer restore()

			message := func() (message string) {
				defer func() { message = fmt.Sprint(recover()) }()
				core.ComputeStats(&proto.ComputeStatsRequest{Raid: swingProcDebuffRaid(debuff, nil), SkipRotation: true})
				return
			}()
			if !strings.Contains(message, fmt.Sprintf("item %d", debuff.itemID)) {
				t.Errorf("without its weapon the raid's debuff stopped on %q, want a message naming item %d", message, debuff.itemID)
			}

			raid := swingProcDebuffRaid(debuff, &proto.SimDatabase{Items: []*proto.SimItem{item}})
			if result := core.ComputeStats(&proto.ComputeStatsRequest{Raid: raid, SkipRotation: true}); result.ErrorResult != "" {
				t.Errorf("with its weapon in the request, ComputeStats failed: %s", result.ErrorResult)
			}
		})
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

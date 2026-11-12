//go:build with_db

package feralcat

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/druid"
)

// A dodged bite gives back its row's miss refund of the base energy and keeps the extra energy and the
// combo points.
func TestBiteDodgeRefundsAndKeepsExtraEnergyAndComboPoints(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	target := cat.CurrentTarget
	target.PseudoStats.BaseDodgeChance = 1
	cat.AddComboPoints(sim, cat.CurrentTarget, 5, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
	refund := spelldata.MustFind(cat.FerociousBite.ActionID.SpellID).MissRefund()
	want := cat.CurrentEnergy() - (1-refund)*cat.FerociousBite.Cost.GetCurrentCost()
	if !cat.FerociousBite.Cast(sim, target) {
		t.Fatal("bite did not cast")
	}
	if cat.FerociousBite.SpellMetrics[target.UnitIndex].Dodges != 1 {
		t.Fatal("bite was not dodged")
	}
	if cat.CurrentEnergy() != want || cat.ComboPoints(cat.CurrentTarget) != 5 {
		t.Fatalf("dodged bite leaves energy=%g CP=%d, want %g and 5", cat.CurrentEnergy(), cat.ComboPoints(cat.CurrentTarget), want)
	}
	t.Logf("dodged bite: energy=%g CP=%d", cat.CurrentEnergy(), cat.ComboPoints(cat.CurrentTarget))
}

func TestBiteLandedSpendsExtraEnergyAndComboPoints(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	target := cat.CurrentTarget
	target.PseudoStats.BaseDodgeChance = 0
	cat.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
	cat.AddComboPoints(sim, cat.CurrentTarget, 5, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
	if !cat.FerociousBite.Cast(sim, target) {
		t.Fatal("bite did not cast")
	}
	if metrics := cat.FerociousBite.SpellMetrics[target.UnitIndex]; metrics.Hits+metrics.Crits != 1 {
		t.Fatal("bite did not land")
	}
	if cat.CurrentEnergy() != 0 || cat.ComboPoints(cat.CurrentTarget) != 0 {
		t.Fatalf("landed bite leaves energy=%g CP=%d, want 0 and 0", cat.CurrentEnergy(), cat.ComboPoints(cat.CurrentTarget))
	}
}

// The trainer's top rank and the book rank of Ferocious Bite each register, and either spends the extra energy and
// the combo points when it lands.
func TestBothBiteRanksSpendExtraEnergyAndComboPoints(t *testing.T) {
	for _, id := range []int32{22829, 31018} {
		sim, cat := newCat(t, "", nil)
		target := cat.CurrentTarget
		target.PseudoStats.BaseDodgeChance = 0
		cat.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
		cat.AddComboPoints(sim, cat.CurrentTarget, 5, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
		bite := cat.GetSpell(core.ActionID{SpellID: id})
		if bite == nil {
			t.Fatalf("spell %d is not registered", id)
		}
		if coef := spelldata.MustFind(id).DamageEffect().APCoef; coef <= 0 {
			t.Errorf("spell %d states an attack power share of %v, want both ranks to scale", id, coef)
		}
		if !bite.Cast(sim, target) {
			t.Fatalf("spell %d did not cast", id)
		}
		if metrics := bite.SpellMetrics[target.UnitIndex]; metrics.Hits+metrics.Crits != 1 {
			t.Fatalf("spell %d did not land", id)
		}
		if cat.CurrentEnergy() != 0 || cat.ComboPoints(cat.CurrentTarget) != 0 {
			t.Errorf("spell %d leaves energy=%g CP=%d, want 0 and 0", id, cat.CurrentEnergy(), cat.ComboPoints(cat.CurrentTarget))
		}
	}
}

// Predatory Discipline 1328613: a landed bite keeps the extra energy and still deals its damage.
func TestPredatoryDisciplineBiteKeepsExtraEnergy(t *testing.T) {
	bite := func(talents string) float64 {
		sim, cat := newCat(t, talents, nil)
		target := cat.CurrentTarget
		target.PseudoStats.BaseDodgeChance = 0
		cat.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
		cat.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
		cat.AddComboPoints(sim, cat.CurrentTarget, 5, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
		want := cat.CurrentEnergy() - cat.FerociousBite.Cost.GetCurrentCost()
		if !cat.FerociousBite.Cast(sim, target) {
			t.Fatal("bite did not cast")
		}
		if cat.ComboPoints(cat.CurrentTarget) != 0 {
			t.Fatalf("landed bite leaves CP=%d, want 0", cat.ComboPoints(cat.CurrentTarget))
		}
		if talents != "" && cat.CurrentEnergy() != want {
			t.Fatalf("bite with Predatory Discipline leaves energy=%g, want %g", cat.CurrentEnergy(), want)
		}
		return cat.FerociousBite.SpellMetrics[target.UnitIndex].TotalDamage
	}
	plain := bite("")
	kept := bite("0000000000000000-00000000000010000000-0000000000000000")
	if kept != plain {
		t.Errorf("bite with Predatory Discipline dealt %g, want the untalented %g", kept, plain)
	}
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

// A Rake or Pounce Bleed tick reads attack power when it ticks.
func TestBleedTicksScaleWithAttackPower(t *testing.T) {
	for _, c := range []struct {
		name  string
		spell func(*FeralDruid) *druid.DruidSpell
	}{
		{"Rake", func(cat *FeralDruid) *druid.DruidSpell { return cat.Rake }},
		{"Pounce Bleed", func(cat *FeralDruid) *druid.DruidSpell { return cat.PounceBleed }},
	} {
		sim, cat := newCat(t, "", nil)
		spell := c.spell(cat)
		cat.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
		dot := spell.Dot(cat.CurrentTarget)
		dot.Apply(sim)
		metric := &spell.SpellMetrics[cat.CurrentTarget.UnitIndex]
		before := metric.TotalDamage
		dot.TickOnce(sim)
		low := metric.TotalDamage - before
		cat.AddStatDynamic(sim, stats.AttackPower, 100)
		before = metric.TotalDamage
		dot.TickOnce(sim)
		high := metric.TotalDamage - before
		want := 100 * spelldata.MustFind(spell.ActionID.SpellID).PeriodicDamageEffect().APCoeff()
		if want <= 0 || math.Abs((high-low)-want) > 1e-5 {
			t.Fatalf("%s: 100 AP tick gain=%g, want %g from the periodic effect's share", c.name, high-low, want)
		}
		t.Logf("%s 100 AP gain per tick=%g", c.name, high-low)
	}
}

// Pounce's stun and combo point land when its bleed misses or is dodged, and the avoided bleed refunds
// no energy.
func TestAvoidedPounceBleedKeepsStunAndComboPoint(t *testing.T) {
	for _, avoid := range []string{"miss", "dodge"} {
		t.Run(avoid, func(t *testing.T) {
			sim, cat := newCat(t, "", nil)
			target := cat.CurrentTarget
			target.PseudoStats.BaseDodgeChance = 0
			if avoid == "miss" {
				cat.AttackTables[target.UnitIndex].Hand(cat.PounceBleed.Spell).BaseMissChance = 1
			} else {
				cat.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
				target.PseudoStats.BaseDodgeChance = 1
			}
			cat.ProwlAura.Activate(sim)
			want := cat.CurrentEnergy() - cat.Pounce.Cost.GetCurrentCost()
			if !cat.Pounce.Cast(sim, target) {
				t.Fatal("pounce did not cast")
			}
			m := cat.PounceBleed.SpellMetrics[target.UnitIndex]
			avoided := m.Dodges
			if avoid == "miss" {
				avoided = m.Misses
			}
			if bleedUp := cat.PounceBleed.Dot(target).IsActive(); avoided != 1 || m.Hits != 0 || bleedUp {
				t.Fatalf("bleed avoided %d times, %d hits, up=%v, want 1, 0 and no bleed", avoided, m.Hits, bleedUp)
			}
			stunUp := target.GetAura("Pounce Stun - " + cat.Label).IsActive()
			if !stunUp || cat.ComboPoints(cat.CurrentTarget) != 1 {
				t.Errorf("stun up=%v and CP=%d, want the stun and 1", stunUp, cat.ComboPoints(cat.CurrentTarget))
			}
			if cat.CurrentEnergy() != want {
				t.Errorf("energy=%g, want %g with no refund", cat.CurrentEnergy(), want)
			}
		})
	}
}

// A second Pounce whose bleed is dodged leaves the first Pounce Bleed ticking to its own end.
func TestAvoidedRePounceKeepsTheTickingBleed(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	target := cat.CurrentTarget
	target.PseudoStats.BaseDodgeChance = 0
	cat.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
	cat.ProwlAura.Activate(sim)
	if !cat.Pounce.Cast(sim, target) {
		t.Fatal("first pounce did not cast")
	}
	dot := cat.PounceBleed.Dot(target)
	if !dot.IsActive() {
		t.Fatal("first bleed did not land")
	}
	expires := dot.ExpiresAt()
	sim.CurrentTime = cat.GCD.ReadyAt()
	fillResource(sim, cat)
	target.PseudoStats.BaseDodgeChance = 1
	cat.ProwlAura.Activate(sim)
	if !cat.Pounce.Cast(sim, target) {
		t.Fatal("second pounce did not cast")
	}
	metric := &cat.PounceBleed.SpellMetrics[target.UnitIndex]
	if metric.Dodges != 1 || !dot.IsActive() || dot.ExpiresAt() != expires {
		t.Fatalf("second bleed dodged %d times; first bleed up=%v, ends %v, want 1 dodge and up to %v", metric.Dodges, dot.IsActive(), dot.ExpiresAt(), expires)
	}
	before := metric.TotalDamage
	dot.TickOnce(sim)
	if metric.TotalDamage <= before {
		t.Error("the first bleed dealt no tick after the dodged second one")
	}
}

func TestRakeHitScalesWithAttackPower(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	target := cat.CurrentTarget
	target.PseudoStats.BaseDodgeChance = 0
	cat.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
	cat.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
	metric := &cat.Rake.SpellMetrics[target.UnitIndex]
	hitDamage := func() float64 {
		before := metric.TotalDamage - metric.TotalTickDamage
		cat.Rake.SkipCastAndApplyEffects(sim, target)
		return metric.TotalDamage - metric.TotalTickDamage - before
	}
	low := hitDamage()
	cat.AddStatDynamic(sim, stats.AttackPower, 100)
	high := hitDamage()
	want := 100 * spelldata.MustFind(cat.Rake.ActionID.SpellID).DamageEffect().APCoeff()
	if want <= 0 || math.Abs((high-low)-want) > 1e-5 {
		t.Fatalf("100 AP hit gain=%g, want %g from the direct effect's share", high-low, want)
	}
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
		cat.AddComboPoints(sim, cat.CurrentTarget, cp, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
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
		share := spelldata.MustFind(cat.Rip.ActionID.SpellID).PeriodicDamageEffect().APCoeff()
		if want := 100 * share * float64(min(cp, 4)); share <= 0 || math.Abs(gain-want) > 1e-5 {
			t.Fatalf("cp=%d tick gain=%g, want %g", cp, gain, want)
		}
		t.Logf("Rip %d CP: 100 AP gains %g per tick / %g total", cp, gain, gain*6)
	}
}

func TestCatShippedPresetRuns(t *testing.T) {
	req := catPowershiftRequest(powershiftTalents, core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "adaptive").Rotation, false)
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

func TestCatStealthOpenersRun(t *testing.T) {
	for name, opener := range map[string]int32{"ravage": 9867, "pounce": 9827} {
		t.Run(name, func(t *testing.T) {
			rotation := defaultCatRotation()
			rotation.PrepullActions = []*proto.APLPrepullAction{{DoAtValue: konst("-1s"), Action: spellAction(spellProwl)}}
			rotation.PriorityList = append([]*proto.APLListItem{castWhen(opener, auraUp(spellProwl))}, rotation.PriorityList...)
			req := catPowershiftRequest(powershiftTalents, rotation, false)
			req.Raid.Parties[0].Players[0].Equipment = core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "default").GearSet
			player := runCat(t, req)
			if castsOf(player.Actions, opener) != 1 || castsOf(player.Actions, 9830) == 0 {
				t.Fatalf("%s must open once then continue with Shred", name)
			}
		})
	}
}

func TestBleedImmuneTargetTakesNoRakeAndNoAssumedBleed(t *testing.T) {
	const rendAndTear5 = "0000000000000000-00000000000100005000-0000000000000000"
	sim, cat := newCat(t, rendAndTear5, nil)
	target := cat.CurrentTarget
	target.BleedImmune = true
	fillResource(sim, cat)
	cat.Rake.Cast(sim, target)
	if cat.Rake.Dot(target).IsActive() || cat.Rake.SpellMetrics[target.UnitIndex].TotalDamage != 0 || cat.ComboPoints(cat.CurrentTarget) != 0 {
		t.Fatal("rake landed on a bleed-immune target")
	}

	shred := func(assume, immune bool) float64 {
		sim, cat := newCat(t, rendAndTear5, nil)
		cat.AssumeBleedActive = assume
		cat.CurrentTarget.BleedImmune = immune
		fillResource(sim, cat)
		cat.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
		cat.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
		cat.CurrentTarget.PseudoStats.BaseDodgeChance = 0
		if !cat.Shred.Cast(sim, cat.CurrentTarget) {
			t.Fatal("shred did not cast")
		}
		return cat.Shred.SpellMetrics[cat.CurrentTarget.UnitIndex].TotalDamage
	}
	if plain, assumed, immune := shred(false, false), shred(true, false), shred(true, true); assumed <= plain*1.01 || immune != plain {
		t.Fatalf("shred %g plain, %g with an assumed bleed, %g on a bleed-immune target", plain, assumed, immune)
	}
}

// Rend and Tear raises Shred on a bleeding target, by the same amount under each of the cat's own
// bleeds. With the bleed assumed, a target with no bleed of the cat's own takes the same raised Shred,
// and another player's bleed counts the same.
func TestAssumedBleedRaisesShredLikeOwnBleed(t *testing.T) {
	const rendAndTear5 = "0000000000000000-00000000000100005000-0000000000000000"
	shred := func(assume bool, bleed func(cat, other *FeralDruid) *druid.DruidSpell) float64 {
		sim, cat, other := newCatBesideAnother(t, rendAndTear5)
		cat.AssumeBleedActive = assume
		target := cat.CurrentTarget
		cat.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
		cat.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
		target.PseudoStats.BaseDodgeChance = 0
		if bleed != nil {
			bleed(cat, other).Dot(target).Apply(sim)
		}
		fillResource(sim, cat)
		if !cat.Shred.Cast(sim, target) {
			t.Fatal("shred did not cast")
		}
		return cat.Shred.SpellMetrics[target.UnitIndex].TotalDamage
	}
	plain := shred(false, nil)
	bleeding := shred(false, func(cat, _ *FeralDruid) *druid.DruidSpell { return cat.Rake })
	if plain <= 0 || bleeding <= plain*1.01 {
		t.Fatalf("shred %g without a bleed, %g with Rake up: Rend and Tear did not apply", plain, bleeding)
	}
	for name, bleed := range map[string]func(cat, other *FeralDruid) *druid.DruidSpell{
		"Rip":            func(cat, _ *FeralDruid) *druid.DruidSpell { return cat.Rip },
		"Lacerate":       func(cat, _ *FeralDruid) *druid.DruidSpell { return cat.Lacerate },
		"Pounce Bleed":   func(cat, _ *FeralDruid) *druid.DruidSpell { return cat.PounceBleed },
		"the other Rake": func(_, other *FeralDruid) *druid.DruidSpell { return other.Rake },
	} {
		if got := shred(false, bleed); math.Abs(got-bleeding) > 1e-6 {
			t.Errorf("shred with %s up = %g, with Rake up = %g", name, got, bleeding)
		}
	}
	if assumed := shred(true, nil); math.Abs(assumed-bleeding) > 1e-6 {
		t.Fatalf("shred with the bleed assumed = %g, with Rake up = %g", assumed, bleeding)
	}
}

// Two cats of one build against one target: the cat under test and another whose bleeds it can read.
func newCatBesideAnother(t *testing.T, talents string) (*core.Simulation, *FeralDruid, *FeralDruid) {
	t.Helper()
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{catPlayer("Cat", talents, nil, nil), catPlayer("Other cat", talents, nil, nil)},
			Buffs:   &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 60},
	}, simsignals.CreateSignals())
	sim.Reset()
	players := sim.Raid.Parties[0].Players
	return sim, players[0].(*FeralDruid), players[1].(*FeralDruid)
}

// Rend and Tear raises Rake, Rip and Pounce Bleed ticks. Each tick counts its own
// bleed as the bleed on the target, and the talent is the only difference between the two builds.
func TestRendAndTearRaisesBleedTicks(t *testing.T) {
	const rendAndTear5 = "0000000000000000-00000000000000005000-0000000000000000"
	for _, c := range []struct {
		name  string
		spell func(*FeralDruid) *druid.DruidSpell
		want  float64
	}{
		{"Rake", func(cat *FeralDruid) *druid.DruidSpell { return cat.Rake }, 1.15},
		{"Rip", func(cat *FeralDruid) *druid.DruidSpell { return cat.Rip }, 1.15},
		{"Pounce Bleed", func(cat *FeralDruid) *druid.DruidSpell { return cat.PounceBleed }, 1.15},
	} {
		tick := func(talents string) float64 {
			sim, cat := newCat(t, talents, nil)
			cat.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
			spell := c.spell(cat)
			if spell == cat.Rip {
				cat.AddComboPoints(sim, cat.CurrentTarget, 5, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
			}
			dot := spell.Dot(cat.CurrentTarget)
			dot.Apply(sim)
			metric := &spell.SpellMetrics[cat.CurrentTarget.UnitIndex]
			before := metric.TotalDamage
			dot.TickOnce(sim)
			return metric.TotalDamage - before
		}
		plain, raised := tick(""), tick(rendAndTear5)
		if plain <= 0 || math.Abs(raised/plain-c.want) > 1e-6 {
			t.Errorf("%s: tick %g without the talent, %g with 5/5, want x%g", c.name, plain, raised, c.want)
		}
	}
}

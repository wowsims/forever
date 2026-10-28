//go:build with_db

package forever_test

import (
	"math"
	"testing"
	"time"

	_ "github.com/wowsims/forever/sim/common/classic" // registers Thunderfury and Enchant Weapon - Crusader.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/core/testagent"
)

// The caster stand-in's spells: a 2 s cast and an instant, both on the global cooldown.
const (
	castTimeSpellID = 403
	instantSpellID  = 8042
)

func init() {
	testagent.Register(proto.Player_ElementalShaman{}, proto.Spec_SpecElementalShaman, func(player *proto.Player, spec any) {
		player.Spec = spec.(*proto.Player_ElementalShaman)
	}, func(agent *testagent.Agent) {
		agent.EnableMelee()
		for _, spell := range []struct {
			id       int32
			castTime time.Duration
		}{{castTimeSpellID, 2 * time.Second}, {instantSpellID, 0}} {
			agent.RegisterSpell(core.SpellConfig{
				ActionID:    core.ActionID{SpellID: spell.id},
				SpellSchool: core.SpellSchoolNature,
				DefenseType: core.DefenseTypeMagic,
				ProcMask:    core.ProcMaskSpellDamage,
				Cast: core.CastConfig{
					DefaultCast: core.Cast{GCD: core.GCDDefault, CastTime: spell.castTime},
				},
				DamageMultiplier: 1,
				ThreatMultiplier: 1,
				ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
					spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeMagicHitAndCrit)
				},
			})
		}
	})
	testagent.Register(proto.Player_Rogue{}, proto.Spec_SpecRogue, func(player *proto.Player, spec any) {
		player.Spec = spec.(*proto.Player_Rogue)
	}, func(agent *testagent.Agent) {
		agent.EnableAutoAttacks(agent, core.AutoAttackOptions{Ranged: agent.WeaponFromRanged(), AutoSwingRanged: true})
	})
}

// A character swinging one weapon in the main hand against level 60 targets with no armor: the
// warrior stand-in, or with a shaman the caster stand-in.
func weaponWearer(t *testing.T, class proto.Class, itemID int32, targets int) (*core.Simulation, *core.Character) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = itemID
	player := &proto.Player{
		Name: "Wearer", Class: class, Race: proto.Race_RaceTauren,
		Equipment:   &proto.EquipmentSpec{Items: items},
		Consumables: &proto.ConsumesSpec{},
		Buffs:       &proto.IndividualBuffs{},
		Rotation:    &proto.APLRotation{},
	}
	if class == proto.Class_ClassShaman {
		core.WithSpec(player, &proto.Player_ElementalShaman{ElementalShaman: &proto.ElementalShaman{}})
	} else {
		core.WithSpec(player, &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{}})
	}
	encounter := &proto.Encounter{Duration: 180}
	for range targets {
		encounter.Targets = append(encounter.Targets, &proto.Target{Name: "target", Level: 60})
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}},
		Encounter:  encounter,
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

// The procs per minute the item's chance-on-hit entry states for the spell, which is the rate the
// weapon's proc registers on.
func itemEntryPPM(t *testing.T, itemID int32, spellID int32) float64 {
	t.Helper()
	for _, effect := range core.GetItemByID(itemID).ItemEffects {
		if effect.BuffId == spellID && effect.GetProc().GetPpm() > 0 {
			return effect.GetProc().GetPpm()
		}
	}
	t.Fatalf("%d's entry states no procs per minute for %d", itemID, spellID)
	return 0
}

// Checks the proc rolls the item entry's procs per minute on the weapon's speed, a chance between
// none and certain.
func expectWeaponRate(t *testing.T, sim *core.Simulation, proc *core.Aura, mask core.ProcMask, itemID, spellID int32) {
	t.Helper()
	want := itemEntryPPM(t, itemID, spellID) * core.GetItemByID(itemID).SwingSpeed / 60
	if got := proc.Dpm.Chance(mask, sim); want <= 0 || want >= 1 || math.Abs(got-want) > 1e-9 {
		t.Errorf("%d's proc rolls %v on a hit, want the entry's procs per minute on the weapon's speed, %v", itemID, got, want)
	}
}

// The ticks a damage over time runs for, as the row's duration over its period.
func rowTickCount(row *spelldata.Spell, tick *spelldata.Effect) int32 {
	return int32(row.Duration() / tick.Period())
}

// Graverobber's Shovel 273817's equip spell 1319047 grants attack power and ranged attack power,
// which the item's stats carry, and its effects 3 and 4 grant as much again of each against Undead.
func TestGraverobbersShovelGrantsAttackPowerAgainstUndead(t *testing.T) {
	const shovel = 273817
	row := spelldata.MustFind(1319047)
	want := stats.Stats{stats.AttackPower: row.EffectN(3).Average(), stats.RangedAttackPower: row.EffectN(4).Average()}
	_, againstUndead := warriorWithConsumes(t, &proto.ConsumesSpec{}, shovel, 0, proto.MobType_MobTypeUndead)
	if got := bonusAgainstTarget(againstUndead); got != want || want[stats.AttackPower] <= 0 {
		t.Errorf("the bonus against Undead is %v, want the row's %v", got, want)
	}
	_, againstDemons := warriorWithConsumes(t, &proto.ConsumesSpec{}, shovel, 0, proto.MobType_MobTypeDemon)
	if got := bonusAgainstTarget(againstDemons); got != (stats.Stats{}) {
		t.Errorf("the bonus against Demons is %v, want none", got)
	}
}

// Sulfuras, Hand of Ragnaros 17182 carries both its effects. Immolation 21142's damage shield deals
// its amount to the attacker of each melee hit taken, and nothing for a spell. Fireball 21162 procs
// at the procs per minute of the item's chance-on-hit entry, deals its row's direct roll, and ticks
// its periodic amount over the row's duration.
func TestSulfurasCarriesItsDamageShieldAndFireball(t *testing.T) {
	sim, character := itemWearer(t, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotMainHand, 17182)
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100, stats.SpellCritPercent: -character.GetStat(stats.SpellCritPercent)})
	enemy := sim.Encounter.ActiveTargetUnits[0]

	var shieldDamage float64
	for _, effect := range spelldata.ItemAuraDamageShields(spelldata.MustFind(21142)) {
		shieldDamage += effect.Average()
	}
	shield := character.GetSpell(core.ActionID{SpellID: 21142, Tag: 2})
	if shield == nil || shieldDamage <= 0 {
		t.Fatal("17182 registers no damage shield")
	}
	takeHit(t, sim, character, meleeSwing(enemy))
	if got := shield.SpellMetrics[enemy.UnitIndex]; got.Casts != 1 || got.TotalDamage != shieldDamage {
		t.Errorf("a melee hit taken cast the shield %d times for %v, want once for the row's %v", got.Casts, got.TotalDamage, shieldDamage)
	}
	takeHit(t, sim, character, fireSpell(enemy))
	if got := shield.SpellMetrics[enemy.UnitIndex].Casts; got != 1 {
		t.Errorf("after a spell hit taken the shield cast %d times, want once still", got)
	}

	proc := character.GetAura("Sulfuras, Hand of Ragnaros")
	if proc == nil || proc.Dpm == nil {
		t.Fatal("17182 registers no Fireball proc on a procs per minute rate")
	}
	expectWeaponRate(t, sim, proc, core.ProcMaskMeleeMHAuto, 17182, 21162)

	row := spelldata.MustFind(21162)
	hit, tick := row.DamageEffect(), row.PeriodicDamageEffect()
	ticks := rowTickCount(row, tick)
	fireball := character.GetSpell(core.ActionID{SpellID: 21162})
	fireball.Cast(sim, enemy)
	dot := fireball.Dot(enemy)
	if dot.BaseTickCount != ticks || dot.BaseTickLength != tick.Period() {
		t.Errorf("the Fireball ticks %d times every %v, want the row's %d every %v", dot.BaseTickCount, dot.BaseTickLength, ticks, tick.Period())
	}
	stepTo(t, sim, sim.CurrentTime+row.Duration()+time.Millisecond)
	metrics := fireball.SpellMetrics[enemy.UnitIndex]
	metrics.TotalDamage -= float64(ticks) * tick.Average()
	if metrics.Casts != 1 || metrics.Ticks != ticks || !damageInRange(metrics, hit) {
		t.Errorf("the Fireball cast %d times and ticked %d times for %v besides its ticks, want once, %d ticks and %v to %v",
			metrics.Casts, metrics.Ticks, metrics.TotalDamage, ticks, hit.Min(), hit.Max())
	}
}

// Ravager 7717 procs Bladestorm 9632 at the procs per minute of the item's chance-on-hit entry.
// Bladestorm is a channel on the wearer for the row's duration that cuts off a cast in progress and
// holds back the auto attacks, and every period strikes each enemy with 9633: main-hand weapon damage
// plus the strike row's flat amount. The strikes do not proc Bladestorm.
func TestRavagerChannelsBladestormOverEveryEnemy(t *testing.T) {
	sim, character := weaponWearer(t, proto.Class_ClassWarrior, 7717, 3)

	proc := character.GetAura("Ravager")
	if proc == nil || proc.Dpm == nil {
		t.Fatal("7717 registers no Bladestorm proc on a procs per minute rate")
	}
	expectWeaponRate(t, sim, proc, core.ProcMaskMeleeMHAuto, 7717, 9632)
	// Every melee hit that may proc Bladestorm now does.
	*proc.Dpm = *character.NewFixedProcChanceManager(1, core.ProcMaskMelee)

	targets := sim.Encounter.ActiveTargetUnits
	for _, target := range targets {
		target.PseudoStats.InFrontOfTarget = false
	}
	bladestorm := character.GetSpell(core.ActionID{SpellID: 9632})
	strike := character.GetSpell(core.ActionID{SpellID: 9633})
	if bladestorm == nil || strike == nil {
		t.Fatal("7717 registers no Bladestorm or no strike")
	}
	row := spelldata.MustFind(9632)
	period := row.EffectN(1).Period()
	ticks := rowTickCount(row, row.EffectN(1))
	mh := character.AutoAttacks.MH()
	extra := spelldata.MustFind(9633).DamageEffect().Average() + strike.MeleeAttackPower(targets[0])*mh.SwingSpeed/mh.AttackPowerPerDPS
	casts := func(spell *core.Spell) (total int32) {
		for _, target := range targets {
			total += spell.SpellMetrics[target.UnitIndex].Casts
		}
		return total
	}
	character.AutoAttacks.EnableAutoSwing(sim)
	swings := casts(character.AutoAttacks.MHAuto())

	start := sim.CurrentTime
	bladestorm.Cast(sim, targets[0])
	dot := bladestorm.AOEDot()
	if dot.Unit != &character.Unit || ticks < 2 || dot.BaseTickCount != ticks || dot.BaseTickLength != period {
		t.Errorf("Bladestorm ticks %d times every %v on %s, want the row's %d every %v on the wearer",
			dot.BaseTickCount, dot.BaseTickLength, dot.Unit.Label, ticks, period)
	}
	stepTo(t, sim, start+(row.Duration()-period)+time.Millisecond)
	if !character.IsChanneling() {
		t.Error("the wearer is not channeling Bladestorm before its last strike")
	}
	if got := casts(bladestorm); got != 1 {
		t.Errorf("the strikes cast Bladestorm %d times in all, want once", got)
	}
	if got := casts(character.AutoAttacks.MHAuto()) - swings; got != 0 {
		t.Errorf("the wearer swung %d times during Bladestorm, want none", got)
	}
	// The first swing after the channel would cast it again.
	proc.Deactivate(sim)

	stepTo(t, sim, start+row.Duration()+time.Millisecond)
	if dot.IsActive() || character.IsChanneling() {
		t.Errorf("Bladestorm is still up after the row's %v", row.Duration())
	}

	for _, target := range targets {
		metrics := strike.SpellMetrics[target.UnitIndex]
		if outcomes := metrics.Hits + metrics.Crits + metrics.Misses + metrics.Dodges; outcomes != ticks {
			t.Errorf("%s took %d strikes, want %d", target.Label, outcomes, ticks)
		}
		hitDamage := metrics.TotalDamage - metrics.TotalCritDamage
		if low, high := float64(metrics.Hits)*(mh.BaseDamageMin+extra), float64(metrics.Hits)*(mh.BaseDamageMax+extra); metrics.Hits == 0 || hitDamage < low || hitDamage > high {
			t.Errorf("%s took %v from %d strikes that hit, want %v to %v", target.Label, hitDamage, metrics.Hits, low, high)
		}
	}

	character.AutoAttacks.CancelAutoSwing(sim)
	proc.Activate(sim)
	strike.Cast(sim, targets[0])
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
	if got := casts(bladestorm); got != 1 {
		t.Errorf("a strike outside the channel cast Bladestorm %d times in all, want once still", got)
	}

	sim, character = weaponWearer(t, proto.Class_ClassShaman, 7717, 1)
	character.AutoAttacks.CancelAutoSwing(sim)
	enemy := sim.Encounter.ActiveTargetUnits[0]
	castTime := character.GetSpell(core.ActionID{SpellID: castTimeSpellID})
	castTime.Cast(sim, enemy)
	character.GetSpell(core.ActionID{SpellID: 9632}).Cast(sim, enemy)
	stepTo(t, sim, sim.CurrentTime+castTime.DefaultCast.CastTime+time.Millisecond)
	if got := castTime.SpellMetrics[enemy.UnitIndex].Casts; got != 0 {
		t.Errorf("the cast Bladestorm started on completed %d times, want none", got)
	}
}

// Heartseeking Crossbow 13040's 29626 procs at the procs per minute of the item's entry, on the
// crossbow's speed, and strikes the target with Shadowbolt 29640 for its row's roll.
func TestHeartseekingCrossbowStrikesWithAShadowbolt(t *testing.T) {
	sim, character := itemWearer(t, proto.Class_ClassRogue, proto.ItemSlot_ItemSlotRanged, 13040)
	enemy := sim.Encounter.ActiveTargetUnits[0]
	proc := character.GetAura("Heartseeking Crossbow")
	shadowbolt := character.GetSpell(core.ActionID{SpellID: 29640})
	if proc == nil || proc.Dpm == nil || shadowbolt == nil {
		t.Fatal("13040 registers no Shadowbolt or no proc on a procs per minute rate")
	}
	expectWeaponRate(t, sim, proc, core.ProcMaskRangedAuto, 13040, 29640)
	// Every ranged hit now procs it.
	*proc.Dpm = *character.NewFixedProcChanceManager(1, core.ProcMaskRanged)
	character.AddStatsDynamic(sim, stats.Stats{stats.PhysicalHitPercent: 100, stats.PhysicalCritPercent: -character.GetStat(stats.PhysicalCritPercent)})

	damage := spelldata.MustFind(29640).DamageEffect()
	shot := &core.Spell{ProcMask: core.ProcMaskRangedAuto, SpellSchool: core.SpellSchoolPhysical, Unit: &character.Unit}
	dealHit(t, sim, character, shot)
	metrics := shadowbolt.SpellMetrics[enemy.UnitIndex]
	if metrics.Casts != 1 || metrics.Hits != 1 || metrics.TotalDamage < damage.Min() || metrics.TotalDamage > damage.Max() {
		t.Errorf("a ranged hit cast the Shadowbolt %d times for %v from %d hits, want one hit for the row's %v to %v",
			metrics.Casts, metrics.TotalDamage, metrics.Hits, damage.Min(), damage.Max())
	}
}

// Hand of Edward the Odd 2243 procs Focus 18803 at the procs per minute of the item's chance-on-hit
// entry. Focus makes the next spell with a cast time instant, on its whole global cooldown, and that
// cast spends it; an instant leaves it up, and unspent it fades after the row's duration.
func TestHandOfEdwardTheOddMakesTheNextCastInstant(t *testing.T) {
	sim, character := weaponWearer(t, proto.Class_ClassShaman, 2243, 1)
	character.AutoAttacks.CancelAutoSwing(sim)
	enemy := sim.Encounter.ActiveTargetUnits[0]

	proc := character.GetAura("Hand of Edward the Odd")
	if proc == nil || proc.Dpm == nil {
		t.Fatal("2243 registers no Focus proc on a procs per minute rate")
	}
	expectWeaponRate(t, sim, proc, core.ProcMaskMeleeMHAuto, 2243, 18803)

	duration := spelldata.MustFind(18803).Duration()
	focus := character.GetAura("Hand of Edward the Odd Focus")
	if focus == nil || focus.Duration != duration {
		t.Fatalf("2243 registers Focus as %v, want an aura of the row's %v", focus, duration)
	}
	castTime := character.GetSpell(core.ActionID{SpellID: castTimeSpellID})
	instant := character.GetSpell(core.ActionID{SpellID: instantSpellID})

	focus.Activate(sim)
	instant.Cast(sim, enemy)
	if !focus.IsActive() {
		t.Error("an instant spent Focus, want it kept")
	}

	stepTo(t, sim, character.GCD.ReadyAt())
	start := sim.CurrentTime
	castTime.Cast(sim, enemy)
	if got := castTime.SpellMetrics[enemy.UnitIndex].Casts; got != 1 {
		t.Errorf("the cast under Focus completed %d times at once, want once", got)
	}
	if focus.IsActive() {
		t.Error("the cast left Focus up, want it spent")
	}
	if got := character.GCD.ReadyAt() - start; got != core.GCDDefault {
		t.Errorf("the instant cast spent a %v global cooldown, want %v", got, core.GCDDefault)
	}

	stepTo(t, sim, character.GCD.ReadyAt())
	castTime.Cast(sim, enemy)
	if got, want := character.Hardcast.Expires-sim.CurrentTime, castTime.DefaultCast.CastTime; got != want {
		t.Errorf("the next cast after Focus is spent takes %v, want its %v", got, want)
	}

	stepTo(t, sim, character.Hardcast.Expires+time.Millisecond)
	focus.Activate(sim)
	stepTo(t, sim, sim.CurrentTime+duration+time.Millisecond)
	if focus.IsActive() {
		t.Errorf("Focus is still up after the row's %v unspent", duration)
	}
}

// Thunderfury, Blessed Blade of the Windseeker 19019 procs 21992 at the procs per minute of the item's
// chance-on-hit entry.
func TestThunderfuryProcsAtItsEntryRate(t *testing.T) {
	sim, character := weaponWearer(t, proto.Class_ClassWarrior, 19019, 1)
	proc := character.GetAura("Thunderfury Proc")
	if proc == nil || proc.Dpm == nil {
		t.Fatal("19019 registers no proc on a procs per minute rate")
	}
	expectWeaponRate(t, sim, proc, core.ProcMaskMeleeMHAuto, 19019, 21992)
}

// Spinal Reaper 288901's 21185 hears a killing blow and fires 21186. A wearer with a mana bar gains the
// mana 21186's effect 1 states, one with only a rage bar the rage its effect 2 states in tenths. A
// death with no kill credit gives nothing, and a second kill inside the row's cooldown gives nothing
// either.
func TestSpinalReaperRestoresPowerOnAKillingBlow(t *testing.T) {
	const reaper = 288901
	cooldown := spelldata.MustFind(21185).ICD()
	gain := spelldata.MustFind(21186)
	for _, tc := range []struct {
		name  string
		class proto.Class
		want  float64
	}{
		{"mana", proto.Class_ClassWarlock, gain.EffectN(1).BaseValue()},
		{"rage", proto.Class_ClassWarrior, gain.EffectN(2).Tenths()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim, character := modWearer(t, tc.class, proto.ItemSlot_ItemSlotMainHand, reaper)
			target := sim.Encounter.AllTargets[0]
			spent := character.NewManaMetrics(core.ActionID{SpellID: 993105})
			// The power the death gives. Mana is spent first, which leaves room for the gain and holds
			// off the regen for the moment the gain is measured.
			gainOn := func(killShare float64) float64 {
				if character.HasManaBar() {
					character.SpendMana(sim, character.CurrentMana(), spent)
				}
				target.Enable(sim)
				before := character.CurrentMana() + character.CurrentRage()
				target.Die(sim, killShare)
				stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
				return character.CurrentMana() + character.CurrentRage() - before
			}

			if tc.want <= 0 || cooldown <= 0 {
				t.Fatalf("the rows state a gain of %v and a cooldown of %v, want both", tc.want, cooldown)
			}
			if got := gainOn(0); got != 0 {
				t.Errorf("a death with no kill credit gave %v, want nothing", got)
			}
			if got := gainOn(1); got != tc.want {
				t.Errorf("a killing blow gave %v, want %v", got, tc.want)
			}
			if got := gainOn(1); got != 0 {
				t.Errorf("a second killing blow inside the cooldown gave %v, want nothing", got)
			}
			stepTo(t, sim, sim.CurrentTime+cooldown)
			if got := gainOn(1); got != tc.want {
				t.Errorf("a killing blow after the cooldown gave %v, want %v", got, tc.want)
			}
		})
	}
}

// Blade of Hanna 2801 adds a stack of Improving Form 1325509 on each main-hand swing. Only the swing
// that brings the stacks to the count the row names casts Perfected Form 1325508: its damage once,
// its +20 to every stat for the row's duration, and the stacks gone so they build again.
func TestBladeOfHannaPerfectsItsFormAtThirtyStacks(t *testing.T) {
	sim, character := weaponWearer(t, proto.Class_ClassWarrior, 2801, 1)
	character.AddStatsDynamic(sim, stats.Stats{stats.PhysicalHitPercent: 100, stats.ExpertiseRating: 1000, stats.PhysicalCritPercent: -character.GetStat(stats.PhysicalCritPercent)})
	enemy := character.CurrentTarget

	stackRow, perfectedRow := spelldata.MustFind(1325509), spelldata.MustFind(1325508)
	threshold := stackRow.EffectN(2).Misc
	stacks := character.GetAuraByID(core.ActionID{SpellID: stackRow.ID})
	buff := character.GetAuraByID(core.ActionID{SpellID: perfectedRow.ID})
	perfect := character.GetSpell(core.ActionID{SpellID: perfectedRow.ID})
	if stacks == nil || buff == nil || perfect == nil || threshold != 30 {
		t.Fatalf("2801 registers no Improving Form, Perfected Form aura or spell, or the row names %d stacks, want 30", threshold)
	}

	for range threshold - 1 {
		dealHit(t, sim, character, meleeSwing(&character.Unit))
	}
	if got := stacks.GetStacks(); got != threshold-1 {
		t.Errorf("after %d swings Improving Form has %d stacks, want %d", threshold-1, got, threshold-1)
	}
	if got := perfect.SpellMetrics[enemy.UnitIndex].Casts; got != 0 {
		t.Errorf("before the %dth swing Perfected Form cast %d times, want none", threshold, got)
	}

	before := character.GetStats()
	dealHit(t, sim, character, meleeSwing(&character.Unit))
	if got := perfect.SpellMetrics[enemy.UnitIndex]; got.Casts != 1 || !damageInRange(got, perfectedRow.DamageEffect()) {
		t.Errorf("the %dth swing cast Perfected Form %d times for %v, want once for the row's %v", threshold, got.Casts, got.TotalDamage, perfectedRow.DamageEffect().Average())
	}
	if stacks.IsActive() {
		t.Errorf("Improving Form kept %d stacks after Perfected Form, want none", stacks.GetStacks())
	}
	if !buff.IsActive() || buff.Duration != perfectedRow.Duration() {
		t.Errorf("Perfected Form is up %v for %v, want up for the row's %v", buff.IsActive(), buff.Duration, perfectedRow.Duration())
	}
	for _, stat := range []stats.Stat{stats.Strength, stats.Agility, stats.Stamina, stats.Intellect, stats.Spirit} {
		if got := gained(character, before, stat); !statNear(got, 20) {
			t.Errorf("Perfected Form grants %v %s, want 20", got, stat.StatName())
		}
	}

	dealHit(t, sim, character, meleeSwing(&character.Unit))
	if got := stacks.GetStacks(); got != 1 {
		t.Errorf("the swing after Perfected Form left Improving Form at %d stacks, want 1", got)
	}
	if got := perfect.SpellMetrics[enemy.UnitIndex].Casts; got != 1 {
		t.Errorf("Perfected Form cast %d times by the end, want once", got)
	}
}

//go:build with_db

package forever_test

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"time"

	_ "github.com/wowsims/forever/sim/common/classic" // registers Thunderfury and Enchant Weapon - Crusader.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/core/testagent"
	"github.com/wowsims/forever/sim/druid"
	"github.com/wowsims/forever/sim/priest"
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
// warrior stand-in, or with a shaman or a druid the caster stand-in.
func weaponWearer(t *testing.T, class proto.Class, itemID int32, targets int) (*core.Simulation, *core.Character) {
	t.Helper()
	return weaponWearerFor(t, class, itemID, targets, 180)
}

// weaponWearer's character in an encounter of the given seconds.
func weaponWearerFor(t *testing.T, class proto.Class, itemID int32, targets int, duration float64) (*core.Simulation, *core.Character) {
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
	if class == proto.Class_ClassShaman || class == proto.Class_ClassDruid {
		core.WithSpec(player, &proto.Player_ElementalShaman{ElementalShaman: &proto.ElementalShaman{}})
	} else {
		core.WithSpec(player, &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{}})
	}
	encounter := &proto.Encounter{Duration: duration}
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

// Sets the procs per minute on the item's chance-on-hit entry for the spell for the length of the
// test; 0 leaves the entry stating no rate.
func withEntryPPM(t *testing.T, itemID int32, spellID int32, ppm float64) {
	t.Helper()
	for _, effect := range core.GetItemByID(itemID).ItemEffects {
		if proc := effect.GetProc(); effect.BuffId == spellID && proc != nil {
			committed := proc.ProcRate
			t.Cleanup(func() { proc.ProcRate = committed })
			proc.ProcRate = nil
			if ppm > 0 {
				proc.ProcRate = &proto.ProcEffect_Ppm{Ppm: ppm}
			}
			return
		}
	}
	t.Fatalf("%d has no chance-on-hit entry for %d", itemID, spellID)
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

// The Green Tower 1204 carries both its generated effects, composed into the one registration core
// takes per item: 1325228's damage shield deals its amount to the attacker of each melee hit taken
// while the off-hand is held, and the stat proc 17154 its hits trigger registers beside it.
func TestTheGreenTowerCarriesItsShieldAndItsProc(t *testing.T) {
	sim, character := itemWearer(t, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotOffHand, 1204)
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100, stats.SpellCritPercent: -character.GetStat(stats.SpellCritPercent)})
	enemy := sim.Encounter.ActiveTargetUnits[0]

	shield, shieldDamage := damageShield(t, character, 1325228)
	if character.GetAura("The Green Tower Proc") == nil {
		t.Fatal("1204 registers no stat proc beside its shield")
	}
	takeHit(t, sim, character, meleeSwing(enemy))
	if got := shield.SpellMetrics[enemy.UnitIndex]; got.Casts != 1 || got.TotalDamage != shieldDamage {
		t.Errorf("a melee hit taken cast the shield %d times for %v, want once for the row's %v", got.Casts, got.TotalDamage, shieldDamage)
	}
}

// Ravager 7717 procs Bladestorm 9632 at the procs per minute of the item's chance-on-hit entry.
// Bladestorm is a channel on the wearer for the row's duration that cuts off a cast in progress and
// holds back the auto attacks, and every period strikes each enemy with 9633: main-hand weapon damage
// plus the strike row's flat amount. The strikes do not proc Bladestorm.
func TestRavagerChannelsBladestormOverEveryEnemy(t *testing.T) {
	t.Skip("Ravager 7717 is not in the live game files; this runs again once they hold it")
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
	if core.GetItemByID(13040) == nil {
		t.Skip("item 13040 is not in the live game files; this runs again once they hold it")
	}
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
	shot := rangedShot(&character.Unit)
	dealHit(t, sim, character, shot)
	metrics := shadowbolt.SpellMetrics[enemy.UnitIndex]
	if metrics.Casts != 1 || metrics.Hits != 1 || metrics.TotalDamage < damage.Min() || metrics.TotalDamage > damage.Max() {
		t.Errorf("a ranged hit cast the Shadowbolt %d times for %v from %d hits, want one hit for the row's %v to %v",
			metrics.Casts, metrics.TotalDamage, metrics.Hits, damage.Min(), damage.Max())
	}
}

// Naga Heartpiercer 3078's 1292647 procs at the procs per minute of the item's entry, on the bow's
// speed, off the cast result of a ranged shot that lands, and Pierce Heart strikes then. A missed shot
// casts none, and the shot's hit, dealt later, rolls no second one.
func TestNagaHeartpiercerStrikesOffLandedShots(t *testing.T) {
	sim, character := itemWearer(t, proto.Class_ClassRogue, proto.ItemSlot_ItemSlotRanged, 3078)
	enemy := sim.Encounter.ActiveTargetUnits[0]
	proc := character.GetAura("Naga Heartpiercer")
	pierceHeart := character.GetSpell(core.ActionID{SpellID: 1292647})
	if proc == nil || proc.Dpm == nil || pierceHeart == nil {
		t.Fatal("3078 registers no Pierce Heart or no proc on a procs per minute rate")
	}
	expectWeaponRate(t, sim, proc, core.ProcMaskRangedAuto, 3078, 1292647)
	// Every ranged hit now procs it.
	*proc.Dpm = *character.NewFixedProcChanceManager(1, core.ProcMaskRanged)
	casts := func() int32 { return pierceHeart.SpellMetrics[enemy.UnitIndex].Casts }

	shot := rangedShot(&character.Unit)
	character.OnCastResult(sim, shot, &core.SpellResult{Target: enemy, Outcome: core.OutcomeMiss})
	if got := casts(); got != 0 {
		t.Errorf("a missed shot cast Pierce Heart %d times, want none", got)
	}
	landed := &core.SpellResult{Target: enemy, Outcome: core.OutcomeHit, Damage: 100}
	character.OnCastResult(sim, shot, landed)
	if got := casts(); got != 1 {
		t.Errorf("a landed shot's cast result cast Pierce Heart %d times, want once then", got)
	}
	character.OnSpellHitDealt(sim, shot, landed)
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
	if got := casts(); got != 1 {
		t.Errorf("the landed shot's hit brought Pierce Heart to %d casts, want it left at one", got)
	}
}

// Skullforge Reaver 13361 in the main hand casts 17484 at the procs per minute of the item's entry, on
// the sword's speed, off landed main-hand swings and specials, and never off a miss, the off hand's
// swings or a ranged hit. The brand drains the target every period for the row's duration and heals
// the wearer the row's heal every period, and a second proc refreshes it.
func TestSkullforgeReaverBrandsOnMainHandHitsAndHealsTheWearer(t *testing.T) {
	sim, character := itemWearerWithEquipment(t, proto.Class_ClassWarrior, &proto.EquipmentSpec{Items: weaponSlots(13361, 0)})
	character.AutoAttacks.CancelAutoSwing(sim)
	proc := character.GetAura("Skullforge Reaver")
	brand := character.GetSpell(core.ActionID{SpellID: 17484})
	transfer := character.GetSpell(core.ActionID{SpellID: 17484, Tag: 1})
	if proc == nil || proc.Dpm == nil || brand == nil || transfer == nil {
		t.Fatal("13361 registers no brand, no heal or no proc on a procs per minute rate")
	}
	expectWeaponRate(t, sim, proc, core.ProcMaskMeleeMHAuto, 13361, 17484)
	expectWeaponRate(t, sim, proc, core.ProcMaskMeleeMHSpecial, 13361, 17484)
	for _, mask := range []core.ProcMask{core.ProcMaskMeleeOHAuto, core.ProcMaskMeleeOHSpecial, core.ProcMaskRangedAuto} {
		if got := proc.Dpm.Chance(mask, sim); got != 0 {
			t.Errorf("a hit of mask %v rolls %v, want no roll off the main hand", mask, got)
		}
	}

	// Every main-hand hit now procs it, and every brand lands.
	*proc.Dpm = *character.NewFixedProcChanceManager(1, core.ProcMaskMeleeMH)
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100})
	character.RemoveHealth(sim, character.MaxHealth()/2)
	row := spelldata.MustFind(17484)
	drain := row.PeriodicDamageEffect()
	heal := row.FirstAura(dbcenums.A_PERIODIC_HEAL)
	ticks := rowTickCount(row, drain)
	if ticks < 2 || rowTickCount(row, heal) != ticks || drain.Min() <= 0 || heal.Min() <= 0 {
		t.Fatalf("17484 drains %v %d times and heals %v %d times, want an amount each at least twice, as many heals as drains",
			drain.Min(), ticks, heal.Min(), rowTickCount(row, heal))
	}
	enemy := character.CurrentTarget

	character.OnSpellHitDealt(sim, meleeSwing(&character.Unit), &core.SpellResult{Target: enemy, Outcome: core.OutcomeMiss})
	if got := brand.SpellMetrics[enemy.UnitIndex].Casts; got != 0 {
		t.Errorf("a missed swing cast %d brands, want none", got)
	}
	start := sim.CurrentTime
	dealHit(t, sim, character, meleeSwing(&character.Unit))
	if !brand.Dot(enemy).IsActive() || !transfer.SelfHot().IsActive() {
		t.Fatal("a main-hand swing applied no brand to the target or no heal to the wearer")
	}
	stepTo(t, sim, start+row.Duration()+time.Millisecond)
	if brand.Dot(enemy).IsActive() || transfer.SelfHot().IsActive() {
		t.Errorf("the brand is still up after the row's %v", row.Duration())
	}
	drained := brand.SpellMetrics[enemy.UnitIndex]
	if drained.Casts != 1 || drained.Ticks != ticks || drained.TotalDamage < float64(ticks)*drain.Min() || drained.TotalDamage > float64(ticks)*drain.Max() {
		t.Errorf("one proc cast %d brands that drained %v over %d ticks, want one that drains %d ticks of the row's %v to %v",
			drained.Casts, drained.TotalDamage, drained.Ticks, ticks, drain.Min(), drain.Max())
	}
	healed := transfer.SpellMetrics[character.UnitIndex]
	if healed.Ticks != ticks || healed.TotalHealing < float64(ticks)*heal.Min() || healed.TotalHealing > float64(ticks)*heal.Max() {
		t.Errorf("the brand healed the wearer %v over %d ticks, want %d ticks of the row's %v to %v",
			healed.TotalHealing, healed.Ticks, ticks, heal.Min(), heal.Max())
	}

	dealHit(t, sim, character, &core.Spell{ProcMask: core.ProcMaskMeleeMHSpecial, DefenseType: core.DefenseTypeMelee, Unit: &character.Unit})
	if got := brand.SpellMetrics[enemy.UnitIndex].Casts; got != 2 {
		t.Errorf("a main-hand special cast %d brands in all, want 2", got)
	}
	stepTo(t, sim, sim.CurrentTime+drain.Period())
	character.OnSpellHitDealt(sim, meleeSwing(&character.Unit), &core.SpellResult{Target: enemy, Outcome: core.OutcomeHit, Damage: 100})
	if got := brand.Dot(enemy).RemainingDuration(sim); got != row.Duration() {
		t.Errorf("a proc on a branded target leaves the brand %v, want the row's full %v", got, row.Duration())
	}
}

// Barovian Family Sword 14541 casts Siphon Health 18652 at the procs per minute of the item's entry, on
// the sword's speed, and the drain goes on a target above the wearer's level with no hit roll. It ticks
// every period for the row's duration, and each tick casts the heal the drain's row links to on the
// wearer, for the amount the drain's row states whatever the tick dealt. The heal crits on a roll of
// its own.
func TestBarovianFamilySwordHealsTheWearerWithEveryTick(t *testing.T) {
	sim, character := itemWearerAgainst(t, proto.Class_ClassWarrior, &proto.EquipmentSpec{Items: weaponSlots(14541, 0)},
		&proto.Target{Name: "target", Level: 63})
	character.AutoAttacks.CancelAutoSwing(sim)
	row := spelldata.MustFind(18652)
	drain := row.PeriodicDamageEffect()
	proc := character.GetAura("Barovian Family Sword")
	siphon := character.GetSpell(core.ActionID{SpellID: row.ID})
	if proc == nil || proc.Dpm == nil || siphon == nil {
		t.Fatal("14541 registers no drain or no proc on a procs per minute rate")
	}
	heal := character.GetSpell(core.ActionID{SpellID: row.Triggered()[0].ID})
	if heal == nil {
		t.Fatal("14541 registers no heal for the spell its drain links to")
	}
	expectWeaponRate(t, sim, proc, core.ProcMaskMeleeMHAuto, 14541, 18652)
	ticks := rowTickCount(row, drain)
	amount := drain.Min()
	if ticks < 2 || amount <= 0 || amount != drain.Max() {
		t.Fatalf("18652 drains %v to %v %d times, want one amount at least twice", amount, drain.Max(), ticks)
	}

	// Every main-hand hit now procs it, nothing crits, and the target takes half of each tick.
	*proc.Dpm = *character.NewFixedProcChanceManager(1, core.ProcMaskMeleeMH)
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: -character.GetStat(stats.SpellCritPercent)})
	character.RemoveHealth(sim, character.MaxHealth()/2)
	enemy := character.CurrentTarget
	enemy.PseudoStats.DamageTakenMultiplier *= 0.5

	dealHit(t, sim, character, meleeSwing(&character.Unit))
	stepTo(t, sim, sim.CurrentTime+row.Duration()+time.Millisecond)
	drained := siphon.SpellMetrics[enemy.UnitIndex]
	if drained.Ticks != ticks || drained.TotalDamage <= 0 || drained.TotalDamage > float64(ticks)*amount/2 {
		t.Fatalf("one proc drained %v over %d ticks, want %d ticks of at most half the row's %v", drained.TotalDamage, drained.Ticks, ticks, amount)
	}
	healed := heal.SpellMetrics[character.UnitIndex]
	if healed.Casts != ticks || healed.Crits != 0 || healed.TotalHealing != float64(ticks)*amount {
		t.Errorf("the drain healed the wearer %v over %d heals, %d of them crits, want %d heals of the row's %v with no crit",
			healed.TotalHealing, healed.Casts, healed.Crits, ticks, amount)
	}

	// The sim rolls the heal's crit on the wearer's spell crit.
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: 100})
	dealHit(t, sim, character, meleeSwing(&character.Unit))
	stepTo(t, sim, sim.CurrentTime+row.Duration()+time.Millisecond)
	healed = heal.SpellMetrics[character.UnitIndex]
	if healed.Casts != 2*ticks || healed.Crits != ticks || healed.TotalHealing <= 2*float64(ticks)*amount {
		t.Errorf("with every spell a crit the second drain brought the heals to %d, %d of them crits, for %v, want %d more that all crit for more than the row's %v each",
			healed.Casts, healed.Crits, healed.TotalHealing, ticks, amount)
	}

	for range 200 {
		siphon.Cast(sim, enemy)
	}
	if cast := siphon.SpellMetrics[enemy.UnitIndex]; cast.Misses != 0 || cast.Hits != cast.Casts {
		t.Errorf("%d of %d drains missed a target above the wearer's level, want every one to land", cast.Misses, cast.Casts)
	}
}

// Malown's Slam 13393 puts Surge of Strength 17499 on the wearer at the procs per minute of the item's
// entry for 17500, on the mace's speed, off landed main-hand swings and specials and never off a miss.
// The buff raises Strength by the row's amount for the row's duration.
func TestMalownsSlamRaisesStrengthOnMainHandHits(t *testing.T) {
	sim, character := itemWearerWithEquipment(t, proto.Class_ClassWarrior, &proto.EquipmentSpec{Items: weaponSlots(13393, 0)})
	character.AutoAttacks.CancelAutoSwing(sim)
	proc := character.GetAura("Malown's Slam")
	surge := character.GetAuraByID(core.ActionID{SpellID: 17499})
	if proc == nil || proc.Dpm == nil || surge == nil {
		t.Fatal("13393 registers no Surge of Strength or no proc on a procs per minute rate")
	}
	expectWeaponRate(t, sim, proc, core.ProcMaskMeleeMHAuto, 13393, 17500)
	expectWeaponRate(t, sim, proc, core.ProcMaskMeleeMHSpecial, 13393, 17500)

	// Every main-hand hit now procs it.
	*proc.Dpm = *character.NewFixedProcChanceManager(1, core.ProcMaskMeleeMH)
	row := spelldata.MustFind(17499)
	amount := row.FirstAura(dbcenums.A_MOD_STAT).Min()
	before := character.GetStats()

	character.OnSpellHitDealt(sim, meleeSwing(&character.Unit), &core.SpellResult{Target: character.CurrentTarget, Outcome: core.OutcomeMiss})
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
	if surge.IsActive() {
		t.Error("a missed swing put up Surge of Strength")
	}
	dealHit(t, sim, character, meleeSwing(&character.Unit))
	if !surge.IsActive() {
		t.Fatal("a main-hand swing put up no Surge of Strength")
	}
	if got := gained(character, before, stats.Strength); amount <= 0 || !statNear(got, amount) {
		t.Errorf("Surge of Strength adds %v Strength, want the row's %v", got, amount)
	}
	if got := surge.RemainingDuration(sim); got != row.Duration() {
		t.Errorf("Surge of Strength went up for %v, want the row's %v", got, row.Duration())
	}
}

// Hand of Edward the Odd 2243 procs Focus 18803 at the procs per minute of the item's chance-on-hit
// entry, which the test states. Focus makes the next spell with a cast time instant, on its whole
// global cooldown, and that cast spends it; an instant leaves it up, and unspent it fades after the
// row's duration.
func TestHandOfEdwardTheOddMakesTheNextCastInstant(t *testing.T) {
	if core.GetItemByID(2243) == nil {
		t.Skip("item 2243 is not in the live game files; this runs again once they hold it")
	}
	withEntryPPM(t, 2243, 18803, 1)
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

// The shortest cast time Blade of Eternal Darkness 17780 and Moonshadow Stave 22458 price their chance
// on (EV-0025, EV-0115).
const castTimeFloor = 1500 * time.Millisecond

// The chance a trigger priced on cast time rolls on a hit whose cast took castTime: the row's procs per
// minute on that time, at least castTimeFloor.
func castTimeChance(t *testing.T, rowID int32, castTime time.Duration) float64 {
	t.Helper()
	ppm := float64(spelldata.MustFind(rowID).RPPM)
	if ppm <= 0 {
		t.Fatalf("%d states no procs per minute, nothing to check", rowID)
	}
	return ppm * max(castTime, castTimeFloor).Seconds() / 60
}

// The caster stand-in wielding Blade of Eternal Darkness 17780 against the given enemies, with no auto
// attacks, and the spell its proc casts.
func boedWearer(t *testing.T, targets int) (*core.Simulation, *core.Character, *core.Spell) {
	t.Helper()
	sim, character := weaponWearerFor(t, proto.Class_ClassShaman, 17780, targets, 7200)
	character.AutoAttacks.CancelAutoSwing(sim)
	proc := character.GetAura("Blade of Eternal Darkness")
	shadows := character.GetSpell(core.ActionID{SpellID: 27860})
	if proc == nil || proc.Dpm != nil || shadows == nil {
		t.Fatal("17780 registers no 27860, or no proc that rolls its own chance")
	}
	return sim, character, shadows
}

// How many times the proc has cast its spell, at any enemy.
func procCasts(sim *core.Simulation, spell *core.Spell) int32 {
	var casts int32
	for _, enemy := range sim.Encounter.ActiveTargetUnits {
		casts += spell.SpellMetrics[enemy.UnitIndex].Casts
	}
	return casts
}

// One landing of a spell cast in castTime on every enemy at once, as an area spell's. Each result is
// dealt at the cast, and the hits land once the next cast, an instant, has started. The sim then steps
// past the cooldown of the trigger's row.
func landCastHits(t *testing.T, sim *core.Simulation, character *core.Character, spell *core.Spell, castTime time.Duration, rowID int32) {
	t.Helper()
	enemies := sim.Encounter.ActiveTargetUnits
	results := make([]*core.SpellResult, len(enemies))
	spell.CurCast.CastTime = castTime
	for i, enemy := range enemies {
		results[i] = &core.SpellResult{Target: enemy, Outcome: core.OutcomeHit, Damage: 100}
		character.OnCastResult(sim, spell, results[i])
	}
	spell.CurCast.CastTime = 0
	for _, result := range results {
		character.OnSpellHitDealt(sim, spell, result)
	}
	stepTo(t, sim, sim.CurrentTime+spelldata.MustFind(rowID).ICD())
}

// Blade of Eternal Darkness 17780 rolls each landed spell hit at 21978's procs per minute on the cast
// time of the hit's cast, at least castTimeFloor: an instant on the floor, a long cast on its own time.
// A missile keeps the time of the cast that launched it after the next cast has started.
func TestBladeOfEternalDarknessPricesEachHitOnItsCastTime(t *testing.T) {
	sim, character, shadows := boedWearer(t, 1)
	spell := fireSpell(&character.Unit)

	const landings = 6000
	for _, castTime := range []time.Duration{0, 2 * castTimeFloor} {
		before := procCasts(sim, shadows)
		for range landings {
			landCastHits(t, sim, character, spell, castTime, 21978)
		}
		want := castTimeChance(t, 21978, castTime)
		mean, deviation := want*landings, math.Sqrt(want*(1-want)*landings)
		if got := float64(procCasts(sim, shadows) - before); math.Abs(got-mean) > 4*deviation {
			t.Errorf("%d hits of a %v cast procced it %v times, want about %.0f", landings, castTime, got, mean)
		}
	}
}

// A long cast made instant, as Presence of Mind makes one, rolls Blade of Eternal Darkness 17780 on the
// floor's chance, and the same cast as a hardcast on its own time.
func TestBladeOfEternalDarknessPricesAnInstantLongCastOnTheFloor(t *testing.T) {
	sim, character, shadows := boedWearer(t, 1)
	enemy := sim.Encounter.ActiveTargetUnits[0]
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100})
	castTime := character.GetSpell(core.ActionID{SpellID: castTimeSpellID})
	// Long enough that a hardcast procs on every hit.
	castTime.DefaultCast.CastTime = time.Minute
	if castTimeChance(t, 21978, castTime.DefaultCast.CastTime) < 1 {
		t.Fatalf("a %v cast rolls %v, want a certain proc", castTime.DefaultCast.CastTime, castTimeChance(t, 21978, castTime.DefaultCast.CastTime))
	}

	castTime.Cast(sim, enemy)
	stepTo(t, sim, character.Hardcast.Expires+time.Millisecond)
	if got := procCasts(sim, shadows); got != 1 {
		t.Errorf("a hardcast hit procced it %d times, want once", got)
	}

	instant := character.AddDynamicMod(core.SpellModConfig{Kind: core.SpellMod_CastTime_Pct, FloatValue: -1})
	instant.Activate()
	const casts = 300
	for range casts {
		stepTo(t, sim, max(character.GCD.ReadyAt(), sim.CurrentTime+spelldata.MustFind(21978).ICD()))
		castTime.Cast(sim, enemy)
	}
	want := castTimeChance(t, 21978, 0)
	mean, deviation := want*casts, math.Sqrt(want*(1-want)*casts)
	if got := float64(procCasts(sim, shadows) - 1); math.Abs(got-mean) > 4*deviation {
		t.Errorf("%d instant hits of the long cast procced it %v times, want about %.0f", casts, got, mean)
	}
}

// Each enemy an area spell lands on rolls Blade of Eternal Darkness 17780, and the row's cooldown lets
// one landing proc it once.
func TestBladeOfEternalDarknessRollsEachEnemyOfALandingOnce(t *testing.T) {
	const enemies = 5
	sim, character, shadows := boedWearer(t, enemies)
	spell := fireSpell(&character.Unit)

	const certain = 50
	for range certain {
		landCastHits(t, sim, character, spell, time.Hour, 21978)
	}
	if got := procCasts(sim, shadows); got != certain {
		t.Errorf("%d landings on %d enemies at a certain chance procced it %d times, want once a landing", certain, enemies, got)
	}

	const landings = 3000
	for range landings {
		landCastHits(t, sim, character, spell, 0, 21978)
	}
	want := 1 - math.Pow(1-castTimeChance(t, 21978, 0), enemies)
	mean, deviation := want*landings, math.Sqrt(want*(1-want)*landings)
	if got := float64(procCasts(sim, shadows) - certain); math.Abs(got-mean) > 4*deviation {
		t.Errorf("%d landings on %d enemies procced it %v times, want about %.0f", landings, enemies, got, mean)
	}
}

// A landed harmful spell rolls Blade of Eternal Darkness 17780 whether or not it deals damage, a
// spell with no proc mask too when its DefenseType is magic, its school is not Physical and the
// wearer cast it. A damage over time tick, a wand shot, a miss, a melee ability or a Physical taunt
// with no proc mask, a spell fired as a proc or cast by an item with no proc mask and a landing on the
// wearer never roll it. Each is dealt at an hour-long cast, a chance that cannot fail, as core deals a cast's result
// before its hit.
func TestBladeOfEternalDarknessRollsLandedHarmfulSpellsOnly(t *testing.T) {
	for _, event := range []struct {
		name string
		want int32
		deal func(*core.Simulation, *core.Character, *core.Spell, *core.SpellResult)
	}{
		{"spell landing that deals no damage", 1, func(sim *core.Simulation, character *core.Character, spell *core.Spell, result *core.SpellResult) {
			character.OnSpellHitDealt(sim, spell, result)
		}},
		{"magic spell with no proc mask", 1, func(sim *core.Simulation, character *core.Character, spell *core.Spell, result *core.SpellResult) {
			spell.ProcMask = core.ProcMaskEmpty
			character.OnSpellHitDealt(sim, spell, result)
		}},
		{"damage over time tick", 0, func(sim *core.Simulation, character *core.Character, spell *core.Spell, result *core.SpellResult) {
			result.Damage = 100
			character.OnPeriodicDamageDealt(sim, spell, result)
		}},
		{"wand shot", 0, func(sim *core.Simulation, character *core.Character, _ *core.Spell, result *core.SpellResult) {
			shot := rangedShot(&character.Unit)
			shot.DefenseType = core.DefenseTypeMagic
			shot.CurCast.CastTime = time.Hour
			character.OnCastResult(sim, shot, result)
			character.OnSpellHitDealt(sim, shot, result)
		}},
		{"miss", 0, func(sim *core.Simulation, character *core.Character, spell *core.Spell, result *core.SpellResult) {
			result.Outcome = core.OutcomeMiss
			character.OnSpellHitDealt(sim, spell, result)
		}},
		{"melee ability with no proc mask", 0, func(sim *core.Simulation, character *core.Character, spell *core.Spell, result *core.SpellResult) {
			spell.ProcMask = core.ProcMaskEmpty
			spell.DefenseType = core.DefenseTypeMelee
			character.OnSpellHitDealt(sim, spell, result)
		}},
		{"Physical taunt with no proc mask", 0, func(sim *core.Simulation, character *core.Character, spell *core.Spell, result *core.SpellResult) {
			spell.ProcMask = core.ProcMaskEmpty
			spell.SpellSchool = core.SpellSchoolPhysical
			character.OnSpellHitDealt(sim, spell, result)
		}},
		{"spell fired as a proc with no proc mask", 0, func(sim *core.Simulation, character *core.Character, spell *core.Spell, result *core.SpellResult) {
			spell.ProcMask = core.ProcMaskEmpty
			spell.Flags |= core.SpellFlagPassiveSpell
			character.OnSpellHitDealt(sim, spell, result)
		}},
		{"spell an item casts with no proc mask", 0, func(sim *core.Simulation, character *core.Character, spell *core.Spell, result *core.SpellResult) {
			spell.ProcMask = core.ProcMaskEmpty
			spell.ActionID = core.ActionID{ItemID: 10646}
			character.OnSpellHitDealt(sim, spell, result)
		}},
		{"landing on the wearer", 0, func(sim *core.Simulation, character *core.Character, spell *core.Spell, result *core.SpellResult) {
			result.Target = &character.Unit
			character.OnSpellHitDealt(sim, spell, result)
		}},
	} {
		t.Run(event.name, func(t *testing.T) {
			sim, character, shadows := boedWearer(t, 1)
			spell := fireSpell(&character.Unit)
			result := &core.SpellResult{Target: character.CurrentTarget, Outcome: core.OutcomeHit}
			spell.CurCast.CastTime = time.Hour
			character.OnCastResult(sim, spell, result)
			event.deal(sim, character, spell, result)
			if got := procCasts(sim, shadows); got != event.want {
				t.Errorf("a %s procced it %d times, want %d", event.name, got, event.want)
			}
		})
	}
}

// An untalented Troll priest wielding Blade of Eternal Darkness 17780 against one level 60 target for
// two hours, its spells free and certain to land, and the spell the proc casts.
func boedPriest(t *testing.T) (*core.Simulation, *priest.Priest, *core.Spell) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 17780
	player := &proto.Player{
		Name: "Priest", Class: proto.Class_ClassPriest, Race: proto.Race_RaceTroll,
		Equipment:   &proto.EquipmentSpec{Items: items},
		Consumables: &proto.ConsumesSpec{},
		Buffs:       &proto.IndividualBuffs{},
		Rotation:    &proto.APLRotation{},
		Spec: &proto.Player_DpsPriest{DpsPriest: &proto.DpsPriest{
			Options: &proto.DpsPriest_Options{ClassOptions: &proto.PriestOptions{}}}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}},
		Encounter:  &proto.Encounter{Duration: 7200, Targets: []*proto.Target{{Name: "target", Level: 60}}},
	}, simsignals.CreateSignals())
	sim.Reset()
	wearer := sim.Raid.Parties[0].Players[0].(priest.PriestAgent).GetPriest()
	wearer.AutoAttacks.CancelAutoSwing(sim)
	wearer.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100})
	wearer.AddDynamicMod(core.SpellModConfig{Kind: core.SpellMod_PowerCost_Pct, FloatValue: -1}).Activate()
	shadows := wearer.GetSpell(core.ActionID{SpellID: 27860})
	if shadows == nil {
		t.Fatal("17780 registers no 27860")
	}
	return sim, wearer, shadows
}

// Hex of Weakness 9035, which deals no damage and has no proc mask, Shadow Word: Pain 589, whose
// landing deals none, and Holy Fire 14914, which deals damage and applies its damage over time effect,
// each roll Blade of Eternal Darkness 17780 once a landing at the chance of their cast time; the
// ticks roll nothing.
func TestBladeOfEternalDarknessRollsEachLandingOfAHexOrDamageOverTimeOnce(t *testing.T) {
	for _, id := range []int32{9035, 589, 14914} {
		t.Run(fmt.Sprint(id), func(t *testing.T) {
			sim, wearer, shadows := boedPriest(t)
			enemy := sim.Encounter.ActiveTargetUnits[0]
			spell := wearer.GetSpell(core.ActionID{SpellID: id})
			if spell == nil {
				t.Fatalf("the priest registers no %d", id)
			}

			const casts = 1500
			for range casts {
				stepTo(t, sim, max(wearer.GCD.ReadyAt(), wearer.Hardcast.Expires+time.Millisecond, sim.CurrentTime+spelldata.MustFind(21978).ICD()))
				spell.Cast(sim, enemy)
			}
			stepTo(t, sim, wearer.Hardcast.Expires+time.Millisecond)
			if got := spell.SpellMetrics[enemy.UnitIndex].Casts; got != casts {
				t.Fatalf("%d casts completed, want %d", got, casts)
			}
			want := castTimeChance(t, 21978, spell.DefaultCast.CastTime)
			mean, deviation := want*casts, math.Sqrt(want*(1-want)*casts)
			if got := float64(procCasts(sim, shadows)); math.Abs(got-mean) > 4*deviation {
				t.Errorf("%d landings of a %v cast procced it %v times, want about %.0f", casts, spell.DefaultCast.CastTime, got, mean)
			}
		})
	}
}

// Each Blade of Eternal Darkness 17780 proc restores 27860's mana to the wearer.
func TestBladeOfEternalDarknessRestoresManaWithEachProc(t *testing.T) {
	sim, character := itemWearer(t, proto.Class_ClassPaladin, proto.ItemSlot_ItemSlotMainHand, 17780)
	want := spelldata.MustFind(27860).EnergizeEffect().Average()
	if want <= 0 {
		t.Fatal("27860 restores no mana, nothing to check")
	}
	character.SpendMana(sim, character.CurrentMana(), character.NewManaMetrics(core.ActionID{SpellID: instantSpellID}))

	const procs = 3
	for range procs {
		landCastHits(t, sim, character, fireSpell(&character.Unit), time.Hour, 21978)
	}
	var gain float64
	var events int32
	for _, r := range character.Metrics.ToProto().Resources {
		if r.Id.GetSpellId() == 27860 {
			gain, events = r.Gain, r.Events
		}
	}
	if events != procs || math.Abs(gain-procs*want) > 1e-6 {
		t.Errorf("%d procs restored %v mana over %d gains, want %v each", procs, gain, events, want)
	}
}

// Moonshadow Stave 22458 rolls 1300435 on each landed harmful spell at the row's procs per minute on the
// cast time of the spell's cast, at least castTimeFloor, a landing that deals no damage too. A miss, a
// damage over time tick and a spell fired as a proc whose row is one never roll. Each proc casts
// Shadow 1300437 or Arcane 1300438, one in two each.
func TestMoonshadowStaveRollsLandedHarmfulSpellsOnTheirCastTime(t *testing.T) {
	const hurricaneTick, flamestrikeTick = 1278759, 1279990
	if !spelldata.MustFind(hurricaneTick).IsAProc() || spelldata.MustFind(flamestrikeTick).IsAProc() {
		t.Fatalf("want %d's row a proc and %d's not one", hurricaneTick, flamestrikeTick)
	}
	wearer := func(t *testing.T) (*core.Simulation, *core.Character, [2]*core.Spell) {
		t.Helper()
		sim, character := weaponWearerFor(t, proto.Class_ClassDruid, 22458, 1, 7200)
		character.AutoAttacks.CancelAutoSwing(sim)
		proc := character.GetAura("Moonshadow Stave")
		bolts := [2]*core.Spell{character.GetSpell(core.ActionID{SpellID: 1300437}), character.GetSpell(core.ActionID{SpellID: 1300438})}
		if proc == nil || proc.Dpm != nil || bolts[0] == nil || bolts[1] == nil {
			t.Fatal("22458 registers no 1300437 or 1300438, or no proc that rolls its own chance")
		}
		return sim, character, bolts
	}
	procs := func(sim *core.Simulation, bolts [2]*core.Spell) int32 {
		return procCasts(sim, bolts[0]) + procCasts(sim, bolts[1])
	}

	t.Run("cast time", func(t *testing.T) {
		sim, character, bolts := wearer(t)
		spell := fireSpell(&character.Unit)
		const landings = 6000
		for _, castTime := range []time.Duration{0, 2 * castTimeFloor} {
			before := procs(sim, bolts)
			for range landings {
				landCastHits(t, sim, character, spell, castTime, 1300435)
			}
			want := castTimeChance(t, 1300435, castTime)
			mean, deviation := want*landings, math.Sqrt(want*(1-want)*landings)
			if got := float64(procs(sim, bolts) - before); math.Abs(got-mean) > 4*deviation {
				t.Errorf("%d hits of a %v cast procced it %v times, want about %.0f", landings, castTime, got, mean)
			}
		}
	})

	t.Run("school", func(t *testing.T) {
		sim, character, bolts := wearer(t)
		const certain = 400
		for range certain {
			landCastHits(t, sim, character, fireSpell(&character.Unit), time.Hour, 1300435)
		}
		shadow, arcane := procCasts(sim, bolts[0]), procCasts(sim, bolts[1])
		if shadow+arcane != certain || math.Abs(float64(shadow)-certain/2) > 4*math.Sqrt(certain/4) {
			t.Errorf("%d certain procs cast 1300437 %d times and 1300438 %d times, want about half each", certain, shadow, arcane)
		}
	})

	// Each result is dealt at an hour-long cast, a chance that cannot fail, as core deals a cast's result
	// before its hit.
	for _, event := range []struct {
		name string
		want int32
		deal func(*core.Simulation, *core.Character, *core.Spell, *core.SpellResult)
	}{
		{"debuff application that deals no damage", 1, func(sim *core.Simulation, character *core.Character, spell *core.Spell, result *core.SpellResult) {
			character.OnSpellHitDealt(sim, spell, result)
		}},
		{"miss", 0, func(sim *core.Simulation, character *core.Character, spell *core.Spell, result *core.SpellResult) {
			result.Outcome = core.OutcomeMiss
			character.OnSpellHitDealt(sim, spell, result)
		}},
		{"damage over time tick", 0, func(sim *core.Simulation, character *core.Character, spell *core.Spell, result *core.SpellResult) {
			result.Damage = 100
			character.OnPeriodicDamageDealt(sim, spell, result)
		}},
		{"spell fired as a proc whose row is one", 0, func(sim *core.Simulation, character *core.Character, spell *core.Spell, result *core.SpellResult) {
			spell.Flags |= core.SpellFlagPassiveSpell
			spell.ActionID = core.ActionID{SpellID: hurricaneTick}
			character.OnSpellHitDealt(sim, spell, result)
		}},
		{"spell fired as a proc whose row is not one", 1, func(sim *core.Simulation, character *core.Character, spell *core.Spell, result *core.SpellResult) {
			spell.Flags |= core.SpellFlagPassiveSpell
			spell.ActionID = core.ActionID{SpellID: flamestrikeTick}
			character.OnSpellHitDealt(sim, spell, result)
		}},
	} {
		t.Run(event.name, func(t *testing.T) {
			sim, character, bolts := wearer(t)
			spell := fireSpell(&character.Unit)
			result := &core.SpellResult{Target: character.CurrentTarget, Outcome: core.OutcomeHit}
			spell.CurCast.CastTime = time.Hour
			character.OnCastResult(sim, spell, result)
			event.deal(sim, character, spell, result)
			if got := procs(sim, bolts); got != event.want {
				t.Errorf("a %s procced it %d times, want %d", event.name, got, event.want)
			}
		})
	}
}

// A hand-written proc that runs at its item entry's procs per minute registers nothing while the entry
// states none: Hand of Edward the Odd 2243 equips and swings with no Focus proc and no Focus aura.
func TestAnEntryRateProcRegistersNothingWhileTheEntryStatesNoRate(t *testing.T) {
	if core.GetItemByID(2243) == nil {
		t.Skip("item 2243 is not in the live game files; this runs again once they hold it")
	}
	withEntryPPM(t, 2243, 18803, 0)
	sim, character := weaponWearer(t, proto.Class_ClassShaman, 2243, 1)

	if got := character.Equipment[proto.ItemSlot_ItemSlotMainHand].ID; got != 2243 {
		t.Fatalf("the main hand holds %d, want 2243", got)
	}
	if character.GetAura("Hand of Edward the Odd") != nil {
		t.Error("2243 registers its Focus proc with no rate on its entry, want none")
	}
	if character.GetAura("Hand of Edward the Odd Focus") != nil {
		t.Error("2243 registers its Focus aura with no rate on its entry, want none")
	}
	stepTo(t, sim, 30*time.Second)
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

// Dragon's Call 10847 casts 13049 at the procs per minute of the item's chance-on-hit entry, on the
// sword's speed, off any landed main-hand hit, an application that deals no damage included. A miss
// casts nothing, and no hit rolls during the row's category cooldown, which starts at the proc.
func TestDragonsCallCastsOnLandedHitsOutsideItsCooldown(t *testing.T) {
	sim, character := weaponWearer(t, proto.Class_ClassWarrior, 10847, 1)
	character.AutoAttacks.CancelAutoSwing(sim)
	proc := character.GetAura("Dragon's Call")
	summon := character.GetSpell(core.ActionID{SpellID: 13049})
	if proc == nil || proc.Dpm == nil || summon == nil {
		t.Fatal("10847 registers no 13049 or no proc on a procs per minute rate")
	}
	expectWeaponRate(t, sim, proc, core.ProcMaskMeleeMHAuto, 10847, 13049)
	expectWeaponRate(t, sim, proc, core.ProcMaskMeleeMHSpecial, 10847, 13049)
	cooldown := spelldata.MustFind(13049).CategoryCooldown()
	if cooldown <= 0 {
		t.Fatalf("13049 states a category cooldown of %v, want one", cooldown)
	}

	// Every landed melee hit outside the cooldown now procs it.
	*proc.Dpm = *character.NewFixedProcChanceManager(1, core.ProcMaskMelee)
	enemy := character.CurrentTarget
	casts := func() int32 { return summon.SpellMetrics[enemy.UnitIndex].Casts }
	hit := func(outcome core.HitOutcome, damage float64) {
		character.OnSpellHitDealt(sim, &core.Spell{ProcMask: core.ProcMaskMeleeMHSpecial, DefenseType: core.DefenseTypeMelee, Unit: &character.Unit},
			&core.SpellResult{Target: enemy, Outcome: outcome, Damage: damage})
	}

	hit(core.OutcomeMiss, 0)
	if got := casts(); got != 0 {
		t.Errorf("a miss cast it %d times, want none", got)
	}
	start := sim.CurrentTime
	hit(core.OutcomeHit, 0)
	if got := casts(); got != 1 {
		t.Errorf("a landed application that dealt no damage cast it %d times, want once", got)
	}
	stepTo(t, sim, start+cooldown-time.Millisecond)
	hit(core.OutcomeHit, 100)
	if got := casts(); got != 1 {
		t.Errorf("a hit 1 ms before the cooldown ends cast it %d times in all, want once still", got)
	}
	stepTo(t, sim, start+cooldown)
	hit(core.OutcomeHit, 100)
	if got := casts(); got != 2 {
		t.Errorf("a hit as the cooldown ends cast it %d times in all, want twice", got)
	}
}

// The Emerald Dragon Whelp the wearer of Dragon's Call 10847 owns, or nil.
func emeraldDragonWhelp(character *core.Character) *core.Pet {
	for _, pet := range character.Pets {
		if pet.Name == "Emerald Dragon Whelp" {
			return pet
		}
	}
	return nil
}

// Dragon's Call's 13049 summons the Emerald Dragon Whelp on the enemy for the row's duration, and the
// whelp swings and casts Acid Spit 9591 while it is up.
func TestDragonsCallSummonsTheWhelpForTheRowsDuration(t *testing.T) {
	sim, character := weaponWearer(t, proto.Class_ClassWarrior, 10847, 1)
	character.AutoAttacks.CancelAutoSwing(sim)
	whelp := emeraldDragonWhelp(character)
	if whelp == nil {
		t.Fatal("the wearer of 10847 owns no Emerald Dragon Whelp")
	}
	enemy := character.CurrentTarget
	start := sim.CurrentTime
	character.GetSpell(core.ActionID{SpellID: 13049}).Cast(sim, enemy)
	duration := spelldata.MustFind(13049).Duration()
	stepTo(t, sim, start+duration-time.Millisecond)
	if !whelp.IsEnabled() {
		t.Fatalf("the whelp is gone 1 ms before the row's %v run out", duration)
	}
	if got := whelp.GetSpell(core.ActionID{SpellID: 9591}).SpellMetrics[enemy.UnitIndex].Casts; got == 0 {
		t.Error("the whelp cast no Acid Spit on the enemy")
	}
	if got := whelp.AutoAttacks.MHAuto().SpellMetrics[enemy.UnitIndex].Casts; got == 0 {
		t.Error("the whelp swung at the enemy no time")
	}
	stepTo(t, sim, start+duration)
	if whelp.IsEnabled() {
		t.Errorf("the whelp is still up as the row's %v run out", duration)
	}
}

// Acid Spit 9591 states its base points as a percent of the creature auto attack dps for the caster's
// level, which is the owner's. Its hits roll the row's variance, so their mean sits at that amount.
func TestAcidSpitDealsItsPercentOfTheCreatureDpsAtTheOwnersLevel(t *testing.T) {
	// Long enough for 100 summons a cooldown apart.
	sim, character := weaponWearerFor(t, proto.Class_ClassWarrior, 10847, 1, 5000)
	character.AutoAttacks.CancelAutoSwing(sim)
	whelp := emeraldDragonWhelp(character)
	enemy := character.CurrentTarget
	summon := character.GetSpell(core.ActionID{SpellID: 13049})
	for range 100 {
		stepTo(t, sim, summon.ReadyAt())
		summon.Cast(sim, enemy)
	}
	stepTo(t, sim, sim.CurrentTime+spelldata.MustFind(13049).Duration())

	metrics := whelp.GetSpell(core.ActionID{SpellID: 9591}).SpellMetrics[enemy.UnitIndex]
	want := spelldata.MustFind(9591).EffectN(1).BaseValue() / 100 * core.CreatureAutoAttackDpsMaxLevel
	got := (metrics.TotalDamage - metrics.TotalCritDamage) / float64(metrics.Hits)
	if metrics.Hits < 200 || math.Abs(got-want) > 0.015*want {
		t.Errorf("Acid Spit's %d hits average %.1f, want the row's percent of the creature dps at level %d, %.1f",
			metrics.Hits, got, core.CharacterLevel, want)
	}
}

// A character carrying Dragon's Call only in its item swap set owns the whelp too, and the proc's cast
// summons it.
func TestDragonsCallInTheSwapSetOwnsTheWhelp(t *testing.T) {
	sim, character := stonedWarrior(t, &proto.ConsumesSpec{}, weaponSlots(10847, 0))
	whelp := emeraldDragonWhelp(character)
	if whelp == nil {
		t.Fatal("a swap set holding 10847 gives its owner no Emerald Dragon Whelp")
	}
	character.GetSpell(core.ActionID{SpellID: 13049}).Cast(sim, character.CurrentTarget)
	if !whelp.IsEnabled() {
		t.Error("13049 cast from the swap set summons no whelp")
	}

	_, without := stonedWarrior(t, &proto.ConsumesSpec{}, nil)
	if emeraldDragonWhelp(without) != nil {
		t.Error("a character with no 10847 owns an Emerald Dragon Whelp")
	}
}

// The spells Shahram can land for the wearer; Fist of Shahram 16601 lands nothing.
var shahramSpellIDs = []int32{16596, 16597, 16598, 16599, 16600}

// Blackblade of Shahram 12592 casts 16602 at the procs per minute of the item's chance-on-hit entry,
// on the sword's speed, off main-hand hits. Each summon lands one of Shahram's spells, none of them
// before its row's hardcast has run, or nothing when he picks Fist of Shahram.
func TestBlackbladeOfShahramLandsOneSpellPerSummonAfterItsHardcast(t *testing.T) {
	sim, character := weaponWearerFor(t, proto.Class_ClassShaman, 12592, 1, 6000)
	character.AutoAttacks.CancelAutoSwing(sim)
	proc := character.GetAura("Blackblade of Shahram")
	summon := character.GetSpell(core.ActionID{SpellID: 16602})
	if proc == nil || proc.Dpm == nil || summon == nil {
		t.Fatal("12592 registers no 16602 or no proc on a procs per minute rate")
	}
	expectWeaponRate(t, sim, proc, core.ProcMaskMeleeMHAuto, 12592, 16602)
	expectWeaponRate(t, sim, proc, core.ProcMaskMeleeMHSpecial, 12592, 16602)

	enemy := character.CurrentTarget
	hardcast := time.Duration(math.MaxInt64)
	casts := func(id int32) int32 {
		return character.GetSpell(core.ActionID{SpellID: id}).SpellMetrics[enemy.UnitIndex].Casts
	}
	landed := func() (n int32) {
		for _, id := range shahramSpellIDs {
			n += casts(id)
		}
		return n
	}
	for _, id := range shahramSpellIDs {
		hardcast = min(hardcast, spelldata.MustFind(id).CastTime())
	}
	if hardcast <= 0 {
		t.Fatalf("a Shahram spell states a hardcast of %v, want one", hardcast)
	}

	// Each summon gets longer than the longest of his spells to land and run out.
	const summons, apart = 200, 25 * time.Second
	nothing := 0
	for range summons {
		start, before := sim.CurrentTime, landed()
		summon.Cast(sim, enemy)
		stepTo(t, sim, start+hardcast)
		if got := landed(); got != before {
			t.Fatalf("a summon landed %d spells within the shortest hardcast, %v, want none yet", got-before, hardcast)
		}
		stepTo(t, sim, start+apart)
		switch landed() - before {
		case 0:
			nothing++
		case 1:
		default:
			t.Fatalf("a summon landed %d spells, want one at most", landed()-before)
		}
	}
	for _, id := range shahramSpellIDs {
		if casts(id) == 0 {
			t.Errorf("%d summons never landed %d", summons, id)
		}
	}
	if nothing == 0 {
		t.Errorf("each of %d summons landed a spell, want some to pick Fist of Shahram and land nothing", summons)
	}
}

// Flames of Shahram 16596 burns every enemy. Shahram's spells take nothing from the wearer's damage
// modifiers or crit chance and put no threat on the wearer.
func TestFlamesOfShahramBurnsEveryEnemy(t *testing.T) {
	sim, character := weaponWearer(t, proto.Class_ClassWarrior, 12592, 3)
	character.AutoAttacks.CancelAutoSwing(sim)
	character.PseudoStats.DamageDealtMultiplier *= 2
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100, stats.SpellCritPercent: 100})
	flames := character.GetSpell(core.ActionID{SpellID: 16596})
	damage := spelldata.MustFind(16596).DamageEffect()
	if damage.Min() <= 0 || damage.Max() <= damage.Min() {
		t.Fatalf("16596 deals %v to %v, want a range", damage.Min(), damage.Max())
	}

	const casts = 100
	for range casts {
		flames.Cast(sim, character.CurrentTarget)
	}
	for _, enemy := range sim.Encounter.ActiveTargetUnits {
		metrics := flames.SpellMetrics[enemy.UnitIndex]
		average := metrics.TotalDamage / float64(metrics.Hits)
		if metrics.Hits != casts || metrics.Crits != 0 || average < damage.Min() || average > damage.Max() {
			t.Errorf("%d casts hit %s %d times with %d crits for %.1f on average, want every cast to hit with no crit within %v to %v",
				casts, enemy.Label, metrics.Hits, metrics.Crits, average, damage.Min(), damage.Max())
		}
		if metrics.TotalThreat != 0 {
			t.Errorf("the hits on %s put %v threat on the wearer, want none", enemy.Label, metrics.TotalThreat)
		}
	}
}

// The Will of Shahram 16598 aura on a wearer whose auto attacks are off, and a check that its stacks,
// and the row's stats once per stack, are what a moment of the test expects.
func willOfShahramWearer(t *testing.T) (*core.Simulation, *core.Character, *core.Spell, *core.Aura, func(string, int32)) {
	t.Helper()
	sim, character := weaponWearer(t, proto.Class_ClassWarrior, 12592, 1)
	character.AutoAttacks.CancelAutoSwing(sim)
	row := spelldata.MustFind(16598)
	amount := row.EffectN(1).Average()
	if amount <= 0 {
		t.Fatalf("16598 grants %v stats, want an amount", amount)
	}
	aura := character.GetAuraByID(core.ActionID{SpellID: 16598})
	if aura == nil {
		t.Fatal("the wearer has no aura of 16598")
	}
	base := character.GetStat(stats.Strength)
	expect := func(when string, stacks int32) {
		t.Helper()
		if got, want := character.GetStat(stats.Strength), base+float64(stacks)*amount; math.Abs(got-want) > 1e-6 {
			t.Errorf("%s the wearer's Strength is %v, want %v", when, got, want)
		}
		if aura.IsActive() != (stacks > 0) || aura.GetStacks() != stacks {
			t.Errorf("%s the aura is active %v with %d stacks, want %d", when, aura.IsActive(), aura.GetStacks(), stacks)
		}
	}
	return sim, character, character.GetSpell(core.ActionID{SpellID: 16598}), aura, expect
}

// Each Will of Shahram 16598 adds a stack worth the row's stats to one aura. Each stack fades on its
// own timer, the aura stays up while any stack does, and its remaining time is the latest stack's.
func TestWillOfShahramStacksAndFadeOneByOne(t *testing.T) {
	sim, character, will, aura, expect := willOfShahramWearer(t)
	duration := spelldata.MustFind(16598).Duration()
	if duration <= 0 {
		t.Fatalf("16598 lasts %v, want a duration", duration)
	}

	start := sim.CurrentTime
	will.Cast(sim, character.CurrentTarget)
	expect("after one Will", 1)
	stepTo(t, sim, start+duration/2)
	will.Cast(sim, character.CurrentTarget)
	expect("after a second Will", 2)
	if got := aura.RemainingDuration(sim); got != duration {
		t.Errorf("after a second Will the aura has %v left, want the second Will's %v", got, duration)
	}
	stepTo(t, sim, start+duration+time.Millisecond)
	expect("past the first Will's duration", 1)
	stepTo(t, sim, start+duration/2+duration+time.Millisecond)
	expect("past the second Will's duration", 0)

	will.Cast(sim, character.CurrentTarget)
	expect("after a Will once every stack has faded", 1)
}

// Will of Shahram 16598 forgets its stacks when its aura ends, so a stack that lands afterwards fades
// on its own timer, and not on one of the stacks that ended with the aura.
func TestWillOfShahramForgetsItsStacksWhenTheAuraEnds(t *testing.T) {
	sim, character, will, aura, expect := willOfShahramWearer(t)
	duration := spelldata.MustFind(16598).Duration()
	if duration <= 0 {
		t.Fatalf("16598 lasts %v, want a duration", duration)
	}

	start := sim.CurrentTime
	will.Cast(sim, character.CurrentTarget)
	stepTo(t, sim, start+duration/2)
	will.Cast(sim, character.CurrentTarget)
	aura.Deactivate(sim)
	expect("after the aura ends", 0)
	will.Cast(sim, character.CurrentTarget)
	expect("after a Will once the aura has ended", 1)
	stepTo(t, sim, start+duration+time.Millisecond)
	expect("past the duration of the first Will before the aura ended", 1)
	stepTo(t, sim, start+duration/2+duration+time.Millisecond)
	expect("past the new Will's duration", 0)
}

// With every stack of Will of Shahram 16598 up, a new Will replaces the stack closest to fading, so the
// bonus stays at the most stacks the wearer can carry and the aura lasts until the new stack fades.
func TestWillOfShahramReplacesTheOldestStackWhenAllAreUp(t *testing.T) {
	sim, character, will, aura, expect := willOfShahramWearer(t)
	duration, full := spelldata.MustFind(16598).Duration(), aura.MaxStacks
	if full < 2 || duration <= time.Duration(full)*time.Second {
		t.Fatalf("16598 lasts %v and the aura holds %d stacks, want more than one stack, each lasting longer than a second per stack", duration, full)
	}

	start := sim.CurrentTime
	for i := range full {
		stepTo(t, sim, start+time.Duration(i)*time.Second)
		will.Cast(sim, character.CurrentTarget)
	}
	expect("with every stack up", full)
	stepTo(t, sim, start+time.Duration(full)*time.Second)
	will.Cast(sim, character.CurrentTarget)
	expect("after one Will more than the aura holds", full)
	if got := aura.RemainingDuration(sim); got != duration {
		t.Errorf("after one Will more than the aura holds it has %v left, want the new Will's %v", got, duration)
	}
	stepTo(t, sim, start+duration+500*time.Millisecond)
	expect("past the first stack's duration", full)
	stepTo(t, sim, start+duration+time.Second+500*time.Millisecond)
	expect("past the second stack's duration", full-1)
}

// Curse of Shahram 16597 slows the attacks of every enemy by the row's share for the row's duration.
func TestCurseOfShahramSlowsEveryEnemysAttacks(t *testing.T) {
	sim, character := weaponWearer(t, proto.Class_ClassWarrior, 12592, 2)
	character.AutoAttacks.CancelAutoSwing(sim)
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100})
	row := spelldata.MustFind(16597)
	slowed := 1 / core.MultiplierFromPercent(-row.FirstAura(dbcenums.A_MOD_MELEE_HASTE_3).Average())
	if slowed >= 1 || row.Duration() <= 0 {
		t.Fatalf("16597 leaves attacks at %v of their speed for %v, want a slow", slowed, row.Duration())
	}
	enemies := sim.Encounter.ActiveTargetUnits
	expect := func(when string, want float64) {
		t.Helper()
		for _, enemy := range enemies {
			if got := enemy.TotalMeleeHasteMultiplier(); math.Abs(got-want) > 1e-9 {
				t.Errorf("%s %s attacks at %v of its speed, want %v", when, enemy.Label, got, want)
			}
		}
	}

	start := sim.CurrentTime
	character.GetSpell(core.ActionID{SpellID: 16597}).Cast(sim, character.CurrentTarget)
	expect("cursed,", slowed)
	stepTo(t, sim, start+row.Duration()+time.Millisecond)
	expect("past the row's duration", 1)
}

// Might of Shahram 16600 stuns every enemy for the row's duration, and a stun-immune enemy not at all.
func TestMightOfShahramStunsEveryEnemyThatCanBeStunned(t *testing.T) {
	sim, character := weaponWearer(t, proto.Class_ClassWarrior, 12592, 2)
	character.AutoAttacks.CancelAutoSwing(sim)
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100})
	row := spelldata.MustFind(16600)
	if row.Duration() <= 0 {
		t.Fatal("16600 states no duration")
	}
	stunnable, immune := sim.Encounter.ActiveTargetUnits[0], sim.Encounter.ActiveTargetUnits[1]
	immune.StunImmune = true

	start := sim.CurrentTime
	character.GetSpell(core.ActionID{SpellID: 16600}).Cast(sim, character.CurrentTarget)
	if !stunnable.PseudoStats.Stunned || immune.PseudoStats.Stunned {
		t.Errorf("after the cast %s is stunned: %v, the stun-immune %s: %v; want only the first",
			stunnable.Label, stunnable.PseudoStats.Stunned, immune.Label, immune.PseudoStats.Stunned)
	}
	stepTo(t, sim, start+row.Duration()+time.Millisecond)
	if stunnable.PseudoStats.Stunned {
		t.Errorf("%s is still stunned past the row's %v", stunnable.Label, row.Duration())
	}
}

// Blessing of Shahram 16599 restores the row's mana each period for the row's duration, and two
// Blessings tick side by side.
func TestBlessingOfShahramRestoresManaForEachBlessing(t *testing.T) {
	sim, character := itemWearer(t, proto.Class_ClassPaladin, proto.ItemSlot_ItemSlotMainHand, 12592)
	row := spelldata.MustFind(16599)
	mana := row.FirstAura(dbcenums.A_PERIODIC_ENERGIZE)
	ticks := int32(row.Duration() / mana.Period())
	if mana.Average() <= 0 || ticks < 2 {
		t.Fatalf("16599 restores %v mana over %d ticks, want an amount at least twice", mana.Average(), ticks)
	}
	character.SpendMana(sim, character.CurrentMana(), character.NewManaMetrics(core.ActionID{SpellID: instantSpellID}))

	blessing := character.GetSpell(core.ActionID{SpellID: 16599})
	start := sim.CurrentTime
	blessing.Cast(sim, character.CurrentTarget)
	stepTo(t, sim, start+mana.Period()/2)
	blessing.Cast(sim, character.CurrentTarget)
	stepTo(t, sim, start+mana.Period()/2+row.Duration()+time.Millisecond)

	var gain float64
	var events int32
	for _, r := range character.Metrics.ToProto().Resources {
		if r.Id.GetSpellId() == 16599 {
			gain, events = r.Gain, r.Events
		}
	}
	if events != 2*ticks || math.Abs(gain-float64(2*ticks)*mana.Average()) > 1e-6 {
		t.Errorf("two Blessings restored %v mana over %d ticks, want %d ticks of %v", gain, events, 2*ticks, mana.Average())
	}
}

// Blessing of Shahram 16599 restores mana to every member of the wearer's party with a mana bar.
func TestBlessingOfShahramRestoresManaForPartyMembers(t *testing.T) {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 12592
	newPaladin := func(name string, equipment []*proto.ItemSpec) *proto.Player {
		player := &proto.Player{
			Name: name, Class: proto.Class_ClassPaladin, Race: proto.Race_RaceHuman,
			Equipment:   &proto.EquipmentSpec{Items: equipment},
			Consumables: &proto.ConsumesSpec{},
			Buffs:       &proto.IndividualBuffs{},
			Rotation:    &proto.APLRotation{},
		}
		core.WithSpec(player, &proto.Player_HolyPaladin{HolyPaladin: &proto.HolyPaladin{}})
		return player
	}
	bare := make([]*proto.ItemSpec, len(items))
	for i := range bare {
		bare[i] = &proto.ItemSpec{}
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{newPaladin("Wearer", items), newPaladin("Friend", bare)},
			Buffs:   &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()
	wearer := sim.Raid.Parties[0].Players[0].GetCharacter()
	friend := sim.Raid.Parties[0].Players[1].GetCharacter()

	row := spelldata.MustFind(16599)
	mana := row.FirstAura(dbcenums.A_PERIODIC_ENERGIZE)
	ticks := int32(row.Duration() / mana.Period())
	for _, member := range []*core.Character{wearer, friend} {
		member.SpendMana(sim, member.CurrentMana(), member.NewManaMetrics(core.ActionID{SpellID: instantSpellID}))
	}

	start := sim.CurrentTime
	wearer.GetSpell(core.ActionID{SpellID: 16599}).Cast(sim, wearer.CurrentTarget)
	stepTo(t, sim, start+row.Duration()+time.Millisecond)

	for _, member := range []*core.Character{wearer, friend} {
		var gain float64
		var events int32
		for _, r := range member.Metrics.ToProto().Resources {
			if r.Id.GetSpellId() == 16599 {
				gain, events = r.Gain, r.Events
			}
		}
		if events != ticks || math.Abs(gain-float64(ticks)*mana.Average()) > 1e-6 {
			t.Errorf("%s gained %v mana over %d ticks, want %d ticks of %v", member.Name, gain, events, ticks, mana.Average())
		}
	}
}

// Spinal Reaper 288901's 21185 hears a killing blow and fires 21186. A wearer with a mana bar gains the
// mana 21186's effect 1 states, one with only a rage bar the rage its effect 2 states in tenths. A
// death with no kill credit gives nothing, and a second kill inside the row's cooldown gives nothing
// either.
func TestSpinalReaperRestoresPowerOnAKillingBlow(t *testing.T) {
	const reaper = 288901
	if core.GetItemByID(reaper) == nil {
		t.Skip("item 288901 is not in the live game files; this runs again once they hold it")
	}
	cooldown := spelldata.MustFind(21185).ICD()
	gain := spelldata.MustFind(21186)
	for _, tc := range []struct {
		name  string
		class proto.Class
		bear  bool
		want  float64
	}{
		{"mana", proto.Class_ClassWarlock, false, gain.EffectN(1).BaseValue()},
		{"rage", proto.Class_ClassWarrior, false, gain.EffectN(2).Tenths()},
		{"rage in bear form", proto.Class_ClassDruid, true, gain.EffectN(2).Tenths()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var sim *core.Simulation
			var character *core.Character
			if tc.bear {
				items := bearItemSlots()
				items[proto.ItemSlot_ItemSlotMainHand].Id = reaper
				var bear *druid.Druid
				sim, bear = equippedBear(t, items)
				bear.ClearForm(sim)
				bear.BearFormAura.Activate(sim)
				character = bear.GetCharacter()
			} else {
				sim, character = modWearer(t, tc.class, proto.ItemSlot_ItemSlotMainHand, reaper)
			}
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

// Ironfoe 11684 rolls 1301046's ProcChance on each landed melee hit on an Orc, and on any other target
// the column divided by its second effect's value (the rule registerExtraAttackVersus states).
func TestIronfoeIsLikelierAgainstOrcs(t *testing.T) {
	expectExtraAttackVersus(t, 11684, "Ironfoe", 1301046, proto.CreatureFamily_CreatureFamilyOrc)
}

// Blade of Hanna 2801 adds a stack of Improving Form 1325509 on each main-hand swing. Only the swing
// that brings the stacks to the count the row names casts Perfected Form 1325508: its damage once,
// its +20 to every stat for the row's duration, and the stacks gone so they build again.
func TestBladeOfHannaPerfectsItsFormAtThirtyStacks(t *testing.T) {
	if core.GetItemByID(2801) == nil {
		t.Skip("item 2801 is not in the live game files; this runs again once they hold it")
	}
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

// Claymore of Two Moons 280080 rolls 1309550 at the row's procs per minute on the weapon's speed,
// on white swings and melee specials alike. Each proc casts Waxing Moon 1319378, a spell hit whose
// damage lands after its flight, or puts up Waning Moon 1319382's Strength for the row's duration,
// and never both.
func TestClaymoreOfTwoMoonsWaxesOrWanesOnEachProc(t *testing.T) {
	if core.GetItemByID(280080) == nil {
		t.Skip("item 280080 is not in the live game files; this runs again once they hold it")
	}
	sim, character := weaponWearer(t, proto.Class_ClassWarrior, 280080, 1)
	character.AutoAttacks.CancelAutoSwing(sim)
	enemy := character.CurrentTarget

	waxRow, waneRow := spelldata.MustFind(1319378), spelldata.MustFind(1319382)
	proc := character.GetAura("Claymore of Two Moons")
	wax := character.GetSpell(core.ActionID{SpellID: waxRow.ID})
	wane := character.GetAuraByID(core.ActionID{SpellID: waneRow.ID})
	if proc == nil || proc.Dpm == nil || wax == nil || wane == nil {
		t.Fatal("280080 registers no proc on a procs per minute rate, no Waxing Moon spell or no Waning Moon aura")
	}
	want := float64(spelldata.MustFind(1309550).RPPM) * core.GetItemByID(280080).SwingSpeed / 60
	amount := waneRow.FirstAura(dbcenums.A_MOD_STAT).Min()
	if want <= 0 || want >= 1 || amount <= 0 || wane.Duration != waneRow.Duration() {
		t.Fatalf("1309550 rolls %v a hit and Waning Moon adds %v Strength for %v, want a chance, a gain and the row's %v", want, amount, wane.Duration, waneRow.Duration())
	}
	for _, mask := range []core.ProcMask{core.ProcMaskMeleeMHAuto, core.ProcMaskMeleeMHSpecial} {
		if got := proc.Dpm.Chance(mask, sim); math.Abs(got-want) > 1e-9 {
			t.Errorf("the proc rolls %v on a hit of mask %v, want the row's procs per minute on the weapon's speed, %v", got, mask, want)
		}
	}
	if wax.ProcMask != core.ProcMaskSpellDamage || wax.Flags&core.SpellFlagProc == 0 {
		t.Errorf("Waxing Moon has mask %v and proc flag %v, want a proc's spell hit", wax.ProcMask, wax.Flags&core.SpellFlagProc != 0)
	}

	// Every melee hit now procs it.
	*proc.Dpm = *character.NewFixedProcChanceManager(1, core.ProcMaskMelee)
	const hits = 100
	var waxes, wanes int
	for range hits {
		wane.Deactivate(sim)
		casts := procCasts(sim, wax)
		before := character.GetStats()
		dealHit(t, sim, character, meleeSwing(&character.Unit))
		waxed, waned := procCasts(sim, wax) > casts, wane.IsActive()
		if waxed == waned {
			t.Fatalf("a proc cast Waxing Moon %v and put up Waning Moon %v, want exactly one", waxed, waned)
		}
		if waxed {
			waxes++
			continue
		}
		wanes++
		if got := gained(character, before, stats.Strength); !statNear(got, amount) {
			t.Fatalf("Waning Moon adds %v Strength, want the row's %v", got, amount)
		}
	}
	if waxes < hits/4 || wanes < hits/4 {
		t.Errorf("%d procs cast Waxing Moon %d times and Waning Moon %d times, want about one in two each", hits, waxes, wanes)
	}

	// Away from the target, a cast deals nothing until its flight ends. The first step lands the
	// casts still in flight from the hits above.
	character.DistanceFromTarget = 20
	stepTo(t, sim, sim.CurrentTime+wax.TravelTime())
	landed := wax.SpellMetrics[enemy.UnitIndex].TotalDamage
	for range 20 {
		wax.Cast(sim, enemy)
	}
	if got := wax.SpellMetrics[enemy.UnitIndex].TotalDamage; wax.TravelTime() <= 0 || got != landed {
		t.Errorf("Waxing Moon dealt %v damage at its cast, want none before its %v flight ends", got-landed, wax.TravelTime())
	}
	stepTo(t, sim, sim.CurrentTime+wax.TravelTime())
	if got := wax.SpellMetrics[enemy.UnitIndex].TotalDamage; got <= landed {
		t.Errorf("20 Waxing Moon casts dealt %v damage once landed, want the row's damage", got-landed)
	}
}

// White Obsidian Wand 274425 rolls 1294058 on a caster's wand shots at the procs per minute of the
// item's chance-on-hit entry, on the wand's speed, and on nothing else. A landed shot puts it up, a
// missed one does not, and Shoot leaves it up. The next spell of a school the row names that pays mana
// costs the row's amount less and spends it; a free cast keeps it.
func TestWhiteObsidianWandProcsOnLandedShotsAndCheapensTheNextSpell(t *testing.T) {
	sim, wearer := priestWearing(t, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotRanged: 274425})
	character := &wearer.Character
	enemy := sim.Encounter.ActiveTargetUnits[0]
	proc := character.GetAura("White Obsidian Wand - Trigger")
	buff := character.GetAura("White Obsidian Wand")
	shoot := character.GetSpell(core.ActionID{SpellID: 5019})
	shot := character.GetSpell(core.ActionID{SpellID: 5019, Tag: 1})
	if character.AutoAttacks.AutoSwingMelee || character.AutoAttacks.AutoSwingRanged || proc == nil || proc.Dpm == nil || buff == nil || shoot == nil || shot == nil {
		t.Fatal("the priest swings an auto attack, or has no Shoot, or 274425 registers no buff or no proc on a procs per minute rate")
	}
	expectWeaponRate(t, sim, proc, core.ProcMaskRangedAuto, 274425, 1294058)
	if got := proc.Dpm.Chance(core.ProcMaskSpellDamage, sim); got != 0 {
		t.Errorf("a spell hit rolls the proc at %v, want never", got)
	}
	row := spelldata.MustFind(1294058)
	effect := row.FirstAura(dbcenums.A_MOD_POWER_COST_SCHOOL)
	smite := character.GetSpell(core.ActionID{SpellID: priest.SmiteRankMap.Highest().ID})
	if effect == nil || effect.Average() >= 0 || smite == nil || smite.Cost == nil || !smite.SpellSchool.Matches(core.SpellSchool(effect.Misc)) {
		t.Fatal("1294058 states no cost reduction, or the priest has no Smite with a cost in a school it names")
	}

	// Every wand shot procs it.
	*proc.Dpm = *character.NewFixedProcChanceManager(1, core.ProcMaskRangedAuto)
	character.AddStatsDynamic(sim, stats.Stats{stats.PhysicalHitPercent: 100})

	character.OnSpellHitDealt(sim, shot, &core.SpellResult{Target: enemy, Outcome: core.OutcomeMiss})
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
	if buff.IsActive() {
		t.Fatal("a missed shot put the buff up, want only a landed one")
	}

	if !shoot.Cast(sim, enemy) {
		t.Fatal("Shoot could not be cast")
	}
	stepTo(t, sim, sim.CurrentTime+3*time.Second)
	if got := shot.SpellMetrics[enemy.UnitIndex]; got.Hits+got.Crits != 1 || !buff.IsActive() {
		t.Fatalf("Shoot landed %d shots and the buff is up %v, want one landed shot and the buff up", got.Hits+got.Crits, buff.IsActive())
	}
	if buff.Duration != row.Duration() {
		t.Errorf("the buff lasts %v, want the row's %v", buff.Duration, row.Duration())
	}
	stepTo(t, sim, max(sim.CurrentTime, character.GCD.ReadyAt()))
	if !shoot.Cast(sim, enemy) || !buff.IsActive() {
		t.Fatal("Shoot could not be cast under the buff, or spent it, want the buff left up")
	}
	stepTo(t, sim, sim.CurrentTime+3*time.Second)

	manaSpent := smite.Cost.ResourceCostImpl.(*core.ManaCost).ResourceMetrics
	castSmite := func() float64 {
		t.Helper()
		stepTo(t, sim, max(sim.CurrentTime, character.GCD.ReadyAt()))
		before := manaSpent.Gain
		if !smite.Cast(sim, enemy) {
			t.Fatal("Smite could not be cast")
		}
		stepTo(t, sim, character.Hardcast.Expires+core.SpellBatchWindow)
		return before - manaSpent.Gain
	}
	free := character.AddDynamicMod(core.SpellModConfig{Kind: core.SpellMod_PowerCost_Pct, School: smite.SpellSchool, FloatValue: -1})
	free.Activate()
	if paid := castSmite(); paid != 0 || !buff.IsActive() {
		t.Fatalf("a free Smite paid %v and left the buff up %v, want it to pay nothing and leave the buff up", paid, buff.IsActive())
	}
	free.Deactivate()
	cheaper := castSmite()
	if buff.IsActive() {
		t.Error("the Smite after the proc left the buff up, want it spent")
	}
	if full := castSmite(); cheaper <= 0 || full-cheaper != -effect.Average() {
		t.Errorf("Smite cost %v under the buff and %v after it, want the row's %v less under it", cheaper, full, -effect.Average())
	}
}

// Iceblade Hacker 13952 carries both its procs: every landed main-hand swing casts 1298414 on the
// target, and 1298416 rolls the procs per minute of the item's chance-on-hit entry on the axe's speed
// and casts its Frostbolt. A missed swing casts neither. With Always Hit on 1298414's row, the swing
// damage never misses, is never dodged, parried or blocked, and can still crit. While the entry
// states no rate, the swing's damage stays and the Frostbolt registers nothing.
func TestIcebladeHackerCarriesBothProcs(t *testing.T) {
	editSpellRow(t, 1298414, func(s *spelldata.Spell) { s.Attr[dbcenums.ATTR_INDEX_EX_3] |= dbcenums.ATTR_EX_3_ALWAYS_HIT })
	sim, character := itemWearerWithEquipment(t, proto.Class_ClassWarrior, &proto.EquipmentSpec{Items: weaponSlots(13952, 0)})
	character.AutoAttacks.CancelAutoSwing(sim)
	proc := character.GetAura("Iceblade Hacker - Frostbolt")
	chill := character.GetSpell(core.ActionID{SpellID: 1298414})
	frostbolt := character.GetSpell(core.ActionID{SpellID: 1298416})
	if character.GetAura("Iceblade Hacker") == nil || proc == nil || proc.Dpm == nil || chill == nil || frostbolt == nil {
		t.Fatal("13952 registers no swing damage, no Frostbolt or no proc on a procs per minute rate")
	}
	expectWeaponRate(t, sim, proc, core.ProcMaskMeleeMHAuto, 13952, 1298416)

	// Every main-hand hit now rolls a Frostbolt.
	*proc.Dpm = *character.NewFixedProcChanceManager(1, core.ProcMaskMeleeMH)
	enemy := character.CurrentTarget
	casts := func(spell *core.Spell) int32 { return spell.SpellMetrics[enemy.UnitIndex].Casts }
	character.OnSpellHitDealt(sim, meleeSwing(&character.Unit), &core.SpellResult{Target: enemy, Outcome: core.OutcomeMiss})
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
	if casts(chill) != 0 || casts(frostbolt) != 0 {
		t.Errorf("a missed swing cast the swing damage %d and the Frostbolt %d times, want neither", casts(chill), casts(frostbolt))
	}
	dealHit(t, sim, character, meleeSwing(&character.Unit))
	if casts(chill) != 1 || casts(frostbolt) != 1 {
		t.Errorf("a landed swing cast the swing damage %d and the Frostbolt %d times, want each once", casts(chill), casts(frostbolt))
	}
	for range 500 {
		chill.Cast(sim, enemy)
	}
	stepTo(t, sim, sim.CurrentTime+time.Second)
	if m := chill.SpellMetrics[enemy.UnitIndex]; m.Misses+m.Dodges+m.Parries+m.Blocks != 0 || m.Crits == 0 {
		t.Errorf("the swing damage missed %d, was dodged %d, parried %d, blocked %d and crit %d times; want it always landing and able to crit",
			m.Misses, m.Dodges, m.Parries, m.Blocks, m.Crits)
	}

	withEntryPPM(t, 13952, 1298416, 0)
	_, unrated := itemWearerWithEquipment(t, proto.Class_ClassWarrior, &proto.EquipmentSpec{Items: weaponSlots(13952, 0)})
	if unrated.GetAura("Iceblade Hacker") == nil || unrated.GetAura("Iceblade Hacker - Frostbolt") != nil {
		t.Error("with no rate on the entry, want the swing damage registered and no Frostbolt")
	}
}

// Mana Channeling Wand 18483 rolls 1302348 on a caster's landed wand shots at the procs per minute of
// the item's chance-on-hit entry, on the wand's speed, and on nothing else. Each proc restores the
// mana of 1302350, the spell 1302348 fires; a missed shot restores none.
func TestManaChannelingWandRestoresManaOnLandedShots(t *testing.T) {
	sim, wearer := priestWearing(t, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotRanged: 18483})
	character := &wearer.Character
	enemy := sim.Encounter.ActiveTargetUnits[0]
	proc := character.GetAura("Mana Channeling Wand")
	shot := character.GetSpell(core.ActionID{SpellID: 5019, Tag: 1})
	if character.AutoAttacks.AutoSwingMelee || character.AutoAttacks.AutoSwingRanged || proc == nil || proc.Dpm == nil || shot == nil {
		t.Fatal("the priest swings an auto attack, or has no Shoot, or 18483 registers no proc on a procs per minute rate")
	}
	expectWeaponRate(t, sim, proc, core.ProcMaskRangedAuto, 18483, 1302350)
	if got := proc.Dpm.Chance(core.ProcMaskSpellDamage, sim); got != 0 {
		t.Errorf("a spell hit rolls the proc at %v, want never", got)
	}

	gain := spelldata.MustFind(1302350).EnergizeEffect().Average()
	restored := func() float64 {
		for _, r := range character.Metrics.ToProto().Resources {
			if r.Id.GetItemId() == 18483 {
				return r.Gain
			}
		}
		return 0
	}
	character.SpendMana(sim, 4*gain, character.NewManaMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionManaRegen}))

	// Every wand shot now rolls a proc.
	*proc.Dpm = *character.NewFixedProcChanceManager(1, core.ProcMaskRangedAuto)
	character.OnSpellHitDealt(sim, shot, &core.SpellResult{Target: enemy, Outcome: core.OutcomeMiss})
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
	if got := restored(); got != 0 {
		t.Errorf("a missed shot restored %v mana, want none", got)
	}
	character.OnSpellHitDealt(sim, shot, &core.SpellResult{Target: enemy, Outcome: core.OutcomeHit, Damage: 100})
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
	if got := restored(); gain <= 0 || got != gain {
		t.Errorf("a landed shot restored %v mana, want the row's %v", got, gain)
	}
}

// A warrior holding the item in the main hand against one target with no armor, which spends the
// execute35 and execute20 shares of the encounter's end below 35% and 20% health. Its auto attacks
// are off, every hit lands and none crits, and the item's proc fires on every main-hand hit.
func procWeaponWearer(t *testing.T, itemID int32, procName string, target *proto.Target, execute35, execute20 float64) (*core.Simulation, *core.Character) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = itemID
	player := &proto.Player{
		Name: "Wearer", Class: proto.Class_ClassWarrior, Race: proto.Race_RaceTauren,
		Equipment:   &proto.EquipmentSpec{Items: items},
		Consumables: &proto.ConsumesSpec{},
		Buffs:       &proto.IndividualBuffs{},
		Rotation:    &proto.APLRotation{},
	}
	core.WithSpec(player, &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{}})
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}},
		Encounter:  &proto.Encounter{Duration: 180, ExecuteProportion_35: execute35, ExecuteProportion_20: execute20, Targets: []*proto.Target{target}},
	}, simsignals.CreateSignals())
	sim.Reset()
	character := sim.Raid.Parties[0].Players[0].GetCharacter()
	character.AutoAttacks.CancelAutoSwing(sim)
	stepTo(t, sim, core.SpellBatchWindow)
	character.AddStatsDynamic(sim, stats.Stats{
		stats.PhysicalHitPercent: 100, stats.SpellHitPercent: 100,
		stats.PhysicalCritPercent: -character.GetStat(stats.PhysicalCritPercent) - 100,
		stats.SpellCritPercent:    -character.GetStat(stats.SpellCritPercent) - 100,
	})
	proc := character.GetAura(procName)
	if proc == nil || proc.Dpm == nil {
		t.Fatalf("%d registers no proc named %q on a procs per minute rate", itemID, procName)
	}
	*proc.Dpm = *character.NewFixedProcChanceManager(1, core.ProcMaskMeleeMH)
	return sim, character
}

// Barman Shanker 12791's Rend 1300787, applied from behind the target, ticks for the multiple of its
// periodic damage its second effect states.
func TestBarmanShankerRendDealsItsMultipleFromBehind(t *testing.T) {
	sim, character := procWeaponWearer(t, 12791, "Barman Shanker", &proto.Target{Name: "target", Level: 60}, 0, 0)
	if character.PseudoStats.InFrontOfTarget {
		t.Fatal("the wearer stands in front of the target")
	}
	enemy := character.CurrentTarget
	rend := character.GetSpell(core.ActionID{SpellID: 1300787})
	if rend == nil {
		t.Fatal("12791 registers no Rend")
	}
	row := spelldata.MustFind(1300787)
	multiple := row.EffectN(2).Average()
	want := row.EffectN(1).Average() * float64(rowTickCount(row, row.EffectN(1))) * multiple
	rend.Cast(sim, enemy)
	stepTo(t, sim, sim.CurrentTime+row.Duration()+time.Second)
	if got := rend.SpellMetrics[enemy.UnitIndex].TotalDamage; multiple <= 1 || math.Abs(got-want) > 0.5 {
		t.Errorf("a Rend from behind dealt %v, want the row's ticks times %v, %v", got, multiple, want)
	}
}

// Sul'thraze the Lasher 9372: a landed Rage of Sul'thraze 1294430 puts Shadow Lash 1294431 on the
// target, which deals its damage when its duration runs out. A Rage that lands while the Lash runs puts
// up no second one and leaves its time alone; a missed Rage puts up none.
func TestSulthrazeTheLasherLashesOnceAfterItsDuration(t *testing.T) {
	sim, character := procWeaponWearer(t, 9372, "Sul'thraze the Lasher", &proto.Target{Name: "target", Level: 60}, 0, 0)
	enemy := character.CurrentTarget
	rage := character.GetSpell(core.ActionID{SpellID: 1294430})
	lash := character.GetSpell(core.ActionID{SpellID: 1294431})
	if rage == nil || lash == nil {
		t.Fatal("9372 registers no Rage of Sul'thraze or no Shadow Lash")
	}
	casts := func() int32 { return lash.SpellMetrics[enemy.UnitIndex].Casts }

	character.OnSpellHitDealt(sim, rage, &core.SpellResult{Target: enemy, Outcome: core.OutcomeMiss})
	if casts() != 0 {
		t.Fatalf("a missed Rage put up %d Lashes, want none", casts())
	}
	dealHit(t, sim, character, meleeSwing(&character.Unit))
	stepTo(t, sim, sim.CurrentTime+time.Second)
	dot := lash.Dot(enemy)
	if casts() != 1 || !dot.IsActive() {
		t.Fatalf("a landed Rage put up %d Lashes, the Lash running %v, want one running", casts(), dot.IsActive())
	}
	expires := dot.ExpiresAt()
	if want := spelldata.MustFind(1294431).Duration(); dot.ExpectedTickCount() != 1 || dot.TickPeriod() != want {
		t.Errorf("the Lash ticks %d times every %v, want once after the row's %v", dot.ExpectedTickCount(), dot.TickPeriod(), want)
	}

	dealHit(t, sim, character, meleeSwing(&character.Unit))
	stepTo(t, sim, sim.CurrentTime+time.Second)
	if casts() != 1 || dot.ExpiresAt() != expires {
		t.Errorf("a Rage landing on a running Lash cast %d Lashes and moved its end from %v to %v, want one Lash and the end kept", casts(), expires, dot.ExpiresAt())
	}
	stepTo(t, sim, expires+core.SpellBatchWindow)
	if dot.IsActive() || lash.SpellMetrics[enemy.UnitIndex].TotalDamage <= 0 {
		t.Errorf("at the Lash's end it is running %v and has dealt %v, want it done and its damage dealt", dot.IsActive(), lash.SpellMetrics[enemy.UnitIndex].TotalDamage)
	}
}

// Pendulum of Doom 9425's Fatal Wound 10373, Diabolic Skiver 9475's Fatal Wound 18206, Blackhand Doomsaw
// 12583's Fatal Wound 16549 and Gut Ripper 2164's Wound 18107 deal the percent their row states more to a target below the health percent
// their row states, from that phase's start, and their plain damage to a target above it.
func TestFatalWoundDealsMoreBelowTheExecuteHealth(t *testing.T) {
	for _, item := range []struct {
		itemID, spellID int32
		procName        string
	}{
		{9425, 10373, "Pendulum of Doom - Fatal Wound"},
		{9475, 18206, "Diabolic Skiver"},
		{2164, 18107, "Gut Ripper"},
		{12583, 16549, "Blackhand Doomsaw"},
	} {
		t.Run(item.procName, func(t *testing.T) {
			row := spelldata.MustFind(item.spellID)
			health := int32(row.EffectN(3).Average())
			if health != 25 {
				t.Fatalf("the row's threshold is %d%%, outside the case the test covers", health)
			}
			meanHit := func(execute35, execute20 float64, at time.Duration, below bool) float64 {
				sim, character := procWeaponWearer(t, item.itemID, item.procName, &proto.Target{Name: "target", Level: 60}, execute35, execute20)
				enemy := character.CurrentTarget
				stepTo(t, sim, at)
				if got := enemy.IsBelowExecuteHealth(health); got != below {
					t.Fatalf("at %v with %v and %v of the fight below 35%% and 20%%, the target below %d%% reads %v, want %v", at, execute35, execute20, health, got, below)
				}
				wound := character.GetSpell(core.ActionID{SpellID: item.spellID})
				if wound == nil {
					t.Fatalf("%d registers no Fatal Wound", item.itemID)
				}
				// A Fatal Wound is a main-hand special that rolls the proc again, so the proc stays off here.
				*character.GetAura(item.procName).Dpm = *character.NewFixedProcChanceManager(0, core.ProcMaskMeleeMH)
				var total float64
				var landed int
				for range 400 {
					if dealt := castDamage(sim, wound, enemy); dealt > 0 {
						total += dealt
						landed++
					}
				}
				if landed < 300 {
					t.Fatalf("%d of 400 Fatal Wounds dealt damage, want most", landed)
				}
				return total / float64(landed)
			}
			want := 1 + row.EffectN(2).Average()/100
			plain := meanHit(0, 0, time.Second, false)
			// Below 35% from the start and 20% only at the end: the target crosses 25% two thirds of the way.
			justBelow := meanHit(1, 0, 121*time.Second, true)
			if ratio := justBelow / plain; want <= 1 || math.Abs(ratio-want) > 0.05*want {
				t.Errorf("Fatal Wound averages %v just below %d%% and %v above it, a ratio of %v, want the row's %v", justBelow, health, plain, ratio, want)
			}
			aboveRow := meanHit(1, 0, 119*time.Second, false)
			if ratio := aboveRow / plain; math.Abs(ratio-1) > 0.05 {
				t.Errorf("Fatal Wound averages %v above %d%% but below 35%%, a ratio of %v to the plain hit, want 1", aboveRow, health, ratio)
			}
		})
	}
}

// Flame Wrath 11809's proc puts 16560 on the wearer for the row's duration, raising its Fire
// resistance by the row's amount, and a melee hit taken then casts the shield's damage at the
// attacker. A later proc refreshes it.
func TestFlameWrathPutsUpItsFireShield(t *testing.T) {
	sim, character := procWeaponWearer(t, 11809, "Flame Wrath", &proto.Target{Name: "target", Level: 60}, 0, 0)
	enemy := character.CurrentTarget
	row := spelldata.MustFind(16560)
	shieldAura := character.GetAura("Flame Wrath Fire Shield")
	resistance := row.FirstAura(dbcenums.A_MOD_RESISTANCE)
	if shieldAura == nil || resistance == nil {
		t.Fatal("11809 registers no fire shield, or 16560 states no resistance")
	}
	shield, _ := damageShield(t, character, 16560)

	before := character.GetStat(stats.FireResistance)
	dealHit(t, sim, character, meleeSwing(&character.Unit))
	if !shieldAura.IsActive() || shieldAura.Duration != row.Duration() {
		t.Fatalf("a proc left the shield up %v for %v, want it up for the row's %v", shieldAura.IsActive(), shieldAura.Duration, row.Duration())
	}
	if got := character.GetStat(stats.FireResistance) - before; got != resistance.Average() {
		t.Errorf("the shield raised Fire resistance by %v, want the row's %v", got, resistance.Average())
	}
	takeHit(t, sim, character, meleeSwing(enemy))
	if got := shield.SpellMetrics[enemy.UnitIndex].Casts; got != 1 {
		t.Errorf("a melee hit taken cast the shield's damage %d times, want once", got)
	}

	expires := shieldAura.ExpiresAt()
	stepTo(t, sim, sim.CurrentTime+5*time.Second)
	dealHit(t, sim, character, meleeSwing(&character.Unit))
	if shieldAura.ExpiresAt() <= expires {
		t.Errorf("a second proc left the shield's end at %v, want it later than %v", shieldAura.ExpiresAt(), expires)
	}
}

// The damage one cast of the spell deals to the target.
func castDamage(sim *core.Simulation, spell *core.Spell, target *core.Unit) float64 {
	before := spell.SpellMetrics[target.UnitIndex].TotalDamage
	spell.Cast(sim, target)
	return spell.SpellMetrics[target.UnitIndex].TotalDamage - before
}

// The factor each item's row states against its tooltip's creatures, read through the row's versus
// bonus: Gahz'rilla Fang 9467's 3742 against Aquatic enemies, The Chief's Enforcer 9477's 1294339
// against Trolls, Angerforge's Battle Axe 11816's 1300783 against Orcs, and Sul'thraze the Lasher
// 9372's Shadow Lash 1294431 against Trolls, which its Rage of Sul'thraze 1294430 states and leaves
// its own hit without.
func TestProcSpellsDealTheirVersusFactor(t *testing.T) {
	for _, tc := range []struct {
		itemID         int32
		procName       string
		rowID, spellID int32
		family         proto.CreatureFamily
	}{
		{9467, "Gahz'rilla Fang - Gahz'rilla Fang", 3742, 3742, proto.CreatureFamily_CreatureFamilyAquatic},
		{9477, "The Chief's Enforcer", 1294339, 1294339, proto.CreatureFamily_CreatureFamilyTroll},
		{11816, "Angerforge's Battle Axe", 1300783, 1300783, proto.CreatureFamily_CreatureFamilyOrc},
		{9372, "Sul'thraze the Lasher", 1294430, 1294431, proto.CreatureFamily_CreatureFamilyTroll},
	} {
		bonus, ok := spelldata.MustFind(tc.rowID).VersusBonus()
		if !ok || bonus.Multiplier <= 1 || !slices.Contains(bonus.Families, tc.family) {
			t.Fatalf("%d states no factor above 1 against %v", tc.rowID, tc.family)
		}
		dealt := func(family proto.CreatureFamily) float64 {
			sim, character := procWeaponWearer(t, tc.itemID, tc.procName, &proto.Target{Name: "target", Level: 60, CreatureFamily: family}, 0, 0)
			*character.GetAura(tc.procName).Dpm = *character.NewFixedProcChanceManager(0, core.ProcMaskMeleeMH)
			spell := character.GetSpell(core.ActionID{SpellID: tc.spellID})
			if spell == nil {
				t.Fatalf("%d registers no %d", tc.itemID, tc.spellID)
			}
			if spell.Dot(character.CurrentTarget) == nil {
				// The first cast that is not dodged or parried.
				for range 50 {
					if damage := castDamage(sim, spell, character.CurrentTarget); damage > 0 {
						return damage
					}
				}
				t.Fatalf("50 casts of %d dealt no damage", tc.spellID)
			}
			spell.Cast(sim, character.CurrentTarget)
			stepTo(t, sim, spell.Dot(character.CurrentTarget).ExpiresAt()+core.SpellBatchWindow)
			return spell.SpellMetrics[character.CurrentTarget.UnitIndex].TotalDamage
		}
		plain, against := dealt(proto.CreatureFamily_CreatureFamilyUnknown), dealt(tc.family)
		if plain <= 0 || math.Abs(against/plain-bonus.Multiplier) > 1e-6 {
			t.Errorf("%d's %d deals %v against %v and %v against others, want the row's %v times", tc.itemID, tc.spellID, against, tc.family, plain, bonus.Multiplier)
		}
	}

	troll := &proto.Target{Name: "target", Level: 60, CreatureFamily: proto.CreatureFamily_CreatureFamilyTroll}
	_, character := procWeaponWearer(t, 9372, "Sul'thraze the Lasher", troll, 0, 0)
	if rage := character.GetSpell(core.ActionID{SpellID: 1294430}); rage == nil || rage.VersusBonus != nil {
		t.Error("Rage of Sul'thraze carries a versus bonus of its own, want the Lash alone to carry it")
	}

	// The Chief's Enforcer's stun lasts the strike's factor longer on a Troll.
	stun := spelldata.MustFind(1294340)
	factor, _ := spelldata.MustFind(1294339).VersusBonus()
	for _, family := range []proto.CreatureFamily{proto.CreatureFamily_CreatureFamilyUnknown, proto.CreatureFamily_CreatureFamilyTroll} {
		_, character := procWeaponWearer(t, 9477, "The Chief's Enforcer", &proto.Target{Name: "target", Level: 60, CreatureFamily: family}, 0, 0)
		want := stun.Duration()
		if family == proto.CreatureFamily_CreatureFamilyTroll {
			want = time.Duration(float64(want) * factor.Multiplier)
		}
		if aura := character.CurrentTarget.GetAura(fmt.Sprintf("%s %d", stun.Name, stun.ID)); aura == nil || aura.Duration != want {
			t.Errorf("against %v the stun is registered %v, want it lasting %v", family, aura, want)
		}
	}
}

// Gahz'rilla Fang 9467's 3742 deals the percent Static Electricity 1294477 states more for each
// stack on the wearer, and spends all of them but one.
func TestGahzrillaFangGrowsWithStaticElectricity(t *testing.T) {
	sim, character := procWeaponWearer(t, 9467, "Gahz'rilla Fang - Gahz'rilla Fang", &proto.Target{Name: "target", Level: 60}, 0, 0)
	enemy := character.CurrentTarget
	fang := character.GetSpell(core.ActionID{SpellID: 3742})
	static := character.GetAuraByID(core.ActionID{SpellID: 1294477})
	if fang == nil || static == nil {
		t.Fatal("9467 registers no damage spell or no Static Electricity")
	}
	plain := castDamage(sim, fang, enemy)
	const stacks = 10
	static.Activate(sim)
	static.SetStacks(sim, stacks)
	charged := castDamage(sim, fang, enemy)
	want := 1 + stacks*spelldata.MustFind(1294477).EffectN(1).Average()/100
	if plain <= 0 || want <= 1 || math.Abs(charged/plain-want) > 1e-6 {
		t.Errorf("with %d stacks it dealt %v against %v with none, a ratio of %v, want %v", stacks, charged, plain, charged/plain, want)
	}
	if got := static.GetStacks(); got != 1 {
		t.Errorf("the proc left %d stacks, want 1", got)
	}
	castDamage(sim, fang, enemy)
	if got := static.GetStacks(); got != 1 {
		t.Errorf("a proc on 1 stack left %d, want 1", got)
	}
}

// The Chief's Enforcer 9477's proc casts 1294340 at the target with each strike, which stuns it for
// the row's duration even when the strike misses. A stun-immune target takes no stun.
func TestTheChiefsEnforcerStunsWhetherOrNotTheStrikeLands(t *testing.T) {
	row := spelldata.MustFind(1294340)
	label := fmt.Sprintf("%s %d", row.Name, row.ID)
	for _, tc := range []struct {
		level   int32
		stunned bool
	}{
		{60, true},
		{63, false},
	} {
		sim, character := procWeaponWearer(t, 9477, "The Chief's Enforcer", &proto.Target{Name: "target", Level: tc.level}, 0, 0)
		enemy := character.CurrentTarget
		strike := character.GetSpell(core.ActionID{SpellID: 1294339})
		stunSpell := character.GetSpell(core.ActionID{SpellID: 1294340})
		if strike == nil || stunSpell == nil {
			t.Fatal("9477 registers no strike or no stun")
		}
		// The strike misses and is dodged at the plain rates again.
		character.AddStatsDynamic(sim, stats.Stats{stats.PhysicalHitPercent: -character.GetStat(stats.PhysicalHitPercent)})
		for range 100 {
			dealHit(t, sim, character, meleeSwing(&character.Unit))
		}
		metrics := strike.SpellMetrics[enemy.UnitIndex]
		if metrics.Hits+metrics.Crits >= metrics.Casts || stunSpell.SpellMetrics[enemy.UnitIndex].Casts != metrics.Casts {
			t.Fatalf("against level %d the proc cast %d strikes, landed %d and cast %d stuns, want some strikes missed and a stun with each",
				tc.level, metrics.Casts, metrics.Hits+metrics.Crits, stunSpell.SpellMetrics[enemy.UnitIndex].Casts)
		}
		stun := enemy.GetAura(label)
		if got := stun != nil && stun.IsActive(); got != tc.stunned {
			t.Errorf("against level %d the target is stunned %v, want %v", tc.level, got, tc.stunned)
		}
		if tc.stunned && (stun.Duration != row.Duration() || !enemy.PseudoStats.Stunned) {
			t.Errorf("the stun lasts %v and marks the target stunned %v, want the row's %v and stunned", stun.Duration, enemy.PseudoStats.Stunned, row.Duration())
		}
	}
}

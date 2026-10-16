//go:build with_db

package forever_test

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func trinketWarrior(t *testing.T, trinketID int32) (*core.Simulation, *core.Character) {
	t.Helper()
	return trinketWarriorWielding(t, trinketID, 0, 0)
}

func trinketWarriorWielding(t *testing.T, trinketID int32, mainHand int32, neck int32) (*core.Simulation, *core.Character) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotTrinket1].Id = trinketID
	items[proto.ItemSlot_ItemSlotMainHand].Id = mainHand
	items[proto.ItemSlot_ItemSlotNeck].Id = neck
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Wearer", Class: proto.Class_ClassWarrior, Race: proto.Race_RaceOrc,
				Spec: &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{Options: &proto.DpsWarrior_Options{
					ClassOptions: &proto.WarriorOptions{DefaultStance: proto.WarriorStance_WarriorStanceBerserker}}}},
				Equipment:   &proto.EquipmentSpec{Items: items},
				Consumables: &proto.ConsumesSpec{},
				Buffs:       &proto.IndividualBuffs{},
				Rotation:    &proto.APLRotation{},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

func stepTo(t *testing.T, sim *core.Simulation, until time.Duration) {
	t.Helper()
	reached := false
	sim.AddPendingAction(core.NewDelayedAction(core.DelayedActionOptions{DoAt: until, OnAction: func(*core.Simulation) { reached = true }}))
	for steps := 0; !reached; steps++ {
		if steps > 10000 {
			t.Fatalf("the sim did not reach %v", until)
		}
		sim.Step()
	}
}

// Uses the item and returns its spell, after checking the major cooldown it is filed under and the
// cooldowns its on-use effect states.
func useTrinket(t *testing.T, sim *core.Simulation, character *core.Character, itemID int32, cdType core.CooldownType,
	cooldown time.Duration, category int32, categoryCooldown time.Duration) *core.Spell {
	t.Helper()
	spell := character.GetSpell(core.ActionID{ItemID: itemID})
	if spell == nil {
		t.Fatalf("%d registers no on-use spell", itemID)
	}
	mcd := character.GetMajorCooldown(spell.ActionID)
	if mcd == nil || mcd.Type != cdType {
		t.Errorf("%d's major cooldown is %v, want type %v", itemID, mcd, cdType)
	}
	if spell.CD.Duration != cooldown {
		t.Errorf("%d's cooldown is %v, want %v", itemID, spell.CD.Duration, cooldown)
	}
	if category == 0 && spell.SharedCD.Timer != nil {
		t.Errorf("%d shares a cooldown, want none", itemID)
	}
	if category != 0 && (spell.SharedCD.Timer != character.GetOrInitSpellCategoryTimer(category) || spell.SharedCD.Duration != categoryCooldown) {
		t.Errorf("%d shares %v on another timer than category %d's, want %v on it", itemID, spell.SharedCD.Duration, category, categoryCooldown)
	}
	if !spell.Cast(sim, character.CurrentTarget) {
		t.Fatalf("%d could not be used", itemID)
	}
	return spell
}

// Uses the item, checking it against the cooldown, category and category cooldown its db entry's
// on-use effect states; a category with no cooldown of its own runs for the effect's duration.
func useTrinketAsEntryStates(t *testing.T, sim *core.Simulation, character *core.Character, itemID int32, cdType core.CooldownType) *core.Spell {
	t.Helper()
	for _, effect := range core.GetItemByID(itemID).ItemEffects {
		onUse := effect.GetOnUse()
		if onUse == nil {
			continue
		}
		if onUse.CategoryId <= 0 {
			return useTrinket(t, sim, character, itemID, cdType, core.DurationFromMillis(onUse.CooldownMs), 0, 0)
		}
		categoryCooldown := onUse.CategoryCooldownMs
		if categoryCooldown <= 0 {
			categoryCooldown = effect.EffectDurationMs
		}
		return useTrinket(t, sim, character, itemID, cdType, core.DurationFromMillis(onUse.CooldownMs),
			onUse.CategoryId, core.DurationFromMillis(categoryCooldown))
	}
	t.Fatalf("%d's entry states no on-use effect", itemID)
	return nil
}

func gained(character *core.Character, before stats.Stats, stat stats.Stat) float64 {
	return character.GetStat(stat) - before[stat]
}

func statNear(a, b float64) bool {
	return math.Abs(a-b) < 1e-6
}

func meleeSwing(unit *core.Unit) *core.Spell {
	return &core.Spell{ProcMask: core.ProcMaskMeleeMHAuto, SpellSchool: core.SpellSchoolPhysical, Unit: unit}
}

func fireSpell(unit *core.Unit) *core.Spell {
	return &core.Spell{ProcMask: core.ProcMaskSpellDamage, SpellSchool: core.SpellSchoolFire, DefenseType: core.DefenseTypeMagic, Unit: unit}
}

// A landed hit taken, and the spell batch window the listener resolves after.
func takeHit(t *testing.T, sim *core.Simulation, character *core.Character, spell *core.Spell) {
	t.Helper()
	character.OnSpellHitTaken(sim, spell, &core.SpellResult{Target: &character.Unit, Outcome: core.OutcomeHit, Damage: 100})
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
}

func dealHit(t *testing.T, sim *core.Simulation, character *core.Character, spell *core.Spell) {
	t.Helper()
	character.OnSpellHitDealt(sim, spell, &core.SpellResult{Target: character.CurrentTarget, Outcome: core.OutcomeHit, Damage: 100})
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
}

// The countdown buff a Zandalarian trinket or the Petrified Scarab puts up: on use it carries all its
// stacks, and each stat it grants is its stack count times one stack's share.
func countdownBuff(t *testing.T, character *core.Character, spellID int32) (*core.Aura, int32) {
	t.Helper()
	buff := character.GetAuraByID(core.ActionID{SpellID: spellID})
	if buff == nil || !buff.IsActive() {
		t.Fatalf("using the item put up no %d", spellID)
	}
	stacks := buff.MaxStacks
	if stacks < 2 || buff.GetStacks() != stacks {
		t.Fatalf("%d went up with %d of %d stacks, want all of at least 2", spellID, buff.GetStacks(), stacks)
	}
	if got, want := buff.Duration, spelldata.MustFind(spellID).Duration(); got != want {
		t.Errorf("%d lasts %v, want the row's %v", spellID, got, want)
	}
	return buff, stacks
}

// Checks each stat is what is left of its full amount with the given stacks of the buff's gone.
func expectCountdown(t *testing.T, when string, character *core.Character, before stats.Stats, full stats.Stats, stacks, lost int32) {
	t.Helper()
	for stat, amount := range full {
		if amount == 0 {
			continue
		}
		want := amount * float64(stacks-lost) / float64(stacks)
		if got := gained(character, before, stats.Stat(stat)); !statNear(got, want) {
			t.Errorf("%s the buff grants %v %s, want %v of the %v on use", when, got, stats.Stat(stat).StatName(), want, amount)
		}
	}
}

// What the buff grants on use over the stats given.
func fullAmounts(character *core.Character, before stats.Stats, of ...stats.Stat) stats.Stats {
	var full stats.Stats
	for _, stat := range of {
		full[stat] = gained(character, before, stat)
	}
	return full
}

// Lifegiving Gem 19341 raises maximum health by a share of it for the buff's duration and heals for
// the same amount, on the cooldown its entry states. A larger maximum gets the same share. When the
// raise fades, health below the lower maximum stays, health above it drops to it, and neither counts
// as damage taken.
func TestLifegivingGemRaisesMaximumHealthByAShare(t *testing.T) {
	sim, character := trinketWarrior(t, 19341)
	maxHealth := character.MaxHealth()
	character.RemoveHealth(sim, maxHealth/2)
	current := character.CurrentHealth()
	damageTaken := character.DamageTakenHealthMetrics.ActualGain

	useTrinketAsEntryStates(t, sim, character, 19341, core.CooldownTypeSurvival)
	gift := character.GetAura("Gift of Life")
	bonus := character.MaxHealth() - maxHealth
	if gift == nil || bonus <= 0 || bonus >= maxHealth {
		t.Fatalf("using the gem raised maximum health by %v of %v, want a share of it", bonus, maxHealth)
	}
	if got := character.CurrentHealth(); !statNear(got, current+bonus) {
		t.Errorf("health is %v, want %v healed by %v", got, current, bonus)
	}

	stepTo(t, sim, sim.CurrentTime+gift.Duration+time.Millisecond)
	if got := character.MaxHealth(); !statNear(got, maxHealth) {
		t.Errorf("after the buff's %v maximum health is %v, want %v back", gift.Duration, got, maxHealth)
	}
	if got := character.CurrentHealth(); !statNear(got, current+bonus) {
		t.Errorf("after the buff below the maximum health is %v, want %v kept", got, current+bonus)
	}

	character.AddStatsDynamic(sim, stats.Stats{stats.Health: maxHealth})
	larger := character.MaxHealth()
	character.GainHealth(sim, larger, character.NewHealthMetrics(core.ActionID{OtherID: proto.OtherAction_OtherActionHealingModel}))
	gift.Activate(sim)
	largerBonus := character.MaxHealth() - larger
	if !statNear(largerBonus/larger, bonus/maxHealth) {
		t.Errorf("on %v maximum health the gem raises it by %v, want the same share as %v of %v", larger, largerBonus, bonus, maxHealth)
	}
	if got := character.CurrentHealth(); !statNear(got, larger+largerBonus) {
		t.Errorf("at full health the gem heals to %v, want %v", got, larger+largerBonus)
	}
	gift.Deactivate(sim)
	if got := character.CurrentHealth(); !statNear(got, larger) {
		t.Errorf("at full health the raise fades to %v health, want %v", got, larger)
	}
	if got := character.DamageTakenHealthMetrics.ActualGain; got != damageTaken {
		t.Errorf("damage taken is %v after the raise fades, want %v", got, damageTaken)
	}
}

// A neck that doubles the wearer's armor, a percent armor modifier.
const armorDoublingNeckID int32 = 992741

// Zandalarian Hero Badge 19948 grants armor and defense on 24574's stacks, and each melee or ranged
// hit taken takes one stack's share of each off; a spell hit taken does not. Brittle Armor 24575
// states a stack's armor with A_MOD_ARMOR_NO_MODS, so the full stacks' armor reaches the wearer whole
// past the neck's doubling.
func TestZandalarianHeroBadgeLosesArmorPerHitTaken(t *testing.T) {
	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{{Id: armorDoublingNeckID, Name: "Test Armor Doubling Neck",
		Type: proto.ItemType_ItemTypeNeck, ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}}}})
	if !core.HasItemEffect(armorDoublingNeckID) {
		core.NewItemEffect(armorDoublingNeckID, func(agent core.Agent) {
			agent.GetCharacter().MultiplyStat(stats.Armor, 2)
		})
	}
	sim, character := trinketWarriorWielding(t, 19948, 0, armorDoublingNeckID)
	before := character.GetStats()

	useTrinketAsEntryStates(t, sim, character, 19948, core.CooldownTypeSurvival)
	_, stacks := countdownBuff(t, character, 24574)
	full := fullAmounts(character, before, stats.Armor, stats.DefenseRating)
	if full[stats.Armor] <= 0 || full[stats.DefenseRating] <= 0 {
		t.Fatalf("the badge grants %v armor and %v defense rating, want both", full[stats.Armor], full[stats.DefenseRating])
	}
	stackArmor := spelldata.MustFind(24575).EffectN(1)
	if stackArmor.Aura != dbcenums.A_MOD_ARMOR_NO_MODS {
		t.Fatalf("24575's first effect is %v, want A_MOD_ARMOR_NO_MODS", stackArmor.Aura)
	}
	if want := stackArmor.BasePoints * float64(stacks); !statNear(full[stats.Armor], want) {
		t.Errorf("the badge's full stacks grant %v armor, want their %v whole", full[stats.Armor], want)
	}

	enemy := sim.Encounter.ActiveTargetUnits[0]
	takeHit(t, sim, character, meleeSwing(enemy))
	expectCountdown(t, "after a melee hit taken", character, before, full, stacks, 1)
	takeHit(t, sim, character, fireSpell(enemy))
	expectCountdown(t, "after a spell hit taken", character, before, full, stacks, 1)
}

// Zandalarian Hero Medallion 19949 adds melee and ranged damage on 24661's stacks, one stack's share
// less for each hit dealt, on the shared cooldown its entry states.
func TestZandalarianHeroMedallionLosesDamagePerHit(t *testing.T) {
	sim, character := trinketWarrior(t, 19949)
	before := character.GetStats()

	useTrinketAsEntryStates(t, sim, character, 19949, core.CooldownTypeDPS)
	buff, stacks := countdownBuff(t, character, 24661)
	full := fullAmounts(character, before, stats.PhysicalDamage)
	if full[stats.PhysicalDamage] <= 0 {
		t.Fatalf("the medallion adds no damage")
	}
	dealHit(t, sim, character, meleeSwing(&character.Unit))
	expectCountdown(t, "after a hit", character, before, full, stacks, 1)
	for range stacks - 1 {
		dealHit(t, sim, character, meleeSwing(&character.Unit))
	}
	expectCountdown(t, "after a hit per stack", character, before, full, stacks, stacks)
	if buff.IsActive() {
		t.Errorf("the buff is still up with no stacks left")
	}
}

// Zandalarian Hero Charm 19950 grants spell damage and healing on 24658's stacks, one stack's share
// less for each spell cast; a melee cast does not count.
func TestZandalarianHeroCharmLosesPowerPerSpellCast(t *testing.T) {
	sim, character := trinketWarrior(t, 19950)
	before := character.GetStats()

	useTrinketAsEntryStates(t, sim, character, 19950, core.CooldownTypeDPS)
	_, stacks := countdownBuff(t, character, 24658)
	full := fullAmounts(character, before, stats.SpellDamage, stats.HealingPower)
	if full[stats.SpellDamage] <= 0 || full[stats.HealingPower] <= 0 {
		t.Fatalf("the charm grants %v spell damage and %v healing, want both", full[stats.SpellDamage], full[stats.HealingPower])
	}

	character.OnCastComplete(sim, fireSpell(&character.Unit))
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
	expectCountdown(t, "after a spell cast", character, before, full, stacks, 1)

	character.OnCastComplete(sim, &core.Spell{ProcMask: core.ProcMaskMeleeMHSpecial, SpellSchool: core.SpellSchoolPhysical, Unit: &character.Unit})
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
	expectCountdown(t, "after a melee cast", character, before, full, stacks, 1)
}

// Petrified Scarab 21685 adds to every resistance on 26463's stacks for the row's duration, one
// stack's share less for each hostile spell that lands; a melee hit taken does not count.
func TestPetrifiedScarabLosesResistancePerSpellTaken(t *testing.T) {
	sim, character := trinketWarrior(t, 21685)
	before := character.GetStats()

	useTrinketAsEntryStates(t, sim, character, 21685, core.CooldownTypeSurvival)
	_, stacks := countdownBuff(t, character, 26463)
	full := fullAmounts(character, before, stats.ArcaneResistance, stats.FireResistance, stats.FrostResistance,
		stats.NatureResistance, stats.ShadowResistance)
	for _, stat := range []stats.Stat{stats.ArcaneResistance, stats.FireResistance, stats.FrostResistance, stats.NatureResistance, stats.ShadowResistance} {
		if full[stat] <= 0 || full[stat] != full[stats.ArcaneResistance] {
			t.Errorf("on use the scarab adds %v %s, want as much as every other resistance", full[stat], stat.StatName())
		}
	}

	enemy := sim.Encounter.ActiveTargetUnits[0]
	takeHit(t, sim, character, fireSpell(enemy))
	expectCountdown(t, "after a spell taken", character, before, full, stacks, 1)
	takeHit(t, sim, character, meleeSwing(enemy))
	expectCountdown(t, "after a melee hit taken", character, before, full, stacks, 1)
}

// Everlook Delivery Bot 274760 marks the target with 1295271; once as many spells or attacks as the
// mark's first two effects bound have landed on it, 1295633 deals its row's Fire to it and the mark
// is gone.
func TestEverlookDeliveryBotDeliversAfterItsCount(t *testing.T) {
	sim, character := trinketWarrior(t, 274760)
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100, stats.SpellCritPercent: -character.GetStat(stats.SpellCritPercent)})
	target := character.CurrentTarget
	markRow := spelldata.MustFind(1295271)
	fewest, most := int(markRow.EffectN(1).BaseValue()), int(markRow.EffectN(2).BaseValue())
	if fewest < 1 || most < fewest {
		t.Fatalf("1295271 bounds the count by %d and %d", fewest, most)
	}

	useTrinketAsEntryStates(t, sim, character, 274760, core.CooldownTypeDPS)
	mark := target.GetAuraByID(core.ActionID{SpellID: 1295271})
	if mark == nil || !mark.IsActive() {
		t.Fatalf("using the bot put no mark on the target")
	}
	delivery := character.GetSpell(core.ActionID{SpellID: 1295633})

	hits := 0
	for mark.IsActive() && hits <= most {
		target.OnSpellHitTaken(sim, meleeSwing(&character.Unit), &core.SpellResult{Target: target, Outcome: core.OutcomeHit, Damage: 100})
		stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
		hits++
	}
	if hits < fewest || hits > most {
		t.Errorf("the delivery came after %d hits, want the row's %d to %d", hits, fewest, most)
	}

	stepTo(t, sim, sim.CurrentTime+10*time.Second)
	damage := spelldata.MustFind(1295633).DamageEffect()
	metrics := delivery.SpellMetrics[target.UnitIndex]
	if metrics.Casts != 1 || !damageInRange(metrics, damage) {
		t.Errorf("the delivery cast %d times for %v damage, want once for the row's %v to %v", metrics.Casts, metrics.TotalDamage, damage.Min(), damage.Max())
	}
}

// Darkmoon Card: Heroism 19287's 23689 procs on the wearer's melee hits, and not on its spells, at the
// procs per minute of the item's entry: the chance on a swing scales with the weapon's speed. It heals
// the wearer through 23682 for that row's roll.
func TestDarkmoonCardHeroismHealsOnMeleeHits(t *testing.T) {
	procOf := func(character *core.Character) *core.Aura {
		t.Helper()
		proc := character.GetAura("Darkmoon Card: Heroism")
		if proc == nil || proc.Dpm == nil {
			t.Fatal("19287 registers no proc on a procs per minute rate")
		}
		return proc
	}
	sim, character := trinketWarrior(t, 19287)
	proc := procOf(character)
	chance, speed := proc.Dpm.Chance(core.ProcMaskMeleeMHAuto, sim), character.AutoAttacks.MH().SwingSpeed
	if want := itemEntryPPM(t, 19287, 23682) * speed / 60; want <= 0 || want >= 1 || math.Abs(chance-want) > 1e-9 {
		t.Errorf("the heal's chance on a main hand swing is %v, want the entry's procs per minute on the swing's speed, %v", chance, want)
	}
	// Ravager 7717 is only a weapon of another speed here.
	armedSim, armed := trinketWarriorWielding(t, 19287, 7717, 0)
	armedChance, armedSpeed := procOf(armed).Dpm.Chance(core.ProcMaskMeleeMHAuto, armedSim), armed.AutoAttacks.MH().SwingSpeed
	if armedSpeed == speed {
		t.Fatalf("the weapon swings at the unarmed %v, want another speed", speed)
	}
	if !statNear(armedChance/armedSpeed, chance/speed) {
		t.Errorf("the chance is %v on a %v s swing and %v on a %v s one, want it in proportion to the speed", chance, speed, armedChance, armedSpeed)
	}

	heal := character.GetSpell(core.ActionID{SpellID: 23682})
	if heal == nil {
		t.Fatal("19287 registers no heal")
	}
	// Every melee hit now procs it.
	*proc.Dpm = *character.NewFixedProcChanceManager(1, core.ProcMaskMelee)
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: -character.GetStat(stats.SpellCritPercent)})
	character.RemoveHealth(sim, character.MaxHealth()/2)

	dealHit(t, sim, character, fireSpell(&character.Unit))
	if got := heal.SpellMetrics[character.UnitIndex].Casts; got != 0 {
		t.Errorf("a spell hit healed %d times, want none", got)
	}
	dealHit(t, sim, character, meleeSwing(&character.Unit))
	amount := spelldata.MustFind(23682).HealEffect()
	metrics := heal.SpellMetrics[character.UnitIndex]
	if metrics.Casts != 1 || metrics.Crits != 0 || metrics.TotalHealing < amount.Min() || metrics.TotalHealing > amount.Max() {
		t.Errorf("a melee hit healed %d times for %v, want once for the row's %v to %v", metrics.Casts, metrics.TotalHealing, amount.Min(), amount.Max())
	}
}

// Dormant Heart of the Mountain 249473's 1249118 answers a heal of any rank of Greater Heal, Healing
// Touch, Healing Wave or Holy Light with 1249119: the row's healing and spell damage for its
// duration. Another heal, such as Flash of Light, does not.
func TestDormantHeartOfTheMountainAnswersTheFourHeals(t *testing.T) {
	sim, character := itemWearer(t, proto.Class_ClassPaladin, proto.ItemSlot_ItemSlotTrinket1, 249473)
	buff := character.GetAuraByID(core.ActionID{SpellID: 1249119})
	if buff == nil {
		t.Fatal("249473 registers no buff")
	}
	row := spelldata.MustFind(1249119)
	healing, damage := row.FirstAura(dbcenums.A_MOD_HEALING_DONE).Average(), row.FirstAura(dbcenums.A_MOD_DAMAGE_DONE).Average()
	heal := func(spellID int32) {
		t.Helper()
		spell := &core.Spell{ActionID: core.ActionID{SpellID: spellID}, ProcMask: core.ProcMaskSpellHealing, SpellSchool: core.SpellSchoolHoly, Unit: &character.Unit}
		character.OnHealDealt(sim, spell, &core.SpellResult{Target: &character.Unit, Outcome: core.OutcomeHit, Damage: 500})
		stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
	}
	before := character.GetStats()

	heal(19750) // Flash of Light rank 1
	if buff.IsActive() {
		t.Fatalf("Flash of Light raised the buff")
	}

	heal(25292) // Holy Light rank 9
	if !buff.IsActive() {
		t.Fatalf("Holy Light raised no buff")
	}
	if got := gained(character, before, stats.HealingPower); healing <= 0 || got != healing {
		t.Errorf("the buff grants %v healing power, want the row's %v", got, healing)
	}
	if got := gained(character, before, stats.SpellDamage); damage <= 0 || got != damage {
		t.Errorf("the buff grants %v spell damage, want the row's %v", got, damage)
	}
	if got := buff.Duration; got != row.Duration() {
		t.Errorf("the buff lasts %v, want the row's %v", got, row.Duration())
	}

	buff.Deactivate(sim)
	for _, rank := range []int32{2060, 5185, 331} { // Greater Heal, Healing Touch, Healing Wave rank 1
		heal(rank)
		if !buff.IsActive() {
			t.Errorf("rank %d raised no buff", rank)
		}
		buff.Deactivate(sim)
	}
}

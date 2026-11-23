//go:build with_db

package forever_test

import (
	"math"
	"slices"
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
	return trinketWarriorFacing(t, trinketID, mainHand, neck, proto.MobType_MobTypeUnknown)
}

func trinketWarriorFacing(t *testing.T, trinketID int32, mainHand int32, neck int32, mobType proto.MobType, areas ...proto.AreaType) (*core.Simulation, *core.Character) {
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
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60, MobType: mobType}}, Duration: 180, AreaTypes: areas},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

// Seal of the Dawn 13209 and Rune of the Dawn 19812 state their bonus against Undead as item stats 128 and
// 136, at the value their equip spells 23930 and 24198 state. Counterattack Lodestone 18537 and Royal Seal
// of Eldre'Thalas 18467 fold equip spells 1318938 and 1318479 into their stats, the bonus against Mechanical
// and Demon targets with them. The sim adds each bonus to the attack table of a target of that mob type
// alone.
func TestTrinketsGrantTheirBonusAgainstTheirMobTypeAlone(t *testing.T) {
	seal, rune := spelldata.MustFind(23930), spelldata.MustFind(24198)
	lodestone, royalSeal := spelldata.MustFind(1318938), spelldata.MustFind(1318479)
	for _, c := range []struct {
		item    int32
		against proto.MobType
		want    stats.Stats
	}{
		{13209, proto.MobType_MobTypeUndead, stats.Stats{stats.AttackPower: seal.EffectN(1).Average(), stats.RangedAttackPower: seal.EffectN(2).Average()}},
		{19812, proto.MobType_MobTypeUndead, stats.Stats{stats.SpellDamage: rune.EffectN(1).Average()}},
		{18537, proto.MobType_MobTypeMechanical, stats.Stats{stats.AttackPower: lodestone.EffectN(3).Average(), stats.RangedAttackPower: lodestone.EffectN(4).Average()}},
		{18467, proto.MobType_MobTypeDemon, stats.Stats{stats.SpellDamage: royalSeal.EffectN(3).Average()}},
	} {
		_, against := trinketWarriorFacing(t, c.item, 0, 0, c.against)
		if got := bonusAgainstTarget(against); got != c.want || c.want == (stats.Stats{}) {
			t.Errorf("%d: the bonus against %v is %v, want the row's %v", c.item, c.against, got, c.want)
		}
		_, againstBeasts := trinketWarriorFacing(t, c.item, 0, 0, proto.MobType_MobTypeBeast)
		if got := bonusAgainstTarget(againstBeasts); got != (stats.Stats{}) {
			t.Errorf("%d: the bonus against Beasts is %v, want none", c.item, got)
		}
	}
}

// Draconic Infused Emblem 22268's buff 1318930 adds its spell damage a second time against Dragonkin
// while it is up, and nothing against other targets.
func TestDraconicInfusedEmblemBuffGrantsSpellDamageAgainstDragonkin(t *testing.T) {
	row := spelldata.MustFind(1318930)
	for _, c := range []struct {
		against proto.MobType
		want    stats.Stats
	}{
		{proto.MobType_MobTypeDragonkin, stats.Stats{stats.SpellDamage: row.EffectN(3).Average()}},
		{proto.MobType_MobTypeBeast, stats.Stats{}},
	} {
		sim, character := trinketWarriorFacing(t, 22268, 0, 0, c.against)
		buff := character.GetAuraByID(core.ActionID{SpellID: 1318930})
		if buff == nil {
			t.Fatal("22268 registers no 1318930")
		}
		if got := bonusAgainstTarget(character); got != (stats.Stats{}) {
			t.Errorf("against %v the bonus before the buff is %v, want none", c.against, got)
		}
		buff.Activate(sim)
		if got := bonusAgainstTarget(character); got != c.want {
			t.Errorf("against %v the bonus with the buff up is %v, want %v", c.against, got, c.want)
		}
		buff.Deactivate(sim)
		if got := bonusAgainstTarget(character); got != (stats.Stats{}) {
			t.Errorf("against %v the bonus after the buff is %v, want none", c.against, got)
		}
	}
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

func stepWhile(t *testing.T, sim *core.Simulation, cond func() bool) {
	t.Helper()
	for steps := 0; cond(); steps++ {
		if steps > 1000 {
			t.Fatal("the sim did not get there in 1000 steps")
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

// Aegis of Preservation 19345's on-use 23780 grants its armor and, while it runs, heals the wearer
// 23781's amount on each melee or ranged hit taken, read from the generated stat on-use: nothing on a
// spell hit, and nothing once the 20 s buff has run out.
func TestAegisOfPreservationHealsOnHitsTakenWhileItRuns(t *testing.T) {
	sim, character := itemWearer(t, proto.Class_ClassPriest, proto.ItemSlot_ItemSlotTrinket1, 19345)
	character.RemoveHealth(sim, character.MaxHealth()/2)
	character.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: -character.GetStat(stats.SpellCritPercent)})
	enemy := sim.Encounter.ActiveTargetUnits[0]
	amount := spelldata.MustFind(23781).ProcHealEffect().Average()

	use := useTrinketAsEntryStates(t, sim, character, 19345, core.CooldownTypeSurvival)
	heal := character.GetSpell(core.ActionID{SpellID: 23781})
	if heal == nil {
		t.Fatal("19345's buff carries no heal proc")
	}

	takeHit(t, sim, character, meleeSwing(enemy))
	takeHit(t, sim, character, rangedShot(enemy))
	takeHit(t, sim, character, fireSpell(enemy))
	if got := heal.SpellMetrics[character.UnitIndex]; got.Casts != 2 || amount <= 0 || got.TotalHealing != 2*amount {
		t.Errorf("a melee, a ranged and a spell hit taken cast the heal %d times for %v, want twice for the row's %v", got.Casts, got.TotalHealing, amount)
	}

	stepTo(t, sim, use.RelatedSelfBuff.ExpiresAt()+time.Millisecond)
	takeHit(t, sim, character, meleeSwing(enemy))
	if got := heal.SpellMetrics[character.UnitIndex].Casts; got != 2 {
		t.Errorf("a melee hit taken after the buff ran out left the heal at %d casts, want 2 still", got)
	}
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

// The damage shield spell a row registers on the character, and the amount the row deals per hit.
func damageShield(t *testing.T, character *core.Character, rowID int32) (*core.Spell, float64) {
	t.Helper()
	var amount float64
	for _, effect := range spelldata.ItemAuraDamageShields(spelldata.MustFind(rowID)) {
		amount += effect.Average()
	}
	shield := character.GetSpell(core.ActionID{SpellID: rowID, Tag: 2})
	if shield == nil || amount <= 0 {
		t.Fatalf("%d registers no damage shield", rowID)
	}
	return shield, amount
}

func rangedShot(unit *core.Unit) *core.Spell {
	return &core.Spell{ProcMask: core.ProcMaskRangedAuto, SpellSchool: core.SpellSchoolPhysical, Unit: unit}
}

func fireSpell(unit *core.Unit) *core.Spell {
	return &core.Spell{ProcMask: core.ProcMaskSpellDamage, SpellSchool: core.SpellSchoolFire, DefenseType: core.DefenseTypeMagic, Unit: unit}
}

func healSpell(unit *core.Unit) *core.Spell {
	return &core.Spell{ProcMask: core.ProcMaskSpellHealing, SpellSchool: core.SpellSchoolHoly, DefenseType: core.DefenseTypeMagic, Unit: unit}
}

// A landed heal on the character, and the spell batch window the listener resolves after.
func dealHeal(t *testing.T, sim *core.Simulation, character *core.Character, spell *core.Spell) {
	t.Helper()
	character.OnHealDealt(sim, spell, &core.SpellResult{Target: &character.Unit, Outcome: core.OutcomeHit, Damage: 100})
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
}

// A landed hit taken, and the spell batch window the listener resolves after.
func takeHit(t *testing.T, sim *core.Simulation, character *core.Character, spell *core.Spell) {
	t.Helper()
	takeOutcome(t, sim, character, spell, core.OutcomeHit, 100)
}

// A hit taken with the outcome and damage, and the spell batch window the listener resolves after.
func takeOutcome(t *testing.T, sim *core.Simulation, character *core.Character, spell *core.Spell, outcome core.HitOutcome, damage float64) {
	t.Helper()
	character.OnSpellHitTaken(sim, spell, &core.SpellResult{Target: &character.Unit, Outcome: outcome, Damage: damage})
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
}

// A melee swing the character dodges, and the spell batch window the listener resolves after.
func dodgeSwing(t *testing.T, sim *core.Simulation, character *core.Character, attacker *core.Unit) {
	t.Helper()
	character.OnSpellHitTaken(sim, meleeSwing(attacker), &core.SpellResult{Target: &character.Unit, Outcome: core.OutcomeDodge})
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
}

func dealHit(t *testing.T, sim *core.Simulation, character *core.Character, spell *core.Spell) {
	t.Helper()
	character.OnSpellHitDealt(sim, spell, &core.SpellResult{Target: character.CurrentTarget, Outcome: core.OutcomeHit, Damage: 100})
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
}

// The buff the item's use put up, checked to carry all of at least 2 stacks or charges for the row's
// duration: a Zandalarian trinket's or the Petrified Scarab's countdown, Vigilance Charm's charges.
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
	// Hammer of the Northern Wind 810 is only a weapon of another speed here.
	armedSim, armed := trinketWarriorWielding(t, 19287, 810, 0)
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

// Highborne Research Tablet 260205's 1318159 rolls its row's procs per minute on the main hand's
// speed for a landed main-hand hit that deals damage, and casts 1265634 on the target. A ranged
// special with no ranged weapon behind it rolls on the main hand's speed too, Thunder Clap excepted,
// which never rolls it. A hit that deals none (an application) never rolls.
func TestHighborneResearchTabletProcsOnLandedDamagingHits(t *testing.T) {
	sim, character := trinketWarriorWielding(t, 260205, 810, 0)
	proc := character.GetAura("Highborne Research Tablet")
	damage := character.GetSpell(core.ActionID{SpellID: 1265634})
	if proc == nil || proc.Dpm == nil || damage == nil {
		t.Fatal("260205 registers no proc on a procs per minute rate")
	}
	ppm := float64(spelldata.MustFind(1318159).RPPM)
	want := ppm * character.AutoAttacks.MH().SwingSpeed / 60
	if ppm <= 0 || want >= 1 {
		t.Fatalf("the row's %v procs per minute on the main hand give %v, want a chance between none and certain", ppm, want)
	}
	for _, hit := range []struct {
		name string
		mask core.ProcMask
	}{
		{"main hand swing", core.ProcMaskMeleeMHAuto},
		{"ranged special", core.ProcMaskRangedSpecial},
	} {
		if chance := proc.Dpm.Chance(hit.mask, sim); math.Abs(chance-want) > 1e-9 {
			t.Errorf("the chance on a %s is %v, want the row's procs per minute on the main hand's speed, %v", hit.name, chance, want)
		}
	}

	*proc.Dpm = *character.NewFixedProcChanceManager(1, core.ProcMaskMelee)
	character.OnSpellHitDealt(sim, meleeSwing(&character.Unit), &core.SpellResult{Target: character.CurrentTarget, Outcome: core.OutcomeHit})
	stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
	if got := damage.SpellMetrics[character.CurrentTarget.UnitIndex].Casts; got != 0 {
		t.Errorf("a hit that dealt no damage cast it %d times, want none", got)
	}
	dealHit(t, sim, character, meleeSwing(&character.Unit))
	if got := damage.SpellMetrics[character.CurrentTarget.UnitIndex].Casts; got != 1 {
		t.Errorf("a landed damaging hit cast it %d times, want once", got)
	}

	*proc.Dpm = *character.NewFixedProcChanceManager(1, core.ProcMaskMelee|core.ProcMaskRanged)
	thunderClap := &core.Spell{ProcMask: core.ProcMaskRangedSpecial, SpellSchool: core.SpellSchoolPhysical, DefenseType: core.DefenseTypeMagic,
		ClassFlags: spelldata.MustFind(6343).ClassFlags, Unit: &character.Unit}
	dealHit(t, sim, character, thunderClap)
	if got := damage.SpellMetrics[character.CurrentTarget.UnitIndex].Casts; got != 1 {
		t.Errorf("a landed Thunder Clap cast it %d more times, want none", got-1)
	}
}

// Highborne Research Tablet 260205 rolls its row's procs per minute on each damaging spell hit,
// priced on the spell's cast time and never under 1.5 s, whatever the weapon in the main hand, for a
// caster who never swings and for a warrior who does.
func TestHighborneResearchTabletPricesSpellsOnCastTime(t *testing.T) {
	const slow, fast = 810, 776
	if core.GetItemByID(slow).SwingSpeed == core.GetItemByID(fast).SwingSpeed {
		t.Fatal("the two weapons share a speed, want two speeds")
	}
	casterWielding := func(weapon int32) (*core.Simulation, *core.Character) {
		return wearing(t, proto.Class_ClassWarlock, &proto.Player_Warlock{Warlock: &proto.Warlock{}},
			map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: 260205, proto.ItemSlot_ItemSlotMainHand: weapon})
	}
	sim, character := casterWielding(slow)
	if character.AutoAttacks.AutoSwingMelee || character.AutoAttacks.AutoSwingRanged {
		t.Fatal("the caster stand-in swings, want no auto attacks")
	}
	proc := character.GetAura("Highborne Research Tablet")
	damage := character.GetSpell(core.ActionID{SpellID: 1265634})
	if proc == nil || proc.Dpm == nil || damage == nil {
		t.Fatal("260205 registers no proc on a procs per minute rate")
	}
	ppm := float64(spelldata.MustFind(1318159).RPPM)
	instant := ppm * core.GCDDefault.Seconds() / 60
	if ppm <= 0 || instant >= 1 {
		t.Fatalf("the row's %v procs per minute on 1.5 s give %v, want a chance between none and certain", ppm, instant)
	}
	fastSim, fastCaster := casterWielding(fast)
	warriorSim, warrior := trinketWarriorWielding(t, 260205, fast, 0)
	for _, wearer := range []struct {
		name   string
		sim    *core.Simulation
		wearer *core.Character
	}{
		{"caster", sim, character},
		{"caster with a faster weapon", fastSim, fastCaster},
		{"warrior who swings", warriorSim, warrior},
	} {
		if chance := wearer.wearer.GetAura("Highborne Research Tablet").Dpm.Chance(core.ProcMaskSpellDamage, wearer.sim); math.Abs(chance-instant) > 1e-9 {
			t.Errorf("the chance on an instant spell hit for a %s is %v, want the row's procs per minute on 1.5 s, %v", wearer.name, chance, instant)
		}
	}

	const hits = 2000
	casts := func() float64 { return float64(damage.SpellMetrics[character.CurrentTarget.UnitIndex].Casts) }
	for _, cast := range []time.Duration{0, 3 * time.Second} {
		want := ppm * max(cast, core.GCDDefault).Seconds() / 60
		before := casts()
		spell := fireSpell(&character.Unit)
		spell.CurCast.CastTime = cast
		for range hits {
			dealHit(t, sim, character, spell)
		}
		mean, deviation := want*hits, math.Sqrt(want*(1-want)*hits)
		if got := casts() - before; math.Abs(got-mean) > 4*deviation {
			t.Errorf("%d hits of a spell cast in %v cast it %v times, want about %.0f", hits, cast, got, mean)
		}
	}
}

// Scented Runewood Brooch 269741's 1318323 rolls its row's procs per minute on the main hand's speed
// for a landed main-hand swing and a landed spell hit alike, and casts 1318324 on the target, but only
// when the target is a Furbolg. A hit on any other target never casts it.
func TestScentedRunewoodBroochProcsOnlyAgainstFurbolg(t *testing.T) {
	sim, character := trinketWarriorWielding(t, 269741, 810, 0)
	proc := character.GetAura("Scented Runewood Brooch")
	damage := character.GetSpell(core.ActionID{SpellID: 1318324})
	if proc == nil || proc.Dpm == nil || damage == nil {
		t.Fatal("269741 registers no damage spell or no proc on a procs per minute rate")
	}
	ppm := float64(spelldata.MustFind(1318323).RPPM)
	want := ppm * character.AutoAttacks.MH().SwingSpeed / 60
	if ppm <= 0 || want >= 1 {
		t.Fatalf("the row's %v procs per minute on the main hand give %v, want a chance between none and certain", ppm, want)
	}
	for _, hit := range []struct {
		name string
		mask core.ProcMask
	}{
		{"main hand swing", core.ProcMaskMeleeMHAuto},
		{"spell hit", core.ProcMaskSpellDamage},
	} {
		if chance := proc.Dpm.Chance(hit.mask, sim); math.Abs(chance-want) > 1e-9 {
			t.Errorf("the chance on a %s is %v, want the row's procs per minute on the main hand's speed, %v", hit.name, chance, want)
		}
	}

	// Every hit now rolls a proc.
	*proc.Dpm = *character.NewFixedProcChanceManager(1, core.ProcMaskMelee|core.ProcMaskSpellDamage)
	enemy := character.CurrentTarget
	casts := func() int32 { return damage.SpellMetrics[enemy.UnitIndex].Casts }
	if enemy.CreatureFamily == proto.CreatureFamily_CreatureFamilyFurbolg {
		t.Fatal("the target is a Furbolg before the test makes it one")
	}
	dealHit(t, sim, character, meleeSwing(&character.Unit))
	dealHit(t, sim, character, fireSpell(&character.Unit))
	if got := casts(); got != 0 {
		t.Errorf("a swing and a spell hit on another target cast it %d times, want none", got)
	}

	enemy.CreatureFamily = proto.CreatureFamily_CreatureFamilyFurbolg
	dealHit(t, sim, character, meleeSwing(&character.Unit))
	if got := casts(); got != 1 {
		t.Errorf("a swing on a Furbolg cast it %d times, want once", got)
	}
	dealHit(t, sim, character, fireSpell(&character.Unit))
	if got := casts(); got != 2 {
		t.Errorf("a swing and a spell hit on a Furbolg cast it %d times, want twice", got)
	}
}

// Scented Runewood Brooch 269741 rolls its row's procs per minute on a caster's spell hits at the
// main-hand weapon's speed though the caster never swings it, and a spell hit casts 1318324 only on
// a Furbolg.
func TestScentedRunewoodBroochProcsOnACastersSpellsOnTheMainHandsSpeed(t *testing.T) {
	const slow, fast = 810, 776
	if core.GetItemByID(slow).SwingSpeed == core.GetItemByID(fast).SwingSpeed {
		t.Fatal("the two weapons share a speed, want two speeds")
	}
	ppm := float64(spelldata.MustFind(1318323).RPPM)
	for _, weapon := range []int32{slow, fast} {
		sim, character := wearing(t, proto.Class_ClassWarlock, &proto.Player_Warlock{Warlock: &proto.Warlock{}},
			map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: 269741, proto.ItemSlot_ItemSlotMainHand: weapon})
		if character.AutoAttacks.AutoSwingMelee || character.AutoAttacks.AutoSwingRanged {
			t.Fatal("the caster stand-in swings, want no auto attacks")
		}
		proc := character.GetAura("Scented Runewood Brooch")
		damage := character.GetSpell(core.ActionID{SpellID: 1318324})
		if proc == nil || proc.Dpm == nil || damage == nil {
			t.Fatal("269741 registers no damage spell or no proc on a procs per minute rate")
		}
		want := ppm * core.GetItemByID(weapon).SwingSpeed / 60
		if chance := proc.Dpm.Chance(core.ProcMaskSpellDamage, sim); ppm <= 0 || want >= 1 || math.Abs(chance-want) > 1e-9 {
			t.Errorf("with %d in the main hand a spell hit rolls %v, want the row's procs per minute on its speed, %v", weapon, chance, want)
		}

		// Every spell hit now rolls a proc.
		*proc.Dpm = *character.NewTargetScaledChanceManager(1, core.ProcMaskSpellDamage, func(*core.Unit) float64 { return 1 })
		enemy := character.CurrentTarget
		dealHit(t, sim, character, fireSpell(&character.Unit))
		if got := damage.SpellMetrics[enemy.UnitIndex].Casts; got != 0 {
			t.Errorf("a spell hit on another target cast it %d times, want none", got)
		}
		enemy.CreatureFamily = proto.CreatureFamily_CreatureFamilyFurbolg
		dealHit(t, sim, character, fireSpell(&character.Unit))
		if got := damage.SpellMetrics[enemy.UnitIndex].Casts; got != 1 {
			t.Errorf("a spell hit on a Furbolg cast it %d times, want once", got)
		}
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

// Vigilance Charm 18370's Vigilant 1302264 goes up with its dodge and its charges, and each attack the
// wearer dodges spends one: a landed hit taken spends none, and the last dodge takes the buff down.
func TestVigilanceCharmSpendsAChargePerDodge(t *testing.T) {
	sim, character := trinketWarrior(t, 18370)
	before := character.GetStats()

	useTrinketAsEntryStates(t, sim, character, 18370, core.CooldownTypeSurvival)
	buff, charges := countdownBuff(t, character, 1302264)
	if got := gained(character, before, stats.DodgePercent); got <= 0 {
		t.Errorf("the buff grants %v dodge, want the row's", got)
	}

	enemy := sim.Encounter.ActiveTargetUnits[0]
	takeHit(t, sim, character, meleeSwing(enemy))
	if got := buff.GetStacks(); !buff.IsActive() || got != charges {
		t.Errorf("after a landed hit taken the buff has %d charges, want all %d", got, charges)
	}

	for dodged := int32(1); dodged < charges; dodged++ {
		dodgeSwing(t, sim, character, enemy)
		if got := buff.GetStacks(); !buff.IsActive() || got != charges-dodged {
			t.Fatalf("after %d dodges the buff has %d charges, want %d", dodged, got, charges-dodged)
		}
	}
	dodgeSwing(t, sim, character, enemy)
	if buff.IsActive() {
		t.Errorf("the buff is still up after %d dodges, want it gone", charges)
	}
}

// Edits the store's row of a spell for the test, restoring it afterwards, and returns the row. The
// effects are copied so an edit to one does not reach the original.
func editSpellRow(t *testing.T, id int32, edit func(*spelldata.Spell)) *spelldata.Spell {
	t.Helper()
	row := spelldata.MustFind(id)
	original := *row
	row.Effects = slices.Clone(row.Effects)
	edit(row)
	t.Cleanup(func() { *row = original })
	return row
}

// A warrior wearing the extra attack item, the trinket in the first trinket slot and any other item
// in the main hand, who swings only when the test deals a hit. The returned func counts the procs of
// the extra attack the row's first effect grants: the extra attacks swung over the extra attacks
// one proc swings.
func extraAttackWearer(t *testing.T, itemID int32, rowID int32) (*core.Simulation, *core.Character, func() int32) {
	t.Helper()
	var sim *core.Simulation
	var character *core.Character
	if core.GetItemByID(itemID).Type == proto.ItemType_ItemTypeTrinket {
		sim, character = trinketWarrior(t, itemID)
	} else {
		sim, character = weaponWearer(t, proto.Class_ClassWarrior, itemID, 1)
	}
	character.AutoAttacks.CancelAutoSwing(sim)
	granted := spelldata.MustFind(rowID).EffectN(1).Trigger()
	extraAttack := character.GetSpell(core.ActionID{OtherID: proto.OtherAction_OtherActionAttack, Tag: granted.ID})
	perProc := int32(granted.ExtraAttacksEffect().BasePoints)
	if extraAttack == nil || perProc <= 0 {
		t.Fatalf("%d registers no extra attack", itemID)
	}
	enemy := character.CurrentTarget
	return sim, character, func() int32 { return extraAttack.SpellMetrics[enemy.UnitIndex].Casts / perProc }
}

// Checks registerExtraAttackVersus's rule for an item: on a target of the family it rolls the row's
// column, on any other target the column divided by the factor, and states that chance against its
// current target. Each case counts the procs of many
// landed melee hits, within five standard deviations of the count. The row rests no cooldown, so every
// hit rolls.
func expectExtraAttackVersus(t *testing.T, itemID int32, name string, rowID int32, family proto.CreatureFamily) {
	t.Helper()
	row := editSpellRow(t, rowID, func(s *spelldata.Spell) { s.ICDMs = 0 })
	column, factor := float64(row.ProcChance)/100, row.EffectN(2).BasePoints
	if column <= 0 || factor <= 1 {
		t.Fatalf("%d states a chance of %v and a factor of %v, want a chance and a factor above 1 to tell the two apart", rowID, column, factor)
	}

	for _, target := range []struct {
		family proto.CreatureFamily
		want   float64
	}{{family, column}, {proto.CreatureFamily_CreatureFamilyUnknown, column / factor}} {
		t.Run(target.family.String(), func(t *testing.T) {
			sim, character, procs := extraAttackWearer(t, itemID, rowID)
			enemy := character.CurrentTarget
			enemy.CreatureFamily = target.family
			trigger := character.GetAura(name)
			if trigger == nil || trigger.Dpm == nil {
				t.Fatalf("%d registers no proc on a chance manager", itemID)
			}
			if got := trigger.Dpm.Chance(core.ProcMaskMeleeMHAuto, sim); math.Abs(got-target.want) > 1e-9 {
				t.Errorf("the stated chance against the current target is %v, want %v", got, target.want)
			}
			swing := meleeSwing(&character.Unit)
			const hits = 40000
			for range hits {
				character.OnSpellHitDealt(sim, swing, &core.SpellResult{Target: enemy, Outcome: core.OutcomeHit, Damage: 100})
			}
			stepTo(t, sim, sim.CurrentTime+time.Millisecond)

			if spread := 5 * math.Sqrt(hits*target.want*(1-target.want)); math.Abs(float64(procs())-hits*target.want) > spread {
				t.Errorf("%d of %d landed melee hits procced it, want %v within %v", procs(), hits, hits*target.want, spread)
			}
		})
	}
}

// Hand of Justice 11815 rolls 15600's ProcChance on each landed melee hit on a Dwarf, and on any other
// target the column divided by its second effect's value (the rule registerExtraAttackVersus states).
func TestHandOfJusticeIsLikelierAgainstDwarves(t *testing.T) {
	expectExtraAttackVersus(t, 11815, "Hand of Justice", 15600, proto.CreatureFamily_CreatureFamilyDwarf)
}

// Hand of Justice 11815 grants nothing on a landed melee hit inside 15600's cooldown after a proc, and
// rolls again once the cooldown ends. The row copy procs on every hit.
func TestHandOfJusticeRestsForItsRowsCooldown(t *testing.T) {
	editSpellRow(t, 15600, func(s *spelldata.Spell) {
		s.ProcChance, s.ProcChanceSource = 100, spelldata.ProcChanceColumn
		s.EffectN(2).BasePoints = 1
	})
	cooldown := spelldata.MustFind(15600).ICD()
	if cooldown <= 0 {
		t.Fatalf("15600 states a cooldown of %v, want one", cooldown)
	}

	sim, character, procs := extraAttackWearer(t, 11815, 15600)

	start := sim.CurrentTime
	dealHit(t, sim, character, meleeSwing(&character.Unit))
	if got := procs(); got != 1 {
		t.Fatalf("a landed melee hit procced it %d times, want once", got)
	}
	stepTo(t, sim, start+cooldown-time.Millisecond)
	dealHit(t, sim, character, meleeSwing(&character.Unit))
	if got := procs(); got != 1 {
		t.Errorf("a hit 1 ms before the cooldown ends procced it %d times in all, want once still", got)
	}
	stepTo(t, sim, start+cooldown)
	dealHit(t, sim, character, meleeSwing(&character.Unit))
	if got := procs(); got != 2 {
		t.Errorf("a hit as the cooldown ends procced it %d times in all, want twice", got)
	}
}

// Mark of the Chosen 17774 puts up 21970 for its row's duration on a melee, ranged or harmful spell
// hit taken, and a heal taken never does. After a proc of any kind every hit rests for 21969's
// cooldown, which outlasts the buff. The row copy procs on every hit.
func TestMarkOfTheChosenProcsOnHitsTakenAndRests(t *testing.T) {
	editSpellRow(t, 21969, func(s *spelldata.Spell) {
		s.ProcChance, s.ProcChanceSource = 100, spelldata.ProcChanceColumn
	})
	cooldown, duration := spelldata.MustFind(21969).ICD(), spelldata.MustFind(21970).Duration()
	if duration <= 0 || cooldown <= duration {
		t.Fatalf("21969's cooldown is %v and 21970 lasts %v, want a buff that ends inside the cooldown", cooldown, duration)
	}
	wearer := func(t *testing.T) (*core.Simulation, *core.Character, *core.Aura) {
		t.Helper()
		sim, character := trinketWarrior(t, 17774)
		buff := character.GetAura("Mark of the Chosen Proc")
		if buff == nil {
			t.Fatal("17774 registers no proc buff")
		}
		return sim, character, buff
	}

	t.Run("heal taken", func(t *testing.T) {
		sim, character, buff := wearer(t)
		heal := &core.Spell{ProcMask: core.ProcMaskSpellHealing, SpellSchool: core.SpellSchoolHoly, Unit: &character.Unit}
		character.OnHealTaken(sim, heal, &core.SpellResult{Target: &character.Unit, Outcome: core.OutcomeHit, Damage: 100})
		stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
		if buff.IsActive() {
			t.Error("a heal taken put the buff up, want no buff")
		}
	})

	for _, hit := range []struct {
		name  string
		spell func(*core.Unit) *core.Spell
	}{{"melee", meleeSwing}, {"ranged", rangedShot}, {"spell", fireSpell}} {
		t.Run(hit.name, func(t *testing.T) {
			sim, character, buff := wearer(t)
			takeHit(t, sim, character, hit.spell(sim.Encounter.ActiveTargetUnits[0]))
			if !buff.IsActive() || buff.Duration != duration {
				t.Errorf("after a %s hit taken the buff is active %v and lasts %v, want it active for the row's %v", hit.name, buff.IsActive(), buff.Duration, duration)
			}
		})
	}

	t.Run("cooldown", func(t *testing.T) {
		sim, character, buff := wearer(t)
		enemy := sim.Encounter.ActiveTargetUnits[0]
		start := sim.CurrentTime
		takeHit(t, sim, character, meleeSwing(enemy))
		// stepTo stops after 10,000 steps, fewer than the sim takes to cross the whole cooldown after
		// a hit taken, so the cooldown is crossed in 10 s legs.
		for at := start + 10*time.Second; at < start+cooldown; at += 10 * time.Second {
			stepTo(t, sim, at)
		}
		stepTo(t, sim, start+cooldown-time.Millisecond)
		takeHit(t, sim, character, rangedShot(enemy))
		if buff.IsActive() {
			t.Error("a ranged hit taken 1 ms before a melee proc's cooldown ends put the buff up, want none")
		}
		stepTo(t, sim, start+cooldown)
		takeHit(t, sim, character, fireSpell(enemy))
		if !buff.IsActive() {
			t.Error("a spell hit taken as the cooldown ends left the buff down, want it up")
		}
	})
}

// The Lion Horn of Stormwind 14557 rolls half of 20847's column on a melee hit taken, and the whole
// column when the encounter is in Strongholds and Cities. Each case counts the procs of many hits
// taken, within four standard deviations of the count. The row copy rests no cooldown, so every hit
// rolls.
func TestLionHornOfStormwindDoublesItsChanceInStrongholdsAndCities(t *testing.T) {
	row := editSpellRow(t, 20847, func(s *spelldata.Spell) { s.ICDMs = 0 })
	column := row.StatedChance()
	if row.ICD() != 0 || column <= 0 || column >= 1 {
		t.Fatalf("20847's copy rests %v and states a chance of %v, want no cooldown and a chance between none and certain", row.ICD(), column)
	}

	for _, tc := range []struct {
		name  string
		areas []proto.AreaType
		want  float64
	}{
		{"outside Strongholds and Cities", nil, column / 2},
		{"in Strongholds and Cities", []proto.AreaType{proto.AreaType_AreaTypeStrongholdsCities}, column},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim, character := trinketWarriorFacing(t, 14557, 0, 0, proto.MobType_MobTypeUnknown, tc.areas...)
			buff := character.GetAura("The Lion Horn of Stormwind Proc")
			if buff == nil {
				t.Fatal("14557 registers no proc buff")
			}
			enemy := sim.Encounter.ActiveTargetUnits[0]
			const hits = 4000
			var procs int
			for range hits {
				buff.Deactivate(sim)
				takeHit(t, sim, character, meleeSwing(enemy))
				if buff.IsActive() {
					procs++
				}
			}
			mean, deviation := tc.want*hits, math.Sqrt(tc.want*(1-tc.want)*hits)
			if math.Abs(float64(procs)-mean) > 4*deviation {
				t.Errorf("%d melee hits taken %s procced it %d times, want about %.0f", hits, tc.name, procs, mean)
			}
		})
	}
}

// Draconic Infused Emblem 22268's 1318931 rolls its row's chance on every landed harmful spell hit,
// one that deals no damage too, and on nothing else: a cast with no hit, a miss, a damage over time
// tick and a spell fired as a proc never roll.
func TestDraconicInfusedEmblemRollsOnLandedSpellHits(t *testing.T) {
	const hurricaneTick, flamestrikeTick = 1278759, 1279990
	if !spelldata.MustFind(hurricaneTick).IsAProc() || spelldata.MustFind(flamestrikeTick).IsAProc() {
		t.Fatalf("want %d's row a proc and %d's not one", hurricaneTick, flamestrikeTick)
	}
	wearer := func(t *testing.T) (*core.Simulation, *core.Character, *core.Aura) {
		t.Helper()
		sim, character := trinketWarrior(t, 22268)
		buff := character.GetAuraByID(core.ActionID{SpellID: 1318930})
		if buff == nil {
			t.Fatal("22268 registers no 1318930")
		}
		return sim, character, buff
	}
	// How many of the given events put the buff up, taking it down before each.
	procs := func(t *testing.T, sim *core.Simulation, buff *core.Aura, events int, event func()) int {
		t.Helper()
		var procs int
		for range events {
			buff.Deactivate(sim)
			event()
			stepTo(t, sim, sim.CurrentTime+core.SpellBatchWindow)
			if buff.IsActive() {
				procs++
			}
		}
		return procs
	}

	t.Run("chance", func(t *testing.T) {
		sim, character, buff := wearer(t)
		spell := fireSpell(&character.Unit)
		want := spelldata.MustFind(1318931).StatedChance()
		if want <= 0 || want >= 1 {
			t.Fatalf("1318931 states a chance of %v, want one between none and certain", want)
		}
		const hits = 4000
		got := procs(t, sim, buff, hits, func() {
			character.OnSpellHitDealt(sim, spell, &core.SpellResult{Target: character.CurrentTarget, Outcome: core.OutcomeHit, Damage: 100})
		})
		mean, deviation := want*hits, math.Sqrt(want*(1-want)*hits)
		if math.Abs(float64(got)-mean) > 4*deviation {
			t.Errorf("%d landed spell hits procced it %d times, want about %.0f", hits, got, mean)
		}
	})

	for _, event := range []struct {
		name string
		want int
		deal func(*core.Simulation, *core.Character, *core.Spell)
	}{
		{"landed hit with no damage", 1, func(sim *core.Simulation, character *core.Character, spell *core.Spell) {
			character.OnSpellHitDealt(sim, spell, &core.SpellResult{Target: character.CurrentTarget, Outcome: core.OutcomeHit})
		}},
		{"cast with no hit", 0, func(sim *core.Simulation, character *core.Character, spell *core.Spell) {
			character.OnCastComplete(sim, spell)
		}},
		{"miss", 0, func(sim *core.Simulation, character *core.Character, spell *core.Spell) {
			character.OnSpellHitDealt(sim, spell, &core.SpellResult{Target: character.CurrentTarget, Outcome: core.OutcomeMiss})
		}},
		{"damage over time tick", 0, func(sim *core.Simulation, character *core.Character, spell *core.Spell) {
			character.OnPeriodicDamageDealt(sim, spell, &core.SpellResult{Target: character.CurrentTarget, Outcome: core.OutcomeHit, Damage: 100})
		}},
		{"landed hit of a proc", 0, func(sim *core.Simulation, character *core.Character, spell *core.Spell) {
			spell.Flags |= core.SpellFlagProc
			character.OnSpellHitDealt(sim, spell, &core.SpellResult{Target: character.CurrentTarget, Outcome: core.OutcomeHit, Damage: 100})
		}},
		{"landed hit of a spell fired as a proc whose row is one", 0, func(sim *core.Simulation, character *core.Character, spell *core.Spell) {
			spell.Flags |= core.SpellFlagPassiveSpell
			spell.ActionID = core.ActionID{SpellID: hurricaneTick}
			character.OnSpellHitDealt(sim, spell, &core.SpellResult{Target: character.CurrentTarget, Outcome: core.OutcomeHit, Damage: 100})
		}},
		{"landed hit of a spell fired as a proc whose row is not one", 1, func(sim *core.Simulation, character *core.Character, spell *core.Spell) {
			spell.Flags |= core.SpellFlagPassiveSpell
			spell.ActionID = core.ActionID{SpellID: flamestrikeTick}
			character.OnSpellHitDealt(sim, spell, &core.SpellResult{Target: character.CurrentTarget, Outcome: core.OutcomeHit, Damage: 100})
		}},
	} {
		t.Run(event.name, func(t *testing.T) {
			editSpellRow(t, 1318931, func(s *spelldata.Spell) {
				s.ProcChance, s.ProcChanceSource = 100, spelldata.ProcChanceColumn
			})
			sim, character, buff := wearer(t)
			spell := fireSpell(&character.Unit)
			if got := procs(t, sim, buff, 1, func() { event.deal(sim, character, spell) }); got != event.want {
				t.Errorf("a %s procced it %d times, want %d", event.name, got, event.want)
			}
		})
	}
}

//go:build with_db

package forever_test

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/core/testagent"
	_ "github.com/wowsims/forever/sim/hunter" // registers the hunter's item effects.
	"github.com/wowsims/forever/sim/priest"
	"github.com/wowsims/forever/sim/shaman"
	"github.com/wowsims/forever/sim/shaman/enhancement"
	_ "github.com/wowsims/forever/sim/warrior" // registers the warrior's item effects.
)

// The stand-ins' spells, under the class flags of the spells the items name: Hamstring (family 4
// bit 1), Arcane Shot (family 9 bit 11), Multi-Shot (family 9 bit 12), and the pet's claw.
var (
	stubHamstring  = core.ActionID{SpellID: 993101}
	stubArcaneShot = core.ActionID{SpellID: 993102}
	stubMultiShot  = core.ActionID{SpellID: 993103}
	stubPetClaw    = core.ActionID{SpellID: 993104}
)

func init() {
	enhancement.RegisterEnhancementShaman()
	priest.RegisterPriest()

	// The warrior stand-in: a rage bar, which the game files' tenths of rage are read against, and a
	// Hamstring costing 10 rage.
	testagent.Register(proto.Player_ProtectionWarrior{}, proto.Spec_SpecProtectionWarrior, func(player *proto.Player, spec any) {
		player.Spec = spec.(*proto.Player_ProtectionWarrior)
	}, func(agent *testagent.Agent) {
		agent.EnableRageBar(core.RageBarOptions{MaxRage: 100, BaseRageMultiplier: 1})
		hamstring := flatHit(stubHamstring, core.SpellSchoolPhysical, core.ProcMaskMeleeMHSpecial, 4, 1)
		hamstring.RageCost = core.RageCostOptions{Cost: 10}
		agent.RegisterSpell(hamstring)
	})

	// The hunter stand-in, on the warlock's spec because the hunter's is taken by the shot stand-in:
	// an Arcane Shot costing 100 mana, a Multi-Shot, and a pet.
	testagent.Register(proto.Player_Warlock{}, proto.Spec_SpecWarlock, func(player *proto.Player, spec any) {
		player.Spec = spec.(*proto.Player_Warlock)
	}, func(agent *testagent.Agent) {
		agent.EnableManaBar()
		arcaneShot := flatHit(stubArcaneShot, core.SpellSchoolArcane, core.ProcMaskRangedSpecial, 9, 11)
		arcaneShot.ManaCost = core.ManaCostOptions{FlatCost: 100}
		agent.RegisterSpell(arcaneShot)
		agent.RegisterSpell(flatHit(stubMultiShot, core.SpellSchoolPhysical, core.ProcMaskRangedSpecial, 9, 12))
		agent.AddPet(&clawPet{Pet: core.NewPet(core.PetConfig{
			Name: "Claw Pet", Owner: &agent.Character, EnabledOnStart: true,
			BaseStats:       stats.Stats{stats.Health: 1000},
			StatInheritance: func(stats.Stats) stats.Stats { return stats.Stats{} },
		})})
	})
}

// A pet whose claw deals 100 physical damage.
type clawPet struct {
	core.Pet
}

func (p *clawPet) GetPet() *core.Pet                        { return &p.Pet }
func (p *clawPet) ApplyTalents()                            {}
func (p *clawPet) Reset(_ *core.Simulation)                 {}
func (p *clawPet) OnEncounterStart(_ *core.Simulation)      {}
func (p *clawPet) ExecuteCustomRotation(_ *core.Simulation) {}

func (p *clawPet) Initialize() {
	p.Pet.Initialize()
	p.RegisterSpell(flatHit(stubPetClaw, core.SpellSchoolPhysical, core.ProcMaskMeleeMHSpecial, 0, 0))
}

// A spell under the family's bit that always hits for 100 no resistance or armor reduces.
func flatHit(id core.ActionID, school core.SpellSchool, mask core.ProcMask, family int32, bit uint) core.SpellConfig {
	config := core.SpellConfig{
		ActionID:         id,
		SpellSchool:      school,
		ProcMask:         mask,
		Flags:            core.SpellFlagIgnoreResists,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeAlwaysHit)
		},
	}
	if family != 0 {
		config.ClassFlags = core.ClassFlags{Family: family, Mask: [4]uint32{1 << bit}}
	}
	return config
}

// The class's stand-in wearing one item, or nothing for item 0, against a level 60 target.
func modWearer(t *testing.T, class proto.Class, slot proto.ItemSlot, itemID int32) (*core.Simulation, *core.Character) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[slot].Id = itemID
	player := &proto.Player{
		Name: "Wearer", Class: class, Race: proto.Race_RaceOrc,
		Equipment:   &proto.EquipmentSpec{Items: items},
		Consumables: &proto.ConsumesSpec{},
		Buffs:       &proto.IndividualBuffs{},
		Rotation:    &proto.APLRotation{},
	}
	if class == proto.Class_ClassWarrior {
		core.WithSpec(player, &proto.Player_ProtectionWarrior{ProtectionWarrior: &proto.ProtectionWarrior{}})
	} else {
		core.WithSpec(player, &proto.Player_Warlock{Warlock: &proto.Warlock{}})
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}},
		Encounter:  &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

func dealtBy(sim *core.Simulation, unit *core.Unit, id core.ActionID) float64 {
	spell := unit.GetSpell(id)
	target := sim.Encounter.ActiveTargetUnits[0]
	before := spell.SpellMetrics[target.UnitIndex].TotalDamage
	spell.SkipCastAndApplyEffects(sim, target)
	return spell.SpellMetrics[target.UnitIndex].TotalDamage - before
}

// The stub hunter with one trinket, against two level 60 targets.
func hunterTrinketWearer(t *testing.T, trinketID int32) (*core.Simulation, *testHunter) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotTrinket2+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotTrinket1].Id = trinketID
	player := &proto.Player{
		Name: "Wearer", Class: proto.Class_ClassHunter, Race: proto.Race_RaceOrc,
		Equipment: &proto.EquipmentSpec{Items: items},
		Buffs:     &proto.IndividualBuffs{},
		Rotation:  &proto.APLRotation{},
	}
	core.WithSpec(player, &proto.Player_Hunter{Hunter: &proto.Hunter{}})

	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{player},
			Buffs:   &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Duration: 180, Targets: []*proto.Target{
			{Name: "target", Level: 60}, {Name: "second", Level: 60}}},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*testHunter)
}

// Casts the shot until one lands.
func landShot(t *testing.T, sim *core.Simulation, hunter *testHunter, shot *core.Spell) {
	t.Helper()
	for casts := 0; ; casts++ {
		if casts > 50 {
			t.Fatalf("50 casts of %v and none landed", shot.ActionID)
		}
		shot.SkipCastAndApplyEffects(sim, hunter.CurrentTarget)
		if hunter.last.Landed() {
			return
		}
	}
}

// The cooldown, category and category cooldown the on-use effect in the item's db entry states.
func classItemOnUse(t *testing.T, itemID int32) (time.Duration, int32, time.Duration) {
	t.Helper()
	for _, effect := range core.GetItemByID(itemID).ItemEffects {
		if onUse := effect.GetOnUse(); onUse != nil {
			return core.DurationFromMillis(onUse.CooldownMs), onUse.CategoryId, core.DurationFromMillis(onUse.CategoryCooldownMs)
		}
	}
	t.Fatalf("%d states no on-use effect", itemID)
	return 0, 0, 0
}

// An Enhancement shaman with Maelstrom Weapon, no weapon and one totem, against a level 63 target.
func enhancementShamanWithTotem(t *testing.T, totemID int32) (*core.Simulation, *shaman.Shaman) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotRanged].Id = totemID
	player := &proto.Player{
		Name: "Shaman", Class: proto.Class_ClassShaman, Race: proto.Race_RaceOrc,
		Equipment:     &proto.EquipmentSpec{Items: items},
		Consumables:   &proto.ConsumesSpec{},
		Buffs:         &proto.IndividualBuffs{},
		Rotation:      &proto.APLRotation{},
		TalentsString: "050533102-053030031005112251",
		Spec: &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{
			Options: &proto.EnhancementShaman_Options{ClassOptions: &proto.ShamanOptions{}}}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}},
		Encounter:  &proto.Encounter{Duration: 25000, Targets: []*proto.Target{{Name: "target", Level: 63}}},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(shaman.ShamanAgent).GetShaman()
}

// An untalented Shadow priest wearing the items, against an Undead target and a Humanoid one.
func priestWearing(t *testing.T, worn map[proto.ItemSlot]int32) (*core.Simulation, *priest.Priest) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{Id: worn[proto.ItemSlot(i)]}
	}
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
		Encounter: &proto.Encounter{Duration: 180, Targets: []*proto.Target{
			{Name: "undead", Level: 60, MobType: proto.MobType_MobTypeUndead},
			{Name: "humanoid", Level: 60, MobType: proto.MobType_MobTypeHumanoid}}},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(priest.PriestAgent).GetPriest()
}

// Finery of Infinite Wisdom's 3-piece bonus 26172 raises Shadow Word: Pain's ticks by the row's
// percentage and leaves Mind Blast alone.
func TestFineryOfInfiniteWisdomRaisesShadowWordPain(t *testing.T) {
	sim, wearer := priestWearing(t, map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotMainHand: 21410,
		proto.ItemSlot_ItemSlotFinger1:  21411,
		proto.ItemSlot_ItemSlotBack:     21412,
	})
	target := sim.Encounter.ActiveTargetUnits[1]
	pain := wearer.GetSpell(core.ActionID{SpellID: priest.ShadowWordPainRankMap.Highest().ID})
	blast := wearer.GetSpell(core.ActionID{SpellID: priest.MindBlastRankMap.Highest().ID})
	bonus := wearer.GetAura("Finery of Infinite Wisdom 3P")
	if pain == nil || blast == nil || bonus == nil || !bonus.IsActive() {
		t.Fatalf("Shadow Word: Pain %v, Mind Blast %v, the 3-piece bonus %v: want all three, the bonus up", pain, blast, bonus)
	}

	painWith, blastWith := pain.ExpectedTickDamage(sim, target), blast.ExpectedInitialDamage(sim, target)
	bonus.Deactivate(sim)
	want := pain.ExpectedTickDamage(sim, target) * (1 + spelldata.MustFind(26172).EffectN(1).Percent())
	if !core.WithinToleranceFloat64(want, painWith, 1e-6) {
		t.Errorf("with the 3-piece bonus, Shadow Word: Pain ticks for %v, want %v", painWith, want)
	}
	if blastWithout := blast.ExpectedInitialDamage(sim, target); !core.WithinToleranceFloat64(blastWithout, blastWith, 1e-6) {
		t.Errorf("with the 3-piece bonus, Mind Blast hits for %v, want %v as without it", blastWith, blastWithout)
	}
}

// Raiments of Conviction's 4-piece bonus 1301087 adds its amount to the priest's spell damage against
// the creature types its effect names, Undead, and nothing against a Humanoid. Three pieces add none.
func TestRaimentsOfConvictionAddSpellDamageAgainstUndead(t *testing.T) {
	versus := func(wearer *priest.Priest, target *core.Unit) float64 {
		return wearer.AttackTables[target.UnitIndex].MobTypeBonusStats[target.MobType][stats.SpellDamage]
	}
	worn := map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotShoulder: 273922,
		proto.ItemSlot_ItemSlotHands:    273921,
		proto.ItemSlot_ItemSlotLegs:     273920,
	}

	sim, three := priestWearing(t, worn)
	if got := versus(three, sim.Encounter.ActiveTargetUnits[0]); got != 0 {
		t.Errorf("with three pieces, the priest has %v spell damage against Undead, want 0", got)
	}

	worn[proto.ItemSlot_ItemSlotFeet] = 273923
	sim, four := priestWearing(t, worn)
	undead, humanoid := sim.Encounter.ActiveTargetUnits[0], sim.Encounter.ActiveTargetUnits[1]
	if got, want := versus(four, undead), spelldata.MustFind(1301087).EffectN(1).BaseValue(); got != want {
		t.Errorf("with four pieces, the priest has %v spell damage against Undead, want %v", got, want)
	}
	if got := versus(four, humanoid); got != 0 {
		t.Errorf("with four pieces, the priest has %v spell damage against a Humanoid, want 0", got)
	}
}

// Arcane Infused Gem 19336 grants Arcane Infused 23721 for the row's duration, on the cooldowns its
// on-use effect states. A landed Aimed Shot leaves it; the next landed Arcane Shot fires Arcane
// Detonation 23722, its whole amount rolled by the row's variance, at every target, and spends it.
func TestArcaneInfusedGemDetonatesOnTheNextArcaneShot(t *testing.T) {
	const gem = 19336
	infused, damage := spelldata.MustFind(23721), spelldata.MustFind(23722).DamageEffect()
	sim, hunter := hunterTrinketWearer(t, gem)
	character := &hunter.Character

	cooldown, category, categoryCooldown := classItemOnUse(t, gem)
	useTrinket(t, sim, character, gem, core.CooldownTypeDPS, cooldown, category, categoryCooldown)
	buff := character.GetAuraByID(core.ActionID{SpellID: infused.ID})
	if buff == nil || !buff.IsActive() || buff.Duration != infused.Duration() {
		t.Fatalf("the gem's buff is %v, want %d up for %v", buff, infused.ID, infused.Duration())
	}
	detonation := character.GetSpell(core.ActionID{SpellID: damage.SpellID})
	if detonation == nil {
		t.Fatalf("the gem registers no Arcane Detonation %d", damage.SpellID)
	}
	results := func(target *core.Unit) int32 {
		m := detonation.SpellMetrics[target.UnitIndex]
		return m.Misses + m.Hits + m.ResistedHits + m.Crits + m.ResistedCrits
	}

	landShot(t, sim, hunter, hunter.aimedShot)
	if !buff.IsActive() || results(character.CurrentTarget) != 0 {
		t.Errorf("a landed Aimed Shot detonated the buff")
	}

	landShot(t, sim, hunter, hunter.arcaneShot)
	if buff.IsActive() {
		t.Errorf("the buff is still up after the Arcane Shot")
	}
	low := damage.Average() * (1 - damage.DirectVariance()/2)
	high := damage.Average() * (1 + damage.DirectVariance()/2)
	for _, target := range sim.Encounter.ActiveTargetUnits {
		if got := results(target); got != 1 {
			t.Errorf("the detonation reached %s %d times, want once", target.Label, got)
		}
		if m := detonation.SpellMetrics[target.UnitIndex]; m.Hits == 1 && (m.TotalDamage < low || m.TotalDamage > high) {
			t.Errorf("the detonation hit %s for %v, want %v to %v", target.Label, m.TotalDamage, low, high)
		}
	}
}

// Renataki's Charm of Beasts 19953 resets the shots' cooldowns, on the cooldowns its on-use effect
// states, and is only worth using while a shot is cooling down.
func TestRenatakisCharmOfBeastsResetsTheShots(t *testing.T) {
	const charm = 19953
	sim, hunter := hunterTrinketWearer(t, charm)
	character := &hunter.Character
	hunter.aimedShot.CD = core.Cooldown{Timer: new(core.Timer), Duration: 6 * time.Second}
	hunter.aimedShot.CD.Reset()

	mcd := character.GetMajorCooldown(core.ActionID{ItemID: charm})
	if mcd == nil || mcd.ShouldActivate(sim, character) {
		t.Fatalf("the charm's major cooldown %v would be used with every shot ready", mcd)
	}

	hunter.aimedShot.CD.Use(sim)
	if !mcd.ShouldActivate(sim, character) {
		t.Errorf("the charm would not be used with Aimed Shot cooling down")
	}
	cooldown, category, categoryCooldown := classItemOnUse(t, charm)
	useTrinket(t, sim, character, charm, core.CooldownTypeDPS, cooldown, category, categoryCooldown)
	if !hunter.aimedShot.IsReady(sim) {
		t.Errorf("Aimed Shot is ready in %v after the charm, want now", hunter.aimedShot.TimeToReady(sim))
	}
}

// The PvP plate gloves' 22778 and Rage of Mugamba's 24428 state a flat modifier on Hamstring's cost
// in tenths of rage, which comes off a warrior's Hamstring as whole rage.
func TestPlateGlovesAndRageOfMugambaCutHamstringsRage(t *testing.T) {
	plate := spelldata.MustFind(22778).EffectN(1).Tenths()
	mugamba := spelldata.MustFind(24428).EffectN(1).Tenths()
	_, bare := modWearer(t, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotHands, 0)
	base := bare.GetSpell(stubHamstring).Cost.GetCurrentCost()

	for _, c := range []struct {
		name string
		slot proto.ItemSlot
		item int32
		cut  float64
	}{
		{"Knight-Lieutenant's Plate Gauntlets", proto.ItemSlot_ItemSlotHands, 16406, plate},
		{"Premier Plate Gauntlets", proto.ItemSlot_ItemSlotHands, 272717, plate},
		{"Rage of Mugamba", proto.ItemSlot_ItemSlotNeck, 19577, mugamba},
	} {
		_, warrior := modWearer(t, proto.Class_ClassWarrior, c.slot, c.item)
		if got, want := warrior.GetSpell(stubHamstring).Cost.GetCurrentCost(), base+c.cut; got != want {
			t.Errorf("wearing %s, Hamstring costs %v rage, want %v", c.name, got, want)
		}
	}
}

// The PvP chain gloves' 23157 states a flat modifier on Arcane Shot's mana cost; the Premier chain
// gloves' 28539 raises Multi-Shot's direct damage by its percentage and leaves Arcane Shot alone.
func TestChainGlovesModifyTheirShot(t *testing.T) {
	bareSim, bare := modWearer(t, proto.Class_ClassHunter, proto.ItemSlot_ItemSlotHands, 0)
	baseCost := bare.GetSpell(stubArcaneShot).Cost.GetCurrentCost()
	baseMultiShot := dealtBy(bareSim, &bare.Unit, stubMultiShot)
	baseArcaneShot := dealtBy(bareSim, &bare.Unit, stubArcaneShot)

	_, cheap := modWearer(t, proto.Class_ClassHunter, proto.ItemSlot_ItemSlotHands, 16403)
	if got, want := cheap.GetSpell(stubArcaneShot).Cost.GetCurrentCost(), baseCost+spelldata.MustFind(23157).EffectN(1).Average(); got != want {
		t.Errorf("wearing Knight-Lieutenant's Chain Gauntlets, Arcane Shot costs %v mana, want %v", got, want)
	}

	sim, hunter := modWearer(t, proto.Class_ClassHunter, proto.ItemSlot_ItemSlotHands, 272495)
	want := baseMultiShot * (1 + spelldata.MustFind(28539).EffectN(1).Percent())
	if got := dealtBy(sim, &hunter.Unit, stubMultiShot); !core.WithinToleranceFloat64(want, got, 1e-6) {
		t.Errorf("wearing Premier Chain Gauntlets, Multi-Shot dealt %v, want %v", got, want)
	}
	if got := dealtBy(sim, &hunter.Unit, stubArcaneShot); !core.WithinToleranceFloat64(baseArcaneShot, got, 1e-6) {
		t.Errorf("wearing Premier Chain Gauntlets, Arcane Shot dealt %v, want %v", got, baseArcaneShot)
	}
}

// Ferra's Collar's 22854 adds its amount to Tamed Pet Passive's damage done, which is 0 on its own:
// the pet deals that many percent more, less than the tooltip states.
func TestFerrasCollarRaisesThePetsDamage(t *testing.T) {
	bareSim, bare := modWearer(t, proto.Class_ClassHunter, proto.ItemSlot_ItemSlotTrinket1, 0)
	base := dealtBy(bareSim, &bare.Pets[0].Unit, stubPetClaw)

	sim, hunter := modWearer(t, proto.Class_ClassHunter, proto.ItemSlot_ItemSlotTrinket1, 18355)
	want := base * (1 + spelldata.MustFind(22854).EffectN(1).Percent())
	if got := dealtBy(sim, &hunter.Pets[0].Unit, stubPetClaw); !core.WithinToleranceFloat64(want, got, 1e-6) {
		t.Errorf("wearing Ferra's Collar, the pet's claw dealt %v, want %v", got, want)
	}
}

// Burning Totem 272433's equip spell 1291077 states a flat duration on Flame Shock, family 11 mask
// 0x10000000: every rank's dot lasts that much longer.
func TestBurningTotemExtendsFlameShock(t *testing.T) {
	_, without := enhancementShamanWithTotem(t, 0)
	_, with := enhancementShamanWithTotem(t, 272433)
	if len(without.FlameShock) == 0 {
		t.Fatal("the shaman registers no Flame Shock")
	}

	extension := spelldata.MustFind(1291077).EffectN(1).TimeValue()
	for i, spell := range with.FlameShock {
		base := without.FlameShock[i].Dot(without.CurrentTarget).BaseDuration()
		if got, want := spell.Dot(with.CurrentTarget).BaseDuration(), base+extension; got != want {
			t.Errorf("Flame Shock %v lasts %v with the totem, want %v", spell.ActionID, got, want)
		}
	}
}

// Totem of the Storm 272432 lets a landed Lightning Bolt grant a Maelstrom Weapon stack at the share
// of the talent's rate per minute its dummy 1291078 states, priced on the bolt's cast time. Chain
// Lightning and the copy Lightning Overload casts never do.
func TestTotemOfTheStormLetsLightningBoltGrantMaelstromWeapon(t *testing.T) {
	sim, wearer := enhancementShamanWithTotem(t, 272432)
	if wearer.Talents.MaelstromWeapon == 0 {
		t.Fatal("the shaman has no Maelstrom Weapon")
	}
	wearer.AutoAttacks.CancelAutoSwing(sim)
	target := wearer.CurrentTarget

	maelstrom := wearer.GetAuraByID(core.ActionID{SpellID: 408505})
	if maelstrom == nil {
		t.Fatal("no Maelstrom Weapon buff 408505 on the shaman")
	}
	var bolt, chain, overload *core.Spell
	for _, spell := range wearer.Spellbook {
		switch {
		case spell.Matches(shaman.SpellMaskLightningBolt):
			bolt = spell
		case spell.Matches(shaman.SpellMaskChainLightning):
			chain = spell
		case spell.Matches(shaman.SpellMaskLightningBoltOverload):
			overload = spell
		}
	}

	// Counts the times n uses left the buff up.
	gains := func(n int, use func()) int {
		count := 0
		for range n {
			maelstrom.Deactivate(sim)
			use()
			if maelstrom.IsActive() {
				count++
			}
		}
		return count
	}
	cast := func(spell *core.Spell) func() {
		return func() {
			spell.SkipCastAndApplyEffects(sim, target)
			stepTo(t, sim, sim.CurrentTime+time.Second)
		}
	}

	if got := gains(200, cast(chain)) + gains(200, cast(overload)); got != 0 {
		t.Errorf("Chain Lightning and the overload granted Maelstrom Weapon %d times, want never", got)
	}

	// The talent's own rate per minute, measured on main-hand hits at the main hand's speed.
	const swings = 60000
	swung := gains(swings, func() { dealHit(t, sim, &wearer.Character, meleeSwing(&wearer.Unit)) })
	talentPPM := float64(swung) / swings / (wearer.AutoAttacks.MH().SwingSpeed / 60)

	// Against a level 63 target about one bolt in six misses, so a miss that rolled would show.
	granted := gains(20000, cast(bolt))
	m := bolt.SpellMetrics[target.UnitIndex]
	landed := m.Hits + m.Crits
	want := talentPPM * spelldata.MustFind(1291078).EffectN(1).Percent() * bolt.DefaultCast.CastTime.Minutes()
	if got := float64(granted) / float64(landed); math.Abs(got-want) > 0.015 {
		t.Errorf("%d of %d landed Lightning Bolts (%v cast) granted a stack, %.3f a bolt, want %.3f",
			granted, landed, bolt.DefaultCast.CastTime, got, want)
	}
}

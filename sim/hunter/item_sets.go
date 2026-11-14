package hunter

import (
	"slices"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var (
	cryptstalkerRapidFire      = spelldata.MustFind(28755) // 2pc: Rapid Fire lasts 4 s longer
	cryptstalkerStalkersAlly   = spelldata.MustFind(28758) // 4pc: Stalker's Ally, the attack power on the hunter and the pet
	cryptstalkerAdrenalineProc = spelldata.MustFind(28752) // 6pc: Adrenaline Rush, the proc
	cryptstalkerAdrenalineMana = spelldata.MustFind(28753) // 6pc: Adrenaline Rush, the mana it grants
	cryptstalkerShotCost       = spelldata.MustFind(28751) // 8pc: Multi-Shot and Aimed Shot cost 20 less
)

// Tier 1
// https://www.wowhead.com/forever/item-set=2101/wildstalker-armor
//
// The game files ship four of the six pieces (shoulders, hands, legs, feet), so the 5-piece bonus
// waits on the head and chest the ItemSet row lists. The 3-piece bonus takes 3 sec off Frost Trap and
// Freezing Trap, which the sim does not cast. The 5-piece's mask names Aimed Shot alone; Multi-Shot
// shares its cooldown category.
var ItemSetWildstalkerArmor = core.NewItemSet(core.ItemSet{
	ID:   2101,
	Name: "Wildstalker Armor",
	Slots: []proto.ItemSlot{
		proto.ItemSlot_ItemSlotHead,
		proto.ItemSlot_ItemSlotShoulder,
		proto.ItemSlot_ItemSlotChest,
		proto.ItemSlot_ItemSlotHands,
		proto.ItemSlot_ItemSlotLegs,
		proto.ItemSlot_ItemSlotFeet,
	},
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(spelldata.MustFind(1300945)), // +1% attack and casting speed
		3: shared.ParsedSetBonus(spelldata.MustFind(1301012)), // Frost and Freezing Trap cooldowns -3 sec
		4: shared.ParsedSetBonus(spelldata.MustFind(1301078)), // +36 melee attack power against Beasts
		5: shared.ParsedSetBonus(spelldata.MustFind(1301252)), // Aimed Shot cooldown -1 sec
	},
})

// Ruins of Ahn'Qiraj
// https://www.wowhead.com/forever/item-set=509/strikers-garb
var ItemSetStrikersGarb = core.NewItemSet(core.ItemSet{
	ID:   509,
	Name: "Striker's Garb",
	Slots: []proto.ItemSlot{
		proto.ItemSlot_ItemSlotHead,
		proto.ItemSlot_ItemSlotShoulder,
		proto.ItemSlot_ItemSlotChest,
		proto.ItemSlot_ItemSlotLegs,
		proto.ItemSlot_ItemSlotFeet,
	},
	Bonuses: map[int32]core.ApplySetBonus{
		3: shared.ParsedSetBonus(spelldata.MustFind(26173)), // Arcane Shot costs 10% less mana
		5: shared.ParsedSetBonus(spelldata.MustFind(26174)), // Rapid Fire cooldown -2 min
	},
})

// Tier 3
var ItemSetCryptstalkerArmor = core.NewItemSet(core.ItemSet{
	Name:  "Cryptstalker Armor",
	ID:    530,
	Slots: shared.TierThreeSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(cryptstalkerRapidFire),
		4: applyStalkersAlly,
		6: func(agent core.Agent, setBonusAura *core.Aura) {
			character := agent.GetCharacter()
			manaMetrics := character.NewManaMetrics(core.ActionID{SpellID: cryptstalkerAdrenalineMana.ID})
			mana := cryptstalkerAdrenalineMana.EnergizeEffect().Average()

			// 28752 states "always" on ranged hits and its proc effect names Multi-Shot alone; the
			// tooltip says your ranged critical hits, and the tooltip wins: the crit is the outcome
			// and the mask comes off.
			trigger := spelldata.ProcTrigger(character, cryptstalkerAdrenalineProc,
				func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
					character.AddMana(sim, mana, manaMetrics)
				})
			trigger.Name = "Cryptstalker Armor - 6PC"
			trigger.Outcome = core.OutcomeCrit
			trigger.ClassFlags = core.ClassFlags{}

			setBonusAura.AttachProcTrigger(trigger)
		},
		8: shared.ParsedSetBonus(cryptstalkerShotCost),
	},
})

// The 4-piece's equip spell 28756 casts 28757 on the pet every 3 s, and 28758, a 4 s pet area aura,
// gives the hunter and the pet its attack power and ranged attack power. With no pet there is nothing
// to cast on, and the 3 s recast keeps the 4 s aura from ever falling off, so both carry it exactly
// while the pet is out. The pet's amount sits on its own aura, so inheritance never copies it; the
// hunter's amount reaches the pet inside the 10% of the hunter's attack power it inherits.
// ParseEffects reads no pet area aura, so the stats attach by hand. The pet is the wearer's one pet
// that is no guardian.
func applyStalkersAlly(agent core.Agent, setBonusAura *core.Aura) {
	character := agent.GetCharacter()
	bonus := stats.Stats{
		stats.AttackPower:       cryptstalkerStalkersAlly.Effect(dbcenums.A_MOD_ATTACK_POWER, 0).Average(),
		stats.RangedAttackPower: cryptstalkerStalkersAlly.Effect(dbcenums.A_MOD_RANGED_ATTACK_POWER, 0).Average(),
	}

	petIndex := slices.IndexFunc(character.Pets, func(pet *core.Pet) bool { return !pet.IsGuardian() })
	if petIndex < 0 {
		return
	}
	pet := character.Pets[petIndex]

	register := func(unit *core.Unit) *core.Aura {
		config := spelldata.AuraConfig(cryptstalkerStalkersAlly)
		config.Duration = core.NeverExpires
		return unit.RegisterAura(config).AttachStatsBuff(bonus)
	}
	ownerAura, petAura := register(&character.Unit), register(&pet.Unit)
	activate := func(sim *core.Simulation) {
		ownerAura.Activate(sim)
		petAura.Activate(sim)
	}
	deactivate := func(sim *core.Simulation) {
		ownerAura.Deactivate(sim)
		petAura.Deactivate(sim)
	}

	setBonusAura.ApplyOnGain(func(_ *core.Aura, sim *core.Simulation) {
		if pet.IsEnabled() {
			activate(sim)
		}
	})
	setBonusAura.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
		deactivate(sim)
	})
	pet.RegisterEnableEffect(func(sim *core.Simulation) {
		if setBonusAura.IsActive() {
			activate(sim)
		}
	})
	pet.RegisterDisableEffect(deactivate)
}

var pvpGloveItemIDs = []int32{23279, 22862, 16463, 16571}

// The gloves' equip spell: Multi-Shot deals 4% more.
var pvpGloveMultiShot = spelldata.MustFind(28539)

func init() {
	for _, itemID := range pvpGloveItemIDs {
		core.NewItemEffect(itemID, func(_ core.Agent) {})
	}
}

func (hunter *Hunter) addPvpGloves() {
	hunter.RegisterPvPGloveMod(
		pvpGloveItemIDs,
		core.SpellModConfig{
			Kind:       core.SpellMod_DirectDamageDone_Pct,
			ClassFlags: multiShotRank.ClassFlags,
			FloatValue: pvpGloveMultiShot.EffectN(1).Percent(),
		})
}

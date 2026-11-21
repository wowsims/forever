package paladin

import (
	"time"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The six slots of a Justice set. The game files ship the sets' current pieces under encryption, so
// the slots follow the older pieces the game files have seen: shoulders, hands, legs and feet, with
// head and chest for the two the ItemSet row still lists.
var justiceSetSlots = []proto.ItemSlot{
	proto.ItemSlot_ItemSlotHead,
	proto.ItemSlot_ItemSlotShoulder,
	proto.ItemSlot_ItemSlotChest,
	proto.ItemSlot_ItemSlotHands,
	proto.ItemSlot_ItemSlotLegs,
	proto.ItemSlot_ItemSlotFeet,
}

// Tier 1 - Retribution
// https://www.wowhead.com/forever/item-set=2106/justice-battlegear
//
// The 3-piece bonus takes 5 sec off Hammer of Justice's cooldown, which the sim does not cast.
var ItemSetJusticeBattlegear = core.NewItemSet(core.ItemSet{
	ID:    2106,
	Name:  "Justice Battlegear",
	Slots: justiceSetSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(_ core.Agent, setBonusAura *core.Aura) {
			// Increases your attack speed and casting speed by 1%.
			setBonusAura.
				AttachMultiplyAttackSpeed(1.01).
				AttachMultiplyCastSpeed(1.01).
				ExposeToAPL(1300951)
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			// +36 Attack Power against Undead. The game files' text says Demons; the effect's
			// creature-type mask is Undead.
			spelldata.ParseEffects(agent.GetCharacter(), setBonusAura, spelldata.MustFind(1301084))
			setBonusAura.ExposeToAPL(1301084)
		},
		5: func(_ core.Agent, setBonusAura *core.Aura) {
			// Reduces the cooldown on your Judgement spell by 0.5 sec.
			setBonusAura.
				AttachSpellMod(core.SpellModConfig{
					ClassMask: SpellMaskJudgement,
					Kind:      core.SpellMod_Cooldown_Flat,
					TimeValue: -500 * time.Millisecond,
				}).
				ExposeToAPL(1301702)
		},
	},
})

// Tier 1 - Holy
// https://www.wowhead.com/forever/item-set=2107/justice-armor
//
// The 3-piece bonus takes 30 sec off Blessing of Protection's cooldown, which the sim does not cast.
var ItemSetJusticeArmor = core.NewItemSet(core.ItemSet{
	ID:    2107,
	Name:  "Justice Armor",
	Slots: justiceSetSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(_ core.Agent, setBonusAura *core.Aura) {
			// Increased Spirit +10.
			setBonusAura.
				AttachStatBuff(stats.Spirit, 10).
				ExposeToAPL(1300948)
		},
		4: func(_ core.Agent, setBonusAura *core.Aura) {
			// Increases healing done by up to 26 and damage done by up to 9 for all magical spells
			// and effects.
			setBonusAura.
				AttachStatsBuff(stats.Stats{stats.HealingPower: 26, stats.SpellDamage: 9}).
				ExposeToAPL(1301080)
		},
		5: func(_ core.Agent, setBonusAura *core.Aura) {
			// Reduces the cooldown on your Holy Shock spell by 1 sec.
			setBonusAura.
				AttachSpellMod(core.SpellModConfig{
					ClassMask: SpellMaskHolyShock | SpellMaskHolyShockHeal,
					Kind:      core.SpellMod_Cooldown_Flat,
					TimeValue: -time.Second,
				}).
				ExposeToAPL(1301694)
		},
	},
})

// Tier 1 - Protection
// https://www.wowhead.com/forever/item-set=2108/justice-battleplate
//
// The 3-piece bonus takes 1.5 sec off Turn Undead's cast time, which the sim does not cast.
var ItemSetJusticeBattleplate = core.NewItemSet(core.ItemSet{
	ID:    2108,
	Name:  "Justice Battleplate",
	Slots: justiceSetSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(_ core.Agent, setBonusAura *core.Aura) {
			// Increased Defense +7.
			setBonusAura.
				AttachStatBuff(stats.DefenseRating, 7*core.DefenseRatingPerDefenseLevel).
				ExposeToAPL(1300949)
		},
		4: func(_ core.Agent, setBonusAura *core.Aura) {
			// Reduces the chance for your melee attacks to be Dodged or Parried by 1.2%: 12
			// expertise rating.
			setBonusAura.
				AttachStatBuff(stats.ExpertiseRating, 12).
				ExposeToAPL(1301083)
		},
		5: func(agent core.Agent, setBonusAura *core.Aura) {
			// Reduces the duration of Forbearance any time you gain it by 10 sec.
			agent.(PaladinAgent).GetPaladin().forbearanceReduction += time.Second * 10
			setBonusAura.ExposeToAPL(1301697)
		},
	},
})

// Ruins of Ahn'Qiraj
// https://www.wowhead.com/forever/item-set=506/battlegear-of-eternal-justice
var ItemSetBattlegearOfEternalJustice = core.NewItemSet(core.ItemSet{
	ID:   506,
	Name: "Battlegear of Eternal Justice",
	Slots: []proto.ItemSlot{
		proto.ItemSlot_ItemSlotBack,
		proto.ItemSlot_ItemSlotFinger1,
		proto.ItemSlot_ItemSlotFinger2,
		proto.ItemSlot_ItemSlotMainHand,
	},
	Bonuses: map[int32]core.ApplySetBonus{
		3: func(agent core.Agent, setBonusAura *core.Aura) {
			// 20% chance to regain 100 mana when you cast a Judgement.
			//
			// 26135 hears every harmful spell cast and names none of the paladin's spells, so the
			// listener keeps to Judgement's own cast. The judgement it unleashes rolls nothing.
			// TODO: Ingame test needed -- whether a Swift Judgement cast rolls it too.
			character := agent.GetCharacter()
			row := spelldata.MustFind(26135)
			trigger := spelldata.ProcTrigger(character, row, nil)
			trigger.Handler = shared.FiredSpellHandler(character, trigger.ActionID, row)
			trigger.Name = "Battlegear of Eternal Justice - 3PC"
			// Judgement's cast carries no proc mask.
			trigger.ProcMask = core.ProcMaskUnknown
			trigger.ClassSpellMask = SpellMaskJudgement

			setBonusAura.AttachProcTrigger(trigger)
			setBonusAura.ExposeToAPL(row.ID)
		},
	},
})

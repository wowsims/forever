package warrior

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The rows ItemSetSpell names for each set's thresholds, and the buffs those rows grant. None of
// them is a warrior class spell, so the generated table carries no reference to them.
var (
	mightBlockValue   = spelldata.MustFind(23562) // 3pc: Block Value 30
	mightRageProc     = spelldata.MustFind(21838) // 5pc: Battlegear of Might, the proc
	mightRageEnergize = spelldata.MustFind(29478) // 5pc: Battlegear of Might, the rage it grants
	mightSunderThreat = spelldata.MustFind(23561) // 8pc: Enhanced Sunder Armor
)

// TODO: items 16861-16868 have no ItemSparse rows in the game files, so the set cannot be worn yet.
var ItemSetBattlegearOfMight = core.NewItemSet(core.ItemSet{
	Name: "Battlegear of Might",
	ID:   209,
	Bonuses: map[int32]core.ApplySetBonus{
		3: func(agent core.Agent, setBonusAura *core.Aura) {
			// A_MOD_BLOCK_VALUE_FLAT has no sim kind in the parse table, so only the amount comes
			// off the row.
			setBonusAura.AttachStatBuff(stats.BlockValue, mightBlockValue.EffectN(1).BaseValue())
		},
		5: func(agent core.Agent, setBonusAura *core.Aura) {
			warrior := agent.(WarriorAgent).GetWarrior()
			gain := shared.FiredSpellHandler(&warrior.Character, core.ActionID{SpellID: mightRageEnergize.ID}, mightRageProc)

			// The rate is the proc chance column, which the tooltip's $h% says is a real roll.
			trigger := spelldata.ProcTrigger(&warrior.Character, mightRageProc, gain)
			trigger.Name = "Battlegear of Might - 5PC"

			setBonusAura.AttachProcTrigger(trigger)
		},
		8: func(agent core.Agent, setBonusAura *core.Aura) {
			// The row states A_ADD_PCT_MODIFIER on the threat, which the parse maps to
			// SpellMod_ThreatMultiplier_Pct: that scales the threat a spell's damage makes, and
			// Sunder Armor deals none. All of its threat is the flat bonus and its attack power
			// share, which SpellMod_FlatThreatBonus_Pct scales together, so only the amount and the
			// spells it names come off the row.
			setBonusAura.AttachSpellMod(core.SpellModConfig{
				ClassFlags: mightSunderThreat.EffectN(1).ClassFlags,
				Kind:       core.SpellMod_FlatThreatBonus_Pct,
				FloatValue: mightSunderThreat.EffectN(1).Percent(),
			})
		},
	},
})

var (
	wrathDiscountProc = spelldata.MustFind(21890) // 5pc: Warrior's Wrath, the proc
	wrathDiscountBuff = spelldata.MustFind(21887) // 5pc: Warrior's Wrath, the discount
	wrathParryProc    = spelldata.MustFind(23548) // 8pc: Parry, the proc
	wrathParryBuff    = spelldata.MustFind(23547) // 8pc: Parry, the buff
)

// TODO: items 16959-16966 have no ItemSparse rows in the game files, so the set cannot be worn yet.
var ItemSetBattlegearOfWrath = core.NewItemSet(core.ItemSet{
	Name: "Battlegear of Wrath",
	ID:   218,
	Bonuses: map[int32]core.ApplySetBonus{
		3: func(agent core.Agent, setBonusAura *core.Aura) {
			// Spell 23563 states 30 attack power on Battle Shout, through the flag the HasBsT2 option sets.
			// The shout reads the flag once, when it is registered, so only a set worn at the start counts.
			if setBonusAura.BuildPhase == core.CharacterBuildPhaseGear {
				agent.(WarriorAgent).GetWarrior().HasBsT2 = true
			}
		},
		5: func(agent core.Agent, setBonusAura *core.Aura) {
			warrior := agent.(WarriorAgent).GetWarrior()

			// The abilities the discount names, which are also the ones the tooltip fires the
			// proc off: the buff's own modifier effect states them and 21890's proc effect
			// states none.
			discounted := wrathDiscountBuff.EffectN(1).ClassFlags

			buff := warrior.RegisterAura(spelldata.AuraConfig(wrathDiscountBuff))
			spelldata.ParseEffects(&warrior.Character, buff, wrathDiscountBuff)

			// The discount is spent at cast, so the buff goes when an ability is cast rather than
			// when one lands, which is what the row's proc flags state.
			buff.AttachProcTrigger(core.ProcTrigger{
				Name:               "Warrior's Wrath - Consume",
				ClassFlags:         discounted,
				Callback:           core.CallbackOnCastComplete,
				TriggerImmediately: true,
				Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
					buff.Deactivate(sim)
				},
			})

			// The rate is the proc chance column, which the tooltip's $h% says is a real roll.
			trigger := spelldata.ProcTrigger(&warrior.Character, wrathDiscountProc,
				func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
					buff.Activate(sim)
				})
			trigger.Name = "Battlegear of Wrath - 5PC"
			trigger.ClassFlags = discounted

			setBonusAura.AttachProcTrigger(trigger)
		},
		8: func(agent core.Agent, setBonusAura *core.Aura) {
			warrior := agent.(WarriorAgent).GetWarrior()

			parry := warrior.RegisterAura(spelldata.AuraConfig(wrathParryBuff,
				spelldata.Label("Battlegear of Wrath Parry")))
			spelldata.ParseEffects(&warrior.Character, parry, wrathParryBuff)

			// The attack that spends the buff is the one it parries, which is an outcome no proc
			// mask states and one that deals no damage. 23547 ships no tooltip, so there is nothing
			// for the row to have read that from and both of its defaults go here.
			consume := spelldata.ProcTrigger(&warrior.Character, wrathParryBuff,
				func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
					parry.Deactivate(sim)
				})
			consume.Name = "Battlegear of Wrath - 8PC Consume"
			consume.Outcome = core.OutcomeEmpty
			consume.RequireDamageDealt = false
			consume.TriggerImmediately = true
			parry.AttachProcTrigger(consume)

			// The rate is the proc chance column, which the tooltip's $h% says is a real roll; the
			// block it fires on is an outcome no proc mask states, so the row states only that the
			// hit it hears carries no damage.
			trigger := spelldata.ProcTrigger(&warrior.Character, wrathParryProc,
				func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
					parry.Activate(sim)
				})
			trigger.Name = "Battlegear of Wrath - 8PC"
			trigger.Outcome = core.OutcomeBlock

			setBonusAura.AttachProcTrigger(trigger)
		},
	},
})

var (
	conquerorShoutCost   = spelldata.MustFind(26109) // 3pc: Conqueror Shout Bonus
	conquerorThunderClap = spelldata.MustFind(26110) // 5pc: Conqueror Thunder Clap Bonus
)

var ItemSetConquerorsBattlegear = core.NewItemSet(core.ItemSet{
	Name: "Conqueror's Battlegear",
	ID:   496,
	Bonuses: map[int32]core.ApplySetBonus{
		3: func(agent core.Agent, setBonusAura *core.Aura) {
			warrior := agent.(WarriorAgent).GetWarrior()
			spelldata.ParseEffects(&warrior.Character, setBonusAura, conquerorShoutCost)
		},
		5: func(agent core.Agent, setBonusAura *core.Aura) {
			warrior := agent.(WarriorAgent).GetWarrior()
			spelldata.ParseEffects(&warrior.Character, setBonusAura, conquerorThunderClap)

			// The one amount the row states raises the slow the tooltip names alongside the
			// damage, which thunder_clap.go reads.
			setBonusAura.AttachAdditivePseudoStatBuff(&warrior.thunderClapEffectBonus, conquerorThunderClap.EffectN(1).Percent())
		},
	},
})

// The six pieces the Tier 1 ItemSet rows list.
// TODO: only the shoulders, hands, legs and feet have ItemSparse rows in the game files, so the
// 5-piece bonuses cannot be worn yet.
var tierOneSlots = []proto.ItemSlot{
	proto.ItemSlot_ItemSlotHead,
	proto.ItemSlot_ItemSlotShoulder,
	proto.ItemSlot_ItemSlotChest,
	proto.ItemSlot_ItemSlotHands,
	proto.ItemSlot_ItemSlotLegs,
	proto.ItemSlot_ItemSlotFeet,
}

// Every aura the Tier 1 rows state has a row in the parse table, so each bonus is its row parsed.
func parsedSetBonus(row *spelldata.Spell) core.ApplySetBonus {
	return func(agent core.Agent, setBonusAura *core.Aura) {
		spelldata.ParseEffects(agent.GetCharacter(), setBonusAura, row)
	}
}

var (
	battlegearGloryHaste          = spelldata.MustFind(1300968) // 2pc: attack speed, and cast speed, which no physical ability takes
	battlegearGloryShoutCooldown  = spelldata.MustFind(1301069) // 3pc: Intimidating Shout's cooldown
	battlegearGloryHumanoidPower  = spelldata.MustFind(1301095) // 4pc: attack power against Humanoids
	battlegearGloryRecklessnessCD = spelldata.MustFind(1301721) // 5pc: Recklessness's cooldown
)

// Tier 1 - Arms/Fury
var ItemSetBattlegearOfGlory = core.NewItemSet(core.ItemSet{
	Name:  "Battlegear of Glory",
	ID:    2102,
	Slots: tierOneSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(battlegearGloryHaste),
		3: parsedSetBonus(battlegearGloryShoutCooldown),
		4: parsedSetBonus(battlegearGloryHumanoidPower),
		5: parsedSetBonus(battlegearGloryRecklessnessCD),
	},
})

var (
	battleplateGloryDefense      = spelldata.MustFind(1300970) // 2pc: Defense skill
	battleplateGloryIntervene    = spelldata.MustFind(1301071) // 3pc: Intervene's cooldown
	battleplateGloryExpertise    = spelldata.MustFind(1301096) // 4pc: expertise rating
	battleplateGloryShieldWallCD = spelldata.MustFind(1301722) // 5pc: Shield Wall's cooldown
)

// Tier 1 - Protection
var ItemSetBattleplateOfGlory = core.NewItemSet(core.ItemSet{
	Name:  "Battleplate of Glory",
	ID:    2103,
	Slots: tierOneSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(battleplateGloryDefense),
		// TODO: the sim does not cast Intervene, so the cooldown the row takes off it reaches no spell.
		3: parsedSetBonus(battleplateGloryIntervene),
		4: parsedSetBonus(battleplateGloryExpertise),
		5: parsedSetBonus(battleplateGloryShieldWallCD),
	},
})

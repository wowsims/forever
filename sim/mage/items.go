package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The rows ItemSetSpell names for each set's thresholds, and the buffs those rows grant. None of
// them is a mage class spell, so the generated table carries no reference to them.
var (
	illusionistSpellDamage = spelldata.MustFind(9417)  // 2pc: +12 spell damage
	illusionistFlamestrike = spelldata.MustFind(24491) // 5pc: Flamestrike casts 0.5 sec faster
)

// (3) Set: Arcane Intellect and Arcane Brilliance cost less, and the sim casts neither.
var ItemSetIllusionistsAttire = core.NewItemSet(core.ItemSet{
	Name: "Illusionist's Attire",
	ID:   482,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			mage := agent.(MageAgent).GetMage()
			spelldata.ParseEffects(&mage.Character, setBonusAura, illusionistSpellDamage)
		},
		5: func(agent core.Agent, setBonusAura *core.Aura) {
			mage := agent.(MageAgent).GetMage()
			spelldata.ParseEffects(&mage.Character, setBonusAura, illusionistFlamestrike)
		},
	},
})

var (
	enigmaResistProc = spelldata.MustFind(26128) // 5pc: Enigma Resist Bonus, the proc
	enigmaAnswer     = spelldata.MustFind(26129) // 5pc: Enigma's Answer, the spell hit buff
)

// (3) Set: a chance for Blizzard not to be interrupted, which the sim never does.
var ItemSetEnigmaVestments = core.NewItemSet(core.ItemSet{
	Name: "Enigma Vestments",
	ID:   503,
	Bonuses: map[int32]core.ApplySetBonus{
		5: func(agent core.Agent, setBonusAura *core.Aura) {
			mage := agent.(MageAgent).GetMage()

			answer := mage.RegisterAura(spelldata.AuraConfig(enigmaAnswer))
			spelldata.ParseEffects(&mage.Character, answer, enigmaAnswer)

			// The row fires on every harmful spell, which decodes to the cast; its tooltip names
			// a resisted spell, an outcome the mask has no bit for and one only the hit knows. A
			// resisted spell deals nothing, so the listener does not ask for damage.
			trigger := spelldata.ProcTrigger(&mage.Character, enigmaResistProc,
				func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
					answer.Activate(sim)
					answer.AddStack(sim)
				})
			trigger.Name = "Enigma Vestments - 5PC"
			trigger.Callback = core.CallbackOnSpellHitDealt
			trigger.Outcome = core.OutcomeMiss
			trigger.RequireDamageDealt = false

			setBonusAura.AttachProcTrigger(trigger)
		},
	},
})

var (
	arcanistSpellDamage = spelldata.MustFind(9346)  // 3pc: +18 spell damage
	arcanistPiercing    = spelldata.MustFind(26175) // 5pc: target resistances 10 lower
	arcanistThreat      = spelldata.MustFind(23545) // 8pc: 15% less threat
)

// TODO: items 16795-16802 have no ItemSparse rows in the game files, so the set cannot be worn yet.
var ItemSetArcanistRegalia = core.NewItemSet(core.ItemSet{
	Name: "Arcanist Regalia",
	ID:   201,
	Bonuses: map[int32]core.ApplySetBonus{
		3: func(agent core.Agent, setBonusAura *core.Aura) {
			mage := agent.(MageAgent).GetMage()
			spelldata.ParseEffects(&mage.Character, setBonusAura, arcanistSpellDamage)
		},
		5: func(agent core.Agent, setBonusAura *core.Aura) {
			mage := agent.(MageAgent).GetMage()
			spelldata.ParseEffects(&mage.Character, setBonusAura, arcanistPiercing)
		},
		8: func(agent core.Agent, setBonusAura *core.Aura) {
			mage := agent.(MageAgent).GetMage()
			spelldata.ParseEffects(&mage.Character, setBonusAura, arcanistThreat)
		},
	},
})

var (
	netherwindFocusProc = spelldata.MustFind(22007) // 8pc: Netherwind Focus, the proc
	netherwindFocusBuff = spelldata.MustFind(22008) // 8pc: Netherwind Focus, the instant cast
)

// TODO: items 16818 and 16912-16918 have no ItemSparse rows in the game files, so the set cannot be
// worn yet.
//
// (5) Set: a larger radius on Arcane Explosion, Flamestrike and Blizzard, which a single target does
// not test.
var ItemSetNetherwindRegalia = core.NewItemSet(core.ItemSet{
	Name: "Netherwind Regalia",
	ID:   210,
	Bonuses: map[int32]core.ApplySetBonus{
		// TODO: (3) Set: spell 21879 states its threat cut on Scorch, Arcane Missiles, Fireball and
		// Frostbolt as flat SPELLMOD_THREAT amounts (-10000, -2000), which have no sim kind.
		8: func(agent core.Agent, setBonusAura *core.Aura) {
			mage := agent.(MageAgent).GetMage()

			// The spells the buff makes instant, which are also the ones whose cast spends it.
			instant := netherwindFocusBuff.EffectN(1).ClassFlags

			buff := mage.RegisterAura(spelldata.AuraConfig(netherwindFocusBuff))
			spelldata.ParseEffects(&mage.Character, buff, netherwindFocusBuff)

			buff.AttachProcTrigger(core.ProcTrigger{
				Name:               "Netherwind Focus - Consume",
				ClassFlags:         instant,
				Callback:           core.CallbackOnCastComplete,
				TriggerImmediately: true,
				Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
					buff.Deactivate(sim)
				},
			})

			// The rate is effect 1's chance. The proc effect states no class flags, so the three
			// spells the tooltip names are matched by rank.
			// TODO: the row states a 10 s cooldown beside no proc cooldown; whether it gates the
			// proc waits on a log.
			trigger := spelldata.ProcTrigger(&mage.Character, netherwindFocusProc,
				func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
					buff.Activate(sim)
				})
			trigger.Name = "Netherwind Regalia - 8PC"
			trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
				return spellData.ArcaneMissiles.Contains(spell) || spellData.Fireball.Contains(spell) ||
					spellData.Frostbolt.Contains(spell)
			}

			setBonusAura.AttachProcTrigger(trigger)
		},
	},
})

var (
	zanzilSpellHit    = spelldata.MustFind(23727) // 2pc: +1% spell hit
	zanzilSpellDamage = spelldata.MustFind(9396)  // 2pc: +6 spell damage
)

// TODO: items 19893 and 19905 have no ItemSparse rows in the game files, so the set cannot be worn yet.
var ItemSetZanzilsConcentration = core.NewItemSet(core.ItemSet{
	Name: "Zanzil's Concentration",
	ID:   462,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			mage := agent.(MageAgent).GetMage()
			spelldata.ParseEffects(&mage.Character, setBonusAura, zanzilSpellHit)
			spelldata.ParseEffects(&mage.Character, setBonusAura, zanzilSpellDamage)
		},
	},
})

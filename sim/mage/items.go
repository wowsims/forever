package mage

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The rows ItemSetSpell names for each set's thresholds, and the buffs and debuffs those rows grant.
// None of them is a mage class spell, so the generated table carries no reference to them.
var (
	frostfireEvocation     = spelldata.MustFind(28763) // 2pc: Evocation cooldown 60 s shorter
	frostfireVulnerability = spelldata.MustFind(28771) // 6pc: Elemental Vulnerability, the proc
	frostfireVulnDebuff    = spelldata.MustFind(28772) // 6pc: Elemental Vulnerability, the debuff
	frostfireNotThere      = spelldata.MustFind(28761) // 8pc: Not There, the proc
	frostfireNotThereBuff  = spelldata.MustFind(28762) // 8pc: Not There, the threat buff
)

// Tier 3
//
// (4) Set: Adaptive Warding raises a resistance when Mage Armor is struck by a spell, which a boss
// fight does not test.
var ItemSetFrostfireRegalia = core.NewItemSet(core.ItemSet{
	Name:  "Frostfire Regalia",
	ID:    526,
	Slots: shared.TierThreeSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(frostfireEvocation),
		6: func(agent core.Agent, setBonusAura *core.Aura) {
			character := agent.GetCharacter()

			// A_MOD_DAMAGE_TAKEN rather than a from-caster aura: every caster's spells take it. "Up to
			// $s1" is spell damage against the target, which each spell takes at its own coefficient,
			// as Judgement of the Crusader does, so it is not parsed into the flat damage taken.
			// The row's one charge is spent by the next spell that hits the target.
			taken := frostfireVulnDebuff.EffectN(1)
			var takenSchools []stats.SchoolIndex
			for school := core.SpellSchoolHoly; school <= core.SpellSchoolArcane; school <<= 1 {
				if taken.Misc&int32(school) != 0 {
					takenSchools = append(takenSchools, school.SchoolIndex())
				}
			}
			debuffs := character.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
				aura := target.RegisterAura(spelldata.AuraConfig(frostfireVulnDebuff,
					spelldata.Label(frostfireVulnDebuff.Name+character.Label)))
				aura.ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) {
					for _, i := range takenSchools {
						target.PseudoStats.SchoolBonusSpellDamage[i] += taken.BaseValue()
					}
				}).ApplyOnExpire(func(_ *core.Aura, _ *core.Simulation) {
					for _, i := range takenSchools {
						target.PseudoStats.SchoolBonusSpellDamage[i] -= taken.BaseValue()
					}
				})

				consume := spelldata.ProcTrigger(character, frostfireVulnDebuff,
					func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
						aura.Deactivate(sim)
					})
				consume.Name = "Elemental Vulnerability - Consume" + character.Label
				consume.TriggerImmediately = true
				return aura.AttachProcTrigger(consume)
			})

			// The rate is the proc chance column. PROC_FLAG_DEAL_HARMFUL_SPELL decodes to the cast,
			// which names no target; the debuff goes on the one a damage spell landed on, the
			// tooltip's "your damage spells".
			trigger := spelldata.ProcTrigger(character, frostfireVulnerability,
				func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
					debuffs.Get(result.Target).Activate(sim)
				})
			trigger.Name = "Frostfire Regalia - 6PC"
			trigger.Callback = core.CallbackOnSpellHitDealt
			trigger.Outcome = core.OutcomeLanded
			trigger.RequireDamageDealt = false
			trigger.RequireDamageComponent = true

			setBonusAura.AttachProcTrigger(trigger)
		},
		8: func(agent core.Agent, setBonusAura *core.Aura) {
			character := agent.GetCharacter()

			// A_MOD_THREAT -100% on the schools the misc value masks is a multiplier of 0, which the
			// parse cannot divide back out, so the buff holds the multipliers it zeroes and puts
			// them back.
			buff := character.RegisterAura(spelldata.AuraConfig(frostfireNotThereBuff))
			threatEffect := frostfireNotThereBuff.EffectN(1)
			var schools []stats.SchoolIndex
			for school := core.SpellSchoolHoly; school <= core.SpellSchoolArcane; school <<= 1 {
				if threatEffect.Misc&int32(school) != 0 {
					schools = append(schools, school.SchoolIndex())
				}
			}
			var held [stats.SchoolLen]float64
			buff.ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) {
				for _, i := range schools {
					held[i] = character.PseudoStats.SchoolThreatMultiplier[i]
					character.PseudoStats.SchoolThreatMultiplier[i] *= core.MultiplierFromPercent(threatEffect.BaseValue())
				}
			}).ApplyOnExpire(func(_ *core.Aura, _ *core.Simulation) {
				for _, i := range schools {
					character.PseudoStats.SchoolThreatMultiplier[i] = held[i]
				}
			})

			// The row spends its one charge on the cast, but the sim reckons threat when a hit
			// deals, which a spell in flight does after its cast completes: the buff goes on the
			// next damage spell's hit.
			consume := spelldata.ProcTrigger(character, frostfireNotThereBuff,
				func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
					buff.Deactivate(sim)
				})
			consume.Name = "Not There - Consume"
			consume.Callback = core.CallbackOnSpellHitDealt
			consume.Outcome = core.OutcomeLanded
			consume.RequireDamageDealt = false
			consume.RequireDamageComponent = true
			consume.TriggerImmediately = true
			buff.AttachProcTrigger(consume)

			// The rate is the proc chance column. The proc effect's class flags name another
			// class's family, which the resolver drops. It listens to landed damage spells for the
			// consume's reason, so the hit that procs it keeps its threat and the next one loses it.
			trigger := spelldata.ProcTrigger(character, frostfireNotThere,
				func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
					buff.Activate(sim)
				})
			trigger.Name = "Frostfire Regalia - 8PC"
			trigger.Callback = core.CallbackOnSpellHitDealt
			trigger.Outcome = core.OutcomeLanded
			trigger.RequireDamageDealt = false
			trigger.RequireDamageComponent = true

			setBonusAura.AttachProcTrigger(trigger)
		},
	},
})

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

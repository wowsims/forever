package shaman

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (shaman *Shaman) registerEnhancementTalents() {
	// Tier 1
	// Earth's Grasp (16043) raises Stoneclaw Totem's health and Earthbind Totem's radius, which the
	// sim does not model.
	shaman.applyThunderingStrikes()
	shaman.applyAncestralKnowledge()

	// Tier 2
	// Guardian Totems (16258) improves Stoneskin, Windwall and Grounding Totem, which the sim does
	// not model.
	shaman.applyMentalDexterity()
	// Improved Ghost Wolf (16262) shortens Ghost Wolf's cast, which the sim does not model.
	shaman.applyImprovedLightningShield()

	// Tier 3
	shaman.applyElementalWeapons()
	shaman.applyShamanisticFocus()
	shaman.applyAnticipation()

	// Tier 4
	shaman.applyToughness()
	shaman.applyFlurry()
	shaman.applyStormstrike()

	// Tier 5
	shaman.applySpiritWeapons()
	shaman.applyMentalQuickness()
	shaman.applyImprovedStormstrike()

	// Tier 6
	shaman.applyMaelstromWeapon()

	// Tier 7
	shaman.applyRageOfTheFarseer()
}

// The spell requires a melee weapon, so the crit is a mod on what the equipped weapon allows.
func (shaman *Shaman) applyThunderingStrikes() {
	if shaman.Talents.ThunderingStrikes == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.ThunderingStrikes.Rank(shaman.Talents.ThunderingStrikes))
}

func (shaman *Shaman) applyAncestralKnowledge() {
	if shaman.Talents.AncestralKnowledge == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.AncestralKnowledge.Rank(shaman.Talents.AncestralKnowledge))
}

func (shaman *Shaman) applyMentalDexterity() {
	if shaman.Talents.MentalDexterity == 0 {
		return
	}

	// Effect 1 is the Intellect share of attack power the tooltip states.
	//
	// TODO: Effect 2, a dummy of 20 plus 5 per level that the tooltip does not mention, is skipped.
	spelldata.ParseStatic(&shaman.Character, spellData.MentalDexterity.Rank(shaman.Talents.MentalDexterity))
}

func (shaman *Shaman) applyImprovedLightningShield() {
	if shaman.Talents.ImprovedLightningShield == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.ImprovedLightningShield.Rank(shaman.Talents.ImprovedLightningShield))
}

func (shaman *Shaman) applyElementalWeapons() {
	if shaman.Talents.ElementalWeapons == 0 {
		return
	}

	rank := spellData.ElementalWeapons.Rank(shaman.Talents.ElementalWeapons)

	// Effect 2 is the Flametongue and Frostbrand damage. Effects 1 and 3 raise effect 1 of Rockbiter
	// Weapon (Passive) and Windfury Weapon, which is attack power on both spells; the parse would read
	// them as damage, so the imbues read them instead.
	spelldata.ParseStatic(&shaman.Character, rank, spelldata.Effects(2))

	shaman.rockbiterWeaponMultiplier = 1 + rank.EffectN(1).Percent()
	shaman.windfuryWeaponMultiplier = 1 + rank.EffectN(3).Percent()
}

// The cost reduction is permanent. No proc in game files.
func (shaman *Shaman) applyShamanisticFocus() {
	if !shaman.Talents.ShamanisticFocus {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.ShamanisticFocus.Highest())
}

func (shaman *Shaman) applyAnticipation() {
	if shaman.Talents.Anticipation == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.Anticipation.Rank(shaman.Talents.Anticipation))
}

func (shaman *Shaman) applyToughness() {
	if shaman.Talents.Toughness == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.Toughness.Rank(shaman.Talents.Toughness))
}

var flurryBuff = spellData.FlurryTriggered.Highest()

func (shaman *Shaman) applyFlurry() {
	if shaman.Talents.Flurry == 0 {
		return
	}

	// The haste is the talent's ranked dummy, which the tooltip of 16256 reads as $m1; the buff's own
	// 10% is replaced in game, its tooltip reading the applied $w1.
	flurryAura := shaman.RegisterAura(spelldata.AuraConfig(flurryBuff)).
		AttachMultiplyMeleeSpeed(spellData.Flurry.MultiplierAt(shaman.Talents.Flurry))

	chargeICD := core.Cooldown{
		Timer:    shaman.NewTimer(),
		Duration: flurryBuff.ICD(),
	}

	// The spell has no roll and the crit is the condition, read off the result here because the same
	// listener has to see the swings that use a charge. Any melee crit sets three charges. 16257's
	// proc flags name swings alone. Charge use in EV-0010. A white swing in the instant the charges
	// are gained uses none.
	var gainedByImbue bool
	trigger := spelldata.ProcTrigger(&shaman.Character, spellData.Flurry.Rank(shaman.Talents.Flurry),
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			imbueHit := spell.SpellID == windfuryWeaponStrike.ID
			if result.Outcome.Matches(core.OutcomeCrit) {
				flurryAura.Activate(sim)
				flurryAura.SetStacks(sim, flurryAura.MaxStacks)
				gainedByImbue = imbueHit
				return
			}

			if !flurryAura.IsActive() {
				return
			}
			spends := spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) || imbueHit
			gainedNow := flurryAura.RemainingDuration(sim) == flurryAura.Duration
			blocked := gainedNow && (!imbueHit || gainedByImbue)
			if spends && !blocked && chargeICD.IsReady(sim) {
				chargeICD.Use(sim)
				flurryAura.RemoveStack(sim)
			}
		})
	trigger.Name = "Flurry - Trigger"
	trigger.ProcMask = core.ProcMaskMelee
	trigger.Outcome = core.OutcomeEmpty
	trigger.RequireDamageDealt = false
	trigger.TriggerImmediately = true

	shaman.MakeProcTriggerAura(trigger)
}

func (shaman *Shaman) applyStormstrike() {
	if !shaman.Talents.Stormstrike {
		return
	}
	shaman.registerStormstrikeSpell()
}

func (shaman *Shaman) applySpiritWeapons() {
	if !shaman.Talents.SpiritWeapons {
		return
	}

	// "Gives a chance to parry enemy melee attacks."
	shaman.PseudoStats.CanParry = true

	rank := spellData.SpiritWeapons.Highest()

	// Effect 1 cuts threat always, and effect 2 raises Rockbiter Weapon (Passive)'s own threat
	// effect, which the imbue reads. Effect 3 is the dummy stating the result under Rockbiter.
	spelldata.ParseStatic(&shaman.Character, rank, spelldata.Effects(1))
	shaman.spiritWeaponsRockbiterThreat = rank.EffectN(2).Percent()
}

func (shaman *Shaman) applyMentalQuickness() {
	if shaman.Talents.MentalQuickness == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.MentalQuickness.Rank(shaman.Talents.MentalQuickness))
}

var improvedStormstrikeBuff = spellData.ImprovedStormstrikeTriggered.Highest()

func (shaman *Shaman) applyImprovedStormstrike() {
	if shaman.Talents.ImprovedStormstrike == 0 {
		return
	}

	rank := spellData.ImprovedStormstrike.Rank(shaman.Talents.ImprovedStormstrike)

	// The tooltip's $m1 is the chance on a Stormstrike to gain the regen and $m2 the chance a dodge
	// or parry resets the cooldown. Effect 3, a flat 20 the tooltip does not name, is unused.
	regenAura := shaman.RegisterAura(spelldata.AuraConfig(improvedStormstrikeBuff))
	spelldata.ParseEffects(&shaman.Character, regenAura, improvedStormstrikeBuff)

	shaman.MakeProcTriggerAura(core.ProcTrigger{
		Name:       "Improved Stormstrike - Regen Trigger",
		ActionID:   core.ActionID{SpellID: improvedStormstrikeBuff.ID},
		Callback:   core.CallbackOnCastComplete,
		ClassFlags: spellData.Stormstrike.Highest().ClassFlags,
		ProcChance: rank.EffectN(1).Percent(),
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			regenAura.Activate(sim)
		},
	})

	// The spell hears melee attacks taken; the dodge or parry is the outcome, which no ProcTypeMask
	// has a bit for. The generator reads the spell's chance off effect 1, where the tooltip names $m2.
	trigger := spelldata.ProcTrigger(&shaman.Character, rank,
		func(_ *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			if shaman.Stormstrike != nil {
				shaman.Stormstrike.CD.Reset()
			}
		},
		spelldata.ChanceFrom(rank.EffectN(2)))
	trigger.Name = "Improved Stormstrike - Reset Trigger"
	trigger.Outcome = core.OutcomeDodge | core.OutcomeParry
	trigger.RequireDamageDealt = false

	shaman.MakeProcTriggerAura(trigger)
}

var maelstromWeaponBuff = spellData.MaelstromWeaponTriggered.Highest()

// The talent's effect 2, the tooltip's $s2, is the rate per minute.
func maelstromWeaponPPM(rank *spelldata.Spell) float64 {
	return rank.EffectN(2).BaseValue()
}

func (shaman *Shaman) applyMaelstromWeapon() {
	if shaman.Talents.MaelstromWeapon == 0 {
		return
	}

	rank := spellData.MaelstromWeapon.Rank(shaman.Talents.MaelstromWeapon)

	perStack := rank.EffectN(1).Percent()
	castTimeMod := shaman.AddDynamicMod(core.SpellModConfig{
		Kind:       core.SpellMod_CastTime_Pct,
		ClassFlags: maelstromWeaponBuff.EffectN(1).ClassFlags,
	})
	costMod := shaman.AddDynamicMod(core.SpellModConfig{
		Kind:       core.SpellMod_PowerCost_Pct,
		ClassFlags: maelstromWeaponBuff.EffectN(2).ClassFlags,
	})
	consumes := maelstromWeaponBuff.EffectN(1).ClassFlags

	// Maelstrom Ready! (408501) goes up as the stacks reach five and comes off with them or at the
	// next Chain Lightning. Clearcasting kept in EV-0180. Maelstrom Ready!'s row, a dummy with one charge
	// and no duration, is in the game files but not in the spell store, so the aura is built here. The
	// Healing Wave and Lesser Healing Wave that also take it off are not in the sim.
	readyAura := shaman.RegisterAura(core.Aura{
		Label:    "Maelstrom Ready!",
		ActionID: core.ActionID{SpellID: 408501},
		Duration: core.NeverExpires,
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.Matches(SpellMaskChainLightning) {
				aura.Deactivate(sim)
			}
		},
	})

	config := spelldata.AuraConfig(maelstromWeaponBuff)
	config.OnGain = func(_ *core.Aura, _ *core.Simulation) {
		castTimeMod.Activate()
		costMod.Activate()
	}
	config.OnExpire = func(_ *core.Aura, sim *core.Simulation) {
		castTimeMod.Deactivate()
		costMod.Deactivate()
		readyAura.Deactivate(sim)
	}
	config.OnStacksChange = func(aura *core.Aura, sim *core.Simulation, _ int32, newStacks int32) {
		castTimeMod.UpdateFloatValue(perStack * float64(newStacks))
		costMod.UpdateFloatValue(perStack * float64(newStacks))
		if newStacks == aura.MaxStacks {
			readyAura.Activate(sim)
		}
	}
	config.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if consumes.Reaches(spell) {
			aura.Deactivate(sim)
		}
	}
	maelstromAura := shaman.RegisterAura(config)
	shaman.maelstromWeaponAura, shaman.maelstromReadyAura = maelstromAura, readyAura

	// The row's chance column is the "no roll" 100; the tooltip prices the chance per hit on the main
	// hand's speed, so the rate per minute is effect 2. Totem hit procs in EV-0009.
	trigger := spelldata.ProcTrigger(&shaman.Character, rank,
		func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			maelstromAura.Activate(sim)
			maelstromAura.AddStack(sim)
		},
		func(_ *core.Character, trigger *core.ProcTrigger) {
			trigger.ProcMask = core.ProcMaskMelee
			trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
				return spell.SpellID != windfuryTotemStrike.ID
			}
		},
		spelldata.PPM(maelstromWeaponPPM(rank)))
	trigger.Name = "Maelstrom Weapon - Trigger"

	shaman.MakeProcTriggerAura(trigger)
}

var rageOfTheFarseerRank = spellData.RageOfTheFarseer.Highest()

func (shaman *Shaman) applyRageOfTheFarseer() {
	if !shaman.Talents.RageOfTheFarseer {
		return
	}

	// Effect 1 is the melee and ranged attack speed, which the parse does not read.
	rageAura := shaman.RegisterAura(spelldata.AuraConfig(rageOfTheFarseerRank)).
		AttachMultiplyAttackSpeed(1 + rageOfTheFarseerRank.EffectN(1).Percent())

	config := spelldata.SpellConfig(&shaman.Unit, rageOfTheFarseerRank)
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		rageAura.Activate(sim)
	}
	config.RelatedSelfBuff = rageAura

	shaman.AddMajorCooldown(core.MajorCooldown{
		Spell: shaman.RegisterSpell(config),
		Type:  core.CooldownTypeDPS,
	})
}

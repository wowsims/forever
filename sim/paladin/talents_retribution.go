package paladin

import (
	"fmt"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (paladin *Paladin) registerRetributionTalents() {
	// Tier 1
	paladin.applyDeflection()
	paladin.applyBenediction()

	// Tier 2
	paladin.applyImprovedJudgement()
	paladin.applyConviction()
	paladin.applySanctifiedJudgement()

	// Tier 3
	paladin.applyVindication()
	// Unyielding Faith shortens Fear and Disorient effects, which the sim never suffers.
	// Seal of Command registered in registerTalentSpells
	paladin.applyPursuitOfJustice()

	// Tier 4
	paladin.applyEyeForAnEye()
	paladin.applySacredArbiter()
	paladin.applyTwoHandedWeaponSpecialization()

	// Tier 5
	// Turn Evil changes Turn Undead, which the sim does not model.
	// Repentance has nothing the sim models.
	paladin.applyVengeance()

	// Tier 6
	paladin.applyChampionOfTheLight()

	// Tier 7
	paladin.applyUnsealedWrath()
}

// Deflection - Increases your Parry chance by 1/2/3/4/5%.
func (paladin *Paladin) applyDeflection() {
	if paladin.Talents.Deflection == 0 {
		return
	}

	spelldata.ParseStatic(&paladin.Character, spellData.Deflection.Rank(paladin.Talents.Deflection))
}

// Benediction - Reduces the Mana cost of all instant cast spells and abilities by 2/4/6/8/10%.
// TODO: Verify in game that Benediction does not reduce Divine Favor's cost.
func (paladin *Paladin) applyBenediction() {
	if paladin.Talents.Benediction == 0 {
		return
	}

	spelldata.ParseStatic(&paladin.Character, spellData.Benediction.Rank(paladin.Talents.Benediction))
}

// Improved Judgement - Decreases the cooldown of your Judgement ability by 1/2 sec.
func (paladin *Paladin) applyImprovedJudgement() {
	if paladin.Talents.ImprovedJudgement == 0 {
		return
	}

	spelldata.ParseStatic(&paladin.Character, spellData.ImprovedJudgement.Rank(paladin.Talents.ImprovedJudgement))
}

// Conviction - Improves your chance to get a critical strike with all spells and attacks by 2/4/6%.
func (paladin *Paladin) applyConviction() {
	if paladin.Talents.Conviction == 0 {
		return
	}

	spelldata.ParseStatic(&paladin.Character, spellData.Conviction.Rank(paladin.Talents.Conviction))
}

// Vindication - Gives your damaging melee attacks a chance to reduce the target's Attack Power by
// 68/136/204 for 30 sec.
//
// The game files state no rate: the ProcChance column reads 100. In game it is 5 procs a minute on
// the weapon's speed, measured on swings at 1.6 and 3.6 speed under Seal of Righteousness and Seal of
// the Crusader. The row's proc mask also hears melee specials, which roll the same chance and were not
// measured. The debuff is a magic spell of its own that can miss.
const vindicationPPM = 5

func (paladin *Paladin) applyVindication() {
	if paladin.Talents.Vindication == 0 {
		return
	}

	rank := spellData.VindicationTriggered.Highest()

	// The raid's Vindication row, priced at the paladin's ranks, so it bids against the raid's
	// attack power reductions rather than stacking with them.
	targetAuras := paladin.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return buffs.NewSpellDataVindicationAura(target, true, paladin.Talents.Vindication)
	})

	// Magic in SpellCategories: the debuff rolls spell hit.
	debuffConfig := spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Magic(core.ProcMaskEmpty), spelldata.Proc())
	debuffConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
		if result.Landed() {
			targetAuras.Get(target).Activate(sim)
		}
		spell.DealOutcome(sim, result)
	}
	debuffConfig.RelatedAuraArrays = targetAuras.ToMap()
	debuff := paladin.RegisterSpell(debuffConfig)

	trigger := spelldata.ProcTrigger(&paladin.Character, spellData.Vindication.Rank(paladin.Talents.Vindication), func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		debuff.Cast(sim, result.Target)
	}, spelldata.WeaponSpeedPPM(vindicationPPM))
	trigger.Name = "Vindication - Trigger" + paladin.Label
	paladin.MakeProcTriggerAura(trigger)
}

// Sanctified Judgement - Gives your Judgement ability a 33/66/100% chance to return 20/40/60% of
// the Mana cost of the judged seal.
//
// Judgement fires no on-cast callbacks, so the trigger hears the judgement the seal unleashes: one
// per Judgement, and it refunds whether or not that judgement lands.
//
// The refund is a share of the seal's base cost, and the seal's percent cost cut goes on top of it
// instead of coming off: Benediction 5/5 makes a 90 mana seal cost 81 and refund 60% of 99, measured
// in game at Benediction 0, 1, 3 and 5 with Sanctified Judgement 3/3.
func (paladin *Paladin) applySanctifiedJudgement() {
	if paladin.Talents.SanctifiedJudgement == 0 {
		return
	}

	manaMetrics := paladin.NewManaMetrics(core.ActionID{SpellID: spellData.SanctifiedJudgement.Highest().ID})
	manaMetrics.NoThreat = spellData.SanctifiedJudgementTriggered.Highest().GeneratesNoThreat()
	refund := spellData.SanctifiedJudgement.EffectAt(2).FractionAt(paladin.Talents.SanctifiedJudgement)

	paladin.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Sanctified Judgement" + paladin.Label,
		Callback:           core.CallbackOnSpellHitDealt,
		ClassSpellMask:     SpellMaskAllJudgements,
		ProcChance:         spellData.SanctifiedJudgement.EffectAt(1).FractionAt(paladin.Talents.SanctifiedJudgement),
		TriggerImmediately: true,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			if seal := paladin.activeSeal(); seal != nil {
				cost := seal.spell.Cost
				paladin.AddMana(sim, float64(cost.BaseCost)*(2-cost.PercentModifier)*refund, manaMetrics)
			}
		},
	})
}

// The spell that deals Eye for an Eye's damage. The talent's dummy names no trigger, so no ladder
// reaches it.
var eyeForAnEyeDamage = spelldata.MustFind(25997)

// Eye for an Eye - All critical strikes against you cause 5/10% of the damage taken to the attacker
// as well. The damage caused by Eye for an Eye will not exceed 50% of the Paladin's total health.
func (paladin *Paladin) applyEyeForAnEye() {
	if paladin.Talents.EyeForAnEye == 0 {
		return
	}

	share := spellData.EyeForAnEye.FractionAt(paladin.Talents.EyeForAnEye)

	var reflected float64
	config := spelldata.SpellConfig(&paladin.Unit, eyeForAnEyeDamage, spelldata.Magic(core.ProcMaskEmpty), spelldata.Proc(),
		spelldata.Flags(core.SpellFlagBinary|core.SpellFlagIgnoreModifiers))
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, reflected, spell.OutcomeAlwaysHit)
	}
	reflect := paladin.RegisterSpell(config)

	trigger := spelldata.ProcTrigger(&paladin.Character, spellData.EyeForAnEye.Rank(paladin.Talents.EyeForAnEye), func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		reflected = min(result.Damage*share, paladin.MaxHealth()/2)
		reflect.Cast(sim, spell.Unit)
	})
	trigger.Name = "Eye for an Eye" + paladin.Label
	paladin.MakeProcTriggerAura(trigger)
}

// Pursuit of Justice - Increases movement speed and mounted movement speed by 8/15%. This does not
// stack with other movement speed increasing effects.
func (paladin *Paladin) applyPursuitOfJustice() {
	if paladin.Talents.PursuitOfJustice == 0 {
		return
	}

	rank := spellData.PursuitOfJustice.Highest()
	paladin.NewPassiveMovementSpeedAura(
		"Pursuit of Justice",
		core.ActionID{SpellID: rank.ID},
		spellData.PursuitOfJustice.Effect(dbcenums.A_MOD_INCREASE_SPEED, 0).FractionAt(paladin.Talents.PursuitOfJustice),
	)
}

// Sacred Arbiter - Increases the damage of your Holy Strike ability by 20% and causes it to refresh
// all Judgement effects on the target.
func (paladin *Paladin) applySacredArbiter() {
	if !paladin.Talents.SacredArbiter {
		return
	}

	spelldata.ParseStatic(&paladin.Character, spellData.SacredArbiter.Rank(1))

	paladin.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Sacred Arbiter" + paladin.Label,
		Callback:           core.CallbackOnSpellHitDealt,
		ClassSpellMask:     SpellMaskHolyStrike,
		Outcome:            core.OutcomeLanded,
		TriggerImmediately: true,
		Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			for _, aura := range result.Target.GetAurasWithTag(buffs.JudgementAuraTag) {
				if aura.IsActive() {
					aura.Refresh(sim)
				}
			}
		},
	})
}

// Two-Handed Weapon Specialization - Increases the physical damage you deal by 2/4/6%, and reduces
// all threat you generate by 7/14/20% while a two-handed melee weapon is equipped.
func (paladin *Paladin) applyTwoHandedWeaponSpecialization() {
	if paladin.Talents.TwoHandedWeaponSpecialization == 0 {
		return
	}

	spelldata.ParseStatic(&paladin.Character,
		spellData.TwoHandedWeaponSpecialization.Rank(paladin.Talents.TwoHandedWeaponSpecialization))
}

// Vengeance - Increases your Physical and Holy damage dealt by 1/2/3% for 30 sec after landing a
// non-periodic critical strike. Stacks up to 3 times.
func (paladin *Paladin) applyVengeance() {
	if paladin.Talents.Vengeance == 0 {
		return
	}

	rank := spellData.VengeanceTriggered.Highest()
	perStack := spellData.Vengeance.FractionAt(paladin.Talents.Vengeance)

	damageMod := paladin.AddDynamicMod(core.SpellModConfig{
		School:     core.SpellSchoolHoly | core.SpellSchoolPhysical,
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: perStack,
	})

	vengeance := paladin.RegisterAura(core.Aura{
		Label:     "Vengeance" + paladin.Label,
		ActionID:  core.ActionID{SpellID: rank.ID},
		Duration:  rank.Duration(),
		MaxStacks: int32(rank.MaxStack),
		OnGain: func(_ *core.Aura, _ *core.Simulation) {
			damageMod.Activate()
		},
		OnExpire: func(_ *core.Aura, _ *core.Simulation) {
			damageMod.Deactivate()
		},
		OnStacksChange: func(_ *core.Aura, _ *core.Simulation, _, newStacks int32) {
			damageMod.UpdateFloatValue(perStack * float64(newStacks))
		},
	})

	trigger := spelldata.ProcTrigger(&paladin.Character, spellData.Vengeance.Rank(paladin.Talents.Vengeance), func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		vengeance.Activate(sim)
		vengeance.AddStack(sim)
	})
	trigger.Name = "Vengeance - Trigger" + paladin.Label
	// The row's harmful-spell bit hears every spell, item spells too, which the sim registers under
	// ProcMaskEmpty, so the listener hears any hit.
	trigger.ProcMask = core.ProcMaskUnknown
	paladin.MakeProcTriggerAura(trigger)
}

// Champion of the Light - Increases your spell damage by up to 12/24/36/48/60% of your Intellect, and
// reduces the cast time of your Hammer of Wrath by 0.2/0.4/0.6/0.8/1.0 sec.
func (paladin *Paladin) applyChampionOfTheLight() {
	if paladin.Talents.ChampionOfTheLight == 0 {
		return
	}

	spelldata.ParseStatic(&paladin.Character, spellData.ChampionOfTheLight.Rank(paladin.Talents.ChampionOfTheLight))
}

// Unsealed Wrath - Your melee attacks store a copy of your active Seal for 15 sec, stacking up to 3
// times. Landing an attack with a different Seal consumes all stacks, giving the stored Seal its
// normal chance to activate once per stack.
//
// Each copy is the Echo of its seal, one aura per seal. The four Echoes state the same duration and
// stack count, and only Echo of Fury is a row of the store, so all four read it. A stored seal's
// effect lands a batch window after the hit that consumes it, as a live seal's does.
func (paladin *Paladin) applyUnsealedWrath() {
	if !paladin.Talents.UnsealedWrath {
		return
	}

	row := spellData.UnsealedWrathTriggered.Highest()
	paladin.echoes = nil
	for _, id := range []int32{echoOfCommandID, echoOfFuryID, echoOfRighteousnessID, echoOfJusticeID} {
		paladin.echoes = append(paladin.echoes, &sealEcho{
			id: id,
			aura: paladin.RegisterAura(core.Aura{
				Label:     fmt.Sprintf("Echo (%d)%s", id, paladin.Label),
				ActionID:  core.ActionID{SpellID: id},
				Duration:  row.Duration(),
				MaxStacks: int32(row.MaxStack),
			}),
		})
	}

	trigger := spelldata.ProcTrigger(&paladin.Character, spellData.UnsealedWrath.Rank(1), func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		seal := paladin.activeSeal()
		if seal == nil {
			return
		}
		target := result.Target
		for _, echo := range paladin.echoes {
			if echo.id == seal.echoID || !echo.aura.IsActive() {
				continue
			}
			stored, stacks := echo.seal, echo.aura.GetStacks()
			echo.seal = nil
			echo.aura.Deactivate(sim)
			if stored == nil || stored.echo == nil {
				continue
			}
			afterBatch(sim, core.ActionPriorityDOT, func(sim *core.Simulation) {
				for range stacks {
					stored.echo(sim, target)
				}
			})
		}
		for _, echo := range paladin.echoes {
			if echo.id == seal.echoID {
				echo.seal = seal
				echo.aura.Activate(sim)
				echo.aura.AddStack(sim)
			}
		}
	})
	trigger.Name = "Unsealed Wrath" + paladin.Label
	// A copy is stored and spent on a hit that lands.
	trigger.Outcome = core.OutcomeLanded
	trigger.TriggerImmediately = true
	paladin.MakeProcTriggerAura(trigger)
}

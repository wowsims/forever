package rogue

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (rogue *Rogue) registerAssassinationTalents() {
	// Tier 1
	// Gouge NYI
	rogue.registerRemorselessAttacks()
	rogue.registerMalice()

	// Tier 2
	// Ruthlessness modeled in rogue.go ApplyFinisher
	rogue.registerMurder()
	// Improved Slice modeled in Slice and Dice directly

	// Tier 3
	// Relentless Strikes modeled in rogue.go ApplyFinisher
	rogue.registerImprovedExposeArmor()
	rogue.registerLethality()

	// Tier 4
	rogue.registerVilePoisons()
	rogue.registerColdBlood()
	rogue.registerImprovedPoisons()

	// Tier 5
	// Vigor modeled in rogue.go during initialization
	rogue.registerMutilate()
	// Kidney Shot NYI

	// Tier 6
	rogue.registerSealFate()

	// Tier 7
	rogue.registerVenom()
}

func (rogue *Rogue) registerImprovedEviscerate() {
	if rogue.Talents.ImprovedEviscerate == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.ImprovedEviscerate.Rank(rogue.Talents.ImprovedEviscerate))
}

var remorselessRank = spellData.RemorselessAttacksTriggered.Highest()

// The talent row hears the killing blow, and a credited kill grants Remorseless. The buff row states
// a flat 40% crit where the talent's effect 1 is 20/40% by rank, so the row is scaled by the ranks
// taken. The spells its crit mod names spend its one charge.
func (rogue *Rogue) registerRemorselessAttacks() {
	if rogue.Talents.RemorselessAttacks == 0 {
		return
	}

	buff := rogue.RegisterAura(spelldata.AuraConfig(remorselessRank))
	spelldata.ParseEffects(&rogue.Character, buff, remorselessRank,
		spelldata.ScaledByRanks(rogue.Talents.RemorselessAttacks, spellData.RemorselessAttacks.Len()), spelldata.IgnoreStacks())
	buff.ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
		aura.SetStacks(sim, aura.MaxStacks)
	})
	// A miss, dodge or parry of a qualifying strike spends the charge as a landed one does.
	spelldata.AttachChargeSpender(&rogue.Character, buff, remorselessRank, "Remorseless Attacks - Spend",
		func(_ *core.Character, trigger *core.ProcTrigger) {
			trigger.Callback = core.CallbackOnCastResult
			trigger.Outcome = core.OutcomeEmpty
			trigger.RequireDamageDealt = false
			trigger.RequireDamageComponent = false
			trigger.TriggerImmediately = true
		})

	rogue.MakeProcTriggerAura(spelldata.ProcTrigger(&rogue.Character,
		spellData.RemorselessAttacks.Rank(rogue.Talents.RemorselessAttacks),
		func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			buff.Activate(sim)
		}))
}

func (rogue *Rogue) registerMalice() {
	if rogue.Talents.Malice == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.Malice.Rank(rogue.Talents.Malice))
}

func (rogue *Rogue) registerMurder() {
	if rogue.Talents.Murder == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.Murder.Rank(rogue.Talents.Murder))
}

func (rogue *Rogue) registerImprovedExposeArmor() {
	if rogue.Talents.ImprovedExposeArmor == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.ImprovedExposeArmor.Rank(rogue.Talents.ImprovedExposeArmor))
}

func (rogue *Rogue) registerLethality() {
	if rogue.Talents.Lethality == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.Lethality.Rank(rogue.Talents.Lethality))
}

func (rogue *Rogue) registerVilePoisons() {
	if rogue.Talents.VilePoisons == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.VilePoisons.Rank(rogue.Talents.VilePoisons))
}

var coldBloodRank = spellData.ColdBlood.Highest()

func (rogue *Rogue) registerColdBlood() {
	if !rogue.Talents.ColdBlood {
		return
	}

	coldBloodAura := rogue.RegisterAura(spelldata.AuraConfig(coldBloodRank))

	coldBloodAura.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if spell.Flags.Matches(SpellFlagColdBlooded) && spell.ProcMask.Matches(core.ProcMaskMeleeMH) {
			aura.Deactivate(sim)
		}
	}

	spelldata.ParseEffects(&rogue.Character, coldBloodAura, coldBloodRank)
	config := spelldata.SpellConfig(&rogue.Unit, coldBloodRank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagMCD))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		coldBloodAura.Activate(sim)
	}

	rogue.ColdBlood = rogue.RegisterSpell(config)

	rogue.AddMajorCooldown(core.MajorCooldown{
		Spell: rogue.ColdBlood,
		Type:  core.CooldownTypeDPS,
	})
}

func (rogue *Rogue) registerImprovedPoisons() {
	if rogue.Talents.ImprovedPoisons == 0 {
		return
	}

	spelldata.ParseStatic(&rogue.Character,
		spellData.ImprovedPoisons.Rank(rogue.Talents.ImprovedPoisons))
}

func (rogue *Rogue) registerSealFate() {
	if rogue.Talents.SealFate == 0 {
		return
	}

	sfRank := spellData.SealFate.Rank(rogue.Talents.SealFate)
	sfMetrics := rogue.NewComboPointMetrics(core.ActionID{SpellID: sfRank.ID})

	trigger := spelldata.ProcTrigger(&rogue.Character, sfRank,
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Outcome.Matches(core.OutcomeCrit) {
				rogue.AddComboPoints(sim, 1, sfMetrics)
			}
		})
	trigger.Name = "Seal Fate - Trigger"
	trigger.SpellFlags = SpellFlagBuilder
	trigger.TriggerImmediately = true

	rogue.MakeProcTriggerAura(trigger)
}

var mutilateCastRank = spellData.Mutilate.Highest()

func (rogue *Rogue) registerMutilate() {
	if !rogue.Talents.Mutilate {
		return
	}

	rogue.MutilateMH = rogue.newMutilateHitSpell(true)
	rogue.MutilateOH = rogue.newMutilateHitSpell(false)

	// TODO: Verify if Mutilate (Cast) can proc MH poisons itself (detected as a double MH poison proc)
	config := spelldata.SpellConfig(&rogue.Unit, mutilateCastRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	pointGen := mutilateCastRank.EnergizeEffect().Average()
	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return rogue.HasDagger(core.MainHand) && rogue.HasDagger(core.OffHand)
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHit) // Miss/Dodge/Parry/Hit
		if result.Landed() {
			rogue.AddComboPoints(sim, int32(pointGen), spell.ComboPointMetrics())
			rogue.MutilateOH.Cast(sim, target)
			rogue.MutilateMH.Cast(sim, target)
		} else {
			spell.IssueRefund(sim)
		}
		spell.DealOutcome(sim, result)
	}

	rogue.Mutilate = rogue.RegisterSpell(config)
}

func (rogue *Rogue) newMutilateHitSpell(isMH bool) *core.Spell {
	hitRank := mutilateCastRank.EffectN(3).Trigger() // 1241590, off hand
	spellOpt := spelldata.Melee(core.ProcMaskMeleeOHSpecial)
	if isMH {
		hitRank = mutilateCastRank.EffectN(2).Trigger() // 1241586, main hand
		spellOpt = spelldata.Melee(core.ProcMaskMeleeMHSpecial)
	}
	mutilateHitBaseDamage := hitRank.DamageEffect().Average()
	config := spelldata.SpellConfig(&rogue.Unit, hitRank, spellOpt)
	// Part of the Mutilate cast: its spell id, so metrics see one cast, and its class flags, so whatever
	// names the cast reaches the strikes, Eureka!'s damage modifier and Head Rush alike.
	// TODO: Ingame test needed: whether Eureka! raises Mutilate's strikes.
	config.Flags |= SpellFlagBuilder | SpellFlagColdBlooded | core.SpellFlagPassiveSpell
	config.Flags &^= core.SpellFlagAPL
	config.SpellID = mutilateCastRank.ID
	config.ActionID = config.ActionID.WithHand(core.Ternary(isMH, proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand))
	for i := range config.ClassFlags.Mask {
		config.ClassFlags.Mask[i] |= mutilateCastRank.ClassFlags.Mask[i]
	}
	config.DamageMultiplier = hitRank.EffectN(2).Percent()

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := mutilateHitBaseDamage
		if isMH {
			baseDamage += rogue.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
		} else {
			baseDamage += rogue.OHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
		}

		oldMultiplier := spell.DamageMultiplier
		// TODO: Detect any poison on the target, not just rogue poisons
		if (rogue.DeadlyPoison != nil && rogue.DeadlyPoison.Dot(target).IsActive()) || rogue.WoundPoisonDebuffAuras.Get(target).IsActive() {
			spell.DamageMultiplier *= 1 + mutilateCastRank.EffectN(4).Percent()
		}

		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialBlockAndCrit)
		spell.DamageMultiplier = oldMultiplier
	}

	return rogue.RegisterSpell(config)
}

var venomRank = spellData.Venom.Highest()

func (rogue *Rogue) registerVenom() {
	if !rogue.Talents.Venom {
		return
	}

	config := spelldata.SpellConfig(&rogue.Unit, venomRank, spelldata.Melee(core.ProcMaskEmpty))
	venomAura := rogue.RegisterAura(spelldata.AuraConfig(venomRank))
	spelldata.ParseEffects(&rogue.Character, venomAura, venomRank)
	config.ExtraCastCondition = rogue.FinisherExtraCastCondition

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim) // As this finisher deals no damage, it may or may not break stealth. To be tested.
		venomAura.Duration = venomRank.Duration() + time.Second*3*time.Duration(rogue.ComboPoints())
		rogue.ApplyFinisher(sim, spell)
		venomAura.Activate(sim)
	}

	rogue.RegisterSpell(config)
}

package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func (warrior *Warrior) registerFuryTalents() {
	// Tier 1
	warrior.registerBoomingVoice()
	warrior.registerCruelty()

	// Tier 2
	// Lingering Rage (1323964) delays the rage decay after leaving combat, which the sim does not model.
	warrior.registerUnbridledWrath()

	// Tier 3
	warrior.registerFuriousPrecision()
	warrior.registerPiercingHowl()
	warrior.registerBloodCraze()

	// Tier 4
	warrior.registerDualWieldSpecialization()
	warrior.registerRagingBlows()
	warrior.registerEnrage()
	warrior.registerImprovedExecute()

	// Tier 5
	// Improved Berserker Rage: berserker_rage.go
	warrior.registerDeathWish()
	warrior.registerImprovedIntercept()

	// Tier 6
	warrior.registerFlurry()
	warrior.registerGoreDrinker()

	// Tier 7
	warrior.registerBloodthirst()
}

func (warrior *Warrior) registerBoomingVoice() {
	if warrior.Talents.BoomingVoice == 0 {
		return
	}

	// The shout radius has no sim kind, so only the shouts' rage cost comes off the row.
	spelldata.ParseStatic(&warrior.Character, spellData.BoomingVoice.Rank(warrior.Talents.BoomingVoice))
}

func (warrior *Warrior) registerCruelty() {
	if warrior.Talents.Cruelty == 0 {
		return
	}

	spelldata.ParseStatic(&warrior.Character, spellData.Cruelty.Rank(warrior.Talents.Cruelty))
}

var unbridledWrathRank = spellData.UnbridledWrathTriggered.Highest()
var unbridledWrathRage = unbridledWrathRank.EnergizeEffect().Tenths()

func (warrior *Warrior) registerUnbridledWrath() {
	if warrior.Talents.UnbridledWrath == 0 {
		return
	}

	rageMetrics := warrior.NewRageMetrics(core.ActionID{SpellID: unbridledWrathRank.ID})

	// The rate is the effect ladder the tooltip's $m1 names - 12/24/36/48/60 by rank, where the
	// proc chance column reads 100 - and the row's mask is the white hits the tooltip means.
	warrior.MakeProcTriggerAura(spelldata.ProcTrigger(&warrior.Character,
		spellData.UnbridledWrath.Rank(warrior.Talents.UnbridledWrath),
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			warrior.AddRage(sim, unbridledWrathRage, rageMetrics)
		}))
}

func (warrior *Warrior) registerFuriousPrecision() {
	if warrior.Talents.FuriousPrecision == 0 {
		return
	}

	// The row requires a one-handed weapon in the off hand, so its hit is a mod on the attacks the
	// off-hand weapon makes.
	spelldata.ParseStatic(&warrior.Character, spellData.FuriousPrecision.Rank(warrior.Talents.FuriousPrecision))
}

func (warrior *Warrior) registerDualWieldSpecialization() {
	if warrior.Talents.DualWieldSpecialization == 0 {
		return
	}

	rank := spellData.DualWieldSpecialization.Rank(warrior.Talents.DualWieldSpecialization)

	spelldata.ParseStatic(&warrior.Character, rank)

	// The off-hand rage the tooltip's $m2 states is the dummy at index 2, which the parse skips.
	warrior.SetOffHandRageMultiplier(core.MultiplierFromPercent(rank.EffectN(2).BaseValue()))
}

func (warrior *Warrior) registerImprovedExecute() {
	if warrior.Talents.ImprovedExecute == 0 {
		return
	}

	spelldata.ParseStatic(&warrior.Character, spellData.ImprovedExecute.Rank(warrior.Talents.ImprovedExecute))
}

var enrageBuff = spellData.EnrageTriggered.Highest()

func (warrior *Warrior) registerEnrage() {
	if warrior.Talents.Enrage == 0 {
		return
	}

	warrior.EnrageAura = warrior.GetOrRegisterAura(spelldata.AuraConfig(enrageBuff))

	// 12880 states the bucket and the duration, and the sim's only Enrage buff row states a flat
	// 10; the percentage per rank is the talent's own ladder, which is what 12317's tooltip reads
	// as $m1%. So the value is the talent's and only the bucket - the physical school's damage
	// dealt multiplier, not a per-spell mod - comes off the buff row.
	warrior.EnrageAura.AttachMultiplicativePseudoStatBuff(
		&warrior.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical],
		spellData.Enrage.MultiplierAt(warrior.Talents.Enrage),
	)

	warrior.EnrageAura.NewExclusiveEffect("Enrage", true, core.ExclusiveEffect{Priority: spellData.Enrage.ValueAt(warrior.Talents.Enrage)})

	// The rate is the shape where the tooltip carries $h%: a real 30% roll on damage taken, which
	// is the proc chance column.
	trigger := spelldata.ProcTrigger(&warrior.Character, spellData.Enrage.Rank(warrior.Talents.Enrage),
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			warrior.EnrageAura.Activate(sim)
			warrior.activateGoreDrinker(sim)
		})
	trigger.Name = "Enrage - Trigger"

	warrior.MakeProcTriggerAura(trigger)
}

var flurryBuff = spellData.FlurryTriggered.Highest()

func (warrior *Warrior) registerFlurry() {
	if warrior.Talents.Flurry == 0 {
		return
	}

	// 12966 supplies the duration and the three charges. The haste is the talent's 5% per point,
	// which 12319 reads as "Increases your melee attack speed by $m1% for your next $12966n swings";
	// the buff row's flat 30% does not apply.
	flurryAura := warrior.RegisterAura(spelldata.AuraConfig(flurryBuff)).
		AttachMultiplyMeleeSpeed(spellData.Flurry.MultiplierAt(warrior.Talents.Flurry))

	// The proc shape with no roll: 12319 states its rate as "always" and the crit is the
	// condition, which this handler reads off the result because the same listener hears every
	// melee result. Every white swing uses a charge whatever its outcome, and a Heroic Strike or
	// Cleave uses one only when it misses. The row's mask reaches ranged and spell hits too; the
	// tooltip says a melee critical strike, so the mask stays the caller's.
	//
	// TODO: Ingame test needed: the Cleave miss follows the measured Heroic Strike miss; no log
	// holds a missed Cleave yet.
	trigger := spelldata.ProcTrigger(&warrior.Character, spellData.Flurry.Rank(warrior.Talents.Flurry),
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spellData.Whirlwind.Contains(spell) && spell.ProcMask.Matches(core.ProcMaskMeleeOH) {
				return
			}

			if result.Outcome.Matches(core.OutcomeCrit) {
				flurryAura.Activate(sim)
				flurryAura.SetStacks(sim, int32(flurryBuff.ProcCharges))
				return
			}

			if !flurryAura.IsActive() {
				return
			}

			if spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit) ||
				(result.Outcome.Matches(core.OutcomeMiss) &&
					(spellData.HeroicStrike.Contains(spell) || spellData.Cleave.Contains(spell))) {
				flurryAura.RemoveStack(sim)
			}
		})
	trigger.Name = "Flurry - Trigger"
	trigger.ProcMask = core.ProcMaskMelee
	trigger.Outcome = core.OutcomeEmpty
	trigger.RequireDamageDealt = false
	trigger.TriggerImmediately = true

	warrior.MakeProcTriggerAura(trigger)
}

var goreDrinkerBuff = spellData.GoreDrinkerTriggered.Highest()
var goreDrinkerHeal = goreDrinkerBuff.EffectN(1).Trigger()

func (warrior *Warrior) registerGoreDrinker() {
	if warrior.Talents.GoreDrinker == 0 {
		return
	}

	charges := int32(spellData.GoreDrinker.EffectAt(1).ValueAt(warrior.Talents.GoreDrinker))
	healthFraction := spellData.GoreDrinker.EffectAt(2).FractionAt(warrior.Talents.GoreDrinker)

	// The heal row states no amount: the buff's trigger carries it as a value, which is the talent's
	// share of maximum health.
	healConfig := spelldata.SpellConfig(&warrior.Unit, goreDrinkerHeal, spelldata.Proc())
	healConfig.ProcMask = core.ProcMaskSpellHealing
	healConfig.DamageMultiplier = 1
	healConfig.ThreatMultiplier = 1
	healConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealHealing(sim, target, warrior.MaxHealth()*healthFraction, spell.OutcomeAlwaysHit)
	}
	heal := warrior.RegisterSpell(healConfig)

	firstAoeHit, resetFirstAoeHit := core.OncePerCast()

	buffConfig := spelldata.AuraConfig(goreDrinkerBuff)
	buffConfig.MaxStacks = charges
	buffConfig.OnReset = func(_ *core.Aura, _ *core.Simulation) {
		resetFirstAoeHit()
	}
	warrior.GoreDrinkerAura = warrior.RegisterAura(buffConfig)

	// TODO: Ingame test needed: a melee hit that deals no damage spends no charge.
	// TODO: Ingame test needed: Cleave and Whirlwind spend one charge and heal once per cast, however
	// many targets they hit.
	consume := spelldata.ProcTrigger(&warrior.Character, goreDrinkerBuff,
		func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			// The hits of one spell batch each queue a handler, which can outnumber the charges left.
			if !warrior.GoreDrinkerAura.IsActive() {
				return
			}
			heal.Cast(sim, &warrior.Unit)
			warrior.GoreDrinkerAura.RemoveStack(sim)
		})
	consume.Name = "Gore Drinker - Heal"
	consume.Outcome = core.OutcomeLanded
	consume.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		if !spell.ProcMask.Matches(core.ProcMaskMeleeSpecial) {
			return true
		}
		if !spellData.Cleave.Contains(spell) && !spellData.Whirlwind.Contains(spell) {
			return true
		}
		return firstAoeHit(sim, spell)
	}
	warrior.GoreDrinkerAura.AttachProcTrigger(consume)

	// The talent's effects name no spells, so the abilities are the tooltip's: Berserker Rage,
	// Bloodrage and Death Wish when cast, Bloodthirst when it lands, and Enrage from its own trigger.
	warrior.MakeProcTriggerAura(core.ProcTrigger{
		Name:     "Gore Drinker - Trigger",
		Callback: core.CallbackOnCastComplete | core.CallbackOnSpellHitDealt,
		Outcome:  core.OutcomeLanded,
		ExtraCondition: func(_ *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
			if result == nil {
				return spellData.BerserkerRage.Contains(spell) || spellData.Bloodrage.Contains(spell) || spellData.DeathWish.Contains(spell)
			}
			return spell.ProcMask.Matches(core.ProcMaskMeleeMHSpecial) && spellData.Bloodthirst.Contains(spell)
		},
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			warrior.activateGoreDrinker(sim)
		},
	})
}

func (warrior *Warrior) activateGoreDrinker(sim *core.Simulation) {
	if warrior.GoreDrinkerAura == nil {
		return
	}

	warrior.GoreDrinkerAura.Activate(sim)
	warrior.GoreDrinkerAura.SetStacks(sim, warrior.GoreDrinkerAura.MaxStacks)
}

var bloodthirstRank = spellData.Bloodthirst.ByID(23894)

func (warrior *Warrior) registerBloodthirst() {
	if !warrior.Talents.Bloodthirst {
		return
	}

	apShare := bloodthirstRank.EffectN(2).Percent()

	config := spelldata.SpellConfig(&warrior.Unit, bloodthirstRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := spell.MeleeAttackPower(target)*apShare + bloodthirstRank.DamageEffect().Roll(sim)
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.RegisterSpell(config)
}

var piercingHowlRank = spellData.PiercingHowl.Highest()

// TODO: In-game test required if there's any threat interaction
func (warrior *Warrior) registerPiercingHowl() {
	if !warrior.Talents.PiercingHowl {
		return
	}

	config := spelldata.SpellConfig(&warrior.Unit, piercingHowlRank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty

	warrior.RegisterSpell(config)
}

var bloodCrazeHot = spellData.BloodCrazeTriggered.Highest()

func (warrior *Warrior) registerBloodCraze() {
	if warrior.Talents.BloodCraze == 0 {
		return
	}

	healthFraction := spellData.BloodCraze.EffectAt(1).FractionAt(warrior.Talents.BloodCraze)
	hitThreshold := spellData.BloodCraze.EffectAt(2).FractionAt(warrior.Talents.BloodCraze)
	tick := bloodCrazeHot.PeriodicEffect()

	config := spelldata.SpellConfig(&warrior.Unit, bloodCrazeHot, spelldata.Proc())
	config.ProcMask = core.ProcMaskSpellHealing
	config.DamageMultiplier = 1
	config.ThreatMultiplier = 1

	config.Hot = spelldata.DotConfig(bloodCrazeHot, tick)
	config.Hot.SelfOnly = true
	config.Hot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		healPerTick := warrior.MaxHealth() * healthFraction / float64(dot.ExpectedTickCount())
		dot.Spell.CalcAndDealPeriodicHealing(sim, target, healPerTick, dot.OutcomeTick)
	}

	bloodCraze := warrior.RegisterSpell(config)

	// 16487 decodes to hits taken, and its crit hint covers only half of the tooltip's condition - a
	// crit taken or a hit over a share of maximum health - so the trigger takes the landed outcome
	// and the condition by hand.
	taken := spelldata.ProcTrigger(&warrior.Character, spellData.BloodCraze.Rank(warrior.Talents.BloodCraze),
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			bloodCraze.SelfHot().Apply(sim)
		})
	taken.Name = "Blood Craze - Damage Taken"
	taken.Outcome = core.OutcomeLanded
	taken.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
		return result.Outcome.Matches(core.OutcomeCrit) || result.Damage > warrior.MaxHealth()*hitThreshold
	}
	warrior.MakeProcTriggerAura(taken)
}

func (warrior *Warrior) registerRagingBlows() {
	if !warrior.Talents.RagingBlows {
		return
	}

	spelldata.ParseStatic(&warrior.Character, spellData.RagingBlows.Rank(1))
}

var deathWishRank = spellData.DeathWish.Highest()

func (warrior *Warrior) registerDeathWish() {
	if !warrior.Talents.DeathWish {
		return
	}

	deathWishAura := warrior.RegisterAura(spelldata.AuraConfig(deathWishRank))
	spelldata.ParseEffects(&warrior.Character, deathWishAura, deathWishRank)

	// Grants immunity to Fear effects, which the row states as A_MECHANIC_IMMUNITY and the parse
	// skips.
	deathWishAura.AttachFearImmunity()

	config := spelldata.SpellConfig(&warrior.Unit, deathWishRank,
		spelldata.Flags(core.SpellFlagCastWhileIncapacitated))

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		deathWishAura.Activate(sim)
	}

	config.RelatedSelfBuff = deathWishAura

	deathWishSpell := warrior.RegisterSpell(config)

	warrior.AddMajorCooldown(core.MajorCooldown{
		Spell: deathWishSpell,
		Type:  core.CooldownTypeDPS,
	})
}

func (warrior *Warrior) registerImprovedIntercept() {
	if warrior.Talents.ImprovedIntercept == 0 {
		return
	}

	spelldata.ParseStatic(&warrior.Character, spellData.ImprovedIntercept.Rank(warrior.Talents.ImprovedIntercept))
}

package mage

import (
	"math"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (mage *Mage) registerFireTalents() {
	// Tier 1
	mage.registerImprovedFireBlast()
	mage.registerIncineration()
	mage.registerAccelerant()

	// Tier 2
	mage.registerIgnite()
	mage.registerFlameThrowing()
	mage.registerImpact()

	// Tier 3
	mage.registerBurningSoul()
	mage.registerImprovedFlamestrike()
	mage.registerPyroblastTalent()

	// Tier 4
	// Improved Scorch implemented in scorch.go
	mage.registerImprovedFireWard()
	mage.registerHeatingUp()
	mage.registerMasterOfElements()

	// Tier 5
	mage.registerCriticalMass()
	mage.registerBlastWaveTalent()

	// Tier 6
	mage.registerFirePower()

	// Tier 7
	mage.registerCombustionTalent()
}

func (mage *Mage) registerImprovedFireBlast() {
	if mage.Talents.ImprovedFireBlast == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ImprovedFireBlast.Rank(mage.Talents.ImprovedFireBlast))
}

func (mage *Mage) registerIncineration() {
	if mage.Talents.Incineration == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.Incineration.Rank(mage.Talents.Incineration))
}

// "Reduces the casting time of your Fireball, Frostfire Bolt, and Arcane Wrath spells by
// 0.1/0.2/0.3/0.4/0.5 sec."
func (mage *Mage) registerAccelerant() {
	if mage.Talents.Accelerant == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.Accelerant.Rank(mage.Talents.Accelerant))
}

// A Fire spell's crit burns the target for a share of its damage (11119, 8% per point) over 4 sec.
// The DoT (412538) stacks to 1, so it rolls: a new crit adds its share to what the current DoT has
// not ticked yet and restarts it over the full duration.
func (mage *Mage) registerIgnite() {
	if mage.Talents.Ignite == 0 {
		return
	}

	igniteFraction := spellData.Ignite.FractionAt(mage.Talents.Ignite)
	igniteDot := spellData.IgniteTriggered.Highest()

	// The damage is already the crit's, after every modifier and resist, so the ticks take none again.
	config := spelldata.SpellConfig(&mage.Unit, igniteDot, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc(),
		spelldata.Flags(core.SpellFlagIgnoreModifiers|core.SpellFlagNoSpellMods|core.SpellFlagIgnoreResists|core.SpellFlagProc))
	config.ClassSpellMask = MageSpellIgnite

	config.Dot = spelldata.DotConfig(igniteDot, igniteDot.EffectN(1))
	// The stacks show the damage per tick, which an APL can read; the row's 1 is the rolling cap.
	config.Dot.Aura.MaxStacks = math.MaxInt32
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, dot.SnapshotBaseDamage, dot.OutcomeTick)
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.Dot(target).Apply(sim)
	}

	mage.Ignite = mage.RegisterSpell(config)

	trigger := spelldata.ProcTrigger(&mage.Character, spellData.Ignite.Rank(mage.Talents.Ignite),
		func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			target := result.Target
			dot := mage.Ignite.Dot(target)
			totalDamage := dot.OutstandingDmg() + result.Damage*igniteFraction

			mage.Ignite.Cast(sim, target)
			dot.SnapshotBaseDamage = totalDamage / float64(dot.BaseTickCount)
			dot.Aura.SetStacks(sim, int32(dot.SnapshotBaseDamage))
		})
	trigger.Name = "Ignite - Trigger"
	trigger.TriggerImmediately = true
	// "Your critical strikes from Fire damage spells": the school, so Frostfire Bolt counts.
	trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return spell.SpellSchool.Matches(core.SpellSchoolFire)
	}

	mage.MakeProcTriggerAura(trigger)
}

func (mage *Mage) registerFlameThrowing() {
	if mage.Talents.FlameThrowing == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.FlameThrowing.Rank(mage.Talents.FlameThrowing))
}

// "Gives your Fire spells a 10% chance to stun the target for 2 sec."
//
// Not modelled: the proc (12355) is an A_MOD_STUN, which a boss the sim fights is immune to, the same
// reason Frostbite's freeze is not modelled.
func (mage *Mage) registerImpact() {
	if mage.Talents.Impact == 0 {
		return
	}
}

func (mage *Mage) registerBurningSoul() {
	if mage.Talents.BurningSoul == 0 {
		return
	}

	// Effect 1 is pushback, which the sim does not model. Effect 2 is the threat, on misc 4, the Fire
	// school.
	spelldata.ParseStatic(&mage.Character, spellData.BurningSoul.Rank(mage.Talents.BurningSoul), spelldata.Effects(2))
}

func (mage *Mage) registerImprovedFlamestrike() {
	if mage.Talents.ImprovedFlamestrike == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.ImprovedFlamestrike.Rank(mage.Talents.ImprovedFlamestrike))
}

func (mage *Mage) registerPyroblastTalent() {
	if !mage.Talents.Pyroblast {
		return
	}

	mage.registerPyroblastSpell()
}

func (mage *Mage) registerBlastWaveTalent() {
	if !mage.Talents.BlastWave {
		return
	}

	mage.registerBlastWaveSpell()
}

func (mage *Mage) registerCombustionTalent() {
	if !mage.Talents.Combustion {
		return
	}

	mage.registerCombustionSpell()
}

// Improved Fire Ward, a chance for Fire Ward to reflect Fire spells (effect 1, a flat modifier on
// Fire Ward's effect 2). The reflect needs an enemy casting Fire at the mage, so it is not modelled.
func (mage *Mage) registerImprovedFireWard() {
	if mage.Talents.ImprovedFireWard == 0 {
		return
	}
}

func (mage *Mage) registerHeatingUp() {
	if !mage.Talents.HeatingUp {
		return
	}

	heatingUpBuff := spellData.HeatingUpTriggered.Highest()

	config := spelldata.AuraConfig(heatingUpBuff)

	// A Pyroblast consumes every stack, checked in game, unless Presence of Mind made it instant.
	// Presence of Mind may drop before or after this runs on the same cast, so the time it has been
	// down tells both apart.
	config.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if spell.Matches(MageSpellPyroblast) && mage.PresenceOfMindAura.TimeInactive(sim) != 0 {
			aura.Deactivate(sim)
		}
	}

	heatingUpAura := mage.RegisterAura(config)

	spelldata.ParseEffects(&mage.Character, heatingUpAura, heatingUpBuff, spelldata.Effects(1))

	addStack := func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		heatingUpAura.Activate(sim)
		heatingUpAura.AddStack(sim)
	}

	trigger := spelldata.ProcTrigger(&mage.Character, spellData.HeatingUp.Rank(1), addStack)
	trigger.Name = "Heating Up - Trigger"
	trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return spellData.Fireball.Contains(spell) || spellData.FrostfireBolt.Contains(spell) ||
			spellData.FireBlast.Contains(spell) || spellData.Scorch.Contains(spell)
	}

	mage.MakeProcTriggerAura(trigger)

	// "Fire Blast grants an additional stack every time it deals damage", crit or not, so a crit
	// grants two.
	fireBlastTrigger := trigger
	fireBlastTrigger.Name = "Heating Up - Fire Blast Trigger"
	fireBlastTrigger.Outcome = core.OutcomeLanded
	fireBlastTrigger.RequireDamageDealt = true
	fireBlastTrigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		return spellData.FireBlast.Contains(spell)
	}

	mage.MakeProcTriggerAura(fireBlastTrigger)
}

func (mage *Mage) registerMasterOfElements() {
	if mage.Talents.MasterOfElements == 0 {
		return
	}

	refundFraction := spellData.MasterOfElements.FractionAt(mage.Talents.MasterOfElements)
	manaMetrics := mage.NewManaMetrics(core.ActionID{SpellID: 29077})

	trigger := spelldata.ProcTrigger(&mage.Character,
		spellData.MasterOfElements.Rank(mage.Talents.MasterOfElements),
		func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			mage.AddMana(sim, float64(spell.Cost.BaseCost)*refundFraction, manaMetrics)
		})
	trigger.Name = "Master of Elements - Trigger"

	trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		if !spell.SpellSchool.Matches(core.SpellSchoolFire | core.SpellSchoolFrost) {
			return false
		}
		return spell.Cost != nil && spell.Cost.BaseCost > 0
	}

	mage.MakeProcTriggerAura(trigger)
}

func (mage *Mage) registerCriticalMass() {
	if mage.Talents.CriticalMass == 0 {
		return
	}

	spelldata.ParseStatic(&mage.Character, spellData.CriticalMass.Rank(mage.Talents.CriticalMass))
}

// The most Fire effects Fire Power counts, from its tooltip's "up to a maximum increase of ${$s1*4}%".
const firePowerMaxEffects = 4

// "Increases the damage done by your Fire spells by 1/2/3% per your Fire effects on the target, up to
// a maximum increase of 12%. Effects which mix multiple schools of magic do not qualify."
//
// Both effects are dummies, so the count is by hand: Ignite, Improved Scorch and the Fireball and
// Pyroblast DoTs, each counted once. Impact's stun is not modelled, so it never counts, and the
// Frostfire Bolt DoT mixes schools; Frostfire Bolt still takes the bonus as a Fire spell. Ignite
// ignores every modifier, so it takes none.
func (mage *Mage) registerFirePower() {
	if mage.Talents.FirePower == 0 {
		return
	}

	perEffect := spellData.FirePower.Rank(mage.Talents.FirePower).EffectN(1).Percent()

	var fireEffects []*core.Spell
	mage.Env.RegisterPreFinalizeEffect(func() {
		for _, spell := range mage.Spellbook {
			if spell.Matches(MageSpellIgnite | MageSpellFireballDot | MageSpellPyroblastDot) {
				fireEffects = append(fireEffects, spell)
			}
		}
	})

	mage.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return core.MakePermanent(target.RegisterAura(core.Aura{
			Label: "Fire Power - " + mage.Label,
		})).AttachDDBC(DDBC_FirePower, DDBC_Total, &mage.AttackTables, func(_ *core.Simulation, spell *core.Spell, _ *core.AttackTable) float64 {
			if !spell.ProcMask.Matches(core.ProcMaskSpellDamage) || !spell.SpellSchool.Matches(core.SpellSchoolFire) {
				return 1
			}

			count := 0
			if mage.ImprovedScorchAuras.Get(target).IsActive() {
				count++
			}
			for _, effect := range fireEffects {
				if effect.Dot(target).IsActive() {
					count++
				}
			}

			return 1 + perEffect*float64(min(count, firePowerMaxEffects))
		})
	})
}

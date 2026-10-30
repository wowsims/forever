package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func (warlock *Warlock) registerDemonologyTalents() {
	// Tier 1
	warlock.applyImprovedHealthFunnel()
	warlock.applyImprovedImp()
	warlock.applyDemonicEmbrace()
	warlock.applyUnholyPower()

	// Tier 2
	// Demonic Aegis implemented in armors.go
	warlock.applyImprovedVoidwalker()
	warlock.applyFelVitality()
	warlock.applyDemonicEnergies()

	// Tier 3
	warlock.applyImprovedSayaad()
	warlock.applyDemonicSacrifice()
	warlock.applyMasterSummoner()

	// Tier 4
	warlock.applyDecimation()
	warlock.applyFelDomination()
	warlock.applyDemonicBrand()

	// Tier 5
	warlock.applyImprovedFelhunter()
	warlock.applySoulLink()
	warlock.applyDemonicKnowledge()

	// Tier 6
	warlock.applyMasterDemonologist()

	// Tier 7
	warlock.applyDemonicPact()
}

// --- Tier 1 ---

// TODO: Not implemented, as it does not affect sims for now: Health Funnel is not modelled.
func (warlock *Warlock) applyImprovedHealthFunnel() {
	if warlock.Talents.ImprovedHealthFunnel == 0 {
		return
	}
}

// Effect 2 raises Firebolt, on the Imp. Effect 1 (Fire Shield) and the dummy effect 3 do not affect sims.
func (warlock *Warlock) applyImprovedImp() {
	if warlock.Talents.ImprovedImp == 0 || warlock.Imp == nil {
		return
	}

	spelldata.ParseStatic(&warlock.Imp.Character, spellData.ImprovedImp.Rank(warlock.Talents.ImprovedImp), spelldata.Effects(2))
}

// Effect 1 raises total Stamina, as the tooltip says.
func (warlock *Warlock) applyDemonicEmbrace() {
	if warlock.Talents.DemonicEmbrace == 0 {
		return
	}

	spelldata.ParseStatic(&warlock.Character, spellData.DemonicEmbrace.Rank(warlock.Talents.DemonicEmbrace), spelldata.Effects(1))
}

// All damage the demons deal. The row's modifier names no spell the sim registers, so it is the pet's own
// damage multiplier; the Firebolt tooltip's $?s18769 reads this talent too.
func (warlock *Warlock) applyUnholyPower() {
	if warlock.Talents.UnholyPower == 0 {
		return
	}

	unholyPowerMult := 1 + spellData.UnholyPower.Rank(warlock.Talents.UnholyPower).EffectN(1).Percent()
	for _, pet := range warlock.demons() {
		pet.PseudoStats.DamageDealtMultiplier *= unholyPowerMult
	}
}

// --- Tier 2 ---

// TODO: Not implemented, as it does not affect sims for now: it raises the Voidwalker's threat and defensive
// abilities, which are not modelled.
func (warlock *Warlock) applyImprovedVoidwalker() {
	if warlock.Talents.ImprovedVoidwalker == 0 {
		return
	}
}

// The demons' maximum health and mana (effect 1), and the warlock's maximum mana (effect 2). The parse skips
// both, so they are applied here.
func (warlock *Warlock) applyFelVitality() {
	if warlock.Talents.FelVitality == 0 {
		return
	}

	rank := spellData.FelVitality.Rank(warlock.Talents.FelVitality)
	petMult, warlockMult := 1+rank.EffectN(1).Percent(), 1+rank.EffectN(2).Percent()

	for _, pet := range warlock.demons() {
		pet.MultiplyStat(stats.Health, petMult)
		pet.MultiplyStat(stats.Mana, petMult)
	}
	warlock.MultiplyStat(stats.Mana, warlockMult)
}

// The demon gains a share of the mana Life Tap returns (lifetap.go). Healing the demon for a share of the
// warlock's spell damage does not affect sims.
func (warlock *Warlock) applyDemonicEnergies() {
	if warlock.Talents.DemonicEnergies == 0 {
		return
	}

	warlock.lifeTapPetShare = spellData.DemonicEnergies.Rank(warlock.Talents.DemonicEnergies).EffectN(2).Percent()
}

// --- Tier 3 ---

// Raises Lash of Pain, on the Succubus. Effect 2 (Seduction's duration) does not affect sims.
func (warlock *Warlock) applyImprovedSayaad() {
	if warlock.Talents.ImprovedSayaad == 0 || warlock.Succubus == nil {
		return
	}

	spelldata.ParseStatic(&warlock.Succubus.Character, spellData.ImprovedSayaad.Rank(warlock.Talents.ImprovedSayaad))
}

// The sacrificed demon grants its effect until a summon cancels it (summonDemon): the Imp 15% more
// Shadow damage (18789), the Succubus 15% more Fire damage (18791), the Voidwalker mana every 4 sec (18792),
// the Felhunter health every 4 sec (18790), which does not affect sims. The demon sacrificed before the
// fight (settled in resolveSacrifice) is sacrificed at the start; casting the talent's spell (18788)
// sacrifices the demon out. The row states no cast time, cost or cooldown.
func (warlock *Warlock) applyDemonicSacrifice() {
	if !warlock.Talents.DemonicSacrifice {
		return
	}

	warlock.DemonicSacrificeAuras = map[proto.WarlockOptions_Summon]*core.Aura{}
	for _, effect := range []struct {
		summon proto.WarlockOptions_Summon
		id     int32
	}{
		{proto.WarlockOptions_Imp, 18789},
		{proto.WarlockOptions_Felhunter, 18790},
		{proto.WarlockOptions_Succubus, 18791},
		{proto.WarlockOptions_Voidwalker, 18792},
	} {
		summon := effect.summon
		effectRow := spellData.DemonicSacrificeTriggered.ByID(effect.id)
		auraConfig := spelldata.AuraConfig(effectRow)
		auraConfig.Duration = core.NeverExpires
		// Auras reset after the reset effects, so the demon sacrificed before the fight is marked here.
		if summon == warlock.Options.SacrificedDemon {
			auraConfig.OnReset = func(aura *core.Aura, sim *core.Simulation) {
				warlock.sacrificedDemon = summon
				aura.Activate(sim)
			}
		}
		aura := warlock.RegisterAura(auraConfig)
		spelldata.ParseEffects(&warlock.Character, aura, effectRow)
		warlock.DemonicSacrificeAuras[summon] = aura

		// Fel Energy's A_OBS_MOD_POWER: a share of maximum mana every period while the aura is up,
		// which the parse skips.
		if summon == proto.WarlockOptions_Voidwalker {
			restore := effectRow.EffectN(1)
			manaMetrics := warlock.NewManaMetrics(core.ActionID{SpellID: effectRow.ID})
			var restoreAction *core.PendingAction
			aura.ApplyOnGain(func(_ *core.Aura, sim *core.Simulation) {
				restoreAction = core.StartPeriodicAction(sim, core.PeriodicActionOptions{
					Period: restore.Period(),
					OnAction: func(sim *core.Simulation) {
						warlock.AddMana(sim, warlock.MaxMana()*restore.Percent(), manaMetrics)
					},
				})
			}).ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
				restoreAction.Cancel(sim)
			})
		}
	}

	warlock.RegisterResetEffect(func(_ *core.Simulation) {
		warlock.sacrificedDemon = proto.WarlockOptions_NoSummon
	})

	config := spelldata.SpellConfig(&warlock.Unit, spellData.DemonicSacrifice.ByID(18788), spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return warlock.ActivePet != nil && warlock.ActivePet.IsEnabled()
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		pet := warlock.ActivePet
		// Before the dismissal, so a summon's stun ending does not set the sacrificed demon's swings going again
		pet.summonStun.Deactivate(sim)
		pet.Disable(sim)
		warlock.ActivePet = nil

		warlock.DemonicSacrificeAuras[warlock.sacrificedDemon].Deactivate(sim)
		warlock.sacrificedDemon = pet.summon
		warlock.DemonicSacrificeAuras[pet.summon].Activate(sim)
	}
	warlock.DemonicSacrifice = warlock.RegisterSpell(config)
}

// Shortens the summons' cast (effect 1) and lowers their cost (effect 2), through the class flags the
// summon rows share.
func (warlock *Warlock) applyMasterSummoner() {
	if warlock.Talents.MasterSummoner == 0 {
		return
	}

	spelldata.ParseStatic(&warlock.Character, spellData.MasterSummoner.Rank(warlock.Talents.MasterSummoner))
}

// --- Tier 4 ---

// Effect 2 shortens Soul Fire's cooldown. Below 35% health (effect 3), Shadow Bolt and Searing Pain deal more
// damage (effect 4), and a cast of either that lands grants Decimation (440873): for 10 sec Soul Fire casts
// faster and, by the buff's A_NO_REAGENT_USE, takes no Soul Shard.
func (warlock *Warlock) applyDecimation() {
	if warlock.Talents.Decimation == 0 {
		return
	}

	rank := spellData.Decimation.Rank(warlock.Talents.Decimation)
	spelldata.ParseStatic(&warlock.Character, rank, spelldata.Effects(2))

	// The execute bonus, up from the moment the fight reaches 35%. It has no ActionID, so the rotation editor
	// lists only the buff (440873) as Decimation's aura.
	executeBonus := rank.EffectN(4)
	executeAura := warlock.RegisterAura(core.Aura{
		Label:    "Decimation - Execute",
		Duration: core.NeverExpires,
	})
	spelldata.ParseEffects(&warlock.Character, executeAura, rank, spelldata.Effects(4))

	threshold := int32(rank.EffectN(3).BasePoints)
	warlock.RegisterResetEffect(func(sim *core.Simulation) {
		sim.RegisterExecutePhaseCallback(func(sim *core.Simulation, isExecute int32) {
			if isExecute <= threshold && !executeAura.IsActive() {
				executeAura.Activate(sim)
			}
		})
	})

	// The buff's row states the 2/2 value, -40%; the talent's effect 1 prices the rank (-20% at 1/2), so the
	// buff takes the row's duration and Soul Fire's flags, and the rank's value.
	decimationRow := spellData.DecimationTriggered.ByID(440873)
	decimationAura := warlock.RegisterAura(spelldata.AuraConfig(decimationRow)).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_CastTime_Pct,
		ClassFlags: decimationRow.EffectN(1).ClassFlags,
		FloatValue: rank.EffectN(1).Percent(),
	})
	noReagent := decimationRow.Effect(dbcenums.A_NO_REAGENT_USE, 0)
	warlock.reagentWaivers = append(warlock.reagentWaivers, reagentWaiver{aura: decimationAura, flags: noReagent.ClassFlags})

	// The buff comes with the cast, and only when its hit lands: a miss or a full resist grants nothing.
	// Shadow Bolt and Searing Pain roll their hit at the cast, so the trigger reads its cast result, before a
	// bolt lands.
	warlock.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Decimation - Trigger",
		Callback:       core.CallbackOnCastResult,
		Outcome:        core.OutcomeLanded,
		ClassFlags:     executeBonus.ClassFlags,
		ClassSpellMask: WarlockSpellShadowBolt | WarlockSpellSearingPain,
		// In the logs the buff comes in the same batch as the cast.
		TriggerImmediately: true,
		ExtraCondition: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) bool {
			return sim.IsExecutePhase35()
		},
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			decimationAura.Activate(sim)
		},
	})
}

// For 15 sec, the next summon casts 5.5 sec faster (effect 1) for half the mana (effect 2), and spends
// the buff.
func (warlock *Warlock) applyFelDomination() {
	if !warlock.Talents.FelDomination {
		return
	}

	rankedSpell := spellData.FelDomination.ByID(18708)
	summons := rankedSpell.EffectN(1).ClassFlags

	auraConfig := spelldata.AuraConfig(rankedSpell)
	auraConfig.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if spell.MatchesFlags(summons) {
			aura.Deactivate(sim)
		}
	}
	warlock.FelDominationAura = warlock.RegisterAura(auraConfig)
	spelldata.ParseEffects(&warlock.Character, warlock.FelDominationAura, rankedSpell)

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Flags(core.SpellFlagAPL))
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		warlock.FelDominationAura.Activate(sim)
	}
	warlock.RegisterSpell(config)
}

// Searing Pain brands the target (1293696) for 10 sec; the demon's next attacks against it each consume a
// charge and deal the brand's damage (pets.go). Effect 1, Searing Pain's lower threat, is parsed.
// The brand row has no charges of its own; the ladder's effect 2, a charges modifier, gives them, 2/4/6 by
// rank: in the logs the brand lands with 6 at 3/3, though the tooltip reads 3.
func (warlock *Warlock) applyDemonicBrand() {
	if warlock.Talents.DemonicBrand == 0 {
		return
	}

	spelldata.ParseStatic(&warlock.Character, spellData.DemonicBrand.Rank(warlock.Talents.DemonicBrand), spelldata.Effects(1))
	charges := int32(spellData.DemonicBrand.Rank(warlock.Talents.DemonicBrand).EffectN(2).BaseValue())

	brandRow := spellData.DemonicBrandTriggered.ByID(1293696)
	warlock.DemonicBrandAuras = warlock.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		auraConfig := spelldata.AuraConfig(brandRow, spelldata.RankLabel(brandRow, &warlock.Unit))
		auraConfig.MaxStacks = charges
		return target.GetOrRegisterAura(auraConfig)
	})

	warlock.MakeProcTriggerAura(core.ProcTrigger{
		Name:     "Demonic Brand - Trigger",
		Callback: core.CallbackOnSpellHitDealt,
		Outcome:  core.OutcomeLanded,
		// In the logs the brand lands in the same batch as Searing Pain's hit.
		TriggerImmediately: true,
		ExtraCondition: func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
			return spellData.SearingPain.Contains(spell)
		},
		Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			brand := warlock.DemonicBrandAuras.Get(result.Target)
			brand.Activate(sim)
			brand.SetStacks(sim, charges)
		},
	})
}

// --- Tier 5 ---

// TODO: Not implemented, as it does not affect sims for now: it improves the Felhunter's utility abilities,
// which are not modelled.
func (warlock *Warlock) applyImprovedFelhunter() {
	if warlock.Talents.ImprovedFelhunter == 0 {
		return
	}
}

// While a demon is out, it and the warlock deal more damage (25228, effect 1). The share of damage
// the demon takes instead does not affect sims.
func (warlock *Warlock) applySoulLink() {
	if !warlock.Talents.SoulLink {
		return
	}

	soulLinkMult := 1 + spellData.SoulLinkTriggered.ByID(25228).EffectN(1).Percent()

	warlock.addPetActiveAura(warlock.RegisterAura(core.Aura{
		Label:    "Soul Link",
		ActionID: core.ActionID{SpellID: spellData.SoulLink.ByID(19028).ID},
		Duration: core.NeverExpires,
	}).AttachMultiplicativePseudoStatBuff(&warlock.PseudoStats.DamageDealtMultiplier, soulLinkMult))

	for _, pet := range warlock.demons() {
		pet.PseudoStats.DamageDealtMultiplier *= soulLinkMult
	}
}

// While a demon is out, the warlock and the demon gain spell damage of up to the rank's share of the
// warlock's level. The row's value is a dummy only the tooltip reads.
func (warlock *Warlock) applyDemonicKnowledge() {
	if warlock.Talents.DemonicKnowledge == 0 {
		return
	}

	rank := spellData.DemonicKnowledge.Rank(warlock.Talents.DemonicKnowledge)
	spellDamage := float64(warlock.Level) * rank.EffectN(1).Percent()

	warlock.DemonicKnowledgeAura = warlock.RegisterAura(core.Aura{
		Label:    "Demonic Knowledge",
		ActionID: core.ActionID{SpellID: rank.ID},
		Duration: core.NeverExpires,
	}).AttachStatBuff(stats.SpellDamage, spellDamage)
	warlock.addPetActiveAura(warlock.DemonicKnowledgeAura)

	for _, pet := range warlock.demons() {
		pet.AddStat(stats.SpellDamage, spellDamage)
	}
}

// --- Tier 6 ---

// Grants the warlock and the demon out an effect by demon, its four dummies read by the tooltip: the Imp
// more Fire damage (effect 1), the Voidwalker less Physical damage taken (effect 2), the Succubus more Shadow
// damage (effect 3), the Felhunter less magic damage taken (effect 4). The warlock's share lasts while
// that demon is out; the demon's own is always on it, as it acts only while out.
func (warlock *Warlock) applyMasterDemonologist() {
	if warlock.Talents.MasterDemonologist == 0 {
		return
	}

	rank := spellData.MasterDemonologist.Rank(warlock.Talents.MasterDemonologist)
	for _, pet := range warlock.demons() {
		warlock.applyMasterDemonologistFor(pet, rank)
	}
}

func (warlock *Warlock) applyMasterDemonologistFor(pet *WarlockPet, rank *spelldata.Spell) {
	var schools []stats.SchoolIndex
	var amount float64
	dealt := true
	switch pet.summon {
	case proto.WarlockOptions_Imp:
		schools, amount = []stats.SchoolIndex{stats.SchoolIndexFire}, rank.EffectN(1).Percent()
	case proto.WarlockOptions_Voidwalker:
		schools, amount, dealt = []stats.SchoolIndex{stats.SchoolIndexPhysical}, rank.EffectN(2).Percent(), false
	case proto.WarlockOptions_Succubus:
		schools, amount = []stats.SchoolIndex{stats.SchoolIndexShadow}, rank.EffectN(3).Percent()
	case proto.WarlockOptions_Felhunter:
		schools, amount, dealt = []stats.SchoolIndex{stats.SchoolIndexArcane, stats.SchoolIndexFire, stats.SchoolIndexFrost,
			stats.SchoolIndexHoly, stats.SchoolIndexNature, stats.SchoolIndexShadow}, rank.EffectN(4).Percent(), false
	default:
		return
	}

	multipliers := func(unit *core.Unit) []*float64 {
		fields := make([]*float64, len(schools))
		for i, school := range schools {
			if dealt {
				fields[i] = &unit.PseudoStats.SchoolDamageDealtMultiplier[school]
			} else {
				fields[i] = &unit.PseudoStats.SchoolDamageTakenMultiplier[school]
			}
		}
		return fields
	}

	aura := warlock.RegisterAura(core.Aura{
		Label:    "Master Demonologist - " + pet.Name,
		ActionID: core.ActionID{SpellID: rank.ID, Tag: int32(pet.summon)},
		Duration: core.NeverExpires,
	})
	for _, field := range multipliers(&warlock.Unit) {
		aura.AttachMultiplicativePseudoStatBuff(field, 1+amount)
	}
	pet.ownerAuras = append(pet.ownerAuras, aura)

	for _, field := range multipliers(&pet.Unit) {
		*field *= 1 + amount
	}
}

// --- Tier 7 ---

// Keeps Demonic Sacrifice's effect through a summon of another demon: resolveSacrifice and summonDemon
// read the talent.
func (warlock *Warlock) applyDemonicPact() {
	if !warlock.Talents.DemonicPact {
		return
	}
}

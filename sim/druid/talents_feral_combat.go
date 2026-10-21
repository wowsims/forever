package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The feral tree in its own order. Feral Charge, Shifting Power and Primal Bite are spells whose cast
// condition is the talent; registerFeralKit registers them. AddRaidBuffs publishes Leader of the
// Pack, and the melee specials read Rend and Tear when they deal damage.
func (druid *Druid) registerFeralCombatTalents() {
	druid.applyFerocity()
	druid.applyHeartOfTheWild()
	druid.applyFeralSwiftness()
	druid.applyFeralInstinct()
	druid.applyBrutalImpact()
	druid.applyThickHide()
	druid.applyShreddingAttacks()
	druid.applySavageFury()
	druid.applySharpenedClaws()
	druid.applyPredatoryStrikes()
	druid.applyBloodFrenzy()
	druid.applyImprovedShiftingPower()
	druid.applyNaturalInstinct()
	druid.applyNaturalReaction()
	druid.applyBerserk()
}

func (druid *Druid) parseTalent(rank int32, ladder spelldata.Ladder) {
	if rank == 0 {
		return
	}
	spelldata.ParseStatic(&druid.Character, ladder.Rank(rank))
}

func (druid *Druid) parseOnFeralForms(rank int32, ladder spelldata.Ladder) {
	if rank == 0 {
		return
	}
	row := ladder.Rank(rank)
	for _, aura := range []*core.Aura{druid.CatFormAura, druid.BearFormAura, druid.DireBearFormAura} {
		if aura != nil {
			spelldata.ParseEffects(&druid.Character, aura, row)
		}
	}
}

func (druid *Druid) applyFerocity() {
	druid.parseTalent(druid.Talents.Ferocity, spellData.Ferocity)
}

func (druid *Druid) applyFeralInstinct() {
	druid.parseOnFeralForms(druid.Talents.FeralInstinct, spellData.FeralInstinct)
}

func (druid *Druid) applyBrutalImpact() {
	druid.parseTalent(druid.Talents.BrutalImpact, spellData.BrutalImpact)
}

func (druid *Druid) applyThickHide() {
	if druid.Talents.ThickHide == 0 {
		return
	}
	row := spellData.ThickHide.Rank(druid.Talents.ThickHide)
	// Effect 2 prices armor as the level times its base points: 60, 120 or 180 at level 60. Effect 1 /
	// 100 is base armor per point of defense skill beyond five times the level, which is the defense the
	// gear and buffs add (one rating is one skill point). Both are base armor, which the bear forms scale.
	// A level 30 character measured 30, 60 and 90 armor from ranks 1 to 3 in Cat Form and 84, 168 and 252 in Bear
	// Form, so ranks 1 and 2 follow the level times the base points, not the game files' per-level formula
	// (178 and 179 at level 60).
	bonus := float64(druid.level()) * row.EffectN(2).BaseValue()
	perDefense := row.EffectN(1).BaseValue() / 100 / core.DefenseRatingPerDefenseLevel
	for _, form := range []struct {
		aura       *core.Aura
		multiplier float64
	}{
		{druid.CatFormAura, 1},
		{druid.BearFormAura, 1 + spellData.BearFormTriggered.Highest().EffectN(2).Percent()},
		{druid.DireBearFormAura, 1 + spellData.DireBearFormTriggered.Highest().EffectN(2).Percent()},
		{druid.MoonkinFormAura, 1 + spellData.MoonkinFormTriggered.ByID(24905).EffectN(1).Percent()},
	} {
		if form.aura != nil {
			form.aura.AttachStatBuff(stats.Armor, bonus*form.multiplier)
			form.aura.AttachStatDependency(druid.NewDynamicStatDependency(stats.DefenseRating, stats.Armor, perDefense*form.multiplier))
		}
	}
}

func (druid *Druid) applyFeralSwiftness() {
	druid.parseTalent(druid.Talents.FeralSwiftness, spellData.FeralSwiftness)
	if druid.Talents.FeralSwiftness > 0 && druid.CatFormAura != nil {
		druid.CatFormAura.NewPassiveMovementSpeedEffect(spellData.FeralSwiftness.EffectAt(2).FractionAt(druid.Talents.FeralSwiftness))
	}
}

func (druid *Druid) applySharpenedClaws() {
	druid.parseOnFeralForms(druid.Talents.SharpenedClaws, spellData.SharpenedClaws)
}

func (druid *Druid) applyShreddingAttacks() {
	druid.parseTalent(druid.Talents.ShreddingAttacks, spellData.ShreddingAttacks)
}

func (druid *Druid) applyImprovedShiftingPower() {
	druid.parseTalent(druid.Talents.ImprovedShiftingPower, spellData.ImprovedShiftingPower)
}

func (druid *Druid) applyPredatoryStrikes() {
	if druid.Talents.PredatoryStrikes == 0 {
		return
	}
	// The dummy is a percent of level as attack power, and only while shapeshifted. 150 at 3/3 is 90 at 60.
	ap := float64(druid.level()) * spellData.PredatoryStrikes.Rank(druid.Talents.PredatoryStrikes).EffectN(1).Percent()
	for _, aura := range []*core.Aura{druid.CatFormAura, druid.BearFormAura, druid.DireBearFormAura} {
		if aura != nil {
			aura.AttachStatBuff(stats.AttackPower, ap)
		}
	}
}

func (druid *Druid) applySavageFury() {
	druid.parseTalent(druid.Talents.SavageFury, spellData.SavageFury)
}

func (druid *Druid) applyHeartOfTheWild() {
	if druid.Talents.HeartOfTheWild == 0 {
		return
	}
	// Effect 1 is the Intellect ($s1), which the parse reads from its stat mask. The bear Stamina ($s2)
	// and cat Strength ($s3) are dummies the forms carry.
	druid.parseTalent(druid.Talents.HeartOfTheWild, spellData.HeartOfTheWild)
	row := spellData.HeartOfTheWild.Rank(druid.Talents.HeartOfTheWild)
	strength := druid.NewDynamicMultiplyStat(stats.Strength, 1+row.EffectN(3).Percent())
	if druid.CatFormAura != nil {
		druid.CatFormAura.AttachStatDependency(strength)
	}
	stamina := druid.NewDynamicMultiplyStat(stats.Stamina, 1+row.EffectN(2).Percent())
	for _, aura := range []*core.Aura{druid.BearFormAura, druid.DireBearFormAura} {
		if aura != nil {
			aura.AttachStatDependency(stamina)
		}
	}
}

func (druid *Druid) applyNaturalInstinct() {
	druid.parseTalent(druid.Talents.NaturalInstinct, spellData.NaturalInstinct)
}

func (druid *Druid) applyBloodFrenzy() {
	if druid.Talents.BloodFrenzy == 0 {
		return
	}
	row := spellData.BloodFrenzy.Rank(druid.Talents.BloodFrenzy)
	rageChance := row.EffectN(1).Percent()  // $m1, the rage in Bear and Dire Bear Form
	comboChance := row.EffectN(2).Percent() // $m2, the combo point in Cat Form
	rageSpell := spellData.BloodFrenzyTriggered.Highest()
	metrics := druid.NewRageMetrics(core.ActionID{SpellID: rageSpell.ID})
	combo := druid.NewComboPointMetrics(core.ActionID{SpellID: row.ID})
	druid.MakeProcTriggerAura(spelldata.ProcTrigger(&druid.Character, row, func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		if !result.Outcome.Matches(core.OutcomeCrit) {
			return
		}
		if druid.InForm(Bear) {
			if sim.Proc(rageChance, "Blood Frenzy Rage") {
				druid.AddRage(sim, rageSpell.EnergizeEffect().Tenths(), metrics)
			}
			return
		}
		// TODO: Ingame research needed -- the triggered spell restores rage in bear form. The cat's
		// extra combo point on a non-periodic builder crit is the tooltip, not an effect on the row.
		if druid.InForm(Cat) && druid.isComboBuilder(spell) && sim.Proc(comboChance, "Blood Frenzy Combo Point") {
			druid.AddComboPoints(sim, 1, combo)
		}
	}, spelldata.Chance(1)))
}

func (druid *Druid) applyNaturalReaction() {
	druid.parseTalent(druid.Talents.NaturalReaction, spellData.NaturalReaction)
	if druid.Talents.NaturalReaction == 0 {
		return
	}
	row := spellData.NaturalReaction.Rank(druid.Talents.NaturalReaction)
	rageSpell := spellData.NaturalReactionTriggered.Highest()
	metrics := druid.NewRageMetrics(core.ActionID{SpellID: rageSpell.ID})
	// The proc mask names every hit taken and has no dodge bit, so the row's trigger listens to landed
	// hits; the tooltip's "each time you dodge" is the outcome. Every rank restores the triggered
	// spell's 5 rage, and the rank sets the chance ($m2).
	trigger := spelldata.ProcTrigger(&druid.Character, row, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		if druid.InForm(Bear) {
			druid.AddRage(sim, rageSpell.EnergizeEffect().Tenths(), metrics)
		}
	})
	trigger.Outcome = core.OutcomeDodge
	druid.MakeProcTriggerAura(trigger)
}

func (druid *Druid) applyBerserk() {
	if !druid.Talents.Berserk {
		return
	}
	row := spellData.Berserk.Highest()
	druid.BerserkAura = druid.RegisterAura(spelldata.AuraConfig(row))
	// Crit on the builder mask and Primal Bite's cooldown. Jump targets and fear immunity are not parsed.
	spelldata.ParseEffects(&druid.Character, druid.BerserkAura, row)
	// TODO: Ingame confirmation needed -- whether Berserk clears a running Primal Bite cooldown
	druid.BerserkAura.ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) {
		if druid.PrimalBite != nil {
			druid.PrimalBite.CD.Reset()
		}
	})
	config := spelldata.SpellConfig(&druid.Unit, row)
	config.Flags |= core.SpellFlagAPL
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		druid.BerserkAura.Activate(sim)
	}
	config.RelatedSelfBuff = druid.BerserkAura
	druid.Berserk = druid.RegisterSpell(config)
	druid.AddMajorCooldown(core.MajorCooldown{
		Spell: druid.Berserk.Spell,
		Type:  core.CooldownTypeDPS,
	})
}

package shaman

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (shaman *Shaman) registerRestorationTalents() {
	// Tier 1
	shaman.applyImprovedHealingWave()
	shaman.applyTidalFocus()

	// Tier 2
	shaman.applyMindfulness()
	shaman.applyNaturalGrace()
	shaman.applyTotemicFocus()
	shaman.applyImprovedReincarnation()

	// Tier 3
	// Ancestral Healing (16176) raises the healed target's armor through Ancestral Fortitude
	// (16177), 8/17/25% by rank.
	//
	// TODO: To be implemented as a buff on the receiver, the way Inspiration is in core/buffs.
	// Healing Focus (16181) reduces pushback only, which the sim does not model.
	// Water Shield (408510): shields.go, registerWaterShieldSpell

	// Tier 4
	shaman.applyTidalMastery()
	// Restorative Totems (16187): effect 1, Mana Spring Totem's, is read by its aura in totems.go.
	//
	// TODO: Effect 2, Healing Stream Totem's, waits on that totem.
	shaman.applyManaTideTotem()

	// Tier 5
	shaman.applyHealingWay()
	shaman.applyNaturesSwiftness()

	// Tier 6
	shaman.applyPurification()

	// Tier 7
	shaman.applyRiptide()
}

func (shaman *Shaman) applyImprovedHealingWave() {
	if shaman.Talents.ImprovedHealingWave == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.ImprovedHealingWave.Rank(shaman.Talents.ImprovedHealingWave))
}

func (shaman *Shaman) applyTotemicFocus() {
	if shaman.Talents.TotemicFocus == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.TotemicFocus.Rank(shaman.Talents.TotemicFocus))
}

func (shaman *Shaman) applyMindfulness() {
	if shaman.Talents.Mindfulness == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.Mindfulness.Rank(shaman.Talents.Mindfulness))
}

func (shaman *Shaman) applyNaturalGrace() {
	if shaman.Talents.NaturalGrace == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.NaturalGrace.Rank(shaman.Talents.NaturalGrace))
}

func (shaman *Shaman) applyTidalFocus() {
	if shaman.Talents.TidalFocus == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.TidalFocus.Rank(shaman.Talents.TidalFocus))
}

// Effect 3, the health and mana Reincarnation restores, is skipped by the parse.
func (shaman *Shaman) applyImprovedReincarnation() {
	if shaman.Talents.ImprovedReincarnation == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.ImprovedReincarnation.Rank(shaman.Talents.ImprovedReincarnation))
}

func (shaman *Shaman) applyTidalMastery() {
	if shaman.Talents.TidalMastery == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.TidalMastery.Rank(shaman.Talents.TidalMastery))
}

var manaTideTotemRank = spellData.ManaTideTotem.Highest()

// The talent teaches Mana Tide Totem, a water totem, so it takes down Mana Spring.
func (shaman *Shaman) applyManaTideTotem() {
	if !shaman.Talents.ManaTideTotem {
		return
	}

	manaTideTotem := shaman.registerBuffTotem(manaTideTotemRank, WaterTotem,
		buffs.NewSpellDataManaTideTotemsAura(&shaman.Unit, true, 0))

	shaman.AddMajorCooldown(core.MajorCooldown{
		Spell: manaTideTotem,
		Type:  core.CooldownTypeMana,
	})
}

func (shaman *Shaman) applyHealingWay() {
	if shaman.Talents.HealingWay == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.HealingWay.Rank(shaman.Talents.HealingWay))
}

var naturesSwiftnessRank = spellData.NaturesSwiftness.Highest()

func (shaman *Shaman) applyNaturesSwiftness() {
	if !shaman.Talents.NaturesSwiftness {
		return
	}

	// The row's mod makes the spells it reaches instant, and the next one cast spends the charge.
	consumes := naturesSwiftnessRank.EffectN(1).ClassFlags

	config := spelldata.AuraConfig(naturesSwiftnessRank)
	config.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if consumes.Reaches(spell) {
			aura.Deactivate(sim)
		}
	}
	nsAura := shaman.RegisterAura(config)
	spelldata.ParseEffects(&shaman.Character, nsAura, naturesSwiftnessRank)

	spellConfig := spelldata.SpellConfig(&shaman.Unit, naturesSwiftnessRank,
		spelldata.Flags(core.SpellFlagNoOnCastComplete))
	spellConfig.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		nsAura.Activate(sim)
	}
	spellConfig.RelatedSelfBuff = nsAura

	shaman.AddMajorCooldown(core.MajorCooldown{
		Spell: shaman.RegisterSpell(spellConfig),
		Type:  core.CooldownTypeDPS,
	})
}

func (shaman *Shaman) applyPurification() {
	if shaman.Talents.Purification == 0 {
		return
	}
	spelldata.ParseStatic(&shaman.Character, spellData.Purification.Rank(shaman.Talents.Purification))
}

// The talent teaches Riptide (408521).
//
// TODO: To be implemented with the healing spells.
func (shaman *Shaman) applyRiptide() {
	if !shaman.Talents.Riptide {
		return
	}
}

package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var combustionRank = spellData.Combustion.Highest()
var combustionBuffRank = spellData.CombustionTriggered.Highest()

// Combustion
// https://www.wowhead.com/forever/spell=11129
//
// "When activated, this spell causes each of your Fire damage spell hits to increase your critical
// strike chance with Fire damage spells by 10%. This effect lasts until you have caused 3
// non-periodic critical strikes with Fire spells."
//
// Two auras: 11129 counts the crits left in its stacks, and 28682 carries the crit.
func (mage *Mage) registerCombustionSpell() {
	buffAura := mage.RegisterAura(spelldata.AuraConfig(combustionBuffRank, spelldata.Label("Combustion Crit")))
	spelldata.ParseEffects(&mage.Character, buffAura, combustionBuffRank)

	var combustionSpell *core.Spell

	chargesConfig := spelldata.AuraConfig(combustionRank, spelldata.Label("Combustion"))

	// Effect 2 casts 28682 on activation, so the crit starts at one stack before any Fire hit.
	chargesConfig.OnGain = func(aura *core.Aura, sim *core.Simulation) {
		aura.SetStacks(sim, aura.MaxStacks)
		buffAura.Activate(sim)
		buffAura.SetStacks(sim, 1)
	}

	// The row is Disabled While Active (Attr0 0x2000000): the cooldown starts when the effect ends.
	chargesConfig.OnExpire = func(aura *core.Aura, sim *core.Simulation) {
		buffAura.Deactivate(sim)
		combustionSpell.CD.Use(sim)
		mage.UpdateMajorCooldowns()
	}

	// Area hits in EV-0197. Periodic damage never reaches OnSpellHitDealt, so DoT ticks and Ignite
	// neither stack nor spend. The row cannot proc from procs, so an item's Fire proc does neither.
	chargesConfig.OnSpellHitDealt = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		if !spell.ProcMask.Matches(core.ProcMaskSpellDamage) || !spell.SpellSchool.Matches(core.SpellSchoolFire) || !result.Landed() ||
			spell.Flags.Matches(core.SpellFlagProc) {
			return
		}

		buffAura.AddStack(sim)

		if result.DidCrit() {
			aura.RemoveStack(sim)
		}
	}

	chargesAura := mage.RegisterAura(chargesConfig)

	// Manaflare Regalia's 5pc: Frostfire Bolt gains effect 2's crit while Combustion is active.
	if mage.manaflareAura != nil {
		frostfireCrit := mage.AddDynamicMod(core.SpellModConfig{
			Kind:       core.SpellMod_BonusCrit_Percent,
			ClassMask:  MageSpellFrostfireBolt,
			FloatValue: manaflareFrostfireBolt.EffectN(2).BaseValue(),
		})
		chargesAura.ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) {
			if mage.manaflareAura.IsActive() {
				frostfireCrit.Activate()
			}
		}).ApplyOnExpire(func(_ *core.Aura, _ *core.Simulation) {
			frostfireCrit.Deactivate()
		})
	}

	// An off-GCD cast of its own.
	config := spelldata.SpellConfig(&mage.Unit, combustionRank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagNoOnCastComplete))
	config.ClassSpellMask = MageSpellCombustion

	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return !chargesAura.IsActive()
	}

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		chargesAura.Activate(sim)
	}

	config.RelatedSelfBuff = chargesAura

	combustionSpell = mage.RegisterSpell(config)

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: combustionSpell,
		Type:  core.CooldownTypeDPS,
	})
}

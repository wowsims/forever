package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Causes an explosion of holy light around the caster, causing X to Y Holy damage to all enemy targets within 10 yards and
// healing all party members within 10 yards for A to B.
// These effects cause no threat. Each time Holy Fire deals damage, you have a 5% chance for your next Holy Nova to cost no Mana.
func (priest *Priest) registerHolyNovaSpell(_ int32, rank *spelldata.Spell) {
	baseDamage := rank.DamageEffect()
	healRow := rank.Refs()[0]
	heal := healRow.HealEffect()

	healConfig := spelldata.SpellConfig(&priest.Unit, healRow, spelldata.Magic(core.ProcMaskSpellHealing), spelldata.Proc())
	healConfig.ClassSpellMask = PriestSpellHolyNova
	healConfig.ThreatMultiplier = 0
	healSpell := priest.RegisterSpell(healConfig)

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ThreatMultiplier = 0 // The tooltip: "These effects cause no threat." No row field states it.
	config.ClassSpellMask = PriestSpellHolyNova

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamage(sim, baseDamage.Average(), spell.OutcomeMagicHitAndCrit)

		if priest.Party == nil || priest.HolyNovaHealTargets <= 0 {
			return
		}

		targets := make([]*core.Unit, 0, len(priest.Party.Players))
		for _, member := range priest.Party.Players {
			targets = append(targets, &member.GetCharacter().Unit)
		}
		healCount := min(int(priest.HolyNovaHealTargets), len(targets))
		for i := 0; i < healCount; i++ {
			j := i + int(sim.RandomFloat("Holy Nova heal target selection")*float64(len(targets)-i))
			targets[i], targets[j] = targets[j], targets[i]
		}
		for _, target := range targets[:healCount] {
			healSpell.CalcAndDealHealing(sim, target, heal.Average(), healSpell.OutcomeHealingCrit)
		}
	}

	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, baseDamage.Average(), spell.OutcomeExpectedMagicCrit)
	}

	priest.RegisterSpell(config)
}

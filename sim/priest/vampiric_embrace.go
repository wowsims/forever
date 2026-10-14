package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerVampiricEmbrace(_ int32, rank *spelldata.Spell) {
	healFraction := rank.EffectN(1).Percent()

	veAuras := priest.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.RegisterAura(spelldata.AuraConfig(rank, spelldata.RankLabel(rank, &priest.Unit)))
	})

	// The party heal 15290 the debuff casts, a heal with no base of its own on the caster's party.
	// TODO: Read the heal off row 15290 once the store carries it.
	heal := priest.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 15290},
		SpellSchool:      core.SpellSchoolShadow,
		DefenseType:      core.DefenseTypeMagic,
		ProcMask:         core.ProcMaskSpellHealing,
		Flags:            core.SpellFlagHelpful | core.SpellFlagPassiveSpell | core.SpellFlagNoOnCastComplete,
		DamageMultiplier: 1,
		ThreatMultiplier: 1, // TODO: Verify in game whether the heal causes threat.
	})

	priest.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Vampiric Embrace Heal",
		Callback:           core.CallbackOnSpellHitDealt | core.CallbackOnPeriodicDamageDealt,
		ClassSpellMask:     PriestShadowSpells, // The tooltip's "Shadow spell damage you deal"; the row's listener names no spells.
		RequireDamageDealt: true,
		CanProcFromProcs:   true,
		ExtraCondition: func(_ *core.Simulation, _ *core.Spell, result *core.SpellResult) bool {
			return veAuras.Get(result.Target).IsActive()
		},
		// Each party member's heal is a heal of its own, which on-heal listeners hear one by one.
		// TODO: Verify in game that each member's heal rolls Revelation separately, not once per damage event.
		Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			amount := result.Damage * healFraction
			for _, agent := range priest.Party.PlayersAndPets {
				member := &agent.GetCharacter().Unit
				if member.IsEnabled() {
					heal.CalcAndDealHealing(sim, member, amount, heal.OutcomeHealing)
				}
			}
		},
	})

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
		if result.Landed() {
			veAuras.Get(target).Activate(sim)
		}
	}
	config.RelatedAuraArrays = veAuras.ToMap()

	priest.VampiricEmbrace = priest.RegisterSpell(config)
}

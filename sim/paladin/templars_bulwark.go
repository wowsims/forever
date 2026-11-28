package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Templar's Bulwark (talent)
// https://www.wowhead.com/forever/spell=1311015
//
// When activated, this ability grants you an absorb shield equal to 100% of your maximum health
// for 8 sec. Applies Forbearance for 1 min. Cannot be cast while Forbearance is active.
func (paladin *Paladin) registerTemplarsBulwark() {
	rank := spellData.TemplarsBulwark.Highest()
	actionID := core.ActionID{SpellID: rank.ID}
	healthShare := rank.Effect(dbcenums.A_SCHOOL_ABSORB, 127).Percent()

	shield := paladin.NewDamageAbsorptionAura(core.AbsorptionAuraConfig{
		Aura: core.Aura{
			Label:    "Templar's Bulwark" + paladin.Label,
			ActionID: actionID,
			Duration: rank.Duration(),
		},
		ShieldStrengthCalculator: func(unit *core.Unit) float64 {
			return unit.MaxHealth() * healthShare
		},
	})

	// The row's excluded caster aura is Forbearance, so the cast requirement refuses it while
	// Forbearance is up.
	config := spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ClassSpellMask = SpellMaskTemplarsBulwark
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		shield.Activate(sim)
		paladin.Forbearance.Activate(sim)
	}
	config.RelatedSelfBuff = shield.Aura
	spell := paladin.RegisterSpell(config)

	paladin.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeSurvival,
	})
}

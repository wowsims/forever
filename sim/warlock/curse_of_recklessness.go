package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The debuff is this warlock's own copy of the one the raid option builds, parsed from the row: -505
// armor. Its category keeps it from stacking with another warlock's copy or a raid one. The row's dummy
// +90 attack power for the target and its fear immunity are left out, as the raid debuff leaves them out.
func (warlock *Warlock) registerCurseOfRecklessness() {
	rankedSpell := spellData.CurseOfRecklessness.ByID(11717)

	warlock.CurseOfRecklessnessAuras = warlock.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return buffs.NewSpellDataCurseOfRecklessnessAuraFor(target, &warlock.Unit)
	})

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ThreatMultiplier = 1
	config.ClassSpellMask = WarlockSpellCurseOfRecklessness
	config.RelatedAuraArrays = warlock.CurseOfRecklessnessAuras.ToMap()

	// Roll to hit and apply in place of any other Curse
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
		if result.Landed() {
			warlock.deactivateOtherCurses(sim, spell, target)
			warlock.CurseOfRecklessnessAuras.Get(target).Activate(sim)
		}
		spell.DealOutcome(sim, result)
	}

	warlock.CurseOfRecklessness = warlock.RegisterSpell(config)
	warlock.Curses = append(warlock.Curses, warlock.CurseOfRecklessness)
}

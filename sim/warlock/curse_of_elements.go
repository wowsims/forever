package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The debuff is this warlock's own copy of the one the raid option builds, parsed from the row: -75 to
// all resistances and 10% more damage taken. Its category keeps it from stacking with another warlock's
// copy or a raid one.
func (warlock *Warlock) registerCurseOfElements() {
	rankedSpell := spellData.CurseOfTheElements.ByID(1311680)

	warlock.CurseOfElementsAuras = warlock.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return buffs.NewSpellDataCurseOfElementsAuraFor(target, &warlock.Unit)
	})
	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ThreatMultiplier = 1
	config.ClassSpellMask = WarlockSpellCurseOfElements
	config.RelatedAuraArrays = warlock.CurseOfElementsAuras.ToMap()

	// Roll to hit and apply in place of any other Curse
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
		if result.Landed() {
			warlock.deactivateOtherCurses(sim, spell, target)
			warlock.CurseOfElementsAuras.Get(target).Activate(sim)
		}
		spell.DealOutcome(sim, result)
	}

	warlock.CurseOfElements = warlock.RegisterSpell(config)
	warlock.Curses = append(warlock.Curses, warlock.CurseOfElements)
}

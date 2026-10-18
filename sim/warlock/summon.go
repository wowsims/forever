package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The summons the APL casts, each by its own row: a 10 sec cast for a share of base mana, both of which
// Master Summoner and Fel Domination lower through the class flags the four rows share. The Voidwalker,
// Succubus and Felhunter take the Soul Shard their rows state; the Imp's row states none.
func (warlock *Warlock) registerSummons() {
	warlock.SummonImp = warlock.registerSummon(spellData.SummonImp.Highest(), warlock.Imp)
	warlock.SummonVoidwalker = warlock.registerSummon(spellData.SummonVoidwalker.Highest(), warlock.Voidwalker)
	warlock.SummonSuccubus = warlock.registerSummon(spellData.SummonSuccubus.Highest(), warlock.Succubus)
	warlock.SummonFelhunter = warlock.registerSummon(spellData.SummonFelhunter.Highest(), warlock.Felhunter)
}

func (warlock *Warlock) registerSummon(row *spelldata.Spell, pet *WarlockPet) *core.Spell {
	config := spelldata.SpellConfig(&warlock.Unit, row, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ClassSpellMask = WarlockSpellSummonDemon

	// The demon out is stunned from the start of the cast. The APL checks the summon can be cast first.
	config.Cast.ModifyCast = func(sim *core.Simulation, _ *core.Spell, _ *core.Cast) {
		if warlock.ActivePet != nil && warlock.ActivePet.IsEnabled() {
			core.ApplyStun(sim, warlock.ActivePet.summonStun)
		}
	}

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		warlock.summonDemon(sim, pet)
	}

	warlock.useReagents(&config, row)
	return warlock.RegisterSpell(config)
}

// Dismisses the demon out and brings in this one, the same demon included, which returns at full mana. Summoning cancels Demonic Sacrifice's effect: any
// demon without Demonic Pact, and with it only the demon sacrificed.
func (warlock *Warlock) summonDemon(sim *core.Simulation, pet *WarlockPet) {
	if warlock.ActivePet != nil {
		// Before the dismissal, so the stun ending does not set the dismissed demon's swings going again
		warlock.ActivePet.summonStun.Deactivate(sim)
		warlock.ActivePet.Disable(sim)
	}
	warlock.ActivePet = pet
	pet.Enable(sim, pet)
	for _, aura := range pet.partyAuras {
		aura.Activate(sim)
	}

	if sacrificed := warlock.sacrificedDemon; sacrificed != proto.WarlockOptions_NoSummon && (!warlock.Talents.DemonicPact || pet.summon == sacrificed) {
		warlock.DemonicSacrificeAuras[sacrificed].Deactivate(sim)
		warlock.sacrificedDemon = proto.WarlockOptions_NoSummon
	}
}

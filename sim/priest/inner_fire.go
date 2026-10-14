package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerInnerFire(_ int32, rank *spelldata.Spell) {
	auraConfig := spelldata.AuraConfig(rank, spelldata.RankLabel(rank, &priest.Unit))
	if rank == spellData.InnerFire.Highest() && priest.SelfBuffs.Armor == proto.PriestOptions_InnerFire {
		auraConfig.OnReset = func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
			aura.SetStacks(sim, aura.MaxStacks)
		}
	}
	aura := priest.RegisterAura(auraConfig)
	aura.NewExclusiveEffect(rank.Name, true, core.ExclusiveEffect{})

	// Improved Inner Fire states its armor bonus as SPELLMOD_ALL_EFFECTS, which the talent's own parse
	// files as damage done, so the armor is priced with it here. Its extra charges reach the aura as a
	// spell mod through RelatedSelfBuff.
	improvedInnerFire := spellData.ImprovedInnerFire.Rank(priest.Talents.ImprovedInnerFire).EffectN(1)
	spelldata.ParseEffects(&priest.Character, aura, rank, spelldata.ScaledBy(improvedInnerFire))

	aura.AttachProcTrigger(spelldata.ProcTrigger(&priest.Character, rank,
		func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			aura.RemoveStack(sim)
		}))

	priest.registerSelfBuffSpell(rank, aura)
}

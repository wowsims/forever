package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var ExorcismRankMap = spellData.Exorcism

// Exorcism
// https://www.wowhead.com/forever/spell=10314
//
// Causes 502 Holy damage to an Undead or Demon target.
//
// spellData.Exorcism holds the six trainer ranks, 879 to 10314. The game files carry a second ladder,
// 415068 to 415073, that the Season of Discovery passive Exorcist (415076) swaps onto the action
// bar so the spell can hit any target; nothing in Forever teaches Exorcist.
func (paladin *Paladin) registerExorcism(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rank.ID},
		SpellSchool:    rank.SpellSchool(),
		DefenseType:    rank.DefenseTypeCore(),
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskExorcism,
		Rank:           rank.RankNumber(),
		MaxRange:       float64(rank.MaxRange),

		ManaCost: manaCost(rank),
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: rank.GCD(),
			},
			CD: core.Cooldown{
				Timer:    paladin.sharedTimer(&paladin.exorcismTimer),
				Duration: cooldown(rank),
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: damage.Coeff(),

		ExtraCastCondition: func(_ *core.Simulation, target *core.Unit) bool {
			return target.MobType == proto.MobType_MobTypeUndead || target.MobType == proto.MobType_MobTypeDemon
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, damage.Roll(sim), spell.OutcomeMagicHitAndCrit)
		},
	})
}

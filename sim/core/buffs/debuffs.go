package buffs

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// The debuffs the raid puts on a target: the generated rows, then the one the game files state in a
// shape no manifest row can carry.
func applyDebuffs(target *core.Unit, debuffs *proto.Debuffs, raid *proto.Raid) {
	applyGeneratedDebuffs(target, debuffs, raid)

	if debuffs.JudgementOfTheCrusader {
		core.MakePermanent(NewSpellDataJudgementOfTheCrusaderAuraRank(target, JudgementOfTheCrusaderMaxRank))
	}
}

// The player's copy of the generated debuff that reads the spell, or nil where none does. An item
// proc whose debuff is a raid debuff lands this copy, which bids in the raid debuff's category.
func NewSpellDataDebuffAuraBySpell(target *core.Unit, spellID int32) *core.Aura {
	m := debuffMetas[spellID]
	if m == nil {
		return nil
	}
	return newDebuff(target, m, true, 0)
}

// A hunter's own Scorpid Sting, one per hunter, tagged with the tag its stings share.
func NewSpellDataScorpidStingAuraFor(target *core.Unit, hunter *core.Unit, stingTag string) *core.Aura {
	return newCasterDebuff(target, scorpidStingMeta, hunter, stingTag, 0)
}

// A warlock's own Curse of the Elements, one per warlock, bidding in the curse's category.
func NewSpellDataCurseOfElementsAuraFor(target *core.Unit, warlock *core.Unit) *core.Aura {
	return newCasterDebuff(target, curseOfElementsMeta, warlock, CurseOfElementsCategory, 0)
}

// A warlock's own Curse of Weakness, one per warlock, bidding in the curse's category.
func NewSpellDataCurseOfWeaknessAuraFor(target *core.Unit, warlock *core.Unit) *core.Aura {
	return newCasterDebuff(target, curseOfWeaknessMeta, warlock, CurseOfWeaknessCategory, 0)
}

// A warlock's own Curse of Recklessness, one per warlock, bidding in the curse's category.
func NewSpellDataCurseOfRecklessnessAuraFor(target *core.Unit, warlock *core.Unit) *core.Aura {
	return newCasterDebuff(target, curseOfRecklessnessMeta, warlock, CurseOfRecklessnessCategory, 0)
}

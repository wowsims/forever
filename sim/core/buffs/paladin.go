package buffs

import (
	"fmt"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// One rank of a paladin aura: the spell the caster used and the number its row states. Retribution
// Aura's damage shield deals Value; every other aura reads its amount off the spell. The paladin
// registers an aura per rank; each joins the same categories as the generated party-buff copy, which
// is the top rank.
type PaladinAuraRank struct {
	SpellID int32
	Rank    int32
	Value   float64
}

var RetributionAuraMaxRank = PaladinAuraRank{SpellID: 10301, Value: RetributionAuraValue(0)}

// A name with the rank the paladin cast; the max rank the raid config applies carries none.
func paladinRankName(name string, rank int32) string {
	if rank > 0 {
		return fmt.Sprintf("%s Rank %d", name, rank)
	}
	return name
}

// A rank the paladin casts is the party-buff row on the rank's own spell, labelled with the rank.
func paladinAuraMeta(base *Meta, rank PaladinAuraRank) *Meta {
	m := *base
	m.Label = paladinRankName(base.Label, rank.Rank)
	m.Spell = spelldata.MustFind(rank.SpellID)
	return &m
}

func NewSpellDataDevotionAuraRank(char *core.Character, rank PaladinAuraRank) *core.Aura {
	return newBuff(&char.Unit, paladinAuraMeta(devotionAuraMeta, rank), true, 0)
}

func NewSpellDataConcentrationAuraRank(char *core.Character, rank PaladinAuraRank) *core.Aura {
	return newBuff(&char.Unit, paladinAuraMeta(concentrationAuraMeta, rank), true, 0)
}

func NewSpellDataFireResistanceAuraRank(char *core.Character, rank PaladinAuraRank) *core.Aura {
	return newBuff(&char.Unit, paladinAuraMeta(fireResistanceAuraMeta, rank), true, 0)
}

func NewSpellDataFrostResistanceAuraRank(char *core.Character, rank PaladinAuraRank) *core.Aura {
	return newBuff(&char.Unit, paladinAuraMeta(frostResistanceAuraMeta, rank), true, 0)
}

func NewSpellDataShadowResistanceAuraRank(char *core.Character, rank PaladinAuraRank) *core.Aura {
	return newBuff(&char.Unit, paladinAuraMeta(shadowResistanceAuraMeta, rank), true, 0)
}

// The aura on the unit the shield protects, at the rank the paladin cast. The external
// (party-buff) copy takes the providing paladin's Holy spell power as externalSpellPower.
func NewSpellDataRetributionAuraRank(char *core.Character, isPlayer bool, rank PaladinAuraRank, externalSpellPower float64) *core.Aura {
	return newSpellPowerDamageShield(&char.Unit, paladinAuraMeta(retributionAuraMeta, rank), isPlayer, rank.Value,
		externalSpellPower)
}

// One rank of a judgement debuff: the spell the target shows and the number its row states. The
// paladin registers a rank per row; the debuff panel applies the max rank, carried by the *MaxRank
// values with Rank 0.
type JudgementRank struct {
	SpellID int32
	Rank    int32
	Value   float64
	HealID  int32 // Judgement of Light's heal, whose row states the amount.
}

var (
	JudgementOfTheCrusaderMaxRank = JudgementRank{SpellID: judgementOfTheCrusaderSpell.ID,
		Value: amount(judgementOfTheCrusaderSpell.Effect(dbcenums.A_MOD_DAMAGE_TAKEN, int32(core.SpellSchoolHoly)))}
	JudgementOfLightMaxRank  = JudgementRank{SpellID: judgementOfLightSpell.ID, HealID: judgementOfLightHealID}
	JudgementOfWisdomMaxRank = JudgementRank{SpellID: judgementOfWisdomSpell.ID,
		Value: amount(spelldata.MustFind(judgementOfWisdomManaID).EnergizeEffect())}
)

// The raid's Judgement of the Crusader, which no generated row reads: the game files state its bonus
// for the holy school alone, which the manifest has no pseudo-stat for.
var judgementOfTheCrusaderSpell = spelldata.MustFind(20303)

// The top ranks' heal and mana. The judgement states neither: its trigger is a dummy, and the
// spells the paladin's own ranks pair with it by hand carry the amounts.
const (
	judgementOfLightHealID  = 20343
	judgementOfWisdomManaID = 20353
)

// Every judgement debuff a paladin puts up carries the tag, so an effect that refreshes "all
// Judgement effects on the target" can find them.
const JudgementAuraTag = "JudgementAura"

// The chance a judgement's row states for the heal or mana it grants attackers. The game files keep
// it outside the spell data, so it is an override, and a row with none would otherwise fire on every
// hit.
func judgementProcChance(spellID int32) float64 {
	chance := spelldata.MustFind(spellID).StatedChance()
	if chance == 0 {
		panic(fmt.Sprintf("buffs: judgement %d states no proc chance", spellID))
	}
	return chance
}

// Judgement of the Crusader raises the Holy damage the target takes by a flat amount. Every rank
// and every paladin share one exclusive category, so the strongest active one is the one that
// counts.
func NewSpellDataJudgementOfTheCrusaderAuraRank(target *core.Unit, rank JudgementRank) *core.Aura {
	bonus := rank.Value
	label := paladinRankName("Judgement of the Crusader", rank.Rank)
	if target.HasAura(label) {
		return target.GetAura(label)
	}

	aura := target.GetOrRegisterAura(core.Aura{
		Label:    label,
		ActionID: core.ActionID{SpellID: rank.SpellID},
		Tag:      JudgementAuraTag,
		Duration: auraDuration(judgementOfTheCrusaderSpell),
	})

	aura.NewExclusiveEffect("Judgement of the Crusader", false, core.ExclusiveEffect{
		Priority: bonus,
		OnGain: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			target.PseudoStats.SchoolBonusSpellDamage[stats.SchoolIndexHoly] += bonus
		},
		OnExpire: func(ee *core.ExclusiveEffect, sim *core.Simulation) {
			target.PseudoStats.SchoolBonusSpellDamage[stats.SchoolIndexHoly] -= bonus
		},
	})

	return aura
}

// The paladin's own Judgement of Light at one rank: the bare 40-second debuff, plus the heal it
// grants whoever strikes the target.
func NewSpellDataJudgementOfLightAuraRank(target *core.Unit, caster *core.Unit, rank JudgementRank) *core.Aura {
	label := paladinRankName("Judgement of Light", rank.Rank)
	if target.HasAura(label) {
		return target.GetAura(label)
	}

	return AttachJudgementOfLightHeal(target.GetOrRegisterAura(core.Aura{
		Label:    label,
		ActionID: core.ActionID{SpellID: rank.SpellID},
		Tag:      JudgementAuraTag,
		Duration: JudgementOfLightDuration(0),
	}), caster, rank)
}

// Whoever lands a damaging melee hit on the judged target rolls the row's chance to be healed for
// the rank's amount. The caster's proc chance mods, e.g. Improved Judgement of Light, raise that
// chance; the raid's copy has no caster. The heal is the attacker's own spell, so its threat goes to
// the attacker.
func AttachJudgementOfLightHeal(aura *core.Aura, caster *core.Unit, rank JudgementRank) *core.Aura {
	healRow := spelldata.MustFind(rank.HealID)
	heal := healRow.HealEffect().Average()
	for _, unit := range aura.Unit.Env.Raid.AllUnits {
		config := spelldata.SpellConfig(unit, healRow, spelldata.Magic(core.ProcMaskSpellHealing), spelldata.Proc(),
			spelldata.Flags(core.SpellFlagProc))
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealHealing(sim, target, heal, spell.OutcomeAlwaysHit)
		}
		unit.GetOrRegisterSpell(config)
	}
	healID := core.ActionID{SpellID: healRow.ID}

	return aura.AttachProcTrigger(core.ProcTrigger{
		Name:                aura.Label + " - Heal",
		Callback:            core.CallbackOnSpellHitTaken,
		ProcMask:            core.ProcMaskMelee,
		Outcome:             core.OutcomeLanded,
		RequireDamageDealt:  true,
		ProcChance:          judgementProcChance(rank.SpellID),
		ProcSpellClassFlags: spelldata.MustFind(rank.SpellID).ClassFlags,
		Caster:              caster,
		Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			spell.Unit.GetSpell(healID).Cast(sim, spell.Unit)
		},
	})
}

// The paladin's own Judgement of Wisdom at one rank.
func NewSpellDataJudgementOfWisdomAuraRank(target *core.Unit, rank JudgementRank) *core.Aura {
	label := paladinRankName("Judgement of Wisdom", rank.Rank)
	if target.HasAura(label) {
		return target.GetAura(label)
	}

	return AttachJudgementOfWisdomMana(target.GetOrRegisterAura(core.Aura{
		Label:    label,
		ActionID: core.ActionID{SpellID: rank.SpellID},
		Tag:      JudgementAuraTag,
		Duration: JudgementOfWisdomDuration(0),
	}), rank)
}

// Whoever lands an attack or spell on the judged target rolls the row's chance to regain the rank's
// mana. Melee claim it returns mana on a miss as well.
func AttachJudgementOfWisdomMana(aura *core.Aura, rank JudgementRank) *core.Aura {
	actionID := core.ActionID{SpellID: rank.SpellID}
	mana := rank.Value

	return aura.AttachProcTrigger(core.ProcTrigger{
		Name:            aura.Label,
		ActionID:        actionID,
		MetricsActionID: actionID,
		ProcChance:      judgementProcChance(rank.SpellID),
		ProcMask:        core.ProcMaskDirect,
		Callback:        core.CallbackOnSpellHitTaken,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !spell.ProcMask.Matches(core.ProcMaskMeleeOrRanged) && !result.Landed() {
				return
			}

			unit := spell.Unit
			if !unit.HasManaBar() {
				return
			}
			if unit.JowManaMetrics == nil {
				unit.JowManaMetrics = unit.NewManaMetrics(actionID)
			}
			unit.AddMana(sim, mana, unit.JowManaMetrics)
		},
	})
}

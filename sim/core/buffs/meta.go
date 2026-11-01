package buffs

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// What the generator states about one manifest row: the spell the aura reads, how the sim names it
// and what it bids in. Which effects are which stats, what they are worth and how they stack is read
// off the row at runtime by spelldata.ParseEffects, with the options Options answers.
type Meta struct {
	// The buff's name in the UI. The aura's label adds " (Player)" or " (External)".
	Label string
	// The row the aura's numbers are read from.
	Spell *spelldata.Spell
	// The cast, for a row whose aura states no duration of its own and whose cooldown is the cast's.
	// Nil where the spell's own row times the buff.
	Cast *spelldata.Spell

	// The aura's own exclusive category, which is also its Tag; "" for none.
	Category string
	// A second category the player's own copy joins without an effect of its own.
	SharedCategory string
	// The category holds one aura at a time, and the aura bids in it as a whole.
	SingleAura bool

	// The improving talent, and the effect of it that prices the buff, counted the way EffectN
	// counts. TalentScalesDuration says the talent prices the duration rather than the amount.
	Talent               spelldata.Ladder
	TalentEffect         int32
	TalentScalesDuration bool
	// TalentDivisor prices the buff at talentPoints/TalentDivisor of the spell's amount, for a talent
	// a tooltip prices by rank; 0 for a talent priced by a spell modifier.
	TalentDivisor int32

	// The aura effects the buff leaves out, by aura.
	SkipAuras []dbcenums.EffectAuraType
	// The finisher the raid config puts on the target is cast at full combo points.
	FullComboPoints bool
}

// The parse options the row states, for a caster with these talent points.
func (m *Meta) Options(talentPoints int32) []spelldata.ParseOpt {
	opts := spelldata.RaidBuffOptions(m.SkipAuras, m.FullComboPoints)
	switch {
	case m.TalentDivisor > 0:
		opts = append(opts, spelldata.ScaledByRanks(talentPoints, m.TalentDivisor))
	case !m.TalentScalesDuration:
		opts = append(opts, spelldata.ScaledBy(m.talentMod(talentPoints)))
	}
	return opts
}

// The first amount the aura applies, in the sim's units: the stat, the multiplier or the pseudo-stat
// change it attaches. A damage shield's amount is the damage it deals.
func (m *Meta) Value(talentPoints int32) float64 {
	for i := range m.Spell.Effects {
		if e := &m.Spell.Effects[i]; e.Aura == dbcenums.A_DAMAGE_SHIELD {
			value := amount(e)
			if mod := m.talentMod(talentPoints); mod != spelldata.NilEffect && !m.TalentScalesDuration {
				value = spelldata.Scaled(value, mod)
			}
			return value
		}
	}

	if applied := spelldata.DryRun(m.Spell, m.Options(talentPoints)...).Applied; len(applied) > 0 {
		return applied[0].Value
	}
	return 0
}

func (m *Meta) Duration(talentPoints int32) time.Duration {
	if m.TalentScalesDuration {
		return talentScaledDuration(m.Spell, m.talentMod(talentPoints))
	}
	if m.Spell.DurationMs <= 0 && m.Cast != nil {
		return auraDuration(m.Cast)
	}
	return auraDuration(m.Spell)
}

func (m *Meta) Cooldown() time.Duration {
	if m.Cast != nil {
		return cooldown(m.Cast)
	}
	return cooldown(m.Spell)
}

// The talent's modifier at a rank, NilEffect for an untaken talent or a row no talent prices.
func (m *Meta) talentMod(talentPoints int32) *spelldata.Effect {
	return m.Talent.Rank(talentPoints).EffectN(int(m.TalentEffect))
}

// Tag 0 is the player's own copy and -1 the external caster's, which the character build phase has
// to see so that stat dependencies are computed with it.
func (m *Meta) label(isPlayer bool) string {
	return m.Label + " (" + core.Ternary(isPlayer, "Player", "External") + ")"
}

// A rank of a generated buff is its row on the rank's own spell, labelled with the rank so that two ranks on one unit
// never share an aura. Both bid in the row's category.
func (m *Meta) atRank(rank *spelldata.Spell) *Meta {
	return m.atState(rank, m.Label+" "+rank.Rank)
}

// The row on another spell under the label given, bidding in the row's categories. A priest's prayer
// reads its own spell this way against the raid's input of its stat.
func (m *Meta) atState(spell *spelldata.Spell, label string) *Meta {
	state := *m
	state.Label = label
	state.Spell = spell
	return &state
}

// The states of a rank select, in the order the proto field counts them from 1.
type rankStates []*Meta

// A state past the last reads as the last.
func (r rankStates) at(state int32) *Meta {
	return r[min(int(state), len(r))-1]
}

func (m *Meta) actionID(isPlayer bool) core.ActionID {
	return core.ActionID{SpellID: m.Spell.ID}.WithTag(core.TernaryInt32(isPlayer, 0, -1))
}

// The aura a generated buff registers on a player: labelled and tagged by label and actionID, with
// Tag set to the category, MaxStacks the row's own stacks (not its charges), the external copy in
// CharacterBuildPhaseBuffs, and the effects read by spelldata.ParseEffects with Options plus
// SchoolResistances, and where Category is set, Exclusive(Category, true) for a single-aura one or
// ExclusivePerStat(Category) for any other. Only the player's own copy joins SharedCategory.
func newBuff(unit *core.Unit, m *Meta, isPlayer bool, talentPoints int32) *core.Aura {
	return m.buff(unit, isPlayer, m.label(isPlayer), talentPoints, m.Options(talentPoints))
}

// A caster's own copy of a generated buff, one per caster: newBuff's player copy labelled
// "<Label>-<caster's label>". It bids in Category all the same.
func newCasterBuff(unit *core.Unit, m *Meta, caster *core.Unit, talentPoints int32) *core.Aura {
	return m.buff(unit, true, m.Label+"-"+caster.Label, talentPoints, m.Options(talentPoints))
}

// A buff worth its amounts once per item in the party: newBuff with Count(count).
func newItemCountBuff(unit *core.Unit, m *Meta, isPlayer bool, count float64) *core.Aura {
	return m.buff(unit, isPlayer, m.label(isPlayer), 0, append(m.Options(0), spelldata.Count(count)))
}

func (m *Meta) buff(unit *core.Unit, isPlayer bool, label string, talentPoints int32, opts []spelldata.ParseOpt) *core.Aura {
	opts = append(opts, spelldata.SchoolResistances())
	if m.Category != "" {
		if m.SingleAura {
			opts = append(opts, spelldata.Exclusive(m.Category, true))
		} else {
			opts = append(opts, spelldata.ExclusivePerStat(m.Category))
		}
	}

	buildPhase := core.Ternary(isPlayer, core.CharacterBuildPhaseNone, core.CharacterBuildPhaseBuffs)
	return m.parsedAura(unit, isPlayer, label, m.Category, talentPoints, buildPhase, opts)
}

// The aura a generated debuff registers on the target: labelled and tagged like newBuff, never in a
// build phase, with Exclusive(Category, SingleAura) where it names a category and no school
// resistance categories of its own. A debuff in a category that holds more than one aura bids each
// effect alone, so a stronger one holds only the effects of its own kind.
func newDebuff(target *core.Unit, m *Meta, isPlayer bool, talentPoints int32) *core.Aura {
	return m.debuff(target, isPlayer, m.label(isPlayer), m.Category, talentPoints)
}

// A caster's own copy of a generated debuff, one per caster: newDebuff's player copy labelled
// "<Label>-<caster's label>" and tagged with tag, which the caster finds it by. It bids in Category
// all the same.
func newCasterDebuff(target *core.Unit, m *Meta, caster *core.Unit, tag string, talentPoints int32) *core.Aura {
	return m.debuff(target, true, m.Label+"-"+caster.Label, tag, talentPoints)
}

// Every debuff lands and only the strongest in its category counts, so debuff rows leave SingleAura
// false. Curse of Recklessness, Faerie Fire and Dust Cloud share MinorArmorReduction this way.
func (m *Meta) debuff(target *core.Unit, isPlayer bool, label string, tag string, talentPoints int32) *core.Aura {
	opts := m.Options(talentPoints)
	if m.Category != "" {
		opts = append(opts, spelldata.Exclusive(m.Category, m.SingleAura))
	}
	return m.parsedAura(target, isPlayer, label, tag, talentPoints, core.CharacterBuildPhaseNone, opts)
}

func (m *Meta) parsedAura(unit *core.Unit, isPlayer bool, label string, tag string, talentPoints int32,
	buildPhase core.CharacterBuildPhase, opts []spelldata.ParseOpt) *core.Aura {
	// A second caller of the same copy shares the aura the first registered, which carries its
	// effects already.
	if aura := unit.GetAura(label); aura != nil {
		return aura
	}

	aura := unit.RegisterAura(core.Aura{
		Label:      label,
		Tag:        tag,
		ActionID:   m.actionID(isPlayer),
		Duration:   m.Duration(talentPoints),
		MaxStacks:  int32(m.Spell.MaxStack),
		BuildPhase: buildPhase,
	})

	// No row a raid buff states acts through a character rather than its unit.
	spelldata.ParseEffects(nil, aura, m.Spell, opts...)
	joinSharedCategory(aura, m.SharedCategory, isPlayer)
	return aura
}

// A damage shield, which the parse does not read: NewDamageShield with the spell's school and
// Value as its damage.
func newDamageShield(unit *core.Unit, m *Meta, isPlayer bool, talentPoints int32) *core.Aura {
	aura := NewDamageShield(unit, m.label(isPlayer), m.actionID(isPlayer), m.Duration(talentPoints),
		m.Category, m.SingleAura, m.Spell.School, m.Value(talentPoints), 0)
	joinSharedCategory(aura, m.SharedCategory, isPlayer)
	return aura
}

// Retribution Aura and Thorns scale with their caster's spell power in Forever even though their
// game files rows carry no coefficient (every rank states EffectBonusCoefficient 0, and the damage
// still moves with spell power in game). The coefficient is the 1.5 s cast-time floor over 3.5, the
// AoE divisor because the shield hits every attacker, and the 0.42 penalty for the aura effect: 6% of
// spell power, the ratio Blizzard states for Retribution Aura and the one measured for Thorns.
const DamageShieldSpellPowerCoefficient = 1.5 / 3.5 / 3 * 0.42

// A damage shield that scales with its caster's spell power. The player's copy reads its own spell
// power through the proc spell; the external copy cannot see the caster, so externalSpellPower
// stands in for it, folded into the damage it bids and deals, and the recipient's own stats stay
// out of the damage.
func newSpellPowerDamageShield(unit *core.Unit, m *Meta, isPlayer bool, damage float64, externalSpellPower float64) *core.Aura {
	label := m.label(isPlayer)
	if unit.HasAura(label) {
		return unit.GetAura(label)
	}

	coefficient := DamageShieldSpellPowerCoefficient
	if !isPlayer {
		damage, coefficient = damage+coefficient*externalSpellPower, 0
	}
	aura := NewDamageShield(unit, label, m.actionID(isPlayer), m.Duration(0), m.Category, m.SingleAura,
		m.Spell.School, damage, coefficient)
	joinSharedCategory(aura, m.SharedCategory, isPlayer)
	return aura
}

package spelldata

import (
	"fmt"
	"math"
	"os"
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/stats"
)

// How a parse is narrowed: which effects it reads, what gates them and whether their values follow
// the aura's stacks.
type ParseOpt func(*parseOptions)

type parseOptions struct {
	only         []int32
	skip         []int32
	skipAuras    []dbcenums.EffectAuraType
	cond         func() bool
	ignoreStacks bool

	level        int32
	scale        *Effect
	rankFraction float64
	comboPoints  bool
	count        float64
	buffAuras    bool
	hand         core.ProcMask

	exclusive         bool
	category          string
	singleAura        bool
	perStatCategory   string
	schoolResistances bool

	casterSlot *casterSlot

	showAsSchoolHit core.SpellSchool
}

// Only these effects, counted from 1 by position the way EffectN counts.
func Effects(idx ...int32) ParseOpt {
	return func(o *parseOptions) {
		o.only = append(o.only, idx...)
	}
}

func SkipEffects(idx ...int32) ParseOpt {
	return func(o *parseOptions) {
		o.skip = append(o.skip, idx...)
	}
}

// No effect of these auras is read, wherever the row states it.
func SkipAuras(auras ...dbcenums.EffectAuraType) ParseOpt {
	return func(o *parseOptions) {
		o.skipAuras = append(o.skipAuras, auras...)
	}
}

// A condition every attachment is gated on, beyond the aura being up: it is read when the aura is
// gained and whenever the caller calls Refresh, and nothing is active while it answers false.
func Conditional(fn func() bool) ParseOpt {
	return func(o *parseOptions) {
		o.cond = fn
	}
}

// The values do not follow the aura's stacks, which is what a row that states charges rather than
// cumulative stacks means.
func IgnoreStacks() ParseOpt {
	return func(o *parseOptions) {
		o.ignoreStacks = true
	}
}

// The amounts are priced at this level rather than the caster's. A raid's buff or debuff is cast by
// a player of the sim's own level whatever unit it lands on, and a debuff on a level 63 boss has no
// character to read that level from.
func Level(n int32) ParseOpt {
	return func(o *parseOptions) {
		o.level = n
	}
}

// An improving talent's modifier on the amount, truncated the way the game resolves one: a
// percentage of it for A_ADD_PCT_MODIFIER, a flat addition otherwise. It prices the first effect the
// table attaches, in the game files' units before the row converts them. An untaken talent is
// NilEffect, which changes nothing.
func ScaledBy(mod *Effect) ParseOpt {
	return func(o *parseOptions) {
		if mod != nil && mod != NilEffect {
			o.scale = mod
		}
	}
}

// The talent's ranks out of divisor, for a talent whose tooltip prices the buff as a fraction of the
// spell's own amount. Every amount the parse attaches is scaled, in the game files' units, and rounded
// to the nearest whole amount the way the tooltip shows it. Rank 0 changes nothing: the full amount.
func ScaledByRanks(ranks, divisor int32) ParseOpt {
	return func(o *parseOptions) {
		if ranks > 0 && divisor > 0 {
			o.rankFraction = float64(ranks) / float64(divisor)
		}
	}
}

// The finisher at full combo points: an effect the game files state per combo point spent is worth
// five of them. A parse without it skips an effect whose amount is its per-point value alone.
func FullComboPoints() ParseOpt {
	return func(o *parseOptions) {
		o.comboPoints = true
	}
}

// The amounts are counted n times over, for one aura that stands for several sources of it. A
// multiplier cannot be counted, so those rows are skipped the way a stacking row skips them.
func Count(n float64) ParseOpt {
	return func(o *parseOptions) {
		o.count = n
	}
}

// The row sits on the weapon in this hand, so a hit, crit or rating aura bound to a matching weapon
// reaches that hand's attacks alone and holds while that hand holds one. An aura this hand cannot
// carry is skipped.
func OnHand(hand core.ProcMask) ParseOpt {
	return func(o *parseOptions) {
		o.hand = hand
	}
}

// The rows only a raid buff reads, on top of the table: see buffAuraTable.
func BuffAuras() ParseOpt {
	return func(o *parseOptions) {
		o.buffAuras = true
	}
}

// The parse bids under an exclusive category, where only the strongest member applies: a multiplier
// priced by how far it is from 1, a stat by its size, and a stacking row once per stack. With
// singleAura everything it attaches bids as a whole, priced at the first attachment, and deactivates
// the others outright. Without, each stat and pseudo-stat bids alone under category + its key +
// "Add" or "Mul", so a stronger member holds only the effects of its own kind and the aura keeps the
// rest.
func Exclusive(category string, singleAura bool) ParseOpt {
	return func(o *parseOptions) {
		o.exclusive, o.category, o.singleAura = true, category, singleAura
	}
}

// Every stat and pseudo-stat the parse attaches bids alone, under the category name core's own
// exclusive stat buffs build: category + the stat or PseudoStats field + "Add" or "Mul". A generated
// buff and a hand-written scroll of the same stat compete that way. The stacks are not followed.
func ExclusivePerStat(category string) ParseOpt {
	return func(o *parseOptions) {
		o.perStatCategory = category
	}
}

// Every flat school resistance the parse attaches bids alone under its school's category, whatever
// else the parse does with the rest, so that only the strongest source of each school applies.
// Armor is not a school.
func SchoolResistances() ParseOpt {
	return func(o *parseOptions) {
		o.schoolResistances = true
	}
}

// The unit that put the aura on the unit it sits on, for the rows that reach only that caster's
// attacks on it: Shadow Weaving's shadow damage, Hemorrhage's Rupture, Winter's Chill's crit. Those
// rows share one slot, index out of total, in the caster's attack table against the aura's unit:
// every damage row multiplies into its DamageDoneByCasterExtraMultiplier handler and every crit row
// adds into its CritChanceByCasterBonus handler, each enabled while any of its rows is up. The class
// numbers its slots the way it numbers its AttachDDBC calls, one per parse. Without a caster those
// rows are skipped.
func FromCaster(caster *core.Unit, index, total int) ParseOpt {
	return func(o *parseOptions) {
		o.casterSlot = &casterSlot{caster: caster, index: index, total: total}
	}
}

// The character sheet shows the row's SPELLMOD_RESIST_MISS_CHANCE hit in these schools' hit, from the
// talents stage on. The row names its spells by class flags, which carry no school, so the caller states
// the schools its tooltip names, for a talent that reaches the class's main spells of them: Shadow Focus
// takes core.SpellSchoolShadow. The spells the mod reaches stay the ones the class flags name.
func ShowAsSchoolHit(schools core.SpellSchool) ParseOpt {
	return func(o *parseOptions) {
		o.showAsSchoolHit = schools
	}
}

// One attachment the parse made, or one part of an attachment that has parts, for a log or a test: the
// effect it came from, the sim kind it became and the value it carries in the sim's own units. A time
// value is stated in milliseconds, and a multiplier as the factor itself.
type Applied struct {
	Effect *Effect
	Kind   string
	Value  float64
}

// An aura effect the parse left out, and its position on the row, counted from 1 the way EffectN
// counts. Reason says why, where the row that refused the effect states one.
type SkippedEffect struct {
	*Effect
	Position int
	Reason   string
}

// What one parse did. Skipped holds the aura effects the table does not know, which are the ones a
// port still has to wire by hand.
type Parsed struct {
	Applied []Applied
	Skipped []SkippedEffect

	attachments []*attachment
	aura        *core.Aura
	cond        func() bool
	stacking    bool
	count       float64
	quiet       bool

	// The attachments the aura's own gain and expiry drive, which is all of them unless a category
	// holds some.
	live []*attachment
}

// Everything the row's aura effects say, attached to an aura: the mods, stat buffs and pseudo-stat
// multipliers follow the aura's gain and expiry, and their values follow its stacks where the row
// states cumulative ones.
//
// The effects act on the aura's own unit, and the character is that unit's, for the few rows whose
// helper lives on a character rather than on a unit. An aura on a unit with no character of its own -
// a debuff on an enemy - takes a nil character and has those rows skipped and reported.
//
// An aura that is already up when it is parsed is caught up the way core's Attach helpers catch one
// up, except for the rows that need a Simulation to act: the haste and energy regen multipliers stay
// off until the aura is applied again. What a category holds waits for the aura's next application
// too.
func ParseEffects(character *core.Character, aura *core.Aura, s *Spell, opts ...ParseOpt) *Parsed {
	if aura == nil {
		return &Parsed{}
	}
	return parse(aura.Unit, character, aura, s, opts, false)
}

// Everything the row's aura effects say, applied to the character now: a talent or a passive, which
// the game files state as an aura the unit is never without.
func ParseStatic(character *core.Character, s *Spell, opts ...ParseOpt) *Parsed {
	if character == nil {
		return &Parsed{}
	}
	return parse(&character.Unit, character, nil, s, opts, false)
}

// What ParseEffects would attach and skip for the row, with nothing registered anywhere: the answer a
// generator reads before it writes the call. The aura sits on a unit with no character, so the rows
// that act through a character are skipped. The amounts are priced at core.CharacterLevel unless
// Level says otherwise.
func DryRun(s *Spell, opts ...ParseOpt) *Parsed {
	return parse(&core.Unit{Level: core.CharacterLevel}, nil, nil, s, opts, true)
}

func parse(unit *core.Unit, character *core.Character, aura *core.Aura, s *Spell, opts []ParseOpt, dry bool) *Parsed {
	parsed := &Parsed{quiet: dry}
	if unit == nil || s == nil || s == Nil || len(s.Effects) == 0 {
		return parsed
	}

	o := &parseOptions{}
	for _, opt := range opts {
		opt(o)
	}

	stacking := s.MaxStack > 0 && !o.ignoreStacks
	p := &parser{
		unit:            unit,
		character:       character,
		spell:           s,
		static:          aura == nil && !dry,
		conditional:     o.cond != nil,
		stacking:        stacking || o.count != 0,
		dry:             dry,
		aura:            aura,
		hand:            o.hand,
		casterSlot:      o.casterSlot,
		showAsSchoolHit: o.showAsSchoolHit,
	}
	if o.casterSlot != nil {
		o.casterSlot.target = unit
	}

	// The amount is the caster's, so a debuff on an enemy is still priced at the character's level
	// and not at the level of the unit it sits on. Without a character the FromCaster unit is the
	// caster, and without either there is only the unit.
	level := unit.Level
	if character != nil {
		level = character.Level
	} else if o.casterSlot != nil {
		level = o.casterSlot.caster.Level
	}
	if o.level != 0 {
		level = o.level
	}

	scale := o.scale
	for i := range s.Effects {
		e := &s.Effects[i]
		if !o.reads(int32(i+1)) || slices.Contains(o.skipAuras, e.Aura) || !AppliesAura(e.Type) {
			continue
		}

		value := e.MeanAt(level)

		if e.PointsPerResource != 0 && value == 0 {
			if !o.comboPoints {
				parsed.skip(s, i, e, "")
				continue
			}
			value = float64(e.PointsPerResource) * fullComboPoints
		}

		attached := o.row(e.Aura)
		if attached == nil {
			parsed.skip(s, i, e, "")
			continue
		}

		if scale != nil {
			value = Scaled(value, scale)
		}
		if o.rankFraction > 0 {
			value = math.Round(value * o.rankFraction)
		}
		a := attached(p, e, value)
		if a == nil {
			parsed.skip(s, i, e, "")
			continue
		}
		if a.skipReason != "" {
			parsed.skip(s, i, e, a.skipReason)
			continue
		}
		scale = nil

		if a.key == "" {
			a.key = a.kind
		}
		if len(a.parts) == 0 {
			parsed.Applied = append(parsed.Applied, Applied{Effect: e, Kind: a.kind, Value: a.value})
		}
		for _, part := range a.parts {
			parsed.Applied = append(parsed.Applied, Applied{Effect: e, Kind: part.kind, Value: part.value})
		}
		parsed.attachments = append(parsed.attachments, a)
	}
	if o.showAsSchoolHit != 0 && !p.showAsSchoolHitStamped {
		panic(fmt.Sprintf("spelldata: spell %d attaches no SPELLMOD_RESIST_MISS_CHANCE hit for ShowAsSchoolHit to show", s.ID))
	}

	parsed.aura = aura
	parsed.cond = o.cond
	parsed.stacking = stacking
	parsed.count = 1
	if o.count != 0 {
		parsed.count = o.count
	}

	if dry {
		return parsed
	}

	if aura == nil {
		parsed.live = parsed.attachments
		parsed.set(nil, parsed.level())
		return parsed
	}

	parsed.live = parsed.bid(aura, o)

	aura.ApplyOnGain(func(_ *core.Aura, sim *core.Simulation) {
		parsed.set(sim, parsed.level())
	}).ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
		parsed.set(sim, 0)
	})

	if stacking {
		aura.ApplyOnStacksChange(func(_ *core.Aura, sim *core.Simulation, _ int32, _ int32) {
			parsed.set(sim, parsed.level())
		})
	}

	if aura.IsActive() {
		parsed.catchUp(parsed.level())
	}

	return parsed
}

// The stats the attachments add to or multiply, in effect order.
func (p *Parsed) Stats() []stats.Stat {
	var sts []stats.Stat
	for _, a := range p.attachments {
		sts = append(sts, a.stats...)
	}
	return sts
}

// The finisher the raid config has on the target is cast at the most combo points there are.
const fullComboPoints = 5

// The row the parse reads an aura with: a buff's own rows first where the caller asked for them.
func (o *parseOptions) row(aura dbcenums.EffectAuraType) row {
	if o.buffAuras {
		if r, ok := buffAuraTable[aura]; ok {
			return r
		}
	}
	return auraTable[aura]
}

// Reads the condition again and turns the attachments on or off accordingly, for a caller whose
// condition changed under a parse that is already in place.
func (p *Parsed) Refresh(sim *core.Simulation) {
	p.set(sim, p.level())
}

// What the attachments are worth right now: nothing while the condition answers false or the aura is
// down, the plain value where the row does not stack, and once per stack where it does, each of
// those counted as often as Count says.
func (p *Parsed) level() float64 {
	if p.cond != nil && !p.cond() {
		return 0
	}
	if p.aura != nil && !p.aura.IsActive() {
		return 0
	}
	if p.stacking && p.aura != nil {
		if stacks := p.aura.GetStacks(); stacks > 0 {
			return float64(stacks) * p.count
		}
	}
	return p.count
}

func (p *Parsed) set(sim *core.Simulation, level float64) {
	for _, a := range p.live {
		a.set(sim, level)
	}
}

// What core's Attach helpers do for an aura that is already up when they are attached: apply now,
// with no Simulation in hand. The rows that read one are left for the aura's next application.
func (p *Parsed) catchUp(level float64) {
	for _, a := range p.live {
		if !a.needsSim {
			a.set(nil, level)
		}
	}
}

func (p *Parsed) skip(s *Spell, i int, e *Effect, reason string) {
	p.Skipped = append(p.Skipped, SkippedEffect{Effect: e, Position: i + 1, Reason: reason})
	if !p.quiet {
		report(s, i+1, e, reason)
	}
}

// Whether the parse reads the effect at this position.
func (o *parseOptions) reads(pos int32) bool {
	if len(o.only) > 0 && !slices.Contains(o.only, pos) {
		return false
	}
	return !slices.Contains(o.skip, pos)
}

// The effect types that put an aura on someone: the plain application and the area auras, which
// carry the same aura and misc values.
func AppliesAura(t dbcenums.SpellEffectType) bool {
	return t == dbcenums.E_APPLY_AURA || t == dbcenums.E_APPLY_AREA_AURA_PARTY ||
		t == dbcenums.E_APPLY_AREA_AURA_RAID
}

// Whether ParseEffects attaches something for the effect's aura. The skill row reads only Defense and
// the weapon skills, so a profession skill is not one. Another row can still refuse one effect by its
// misc value, which only a parse against a unit finds out.
func (e *Effect) ParsedAura() bool {
	if _, weapon := weaponSkillLines[e.Misc]; e.Aura == dbcenums.A_MOD_SKILL && e.Misc != miscDefense && !weapon {
		return false
	}
	return AppliesAura(e.Type) && auraTable[e.Aura] != nil
}

// What a port has to wire by hand, on the console when SPELLDATA_REPORT is set.
func report(s *Spell, pos int, e *Effect, reason string) {
	if os.Getenv("SPELLDATA_REPORT") == "" {
		return
	}
	fmt.Printf("spelldata: unparsed %d %s %s value %v%s\n", s.ID, s.Name, effectNote(pos, e), e.BasePoints,
		reasonNote(reason))
}

func reasonNote(reason string) string {
	if reason == "" {
		return ""
	}
	return ": " + reason
}

func effectNote(pos int, e *Effect) string {
	return fmt.Sprintf("effect %d %s(%d) misc %d", pos, auraName(e.Aura), e.Aura, e.Misc)
}

// The game files' name for an aura the parser skipped, or its number where the game files name none.
func auraName(a dbcenums.EffectAuraType) string {
	if name, ok := dbcenums.Named(a); ok {
		return name
	}
	return fmt.Sprintf("A_%d", a)
}

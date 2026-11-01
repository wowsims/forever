package database

// Resolves tools/database/buffmanifest against the spell store's captured game files
// rows and renders sim/core/buffs/buffs_auto_gen.go and
// sim/core/buffs/debuffs_auto_gen.go from the result, in the same pass that
// renders the store.
//
// The manifest names the spell each proto field reads, and the generated file
// states it as a buffs.Meta: which of its effects are which stats, what they are
// worth and how they stack is read off the row at runtime by
// spelldata.ParseEffects. The generator asks the same parse, through
// spelldata.DryRun, whether the row attaches anything and what it leaves out, so
// a row it writes is one the sim builds the same way. A row the parse attaches
// nothing of renders as a commented shell carrying the reason.

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"text/template"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/buffmanifest"
)

const buffsGenFile = "sim/core/buffs/buffs_auto_gen.go"
const debuffsGenFile = "sim/core/buffs/debuffs_auto_gen.go"

// ResolvedBuff is one manifest row plus everything the database states about it.
type ResolvedBuff struct {
	buffmanifest.BuffSpec

	Scope buffmanifest.BuffScope
	Go    string // identifier stem: "BattleShout"
	Proto buffmanifest.BuffProtoType
	// Kind is the manifest's, or a damage shield where the spell states one.
	Kind buffmanifest.BuffKind

	// Name is the game files' name for the spell the player learns, and AuraName the aura's where the
	// manifest names a cast beside it. Both are "" for a spell no class family files, which is also
	// when Owner is the manifest's rather than the family's.
	Name     string
	AuraName string
	Owner    proto.Class

	SpellID     int32 // the spell the aura's numbers are read from
	CastSpellID int32 // the cast, where the manifest pins one; SpellID otherwise
	DurationMs  int32 // -1 or 0 never expires
	CooldownMs  int32 // 0 when the game files state none; read from the cast
	Totem       bool  // the cast summons into a totem slot, so the owner's true pets stand under it too

	// Spell is the row the store carries for SpellID, built from the same
	// captured game files rows the store is rendered from.
	Spell *spelldata.Spell
	// ScopeSpell is the row whose targeting states who the aura reaches: Spell, or the group spell a
	// raid input stands for.
	ScopeSpell *spelldata.Spell

	// A rank select's states, in the order the proto field counts them from 1.
	RankStates []rankState

	// SkipAuraTypes are the auras the row states at 0 beside the buff, which the parse leaves out.
	SkipAuraTypes []dbcenums.EffectAuraType
	// FullComboPoints says a debuff's amount is stated per combo point, and the
	// raid config's copy is the finisher at full combo points.
	FullComboPoints bool

	// Applied is what the parse attaches for the row, and LeftOut the aura
	// effects it leaves out, one note each.
	Applied []spelldata.Applied
	LeftOut []string

	// TalentRanks is how many points the improving talent takes, 0 for a row
	// no talent prices.
	TalentRanks          int32
	TalentScalesDuration bool

	// TalentSpellID is the spell of the trait node that prices the improvement,
	// which is the icon the UI shows for the improved state. TalentPosition is
	// the effect of that spell the improvement is read from, counted the way
	// EffectN counts.
	TalentSpellID  int32
	TalentPosition int32
	// TalentDivisor prices a row whose talent states its effect only in its tooltip, as
	// ${$m1/-N*$<SpellID>m1}: rank m1 of N is worth m1/N of every amount the spell attaches. A row
	// whose own tooltip shows the talent's $<Talent>s1 takes the node's ranks as N. 0 for a talent
	// priced by a spell modifier.
	TalentDivisor int32

	ScopeFromClient buffmanifest.BuffScope

	DBName string

	Supported bool
	Reason    string
	Warnings  []string
}

// One state of a rank select, as its row and its cast state it.
type rankState struct {
	SpellID     int32
	CastSpellID int32
	Label       string // the cast's name and rank
}

func (r *ResolvedBuff) warn(format string, args ...any) {
	r.Warnings = append(r.Warnings, fmt.Sprintf(format, args...))
}

func (r *ResolvedBuff) unsupported(format string, args ...any) {
	r.Supported = false
	r.Reason = fmt.Sprintf(format, args...)
}

// Every manifest row, read out of the game files rows the store is built from.
func resolveBuffManifest(in *storeInputs) ([]ResolvedBuff, error) {
	t := in.tables()
	var rows []ResolvedBuff
	for _, spec := range buffmanifest.All() {
		row, err := resolveBuff(t, in.TraitNodes, spec)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", spec.Field, err)
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func resolveBuff(t *spellTables, nodes []traitNode, spec buffmanifest.Row) (ResolvedBuff, error) {
	row := ResolvedBuff{
		BuffSpec: spec.BuffSpec, Scope: spec.Scope, Go: spec.GoStem(), Proto: spec.ProtoType(), Kind: spec.Kind,
		Supported: true, ScopeFromClient: spec.Scope,
	}

	compiled, err := compiledProtoType(row)
	if err != nil {
		return row, err
	}

	if row.Kind == buffmanifest.KindFlag {
		row.unsupported("%s", spec.Reason)
		return row, nil
	}

	if err := loadBuffSpell(t, &row); err != nil {
		return row, err
	}
	if err := nameBuff(t, &row); err != nil {
		return row, err
	}
	if row.Kind == buffmanifest.KindPlain && row.hasDamageShield() {
		row.Kind = buffmanifest.KindDamageShield
	}
	parseBuff(&row)
	if err := resolveRankStates(t, &row); err != nil {
		return row, err
	}
	if err := resolveTalent(t, nodes, &row); err != nil {
		return row, err
	}
	warnUnpricedTalents(t, nodes, &row)
	validateScope(&row)

	if compiled != buffProtoTypeNames[row.Proto] && row.Supported {
		row.unsupported("proto field not yet retyped: the sim compiled against %s, the manifest declares %s",
			compiled, buffProtoTypeNames[row.Proto])
	}
	return row, nil
}

// The spell's name, its row and duration, and the cooldown of a buff other
// players cast on their own: the shared timer the sim hands the next source is
// the cast's cooldown, and a totem's aura states neither how long the totem
// stands nor when the next one may be dropped. Mana Tide is that row - 1325257
// carries the mana, the cast 1325256 the 13 seconds and the 5 minutes.
func loadBuffSpell(t *spellTables, row *ResolvedBuff) error {
	if row.SpellID = row.BuffSpec.SpellID; row.SpellID == 0 {
		return fmt.Errorf("names no spell")
	}
	row.CastSpellID = cmp.Or(row.CastID, row.SpellID)
	for _, id := range []int32{row.SpellID, row.CastSpellID} {
		if _, named := t.Names[id]; !named {
			return fmt.Errorf("spell %d is no spell in the game files", id)
		}
	}

	spell := t.row(row.SpellID)
	row.Spell = storeRowSpell(spell)
	row.ScopeSpell = row.Spell
	if group, ok := groupInputSpells[row.Field]; ok {
		row.ScopeSpell = storeRowSpell(t.row(group))
	}
	row.DBName = spell.Name
	row.DurationMs = spell.DurationMs
	row.Totem = summonsTotem(t.row(row.CastSpellID))

	if row.Kind != buffmanifest.KindExternalCD {
		return nil
	}
	cast := t.row(row.CastSpellID)
	row.CooldownMs = max(cast.CooldownMs, cast.CategoryCooldownMs)
	if row.CooldownMs == 0 {
		row.unsupported("spell %d states no cooldown, which an external cooldown is scheduled by", row.CastSpellID)
		return nil
	}
	if row.DurationMs <= 0 && row.CastSpellID != row.SpellID {
		row.DurationMs = cast.DurationMs
	}
	if row.DurationMs <= 0 {
		row.unsupported("spell %d states no duration, which the external cooldown's aura needs", row.SpellID)
	}
	return nil
}

// The SummonProperties rows of the four totem slots (fire, earth, water and air), which a totem's
// summon names in its second misc value. The game files database carries no SummonProperties table.
var totemSlotSummonProperties = []int32{63, 81, 82, 83}

func summonsTotem(cast storeSpell) bool {
	return slices.ContainsFunc(cast.Effects, func(e storeEffect) bool {
		return e.Type == dbcenums.E_SUMMON && slices.Contains(totemSlotSummonProperties, e.Misc2)
	})
}

// The store's row as sim/core/spelldata states it, copied field by field from the generator's
// mirror of it. The parse reads this rather than the compiled store, which is the output the same
// pass rewrites.
func storeRowSpell(row storeSpell) *spelldata.Spell {
	s := &spelldata.Spell{}
	copyStoreFields(reflect.ValueOf(s).Elem(), reflect.ValueOf(row))
	return s
}

func copyStoreFields(to reflect.Value, from reflect.Value) {
	for i := 0; i < from.NumField(); i++ {
		field := from.Type().Field(i)
		dst := to.FieldByName(field.Name)
		if !field.IsExported() || !dst.IsValid() {
			continue
		}

		src := from.Field(i)
		if src.Kind() == reflect.Slice && src.Type().Elem().Kind() == reflect.Struct {
			dst.Set(reflect.MakeSlice(dst.Type(), src.Len(), src.Len()))
			for j := 0; j < src.Len(); j++ {
				copyStoreFields(dst.Index(j), src.Index(j))
			}
			continue
		}
		dst.Set(src.Convert(dst.Type()))
	}
}

// The names the game files give the row's spells, and the class whose family files them. A spell no
// class family files takes the manifest's owner, and one that a family files cannot.
func nameBuff(t *spellTables, row *ResolvedBuff) error {
	row.Owner = row.BuffSpec.Owner
	family := row.Spell.ClassFlags.Family
	if family == 0 {
		return nil
	}
	if row.Owner != proto.Class_ClassUnknown {
		return fmt.Errorf("names owner %s, but spell %d is filed under class family %d", row.Owner, row.SpellID, family)
	}
	for class, classFamily := range core.ClassSpellFamilies {
		if classFamily == family {
			row.Owner = class
		}
	}
	row.Name = t.Names[row.CastSpellID]
	if row.CastID != 0 {
		row.AuraName = t.Names[row.SpellID]
	}
	return nil
}

// What the parse makes of the row with the options the generated Meta states:
// the amounts it attaches and the aura effects it leaves out. A kind whose
// behaviour is a driver's is written whatever the parse attaches, a damage
// shield needs its shield, and every other kind needs an amount.
func parseBuff(row *ResolvedBuff) {
	// The value a debuff states per combo point is 0 on the spell itself. The
	// raid config is one debuff that is simply on the target, which is the
	// finisher at full combo points; a caster spending fewer of them takes its
	// own value through a driver.
	if row.Scope == buffmanifest.ScopeDebuff {
		row.FullComboPoints = slices.ContainsFunc(row.Spell.Effects, func(e spelldata.Effect) bool {
			return e.PointsPerResource != 0 && e.Average() == 0
		})
	}

	// An aura the row states at 0 beside the buff attaches nothing but an
	// aura, so the parse leaves it out: the healing-taken row every paladin
	// aura carries is one.
	for _, applied := range spelldata.DryRun(row.Spell, row.parseOptions()...).Applied {
		e := applied.Effect
		if e.BasePoints == 0 && e.PointsPerResource == 0 && !slices.Contains(row.SkipAuraTypes, e.Aura) {
			row.SkipAuraTypes = append(row.SkipAuraTypes, e.Aura)
		}
	}

	// The constructors parse a buff with no character, as they do a debuff.
	parsed := spelldata.DryRun(row.Spell, row.parseOptions()...)
	row.Applied = parsed.Applied
	row.LeftOut = parsed.SkippedNotes(row.Spell)

	switch {
	case row.Kind == buffmanifest.KindDamageShield:
		if !row.hasDamageShield() {
			row.unsupported("no A_DAMAGE_SHIELD effect on spell %d", row.SpellID)
		}
	case isDriverKind(row.Kind):
		// Driver kinds: the hand-written driver decides what the numbers mean.
	default:
		if len(row.Applied) > 0 {
			return
		}
		if len(row.LeftOut) == 0 {
			row.unsupported("spell %d states no aura effect the parse attaches", row.SpellID)
			return
		}
		row.unsupported("spell %d states no aura effect the parse attaches: %s", row.SpellID,
			strings.Join(row.LeftOut, "; "))
	}
}

// The rank select's states. Each stands in for the row, so a state must attach the amounts the row
// attaches, by kind, and a totem's must summon into a totem slot as the row's cast does.
func resolveRankStates(t *spellTables, row *ResolvedBuff) error {
	want := appliedKinds(row.Applied)
	for _, rank := range row.Ranks {
		state := rankState{SpellID: rank.SpellID, CastSpellID: cmp.Or(rank.CastID, rank.SpellID)}
		for _, id := range []int32{state.SpellID, state.CastSpellID} {
			if _, named := t.Names[id]; !named {
				return fmt.Errorf("rank spell %d is no spell in the game files", id)
			}
		}

		cast := t.row(state.CastSpellID)
		state.Label = strings.TrimSpace(t.Names[state.CastSpellID] + " " + cast.Rank)
		if summonsTotem(cast) != row.Totem {
			return fmt.Errorf("rank spell %d summons a totem where the row's cast %d does not, or the other way round", state.CastSpellID, row.CastSpellID)
		}
		got := appliedKinds(spelldata.DryRun(storeRowSpell(t.row(state.SpellID)), row.parseOptions()...).Applied)
		if !slices.Equal(got, want) {
			return fmt.Errorf("rank spell %d attaches %v where the row attaches %v", state.SpellID, got, want)
		}
		row.RankStates = append(row.RankStates, state)
	}
	return nil
}

func appliedKinds(applied []spelldata.Applied) []string {
	kinds := make([]string, len(applied))
	for i, a := range applied {
		kinds[i] = a.Kind
	}
	slices.Sort(kinds)
	return kinds
}

// The options buffs.Meta.Options states for the row, less the talent's scaling,
// which prices an amount and never changes what is attached. A damage shield's
// own effect is left to newDamageShield, and an item-count row counts its
// amounts, which is what a multiplier cannot be read through.
func (row *ResolvedBuff) parseOptions() []spelldata.ParseOpt {
	skip := row.SkipAuraTypes
	if row.Kind == buffmanifest.KindDamageShield {
		skip = append(slices.Clone(skip), dbcenums.A_DAMAGE_SHIELD)
	}
	opts := spelldata.RaidBuffOptions(skip, row.FullComboPoints)
	if row.Kind == buffmanifest.KindItemCount {
		opts = append(opts, spelldata.Count(1))
	}
	return opts
}

func (row *ResolvedBuff) hasDamageShield() bool {
	return slices.ContainsFunc(row.Spell.Effects, func(e spelldata.Effect) bool {
		return e.Aura == dbcenums.A_DAMAGE_SHIELD && spelldata.AppliesAura(e.Type)
	})
}

// The improving talent the manifest pins, which the store keeps as a ladder of
// its ranks. The one spell modifier of it whose class mask reaches the buff's
// spell is the improvement; one that modifies misc 1 scales the duration,
// anything else the value. Where none reaches it, a tooltip term naming the
// buff's spell prices the value by rank (see tooltipTalentDivisor), or the
// buff's own tooltip shows the talent's amount as $<talent>s1, which the buff's
// row states at full rank, and each rank is its share of the node's ranks.
func resolveTalent(t *spellTables, nodes []traitNode, row *ResolvedBuff) error {
	id := row.Talent
	if id == 0 {
		return nil
	}
	node := slices.IndexFunc(nodes, func(n traitNode) bool { return n.SpellID == id })
	if node < 0 {
		return fmt.Errorf("talent %d is no node of a class tree", id)
	}
	talent := t.row(id)
	position := -1
	for i, e := range talent.Effects {
		if !isSpellModifier(e) || !e.ClassFlags.Matches(row.Spell.ClassFlags) {
			continue
		}
		if position >= 0 {
			return fmt.Errorf("talent %d modifies spell %d with effects %d and %d", id, row.SpellID, position+1, i+1)
		}
		position = i
	}
	if position < 0 {
		divisor := tooltipTalentDivisor(t.Descriptions[id], row.SpellID)
		if divisor == 0 && strings.Contains(t.Descriptions[row.SpellID], fmt.Sprintf("$%ds1", id)) {
			divisor = nodes[node].MaxRanks
		}
		if divisor == 0 {
			return fmt.Errorf("talent %d has no spell modifier that reaches spell %d", id, row.SpellID)
		}
		row.TalentSpellID = id
		row.TalentPosition = 1
		row.TalentDivisor = divisor
		row.TalentRanks = nodes[node].MaxRanks
		return nil
	}

	row.TalentSpellID = id
	row.TalentPosition = int32(position + 1)
	row.TalentScalesDuration = talent.Effects[position].Misc == int32(dbcenums.SPELLMOD_DURATION)
	if !row.TalentScalesDuration && len(row.Applied) == 0 && !row.hasDamageShield() {
		row.warn("talent %q has nothing to scale: the parse attaches no amount of spell %d", talent.Name, row.SpellID)
		return nil
	}
	row.TalentRanks = nodes[node].MaxRanks
	return nil
}

// A talent tooltip term that scales another spell by the talent's rank: ${$m1/-3*$440667m1} is
// rank/3 of spell 440667's effect 1, and the fraction applies to every amount the spell attaches.
// The sign only flips the tooltip's wording; the amount keeps the spell's own sign.
var tooltipRankFraction = regexp.MustCompile(`\$\{\$m1/-?(\d+)\*\$(\d+)m1\}`)

func tooltipTalentDivisor(description string, spellID int32) int32 {
	for _, m := range tooltipRankFraction.FindAllStringSubmatch(description, -1) {
		divisor, _ := strconv.Atoi(m[1])
		id, _ := strconv.Atoi(m[2])
		if int32(id) == spellID && divisor > 0 {
			return int32(divisor)
		}
	}
	return 0
}

// A passive talent that raises an effect's value on the buff's spell prices an
// improved state, which a row naming no talent does not offer.
func warnUnpricedTalents(t *spellTables, nodes []traitNode, row *ResolvedBuff) {
	if row.Talent != 0 {
		return
	}
	for _, node := range nodes {
		talent := t.row(node.SpellID)
		if talent.Attr[dbcenums.ATTR_INDEX_BASE]&dbcenums.ATTR_PASSIVE == 0 {
			continue
		}
		for i, e := range talent.Effects {
			if isSpellModifier(e) && raisesValue(e.Misc) && e.ClassFlags.Matches(row.Spell.ClassFlags) {
				row.warn("talent %d %q effect %d raises spell %d (misc %d), and the manifest names no Talent",
					node.SpellID, talent.Name, i+1, row.SpellID, e.Misc)
			}
		}
	}
}

func isSpellModifier(e storeEffect) bool {
	return e.Aura == dbcenums.A_ADD_FLAT_MODIFIER || e.Aura == dbcenums.A_ADD_PCT_MODIFIER
}

func raisesValue(misc int32) bool {
	switch dbcenums.SpellModOp(misc) {
	case dbcenums.SPELLMOD_EFFECT1, dbcenums.SPELLMOD_ALL_EFFECTS, dbcenums.SPELLMOD_EFFECT2, dbcenums.SPELLMOD_EFFECT3:
		return true
	}
	return false
}

// The group spell of each raid input that reads the single-target spell's numbers.
var groupInputSpells = map[string]int32{
	"mark_of_the_wild":     21850,
	"power_word_fortitude": 21564,
	"arcane_intellect":     23028,
	"shadow_protection":    27683,
}

// Who the game files say the aura reaches. The manifest wins - the sim's scopes are
// a UI grouping as much as a game fact - but a disagreement is worth printing.
func validateScope(row *ResolvedBuff) {
	scope, known := clientScope(*row)
	if !known {
		return
	}
	row.ScopeFromClient = scope
	if scope != row.Scope {
		row.warn("scope mismatch: the manifest says %s, spell %d reads as %s", row.Scope, row.ScopeSpell.ID, scope)
	}
}

// The group the game files state a buff reaches: an area aura names it in the
// effect itself, and an aura the game files apply over an area or on one ally
// names it in the effect's target.
func clientScope(row ResolvedBuff) (buffmanifest.BuffScope, bool) {
	if row.ScopeSpell == nil {
		return buffmanifest.ScopeIndividual, false
	}
	for _, effect := range row.ScopeSpell.Effects {
		switch effect.Type {
		case dbcenums.E_APPLY_AREA_AURA_RAID:
			return buffmanifest.ScopeRaid, true
		case dbcenums.E_APPLY_AREA_AURA_PARTY:
			return buffmanifest.ScopeParty, true
		case dbcenums.E_APPLY_AURA:
			switch effect.Target[0] {
			case dbcenums.TARGET_UNIT_CASTER_AREA_RAID:
				return buffmanifest.ScopeRaid, true
			case dbcenums.TARGET_UNIT_CASTER_AREA_PARTY:
				return buffmanifest.ScopeParty, true
			case dbcenums.TARGET_UNIT_TARGET_ALLY, dbcenums.TARGET_153:
				return buffmanifest.ScopeIndividual, true
			}
		}
	}
	return buffmanifest.ScopeIndividual, false
}

// The repository root, found by walking up from the working directory until
// go.mod appears. The generator runs from the root and the regeneration test
// from tools/database, and both have to reach sim/core/buffs.
func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("no go.mod above the working directory, so the repository root is unknown")
		}
		dir = parent
	}
}

// The proto message each scope's fields live on. The generated apply blocks read
// the field through the type the sim compiled against, so the manifest's declared
// type is checked against this rather than against proto/buffs.proto: a row
// whose proto field has not been retyped yet would otherwise generate code that
// cannot compile.
var buffScopeMessages = map[buffmanifest.BuffScope]reflect.Type{
	buffmanifest.ScopeRaid:       reflect.TypeOf(proto.RaidBuffs{}),
	buffmanifest.ScopeParty:      reflect.TypeOf(proto.PartyBuffs{}),
	buffmanifest.ScopeIndividual: reflect.TypeOf(proto.IndividualBuffs{}),
	buffmanifest.ScopeDebuff:     reflect.TypeOf(proto.Debuffs{}),
}

var buffProtoTypeNames = map[buffmanifest.BuffProtoType]string{
	buffmanifest.ProtoBool:     "bool",
	buffmanifest.ProtoTristate: "proto.TristateEffect",
	buffmanifest.ProtoInt32:    "int32",
	buffmanifest.ProtoDouble:   "float64",
}

// The Go type the compiled proto states for a field, spelled the way the manifest
// spells it.
func compiledProtoType(row ResolvedBuff) (string, error) {
	message, ok := buffScopeMessages[row.Scope]
	if !ok {
		return "", fmt.Errorf("unknown scope %s", row.Scope)
	}
	field, ok := message.FieldByName(row.GoField())
	if !ok {
		return "", fmt.Errorf("%s has no field %s", message.Name(), row.GoField())
	}
	switch field.Type.Kind() {
	case reflect.Bool:
		return "bool", nil
	case reflect.Float64:
		return "float64", nil
	case reflect.Int32:
		if name := field.Type.Name(); name != "int32" {
			return "proto." + name, nil
		}
		return "int32", nil
	}
	return field.Type.String(), nil
}

func isDriverKind(kind buffmanifest.BuffKind) bool {
	switch kind {
	case buffmanifest.KindExternalCD, buffmanifest.KindProc, buffmanifest.KindManual:
		return true
	}
	return false
}

func lowerFirst(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}

///////////////////////////////////////////////////////////////////////////
//							Rendering
///////////////////////////////////////////////////////////////////////////

// buffRow is what the templates see: every expression the generated file needs,
// already spelled as Go source.
type buffRow struct {
	Go          string
	Builder     string
	Field       string
	Label       string
	SpellID     int32
	Kind        string
	Reason      string
	LeftOut     []string
	Supported   bool
	HasSpell    bool
	SpellVar    string
	MetaVar     string
	MetaFields  string
	Category    string
	CategoryVar string
	HasValue    bool
	HasCooldown bool
	ExtraParams string
	Constructor string
	ApplyIf     string
	ApplyBody   string
	HasApply    bool
	RanksVar    string
	RankStates  []string
}

// Every file the manifest renders, by the path it is written to: the two Go
// files, the settings inputs and the proto messages.
func renderBuffOutputs(in *storeInputs) (map[string][]byte, error) {
	rows, err := resolveBuffManifest(in)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		for _, warning := range row.Warnings {
			fmt.Fprintf(progress, "buffs: %s: %s\n", row.Field, warning)
		}
	}

	files, err := renderBuffFiles(rows)
	if err != nil {
		return nil, err
	}
	if files[buffsDebuffsTSFile], err = RenderBuffsDebuffsTS(rows); err != nil {
		return nil, err
	}
	files[buffsProtoFile] = buffmanifest.RenderProto()
	return files, nil
}

const buffsProtoFile = "proto/buffs.proto"

func renderBuffFiles(rows []ResolvedBuff) (map[string][]byte, error) {
	buffs, err := renderBuffFile(rows, false)
	if err != nil {
		return nil, err
	}
	debuffs, err := renderBuffFile(rows, true)
	if err != nil {
		return nil, err
	}
	return map[string][]byte{buffsGenFile: buffs, debuffsGenFile: debuffs}, nil
}

func renderBuffFile(resolved []ResolvedBuff, debuffs bool) ([]byte, error) {
	var rows []buffRow
	var shared []sharedCategoryRow
	needsSpells, needsEnums := false, false
	for _, row := range resolved {
		if (row.Scope == buffmanifest.ScopeDebuff) != debuffs {
			continue
		}
		rendered := renderRow(row)
		if rendered.Supported {
			needsSpells = true
			needsEnums = needsEnums || len(row.SkipAuraTypes) > 0
			if row.SharedCategory != "" && !slices.ContainsFunc(shared, func(c sharedCategoryRow) bool { return c.Name == row.SharedCategory }) {
				shared = append(shared, sharedCategoryRow{Var: sharedCategoryVar(row.SharedCategory), Name: row.SharedCategory})
			}
		}
		rows = append(rows, rendered)
	}
	slices.SortFunc(shared, func(a, b sharedCategoryRow) int { return strings.Compare(a.Name, b.Name) })

	name := "buffs"
	tmplStr := TmplStrBuffs
	if debuffs {
		name, tmplStr = "debuffs", TmplStrDebuffs
	}

	tmpl, err := template.New(name).Parse(tmplStr)
	if err != nil {
		return nil, err
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, map[string]any{
		"Rows": rows, "NeedsSpells": needsSpells, "NeedsEnums": needsEnums,
		"SharedCategories": shared,
	}); err != nil {
		return nil, fmt.Errorf("rendering %s: %w", name, err)
	}

	// The template cannot indent a commented-out shell the way gofmt wants, and
	// writing unformatted Go would make every later run diff on whitespace.
	out, err := format.Source(rendered.Bytes())
	if err != nil {
		return nil, fmt.Errorf("generated %s is not valid Go: %w", name, err)
	}
	return out, nil
}

// sharedCategoryRow is one `var <Name>Category = "<Name>"` line: several rows
// name the same shared category, so the generated file declares it once.
type sharedCategoryRow struct {
	Var  string
	Name string
}

// The identifier the generated file gives a shared category, which every row
// that joins it and every caller in sim/core/buffs name.
func sharedCategoryVar(category string) string {
	return category + "Category"
}

func buffScopeField(scope buffmanifest.BuffScope) string {
	switch scope {
	case buffmanifest.ScopeRaid:
		return "raid"
	case buffmanifest.ScopeParty:
		return "party"
	case buffmanifest.ScopeDebuff:
		return "debuffs"
	}
	return "individual"
}

func renderRow(row ResolvedBuff) buffRow {
	out := buffRow{
		Go: row.Go, Builder: buffBuilder(row.Go), Field: row.Field, Label: buffLabel(row),
		SpellID: row.SpellID, Reason: row.Reason, LeftOut: row.LeftOut,
		Supported: row.Supported, HasSpell: row.SpellID != 0,
	}
	if row.Kind != buffmanifest.KindPlain {
		out.Kind = row.Kind.String()
	}

	// A row whose amounts are worth one item each takes the number of them, so
	// that a party with three of the same staff gets three times the aura.
	if row.Kind == buffmanifest.KindItemCount {
		out.ExtraParams = ", count float64"
	}

	if !row.Supported {
		if out.Reason == "" {
			out.Reason = "no reason given"
		}
		return out
	}

	if row.Category != "" {
		out.Category = row.Category
		out.CategoryVar = row.Go + "Category"
	}
	out.SpellVar = lowerFirst(row.Go) + "Spell"
	out.MetaVar = lowerFirst(row.Go) + "Meta"
	out.MetaFields = buffMetaFields(row, out)
	out.HasValue = len(row.Applied) > 0 || row.Kind == buffmanifest.KindDamageShield
	out.HasCooldown = row.CooldownMs > 0
	out.Constructor = buffConstructor(row, out)
	if len(row.RankStates) > 0 {
		out.RanksVar = buffRanksVar(row)
		for _, state := range row.RankStates {
			out.RankStates = append(out.RankStates, fmt.Sprintf("%s.atRank(spelldata.MustFind(%d)),", out.MetaVar, state.SpellID))
		}
	}
	out.ApplyIf, out.ApplyBody, out.HasApply = buffApply(row, out.SpellVar)
	return out
}

func buffRanksVar(row ResolvedBuff) string {
	return lowerFirst(row.Go) + "Ranks"
}

// The constructor's name: the stem with the NewSpellData prefix, and Aura after it unless the stem
// already ends in Aura.
func buffBuilder(stem string) string {
	if strings.HasSuffix(stem, "Aura") {
		return "NewSpellData" + stem
	}
	return "NewSpellData" + stem + "Aura"
}

// The label the generated Meta and the settings input both show, which is the
// game files' name for the spell unless the manifest overrides it.
func buffLabel(row ResolvedBuff) string {
	if row.Label != "" {
		return row.Label
	}
	if row.Name != "" {
		return row.Name
	}
	return row.DBName
}

// The fields of the row's buffs.Meta literal, one per line, leaving out the
// ones at their zero value.
func buffMetaFields(row ResolvedBuff, rendered buffRow) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Label: %q,\n", rendered.Label)
	fmt.Fprintf(&b, "Spell: %s,\n", rendered.SpellVar)
	if row.CastSpellID != row.SpellID && row.CooldownMs > 0 {
		fmt.Fprintf(&b, "Cast: spelldata.MustFind(%d),\n", row.CastSpellID)
	}
	if rendered.CategoryVar != "" {
		fmt.Fprintf(&b, "Category: %s,\n", rendered.CategoryVar)
	}
	if row.SharedCategory != "" {
		fmt.Fprintf(&b, "SharedCategory: %s,\n", sharedCategoryVar(row.SharedCategory))
	}
	if row.SingleAura {
		b.WriteString("SingleAura: true,\n")
	}
	if row.TalentRanks > 0 {
		fmt.Fprintf(&b, "Talent: spelldata.Talent(%d, %d),\n", row.TalentSpellID, row.TalentRanks)
		fmt.Fprintf(&b, "TalentEffect: %d,\n", row.TalentPosition)
		if row.TalentDivisor > 0 {
			fmt.Fprintf(&b, "TalentDivisor: %d,\n", row.TalentDivisor)
		}
		if row.TalentScalesDuration {
			b.WriteString("TalentScalesDuration: true,\n")
		}
	}
	if len(row.SkipAuraTypes) > 0 {
		names := make([]string, len(row.SkipAuraTypes))
		for i, aura := range row.SkipAuraTypes {
			name, _ := dbcenums.Named(aura)
			names[i] = "dbcenums." + name
		}
		fmt.Fprintf(&b, "SkipAuras: []dbcenums.EffectAuraType{%s},\n", strings.Join(names, ", "))
	}
	if row.FullComboPoints {
		b.WriteString("FullComboPoints: true,\n")
	}
	return b.String()
}

// The call the constructor makes into sim/core/buffs/meta.go.
func buffConstructor(row ResolvedBuff, rendered buffRow) string {
	switch {
	case row.Kind == buffmanifest.KindDamageShield:
		return fmt.Sprintf("return newDamageShield(unit, %s, isPlayer, talentPoints)", rendered.MetaVar)
	case row.Scope == buffmanifest.ScopeDebuff:
		return fmt.Sprintf("return newDebuff(unit, %s, isPlayer, talentPoints)", rendered.MetaVar)
	case row.Kind == buffmanifest.KindItemCount:
		return fmt.Sprintf("return newItemCountBuff(unit, %s, isPlayer, count)", rendered.MetaVar)
	}
	return fmt.Sprintf("return newBuff(unit, %s, isPlayer, talentPoints)", rendered.MetaVar)
}

// The apply block: the condition the proto field is read by, and the call that
// puts the buff on the unit. An int32 field a talent prices is the talent's
// rank, up to the node's ranks. A kind whose behaviour is a cooldown, a proc or an
// uptime calls a driver of a fixed name that sim/core/buffs/drivers.go declares,
// and so does a row the manifest marks as driven. A driver is handed the whole
// scope message rather than its own field, because a driven buff often reads a
// second one: Flametongue Totem does nothing under Windfury Totem. Any other totem
// stands through standTotem, which puts it on the character's true pets too
// unless the row's spell says only a player can carry it. A rank select applies
// the state its field holds in place of the row.
func buffApply(row ResolvedBuff, spellVar string) (string, string, bool) {
	unit, field := "char", buffScopeField(row.Scope)
	if row.Scope == buffmanifest.ScopeDebuff {
		unit = "target"
	}
	access := field + "." + row.GoField()

	if len(row.RankStates) > 0 {
		state := fmt.Sprintf("%s.at(%s)", buffRanksVar(row), access)
		if row.Totem {
			return access + " > 0", fmt.Sprintf("standRankedTotem(char, %s)", state), true
		}
		return access + " > 0", fmt.Sprintf("core.MakePermanent(newBuff(&char.Unit, %s, false, 0))", state), true
	}

	var cond, points string
	switch row.Proto {
	case buffmanifest.ProtoBool:
		cond, points = access, "0"
	case buffmanifest.ProtoTristate:
		cond = access + " != proto.TristateEffect_TristateEffectMissing"
		points = fmt.Sprintf("core.GetTristateValueInt32(%s, 0, %d)", access, row.TalentRanks)
	case buffmanifest.ProtoInt32, buffmanifest.ProtoDouble:
		cond, points = access+" > 0", "0"
		if row.Proto == buffmanifest.ProtoInt32 && row.TalentRanks > 0 {
			points = fmt.Sprintf("min(%s, %d)", access, row.TalentRanks)
		}
	default:
		return "", "", false
	}

	var body string
	switch {
	case row.Driver, isDriverKind(row.Kind), row.Kind == buffmanifest.KindItemCount:
		scope := field
		if row.Scope == buffmanifest.ScopeDebuff {
			scope = "debuffs, raid"
		}
		body = fmt.Sprintf("drive%s(%s, %s)", row.Go, unit, scope)
	case row.Totem && row.Scope != buffmanifest.ScopeDebuff:
		body = fmt.Sprintf("standTotem(char, %s, %s, %s)", spellVar, buffBuilder(row.Go), points)
	default:
		target := "&char.Unit"
		if row.Scope == buffmanifest.ScopeDebuff {
			target = "target"
		}
		body = fmt.Sprintf("core.MakePermanent(%s(%s, false, %s))", buffBuilder(row.Go), target, points)
	}
	return cond, body, true
}

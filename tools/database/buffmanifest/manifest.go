// Package buffmanifest is the checked-in manifest of every raid, party, individual
// and enemy-debuff proto field the sim knows about. It is a leaf package: it may
// import the standard library and sim/core/proto and nothing else, so the buff
// generator and the proto emitter can both read it while the generated proto and
// sim/core files are stale.
package buffmanifest

import (
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
)

type BuffScope int

const (
	ScopeRaid BuffScope = iota
	ScopeParty
	ScopeIndividual
	ScopeDebuff
)

func (s BuffScope) String() string {
	switch s {
	case ScopeRaid:
		return "ScopeRaid"
	case ScopeParty:
		return "ScopeParty"
	case ScopeIndividual:
		return "ScopeIndividual"
	case ScopeDebuff:
		return "ScopeDebuff"
	}
	return "BuffScope(unknown)"
}

type BuffProtoType int

// The zero BuffProtoType is no override: the row's type is derived, see ProtoType.
const (
	ProtoBool BuffProtoType = iota + 1
	ProtoTristate
	ProtoInt32
	ProtoDouble
)

func (t BuffProtoType) String() string {
	switch t {
	case ProtoBool:
		return "ProtoBool"
	case ProtoTristate:
		return "ProtoTristate"
	case ProtoInt32:
		return "ProtoInt32"
	case ProtoDouble:
		return "ProtoDouble"
	}
	return "BuffProtoType(unknown)"
}

// BuffKind names a row whose behaviour is not simply the aura its spell states. A plain row states
// none, and a damage shield is read off the spell.
type BuffKind int

const (
	KindPlain BuffKind = iota
	KindDamageShield
	KindExternalCD
	KindProc
	KindItemCount
	KindManual
	KindFlag
)

func (k BuffKind) String() string {
	switch k {
	case KindPlain:
		return "KindPlain"
	case KindDamageShield:
		return "KindDamageShield"
	case KindExternalCD:
		return "KindExternalCD"
	case KindProc:
		return "KindProc"
	case KindItemCount:
		return "KindItemCount"
	case KindManual:
		return "KindManual"
	case KindFlag:
		return "KindFlag"
	}
	return "BuffKind(unknown)"
}

type ActionRef struct {
	SpellID int32
	ItemID  int32
}

// BuffSpec is what the game files cannot state about a proto field. Its scope is the slice it sits in,
// its number its position there, and its names, owner and left-out auras are read off its spell.
type BuffSpec struct {
	Field string // proto field name, snake_case
	// SpellID is the spell the aura's numbers are read from: the top rank of the castable family (the
	// trainer's top rank where only a class book teaches the rank above it), or the aura that family's
	// cast applies when the cast is a summon or a dummy. CastID is that cast,
	// the spell the player learns, which names the buff and times an external cooldown. Both are
	// roots of the spell store.
	SpellID int32
	CastID  int32
	// TooltipID is the spell the settings input shows as its icon and tooltip, for a row whose aura
	// states no text and whose CastID is not that spell either. 0 defers to CastID, then SpellID; it
	// never changes the numbers the row reads.
	TooltipID int32
	// Talent is the spell of the trait node that prices the improved state, 0 when none does.
	Talent         int32
	Category       string // exclusive-effect category value, "" = none
	SharedCategory string // second exclusive category the aura also joins, "" = none
	SingleAura     bool
	Driver         bool         // apply block hands the field to drive<Go>; the aura is not simply always up
	Stats          []proto.Stat // UI relevance tags
	// ImpAction names the improved state's source when it is not a talent: an
	// item, or the spell an item set grants at a piece threshold. It is the icon
	// the improved state shows.
	ImpAction *ActionRef
	Label     string // UI label override, "" = the spell's name
	// Owner is the class that provides a buff whose spell names no class family.
	Owner proto.Class
	Kind  BuffKind
	Proto BuffProtoType // override of the derived type
	// Reason says what a Flag row is, since no spell does; the shell it renders as carries it.
	Reason string
}

// GoStem is the identifier stem the generated code names the row by: GoField without underscores.
func (s BuffSpec) GoStem() string {
	return strings.ReplaceAll(s.GoField(), "_", "")
}

// ProtoType is the field's proto type: a tristate when a talent or an action prices an improved
// state, a count for an external cooldown or an item, and a bool otherwise.
func (s BuffSpec) ProtoType() BuffProtoType {
	switch {
	case s.Proto != 0:
		return s.Proto
	case s.Talent != 0 || s.ImpAction != nil:
		return ProtoTristate
	case s.Kind == KindExternalCD || s.Kind == KindItemCount:
		return ProtoInt32
	}
	return ProtoBool
}

// GoField is the field name protoc-gen-go generates for Field. It ports
// protobuf-go's strs.GoCamelCase, including the branch that keeps the underscore
// in front of a digit ("soe_enhancement_2pt4" gives "SoeEnhancement_2Pt4").
func (s BuffSpec) GoField() string {
	var b []byte
	for i := 0; i < len(s.Field); i++ {
		c := s.Field[i]
		switch {
		case c == '_' && i+1 < len(s.Field) && isASCIILower(s.Field[i+1]):
		case isASCIIDigit(c):
			b = append(b, c)
		default:
			if isASCIILower(c) {
				c -= 'a' - 'A'
			}
			b = append(b, c)
			for ; i+1 < len(s.Field) && isASCIILower(s.Field[i+1]); i++ {
				b = append(b, s.Field[i+1])
			}
		}
	}
	return string(b)
}

// TSField is the property name protobuf-ts generates for Field: lower camel case
// in which a digit also capitalises the character that follows it
// ("soe_enhancement_2pt4" gives "soeEnhancement2Pt4").
func (s BuffSpec) TSField() string {
	var b strings.Builder
	capNext := false
	for i := 0; i < len(s.Field); i++ {
		c := s.Field[i]
		switch {
		case c == '_':
			capNext = true
		case isASCIIDigit(c):
			b.WriteByte(c)
			capNext = true
		case capNext:
			if isASCIILower(c) {
				c -= 'a' - 'A'
			}
			b.WriteByte(c)
			capNext = false
		case i == 0:
			if !isASCIILower(c) {
				c += 'a' - 'A'
			}
			b.WriteByte(c)
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// Scopes is the order the rows are resolved, rendered and applied in. The apply order is the order
// the auras register in, which the sim's results follow.
var Scopes = []BuffScope{ScopeParty, ScopeRaid, ScopeIndividual, ScopeDebuff}

// ByScope is the rows of one message, in proto number order.
func ByScope(scope BuffScope) []BuffSpec {
	switch scope {
	case ScopeRaid:
		return Raid
	case ScopeParty:
		return Party
	case ScopeIndividual:
		return Individual
	case ScopeDebuff:
		return Debuffs
	}
	return nil
}

// Row is a manifest row with the scope and the proto number its place in the manifest gives it.
type Row struct {
	BuffSpec
	Scope  BuffScope
	Number int32
}

// All is every row, scope by scope in Scopes order.
func All() []Row {
	var rows []Row
	for _, scope := range Scopes {
		for i, spec := range ByScope(scope) {
			rows = append(rows, Row{BuffSpec: spec, Scope: scope, Number: int32(i + 1)})
		}
	}
	return rows
}

func isASCIILower(c byte) bool { return c >= 'a' && c <= 'z' }

func isASCIIDigit(c byte) bool { return c >= '0' && c <= '9' }

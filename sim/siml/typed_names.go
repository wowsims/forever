package siml

import (
	"fmt"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Names (SPEC.md §7): what a name means in its position's namespace for the file's class, the namespace an ActionID
// field reads, and the raw JSON that holds an id a name could write.

// A name a file writes, as typing reads it.
type nameRef struct {
	span
	name     string
	ns       Namespace
	id       int32
	actionID *proto.ActionID // What a spell-like name writes: a new message of its own.
	row      *SlotRow
	path     []any // The JSON path, from IndividualSimSettings, of the field it fills.
	known    bool  // Whether the namespace holds it.
}

// What a namespace holds, as messages name it: "battle_shout is a spell".
var namespaceThings = map[Namespace]string{
	NamespaceSpell:       "a spell",
	NamespaceAura:        "an aura",
	NamespaceTargetAura:  "an aura on a target",
	NamespaceTargetSpell: "a boss spell",
	NamespacePetAura:     "a pet aura",
	NamespaceItem:        "an item",
	NamespaceEnchant:     "an enchant",
	NamespaceSuffix:      "a random suffix",
}

// What n means in ns for the file (SPEC.md §7.2); false, and unknown-name, when ns holds no such name. position is
// what takes the name, for the message: "this position", a field's name, "a slot"; cast, whether it is the spell an
// action casts, whose did-you-mean offers castable spells only. A name the namespace does not hold goes into refs here;
// the caller adds one it holds.
func (t *typer) lookup(n name, ns Namespace, position string, cast bool) (nameRef, bool) {
	r := nameRef{span: n.span, name: n.text, ns: ns, path: append([]any{}, t.at...)}
	var ok bool
	switch ns {
	case NamespaceItem:
		r.id, ok = t.names.Item(n.text)
	case NamespaceEnchant:
		r.id, ok = t.names.Enchant(n.text)
	case NamespaceSuffix:
		r.id, ok = t.names.Suffix(n.text)
	default:
		r.actionID, ok = t.names.Spell(t.class, ns, n.text)
		r.id = r.actionID.GetSpellId()
	}
	r.known = ok
	if !ok {
		t.report("unknown-name", t.unknownName(n.text, ns, position, cast), n.span)
		t.addRef(r)
	}
	return r, ok
}

// at, or the call in this file that reads it when at lies in a used file's def: a used file's read types a def only
// where it calls it (SPEC.md §6.10), so the names in the def's value are this file's. false when no call here holds it.
func (t *typer) ownSpan(at span) (span, bool) {
	if t.lines == nil || at.end <= len(t.lines.src) {
		return at, true
	}
	for s := t.site; s != nil; s = s.outer {
		if s.call.end <= len(t.lines.src) {
			return s.call, true
		}
	}
	return span{}, false
}

// Adds r to the names the file writes; one in a used file's def to usedRefs, at the call that reads the def.
func (t *typer) addRef(r nameRef) {
	at, ok := t.ownSpan(r.span)
	switch {
	case !ok:
	case at == r.span:
		t.refs = append(t.refs, r)
	default:
		r.span = at
		t.usedRefs = append(t.usedRefs, r)
	}
}

// Why ns holds no name: what the name is where another namespace holds it, else the names it may have meant, castable
// ones where an action casts (cast).
func (t *typer) unknownName(name string, ns Namespace, position string, cast bool) string {
	for _, other := range t.names.Holding(t.class, name) {
		switch {
		case other == ns:
		case other == NamespaceItem && ns.SpellLike():
			return fmt.Sprintf("%s is an item; write item.%[1]s", name)
		default:
			return fmt.Sprintf("%s is %s; %s takes %s", name, namespaceThings[other], position, namespaceThings[ns])
		}
	}
	thing, _ := strings.CutPrefix(namespaceThings[ns], "a ")
	thing, _ = strings.CutPrefix(thing, "an ")
	candidates := t.names.Visible(t.class, ns)
	if cast {
		candidates = slices.DeleteFunc(candidates, func(c string) bool {
			e, _ := t.names.spellEntry(t.class, ns, c)
			return e.Uncastable
		})
	}
	message := fmt.Sprintf("no %s is named %s%s", thing, name, didYouMean(name, candidates))
	if ns.SpellLike() && t.class == proto.Class_ClassUnknown {
		message += "; with no spec, a file reads only the spells and auras every class shares"
	}
	return message
}

// The kind whose own arguments are being typed or printed, and the type of the unit they give it.
type kindScope struct {
	kind *aplKind
	unit proto.UnitReference_Type
}

// The namespace the names of the ActionID field f read (SPEC.md §7.2): a kind's aura_id reads the auras of its unit,
// the ActionID of a boss_spell_<prop> kind what a target casts, and any other ActionID what the player casts.
func (s kindScope) namespace(f *field) Namespace {
	switch {
	case s.kind == nil || f.desc.ContainingMessage() != s.kind.message.desc:
		return NamespaceSpell
	case f.Name == "aura_id":
		return auraNamespace(s.unit)
	case strings.HasPrefix(s.kind.Kind, "boss_spell_"):
		return NamespaceTargetSpell
	}
	return NamespaceSpell
}

// Whether f is the spell an action kind casts: the spell_id of cast_friendly_spell, multidot and the others.
func (s kindScope) casts(f *field) bool {
	return s.kind != nil && f.desc.ContainingMessage() == s.kind.message.desc && f.Name == "spell_id" &&
		s.kind.field.ContainingMessage().FullName() == actionMessage.FullName()
}

// Why a rotation of the file's spec cannot cast name, with the castable spells the spec registers whose names extend
// it or that it extends, which it may have meant: raptor_strike for raptor_strike_tag0.
func (t *typer) notCastable(name string) string {
	var near []string
	for _, n := range t.names.Registered(t.class, t.spec, NamespaceSpell) {
		if !n.Entry.Uncastable && (strings.HasPrefix(name, n.Name+"_") || strings.HasPrefix(n.Name, name+"_")) {
			near = append(near, n.Name)
		}
	}
	message := fmt.Sprintf("%s is a spell no build lets a rotation cast", name)
	if len(near) > 0 {
		message += "; did you mean " + strings.Join(near[:min(len(near), 3)], ", ") + "?"
	}
	return message
}

// Whether the kind finds an aura by its internal cooldown trigger too, as the sim's GetIcdAuraByID does.
func (s kindScope) readsICD() bool {
	if s.kind == nil {
		return false
	}
	switch s.kind.message.desc.FullName() {
	case icdIsReadyMessage.FullName(), triggerICDMessage.FullName(), internalCooldownMessage.FullName():
		return true
	}
	return false
}

var (
	icdIsReadyMessage       = (&proto.APLValueAuraICDIsReady{}).ProtoReflect().Descriptor()
	triggerICDMessage       = (&proto.APLActionTriggerICD{}).ProtoReflect().Descriptor()
	internalCooldownMessage = (&proto.APLValueAuraInternalCooldown{}).ProtoReflect().Descriptor()
)

// The namespace of the auras on a unit of type typ: a pet's, a target's, or else the player's.
func auraNamespace(typ proto.UnitReference_Type) Namespace {
	switch typ {
	case proto.UnitReference_Pet:
		return NamespacePetAura
	case proto.UnitReference_Unknown, proto.UnitReference_Self, proto.UnitReference_Player, proto.UnitReference_AllPlayers:
		return NamespaceAura
	}
	return NamespaceTargetAura
}

// The kind's UnitReference field, the first in declaration order, which says whose auras its aura_id reads; nil when
// it has none.
func (k *aplKind) unitField() *field {
	for _, f := range k.message.fields {
		if f.Printed == "unit" && !f.Repeated {
			return f
		}
	}
	return nil
}

// Types the arguments of the kind k next, written with the unit prefix and the arguments args, either of which may
// be nil; gives what restores the kind before.
func (t *typer) enter(k *aplKind, prefix *name, args *argList) func() {
	outer := t.kind
	t.kind = kindScope{kind: k, unit: t.writtenUnit(k, prefix, args)}
	return func() { t.kind = outer }
}

// The type of the unit written for k's unit field: the prefix, else the first argument that names the field. One that
// does not read, like none written, is Unknown.
func (t *typer) writtenUnit(k *aplKind, prefix *name, args *argList) proto.UnitReference_Type {
	unit := k.unitField()
	var written node
	switch {
	case unit == nil:
	case prefix != nil:
		written = &callNode{span: prefix.span, name: *prefix}
	case args != nil:
		for _, a := range args.items {
			if a.name != nil && k.message.byName[a.name.text] == unit {
				written = a.value
				break
			}
		}
	}
	switch n := written.(type) {
	case *callNode:
		if u := t.ix.unitReference(n.name.text); n.args == nil && u != nil {
			return u.Type
		}
	case *rawNode:
		u := newMessage(unitMessage)
		if t.ix.readJSON(n.json, u, protojson.UnmarshalOptions{}) == "" {
			return u.Interface().(*proto.UnitReference).Type
		}
	}
	return proto.UnitReference_Unknown
}

// Notes the raw token r when the message m read from it holds an id a name could write (raw-ref, SPEC.md §7.3).
func (t *typer) noteRaw(r *rawNode, m protoreflect.Message) {
	if at, ok := t.ownSpan(r.span); ok && t.ix.holdsID(m) && !slices.Contains(t.rawRefs, at) {
		t.rawRefs = append(t.rawRefs, at)
	}
}

// Whether m holds a spell id, an item id, an enchant effect id or a random suffix id.
func (ix *schemaIndex) holdsID(m protoreflect.Message) bool {
	held := false
	walkMessages(m, func(x protoreflect.Message) {
		msg := ix.messages[string(x.Descriptor().FullName())]
		if msg == nil {
			return
		}
		action := x.Descriptor().FullName() == actionIDMessage.FullName()
		for _, f := range msg.fields {
			v, ok := fieldValue(x, f.desc)
			switch {
			case !ok:
			case action && (f.Name == "spell_id" || f.Name == "item_id"):
				held = held || v.Int() != 0
			case f.Printed == "ref" && f.isList():
				for i := range v.List().Len() {
					held = held || intOf(f.desc, v.List().Get(i)) != 0
				}
			case f.Printed == "ref":
				held = held || intOf(f.desc, v) != 0
			}
		}
	})
	return held
}

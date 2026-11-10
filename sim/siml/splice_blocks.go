package siml

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// The panel's edits of when blocks and defs (SPEC.md §6.9, §6.10). A block's condition is joined to each row in it,
// so an edit of the condition is made on those rows too; a block an edit adds or takes apart moves rows in or out of
// it, and the rows keep the condition they read with. A def call an edit writes reads as what the new text gives it.

// A when block as the edits leave it.
type spliceWhen struct {
	stmt    *whenStmt // Its line in the text, or one with no condition for a block an edit adds.
	list    *spliceList
	value   *proto.APLValue
	split   bool // Its condition is an and whose values are its conjuncts, as a when line writes it without parentheses.
	hide    bool
	calls   map[string]spliceCall // The def calls its condition writes, by their path in it.
	forced  map[string]string     // The def calls edits write in its condition, by their path in it.
	adopt   bool                  // An edit wrote a call in its condition, whose value the new text gives.
	changed bool                  // Printed again.
}

// The conjuncts the block joins to the condition of a row in it.
func (st *spliceWhen) conjuncts() []*proto.APLValue {
	if st.split {
		return st.value.GetAnd().GetVals()
	}
	return []*proto.APLValue{st.value}
}

// A def line an edit adds: its text, or the value an extract names, written as the statement it is in writes it.
type spliceDef struct {
	text   string
	name   string
	value  *proto.APLValue
	holder protoreflect.Message
	calls  map[string]spliceCall
	forced map[string]string
	scope  scope
}

var aplValueMessage = (&proto.APLValue{}).ProtoReflect().Descriptor()

// A condition of conjuncts: the one alone, or their and.
func conditionOf(vals []*proto.APLValue) *proto.APLValue {
	switch len(vals) {
	case 0:
		return nil
	case 1:
		return vals[0]
	}
	return andOf(vals)
}

func isSplit(v *proto.APLValue) bool { return len(v.GetAnd().GetVals()) > 1 }

// Removes the def calls at path in a statement, inside it and around it: an edit there changes their values.
func clearCalls(forced map[string]string, calls map[string]spliceCall, path []any) {
	at := pathKey(path)
	related := func(key string) bool { return under(key, at) || under(at, key) }
	maps.DeleteFunc(forced, func(key, _ string) bool { return related(key) })
	maps.DeleteFunc(calls, func(key string, _ spliceCall) bool { return related(key) })
}

// Whether the path key is prefix's or one below it.
func under(key, prefix string) bool {
	return prefix == "[]" || key == prefix || strings.HasPrefix(key, strings.TrimSuffix(prefix, "]")+" ")
}

// p with the def calls a statement writes: those its text writes, while the edits leave their values as they are,
// and those the edits write.
func withCalls(p *printer, msg protoreflect.Message, calls map[string]spliceCall, forced map[string]string) *printer {
	q := *p
	q.calls = map[*proto.APLValue]string{}
	valuesIn(msg, nil, func(path []any, v *proto.APLValue) {
		key := pathKey(path)
		if text, ok := forced[key]; ok {
			q.calls[v] = text
		} else if c, ok := calls[key]; ok && googleProto.Equal(c.value, v) {
			q.calls[v] = c.text
		}
	})
	return &q
}

// A value from its JSON.
func (m *rotationModel) readValue(text []byte, path []any) (*proto.APLValue, string) {
	msg := newMessage(aplValueMessage)
	if problem := m.ix.readJSON(string(text), msg, protojson.UnmarshalOptions{DiscardUnknown: true}); problem != "" {
		return nil, fmt.Sprintf("the value for %s is no APLValue: %s", describePath(path), problem)
	}
	return msg.Interface().(*proto.APLValue), ""
}

// Gives a slot the row item, which it is printed from.
func (m *rotationModel) setItem(s *spliceSlot, item *proto.APLListItem) {
	s.msg = item.ProtoReflect()
	s.json = m.ix.jsonValue(s.msg)
	s.dirty = true
}

// The rows in block w, at any depth, but raw rows, which take no condition from it.
func (m *rotationModel) rowsIn(w *whenStmt) []*spliceSlot {
	var out []*spliceSlot
	for _, s := range m.blocks[w].list.slots {
		if slices.Contains(s.blocks, w) && !(s.unit != nil && isRawRow(s.unit.statement)) {
			out = append(out, s)
		}
	}
	return out
}

func isRawRow(s stmt) bool {
	r, ok := s.(*rowStmt)
	return ok && rawRow(r) != nil
}

// A row's own conjuncts, after the n its blocks join to its condition.
type ownConjuncts struct {
	vals []*proto.APLValue
	n    int
}

// Each row's own conjuncts in its blocks, or why a row is no longer in them. A row in no block has the values of its
// condition's and as its own.
func (m *rotationModel) owns(rows []*spliceSlot) (map[*spliceSlot]ownConjuncts, string) {
	out := map[*spliceSlot]ownConjuncts{}
	for _, s := range rows {
		own, ok := m.fits(s.msg, s.blocks)
		if !ok {
			return nil, "a row of the block no longer reads its condition; edit the block on its own"
		}
		if len(s.blocks) == 0 && len(own) == 1 && isSplit(own[0]) {
			own = own[0].GetAnd().GetVals()
		}
		out[s] = ownConjuncts{vals: own, n: len(m.conjunctsOf(s.blocks))}
	}
	return out, ""
}

// The JSON path in a row of its condition's conjunct i of total.
func conjunctPath(i, total int) []any {
	if total == 1 {
		return []any{"action", "condition"}
	}
	return []any{"action", "condition", "and", "vals", i}
}

// Gives a row the condition its blocks and own give it. The def calls in its own conjuncts move with them.
func (m *rotationModel) rejoin(s *spliceSlot, own ownConjuncts, keep int) {
	item := googleProto.Clone(s.msg.Interface()).(*proto.APLListItem)
	if item.Action == nil {
		item.Action = &proto.APLAction{}
	}
	vals := own.vals[len(own.vals)-keep:]
	joined := m.conjunctsOf(s.blocks)
	item.Action.Condition = conditionOf(slices.Concat(joined, vals))
	moves := map[string][2][]any{}
	for i := range vals {
		from := conjunctPath(own.n+len(own.vals)-keep+i, own.n+len(own.vals))
		moves[pathKey(from)] = [2][]any{from, conjunctPath(len(joined)+i, len(joined)+len(vals))}
	}
	// Calls outside the condition stay where they are, and those of the conjuncts its blocks join are theirs.
	condition := pathKey([]any{"action", "condition"})
	calls, forced := map[string]spliceCall{}, map[string]string{}
	for key, c := range s.calls {
		if !under(key, condition) {
			calls[key] = c
		}
		for prefix, move := range moves {
			if under(key, prefix) {
				c.path = slices.Concat(move[1], c.path[len(move[0]):])
				calls[pathKey(c.path)] = c
			}
		}
	}
	for key, text := range s.forced {
		if !under(key, condition) {
			forced[key] = text
		}
		for prefix, move := range moves {
			if under(key, prefix) {
				var path []any
				for x := range strings.FieldsSeq(strings.Trim(key, "[]")) {
					path = append(path, x)
				}
				forced[pathKey(slices.Concat(move[1], path[len(move[0]):]))] = text
			}
		}
	}
	s.calls, s.forced = calls, forced
	m.setItem(s, item)
}

func (m *rotationModel) applyBlocks(e AplEdit) (bool, string) {
	switch e.Op {
	case "wrap":
		return true, m.wrap(e)
	case "inline":
		return true, m.inline(e)
	case "extract":
		return true, m.extract(e)
	case "defInsert":
		return true, m.defInsert(e)
	case "defDelete":
		return true, m.defDelete(e)
	}
	if len(e.Path) > 0 && e.Path[0] == "whens" {
		return true, m.applyWhen(e)
	}
	if e.Call != "" {
		return true, m.setCall(e)
	}
	return false, ""
}

// The block a ["whens", id, …] path names, and the path after the id.
func (m *rotationModel) whenAt(path []any) (*whenStmt, []any, string) {
	if len(path) < 2 {
		return nil, nil, describePath(path) + " names no when block"
	}
	id, ok := pathIndex(path[1], len(m.whens))
	if !ok || m.dropped[m.whens[id]] {
		return nil, nil, describePath(path[:2]) + " does not exist"
	}
	return m.whens[id], path[2:], ""
}

func (m *rotationModel) applyWhen(e AplEdit) string {
	w, rest, problem := m.whenAt(e.Path)
	if problem != "" {
		return problem
	}
	if problem := m.read(); problem != "" {
		return problem
	}
	switch {
	case len(rest) == 0 && e.Op == "set" && len(e.Value) == 0:
		return m.unwrap(w)
	case len(rest) == 1 && rest[0] == "hide" && e.Op == "set":
		var hide bool
		if len(e.Value) > 0 {
			if err := json.Unmarshal(e.Value, &hide); err != nil {
				return describePath(e.Path) + " takes true or none"
			}
		}
		return m.setHide(w, hide)
	case len(rest) > 0 && rest[0] == "condition":
		return m.editCondition(w, e, rest[1:])
	}
	return describePath(e.Path) + " is no condition or hide of a when block, and a block itself is only set to none"
}

// Takes block w apart: its rows keep the condition they read with, and are written with it.
func (m *rotationModel) unwrap(w *whenStmt) string {
	for _, other := range m.whens {
		if other.parent == w && !m.dropped[other] {
			return "the block holds other blocks; take those apart first"
		}
	}
	rows := m.rowsIn(w)
	for _, s := range rows {
		s.blocks = slices.DeleteFunc(slices.Clone(s.blocks), func(b *whenStmt) bool { return b == w })
		s.pinned, s.dirty = true, true
	}
	// The comments above the when line go above its first row.
	if o := m.blocks[w].list.opens[w]; o != nil && o.unit != nil && len(rows) > 0 {
		rows[0].lead = append(slices.Clone(o.unit.comments), rows[0].lead...)
	}
	m.dropped[w] = true
	return ""
}

// Hides or shows block w with its rows. A row shown again stays hidden when its own line or another block hides it.
func (m *rotationModel) setHide(w *whenStmt, hide bool) string {
	st := m.blocks[w]
	st.hide, st.changed = hide, true
	for _, s := range m.rowsIn(w) {
		item := googleProto.Clone(s.msg.Interface()).(*proto.APLListItem)
		item.Hide = hide || s.ownHide() || m.hidden(s.blocks)
		m.setItem(s, item)
	}
	return ""
}

// An edit of block w's condition at path, made on each row in the block too.
func (m *rotationModel) editCondition(w *whenStmt, e AplEdit, path []any) string {
	st := m.blocks[w]
	owns, problem := m.owns(m.rowsIn(w))
	if problem != "" {
		return problem
	}
	v := m.ix.jsonValue(st.value.ProtoReflect())
	at := slices.Concat([]any{"condition"}, path)
	root := any(map[string]any{"condition": v})
	var value any
	has := len(e.Value) > 0
	if has {
		decoded, err := decodeJSON(e.Value)
		if err != nil {
			return fmt.Sprintf("the value for %s is no JSON: %v", describePath(e.Path), err)
		}
		value = decoded
	}
	switch e.Op {
	case "set":
		if e.Call != "" && !has {
			value, has = map[string]any{}, true
			if old, ok := jsonAt(root, at); ok {
				value = old
			}
		}
		if len(path) == 0 && !has {
			return "a when line needs a condition"
		}
		if len(path) == 0 {
			root = map[string]any{"condition": value}
		} else {
			root, problem = setIn(root, at, value, has, e.Path)
		}
	case "insert":
		root, problem = insertIn(root, at, e.Index, value, e.Path)
	case "move":
		list, _ := jsonAt(root, at)
		items, isList := list.([]any)
		if !isList {
			return describePath(e.Path) + " is no list"
		}
		problem = moveIn(items, e.From, e.To, e.Path)
	default:
		return fmt.Sprintf("no edit is named %q in a when block's condition", e.Op)
	}
	if problem != "" {
		return problem
	}
	text, _ := json.Marshal(root.(map[string]any)["condition"])
	value2, problem := m.readValue(text, e.Path)
	if problem != "" {
		return problem
	}
	clearCalls(st.forced, st.calls, path)
	if e.Call != "" {
		if st.forced == nil {
			st.forced = map[string]string{}
		}
		st.forced[pathKey(path)], st.adopt = e.Call, true
	}
	// A condition that was one conjunct, an and in parentheses, stays one.
	st.split = isSplit(value2) && (st.split || !isSplit(st.value))
	st.value, st.changed = value2, true
	for s, own := range owns {
		m.rejoin(s, own, len(own.vals))
		s.adopt = s.adopt || st.adopt
	}
	return ""
}

// Puts rows e.From to e.To of the list at e.Path in a new block with condition e.Value, or the call e.Call. The rows
// must be in the same blocks. A row whose own conjuncts start with the block's gives them to it.
func (m *rotationModel) wrap(e AplEdit) string {
	t, problem := m.locate(e.Path)
	if problem != "" {
		return problem
	}
	if t.list == nil || t.index >= 0 || !t.list.row || t.list.prepull {
		return "wrap puts rows of the priority list or a group in a when block"
	}
	if problem := m.read(); problem != "" {
		return problem
	}
	list := t.list
	if e.From < 0 || e.From > e.To || e.To >= len(list.slots) {
		return fmt.Sprintf("%s holds %d rows, so rows %d to %d do not go in a block", describePath(e.Path), len(list.slots), e.From, e.To)
	}
	outer := list.slots[e.From].blocks
	rows := list.slots[e.From : e.To+1]
	for _, s := range rows {
		if !slices.Equal(s.blocks, outer) {
			return "the rows are in different when blocks; wrap rows of one block"
		}
		if s.unit != nil && isRawRow(s.unit.statement) {
			return "a raw row takes no condition from a block"
		}
	}
	owns, problem := m.owns(rows)
	if problem != "" {
		return problem
	}
	if len(e.Value) == 0 && e.Call == "" {
		return "a when block needs a condition"
	}
	value := &proto.APLValue{}
	if len(e.Value) > 0 {
		if value, problem = m.readValue(e.Value, e.Path); problem != "" {
			return problem
		}
	}
	var parent *whenStmt
	if len(outer) > 0 {
		parent = outer[len(outer)-1]
	}
	w := &whenStmt{parent: parent}
	st := &spliceWhen{stmt: w, list: list, value: value, split: isSplit(value), calls: map[string]spliceCall{}, forced: map[string]string{}, changed: true}
	if e.Call != "" {
		st.forced[pathKey(nil)], st.adopt = e.Call, true
	}
	m.blocks[w] = st
	m.whens = append(m.whens, w)
	if list.opens == nil {
		list.opens = map[*whenStmt]*spliceSlot{}
	}
	if list.lines == nil {
		list.lines = slices.Clone(list.old)
	}
	list.opens[w] = &spliceSlot{when: w}
	added := st.conjuncts()
	for _, s := range rows {
		own := owns[s]
		keep := len(own.vals)
		if !st.adopt && keep >= len(added) && slices.EqualFunc(own.vals[:len(added)], added, m.ix.sameValue) {
			keep -= len(added)
		}
		s.blocks = append(slices.Clone(outer), w)
		s.pinned, s.adopt = true, st.adopt
		m.rejoin(s, own, keep)
	}
	return ""
}

// The statement a path into a row or var line leads to, and the path in it.
func (m *rotationModel) slotAt(path []any) (*spliceSlot, []any, string) {
	t, problem := m.locate(path)
	if problem != "" {
		return nil, nil, problem
	}
	if t.list == nil || t.index < 0 {
		return nil, nil, describePath(path) + " is no value in a row or var line"
	}
	if problem := m.read(); problem != "" {
		return nil, nil, problem
	}
	return t.list.slots[t.index], t.rest, ""
}

// Writes the def call e.Call at e.Path, in a row, var line or when block's condition.
func (m *rotationModel) setCall(e AplEdit) string {
	s, rest, problem := m.slotAt(e.Path)
	if problem != "" {
		return problem
	}
	if len(rest) == 0 {
		return describePath(e.Path) + " is a whole row or var line, which no def call writes"
	}
	if len(e.Value) == 0 && valueAt(s.msg, rest) == nil {
		root, problem := setIn(s.json, rest, map[string]any{}, true, e.Path)
		if problem != "" {
			return problem
		}
		s.json, s.msg = root, nil
		if problem := m.read(); problem != "" {
			return problem
		}
	} else if len(e.Value) > 0 {
		value, err := decodeJSON(e.Value)
		if err != nil {
			return fmt.Sprintf("the value for %s is no JSON: %v", describePath(e.Path), err)
		}
		if problem := changed(s, func(root any) (any, string) { return setIn(root, rest, value, true, e.Path) }); problem != "" {
			return problem
		}
		if problem := m.read(); problem != "" {
			return problem
		}
	}
	clearCalls(s.forced, s.calls, rest)
	if s.forced == nil {
		s.forced = map[string]string{}
	}
	s.forced[pathKey(rest)] = e.Call
	s.adopt, s.dirty = true, true
	return ""
}

// Writes the def call at e.Path as the value it reads as.
func (m *rotationModel) inline(e AplEdit) string {
	if len(e.Path) > 0 && e.Path[0] == "whens" {
		w, rest, problem := m.whenAt(e.Path)
		if problem != "" {
			return problem
		}
		st := m.blocks[w]
		if len(rest) == 0 || rest[0] != "condition" {
			return describePath(e.Path) + " is no def call"
		}
		key := pathKey(rest[1:])
		if _, ok := st.calls[key]; !ok && st.forced[key] == "" {
			return describePath(e.Path) + " is no def call"
		}
		delete(st.calls, key)
		delete(st.forced, key)
		st.changed = true
		return ""
	}
	s, rest, problem := m.slotAt(e.Path)
	if problem != "" {
		return problem
	}
	key := pathKey(rest)
	if _, ok := s.calls[key]; !ok && s.forced[key] == "" {
		return describePath(e.Path) + " is no def call"
	}
	delete(s.calls, key)
	delete(s.forced, key)
	s.dirty = true
	return ""
}

// Why a new def cannot be named name, or "".
func (m *rotationModel) defNameProblem(name string) string {
	if why := identifierRefusal("def", name); why != "" {
		return why
	}
	if m.f.typer.defs.byName[name] != nil || slices.ContainsFunc(m.newDefs, func(d *spliceDef) bool { return d.name == name }) {
		return "a def named " + name + " is written already"
	}
	if what := m.f.typer.meaning(name); what != "" {
		return fmt.Sprintf("%s is %s; a def takes a name of its own", name, what)
	}
	return ""
}

// Names the value at e.Path with a new def e.Name, which the value is then written as a call of.
func (m *rotationModel) extract(e AplEdit) string {
	if problem := m.defNameProblem(e.Name); problem != "" {
		return problem
	}
	d := &spliceDef{name: e.Name}
	if len(e.Path) > 0 && e.Path[0] == "whens" {
		w, rest, problem := m.whenAt(e.Path)
		if problem != "" {
			return problem
		}
		if len(rest) == 0 || rest[0] != "condition" {
			return describePath(e.Path) + " is no value"
		}
		st := m.blocks[w]
		if d.value = valueAt(st.value.ProtoReflect(), rest[1:]); d.value == nil {
			return describePath(e.Path) + " is no value"
		}
		// A call reads as one conjunct, so a block whose condition the def takes whole joins it to its rows as one.
		if len(rest) == 1 && st.split {
			rows := m.rowsIn(w)
			owns, problem := m.owns(rows)
			if problem != "" {
				return problem
			}
			st.split = false
			for s, own := range owns {
				m.rejoin(s, own, len(own.vals))
			}
		}
		d.holder, d.calls, d.forced, d.scope = st.value.ProtoReflect(), st.calls, maps.Clone(st.forced), groupScope(st.list.group)
		if st.forced == nil {
			st.forced = map[string]string{}
		}
		st.forced[pathKey(rest[1:])], st.changed = e.Name, true
	} else {
		t, _ := m.locate(e.Path)
		s, rest, problem := m.slotAt(e.Path)
		if problem != "" {
			return problem
		}
		if d.value = valueAt(s.msg, rest); d.value == nil || len(rest) == 0 {
			return describePath(e.Path) + " is no value"
		}
		d.holder, d.calls, d.forced, d.scope = s.msg, s.calls, maps.Clone(s.forced), groupScope(t.list.group)
		if s.forced == nil {
			s.forced = map[string]string{}
		}
		s.forced[pathKey(rest)], s.dirty = e.Name, true
	}
	m.newDefs = append(m.newDefs, d)
	return ""
}

// `def name(p, …) = value`, `def name = value` with no parameters.
func defText(name string, params []string, value string) string {
	head := "def " + name
	if len(params) > 0 {
		head += "(" + strings.Join(params, ", ") + ")"
	}
	lines := strings.Split(value, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return head + " = " + strings.TrimSpace(strings.Join(lines, " "))
}

// Adds the def line e.Name(e.Params) = e.Text.
func (m *rotationModel) defInsert(e AplEdit) string {
	if problem := m.defNameProblem(e.Name); problem != "" {
		return problem
	}
	if strings.TrimSpace(e.Text) == "" {
		return "a def needs a value"
	}
	m.newDefs = append(m.newDefs, &spliceDef{name: e.Name, text: defText(e.Name, e.Params, e.Text)})
	return ""
}

// Removes the def e.Name, its calls written as the values they read as. A def another def calls stays.
func (m *rotationModel) defDelete(e AplEdit) string {
	d := m.f.typer.defs.byName[e.Name]
	if d == nil || m.gone[d] {
		return "no def is named " + e.Name
	}
	for _, other := range m.f.typer.defs.byName {
		if other == d || m.gone[other] {
			continue
		}
		calls := false
		walkTree(other.value, nil, func(n, _ any) {
			if c, ok := n.(*callNode); ok && c.name.text == e.Name && !other.isParam(c.name.text) {
				calls = true
			}
		})
		if calls {
			return fmt.Sprintf("the def %s calls %s; remove that call first", other.name.text, e.Name)
		}
	}
	for _, list := range m.lists() {
		for _, s := range list.slots {
			for key, c := range s.calls {
				if c.name == e.Name {
					delete(s.calls, key)
					s.dirty = true
				}
			}
		}
	}
	for _, st := range m.blocks {
		for key, c := range st.calls {
			if c.name == e.Name {
				delete(st.calls, key)
				st.changed = true
			}
		}
	}
	m.gone[d] = true
	return ""
}

// The rotation path of a list.
func (m *rotationModel) listPath(list *spliceList) []any {
	if list.group == nil {
		return []any{list.key}
	}
	return []any{"groups", slices.Index(m.groups, list.group), list.key}
}

// The elements the rotation read back holds in place of a list's.
func (m *rotationModel) readBack(got *proto.APLRotation, list *spliceList) []protoreflect.Message {
	if list.group == nil {
		return elements(got.ProtoReflect(), list.key)
	}
	if g := slices.Index(m.groups, list.group); g < len(got.GetGroups()) {
		return elements(got.GetGroups()[g].ProtoReflect(), list.key)
	}
	return nil
}

// Takes what the new text reads as for each statement an edit wrote a def call in, and for each block whose condition
// holds one; true when it took any.
func (m *rotationModel) adopt(back *typedFile, got *proto.APLRotation) bool {
	took := false
	stmts := map[string]stmt{}
	for s, path := range back.paths {
		stmts[pathKey(path)] = s
	}
	for _, w := range m.whens {
		st := m.blocks[w]
		if !st.adopt || m.dropped[w] {
			continue
		}
		k := slices.IndexFunc(st.list.slots, func(s *spliceSlot) bool { return slices.Contains(s.blocks, w) })
		if k < 0 {
			continue
		}
		depth := slices.Index(st.list.slots[k].blocks, w)
		r, ok := stmts[pathKey(append(m.listPath(st.list), k))].(*rowStmt)
		if !ok {
			continue
		}
		read := blocksOf(r)
		if depth >= len(read) || back.whens[read[depth]] == nil {
			continue
		}
		st.value, st.adopt, took = back.whens[read[depth]], false, true
	}
	for _, list := range m.lists() {
		read := m.readBack(got, list)
		if len(read) != len(list.slots) {
			continue
		}
		for k, s := range list.slots {
			if !s.adopt {
				continue
			}
			s.msg, s.adopt, took = read[k], false, true
			s.json = m.ix.jsonValue(s.msg)
		}
	}
	return took
}

// The def lines edits add and remove. New ones go after the last var or def line at the top of the file, or after its
// top-level lines.
func (w *spliceWriter) defs() {
	m := w.model
	for d := range m.gone {
		u := unitOf(d, nil, w.lines)
		w.out.drop(u.first, u.last)
	}
	if len(m.newDefs) == 0 {
		return
	}
	var lines []spliceLine
	for _, d := range m.newDefs {
		text := d.text
		if text == "" {
			p := withCalls(w.printer, d.holder, d.calls, d.forced)
			text = defText(d.name, nil, strings.Join(p.value(d.value.ProtoReflect(), precTop, d.scope), ""))
		}
		lines = append(lines, spliceLine{text: text, from: -1})
	}
	last := -1
	for _, s := range slices.Concat(w.tree.top, w.tree.sections[0].body) {
		switch s := s.(type) {
		case *varStmt:
			last = max(last, w.lines.line(s.end))
		case *defStmt:
			if !m.gone[s] {
				last = max(last, w.lines.line(s.end))
			}
		}
	}
	if last >= 0 {
		w.out.insert(last+1, lines, nil)
		return
	}
	w.out.insert(w.after(), append([]spliceLine{blankSpliceLine}, lines...), nil)
}

// A when slot's lines: as written, or printed from the block's condition when an edit changes it.
func (w *spliceWriter) whenLines(s *spliceSlot, list *spliceList) []spliceLine {
	st := w.model.blocks[s.when]
	if !st.changed {
		if s.unit != nil {
			return w.out.written(s.unit.first, s.unit.last)
		}
		return w.out.written(w.lines.line(s.when.start), w.lines.line(s.when.end))
	}
	p := withCalls(w.printer, st.value.ProtoReflect(), st.calls, st.forced)
	ps := p.whenLine(st.value.ProtoReflect(), st.split, st.hide, groupScope(list.group), w.lineIndent(list, s.when))
	var above []string
	if s.unit != nil {
		for _, c := range s.unit.comments {
			above = append(above, w.out.lines.text(c.line))
		}
		trailing, inner := innerComments(s.when.inner, w.lines.line(s.when.end))
		ps.trailing = cmp.Or(trailing, ps.trailing)
		for _, c := range inner {
			above = append(above, commentLine(c, w.lineIndent(list, s.when)).line)
		}
	}
	return newSpliceLines(append(above, statementLines(ps)...))
}

// The column of a when line: its own in the text, or that of the body it goes in for a block an edit adds.
func (w *spliceWriter) lineIndent(list *spliceList, b *whenStmt) int {
	if b.condition != nil {
		return columnOf(b, w.lines)
	}
	return w.bodyIndent(list, whenChain(b.parent))
}

// The column of the rows in the innermost of blocks.
func (w *spliceWriter) bodyIndent(list *spliceList, blocks []*whenStmt) int {
	if len(blocks) == 0 {
		return list.indent
	}
	b := blocks[len(blocks)-1]
	if b.condition != nil && len(b.body) > 0 {
		return columnOf(b.body[0], w.lines)
	}
	return w.lineIndent(list, b) + indentWidth
}

// A def edit made on its line alone: defSet writes the def's parameters and value, and defRename renames it with
// its calls.
func (d *Doc) editDef(e AplEdit) ([]TextEdit, error) {
	f := d.typed
	def := f.typer.defs.byName[e.Name]
	if def == nil {
		return nil, errors.New("no def is named " + e.Name)
	}
	if e.Op == "defRename" {
		files, err := d.Rename(d.lines.position(def.name.start), e.Text)
		if err != nil {
			return nil, err
		}
		// The panel edits this file alone, so the calls in a file that uses it are renamed with F2.
		for _, file := range files {
			if !SamePath(file.Path, d.path) {
				return nil, fmt.Errorf("the def %s is called in %s; rename it in the text with F2, which renames those calls too", e.Name, filepath.Base(file.Path))
			}
		}
		if len(files) == 0 {
			return nil, errors.New("the def does not rename")
		}
		return files[0].Edits, nil
	}
	if strings.TrimSpace(e.Text) == "" {
		return nil, errors.New("a def needs a value")
	}
	line := defText(e.Name, e.Params, e.Text)
	back := d.readBack(d.src[:def.start]+line+d.src[def.end:], f.kind)
	if problem := firstError(back); problem != "" {
		return nil, errors.New("the def has an error: " + problem)
	}
	// The def as format writes it, with the comments inside it above it; one at the end of its last line stays there.
	indent := columnOf(def, d.lines)
	var lines []string
	_, above := innerComments(def.inner, d.lines.line(def.end))
	for _, c := range above {
		lines = append(lines, commentLine(c, indent).line)
	}
	lines = append(lines, statementLines(defLine(back.typer.defs.byName[e.Name], back.tree.lines.src, indent))...)
	eol := "\n"
	if strings.Contains(d.src, "\r\n") {
		eol = "\r\n"
	}
	return []TextEdit{{Range: d.rangeOf(def.span), NewText: strings.Join(lines, eol)[indent:]}}, nil
}

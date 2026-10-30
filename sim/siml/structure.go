package siml

import (
	"encoding/json"
	"maps"
	"slices"
	"strings"
)

// What the rotation panel shows of an x.apl.siml beyond the rotation's JSON: the when blocks its rows are in, the
// def calls its values are written as, and its defs (SPEC.md §6.9, §6.10).
type Structure struct {
	Whens []WhenBlock `json:"whens"`
	Rows  []BlockRow  `json:"rows"`
	Calls []DefCall   `json:"calls"`
	Defs  []DefInfo   `json:"defs"`
}

// A when block. Its id is its place among the file's when lines, which a ["whens", id, …] edit names.
type WhenBlock struct {
	ID        int               `json:"id"`
	List      []any             `json:"list"`   // The rotation path of the list its rows are in.
	Parent    int               `json:"parent"` // The block it is in, -1 for none.
	First     int               `json:"first"`  // The index of its first row in the list.
	Last      int               `json:"last"`   // The index of its last row, nested blocks' included; First-1 when it has none.
	Hide      bool              `json:"hide,omitempty"`
	Condition json.RawMessage   `json:"condition"`
	Conjuncts []json.RawMessage `json:"conjuncts"` // What it joins to the condition of each row in it, in order.
	Range     Range             `json:"-"`
}

// A row of a when block: how many conjuncts at the start of its condition its blocks join to it, and whether its
// condition holds conjuncts of its own after them.
type BlockRow struct {
	Path      []any `json:"path"`
	Inherited int   `json:"inherited"`
	Own       bool  `json:"own"`
}

// A def call: the rotation path of the value it reads as, or ["whens", id, "condition", …] in a block's condition.
type DefCall struct {
	Path []any    `json:"path"`
	Name string   `json:"name"`
	Args []string `json:"args"`
}

type DefInfo struct {
	Name   string   `json:"name"`
	Params []string `json:"params"`
	Value  string   `json:"value"` // As format writes it, on one line and without the comments in it.
	Used   bool     `json:"used"`
	Range  Range    `json:"-"`
}

// The structure of an x.apl.siml's rotation; nil when the text has an error, uses another file, or is no x.apl.siml.
func (d *Doc) Structure() *Structure {
	f, err := d.aplFile()
	if err != nil {
		return nil
	}
	m := newRotationModel(d.ix, f)
	out := &Structure{Whens: []WhenBlock{}, Rows: []BlockRow{}, Calls: []DefCall{}, Defs: []DefInfo{}}
	ids := map[*whenStmt]int{}
	for i, w := range m.whens {
		ids[w] = i
	}
	first, last := map[*whenStmt]int{}, map[*whenStmt]int{}
	for _, list := range m.lists() {
		at := m.listPath(list)
		k := 0
		for _, s := range list.lines {
			if s.when != nil {
				first[s.when], last[s.when] = k, k-1
				continue
			}
			for _, b := range s.blocks {
				last[b] = k
			}
			k++
		}
		for k, s := range list.slots {
			path := append(slices.Clone(at), k)
			calls := s.calls
			if len(s.blocks) > 0 {
				own, _ := m.fits(s.msg, s.blocks)
				n := len(m.conjunctsOf(s.blocks))
				out.Rows = append(out.Rows, BlockRow{Path: path, Inherited: n, Own: len(own) > 0})
				// The calls of the conjuncts its blocks join are the blocks' own.
				calls = maps.Clone(calls)
				for i := range n {
					prefix := pathKey(conjunctPath(i, n+len(own)))
					maps.DeleteFunc(calls, func(key string, _ spliceCall) bool { return under(key, prefix) })
				}
			}
			out.Calls = append(out.Calls, outermostCalls(path, calls)...)
		}
	}
	for i, w := range m.whens {
		st := m.blocks[w]
		parent := -1
		if w.parent != nil {
			parent = ids[w.parent]
		}
		b := WhenBlock{ID: i, List: m.listPath(st.list), Parent: parent, First: first[w], Last: last[w], Hide: st.hide,
			Condition: json.RawMessage(d.ix.jsonText(st.value.ProtoReflect())), Range: d.rangeOf(w.span)}
		for _, c := range st.conjuncts() {
			b.Conjuncts = append(b.Conjuncts, json.RawMessage(d.ix.jsonText(c.ProtoReflect())))
		}
		out.Whens = append(out.Whens, b)
		out.Calls = append(out.Calls, outermostCalls([]any{"whens", i, "condition"}, st.calls)...)
	}
	for _, s := range f.tree.inFileOrder() {
		def, ok := s.(*defStmt)
		if !ok || f.typer.defs.byName[def.name.text] != def {
			continue
		}
		info := DefInfo{Name: def.name.text, Params: []string{}, Used: f.typer.defs.used[def], Range: d.rangeOf(def.span)}
		for _, p := range def.params {
			info.Params = append(info.Params, p.text)
		}
		if def.value != nil {
			info.Value = nodeText(def.value, d.src)
		}
		out.Defs = append(out.Defs, info)
	}
	return out
}

// The calls of a statement no other call of it holds, at their paths below at.
func outermostCalls(at []any, calls map[string]spliceCall) []DefCall {
	var out []DefCall
	for key, c := range calls {
		inner := false
		for other := range calls {
			if other != key && under(key, other) {
				inner = true
			}
		}
		if !inner {
			out = append(out, DefCall{Path: slices.Concat(at, c.path), Name: c.name, Args: append([]string{}, c.args...)})
		}
	}
	slices.SortFunc(out, func(a, b DefCall) int { return strings.Compare(pathKey(a.Path), pathKey(b.Path)) })
	return out
}

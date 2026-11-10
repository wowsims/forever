package siml

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
)

// The editor's checks beyond the parser's: the raw refs a file under ui/specs may not hold, names the spec registers
// for no build, the rotation lints, and slot rows the item database says do not fit. Under ui/specs, where the
// repository's checks hold, warnings are errors.

// The lint codes the parser reports too (SPEC.md §7.5). Where both report one at the same text, the parser's stands.
var parserChecks = []string{"undefined-group", "undefined-variable", "undefined-sequence", "duplicate-variable"}

var lintSeverity = map[string]Severity{"error": Error, "warning": Warning, "info": Information}

// The parser's diagnostics and the checks, as the editor shows them: under ui/specs the parser's warnings are errors.
func (d *Doc) EditorDiagnostics(c *DatabaseCatalog) Diagnostics {
	strict := UnderSpecs(d.path)
	var out Diagnostics
	for _, x := range d.Diagnostics() {
		if strict && x.Severity == Warning {
			x.Severity = Error
		}
		out = append(out, x)
	}
	return append(out, d.checks(c, strict)...)
}

func (d *Doc) checks(c *DatabaseCatalog, strict bool) Diagnostics {
	if d.typed == nil {
		return nil
	}
	var out Diagnostics
	if strict && d.file != nil {
		out = append(out, d.file.RawRefs...)
	}
	check := func(at Range, code, message string) {
		severity := Warning
		if strict {
			severity = Error
		}
		out = append(out, Diagnostic{Path: d.path, Start: at.Start, End: at.End, Severity: severity, Code: code, Message: message})
	}
	d.unregistered(check)
	if c != nil {
		d.gear(c, check)
	}
	parser := d.Diagnostics()
	for _, f := range d.lints() {
		at, ok := d.SpanFor(f.Path)
		if !ok {
			continue
		}
		if slices.Contains(parserChecks, f.Code) && slices.ContainsFunc(parser, func(x Diagnostic) bool {
			return x.Code == f.Code && !before(at.End, x.Start) && !before(x.End, at.Start)
		}) {
			continue
		}
		out = append(out, Diagnostic{Path: d.path, Start: at.Start, End: at.End, Severity: lintSeverity[f.Severity], Code: f.Code, Message: f.Message})
	}
	return out
}

// Whether a lies before b.
func before(a, b Position) bool { return a.Line < b.Line || a.Line == b.Line && a.Col < b.Col }

// The rotation lints, when the text reads and holds a rotation.
func (d *Doc) lints() []Finding {
	if d.err != nil || d.file == nil || !slices.Contains(d.file.Categories, rotationCategory) {
		return nil
	}
	return Lint(d.file.Rotation())
}

// A spell-like name the spec registers for no build, as the repository's registration check finds it: the sim does not
// have it unless gear, an item or a setting the builds leave out registers it.
func (d *Doc) unregistered(check func(Range, string, string)) {
	spec := d.specName()
	for _, r := range d.Refs() {
		if !r.Namespace.SpellLike() {
			continue
		}
		entry, ok := d.names.spellEntry(d.class(), r.Namespace, r.Name)
		if !ok {
			continue
		}
		registered := len(entry.RegisteredBy) > 0
		builds := ""
		if spec != "" {
			registered, builds = slices.Contains(entry.RegisteredBy, spec), spec+" "
		}
		if !registered {
			check(Range{r.Start, r.End}, "unregistered-name", fmt.Sprintf("%s is %s no %sbuild registers", r.Name, namespaceThings[r.Namespace], builds))
		}
	}
}

// Slot rows whose item, suffix or enchant the database says do not go together: the item goes in another slot, the
// item rolls no such suffix, or no row of the enchant fits the slot or the item.
func (d *Doc) gear(c *DatabaseCatalog, check func(Range, string, string)) {
	var rows []*slotRowStmt
	for _, s := range d.tree().statements {
		switch x := s.(type) {
		case *slotRowStmt:
			rows = append(rows, x)
		case *headerStmt:
			if row, ok := x.body.(*slotRowStmt); ok {
				rows = append(rows, row)
			}
		}
	}
	for _, row := range rows {
		if row.item == nil {
			continue
		}
		slot := row.slot.text
		slotValue := d.ix.enums[slotEnum].byName[slot]
		id, named := d.names.Item(row.item.text)
		item, found := CatalogItem{}, false
		if named {
			item, found = c.Item(id)
		}
		switch {
		case named && !found:
			if known, ok := c.ItemName(id); ok {
				check(d.rangeOf(row.item.span), "gear-item", known+" is no gear the database lists")
			} else {
				check(d.rangeOf(row.item.span), "gear-item", fmt.Sprintf("the database has no item %d", id))
			}
		case found && slotValue != nil && !slices.Contains(item.slots(), proto.ItemSlot(slotValue.Number)):
			slots := strings.Join(d.slotWords(item.slots()), " or ")
			if slots == "" {
				slots = "no slot"
			}
			check(d.rangeOf(row.item.span), "gear-slot", fmt.Sprintf("%s goes in %s, not %s", item.Name, slots, slot))
		}
		if row.suffix != nil && found {
			if suffix, ok := d.names.Suffix(row.suffix.text); ok && !slices.Contains(c.Suffixes(item.ID), suffix) {
				check(d.rangeOf(row.suffix.span), "gear-suffix", fmt.Sprintf("%s rolls no %s", item.Name, row.suffix.text))
			}
		}
		if row.enchant == nil {
			continue
		}
		entry, ok := d.names.Enchants[row.enchant.text]
		switch {
		case !ok:
		case slotValue == nil || !slices.Contains(entry.Slots, proto.ItemSlot(slotValue.Number)):
			check(d.rangeOf(row.enchant.span), "gear-enchant", fmt.Sprintf("%s fits %s, not %s", row.enchant.text, strings.Join(d.slotWords(entry.Slots), ", "), slot))
		case found && !enchantEffectFits(c, entry.EffectID, item):
			check(d.rangeOf(row.enchant.span), "gear-enchant", fmt.Sprintf("%s does not fit %s", row.enchant.text, item.Name))
		}
	}
}

// --- Code actions.

// The end of a parser message that names the closest candidates to a word it does not know.
var didYouMeanCandidates = regexp.MustCompile(`; did you mean ([^?]+)\?`)

// The code actions at a range. A word the parser does not know, where a diagnostic of the parser's touches the range,
// changes to one of the candidates the diagnostic names. The rewrites of a rotation's conditions each change the proto,
// so none is made without asking (SPEC.md §16.2): an `and` or `or` of one operand is that operand, and an `and` inside
// an `and`, or an `or` inside an `or`, is one list.
func (d *Doc) quickFixes(r Range) []Fix {
	tree := d.tree()
	if tree == nil {
		return nil
	}
	from, to := d.Offset(r.Start), d.Offset(r.End)
	var out []Fix
	for _, x := range d.Diagnostics() {
		if before(x.End, r.Start) || before(r.End, x.Start) {
			continue
		}
		m := didYouMeanCandidates.FindStringSubmatch(x.Message)
		if m == nil || !identPattern.MatchString(d.src[d.Offset(x.Start):d.Offset(x.End)]) {
			continue
		}
		for i, candidate := range strings.Split(m[1], ", ") {
			out = append(out, Fix{
				Title: "Change to " + candidate, Kind: "quickfix", Diagnostic: x, Preferred: i == 0,
				Edits: []TextEdit{{Range: Range{x.Start, x.End}, NewText: candidate}},
			})
		}
	}
	text := func(n node) string { s := n.pos(); return d.src[s.start:s.end] }
	rewrite := func(title string, n node, newText string) {
		out = append(out, Fix{Title: title, Kind: "refactor.rewrite", Edits: []TextEdit{{Range: d.rangeOf(n.pos()), NewText: newText}}})
	}
	for _, s := range tree.statements {
		walkTree(s, nil, func(n, parent any) {
			at, ok := n.(node)
			if !ok || at.pos().end < from || at.pos().start > to {
				return
			}
			switch x := n.(type) {
			case *callNode:
				if (x.name.text != "and" && x.name.text != "or") || x.args == nil {
					return
				}
				operands := x.args.items
				// A named operand, or a list, which and/or reads as its operands, writes no operator.
				if slices.ContainsFunc(operands, func(a arg) bool { _, list := a.value.(*listNode); return a.name != nil || list }) {
					return
				}
				op := x.name.text
				switch {
				case len(operands) == 1:
					operand := operands[0].value
					newText := text(operand)
					if parenthesized(nodeType(operand), nodeType(parent)) {
						newText = "(" + newText + ")"
					}
					rewrite(fmt.Sprintf("Write %s(…) of one operand as the operand", op), x, newText)
				case len(operands) > 1 && nodeType(parent) == op:
					parts := make([]string, len(operands))
					for i, a := range operands {
						parts[i] = text(a.value)
						if nodeType(a.value) == "or" {
							parts[i] = "(" + parts[i] + ")"
						}
					}
					rewrite(fmt.Sprintf("Join this %s(…) into the %[1]s around it", op), x, strings.Join(parts, " "+op+" "))
				}
			case *parenNode:
				if inner := nodeType(x.expr); (inner == "and" || inner == "or") && nodeType(parent) == inner {
					rewrite(fmt.Sprintf("Join this %s into the %[1]s around it", inner), x, text(x.expr))
				}
			}
		})
	}
	return out
}

// What kind of expression a node is, as the rewrites compare them: and, or, not, compare, math, call, paren; "" for
// any other.
func nodeType(n any) string {
	switch x := n.(type) {
	case *logicNode:
		return x.op
	case *notNode:
		return "not"
	case *compareNode:
		return "compare"
	case *mathNode:
		return "math"
	case *callNode:
		return "call"
	case *parenNode:
		return "paren"
	}
	return ""
}

// Whether an expression of type inner needs parentheses to stand where a call stood inside one of type outer (SPEC.md
// §6.5): to keep an `and` in an `and` (or an `or` in an `or`) apart, and where it binds looser.
func parenthesized(inner, outer string) bool {
	looseness := map[string]int{"or": 4, "and": 3, "not": 2, "compare": 1, "math": 0}
	li, innerOK := looseness[inner]
	lo, outerOK := looseness[outer]
	if !innerOK || !outerOK {
		return false
	}
	if outer == "math" || inner == outer {
		return true
	}
	return li > lo || outer == "compare" && inner == "compare"
}

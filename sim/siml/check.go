package siml

import (
	"fmt"
	"slices"
	"strings"
)

// The checks that need the whole file: the references between variables, groups, placeholders and sequences
// (SPEC.md §7.5).

// What a variable reference refers to among the names declared in its scope: one name (found), several
// (ambiguous), or none.
type resolution struct {
	name      string
	found     bool
	ambiguous []string
}

// What `$written` (or `${written}` when braced) refers to among declared: the exact name, else the one name
// whose identifier is written (SPEC.md §7.5).
func resolveVariable(written string, braced bool, declared []string) resolution {
	if slices.Contains(declared, written) {
		return resolution{name: written, found: true}
	}
	if braced {
		return resolution{}
	}
	var folded []string
	for _, d := range declared {
		if Identifier(d) == written && !slices.Contains(folded, d) {
			folded = append(folded, d)
		}
	}
	if len(folded) == 1 {
		return resolution{name: folded[0], found: true}
	}
	return resolution{ambiguous: folded}
}

// A variable a rotation declares: a value variable, or a variable of the group scope names.
type declared struct {
	name  string
	scope scope
}

// The names a reference in sc sees: the value variables, and inside a group that group's variables.
func declaredIn(declarations []declared, sc scope) []string {
	var names []string
	for _, d := range declarations {
		if (!d.scope.inGroup || sc.inGroup && d.scope == sc) && !slices.Contains(names, d.name) {
			names = append(names, d.name)
		}
	}
	return names
}

func declaredNames(declarations []*variableDecl) []declared {
	out := make([]declared, len(declarations))
	for i, d := range declarations {
		out[i] = declared{name: d.name, scope: d.scope}
	}
	return out
}

// The references between variables, groups, placeholders and sequences (SPEC.md §7.5). They may name what the
// rotations of used files declare.
func (t *typer) checkReferences(declarations []*variableDecl, groups []nameUse, used usedDeclarations) {
	seen := map[scope]map[string]bool{}
	for _, d := range declarations {
		if seen[d.scope] == nil {
			seen[d.scope] = map[string]bool{}
		}
		if seen[d.scope][d.name] {
			in := ""
			if d.scope.inGroup {
				in = fmt.Sprintf(` in group "%s"`, d.scope.group)
			}
			t.report("duplicate-variable", d.name+" is declared twice"+in, d.span)
		}
		seen[d.scope][d.name] = true
		if from, ok := used.from[d.name]; ok && !d.scope.inGroup {
			t.report("duplicate-variable", fmt.Sprintf("%s is declared in both this file and %s; set it with use … with", d.name, from), d.span)
		}
	}

	visible := append(declaredNames(declarations), used.variables...)
	targets := map[*variableUse]*variableDecl{}
	for _, use := range t.variableUses {
		t.site = use.site
		n := use.node
		// The sim finds a reference among the value variables only; a group's own variables and a reference's
		// `with` replace one of the same name and declare none (SPEC.md §7.5).
		res := resolveVariable(n.name, n.braced, declaredIn(visible, scope{}))
		use.name(n.name)
		switch {
		case res.found:
			use.name(res.name)
			targets[use] = declarationOf(declarations, res.name, use.scope)
		case t.reportAmbiguous(n, res):
		default:
			t.warn("undefined-variable", "no variable here is named "+nodeText(n, ""), n.span)
		}
	}
	t.site = nil
	uses := func(d *variableDecl) []*variableDecl {
		var out []*variableDecl
		for _, u := range d.uses {
			if target := targets[u]; target != nil {
				out = append(out, target)
			}
		}
		return out
	}
	for _, d := range declarations {
		if reaches(d, d, uses) {
			t.report("reference-cycle", d.name+" refers to itself through other variables", d.span)
		}
	}

	defined := map[string]bool{}
	for _, g := range groups {
		defined[g.name] = true
	}
	for g := range used.groups {
		defined[g] = true
	}
	for _, use := range t.groupUses {
		if !defined[use.name] {
			t.warn("undefined-group", fmt.Sprintf(`no group is named "%s"`, use.name), use.span)
			continue
		}
		placeholders := t.placeholders[scope{group: use.name, inGroup: true}]
		for _, w := range use.with {
			if !placeholders[w.name] && !used.groups[use.name][w.name] {
				t.warn("undefined-placeholder", fmt.Sprintf(`group "%s" holds no %%%s`, use.name, w.name), w.span)
			}
		}
	}
	referenced := func(name string) []string {
		var out []string
		for _, u := range t.groupUses {
			if u.scope == (scope{group: name, inGroup: true}) && defined[u.name] && !slices.Contains(out, u.name) {
				out = append(out, u.name)
			}
		}
		return out
	}
	reported := map[string]bool{}
	for _, g := range groups {
		if reported[g.name] || !reaches(g.name, g.name, referenced) {
			continue
		}
		reported[g.name] = true
		t.report("reference-cycle", fmt.Sprintf(`group "%s" refers to itself through group references`, g.name), g.span)
	}

	for _, use := range t.sequenceUses {
		if !t.sequenceDefs[use.name] {
			t.site = use.site
			t.warn("undefined-sequence", fmt.Sprintf(`no sequence is named "%s"`, use.name), use.span)
			t.site = nil
		}
	}
}

// The declaration a resolved name refers to from sc: the group's own, else the value variable.
func declarationOf(declarations []*variableDecl, name string, sc scope) *variableDecl {
	for _, d := range declarations {
		if d.name == name && sc.inGroup && d.scope == sc {
			return d
		}
	}
	for _, d := range declarations {
		if d.name == name && !d.scope.inGroup {
			return d
		}
	}
	return nil
}

// Whether to is reachable from `from` in one step or more.
func reaches[T comparable](from, to T, next func(T) []T) bool {
	seen := map[T]bool{}
	stack := next(from)
	for len(stack) > 0 {
		x := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if x == to {
			return true
		}
		if seen[x] {
			continue
		}
		seen[x] = true
		stack = append(stack, next(x)...)
	}
	return false
}

// Reports ambiguous-variable at v when res holds several names; false when it holds none.
func (t *typer) reportAmbiguous(v *varNode, res resolution) bool {
	if len(res.ambiguous) == 0 {
		return false
	}
	exact := make([]string, len(res.ambiguous))
	for i, a := range res.ambiguous {
		exact[i] = braced("$", a)
	}
	t.report("ambiguous-variable", fmt.Sprintf("%s could be any of %s; write the exact name", nodeText(v, ""), strings.Join(exact, ", ")), v.span)
	return true
}

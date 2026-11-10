package reforgeoptimizer

import (
	"maps"
	"math"
	"strconv"
	"strings"
)

// This file produces a stable, exact CPLEX LP text encoding of the model for the HiGHS solver.
// The serialization is fully deterministic down to the byte level (including the quirk that the
// leading space is dropped on the first wrapped line), so an unchanged model always yields the
// same LP text and therefore the same HiGHS tie-breaking among equal-objective solutions.

// The longest LP line written before the expression wraps.
const maxLPLineLength = 200

// An optional bound row: a lower bound (min), an upper bound (max), or both. lessEq builds a "<="
// row (max) and greaterEq builds a ">=" row (min).
type lpConstraint struct {
	min, max       float64
	hasMin, hasMax bool
}

func lessEq(n float64) lpConstraint    { return lpConstraint{max: n, hasMax: true} }
func greaterEq(n float64) lpConstraint { return lpConstraint{min: n, hasMin: true} }

// An insertion-ordered map name -> coefficients. Order is load-bearing: it determines x-index
// assignment and therefore HiGHS's tie-breaking among equal-EP solutions.
//
// Each variable carries coefficients in TWO parallel spaces:
//   - byName holds the CAP coefficients (fully stat-dependency-resolved via resolveStatDelta)
//     plus the structural/constraint keys. modelToLPFormat's constraint
//     rows and checkCaps read byName, so every dependency counts toward caps.
//   - objByName holds the OBJECTIVE coefficients (the EP-calibrated applyReforgeStat output).
//     updateReforgeScores reads objByName to compute each variable's 'score' coefficient, so
//     the LP objective stays exactly as calibrated.
type lpVariables struct {
	order     []string
	byName    map[string]map[string]float64
	objByName map[string]map[string]float64
}

func newLPVariables() *lpVariables {
	return &lpVariables{
		byName:    map[string]map[string]float64{},
		objByName: map[string]map[string]float64{},
	}
}

// Stores a variable's CAP coefficients (byName). A new name goes to the end of the insertion
// order; overwriting an existing one keeps its original position.
func (v *lpVariables) set(name string, coeffs map[string]float64) {
	if _, ok := v.byName[name]; !ok {
		v.order = append(v.order, name)
	}
	v.byName[name] = coeffs
}

// Stores a variable's OBJECTIVE coefficients (objByName). The insertion order is left alone; only
// the CAP coefficients' setter adds to it.
func (v *lpVariables) setObj(name string, coeffs map[string]float64) {
	v.objByName[name] = coeffs
}

func (v *lpVariables) get(name string) (map[string]float64, bool) {
	c, ok := v.byName[name]
	return c, ok
}

func (v *lpVariables) getObj(name string) map[string]float64 {
	return v.objByName[name]
}

func (v *lpVariables) each(fn func(name string, coeffs map[string]float64)) {
	for _, name := range v.order {
		fn(name, v.byName[name])
	}
}

func (v *lpVariables) len() int { return len(v.order) }

// An insertion-ordered map name -> constraint.
type lpConstraints struct {
	order  []string
	byName map[string]lpConstraint
}

func newLPConstraints() *lpConstraints {
	return &lpConstraints{byName: map[string]lpConstraint{}}
}

func (c *lpConstraints) set(name string, con lpConstraint) {
	if _, ok := c.byName[name]; !ok {
		c.order = append(c.order, name)
	}
	c.byName[name] = con
}

func (c *lpConstraints) has(name string) bool {
	_, ok := c.byName[name]
	return ok
}

// A shallow copy that can be extended without mutating the original; checkCaps tightens it into
// the constraint set for the next refinement pass.
func (c *lpConstraints) clone() *lpConstraints {
	out := &lpConstraints{
		order:  append([]string(nil), c.order...),
		byName: make(map[string]lpConstraint, len(c.byName)),
	}
	maps.Copy(out.byName, c.byName)
	return out
}

func (c *lpConstraints) each(fn func(name string, con lpConstraint)) {
	for _, name := range c.order {
		fn(name, c.byName[name])
	}
}

type lpModel struct {
	direction   string // "maximize" | "minimize"
	objective   string
	variables   *lpVariables
	constraints *lpConstraints
	binaries    bool
}

// The parts of a solved LP the algorithm consumes: the status, the objective value, and the
// ORIGINAL names of the variables the solver selected (column primal >= 0.5), in x-index
// (variable insertion) order.
type lpSolution struct {
	status    string // "optimal" | "infeasible" | "unbounded" | "timedout" | "unknown"
	result    float64
	variables []string
}

// Serializes the model to CPLEX LP text for the HiGHS solver. Returns the LP text and the
// original variable names, where entry i is the name mapped to "x{i}".
func modelToLPFormat(model *lpModel) (string, []string) {
	varNameMap := make(map[string]string, model.variables.len())
	reverseNames := make([]string, 0, model.variables.len())
	idx := 0
	model.variables.each(func(name string, _ map[string]float64) {
		escaped := "x" + strconv.Itoa(idx)
		varNameMap[name] = escaped
		reverseNames = append(reverseNames, name)
		idx++
	})

	lines := make([]string, 0, 8+idx)
	if model.direction == "maximize" {
		lines = append(lines, "Maximize")
	} else {
		lines = append(lines, "Minimize")
	}
	lines = append(lines, buildLPObjective(model, varNameMap)...)

	constraintLines := buildLPConstraints(model, varNameMap)
	if len(constraintLines) > 0 {
		lines = append(lines, "Subject To")
		lines = append(lines, constraintLines...)
	}

	// buildBounds emits nothing for binary variables; the "Bounds" header is only written
	// when there are bound lines.
	if !model.binaries && idx > 0 {
		lines = append(lines, "Bounds")
		for i := 0; i < idx; i++ {
			lines = append(lines, " 0 <= x"+strconv.Itoa(i))
		}
	}

	if model.binaries {
		lines = append(lines, "Binary")
		for i := 0; i < idx; i++ {
			lines = append(lines, " x"+strconv.Itoa(i))
		}
	}

	lines = append(lines, "End")
	return strings.Join(lines, "\n"), reverseNames
}

// Appends one "coeff var" term with LP sign formatting: the first term is bare, later terms carry
// an explicit "+ " / "- " with the absolute coefficient.
func appendLPTerm(terms []string, coeff float64, escaped string) []string {
	if len(terms) == 0 {
		return append(terms, formatLPNumber(coeff)+" "+escaped)
	}
	if coeff >= 0 {
		return append(terms, "+ "+formatLPNumber(coeff)+" "+escaped)
	}
	return append(terms, "- "+formatLPNumber(math.Abs(coeff))+" "+escaped)
}

func buildLPObjective(model *lpModel, varNameMap map[string]string) []string {
	var terms []string
	model.variables.each(func(name string, coeffs map[string]float64) {
		score, ok := coeffs[model.objective]
		if !ok || score == 0 {
			return
		}
		terms = appendLPTerm(terms, score, varNameMap[name])
	})

	if len(terms) == 0 {
		return []string{" obj: 0"}
	}
	return wrapLPExpression(" obj: "+strings.Join(terms, " "), maxLPLineLength)
}

func buildLPConstraints(model *lpModel, varNameMap map[string]string) []string {
	var lines []string
	constraintIndex := 0

	nextLabel := func() string {
		label := "c" + strconv.Itoa(constraintIndex)
		constraintIndex++
		return label
	}

	model.constraints.each(func(cname string, c lpConstraint) {
		if !c.hasMin && !c.hasMax {
			return
		}

		var terms []string
		model.variables.each(func(vname string, coeffs map[string]float64) {
			coeff, ok := coeffs[cname]
			if !ok || coeff == 0 {
				return
			}
			terms = appendLPTerm(terms, coeff, varNameMap[vname])
		})

		if len(terms) == 0 {
			return
		}
		lhs := strings.Join(terms, " ")

		if c.hasMax {
			lines = append(lines, wrapLPExpression(" "+nextLabel()+": "+lhs+" <= "+formatLPNumber(c.max), maxLPLineLength)...)
		}
		if c.hasMin {
			lines = append(lines, wrapLPExpression(" "+nextLabel()+": "+lhs+" >= "+formatLPNumber(c.min), maxLPLineLength)...)
		}
	})

	return lines
}

// Splits a long LP expression across multiple lines at token boundaries so no line exceeds
// maxLength. The first wrapped line drops its leading space (an empty currentLine
// has length 0, so the next token replaces it without the leading space).
func wrapLPExpression(expression string, maxLength int) []string {
	if len(expression) <= maxLength {
		return []string{expression}
	}

	var lines []string
	var currentLine strings.Builder
	currentLine.Grow(maxLength)
	for token := range strings.SplitSeq(expression, " ") {
		if currentLine.Len() == 0 {
			currentLine.WriteString(token)
		} else if currentLine.Len()+1+len(token) <= maxLength {
			currentLine.WriteByte(' ')
			currentLine.WriteString(token)
		} else {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLine.Grow(maxLength)
			currentLine.WriteByte(' ')
			currentLine.WriteString(token)
		}
	}
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}
	return lines
}

// Renders a coefficient as LP text: fixed 10 decimal places, then strip trailing
// zeros and an optional trailing dot; ±Inf -> ±1e30; NaN -> 0.
func formatLPNumber(num float64) string {
	if math.IsInf(num, 1) {
		return "1e30"
	}
	if math.IsInf(num, -1) {
		return "-1e30"
	}
	if math.IsNaN(num) {
		return "0"
	}
	s := strconv.FormatFloat(num, 'f', 10, 64)
	if strings.ContainsRune(s, '.') {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}

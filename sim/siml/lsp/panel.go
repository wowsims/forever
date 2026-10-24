package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/wowsims/forever/sim/siml"
)

// The requests the rotation panel of the VS Code extension makes about an open .apl.siml: the rotation the text reads
// as, the rotation path at a position, the text edits that make the panel's edits, the database's names and icons,
// the patterns Insert Pattern offers, and a sim of the text.

func (s *server) panelRequest(ctx context.Context, method string, params json.RawMessage) (any, *ResponseError) {
	var p struct {
		URI          string         `json:"uri"`
		Version      *int           `json:"version"`
		Edits        []siml.AplEdit `json:"edits"`
		ID           string         `json:"id"`
		Values       map[string]any `json:"values"`
		Path         []any          `json:"path"`
		Compare      bool           `json:"compare"`
		SettingsFile string         `json:"settingsFile"`
		Iterations   int            `json:"iterations"`
		Seed         int64          `json:"seed"`
		Timeline     bool           `json:"timeline"`
		Position     Position       `json:"position"`
	}
	if len(params) > 0 {
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, invalidParams(err)
		}
	}
	p.Path = jsonPath(p.Path)
	switch method {
	case "siml/rotation":
		return s.rotation(p.URI)
	case "siml/splice":
		return s.splice(p.URI, p.Version, p.Edits)
	case "siml/registered":
		return s.registered(p.URI)
	case "siml/pathAt":
		return s.pathAt(p.URI, p.Position)
	case "siml/database":
		return s.panelDatabase(p.URI)
	case "siml/patterns":
		return patternList(), nil
	case "siml/insertPattern":
		return s.insertPattern(p.URI, p.Version, p.ID, p.Values, p.Path)
	case "siml/sim":
		return s.sim(ctx, simParams{uri: p.URI, compare: p.Compare, settingsFile: p.SettingsFile, iterations: p.Iterations, seed: p.Seed, timeline: p.Timeline})
	}
	return nil, &ResponseError{Code: codeMethodNotFound, Message: "method not found: " + method}
}

// A JSON path as siml reads one: list indices as int.
func jsonPath(path []any) []any {
	for i, part := range path {
		if n, ok := part.(float64); ok {
			path[i] = int(n)
		}
	}
	return path
}

// A finding of the panel's rows: a diagnostic of the file, at the rotation path of the text it is about.
type panelFinding struct {
	Path     []any  `json:"path"`
	Severity string `json:"severity"` // error, warning or info.
	Code     string `json:"code"`
	Message  string `json:"message"`
}

type panelSpan struct {
	Path  []any `json:"path"`
	Range Range `json:"range"`
}

// An ActionID as JSON: a spell's id, tag and rank, or an item's id.
type actionID struct {
	SpellID int32 `json:"spellId,omitempty"`
	ItemID  int32 `json:"itemId,omitempty"`
	Tag     int32 `json:"tag,omitempty"`
	Rank    int32 `json:"rank,omitempty"`
}

type registeredName struct {
	Name string   `json:"name"`
	ID   actionID `json:"id"`
}

// What the panel's pickers offer: the names a spec's builds register by the namespace each position reads (SPEC.md
// §7.1), and the items with a use. Status is ready, or unavailable with the reason.
type registered struct {
	Status     string           `json:"status"`
	Spec       string           `json:"spec,omitempty"` // The field of Player's spec oneof.
	Reason     string           `json:"reason,omitempty"`
	Registered *registeredNames `json:"registered,omitempty"`
}

type registeredNames struct {
	Spells       []registeredName `json:"spells"`
	Auras        []registeredName `json:"auras"`
	TargetAuras  []registeredName `json:"targetAuras"`
	PetAuras     []registeredName `json:"petAuras"`
	TargetSpells []registeredName `json:"targetSpells"`
	Items        []registeredName `json:"items"`
}

// The rotation an .apl.siml reads as at a document version, for the panel.
type rotationResult struct {
	Version   int             `json:"version"`
	Rotation  json.RawMessage `json:"rotation,omitempty"`  // Left out while the text has errors.
	ReadOnly  string          `json:"readOnly,omitempty"`  // Why the panel cannot edit it.
	ClassName string          `json:"className,omitempty"` // The class the names are read with: hunter.
	Findings  []panelFinding  `json:"findings"`
	Spans     []panelSpan     `json:"spans"`
	Used      []string        `json:"used"`                // The files the use lines read, and the files those use.
	Spec      string          `json:"spec,omitempty"`      // The field of Player's spec oneof; siml/registered answers for it.
	Names     string          `json:"names"`               // Which name table the file is read with; siml/registered answers again when it changes.
	Database  string          `json:"database"`            // Which database read the answer holds; siml/database answers again when it changes.
	Structure *panelStructure `json:"structure,omitempty"` // Left out while the text has errors.
}

// The when blocks, def calls and defs of the rotation, with the text of each block's line and each def.
type panelStructure struct {
	*siml.Structure
	WhenRanges []Range `json:"whenRanges"` // By block id.
	DefRanges  []Range `json:"defRanges"`  // In the order of the defs.
}

func newPanelStructure(s *siml.Structure) *panelStructure {
	if s == nil {
		return nil
	}
	out := &panelStructure{Structure: s, WhenRanges: []Range{}, DefRanges: []Range{}}
	for _, w := range s.Whens {
		out.WhenRanges = append(out.WhenRanges, toRange(w.Range))
	}
	for _, d := range s.Defs {
		out.DefRanges = append(out.DefRanges, toRange(d.Range))
	}
	return out
}

var severityNames = map[int]string{severityError: "error", severityWarning: "warning", severityInformation: "info"}

func (s *server) rotation(uri string) (any, *ResponseError) {
	catalog := s.readCatalog(uri)
	s.mu.Lock()
	defer s.mu.Unlock()
	a := s.analyzeLocked(uri)
	if a == nil {
		return nil, failed("%s is not open", uri)
	}
	out := rotationResult{Version: a.version, Findings: []panelFinding{}, Spans: []panelSpan{}, Used: []string{}}
	out.Names, out.Database = fmt.Sprintf("%p", a.names), fmt.Sprintf("%p", catalog.of(a))
	if a.err != nil {
		out.ReadOnly = a.err.Error()
		return out, nil
	}
	d := a.doc
	out.Rotation = d.RotationJSON()
	out.ReadOnly = readOnlyReason(d, a.used)
	out.Structure = newPanelStructure(d.Structure())
	out.Used = append(out.Used, a.used...)
	f, _ := d.File()
	var info siml.SpecInfo
	var hasSpec bool
	if f != nil {
		info, hasSpec = siml.SpecOf(f.Spec)
	}
	if hasSpec {
		out.ClassName = siml.ClassName(info.Class)
		out.Spec = string(info.Field.Name())
	}
	if d.Kind().Section == "apl" {
		for _, span := range d.Spans() {
			out.Spans = append(out.Spans, panelSpan{Path: span.Path, Range: toRange(span.Range)})
		}
	}
	for _, diag := range s.diagnosticsLocked(a, catalog.of(a)) {
		path := d.PathAt(fromPosition(diag.Range.Start))
		out.Findings = append(out.Findings, panelFinding{Path: nonNil(path), Severity: severityNames[diag.Severity], Code: diag.Code, Message: diag.Message})
	}
	return out, nil
}

// The rotation path the text at a position of an open file writes, which the panel selects with the cursor and
// Insert Pattern inserts at; [] outside the rotation's spans and while the file cannot be read.
func (s *server) pathAt(uri string, at Position) (any, *ResponseError) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a := s.analyzeLocked(uri)
	if a == nil {
		return nil, failed("%s is not open", uri)
	}
	if a.err != nil {
		return []any{}, nil
	}
	return a.doc.PathAt(fromPosition(at)), nil
}

// The names the panel's pickers offer for an open file's spec. They depend on the spec line and the name table only,
// so the panel asks again when the spec or rotationResult.Names changes, not on each edit.
func (s *server) registered(uri string) (any, *ResponseError) {
	catalog := s.readCatalog(uri)
	s.mu.Lock()
	defer s.mu.Unlock()
	a := s.analyzeLocked(uri)
	if a == nil {
		return nil, failed("%s is not open", uri)
	}
	if a.err != nil {
		return registered{Status: "unavailable", Reason: a.err.Error()}, nil
	}
	if f, _ := a.doc.File(); f != nil {
		if info, ok := siml.SpecOf(f.Spec); ok {
			return s.registeredFor(a, info, catalog.of(a)), nil
		}
	}
	return registered{Status: "unavailable", Reason: "the file has no spec line"}, nil
}

// Why the panel cannot edit the rotation: the text has an error, or the rotation is put together with another file,
// whose rows the panel shows but this text does not hold. "" when it can.
func readOnlyReason(d *siml.Doc, used []string) string {
	for _, diag := range d.Diagnostics() {
		if diag.Severity == siml.Error {
			return fmt.Sprintf("The text has errors (line %d: %s). Edits wait until it reads again.", diag.Start.Line, diag.Message)
		}
	}
	if d.Kind().Section != "apl" {
		return "The panel edits the rotation of an x.apl.siml file."
	}
	if len(used) > 0 {
		if _, err := d.Splice(nil); err != nil {
			return sentence(err.Error())
		}
	}
	return ""
}

// A message as a sentence: its first letter upper case and a full stop at its end.
func sentence(message string) string {
	r, size := utf8.DecodeRuneInString(message)
	message = string(unicode.ToUpper(r)) + message[size:]
	if !strings.HasSuffix(message, ".") {
		message += "."
	}
	return message
}

func (s *server) registeredFor(a *analysis, info siml.SpecInfo, c *siml.DatabaseCatalog) registered {
	spec := string(info.Field.Name())
	list := func(ns siml.Namespace) []registeredName {
		out := []registeredName{}
		for _, n := range a.names.Registered(info.Class, spec, ns) {
			out = append(out, registeredName{Name: n.Name, ID: actionID{SpellID: n.Entry.SpellID, Tag: n.Entry.Tag, Rank: n.Entry.Rank}})
		}
		return out
	}
	out := registered{Status: "ready", Spec: spec, Registered: &registeredNames{
		Spells:       list(siml.NamespaceSpell),
		Auras:        list(siml.NamespaceAura),
		TargetAuras:  list(siml.NamespaceTargetAura),
		PetAuras:     list(siml.NamespacePetAura),
		TargetSpells: list(siml.NamespaceTargetSpell),
		Items:        []registeredName{},
	}}
	for _, id := range usableItems(c) {
		if name, ok := a.names.ItemName(id); ok {
			out.Registered.Items = append(out.Registered.Items, registeredName{Name: "item." + name, ID: actionID{ItemID: id}})
		}
	}
	return out
}

// The document's analysis at version, or the error a request at another version answers with.
func (s *server) atVersionLocked(uri string, version *int) (*analysis, *ResponseError) {
	a := s.analyzeLocked(uri)
	if a == nil {
		return nil, failed("%s is not open", uri)
	}
	if version != nil && *version != a.version {
		return nil, &ResponseError{Code: codeContentModified, Message: fmt.Sprintf("the document is at version %d, not %d", a.version, *version)}
	}
	if a.err != nil {
		return nil, failed("%s", a.err.Error())
	}
	return a, nil
}

// The text edits that make the panel's edits on the document at version.
func (s *server) splice(uri string, version *int, edits []siml.AplEdit) (any, *ResponseError) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, rpcErr := s.atVersionLocked(uri, version)
	if rpcErr != nil {
		return nil, rpcErr
	}
	out, err := a.doc.Splice(edits)
	if err != nil {
		return nil, failed("%s", err.Error())
	}
	return map[string]any{"edits": toTextEdits(out)}, nil
}

func (s *server) panelDatabase(uri string) (any, *ResponseError) {
	s.mu.Lock()
	var path string
	if d := s.docs[uri]; d != nil {
		path = d.path
	} else if p, ok := uriPath(uri); ok {
		path = p
	}
	root, ok := s.rootLocked(path)
	s.mu.Unlock()
	if !ok {
		return panelDatabase{Spells: map[int32][2]string{}, Items: map[int32][2]string{}}, nil
	}
	c := siml.CatalogAt(root)
	return panelDatabase{Spells: c.SpellIcons(), Items: c.ItemIcons()}, nil
}

// The edits that insert a pattern at a rotation path, with the values the reader gave.
func (s *server) insertPattern(uri string, version *int, id string, values map[string]any, path []any) (any, *ResponseError) {
	pattern, ok := patternByID(id)
	if !ok {
		return nil, invalidParams(fmt.Errorf("no pattern is named %s", id))
	}
	place := placementAt(path, pattern.Target)
	if pattern.Target == "condition" && place.condition == nil {
		return nil, failed("%s is a condition: put the cursor on a priority or group row to add it there.", pattern.Label)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	a, rpcErr := s.atVersionLocked(uri, version)
	if rpcErr != nil {
		if rpcErr.Code == codeContentModified {
			rpcErr.Message = "The rotation changed while the pattern was being filled in; insert it again."
		}
		return nil, rpcErr
	}
	if why := readOnlyReason(a.doc, a.used); why != "" {
		return nil, failed("Cannot insert a pattern: %s", why)
	}
	var rotation any
	if raw := a.doc.RotationJSON(); raw != nil {
		if err := json.Unmarshal(raw, &rotation); err != nil {
			return nil, failed("%s", err.Error())
		}
	}
	edits, err := patternEdits(rotation, pattern, values, place)
	if err != nil {
		return nil, failed("The pattern was not inserted: %s.", err.Error())
	}
	out, err := a.doc.Splice(edits)
	if err != nil {
		return nil, failed("The pattern was not inserted: %s.", err.Error())
	}
	return map[string]any{"edits": toTextEdits(out)}, nil
}

var errNoSim = errors.New("this server sims nothing")

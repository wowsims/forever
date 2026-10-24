package lsp

import (
	"github.com/wowsims/forever/sim/siml"
)

// Definition, references and rename of what a rotation names itself, which the analysis follows into the files a
// `use … with` list sets variables of and the .siml files near them; here their files become URIs.

// The URI of a file a location is in: the requesting document's for its own path or none, an open document's, else
// the file's.
func (s *server) locationURILocked(a *analysis, uri, path string) string {
	if path == "" || siml.SamePath(path, a.doc.Path()) {
		return uri
	}
	return s.uriOfLocked(path)
}

func (s *server) locationsLocked(a *analysis, uri string, found []siml.Location) []Location {
	out := []Location{}
	for _, l := range found {
		out = append(out, Location{URI: s.locationURILocked(a, uri, l.Path), Range: toRange(l.Range)})
	}
	return out
}

func (s *server) prepareRenameLocked(a *analysis, at Position) (any, *ResponseError) {
	r, placeholder, err := a.doc.PrepareRename(fromPosition(at))
	if err != nil {
		return nil, &ResponseError{Code: codeInvalidRequest, Message: err.Error()}
	}
	return map[string]any{"range": toRange(r), "placeholder": placeholder}, nil
}

func (s *server) renameLocked(a *analysis, uri string, at Position, newName string) (any, *ResponseError) {
	files, err := a.doc.Rename(fromPosition(at), newName)
	if err != nil {
		return nil, &ResponseError{Code: codeInvalidRequest, Message: err.Error()}
	}
	if files == nil {
		return nil, nil
	}
	edit := WorkspaceEdit{Changes: map[string][]TextEdit{}}
	for _, f := range files {
		target := s.locationURILocked(a, uri, f.Path)
		edit.Changes[target] = append(edit.Changes[target], toTextEdits(f.Edits)...)
	}
	return edit, nil
}

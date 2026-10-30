package lsp

import (
	"github.com/wowsims/forever/sim/siml"
)

// The language server protocol's structures the server reads and writes. Positions are 0-based lines and UTF-16
// columns; siml's count both from 1.

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

type WorkspaceEdit struct {
	Changes map[string][]TextEdit `json:"changes"`
}

type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"`
	Code     string `json:"code,omitempty"`
	Source   string `json:"source"`
	Message  string `json:"message"`

	RelatedInformation []RelatedInformation `json:"relatedInformation,omitempty"`
}

type RelatedInformation struct {
	Location Location `json:"location"`
	Message  string   `json:"message"`
}

const (
	severityError       = 1
	severityWarning     = 2
	severityInformation = 3
)

type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

type CompletionItem struct {
	Label            string         `json:"label"`
	Kind             int            `json:"kind,omitempty"`
	Detail           string         `json:"detail,omitempty"`
	Documentation    *MarkupContent `json:"documentation,omitempty"`
	SortText         string         `json:"sortText,omitempty"`
	FilterText       string         `json:"filterText,omitempty"`
	InsertText       string         `json:"insertText,omitempty"`
	InsertTextFormat int            `json:"insertTextFormat,omitempty"`
	TextEdit         *TextEdit      `json:"textEdit,omitempty"`
	Command          *Command       `json:"command,omitempty"`
}

type Command struct {
	Title   string `json:"title"`
	Command string `json:"command"`
}

type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           int              `json:"kind"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

type FoldingRange struct {
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
	Kind      string `json:"kind,omitempty"` // comment, imports or region.
}

type DocumentLink struct {
	Range  Range  `json:"range"`
	Target string `json:"target"`
}

type CodeAction struct {
	Title       string         `json:"title"`
	Kind        string         `json:"kind"`
	Diagnostics []Diagnostic   `json:"diagnostics,omitempty"`
	IsPreferred bool           `json:"isPreferred,omitempty"`
	Edit        *WorkspaceEdit `json:"edit,omitempty"`
}

type SemanticTokens struct {
	Data []uint32 `json:"data"`
}

// The params of the textDocument requests the server answers; each reads the fields it needs.
type textDocumentParams struct {
	TextDocument struct {
		URI        string `json:"uri"`
		LanguageID string `json:"languageId"`
		Version    int    `json:"version"`
		Text       string `json:"text"`
	} `json:"textDocument"`
	ContentChanges []struct {
		Text string `json:"text"`
	} `json:"contentChanges"`
	Position Position `json:"position"`
	Range    Range    `json:"range"`
	NewName  string   `json:"newName"`
	Context  struct {
		IncludeDeclaration bool         `json:"includeDeclaration"`
		Diagnostics        []Diagnostic `json:"diagnostics"`
	} `json:"context"`
}

// A siml position as the protocol's.
func toPosition(p siml.Position) Position {
	return Position{Line: max(p.Line-1, 0), Character: max(p.Col-1, 0)}
}

func toRange(r siml.Range) Range { return Range{toPosition(r.Start), toPosition(r.End)} }

// A protocol position as siml's.
func fromPosition(p Position) siml.Position {
	return siml.Position{Line: p.Line + 1, Col: p.Character + 1}
}

func fromRange(r Range) siml.Range {
	return siml.Range{Start: fromPosition(r.Start), End: fromPosition(r.End)}
}

func toTextEdits(edits []siml.TextEdit) []TextEdit {
	out := make([]TextEdit, len(edits))
	for i, e := range edits {
		out[i] = TextEdit{Range: toRange(e.Range), NewText: e.NewText}
	}
	return out
}

func completionItem(c siml.Completion) CompletionItem {
	item := CompletionItem{
		Label:      c.Label,
		Kind:       int(c.Kind), // Numbered as the protocol's CompletionItemKind.
		Detail:     c.Detail,
		SortText:   c.SortText,
		FilterText: c.FilterText,
	}
	if c.Retrigger {
		item.Command = &Command{Title: "Suggest", Command: "editor.action.triggerSuggest"}
	}
	if c.Documentation != "" {
		item.Documentation = &MarkupContent{Kind: "markdown", Value: c.Documentation}
	}
	text := c.InsertText
	if text == "" {
		text = c.Label
	}
	item.TextEdit = &TextEdit{Range: toRange(c.Replace), NewText: text}
	return item
}

func documentSymbols(symbols []siml.Symbol) []DocumentSymbol {
	out := make([]DocumentSymbol, len(symbols))
	for i, s := range symbols {
		out[i] = DocumentSymbol{
			Name:           s.Name,
			Detail:         s.Detail,
			Kind:           max(int(s.Kind), 1),
			Range:          toRange(s.Range),
			SelectionRange: toRange(s.Selection),
			Children:       documentSymbols(s.Children),
		}
	}
	return out
}

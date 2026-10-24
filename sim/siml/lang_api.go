package siml

import (
	"slices"

	"github.com/wowsims/forever/sim/core/proto"
)

type CatalogItem struct {
	ID         int32
	Name       string
	Icon       string
	Type       proto.ItemType
	RangedType proto.RangedWeaponType
	HandType   proto.HandType
	WeaponType proto.WeaponType
	Quality    proto.ItemQuality
	Ilvl       int32
	Phase      int32
	SetName    string
	Effects    []string // The buff names of its item effects.
	OnUse      bool     // Whether it has a use, which item.<name> casts.
}

type CatalogEnchant struct {
	EffectID    int32
	Name        string
	Type        proto.ItemType
	EnchantType proto.EnchantType
	ExtraTypes  []proto.ItemType // Further item types the row goes on, as an armor kit's.
	SpellID     int32
	ItemID      int32
}

// What a completion offers, numbered as the protocol's CompletionItemKind.
type CompletionKind int

const (
	CompletionMethod        CompletionKind = 2
	CompletionFunction      CompletionKind = 3
	CompletionField         CompletionKind = 5
	CompletionVariable      CompletionKind = 6
	CompletionClass         CompletionKind = 7
	CompletionModule        CompletionKind = 9
	CompletionProperty      CompletionKind = 10
	CompletionValue         CompletionKind = 12
	CompletionKeyword       CompletionKind = 14
	CompletionFile          CompletionKind = 17
	CompletionReference     CompletionKind = 18
	CompletionEnumMember    CompletionKind = 20
	CompletionConstant      CompletionKind = 21
	CompletionTypeParameter CompletionKind = 25
)

type Completion struct {
	Label, Detail, Documentation string
	Kind                         CompletionKind
	InsertText                   string
	Replace                      Range
	SortText, FilterText         string
	Retrigger                    bool // Completion opens again once it is inserted: after `cast `, `aura(`, `target.`.
}

// What may go at a position: keywords, kinds, names, fields and values, each with the text it replaces.
func (d *Doc) Complete(at Position, c *DatabaseCatalog) []Completion { return d.complete(at, c) }

type Hover struct {
	Markdown string
	Range    Range
	SpellID  int32 // >0: the server appends the game files row for this spell
}

// What the word at a position is: a name and what it means, a kind, field, unit, section or spec, or a variable,
// group, sequence or placeholder with its declaration.
func (d *Doc) Hover(at Position, c *DatabaseCatalog) (Hover, bool) { return d.hover(at, c) }

type Token struct {
	Range     Range
	Type      string
	Modifiers []string
}

func TokenTypes() []string     { return slices.Clone(tokenTypes) }
func TokenModifiers() []string { return slices.Clone(tokenModifiers) }

// The words the schema and the table give a meaning, in file order: each on one line, none empty, none overlapping.
func (d *Doc) Tokens() []Token { return d.tokens() }

// What a symbol is, numbered as the protocol's SymbolKind.
type SymbolKind int

const (
	SymbolModule    SymbolKind = 2
	SymbolNamespace SymbolKind = 3
	SymbolFunction  SymbolKind = 12
	SymbolVariable  SymbolKind = 13
	SymbolBoolean   SymbolKind = 17
	SymbolEvent     SymbolKind = 24
)

type Symbol struct {
	Name, Detail     string
	Kind             SymbolKind
	Range, Selection Range
	Children         []Symbol
}

// The sections, with their variables, sequences, defs and when blocks, and the file's variables and defs before the
// first header, for the outline.
func (d *Doc) Symbols() []Symbol { return d.symbols() }

// Each section, each statement that runs over several lines, the block of use lines and each run of comment lines.
func (d *Doc) Folds() []Fold { return d.folds() }

// A use line's file, resolved.
type Link struct {
	Range Range
	Path  string
}

// The files the use lines name that exist.
func (d *Doc) Links() []Link { return d.linkedFiles() }

type Fix struct {
	Title      string
	Kind       string // quickfix, or refactor.rewrite for a rewrite no diagnostic asks for.
	Diagnostic Diagnostic
	Edits      []TextEdit
	Preferred  bool
}

// The code actions at a range: a word the parser does not know changed to a candidate its diagnostic names, and the
// rewrites of conditions that change the proto, which are only made when asked (SPEC.md §16.2).
func (d *Doc) QuickFixes(r Range) []Fix { return d.quickFixes(r) }

// What tools/spelldata prints, for the command and for a server that reads a row in-process.
package printer

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core/spelldata"
)

type options struct {
	// spell, family, expr, config, hover or lsp.
	mode string
	// What the mode reads: the spell, the family, the chain, the call or the file.
	arg string
	// -hover's <line>:<column>.
	position string
	pkg      string
	json     bool
	all      bool
}

const usage = "usage: go run ./tools/spelldata <id | name> [-all] [-json] | -family <class>/<Family> | -expr <call> [-package <class>] | -config <SpellConfig call | core.SpellModConfig literal> [-package <class>] | -hover <file> <line>:<column> | -lsp | -dump <dir>"

// The flags are taken in any position: `11574 -json` is how a caller writes it, and the stdlib flag
// package stops reading flags at the first positional argument. The ones that take a value read it as
// the next argument or after an `=`. An editor's language client may add its own transport flag after
// -lsp, such as --stdio.
func parseArgs(args []string) (options, error) {
	var opts options
	var positional []string
	askOne := fmt.Errorf("ask for one thing: a spell, -family, -expr, -config, -hover, -lsp or -dump")

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			positional = append(positional, arg)
			continue
		}
		name, value, valued := strings.Cut(strings.TrimLeft(arg, "-"), "=")
		switch {
		case name == "json" && !valued:
			opts.json = true
		case name == "all" && !valued:
			opts.all = true
		case name == "stdio" && !valued && opts.mode == "lsp":
		case name == "lsp" && !valued, name == "package", name == "family", name == "expr", name == "config", name == "hover", name == "dump":
			if name != "lsp" && !valued {
				if i+1 >= len(args) {
					return opts, fmt.Errorf("%s takes a value", arg)
				}
				i++
				value = args[i]
			}
			if name == "package" {
				opts.pkg = value
				continue
			}
			if opts.mode != "" {
				return opts, askOne
			}
			opts.mode, opts.arg = name, value
		default:
			return opts, fmt.Errorf("unknown argument %q", arg)
		}
	}

	switch {
	case opts.mode == "hover":
		if len(positional) != 1 {
			return opts, fmt.Errorf("-hover takes a file and a position: -hover <file> <line>:<column>")
		}
		opts.position = positional[0]
	case len(positional) > 1:
		return opts, fmt.Errorf("name one spell, not both %q and %q", positional[0], positional[1])
	case len(positional) == 1 && opts.mode != "":
		return opts, askOne
	case len(positional) == 1:
		opts.mode, opts.arg = "spell", positional[0]
	case opts.mode == "":
		return opts, errors.New(usage)
	}
	return opts, nil
}

// Answers one command line of tools/spelldata into out.
func Run(args []string, out io.Writer) error {
	opts, err := parseArgs(args)
	if err != nil {
		return err
	}

	switch opts.mode {
	case "family":
		return runFamily(out, opts)
	case "expr":
		return runExpr(out, opts)
	case "hover":
		return runHover(out, opts)
	case "config":
		return runConfig(out, opts)
	case "lsp":
		return runLSP(os.Stdin, out)
	case "dump":
		return runDump(opts.arg)
	}

	if id, err := strconv.ParseInt(opts.arg, 10, 32); err == nil {
		s, err := findSpell(int32(id))
		if err != nil {
			return err
		}
		c := newCard(s, s.Rank, 0)
		if opts.json {
			return writeJSON(out, c)
		}
		c.writeText(out)
		return nil
	}

	matches := spelldata.ByName(opts.arg)
	if len(matches) == 0 {
		return fmt.Errorf("no spell is named %q", opts.arg)
	}

	picked := matches
	if !opts.all {
		picked = []*spelldata.Spell{highestRank(matches)}
	}

	cards := make([]card, 0, len(picked))
	for _, s := range picked {
		cards = append(cards, newCard(s, s.Rank, 0))
	}
	if opts.json {
		return writeJSON(out, cards)
	}

	if len(matches) > 1 {
		for _, s := range matches {
			fmt.Fprintf(out, "%-8d %s  %s\n", s.ID, s.Name, s.Rank)
		}
		fmt.Fprintln(out)
	}
	for i, c := range cards {
		if i > 0 {
			fmt.Fprintln(out)
		}
		c.writeText(out)
	}
	return nil
}

// A ladder as a whole: its rank index, then the highest rank in full, which is the rank most callers
// are after and the one a bare Highest() reaches.
func runFamily(out io.Writer, opts options) error {
	family, err := findFamily(ladderFamilies(), opts.arg, opts.pkg)
	if err != nil {
		return err
	}

	if family.err != nil {
		return family.err
	}
	highest := family.ladder.Highest()
	top := newCard(highest, rankLabel(family, highest), 0)

	if opts.json {
		return writeJSON(out, familyJSON{Family: family.key(), Ranks: familyRows(family), Highest: top})
	}

	writeFamilyText(out, family)
	top.writeText(out)
	return nil
}

func runExpr(out io.Writer, opts options) error {
	c, err := parseChain(opts.arg)
	if err != nil {
		return err
	}
	c, err = resolveChain(c, nil, &tracer{})
	if err != nil {
		return err
	}
	result, err := evalExpr(ladderFamilies(), c, opts.pkg, nil)
	if err != nil {
		return err
	}
	doc := ""
	if n := len(c.segments); n > 0 {
		doc = methodDoc(result.owner, c.segments[n-1].name)
	}

	if opts.json {
		return writeJSON(out, exprJSON{
			card:      result.card(),
			Kind:      result.kind,
			Trail:     result.trail,
			Value:     result.value,
			Doc:       doc,
			Accessors: result.accessors,
		})
	}

	writeExprText(out, result, opts.arg, doc)
	return nil
}

func runConfig(out io.Writer, opts options) error {
	node, err := parser.ParseExpr(opts.arg)
	call, isCall := node.(*ast.CallExpr)
	lit, isLit := node.(*ast.CompositeLit)
	if err != nil || !isCall && !(isLit && isSpellModConfig(lit)) {
		return fmt.Errorf("%q is not a SpellConfig call or a core.SpellModConfig literal", opts.arg)
	}
	trace := &tracer{}
	var declarations map[string]declaration
	if ws := newWorkspace(); ws.root != "" && opts.pkg != "" {
		declarations = ws.declarations(filepath.Join(ws.root, "sim", opts.pkg), nil, 0, trace)
	}
	if isLit {
		evalSpellModConfig(lit, declarations, opts.pkg, trace).table().writeText(out)
		return nil
	}
	result, err := evalSpellConfig(call, declarations, opts.pkg, trace, nil)
	if err != nil {
		return err
	}
	result.table().writeText(out)
	return nil
}

func runHover(out io.Writer, opts options) error {
	lineText, colText, ok := strings.Cut(opts.position, ":")
	line, lineErr := strconv.Atoi(lineText)
	col, colErr := strconv.Atoi(colText)
	if !ok || lineErr != nil || colErr != nil || line < 1 || col < 1 {
		return fmt.Errorf("%q is not a position: write <line>:<column>, both counted from 1", opts.position)
	}
	text, err := os.ReadFile(opts.arg)
	if err != nil {
		return err
	}

	markdown, trace, found := newWorkspace().hover(string(text), line-1, col-1, pathURI(opts.arg), nil)
	for _, entry := range trace {
		fmt.Fprintln(os.Stderr, entry)
	}
	if !found {
		return fmt.Errorf("no hover at %s:%s", opts.arg, opts.position)
	}
	_, err = io.WriteString(out, markdown)
	return err
}

func runLSP(in io.Reader, out io.Writer) error {
	shutdown, err := serveLSP(in, out)
	if err == nil && !shutdown {
		err = fmt.Errorf("the client exited without a shutdown")
	}
	return err
}

// What the chain answered, then the row it was read off. A pick states the call as the caller wrote it,
// since nothing in it was substituted; a longer chain states the trail, which is that call with every
// name resolved to the number it stands for.
func writeExprText(out io.Writer, result *exprResult, expr, doc string) {
	c := result.card()
	switch result.kind {
	case kindSpell:
		fmt.Fprintf(out, "%s = %s\n\n", expr, result.value)
	case kindEffect:
		fmt.Fprintf(out, "%s = %s of %s\n\n", result.trail, result.value, c.heading())
	default:
		fmt.Fprintf(out, "%s = %s\n\n", result.trail, result.value)
	}

	if doc != "" {
		for line := range strings.SplitSeq(doc, "\n") {
			fmt.Fprintln(out, strings.TrimRight("    "+line, " "))
		}
		fmt.Fprintln(out)
	}

	for i, accessor := range result.accessors {
		label := ""
		if i == 0 {
			label = "accessors"
		}
		fmt.Fprintf(out, "%-9s %s\n", label, accessor)
	}
	if len(result.accessors) > 0 {
		fmt.Fprintln(out)
	}

	c.writeText(out)
}

// The rank a caller means out of several rows with one name: the highest one the game files state, and
// the lowest id among rows that state no rank, which is the ability itself ahead of the item and set
// rows that share its name.
func highestRank(matches []*spelldata.Spell) *spelldata.Spell {
	return slices.MaxFunc(matches, func(a, b *spelldata.Spell) int { return cmp.Compare(a.RankNumber(), b.RankNumber()) })
}

func findSpell(id int32) (*spelldata.Spell, error) {
	if s := spelldata.Find(id); s != spelldata.Nil {
		return s, nil
	}
	return nil, fmt.Errorf("spell %d is not in the store", id)
}

// The row a chain reached, and what the chain answered: the kind of value, the chain with every name
// resolved, the value as it prints, the doc comment of the accessor that answered it and, where it
// stopped on an effect, the accessors that read something off it.
type exprJSON struct {
	card
	Kind      string   `json:"kind"`
	Trail     string   `json:"trail"`
	Value     string   `json:"value"`
	Doc       string   `json:"doc,omitempty"`
	Accessors []string `json:"accessors,omitempty"`
}

func writeJSON(out io.Writer, v any) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

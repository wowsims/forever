package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/wowsims/forever/sim/core/spelldata"
)

// One row as the tool states it. The text form, the hover's markdown and -json are renderings of it.
type card struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	// The store's rank column, or `rank n of N` on a talent's rank.
	Rank string `json:"rank"`
	// The ladder calls a class file reaches this id through.
	Ladder  []string `json:"ladder"`
	Header  []field  `json:"header"`
	Effects []Line   `json:"effects"`
	Wowhead string   `json:"wowhead"`
}

type field struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// One effect as both readings: what it does in words, and the game files' own columns, which is what a
// developer checks the words against. Human is "unrecognised shape" on an effect no branch covers,
// and the literal is then all there is to read. Read marks the effect a chain read.
type Line struct {
	Human   string `json:"human"`
	Literal string `json:"literal"`
	Read    bool   `json:"read,omitempty"`
}

// The card of a row under the rank it was reached as, with effect read (counted from 1) marked.
func newCard(s *spelldata.Spell, rank string, read int) card {
	header := headerFields(s)
	if proc := procSummary(s); proc != "" {
		header = append(header, field{"proc", proc})
	}
	if refs := refList(s); len(refs) > 0 {
		header = append(header, field{"refs", strings.Join(refs, ", ")})
	}

	effects := effectLines(s)
	if read > 0 && read <= len(effects) {
		effects[read-1].Read = true
	}

	return card{
		ID:      s.ID,
		Name:    s.Name,
		Rank:    rank,
		Ladder:  ladderRefs(s.ID),
		Header:  header,
		Effects: effects,
		Wowhead: fmt.Sprintf("https://www.wowhead.com/forever/spell=%d", s.ID),
	}
}

func (c card) heading() string {
	if c.Rank == "" {
		return fmt.Sprintf("%d %s", c.ID, c.Name)
	}
	return fmt.Sprintf("%d %s (%s)", c.ID, c.Name, c.Rank)
}

func title(s *spelldata.Spell) string {
	return card{ID: s.ID, Name: s.Name, Rank: s.Rank}.heading()
}

func (c card) writeText(out io.Writer) {
	fmt.Fprintln(out, join(c.heading(), strings.Join(c.Ladder, "  ")))
	for _, f := range c.Header {
		fmt.Fprintf(out, "%-9s %s\n", f.Key, f.Value)
	}

	if len(c.Effects) > 0 {
		fmt.Fprintln(out)
	}
	for i, line := range c.Effects {
		label := fmt.Sprintf("effect %-2d", i+1)
		if line.Read {
			label = fmt.Sprintf("effect %d (read)", i+1)
		}
		fmt.Fprintf(out, "%s %s\n", label, line.Human)
		fmt.Fprintf(out, "%9s %s\n", "", line.Literal)
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, c.Wowhead)
}

// A heading, the ladder calls that reach the row, the row's own columns as a table titled with its
// name, then its effects below a rule.
func (c card) writeMarkdown(md *strings.Builder) {
	name, heading := c.Name, fmt.Sprintf("%d %s", c.ID, c.Name)
	if c.Rank != "" {
		name = fmt.Sprintf("%s (%s)", c.Name, c.Rank)
		heading += " · " + c.Rank
	}

	fmt.Fprintf(md, "### %s\n", heading)
	for _, ref := range c.Ladder {
		fmt.Fprintf(md, "`%s`  \n", ref)
	}
	if len(c.Header) > 0 {
		fmt.Fprintf(md, "\n| | %s |\n|--:|:--|\n", cell(name))
		for _, f := range c.Header {
			fmt.Fprintf(md, "| **%s** | %s |\n", f.Key, cell(f.Value))
		}
	}

	effectParagraphs(md, c.Effects, false)
	fmt.Fprintf(md, "\n[Wowhead](%s)\n", c.Wowhead)
}

// Each effect's wording over its game files row, a paragraph each: a table cell holds no line break
// short of HTML, which not every editor renders, and a long row then scrolls the hover sideways. The
// row's fields are one inline code span each, so it wraps between fields.
func effectParagraphs(md *strings.Builder, effects []Line, onlyRead bool) {
	if len(effects) == 0 {
		return
	}
	md.WriteString("\n---\n")
	for i, effect := range effects {
		if onlyRead && !effect.Read {
			continue
		}
		marker := fmt.Sprint(i + 1)
		if effect.Read {
			marker += " ▶"
		}
		fields := strings.Fields(effect.Literal)
		for i, field := range fields {
			fields[i] = "`" + field + "`"
		}
		fmt.Fprintf(md, "\n**%s** · %s  \n%s\n", marker, effect.Human, strings.Join(fields, " "))
	}
}

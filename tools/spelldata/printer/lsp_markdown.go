package printer

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/wowsims/forever/sim/core/spelldata"
)

func cell(text string) string {
	return strings.ReplaceAll(text, "|", `\|`)
}

func codeCell(text string) string {
	return "`" + cell(text) + "`"
}

func idMarkdown(s *spelldata.Spell) string {
	var md strings.Builder
	newCard(s, s.Rank, 0).writeMarkdown(&md)
	return md.String()
}

func familyMarkdown(f *ladderFamily) string {
	var md strings.Builder
	fmt.Fprintf(&md, "### %s\n\n", f.key())
	rows := familyRows(f)
	if len(rows) > 0 && rows[0].Value != "" {
		md.WriteString("| id | name | rank | call | value |\n|--:|:--|:--|:--|:--|\n")
		for _, row := range rows {
			fmt.Fprintf(&md, "| %d | %s | %s | %s | %s |\n", row.ID, cell(row.Name), cell(row.Rank), codeCell(row.Accessor), cell(row.Value))
		}
	} else {
		md.WriteString("| id | name | rank | call |\n|--:|:--|:--|:--|\n")
		for _, row := range rows {
			fmt.Fprintf(&md, "| %d | %s | %s | %s |\n", row.ID, cell(row.Name), cell(row.Rank), codeCell(row.Accessor))
		}
	}
	md.WriteString("\n")
	highest := f.ladder.Highest()
	newCard(highest, rankLabel(f, highest), 0).writeMarkdown(&md)
	return md.String()
}

// With links false the rank selector names the current rank only: an editor that cannot run a command
// link picks another rank through the code actions.
func exprMarkdown(result *exprResult, hover chainHover, links bool) string {
	var md strings.Builder

	c := result.card()
	called := ""
	if n := len(hover.chain.segments); n > 0 {
		called = hover.chain.segments[n-1].text(true)
		// An index reads as the accessor it indexes: `Refs()[0]`.
		if hover.chain.segments[n-1].index && n > 1 {
			called = hover.chain.segments[n-2].text(true) + called
		}
	}

	// Only show the rank selector when the cursor is actually on the
	// rank parameter itself. Expressions which resolve through rank still
	// use the selected virtual rank, but do not show the selector again.
	showRankSelector := hover.rank != nil &&
		hover.name == hover.rank.parameter

	if showRankSelector && result.family != nil {
		selectedRank := int32(result.family.ladder.Len())

		if hover.rank.rank > 0 {
			selectedRank = hover.rank.rank
		}

		fmt.Fprintf(
			&md,
			"**Rank:** **%d** / %d\n\n",
			selectedRank,
			result.family.ladder.Len(),
		)

		if !links {
			md.WriteString("Pick another rank from the code actions.\n\n")
		} else {
			for rank := int32(1); rank <= int32(result.family.ladder.Len()); rank++ {
				args, _ := json.Marshal(map[string]any{
					"parameter": hover.rank.parameter,
					"function":  hover.rank.function,
					"rank":      rank,
				})

				if rank > 1 {
					md.WriteString(" ")
				}

				if rank == selectedRank {
					fmt.Fprintf(&md, "**[%d]**", rank)
				} else {
					fmt.Fprintf(
						&md,
						"[%d](command:wowsims-spelldata.changeRank?%s)",
						rank,
						url.QueryEscape(string(args)),
					)
				}
			}

			md.WriteString("\n\n")
		}
	}

	// Accessors on a Ladder.Each virtual rank should render only the
	// resolved spell/rank heading and the value under the cursor.
	// The full spell card and evaluation trail are redundant here.
	// Another name bound to a row or an effect renders in full, as any bound name does.
	if hover.rank != nil && hover.parameter != nil && len(hover.chain.segments) > 0 &&
		(hover.name == "" || hover.name == hover.rank.parameter || result.kind == kindValue) {
		fmt.Fprintf(&md, "### %s\n\n", c.heading())
		fmt.Fprintf(&md, "`%s` = **%s**\n\n", called, result.value)
		return md.String()
	}

	switch result.kind {
	case kindSpell:
		c.writeMarkdown(&md)

	case kindEffect:
		fmt.Fprintf(&md, "`%s` = **%s** of %s\n\n", called, result.value, c.heading())
		fmt.Fprintf(&md, "`%s`\n", result.trail)
		if result.readEffect > 0 {
			effectParagraphs(&md, c.Effects, true)
		}
		if len(result.accessors) > 0 {
			md.WriteString("\n")
		}
		for _, accessor := range result.accessors {
			fmt.Fprintf(&md, "`%s`  \n", accessor)
		}
		fmt.Fprintf(&md, "\n[Wowhead](%s)\n", c.Wowhead)

	default:
		label := called
		if hover.name != "" {
			label = hover.name
		}

		if hover.name == "" {
			c.writeMarkdown(&md)
			md.WriteString("\n")
			fmt.Fprintf(&md, "`%s` = **%s**\n\n", label, result.value)
			fmt.Fprintf(&md, "`%s`\n", result.trail)
		} else {
			fmt.Fprintf(&md, "`%s` = **%s**\n\n", label, result.value)
			fmt.Fprintf(&md, "`%s`\n\n", result.trail)
			c.writeMarkdown(&md)
		}
	}

	return md.String()
}

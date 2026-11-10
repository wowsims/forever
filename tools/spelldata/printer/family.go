package printer

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/wowsims/forever/sim/core/spelldata"
)

// The ladder a caller names: `warrior/Execute`, or `Execute` alone when one class states it, with the
// package the caller's file sits in as the tiebreak. A name several classes state and no package to
// choose between them is an error rather than a guess.
func findFamily(index map[string]*ladderFamily, spec, pkg string) (*ladderFamily, error) {
	field := spec
	if class, name, ok := strings.Cut(spec, "/"); ok {
		pkg, field = class, name
	}
	if field == "" {
		return nil, fmt.Errorf("name a ladder, as <class>/<Family> or <Family>")
	}

	if pkg != "" {
		family, ok := index[pkg+"/"+field]
		if !ok {
			return nil, fmt.Errorf("%s states no ladder named %q", pkg, field)
		}
		return family, nil
	}

	var found []*ladderFamily
	for _, family := range index {
		if family.field == field {
			found = append(found, family)
		}
	}
	switch len(found) {
	case 0:
		return nil, fmt.Errorf("no class file states a ladder named %q", field)
	case 1:
		return found[0], nil
	}

	classes := make([]string, 0, len(found))
	for _, family := range found {
		classes = append(classes, family.pkg)
	}
	sort.Strings(classes)
	return nil, fmt.Errorf("%s is a ladder in %s - name one as <class>/%s",
		field, strings.Join(classes, ", "), field)
}

// Whether any class file states a ladder by this name, which is what makes a bare `Execute.Highest()`
// a ladder call rather than a name the package has yet to bind.
func isFamilyName(name string) bool {
	loadLadders.Do(scanLadders)
	return familyFields[name]
}

// The rank of a talent's ladder a row is, counted from 1, or 0 on a row that is not one of them: a
// talent's rank is not the store's rank column, so it is stated by position.
func (f *ladderFamily) talentRank(s *spelldata.Spell) int32 {
	var rank int32
	if f != nil && f.talent {
		f.ladder.Each(func(n int32, r *spelldata.Spell) {
			if r == s {
				rank = n
			}
		})
	}
	return rank
}

// The rank a row a ladder reached is: the talent rank where it is one, else the store's rank column.
func rankLabel(f *ladderFamily, s *spelldata.Spell) string {
	if rank := f.talentRank(s); rank > 0 {
		return fmt.Sprintf("rank %d of %d", rank, f.ladder.Len())
	}
	return s.Rank
}

type familyRow struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	Rank     string `json:"rank"`
	Accessor string `json:"accessor"`
	Value    string `json:"value,omitempty"`
}

type familyJSON struct {
	Family  string      `json:"family"`
	Ranks   []familyRow `json:"ranks"`
	Highest card        `json:"highest"`
}

// One row per rank of the ladder, with effect 1's value where it changes by rank.
func familyRows(f *ladderFamily) []familyRow {
	first := f.ladder.Rank(1).EffectN(1).BasePoints
	varies := false
	f.ladder.Each(func(_ int32, s *spelldata.Spell) {
		varies = varies || s.EffectN(1).BasePoints != first
	})

	rows := make([]familyRow, 0, f.ladder.Len())
	f.ladder.Each(func(rank int32, s *spelldata.Spell) {
		row := familyRow{ID: s.ID, Name: s.Name, Rank: rankLabel(f, s), Accessor: fmt.Sprintf("Rank(%d)", rank)}
		if rank == f.ladder.Len() {
			row.Accessor = "Highest()"
		}
		if varies {
			row.Value = "effect 1 = " + number(s.EffectN(1).BasePoints)
		}
		rows = append(rows, row)
	})
	return rows
}

func writeFamilyText(out io.Writer, f *ladderFamily) {
	rows := familyRows(f)

	width, rankWidth := 0, 8
	for _, row := range rows {
		width = max(width, len(row.Name))
		rankWidth = max(rankWidth, len(row.Rank))
	}

	fmt.Fprintf(out, "%s spellData.%s\n", f.pkg, f.field)
	for _, row := range rows {
		if row.Value == "" {
			fmt.Fprintf(out, "%-8d %-*s  %-*s %s\n", row.ID, width, row.Name, rankWidth, row.Rank, row.Accessor)
			continue
		}
		fmt.Fprintf(out, "%-8d %-*s  %-*s %-9s %s\n", row.ID, width, row.Name, rankWidth, row.Rank, row.Accessor, row.Value)
	}
	fmt.Fprintln(out)
}

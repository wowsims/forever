package database

import (
	"database/sql"
	"fmt"
	"go/format"
	"sort"
	"strings"

	"github.com/wowsims/forever/sim/core/dbcenums"
)

// The hunter's pet families, from the game files' skill lines: one "Pet - <family>" line per family in
// the class-skills category, told apart from a warlock's demons by the Hunter Pet Scaling passive
// every hunter line carries. Two lines can name one family (the game files have two "Pet - Bat"), and
// they merge.
//
// A family is what the sim reads off it: the Tamed Pet Passive (DND) that states the family's
// damage, armor and health, and the abilities the line teaches, pet side. Every ability is already the
// triggered ladder of a hunter family, so the file names those ladders, and an ability no hunter ladder
// carries is reported.
type petFamily struct {
	// The line's name after "Pet - ", and the Go identifier fieldNameOf makes of it, which is the
	// map key and what proto.HunterOptions_PetType names the family.
	Name  string
	Field string

	// The family's Tamed Pet Passive: the one on the line with an A_MOD_BASE_RESISTANCE_PCT effect,
	// which the shared ones do not carry.
	Passive int32

	// The hunter ladders whose triggered ladder the family learns, in name order.
	Abilities []string
}

const hunterPetScaling = "Hunter Pet Scaling"
const tamedPetPassive = "Tamed Pet Passive (DND)"

func discoverPetFamilies(db *sql.DB, ladders []rankLadder, triggered map[string][]triggeredRow) ([]petFamily, []string, error) {
	rows, err := db.Query(`
		SELECT sl.DisplayName_lang, sla.Spell, n.Name_lang, s.NameSubtext_lang,
		       (COALESCE(json_extract(sm.Attributes, '$[0]'), 0) & ?) != 0,
		       EXISTS (SELECT 1 FROM SpellEffect se WHERE se.SpellID = sla.Spell AND se.EffectAura = ?)
		FROM SkillLine sl
		JOIN SkillLineAbility sla ON sla.SkillLine = sl.ID
		JOIN SpellName n ON n.ID = sla.Spell
		JOIN Spell s ON s.ID = sla.Spell
		LEFT JOIN SpellMisc sm ON sm.SpellID = sla.Spell AND sm.DifficultyID = 0
		WHERE sl.CategoryID = 7 AND sl.DisplayName_lang LIKE 'Pet - %'
		AND EXISTS (
			SELECT 1 FROM SkillLineAbility h JOIN SpellName hn ON hn.ID = h.Spell
			WHERE h.SkillLine = sl.ID AND hn.Name_lang = ?)
		ORDER BY sl.DisplayName_lang, n.Name_lang, sla.Spell`,
		dbcenums.ATTR_PASSIVE, dbcenums.A_MOD_BASE_RESISTANCE_PCT, hunterPetScaling)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	triggeredIDs := map[string]map[int32]bool{}
	for _, l := range ladders {
		if trig := triggered[l.Field]; trig != nil {
			ids := map[int32]bool{}
			for _, row := range trig {
				ids[row.SpellID] = true
			}
			triggeredIDs[l.Name] = ids
		}
	}

	type abilityRow struct {
		name string
		id   int32
	}
	byFamily := map[string]*petFamily{}
	abilities := map[string][]abilityRow{}
	var order []string
	var skipped []string
	for rows.Next() {
		var line, name, subtext string
		var id int32
		var passive, familyPassive bool
		if err := rows.Scan(&line, &id, &name, &subtext, &passive, &familyPassive); err != nil {
			return nil, nil, err
		}
		family := strings.TrimPrefix(line, "Pet - ")
		f := byFamily[family]
		if f == nil {
			f = &petFamily{Name: family, Field: fieldNameOf(family)}
			byFamily[family] = f
			order = append(order, family)
		}

		switch {
		case name == tamedPetPassive:
			if !familyPassive {
				continue
			}
			if f.Passive != 0 && f.Passive != id {
				skipped = append(skipped, fmt.Sprintf("%s: two family passives, %d and %d", family, f.Passive, id))
			}
			f.Passive = id
		case name == hunterPetScaling, passive:
			// Shared by every family, and stated by the server rather than the rows.
		case subtext == "" || rankSubtext.MatchString(subtext):
			abilities[family] = append(abilities[family], abilityRow{name, id})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	var families []petFamily
	for _, family := range order {
		f := byFamily[family]
		if f.Field == "" {
			skipped = append(skipped, fmt.Sprintf("%s: no usable Go identifier", family))
			continue
		}
		if f.Passive == 0 {
			skipped = append(skipped, fmt.Sprintf("%s: no Tamed Pet Passive states its damage, armor and health", family))
			continue
		}

		seen := map[string]bool{}
		var names []string
		for _, a := range abilities[family] {
			if !seen[a.name] {
				seen[a.name] = true
				names = append(names, a.name)
			}
		}
		sort.Strings(names)
		for _, name := range names {
			ids := triggeredIDs[name]
			if ids == nil {
				skipped = append(skipped, fmt.Sprintf("%s: %s: no hunter family teaches it", family, name))
				continue
			}
			missing := false
			for _, a := range abilities[family] {
				if a.name == name && !ids[a.id] {
					skipped = append(skipped, fmt.Sprintf("%s: %s: spell %d is not a rank the hunter's %s teaches", family, name, a.id, fieldNameOf(name)))
					missing = true
				}
			}
			if !missing {
				f.Abilities = append(f.Abilities, fieldNameOf(name))
			}
		}
		families = append(families, *f)
	}
	sort.Slice(families, func(i, j int) bool { return families[i].Name < families[j].Name })
	sort.Strings(skipped)
	return families, skipped, nil
}

// The families' passives: what the file above reaches that the hunter's own ladders do not, so the
// store has to carry them as roots of their own.
func petFamilyRoots(families []petFamily) []int32 {
	ids := make([]int32, 0, len(families))
	for _, f := range families {
		ids = append(ids, f.Passive)
	}
	return ids
}

func renderPetFamiliesFile(families []petFamily, skipped []string) ([]byte, error) {
	var b strings.Builder
	b.WriteString("// Code generated by tools/database/gen_spelldata. DO NOT EDIT.\n\n")
	b.WriteString("package hunter\n\n")
	b.WriteString("import (\n\t\"github.com/wowsims/forever/sim/core/spelldata\"\n)\n\n")

	if len(skipped) > 0 {
		b.WriteString("// Not generated:\n")
		for _, s := range skipped {
			fmt.Fprintf(&b, "//   %s\n", s)
		}
		b.WriteString("\n")
	}

	b.WriteString(`// A pet family as the client's skill lines describe it: the Tamed Pet Passive that states the
// family's damage, armor and health, and the abilities the family learns, pet side, as the
// triggered ladders of the hunter families that teach them.
type generatedPetFamily struct {
	Name      string
	Passive   spelldata.Ladder
	Abilities []spelldata.Ladder
}

// Keyed by the family's Go identifier, which is what proto.HunterOptions_PetType names it.
var petFamilies = map[string]generatedPetFamily{
`)
	for _, f := range families {
		refs := make([]string, len(f.Abilities))
		for i, field := range f.Abilities {
			refs[i] = "spellData." + field + "Triggered"
		}
		fmt.Fprintf(&b, "\t%q: {Name: %q, Passive: spelldata.Ranked(%d), Abilities: []spelldata.Ladder{%s}},\n",
			f.Field, f.Name, f.Passive, strings.Join(refs, ", "))
	}
	b.WriteString("}\n")

	out, err := format.Source([]byte(b.String()))
	if err != nil {
		return nil, fmt.Errorf("generated pet families file does not parse, refusing to write it: %w", err)
	}
	return out, nil
}

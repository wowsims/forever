package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go/format"
	"math"
	"sort"
	"strings"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/tools/database/dbc"
)

// The hunter's arrows and quivers, from the game files' items: every Projectile - Arrow and Quiver -
// Quiver item the item generator would not throw out, by the DEPRECATED flag or the name deny-list.
// Nothing hand-maintained.
//
// An arrow is the damage per second it adds, which the item row does not state: the game reads it
// off ItemDamageAmmo by the arrow's item level and quality, and its tooltip rounds that down to the
// half point: Thorium Headed Arrow's 17.71 reads "Adds 17.5 damage per second", Ice Threaded
// Arrow's 16.79 reads 16.5, Rough Arrow's 1.82 reads 1.5.
//
// A quiver is the ranged attack speed its equip spell's A_MOD_RANGED_HASTE_QUIVER effect states.
// Quivers that state the same speed are one choice to the sim, so only the highest item level of
// them is kept.
type ammoItem struct {
	// The item's name, and the Go identifier fieldNameOf makes of it, which is the map key and what
	// the proto.HunterOptions enum names the item.
	Name   string
	Field  string
	ItemID int32
	// The arrow's damage per second, or the quiver's ranged attack speed in percent.
	Value float64
}

const itemClassProjectile = 6
const itemSubclassArrow = 2
const itemClassQuiver = 11
const itemSubclassQuiver = 2

func discoverArrows(db *sql.DB) ([]ammoItem, []string, error) {
	rows, err := db.Query(`
		SELECT i.ID, s.Display_lang, COALESCE(json_extract(s.Flags, '$[0]'), 0), a.Quality, s.OverallQualityID
		FROM Item i
		JOIN ItemSparse s ON s.ID = i.ID
		LEFT JOIN ItemDamageAmmo a ON a.ItemLevel = s.ItemLevel
		WHERE i.ClassID = ? AND i.SubclassID = ?
		ORDER BY s.Display_lang, i.ID`,
		itemClassProjectile, itemSubclassArrow)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var arrows []ammoItem
	var skipped []string
	for rows.Next() {
		var id int32
		var name string
		var flags int64
		var qualities sql.NullString
		var quality int
		if err := rows.Scan(&id, &name, &flags, &qualities, &quality); err != nil {
			return nil, nil, err
		}
		a, ok := ammoItemOf(id, name, flags, &skipped)
		if !ok {
			continue
		}
		var dps []float64
		if qualities.Valid {
			if err := json.Unmarshal([]byte(qualities.String), &dps); err != nil {
				return nil, nil, fmt.Errorf("%s (%d): ItemDamageAmmo quality row: %w", name, id, err)
			}
		}
		if quality >= len(dps) || dps[quality] <= 0 {
			skipped = append(skipped, fmt.Sprintf("%s (%d): ItemDamageAmmo states no damage for its item level and quality", name, id))
			continue
		}
		a.Value = math.Floor(dps[quality]*2) / 2
		arrows = append(arrows, a)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	sort.SliceStable(arrows, func(i, j int) bool { return arrows[i].Value < arrows[j].Value })
	return arrows, skipped, nil
}

func discoverQuivers(db *sql.DB) ([]ammoItem, []string, error) {
	// The highest item level first, so the first quiver seen at a speed is the one kept.
	rows, err := db.Query(`
		SELECT i.ID, s.Display_lang, COALESCE(json_extract(s.Flags, '$[0]'), 0), se.EffectBasePointsF
		FROM Item i
		JOIN ItemSparse s ON s.ID = i.ID
		JOIN ItemXItemEffect x ON x.ItemID = i.ID
		JOIN ItemEffect e ON e.ID = x.ItemEffectID
		JOIN SpellEffect se ON se.SpellID = e.SpellID AND se.EffectAura = ?
		WHERE i.ClassID = ? AND i.SubclassID = ?
		ORDER BY s.ItemLevel DESC, s.OverallQualityID DESC, i.ID`,
		dbcenums.A_MOD_RANGED_HASTE_QUIVER, itemClassQuiver, itemSubclassQuiver)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var quivers []ammoItem
	var skipped []string
	kept := map[float64]bool{}
	for rows.Next() {
		var id int32
		var name string
		var flags int64
		var haste float64
		if err := rows.Scan(&id, &name, &flags, &haste); err != nil {
			return nil, nil, err
		}
		q, ok := ammoItemOf(id, name, flags, &skipped)
		if !ok || haste <= 0 || kept[haste] {
			continue
		}
		kept[haste] = true
		q.Value = haste
		quivers = append(quivers, q)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	sort.SliceStable(quivers, func(i, j int) bool { return quivers[i].Value < quivers[j].Value })
	return quivers, skipped, nil
}

// The item as the tables name it, or false for an item the item generator throws out or that has no
// Go identifier, the latter reported in skipped.
func ammoItemOf(id int32, name string, flags int64, skipped *[]string) (ammoItem, bool) {
	if dbc.ItemStaticFlags0(flags)&dbc.DEPRECATED != 0 {
		return ammoItem{}, false
	}
	for _, re := range DenyListNameRegexes {
		if re.MatchString(name) {
			return ammoItem{}, false
		}
	}
	field := fieldNameOf(name)
	if field == "" {
		*skipped = append(*skipped, fmt.Sprintf("%s (%d): no usable Go identifier", name, id))
		return ammoItem{}, false
	}
	return ammoItem{Name: name, Field: field, ItemID: id}, true
}

func renderAmmoFile(arrows, quivers []ammoItem, skipped []string) ([]byte, error) {
	var b strings.Builder
	b.WriteString("// Code generated by tools/database/gen_spelldata. DO NOT EDIT.\n\n")
	b.WriteString("package hunter\n\n")

	if len(skipped) > 0 {
		sort.Strings(skipped)
		b.WriteString("// Not generated:\n")
		for _, s := range skipped {
			fmt.Fprintf(&b, "//   %s\n", s)
		}
		b.WriteString("\n")
	}

	b.WriteString(`// An arrow as the client's items describe it: the damage per second it adds to a bow or crossbow,
// ItemDamageAmmo's value for its item level and quality rounded down to the half point, as the
// tooltip reads it.
type generatedArrow struct {
	Name   string
	ItemID int32
	DPS    float64
}

// Keyed by the arrow's Go identifier, which is what proto.HunterOptions_Ammo names it.
var arrows = map[string]generatedArrow{
`)
	for _, a := range arrows {
		fmt.Fprintf(&b, "\t%q: {Name: %q, ItemID: %d, DPS: %g},\n", a.Field, a.Name, a.ItemID, a.Value)
	}
	b.WriteString(`}

// A quiver as the client's items describe it: the ranged attack speed, in percent, its equip spell
// grants. One quiver a speed, the highest item level of those that grant it.
type generatedQuiver struct {
	Name   string
	ItemID int32
	Haste  float64
}

// Keyed by the quiver's Go identifier, which is what proto.HunterOptions_QuiverBonus names it.
var quivers = map[string]generatedQuiver{
`)
	for _, q := range quivers {
		fmt.Fprintf(&b, "\t%q: {Name: %q, ItemID: %d, Haste: %g},\n", q.Field, q.Name, q.ItemID, q.Value)
	}
	b.WriteString("}\n")

	out, err := format.Source([]byte(b.String()))
	if err != nil {
		return nil, fmt.Errorf("generated ammo file does not parse, refusing to write it: %w", err)
	}
	return out, nil
}

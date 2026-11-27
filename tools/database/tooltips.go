package database

import (
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"os"
	"slices"
	"strconv"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/dbc"
	"github.com/wowsims/forever/tools/tooltip"
)

// One spell's tooltip for the site (ui/sim/proto/tooltips.ts): its name and rank, the cost, range, cast and
// cooldown lines as [left, right] pairs, and its description and buff text compiled by tooltip.Compile.
type spellTooltip struct {
	Name  string            `json:"n"`
	Rank  string            `json:"r,omitempty"`
	Lines [][2]string       `json:"h,omitempty"`
	Text  []tooltip.Segment `json:"d,omitempty"`
	Aura  []tooltip.Segment `json:"a,omitempty"`
	// A talent's points and, where its curve gives each rank its own amounts, the description at each rank
	// from rank 1; the description above is the top rank's, as the row states it.
	MaxPoints int32               `json:"m,omitempty"`
	RankTexts [][]tooltip.Segment `json:"x,omitempty"`
}

// Writes the tooltip of every spell the site can show, the spells the sim's store holds and every talent in
// the trees (their spell ids and points), to path. A condition on a rank of a talent the trees offer is left
// to the player's points, one on a talent they do not offer is false; the store's rows give the header lines.
func WriteSpellTooltips(path string, instance *dbc.DBC, treeTalents map[int32]int32, rawTalents []RawTalent) error {
	ids := map[int32]bool{}
	talents := map[int64]tooltip.Talent{}
	for id := range treeTalents {
		ids[id] = true
		talents[int64(id)] = tooltip.Talent{Spell: int64(id), Rank: 1}
	}
	for _, raw := range rawTalents {
		ranks, err := parseIntArrayField(raw.SpellRank, 9)
		if err != nil || len(ranks) == 0 || ranks[0] == 0 {
			continue
		}
		first := int64(ranks[0])
		offered := ids[int32(first)]
		for i, id := range ranks {
			if id == 0 {
				continue
			}
			talent := tooltip.Talent{Rank: i + 1}
			if offered {
				talent.Spell = first
			}
			talents[int64(id)] = talent
		}
	}
	for _, s := range spelldata.All() {
		ids[s.ID] = true
	}

	provider := tooltip.DBCTooltipDataProvider{DBC: instance}
	tooltips := map[int32]spellTooltip{}
	failed := 0
	for _, id := range slices.Sorted(maps.Keys(ids)) {
		row, ok := instance.Spells[int(id)]
		if !ok || row.NameLang == "" {
			continue
		}
		t := spellTooltip{Name: row.NameLang, Rank: row.Subtext, Lines: headerLines(spelldata.Find(id)), MaxPoints: treeTalents[id]}
		t.RankTexts = rankTexts(provider, id, row.Description, treeTalents[id], talents)
		var err error
		if row.Description != "" {
			if t.Text, err = tooltip.Compile(row.Description, provider, int64(id), talents); err != nil {
				failed++
			}
		}
		if row.AuraDescription != "" {
			if t.Aura, err = tooltip.Compile(row.AuraDescription, provider, int64(id), talents); err != nil {
				failed++
			}
		}
		tooltips[id] = t
	}

	data, err := json.Marshal(tooltips)
	if err != nil {
		return err
	}
	fmt.Printf("Wrote %d spell tooltips to %s (%d KB); %d texts did not compile\n", len(tooltips), path, len(data)/1024, failed)
	return os.WriteFile(path, data, 0o644)
}

// A talent's description at each rank, read off the store's ladder for it; none when every rank reads alike.
func rankTexts(provider tooltip.DBCTooltipDataProvider, id int32, description string, maxPoints int32, talents map[int64]tooltip.Talent) (texts [][]tooltip.Segment) {
	if maxPoints < 2 || description == "" || spelldata.Find(id) == spelldata.Nil {
		return nil
	}
	// A curve whose length disagrees with the tree's points panics in the ladder; that talent keeps the row's text.
	defer func() {
		if recover() != nil {
			texts = nil
		}
	}()
	ladder := spelldata.Talent(id, maxPoints)
	differ := false
	for n := int32(1); n <= maxPoints; n++ {
		rank := tooltip.RankProvider{DBCTooltipDataProvider: provider, SpellID: int64(id), Rank: ladder.Rank(n)}
		text, err := tooltip.Compile(description, rank, int64(id), talents)
		if err != nil {
			return nil
		}
		differ = differ || (len(texts) > 0 && fmt.Sprint(text) != fmt.Sprint(texts[0]))
		texts = append(texts, text)
	}
	if !differ {
		return nil
	}
	return texts
}

// The lines under a spell's name as the game shows them: cost and range, then cast time and cooldown.
func headerLines(s *spelldata.Spell) [][2]string {
	if s == nil || s.IsPassive() {
		return nil
	}
	var cost string
	for i := range s.Powers {
		p := &s.Powers[i]
		switch {
		case p.CostPct != 0:
			cost = strconv.FormatFloat(float64(p.CostPct), 'f', -1, 32) + "% of base mana"
		case p.Cost != 0:
			cost = strconv.FormatFloat(s.PowerCost(p.Type), 'f', -1, 64) + " " + powerLabel(p.Type)
		}
	}
	var reach string
	switch {
	case s.MaxRange == 0:
	case s.MaxRange <= 5:
		reach = "Melee Range"
	case s.MinRange != 0:
		reach = fmt.Sprintf("%s-%s yd range", trim(float64(s.MinRange)), trim(float64(s.MaxRange)))
	default:
		reach = trim(float64(s.MaxRange)) + " yd range"
	}
	cast := "Instant"
	switch {
	case s.IsChanneled():
		cast = "Channeled"
	case s.CastTimeMs > 0:
		cast = trim(float64(s.CastTimeMs)/1000) + " sec cast"
	}
	var cooldown string
	if ms := max(s.CooldownMs, s.CategoryCooldownMs); ms > 0 {
		if ms >= 60000 && ms%60000 == 0 {
			cooldown = strconv.Itoa(int(ms/60000)) + " min cooldown"
		} else {
			cooldown = trim(float64(ms)/1000) + " sec cooldown"
		}
	}
	lines := [][2]string{{cast, cooldown}}
	if cost != "" || reach != "" {
		lines = append([][2]string{{cost, reach}}, lines...)
	}
	return lines
}

func trim(v float64) string {
	return strconv.FormatFloat(math.Round(v*100)/100, 'f', -1, 64)
}

func powerLabel(t dbcenums.PowerType) string {
	switch t {
	case dbcenums.POWER_RAGE:
		return "Rage"
	case dbcenums.POWER_FOCUS:
		return "Focus"
	case dbcenums.POWER_ENERGY:
		return "Energy"
	case dbcenums.POWER_HEALTH:
		return "Health"
	}
	return "Mana"
}

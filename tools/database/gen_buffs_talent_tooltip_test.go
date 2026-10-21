package database

import "testing"

// Vindication's talent states its effect only in its tooltip: rank/N of the debuff's amount.
func TestTalentPricedByItsTooltip(t *testing.T) {
	for _, row := range resolveCommittedBuffs(t) {
		if row.Field != "vindication" {
			continue
		}
		ranks := talentNodeRanks(t, row.Talent)
		if row.TalentRanks == 0 || row.TalentRanks != ranks || row.TalentDivisor <= 1 || row.TalentSpellID != row.Talent {
			t.Errorf("vindication resolved to ranks %d divisor %d talent %d, want its node's %d ranks, a tooltip divisor and talent %d",
				row.TalentRanks, row.TalentDivisor, row.TalentSpellID, ranks, row.Talent)
		}
		return
	}
	t.Fatal("no vindication row")
}

// Dark Inspiration's own tooltip shows its talent's amount, which the buff's row states at full rank:
// each rank is its share of the node's ranks.
func TestTalentPricedByTheBuffsTooltip(t *testing.T) {
	for _, row := range resolveCommittedBuffs(t) {
		if row.Field != "dark_inspiration" {
			continue
		}
		ranks := talentNodeRanks(t, row.Talent)
		if ranks == 0 || row.TalentRanks != ranks || row.TalentDivisor != ranks || row.TalentSpellID != row.Talent {
			t.Errorf("dark_inspiration resolved to ranks %d divisor %d talent %d, want its node's %d ranks as both and talent %d",
				row.TalentRanks, row.TalentDivisor, row.TalentSpellID, ranks, row.Talent)
		}
		return
	}
	t.Fatal("no dark_inspiration row")
}

func TestTooltipDivisor(t *testing.T) {
	cases := []struct {
		text    string
		spellID int32
		want    int32
	}{
		{"reduce the target's Attack Power by ${$m1/-3*$440667m1}, and", 440667, 3},
		{"reduce the target's Attack Power by ${$m1/-3*$440667m1}, and", 11556, 0},
		{"increase your Attack Power by $m1%", 440667, 0},
	}
	for _, c := range cases {
		if got := tooltipTalentDivisor(c.text, c.spellID); got != c.want {
			t.Errorf("tooltipTalentDivisor(%q, %d) = %d, want %d", c.text, c.spellID, got, c.want)
		}
	}
}

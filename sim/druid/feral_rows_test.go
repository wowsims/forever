package druid

import (
	"testing"

	"github.com/wowsims/forever/sim/core/spelldata"
)

func TestOmenOfClarityUsesThePlaceholderRate(t *testing.T) {
	if got := spellData.OmenOfClarity.Highest().RPPM; got != 2 {
		t.Errorf("Omen of Clarity RPPM = %v, want the placeholder 2", got)
	}
}

// The game files state no attack power share on Ferocious Bite; every rank carries the top rank's measured one,
// so a lower rank registered later does not read 0.
func TestFerociousBiteRanksShareTheAttackPowerPerComboPoint(t *testing.T) {
	top := spellData.FerociousBite.Highest().DamageEffect().APCoeff()
	if top <= 0 {
		t.Fatalf("Ferocious Bite's top rank has no attack power per combo point")
	}
	spellData.FerociousBite.Each(func(rank int32, s *spelldata.Spell) {
		if got := s.DamageEffect().APCoeff(); got != top {
			t.Errorf("Ferocious Bite rank %d (%d) attack power per combo point = %v, want the top rank's %v", rank, s.ID, got, top)
		}
	})
}

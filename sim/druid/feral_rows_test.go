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

func TestFerociousBiteAttackPowerPerComboPoint(t *testing.T) {
	spellData.FerociousBite.Each(func(rank int32, s *spelldata.Spell) {
		if got := s.DamageEffect().APCoeff(); got != 0.037 {
			t.Errorf("Ferocious Bite rank %d (%d) attack power per combo point = %v, want the measured 0.037", rank, s.ID, got)
		}
	})
}

package druid

import "testing"

func TestOmenOfClarityUsesThePlaceholderRate(t *testing.T) {
	if got := spellData.OmenOfClarity.Highest().RPPM; got != 2 {
		t.Errorf("Omen of Clarity RPPM = %v, want the placeholder 2", got)
	}
}

func TestFinisherAttackPowerPlaceholders(t *testing.T) {
	if got := spellData.FerociousBite.ByID(22829).DamageEffect().APCoeff(); got != 0.0375 {
		t.Errorf("Ferocious Bite attack power per combo point = %v, want the placeholder 0.0375", got)
	}
}

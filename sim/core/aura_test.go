package core

import "testing"

// A hit's callbacks run over the auras that were listening when it landed. An extra attack that
// lands inside the first hit's callbacks, and drops an aura there, does not make a later aura run
// twice for either hit.
func TestOnSpellHitDealtRunsEachAuraOncePerHit(t *testing.T) {
	sim := &Simulation{}
	unit := &Unit{auraTracker: newAuraTracker()}
	swing := &Spell{}
	extraAttack := &Spell{}

	sealHits := map[*Spell]int{}
	unit.RegisterAura(Aura{
		Label:    "Extra Attack",
		Duration: NeverExpires,
		OnSpellHitDealt: func(_ *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
			if spell == swing {
				unit.OnSpellHitDealt(sim, extraAttack, result)
			}
		},
	}).Activate(sim)
	charge := unit.RegisterAura(Aura{
		Label:    "Charge",
		Duration: NeverExpires,
		OnSpellHitDealt: func(aura *Aura, sim *Simulation, _ *Spell, _ *SpellResult) {
			aura.Deactivate(sim)
		},
	})
	charge.Activate(sim)
	seal := unit.RegisterAura(Aura{
		Label:    "Seal",
		Duration: NeverExpires,
		OnSpellHitDealt: func(_ *Aura, _ *Simulation, spell *Spell, _ *SpellResult) {
			sealHits[spell]++
		},
	})
	seal.Activate(sim)

	unit.OnSpellHitDealt(sim, swing, &SpellResult{})

	if sealHits[swing] != 1 || sealHits[extraAttack] != 1 {
		t.Errorf("seal ran %d times for the swing and %d for the extra attack, want 1 and 1", sealHits[swing], sealHits[extraAttack])
	}
	if charge.IsActive() {
		t.Errorf("the charge is still up after two hits")
	}
}

// An aura gained inside a hit's callbacks first hears the next hit.
func TestOnSpellHitDealtSkipsAurasGainedDuringTheHit(t *testing.T) {
	sim := &Simulation{}
	unit := &Unit{auraTracker: newAuraTracker()}

	gainedHits := 0
	gained := unit.RegisterAura(Aura{
		Label:    "Gained",
		Duration: NeverExpires,
		OnSpellHitDealt: func(_ *Aura, _ *Simulation, _ *Spell, _ *SpellResult) {
			gainedHits++
		},
	})
	var dropped *Aura
	unit.RegisterAura(Aura{
		Label:    "Gainer",
		Duration: NeverExpires,
		OnSpellHitDealt: func(_ *Aura, sim *Simulation, _ *Spell, _ *SpellResult) {
			dropped.Deactivate(sim)
			gained.Activate(sim)
		},
	}).Activate(sim)
	dropped = unit.RegisterAura(Aura{
		Label:           "Dropped",
		Duration:        NeverExpires,
		OnSpellHitDealt: func(_ *Aura, _ *Simulation, _ *Spell, _ *SpellResult) {},
	})
	dropped.Activate(sim)

	unit.OnSpellHitDealt(sim, &Spell{}, &SpellResult{})
	if gainedHits != 0 {
		t.Errorf("the aura gained during the hit ran %d times for it, want 0", gainedHits)
	}

	unit.OnSpellHitDealt(sim, &Spell{}, &SpellResult{})
	if gainedHits != 1 {
		t.Errorf("the aura gained during the first hit ran %d times for the second, want 1", gainedHits)
	}
}

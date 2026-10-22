//go:build with_db

package balance

import (
	"fmt"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var innervateRow = spelldata.MustFind(29166)

const majorManaPotion = 13444

// A Moonkin with no gear, the default talents, a Major Mana Potion and no rotation, in a fight of the
// given length, reset at the pull.
func newMoonkin(t *testing.T, fight time.Duration) (*core.Simulation, *BalanceDruid) {
	t.Helper()
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Moonkin", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
				TalentsString: DefaultTalents,
				Spec: &proto.Player_BalanceDruid{BalanceDruid: &proto.BalanceDruid{
					Options: &proto.BalanceDruid_Options{ClassOptions: &proto.DruidOptions{}},
				}},
				Equipment:   &proto.EquipmentSpec{},
				Consumables: &proto.ConsumesSpec{ResourcePotId: majorManaPotion},
				Buffs:       &proto.IndividualBuffs{},
				Rotation:    &proto.APLRotation{},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 63}}, Duration: fight.Seconds()},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*BalanceDruid)
}

// The state of autocast's Innervate check `at` into a fight of the given length, once the given share
// of Innervate's gain has gone from the druid's bar, with the Major Mana Potion still ready or drunk
// just then.
type innervateCheck struct {
	cast        bool // autocast cast Innervate
	runsDry     bool // the bar would run dry before the end at the rate it emptied
	potionBack  bool // the potion is ready again before the end
	belowFloor  bool // the bar is under 30%
	anotherFits bool // another Innervate would fit in the fight after this one
}

// The gain is the spirit regen the row raises by its A_MOD_POWER_REGEN_PERCENT effect, for the row's
// duration.
func checkInnervate(t *testing.T, fight, at time.Duration, missingShare float64, potionReady bool) innervateCheck {
	t.Helper()
	sim, moonkin := newMoonkin(t, fight)
	potion := moonkin.GetMajorCooldown(core.ActionID{ItemID: majorManaPotion})
	if potion == nil || !potion.Type.Matches(core.CooldownTypeMana) {
		t.Fatalf("the Major Mana Potion is not a mana major cooldown")
	}
	for sim.CurrentTime < at {
		if sim.Step() {
			t.Fatalf("the fight ended at %v, before %v", sim.CurrentTime, at)
		}
	}
	if !potionReady {
		potion.Spell.SharedCD.Use(sim)
	}
	regen := innervateRow.Effect(dbcenums.A_MOD_POWER_REGEN_PERCENT, int32(dbcenums.POWER_MANA)).BaseValue()
	gain := moonkin.SpiritManaRegenPerSecond() * core.MultiplierFromPercent(regen) * core.DurationFromMillis(innervateRow.DurationMs).Seconds()
	if gain <= 0 || missingShare*gain > moonkin.CurrentMana() {
		t.Fatalf("an Innervate brings %.0f mana to a bar of %.0f; the test needs a gain that fits the bar", gain, moonkin.MaxMana())
	}
	moonkin.SpendMana(sim, missingShare*gain, moonkin.NewManaMetrics(core.ActionID{SpellID: 1}))
	missing := moonkin.MaxMana() - moonkin.CurrentMana()
	end := sim.CurrentTime + sim.GetRemainingDuration()
	check := innervateCheck{
		runsDry:     moonkin.CurrentMana() < missing/sim.CurrentTime.Seconds()*sim.GetRemainingDuration().Seconds(),
		potionBack:  potion.ReadyAt() < end,
		belowFloor:  moonkin.CurrentMana() < 0.3*moonkin.MaxMana(),
		anotherFits: sim.GetRemainingDuration() > core.DurationFromMillis(innervateRow.CooldownMs)+core.DurationFromMillis(innervateRow.DurationMs),
	}
	mcd := moonkin.GetMajorCooldown(core.ActionID{SpellID: innervateRow.ID})
	if mcd == nil {
		t.Fatalf("Innervate is not a major cooldown")
	}
	check.cast = mcd.TryActivate(sim, moonkin.GetCharacter())
	return check
}

// Autocast casts Innervate once half its gain fits under the bar while another Innervate would still
// fit in the fight after this one (the row's cooldown and duration). The last one waits for three
// quarters, for the potion to be drunk, and, while the potion will be back before the end, for the
// bar to drop below 30%. None goes while the bar would last to the end at the rate it has emptied.
func TestInnervateAutocastsAsEarlyAsItsManaFits(t *testing.T) {
	at := 30 * time.Second
	long := at + 405*time.Second  // another Innervate fits after this one
	last := at + 375*time.Second  // only this one fits, and the potion drunk now is back before the end
	short := at + 110*time.Second // only this one fits, and the potion drunk now is not back before the end
	for _, tc := range []struct {
		name        string
		fight, at   time.Duration
		missing     float64 // share of the gain gone from the bar
		potionReady bool
		want        innervateCheck
	}{
		{"last one, full bar", short, at, 0, false, innervateCheck{potionBack: false}},
		{"last one, under three quarters", short, at, 0.70, false, innervateCheck{runsDry: true}},
		{"last one, over three quarters", short, at, 0.80, false, innervateCheck{runsDry: true, cast: true}},
		{"last one, the potion still ready", short, at, 0.95, true, innervateCheck{runsDry: true, potionBack: true, belowFloor: true}},
		{"last one, the potion back before the end, bar over 30%", last, at, 0.80, false, innervateCheck{runsDry: true, potionBack: true}},
		{"last one, the potion back before the end, bar under 30%", last, at, 0.95, false, innervateCheck{runsDry: true, potionBack: true, belowFloor: true, cast: true}},
		{"another fits after it, under half", long, at, 0.45, false, innervateCheck{runsDry: true, potionBack: true, anotherFits: true}},
		{"another fits after it, over half", long, at, 0.55, false, innervateCheck{runsDry: true, potionBack: true, anotherFits: true, cast: true}},
		{"another fits after it, the potion still ready", long, at, 0.55, true, innervateCheck{runsDry: true, potionBack: true, anotherFits: true, cast: true}},
		{"the bar lasts to the end", 180 * time.Second, 170 * time.Second, 0.80, false, innervateCheck{}},
	} {
		t.Run(fmt.Sprintf("%s, %.2f missing", tc.name, tc.missing), func(t *testing.T) {
			got := checkInnervate(t, tc.fight, tc.at, tc.missing, tc.potionReady)
			want := tc.want
			if got.runsDry != want.runsDry || got.potionBack != want.potionBack || got.belowFloor != want.belowFloor || got.anotherFits != want.anotherFits {
				t.Fatalf("setup: %+v, want %+v", got, want)
			}
			if got.cast != want.cast {
				t.Errorf("autocast cast Innervate: %v, want %v", got.cast, want.cast)
			}
		})
	}
}

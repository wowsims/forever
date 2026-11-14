package holy

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func innervatedPaladin(t *testing.T, talents string, individual *proto.IndividualBuffs, party *proto.PartyBuffs) (*core.Simulation, *core.Character) {
	t.Helper()
	player := core.WithSpec(&proto.Player{
		Class:         proto.Class_ClassPaladin,
		Race:          proto.Race_RaceHuman,
		Equipment:     core.GetGearSet("../../../ui/specs/paladin/holy/gear_sets", "default").GearSet,
		Buffs:         individual,
		TalentsString: talents,
		Rotation:      &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}, PlayerOptions)
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, party, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

// Runs the sim's queued actions for d.
func runFor(sim *core.Simulation, d time.Duration) {
	done := false
	sim.AddPendingAction(&core.PendingAction{
		NextActionAt: sim.CurrentTime + d,
		OnAction:     func(_ *core.Simulation) { done = true },
	})
	for !done {
		sim.Step()
	}
}

// Innervate lets all of the spirit regen continue while casting and multiplies all of the regen, MP5
// included, by its row's mana regen factor, whatever share Reverence already lets through: with the
// standard talents' 2/2 Reverence it is the same regen, and Reverence's 30% is back once the aura falls
// off. The mana bar's regen follows each change.
func TestInnervateMultipliesTheFullRegen(t *testing.T) {
	factor := core.MultiplierFromPercent(spelldata.MustFind(29166).Effect(dbcenums.A_MOD_POWER_REGEN_PERCENT, int32(dbcenums.POWER_MANA)).BaseValue())
	for _, c := range []struct {
		talents string
		share   float64
	}{{"", 0}, {StandardTalents, 0.3}} {
		sim, pally := innervatedPaladin(t, c.talents, &proto.IndividualBuffs{Innervates: 1}, &proto.PartyBuffs{})
		if got := pally.PseudoStats.SpiritRegenRateCasting; got != c.share {
			t.Fatalf("share %v: the talents let %v of the spirit regen through while casting", c.share, got)
		}

		metrics := pally.NewManaMetrics(core.ActionID{SpellID: 29166, Tag: 1})
		// Empties the bar inside the five-second rule and returns the mana the next 2 s restore.
		castingRegen := func() float64 {
			pally.SpendMana(sim, pally.CurrentMana(), metrics)
			pally.PseudoStats.FiveSecondRuleRefreshTime = sim.CurrentTime + time.Second*5
			runFor(sim, 2*time.Second)
			return pally.CurrentMana()
		}

		if pally.MP5ManaRegenPerSecond() == 0 {
			t.Fatal("the paladin has no MP5")
		}
		idle, casting := pally.ManaRegenPerSecondWhileNotCasting(), pally.ManaRegenPerSecondWhileCasting()
		want := factor * idle

		aura := pally.GetAura("Innervates (External)")
		if aura == nil {
			t.Fatal("the paladin has no Innervates aura")
		}
		aura.Activate(sim)
		if got := pally.ManaRegenPerSecondWhileNotCasting(); math.Abs(got-want) > 1e-9 {
			t.Errorf("share %v: %v mana per second while not casting under Innervate, want %v", c.share, got, want)
		}
		if got, idle := pally.ManaRegenPerSecondWhileCasting(), pally.ManaRegenPerSecondWhileNotCasting(); got != idle {
			t.Errorf("share %v: %v mana per second while casting under Innervate, want the %v while not casting",
				c.share, got, idle)
		}
		if got := castingRegen(); math.Abs(got-2*want) > 1e-9 {
			t.Errorf("share %v: 2 s of casting under Innervate restore %v mana, want %v", c.share, got, 2*want)
		}

		aura.Deactivate(sim)
		if got := pally.ManaRegenPerSecondWhileNotCasting(); got != idle {
			t.Errorf("share %v: %v mana per second while not casting after Innervate, want %v", c.share, got, idle)
		}
		if got := pally.ManaRegenPerSecondWhileCasting(); math.Abs(got-casting) > 1e-9 {
			t.Errorf("share %v: %v mana per second while casting after Innervate, want %v", c.share, got, casting)
		}
		if got := castingRegen(); math.Abs(got-2*casting) > 1e-9 {
			t.Errorf("share %v: 2 s of casting after Innervate restore %v mana, want %v", c.share, got, 2*casting)
		}
	}
}

// Innervate multiplies the Spirit regen and the MP5 from gear, but a periodic energize buff ticks its
// own mana under it: over the whole Innervate the paladin gains the multiplied Spirit and gear regen
// plus the buff's mana as it ticks without Innervate. Any other mana regen multiplier, such as Skyborne
// Energized's, scales the buff's mana with the rest.
func TestInnervateLeavesEnergizeBuffsUnscaled(t *testing.T) {
	factor := core.MultiplierFromPercent(spelldata.MustFind(29166).Effect(dbcenums.A_MOD_POWER_REGEN_PERCENT, int32(dbcenums.POWER_MANA)).BaseValue())
	for _, c := range []struct {
		name       string
		spellID    int32
		individual *proto.IndividualBuffs
		party      *proto.PartyBuffs
	}{
		{"Greater Blessing of Wisdom", 25918, &proto.IndividualBuffs{Innervates: 1, GreaterBlessingOfWisdom: 2}, &proto.PartyBuffs{}},
		{"Mana Spring Totem", 10494, &proto.IndividualBuffs{Innervates: 1}, &proto.PartyBuffs{ManaSpringTotem: proto.TristateEffect_TristateEffectRegular}},
	} {
		t.Run(c.name, func(t *testing.T) {
			sim, pally := innervatedPaladin(t, "", c.individual, c.party)
			if pally.GetAuraByID(core.ActionID{SpellID: c.spellID, Tag: -1}) == nil {
				t.Fatalf("the paladin has no %s", c.name)
			}
			energize := spelldata.MustFind(c.spellID).Effect(dbcenums.A_PERIODIC_ENERGIZE, int32(dbcenums.POWER_MANA))
			if energize.Average() == 0 {
				t.Fatalf("%s states no mana over time", c.name)
			}
			buffPerSecond := energize.Average() * 1000 / float64(energize.PeriodMs)
			gearPerSecond := pally.GetStat(stats.MP5)/5 - buffPerSecond
			if gearPerSecond <= 0 {
				t.Fatal("the paladin has no MP5 besides the buff")
			}

			idle := pally.ManaRegenPerSecondWhileNotCasting()
			pally.MultiplyManaRegenSpeed(sim, 2)
			if got := pally.ManaRegenPerSecondWhileNotCasting(); got != 2*idle {
				t.Errorf("a mana regen speed of 2 gives %v mana per second while not casting, want %v", got, 2*idle)
			}
			pally.MultiplyManaRegenSpeed(sim, 0.5)

			aura := pally.GetAura("Innervates (External)")
			if aura == nil {
				t.Fatal("the paladin has no Innervates aura")
			}
			aura.Activate(sim)
			metrics := pally.NewManaMetrics(core.ActionID{SpellID: 29166, Tag: 2})
			pally.SpendMana(sim, pally.CurrentMana(), metrics)
			runFor(sim, aura.Duration)

			spirit := pally.SpiritManaRegenPerSecond() * pally.PseudoStats.SpiritRegenMultiplier
			want := (factor*(spirit+gearPerSecond) + buffPerSecond) * aura.Duration.Seconds()
			if got := pally.CurrentMana(); math.Abs(got-want) > 1e-6 {
				t.Errorf("Innervate restores %v mana, want %v with the %v from %s unscaled", got, want,
					buffPerSecond*aura.Duration.Seconds(), c.name)
			}
		})
	}
}

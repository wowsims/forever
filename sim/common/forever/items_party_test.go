//go:build with_db

package forever_test

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	_ "github.com/wowsims/forever/sim/mage" // registers Atiesh, Greatstaff of the Guardian.
)

// The warrior stand-in of class class wearing one item, in a party with these party buffs.
func partyItemWearer(t *testing.T, seed int64, class proto.Class, slot proto.ItemSlot, itemID int32, party *proto.PartyBuffs) (*core.Simulation, *core.Character) {
	t.Helper()
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: seed},
		Raid:       partyItemRaid(class, slot, itemID, party),
		Encounter:  &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

// The character sheet's stats for partyItemWearer's stand-in.
func partyItemSheet(t *testing.T, class proto.Class, slot proto.ItemSlot, itemID int32, party *proto.PartyBuffs) *proto.UnitStats {
	t.Helper()
	result := core.ComputeStats(&proto.ComputeStatsRequest{Raid: partyItemRaid(class, slot, itemID, party), SkipRotation: true})
	if result.ErrorResult != "" {
		t.Fatalf("ComputeStats: %s", result.ErrorResult)
	}
	return result.RaidStats.Parties[0].Players[0].FinalStats
}

func partyItemRaid(class proto.Class, slot proto.ItemSlot, itemID int32, party *proto.PartyBuffs) *proto.Raid {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[slot].Id = itemID
	player := &proto.Player{
		Name: "Wearer", Class: class, Race: proto.Race_RaceTroll,
		Equipment:   &proto.EquipmentSpec{Items: items},
		Consumables: &proto.ConsumesSpec{},
		Buffs:       &proto.IndividualBuffs{},
		Rotation:    &proto.APLRotation{},
	}
	core.WithSpec(player, &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{}})
	return &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: party}}}
}

// Darkspear 12802's 1314143 grants 16939's crit, physical and spell, for 16939's duration, on the
// chance of melee hits 1314143 states and with its time between procs. Over 300 wielders the first
// proc takes one over the chance hits on average.
func TestDarkspearGrantsItsPartyCritOnItsChance(t *testing.T) {
	trigger, crit := spelldata.MustFind(1314143), spelldata.MustFind(16939)
	chance := trigger.StatedChance()
	if chance <= 0 || chance >= 1 {
		t.Fatalf("1314143 states a %v chance, want a roll", chance)
	}
	hitsToProc := func(sim *core.Simulation, character *core.Character, buff *core.Aura) int {
		t.Helper()
		limit := int(50 / chance)
		hits := 0
		for ; !buff.IsActive(); hits++ {
			if hits == limit {
				t.Fatalf("%d melee hits granted no Darkspear crit", limit)
			}
			dealHit(t, sim, character, meleeSwing(&character.Unit))
		}
		return hits
	}

	total := 0
	const wielders = 300
	for seed := range int64(wielders) {
		sim, character := partyItemWearer(t, seed+1, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotMainHand, 12802, &proto.PartyBuffs{})
		buff := character.GetAura("Darkspear (Player)")
		if buff == nil {
			t.Fatal("12802 registers no Darkspear crit")
		}
		total += hitsToProc(sim, character, buff)
	}
	// The hits to the first proc are geometric: mean 1/p, spread sqrt(1-p)/p, over the wielders.
	want := 1 / chance
	band := 4 * math.Sqrt(1-chance) / chance / math.Sqrt(wielders)
	if mean := float64(total) / wielders; math.Abs(mean-want) > band {
		t.Errorf("the first proc took %v hits on average, want %v within %v", mean, want, band)
	}

	sim, character := partyItemWearer(t, 1, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotMainHand, 12802, &proto.PartyBuffs{})
	buff := character.GetAura("Darkspear (Player)")
	before := character.GetStats()
	hitsToProc(sim, character, buff)
	raised := sim.CurrentTime
	for _, stat := range []stats.Stat{stats.PhysicalCritPercent, stats.SpellCritPercent} {
		if got, want := character.GetStat(stat)-before[stat], crit.EffectN(1).Average(); got != want {
			t.Errorf("the buff adds %v %v, want %v", got, stat, want)
		}
	}
	if buff.Duration != crit.Duration() {
		t.Errorf("the buff lasts %v, want %v", buff.Duration, crit.Duration())
	}

	stepTo(t, sim, raised+crit.Duration()+time.Millisecond)
	for range int(10 / chance) {
		dealHit(t, sim, character, meleeSwing(&character.Unit))
	}
	if buff.IsActive() {
		t.Errorf("the buff came back within the %v between procs", trigger.ICD())
	}
	stepTo(t, sim, raised+trigger.ICD())
	hitsToProc(sim, character, buff)
}

// The party's Darkspear and the wielder's own bid in one category: with both up, the wielder has
// 16939's crit once, not twice.
func TestDarkspearAndThePartysDarkspearDoNotStack(t *testing.T) {
	sim, character := partyItemWearer(t, 1, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotMainHand, 12802, &proto.PartyBuffs{Darkspear: true})
	own, party := character.GetAura("Darkspear (Player)"), character.GetAura("Darkspear (External)")
	if own == nil || party == nil {
		t.Fatalf("the wielder carries %v and %v, want both copies", own, party)
	}
	own.Deactivate(sim)
	party.Deactivate(sim)
	before := character.GetStat(stats.PhysicalCritPercent)

	party.Activate(sim)
	own.Activate(sim)
	if got, want := character.GetStat(stats.PhysicalCritPercent)-before, spelldata.MustFind(16939).EffectN(1).Average(); got != want {
		t.Errorf("both copies up add %v crit, want %v", got, want)
	}
}

// Admiral's Hat 10030 uses 12022 on the cooldown its on-use effect states: 12022's Stamina for
// 12022's duration. Used beside the party's Admiral's Hat it adds nothing more.
func TestAdmiralsHatGrantsItsStaminaOnce(t *testing.T) {
	const hat = 10030
	buff := spelldata.MustFind(12022)
	stamina := buff.EffectN(1).Average()
	cooldown, category, categoryCooldown := classItemOnUse(t, hat)

	for _, tc := range []struct {
		name  string
		party *proto.PartyBuffs
		want  float64
	}{
		{"alone", &proto.PartyBuffs{}, stamina},
		{"beside the party's", &proto.PartyBuffs{AdmiralsHat: true}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim, character := partyItemWearer(t, 1, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotHead, hat, tc.party)
			before := character.GetStats()
			useTrinket(t, sim, character, hat, core.CooldownTypeSurvival, cooldown, category, categoryCooldown)
			if got := gained(character, before, stats.Stamina); got != tc.want {
				t.Errorf("the hat adds %v Stamina, want %v", got, tc.want)
			}
			if got := character.GetAura("Admiral's Hat (Player)").Duration; got != buff.Duration() {
				t.Errorf("the hat's buff lasts %v, want %v", got, buff.Duration())
			}
		})
	}

	_, bare := partyItemWearer(t, 1, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotHead, 0, &proto.PartyBuffs{})
	_, buffed := partyItemWearer(t, 1, proto.Class_ClassWarrior, proto.ItemSlot_ItemSlotHead, 0, &proto.PartyBuffs{AdmiralsHat: true})
	if got := buffed.GetStat(stats.Stamina) - bare.GetStat(stats.Stamina); got != stamina {
		t.Errorf("the party's Admiral's Hat adds %v Stamina, want %v", got, stamina)
	}
}

// Atiesh, Greatstaff of the Guardian 22589's party aura 28142 gives the mage wearing it 28142's
// spell crit, and the party buff's count adds the other members' staves on top.
func TestAtieshGreatstaffOfTheGuardianGivesItsWearerItsSpellCrit(t *testing.T) {
	_, bare := partyItemWearer(t, 1, proto.Class_ClassMage, proto.ItemSlot_ItemSlotMainHand, 0, &proto.PartyBuffs{})
	_, wearer := partyItemWearer(t, 1, proto.Class_ClassMage, proto.ItemSlot_ItemSlotMainHand, 22589, &proto.PartyBuffs{})
	_, counted := partyItemWearer(t, 1, proto.Class_ClassMage, proto.ItemSlot_ItemSlotMainHand, 22589, &proto.PartyBuffs{AtieshMage: 1})
	staff := spelldata.MustFind(28142).EffectN(1).Average()

	spellCrit := func(character *core.Character) float64 { return character.GetStat(stats.SpellCritPercent) }
	if got := spellCrit(wearer) - spellCrit(bare); math.Abs(got-staff) > 1e-9 {
		t.Errorf("the staff adds %v spell crit, want %v", got, staff)
	}
	if got := spellCrit(counted) - spellCrit(bare); math.Abs(got-2*staff) > 1e-9 {
		t.Errorf("the staff and one more in the party add %v spell crit, want %v", got, 2*staff)
	}
	if got := wearer.GetStat(stats.PhysicalCritPercent) - bare.GetStat(stats.PhysicalCritPercent); got != 0 {
		t.Errorf("the staff adds %v physical crit, want none", got)
	}

	bareSheet := partyItemSheet(t, proto.Class_ClassMage, proto.ItemSlot_ItemSlotMainHand, 0, &proto.PartyBuffs{})
	wearerSheet := partyItemSheet(t, proto.Class_ClassMage, proto.ItemSlot_ItemSlotMainHand, 22589, &proto.PartyBuffs{})
	sheetCrit := func(sheet *proto.UnitStats) float64 {
		return stats.PseudoStatValue(sheet.PseudoStats, proto.PseudoStat_PseudoStatSpellCritPercent)
	}
	if got := sheetCrit(wearerSheet) - sheetCrit(bareSheet); math.Abs(got-staff) > 1e-9 {
		t.Errorf("the character sheet shows the staff's spell crit as %v, want %v", got, staff)
	}
}

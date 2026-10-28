package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestRaidRandomPlayerReachesEveryPartyAndExcludesTheCurrentTarget(t *testing.T) {
	request := fakeRageSimRequest()
	request.Raid.TargetDummies = 9
	request.Raid.NumActiveParties = 2
	request.Raid.Parties = append(request.Raid.Parties, &proto.Party{})
	sim := setupFakeRageSimFrom(request)
	exclude := &sim.Raid.Parties[0].Players[0].GetCharacter().Unit
	seen := make(map[*Unit]bool)
	for range 200 {
		unit := sim.Raid.GetRandomPlayer(sim, "Raid player test", exclude)
		if unit == nil || unit == exclude {
			t.Fatal("random player was missing or the excluded unit")
		}
		seen[unit] = true
	}
	if len(seen) != len(sim.Raid.AllPlayerUnits)-1 {
		t.Errorf("visited %d raid members, want %d", len(seen), len(sim.Raid.AllPlayerUnits)-1)
	}
	if !seen[&sim.Raid.Parties[1].Players[0].GetCharacter().Unit] {
		t.Fatal("never picked a member of the second party")
	}
	lone := &Raid{AllPlayerUnits: []*Unit{exclude}}
	if lone.GetRandomPlayer(sim, "Raid player test", exclude) != nil {
		t.Fatal("a raid holding only the excluded player returned that player")
	}
}

func TestPlayerAfterAnEmptyPartySlotGetsItsOwnConfig(t *testing.T) {
	request := fakeRageSimRequest()
	party := request.Raid.Parties[0]
	party.Players[0].Buffs = &proto.IndividualBuffs{GreaterBlessingOfKings: true}
	party.Players = append([]*proto.Player{{}}, party.Players...)
	sim := setupFakeRageSimFrom(request)
	char := sim.Raid.Parties[0].Players[0].GetCharacter()
	if char.PartyIndex != 1 {
		t.Fatalf("player sits at party slot %d, want 1", char.PartyIndex)
	}
	if char.GetAura("Greater Blessing of Kings (External)") == nil {
		t.Error("the player after an empty slot did not get the buffs from its own config")
	}
}

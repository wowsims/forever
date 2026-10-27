package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestPartyRandomPlayerStaysInPartyAndExcludesTheCurrentTarget(t *testing.T) {
	request := fakeRageSimRequest()
	request.Raid.TargetDummies = 9
	request.Raid.NumActiveParties = 2
	request.Raid.Parties = append(request.Raid.Parties, &proto.Party{})
	sim := setupFakeRageSimFrom(request)
	party := sim.Raid.Parties[0]
	exclude := &party.Players[0].GetCharacter().Unit
	seen := make(map[*Unit]bool)
	for range 40 {
		unit := party.GetRandomPlayer(sim, "Party player test", exclude)
		if unit == exclude || !party.HasPlayer(unit) {
			t.Fatal("random player was the excluded unit or outside the party")
		}
		seen[unit] = true
	}
	if len(seen) != party.Size()-1 {
		t.Errorf("visited %d party members, want %d", len(seen), party.Size()-1)
	}
	if party.HasPlayer(&sim.Raid.Parties[1].Players[0].GetCharacter().Unit) {
		t.Fatal("a player from the other party matched this party")
	}
	empty := &Party{}
	if empty.GetRandomPlayer(sim, "Party player test", nil) != nil {
		t.Fatal("an empty party returned a player")
	}
	lone := &Party{Players: party.Players[:1]}
	if lone.GetRandomPlayer(sim, "Party player test", exclude) != nil {
		t.Fatal("a party containing only the excluded player returned that player")
	}
}

package gen

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/siml"
)

func spellID(id, rank int32) *proto.ActionID {
	return &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}, Rank: rank}
}

// A build's talents go into spell, uncastable, with no rank and marked talent only, and a talent that is also a
// castable spell keeps the spell's rank, stays castable and is not marked.
func TestRegistrationsTakeTheTalents(t *testing.T) {
	result := &proto.ComputeStatsResult{
		RaidStats: &proto.RaidStats{Parties: []*proto.PartyStats{{Players: []*proto.PlayerStats{{Metadata: &proto.UnitMetadata{
			Spells:  []*proto.SpellStats{{Id: spellID(20375, 1), IsCastable: true}},
			Talents: []*proto.ActionID{spellID(1310905, 2), spellID(20375, 1)},
		}}}}}},
		EncounterStats: &proto.EncounterStats{},
	}
	spells := registrations(result)[siml.NamespaceSpell]
	if got, ok := spells[Identity{SpellID: 1310905}]; !ok || got.Rank != 0 || got.Castable || !got.TalentOnly {
		t.Errorf("Consecrated Ground registers as %+v (listed %v), want no rank, uncastable and talent only", got, ok)
	}
	if got := spells[Identity{SpellID: 20375}]; got.Rank != 1 || !got.Castable || got.TalentOnly {
		t.Errorf("Seal of Command registers as %+v, want rank 1, castable and listed as a spell", got)
	}
}

package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestIndividualSimRequestWithoutPlayer(t *testing.T) {
	request, _ := IndividualSimRequest(&proto.IndividualSimSettings{}, nil)
	if player := request.Raid.Parties[0].Players[0]; player == nil {
		t.Fatalf("settings without a player gave no player in the raid")
	}
}

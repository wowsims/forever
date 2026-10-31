package bulk

import (
	"fmt"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func BulkCombinationCount(request *proto.BulkCombinationCountRequest) *proto.BulkCombinationCountResult {
	if request == nil {
		return &proto.BulkCombinationCountResult{Error: &proto.ErrorOutcome{Message: "bulk combination count request is missing"}}
	}
	if request.GetBaseRequest() == nil {
		return &proto.BulkCombinationCountResult{Error: &proto.ErrorOutcome{Message: "bulk combination count request is missing base request"}}
	}
	if request.GetBulkSettings() == nil {
		return &proto.BulkCombinationCountResult{Error: &proto.ErrorOutcome{Message: "bulk combination count request is missing bulk settings"}}
	}

	bulkRequest := &proto.BulkSimRequest{
		BaseRequest:  request.GetBaseRequest(),
		BulkSettings: request.GetBulkSettings(),
	}
	generator, err := newGeneratorFromRequest(bulkRequest)
	if err != nil {
		return &proto.BulkCombinationCountResult{Error: &proto.ErrorOutcome{Message: err.Error()}}
	}

	rawCombinations := generator.rawCombinationsCount()
	matchingCombinations := rawCombinations

	return &proto.BulkCombinationCountResult{
		RawCombinations: int32(rawCombinations),
		Combinations:    int32(matchingCombinations),
	}
}

func BulkCandidates(request *proto.BulkCandidatesRequest) *proto.BulkCandidatesResult {
	if request == nil {
		return &proto.BulkCandidatesResult{Error: &proto.ErrorOutcome{Message: "bulk candidates request is missing"}}
	}
	if request.GetBaseRequest() == nil {
		return &proto.BulkCandidatesResult{Error: &proto.ErrorOutcome{Message: "bulk candidates request is missing base request"}}
	}
	if request.GetBulkSettings() == nil {
		return &proto.BulkCandidatesResult{Error: &proto.ErrorOutcome{Message: "bulk candidates request is missing bulk settings"}}
	}

	bulkRequest := &proto.BulkSimRequest{
		BaseRequest:  request.GetBaseRequest(),
		BulkSettings: request.GetBulkSettings(),
	}
	generator, err := newGeneratorFromRequest(bulkRequest)
	if err != nil {
		return &proto.BulkCandidatesResult{Error: &proto.ErrorOutcome{Message: err.Error()}}
	}

	rawCombinations := generator.rawCombinationsCount()
	candidates := generator.buildCandidates()

	return &proto.BulkCandidatesResult{
		Candidates:      candidates,
		RawCombinations: int32(rawCombinations),
		Combinations:    int32(len(candidates)),
	}
}

// Validates the shared preconditions of every bulk entry point (player present, equipment
// present), registers the request's database items, and builds the candidate generator. Callers
// map the returned error into their own result type.
func newGeneratorFromRequest(request *proto.BulkSimRequest) (*bulkSimCandidateGenerator, error) {
	player, playerErr := getPlayer(request)
	if playerErr != nil {
		return nil, playerErr
	}
	if player.GetEquipment() == nil {
		return nil, fmt.Errorf("bulk sim request is missing player equipment")
	}
	if player.GetDatabase() != nil {
		// Safe to run unserialized: core guards the shared database itself, so generation
		// (minutes, for a large selection) does not block other bulk requests.
		core.AddToDatabase(player.GetDatabase())
	}
	return newBulkSimCandidateGenerator(request, player)
}

func EnsureBulkSimCandidatesGenerated(request *proto.BulkSimRequest) error {
	if request == nil || request.GetBulkSettings() == nil || len(request.GetCandidates()) > 0 {
		return nil
	}
	if request.GetBaseRequest() == nil || request.GetBaseRequest().GetRaid() == nil {
		return fmt.Errorf("bulk sim request is missing base raid")
	}
	generator, buildErr := newGeneratorFromRequest(request)
	if buildErr != nil {
		return buildErr
	}
	request.Candidates = generator.buildCandidates()
	return nil
}

func getPlayer(request *proto.BulkSimRequest) (*proto.Player, error) {
	if request == nil || request.GetBaseRequest() == nil || request.GetBaseRequest().GetRaid() == nil {
		return nil, fmt.Errorf("bulk sim request is missing base raid")
	}
	parties := request.GetBaseRequest().GetRaid().GetParties()
	if len(parties) == 0 || parties[0] == nil {
		return nil, fmt.Errorf("bulk sim request raid is missing parties")
	}
	players := parties[0].GetPlayers()
	if len(players) == 0 || players[0] == nil {
		return nil, fmt.Errorf("bulk sim request raid is missing player")
	}
	return players[0], nil
}

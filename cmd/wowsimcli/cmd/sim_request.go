package cmd

import (
	"fmt"
	"math/rand/v2"
	"os"
	"strings"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	defaultIterations = 12500     // The site's, for settings that set none.
	maxRandomSeed     = 1<<32 - 1 // The site draws a seed below this when the settings fix none.
)

// The request in the file at path: a RaidSimRequest in protojson, or an x.siml, a full setup or a preset build loaded
// over its spec's defaults.siml, simmed as the site sims it. The warnings say what a setup leaves for the request to
// fill in, and a rotation the site builds that the sim cannot.
func readSimRequest(path string) (*proto.RaidSimRequest, []string, error) {
	if !strings.HasSuffix(path, ".siml") {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, err
		}
		request := &proto.RaidSimRequest{}
		if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(data, request); err != nil {
			return nil, nil, fmt.Errorf("%s: %w", path, err)
		}
		if request.SimOptions == nil {
			return nil, nil, fmt.Errorf("%s: a RaidSimRequest needs simOptions", path)
		}
		return request, nil, nil
	}
	c, err := readSetup(".", path)
	if err != nil {
		return nil, nil, err
	}
	player, warnings := c.settings.GetPlayer(), c.info.Warnings
	// The site builds an auto or simple rotation into an APL in the browser, and the sim runs only APL rows. A player
	// with no rotation at all stops the sim at its first swing.
	switch rotation := player.Rotation; {
	case rotation == nil:
		player.Rotation = &proto.APLRotation{}
		warnings = append(warnings, "the settings hold no rotation; simmed with none")
	case rotation.Type != proto.APLRotation_TypeAPL && len(rotation.PrepullActions)+len(rotation.PriorityList) == 0:
		warnings = append(warnings, fmt.Sprintf("the rotation is %s, which the site builds into an APL before simming; the file holds no APL row, so this sims no rotation", rotation.Type))
	case rotation.Type != proto.APLRotation_TypeAPL:
		warnings = append(warnings, fmt.Sprintf("the rotation is %s, which the site builds into an APL before simming; this sims the APL rows the file holds instead", rotation.Type))
	}
	request, more := settingsRequest(c.settings)
	return request, append(warnings, more...), nil
}

// The request the site's individual sim sends for settings (core.IndividualSimRequest), for the iterations and fixed
// seed of the sim settings, or the site's 12500 iterations and a random seed.
func settingsRequest(s *proto.IndividualSimSettings) (*proto.RaidSimRequest, []string) {
	options := &proto.SimOptions{Iterations: s.GetSettings().GetIterations(), RandomSeed: s.GetSettings().GetFixedRngSeed()}
	if options.Iterations == 0 {
		options.Iterations = defaultIterations
	}
	if options.RandomSeed == 0 {
		options.RandomSeed = rand.Int64N(maxRandomSeed)
	}
	return core.IndividualSimRequest(s, options)
}

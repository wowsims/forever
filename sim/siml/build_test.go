package siml

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

// A build's encounter keeps the defaults' healing model where the build sets none, as the site's applyBuild does, and
// replaces it where the build sets one.
func TestLoadBuildKeepsTheHealingModelAnEncounterLeavesOut(t *testing.T) {
	defaults := &proto.HealingModel{Hps: 1200}
	s := &proto.IndividualSimSettings{Player: &proto.Player{HealingModel: defaults}}
	LoadBuild(s, &proto.IndividualSimSettings{Player: &proto.Player{}, Encounter: &proto.Encounter{Duration: 120}})
	if s.Player.HealingModel != defaults {
		t.Errorf("healing model %v under a build with no healing model, want the defaults' %v", s.Player.HealingModel, defaults)
	}

	own := &proto.HealingModel{Hps: 900}
	LoadBuild(s, &proto.IndividualSimSettings{Player: &proto.Player{HealingModel: own}, Encounter: &proto.Encounter{Duration: 120}})
	if s.Player.HealingModel != own {
		t.Errorf("healing model %v, want the build's own %v", s.Player.HealingModel, own)
	}
}

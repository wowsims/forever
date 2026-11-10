package cmd

import (
	"fmt"
	"math/rand/v2"
	"os"
	"strings"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
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
	c, err := readSetup(repositoryRootOf(path), path)
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

// The request the site's individual sim sends for settings (individualSimRequest), for the iterations and fixed seed
// of the sim settings, or the site's 12500 iterations and a random seed.
func settingsRequest(s *proto.IndividualSimSettings) (*proto.RaidSimRequest, []string) {
	options := &proto.SimOptions{Iterations: s.GetSettings().GetIterations(), RandomSeed: s.GetSettings().GetFixedRngSeed()}
	if options.Iterations == 0 {
		options.Iterations = defaultIterations
	}
	if options.RandomSeed == 0 {
		options.RandomSeed = rand.Int64N(maxRandomSeed)
	}
	return individualSimRequest(s, options)
}

// The request the site's individual sim sends for settings, with options: the player alone in the first of the raid's
// five parties, with the party buffs, the raid buffs, debuffs, tanks and target dummies, against the encounter. A ring enchant needs
// Enchanting, so a player without it is simmed without its ring enchants. Settings with no encounter, or one with no
// target, are simmed against one default target, in the site's default fight, which the warnings say. s is not
// changed. The potions and conjured items the site adds for simming (getConsumes(true)) are not added: which are
// relevant comes from the spec configs under ui/specs.
func individualSimRequest(s *proto.IndividualSimSettings, options *proto.SimOptions) (*proto.RaidSimRequest, []string) {
	var warnings []string
	player := googleProto.Clone(orEmpty(s.GetPlayer())).(*proto.Player)
	if player.Profession1 != proto.Profession_Enchanting && player.Profession2 != proto.Profession_Enchanting {
		for _, slot := range []proto.ItemSlot{proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2} {
			if items := player.GetEquipment().GetItems(); int(slot) < len(items) && items[slot] != nil {
				items[slot].Enchant = 0
			}
		}
	}
	raid := core.SinglePlayerRaidProto(player, orEmpty(s.PartyBuffs), orEmpty(s.RaidBuffs), orEmpty(s.Debuffs))
	// The target dummies take the free places after the player, up to 24.
	for len(raid.Parties) < 5 {
		raid.Parties = append(raid.Parties, &proto.Party{})
	}
	raid.Tanks = s.Tanks
	raid.TargetDummies = s.TargetDummies

	encounter := s.Encounter
	switch {
	case encounter == nil:
		encounter = siteDefaultEncounter()
		warnings = append(warnings, "the settings hold no encounter; simmed against one default target")
	case len(encounter.Targets) == 0:
		encounter = googleProto.Clone(encounter).(*proto.Encounter)
		encounter.Targets = []*proto.Target{siteDefaultTarget()}
		warnings = append(warnings, "the encounter has no target; simmed against one default target")
	}
	return &proto.RaidSimRequest{Raid: raid, Encounter: encounter, SimOptions: options, Type: proto.SimType_SimTypeIndividual}, warnings
}

// The fight the site starts with (Encounter.defaultEncounterProto in ui/sim/raid/encounter.ts): 180 s varied by 5 s,
// against one default target. Each call returns a new message.
func siteDefaultEncounter() *proto.Encounter {
	return &proto.Encounter{
		Duration:             180,
		DurationVariation:    5,
		ExecuteProportion_20: 0.2,
		ExecuteProportion_35: 0.35,
		Targets:              []*proto.Target{siteDefaultTarget()},
	}
}

// The site's default target (Encounter.defaultTargetProto): the sim tests' default target with the raid target's id,
// name and health, and the block value the site gives it. Each call returns a new message.
func siteDefaultTarget() *proto.Target {
	target := core.FreshDefaultTargetConfig()
	target.Id, target.Name = 31146, "Raid Target"
	target.Stats[stats.Health] = 1428900
	target.Stats[stats.BlockValue] = 54
	return target
}

// m, or an empty message of its type when it is nil.
func orEmpty[M googleProto.Message](m M) M {
	if m.ProtoReflect().IsValid() {
		return m
	}
	return m.ProtoReflect().Type().New().Interface().(M)
}

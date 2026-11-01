package core

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/siml"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var DefaultSimTestOptions = &proto.SimOptions{
	Iterations: 20,
	IsTest:     true,
	Debug:      false,
	RandomSeed: 101,
}
var StatWeightsDefaultSimTestOptions = &proto.SimOptions{
	Iterations: 300,
	IsTest:     true,
	Debug:      false,
	RandomSeed: 101,
}
var AverageDefaultSimTestOptions = &proto.SimOptions{
	Iterations: 2000,
	IsTest:     true,
	Debug:      false,
	RandomSeed: 101,
}

const ShortDuration = 60
const LongDuration = 180

func FreshDefaultTargetConfig() *proto.Target {
	return &proto.Target{
		Level: CharacterLevel + 3,
		Stats: stats.Stats{
			stats.Armor:       3731,
			stats.AttackPower: 320,
		}.ToProtoArray(),
		MobType: proto.MobType_MobTypeMechanical,

		SwingSpeed:    2,
		MinBaseDamage: 2601,
		ParryHaste:    true,
		CanCrush:      true,
		DamageSpread:  0.5,
	}
}

var DefaultTargetProto = FreshDefaultTargetConfig()

var FullRaidBuffs = &proto.RaidBuffs{
	ArcaneBrilliance:         1,
	PrayerOfFortitude:        1,
	PrayerOfSpirit:           true,
	GiftOfTheWild:            1,
	PrayerOfShadowProtection: 1,

	GraceOfAirTotem:      1,
	StrengthOfEarthTotem: 1,
	WindfuryTotem:        true,
	BattleShout:          1, // Rank 6
	TrueshotAura:         true,
	LeaderOfThePack:      true,
	MoonkinAura:          true,
}

var FullPartyBuffs = &proto.PartyBuffs{
	BloodPact:         true,
	DevotionAura:      true,
	RetributionAura:   true,
	ConcentrationAura: true,
	AtieshDruid:       1,
	AtieshMage:        1,
	AtieshPriest:      1,
	AtieshWarlock:     1,

	ManaSpringTotem: proto.TristateEffect_TristateEffectImproved,
	ManaTideTotems:  1,
}

var FullIndividualBuffs = &proto.IndividualBuffs{
	GreaterBlessingOfKings:  true,
	GreaterBlessingOfWisdom: 1,
	GreaterBlessingOfMight:  1,
	Thorns:                  true,
}

var FullTankIndividualBuffs = &proto.IndividualBuffs{
	GreaterBlessingOfKings:  true,
	GreaterBlessingOfWisdom: 1,
	GreaterBlessingOfMight:  1,
	Thorns:                  true,
}

var FullDebuffs = &proto.Debuffs{
	JudgementOfWisdom:      true,
	JudgementOfLight:       true,
	JudgementOfTheCrusader: true,
	CurseOfElements:        true,
	GiftOfArthas:           true,
	ExposeArmor:            true,
	FaerieFire:             true,
	SunderArmor:            true,
	CurseOfRecklessness:    true,
	HuntersMark:            true,
	DemoralizingRoar:       true,
	// DemoralizingShout:         true,
	ThunderClap:  true,
	InsectSwarm:  true,
	ScorpidSting: true,
}

func NewDefaultTarget() *proto.Target {
	return DefaultTargetProto // seems to be read-only
}

func MakeDefaultEncounterCombos() []EncounterCombo {
	var DefaultTarget = NewDefaultTarget()

	multipleTargets := make([]*proto.Target, 21)
	for i := range multipleTargets {
		if i != 10 {
			multipleTargets[i] = DefaultTarget
		} else {
			disabledTarget := FreshDefaultTargetConfig()
			disabledTarget.DisabledAtStart = true
			multipleTargets[i] = disabledTarget
		}
	}

	return []EncounterCombo{
		{
			Label: "ShortSingleTarget",
			Encounter: &proto.Encounter{
				Duration:             ShortDuration,
				ExecuteProportion_20: 0.2,
				ExecuteProportion_35: 0.35,
				Targets: []*proto.Target{
					DefaultTarget,
				},
			},
		},
		{
			Label: "LongSingleTarget",
			Encounter: &proto.Encounter{
				Duration:             LongDuration,
				ExecuteProportion_20: 0.2,
				ExecuteProportion_35: 0.35,
				Targets: []*proto.Target{
					DefaultTarget,
				},
			},
		},
		{
			Label: "LongMultiTarget",
			Encounter: &proto.Encounter{
				Duration:             LongDuration,
				ExecuteProportion_20: 0.2,
				ExecuteProportion_35: 0.35,
				Targets:              multipleTargets,
			},
		},
	}
}

func MakeSingleTargetEncounter(variation float64) *proto.Encounter {
	return &proto.Encounter{
		Duration:             LongDuration,
		DurationVariation:    variation,
		ExecuteProportion_20: 0.2,
		ExecuteProportion_35: 0.35,
		Targets: []*proto.Target{
			NewDefaultTarget(),
		},
	}
}

func RaidSimTest(label string, t *testing.T, rsr *proto.RaidSimRequest, expectedDps float64) {
	result := RunRaidSim(rsr)
	if result.Error != nil {
		t.Fatalf("Sim failed with error: %s", result.Error.Message)
	}
	tolerance := 0.5
	if result.RaidMetrics.Dps.Avg < expectedDps-tolerance || result.RaidMetrics.Dps.Avg > expectedDps+tolerance {
		// Automatically print output if we had debugging enabled.
		if rsr.SimOptions.Debug {
			log.Printf("LOGS:\n%s\n", result.Logs)
		}
		t.Fatalf("%s failed: expected %0f dps from sim but was %0f", label, expectedDps, result.RaidMetrics.Dps.Avg)
	}
}

func RaidBenchmark(b *testing.B, rsr *proto.RaidSimRequest) {
	rsr.Encounter.Duration = LongDuration
	rsr.SimOptions.Iterations = 1

	// Set to false because IsTest adds a lot of computation.
	rsr.SimOptions.IsTest = false

	for i := 0; i < b.N; i++ {
		result := RunRaidSim(rsr)
		if result.Error != nil {
			b.Fatalf("RaidBenchmark() at iteration %d failed: %v", i, result.Error.Message)
		}
	}
}

// CharacterBenchmark sims a suite config's default setup against one target. The seed is fixed,
// so the reported dps only moves when a change alters results.
func CharacterBenchmark(b *testing.B, config CharacterSuiteConfig) {
	_, raid := config.defaultPlayerAndRaid()
	rsr := &proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  Ternary(config.Encounter.Encounter != nil, config.Encounter.Encounter, MakeSingleTargetEncounter(0)),
		SimOptions: &proto.SimOptions{Iterations: 1000, RandomSeed: 101},
	}

	b.ReportAllocs()
	var dps float64
	for b.Loop() {
		result := RunRaidSim(rsr)
		if result.Error != nil {
			b.Fatalf("CharacterBenchmark() failed: %v", result.Error.Message)
		}
		dps = result.RaidMetrics.Dps.Avg
	}
	b.ReportMetric(dps, "dps")
}

// Reads a preset under ui/specs that a spec writes in SIML, with the files its use lines name: nil when there is no
// such file, and the spec's JSON is read instead. One that does not read panics with its diagnostics.
func readSIMLPreset(path string) *siml.File {
	f, err := siml.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		panic(err)
	}
	return f
}

// The gear sets and encounter builds were removed with the rest of the TBC presets, so a
// missing file is now an expected state rather than a broken checkout. log.Fatalf killed
// the whole process and took the test output with it; returning an empty combo lets the
// suite run and report a normal failure instead.
func GetAplRotation(dir string, file string) RotationCombo {
	if f := readSIMLPreset(dir + "/" + file + ".apl.siml"); f != nil {
		return RotationCombo{Label: file, Rotation: f.Rotation()}
	}
	filePath := dir + "/" + file + ".apl.json"
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("no apl json file, skipping rotation: %s", filePath)
		return RotationCombo{Label: file}
	}

	return RotationCombo{Label: file, Rotation: APLRotationFromJsonString(string(data))}
}

// The rotation with a spell's rank moved to another rank by spell id, so the other rank reaches the results.
func RotationWithSpellIds(label string, combo RotationCombo, ids map[int32]int32) RotationCombo {
	rotation := googleProto.Clone(combo.Rotation).(*proto.APLRotation)
	swapped := map[int32]bool{}
	swapSpellIds(rotation.ProtoReflect(), ids, swapped)
	for from := range ids {
		if !swapped[from] {
			panic(fmt.Sprintf("the %s rotation does not name spell %d", combo.Label, from))
		}
	}
	return RotationCombo{Label: label, Rotation: rotation}
}

// Swaps every spell id that ids maps for the id it maps to, except an external buff copy (tag -1), and marks each id
// it swapped. A swapped id drops its rank, which belonged to the old id.
func swapSpellIds(message protoreflect.Message, ids map[int32]int32, swapped map[int32]bool) {
	if id, ok := message.Interface().(*proto.ActionID); ok {
		if to, ok := ids[id.GetSpellId()]; ok && id.Tag != -1 {
			swapped[id.GetSpellId()] = true
			id.RawId = &proto.ActionID_SpellId{SpellId: to}
			id.Rank = 0
		}
		return
	}
	message.Range(func(field protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		switch {
		case field.IsList() && field.Message() != nil:
			for i := 0; i < value.List().Len(); i++ {
				swapSpellIds(value.List().Get(i).Message(), ids, swapped)
			}
		case field.Message() != nil && !field.IsMap():
			swapSpellIds(value.Message(), ids, swapped)
		}
		return true
	})
}

func GetGearSet(dir string, file string) GearSetCombo {
	if f := readSIMLPreset(dir + "/" + file + ".gear.siml"); f != nil {
		return GearSetCombo{Label: file, GearSet: f.Equipment()}
	}
	filePath := dir + "/" + file + ".gear.json"
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("no gear json file, skipping gear set: %s", filePath)
		return GearSetCombo{Label: file, GearSet: &proto.EquipmentSpec{}}
	}

	return GearSetCombo{Label: file, GearSet: EquipmentSpecFromJsonString(string(data))}
}

func GetItemSwapGearSet(dir string, file string) ItemSwapSetCombo {
	if f := readSIMLPreset(dir + "/" + file + ".gear.siml"); f != nil {
		return ItemSwapSetCombo{Label: file, ItemSwap: &proto.ItemSwap{Items: f.Equipment().Items}}
	}
	filePath := dir + "/" + file + ".gear.json"
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("failed to load gear json file: %s, %s", filePath, err)
	}

	return ItemSwapSetCombo{Label: file, ItemSwap: ItemSwapFromJsonString(string(data))}
}

func GenerateTalentVariations(baseTalents string) []TalentsCombo {
	return GenerateTalentVariationsForRows(baseTalents, []int{0, 1, 2, 3, 4, 5})
}

func GenerateTalentVariationsForRows(baseTalents string, rowsToVary []int) []TalentsCombo {
	if len(baseTalents) != 6 {
		log.Fatalf("Expected 6-digit talent string, got: %s", baseTalents)
	}

	var combinations []TalentsCombo

	baseRunes := []rune(baseTalents)
	for _, row := range rowsToVary {
		if row < 0 || row >= 6 {
			log.Fatalf("Invalid row index: %d, must be between 0 and 5", row)
		}

		for choice := 1; choice <= 3; choice++ {
			if int(baseRunes[row]-'0') == choice {
				continue
			}

			variation := make([]rune, 6)
			copy(variation, baseRunes)
			variation[row] = rune('0' + choice)

			combinations = append(combinations, TalentsCombo{
				Label:   fmt.Sprintf("Row%d_Talent%d", row+1, choice),
				Talents: string(variation),
			})
		}
	}

	return combinations
}

func GetTestBuildFromJSON(class proto.Class, dir string, file string, itemFilter ItemFilter, epReferenceStat *proto.Stat, statsToWeigh *[]proto.Stat) CharacterSuiteConfig {
	filePath := dir + "/" + file + ".build.json"
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("failed to load gear json file: %s, %s", filePath, err)
	}

	simSettings := &proto.IndividualSimSettings{}
	if err := protojson.Unmarshal(data, simSettings); err != nil {
		panic(err)
	}

	config := CharacterSuiteConfig{
		Class:       class,
		Race:        simSettings.Player.Race,
		Profession1: simSettings.Player.Profession1,
		Profession2: simSettings.Player.Profession2,
		GearSet: GearSetCombo{
			Label:   file,
			GearSet: simSettings.Player.Equipment,
		},
		SpecOptions: SpecOptionsCombo{
			Label:       file,
			SpecOptions: getPlayerSpecOptions(simSettings.Player),
		},
		Talents: simSettings.Player.TalentsString,
		Rotation: RotationCombo{
			Label:    file,
			Rotation: simSettings.Player.Rotation,
		},
		Encounter: EncounterCombo{
			Label:     file,
			Encounter: simSettings.Encounter,
		},
		ItemSwapSet: ItemSwapSetCombo{
			Label:    file,
			ItemSwap: simSettings.Player.ItemSwap,
		},
		StartingDistance:   simSettings.Player.DistanceFromTarget,
		ReactionTimeMs:     simSettings.Player.ReactionTimeMs,
		ChannelClipDelayMs: simSettings.Player.ChannelClipDelayMs,

		Consumables:     simSettings.Player.Consumables,
		IndividualBuffs: simSettings.Player.Buffs,
		PartyBuffs:      simSettings.PartyBuffs,
		RaidBuffs:       simSettings.RaidBuffs,
		Debuffs:         simSettings.Debuffs,
		Cooldowns:       simSettings.Player.Cooldowns,

		InFrontOfTarget: simSettings.Player.InFrontOfTarget,
		TargetDummies:   simSettings.TargetDummies,
		HealingModel:    simSettings.Player.HealingModel,

		ItemFilter: itemFilter,
	}

	if simSettings.Tanks != nil {
		config.Tanks = simSettings.Tanks

		// Check if any of the tanks is the player.
		for _, tank := range simSettings.Tanks {
			if tank.Type == proto.UnitReference_Player {
				config.IsTank = true
				break
			}
		}
	}

	if epReferenceStat != nil {
		config.EPReferenceStat = *epReferenceStat
	}
	if statsToWeigh != nil {
		config.StatsToWeigh = *statsToWeigh
	}

	return config
}

func getPlayerSpecOptions(player *proto.Player) interface{} {
	if playerSpec, ok := player.Spec.(*proto.Player_BalanceDruid); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_FeralCatDruid); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_FeralBearDruid); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_RestorationDruid); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_Hunter); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_Mage); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_HolyPaladin); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_ProtectionPaladin); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_RetributionPaladin); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_DpsPriest); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_HealerPriest); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_Rogue); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_ElementalShaman); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_EnhancementShaman); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_RestorationShaman); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_Warlock); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_DpsWarrior); ok {
		return playerSpec
	}
	if playerSpec, ok := player.Spec.(*proto.Player_ProtectionWarrior); ok {
		return playerSpec
	}

	panic("Unsupported spec provided to getPlayerSpecOptions. Please add a case for the spec.")
}

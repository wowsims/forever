package buffmanifest

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestUniqueScopeField(t *testing.T) {
	seen := map[string]bool{}
	for _, row := range All() {
		key := fmt.Sprintf("%s/%s", row.Scope, row.Field)
		if seen[key] {
			t.Errorf("duplicate (scope, field): %s", key)
		}
		seen[key] = true
	}
}

func TestUniqueGoStem(t *testing.T) {
	seen := map[string]bool{}
	for _, row := range All() {
		if seen[row.GoStem()] {
			t.Errorf("duplicate Go stem: %s", row.GoStem())
		}
		seen[row.GoStem()] = true
	}
}

func TestFlagRowsHaveAReason(t *testing.T) {
	for _, row := range All() {
		if row.Kind == KindFlag && row.Reason == "" {
			t.Errorf("%s is %s and needs a Reason", row.Field, row.Kind)
		}
		if row.Kind != KindFlag && row.Reason != "" {
			t.Errorf("%s is %s, which renders no Reason", row.Field, row.Kind)
		}
	}
}

func TestResolvableRowsNameASpell(t *testing.T) {
	for _, row := range All() {
		if (row.SpellID == 0) != (row.Kind == KindFlag) {
			t.Errorf("%s (%s) names spell %d", row.Field, row.Kind, row.SpellID)
		}
	}
}

func TestProtoType(t *testing.T) {
	cases := []struct {
		spec BuffSpec
		want BuffProtoType
	}{
		{BuffSpec{}, ProtoBool},
		{BuffSpec{Talent: 16187}, ProtoTristate},
		{BuffSpec{ImpAction: &ActionRef{SpellID: 23563}}, ProtoTristate},
		{BuffSpec{Kind: KindExternalCD}, ProtoInt32},
		{BuffSpec{Kind: KindItemCount}, ProtoInt32},
		{BuffSpec{Kind: KindFlag, Proto: ProtoDouble}, ProtoDouble},
	}
	for _, c := range cases {
		if got := c.spec.ProtoType(); got != c.want {
			t.Errorf("%+v is %s, want %s", c.spec, got, c.want)
		}
	}
}

func TestFieldNaming(t *testing.T) {
	cases := []struct {
		field, goName, tsName string
	}{
		{"battle_shout", "BattleShout", "battleShout"},
		{"soe_enhancement_2pt4", "SoeEnhancement_2Pt4", "soeEnhancement2Pt4"},
		{"joc_retribution_2pt4", "JocRetribution_2Pt4", "jocRetribution2Pt4"},
		{"snapshot_bs_booming_voice_rank", "SnapshotBsBoomingVoiceRank", "snapshotBsBoomingVoiceRank"},
		{"isb_uptime", "IsbUptime", "isbUptime"},
		{"expose_weakness_hunter_agility", "ExposeWeaknessHunterAgility", "exposeWeaknessHunterAgility"},
	}
	for _, c := range cases {
		spec := BuffSpec{Field: c.field}
		if got := spec.GoField(); got != c.goName {
			t.Errorf("GoField(%q) = %q, want %q", c.field, got, c.goName)
		}
		if got := spec.TSField(); got != c.tsName {
			t.Errorf("TSField(%q) = %q, want %q", c.field, got, c.tsName)
		}
	}
}

func TestFieldNamesRoundTrip(t *testing.T) {
	for _, spec := range All() {
		goName := spec.GoField()
		if goName == "" {
			t.Errorf("%s has an empty GoField", spec.Field)
			continue
		}
		want := strings.ReplaceAll(goName, "_", "")
		want = strings.ToLower(want[:1]) + want[1:]
		if got := spec.TSField(); got != want {
			t.Errorf("%s: TSField %q does not match GoField %q", spec.Field, got, goName)
		}
	}
}

// proto/buffs.proto is rendered from this manifest, so this test is not an independent oracle for
// the field set: it catches a hand edit to the committed proto drifting from the manifest, and
// nothing more. What the numbers and types are checked against is tools/database's
// TestBuffFilesRegenerateFromTheCommittedInputs: the committed file is what the manifest renders.
func TestBuffsMatchProto(t *testing.T) {
	messages := map[string]BuffScope{
		"RaidBuffs":       ScopeRaid,
		"PartyBuffs":      ScopeParty,
		"IndividualBuffs": ScopeIndividual,
		"Debuffs":         ScopeDebuff,
	}

	want := map[string]protoField{}
	for message, scope := range messages {
		for field, declared := range parseProtoMessage(t, message) {
			want[fmt.Sprintf("%s/%s", scope, field)] = declared
		}
	}

	got := map[string]bool{}
	for _, row := range All() {
		key := fmt.Sprintf("%s/%s", row.Scope, row.Field)
		got[key] = true

		declared, ok := want[key]
		if !ok {
			t.Errorf("%s is in the manifest but not in buffs.proto", key)
			continue
		}
		if row.Number != declared.number {
			t.Errorf("%s has number %d in the manifest and %d in buffs.proto", key, row.Number, declared.number)
		}
		if !declared.allows(row.ProtoType()) {
			t.Errorf("%s is %s in the manifest and %s in buffs.proto", key, row.ProtoType(), declared.kind)
		}
	}
	for key := range want {
		if !got[key] {
			t.Errorf("%s is in buffs.proto but not in the manifest", key)
		}
	}
}

type protoField struct {
	kind   string
	number int32
}

// allows reports whether a manifest row may carry proto type p.
func (f protoField) allows(p BuffProtoType) bool {
	switch f.kind {
	case "bool":
		return p == ProtoBool
	case "int32":
		return p == ProtoInt32
	case "double":
		return p == ProtoDouble
	case "TristateEffect":
		return p == ProtoTristate
	}
	return false
}

var protoFieldRE = regexp.MustCompile(`^\s*([A-Za-z][\w.]*)\s+([a-z][a-z0-9_]*)\s*=\s*(\d+)\s*;`)

func parseProtoMessage(t *testing.T, message string) map[string]protoField {
	t.Helper()

	raw, err := os.ReadFile("../../../proto/buffs.proto")
	if err != nil {
		t.Fatalf("read buffs.proto: %v", err)
	}

	fields := map[string]protoField{}
	inMessage := false
	for line := range strings.SplitSeq(string(raw), "\n") {
		if !inMessage {
			inMessage = strings.HasPrefix(strings.TrimSpace(line), "message "+message+" {")
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "}") {
			break
		}
		match := protoFieldRE.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		number, err := strconv.Atoi(match[3])
		if err != nil {
			t.Fatalf("field %s of %s: %v", match[2], message, err)
		}
		fields[match[2]] = protoField{kind: match[1], number: int32(number)}
	}

	if len(fields) == 0 {
		t.Fatalf("no fields parsed for message %s", message)
	}
	return fields
}

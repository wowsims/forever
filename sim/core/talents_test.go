package core

import (
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	googleproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// A digit above a talent's talent_max_rank fills the field at that rank, and the rank the APL reads
// is the capped one.
func TestFillTalentsProtoCapsADigitAtTheTalentsMaxRank(t *testing.T) {
	talents := &proto.WarriorTalents{}
	fields := talents.ProtoReflect().Descriptor().Fields()
	digits := []byte(strings.Repeat("0", fields.Len()))

	var ranked protoreflect.FieldDescriptor
	for i := range fields.Len() {
		if fd := fields.Get(i); fd.Kind() == protoreflect.Int32Kind {
			ranked = fd
			break
		}
	}
	if ranked == nil {
		t.Fatal("WarriorTalents has no ranked talent")
	}
	spellID := googleproto.GetExtension(ranked.Options(), proto.E_TalentSpellId).(int32)
	maxRank := googleproto.GetExtension(ranked.Options(), proto.E_TalentMaxRank).(int32)
	if spellID == 0 || maxRank == 0 {
		t.Fatalf("%s states no talent_spell_id or no talent_max_rank", ranked.Name())
	}
	digits[ranked.Number()-1] = '9'

	char := &Character{}
	char.FillTalentsProto(talents.ProtoReflect(), string(digits), [3]int{fields.Len(), 0, 0})
	if got := talents.ProtoReflect().Get(ranked).Int(); got != int64(maxRank) {
		t.Errorf("a digit of 9 fills %s at rank %d, want the talent's %d", ranked.Name(), got, maxRank)
	}
	if got := char.TalentRank(spellID); got != maxRank {
		t.Errorf("a digit of 9 records rank %d for the APL, want the talent's %d", got, maxRank)
	}
}

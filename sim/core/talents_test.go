package core

import (
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	googleproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// A digit above a talent's talent_max_rank fills the field at that rank, any digit above 0 takes a
// one-point talent, and the rank the APL reads is the capped one, for every talent of every class.
func TestFillTalentsProtoCapsADigitAtTheTalentsMaxRank(t *testing.T) {
	for _, msg := range []googleproto.Message{
		&proto.DruidTalents{}, &proto.HunterTalents{}, &proto.MageTalents{}, &proto.PaladinTalents{}, &proto.PriestTalents{},
		&proto.RogueTalents{}, &proto.ShamanTalents{}, &proto.WarlockTalents{}, &proto.WarriorTalents{},
	} {
		talents := msg.ProtoReflect()
		fields := talents.Descriptor().Fields()
		char := &Character{}
		char.FillTalentsProto(talents, strings.Repeat("9", fields.Len()), [3]int{fields.Len(), 0, 0})
		for i := range fields.Len() {
			fd := fields.Get(i)
			want := int32(1)
			if fd.Kind() == protoreflect.Int32Kind {
				want = googleproto.GetExtension(fd.Options(), proto.E_TalentMaxRank).(int32)
				if got := talents.Get(fd).Int(); want == 0 || got != int64(want) {
					t.Errorf("a digit of 9 fills %s at rank %d, want its talent_max_rank %d", fd.FullName(), got, want)
				}
			} else if !talents.Get(fd).Bool() {
				t.Errorf("a digit of 9 leaves %s untaken", fd.FullName())
			}
			spellID := googleproto.GetExtension(fd.Options(), proto.E_TalentSpellId).(int32)
			if got := char.TalentRank(spellID); got != want {
				t.Errorf("a digit of 9 records rank %d for %s, want %d", got, fd.FullName(), want)
			}
		}
	}
}

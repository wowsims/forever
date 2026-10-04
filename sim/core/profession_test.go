package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestMeetsProfession(t *testing.T) {
	engineering, tailoring, alchemy := proto.Profession_Engineering, proto.Profession_Tailoring, proto.Profession_Alchemy
	maxed := &Character{professions: [2]proto.Profession{engineering, tailoring}}
	ranked := &Character{professions: [2]proto.Profession{engineering, tailoring}, professionRank: [2]int32{200, 300}}

	for _, c := range []struct {
		name      string
		character *Character
		prof      proto.Profession
		rank      int32
		want      bool
	}{
		{"no profession required", ranked, proto.Profession_ProfessionUnknown, 0, true},
		{"maxed meets any rank", maxed, engineering, 300, true},
		{"rank above the requirement", ranked, engineering, 175, true},
		{"rank equal to the requirement", ranked, engineering, 200, true},
		{"rank below the requirement", ranked, engineering, 285, false},
		{"second profession's rank", ranked, tailoring, 285, true},
		{"profession not held", maxed, alchemy, 0, false},
	} {
		if got := c.character.MeetsProfession(c.prof, c.rank); got != c.want {
			t.Errorf("%s: MeetsProfession(%v, %d) = %v, want %v", c.name, c.prof, c.rank, got, c.want)
		}
	}
}

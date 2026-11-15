package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestAreaTypeSetDropsUnknownAndDuplicates(t *testing.T) {
	got := areaTypeSet([]proto.AreaType{
		proto.AreaType_AreaTypeUnknown,
		proto.AreaType_AreaTypeForestGrassland,
		proto.AreaType_AreaTypeMountainous,
		proto.AreaType_AreaTypeForestGrassland,
	})
	want := []proto.AreaType{proto.AreaType_AreaTypeForestGrassland, proto.AreaType_AreaTypeMountainous}
	if len(got) != len(want) {
		t.Fatalf("areaTypeSet = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("areaTypeSet = %v, want %v", got, want)
		}
	}

	encounter := Encounter{AreaTypes: got}
	if !encounter.InArea(proto.AreaType_AreaTypeMountainous) || encounter.InArea(proto.AreaType_AreaTypeDesert) {
		t.Fatalf("InArea reads the wrong set: %v", got)
	}
}

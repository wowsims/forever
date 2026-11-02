package main

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database"
)

func TestPruneQuestieSources(t *testing.T) {
	drop := func(npcID, zoneID int32, otherName string) *proto.UIItemSource {
		return &proto.UIItemSource{Source: &proto.UIItemSource_Drop{Drop: &proto.DropSource{NpcId: npcID, ZoneId: zoneID, OtherName: otherName}}}
	}
	soldBy := func(zoneID int32) *proto.UIItemSource {
		return &proto.UIItemSource{Source: &proto.UIItemSource_SoldBy{SoldBy: &proto.SoldBySource{NpcId: 1, ZoneId: zoneID}}}
	}
	quest := &proto.UIItemSource{Source: &proto.UIItemSource_Quest{Quest: &proto.QuestSource{Id: 1, Name: "Bounty"}}}
	worldDrop := drop(0, 0, database.WorldDropName)
	inGame, unknownZone, noZone := drop(1, 28, ""), drop(2, 1585, ""), drop(3, 0, "")

	questie := database.NewWowDatabase()
	questie.Items[1] = &proto.UIItem{Id: 1, Sources: []*proto.UIItemSource{quest, soldBy(28), soldBy(1585), inGame, unknownZone, noZone, worldDrop}}
	pruneQuestieSources(questie, map[int32]bool{28: true})

	want := []*proto.UIItemSource{quest, questie.Items[1].Sources[1], inGame, worldDrop}
	got := questie.Items[1].Sources
	if len(got) != len(want) {
		t.Fatalf("sources = %v, want the quest, the vendor and NPC in a known zone, and the world drop", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("source %d = %v, want %v", i, got[i], want[i])
		}
	}
}

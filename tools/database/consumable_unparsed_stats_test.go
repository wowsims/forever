package database

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestUnparsedStatsListVersusStatsChannelPercentsAndZoneBonus(t *testing.T) {
	flask := func(id, spellID int32) *proto.Consumable {
		stats := make([]float64, len(proto.Stat_name))
		stats[proto.Stat_StatStamina] = 60
		return &proto.Consumable{Id: id, Stats: stats, ItemEffects: []*proto.ItemEffect{{BuffId: spellID}}}
	}
	db := NewWowDatabase()
	db.Consumables[9224] = &proto.Consumable{Id: 9224, ItemEffects: []*proto.ItemEffect{{BuffId: 11406}}}
	db.Consumables[23327] = &proto.Consumable{Id: 23327, ItemEffects: []*proto.ItemEffect{{BuffId: 29332}}}
	db.Consumables[22237] = &proto.Consumable{Id: 22237, ItemEffects: []*proto.ItemEffect{{BuffId: 27723}}}
	db.Consumables[13510] = &proto.Consumable{Id: 13510, ItemEffects: []*proto.ItemEffect{{BuffId: 17626}}}
	db.Consumables[211954] = &proto.Consumable{Id: 211954, ItemEffects: []*proto.ItemEffect{{BuffId: 430949}}}
	db.Consumables[211957] = &proto.Consumable{Id: 211957, ItemEffects: []*proto.ItemEffect{{BuffId: 430952}}}
	db.Consumables[13452] = &proto.Consumable{Id: 13452, ItemEffects: []*proto.ItemEffect{{BuffId: 17538}}}
	db.Consumables[250337] = &proto.Consumable{Id: 250337, ItemEffects: []*proto.ItemEffect{{BuffId: 1250940}}}
	for id, spellID := range map[int32]int32{274273: 1293740, 274274: 1293741, 274275: 1293742, 274276: 1293743} {
		db.Consumables[id] = flask(id, spellID)
	}
	if err := AddUnparsedStats(db); err != nil {
		t.Fatal(err)
	}

	meleeHit := map[proto.PseudoStat]float64{proto.PseudoStat_PseudoStatMeleeHitPercent: 2, proto.PseudoStat_PseudoStatRangedHitPercent: 2}
	allCrit := map[proto.PseudoStat]float64{
		proto.PseudoStat_PseudoStatMeleeCritPercent: 2, proto.PseudoStat_PseudoStatRangedCritPercent: 2, proto.PseudoStat_PseudoStatSpellCritPercent: 2,
	}
	for id, want := range map[int32]struct {
		stats       map[proto.Stat]float64
		pseudoStats map[proto.PseudoStat]float64
	}{
		9224:   {stats: map[proto.Stat]float64{proto.Stat_StatAttackPower: 180}},
		23327:  {pseudoStats: meleeHit},
		22237:  {pseudoStats: meleeHit},
		211954: {pseudoStats: map[proto.PseudoStat]float64{proto.PseudoStat_PseudoStatSpellHitPercent: 1}},
		211957: {pseudoStats: map[proto.PseudoStat]float64{proto.PseudoStat_PseudoStatSpellCritPercent: 1}},
		13452:  {pseudoStats: allCrit},
		250337: {pseudoStats: allCrit},
		274273: {stats: map[proto.Stat]float64{proto.Stat_StatStamina: 60, proto.Stat_StatHitRating: 50}},
		274274: {stats: map[proto.Stat]float64{proto.Stat_StatStamina: 60, proto.Stat_StatCritRating: 56}},
		274275: {stats: map[proto.Stat]float64{proto.Stat_StatStamina: 60, proto.Stat_StatExpertiseRating: 50}},
		274276: {stats: map[proto.Stat]float64{proto.Stat_StatStamina: 60, proto.Stat_StatHasteRating: 50}},
	} {
		consumable := db.Consumables[id]
		if (len(consumable.Stats) == 0) != (len(want.stats) == 0) {
			t.Errorf("%d lists stats %v, want %v", id, consumable.Stats, want.stats)
		}
		for stat, value := range consumable.Stats {
			if value != want.stats[proto.Stat(stat)] {
				t.Errorf("%d lists %v %v, want %v", id, proto.Stat(stat), value, want.stats[proto.Stat(stat)])
			}
		}
		if (len(consumable.PseudoStats) == 0) != (len(want.pseudoStats) == 0) {
			t.Errorf("%d lists pseudo stats %v, want %v", id, consumable.PseudoStats, want.pseudoStats)
		}
		for stat, value := range consumable.PseudoStats {
			if value != want.pseudoStats[proto.PseudoStat(stat)] {
				t.Errorf("%d lists %v %v, want %v", id, proto.PseudoStat(stat), value, want.pseudoStats[proto.PseudoStat(stat)])
			}
		}
	}
	if titans := db.Consumables[13510]; len(titans.Stats) != 0 || len(titans.PseudoStats) != 0 {
		t.Errorf("Flask of the Titans lists %v %v, want nothing added", titans.Stats, titans.PseudoStats)
	}
}

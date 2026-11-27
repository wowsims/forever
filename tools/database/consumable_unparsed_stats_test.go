package database

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	googleProto "google.golang.org/protobuf/proto"
)

func TestUnparsedStatsListChannelPercentsAndZoneBonus(t *testing.T) {
	flask := func(id, spellID int32) *proto.Consumable {
		stats := make([]float64, len(proto.Stat_name))
		stats[proto.Stat_StatStamina] = 60
		return &proto.Consumable{Id: id, Stats: stats, ItemEffects: []*proto.ItemEffect{{BuffId: spellID}}}
	}
	db := NewWowDatabase()
	db.Consumables[23327] = &proto.Consumable{Id: 23327, ItemEffects: []*proto.ItemEffect{{BuffId: 29332}}}
	db.Consumables[22237] = &proto.Consumable{Id: 22237, ItemEffects: []*proto.ItemEffect{{BuffId: 27723}}}
	db.Consumables[13510] = &proto.Consumable{Id: 13510, ItemEffects: []*proto.ItemEffect{{BuffId: 17626}}}
	db.Consumables[211954] = &proto.Consumable{Id: 211954, ItemEffects: []*proto.ItemEffect{{BuffId: 430949}}}
	db.Consumables[211957] = &proto.Consumable{Id: 211957, ItemEffects: []*proto.ItemEffect{{BuffId: 430952}}}
	db.Consumables[13452] = &proto.Consumable{Id: 13452, ItemEffects: []*proto.ItemEffect{{BuffId: 17538}}}
	db.Consumables[250337] = &proto.Consumable{Id: 250337, ItemEffects: []*proto.ItemEffect{{BuffId: 1250940}}}
	db.Consumables[13928] = &proto.Consumable{Id: 13928, ItemEffects: []*proto.ItemEffect{{BuffId: 1249522}}}
	for id, spellID := range map[int32]int32{274273: 1293740, 274274: 1293741, 274275: 1293742, 274276: 1293743} {
		db.Consumables[id] = flask(id, spellID)
	}
	// Grilled Squid's Well Fed states a crit of its own other than the amount its periodic trigger hands
	// over, so its case tells the two apart.
	editRow(t, spelldata.MustFind(1249522).FirstAura(dbcenums.A_PERIODIC_TRIGGER_SPELL_WITH_VALUE).TriggerID, func(s *spelldata.Spell) {
		s.Effects = slices.Clone(s.Effects)
		for i := range s.Effects {
			if s.Effects[i].Aura == dbcenums.A_MOD_CRIT_PCT {
				s.Effects[i].BasePoints++
			}
		}
	})
	if err := AddUnparsedStats(db); err != nil {
		t.Fatal(err)
	}

	amount := func(spellID int32, aura dbcenums.EffectAuraType) float64 {
		return spelldata.MustFind(spellID).FirstAura(aura).Average()
	}
	zoneBonus := func(spellID int32) float64 {
		bonus := spelldata.MustFind(spellID).EffectN(2).Average()
		if bonus == 0 {
			t.Fatalf("%d states no zone bonus", spellID)
		}
		return bonus
	}
	stamina := map[proto.Stat]float64{proto.Stat_StatStamina: 60}
	meleeAndRangedHit := func(spellID int32) map[proto.PseudoStat]float64 {
		hit := amount(spellID, dbcenums.A_MOD_HIT_CHANCE)
		return map[proto.PseudoStat]float64{proto.PseudoStat_PseudoStatMeleeHitPercent: hit, proto.PseudoStat_PseudoStatRangedHitPercent: hit}
	}
	allCrit := func(crit float64) map[proto.PseudoStat]float64 {
		return map[proto.PseudoStat]float64{
			proto.PseudoStat_PseudoStatMeleeCritPercent: crit, proto.PseudoStat_PseudoStatRangedCritPercent: crit, proto.PseudoStat_PseudoStatSpellCritPercent: crit,
		}
	}
	for id, want := range map[int32]struct {
		stats       map[proto.Stat]float64
		pseudoStats map[proto.PseudoStat]float64
		zoneStats   map[int32]float64
	}{
		23327:  {pseudoStats: meleeAndRangedHit(29332)},
		22237:  {pseudoStats: meleeAndRangedHit(27723)},
		211954: {pseudoStats: map[proto.PseudoStat]float64{proto.PseudoStat_PseudoStatSpellHitPercent: amount(430949, dbcenums.A_MOD_SPELL_HIT_CHANCE)}},
		211957: {pseudoStats: map[proto.PseudoStat]float64{proto.PseudoStat_PseudoStatSpellCritPercent: amount(430952, dbcenums.A_MOD_SPELL_CRIT_CHANCE)}},
		13452:  {pseudoStats: allCrit(amount(17538, dbcenums.A_MOD_CRIT_PCT))},
		250337: {pseudoStats: allCrit(amount(1250940, dbcenums.A_MOD_CRIT_PCT))},
		// Grilled Squid's crit is on the Well Fed 1249522 triggers, at the value its periodic trigger hands over.
		13928:  {pseudoStats: allCrit(amount(1249522, dbcenums.A_PERIODIC_TRIGGER_SPELL_WITH_VALUE))},
		274273: {stats: stamina, zoneStats: map[int32]float64{int32(proto.Stat_StatHitRating): zoneBonus(1293740) * core.HitRatingPerHitPercent}},
		274274: {stats: stamina, zoneStats: map[int32]float64{int32(proto.Stat_StatCritRating): zoneBonus(1293741) * core.CritRatingPerCritPercent}},
		274275: {stats: stamina, zoneStats: map[int32]float64{int32(proto.Stat_StatExpertiseRating): zoneBonus(1293742) * core.ExpertiseRatingPerExpertisePercent}},
		274276: {stats: stamina, zoneStats: map[int32]float64{int32(proto.Stat_StatHasteRating): zoneBonus(1293743) * core.HasteRatingPerHastePercent}},
	} {
		consumable := db.Consumables[id]
		var wantZoneStats []*proto.ZoneStats
		if want.zoneStats != nil {
			wantZoneStats = []*proto.ZoneStats{{ZoneIds: naturalFlaskZones, Stats: want.zoneStats}}
		}
		if !slices.EqualFunc(consumable.ZoneStats, wantZoneStats, func(a, b *proto.ZoneStats) bool { return googleProto.Equal(a, b) }) {
			t.Errorf("%d lists zone stats %v, want %v", id, consumable.ZoneStats, wantZoneStats)
		}
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
	if titans := db.Consumables[13510]; len(titans.Stats) != 0 || len(titans.PseudoStats) != 0 || len(titans.MobTypeStats) != 0 || len(titans.ZoneStats) != 0 {
		t.Errorf("Flask of the Titans lists %v %v %v %v, want nothing added", titans.Stats, titans.PseudoStats, titans.MobTypeStats, titans.ZoneStats)
	}
}

func TestUnparsedStatsListVersusStatsUnderTheirMobTypes(t *testing.T) {
	db := NewWowDatabase()
	db.Consumables[9224] = &proto.Consumable{Id: 9224, ItemEffects: []*proto.ItemEffect{{BuffId: 11406}}}
	db.Consumables[250954] = &proto.Consumable{Id: 250954, ItemEffects: []*proto.ItemEffect{{BuffId: 1252558}}}
	if err := AddUnparsedStats(db); err != nil {
		t.Fatal(err)
	}

	for id, versus := range map[int32]struct {
		spellID int32
		aura    dbcenums.EffectAuraType
		stat    proto.Stat
	}{
		9224:   {11406, dbcenums.A_MOD_MELEE_ATTACK_POWER_VERSUS, proto.Stat_StatAttackPower},
		250954: {1252558, dbcenums.A_MOD_FLAT_SPELL_DAMAGE_VERSUS, proto.Stat_StatSpellDamage},
	} {
		e := spelldata.MustFind(versus.spellID).FirstAura(versus.aura)
		mobTypes := spelldata.MobTypes(e.Misc)
		if e.Average() == 0 || len(mobTypes) == 0 {
			t.Fatalf("%d's versus aura states %v against %v, want an amount against some mob type", id, e.Average(), mobTypes)
		}
		consumable := db.Consumables[id]
		if len(consumable.Stats) != 0 || len(consumable.PseudoStats) != 0 {
			t.Errorf("%d lists %v %v, want its versus stat kept out of the unconditional stats", id, consumable.Stats, consumable.PseudoStats)
		}
		want := []*proto.MobTypeStats{{MobTypes: mobTypes, Stats: map[int32]float64{int32(versus.stat): e.Average()}}}
		if !slices.EqualFunc(consumable.MobTypeStats, want, func(a, b *proto.MobTypeStats) bool { return googleProto.Equal(a, b) }) {
			t.Errorf("%d lists mob type stats %v, want %v", id, consumable.MobTypeStats, want)
		}
	}
}

func TestUnparsedStatsListAVersusStatWithNoMobTypeAsAPlainStat(t *testing.T) {
	versus := func(mask int32, amount float64) spelldata.Effect {
		return spelldata.Effect{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_MELEE_ATTACK_POWER_VERSUS, Misc: mask, BasePoints: amount}
	}
	const demons = 1 << (3 - 1)
	row := &spelldata.Spell{Effects: []spelldata.Effect{versus(0, 40), versus(demons, 60)}}
	melee := spelldata.VersusAura(dbcenums.A_MOD_MELEE_ATTACK_POWER_VERSUS)
	stat := melee.Stats[0]
	consumable := &proto.Consumable{}
	if err := addUnparsedStats(0, consumable, row); err != nil {
		t.Fatal(err)
	}

	if len(consumable.Stats) <= int(stat) || consumable.Stats[stat] != row.Effects[0].Average() {
		t.Errorf("lists stats %v, want %v %v from the aura that names no mob type", consumable.Stats, stat, row.Effects[0].Average())
	}
	want := []*proto.MobTypeStats{{MobTypes: spelldata.MobTypes(demons), Stats: map[int32]float64{int32(stat): row.Effects[1].Average()}}}
	if !slices.EqualFunc(consumable.MobTypeStats, want, func(a, b *proto.MobTypeStats) bool { return googleProto.Equal(a, b) }) {
		t.Errorf("lists mob type stats %v, want %v", consumable.MobTypeStats, want)
	}
}

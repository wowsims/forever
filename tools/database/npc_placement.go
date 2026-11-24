package database

import (
	"database/sql"
	"fmt"
)

// The difficulty of the Forever dungeons: every live dungeon, vanilla and new, has a MapDifficulty row with it, and
// none of the Season of Discovery maps and raids in the game files has one.
const foreverDungeonDifficulty = 201

type instanceMapRow struct {
	ID                  int32
	Name                string
	InstanceType        int32
	AreaTableID         int32
	CosmeticParentMapID int32
	ForeverDifficulty   bool
}

type instanceAreaRow struct {
	ID          int32
	Name        string
	ContinentID int32
	ParentID    int32
}

type encounterRow struct {
	Name  string
	MapID int32
}

// Places QuestieDB's NPCs in the dungeon or raid the game files put them in, where QuestieDB gives an open-world zone,
// a zone AreaTable lacks, or none.
type NpcPlacement struct {
	maps       map[int32]instanceMapRow
	areas      map[int32]instanceAreaRow
	zoneMaps   map[int32]int32    // Each instance zone's map.
	targets    map[int32]int32    // Each instance map's zone, which its NPCs' drops take.
	encounters map[string][]int32 // The maps of each DungeonEncounter name.
}

// An instance map's zone is its AreaTableID, else its top-level area named as the map. An area is in an instance when
// it is the map's AreaTableID or lies on the map.
func newNpcPlacement(maps []instanceMapRow, areas []instanceAreaRow, encounters []encounterRow) *NpcPlacement {
	p := &NpcPlacement{
		maps:       map[int32]instanceMapRow{},
		areas:      map[int32]instanceAreaRow{},
		zoneMaps:   map[int32]int32{},
		targets:    map[int32]int32{},
		encounters: map[string][]int32{},
	}
	for _, m := range maps {
		p.maps[m.ID] = m
		if m.AreaTableID != 0 {
			p.zoneMaps[m.AreaTableID] = m.ID
			p.targets[m.ID] = m.AreaTableID
		}
	}
	for _, area := range areas {
		p.areas[area.ID] = area
		m, ok := p.maps[area.ContinentID]
		if !ok {
			continue
		}
		p.zoneMaps[area.ID] = m.ID
		if _, set := p.targets[m.ID]; !set && area.ParentID == 0 && area.Name == m.Name {
			p.targets[m.ID] = area.ID
		}
	}
	for _, encounter := range encounters {
		if _, ok := p.maps[encounter.MapID]; ok {
			p.encounters[encounter.Name] = append(p.encounters[encounter.Name], encounter.MapID)
		}
	}
	return p
}

// The instance map a zone is in, for the quests filed under it.
func (p *NpcPlacement) instanceOf(zone int32) (int32, bool) {
	m, ok := p.zoneMaps[zone]
	return m, ok
}

// The zone of a QuestieDB NPC, given the instances whose quests ask for its kill. The NPC belongs to an instance when a
// DungeonEncounter there has its name or one of those quests is filed there. It keeps its zone when that names more
// than one instance (Balnazzar is a Stratholme and a Scarlet Enclave encounter), names one with no zone to move to, or
// when it already sits in a dungeon or raid. It moves from a zone AreaTable lacks (BRD's 1585) or from a world zone on
// the continent the map's CosmeticParentMapID names (City of Dalaran's bosses in Alterac Mountains), and from no other
// zone. That keeps open-world NPCs named as a Season of Discovery encounter (the Emerald dragons, Azuregos, Kazzak)
// and Highland Horror in the world. An NPC with no zone joins a dungeon with the Forever dungeon difficulty only,
// never a raid, nor a map without that difficulty.
func (p *NpcPlacement) zone(name string, zone int32, questMaps []int32) int32 {
	signal := map[int32]bool{}
	for _, maps := range [][]int32{questMaps, p.encounters[name]} {
		for _, m := range maps {
			signal[m] = true
		}
	}
	if _, inInstance := p.zoneMaps[zone]; inInstance || len(signal) != 1 {
		return zone
	}
	var m instanceMapRow
	for id := range signal {
		m = p.maps[id]
	}
	target, ok := p.targets[m.ID]
	area, known := p.areas[zone]
	switch {
	case !ok:
		return zone
	case zone == 0:
		if m.InstanceType != 1 || !m.ForeverDifficulty {
			return zone
		}
	case known:
		if m.CosmeticParentMapID == -1 || area.ContinentID != m.CosmeticParentMapID {
			return zone
		}
	}
	return target
}

func LoadNpcPlacement(dbHelper *DBHelper) (*NpcPlacement, error) {
	maps, err := LoadRows(dbHelper.db, `
		SELECT ID, MapName_lang, InstanceType, AreaTableID, CosmeticParentMapID,
			EXISTS (SELECT 1 FROM MapDifficulty WHERE MapID = Map.ID AND DifficultyID = ?)
		FROM Map WHERE InstanceType IN (1, 2)
	`, func(rows *sql.Rows) (instanceMapRow, error) {
		var m instanceMapRow
		return m, rows.Scan(&m.ID, &m.Name, &m.InstanceType, &m.AreaTableID, &m.CosmeticParentMapID, &m.ForeverDifficulty)
	}, foreverDungeonDifficulty)
	if err != nil {
		return nil, fmt.Errorf("loading instance maps: %w", err)
	}
	areas, err := LoadRows(dbHelper.db, `SELECT ID, AreaName_lang, ContinentID, ParentAreaID FROM AreaTable`,
		func(rows *sql.Rows) (instanceAreaRow, error) {
			var a instanceAreaRow
			return a, rows.Scan(&a.ID, &a.Name, &a.ContinentID, &a.ParentID)
		})
	if err != nil {
		return nil, fmt.Errorf("loading AreaTable: %w", err)
	}
	encounters, err := LoadRows(dbHelper.db, `SELECT Name_lang, MapID FROM DungeonEncounter`,
		func(rows *sql.Rows) (encounterRow, error) {
			var e encounterRow
			return e, rows.Scan(&e.Name, &e.MapID)
		})
	if err != nil {
		return nil, fmt.Errorf("loading DungeonEncounter: %w", err)
	}
	return newNpcPlacement(maps, areas, encounters), nil
}

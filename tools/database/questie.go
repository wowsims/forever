package database

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools"
)

// QuestieDB's Forever data (github.com/Questie/QuestieDB, GPLv3) names each item's quest rewards and vendors, which
// the game files do not hold.
const questieDataURL = "https://raw.githubusercontent.com/Questie/QuestieDB/master/data/Forever/"

// Every item QuestieDB lists a reward quest or a vendor for, with each as a Quest or SoldBy source.
func ReadQuestieData() *WowDatabase {
	db, err := parseQuestieData(
		tools.ReadWebRequired(questieDataURL+"foreverItemDB.lua"),
		tools.ReadWebRequired(questieDataURL+"foreverQuestDB.lua"),
		tools.ReadWebRequired(questieDataURL+"foreverNpcDB.lua"),
	)
	if err != nil {
		log.Fatalf("parsing the QuestieDB data: %v", err)
	}
	return db
}

func parseQuestieData(itemSrc, questSrc, npcSrc string) (*WowDatabase, error) {
	itemKeys, items, err := readQuestieFile(itemSrc, "item")
	if err != nil {
		return nil, err
	}
	questKeys, quests, err := readQuestieFile(questSrc, "quest")
	if err != nil {
		return nil, err
	}
	npcKeys, npcs, err := readQuestieFile(npcSrc, "npc")
	if err != nil {
		return nil, err
	}
	var questRewards, vendors, questName, npcName, npcZone int
	for _, key := range []struct {
		keys  map[string]int
		name  string
		index *int
	}{
		{itemKeys, "questRewards", &questRewards},
		{itemKeys, "vendors", &vendors},
		{questKeys, "name", &questName},
		{npcKeys, "name", &npcName},
		{npcKeys, "zoneID", &npcZone},
	} {
		index, ok := key.keys[key.name]
		if !ok {
			return nil, fmt.Errorf("no key %q", key.name)
		}
		*key.index = index
	}

	db := NewWowDatabase()
	for id, item := range items {
		var sources []*proto.UIItemSource
		for _, questID := range item.ids(questRewards) {
			if quest, ok := quests[questID]; ok {
				name, _ := quest.field(questName).(string)
				sources = append(sources, &proto.UIItemSource{Source: &proto.UIItemSource_Quest{Quest: &proto.QuestSource{
					Id:   int32(questID),
					Name: name,
				}}})
			}
		}
		for _, npcID := range item.ids(vendors) {
			if npc, ok := npcs[npcID]; ok {
				name, _ := npc.field(npcName).(string)
				zone, _ := npc.field(npcZone).(float64)
				sources = append(sources, &proto.UIItemSource{Source: &proto.UIItemSource_SoldBy{SoldBy: &proto.SoldBySource{
					NpcId:   int32(npcID),
					NpcName: name,
					ZoneId:  int32(zone),
				}}})
			}
		}
		if len(sources) > 0 {
			db.Items[int32(id)] = &proto.UIItem{Id: int32(id), Sources: sources}
		}
	}
	return db, nil
}

// The field indexes of QuestieDB.<kind>Keys and the rows of QuestieDB.<kind>Data, a Lua chunk in a long string.
func readQuestieFile(src, kind string) (map[string]int, map[int]*luaTable, error) {
	keysTable, err := parseLuaAfter(src, "QuestieDB."+kind+"Keys =")
	if err != nil {
		return nil, nil, err
	}
	keys := map[string]int{}
	for key, value := range keysTable.keyed {
		name, isName := key.(string)
		index, isIndex := value.(float64)
		if !isName || !isIndex {
			return nil, nil, fmt.Errorf("QuestieDB.%sKeys: %v = %v is not a name and an index", kind, key, value)
		}
		keys[name] = int(index)
	}

	dataTable, err := parseLuaAfter(src, "QuestieDB."+kind+"Data = [[return")
	if err != nil {
		return nil, nil, fmt.Errorf("QuestieDB.%sData: %w", kind, err)
	}
	rows := make(map[int]*luaTable, len(dataTable.keyed))
	for key, value := range dataTable.keyed {
		id, isID := key.(float64)
		row, isRow := value.(*luaTable)
		if !isID || !isRow {
			return nil, nil, fmt.Errorf("QuestieDB.%sData: %v is not an id and a row", kind, key)
		}
		rows[int(id)] = row
	}
	return keys, rows, nil
}

// A Lua table constructor: the positional values in order, nil included, and and the [key] = value fields. Numbers are float64 and strings string.
type luaTable struct {
	list  []any
	keyed map[any]any
}

// The positional value at a 1-based index, nil when the table is shorter.
func (t *luaTable) field(index int) any {
	if index < 1 || index > len(t.list) {
		return nil
	}
	return t.list[index-1]
}

func (t *luaTable) ids(index int) []int {
	list, _ := t.field(index).(*luaTable)
	if list == nil {
		return nil
	}
	ids := make([]int, 0, len(list.list))
	for _, value := range list.list {
		if id, ok := value.(float64); ok {
			ids = append(ids, int(id))
		}
	}
	return ids
}

func parseLuaAfter(src, marker string) (*luaTable, error) {
	start := strings.Index(src, marker)
	if start < 0 {
		return nil, fmt.Errorf("no %q", marker)
	}
	p := &luaParser{src: src, pos: start + len(marker)}
	value, err := p.value()
	if err != nil {
		return nil, err
	}
	table, ok := value.(*luaTable)
	if !ok {
		return nil, fmt.Errorf("%q is not followed by a table", marker)
	}
	return table, nil
}

// Reads the subset of Lua literals QuestieDB writes: tables, quoted strings, numbers and nil, with -- comments to
// the end of the line.
type luaParser struct {
	src string
	pos int
}

func (p *luaParser) errorf(format string, args ...any) error {
	return fmt.Errorf("offset %d: %s", p.pos, fmt.Sprintf(format, args...))
}

func (p *luaParser) skipSpace() {
	for p.pos < len(p.src) {
		switch {
		case strings.HasPrefix(p.src[p.pos:], "--"):
			if end := strings.IndexByte(p.src[p.pos:], '\n'); end >= 0 {
				p.pos += end + 1
			} else {
				p.pos = len(p.src)
			}
		case strings.IndexByte(" \t\r\n", p.src[p.pos]) >= 0:
			p.pos++
		default:
			return
		}
	}
}

func (p *luaParser) value() (any, error) {
	p.skipSpace()
	if p.pos >= len(p.src) {
		return nil, p.errorf("unexpected end")
	}
	switch c := p.src[p.pos]; {
	case c == '{':
		return p.table()
	case c == '\'' || c == '"':
		return p.string()
	case c == '-' || c == '.' || (c >= '0' && c <= '9'):
		return p.number()
	}
	if strings.HasPrefix(p.src[p.pos:], "nil") {
		p.pos += len("nil")
		return nil, nil
	}
	return nil, p.errorf("unexpected %q", p.src[p.pos])
}

func (p *luaParser) table() (*luaTable, error) {
	p.pos++
	table := &luaTable{keyed: map[any]any{}}
	for {
		p.skipSpace()
		if p.pos >= len(p.src) {
			return nil, p.errorf("unclosed table")
		}
		if p.src[p.pos] == '}' {
			p.pos++
			return table, nil
		}

		var key any
		if p.src[p.pos] == '[' {
			p.pos++
			k, err := p.value()
			if err != nil {
				return nil, err
			}
			p.skipSpace()
			if !strings.HasPrefix(p.src[p.pos:], "]") {
				return nil, p.errorf("expected ]")
			}
			p.pos++
			if err := p.expect('='); err != nil {
				return nil, err
			}
			key = k
		}

		value, err := p.value()
		if err != nil {
			return nil, err
		}
		if key == nil {
			table.list = append(table.list, value)
		} else {
			table.keyed[key] = value
		}

		p.skipSpace()
		if p.pos < len(p.src) && p.src[p.pos] == ',' {
			p.pos++
		} else if p.pos >= len(p.src) || p.src[p.pos] != '}' {
			return nil, p.errorf("expected , or }")
		}
	}
}

func (p *luaParser) expect(c byte) error {
	p.skipSpace()
	if p.pos >= len(p.src) || p.src[p.pos] != c {
		return p.errorf("expected %q", c)
	}
	p.pos++
	return nil
}

func (p *luaParser) string() (string, error) {
	quote := p.src[p.pos]
	p.pos++
	var text strings.Builder
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		p.pos++
		switch c {
		case quote:
			return text.String(), nil
		case '\n':
			return "", p.errorf("newline in string")
		case '\\':
			if p.pos >= len(p.src) {
				return "", p.errorf("unclosed string")
			}
			escaped := p.src[p.pos]
			p.pos++
			switch escaped {
			case 'n':
				text.WriteByte('\n')
			case 't':
				text.WriteByte('\t')
			case '\\', '\'', '"':
				text.WriteByte(escaped)
			default:
				return "", p.errorf("unknown escape \\%c", escaped)
			}
		default:
			text.WriteByte(c)
		}
	}
	return "", p.errorf("unclosed string")
}

func (p *luaParser) number() (float64, error) {
	start := p.pos
	for p.pos < len(p.src) && strings.IndexByte("-.0123456789", p.src[p.pos]) >= 0 {
		p.pos++
	}
	number, err := strconv.ParseFloat(p.src[start:p.pos], 64)
	if err != nil {
		return 0, p.errorf("bad number %q", p.src[start:p.pos])
	}
	return number, nil
}

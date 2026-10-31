package database

import (
	"archive/zip"
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools"
)

// QuestieDB (github.com/Questie/QuestieDB, GPLv3) names each item's quest rewards and vendors, which the game files
// do not hold. A release's baked Forever data has QuestieDB's Forever corrections applied; the repository's
// data/Forever Lua files do not.
const questieRelease = "v1.0.9"

// Every item QuestieDB lists a reward quest or a vendor for, with each as a Quest or SoldBy source.
func ReadQuestieData() *WowDatabase {
	archive := tools.ReadWebRequired("https://github.com/Questie/QuestieDB/releases/download/" + questieRelease + "/QuestieDB-Forever.zip")
	reader, err := zip.NewReader(strings.NewReader(archive), int64(len(archive)))
	if err != nil {
		log.Fatalf("opening the QuestieDB release: %v", err)
	}
	read := func(name string) string {
		text, err := readZipFile(reader, "QuestieDB/"+name)
		if err != nil {
			log.Fatalf("reading the QuestieDB release: %v", err)
		}
		return text
	}
	db, err := parseQuestieData(read("QuestieDB_Forever.toc"), read("src/meta/itemMeta.lua"), read("src/meta/questMeta.lua"), read("src/meta/npcMeta.lua"))
	if err != nil {
		log.Fatalf("parsing the QuestieDB data: %v", err)
	}
	return db
}

func readZipFile(reader *zip.Reader, name string) (string, error) {
	file, err := reader.Open(name)
	if err != nil {
		return "", err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	return string(data), err
}

// The baked data is the TOC's "## X-<entity>-<id>-<field>: <value>" lines; each meta file maps an entity's field
// names to indexes.
func parseQuestieData(toc, itemMeta, questMeta, npcMeta string) (*WowDatabase, error) {
	store := readQuestieStore(toc)
	items, err := newQuestieEntity(store, "Item", itemMeta)
	if err != nil {
		return nil, err
	}
	quests, err := newQuestieEntity(store, "Quest", questMeta)
	if err != nil {
		return nil, err
	}
	npcs, err := newQuestieEntity(store, "Npc", npcMeta)
	if err != nil {
		return nil, err
	}
	var questRewards, vendors, questName, npcName, npcZone int
	for _, key := range []struct {
		entity *questieEntity
		name   string
		index  *int
	}{
		{items, "questRewards", &questRewards},
		{items, "vendors", &vendors},
		{quests, "name", &questName},
		{npcs, "name", &npcName},
		{npcs, "zoneID", &npcZone},
	} {
		index, ok := key.entity.keys[key.name]
		if !ok {
			return nil, fmt.Errorf("%s has no key %q", key.entity.name, key.name)
		}
		*key.index = index
	}

	db := NewWowDatabase()
	for _, id := range items.ids {
		item, err := items.row(id)
		if err != nil {
			return nil, err
		}
		questIDs, err := item.ids(questRewards)
		if err != nil {
			return nil, err
		}
		npcIDs, err := item.ids(vendors)
		if err != nil {
			return nil, err
		}

		var sources []*proto.UIItemSource
		for _, questID := range questIDs {
			quest, err := quests.row(questID)
			if err != nil {
				return nil, err
			}
			if quest == nil {
				continue
			}
			name, err := quest.text(questName)
			if err != nil {
				return nil, err
			}
			sources = append(sources, &proto.UIItemSource{Source: &proto.UIItemSource_Quest{Quest: &proto.QuestSource{
				Id:   int32(questID),
				Name: name,
			}}})
		}
		for _, npcID := range npcIDs {
			npc, err := npcs.row(npcID)
			if err != nil {
				return nil, err
			}
			if npc == nil {
				continue
			}
			name, err := npc.text(npcName)
			if err != nil {
				return nil, err
			}
			zone, err := npc.integer(npcZone)
			if err != nil {
				return nil, err
			}
			sources = append(sources, &proto.UIItemSource{Source: &proto.UIItemSource_SoldBy{SoldBy: &proto.SoldBySource{
				NpcId:   int32(npcID),
				NpcName: name,
				ZoneId:  int32(zone),
			}}})
		}
		if len(sources) > 0 {
			db.Items[int32(id)] = &proto.UIItem{Id: int32(id), Sources: sources}
		}
	}
	return db, nil
}

// The TOC's X- fields by key, without the "X-".
type questieStore map[string]string

func readQuestieStore(toc string) questieStore {
	store := questieStore{}
	for _, line := range strings.Split(toc, "\n") {
		if key, value, ok := strings.Cut(strings.TrimSuffix(line, "\r"), ": "); ok && strings.HasPrefix(key, "## X-") {
			store[strings.TrimPrefix(key, "## X-")] = value
		}
	}
	return store
}

var questieChunkHeader = regexp.MustCompile(`^~(\d+)~$`)

// The base64-decoded value of a key, false when the key is missing. A "~N~" value stands for the values of
// <key>-1 to <key>-N joined.
func (s questieStore) base64(key string) ([]byte, bool, error) {
	value, ok := s[key]
	if !ok {
		return nil, false, nil
	}
	if header := questieChunkHeader.FindStringSubmatch(value); header != nil {
		parts, _ := strconv.Atoi(header[1])
		var joined strings.Builder
		for i := 1; i <= parts; i++ {
			part, ok := s[key+"-"+strconv.Itoa(i)]
			if !ok {
				return nil, false, fmt.Errorf("%s declares %d parts but part %d is missing", key, parts, i)
			}
			joined.WriteString(part)
		}
		value = joined.String()
	}
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, false, fmt.Errorf("%s: %w", key, err)
	}
	return data, true, nil
}

// The CBOR value of a key, nil when the key is missing.
func (s questieStore) value(key string) (any, error) {
	data, ok, err := s.base64(key)
	if err != nil || !ok {
		return nil, err
	}
	value, err := decodeCBOR(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", key, err)
	}
	return value, nil
}

type questieEntity struct {
	name  string
	store questieStore
	keys  map[string]int // Field names to indexes, from the keys table of the entity's meta file.
	ids   []int          // Every id the entity has, from <name>-IDS: a CBOR list, zlib compressed.
	known map[int]bool
}

func newQuestieEntity(store questieStore, name, meta string) (*questieEntity, error) {
	keys, err := readQuestieKeys(meta)
	if err != nil {
		return nil, fmt.Errorf("%s meta: %w", name, err)
	}
	compressed, ok, err := store.base64(name + "-IDS")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("no %s-IDS", name)
	}
	ids, err := decodeQuestieIDs(compressed)
	if err != nil {
		return nil, fmt.Errorf("%s-IDS: %w", name, err)
	}
	entity := &questieEntity{name: name, store: store, keys: keys, ids: ids, known: make(map[int]bool, len(ids))}
	for _, id := range ids {
		entity.known[id] = true
	}
	return entity, nil
}

func decodeQuestieIDs(compressed []byte) ([]int, error) {
	reader, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	value, err := decodeCBOR(data)
	if err != nil {
		return nil, err
	}
	return cborIDs(value)
}

// The scalar row of an id, nil when the entity has no such id. It is stored under <name>-<id>-S, as a map from field
// index or as a list when the indexes run from 1 without a gap.
func (e *questieEntity) row(id int) (*questieRow, error) {
	if !e.known[id] {
		return nil, nil
	}
	key := fmt.Sprintf("%s-%d", e.name, id)
	value, err := e.store.value(key + "-S")
	if err != nil {
		return nil, err
	}
	row := &questieRow{entity: e, key: key, fields: map[int64]any{}}
	switch value := value.(type) {
	case nil:
	case []any:
		for i, field := range value {
			row.fields[int64(i+1)] = field
		}
	case map[any]any:
		for index, field := range value {
			switch index := index.(type) {
			case int64:
				row.fields[index] = field
			case string:
				presence, ok := field.(int64)
				if index != "p" || !ok {
					return nil, fmt.Errorf("%s-S: %q = %v is not the presence bits", key, index, field)
				}
				row.presence = presence
			}
		}
	default:
		return nil, fmt.Errorf("%s-S: %v is not a row", key, value)
	}
	return row, nil
}

type questieRow struct {
	entity   *questieEntity
	key      string
	fields   map[int64]any
	presence int64 // Bit index-1 is set for each table field the row has, each stored under <key>-<index>.
}

func (r *questieRow) text(index int) (string, error) {
	switch value := r.fields[int64(index)].(type) {
	case nil:
		return "", nil
	case string:
		return value, nil
	default:
		return "", fmt.Errorf("%s field %d: %v is not a string", r.key, index, value)
	}
}

func (r *questieRow) integer(index int) (int64, error) {
	switch value := r.fields[int64(index)].(type) {
	case nil:
		return 0, nil
	case int64:
		return value, nil
	default:
		return 0, fmt.Errorf("%s field %d: %v is not an integer", r.key, index, value)
	}
}

func (r *questieRow) ids(index int) ([]int, error) {
	if r.presence&(1<<(index-1)) == 0 {
		return nil, nil
	}
	key := fmt.Sprintf("%s-%d", r.key, index)
	value, err := r.entity.store.value(key)
	if err != nil {
		return nil, err
	}
	ids, err := cborIDs(value)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", key, err)
	}
	return ids, nil
}

func cborIDs(value any) ([]int, error) {
	list, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("%v is not a list", value)
	}
	ids := make([]int, len(list))
	for i, id := range list {
		number, ok := id.(int64)
		if !ok {
			return nil, fmt.Errorf("%v is not an id", id)
		}
		ids[i] = int(number)
	}
	return ids, nil
}

// Decodes the CBOR (RFC 8949) the baked data holds: integers as int64, byte and text strings as string, floats as
// float64, false, true, null, lists as []any and maps as map[any]any with int64 or string keys. Tags and indefinite
// lengths fail.
func decodeCBOR(data []byte) (any, error) {
	decoder := &cborDecoder{data: data}
	value, err := decoder.value()
	if err != nil {
		return nil, err
	}
	if decoder.pos != len(data) {
		return nil, fmt.Errorf("CBOR: %d bytes after the value", len(data)-decoder.pos)
	}
	return value, nil
}

type cborDecoder struct {
	data []byte
	pos  int
}

func (d *cborDecoder) next(n uint64) ([]byte, error) {
	if n > uint64(len(d.data)-d.pos) {
		return nil, fmt.Errorf("CBOR: %d bytes past the end at offset %d", n, d.pos)
	}
	chunk := d.data[d.pos : d.pos+int(n)]
	d.pos += int(n)
	return chunk, nil
}

func (d *cborDecoder) value() (any, error) {
	head, err := d.next(1)
	if err != nil {
		return nil, err
	}
	major, info := head[0]>>5, head[0]&31
	var arg uint64
	switch {
	case info < 24:
		arg = uint64(info)
	case info <= 27:
		chunk, err := d.next(1 << (info - 24))
		if err != nil {
			return nil, err
		}
		for _, b := range chunk {
			arg = arg<<8 | uint64(b)
		}
	default:
		return nil, fmt.Errorf("CBOR: additional information %d at offset %d", info, d.pos-1)
	}

	switch major {
	case 0, 1:
		if arg > math.MaxInt64 {
			return nil, fmt.Errorf("CBOR: integer out of range at offset %d", d.pos)
		}
		if major == 1 {
			return -1 - int64(arg), nil
		}
		return int64(arg), nil
	case 2, 3:
		chunk, err := d.next(arg)
		return string(chunk), err
	case 4:
		// Every value takes a byte at least, which bounds the length before allocating.
		if arg > uint64(len(d.data)-d.pos) {
			return nil, fmt.Errorf("CBOR: list of %d past the end at offset %d", arg, d.pos)
		}
		list := make([]any, arg)
		for i := range list {
			if list[i], err = d.value(); err != nil {
				return nil, err
			}
		}
		return list, nil
	case 5:
		if arg > uint64(len(d.data)-d.pos)/2 {
			return nil, fmt.Errorf("CBOR: map of %d past the end at offset %d", arg, d.pos)
		}
		table := make(map[any]any, arg)
		for range arg {
			key, err := d.value()
			if err != nil {
				return nil, err
			}
			switch key.(type) {
			case int64, string:
			default:
				return nil, fmt.Errorf("CBOR: map key %v at offset %d", key, d.pos)
			}
			if table[key], err = d.value(); err != nil {
				return nil, err
			}
		}
		return table, nil
	case 7:
		switch info {
		case 20:
			return false, nil
		case 21:
			return true, nil
		case 22:
			return nil, nil
		case 25:
			return halfFloat(uint16(arg)), nil
		case 26:
			return float64(math.Float32frombits(uint32(arg))), nil
		case 27:
			return math.Float64frombits(arg), nil
		}
	}
	return nil, fmt.Errorf("CBOR: major type %d, additional information %d at offset %d", major, info, d.pos)
}

// An IEEE 754 binary16 value.
func halfFloat(bits uint16) float64 {
	sign := 1.0
	if bits&0x8000 != 0 {
		sign = -1
	}
	exponent, fraction := int(bits>>10)&0x1f, float64(bits&0x3ff)
	switch exponent {
	case 0:
		return sign * math.Ldexp(fraction, -24)
	case 0x1f:
		if fraction != 0 {
			return math.NaN()
		}
		return math.Inf(int(sign))
	}
	return sign * math.Ldexp(fraction+1024, exponent-25)
}

// A line that opens the keys table, so that a name ending in "keys" or a comment quoting it does not match.
var questieKeysTable = regexp.MustCompile(`(?m)^\s*keys = \{`)

// The field names and indexes of a meta file's table `keys = { ['name'] = 1, -- comment ... }`.
func readQuestieKeys(meta string) (map[string]int, error) {
	start := questieKeysTable.FindStringIndex(meta)
	if start == nil {
		return nil, fmt.Errorf("no keys table")
	}
	p := &luaParser{src: meta, pos: start[1]}
	keys := map[string]int{}
	for {
		p.skipSpace()
		if p.pos >= len(p.src) {
			return nil, p.errorf("unclosed table")
		}
		if p.src[p.pos] == '}' {
			return keys, nil
		}
		if err := p.expect('['); err != nil {
			return nil, err
		}
		name, err := p.string()
		if err != nil {
			return nil, err
		}
		if err := p.expect(']'); err != nil {
			return nil, err
		}
		if err := p.expect('='); err != nil {
			return nil, err
		}
		index, err := p.index()
		if err != nil {
			return nil, err
		}
		keys[name] = index
		p.skipSpace()
		if p.pos < len(p.src) && p.src[p.pos] == ',' {
			p.pos++
		} else if p.pos >= len(p.src) || p.src[p.pos] != '}' {
			return nil, p.errorf("expected , or }")
		}
	}
}

// Reads the Lua a keys table holds: quoted strings without escapes, field indexes and -- comments to the end of the
// line.
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

func (p *luaParser) expect(c byte) error {
	p.skipSpace()
	if p.pos >= len(p.src) || p.src[p.pos] != c {
		return p.errorf("expected %q", c)
	}
	p.pos++
	return nil
}

func (p *luaParser) string() (string, error) {
	p.skipSpace()
	if p.pos >= len(p.src) || (p.src[p.pos] != '\'' && p.src[p.pos] != '"') {
		return "", p.errorf("expected a string")
	}
	quote := p.src[p.pos]
	end := strings.IndexByte(p.src[p.pos+1:], quote)
	if end < 0 {
		return "", p.errorf("unclosed string")
	}
	text := p.src[p.pos+1 : p.pos+1+end]
	if strings.ContainsAny(text, "\\\n") {
		return "", p.errorf("escape or newline in string %q", text)
	}
	p.pos += end + 2
	return text, nil
}

// A field index, from 1 to 63 so that it has a presence bit.
func (p *luaParser) index() (int, error) {
	p.skipSpace()
	start := p.pos
	for p.pos < len(p.src) && p.src[p.pos] >= '0' && p.src[p.pos] <= '9' {
		p.pos++
	}
	index, err := strconv.Atoi(p.src[start:p.pos])
	if err != nil || index < 1 || index > 63 {
		return 0, p.errorf("bad field index %q", p.src[start:p.pos])
	}
	return index, nil
}

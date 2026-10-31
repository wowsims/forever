package database

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

// The subset of CBOR the baked data holds, with strings as byte strings as it writes them.
func encodeCBOR(value any) []byte {
	head := func(major byte, arg int) []byte {
		switch {
		case arg < 24:
			return []byte{major<<5 | byte(arg)}
		case arg < 1<<8:
			return []byte{major<<5 | 24, byte(arg)}
		default:
			return []byte{major<<5 | 25, byte(arg >> 8), byte(arg)}
		}
	}
	switch value := value.(type) {
	case nil:
		return []byte{0xf6}
	case int:
		if value < 0 {
			return head(1, -1-value)
		}
		return head(0, value)
	case string:
		return append(head(2, len(value)), value...)
	case []any:
		out := head(4, len(value))
		for _, element := range value {
			out = append(out, encodeCBOR(element)...)
		}
		return out
	case map[any]any:
		out := head(5, len(value))
		for key, element := range value {
			out = append(append(out, encodeCBOR(key)...), encodeCBOR(element)...)
		}
		return out
	}
	panic(fmt.Sprintf("cannot encode %T", value))
}

func questieValue(value any) string {
	return base64.StdEncoding.EncodeToString(encodeCBOR(value))
}

func questieIDs(ids ...any) string {
	var compressed bytes.Buffer
	writer := zlib.NewWriter(&compressed)
	writer.Write(encodeCBOR(ids))
	writer.Close()
	return base64.StdEncoding.EncodeToString(compressed.Bytes())
}

// Splits a value into a "~N~" header and its N parts.
func questieChunks(key, value string, parts int) string {
	lines := []string{fmt.Sprintf("## X-%s: ~%d~", key, parts)}
	size := (len(value) + parts - 1) / parts
	for i := range parts {
		lines = append(lines, fmt.Sprintf("## X-%s-%d: %s", key, i+1, value[min(i*size, len(value)):min((i+1)*size, len(value))]))
	}
	return strings.Join(lines, "\n")
}

const questieTestItemMeta = `local meta = {
  entity = 'Item',
  keys = {
    -- Localized item name.
    ['name'] = 1,
    -- Quest IDs that reward this item.
    ['questRewards'] = 2,
    ['vendors'] = 3, -- NPC IDs for vendors that sell this item.
  },
  types = { [1]='string', [2]='table', [3]='table' },
}`

const questieTestQuestMeta = `local meta = {
  keys = {
    ['name'] = 1,
    ['zoneOrSort'] = 2, -- int, <0: QuestSort.dbc ID
  },
}`

const questieTestNpcMeta = `local meta = {
  keys = {
    ['name'] = 1,
    ['spawns'] = 2, -- table {[zoneID(int)] = {coordPair(floatVector2D),...},...}
    ['zoneID'] = 3,
  },
}`

const questieTestQuestRewards = "## X-Item-100-2: "
const questieTestQuestRow = "## X-Quest-8561-S: "

// Item 100 has a quest with no row, item 101's vendors are chunked and sold by a vendor stored as a list row and one
// with a null zone, item 102 has no table fields, and item 103 has a vendors field its presence bits leave out.
func questieTestTOC() string {
	return strings.Join([]string{
		"## Interface: 16001",
		"## X-Flavor: Forever",
		questieChunks("Item-IDS", questieIDs(100, 101, 102, 103), 3),
		"## X-Item-100-S: " + questieValue(map[any]any{1: "Conqueror's Crown", "p": 1 << 1}),
		questieTestQuestRewards + questieValue([]any{8561, 9999}),
		"## X-Item-101-S: " + questieValue(map[any]any{1: "Plain Ring", "p": 1 << 2}),
		questieChunks("Item-101-3", questieValue([]any{5, 6}), 2),
		"## X-Item-102-S: " + questieValue([]any{"Unsold, Unrewarded"}),
		"## X-Item-103-S: " + questieValue(map[any]any{1: "Stale Vendors"}),
		"## X-Item-103-3: " + questieValue([]any{5}),
		"## X-Quest-IDS: " + questieIDs(8561),
		questieTestQuestRow + questieValue(map[any]any{1: "Conqueror's Crown", 2: -15}),
		"## X-Npc-IDS: " + questieIDs(5, 6),
		"## X-Npc-5-S: " + questieValue([]any{`Grizzle "Rings, Bands" O'Hare`, nil, 1519}),
		"## X-Npc-6-S: " + questieValue(map[any]any{1: "Nowhere Vendor", 3: nil}),
	}, "\r\n")
}

func TestParseQuestieData(t *testing.T) {
	db, err := parseQuestieData(questieTestTOC(), questieTestItemMeta, questieTestQuestMeta, questieTestNpcMeta)
	if err != nil {
		t.Fatal(err)
	}

	want := map[int32][]*proto.UIItemSource{
		100: {{Source: &proto.UIItemSource_Quest{Quest: &proto.QuestSource{Id: 8561, Name: "Conqueror's Crown"}}}},
		101: {
			{Source: &proto.UIItemSource_SoldBy{SoldBy: &proto.SoldBySource{NpcId: 5, NpcName: `Grizzle "Rings, Bands" O'Hare`, ZoneId: 1519}}},
			{Source: &proto.UIItemSource_SoldBy{SoldBy: &proto.SoldBySource{NpcId: 6, NpcName: "Nowhere Vendor"}}},
		},
	}
	if len(db.Items) != len(want) {
		t.Errorf("got %d items, want %d: an item with neither quests nor vendors is left out, and so is a table field without its presence bit", len(db.Items), len(want))
	}
	for id, sources := range want {
		item, ok := db.Items[id]
		if !ok {
			t.Errorf("item %d is missing", id)
			continue
		}
		if len(item.Sources) != len(sources) {
			t.Errorf("item %d has %d sources, want %d: a quest with no row is left out", id, len(item.Sources), len(sources))
			continue
		}
		for i := range sources {
			if !googleProto.Equal(item.Sources[i], sources[i]) {
				t.Errorf("item %d source %d = %v, want %v", id, i, item.Sources[i], sources[i])
			}
		}
	}
}

func TestParseQuestieDataRejectsMalformedData(t *testing.T) {
	toc := questieTestTOC()
	vendorsPart := "## X-Item-101-3-2: "
	for name, test := range map[string]struct{ toc, itemMeta string }{
		"missing chunk":      {strings.Replace(toc, vendorsPart, "## X-Item-101-3-9: ", 1), questieTestItemMeta},
		"bad base64":         {strings.Replace(toc, vendorsPart, vendorsPart+"!", 1), questieTestItemMeta},
		"truncated CBOR":     {strings.Replace(toc, questieTestQuestRewards, questieTestQuestRewards+questieValue([]any{8561, 9999})[:4]+"\n## X-Unused: ", 1), questieTestItemMeta},
		"id is not a number": {strings.Replace(toc, questieTestQuestRewards, questieTestQuestRewards+questieValue([]any{"8561"})+"\n## X-Unused: ", 1), questieTestItemMeta},
		"name is a number":   {strings.Replace(toc, questieTestQuestRow, questieTestQuestRow+questieValue([]any{8561})+"\n## X-Unused: ", 1), questieTestItemMeta},
		"missing key":        {toc, strings.Replace(questieTestItemMeta, "['vendors'] = 3,", "", 1)},
		"no IDS":             {strings.Replace(toc, "## X-Npc-IDS: ", "## X-Npc-Ids: ", 1), questieTestItemMeta},
		"bad field index":    {toc, strings.Replace(questieTestItemMeta, "['name'] = 1,", "['name'] = 64,", 1)},
	} {
		if _, err := parseQuestieData(test.toc, test.itemMeta, questieTestQuestMeta, questieTestNpcMeta); err == nil {
			t.Errorf("%s: parsed without an error", name)
		}
	}
}

func TestDecodeQuestieCBOR(t *testing.T) {
	for _, test := range []struct {
		data []byte
		want any
	}{
		{[]byte{0x19, 0x01, 0xf4}, int64(500)},
		{[]byte{0x39, 0x01, 0xf3}, int64(-500)},
		{[]byte{0x63, 'a', 'b', 'c'}, "abc"},
		{[]byte{0xf9, 0x3e, 0x00}, 1.5},
		{[]byte{0xf9, 0xc4, 0x00}, -4.0},
		{[]byte{0xf9, 0x00, 0x01}, math.Ldexp(1, -24)},
		{[]byte{0xfa, 0x3f, 0xc0, 0x00, 0x00}, 1.5},
		{[]byte{0xfb, 0x40, 0x39, 0x0f, 0x5c, 0x28, 0xf5, 0xc2, 0x8f}, 25.06},
		{[]byte{0xf5}, true},
	} {
		got, err := decodeCBOR(test.data)
		if err != nil || got != test.want {
			t.Errorf("% x = %v (%v), want %v", test.data, got, err, test.want)
		}
	}

	for name, data := range map[string][]byte{
		"indefinite length": {0x9f, 0x01, 0xff},
		"tag":               {0xc1, 0x00},
		"truncated":         {0x18},
		"trailing bytes":    {0x01, 0x02},
		"list past the end": {0x99, 0xff, 0xff},
		"list as map key":   {0xa1, 0x80, 0x01},
	} {
		if _, err := decodeCBOR(data); err == nil {
			t.Errorf("%s: decoded without an error", name)
		}
	}
}

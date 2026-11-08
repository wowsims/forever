package database

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func itemsWithIds(ids ...int32) map[int32]*proto.UIItem {
	items := make(map[int32]*proto.UIItem, len(ids))
	for _, id := range ids {
		items[id] = &proto.UIItem{Id: id}
	}
	return items
}

func TestStampItemFirstSeenKeepsKnownStampsAndStampsNewItems(t *testing.T) {
	const earlier, now = int64(100), int64(200)
	known := []ItemFirstSeen{{ID: 3, FirstSeen: earlier}, {ID: 1, FirstSeen: earlier}}
	got := StampItemFirstSeen(known, itemsWithIds(1, 2, 3), now)
	want := []ItemFirstSeen{{ID: 1, FirstSeen: earlier}, {ID: 2, FirstSeen: now}, {ID: 3, FirstSeen: earlier}}
	if !slices.Equal(got, want) {
		t.Errorf("stamped entries = %v, want the known stamps kept, the new item stamped now, sorted by id: %v", got, want)
	}
}

func TestStampItemFirstSeenKeepsTheLineOfAnItemThatLeftTheDatabase(t *testing.T) {
	const earlier, now = int64(100), int64(200)
	known := []ItemFirstSeen{{ID: 1, FirstSeen: earlier}, {ID: 2, FirstSeen: earlier}}
	got := StampItemFirstSeen(known, itemsWithIds(1), now)
	if !slices.Equal(got, known) {
		t.Errorf("stamped entries = %v, want every known entry kept: %v", got, known)
	}
}

func TestItemFirstSeenRoundTripsByteIdentical(t *testing.T) {
	path := filepath.Join(t.TempDir(), "item_first_seen.json")
	entries := []ItemFirstSeen{{ID: 1, FirstSeen: 100}, {ID: 2, FirstSeen: 200}}
	WriteItemFirstSeen(path, entries)
	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "[\n{\"id\":1,\"firstSeen\":100},\n{\"id\":2,\"firstSeen\":200}\n]\n"
	if string(written) != want {
		t.Errorf("written file = %q, want one entry per line: %q", written, want)
	}
	if got := ReadItemFirstSeen(path); !slices.Equal(got, entries) {
		t.Errorf("read back %v, want %v", got, entries)
	}
	WriteItemFirstSeen(path, ReadItemFirstSeen(path))
	if rewritten, _ := os.ReadFile(path); !bytes.Equal(rewritten, written) {
		t.Errorf("rewriting what was read changed the file: %q, want %q", rewritten, written)
	}
}

func TestItemFirstSeenReadsAMissingFileAsEmpty(t *testing.T) {
	if got := ReadItemFirstSeen(filepath.Join(t.TempDir(), "missing.json")); len(got) != 0 {
		t.Errorf("a missing file read as %v, want no entries", got)
	}
}

func TestCommittedItemFirstSeenCoversEveryItemInCanonicalForm(t *testing.T) {
	const path = "../../assets/database/item_first_seen.json"
	dbJson, err := os.ReadFile("../../assets/database/db.json")
	if err != nil {
		t.Fatal(err)
	}
	committed, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	entries := ReadItemFirstSeen(path)
	stamped := make(map[int32]bool, len(entries))
	for i, entry := range entries {
		if i > 0 && entry.ID <= entries[i-1].ID {
			t.Errorf("item_first_seen.json lists item %d after item %d, want ids strictly ascending", entry.ID, entries[i-1].ID)
		}
		stamped[entry.ID] = true
	}
	for id := range ReadDatabaseFromJson(string(dbJson)).Items {
		if !stamped[id] {
			t.Errorf("item %d in db.json has no entry in item_first_seen.json; run gen_db -gen=db", id)
		}
	}

	rewritten := filepath.Join(t.TempDir(), "item_first_seen.json")
	WriteItemFirstSeen(rewritten, entries)
	if got, _ := os.ReadFile(rewritten); !bytes.Equal(got, committed) {
		t.Errorf("item_first_seen.json is not in the form gen_db writes, one entry per line")
	}
}

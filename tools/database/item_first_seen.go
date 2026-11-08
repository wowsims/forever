package database

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"os"
	"slices"

	"github.com/wowsims/forever/sim/core/proto"
)

type ItemFirstSeen struct {
	ID        int32 `json:"id"`
	FirstSeen int64 `json:"firstSeen"` // Unix seconds of the gen_db run that first saw the item.
}

// A missing file reads as no entries.
func ReadItemFirstSeen(path string) []ItemFirstSeen {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		log.Fatalf("[ERROR] Failed to read %s: %s", path, err.Error())
	}
	var entries []ItemFirstSeen
	if err := json.Unmarshal(data, &entries); err != nil {
		log.Fatalf("[ERROR] Failed to parse %s: %s", path, err.Error())
	}
	return entries
}

// Every known entry stays, the ids of items no longer in the database too, so an item that returns keeps its first
// sighting and the file only ever gains lines.
func StampItemFirstSeen(known []ItemFirstSeen, items map[int32]*proto.UIItem, now int64) []ItemFirstSeen {
	entries := slices.Clone(known)
	seen := make(map[int32]bool, len(known))
	for _, entry := range known {
		seen[entry.ID] = true
	}
	for id := range items {
		if !seen[id] {
			entries = append(entries, ItemFirstSeen{ID: id, FirstSeen: now})
		}
	}
	slices.SortFunc(entries, func(a, b ItemFirstSeen) int { return int(a.ID) - int(b.ID) })
	return entries
}

// One entry per line, like db.json, so the file diffs by line.
func WriteItemFirstSeen(path string, entries []ItemFirstSeen) {
	var buf bytes.Buffer
	buf.WriteString("[\n")
	for i, entry := range entries {
		line, err := json.Marshal(entry)
		if err != nil {
			log.Fatalf("[ERROR] Failed to marshal %s: %s", path, err.Error())
		}
		buf.Write(line)
		if i < len(entries)-1 {
			buf.WriteByte(',')
		}
		buf.WriteByte('\n')
	}
	buf.WriteString("]\n")
	if existing, err := os.ReadFile(path); err == nil && bytes.Equal(existing, buf.Bytes()) {
		log.Printf("No changes detected, skipping write of %s", path)
		return
	}
	if err := os.WriteFile(path, buf.Bytes(), 0666); err != nil {
		log.Fatalf("[ERROR] Failed to write %s: %s", path, err.Error())
	}
}

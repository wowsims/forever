package lsp

import (
	"strings"
	"sync"
)

// The game files' row of a spell, as `go run ./tools/spelldata <id>` prints it, for a hover, kept by id. The rows
// come from the server's own spell store, the same for every repository it serves.
type cards struct {
	read func(spellID int32) (string, bool)
	mu   sync.Mutex
	rows map[int32]string
}

func newCards(read func(spellID int32) (string, bool)) *cards {
	return &cards{read: read, rows: map[int32]string{}}
}

// The row of a spell, or false for an id the store does not carry.
func (c *cards) row(spellID int32) (string, bool) {
	c.mu.Lock()
	row, ok := c.rows[spellID]
	c.mu.Unlock()
	if ok {
		return row, true
	}
	row, ok = c.read(spellID)
	if !ok {
		return "", false
	}
	row = strings.TrimSpace(row)
	c.mu.Lock()
	c.rows[spellID] = row
	c.mu.Unlock()
	return row, true
}

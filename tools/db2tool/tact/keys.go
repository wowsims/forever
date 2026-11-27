package tact

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

// The community-maintained key list (https://github.com/wowdev/TACTKeys):
// one "KEYNAME KEY" pair per line, the name as 16 hex digits and the key as
// 32, optionally followed by a description.
const defaultKeyFileURL = "https://raw.githubusercontent.com/wowdev/TACTKeys/master/WoW.txt"

// KeyStore holds TACT keys by their 64-bit lookup name — the value a BLTE
// 'E' chunk and a WDC5 section header both carry to say which key they
// need. Keys reach it from three places: the game files' own TactKey.db2,
// TACTKEY hotfix records the server pushed into DBCache.bin, and the
// community list. It also remembers every name a chunk asked for and did
// not find, so the extractor can say what is still locked.
type KeyStore struct {
	keys    map[uint64][16]byte
	mu      sync.Mutex // Guards missing, which decryption writes while several files may be opening at once.
	missing map[uint64]struct{}
}

func NewKeyStore() *KeyStore {
	return &KeyStore{keys: make(map[uint64][16]byte), missing: make(map[uint64]struct{})}
}

func (s *KeyStore) Add(name uint64, key [16]byte) { s.keys[name] = key }

func (s *KeyStore) Len() int { return len(s.keys) }

func (s *KeyStore) Has(name uint64) bool {
	_, ok := s.keys[name]
	return ok
}

// Missing lists the key names decryption asked for and did not have, sorted.
func (s *KeyStore) Missing() []uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]uint64, 0, len(s.missing))
	for n := range s.missing {
		out = append(out, n)
	}
	slices.Sort(out)
	return out
}

func (s *KeyStore) lookup(name uint64) ([16]byte, bool) {
	if s == nil {
		return [16]byte{}, false
	}
	k, ok := s.keys[name]
	if !ok {
		s.mu.Lock()
		s.missing[name] = struct{}{}
		s.mu.Unlock()
	}
	return k, ok
}

// LoadKeyFile adds every key in a community-format file. Malformed lines are
// skipped rather than fatal, since the list is hand-maintained.
func (s *KeyStore) LoadKeyFile(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	added := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 2 || strings.HasPrefix(fields[0], "#") || len(fields[0]) != 16 || len(fields[1]) != 32 {
			continue
		}
		name, err := strconv.ParseUint(fields[0], 16, 64)
		if err != nil {
			continue
		}
		raw, err := hex.DecodeString(fields[1])
		if err != nil {
			continue
		}
		var key [16]byte
		copy(key[:], raw)
		if !s.Has(name) {
			added++
		}
		s.Add(name, key)
	}
	return added, sc.Err()
}

// RefreshKeyFile keeps a local copy of the community list current with the
// same download-if-stale rule as the listfile: a usable existing file is
// never lost to a network failure, and only a missing file makes the
// download mandatory.
func RefreshKeyFile(path, url string) error {
	if url == "" {
		url = defaultKeyFileURL
	}
	info, statErr := os.Stat(path)
	if statErr == nil {
		if time.Since(info.ModTime()) < 24*time.Hour {
			return nil
		}
		if err := downloadFile(url, path); err != nil {
			fmt.Fprintf(os.Stderr, "db2tool: TACT key list refresh failed (%v), using existing %s\n", err, path)
		}
		return nil
	}
	if err := downloadFile(url, path); err != nil {
		return fmt.Errorf("downloading TACT key list: %w", err)
	}
	return nil
}

// KeyName formats a lookup the way the community list and wow.tools spell it.
func KeyName(name uint64) string {
	return fmt.Sprintf("%016X", name)
}

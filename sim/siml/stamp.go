package siml

import (
	"fmt"
	"os"
	"slices"
	"sync"
	"time"
)

// A file's size and modification time, which say it changed since a read. A missing file's size is -1.
type FileStamp struct {
	Size int64
	Time time.Time
}

func StampOf(path string) FileStamp {
	info, err := os.Stat(path)
	if err != nil {
		return FileStamp{Size: -1}
	}
	return FileStamp{info.Size(), info.ModTime()}
}

// A value read from files, read again when one of them changes on disk. Callers of one Reread wait for its read;
// those of another do not.
type Reread[T any] struct {
	mu     sync.Mutex
	stamps []FileStamp
	value  T
	err    error
}

// The value read from paths, by read when one of them changed since the last read.
func (r *Reread[T]) Get(paths []string, read func() (T, error)) (T, error) {
	stamps := make([]FileStamp, len(paths))
	for i, p := range paths {
		stamps[i] = StampOf(p)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.stamps == nil || !slices.Equal(r.stamps, stamps) {
		r.stamps = stamps
		r.value, r.err = read()
	}
	return r.value, r.err
}

// The name table at Path, read again when the file changes on disk.
type NamesFile struct {
	Path string
	held Reread[*Names]
}

func (n *NamesFile) Get() (*Names, error) {
	return n.held.Get([]string{n.Path}, func() (*Names, error) {
		data, err := ReadSource(n.Path)
		if err == nil {
			var names *Names
			if names, err = ParseNames(data); err == nil {
				return names, nil
			}
		}
		return nil, fmt.Errorf("%s does not load: %w", n.Path, err)
	})
}

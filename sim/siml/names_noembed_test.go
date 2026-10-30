//go:build !with_db

package siml

import (
	"strings"
	"testing"
)

func TestDefaultNamesNeedsTheTag(t *testing.T) {
	if _, err := DefaultNames(); err == nil || !strings.Contains(err.Error(), "with_db") {
		t.Errorf("DefaultNames gave %v, want an error naming the with_db tag", err)
	}
}

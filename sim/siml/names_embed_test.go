//go:build with_db

package siml

import (
	"bytes"
	"os"
	"testing"
)

func TestDefaultNamesIsTheCommittedTable(t *testing.T) {
	names, err := DefaultNames()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("names.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(names.Encode(), data) {
		t.Error("the built-in table does not encode to names.json")
	}
}

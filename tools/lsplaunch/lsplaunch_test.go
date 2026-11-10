package lsplaunch

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProtoStale(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"proto", filepath.Dir(generatedProto)} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	source := filepath.Join(root, "proto", "api.proto")
	generated := filepath.Join(root, generatedProto)
	if err := os.WriteFile(source, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	expect := func(when string, want bool) {
		t.Helper()
		got, err := protoStale(root)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("%s: stale %v, expected %v", when, got, want)
		}
	}

	expect("without the generated code", true)

	if err := os.WriteFile(generated, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	if err := os.Chtimes(source, now.Add(-time.Hour), now.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	expect("with the generated code newer than every .proto file", false)

	if err := os.Chtimes(source, now.Add(time.Hour), now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	expect("with a .proto file newer than the generated code", true)
}

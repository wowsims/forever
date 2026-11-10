//go:build with_db

package siml

import (
	"strings"
	"testing"
)

// The example rotation of llm_head.txt reads with the sim's names.
func TestLLMExampleReads(t *testing.T) {
	_, rest, ok := strings.Cut(llmHead, "## Rotations\n\n```\n")
	example, _, closed := strings.Cut(rest, "```")
	if !ok || !closed {
		t.Fatal("llm_head.txt has no example under ## Rotations")
	}
	if _, err := Parse("example.apl.siml", []byte(example), langNames(t), nil); err != nil {
		t.Errorf("the example rotation does not read:\n%v", err)
	}
}

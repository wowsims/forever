package siml

import (
	"os"
	"path/filepath"
	"testing"
)

// The wowsims-siml skill lists the kinds the protos give, so a change to apl.proto leaves it stale until it is written again.
func TestLLMTextIsCurrent(t *testing.T) {
	want, err := LLMText()
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join("..", "..", ".github", "skills", "wowsims-siml", "SKILL.md"))
	if err != nil || string(got) != want {
		t.Errorf("the wowsims-siml skill is not what the schema writes (%v); run make siml, or go run ./cmd/wowsimcli siml llm -o .github/skills/wowsims-siml/SKILL.md", err)
	}
}

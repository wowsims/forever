package cmd

import (
	"io"

	"github.com/wowsims/forever/sim/siml"
)

// The wowsims-siml skill, a short guide to SIML with every action, value and accessor the protos give:
//
//	go run ./cmd/wowsimcli siml llm -o .github/skills/wowsims-siml/SKILL.md
//
// With no -o it writes to stdout. make siml writes it.
func simlLLM(out, errOut io.Writer, args []string) error {
	flags := newFlags("llm [-o <file>]", errOut)
	output := flags.String("o", "", "the file to write; stdout when empty")
	if err := parseFlags(flags, args, 0); err != nil {
		return err
	}
	text, err := siml.LLMText()
	if err != nil {
		return err
	}
	return writeOutput(out, *output, []byte(text))
}

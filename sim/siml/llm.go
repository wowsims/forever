package siml

import (
	_ "embed"
	"fmt"
	"slices"
	"strings"
)

// The wowsims-siml skill (.github/skills/wowsims-siml/SKILL.md): a short guide to SIML for a language model or a
// newcomer, llm_head.txt, then every action, value and accessor the protos give, written from the schema, so a change
// to apl.proto shows up in it when it is written again (wowsimcli siml llm, which make siml runs).

//go:embed llm_head.txt
var llmHead string

// The text of the wowsims-siml skill.
func LLMText() (string, error) {
	ix, err := loadIndex()
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString(llmHead)
	b.WriteString("\n## Actions\n\nWritten from `proto/apl.proto`. Every kind also takes the generic `kind(field=value, …)`.\n\n")
	for _, k := range ix.actionKindList() {
		b.WriteString(ix.llmLine(k) + "\n")
	}
	b.WriteString("\n## Values\n\nWritten from `proto/apl.proto`, but for the accessors below.\n\n")
	for _, k := range ix.valueKindList {
		if k.Form != "accessor" {
			b.WriteString(ix.llmLine(k) + "\n")
		}
	}
	b.WriteString("\n## Accessors\n\nWritten `[unit.]head(subject).property(arguments)`; each head with its properties.\n\n")
	for _, head := range ix.heads {
		var props []string
		unit := ""
		for _, k := range ix.headKinds(head) {
			props = append(props, "`"+k.Accessor.Prop+ix.llmArgs(k, k.Accessor.Subject, k.Accessor.Unit)+"`")
			if k.Accessor.Unit != "" {
				unit = "[unit.]"
			}
		}
		b.WriteString(llmWrap(fmt.Sprintf("- `%s%s(x)`: %s", unit, head, strings.Join(props, ", "))))
	}
	return b.String(), nil
}

// One kind as a list item, with its fields.
func (ix *schemaIndex) llmLine(k *aplKind) string {
	line := k.Kind + ix.llmArgs(k)
	switch k.Form {
	case "bare":
		line = k.Kind
	case "unit_prefix":
		line = "[unit.]" + k.Kind + ix.llmArgs(k, k.UnitField)
	case "list_args":
		line = k.Kind + "(a, b, …)"
	}
	line = "- `" + line + "`"
	if k.Form == "short" {
		line += ", short form above"
	}
	return line
}

// A list item broken at spaces outside code into lines of at most 116 runes, as the hand-written skills wrap,
// continued under the item's text.
func llmWrap(item string) string {
	var b strings.Builder
	line, inCode, start := "", false, 0
	for i, r := range item {
		if r == '`' {
			inCode = !inCode
		}
		if r != ' ' || inCode || i+1 >= len(item) {
			continue
		}
		word := item[start:i]
		if line != "" && len([]rune(line+word)) > 116 {
			b.WriteString(strings.TrimRight(line, " ") + "\n")
			line = "  "
		}
		line += word + " "
		start = i + 1
	}
	word := item[start:]
	if line != "" && len([]rune(line+word)) > 116 {
		b.WriteString(strings.TrimRight(line, " ") + "\n")
		line = "  "
	}
	b.WriteString(line + word + "\n")
	return b.String()
}

// A kind's fields, but those in leave, as (name: type, …); "" for none. The positional field comes first, unnamed.
func (ix *schemaIndex) llmArgs(k *aplKind, leave ...string) string {
	var parts []string
	for _, f := range k.message.fields {
		if slices.Contains(leave, f.Name) {
			continue
		}
		if f.Name == k.Positional {
			parts = slices.Insert(parts, 0, ix.llmType(f))
			continue
		}
		parts = append(parts, f.Siml+": "+ix.llmType(f))
	}
	if len(parts) == 0 {
		return ""
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

// What a field takes, in a word or two.
func (ix *schemaIndex) llmType(f *field) string {
	t := f.Printed
	switch f.Printed {
	case "action_id":
		t = "name"
	case "enum", "flag_enum":
		if e := ix.enums[f.Type]; e != nil {
			var shorts []string
			for _, v := range e.values {
				if v.Short != "" && !strings.Contains(v.Short, "unknown") && !slices.Contains(shorts, v.Short) {
					shorts = append(shorts, v.Short)
				}
			}
			if len(shorts) > 8 {
				shorts = append(shorts[:8], "…")
			}
			t = strings.Join(shorts, "|")
		}
	case "message":
		t = "(…)"
	}
	if f.Ref != "" {
		t = f.Ref + " name"
	}
	if f.Repeated {
		t = "[" + t + ", …]"
	}
	return t
}

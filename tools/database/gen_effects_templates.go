package database

const TmplStrOnUse = `package forever

import (
{{- if .HasStacking }}
	"time"

{{ end }}
	"github.com/wowsims/forever/sim/common/shared"
{{- if .HasStacking }}
	"github.com/wowsims/forever/sim/core"
{{- end }}
)

func RegisterAllOnUseCds() {
{{- range .Groups }}

	// {{ .Name }}
{{- range .Entries }}
	{{- if .Skipped}}
	{{- range (.Tooltip | formatStrings 100) }}
	// Not simulated: {{.}}
	{{- end}}
	{{with index .Variants 0 -}}
	// https://www.wowhead.com/forever/spell={{.SpellID}}
	{{- end}}
	{{- else}}
	{{- if .StackingOnUse}}
  	{{- with index .Variants 0}}
	// {{ .Name }} - https://www.wowhead.com/forever/spell={{.SpellID}}
	{{- end}}
	shared.NewStackingStatBonusCD(shared.StackingStatBonusCD{
		Name:                  "{{ .StackingOnUse.Name }}",
		ID:                    {{ (index .Variants 0).ID }},
		CD:                    time.Millisecond * {{ .StackingOnUse.CooldownMs }},
		Callback:              {{ .StackProcInfo.Callback | asCoreCallback }},
		ProcMask:              {{ .StackProcInfo.ProcMask | asCoreProcMask }},
		Outcome:               {{ .StackProcInfo.Outcome | asCoreOutcome }},
		RequireDamageDealt:    {{ .StackProcInfo.RequireDamageDealt }},
		{{- if .StackProcInfo.RequireDamageComponent }}
		RequireDamageComponent: true,
		{{- end}}
		{{- if .StackProcInfo.CanProcFromProcs }}
		CanProcFromProcs:      true,
		{{- end}}
		{{- if .StackProcInfo.HonoursWeaponProcSuppression }}
		SpellFlagsExclude:     core.SpellFlagSuppressWeaponProcs,
		{{- end}}
		TrinketLimitsDuration: true,
	})
	{{- else if .Proc}}
	{{- $call := .Proc.OnUseConstructor }}
	{{- if .Proc.Summary }}
	// {{ .Proc.Summary }}
	{{- end}}
	{{- if .Supported}}
  	{{- with index .Variants 0}}
	shared.{{ $call }}({{ .ID }}) // {{ .Name }} - https://www.wowhead.com/forever/spell={{.SpellID}}
	{{- end}}
	{{- else}}
	// unsupported: {{ .Proc.Reason }}
  	{{- with index .Variants 0}}
	// shared.{{ $call }}({{ .ID }}) // {{ .Name }} - https://www.wowhead.com/forever/spell={{.SpellID}}
	{{- end}}
	{{- end}}
	{{- else if not .Supported}}
  	{{- with index .Variants 0}}
	// shared.NewSimpleStatActive({{ .ID }}) // {{ .Name }} - https://www.wowhead.com/forever/spell={{.SpellID}}
	{{- end}}
	{{- else}}
	{{- if .NotSimulated}}
	// not simulated: {{ .NotSimulated }}
	{{- end}}
  	{{- with index .Variants 0}}
	shared.NewSimpleStatActive({{ .ID }}) // {{ .Name }} - https://www.wowhead.com/forever/spell={{.SpellID}}
	{{- end}}
	{{- end}}
	{{- end}}
{{- end }}

{{- end }}
}`
const TmplStrProc = `package forever
{{ if .HasLive }}
import (
{{- if .UsesCore }}
	"github.com/wowsims/forever/sim/core"
{{- end }}
 	"github.com/wowsims/forever/sim/common/shared"
)
{{- end }}

func RegisterAllProcs() {
{{- range .Groups }}

	// {{ .Name }}
{{- range .Entries }}
	{{- if .Skipped}}
	{{- range (.Tooltip | formatStrings 100) }}
	// Not simulated: {{.}}
	{{- end}}
	{{with index .Variants 0 -}}
	// https://www.wowhead.com/forever/spell={{.SpellID}}
	{{- end}}
	{{- else}}
	{{if not .Supported}}
	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	{{- end}}
	{{- range (.Tooltip | formatStrings 100) }}
	// {{.}}
	{{- end}}
	{{with index .Variants 0 -}}
	// https://www.wowhead.com/forever/spell={{.SpellID}}
	{{- end}}
	{{- if .Supported}}
		{{- if .Proc}}
			// {{ .Proc.Summary }}
			shared.{{ .Proc.ProcConstructor }}(shared.SpellDataProc{TriggerSpellID: {{ .Proc.TriggerSpellID }}{{ if .Proc.BuffSpellID }}, BuffSpellID: {{ .Proc.BuffSpellID }}{{ end }}{{ if .Proc.IsWeaponProc }}, IsWeaponProc: true{{ end }}},
				[]shared.ItemVariant{
				{{- range .Variants }}
				{ItemID: {{.ID}}, ItemName: "{{.Name}}"},
				{{- end}}
			})
		{{- else if gt .ProcInfo.MaxCumulativeStacks 0 }}
			shared.NewStackingStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
				Callback:           {{ .ProcInfo.Callback | asCoreCallback }},
				ProcMask:           {{ .ProcInfo.ProcMask | asCoreProcMask }},
				Outcome:            {{ .ProcInfo.Outcome | asCoreOutcome }},
				RequireDamageDealt: {{ .ProcInfo.RequireDamageDealt }},
				{{- if .ProcInfo.RequireDamageComponent }}
				RequireDamageComponent: true,
				{{- end}}
				{{- if .ProcInfo.ClassSpellsOnly }}
				ClassSpellsOnly:    {{ .ProcInfo.ClassSpellsOnly }},
				{{- end}}
				{{- if .ProcInfo.CanProcFromProcs }}
				CanProcFromProcs:   true,
				{{- end}}
				{{- if .ProcInfo.IsWeaponProc }}
				IsWeaponProc:       true,
				{{- end}}
				{{- if .ProcInfo.HonoursWeaponProcSuppression }}
				SpellFlagsExclude:  core.SpellFlagSuppressWeaponProcs,
				{{- end}}
			{{- if .StackProcInfo }}
				StackCallback:      {{ .StackProcInfo.Callback | asCoreCallback }},
				StackProcMask:      {{ .StackProcInfo.ProcMask | asCoreProcMask }},
				StackOutcome:       {{ .StackProcInfo.Outcome | asCoreOutcome }},
			{{- end}}
			}, []shared.ItemVariant{
				{{- range .Variants }}
				{ItemID: {{.ID}}, ItemName: "{{.Name}}"},
				{{- end}}
			})
		{{- else}}
			shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
				Callback:           {{ .ProcInfo.Callback | asCoreCallback }},
				ProcMask:           {{ .ProcInfo.ProcMask | asCoreProcMask }},
				Outcome:            {{ .ProcInfo.Outcome | asCoreOutcome }},
				RequireDamageDealt: {{ .ProcInfo.RequireDamageDealt }},
				{{- if .ProcInfo.RequireDamageComponent }}
				RequireDamageComponent: true,
				{{- end}}
				{{- if .ProcInfo.ClassSpellsOnly }}
				ClassSpellsOnly:    {{ .ProcInfo.ClassSpellsOnly }},
				{{- end}}
				{{- if .ProcInfo.CanProcFromProcs }}
				CanProcFromProcs:   true,
				{{- end}}
				{{- if .ProcInfo.IsWeaponProc }}
				IsWeaponProc:       true,
				{{- end}}
				{{- if .ProcInfo.HonoursWeaponProcSuppression }}
				SpellFlagsExclude:  core.SpellFlagSuppressWeaponProcs,
				{{- end}}
			{{- if .StackProcInfo }}
				StackCallback:      {{ .StackProcInfo.Callback | asCoreCallback }},
				StackProcMask:      {{ .StackProcInfo.ProcMask | asCoreProcMask }},
				StackOutcome:       {{ .StackProcInfo.Outcome | asCoreOutcome }},
			{{- end}}
			}, []shared.ItemVariant{
				{{- range .Variants }}
				{ItemID: {{.ID}}, ItemName: "{{.Name}}"},
				{{- end}}
			})
		{{- end}}
	{{- else}}
		{{- if .Proc}}
			// unsupported: {{ .Proc.Reason }}
			// {{ .Proc.Summary }}
			// shared.{{ .Proc.ProcConstructor }}(shared.SpellDataProc{TriggerSpellID: {{ .Proc.TriggerSpellID }}{{ if .Proc.BuffSpellID }}, BuffSpellID: {{ .Proc.BuffSpellID }}{{ end }}{{ if .Proc.IsWeaponProc }}, IsWeaponProc: true{{ end }}},
			//	[]shared.ItemVariant{
			{{- range .Variants }}
			//	{ItemID: {{.ID}}, ItemName: "{{.Name}}"},
			{{- end}}
			// })
		{{- else if gt .ProcInfo.MaxCumulativeStacks 0 }}
			// shared.NewStackingStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
			//	Callback:           {{ .ProcInfo.Callback | asCoreCallback }},
			//	ProcMask:           {{ .ProcInfo.ProcMask | asCoreProcMask }},
			//	Outcome:            {{ .ProcInfo.Outcome | asCoreOutcome }},
			//	RequireDamageDealt: {{ .ProcInfo.RequireDamageDealt }},
			{{- if .ProcInfo.ClassSpellsOnly }}
			//	ClassSpellsOnly:    {{ .ProcInfo.ClassSpellsOnly }},
			{{- end}}
			// }, []shared.ItemVariant{
				{{- range .Variants }}
			//	{ItemID: {{.ID}}, ItemName: "{{.Name}}"},
				{{- end}}
			// })
		{{- else}}
			// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
			//	Callback:           {{ .ProcInfo.Callback | asCoreCallback }},
			//	ProcMask:           {{ .ProcInfo.ProcMask | asCoreProcMask }},
			//	Outcome:            {{ .ProcInfo.Outcome | asCoreOutcome }},
			//	RequireDamageDealt: {{ .ProcInfo.RequireDamageDealt }}
			{{- if .ProcInfo.ClassSpellsOnly }}
			// ClassSpellsOnly:    {{ .ProcInfo.ClassSpellsOnly }},
			{{- end}}
			// }, []shared.ItemVariant{
				{{- range .Variants }}
			//	{ItemID: {{.ID}}, ItemName: "{{.Name}}"},
				{{- end}}
			// })
		{{- end}}
	{{- end}}
	{{- end}}
{{- end }}

{{- end }}
}`

const TmplStrEnchant = `package forever
{{ if .HasLive }}
import (
{{- if .UsesCore }}
	"github.com/wowsims/forever/sim/core"
{{- end }}
 	"github.com/wowsims/forever/sim/common/shared"
)
{{- end }}

func RegisterAllEnchants() {
{{- range .Groups }}

	// {{ .Name }}
{{- range .Entries }}
	{{- if .Skipped}}
	{{- range (.Tooltip | formatStrings 100) }}
	// Not simulated: {{.}}
	{{- end}}
	{{with index .Variants 0 -}}
	// https://www.wowhead.com/forever/spell={{.SpellID}}
	{{- end}}
	{{- else}}
	{{if not .Supported}}
	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	{{- end}}
	{{- range (.Tooltip | formatStrings 100) }}
	// {{.}}
	{{- end}}
	{{with index .Variants 0 -}}
	// https://www.wowhead.com/forever/spell={{.SpellID}}
	{{- end}}
	{{- if .Proc}}
		{{- if .Supported}}
		// {{ .Proc.Summary }}
		shared.{{ .Proc.ProcConstructor }}(shared.SpellDataProc{
			{{- with index .Variants 0 }}
			Name:           "{{ .Name }}",
			EnchantID:      {{ .ID }},
			{{- end}}
			TriggerSpellID: {{ .Proc.TriggerSpellID }},
			{{- if .Proc.BuffSpellID }}
			BuffSpellID:    {{ .Proc.BuffSpellID }},
			{{- end}}
			{{- if .Proc.IsWeaponProc }}
			IsWeaponProc:   true,
			{{- end}}
		}, nil)
		{{- else}}
		// unsupported: {{ .Proc.Reason }}
		// {{ .Proc.Summary }}
		// shared.{{ .Proc.ProcConstructor }}(shared.SpellDataProc{
		{{- with index .Variants 0 }}
		//	Name:           "{{ .Name }}",
		//	EnchantID:      {{ .ID }},
		{{- end}}
		//	TriggerSpellID: {{ .Proc.TriggerSpellID }},
		{{- if .Proc.BuffSpellID }}
		//	BuffSpellID:    {{ .Proc.BuffSpellID }},
		{{- end}}
		{{- if .Proc.IsWeaponProc }}
		//	IsWeaponProc:   true,
		{{- end}}
		// }, nil)
		{{- end}}
	{{- else if .Supported}}
		shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
			{{with index .Variants 0 -}}
			Name:               "{{ .Name }}",
			EnchantID:          {{ .ID }},
			{{- end}}
			Callback:           {{ .ProcInfo.Callback | asCoreCallback }},
			ProcMask:           {{ .ProcInfo.ProcMask | asCoreProcMask }},
			Outcome:            {{ .ProcInfo.Outcome | asCoreOutcome }},
			RequireDamageDealt: {{ .ProcInfo.RequireDamageDealt }},
			{{- if .ProcInfo.RequireDamageComponent }}
			RequireDamageComponent: true,
			{{- end}}
			{{- if .ProcInfo.ClassSpellsOnly }}
			ClassSpellsOnly:    {{ .ProcInfo.ClassSpellsOnly }},
			{{- end}}
			{{- if .ProcInfo.CanProcFromProcs }}
			CanProcFromProcs:   true,
			{{- end}}
			{{- if .ProcInfo.IsWeaponProc }}
			IsWeaponProc:       true,
			{{- end}}
			{{- if .ProcInfo.HonoursWeaponProcSuppression }}
			SpellFlagsExclude:  core.SpellFlagSuppressWeaponProcs,
			{{- end}}
		})
	{{- else}}
		// shared.NewProcStatBonusEffect(shared.ProcStatBonusEffect{
		{{- with index .Variants 0 }}
		//	Name:               "{{ .Name }}",
		//	EnchantID:          {{ .ID }},
		{{- end}}
		//	Callback:           {{ .ProcInfo.Callback | asCoreCallback }},
		//	ProcMask:           {{ .ProcInfo.ProcMask | asCoreProcMask }},
		//	Outcome:            {{ .ProcInfo.Outcome | asCoreOutcome }},
		//	RequireDamageDealt: {{ .ProcInfo.RequireDamageDealt }},
		{{- if .ProcInfo.ClassSpellsOnly }}
		//  ClassSpellsOnly:    {{ .ProcInfo.ClassSpellsOnly }},
		{{- end}}
		// })
	{{- end}}
	{{- end}}
{{- end }}

{{- end }}
}`

const TmplStrMissingEffects = `
// This file is auto generated
// Changes will be overwritten on next database generation

export const MISSING_ITEM_EFFECTS = new Map<number, string[]>([
{{- range .ItemEffects }}
	[
		{{.ItemID}}, // {{ .Name }}
		[
			{{- range .Effects }}
			"{{ jsString .Name }}", // {{.SpellID}} - https://www.wowhead.com/forever/spell={{.SpellID}}
			{{- end}}
		]
	],
{{- end }}
])

export const MISSING_ENCHANT_EFFECTS = new Map<number, string[]>([
{{- range .EnchantEffects }}
	[
		{{.ItemID}}, // {{ .Name }}
		[
			{{- range .Effects }}
			"{{ jsString .Name }}", // {{.SpellID}} - https://www.wowhead.com/forever/spell={{.SpellID}}
			{{- end}}
		]
	],
{{- end }}
])

// Keyed by set id, then by the bonus's piece count; the tooltips of the spells the sim leaves out.
export const MISSING_SET_BONUSES = new Map<number, Map<number, string[]>>([
{{- range .MissingSetBonuses }}
	[
		{{(index . 0).ID}}, // {{ (index . 0).Name }}
		new Map([
			{{- range $bonus := . }}
			[
				{{ $bonus.Pieces }},
				[
					{{- range $i, $tooltip := $bonus.Tooltips }}
					"{{ jsString $tooltip }}", // {{ index $bonus.Spells $i }} - https://www.wowhead.com/forever/spell={{ index $bonus.Spells $i }}
					{{- end }}
				],
			],
			{{- end }}
		]),
	],
{{- end }}
])

// A proc rate the game files do not state and nobody has measured on Forever, and where it comes from.
export type UnvalidatedRate = { kind: 'ppm' | 'chance'; origin: 'vanilla' | 'tbc' | 'unknown' }

export const UNVALIDATED_RATE_ITEMS = new Map<number, UnvalidatedRate>([
{{- range .UnvalidatedRateItems }}
	[{{.ID}}, { kind: '{{.Kind}}', origin: '{{.Origin}}' }], // {{ .Name }}
{{- end }}
])

export const UNVALIDATED_RATE_ENCHANTS = new Map<number, UnvalidatedRate>([
{{- range .UnvalidatedRateEnchants }}
	[{{.ID}}, { kind: '{{.Kind}}', origin: '{{.Origin}}' }], // {{ .Name }}
{{- end }}
])

// Keyed by set id, then by the bonus's piece count.
export const UNVALIDATED_RATE_SET_BONUSES = new Map<number, Map<number, UnvalidatedRate>>([
{{- range .UnvalidatedRateSetBonuses }}
	[
		{{(index . 0).ID}}, // {{ (index . 0).Name }}
		new Map([
			{{- range . }}
			[{{.Pieces}}, { kind: '{{.Kind}}', origin: '{{.Origin}}' }],
			{{- end }}
		]),
	],
{{- end }}
])
`

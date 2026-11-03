# WoWSims SIML for Zed

SIML files (`*.siml`, `sim/siml/SPEC.md`) in Zed. The tree-sitter grammar of `tools/tree-sitter-siml` colours,
indents and outlines them, and the SIML language server (`wowsimcli siml lsp`, `sim/siml/lsp`) checks, completes,
hovers, navigates, renames and formats them, as it does in VS Code (`tools/vscode-apl`).

## Install

Zed builds a dev extension itself, which needs Rust installed through `rustup`:

1. Run `zed: install dev extension` from the command palette and pick this folder, `tools/zed-siml`. Zed compiles
   `src/lib.rs` and fetches the grammar from the commit `extension.toml` names (`[grammars.siml]`).
2. Open the repository root as the worktree, not a file in it: the server runs from there.

The extension starts the server with `go -C <worktree root> run ./sim/siml/lsp/launch`, which regenerates the protos
when the `.proto` files are newer, builds `wowsimcli` with the item database and runs `siml lsp`. The first start waits
for that build; `go` has to be on the PATH of the shell Zed starts in. Where the build fails, Zed shows the error as
a notification and keeps the server, so _editor: restart language server_ after the fix starts it. Restart it after
pulling Go changes too. A fresh `make siml-names` or `make db` needs no restart.

The repository's `.zed/settings.json` turns the server's semantic tokens on for SIML (`"semantic_tokens":
"combined"`): the grammar colours the shapes, and the server's tokens colour what only the schema and the name table
know on top, such as a spell, an aura, an item, an enum value or a unit. `languages/siml/semantic_token_rules.json`
maps them to the theme.

The extension starts the server with the initialization option `simCommands`, which turns the sims on. A sim reads
the spec's `defaults.siml` and the one preset build that uses the rotation, as VS Code's does; Zed's settings may pass
more, which the extension adds to its own:

```json
{
	"lsp": {
		"wowsims-siml": {
			"initialization_options": { "simIterations": 5000, "simSettingsFile": "/path/to/exported.json" }
		}
	}
}
```

A change to `src/lib.rs` or `extension.toml` takes effect after _zed: install dev extension_ again; a change to a
query, `config.toml` or a snippet after the extension is rebuilt from the Extensions page.

## What it does

- **Colours.** Keywords, sections, kinds and defs, accessor heads and properties, units, fields, variables,
  placeholders, numbers with their units, strings and raw JSON.
- **Outline.** Sections, groups, variables, defs, sequences and `when` blocks, in _outline: toggle_, the outline
  panel and the breadcrumbs.
- **Indentation.** The line after a header or a `when` line goes in a level, as do the lines inside brackets.
- **Brackets.** `( )`, `[ ]`, quotes and the backticks of raw JSON match and close, and the brackets take rainbow
  colours where `colorize_brackets` is on.
- **Comments.** _editor: toggle comments_ (`gc` in vim mode) writes and removes `// `.
- **Vim text objects.** `af`/`if` a def and its value, `af` a `when` block, `ac` a section.
- **Snippets.** `siml` a rotation file, `when`, `group`, `sequence`, `def`, `var`, and the shapes of VS Code's Insert
  APL Pattern: `dot`, `debuff`, `buff`, `during`, `lineup`, `execute`, `notexecute`, `resource`, `toggle`,
  `potion`, `prepull`.
- **The language server.** Diagnostics, completion, hover (a spell's row of the game files among it), go to
  definition, find references, rename, formatting (_editor: format_), code actions (`ctrl-.`), and links on `use`
  paths.
- **Sims.** In an `.apl.siml`, `ctrl-.` offers _Sim this rotation_ and _Compare this rotation with the saved file, or
  HEAD when nothing is unsaved_, as VS Code's Sim and Compare. A notification gives the DPS, and each row shows how
  often it ran per fight at the end of its line until the text changes. The repository's `.zed/settings.json` turns
  inlay hints on for SIML, which the counts are. With `"code_lens": "on"` in Zed's settings (a setting for every
  language) the same two commands also sit above the first line, the last result in their titles.

Zed extensions draw no panels, so the VS Code panel (the rotation as a tree and its forms) has no counterpart here. Raw JSON is not coloured as JSON: Zed lists an injected language's outline in the file's, and JSON's
would list every key of a raw row.

## The grammar

`extension.toml` fetches the grammar from a commit of this repository, `rev` under `[grammars.siml]`. After a change
to `tools/tree-sitter-siml` lands, set `rev` to the commit that made it (its README says more); `npm run check` there
fails while `rev` names another grammar.

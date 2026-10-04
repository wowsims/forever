# tools/spelldata

Reads the spell store, `sim/core/spelldata`, from the command line and from an editor. Run it from the
repository root.

```
go run ./tools/spelldata 11574                  # one row: header, ladder call, each effect worded and literal
go run ./tools/spelldata Whirlwind [-all]       # every row with that name, then the highest rank (or all)
go run ./tools/spelldata 11574 -json            # the same as JSON
go run ./tools/spelldata -family warrior/Execute
go run ./tools/spelldata -expr 'spellData.Execute.Highest().EffectN(1).Average(core.CharacterLevel)' -package warrior
go run ./tools/spelldata -config 'spelldata.SpellConfig(&warrior.Unit, executeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))' -package warrior
go run ./tools/spelldata -config 'core.SpellModConfig{Kind: core.SpellMod_BonusCrit_Percent, ProcMask: core.ProcMaskMelee, EquipSubclass: dbcenums.ITEM_SUBCLASS_BIT_WEAPON_1H_AXE | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_2H_AXE}'
go run ./tools/spelldata -hover sim/warrior/execute.go 12:40   # the markdown an editor shows at line:column (1-based)
go run ./tools/spelldata -lsp                   # a language server on stdio
```

`-hover` prints the hover's markdown on stdout and its trace on stderr, and exits 1 where there is no
hover. `-config` resolves a `spelldata.SpellConfig` call - the row pick substituted through the
package's declarations, each option applied - and states each field with the step that filled it. Given
a `core.SpellModConfig` literal it states each field it can read, the weapon its Equip fields require
and the rule core applies to it: per weapon, or ranged weapons only; magic-defense spells and pets never
match; an item swap re-filters it; `FromRating` counts on the sheet as rating.
`-expr` reads a chain the way a class file writes it: a family's `spelldata.Ladder`, then any accessor
of the store (`Highest()`, `Rank(n)`, `EffectAt(1).TenthsAt(1)`, `Effect(dbcenums.A_X, misc)`, ...),
with arguments that are literals or constants of `dbcenums`, of `stats` (`stats.HitRating`) or of
`sim/core`'s `flags.go`, `constants.go`, `base_stats_auto_gen.go` (`core.HitRatingPerHitPercent`) and
`spell_mod.go` (`core.SpellMod_BonusCrit_Percent`).

A hit, crit or rating effect names the rule `spelldata.WeaponScope` gives it, which is the one the parse
attaches it by: `per weapon: only attacks made with <weapons>`, `ranged weapons only`,
`while <weapons> is equipped: every spell and attack`, `the weapon requirement only gates casting` for an
aura on the enemy, and for a rating the ratings `spelldata.RatingStats` reads (the expertise bit among
them) or, on a scoped row, the percent it is worth and the ratings it leaves unread.

## JSON output

`-json` states the same reading as data, for a script or an AI agent that calls the tool. A row is one
object:

```json
{
	"id": 11574,
	"name": "Rend",
	"rank": "Rank 7",
	"ladder": ["warrior spellData.Rend.Highest()"],
	"header": [
		{ "key": "school", "value": "physical" },
		{ "key": "cost", "value": "10 rage" }
	],
	"effects": [
		{
			"human": "21 physical damage every 3 s to the enemy (7 ticks)",
			"literal": "E_APPLY_AURA A_PERIODIC_DAMAGE base=21 period=3000ms target=[6,0]"
		}
	],
	"wowhead": "https://www.wowhead.com/forever/spell=11574"
}
```

- `rank` is the store's rank column, or `rank n of N` on a talent's rank.
- `ladder` holds the class-file calls that reach the id; empty where no class file names it.
- `header` holds the columns the row states, in the order the text form prints them. It is a list
  because a key can repeat (`cost`, `cooldown`); `proc` and `refs` are keys like the rest.
- `effects[].human` is `unrecognised shape` where no wording applies, and `literal` is always the
  game files' own columns. `read: true` marks the effect an `-expr` chain read.

By mode:

- `<id> -json`: one row.
- `<name> -json [-all]`: an array of rows, the highest rank or every match.
- `-family <class>/<Family> -json`: `{"family", "ranks": [{"id", "name", "rank", "accessor", "value"}], "highest": <row>}`.
  `value` (`effect 1 = 600`) is there only where effect 1 changes by rank.
- `-expr <chain> -json`: the row the chain reached, plus `kind` (`spell`, `effect` or `value`), `trail`
  (the chain with every constant resolved), `value`, `doc` (the doc comment of the accessor that
  answered, where it has one) and, on an effect, `accessors` (`Name(args) = value` for each accessor
  that reads something off it).
- `-config` and `-hover` have no JSON form.

## The language server

`-lsp` speaks JSON-RPC 2.0 over stdio with Content-Length framing.

| method                                                           |                                                                                                               |
| ---------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| `initialize`                                                     | `hoverProvider`, incremental `textDocumentSync`; `initializationOptions.trace` is `"on"` (default) or `"off"` |
| `initialized`, `$/cancelRequest`, `didSave`, other notifications | ignored                                                                                                       |
| `textDocument/didOpen`, `didChange`                              | keeps the buffer, so a hover reads unsaved edits; the next hover re-reads that file's declarations            |
| `textDocument/didClose`                                          | drops the buffer; the next hover re-reads the file from disk                                                  |
| `textDocument/hover`                                             | markdown `MarkupContent`, or `null`                                                                           |
| `shutdown`, `exit`                                               | exits 0 after a `shutdown`, 1 without one                                                                     |
| any other request                                                | `MethodNotFound`                                                                                              |

Every hover sends its trace as `window/logMessage` of type Log: the position, what matched (`id`,
`family`, `ident`, `segment`, `SpellConfig`, `SpellModConfig`), each substitution with the file and line it came from,
the declaration cache hit or miss, and the result (`✓ 20662 Execute (Rank 5) → effect 1 → 600`) or why
it stopped (`✗ ...`).

What a hover reads: a spell id in the shapes hand-written code states (`MustFind(n)`, `Find(n)`,
`.ByID(n)`, `SpellID: n`, `"spellId": n`, `fromSpellId(n)`, `spellId: n`) in any file; in a Go file
also a `spellData.<Family>` token, a name the package folder binds to a ladder chain with `var`, `=` or
`:=` at package level, or the hovered function binds before the cursor (substituted through at most
four names, cycles refused), one accessor of a chain anywhere in an expression,
`spelldata.SpellConfig`, and the type name of a `core.SpellModConfig` literal. A file the editor does not hold is read again when its modification time
moves. The server compiles the store in; restart it after regenerating.

## Editors

- **VS Code**: `.vscode/extensions/wowsims-spelldata`, a workspace extension built from
  `tools/vscode-spelldata` by `make vscode-spelldata` (or `npm run build:lsp` from the root, which also compiles the server).
- **Zed**: `tools/zed-spelldata`, installed with `zed: install dev extension`.

### Other editors

Neovim (0.10 or later), in `init.lua`:

```lua
vim.api.nvim_create_autocmd('FileType', {
	pattern = { 'go', 'json', 'typescript', 'typescriptreact' },
	callback = function(args)
		local root = vim.fs.root(args.buf, 'go.mod')
		if root and vim.uv.fs_stat(root .. '/tools/spelldata') then
			vim.lsp.start({
				name = 'wowsims-spelldata',
				cmd = { 'go', 'run', './tools/spelldata', '-lsp' },
				cmd_cwd = root,
				root_dir = root,
				init_options = { trace = 'on' },
			})
		end
	end,
})
```

Helix, in `languages.toml` (the repository's `.helix/languages.toml` or the user's):

```toml
[language-server.wowsims-spelldata]
command = "go"
args = ["run", "./tools/spelldata", "-lsp"]
config = { trace = "on" }

[[language]]
name = "go"
language-servers = ["gopls", "wowsims-spelldata"]

[[language]]
name = "json"
language-servers = ["vscode-json-language-server", "wowsims-spelldata"]

[[language]]
name = "typescript"
language-servers = ["typescript-language-server", "wowsims-spelldata"]

[[language]]
name = "tsx"
language-servers = ["typescript-language-server", "wowsims-spelldata"]
```

Helix starts the server in the root it finds for the file, which in this repository is its root - where
`go run ./tools/spelldata` has to run. Its hover shows every server's answer.

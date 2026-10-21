# WoWSims Spelldata hover

Hover a spell id in Go, JSON or TypeScript and read the game files' row: what the spell is, a table of
the header columns, and a table of the effects - each worded where the shape is known and always
over the game files' own literal.

The ids it recognises are the ones hand-written code states: `spelldata.MustFind(11574)`,
`spelldata.Find(116)`, `spellData.Rend.ByID(11574)`, `core.ActionID{SpellID: 11574}`, an APL file's
`"spellId": 11574`, `ActionId.fromSpellId(23563)` and a TS `spellId: 23563`.

In a Go file it also reads names. The family in `spellData.Execute` hovers as the whole ladder; a name
bound to a rank, `executeRank` in `var executeRank = spellData.Execute.Highest()`, hovers as that rank
wherever the package uses it; a name bound to a value read off a rank, `executeBaseDamage` in
`var executeBaseDamage = executeRank.EffectN(1).Average(core.CharacterLevel)`, hovers as the value
(**600**), the chain with every name substituted, and the row with the effect it read marked `▶`. On a
declaration line each accessor answers for itself: `EffectN(1)` the effect and every accessor that reads
something off it, `Average(core.CharacterLevel)` the value and the doc comment `sim/core/spelldata`
writes on that accessor. The cursor on `SpellConfig` in a `spelldata.SpellConfig(...)` call shows the
config the resolver builds from the row and the options, each field with the step that filled it; on
`SpellModConfig` in a `core.SpellModConfig{...}` literal, the fields it can read and the weapon its
Equip fields require. A hit, crit or rating effect names the weapon rule the parse applies to it. A
`Ladder.Each` callback parameter, `rank` in `func(_ int32, rank *spelldata.Spell)` and in the function
the callback passes it to, hovers as the highest rank, and the hover on the parameter links the other
ranks: a click hovers it as that rank until the server restarts.

## How it answers

This extension is a language client. It starts `go run ./tools/spelldata/launch -lsp` in the workspace
folder holding `go.mod`, which regenerates stale protos and builds the server, and the server does all
the reading: the id patterns, the ladder families, the names a package folder binds (read from the open
buffers first, then disk), the substitution and the markdown.
The first hover of a session waits for the server to build, a few seconds; the rest answer in-process.
The server compiles the store in, so restart it (_Developer: Restart Extension Host_, or reload the
window) after regenerating `sim/core/spelldata`.

Each hover logs how it was resolved - what matched, each substitution with the file and line it came
from, and the result or why it stopped - to the `WoWSims Spelldata` output channel.

## Settings

|                              |                                                                                                        |
| ---------------------------- | ------------------------------------------------------------------------------------------------------ |
| `wowsims-spelldata.goBinary` | the `go` binary that starts the server; `go` looks on the PATH, then GOROOT and the usual install dirs |
| `wowsims-spelldata.trace`    | `on` (default) or `off`: whether each hover's trace goes to the output channel                         |

## Build and install

The source is `tools/vscode-spelldata`; `.vscode/extensions/wowsims-spelldata` is its build output
and is not edited by hand. From the repository root:

```
make vscode-spelldata                                   # npm ci, typecheck, bundle, write the manifest
cd tools/vscode-spelldata && npm run check && npm test  # the committed output matches a fresh build
```

VS Code (1.91 or later) treats `.vscode/extensions/wowsims-spelldata` as a workspace extension:
opening the repository shows an _Install Workspace Extension_ prompt once, and after that the hover
loads for this workspace only. `out/extension.js` is one esbuild bundle with `vscode-languageclient`
inside, so a fresh clone needs no `npm install` to use it.

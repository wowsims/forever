# WoWSims Spelldata for Zed

The spell hover of `tools/spelldata` in Zed: a spell id in Go, JSON, TypeScript or TSX, a ladder family,
a name bound to a ladder chain, a `spelldata.SpellConfig` call or a `core.SpellModConfig` literal hovers
as the game files' row, the value the chain reads, the config the resolver builds or the mod's fields and
the weapon they require; a code action picks the rank a `Ladder.Each` callback parameter hovers as. The
extension only starts the server,
`go -C <worktree root> run ./tools/spelldata/launch -lsp`, which regenerates stale protos and builds the
server first; everything else is in `tools/spelldata` (see its README).

## Install

Zed builds a dev extension itself, which needs Rust installed through `rustup`:

1. Run `zed: install dev extension` from the command palette.
2. Pick this folder, `tools/zed-spelldata`.

With Zed on Windows and the repository in WSL, copy this folder to a Windows drive first and pick the
copy: Zed links the installed extension to its folder with a junction, which cannot point at
`\\wsl.localhost`. Zed compiles it on Windows, so the toolchain there has to link on its own:
`rustup default stable-x86_64-pc-windows-gnu`, or the MSVC toolchain with Visual Studio's C++ build
tools. The server itself still runs in WSL, from the repository.

`.zed/settings.json` lists gopls before this server for Go. Zed shows each server's hover as a
separate popover.

Open the repository root as the worktree - the server is run from there and reads the store it
compiles. The first hover waits for `go run` to compile, a few seconds. `go` has to be on the PATH of
the shell Zed starts in.

A `Ladder.Each` callback parameter hovers as the highest rank of its ladder. VS Code's hover links the
other ranks; Zed cannot run those links, so the server offers the ranks as code actions wherever a hover
reads through the parameter: _editor: toggle code actions_ (`ctrl-.`) lists
`Hover rank as 19895 Shadow Resistance Aura (Rank 2)` next to gopls' actions. Hover again to read the
rank picked; the pick lasts until the server restarts.

The server compiles the store in, so after regenerating `sim/core/spelldata` or editing
`tools/spelldata`, run _editor: restart language server_ (it restarts gopls too). A restart also
regenerates `sim/core/proto` when a pull moved the `.proto` files. Where the server does not build, Zed
shows the build error as a notification and keeps the server, so a restart after the fix starts it.

A change to `src/lib.rs` takes effect after _zed: install dev extension_ again.

The server logs each hover's trace as a log message: open _dev: open language server logs_, pick
`wowsims-spelldata` and _Server Logs_ to read it. To silence the trace, set the server's initialization
option in Zed's settings:

```json
{
	"lsp": {
		"wowsims-spelldata": {
			"initialization_options": { "trace": "off" }
		}
	}
}
```

The extension is registered next to Zed's own servers for those languages, so gopls and the TypeScript
server keep answering too; Zed shows both hovers.

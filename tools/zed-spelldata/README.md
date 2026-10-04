# WoWSims Spelldata for Zed

The spell hover of `tools/spelldata` in Zed: a spell id in Go, JSON, TypeScript or TSX, a ladder family,
a name bound to a ladder chain, a `spelldata.SpellConfig` call or a `core.SpellModConfig` literal hovers
as the game files' row, the value the chain reads, the config the resolver builds or the mod's fields and
the weapon they require. The extension only starts the server,
`go -C <worktree root> run ./tools/spelldata -lsp`; everything else is in `tools/spelldata` (see its
README).

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

The server logs each hover's trace as a log message: open _debug: open language server logs_ and pick
`wowsims-spelldata` to read it. To silence it, set the server's initialization option in Zed's
settings:

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

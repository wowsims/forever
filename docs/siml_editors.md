# SIML in another editor

`tools/vscode-apl` is the VS Code client of SIML (`sim/siml/SPEC.md`). Every other editor needs the same three things,
and none of them is VS Code specific:

| Part                        | What it gives                                                                     | Where it comes from                                                                               |
| --------------------------- | --------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------- |
| File type and language data | `.siml` files, `//` comments, brackets, quotes                                    | the table below                                                                                   |
| Highlighting                | keywords, units, sections, strings, numbers                                       | the TextMate grammar, or the semantic tokens of the server in an editor that applies them         |
| The language server         | diagnostics, completion, hover, navigation, rename, formatting, code actions, ... | `wowsimcli siml lsp` over stdio, which finds the repository above each file and needs no settings |

The server is the only part with logic. Names, items, enchants, did-you-mean and the lints all come from it, so a
new editor never reimplements them; it starts the server and says which files it serves.

## The language data

|                 |                                                                                                                                                   |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| Language id     | `siml`                                                                                                                                            |
| Files           | `*.siml`; the kind is the double extension (`x.apl.siml`, `x.gear.siml`, `x.<section>.siml`), so an editor matches on the last `.siml` only       |
| Line comment    | `//`, and `///` for a note that stays with the next row; there are no block comments                                                              |
| Brackets        | `( )` and `[ ]`; `"` and `` ` `` close on themselves                                                                                              |
| Indentation     | the line after a `section:` header or a `group "name":` header is indented one level; nothing dedents by pattern                                  |
| Word characters | letters, digits and `_`, with a leading `$` or `%` for a variable or placeholder, and `${…}` or `%{…}` for a name with spaces; `2.5s` is one word |
| Scope name      | `source.siml`                                                                                                                                     |
| Server command  | `wowsimcli siml lsp`, with the repository root as its working directory when the editor lets it choose                                            |
| Server root     | the first folder above a file that holds `sim/siml/names.json`; an editor's own root only has to hold the binary or `go.mod`                      |

`tools/vscode-apl/language-configuration.json` is the same data in VS Code's format, with the exact regular
expressions for the word and the indentation; copy from there when an editor wants them.

## Build the server

```sh
go build --tags=with_db -o wowsimcli ./cmd/wowsimcli   # in the repository root
./wowsimcli siml lsp                                  # speaks LSP on stdin and stdout, logs on stderr
```

The server reads the item database at run time, so it checks files, hovers and completes with or without `with_db`;
only the panel's sim (`siml/sim`) needs the tag, as `siml sim` does. Rebuild after pulling Go changes. The server reads the name table
(`sim/siml/names.json`) and the database from the repository the file is in, so `make siml-names` or
`make db` needs no restart.

## What the server provides

Taken from its `initialize` answer (`serverInfo.name` is `wowsims-siml`):

| Capability                                                                                   | Notes                                                                                                        |
| -------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------ |
| `textDocumentSync`                                                                           | open and close, full text on change                                                                          |
| `publishDiagnostics`                                                                         | parser errors and warnings with did-you-mean, the registered-name check, the rotation lints, the slot checks |
| `completionProvider`                                                                         | trigger characters ` ` `(` `.` `"` `$` `%` `,` `=` `[` `+`                                                   |
| `hoverProvider`, `definitionProvider`, `referencesProvider`, `renameProvider` (with prepare) | variables, groups, sequences, placeholders, defs and parameters; a spell's row of the game files on hover    |
| `documentSymbolProvider`, `foldingRangeProvider`, `documentLinkProvider`                     | outline, folds, `use` paths                                                                                  |
| `documentFormattingProvider`                                                                 | the printer every committed file is a fixed point of                                                         |
| `codeActionProvider`                                                                         | kinds `quickfix` and `refactor.rewrite`                                                                      |
| `semanticTokensProvider` (`full`)                                                            | legend below                                                                                                 |

Semantic token types: `namespace`, `class`, `property`, `function`, `method`, `variable`, `parameter`,
`enumMember`, `type`. Modifiers: `declaration`, `readonly`, `defaultLibrary`. When the client registers file
watchers dynamically, the server also watches the name table, the database and the `*.siml` files.

The panel's `siml/*` requests (`siml/rotation`, `siml/sim`, ...) are custom and optional; a plain editor never sends
them.

## Highlighting

An editor colours SIML in one of three ways, from the most to the least that it needs from us:

1. **TextMate grammar.** `wowsimcli siml grammar -o siml.tmLanguage.json` writes it; no database tag is needed. The
   copy `.vscode/extensions/wowsims-apl/syntaxes/siml.tmLanguage.json` is committed and is what VS Code, and an
   IDE that loads VS Code bundles, reads. Its keywords, accessor heads, unit words and section names come from the
   schema, so regenerate it after the schema grows (`make vscode-apl` does).
2. **Semantic tokens.** The server colours what only the schema and the name table know (a spell is not an aura
   is not an item). It works on its own in an editor that applies semantic tokens, and on top of a grammar in one
   that also has one. Map the legend's types to the theme's groups.
3. **Tree-sitter.** Neovim's built-in highlighting, Helix and Zed colour from a tree-sitter grammar, which this
   repository does not have. Until one exists, those editors get the server's features, and the colours only from
   its semantic tokens where the editor applies them.

## Editors

None of these snippets has been run against the editor named; the server side of each is the verified
`initialize` answer above. Report what an editor needs beyond them.

### Neovim (0.10 or later)

```lua
vim.filetype.add({ extension = { siml = 'siml' } })
vim.api.nvim_create_autocmd('FileType', {
	pattern = 'siml',
	callback = function(args)
		local root = vim.fs.root(args.buf, 'go.mod')
		if root then
			vim.lsp.start({
				name = 'wowsims-siml',
				cmd = { root .. '/wowsimcli', 'siml', 'lsp' },
				cmd_cwd = root,
				root_dir = root,
			})
		end
	end,
})
vim.bo.commentstring = '// %s' -- or set it in a ftplugin/siml.lua
```

Neovim 0.11 can replace the autocmd with `vim.lsp.config('siml', { cmd = { '/path/to/wowsimcli', 'siml', 'lsp' },
cmd_cwd = '/path/to/forever', filetypes = { 'siml' } })` and `vim.lsp.enable('siml')`. Semantic tokens are on by
default; `:Inspect` lists the `@lsp.type.*` groups to link to your theme.

### Helix

In the repository's `.helix/languages.toml` or the user's:

```toml
[language-server.wowsims-siml]
command = "./wowsimcli"
args = ["siml", "lsp"]

[[language]]
name = "siml"
scope = "source.siml"
file-types = ["siml"]
comment-tokens = ["//"]
language-servers = ["wowsims-siml"]
roots = ["go.mod"]
```

Helix starts the server in the root it finds for the file. It colours from tree-sitter only, so a SIML file is
uncoloured until a grammar exists.

### Zed

Zed starts a language server only for an extension, so a settings file cannot do it. The shape is the one of
`tools/zed-spelldata`: a Rust extension that returns the command, and a language directory.

```toml
# extension.toml
id = "wowsims-siml"
name = "WoWSims SIML"
version = "0.1.0"
schema_version = 1
authors = ["WoWSims"]
description = "SIML files: the language server of wowsimcli."
repository = "https://github.com/wowsims/forever"

[language_servers.wowsims-siml]
name = "WoWSims SIML"
languages = ["SIML"]
```

```toml
# languages/siml/config.toml
name = "SIML"
path_suffixes = ["siml"]
line_comments = ["// "]
brackets = [
	{ start = "(", end = ")", close = true, newline = false },
	{ start = "[", end = "]", close = true, newline = false },
	{ start = "\"", end = "\"", close = true, newline = false, not_in = ["string", "comment"] },
]
```

`language_server_command` returns `<root>/wowsimcli siml lsp` (or `go -C <root> run --tags=with_db ./cmd/wowsimcli
siml lsp`, as the spelldata extension runs `go`). `config.toml` names no `grammar`, because none exists, and the
colours then come from the server's semantic tokens, which Zed applies when the language asks for them in
`settings.json`:

```json
{ "languages": { "SIML": { "semantic_tokens": "full" } } }
```

Install it with `zed: install dev extension`, which needs Rust through `rustup`. Run _editor: restart language
server_ after rebuilding `wowsimcli`.

### Emacs (eglot, 29 or later)

```elisp
(define-derived-mode siml-mode prog-mode "SIML"
  (setq-local comment-start "// "))
(add-to-list 'auto-mode-alist '("\\.siml\\'" . siml-mode))
(with-eval-after-load 'eglot
  (add-to-list 'eglot-server-programs
               '(siml-mode . ("/path/to/forever/wowsimcli" "siml" "lsp"))))
```

Eglot sends the language id `siml` from the mode's name. It does not apply semantic tokens; `lsp-mode` does.

### Sublime Text (LSP package)

Add a client in the LSP package's settings with `"command": ["/path/to/wowsimcli", "siml", "lsp"]`,
`"selector": "source.siml"` and a syntax whose scope is `source.siml`. Sublime reads `.sublime-syntax` and
`.tmLanguage` (plist) files, so convert the generated grammar first.

### JetBrains IDEs

The _TextMate Bundles_ settings page takes a VS Code extension folder: add `.vscode/extensions/wowsims-apl` for the
colours. The server runs through the built-in LSP support of the paid IDEs or the LSP4IJ plugin, as a user-defined
server for `*.siml` with the command above.

## When something is not served

- Nothing answers: run the command by hand in the repository root; it waits on stdin, and a build or a Go error
  shows at once on stderr.
- A file is served by the wrong repository: a repository nested inside another (a worktree under
  `.claude/worktrees`) is served by its own server only; open it as its own root.
- Names are missing: `make siml-names`, then save a file so the server reads the table again.
- An untitled buffer reads the first workspace folder's repository.
- Extend the server, not the client: a feature that goes in `sim/siml/lsp` or `sim/siml/lang_*.go` reaches every
  editor above. A change to what the server advertises changes the table in this document.

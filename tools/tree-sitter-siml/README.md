# tree-sitter-siml

The tree-sitter grammar of SIML (`sim/siml/SPEC.md`), for editors that colour, indent and outline with tree-sitter.
`tools/zed-siml` is built on it.

The parser in `sim/siml` is the language. This grammar reads the shapes of a file (sections, rows, entries,
expressions, `when` blocks) and leaves every name to the language server, so it needs no change when the schema or the
name table grows, and a file it reads may still hold errors that the server reports. A section's header decides how its
body reads: `prepull:`, `priority:` and `group "…":` hold rows, every other section entries. Before the first header
a line reads as a row when it starts with an action keyword, `#` or a time, or has an `if`, and as an entry otherwise.
A row whose keyword has nothing after it yet (`cast` alone) reads as that step, so a row being typed ends at its line.

- `grammar.js`: the rules.
- `src/scanner.c`: where a statement ends (SPEC.md §3.1), where a `when` body starts and ends (§6.9), and which lines
  are headers. On every committed `.siml` file it starts each statement where the Go parser does.
- `src/parser.c`, `src/grammar.json`, `src/node-types.json`, `src/tree_sitter/`: written by `npm run generate` and
  committed, since an editor compiles `src/parser.c` and `src/scanner.c` as they are.
- `test/corpus/`: the shapes that need care, each with the tree it reads as.

## Changing it

```sh
cd tools/tree-sitter-siml
npm ci
npm run generate   # after editing grammar.js
npm test           # the corpus; `npx tree-sitter test -u` records the tree of a new case, which you then read
npm run check      # what CI runs
```

`npm run check` regenerates the parser with the pinned `tree-sitter-cli` and compares it with `src/`, runs the corpus,
parses every committed `.siml` file but the error cases of `sim/siml/testdata/errors`, compiles the Zed queries, parses
each Zed snippet, and checks that the commit `tools/zed-siml/extension.toml` fetches the grammar from holds this
`src/`. `npx tree-sitter parse <file>` prints the tree of one file, and `npx tree-sitter query
../zed-siml/languages/siml/highlights.scm <file>` what a query captures in it.

A change to the syntax in `sim/siml` (`lex.go`, `syntax.go`) is a change here too.

Zed reads the grammar from a commit, so a change to it takes two commits: the change, then, once it is pushed,
`rev` in `tools/zed-siml/extension.toml` set to that commit's hash. Until then `npm run check` names the rev as stale.

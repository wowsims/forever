# Editing rotations in VS Code

A rotation is an `.apl.siml` file, the text form `sim/siml/SPEC.md` describes. `make vscode-apl` builds the workspace
extension, which starts the SIML language server (`wowsimcli siml lsp`) that checks, completes, explains and formats
every `*.siml` file; see `tools/vscode-apl/README.md`.

A rotation with one prepull cast and one priority row looks like

```
siml 1
spec hunter

prepull:
  -1s cast hunters_mark

priority:
  cast hunters_mark if not target.aura(hunters_mark).is_active
```

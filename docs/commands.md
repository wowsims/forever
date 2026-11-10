# Commands

We use a makefile for our build system. These commands will usually be all you need while developing for this project:

```sh
# Installs a pre-commit git hook so that your go code is automatically formatted (if you don't use an IDE that supports that).  If you want to manually format go code you can run make fmt.
# Also installs `air` to reload the dev servers automatically
make setup

# Run all the tests. Currently only the backend sim has tests.
make test

# Update the expected test results. This will need to be run after adding/removing any tests, and also if test results change due to code changes.
make update-tests

# Host a local version of the UI at http://localhost:8080. Visit it by pointing a browser to
# http://localhost:8080/forever/YOUR_SPEC_HERE, where YOUR_SPEC_HERE is the directory under ui/ with your custom code.
# Recompiles the entire client before launching using `make dist/forever`
npm start
# Or
make host

# With file-watching so the server auto-restarts and recompiles on Go or TS changes:
npm start
# Or
WATCH=1 make host

# Serve the component catalogue (every story and the design-system pages) at http://localhost:6006
npm run storybook:dev

# Build the same catalogue as static files into storybook-static/
npm run storybook:build

# Delete all generated files (.pb.go and .ts proto files, and dist/)
make clean

# Recompiles the ts only for the given spec (e.g. make host_elemental_shaman)
make host_$spec

# Recompiles the `wowsimforever` server binary and runs it, hosting /dist directory at http://localhost:3333/forever.
# This is the fastest way to iterate on core go simulator code so you don't have to wait for client rebuilds.
# To rebuild client for a spec just do 'make $spec' and refresh browser.
make rundevserver

# With file-watching so the server auto-restarts and recompiles on Go or TS changes:
WATCH=1 make rundevserver

# The same as rundevserver, recompiles  `wowsimforever` binary and runs it on port 3333. Instead of serving content from the dist folder,
# this command also runs `vite serve` to start the Vite dev server on port 5173 (or similar) and automatically reloads the page on .ts changes in less than a second.
# This allows for more rapid development, with sub second reloads on TS changes. This combines the benefits of `WATCH=1 make rundevserver` and `WATCH=1 make host`
# to create something that allows you to work in any part of the code with ease and speed.
# This might get rolled into `WATCH=1 make rundevserver` at some point.
WATCH=1 make devmode

# This is just the same as rundevserver currently
make devmode

# This command recompiles the workers in the /ui/worker folder for easier debugging/development
# Can be used with or without WATCH command
make webworkers

# With file watch enabled
WATCH=1 make webworkers

# Creates the 'wowsimforever' binary that can host the UI and run simulations natively (instead of with wasm).
# Builds the UI and the compiles it into the binary so that you can host the sim as a server instead of wasm on the client.
# It does this by first doing make dist/forever and then copying all those files to binary_dist/forever and loading all the files in that directory into its binary on compile.
make wowsimforever

# Using the --usefs flag will instead of hosting the client built into the binary, it will host whatever code is found in the /dist directory.
# Use --wasm to host the client with the wasm simulator.
# The server also disables all caching so that refreshes should pickup any changed files in dist/. The client will still call to the server to run simulations so you can iterate more quickly on client changes.
# make dist/forever && ./wowsimforever --usefs would rebuild the whole client and host it. (you would have had to run `make devserver` to build the wowsimforever binary first.)
./wowsimforever --usefs

# Generate code for the sim database (db.json). Only necessary if you changed the items generator.
# Useful only if you're actively working on the generator and have already run make db locally at least once.
make simdb

# Generate data from WoW game files
# Requires a local WoW install. It reads the folder in WOW_DIR, or Settings.BaseDir in the settings
# file (/mnt/c/Program Files/World of Warcraft) when WOW_DIR is not set.
# Uses tools/database/generator-settings.json for settings
# Also runs make simdb
# This is what you will use most of the time for generation
make db

# Same as make db but from the PTR install, and WOW_DIR applies the same way
# Uses tools/database/ptr-generator-settings.json for settings
make ptrdb

# WoW installed somewhere other than C:? Point make db and make ptrdb at the folder that holds
# _classic_beta_ and the other game folders. For one run:
make db WOW_DIR="/mnt/e/World of Warcraft"
# Or for every run, from your shell profile (then source the profile or open a new terminal).
# Linux and WSL (bash):
echo 'export WOW_DIR="/mnt/e/World of Warcraft"' >> ~/.bashrc
# macOS (zsh):
echo 'export WOW_DIR="/Applications/World of Warcraft"' >> ~/.zshrc
# Windows, running make natively (PowerShell; open a new terminal afterwards):
[Environment]::SetEnvironmentVariable("WOW_DIR", "E:\World of Warcraft", "User")
# Windows, Command Prompt (open a new terminal afterwards):
setx WOW_DIR "E:\World of Warcraft"
# A variable set on the Windows side does not reach WSL; under WSL, use the bash line above.

# Same as make db but without a WoW install: reads the current build off Blizzard's CDN
# and applies the hotfix cache you point it at (see tools/db2tool/README.md).
# The "Update DB" GitHub workflow runs this with raidbots' mirror of the live DBCache.bin.
make db DB2TOOL_FLAGS="--cdn --dbcache tools/db2tool/caches/DBCache.bin --carry-hotfixes"

# Regenerate the spell store and the sim/<class>/spell_data_auto_gen.go tables from the game files
# Also rewrites the buff and debuff files in sim/core/buffs, their settings inputs and proto/buffs.proto
# Needs tools/database/wowsims.db, so run make db at least once first
make spelldata

# Rewrite proto/buffs.proto alone from tools/database/buffmanifest
# Needs no game files database and no sim that builds; make sim/core/proto/api.pb.go runs it before protoc
go run ./tools/gen_buffs_proto

# Run one sim from a RaidSimRequest in protojson, or from an x.siml (the request built as the site builds it, without
# the potions and conjured items the site adds for its APL; with no encounter, against one default target), and print
# the RaidSimResult
# A preset build, which names no race, is loaded over its spec's defaults.siml as the site loads it, and the command
# says so; an auto or simple rotation, which the site builds into an APL, sims only the APL rows the file holds
# Nothing migrates an input file: export the settings from the UI again after a proto change
go run --tags=with_db ./cmd/wowsimcli sim --infile input.json

# SIML files (sim/siml/SPEC.md); the siml commands need the name table, built in with --tags=with_db
go run --tags=with_db ./cmd/wowsimcli siml parse [-json] x.apl.siml               # check a file loads, or print its proto
go run --tags=with_db ./cmd/wowsimcli siml print -kind apl -spec hunter x.apl.json # print a JSON file as SIML
go run --tags=with_db ./cmd/wowsimcli siml fmt -check ui/specs                    # list files not canonical or failing the check (SPEC.md §7.4)
go run --tags=with_db ./cmd/wowsimcli siml verify x.apl.json x.apl.siml           # check a print reads back as its JSON

# Sim the text of an .apl.siml for the APL editor: a JSON request on stdin, DPS, row counts, spells and a timeline
# as JSON on stdout (the request and result are described in cmd/wowsimcli/cmd/siml_sim.go)
echo '{"siml": "...", "path": "ui/specs/hunter/dps/apls/default.apl.siml"}' | go run --tags=with_db ./cmd/wowsimcli siml sim

# The SIML language server on stdin and stdout, run from the repository root; the VS Code extension builds and starts
# it (tools/vscode-apl/README.md)
go run --tags=with_db ./cmd/wowsimcli siml lsp
# Write the extension's TextMate grammar, which make vscode-apl does
go run ./cmd/wowsimcli siml grammar -o .vscode/extensions/wowsims-apl/syntaxes/siml.tmLanguage.json
# Write the wowsims-siml skill, the short guide with every action, value and accessor from proto/apl.proto, which make
# siml does
go run ./cmd/wowsimcli siml llm -o .github/skills/wowsims-siml/SKILL.md

# Add names for newly registered spells and new database rows to the SIML name table; commit it with the change
make siml-names
# Before committing, and after a change to proto/apl.proto: the name table as above, the wowsims-siml skill, then siml
# fmt -check over ui/specs
make siml
# The wowsimcli vite reads .siml imports with (tools/vite/siml.mts); make builds it with --tags=with_db, so the siml
# commands above run on it too, and rebuilds it after a Go change
make bin/wowsimcli
```

# A partially written target (e.g. the two-step wasm gzip recipe killed mid-write) must
# not be treated as up to date on the next run.
.DELETE_ON_ERROR:

OUT_DIR := dist/forever
# Windows won't launch an extensionless binary -- air just pops a file-association prompt.
BIN_EXT := $(shell go env GOEXE)
ASSETS_INPUT := $(shell find assets/ -type f)
ASSETS := $(patsubst assets/%,$(OUT_DIR)/assets/%,$(ASSETS_INPUT))
# Recursive wildcard function. Needs to be '=' instead of ':=' because of recursion.
rwildcard = $(foreach d,$(wildcard $(1:=/*)),$(call rwildcard,$d,$2) $(filter $(subst *,%,$2),$d))
GOROOT := $(shell go env GOROOT)
UI_SRC := $(shell find ui -name '*.ts' -o -name '*.tsx' -o -name '*.css' -o -name '*.html')
AUTO_GEN_FILES_TS := ui/sim/player/classes/capabilities_auto_gen.ts ui/sim/bulk/constants_auto_gen.ts ui/sim/wasm/bulk_sim/constants_auto_gen.ts ui/sim/proto/sim_tables_auto_gen.ts ui/sim/wasm/stat_weights_constants_auto_gen.ts
AUTO_GEN_FILES_TS_DEPS := sim/core/base_stats.go sim/core/bulk/candidates.go sim/core/bulk/statistics.go sim/core/consumes_imbues.go sim/core/database.go sim/core/statweight.go tools/database/gen_character_constants_ts.go tools/database/gen_bulksim_constants.ts.go tools/database/gen_sim_tables_ts.go tools/database/gen_stat_weight_constants_ts.go sim/core/proto/api.pb.go

$(OUT_DIR)/.dirstamp: \
  $(OUT_DIR)/lib.wasm.gz \
  ui/generated/proto/api.ts \
  $(ASSETS) \
  $(OUT_DIR)/bundle/.dirstamp
	touch $@

$(OUT_DIR)/bundle/.dirstamp: \
  $(UI_SRC) \
  $(AUTO_GEN_FILES_TS) \
  vite.config.mts \
  vite.build-workers.mts \
  tools/vite/spec_pages.mts \
  node_modules \
  tsconfig.json \
  ui/generated/proto/api.ts
	node_modules/typescript/bin/tsc --noEmit
	npx tsx vite.build-workers.mts
	npx vite build
	touch $@

.PHONY: clean
clean:
	rm -rf ui/generated/proto/*.ts \
	  sim/core/proto/*.pb.go \
	  wowsimforever$(BIN_EXT) \
	  wowsimforever-windows.exe \
	  wowsimforever-amd64-darwin \
	  wowsimforever-arm64-darwin \
	  wowsimforever-amd64-linux \
	  dist \
	  binary_dist \
	  ui/generated/proto/*.ts \
	  node_modules
	find . -name "*.results.tmp" -type f -delete

ui/generated/proto/api.ts: proto/*.proto node_modules
	mkdir -p ui/generated/proto
	npx protoc --ts_opt generate_dependencies --ts_out ui/generated/proto --proto_path proto proto/api.proto
	npx protoc --ts_out ui/generated/proto --proto_path proto proto/test.proto
	npx protoc --ts_out ui/generated/proto --proto_path proto proto/ui.proto

.PHONY: package.json

package.json:
# Checks if the system is FreeBSD and jq is installed. This is due to the need to switch out the vite package for rollup on FreeBSD.
ifeq ($(shell uname -s), FreeBSD)
	@if ! command -v jq > /dev/null; then \
		echo "jq is not installed. Please install jq to proceed."; \
		exit 1; \
	fi; \
	\
	echo "Checking and updating package.json for FreeBSD..."; \
	\
	if ! grep -q '"overrides"' package.json; then \
		jq '. + { "overrides": { "vite": { "rollup": "npm:@rollup/wasm-node@4.13.0" } } }' package.json > package.json.tmp && mv package.json.tmp package.json && npm install; \
	else \
		jq '.overrides += { "vite": { "rollup": "npm:@rollup/wasm-node@4.13.0" } }' package.json > package.json.tmp && mv package.json.tmp package.json && npm install; \
	fi
endif

package-lock.json:
	npm install

node_modules: package-lock.json
	npm ci

# Generic rule for hosting any class directory
.PHONY: host_%
host_%: $(OUT_DIR) node_modules $(AUTO_GEN_FILES_TS)
	npx http-server $(OUT_DIR)/..

.PHONY: wasm
wasm: $(OUT_DIR)/lib.wasm.gz

# Builds the generic .wasm, with all items included.
# Published gzipped: Cloudflare Pages caps files at 25 MiB.
# The main thread decompresses and compiles it once (see getSharedWasmModule in
# ui/sim/workers/worker_pool.ts) and shares the compiled module with every worker.
WASM_FEATURES := --enable-sign-ext --enable-nontrapping-float-to-int --enable-mutable-globals --enable-bulk-memory
$(OUT_DIR)/lib.wasm.gz: sim/wasm/* sim/core/proto/api.pb.go $(filter-out sim/core/items/all_items.go, $(call rwildcard,sim,*.go))
	@echo "Starting webassembly compile now..."
	@if GOWASM=satconv,signext GOOS=js GOARCH=wasm go build -ldflags "-w -s" -o ./$(OUT_DIR)/lib.wasm ./sim/wasm/; then \
		printf "\033[1;32mWASM compile successful.\033[0m\n"; \
	else \
		printf "\033[1;31mWASM COMPILE FAILED\033[0m\n"; \
		exit 1; \
	fi
	@if command -v wasm-opt >/dev/null 2>&1; then \
		echo "Optimizing wasm with wasm-opt..."; \
		wasm-opt -O3 $(WASM_FEATURES) $(OUT_DIR)/lib.wasm -o $(OUT_DIR)/lib.wasm.tmp && mv -f $(OUT_DIR)/lib.wasm.tmp $(OUT_DIR)/lib.wasm; \
	else \
		printf "\033[1;33mwasm-opt not found -- skipping optimization (install binaryen to enable).\033[0m\n"; \
	fi
	gzip -9 -f -n $(OUT_DIR)/lib.wasm

$(OUT_DIR)/assets/%: assets/%
	mkdir -p $(@D)
	cp $< $@
	rm -rf $(OUT_DIR)/assets/db_inputs


binary_dist/dist.go: sim/web/dist.go.tmpl
	mkdir -p binary_dist/forever
	touch binary_dist/forever/embedded
	cp sim/web/dist.go.tmpl binary_dist/dist.go

binary_dist: $(OUT_DIR)/.dirstamp
	rm -rf binary_dist
	mkdir -p binary_dist
	cp -r $(OUT_DIR) binary_dist/
	rm -f binary_dist/forever/lib.wasm binary_dist/forever/lib.wasm.gz
	rm -rf binary_dist/forever/assets/db_inputs
	rm binary_dist/forever/assets/database/db.bin
	rm binary_dist/forever/assets/database/leftover_db.bin

# Rebuild the protobuf generated code.
.PHONY: proto
proto: sim/core/proto/api.pb.go ui/generated/proto/api.ts

# The buff messages come from the checked-in manifest, not from hand-edited proto. The manifest
# imports the compiled protos, so this is run by hand after a manifest edit rather than as a
# prerequisite of them; TestBuffFilesRegenerateFromTheCommittedInputs fails while buffs.proto is stale.
.PHONY: buffs-proto
buffs-proto:
	go run ./tools/gen_buffs_proto

# Builds the web server with the compiled client.
.PHONY: wowsimforever
wowsimforever: binary_dist devserver

.PHONY: devserver
devserver: sim/core/proto/api.pb.go sim/web/*.go binary_dist/dist.go
	@echo "Starting server compile now..."
	@if go build -o wowsimforever$(BIN_EXT) ./sim/web ; then \
		printf "\033[1;32mBuild Completed Successfully\033[0m\n"; \
	else \
		printf "\033[1;31mBUILD FAILED\033[0m\n"; \
		exit 1; \
	fi

.PHONY: air
air:
ifeq ($(WATCH), 1)
	@if ! command -v air; then \
		echo "Missing air dependency. Please run \`make setup\`"; \
		exit 1; \
	fi
endif

rundevserver: air devserver $(AUTO_GEN_FILES_TS)
ifeq ($(WATCH), 1)
	npx tsx vite.build-workers.mts & npx vite build -m development --watch &
	ulimit -n 10240 && air -tmp_dir "/tmp" -build.include_ext "go,proto" -build.args_bin "--usefs=true --launch=false" -build.bin "./wowsimforever$(BIN_EXT)" -build.cmd "make devserver" -build.exclude_dir "assets,dist,node_modules,ui,tools"
else
	./wowsimforever$(BIN_EXT) --usefs=true --launch=false --host=":3333"
endif

wowsimforever-windows.exe: wowsimforever
# go build only considers syso files when invoked without specifying .go files: https://github.com/golang/go/issues/16090
	cp ./assets/favicon_io/icon-windows_amd64.syso ./sim/web/icon-windows_amd64.syso
	cd ./sim/web/ && GOOS=windows GOARCH=amd64 GOAMD64=v2 go build -o wowsimforever-windows.exe -ldflags="-X 'main.Version=$(VERSION)' -s -w"
	cd ./cmd/wowsimcli && GOOS=windows GOARCH=amd64 GOAMD64=v2 go build -o wowsimcli-windows.exe --tags=with_db -ldflags="-X 'main.Version=$(VERSION)' -s -w"
	rm ./sim/web/icon-windows_amd64.syso
	mv ./sim/web/wowsimforever-windows.exe ./wowsimforever-windows.exe
	mv ./cmd/wowsimcli/wowsimcli-windows.exe ./wowsimcli-windows.exe

release: wowsimforever wowsimforever-windows.exe
	GOOS=darwin GOARCH=amd64 GOAMD64=v2 go build -o wowsimforever-amd64-darwin -ldflags="-X 'main.Version=$(VERSION)' -s -w" ./sim/web
	GOOS=darwin GOARCH=arm64 go build -o wowsimforever-arm64-darwin -ldflags="-X 'main.Version=$(VERSION)' -s -w" ./sim/web
	GOOS=darwin GOARCH=arm64 go build -o wowsimcli-arm64-darwin --tags=with_db -ldflags="-X 'main.Version=$(VERSION)' -s -w" ./cmd/wowsimcli/cli_main.go
	GOOS=linux GOARCH=amd64 GOAMD64=v2 go build -o wowsimforever-amd64-linux   -ldflags="-X 'main.Version=$(VERSION)' -s -w" ./sim/web
	GOOS=linux GOARCH=amd64 GOAMD64=v2 go build -o wowsimcli-amd64-linux --tags=with_db -ldflags="-X 'main.Version=$(VERSION)' -s -w" ./cmd/wowsimcli/cli_main.go
# Now compress into a zip because the files are getting large.
	zip wowsimforever-windows.exe.zip wowsimforever-windows.exe
	zip wowsimforever-amd64-darwin.zip wowsimforever-amd64-darwin
	zip wowsimforever-arm64-darwin.zip wowsimforever-arm64-darwin
	zip wowsimforever-amd64-linux.zip wowsimforever-amd64-linux
	zip wowsimcli-amd64-linux.zip wowsimcli-amd64-linux
	zip wowsimcli-arm64-darwin.zip wowsimcli-arm64-darwin
	zip wowsimcli-windows.exe.zip wowsimcli-windows.exe

sim/core/proto/api.pb.go: proto/*.proto
	@if go version -m "$$(command -v protoc-gen-go)" 2>/dev/null | grep -qE '^[[:space:]]+mod[[:space:]]+github\.com/golang/protobuf[[:space:]]'; then \
		echo "ERROR: your protoc-gen-go is the deprecated github.com/golang/protobuf plugin;"; \
		echo "it generates code that no longer builds against this repo's protobuf version."; \
		echo "Fix:  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"; \
		echo "then: rm -f sim/core/proto/*.pb.go && retry"; \
		exit 1; \
	fi
# Distro protoc packages (e.g. Ubuntu/WSL's protobuf-compiler) bundle a
# descriptor.proto whose go_package still points at the deprecated
# github.com/golang/protobuf path, which is no longer a dependency. Pin the
# mapping so common.proto's MessageOptions extension resolves to descriptorpb.
	protoc -I=./proto \
		--go_opt=Mgoogle/protobuf/descriptor.proto=google.golang.org/protobuf/types/descriptorpb \
		--go_out=./sim/core ./proto/*.proto

$(AUTO_GEN_FILES_TS): $(AUTO_GEN_FILES_TS_DEPS)
	go run ./tools/database/gen_db -gen=go-to-ts

.PHONY: go-to-ts
go-to-ts: $(AUTO_GEN_FILES_TS)

.PHONY: character-constants-ts
character-constants-ts: go-to-ts

# Only useful for building the lib on a host platform that matches the target platform
.PHONY: locallib
locallib: sim/core/proto/api.pb.go
	go build -buildmode=c-shared -o wowsimforever.so --tags=with_db ./sim/lib/library.go

.PHONY: nixlib
nixlib: sim/core/proto/api.pb.go
	GOOS=linux GOARCH=amd64 GOAMD64=v2 go build -buildmode=c-shared -o wowsimforever-linux.so --tags=with_db ./sim/lib/library.go

.PHONY: winlib
winlib: sim/core/proto/api.pb.go
	GOOS=windows GOARCH=amd64 GOAMD64=v2 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -buildmode=c-shared -o wowsimforever-windows.dll --tags=with_db ./sim/lib/library.go

.PHONY: simdb
simdb: sim/core/items/all_items.go sim/core/proto/api.pb.go

# $(abspath) is make's own lexical resolver: unlike `realpath` it neither requires
# the path to exist nor follows symlinks. wowsims.db is gitignored and may well be
# a symlink to a sibling checkout's copy, and `realpath` would then hand db2tool
# THAT repo's database to overwrite.
CLIENTDATA_SETTINGS := $(abspath ./tools/database/generator-settings.json)
CLIENTDATAPTR_SETTINGS := $(abspath ./tools/database/ptr-generator-settings.json)
CLIENTDATA_OUTPUT   := $(abspath ./tools/database/wowsims.db)
# Extra db2tool flags. `make db DB2TOOL_FLAGS="--cdn --dbcache path/to/DBCache.bin"`
# reads the build off Blizzard's CDN instead of BaseDir (no install needed).
DB2TOOL_FLAGS ?=
# The WoW install `make db` and `make ptrdb` read, in place of the settings file's BaseDir.
# `make db WOW_DIR="/mnt/e/World of Warcraft"`, or export WOW_DIR once in your shell profile.
WOW_DIR ?=
ifneq ($(WOW_DIR),)
DB2TOOL_FLAGS += --basedir "$(WOW_DIR)"
endif
# CLASSES, RACIALS and SHARED choose what `make db`, `make ptrdb` and `make spelldata` take from the
# extracted build. `make db CLASSES=druid` takes the druid and keeps every other class as committed;
# `CLASSES=all` and `CLASSES=none` name every class or none. `RACIALS=1` takes the racials, and
# `SHARED=0` keeps the shared group as committed, so `CLASSES=none SHARED=0` takes the items alone.
# With none set, the run takes the default spell data groups. See docs/spell_data/partial-updates.md.
CLASSES ?=
RACIALS ?=
SHARED ?=
SPELLDATA_FLAGS := $(if $(CLASSES),-classes=$(CLASSES)) $(if $(filter 1 true yes,$(RACIALS)),-racials) $(if $(filter 0 false no,$(SHARED)),-shared=false)
GEN_DB_FLAGS := $(if $(CLASSES),-classes=$(CLASSES))

# The spell store is regenerated before the item database, and the order is load-bearing: gen_db
# classifies every item and enchant proc from the store compiled into it, so running it against a
# refreshed game files and a stale store writes item files from last week's rows. gen_spelldata reads the
# store too, and writes nothing the sim does not compile against, so it is what may rewrite it.
.PHONY: db
db:
	@echo "Extracting game files data"
	go run ./tools/db2tool -s $(CLIENTDATA_SETTINGS) --output $(CLIENTDATA_OUTPUT) $(DB2TOOL_FLAGS)
	@echo "Regenerating the spell store"
	go run ./tools/database/gen_spelldata $(SPELLDATA_FLAGS)
	@echo "Running DBC generation tool"
	go run tools/database/gen_db/*.go -outDir=./assets -gen=db $(GEN_DB_FLAGS)

# Regenerates sim/<class>/spell_data_auto_gen.go from tools/database/wowsims.db, so it
# runs after `make db` and from the repo root.
.PHONY: spelldata
spelldata:
	go run ./tools/database/gen_spelldata $(SPELLDATA_FLAGS)

# Regenerates sim/core/base_stats_auto_gen.go and ui/sim/constants/mechanics.ts from
# the same numbers, so the sim and the UI cannot disagree about a rating conversion.
# Needs tools/database/wowsims.db, so it runs after `make db` and from the repo root.
.PHONY: basestats
basestats:
	python3 tools/base_stats_parser.py

# Same order as `db` above, and for the same reason.
.PHONY: ptrdb
ptrdb:
	@echo "Extracting game files data"
	go run ./tools/db2tool -s $(CLIENTDATAPTR_SETTINGS) --output $(CLIENTDATA_OUTPUT) $(DB2TOOL_FLAGS)
	@echo "Regenerating the spell store"
	go run ./tools/database/gen_spelldata $(SPELLDATA_FLAGS)
	@echo "Running DBC generation tool"
	go run tools/database/gen_db/*.go -outDir=./assets -gen=db $(GEN_DB_FLAGS)

# The same order as `db`, and for the same reason: this rule depends on tools/database/*.go, so a
# change to the store generator triggers it, and gen_db has to read a store written by that change.
sim/core/items/all_items.go: $(call rwildcard,tools/database,*.go) $(call rwildcard,sim/core/proto,*.go)
	@test -f tools/database/wowsims.db || { \
		echo "ERROR: tools/database/wowsims.db is missing (gitignored, produced by 'make db')."; \
		echo "Run 'make db' to extract it from a local WoW install."; \
		exit 1; }
	@test -f tools/db2tool/listfile.csv || { \
		echo "tools/db2tool/listfile.csv is missing, downloading it..."; \
		curl -fL -o tools/db2tool/listfile.csv https://github.com/wowdev/wow-listfile/releases/latest/download/community-listfile.csv; }
	@echo "Regenerating the spell store"
	go run ./tools/database/gen_spelldata
	go run tools/database/gen_db/*.go -outDir=./assets -gen=db

# Syncs the HiGHS solver artifacts from the pinned npm `highs` package: copies its wasm to
# ui/worker/highs.wasm (which the Go backend go:embeds) and regenerates the minified-name map the
# wazero host looks its imports and exports up through. Run after changing the pinned version in
# package.json, and commit both outputs. TestHiGHSArtifactsMatchPinnedVersion fails if you don't.
.PHONY: update-highs
update-highs: node_modules
	go run ./tools/gen_highs

.PHONY: test
test: $(OUT_DIR)/lib.wasm.gz binary_dist/dist.go
	GOARCH=amd64 go test --tags=with_db ./sim/...

.PHONY: update-tests
update-tests:
	find . -name "*.results" -type f -delete
	find . -name "*.results.tmp" -exec bash -c 'cp "$$1" "$${1%.results.tmp}".results' _ {} \;

# Names the generated spell data files that are not what tools/database/gen_spelldata writes
# today, and writes nothing. Exits 1 with the paths when the committed files have drifted, which a
# game files update or a hand edit both do. A spell data group the inputs record on a different game
# build from the database's is printed as not ported, and the check still passes. Needs tools/database/wowsims.db.
.PHONY: spelldata-check
spelldata-check:
	go run ./tools/database/gen_spelldata -check

.PHONY: fmt
fmt: tsfmt
	gofmt -w ./sim
	gofmt -w ./tools

.PHONY: tsfmt
tsfmt:
	npx oxfmt ui

# one time setup to install pre-commit hook for gofmt and npm install needed packages
setup:
	cp pre-commit .git/hooks
	chmod +x .git/hooks/pre-commit
	! command -v air && curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin || true

# Host a local server, for dev testing
.PHONY: host
host: air $(OUT_DIR)/.dirstamp node_modules $(AUTO_GEN_FILES_TS)
ifeq ($(WATCH), 1)
	ulimit -n 10240 && air -tmp_dir "/tmp" -build.include_ext "go,ts,js,html" -build.bin "npx" -build.args_bin "http-server $(OUT_DIR)/.." -build.cmd "make" -build.exclude_dir "dist,node_modules,tools"
else
	# Intentionally serve one level up, so the local site has 'forever' as the first
	# directory just like github pages.
	npx http-server $(OUT_DIR)/..
endif

devmode: air devserver $(AUTO_GEN_FILES_TS)
ifeq ($(WATCH), 1)
	npx tsx vite.build-workers.mts & npx vite serve --host &
	air -tmp_dir "/tmp" -build.include_ext "go,proto" -build.args_bin "--usefs=true --launch=false --wasm=false" -build.bin "./wowsimforever$(BIN_EXT)" -build.cmd "make devserver" -build.exclude_dir "assets,dist,node_modules,ui,tools"
else
	./wowsimforever$(BIN_EXT) --usefs=true --launch=false --host=":3333"
endif

webworkers:
	npx tsx vite.build-workers.mts --watch=$(if $(WATCH),true,false)

.PHONY: vscode-spelldata
vscode-spelldata:
	cd tools/vscode-spelldata && npm ci --no-audit --no-fund && npm run build

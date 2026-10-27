// db2tool extracts data from the World of Warcraft game files into tools/database/wowsims.db.
//
// The default mode reads the local install named by the settings' BaseDir, or
// by --basedir when given:
// .build.info picks the build, files come from local CASC
// (root → encoding → .idx → data.NNN → BLTE), .dbd definitions and the
// community listfile are fetched/cached over plain HTTPS. The game files'
// DBCache.bin hotfixes for the extracted build are applied to the decoded
// rows; --dbcache <file> pins specific cache files instead of the default
// scan and --no-hotfixes disables the overlay. --carry-hotfixes also applies
// the given caches of older builds (rules in README.md). Encrypted sections are
// decrypted with every TACT key the game files ship, the server pushed into
// DBCache.bin, or the community list carries (--keys <file>, default
// tools/db2tool/TACTKeys.txt, refreshed from wowdev/TACTKeys).
//
// With --cdn, the same pipeline reads the product's current build straight
// off Blizzard's CDN (--region picks the version service, default us) — no
// install required, and hotfixes come from --dbcache or tools/db2tool/caches.
//
// With --build (and optionally --db2dir/--dbddir), the offline mode decodes
// pre-extracted .db2 files instead — no install required and no hotfixes
// unless --dbcache is given.
package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wowsims/forever/tools/db2tool/config"
	"github.com/wowsims/forever/tools/db2tool/dbd"
	"github.com/wowsims/forever/tools/db2tool/sqlite"
	"github.com/wowsims/forever/tools/db2tool/tact"
	"github.com/wowsims/forever/tools/db2tool/wdc"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "db2tool:", err)
		os.Exit(1)
	}
}

type options struct {
	settingsFile  string
	databaseFile  string
	baseDir       string   // overrides the settings' BaseDir when set
	db2Dir        string   // offline mode only
	dbdDir        string   // offline mode override
	buildNumber   uint32   // nonzero → offline mode
	buildArg      string   // the --build value as given, recorded as the game version in offline mode
	dbCaches      []string // explicit DBCache files, overriding the default scan
	noHotfixes    bool     // skip hotfix application entirely
	carryHotfixes bool     // also apply caches of older builds, under the layout guard
	keyFile       string   // community TACT key list; empty → tools/db2tool/TACTKeys.txt
	cdn           bool     // read the current build from Blizzard's CDN instead of BaseDir
	region        string   // version service region for --cdn
}

// parseArgs scans the args pairwise. --settings/-s and --output/-output/-o
// are aliases, plus the offline-mode and hotfix flags.
func parseArgs(args []string) (options, error) {
	opts := options{
		settingsFile: "appsettings.json",
		databaseFile: "wowsims.db",
		region:       "us",
	}
	for i := 0; i < len(args); i++ {
		next := func() (string, error) {
			if i+1 < len(args) {
				i++
				return args[i], nil
			}
			return "", fmt.Errorf("flag %s needs a value", args[i])
		}
		var err error
		switch args[i] {
		case "--settings", "-s":
			opts.settingsFile, err = next()
		case "--output", "-output", "-o":
			opts.databaseFile, err = next()
		case "--basedir":
			opts.baseDir, err = next()
		case "--db2dir":
			opts.db2Dir, err = next()
		case "--dbddir":
			opts.dbdDir, err = next()
		case "--dbcache":
			var f string
			if f, err = next(); err == nil {
				opts.dbCaches = append(opts.dbCaches, f)
			}
		case "--no-hotfixes":
			opts.noHotfixes = true
		case "--carry-hotfixes":
			opts.carryHotfixes = true
		case "--keys":
			opts.keyFile, err = next()
		case "--cdn":
			opts.cdn = true
		case "--region":
			opts.region, err = next()
		case "--build":
			var v string
			if v, err = next(); err == nil {
				// Accept either a bare build number (68571) or a full version
				// string (5.5.4.68571), of which only the trailing component
				// identifies the build.
				n := v
				if dot := strings.LastIndexByte(n, '.'); dot >= 0 {
					n = n[dot+1:]
				}
				b, perr := strconv.ParseUint(n, 10, 32)
				if perr != nil {
					return opts, fmt.Errorf("invalid --build %q: want a build number like 68571 or a version like 5.5.4.68571", v)
				}
				opts.buildNumber = uint32(b)
				opts.buildArg = v
			}
		default:
			return opts, fmt.Errorf("unknown argument %q", args[i])
		}
		if err != nil {
			return opts, err
		}
	}
	// The default scan finds every game's DBCache.bin under the install (retail, Classic, Era), and
	// carried, any of them would overlay Forever's tables; carrying takes only the caches named.
	if opts.carryHotfixes && len(opts.dbCaches) == 0 {
		return opts, fmt.Errorf("--carry-hotfixes needs --dbcache <file>: it carries only the caches named")
	}
	return opts, nil
}

// resolvePath resolves a possibly-relative settings path against the tool
// home directory tools/db2tool — relative settings values like
// "../../assets/db_inputs/basestats" are anchored there.
func resolvePath(toolHome, value string) string {
	if filepath.IsAbs(value) {
		return value
	}
	return filepath.Join(toolHome, value)
}

func run(args []string) error {
	opts, err := parseArgs(args)
	if err != nil {
		return err
	}

	settingsAbs, err := filepath.Abs(opts.settingsFile)
	if err != nil {
		return err
	}
	settings, err := config.Load(settingsAbs)
	if err != nil {
		return fmt.Errorf("loading settings: %w", err)
	}
	if len(settings.Tables) == 0 {
		return fmt.Errorf("settings file lists no Tables")
	}
	if opts.baseDir != "" {
		settings.Settings.BaseDir = opts.baseDir
	}

	// The tool home is resolved from the working directory, which must be the
	// repo root — the same invariant gen_db already has (its ./tools/... and
	// ./assets literals). Fail loud otherwise.
	toolHome, err := filepath.Abs(filepath.Join("tools", "db2tool"))
	if err != nil {
		return err
	}
	if _, err := os.Stat(toolHome); err != nil {
		return fmt.Errorf("tools/db2tool not found — run from the repository root (CWD-dependent like gen_db): %w", err)
	}
	dbdCacheDir := opts.dbdDir
	if dbdCacheDir == "" {
		dbdCacheDir = filepath.Join(toolHome, "DBDCache")
	}

	var buildNumber uint32
	var version string
	var openTable func(tableName string) (*wdc.Table, error)
	var build *tact.Build // local-CASC and CDN modes only

	if opts.buildNumber != 0 {
		// Offline mode: pre-extracted .db2 files.
		buildNumber = opts.buildNumber
		version = opts.buildArg
		if !strings.Contains(version, ".") {
			fmt.Printf("Warning: BuildInfo records the game build as %q; pass the full version, such as 1.60.1.70170, so it matches an install or CDN extraction\n",
				version)
		}
		db2Dir := opts.db2Dir
		if db2Dir == "" {
			db2Dir = filepath.Join(toolHome, "dbfilesclient")
		}
		openTable = func(tableName string) (*wdc.Table, error) {
			return wdc.ReadFile(filepath.Join(db2Dir, tableName+".db2"))
		}
	} else {
		// Local-CASC mode (the default) and CDN mode: everything from the
		// install or the CDN, through the same Build.
		if opts.cdn {
			build, err = tact.OpenCDN(settings.Settings.Product, opts.region, filepath.Join(toolHome, "cdncache"))
		} else if settings.Settings.BaseDir == "" {
			return fmt.Errorf("settings BaseDir is required (or pass --cdn, or --build for offline mode)")
		} else {
			build, err = tact.Open(settings.Settings.BaseDir, settings.Settings.Product)
		}
		if err != nil {
			return err
		}
		buildNumber = build.BuildNumber
		version = build.Entry.Version
		source := "local install"
		if opts.cdn {
			source = opts.region + " CDN"
		}
		fmt.Printf("Extracting %s %s (build %d) from %s\n", build.Entry.Product, build.Entry.Version, buildNumber, source)

		listfile := &tact.Listfile{Path: filepath.Join(toolHome, "listfile.csv")}
		if err := listfile.Refresh(); err != nil {
			return err
		}

		// GameTables: raw bytes, filename casing preserved from settings.
		gameTablesOutDir := resolvePath(toolHome, settings.GameTablesOutDirectory)
		if err := os.MkdirAll(gameTablesOutDir, 0o755); err != nil {
			return err
		}
		for _, gameTable := range settings.GameTables {
			fdid, err := listfile.GetFDID("gametables/" + gameTable + ".txt")
			if err != nil {
				return err
			}
			data, err := build.OpenFileByFDID(fdid)
			if err != nil {
				return fmt.Errorf("gametable %s: %w", gameTable, err)
			}
			if err := os.WriteFile(filepath.Join(gameTablesOutDir, gameTable+".txt"), data, 0o644); err != nil {
				return err
			}
		}

		// Tables: extract each .db2 to the target directory, then parse it.
		// The FDID key uses the RAW settings TargetDirectory value; only the
		// on-disk output use is resolved.
		targetDirOnDisk := resolvePath(toolHome, settings.TargetDirectory)
		if err := os.MkdirAll(targetDirOnDisk, 0o755); err != nil {
			return err
		}
		openTable = func(tableName string) (*wdc.Table, error) {
			fdid, err := listfile.GetFDID(settings.TargetDirectory + "/" + tableName + ".db2")
			if err != nil {
				return nil, err
			}
			data, err := build.OpenFileByFDID(fdid)
			if err != nil {
				return nil, fmt.Errorf("table %s: %w", tableName, err)
			}
			path := filepath.Join(targetDirOnDisk, tableName+".db2")
			if err := os.WriteFile(path, data, 0o644); err != nil {
				return nil, err
			}
			return wdc.ReadFile(path)
		}
	}

	// Hotfixes: overlay the game files' DBCache.bin records on every decoded
	// table. By default only a cache for this exact build applies; having none
	// is not an error. --dbcache pins specific cache files (deterministic
	// runs); with no override, local-CASC mode scans tools/db2tool/caches plus
	// <BaseDir>/**/DBCache.bin, CDN mode only tools/db2tool/caches, and the
	// offline --build mode stays hotfix-free.
	// --carry-hotfixes adds older builds' caches, in OrderHotfixReaders' order.
	var carried []*wdc.HotfixReader
	var own *wdc.HotfixReader
	if !opts.noHotfixes {
		var readers map[uint32]*wdc.HotfixReader
		if len(opts.dbCaches) > 0 {
			if readers, err = wdc.CombineHotfixFiles(opts.dbCaches); err != nil {
				return err
			}
		} else if opts.buildNumber == 0 {
			baseDir := settings.Settings.BaseDir
			if opts.cdn {
				baseDir = ""
			}
			if readers, err = wdc.LoadHotfixCaches(filepath.Join(toolHome, "caches"), baseDir); err != nil {
				return err
			}
		}
		var newer []uint32
		carried, own, newer = wdc.OrderHotfixReaders(readers, buildNumber, opts.carryHotfixes)
		for _, b := range newer {
			fmt.Fprintf(os.Stderr, "db2tool: warning: not carrying hotfixes from build %d: newer than build %d\n", b, buildNumber)
		}
		for _, r := range carried {
			fmt.Printf("Carrying hotfixes from build %d onto build %d\n", r.BuildID, buildNumber)
		}
		// A hotfix-free run is easy to miss in the output, and it also means
		// no server-pushed TACT keys were picked up, so say so.
		noReaders := len(carried) == 0 && own == nil
		switch {
		case noReaders && len(opts.dbCaches) > 0:
			fmt.Fprintf(os.Stderr, "db2tool: warning: none of the given --dbcache files hold hotfixes for build %d; continuing without the overlay\n", buildNumber)
		case noReaders && opts.cdn:
			fmt.Fprintf(os.Stderr, "db2tool: warning: no DBCache.bin for build %d found under %s; continuing without the overlay\n", buildNumber, filepath.Join(toolHome, "caches"))
		case noReaders && opts.buildNumber == 0:
			fmt.Fprintf(os.Stderr, "db2tool: warning: no DBCache.bin for build %d found under %s or %s; continuing without the overlay\n", buildNumber, settings.Settings.BaseDir, filepath.Join(toolHome, "caches"))
		}
	}
	carriedApplied := make(map[*wdc.HotfixReader]bool)

	// decodeTable opens a table, decodes it against its definition for this
	// build and overlays the hotfixes.
	decodeTable := func(tableName string) (*wdc.Table, sqlite.TableDef, *wdc.Decoded, error) {
		var td sqlite.TableDef
		table, err := openTable(tableName)
		if err != nil {
			return nil, td, nil, err
		}
		dbdPath, err := dbd.FetchCached(dbdCacheDir, tableName)
		if err != nil {
			return nil, td, nil, err
		}
		def, err := dbd.ReadFile(dbdPath, true)
		if err != nil {
			return nil, td, nil, err
		}
		version, err := dbd.SelectVersion(def, buildNumber)
		if err != nil {
			return nil, td, nil, fmt.Errorf("table %s: %w", tableName, err)
		}
		decoded, err := table.DecodeRows(def, version, buildNumber)
		if err != nil {
			return nil, td, nil, fmt.Errorf("table %s: %w", tableName, err)
		}
		for _, r := range carried {
			applied, layoutSkipped, err := r.ApplyHotfixes(table, def, version, buildNumber, decoded)
			if err != nil {
				return nil, td, nil, fmt.Errorf("table %s: applying hotfixes: %w", tableName, err)
			}
			if layoutSkipped {
				fmt.Fprintf(os.Stderr, "db2tool: warning: table %s: build %d does not share build %d's layout; not carrying its hotfixes into this table\n",
					tableName, r.BuildID, buildNumber)
			}
			if applied {
				carriedApplied[r] = true
			}
		}
		if own != nil {
			if _, _, err := own.ApplyHotfixes(table, def, version, buildNumber, decoded); err != nil {
				return nil, td, nil, fmt.Errorf("table %s: applying hotfixes: %w", tableName, err)
			}
		}
		return table, sqlite.TableDef{Name: tableName, Def: def, Version: version}, decoded, nil
	}

	// TACT keys: the game files' own TactKey/TactKeyLookup tables (with any
	// pushed keys the hotfix overlay adds) plus the community list. Only the
	// local-CASC and CDN modes can use them: offline .db2 files were already
	// written with their encrypted chunks zero-filled.
	var keys *tact.KeyStore
	if build != nil {
		keys = tact.NewKeyStore()
		keyPath := opts.keyFile
		if keyPath == "" {
			keyPath = filepath.Join(toolHome, "TACTKeys.txt")
		}
		fromClient, err := loadClientKeys(keys, decodeTable)
		if err != nil {
			return err
		}
		community := 0
		if err := tact.RefreshKeyFile(keyPath, ""); err != nil {
			fmt.Fprintf(os.Stderr, "db2tool: %v; continuing without the community key list\n", err)
		} else if community, err = keys.LoadKeyFile(keyPath); err != nil {
			return err
		}
		build.Keys = keys
		fmt.Printf("TACT keys: %d loaded (%d from the game files and their hotfixes, %d more from %s)\n", keys.Len(), fromClient, community, keyPath)
	}

	type loaded struct {
		def     sqlite.TableDef
		table   *wdc.Table
		decoded *wdc.Decoded
	}
	tables := make([]loaded, 0, len(settings.Tables))
	tableDefs := make([]sqlite.TableDef, 0, len(settings.Tables))

	for _, tableName := range settings.Tables {
		table, td, decoded, err := decodeTable(tableName)
		if err != nil {
			return err
		}
		reportLockedSections(tableName, table, keys)
		tables = append(tables, loaded{def: td, table: table, decoded: decoded})
		tableDefs = append(tableDefs, td)
	}

	// The push counts this build's own reader and each carried reader that
	// applied to at least one table. The workflows read the Hotfixes line.
	var hotfixPush int32
	applied := own != nil
	if own != nil {
		hotfixPush = own.HighestPush()
	}
	var carriedFrom string
	for _, r := range carried {
		if !carriedApplied[r] {
			fmt.Fprintf(os.Stderr, "db2tool: warning: the cache for build %d applied to no table\n", r.BuildID)
			continue
		}
		hotfixPush = max(hotfixPush, r.HighestPush())
		carriedFrom += fmt.Sprintf(", carried from build %d", r.BuildID)
		applied = true
	}
	if applied {
		fmt.Printf("Hotfixes: build %d up to push %d%s\n", buildNumber, hotfixPush, carriedFrom)
	} else {
		fmt.Printf("Hotfixes: none applied for build %d\n", buildNumber)
	}

	db, err := sqlite.Open(opts.databaseFile)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := sqlite.CreateTables(db, tableDefs); err != nil {
		return err
	}

	for _, t := range tables {
		if err := sqlite.InsertRows(db, t.def, t.decoded); err != nil {
			return err
		}
	}
	if err := sqlite.WriteBuildInfo(db, version, settings.Settings.Product, hotfixPush); err != nil {
		return err
	}

	fmt.Println("Processing completed.")
	return nil
}

// loadClientKeys decodes TactKeyLookup (key id → 64-bit name) and TactKey
// (key id → key) and adds every pair to keys. Both tables take the hotfix
// overlay, which is how a key the server pushes into DBCache.bin arrives.
func loadClientKeys(keys *tact.KeyStore, decodeTable func(string) (*wdc.Table, sqlite.TableDef, *wdc.Decoded, error)) (int, error) {
	_, lookupDef, lookups, err := decodeTable("TactKeyLookup")
	if err != nil {
		return 0, err
	}
	_, keyDef, tactKeys, err := decodeTable("TactKey")
	if err != nil {
		return 0, err
	}
	names := make(map[int32]uint64, len(lookups.Rows))
	col := columnIndex(lookupDef, "TACTID")
	for _, row := range lookups.Rows {
		if raw := byteArray(row.Values[col]); len(raw) == 8 {
			names[row.ID] = binary.LittleEndian.Uint64(raw)
		}
	}
	added := 0
	col = columnIndex(keyDef, "Key")
	for _, row := range tactKeys.Rows {
		raw := byteArray(row.Values[col])
		name, ok := names[row.ID]
		if !ok || len(raw) != 16 {
			continue
		}
		var key [16]byte
		copy(key[:], raw)
		if !keys.Has(name) {
			added++
		}
		keys.Add(name, key)
	}
	return added, nil
}

func columnIndex(td sqlite.TableDef, name string) int {
	for i, d := range td.Version.Definitions {
		if d.Name == name {
			return i
		}
	}
	panic(fmt.Sprintf("table %s has no column %s", td.Name, name))
}

// byteArray flattens a decoded u8 array column, which the decoder hands back
// as []int64 (or []uint64 for 64-bit unsigned elements).
func byteArray(v any) []byte {
	switch a := v.(type) {
	case []int64:
		out := make([]byte, len(a))
		for i, x := range a {
			out[i] = byte(x)
		}
		return out
	case []uint64:
		out := make([]byte, len(a))
		for i, x := range a {
			out[i] = byte(x)
		}
		return out
	}
	return nil
}

// reportLockedSections names every key a table's encrypted sections still
// need, so a run says what it could not read instead of silently dropping
// rows.
func reportLockedSections(tableName string, t *wdc.Table, keys *tact.KeyStore) {
	locked := make(map[uint64]int32)
	for _, s := range t.Sections {
		if s.TactKeyLookup != 0 && (keys == nil || !keys.Has(s.TactKeyLookup)) {
			locked[s.TactKeyLookup] += s.NumRecords
		}
	}
	for name, rows := range locked {
		fmt.Printf("  %s: %d rows locked by TACT key %s\n", tableName, rows, tact.KeyName(name))
	}
}

// icons writes the icon of every item, enchant, consumable and spell the database names into
// assets/icons as <name>.jpg, decoded from the game files' BLP on Blizzard's CDN, and lists the names it
// wrote in assets/icons/manifest.json. The folder is gitignored: deploy builds it.
//
// Run from the repository root: go run ./tools/icons
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/wowsims/forever/assets/database"
	"github.com/wowsims/forever/tools/db2tool/tact"
	"golang.org/x/sync/errgroup"
)

// The folder the icons and manifest.json go to, and the product and version service region whose current build
// they are read from.
const (
	out     = "assets/icons"
	product = "wowf"
	region  = "us"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "icons:", err)
		os.Exit(1)
	}
}

func run() error {
	toolHome := filepath.Join("tools", "db2tool")
	build, err := tact.OpenCDN(product, region, filepath.Join(toolHome, "cdncache"))
	if err != nil {
		return err
	}
	listfile := &tact.Listfile{Path: filepath.Join(toolHome, "listfile.csv")}
	if err := listfile.Refresh(); err != nil {
		return err
	}
	fdids, err := iconFDIDs(listfile.Path)
	if err != nil {
		return err
	}
	// Without the key list the few encrypted icons fail to decode and stay on zamimg; the rest still decode.
	keyFile := filepath.Join(toolHome, "TACTKeys.txt")
	build.Keys = tact.NewKeyStore()
	if err := tact.RefreshKeyFile(keyFile, ""); err != nil {
		fmt.Fprintln(os.Stderr, "icons: warning:", err)
	} else if _, err := build.Keys.LoadKeyFile(keyFile); err != nil {
		fmt.Fprintln(os.Stderr, "icons: warning:", err)
	}
	fmt.Printf("Reading icons from %s %s\n", build.Entry.Product, build.Entry.Version)

	names := iconNames()
	var missing, found []string
	for _, name := range names {
		if _, ok := fdids[name]; ok {
			found = append(found, name)
		} else {
			missing = append(missing, name+": not in the listfile")
		}
	}

	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	errs := make([]error, len(found))
	var g errgroup.Group
	g.SetLimit(16)
	for i, name := range found {
		g.Go(func() error {
			errs[i] = writeIcon(build, fdids[name], filepath.Join(out, name+".jpg"))
			return nil
		})
	}
	g.Wait()
	var written []string
	for i, name := range found {
		if errs[i] != nil {
			missing = append(missing, fmt.Sprintf("%s (%d): %v", name, fdids[name], errs[i]))
		} else {
			written = append(written, name)
		}
	}

	for _, m := range missing {
		fmt.Println("Left to zamimg: " + m)
	}
	// With nothing decoded the site would quietly fall back to zamimg for every icon, so the run fails instead.
	if len(written) == 0 {
		return fmt.Errorf("decoded none of the %d icons the database names", len(names))
	}
	manifest, err := json.Marshal(written)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "manifest.json"), manifest, 0o644); err != nil {
		return err
	}
	fmt.Printf("Wrote %d icons to %s; %d left to zamimg\n", len(written), out, len(missing))
	return nil
}

// The file data id of every icon in the listfile, by the name gen_db gives it (database.GetIconName): the
// file name without its extension, lowercased, with its spaces as dashes.
func iconFDIDs(listfile string) (map[string]uint32, error) {
	f, err := os.Open(listfile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fdids := make(map[string]uint32)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		id, path, ok := strings.Cut(scanner.Text(), ";")
		path = strings.ToLower(path)
		if !ok || !strings.HasPrefix(path, "interface/icons/") {
			continue
		}
		fdid, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			continue
		}
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		fdids[strings.ReplaceAll(name, " ", "-")] = uint32(fdid)
	}
	return fdids, scanner.Err()
}

// Every icon name the shipped database holds, db.bin and leftover_db.bin together.
func iconNames() []string {
	db := database.Load()
	var names []string
	for _, item := range db.Items {
		names = append(names, item.Icon)
	}
	for _, enchant := range db.Enchants {
		names = append(names, enchant.Icon)
	}
	for _, consumable := range db.Consumables {
		names = append(names, consumable.Icon)
	}
	for _, data := range slices.Concat(db.ItemIcons, db.SpellIcons) {
		names = append(names, data.Icon)
	}
	slices.Sort(names)
	names = slices.Compact(names)
	return slices.DeleteFunc(names, func(name string) bool { return name == "" })
}

func writeIcon(build *tact.Build, fdid uint32, path string) error {
	data, err := build.OpenFileByFDID(fdid)
	if err != nil {
		return err
	}
	img, err := decodeBLP(data)
	if err != nil {
		return err
	}
	// A game files icon has a frame a sixteenth of its width wide (4 px of 64). The site draws its own border,
	// so the frame is cut off, which leaves 56 px, the size of zamimg's large icons.
	inset := img.Rect.Dx() / 16
	framed := img.SubImage(image.Rect(inset, inset, img.Rect.Dx()-inset, img.Rect.Dy()-inset))
	// JPEG has no alpha: the encoder reads premultiplied colours, so a transparent pixel comes out black.
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, framed, &jpeg.Options{Quality: 80}); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

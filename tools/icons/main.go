// icons writes the icon of every item, enchant, consumable and spell the database names into
// assets/icons as <name>.jpg, decoded from the game files' BLP on Blizzard's CDN, and lists the names it
// wrote in assets/icons/manifest.json. The site serves an icon from there when the manifest lists it and
// from zamimg otherwise (ui/sim/proto/wowhead.ts). The folder is gitignored: deploy builds it.
//
// Run from the repository root: go run ./tools/icons
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/wowsims/forever/assets/database"
	"github.com/wowsims/forever/tools/db2tool/tact"
)

func main() {
	out := flag.String("out", "assets/icons", "the folder the icons and manifest.json are written to")
	product := flag.String("product", "wowf", "the product whose current build is read")
	region := flag.String("region", "us", "the version service region")
	flag.Parse()
	if err := run(*out, *product, *region); err != nil {
		fmt.Fprintln(os.Stderr, "icons:", err)
		os.Exit(1)
	}
}

func run(out, product, region string) error {
	toolHome := filepath.Join("tools", "db2tool")
	build, err := tact.OpenCDN(product, region, filepath.Join(toolHome, "cdncache"))
	if err != nil {
		return err
	}
	listfile := &tact.Listfile{Path: filepath.Join(toolHome, "listfile.csv")}
	if err := listfile.Refresh(); err != nil {
		return err
	}
	keyFile := filepath.Join(toolHome, "TACTKeys.txt")
	if err := tact.RefreshKeyFile(keyFile, ""); err != nil {
		return err
	}
	build.Keys = tact.NewKeyStore()
	if _, err := build.Keys.LoadKeyFile(keyFile); err != nil {
		return err
	}
	fmt.Printf("Reading icons from %s %s\n", build.Entry.Product, build.Entry.Version)

	// The listfile loads on the first lookup and is not safe for concurrent use, so every id is found
	// before the downloads start.
	type icon struct {
		name string
		fdid uint32
	}
	var icons []icon
	var missing []string
	for _, name := range iconNames() {
		fdid, err := listfile.GetFDID("interface/icons/" + name + ".blp")
		if err != nil {
			// The database writes a space in a game path as a dash.
			fdid, err = listfile.GetFDID("interface/icons/" + strings.ReplaceAll(name, "-", " ") + ".blp")
		}
		if err != nil {
			missing = append(missing, name+": not in the listfile")
			continue
		}
		icons = append(icons, icon{name, fdid})
	}

	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	var mu sync.Mutex
	var written []string
	work := make(chan icon)
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			for ic := range work {
				err := writeIcon(build, ic.fdid, filepath.Join(out, ic.name+".jpg"))
				mu.Lock()
				if err != nil {
					missing = append(missing, fmt.Sprintf("%s (%d): %v", ic.name, ic.fdid, err))
				} else {
					written = append(written, ic.name)
				}
				mu.Unlock()
			}
		})
	}
	for _, ic := range icons {
		work <- ic
	}
	close(work)
	wg.Wait()

	slices.Sort(written)
	manifest, err := json.Marshal(written)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "manifest.json"), manifest, 0o644); err != nil {
		return err
	}
	slices.Sort(missing)
	fmt.Printf("Wrote %d icons to %s; %d left to zamimg\n", len(written), out, len(missing))
	for _, m := range missing {
		fmt.Println("  " + m)
	}
	return nil
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

package siml

import (
	googleProto "google.golang.org/protobuf/proto"

	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
)

// Which sections a file holds, from its double extension (SPEC.md §1).
type Kind struct {
	Section string // The section an x.<section>.siml holds, such as apl or gear; "" for a full setup, x.siml.
}

// The kind of the file at path. A path that does not end in .siml, or whose double extension names no section,
// is an error.
func KindOf(path string) (Kind, error) {
	if !strings.HasSuffix(path, ".siml") {
		return Kind{}, fmt.Errorf("%s is not a .siml file", path)
	}
	k := kindOf(path)
	if k.Section == "" {
		return k, nil
	}
	ix, err := loadIndex()
	if err != nil {
		return Kind{}, err
	}
	if ix.sections[k.Section] == nil {
		return Kind{}, fmt.Errorf("%s: the file name names no section: %s", path, k.Section)
	}
	return k, nil
}

// The kind a path's double extension gives; text with no file name, or one not ending in .siml, is a full setup.
func kindOf(path string) Kind {
	base := path[strings.LastIndexAny(path, `/\`)+1:]
	stem, ok := strings.CutSuffix(base, ".siml")
	if !ok {
		return Kind{}
	}
	if dot := strings.LastIndexByte(stem, '.'); dot > 0 && dot < len(stem)-1 {
		return Kind{Section: stem[dot+1:]}
	}
	return Kind{}
}

// Reads a file that a use line names, at a path resolved against the using file.
type Opener func(path string) ([]byte, error)

// A name a file writes, and what it means there (SPEC.md §7).
type NameRef struct {
	Name       string
	Namespace  Namespace
	Start, End Position
	ID         int32    // The spell id, item id, enchant effect id or random suffix id it means.
	Row        *SlotRow // The slot row it is written on, if any.
	Path       []any    // The JSON path, within the loader's message (SPEC.md §1), of the field it fills; nil outside it.
	Known      bool     // Whether its namespace holds it; Doc.Refs gives the names it does not hold too.
}

// A slot row: its slot, and the item its suffix and enchant are on.
type SlotRow struct {
	Slot proto.ItemSlot
	Item int32
}

// A parsed file. One with an error holds no settings, categories or warnings, and still lists the names and raw refs
// that read.
type File struct {
	Path       string
	Kind       Kind
	Spec       proto.Spec                   // SpecUnknown when the file names no spec.
	Settings   *proto.IndividualSimSettings // What the file and the files it uses set (SPEC.md §5.1).
	Categories []string                     // The import categories the file holds, in the site's order (SPEC.md §5.1).
	Names      []NameRef                    // The names the file itself writes, in file order.
	Uses       []string                     // The paths its use lines name, read against its own path, in file order.
	RawRefs    Diagnostics                  // raw-ref at each raw token that holds an id a name could write, which no file under ui/specs holds (SPEC.md §7.3).
	Warnings   Diagnostics
}

// The rotation of an x.apl.siml, or of a full file that holds one.
func (f *File) Rotation() *proto.APLRotation { return f.Settings.GetPlayer().GetRotation() }

// The 17 items of an x.gear.siml, or of a full file that holds gear:.
func (f *File) Equipment() *proto.EquipmentSpec { return f.Settings.GetPlayer().GetEquipment() }

// What the file loads as (SPEC.md §1): an x.apl.siml its rotation, an x.gear.siml its equipment, any other file its
// settings.
func (f *File) Message() googleProto.Message {
	switch f.Kind.Section {
	case sectionAPL:
		return f.Rotation()
	case sectionGear:
		return f.Equipment()
	}
	return f.Settings
}

// The loader's message for a file of kind (SPEC.md §1), empty: an APLRotation, an EquipmentSpec or
// IndividualSimSettings.
func (k Kind) Message() googleProto.Message {
	switch k.Section {
	case sectionAPL:
		return &proto.APLRotation{}
	case sectionGear:
		return &proto.EquipmentSpec{}
	}
	return &proto.IndividualSimSettings{}
}

// The settings a loader's message stands for: an APLRotation as the rotation, an EquipmentSpec as the equipment.
func SettingsOf(m googleProto.Message) *proto.IndividualSimSettings {
	switch m := m.(type) {
	case *proto.APLRotation:
		return &proto.IndividualSimSettings{Player: &proto.Player{Rotation: m}}
	case *proto.EquipmentSpec:
		return &proto.IndividualSimSettings{Player: &proto.Player{Equipment: m}}
	}
	return m.(*proto.IndividualSimSettings)
}

// Parses src, the text of the file at path, whose names mean what names says (SPEC.md §7). The error is Diagnostics
// when the file has an error, which then holds its warnings too, and the file holds what read. open reads the files
// that use lines name, at paths read against path; with no open, a use line is use-refused.
func Parse(path string, src []byte, names *Names, open Opener) (*File, error) {
	return Analyze(path, src, names, open).File()
}

// The largest file ReadSource reads.
const maxSourceSize = 16 << 20

// Reads the file at path, which a use line or a command names: a regular file of at most 16 MB. A FIFO or a device
// is refused unread, since reading one can block forever or never end.
func ReadSource(path string) ([]byte, error) {
	if info, err := os.Stat(path); err != nil {
		return nil, err
	} else if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", path)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxSourceSize+1))
	if err == nil && len(data) > maxSourceSize {
		err = fmt.Errorf("%s is larger than %d MB", path, maxSourceSize>>20)
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Reads and parses the file at path with the committed name table, reading the files its use lines name from the
// file system. A build without the with_db tag holds no table, and reads only a file that writes no name.
func ReadFile(path string) (*File, error) {
	src, err := ReadSource(path)
	if err != nil {
		return nil, err
	}
	names, noTable := DefaultNames()
	if noTable != nil {
		names = &Names{}
	}
	f, err := Parse(path, src, names, ReadSource)
	var diags Diagnostics
	if noTable != nil && errors.As(err, &diags) && slices.ContainsFunc(diags, func(d Diagnostic) bool { return d.Code == "unknown-name" }) {
		return nil, fmt.Errorf("%s writes names: %w", path, noTable)
	}
	return f, err
}

package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/siml"
)

// What is simmed besides the APL: one player's settings, as the site sims them.
type character struct {
	settings *proto.IndividualSimSettings
	info     characterInfo
}

type characterInfo struct {
	Source   string   `json:"source"` // The files the settings come from.
	Race     string   `json:"race"`
	Talents  string   `json:"talents"`
	Gear     string   `json:"gear,omitempty"` // The spec's gear set the gear is, if it is one.
	Warnings []string `json:"warnings,omitempty"`
}

// The field of the spec oneof a player sets, or nil.
func specField(p *proto.Player) protoreflect.FieldDescriptor {
	m := p.ProtoReflect()
	return m.WhichOneof(m.Descriptor().Oneofs().ByName("spec"))
}

// The spec's folder under ui/specs, such as hunter/dps, and its defaults: the folder of the spec's class whose
// defaults.siml is for the spec. A defaults.siml that does not read is skipped, so another spec's broken file does not
// stop this one; the error names it when no folder is the spec's.
func specDefaults(root string, spec proto.Spec) (string, *siml.File, error) {
	var skipped []error
	if info, ok := siml.SpecOf(spec); ok {
		class := siml.ClassName(info.Class)
		paths, err := filepath.Glob(filepath.Join(root, "ui/specs", class, "*", "defaults.siml"))
		if err != nil {
			return "", nil, err
		}
		for _, path := range paths {
			f, err := siml.ReadFile(path)
			if err != nil {
				skipped = append(skipped, err)
				continue
			}
			if f.Spec == spec {
				return class + "/" + filepath.Base(filepath.Dir(path)), f, nil
			}
		}
	}
	return "", nil, errors.Join(append([]error{fmt.Errorf("ui/specs holds no defaults.siml for %s", spec)}, skipped...)...)
}

// The character the site sims a rotation with: the spec's defaults, with the build loaded over them that uses the
// file at path, when exactly one of the spec's builds does. A rotation several builds use, or none, is simmed on the
// defaults, which the warnings say for several.
func characterFromRepo(root, folder string, defaults *siml.File, path string) (*character, error) {
	c := &character{
		settings: cloneProto(defaults.Settings),
		info:     characterInfo{Source: defaults.Path},
	}
	builds, err := buildsUsing(root, folder, path)
	if err != nil {
		return nil, err
	}
	switch len(builds) {
	case 1:
		siml.LoadBuild(c.settings, builds[0].Settings)
		c.info.Source = builds[0].Path + " over " + c.info.Source
	case 0:
	default:
		names := make([]string, len(builds))
		for i, b := range builds {
			names[i] = filepath.Base(b.Path)
		}
		c.info.Warnings = append(c.info.Warnings, fmt.Sprintf("builds %s all use this file; simmed on the defaults", strings.Join(names, ", ")))
	}
	c.info.Gear = gearSetOf(root, folder, c.settings.GetPlayer().GetEquipment())
	return c, nil
}

// The builds of the spec in folder whose use lines name the file at path.
func buildsUsing(root, folder, path string) ([]*siml.File, error) {
	paths, err := filepath.Glob(filepath.Join(root, "ui/specs", folder, "builds", "*.siml"))
	if err != nil {
		return nil, err
	}
	target, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	var found []*siml.File
	for _, p := range paths {
		build, err := repoFiles.read(p)
		if err != nil {
			return nil, err
		}
		if slices.ContainsFunc(build.Uses, func(u string) bool {
			used, err := filepath.Abs(filepath.FromSlash(u))
			return err == nil && used == target
		}) {
			found = append(found, build)
		}
	}
	return found, nil
}

// The name of the spec's gear set e is, or "".
func gearSetOf(root, folder string, e *proto.EquipmentSpec) string {
	paths, _ := filepath.Glob(filepath.Join(root, "ui/specs", folder, "gear_sets", "*.gear.siml"))
	for _, path := range paths {
		if f, err := repoFiles.read(path); err == nil && googleProto.Equal(f.Equipment(), e) {
			return strings.TrimSuffix(filepath.Base(path), ".gear.siml")
		}
	}
	return ""
}

// The builds and gear sets under ui/specs as last read, read again when the file or a file it uses changed. A file
// that does not read is read again each time.
var repoFiles = &fileCache{files: map[string]cachedFile{}}

type fileCache struct {
	mu    sync.Mutex
	files map[string]cachedFile
}

type cachedFile struct {
	stamps []siml.FileStamp // The file's, then each used file's.
	file   *siml.File
}

func (c *fileCache) read(path string) (*siml.File, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if held, ok := c.files[path]; ok && slices.Equal(held.stamps, stampsOf(path, held.file)) {
		return held.file, nil
	}
	f, err := siml.ReadFile(path)
	if err == nil {
		c.files[path] = cachedFile{stampsOf(path, f), f}
	}
	return f, err
}

func stampsOf(path string, f *siml.File) []siml.FileStamp {
	stamps := []siml.FileStamp{siml.StampOf(path)}
	if f != nil {
		for _, u := range f.Uses {
			stamps = append(stamps, siml.StampOf(filepath.FromSlash(u)))
		}
	}
	return stamps
}

// The character a settings file describes: an IndividualSimSettings the site exports, in protojson or as an x.siml.
func characterFromSettings(root, path string, spec proto.Spec, folder string) (*character, error) {
	c := &character{settings: &proto.IndividualSimSettings{}, info: characterInfo{Source: path}}
	if strings.HasSuffix(path, ".siml") {
		var err error
		if c, err = readSetup(root, path); err != nil {
			return nil, err
		}
	} else if data, err := os.ReadFile(path); err != nil {
		return nil, err
	} else if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(data, c.settings); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if info, ok := siml.SpecOf(spec); !ok || specField(c.settings.GetPlayer()) != info.Field {
		return nil, fmt.Errorf("%s is not a %s character", path, folder)
	}
	return c, nil
}

// The character of the x.siml at path: a full setup as it is, or a preset build, which names no race, loaded over its
// spec's defaults.siml under root. The warnings are the file's, and for a build the defaults it is loaded over.
func readSetup(root, path string) (*character, error) {
	f, err := siml.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if f.Kind.Section != "" {
		return nil, fmt.Errorf("%s holds only its %s section, and a setup is an x.siml: a full setup or a preset build", path, f.Kind.Section)
	}
	c := &character{settings: f.Settings, info: characterInfo{Source: path}}
	for _, w := range f.Warnings {
		c.info.Warnings = append(c.info.Warnings, diagnosticLine(w))
	}
	// A preset build sits in a builds folder, and may set its race; any other file is a build only when it names none.
	if filepath.Base(filepath.Dir(path)) != "builds" && f.Settings.GetPlayer().GetRace() != proto.Race_RaceUnknown {
		return c, nil
	}
	_, defaults, err := specDefaults(root, f.Spec)
	if err != nil {
		return nil, fmt.Errorf("%s is a preset build, and %w", path, err)
	}
	c.settings = cloneProto(defaults.Settings)
	siml.LoadBuild(c.settings, f.Settings)
	c.info.Source = path + " over " + defaults.Path
	c.info.Warnings = append(c.info.Warnings, fmt.Sprintf("%s is a preset build: simmed over %s", path, defaults.Path))
	return c, nil
}

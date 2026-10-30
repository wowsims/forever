package gen

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/siml"
)

// The repin record, relative to the repository's root: a line for each name a -repin gave another meaning.
const RecordPath = "sim/siml/gen/repinned.txt"

// The lines that open a repin record.
const recordHeader = `# The names simlnames -repin gave another meaning, one line each, as the run reports them (sim/siml/SPEC.md §7.4).
# The check against a pull request's base branch accepts a name of the base's table that means something else, or
# nothing, only when the pull request adds the name's line here. simlnames -w -repin writes the lines; keep them.
`

// The line a repin record holds for r.
func (r Repin) String() string { return fmt.Sprintf("%s: was %s, now %s", r.Pin, r.Was, r.Now) }

// record with a line for each of repins that gave its name another meaning and is not in it yet, and the header
// first when record is empty.
func AppendRecord(record []byte, repins []Repin) []byte {
	out := slices.Clone(record)
	if len(bytes.TrimSpace(out)) == 0 {
		out = []byte(recordHeader)
	}
	have := map[string]bool{}
	for _, line := range strings.Split(string(out), "\n") {
		have[strings.TrimSpace(line)] = true
	}
	for _, r := range repins {
		if line := r.String(); r.Was != r.Now && !have[line] {
			if len(out) > 0 && out[len(out)-1] != '\n' {
				out = append(out, '\n')
			}
			out = append(out, line+"\n"...)
			have[line] = true
		}
	}
	return out
}

// The names of the lines record holds that base, an earlier record, does not hold.
func RecordedNames(record, base []byte) map[string]bool {
	old := map[string]bool{}
	for _, line := range strings.Split(string(base), "\n") {
		old[strings.TrimSpace(line)] = true
	}
	names := map[string]bool{}
	for _, line := range strings.Split(string(record), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || old[line] {
			continue
		}
		if name, _, ok := strings.Cut(line, ":"); ok {
			names[strings.TrimSpace(name)] = true
		}
	}
	return names
}

// One line for each name of base that now means something else or nothing, unless repinned holds it, such as
// "hunter.spell.arcane_shot: was spell 3044, now spell 2643". A name means its ids, tag, rank and slots; registeredBy
// is no part of it.
func Changed(base, now *siml.Names, repinned map[string]bool) []string {
	was, is := meanings(base), meanings(now)
	var lines []string
	for _, p := range slices.SortedFunc(maps.Keys(was), comparePins) {
		got, ok := is[p]
		if !ok {
			got = "nothing"
		}
		if got != was[p] && !repinned[p.String()] {
			lines = append(lines, Repin{p, was[p], got}.String())
		}
	}
	return lines
}

func comparePins(a, b Pin) int { return strings.Compare(a.String(), b.String()) }

// What every name of t means, as the report writes it.
func meanings(t *siml.Names) map[Pin]string {
	m := map[Pin]string{}
	spells := func(class proto.Class, names siml.SpellNames) {
		for ns, entries := range names {
			for name, e := range entries {
				m[Pin{class, ns, name}] = entryText(e)
			}
		}
	}
	spells(proto.Class_ClassUnknown, t.Shared)
	for class, names := range t.Classes {
		spells(class, names)
	}
	for name, id := range t.Items {
		m[Pin{Namespace: siml.NamespaceItem, Name: name}] = "item " + strconv.Itoa(int(id))
	}
	for name, e := range t.Enchants {
		m[Pin{Namespace: siml.NamespaceEnchant, Name: name}] = enchantText(e)
	}
	for name, id := range t.Suffixes {
		m[Pin{Namespace: siml.NamespaceSuffix, Name: name}] = "suffix " + strconv.Itoa(int(id))
	}
	return m
}

// Drops from old, in place, the entries whose names head does not hold and that c no longer gives: spell-like
// entries no build registers, and items, enchants and suffixes the database no longer holds. Returns a line for each.
func prune(old, head *siml.Names, c *Collection) []string {
	var dropped []string
	drop := func(p Pin, meaning string) { dropped = append(dropped, fmt.Sprintf("%s = %s", p, meaning)) }
	spells := func(class proto.Class, names, committed siml.SpellNames) {
		for ns, entries := range names {
			for name, e := range entries {
				if _, ok := committed[ns][name]; !ok && !c.registers(class, ns, Identity{e.SpellID, e.Tag}) {
					drop(Pin{class, ns, name}, entryText(e))
					delete(entries, name)
				}
			}
		}
	}
	spells(proto.Class_ClassUnknown, old.Shared, head.Shared)
	for class, names := range old.Classes {
		spells(class, names, head.Classes[class])
	}
	items := map[int32]bool{}
	for _, it := range c.Items {
		items[it.ID] = true
	}
	for name, id := range old.Items {
		if _, ok := head.Items[name]; !ok && !items[id] {
			drop(Pin{Namespace: siml.NamespaceItem, Name: name}, "item "+strconv.Itoa(int(id)))
			delete(old.Items, name)
		}
	}
	rows := map[string]bool{}
	for _, row := range c.Enchants {
		rows[enchantKey(row.EffectID, row.Slots)] = true
	}
	for name, e := range old.Enchants {
		if _, ok := head.Enchants[name]; !ok && !rows[enchantKey(e.EffectID, e.Slots)] {
			drop(Pin{Namespace: siml.NamespaceEnchant, Name: name}, enchantText(e))
			delete(old.Enchants, name)
		}
	}
	suffixes := map[int32]bool{}
	for _, s := range c.Suffixes {
		suffixes[s.ID] = true
	}
	for name, id := range old.Suffixes {
		if _, ok := head.Suffixes[name]; !ok && !suffixes[id] {
			drop(Pin{Namespace: siml.NamespaceSuffix, Name: name}, "suffix "+strconv.Itoa(int(id)))
			delete(old.Suffixes, name)
		}
	}
	slices.Sort(dropped)
	return dropped
}

// Whether a build registers id in ns: a build of one of class's specs, or of any spec for ClassUnknown.
func (c *Collection) registers(class proto.Class, ns siml.Namespace, id Identity) bool {
	return slices.ContainsFunc(c.Registered[ns][id].Specs, func(s string) bool {
		return class == proto.Class_ClassUnknown || slices.Contains(c.Specs[class], s)
	})
}

// Whether root is the top of a git checkout, which a source tarball is not.
func IsCheckout(root string) bool {
	out, err := exec.Command("git", "-C", root, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return false
	}
	top, err1 := filepath.EvalSymlinks(filepath.FromSlash(strings.TrimSpace(string(out))))
	dir, err2 := filepath.EvalSymlinks(root)
	return err1 == nil && err2 == nil && top == dir
}

// The file at path in commit rev of the checkout at root; ok is false when rev does not hold it.
func GitFile(root, rev, path string) (data []byte, ok bool, err error) {
	if err := exec.Command("git", "-C", root, "rev-parse", "--verify", "--quiet", "--end-of-options", rev+"^{commit}").Run(); err != nil {
		return nil, false, fmt.Errorf("%s is no commit of the checkout at %s", rev, root)
	}
	data, err = exec.Command("git", "-C", root, "show", "--end-of-options", rev+":"+path).Output()
	if exit := (*exec.ExitError)(nil); errors.As(err, &exit) {
		return nil, false, nil
	}
	return data, err == nil, err
}

// The table HEAD commits; nil when root is no git checkout or HEAD holds no table.
func headTable(root string) (*siml.Names, error) {
	if !IsCheckout(root) {
		return nil, nil
	}
	data, ok, err := GitFile(root, "HEAD", TablePath)
	if err != nil || !ok {
		// A checkout with no commit yet has no HEAD to read.
		return nil, nil
	}
	head, err := siml.ParseNames(data)
	if err != nil {
		return nil, fmt.Errorf("%s at HEAD: %w", TablePath, err)
	}
	return head, nil
}

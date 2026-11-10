package gen

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/siml"
)

// The committed table, relative to the repository's root.
const TablePath = "sim/siml/names.json"

// The command that regenerates the committed table.
const MakeCommand = "make siml-names"

// Reads the table under root, collects what the sim and the database hold, and returns the table the one under
// root grows into, with pin handed back to the new-name rules when it is not nil. In a git checkout, the entries
// whose names HEAD's table does not hold and that the sim and the database no longer give are dropped first, so a
// name added during work in progress does not outlive what it named.
func Generate(root string, pin *Pin) (*siml.Names, *Report, error) {
	data, err := os.ReadFile(filepath.Join(root, TablePath))
	if err != nil {
		return nil, nil, err
	}
	old, err := siml.ParseNames(data)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w; it is generated, so take the committed one and run %s", TablePath, err, MakeCommand)
	}
	head, err := headTable(root)
	if err != nil {
		return nil, nil, err
	}
	c, err := Collect(root, old)
	if err != nil {
		return nil, nil, err
	}
	var dropped []string
	if head != nil {
		dropped = prune(old, head, c)
	}
	names, report, err := Update(old, c, pin)
	if report != nil {
		report.Dropped = dropped
	}
	return names, report, err
}

// The .siml files under root/ui/specs that write the name pin holds and mean its entry by it, relative to root: every
// committed file whose meaning a repin of the name can change. The files are read with the committed table, and a
// class's name is meant only in a file of that class.
func FilesUsing(root string, pin Pin) ([]string, error) {
	data, err := os.ReadFile(filepath.Join(root, TablePath))
	if err != nil {
		return nil, err
	}
	names, err := siml.ParseNames(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", TablePath, err)
	}
	var files []string
	err = filepath.WalkDir(filepath.Join(root, "ui", "specs"), func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".siml") {
			return err
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		f, err := siml.Parse(path, src, names, os.ReadFile)
		if diags := (siml.Diagnostics{}); err != nil && !errors.As(err, &diags) {
			return err
		}
		if !means(f, pin, names) {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	return files, err
}

// Whether f writes the name pin holds, and means pin's entry by it with names: a shared name means a class's own
// entry of that name in a file of the class.
func means(f *siml.File, pin Pin, names *siml.Names) bool {
	info, ok := siml.SpecOf(f.Spec)
	if pin.Class != proto.Class_ClassUnknown && (!ok || info.Class != pin.Class) {
		return false
	}
	if ok && pin.Class == proto.Class_ClassUnknown && pin.Namespace.SpellLike() {
		if _, own := names.Classes[info.Class][pin.Namespace][pin.Name]; own {
			return false
		}
	}
	return slices.ContainsFunc(f.Names, func(r siml.NameRef) bool { return r.Name == pin.Name && r.Namespace == pin.Namespace })
}

// Writes the report: the entries added, or their count by scope unless verbose; the new class entries a shared name
// holds the plain name of; the repinned names; the dropped names; the stale pinned names; and what could not be
// built, or its count unless verbose.
func (r *Report) Write(w io.Writer, verbose bool) {
	if verbose || len(r.Added) <= 50 {
		for _, line := range r.Added {
			fmt.Fprintln(w, "added", line)
		}
	} else {
		counts := map[string]int{}
		var scopes []string
		for _, line := range r.Added {
			scope := line[:strings.LastIndexByte(line[:strings.Index(line, " = ")], '.')]
			if counts[scope] == 0 {
				scopes = append(scopes, scope)
			}
			counts[scope]++
		}
		for _, scope := range scopes {
			fmt.Fprintf(w, "added %d names to %s\n", counts[scope], scope)
		}
		fmt.Fprintln(w, "(-v lists them)")
	}
	for _, line := range r.Qualified {
		fmt.Fprintln(w, "qualified:", line)
	}
	for _, p := range r.Repinned {
		fmt.Fprintln(w, "repinned", p)
	}
	for _, line := range r.Dropped {
		fmt.Fprintln(w, "dropped", line, "(HEAD's table does not hold it, and nothing gives it now)")
	}
	for _, line := range r.Stale {
		fmt.Fprintln(w, "stale:", line)
	}
	if r.Unregistered > 0 {
		fmt.Fprintf(w, "%d pinned spell and aura names are registered by no build\n", r.Unregistered)
	}
	if verbose {
		for _, line := range r.Skipped {
			fmt.Fprintln(w, "skipped", line)
		}
	} else if len(r.Skipped) > 0 {
		fmt.Fprintf(w, "skipped %d specs and sources the sim cannot build (-v lists them); a spec that does not build adds none of its spell and aura names until it does\n", len(r.Skipped))
	}
}

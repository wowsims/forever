// Generates the SIML name table (sim/siml/SPEC.md §7.4) from what the sim registers and the database, keeping
// every name the committed table holds, and reports what it added on standard error.
//
//	go run --tags=with_db ./sim/siml/cmd/simlnames -w (or make siml-names)
//	go run --tags=with_db ./sim/siml/cmd/simlnames -w -repin warrior.aura.overpower
//	go run ./sim/siml/cmd/simlnames -base origin/master
//
// With -base it generates nothing: it checks that every name of the table at that commit means what it meant there,
// unless the repin record gains the name's line over that commit's record, and exits 1 naming each one that does not.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wowsims/forever/sim/siml"
	"github.com/wowsims/forever/sim/siml/gen"
)

var (
	write   = flag.Bool("w", false, "write the table to "+gen.TablePath+" instead of standard output")
	repin   = flag.String("repin", "", "hand <class>.<namespace>.<name>, shared.<namespace>.<name>, item.<name>, enchant.<name> or suffix.<name> back to the new-name rules, with the other names its game name gives")
	verbose = flag.Bool("v", false, "list every added entry")
	base    = flag.String("base", "", "check the table against the one this commit holds, and write nothing")
)

func main() {
	flag.Parse()
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "simlnames: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	root, err := repositoryRoot()
	if err != nil {
		return err
	}
	if *base != "" {
		if *write || *repin != "" {
			return errors.New("-base checks the table and writes nothing, so it takes neither -w nor -repin")
		}
		return checkBase(root, *base)
	}
	var pin *gen.Pin
	if *repin != "" {
		p, err := gen.ParsePin(*repin)
		if err != nil {
			return err
		}
		pin = &p
	}
	table, report, err := gen.Generate(root, pin)
	if err != nil {
		return err
	}
	report.Write(os.Stderr, *verbose)
	for _, p := range report.Repinned {
		files, err := gen.FilesUsing(root, p.Pin)
		if err != nil {
			return err
		}
		for _, f := range files {
			fmt.Fprintf(os.Stderr, "%s uses %s, whose meaning the repin changes\n", f, p.Name)
		}
	}
	data := table.Encode()
	if !*write {
		_, err := os.Stdout.Write(data)
		return err
	}
	if len(report.Repinned) > 0 {
		path := filepath.Join(root, gen.RecordPath)
		record, err := os.ReadFile(path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if err := replace(path, gen.AppendRecord(record, report.Repinned), "repinned-*.txt"); err != nil {
			return err
		}
	}
	return replace(filepath.Join(root, gen.TablePath), data, "names-*.json")
}

// Writes data to path. Left alone when it holds data already, so what is built from it is not rebuilt. Otherwise
// written whole to a temporary file matching pattern and renamed into place, so a failed run leaves it as it was.
func replace(path string, data []byte, pattern string) error {
	if old, err := os.ReadFile(path); err == nil && bytes.Equal(old, data) {
		return nil
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), pattern)
	if err != nil {
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// Fails, naming each, when a name of the table at commit rev means something else, or nothing, in the table under
// root, and the repin record under root does not add its line over rev's record.
func checkBase(root, rev string) error {
	if !gen.IsCheckout(root) {
		return fmt.Errorf("-base needs a git checkout, and %s is none", root)
	}
	baseData, ok, err := gen.GitFile(root, rev, gen.TablePath)
	if err != nil || !ok {
		return err
	}
	baseTable, err := siml.ParseNames(baseData)
	if err != nil {
		return fmt.Errorf("%s at %s: %w", gen.TablePath, rev, err)
	}
	data, err := os.ReadFile(filepath.Join(root, gen.TablePath))
	if err != nil {
		return err
	}
	table, err := siml.ParseNames(data)
	if err != nil {
		return fmt.Errorf("%s: %w", gen.TablePath, err)
	}
	baseRecord, _, err := gen.GitFile(root, rev, gen.RecordPath)
	if err != nil {
		return err
	}
	record, err := os.ReadFile(filepath.Join(root, gen.RecordPath))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	changed := gen.Changed(baseTable, table, gen.RecordedNames(record, baseRecord))
	if len(changed) == 0 {
		return nil
	}
	for _, line := range changed {
		fmt.Fprintln(os.Stderr, line)
	}
	return fmt.Errorf("%d names of %s at %s mean something else now. A name keeps its meaning: take %s's entries back, "+
		"or, to give a name another meaning on purpose, run simlnames -w -repin <name>, which adds its line to %s",
		len(changed), gen.TablePath, rev, rev, gen.RecordPath)
}

// The directory that holds go.mod, from the working directory up.
func repositoryRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("no go.mod above the working directory")
		}
		dir = parent
	}
}

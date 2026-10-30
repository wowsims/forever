package cmd

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/siml"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
)

// The siml commands read and write SIML files (sim/siml/SPEC.md) with the committed name table, which a build holds
// only with the with_db tag. Their flags are Go flags, written -check or --check as the go tools take them.

var simlCmd = &cobra.Command{
	Use:   "siml",
	Short: "read, print, format, verify and sim SIML files",
}

func init() {
	simlCmd.AddCommand(
		simlSubcommand("parse [-json] <file>", "check that a .siml file loads, or print what it loads as protojson", simlParse),
		simlSubcommand("print -kind apl|gear|siml [-spec <field>] [-categories <letters>] <file.json>", "print a rotation, gear set or settings JSON file as SIML", simlPrint),
		simlSubcommand("fmt [-check | -w] <files or folders>", "format .siml files, or check that they are in canonical form", simlFmt),
		simlSubcommand("verify <old.json> <new.siml>", "check that a .siml file reads back as the JSON file it was printed from", simlVerify),
		simlSubcommand("sim < request.json", "sim the text of an .apl.siml for the APL editor: a JSON request on stdin, a JSON result on stdout", simlSim),
		simlSubcommand("serve [-names <names.json>]", "read .siml texts for the Vite plugin: one JSON request a line on stdin, one JSON response a line on stdout", simlServe),
		simlSubcommand("lsp", "serve the SIML language server on stdin and stdout", simlLSP),
		simlSubcommand("grammar -o <file>", "write the TextMate grammar of SIML, built from the schema", simlGrammar),
		simlSubcommand("llm [-o <file>]", "write the wowsims-siml skill, a short guide to SIML with every kind the protos give", simlLLM),
	)
}

// A siml subcommand, whose run parses its own arguments with the flag package.
func simlSubcommand(use, short string, run func(out, errOut io.Writer, args []string) error) *cobra.Command {
	return &cobra.Command{
		Use:                use,
		Short:              short,
		DisableFlagParsing: true,
		SilenceUsage:       true,
		SilenceErrors:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := run(cmd.OutOrStdout(), cmd.ErrOrStderr(), args); !errors.Is(err, flag.ErrHelp) {
				return err
			}
			return nil
		},
	}
}

func newFlags(usage string, errOut io.Writer) *flag.FlagSet {
	flags := flag.NewFlagSet("siml", flag.ContinueOnError)
	flags.SetOutput(errOut)
	flags.Usage = func() {
		fmt.Fprintf(errOut, "usage: wowsimcli siml %s\n", usage)
		flags.PrintDefaults()
	}
	return flags
}

// Parses args into flags, which must leave n arguments, or at least one when n is -1.
func parseFlags(flags *flag.FlagSet, args []string, n int) error {
	if err := flags.Parse(args); err != nil {
		return err
	}
	if n == -1 && flags.NArg() == 0 || n >= 0 && flags.NArg() != n {
		flags.Usage()
		return errors.New("wrong number of arguments")
	}
	return nil
}

// The JSON file at path, read as m.
func readJSON(path string, m googleProto.Message) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := protojson.Unmarshal(data, m); err != nil {
		return fmt.Errorf("%s does not read as %s: %w", path, googleProto.MessageName(m).Name(), err)
	}
	return nil
}

// The spec whose spec line writes field.
func specNamed(field string) (siml.SpecInfo, error) {
	if info, ok := siml.SpecNamed(field); ok {
		return info, nil
	}
	var fields []string
	for _, info := range siml.Specs() {
		fields = append(fields, string(info.Field.Name()))
	}
	return siml.SpecInfo{}, fmt.Errorf("no spec is named %q; the specs are %s", field, strings.Join(fields, ", "))
}

// The diagnostics of a parse of the file at path that fail its check: every error, and for a file under ui/specs its
// warnings and raw refs too; and the warnings that do not.
func failures(path string, f *siml.File, err error) (fail, warn siml.Diagnostics, _ error) {
	var diags siml.Diagnostics
	switch {
	case errors.As(err, &diags):
	case err != nil:
		return nil, nil, err
	default:
		diags = f.Warnings
	}
	repository := siml.UnderSpecs(path)
	for _, d := range diags {
		if d.Severity == siml.Error || repository {
			fail = append(fail, d)
		} else {
			warn = append(warn, d)
		}
	}
	if repository && f != nil && len(f.RawRefs) > 0 {
		fail = append(fail, f.RawRefs...)
		slices.SortStableFunc(fail, func(a, b siml.Diagnostic) int {
			return cmp.Or(a.Start.Line-b.Start.Line, a.Start.Col-b.Start.Col)
		})
	}
	return fail, warn, nil
}

// A diagnostic as a line: "path:line:col code message", and "(warning)" after a warning.
func diagnosticLine(d siml.Diagnostic) string {
	if d.Severity == siml.Warning {
		return d.String() + " (warning)"
	}
	return d.String()
}

func printDiagnostics(w io.Writer, diags siml.Diagnostics) {
	for _, d := range diags {
		fmt.Fprintln(w, diagnosticLine(d))
	}
}

func simlParse(out, errOut io.Writer, args []string) error {
	flags := newFlags("parse [-json] <file>", errOut)
	asJSON := flags.Bool("json", false, "print the message the file loads as, as protojson")
	if err := parseFlags(flags, args, 1); err != nil {
		return err
	}
	path := flags.Arg(0)
	f, err := siml.ReadFile(path)
	fail, warn, err := failures(path, f, err)
	if err != nil {
		return err
	}
	printDiagnostics(errOut, warn)
	printDiagnostics(errOut, fail)
	if f.Settings != nil {
		if *asJSON {
			data, err := protojson.Marshal(f.Message())
			if err != nil {
				return err
			}
			var indented bytes.Buffer
			if err := json.Indent(&indented, data, "", "  "); err != nil {
				return err
			}
			fmt.Fprintln(out, indented.String())
		} else {
			fmt.Fprintf(out, "%s: loads as %s\n", path, loadsAs(f))
		}
	}
	switch {
	case f.Settings == nil:
		return fmt.Errorf("%s does not load", path)
	case len(fail) > 0:
		return fmt.Errorf("%s fails the check of the files under ui/specs", path)
	}
	return nil
}

// What a loaded file is, for parse to say: "a rotation for hunter", "a gear set", "settings for hunter with the
// categories g, t, r".
func loadsAs(f *siml.File) string {
	text, part := map[string]string{"apl": "a rotation", "gear": "a gear set"}[f.Kind.Section]
	if !part {
		text = "settings"
	}
	if info, ok := siml.SpecOf(f.Spec); ok {
		text += " for " + string(info.Field.Name())
	}
	switch {
	case part:
	case len(f.Categories) == 0:
		text += " with no category"
	default:
		text += " with the categories " + strings.Join(f.Categories, ", ")
	}
	return text
}

const categoryLetters = "gtrcmxeu" // The site's import categories (SPEC.md §5.1).

func simlPrint(out, errOut io.Writer, args []string) error {
	flags := newFlags("print -kind apl|gear|siml [-spec <field>] [-categories <letters>] <file.json>", errOut)
	kindName := flags.String("kind", "", "what the JSON file holds: apl, an APLRotation printed as an x.apl.siml; gear, an EquipmentSpec printed as an x.gear.siml; siml, IndividualSimSettings printed as a full x.siml")
	specName := flags.String("spec", "", "the spec the file is for, as its spec line writes it (hunter, dps_warrior): a rotation needs it; a full setup takes its player's; a gear file writes none")
	letters := flags.String("categories", "", "with -kind siml, the import categories ("+categoryLetters+") whose sections to write, an empty section for one that holds nothing; without it, each section that holds something")
	if err := parseFlags(flags, args, 1); err != nil {
		return err
	}
	printed, ok := map[string]string{"apl": "print.apl.siml", "gear": "print.gear.siml", "siml": "print.siml"}[*kindName]
	if !ok {
		return errors.New("-kind is apl, gear or siml")
	}
	kind, err := siml.KindOf(printed)
	if err != nil {
		return err
	}
	spec := proto.Spec_SpecUnknown
	if *specName != "" {
		info, err := specNamed(*specName)
		if err != nil {
			return err
		}
		spec = info.Spec
	} else if *kindName == "apl" {
		return errors.New("-kind apl needs -spec: an x.apl.siml writes its spec line, and its names are its class's")
	}
	var categories []string
	if isSet(flags, "categories") {
		if *kindName != "siml" {
			return errors.New("-categories is for -kind siml")
		}
		if *letters == "" {
			return errors.New("-categories needs a category letter")
		}
		for _, c := range strings.Split(*letters, "") {
			if !strings.Contains(categoryLetters, c) || slices.Contains(categories, c) {
				return fmt.Errorf("-categories takes each of %s at most once, not %q", categoryLetters, *letters)
			}
			categories = append(categories, c)
		}
	}
	names, err := siml.DefaultNames()
	if err != nil {
		return err
	}

	path := flags.Arg(0)
	m := kind.Message()
	if err := readJSON(path, m); err != nil {
		return err
	}
	settings := siml.SettingsOf(m)
	if held, ok := siml.SpecOf(heldSpec(settings)); ok && spec != proto.Spec_SpecUnknown && held.Spec != spec {
		return fmt.Errorf("%s holds a %s player, not %s", path, held.Field.Name(), *specName)
	}
	text := siml.Print(kind, spec, settings, categories, names)

	// The printer checks that each line reads back, which misses a field SIML writes nowhere. Without -categories the
	// whole print reads back as the JSON.
	f, err := siml.Parse(printed, text, names, nil)
	if err != nil {
		return fmt.Errorf("the print of %s does not read back, which is a bug of the printer:\n%w", path, err)
	}
	if diffs := readBackDifferences(m, f, names); categories == nil && len(diffs) > 0 {
		return fmt.Errorf("the print of %s loses what SIML has no form for, which is a bug of the printer (sim/siml/SPEC.md §12.4):\n  %s",
			path, strings.Join(diffs, "\n  "))
	}
	for _, r := range f.RawRefs {
		fmt.Fprintf(errOut, "%s: the print writes raw JSON at %d:%d, where the name table has no name\n", path, r.Start.Line, r.Start.Col)
	}
	_, err = out.Write(text)
	return err
}

// Whether the flag name is given on the command line.
func isSet(flags *flag.FlagSet, name string) bool {
	set := false
	flags.Visit(func(f *flag.Flag) { set = set || f.Name == name })
	return set
}

// The spec whose message the settings' player holds, SpecUnknown for none.
func heldSpec(s *proto.IndividualSimSettings) proto.Spec {
	if fd := specField(s.GetPlayer()); fd != nil {
		if info, ok := siml.SpecNamed(string(fd.Name())); ok {
			return info.Spec
		}
	}
	return proto.Spec_SpecUnknown
}

func simlVerify(out, errOut io.Writer, args []string) error {
	flags := newFlags("verify <old.json> <new.siml>", errOut)
	if err := parseFlags(flags, args, 2); err != nil {
		return err
	}
	oldPath, newPath := flags.Arg(0), flags.Arg(1)
	kind, err := siml.KindOf(newPath)
	if err != nil {
		return err
	}
	old := kind.Message()
	if err := readJSON(oldPath, old); err != nil {
		return err
	}
	names, err := siml.DefaultNames()
	if err != nil {
		return err
	}
	f, err := siml.ReadFile(newPath)
	if diags := (siml.Diagnostics{}); errors.As(err, &diags) {
		printDiagnostics(errOut, diags)
		return fmt.Errorf("%s does not load", newPath)
	} else if err != nil {
		return err
	}
	printDiagnostics(errOut, f.Warnings)

	diffs := readBackDifferences(old, f, names)
	if len(diffs) == 0 {
		fmt.Fprintf(out, "%s reads back as %s\n", newPath, oldPath)
		return nil
	}
	fmt.Fprintf(out, "%s does not read back as %s. Where they differ after the normalisation of SPEC.md §12.4, by JSON path:\n", newPath, oldPath)
	for _, d := range diffs {
		fmt.Fprintln(out, "  "+d)
	}
	return fmt.Errorf("%s differs from %s", newPath, oldPath)
}

// Where f reads back differently from the JSON message want, after the normalisation of SPEC.md §12.4: one
// "path: <JSON value> in the JSON, <its value> in the SIML" for each difference, by JSON path.
func readBackDifferences(want googleProto.Message, f *siml.File, names *siml.Names) []string {
	a, b := siml.Normalize(want, names, f.Spec), siml.Normalize(f.Message(), names, f.Spec)
	var lines []string
	for _, d := range messageDiff(a.ProtoReflect(), b.ProtoReflect()) {
		lines = append(lines, fmt.Sprintf("%s: %s in the JSON, %s in the SIML", d.path, d.old, d.new))
	}
	return lines
}

// The database catalog of the repository the file at path sits in, which the checks of files under ui/specs read.
func repositoryCatalog(path string) (*siml.DatabaseCatalog, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	root, ok := siml.RepositoryRoot(abs)
	if !ok {
		return nil, fmt.Errorf("%s: no repository holds the file, so no database checks it", path)
	}
	return siml.CatalogAt(root), nil
}

func simlFmt(out, errOut io.Writer, args []string) error {
	flags := newFlags("fmt [-check | -w] <files or folders>", errOut)
	check := flags.Bool("check", false, "print each file that is not in canonical form or has an error, a file under ui/specs a warning, a raw ref or a name that no longer holds too, with what it has, and fail if there is one")
	write := flags.Bool("w", false, "write the canonical form to each file that is not in it")
	if err := parseFlags(flags, args, -1); err != nil {
		return err
	}
	if *check && *write {
		return errors.New("-check and -w do not go together")
	}
	files, err := simlFiles(flags.Args())
	if err != nil {
		return err
	}
	names, err := siml.DefaultNames()
	if err != nil {
		return err
	}
	failed := 0
	for _, path := range files {
		src, err := siml.ReadSource(path)
		if err != nil {
			return err
		}
		doc := siml.Analyze(path, src, names, siml.ReadSource)
		f, err := doc.File()
		fail, warn, err := failures(path, f, err)
		if err != nil {
			return err
		}
		var problems []string
		if len(fail) == 0 && siml.UnderSpecs(path) {
			catalog, err := repositoryCatalog(path)
			if err != nil {
				return err
			}
			problems = repositoryProblems(doc, catalog, path, f)
		}
		// A file that Format finds an error in has that error in its parse too, so it fails.
		formatted, err := siml.Format(path, src, names, siml.ReadSource)
		if diags := (siml.Diagnostics{}); err != nil && !errors.As(err, &diags) {
			return err
		}
		changed := err == nil && !bytes.Equal(formatted, src)
		if !*check {
			printDiagnostics(errOut, warn)
		}
		printDiagnostics(errOut, fail)
		for _, p := range problems {
			fmt.Fprintln(errOut, p)
		}
		if *check && changed {
			fmt.Fprintf(out, "%s: not in canonical form\n", path)
		}
		if len(fail) > 0 || len(problems) > 0 || *check && changed {
			failed++
			continue
		}
		switch {
		case *write && changed:
			if err := os.WriteFile(path, formatted, 0o644); err != nil {
				return err
			}
		case !*check && !*write:
			if _, err := out.Write(formatted); err != nil {
				return err
			}
		}
	}
	switch {
	case failed > 0 && *check:
		return fmt.Errorf("%d of %d .siml files are not in canonical form or fail the check", failed, len(files))
	case failed > 0:
		return fmt.Errorf("%d of %d .siml files fail the check", failed, len(files))
	}
	return nil
}

// The .siml files args name: each file, and every .siml under each folder, skipping node_modules and hidden folders.
func simlFiles(args []string) ([]string, error) {
	var files []string
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			if !strings.HasSuffix(arg, ".siml") {
				return nil, fmt.Errorf("%s is not a .siml file", arg)
			}
			files = append(files, arg)
			continue
		}
		err = filepath.WalkDir(arg, func(path string, d fs.DirEntry, err error) error {
			switch {
			case err != nil:
				return err
			case d.IsDir() && path != arg && (d.Name() == "node_modules" || strings.HasPrefix(d.Name(), ".")):
				return filepath.SkipDir
			case !d.IsDir() && strings.HasSuffix(path, ".siml"):
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no .siml file in %s", strings.Join(args, ", "))
	}
	return files, nil
}

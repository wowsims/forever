package siml

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
)

// The conformance corpus of SPEC.md §15 in testdata.

// The case files under a corpus folder whose names end in suffix (SPEC.md §15). In a nested folder, a multi-file
// case, only main.<kind>.siml is a case; the other files are the ones it uses.
func corpusCases(t *testing.T, folder, suffix string) []string {
	t.Helper()
	root := filepath.Join("testdata", folder)
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		nested := filepath.Dir(path) != root
		if err == nil && !d.IsDir() && strings.HasSuffix(path, suffix) && (!nested || strings.HasPrefix(d.Name(), "main.")) {
			files = append(files, path)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatalf("%s holds no *%s case", root, suffix)
	}
	slices.Sort(files)
	return files
}

func caseName(file string) string {
	name, _ := filepath.Rel("testdata", file)
	return filepath.ToSlash(name)
}

// How SPEC.md §15 reads a case: with file access, except an error case directly in errors/, which is read as a host
// without it.
func caseOpener(file string) Opener {
	if filepath.Base(filepath.Dir(file)) == "errors" {
		return nil
	}
	return os.ReadFile
}

func parseCase(t *testing.T, file string) (*File, error) {
	t.Helper()
	src, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	return Parse(file, src, corpusNames(t), caseOpener(file))
}

// The diagnostics of a parse, warnings included.
func diagnosticsOf(t *testing.T, f *File, err error) Diagnostics {
	t.Helper()
	var diags Diagnostics
	switch {
	case err == nil:
		diags = f.Warnings
	case !errors.As(err, &diags):
		t.Fatal(err)
	}
	return diags
}

// Diagnostics as the lines of a .diag file: "LINE:COL severity code" (SPEC.md §15).
func diagnosticLines(diags Diagnostics) []string {
	lines := make([]string, len(diags))
	for i, d := range diags {
		lines[i] = fmt.Sprintf("%d:%d %s %s", d.Start.Line, d.Start.Col, d.Severity, d.Code)
	}
	return lines
}

// The JSON file at path, read as the loader's message for kind.
func readMessage(t *testing.T, path string, kind Kind) googleProto.Message {
	t.Helper()
	text, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	m := kind.Message()
	if err := protojson.Unmarshal(text, m); err != nil {
		t.Fatalf("%s does not read as %T: %v", path, m, err)
	}
	return m
}

// The raw refs of a case are those its .raw file lists, as a .diag file lists diagnostics, when it has one.
func checkRawRefs(t *testing.T, file string, f *File) {
	t.Helper()
	want, err := os.ReadFile(file + ".raw")
	if errors.Is(err, fs.ErrNotExist) {
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if got := diagnosticLines(f.RawRefs); !slices.Equal(got, nonEmptyLines(string(want))) {
		t.Errorf("raw refs differ\ngot:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.TrimSpace(string(want)))
	}
}

// SPEC.md §15 item 1: every parse case reads to its expected JSON, with no error.
func TestParseCases(t *testing.T) {
	for _, file := range corpusCases(t, "parse", ".siml") {
		t.Run(caseName(file), func(t *testing.T) {
			f, err := parseCase(t, file)
			if err != nil {
				t.Fatal(err)
			}
			checkRawRefs(t, file, f)
			want := readMessage(t, strings.TrimSuffix(file, ".siml")+".json", f.Kind)
			if got := f.Message(); !googleProto.Equal(got, want) {
				t.Errorf("the parse differs from the expected JSON\ngot:  %s\nwant: %s", protojson.Format(got), protojson.Format(want))
			}
		})
	}
}

// SPEC.md §15 item 4: every errors case gives exactly its listed diagnostics, in order.
func TestErrorCases(t *testing.T) {
	for _, file := range corpusCases(t, "errors", ".siml") {
		t.Run(caseName(file), func(t *testing.T) {
			want, err := os.ReadFile(file + ".diag")
			if err != nil {
				t.Fatal(err)
			}
			f, err := parseCase(t, file)
			diags := diagnosticsOf(t, f, err)
			if got := diagnosticLines(diags); !slices.Equal(got, nonEmptyLines(string(want))) {
				t.Errorf("diagnostics differ\ngot:\n%s\nwant:\n%s\nmessages:\n%v", strings.Join(got, "\n"), strings.TrimSpace(string(want)), diags)
			}
			checkRawRefs(t, file, f)
		})
	}
}

// The table at path, in the form of SPEC.md §7.4.
func readNames(t *testing.T, path string) *Names {
	t.Helper()
	text, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	names, err := ParseNames(text)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return names
}

// SPEC.md §15 item 7: every table in tables/refused is refused, and every table in tables/accepted loads.
func TestNameTables(t *testing.T) {
	for folder, refused := range map[string]bool{"refused": true, "accepted": false} {
		files, err := filepath.Glob(filepath.Join("testdata", "tables", folder, "*.json"))
		if err != nil || len(files) == 0 {
			t.Fatalf("testdata/tables/%s holds no table (%v)", folder, err)
		}
		for _, file := range files {
			t.Run(caseName(file), func(t *testing.T) {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := ParseNames(data); refused && err == nil {
					t.Error("the table reads")
				} else if !refused && err != nil {
					t.Error(err)
				}
			})
		}
	}
}

// The fixture table every corpus case is read and printed with (SPEC.md §15).
func corpusNames(t *testing.T) *Names {
	t.Helper()
	return readNames(t, filepath.Join("testdata", "names.json"))
}

// Prints r as an x.apl.siml for spec.
func printRotation(spec proto.Spec, r *proto.APLRotation, names *Names) []byte {
	return Print(Kind{Section: sectionAPL}, spec, &proto.IndividualSimSettings{Player: &proto.Player{Rotation: r}}, nil, names)
}

// Prints e as an x.gear.siml.
func printGear(e *proto.EquipmentSpec, names *Names) []byte {
	return Print(Kind{Section: sectionGear}, proto.Spec_SpecUnknown, &proto.IndividualSimSettings{Player: &proto.Player{Equipment: e}}, nil, names)
}

// A print case's spec: its name starts with the spec's field (SPEC.md §15), the longest that fits.
func caseSpec(file string) proto.Spec {
	base := filepath.Base(file)
	spec, length := proto.Spec_SpecUnknown, 0
	for _, info := range Specs() {
		field := string(info.Field.Name())
		if (strings.HasPrefix(base, field+"_") || strings.HasPrefix(base, field+".")) && len(field) > length {
			spec, length = info.Spec, len(field)
		}
	}
	return spec
}

// The categories a print case is printed with: the letters of its .categories file, or nil when it has none.
func caseCategories(file string) []string {
	letters, err := os.ReadFile(strings.TrimSuffix(file, ".json") + ".categories")
	if err != nil {
		return nil
	}
	return strings.Split(strings.TrimSpace(string(letters)), "")
}

// SPEC.md §15 item 2: every print case prints to its expected text exactly, and reads back to its input, or with
// categories to a file holding exactly those.
func TestPrintCases(t *testing.T) {
	names := corpusNames(t)
	for _, file := range corpusCases(t, "print", ".json") {
		t.Run(caseName(file), func(t *testing.T) {
			target := strings.TrimSuffix(file, ".json") + ".siml"
			kind := kindOf(target)
			input := readMessage(t, file, kind)
			categories := caseCategories(file)
			got := Print(kind, caseSpec(file), SettingsOf(input), categories, names)
			want, err := os.ReadFile(target)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != string(want) {
				t.Fatalf("the print differs\ngot:\n%s\nwant:\n%s", got, want)
			}
			f, err := Parse(target, got, names, nil)
			if err != nil {
				t.Fatal(err)
			}
			if categories != nil {
				if !slices.Equal(f.Categories, categories) {
					t.Errorf("the print holds the categories %v, want %v", f.Categories, categories)
				}
			} else if back, in := Normalize(f.Message(), names, f.Spec), Normalize(input, names, f.Spec); !googleProto.Equal(back, in) {
				t.Errorf("the print reads back differently\ngot:  %s\nwant: %s", protojson.Format(back), protojson.Format(in))
			}
		})
	}
}

// SPEC.md §15 item 3: every format case formats to its .want exactly.
func TestFormatCases(t *testing.T) {
	names := corpusNames(t)
	for _, file := range corpusCases(t, "format", ".siml") {
		t.Run(caseName(file), func(t *testing.T) {
			src, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(file + ".want")
			if err != nil {
				t.Fatal(err)
			}
			got, err := Format(file, src, names, caseOpener(file))
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != string(want) {
				t.Errorf("the format differs\ngot:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}

// SPEC.md §15 item 5, for every parse, print and format case: format reaches a fixed point at once, and the print of
// a parse reads back to that parse.
func TestCorpusRoundTrips(t *testing.T) {
	names := corpusNames(t)
	files := slices.Concat(corpusCases(t, "parse", ".siml"), corpusCases(t, "print", ".siml"), corpusCases(t, "format", ".siml"))
	for _, file := range files {
		t.Run(caseName(file), func(t *testing.T) {
			src, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			once, err := Format(file, src, names, caseOpener(file))
			if err != nil {
				t.Fatal(err)
			}
			if twice, err := Format(file, once, names, caseOpener(file)); err != nil || string(twice) != string(once) {
				t.Errorf("format is no fixed point (%v)\nonce:\n%s\ntwice:\n%s", err, once, twice)
			}
			f, err := parseCase(t, file)
			if err != nil {
				t.Fatal(err)
			}
			printed := Print(f.Kind, f.Spec, f.Settings, nil, names)
			back, err := Parse(file, printed, names, nil)
			if err != nil {
				t.Fatalf("the print does not read: %v\n%s", err, printed)
			}
			if got, want := Normalize(back.Message(), names, f.Spec), Normalize(f.Message(), names, f.Spec); !googleProto.Equal(got, want) {
				t.Errorf("the print reads back differently\n%s\ngot:  %s\nwant: %s", printed, protojson.Format(got), protojson.Format(want))
			}
		})
	}
}

func nonEmptyLines(text string) []string {
	return slices.DeleteFunc(strings.Split(text, "\n"), func(line string) bool { return line == "" })
}

// A file SPEC.md §16 gives as an example.
type specExample struct{ path, text string }

var (
	exampleName  = regexp.MustCompile("`([^`]*\\.siml)`")
	exampleBlock = regexp.MustCompile("(?s)\n```\n(.*?)```\n")
	exampleKind  = regexp.MustCompile(`^(.*?)((?:\.[a-z_]+)?\.siml)$`)
)

// The examples of SPEC.md §16, as the files they stand for: each code block under a `### 16.` heading. The first
// block takes the path the heading gives, and a further block one beside it, so that a `use` of the first reads it.
func specExamples(t *testing.T) []specExample {
	t.Helper()
	spec, err := os.ReadFile("SPEC.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(spec)
	text = text[strings.Index(text, "\n## 16. "):]
	var out []specExample
	for _, part := range strings.Split(text, "\n### 16.")[1:] {
		heading, _, _ := strings.Cut(part, "\n")
		m := exampleName.FindStringSubmatch(heading)
		if m == nil {
			t.Fatalf("the heading %q names no .siml file", heading)
		}
		file := m[1]
		if strings.HasPrefix(file, ".") {
			file = "example" + file
		}
		for i, block := range exampleBlock.FindAllStringSubmatch(part, -1) {
			path := file
			if i > 0 {
				path = exampleKind.ReplaceAllString(file, fmt.Sprintf("${1}_%d${2}", i+1))
			}
			out = append(out, specExample{path, block[1]})
		}
	}
	return out
}

// Every example of SPEC.md §16 reads with the committed name table, with no diagnostic and no raw ref (§7.3).
func TestSpecExamples(t *testing.T) {
	examples := specExamples(t)
	var paths []string
	for _, e := range examples {
		paths = append(paths, e.path)
	}
	want := []string{
		"ui/specs/hunter/dps/apls/default.apl.siml",
		"ui/specs/hunter/dps/apls/default_2.apl.siml",
		"arms.apl.siml",
		"ui/specs/hunter/dps/gear_sets/p1.gear.siml",
		"example.siml",
	}
	if !slices.Equal(paths, want) {
		t.Fatalf("the examples are %v, want %v", paths, want)
	}
	// An example's own text, else the committed file it uses, such as the hunter's apls/weave_common.apl.siml.
	open := func(path string) ([]byte, error) {
		for _, e := range examples {
			if e.path == filepath.ToSlash(path) {
				return []byte(e.text), nil
			}
		}
		return os.ReadFile(filepath.Join("..", "..", path))
	}
	names := readNames(t, "names.json")
	for _, e := range examples {
		t.Run(e.path, func(t *testing.T) {
			f, err := Parse(e.path, []byte(e.text), names, open)
			if diags := diagnosticsOf(t, f, err); len(diags) != 0 {
				t.Errorf("the diagnostics are\n%s", strings.Join(diagnosticLines(diags), "\n"))
			}
			if f != nil && len(f.RawRefs) != 0 {
				t.Errorf("the raw refs are %v", f.RawRefs)
			}
		})
	}
}

// Every diagnostic code of SPEC.md §13 that the package emits has an errors case, and every code an errors case gives
// is in the table.
func TestErrorCasesCoverEveryCode(t *testing.T) {
	spec, err := os.ReadFile("SPEC.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(spec)
	table := text[strings.Index(text, "## 13. Diagnostics"):strings.Index(text, "**Recovery.**")]
	codes := map[string]bool{}
	for _, line := range strings.Split(table, "\n") {
		if cells := strings.Split(line, "|"); strings.HasPrefix(line, "| `") {
			for _, m := range regexp.MustCompile("`([a-z-]+)`").FindAllStringSubmatch(cells[1], -1) {
				codes[m[1]] = true
			}
		}
	}
	sources, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	var code strings.Builder
	for _, file := range sources {
		if !strings.HasSuffix(file, "_test.go") {
			src, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			code.Write(src)
		}
	}
	covered := map[string]bool{}
	err = filepath.WalkDir(filepath.Join("testdata", "errors"), func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".diag") && !strings.HasSuffix(path, ".raw") {
			return err
		}
		src, err := os.ReadFile(path)
		for _, line := range nonEmptyLines(string(src)) {
			if fields := strings.Fields(line); len(fields) > 2 {
				covered[fields[2]] = true
			}
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	emitted := 0
	for c := range codes {
		if strings.Contains(code.String(), `"`+c+`"`) {
			emitted++
			if !covered[c] {
				t.Errorf("no errors case gives %s", c)
			}
		}
	}
	if emitted < 50 {
		t.Errorf("the package emits only %d of the codes of SPEC.md §13", emitted)
	}
	for c := range covered {
		if !codes[c] {
			t.Errorf("an errors case gives %s, which SPEC.md §13 does not list", c)
		}
	}
}

// The parse cases write every section, every form of statement, every short form and every other form of SPEC.md
// §8.3.
func TestParseCasesCoverTheLanguage(t *testing.T) {
	ix, err := loadIndex()
	if err != nil {
		t.Fatal(err)
	}
	sections, statements := map[string]bool{}, map[string]bool{}
	for _, file := range corpusCases(t, "parse", ".siml") {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		kind, err := KindOf(file)
		if err != nil {
			t.Fatal(err)
		}
		tree, _ := parseSyntax(sourceText(src), ix, kind)
		for _, s := range tree.sections {
			if s.header != nil || len(s.body) > 0 {
				sections[s.name] = true
			}
		}
		for _, st := range tree.inFileOrder() {
			statements[strings.TrimPrefix(fmt.Sprintf("%T", st), "*siml.")] = true
		}
	}
	for _, def := range ix.sectionList {
		if !sections[def.Name] {
			t.Errorf("no parse case writes the section %s", def.Name)
		}
	}
	for _, st := range []string{"versionStmt", "specStmt", "talentsStmt", "useStmt", "headerStmt", "varStmt", "rowStmt", "slotRowStmt",
		"flagListStmt", "fieldEntryStmt", "messageEntryStmt", "rawEntryStmt", "whenStmt", "defStmt"} {
		if !statements[st] {
			t.Errorf("no parse case writes a %s", st)
		}
	}

	var json strings.Builder
	for _, file := range corpusCases(t, "parse", ".json") {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		json.Write(src)
	}
	s, err := buildSchema()
	if err != nil {
		t.Fatal(err)
	}
	forms := map[string]bool{}
	for _, k := range slices.Concat(s.Kinds.Action, s.Kinds.Value) {
		switch written := strings.Contains(json.String(), `"`+k.JSON+`":`); {
		case written:
			forms[k.Form] = true
		case k.Form == "short":
			t.Errorf("no parse case writes the short form of %s", k.Kind)
		}
	}
	for _, form := range []string{"accessor", "bare", "unit_prefix", "list_args", "generic"} {
		if !forms[form] {
			t.Errorf("no parse case writes a kind of the %s form", form)
		}
	}
}

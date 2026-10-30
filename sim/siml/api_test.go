package siml

import (
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

// The site's requests, read with the committed name table as the site reads them.

const apiRotation = "siml 1\nspec hunter\n\npriority:\n  cast multi_shot\n"

func TestParseRequestRotation(t *testing.T) {
	r := ParseRequest(&proto.SimlParseRequest{Path: "rotation.apl.siml", Text: apiRotation}, readNames(t, "names.json"))
	if r.Error != "" || len(r.Diagnostics) > 0 {
		t.Fatalf("got error %q, diagnostics %v", r.Error, r.Diagnostics)
	}
	if r.Spec != proto.Spec_SpecHunter {
		t.Errorf("spec %v, want SpecHunter", r.Spec)
	}
	if strings.Join(r.Categories, "") != "r" {
		t.Errorf("categories %v, want [r]", r.Categories)
	}
	if n := len(r.Settings.GetPlayer().GetRotation().GetPriorityList()); n != 1 {
		t.Errorf("%d priority list items, want 1", n)
	}
}

func TestParseRequestGear(t *testing.T) {
	text := "siml 1\n\nhead lieutenant_commanders_chain_helmet\n"
	r := ParseRequest(&proto.SimlParseRequest{Path: "p1.gear.siml", Text: text}, readNames(t, "names.json"))
	if r.Error != "" || len(r.Diagnostics) > 0 {
		t.Fatalf("got error %q, diagnostics %v", r.Error, r.Diagnostics)
	}
	if strings.Join(r.Categories, "") != "g" {
		t.Errorf("categories %v, want [g]", r.Categories)
	}
	items := r.Settings.GetPlayer().GetEquipment().GetItems()
	if len(items) != 17 || items[0].Id == 0 {
		t.Errorf("items %v, want 17 with the head set", items)
	}
}

func TestParseRequestSetup(t *testing.T) {
	text := "siml 1\nspec hunter\ntalents \"-0053\"\n\nconsumables:\n\npriority:\n  cast multi_shot\n  group \"Nowhere\"\n"
	r := ParseRequest(&proto.SimlParseRequest{Path: "wowsims.siml", Text: text}, readNames(t, "names.json"))
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if got := strings.Join(r.Categories, ""); got != "trc" {
		t.Errorf("categories %q, want trc", got)
	}
	if r.Settings.GetPlayer().GetTalentsString() != "-0053" {
		t.Errorf("talents %q, want -0053", r.Settings.GetPlayer().GetTalentsString())
	}
	if len(r.Diagnostics) != 1 || !r.Diagnostics[0].Warning {
		t.Errorf("diagnostics %v, want one warning", r.Diagnostics)
	}
}

func TestParseRequestError(t *testing.T) {
	text := "siml 1\nspec hunter\n\npriority:\n  cast multi_shto\n"
	r := ParseRequest(&proto.SimlParseRequest{Path: "rotation.apl.siml", Text: text}, readNames(t, "names.json"))
	if r.Settings != nil || len(r.Categories) > 0 {
		t.Errorf("a text with an error gave settings %v, categories %v", r.Settings, r.Categories)
	}
	if len(r.Diagnostics) != 1 {
		t.Fatalf("diagnostics %v, want one", r.Diagnostics)
	}
	d := r.Diagnostics[0]
	if d.Warning || d.Code != "unknown-name" || d.StartLine != 5 || d.StartCol != 8 || d.EndLine != 5 || d.EndCol != 18 {
		t.Errorf("diagnostic %v, want the error unknown-name at 5:8-5:18", d)
	}
}

func TestParseRequestRefusesUse(t *testing.T) {
	text := "siml 1\nspec hunter\n\nuse \"default.apl.siml\"\n"
	r := ParseRequest(&proto.SimlParseRequest{Path: "wowsims.siml", Text: text}, readNames(t, "names.json"))
	if len(r.Diagnostics) != 1 || r.Diagnostics[0].Code != "use-refused" || r.Diagnostics[0].StartLine != 4 {
		t.Errorf("diagnostics %v, want use-refused on line 4", r.Diagnostics)
	}
}

func TestPrintRequest(t *testing.T) {
	names := readNames(t, "names.json")
	parsed := ParseRequest(&proto.SimlParseRequest{Path: "rotation.apl.siml", Text: apiRotation}, names)
	r := PrintRequest(&proto.SimlPrintRequest{Section: "apl", Spec: proto.Spec_SpecHunter, Settings: parsed.Settings}, names)
	if r.Error != "" || len(r.Diagnostics) > 0 || r.Text != apiRotation {
		t.Errorf("got text %q, error %q, diagnostics %v; want %q", r.Text, r.Error, r.Diagnostics, apiRotation)
	}

	setup := PrintRequest(&proto.SimlPrintRequest{Spec: proto.Spec_SpecHunter, Settings: parsed.Settings, Categories: []string{"r", "c"}}, names)
	if want := "siml 1\nspec hunter\n\nconsumables:\n\npriority:\n  cast multi_shot\n"; setup.Text != want {
		t.Errorf("full setup %q, error %q; want %q", setup.Text, setup.Error, want)
	}
}

// With no spec line, a rotation does not read, so the print gives the problems instead of the text.
func TestPrintRequestRefusesWhatDoesNotReadBack(t *testing.T) {
	names := readNames(t, "names.json")
	parsed := ParseRequest(&proto.SimlParseRequest{Path: "rotation.apl.siml", Text: apiRotation}, names)
	r := PrintRequest(&proto.SimlPrintRequest{Section: "apl", Settings: &proto.IndividualSimSettings{Player: &proto.Player{Rotation: parsed.Settings.Player.Rotation}}}, names)
	if r.Text != "" || len(r.Diagnostics) == 0 {
		t.Errorf("got text %q, diagnostics %v; want no text and the problems", r.Text, r.Diagnostics)
	}
}

func TestPrintRequestUnknownSection(t *testing.T) {
	r := PrintRequest(&proto.SimlPrintRequest{Section: "nowhere"}, readNames(t, "names.json"))
	if r.Error == "" || r.Text != "" {
		t.Errorf("got text %q, error %q; want an error", r.Text, r.Error)
	}
}

// A literal run as long as a file may be reads without a stack of its own length (SPEC.md §2 reads a compound
// duration as any number of number-and-unit pairs).
func TestParseRequestReadsALongLiteralRun(t *testing.T) {
	text := "siml 1\nspec hunter\n\nvar $x = " + strings.Repeat("1s", maxSourceSize/4) + "\n"
	r := ParseRequest(&proto.SimlParseRequest{Path: "x.apl.siml", Text: text}, readNames(t, "names.json"))
	if r.Error != "" || len(r.Diagnostics) > 0 {
		t.Fatalf("got error %q, diagnostics %v, want the run read as one duration", r.Error, r.Diagnostics)
	}
}

func TestParseRequestRefusesTextLargerThanAFile(t *testing.T) {
	text := "siml 1\nspec hunter\n\n// " + strings.Repeat("x", maxSourceSize) + "\n"
	names := readNames(t, "names.json")
	if r := ParseRequest(&proto.SimlParseRequest{Path: "x.apl.siml", Text: text}, names); !strings.Contains(r.Error, "larger than") {
		t.Errorf("ParseRequest gave the error %q, want one naming the size", r.Error)
	}
}

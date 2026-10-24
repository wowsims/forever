package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	_ "github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

var basicSpec = &proto.Player_ElementalShaman{
	ElementalShaman: &proto.ElementalShaman{
		Options: &proto.ElementalShaman_Options{
			ClassOptions: &proto.ShamanOptions{},
		},
	},
}

var p1Equip = &proto.EquipmentSpec{
	Items: []*proto.ItemSpec{
		{Id: 40516, Enchant: 3820},
		{Id: 44661},
		{Id: 40286, Enchant: 3810},
		{Id: 44005, Enchant: 3722},
		{Id: 40514, Enchant: 3832},
		{Id: 40324, Enchant: 2332},
		{Id: 40302, Enchant: 3246},
		{Id: 40301},
		{Id: 40560, Enchant: 3721},
		{Id: 40519, Enchant: 3826},
		{Id: 37694},
		{Id: 40399},
		{Id: 40432},
		{Id: 40255},
		{Id: 40395, Enchant: 3834},
		{Id: 40401, Enchant: 1128},
		{Id: 40267},
	},
}

func init() {
	s := &server{
		progMut:         sync.RWMutex{},
		asyncProgresses: map[string]*asyncProgress{},
	}
	go func() {
		s.runServer(true, "localhost:3339", false, "", false, bufio.NewReader(bytes.NewBuffer([]byte{})))
	}()

	time.Sleep(time.Second) // hack so we have time for server to startup. Probably could repeatedly curl the endpoint until it responds.
}

// TestIndividualSim is just a smoke test to make sure the http server works as expected.
//
//	Don't modify this test unless the proto defintions change and this no longer compiles.
func TestIndividualSim(t *testing.T) {
	req := &proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Race:      proto.Race_RaceTroll,
				Class:     proto.Class_ClassShaman,
				Equipment: p1Equip,
				Spec:      basicSpec,
			},
			&proto.PartyBuffs{},
			&proto.RaidBuffs{},
			&proto.Debuffs{}),
		Encounter: &proto.Encounter{
			Duration: 120,
			Targets: []*proto.Target{
				{},
			},
		},
		SimOptions: &proto.SimOptions{
			Iterations: 5000,
			RandomSeed: 1,
			Debug:      false,
		},
	}

	msgBytes, err := googleProto.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to encode request: %s", err.Error())
	}

	r, err := http.Post("http://localhost:3339/raidSim", "application/x-protobuf", bytes.NewReader(msgBytes))
	if err != nil {
		t.Fatalf("Failed to POST request: %s", err.Error())
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("Failed to read result body: %s", err.Error())
		return
	}

	rsr := &proto.RaidSimResult{}
	if err := googleProto.Unmarshal(body, rsr); err != nil {
		t.Fatalf("Failed to parse request: %s", err.Error())
		return
	}

	log.Printf("RESULT: %#v", rsr)
}

// A table missing on the first SIML request is read on a later one.
func TestSimlNamesMissingAtFirstIsReadLater(t *testing.T) {
	site, saved := fstest.MapFS{}, siteFS
	siteFS = site
	defer func() { siteFS, simlNamesTable = saved, nil }()
	if _, err := simlNames(); err == nil {
		t.Fatal("a site without the table gave one")
	}
	data, err := os.ReadFile("../siml/names.json")
	if err != nil {
		t.Fatal(err)
	}
	site["forever/siml_names.json"] = &fstest.MapFile{Data: data}
	if _, err := simlNames(); err != nil {
		t.Errorf("the table, once the site holds it: %v", err)
	}
}

// The SIML routes, read with the committed name table served as the site's.
func TestSimlRoutes(t *testing.T) {
	data, err := os.ReadFile("../siml/names.json")
	if err != nil {
		t.Fatal(err)
	}
	saved := siteFS
	siteFS = fstest.MapFS{"forever/siml_names.json": &fstest.MapFile{Data: data}}
	defer func() { siteFS, simlNamesTable = saved, nil }()
	post := func(route string, req, result googleProto.Message) {
		t.Helper()
		msgBytes, err := googleProto.Marshal(req)
		if err != nil {
			t.Fatal(err)
		}
		r, err := http.Post("http://localhost:3339"+route, "application/x-protobuf", bytes.NewReader(msgBytes))
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if err := googleProto.Unmarshal(body, result); err != nil {
			t.Fatalf("%s answered %d %q: %v", route, r.StatusCode, body, err)
		}
	}

	const text = "siml 1\nspec hunter\n\npriority:\n  cast multi_shot\n"
	parsed := &proto.SimlParseResult{}
	post("/simlParse", &proto.SimlParseRequest{Path: "rotation.apl.siml", Text: text}, parsed)
	if parsed.Error != "" || parsed.Spec != proto.Spec_SpecHunter || len(parsed.Categories) != 1 || parsed.Categories[0] != "r" {
		t.Fatalf("/simlParse gave %v", parsed)
	}

	printed := &proto.SimlPrintResult{}
	post("/simlPrint", &proto.SimlPrintRequest{Section: "apl", Spec: proto.Spec_SpecHunter, Settings: parsed.Settings}, printed)
	if printed.Text != text {
		t.Errorf("/simlPrint gave %v, want %q", printed, text)
	}

}

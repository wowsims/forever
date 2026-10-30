//go:build with_db

package core

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/siml"
)

// Writes each name → text into a new temporary folder, and returns the folder.
func presetDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, text := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// What read panics with, and whether it panicked.
func panicOf(read func()) (message string, panicked bool) {
	defer func() {
		if r := recover(); r != nil {
			message, panicked = fmt.Sprint(r), true
		}
	}()
	read()
	return "", false
}

func TestPresetsReadAsSIML(t *testing.T) {
	names, err := siml.DefaultNames()
	if err != nil {
		t.Fatal(err)
	}
	huntersMark, _ := names.Spell(proto.Class_ClassHunter, siml.NamespaceSpell, "hunters_mark")
	bow, _ := names.Item("fahrads_reloading_repeater")
	// A JSON file beside each .siml, which the .siml goes ahead of.
	dir := presetDir(t, map[string]string{
		"x.apl.siml":  "siml 1\nspec hunter\n\npriority:\n  cast hunters_mark\n",
		"x.gear.siml": "siml 1\n\nranged  fahrads_reloading_repeater\n",
		"x.apl.json":  "{}",
		"x.gear.json": "{}",
	})

	rotation := GetAplRotation(dir, "x")
	if rotation.Label != "x" || rotation.Rotation.GetPriorityList()[0].GetAction().GetCastSpell().GetSpellId().GetSpellId() != huntersMark.GetSpellId() {
		t.Errorf("GetAplRotation gave %v, want the label x and a cast of Hunter's Mark", rotation)
	}
	if items := GetGearSet(dir, "x").GearSet.GetItems(); len(items) != 17 || items[proto.ItemSlot_ItemSlotRanged].GetId() != bow {
		t.Errorf("GetGearSet gave the items %v, want 17 with Fahrad's Reloading Repeater in the ranged slot", items)
	}
	if items := GetItemSwapGearSet(dir, "x").ItemSwap.GetItems(); len(items) != 17 || items[proto.ItemSlot_ItemSlotRanged].GetId() != bow {
		t.Errorf("GetItemSwapGearSet gave the items %v, want 17 with Fahrad's Reloading Repeater in the ranged slot", items)
	}
}

// A spec that has no .siml preset of a name reads its JSON one.
func TestPresetsReadAsJSONWithNoSIML(t *testing.T) {
	dir := presetDir(t, map[string]string{
		"x.apl.json":  `{"priorityList":[{"action":{"castSpell":{"spellId":{"spellId":14325}}}}]}`,
		"x.gear.json": `{"items":[{"id":18323}]}`,
	})
	if id := GetAplRotation(dir, "x").Rotation.GetPriorityList()[0].GetAction().GetCastSpell().GetSpellId().GetSpellId(); id != 14325 {
		t.Errorf("GetAplRotation cast %d, want the JSON file's spell", id)
	}
	if items := GetGearSet(dir, "x").GearSet.GetItems(); len(items) != 1 || items[0].GetId() != 18323 {
		t.Errorf("GetGearSet gave %v, want the JSON file's item", items)
	}
	if items := GetItemSwapGearSet(dir, "x").ItemSwap.GetItems(); len(items) != 1 || items[0].GetId() != 18323 {
		t.Errorf("GetItemSwapGearSet gave %v, want the JSON file's item", items)
	}
}

func TestAPresetThatDoesNotReadPanicsWithItsDiagnostics(t *testing.T) {
	dir := presetDir(t, map[string]string{"x.apl.siml": "siml 1\nspec hunter\n\npriority:\n  cast hunters_mrak\n"})
	message, panicked := panicOf(func() { GetAplRotation(dir, "x") })
	if !panicked || !strings.Contains(message, "x.apl.siml:5:8 unknown-name no spell is named hunters_mrak; did you mean hunters_mark?") {
		t.Errorf("GetAplRotation gave the panic %q (panicked: %v), want the unknown-name diagnostic", message, panicked)
	}
}

// A preset missing in both forms is logged and left empty, so a suite runs on what a spec ships.
func TestAMissingPresetIsSkipped(t *testing.T) {
	var logged bytes.Buffer
	log.SetOutput(&logged)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	dir := t.TempDir()

	if rotation := GetAplRotation(dir, "x"); rotation.Label != "x" || rotation.Rotation != nil {
		t.Errorf("GetAplRotation gave %v, want the label x and no rotation", rotation)
	}
	if gear := GetGearSet(dir, "x"); gear.Label != "x" || len(gear.GearSet.GetItems()) != 0 {
		t.Errorf("GetGearSet gave %v, want the label x and no items", gear)
	}
	for _, want := range []string{
		"no apl json file, skipping rotation: " + dir + "/x.apl.json",
		"no gear json file, skipping gear set: " + dir + "/x.gear.json",
	} {
		if !strings.Contains(logged.String(), want) {
			t.Errorf("the log %q does not say %q", logged.String(), want)
		}
	}
}

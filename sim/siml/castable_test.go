package siml

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

// Where an action casts, completion and did-you-mean offer only the spells a rotation may cast, and the parser refuses
// the others with the castable spell of the name as the quick fix. A value position takes either.
func TestCastPositionsTakeCastableSpells(t *testing.T) {
	names := &Names{Classes: map[proto.Class]SpellNames{proto.Class_ClassHunter: {NamespaceSpell: {
		"strike":      {SpellID: 900001, Tag: 1, RegisteredBy: []string{"hunter"}},
		"strike_tag0": {SpellID: 900001, RegisteredBy: []string{"hunter"}, Uncastable: true},
	}}}}
	for _, text := range []string{"priority:\n  cast |\n", "priority:\n  channel |\n", "priority:\n  multidot(|\n"} {
		if got := labelsOf(completionsIn(t, names, nil, hunterPath, plainPreface+text)); !slices.Contains(got, "strike") || slices.Contains(got, "strike_tag0") {
			t.Errorf("%q offers %v, want strike and not strike_tag0", text, got)
		}
	}
	if got := labelsOf(completionsIn(t, names, nil, hunterPath, plainPreface+"priority:\n  cast strike if spell(|\n")); !slices.Contains(got, "strike_tag0") {
		t.Errorf("spell( offers %v, want strike_tag0 too", got)
	}

	text := plainPreface + "priority:\n  cast strike_tag0 if spell(strike_tag0).is_ready\n  cast strike_tag1 if spell(strike_tag1).is_ready\n"
	d := Analyze(hunterPath, []byte(text), names, langOpen)
	var got []string
	for _, x := range d.Diagnostics() {
		got = append(got, x.Message)
	}
	want := []string{
		"strike_tag0 is a spell no build lets a rotation cast; did you mean strike?",
		"no spell is named strike_tag1",
		"no spell is named strike_tag1; did you mean strike_tag0?",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
	first := d.Diagnostics()[0]
	if fixes := d.QuickFixes(Range{first.Start, first.End}); len(fixes) == 0 || fixes[0].Title != "Change to strike" {
		t.Errorf("quick fixes %v, want Change to strike", fixes)
	}
}

// A talent kind takes a talent's own spell, and a spell the spec knows only as a talent reads nowhere else: completion
// and did-you-mean offer each name only where it fits.
func TestTalentPositionsTakeTalents(t *testing.T) {
	names := &Names{Classes: map[proto.Class]SpellNames{proto.Class_ClassHunter: {NamespaceSpell: {
		"aimed_strike":   {SpellID: 900001, RegisteredBy: []string{"hunter"}},
		"deadly_aspects": {SpellID: 19552, RegisteredBy: []string{"hunter"}, TalentOnlyBy: []string{"hunter"}, Uncastable: true},
	}}}}
	offers := func(text string) []string {
		return labelsOf(completionsIn(t, names, nil, hunterPath, plainPreface+text))
	}
	if got := offers("priority:\n  cast aimed_strike if talent_is_known(|\n"); !slices.Contains(got, "deadly_aspects") || slices.Contains(got, "aimed_strike") {
		t.Errorf("talent_is_known( offers %v, want deadly_aspects and not aimed_strike", got)
	}
	if got := offers("priority:\n  cast aimed_strike if spell(|\n"); !slices.Contains(got, "aimed_strike") || slices.Contains(got, "deadly_aspects") {
		t.Errorf("spell( offers %v, want aimed_strike and not deadly_aspects", got)
	}

	text := plainPreface + "priority:\n  cast aimed_strike if talent_is_known(deadly_aspect)\n  cast aimed_strike if spell(deadly_aspect).is_ready\n"
	var got []string
	for _, x := range Analyze(hunterPath, []byte(text), names, langOpen).Diagnostics() {
		got = append(got, x.Message)
	}
	want := []string{"no spell is named deadly_aspect; did you mean deadly_aspects?", "no spell is named deadly_aspect"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

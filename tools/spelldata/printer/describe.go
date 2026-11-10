package printer

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The spell's own columns, leaving out what the row does not state.
func headerFields(s *spelldata.Spell) []field {
	out := []field{}
	add := func(key, value string) {
		if value != "" {
			out = append(out, field{key, value})
		}
	}

	add("school", schoolName(s.SpellSchool()))
	add("defense", defenseName(s.DefenseTypeCore()))
	if s.IsBleed() {
		add("mechanic", "bleed")
	} else if s.Mechanic != 0 {
		add("mechanic", fmt.Sprintf("mechanic %d", s.Mechanic))
	}
	if s.DurationMs == -1 {
		add("duration", "permanent")
	} else {
		add("duration", seconds(s.DurationMs))
	}
	add("cast", seconds(s.CastTimeMs))
	add("cooldown", seconds(s.CooldownMs))
	if s.CategoryCooldownMs != 0 {
		add("cooldown", seconds(s.CategoryCooldownMs)+" (category)")
	}
	add("gcd", seconds(s.GCDMs))
	for i := range s.Powers {
		add("cost", cost(s, &s.Powers[i]))
	}
	add("range", rangePhrase(s))
	add("stance", stanceList(s.StanceMask))
	add("stance exclude", stanceList(s.StanceExclude))
	if s.CasterAura != 0 {
		add("caster aura", spellRef(s.CasterAura))
	}
	if s.ExcludeCasterAura != 0 {
		add("excluded caster aura", spellRef(s.ExcludeCasterAura))
	}
	add("equip", equipRequirement(s))
	if s.MaxTargets != 0 {
		add("targets", fmt.Sprintf("up to %d", s.MaxTargets))
	}
	if s.MaxStack != 0 {
		add("stack", fmt.Sprintf("up to %d", s.MaxStack))
	}
	if s.ProcCharges != 0 {
		add("charges", strconv.Itoa(int(s.ProcCharges)))
	}
	add("icd", seconds(s.ICDMs))
	add("attrs", attributeList(s))
	add("labels", labelList(s))
	return out
}

// SpellPower states either a flat amount out of the bar or a share of the caster's base.
func cost(s *spelldata.Spell, p *spelldata.Power) string {
	if p.CostPct != 0 {
		return fmt.Sprintf("%s%% of base %s", number(float64(p.CostPct)), powerName(p.Type))
	}
	if p.Cost == 0 {
		return ""
	}
	return join(number(s.PowerCost(p.Type)), powerName(p.Type))
}

// SpellRange, with the dead zone a charge states as a minimum.
func rangePhrase(s *spelldata.Spell) string {
	if s.MaxRange == 0 {
		return ""
	}
	if s.MinRange != 0 {
		return fmt.Sprintf("%s-%s yd", number(float64(s.MinRange)), number(float64(s.MaxRange)))
	}
	return number(float64(s.MaxRange)) + " yd"
}

// The attributes sim/core/spelldata/attributes.go names, which are the ones the sim acts on. IsAProc
// is stated only when the game files deny it, since every other row is one.
func attributeList(s *spelldata.Spell) string {
	var out []string
	for _, attr := range []struct {
		set  bool
		name string
	}{
		{s.IsPassive(), "passive"},
		{s.IsChanneled(), "channeled"},
		{s.RefundsOnMiss(), "refund on miss"},
		{s.PeriodicCanCrit(), "periodic can crit"},
		{s.CannotCrit(), "cannot crit"},
		{!s.IsAProc(), "not a proc"},
		{s.CanProcFromProcs(), "can proc from procs"},
		{s.SuppressesWeaponProcs(), "suppresses weapon procs"},
		{s.IsWeaponProcAura(), "hears hits as a weapon proc"},
		{s.ClassSpellsOnly(), "class abilities only"},
	} {
		if attr.set {
			out = append(out, attr.name)
		}
	}
	return strings.Join(out, ", ")
}

// What the row says about a proc: where the rate is stated, the rate itself, the hits the mask lets
// through and what the tooltip added that the mask cannot say.
//
// The heartbeat bit alone is the game files' default on any aura, so a row carrying nothing else is
// not a proc and states no summary.
func procSummary(s *spelldata.Spell) string {
	if s.ProcFlags[0]&^dbcenums.PROC_FLAG_HEARTBEAT == 0 && s.ProcFlags[1] == 0 && s.RPPM == 0 {
		return ""
	}

	var parts []string
	if flags := procFlagNames(s.ProcFlags); len(flags) > 0 {
		parts = append(parts, "hears "+strings.Join(flags, ", "))
	}
	parts = append(parts, procRate(s))
	if hints := procHintNames(s.ProcHint); len(hints) > 0 {
		parts = append(parts, "tooltip says "+strings.Join(hints, ", "))
	}
	return strings.Join(parts, "; ")
}

// The rate by the source that says where it is stated: the ProcChance column is the roll only under
// ProcChanceColumn, which is what ProcChanceSource exists to say.
func procRate(s *spelldata.Spell) string {
	if s.RPPM > 0 {
		return number(float64(s.RPPM)) + " procs per minute"
	}
	switch s.ProcChanceSource {
	case spelldata.ProcChanceColumn:
		return fmt.Sprintf("%d%% chance", s.ProcChance)
	case spelldata.ProcChanceEffectN:
		return fmt.Sprintf("%s%% chance, from effect %d",
			number(s.EffectN(int(s.ProcChanceEffect)).BaseValue()), s.ProcChanceEffect)
	case spelldata.ProcChanceAlways:
		return "no roll, it fires whenever its condition is met"
	}
	return "no stated chance, the rate has to come from an override"
}

func refList(s *spelldata.Spell) []string {
	out := []string{}
	for _, ref := range s.Refs() {
		out = append(out, fmt.Sprintf("%d %s", ref.ID, ref.Name))
	}
	return out
}

// A SpellAuraRestrictions caster aura by id, named where the store carries it.
func spellRef(id int32) string {
	if s := spelldata.Find(id); s != spelldata.Nil {
		return fmt.Sprintf("%d %s", s.ID, s.Name)
	}
	return fmt.Sprint(id)
}

func labelList(s *spelldata.Spell) string {
	var out []string
	for _, label := range s.Labels {
		out = append(out, strconv.Itoa(int(label)))
	}
	return strings.Join(out, ", ")
}

func seconds(ms int32) string {
	if ms == 0 {
		return ""
	}
	return number(float64(ms)/1000) + " s"
}

// Read back as the float32 the game files' columns are, so a base point of 5.449999809265137 - the
// widening of the game files' 5.45 - prints as the game files state it.
func number(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 32)
}

func signed(v float64) string {
	if v >= 0 {
		return "+" + number(v)
	}
	return number(v)
}

func join(parts ...string) string {
	return strings.Join(nonEmpty(parts), " ")
}

func nonEmpty(parts []string) []string {
	return slices.DeleteFunc(slices.Clone(parts), func(part string) bool { return part == "" })
}

package buffs

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The generated rows whose first amount is not the number the game files state, at every talent point
// they price: mana over a tick period read as MP5, a percentage read as a multiplier or a pseudo-stat,
// a talent's share of the amount, and a finisher's per-combo-point amount at full combo points.
func TestGeneratedRowsReadInTheSimsUnits(t *testing.T) {
	mp5 := func(s *spelldata.Spell) func(*testing.T, int32) float64 {
		return func(t *testing.T, _ int32) float64 {
			e := firstAura(t, s, dbcenums.A_PERIODIC_ENERGIZE)
			return e.Average() * 5000 / float64(e.PeriodMs)
		}
	}
	multiplier := func(s *spelldata.Spell, aura dbcenums.EffectAuraType) func(*testing.T, int32) float64 {
		return func(t *testing.T, _ int32) float64 {
			return 1 + firstAura(t, s, aura).Average()/100
		}
	}

	cases := []struct {
		name   string
		value  func(int32) float64
		points int32
		want   func(t *testing.T, points int32) float64
	}{
		{"ManaTideTotems", ManaTideTotemsValue, 0, mp5(manaTideTotemsSpell)},
		{"GreaterBlessingOfWisdom", GreaterBlessingOfWisdomValue, 0, mp5(greaterBlessingOfWisdomSpell)},
		{"ManaSpringTotem", ManaSpringTotemValue, manaSpringTotemMeta.Talent.Len(), func(t *testing.T, points int32) float64 {
			e := firstAura(t, manaSpringTotemSpell, dbcenums.A_PERIODIC_ENERGIZE)
			return talentScaled(e.Average(), manaSpringTotemMeta.talentMod(points)) * 5000 / float64(e.PeriodMs)
		}},
		{"GreaterBlessingOfKings", GreaterBlessingOfKingsValue, 0, multiplier(greaterBlessingOfKingsSpell, dbcenums.A_MOD_TOTAL_STAT_PERCENTAGE)},
		{"PowerInfusions", PowerInfusionsValue, 0, multiplier(powerInfusionsSpell, dbcenums.A_MOD_HEALING_DONE_PERCENT)},
		{"ConcentrationAura", ConcentrationAuraValue, 0, func(t *testing.T, _ int32) float64 {
			return -firstAura(t, concentrationAuraSpell, dbcenums.A_REDUCE_PUSHBACK).Average() / 100
		}},
		// A slow divides the speed by the factor a haste of the same percentage multiplies it by.
		{"ThunderClap", ThunderClapValue, 0, func(t *testing.T, _ int32) float64 {
			pct := firstAura(t, thunderClapSpell, dbcenums.A_MOD_MELEE_HASTE_3).Average()
			if pct < 0 {
				return 1 / (1 - pct/100)
			}
			return 1 + pct/100
		}},
		{"ExposeArmor", ExposeArmorValue, 0, func(t *testing.T, _ int32) float64 {
			return float64(firstAura(t, exposeArmorSpell, dbcenums.A_MOD_ARMOR_NO_MODS).PointsPerResource) * 5
		}},
		// Each rank is its share of the debuff's amount, rounded half away from zero; untaken is the whole.
		{"Vindication", VindicationValue, vindicationMeta.Talent.Len(), func(t *testing.T, points int32) float64 {
			amount := firstAura(t, vindicationSpell, dbcenums.A_MOD_ATTACK_POWER).Average()
			if points == 0 {
				return amount
			}
			return math.Round(amount * float64(points) / float64(vindicationMeta.TalentDivisor))
		}},
		{"DarkInspiration", DarkInspirationValue, darkInspirationMeta.Talent.Len(), func(t *testing.T, points int32) float64 {
			pct := firstAura(t, darkInspirationSpell, dbcenums.A_MOD_TOTAL_STAT_PERCENTAGE).Average()
			if points > 0 {
				pct = math.Round(pct * float64(points) / float64(darkInspirationMeta.TalentDivisor))
			}
			return 1 + pct/100
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for points := int32(0); points <= c.points; points++ {
				if got, want := c.value(points), c.want(t, points); math.Abs(got-want) > 1e-9 {
					t.Errorf("Value(%d) = %v, want %v", points, got, want)
				}
			}
		})
	}
}

// An improving talent's modifier on an amount, truncated the way the game resolves one: a percentage
// of it, or a flat addition. An untaken talent adds nothing.
func talentScaled(amount float64, mod *spelldata.Effect) float64 {
	switch {
	case mod == spelldata.NilEffect:
		return amount
	case mod.Aura == dbcenums.A_ADD_PCT_MODIFIER:
		return math.Trunc(amount * (1 + mod.BaseValue()/100))
	default:
		return math.Trunc(amount + mod.BaseValue())
	}
}

// The row's first effect that applies this aura, whatever the effect's type: a party aura states the
// same aura as one on the caster alone.
func firstAura(t *testing.T, s *spelldata.Spell, aura dbcenums.EffectAuraType) *spelldata.Effect {
	t.Helper()
	for i := range s.Effects {
		if s.Effects[i].Aura == aura {
			return &s.Effects[i]
		}
	}
	t.Fatalf("spell %d applies no aura %v", s.ID, aura)
	return nil
}

// The game files state a permanent aura as -1 and one with no duration of its own, a totem's, as 0:
// both never expire. A row with no duration that names its cast lasts as long as the cast, and a
// driven buff waits the longer of the cast's own and its category's cooldown.
func TestMetaReadsTheDurationAndCooldown(t *testing.T) {
	durations := []struct {
		name string
		meta *Meta
		want time.Duration
	}{
		{"permanent", &Meta{Spell: &spelldata.Spell{DurationMs: -1}}, core.NeverExpires},
		{"no duration of its own", &Meta{Spell: &spelldata.Spell{}}, core.NeverExpires},
		{"timed", &Meta{Spell: &spelldata.Spell{DurationMs: 13000}}, 13 * time.Second},
		{"timed by its cast", &Meta{Spell: &spelldata.Spell{}, Cast: &spelldata.Spell{DurationMs: 13000}}, 13 * time.Second},
		{"permanent cast", &Meta{Spell: &spelldata.Spell{}, Cast: &spelldata.Spell{DurationMs: -1}}, core.NeverExpires},
		{"timed despite its cast", &Meta{Spell: &spelldata.Spell{DurationMs: 15000}, Cast: &spelldata.Spell{DurationMs: 13000}}, 15 * time.Second},
	}
	for _, c := range durations {
		if got := c.meta.Duration(0); got != c.want {
			t.Errorf("%s: Duration(0) = %v, want %v", c.name, got, c.want)
		}
	}

	cooldowns := []struct {
		name string
		meta *Meta
		want time.Duration
	}{
		{"own cooldown", &Meta{Spell: &spelldata.Spell{CooldownMs: 180000}}, 3 * time.Minute},
		{"category cooldown", &Meta{Spell: &spelldata.Spell{CategoryCooldownMs: 300000}}, 5 * time.Minute},
		{"the longer of both", &Meta{Spell: &spelldata.Spell{CooldownMs: 360000, CategoryCooldownMs: 300000}}, 6 * time.Minute},
		{"the cast's", &Meta{Spell: &spelldata.Spell{CooldownMs: 1000}, Cast: &spelldata.Spell{CategoryCooldownMs: 300000}}, 5 * time.Minute},
	}
	for _, c := range cooldowns {
		if got := c.meta.Cooldown(); got != c.want {
			t.Errorf("%s: Cooldown() = %v, want %v", c.name, got, c.want)
		}
	}
}

// A raid aura that is a damage shield and a healing-taken row of 0, as Retribution Aura is. Its value
// is the damage the shield deals, whatever else the row states: the parse would attach the 0 as a
// multiplier of 1.
func TestMetaValueOfADamageShieldIsItsDamage(t *testing.T) {
	aura := &spelldata.Spell{Effects: []spelldata.Effect{
		{Type: dbcenums.E_APPLY_AREA_AURA_PARTY, Aura: dbcenums.A_DAMAGE_SHIELD, BasePoints: 20},
		{Type: dbcenums.E_APPLY_AREA_AURA_PARTY, Aura: dbcenums.A_MOD_HEALING_PCT, Index: 1},
	}}
	if got := (&Meta{Spell: aura}).Value(0); got != 20 {
		t.Errorf("the aura without the skips reads %v, want the shield's 20", got)
	}
}

// A Vindication rank is worth rank/3 of the debuff's amount at the caster's level, rounded half away
// from zero, on both of its attack powers.
func TestVindicationRankReadsLikeItsTooltip(t *testing.T) {
	for _, level := range []int32{20, 16} {
		opts := append(vindicationMeta.Options(1), spelldata.Level(level))
		applied := spelldata.DryRun(vindicationMeta.Spell, opts...).Applied
		if len(applied) != 2 {
			t.Fatalf("level %d: the parse attaches %d amounts, want attack power and ranged attack power", level, len(applied))
		}
		for _, a := range applied {
			if want := math.Round(a.Effect.AverageAt(level) / float64(vindicationMeta.TalentDivisor)); a.Value != want {
				t.Errorf("level %d: %s is %v, want %v", level, a.Kind, a.Value, want)
			}
		}
	}
}

// Each state puts up the external copy of its own row, worth what that row states and bidding in the
// row's category, whichever spell the state reads. Off puts up nothing, and a state past the last
// reads as the last.
func TestRankSelectStatesGrantTheirOwnRows(t *testing.T) {
	stated := func(aura dbcenums.EffectAuraType) func(*testing.T, *spelldata.Spell) float64 {
		return func(t *testing.T, s *spelldata.Spell) float64 { return rowAmount(t, firstAura(t, s, aura)) }
	}
	mp5 := func(t *testing.T, s *spelldata.Spell) float64 {
		e := firstAura(t, s, dbcenums.A_PERIODIC_ENERGIZE)
		return e.Average() * 5000 / float64(e.PeriodMs)
	}

	cases := []struct {
		name   string
		ranks  rankStates
		stat   stats.Stat
		amount func(*testing.T, *spelldata.Spell) float64
		set    func(*proto.RaidBuffs, *proto.IndividualBuffs, int32)
	}{
		{"StrengthOfEarthTotem", strengthOfEarthTotemRanks, stats.Strength, stated(dbcenums.A_MOD_STAT),
			func(r *proto.RaidBuffs, _ *proto.IndividualBuffs, state int32) { r.StrengthOfEarthTotem = state }},
		{"GraceOfAirTotem", graceOfAirTotemRanks, stats.Agility, stated(dbcenums.A_MOD_STAT),
			func(r *proto.RaidBuffs, _ *proto.IndividualBuffs, state int32) { r.GraceOfAirTotem = state }},
		{"GreaterBlessingOfMight", greaterBlessingOfMightRanks, stats.AttackPower, stated(dbcenums.A_MOD_ATTACK_POWER),
			func(_ *proto.RaidBuffs, i *proto.IndividualBuffs, state int32) { i.GreaterBlessingOfMight = state }},
		{"GreaterBlessingOfWisdom", greaterBlessingOfWisdomRanks, stats.MP5, mp5,
			func(_ *proto.RaidBuffs, i *proto.IndividualBuffs, state int32) { i.GreaterBlessingOfWisdom = state }},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			applied := func(state int32) *core.Character {
				char := shapeCharacter()
				raid, individual := &proto.RaidBuffs{}, &proto.IndividualBuffs{}
				c.set(raid, individual, state)
				applyGeneratedBuffs(char, raid, &proto.PartyBuffs{}, individual)
				return char
			}
			granted := func(char *core.Character, s *spelldata.Spell) (float64, *core.Aura) {
				aura := char.GetAuraByID(core.ActionID{SpellID: s.ID, Tag: -1})
				if aura == nil {
					t.Fatalf("no external copy of spell %d is up", s.ID)
				}
				before := char.GetStats()[c.stat]
				aura.Activate(&core.Simulation{})
				return char.GetStats()[c.stat] - before, aura
			}

			off := applied(0)
			for _, state := range c.ranks {
				if aura := off.GetAuraByID(core.ActionID{SpellID: state.Spell.ID, Tag: -1}); aura != nil {
					t.Errorf("off puts up %s", aura.Label)
				}
			}

			for i, state := range c.ranks {
				got, aura := granted(applied(int32(i+1)), state.Spell)
				if want := c.amount(t, state.Spell); got != want {
					t.Errorf("state %d grants %v, want its row's %v", i+1, got, want)
				}
				if aura.Tag != state.Category {
					t.Errorf("state %d bids in %q, want the row's %q", i+1, aura.Tag, state.Category)
				}
			}

			last := c.ranks[len(c.ranks)-1].Spell
			if got, _ := granted(applied(int32(len(c.ranks)+1)), last); got != c.amount(t, last) {
				t.Errorf("a state past the last grants %v, want the last row's %v", got, c.amount(t, last))
			}
		})
	}
}

// A priest's prayer reads its own row and duration, and bids against the raid's copy of its stat, so
// the two count once.
func TestPrayerCountsOnceWithTheRaidsCopy(t *testing.T) {
	cases := []struct {
		label   string
		prayer  *spelldata.Spell
		newAura func(*core.Unit, bool, *spelldata.Spell) *core.Aura
		raid    *Meta
		stat    stats.Stat
		aura    dbcenums.EffectAuraType
	}{
		{"Prayer of Fortitude", spelldata.MustFind(21564), NewSpellDataPrayerOfFortitudeAura, powerWordFortitudeMeta, stats.Stamina, dbcenums.A_MOD_STAT},
		{"Prayer of Shadow Protection", spelldata.MustFind(27683), NewSpellDataPrayerOfShadowProtectionAura, shadowProtectionMeta, stats.ShadowResistance, dbcenums.A_MOD_RESISTANCE},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			char := shapeCharacter()
			own := c.newAura(&char.Unit, true, c.prayer)
			external := newBuff(&char.Unit, c.raid, false, 0)
			before := char.GetStats()[c.stat]

			if want := c.label + " (Player)"; own.Label != want {
				t.Errorf("the own copy is labelled %q, want %q", own.Label, want)
			}
			if want := auraDuration(c.prayer); own.Duration != want {
				t.Errorf("the own copy lasts %v, want the prayer's %v", own.Duration, want)
			}

			sim := &core.Simulation{}
			own.Activate(sim)
			external.Activate(sim)
			if got, want := char.GetStats()[c.stat]-before, rowAmount(t, firstAura(t, c.prayer, c.aura)); got != want {
				t.Errorf("both copies grant %v, want the row's %v counted once", got, want)
			}
		})
	}
}

// A totem state stands on a true pet where its row lets a pet carry one, and a state whose row says
// only a player can carry it stays off the pet. The row copy clears that attribute.
func TestRankSelectTotemStateOnATruePet(t *testing.T) {
	petCarries := func(state *Meta) bool {
		owner := core.NewCharacter(&core.Party{Raid: &core.Raid{}}, 0, &proto.Player{
			Name:      "Totem Owner",
			Race:      proto.Race_RaceOrc,
			Class:     proto.Class_ClassWarrior,
			Spec:      &proto.Player_ProtectionWarrior{ProtectionWarrior: &proto.ProtectionWarrior{}},
			Equipment: &proto.EquipmentSpec{Items: []*proto.ItemSpec{}},
		})
		pet := core.NewPet(core.PetConfig{
			Name:            "True Pet",
			Owner:           &owner,
			StatInheritance: func(stats.Stats) stats.Stats { return stats.Stats{} },
			EnabledOnStart:  true,
		})
		owner.Pets = append(owner.Pets, &pet)
		owner.Env = &core.Environment{MeasuringStats: true}

		standRankedTotem(&owner, state)
		if owner.GetAura(state.label(false)) == nil {
			t.Fatalf("the owner carries no %s", state.Label)
		}
		return pet.GetAura(state.label(false)) != nil
	}

	for _, state := range strengthOfEarthTotemRanks {
		row := *state.Spell
		row.Attr[dbcenums.ATTR_INDEX_EX_3] |= uint32(dbcenums.ATTR_EX_3_ONLY_ON_PLAYER)
		if petCarries(strengthOfEarthTotemMeta.atState(&row, state.Label)) {
			t.Errorf("%s stands on the pet with a row only a player can carry", state.Label)
		}
		row.Attr[dbcenums.ATTR_INDEX_EX_3] &^= uint32(dbcenums.ATTR_EX_3_ONLY_ON_PLAYER)
		if !petCarries(strengthOfEarthTotemMeta.atState(&row, state.Label)) {
			t.Errorf("%s does not stand on the pet", state.Label)
		}
	}
}

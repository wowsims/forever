package core_test

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/assets/database"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestPartyBattleShoutAppliesTheGeneratedAura(t *testing.T) {
	char := core.NewGeneratedBuffTestCharacter()

	core.ApplyBuffEffects(char,
		&proto.RaidBuffs{}, &proto.PartyBuffs{BattleShout: proto.TristateEffect_TristateEffectRegular}, &proto.IndividualBuffs{})
	char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)

	shout := battleShout(t)
	if got := char.GetStats()[stats.AttackPower]; got != shout {
		t.Errorf("the party's Battle Shout applied %v attack power, want the game files' %v", got, shout)
	}

	aura := char.GetAura("Battle Shout (External)")
	if aura == nil {
		t.Fatalf("no aura is labelled %q; the unit has %v", "Battle Shout (External)", auraLabels(char))
	}
	if want := (core.ActionID{SpellID: 25289, Tag: -1}); aura.ActionID != want {
		t.Errorf("the external copy is %v, want %v", aura.ActionID, want)
	}
	if want := rowDuration(t, 25289); aura.Duration != want {
		t.Errorf("the external copy lasts %v, want %v", aura.Duration, want)
	}

	category := char.ExclusiveEffectManager.GetExclusiveEffectCategory(buffs.BattleShoutCategory)
	if !category.SingleAura {
		t.Error("the Battle Shout category is not single-aura, so a second copy could sit next to it")
	}
	if len(category.Effects()) != 1 {
		t.Errorf("the category holds %d effects, want one", len(category.Effects()))
	} else if category.Effects()[0].Priority != shout {
		t.Errorf("the only effect in the category bids %v, want %v", category.Effects()[0].Priority, shout)
	}

	// The chain the driver installs finds the player's own shout by this tag.
	if tagged := char.GetAurasWithTag(buffs.BattleShoutCategory); len(tagged) != 1 || tagged[0] != aura {
		t.Errorf("%d auras carry the Battle Shout tag, want only the external copy", len(tagged))
	}
}

func battleShout(t *testing.T) float64 {
	t.Helper()
	return rowAmount(t, 25289, dbcenums.A_MOD_ATTACK_POWER)
}

// The attack power three pieces of Battlegear of Wrath add to the shout.
func battleShoutSetBonus(t *testing.T) float64 {
	t.Helper()
	return rowAmount(t, 23563, dbcenums.A_ADD_FLAT_MODIFIER)
}

// A warrior who casts Battle Shout and has the external one ticked: both copies
// are worth the same attack power, and both are equally long, so the tie goes to
// whichever the build phase activates last. ExclusiveEffect.Activate only turns a
// newcomer away when the incumbent's remaining duration is longer than the
// newcomer's, so the player's copy takes the category and the external one is
// deactivated - the character sheet shows one shout either way.
func TestPlayerBattleShoutTakesTheCategoryOnATie(t *testing.T) {
	char := core.NewGeneratedBuffTestCharacter()

	core.ApplyBuffEffects(char,
		&proto.RaidBuffs{}, &proto.PartyBuffs{BattleShout: proto.TristateEffect_TristateEffectRegular}, &proto.IndividualBuffs{})

	player := buffs.NewSpellDataBattleShoutAura(&char.Unit, true, 0)
	if want := (core.ActionID{SpellID: 25289, Tag: 0}); player.ActionID != want {
		t.Errorf("the player's copy is %v, want %v", player.ActionID, want)
	}
	if player.Label != "Battle Shout (Player)" {
		t.Errorf("the player's copy is labelled %q, want %q", player.Label, "Battle Shout (Player)")
	}
	// What a warrior whose UseBattleShout is set does with its own aura.
	player.BuildPhase = core.CharacterBuildPhaseBuffs

	char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)

	external := char.GetAura("Battle Shout (External)")
	if !player.IsActive() {
		t.Error("the player's own shout did not activate next to the external one")
	}
	if external.IsActive() {
		t.Error("the external copy stayed active, so both copies of the shout are on the character")
	}
	if got, want := char.GetStats()[stats.AttackPower], battleShout(t); got != want {
		t.Errorf("both copies together applied %v attack power, want %v", got, want)
	}
	if tagged := char.GetAurasWithTag(buffs.BattleShoutCategory); len(tagged) != 2 {
		t.Errorf("%d auras carry the Battle Shout tag, want both copies", len(tagged))
	}
}

// The improved state says the warrior who shouted for the party wears three
// pieces of Battlegear of Wrath, so the external copy is worth the game files'
// shout plus the set's bonus.
func TestImprovedPartyBattleShoutAddsTheTierTwoBonus(t *testing.T) {
	for _, row := range []struct {
		name  string
		party *proto.PartyBuffs
		want  float64
	}{
		{"with the set", &proto.PartyBuffs{BattleShout: proto.TristateEffect_TristateEffectImproved}, battleShout(t) + battleShoutSetBonus(t)},
		{"without it", &proto.PartyBuffs{BattleShout: proto.TristateEffect_TristateEffectRegular}, battleShout(t)},
	} {
		t.Run(row.name, func(t *testing.T) {
			char := core.NewGeneratedBuffTestCharacter()

			core.ApplyBuffEffects(char, &proto.RaidBuffs{}, row.party, &proto.IndividualBuffs{})
			char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)

			if got := char.GetStats()[stats.AttackPower]; got != row.want {
				t.Errorf("the party's Battle Shout applied %v attack power, want %v", got, row.want)
			}
			if got := char.GetAura("Battle Shout (External)").ExclusiveEffects[0].Priority; got != row.want {
				t.Errorf("the external copy bids %v, want %v", got, row.want)
			}
		})
	}
}

// Two copies of the shout that are not worth the same: the stronger one
// takes the category whichever side it is on, and the character sheet reads it
// once rather than both added up.
func TestTheStrongerBattleShoutTakesTheCategory(t *testing.T) {
	for _, row := range []struct {
		name         string
		party        *proto.PartyBuffs
		playerBonus  float64
		wantPlayerUp bool
	}{
		{"the player wears the set", &proto.PartyBuffs{BattleShout: proto.TristateEffect_TristateEffectRegular}, buffs.BattleShoutT2Bonus, true},
		{"the external warrior does", &proto.PartyBuffs{BattleShout: proto.TristateEffect_TristateEffectImproved}, 0, false},
	} {
		t.Run(row.name, func(t *testing.T) {
			char := core.NewGeneratedBuffTestCharacter()

			core.ApplyBuffEffects(char, &proto.RaidBuffs{}, row.party, &proto.IndividualBuffs{})

			player := buffs.NewSpellDataBattleShoutAura(&char.Unit, true, 0)
			if row.playerBonus != 0 {
				buffs.AddGeneratedFlatBonus(player, stats.AttackPower, buffs.BattleShoutValue(0), row.playerBonus)
			}
			player.BuildPhase = core.CharacterBuildPhaseBuffs

			char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)

			external := char.GetAura("Battle Shout (External)")
			if player.IsActive() != row.wantPlayerUp {
				t.Errorf("the player's copy is active: %v, want %v", player.IsActive(), row.wantPlayerUp)
			}
			if external.IsActive() == row.wantPlayerUp {
				t.Errorf("the external copy is active: %v, want %v", external.IsActive(), !row.wantPlayerUp)
			}
			if got, want := char.GetStats()[stats.AttackPower], battleShout(t)+battleShoutSetBonus(t); got != want {
				t.Errorf("both copies together applied %v attack power, want the stronger one's %v", got, want)
			}
		})
	}
}

// Two flat stat rows: Prayer of Fortitude's stamina bids only against its own stat's category and
// Blood Pact names none, so the two sources of stamina add up.
func TestGeneratedFlatStatBuffsAddUp(t *testing.T) {
	char := core.NewGeneratedBuffTestCharacter()

	core.ApplyBuffEffects(char,
		&proto.RaidBuffs{PrayerOfFortitude: true}, &proto.PartyBuffs{BloodPact: true}, &proto.IndividualBuffs{})
	char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)

	fortitudeStamina, pactStamina := rowAmount(t, 21564, dbcenums.A_MOD_STAT), rowAmount(t, 11767, dbcenums.A_MOD_STAT)
	if got := char.GetStats()[stats.Stamina]; got != fortitudeStamina+pactStamina {
		t.Errorf("Prayer of Fortitude and Blood Pact applied %v stamina, want the game files' %v + %v", got, fortitudeStamina, pactStamina)
	}

	fortitude := char.GetAura("Prayer of Fortitude (External)")
	if fortitude == nil {
		t.Fatalf("no aura is labelled %q; the unit has %v", "Prayer of Fortitude (External)", auraLabels(char))
	}
	if want := (core.ActionID{SpellID: 21564, Tag: -1}); fortitude.ActionID != want {
		t.Errorf("the external copy is %v, want %v", fortitude.ActionID, want)
	}
}

// A percentage row: every stat the game files' A_MOD_TOTAL_STAT_PERCENTAGE names
// goes through a multiplying dependency rather than a flat amount, floored to a whole number.
func TestGeneratedGreaterBlessingOfKingsMultipliesEveryStat(t *testing.T) {
	multiplier := core.MultiplierFromPercent(rowAmount(t, 25898, dbcenums.A_MOD_TOTAL_STAT_PERCENTAGE))
	if multiplier <= 1 {
		t.Fatalf("the blessing is worth %v, want a multiplier above 1", multiplier)
	}
	char := core.NewGeneratedBuffTestCharacter()
	char.SetStats(stats.Stats{
		stats.Strength: 100, stats.Agility: 100, stats.Stamina: 100,
		stats.Intellect: 100, stats.Spirit: 100, stats.Armor: 100,
	})

	core.ApplyBuffEffects(char,
		&proto.RaidBuffs{}, &proto.PartyBuffs{}, &proto.IndividualBuffs{GreaterBlessingOfKings: true})
	measureGeneratedBuffStats(char)

	want := math.Floor(100 * multiplier)
	for _, stat := range []stats.Stat{stats.Strength, stats.Agility, stats.Stamina, stats.Intellect, stats.Spirit} {
		if got := char.GetStats()[stat]; got != want {
			t.Errorf("%s is %v, want 100 multiplied by the game files' %v, %v", stat.StatName(), got, multiplier, want)
		}
	}
	if got := char.GetStats()[stats.Armor]; got != 100 {
		t.Errorf("armor is %v, want the blessing to have left it alone", got)
	}
}

// Two sources of shadow resistance compete for the school, so the stronger one
// holds it alone, while everything else Gift of the Wild grants is applied outright.
func TestGeneratedResistancesCompeteAcrossBuffs(t *testing.T) {
	measure := func(raid *proto.RaidBuffs) (*core.Character, stats.Stats) {
		char := core.NewGeneratedBuffTestCharacter()
		core.ApplyBuffEffects(char, raid, &proto.PartyBuffs{}, &proto.IndividualBuffs{})
		char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)
		return char, char.GetStats()
	}
	_, giftOnly := measure(&proto.RaidBuffs{GiftOfTheWild: true})
	_, prayerOnly := measure(&proto.RaidBuffs{PrayerOfShadowProtection: true})
	char, both := measure(&proto.RaidBuffs{GiftOfTheWild: true, PrayerOfShadowProtection: true})

	for _, stat := range []stats.Stat{stats.ShadowResistance, stats.FrostResistance, stats.Armor, stats.Stamina} {
		if giftOnly[stat] <= 0 {
			t.Fatalf("Gift of the Wild alone applied %v %s, want some", giftOnly[stat], stat.StatName())
		}
	}
	if prayerOnly[stats.ShadowResistance] <= 0 {
		t.Fatalf("Prayer of Shadow Protection alone applied %v shadow resistance, want some", prayerOnly[stats.ShadowResistance])
	}
	if got, sum := both[stats.ShadowResistance], giftOnly[stats.ShadowResistance]+prayerOnly[stats.ShadowResistance]; got >= sum {
		t.Errorf("shadow resistance is %v, want less than both added up, %v", got, sum)
	}
	if got, want := both[stats.ShadowResistance], max(giftOnly[stats.ShadowResistance], prayerOnly[stats.ShadowResistance]); got != want {
		t.Errorf("shadow resistance is %v, want only the stronger of the two, %v", got, want)
	}
	for _, stat := range []stats.Stat{stats.FrostResistance, stats.Armor, stats.Stamina} {
		if got, want := both[stat], giftOnly[stat]; got != want {
			t.Errorf("%s is %v, want Gift of the Wild's %v", stat.StatName(), got, want)
		}
	}
	for _, label := range []string{"Gift of the Wild (External)", "Prayer of Shadow Protection (External)"} {
		if aura := char.GetAura(label); aura == nil || !aura.IsActive() {
			t.Errorf("%s is not active, want both buffs up with only the shadow resistance held", label)
		}
	}

	school := char.ExclusiveEffectManager.GetExclusiveEffectCategory(
		core.ResistanceCategoryShadow + stats.ShadowResistance.StatName() + "Add")
	if len(school.Effects()) != 2 {
		t.Errorf("the shadow school holds %d effects, want both buffs", len(school.Effects()))
	}
}

// A paladin aura holds its own slot and joins the shared one that stops a
// paladin from having two of them up. Only the player's own copy joins it: the
// external copy has to be able to sit next to the one the paladin casts.
func TestGeneratedPaladinAuraJoinsTheSharedCategoryOnThePlayerCopyOnly(t *testing.T) {
	char := core.NewGeneratedBuffTestCharacter()

	core.ApplyBuffEffects(char,
		&proto.RaidBuffs{}, &proto.PartyBuffs{DevotionAura: true}, &proto.IndividualBuffs{})

	shared := char.ExclusiveEffectManager.GetExclusiveEffectCategory("PaladinAura")
	if len(shared.Effects()) != 0 {
		t.Errorf("the external copy joined the shared category with %d effects, want none", len(shared.Effects()))
	}

	player := buffs.NewSpellDataDevotionAura(&char.Unit, true, 0)
	if len(shared.Effects()) != 1 {
		t.Errorf("the player's copy put %d effects in the shared category, want one", len(shared.Effects()))
	}
	player.BuildPhase = core.CharacterBuildPhaseBuffs

	measureGeneratedBuffStats(char)

	own := char.ExclusiveEffectManager.GetExclusiveEffectCategory(buffs.DevotionAuraCategory)
	if !own.SingleAura {
		t.Error("the aura's own category is not single-aura, so a second copy could sit next to it")
	}

	// Both copies bid the same armor and both are permanent, so the paladin's own
	// cast takes the slot by activating after the external one, and the armor on
	// the character sheet does not move.
	external := char.GetAura("Devotion Aura (External)")
	if !player.IsActive() {
		t.Error("the paladin's own aura did not activate next to the external one")
	}
	if external.IsActive() {
		t.Error("the external copy stayed active, so both copies of the aura are on the character")
	}
	if got, want := char.GetStats()[stats.Armor], armorAmount(t, 10293); got != want {
		t.Errorf("both copies together applied %v armor, want the game files' %v", got, want)
	}
}

// A totem the game files only tie to its cast by name, improved by the talent the
// live tree still prices.
func TestGeneratedManaSpringTotemTakesTheTalentedValue(t *testing.T) {
	char := core.NewGeneratedBuffTestCharacter()

	core.ApplyBuffEffects(char, &proto.RaidBuffs{},
		&proto.PartyBuffs{ManaSpringTotem: proto.TristateEffect_TristateEffectImproved},
		&proto.IndividualBuffs{})
	char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)

	tick := rowEffect(t, 10494, dbcenums.A_PERIODIC_ENERGIZE)
	untalented := rowAmount(t, 10494, dbcenums.A_PERIODIC_ENERGIZE) * 5000 / float64(tick.PeriodMs)
	if got := buffs.ManaSpringTotemValue(0); got != untalented {
		t.Errorf("the untalented totem is worth %v MP5, want the game files' %v per %v as %v", got, tick.Average(), tick.Period(), untalented)
	}
	for points := int32(1); points <= 5; points++ {
		if got, below := buffs.ManaSpringTotemValue(points), buffs.ManaSpringTotemValue(points-1); got < below {
			t.Errorf("the totem with %d/5 of its talent is worth %v MP5, want at least the %v of one point less", points, got, below)
		}
	}
	improved := buffs.ManaSpringTotemValue(5)
	if improved <= untalented {
		t.Errorf("the totem with 5/5 of its talent is worth %v MP5, want more than the untalented %v", improved, untalented)
	}
	if got := char.GetStats()[stats.MP5]; got != improved {
		t.Errorf("the improved totem applied %v MP5, want the talented %v", got, improved)
	}

	aura := char.GetAura("Mana Spring Totem (External)")
	if aura == nil {
		t.Fatalf("no aura is labelled %q; the unit has %v", "Mana Spring Totem (External)", auraLabels(char))
	}
	if want := (core.ActionID{SpellID: 10494, Tag: -1}); aura.ActionID != want {
		t.Errorf("the totem's aura is %v, want the aura family member %v", aura.ActionID, want)
	}

	category := char.ExclusiveEffectManager.GetExclusiveEffectCategory(
		buffs.ManaSpringTotemCategory + stats.MP5.StatName() + "Add")
	if len(category.Effects()) != 1 || category.Effects()[0].Priority != improved {
		t.Errorf("the totem's category holds %d effects, first bid %v; want one bidding %v",
			len(category.Effects()), category.Effects()[0].Priority, improved)
	}
}

// The two rows whose stats the manifest names. Each is worth its game files'
// amount on every kind of crit, and neither touches RangedCritPercent, which the
// sim adds on top of the physical one, so a ranged attack gains the amount once.
//
// Spell 17007 calls Leader of the Pack exclusive with Moonkin Aura, so the two
// share a category that holds one aura: a party with both ticked is worth the
// stronger, and the copy that loses stays registered without applying anything.
func TestGeneratedCritAurasApplyToEveryKindOfCrit(t *testing.T) {
	for _, row := range []struct {
		name  string
		party *proto.PartyBuffs
		want  float64
	}{
		{"Leader of the Pack", &proto.PartyBuffs{LeaderOfThePack: true}, leaderOfThePack(t)},
		{"Moonkin Aura", &proto.PartyBuffs{MoonkinAura: true}, moonkinAura(t)},
		{"both", &proto.PartyBuffs{LeaderOfThePack: true, MoonkinAura: true}, max(leaderOfThePack(t), moonkinAura(t))},
	} {
		t.Run(row.name, func(t *testing.T) {
			char := core.NewGeneratedBuffTestCharacter()

			core.ApplyBuffEffects(char, &proto.RaidBuffs{}, row.party, &proto.IndividualBuffs{})
			char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)

			if got := char.GetStats()[stats.PhysicalCritPercent]; got != row.want {
				t.Errorf("%s applied %v physical crit, want the game files' %v", row.name, got, row.want)
			}
			if got := char.GetStats()[stats.SpellCritPercent]; got != row.want {
				t.Errorf("%s applied %v spell crit, want the game files' %v", row.name, got, row.want)
			}
			if got := char.GetStats()[stats.RangedCritPercent]; got != 0 {
				t.Errorf("%s applied %v on top of the physical crit a ranged attack already reads", row.name, got)
			}
		})
	}
}

func leaderOfThePack(t *testing.T) float64 {
	t.Helper()
	return rowAmount(t, 24932, dbcenums.A_MOD_CRIT_PCT)
}

func moonkinAura(t *testing.T) float64 {
	t.Helper()
	return rowAmount(t, 24907, dbcenums.A_MOD_CRIT_PCT)
}

// The loser of the category is still a registered aura on the character, so a
// class port can find it and a log can name it; it simply holds nothing.
func TestGeneratedCritAurasLeaveTheOutbidCopyRegisteredAndInert(t *testing.T) {
	char := core.NewGeneratedBuffTestCharacter()

	core.ApplyBuffEffects(char, &proto.RaidBuffs{},
		&proto.PartyBuffs{LeaderOfThePack: true, MoonkinAura: true}, &proto.IndividualBuffs{})
	char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)

	pack := char.GetAura("Leader of the Pack (External)")
	moonkin := char.GetAura("Moonkin Aura (External)")
	if pack == nil || moonkin == nil {
		t.Fatalf("the two copies are registered as %v and %v, want both", pack, moonkin)
	}
	if pack.IsActive() == moonkin.IsActive() {
		t.Errorf("both copies are active: %v, want exactly one holding the category", pack.IsActive())
	}

	category := char.ExclusiveEffectManager.GetExclusiveEffectCategory(buffs.LeaderOfThePackCategory)
	if !category.SingleAura {
		t.Error("the DruidCritAura category is not single-aura, so both copies could sit on the character")
	}
	if len(category.Effects()) != 2 {
		t.Errorf("%d effects bid for DruidCritAura, want both copies", len(category.Effects()))
	}
	want := max(leaderOfThePack(t), moonkinAura(t))
	if active := category.GetActiveEffect(); active == nil || active.Priority != want {
		t.Errorf("the category is held at %v, want the stronger copy's %v", active, want)
	}
}

// The row whose only amount is a pseudo-stat the game files state as a reduction.
func TestGeneratedPartyConcentrationAuraReducesPushback(t *testing.T) {
	char := core.NewGeneratedBuffTestCharacter()

	core.ApplyBuffEffects(char, &proto.RaidBuffs{},
		&proto.PartyBuffs{ConcentrationAura: true}, &proto.IndividualBuffs{})
	char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)

	reduction := -rowAmount(t, 19746, dbcenums.A_REDUCE_PUSHBACK) / 100
	if reduction >= 0 {
		t.Fatalf("Concentration Aura is worth %v pushback chance, want a reduction", reduction)
	}
	if got, want := char.PseudoStats.PushbackChance, 1+reduction; got != want {
		t.Errorf("Concentration Aura left the pushback chance at %v, want 1 reduced by the game files' %v, %v", got, -reduction, want)
	}
}

// The rank of Trueshot Aura the resolver takes states one
// A_MOD_RANGED_ATTACK_POWER effect: melee attack power is untouched.
func TestGeneratedTrueshotAuraGivesTheTopRanksRangedAttackPower(t *testing.T) {
	char := core.NewGeneratedBuffTestCharacter()

	core.ApplyBuffEffects(char,
		&proto.RaidBuffs{}, &proto.PartyBuffs{TrueshotAura: true}, &proto.IndividualBuffs{})
	char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)

	if got, want := char.GetStats()[stats.RangedAttackPower], rowAmount(t, 20906, dbcenums.A_MOD_RANGED_ATTACK_POWER); got != want {
		t.Errorf("Trueshot Aura applied %v ranged attack power, want the game files' %v", got, want)
	}
	if got := char.GetStats()[stats.AttackPower]; got != 0 {
		t.Errorf("Trueshot Aura applied %v melee attack power, want none", got)
	}

	aura := char.GetAura("Trueshot Aura (External)")
	if aura == nil {
		t.Fatalf("no aura is labelled %q; the unit has %v", "Trueshot Aura (External)", auraLabels(char))
	}
	if want := (core.ActionID{SpellID: 20906, Tag: -1}); aura.ActionID != want {
		t.Errorf("the external copy is %v, want %v", aura.ActionID, want)
	}
	if aura.Duration != core.NeverExpires {
		t.Errorf("the party's Trueshot Aura lasts %v, want it permanent", aura.Duration)
	}
}

// The build phase enables a multiplying dependency without recomputing the
// unit's stats; applyAllEffects measures them afterwards, and so does this.
func measureGeneratedBuffStats(char *core.Character) {
	char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)
	char.SetStats(char.SortAndApplyStatDependencies(char.GetStats()).FloorGameStats())
}

// Grace of Air stands for the fight on its own, and gives way to Windfury
// Totem: under both, the party gets no agility from it.
func TestGeneratedGraceOfAirGivesWayToWindfuryTotem(t *testing.T) {
	standing := core.NewGeneratedBuffTestCharacter()
	core.ApplyBuffEffects(standing,
		&proto.RaidBuffs{}, &proto.PartyBuffs{GraceOfAirTotem: true}, &proto.IndividualBuffs{})
	standing.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)

	if got, want := standing.GetStats()[stats.Agility], graceOfAirTotemAgility(t); got != want {
		t.Errorf("the totem applied %v agility, want the game files' %v", got, want)
	}
	aura := standing.GetAura("Grace of Air Totem (External)")
	if aura == nil {
		t.Fatalf("no aura is labelled %q; the unit has %v", "Grace of Air Totem (External)", auraLabels(standing))
	}
	if aura.Duration != core.NeverExpires {
		t.Errorf("the totem's aura lasts %v, want it to stand for the fight", aura.Duration)
	}

	sim := setupFakeSimWithBuffs(&proto.RaidBuffs{},
		&proto.PartyBuffs{GraceOfAirTotem: true}, &proto.IndividualBuffs{})
	permanent := sim.Raid.Parties[0].Players[0].GetCharacter().GetAura("Grace of Air Totem (External)")
	if !permanent.IsActive() {
		t.Error("the totem is down at the pull, want it up for the fight")
	}

	underWindfury := core.NewGeneratedBuffTestCharacter()
	core.ApplyBuffEffects(underWindfury, &proto.RaidBuffs{},
		&proto.PartyBuffs{GraceOfAirTotem: true, WindfuryTotem: true}, &proto.IndividualBuffs{})
	underWindfury.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)

	if got := underWindfury.GetStats()[stats.Agility]; got != 0 {
		t.Errorf("under Windfury Totem the totem applied %v agility, want none", got)
	}
	if underWindfury.GetAura("Grace of Air Totem (External)") != nil {
		t.Error("under Windfury Totem the totem registered its aura")
	}
}

func graceOfAirTotemAgility(t *testing.T) float64 {
	t.Helper()
	return rowAmount(t, 25360, dbcenums.A_MOD_STAT)
}

// Each staff in the party is worth its own copy of the aura's amounts: MP5 per
// druid staff, spell crit per mage one, healing per priest one and spell damage
// plus healing per warlock one.
func TestGeneratedAtieshStavesCountTheStaves(t *testing.T) {
	char := core.NewGeneratedBuffTestCharacter()

	core.ApplyBuffEffects(char, &proto.RaidBuffs{},
		&proto.PartyBuffs{AtieshDruid: 1, AtieshMage: 3, AtieshPriest: 2, AtieshWarlock: 1},
		&proto.IndividualBuffs{})
	char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)

	for stat, want := range map[stats.Stat]float64{
		stats.MP5:              1 * rowAmount(t, 28145, dbcenums.A_MOD_POWER_REGEN),
		stats.SpellCritPercent: 3 * rowAmount(t, 28142, dbcenums.A_MOD_SPELL_CRIT_CHANCE),
		stats.HealingPower:     2*rowAmount(t, 28144, dbcenums.A_MOD_HEALING_DONE) + 1*rowAmount(t, 28143, dbcenums.A_MOD_HEALING_DONE),
		stats.SpellDamage:      1 * rowAmount(t, 28143, dbcenums.A_MOD_DAMAGE_DONE),
	} {
		if got := char.GetStats()[stat]; got != want {
			t.Errorf("the party's staves applied %v %s, want %v", got, stat.StatName(), want)
		}
	}

	if char.GetAura("Atiesh - Mage (External)") == nil {
		t.Fatalf("the unit has %v, want an aura for the mage's staff", auraLabels(char))
	}
}

// The party's Darkspear is up for a share of the fight only, so the stats the character sheet and
// the EP are measured with carry none of its crit.
func TestGeneratedPartyDarkspearStaysOffTheCharacterSheet(t *testing.T) {
	char := core.NewGeneratedBuffTestCharacter()

	core.ApplyBuffEffects(char, &proto.RaidBuffs{}, &proto.PartyBuffs{Darkspear: true}, &proto.IndividualBuffs{})
	char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)

	aura := char.GetAura("Darkspear (External)")
	if aura == nil {
		t.Fatalf("the unit has %v, want the party's Darkspear", auraLabels(char))
	}
	if aura.BuildPhase != core.CharacterBuildPhaseNone {
		t.Errorf("the party's Darkspear is in build phase %v, want none", aura.BuildPhase)
	}
	for _, stat := range []stats.Stat{stats.PhysicalCritPercent, stats.SpellCritPercent} {
		if got := char.GetStats()[stat]; got != 0 {
			t.Errorf("the party's Darkspear applied %v %s to the measured stats, want 0", got, stat.StatName())
		}
	}
}

const darkspearItemID = 12802

// A raid of one in a party with the Darkspear party buff, its database the one the request carries.
func darkspearRaid(db *proto.SimDatabase) *proto.Raid {
	return &proto.Raid{
		Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name:      "Caster",
				Class:     proto.Class_ClassShaman,
				Buffs:     &proto.IndividualBuffs{},
				Spec:      &proto.Player_ElementalShaman{},
				Equipment: &proto.EquipmentSpec{},
				Database:  db,
			}},
			Buffs: &proto.PartyBuffs{Darkspear: true},
		}},
		Buffs: &proto.RaidBuffs{},
	}
}

// The Darkspear row as a browser request carries it: the UI database's item, cut down to the
// fields the sim reads.
func darkspearRequestItem(t *testing.T) *proto.SimItem {
	t.Helper()
	for _, item := range database.Load().Items {
		if item.Id != darkspearItemID {
			continue
		}
		data, err := protojson.Marshal(item)
		if err != nil {
			t.Fatal(err)
		}
		simItem := &proto.SimItem{}
		if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(data, simItem); err != nil {
			t.Fatal(err)
		}
		return simItem
	}
	t.Fatalf("the database has no item %d", darkspearItemID)
	return nil
}

// Without the full item database the sim knows only the items the request carries; the party's
// Darkspear then stops on a message that names the item the request is missing.
func TestPartyDarkspearNamesTheItemTheRequestMustCarry(t *testing.T) {
	restore := core.KeepOnlyItems(nil)
	defer restore()

	message := func() (message string) {
		defer func() { message = fmt.Sprint(recover()) }()
		core.ComputeStats(&proto.ComputeStatsRequest{Raid: darkspearRaid(nil), SkipRotation: true})
		return
	}()
	if !strings.Contains(message, fmt.Sprintf("item %d", darkspearItemID)) {
		t.Errorf("the party's Darkspear without its item stopped on %q, want a message naming item %d", message, darkspearItemID)
	}
}

// A request that carries the Darkspear item computes stats with the party's Darkspear on.
func TestPartyDarkspearReadsTheItemTheRequestCarries(t *testing.T) {
	item := darkspearRequestItem(t)
	restore := core.KeepOnlyItems(nil)
	defer restore()

	raid := darkspearRaid(&proto.SimDatabase{Items: []*proto.SimItem{item}})
	result := core.ComputeStats(&proto.ComputeStatsRequest{Raid: raid, SkipRotation: true})
	if result.ErrorResult != "" {
		t.Fatalf("ComputeStats: %s", result.ErrorResult)
	}
	darkspear := buffs.NewSpellDataDarkspearAura(&core.NewGeneratedBuffTestCharacter().Unit, false, 0).ActionID.ToProto()
	for _, aura := range result.RaidStats.Parties[0].Players[0].Metadata.Auras {
		if aura.Id.GetSpellId() == darkspear.GetSpellId() {
			return
		}
	}
	t.Errorf("the character carries no aura of spell %d, want the party's Darkspear", darkspear.GetSpellId())
}

// The windfury proc grants the attack power the game files state for as long as
// the aura lasts; the totem aura around it holds the category and the trigger.
func TestGeneratedWindfuryTotemProcAppliesItsAttackPower(t *testing.T) {
	sim := setupFakeSimWithBuffs(&proto.RaidBuffs{}, &proto.PartyBuffs{WindfuryTotem: true}, &proto.IndividualBuffs{})
	char := sim.Raid.Parties[0].Players[0].GetCharacter()

	proc := char.GetAura("Windfury Totem (External)")
	if proc == nil {
		t.Fatalf("no aura is labelled %q; the unit has %v", "Windfury Totem (External)", auraLabels(char))
	}
	if want := rowDuration(t, 10610); proc.Duration != want {
		t.Errorf("the proc buff lasts %v, want the game files' %v", proc.Duration, want)
	}

	before := char.GetStats()[stats.AttackPower]
	proc.Activate(sim)
	if got, want := char.GetStats()[stats.AttackPower]-before, windfuryTotem(t); got != want {
		t.Errorf("the proc applied %v attack power, want the game files' %v", got, want)
	}
	proc.Deactivate(sim)
	if got := char.GetStats()[stats.AttackPower]; got != before {
		t.Errorf("the proc left %v attack power behind when it expired", got-before)
	}

	totem := char.GetAura("Windfury Totem")
	if totem == nil {
		t.Fatalf("no aura is labelled %q; the unit has %v", "Windfury Totem", auraLabels(char))
	}
	category := char.ExclusiveEffectManager.GetExclusiveEffectCategory(buffs.WindfuryTotemCategory)
	if bid := windfuryTotem(t); len(category.Effects()) != 1 || category.Effects()[0].Priority != bid {
		t.Errorf("the category holds %d effects, first bid %v; want one bidding %v",
			len(category.Effects()), category.Effects()[0].Priority, bid)
	}
	if char.GetAura("Windfury Totem Trigger") == nil {
		t.Error("the driver registered no proc trigger for the totem")
	}
}

func windfuryTotem(t *testing.T) float64 {
	t.Helper()
	return rowAmount(t, 10610, dbcenums.A_MOD_ATTACK_POWER)
}

// The proc holds the game files' charges and each landed auto attack, in either
// hand, spends one. A melee special or a missed swing spends none, and a proc no
// auto lands on keeps its attack power for its whole duration.
func TestGeneratedWindfuryTotemProcSpendsItsChargesOnAutoAttacks(t *testing.T) {
	sim := setupFakeSimWithBuffs(&proto.RaidBuffs{}, &proto.PartyBuffs{WindfuryTotem: true}, &proto.IndividualBuffs{})
	char := sim.Raid.Parties[0].Players[0].GetCharacter()
	proc := char.GetAura("Windfury Totem (External)")
	if proc == nil {
		t.Fatalf("no aura is labelled %q; the unit has %v", "Windfury Totem (External)", auraLabels(char))
	}

	mhAuto := &core.Spell{ProcMask: core.ProcMaskMeleeMHAuto}
	ohAuto := &core.Spell{ProcMask: core.ProcMaskMeleeOHAuto}
	special := &core.Spell{ProcMask: core.ProcMaskMeleeMHSpecial}
	landed := &core.SpellResult{Target: &char.Unit, Outcome: core.OutcomeHit, Damage: 100}
	missed := &core.SpellResult{Target: &char.Unit, Outcome: core.OutcomeMiss}
	swing := func(spell *core.Spell, result *core.SpellResult) {
		proc.OnSpellHitDealt(proc, sim, spell, result)
	}

	charges := proc.MaxStacks
	if charges < 2 {
		t.Fatalf("the proc holds %d charges, want at least 2 to spend one in each hand", charges)
	}
	fullCharges := func() {
		proc.Activate(sim)
		proc.SetStacks(sim, charges)
	}

	before := char.GetStats()[stats.AttackPower]
	fullCharges()

	swing(special, landed)
	swing(mhAuto, missed)
	if got := proc.GetStacks(); got != charges {
		t.Errorf("a special and a missed auto left %d charges, want all %d still there", got, charges)
	}

	swing(mhAuto, landed)
	if !proc.IsActive() || proc.GetStacks() != charges-1 {
		t.Fatalf("one landed auto left the proc active %v with %d charges, want it up with %d",
			proc.IsActive(), proc.GetStacks(), charges-1)
	}
	if got, want := char.GetStats()[stats.AttackPower]-before, windfuryTotem(t); got != want {
		t.Errorf("the proc holds %v attack power with a charge spent, want the game files' %v", got, want)
	}

	for range charges - 2 {
		swing(mhAuto, landed)
	}
	swing(ohAuto, landed)
	if proc.IsActive() {
		t.Error("the proc is still up after its last landed auto, an off-hand one, want the charges spent")
	}
	if got := char.GetStats()[stats.AttackPower]; got != before {
		t.Errorf("the spent proc left %v attack power behind", got-before)
	}

	fullCharges()
	start := sim.CurrentTime
	swing(special, landed)
	if !proc.IsActive() || proc.RemainingDuration(sim) != proc.Duration {
		t.Fatalf("a proc no auto landed on is active %v with %v left, want the full %v",
			proc.IsActive(), proc.RemainingDuration(sim), proc.Duration)
	}
	sim.AddPendingAction(core.NewDelayedAction(core.DelayedActionOptions{
		DoAt:     start + proc.Duration*2,
		OnAction: func(*core.Simulation) {},
	}))
	for proc.IsActive() && sim.CurrentTime < start+proc.Duration*2 {
		sim.Step()
	}
	if proc.IsActive() || proc.GetStacks() != 0 {
		t.Errorf("the lingering proc is active %v with %d charges at %v, want it gone after its %v",
			proc.IsActive(), proc.GetStacks(), sim.CurrentTime-start, proc.Duration)
	}
}

// Windfury Totem's proc is the party aura 10612: a main-hand hit that deals
// damage can fire it, auto or special, and no other hit can, a proc's hit
// included. It rolls off the hit's cast result, so the proc starts with every
// charge whichever main-hand hit fired it.
func TestGeneratedWindfuryTotemProcsOffMainHandHitsWithEveryCharge(t *testing.T) {
	setup := func(t *testing.T) (*core.Simulation, *core.Unit, *core.Aura, *[]int32) {
		t.Helper()

		sim := setupFakeSimWithBuffs(&proto.RaidBuffs{}, &proto.PartyBuffs{WindfuryTotem: true}, &proto.IndividualBuffs{})
		char := sim.Raid.Parties[0].Players[0].GetCharacter()
		trigger := char.GetAura("Windfury Totem Trigger")
		proc := char.GetAura("Windfury Totem (External)")
		if trigger == nil || proc == nil {
			t.Fatalf("trigger %v, proc %v; want both registered", trigger, proc)
		}
		if !trigger.IsActive() {
			t.Fatal("the totem's trigger is not up at the pull")
		}

		var started []int32
		proc.ApplyOnStacksChange(func(_ *core.Aura, _ *core.Simulation, oldStacks int32, newStacks int32) {
			if oldStacks == 0 {
				started = append(started, newStacks)
			}
		})
		return sim, sim.Encounter.AllTargetUnits[0], trigger, &started
	}

	// Offers the trigger up to 200 landed hits of one kind and answers whether any of them procced.
	offer := func(sim *core.Simulation, target *core.Unit, trigger *core.Aura, started *[]int32, spell *core.Spell, damage float64) bool {
		for i := 0; i < 200 && len(*started) == 0; i++ {
			trigger.OnCastResult(trigger, sim, spell, &core.SpellResult{Target: target, Outcome: core.OutcomeHit, Damage: damage})
		}
		return len(*started) > 0
	}

	for _, row := range []struct {
		name  string
		spell *core.Spell
	}{
		{"a main-hand auto", &core.Spell{ProcMask: core.ProcMaskMeleeMHAuto}},
		{"a main-hand special", &core.Spell{ProcMask: core.ProcMaskMeleeMHSpecial}},
	} {
		t.Run(row.name, func(t *testing.T) {
			sim, target, trigger, started := setup(t)
			charges := trigger.Unit.GetAura("Windfury Totem (External)").MaxStacks
			if !offer(sim, target, trigger, started, row.spell, 100) {
				t.Fatal("200 landed hits never procced the totem")
			}
			if got := (*started)[0]; got != charges {
				t.Errorf("%s procced it with %d of its %d charges, want all of them", row.name, got, charges)
			}
		})
	}

	for _, row := range []struct {
		name   string
		spell  *core.Spell
		damage float64
	}{
		{"an off-hand auto", &core.Spell{ProcMask: core.ProcMaskMeleeOHAuto}, 100},
		{"an off-hand special", &core.Spell{ProcMask: core.ProcMaskMeleeOHSpecial}, 100},
		{"a ranged auto", &core.Spell{ProcMask: core.ProcMaskRangedAuto}, 100},
		{"a ranged special", &core.Spell{ProcMask: core.ProcMaskRangedSpecial}, 100},
		{"a spell", &core.Spell{ProcMask: core.ProcMaskSpellDamage}, 100},
		{"a main-hand special that deals no damage", &core.Spell{ProcMask: core.ProcMaskMeleeMHSpecial}, 0},
		{"a main-hand proc's hit", &core.Spell{ProcMask: core.ProcMaskMeleeMHSpecial, Flags: core.SpellFlagProc}, 100},
	} {
		t.Run(row.name+" never procs it", func(t *testing.T) {
			sim, target, trigger, started := setup(t)
			if offer(sim, target, trigger, started, row.spell, row.damage) {
				t.Errorf("%s procced the totem", row.name)
			}
		})
	}

	t.Run("the trigger waits the party aura's cooldown between procs", func(t *testing.T) {
		_, _, trigger, _ := setup(t)
		if want := spelldata.MustFind(10612).ICD(); trigger.Icd == nil || trigger.Icd.Duration != want {
			t.Errorf("the trigger's cooldown is %v, want the game files' %v", trigger.Icd, want)
		}
	})
}

// Windfury Totem's proc puts every charge up at the hit's cast result. The hit
// that fired it is dealt first and spends a charge if it is an auto; the extra
// attack swings after it, in the same instant, and spends the next if it lands.
// A special spends none, so its extra attack and the auto after it spend them.
func TestGeneratedWindfuryTotemSpendsItsChargesOnTheHitThenTheExtraAttack(t *testing.T) {
	setup := func(t *testing.T) (*core.Simulation, *core.Unit, *core.Aura, *core.Aura, *core.Spell, *[]int32) {
		t.Helper()
		sim, owner := setupPetOwnerSim(&proto.PartyBuffs{WindfuryTotem: true})
		for _, unit := range slices.Concat(sim.Raid.AllUnits, sim.Encounter.AllTargetUnits) {
			unit.AutoAttacks.CancelAutoSwing(sim)
		}
		pet := owner.pet
		trigger := pet.GetAura("Windfury Totem Trigger")
		proc := pet.GetAura("Windfury Totem (External)")
		if trigger == nil || proc == nil || !trigger.IsActive() {
			t.Fatalf("the pet's trigger is %v and proc %v, want both registered and the trigger up", trigger, proc)
		}
		if proc.MaxStacks < 2 {
			t.Fatalf("the proc holds %d charges, want at least 2", proc.MaxStacks)
		}
		extraAttack := pet.GetSpell(pet.AutoAttacks.MHAuto().ActionID.WithTag(proc.ActionID.SpellID))
		if extraAttack == nil {
			t.Fatal("the pet registered no extra attack under the totem's tag")
		}
		var changes []int32
		proc.ApplyOnStacksChange(func(_ *core.Aura, _ *core.Simulation, oldStacks int32, newStacks int32) {
			if newStacks != oldStacks {
				changes = append(changes, newStacks)
			}
		})
		return sim, sim.Encounter.AllTargetUnits[0], trigger, proc, extraAttack, &changes
	}
	casts := func(spell *core.Spell, target *core.Unit) int32 {
		return spell.SpellMetrics[target.UnitIndex].Casts
	}
	landed := func(spell *core.Spell, target *core.Unit) int32 {
		m := spell.SpellMetrics[target.UnitIndex]
		return m.Casts - m.Misses - m.Dodges - m.Parries
	}
	// Steps the sim until the proc's extra attack has swung, which it does without the clock moving.
	swingExtraAttack := func(t *testing.T, sim *core.Simulation, extraAttack *core.Spell, target *core.Unit) {
		t.Helper()
		at, swung := sim.CurrentTime, casts(extraAttack, target)
		for i := 0; i < 10 && casts(extraAttack, target) == swung; i++ {
			sim.Step()
		}
		if got := casts(extraAttack, target) - swung; got != 1 {
			t.Fatalf("the proc swung %d extra attacks, want 1", got)
		}
		if sim.CurrentTime != at {
			t.Fatalf("the extra attack swung at %v, want the triggering hit's %v", sim.CurrentTime, at)
		}
	}

	t.Run("a main-hand auto", func(t *testing.T) {
		sim, target, trigger, proc, extraAttack, changes := setup(t)
		mhAuto := proc.Unit.AutoAttacks.MHAuto()
		charges := proc.MaxStacks
		seen := map[bool]bool{}
		for i := 0; i < 2000 && !(seen[true] && seen[false]); i++ {
			*changes = nil
			swung, hits := casts(extraAttack, target), landed(extraAttack, target)
			mhAuto.Cast(sim, target)
			if len(*changes) == 0 {
				continue
			}
			if got, want := *changes, []int32{charges, charges - 1}; !slices.Equal(got, want) {
				t.Fatalf("the auto that procced the totem left the charges going %v, want %v", got, want)
			}
			if casts(extraAttack, target) != swung {
				t.Fatal("the extra attack swung before the auto that procced it was dealt")
			}
			swingExtraAttack(t, sim, extraAttack, target)
			extraLanded := landed(extraAttack, target) > hits
			want := []int32{charges, charges - 1}
			if extraLanded {
				want = append(want, charges-2)
			}
			if got := *changes; !slices.Equal(got, want) {
				t.Errorf("with the extra attack landed %v the charges went %v, want %v", extraLanded, got, want)
			}
			seen[extraLanded] = true
			proc.Deactivate(sim)
			trigger.Icd.Reset()
		}
		if !seen[true] || !seen[false] {
			t.Errorf("2000 autos saw an extra attack land %v and miss %v, want both", seen[true], seen[false])
		}
	})

	t.Run("a main-hand special", func(t *testing.T) {
		sim, target, trigger, proc, extraAttack, changes := setup(t)
		special := &core.Spell{ProcMask: core.ProcMaskMeleeMHSpecial}
		hit := &core.SpellResult{Target: target, Outcome: core.OutcomeHit, Damage: 100}
		charges := proc.MaxStacks
		for i := 0; i < 2000; i++ {
			*changes = nil
			trigger.OnCastResult(trigger, sim, special, hit)
			if len(*changes) == 0 {
				continue
			}
			proc.OnSpellHitDealt(proc, sim, special, hit)
			if got, want := *changes, []int32{charges}; !slices.Equal(got, want) {
				t.Fatalf("the special that procced the totem left the charges going %v, want %v", got, want)
			}
			hits := landed(extraAttack, target)
			swingExtraAttack(t, sim, extraAttack, target)
			if landed(extraAttack, target) == hits {
				proc.Deactivate(sim)
				trigger.Icd.Reset()
				continue
			}
			proc.OnSpellHitDealt(proc, sim, &core.Spell{ProcMask: core.ProcMaskMeleeMHAuto}, hit)
			if got, want := *changes, []int32{charges, charges - 1, charges - 2}; !slices.Equal(got, want) {
				t.Errorf("the special's extra attack and the auto after it left the charges going %v, want %v", got, want)
			}
			return
		}
		t.Error("2000 specials never procced the totem with an extra attack that landed")
	})
}

// The fire damage a main-hand hit adds is proportional to the weapon's speed.
func TestGeneratedFlametongueTotemDamageFollowsWeaponSpeed(t *testing.T) {
	slow, fast := buffs.FlametongueTotemDamage(2.6), buffs.FlametongueTotemDamage(1.3)
	if fast <= 0 || !core.WithinToleranceFloat64(2*fast, slow, 1e-9) {
		t.Errorf("a 2.6 speed weapon procs %v and a 1.3 speed one %v, want twice as much and above 0", slow, fast)
	}
}

// A main-hand auto attack that lands strikes the target for the game files'
// fire damage scaled by the weapon's speed, on the spell hit and crit tables; a
// special, an off-hand hit or a missed swing strikes nothing.
func TestGeneratedFlametongueTotemStrikesOnMainHandHits(t *testing.T) {
	sim := setupFakeSimWithBuffs(&proto.RaidBuffs{}, &proto.PartyBuffs{FlametongueTotem: true}, &proto.IndividualBuffs{})
	char := sim.Raid.Parties[0].Players[0].GetCharacter()
	target := sim.Encounter.AllTargetUnits[0]

	trigger := char.GetAura("Flametongue Totem Trigger")
	if trigger == nil {
		t.Fatalf("no aura is labelled %q; the unit has %v", "Flametongue Totem Trigger", auraLabels(char))
	}
	if !trigger.IsActive() {
		t.Fatal("the totem's trigger is not up at the pull")
	}
	category := char.ExclusiveEffectManager.GetExclusiveEffectCategory(buffs.FlametongueTotemCategory)
	if len(category.Effects()) != 1 {
		t.Errorf("the category holds %d effects, want the totem's one", len(category.Effects()))
	}

	strike := char.GetSpell(core.ActionID{SpellID: 16387, Tag: 1})
	if strike == nil {
		t.Fatal("the totem registered no spell for its proc")
	}
	if strike.SpellSchool != core.SpellSchoolFire || strike.DefenseType != core.DefenseTypeMagic {
		t.Errorf("the proc is %v on %v, want fire on the spell tables", strike.SpellSchool, strike.DefenseType)
	}
	if !strike.Flags.Matches(core.SpellFlagProc) {
		t.Error("the proc's hits are not flagged as a proc, which 16368 does not opt out of")
	}

	metrics := &strike.SpellMetrics[target.UnitIndex]
	offer := func(procMask core.ProcMask, outcome core.HitOutcome) (casts int32, damage float64, crit bool) {
		casts, damage, crits := metrics.Casts, metrics.TotalDamage, metrics.Crits
		trigger.OnSpellHitDealt(trigger, sim, &core.Spell{ProcMask: procMask},
			&core.SpellResult{Target: target, Outcome: outcome, Damage: 100})
		return metrics.Casts - casts, metrics.TotalDamage - damage, metrics.Crits > crits
	}

	for _, row := range []struct {
		name     string
		speed    float64
		procMask core.ProcMask
	}{
		{"a 2.6 speed main-hand auto", 2.6, core.ProcMaskMeleeMHAuto},
		{"a 1.3 speed main-hand auto", 1.3, core.ProcMaskMeleeMHAuto},
	} {
		char.AutoAttacks.SetMH(core.Weapon{SwingSpeed: row.speed})
		want := buffs.FlametongueTotemDamage(row.speed)
		landedDamage := false
		for i := 0; i < 20; i++ {
			casts, damage, crit := offer(row.procMask, core.OutcomeHit)
			if casts != 1 {
				t.Fatalf("%s cast the proc %d times, want once", row.name, casts)
			}
			if damage == 0 {
				continue
			}
			landedDamage = true
			if expected := core.TernaryFloat64(crit, want*1.5, want); math.Abs(damage-expected) > 1e-9 {
				t.Errorf("%s struck for %v (crit %v), want %v", row.name, damage, crit, expected)
			}
		}
		if !landedDamage {
			t.Errorf("20 of %s dealt no damage", row.name)
		}
	}

	for _, row := range []struct {
		name     string
		procMask core.ProcMask
		outcome  core.HitOutcome
	}{
		{"a main-hand special", core.ProcMaskMeleeMHSpecial, core.OutcomeHit},
		{"an off-hand auto", core.ProcMaskMeleeOHAuto, core.OutcomeHit},
		{"a ranged auto", core.ProcMaskRangedAuto, core.OutcomeHit},
		{"a missed main-hand auto", core.ProcMaskMeleeMHAuto, core.OutcomeMiss},
	} {
		if casts, _, _ := offer(row.procMask, row.outcome); casts != 0 {
			t.Errorf("%s cast the proc %d times, want none", row.name, casts)
		}
	}
}

// The totem is a party aura, so it stands for the fight and a weapon swapped
// into the main hand keeps it.
func TestGeneratedFlametongueTotemStaysThroughAMainHandSwap(t *testing.T) {
	const arcaniteReaper, whirlwindAxe = 12784, 6975
	withMainHand := func(id int32) []*proto.ItemSpec {
		items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
		for i := range items {
			items[i] = &proto.ItemSpec{}
		}
		items[proto.ItemSlot_ItemSlotMainHand].Id = id
		return items
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 100},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name:           "Caster",
				Class:          proto.Class_ClassShaman,
				Buffs:          &proto.IndividualBuffs{},
				Spec:           &proto.Player_ElementalShaman{},
				Equipment:      &proto.EquipmentSpec{Items: withMainHand(arcaniteReaper)},
				EnableItemSwap: true,
				ItemSwap:       &proto.ItemSwap{Items: withMainHand(whirlwindAxe)},
			}},
			Buffs: &proto.PartyBuffs{FlametongueTotem: true},
		}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()
	char := sim.Raid.Parties[0].Players[0].GetCharacter()

	totem := char.GetAura("Flametongue Totem (External)")
	trigger := char.GetAura("Flametongue Totem Trigger")
	if totem == nil || trigger == nil {
		t.Fatalf("totem %v, trigger %v; want both registered", totem, trigger)
	}
	if totem.Duration != core.NeverExpires {
		t.Errorf("the totem lasts %v, want it to stand for the fight", totem.Duration)
	}
	if !totem.IsActive() || !trigger.IsActive() {
		t.Fatalf("at the pull the totem is active %v and its trigger %v, want both up",
			totem.IsActive(), trigger.IsActive())
	}

	char.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)
	if !totem.IsActive() || !trigger.IsActive() {
		t.Errorf("after the swap the totem is active %v and its trigger %v, want both still up",
			totem.IsActive(), trigger.IsActive())
	}
}

// The two totems do not stack: under both, Windfury Totem holds and Flametongue
// Totem registers nothing.
func TestGeneratedWindfuryTotemHoldsOverFlametongueTotem(t *testing.T) {
	sim := setupFakeSimWithBuffs(&proto.RaidBuffs{},
		&proto.PartyBuffs{WindfuryTotem: true, FlametongueTotem: true}, &proto.IndividualBuffs{})
	char := sim.Raid.Parties[0].Players[0].GetCharacter()

	if char.GetAura("Flametongue Totem Trigger") != nil || char.GetSpell(core.ActionID{SpellID: 16387, Tag: 1}) != nil {
		t.Error("Flametongue Totem registered its proc next to Windfury Totem")
	}
	windfury := char.GetAura("Windfury Totem")
	for _, effect := range char.ExclusiveEffectManager.GetExclusiveEffectCategory(buffs.FlametongueTotemCategory).Effects() {
		if effect.Aura != windfury {
			t.Errorf("%q bids in Flametongue Totem's category next to Windfury Totem", effect.Aura.Label)
		}
	}

	trigger := char.GetAura("Windfury Totem Trigger")
	proc := char.GetAura("Windfury Totem (External)")
	if trigger == nil || !trigger.IsActive() || proc == nil {
		t.Fatalf("Windfury Totem's trigger %v, proc %v; want both registered and the trigger up", trigger, proc)
	}
	target := sim.Encounter.AllTargetUnits[0]
	for i := 0; i < 200 && !proc.IsActive(); i++ {
		trigger.OnCastResult(trigger, sim, &core.Spell{ProcMask: core.ProcMaskMeleeMHSpecial},
			&core.SpellResult{Target: target, Outcome: core.OutcomeHit, Damage: 100})
	}
	if !proc.IsActive() {
		t.Error("200 landed main-hand hits never procced Windfury Totem")
	}
}

// What each test shaman stands in Initialize, set by the test before it builds the sim.
var totemSetup func(c *core.Character)

type totemShaman struct {
	core.Character
}

func init() {
	core.RegisterAgentFactory(
		proto.Player_EnhancementShaman{},
		proto.Spec_SpecEnhancementShaman,
		func(character *core.Character, _ *proto.Player, _ *proto.Raid) core.Agent {
			return &totemShaman{Character: *character}
		},
		func(player *proto.Player, spec interface{}) {
			player.Spec = spec.(*proto.Player_EnhancementShaman)
		},
	)
}

func (c *totemShaman) GetCharacter() *core.Character       { return &c.Character }
func (c *totemShaman) ApplyTalents()                       {}
func (c *totemShaman) Reset(_ *core.Simulation)            {}
func (c *totemShaman) OnGCDReady(_ *core.Simulation)       {}
func (c *totemShaman) OnEncounterStart(_ *core.Simulation) {}
func (c *totemShaman) Initialize()                         { totemSetup(&c.Character) }

// A shaman's own weapon totem, the way the shaman stands one: an aura that bids
// for the totem's category and turns its proc on while it holds it.
func standWeaponTotem(c *core.Character, label string, category string, priority float64, procTrigger *core.Aura) *core.Aura {
	totem := c.RegisterAura(core.Aura{Label: label, Duration: time.Minute * 2})
	totem.NewExclusiveEffect(category, false, core.ExclusiveEffect{
		Priority: priority,
		OnGain: func(_ *core.ExclusiveEffect, sim *core.Simulation) {
			procTrigger.Activate(sim)
		},
		OnExpire: func(_ *core.ExclusiveEffect, sim *core.Simulation) {
			procTrigger.Deactivate(sim)
		},
	})
	return totem
}

func standWindfuryTotem(c *core.Character) *core.Aura {
	totem := standWeaponTotem(c, "Windfury Totem (Self)", buffs.WindfuryTotemCategory,
		buffs.WindfuryTotemValue(0), buffs.NewWindfuryTotemProcTrigger(c))
	buffs.HoldOverWindfuryTotemRivals(totem)
	return totem
}

// One shaman under the party buffs, standing what setup gives it.
func newTotemSim(t *testing.T, party *proto.PartyBuffs, setup func(c *core.Character)) (*core.Simulation, *core.Character) {
	t.Helper()
	totemSetup = setup
	t.Cleanup(func() { totemSetup = nil })

	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 100},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name:      "Shaman",
				Class:     proto.Class_ClassShaman,
				Buffs:     &proto.IndividualBuffs{},
				Spec:      &proto.Player_EnhancementShaman{},
				Equipment: &proto.EquipmentSpec{},
			}},
			Buffs: party,
		}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

var graceOfAirAgility = core.ExclusiveStatCategory(buffs.GraceOfAirTotemCategory, stats.Agility.StatName(), false)

// A Windfury Totem of the shaman's own casting stands only part of the fight:
// while it is down, the party's Grace of Air and Flametongue Totem work, and
// while it stands, neither does.
func TestGeneratedOwnWindfuryTotemHoldsOverThePartysTotems(t *testing.T) {
	var windfury *core.Aura
	sim, char := newTotemSim(t, &proto.PartyBuffs{GraceOfAirTotem: true, FlametongueTotem: true}, func(c *core.Character) {
		windfury = standWindfuryTotem(c)
	})
	graceOfAir := char.GetAura("Grace of Air Totem (External)")
	flametongue := char.GetAura("Flametongue Totem Trigger")
	agility := char.ExclusiveEffectManager.GetExclusiveEffectCategory(graceOfAirAgility)

	before := char.GetStats()[stats.Agility]
	if agility.GetActiveAura() != graceOfAir || !flametongue.IsActive() {
		t.Fatalf("before the totem the agility is held by %v and Flametongue Totem's trigger is up %v, want Grace of Air and up",
			agility.GetActiveAura(), flametongue.IsActive())
	}

	windfury.Activate(sim)
	if agility.GetActiveAura() == graceOfAir {
		t.Error("under the shaman's Windfury Totem, Grace of Air still holds the agility")
	}
	if got, want := before-char.GetStats()[stats.Agility], graceOfAirTotemAgility(t); got != want {
		t.Errorf("the totem took %v agility away, want Grace of Air's %v", got, want)
	}
	if flametongue.IsActive() {
		t.Error("under the shaman's Windfury Totem, Flametongue Totem's trigger is up")
	}
	if !graceOfAir.IsActive() {
		t.Error("the party's Grace of Air went down, want it standing and doing nothing")
	}

	windfury.Deactivate(sim)
	if agility.GetActiveAura() != graceOfAir || char.GetStats()[stats.Agility] != before {
		t.Errorf("after the totem the agility is held by %v at %v, want Grace of Air back at %v",
			agility.GetActiveAura(), char.GetStats()[stats.Agility], before)
	}
	if !flametongue.IsActive() {
		t.Error("after the totem, Flametongue Totem's trigger is still down")
	}
}

// The party's Windfury Totem holds over the shaman's own Grace of Air and
// Flametongue Totem: each stands and does nothing.
func TestGeneratedPartyWindfuryTotemHoldsOverOwnTotems(t *testing.T) {
	var graceOfAir, flametongue *core.Aura
	sim, char := newTotemSim(t, &proto.PartyBuffs{WindfuryTotem: true}, func(c *core.Character) {
		graceOfAir = buffs.NewSpellDataGraceOfAirTotemAura(&c.Unit, true, 0)
		flametongue = standWeaponTotem(c, "Flametongue Totem (Self)", buffs.FlametongueTotemCategory,
			buffs.FlametongueTotemPriority(), buffs.NewFlametongueTotemProcTrigger(c))
	})
	before := char.GetStats()[stats.Agility]

	graceOfAir.Activate(sim)
	flametongue.Activate(sim)
	if !graceOfAir.IsActive() || !flametongue.IsActive() {
		t.Fatalf("Grace of Air is up %v and Flametongue Totem %v, want both standing", graceOfAir.IsActive(), flametongue.IsActive())
	}
	if got := char.GetStats()[stats.Agility]; got != before {
		t.Errorf("the shaman's Grace of Air moved agility from %v to %v under the party's Windfury Totem", before, got)
	}
	if char.GetAura("Flametongue Totem Trigger").IsActive() {
		t.Error("the shaman's Flametongue Totem turned its trigger on under the party's Windfury Totem")
	}
}

// Windfury Weapon on the main hand keeps the shaman's own Windfury Totem's proc
// down, and the standing totem still holds over Flametongue Totem.
func TestGeneratedWindfuryTotemHoldsOverFlametongueTotemUnderWindfuryWeapon(t *testing.T) {
	var windfury, imbue *core.Aura
	sim, char := newTotemSim(t, &proto.PartyBuffs{FlametongueTotem: true}, func(c *core.Character) {
		windfury = standWindfuryTotem(c)
		imbue = c.RegisterAura(core.Aura{Label: "Windfury Weapon (Main Hand)", Duration: core.NeverExpires})
		imbue.NewExclusiveEffect(buffs.WindfuryTotemCategory, false, core.ExclusiveEffect{Priority: math.MaxFloat64})
	})

	imbue.Activate(sim)
	windfury.Activate(sim)
	if char.GetAura("Windfury Totem Trigger").IsActive() {
		t.Fatal("under Windfury Weapon the shaman's Windfury Totem turned its trigger on")
	}
	if char.GetAura("Flametongue Totem Trigger").IsActive() {
		t.Error("under Windfury Weapon and the shaman's Windfury Totem, Flametongue Totem's trigger is up")
	}
}

// A generated damage shield deals the game files' damage back to whoever lands a
// melee hit, and nothing to a spell.
func TestGeneratedThornsStrikesBackAtAMeleeHit(t *testing.T) {
	sim := setupFakeSimWithBuffs(&proto.RaidBuffs{}, &proto.PartyBuffs{}, &proto.IndividualBuffs{Thorns: true})
	char := sim.Raid.Parties[0].Players[0].GetCharacter()
	attacker := sim.Encounter.AllTargetUnits[0]

	aura := char.GetAura("Thorns (External)")
	if aura == nil {
		t.Fatalf("no aura is labelled %q; the unit has %v", "Thorns (External)", auraLabels(char))
	}
	if !aura.IsActive() {
		t.Fatal("the individual Thorns is not up at the start of the fight")
	}

	shield := char.GetSpell(core.ActionID{SpellID: 9910, Tag: 1})
	if shield == nil {
		t.Fatal("the shield registered no spell to deal its damage with")
	}

	// The shield's own swings would land on the character too, so they are
	// cancelled and every hit the test asks about is delivered by hand.
	attacker.AutoAttacks.CancelAutoSwing(sim)

	landed := &core.SpellResult{Target: &char.Unit, Outcome: core.OutcomeHit}
	aura.OnSpellHitTaken(aura, sim, attacker.AutoAttacks.MHAuto(), landed)
	sim.Step()
	if got := shield.SpellMetrics[attacker.UnitIndex].Casts; got != 1 {
		t.Errorf("a melee hit taken cast the shield %d times, want once", got)
	}
	if got, want := shield.SpellMetrics[attacker.UnitIndex].TotalDamage, rowAmount(t, 9910, dbcenums.A_DAMAGE_SHIELD); got != want {
		t.Errorf("the shield dealt %v damage, want the game files' %v", got, want)
	}

	caster := sim.Raid.Parties[0].Players[0].(*core.FakeAgent)
	aura.OnSpellHitTaken(aura, sim, caster.Spell, landed)
	sim.Step()
	if got := shield.SpellMetrics[attacker.UnitIndex].Casts; got != 1 {
		t.Errorf("a shadow spell taken cast the shield %d times, want the shield to ignore it", got)
	}
}

// Retribution Aura is the paladin's slot: the external copy keeps its own
// category and stays out of the shared one only the paladin's own cast joins.
func TestGeneratedRetributionAuraHoldsThePaladinSlot(t *testing.T) {
	char := core.NewGeneratedBuffTestCharacter()

	core.ApplyBuffEffects(char,
		&proto.RaidBuffs{}, &proto.PartyBuffs{RetributionAura: true}, &proto.IndividualBuffs{})

	external := char.GetAura("Retribution Aura (External)")
	if external == nil {
		t.Fatalf("no aura is labelled %q; the unit has %v", "Retribution Aura (External)", auraLabels(char))
	}
	category := char.ExclusiveEffectManager.GetExclusiveEffectCategory(buffs.RetributionAuraCategory)
	if bid := rowAmount(t, 10301, dbcenums.A_DAMAGE_SHIELD); !category.SingleAura || len(category.Effects()) != 1 || category.Effects()[0].Priority != bid {
		t.Errorf("the category is single-aura %v with %d effects, first bid %v; want one bidding %v",
			category.SingleAura, len(category.Effects()), category.Effects()[0].Priority, bid)
	}
	if shared := char.ExclusiveEffectManager.GetExclusiveEffectCategory(buffs.PaladinAuraCategory); len(shared.Effects()) != 0 {
		t.Errorf("the external copy put %d effects in the paladin's shared category, want none",
			len(shared.Effects()))
	}

	player := buffs.NewSpellDataRetributionAura(&char.Unit, true, 0)
	if shared := char.ExclusiveEffectManager.GetExclusiveEffectCategory(buffs.PaladinAuraCategory); len(shared.Effects()) != 1 {
		t.Errorf("the paladin's own copy put %d effects in the shared category, want one", len(shared.Effects()))
	}
	if player == external {
		t.Error("the paladin's own copy and the external one are the same aura")
	}
}

// The party's copy cannot see the providing paladin, so the spell power the party states is folded
// into the damage it bids and deals.
func TestRetributionAuraCarriesThePartysSpellPower(t *testing.T) {
	char := core.NewGeneratedBuffTestCharacter()

	core.ApplyBuffEffects(char,
		&proto.RaidBuffs{}, &proto.PartyBuffs{RetributionAura: true, RetributionAuraSpellPower: 450}, &proto.IndividualBuffs{})

	if char.GetAura("Retribution Aura (External)") == nil {
		t.Fatalf("no aura is labelled %q; the unit has %v", "Retribution Aura (External)", auraLabels(char))
	}

	want := rowAmount(t, 10301, dbcenums.A_DAMAGE_SHIELD) + buffs.RetributionAuraSpellPowerCoefficient*450
	category := char.ExclusiveEffectManager.GetExclusiveEffectCategory(buffs.RetributionAuraCategory)
	if len(category.Effects()) != 1 || category.Effects()[0].Priority != want {
		t.Errorf("the category has %d effects, first bid %v; want one bidding %v",
			len(category.Effects()), category.Effects()[0].Priority, want)
	}
}

// A whole environment, because an external cooldown registers a spell, a timer
// per source and a major cooldown, none of which a bare Character has.
func setupFakeSimWithBuffs(raid *proto.RaidBuffs, party *proto.PartyBuffs, individual *proto.IndividualBuffs) *core.Simulation {
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 100},
		Raid: &proto.Raid{
			Parties: []*proto.Party{
				{
					Players: []*proto.Player{
						{
							Name:      "Caster",
							Class:     proto.Class_ClassShaman,
							Buffs:     individual,
							Spec:      &proto.Player_ElementalShaman{},
							Equipment: &proto.EquipmentSpec{},
						},
					},
					Buffs: party,
				},
			},
			Buffs: raid,
		},
		Encounter: &proto.Encounter{
			Targets: []*proto.Target{{
				Name: "target", Level: 60, MobType: proto.MobType_MobTypeDemon,
				SwingSpeed: 2, MinBaseDamage: 100,
			}},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()

	return sim
}

// Two druids innervating one character take turns: the second waits for the
// first one's aura to fall off, and a third cast waits for the cooldown the
// game files state.
func TestGeneratedInnervatesTakeTurnsBetweenTheirSources(t *testing.T) {
	sim := setupFakeSimWithBuffs(&proto.RaidBuffs{}, &proto.PartyBuffs{}, &proto.IndividualBuffs{Innervates: 2})
	char := sim.Raid.Parties[0].Players[0].GetCharacter()

	aura := char.GetAura("Innervates (External)")
	if aura == nil {
		t.Fatalf("no aura is labelled %q; the unit has %v", "Innervates (External)", auraLabels(char))
	}
	if want := rowDuration(t, 29166); aura.Duration != want {
		t.Errorf("the innervate lasts %v, want the game files' %v", aura.Duration, want)
	}

	spell := char.GetSpell(core.ActionID{SpellID: 29166, Tag: -1})
	if spell == nil {
		t.Fatal("no spell stands for the external innervates")
	}

	// What the sim asks before it spends one of the sources: its own timer and
	// the condition the external approximation installs.
	ready := func() bool {
		return spell.CD.Timer.IsReady(sim) && spell.ExtraCastCondition(sim, &char.Unit)
	}

	if !ready() {
		t.Fatal("the first of two innervates cannot be cast at the start of the fight")
	}
	spell.SkipCastAndApplyEffects(sim, &char.Unit)
	if !aura.IsActive() {
		t.Fatal("casting the first innervate left the character without the aura")
	}
	if ready() {
		t.Error("a second innervate lands while the first one is still up")
	}

	sim.CurrentTime = aura.Duration + time.Second
	aura.Deactivate(sim)
	if !ready() {
		t.Fatal("the second druid cannot innervate once the first aura has fallen off")
	}
	spell.SkipCastAndApplyEffects(sim, &char.Unit)

	sim.CurrentTime += aura.Duration + time.Second
	aura.Deactivate(sim)
	if ready() {
		t.Error("a third innervate lands before either druid's cooldown is up")
	}

	sim.CurrentTime = buffs.InnervatesCooldown() + time.Minute
	if !ready() {
		t.Error("the first druid cannot innervate again once its cooldown is up")
	}
}

// Power Infusion raises damage and healing done while it is up, and nothing
// once it has expired. The damage half is the six schools the game files' mask
// names, which does not include physical.
func TestGeneratedPowerInfusionRaisesDamageAndHealingDone(t *testing.T) {
	sim := setupFakeSimWithBuffs(&proto.RaidBuffs{}, &proto.PartyBuffs{}, &proto.IndividualBuffs{PowerInfusions: 1})
	char := sim.Raid.Parties[0].Players[0].GetCharacter()

	aura := char.GetAura("Power Infusions (External)")
	if aura == nil {
		t.Fatalf("no aura is labelled %q; the unit has %v", "Power Infusions (External)", auraLabels(char))
	}
	if want := rowDuration(t, 10060); aura.Duration != want {
		t.Errorf("the infusion lasts %v, want the game files' %v", aura.Duration, want)
	}

	damageDone := core.MultiplierFromPercent(rowAmount(t, 10060, dbcenums.A_MOD_DAMAGE_PERCENT_DONE))
	healingDone := core.MultiplierFromPercent(rowAmount(t, 10060, dbcenums.A_MOD_HEALING_DONE_PERCENT))
	healing := char.PseudoStats.HealingDealtMultiplier
	physical := stats.SchoolIndexPhysical
	before := char.PseudoStats.SchoolDamageDealtMultiplier[physical]

	aura.Activate(sim)
	// Spell 10060 states mask 126, which is every school but physical, so a
	// melee swing is not part of what the infusion raises.
	for _, school := range magicSchools {
		if got := char.PseudoStats.SchoolDamageDealtMultiplier[school]; got != damageDone {
			t.Errorf("school %d deals %v times the damage, want the game files' %v", school, got, damageDone)
		}
	}
	if got := char.PseudoStats.SchoolDamageDealtMultiplier[physical]; got != before {
		t.Errorf("physical damage dealt is %v, want the mask to have left it at %v", got, before)
	}
	if got := char.PseudoStats.HealingDealtMultiplier; got != healing*healingDone {
		t.Errorf("the infusion multiplies healing dealt by %v, want the game files' %v", got/healing, healingDone)
	}

	aura.Deactivate(sim)
	for _, school := range magicSchools {
		if got := char.PseudoStats.SchoolDamageDealtMultiplier[school]; got != 1 {
			t.Errorf("school %d still deals %v times the damage once the infusion expired", school, got)
		}
	}
	if char.PseudoStats.HealingDealtMultiplier != healing {
		t.Error("the infusion left part of itself behind when it expired")
	}
}

// Mana Tide's aura states mana per tick, which reaches the sim as the mana per
// 5 seconds it is worth; the totem cast states how long it stands.
func TestGeneratedManaTideTotemRestoresItsMana(t *testing.T) {
	sim := setupFakeSimWithBuffs(&proto.RaidBuffs{}, &proto.PartyBuffs{ManaTideTotems: 1}, &proto.IndividualBuffs{})
	char := sim.Raid.Parties[0].Players[0].GetCharacter()

	aura := char.GetAura("Mana Tide Totem (External)")
	if aura == nil {
		t.Fatalf("no aura is labelled %q; the unit has %v", "Mana Tide Totem (External)", auraLabels(char))
	}
	if want := rowDuration(t, 17359); aura.Duration != want {
		t.Errorf("the totem stands for %v, want the cast's %v", aura.Duration, want)
	}
	tick := rowEffect(t, 17360, dbcenums.A_PERIODIC_ENERGIZE)
	mp5 := rowAmount(t, 17360, dbcenums.A_PERIODIC_ENERGIZE) * 5000 / float64(tick.PeriodMs)

	before := char.GetStats()[stats.MP5]
	aura.Activate(sim)
	if got := char.GetStats()[stats.MP5] - before; got != mp5 {
		t.Errorf("the totem applied %v MP5, want the %v per %v the game files state as %v",
			got, tick.Average(), tick.Period(), mp5)
	}
}

// A buff that is not simply up is not part of the stats the character sheet is
// measured with: the cooldowns and the weapon totems. A main hand carries one
// totem enchant, so Flametongue Totem is measured on a party without Windfury.
func TestGeneratedDrivenBuffsAreNotBuildPhaseAuras(t *testing.T) {
	for _, tc := range []struct {
		party      *proto.PartyBuffs
		individual *proto.IndividualBuffs
		labels     []string
	}{
		{
			party:      &proto.PartyBuffs{ManaTideTotems: 1, WindfuryTotem: true},
			individual: &proto.IndividualBuffs{Innervates: 1, PowerInfusions: 1},
			labels: []string{
				"Innervates (External)", "Power Infusions (External)", "Mana Tide Totem (External)",
				"Windfury Totem (External)",
			},
		},
		{
			party:      &proto.PartyBuffs{FlametongueTotem: true},
			individual: &proto.IndividualBuffs{},
			labels:     []string{"Flametongue Totem (External)"},
		},
	} {
		sim := setupFakeSimWithBuffs(&proto.RaidBuffs{}, tc.party, tc.individual)
		char := sim.Raid.Parties[0].Players[0].GetCharacter()

		for _, label := range tc.labels {
			aura := char.GetAura(label)
			if aura == nil {
				t.Fatalf("no aura is labelled %q; the unit has %v", label, auraLabels(char))
			}
			if aura.BuildPhase != core.CharacterBuildPhaseNone {
				t.Errorf("%s is measured in build phase %v, want none of them", label, aura.BuildPhase)
			}
		}
	}
}

// Stoneskin takes a flat amount off every physical hit, and Guardian Totems raises it.
// The reduction lands on PseudoStats.BonusPhysicalDamageTaken, not a stats.Stats entry.
func TestGeneratedStoneskinTotemTakesTheTalentedValue(t *testing.T) {
	untalented := rowAmount(t, 10405, dbcenums.A_MOD_DAMAGE_TAKEN)
	if got := buffs.StoneskinTotemValue(0); got != untalented {
		t.Errorf("untalented Stoneskin is worth %v, want the game files' %v", got, untalented)
	}
	talented := buffs.StoneskinTotemValue(2)
	if untalented >= 0 || talented >= untalented {
		t.Errorf("Stoneskin is worth %v untalented and %v with 2/2 Guardian Totems, want a reduction the talent makes larger",
			untalented, talented)
	}

	char := core.NewGeneratedBuffTestCharacter()
	core.ApplyBuffEffects(char, &proto.RaidBuffs{},
		&proto.PartyBuffs{StoneskinTotem: proto.TristateEffect_TristateEffectImproved}, &proto.IndividualBuffs{})
	char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)

	aura := char.GetAura("Stoneskin Totem (External)")
	if aura == nil {
		t.Fatalf("no aura is labelled %q; the unit has %v", "Stoneskin Totem (External)", auraLabels(char))
	}
	if got := char.PseudoStats.BonusPhysicalDamageTaken; got != talented {
		t.Errorf("the talented totem applied %v physical damage taken, want %v", got, talented)
	}
}

// Furious Howl does not stack with Battle Shout: the two share one category, and the stronger shout
// is the one that applies. The howl's amount gains per level past its BaseLevel, floored.
func TestGeneratedFuriousHowlDoesNotStackWithBattleShout(t *testing.T) {
	howl := levelScaledFloor(rowEffect(t, 24597, dbcenums.A_MOD_ATTACK_POWER))
	if howl == 0 {
		t.Fatal("Furious Howl states no attack power")
	}
	if got := buffs.FuriousHowlValue(0); got != howl {
		t.Fatalf("Furious Howl is worth %v attack power, want the game files' amount at the character's level, floored to %v",
			got, howl)
	}

	howlOnly := core.NewGeneratedBuffTestCharacter()
	core.ApplyBuffEffects(howlOnly, &proto.RaidBuffs{}, &proto.PartyBuffs{FuriousHowl: true}, &proto.IndividualBuffs{})
	howlOnly.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)
	if got := howlOnly.GetStats()[stats.AttackPower]; got != howl {
		t.Errorf("Furious Howl alone applied %v attack power, want %v", got, howl)
	}

	char := core.NewGeneratedBuffTestCharacter()
	core.ApplyBuffEffects(char, &proto.RaidBuffs{},
		&proto.PartyBuffs{BattleShout: proto.TristateEffect_TristateEffectRegular, FuriousHowl: true}, &proto.IndividualBuffs{})
	char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)
	if got, want := char.GetStats()[stats.AttackPower], max(battleShout(t), howl); got != want {
		t.Errorf("attack power is %v under both shouts, want the stronger one's %v alone", got, want)
	}
}

// An owner with one true pet and one guardian, both meleeing and summoned at the pull, and a dynamic
// pet that stays away until a test enables it.
type petOwner struct {
	core.Character
	pet      *testPet
	guardian *testPet
	dynamic  *testPet
}

type testPet struct {
	core.Pet
}

func init() {
	core.RegisterAgentFactory(
		proto.Player_Hunter{},
		proto.Spec_SpecHunter,
		func(character *core.Character, _ *proto.Player, _ *proto.Raid) core.Agent {
			owner := &petOwner{Character: *character}
			owner.pet = newTestPet(owner, "True Pet", false)
			owner.guardian = newTestPet(owner, "Guardian", true)
			owner.dynamic = &testPet{Pet: core.NewPet(core.PetConfig{
				Name:  "Dynamic Pet",
				Owner: &owner.Character,
				StatInheritance: func(ownerStats stats.Stats) stats.Stats {
					return stats.Stats{stats.AttackPower: max(ownerStats[stats.Strength], ownerStats[stats.Agility])}
				},
				IsDynamic: true,
			})}
			owner.AddPet(owner.dynamic)
			return owner
		},
		func(player *proto.Player, spec interface{}) {
			player.Spec = spec.(*proto.Player_Hunter)
		},
	)
}

func newTestPet(owner *petOwner, name string, guardian bool) *testPet {
	pet := &testPet{Pet: core.NewPet(core.PetConfig{
		Name:            name,
		Owner:           &owner.Character,
		StatInheritance: func(stats.Stats) stats.Stats { return stats.Stats{} },
		EnabledOnStart:  true,
		IsGuardian:      guardian,
	})}
	pet.EnableAutoAttacks(pet, core.AutoAttackOptions{
		MainHand:       core.Weapon{BaseDamageMin: 10, BaseDamageMax: 20, SwingSpeed: 2, NormalizedSwingSpeed: 2},
		AutoSwingMelee: true,
	})
	owner.AddPet(pet)
	return pet
}

func (c *petOwner) GetCharacter() *core.Character       { return &c.Character }
func (c *petOwner) ApplyTalents()                       {}
func (c *petOwner) Reset(_ *core.Simulation)            {}
func (c *petOwner) OnGCDReady(_ *core.Simulation)       {}
func (c *petOwner) OnEncounterStart(_ *core.Simulation) {}
func (c *petOwner) Initialize()                         {}

func (p *testPet) GetPet() *core.Pet                   { return &p.Pet }
func (p *testPet) Reset(_ *core.Simulation)            {}
func (p *testPet) OnEncounterStart(_ *core.Simulation) {}

func (p *testPet) ExecuteCustomRotation(_ *core.Simulation) {}

func setupPetOwnerSim(party *proto.PartyBuffs) (*core.Simulation, *petOwner) {
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 100},
		Raid: &proto.Raid{
			Parties: []*proto.Party{{
				Players: []*proto.Player{{
					Name:      "Owner",
					Class:     proto.Class_ClassHunter,
					Spec:      &proto.Player_Hunter{},
					Equipment: &proto.EquipmentSpec{},
				}},
				Buffs: party,
			}},
		},
		Encounter: &proto.Encounter{
			Targets:  []*proto.Target{{Name: "target", Level: 60}},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*petOwner)
}

// A true pet stands under its owner's totems with the owner's exclusivity, except those whose aura
// only a player can carry: it has Mana Spring and, under Windfury Totem, the totem's proc, but never
// Strength of Earth or Grace of Air. A guardian stands under none of them.
func TestTruePetsStandUnderTheOwnersTotems(t *testing.T) {
	for _, tc := range []struct {
		name     string
		party    *proto.PartyBuffs
		windfury bool
	}{
		{"grace of air", &proto.PartyBuffs{ManaSpringTotem: proto.TristateEffect_TristateEffectRegular, StrengthOfEarthTotem: true, GraceOfAirTotem: true}, false},
		{"windfury", &proto.PartyBuffs{ManaSpringTotem: proto.TristateEffect_TristateEffectRegular, StrengthOfEarthTotem: true, GraceOfAirTotem: true, WindfuryTotem: true}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, owner := setupPetOwnerSim(tc.party)
			pet := owner.pet

			if aura := owner.GetAura("Strength of Earth Totem (External)"); aura == nil || !aura.IsActive() {
				t.Fatalf("the owner's Strength of Earth Totem is %v; the owner has %v", aura, auraLabels(&owner.Character))
			}
			if aura := pet.GetAura("Mana Spring Totem (External)"); aura == nil || !aura.IsActive() {
				t.Fatalf("the pet's Mana Spring Totem is %v; the pet has %v", aura, auraLabels(&pet.Character))
			}
			if got, want := pet.GetStats()[stats.MP5], buffs.ManaSpringTotemValue(0); got != want {
				t.Errorf("the pet has %v MP5, want the totem's %v", got, want)
			}
			for _, label := range []string{"Strength of Earth Totem (External)", "Grace of Air Totem (External)"} {
				if pet.GetAura(label) != nil {
					t.Errorf("the pet carries %q", label)
				}
			}
			if got := pet.GetStats(); got[stats.Strength] != 0 || got[stats.Agility] != 0 {
				t.Errorf("the pet has %v strength and %v agility, want none", got[stats.Strength], got[stats.Agility])
			}
			trigger := pet.GetAura("Windfury Totem Trigger")
			if got := trigger != nil && trigger.IsActive(); got != tc.windfury {
				t.Errorf("the pet's Windfury Totem proc is up %v, want %v", got, tc.windfury)
			}

			for _, label := range []string{"Mana Spring Totem (External)", "Windfury Totem"} {
				if owner.guardian.GetAura(label) != nil {
					t.Errorf("the guardian carries %q", label)
				}
			}
			if got := owner.guardian.GetStats()[stats.MP5]; got != 0 {
				t.Errorf("the guardian has %v MP5, want none", got)
			}
		})
	}
}

// A pet dismissed and summoned again stands under the totem again.
func TestTruePetsStandUnderTheOwnersTotemsAfterASummon(t *testing.T) {
	sim, owner := setupPetOwnerSim(&proto.PartyBuffs{ManaSpringTotem: proto.TristateEffect_TristateEffectRegular})
	pet := owner.pet
	totem := pet.GetAura("Mana Spring Totem (External)")
	if totem == nil {
		t.Fatalf("the pet has %v, want Mana Spring Totem", auraLabels(&pet.Character))
	}

	pet.Disable(sim)
	if totem.IsActive() || pet.GetStats()[stats.MP5] != 0 {
		t.Errorf("the dismissed pet stands under the totem %v with %v MP5, want neither",
			totem.IsActive(), pet.GetStats()[stats.MP5])
	}
	pet.Enable(sim, pet)
	if !totem.IsActive() {
		t.Fatal("the summoned pet does not stand under the totem")
	}
	if got, want := pet.GetStats()[stats.MP5], buffs.ManaSpringTotemValue(0); got != want {
		t.Errorf("the summoned pet has %v MP5, want the totem's %v", got, want)
	}
}

// Mana Tide is up on a true pet for as long as it is up on the owner.
func TestTruePetsStandUnderTheOwnersManaTide(t *testing.T) {
	sim, owner := setupPetOwnerSim(&proto.PartyBuffs{ManaTideTotems: 1})
	totem := owner.GetAura("Mana Tide Totem (External)")
	petTotem := owner.pet.GetAura("Mana Tide Totem (External)")
	if totem == nil || petTotem == nil {
		t.Fatalf("the owner's Mana Tide is %v and the pet's %v, want both", totem, petTotem)
	}
	if petTotem.IsActive() {
		t.Error("the pet stands under Mana Tide before the owner does")
	}

	totem.Activate(sim)
	if !petTotem.IsActive() {
		t.Error("the owner's Mana Tide left the pet without it")
	}
	if owner.guardian.GetAura("Mana Tide Totem (External)") != nil {
		t.Error("the guardian carries Mana Tide")
	}
	totem.Deactivate(sim)
	if petTotem.IsActive() {
		t.Error("the pet keeps Mana Tide after the owner's has gone")
	}
}

func auraLabels(char *core.Character) []string {
	labels := make([]string, 0, len(char.GetAuras()))
	for _, aura := range char.GetAuras() {
		labels = append(labels, aura.Label)
	}
	return labels
}

package core_test

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

func TestPartyBattleShoutAppliesTheGeneratedAura(t *testing.T) {
	char := core.NewGeneratedBuffTestCharacter()

	core.ApplyBuffEffects(char,
		&proto.RaidBuffs{}, &proto.PartyBuffs{BattleShout: proto.TristateEffect_TristateEffectRegular}, &proto.IndividualBuffs{})
	char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)

	if got := char.GetStats()[stats.AttackPower]; got != 139 {
		t.Errorf("the party's Battle Shout applied %v attack power, want the game files' 139", got)
	}

	aura := char.GetAura("Battle Shout (External)")
	if aura == nil {
		t.Fatalf("no aura is labelled %q; the unit has %v", "Battle Shout (External)", auraLabels(char))
	}
	if want := (core.ActionID{SpellID: 25289, Tag: -1}); aura.ActionID != want {
		t.Errorf("the external copy is %v, want %v", aura.ActionID, want)
	}
	if buffs.BattleShoutDuration(0) != 3*time.Minute {
		t.Errorf("Battle Shout lasts %v, want the game files' 3 minutes", buffs.BattleShoutDuration(0))
	}
	if aura.Duration != buffs.BattleShoutDuration(0) {
		t.Errorf("the external copy lasts %v, want %v", aura.Duration, buffs.BattleShoutDuration(0))
	}

	category := char.ExclusiveEffectManager.GetExclusiveEffectCategory(buffs.BattleShoutCategory)
	if !category.SingleAura {
		t.Error("the Battle Shout category is not single-aura, so a second copy could sit next to it")
	}
	if len(category.Effects()) != 1 {
		t.Errorf("the category holds %d effects, want one", len(category.Effects()))
	} else if category.Effects()[0].Priority != 139 {
		t.Errorf("the only effect in the category bids %v, want 139", category.Effects()[0].Priority)
	}

	// The chain the driver installs finds the player's own shout by this tag.
	if tagged := char.GetAurasWithTag(buffs.BattleShoutCategory); len(tagged) != 1 || tagged[0] != aura {
		t.Errorf("%d auras carry the Battle Shout tag, want only the external copy", len(tagged))
	}
}

// A warrior who casts Battle Shout and has the external one ticked: both copies
// are 139 attack power, and both are equally long, so the tie goes to whichever
// the build phase activates last. ExclusiveEffect.Activate only turns a newcomer
// away when the incumbent's remaining duration is longer than the newcomer's, so
// the player's copy takes the category and the external one is deactivated - the
// character sheet shows 139 either way.
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
	if got := char.GetStats()[stats.AttackPower]; got != 139 {
		t.Errorf("both copies together applied %v attack power, want 139", got)
	}
	if tagged := char.GetAurasWithTag(buffs.BattleShoutCategory); len(tagged) != 2 {
		t.Errorf("%d auras carry the Battle Shout tag, want both copies", len(tagged))
	}
}

// The improved state says the warrior who shouted for the party wears three
// pieces of Battlegear of Wrath, so the external copy is worth the game files' 139
// plus the set's 30.
func TestImprovedPartyBattleShoutAddsTheTierTwoBonus(t *testing.T) {
	for _, row := range []struct {
		name  string
		party *proto.PartyBuffs
		want  float64
	}{
		{"with the set", &proto.PartyBuffs{BattleShout: proto.TristateEffect_TristateEffectImproved}, 169},
		{"without it", &proto.PartyBuffs{BattleShout: proto.TristateEffect_TristateEffectRegular}, 139},
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

// Two copies of the shout that are no longer worth the same: the stronger one
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
			if got := char.GetStats()[stats.AttackPower]; got != 169 {
				t.Errorf("both copies together applied %v attack power, want the stronger one's 169", got)
			}
		})
	}
}

// A flat stat row: no category at all, so the two sources of stamina add up.
func TestGeneratedFlatStatBuffsAddUp(t *testing.T) {
	char := core.NewGeneratedBuffTestCharacter()

	core.ApplyBuffEffects(char,
		&proto.RaidBuffs{PrayerOfFortitude: true}, &proto.PartyBuffs{BloodPact: true}, &proto.IndividualBuffs{})
	char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)

	if got := char.GetStats()[stats.Stamina]; got != 124 {
		t.Errorf("Prayer of Fortitude and Blood Pact applied %v stamina, want the game files' 70 + 54", got)
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
// goes through a multiplying dependency rather than a flat amount.
func TestGeneratedGreaterBlessingOfKingsMultipliesEveryStat(t *testing.T) {
	char := core.NewGeneratedBuffTestCharacter()
	char.SetStats(stats.Stats{
		stats.Strength: 100, stats.Agility: 100, stats.Stamina: 100,
		stats.Intellect: 100, stats.Spirit: 100, stats.Armor: 100,
	})

	core.ApplyBuffEffects(char,
		&proto.RaidBuffs{}, &proto.PartyBuffs{}, &proto.IndividualBuffs{GreaterBlessingOfKings: true})
	measureGeneratedBuffStats(char)

	for _, stat := range []stats.Stat{stats.Strength, stats.Agility, stats.Stamina, stats.Intellect, stats.Spirit} {
		if got := char.GetStats()[stat]; got != 110 {
			t.Errorf("%s is %v, want 100 multiplied by the game files' 1.1", stat.StatName(), got)
		}
	}
	if got := char.GetStats()[stats.Armor]; got != 100 {
		t.Errorf("armor is %v, want the blessing to have left it alone", got)
	}
}

// Two sources of shadow resistance compete for the school, while everything
// else Gift of the Wild grants is applied outright.
func TestGeneratedResistancesCompeteAcrossBuffs(t *testing.T) {
	char := core.NewGeneratedBuffTestCharacter()

	core.ApplyBuffEffects(char,
		&proto.RaidBuffs{GiftOfTheWild: true, PrayerOfShadowProtection: true},
		&proto.PartyBuffs{}, &proto.IndividualBuffs{})
	char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)

	if got := char.GetStats()[stats.ShadowResistance]; got != 60 {
		t.Errorf("shadow resistance is %v, want only Prayer of Shadow Protection's 60", got)
	}
	if got := char.GetStats()[stats.FrostResistance]; got != 27 {
		t.Errorf("frost resistance is %v, want Gift of the Wild's 27", got)
	}
	if got := char.GetStats()[stats.Armor]; got != 385 {
		t.Errorf("armor is %v, want Gift of the Wild's 385", got)
	}
	if got := char.GetStats()[stats.Stamina]; got != 16 {
		t.Errorf("stamina is %v, want Gift of the Wild's 16", got)
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

	// Both copies bid the same 735 and both are permanent, so the paladin's own
	// cast takes the slot by activating after the external one, and the armor on
	// the character sheet does not move.
	external := char.GetAura("Devotion Aura (External)")
	if !player.IsActive() {
		t.Error("the paladin's own aura did not activate next to the external one")
	}
	if external.IsActive() {
		t.Error("the external copy stayed active, so both copies of the aura are on the character")
	}
	if got := char.GetStats()[stats.Armor]; got != 735 {
		t.Errorf("both copies together applied %v armor, want the game files' 735", got)
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

	if got := char.GetStats()[stats.MP5]; got != 30 {
		t.Errorf("the improved totem applied %v MP5, want the curve's top value 30", got)
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
	if len(category.Effects()) != 1 || category.Effects()[0].Priority != 30 {
		t.Errorf("the totem's category holds %d effects, first bid %v; want one bidding 30",
			len(category.Effects()), category.Effects()[0].Priority)
	}
}

// The two rows whose stats the manifest names. Each is worth the game files' 3 on
// every kind of crit, and neither touches RangedCritPercent, which the sim adds
// on top of the physical one, so a ranged attack gains 3 and not 6.
//
// Spell 17007 calls Leader of the Pack exclusive with Moonkin Aura, so the two
// share a category that holds one aura: a party with both ticked is worth 3,
// and the copy that loses stays registered without applying anything.
func TestGeneratedCritAurasApplyTheClientsThreeToEveryKindOfCrit(t *testing.T) {
	for _, row := range []struct {
		name  string
		party *proto.PartyBuffs
	}{
		{"Leader of the Pack", &proto.PartyBuffs{LeaderOfThePack: true}},
		{"Moonkin Aura", &proto.PartyBuffs{MoonkinAura: true}},
		{"both", &proto.PartyBuffs{LeaderOfThePack: true, MoonkinAura: true}},
	} {
		t.Run(row.name, func(t *testing.T) {
			char := core.NewGeneratedBuffTestCharacter()

			core.ApplyBuffEffects(char, &proto.RaidBuffs{}, row.party, &proto.IndividualBuffs{})
			char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)

			if got := char.GetStats()[stats.PhysicalCritPercent]; got != 3 {
				t.Errorf("%s applied %v physical crit, want the game files' 3", row.name, got)
			}
			if got := char.GetStats()[stats.SpellCritPercent]; got != 3 {
				t.Errorf("%s applied %v spell crit, want the game files' 3", row.name, got)
			}
			if got := char.GetStats()[stats.RangedCritPercent]; got != 0 {
				t.Errorf("%s applied %v on top of the physical crit a ranged attack already reads", row.name, got)
			}
		})
	}
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
	if active := category.GetActiveEffect(); active == nil || active.Priority != 3 {
		t.Errorf("the category is held at %v, want the game files' 3", active)
	}
}

// The row whose only amount is a pseudo-stat the game files state as a reduction.
func TestGeneratedPartyConcentrationAuraReducesPushback(t *testing.T) {
	char := core.NewGeneratedBuffTestCharacter()

	core.ApplyBuffEffects(char, &proto.RaidBuffs{},
		&proto.PartyBuffs{ConcentrationAura: true}, &proto.IndividualBuffs{})
	char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)

	if got := char.PseudoStats.PushbackChance; got != 0.65 {
		t.Errorf("Concentration Aura left the pushback chance at %v, want 1 reduced by the game files' 35%%", got)
	}
}

// Rank 5 of Trueshot Aura, which the resolver takes, states one
// A_MOD_RANGED_ATTACK_POWER effect worth 50: melee attack power is untouched and
// the 75 of rank 4 is not what the party gets.
func TestGeneratedTrueshotAuraGivesTheTopRanksRangedAttackPower(t *testing.T) {
	char := core.NewGeneratedBuffTestCharacter()

	core.ApplyBuffEffects(char,
		&proto.RaidBuffs{}, &proto.PartyBuffs{TrueshotAura: true}, &proto.IndividualBuffs{})
	char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)

	if got := char.GetStats()[stats.RangedAttackPower]; got != 50 {
		t.Errorf("Trueshot Aura applied %v ranged attack power, want the game files' 50", got)
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

	if got := standing.GetStats()[stats.Agility]; got != 89 {
		t.Errorf("the totem applied %v agility, want the game files' 89", got)
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

// Each staff in the party is worth its own copy of the aura's amounts: 11 MP5
// per druid staff, 2 spell crit per mage one, 62 healing per priest one and
// 33 spell damage plus 33 healing per warlock one.
func TestGeneratedAtieshStavesCountTheStaves(t *testing.T) {
	char := core.NewGeneratedBuffTestCharacter()

	core.ApplyBuffEffects(char, &proto.RaidBuffs{},
		&proto.PartyBuffs{AtieshDruid: 1, AtieshMage: 3, AtieshPriest: 2, AtieshWarlock: 1},
		&proto.IndividualBuffs{})
	char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)

	for stat, want := range map[stats.Stat]float64{
		stats.MP5:              11,
		stats.SpellCritPercent: 6,
		stats.HealingPower:     62*2 + 33,
		stats.SpellDamage:      33,
	} {
		if got := char.GetStats()[stat]; got != want {
			t.Errorf("the party's staves applied %v %s, want %v", got, stat.StatName(), want)
		}
	}

	if char.GetAura("Atiesh - Mage (External)") == nil {
		t.Fatalf("the unit has %v, want an aura for the mage's staff", auraLabels(char))
	}
}

// The windfury proc grants the attack power the game files state for the second
// the aura lasts; the totem aura around it holds the category and the trigger.
func TestGeneratedWindfuryTotemProcAppliesTheClientsAttackPower(t *testing.T) {
	sim := setupFakeSimWithBuffs(&proto.RaidBuffs{}, &proto.PartyBuffs{WindfuryTotem: true}, &proto.IndividualBuffs{})
	char := sim.Raid.Parties[0].Players[0].GetCharacter()

	proc := char.GetAura("Windfury Totem (External)")
	if proc == nil {
		t.Fatalf("no aura is labelled %q; the unit has %v", "Windfury Totem (External)", auraLabels(char))
	}
	if proc.Duration != time.Second {
		t.Errorf("the proc buff lasts %v, want the game files' 1 second", proc.Duration)
	}

	before := char.GetStats()[stats.AttackPower]
	proc.Activate(sim)
	if got := char.GetStats()[stats.AttackPower] - before; got != 246 {
		t.Errorf("the proc applied %v attack power, want the game files' 246", got)
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
	if len(category.Effects()) != 1 || category.Effects()[0].Priority != buffs.WindfuryTotemValue(0) {
		t.Errorf("the category holds %d effects, first bid %v; want one bidding 246",
			len(category.Effects()), category.Effects()[0].Priority)
	}
	if char.GetAura("Windfury Totem Trigger") == nil {
		t.Error("the driver registered no proc trigger for the totem")
	}
}

// The proc holds the game files' 2 charges and each landed auto attack, in either
// hand, spends one. A melee special or a missed swing spends none, and a proc no
// auto lands on keeps its attack power for the whole second.
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

	if proc.MaxStacks != 2 {
		t.Fatalf("the proc holds %d charges, want the game files' 2", proc.MaxStacks)
	}
	fullCharges := func() {
		proc.Activate(sim)
		proc.SetStacks(sim, proc.MaxStacks)
	}

	before := char.GetStats()[stats.AttackPower]
	fullCharges()

	swing(special, landed)
	swing(mhAuto, missed)
	if got := proc.GetStacks(); got != 2 {
		t.Errorf("a special and a missed auto left %d charges, want both still there", got)
	}

	swing(mhAuto, landed)
	if !proc.IsActive() || proc.GetStacks() != 1 {
		t.Fatalf("one landed auto left the proc active %v with %d charges, want it up with 1",
			proc.IsActive(), proc.GetStacks())
	}
	if got := char.GetStats()[stats.AttackPower] - before; got != 246 {
		t.Errorf("the proc holds %v attack power on its last charge, want the game files' 246", got)
	}

	swing(ohAuto, landed)
	if proc.IsActive() {
		t.Error("the proc is still up after its second landed auto, want the charges spent")
	}
	if got := char.GetStats()[stats.AttackPower]; got != before {
		t.Errorf("the spent proc left %v attack power behind", got-before)
	}

	fullCharges()
	start := sim.CurrentTime
	swing(special, landed)
	if !proc.IsActive() || proc.RemainingDuration(sim) != time.Second {
		t.Fatalf("a proc no auto landed on is active %v with %v left, want the full second",
			proc.IsActive(), proc.RemainingDuration(sim))
	}
	sim.AddPendingAction(core.NewDelayedAction(core.DelayedActionOptions{
		DoAt:     start + time.Second*2,
		OnAction: func(*core.Simulation) {},
	}))
	for proc.IsActive() && sim.CurrentTime < start+time.Second*2 {
		sim.Step()
	}
	if proc.IsActive() || proc.GetStacks() != 0 {
		t.Errorf("the lingering proc is active %v with %d charges at %v, want it gone after its second",
			proc.IsActive(), proc.GetStacks(), sim.CurrentTime-start)
	}
}

// Windfury Totem's proc is the party aura 10612: a main-hand hit that deals
// damage can fire it, auto or special, and no other hit can, a proc's hit
// included. A main-hand auto that procs it has spent the first charge itself
// and leaves one for the extra attack; a special leaves both.
func TestGeneratedWindfuryTotemProcsOffMainHandHitsWithTheChargesTheTriggerLeaves(t *testing.T) {
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
			trigger.OnSpellHitDealt(trigger, sim, spell, &core.SpellResult{Target: target, Outcome: core.OutcomeHit, Damage: damage})
		}
		return len(*started) > 0
	}

	for _, row := range []struct {
		name    string
		spell   *core.Spell
		charges int32
	}{
		{"a main-hand auto leaves one charge", &core.Spell{ProcMask: core.ProcMaskMeleeMHAuto}, 1},
		{"a main-hand special leaves both charges", &core.Spell{ProcMask: core.ProcMaskMeleeMHSpecial}, 2},
	} {
		t.Run(row.name, func(t *testing.T) {
			sim, target, trigger, started := setup(t)
			if !offer(sim, target, trigger, started, row.spell, 100) {
				t.Fatal("200 landed hits never procced the totem")
			}
			if got := (*started)[0]; got != row.charges {
				t.Errorf("the proc started with %d charges, want %d", got, row.charges)
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

	t.Run("the trigger waits 100 ms between procs", func(t *testing.T) {
		_, _, trigger, _ := setup(t)
		if trigger.Icd == nil || trigger.Icd.Duration != 100*time.Millisecond {
			t.Errorf("the trigger's cooldown is %v, want the game files' 100 ms", trigger.Icd)
		}
	})
}

// Every main-hand hit adds 1363 x weapon speed / 100 fire damage.
func TestGeneratedFlametongueTotemDamageFollowsWeaponSpeed(t *testing.T) {
	if got := buffs.FlametongueTotemDamage(2.6); math.Abs(got-35.438) > 0.001 {
		t.Errorf("a 2.6 speed weapon procs %v, want 1363 x 2.6 / 100", got)
	}
}

// A main-hand auto attack that lands strikes the target for the game files'
// 1363 x the weapon's speed / 100 in fire damage on the spell hit and crit
// tables; a special, an off-hand hit or a missed swing strikes nothing.
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

	strike := char.GetSpell(core.ActionID{SpellID: 16389, Tag: -1})
	if strike == nil {
		t.Fatal("the totem registered no spell for its proc")
	}
	if strike.SpellSchool != core.SpellSchoolFire || strike.DefenseType != core.DefenseTypeMagic {
		t.Errorf("the proc is %v on %v, want fire on the spell tables", strike.SpellSchool, strike.DefenseType)
	}
	if !strike.Flags.Matches(core.SpellFlagProc) {
		t.Error("the proc's hits are not flagged as a proc, which 16389 does not opt out of")
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
		want := 1363 * row.speed / 100
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

	if char.GetAura("Flametongue Totem Trigger") != nil || char.GetSpell(core.ActionID{SpellID: 16389, Tag: -1}) != nil {
		t.Error("Flametongue Totem registered its proc next to Windfury Totem")
	}
	if len(char.ExclusiveEffectManager.GetExclusiveEffectCategory(buffs.FlametongueTotemCategory).Effects()) != 0 {
		t.Error("Flametongue Totem bids in its category next to Windfury Totem")
	}

	trigger := char.GetAura("Windfury Totem Trigger")
	proc := char.GetAura("Windfury Totem (External)")
	if trigger == nil || !trigger.IsActive() || proc == nil {
		t.Fatalf("Windfury Totem's trigger %v, proc %v; want both registered and the trigger up", trigger, proc)
	}
	target := sim.Encounter.AllTargetUnits[0]
	for i := 0; i < 200 && !proc.IsActive(); i++ {
		trigger.OnSpellHitDealt(trigger, sim, &core.Spell{ProcMask: core.ProcMaskMeleeMHSpecial},
			&core.SpellResult{Target: target, Outcome: core.OutcomeHit, Damage: 100})
	}
	if !proc.IsActive() {
		t.Error("200 landed main-hand hits never procced Windfury Totem")
	}
}

// A generated damage shield deals the game files' damage back to whoever lands a
// melee hit, and nothing to a spell.
func TestGeneratedThornsStrikesBackAtAMeleeHit(t *testing.T) {
	sim := setupFakeSimWithBuffs(&proto.RaidBuffs{}, &proto.PartyBuffs{}, &proto.IndividualBuffs{Thorns: true})
	char := sim.Raid.Parties[0].Players[0].GetCharacter()
	attacker := sim.Encounter.AllTargetUnits[0]

	if buffs.ThornsValue(0) != 22 {
		t.Errorf("Thorns is worth %v, want the game files' 22", buffs.ThornsValue(0))
	}

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
	if got := shield.SpellMetrics[attacker.UnitIndex].TotalDamage; got != 22 {
		t.Errorf("the shield dealt %v damage, want the game files' 22", got)
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
	if buffs.RetributionAuraValue(0) != 30 {
		t.Errorf("the aura is worth %v holy damage, want the game files' 30", buffs.RetributionAuraValue(0))
	}

	category := char.ExclusiveEffectManager.GetExclusiveEffectCategory(buffs.RetributionAuraCategory)
	if !category.SingleAura || len(category.Effects()) != 1 || category.Effects()[0].Priority != 30 {
		t.Errorf("the category is single-aura %v with %d effects, first bid %v; want one bidding 30",
			category.SingleAura, len(category.Effects()), category.Effects()[0].Priority)
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

	want := buffs.RetributionAuraValue(0) + buffs.RetributionAuraSpellPowerCoefficient*450
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
// first one's aura to fall off, and a third cast waits for the six minutes the
// game files state.
func TestGeneratedInnervatesTakeTurnsBetweenTheirSources(t *testing.T) {
	sim := setupFakeSimWithBuffs(&proto.RaidBuffs{}, &proto.PartyBuffs{}, &proto.IndividualBuffs{Innervates: 2})
	char := sim.Raid.Parties[0].Players[0].GetCharacter()

	aura := char.GetAura("Innervates (External)")
	if aura == nil {
		t.Fatalf("no aura is labelled %q; the unit has %v", "Innervates (External)", auraLabels(char))
	}
	if aura.Duration != buffs.InnervatesDuration(0) || aura.Duration != time.Second*20 {
		t.Errorf("the innervate lasts %v, want the game files' 20 seconds", aura.Duration)
	}
	if buffs.InnervatesCooldown() != time.Minute*6 {
		t.Errorf("the innervate's cooldown is %v, want the game files' 6 minutes", buffs.InnervatesCooldown())
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
		t.Error("a third innervate lands before either druid's six minutes are up")
	}

	sim.CurrentTime = buffs.InnervatesCooldown() + time.Minute
	if !ready() {
		t.Error("the first druid cannot innervate again six minutes later")
	}
}

// Power Infusion is +20% damage and healing done while it is up, and nothing
// once it has expired. The damage half is the six schools the game files' mask
// names, which does not include physical.
func TestGeneratedPowerInfusionRaisesDamageAndHealingDone(t *testing.T) {
	sim := setupFakeSimWithBuffs(&proto.RaidBuffs{}, &proto.PartyBuffs{}, &proto.IndividualBuffs{PowerInfusions: 1})
	char := sim.Raid.Parties[0].Players[0].GetCharacter()

	aura := char.GetAura("Power Infusions (External)")
	if aura == nil {
		t.Fatalf("no aura is labelled %q; the unit has %v", "Power Infusions (External)", auraLabels(char))
	}
	if aura.Duration != time.Second*15 || buffs.PowerInfusionsCooldown() != time.Minute*3 {
		t.Errorf("the infusion lasts %v on a %v cooldown, want the game files' 15 seconds and 3 minutes",
			aura.Duration, buffs.PowerInfusionsCooldown())
	}

	healing := char.PseudoStats.HealingDealtMultiplier
	physical := stats.SchoolIndexPhysical
	before := char.PseudoStats.SchoolDamageDealtMultiplier[physical]

	aura.Activate(sim)
	// Spell 10060 states mask 126, which is every school but physical, so a
	// melee swing is not part of what the infusion raises.
	for _, school := range magicSchools {
		if got := char.PseudoStats.SchoolDamageDealtMultiplier[school]; got != 1.2 {
			t.Errorf("school %d deals %v times the damage, want the game files' 1.2", school, got)
		}
	}
	if got := char.PseudoStats.SchoolDamageDealtMultiplier[physical]; got != before {
		t.Errorf("physical damage dealt is %v, want the mask to have left it at %v", got, before)
	}
	if got := char.PseudoStats.HealingDealtMultiplier; got != healing*1.2 {
		t.Errorf("the infusion multiplies healing dealt by %v, want the game files' 1.2", got/healing)
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

// Mana Tide's aura states 290 mana every 3 seconds, which reaches the sim as
// the mana per 5 seconds it is worth; the totem cast states how long it stands.
func TestGeneratedManaTideTotemRestoresTheClientsMana(t *testing.T) {
	sim := setupFakeSimWithBuffs(&proto.RaidBuffs{}, &proto.PartyBuffs{ManaTideTotems: 1}, &proto.IndividualBuffs{})
	char := sim.Raid.Parties[0].Players[0].GetCharacter()

	aura := char.GetAura("Mana Tide Totem (External)")
	if aura == nil {
		t.Fatalf("no aura is labelled %q; the unit has %v", "Mana Tide Totem (External)", auraLabels(char))
	}
	if aura.Duration != time.Second*13 || buffs.ManaTideTotemsCooldown() != time.Minute*5 {
		t.Errorf("the totem stands for %v on a %v cooldown, want the game files' 13 seconds and 5 minutes",
			aura.Duration, buffs.ManaTideTotemsCooldown())
	}
	if want := 290 * 5000.0 / 3000.0; buffs.ManaTideTotemsValue(0) != want {
		t.Errorf("the totem is worth %v MP5, want the 290 per 3 seconds the game files state as %v",
			buffs.ManaTideTotemsValue(0), want)
	}

	before := char.GetStats()[stats.MP5]
	aura.Activate(sim)
	if got := char.GetStats()[stats.MP5] - before; got != buffs.ManaTideTotemsValue(0) {
		t.Errorf("the totem applied %v MP5, want %v", got, buffs.ManaTideTotemsValue(0))
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

// Stoneskin takes a flat amount off every physical hit, and Guardian Totems raises it by 10% a rank.
// The reduction lands on PseudoStats.BonusPhysicalDamageTaken, not a stats.Stats entry.
func TestGeneratedStoneskinTotemTakesTheTalentedValue(t *testing.T) {
	if got := buffs.StoneskinTotemValue(0); got != -30 {
		t.Errorf("untalented Stoneskin is worth %v, want the game files' -30", got)
	}
	if got := buffs.StoneskinTotemValue(2); got != -36 {
		t.Errorf("Stoneskin with 2/2 Guardian Totems is worth %v, want -36", got)
	}

	char := core.NewGeneratedBuffTestCharacter()
	core.ApplyBuffEffects(char, &proto.RaidBuffs{},
		&proto.PartyBuffs{StoneskinTotem: proto.TristateEffect_TristateEffectImproved}, &proto.IndividualBuffs{})
	char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)

	aura := char.GetAura("Stoneskin Totem (External)")
	if aura == nil {
		t.Fatalf("no aura is labelled %q; the unit has %v", "Stoneskin Totem (External)", auraLabels(char))
	}
	if got := char.PseudoStats.BonusPhysicalDamageTaken; got != -36 {
		t.Errorf("the talented totem applied %v physical damage taken, want -36", got)
	}
}

// Furious Howl does not stack with Battle Shout: the two share one category, and the stronger shout
// is the one that applies.
func TestGeneratedFuriousHowlDoesNotStackWithBattleShout(t *testing.T) {
	if got := buffs.FuriousHowlValue(0); got != 83 {
		t.Fatalf("Furious Howl is worth %v attack power, want the game files' 82 plus 4 levels of 0.4, floored to 83", got)
	}

	howlOnly := core.NewGeneratedBuffTestCharacter()
	core.ApplyBuffEffects(howlOnly, &proto.RaidBuffs{}, &proto.PartyBuffs{FuriousHowl: true}, &proto.IndividualBuffs{})
	howlOnly.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)
	if got := howlOnly.GetStats()[stats.AttackPower]; got != buffs.FuriousHowlValue(0) {
		t.Errorf("Furious Howl alone applied %v attack power, want %v", got, buffs.FuriousHowlValue(0))
	}

	char := core.NewGeneratedBuffTestCharacter()
	core.ApplyBuffEffects(char, &proto.RaidBuffs{},
		&proto.PartyBuffs{BattleShout: proto.TristateEffect_TristateEffectRegular, FuriousHowl: true}, &proto.IndividualBuffs{})
	char.ApplyBuildPhaseAuras(core.CharacterBuildPhaseBuffs)
	if got := char.GetStats()[stats.AttackPower]; got != buffs.BattleShoutValue(0) {
		t.Errorf("attack power is %v under both shouts, want Battle Shout's %v alone", got, buffs.BattleShoutValue(0))
	}
}

func auraLabels(char *core.Character) []string {
	labels := make([]string, 0, len(char.GetAuras()))
	for _, aura := range char.GetAuras() {
		labels = append(labels, aura.Label)
	}
	return labels
}

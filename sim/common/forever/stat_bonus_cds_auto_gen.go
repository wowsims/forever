package forever

import (
	"time"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
)

func RegisterAllOnUseCds() {

	//
	// unsupported: 12257 deals no damage and heals no one (A_PERIODIC_TRIGGER_SPELL, E_TRIGGER_SPELL)
	// shared.NewSimpleStatActive(744) // Thunderbrew's Boot Flask - https://www.wowhead.com/forever/spell=12257
	// unsupported: 835 deals no damage and heals no one (A_MOD_STUN)
	// shared.NewSimpleStatActive(1404) // Tidal Charm - https://www.wowhead.com/forever/spell=835
	// on use: 1317424 (E_HEAL)
	// unsupported: the heal lands on implicit target 21, not the wearer
	// shared.NewSpellDataHealOnUse(1713) // Ankh of Life - https://www.wowhead.com/forever/spell=1317424
	// unsupported: 14530 deals no damage and heals no one (A_MOD_INCREASE_SPEED)
	// shared.NewSimpleStatActive(2820) // Nifty Stopwatch - https://www.wowhead.com/forever/spell=14530
	// unsupported: 1317740 deals no damage and heals no one (E_DUMMY)
	// shared.NewSimpleStatActive(4130) // Smotts' Compass - https://www.wowhead.com/forever/spell=1317740
	// on use: 5024 (A_PERIODIC_DAMAGE); not simulated: A_MOD_INCREASE_SPEED, A_PERIODIC_MANA_LEECH
	// unsupported: the damage lands on implicit target 1, not an enemy
	// shared.NewSpellDataDamageOnUse(4984) // Skull of Impending Doom - https://www.wowhead.com/forever/spell=5024
	// unsupported: 1139 deals no damage and heals no one (A_MOD_DECREASE_SPEED, A_MOD_MELEE_RANGED_HASTE_2)
	// shared.NewSimpleStatActive(5079) // Cold Basilisk Eye - https://www.wowhead.com/forever/spell=1139
	// unsupported: 8892 deals no damage and heals no one (A_MOD_INCREASE_SPEED)
	// shared.NewSimpleStatActive(7189) // Goblin Rocket Boots - https://www.wowhead.com/forever/spell=8892
	// unsupported: 9175 deals no damage and heals no one (A_MOD_INCREASE_SPEED)
	// shared.NewSimpleStatActive(7391) // Swift Boots - https://www.wowhead.com/forever/spell=9175
	// unsupported: 14537 deals no damage and heals no one (E_DUMMY)
	// shared.NewSimpleStatActive(7734) // Six Demon Bag - https://www.wowhead.com/forever/spell=14537
	// on use: 10577 (E_HEAL)
	// unsupported: the heal lands on implicit target 21, not the wearer
	// shared.NewSpellDataHealOnUse(8346) // Gauntlets of the Sea - https://www.wowhead.com/forever/spell=10577
	// unsupported: 12022 deals no damage and heals no one (E_APPLY_AREA_AURA_PARTY)
	// shared.NewSimpleStatActive(10030) // Admiral's Hat - https://www.wowhead.com/forever/spell=12022
	// unsupported: 12438 deals no damage and heals no one (A_FEATHER_FALL)
	// shared.NewSimpleStatActive(10518) // Parachute Cloak - https://www.wowhead.com/forever/spell=12438
	// unsupported: 22641 deals no damage and heals no one (E_CHARGE, E_TRIGGER_SPELL)
	// shared.NewSimpleStatActive(10588) // Goblin Rocket Helmet - https://www.wowhead.com/forever/spell=22641
	// unsupported: 13141 deals no damage and heals no one (A_MOD_INCREASE_SPEED)
	// shared.NewSimpleStatActive(10724) // Gnomish Rocket Boots - https://www.wowhead.com/forever/spell=13141
	// unsupported: 13180 deals no damage and heals no one (E_DUMMY)
	// shared.NewSimpleStatActive(10726) // Gnomish Mind Control Cap - https://www.wowhead.com/forever/spell=13180
	// unsupported: 18364 deals no damage and heals no one (A_PERIODIC_TRIGGER_SPELL)
	// shared.NewSimpleStatActive(14134) // Cloak of Fire - https://www.wowhead.com/forever/spell=18364
	// on use: 18386 (E_HEAL)
	// unsupported: the heal lands on implicit target 5, not the wearer
	// shared.NewSpellDataHealOnUse(14153) // Robe of the Void - https://www.wowhead.com/forever/spell=18386
	// on use: 1287808 (A_SCHOOL_ABSORB); not simulated: A_DUMMY
	// unsupported: the absorb of 10000000000 beside an A_DUMMY absorbs only the spells a script names, which the game files do not list
	// shared.NewSpellDataAbsorbOnUse(18406) // Onyxia Blood Talisman - https://www.wowhead.com/forever/spell=1287808
	// unsupported: 23131 deals no damage and heals no one (A_REFLECT_SPELLS_SCHOOL)
	// shared.NewSimpleStatActive(18634) // Gyrofreeze Ice Reflector - https://www.wowhead.com/forever/spell=23131
	// on use: 23064 (E_HEAL); not simulated: E_DISPEL_MECHANIC, E_ENERGIZE
	// unsupported: the heal lands on implicit target 21, not the wearer
	// shared.NewSpellDataHealOnUse(18637) // Major Recombobulator - https://www.wowhead.com/forever/spell=23064
	// unsupported: 23097 deals no damage and heals no one (A_REFLECT_SPELLS_SCHOOL)
	// shared.NewSimpleStatActive(18638) // Hyper-Radiant Flame Reflector - https://www.wowhead.com/forever/spell=23097
	// unsupported: 23132 deals no damage and heals no one (A_REFLECT_SPELLS_SCHOOL)
	// shared.NewSimpleStatActive(18639) // Ultra-Flash Shadow Reflector - https://www.wowhead.com/forever/spell=23132
	// unsupported: 23595 deals no damage and heals no one (E_DISPEL_MECHANIC)
	// shared.NewSimpleStatActive(19141) // Luffa - https://www.wowhead.com/forever/spell=23595
	// unsupported: 23721 deals no damage and heals no one (A_PROC_TRIGGER_SPELL)
	// shared.NewSimpleStatActive(19336) // Arcane Infused Gem - https://www.wowhead.com/forever/spell=23721
	// unsupported: 23724 deals no damage and heals no one (A_ADD_PCT_MODIFIER)
	// shared.NewSimpleStatActive(19340) // Rune of Metamorphosis - https://www.wowhead.com/forever/spell=23724
	// unsupported: 23725 deals no damage and heals no one (E_DUMMY)
	// shared.NewSimpleStatActive(19341) // Lifegiving Gem - https://www.wowhead.com/forever/spell=23725
	// unsupported: 23726 deals no damage and heals no one (A_ADD_FLAT_MODIFIER)
	// shared.NewSimpleStatActive(19342) // Venomous Totem - https://www.wowhead.com/forever/spell=23726
	// unsupported: 24574 deals no damage and heals no one (A_PROC_TRIGGER_SPELL, E_TRIGGER_SPELL)
	// shared.NewSimpleStatActive(19948) // Zandalarian Hero Badge - https://www.wowhead.com/forever/spell=24574
	// unsupported: 24661 deals no damage and heals no one (A_DUMMY)
	// shared.NewSimpleStatActive(19949) // Zandalarian Hero Medallion - https://www.wowhead.com/forever/spell=24661
	// unsupported: 24658 deals no damage and heals no one (A_DUMMY, E_TRIGGER_SPELL)
	// shared.NewSimpleStatActive(19950) // Zandalarian Hero Charm - https://www.wowhead.com/forever/spell=24658
	// unsupported: 24531 deals no damage and heals no one (E_DUMMY)
	// shared.NewSimpleStatActive(19953) // Renataki's Charm of Beasts - https://www.wowhead.com/forever/spell=24531
	// unsupported: 24542 deals no damage and heals no one (A_ADD_PCT_MODIFIER, A_ADD_PCT_MODIFIER)
	// shared.NewSimpleStatActive(19955) // Wushoolay's Charm of Nature - https://www.wowhead.com/forever/spell=24542
	// unsupported: 24499 deals no damage and heals no one (A_ADD_PCT_MODIFIER)
	// shared.NewSimpleStatActive(19956) // Wushoolay's Charm of Spirits - https://www.wowhead.com/forever/spell=24499
	// unsupported: 24543 deals no damage and heals no one (A_ADD_FLAT_MODIFIER)
	// shared.NewSimpleStatActive(19957) // Hazza'rah's Charm of Destruction - https://www.wowhead.com/forever/spell=24543
	// unsupported: 24546 deals no damage and heals no one (A_ADD_PCT_MODIFIER, A_ADD_PCT_MODIFIER)
	// shared.NewSimpleStatActive(19958) // Hazza'rah's Charm of Healing - https://www.wowhead.com/forever/spell=24546
	// unsupported: 24544 deals no damage and heals no one (A_ADD_FLAT_MODIFIER, A_ADD_PCT_MODIFIER)
	// shared.NewSimpleStatActive(19959) // Hazza'rah's Charm of Magic - https://www.wowhead.com/forever/spell=24544
	// unsupported: 25892 deals no damage and heals no one (E_THREAT)
	// shared.NewSimpleStatActive(21181) // Grace of Earth - https://www.wowhead.com/forever/spell=25892
	// unsupported: 26168 deals no damage and heals no one (A_DAMAGE_SHIELD)
	// shared.NewSimpleStatActive(21488) // Fetish of Chitinous Spikes - https://www.wowhead.com/forever/spell=26168
	// unsupported: 26467 deals no damage and heals no one (A_PROC_TRIGGER_SPELL)
	// shared.NewSimpleStatActive(21625) // Scarab Brooch - https://www.wowhead.com/forever/spell=26467
	// unsupported: 26400 deals no damage and heals no one (A_MOD_THREAT)
	// shared.NewSimpleStatActive(21647) // Fetish of the Sand Reaver - https://www.wowhead.com/forever/spell=26400
	// unsupported: 26463 deals no damage and heals no one (A_PROC_TRIGGER_SPELL, E_TRIGGER_SPELL)
	// shared.NewSimpleStatActive(21685) // Petrified Scarab - https://www.wowhead.com/forever/spell=26463
	// unsupported: 28862 deals no damage and heals no one (A_MOD_THREAT)
	// shared.NewSimpleStatActive(23001) // Eye of Diminution - https://www.wowhead.com/forever/spell=28862
	// unsupported: 28773 deals no damage and heals no one (A_MOD_BLOCK_VALUE_FLAT)
	// shared.NewSimpleStatActive(23040) // Glyph of Deflection - https://www.wowhead.com/forever/spell=28773
	// unsupported: 1306267 deals no damage and heals no one (E_FORCE_CAST_2)
	// shared.NewSimpleStatActive(221315) // Traveler's Symbols - https://www.wowhead.com/forever/spell=1306267
	// unsupported: 23453 deals no damage and heals no one (E_DUMMY)
	// shared.NewSimpleStatActive(260819) // EZ-Thro Field Transporter: Gadgetzan - https://www.wowhead.com/forever/spell=23453
	// unsupported: 1269339 deals no damage and heals no one (E_DUMMY)
	// shared.NewSimpleStatActive(260821) // EZ and SAF Field Transporter: Mt. Hyjal - https://www.wowhead.com/forever/spell=1269339
	// unsupported: 1270941 deals no damage and heals no one (E_DUMMY)
	// shared.NewSimpleStatActive(260824) // Gnomish Poultryizer - https://www.wowhead.com/forever/spell=1270941
	// unsupported: 1296664 deals no damage and heals no one (A_MOD_CHARM)
	// shared.NewSimpleStatActive(269741) // Scented Runewood Brooch - https://www.wowhead.com/forever/spell=1296664
	// on use: 1291097 (A_SCHOOL_ABSORB)
	// unsupported: the damage of 1291099 (Sigmoid Revenge) the absorb's row names is not simulated
	// shared.NewSpellDataAbsorbOnUse(272437) // Adaptive Combat Assistant - https://www.wowhead.com/forever/spell=1291097
	// on use: 1291101 (A_MOD_CRIT_PCT)
	// unsupported: the buff drops after ProcCharges 1, which the aura does not spend
	// shared.NewSpellDataAuraOnUse(272438) // Weakness Analyzer - https://www.wowhead.com/forever/spell=1291101
	// unsupported: 1291105 deals no damage and heals no one (A_MOD_BLOCK_PERCENT)
	// shared.NewSimpleStatActive(272440) // Defender's Grip Stabilizer - https://www.wowhead.com/forever/spell=1291105
	// unsupported: 1293820 deals no damage and heals no one (A_MOD_DODGE_PERCENT)
	// shared.NewSimpleStatActive(274386) // Toy Soldier - https://www.wowhead.com/forever/spell=1293820
	// unsupported: 1295271 deals no damage and heals no one (E_DUMMY, E_DUMMY, A_DUMMY)
	// shared.NewSimpleStatActive(274760) // Everlook Delivery Bot - https://www.wowhead.com/forever/spell=1295271
	// unsupported: 1295272 deals no damage and heals no one (E_TRIGGER_MISSILE)
	// shared.NewSimpleStatActive(274761) // Parachute-Priest Pager - https://www.wowhead.com/forever/spell=1295272
	// unsupported: 1296564 deals no damage and heals no one (A_MOD_STUN)
	// shared.NewSimpleStatActive(275347) // Lichbane - https://www.wowhead.com/forever/spell=1296564
	// unsupported: 1297762 deals no damage and heals no one (A_FEATHER_FALL)
	// shared.NewSimpleStatActive(275729) // Rusty Propeller Blade - https://www.wowhead.com/forever/spell=1297762
	// unsupported: 1299440 deals no damage and heals no one (A_MOD_FEAR)
	// shared.NewSimpleStatActive(276337) // Thaelemaches' Talisman - https://www.wowhead.com/forever/spell=1299440
	// unsupported: 1302833 deals no damage and heals no one (A_MOD_SPEED_ALWAYS, E_DUMMY, E_FORCE_CAST)
	// shared.NewSimpleStatActive(277844) // Mirage Mirror - https://www.wowhead.com/forever/spell=1302833
	// unsupported: 1320579 deals no damage and heals no one (A_MOD_DECREASE_SPEED, E_DUMMY)
	// shared.NewSimpleStatActive(286556) // Winds of Tanaris - https://www.wowhead.com/forever/spell=1320579

	// Absorbs
	// on use: 10618 (A_SCHOOL_ABSORB)
	shared.NewSpellDataAbsorbOnUse(8367) // Dragonscale Breastplate - https://www.wowhead.com/forever/spell=10618
	// on use: 12561 (A_SCHOOL_ABSORB)
	shared.NewSpellDataAbsorbOnUse(10543) // Goblin Construction Helmet - https://www.wowhead.com/forever/spell=12561
	// on use: 13234 (A_SCHOOL_ABSORB)
	shared.NewSpellDataAbsorbOnUse(10721) // Gnomish Harm Prevention Belt - https://www.wowhead.com/forever/spell=13234
	// on use: 21956 (A_SCHOOL_ABSORB)
	shared.NewSpellDataAbsorbOnUse(17759) // Mark of Resolution - https://www.wowhead.com/forever/spell=21956
	// on use: 23506 (A_SCHOOL_ABSORB)
	shared.NewSpellDataAbsorbOnUse(19024) // Arena Grand Master - https://www.wowhead.com/forever/spell=23506
	// on use: 23991 (A_SCHOOL_ABSORB)
	shared.NewSpellDataAbsorbOnUse(20071) // Talisman of Arathor - https://www.wowhead.com/forever/spell=23991
	// on use: 23991 (A_SCHOOL_ABSORB)
	shared.NewSpellDataAbsorbOnUse(20072) // Defiler's Talisman - https://www.wowhead.com/forever/spell=23991
	// on use: 25746 (A_SCHOOL_ABSORB)
	shared.NewSpellDataAbsorbOnUse(21115) // Defiler's Talisman - https://www.wowhead.com/forever/spell=25746
	// on use: 25747 (A_SCHOOL_ABSORB)
	shared.NewSpellDataAbsorbOnUse(21116) // Defiler's Talisman - https://www.wowhead.com/forever/spell=25747
	// on use: 25746 (A_SCHOOL_ABSORB)
	shared.NewSpellDataAbsorbOnUse(21117) // Talisman of Arathor - https://www.wowhead.com/forever/spell=25746
	// on use: 25747 (A_SCHOOL_ABSORB)
	shared.NewSpellDataAbsorbOnUse(21118) // Talisman of Arathor - https://www.wowhead.com/forever/spell=25747
	// on use: 29506 (A_SCHOOL_ABSORB)
	shared.NewSpellDataAbsorbOnUse(23558) // The Burrower's Shell - https://www.wowhead.com/forever/spell=29506
	// on use: 1314412 (A_SCHOOL_ABSORB)
	shared.NewSpellDataAbsorbOnUse(282778) // Mark of the Red Flight - https://www.wowhead.com/forever/spell=1314412

	// Agility / Intellect / Spirit / Stamina / Strength
	shared.NewSimpleStatActive(270226) // Golden Banana - https://www.wowhead.com/forever/spell=1287571

	// ArcaneResistance / FireResistance / FrostResistance / NatureResistance / ShadowResistance
	shared.NewSimpleStatActive(23042) // Loatheb's Reflection - https://www.wowhead.com/forever/spell=28778

	// Armor
	shared.NewSimpleStatActive(1447)  // Ring of Saviors - https://www.wowhead.com/forever/spell=18826
	shared.NewSimpleStatActive(11811) // Smoking Heart of the Mountain - https://www.wowhead.com/forever/spell=1300752
	// not simulated: the proc the buff carries, 23781 (E_HEAL)
	shared.NewSimpleStatActive(19345) // Aegis of Preservation - https://www.wowhead.com/forever/spell=23780

	// ArmorPenetration
	// Badge of the Swarmguard - https://www.wowhead.com/forever/spell=26480
	shared.NewStackingStatBonusCD(shared.StackingStatBonusCD{
		Name:                   "Badge of the Swarmguard",
		ID:                     21670,
		CD:                     time.Millisecond * 180000,
		Callback:               core.CallbackOnSpellHitDealt,
		ProcMask:               core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial | core.ProcMaskRangedAuto | core.ProcMaskRangedSpecial,
		Outcome:                core.OutcomeLanded,
		RequireDamageDealt:     false,
		RequireDamageComponent: true,
		TrinketLimitsDuration:  true,
	})

	// AttackPower / RangedAttackPower
	shared.NewSimpleStatActive(21180)  // Earthstrike - https://www.wowhead.com/forever/spell=25891
	shared.NewSimpleStatActive(23041)  // Slayer's Crest - https://www.wowhead.com/forever/spell=28777
	shared.NewSimpleStatActive(249470) // Molten Heart of the Mountain - https://www.wowhead.com/forever/spell=1249113

	// Auras
	// on use: 23734 (A_MOD_DAMAGE_PERCENT_DONE, A_MOD_HEALING_DONE_PERCENT, A_MOD_POWER_COST_SCHOOL_PCT)
	shared.NewSpellDataAuraOnUse(19344) // Natural Alignment Crystal - https://www.wowhead.com/forever/spell=23734
	// on use: 24498 (A_MOD_SPELL_CRIT_CHANCE)
	shared.NewSpellDataAuraOnUse(19952) // Gri'lek's Charm of Valor - https://www.wowhead.com/forever/spell=24498

	// Damage
	// on use: 10578 (E_SCHOOL_DAMAGE, A_PERIODIC_DAMAGE)
	shared.NewSpellDataDamageOnUse(8348) // Helm of Fire - https://www.wowhead.com/forever/spell=10578
	// on use: 15712 (E_SCHOOL_DAMAGE); not simulated: E_TRIGGER_SPELL, E_TRIGGER_SPELL
	shared.NewSpellDataDamageOnUse(11905) // Linken's Boomerang - https://www.wowhead.com/forever/spell=15712
	// on use: 26789 (E_SCHOOL_DAMAGE)
	shared.NewSpellDataDamageOnUse(21891) // Shard of the Fallen Star - https://www.wowhead.com/forever/spell=26789
	// on use: 1295270 (E_SCHOOL_DAMAGE); not simulated: E_NONE, E_NONE
	shared.NewSpellDataDamageOnUse(274759) // Everlook Pathcarver - https://www.wowhead.com/forever/spell=1295270

	// FireResistance
	shared.NewSimpleStatActive(2802)  // Blazing Emblem - https://www.wowhead.com/forever/spell=13744
	shared.NewSimpleStatActive(13164) // Heart of the Scale - https://www.wowhead.com/forever/spell=17275

	// FrostDamage / ShadowDamage
	shared.NewSimpleStatActive(249469) // Frozen Heart of the Mountain - https://www.wowhead.com/forever/spell=1249110

	// HealingPower
	shared.NewSimpleStatActive(20636) // Hibernation Crystal - https://www.wowhead.com/forever/spell=24998
	shared.NewSimpleStatActive(23047) // Eye of the Dead - https://www.wowhead.com/forever/spell=28780

	// HealingPower / SpellDamage
	shared.NewSimpleStatActive(18820) // Talisman of Ephemeral Power - https://www.wowhead.com/forever/spell=23271
	// Talisman of Ascendance - https://www.wowhead.com/forever/spell=28200
	shared.NewStackingStatBonusCD(shared.StackingStatBonusCD{
		Name:                   "Talisman of Ascendance",
		ID:                     22678,
		CD:                     time.Millisecond * 60000,
		Callback:               core.CallbackOnSpellHitDealt | core.CallbackOnHealDealt,
		ProcMask:               core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial | core.ProcMaskRangedAuto | core.ProcMaskRangedSpecial | core.ProcMaskSpellDamage | core.ProcMaskSpellHealing,
		Outcome:                core.OutcomeLanded,
		RequireDamageDealt:     false,
		RequireDamageComponent: true,
		TrinketLimitsDuration:  true,
	})
	shared.NewSimpleStatActive(23046) // The Restrained Essence of Sapphiron - https://www.wowhead.com/forever/spell=28779

	// Heals
	// on use: 18957 (A_PERIODIC_HEAL)
	shared.NewSpellDataHealOnUse(15107) // Orb of Noh'Orahil - https://www.wowhead.com/forever/spell=18957
	// on use: 18957 (A_PERIODIC_HEAL)
	shared.NewSpellDataHealOnUse(15108) // Orb of Dar'Orahil - https://www.wowhead.com/forever/spell=18957
	// on use: 20631 (A_PERIODIC_HEAL)
	shared.NewSpellDataHealOnUse(16768) // Furbolg Medicine Pouch - https://www.wowhead.com/forever/spell=20631

	// MP5
	shared.NewSimpleStatActive(4696) // Lapidis Tankard of Tidesippe - https://www.wowhead.com/forever/spell=1135

	// Resources
	// on use: 9174 (E_ENERGIZE)
	shared.NewSpellDataEnergizeOnUse(4264) // Barbaric Belt - https://www.wowhead.com/forever/spell=9174
	// on use: 9253 (E_ENERGIZE)
	shared.NewSpellDataEnergizeOnUse(7515) // Celestial Orb - https://www.wowhead.com/forever/spell=9253
	// on use: 18385 (E_ENERGIZE)
	shared.NewSpellDataEnergizeOnUse(14152) // Robe of the Archmage - https://www.wowhead.com/forever/spell=18385
	// on use: 24571 (E_ENERGIZE)
	shared.NewSpellDataEnergizeOnUse(19951) // Gri'lek's Charm of Might - https://www.wowhead.com/forever/spell=24571
	// on use: 24532 (E_ENERGIZE)
	shared.NewSpellDataEnergizeOnUse(19954) // Renataki's Charm of Trickery - https://www.wowhead.com/forever/spell=24532
	// on use: 24884 (A_PERIODIC_ENERGIZE)
	shared.NewSpellDataEnergizeOnUse(20525) // Earthen Sigil - https://www.wowhead.com/forever/spell=24884
	// on use: 28760 (E_ENERGIZE)
	shared.NewSpellDataEnergizeOnUse(23027) // Warmth of Forgiveness - https://www.wowhead.com/forever/spell=28760

	// Skipped
	// Not simulated: Staff of Conjuring: "Conjure Food" (8736) - ignored effect type 24
	// https://www.wowhead.com/forever/spell=8736
	// Not simulated: Orb of Deception: "Orb of Deception" (16739) - ignored aura type 56
	// https://www.wowhead.com/forever/spell=16739
	// Not simulated: Dog Whistle: "Summon Tracking Hound" (9515) - ignored effect type 28
	// https://www.wowhead.com/forever/spell=9515
	// Not simulated: Spider Belt: "Immune Root" (9774) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=9774
	// Not simulated: Enchanted Moonstalker Cloak: "Form of the Moonstalker" (6298) - ignored aura type 56
	// https://www.wowhead.com/forever/spell=6298
	// Not simulated: Ornate Mithril Boots: "Immune Root" (9774) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=9774
	// Not simulated: Glimmering Mithril Insignia: "Fearless" (12733) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=12733
	// Not simulated: Arcanite Dragonling: "Arcanite Dragonling" (19804) - ignored effect type 28
	// https://www.wowhead.com/forever/spell=19804
	// Not simulated: Insignia of the Horde: "Immune Root/Snare/Stun" (5579) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=5579
	// Not simulated: Insignia of the Horde: "Immune Root/Snare/Stun" (5579) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=5579
	// Not simulated: Insignia of the Horde: "Immune Root/Snare/Stun" (5579) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=5579
	// Not simulated: Insignia of the Horde: "Immune Charm/Fear/Polymorph" (23273) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=23273
	// Not simulated: Insignia of the Horde: "Immune Fear/Polymorph/Snare" (23274) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=23274
	// Not simulated: Insignia of the Horde: "Immune Fear/Polymorph/Stun" (23276) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=23276
	// Not simulated: Insignia of the Horde: "Immune Charm/Fear/Polymorph" (23273) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=23273
	// Not simulated: Insignia of the Horde: "Immune Charm/Fear/Stun" (23277) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=23277
	// Not simulated: Insignia of the Alliance: "Immune Root/Snare/Stun" (5579) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=5579
	// Not simulated: Insignia of the Alliance: "Immune Root/Snare/Stun" (5579) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=5579
	// Not simulated: Insignia of the Alliance: "Immune Charm/Fear/Polymorph" (23273) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=23273
	// Not simulated: Insignia of the Alliance: "Immune Charm/Fear/Polymorph" (23273) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=23273
	// Not simulated: Insignia of the Alliance: "Immune Fear/Polymorph/Snare" (23274) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=23274
	// Not simulated: Insignia of the Alliance: "Immune Fear/Polymorph/Stun" (23276) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=23276
	// Not simulated: Insignia of the Alliance: "Immune Charm/Fear/Stun" (23277) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=23277
	// Not simulated: Insignia of the Alliance: "Immune Fear/Polymorph/Stun" (23276) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=23276
	// Not simulated: Defender of the Timbermaw: "Defender of the Timbermaw" (26066) - ignored effect type 28
	// https://www.wowhead.com/forever/spell=26066
	// Not simulated: Vanquished Tentacle of C'Thun: "Tentacle Call" (26391) - ignored effect type 28
	// https://www.wowhead.com/forever/spell=26391
	// Not simulated: Insignia of the Alliance: "Immune Root/Snare/Stun" (5579) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=5579
	// Not simulated: Insignia of the Alliance: "Immune Charm/Fear/Polymorph" (23273) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=23273
	// Not simulated: Insignia of the Alliance: "Immune Fear/Polymorph/Stun" (23276) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=23276
	// Not simulated: Insignia of the Alliance: "Immune Fear/Polymorph/Stun" (23276) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=23276
	// Not simulated: Insignia of the Alliance: "Immune Charm/Fear/Polymorph" (23273) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=23273
	// Not simulated: Insignia of the Alliance: "Immune Root/Snare/Stun" (5579) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=5579
	// Not simulated: Insignia of the Alliance: "Immune Charm/Fear/Stun" (23277) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=23277
	// Not simulated: Insignia of the Alliance: "Immune Fear/Polymorph/Snare" (23274) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=23274
	// Not simulated: Insignia of the Horde: "Immune Root/Snare/Stun" (5579) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=5579
	// Not simulated: Insignia of the Horde: "Immune Charm/Fear/Polymorph" (23273) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=23273
	// Not simulated: Insignia of the Horde: "Immune Fear/Polymorph/Stun" (23276) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=23276
	// Not simulated: Insignia of the Horde: "Immune Charm/Fear/Polymorph" (23273) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=23273
	// Not simulated: Insignia of the Horde: "Immune Fear/Polymorph/Snare" (23274) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=23274
	// Not simulated: Insignia of the Horde: "Immune Charm/Fear/Stun" (23277) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=23277
	// Not simulated: Insignia of the Horde: "Immune Root/Snare/Stun" (5579) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=5579
	// Not simulated: Insignia of the Horde: "Immune Root/Snare/Stun" (5579) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=5579
	// Not simulated: SAF-T Emergency Ripper: Everlook: "Everlook Transporter" (23442) - ignored effect type 252
	// https://www.wowhead.com/forever/spell=23442
	// Not simulated: Defender of the Barkskin: "Defender of the Barkskin" (1294063) - ignored effect type 28
	// https://www.wowhead.com/forever/spell=1294063
	// Not simulated: Graverobber's Shovel: "Dig" (1292560) - ignored effect type 24
	// https://www.wowhead.com/forever/spell=1292560
	// Not simulated: Greater Insignia of the Horde: "PvP Trinket" (438273) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=438273
	// Not simulated: Greater Insignia of the Alliance: "PvP Trinket" (438273) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=438273
	// Not simulated: Field Agent Beverage: "Field Agent Beverage" (1297448) - ignored aura type 17
	// https://www.wowhead.com/forever/spell=1297448

	// Speed
	// on use: 23723 (A_MOD_CASTING_SPEED_NOT_STACK)
	shared.NewSpellDataSpeedOnUse(19339) // Mind Quickening Gem - https://www.wowhead.com/forever/spell=23723
	// on use: 23733 (A_MOD_CASTING_SPEED_NOT_STACK, A_MOD_MELEE_HASTE_3)
	shared.NewSpellDataSpeedOnUse(19343) // Scrolls of Blinding Light - https://www.wowhead.com/forever/spell=23733
	// on use: 28866 (A_MOD_MELEE_HASTE_3, A_MOD_RANGED_HASTE)
	shared.NewSpellDataSpeedOnUse(22954) // Kiss of the Spider - https://www.wowhead.com/forever/spell=28866

	// SpellDamage / SpellPiercing
	shared.NewSimpleStatActive(21473) // Eye of Moam - https://www.wowhead.com/forever/spell=26166

	// Spirit
	shared.NewSimpleStatActive(272439) // Serenity Field - https://www.wowhead.com/forever/spell=1291103
}

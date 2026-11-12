package forever

import (
	"time"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
)

func RegisterAllOnUseCds() {
	defer shared.Generated()()

	//
	// unsupported: 18820 deals no damage and heals no one (A_MOD_POWER_COST_SCHOOL)
	// shared.NewSimpleStatActive(940) // Robes of Insight - https://www.wowhead.com/forever/spell=18820
	// unsupported: 835 deals no damage and heals no one (A_MOD_STUN)
	// shared.NewSimpleStatActive(1404) // Tidal Charm - https://www.wowhead.com/forever/spell=835
	// unsupported: 1300364 deals no damage and heals no one (A_MOD_DETECTED_RANGE)
	// shared.NewSimpleStatActive(1490) // Guardian Talisman - https://www.wowhead.com/forever/spell=1300364
	// unsupported: 14530 deals no damage and heals no one (A_MOD_INCREASE_SPEED)
	// shared.NewSimpleStatActive(2820) // Nifty Stopwatch - https://www.wowhead.com/forever/spell=14530
	// unsupported: 1317740 deals no damage and heals no one (E_DUMMY)
	// shared.NewSimpleStatActive(4130) // Smotts' Compass - https://www.wowhead.com/forever/spell=1317740
	// on use: 5024 (A_PERIODIC_DAMAGE); not simulated: A_MOD_INCREASE_SPEED, A_PERIODIC_MANA_LEECH
	// unsupported: the damage lands on implicit target 1, not an enemy
	// shared.NewSpellDataDamageOnUse(4984) // Skull of Impending Doom - https://www.wowhead.com/forever/spell=5024
	// unsupported: 8892 deals no damage and heals no one (A_MOD_INCREASE_SPEED)
	// shared.NewSimpleStatActive(7189) // Goblin Rocket Boots - https://www.wowhead.com/forever/spell=8892
	// unsupported: 9175 deals no damage and heals no one (A_MOD_INCREASE_SPEED)
	// shared.NewSimpleStatActive(7391) // Swift Boots - https://www.wowhead.com/forever/spell=9175
	// unsupported: 14537 deals no damage and heals no one (E_DUMMY)
	// shared.NewSimpleStatActive(7734) // Six Demon Bag - https://www.wowhead.com/forever/spell=14537
	// unsupported: 22641 deals no damage and heals no one (E_CHARGE, E_TRIGGER_SPELL)
	// shared.NewSimpleStatActive(9394) // Horned Viking Helmet - https://www.wowhead.com/forever/spell=22641
	// unsupported: 12438 deals no damage and heals no one (A_FEATHER_FALL)
	// shared.NewSimpleStatActive(9404) // Olaf's All Purpose Shield - https://www.wowhead.com/forever/spell=12438
	// unsupported: 12438 deals no damage and heals no one (A_FEATHER_FALL)
	// shared.NewSimpleStatActive(10518) // Parachute Cloak - https://www.wowhead.com/forever/spell=12438
	// unsupported: 22641 deals no damage and heals no one (E_CHARGE, E_TRIGGER_SPELL)
	// shared.NewSimpleStatActive(10588) // Goblin Rocket Helmet - https://www.wowhead.com/forever/spell=22641
	// unsupported: 13141 deals no damage and heals no one (A_MOD_INCREASE_SPEED)
	// shared.NewSimpleStatActive(10724) // Gnomish Rocket Boots - https://www.wowhead.com/forever/spell=13141
	// unsupported: 13180 deals no damage and heals no one (E_DUMMY)
	// shared.NewSimpleStatActive(10726) // Gnomish Mind Control Cap - https://www.wowhead.com/forever/spell=13180
	// unsupported: 1300754 deals no damage and heals no one (A_MOD_CHARM)
	// shared.NewSimpleStatActive(11625) // Enthralled Sphere - https://www.wowhead.com/forever/spell=1300754
	// unsupported: 1300763 deals no damage and heals no one (E_DUMMY)
	// shared.NewSimpleStatActive(11702) // Grizzle's Skinner - https://www.wowhead.com/forever/spell=1300763
	// on use: 17447 (A_PERIODIC_DAMAGE); not simulated: A_PERIODIC_ENERGIZE
	// unsupported: the damage lands on implicit target 1, not an enemy
	// shared.NewSpellDataDamageOnUse(11808) // Circle of Flame - https://www.wowhead.com/forever/spell=17447
	// unsupported: 15646 deals no damage and heals no one (A_MOD_POWER_COST_SCHOOL)
	// shared.NewSimpleStatActive(11832) // Burst of Knowledge - https://www.wowhead.com/forever/spell=15646
	// unsupported: 1308917 deals no damage and heals no one (E_DISPEL_MECHANIC)
	// shared.NewSimpleStatActive(13213) // Smolderweb's Eye - https://www.wowhead.com/forever/spell=1308917
	// unsupported: 18400 deals no damage and heals no one (A_PERIODIC_TRIGGER_SPELL)
	// shared.NewSimpleStatActive(13379) // Piccolo of the Flaming Fire - https://www.wowhead.com/forever/spell=18400
	// unsupported: 1299443 deals no damage and heals no one (A_MOD_STUN)
	// shared.NewSimpleStatActive(13515) // Ramstein's Lightning Bolts - https://www.wowhead.com/forever/spell=1299443
	// unsupported: 1298497 deals no damage and heals no one (E_DUMMY)
	// shared.NewSimpleStatActive(13544) // Spectral Essence - https://www.wowhead.com/forever/spell=1298497
	// unsupported: 18264 deals no damage and heals no one (E_APPLY_AREA_AURA_PARTY)
	// shared.NewSimpleStatActive(13937) // Headmaster's Charge - https://www.wowhead.com/forever/spell=18264
	// unsupported: 1318944 deals no damage and heals no one (A_MOD_CRIT_CHANCE_FOR_CASTER)
	// shared.NewSimpleStatActive(13965) // Blackhand's Breadth - https://www.wowhead.com/forever/spell=1318944
	// on use: 18386 (E_HEAL)
	// unsupported: the heal lands on implicit target 5, not the wearer
	// shared.NewSpellDataHealOnUse(14153) // Robe of the Void - https://www.wowhead.com/forever/spell=18386
	// unsupported: 21954 deals no damage and heals no one (E_DISPEL)
	// shared.NewSimpleStatActive(17744) // Heart of Noxxion - https://www.wowhead.com/forever/spell=21954
	// on use: 1287808 (A_SCHOOL_ABSORB); not simulated: A_DUMMY
	// unsupported: the absorb of 10000000000 beside an A_DUMMY absorbs only the spells a script names, which the game files do not list
	// shared.NewSpellDataAbsorbOnUse(18406) // Onyxia Blood Talisman - https://www.wowhead.com/forever/spell=1287808
	// unsupported: 1302356 deals no damage and heals no one (A_PROC_TRIGGER_SPELL)
	// shared.NewSimpleStatActive(18537) // Counterattack Lodestone - https://www.wowhead.com/forever/spell=1302356
	// unsupported: 23131 deals no damage and heals no one (A_REFLECT_SPELLS_SCHOOL)
	// shared.NewSimpleStatActive(18634) // Gyrofreeze Ice Reflector - https://www.wowhead.com/forever/spell=23131
	// unsupported: 23097 deals no damage and heals no one (A_REFLECT_SPELLS_SCHOOL)
	// shared.NewSimpleStatActive(18638) // Hyper-Radiant Flame Reflector - https://www.wowhead.com/forever/spell=23097
	// unsupported: 23132 deals no damage and heals no one (A_REFLECT_SPELLS_SCHOOL)
	// shared.NewSimpleStatActive(18639) // Ultra-Flash Shadow Reflector - https://www.wowhead.com/forever/spell=23132
	// unsupported: 23595 deals no damage and heals no one (E_DISPEL_MECHANIC)
	// shared.NewSimpleStatActive(19141) // Luffa - https://www.wowhead.com/forever/spell=23595
	// unsupported: 24352 deals no damage and heals no one (A_DUMMY)
	// shared.NewSimpleStatActive(19991) // Devilsaur Eye - https://www.wowhead.com/forever/spell=24352
	// on use: 24353 (A_ADD_PCT_MODIFIER)
	// unsupported: the buff drops after ProcCharges 1, which the aura does not spend
	// shared.NewSpellDataAuraOnUse(19992) // Devilsaur Tooth - https://www.wowhead.com/forever/spell=24353
	// on use: 24389 (A_ADD_FLAT_MODIFIER, A_DUMMY)
	// unsupported: effect 1 A_ADD_FLAT_MODIFIER misc 0 is not parsed on the wearer; effect 2 A_DUMMY misc 0 is not parsed on the wearer
	// shared.NewSpellDataAuraOnUse(20036) // Fire Ruby - https://www.wowhead.com/forever/spell=24389
	// unsupported: 25112 deals no damage and heals no one (E_SUMMON_PET)
	// shared.NewSimpleStatActive(20534) // Abyss Shard - https://www.wowhead.com/forever/spell=25112
	// unsupported: 25892 deals no damage and heals no one (E_THREAT)
	// shared.NewSimpleStatActive(21181) // Grace of Earth - https://www.wowhead.com/forever/spell=25892
	// unsupported: 26467 deals no damage and heals no one (A_PROC_TRIGGER_SPELL)
	// shared.NewSimpleStatActive(21625) // Scarab Brooch - https://www.wowhead.com/forever/spell=26467
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
	// unsupported: 1295272 deals no damage and heals no one (E_TRIGGER_MISSILE)
	// shared.NewSimpleStatActive(274761) // Parachute-Priest Pager - https://www.wowhead.com/forever/spell=1295272
	// unsupported: 1315972 deals no damage and heals no one (A_MOD_STUN, A_MOD_STUN, A_DUMMY)
	// shared.NewSimpleStatActive(275009) // Emblem of Slamming - https://www.wowhead.com/forever/spell=1315972
	// unsupported: 1296564 deals no damage and heals no one (A_MOD_STUN)
	// shared.NewSimpleStatActive(275347) // Lichbane - https://www.wowhead.com/forever/spell=1296564
	// unsupported: 1297762 deals no damage and heals no one (A_FEATHER_FALL)
	// shared.NewSimpleStatActive(275729) // Rusty Propeller Blade - https://www.wowhead.com/forever/spell=1297762
	// unsupported: 1299440 deals no damage and heals no one (A_MOD_FEAR)
	// shared.NewSimpleStatActive(276337) // Thaelemaches' Talisman - https://www.wowhead.com/forever/spell=1299440
	// unsupported: 1302359 deals no damage and heals no one (A_PERIODIC_TRIGGER_SPELL)
	// shared.NewSimpleStatActive(277559) // Shen'dralar Relic - https://www.wowhead.com/forever/spell=1302359
	// unsupported: 1302833 deals no damage and heals no one (A_MOD_SPEED_ALWAYS, E_DUMMY, E_FORCE_CAST)
	// shared.NewSimpleStatActive(277844) // Mirage Mirror - https://www.wowhead.com/forever/spell=1302833
	// on use: 1309551 (A_MOD_DAMAGE_PERCENT_TAKEN, A_MOD_SPEED_ALWAYS)
	// unsupported: effect 2 A_MOD_SPEED_ALWAYS misc 0 is not parsed on the wearer
	// shared.NewSpellDataAuraOnUse(280074) // Scalding Flamehoof - https://www.wowhead.com/forever/spell=1309551
	// unsupported: 1315557 deals no damage and heals no one (A_MOD_STUN, A_ANIM_REPLACEMENT_SET)
	// shared.NewSimpleStatActive(284231) // Sayaad Charm - https://www.wowhead.com/forever/spell=1315557
	// unsupported: 1320579 deals no damage and heals no one (A_MOD_DECREASE_SPEED, E_DUMMY)
	// shared.NewSimpleStatActive(286556) // Winds of Tanaris - https://www.wowhead.com/forever/spell=1320579
	// unsupported: 13534 deals no damage and heals no one (A_MOD_DISARM)
	// shared.NewSimpleStatActive(287026) // Escape Plan - https://www.wowhead.com/forever/spell=13534
	// unsupported: 1325180 deals no damage and heals no one (E_ACTIVATE_OBJECT, E_CREATE_LOOT)
	// shared.NewSimpleStatActive(287045) // Ancient Mattock - https://www.wowhead.com/forever/spell=1325180

	// Absorbs
	// on use: 13744 (A_SCHOOL_ABSORB, A_MOD_RESISTANCE)
	shared.NewSpellDataAbsorbOnUse(2802) // Blazing Emblem - https://www.wowhead.com/forever/spell=13744
	// on use: 10618 (A_SCHOOL_ABSORB)
	shared.NewSpellDataAbsorbOnUse(8367) // Dragonscale Breastplate - https://www.wowhead.com/forever/spell=10618
	// on use: 12561 (A_SCHOOL_ABSORB)
	shared.NewSpellDataAbsorbOnUse(10543) // Goblin Construction Helmet - https://www.wowhead.com/forever/spell=12561
	// on use: 13234 (A_SCHOOL_ABSORB)
	shared.NewSpellDataAbsorbOnUse(10721) // Gnomish Harm Prevention Belt - https://www.wowhead.com/forever/spell=13234
	// on use: 17252 (A_SCHOOL_ABSORB); not simulated: A_PERIODIC_ENERGIZE
	shared.NewSpellDataAbsorbOnUse(13143) // Mark of the Dragon Lord - https://www.wowhead.com/forever/spell=17252
	// on use: 21956 (A_SCHOOL_ABSORB)
	shared.NewSpellDataAbsorbOnUse(17759) // Mark of Resolution - https://www.wowhead.com/forever/spell=21956
	// on use: 1301625 (A_SCHOOL_ABSORB)
	shared.NewSpellDataAbsorbOnUse(18047) // Flame Walkers - https://www.wowhead.com/forever/spell=1301625
	// on use: 1302362 (A_SCHOOL_ABSORB)
	shared.NewSpellDataAbsorbOnUse(18340) // Eidolon Talisman - https://www.wowhead.com/forever/spell=1302362
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
	// on use: 1291097 (A_SCHOOL_ABSORB)
	shared.NewSpellDataAbsorbOnUse(272437) // Adaptive Combat Assistant - https://www.wowhead.com/forever/spell=1291097
	// on use: 1314412 (A_SCHOOL_ABSORB)
	shared.NewSpellDataAbsorbOnUse(282778) // Mark of the Red Flight - https://www.wowhead.com/forever/spell=1314412

	// Agility
	shared.NewSimpleStatActive(280071) // Flamebreak Talisman - https://www.wowhead.com/forever/spell=1315878

	// Agility / Intellect / Spirit / Stamina / Strength
	shared.NewSimpleStatActive(270226) // Golden Banana - https://www.wowhead.com/forever/spell=1287571

	// Agility / Stamina / Strength
	shared.NewSimpleStatActive(15873) // Ragged John's Neverending Cup - https://www.wowhead.com/forever/spell=20587

	// ArcaneResistance / FireResistance / FrostResistance / NatureResistance / ShadowResistance
	shared.NewSimpleStatActive(15867) // Prismcharm - https://www.wowhead.com/forever/spell=19638
	shared.NewSimpleStatActive(23042) // Loatheb's Reflection - https://www.wowhead.com/forever/spell=28778

	// ArmorNoMods
	shared.NewSimpleStatActive(1447)  // Ring of Saviors - https://www.wowhead.com/forever/spell=18826
	shared.NewSimpleStatActive(11811) // Smoking Heart of the Mountain - https://www.wowhead.com/forever/spell=1300752
	shared.NewSimpleStatActive(12532) // Spire of the Stoneshaper - https://www.wowhead.com/forever/spell=16470
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
	// on use: 1139 (A_MOD_MELEE_RANGED_HASTE_2); not simulated: A_MOD_DECREASE_SPEED
	shared.NewSpellDataAuraOnUse(5079) // Cold Basilisk Eye - https://www.wowhead.com/forever/spell=1139
	// on use: 17275 (A_DAMAGE_SHIELD, A_MOD_RESISTANCE)
	shared.NewSpellDataAuraOnUse(13164) // Heart of the Scale - https://www.wowhead.com/forever/spell=17275
	// on use: 1287840 (A_MOD_SPELL_HIT_CHANCE)
	shared.NewSpellDataAuraOnUse(13968) // Eye of the Beast - https://www.wowhead.com/forever/spell=1287840
	// on use: 23724 (A_ADD_PCT_MODIFIER)
	shared.NewSpellDataAuraOnUse(19340) // Rune of Metamorphosis - https://www.wowhead.com/forever/spell=23724
	// on use: 23726 (A_ADD_FLAT_MODIFIER)
	shared.NewSpellDataAuraOnUse(19342) // Venomous Totem - https://www.wowhead.com/forever/spell=23726
	// on use: 23734 (A_MOD_DAMAGE_PERCENT_DONE, A_MOD_HEALING_DONE_PERCENT, A_MOD_POWER_COST_SCHOOL_PCT)
	shared.NewSpellDataAuraOnUse(19344) // Natural Alignment Crystal - https://www.wowhead.com/forever/spell=23734
	// on use: 24498 (A_MOD_SPELL_CRIT_CHANCE)
	shared.NewSpellDataAuraOnUse(19952) // Gri'lek's Charm of Valor - https://www.wowhead.com/forever/spell=24498
	// on use: 24542 (A_ADD_PCT_MODIFIER, A_ADD_PCT_MODIFIER)
	shared.NewSpellDataAuraOnUse(19955) // Wushoolay's Charm of Nature - https://www.wowhead.com/forever/spell=24542
	// on use: 24499 (A_ADD_PCT_MODIFIER)
	shared.NewSpellDataAuraOnUse(19956) // Wushoolay's Charm of Spirits - https://www.wowhead.com/forever/spell=24499
	// on use: 24543 (A_ADD_FLAT_MODIFIER)
	shared.NewSpellDataAuraOnUse(19957) // Hazza'rah's Charm of Destruction - https://www.wowhead.com/forever/spell=24543
	// on use: 24546 (A_ADD_PCT_MODIFIER, A_ADD_PCT_MODIFIER)
	shared.NewSpellDataAuraOnUse(19958) // Hazza'rah's Charm of Healing - https://www.wowhead.com/forever/spell=24546
	// on use: 24544 (A_ADD_FLAT_MODIFIER, A_ADD_PCT_MODIFIER)
	shared.NewSpellDataAuraOnUse(19959) // Hazza'rah's Charm of Magic - https://www.wowhead.com/forever/spell=24544
	// on use: 26168 (A_DAMAGE_SHIELD)
	shared.NewSpellDataAuraOnUse(21488) // Fetish of Chitinous Spikes - https://www.wowhead.com/forever/spell=26168
	// on use: 26400 (A_MOD_THREAT)
	shared.NewSpellDataAuraOnUse(21647) // Fetish of the Sand Reaver - https://www.wowhead.com/forever/spell=26400
	// on use: 28862 (A_MOD_THREAT)
	shared.NewSpellDataAuraOnUse(23001) // Eye of Diminution - https://www.wowhead.com/forever/spell=28862
	// on use: 1291101 (A_MOD_CRIT_PCT)
	shared.NewSpellDataAuraOnUse(272438) // Weakness Analyzer - https://www.wowhead.com/forever/spell=1291101
	// on use: 1291105 (A_MOD_BLOCK_PERCENT)
	shared.NewSpellDataAuraOnUse(272440) // Defender's Grip Stabilizer - https://www.wowhead.com/forever/spell=1291105
	// on use: 1293820 (A_MOD_DODGE_PERCENT)
	shared.NewSpellDataAuraOnUse(274386) // Toy Soldier - https://www.wowhead.com/forever/spell=1293820
	// on use: 1315844 (A_MOD_DODGE_PERCENT)
	shared.NewSpellDataAuraOnUse(280060) // Lucky Shovelnose Tuft - https://www.wowhead.com/forever/spell=1315844

	// BlockValue
	shared.NewSimpleStatActive(23040) // Glyph of Deflection - https://www.wowhead.com/forever/spell=28773

	// Damage
	// on use: 12257 (A_PERIODIC_TRIGGER_SPELL); not simulated: E_TRIGGER_SPELL; every 1s for 5s: 12278 (E_SCHOOL_DAMAGE)
	shared.NewSpellDataPeriodicDamageOnUse(744) // Thunderbrew's Boot Flask - https://www.wowhead.com/forever/spell=12257
	// on use: 10578 (E_SCHOOL_DAMAGE, A_PERIODIC_DAMAGE)
	shared.NewSpellDataDamageOnUse(8348) // Helm of Fire - https://www.wowhead.com/forever/spell=10578
	// on use: 15712 (E_SCHOOL_DAMAGE); not simulated: E_TRIGGER_SPELL, E_TRIGGER_SPELL
	shared.NewSpellDataDamageOnUse(11905) // Linken's Boomerang - https://www.wowhead.com/forever/spell=15712
	// on use: 17283 (E_SCHOOL_DAMAGE); not simulated: E_NONE
	shared.NewSpellDataDamageOnUse(13171) // Smokey's Lighter - https://www.wowhead.com/forever/spell=17283
	// on use: 18364 (A_PERIODIC_TRIGGER_SPELL); every 5s for 15s: 18374 (E_SCHOOL_DAMAGE)
	shared.NewSpellDataPeriodicDamageOnUse(14134) // Cloak of Fire - https://www.wowhead.com/forever/spell=18364
	// on use: 26789 (E_SCHOOL_DAMAGE)
	shared.NewSpellDataDamageOnUse(21891) // Shard of the Fallen Star - https://www.wowhead.com/forever/spell=26789
	// on use: 1295039 (E_SCHOOL_DAMAGE, A_PERIODIC_DAMAGE); not simulated: E_NONE
	shared.NewSpellDataDamageOnUse(274609) // Serpentbane Crown - https://www.wowhead.com/forever/spell=1295039
	// on use: 1295270 (E_SCHOOL_DAMAGE); not simulated: E_NONE, E_NONE
	shared.NewSpellDataDamageOnUse(274759) // Everlook Pathcarver - https://www.wowhead.com/forever/spell=1295270

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
	// on use: 17712 (E_HEAL); not simulated: E_NONE
	shared.NewSpellDataHealOnUse(833) // Lifestone - https://www.wowhead.com/forever/spell=17712
	// on use: 1317424 (E_HEAL)
	shared.NewSpellDataHealOnUse(1713) // Ankh of Life - https://www.wowhead.com/forever/spell=1317424
	// on use: 10577 (E_HEAL)
	shared.NewSpellDataHealOnUse(8346) // Gauntlets of the Sea - https://www.wowhead.com/forever/spell=10577
	// on use: 18957 (A_PERIODIC_HEAL)
	shared.NewSpellDataHealOnUse(15107) // Orb of Noh'Orahil - https://www.wowhead.com/forever/spell=18957
	// on use: 18957 (A_PERIODIC_HEAL)
	shared.NewSpellDataHealOnUse(15108) // Orb of Dar'Orahil - https://www.wowhead.com/forever/spell=18957
	// on use: 20631 (A_PERIODIC_HEAL)
	shared.NewSpellDataHealOnUse(16768) // Furbolg Medicine Pouch - https://www.wowhead.com/forever/spell=20631
	// on use: 23064 (E_HEAL); not simulated: E_DISPEL_MECHANIC, E_ENERGIZE
	shared.NewSpellDataHealOnUse(18637) // Major Recombobulator - https://www.wowhead.com/forever/spell=23064
	// on use: 24354 (E_HEAL)
	shared.NewSpellDataHealOnUse(19990) // Blessed Prayer Beads - https://www.wowhead.com/forever/spell=24354
	// on use: 363881 (A_PERIODIC_HEAL); not simulated: A_DUMMY
	shared.NewSpellDataHealOnUse(20130) // Diamond Flask - https://www.wowhead.com/forever/spell=363881
	// on use: 1315849 (E_HEAL); not simulated: E_TRIGGER_SPELL_WITH_VALUE
	shared.NewSpellDataHealOnUse(280065) // Behemoth Bones - https://www.wowhead.com/forever/spell=1315849

	// Health
	shared.NewSimpleStatActive(13966) // Mark of Tyranny - https://www.wowhead.com/forever/spell=1287842

	// MP5
	shared.NewSimpleStatActive(4696) // Lapidis Tankard of Tidesippe - https://www.wowhead.com/forever/spell=1135

	// Resources
	// on use: 9174 (E_ENERGIZE)
	shared.NewSpellDataEnergizeOnUse(4264) // Barbaric Belt - https://www.wowhead.com/forever/spell=9174
	// on use: 9253 (E_ENERGIZE)
	shared.NewSpellDataEnergizeOnUse(7515) // Celestial Orb - https://www.wowhead.com/forever/spell=9253
	// on use: 15604 (A_PERIODIC_ENERGIZE)
	shared.NewSpellDataEnergizeOnUse(11819) // Second Wind - https://www.wowhead.com/forever/spell=15604
	// on use: 18385 (E_ENERGIZE)
	shared.NewSpellDataEnergizeOnUse(14152) // Robe of the Archmage - https://www.wowhead.com/forever/spell=18385
	// on use: 1302205 (A_PERIODIC_ENERGIZE, E_TRIGGER_SPELL)
	shared.NewSpellDataEnergizeOnUse(18371) // Mindtap Talisman - https://www.wowhead.com/forever/spell=1302205
	// on use: 24571 (E_ENERGIZE)
	shared.NewSpellDataEnergizeOnUse(19951) // Gri'lek's Charm of Might - https://www.wowhead.com/forever/spell=24571
	// on use: 24532 (E_ENERGIZE)
	shared.NewSpellDataEnergizeOnUse(19954) // Renataki's Charm of Trickery - https://www.wowhead.com/forever/spell=24532
	// on use: 24865 (E_ENERGIZE)
	shared.NewSpellDataEnergizeOnUse(20512) // Sanctified Orb - https://www.wowhead.com/forever/spell=24865
	// on use: 24884 (A_PERIODIC_ENERGIZE)
	shared.NewSpellDataEnergizeOnUse(20525) // Earthen Sigil - https://www.wowhead.com/forever/spell=24884
	// on use: 28760 (E_ENERGIZE)
	shared.NewSpellDataEnergizeOnUse(23027) // Warmth of Forgiveness - https://www.wowhead.com/forever/spell=28760

	// Skipped
	// Not simulated: Lei of Lilies: "Conjure Lily Root" (18831) - ignored effect type 24
	// https://www.wowhead.com/forever/spell=18831
	// Not simulated: Orb of Deception: "Orb of Deception" (16739) - ignored aura type 56
	// https://www.wowhead.com/forever/spell=16739
	// Not simulated: Dog Whistle: "Summon Tracking Hound" (9515) - ignored effect type 28
	// https://www.wowhead.com/forever/spell=9515
	// Not simulated: Ornate Mithril Boots: "Immune Root" (9774) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=9774
	// Not simulated: Glimmering Mithril Insignia: "Fearless" (12733) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=12733
	// Not simulated: Chained Essence of Eranikus: "Cloud of Chained Essence" (12766) - ignored effect type 28
	// https://www.wowhead.com/forever/spell=12766
	// Not simulated: Stealthblade: "Stealth" (1300180) - ignored aura type 16
	// https://www.wowhead.com/forever/spell=1300180
	// Not simulated: Book of the Dead: "Summon Skeleton" (17490) - ignored effect type 28
	// https://www.wowhead.com/forever/spell=17490
	// Not simulated: Cannonball Runner: "Summon Scarlet Cannon" (6251) - ignored effect type 28
	// https://www.wowhead.com/forever/spell=6251
	// Not simulated: Barov Peasant Caller: "Death by Peasant" (18307) - ignored effect type 28
	// https://www.wowhead.com/forever/spell=18307
	// Not simulated: Barov Peasant Caller: "Death by Peasant" (18308) - ignored effect type 28
	// https://www.wowhead.com/forever/spell=18308
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
	// Not simulated: Enamored Water Spirit: "Mana Spring Totem" (24854) - ignored effect type 28
	// https://www.wowhead.com/forever/spell=24854
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
	// Not simulated: Theka's Last Breath: "Theka's Last Breath" (1294617) - ignored aura type 56
	// https://www.wowhead.com/forever/spell=1294617
	// Not simulated: Field Agent Beverage: "Field Agent Beverage" (1297448) - ignored aura type 17
	// https://www.wowhead.com/forever/spell=1297448
	// Not simulated: Zarla's Snack Pouch: "Zarla's Snack Pouch" (1315190) - ignored effect type 28
	// https://www.wowhead.com/forever/spell=1315190
	// Not simulated: Sand-Covered Silk Sash: "Immune Root" (9774) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=9774

	// Speed
	// on use: 13494 (A_MOD_MELEE_HASTE_3)
	shared.NewSpellDataSpeedOnUse(9449) // Manual Crowd Pummeler - https://www.wowhead.com/forever/spell=13494
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

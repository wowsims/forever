package forever

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core/proto"
)

// The sets no hand-written set covers. A bonus is registered when every row it grants parses whole
// onto the set bonus aura; the others are comments. A set written by hand leaves this file on the next
// db run.
func RegisterAllSetBonuses() {
	// The Gladiator
	// https://www.wowhead.com/forever/item-set=1
	shared.NewParsedItemSet(1, "The Gladiator", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// TODO: Manual implementation required
		//
		// Reduces the hit chance of Fear effects against you by 8%.
		// https://www.wowhead.com/forever/spell=1301126
		// unsupported: effect 1 A_MOD_MECHANIC_RESISTANCE misc 5 is not parsed
		// 2: {1301126},

		// +20 Attack Power.
		// https://www.wowhead.com/forever/spell=9331
		3: {9331},

		// TODO: Manual implementation required
		//
		// Chance on hit to gain 5% attack speed for 10s. Attacks while no allies are within 20 yds are 2 times as
		// likely to activate this effect.
		// https://www.wowhead.com/forever/spell=1301123
		// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed, effect 2 A_DUMMY misc 0 is not parsed, effect 3 A_DUMMY misc 0 is not parsed
		// 4: {1301123},

		// Improves your chance to get a critical strike by 1.0%.
		// https://www.wowhead.com/forever/spell=1314795
		5: {1314795},
	})

	// Dal'Rend's Arms
	// https://www.wowhead.com/forever/item-set=41
	shared.NewParsedItemSet(41, "Dal'Rend's Arms", []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand}, map[int32][]int32{
		// +50 Attack Power.
		// https://www.wowhead.com/forever/spell=14056
		2: {14056},
	})

	// TODO: Manual implementation required
	//
	// Spider's Kiss
	// https://www.wowhead.com/forever/item-set=65
	// shared.NewParsedItemSet(65, "Spider's Kiss", []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand}, map[int32][]int32{
	// 	// Chance on Hit: Immobilizes the target and lowers their armor by 100 for 10s.
	// 	// https://www.wowhead.com/forever/spell=17332
	// 	// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
	// 	2: {17332},
	// })

	// The Postmaster
	// https://www.wowhead.com/forever/item-set=81
	shared.NewParsedItemSet(81, "The Postmaster", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet, proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2}, map[int32][]int32{
		// TODO: Manual implementation required
		//
		// Increases run speed by 8%.
		// https://www.wowhead.com/forever/spell=1302376
		// unsupported: effect 1 A_MOD_INCREASE_SPEED misc 0 is not parsed
		// 2: {1302376},

		// Increases damage and healing done by magical spells and effects by up to 23.
		// https://www.wowhead.com/forever/spell=14047
		3: {14047},

		// TODO: Manual implementation required
		//
		// Falling below 25% health grants Return to Sender, reflecting the next spell cast on you within 6s. Can
		// only occur once every 3 min.
		// https://www.wowhead.com/forever/spell=1302380
		// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
		// 4: {1302380},

		// Improves your chance to hit by 1.0%.
		// https://www.wowhead.com/forever/spell=432639
		5: {432639},
	})

	// Necropile Raiment
	// https://www.wowhead.com/forever/item-set=122
	shared.NewParsedItemSet(122, "Necropile Raiment", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotWrist, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Improves your chance to hit by 0.5%.
		// https://www.wowhead.com/forever/spell=1299734
		2: {1299734},

		// TODO: Manual implementation required
		//
		// Harmful spell casts have a chance to steal 28 life from the target enemy.
		// https://www.wowhead.com/forever/spell=1299737
		// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
		// 3: {1299737},

		// +5 All Resistances.
		// https://www.wowhead.com/forever/spell=18676
		4: {18676},
		// Increases damage and healing done by magical spells and effects by up to 23.
		// https://www.wowhead.com/forever/spell=14047
		5: {14047},
	})

	// Bloodmail Regalia
	// https://www.wowhead.com/forever/item-set=123
	shared.NewParsedItemSet(123, "Bloodmail Regalia", []proto.ItemSlot{proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +10 Attack Power.
		// https://www.wowhead.com/forever/spell=9140
		2: {9140},

		// TODO: Manual implementation required
		//
		// Grants your melee attacks a chance to wound the target, causing them to Bleed for 70 damage over 10s.
		// https://www.wowhead.com/forever/spell=1299740
		// unsupported: effect 1 lands on implicit target 6, effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
		// 3: {1299740},

		// +5 All Resistances.
		// https://www.wowhead.com/forever/spell=18676
		4: {18676},
		// Improves your chance to get a critical strike by 1.5%.
		// https://www.wowhead.com/forever/spell=1299738
		5: {1299738},
	})

	// Deathbone Guardian
	// https://www.wowhead.com/forever/item-set=124
	shared.NewParsedItemSet(124, "Deathbone Guardian", []proto.ItemSlot{proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Increased Defense +3.
		// https://www.wowhead.com/forever/spell=7516
		2: {7516},

		// TODO: Manual implementation required
		//
		// Grants a chance when struck in combat to increase damage and healing done by magical spells and effects
		// by up to 70 for 10s.
		// https://www.wowhead.com/forever/spell=1299743
		// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
		// 3: {1299743},

		// +5 All Resistances.
		// https://www.wowhead.com/forever/spell=18676
		4: {18676},

		// TODO: Manual implementation required
		//
		// Reduces the chance for your attacks to be dodged or parried by 2%.
		// https://www.wowhead.com/forever/spell=1213289
		// unsupported: effect 1 A_MOD_COMBAT_RESULT_CHANCE2 misc 2 is not parsed, effect 2 A_MOD_COMBAT_RESULT_CHANCE2 misc 4 is not parsed
		// 5: {1213289},
	})

	// TODO: Manual implementation required
	//
	// Volcanic Armor
	// https://www.wowhead.com/forever/item-set=141
	// shared.NewParsedItemSet(141, "Volcanic Armor", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotLegs}, map[int32][]int32{
	// 	// 5% chance of dealing 40 Fire damage on a successful melee attack.
	// 	// https://www.wowhead.com/forever/spell=9233
	// 	// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
	// 	3: {9233},
	// })

	// Stormshroud Armor
	// https://www.wowhead.com/forever/item-set=142
	shared.NewParsedItemSet(142, "Stormshroud Armor", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs}, map[int32][]int32{
		// TODO: Manual implementation required
		//
		// 10% chance of dealing 20 Nature damage on a successful melee attack.
		// https://www.wowhead.com/forever/spell=18979
		// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
		// 2: {18979},

		// TODO: Manual implementation required
		//
		// 4% chance on melee attack of restoring 30 energy.
		// https://www.wowhead.com/forever/spell=23863
		// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
		// 3: {23863},

		// +14 Attack Power.
		// https://www.wowhead.com/forever/spell=9142
		4: {9142},
	})

	// Devilsaur Armor
	// https://www.wowhead.com/forever/item-set=143
	shared.NewParsedItemSet(143, "Devilsaur Armor", []proto.ItemSlot{proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs}, map[int32][]int32{
		// Improves your chance to hit by 2.0%.
		// https://www.wowhead.com/forever/spell=460230
		2: {460230},
	})

	// Ironfeather Armor
	// https://www.wowhead.com/forever/item-set=144
	shared.NewParsedItemSet(144, "Ironfeather Armor", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest}, map[int32][]int32{
		// Increases damage and healing done by magical spells and effects by up to 20.
		// https://www.wowhead.com/forever/spell=14799
		2: {14799},
	})

	// Defias Leather
	// https://www.wowhead.com/forever/item-set=161
	shared.NewParsedItemSet(161, "Defias Leather", []proto.ItemSlot{proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +5 Arcane Resistance.
		// https://www.wowhead.com/forever/spell=14707
		2: {14707},
		// +15 Attack Power against Humanoids.
		// https://www.wowhead.com/forever/spell=1292025
		3: {1292025},

		// TODO: Manual implementation required
		//
		// Grants your melee attacks a 5% chance when striking from behind to hit a vital point, causing the target
		// to Bleed for 75 Physical damage over 5s.
		// https://www.wowhead.com/forever/spell=1292028
		// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
		// 4: {1292028},

		// Increased Daggers +1.
		// https://www.wowhead.com/forever/spell=7534
		5: {7534},
	})

	// Embrace of the Viper
	// https://www.wowhead.com/forever/item-set=162
	shared.NewParsedItemSet(162, "Embrace of the Viper", []proto.ItemSlot{proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +10 Intellect.
		// https://www.wowhead.com/forever/spell=9106
		2: {9106},
		// +10 Attack Power.
		// https://www.wowhead.com/forever/spell=9140
		3: {9140},

		// TODO: Manual implementation required
		//
		// Taking damage while at or below 25% health causes you to restore 100 health and mana over 5s. This effect
		// can only trigger once every 5 min.
		// https://www.wowhead.com/forever/spell=1291796
		// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
		// 4: {1291796},

		// TODO: Manual implementation required
		//
		// Grants your melee attacks a chance to inflict Dream Venom, stunning the target for 1s.
		// https://www.wowhead.com/forever/spell=1291807
		// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
		// 5: {1291807},
	})

	// Chain of the Scarlet Crusade
	// https://www.wowhead.com/forever/item-set=163
	shared.NewParsedItemSet(163, "Chain of the Scarlet Crusade", []proto.ItemSlot{proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotWrist, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +10 Shadow Resistance.
		// https://www.wowhead.com/forever/spell=14673
		2: {14673},
		// +30 Attack Power against Undead.
		// https://www.wowhead.com/forever/spell=18074
		3: {18074},

		// TODO: Manual implementation required
		//
		// Taking damage has a 15% chance to grant Enraging Light for 10s, causing your melee attacks to deal 20
		// Holy damage. Deals 3 times as much damage to Undead.
		// https://www.wowhead.com/forever/spell=1293674
		// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
		// 4: {1293674},

		// Improves your chance to hit by 1.0%.
		// https://www.wowhead.com/forever/spell=432639
		5: {432639},

		// TODO: Manual implementation required
		//
		// Taking damage while at or below 20% health grants you a shield, absorbing 1 damage for 10s. This effect
		// can only trigger once every 4 min.
		// https://www.wowhead.com/forever/spell=1293670
		// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
		// 6: {1293670},
	})

	// Garb of Thero-shan
	// https://www.wowhead.com/forever/item-set=221
	shared.NewParsedItemSet(221, "Garb of Thero-shan", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotChest}, map[int32][]int32{
		// Improves your chance to get a critical strike with melee attacks by 1.0%.
		// https://www.wowhead.com/forever/spell=7597
		6: {7597},
	})

	// Champion's Battlegear
	// https://www.wowhead.com/forever/item-set=281
	shared.NewParsedItemSet(281, "Champion's Battlegear", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +40 Attack Power.
		// https://www.wowhead.com/forever/spell=14049
		2: {14049},
		// Reduces the cooldown of your Intercept ability by 5.0 sec.
		// https://www.wowhead.com/forever/spell=22738
		4: {22738},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Lieutenant Commander's Battlegear
	// https://www.wowhead.com/forever/item-set=282
	shared.NewParsedItemSet(282, "Lieutenant Commander's Battlegear", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +40 Attack Power.
		// https://www.wowhead.com/forever/spell=14049
		2: {14049},
		// Reduces the cooldown of your Intercept ability by 5.0 sec.
		// https://www.wowhead.com/forever/spell=22738
		4: {22738},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Champion's Earthshaker
	// https://www.wowhead.com/forever/item-set=301
	shared.NewParsedItemSet(301, "Champion's Earthshaker", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +40 Attack Power.
		// https://www.wowhead.com/forever/spell=14049
		2: {14049},
		// Improves your chance to get a critical strike with all Shock spells by 2%.
		// https://www.wowhead.com/forever/spell=22804
		4: {22804},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Lieutenant Commander's Regalia
	// https://www.wowhead.com/forever/item-set=343
	shared.NewParsedItemSet(343, "Lieutenant Commander's Regalia", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Increases damage and healing done by magical spells and effects by up to 23.
		// https://www.wowhead.com/forever/spell=14047
		2: {14047},
		// Reduces the cooldown of your Blink spell by 1.5 sec.
		// https://www.wowhead.com/forever/spell=23025
		4: {23025},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Lieutenant Commander's Vestments
	// https://www.wowhead.com/forever/item-set=348
	shared.NewParsedItemSet(348, "Lieutenant Commander's Vestments", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +40 Attack Power.
		// https://www.wowhead.com/forever/spell=14049
		2: {14049},
		// Reduces the cooldown of your Gouge ability by 1.0 sec.
		// https://www.wowhead.com/forever/spell=23048
		4: {23048},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Champion's Pursuit
	// https://www.wowhead.com/forever/item-set=361
	shared.NewParsedItemSet(361, "Champion's Pursuit", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Agility.
		// https://www.wowhead.com/forever/spell=14384
		2: {14384},
		// Reduces the cooldown of your Concussive Shot by 1.0 sec.
		// https://www.wowhead.com/forever/spell=23158
		4: {23158},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Lieutenant Commander's Pursuit
	// https://www.wowhead.com/forever/item-set=362
	shared.NewParsedItemSet(362, "Lieutenant Commander's Pursuit", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Agility.
		// https://www.wowhead.com/forever/spell=14384
		2: {14384},
		// Reduces the cooldown of your Concussive Shot by 1.0 sec.
		// https://www.wowhead.com/forever/spell=23158
		4: {23158},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Lieutenant Commander's Sanctuary
	// https://www.wowhead.com/forever/item-set=381
	shared.NewParsedItemSet(381, "Lieutenant Commander's Sanctuary", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +40 Attack Power.
		// https://www.wowhead.com/forever/spell=14049
		2: {14049},

		// TODO: Manual implementation required
		//
		// Increases your movement speed by 15% while in Bear, Cat, or Travel Form. Only active outdoors.
		// https://www.wowhead.com/forever/spell=23218
		// unsupported: effect 1 A_MOD_SPEED_ALWAYS misc 0 is not parsed
		// 4: {23218},

		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Champion's Sanctuary
	// https://www.wowhead.com/forever/item-set=382
	shared.NewParsedItemSet(382, "Champion's Sanctuary", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +40 Attack Power.
		// https://www.wowhead.com/forever/spell=14049
		2: {14049},

		// TODO: Manual implementation required
		//
		// Increases your movement speed by 15% while in Bear, Cat, or Travel Form. Only active outdoors.
		// https://www.wowhead.com/forever/spell=23218
		// unsupported: effect 1 A_MOD_SPEED_ALWAYS misc 0 is not parsed
		// 4: {23218},

		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Warlord's Battlegear
	// https://www.wowhead.com/forever/item-set=383
	shared.NewParsedItemSet(383, "Warlord's Battlegear", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		2: {14467},
		// Reduces the cooldown of your Intercept ability by 5.0 sec.
		// https://www.wowhead.com/forever/spell=22738
		3: {22738},
		// +40 Attack Power.
		// https://www.wowhead.com/forever/spell=14049
		6: {14049},
	})

	// Field Marshal's Battlegear
	// https://www.wowhead.com/forever/item-set=384
	shared.NewParsedItemSet(384, "Field Marshal's Battlegear", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		2: {14467},
		// Reduces the cooldown of your Intercept ability by 5.0 sec.
		// https://www.wowhead.com/forever/spell=22738
		3: {22738},
		// +40 Attack Power.
		// https://www.wowhead.com/forever/spell=14049
		6: {14049},
	})

	// Warlord's Earthshaker
	// https://www.wowhead.com/forever/item-set=386
	shared.NewParsedItemSet(386, "Warlord's Earthshaker", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		2: {14467},
		// Improves your chance to get a critical strike with all Shock spells by 2%.
		// https://www.wowhead.com/forever/spell=22804
		3: {22804},
		// +40 Attack Power.
		// https://www.wowhead.com/forever/spell=14049
		6: {14049},
	})

	// Warlord's Regalia
	// https://www.wowhead.com/forever/item-set=387
	shared.NewParsedItemSet(387, "Warlord's Regalia", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		2: {14467},
		// Reduces the cooldown of your Blink spell by 1.5 sec.
		// https://www.wowhead.com/forever/spell=23025
		3: {23025},
		// Increases damage and healing done by magical spells and effects by up to 23.
		// https://www.wowhead.com/forever/spell=14047
		6: {14047},
	})

	// Field Marshal's Regalia
	// https://www.wowhead.com/forever/item-set=388
	shared.NewParsedItemSet(388, "Field Marshal's Regalia", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		2: {14467},
		// Reduces the cooldown of your Blink spell by 1.5 sec.
		// https://www.wowhead.com/forever/spell=23025
		3: {23025},
		// Increases damage and healing done by magical spells and effects by up to 23.
		// https://www.wowhead.com/forever/spell=14047
		6: {14047},
	})

	// Warlord's Vestments
	// https://www.wowhead.com/forever/item-set=393
	shared.NewParsedItemSet(393, "Warlord's Vestments", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		2: {14467},
		// Reduces the cooldown of your Gouge ability by 1.0 sec.
		// https://www.wowhead.com/forever/spell=23048
		3: {23048},
		// +40 Attack Power.
		// https://www.wowhead.com/forever/spell=14049
		6: {14049},
	})

	// Field Marshal's Vestments
	// https://www.wowhead.com/forever/item-set=394
	shared.NewParsedItemSet(394, "Field Marshal's Vestments", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		2: {14467},
		// Reduces the cooldown of your Gouge ability by 1.0 sec.
		// https://www.wowhead.com/forever/spell=23048
		3: {23048},
		// +40 Attack Power.
		// https://www.wowhead.com/forever/spell=14049
		6: {14049},
	})

	// Field Marshal's Pursuit
	// https://www.wowhead.com/forever/item-set=395
	shared.NewParsedItemSet(395, "Field Marshal's Pursuit", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		2: {14467},
		// Reduces the cooldown of your Concussive Shot by 1.0 sec.
		// https://www.wowhead.com/forever/spell=23158
		3: {23158},
		// +20 Agility.
		// https://www.wowhead.com/forever/spell=14384
		6: {14384},
	})

	// Warlord's Pursuit
	// https://www.wowhead.com/forever/item-set=396
	shared.NewParsedItemSet(396, "Warlord's Pursuit", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		2: {14467},
		// Reduces the cooldown of your Concussive Shot by 1.0 sec.
		// https://www.wowhead.com/forever/spell=23158
		3: {23158},
		// +20 Agility.
		// https://www.wowhead.com/forever/spell=14384
		6: {14384},
	})

	// Field Marshal's Sanctuary
	// https://www.wowhead.com/forever/item-set=397
	shared.NewParsedItemSet(397, "Field Marshal's Sanctuary", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		2: {14467},

		// TODO: Manual implementation required
		//
		// Increases your movement speed by 15% while in Bear, Cat, or Travel Form. Only active outdoors.
		// https://www.wowhead.com/forever/spell=23218
		// unsupported: effect 1 A_MOD_SPEED_ALWAYS misc 0 is not parsed
		// 3: {23218},

		// +40 Attack Power.
		// https://www.wowhead.com/forever/spell=14049
		6: {14049},
	})

	// Warlord's Sanctuary
	// https://www.wowhead.com/forever/item-set=398
	shared.NewParsedItemSet(398, "Warlord's Sanctuary", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		2: {14467},

		// TODO: Manual implementation required
		//
		// Increases your movement speed by 15% while in Bear, Cat, or Travel Form. Only active outdoors.
		// https://www.wowhead.com/forever/spell=23218
		// unsupported: effect 1 A_MOD_SPEED_ALWAYS misc 0 is not parsed
		// 3: {23218},

		// +40 Attack Power.
		// https://www.wowhead.com/forever/spell=14049
		6: {14049},
	})

	// Lieutenant Commander's Aegis
	// https://www.wowhead.com/forever/item-set=401
	shared.NewParsedItemSet(401, "Lieutenant Commander's Aegis", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Increases healing done by up to 44 and damage done by up to 15 for all magical spells and effects.
		// https://www.wowhead.com/forever/spell=467550
		2: {467550},
		// Reduces the cooldown of your Hammer of Justice by 10.0 sec.
		// https://www.wowhead.com/forever/spell=23302
		4: {23302},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Field Marshal's Aegis
	// https://www.wowhead.com/forever/item-set=402
	shared.NewParsedItemSet(402, "Field Marshal's Aegis", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		2: {14467},
		// Reduces the cooldown of your Hammer of Justice by 10.0 sec.
		// https://www.wowhead.com/forever/spell=23302
		3: {23302},
		// Increases healing done by up to 44 and damage done by up to 15 for all magical spells and effects.
		// https://www.wowhead.com/forever/spell=467550
		6: {467550},
	})

	// Bloodvine Garb
	// https://www.wowhead.com/forever/item-set=421
	shared.NewParsedItemSet(421, "Bloodvine Garb", []proto.ItemSlot{proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Improves your chance to get a critical strike with spells by 2.0%.
		// https://www.wowhead.com/forever/spell=18382
		3: {18382},
	})

	// TODO: Manual implementation required
	//
	// Primal Batskin
	// https://www.wowhead.com/forever/item-set=441
	// shared.NewParsedItemSet(441, "Primal Batskin", []proto.ItemSlot{proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotWrist, proto.ItemSlot_ItemSlotHands}, map[int32][]int32{
	// 	// Minor increase to running and swimming speed. Does not stack with similar effects.
	// 	// https://www.wowhead.com/forever/spell=24090
	// 	// unsupported: effect 1 A_MOD_INCREASE_SPEED misc 0 is not parsed, effect 2 A_MOD_INCREASE_SWIM_SPEED misc 0 is not parsed
	// 	3: {24090},
	// })

	// Blood Tiger Harness
	// https://www.wowhead.com/forever/item-set=442
	shared.NewParsedItemSet(442, "Blood Tiger Harness", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest}, map[int32][]int32{
		// Improves your chance to get a critical strike with melee attacks by 1.0%.
		// https://www.wowhead.com/forever/spell=7597
		// Improves your chance to get a critical strike with spells by 1.0%.
		// https://www.wowhead.com/forever/spell=18384
		2: {7597, 18384},
	})

	// Bloodsoul Embrace
	// https://www.wowhead.com/forever/item-set=443
	shared.NewParsedItemSet(443, "Bloodsoul Embrace", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands}, map[int32][]int32{
		// Restores 12 mana per 5 sec.
		// https://www.wowhead.com/forever/spell=21636
		3: {21636},
	})

	// The Darksoul
	// https://www.wowhead.com/forever/item-set=444
	shared.NewParsedItemSet(444, "The Darksoul", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotLegs}, map[int32][]int32{
		// Increased Defense +20.
		// https://www.wowhead.com/forever/spell=21416
		3: {21416},
	})

	// The Highlander's Resolution
	// https://www.wowhead.com/forever/item-set=467
	shared.NewParsedItemSet(467, "The Highlander's Resolution", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +5 Stamina.
		// https://www.wowhead.com/forever/spell=7503
		2: {7503},
		// Improves your chance to get a critical strike with melee attacks by 1.0%.
		// https://www.wowhead.com/forever/spell=7597
		3: {7597},
	})

	// The Highlander's Resolve
	// https://www.wowhead.com/forever/item-set=468
	shared.NewParsedItemSet(468, "The Highlander's Resolve", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +5 Stamina.
		// https://www.wowhead.com/forever/spell=7503
		2: {7503},
		// Improves your chance to get a critical strike with melee attacks by 1.0%.
		// https://www.wowhead.com/forever/spell=7597
		3: {7597},
	})

	// The Highlander's Determination
	// https://www.wowhead.com/forever/item-set=469
	shared.NewParsedItemSet(469, "The Highlander's Determination", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +5 Stamina.
		// https://www.wowhead.com/forever/spell=7503
		2: {7503},
		// Improves your chance to get a critical strike with melee attacks by 1.0%.
		// https://www.wowhead.com/forever/spell=7597
		3: {7597},
	})

	// The Highlander's Fortitude
	// https://www.wowhead.com/forever/item-set=470
	shared.NewParsedItemSet(470, "The Highlander's Fortitude", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +5 Stamina.
		// https://www.wowhead.com/forever/spell=7503
		2: {7503},
		// Improves your chance to get a critical strike with spells by 1.0%.
		// https://www.wowhead.com/forever/spell=18384
		3: {18384},
	})

	// The Highlander's Will
	// https://www.wowhead.com/forever/item-set=472
	shared.NewParsedItemSet(472, "The Highlander's Will", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +5 Stamina.
		// https://www.wowhead.com/forever/spell=7503
		2: {7503},
		// Improves your chance to get a critical strike with spells by 1.0%.
		// https://www.wowhead.com/forever/spell=18384
		3: {18384},
	})

	// The Highlander's Intent
	// https://www.wowhead.com/forever/item-set=473
	shared.NewParsedItemSet(473, "The Highlander's Intent", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +5 Stamina.
		// https://www.wowhead.com/forever/spell=7503
		2: {7503},
		// Improves your chance to get a critical strike with spells by 1.0%.
		// https://www.wowhead.com/forever/spell=18384
		3: {18384},
	})

	// Vindicator's Battlegear
	// https://www.wowhead.com/forever/item-set=474
	shared.NewParsedItemSet(474, "Vindicator's Battlegear", []proto.ItemSlot{proto.ItemSlot_ItemSlotNeck, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotWrist, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotTrinket1, proto.ItemSlot_ItemSlotTrinket2}, map[int32][]int32{
		// Increases your chance to block attacks with a shield by 2%.
		// https://www.wowhead.com/forever/spell=13675
		2: {13675},
		// Decreases the cooldown of Intimidating Shout by 15.0 sec.
		// https://www.wowhead.com/forever/spell=24456
		3: {24456},
		// Decrease the rage cost of Whirlwind by 3.0.
		// https://www.wowhead.com/forever/spell=24431
		5: {24431},
	})

	// Freethinker's Armor
	// https://www.wowhead.com/forever/item-set=475
	shared.NewParsedItemSet(475, "Freethinker's Armor", []proto.ItemSlot{proto.ItemSlot_ItemSlotNeck, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotWrist, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotTrinket1, proto.ItemSlot_ItemSlotTrinket2}, map[int32][]int32{
		// Restores 4 mana per 5 sec.
		// https://www.wowhead.com/forever/spell=21618
		2: {21618},
		// Reduces the casting time of your Holy Light spell by 0.1 sec.
		// https://www.wowhead.com/forever/spell=24457
		3: {24457},
		// Increases the duration of all Blessings by 10%.
		// https://www.wowhead.com/forever/spell=24460
		5: {24460},
	})

	// Augur's Regalia
	// https://www.wowhead.com/forever/item-set=476
	shared.NewParsedItemSet(476, "Augur's Regalia", []proto.ItemSlot{proto.ItemSlot_ItemSlotNeck, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotWrist, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotTrinket1, proto.ItemSlot_ItemSlotTrinket2}, map[int32][]int32{
		// Restores 4 mana per 5 sec.
		// https://www.wowhead.com/forever/spell=21618
		2: {21618},
		// Improves the duration of your Frost Shock spell by 1.0 sec.
		// https://www.wowhead.com/forever/spell=24461
		3: {24461},
		// Increase the range of your Lightning Bolt spell by 5 yds.
		// https://www.wowhead.com/forever/spell=24462
		5: {24462},
	})

	// Predator's Armor
	// https://www.wowhead.com/forever/item-set=477
	shared.NewParsedItemSet(477, "Predator's Armor", []proto.ItemSlot{proto.ItemSlot_ItemSlotNeck, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotWrist, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotTrinket1, proto.ItemSlot_ItemSlotTrinket2}, map[int32][]int32{
		// +20 Attack Power.
		// https://www.wowhead.com/forever/spell=9331
		2: {9331},
		// Decreases the cooldown of Concussive Shot by 1.0 sec.
		// https://www.wowhead.com/forever/spell=24465
		3: {24465},
		// Increases the duration of Serpent Sting by 3.0 sec.
		// https://www.wowhead.com/forever/spell=24467
		5: {24467},
	})

	// Madcap's Outfit
	// https://www.wowhead.com/forever/item-set=478
	shared.NewParsedItemSet(478, "Madcap's Outfit", []proto.ItemSlot{proto.ItemSlot_ItemSlotNeck, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotWrist, proto.ItemSlot_ItemSlotTrinket1, proto.ItemSlot_ItemSlotTrinket2}, map[int32][]int32{
		// +20 Attack Power.
		// https://www.wowhead.com/forever/spell=9331
		2: {9331},
		// Decreases the cooldown of Blind by 20.0 sec.
		// https://www.wowhead.com/forever/spell=24469
		3: {24469},
		// Decrease the energy cost of Eviscerate and Rupture by 5.
		// https://www.wowhead.com/forever/spell=24471
		5: {24471},
	})

	// The Defiler's Determination
	// https://www.wowhead.com/forever/item-set=483
	shared.NewParsedItemSet(483, "The Defiler's Determination", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +5 Stamina.
		// https://www.wowhead.com/forever/spell=7503
		2: {7503},
		// Improves your chance to get a critical strike with melee attacks by 1.0%.
		// https://www.wowhead.com/forever/spell=7597
		3: {7597},
	})

	// The Defiler's Fortitude
	// https://www.wowhead.com/forever/item-set=484
	shared.NewParsedItemSet(484, "The Defiler's Fortitude", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +5 Stamina.
		// https://www.wowhead.com/forever/spell=7503
		2: {7503},
		// Improves your chance to get a critical strike with melee attacks by 1.0%.
		// https://www.wowhead.com/forever/spell=7597
		3: {7597},
	})

	// The Defiler's Intent
	// https://www.wowhead.com/forever/item-set=485
	shared.NewParsedItemSet(485, "The Defiler's Intent", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +5 Stamina.
		// https://www.wowhead.com/forever/spell=7503
		2: {7503},
		// Improves your chance to get a critical strike with spells by 1.0%.
		// https://www.wowhead.com/forever/spell=18384
		3: {18384},
	})

	// The Defiler's Resolution
	// https://www.wowhead.com/forever/item-set=487
	shared.NewParsedItemSet(487, "The Defiler's Resolution", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +5 Stamina.
		// https://www.wowhead.com/forever/spell=7503
		2: {7503},
		// Improves your chance to get a critical strike with melee attacks by 1.0%.
		// https://www.wowhead.com/forever/spell=7597
		3: {7597},
	})

	// The Defiler's Will
	// https://www.wowhead.com/forever/item-set=488
	shared.NewParsedItemSet(488, "The Defiler's Will", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +5 Stamina.
		// https://www.wowhead.com/forever/spell=7503
		2: {7503},
		// Improves your chance to get a critical strike with spells by 1.0%.
		// https://www.wowhead.com/forever/spell=18384
		3: {18384},
	})

	// Black Dragon Mail
	// https://www.wowhead.com/forever/item-set=489
	shared.NewParsedItemSet(489, "Black Dragon Mail", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotLegs}, map[int32][]int32{
		// Improves your chance to hit by 1.0%.
		// https://www.wowhead.com/forever/spell=15464
		2: {15464},
		// Improves your chance to get a critical strike with melee attacks by 2.0%.
		// https://www.wowhead.com/forever/spell=7598
		3: {7598},
	})

	// Green Dragon Mail
	// https://www.wowhead.com/forever/item-set=490
	shared.NewParsedItemSet(490, "Green Dragon Mail", []proto.ItemSlot{proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs}, map[int32][]int32{
		// Restores 3 mana per 5 sec.
		// https://www.wowhead.com/forever/spell=21625
		2: {21625},
		// Allows 15% of your Mana regeneration to continue while casting.
		// https://www.wowhead.com/forever/spell=21894
		3: {21894},
	})

	// Blue Dragon Mail
	// https://www.wowhead.com/forever/item-set=491
	shared.NewParsedItemSet(491, "Blue Dragon Mail", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotLegs}, map[int32][]int32{
		// +4 All Resistances.
		// https://www.wowhead.com/forever/spell=18675
		2: {18675},
		// Increases damage and healing done by magical spells and effects by up to 28.
		// https://www.wowhead.com/forever/spell=14127
		3: {14127},
	})

	// TODO: Manual implementation required
	//
	// Twilight Trappings
	// https://www.wowhead.com/forever/item-set=492
	// shared.NewParsedItemSet(492, "Twilight Trappings", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest}, map[int32][]int32{
	// 	// Bestows the wearer with the evil aura of a Twilight's Hammer cultist.
	// 	// https://www.wowhead.com/forever/spell=24746
	// 	// unsupported: effect 1 A_DUMMY misc 0 is not parsed
	// 	3: {24746},
	// })

	// Genesis Raiment
	// https://www.wowhead.com/forever/item-set=493
	shared.NewParsedItemSet(493, "Genesis Raiment", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Increased Defense +15.
		// https://www.wowhead.com/forever/spell=18196
		// +150 Armor.
		// https://www.wowhead.com/forever/spell=15693
		3: {18196, 15693},
		// Reduces the cooldown of Rebirth by 10.0 minutes.
		// https://www.wowhead.com/forever/spell=26106
		5: {26106},
	})

	// TODO: Manual implementation required
	//
	// Symbols of Unending Life
	// https://www.wowhead.com/forever/item-set=494
	// shared.NewParsedItemSet(494, "Symbols of Unending Life", []proto.ItemSlot{proto.ItemSlot_ItemSlotBack, proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2, proto.ItemSlot_ItemSlotMainHand}, map[int32][]int32{
	// 	// Your finishing moves now refund 30 energy on a Miss, Dodge, Block, or Parry.
	// 	// https://www.wowhead.com/forever/spell=26107
	// 	// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
	// 	3: {26107},
	// })

	// Battlegear of Unyielding Strength
	// https://www.wowhead.com/forever/item-set=495
	shared.NewParsedItemSet(495, "Battlegear of Unyielding Strength", []proto.ItemSlot{proto.ItemSlot_ItemSlotBack, proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2, proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand}, map[int32][]int32{
		// rage cost to Intercept.
		// https://www.wowhead.com/forever/spell=26111
		3: {26111},
	})

	// Deathdealer's Embrace
	// https://www.wowhead.com/forever/item-set=497
	shared.NewParsedItemSet(497, "Deathdealer's Embrace", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Reduces the cooldown of your Evasion ability by min.
		// https://www.wowhead.com/forever/spell=26112
		3: {26112},
		// 15% increased damage to your Eviscerate ability.
		// https://www.wowhead.com/forever/spell=26113
		5: {26113},
	})

	// Emblems of Veiled Shadows
	// https://www.wowhead.com/forever/item-set=498
	shared.NewParsedItemSet(498, "Emblems of Veiled Shadows", []proto.ItemSlot{proto.ItemSlot_ItemSlotBack, proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2, proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand}, map[int32][]int32{
		// energy cost for your Slice and Dice ability.
		// https://www.wowhead.com/forever/spell=26114
		3: {26114},
	})

	// Stormcaller's Garb
	// https://www.wowhead.com/forever/item-set=501
	shared.NewParsedItemSet(501, "Stormcaller's Garb", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// TODO: Manual implementation required
		//
		// Your Lightning Bolt, Chain Lightning, and Shock spells have a 20% chance to grant up to 50 Nature damage
		// to spells for 8s.
		// https://www.wowhead.com/forever/spell=26119
		// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
		// 3: {26119},

		// seconds on the casting time of your Chain Heal spell.
		// https://www.wowhead.com/forever/spell=26122
		5: {26122},
	})

	// TODO: Manual implementation required
	//
	// Gift of the Gathering Storm
	// https://www.wowhead.com/forever/item-set=502
	// shared.NewParsedItemSet(502, "Gift of the Gathering Storm", []proto.ItemSlot{proto.ItemSlot_ItemSlotBack, proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2, proto.ItemSlot_ItemSlotMainHand}, map[int32][]int32{
	// 	// Increases the chain target damage multiplier of your Chain Lightning spell by 5%.
	// 	// https://www.wowhead.com/forever/spell=26123
	// 	// unsupported: effect 1 A_ADD_PCT_MODIFIER misc 20 is not parsed
	// 	3: {26123},
	// })

	// Trappings of Vaulted Secrets
	// https://www.wowhead.com/forever/item-set=504
	shared.NewParsedItemSet(504, "Trappings of Vaulted Secrets", []proto.ItemSlot{proto.ItemSlot_ItemSlotBack, proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2, proto.ItemSlot_ItemSlotMainHand}, map[int32][]int32{
		// 15% increase to the total damage absorbed by Mana Shield.
		// https://www.wowhead.com/forever/spell=26131
		3: {26131},
	})

	// Avenger's Battlegear
	// https://www.wowhead.com/forever/item-set=505
	shared.NewParsedItemSet(505, "Avenger's Battlegear", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Increases the duration of your Judgements by 20%.
		// https://www.wowhead.com/forever/spell=26130
		3: {26130},
		// Increases damage and healing done by magical spells and effects by up to 71.
		// https://www.wowhead.com/forever/spell=23929
		5: {23929},
	})

	// TODO: Manual implementation required
	//
	// Battlegear of Eternal Justice
	// https://www.wowhead.com/forever/item-set=506
	// shared.NewParsedItemSet(506, "Battlegear of Eternal Justice", []proto.ItemSlot{proto.ItemSlot_ItemSlotBack, proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2, proto.ItemSlot_ItemSlotMainHand}, map[int32][]int32{
	// 	// 20% chance to regain 100 mana when you cast a Judgement.
	// 	// https://www.wowhead.com/forever/spell=26135
	// 	// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
	// 	3: {26135},
	// })

	// TODO: Manual implementation required
	//
	// Trappings of the Unseen Path
	// https://www.wowhead.com/forever/item-set=510
	// shared.NewParsedItemSet(510, "Trappings of the Unseen Path", []proto.ItemSlot{proto.ItemSlot_ItemSlotBack, proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2, proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand}, map[int32][]int32{
	// 	// Increases your pet's damage by 3%.
	// 	// https://www.wowhead.com/forever/spell=26176
	// 	// unsupported: effect 1 A_PERIODIC_TRIGGER_SPELL misc 0 is not parsed
	// 	3: {26176},
	// })

	// Ironweave Battlesuit
	// https://www.wowhead.com/forever/item-set=520
	shared.NewParsedItemSet(520, "Ironweave Battlesuit", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotWrist, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +200 Armor.
		// https://www.wowhead.com/forever/spell=14803
		2: {14803},
		// Decreases the magical resistances of your spell targets by 5.
		// https://www.wowhead.com/forever/spell=26283
		3: {26283},

		// TODO: Manual implementation required
		//
		// Reduces the hit chance of Silence and Interrupt effects against you by 10%.
		// https://www.wowhead.com/forever/spell=27733
		// unsupported: effect 1 A_MOD_MECHANIC_RESISTANCE misc 26 is not parsed, effect 2 A_MOD_MECHANIC_RESISTANCE misc 9 is not parsed
		// 4: {27733},

		// Increases damage and healing done by magical spells and effects by up to 23.
		// https://www.wowhead.com/forever/spell=14047
		5: {14047},

		// TODO: Manual implementation required
		//
		// The Ironweave hardens under duress, reducing damage taken while Stunned by 15%.
		// https://www.wowhead.com/forever/spell=1302374
		// unsupported: effect 1 A_MOD_MECHANIC_DAMAGE_TAKEN_PERCENT misc 12 is not parsed
		// 6: {1302374},
	})

	// Dreamwalker Raiment
	// https://www.wowhead.com/forever/item-set=521
	shared.NewParsedItemSet(521, "Dreamwalker Raiment", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotWrist, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet, proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2}, map[int32][]int32{
		// TODO: Manual implementation required
		//
		// Your Rejuvenation ticks have a chance to restore 60 mana, 8 energy, or 2 rage to your target.
		// https://www.wowhead.com/forever/spell=28716
		// unsupported: effect 1 A_OVERRIDE_CLASS_SCRIPTS misc 4533 is not parsed
		// 2: {28716},

		// Reduces the mana cost of your Healing Touch, Regrowth, Rejuvenation, and Tranquility spells by 3%.
		// https://www.wowhead.com/forever/spell=28743
		4: {28743},

		// TODO: Manual implementation required
		//
		// Your initial cast and Regrowth ticks will increase the maximum health of your target by up to 50, stacking
		// up to 7 times.
		// https://www.wowhead.com/forever/spell=28744
		// unsupported: effect 1 A_OVERRIDE_CLASS_SCRIPTS misc 4537 is not parsed
		// 6: {28744},

		// TODO: Manual implementation required
		//
		// On Healing Touch critical hits, you regain 30% of the mana cost of the spell.
		// https://www.wowhead.com/forever/spell=28719
		// unsupported: effect 1 A_DUMMY misc 0 is not parsed
		// 8: {28719},
	})

	// Champion's Vestments
	// https://www.wowhead.com/forever/item-set=522
	shared.NewParsedItemSet(522, "Champion's Vestments", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +40 Attack Power.
		// https://www.wowhead.com/forever/spell=14049
		2: {14049},
		// Reduces the cooldown of your Gouge ability by 1.0 sec.
		// https://www.wowhead.com/forever/spell=23048
		4: {23048},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Bonescythe Armor
	// https://www.wowhead.com/forever/item-set=524
	shared.NewParsedItemSet(524, "Bonescythe Armor", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotWrist, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet, proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2}, map[int32][]int32{
		// TODO: Manual implementation required
		//
		// Your normal melee swings have a chance to Invigorate you, healing you for 100.
		// https://www.wowhead.com/forever/spell=28816
		// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
		// 2: {28816},

		// TODO: Manual implementation required
		//
		// Your Backstab, Sinister Strike, and Hemorrhage critical hits cause you to regain 5 energy.
		// https://www.wowhead.com/forever/spell=28812
		// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
		// 4: {28812},

		// Reduces the threat from your Backstab, Sinister Strike, Hemorrhage, and Eviscerate abilities.
		// https://www.wowhead.com/forever/spell=28811
		6: {28811},

		// TODO: Manual implementation required
		//
		// Your Eviscerate has a chance per combo point to reveal a flaw in your opponent's armor, granting a 100%
		// critical hit chance for your next Backstab, Sinister Strike, or Hemorrhage.
		// https://www.wowhead.com/forever/spell=28814
		// unsupported: effect 1 A_ADD_TARGET_TRIGGER misc 0 is not parsed
		// 8: {28814},
	})

	// The Earthshatterer
	// https://www.wowhead.com/forever/item-set=527
	shared.NewParsedItemSet(527, "The Earthshatterer", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotWrist, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet, proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2}, map[int32][]int32{
		// Reduces the mana cost of your totem spells by 12%.
		// https://www.wowhead.com/forever/spell=28818
		2: {28818},
		// Increases the mana gained from your Mana Spring totems by 25%.
		// https://www.wowhead.com/forever/spell=29171
		4: {29171},

		// TODO: Manual implementation required
		//
		// Your Healing Wave and Lesser Healing Wave spells have a chance to imbue your target with Totemic Power.
		// https://www.wowhead.com/forever/spell=28823
		// unsupported: effect 1 A_DUMMY misc 0 is not parsed
		// 6: {28823},

		// TODO: Manual implementation required
		//
		// Your Lightning Shield spell also grants you 15 mana per 5 sec. while active.
		// https://www.wowhead.com/forever/spell=28821
		// unsupported: effect 1 A_ADD_TARGET_TRIGGER misc 0 is not parsed
		// 8: {28821},
	})

	// Redemption Armor
	// https://www.wowhead.com/forever/item-set=528
	shared.NewParsedItemSet(528, "Redemption Armor", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotWrist, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet, proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2}, map[int32][]int32{
		// TODO: Manual implementation required
		//
		// Increases the amount healed by your Judgement of Light by 20.
		// https://www.wowhead.com/forever/spell=28775
		// unsupported: effect 1 A_DUMMY misc 0 is not parsed
		// 2: {28775},

		// Reduces cooldown on your Lay on Hands by 12.0 min.
		// https://www.wowhead.com/forever/spell=28774
		4: {28774},

		// TODO: Manual implementation required
		//
		// Your Flash of Light and Holy Light spells have a chance to imbue your target with Holy Power.
		// https://www.wowhead.com/forever/spell=28789
		// unsupported: effect 1 A_DUMMY misc 0 is not parsed
		// 6: {28789},

		// TODO: Manual implementation required
		//
		// Your Cleanse spell also heals the target for 200.
		// https://www.wowhead.com/forever/spell=28787
		// unsupported: effect 1 A_ADD_TARGET_TRIGGER misc 0 is not parsed
		// 8: {28787},
	})

	// Battlegear of Undead Slaying
	// https://www.wowhead.com/forever/item-set=533
	shared.NewParsedItemSet(533, "Battlegear of Undead Slaying", []proto.ItemSlot{proto.ItemSlot_ItemSlotHands}, map[int32][]int32{
		// Increases your damage against undead by 2%.
		// https://www.wowhead.com/forever/spell=29068
		3: {29068},
	})

	// Undead Slayer's Armor
	// https://www.wowhead.com/forever/item-set=534
	shared.NewParsedItemSet(534, "Undead Slayer's Armor", []proto.ItemSlot{proto.ItemSlot_ItemSlotHands}, map[int32][]int32{
		// Increases your damage against undead by 2%.
		// https://www.wowhead.com/forever/spell=29068
		3: {29068},
	})

	// Garb of the Undead Slayer
	// https://www.wowhead.com/forever/item-set=535
	shared.NewParsedItemSet(535, "Garb of the Undead Slayer", []proto.ItemSlot{proto.ItemSlot_ItemSlotHands}, map[int32][]int32{
		// Increases your damage against undead by 2%.
		// https://www.wowhead.com/forever/spell=29068
		3: {29068},
	})

	// Regalia of Undead Cleansing
	// https://www.wowhead.com/forever/item-set=536
	shared.NewParsedItemSet(536, "Regalia of Undead Cleansing", []proto.ItemSlot{proto.ItemSlot_ItemSlotHands}, map[int32][]int32{
		// Increases your damage against undead by 2%.
		// https://www.wowhead.com/forever/spell=29068
		3: {29068},
	})

	// Champion's Battlearmor
	// https://www.wowhead.com/forever/item-set=537
	shared.NewParsedItemSet(537, "Champion's Battlearmor", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +40 Attack Power.
		// https://www.wowhead.com/forever/spell=14049
		2: {14049},
		// Reduces the cooldown of your Intercept ability by 5.0 sec.
		// https://www.wowhead.com/forever/spell=22738
		4: {22738},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Champion's Stormcaller
	// https://www.wowhead.com/forever/item-set=538
	shared.NewParsedItemSet(538, "Champion's Stormcaller", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +40 Attack Power.
		// https://www.wowhead.com/forever/spell=14049
		2: {14049},
		// Improves your chance to get a critical strike with all Shock spells by 2%.
		// https://www.wowhead.com/forever/spell=22804
		4: {22804},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Champion's Arcanum
	// https://www.wowhead.com/forever/item-set=542
	shared.NewParsedItemSet(542, "Champion's Arcanum", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Increases damage and healing done by magical spells and effects by up to 23.
		// https://www.wowhead.com/forever/spell=14047
		2: {14047},
		// Reduces the cooldown of your Blink spell by 1.5 sec.
		// https://www.wowhead.com/forever/spell=23025
		4: {23025},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Champion's Pursuance
	// https://www.wowhead.com/forever/item-set=543
	shared.NewParsedItemSet(543, "Champion's Pursuance", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Agility.
		// https://www.wowhead.com/forever/spell=14384
		2: {14384},
		// Reduces the cooldown of your Concussive Shot by 1.0 sec.
		// https://www.wowhead.com/forever/spell=23158
		4: {23158},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Lieutenant Commander's Redoubt
	// https://www.wowhead.com/forever/item-set=544
	shared.NewParsedItemSet(544, "Lieutenant Commander's Redoubt", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Increases damage and healing done by magical spells and effects by up to 23.
		// https://www.wowhead.com/forever/spell=14047
		2: {14047},
		// Reduces the cooldown of your Hammer of Justice by 10.0 sec.
		// https://www.wowhead.com/forever/spell=23302
		4: {23302},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Lieutenant Commander's Battlearmor
	// https://www.wowhead.com/forever/item-set=545
	shared.NewParsedItemSet(545, "Lieutenant Commander's Battlearmor", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +40 Attack Power.
		// https://www.wowhead.com/forever/spell=14049
		2: {14049},
		// Reduces the cooldown of your Intercept ability by 5.0 sec.
		// https://www.wowhead.com/forever/spell=22738
		4: {22738},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Champion's Regalia
	// https://www.wowhead.com/forever/item-set=546
	shared.NewParsedItemSet(546, "Champion's Regalia", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Increases damage and healing done by magical spells and effects by up to 23.
		// https://www.wowhead.com/forever/spell=14047
		2: {14047},
		// Reduces the cooldown of your Blink spell by 1.5 sec.
		// https://www.wowhead.com/forever/spell=23025
		4: {23025},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Lieutenant Commander's Guard
	// https://www.wowhead.com/forever/item-set=548
	shared.NewParsedItemSet(548, "Lieutenant Commander's Guard", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +40 Attack Power.
		// https://www.wowhead.com/forever/spell=14049
		2: {14049},
		// Reduces the cooldown of your Gouge ability by 1.0 sec.
		// https://www.wowhead.com/forever/spell=23048
		4: {23048},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Lieutenant Commander's Pursuance
	// https://www.wowhead.com/forever/item-set=550
	shared.NewParsedItemSet(550, "Lieutenant Commander's Pursuance", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Agility.
		// https://www.wowhead.com/forever/spell=14384
		2: {14384},
		// Reduces the cooldown of your Concussive Shot by 1.0 sec.
		// https://www.wowhead.com/forever/spell=23158
		4: {23158},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Stormcloth Regalia
	// https://www.wowhead.com/forever/item-set=1972
	shared.NewParsedItemSet(1972, "Stormcloth Regalia", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +10 Spirit.
		// https://www.wowhead.com/forever/spell=1258496
		3: {1258496},
		// Increases healing done by spells and effects by up to 22 and damage done by spells and effects by up to
		// 8.
		// https://www.wowhead.com/forever/spell=1258494
		6: {1258494},
	})

	// TODO: Manual implementation required
	//
	// Teachings of the Furbolgs
	// https://www.wowhead.com/forever/item-set=2071
	// shared.NewParsedItemSet(2071, "Teachings of the Furbolgs", []proto.ItemSlot{proto.ItemSlot_ItemSlotTrinket1, proto.ItemSlot_ItemSlotTrinket2}, map[int32][]int32{
	// 	// Your attacks have a chance to increase your Attack Power by 40 and your healing spells have a chance to
	// 	// increase your Intellect by 20 for 15s.
	// 	// https://www.wowhead.com/forever/spell=1294087
	// 	// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
	// 	2: {1294087},
	// })

	// Champion's Aegis
	// https://www.wowhead.com/forever/item-set=2082
	shared.NewParsedItemSet(2082, "Champion's Aegis", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Increases healing done by up to 44 and damage done by up to 15 for all magical spells and effects.
		// https://www.wowhead.com/forever/spell=467550
		2: {467550},
		// Reduces the cooldown of your Hammer of Justice by 10.0 sec.
		// https://www.wowhead.com/forever/spell=23302
		4: {23302},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Warlord's Aegis
	// https://www.wowhead.com/forever/item-set=2083
	shared.NewParsedItemSet(2083, "Warlord's Aegis", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		2: {14467},
		// Reduces the cooldown of your Hammer of Justice by 10.0 sec.
		// https://www.wowhead.com/forever/spell=23302
		3: {23302},
		// Increases healing done by up to 44 and damage done by up to 15 for all magical spells and effects.
		// https://www.wowhead.com/forever/spell=467550
		6: {467550},
	})

	// Champion's Vindication
	// https://www.wowhead.com/forever/item-set=2084
	shared.NewParsedItemSet(2084, "Champion's Vindication", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Increases damage and healing done by magical spells and effects by up to 23.
		// https://www.wowhead.com/forever/spell=14047
		2: {14047},
		// Reduces the cooldown of your Hammer of Justice by 10.0 sec.
		// https://www.wowhead.com/forever/spell=23302
		4: {23302},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Warlord's Vindication
	// https://www.wowhead.com/forever/item-set=2085
	shared.NewParsedItemSet(2085, "Warlord's Vindication", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		2: {14467},
		// Reduces the cooldown of your Hammer of Justice by 10.0 sec.
		// https://www.wowhead.com/forever/spell=23302
		3: {23302},
		// Increases damage and healing done by magical spells and effects by up to 23.
		// https://www.wowhead.com/forever/spell=14047
		6: {14047},
	})

	// Lieutenant Commander's Vindication
	// https://www.wowhead.com/forever/item-set=2086
	shared.NewParsedItemSet(2086, "Lieutenant Commander's Vindication", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Increases damage and healing done by magical spells and effects by up to 23.
		// https://www.wowhead.com/forever/spell=14047
		2: {14047},
		// Reduces the cooldown of your Hammer of Justice by 10.0 sec.
		// https://www.wowhead.com/forever/spell=23302
		4: {23302},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Field Marshal's Vindication
	// https://www.wowhead.com/forever/item-set=2087
	shared.NewParsedItemSet(2087, "Field Marshal's Vindication", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		2: {14467},
		// Reduces the cooldown of your Hammer of Justice by 10.0 sec.
		// https://www.wowhead.com/forever/spell=23302
		3: {23302},
		// Increases damage and healing done by magical spells and effects by up to 23.
		// https://www.wowhead.com/forever/spell=14047
		6: {14047},
	})

	// Lieutenant Commander's Earthshaker
	// https://www.wowhead.com/forever/item-set=2088
	shared.NewParsedItemSet(2088, "Lieutenant Commander's Earthshaker", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +40 Attack Power.
		// https://www.wowhead.com/forever/spell=14049
		2: {14049},
		// Improves your chance to get a critical strike with all Shock spells by 2%.
		// https://www.wowhead.com/forever/spell=22804
		4: {22804},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Field Marshal's Earthshaker
	// https://www.wowhead.com/forever/item-set=2089
	shared.NewParsedItemSet(2089, "Field Marshal's Earthshaker", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		2: {14467},
		// Improves your chance to get a critical strike with all Shock spells by 2%.
		// https://www.wowhead.com/forever/spell=22804
		3: {22804},
		// +40 Attack Power.
		// https://www.wowhead.com/forever/spell=14049
		6: {14049},
	})

	// Champion's Wartide
	// https://www.wowhead.com/forever/item-set=2090
	shared.NewParsedItemSet(2090, "Champion's Wartide", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Increases healing done by up to 44 and damage done by up to 15 for all magical spells and effects.
		// https://www.wowhead.com/forever/spell=467550
		2: {467550},
		// Improves your chance to get a critical strike with all Shock spells by 2%.
		// https://www.wowhead.com/forever/spell=22804
		4: {22804},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Warlord's Wartide
	// https://www.wowhead.com/forever/item-set=2091
	shared.NewParsedItemSet(2091, "Warlord's Wartide", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		2: {14467},
		// Improves your chance to get a critical strike with all Shock spells by 2%.
		// https://www.wowhead.com/forever/spell=22804
		3: {22804},
		// Increases healing done by up to 44 and damage done by up to 15 for all magical spells and effects.
		// https://www.wowhead.com/forever/spell=467550
		6: {467550},
	})

	// Lieutenant Commander's Wartide
	// https://www.wowhead.com/forever/item-set=2092
	shared.NewParsedItemSet(2092, "Lieutenant Commander's Wartide", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Increases healing done by up to 44 and damage done by up to 15 for all magical spells and effects.
		// https://www.wowhead.com/forever/spell=467550
		2: {467550},
		// Improves your chance to get a critical strike with all Shock spells by 2%.
		// https://www.wowhead.com/forever/spell=22804
		4: {22804},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Field Marshal's Wartide
	// https://www.wowhead.com/forever/item-set=2093
	shared.NewParsedItemSet(2093, "Field Marshal's Wartide", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		2: {14467},
		// Improves your chance to get a critical strike with all Shock spells by 2%.
		// https://www.wowhead.com/forever/spell=22804
		3: {22804},
		// Increases healing done by up to 44 and damage done by up to 15 for all magical spells and effects.
		// https://www.wowhead.com/forever/spell=467550
		6: {467550},
	})

	// Champion's Thunderfist
	// https://www.wowhead.com/forever/item-set=2094
	shared.NewParsedItemSet(2094, "Champion's Thunderfist", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Increases damage and healing done by magical spells and effects by up to 23.
		// https://www.wowhead.com/forever/spell=14047
		2: {14047},
		// Improves your chance to get a critical strike with all Shock spells by 2%.
		// https://www.wowhead.com/forever/spell=22804
		4: {22804},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Warlord's Thunderfist
	// https://www.wowhead.com/forever/item-set=2095
	shared.NewParsedItemSet(2095, "Warlord's Thunderfist", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		2: {14467},
		// Improves your chance to get a critical strike with all Shock spells by 2%.
		// https://www.wowhead.com/forever/spell=22804
		3: {22804},
		// Increases damage and healing done by magical spells and effects by up to 23.
		// https://www.wowhead.com/forever/spell=14047
		6: {14047},
	})

	// Lieutenant Commander's Thunderfist
	// https://www.wowhead.com/forever/item-set=2096
	shared.NewParsedItemSet(2096, "Lieutenant Commander's Thunderfist", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Increases damage and healing done by magical spells and effects by up to 23.
		// https://www.wowhead.com/forever/spell=14047
		2: {14047},
		// Improves your chance to get a critical strike with all Shock spells by 2%.
		// https://www.wowhead.com/forever/spell=22804
		4: {22804},
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		6: {14467},
	})

	// Field Marshal's Thunderfist
	// https://www.wowhead.com/forever/item-set=2097
	shared.NewParsedItemSet(2097, "Field Marshal's Thunderfist", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +20 Stamina.
		// https://www.wowhead.com/forever/spell=14467
		2: {14467},
		// Improves your chance to get a critical strike with all Shock spells by 2%.
		// https://www.wowhead.com/forever/spell=22804
		3: {22804},
		// Increases damage and healing done by magical spells and effects by up to 23.
		// https://www.wowhead.com/forever/spell=14047
		6: {14047},
	})

	// Manaflare Regalia
	// https://www.wowhead.com/forever/item-set=2098
	shared.NewParsedItemSet(2098, "Manaflare Regalia", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Improves your chance to hit by 1.0%.
		// https://www.wowhead.com/forever/spell=1300947
		2: {1300947},
		// Reduces the cooldown of your Counterspell by 5 sec.
		// https://www.wowhead.com/forever/spell=1301013
		3: {1301013},

		// TODO: Manual implementation required
		//
		// Increases damage done by your spells and effects by up to 21 when fighting Elementals.
		// https://www.wowhead.com/forever/spell=1301079
		// unsupported: effect 1 A_MOD_DAMAGE_DONE_CREATURE misc 8 is not parsed
		// 4: {1301079},

		// TODO: Manual implementation required
		//
		// Your Frostfire Bolt spell has a 10% increased chance to trigger Missile Barrage, gains 10% increased critical
		// strike chance while your Combustion spell is active, and has a 10% increased chance to trigger Fingers
		// of Frost.
		// https://www.wowhead.com/forever/spell=1301488
		// unsupported: effect 1 A_DUMMY misc 0 is not parsed, effect 2 A_DUMMY misc 0 is not parsed, effect 3 A_DUMMY misc 0 is not parsed
		// 5: {1301488},
	})

	// Grimstitch Armor
	// https://www.wowhead.com/forever/item-set=2099
	shared.NewParsedItemSet(2099, "Grimstitch Armor", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Improves your chance to hit by 1.0%.
		// https://www.wowhead.com/forever/spell=1300956
		2: {1300956},
		// Reduces the cooldown of your Kidney Shot ability by 3 sec.
		// https://www.wowhead.com/forever/spell=1301047
		3: {1301047},
		// +36 Attack Power against Humanoids.
		// https://www.wowhead.com/forever/spell=1301088
		4: {1301088},
		// Reduces the cost of your Envenom and Eviscerate abilities by 5 Energy.
		// https://www.wowhead.com/forever/spell=1301708
		5: {1301708},
	})

	// The Spiritcaller
	// https://www.wowhead.com/forever/item-set=2109
	shared.NewParsedItemSet(2109, "The Spiritcaller", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Increased Spirit +10
		// https://www.wowhead.com/forever/spell=1300960
		2: {1300960},

		// TODO: Manual implementation required
		//
		// Increases the radius of effect of your beneficial Fire and Water totems by 10 yds.
		// https://www.wowhead.com/forever/spell=1301059
		// unsupported: effect 1 A_ADD_FLAT_MODIFIER misc 6 is not parsed
		// 3: {1301059},

		// Increases healing done by up to 26 and damage done by up to 9 for all magical spells and effects.
		// https://www.wowhead.com/forever/spell=1301093
		4: {1301093},
		// Reduces the cooldown of your Riptide spell by 1 sec.
		// https://www.wowhead.com/forever/spell=1301713
		5: {1301713},
	})

	// The Spiritcaller's Rage
	// https://www.wowhead.com/forever/item-set=2110
	shared.NewParsedItemSet(2110, "The Spiritcaller's Rage", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Increases your attack speed and casting speed by 1%.
		// https://www.wowhead.com/forever/spell=1300957
		2: {1300957},

		// TODO: Manual implementation required
		//
		// Increases the radius of effect of your beneficial Air and Earth totems by 10 yds.
		// https://www.wowhead.com/forever/spell=1301051
		// unsupported: effect 1 A_ADD_FLAT_MODIFIER misc 6 is not parsed
		// 3: {1301051},

		// +36 Attack Power against Elementals.
		// https://www.wowhead.com/forever/spell=1301090
		4: {1301090},
		// Reduces the cooldown of your Stormstrike ability by 0.5 sec.
		// https://www.wowhead.com/forever/spell=1301709
		5: {1301709},
	})

	// The Spiritcaller's Storm
	// https://www.wowhead.com/forever/item-set=2111
	shared.NewParsedItemSet(2111, "The Spiritcaller's Storm", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Improves your chance to hit by 1.0%.
		// https://www.wowhead.com/forever/spell=1300959
		2: {1300959},
		// Reduces the cooldown of your Earthbind Totem spell by 2 sec.
		// https://www.wowhead.com/forever/spell=1301053
		3: {1301053},

		// TODO: Manual implementation required
		//
		// Increases damage done by your spells and effects by up to 21 when fighting Elementals.
		// https://www.wowhead.com/forever/spell=1301092
		// unsupported: effect 1 A_MOD_DAMAGE_DONE_CREATURE misc 8 is not parsed
		// 4: {1301092},

		// Reduces the cooldown of your Lava Burst spell by 1 sec.
		// https://www.wowhead.com/forever/spell=1301711
		5: {1301711},
	})

	// Grovekeeper Raiment
	// https://www.wowhead.com/forever/item-set=2112
	shared.NewParsedItemSet(2112, "Grovekeeper Raiment", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Increased Spirit +10
		// https://www.wowhead.com/forever/spell=1300943
		2: {1300943},
		// Reduces the cooldown of your Barkskin spell by 10 sec.
		// https://www.wowhead.com/forever/spell=1301006
		3: {1301006},
		// Increases healing done by up to 26 and damage done by up to 9 for all magical spells and effects.
		// https://www.wowhead.com/forever/spell=1301077
		4: {1301077},
		// Reduces the cooldown of your Swiftmend spell by 3.0 sec.
		// https://www.wowhead.com/forever/spell=1301250
		5: {1301250},
	})

	// Grovekeeper Rage
	// https://www.wowhead.com/forever/item-set=2113
	shared.NewParsedItemSet(2113, "Grovekeeper Rage", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Increased Defense +7
		// https://www.wowhead.com/forever/spell=1300942
		2: {1300942},

		// TODO: Manual implementation required
		//
		// Your Rebirth spell is now castable in Bear Form and Dire Bear Form and its cast time is reduced by 2.0
		// sec while you are in Bear Form or Dire Bear Form.
		// https://www.wowhead.com/forever/spell=1301005
		// unsupported: effect 2 A_MOD_IGNORE_SHAPESHIFT misc 0 is not parsed
		// 3: {1301005},

		// Reduces the chance for your melee attacks to be Dodged or Parried by 1.2%.
		// https://www.wowhead.com/forever/spell=1301076
		4: {1301076},
		// Reduces the cooldown of your Berserk ability by 15.0 sec.
		// https://www.wowhead.com/forever/spell=1301249
		5: {1301249},
	})

	// Grovekeeper Ferocity
	// https://www.wowhead.com/forever/item-set=2115
	shared.NewParsedItemSet(2115, "Grovekeeper Ferocity", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// Increases your attack speed and casting speed by 1%.
		// https://www.wowhead.com/forever/spell=1300934
		2: {1300934},

		// TODO: Manual implementation required
		//
		// Your Hibernate spell is now castable in Cat Form and its cast time is reduced by 1.5 sec while you are
		// in Cat Form, but Hibernate gains a cooldown of 15 sec.
		// https://www.wowhead.com/forever/spell=1300999
		// unsupported: effect 3 A_MOD_IGNORE_SHAPESHIFT misc 0 is not parsed
		// 3: {1300999},

		// +36 Attack Power against Demons.
		// https://www.wowhead.com/forever/spell=1301075
		4: {1301075},
		// Reduces the cooldown of your Shifting Power ability by 1.0 sec.
		// https://www.wowhead.com/forever/spell=1301247
		5: {1301247},
	})

	// TODO: Manual implementation required
	//
	// Rider of the Plaguelands
	// https://www.wowhead.com/forever/item-set=2130
	// shared.NewParsedItemSet(2130, "Rider of the Plaguelands", []proto.ItemSlot{proto.ItemSlot_ItemSlotBack, proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2, proto.ItemSlot_ItemSlotMainHand}, map[int32][]int32{
	// 	// Increases movement speed while mounted by 5% in Eastern and Western Plaguelands.
	// 	// https://www.wowhead.com/forever/spell=1315080
	// 	// unsupported: the row applies only in area group 9317, effect 1 A_MOD_MOUNTED_SPEED_ALWAYS misc 0 is not parsed
	// 	2: {1315080},
	// 	// Chance on melee attack to lash the target, inflicting 140 Shadow damage.
	// 	// https://www.wowhead.com/forever/spell=1315077
	// 	// unsupported: effect 1 is E_SCHOOL_DAMAGE
	// 	3: {1315077},
	// })

	// TODO: Manual implementation required
	//
	// Partners in Crime
	// https://www.wowhead.com/forever/item-set=2131
	// shared.NewParsedItemSet(2131, "Partners in Crime", []proto.ItemSlot{proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2}, map[int32][]int32{
	// 	// Chance on melee swing to become Angry, increasing Strength by 50 for 10s.
	// 	// https://www.wowhead.com/forever/spell=1317039
	// 	// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
	// 	// +22 Strength.
	// 	// https://www.wowhead.com/forever/spell=1317038
	// 	2: {1317039, 1317038},
	// })

	// TODO: Manual implementation required
	//
	// Blessing of Kalimdor
	// https://www.wowhead.com/forever/item-set=2132
	// shared.NewParsedItemSet(2132, "Blessing of Kalimdor", []proto.ItemSlot{proto.ItemSlot_ItemSlotNeck, proto.ItemSlot_ItemSlotBack, proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2}, map[int32][]int32{
	// 	// Increases movement speed by 5% in The Barrens and Stonetalon Mountains.
	// 	// https://www.wowhead.com/forever/spell=1318244
	// 	// unsupported: the row applies only in area group 9341, effect 1 A_MOD_SPEED_ALWAYS misc 0 is not parsed
	// 	2: {1318244},
	// })

	// Rotmender's Raiment
	// https://www.wowhead.com/forever/item-set=2133
	shared.NewParsedItemSet(2133, "Rotmender's Raiment", []proto.ItemSlot{proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotWaist, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +5 Shadow Resistance.
		// https://www.wowhead.com/forever/spell=14668
		2: {14668},
		// +10 Intellect.
		// https://www.wowhead.com/forever/spell=9106
		3: {9106},

		// TODO: Manual implementation required
		//
		// When your Mana falls below 15%, restore 200 Mana over 10s. Restores twice as much Mana in Haunted and
		// Wasteland areas. May only occur once every 5 min.
		// https://www.wowhead.com/forever/spell=1321560
		// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
		// 4: {1321560},

		// TODO: Manual implementation required
		//
		// Grants your healing spells a chance to heal for an additional 40 Health every 3.0 sec for 15s.
		// https://www.wowhead.com/forever/spell=1321587
		// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
		// 5: {1321587},
	})

	// Violet Sorcerer's Vestments
	// https://www.wowhead.com/forever/item-set=2134
	shared.NewParsedItemSet(2134, "Violet Sorcerer's Vestments", []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +5 Arcane Resistance.
		// https://www.wowhead.com/forever/spell=1324311
		2: {1324311},
		// +10 Spirit.
		// https://www.wowhead.com/forever/spell=1324312
		3: {1324312},

		// TODO: Manual implementation required
		//
		// Chance on damaging spell to allow 15% of your Mana regeneration to continue while casting for 15s. This
		// Mana Regeneration is doubled in Strongholds and Cities.
		// https://www.wowhead.com/forever/spell=1324318
		// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
		// 4: {1324318},

		// TODO: Manual implementation required
		//
		// Chance on damaging spell to summon an Arcane elemental to aid you in combat.
		// https://www.wowhead.com/forever/spell=1324323
		// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
		// 5: {1324323},
	})

	// Krol'dok Battlegear
	// https://www.wowhead.com/forever/item-set=2135
	shared.NewParsedItemSet(2135, "Krol'dok Battlegear", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotLegs}, map[int32][]int32{
		// +100 Armor.
		// https://www.wowhead.com/forever/spell=1324287
		2: {1324287},
		// Increased Defense +5.
		// https://www.wowhead.com/forever/spell=1324288
		3: {1324288},

		// TODO: Manual implementation required
		//
		// Chance on block to heal for 18. Healing is doubled in Strongholds and Cities.
		// https://www.wowhead.com/forever/spell=1324298
		// unsupported: effect 2 is E_DUMMY, effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
		// 4: {1324298},

		// TODO: Manual implementation required
		//
		// Taking damage has a high chance to increase your Stamina by 1 for 30s, stacking up to 20 times.
		// https://www.wowhead.com/forever/spell=1324309
		// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
		// 5: {1324309},
	})

	// Defias Enforcer's Garb
	// https://www.wowhead.com/forever/item-set=2136
	shared.NewParsedItemSet(2136, "Defias Enforcer's Garb", []proto.ItemSlot{proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder, proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet}, map[int32][]int32{
		// +10 Fire Resistance.
		// https://www.wowhead.com/forever/spell=1324673
		2: {1324673},
		// +15 Stamina.
		// https://www.wowhead.com/forever/spell=1324674
		3: {1324674},
		// Increased Swords +1.
		// https://www.wowhead.com/forever/spell=1324675
		4: {1324675},

		// TODO: Manual implementation required
		//
		// Melee attacks have a chance to increase your Critical Strike chance by 2% for 20s. This effect is doubled
		// in swamps.
		// https://www.wowhead.com/forever/spell=1324684
		// unsupported: effect 1 A_PROC_TRIGGER_SPELL misc 0 is not parsed
		// 5: {1324684},
	})
}

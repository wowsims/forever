package forever

import (
	"github.com/wowsims/forever/sim/common/shared"
)

func RegisterAllEnchants() {

	// Enchants

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Attaches an Iron Spike to your shield that deals damage every time you block with it.
	// https://www.wowhead.com/forever/spell=7216
	// unsupported: an outcome the proc mask has no bit for; the damage spell's row states no damage
	// trigger 9784 (every time, core.CallbackOnSpellHitTaken, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{
	//	Name:           "Iron Shield Spike",
	//	EnchantID:      43,
	//	TriggerSpellID: 9784,
	//	BuffSpellID:    9784,
	// }, nil)

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Attaches a Mithril Spike to your shield that deals damage every time you block with it.
	// https://www.wowhead.com/forever/spell=9781
	// unsupported: an outcome the proc mask has no bit for; the damage spell's row states no damage
	// trigger 9782 (every time, core.CallbackOnSpellHitTaken, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{
	//	Name:           "Mithril Shield Spike",
	//	EnchantID:      463,
	//	TriggerSpellID: 9782,
	//	BuffSpellID:    9782,
	// }, nil)

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Permanently enchant a melee weapon to often strike for 40 additional fire damage.
	// https://www.wowhead.com/forever/spell=13898
	// unsupported: states no rate
	// trigger 13897 (0%, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{
	//	Name:           "Enchant Weapon - Fiery Weapon",
	//	EnchantID:      803,
	//	TriggerSpellID: 13897,
	//	BuffSpellID:    13897,
	//	IsWeaponProc:   true,
	// }, nil)

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Permanently enchant a melee weapon to have a chance of stunning and doing heavy damage to demons.
	// https://www.wowhead.com/forever/spell=13915
	// unsupported: states no rate; the damage spell hits demons (TargetCreatureType 4) only
	// trigger 13907 (0%, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{
	//	Name:           "Enchant Weapon - Demonslaying",
	//	EnchantID:      912,
	//	TriggerSpellID: 13907,
	//	BuffSpellID:    13907,
	//	IsWeaponProc:   true,
	// }, nil)

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Attaches a Thorium Spike to your shield that deals damage every time you block with it.
	// https://www.wowhead.com/forever/spell=16623
	// unsupported: an outcome the proc mask has no bit for; the damage spell's row states no damage
	// trigger 16624 (every time, core.CallbackOnSpellHitTaken, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{
	//	Name:           "Thorium Shield Spike",
	//	EnchantID:      1704,
	//	TriggerSpellID: 16624,
	//	BuffSpellID:    16624,
	// }, nil)

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Permanently enchant a melee weapon to often chill the target reducing their movement and attack speed.
	// https://www.wowhead.com/forever/spell=20029
	// unsupported: states no rate
	// trigger 20005 (0%, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDebuffProc(shared.SpellDataProc{
	//	Name:           "Enchant Weapon - Icy Chill",
	//	EnchantID:      1894,
	//	TriggerSpellID: 20005,
	//	IsWeaponProc:   true,
	// }, nil)

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Permanently enchant a melee weapon to often steal life from the enemy and give it to the wielder.
	// https://www.wowhead.com/forever/spell=20032
	// unsupported: states no rate; the damage spell's row states no damage
	// trigger 20004 (0%, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{
	//	Name:           "Enchant Weapon - Lifestealing",
	//	EnchantID:      1898,
	//	TriggerSpellID: 20004,
	//	BuffSpellID:    20004,
	//	IsWeaponProc:   true,
	// }, nil)

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Permanently enchant a melee weapon to often inflict a curse on the target reducing their melee damage.
	// https://www.wowhead.com/forever/spell=20033
	// unsupported: states no rate
	// trigger 20006 (0%, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDebuffProc(shared.SpellDataProc{
	//	Name:           "Enchant Weapon - Unholy Weapon",
	//	EnchantID:      1899,
	//	TriggerSpellID: 20006,
	//	IsWeaponProc:   true,
	// }, nil)

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Permanently enchant a Weapon to cause all spells and attacks to sometimes deal 75 additional damage to
	// mechanical creatures.
	// https://www.wowhead.com/forever/spell=435481
	// unsupported: states no rate; the damage spell hits mechanicals (TargetCreatureType 256) only
	// trigger 435467 (no stated rate, core.CallbackOnSpellHitDealt, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial | core.ProcMaskRangedAuto | core.ProcMaskRangedSpecial) -> buff 439164
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{
	//	Name:           "Enchant Weapon - Dismantle",
	//	EnchantID:      7210,
	//	TriggerSpellID: 435467,
	//	BuffSpellID:    439164,
	// }, nil)

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Permanently enchant a Weapon to cause all spells and attacks to sometimes deal 75 additional damage to
	// mechanical creatures.
	// https://www.wowhead.com/forever/spell=435481
	// unsupported: the damage spell hits mechanicals (TargetCreatureType 256) only
	// trigger 442206 (30%, core.CallbackOnCastComplete, core.ProcMaskSpellDamage) -> buff 439164
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{
	//	Name:           "Enchant Weapon - Dismantle",
	//	EnchantID:      7210,
	//	TriggerSpellID: 442206,
	//	BuffSpellID:    439164,
	// }, nil)

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Teaches you how to permanently enchant a piece of chest armor to reflect 9 damage back to the attacker
	// when the bearer is struck in Melee.
	//
	// https://www.wowhead.com/forever/spell=435903
	// unsupported: the enchant's effect entry resolves no stats from 435901 (A_DAMAGE_SHIELD)
	// trigger 435901 (every time, core.CallbackOnSpellHitTaken, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial)
	// shared.NewSpellDataProc(shared.SpellDataProc{
	//	Name:           "Enchant Chest - Retricutioner",
	//	EnchantID:      7223,
	//	TriggerSpellID: 435901,
	// }, nil)

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Permanently enchant a melee weapon so that often when attacking in melee it heals for 400 and increases
	// Strength by 120 for 20s.
	// https://www.wowhead.com/forever/spell=1231128
	// unsupported: states no rate
	// trigger 1231124 (0%, core.CallbackEmpty, core.ProcMaskUnknown); slot spell 1231126 applies the same spell and is not registered
	// shared.NewSpellDataProc(shared.SpellDataProc{
	//	Name:           "Enchant Weapon - Grand Crusader",
	//	EnchantID:      7940,
	//	TriggerSpellID: 1231124,
	//	IsWeaponProc:   true,
	// }, nil)

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Permanently enchant a two-handed melee weapon so that often when attacking in melee it heals for 400 and
	// increases Strength by 200 for 20s.
	// https://www.wowhead.com/forever/spell=1232172
	// unsupported: states no rate
	// trigger 1232169 (0%, core.CallbackEmpty, core.ProcMaskUnknown); slot spell 1232170 applies the same spell and is not registered
	// shared.NewSpellDataProc(shared.SpellDataProc{
	//	Name:           "Enchant 2H Weapon - Grand Inquisitor",
	//	EnchantID:      7943,
	//	TriggerSpellID: 1232169,
	//	IsWeaponProc:   true,
	// }, nil)

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Permanently enchant a pair of gloves to give a small chance to acquire Death Lotus when gathering any
	// herb in Hyjal.
	// https://www.wowhead.com/forever/spell=1294054
	// unsupported: no callback in the proc mask; states no rate; the enchant's effect entry resolves no stats from 1294053 (A_DUMMY)
	// trigger 1294053 (0%, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataProc(shared.SpellDataProc{
	//	Name:           "Enchant Gloves - Lotus Claw",
	//	EnchantID:      8695,
	//	TriggerSpellID: 1294053,
	// }, nil)

	// Enchants a weapon to have a 15% chance to inflict 11 Fire damage to all enemies within 3 yards.
	// https://www.wowhead.com/forever/spell=6296
	// trigger 6297 (15%, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{
		Name:           "Enchant: Fiery Blaze",
		EnchantID:      36,
		TriggerSpellID: 6297,
		BuffSpellID:    6297,
		IsWeaponProc:   true,
	}, nil)

	// Enchant a piece of chest armor so it has a 2% chance per hit of giving you 10 points of damage absorption.
	// Cannot occur more often than once every 5 sec.
	// https://www.wowhead.com/forever/spell=7426
	// trigger 7445 (2%, core.CallbackOnSpellHitTaken, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial) -> buff 7423
	shared.NewSpellDataAbsorbProc(shared.SpellDataProc{
		Name:           "Enchant Chest - Minor Absorption",
		EnchantID:      44,
		TriggerSpellID: 7445,
		BuffSpellID:    7423,
	}, nil)

	// Enchant a piece of chest armor so it has a 5% chance per hit of giving you 25 points of damage absorption.
	// Cannot occur more often than once every 5 sec.
	// https://www.wowhead.com/forever/spell=13538
	// trigger 7446 (5%, core.CallbackOnSpellHitTaken, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial) -> buff 7447
	shared.NewSpellDataAbsorbProc(shared.SpellDataProc{
		Name:           "Enchant Chest - Lesser Absorption",
		EnchantID:      63,
		TriggerSpellID: 7446,
		BuffSpellID:    7447,
	}, nil)

	// Permanently enchant a two-handed melee weapon so that often when striking with a spell it restores 400
	// mana and increases Spell Power by 140 for 20s.
	// https://www.wowhead.com/forever/spell=1231139
	// trigger 1231152 (7%, core.CallbackOnSpellHitDealt | core.CallbackOnHealDealt, core.ProcMaskSpellDamage | core.ProcMaskSpellHealing) -> buff 1231138
	shared.NewSpellDataProc(shared.SpellDataProc{
		Name:           "Enchant 2H Weapon - Grand Arcanist",
		EnchantID:      7941,
		TriggerSpellID: 1231152,
		BuffSpellID:    1231138,
	}, nil)

	// Permanently enchant a melee weapon so that often when striking with a spell it restores 400 mana and increases
	// Spell Power by 70 for 20s.
	// https://www.wowhead.com/forever/spell=1231164
	// trigger 1231163 (7%, core.CallbackOnSpellHitDealt | core.CallbackOnHealDealt, core.ProcMaskSpellDamage | core.ProcMaskSpellHealing) -> buff 1231162
	shared.NewSpellDataProc(shared.SpellDataProc{
		Name:           "Enchant Weapon - Grand Sorcerer",
		EnchantID:      7942,
		TriggerSpellID: 1231163,
		BuffSpellID:    1231162,
	}, nil)

	// Permanently enchant a Melee Weapon to have a chance to grant Insight when you cast a spell, increasing
	// Spirit by 100% for 10s.
	// https://www.wowhead.com/forever/spell=1248757
	// trigger 1248758 (35%, core.CallbackOnSpellHitDealt | core.CallbackOnHealDealt, core.ProcMaskSpellDamage | core.ProcMaskSpellHealing) -> buff 1299796
	shared.NewSpellDataAuraProc(shared.SpellDataProc{
		Name:           "Enchant Weapon - Insight",
		EnchantID:      8216,
		TriggerSpellID: 1248758,
		BuffSpellID:    1299796,
	}, nil)

	// Enchant a piece of chest armor so it has a 25% chance per hit of giving you 50 points of damage absorption.
	// Cannot occur more often than once every 5 sec.
	// https://www.wowhead.com/forever/spell=1249071
	// trigger 1249072 (25%, core.CallbackOnSpellHitTaken, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial) -> buff 1249073
	shared.NewSpellDataAbsorbProc(shared.SpellDataProc{
		Name:           "Enchant Chest - Absorption",
		EnchantID:      8220,
		TriggerSpellID: 1249072,
		BuffSpellID:    1249073,
	}, nil)

	// Permanently enchant a Melee Weapon to trigger Recovery when you are Parried or Dodged, healing you for
	// 5% of your maximum health. Cannot occur more often than once every 10 sec.
	// https://www.wowhead.com/forever/spell=1248760
	// trigger 1248761 (every time, core.CallbackOnSpellHitDealt, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial) -> buff 1248759; the enchant's tooltip restricts it to core.OutcomeDodge | core.OutcomeParry
	shared.NewSpellDataHealProc(shared.SpellDataProc{
		Name:           "Enchant Weapon - Recovery",
		EnchantID:      8721,
		TriggerSpellID: 1248761,
		BuffSpellID:    1248759,
	}, nil)
}

package forever

import (
	"github.com/wowsims/forever/sim/common/shared"
)

func RegisterAllProcs() {

	// Equip

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases Feral Combat +1.
	// https://www.wowhead.com/forever/spell=1320726
	// unsupported: effect 1 A_MOD_SKILL misc 3014 is not parsed on the wearer
	// equip: 1320726 (A_MOD_SKILL)
	// shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 1320726},
	//	[]shared.ItemVariant{
	//	{ItemID: 9414, ItemName: "Oilskin Leggings"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Reduces your damage taken by 5% in the Battle Ring and The Maul.
	// https://www.wowhead.com/forever/spell=1317285
	// unsupported: the row applies only in area group 9337, which names no area type
	// equip: 1317285 (A_MOD_DAMAGE_PERCENT_TAKEN)
	// shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 1317285},
	//	[]shared.ItemVariant{
	//	{ItemID: 18706, ItemName: "Arena Master"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Reduces your damage taken by 6% in the Battle Ring and The Maul.
	// https://www.wowhead.com/forever/spell=1318318
	// unsupported: the row applies only in area group 9337, which names no area type
	// equip: 1318318 (A_MOD_DAMAGE_PERCENT_TAKEN)
	// shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 1318318},
	//	[]shared.ItemVariant{
	//	{ItemID: 19024, ItemName: "Arena Grand Master"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases your pet's armor by 10%.
	// https://www.wowhead.com/forever/spell=27225
	// unsupported: effect 1 A_MOD_BASE_RESISTANCE_PCT misc 1 is not parsed on a pet
	// equip: 27225 keeps 27208 up (A_MOD_BASE_RESISTANCE_PCT)
	// shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 27225},
	//	[]shared.ItemVariant{
	//	{ItemID: 22060, ItemName: "Beastmaster's Tunic"},
	// })

	// Deals 5 Fire damage to anyone who strikes you with a melee attack.
	// https://www.wowhead.com/forever/spell=21142
	// equip: 21142 (A_DAMAGE_SHIELD)
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 21142},
		[]shared.ItemVariant{
			{ItemID: 3475, ItemName: "Cloak of Flames"},
		})

	// When struck in combat inflicts 13 Fire damage to the attacker.
	// https://www.wowhead.com/forever/spell=23266
	// equip: 23266 (A_DAMAGE_SHIELD)
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 23266},
		[]shared.ItemVariant{
			{ItemID: 18815, ItemName: "Essence of the Pure Flame"},
		})

	// Increases your pet's maximum health by 3%.
	// https://www.wowhead.com/forever/spell=27038
	// equip: 27038 keeps 27039 up (A_MOD_INCREASE_HEALTH_PERCENT)
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 27038},
		[]shared.ItemVariant{
			{ItemID: 22013, ItemName: "Beastmaster's Cap"},
		})

	// Increases your pet's critical strike chance by 2%.
	// https://www.wowhead.com/forever/spell=27043
	// equip: 27043 keeps 27042 up (A_MOD_WEAPON_CRIT_PERCENT)
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 27043},
		[]shared.ItemVariant{
			{ItemID: 22015, ItemName: "Beastmaster's Gloves"},
		})

	// Increases damage dealt by your pet by 3%.
	// https://www.wowhead.com/forever/spell=27206
	// equip: 27206 keeps 27205 up (A_MOD_DAMAGE_PERCENT_DONE)
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 27206},
		[]shared.ItemVariant{
			{ItemID: 22061, ItemName: "Beastmaster's Boots"},
		})

	// Spell Damage received is reduced by 10.
	// https://www.wowhead.com/forever/spell=27518
	// equip: 27518 (A_MOD_DAMAGE_TAKEN)
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 27518},
		[]shared.ItemVariant{
			{ItemID: 22191, ItemName: "Obsidian Mail Tunic"},
		})

	// When struck in combat, inflicts 3 Nature damage to the attacker.
	// https://www.wowhead.com/forever/spell=1297378
	// equip: 1297378 (A_DAMAGE_SHIELD)
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 1297378},
		[]shared.ItemVariant{
			{ItemID: 275649, ItemName: "Bramblebark Barrier"},
		})

	// When struck in combat, inflicts 3 Nature damage to the attacker.
	// https://www.wowhead.com/forever/spell=1297910
	// equip: 1297910 (A_DAMAGE_SHIELD)
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 1297910},
		[]shared.ItemVariant{
			{ItemID: 275833, ItemName: "Bristlecone Cloak"},
		})

	// When struck in combat, inflicts 2 Fire damage to the attacker.
	// https://www.wowhead.com/forever/spell=1312955
	// equip: 1312955 (A_DAMAGE_SHIELD)
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 1312955},
		[]shared.ItemVariant{
			{ItemID: 282080, ItemName: "Flame Seared Signet"},
		})

	// Procs

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Decrease the armor of the target by 100 for 30s. While affected, the target cannot stealth or turn invisible.
	// https://www.wowhead.com/forever/spell=13752
	// unsupported: states no rate
	// trigger 13752 (every time, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDebuffProc(shared.SpellDataProc{TriggerSpellID: 13752, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 869, ItemName: "Dazzling Longsword"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Wounds the target for 102 damage.
	// https://www.wowhead.com/forever/spell=18090
	// unsupported: states no rate
	// trigger 18090 (every time, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 18090, BuffSpellID: 18090, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 1986, ItemName: "Gutrender"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Launches a bolt of frost at the enemy causing 50 Frost damage and slowing movement speed by 50% for 5s.
	// https://www.wowhead.com/forever/spell=13439
	// unsupported: states no rate
	// trigger 13439 (every time, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 13439, BuffSpellID: 13439, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 3854, ItemName: "Frost Tiger Blade"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Smite your target with the power of the holy spring, inflicting 56 Holy damage.
	// https://www.wowhead.com/forever/spell=1312330
	// unsupported: states no rate
	// trigger 1312330 (every time, core.CallbackOnSpellHitDealt, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeMHSpecial)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 1312330, BuffSpellID: 1312330, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 5028, ItemName: "Lord Sakrasis' Scepter"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Disarm target's weapon for 10s.
	// https://www.wowhead.com/forever/spell=13534
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 7954, ItemName: "The Shatterer"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Diseases a target for 100 Nature damage and an additional 360 damage over 1min.
	// https://www.wowhead.com/forever/spell=9796
	// unsupported: states no rate
	// trigger 9796 (0%, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 9796, BuffSpellID: 9796, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 7959, ItemName: "Blight"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Protects the caster with a holy shield.
	// https://www.wowhead.com/forever/spell=9800
	// unsupported: states no rate
	// trigger 9800 (every time, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataAbsorbProc(shared.SpellDataProc{TriggerSpellID: 9800, BuffSpellID: 9800, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 7960, ItemName: "Truesilver Champion"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Deal 112 Fire damage. Deals 3 times as much damage to Earth Elementals, Fire Elementals, and Mountain
	// Giants.
	// https://www.wowhead.com/forever/spell=1317432
	// unsupported: states no rate
	// trigger 1317432 (0%, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 1317432, BuffSpellID: 1317432, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 8708, ItemName: "Hammer of Expertise"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Chance on harmful spell cast to blast a target for 77 Fire damage.
	// https://www.wowhead.com/forever/spell=18083
	// unsupported: states no rate
	// trigger 1293980 (no stated rate, core.CallbackOnCastComplete, core.ProcMaskSpellDamage) -> buff 18083
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 1293980, BuffSpellID: 18083},
	//	[]shared.ItemVariant{
	//	{ItemID: 9419, ItemName: "Galgann's Firehammer"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Water breathing lasts 100% longer than normal.
	// https://www.wowhead.com/forever/spell=1320643
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 10506, ItemName: "Deepdive Helmet"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Wounds the target causing them to bleed for 70 damage over 7s.
	// https://www.wowhead.com/forever/spell=18075
	// unsupported: states no rate
	// trigger 18075 (every time, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 18075, BuffSpellID: 18075, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 10772, ItemName: "Glutton's Cleaver"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Reduces the duration of all movement impairing effects on you by 7%.
	// https://www.wowhead.com/forever/spell=1293819
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 10776, ItemName: "Silky Spider Cape"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Reduces targets armor by 300 for 20s.
	// https://www.wowhead.com/forever/spell=15280
	// unsupported: states no rate
	// trigger 15280 (every time, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDebuffProc(shared.SpellDataProc{TriggerSpellID: 15280, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 11607, ItemName: "Dark Iron Sunderer"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Stuns target for 8s.
	// https://www.wowhead.com/forever/spell=15283
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 11608, ItemName: "Dark Iron Pulverizer"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// When struck in combat has a 5% chance to make you invulnerable to melee damage for 3s. This effect can
	// only occur once every 10 sec.
	// https://www.wowhead.com/forever/spell=16621
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 12641, ItemName: "Invulnerable Mail"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Enemy is inflicted with the Bleakwood Curse that reduces their magic resistances by 25. Can be applied
	// up to 3 times.
	// https://www.wowhead.com/forever/spell=16871
	// unsupported: states no rate
	// trigger 16871 (every time, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDebuffProc(shared.SpellDataProc{TriggerSpellID: 16871, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 12769, ItemName: "Bleakwood Hew"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Dispels a magic effect on the current foe.
	// https://www.wowhead.com/forever/spell=16908
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 12781, ItemName: "Serenity"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Stuns target for 3s.
	// https://www.wowhead.com/forever/spell=56
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 12796, ItemName: "Hammer of the Titans"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Strike the target with a Flaming Shell for 38 Fire damage.
	// https://www.wowhead.com/forever/spell=29647
	// unsupported: states no rate
	// trigger 29647 (0%, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 29647, BuffSpellID: 29647, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 13146, ItemName: "Shell Launcher Shotgun"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Blasts a target for 700 Fire damage.
	// https://www.wowhead.com/forever/spell=18112
	// unsupported: states no rate
	// trigger 18112 (every time, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 18112, BuffSpellID: 18112, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 13262, ItemName: "Ashbringer"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Reduces the cooldown of your Fade ability by 2.0 sec.
	// https://www.wowhead.com/forever/spell=18388
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 14154, ItemName: "Truefaith Vestments"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Protects the wearer from being fully engulfed by Shadow Flame.
	// https://www.wowhead.com/forever/spell=22683
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 15138, ItemName: "Onyxia Scale Cloak"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Fires a Shadow Shot at the target for 26 Shadow damage.
	// https://www.wowhead.com/forever/spell=29641
	// unsupported: states no rate
	// trigger 29641 (0%, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 29641, BuffSpellID: 29641, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 16004, ItemName: "Dark Iron Rifle"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// https://www.wowhead.com/forever/spell=1318325
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 16022, ItemName: "Arcanite Dragonling"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the damage absorbed by your Mana Shield by 285.
	// https://www.wowhead.com/forever/spell=23037
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 16391, ItemName: "Knight-Lieutenant's Silk Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the duration of your Sprint ability by 3.0 sec.
	// https://www.wowhead.com/forever/spell=23049
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 16392, ItemName: "Knight-Lieutenant's Leather Boots"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the damage absorbed by your Mana Shield by 285.
	// https://www.wowhead.com/forever/spell=23037
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 16440, ItemName: "Marshal's Silk Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the duration of your Sprint ability by 3.0 sec.
	// https://www.wowhead.com/forever/spell=23049
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 16446, ItemName: "Marshal's Leather Footguards"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the damage absorbed by your Mana Shield by 285.
	// https://www.wowhead.com/forever/spell=23037
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 16487, ItemName: "Blood Guard's Silk Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the duration of your Sprint ability by 3.0 sec.
	// https://www.wowhead.com/forever/spell=23049
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 16498, ItemName: "Blood Guard's Leather Treads"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the speed of your Ghost Wolf ability by 15%.
	// https://www.wowhead.com/forever/spell=22801
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 16518, ItemName: "Blood Guard's Mail Walkers"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the damage absorbed by your Mana Shield by 285.
	// https://www.wowhead.com/forever/spell=23037
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 16540, ItemName: "General's Silk Handguards"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the duration of your Sprint ability by 3.0 sec.
	// https://www.wowhead.com/forever/spell=23049
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 16558, ItemName: "General's Leather Treads"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the speed of your Ghost Wolf ability by 15%.
	// https://www.wowhead.com/forever/spell=22801
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 16573, ItemName: "General's Mail Boots"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Poisons target for 8 Nature damage every 2.0 sec for 20s.
	// https://www.wowhead.com/forever/spell=17511
	// unsupported: states no rate
	// trigger 17511 (every time, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 17511, BuffSpellID: 17511, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 17002, ItemName: "Ichor Spitter"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Knocks down all nearby enemies for 3s.
	// https://www.wowhead.com/forever/spell=21152
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 17073, ItemName: "Earthshaker"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Searing Pain.
	// https://www.wowhead.com/forever/spell=23046
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 17564, ItemName: "Knight-Lieutenant's Dreadweave Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Searing Pain.
	// https://www.wowhead.com/forever/spell=23046
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 17577, ItemName: "Blood Guard's Dreadweave Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Searing Pain.
	// https://www.wowhead.com/forever/spell=23046
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 17584, ItemName: "Marshal's Dreadweave Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Searing Pain.
	// https://www.wowhead.com/forever/spell=23046
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 17588, ItemName: "General's Dreadweave Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Mind Blast.
	// https://www.wowhead.com/forever/spell=23043
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 17596, ItemName: "Knight-Lieutenant's Satin Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Mind Blast.
	// https://www.wowhead.com/forever/spell=23043
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 17608, ItemName: "Marshal's Satin Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Mind Blast.
	// https://www.wowhead.com/forever/spell=23043
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 17617, ItemName: "Blood Guard's Satin Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Mind Blast.
	// https://www.wowhead.com/forever/spell=23043
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 17620, ItemName: "General's Satin Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Strike your target with Keeper's Sting for 28 Nature damage.
	// https://www.wowhead.com/forever/spell=29655
	// unsupported: states no rate
	// trigger 29655 (0%, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 29655, BuffSpellID: 29655, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 17753, ItemName: "Verdant Keeper's Aim"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Blasts a target for 90 Fire damage.
	// https://www.wowhead.com/forever/spell=13442
	// unsupported: states no rate
	// trigger 13442 (0%, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 13442, BuffSpellID: 13442, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 18671, ItemName: "Baron Charr's Sceptre"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// While equipped, the wearer suffers less damage from falls.
	// https://www.wowhead.com/forever/spell=23409
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 18951, ItemName: "Evonice's Landin' Pilla"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Reduce your threat to the current target making them less likely to attack you.
	// https://www.wowhead.com/forever/spell=23604
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 19166, ItemName: "Black Amnesty"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives the wearer a 10% chance of being able to resurrect with 20% health and mana.
	// https://www.wowhead.com/forever/spell=23701
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 19290, ItemName: "Darkmoon Card: Twisting Nether"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Stuns target for 1s.
	// https://www.wowhead.com/forever/spell=23454
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 19323, ItemName: "The Unstoppable Force"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the duration of Hammer of Justice by 0.5 sec.
	// https://www.wowhead.com/forever/spell=24188
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 19588, ItemName: "Hero's Brand"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the amount of damage absorbed by Power Word: Shield by 35.
	// https://www.wowhead.com/forever/spell=24191
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 19594, ItemName: "The All-Seeing Eye of Zuldazar"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Reduces the cooldown of Counterspell by 2.0 sec.
	// https://www.wowhead.com/forever/spell=24429
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 19601, ItemName: "Jewel of Kajaro"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the radius of Rain of Fire and Hellfire by 1 yard.
	// https://www.wowhead.com/forever/spell=24430
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 19605, ItemName: "Kezan's Unstoppable Taint"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Decreases the cooldown of Kick by 0.5 sec.
	// https://www.wowhead.com/forever/spell=24434
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 19617, ItemName: "Zandalarian Shadow Mastery Talisman"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Decreases the cooldown of Feign Death by 2.0 sec.
	// https://www.wowhead.com/forever/spell=24432
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 19621, ItemName: "Maelstrom's Wrath"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Run speed increased slightly.
	// https://www.wowhead.com/forever/spell=23990
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 20048, ItemName: "Highlander's Plate Greaves"},
	//	{ItemID: 20127, ItemName: "Highlander's Plate Greaves"},
	//	{ItemID: 20128, ItemName: "Highlander's Plate Greaves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Run speed increased slightly.
	// https://www.wowhead.com/forever/spell=23990
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 20049, ItemName: "Highlander's Lamellar Greaves"},
	//	{ItemID: 20109, ItemName: "Highlander's Lamellar Greaves"},
	//	{ItemID: 20110, ItemName: "Highlander's Lamellar Greaves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Run speed increased slightly.
	// https://www.wowhead.com/forever/spell=23990
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 20050, ItemName: "Highlander's Chain Greaves"},
	//	{ItemID: 20091, ItemName: "Highlander's Chain Greaves"},
	//	{ItemID: 20092, ItemName: "Highlander's Chain Greaves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Run speed increased slightly.
	// https://www.wowhead.com/forever/spell=23990
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 20051, ItemName: "Highlander's Mail Greaves"},
	//	{ItemID: 20121, ItemName: "Highlander's Mail Greaves"},
	//	{ItemID: 20122, ItemName: "Highlander's Mail Greaves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Run speed increased slightly.
	// https://www.wowhead.com/forever/spell=23990
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 20052, ItemName: "Highlander's Leather Boots"},
	//	{ItemID: 20112, ItemName: "Highlander's Leather Boots"},
	//	{ItemID: 20113, ItemName: "Highlander's Leather Boots"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Run speed increased slightly.
	// https://www.wowhead.com/forever/spell=23990
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 20053, ItemName: "Highlander's Lizardhide Boots"},
	//	{ItemID: 20100, ItemName: "Highlander's Lizardhide Boots"},
	//	{ItemID: 20101, ItemName: "Highlander's Lizardhide Boots"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Run speed increased slightly.
	// https://www.wowhead.com/forever/spell=23990
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 20054, ItemName: "Highlander's Cloth Boots"},
	//	{ItemID: 20094, ItemName: "Highlander's Cloth Boots"},
	//	{ItemID: 20095, ItemName: "Highlander's Cloth Boots"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Run speed increased slightly.
	// https://www.wowhead.com/forever/spell=23990
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 20154, ItemName: "Defiler's Chain Greaves"},
	//	{ItemID: 20155, ItemName: "Defiler's Chain Greaves"},
	//	{ItemID: 20156, ItemName: "Defiler's Chain Greaves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Run speed increased slightly.
	// https://www.wowhead.com/forever/spell=23990
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 20159, ItemName: "Defiler's Cloth Boots"},
	//	{ItemID: 20160, ItemName: "Defiler's Cloth Boots"},
	//	{ItemID: 20161, ItemName: "Defiler's Cloth Boots"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Run speed increased slightly.
	// https://www.wowhead.com/forever/spell=23990
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 20167, ItemName: "Defiler's Lizardhide Boots"},
	//	{ItemID: 20168, ItemName: "Defiler's Lizardhide Boots"},
	//	{ItemID: 20170, ItemName: "Defiler's Lizardhide Boots"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Run speed increased slightly.
	// https://www.wowhead.com/forever/spell=23990
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 20181, ItemName: "Defiler's Lamellar Greaves"},
	//	{ItemID: 20183, ItemName: "Defiler's Lamellar Greaves"},
	//	{ItemID: 20185, ItemName: "Defiler's Lamellar Greaves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Run speed increased slightly.
	// https://www.wowhead.com/forever/spell=23990
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 20186, ItemName: "Defiler's Leather Boots"},
	//	{ItemID: 20187, ItemName: "Defiler's Leather Boots"},
	//	{ItemID: 20189, ItemName: "Defiler's Leather Boots"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Run speed increased slightly.
	// https://www.wowhead.com/forever/spell=23990
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 20199, ItemName: "Defiler's Mail Greaves"},
	//	{ItemID: 20200, ItemName: "Defiler's Mail Greaves"},
	//	{ItemID: 20202, ItemName: "Defiler's Mail Greaves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Run speed increased slightly.
	// https://www.wowhead.com/forever/spell=23990
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 20208, ItemName: "Defiler's Plate Greaves"},
	//	{ItemID: 20209, ItemName: "Defiler's Plate Greaves"},
	//	{ItemID: 20211, ItemName: "Defiler's Plate Greaves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Disorients the target, causing it to wander aimlessly for up to 3s.
	// https://www.wowhead.com/forever/spell=26108
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 21134, ItemName: "Dark Edge of Insanity"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// https://www.wowhead.com/forever/spell=1318470
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 21326, ItemName: "Defender of the Timbermaw"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Underwater Breath lasts 50% longer than normal.
	// https://www.wowhead.com/forever/spell=11789
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 21526, ItemName: "Band of Icy Depths"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// When struck by a harmful spell, the caster of that spell has a 5% chance to be silenced for 3s.
	// https://www.wowhead.com/forever/spell=27559
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 22198, ItemName: "Jagged Obsidian Shield"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases healing done by Lesser Healing Wave by up to 80.
	// https://www.wowhead.com/forever/spell=27855
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 22396, ItemName: "Totem of Life"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Reduces the mana cost of your Cleanse spell by 25.
	// https://www.wowhead.com/forever/spell=27847
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 22402, ItemName: "Libram of Grace"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the duration of your Sprint ability by 3.0 sec.
	// https://www.wowhead.com/forever/spell=23049
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 22856, ItemName: "Blood Guard's Leather Walkers"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the speed of your Ghost Wolf ability by 15%.
	// https://www.wowhead.com/forever/spell=22801
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 22857, ItemName: "Blood Guard's Mail Greaves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Searing Pain.
	// https://www.wowhead.com/forever/spell=23046
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 22865, ItemName: "Blood Guard's Dreadweave Handwraps"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Mind Blast.
	// https://www.wowhead.com/forever/spell=23043
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 22869, ItemName: "Blood Guard's Satin Handwraps"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the damage absorbed by your Mana Shield by 285.
	// https://www.wowhead.com/forever/spell=23037
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 22870, ItemName: "Blood Guard's Silk Handwraps"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gain up to 25 mana each time you cast Healing Touch.
	// https://www.wowhead.com/forever/spell=28847
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 23004, ItemName: "Idol of Longevity"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Regain up to 10 mana each time you cast Lesser Healing Wave.
	// https://www.wowhead.com/forever/spell=28849
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 23005, ItemName: "Totem of Flowing Water"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Searing Pain.
	// https://www.wowhead.com/forever/spell=23046
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 23282, ItemName: "Knight-Lieutenant's Dreadweave Handwraps"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the duration of your Sprint ability by 3.0 sec.
	// https://www.wowhead.com/forever/spell=23049
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 23285, ItemName: "Knight-Lieutenant's Leather Walkers"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Mind Blast.
	// https://www.wowhead.com/forever/spell=23043
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 23288, ItemName: "Knight-Lieutenant's Satin Handwraps"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the damage absorbed by your Mana Shield by 285.
	// https://www.wowhead.com/forever/spell=23037
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 23290, ItemName: "Knight-Lieutenant's Silk Handwraps"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Prevents an attack that would otherwise kill you. Triggering this effect also grants you 3s of damage
	// immunity and shatters the phylactery.
	// https://www.wowhead.com/forever/spell=370391
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 191312, ItemName: "Failsafe Phylactery"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// https://www.wowhead.com/forever/spell=463001
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 208424, ItemName: "Sun Shades"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Causes Wrath to have a 50% chance to restore 35 Mana.
	// https://www.wowhead.com/forever/spell=1302521
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 249441, ItemName: "Talons of Wrath"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Reduces the cooldown of Grounding Totem by 1 sec.
	// https://www.wowhead.com/forever/spell=1249011
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 249443, ItemName: "Totem of Ancestral Protection"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Sends a shadowy bolt at the enemy causing 140 Shadow damage.
	// https://www.wowhead.com/forever/spell=16784
	// unsupported: states no rate
	// trigger 16784 (every time, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 16784, BuffSpellID: 16784, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 260186, ItemName: "Darkblade"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Chance on hit to deal 75 Arcane damage. Deals 2 times as much damage to Naga and Satyrs.
	// https://www.wowhead.com/forever/spell=1265634
	// unsupported: states no rate
	// trigger 1318159 (no stated rate, core.CallbackOnSpellHitDealt, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial | core.ProcMaskRangedAuto | core.ProcMaskRangedSpecial | core.ProcMaskSpellDamage) -> buff 1265634
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 1318159, BuffSpellID: 1265634},
	//	[]shared.ItemVariant{
	//	{ItemID: 260205, ItemName: "Highborne Research Tablet"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// https://www.wowhead.com/forever/spell=1318514
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 268873, ItemName: "Defender of the Barkskin"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Harmful spell casts and attacks against Furbolg have a chance to deal 105 Nature damage.
	// https://www.wowhead.com/forever/spell=1318324
	// unsupported: states no rate
	// trigger 1318323 (no stated rate, core.CallbackOnSpellHitDealt, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial | core.ProcMaskRangedAuto | core.ProcMaskRangedSpecial | core.ProcMaskSpellDamage) -> buff 1318324
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 1318323, BuffSpellID: 1318324},
	//	[]shared.ItemVariant{
	//	{ItemID: 269741, ItemName: "Scented Runewood Brooch"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Your Enrage ability generates an additional 10 Rage over its duration.
	// https://www.wowhead.com/forever/spell=1291060
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272428, ItemName: "Enraged Idol"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Each of your heal over time effects on the target reduces the remaining cooldown on Swiftmend by 1 sec
	// when you cast it.
	// https://www.wowhead.com/forever/spell=1291075
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272429, ItemName: "Idol of Synthesis"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the damage absorbed by your Mana Shield by 285.
	// https://www.wowhead.com/forever/spell=23037
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272455, ItemName: "Premier Magus Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the speed of your Ghost Wolf ability by 15%.
	// https://www.wowhead.com/forever/spell=22801
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272484, ItemName: "Premier Mail Walkers"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the damage absorbed by your Mana Shield by 285.
	// https://www.wowhead.com/forever/spell=23037
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272505, ItemName: "Premier Champion's Magus Handguards"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the duration of your Sprint ability by 3.0 sec.
	// https://www.wowhead.com/forever/spell=23049
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272523, ItemName: "Premier Centurion's Shadowhide Treads"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the speed of your Ghost Wolf ability by 15%.
	// https://www.wowhead.com/forever/spell=22801
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272538, ItemName: "Premier Centurion's Linked Boots"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Searing Pain.
	// https://www.wowhead.com/forever/spell=23046
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272553, ItemName: "Premier Felweave Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Searing Pain.
	// https://www.wowhead.com/forever/spell=23046
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272556, ItemName: "Premier Champion's Felweave Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Mind Blast.
	// https://www.wowhead.com/forever/spell=23043
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272569, ItemName: "Premier Satin Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Mind Blast.
	// https://www.wowhead.com/forever/spell=23043
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272572, ItemName: "Premier Champion's Satin Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the duration of your Sprint ability by 3.0 sec.
	// https://www.wowhead.com/forever/spell=23049
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272608, ItemName: "Premier Shadowhide Walkers"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the speed of your Ghost Wolf ability by 15%.
	// https://www.wowhead.com/forever/spell=22801
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272609, ItemName: "Premier Linked Greaves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Flash Heal.
	// https://www.wowhead.com/forever/spell=1293542
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272619, ItemName: "Premier Silken Handwraps"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the damage absorbed by your Mana Shield by 285.
	// https://www.wowhead.com/forever/spell=23037
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272703, ItemName: "Premier Silk Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the duration of your Sprint ability by 3.0 sec.
	// https://www.wowhead.com/forever/spell=23049
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272704, ItemName: "Premier Leather Boots"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the duration of your Repentance by 1 sec.
	// https://www.wowhead.com/forever/spell=1316101
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272719, ItemName: "Premier Chevalier Gauntlets"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the damage absorbed by your Mana Shield by 285.
	// https://www.wowhead.com/forever/spell=23037
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272749, ItemName: "Premier Lieutenant Commander's Silk Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the duration of your Sprint ability by 3.0 sec.
	// https://www.wowhead.com/forever/spell=23049
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272755, ItemName: "Premier Knight-Champion's Leather Footguards"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the duration of your Repentance by 1 sec.
	// https://www.wowhead.com/forever/spell=1316101
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272780, ItemName: "Premier Lieutenant Commander's Chevalier Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Searing Pain.
	// https://www.wowhead.com/forever/spell=23046
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272796, ItemName: "Premier Dreadweave Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Searing Pain.
	// https://www.wowhead.com/forever/spell=23046
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272808, ItemName: "Premier Lieutenant Commander's Dreadweave Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Mind Blast.
	// https://www.wowhead.com/forever/spell=23043
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272812, ItemName: "Premier Voidcloth Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Mind Blast.
	// https://www.wowhead.com/forever/spell=23043
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 272824, ItemName: "Premier Lieutenant Commander's Voidcloth Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the speed of your Ghost Wolf ability by 15%.
	// https://www.wowhead.com/forever/spell=22801
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 273336, ItemName: "Premier Centurion's Mail Boots"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the speed of your Ghost Wolf ability by 15%.
	// https://www.wowhead.com/forever/spell=22801
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 273344, ItemName: "Premier Centurion's Ringmail Boots"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the speed of your Ghost Wolf ability by 15%.
	// https://www.wowhead.com/forever/spell=22801
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 273355, ItemName: "Premier Ringmail Walkers"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Flash Heal.
	// https://www.wowhead.com/forever/spell=1293542
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 273427, ItemName: "Premier Lieutenant Commander's Mooncloth Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Flash Heal.
	// https://www.wowhead.com/forever/spell=1293542
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 273431, ItemName: "Premier Champion's Silken Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Gives you a 50% chance to avoid interruption caused by damage while casting Flash Heal.
	// https://www.wowhead.com/forever/spell=1293542
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 273445, ItemName: "Premier Mooncloth Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the speed of your Ghost Wolf ability by 15%.
	// https://www.wowhead.com/forever/spell=22801
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 274176, ItemName: "Premier Scalemail Walkers"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the speed of your Ghost Wolf ability by 15%.
	// https://www.wowhead.com/forever/spell=22801
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 274183, ItemName: "Premier Knight-Champion's Flatmail Boots"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the speed of your Ghost Wolf ability by 15%.
	// https://www.wowhead.com/forever/spell=22801
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 274192, ItemName: "Premier Flatmail Greaves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the speed of your Ghost Wolf ability by 15%.
	// https://www.wowhead.com/forever/spell=22801
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 274198, ItemName: "Premier Knight-Champion's Scalemail Boots"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the speed of your Ghost Wolf ability by 15%.
	// https://www.wowhead.com/forever/spell=22801
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 274206, ItemName: "Premier Knight-Champion's Linkmail Boots"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the speed of your Ghost Wolf ability by 15%.
	// https://www.wowhead.com/forever/spell=22801
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 274217, ItemName: "Premier Linkmail Walkers"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the duration of your Repentance by 1 sec.
	// https://www.wowhead.com/forever/spell=1316101
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 274227, ItemName: "Premier Scaled Gauntlets"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Increases the duration of your Repentance by 1 sec.
	// https://www.wowhead.com/forever/spell=1316101
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 274236, ItemName: "Premier Champion's Scaled Gloves"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Chance on hit to reduce the mana cost of your next spell cast within 10s by 24.
	// https://www.wowhead.com/forever/spell=1294058
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 274425, ItemName: "White Obsidian Wand"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// While in combat, your healing spells have a chance to grant an ally 300 armor for 10s.
	// https://www.wowhead.com/forever/spell=1294065
	// unsupported: states no rate
	// trigger 1294064 (no stated rate, core.CallbackOnHealDealt, core.ProcMaskSpellHealing) -> buff 1294065
	// shared.NewSpellDataProc(shared.SpellDataProc{TriggerSpellID: 1294064, BuffSpellID: 1294065},
	//	[]shared.ItemVariant{
	//	{ItemID: 274430, ItemName: "Rock Sentinel Slicer"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// https://www.wowhead.com/forever/spell=1318317
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 274761, ItemName: "Parachute-Priest Pager"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Diseases target for 24 Nature damage every 1.0 sec for 7s. Deals 2 times as much damage to Aquatic enemies.
	// https://www.wowhead.com/forever/spell=1295744
	// unsupported: states no rate
	// trigger 1295744 (every time, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 1295744, BuffSpellID: 1295744, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 274963, ItemName: "Rot-Covered Harpoon"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Blunderflame the target for 56 Fire damage.
	// https://www.wowhead.com/forever/spell=1296730
	// unsupported: states no rate
	// trigger 1296730 (0%, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 1296730, BuffSpellID: 1296730, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 275439, ItemName: "Blunderflame Bone Bow"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Cremate the target for 56 Fire damage.
	// https://www.wowhead.com/forever/spell=1297921
	// unsupported: states no rate
	// trigger 1297921 (0%, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 1297921, BuffSpellID: 1297921, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 275441, ItemName: "Cremation Longbow"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Chance on harmful spell cast to reduce target's attack power by 60 for 30s.
	//
	// https://www.wowhead.com/forever/spell=1297082
	// unsupported: states no rate
	// trigger 1297085 (no stated rate, core.CallbackOnCastComplete, core.ProcMaskSpellDamage) -> buff 1297082
	// shared.NewSpellDataDebuffProc(shared.SpellDataProc{TriggerSpellID: 1297085, BuffSpellID: 1297082},
	//	[]shared.ItemVariant{
	//	{ItemID: 275630, ItemName: "Depleted Eye of Influence"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Heals the wielder for 182 damage over 14s.
	// https://www.wowhead.com/forever/spell=1297357
	// unsupported: states no rate
	// trigger 1297357 (0%, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataHealProc(shared.SpellDataProc{TriggerSpellID: 1297357, BuffSpellID: 1297357, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 275645, ItemName: "Reforged Spear"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Your healing spells have a 5% chance to remove 1 Poison effect from the target.
	// https://www.wowhead.com/forever/spell=1297369
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 275647, ItemName: "Cleansed Ritual Kris"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// https://www.wowhead.com/forever/spell=1317283
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 277910, ItemName: "Mysticbloom Seed"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Your Well Fed buffs metabolize 50% slower. This effect is increased by an additional 50% in Wasteland
	// areas.
	// https://www.wowhead.com/forever/spell=1302968
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 277950, ItemName: "Wastewanderer Rations"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Your Lacerate hits have a 10% chance to reset the cooldown on Mangle (Bear).
	// https://www.wowhead.com/forever/spell=1306483
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 279251, ItemName: "Idol of the Ursine Twins"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Blast the target with the fires of a forge, inflicting 51 Fire damage.
	// https://www.wowhead.com/forever/spell=1312176
	// unsupported: states no rate
	// trigger 1312176 (every time, core.CallbackOnSpellHitDealt, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeMHSpecial)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 1312176, BuffSpellID: 1312176, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 281665, ItemName: "Twilight Forgehammer"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Strike the target with a Flaming Shell, inflicting 7 Fire damage every 3 sec for 9 sec.
	// https://www.wowhead.com/forever/spell=1314040
	// unsupported: states no rate
	// trigger 1314040 (0%, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 1314040, BuffSpellID: 1314040, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 282561, ItemName: "Siegebreaker's Blaster"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Reduces the target enemy's attack power by 60 for 30s.
	// https://www.wowhead.com/forever/spell=1315767
	// unsupported: states no rate
	// trigger 1315767 (0%, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDebuffProc(shared.SpellDataProc{TriggerSpellID: 1315767, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 284262, ItemName: "Howling Hide"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Your attacks have a high chance to poison the target for 8 Nature damage every 3.0 sec for 9s.
	// https://www.wowhead.com/forever/spell=1317583
	// unsupported: states no rate
	// trigger 1317583 (every time, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 1317583, BuffSpellID: 1317583, IsWeaponProc: true},
	//	[]shared.ItemVariant{
	//	{ItemID: 285102, ItemName: "Darkmist Dirk"},
	// })

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Stun the target for 1s.
	// https://www.wowhead.com/forever/spell=1318178
	// shared.NewProcStatBonusEffectWithVariants(shared.ProcStatBonusEffect{
	//	Callback:           core.CallbackEmpty,
	//	ProcMask:           core.ProcMaskUnknown,
	//	Outcome:            core.OutcomeEmpty,
	//	RequireDamageDealt: false
	// }, []shared.ItemVariant{
	//	{ItemID: 285281, ItemName: "Arcanite Blacksmith Hammer"},
	// })

	// Smites an enemy for 30 Holy damage.
	// https://www.wowhead.com/forever/spell=13519
	// trigger 13519 (1 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 13519, BuffSpellID: 13519, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 754, ItemName: "Shortsword of Vengeance"},
		})

	// Launches a bolt of frost at the enemy causing 50 Frost damage and slowing movement speed by 50% for 5s.
	// https://www.wowhead.com/forever/spell=13439
	// trigger 13439 (3.5 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 13439, BuffSpellID: 13439, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 810, ItemName: "Hammer of the Northern Wind"},
		})

	// Hurls a fiery ball that causes 176 Fire damage and an additional 24 damage over 6s.
	// https://www.wowhead.com/forever/spell=18796
	// trigger 18796 (1 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 18796, BuffSpellID: 18796, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 870, ItemName: "Fiery War Axe"},
		})

	// Grants 1 extra attack on your next swing.
	// https://www.wowhead.com/forever/spell=18797
	// trigger 18797 (1.9 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataExtraAttackProc(shared.SpellDataProc{TriggerSpellID: 18797, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 871, ItemName: "Flurry Axe"},
		})

	// When struck in combat has a 1% chance of raising a thorny shield that inflicts 3 Nature damage to attackers
	// when hit and increases Nature resistance by 50 for 30s.
	//
	// https://www.wowhead.com/forever/spell=17154
	// trigger 18097 (1%, core.CallbackOnSpellHitTaken, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial) -> buff 17154
	shared.NewSpellDataProc(shared.SpellDataProc{TriggerSpellID: 18097, BuffSpellID: 17154},
		[]shared.ItemVariant{
			{ItemID: 1204, ItemName: "The Green Tower"},
		})

	// Blasts a target for 140 Fire damage.
	// https://www.wowhead.com/forever/spell=1300753
	// trigger 1300753 (1 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 1300753, BuffSpellID: 1300753, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 1728, ItemName: "Teebu's Blazing Longsword"},
		})

	// Sends a shadowy bolt at the enemy causing 200 Shadow damage.
	// https://www.wowhead.com/forever/spell=18211
	// trigger 18211 (1 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 18211, BuffSpellID: 18211, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 1982, ItemName: "Nightblade"},
		})

	// Wounds the target for 108 damage.
	// https://www.wowhead.com/forever/spell=18107
	// trigger 18107 (1 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 18107, BuffSpellID: 18107, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 2164, ItemName: "Gut Ripper"},
		})

	// Fires a Searing Arrow at the target for 28 Fire damage.
	// https://www.wowhead.com/forever/spell=29638
	// trigger 29638 (3.35 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 29638, BuffSpellID: 29638, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 2825, ItemName: "Bow of Searing Arrows"},
		})

	// A burst of energy fills the caster, increasing his damage by 10 and armor by 150 for 15s.
	// https://www.wowhead.com/forever/spell=8191
	// trigger 8191 (1.8 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataProc(shared.SpellDataProc{TriggerSpellID: 8191, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 6622, ItemName: "Sword of Zeal"},
		})

	// Heals wielder of 78 damage over 12s.
	// https://www.wowhead.com/forever/spell=8348
	// trigger 8348 (1 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataHealProc(shared.SpellDataProc{TriggerSpellID: 8348, BuffSpellID: 8348, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 6660, ItemName: "Julie's Dagger"},
		})

	// When struck in combat has a 5% chance of inflicting 140 Shadow damage to the attacker.
	// https://www.wowhead.com/forever/spell=1293421
	// trigger 7619 (5%, core.CallbackOnSpellHitTaken, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial | core.ProcMaskRangedAuto | core.ProcMaskRangedSpecial) -> buff 1293421
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 7619, BuffSpellID: 1293421},
		[]shared.ItemVariant{
			{ItemID: 7747, ItemName: "Vile Protector"},
		})

	// When struck in combat has a 3% chance to heal you for 80.
	// https://www.wowhead.com/forever/spell=9777
	// trigger 9778 (3%, core.CallbackOnSpellHitTaken, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial) -> buff 9777
	shared.NewSpellDataHealProc(shared.SpellDataProc{TriggerSpellID: 9778, BuffSpellID: 9777},
		[]shared.ItemVariant{
			{ItemID: 7939, ItemName: "Truesilver Breastplate"},
		})

	// Decrease the armor of the target by 100 for 20s. While affected, the target cannot stealth or turn invisible.
	// https://www.wowhead.com/forever/spell=9806
	// trigger 9806 (1 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDebuffProc(shared.SpellDataProc{TriggerSpellID: 9806, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 7961, ItemName: "Phantom Blade"},
		})

	// Wounds the target for 75 damage.
	// https://www.wowhead.com/forever/spell=16405
	// trigger 16405 (1 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 16405, BuffSpellID: 16405, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 8190, ItemName: "Hanzo Sword"},
		})

	// Blasts a target for 91 Nature damage.
	// https://www.wowhead.com/forever/spell=18081
	// trigger 18081 (1 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 18081, BuffSpellID: 18081, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 9651, ItemName: "Gryphon Rider's Stormhammer"},
		})

	// 2% chance when struck in melee to gain a holy shield, absorbing 216 damage for 15s. This chance is doubled
	// in Wasteland and Haunted areas.
	// https://www.wowhead.com/forever/spell=10368
	// trigger 8397 (4%, the column over effect 1's 2%, core.CallbackOnSpellHitTaken, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial) -> buff 10368
	shared.NewSpellDataAbsorbProc(shared.SpellDataProc{TriggerSpellID: 8397, BuffSpellID: 10368},
		[]shared.ItemVariant{
			{ItemID: 11302, ItemName: "Uther's Strength"},
		})

	// When struck in combat has a 1% chance of inflicting 100 Shadow damage to the attacker.
	// https://www.wowhead.com/forever/spell=16783
	// trigger 7617 (1%, core.CallbackOnSpellHitTaken, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial | core.ProcMaskRangedAuto | core.ProcMaskRangedSpecial) -> buff 16783
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 7617, BuffSpellID: 16783},
		[]shared.ItemVariant{
			{ItemID: 11861, ItemName: "Girdle of Reprisal"},
		})

	// Blasts a target for 98 Nature damage.
	// https://www.wowhead.com/forever/spell=18089
	// trigger 18089 (1 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 18089, BuffSpellID: 18089, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 11902, ItemName: "Linken's Sword of Mastery"},
		})

	// Adds 4 Fire damage to your weapon attack.
	// https://www.wowhead.com/forever/spell=7714
	// trigger 7721 (every time, core.CallbackOnSpellHitDealt, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial) -> buff 7714
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 7721, BuffSpellID: 7714},
		[]shared.ItemVariant{
			{ItemID: 12631, ItemName: "Fiery Plate Gauntlets"},
		})

	// Adds 3 Lightning damage to your melee attacks.
	// https://www.wowhead.com/forever/spell=16614
	// trigger 16615 (every time, core.CallbackOnSpellHitDealt, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial) -> buff 16614
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 16615, BuffSpellID: 16614},
		[]shared.ItemVariant{
			{ItemID: 12632, ItemName: "Storm Gauntlets"},
		})

	// Burns the enemy for 200 damage over 30s.
	// https://www.wowhead.com/forever/spell=16898
	// trigger 16898 (1 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 16898, BuffSpellID: 16898, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 12777, ItemName: "Blazing Rapier"},
		})

	// Heal yourself for 360 and increase your Strength by 120 for 30s.
	// https://www.wowhead.com/forever/spell=16916
	// trigger 16916 (1 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataProc(shared.SpellDataProc{TriggerSpellID: 16916, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 12790, ItemName: "Arcanite Champion"},
		})

	// Hurls a fiery ball that causes 228 Fire damage and an additional 36 damage over 6s.
	// https://www.wowhead.com/forever/spell=18082
	// trigger 18082 (1 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 18082, BuffSpellID: 18082, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 12792, ItemName: "Volcanic Hammer"},
		})

	// Blasts up to 3 targets for 125 Nature damage.
	// https://www.wowhead.com/forever/spell=16921
	// trigger 16921 (0.5 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 16921, BuffSpellID: 16921, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 12794, ItemName: "Masterwork Stormhammer"},
		})

	// Wounds the target causing them to bleed for 100 damage over 30s.
	// https://www.wowhead.com/forever/spell=13318
	// trigger 13318 (1 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 13318, BuffSpellID: 13318, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 12795, ItemName: "Blood Talon"},
		})

	// Slows the target's movement speed by 30% and increases the time between their attacks by 25% for 5s.
	// https://www.wowhead.com/forever/spell=16927
	// trigger 16927 (1 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDebuffProc(shared.SpellDataProc{TriggerSpellID: 16927, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 12797, ItemName: "Frostguard"},
		})

	// Reduces an enemy's armor by 165. Stacks up to 3 times.
	// https://www.wowhead.com/forever/spell=16928
	// trigger 16928 (1 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDebuffProc(shared.SpellDataProc{TriggerSpellID: 16928, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 12798, ItemName: "Annihilator"},
		})

	// Grants a chance on striking the enemy for 60 Fire damage.
	// https://www.wowhead.com/forever/spell=13441
	// trigger 16982 (1%, core.CallbackOnSpellHitDealt, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial) -> buff 13441
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 16982, BuffSpellID: 13441},
		[]shared.ItemVariant{
			{ItemID: 12805, ItemName: "Orb of Fire"},
		})

	// Poisons target for 8 Nature damage every 2.0 sec for 20s.
	// https://www.wowhead.com/forever/spell=17511
	// trigger 17511 (1 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 17511, BuffSpellID: 17511, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 13035, ItemName: "Serpent Slicer"},
		})

	// Sends a shadowy bolt at the enemy causing 150 Shadow damage.
	// https://www.wowhead.com/forever/spell=18214
	// trigger 18214 (1.25 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 18214, BuffSpellID: 18214, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 13051, ItemName: "Witchfury"},
		})

	// Sends a shadowy bolt at the enemy causing 200 Shadow damage.
	// https://www.wowhead.com/forever/spell=18211
	// trigger 18211 (0.75 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 18211, BuffSpellID: 18211, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 13053, ItemName: "Doombringer"},
		})

	// Wounds the target for 130 damage.
	// https://www.wowhead.com/forever/spell=14126
	// trigger 14126 (0.87 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 14126, BuffSpellID: 14126, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 13054, ItemName: "Grim Reaper"},
		})

	// Wounds the target for 75 damage.
	// https://www.wowhead.com/forever/spell=16405
	// trigger 16405 (3 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 16405, BuffSpellID: 16405, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 13060, ItemName: "The Needler"},
		})

	// Wounds the target causing them to bleed for 80 damage over 30s.
	// https://www.wowhead.com/forever/spell=16406
	// trigger 16406 (1 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 16406, BuffSpellID: 16406, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 15814, ItemName: "Hameya's Slayer"},
		})

	// Adds 2 fire damage to your melee attacks.
	// https://www.wowhead.com/forever/spell=7712
	// trigger 7711 (every time, core.CallbackOnSpellHitDealt, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial) -> buff 7712
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 7711, BuffSpellID: 7712},
		[]shared.ItemVariant{
			{ItemID: 17111, ItemName: "Blazefury Medallion"},
		})

	// Hurls a fiery ball that causes 92 Fire damage and an additional 16 damage over 8s.
	// https://www.wowhead.com/forever/spell=21159
	// trigger 21159 (1 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 21159, BuffSpellID: 21159, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 17193, ItemName: "Sulfuron Hammer"},
		})

	// Blasts up to 3 targets for 200 Nature damage. Each target after the first takes less damage.
	// https://www.wowhead.com/forever/spell=21179
	// trigger 21179 (1.5 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 21179, BuffSpellID: 21179, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 17223, ItemName: "Thunderstrike"},
		})

	// Grants an extra attack on your next swing.
	// https://www.wowhead.com/forever/spell=21919
	// trigger 21919 (1 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataExtraAttackProc(shared.SpellDataProc{TriggerSpellID: 21919, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 17705, ItemName: "Thrash Blade"},
		})

	// When struck in combat has a 5% chance of inflicting 50 Nature damage to the attacker.
	// https://www.wowhead.com/forever/spell=16782
	// trigger 13959 (5%, core.CallbackOnSpellHitTaken, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial | core.ProcMaskRangedAuto | core.ProcMaskRangedSpecial) -> buff 16782
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 13959, BuffSpellID: 16782},
		[]shared.ItemVariant{
			{ItemID: 18825, ItemName: "Grand Marshal's Aegis"},
		})

	// When struck in combat has a 5% chance of inflicting 50 Nature damage to the attacker.
	// https://www.wowhead.com/forever/spell=16782
	// trigger 13959 (5%, core.CallbackOnSpellHitTaken, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial | core.ProcMaskRangedAuto | core.ProcMaskRangedSpecial) -> buff 16782
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 13959, BuffSpellID: 16782},
		[]shared.ItemVariant{
			{ItemID: 18826, ItemName: "High Warlord's Shield Wall"},
		})

	// Blasts a target for 45 Frost damage.
	// https://www.wowhead.com/forever/spell=18398
	// trigger 18398 (1.4 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 18398, BuffSpellID: 18398, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 19099, ItemName: "Glacial Blade"},
		})

	// Blasts a target for 45 Nature damage.
	// https://www.wowhead.com/forever/spell=23592
	// trigger 23592 (1.4 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 23592, BuffSpellID: 23592, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 19100, ItemName: "Electrified Dagger"},
		})

	// Spell damage taken by target increased by 15% for 5s.
	// https://www.wowhead.com/forever/spell=23605
	// trigger 23605 (2 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataAuraProc(shared.SpellDataProc{TriggerSpellID: 23605, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 19169, ItemName: "Nightfall"},
		})

	// Sends a shadowy bolt at the enemy causing 200 Shadow damage.
	// https://www.wowhead.com/forever/spell=18211
	// trigger 18211 (0.4 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 18211, BuffSpellID: 18211, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 19170, ItemName: "Ebon Hand"},
		})

	// 2% chance on successful spellcast to increase your Spirit by 150 for 15s.
	// https://www.wowhead.com/forever/spell=23684
	// trigger 23688 (2%, core.CallbackOnCastComplete, core.ProcMaskSpellDamage | core.ProcMaskSpellHealing) -> buff 23684
	shared.NewSpellDataProc(shared.SpellDataProc{TriggerSpellID: 23688, BuffSpellID: 23684},
		[]shared.ItemVariant{
			{ItemID: 19288, ItemName: "Darkmoon Card: Blue Dragon"},
		})

	// Chance to strike your melee target with lightning for 250 Nature damage.
	// https://www.wowhead.com/forever/spell=23687
	// trigger 23686 (1 ppm, core.CallbackOnSpellHitDealt, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial) -> buff 23687
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 23686, BuffSpellID: 23687},
		[]shared.ItemVariant{
			{ItemID: 19289, ItemName: "Darkmoon Card: Maelstrom"},
		})

	// Wounds the target for 250 damage and increases the casting time of all spells by 25% for 30s.
	// https://www.wowhead.com/forever/spell=1290950
	// trigger 1290950 (0.4 ppm on the item, core.CallbackEmpty, core.ProcMaskUnknown)
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 1290950, BuffSpellID: 1290950, IsWeaponProc: true},
		[]shared.ItemVariant{
			{ItemID: 19324, ItemName: "The Lobotomizer"},
		})

	// Gives a chance when your harmful spells land to increase the damage of your spells and effects by 132
	// for 10s.
	// https://www.wowhead.com/forever/spell=25907
	// trigger 25906 (5%, core.CallbackOnCastComplete, core.ProcMaskSpellDamage) -> buff 25907
	shared.NewSpellDataProc(shared.SpellDataProc{TriggerSpellID: 25906, BuffSpellID: 25907},
		[]shared.ItemVariant{
			{ItemID: 21190, ItemName: "Wrath of Cenarius"},
		})

	// When struck in combat has a 5% chance of inflicting 50 Nature damage to the attacker.
	// https://www.wowhead.com/forever/spell=16782
	// trigger 13959 (5%, core.CallbackOnSpellHitTaken, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial | core.ProcMaskRangedAuto | core.ProcMaskRangedSpecial) -> buff 16782
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 13959, BuffSpellID: 16782},
		[]shared.ItemVariant{
			{ItemID: 272591, ItemName: "Premier High Warlord's Shield Wall"},
		})

	// When struck in combat has a 5% chance of inflicting 50 Nature damage to the attacker.
	// https://www.wowhead.com/forever/spell=16782
	// trigger 13959 (5%, core.CallbackOnSpellHitTaken, core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto | core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeOHSpecial | core.ProcMaskRangedAuto | core.ProcMaskRangedSpecial) -> buff 16782
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 13959, BuffSpellID: 16782},
		[]shared.ItemVariant{
			{ItemID: 272838, ItemName: "Premier Grand Marshal's Aegis"},
		})

	// Thrown attacks explode on impact, causing 17 Fire damage to nearby enemies.
	// https://www.wowhead.com/forever/spell=1318068
	// trigger 1318069 (every time, core.CallbackOnSpellHitDealt, core.ProcMaskRangedAuto) -> buff 1318068
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 1318069, BuffSpellID: 1318068},
		[]shared.ItemVariant{
			{ItemID: 285277, ItemName: "Satchel of Iron Bombs"},
		})

	// Thrown attacks explode on impact, causing 21 Fire damage to nearby enemies.
	// https://www.wowhead.com/forever/spell=1318121
	// trigger 1318123 (every time, core.CallbackOnSpellHitDealt, core.ProcMaskRangedAuto) -> buff 1318121
	shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 1318123, BuffSpellID: 1318121},
		[]shared.ItemVariant{
			{ItemID: 285278, ItemName: "Satchel of Dark Iron Bombs"},
		})

	// Skipped
	// Not simulated: Rod of the Sleepwalker: "Resist Sleep 05" (1292652) - ignored aura type 117
	// https://www.wowhead.com/forever/spell=1292652
	// Not simulated: Mug O' Hurt: "Dazed" (13496) - ignored aura type 33
	// https://www.wowhead.com/forever/spell=13496
	// Not simulated: Girdle of the Blindwatcher: "Stealth Detection 05" (1292149) - ignored aura type 17
	// https://www.wowhead.com/forever/spell=1292149
	// Not simulated: Medal of Courage: "Resist Fear 02" (1324578) - ignored aura type 117
	// https://www.wowhead.com/forever/spell=1324578
	// Not simulated: Mask of Thero-shan: "Stealth 06" (17746) - ignored aura type 154
	// https://www.wowhead.com/forever/spell=17746
	// Not simulated: Nightscape Boots: "Stealth 06" (17746) - ignored aura type 154
	// https://www.wowhead.com/forever/spell=17746
	// Not simulated: Catseye Ultra Goggles: "Stealth Detection 15" (12418) - ignored aura type 17
	// https://www.wowhead.com/forever/spell=12418
	// Not simulated: Carrot on a Stick: "Mount Speed" (13587) - ignored aura type 130
	// https://www.wowhead.com/forever/spell=13587
	// Not simulated: Stronghold Gauntlets: "Immune to Disarm" (7219) - ignored aura type 77
	// https://www.wowhead.com/forever/spell=7219
	// Not simulated: Voice Amplification Modulator: "Resist Silence 07" (19786) - ignored aura type 117
	// https://www.wowhead.com/forever/spell=19786
	// Not simulated: Knight-Lieutenant's Dragonhide Gloves: "Stealth Detection 10" (23217) - ignored aura type 17
	// https://www.wowhead.com/forever/spell=23217
	// Not simulated: Marshal's Dragonhide Gauntlets: "Stealth Detection 10" (23217) - ignored aura type 17
	// https://www.wowhead.com/forever/spell=23217
	// Not simulated: Blood Guard's Dragonhide Gauntlets: "Stealth Detection 10" (23217) - ignored aura type 17
	// https://www.wowhead.com/forever/spell=23217
	// Not simulated: General's Dragonhide Gloves: "Stealth Detection 10" (23217) - ignored aura type 17
	// https://www.wowhead.com/forever/spell=23217
	// Not simulated: Shard of the Defiler: "Echo of Archimonde" (21079) - ignored aura type 56
	// https://www.wowhead.com/forever/spell=21079
	// Not simulated: Mark of Resolution: "Stout Heart" (21958) - ignored aura type 117
	// https://www.wowhead.com/forever/spell=21958
	// Not simulated: The Eye of Divinity: "Eye of Divinity" (23101) - ignored aura type 19
	// https://www.wowhead.com/forever/spell=23101
	// Not simulated: Pirate's Eye Patch: "Fear Resistance 4" (24351) - ignored aura type 117
	// https://www.wowhead.com/forever/spell=24351
	// Not simulated: Bloodvine Lens: "Stealth Detection 10" (23217) - ignored aura type 17
	// https://www.wowhead.com/forever/spell=23217
	// Not simulated: Gnomish Turban of Psychic Might: "Resist Silence 10" (26208) - ignored aura type 117
	// https://www.wowhead.com/forever/spell=26208
	// Not simulated: Darkmantle Boots: "Stealth 08" (27037) - ignored aura type 154
	// https://www.wowhead.com/forever/spell=27037
	// Not simulated: Blood Guard's Dragonhide Grips: "Stealth Detection 10" (23217) - ignored aura type 17
	// https://www.wowhead.com/forever/spell=23217
	// Not simulated: Knight-Lieutenant's Dragonhide Grips: "Stealth Detection 10" (23217) - ignored aura type 17
	// https://www.wowhead.com/forever/spell=23217
	// Not simulated: Premier Kodohide Gauntlets: "Stealth Detection 10" (23217) - ignored aura type 17
	// https://www.wowhead.com/forever/spell=23217
	// Not simulated: Premier Champion's Dragonhide Gloves: "Stealth Detection 10" (23217) - ignored aura type 17
	// https://www.wowhead.com/forever/spell=23217
	// Not simulated: Premier Dragonhide Grips: "Stealth Detection 10" (23217) - ignored aura type 17
	// https://www.wowhead.com/forever/spell=23217
	// Not simulated: Premier Beasthide Gloves: "Stealth Detection 10" (23217) - ignored aura type 17
	// https://www.wowhead.com/forever/spell=23217
	// Not simulated: Premier Lieutenant Commander's Beasthide Gauntlets: "Stealth Detection 10" (23217) - ignored aura type
	// Not simulated: 17
	// https://www.wowhead.com/forever/spell=23217
	// Not simulated: Cloak of Hermitic Bliss: "-25% Movement Speed" (1291748) - ignored aura type 33
	// https://www.wowhead.com/forever/spell=1291748
	// Not simulated: Premier Lunarhide Gloves: "Stealth Detection 10" (23217) - ignored aura type 17
	// https://www.wowhead.com/forever/spell=23217
	// Not simulated: Premier Dreamhide Gloves: "Stealth Detection 10" (23217) - ignored aura type 17
	// https://www.wowhead.com/forever/spell=23217
	// Not simulated: Premier Wyrmhide Gauntlets: "Stealth Detection 10" (23217) - ignored aura type 17
	// https://www.wowhead.com/forever/spell=23217
	// Not simulated: Premier Lieutenant Commander's Lunarhide Gauntlets: "Stealth Detection 10" (23217) - ignored aura type
	// Not simulated: 17
	// https://www.wowhead.com/forever/spell=23217
	// Not simulated: Premier Lieutenant Commander's Dreamhide Gauntlets: "Stealth Detection 10" (23217) - ignored aura type
	// Not simulated: 17
	// https://www.wowhead.com/forever/spell=23217
	// Not simulated: Premier Champion's Wyrmhide Gloves: "Stealth Detection 10" (23217) - ignored aura type 17
	// https://www.wowhead.com/forever/spell=23217
	// Not simulated: Premier Champion's Kodohide Gloves: "Stealth Detection 10" (23217) - ignored aura type 17
	// https://www.wowhead.com/forever/spell=23217
	// Not simulated: Sorcerer Collar: "Resist Silence 04" (1292222) - ignored aura type 117
	// https://www.wowhead.com/forever/spell=1292222
	// Not simulated: Blindwatcher's Sight: "Resist Disorient 07" (1292268) - ignored aura type 117
	// https://www.wowhead.com/forever/spell=1292268
	// Not simulated: Nightskulker Ring: "Resist Fear 04" (1292574) - ignored aura type 117
	// https://www.wowhead.com/forever/spell=1292574
	// Not simulated: Skullduggery Belt: "Stealth 05" (1293068) - ignored aura type 154
	// https://www.wowhead.com/forever/spell=1293068
	// Not simulated: Field Agent Beverage: "Resist Charm 06" (1297455) - ignored aura type 117
	// https://www.wowhead.com/forever/spell=1297455
	// Not simulated: Wintersaber Hide Lined Gloves: "Mount Speed" (1315778) - ignored aura type 130
	// https://www.wowhead.com/forever/spell=1315778
	// Not simulated: Mithril Blacksmith Hammer: "Concussed" (1318163) - ignored aura type 33
	// https://www.wowhead.com/forever/spell=1318163
}

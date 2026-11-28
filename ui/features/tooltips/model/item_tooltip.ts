import { ArmorType, Class, HandType, ItemType, RangedWeaponType, SpellSchool, Stat, WeaponType } from '@generated/proto/common';
import { RepLevel, type UIEnchant as Enchant, type UIItem as Item } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { translateArmorType, translateRangedWeaponType, translateSpellSchool, translateStat, translateWeaponType } from '@i18n/localization';
import {
	BLOCK_RATING_PER_BLOCK_PERCENT,
	CRIT_RATING_PER_CRIT_PERCENT,
	DEFENSE_RATING_PER_DEFENSE_LEVEL,
	DODGE_RATING_PER_DODGE_PERCENT,
	HASTE_RATING_PER_HASTE_PERCENT,
	HIT_RATING_PER_HIT_PERCENT,
	PARRY_RATING_PER_PARRY_PERCENT,
} from '@sim/constants/mechanics';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { classNames, REP_LEVEL_NAMES } from '@sim/proto/names';
import type { ItemTooltipData, SetTooltipData } from '@sim/proto/tooltips';
import type { GameTooltipLine } from '@ui-kit/GameTooltip';

export interface ItemTooltipInput {
	equipped: EquippedItem;
	data?: ItemTooltipData;
	set?: SetTooltipData;
	/** The items the player wears, which the set counts. */
	worn: ReadonlySet<number>;
	/** The applied enchant as the game names it on the item, and its icon. */
	enchantName?: string;
	enchantIcon?: string;
	itemName: (itemId: number) => string;
	/** A spell's text worked out for the player; nothing for a spell without text. */
	spellText: (spellId: number) => string | undefined;
}

const t = (key: string, values: Record<string, unknown> = {}): string => i18n.t(`tooltips.${key}`, { defaultValue: key, ...values });

const SLOT_KEYS: Partial<Record<ItemType, string>> = {
	[ItemType.ItemTypeHead]: 'head',
	[ItemType.ItemTypeNeck]: 'neck',
	[ItemType.ItemTypeShoulder]: 'shoulder',
	[ItemType.ItemTypeBack]: 'back',
	[ItemType.ItemTypeChest]: 'chest',
	[ItemType.ItemTypeWrist]: 'wrist',
	[ItemType.ItemTypeHands]: 'hands',
	[ItemType.ItemTypeWaist]: 'waist',
	[ItemType.ItemTypeLegs]: 'legs',
	[ItemType.ItemTypeFeet]: 'feet',
	[ItemType.ItemTypeFinger]: 'finger',
	[ItemType.ItemTypeTrinket]: 'trinket',
	[ItemType.ItemTypeRanged]: 'ranged',
};

const HAND_KEYS: Partial<Record<HandType, string>> = {
	[HandType.HandTypeOneHand]: 'one_hand',
	[HandType.HandTypeMainHand]: 'main_hand',
	[HandType.HandTypeOffHand]: 'off_hand',
	[HandType.HandTypeTwoHand]: 'two_hand',
};

// The slot an item the item database does not hold goes in, by its inventory type. A bag and a quiver write their
// slots instead, and a cloak writes no armor type.
const INVENTORY_SLOT_KEYS: Record<number, string> = {
	1: 'head',
	2: 'neck',
	3: 'shoulder',
	4: 'shirt',
	5: 'chest',
	6: 'waist',
	7: 'legs',
	8: 'feet',
	9: 'wrist',
	10: 'hands',
	11: 'finger',
	12: 'trinket',
	13: 'one_hand',
	14: 'off_hand',
	15: 'ranged',
	16: 'back',
	17: 'two_hand',
	19: 'tabard',
	20: 'chest',
	21: 'main_hand',
	22: 'off_hand',
	23: 'held_in_off_hand',
	24: 'ammo',
	25: 'thrown',
	26: 'ranged',
	28: 'relic',
};
const INVENTORY_TYPE_BACK = 16;

const BINDING_KEYS: Record<number, string> = { 1: 'picked_up', 2: 'equipped', 3: 'used' };

const WHITE_STATS = [
	Stat.StatStrength,
	Stat.StatAgility,
	Stat.StatStamina,
	Stat.StatIntellect,
	Stat.StatSpirit,
	Stat.StatArcaneResistance,
	Stat.StatFireResistance,
	Stat.StatFrostResistance,
	Stat.StatNatureResistance,
	Stat.StatShadowResistance,
];

// The ratings the game writes as a share of a percent: 14 crit rating is 1%.
const PERCENT_STATS: Array<[Stat, string, number]> = [
	[Stat.StatHitRating, 'hit', HIT_RATING_PER_HIT_PERCENT],
	[Stat.StatCritRating, 'crit', CRIT_RATING_PER_CRIT_PERCENT],
	[Stat.StatHasteRating, 'haste', HASTE_RATING_PER_HASTE_PERCENT],
	[Stat.StatDodgeRating, 'dodge', DODGE_RATING_PER_DODGE_PERCENT],
	[Stat.StatParryRating, 'parry', PARRY_RATING_PER_PARRY_PERCENT],
	[Stat.StatBlockRating, 'block', BLOCK_RATING_PER_BLOCK_PERCENT],
];

const SCHOOL_STATS: Array<[Stat, SpellSchool]> = [
	[Stat.StatArcaneDamage, SpellSchool.SpellSchoolArcane],
	[Stat.StatFireDamage, SpellSchool.SpellSchoolFire],
	[Stat.StatFrostDamage, SpellSchool.SpellSchoolFrost],
	[Stat.StatHolyDamage, SpellSchool.SpellSchoolHoly],
	[Stat.StatNatureDamage, SpellSchool.SpellSchoolNature],
	[Stat.StatShadowDamage, SpellSchool.SpellSchoolShadow],
];

const OTHER_STATS = [Stat.StatExpertiseRating, Stat.StatArmorPenetration, Stat.StatHealth, Stat.StatMana];

const number = (value: number) => String(Number(value.toFixed(2)));
// A share as the game writes one on an item, to one decimal: 1.0%.
const percent = (value: number) => value.toFixed(1);

/** The time in an item spell's cooldown the way the game writes it: 30 Sec, 5 Min, 1 Hour. */
const cooldownText = (ms: number) => {
	const seconds = ms / 1000;
	if (seconds >= 7200) return t('time.hours', { n: number(seconds / 3600) });
	if (seconds >= 3600) return t('time.hour', { n: number(seconds / 3600) });
	if (seconds >= 60) return t('time.min', { n: number(seconds / 60) });
	return t('time.sec', { n: number(seconds) });
};

const slotLines = (item: Item): Array<GameTooltipLine> => {
	if (item.weaponType === WeaponType.WeaponTypeOffHand) return [{ left: t('slots.held_in_off_hand') }];
	const slotKey = item.type === ItemType.ItemTypeWeapon ? HAND_KEYS[item.handType] : SLOT_KEYS[item.type];
	let type: string | undefined;
	if (item.type === ItemType.ItemTypeWeapon && item.weaponType !== WeaponType.WeaponTypeUnknown) type = translateWeaponType(item.weaponType);
	else if (item.type === ItemType.ItemTypeRanged && item.rangedWeaponType !== RangedWeaponType.RangedWeaponTypeUnknown)
		type = translateRangedWeaponType(item.rangedWeaponType);
	else if (item.armorType !== ArmorType.ArmorTypeUnknown && item.type !== ItemType.ItemTypeBack) type = translateArmorType(item.armorType);
	return slotKey || type ? [{ left: slotKey && t(`slots.${slotKey}`), right: type }] : [];
};

// A weapon's damage, with its school where it is not physical, as a wand's "35 - 66 Fire Damage".
const weaponLines = (item: Item): Array<GameTooltipLine> => {
	if (!item.weaponSpeed || !item.weaponDamageMax) return [];
	const dps = (item.weaponDamageMin + item.weaponDamageMax) / 2 / item.weaponSpeed;
	const range = { min: Math.round(item.weaponDamageMin), max: Math.round(item.weaponDamageMax) };
	return [
		{
			left: item.spellSchool ? t('damage_school', { ...range, school: translateSpellSchool(item.spellSchool) }) : t('damage', range),
			right: t('speed', { speed: item.weaponSpeed.toFixed(2) }),
		},
		{ left: t('dps', { dps: dps.toFixed(2) }) },
	];
};

// The lines the game writes for an item's bonus stats: the armor, a shield's block, the primary stats and
// resistances in white, the rest as green Equip lines, which the game writes below the requirements. An attack power
// bonus is melee and ranged both (item stat 38), so the ranged line carries only a ranged bonus beyond it, and a
// healing line only healing beyond the spell damage that also heals.
const statLines = (equipped: EquippedItem): { white: Array<GameTooltipLine>; green: Array<GameTooltipLine> } => {
	const stats = equipped.calcStats();
	const value = (stat: Stat) => stats.getStat(stat);
	const white: Array<GameTooltipLine> = [];
	const green: Array<GameTooltipLine> = [];
	const armor = value(Stat.StatArmor) + value(Stat.StatBonusArmor);
	if (armor > 0) white.push({ left: t('armor', { armor: Math.round(armor) }) });
	const shield = equipped.item.weaponType === WeaponType.WeaponTypeShield;
	if (shield && value(Stat.StatBlockValue) > 0) white.push({ left: t('block', { value: Math.round(value(Stat.StatBlockValue)) }) });
	for (const stat of WHITE_STATS) {
		if (value(stat) > 0) white.push({ left: t('stat', { value: Math.round(value(stat)), stat: translateStat(stat) }) });
	}

	const equip = (key: string, values: Record<string, unknown>) => green.push({ left: t('equip', { text: t(`equip_stats.${key}`, values) }), tone: 'green' });
	for (const [stat, key, perPercent] of PERCENT_STATS) {
		if (value(stat) > 0) equip(key, { value: percent(value(stat) / perPercent) });
	}
	if (value(Stat.StatDefenseRating) > 0) equip('defense', { value: number(value(Stat.StatDefenseRating) / DEFENSE_RATING_PER_DEFENSE_LEVEL) });
	const attackPower = value(Stat.StatAttackPower);
	if (attackPower > 0) equip('attack_power', { value: Math.round(attackPower) });
	if (value(Stat.StatRangedAttackPower) > attackPower) equip('ranged_attack_power', { value: Math.round(value(Stat.StatRangedAttackPower) - attackPower) });
	if (value(Stat.StatFeralAttackPower) > 0) equip('feral_attack_power', { value: Math.round(value(Stat.StatFeralAttackPower)) });
	const spellDamage = value(Stat.StatSpellDamage);
	if (spellDamage > 0) equip('spell_damage', { value: Math.round(spellDamage) });
	for (const [stat, school] of SCHOOL_STATS) {
		if (value(stat) > 0) equip('school_damage', { value: Math.round(value(stat)), school: translateSpellSchool(school) });
	}
	if (value(Stat.StatHealingPower) > spellDamage) equip('healing', { value: Math.round(value(Stat.StatHealingPower)) });
	if (value(Stat.StatMP5) > 0) equip('mp5', { value: number(value(Stat.StatMP5)) });
	if (!shield && value(Stat.StatBlockValue) > 0) equip('block_value', { value: Math.round(value(Stat.StatBlockValue)) });
	if (value(Stat.StatSpellPiercing) > 0) equip('spell_piercing', { value: Math.round(value(Stat.StatSpellPiercing)) });
	for (const stat of OTHER_STATS) {
		if (value(stat) > 0) equip('other', { value: Math.round(value(stat)), stat: translateStat(stat) });
	}
	return { white, green };
};

const bindingLines = (data: ItemTooltipData | undefined): Array<GameTooltipLine> =>
	data?.b && BINDING_KEYS[data.b] ? [{ left: t(`binding.${BINDING_KEYS[data.b]}`) }] : [];

const uniqueLines = (unique: number | undefined): Array<GameTooltipLine> => {
	if (!unique) return [];
	if (unique < 0) return [{ left: t('unique_equipped') }];
	return [{ left: unique > 1 ? t('unique_count', { n: unique }) : t('unique') }];
};

// What the item asks of the player, in the game's order: the classes, the level, a profession's skill, a standing.
const requirementLines = (classes: ReadonlyArray<Class>, level: number, data: ItemTooltipData | undefined): Array<GameTooltipLine> => {
	const lines: Array<GameTooltipLine> = [];
	const names = classes.filter(cls => cls !== Class.ClassUnknown).map(cls => classNames.get(cls) ?? '');
	if (names.length) lines.push({ left: t('classes', { classes: names.join(', ') }) });
	if (level > 1) lines.push({ left: t('requires_level', { level }) });
	if (data?.k) lines.push({ left: data.k.r ? t('requires_skill', { skill: data.k.n, rank: data.k.r }) : t('requires', { name: data.k.n }) });
	if (data?.g) {
		const standing = REP_LEVEL_NAMES[((data.g.r ?? 0) + 1) as RepLevel];
		lines.push({ left: t('requires_reputation', { faction: data.g.n, standing }) });
	}
	return lines;
};

const spellLines = (data: ItemTooltipData | undefined, spellText: (spellId: number) => string | undefined): Array<GameTooltipLine> =>
	(data?.e ?? []).flatMap(([trigger, spellId, cooldownMs, charges]) => {
		const text = spellText(spellId);
		if (!text) return [];
		const lines: Array<GameTooltipLine> = [];
		if (trigger === 0) {
			const cooldown = cooldownMs > 0 ? ` ${t('cooldown', { time: cooldownText(cooldownMs) })}` : '';
			lines.push({ left: t('use', { text }) + cooldown, tone: 'green' });
		} else {
			lines.push({ left: t(trigger === 2 ? 'chance_on_hit' : 'equip', { text }), tone: 'green' });
		}
		if (charges) lines.push({ left: t('charges', { n: charges }) });
		return lines;
	});

// A set as the game lays it out: a blank line, its name with the count worn and each piece, lit when worn, then a
// blank line and the bonuses, green without their count once active and grey with it until then.
const setLines = ({ set, worn, itemName, spellText }: ItemTooltipInput): Array<GameTooltipLine> => {
	if (!set) return [];
	const count = set.p.filter(id => worn.has(id)).length;
	const lines: Array<GameTooltipLine> = [{ left: t('set', { name: set.n, worn: count, total: set.p.length }), tone: 'gold', spaced: true }];
	for (const id of set.p) lines.push({ left: itemName(id), tone: worn.has(id) ? 'yellow' : 'grey', indent: true });
	const pieceLines = lines.length;
	for (const [pieces, spellId] of set.b) {
		const text = spellText(spellId);
		if (!text) continue;
		const active = count >= pieces;
		lines.push({
			left: active ? t('set_bonus_active', { text }) : t('set_bonus', { count: pieces, text }),
			tone: active ? 'green' : 'grey',
			spaced: lines.length === pieceLines,
		});
	}
	return lines;
};

// The flavour text, then the price a vendor pays, which ends the game's tooltip.
const closingLines = (data: ItemTooltipData | undefined): Array<GameTooltipLine> => [
	...(data?.f ? [{ left: `"${data.f}"`, tone: 'gold' as const }] : []),
	...(data?.s ? [{ left: t('sell_price'), money: data.s }] : []),
];

// What an enchant adds, for one the tooltip files do not name: its stats, such as "+7 Agility", or for one without, its
// name without the recipe's "Enchant Weapon - " (Crusader).
export const enchantText = (enchant: Enchant) => {
	const stats = enchant.stats.flatMap((value, stat) => (value > 0 ? [t('stat', { value: Math.round(value), stat: translateStat(stat) })] : []));
	if (enchant.weaponDamage > 0) stats.push(t('stat', { value: Math.round(enchant.weaponDamage), stat: t('weapon_damage') }));
	return stats.length ? stats.join(', ') : enchant.name.replace(/^Enchant [^-]+ - /, '');
};

/**
 * An item's tooltip below its name, line by line as the game lays it out: what it is, its stats and enchant, what it
 * asks of the player, its Equip and Use lines, its set, its flavour text and its price.
 */
export const itemTooltipLines = (input: ItemTooltipInput): Array<GameTooltipLine> => {
	const { equipped, data } = input;
	const item = equipped.item;
	const stats = statLines(equipped);
	const lines: Array<GameTooltipLine> = [{ left: t('item_level', { ilvl: equipped.ilvl }), tone: 'gold' }];
	lines.push(...bindingLines(data), ...uniqueLines(data ? data.u : Number(item.unique)));
	lines.push(...slotLines(item), ...weaponLines(item), ...stats.white);
	if (equipped.enchant) {
		const name = input.enchantName ?? enchantText(equipped.enchant);
		lines.push({ left: t('enchanted', { name }), tone: 'green', icon: input.enchantIcon });
	}
	lines.push(...requirementLines(item.classAllowlist, item.reqLvl, data), ...stats.green);
	lines.push(...spellLines(data, input.spellText), ...setLines(input), ...closingLines(data));
	return lines;
};

/** The tooltip of an item the item database does not hold, a consumable or an enchant's item: the parts the tooltips carry. */
export const otherItemTooltipLines = (data: ItemTooltipData, spellText: (spellId: number) => string | undefined): Array<GameTooltipLine> => {
	const lines: Array<GameTooltipLine> = [];
	// The game writes an item level only for what goes in a slot, which a potion does not.
	if (data.l && data.o) lines.push({ left: t('item_level', { ilvl: data.l }), tone: 'gold' });
	lines.push(...bindingLines(data), ...uniqueLines(data.u));
	const slotKey = data.o ? INVENTORY_SLOT_KEYS[data.o] : undefined;
	if (data.c && data.t) lines.push({ left: t('container', { slots: data.c, type: data.t }) });
	else if (slotKey) lines.push({ left: t(`slots.${slotKey}`), right: data.o === INVENTORY_TYPE_BACK ? undefined : data.t });
	if (data.d) lines.push({ left: t('ammo_dps', { dps: number(data.d) }) });
	lines.push(...requirementLines(data.a ?? [], data.r ?? 0, data), ...spellLines(data, spellText), ...closingLines(data));
	return lines;
};

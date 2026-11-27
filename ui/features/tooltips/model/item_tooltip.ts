import { ArmorType, Class, HandType, ItemType, RangedWeaponType, SpellSchool, Stat, WeaponType } from '@generated/proto/common';
import type { UIEnchant as Enchant, UIItem as Item } from '@generated/proto/ui';
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
import { classNames } from '@sim/proto/names';
import type { ItemTooltipData, SetTooltipData } from '@sim/proto/tooltips';
import type { GameTooltipLine } from '@ui-kit/GameTooltip';

export interface ItemTooltipInput {
	equipped: EquippedItem;
	data?: ItemTooltipData;
	set?: SetTooltipData;
	/** The items the player wears, which the set counts. */
	worn: ReadonlySet<number>;
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

/** The time in an item spell's cooldown the way the game writes it: 30 Sec, 5 Min, 1 Hour. */
const cooldownText = (ms: number) => {
	const seconds = ms / 1000;
	if (seconds >= 7200) return t('time.hours', { n: number(seconds / 3600) });
	if (seconds >= 3600) return t('time.hour', { n: number(seconds / 3600) });
	if (seconds >= 60) return t('time.min', { n: number(seconds / 60) });
	return t('time.sec', { n: number(seconds) });
};

const slotLines = (item: Item): Array<GameTooltipLine> => {
	const slotKey = item.type === ItemType.ItemTypeWeapon ? HAND_KEYS[item.handType] : SLOT_KEYS[item.type];
	let type: string | undefined;
	if (item.type === ItemType.ItemTypeWeapon && item.weaponType !== WeaponType.WeaponTypeUnknown) type = translateWeaponType(item.weaponType);
	else if (item.type === ItemType.ItemTypeRanged && item.rangedWeaponType !== RangedWeaponType.RangedWeaponTypeUnknown)
		type = translateRangedWeaponType(item.rangedWeaponType);
	else if (item.armorType !== ArmorType.ArmorTypeUnknown) type = translateArmorType(item.armorType);
	return slotKey || type ? [{ left: slotKey && t(`slots.${slotKey}`), right: type }] : [];
};

const weaponLines = (item: Item): Array<GameTooltipLine> => {
	if (!item.weaponSpeed || !item.weaponDamageMax) return [];
	const dps = (item.weaponDamageMin + item.weaponDamageMax) / 2 / item.weaponSpeed;
	return [
		{
			left: t('damage', { min: Math.round(item.weaponDamageMin), max: Math.round(item.weaponDamageMax) }),
			right: t('speed', { speed: item.weaponSpeed.toFixed(2) }),
		},
		{ left: t('dps', { dps: dps.toFixed(2) }) },
	];
};

// The lines the game writes for an item's bonus stats: the armor, primary stats and resistances in white, the rest as
// green Equip lines. An attack power bonus is melee and ranged both (item stat 38), so the ranged line carries only a
// ranged bonus beyond it, and a healing line only healing beyond the spell damage that also heals.
const statLines = (equipped: EquippedItem): Array<GameTooltipLine> => {
	const stats = equipped.calcStats();
	const value = (stat: Stat) => stats.getStat(stat);
	const lines: Array<GameTooltipLine> = [];
	const armor = value(Stat.StatArmor) + value(Stat.StatBonusArmor);
	if (armor > 0) lines.push({ left: t('armor', { armor: Math.round(armor) }) });
	for (const stat of WHITE_STATS) {
		if (value(stat) > 0) lines.push({ left: t('stat', { value: Math.round(value(stat)), stat: translateStat(stat) }) });
	}

	const equip = (key: string, values: Record<string, unknown>) => lines.push({ left: t('equip', { text: t(`equip_stats.${key}`, values) }), tone: 'green' });
	for (const [stat, key, perPercent] of PERCENT_STATS) {
		if (value(stat) > 0) equip(key, { value: number(value(stat) / perPercent) });
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
	if (value(Stat.StatBlockValue) > 0) equip('block_value', { value: Math.round(value(Stat.StatBlockValue)) });
	if (value(Stat.StatSpellPiercing) > 0) equip('spell_piercing', { value: Math.round(value(Stat.StatSpellPiercing)) });
	for (const stat of OTHER_STATS) {
		if (value(stat) > 0) equip('other', { value: Math.round(value(stat)), stat: translateStat(stat) });
	}
	return lines;
};

const spellLines = (data: ItemTooltipData | undefined, spellText: ItemTooltipInput['spellText']): Array<GameTooltipLine> =>
	(data?.e ?? []).flatMap(([trigger, spellId, cooldownMs]) => {
		const text = spellText(spellId);
		if (!text) return [];
		if (trigger === 0) {
			const cooldown = cooldownMs > 0 ? ` ${t('cooldown', { time: cooldownText(cooldownMs) })}` : '';
			return [{ left: t('use', { text }) + cooldown, tone: 'green' as const }];
		}
		return [{ left: t(trigger === 2 ? 'chance_on_hit' : 'equip', { text }), tone: 'green' as const }];
	});

const setLines = ({ set, worn, itemName, spellText }: ItemTooltipInput): Array<GameTooltipLine> => {
	if (!set) return [];
	const count = set.p.filter(id => worn.has(id)).length;
	const lines: Array<GameTooltipLine> = [{ left: t('set', { name: set.n, worn: count, total: set.p.length }), tone: 'gold', spaced: true }];
	for (const id of set.p) lines.push({ left: itemName(id), tone: worn.has(id) ? 'white' : 'grey', indent: true });
	set.b.forEach(([pieces, spellId], index) => {
		const text = spellText(spellId);
		if (text) lines.push({ left: t('set_bonus', { count: pieces, text }), tone: count >= pieces ? 'green' : 'grey', spaced: index === 0 });
	});
	return lines;
};

// An enchant as the game writes it on the item: its stats, such as "+7 Agility", or for one without, its name without the
// recipe's "Enchant Weapon - " (Crusader).
const enchantText = (enchant: Enchant) => {
	const stats = enchant.stats.flatMap((value, stat) => (value > 0 ? [t('stat', { value: Math.round(value), stat: translateStat(stat) })] : []));
	if (enchant.weaponDamage > 0) stats.push(t('stat', { value: Math.round(enchant.weaponDamage), stat: t('weapon_damage') }));
	return stats.length ? stats.join(', ') : enchant.name.replace(/^Enchant [^-]+ - /, '');
};

/** An item's tooltip below its name, line by line as the game lays it out: what it is, its stats, its spells, its set. */
export const itemTooltipLines = (input: ItemTooltipInput): Array<GameTooltipLine> => {
	const { equipped, data } = input;
	const item = equipped.item;
	const lines: Array<GameTooltipLine> = [{ left: t('item_level', { ilvl: equipped.ilvl }), tone: 'gold' }];
	if (data?.b && BINDING_KEYS[data.b]) lines.push({ left: t(`binding.${BINDING_KEYS[data.b]}`) });
	if (item.unique) lines.push({ left: t('unique') });
	lines.push(...slotLines(item), ...weaponLines(item), ...statLines(equipped));
	if (equipped.enchant) lines.push({ left: enchantText(equipped.enchant), tone: 'green' });
	if (item.reqLvl > 1) lines.push({ left: t('requires_level', { level: item.reqLvl }) });
	const classes = item.classAllowlist.filter(cls => cls !== Class.ClassUnknown).map(cls => classNames.get(cls) ?? '');
	if (classes.length) lines.push({ left: t('classes', { classes: classes.join(', ') }) });
	lines.push(...spellLines(data, input.spellText), ...setLines(input));
	if (data?.f) lines.push({ left: `"${data.f}"`, tone: 'gold', spaced: true });
	return lines;
};

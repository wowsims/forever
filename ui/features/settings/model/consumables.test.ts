import { ConsumableType, PseudoStat, Stat } from '@generated/proto/common';
import { Consumable } from '@generated/proto/db';
import i18n from '@i18n/config';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';

import english from '../../../../assets/locales/en/translation.json';
import { elixirGroups } from './consumables';

const elixir = (name: string, stats: Partial<Record<Stat, number>>, crit = false) =>
	Consumable.create({
		name,
		type: ConsumableType.ConsumableTypeElixir,
		stats: Object.assign(new Array<number>(42).fill(0), stats),
		pseudoStats: crit
			? Object.assign(new Array<number>(36).fill(0), {
					[PseudoStat.PseudoStatMeleeCritPercent]: 2,
					[PseudoStat.PseudoStatSpellCritPercent]: 2,
					[PseudoStat.PseudoStatRangedCritPercent]: 2,
				})
			: [],
	});

const ELIXIRS = [
	elixir('Bloodkelp Elixir of Dodging', {}),
	elixir('Bogling Root', { [Stat.StatPhysicalDamage]: 1 }),
	elixir('Elixir of the Phalanx', { [Stat.StatHealth]: 400, [Stat.StatArmorNoMods]: 500 }),
	elixir('Elixir of Fortitude', { [Stat.StatHealth]: 200 }),
	elixir('Greater Mageblood Elixir', { [Stat.StatMP5]: 20 }),
	elixir('Elixir of Fire Power', { [Stat.StatFireDamage]: 10 }),
	elixir('Arcane Elixir', { [Stat.StatSpellDamage]: 20 }),
	elixir('Elixir of the Sages', { [Stat.StatIntellect]: 18, [Stat.StatSpirit]: 18 }),
	elixir('Elixir of Cunning', { [Stat.StatAgility]: 25, [Stat.StatIntellect]: 25 }),
	elixir('Elixir of Ferocity', { [Stat.StatStrength]: 18, [Stat.StatAgility]: 18 }),
	elixir('Juju Power', { [Stat.StatStrength]: 30 }),
	elixir('Elixir of the Mongoose', { [Stat.StatAgility]: 25 }, true),
	elixir('Elixir of the Owl', { [Stat.StatIntellect]: 25 }, true),
];

// The test setup loads no locale, so the group names would read as their keys.
beforeAll(() => {
	i18n.addResourceBundle('en', 'translation', english, true, true);
});
afterAll(() => {
	i18n.removeResourceBundle('en', 'translation');
});

describe('elixirGroups', () => {
	const grouped = () => elixirGroups(ELIXIRS, consumable => consumable).map(group => [group.label, group.items.map(consumable => consumable.name)] as const);

	it('lists the crit elixirs first, then one line per stat under the first stat each grants, and those granting nothing read last', () => {
		expect(grouped()).toEqual([
			['Critical Strike', ['Elixir of the Mongoose', 'Elixir of the Owl']],
			['Strength', ['Elixir of Ferocity', 'Juju Power']],
			['Agility', ['Elixir of Cunning']],
			['Intellect', ['Elixir of the Sages']],
			['Spell Power', ['Arcane Elixir']],
			['Fire Spell Damage', ['Elixir of Fire Power']],
			['Armor', ['Elixir of the Phalanx']],
			['Health', ['Elixir of Fortitude']],
			['Mana Regeneration', ['Greater Mageblood Elixir']],
			['Physical Damage', ['Bogling Root']],
			['Other', ['Bloodkelp Elixir of Dodging']],
		]);
	});

	it('keeps the order it is given within each line', () => {
		const reversed = elixirGroups([...ELIXIRS].reverse(), consumable => consumable);
		expect(reversed.find(group => group.label === 'Strength')!.items.map(consumable => consumable.name)).toEqual(['Juju Power', 'Elixir of Ferocity']);
	});
});

import { ActionCategory } from '@generated/proto/api';
import type { UnitMetadata } from '@sim/player/player';
import { ActionId } from '@sim/proto/action_id';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { actionIdSets } from './action_id_sets';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const named = (spellId: number, name: string, rank = 0) =>
	Object.assign(Object.create(ActionId.prototype), ActionId.fromSpellId(spellId), { name, baseName: name.replace(/ \(Rank \d+\)/, ''), rank }) as ActionId;

const aura = (id: ActionId, category = ActionCategory.ActionCategoryUnknown, extra: object = {}) => ({
	id,
	data: { maxStacks: 0, hasIcd: false, hasExclusiveEffect: false, category, ...extra },
});

const spell = (id: ActionId, category = ActionCategory.ActionCategoryUnknown, extra: object = {}) => ({
	id,
	data: { isCastable: true, prepullOnly: false, encounterOnly: false, category, ...extra },
});

const metadata = (auras: Array<ReturnType<typeof aura>>, spells: Array<ReturnType<typeof spell>> = []) =>
	({ getAuras: () => auras, getSpells: () => spells, getName: () => '' }) as unknown as UnitMetadata;

const menu = (options: Array<{ value: ActionId; submenu?: Array<unknown> }>) =>
	options.map(option => [option.value.name, ...(option.submenu ?? []).map(segment => (segment instanceof ActionId ? segment.name : segment))].join(' > '));

describe('aura sets', () => {
	// Sorted by name, as the metadata arrives.
	const auras = [
		aura(named(1, 'Battle Shout')),
		aura(named(2, 'Battle Shout (External)'), ActionCategory.ActionCategoryExternalBuff),
		aura(named(3, 'Flurry'), ActionCategory.ActionCategoryTalent, { maxStacks: 3 }),
		aura(named(4, 'Mighty Rage Potion'), ActionCategory.ActionCategoryConsumable),
		aura(named(5, 'Recklessness'), ActionCategory.ActionCategoryCooldown),
		aura(named(6, 'Unbridled Wrath (Rank 1)', 1), ActionCategory.ActionCategoryTalent),
		aura(named(7, 'Unbridled Wrath (Rank 2)', 2), ActionCategory.ActionCategoryTalent),
	];

	it('lists the categories in a fixed order ahead of the class auras', async () => {
		expect(menu(await actionIdSets.auras.getActionIDs(metadata(auras)))).toEqual([
			'Recklessness > cooldowns',
			'Mighty Rage Potion > consumables',
			'Flurry > talents',
			'Unbridled Wrath (Rank 1) > talents > Unbridled Wrath (Rank 2)',
			'Unbridled Wrath (Rank 2) > talents',
			'Battle Shout (External) > external_buffs',
			'Battle Shout',
		]);
	});

	it('categorises the filtered sets the same way', async () => {
		expect(menu(await actionIdSets.stackable_auras.getActionIDs(metadata(auras)))).toEqual(['Flurry > talents']);
	});

	it("lists an internal cooldown trigger that shares its aura's id once", async () => {
		const clearcasting = aura(named(8, 'Clearcasting'), ActionCategory.ActionCategoryUnknown, { hasIcd: true });
		const trigger = aura(named(8, 'Clearcasting'), ActionCategory.ActionCategoryUnknown, { hasIcd: true });
		const omen = aura(named(9, 'Omen of Clarity'), ActionCategory.ActionCategoryUnknown, { hasIcd: true });
		const withTriggers = { ...metadata([clearcasting]), getIcdAuras: () => [trigger, omen] } as unknown as UnitMetadata;
		expect(menu(await actionIdSets.icd_auras.getActionIDs(withTriggers))).toEqual(['Clearcasting', 'Omen of Clarity']);
	});
});

describe('castable_spells', () => {
	const spells = [
		spell(named(10, 'Bloodthirst')),
		spell(named(11, 'Heroic Strike')),
		spell(named(12, 'Mighty Rage Potion'), ActionCategory.ActionCategoryConsumable, { encounterOnly: true }),
		spell(named(13, 'Recklessness'), ActionCategory.ActionCategoryCooldown),
	];

	it('keeps the class spells under Spell, then the categories, then the placeholders', async () => {
		const options = await actionIdSets.castable_spells.getActionIDs(metadata([], spells));
		expect(menu(options).slice(0, 4)).toEqual([
			'Bloodthirst > spell',
			'Heroic Strike > spell',
			'Recklessness > cooldowns',
			'Mighty Rage Potion > consumables',
		]);
		expect(options.at(-1)?.submenu).toEqual(['placeholders']);
		expect(options.find(option => option.value.name === 'Mighty Rage Potion')?.listScope).toBe('priority');
	});

	it('files the lower ranks of a spell under its highest rank, so the submenu row picks the highest', async () => {
		const ranked = [spell(named(20, 'Fireball (Rank 1)', 1)), spell(named(21, 'Fireball (Rank 3)', 3)), spell(named(22, 'Fireball (Rank 2)', 2))];
		expect(menu(await actionIdSets.castable_spells.getActionIDs(metadata([], ranked))).slice(0, 3)).toEqual([
			'Fireball (Rank 1) > spell > Fireball (Rank 3)',
			'Fireball (Rank 3) > spell',
			'Fireball (Rank 2) > spell > Fireball (Rank 3)',
		]);
	});
});

describe('menu groups', () => {
	it('lists the groups, then Spell and the categories', async () => {
		const warrior = [
			spell(named(2457, 'Battle Stance'), ActionCategory.ActionCategoryUnknown, { menuGroup: 'stances' }),
			spell(named(30, 'Bloodthirst')),
			spell(named(2458, 'Berserker Stance'), ActionCategory.ActionCategoryUnknown, { menuGroup: 'stances' }),
			spell(named(34, 'Recklessness'), ActionCategory.ActionCategoryCooldown),
		];
		expect(menu(await actionIdSets.castable_spells.getActionIDs(metadata([], warrior))).slice(0, 4)).toEqual([
			'Battle Stance > stances',
			'Berserker Stance > stances',
			'Bloodthirst > spell',
			'Recklessness > cooldowns',
		]);
	});

	it('lists the groups, then the categories and the rest of the auras', async () => {
		const auras = [
			aura(named(60, 'Battle Shout')),
			aura(named(2457, 'Battle Stance'), ActionCategory.ActionCategoryUnknown, { menuGroup: 'stances' }),
			aura(named(63, 'Flurry'), ActionCategory.ActionCategoryTalent),
			aura(named(64, 'Recklessness'), ActionCategory.ActionCategoryCooldown),
		];
		expect(menu(await actionIdSets.auras.getActionIDs(metadata(auras)))).toEqual([
			'Battle Stance > stances',
			'Recklessness > cooldowns',
			'Flurry > talents',
			'Battle Shout',
		]);
	});

	it('files a cooldown with a group under the group, not cooldowns', async () => {
		const grouped = [spell(named(72, 'Recklessness'), ActionCategory.ActionCategoryCooldown, { menuGroup: 'my_group' })];
		expect(menu(await actionIdSets.castable_spells.getActionIDs(metadata([], grouped)))[0]).toBe('Recklessness > my_group');
	});

	it('files a dotted group as a path of submenus', async () => {
		const nested = [spell(named(78, 'Ravage'), ActionCategory.ActionCategoryUnknown, { menuGroup: 'spell.cat.stealth' })];
		expect(menu(await actionIdSets.castable_spells.getActionIDs(metadata([], nested)))[0]).toBe('Ravage > spell > cat > stealth');
	});

	it('lists the groups in their fixed order, a nested group after its parent and an unlisted one last', async () => {
		// Sorted by name, as the metadata arrives.
		const druid = [
			spell(named(80, 'Bash'), ActionCategory.ActionCategoryUnknown, { menuGroup: 'spell.utility' }),
			spell(named(81, 'Battle Stance'), ActionCategory.ActionCategoryUnknown, { menuGroup: 'stances' }),
			spell(named(82, 'Bear Form'), ActionCategory.ActionCategoryUnknown, { menuGroup: 'spell.forms' }),
			spell(named(83, 'Claw'), ActionCategory.ActionCategoryUnknown, { menuGroup: 'spell.cat' }),
			spell(named(84, 'Maul'), ActionCategory.ActionCategoryUnknown, { menuGroup: 'spell.bear' }),
			spell(named(85, 'Prowl'), ActionCategory.ActionCategoryUnknown, { menuGroup: 'spell.cat.stealth' }),
			spell(named(86, 'Shred'), ActionCategory.ActionCategoryUnknown, { menuGroup: 'spell.cat' }),
		];
		expect(menu(await actionIdSets.castable_spells.getActionIDs(metadata([], druid))).slice(0, 7)).toEqual([
			'Bear Form > spell > forms',
			'Claw > spell > cat',
			'Shred > spell > cat',
			'Prowl > spell > cat > stealth',
			'Maul > spell > bear',
			'Bash > spell > utility',
			'Battle Stance > stances',
		]);
	});

	it('opens a parent group with a nested group listed before it', async () => {
		// Sorted by name, as the metadata arrives.
		const auras = [
			aura(named(90, 'Cat Form'), ActionCategory.ActionCategoryUnknown, { menuGroup: 'druid.forms' }),
			aura(named(91, 'Clearcasting'), ActionCategory.ActionCategoryUnknown, { menuGroup: 'druid' }),
			aura(named(92, 'Dire Bear Form'), ActionCategory.ActionCategoryUnknown, { menuGroup: 'druid.forms' }),
			aura(named(93, 'Moving')),
		];
		expect(menu(await actionIdSets.auras.getActionIDs(metadata(auras)))).toEqual([
			'Cat Form > druid > forms',
			'Dire Bear Form > druid > forms',
			'Clearcasting > druid',
			'Moving',
		]);
	});

	it('files every rank of a grouped spell in the group, behind its highest rank', async () => {
		const ranked = [
			spell(named(74, 'Cat Form (Rank 1)', 1), ActionCategory.ActionCategoryUnknown, { menuGroup: 'forms' }),
			spell(named(75, 'Cat Form (Rank 2)', 2), ActionCategory.ActionCategoryUnknown, { menuGroup: 'forms' }),
			spell(named(76, 'Claw')),
		];
		expect(menu(await actionIdSets.castable_spells.getActionIDs(metadata([], ranked))).slice(0, 3)).toEqual([
			'Cat Form (Rank 1) > forms > Cat Form (Rank 2)',
			'Cat Form (Rank 2) > forms',
			'Claw > spell',
		]);
	});
});

describe('talents', () => {
	const talentNames: Record<number, string> = { 40: 'Anger Management', 41: 'Cruelty' };
	const talentMetadata = (talents: Array<ActionId>) => ({ getTalents: () => talents }) as unknown as UnitMetadata;
	afterEach(() => vi.restoreAllMocks());

	it('offers each held talent at its rank, with the lower ranks behind it', async () => {
		// The lower ranks are built and filled by the set itself; the tooltip lookup is not reachable here.
		vi.spyOn(ActionId.prototype, 'fill').mockImplementation(function (this: ActionId) {
			return Promise.resolve(named(this.spellId, `${talentNames[this.spellId]} (Rank ${this.rank})`, this.rank));
		});

		const options = await actionIdSets.talents.getActionIDs(talentMetadata([named(40, 'Anger Management (Rank 1)', 1), named(41, 'Cruelty (Rank 3)', 3)]));
		expect(menu(options)).toEqual([
			'Anger Management (Rank 1)',
			'Cruelty (Rank 3)',
			'Cruelty (Rank 1) > Cruelty (Rank 3)',
			'Cruelty (Rank 2) > Cruelty (Rank 3)',
		]);
		expect(options.map(option => [option.value.spellId, option.value.rank])).toEqual([
			[40, 1],
			[41, 3],
			[41, 1],
			[41, 2],
		]);
	});

	it('offers each held talent once and without a rank for Talent Rank', async () => {
		const options = await actionIdSets.talent_ranks.getActionIDs(
			talentMetadata([named(40, 'Anger Management (Rank 1)', 1), named(41, 'Cruelty (Rank 3)', 3)]),
		);
		expect(options.map(option => [option.value.spellId, option.value.rank])).toEqual([
			[40, 0],
			[41, 0],
		]);
	});
});

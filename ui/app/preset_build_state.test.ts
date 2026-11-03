import { ItemSpec, ItemSwap } from '@generated/proto/common';
import { PresetConfigurationCategory } from '@sim/constants/preset_categories';
import { Stats } from '@sim/proto/stats';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));
vi.mock('@i18n/localization', () => ({ translatePresetConfigurationCategory: (category: string) => `category:${category}` }));

const { buildCategories, isBuildActive } = await import('./preset_build_state');

const build = (parts: Record<string, unknown>) => ({ name: 'P1', ...parts }) as any;

describe('buildCategories', () => {
	it('names each top-level category a build carries, and never the build name', () => {
		expect(buildCategories(build({ gear: {}, talents: {} }))).toEqual(['category:gear', 'category:talents']);
	});

	// `encounter`, `settings` and `reforgeSettings` are excluded from the key walk and then added back
	// by their own rules, so a build carrying only `encounter.encounter` still lists the category.
	it('reads the encounter category off the nested field rather than the key', () => {
		expect(buildCategories(build({ encounter: {} }))).toEqual([]);
		expect(buildCategories(build({ encounter: { encounter: {} } }))).toEqual(['category:encounter']);
	});

	it('folds every settings key except the named ones into one label, and lists buffs once', () => {
		expect(buildCategories(build({ settings: { specOptions: {}, consumables: {}, race: 1, buffs: {}, raidBuffs: {}, partyBuffs: {} } }))).toEqual([
			'common.preset.buffs',
			'common.preset.class_spec_options',
			'common.preset.consumables',
			'common.preset.other_settings',
		]);
	});

	it('names the debuffs settings key separately from the buffs', () => {
		expect(buildCategories(build({ settings: { debuffs: {}, buffs: {} } }))).toEqual(['common.preset.buffs', 'common.preset.debuffs']);
	});

	it('de-duplicates and sorts, so two settings keys of the same kind read once', () => {
		const categories = buildCategories(build({ settings: { race: 1, profession1: 2 } }));
		expect(categories).toEqual(['common.preset.other_settings']);
	});

	// Each of these is excluded from the key walk and re-added by its own rule, so it reads once
	// rather than under two names. `epWeights` was the one that was not, and listed itself twice.
	it('names each specially-handled category exactly once', () => {
		expect(buildCategories(build({ reforgeSettings: {} }))).toEqual(['common.preset.reforge_settings']);
		expect(buildCategories(build({ epWeights: {} }))).toEqual(['common.preset.stat_weights']);
	});

	it('lists a build swap set under gear', () => {
		expect(buildCategories(build({ gear: {}, itemSwap: {} }))).toEqual(['category:gear']);
	});
});

describe('isBuildActive', () => {
	const swapWith = (id: number) => ItemSwap.create({ items: [ItemSpec.create(), ItemSpec.create({ id, enchant: 8211 })] });
	const hostWithSwap = (itemSwap: ItemSwap) => ({ player: { itemSwapSettings: { getEnableItemSwap: () => true, toProto: () => itemSwap } } }) as any;
	const swapBuild = (itemSwap: ItemSwap) => build({ settings: { name: 'P1', playerOptions: { enableItemSwap: true, itemSwap } } });

	// The live swap always carries a prepull stats message; a preset's JSON leaves it out.
	it('matches a preset item swap without prepull stats against the live all-zero ones', () => {
		const live = { ...swapWith(15257), prepullBonusStats: { ...new Stats().toProto(), apiVersion: 2 } };

		expect(isBuildActive(swapBuild(swapWith(15257)), hostWithSwap(live), [PresetConfigurationCategory.Settings])).toBe(true);
	});

	it('tells a different swap item apart', () => {
		const live = { ...swapWith(15257), prepullBonusStats: new Stats().toProto() };

		expect(isBuildActive(swapBuild(swapWith(9449)), hostWithSwap(live), [PresetConfigurationCategory.Settings])).toBe(false);
	});
});

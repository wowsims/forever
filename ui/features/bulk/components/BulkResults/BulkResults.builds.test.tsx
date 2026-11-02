// The Results pane over a real Sim and Player, with the real rows, for one finished Builds encounter. No i18n mock: with
// no locale loaded every string is its key.
//
// The encounter ranks Raid (helm 2, +20 strength) at 1,100, with the current talents, rotation and settings or the build
// a case passes, then the current gear set with a helm enchant it does not wear at 1,050, then the current gear as
// equipped at 1,000; those two run under the settings the case passes.
import { DistributionMetrics } from '@generated/proto/api';
import { APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { ItemSlot, Race, Stat } from '@generated/proto/common';
import { SavedSettings, UIEnchant } from '@generated/proto/ui';
import type { BulkEncounterResults, BulkGearSetup, BulkPlayerSettingsSetup, BulkRankedResult } from '@sim/bulk/types';
import { Gear } from '@sim/proto/gear';
import { Stats } from '@sim/proto/stats';
import { patchBulkState, seedBulkSettings } from '@sim/settings/bulk_settings';
import { subscribeGated } from '@sim/state/batch';
import { classTalentsConfig } from '@sim/talents/factory';
import { parseTalentsString, treePointTotal } from '@sim/talents/talents_string';
import { renderWithHost } from '@testing/react';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { compareGearSet, CURRENT_ROTATION_SETUP, CURRENT_SETTINGS_SETUP, CURRENT_TALENT_SETUP } from '@testing/storybook/fixtures/bulk-compare';
import { act, cleanup, fireEvent, screen, within } from '@testing-library/react';
import { toastManager } from '@ui-kit/Toast';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { BulkResults } from './BulkResults';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());
vi.mock('@sim/hooks/useActionId', () => ({
	useActionId: () => ({ iconUrl: '', name: '', href: '', ready: true }),
}));

beforeAll(() => serveDatabase());

const STRENGTH = new Stats().withStat(Stat.StatStrength, 20);
const RAID: BulkGearSetup = { gearSet: 'Raid', gear: compareGearSet(2), gearSetGear: compareGearSet(2), bonusStats: STRENGTH };
const HELM_ENCHANT = UIEnchant.create({ effectId: 2545, name: 'Fixture Helm Enchant' });
const CURRENT_ENCHANTED: BulkGearSetup = {
	gearSet: '',
	gear: new Gear({ [ItemSlot.ItemSlotHead]: compareGearSet(1).getEquippedItem(ItemSlot.ItemSlotHead)!.withEnchant(HELM_ENCHANT) }),
	gearSetGear: compareGearSet(1),
	bonusStats: new Stats(),
};
// Within the warrior's Arms tree: the player keeps a string its trees accept as it is (ui/sim/talents/factory.ts).
const ARMS = { name: 'Arms', talentsString: '3030500110', isCurrent: false };
const OPENER = {
	name: 'Opener',
	isCurrent: false,
	rotation: APLRotation.create({ type: APLRotationType.TypeSimple, simple: { specRotationJson: '{"useOpener":true}' } }),
};

const HEALER: BulkPlayerSettingsSetup = { name: 'Healer', isCurrent: false, settings: SavedSettings.create({ race: Race.RaceDwarf }) };

type Build = Pick<BulkRankedResult, 'talents' | 'rotation' | 'settings'>;

const ARMS_BUILD: Build = { talents: ARMS, rotation: OPENER, settings: HEALER };

const ranked = (avg: number, fields: Omit<BulkRankedResult, 'dpsMetrics'>): BulkRankedResult => ({
	...fields,
	dpsMetrics: DistributionMetrics.create({ avg, stdev: 10 }),
});

const entry = (settings: BulkPlayerSettingsSetup, build: Partial<Build> = {}): BulkEncounterResults => {
	const current = { talents: CURRENT_TALENT_SETUP, rotation: CURRENT_ROTATION_SETUP, settings };
	const base = ranked(1000, { gear: compareGearSet(1), ...current, gearSetup: null });
	const raid = ranked(1100, { gear: RAID.gear, ...current, ...build, gearSetup: RAID });
	const currentGear = ranked(1050, { gear: CURRENT_ENCHANTED.gear, ...current, gearSetup: CURRENT_ENCHANTED });
	return {
		name: '',
		isCurrent: true,
		duration: 180,
		durationVariation: 5,
		results: {
			chains: [[raid], [currentGear], [base]],
			originalGearResults: base,
			referenceIsCurrent: true,
			currentGear: compareGearSet(1),
			iterations: 100,
		},
	};
};

const mount = async (shown: BulkEncounterResults) => {
	const { sim, player } = await makeIndividualSim();
	seedBulkSettings(player);
	patchBulkState(player, { gearStarted: true, gearResults: [shown] });
	const onActivateTab = vi.fn();
	renderWithHost(<BulkResults onActivateTab={onActivateTab} selected={0} />, { sim, player });
	return { sim, player, onActivateTab };
};

const rows = () => screen.getAllByTestId('bulk-sim-result-root');
const buildLine = (row: HTMLElement) => within(row).queryByTestId('bulk-result-talents');
// The build line in order: the gear pill, then its enchant count and each pill by their test ids.
const lineParts = (row: HTMLElement) =>
	Array.from(buildLine(row)?.childNodes ?? []).flatMap(part => {
		if (part.nodeType === Node.TEXT_NODE) return [['text', part.textContent]];
		const tagged = part instanceof Element ? (part.matches('[data-testid]') ? part : part.querySelector('[data-testid]')) : null;
		return tagged ? [[tagged.getAttribute('data-testid'), tagged.textContent]] : [];
	});

const CURRENT_PILLS = [
	['bulk-result-talents-anchor', 'bulk_tab.results.talentsbulk_tab.results.current_talents'],
	['bulk-result-rotation', 'bulk_tab.results.rotationbulk_tab.results.current_rotation'],
	['bulk-result-settings', 'bulk_tab.results.settingsbulk_tab.results.current_settings'],
];

describe('BulkResults in Builds', () => {
	it('names each row’s gear setup pill by its gear set, the current gear set’s and the reference row’s reading "Current gear"', async () => {
		await mount(entry(CURRENT_SETTINGS_SETUP));

		const [build, currentGear, reference] = rows();
		const gearPill = (row: HTMLElement) => within(row).getByTestId('bulk-result-gear-set').textContent;
		expect([build, currentGear, reference].map(gearPill)).toEqual([
			'bulk_tab.results.gearRaid',
			'bulk_tab.results.gearbulk_tab.results.current_gear',
			'bulk_tab.results.gearbulk_tab.results.current_gear',
		]);
		expect(within(reference).getByTestId('results-sim-dps').firstElementChild!.textContent).toBe('bulk_tab.results.current_gear');
	});

	it('counts the enchants a setup changes on its gear set after its label and lists them by slot on hover', async () => {
		await mount(entry(CURRENT_SETTINGS_SETUP));

		const [build, currentGear] = rows();
		expect(within(build).queryByTestId('bulk-enchants-differ')).toBeNull();
		const count = within(currentGear).getByTestId('bulk-enchants-differ');
		expect(count.tabIndex).toBe(0);

		fireEvent.mouseEnter(count);

		const tooltip = await screen.findByRole('tooltip');
		expect(
			within(tooltip)
				.getAllByRole('listitem')
				.map(item => item.textContent),
		).toEqual(['slots.head: Fixture Helm Enchant']);
	});

	it('shows the gear, talents, rotation and settings pills on every Builds row, in that order, even when every row shares them', async () => {
		await mount(entry(CURRENT_SETTINGS_SETUP));

		const [build, currentGear, reference] = rows();
		expect(lineParts(build)).toEqual([['bulk-result-gear-set', 'bulk_tab.results.gearRaid'], ...CURRENT_PILLS]);
		expect(lineParts(currentGear)).toEqual([
			['bulk-result-gear-set', 'bulk_tab.results.gearbulk_tab.results.current_gear'],
			['bulk-enchants-differ', 'bulk_tab.compare.enchants_differ'],
			...CURRENT_PILLS,
		]);
		expect(lineParts(reference)).toEqual([['bulk-result-gear-set', 'bulk_tab.results.gearbulk_tab.results.current_gear'], ...CURRENT_PILLS]);

		cleanup();
		await mount(entry(CURRENT_SETTINGS_SETUP, ARMS_BUILD));

		const [arms] = rows();
		expect(lineParts(arms)).toEqual([
			['bulk-result-gear-set', 'bulk_tab.results.gearRaid'],
			['bulk-result-talents-anchor', 'bulk_tab.results.talentsArms'],
			['bulk-result-rotation', 'bulk_tab.results.rotationOpener'],
			['bulk-result-settings', 'bulk_tab.results.settingsHealer'],
		]);
		for (const testId of ['bulk-result-rotation', 'bulk-result-settings']) {
			const pill = within(arms).getByTestId(testId);
			expect(pill.hasAttribute('tabindex')).toBe(false);
			expect(pill.hasAttribute('data-tooltip-id')).toBe(false);
		}
	});

	it('shows no talents, rotation or settings pill on a Gear-mode row', async () => {
		const gearMode = { talents: CURRENT_TALENT_SETUP, rotation: CURRENT_ROTATION_SETUP, gearSetup: null, settings: null };
		const base = ranked(1000, { gear: compareGearSet(1), ...gearMode });
		await mount({
			name: '',
			isCurrent: true,
			duration: 180,
			durationVariation: 5,
			results: {
				chains: [[ranked(1100, { gear: RAID.gear, ...gearMode })], [base]],
				originalGearResults: base,
				referenceIsCurrent: true,
				currentGear: compareGearSet(1),
				iterations: 100,
			},
		});

		expect(rows().map(buildLine)).toEqual([null, null]);
	});

	it('lists the talents that hold points under their trees in the talent pill’s tooltip, each with its rank', async () => {
		const { player } = await mount(entry(CURRENT_SETTINGS_SETUP, ARMS_BUILD));
		const config = classTalentsConfig[player.getClass()];
		const points = parseTalentsString(config, ARMS.talentsString);
		const spent = points.flatMap((tree, treeIdx) =>
			tree.flatMap((value, talentIdx) => (value > 0 ? [`${value}/${config[treeIdx].talents[talentIdx].maxPoints}`] : [])),
		);
		const treesWithPoints = config.filter((_, treeIdx) => treePointTotal(points, treeIdx) > 0);

		fireEvent.mouseEnter(within(rows()[0]).getByTestId('bulk-result-talents-anchor'));

		const grid = await screen.findByTestId('bulk-talents-grid');
		const trees = within(grid).getAllByTestId('bulk-talents-grid-tree');
		expect(trees.map(tree => tree.firstElementChild!.textContent)).toEqual(
			treesWithPoints.map(tree => `${tree.name}${treePointTotal(points, config.indexOf(tree))}`),
		);
		const talents = within(grid).getAllByTestId('bulk-talents-grid-talent');
		expect(talents.map(talent => talent.querySelector('.ui-talent-picker-points')!.textContent)).toEqual(spent);
	});

	it('says the gear is unchanged on a row that wears the current gear with another bundle’s talents, and offers Equip', async () => {
		const base = ranked(1000, {
			gear: compareGearSet(1),
			talents: CURRENT_TALENT_SETUP,
			rotation: CURRENT_ROTATION_SETUP,
			gearSetup: null,
			settings: CURRENT_SETTINGS_SETUP,
		});
		const wearsCurrentGear: BulkGearSetup = { gearSet: '', gear: compareGearSet(1), gearSetGear: compareGearSet(1), bonusStats: new Stats() };
		const arms = ranked(1100, {
			gear: compareGearSet(1),
			talents: ARMS,
			rotation: CURRENT_ROTATION_SETUP,
			gearSetup: wearsCurrentGear,
			settings: CURRENT_SETTINGS_SETUP,
		});
		await mount({
			...entry(CURRENT_SETTINGS_SETUP),
			results: { chains: [[arms], [base]], originalGearResults: base, referenceIsCurrent: true, currentGear: compareGearSet(1), iterations: 100 },
		});

		const [armsRow, reference] = rows();
		expect(within(armsRow).getByTestId('bulk-result-gear-unchanged').textContent).toBe('bulk_tab.results.gear_unchanged');
		expect(within(armsRow).getByTestId('bulk-equip-btn').hasAttribute('data-hidden')).toBe(false);
		expect(within(reference).queryByTestId('bulk-result-gear-unchanged')).toBeNull();
	});

	it('stands the best row as the reference when the current setup did not run: no delta, with its items and Equip, the others a loss', async () => {
		const current = { talents: CURRENT_TALENT_SETUP, rotation: CURRENT_ROTATION_SETUP, settings: CURRENT_SETTINGS_SETUP };
		const raid = ranked(156.45, { gear: RAID.gear, ...current, gearSetup: RAID });
		const enchanted = ranked(71.17, { gear: CURRENT_ENCHANTED.gear, ...current, gearSetup: CURRENT_ENCHANTED });
		await mount({
			...entry(CURRENT_SETTINGS_SETUP),
			results: { chains: [[raid], [enchanted]], originalGearResults: raid, referenceIsCurrent: false, currentGear: compareGearSet(1), iterations: 100 },
		});

		const [best, other] = rows();
		expect(within(best).queryByTestId('results-reference-diff')).toBeNull();
		expect(within(best).getByTestId('results-sim-dps').textContent).not.toContain('bulk_tab.results.current_gear');
		expect(within(best).getByTestId('bulk-equip-btn').hasAttribute('data-hidden')).toBe(false);
		expect(within(best).getAllByTestId('bulk-result-item').length).toBeGreaterThan(0);
		expect(within(other).getByTestId('results-reference-diff').getAttribute('data-sign')).toBe('negative');
		// In percent of the best set's DPS, as Compare gives it, not of the row's own (119.83%).
		expect(within(other).getByTestId('results-reference-diff').textContent).toBe('-85.28 (54.51%)');
		expect(within(other).getByTestId('bulk-equip-btn').hasAttribute('data-hidden')).toBe(false);
	});

	it('equips the row’s gear setup, talents, rotation and settings with one build toast', async () => {
		const { sim, player, onActivateTab } = await mount(entry(CURRENT_SETTINGS_SETUP, ARMS_BUILD));
		player.setRace(Race.RaceOrc);
		const toast = vi.spyOn(toastManager, 'add').mockReturnValue('toast');
		let notified = 0;
		const unsubscribe = subscribeGated(
			sim.store.subscribe,
			state => state.players[player.storeKey],
			() => notified++,
		);

		act(() => {
			fireEvent.click(within(rows()[0]).getByTestId('bulk-equip-btn'));
		});
		unsubscribe();

		expect(player.getGear().equals(RAID.gear)).toBe(true);
		expect(player.getTalentsString()).toBe(ARMS.talentsString);
		expect(APLRotation.equals(player.aplRotation, OPENER.rotation)).toBe(true);
		expect(player.getBonusStats().equals(STRENGTH)).toBe(true);
		expect(player.getRace()).toBe(Race.RaceDwarf);
		expect(notified).toBe(1);
		expect(toast).toHaveBeenCalledOnce();
		expect(toast).toHaveBeenCalledWith({ variant: 'success', body: 'bulk_tab.results.build_equipped' });
		expect(onActivateTab).toHaveBeenCalledWith('gear-tab');
	});

	it('leaves the settings alone when the row ran under the current settings', async () => {
		const { player } = await mount(entry(HEALER, { talents: ARMS, rotation: OPENER, settings: CURRENT_SETTINGS_SETUP }));
		player.setRace(Race.RaceOrc);
		const toast = vi.spyOn(toastManager, 'add').mockReturnValue('toast');

		act(() => {
			fireEvent.click(within(rows()[0]).getByTestId('bulk-equip-btn'));
		});

		expect(player.getGear().equals(RAID.gear)).toBe(true);
		expect(player.getRace()).toBe(Race.RaceOrc);
		expect(toast).toHaveBeenCalledWith({ variant: 'success', body: 'bulk_tab.results.build_equipped' });
	});
});

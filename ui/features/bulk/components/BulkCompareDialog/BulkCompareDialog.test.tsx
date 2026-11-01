// The dialog over a real store holding three finished encounters. With no locale loaded every string is its key, so sets
// are told apart by their gains.
//
// The fixture, in display order: Council (1:30) ranks S1 1,030, S2 1,020, S3 990; Patchwerk (5:00) S2 1,100, S4 1,090,
// S5 1,080; Ragnaros (10:00) S6 12,500, S1 1,150, S2 1,120. The current gear does 1,000 in each, and every result
// spreads ±10 a fight.
//
// The Builds fixture runs two encounters, the current one (3:00) and Patchwerk (5:00). The first ranks Raid (helm 2) with
// Arms, Opener and Healer, then a setup of the current gear set (helm 3) with the current talents, rotation and settings;
// Patchwerk ranks Raid with the current ones.
import { DistributionMetrics } from '@generated/proto/api';
import { APLRotation } from '@generated/proto/apl';
import { ItemSlot, ItemType } from '@generated/proto/common';
import { SavedSettings, UIEnchant, UIItem } from '@generated/proto/ui';
import type { BulkEncounterResults, BulkGearSetup, BulkRankedResult } from '@sim/bulk/types';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { EquippedItem } from '@sim/proto/equipped_item';
import { Gear } from '@sim/proto/gear';
import { Stats } from '@sim/proto/stats';
import {
	compareGearSet,
	CURRENT_ROTATION_SETUP,
	CURRENT_SETTINGS_SETUP,
	CURRENT_TALENT_SETUP,
	makeBulkCompareHost,
} from '@testing/storybook/fixtures/bulk-compare';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { PortalContainerContext } from '@ui-kit/hooks/usePortalContainer';
import { describe, expect, it } from 'vitest';

import { BulkCompareDialog } from './BulkCompareDialog';

const rootElem = document.createElement('div');

const RAID: BulkGearSetup = { gearSet: 'Raid', gear: compareGearSet(2), bonusStats: new Stats() };
const CURRENT_UNDEAD: BulkGearSetup = { gearSet: '', gear: compareGearSet(3), bonusStats: new Stats() };

// A gear set of a helm and a cloak, and two setups of it: one enchants both, one neither.
const enchantedSet = (helm?: UIEnchant, cloak?: UIEnchant) =>
	new Gear({
		[ItemSlot.ItemSlotHead]: new EquippedItem({ item: UIItem.create({ id: 30, name: 'Helm 30', type: ItemType.ItemTypeHead }), enchant: helm }),
		[ItemSlot.ItemSlotBack]: new EquippedItem({ item: UIItem.create({ id: 31, name: 'Cloak 31', type: ItemType.ItemTypeBack }), enchant: cloak }),
	});
const enchantVariant = (gear: Gear): BulkGearSetup => ({ gearSet: '', gear, gearSetGear: enchantedSet(), bonusStats: new Stats() });
const BOTH_ENCHANTED = enchantVariant(
	enchantedSet(UIEnchant.create({ effectId: 1, name: 'Fixture Helm Enchant' }), UIEnchant.create({ effectId: 2, name: 'Fixture Cloak Enchant' })),
);
const NONE_ENCHANTED = enchantVariant(enchantedSet());
const ARMS_BUILD = {
	talents: { name: 'Arms', talentsString: '3030500130', isCurrent: false },
	rotation: { name: 'Opener', isCurrent: false, rotation: APLRotation.create() },
	settings: { name: 'Healer', isCurrent: false, settings: SavedSettings.create() },
};
const CURRENT_BUILD = { talents: CURRENT_TALENT_SETUP, rotation: CURRENT_ROTATION_SETUP, settings: CURRENT_SETTINGS_SETUP };

const buildRow = (gearSetup: BulkGearSetup | null, avg: number, build = CURRENT_BUILD): BulkRankedResult => ({
	gear: gearSetup?.gear ?? compareGearSet(1),
	dpsMetrics: DistributionMetrics.create({ avg, stdev: 10 }),
	...build,
	gearSetup,
});

const buildEntry = (encounter: string, duration: number, ranked: BulkRankedResult[]): BulkEncounterResults => {
	const base = buildRow(null, 1000);
	return {
		name: encounter,
		isCurrent: encounter === '',
		duration,
		durationVariation: 5,
		results: {
			chains: [...ranked, base].map(result => [result]),
			originalGearResults: base,
			referenceIsCurrent: true,
			currentGear: compareGearSet(1),
			iterations: 100,
		},
	};
};

// As a batch that did not sim the current setup stores an encounter: `ranked` best first, the best the reference.
const bestReferenceEntry = (encounter: string, duration: number, ranked: BulkRankedResult[]): BulkEncounterResults => ({
	name: encounter,
	isCurrent: encounter === '',
	duration,
	durationVariation: 5,
	results: {
		chains: ranked.map(result => [result]),
		originalGearResults: ranked[0],
		referenceIsCurrent: false,
		currentGear: compareGearSet(1),
		iterations: 100,
	},
});

const BUILD_ENTRIES: BulkEncounterResults[] = [
	buildEntry('', 180, [buildRow(RAID, 1050, ARMS_BUILD), buildRow(CURRENT_UNDEAD, 1020)]),
	buildEntry('Patchwerk', 300, [buildRow(RAID, 1100)]),
];

const mount = (entries?: BulkEncounterResults[]) => {
	document.body.appendChild(rootElem);
	render(
		<PortalContainerContext value={rootElem}>
			<SimHostProvider host={makeBulkCompareHost(entries)}>
				<BulkCompareDialog open onOpenChange={() => {}} />
			</SimHostProvider>
		</PortalContainerContext>,
	);
};

const dialog = () => within(screen.getByRole('dialog'));
const bodyRows = () => dialog().getAllByRole('row').slice(1);
const cells = (row: HTMLElement) => within(row).getAllByRole('cell');
const shown = (row: HTMLElement) =>
	cells(row).map(cell => (cell.hasAttribute('data-gap') ? 'gap' : `${cell.firstElementChild!.textContent}${cell.hasAttribute('data-best') ? ' best' : ''}`));

describe('BulkCompareDialog', () => {
	it('shows the table straight away, with no view switch', () => {
		mount();

		expect(dialog().getByRole('table')).toBeTruthy();
		expect(screen.getByRole('dialog').querySelector('[aria-pressed]')).toBeNull();
	});

	it('gives every encounter a column by length', () => {
		mount();

		const headers = within(dialog().getAllByRole('row')[0]).getAllByRole('columnheader');
		expect(headers.map(header => header.firstChild!.textContent)).toEqual(['bulk_tab.compare.table_set', 'Council', 'Patchwerk', 'Ragnaros']);
	});

	it('lists the current gear, then every set by number, with gaps and each encounter’s best marked', () => {
		mount();

		const rows = bodyRows();
		expect(rows.map(row => within(row).getByRole('rowheader').firstChild!.textContent)).toEqual([
			'bulk_tab.results.current',
			...Array.from({ length: 6 }, () => 'bulk_tab.compare.short_set'),
		]);
		const [current, s1, , , , , s6] = rows;
		expect(shown(current)).toEqual(['0.00%', '0.00%', '0.00%']);
		expect(shown(s1)).toEqual(['+3.00% best', 'gap', '+15.00%']);
		expect(shown(s6)).toEqual(['gap', 'gap', '+1,150.00% best']);
	});

	it('names the items a set changes under its label', () => {
		mount();

		const [, s1] = bodyRows();
		expect(within(s1).getByText('Helm 3')).toBeTruthy();
	});

	it('gives each cell its rank, spread and gain as a tooltip, and the current gear’s its DPS and spread only', () => {
		mount();

		const [current, s1, , s3] = bodyRows();
		expect(cells(s1)[0].getAttribute('data-tooltip-content')).toBe('bulk_tab.compare.readout_rank bulk_tab.compare.readout_spread · +30.00 (+3.00%)');
		expect(cells(s3)[0].getAttribute('data-tooltip-content')).toBe('bulk_tab.compare.readout_rank bulk_tab.compare.readout_spread · -10.00 (-1.00%)');
		expect(cells(current)[0].getAttribute('data-tooltip-content')).toBe('bulk_tab.compare.readout_current bulk_tab.compare.readout_spread');
	});

	it('heads each column with its encounter and names each Builds set by its gear setup, talents, rotation and settings', () => {
		mount(BUILD_ENTRIES);

		const headers = within(dialog().getAllByRole('row')[0]).getAllByRole('columnheader');
		expect(headers.map(header => header.firstChild!.textContent)).toEqual(['bulk_tab.compare.table_build', 'bulk_tab.results.current', 'Patchwerk']);
		const current =
			'bulk_tab.results.talentsbulk_tab.results.current_talentsbulk_tab.results.rotationbulk_tab.results.current_rotationbulk_tab.results.settingsbulk_tab.results.current_settings';
		expect(bodyRows().map(row => within(row).getByRole('rowheader').textContent)).toEqual([
			`bulk_tab.results.currentbulk_tab.results.gearbulk_tab.results.current_gear${current}`,
			'bulk_tab.compare.short_setbulk_tab.results.gearRaidbulk_tab.results.talentsArmsbulk_tab.results.rotationOpenerbulk_tab.results.settingsHealer',
			`bulk_tab.compare.short_setbulk_tab.results.gearbulk_tab.results.current_gear${current}`,
			`bulk_tab.compare.short_setbulk_tab.results.gearRaid${current}`,
		]);
		expect(bodyRows().map(shown)).toEqual([
			['0.00%', '0.00%'],
			['+5.00% best', 'gap'],
			['+2.00%', 'gap'],
			['gap', '+10.00% best'],
		]);
	});

	it('counts a Builds set’s enchants that differ from its gear set’s own and lists them by slot on hover', async () => {
		mount([buildEntry('', 180, [buildRow(BOTH_ENCHANTED, 1030), buildRow(NONE_ENCHANTED, 1010), buildRow(RAID, 1005)])]);

		const [current, both, none, raid] = bodyRows();
		expect([current, none, raid].map(row => within(row).queryByTestId('bulk-enchants-differ'))).toEqual([null, null, null]);
		const count = within(both).getByTestId('bulk-enchants-differ');
		expect(count.textContent).toBe('bulk_tab.compare.enchants_differ');
		expect(count.tabIndex).toBe(0);

		fireEvent.mouseEnter(count);

		const tooltip = await screen.findByRole('tooltip');
		expect(
			within(tooltip)
				.getAllByRole('listitem')
				.map(item => item.textContent),
		).toEqual(['slots.head: Fixture Helm Enchant', 'slots.back: Fixture Cloak Enchant']);
	});

	it('shows no current row when the batch did not sim the current setup, each column’s best set at 0.00% and the rest a loss', () => {
		mount([
			bestReferenceEntry('', 180, [buildRow(RAID, 1000, ARMS_BUILD), buildRow(CURRENT_UNDEAD, 800)]),
			bestReferenceEntry('Patchwerk', 300, [buildRow(CURRENT_UNDEAD, 1250), buildRow(RAID, 1000, ARMS_BUILD)]),
		]);

		expect(bodyRows().map(row => within(row).getByRole('rowheader').firstChild!.textContent)).toEqual([
			'bulk_tab.compare.short_set',
			'bulk_tab.compare.short_set',
		]);
		expect(bodyRows().map(shown)).toEqual([
			['0.00% best', '-20.00%'],
			['-20.00%', '0.00% best'],
		]);
		expect(dialog().getByText('bulk_tab.compare.table_caption_best')).toBeTruthy();
		const legend = dialog().getByText(/table_legend_best/).textContent;
		expect(legend).toContain('bulk_tab.compare.table_legend_loss');
		expect(legend).toContain('bulk_tab.compare.table_legend_none');
		expect(legend).not.toContain('bulk_tab.compare.table_legend_gain');
	});

	it('names a Gear-mode set by the items it changes, with no talents, rotation or settings', () => {
		mount();

		const [current, s1] = bodyRows();
		expect(within(current).getByRole('rowheader').textContent).toBe('bulk_tab.results.currentbulk_tab.results.gearbulk_tab.results.current_gear');
		expect(within(s1).getByRole('rowheader').textContent).toBe('bulk_tab.compare.short_setHelm 3');
	});
});

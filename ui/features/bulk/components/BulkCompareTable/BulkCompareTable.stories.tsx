import type { Meta, StoryObj } from '@storybook/react-vite';

import { type GainScale, type GainTint, gainTint } from '../../model/results/compare';
import {
	BulkCompareTable,
	type BulkCompareTableCell,
	type BulkCompareTableColumn,
	type BulkCompareTableProps,
	type BulkCompareTableRow,
} from './BulkCompareTable';

interface StoryCell {
	gainPct: number;
	best: boolean;
}

interface StoryRow extends Omit<BulkCompareTableRow, 'cells'> {
	cells: Array<StoryCell | null>;
}

// A cell `gainPct` over the current gear's DPS; null is a gap.
const cell = (gainPct: number | null, best = false): StoryCell | null => (gainPct === null ? null : { gainPct, best });

const row = ({
	key,
	label,
	cells,
	changes = [],
	talents = null,
	rotation = null,
	settings = null,
	gear = null,
	enchants = [],
}: Pick<StoryRow, 'key' | 'label' | 'cells'> & Partial<StoryRow>): StoryRow => ({ key, label, talents, rotation, settings, gear, changes, enchants, cells });

const gainLabel = (gainPct: number) => `${gainPct > 0 ? '+' : ''}${gainPct.toFixed(2)}%`;

// The table's props, each column's cells tinted on that column's scale from its lowest gain to its highest, 0 included,
// as the dialog builds them; `currentDps` is the current gear's DPS in each column.
const tableArgs = (columns: BulkCompareTableColumn[], currentDps: number[], rows: StoryRow[]): BulkCompareTableProps => {
	const scales = columns.map((_, index): GainScale => {
		const gains = rows.flatMap(line => line.cells[index]?.gainPct ?? []);
		return { low: Math.min(0, ...gains), high: Math.max(0, ...gains) };
	});
	const tableCell = ({ gainPct, best }: StoryCell, index: number): BulkCompareTableCell => {
		const dps = (currentDps[index] * (1 + gainPct / 100)).toFixed(2);
		const gain = gainLabel(gainPct);
		return { gain, dps, tint: gainTint(gainPct, scales[index]), best, tooltip: `#1 · ${dps} DPS ±12 per fight · ${gain}` };
	};
	const tableRows = rows.map(line => ({ ...line, cells: line.cells.map((entry, index) => entry && tableCell(entry, index)) }));
	const tints = tableRows.flatMap(line => line.cells.flatMap(entry => entry?.tint ?? []));
	return {
		columns,
		rows: tableRows,
		legend: { lowTint: Math.min(0, ...tints) as GainTint, highTint: Math.max(0, ...tints) as GainTint },
	};
};

// Seven timed encounters by length, two of them 1:30 long, then a health-based one, which the table lists too.
const COLUMNS = [
	{ label: 'Short fight', length: '0:45' },
	{ label: 'Council', length: '1:30' },
	{ label: 'Council (adds)', length: '1:30' },
	{ label: 'Default', length: '2:00' },
	{ label: 'Current', length: '3:00' },
	{ label: 'Patchwerk', length: '5:00' },
	{ label: 'Long fight', length: '8:00' },
	{ label: 'Ragnaros', length: 'Health' },
];

const CURRENT_DPS = [1203, 980, 1010, 870, 776, 769, 760, 1100];

// S1 wins most encounters; S2 is missing from two; S3 ranks only twice; S4 loses everywhere it ranks. The current gear
// is best in Default, where every set ranked does worse.
const ROWS = [
	row({ key: 'current', label: 'Current', gear: 'Current gear', cells: COLUMNS.map((_, index) => cell(0, index === 3)) }),
	row({
		key: 'set-1',
		label: 'S1',
		changes: ['Stormrage’s Talisman of Seething', 'Ring of the Godslayer', 'Badge of the Swarmguard'],
		cells: [cell(8.58, true), cell(1.55, true), cell(1.2), cell(-0.05), cell(1.55, true), cell(1.75, true), cell(2.4, true), cell(0.4)],
	}),
	row({
		key: 'set-2',
		label: 'S2',
		changes: ['Badge of the Swarmguard', 'Cloak of Concentrated Hatred (Enchant Cloak - Greater Agility)'],
		cells: [cell(4.1), null, cell(1.9, true), cell(-0.3), null, cell(0.7), cell(0.5), cell(0.9, true)],
	}),
	row({ key: 'set-3', label: 'S3', changes: ['Kiss of the Spider'], cells: [null, cell(0.52), null, null, cell(0.52), null, null, null] }),
	row({ key: 'set-4', label: 'S4', changes: ['Hand of Justice'], cells: [cell(-0.4), null, cell(-1.4), cell(-3.2), null, cell(-0.7), null, null] }),
];

// Four timed encounters where every set gains, S1 the most.
const GAIN_COLUMNS = [
	{ label: 'Default - 60', length: '1:00' },
	{ label: 'Default - 90', length: '1:30' },
	{ label: 'Default - 120', length: '2:00' },
	{ label: 'Current', length: '3:00' },
];

const GAIN_CURRENT_DPS = [653.53, 622.04, 610.76, 600.3];

const GAIN_ROWS = [
	row({ key: 'current', label: 'Current', gear: 'Current gear', cells: GAIN_COLUMNS.map(() => cell(0)) }),
	row({
		key: 'set-1',
		label: 'S1',
		changes: ['Slayer’s Crest', 'Kiss of the Spider'],
		cells: [cell(11.18, true), cell(8.79, true), cell(7.54, true), cell(8.9, true)],
	}),
	row({ key: 'set-2', label: 'S2', changes: ['Slayer’s Crest'], cells: [cell(6.07), cell(4.85), cell(4.29), cell(4.63)] }),
	row({ key: 'set-3', label: 'S3', changes: ['Kiss of the Spider'], cells: [cell(5.89), cell(4.76), cell(4.1), cell(4.81)] }),
	row({ key: 'set-4', label: 'S4', changes: ['Slayer’s Crest', 'Frozen Heart of the Mountain'], cells: [cell(5.42), cell(4.25), cell(3.7), cell(4.07)] }),
	row({
		key: 'set-5',
		label: 'S5',
		changes: ['Kiss of the Spider', 'Frozen Heart of the Mountain'],
		cells: [cell(5.27), cell(4.26), cell(3.59), cell(4.23)],
	}),
];

// A Builds batch over three encounters: each set names its gear setup in place of the items it changes, then its talents,
// rotation and settings. S1 is P1 Fury with an enchant preset's cloak enchant, and S3 and S4 are two setups of the
// current gear set, so each counts the enchants that differ from its gear set's own and lists them on hover.
const BUILD_COLUMNS = [
	{ label: 'Short fight', length: '1:30' },
	{ label: 'Current', length: '3:00' },
	{ label: 'Patchwerk', length: '5:00' },
];

const BUILD_CURRENT_DPS = [622.04, 600.3, 588.12];

const CURRENT_BUILD = { gear: 'Current gear', talents: 'Current talents', rotation: 'Current rotation', settings: 'Current settings' };
const FURY_BUILD = { gear: 'P1 Fury', talents: 'Fury', rotation: 'Fury', settings: 'Pre-BiS - Fury' };

const BUILD_ROWS = [
	row({ key: 'current', label: 'Current', cells: BUILD_COLUMNS.map(() => cell(0)), ...CURRENT_BUILD }),
	row({ key: 'set-1', label: 'S1', cells: [cell(6.4, true), cell(5.1, true), cell(4.8, true)], ...FURY_BUILD, enchants: ['Back: Agility'] }),
	row({ key: 'set-2', label: 'S2', cells: [cell(5.2), cell(4.4), null], ...FURY_BUILD, rotation: 'My APL' }),
	row({ key: 'set-3', label: 'S3', cells: [cell(1.9), cell(2.3), cell(2.6)], ...CURRENT_BUILD, enchants: ['Back: Agility', 'Main hand: Crusader'] }),
	row({ key: 'set-4', label: 'S4', cells: [cell(1.2), cell(1.6), cell(1.4)], ...CURRENT_BUILD, enchants: ['Back: Agility'], settings: 'Tank healing' }),
];

// The same three encounters with the Current bundle unticked: no current row, and each cell is a loss against its
// encounter's best set, whose DPS each column is measured from.
const BEST_DPS = [662.85, 631.0, 616.35];

const BEST_ROWS = [
	row({ key: 'set-1', label: 'S1', cells: [cell(0, true), cell(0, true), cell(-1.2)], ...FURY_BUILD, enchants: ['Back: Agility'] }),
	row({ key: 'set-2', label: 'S2', cells: [cell(-1.1), cell(-0.7), null], ...FURY_BUILD, rotation: 'My APL' }),
	row({
		key: 'set-3',
		label: 'S3',
		cells: [cell(-4.3), cell(-2.7), cell(0, true)],
		gear: 'Pre-BiS (Arms)',
		talents: 'Arms',
		rotation: 'Arms APL',
		settings: 'Pre-BiS - Arms',
	}),
];

const meta = {
	title: 'Features/Bulk/BulkCompareTable',
	component: BulkCompareTable,
	tags: ['autodocs'],
	args: tableArgs(COLUMNS, CURRENT_DPS, ROWS),
	argTypes: { columns: { control: false }, rows: { control: false }, legend: { control: false } },
} satisfies Meta<typeof BulkCompareTable>;

export default meta;
type Story = StoryObj<typeof meta>;

// Gains and losses, so the legend's strip runs from the deepest loss's tint to the largest gain's.
export const EightEncounters: Story = {};

// Every set gains, so every cell but the current gear's is green, each encounter's largest gain darkest, and the legend
// starts at no change.
export const AllGains: Story = {
	args: tableArgs(GAIN_COLUMNS, GAIN_CURRENT_DPS, GAIN_ROWS),
};

export const Builds: Story = {
	args: { ...tableArgs(BUILD_COLUMNS, BUILD_CURRENT_DPS, BUILD_ROWS), builds: true },
};

// Every cell is 0.00% or a loss, so the legend runs from the deepest loss to no change.
export const CurrentBundleUnticked: Story = {
	args: { ...tableArgs(BUILD_COLUMNS, BEST_DPS, BEST_ROWS), builds: true, againstBest: true },
};

// Narrower than the columns need, so the table scrolls sideways under its fixed set column.
export const NarrowScrolls: Story = {
	decorators: [
		Story => (
			<div className="max-w-160">
				<Story />
			</div>
		),
	],
};

// A phone's width: the set column takes a third of it and its long item names wrap and hyphenate, two encounters fill
// the rest and the others scroll sideways.
export const Phone: Story = {
	globals: { viewport: { value: 'mobile1', isRotated: false } },
	args: tableArgs(COLUMNS, CURRENT_DPS, [
		...ROWS,
		row({
			key: 'set-5',
			label: 'S5',
			changes: ['Thunderfury, Blessed Blade of the Windseeker', 'Spaulders of Valor (Enchant Shoulders - Zandalar Signet of Might)'],
			cells: [cell(6.2), cell(1.1), cell(0.8), cell(-0.6), cell(1.3), cell(1.5), cell(2.1), cell(0.7)],
		}),
	]),
};

export const TwoEncounters: Story = {
	args: tableArgs(
		COLUMNS.slice(0, 2),
		CURRENT_DPS,
		ROWS.map(line => ({ ...line, cells: line.cells.slice(0, 2) })),
	),
};

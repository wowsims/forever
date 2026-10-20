import type { Meta, StoryObj } from '@storybook/react-vite';

import { type GainScale, gainTint } from '../../model/compare';
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

const row = (key: string, label: string, changes: string[], cells: Array<StoryCell | null>, talents: string | null = null): StoryRow => ({
	key,
	label,
	talents,
	changes,
	cells,
});

const gainLabel = (gainPct: number) => `${gainPct > 0 ? '+' : ''}${gainPct.toFixed(2)}%`;

// The table's props, every cell tinted on one scale from the lowest gain to the highest, 0 included, as the dialog
// builds them; `currentDps` is the current gear's DPS in each column.
const tableArgs = (columns: BulkCompareTableColumn[], currentDps: number[], rows: StoryRow[]): BulkCompareTableProps => {
	const gains = rows.flatMap(line => line.cells.flatMap(entry => (entry ? [entry.gainPct] : [])));
	const scale: GainScale = { low: Math.min(0, ...gains), high: Math.max(0, ...gains) };
	const tableCell = ({ gainPct, best }: StoryCell, index: number): BulkCompareTableCell => {
		const dps = (currentDps[index] * (1 + gainPct / 100)).toFixed(2);
		const gain = gainLabel(gainPct);
		return { gain, dps, tint: gainTint(gainPct, scale), best, tooltip: `#1 · ${dps} DPS ±12 per fight · ${gain}` };
	};
	return {
		columns,
		rows: rows.map(line => ({ ...line, cells: line.cells.map((entry, index) => entry && tableCell(entry, index)) })),
		legend: {
			low: gainLabel(scale.low),
			high: gainLabel(scale.high),
			lowTint: gainTint(scale.low, scale),
			highTint: gainTint(scale.high, scale),
		},
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
	row(
		'current',
		'Current gear',
		[],
		COLUMNS.map((_, index) => cell(0, index === 3)),
	),
	row(
		'set-1',
		'S1',
		['Stormrage’s Talisman of Seething', 'Ring of the Godslayer', 'Badge of the Swarmguard'],
		[cell(8.58, true), cell(1.55, true), cell(1.2), cell(-0.05), cell(1.55, true), cell(1.75, true), cell(2.4, true), cell(0.4)],
	),
	row(
		'set-2',
		'S2',
		['Badge of the Swarmguard', 'Cloak of Concentrated Hatred (Enchant Cloak - Greater Agility)'],
		[cell(4.1), null, cell(1.9, true), cell(-0.3), null, cell(0.7), cell(0.5), cell(0.9, true)],
	),
	row('set-3', 'S3', ['Kiss of the Spider'], [null, cell(0.52), null, null, cell(0.52), null, null, null]),
	row('set-4', 'S4', ['Hand of Justice'], [cell(-0.4), null, cell(-1.4), cell(-3.2), null, cell(-0.7), null, null]),
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
	row(
		'current',
		'Current gear',
		[],
		GAIN_COLUMNS.map(() => cell(0)),
	),
	row('set-1', 'S1', ['Slayer’s Crest', 'Kiss of the Spider'], [cell(11.18, true), cell(8.79, true), cell(7.54, true), cell(8.9, true)]),
	row('set-2', 'S2', ['Slayer’s Crest'], [cell(6.07), cell(4.85), cell(4.29), cell(4.63)]),
	row('set-3', 'S3', ['Kiss of the Spider'], [cell(5.89), cell(4.76), cell(4.1), cell(4.81)]),
	row('set-4', 'S4', ['Slayer’s Crest', 'Frozen Heart of the Mountain'], [cell(5.42), cell(4.25), cell(3.7), cell(4.07)]),
	row('set-5', 'S5', ['Kiss of the Spider', 'Frozen Heart of the Mountain'], [cell(5.27), cell(4.26), cell(3.59), cell(4.23)]),
];

// Two talent setups in each encounter: every row names its setup, and the current gear with Arms is a set with no item
// changes.
const TALENT_ROWS = [
	row(
		'current',
		'Current gear',
		[],
		GAIN_COLUMNS.map(() => cell(0)),
		'Current talents',
	),
	row('set-1', 'S1', ['Slayer’s Crest'], [cell(9.4, true), cell(7.9, true), cell(6.8, true), cell(7.2, true)], 'Arms'),
	row('set-2', 'S2', ['No gear changes'], [cell(3.1), cell(2.6), cell(2.2), cell(2.5)], 'Arms'),
	row('set-3', 'S3', ['Slayer’s Crest'], [cell(6.07), cell(4.85), cell(4.29), cell(4.63)], 'Current talents'),
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

// Every set gains, so every cell but the current gear's is green, the largest gain darkest, and the legend starts at 0%.
export const AllGains: Story = {
	args: tableArgs(GAIN_COLUMNS, GAIN_CURRENT_DPS, GAIN_ROWS),
};

export const TalentSetups: Story = {
	args: tableArgs(GAIN_COLUMNS, GAIN_CURRENT_DPS, TALENT_ROWS),
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
		row(
			'set-5',
			'S5',
			['Thunderfury, Blessed Blade of the Windseeker', 'Spaulders of Valor (Enchant Shoulders - Zandalar Signet of Might)'],
			[cell(6.2), cell(1.1), cell(0.8), cell(-0.6), cell(1.3), cell(1.5), cell(2.1), cell(0.7)],
		),
	]),
};

export const TwoEncounters: Story = {
	args: tableArgs(
		COLUMNS.slice(0, 2),
		CURRENT_DPS,
		ROWS.map(line => ({ ...line, cells: line.cells.slice(0, 2) })),
	),
};

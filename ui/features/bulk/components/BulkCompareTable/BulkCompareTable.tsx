import i18n from '@i18n/config';
import { Table, TableCell, TableHeaderCell, TableRow } from '@ui-kit/Table';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { cssVars } from '@ui-kit/utils/css';
import clsx from 'clsx';
import { useId } from 'react';

import type { GainTint } from '../../model/compare';

export interface BulkCompareTableColumn {
	label: string;
	/** The encounter's length as "m:ss", or how a health-based one ends. */
	length: string;
}

export interface BulkCompareTableCell {
	/** The gain over the current gear in percent, signed. */
	gain: string;
	dps: string;
	tint: GainTint;
	/** The encounter's highest DPS. */
	best: boolean;
	/** The rank, DPS, spread and gain, shown on hover and focus. */
	tooltip: string;
}

export interface BulkCompareTableRow {
	key: string;
	label: string;
	/** The items the set wears where the current gear wears something else. */
	changes: readonly string[];
	/** One per column; null where the encounter did not rank the set. */
	cells: ReadonlyArray<BulkCompareTableCell | null>;
}

/** The table's lowest gain and its highest, each with the tint it takes. */
interface BulkCompareTableLegend {
	low: string;
	high: string;
	lowTint: GainTint;
	highTint: GainTint;
}

export interface BulkCompareTableProps {
	columns: readonly BulkCompareTableColumn[];
	/** The current gear's row first. */
	rows: readonly BulkCompareTableRow[];
	legend: BulkCompareTableLegend;
}

const t = (key: string) => i18n.t(`bulk_tab.compare.${key}`);

// A diverging scale on the result rows' gain and loss tones, neutral next to no change. The loss tone is the darker, so
// its steps run deeper; the darkest of either keeps the foreground text readable.
const TINT_CLASSES: Record<GainTint, string> = {
	[-5]: 'bg-danger/50',
	[-4]: 'bg-danger/40',
	[-3]: 'bg-danger/30',
	[-2]: 'bg-danger/20',
	[-1]: 'bg-danger/10',
	0: 'bg-white/5',
	1: 'bg-success/8',
	2: 'bg-success/16',
	3: 'bg-success/24',
	4: 'bg-success/32',
	5: 'bg-success/40',
};
const TINTS: GainTint[] = [-5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5];

// From md up, the set column's width (its col's md:w-64) and each encounter column's narrowest, in spacing steps; past
// that the table scrolls sideways.
const SET_COLUMN_STEPS = 64;
const MIN_COLUMN_STEPS = 36;

const CELL_CLASS = 'py-4 text-center text-foreground tabular-nums data-best:outline-2 data-best:-outline-offset-2 data-best:outline-foreground';
const GAP_CLASS = 'border border-dashed border-surface-border py-4 text-center text-muted';
const STICKY_CLASS = 'sticky left-0 z-10 bg-background text-left';
// The ring fills the border spacing around each header cell, so rows scrolling under the header do not show through it.
const HEADER_CLASS = 'sticky top-0 z-20 bg-background align-bottom text-base ring-4 ring-background';

const Cell = ({ cell, tooltipId }: { cell: BulkCompareTableCell | null; tooltipId: string }) =>
	cell ? (
		<TableCell
			tabIndex={0}
			className={clsx(CELL_CLASS, TINT_CLASSES[cell.tint])}
			data-best={cell.best ? '' : undefined}
			{...tooltipAnchorProps(tooltipId, cell.tooltip)}>
			<span className="block text-lg font-bold">{cell.gain}</span>
			<span className="block text-sm">{cell.dps}</span>
		</TableCell>
	) : (
		<TableCell className={GAP_CLASS} data-gap="">
			<span aria-hidden>—</span>
			<span className="sr-only">{t('not_ranked')}</span>
		</TableCell>
	);

// The strip holds only the tints from the lowest gain's to the highest's, so a side no cell reaches is left out.
const Legend = ({ low, high, lowTint, highTint }: BulkCompareTableLegend) => (
	<div className="flex flex-wrap items-center gap-2 text-xs text-muted">
		{low}
		<span aria-hidden className="flex">
			{TINTS.filter(tint => tint >= lowTint && tint <= highTint).map(tint => (
				<span key={tint} className={clsx('h-4 w-6', TINT_CLASSES[tint])} />
			))}
		</span>
		{high}
		<span aria-hidden className="ml-2 h-4 w-6 outline-2 -outline-offset-2 outline-foreground" />
		{t('table_legend_best')}
		<span aria-hidden className="ml-2 h-4 w-6 border border-dashed border-surface-border" />
		{t('table_legend_gap')}
	</div>
);

export const BulkCompareTable = ({ columns, rows, legend }: BulkCompareTableProps) => {
	const tooltipId = useId();
	return (
		// The table scrolls inside the dialog's height, so the header row stays at its top and the legend stays below it. The
		// scroller is positioned so the gap cells' screen-reader text stays inside it and does not scroll the dialog body.
		<div className="flex min-h-0 min-w-0 flex-col gap-4">
			<p className="mb-0 text-sm">{t('table_caption')}</p>
			<div className="@container relative min-h-0 min-w-0 overflow-auto">
				{/* A fixed layout gives the set column its width and shares the rest equally among the encounters. Below md
				every column takes a third of the scroller's width, so two encounters show next to the set column and the rest
				scroll sideways. */}
				<Table
					variant="plain"
					dense
					className="min-w-(--phone-min-width) table-fixed border-separate border-spacing-1 md:min-w-(--min-width)"
					style={cssVars({
						'--min-width': `calc((${SET_COLUMN_STEPS} + ${columns.length} * ${MIN_COLUMN_STEPS}) * var(--spacing))`,
						'--phone-min-width': `calc(${columns.length + 1} * 100cqw / 3)`,
					})}>
					<colgroup>
						<col className="w-[calc(100cqw/3)] md:w-64" />
						{columns.map((_, index) => (
							<col key={index} />
						))}
					</colgroup>
					<thead>
						<TableRow>
							<TableHeaderCell scope="col" className={clsx(STICKY_CLASS, HEADER_CLASS, 'z-30')}>
								{t('table_set')}
							</TableHeaderCell>
							{columns.map((column, index) => (
								<TableHeaderCell key={index} scope="col" wrap className={HEADER_CLASS}>
									{column.label}
									<span className="block text-sm font-normal text-muted">{column.length}</span>
								</TableHeaderCell>
							))}
						</TableRow>
					</thead>
					<tbody>
						{rows.map(row => (
							<TableRow key={row.key}>
								<TableHeaderCell
									scope="row"
									wrap
									className={clsx(STICKY_CLASS, 'align-middle font-normal max-md:wrap-break-word max-md:hyphens-auto')}>
									{row.label}
									{row.changes.map((item, index) => (
										<span key={index} className="block text-xs text-muted md:truncate" title={item}>
											{item}
										</span>
									))}
								</TableHeaderCell>
								{row.cells.map((cell, index) => (
									<Cell key={index} cell={cell} tooltipId={tooltipId} />
								))}
							</TableRow>
						))}
					</tbody>
				</Table>
			</div>
			<Legend {...legend} />
			{/* To the cell's left, so a cell at the table's right edge never pushes it past the dialog's; react-tooltip flips it
			right where the left has no room. */}
			<Tooltip id={tooltipId} place="left" />
		</div>
	);
};

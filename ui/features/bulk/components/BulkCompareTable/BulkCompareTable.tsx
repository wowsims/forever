import i18n from '@i18n/config';
import { Table, TableCell, TableHeaderCell, TableRow } from '@ui-kit/Table';

export interface BulkCompareTableEncounter {
	label: string;
	summary: string;
}

export interface BulkCompareTableCell {
	dps: string;
	margin: string;
	/** The gain over the current gear with its percentage; null for the current gear. */
	gain: string | null;
}

export interface BulkCompareTableRow {
	key: string;
	label: string;
	/** The items the set wears where the current gear wears something else. */
	changes: readonly string[];
	/** One per encounter, in the encounters' order; null where the encounter did not rank the set. */
	cells: ReadonlyArray<BulkCompareTableCell | null>;
}

export interface BulkCompareTableProps {
	encounters: readonly BulkCompareTableEncounter[];
	rows: readonly BulkCompareTableRow[];
	/** The key of the set picked out; its row is marked. */
	highlighted: string | null;
	onHighlight: (key: string | null) => void;
}

const t = (key: string) => i18n.t(`bulk_tab.compare.${key}`);

export const BulkCompareTable = ({ encounters, rows, highlighted, onHighlight }: BulkCompareTableProps) => (
	<div className="min-w-0 overflow-x-auto">
		<Table dense>
			<caption className="caption-bottom pt-2 text-left text-sm text-muted">{t('table_caption')}</caption>
			<thead>
				<TableRow>
					<TableHeaderCell scope="col">{t('table_set')}</TableHeaderCell>
					{encounters.map((encounter, index) => (
						<TableHeaderCell key={index} scope="col" wrap>
							{encounter.label}
							<span className="block text-xs font-normal text-muted">{encounter.summary}</span>
						</TableHeaderCell>
					))}
				</TableRow>
			</thead>
			<tbody>
				{rows.map(row => (
					<TableRow
						key={row.key}
						className="data-active:bg-scrim"
						data-active={row.key === highlighted ? '' : undefined}
						onMouseEnter={() => onHighlight(row.key)}
						onMouseLeave={() => onHighlight(null)}>
						<TableHeaderCell scope="row" wrap className="min-w-48 text-left">
							{row.label}
							{row.changes.length > 0 && <span className="block text-xs font-normal text-muted">{row.changes.join(', ')}</span>}
						</TableHeaderCell>
						{row.cells.map((cell, index) => (
							<TableCell key={index} className="whitespace-nowrap tabular-nums">
								{cell && (
									<>
										{cell.dps} <span className="text-xs text-muted">±{cell.margin}</span>
										{cell.gain !== null && <span className="block text-xs">{cell.gain}</span>}
									</>
								)}
							</TableCell>
						))}
					</TableRow>
				))}
			</tbody>
		</Table>
	</div>
);

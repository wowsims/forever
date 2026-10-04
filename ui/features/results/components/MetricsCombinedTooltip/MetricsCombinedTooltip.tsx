import type { SpellSchool } from '@generated/proto/common';
import i18n from '@i18n/config';
import { formatToCompactNumber } from '@sim/utils/format';
import { Table, TableCell, TableHeaderCell, TableRow } from '@ui-kit/Table';
import { Fragment } from 'react';

import { MetricsTotalBar } from '../MetricsTotalBar';

export interface MetricsCombinedTooltipEntry {
	name: string;
	value: number;
	percentage: number;
	average?: number;
}

export interface MetricsCombinedTooltipGroup {
	name?: string;
	className?: string;
	totalPercentage: number;
	spellSchool?: SpellSchool | null;
	data: Array<MetricsCombinedTooltipEntry>;
}

export interface MetricsCombinedTooltipProps {
	groups: Array<MetricsCombinedTooltipGroup>;
	/** Overrides for the Type / Count / Average headers, by position. */
	headerValues?: Array<string | undefined>;
}

/** The tooltip *body* only — a nested breakdown table. It parameterises the rows and the bars; the anchor, the open/close and the veto belong to the column's one `<Tooltip>`, which is what keeps a 19-row table at seven tooltip instances instead of 133. */
export const MetricsCombinedTooltip = ({ groups, headerValues }: MetricsCombinedTooltipProps) => {
	const displayGroups = groups
		.filter(group => group.data.some(entry => entry.value))
		.map(group => ({ ...group, data: group.data.filter(entry => entry.value) }));
	const hasAverageColumn = displayGroups.some(group => group.data.some(entry => typeof entry.average === 'number'));

	return (
		<Table striped hover className="ui-metrics-table" testId="metrics-table">
			<thead data-testid="metrics-table-header">
				<TableRow testId="metrics-table-header-row">
					<TableHeaderCell testId="metrics-table-header-cell" className="ui-metrics-header-cell">
						{headerValues?.[0] || i18n.t('results_tab.details.tooltip_table.type')}
					</TableHeaderCell>
					<TableHeaderCell testId="metrics-table-header-cell" className="ui-metrics-header-cell">
						{headerValues?.[1] || i18n.t('results_tab.details.tooltip_table.count')}
					</TableHeaderCell>
					{hasAverageColumn && (
						<TableHeaderCell testId="metrics-table-header-cell" className="ui-metrics-header-cell">
							{headerValues?.[2] || i18n.t('results_tab.details.tooltip_table.average')}
						</TableHeaderCell>
					)}
				</TableRow>
			</thead>
			<tbody data-testid="metrics-table-body">
				{displayGroups.map(({ name: groupName, className, data, spellSchool, totalPercentage }, groupIndex) => {
					const maxValue = Math.max(...data.map(entry => entry.value));
					const columnCount = data.some(entry => typeof entry.average === 'number') ? 3 : 2;
					return (
						<Fragment key={groupName ?? groupIndex}>
							{groupName && displayGroups.length > 1 && (
								<TableRow testId="metrics-table-group-header" className={className}>
									<TableHeaderCell className="p-0 text-left font-normal" colSpan={columnCount}>
										{groupName}
									</TableHeaderCell>
								</TableRow>
							)}
							{[...data]
								.sort((a, b) => b.value - a.value)
								.map(({ name, value, percentage, average }) => (
									<TableRow className={className} key={name}>
										<TableCell className="ui-metrics-cell">{name}</TableCell>
										<TableCell className="ui-metrics-cell">
											<MetricsTotalBar
												spellSchool={spellSchool}
												percentage={(percentage / totalPercentage) * 100}
												max={maxValue}
												total={value}
												value={value}
											/>
										</TableCell>
										{typeof average === 'number' && <TableCell className="ui-metrics-cell">{formatToCompactNumber(average)}</TableCell>}
									</TableRow>
								))}
						</Fragment>
					);
				})}
			</tbody>
		</Table>
	);
};

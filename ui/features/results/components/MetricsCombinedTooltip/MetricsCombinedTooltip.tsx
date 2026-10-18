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
	min?: number;
	max?: number;
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
	/** Overrides for the Type and Count headers, by position. The Min, Average and Max headers are fixed. */
	headerValues?: Array<string | undefined>;
}

type ValueField = 'min' | 'average' | 'max';

const VALUE_FIELDS: Array<ValueField> = ['min', 'average', 'max'];

/** The tooltip *body* only — a nested breakdown table. It parameterises the rows and the bars; the anchor, the open/close and the veto belong to the column's one `<Tooltip>`, which is what keeps a 19-row table at seven tooltip instances instead of 133. */
export const MetricsCombinedTooltip = ({ groups, headerValues }: MetricsCombinedTooltipProps) => {
	const displayGroups = groups
		.filter(group => group.data.some(entry => entry.value))
		.map(group => ({ ...group, data: group.data.filter(entry => entry.value) }));
	const valueColumns = VALUE_FIELDS.filter(field => displayGroups.some(group => group.data.some(entry => typeof entry[field] === 'number')));
	const columnCount = 2 + valueColumns.length;

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
					{valueColumns.map(field => (
						<TableHeaderCell key={field} testId="metrics-table-header-cell" className="ui-metrics-header-cell">
							{i18n.t(`results_tab.details.tooltip_table.${field}`)}
						</TableHeaderCell>
					))}
				</TableRow>
			</thead>
			<tbody data-testid="metrics-table-body">
				{displayGroups.map(({ name: groupName, className, data, spellSchool, totalPercentage }, groupIndex) => {
					const maxValue = Math.max(...data.map(entry => entry.value));
					return (
						<Fragment key={groupName ?? groupIndex}>
							{groupName && displayGroups.length > 1 && (
								<TableRow testId="metrics-table-group-header" className={className}>
									<TableHeaderCell
										className="border-b border-border-muted px-2 pt-2 pb-1 text-left font-medium text-muted"
										colSpan={columnCount}>
										{groupName}
									</TableHeaderCell>
								</TableRow>
							)}
							{[...data]
								.sort((a, b) => b.value - a.value)
								.map(entry => (
									<TableRow className={className} key={entry.name}>
										<TableCell className="ui-metrics-cell whitespace-nowrap">{entry.name}</TableCell>
										<TableCell className="ui-metrics-cell">
											<MetricsTotalBar
												spellSchool={spellSchool}
												percentage={(entry.percentage / totalPercentage) * 100}
												max={maxValue}
												total={entry.value}
												value={entry.value}
											/>
										</TableCell>
										{valueColumns.map(field => {
											const fieldValue = entry[field];
											return (
												<TableCell key={field} className="ui-metrics-cell">
													{typeof fieldValue === 'number' ? formatToCompactNumber(fieldValue) : ''}
												</TableCell>
											);
										})}
									</TableRow>
								))}
						</Fragment>
					);
				})}
			</tbody>
		</Table>
	);
};

import i18n from '@i18n/config';
import { ContentBlock } from '@ui-kit/ContentBlock';
import { DialogColumns } from '@ui-kit/DialogColumns';
import { MultiComboBox } from '@ui-kit/MultiComboBox';
import { NumberField } from '@ui-kit/TextField';
import { type ReactNode, useMemo } from 'react';

export interface StatSweepPanelProps {
	/** Keyed by `UnitStat.getKey()`. */
	statOptions: ReadonlyArray<{ value: string; label: string }>;
	stats: ReadonlyArray<string>;
	onStatsChange: (stats: string[]) => void;
	range: number;
	onRangeChange: (range: number) => void;
	step: number;
	onStepChange: (step: number) => void;
	/** Percent stats step in fractions of a point. */
	fractional?: boolean;
	precision: ReactNode;
	chart: ReactNode;
	history: ReactNode;
	running: boolean;
}

const t = (key: string) => i18n.t(`sidebar.buttons.stat_sweep.${key}`);

export const StatSweepPanel = ({
	statOptions,
	stats,
	onStatsChange,
	range,
	onRangeChange,
	step,
	onStepChange,
	fractional,
	precision,
	chart,
	history,
	running,
}: StatSweepPanelProps) => {
	// The combo box picks numbers, so each option goes by its index.
	const statValues = useMemo(() => statOptions.map((option, index) => ({ name: option.label, value: index })), [statOptions]);
	const picked = useMemo(() => statOptions.flatMap((option, index) => (stats.includes(option.value) ? [index] : [])), [statOptions, stats]);

	return (
		<DialogColumns.Root>
			<DialogColumns.Main className="flex flex-col gap-4">
				<div className="flex flex-wrap items-start gap-4">
					<MultiComboBox
						id="stat-sweep-stats"
						fitOptions
						label={t('stats.label')}
						placeholder={t('stats.placeholder')}
						emptyText={t('stats.empty')}
						removeLabel={t('stats.remove')}
						openLabel={t('stats.open')}
						values={statValues}
						value={picked}
						onChange={indices => onStatsChange(indices.map(index => statOptions[index].value))}
						disabled={running}
						testId="stat-sweep-stats"
					/>
					<NumberField
						id="stat-sweep-range"
						label={t('range')}
						float={fractional}
						positive
						disabled={running}
						value={range}
						onChange={onRangeChange}
					/>
					<NumberField id="stat-sweep-step" label={t('step')} float={fractional} positive disabled={running} value={step} onChange={onStepChange} />
				</div>
				{chart}
			</DialogColumns.Main>
			<DialogColumns.Sidebar>
				<ContentBlock config={{ header: { title: t('settings') } }}>{precision}</ContentBlock>
				{history}
			</DialogColumns.Sidebar>
		</DialogColumns.Root>
	);
};

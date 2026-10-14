import i18n from '@i18n/config';
import type { StatSweepMetric } from '@sim/state/sim_store';
import { Button } from '@ui-kit/Button';
import { ContentBlock } from '@ui-kit/ContentBlock';
import { Icon } from '@ui-kit/Icon';
import { Pill } from '@ui-kit/Pill';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx from 'clsx';
import { useId, useMemo, useState } from 'react';

import { SERIES_SWATCH_CLASSES } from '../../model/colors';
import type { SweepBatch } from '../../model/stat_sweep';

export interface StatSweepHistoryEntry {
	id: string;
	label: string;
	/** The swept stat, as its batch's header lists it. */
	statLabel: string;
	/** Epoch milliseconds. */
	createdAt: number;
	overlaid: boolean;
	/** Set while the run is drawn on the chart. */
	colorSlot?: number;
	/** The metric plotted when the run started, which Target ± followed; tagged on the row unless it is DPS. */
	metric: StatSweepMetric;
}

export interface StatSweepHistoryProps {
	/** Newest sweep first, each with its runs in run order. */
	batches: ReadonlyArray<SweepBatch<StatSweepHistoryEntry>>;
	onOverlayChange: (ids: string[], overlaid: boolean) => void;
	onDelete: (ids: string[]) => void;
	/** Exports the runs the chart shows; unset while it shows none. */
	onExportCsv?: () => void;
}

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`sidebar.buttons.stat_sweep.${key}`, options);

// A batch header names this many stats and counts the rest.
const HEADER_STATS = 3;

// The swatch, the text, the eye, the delete button, then a batch's chevron. A batch's header button spans the whole row, under its eye and delete button.
const ROW_CLASS = 'grid grid-cols-[auto_minmax(0,1fr)_auto_auto_1.25em] items-center gap-x-2 text-ui';

const batchTitle = (runs: ReadonlyArray<StatSweepHistoryEntry>) => {
	const stats = runs.map(run => run.statLabel);
	const named = stats.slice(0, HEADER_STATS).join(', ');
	return stats.length > HEADER_STATS ? t('history.more_stats', { stats: named, count: stats.length - HEADER_STATS }) : named;
};

const RunDetails = ({ createdAt, metric, dateFormat }: Pick<StatSweepHistoryEntry, 'createdAt' | 'metric'> & { dateFormat: Intl.DateTimeFormat }) => (
	<span className="col-span-2 col-start-1 flex items-center gap-2 text-xs text-muted">
		{dateFormat.format(createdAt)}
		{metric !== 'dps' && (
			<Pill size="xs" tone="primary" title={t('history.metric')} testId="stat-sweep-history-metric">
				{i18n.t(`sidebar.results.metrics.${metric}.label`)}
			</Pill>
		)}
	</span>
);

interface EyeToggleProps {
	label: string;
	/** How many of the runs it stands for are drawn. */
	drawn: number;
	total: number;
	onChange: (overlaid: boolean) => void;
}

/** The eye that draws or hides the runs it stands for; a click on a mixed group draws them all. */
const EyeToggle = ({ label, drawn, total, onChange }: EyeToggleProps) => {
	const all = drawn === total;
	const mixed = drawn > 0 && !all;
	return (
		<Button
			iconOnly
			aria-label={t(mixed ? 'history.show_hidden' : all ? 'history.hide' : 'history.show', { label })}
			aria-pressed={mixed ? 'mixed' : all}
			data-mixed={mixed || undefined}
			className="data-mixed:opacity-50"
			onClick={() => onChange(!all)}>
			<Icon name={drawn > 0 ? 'eye' : 'eye-slash'} style="base" />
		</Button>
	);
};

interface RunRowProps extends Pick<StatSweepHistoryProps, 'onOverlayChange' | 'onDelete'> {
	run: StatSweepHistoryEntry;
}

const RunRow = ({ run, onOverlayChange, onDelete }: RunRowProps) => (
	<li className={ROW_CLASS}>
		<span aria-hidden className={clsx('size-2 rounded-full', run.colorSlot === undefined ? 'bg-muted/40' : SERIES_SWATCH_CLASSES[run.colorSlot])} />
		<span className="min-w-0 truncate" title={run.label}>
			{run.label}
		</span>
		<EyeToggle label={run.label} drawn={run.overlaid ? 1 : 0} total={1} onChange={overlaid => onOverlayChange([run.id], overlaid)} />
		<Button iconOnly variant="link-danger" aria-label={t('history.delete', { label: run.label })} onClick={() => onDelete([run.id])}>
			<Icon name="xmark" />
		</Button>
	</li>
);

interface BatchRowProps extends Pick<StatSweepHistoryProps, 'onOverlayChange' | 'onDelete'> {
	batch: SweepBatch<StatSweepHistoryEntry>;
	open: boolean;
	onOpenChange: (open: boolean) => void;
	dateFormat: Intl.DateTimeFormat;
	statsTooltipId: string;
}

const BatchRow = ({ batch, open, onOpenChange, onOverlayChange, onDelete, dateFormat, statsTooltipId }: BatchRowProps) => {
	const runsId = useId();
	const [first] = batch.runs;
	const ids = batch.runs.map(run => run.id);
	const ticked = batch.runs.filter(run => run.overlaid).length;
	const title = batchTitle(batch.runs);
	return (
		<li className={ROW_CLASS} data-testid="stat-sweep-history-batch">
			<Button
				variant="unstyled"
				className="col-span-full row-span-2 row-start-1 grid cursor-pointer grid-cols-subgrid grid-rows-subgrid items-center text-left text-foreground hover:text-link-hover focus-visible:focus-ring"
				aria-expanded={open}
				aria-controls={runsId}
				onClick={() => onOpenChange(!open)}
				data-batch-id={batch.id}
				{...tooltipAnchorProps(statsTooltipId)}>
				<span className="col-span-2 min-w-0 truncate">{title}</span>
				<Icon name="chevron-down" fixedWidth className={clsx('col-start-5 transition-transform duration-150', !open && '-rotate-90')} />
				<RunDetails createdAt={first.createdAt} metric={first.metric} dateFormat={dateFormat} />
			</Button>
			<span className="relative z-10 col-span-2 col-start-3 row-start-1 flex items-center gap-x-2">
				<EyeToggle label={title} drawn={ticked} total={ids.length} onChange={overlaid => onOverlayChange(ids, overlaid)} />
				<Button iconOnly variant="link-danger" aria-label={t('history.delete_batch', { label: title })} onClick={() => onDelete(ids)}>
					<Icon name="xmark" />
				</Button>
			</span>
			<ul id={runsId} hidden={!open} className="col-span-full m-0 mt-2 grid list-none gap-2 p-0">
				{batch.runs.map(run => (
					<RunRow key={run.id} run={run} onOverlayChange={onOverlayChange} onDelete={onDelete} />
				))}
			</ul>
		</li>
	);
};

type Batches = StatSweepHistoryProps['batches'];

// The newest run's time: a newer run arriving raises it, and deleting runs never does.
const newestAt = (batches: Batches) => (batches.length ? Math.max(...batches[0].runs.map(run => run.createdAt)) : -Infinity);
const newestOnly = (batches: Batches) => ({ at: newestAt(batches), ids: batches.length ? [batches[0].id] : [] });

export const StatSweepHistory = ({ batches, onOverlayChange, onDelete, onExportCsv }: StatSweepHistoryProps) => {
	const language = i18n.language;
	const dateFormat = useMemo(() => new Intl.DateTimeFormat(language, { dateStyle: 'short', timeStyle: 'short' }), [language]);
	const statsTooltipId = useId();
	// The newest batch opens and the older ones close each time a newer run arrives; a delete leaves them as they are.
	const [expanded, setExpanded] = useState(() => newestOnly(batches));
	if (newestAt(batches) > expanded.at) setExpanded(newestOnly(batches));
	const setOpen = (id: string, open: boolean) =>
		setExpanded(current => ({ ...current, ids: open ? [...current.ids, id] : current.ids.filter(other => other !== id) }));
	return (
		<ContentBlock config={{ header: { title: t('history.title') } }}>
			<div className="flex flex-col gap-4">
				{batches.length ? (
					<ul className="m-0 grid list-none gap-2 p-0">
						{batches.map(batch => (
							<BatchRow
								key={batch.id}
								batch={batch}
								open={expanded.ids.includes(batch.id)}
								onOpenChange={open => setOpen(batch.id, open)}
								onOverlayChange={onOverlayChange}
								onDelete={onDelete}
								dateFormat={dateFormat}
								statsTooltipId={statsTooltipId}
							/>
						))}
					</ul>
				) : (
					<p className="mb-0 text-ui text-muted">{t('history.empty')}</p>
				)}
				<Tooltip
					id={statsTooltipId}
					align="start"
					render={({ activeAnchor }) => {
						const runs = batches.find(batch => batch.id === activeAnchor?.getAttribute('data-batch-id'))?.runs;
						return (
							runs && (
								<ul className="m-0 list-none p-0">
									{runs.map(run => (
										<li key={run.id}>{run.statLabel}</li>
									))}
								</ul>
							)
						);
					}}
				/>
				<Button variant="outline-primary" size="sm" className="self-start" disabled={!onExportCsv} onClick={onExportCsv}>
					<Icon name="download" />
					{t('export_csv')}
				</Button>
			</div>
		</ContentBlock>
	);
};

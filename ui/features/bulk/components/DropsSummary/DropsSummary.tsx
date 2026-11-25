import i18n from '@i18n/config';
import { formatToNumber } from '@sim/utils/format';
import { Box } from '@ui-kit/Box';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { type ReactNode, useId } from 'react';

import { formatBulkDps, formatBulkPct } from '../../model/results/format';

export interface DropsSummaryViewProps {
	current: { avg: number; plusMinus: number };
	/** Null when nothing upgrades. */
	bestDrop: { name: string; dps: number; pct: number } | null;
	/** Null when nothing upgrades, or when no card holds the best drop (the tile then shows a dash). Its gain is the best drop's. */
	bestSource: { label: string; detail: string } | null;
}

const t = (key: string) => i18n.t(`bulk_tab.drops.results.${key}`);

interface TileProps {
	title: string;
	testId: string;
	children: ReactNode;
}

const Tile = ({ title, testId, children }: TileProps) => (
	<Box className="grid content-start gap-1" testId={testId}>
		<span className="text-xs text-gray-400">{title}</span>
		{children}
	</Box>
);

const None = () => <span className="text-gray-400">{t('no_upgrade')}</span>;

export const DropsSummaryView = ({ current, bestDrop, bestSource }: DropsSummaryViewProps) => {
	const tooltipId = useId();
	return (
		<div className="grid grid-cols-1 gap-2 sm:grid-cols-3" data-testid="drops-summary">
			<Tile title={t('current_gear')} testId="drops-summary-current">
				<span className="text-xl font-bold">
					{formatBulkDps(current.avg)} <span className="text-sm font-normal">{i18n.t('sidebar.results.metrics.dps.label')}</span>
					{current.plusMinus > 0 && (
						<span className="text-sm font-normal"> (±{formatToNumber(current.plusMinus, { maximumFractionDigits: 0 })})</span>
					)}
				</span>
			</Tile>
			<Tile title={t('best_drop')} testId="drops-summary-best-drop">
				{bestDrop ? (
					<>
						<span className="truncate text-xl font-bold" {...tooltipAnchorProps(tooltipId, bestDrop.name)}>
							{bestDrop.name}
						</span>
						<span className="text-sm">
							<span className="text-success">{formatBulkPct(bestDrop.pct)}</span> {formatBulkDps(bestDrop.dps)}
						</span>
					</>
				) : (
					<None />
				)}
			</Tile>
			<Tile title={t('best_source')} testId="drops-summary-best-source">
				{bestSource && bestDrop ? (
					<>
						<span className="truncate text-xl font-bold" {...tooltipAnchorProps(tooltipId, bestSource.label)}>
							{bestSource.label}
						</span>
						<span className="text-sm">
							<span className="text-success">{formatBulkPct(bestDrop.pct)}</span> <span className="text-gray-400">{bestSource.detail}</span>
						</span>
					</>
				) : bestDrop ? (
					<span className="text-xl text-gray-400">—</span>
				) : (
					<None />
				)}
			</Tile>
			<Tooltip id={tooltipId} />
		</div>
	);
};

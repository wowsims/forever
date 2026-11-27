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
	/**
	 * The card the Recommended order puts first, with its upgrades of its drops, worn ones included, as the card counts them.
	 * Null when no card holds an upgrade; the tile then shows a dash beside a best drop.
	 */
	farmFirst: { label: string; detail: string; upgrades: number; total: number } | null;
}

const t = (key: string, options: { count?: number; total?: number } = {}) => i18n.t(`bulk_tab.drops.results.${key}`, options);

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

export const DropsSummaryView = ({ current, bestDrop, farmFirst }: DropsSummaryViewProps) => {
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
			<Tile title={t('farm_first')} testId="drops-summary-farm-first">
				{farmFirst ? (
					<>
						<span className="truncate text-xl font-bold" {...tooltipAnchorProps(tooltipId, farmFirst.label)}>
							{farmFirst.label}
						</span>
						<span className="text-sm">
							{t('farm_first_upgrades', { count: farmFirst.upgrades, total: farmFirst.total })}
							<span className="text-gray-400"> · {farmFirst.detail}</span>
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

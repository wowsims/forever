import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import clsx from 'clsx';

import type { DropsSourceLabel } from '../../model/drops/sources';
import { formatBulkDps, formatBulkPct } from '../../model/results/format';

export interface DropsBossCardViewProps {
	/** Null on a card with no upgrade. */
	rank: number | null;
	label: DropsSourceLabel;
	/** Every item the source dropped, best first, then the worn ones. */
	strip: ReadonlyArray<{ itemId: number; icon: string; name: string; gainPct: number | null /* null: worn */; upgrade: boolean }>;
	best: { dps: number; pct: number } | null;
	selected: boolean;
	onSelect: () => void;
}

const t = (key: string, options: { count?: number } = {}) => i18n.t(`bulk_tab.drops.results.${key}`, options);

// "N / M upgrades": N counts the strip's upgrades, M its items, worn ones included.
export const DropsBossCardView = ({ rank, label, strip, best, selected, onSelect }: DropsBossCardViewProps) => {
	const upgrades = strip.filter(entry => entry.upgrade).length;
	return (
		<Button
			variant="unstyled"
			className={clsx(
				'flex flex-col gap-2 rounded-sm border border-border bg-background p-2 text-left',
				'hover:border-primary focus-visible:focus-ring data-[selected=true]:border-primary data-[upgrade=none]:opacity-60',
			)}
			data-testid="drops-boss-card"
			data-upgrade={rank === null ? 'none' : undefined}
			data-selected={selected}
			aria-pressed={selected}
			onClick={onSelect}>
			<span className="flex items-baseline gap-2">
				{rank !== null && <span className="text-sm text-gray-400">{rank}</span>}
				<span className="min-w-0 truncate font-bold" data-testid="drops-boss-label" title={label.label}>
					{label.label}
				</span>
			</span>
			<span className="truncate text-xs text-gray-400">{label.detail}</span>
			<span className="flex flex-wrap gap-1" data-testid="drops-boss-strip">
				{strip.map(entry => (
					<span
						key={entry.itemId}
						className="flex flex-col items-center gap-1 text-xs"
						data-testid="drops-boss-strip-item"
						data-item-id={entry.itemId}
						data-worn={entry.gainPct === null ? '' : undefined}
						data-upgrade={entry.upgrade ? '' : undefined}>
						<span
							className={clsx(
								'inline-block size-8 rounded-xs bg-cover bg-center',
								entry.gainPct === null && 'opacity-50',
								entry.upgrade && 'ring-1 ring-success',
							)}
							role="img"
							aria-label={entry.name}
							title={entry.name}
							style={{ backgroundImage: `url('${entry.icon}')` }}
						/>
						<span className={clsx(entry.upgrade ? 'text-success' : 'text-gray-400')}>
							{entry.gainPct === null ? t('worn') : formatBulkPct(entry.gainPct)}
						</span>
					</span>
				))}
			</span>
			<span className="flex items-baseline justify-between gap-2 text-sm">
				{best ? (
					<span>
						{t('best')} <span className="font-bold text-success">{formatBulkPct(best.pct)}</span>{' '}
						<span className="text-xs text-gray-400">{formatBulkDps(best.dps)}</span>
					</span>
				) : (
					<span className="text-gray-400">{t('no_upgrade')}</span>
				)}
				<span className="text-xs text-gray-400" data-testid="drops-boss-upgrades">
					{upgrades} / {strip.length} {t('upgrades', { count: upgrades })}
				</span>
			</span>
		</Button>
	);
};

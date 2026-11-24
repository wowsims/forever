import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { ItemCellIcon } from '@ui-kit/ItemCell';
import clsx from 'clsx';

import type { DropsSourceLabel } from '../../model/drops/sources';
import { formatBulkDps, formatBulkPct } from '../../model/results/format';

export interface DropsBossCardViewProps {
	/** Null on a card with no upgrade. */
	rank: number | null;
	label: DropsSourceLabel;
	/** Every item the source dropped, best first, then the worn ones. */
	strip: ReadonlyArray<{
		itemId: number;
		icon: string;
		name: string;
		/** Absent when the database does not hold the item. */
		ilvl?: number;
		gainPct: number | null /* null: worn */;
		upgrade: boolean;
		/** Equips the item's best row; absent on a worn item. */
		onEquip?: () => void;
	}>;
	best: { dps: number; pct: number } | null;
	selected: boolean;
	onSelect: () => void;
}

const t = (key: string, options: { count?: number; name?: string } = {}) => i18n.t(`bulk_tab.drops.results.${key}`, options);

// "N / M upgrades", plural on M: N counts the strip's upgrades, M its items, worn ones included.
// Each strip entry is a fixed column, its icon a fixed tile with a surface fill and the item level, so an icon that fails to
// load keeps its place over its %. The card selects through a button laid under its content, so each icon's [+] equip
// button can sit on top of it.
export const DropsBossCardView = ({ rank, label, strip, best, selected, onSelect }: DropsBossCardViewProps) => {
	const upgrades = strip.filter(entry => entry.upgrade).length;
	return (
		<div
			className={clsx(
				'relative flex flex-col gap-2 border border-border p-2 text-left text-foreground',
				'hover:border-primary has-focus-visible:border-primary data-[selected=true]:border-primary data-[upgrade=none]:opacity-60',
			)}
			data-testid="drops-boss-card"
			data-upgrade={rank === null ? 'none' : undefined}
			data-selected={selected}>
			<Button
				variant="unstyled"
				className="absolute inset-0 focus-visible:focus-ring"
				aria-pressed={selected}
				aria-label={label.label}
				data-testid="drops-boss-select"
				onClick={onSelect}
			/>
			<span className="flex items-baseline gap-2">
				{rank !== null && <span className="text-sm text-gray-400">{rank}</span>}
				<span className="min-w-0 truncate font-bold" data-testid="drops-boss-label" title={label.label}>
					{label.label}
				</span>
			</span>
			<span className="truncate text-xs text-gray-400">{label.detail}</span>
			<span className="flex flex-wrap gap-2" data-testid="drops-boss-strip">
				{strip.map(entry => (
					<span
						key={entry.itemId}
						className="pointer-events-none relative flex w-14 flex-col items-center gap-1 text-xs"
						data-testid="drops-boss-strip-item"
						data-item-id={entry.itemId}
						data-worn={entry.gainPct === null ? '' : undefined}
						data-upgrade={entry.upgrade ? '' : undefined}>
						<ItemCellIcon ilvl={entry.ilvl} size="sm" className="[--icon-size:--spacing(14)]">
							<span
								className={clsx(
									'block size-full bg-surface bg-cover bg-center',
									entry.gainPct === null && 'opacity-50',
									entry.upgrade && 'ring-1 ring-success',
								)}
								role="img"
								aria-label={entry.name}
								title={entry.name}
								style={entry.icon ? { backgroundImage: `url('${entry.icon}')` } : undefined}
							/>
							{entry.onEquip && (
								<Button
									variant="unstyled"
									className="pointer-events-auto absolute right-0 bottom-0 flex size-4 items-center justify-center bg-black/75 text-2xs text-white hover:text-primary focus-visible:focus-ring"
									aria-label={t('equip_item', { name: entry.name })}
									title={t('equip_item', { name: entry.name })}
									data-testid="drops-boss-equip"
									onClick={entry.onEquip}>
									<Icon name="plus" />
								</Button>
							)}
						</ItemCellIcon>
						<span className={clsx('whitespace-nowrap tabular-nums', entry.upgrade ? 'text-success' : 'text-gray-400')}>
							{entry.gainPct === null ? t('worn') : formatBulkPct(entry.gainPct)}
						</span>
					</span>
				))}
			</span>
			<span className="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1 text-sm">
				{best ? (
					<span>
						{t('best')} <span className="font-bold text-success">{formatBulkPct(best.pct)}</span>{' '}
						<span className="text-xs text-gray-400">{formatBulkDps(best.dps)}</span>
					</span>
				) : (
					<span className="text-gray-400">{t('no_upgrade')}</span>
				)}
				<span className="text-xs text-gray-400" data-testid="drops-boss-upgrades">
					{upgrades} / {strip.length} {t('upgrades', { count: strip.length })}
				</span>
			</span>
		</div>
	);
};

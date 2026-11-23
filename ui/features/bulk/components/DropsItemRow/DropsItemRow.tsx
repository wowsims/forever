import i18n from '@i18n/config';
import { useActionId } from '@sim/hooks/useActionId';
import { ActionId } from '@sim/proto/action_id';
import type { EquippedItem } from '@sim/proto/equipped_item';
import type { DeltaText } from '@sim/utils/format';
import { Button } from '@ui-kit/Button';
import { itemQualityClassName } from '@ui-kit/utils/css';
import { WowheadIcon } from '@ui-kit/WowheadIcon';

import { formatBulkItemLabel } from '../../model/results/format';
import { BulkDpsCell } from '../BulkDpsCell';

export interface DropsItemRowViewProps {
	/** Null on the current gear's row. */
	rank: number | null;
	/** The item the row adds to the worn gear; null on the current gear's row. */
	item: EquippedItem | null;
	slot: string;
	source: string;
	moreSources: readonly string[];
	avg: number;
	/** Null on a lower-precision row, which was stopped early and carries no margin. */
	plusMinus: number | null;
	delta: DeltaText;
	significance: string;
	hiddenVariations: number;
	onShowVariations: () => void;
	/** Null on the current gear's row. */
	onEquip: (() => void) | null;
}

// Shared with the header so the columns line up.
const RANK_CLASS = 'w-6 shrink-0';
const SLOT_CLASS = 'w-24 shrink-0';
const SOURCE_CLASS = 'w-40 shrink-0';
const DPS_CLASS = 'w-52 shrink-0';
const EQUIP_CLASS = 'ml-auto w-16 shrink-0';

const t = (key: string, options: { count?: number } = {}) => i18n.t(`bulk_tab.drops.results.${key}`, options);

interface ItemNameProps {
	item: EquippedItem;
}

const ItemName = ({ item }: ItemNameProps) => {
	const { iconUrl, href } = useActionId(ActionId.fromItem(item.item));
	return (
		<span className="flex min-w-0 items-center gap-2">
			<WowheadIcon className="size-8 shrink-0" iconUrl={iconUrl} href={href || undefined} navigate={false} tabIndex={-1} />
			<span className={`min-w-0 wrap-anywhere ${itemQualityClassName(item.item.quality)}`}>{formatBulkItemLabel(item)}</span>
		</span>
	);
};

/** The column titles over the rows, from `md` up where the rows stay on one line. */
export const DropsItemRowHeader = () => (
	<div className="mb-2 hidden items-center gap-4 border-b border-b-border pb-2 text-xs text-gray-400 md:flex" data-testid="drops-result-header">
		<span className={RANK_CLASS}>{t('columns.rank')}</span>
		<span className="min-w-0 flex-1">{t('columns.item')}</span>
		<span className={SLOT_CLASS}>{t('columns.slot')}</span>
		<span className={SOURCE_CLASS}>{t('columns.source')}</span>
		<span className={DPS_CLASS}>
			{t('columns.dps')} / {t('columns.delta')}
		</span>
		<span className={EQUIP_CLASS} />
	</div>
);

export const DropsItemRowView = ({
	rank,
	item,
	slot,
	source,
	moreSources,
	avg,
	plusMinus,
	delta,
	significance,
	hiddenVariations,
	onShowVariations,
	onEquip,
}: DropsItemRowViewProps) => {
	const lowPrecision = plusMinus === null;

	return (
		<div
			className="flex items-center gap-4 not-last:not-only:mb-6 not-last:not-only:border-b not-last:not-only:border-b-border not-last:not-only:pb-6 max-md:flex-wrap"
			data-testid="drops-result-row"
			data-item-id={item?.id}
			data-current={item === null ? '' : undefined}
			data-precision={lowPrecision ? 'low' : undefined}>
			<span className={`${RANK_CLASS} text-sm text-gray-400`}>{rank}</span>
			<div className="flex min-w-0 flex-1 flex-col gap-1">
				{item ? <ItemName item={item} /> : <span className="font-bold">{t('current_gear')}</span>}
				{hiddenVariations > 0 && (
					<span className="text-xs text-gray-400" data-testid="drops-result-hidden">
						{t('variations_hidden', { count: hiddenVariations })}{' '}
						<Button variant="link" size="xs" onClick={onShowVariations}>
							{i18n.t('bulk_tab.drops.setup.show')}
						</Button>
					</span>
				)}
			</div>
			<span className={`${SLOT_CLASS} text-sm`}>{slot}</span>
			<span className={`${SOURCE_CLASS} truncate text-sm text-gray-400`} title={[source, ...moreSources].join('\n')}>
				{source}
				{moreSources.length > 0 && ` +${moreSources.length}`}
			</span>
			<BulkDpsCell
				className={DPS_CLASS}
				avg={avg}
				plusMinus={plusMinus}
				diff={item ? { ...delta, significance } : null}
				lowPrecision={lowPrecision ? { note: t('low_precision'), tooltip: t('low_precision_tooltip') } : undefined}
			/>
			<div className={`${EQUIP_CLASS} flex justify-end`}>
				{onEquip && (
					<Button data-testid="bulk-equip-btn" onClick={onEquip}>
						{i18n.t('bulk_tab.results.equip_button')}
					</Button>
				)}
			</div>
		</div>
	);
};

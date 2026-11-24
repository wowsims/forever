import type { EquippedItem } from '@sim/proto/equipped_item';
import { itemQualityClassName } from '@ui-kit/utils/css';
import clsx from 'clsx';

import { formatBulkItemLabel } from '../../model/results/format';

export interface DropsItemLabelProps {
	/** The item the row adds to the worn gear. */
	item: EquippedItem;
	slot: string;
	/** Empty when no source of the run drops it. */
	source: string;
	moreSources: readonly string[];
}

/** A Drops result row's new item: its name, then its slot and source, with every source in the title. */
export const DropsItemLabel = ({ item, slot, source, moreSources }: DropsItemLabelProps) => (
	<div className="flex min-w-0 flex-col gap-1 text-sm" data-testid="drops-item-label">
		<span className={clsx('wrap-break-word', itemQualityClassName(item.item.quality))}>{formatBulkItemLabel(item)}</span>
		<span className="text-xs text-muted" title={[source, ...moreSources].join('\n')}>
			{[slot, source].filter(Boolean).join(' · ')}
			{moreSources.length > 0 && ` +${moreSources.length}`}
		</span>
	</div>
);

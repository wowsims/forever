import { EnchantLabel } from '@features/gear/components/GearPicker/EnchantLabel';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { itemQualityClassName } from '@ui-kit/utils/css';
import clsx from 'clsx';
import { useId } from 'react';

import { DropsBoePill } from '../DropsBoePill';

export interface DropsItemLabelProps {
	/** The item the row adds to the worn gear. */
	item: EquippedItem;
	slot: string;
	/** Empty when no source of the run drops it. */
	source: string;
	moreSources: readonly string[];
}

/**
 * A Drops result row's new item: its name, marked when it binds on equip, its enchant under it as the gear tab shows one,
 * then its slot and source, with every source in a tooltip on them.
 */
export const DropsItemLabel = ({ item, slot, source, moreSources }: DropsItemLabelProps) => {
	const tooltipId = useId();
	return (
		<div className="flex min-w-0 flex-col gap-1 text-sm" data-testid="drops-item-label">
			<span className="flex flex-wrap items-center gap-2">
				<span className={clsx('wrap-break-word', itemQualityClassName(item.item.quality))}>{item.item.name}</span>
				<DropsBoePill item={item.item} />
			</span>
			<EnchantLabel className="ui-item-picker-label-muted tracking-normal" testId="drops-item-enchant" enchant={item.enchant} />
			<span className="text-xs text-muted" {...tooltipAnchorProps(tooltipId, [source, ...moreSources].join('\n'))}>
				{[slot, source].filter(Boolean).join(' · ')}
				{moreSources.length > 0 && ` +${moreSources.length}`}
			</span>
			<Tooltip id={tooltipId} className="whitespace-pre-line" />
		</div>
	);
};

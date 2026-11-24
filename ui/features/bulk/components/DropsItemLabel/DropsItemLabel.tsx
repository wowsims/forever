import { EnchantLabel } from '@features/gear/components/GearPicker/EnchantLabel';
import { BindType } from '@generated/proto/ui';
import i18n from '@i18n/config';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { Pill } from '@ui-kit/Pill';
import { itemQualityClassName } from '@ui-kit/utils/css';
import clsx from 'clsx';

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
 * then its slot and source, with every source in the title.
 */
export const DropsItemLabel = ({ item, slot, source, moreSources }: DropsItemLabelProps) => (
	<div className="flex min-w-0 flex-col gap-1 text-sm" data-testid="drops-item-label">
		<span className="flex flex-wrap items-center gap-2">
			<span className={clsx('wrap-break-word', itemQualityClassName(item.item.quality))}>{item.item.name}</span>
			{item.item.bindType === BindType.BindTypeOnEquip && (
				<Pill size="xs" rounded="xs" title={i18n.t('bulk_tab.drops.results.boe_tooltip')} testId="drops-item-boe">
					{i18n.t('bulk_tab.drops.results.boe')}
				</Pill>
			)}
		</span>
		<EnchantLabel className="ui-item-picker-label-muted tracking-normal" testId="drops-item-enchant" enchant={item.enchant} />
		<span className="text-xs text-muted" title={[source, ...moreSources].join('\n')}>
			{[slot, source].filter(Boolean).join(' · ')}
			{moreSources.length > 0 && ` +${moreSources.length}`}
		</span>
	</div>
);

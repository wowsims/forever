import { BindType, type UIItem } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { Pill } from '@ui-kit/Pill';

export interface DropsBoePillProps {
	item: Pick<UIItem, 'bindType'>;
}

/** The pill that marks an item binding on equip in the Drops rows, nothing for any other item. */
export const DropsBoePill = ({ item }: DropsBoePillProps) =>
	item.bindType === BindType.BindTypeOnEquip && (
		<Pill size="xs" rounded="xs" className="shrink-0" title={i18n.t('bulk_tab.drops.results.boe_tooltip')} testId="drops-item-boe">
			{i18n.t('bulk_tab.drops.results.boe')}
		</Pill>
	);

import { BindType, type UIItem } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { Pill } from '@ui-kit/Pill';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { useId } from 'react';

export interface DropsBoePillProps {
	item: Pick<UIItem, 'bindType'>;
}

/** The pill that marks an item binding on equip in the Drops rows, nothing for any other item. */
export const DropsBoePill = ({ item }: DropsBoePillProps) => {
	const tooltipId = useId();
	if (item.bindType !== BindType.BindTypeOnEquip) return null;
	return (
		<>
			<Pill
				size="xs"
				rounded="xs"
				className="shrink-0"
				testId="drops-item-boe"
				{...tooltipAnchorProps(tooltipId, i18n.t('bulk_tab.drops.results.boe_tooltip'))}>
				{i18n.t('bulk_tab.drops.results.boe')}
			</Pill>
			<Tooltip id={tooltipId} />
		</>
	);
};

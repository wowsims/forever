import { getEmptySlotIconUrl } from '@features/gear/model/empty_slot_icons';
import i18n from '@i18n/config';
import { translateBulkSlotName } from '@i18n/localization';
import { Button } from '@ui-kit/Button';
import { Meter } from '@ui-kit/Meter';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { useId } from 'react';

import { type DropsItemList, dropsItemsBySlot, dropsTickCount } from '../../model/drops/items';
import { bulkGroupItemSlots } from '../../model/gear/gear_data';

export interface DropsItemsCardViewProps {
	list: DropsItemList;
	/** The ticked, non-worn items the dialog's filters hide; they still run. */
	hidden: number;
	onChoose: () => void;
}

const t = (key: string) => i18n.t(`bulk_tab.drops.setup.${key}`);

// One tile per slot the list reaches: its empty-slot icon over a strip with its ticked count, edged in the
// accent when every item is ticked and dimmed when none are.
export const DropsItemsCardView = ({ list, hidden, onChoose }: DropsItemsCardViewProps) => {
	const tooltipId = useId();
	const { ticked, total } = dropsTickCount(list.items);
	const slots = dropsItemsBySlot(list)
		.map(({ slot, items }) => ({ slot, ...dropsTickCount(items) }))
		.filter(slot => slot.total > 0);

	return (
		<div className="grid content-start gap-2" data-testid="drops-items">
			<div className="flex flex-wrap items-center gap-2">
				<h5 className="mr-auto mb-0">{t('items_title')}</h5>
				<Button size="sm" disabled={list.items.length === 0} onClick={onChoose}>
					{t('choose_items')}
				</Button>
			</div>
			{list.items.length === 0 ? (
				<p className="mb-0 text-sm text-gray-400">{t('empty')}</p>
			) : (
				<>
					<p className="mb-0 text-sm" data-testid="drops-items-ticked">
						{i18n.t('bulk_tab.drops.setup.items_ticked', { ticked, total })}
					</p>
					<Meter value={ticked} max={total} track="gray" />
					{hidden > 0 && (
						<p className="mb-0 text-sm text-gray-400" data-testid="drops-items-hidden">
							{i18n.t('bulk_tab.drops.setup.hidden_ticked', { count: hidden })}
						</p>
					)}
					<ul className="mb-0 flex list-none flex-wrap gap-2 p-0">
						{slots.map(slot => {
							const state = slot.ticked === 0 ? 'none' : slot.ticked === slot.total ? 'all' : 'some';
							const label = `${translateBulkSlotName(slot.slot)} ${slot.ticked}/${slot.total}`;
							return (
								<li key={slot.slot}>
									<div
										role="img"
										aria-label={label}
										data-state={state}
										data-testid="drops-items-slot"
										className="relative size-12 border border-border bg-cover data-[state=all]:border-primary data-[state=none]:opacity-40"
										style={{ backgroundImage: `url(${getEmptySlotIconUrl(bulkGroupItemSlots(slot.slot)[0])})` }}
										{...tooltipAnchorProps(tooltipId, label)}>
										<span className="absolute inset-x-0 bottom-0 bg-black py-px text-center text-2xs leading-none tabular-nums" aria-hidden>
											{slot.ticked}
										</span>
									</div>
								</li>
							);
						})}
					</ul>
					<Tooltip id={tooltipId} />
				</>
			)}
		</div>
	);
};

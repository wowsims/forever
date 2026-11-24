import i18n from '@i18n/config';
import { translateBulkSlotName } from '@i18n/localization';
import { Button } from '@ui-kit/Button';
import { Pill } from '@ui-kit/Pill';

import { type DropsItemList, dropsItemsBySlot, dropsTickCount } from '../../model/drops/items';

export interface DropsItemsCardViewProps {
	list: DropsItemList;
	/** The ticked, non-worn items the dialog's filters hide; they still run. */
	hidden: number;
	onChoose: () => void;
}

const t = (key: string) => i18n.t(`bulk_tab.drops.setup.${key}`);

export const DropsItemsCardView = ({ list, hidden, onChoose }: DropsItemsCardViewProps) => {
	const { ticked, total } = dropsTickCount(list.items);
	const slots = dropsItemsBySlot(list)
		.map(({ slot, items }) => ({ slot, ticked: dropsTickCount(items).ticked }))
		.filter(slot => slot.ticked > 0);

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
					{hidden > 0 && (
						<p className="mb-0 text-sm text-gray-400" data-testid="drops-items-hidden">
							{i18n.t('bulk_tab.drops.setup.hidden_ticked', { count: hidden })}
						</p>
					)}
					{slots.length > 0 && (
						<ul className="mb-0 flex list-none flex-wrap gap-1 p-0">
							{slots.map(({ slot, ticked: slotTicked }) => (
								<li key={slot}>
									<Pill size="xs" testId="drops-items-slot" className="gap-1">
										{translateBulkSlotName(slot)}
										<span className="tabular-nums">{slotTicked}</span>
									</Pill>
								</li>
							))}
						</ul>
					)}
				</>
			)}
		</div>
	);
};

import i18n from '@i18n/config';
import { SortDirection } from '@sim/constants/other';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';

import { ItemListSortBy, tabHasLevel } from '../../model/item_sort';
import { getTranslatedTabLabel, SelectorModalTabs } from '../../types';
import { columnHeaderLabel } from './utils';

export interface ItemListHeaderRowProps {
	label: SelectorModalTabs;
	/** The EP column; omitted for a list with none. `showValues` shows or hides it. */
	ep?: { showValues: boolean; tooltipId: string };
	sortBy: ItemListSortBy;
	sortDirection: SortDirection;
	onSort: (sortBy: ItemListSortBy) => void;
}

const sortedLabel = (sortBy: ItemListSortBy, sortDirection: SortDirection, column: ItemListSortBy) =>
	sortBy !== column
		? undefined
		: i18n.t(sortDirection === SortDirection.ASC ? 'gear_tab.gear_picker.sorted_ascending' : 'gear_tab.gear_picker.sorted_descending');

export const ItemListHeaderRow = ({ label, ep, sortBy, sortDirection, onSort }: ItemListHeaderRowProps) => {
	const isItemsTab = label === SelectorModalTabs.Items;
	const ilvlSortedLabel = sortedLabel(sortBy, sortDirection, ItemListSortBy.ILVL);
	const epSortedLabel = sortedLabel(sortBy, sortDirection, ItemListSortBy.EP);
	return (
		<div className="mr-4 flex justify-between gap-4 pr-2 text-md max-xl:mr-0" data-testid="selector-modal-list-labels">
			{tabHasLevel(label) && (
				<h6 className="w-12 select-none" data-testid="ilvl-label">
					<Button variant="unstyled" className="text-inherit [font:inherit] focus-visible:focus-ring" onClick={() => onSort(ItemListSortBy.ILVL)}>
						{i18n.t('gear_tab.gear_picker.table_headers.ilvl')}
						{ilvlSortedLabel && <span className="sr-only">, {ilvlSortedLabel}</span>}
					</Button>
				</h6>
			)}
			<h6 className="mr-2 flex-1" data-testid="item-label">
				{columnHeaderLabel(label, getTranslatedTabLabel)}
			</h6>
			{isItemsTab && (
				<h6 className="w-64" data-testid="source-label">
					{i18n.t('gear_tab.gear_picker.table_headers.source')}
				</h6>
			)}
			{ep && (
				<h6 className="float-right flex w-24 items-center select-none" data-testid="ep-label" style={{ display: ep.showValues ? undefined : 'none' }}>
					<Button
						variant="unstyled"
						className="flex items-center text-inherit [font:inherit] focus-visible:focus-ring"
						onClick={() => onSort(ItemListSortBy.EP)}>
						<span>{i18n.t('gear_tab.gear_picker.table_headers.ep')}</span>
						<Icon name="plus-minus" size="2xs" />
						{epSortedLabel && <span className="sr-only">, {epSortedLabel}</span>}
					</Button>
					<Button iconOnly aria-label={i18n.t('gear_tab.gear_picker.ep_tooltip')} className="ml-1" data-tooltip-id={ep.tooltipId}>
						<Icon name="circle-question" style="regular" size="lg" />
					</Button>
				</h6>
			)}
			<h6 className="w-8" data-testid="favorite-label" />
			{isItemsTab && <h6 className="w-8" data-testid="compare-label" />}
		</div>
	);
};

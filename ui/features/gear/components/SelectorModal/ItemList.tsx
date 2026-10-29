import { Class, ItemSlot } from '@generated/proto/common';
import { UIEnchant as Enchant, UIItem as Item } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { SortDirection } from '@sim/constants/other';
import { useSimHost } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { useSimStore } from '@sim/hooks/useSimStore';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { useUiStore } from '@sim/hooks/useUiStore';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { subscribeSimField } from '@sim/state/subscriptions';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { SearchField } from '@ui-kit/TextField';
import { Tooltip } from '@ui-kit/Tooltip';
import { VirtualList } from '@ui-kit/VirtualList';
import clsx from 'clsx';
import type { ReactNode } from 'react';
import { useCallback, useId, useMemo, useRef, useState } from 'react';

import { applyFavourite, isItemFavourited } from '../../model/favourites';
import { getItemIdByItemType } from '../../model/item_ids';
import { matchesSearch } from '../../model/item_search';
import { defaultSortBy, ItemListSortBy, sortItemIdxs, tabSortsByEP } from '../../model/item_sort';
import { type ItemDataFields, type ItemListType, SelectorModalTabs } from '../../types';
import { FiltersMenu } from '../FiltersMenu';
import { ItemListHeaderRow } from './ItemListHeaderRow';
import { ItemListRow } from './ItemListRow';
import { removeButtonLabel, type SelectorTab } from './utils';

const ROW_HEIGHT = 56;
const ROW_CLASS = clsx(
	'relative flex items-center gap-4 bg-gray-900 p-2',
	'hover:bg-gray-800 data-[stripe=even]:not-hover:bg-table-odd data-[stripe=odd]:not-hover:bg-table-even',
	'[&_[data-active]_.ui-selector-modal-list-item-icon]:outline-2 [&_[data-active]_.ui-selector-modal-list-item-icon]:outline-success',
);

type FilterFlag = 'oneHandedWeapons' | 'twoHandedWeapons' | 'showAllEnchants';

const phaseOptions = () => [
	{ value: 1, label: i18n.t('common.phases.1') },
	{ value: 2, label: i18n.t('common.phases.2') },
	{ value: 3, label: i18n.t('common.phases.3') },
	{ value: 4, label: i18n.t('common.phases.4') },
	{ value: 5, label: i18n.t('common.phases.5') },
];

export interface ItemListToggle {
	id: string;
	testId: string;
	label: string;
	value: boolean;
	onChange: (value: boolean) => void;
}

export interface ItemListToolbarViewProps {
	search: string;
	onSearchChange: (search: string) => void;
	/** The filters dialog. The filters button renders only beside one. */
	filtersMenu?: ReactNode;
	onOpenFilters?: () => void;
	phase: number;
	onPhaseChange: (phase: number) => void;
	toggles: ReadonlyArray<ItemListToggle>;
	/** The remove button renders only beside one. */
	remove?: { label: string; onClick: () => void };
}

export const ItemListToolbarView = ({
	search,
	onSearchChange,
	filtersMenu,
	onOpenFilters,
	phase,
	onPhaseChange,
	toggles,
	remove,
}: ItemListToolbarViewProps) => {
	return (
		<div className="ui-selector-modal-filters mb-(--modal-padding) flex items-center [&>*:not(:last-child)]:mr-2">
			<SearchField
				value={search}
				onChange={onSearchChange}
				placeholder={i18n.t('common.search')}
				inputClassName="max-w-48"
				inputTestId="selector-modal-search"
			/>
			{filtersMenu !== undefined && (
				<>
					<Button data-testid="selector-modal-filters-button" onClick={onOpenFilters}>
						{i18n.t('gear_tab.gear_picker.filters_button')}
					</Button>
					{filtersMenu}
				</>
			)}
			<div className="min-w-28" data-testid="selector-modal-phase-selector">
				<ChoiceField options={phaseOptions()} value={phase} onChange={onPhaseChange} />
			</div>
			{toggles.map(toggle => (
				<BooleanPicker
					key={toggle.id}
					id={toggle.id}
					testId={toggle.testId}
					label={toggle.label}
					layout="inline"
					value={toggle.value}
					onChange={toggle.onChange}
				/>
			))}
			{remove && (
				<Button variant="danger" className="-mt-2 mr-4 -mb-2 ml-auto" data-testid="selector-modal-remove-button" onClick={remove.onClick}>
					{remove.label}
				</Button>
			)}
		</div>
	);
};

const toggle = (name: string, label: string, value: boolean, onChange: (value: boolean) => void): ItemListToggle => ({
	id: `show-${name}-selector`,
	testId: `selector-modal-show-${name}`,
	label,
	value,
	onChange,
});

export interface ItemListProps {
	tab: SelectorTab;
	slot: ItemSlot;
	equippedItem: EquippedItem | null;
}

export const ItemList = ({ tab, slot, equippedItem }: ItemListProps) => {
	const host = useSimHost();
	const player = host.player;
	const sim = player.sim;
	const { label, itemData, computeEP, equippedToItem, onRemove } = tab;

	const [search, setSearch] = useState('');
	const [sortBy, setSortBy] = useState(() => defaultSortBy(label, slot));
	const [sortDirection, setSortDirection] = useState(SortDirection.DESC);
	const listRef = useRef<HTMLDivElement>(null);
	const [filtersOpen, setFiltersOpen] = useState(false);
	const tooltipId = useId();

	const filters = useStoreSubscribe(subscribeSimField(sim, 'filters'), () => sim.getFilters());
	const phase = useSimStore('phase');
	const race = usePlayerStore('race');
	const showEPValues = useUiStore('showEPValues');

	const isFavourited = useCallback((row: ItemDataFields<ItemListType>) => isItemFavourited(filters, label, row), [filters, label]);

	const itemsToDisplay = useMemo(() => {
		let idxs = itemData.map((_, index) => index);
		const currentEquippedItem = player.getEquippedItem(slot);

		if (label === SelectorModalTabs.Items) {
			idxs = player.filterItemData(idxs, index => itemData[index].item as unknown as Item, slot);
		} else if (label === SelectorModalTabs.Enchants) {
			idxs = player.filterEnchantData(idxs, index => itemData[index].item as unknown as Enchant, slot, currentEquippedItem);
		}

		idxs = idxs.filter(
			index =>
				itemData[index].phase <= phase &&
				matchesSearch(
					search,
					itemData[index],
					npcId => sim.db.getNpc(npcId)?.name,
					factionId => sim.db.getRepFaction(factionId)?.name,
				),
		);
		return sortItemIdxs(idxs, itemData, { sortBy, sortDirection, computeEP, isFavourited });
		// `filters` is named here even though the body never reads it: `player.filterItemData` and
		// `player.filterEnchantData` reach `sim.getFilters()` internally, so a filter change has to
		// rebuild this. The lint calls it unnecessary; deleting it leaves correctness resting on
		// `isFavourited` happening to close over `filters` too. `race` is named for the same reason:
		// the faction filter follows the race unless the user picked a faction.
		// oxlint-disable-next-line react-hooks/exhaustive-deps
	}, [itemData, player, sim, slot, label, phase, search, sortBy, sortDirection, computeEP, isFavourited, filters, race]);

	const showEPOptions = tabSortsByEP(label, slot);
	const equipped = equippedToItem(equippedItem);
	const equippedId = getItemIdByItemType(label, equipped);
	const equippedEP = showEPOptions && equipped !== null && equipped !== undefined ? computeEP(equipped) : null;

	const sort = (next: ItemListSortBy) => {
		setSortDirection(current => (sortBy === next ? 1 - current : SortDirection.DESC));
		setSortBy(next);
	};

	const toggleFavourite = (row: ItemDataFields<ItemListType>) => {
		const next = sim.getFilters();
		if (!applyFavourite(next, label, row, !isFavourited(row))) return;
		sim.setFilters(next);
	};

	const setFilterFlag = (field: FilterFlag) => (value: boolean) => {
		const next = sim.getFilters();
		next[field] = value;
		sim.setFilters(next);
	};

	const showWeaponOptions =
		label === SelectorModalTabs.Items &&
		(slot === ItemSlot.ItemSlotMainHand || (slot === ItemSlot.ItemSlotOffHand && player.getClass() === Class.ClassWarrior));

	const toggles = [
		showWeaponOptions &&
			toggle('1h-weapons', i18n.t('settings_tab.other.show_1h_weapons.label'), filters.oneHandedWeapons, setFilterFlag('oneHandedWeapons')),
		showWeaponOptions &&
			toggle('2h-weapons', i18n.t('settings_tab.other.show_2h_weapons.label'), filters.twoHandedWeapons, setFilterFlag('twoHandedWeapons')),
		label === SelectorModalTabs.Enchants &&
			toggle('all-enchants', i18n.t('settings_tab.other.show_all_enchants.label'), filters.showAllEnchants, setFilterFlag('showAllEnchants')),
		showEPOptions && toggle('ep-values', i18n.t('settings_tab.other.show_ep_values.label'), showEPValues, value => sim.setShowEPValues(value)),
	].filter((entry): entry is ItemListToggle => !!entry);

	return (
		<>
			<ItemListToolbarView
				search={search}
				onSearchChange={setSearch}
				filtersMenu={label === SelectorModalTabs.Items ? <FiltersMenu slot={slot} open={filtersOpen} onOpenChange={setFiltersOpen} /> : undefined}
				onOpenFilters={() => setFiltersOpen(true)}
				phase={phase}
				onPhaseChange={next => sim.setPhase(next)}
				toggles={toggles}
				remove={onRemove && { label: removeButtonLabel(label, key => i18n.t(key)), onClick: onRemove }}
			/>
			<ItemListHeaderRow
				label={label}
				ep={showEPOptions ? { showValues: showEPValues, tooltipId: `${tooltipId}-ep` } : undefined}
				sortBy={sortBy}
				sortDirection={sortDirection}
				onSort={sort}
			/>
			<div
				ref={listRef}
				className="mb-0 h-[65dvh] w-full overflow-x-hidden overflow-y-scroll p-0"
				data-testid="selector-modal-list"
				data-hide-ep={!showEPValues ? '' : undefined}
				tabIndex={0}>
				<VirtualList
					count={itemsToDisplay.length}
					rowHeight={ROW_HEIGHT}
					getScrollElement={() => listRef.current}
					rowClassName={() => ROW_CLASS}
					renderRow={index => {
						const row = itemData[itemsToDisplay[index]];
						return (
							<ItemListRow
								itemData={row}
								label={label}
								active={row.id === equippedId}
								ep={showEPOptions ? { ep: computeEP(row.item), equippedEP } : undefined}
								favourited={isFavourited(row)}
								favouriteTooltipId={`${tooltipId}-favourite`}
								compareTooltipId={`${tooltipId}-compare`}
								onEquip={() => row.onEquip(row.item)}
								onToggleFavourite={() => toggleFavourite(row)}
							/>
						);
					}}
				/>
			</div>
			{showEPOptions && <Tooltip id={`${tooltipId}-ep`} content={i18n.t('gear_tab.gear_picker.ep_tooltip')} />}
			<Tooltip
				id={`${tooltipId}-favourite`}
				render={({ activeAnchor }) =>
					(activeAnchor as HTMLElement | null)?.dataset.favourited === 'true'
						? i18n.t('gear_tab.gear_picker.row_actions.remove_favorite')
						: i18n.t('gear_tab.gear_picker.row_actions.add_favorite')
				}
			/>
			<Tooltip
				id={`${tooltipId}-compare`}
				render={({ activeAnchor }) =>
					(activeAnchor as HTMLElement | null)?.dataset.inBatch === 'true'
						? i18n.t('gear_tab.gear_picker.row_actions.remove_from_batch')
						: i18n.t('gear_tab.gear_picker.row_actions.add_to_batch')
				}
			/>
		</>
	);
};

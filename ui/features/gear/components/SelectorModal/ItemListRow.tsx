import { addBulkItem, hasBulkItem, removeBulkItem } from '@features/bulk/model/gear/items';
import { type ItemQuality, ItemSpec } from '@generated/proto/common';
import type { UIItem as Item } from '@generated/proto/ui';
import { useSimHost } from '@sim/context/SimHostContext';
import { useActionId } from '@sim/hooks/useActionId';
import { isIndividualSimHost } from '@sim/sim_host';
import { FakeLink } from '@ui-kit/FakeLink';
import { itemQualityClassName } from '@ui-kit/utils/css';
import { WowheadAnchor } from '@ui-kit/WowheadAnchor';
import clsx from 'clsx';
import { type ReactNode, useMemo } from 'react';
import { useStore } from 'zustand';

import { trackEvent } from '../../../../tracking/utils';
import { tabHasLevel } from '../../model/item_sort';
import type { ItemData, ItemListType } from '../../types';
import { SelectorModalTabs } from '../../types';
import { ItemNoticeIcon } from '../GearPicker/ItemNoticeIcon';
import { NameDescriptionLabel } from '../ItemCell';
import { EpCell, type EpCellProps } from './EpCell';
import { ItemSource } from './ItemSource';
import { RowActions, type RowActionsProps } from './RowActions';

export interface ItemListRowViewProps extends RowActionsProps {
	name: ReactNode;
	quality: ItemQuality;
	nameDescription: string;
	iconUrl: string;
	href: string;
	/** The level column: an item's ilvl on the items tab, an enchant's lvl on the enchants tab; 0 leaves the cell empty. */
	ilvl?: number;
	active: boolean;
	onEquip: () => void;
	notice?: ReactNode;
	/** The source column; the items tab alone has one. */
	source?: ReactNode;
	/** Omitted for a slot the list computes no EP for. */
	ep?: EpCellProps;
}

export const ItemListRowView = ({
	name,
	quality,
	nameDescription,
	iconUrl,
	href,
	ilvl,
	active,
	onEquip,
	notice,
	source,
	ep,
	...actions
}: ItemListRowViewProps) => (
	<>
		{ilvl !== undefined && (
			<div className="w-12 text-center" data-testid="selector-modal-list-item-ilvl-container">
				{ilvl || null}
			</div>
		)}
		<div className="flex flex-1 items-center gap-1" data-testid="selector-modal-list-label-cell" data-active={active ? '' : undefined}>
			<WowheadAnchor className="ui-selector-modal-list-item-link" testId="selector-modal-list-item-link" href={href || undefined} onActivate={onEquip}>
				<img className="ui-selector-modal-list-item-icon" data-testid="selector-modal-list-item-icon" src={iconUrl || undefined} alt="" />
				<span
					className={clsx('ui-selector-modal-list-item-name', itemQualityClassName(quality))}
					data-testid="selector-modal-list-item-name"
					data-quality={quality}>
					{name}
					{!!nameDescription && <NameDescriptionLabel nameDescription={nameDescription} />}
				</span>
			</WowheadAnchor>
			{notice}
		</div>
		{source !== undefined && (
			<div className="w-64" data-testid="selector-modal-list-item-source-container">
				{source}
			</div>
		)}
		{ep && <EpCell {...ep} />}
		<RowActions {...actions} />
		<FakeLink onActivate={onEquip} testId="selector-modal-list-item-row-link" />
	</>
);

export interface ItemListRowProps {
	itemData: ItemData<ItemListType, ReactNode>;
	label: SelectorModalTabs;
	active: boolean;
	ep?: EpCellProps;
	favourited: boolean;
	favouriteTooltipId: string;
	compareTooltipId: string;
	onEquip: () => void;
	onToggleFavourite: () => void;
}

export const ItemListRow = ({
	itemData,
	label,
	active,
	ep,
	favourited,
	favouriteTooltipId,
	compareTooltipId,
	onEquip,
	onToggleFavourite,
}: ItemListRowProps) => {
	const host = useSimHost();
	const { iconUrl, href } = useActionId(itemData.actionId);
	const isItemsTab = label === SelectorModalTabs.Items;
	const isEnchantsTab = label === SelectorModalTabs.Enchants;
	const level = tabHasLevel(label) ? itemData.ilvl || (isItemsTab ? (itemData.item as unknown as Item).ilvl : 0) : undefined;

	const batchPlayer = isIndividualSimHost(host) ? host.player : null;
	const batchSpec = useMemo(() => ItemSpec.create({ id: itemData.id }), [itemData.id]);
	// Read through zustand rather than `useStoreSubscribe`: the answer is keyed by a prop, and that
	// hook re-reads on a notification only, so a recycled row would keep the previous item's flag.
	const inBatch = useStore(host.sim.store, () => !!batchPlayer && hasBulkItem(batchPlayer, batchSpec));

	const toggleCompare = () => {
		const present = !!batchPlayer && hasBulkItem(batchPlayer, batchSpec);
		if (batchPlayer) (present ? removeBulkItem : addBulkItem)(batchPlayer, batchSpec);
		trackEvent({ action: 'click', category: 'batch', label: present ? 'remove-item' : 'add-item' });
	};

	return (
		<ItemListRowView
			name={itemData.name}
			quality={itemData.quality}
			nameDescription={itemData.nameDescription}
			iconUrl={iconUrl}
			href={href}
			ilvl={level}
			active={active}
			onEquip={onEquip}
			notice={<ItemNoticeIcon itemId={isEnchantsTab ? undefined : itemData.id} enchantId={isEnchantsTab ? itemData.id : undefined} />}
			source={isItemsTab ? <ItemSource item={itemData.item as unknown as Item} sim={host.sim} /> : undefined}
			ep={ep}
			favourited={favourited}
			favouriteTooltipId={favouriteTooltipId}
			onToggleFavourite={onToggleFavourite}
			compare={isItemsTab ? { inBatch, tooltipId: compareTooltipId, onToggle: toggleCompare } : undefined}
		/>
	);
};

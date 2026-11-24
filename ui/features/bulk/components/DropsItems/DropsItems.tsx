import { getEmptySlotIconUrl } from '@features/gear/model/empty_slot_icons';
import i18n from '@i18n/config';
import { translateBulkSlotName } from '@i18n/localization';
import { BULK_SIM_ITEM_SLOT_TO_ITEM_SLOT_PAIRS, BULK_SIM_ITEM_SLOT_TO_SINGLE_ITEM_SLOT, type BulkSimItemSlot } from '@sim/bulk/utils';
import { useActionId } from '@sim/hooks/useActionId';
import { ActionId } from '@sim/proto/action_id';
import { Accordion, AccordionItem } from '@ui-kit/Accordion';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { Field } from '@ui-kit/Field';
import { IconSwatch } from '@ui-kit/IconSwatch';
import { ItemCellIcon } from '@ui-kit/ItemCell';
import { Pill } from '@ui-kit/Pill';
import { ToggleGroup } from '@ui-kit/ToggleGroup';
import { Tooltip } from '@ui-kit/Tooltip';
import { TooltipAnchor } from '@ui-kit/TooltipAnchor';
import { itemQualityClassName } from '@ui-kit/utils/css';
import { WowheadIcon } from '@ui-kit/WowheadIcon';
import clsx from 'clsx';
import { useId, useMemo, useState } from 'react';

import { type DropsItem, type DropsItemList, dropsItemsBySlot, dropsItemsBySource } from '../../model/drops/items';
import { type DropsNames, dropsSourceLabel, type DropsSourceNode } from '../../model/drops/sources';
import { baseIlvl } from '../../model/gear/search';

export interface DropsItemsViewProps {
	list: DropsItemList;
	grouping: 'slot' | 'source';
	onGroupingChange: (grouping: 'slot' | 'source') => void;
	/** The Sources tree the list was picked from, whose leaf order the source grouping follows. */
	tree: readonly DropsSourceNode[];
	names: DropsNames;
	/** Called with the listed, non-worn ids of Include all / Exclude all, or one row's id. */
	onIncludedChange: (ids: readonly number[], included: boolean) => void;
}

const ITEM_ROW_CLASS = clsx('hover:bg-gray-800 data-[stripe=even]:not-hover:bg-table-odd data-[stripe=odd]:not-hover:bg-table-even');

const t = (key: string) => i18n.t(`bulk_tab.drops.setup.${key}`);

// `source` is the source key of a source group, which its rows leave out of their own sources.
interface ItemGroup {
	key: string;
	source?: string;
	title: string;
	/** The empty-slot art of a slot group, from the first gear slot of its bulk slot. */
	iconUrl?: string;
	detail?: string;
	items: DropsItem[];
}

const slotIconUrl = (slot: BulkSimItemSlot): string | undefined => {
	const gearSlot = BULK_SIM_ITEM_SLOT_TO_SINGLE_ITEM_SLOT.get(slot) ?? BULK_SIM_ITEM_SLOT_TO_ITEM_SLOT_PAIRS.get(slot)?.[0];
	return gearSlot === undefined ? undefined : getEmptySlotIconUrl(gearSlot);
};

const itemGroups = (list: DropsItemList, grouping: DropsItemsViewProps['grouping'], tree: readonly DropsSourceNode[], names: DropsNames): ItemGroup[] =>
	grouping === 'slot'
		? dropsItemsBySlot(list).map(({ slot, items }) => ({ key: `slot:${slot}`, title: translateBulkSlotName(slot), iconUrl: slotIconUrl(slot), items }))
		: dropsItemsBySource(list, tree, names).map(({ key, label, items }) => ({ key, source: key, title: label.label, detail: label.detail, items }));

interface ItemRowProps {
	entry: DropsItem;
	index: number;
	group?: string;
	names: DropsNames;
	onIncludedChange: DropsItemsViewProps['onIncludedChange'];
}

// Box, icon and name share one field, so the name ticks the box and a worn row dims once, through the field's disabled look. The name
// stretches over the free space, and a press on the sources or the tags ticks the box too; the box stays the keyboard's way in.
// The other sources open in a kit tooltip, which a tap opens on a touch screen.
const ItemRow = ({ entry, index, group, names, onIncludedChange }: ItemRowProps) => {
	const { item, worn, included, tags, sourceKeys } = entry;
	const { iconUrl, href } = useActionId(ActionId.fromItem(item));
	const id = useId();
	const [first, ...more] = sourceKeys.filter(key => key !== group).map(key => dropsSourceLabel(key, names).label);

	return (
		<div
			className={clsx('flex h-10 cursor-pointer items-center gap-2 px-1', ITEM_ROW_CLASS)}
			data-stripe={index % 2 === 0 ? 'even' : 'odd'}
			data-testid="drops-item-row"
			data-item-id={item.id}
			data-worn={worn ? '' : undefined}
			onClick={event => {
				if (!worn && !(event.target as Element).closest('label, input, button, a, [role="checkbox"]')) onIncludedChange([item.id], !included);
			}}>
			<Field.Root id={id} layout="inline" size="xs" disabled={worn} className="ui-boolean-picker min-w-0 grow">
				<BooleanPicker id={id} value={included && !worn} disabled={worn} onChange={checked => onIncludedChange([item.id], checked)} />
				<ItemCellIcon ilvl={baseIlvl(item)} size="sm">
					<WowheadIcon className="size-full" iconUrl={iconUrl} href={href || undefined} navigate={false} tabIndex={-1} />
				</ItemCellIcon>
				<Field.Label className={clsx('min-w-0 grow cursor-pointer truncate text-base', itemQualityClassName(item.quality))}>{item.name}</Field.Label>
			</Field.Root>
			{worn ? (
				<span className="max-w-2/5 min-w-0 truncate text-xs text-gray-400">{i18n.t('bulk_tab.drops.items.worn')}</span>
			) : (
				first && (
					<span className="flex max-w-2/5 min-w-0 items-center gap-1 text-xs text-gray-400">
						<span className="min-w-0 truncate">{first}</span>
						{more.length > 0 && (
							<>
								<TooltipAnchor tooltipId={`${id}-sources`} className="shrink-0" testId="drops-item-more-sources">
									+{more.length}
								</TooltipAnchor>
								<Tooltip
									id={`${id}-sources`}
									content={
										<ul className="mb-0 list-none p-0">
											{more.map(label => (
												<li key={label}>{label}</li>
											))}
										</ul>
									}
								/>
							</>
						)}
					</span>
				)
			)}
			{tags.map(tag => (
				<Pill key={tag} size="xs" rounded="xs" testId="drops-item-tag" className="shrink-0">
					{i18n.t(`bulk_tab.drops.items.tags.${tag}`)}
				</Pill>
			))}
		</div>
	);
};

interface ItemGroupSectionProps {
	group: ItemGroup;
	names: DropsNames;
	onIncludedChange: DropsItemsViewProps['onIncludedChange'];
}

// The badge counts the included non-worn rows and hides at none, as the enchant preset dialog's does. Worn rows are left out of it and of the ids the box sends, as Include all and Exclude all leave them out.
// A partly ticked box passes `value={false}` with `indeterminate`, so pressing it reports `true` and ticks the rest.
// A section's rows mount only while it is open, so the list holds as many as the open sections list.
const ItemGroupSection = ({ group, names, onIncludedChange }: ItemGroupSectionProps) => {
	const id = useId();
	const pickable = group.items.filter(entry => !entry.worn);
	const included = pickable.filter(entry => entry.included).length;
	const ids = pickable.map(entry => entry.item.id);

	return (
		<AccordionItem
			value={group.key}
			testId="drops-group"
			panelClassName="px-1"
			start={
				<Field.Root id={id} layout="inline" size="xs" disabled={ids.length === 0} className="ui-boolean-picker ms-2 grow-0">
					<BooleanPicker
						id={id}
						value={ids.length > 0 && included === ids.length}
						indeterminate={included > 0 && included < ids.length}
						disabled={ids.length === 0}
						onChange={checked => onIncludedChange(ids, checked)}
					/>
					<Field.Label className="sr-only">{group.title}</Field.Label>
				</Field.Root>
			}
			// A source section has no slot icon, so the row keeps the icon's height to pad like a slot section.
			title={
				<span className="flex min-h-8 min-w-0 grow items-center gap-2">
					{group.iconUrl && <IconSwatch icon={{ iconUrl: group.iconUrl, name: group.title }} className="size-8 min-w-8" aria-hidden />}
					<span className="truncate">{group.title}</span>
					{group.detail && <span className="min-w-0 truncate text-xs font-normal text-gray-400">{group.detail}</span>}
					{included > 0 && (
						<Pill size="xs" className="min-w-[1lh] shrink-0 justify-center" testId="drops-group-count">
							{included}
						</Pill>
					)}
				</span>
			}>
			{group.items.map((entry, index) => (
				<ItemRow key={entry.item.id} entry={entry} index={index} group={group.source} names={names} onIncludedChange={onIncludedChange} />
			))}
		</AccordionItem>
	);
};

export const DropsItemsView = ({ list, grouping, onGroupingChange, tree, names, onIncludedChange }: DropsItemsViewProps) => {
	const groups = useMemo(() => itemGroups(list, grouping, tree, names), [list, grouping, tree, names]);
	const [open, setOpen] = useState<string[]>([]);
	const pickable = list.items.filter(entry => !entry.worn).map(entry => entry.item.id);

	return (
		<div className="grid content-start gap-4" data-testid="drops-items">
			<div className="flex flex-wrap items-center gap-2">
				<h5 className="mr-auto mb-0">{t('items_title')}</h5>
				{groups.length > 0 && (
					<>
						<ToggleGroup
							size="sm"
							tone="neutral"
							aria-label={t('items_title')}
							value={grouping}
							options={[
								{ value: 'slot', label: t('by_slot') },
								{ value: 'source', label: t('by_source') },
							]}
							onChange={next => {
								setOpen([]);
								onGroupingChange(next);
							}}
						/>
						<div className="flex items-center gap-2">
							<Button variant="secondary" size="sm" disabled={pickable.length === 0} onClick={() => onIncludedChange(pickable, true)}>
								{t('include_all')}
							</Button>
							<Button variant="secondary" size="sm" disabled={pickable.length === 0} onClick={() => onIncludedChange(pickable, false)}>
								{t('exclude_all')}
							</Button>
						</div>
					</>
				)}
			</div>
			{groups.length === 0 ? (
				<p className="mb-0 text-sm text-gray-400">{t('empty')}</p>
			) : (
				<div className="max-h-[60dvh] overflow-x-hidden overflow-y-auto" data-testid="drops-item-list" tabIndex={0}>
					<Accordion value={open} onValueChange={setOpen} variant="boxed">
						{groups.map(group => (
							<ItemGroupSection key={group.key} group={group} names={names} onIncludedChange={onIncludedChange} />
						))}
					</Accordion>
				</div>
			)}
		</div>
	);
};

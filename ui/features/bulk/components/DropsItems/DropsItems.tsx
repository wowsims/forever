import i18n from '@i18n/config';
import { translateBulkSlotName } from '@i18n/localization';
import { useActionId } from '@sim/hooks/useActionId';
import { ActionId } from '@sim/proto/action_id';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { Field } from '@ui-kit/Field';
import { ItemCellIcon } from '@ui-kit/ItemCell';
import { Pill } from '@ui-kit/Pill';
import { ToggleGroup } from '@ui-kit/ToggleGroup';
import { Tooltip } from '@ui-kit/Tooltip';
import { TooltipAnchor } from '@ui-kit/TooltipAnchor';
import { itemQualityClassName } from '@ui-kit/utils/css';
import { VirtualList } from '@ui-kit/VirtualList';
import { WowheadIcon } from '@ui-kit/WowheadIcon';
import clsx from 'clsx';
import { useId, useMemo, useRef } from 'react';

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

const ROW_HEIGHT = 40;
const ITEM_ROW_CLASS = clsx('hover:bg-gray-800 data-[stripe=even]:not-hover:bg-table-odd data-[stripe=odd]:not-hover:bg-table-even');

const t = (key: string) => i18n.t(`bulk_tab.drops.setup.${key}`);

// `group` is the source key of the group a row sits under, which the row then leaves out of its own sources.
type ListRow = { kind: 'header'; key: string; title: string; detail?: string } | { kind: 'item'; key: string; group?: string; entry: DropsItem };

const listRows = (list: DropsItemList, grouping: DropsItemsViewProps['grouping'], tree: readonly DropsSourceNode[], names: DropsNames): ListRow[] => {
	const groups =
		grouping === 'slot'
			? dropsItemsBySlot(list).map(({ slot, items }) => ({
					key: `slot:${slot}`,
					source: undefined,
					title: translateBulkSlotName(slot),
					detail: undefined,
					items,
				}))
			: dropsItemsBySource(list, tree, names).map(({ key, label, items }) => ({ key, source: key, title: label.label, detail: label.detail, items }));
	return groups.flatMap(group => [
		{ kind: 'header' as const, key: group.key, title: group.title, detail: group.detail },
		...group.items.map(entry => ({ kind: 'item' as const, key: `${group.key}/${entry.item.id}`, group: group.source, entry })),
	]);
};

interface ItemRowProps {
	entry: DropsItem;
	group?: string;
	names: DropsNames;
	onIncludedChange: DropsItemsViewProps['onIncludedChange'];
}

// Box, icon and name share one field, so the name ticks the box and a worn row dims once, through the field's disabled look.
// The other sources open in a kit tooltip, which a tap opens on a touch screen.
const ItemRow = ({ entry, group, names, onIncludedChange }: ItemRowProps) => {
	const { item, worn, included, tags, sourceKeys } = entry;
	const { iconUrl, href } = useActionId(ActionId.fromItem(item));
	const id = useId();
	const [first, ...more] = sourceKeys.filter(key => key !== group).map(key => dropsSourceLabel(key, names).label);

	return (
		<div className="flex h-full items-center gap-2 px-1" data-testid="drops-item-row" data-item-id={item.id} data-worn={worn ? '' : undefined}>
			<Field.Root id={id} layout="inline" size="xs" disabled={worn} className="ui-boolean-picker min-w-0 grow">
				<BooleanPicker id={id} value={included && !worn} disabled={worn} onChange={checked => onIncludedChange([item.id], checked)} />
				<ItemCellIcon ilvl={baseIlvl(item)} size="sm">
					<WowheadIcon className="size-full" iconUrl={iconUrl} href={href || undefined} navigate={false} tabIndex={-1} />
				</ItemCellIcon>
				<Field.Label className={clsx('min-w-0 truncate text-base', itemQualityClassName(item.quality))}>{item.name}</Field.Label>
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

export const DropsItemsView = ({ list, grouping, onGroupingChange, tree, names, onIncludedChange }: DropsItemsViewProps) => {
	const scroller = useRef<HTMLDivElement>(null);
	const rows = useMemo(() => listRows(list, grouping, tree, names), [list, grouping, tree, names]);
	const pickable = list.items.filter(entry => !entry.worn).map(entry => entry.item.id);

	return (
		<div className="grid content-start gap-4" data-testid="drops-items">
			<div className="flex flex-wrap items-center gap-2">
				<h5 className="mr-auto mb-0">{t('items_title')}</h5>
				{rows.length > 0 && (
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
							onChange={onGroupingChange}
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
			{rows.length === 0 ? (
				<p className="mb-0 text-sm text-gray-400">{t('empty')}</p>
			) : (
				<div ref={scroller} className="max-h-[60dvh] overflow-x-hidden overflow-y-auto" data-testid="drops-item-list" tabIndex={0}>
					<VirtualList
						count={rows.length}
						rowHeight={ROW_HEIGHT}
						getScrollElement={() => scroller.current}
						rowClassName={index => (rows[index].kind === 'item' ? ITEM_ROW_CLASS : undefined)}
						renderRow={index => {
							const row = rows[index];
							return row.kind === 'header' ? (
								<div
									className="flex h-full items-end gap-2 border-b border-border pb-1 text-xs font-bold text-gray-400"
									data-testid="drops-group-header">
									<span>{row.title}</span>
									{row.detail && <span className="font-normal">{row.detail}</span>}
								</div>
							) : (
								<ItemRow entry={row.entry} group={row.group} names={names} onIncludedChange={onIncludedChange} />
							);
						}}
					/>
				</div>
			)}
		</div>
	);
};

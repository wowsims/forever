import { FiltersButton, type FiltersMenuRow, FiltersMenuView, type FiltersMenuViewProps } from '@features/gear/components/FiltersMenu';
import { railStep } from '@features/gear/model/rail';
import type { DatabaseFilters } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { translateBulkSlotName } from '@i18n/localization';
import { useActionId } from '@sim/hooks/useActionId';
import { ActionId } from '@sim/proto/action_id';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { Dialog } from '@ui-kit/Dialog';
import { Field } from '@ui-kit/Field';
import { ItemCellIcon } from '@ui-kit/ItemCell';
import { Pill } from '@ui-kit/Pill';
import { SearchField } from '@ui-kit/TextField';
import { ToggleGroup } from '@ui-kit/ToggleGroup';
import { Tooltip } from '@ui-kit/Tooltip';
import { TooltipAnchor } from '@ui-kit/TooltipAnchor';
import { itemQualityClassName } from '@ui-kit/utils/css';
import { VirtualList } from '@ui-kit/VirtualList';
import { WowheadIcon } from '@ui-kit/WowheadIcon';
import clsx from 'clsx';
import { type KeyboardEvent, useId, useMemo, useRef, useState } from 'react';

import { dropsMatchesSearch, dropsPassesFilters } from '../../model/drops/filters';
import { type DropsItem, type DropsItemList, dropsItemsBySlot, dropsItemsBySource, dropsTickCount } from '../../model/drops/items';
import { type DropsNames, dropsSourceLabel, type DropsSourceNode } from '../../model/drops/sources';
import { baseIlvl } from '../../model/gear/search';

export type DropsGrouping = 'slot' | 'source';

export interface DropsItemsDialogViewProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	list: DropsItemList;
	grouping: DropsGrouping;
	onGroupingChange: (grouping: DropsGrouping) => void;
	/** The Sources tree the list was picked from, whose leaf order the source grouping follows. */
	tree: readonly DropsSourceNode[];
	names: DropsNames;
	/** Drops' own filters, which only change what the dialog shows. */
	filters: DatabaseFilters;
	activeFilters: number;
	filtersMenu: Omit<FiltersMenuViewProps, 'open' | 'onOpenChange' | 'slot' | 'rows' | 'filters'>;
	/** Called with the shown, non-worn ids of Include shown or Exclude shown, or one row's id. */
	onIncludedChange: (ids: readonly number[], included: boolean) => void;
}

// The rows a list of every slot can use: no source, raid, faction or required level row, which the Sources tree and the
// level 45 cutoff decide.
const FILTER_ROWS: ReadonlyArray<FiltersMenuRow> = ['ilvl', 'stats', 'bonus', 'armor', 'weapons', 'ranged'];

const ALL = 'all';
// Below md a row puts its sources under the name; both lines fit the same height.
const ROW_HEIGHT = 40;
const ROW_CLASS = clsx('hover:bg-gray-800 data-[stripe=even]:not-hover:bg-table-odd data-[stripe=odd]:not-hover:bg-table-even');

const t = (key: string) => i18n.t(`bulk_tab.drops.setup.${key}`);

// `source` is the source key of a source group, which its rows leave out of their own sources.
interface RailGroup {
	key: string;
	source?: string;
	title: string;
	detail?: string;
	items: readonly DropsItem[];
}

// "All" lists the list itself, which holds each item once whatever the grouping.
const railGroups = (list: DropsItemList, grouping: DropsGrouping, tree: readonly DropsSourceNode[], names: DropsNames): RailGroup[] => [
	{ key: ALL, title: t('all'), items: list.items },
	...(grouping === 'slot'
		? dropsItemsBySlot(list).map(({ slot, items }) => ({ key: `slot:${slot}`, title: translateBulkSlotName(slot), items }))
		: dropsItemsBySource(list, tree, names).map(({ key, label, items }) => ({ key, source: key, title: label.label, detail: label.detail, items }))),
];

const otherSources = (entry: DropsItem, source: string | undefined, names: DropsNames) =>
	entry.sourceKeys.filter(key => key !== source).map(key => dropsSourceLabel(key, names).label);

interface ItemRowProps {
	entry: DropsItem;
	sources: readonly string[];
	sourcesTooltipId: string;
	onIncludedChange: DropsItemsDialogViewProps['onIncludedChange'];
}

// Box, icon and name share one field, so the name ticks the box and a worn row dims once, through the field's disabled look. The name
// stretches over the free space, and a press on the sources or the tags ticks the box too; the box stays the keyboard's way in.
// The sources sit beside the name from md up and under it below. The other sources open in the dialog's one tooltip, which a tap
// opens on a touch screen.
const ItemRow = ({ entry, sources, sourcesTooltipId, onIncludedChange }: ItemRowProps) => {
	const { item, worn, included, tags } = entry;
	const { iconUrl, href } = useActionId(ActionId.fromItem(item));
	const id = useId();
	const [first, ...more] = sources;

	return (
		<div
			className="flex h-full cursor-pointer items-center gap-2 px-1"
			data-testid="drops-item-row"
			data-item-id={item.id}
			data-worn={worn ? '' : undefined}
			onClick={event => {
				if (!worn && !(event.target as Element).closest('label, input, button, a, [role="checkbox"]')) onIncludedChange([item.id], !included);
			}}>
			<Field.Root
				id={id}
				layout="inline"
				size="xs"
				disabled={worn}
				className="ui-boolean-picker min-w-0 grow max-md:grid max-md:grid-cols-[auto_minmax(0,1fr)] max-md:content-center max-md:gap-y-0">
				<span className="flex shrink-0 items-center gap-2 max-md:row-span-2">
					<BooleanPicker id={id} value={included && !worn} disabled={worn} onChange={checked => onIncludedChange([item.id], checked)} />
					<ItemCellIcon ilvl={baseIlvl(item)} size="sm">
						<WowheadIcon className="size-full" iconUrl={iconUrl} href={href || undefined} navigate={false} tabIndex={-1} />
					</ItemCellIcon>
				</span>
				<Field.Label className={clsx('min-w-0 grow cursor-pointer truncate text-base', itemQualityClassName(item.quality))}>{item.name}</Field.Label>
				{!worn && first && (
					<span className="flex min-w-0 items-center gap-1 text-xs text-gray-400 md:max-w-2/5">
						<span className="min-w-0 truncate">{first}</span>
						{more.length > 0 && (
							<TooltipAnchor tooltipId={sourcesTooltipId} className="shrink-0" testId="drops-item-more-sources">
								+{more.length}
							</TooltipAnchor>
						)}
					</span>
				)}
			</Field.Root>
			{worn && <span className="max-w-2/5 min-w-0 truncate text-xs text-gray-400">{i18n.t('bulk_tab.drops.items.worn')}</span>}
			{tags.map(tag => (
				<Pill key={tag} size="xs" rounded="xs" testId="drops-item-tag" className="shrink-0">
					{i18n.t(`bulk_tab.drops.items.tags.${tag}`)}
				</Pill>
			))}
		</div>
	);
};

// Search, the filters and the rail pick live here and reach nothing that runs. The rail counts the list before them, so
// its numbers match the card's. A pick whose group leaves the rail reads as All. Below md the rail is one row of chips
// that scrolls sideways above the list, and the list fills the rest of the fullscreen dialog.
export const DropsItemsDialogView = ({
	open,
	onOpenChange,
	list,
	grouping,
	onGroupingChange,
	tree,
	names,
	filters,
	activeFilters,
	filtersMenu,
	onIncludedChange,
}: DropsItemsDialogViewProps) => {
	const [search, setSearch] = useState('');
	const [pick, setPick] = useState(ALL);
	const [filtersOpen, setFiltersOpen] = useState(false);
	const listRef = useRef<HTMLDivElement>(null);
	const sourcesTooltipId = useId();

	const groups = useMemo(() => railGroups(list, grouping, tree, names), [list, grouping, tree, names]);
	const picked = groups.find(group => group.key === pick) ?? groups[0];
	const shown = useMemo(
		() => picked.items.filter(entry => dropsPassesFilters(entry.item, filters) && dropsMatchesSearch(entry, search, names)),
		[picked, filters, search, names],
	);
	const shownIds = shown.filter(entry => !entry.worn).map(entry => entry.item.id);
	const sourcesOf = (entry: DropsItem) => otherSources(entry, picked.source, names);

	// On the popup, through `Dialog`'s own prop, as the gear picker's rail does. The filters menu is a dialog of its own.
	const onRailKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
		if (!open || filtersOpen) return;
		const next = railStep(
			groups.map(group => group.key),
			picked.key,
			event.key,
		);
		if (next === null) return;
		event.preventDefault();
		setPick(next);
	};

	return (
		<Dialog
			open={open}
			onOpenChange={onOpenChange}
			size="full"
			bodyClassName="min-h-0"
			title={t('items_title')}
			onKeyDown={onRailKeyDown}
			testId="drops-items-dialog">
			<div className="flex flex-wrap items-center gap-2">
				<SearchField
					value={search}
					onChange={setSearch}
					placeholder={i18n.t('common.search')}
					inputClassName="max-w-48"
					inputTestId="drops-items-search"
				/>
				<FiltersButton active={activeFilters} onClick={() => setFiltersOpen(true)} testId="drops-items-filters-button" />
				<ToggleGroup
					size="sm"
					tone="neutral"
					className="ms-auto"
					aria-label={t('items_title')}
					value={grouping}
					options={[
						{ value: 'slot', label: t('by_slot') },
						{ value: 'source', label: t('by_source') },
					]}
					onChange={next => {
						setPick(ALL);
						onGroupingChange(next);
					}}
				/>
			</div>
			<div className="flex min-h-0 flex-1 gap-(--modal-padding) max-md:flex-col max-md:gap-2">
				<nav
					aria-label={t(grouping === 'slot' ? 'by_slot' : 'by_source')}
					className="shrink-0 border-border max-md:-mx-(--modal-padding) max-md:flex max-md:scrollbar-none max-md:overflow-x-auto max-md:border-b max-md:px-(--modal-padding) md:w-52 md:overflow-y-auto md:border-e"
					data-testid="drops-items-rail">
					{groups.map(group => {
						const { ticked, total } = dropsTickCount(group.items);
						const active = group.key === picked.key;
						return (
							<Button
								key={group.key}
								variant={null}
								aria-current={active ? 'true' : undefined}
								data-active={active ? '' : undefined}
								className="flex w-full items-center gap-2 border-0 border-e-2 border-transparent px-2 py-1 text-left text-sm text-gray-400 hover:bg-gray-800 data-active:border-primary data-active:bg-gray-800 data-active:text-white max-md:w-auto max-md:shrink-0 max-md:gap-0 max-md:border-0 max-md:border-b-2 max-md:py-2 max-md:font-bold max-md:text-link max-md:hover:bg-transparent max-md:hover:text-link-hover max-md:data-active:bg-transparent max-md:data-active:text-white"
								onClick={() => setPick(group.key)}>
								<span className="grid min-w-0 grow">
									<span className="truncate">{group.title}</span>
									{group.detail && <span className="truncate text-xs">{group.detail}</span>}
								</span>
								<span className="shrink-0 tabular-nums max-md:-mt-px max-md:self-start max-md:ps-px max-md:text-xs max-md:leading-none max-md:font-normal">
									{ticked}/{total}
								</span>
							</Button>
						);
					})}
				</nav>
				<div className="flex min-h-0 min-w-0 flex-1 flex-col gap-2">
					<div className="flex flex-wrap items-center gap-2">
						<Button variant="secondary" size="sm" disabled={shownIds.length === 0} onClick={() => onIncludedChange(shownIds, true)}>
							{t('include_shown')}
						</Button>
						<Button variant="secondary" size="sm" disabled={shownIds.length === 0} onClick={() => onIncludedChange(shownIds, false)}>
							{t('exclude_shown')}
						</Button>
					</div>
					{shown.length === 0 ? (
						<p className="mb-0 text-sm text-gray-400">{t('no_match')}</p>
					) : (
						<div ref={listRef} className="min-h-0 flex-1 overflow-x-hidden overflow-y-auto" data-testid="drops-item-list" tabIndex={0}>
							<VirtualList
								count={shown.length}
								rowHeight={ROW_HEIGHT}
								getScrollElement={() => listRef.current}
								rowClassName={() => ROW_CLASS}
								renderRow={index => (
									<ItemRow
										entry={shown[index]}
										sources={sourcesOf(shown[index])}
										sourcesTooltipId={sourcesTooltipId}
										onIncludedChange={onIncludedChange}
									/>
								)}
							/>
						</div>
					)}
				</div>
			</div>
			<Tooltip
				id={sourcesTooltipId}
				render={({ activeAnchor }) => {
					const id = Number(activeAnchor?.closest('[data-item-id]')?.getAttribute('data-item-id'));
					const entry = shown.find(row => row.item.id === id);
					return (
						entry && (
							<ul className="mb-0 list-none p-0">
								{sourcesOf(entry)
									.slice(1)
									.map(label => (
										<li key={label}>{label}</li>
									))}
							</ul>
						)
					);
				}}
			/>
			<FiltersMenuView rows={FILTER_ROWS} open={filtersOpen} onOpenChange={setFiltersOpen} filters={filters} {...filtersMenu} />
		</Dialog>
	);
};

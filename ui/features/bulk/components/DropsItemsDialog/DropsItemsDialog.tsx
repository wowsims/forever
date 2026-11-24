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
import { useMediumLayout } from '@ui-kit/hooks/useMediumLayout';
import { Icon } from '@ui-kit/Icon';
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
import { Fragment, type KeyboardEvent, useId, useMemo, useRef, useState } from 'react';

import { dropsMatchesSearch, dropsPassesFilters } from '../../model/drops/filters';
import { type DropsItem, type DropsItemList, dropsItemsBySlot, dropsItemsBySourceParent, dropsTickCount } from '../../model/drops/items';
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

// `source` is the source key of a source group, which its rows leave out of their own sources. A source group sits under
// its `parent`, a group of its own that holds every item of its source groups.
interface RailGroup {
	key: string;
	source?: string;
	parent?: string;
	title: string;
	items: readonly DropsItem[];
}

interface RailParent {
	group: RailGroup;
	leaves: RailGroup[];
}

const parentKey = (key: string) => `parent:${key}`;

const sourceParents = (list: DropsItemList, tree: readonly DropsSourceNode[], names: DropsNames): RailParent[] =>
	dropsItemsBySourceParent(list, tree, names).map(parent => ({
		group: { key: parentKey(parent.key), title: parent.label, items: parent.items },
		leaves: parent.leaves.map(({ key, label, items }) => ({ key, source: key, parent: parentKey(parent.key), title: label.label, items })),
	}));

// "All" lists the list itself, which holds each item once whatever the grouping.
const railGroups = (list: DropsItemList, grouping: DropsGrouping, parents: readonly RailParent[]): RailGroup[] => [
	{ key: ALL, title: t('all'), items: list.items },
	...(grouping === 'slot'
		? dropsItemsBySlot(list).map(({ slot, items }) => ({ key: `slot:${slot}`, title: translateBulkSlotName(slot), items }))
		: parents.flatMap(parent => [parent.group, ...parent.leaves])),
];

const RAIL_NAV = clsx(
	'shrink-0 border-border max-md:-mx-(--modal-padding) max-md:flex max-md:scrollbar-none max-md:overflow-x-auto max-md:border-b max-md:px-(--modal-padding) md:w-72 md:overflow-y-auto md:border-e',
);
const RAIL_ENTRY = clsx(
	'flex w-full items-center gap-2 border-0 border-e-2 border-transparent px-2 py-1 text-left text-sm text-gray-400 hover:bg-gray-800 data-active:border-primary data-active:bg-gray-800 data-active:text-white max-md:w-auto max-md:shrink-0 max-md:gap-0 max-md:border-0 max-md:border-b-2 max-md:py-2 max-md:font-bold max-md:text-link max-md:hover:bg-transparent max-md:hover:text-link-hover max-md:data-active:bg-transparent max-md:data-active:text-white',
);

interface RailEntryProps {
	group: RailGroup;
	title?: string;
	active: boolean;
	onClick: () => void;
	className?: string;
}

// From md up a rail row with its counts at the end; below md a tab with its counts raised beside the name.
const RailEntry = ({ group, title = group.title, active, onClick, className }: RailEntryProps) => {
	const { ticked, total } = dropsTickCount(group.items);
	return (
		<Button
			variant={null}
			aria-current={active ? 'true' : undefined}
			data-active={active ? '' : undefined}
			className={clsx(RAIL_ENTRY, className)}
			onClick={onClick}>
			<span className="min-w-0 grow truncate">{title}</span>
			<span className="shrink-0 tabular-nums max-md:-mt-px max-md:self-start max-md:ps-px max-md:text-xs max-md:leading-none max-md:font-normal">
				{ticked}/{total}
			</span>
		</Button>
	);
};

interface SourceRailRowProps {
	group: RailGroup;
	active: boolean;
	onClick: () => void;
	onIncludedChange: DropsItemsDialogViewProps['onIncludedChange'];
	/** Set on a parent: the chevron at the end shows whether its sources are open. */
	expanded?: boolean;
	nested?: boolean;
}

// The source rail from md up, laid out as the Sources tree: a box that ticks every shown, non-worn item of the row, then
// the name and counts, which pick the row or open a parent, and a parent's chevron at the end.
const SourceRailRow = ({ group, active, onClick, onIncludedChange, expanded, nested }: SourceRailRowProps) => {
	const id = useId();
	const { ticked, total } = dropsTickCount(group.items);
	const ids = group.items.filter(entry => !entry.worn).map(entry => entry.item.id);
	return (
		<div
			data-active={active ? '' : undefined}
			className={clsx(
				'group flex items-center gap-2 border-e-2 border-transparent ps-2 hover:bg-gray-800 data-active:border-primary data-active:bg-gray-800',
				nested && 'ps-6',
			)}>
			<Field.Root
				id={id}
				layout="inline"
				size="xs"
				className="ui-boolean-picker grow-0 [--control-xs-height:--spacing(4)] [&_.ui-boolean-picker-indicator]:text-2xs">
				<BooleanPicker
					id={id}
					value={total > 0 && ticked === total}
					indeterminate={ticked > 0 && ticked < total}
					disabled={total === 0}
					onChange={checked => onIncludedChange(ids, checked)}
				/>
				<Field.Label className="sr-only">{group.title}</Field.Label>
			</Field.Root>
			<Button
				variant={null}
				aria-current={active ? 'true' : undefined}
				aria-expanded={expanded}
				className="flex min-w-0 grow items-center gap-2 py-1 ps-0 pe-2 text-left text-sm text-gray-400 group-data-active:text-white"
				onClick={onClick}>
				<span className="min-w-0 grow truncate">{group.title}</span>
				<span className="shrink-0 tabular-nums">
					{ticked}/{total}
				</span>
				{expanded !== undefined && (
					<Icon name="chevron-down" fixedWidth className={clsx('size-4 shrink-0 transition-transform duration-150', !expanded && '-rotate-90')} />
				)}
			</Button>
		</div>
	);
};

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

// Search, the filters and the rail pick live here and reach nothing that runs. The rail is built from the items search and
// the filters leave, so its counts and boxes follow them and a group they empty leaves it; a pick whose group leaves
// the rail reads as All. By source, a parent only opens from md up,
// where its sources are what a press picks. Below md the rail is one row of chips that scrolls sideways above the list,
// by source one per parent with a second row of the picked parent's sources, and the list fills the rest of the dialog.
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

	const [opened, setOpened] = useState<ReadonlySet<string>>(new Set());
	const medium = useMediumLayout();

	const visible = useMemo(
		() => ({ items: list.items.filter(entry => dropsPassesFilters(entry.item, filters) && dropsMatchesSearch(entry, search, names)) }),
		[list, filters, search, names],
	);
	const parents = useMemo(() => (grouping === 'source' ? sourceParents(visible, tree, names) : []), [visible, grouping, tree, names]);
	const groups = useMemo(() => railGroups(visible, grouping, parents), [visible, grouping, parents]);
	const picked = groups.find(group => group.key === pick) ?? groups[0];
	const pickedParent = parents.find(parent => parent.group.key === (picked.parent ?? picked.key));
	const toggle = (key: string) =>
		setOpened(was => {
			const next = new Set(was);
			if (!next.delete(key)) next.add(key);
			return next;
		});
	const shown = picked.items;
	const shownIds = shown.filter(entry => !entry.worn).map(entry => entry.item.id);
	const sourcesOf = (entry: DropsItem) => otherSources(entry, picked.source, names);

	const railEntry = (group: RailGroup, extra?: Partial<RailEntryProps>) => (
		<RailEntry key={group.key} group={group} active={group.key === picked.key} onClick={() => setPick(group.key)} {...extra} />
	);

	// The entries a press can pick, in rail order: by source, from md up the sources of the open parents, below md the
	// parents of the chip row.
	const railKeys =
		grouping === 'slot'
			? groups.map(group => group.key)
			: [ALL, ...parents.flatMap(parent => (medium ? (opened.has(parent.group.key) ? parent.leaves : []) : [parent.group])).map(group => group.key)];

	// On the popup, through `Dialog`'s own prop, as the gear picker's rail does. The filters menu is a dialog of its own.
	const onRailKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
		if (!open || filtersOpen) return;
		const next = railStep(railKeys, picked.key, event.key);
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
				{grouping === 'slot' || !medium ? (
					<nav aria-label={t(grouping === 'slot' ? 'by_slot' : 'by_source')} className={RAIL_NAV} data-testid="drops-items-rail">
						{(grouping === 'slot' ? groups : [groups[0], ...parents.map(parent => parent.group)]).map(group => railEntry(group))}
					</nav>
				) : (
					<nav aria-label={t('by_source')} className={RAIL_NAV} data-testid="drops-items-rail">
						<SourceRailRow
							group={groups[0]}
							active={groups[0].key === picked.key}
							onClick={() => setPick(ALL)}
							onIncludedChange={onIncludedChange}
						/>
						{parents.map(parent => {
							const expanded = opened.has(parent.group.key);
							return (
								<Fragment key={parent.group.key}>
									<SourceRailRow
										group={parent.group}
										active={parent.group.key === picked.key}
										expanded={expanded}
										onClick={() => toggle(parent.group.key)}
										onIncludedChange={onIncludedChange}
									/>
									{expanded &&
										parent.leaves.map(leaf => (
											<SourceRailRow
												key={leaf.key}
												group={leaf}
												active={leaf.key === picked.key}
												nested
												onClick={() => setPick(leaf.key)}
												onIncludedChange={onIncludedChange}
											/>
										))}
								</Fragment>
							);
						})}
					</nav>
				)}
				{grouping === 'source' && !medium && pickedParent && (
					<nav aria-label={pickedParent.group.title} className={clsx(RAIL_NAV, 'max-md:-mt-2')} data-testid="drops-items-subrail">
						{railEntry(pickedParent.group, {
							title: i18n.t('bulk_tab.drops.setup.all_of', { name: pickedParent.group.title }),
							className: 'max-md:text-xs max-md:font-normal',
						})}
						{pickedParent.leaves.map(leaf => railEntry(leaf, { className: 'max-md:text-xs max-md:font-normal' }))}
					</nav>
				)}
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

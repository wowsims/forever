import i18n from '@i18n/config';
import { translateBulkSlotName } from '@i18n/localization';
import { useActionId } from '@sim/hooks/useActionId';
import { ActionId } from '@sim/proto/action_id';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { ToggleGroup } from '@ui-kit/ToggleGroup';
import { itemQualityClassName } from '@ui-kit/utils/css';
import { VirtualList } from '@ui-kit/VirtualList';
import { WowheadIcon } from '@ui-kit/WowheadIcon';
import clsx from 'clsx';
import { useMemo, useRef, useState } from 'react';

import { type DropsItem, type DropsItemList, dropsItemsBySlot, dropsItemsBySource, dropsLeftOutCounts, type DropsLeftOutReason } from '../../model/drops/items';
import { type DropsNames, dropsSourceLabel } from '../../model/drops/sources';

export interface DropsItemsViewProps {
	list: DropsItemList;
	grouping: 'slot' | 'source';
	onGroupingChange: (grouping: 'slot' | 'source') => void;
	names: DropsNames;
	/** Called with the listed, non-worn ids of Include all / Exclude all, or one row's id. */
	onIncludedChange: (ids: readonly number[], included: boolean) => void;
}

const ROW_HEIGHT = 40;
const LEFT_OUT_REASONS: readonly DropsLeftOutReason[] = ['class', 'random_suffix'];

const t = (key: string, options: { count?: number; reasons?: string } = {}) => i18n.t(`bulk_tab.drops.setup.${key}`, options);

type ListRow = { kind: 'header'; key: string; title: string; detail?: string } | { kind: 'item'; key: string; entry: DropsItem };

const listRows = (list: DropsItemList, grouping: DropsItemsViewProps['grouping'], names: DropsNames): ListRow[] => {
	const groups =
		grouping === 'slot'
			? dropsItemsBySlot(list).map(({ slot, items }) => ({ key: `slot:${slot}`, title: translateBulkSlotName(slot), detail: undefined, items }))
			: dropsItemsBySource(list, names).map(({ key, label, items }) => ({ key, title: label.label, detail: label.detail, items }));
	return groups.flatMap(group => [
		{ kind: 'header' as const, key: group.key, title: group.title, detail: group.detail },
		...group.items.map(entry => ({ kind: 'item' as const, key: `${group.key}/${entry.item.id}`, entry })),
	]);
};

interface ItemRowProps {
	entry: DropsItem;
	names: DropsNames;
	onIncludedChange: DropsItemsViewProps['onIncludedChange'];
}

const ItemRow = ({ entry, names, onIncludedChange }: ItemRowProps) => {
	const { item, worn, included, tags, sourceKeys } = entry;
	const { iconUrl, href } = useActionId(ActionId.fromItem(item));
	const [first, ...more] = sourceKeys.map(key => dropsSourceLabel(key, names).label);

	return (
		<div
			className="flex h-full items-center gap-2 px-1 data-worn:opacity-50"
			data-testid="drops-item-row"
			data-item-id={item.id}
			data-worn={worn ? '' : undefined}>
			<WowheadIcon className="size-8 shrink-0" iconUrl={iconUrl} href={href || undefined} navigate={false} tabIndex={-1} />
			<BooleanPicker
				className={clsx('min-w-0 grow wrap-anywhere', itemQualityClassName(item.quality))}
				label={item.name}
				layout="inline"
				size="xs"
				value={included && !worn}
				disabled={worn}
				onChange={checked => onIncludedChange([item.id], checked)}
			/>
			<span
				className="shrink-0 truncate text-xs text-gray-400"
				title={more.length ? sourceKeys.map(key => dropsSourceLabel(key, names).label).join('\n') : undefined}>
				{worn ? i18n.t('bulk_tab.drops.items.worn') : first}
				{!worn && more.length > 0 && ` +${more.length}`}
			</span>
			{tags.map(tag => (
				<span key={tag} className="shrink-0 rounded-sm bg-gray-700 px-1 text-xs" data-testid="drops-item-tag">
					{i18n.t(`bulk_tab.drops.items.tags.${tag}`)}
				</span>
			))}
		</div>
	);
};

interface LeftOutProps {
	list: DropsItemList;
}

const LeftOut = ({ list }: LeftOutProps) => {
	const [shown, setShown] = useState(false);
	if (list.leftOut.length === 0) return null;
	const counts = dropsLeftOutCounts(list);
	const reasons = LEFT_OUT_REASONS.filter(reason => counts[reason] > 0)
		.map(reason => i18n.t(`bulk_tab.drops.items.left_out.${reason}`, { count: counts[reason] }))
		.join(', ');

	return (
		<div className="grid gap-1 text-sm text-gray-400" data-testid="drops-left-out">
			<p className="mb-0">
				{t('left_out_summary', { count: list.leftOut.length, reasons })}{' '}
				<Button variant="link" size="xs" onClick={() => setShown(wasShown => !wasShown)}>
					{t(shown ? 'hide' : 'show')}
				</Button>
			</p>
			{shown && (
				<ul className="mb-0 grid max-h-48 list-none gap-1 overflow-y-auto p-0 text-xs">
					{list.leftOut.map(({ item, reason }) => (
						<li key={item.id} className="flex justify-between gap-2">
							<span className={itemQualityClassName(item.quality)}>{item.name}</span>
							<span>{t(`left_out_reason.${reason}`)}</span>
						</li>
					))}
				</ul>
			)}
		</div>
	);
};

export const DropsItemsView = ({ list, grouping, onGroupingChange, names, onIncludedChange }: DropsItemsViewProps) => {
	const scroller = useRef<HTMLDivElement>(null);
	const rows = useMemo(() => listRows(list, grouping, names), [list, grouping, names]);
	const pickable = list.items.filter(entry => !entry.worn).map(entry => entry.item.id);

	return (
		<div className="grid content-start gap-2" data-testid="drops-items">
			<div className="flex flex-wrap items-center gap-2">
				<h5 className="mr-auto mb-0">{t('items_title')}</h5>
				<ToggleGroup
					size="xs"
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
					<Button variant="link" size="xs" disabled={pickable.length === 0} onClick={() => onIncludedChange(pickable, true)}>
						{t('include_all')}
					</Button>
					<Button variant="link" size="xs" disabled={pickable.length === 0} onClick={() => onIncludedChange(pickable, false)}>
						{t('exclude_all')}
					</Button>
				</div>
			</div>
			{rows.length === 0 ? (
				<p className="mb-0 text-sm text-gray-400">{t('empty')}</p>
			) : (
				<div ref={scroller} className="h-[60dvh] overflow-x-hidden overflow-y-scroll" data-testid="drops-item-list" tabIndex={0}>
					<VirtualList
						count={rows.length}
						rowHeight={ROW_HEIGHT}
						getScrollElement={() => scroller.current}
						renderRow={index => {
							const row = rows[index];
							return row.kind === 'header' ? (
								<div className="flex h-full items-end gap-2 border-b border-border pb-1 text-sm font-bold" data-testid="drops-group-header">
									<span>{row.title}</span>
									{row.detail && <span className="text-xs font-normal text-gray-400">{row.detail}</span>}
								</div>
							) : (
								<ItemRow entry={row.entry} names={names} onIncludedChange={onIncludedChange} />
							);
						}}
					/>
				</div>
			)}
			<LeftOut list={list} />
		</div>
	);
};

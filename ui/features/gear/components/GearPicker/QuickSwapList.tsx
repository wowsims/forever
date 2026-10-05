import type { UIEnchant as Enchant } from '@generated/proto/ui';
import { useActionId } from '@sim/hooks/useActionId';
import { ActionId } from '@sim/proto/action_id';
import { Button } from '@ui-kit/Button';
import { itemQualityClassName } from '@ui-kit/utils/css';
import { WowheadAnchor } from '@ui-kit/WowheadAnchor';
import clsx from 'clsx';

export interface QuickSwapEntry {
	item: Enchant;
	active: boolean;
}

export interface QuickSwapListProps {
	title: string;
	emptyMessage: string;
	entries: ReadonlyArray<QuickSwapEntry>;
	onItemClick: (item: Enchant) => void;
	footerButton: { label: string; onClick: () => void };
}

const QuickSwapRow = ({ entry, onItemClick }: { entry: QuickSwapEntry; onItemClick: (item: Enchant) => void }) => {
	const { iconUrl, href } = useActionId(ActionId.fromSpellId(entry.item.spellId));

	return (
		<li className="not-last:border-b not-last:border-b-border odd:[&_a]:bg-table-odd">
			<WowheadAnchor
				href={href || undefined}
				className="flex cursor-pointer items-center border border-transparent bg-table-even -outline-offset-1 transition-[background-color,color,border] duration-150 ease-in-out focus-visible:outline focus-visible:outline-1 focus-visible:outline-link data-active:border-success pointer-fine:hover:bg-table-row-even-hover"
				testId="tooltip-quick-swap__anchor"
				data-active={entry.active ? '' : undefined}
				onActivate={() => onItemClick(entry.item)}>
				<img
					alt={entry.item.name}
					className="static z-1 mr-1 inline-block size-8 shrink-0 cursor-pointer rounded-none bg-cover bg-center bg-no-repeat"
					src={iconUrl || undefined}
				/>
				<span className={clsx('truncate text-left', itemQualityClassName(entry.item.quality))} data-testid="tooltip-quick-swap__label">
					{entry.item.name}
				</span>
			</WowheadAnchor>
		</li>
	);
};

export const QuickSwapList = ({ title, emptyMessage, entries, onItemClick, footerButton }: QuickSwapListProps) => (
	<>
		<h3 className="mb-0 px-2 pt-2 pb-2 text-center text-(length:--h6-font-size)" data-testid="tooltip-quick-swap__title">
			{title}
		</h3>
		{entries.length ? (
			<ul className="mb-0 max-h-50 list-none gap-px overflow-y-auto pl-0">
				{entries.map(entry => (
					<QuickSwapRow key={`enchant-${entry.item.effectId}-${entry.item.type}`} entry={entry} onItemClick={onItemClick} />
				))}
			</ul>
		) : (
			<p className="mb-0 px-2 text-link-danger" data-testid="tooltip-quick-swap__empty">
				{emptyMessage}
			</p>
		)}
		<div className="mb-0 flex justify-center px-2 pt-2 pb-2" data-testid="tooltip-quick-swap__footer">
			<Button size="sm" onClick={footerButton.onClick}>
				{footerButton.label}
			</Button>
		</div>
	</>
);

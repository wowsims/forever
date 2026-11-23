import i18n from '@i18n/config';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import clsx from 'clsx';
import { useState } from 'react';

import { dropsNodeState, type DropsSourceNode } from '../../model/drops/sources';

export interface DropsSourcesViewProps {
	tree: readonly DropsSourceNode[];
	/** The ticked leaf keys, which may name nodes the tree does not list. */
	ticked: ReadonlySet<string>;
	onTick: (node: DropsSourceNode, checked: boolean) => void;
	onQuickPick: (pick: 'l60_dungeons' | 'clear') => void;
}

const t = (key: string) => i18n.t(`bulk_tab.drops.setup.${key}`);

interface SourceRowProps extends Pick<DropsSourcesViewProps, 'ticked' | 'onTick'> {
	node: DropsSourceNode;
	depth: number;
}

// A partly ticked node passes `value={false}` with `indeterminate`, so pressing it reports `true` and ticks the rest.
// Children mount only while the node is open.
const SourceRow = ({ node, depth, ticked, onTick }: SourceRowProps) => {
	const [open, setOpen] = useState(false);
	const state = dropsNodeState(node, ticked);
	const hasChildren = node.children.length > 0;

	return (
		<div>
			<div className="flex items-center gap-1" style={{ paddingInlineStart: `${depth * 1}rem` }}>
				{hasChildren ? (
					<Button
						variant={null}
						iconOnly
						aria-label={node.label}
						aria-expanded={open}
						className="size-4 shrink-0"
						onClick={() => setOpen(wasOpen => !wasOpen)}>
						<Icon name="chevron-down" fixedWidth className={clsx('transition-transform duration-150', !open && '-rotate-90')} />
					</Button>
				) : (
					<span className="size-4 shrink-0" aria-hidden />
				)}
				<BooleanPicker
					className="min-w-0 grow wrap-anywhere"
					label={node.label}
					layout="inline"
					size="xs"
					value={state === 'checked'}
					indeterminate={state === 'mixed'}
					onChange={checked => onTick(node, checked)}
				/>
				<span className="shrink-0 text-xs text-gray-400" data-testid="drops-source-count">
					{node.itemIds.length}
				</span>
			</div>
			{open && node.children.map(child => <SourceRow key={child.key} node={child} depth={depth + 1} ticked={ticked} onTick={onTick} />)}
		</div>
	);
};

export const DropsSourcesView = ({ tree, ticked, onTick, onQuickPick }: DropsSourcesViewProps) => (
	<div className="grid content-start gap-2" data-testid="drops-sources">
		<h5 className="mb-0">{t('sources_title')}</h5>
		<div className="flex flex-wrap gap-2">
			<Button variant="secondary" size="xs" onClick={() => onQuickPick('l60_dungeons')}>
				{t('quick_picks.l60_dungeons')}
			</Button>
			<Button variant="secondary" size="xs" onClick={() => onQuickPick('clear')}>
				{t('quick_picks.clear')}
			</Button>
		</div>
		<div className="grid gap-1">
			{tree.map(node => (
				<SourceRow key={node.key} node={node} depth={0} ticked={ticked} onTick={onTick} />
			))}
		</div>
	</div>
);

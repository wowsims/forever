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
	onQuickPick: (pick: 'endgame_dungeons' | 'clear') => void;
}

const t = (key: string) => i18n.t(`bulk_tab.drops.setup.${key}`);

interface SourceRowProps extends Pick<DropsSourcesViewProps, 'ticked' | 'onTick'> {
	node: DropsSourceNode;
}

// A partly ticked node passes `value={false}` with `indeterminate`, so pressing it reports `true` and ticks the rest.
// Children mount only while the node is open, indented under a guide line centred on the chevron, as the kit Tree indents.
const SourceRow = ({ node, ticked, onTick }: SourceRowProps) => {
	const [open, setOpen] = useState(false);
	const state = dropsNodeState(node, ticked);
	const hasChildren = node.children.length > 0;

	return (
		<div>
			<div className="flex items-center gap-1">
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
			{open && (
				<div className="ms-2 grid gap-1 border-s border-border ps-4 pt-1">
					{node.children.map(child => (
						<SourceRow key={child.key} node={child} ticked={ticked} onTick={onTick} />
					))}
				</div>
			)}
		</div>
	);
};

export const DropsSourcesView = ({ tree, ticked, onTick, onQuickPick }: DropsSourcesViewProps) => (
	<div className="grid content-start gap-4" data-testid="drops-sources">
		<div className="grid gap-2">
			<h5 className="mb-0">{t('sources_title')}</h5>
			<div className="flex flex-wrap gap-2">
				<Button variant="secondary" size="sm" onClick={() => onQuickPick('endgame_dungeons')}>
					{t('quick_picks.endgame_dungeons')}
				</Button>
				<Button variant="secondary" size="sm" onClick={() => onQuickPick('clear')}>
					{t('quick_picks.clear')}
				</Button>
			</div>
		</div>
		<div className="-m-1 grid max-h-[60dvh] gap-1 overflow-y-auto p-1">
			{tree.map(node => (
				<SourceRow key={node.key} node={node} ticked={ticked} onTick={onTick} />
			))}
		</div>
	</div>
);

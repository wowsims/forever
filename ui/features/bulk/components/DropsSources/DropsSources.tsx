import i18n from '@i18n/config';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { DisclosureChevron } from '@ui-kit/Icon';
import { Pill } from '@ui-kit/Pill';
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
// Children mount only while the node is open, indented under a guide line. The row reads like an accordion header: box, name, count badge, chevron on the right.
// On a node with children the name, badge and chevron are one button that opens it, apart from the box; a leaf's name ticks its box.
const SourceRow = ({ node, ticked, onTick }: SourceRowProps) => {
	const [open, setOpen] = useState(false);
	const state = dropsNodeState(node, ticked);
	const hasChildren = node.children.length > 0;

	return (
		<div>
			<div className="flex items-center gap-2 py-1 text-sm font-semibold">
				{hasChildren ? (
					<>
						<BooleanPicker
							aria-label={node.label}
							size="xs"
							value={state === 'checked'}
							indeterminate={state === 'mixed'}
							onChange={checked => onTick(node, checked)}
						/>
						<Button
							variant="unstyled"
							aria-label={node.label}
							aria-expanded={open}
							className="flex min-w-0 grow cursor-pointer items-center gap-2 border border-transparent pe-2 text-left leading-normal select-none focus-visible:focus-ring"
							onClick={() => setOpen(wasOpen => !wasOpen)}>
							<span className="min-w-0 wrap-anywhere">{node.label}</span>
							<Pill size="xs" count className="shrink-0" testId="drops-source-count">
								{node.itemIds.length}
							</Pill>
							<span className="grow" />
							<DisclosureChevron open={open} className="size-4 shrink-0" />
						</Button>
					</>
				) : (
					<>
						<BooleanPicker
							className="min-w-0 wrap-anywhere"
							label={node.label}
							layout="inline"
							size="xs"
							value={state === 'checked'}
							indeterminate={state === 'mixed'}
							onChange={checked => onTick(node, checked)}
						/>
						<Pill size="xs" count className="shrink-0" testId="drops-source-count">
							{node.itemIds.length}
						</Pill>
					</>
				)}
			</div>
			{open && (
				<div className="grid border-s border-border ps-2 pb-1">
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
		<div className="-m-1 grid max-h-[60dvh] overflow-y-auto p-1">
			{tree.map(node => (
				<SourceRow key={node.key} node={node} ticked={ticked} onTick={onTick} />
			))}
		</div>
	</div>
);

import { useActionId } from '@sim/hooks/useActionId';
import { WowheadIcon } from '@ui-kit/WowheadIcon';
import { memo } from 'react';

import type { ContentRow, RowItem } from '../../../model/timeline/rotation';
import { AuraItem } from './AuraItem';
import { CastItem } from './CastItem';
import { DelayItem } from './DelayItem';
import { GcdExtensionItem } from './GcdExtensionItem';
import { GcdSegmentItem } from './GcdSegmentItem';
import { ResourceItem } from './ResourceItem';
import { RotationRowIcon } from './RotationRowIcon';
import { RotationRowLabel } from './RotationRowLabel';
import { TickItem } from './TickItem';
import { rowStyle } from './utils';

export interface RowItemViewProps {
	item: RowItem;
	index: number;
	iconUrl: string;
	cssName: string;
}

/**
 * Memoised because a pan only changes the item window at its edges: every item that was already on
 * screen keeps the same model object and index, so its subtree is never diffed again.
 */
export const RowItemView = memo(({ item, index, iconUrl, cssName }: RowItemViewProps) => {
	switch (item.kind) {
		case 'cast':
			return <CastItem item={item} index={index} iconUrl={iconUrl} />;
		case 'delay':
			return <DelayItem item={item} index={index} />;
		case 'gcdExtension':
			return <GcdExtensionItem item={item} />;
		case 'gcdSegment':
			return <GcdSegmentItem item={item} index={index} />;
		case 'tick':
			return <TickItem item={item} index={index} />;
		case 'aura':
			return <AuraItem item={item} index={index} />;
		case 'resource':
			return <ResourceItem item={item} index={index} cssName={cssName} />;
	}
});
RowItemView.displayName = 'RowItemView';

export interface RotationRowProps {
	row: ContentRow;
	/** The row's item indexes inside the current track window, ascending. */
	items: ReadonlyArray<number>;
	onHide: (key: string) => void;
	hideTooltipId?: string;
}

/**
 * Memoised on the row and its item window: a frame that leaves both alone — every vertical scroll,
 * and every horizontal one that does not cross an item's edge — never renders the row again.
 */
export const RotationRow = memo(({ row, items, onHide, hideTooltipId }: RotationRowProps) => {
	// Every cast in a row is the same action, so the icon is resolved once here rather than per item.
	const { iconUrl } = useActionId(row.kind === 'resource' || row.kind === 'gcd' ? undefined : row.actionId);

	return (
		<div data-testid="rotation-row" data-row-kind={row.kind} className="ui-timeline-row" style={rowStyle(row)} data-row-key={row.key}>
			<RotationRowLabel
				text={row.label}
				icon={
					row.kind === 'resource' ? (
						<WowheadIcon testId="rotation-row-icon" className="ui-timeline-row-icon" iconUrl={row.icon} />
					) : row.kind === 'gcd' ? (
						// The GCD strip covers every ability, so its icon cell stays empty and only holds the column.
						<span className="ui-timeline-row-icon" />
					) : (
						<RotationRowIcon actionId={row.actionId} tooltip={row.kind === 'aura' ? 'buffAura' : 'spell'} />
					)
				}
				onHide={() => onHide(row.key)}
				hideTooltipId={hideTooltipId}
			/>
			<div data-testid="rotation-row-track" className="ui-timeline-row-track">
				{items.map(index => (
					<RowItemView key={index} item={row.items[index]} index={index} iconUrl={iconUrl} cssName={row.kind === 'resource' ? row.cssName : ''} />
				))}
			</div>
		</div>
	);
});
RotationRow.displayName = 'RotationRow';

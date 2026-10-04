import type { RowItem } from '../../../model/timeline/rotation';
import { ResourceTooltip } from '../tooltips/ResourceTooltip';
import { AuraTooltip } from './AuraTooltip';
import { CastTooltip } from './CastTooltip';
import { DelayTooltip } from './DelayTooltip';
import { GcdSegmentTooltip } from './GcdSegmentTooltip';
import { TickTooltip } from './TickTooltip';

export interface RowItemTooltipProps {
	item: RowItem;
}

/** What hovering one bar, tick or resource block shows. One entry per item kind, so a new kind is a type error. */
export const RowItemTooltip = ({ item }: RowItemTooltipProps) => {
	switch (item.kind) {
		case 'cast':
			return <CastTooltip log={item.log} stack={item.stack} />;
		case 'delay':
			return <DelayTooltip log={item.log} />;
		case 'gcdSegment':
			return <GcdSegmentTooltip log={item.log} />;
		// Named by the cast it extends, which owns the tooltip, so it carries no index to hover.
		case 'gcdExtension':
			return null;
		case 'tick':
			return <TickTooltip log={item.log} />;
		case 'aura':
			return <AuraTooltip log={item.log} />;
		case 'resource':
			return <ResourceTooltip log={item.log} maxValue={item.startValue} includeAuras={false} />;
	}
};

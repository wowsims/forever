import type { GcdSegmentItem as GcdSegmentItemModel } from '../../../model/timeline/rotation';
import { spanStyle } from './utils';

export interface GcdSegmentItemProps {
	item: GcdSegmentItemModel;
	index: number;
}

/** One cast's GCD on the strip row, centred in the item band rather than filling it. */
export const GcdSegmentItem = ({ item, index }: GcdSegmentItemProps) => (
	<div
		data-testid="rotation-item-gcd-segment"
		className="ui-timeline-item top-(--rotation-gcd-top) h-2 w-timeline-segment border-r border-black/35 bg-[rgb(80_180_235/0.55)]"
		data-item-index={index}
		style={spanStyle(item.start, item.end - item.start)}
	/>
);

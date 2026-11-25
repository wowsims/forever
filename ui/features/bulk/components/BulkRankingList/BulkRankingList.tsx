import { useScrollMargin, useScrollParent, VirtualList } from '@ui-kit/VirtualList';
import clsx from 'clsx';
import { Fragment, type ReactNode, useRef } from 'react';

import { BulkResultTieGroup } from '../BulkResults/BulkResultTieGroup';

export interface BulkRankingListProps<T> {
	/** The ranking in order: each entry one row, or the rows of a tie group. */
	chains: readonly (readonly T[])[];
	renderRow: (row: T) => ReactNode;
}

// About one plain row; a tie group or a row with a note measures taller once it renders.
const ESTIMATED_RANKING_HEIGHT = 120;

// One entry per ranking, a row or a tie group, measured as it renders, since a note, a tie group or a wrapped name
// changes its height. The list scrolls with whatever scrolls the pane.
export const BulkRankingList = <T,>({ chains, renderRow }: BulkRankingListProps<T>) => {
	const listRef = useRef<HTMLDivElement>(null);
	const scroller = useScrollParent(listRef);
	const scrollMargin = useScrollMargin(listRef, scroller);

	// Each ranking sits alone in its list entry, so the gap and the rule the rows draw between siblings go on the entry: a
	// row before a tie group keeps the gap without the rule. The rows inside a tie group are siblings and draw their own.
	const renderRanking = (index: number) => {
		const chain = chains[index];
		if (chain.length > 1) {
			return (
				<BulkResultTieGroup>
					{chain.map((row, rowIdx) => (
						<Fragment key={rowIdx}>{renderRow(row)}</Fragment>
					))}
				</BulkResultTieGroup>
			);
		}
		const last = index === chains.length - 1;
		const beforeGroup = (chains[index + 1]?.length ?? 0) > 1;
		return <div className={clsx(!last && 'mb-6', !last && !beforeGroup && 'border-b border-b-border pb-6')}>{renderRow(chain[0])}</div>;
	};

	return (
		<div ref={listRef}>
			<VirtualList
				measured
				count={chains.length}
				rowHeight={ESTIMATED_RANKING_HEIGHT}
				getScrollElement={() => scroller}
				scrollMargin={scrollMargin}
				renderRow={renderRanking}
				testId="bulk-ranking-list"
			/>
		</div>
	);
};

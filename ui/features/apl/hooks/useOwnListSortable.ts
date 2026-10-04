import type { TreeSortable } from '@ui-kit/Tree';
import { useMemo, useRef } from 'react';

/**
 * The drag of a nested APL list standing outside an `AplList` root, as in a story: the list owns the
 * drag, and the lists nested in its rows report their moves to it too, so it applies only the moves
 * inside its own `group`. Under an `AplList` root the root's drag wins and this one is never read.
 */
export const useOwnListSortable = (move: (from: number, to: number) => void, group: string): TreeSortable => {
	const latest = useRef({ move, group });
	latest.current = { move, group };
	return useMemo(
		() => ({
			onMove: ({ from, to }) => {
				if (from.group === latest.current.group && to.group === latest.current.group) latest.current.move(from.index, to.index);
			},
		}),
		[],
	);
};

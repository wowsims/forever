import { findScrollParent } from '@ui-kit/utils/dom';
import type { RefObject } from 'react';
import { useLayoutEffect, useState } from 'react';

/** The element that scrolls `list`, or `window` when none does; `null` until the list is in the document. */
export const useScrollParent = (list: RefObject<HTMLElement | null>): HTMLElement | Window | null => {
	const [scroller, setScroller] = useState<HTMLElement | Window | null>(null);

	useLayoutEffect(() => {
		if (list.current) setScroller(findScrollParent(list.current) ?? window);
	}, [list]);

	return scroller;
};

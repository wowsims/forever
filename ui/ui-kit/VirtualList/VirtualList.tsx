import { useVirtualizer } from '@tanstack/react-virtual';
import clsx from 'clsx';
import { type ReactNode, useEffect, useReducer } from 'react';

import { WINDOW_SCROLLER } from './window_scroller';

export interface VirtualListProps {
	count: number;
	// Every row is exactly this tall, unless `measured`, where it is the estimate a row keeps until it
	// has rendered once.
	rowHeight?: number;
	// For rows whose heights differ: each mounted row is measured, through the virtualiser's own
	// `measureElement`, and re-measured when it resizes. A row's height must not depend on where the
	// list places it, or a measurement moves the row and the next one measures it again.
	measured?: boolean;
	// The element that scrolls. May be an ancestor the list shares with other content, or `window`
	// when the list scrolls with the page. Returning `window` switches the list into window mode, so
	// a caller resolving its scroller with `findScrollParent` may report either without knowing which
	// it will be until the list is in the document.
	getScrollElement: () => HTMLElement | Window | null;
	// Pixels of a shared scroller above the first row: a sticky header, a filter bar. In window
	// mode this is the list's own offset down the document.
	scrollMargin?: number;
	className?: string;
	// Alternating row backgrounds by absolute index, and a lighter hovered row.
	striped?: boolean;
	rowClassName?: (index: number) => string | undefined;
	renderRow: (index: number) => ReactNode;
	testId?: string;
}

const DEFAULT_ROW_HEIGHT = 28;
const OVERSCAN = 10;

/**
 * Rows are absolutely positioned and moved with `transform`, which is how `@tanstack/react-virtual`
 * is built to work.
 *
 * The consequence to know: **`:nth-child` striping does not work here.** A row's position among its
 * siblings is the position within the rendered window, not within the list, so it changes as you
 * scroll. Every row carries `data-index` and `data-stripe`, and `striped` paints the stripes off
 * `data-stripe` instead.
 *
 * The transform is also a containing block for `position: fixed` descendants, so a tooltip, popover
 * or menu that a row renders **into itself** lands against the row rather than the viewport. Rows
 * that only ever anchor a popup rendered elsewhere — a Wowhead tooltip on `<body>`, a `Tooltip`
 * portalled out — are unaffected.
 */
export const VirtualList = ({
	count,
	rowHeight = DEFAULT_ROW_HEIGHT,
	getScrollElement,
	scrollMargin = 0,
	measured = false,
	className,
	striped = false,
	rowClassName,
	renderRow,
	testId = 'virtual-list',
}: VirtualListProps) => {
	// Read every render, because a caller that walks up for its scrolling ancestor has nothing to
	// report until the list is in the document. `null` takes the element branch: neither mode can
	// measure without a scroller, and the element one is what a ref-held scroller resolves to.
	const scroller = getScrollElement();
	const scrollsWithWindow = scroller !== null && !(scroller instanceof HTMLElement);

	const virtualizer = useVirtualizer<HTMLElement, HTMLElement>({
		count,
		getScrollElement: getScrollElement as () => HTMLElement | null,
		estimateSize: () => rowHeight,
		overscan: OVERSCAN,
		scrollMargin,
		...(scrollsWithWindow ? WINDOW_SCROLLER : {}),
	});

	// The virtualiser first looks for its scroller in a layout effect, which runs before a parent's ref
	// is attached; when it found none at mount but one is there now, one more render lets it subscribe.
	const [, rerender] = useReducer((renders: number) => renders + 1, 0);
	const subscribedTo = virtualizer.scrollElement;
	useEffect(() => {
		if (subscribedTo === null && getScrollElement() !== null) rerender();
	}, [subscribedTo, getScrollElement]);

	return (
		<div
			className={clsx('ui-virtual-list', className)}
			data-striped={striped ? '' : undefined}
			data-testid={testId}
			style={{ position: 'relative', height: virtualizer.getTotalSize() }}>
			{virtualizer.getVirtualItems().map(item => (
				<div
					key={item.key}
					data-index={item.index}
					data-stripe={item.index % 2 === 0 ? 'even' : 'odd'}
					data-testid="virtual-list-row"
					ref={measured ? virtualizer.measureElement : undefined}
					className={clsx('ui-virtual-list-row', rowClassName?.(item.index))}
					style={{
						position: 'absolute',
						top: 0,
						left: 0,
						width: '100%',
						height: measured ? undefined : rowHeight,
						transform: `translateY(${item.start - scrollMargin}px)`,
					}}>
					{renderRow(item.index)}
				</div>
			))}
		</div>
	);
};

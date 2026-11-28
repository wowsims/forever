import { usePortalContainer } from '@ui-kit/hooks/usePortalContainer';
import clsx from 'clsx';
import { forwardRef, type ReactNode, useCallback, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { type ITooltip, Tooltip as ReactTooltip, type TooltipRefProps } from 'react-tooltip';

import { hasHover } from '../../shared/pointer';

export type TooltipPlace = 'top' | 'right' | 'bottom' | 'left';

const DEFAULT_MAX_WIDTH = 'max-w-tooltip max-xl:max-w-tooltip-lg max-md:max-w-tooltip-sm';

export interface TooltipProps {
	/** Anchors opt in with `data-tooltip-id={id}`; one Tooltip can serve many of them. */
	id: string;
	/** A selector for the anchors in place of `data-tooltip-id`, for a tooltip every element of a kind opens. */
	anchorSelect?: string;
	content?: ReactNode;
	/** Per-anchor content, for one Tooltip serving a whole column: read the row off `activeAnchor`. Returning nothing renders no tooltip at all. */
	render?: ITooltip['render'];
	place?: TooltipPlace;
	/** Lets the pointer enter the tooltip, for content with links or buttons in it. */
	clickable?: boolean;
	/**
	 * The content links to the page the anchor shows, as the game tooltips do, and brings its own padding. The tooltip touches
	 * its anchor, and the pointer can move onto its links while the rest of it lets the pointer through to what lies beneath,
	 * so moving along a row of anchors still switches between them. Without hover, a tap on a link anchor opens the tooltip
	 * instead of following the link.
	 */
	carriesLink?: boolean;
	/** A popover: opens on click instead of hover, and closes on a click outside itself or Escape. */
	openOnClick?: boolean;
	/** Unmounts the tooltip. Use it to suppress one while a popover on the same control is open. */
	hidden?: boolean;
	/** Fires after the open and close transitions, not at the request — see `BonusStatsLink`. */
	onOpenChange?: (open: boolean) => void;
	widthClassName?: string;
	maxWidth?: 'default' | 'none' | string;
	align?: 'start' | 'center';
	padded?: boolean;
	className?: string;
	testId?: string;
}

// react-tooltip splits click-close into an anchor half and a global half, and defaults the anchor half off (`closeEvents.click`), so without `HOVER_CLOSE` the bonus-stat icon's tooltip stayed open behind the popover its own button had just opened.
const GLOBAL_CLOSE = { clickOutsideAnchor: true, escape: true };
// `openEvents` is deliberately never passed.
const HOVER_CLOSE = { mouseleave: true, blur: true, click: true };
// A tapped button that opens a dialog moves focus into it, which closes its label instead of leaving it over the dialog.
const TAP_CLOSE = { blur: true };

/** Content is `children`, which react-tooltip does not render until the tooltip first opens — so a picker built inside one costs nothing until it is shown. */
export const Tooltip = forwardRef<TooltipRefProps, TooltipProps>(
	(
		{
			id,
			anchorSelect,
			content,
			render,
			place = 'top',
			clickable,
			carriesLink,
			openOnClick,
			hidden,
			onOpenChange,
			widthClassName,
			maxWidth = 'default',
			align,
			padded,
			className,
			testId,
		},
		ref,
	) => {
		const maxWidthClassName = maxWidth === 'default' ? DEFAULT_MAX_WIDTH : maxWidth === 'none' ? 'max-w-none' : maxWidth;
		const portalContainer = usePortalContainer();
		// Without hover, react-tooltip opens a hover tooltip on `touchstart`, so a swipe that starts on an anchor opens it and no
		// `mouseleave` ever closes it. There a tap opens it and a tap anywhere else closes it.
		const touch = !hasHover();
		// The tooltip's own link matches the anchors it opens on, but is none of them.
		const inTooltip = useCallback((element: Element | null) => !!element && !!document.getElementById(id)?.contains(element), [id]);
		useEffect(() => {
			if (!carriesLink || !touch) return;
			const selector = anchorSelect ?? `[data-tooltip-id="${id}"]`;
			const tapLink = (event: MouseEvent) => {
				const anchor = event.target instanceof Element ? event.target.closest(selector) : null;
				if (anchor instanceof HTMLAnchorElement && anchor.href && !inTooltip(anchor)) event.preventDefault();
			};
			document.addEventListener('click', tapLink, true);
			return () => document.removeEventListener('click', tapLink, true);
		}, [carriesLink, touch, anchorSelect, id, inTooltip]);
		const tooltip = (
			<ReactTooltip
				ref={ref}
				id={id}
				anchorSelect={anchorSelect}
				render={render}
				place={place}
				// react-tooltip moves the tooltip to any other anchor the pointer touches, so a tooltip with a link overlaps its anchor
				// by its border. The pointer then reaches the link without crossing a gap that another anchor may reach into.
				offset={carriesLink ? -1 : undefined}
				clickable={clickable || carriesLink}
				disableTooltip={carriesLink ? inTooltip : undefined}
				openOnClick={openOnClick || touch}
				closeEvents={openOnClick ? undefined : touch ? TAP_CLOSE : HOVER_CLOSE}
				globalCloseEvents={GLOBAL_CLOSE}
				hidden={hidden}
				afterShow={onOpenChange && (() => onOpenChange(true))}
				afterHide={
					carriesLink || onOpenChange
						? () => {
								// react-tooltip clears its "pointer on the tooltip" flag only on a mouseout from the tooltip. A tooltip closed
								// with the pointer on its link, by Escape, never gets one, and until it does no tooltip hides on leaving its anchor.
								if (carriesLink) document.getElementById(id)?.dispatchEvent(new MouseEvent('mouseout'));
								onOpenChange?.(false);
							}
						: undefined
				}
				delayHide={0}
				delayShow={0}
				className={clsx(
					'sim-tooltip',
					'ui-tooltip',
					maxWidthClassName,
					widthClassName,
					align === 'start' && 'text-left',
					carriesLink && 'pointer-events-none [&_a]:pointer-events-auto',
					(padded === false || carriesLink) && 'sim-tooltip--unpadded',
					testId,
					className,
				)}
				noArrow
				disableStyleInjection="core">
				{content}
			</ReactTooltip>
		);
		return portalContainer ? createPortal(tooltip, portalContainer) : tooltip;
	},
);
Tooltip.displayName = 'Tooltip';

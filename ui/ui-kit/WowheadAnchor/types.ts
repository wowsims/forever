import type { AnchorHTMLAttributes, MouseEvent, Ref } from 'react';

export interface WowheadAnchorProps extends Omit<AnchorHTMLAttributes<HTMLAnchorElement>, 'href'> {
	href?: string;
	/** `false` leaves `data-whtticon` off a linked anchor. */
	icon?: boolean;
	/** `false` cancels the click; on its own it adds no role, tab index or touch opt-out — those stay keyed on `onActivate`. Default `true`. */
	navigate?: boolean;
	/**
	 * Cancels the click's navigation, then acts. It runs before `onClick`, which still receives the click. The anchor carries
	 * `data-disable-wowhead-touch-tooltip`, so a tap acts rather than opening wowhead's touch tooltip. Space activates it, linked or not;
	 * without an `href` it is a focusable `role="button"` that Enter activates too.
	 */
	onActivate?: (event: MouseEvent<HTMLAnchorElement>) => void;
	/** Marks the anchor `aria-disabled`, keeps an acting anchor out of the tab order, and never calls `onActivate`. An anchor that acts or has `navigate={false}` still cancels the click; a plain link still navigates. */
	disabled?: boolean;
	ref?: Ref<HTMLAnchorElement>;
	testId?: string;
}

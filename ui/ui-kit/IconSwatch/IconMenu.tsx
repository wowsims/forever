import { Menu } from '@ui-kit/Menu';
import clsx from 'clsx';
import type { ReactNode } from 'react';

import { IconSwatch, type IconSwatchProps } from './IconSwatch';

export interface IconMenuProps {
	/** The `IconSwatch` props for the trigger `IconMenu` renders: the swatch consumes `onActivate` (through `triggerProps`) to cancel its own click and let the menu's own toggle run instead. */
	trigger: IconSwatchProps;
	side: 'right' | 'bottom';
	onOpenChange?: (open: boolean) => void;
	className?: string;
	portalProps?: Record<string, unknown>;
	positionerProps?: Record<string, unknown>;
	popupProps?: Record<string, unknown>;
	children: ReactNode;
}

/** A grid of icons that opens on hover from a swatch, overlapping its border by a pixel. */
export const IconMenu = ({ trigger, side, onOpenChange, className, portalProps, positionerProps, popupProps, children }: IconMenuProps) => {
	return (
		<Menu
			onOpenChange={onOpenChange}
			surface="none"
			triggerRender={<IconSwatch {...trigger} />}
			// No href on the swatch leaves an anchor that is not focusable; `nativeButton: false` gives it the button role and a tab stop.
			// A swatch with a wowhead href is still a link. Base UI hands `onActivate` on to the swatch, which cancels the click, and its own click handler toggles the menu: a tap, which never hovers, opens the menu rather than the link.
			// The trigger's own disabled swatch dims and blocks its clicks; this also stops Base UI opening the menu from the trigger or a click relayed onto it.
			triggerProps={{ nativeButton: false, openOnHover: true, delay: 0, onActivate: () => {}, disabled: trigger.disabled }}
			side={side}
			align="start"
			sideOffset={-1}
			// `positionMethod="fixed"`, not the default `absolute`: Base UI renders the positioner `position: fixed` until it has a position, so Floating UI measures it against the viewport, and the switch to `absolute` then reads those viewport coordinates against whatever ancestor is `position: relative`.
			positionMethod="fixed"
			portalProps={portalProps}
			positionerProps={positionerProps}
			className={clsx('grid border-0 bg-grey', className)}
			popupProps={popupProps}>
			{children}
		</Menu>
	);
};

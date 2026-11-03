import { Menu as BaseMenu } from '@base-ui/react/menu';
import { isNativeButtonTrigger } from '@ui-kit/Button';
import { FieldContext } from '@ui-kit/Field/context/FieldContext';
import { usePortalContainer } from '@ui-kit/hooks/usePortalContainer';
import clsx from 'clsx';
import type { ReactElement, ReactNode } from 'react';

const POSITIONER_Z_CLASSES = {
	menu: 'ui-menu-positioner',
	plain: 'ui-menu-positioner-plain',
	none: 'z-dropdown',
} as const;

const SURFACE_CLASSES = {
	menu: 'ui-popup-surface ui-menu',
	plain: 'ui-menu-plain',
	none: '',
} as const;

interface MenuSharedProps {
	trigger?: ReactNode;
	triggerProps?: Record<string, unknown>;
	surface: 'menu' | 'plain' | 'none';
	anchorWidth?: boolean;
	side?: BaseMenu.Positioner.Props['side'];
	align?: BaseMenu.Positioner.Props['align'];
	sideOffset?: BaseMenu.Positioner.Props['sideOffset'];
	collisionPadding?: BaseMenu.Positioner.Props['collisionPadding'];
	positionMethod?: BaseMenu.Positioner.Props['positionMethod'];
	container?: HTMLElement | null;
	keepMounted?: boolean;
	className?: string;
	positionerClassName?: string;
	portalProps?: Record<string, unknown>;
	positionerProps?: Record<string, unknown>;
	popupRender?: ReactElement;
	popupProps?: Record<string, unknown>;
	children: ReactNode;
}

interface MenuRootProps extends MenuSharedProps {
	submenu?: false;
	triggerRender?: ReactElement;
	open?: boolean;
	onOpenChange?: (open: boolean) => void;
	modal?: boolean;
}

// A submenu's open state belongs to the menu tree, so the root's controlled-open props are not its
// to take, and its row element is the caller's rather than a default nothing forces it to supply.
interface MenuSubmenuProps extends MenuSharedProps {
	submenu: true;
	triggerRender: ReactElement;
	open?: never;
	onOpenChange?: never;
	modal?: never;
}

export type MenuProps = MenuRootProps | MenuSubmenuProps;

export const Menu = ({
	trigger,
	triggerRender,
	triggerProps,
	submenu,
	open,
	onOpenChange,
	modal = false,
	surface,
	anchorWidth,
	side,
	align,
	sideOffset,
	collisionPadding,
	positionMethod,
	container,
	keepMounted,
	className,
	positionerClassName,
	portalProps,
	positionerProps,
	popupRender,
	popupProps,
	children,
}: MenuProps) => {
	const portalContainer = usePortalContainer();
	// Base UI focuses a hover-opened popup and, with `highlightItemOnHover`, the item under the pointer. Both are script
	// focus, which Chrome paints as `:focus-visible` until the page has had a mouse focus, so the menu shows focus rings
	// under the mouse. The pointer highlight is `:hover` in Menu.css instead, and the popup itself draws no outline.
	// The popup leaves any Field the trigger sits in, so its controls take no size from that field.
	const popup = (
		<FieldContext.Provider value={null}>
			<BaseMenu.Portal container={container ?? portalContainer ?? undefined} keepMounted={keepMounted} {...portalProps}>
				<BaseMenu.Positioner
					side={side}
					align={align}
					sideOffset={sideOffset}
					collisionPadding={collisionPadding}
					positionMethod={positionMethod}
					className={clsx(POSITIONER_Z_CLASSES[surface], positionerClassName)}
					{...positionerProps}>
					<BaseMenu.Popup
						render={popupRender ?? <ul />}
						className={clsx('m-0 list-none p-0 outline-none', SURFACE_CLASSES[surface], anchorWidth && 'ui-menu-anchor-width', className)}
						{...popupProps}>
						{children}
					</BaseMenu.Popup>
				</BaseMenu.Positioner>
			</BaseMenu.Portal>
		</FieldContext.Provider>
	);

	// The row owns the popup: Base UI leaves focus guards where the portal sits, and they belong inside the `li`, not between the list's rows.
	if (submenu)
		return (
			<BaseMenu.SubmenuRoot highlightItemOnHover={false}>
				<li role="none">
					<BaseMenu.SubmenuTrigger render={triggerRender} nativeButton={isNativeButtonTrigger(triggerRender)} {...triggerProps}>
						{trigger}
					</BaseMenu.SubmenuTrigger>
					{popup}
				</li>
			</BaseMenu.SubmenuRoot>
		);

	return (
		<BaseMenu.Root open={open} onOpenChange={onOpenChange} modal={modal} highlightItemOnHover={false}>
			<BaseMenu.Trigger render={triggerRender} nativeButton={isNativeButtonTrigger(triggerRender)} {...triggerProps}>
				{trigger}
			</BaseMenu.Trigger>
			{popup}
		</BaseMenu.Root>
	);
};

import { ContextMenu as BaseContextMenu } from '@base-ui/react/context-menu';
import { usePortalContainer } from '@ui-kit/hooks/usePortalContainer';
import type { ReactNode } from 'react';

export interface ContextMenuProps {
	/** The area that opens the menu on a right click, a long press, the ContextMenu key or Shift+F10. */
	children: ReactNode;
	/** `MenuItem` rows. */
	menu: ReactNode;
	onOpenChange?: (open: boolean) => void;
	className?: string;
	triggerProps?: Record<string, unknown>;
}

// Focus stays where the caller puts it: a row's action decides where focus goes next, so the popup returns none.
export const ContextMenu = ({ children, menu, onOpenChange, className, triggerProps }: ContextMenuProps) => {
	const portalContainer = usePortalContainer();
	return (
		<BaseContextMenu.Root onOpenChange={onOpenChange && (open => onOpenChange(open))}>
			<BaseContextMenu.Trigger className={className} {...triggerProps}>
				{children}
			</BaseContextMenu.Trigger>
			<BaseContextMenu.Portal container={portalContainer ?? undefined}>
				<BaseContextMenu.Positioner className="ui-menu-positioner">
					<BaseContextMenu.Popup render={<ul />} finalFocus={false} className="ui-popup-surface ui-menu m-0 list-none p-0 outline-none">
						{menu}
					</BaseContextMenu.Popup>
				</BaseContextMenu.Positioner>
			</BaseContextMenu.Portal>
		</BaseContextMenu.Root>
	);
};

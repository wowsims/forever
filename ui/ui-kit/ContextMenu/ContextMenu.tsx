import { ContextMenu as BaseContextMenu } from '@base-ui/react/context-menu';
import { usePortalContainer } from '@ui-kit/hooks/usePortalContainer';
import type { ComponentProps, ReactNode } from 'react';

export interface ContextMenuProps {
	/** The area that opens the menu on a right click, a long press, the ContextMenu key or Shift+F10. */
	children: ReactNode;
	/** `MenuItem` rows. */
	menu: ReactNode;
	open?: boolean;
	onOpenChange?: (open: boolean) => void;
	/** After the open or close transition has ended and focus has settled. */
	onOpenChangeComplete?: (open: boolean) => void;
	/** Where focus goes when the menu closes: the element, `true` for the area, `false` to leave it where it is. */
	finalFocus?: ComponentProps<typeof BaseContextMenu.Popup>['finalFocus'];
	className?: string;
	triggerProps?: Record<string, unknown>;
}

export const ContextMenu = ({ children, menu, open, onOpenChange, onOpenChangeComplete, finalFocus, className, triggerProps }: ContextMenuProps) => {
	const portalContainer = usePortalContainer();
	return (
		<BaseContextMenu.Root open={open} onOpenChange={onOpenChange} onOpenChangeComplete={onOpenChangeComplete} highlightItemOnHover={false}>
			<BaseContextMenu.Trigger className={className} {...triggerProps}>
				{children}
			</BaseContextMenu.Trigger>
			<BaseContextMenu.Portal container={portalContainer ?? undefined}>
				<BaseContextMenu.Positioner className="ui-menu-positioner">
					<BaseContextMenu.Popup render={<ul />} finalFocus={finalFocus} className="ui-popup-surface ui-menu m-0 list-none p-0 outline-none">
						{menu}
					</BaseContextMenu.Popup>
				</BaseContextMenu.Positioner>
			</BaseContextMenu.Portal>
		</BaseContextMenu.Root>
	);
};

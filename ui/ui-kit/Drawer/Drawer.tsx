import { Drawer as BaseDrawer } from '@base-ui/react/drawer';
import { FieldContext } from '@ui-kit/Field/context/FieldContext';
import { usePortalContainer } from '@ui-kit/hooks/usePortalContainer';
import clsx from 'clsx';
import type { CSSProperties, ReactElement, ReactNode } from 'react';

export type DrawerSide = 'left' | 'bottom' | 'right';

const SWIPE_TO_CLOSE = { left: 'left', bottom: 'down', right: 'right' } as const satisfies Record<DrawerSide, BaseDrawer.Root.Props['swipeDirection']>;

export interface DrawerProps {
	trigger: ReactElement;
	open?: boolean;
	onOpenChange?: (open: boolean) => void;
	ignoreOutsidePress?: (event: Event) => boolean;
	modal?: boolean;
	/** Pins the drawer to that edge of the viewport as a sheet the kit paints, and a swipe toward the edge closes it. Without a side the caller places and paints the popup. */
	side?: DrawerSide;
	/** On the popup, which carries position and fade but no paint of its own: the caller's class places and paints it. */
	className?: string;
	/** Positions the portalled popup, since it does not share a containing block with its anchor. */
	style?: CSSProperties;
	children?: ReactNode;
	testId?: string;
}

export const Drawer = ({ trigger, open, onOpenChange, ignoreOutsidePress, modal = true, side, className, style, children, testId }: DrawerProps) => {
	const portalContainer = usePortalContainer();
	return (
		<BaseDrawer.Root
			open={open}
			modal={modal}
			swipeDirection={side ? SWIPE_TO_CLOSE[side] : undefined}
			onOpenChange={(nextOpen, eventDetails) => {
				if (!nextOpen && eventDetails.reason === 'outside-press' && ignoreOutsidePress?.(eventDetails.event)) {
					eventDetails.cancel();
					return;
				}
				onOpenChange?.(nextOpen);
			}}>
			<BaseDrawer.Trigger render={trigger} />
			{/* Out of any Field the trigger sits in, so the drawer's controls take no size from that field. */}
			<FieldContext.Provider value={null}>
				<BaseDrawer.Portal container={portalContainer ?? undefined}>
					<BaseDrawer.Viewport className="contents">
						<BaseDrawer.Popup className={clsx('ui-drawer', className)} style={style} data-side={side} data-testid={testId ?? 'sim-drawer-popup'}>
							{children}
						</BaseDrawer.Popup>
					</BaseDrawer.Viewport>
				</BaseDrawer.Portal>
			</FieldContext.Provider>
		</BaseDrawer.Root>
	);
};

import { Popover as BasePopover } from '@base-ui/react/popover';
import { isNativeButtonTrigger } from '@ui-kit/Button';
import { FieldContext } from '@ui-kit/Field/context/FieldContext';
import { usePortalContainer } from '@ui-kit/hooks/usePortalContainer';
import type { ClassValue } from 'clsx';
import clsx from 'clsx';
import { type ComponentPropsWithoutRef, type ReactElement, type ReactNode, type RefObject, useState } from 'react';

export type PopoverSide = 'top' | 'right' | 'bottom' | 'left';
export type PopoverAlign = 'start' | 'center' | 'end';

export interface PopoverProps {
	/** Contents of the trigger button. `Popover.Root` renders no element of its own, so the trigger is the only element the popover puts in the page's flow. Omit it and give `anchor` instead, for a popover a control opens without becoming its trigger. */
	trigger?: ReactNode;
	/** The trigger's own element, e.g. a kit `<Button variant="primary" />`, with `trigger` as its children; a plain `<button>` otherwise. */
	triggerRender?: ReactElement;
	/** What the popup is positioned against when there is no trigger. */
	anchor?: BasePopover.Positioner.Props['anchor'];
	triggerClassName?: ClassValue;
	triggerTestId?: string;
	/** Forwarded to the trigger `<button>`: `aria-label` for an icon-only trigger, and `tooltipAnchorProps(id)` for one that also carries a hover tooltip. */
	triggerProps?: Omit<ComponentPropsWithoutRef<'button'>, 'className' | 'children'>;
	/** The element focused on close, or its first tabbable child. Without it Base UI focuses the trigger, or what held focus before. */
	finalFocus?: RefObject<HTMLElement | null>;
	/** Omit for an uncontrolled popover — the trigger drives it. `onOpenChange` fires either way. */
	open?: boolean;
	onOpenChange?: (open: boolean) => void;
	/** On the popup, which is the box a consumer sizes. */
	className?: ClassValue;
	maxWidthClassName?: string;
	/** Base UI's default is `<body>`, and that is outside `.sim-ui` — where the spec theme lives. */
	container?: HTMLElement | null;
	side?: PopoverSide;
	align?: PopoverAlign;
	sideOffset?: BasePopover.Positioner.Props['sideOffset'];
	/** Dims the page behind the popup, as a dialog does; a press on it closes the popover. */
	backdrop?: boolean;
	/**
	 * A heading at the top of the popup that also names it for screen readers. The popup is then a column, so the
	 * heading stays in view above a body carrying `ui-popover-body`, which scrolls on its own across the popup's padding.
	 */
	title?: ReactNode;
	children?: ReactNode;
	testId?: string;
}

// Every anchored popup in the tree sits 10px off its anchor.
const ANCHOR_DISTANCE = 10;

export const Popover = ({
	trigger,
	triggerRender,
	anchor,
	triggerClassName,
	triggerTestId,
	triggerProps,
	finalFocus,
	open,
	onOpenChange,
	className,
	maxWidthClassName,
	container,
	side = 'bottom',
	align = 'center',
	sideOffset = ANCHOR_DISTANCE,
	backdrop = false,
	title,
	children,
	testId,
}: PopoverProps) => {
	const portalContainer = usePortalContainer();
	const [triggerElement, setTriggerElement] = useState<HTMLElement | null>(null);
	// The popup is portaled out of a dialog its trigger sits in, so only this keeps it inside that dialog.
	const collisionBoundary = triggerElement?.closest<HTMLElement>('[role="dialog"]') ?? undefined;
	return (
		<BasePopover.Root open={open} modal={false} onOpenChange={nextOpen => onOpenChange?.(nextOpen)}>
			{trigger != null && (
				<BasePopover.Trigger
					ref={setTriggerElement}
					render={triggerRender}
					nativeButton={isNativeButtonTrigger(triggerRender)}
					className={clsx(triggerClassName)}
					data-testid={triggerTestId}
					{...triggerProps}>
					{trigger}
				</BasePopover.Trigger>
			)}
			{/* Out of any Field the trigger sits in, so the popup's controls take no size from that field. */}
			<FieldContext.Provider value={null}>
				{/* Named, because with a `container` the portal renders a wrapper element of its own. */}
				<BasePopover.Portal className="contents" data-testid="sim-popover-portal" container={container ?? portalContainer ?? undefined}>
					{backdrop && <BasePopover.Backdrop className="ui-popover-backdrop" data-testid="sim-popover-backdrop" />}
					<BasePopover.Positioner
						className="ui-popover-positioner"
						data-testid="sim-popover-positioner"
						anchor={anchor}
						side={side}
						align={align}
						sideOffset={sideOffset}
						collisionBoundary={collisionBoundary}>
						<BasePopover.Popup
							className={clsx(
								maxWidthClassName ?? 'max-w-(--available-width)',
								'ui-popup-surface ui-popover',
								title != null && 'flex flex-col',
								className,
							)}
							data-testid={testId ?? 'sim-popover-popup'}
							initialFocus={false}
							finalFocus={finalFocus}>
							{title != null && <BasePopover.Title className="ui-popover-title">{title}</BasePopover.Title>}
							{children}
						</BasePopover.Popup>
					</BasePopover.Positioner>
				</BasePopover.Portal>
			</FieldContext.Provider>
		</BasePopover.Root>
	);
};

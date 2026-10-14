import { Dialog as BaseDialog } from '@base-ui/react/dialog';
import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { usePortalContainer } from '@ui-kit/hooks/usePortalContainer';
import { Icon } from '@ui-kit/Icon';
import clsx from 'clsx';
import type { KeyboardEventHandler, ReactNode } from 'react';

export type DialogSize = 'sm' | 'md' | 'lg' | 'xl' | 'full';

export interface DialogProps {
	open: boolean;
	/** Not called for a close the user is not allowed to make — see `preventClose`. */
	onOpenChange: (open: boolean) => void;
	className?: string;
	size?: DialogSize;
	maxWidthClassName?: string;
	title?: ReactNode;
	headerFlush?: boolean;
	/** Content beside the title, inside the header bar. */
	headerChildren?: ReactNode;
	footer?: ReactNode;
	/** Cap the popup at the viewport height and scroll the body, instead of scrolling the viewport. */
	scrollContents?: boolean;
	/** Removes the close button, the backdrop press and the Escape key. */
	preventClose?: boolean;
	/** Keep the dialog in the DOM while closed. */
	keepMounted?: boolean;
	/** For a dialog opened from another dialog or a popover, so its backdrop renders over what opened it. A popover opened inside an elevated dialog renders under it. */
	elevated?: boolean;
	verticalAlign?: 'top' | 'center';
	bodyClassName?: string;
	closeClassName?: string;
	/** On the popup, which is where Base UI stops keydown propagation — a listener above it never sees a key. */
	onKeyDown?: KeyboardEventHandler<HTMLDivElement>;
	children?: ReactNode;
	testId?: string;
}

export const Dialog = ({
	open,
	onOpenChange,
	className,
	size = 'lg',
	maxWidthClassName,
	title,
	headerFlush = false,
	headerChildren,
	footer,
	scrollContents = false,
	preventClose = false,
	keepMounted = false,
	elevated = false,
	verticalAlign = 'top',
	bodyClassName,
	closeClassName,
	onKeyDown,
	children,
	testId,
}: DialogProps) => {
	const portalContainer = usePortalContainer();
	return (
		<BaseDialog.Root
			open={open}
			onOpenChange={(nextOpen, details) => {
				if (!nextOpen && preventClose) {
					details.cancel();
					return;
				}
				onOpenChange(nextOpen);
			}}>
			{/* Named, because with a `container` the portal renders a wrapper element of its own. */}
			<BaseDialog.Portal data-testid="sim-dialog-portal" container={portalContainer ?? undefined} keepMounted={keepMounted}>
				{/* Base UI renders no backdrop for a nested dialog (`enabled: forceRender || !nested`), so an elevated one has to ask for its own. */}
				<BaseDialog.Backdrop className="ui-dialog-backdrop" data-testid="sim-dialog-backdrop" data-elevated={elevated} forceRender={elevated} />
				<BaseDialog.Viewport className="ui-dialog-viewport" data-testid="sim-dialog-viewport" data-elevated={elevated}>
					<BaseDialog.Popup
						className={clsx(
							'ui-dialog',
							maxWidthClassName,
							scrollContents && 'ui-dialog-scroll',
							verticalAlign === 'center' ? 'ui-dialog-centered' : 'ui-dialog-top',
							className,
						)}
						data-testid={testId ?? 'sim-dialog-popup'}
						data-size={size}
						onKeyDown={onKeyDown}>
						{(title != null || headerChildren != null || !preventClose) && (
							<div
								className={clsx('flex shrink-0 items-start', 'ui-dialog-header', headerFlush && 'ui-dialog-header-flush')}
								data-testid="sim-dialog-header"
								data-bare={false}>
								{title != null && (
									<BaseDialog.Title className="ui-dialog-title" data-testid="sim-dialog-title">
										{title}
									</BaseDialog.Title>
								)}
								{headerChildren}
								{!preventClose && (
									<BaseDialog.Close
										render={<Button iconOnly aria-label={i18n.t('common.close')} className={clsx('ui-dialog-close', closeClassName)} />}
										data-testid="sim-dialog-close">
										<Icon name="xmark" size="2xl" />
									</BaseDialog.Close>
								)}
							</div>
						)}
						<div className={clsx('ui-dialog-body', scrollContents && 'overflow-auto', bodyClassName)} data-testid="sim-dialog-body">
							{children}
						</div>
						{footer != null && (
							<div className="ui-dialog-footer" data-testid="sim-dialog-footer">
								{footer}
							</div>
						)}
					</BaseDialog.Popup>
				</BaseDialog.Viewport>
			</BaseDialog.Portal>
		</BaseDialog.Root>
	);
};

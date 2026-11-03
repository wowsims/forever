import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import type { PopoverProps } from '@ui-kit/Popover';
import { Popover } from '@ui-kit/Popover';
import type { ReactNode } from 'react';

export interface ConfirmPopoverProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	/** What the popover hangs off when the control that opens it is not its trigger. */
	anchor?: PopoverProps['anchor'];
	trigger?: ReactNode;
	triggerClassName?: PopoverProps['triggerClassName'];
	triggerProps?: PopoverProps['triggerProps'];
	finalFocus?: PopoverProps['finalFocus'];
	side?: PopoverProps['side'];
	backdrop?: PopoverProps['backdrop'];
	confirmLabel?: string;
	/** Omit for a message that is only acknowledged: one dismissing button, no cancel. */
	onConfirm?: () => void;
	children?: ReactNode;
	testId?: string;
}

export const ConfirmPopover = ({
	open,
	onOpenChange,
	anchor,
	trigger,
	triggerClassName,
	triggerProps,
	finalFocus,
	side = 'top',
	backdrop = true,
	confirmLabel,
	onConfirm,
	children,
	testId,
}: ConfirmPopoverProps) => (
	<Popover
		open={open}
		onOpenChange={onOpenChange}
		anchor={anchor}
		trigger={trigger}
		triggerClassName={triggerClassName}
		triggerProps={triggerProps}
		finalFocus={finalFocus}
		side={side}
		backdrop={backdrop}
		testId={testId ?? 'sim-confirm-popover'}
		maxWidthClassName="max-w-54">
		<p className="mb-2" data-testid="sim-confirm-popover-message">
			{children}
		</p>
		<div className="flex justify-end gap-2" data-testid="sim-confirm-popover-actions">
			{onConfirm ? (
				<>
					<Button variant="outline-cancel" size="sm" onClick={() => onOpenChange(false)}>
						{i18n.t('common.cancel')}
					</Button>
					<Button
						variant="cancel"
						size="sm"
						onClick={() => {
							onConfirm();
							onOpenChange(false);
						}}>
						{confirmLabel ?? i18n.t('common.confirm')}
					</Button>
				</>
			) : (
				<Button variant="primary" size="sm" onClick={() => onOpenChange(false)}>
					{confirmLabel ?? i18n.t('common.ok')}
				</Button>
			)}
		</div>
	</Popover>
);

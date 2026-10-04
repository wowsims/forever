import { ConfirmPopover } from '@ui-kit/ConfirmPopover';
import { Icon } from '@ui-kit/Icon';
import { tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx from 'clsx';
import { type ComponentPropsWithoutRef, type ElementType, type ReactNode, useState } from 'react';

export interface ChipProps {
	label: ReactNode;
	active?: boolean;
	disabled?: boolean;
	onSelect?: () => void;
	onDelete?: () => void;
	deleteLabel?: string;
	deleteTooltipId?: string;
	chipTooltipId?: string;
	tooltip?: string;
	deleteMessage?: ReactNode;
	deleteConfirmLabel?: string;
	className?: string;
	testId?: string;
	nameTestId?: string;
	deleteTestId?: string;
	as?: ElementType;
	nameAs?: ElementType;
	rootProps?: Record<string, unknown>;
	nameProps?: Record<string, unknown>;
	confirmDelete?: boolean;
}

export const Chip = ({
	label,
	active,
	disabled,
	onSelect,
	onDelete,
	deleteLabel,
	deleteTooltipId,
	chipTooltipId,
	tooltip,
	deleteMessage,
	deleteConfirmLabel,
	className,
	testId,
	nameTestId,
	deleteTestId,
	as: Root = 'div',
	nameAs: Name = 'button',
	rootProps,
	nameProps,
	confirmDelete = true,
}: ChipProps) => {
	const [confirming, setConfirming] = useState(false);

	return (
		<Root
			className={clsx('ui-chip', className)}
			data-testid={testId}
			data-active={active ? '' : undefined}
			data-disabled={disabled ? '' : undefined}
			{...rootProps}>
			<Name
				className="ui-chip-name"
				data-testid={nameTestId}
				onClick={onSelect}
				{...(Name === 'button' ? { type: 'button' } : {})}
				{...tooltipAnchorProps(tooltip ? chipTooltipId : undefined, tooltip)}
				{...nameProps}>
				{label}
			</Name>
			{onDelete && !confirmDelete && (
				<button
					type="button"
					className="ui-chip-delete"
					data-testid={deleteTestId}
					aria-label={deleteLabel}
					{...tooltipAnchorProps(deleteTooltipId)}
					onClick={onDelete}>
					<Icon name="xmark" style="base" size="lg" />
				</button>
			)}
			{onDelete && confirmDelete && (
				<ConfirmPopover
					open={confirming}
					onOpenChange={setConfirming}
					trigger={<Icon name="xmark" style="base" size="lg" />}
					triggerClassName="ui-chip-delete"
					triggerProps={
						{
							'aria-label': deleteLabel,
							'data-testid': deleteTestId,
							...tooltipAnchorProps(deleteTooltipId),
						} as ComponentPropsWithoutRef<'button'>
					}
					confirmLabel={deleteConfirmLabel}
					onConfirm={onDelete}>
					{deleteMessage}
				</ConfirmPopover>
			)}
		</Root>
	);
};

import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { Tooltip, tooltipAnchorProps, type TooltipPlace } from '@ui-kit/Tooltip';
import clsx from 'clsx';
import { type ReactNode, useId } from 'react';

interface TooltipButtonBaseProps {
	place?: TooltipPlace;
	className?: string;
	testId?: string;
}

/** The accessible name is `label`, or a string tooltip itself; rich tooltip content needs a `label`. */
export type TooltipButtonProps = TooltipButtonBaseProps & ({ tooltip: string; label?: string } | { tooltip: ReactNode; label: string });

export const TooltipButton = ({ tooltip, label, place, className, testId = 'tooltip-button' }: TooltipButtonProps) => {
	const id = useId();
	return (
		<>
			<Button
				variant="link"
				iconOnly
				aria-label={label ?? String(tooltip)}
				className={clsx('flex items-center justify-center', className)}
				data-testid={testId}
				{...tooltipAnchorProps(id)}>
				<Icon name="circle-question" style="regular" />
			</Button>
			<Tooltip id={id} content={tooltip} place={place} />
		</>
	);
};

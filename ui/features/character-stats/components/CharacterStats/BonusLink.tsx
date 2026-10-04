import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { Tooltip, tooltipAnchorProps, type TooltipRefProps } from '@ui-kit/Tooltip';
import { type ReactNode, useCallback, useId, useRef, useState } from 'react';

export interface BonusLinkProps {
	label: string;
	/** The popover's field; `onCommit` closes the popover. */
	picker: (onCommit: () => void) => ReactNode;
}

export const BonusLink = ({ label, picker }: BonusLinkProps) => {
	const id = useId();
	const popover = useRef<TooltipRefProps>(null);
	const [popoverOpen, setPopoverOpen] = useState(false);
	const closePopover = useCallback(() => popover.current?.close(), []);

	return (
		<>
			<Button variant="unstyled" data-testid="add-bonus-stats" className="ml-2 text-white" aria-label={label} {...tooltipAnchorProps(`${id}-popover`)}>
				<Icon name="plus-minus" {...tooltipAnchorProps(`${id}-icon`)} />
			</Button>
			<Tooltip id={`${id}-icon`} content={label} hidden={popoverOpen} />
			<Tooltip
				ref={popover}
				id={`${id}-popover`}
				testId="bonus-stats-popover"
				align="start"
				place="right"
				openOnClick
				clickable
				onOpenChange={setPopoverOpen}
				content={picker(closePopover)}
			/>
		</>
	);
};

import { Tooltip } from '@ui-kit/Tooltip';
import { TooltipAnchor } from '@ui-kit/TooltipAnchor';
import { useId } from 'react';

export interface DiscoveredDotProps {
	/** When the item was discovered, shown in the dot's tooltip and read as its name. */
	label: string;
}

// The anchor is larger than the dot, so the dot stays easy to hover.
export const DiscoveredDot = ({ label }: DiscoveredDotProps) => {
	const tooltipId = useId();
	return (
		<>
			<TooltipAnchor
				tooltipId={tooltipId}
				className="pointer-events-auto absolute top-0 right-0 flex size-4 translate-x-1/2 -translate-y-1/2 items-center justify-center"
				testId="selector-modal-list-item-discovered">
				<span className="size-[6px] rounded-full bg-success shadow-discovered-dot" />
				<span className="sr-only">{label}</span>
			</TooltipAnchor>
			<Tooltip id={tooltipId} content={label} />
		</>
	);
};

import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Icon, type IconName } from '@ui-kit/Icon';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { forwardRef, useId } from 'react';

export interface RotationToolbarProps {
	onZoomOut: () => void;
	onZoomIn: () => void;
	onFit: () => void;
	onReset: () => void;
}

/**
 * The corner above the row labels. Exactly `--label-w` wide — the view measures it to find the
 * resolved clamp, which is the only reason it takes a ref — so the ruler beside it lines up with
 * every track.
 */
export const RotationToolbar = forwardRef<HTMLDivElement, RotationToolbarProps>(({ onZoomOut, onZoomIn, onFit, onReset }, ref) => {
	const tooltipId = useId();
	const buttons: Array<{ key: string; icon: IconName; run: () => void }> = [
		{ key: 'zoom_out', icon: 'magnifying-glass-minus', run: onZoomOut },
		{ key: 'zoom_in', icon: 'magnifying-glass-plus', run: onZoomIn },
		{ key: 'fit', icon: 'expand', run: onFit },
		{ key: 'reset', icon: 'rotate-left', run: onReset },
	];

	return (
		<div ref={ref} data-testid="rotation-corner" className="ui-timeline-label-col">
			{buttons.map(button => {
				const label = i18n.t(`results_tab.details.timeline.chart_options.${button.key}`);
				return (
					<Button
						key={button.key}
						iconOnly
						aria-label={label}
						data-testid="rotation-zoom-button"
						className="px-1 py-0 leading-none text-white hover:text-brand"
						onClick={button.run}
						{...tooltipAnchorProps(tooltipId, label)}>
						<Icon name={button.icon} />
					</Button>
				);
			})}
			<Tooltip id={tooltipId} place="bottom" />
		</div>
	);
});
RotationToolbar.displayName = 'RotationToolbar';

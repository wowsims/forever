import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { ButtonGroup } from '@ui-kit/ButtonGroup';
import { Icon, type IconName } from '@ui-kit/Icon';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { useId } from 'react';

export interface ChartToolbarProps {
	onReset: () => void;
	onZoomIn: () => void;
	onZoomOut: () => void;
	onPanLeft: () => void;
	onPanRight: () => void;
}

export const ChartToolbar = ({ onReset, onZoomIn, onZoomOut, onPanLeft, onPanRight }: ChartToolbarProps) => {
	const tooltipId = useId();
	const buttons: Array<{ key: string; icon: IconName; run: () => void }> = [
		{ key: 'reset', icon: 'arrows-rotate', run: onReset },
		{ key: 'zoom_out', icon: 'magnifying-glass-minus', run: onZoomOut },
		{ key: 'zoom_in', icon: 'magnifying-glass-plus', run: onZoomIn },
		{ key: 'pan_left', icon: 'chevron-left', run: onPanLeft },
		{ key: 'pan_right', icon: 'chevron-right', run: onPanRight },
	];

	return (
		<ButtonGroup data-testid="timeline-chart-toolbar" className="ml-auto w-auto shrink-0 self-start" size="sm">
			{buttons.map(button => {
				const label = i18n.t(`results_tab.details.timeline.chart_options.${button.key}`);
				return (
					<Button key={button.key} variant="outline-primary" aria-label={label} onClick={button.run} {...tooltipAnchorProps(tooltipId, label)}>
						<Icon name={button.icon} />
					</Button>
				);
			})}
			<Tooltip id={tooltipId} place="bottom" />
		</ButtonGroup>
	);
};

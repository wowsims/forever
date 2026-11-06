import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { Icon, type IconName } from '@ui-kit/Icon';
import { Menu, MenuItem } from '@ui-kit/Menu';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { useId } from 'react';

import type { PlanEventKind } from '../../model/lanes';
import { eventLabelKey, LANE_EVENT_KINDS, PLAN_LANES } from '../../model/lanes';
import { plannerText } from '../EncounterPlanner/utils/configs';
import { formatSeconds } from '../EncounterPlanner/utils/tips';

export interface PlannerToolbarProps {
	snapSeconds: number;
	onSnapChange: (seconds: number) => void;
	onZoomIn: () => void;
	onZoomOut: () => void;
	onFit: () => void;
	onAdd: (kind: PlanEventKind) => void;
}

const SNAP_STEPS: ReadonlyArray<number> = [0, 0.5, 1, 5, 10];

export const PlannerToolbar = ({ snapSeconds, onSnapChange, onZoomIn, onZoomOut, onFit, onAdd }: PlannerToolbarProps) => {
	const tooltipId = useId();
	const zoomButtons: Array<{ key: string; icon: IconName; run: () => void }> = [
		{ key: 'zoom_out', icon: 'magnifying-glass-minus', run: onZoomOut },
		{ key: 'zoom_in', icon: 'magnifying-glass-plus', run: onZoomIn },
		{ key: 'fit', icon: 'expand', run: onFit },
	];

	return (
		<div className="flex flex-wrap items-center gap-4">
			<div className="flex items-center gap-2">
				{zoomButtons.map(button => {
					const label = plannerText(`toolbar.${button.key}`);
					return (
						<Button key={button.key} iconOnly aria-label={label} onClick={button.run} {...tooltipAnchorProps(tooltipId, label)}>
							<Icon name={button.icon} />
						</Button>
					);
				})}
			</div>
			<ChoiceField
				id="encounter-planner-snap"
				label={plannerText('toolbar.snap')}
				layout="inline"
				size="xs"
				options={SNAP_STEPS.map(step => ({ value: step, label: step > 0 ? formatSeconds(step) : plannerText('toolbar.snap_off') }))}
				value={snapSeconds}
				onChange={onSnapChange}
			/>
			<Menu
				surface="menu"
				align="start"
				trigger={
					<>
						<Icon name="plus" /> {plannerText('toolbar.add_event')}
					</>
				}
				triggerRender={<Button variant="primary" size="sm" />}>
				{PLAN_LANES.flatMap(lane => LANE_EVENT_KINDS[lane]).map(kind => (
					<MenuItem key={kind} layout="block" onClick={() => onAdd(kind)}>
						{i18n.t(eventLabelKey(kind))}
					</MenuItem>
				))}
			</Menu>
			<Tooltip id={tooltipId} place="bottom" />
		</div>
	);
};

import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { cssVars } from '@ui-kit/utils/css';

import { BAR_CLASSES, barVars } from '../DesignerEventBar/utils';
import { LANE_CLASSES, LANE_GRID_CLASSES, LANE_LABEL_CLASSES } from '../DesignerLane/utils';
import { designerText } from '../EncounterDesigner/utils/configs';

export interface DesignerTargetLaneProps {
	/** The encounter's target names, one bar each. */
	targetNames: ReadonlyArray<string>;
	baseSeconds: number;
	selected?: number;
	onSelect: (index: number) => void;
	/** Enter on a bar, after `onSelect`; a click only selects. */
	onActivate: (index: number) => void;
	onAdd: () => void;
}

// A target is up for the whole fight, so its bar spans the fight and takes no drag, resize or arrow key.
export const DesignerTargetLane = ({ targetNames, baseSeconds, selected, onSelect, onActivate, onAdd }: DesignerTargetLaneProps) => {
	const label = designerText('targets');
	const targetLabel = i18n.t('settings_tab.encounter.target');
	return (
		<div role="group" aria-label={label} className={LANE_CLASSES} style={cssVars({ '--bar': 'var(--designer-lane-targets-bar)' })}>
			<div className={LANE_LABEL_CLASSES}>
				<span className="text-sm font-bold text-white">{label}</span>
				<div className="flex flex-wrap gap-1">
					<Button variant="outline-secondary" size="xs" aria-label={designerText('add', { event: targetLabel })} onClick={onAdd}>
						<Icon name="plus" /> {targetLabel}
					</Button>
				</div>
			</div>
			<div className={LANE_GRID_CLASSES}>
				{targetNames.map((_, index) => (
					<div key={index} aria-hidden="true" />
				))}
				{targetNames.map((name, index) => (
					<Button
						key={index}
						variant="unstyled"
						aria-pressed={index === selected}
						data-selected={index === selected ? '' : undefined}
						className={`absolute left-designer-pos z-1 col-start-1 row-(--row) text-xs font-bold text-white outline-none focus-visible:focus-ring data-selected:z-2 ${BAR_CLASSES}`}
						style={barVars(index, 0, baseSeconds)}
						onClick={() => onSelect(index)}
						onKeyDown={event => {
							if (event.key !== 'Enter') return;
							event.preventDefault();
							onSelect(index);
							onActivate(index);
						}}>
						<span className="min-w-0 truncate px-1">{name}</span>
					</Button>
				))}
			</div>
		</div>
	);
};

import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { cssVars } from '@ui-kit/utils/css';

import { barVars } from '../DesignerEventBar/utils';
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
		<div
			role="group"
			aria-label={label}
			className="flex min-w-full border-b border-gray-700"
			style={cssVars({ '--bar': 'var(--designer-lane-targets-bar)' })}>
			<div className="sticky left-0 z-3 box-border flex w-(--label-w) shrink-0 grow-0 basis-(--label-w) flex-col justify-center gap-1 border-r border-gray-700 bg-background px-2 py-1">
				<span className="text-sm font-bold text-white">{label}</span>
				<div className="flex flex-wrap gap-1">
					<Button variant="outline-secondary" size="xs" aria-label={designerText('add', { event: targetLabel })} onClick={onAdd}>
						<Icon name="plus" /> {targetLabel}
					</Button>
				</div>
			</div>
			<div className="relative grid shrink-0 grow basis-designer-total auto-rows-(--designer-lane-height) grid-cols-1">
				{targetNames.map((_, index) => (
					<div key={index} aria-hidden="true" />
				))}
				{targetNames.map((name, index) => (
					<Button
						key={index}
						variant="unstyled"
						aria-pressed={index === selected}
						data-selected={index === selected ? '' : undefined}
						className="absolute top-1 bottom-1 left-designer-pos z-1 col-start-1 row-(--row) flex w-timeline-segment min-w-2 items-center rounded-xs border border-transparent bg-(--bar) text-xs font-bold text-white outline-none focus-visible:focus-ring data-selected:z-2 data-selected:border-white"
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

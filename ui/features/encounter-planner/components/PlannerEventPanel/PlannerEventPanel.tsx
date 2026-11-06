import type { ChoiceFieldConfig } from '@features/inputs/configs';
import type { AddWave, PresetTarget, Target } from '@generated/proto/common';
import { EncounterPlan } from '@generated/proto/common';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { FieldGrid } from '@ui-kit/FieldGrid';
import { Icon } from '@ui-kit/Icon';
import { MultiComboBox } from '@ui-kit/MultiComboBox';
import { NumberField, StringField } from '@ui-kit/TextField';
import type { Ref } from 'react';

import type { PlanEvent, PlannerSelection } from '../../model/lanes';
import { getEvent, isEventRef, removeEvent } from '../../model/lanes';
import type { PlanFieldConfig, PlanFieldContext } from '../EncounterPlanner/utils/configs';
import { editEvent, planFieldConfigs, plannerText } from '../EncounterPlanner/utils/configs';
import { eventName } from '../EncounterPlanner/utils/tips';
import { PlannerAdds } from '../PlannerAdds';
import { PlannerAddStats } from '../PlannerAddStats';

/** The selected encounter target, whose edits go to the encounter as they happen rather than into the plan. */
export interface PlannerTargetEditor {
	target: Target;
	/** Starts the field ids; the NPC and preset target fields are built with it too. */
	idPrefix: string;
	/** The NPC and preset target fields, bound to the encounter. */
	identity: ReadonlyArray<ChoiceFieldConfig<null>>;
	/** Shows the fields only a threat sim reads. */
	threatMetrics: boolean;
	canRemove: boolean;
	onModify: (edit: (target: Target) => void) => void;
	/** Swaps the target with the one before it (-1) or after it (1). */
	onMove: (direction: -1 | 1) => void;
	/** The copy goes at the end of the list. */
	onCopy: () => void;
	onRemove: () => void;
}

export interface PlannerEventPanelProps {
	plan: EncounterPlan;
	selected?: PlannerSelection;
	baseSeconds: number;
	/** The encounter's targets by index, for the target pickers. */
	targetNames: ReadonlyArray<string>;
	/** The preset targets a wave can copy as its add. */
	presetTargets: ReadonlyArray<PresetTarget>;
	/** A new add of a wave: the default template's copy. */
	newAdd: () => Target;
	/** The selected target's fields and actions, when a target is selected. */
	targetEditor?: PlannerTargetEditor;
	onChange: (plan: EncounterPlan) => void;
	onSelect: (selection: PlannerSelection | undefined) => void;
	/** The panel takes focus from script only, for a container that moves focus here when a bar is chosen. */
	ref?: Ref<HTMLElement>;
}

export const PlannerEventPanel = ({
	plan,
	selected,
	baseSeconds,
	targetNames,
	presetTargets,
	newAdd,
	targetEditor,
	onChange,
	onSelect,
	ref,
}: PlannerEventPanelProps) => {
	if (selected?.kind === 'target' && targetEditor) {
		const name = targetNames[selected.index];
		return (
			<section ref={ref} tabIndex={-1} aria-label={name} className="flex flex-col gap-2 outline-none">
				<div className="flex items-center justify-between gap-2">
					<span className="text-base font-bold text-white">{name}</span>
					{targetEditor.canRemove && (
						<Button variant="link-danger" size="sm" onClick={targetEditor.onRemove}>
							{plannerText('panel.remove')}
							<Icon name="xmark" />
						</Button>
					)}
				</div>
				<div className="flex flex-wrap gap-2">
					<Button variant="outline-secondary" size="xs" disabled={selected.index === 0} onClick={() => targetEditor.onMove(-1)}>
						{plannerText('panel.move_up')}
					</Button>
					<Button variant="outline-secondary" size="xs" disabled={selected.index === targetNames.length - 1} onClick={() => targetEditor.onMove(1)}>
						{plannerText('panel.move_down')}
					</Button>
					<Button variant="outline-secondary" size="xs" onClick={targetEditor.onCopy}>
						{plannerText('menu.duplicate')}
					</Button>
				</div>
				<PlannerAddStats
					target={targetEditor.target}
					identity={targetEditor.identity}
					idPrefix={targetEditor.idPrefix}
					threatMetrics={targetEditor.threatMetrics}
					onModify={targetEditor.onModify}
				/>
			</section>
		);
	}

	const event = isEventRef(selected) ? getEvent(plan, selected) : undefined;

	if (!isEventRef(selected) || !event)
		return (
			<section ref={ref} tabIndex={-1} className="flex flex-col gap-2 outline-none">
				<span className="text-sm text-foreground-disabled">{plannerText('panel.empty')}</span>
			</section>
		);

	const context: PlanFieldContext = {
		baseSeconds,
		targetNames,
		waveNames: plan.addWaves.map((_, index) => eventName(plan, { kind: 'wave', index })),
	};
	const name = eventName(plan, selected);
	const prefix = `encounter-planner-${selected.kind}-${selected.index}`;
	const commit = (edit: (event: PlanEvent) => void) => onChange(editEvent(plan, selected, edit));

	const remove = () => {
		const next = EncounterPlan.clone(plan);
		removeEvent(next, selected);
		onSelect(undefined);
		onChange(next);
	};

	const renderField = (field: PlanFieldConfig<PlanEvent>) => {
		const id = `${prefix}-${field.id}`;
		const frame = { id, label: field.label, labelTooltip: field.tooltip };
		switch (field.type) {
			case 'number':
				return (
					<NumberField
						key={id}
						{...frame}
						float={field.float}
						positive
						value={field.get(event, context)}
						onChange={value => commit(e => field.set(e, value, context))}
					/>
				);
			case 'boolean':
				return <BooleanPicker key={id} {...frame} className="self-end" value={field.get(event)} onChange={value => commit(e => field.set(e, value))} />;
			case 'text':
				return <StringField key={id} {...frame} value={field.get(event)} onChange={value => commit(e => field.set(e, value))} />;
			case 'choice':
				return (
					<ChoiceField
						key={id}
						{...frame}
						options={field.options(context)}
						value={field.get(event)}
						onChange={value => commit(e => field.set(e, value))}
					/>
				);
			case 'list':
				return (
					<MultiComboBox
						key={id}
						{...frame}
						className="col-span-full"
						values={field.options(context)}
						value={field.get(event)}
						onChange={value => commit(e => field.set(e, value))}
					/>
				);
		}
	};

	const fields = planFieldConfigs(selected.kind) as Array<PlanFieldConfig<PlanEvent>>;
	const shown = fields.filter(field => field.show?.(event) ?? true);
	// A wave's adds follow its name; every other kind is one grid.
	const templateEnd = selected.kind === 'wave' ? shown.findIndex(field => field.id === 'name') + 1 : shown.length;

	return (
		<section ref={ref} tabIndex={-1} aria-label={name} className="flex flex-col gap-2 outline-none">
			<div className="flex items-center justify-between gap-2">
				<span className="text-base font-bold text-white">{name}</span>
				<Button variant="link-danger" size="sm" onClick={remove}>
					{plannerText('panel.remove')}
					<Icon name="xmark" />
				</Button>
			</div>
			<FieldGrid>{shown.slice(0, templateEnd).map(renderField)}</FieldGrid>
			{selected.kind === 'wave' && (
				<PlannerAdds
					key={prefix}
					idPrefix={`${prefix}-add`}
					adds={(event as AddWave).adds}
					presetTargets={presetTargets}
					newAdd={newAdd}
					onEdit={edit => commit(e => edit((e as AddWave).adds))}
				/>
			)}
			<FieldGrid>{shown.slice(templateEnd).map(renderField)}</FieldGrid>
		</section>
	);
};

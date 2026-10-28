import type { AddWave, PresetTarget } from '@generated/proto/common';
import { EncounterDesign, Target } from '@generated/proto/common';
import { Accordion, AccordionItem } from '@ui-kit/Accordion';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { FieldGrid } from '@ui-kit/FieldGrid';
import { Icon } from '@ui-kit/Icon';
import { MultiComboBox } from '@ui-kit/MultiComboBox';
import { NumberField, StringField } from '@ui-kit/TextField';
import { type Ref, useState } from 'react';

import type { DesignEvent, DesignEventRef } from '../../../model/design/lanes';
import { getEvent, removeEvent } from '../../../model/design/lanes';
import { DesignerAddStats } from '../DesignerAddStats';
import type { DesignFieldConfig, DesignFieldContext } from '../utils/configs';
import { designerText, designFieldConfigs, editEvent, eventName } from '../utils/configs';

export interface DesignerEventPanelProps {
	design: EncounterDesign;
	selected?: DesignEventRef;
	baseSeconds: number;
	/** The encounter's targets by index, for the target pickers. */
	targetNames: ReadonlyArray<string>;
	/** The preset targets a wave can copy as its add. */
	presetTargets: ReadonlyArray<PresetTarget>;
	onChange: (design: EncounterDesign) => void;
	onSelect: (ref: DesignEventRef | undefined) => void;
	/** The panel takes focus from script only, for a container that moves focus here when a bar is chosen. */
	ref?: Ref<HTMLElement>;
}

export const DesignerEventPanel = ({ design, selected, baseSeconds, targetNames, presetTargets, onChange, onSelect, ref }: DesignerEventPanelProps) => {
	const event = selected && getEvent(design, selected);
	const [sections, setSections] = useState<string[]>([]);

	if (!selected || !event)
		return (
			<section ref={ref} tabIndex={-1} className="flex flex-col gap-2 outline-none">
				<span className="text-sm text-foreground-disabled">{designerText('panel.empty')}</span>
			</section>
		);

	const context: DesignFieldContext = {
		baseSeconds,
		targetNames,
		presetTargets,
		waveNames: design.addWaves.map((_, index) => eventName(design, { kind: 'wave', index })),
	};
	const name = eventName(design, selected);
	const prefix = `encounter-designer-${selected.kind}-${selected.index}`;
	const commit = (edit: (event: DesignEvent) => void) => onChange(editEvent(design, selected, edit));

	const remove = () => {
		const next = EncounterDesign.clone(design);
		removeEvent(next, selected);
		onSelect(undefined);
		onChange(next);
	};

	const renderField = (field: DesignFieldConfig<DesignEvent>) => {
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
				return <BooleanPicker key={id} {...frame} value={field.get(event)} onChange={value => commit(e => field.set(e, value))} />;
			case 'text':
				return <StringField key={id} {...frame} value={field.get(event)} onChange={value => commit(e => field.set(e, value))} />;
			case 'choice':
				return (
					<ChoiceField
						key={id}
						{...frame}
						options={field.options(context)}
						value={field.get(event, context)}
						onChange={value => commit(e => field.set(e, value, context))}
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

	const fields = designFieldConfigs(selected.kind) as Array<DesignFieldConfig<DesignEvent>>;
	const shown = fields.filter(field => field.show?.(event) ?? true);
	// A wave's add stats follow its template, whose pick they refine; every other kind is one grid.
	const templateEnd = selected.kind === 'wave' ? shown.findIndex(field => field.id === 'template') + 1 : shown.length;

	return (
		<section ref={ref} tabIndex={-1} aria-label={name} className="flex flex-col gap-2 outline-none">
			<div className="flex items-center justify-between gap-2">
				<span className="text-base font-bold text-white">{name}</span>
				<Button variant="link-danger" size="sm" onClick={remove}>
					<Icon name="xmark" /> {designerText('panel.remove')}
				</Button>
			</div>
			<FieldGrid>{shown.slice(0, templateEnd).map(renderField)}</FieldGrid>
			{selected.kind === 'wave' && (
				<Accordion value={sections} onValueChange={setSections} density="compact">
					<AccordionItem value="add-stats" title={designerText('panel.add_stats')} testId="encounter-designer-add-stats">
						<DesignerAddStats
							idPrefix={`${prefix}-add`}
							target={(event as AddWave).target}
							onModify={edit => commit(e => edit(((e as AddWave).target ??= Target.create())))}
						/>
					</AccordionItem>
				</Accordion>
			)}
			<FieldGrid>{shown.slice(templateEnd).map(renderField)}</FieldGrid>
		</section>
	);
};

import { type PresetTarget, Target } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateMobType } from '@i18n/localization';
import { Accordion, AccordionItem } from '@ui-kit/Accordion';
import { Button } from '@ui-kit/Button';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { Icon } from '@ui-kit/Icon';
import { useState } from 'react';

import { DesignerAddStats } from '../DesignerAddStats';
import { designerText } from '../utils/configs';

export interface DesignerAddsProps {
	adds: ReadonlyArray<Target>;
	/** The preset targets an add can copy. */
	presetTargets: ReadonlyArray<PresetTarget>;
	/** A new add: the default template's copy. */
	newAdd: () => Target;
	/** Edits the wave's list of adds in the draft. */
	onEdit: (edit: (adds: Array<Target>) => void) => void;
	/** Starts the field ids, so two waves never share one. */
	idPrefix: string;
}

const CUSTOM = -1;

/** The adds one occurrence of a wave spawns: one row each, with its template and its own stats. */
export const DesignerAdds = ({ adds, presetTargets, newAdd, onEdit, idPrefix }: DesignerAddsProps) => {
	const [open, setOpen] = useState<string[]>([]);
	const templates = [{ value: CUSTOM, label: i18n.t('common.custom') }, ...presetTargets.map((preset, value) => ({ value, label: preset.path }))];

	return (
		<div className="flex flex-col gap-2" data-testid="encounter-designer-adds">
			<div className="flex items-center justify-between gap-2">
				<span className="text-sm font-bold text-white">{designerText('adds.title')}</span>
				<Button variant="outline-secondary" size="xs" aria-label={designerText('adds.add')} onClick={() => onEdit(list => void list.push(newAdd()))}>
					<Icon name="plus" /> {designerText('adds.add_short')}
				</Button>
			</div>
			{adds.map((add, index) => {
				const key = `${idPrefix}-${index}`;
				const found = presetTargets.findIndex(preset => !!preset.target && Target.equals(add, preset.target));
				return (
					<div key={key} className="flex flex-col gap-1 border-l-2 border-gray-700 pl-2">
						<div className="flex items-center justify-between gap-2">
							<span className="min-w-0 truncate text-sm text-white">
								{designerText('adds.add_row', {
									name: add.name || designerText('adds.add_title', { number: index + 1 }),
									level: add.level,
									type: translateMobType(add.mobType),
								})}
							</span>
							<span className="flex shrink-0 gap-1">
								<Button
									variant="link"
									size="sm"
									aria-label={designerText('adds.duplicate')}
									onClick={() => onEdit(list => list.splice(index + 1, 0, Target.clone(list[index])))}>
									<Icon name="copy" />
								</Button>
								<Button
									variant="link-danger"
									size="sm"
									aria-label={designerText('adds.remove')}
									onClick={() => onEdit(list => void list.splice(index, 1))}>
									<Icon name="xmark" />
								</Button>
							</span>
						</div>
						<Accordion
							value={open.filter(item => item === key)}
							onValueChange={value => setOpen(current => [...current.filter(item => item !== key), ...value])}
							density="compact">
							<AccordionItem value={key} title={designerText('panel.add_stats')} testId={`${key}-stats`}>
								<div className="flex flex-col gap-2">
									<ChoiceField<number>
										id={`${key}-template`}
										label={designerText('fields.template')}
										labelTooltip={designerText('tooltips.template')}
										options={templates}
										value={found}
										onChange={value => {
											const preset = presetTargets[value]?.target;
											if (preset) onEdit(list => void (list[index] = Target.clone(preset)));
										}}
									/>
									<DesignerAddStats idPrefix={key} target={add} onModify={edit => onEdit(list => edit(list[index]))} />
								</div>
							</AccordionItem>
						</Accordion>
					</div>
				);
			})}
		</div>
	);
};

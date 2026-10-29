import { type PresetTarget, Target } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateMobType } from '@i18n/localization';
import { Button } from '@ui-kit/Button';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { Dialog } from '@ui-kit/Dialog';
import { Icon } from '@ui-kit/Icon';
import { useState } from 'react';

import { DesignerAddStats } from '../DesignerAddStats';
import { designerText } from '../EncounterDesigner/utils/configs';

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

/** The adds one occurrence of a wave spawns: one row each, whose template and stats open in a dialog over the designer. */
export const DesignerAdds = ({ adds, presetTargets, newAdd, onEdit, idPrefix }: DesignerAddsProps) => {
	// The add stays picked while the dialog closes, so its fields fade out with it.
	const [picked, setPicked] = useState(0);
	const [open, setOpen] = useState(false);
	const templates = [{ value: CUSTOM, label: i18n.t('common.custom') }, ...presetTargets.map((preset, value) => ({ value, label: preset.path }))];
	const addName = (add: Target, index: number) => add.name || designerText('adds.add_title', { number: index + 1 });
	const edit = (index: number) => {
		setPicked(index);
		setOpen(true);
	};
	const edited = adds.at(picked);
	const key = `${idPrefix}-${picked}`;

	return (
		<div className="flex flex-col gap-2" data-testid="encounter-designer-adds">
			<div className="flex items-center justify-between gap-2">
				<span className="text-sm font-bold text-white">{designerText('adds.title')}</span>
				<Button
					variant="outline-secondary"
					size="xs"
					aria-label={designerText('adds.add')}
					onClick={() => {
						onEdit(list => void list.push(newAdd()));
						edit(adds.length);
					}}>
					<Icon name="plus" /> {designerText('adds.add_short')}
				</Button>
			</div>
			{adds.map((add, index) => (
				<div key={`${idPrefix}-${index}`} className="flex items-center justify-between gap-2 border-l-2 border-gray-700 pl-2">
					{/* The edit button is the keyboard's way in, so the label stays out of the tab order. */}
					<Button variant="unstyled" tabIndex={-1} className="min-w-0 truncate text-left text-sm text-white" onClick={() => edit(index)}>
						{designerText('adds.add_row', { name: addName(add, index), level: add.level, type: translateMobType(add.mobType) })}
					</Button>
					<span className="flex shrink-0 gap-1">
						<Button variant="link" size="sm" aria-label={designerText('adds.edit')} onClick={() => edit(index)}>
							<Icon name="pencil" />
						</Button>
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
			))}
			{/* Edits go straight into the designer's draft, whose own Save or Cancel decides them, so this dialog only closes. */}
			<Dialog
				open={open && !!edited}
				onOpenChange={setOpen}
				elevated
				size="md"
				scrollContents
				title={edited && addName(edited, picked)}
				testId="encounter-designer-add-dialog"
				footer={
					<Button variant="primary" onClick={() => setOpen(false)}>
						{designerText('adds.done')}
					</Button>
				}>
				{edited && (
					<div className="flex flex-col gap-2">
						<ChoiceField<number>
							id={`${key}-template`}
							label={designerText('fields.template')}
							labelTooltip={designerText('tooltips.template')}
							options={templates}
							value={presetTargets.findIndex(preset => !!preset.target && Target.equals(edited, preset.target))}
							onChange={value => {
								const preset = presetTargets[value]?.target;
								if (preset) onEdit(list => void (list[picked] = Target.clone(preset)));
							}}
						/>
						<DesignerAddStats idPrefix={key} target={edited} onModify={change => onEdit(list => change(list[picked]))} />
					</div>
				)}
			</Dialog>
		</div>
	);
};

import { BulkBuildBundle, BulkGearSetRef, BulkRotationRef, BulkSettingsRef, BulkTalentRef } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { Dialog } from '@ui-kit/Dialog';
import { Field } from '@ui-kit/Field';
import { Input } from '@ui-kit/FormControl';
import { useId, useState } from 'react';

import { bundleLabel, bundleNameTaken, type bundleOptions, CURRENT_REF } from '../../model/builds/bundles';
import { sameRef, type SavedRef } from '../../model/builds/saved_refs';

type PickerId = 'gearSet' | 'talents' | 'rotation' | 'settings';

export interface BulkBundleDialogViewProps {
	open: boolean;
	onOpenChange(open: boolean): void;
	title: string;
	/** `placeholder` is the label the picks give the bundle, which it keeps while the name is empty. */
	name: { value: string; placeholder: string; taken: boolean; onChange(name: string): void };
	pickers: ReadonlyArray<{
		id: PickerId;
		label: string;
		options: ReadonlyArray<{ value: SavedRef; label: string }>;
		value: SavedRef;
		onChange(ref: SavedRef): void;
	}>;
	canSave: boolean;
	onSave(): void;
}

export const BulkBundleDialogView = ({ open, onOpenChange, title, name, pickers, canSave, onSave }: BulkBundleDialogViewProps) => {
	const idPrefix = useId();
	const nameId = `${idPrefix}-name`;
	return (
		<Dialog
			open={open}
			onOpenChange={onOpenChange}
			size="md"
			title={title}
			testId="bulk-bundle-dialog"
			footer={
				<div className="flex gap-2">
					<Button variant="outline-cancel" onClick={() => onOpenChange(false)}>
						{i18n.t('common.cancel')}
					</Button>
					<Button variant="primary" disabled={!canSave} onClick={onSave} data-testid="bulk-bundle-save">
						{i18n.t('bulk_tab.builds.bundles.save')}
					</Button>
				</div>
			}>
			<div className="grid gap-4">
				<Field
					id={nameId}
					label={i18n.t('bulk_tab.builds.bundles.name')}
					description={name.taken ? i18n.t('bulk_tab.builds.bundles.name_taken') : undefined}>
					<Input
						id={nameId}
						type="text"
						placeholder={name.placeholder}
						aria-invalid={name.taken}
						value={name.value}
						onChange={event => name.onChange(event.target.value)}
						onKeyDown={event => {
							if (event.key === 'Enter' && canSave) onSave();
						}}
						data-testid="bulk-bundle-name"
					/>
				</Field>
				{pickers.map(picker => (
					<ChoiceField<SavedRef>
						key={picker.id}
						id={`${idPrefix}-${picker.id}`}
						label={picker.label}
						options={picker.options}
						value={picker.value}
						equals={sameRef}
						onChange={picker.onChange}
					/>
				))}
			</div>
		</Dialog>
	);
};

export interface BulkBundleDialogProps {
	open: boolean;
	/** Null for a new bundle. */
	bundle: BulkBuildBundle | null;
	options: ReturnType<typeof bundleOptions>;
	/** The labels the name may not take, the bundle's own left out. */
	takenLabels: readonly string[];
	onSave(bundle: BulkBuildBundle): void;
	onClose(): void;
}

const PICKERS: ReadonlyArray<{ id: PickerId; options: keyof BulkBundleDialogProps['options']; labelKey: string }> = [
	{ id: 'gearSet', options: 'gearSets', labelKey: 'gear_set' },
	{ id: 'talents', options: 'talents', labelKey: 'talents' },
	{ id: 'rotation', options: 'rotations', labelKey: 'rotation' },
	{ id: 'settings', options: 'settings', labelKey: 'settings' },
];

type Picks = Record<PickerId, SavedRef>;

// A ref no option holds starts at Current.
const startPick = (options: ReadonlyArray<{ ref: SavedRef }>, ref: SavedRef | undefined): SavedRef =>
	ref && options.some(option => sameRef(option.ref, ref)) ? ref : CURRENT_REF;

const optionLabel = (options: ReadonlyArray<{ ref: SavedRef; label: string }>, ref: SavedRef): string =>
	options.find(option => sameRef(option.ref, ref))?.label ?? '';

const bundleOf = ({ gearSet, talents, rotation, settings }: Picks, name: string): BulkBuildBundle =>
	BulkBuildBundle.create({
		gearSet: BulkGearSetRef.create({ name: gearSet.name, preset: gearSet.preset, group: gearSet.group ?? '' }),
		talents: BulkTalentRef.create({ name: talents.name, preset: talents.preset }),
		rotation: BulkRotationRef.create({ name: rotation.name, preset: rotation.preset }),
		settings: BulkSettingsRef.create({ name: settings.name, preset: settings.preset }),
		checked: true,
		name: name.trim(),
	});

// Save stays off while every pick is Current, since that bundle is the Current bundle's row, and while the name is another
// bundle's label.
export const BulkBundleDialog = ({ open, bundle, options, takenLabels, onSave, onClose }: BulkBundleDialogProps) => {
	const [name, setName] = useState(bundle?.name ?? '');
	const [picks, setPicks] = useState<Picks>(() => ({
		gearSet: startPick(options.gearSets, bundle?.gearSet),
		talents: startPick(options.talents, bundle?.talents),
		rotation: startPick(options.rotations, bundle?.rotation),
		settings: startPick(options.settings, bundle?.settings),
	}));
	const nameTaken = bundleNameTaken(name, takenLabels);

	return (
		<BulkBundleDialogView
			open={open}
			onOpenChange={next => !next && onClose()}
			title={i18n.t(bundle ? 'bulk_tab.builds.bundles.dialog_edit' : 'bulk_tab.builds.bundles.dialog_new')}
			name={{
				value: name,
				placeholder: bundleLabel({
					gear: optionLabel(options.gearSets, picks.gearSet),
					talents: optionLabel(options.talents, picks.talents),
					rotation: optionLabel(options.rotations, picks.rotation),
					settings: optionLabel(options.settings, picks.settings),
				}),
				taken: nameTaken,
				onChange: setName,
			}}
			pickers={PICKERS.map(picker => ({
				id: picker.id,
				label: i18n.t(`bulk_tab.builds.parts.${picker.labelKey}`),
				options: options[picker.options].map(({ ref, label }) => ({ value: ref, label })),
				value: picks[picker.id],
				onChange: ref => setPicks(current => ({ ...current, [picker.id]: ref })),
			}))}
			canSave={!nameTaken && Object.values(picks).some(ref => !sameRef(ref, CURRENT_REF))}
			onSave={() => {
				onSave(bundleOf(picks, name));
				onClose();
			}}
		/>
	);
};

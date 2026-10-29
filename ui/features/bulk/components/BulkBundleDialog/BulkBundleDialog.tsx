import { BulkBuildBundle, BulkGearSetRef, BulkRotationRef, BulkSettingsRef, BulkTalentRef } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { Dialog } from '@ui-kit/Dialog';
import { useId, useState } from 'react';

import { type bundleOptions, CURRENT_REF } from '../../model/bundles';
import { sameRef, type SavedRef } from '../../model/saved_refs';

type PickerId = 'gearSet' | 'talents' | 'rotation' | 'settings';

export interface BulkBundleDialogViewProps {
	open: boolean;
	onOpenChange(open: boolean): void;
	title: string;
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

export const BulkBundleDialogView = ({ open, onOpenChange, title, pickers, canSave, onSave }: BulkBundleDialogViewProps) => {
	const idPrefix = useId();
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

const bundleOf = ({ gearSet, talents, rotation, settings }: Picks): BulkBuildBundle =>
	BulkBuildBundle.create({
		gearSet: BulkGearSetRef.create({ name: gearSet.name, preset: gearSet.preset, group: gearSet.group ?? '' }),
		talents: BulkTalentRef.create({ name: talents.name, preset: talents.preset }),
		rotation: BulkRotationRef.create({ name: rotation.name, preset: rotation.preset }),
		settings: BulkSettingsRef.create({ name: settings.name, preset: settings.preset }),
		checked: true,
	});

// Save stays off while every pick is Current: that bundle is the Current bundle's row.
export const BulkBundleDialog = ({ open, bundle, options, onSave, onClose }: BulkBundleDialogProps) => {
	const [picks, setPicks] = useState<Picks>(() => ({
		gearSet: startPick(options.gearSets, bundle?.gearSet),
		talents: startPick(options.talents, bundle?.talents),
		rotation: startPick(options.rotations, bundle?.rotation),
		settings: startPick(options.settings, bundle?.settings),
	}));

	return (
		<BulkBundleDialogView
			open={open}
			onOpenChange={next => !next && onClose()}
			title={i18n.t(bundle ? 'bulk_tab.builds.bundles.dialog_edit' : 'bulk_tab.builds.bundles.dialog_new')}
			pickers={PICKERS.map(picker => ({
				id: picker.id,
				label: i18n.t(`bulk_tab.builds.parts.${picker.labelKey}`),
				options: options[picker.options].map(({ ref, label }) => ({ value: ref, label })),
				value: picks[picker.id],
				onChange: ref => setPicks(current => ({ ...current, [picker.id]: ref })),
			}))}
			canSave={Object.values(picks).some(ref => !sameRef(ref, CURRENT_REF))}
			onSave={() => {
				onSave(bundleOf(picks));
				onClose();
			}}
		/>
	);
};

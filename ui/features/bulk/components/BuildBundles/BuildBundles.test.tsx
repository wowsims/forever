import { EnchantPresetDialogView } from '@features/enchant-presets/components/EnchantPresets/EnchantPresetDialog';
import { BulkBuildBundle, BulkGearSetRef, BulkRotationRef, BulkSettingsRef, BulkTalentRef } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useRef, useState } from 'react';
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest';

import english from '../../../../../assets/locales/en/translation.json';
import { CURRENT_REF } from '../../model/builds/bundles';
import { BulkBundleDialog, type BulkBundleDialogProps } from '../BulkBundleDialog';
import type { BulkBundleListProps } from '../BulkBundleList';
import type { BulkCheckRow } from '../BulkCheckList';
import { BuildBundlesView } from './BuildBundles';

beforeAll(() => {
	i18n.addResourceBundle('en', 'translation', english, true, true);
});
afterAll(() => {
	i18n.removeResourceBundle('en', 'translation');
});

const ROWS: BulkBundleListProps['rows'] = [
	{ key: '0', label: 'Fury', checked: true },
	{ key: '1', label: 'Arms', checked: true },
];

const StatefulBundles = () => {
	const [rows, setRows] = useState(ROWS);
	return (
		<BuildBundlesView
			list={{
				current: { label: 'Current bundle', checked: true, onCheckedChange: vi.fn() },
				rows,
				onCheckedChange: vi.fn(),
				onEdit: vi.fn(),
				onDelete: key => setRows(current => current.filter(row => row.key !== key)),
				testIds: { list: 'bulk-bundles', row: 'bulk-bundle', current: 'bulk-bundle-current' },
			}}
			onNew={vi.fn()}
			dialog={null}
			enchantPicks={[]}
			onEnchantPickCheckedChange={vi.fn()}
			onNewEnchantPreset={vi.fn()}
			enchantPresetDialog={null}
		/>
	);
};

const AGILITY_CLOAK: BulkCheckRow = {
	ref: { name: 'Agility cloak', preset: false },
	name: 'Agility cloak',
	summary: '1 slot.',
	checked: false,
	duplicate: false,
};

// As `BuildBundles` wires it: the dialog opens from New preset, and its save lists the preset.
const FirstEnchantPreset = () => {
	const [picks, setPicks] = useState<readonly BulkCheckRow[]>([]);
	const [creating, setCreating] = useState(false);
	const enchantPicksRef = useRef<HTMLDivElement>(null);
	return (
		<BuildBundlesView
			list={{
				current: { label: 'Current bundle', checked: true, onCheckedChange: vi.fn() },
				rows: [],
				onCheckedChange: vi.fn(),
				onEdit: vi.fn(),
				onDelete: vi.fn(),
				testIds: { list: 'bulk-bundles', row: 'bulk-bundle', current: 'bulk-bundle-current' },
			}}
			onNew={vi.fn()}
			dialog={null}
			enchantPicks={picks}
			onEnchantPickCheckedChange={vi.fn()}
			onNewEnchantPreset={() => setCreating(true)}
			enchantPicksRef={enchantPicksRef}
			enchantPresetDialog={
				creating && (
					<EnchantPresetDialogView
						open
						onOpenChange={next => !next && setCreating(false)}
						finalFocus={enchantPicksRef}
						title="New enchant preset"
						name="Agility cloak"
						onNameChange={vi.fn()}
						nameTaken={false}
						canSave
						onSave={() => {
							setPicks([AGILITY_CLOAK]);
							setCreating(false);
						}}
						groups={[]}
						openSlots={[]}
						onOpenSlotsChange={vi.fn()}
						search=""
						onSearchChange={vi.fn()}
						showAll={false}
						onShowAllChange={vi.fn()}
					/>
				)
			}
		/>
	);
};

describe('BuildBundlesView', () => {
	it('moves focus to New bundle once a bundle is deleted, since its × goes with it', async () => {
		render(<StatefulBundles />);

		await userEvent.click(screen.getByRole('button', { name: 'Delete Arms' }));
		await userEvent.click(screen.getByRole('button', { name: 'Delete' }));

		expect(screen.getAllByTestId('bulk-bundle')).toHaveLength(1);
		expect(document.activeElement).toBe(screen.getByTestId('bulk-bundle-new'));
	});

	it('moves focus to the new enchant preset once the first is saved, since New preset goes with it, and back to New preset on a cancel', async () => {
		render(<FirstEnchantPreset />);
		const newPreset = () => screen.getByRole('button', { name: 'New preset' });

		await userEvent.click(newPreset());
		await userEvent.click(screen.getByRole('button', { name: 'Cancel' }));
		await waitFor(() => expect(document.activeElement).toBe(newPreset()));

		await userEvent.click(newPreset());
		await userEvent.click(screen.getByTestId('bulk-enchant-preset-save'));
		expect(screen.queryByTestId('bulk-enchant-pick-new')).toBeNull();
		await waitFor(() => expect(document.activeElement).toBe(screen.getByRole('checkbox', { name: 'Agility cloak' })));
	});
});

const ref = (name: string, preset = true) => ({ name, preset });

const OPTIONS: BulkBundleDialogProps['options'] = {
	gearSets: [
		{ ref: CURRENT_REF, label: 'Current gear' },
		{ ref: { name: 'Pre-BiS', preset: true, group: 'Arms' }, label: 'Pre-BiS (Arms)' },
	],
	talents: [
		{ ref: CURRENT_REF, label: 'Current talents' },
		{ ref: ref('Arms'), label: 'Arms' },
	],
	rotations: [
		{ ref: CURRENT_REF, label: 'Current rotation' },
		{ ref: ref('Arms APL'), label: 'Arms APL' },
	],
	settings: [
		{ ref: CURRENT_REF, label: 'Current settings' },
		{ ref: ref('Pre-BiS - Arms'), label: 'Pre-BiS - Arms' },
	],
};

const ARMS_BUNDLE = BulkBuildBundle.create({
	gearSet: BulkGearSetRef.create({ name: 'Pre-BiS', preset: true, group: 'Arms' }),
	talents: BulkTalentRef.create(ref('Arms')),
	rotation: BulkRotationRef.create(ref('Arms APL')),
	settings: BulkSettingsRef.create(ref('Pre-BiS - Arms')),
	checked: true,
	name: 'Arms raid',
});

const renderDialog = (props: Partial<BulkBundleDialogProps> = {}) => {
	const onSave = vi.fn();
	render(
		<BulkBundleDialog open bundle={null} options={OPTIONS} takenLabels={['Current bundle', 'Fury raid']} onSave={onSave} onClose={vi.fn()} {...props} />,
	);
	return onSave;
};

const nameField = () => screen.getByRole('textbox', { name: 'Name' });

describe('BulkBundleDialog names', () => {
	it('shows the label the picks give the bundle as the name’s placeholder, following the picks', async () => {
		renderDialog();

		expect(nameField().getAttribute('placeholder')).toBe('Current gear · Current talents · Current rotation · Current settings');
		await userEvent.click(screen.getByRole('combobox', { name: 'Talents' }));
		await userEvent.click(await screen.findByRole('option', { name: 'Arms' }));
		expect(nameField().getAttribute('placeholder')).toBe('Current gear · Arms · Current rotation · Current settings');
	});

	it('starts an edit from the bundle’s name', () => {
		renderDialog({ bundle: ARMS_BUNDLE });

		expect((nameField() as HTMLInputElement).value).toBe('Arms raid');
		expect(screen.getByTestId('bulk-bundle-save').hasAttribute('disabled')).toBe(false);
	});

	it('keeps Save off while the name is another bundle’s label, whatever its case and spaces, and says so', async () => {
		renderDialog({ bundle: ARMS_BUNDLE });
		const save = screen.getByTestId('bulk-bundle-save');

		await userEvent.clear(nameField());
		await userEvent.type(nameField(), ' fury RAID ');

		expect(save.hasAttribute('disabled')).toBe(true);
		expect(nameField().getAttribute('aria-invalid')).toBe('true');
		expect(screen.getByText('Name already used')).toBeTruthy();

		await userEvent.type(nameField(), '2');

		expect(save.hasAttribute('disabled')).toBe(false);
		expect(screen.queryByText('Name already used')).toBeNull();
	});

	it('saves the name trimmed on Enter', async () => {
		const onSave = renderDialog({ bundle: ARMS_BUNDLE });

		await userEvent.clear(nameField());
		await userEvent.type(nameField(), '  Arms main  {Enter}');

		expect(onSave).toHaveBeenCalledWith(expect.objectContaining({ name: 'Arms main' }));
	});
});

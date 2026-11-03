import { EnchantPresetDialogView } from '@features/enchant-presets/components/EnchantPresets/EnchantPresetDialog';
import i18n from '@i18n/config';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useRef, useState } from 'react';
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest';

import english from '../../../../../assets/locales/en/translation.json';
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

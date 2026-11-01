import { encodeSavedData } from '@features/saved-data-transfer/model/transfer_text';
import { EquipmentSpec } from '@generated/proto/common';
import { SavedGearSet } from '@generated/proto/ui';
import { Stats } from '@sim/proto/stats';
import { createSimStore, patchSlice } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));
vi.mock('@sim/state/subscriptions', async () => (await import('@testing/sim/subscriptions')).mockSubscriptions());
vi.mock('@sim/hooks/useSimReady', () => ({ useSimReady: () => true }));
vi.mock('../preset_build_state', () => ({ isBuildActive: () => false }));

const { GearPresets } = await import('./GearPresets');

const setup = (presets: { gear: Array<Record<string, unknown>>; builds: Array<Record<string, unknown>> }, powerTools = false) => {
	const store = createSimStore();
	patchSlice(store, 'ui', { showPowerTools: powerTools });
	return renderWithHost(<GearPresets />, {
		player: { getGear: () => ({ asSpec: () => EquipmentSpec.create() }), getBonusStats: () => new Stats() },
		sim: { store, db: { lookupEquipmentSpec: (spec: EquipmentSpec) => ({ asSpec: () => spec }) } },
		individualConfig: { presets },
		getStorageKey: (part: string) => `spec${part}`,
		getSavedGearStorageKey: () => 'specsavedGearSets',
	});
};

const saveInputs = (container: HTMLElement) => container.querySelectorAll('[data-testid="saved-data-save-input"]');

describe('GearPresets', () => {
	// A production build drops the dev-only gearsets, which leaves grouped builds with no gear.
	it('keeps the saved gear sets when grouped builds carry no gear and no gear presets remain', () => {
		const { container } = setup({ gear: [], builds: [{ name: 'Pre-BiS', group: 'Fury', talents: {} }] });

		expect(container.querySelector('[data-testid="preset-group-picker"]')).not.toBeNull();
		expect(saveInputs(container)).toHaveLength(1);
	});

	it('puts the transfer cog in the grouped layout’s untitled saved panel header', () => {
		const { container } = setup({ gear: [], builds: [{ name: 'Pre-BiS', group: 'Fury', talents: {} }] }, true);

		const cog = container.querySelector('[data-testid="saved-data-transfer-button"]');
		expect(cog?.closest('[data-testid="content-block-header"]')?.querySelector('[data-testid="content-block-title"]')?.textContent).toBe('');
	});

	it('steps an import around the gear presets, which its saved panel does not list', async () => {
		window.localStorage.clear();
		setup({ gear: [{ name: 'BiS', group: 'Fury', gear: EquipmentSpec.create() }], builds: [] }, true);

		fireEvent.click(screen.getByTestId('saved-data-transfer-button'));
		fireEvent.click(await screen.findByTestId('saved-data-transfer-import'));
		const text = await encodeSavedData({ type: 'gear', spec: 'SpecDpsWarrior', entries: { BiS: SavedGearSet.toJson(SavedGearSet.create()) } });
		await act(async () => {
			fireEvent.change(await screen.findByTestId('importer-textarea'), { target: { value: text } });
		});
		await screen.findByTestId('saved-data-tick-list');
		await act(async () => {
			fireEvent.click(screen.getByTestId('import-button'));
		});

		await waitFor(() => expect(Object.keys(JSON.parse(window.localStorage.getItem('specsavedGearSets')!))).toEqual(['BiS (2)']));
	});
});

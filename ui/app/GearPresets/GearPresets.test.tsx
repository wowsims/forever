import { EquipmentSpec } from '@generated/proto/common';
import { Stats } from '@sim/proto/stats';
import { renderWithHost } from '@testing/react';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));
vi.mock('@sim/state/subscriptions', async () => (await import('@testing/sim/subscriptions')).mockSubscriptions());
vi.mock('@sim/hooks/useSimReady', () => ({ useSimReady: () => true }));
vi.mock('../preset_build_state', () => ({ isBuildActive: () => false }));

const { GearPresets } = await import('./GearPresets');

const setup = (presets: { gear: Array<Record<string, unknown>>; builds: Array<Record<string, unknown>> }) =>
	renderWithHost(<GearPresets />, {
		player: { getGear: () => ({ asSpec: () => EquipmentSpec.create() }), getBonusStats: () => new Stats() },
		sim: { db: { lookupEquipmentSpec: (spec: EquipmentSpec) => ({ asSpec: () => spec }) } },
		individualConfig: { presets },
		getStorageKey: (part: string) => `spec${part}`,
		getSavedGearStorageKey: () => 'specsavedGearSets',
	});

const saveInputs = (container: HTMLElement) => container.querySelectorAll('[data-testid="saved-data-save-input"]');

describe('GearPresets', () => {
	// A production build drops the dev-only gearsets, which leaves grouped builds with no gear.
	it('keeps the saved gear sets when grouped builds carry no gear and no gear presets remain', () => {
		const { container } = setup({ gear: [], builds: [{ name: 'Pre-BiS', group: 'Fury', talents: {} }] });

		expect(container.querySelector('[data-testid="preset-group-picker"]')).not.toBeNull();
		expect(saveInputs(container)).toHaveLength(1);
	});
});

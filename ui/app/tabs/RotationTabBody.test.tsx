import { APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { renderWithHost } from '@testing/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

// Everything the panes mount besides the Cooldowns block and the sidebar's SIML actions; none of it is under test here.
vi.mock('@features/apl/components/AplNavbar', () => ({ AplNavbar: () => null }));
vi.mock('@features/apl/components/GroupList', () => ({ GroupList: () => null }));
vi.mock('@features/apl/components/PrePullList', () => ({ PrePullList: () => null }));
vi.mock('@features/apl/components/PriorityList', () => ({ PriorityList: () => null }));
vi.mock('@features/apl/components/PriorityPane', () => ({ PriorityPane: () => null }));
vi.mock('@features/apl/components/RotationTypePicker', () => ({ RotationTypePicker: () => null }));
vi.mock('@features/apl/components/SavedRotation', () => ({ SavedRotation: () => null }));
vi.mock('@features/apl/components/SimpleRotationInputs', () => ({ SimpleRotationInputs: () => null }));
vi.mock('@features/apl/components/VariablesList', () => ({ VariablesList: () => null }));
vi.mock('../PresetConfigurationPicker', () => ({ PresetConfigurationPicker: () => null }));
vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));
vi.mock('@features/settings', () => ({
	CooldownsPicker: () => null,
	useAvailableCooldowns: () => available,
}));
vi.mock('@features/import-export', () => ({ SimlRotationActions: () => <div data-testid="siml-rotation" /> }));

let available: Array<unknown> = [];
const { RotationTabBody } = await import('./RotationTabBody');

const setup = (hideSimpleCooldowns?: boolean, rotationType = APLRotationType.TypeSimple) =>
	renderWithHost(<RotationTabBody rotationType={rotationType} />, {
		player: { hasSimpleRotationGenerator: () => true },
		individualConfig: { rotationInputs: { inputs: [] }, hideSimpleCooldowns },
	});

const cooldownBlock = (container: HTMLElement) => container.querySelector('[data-testid="cooldown-settings"]');

beforeEach(() => {
	available = [{}, {}];
});

describe('RotationTabBody simple pane', () => {
	it('shows the Cooldowns block when the spec has available cooldowns', () => {
		expect(cooldownBlock(setup().container)).not.toBeNull();
	});

	// DPS warrior's `simpleRotation` generator never reads its Cooldowns argument, so the picker
	// could not affect the APL it builds. Vanilla hid the block with a per-spec SCSS rule.
	it('hides it for a spec that opts out, even with cooldowns available', () => {
		expect(cooldownBlock(setup(true).container)).toBeNull();
	});

	it('hides it when the spec has no available cooldowns', () => {
		available = [];
		expect(cooldownBlock(setup().container)).toBeNull();
	});
});

describe('RotationTabBody sidebar', () => {
	it.each([APLRotationType.TypeAuto, APLRotationType.TypeSimple, APLRotationType.TypeAPL])('holds the SIML actions beside rotation type %s', type => {
		expect(setup(false, type).container.querySelector('[data-testid="siml-rotation"]')).not.toBeNull();
	});
});

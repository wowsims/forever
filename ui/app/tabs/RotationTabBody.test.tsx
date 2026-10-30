import { requestAplReveal } from '@features/apl/model/reveal';
import { APLListItem, APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { renderWithHost } from '@testing/react';
import { act, waitFor } from '@testing-library/react';
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

describe('RotationTabBody APL sub-tabs', () => {
	const rotation = APLRotation.create({
		priorityList: [APLListItem.create()],
		groups: [{ name: 'execute', actions: [APLListItem.create()], variables: [] }],
	});

	const panes = (container: HTMLElement) => [...container.querySelectorAll<HTMLElement>('[data-testid="tab-pane"]')];
	const openPane = (container: HTMLElement) => panes(container).find(pane => !pane.hidden)?.id;

	it('opens Action Groups for a group entry and Priority List for a priority entry', async () => {
		const { container } = renderWithHost(<RotationTabBody rotationType={APLRotationType.TypeAPL} />, {
			player: { aplRotation: rotation, hasSimpleRotationGenerator: () => false },
		});
		expect(openPane(container)).toBe('apl-priority-list');

		act(() => {
			requestAplReveal(rotation, [{ kind: 'group', group: 0, index: 0 }]);
		});
		await waitFor(() => expect(openPane(container)).toBe('apl-action-groups'));

		act(() => {
			requestAplReveal(rotation, [{ kind: 'priority', index: 0 }]);
		});
		await waitFor(() => expect(openPane(container)).toBe('apl-priority-list'));
	});
});

describe('RotationTabBody sidebar', () => {
	it.each([APLRotationType.TypeAuto, APLRotationType.TypeSimple, APLRotationType.TypeAPL])('holds the SIML actions beside rotation type %s', type => {
		expect(setup(false, type).container.querySelector('[data-testid="siml-rotation"]')).not.toBeNull();
	});
});

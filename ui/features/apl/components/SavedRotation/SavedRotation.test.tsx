import { APLRotation, APLRotation_Type } from '@generated/proto/apl';
import { SavedRotation as SavedRotationProto } from '@generated/proto/ui';
import { createSimStore } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { makeAplRotation } from '@testing/storybook/fixtures/protos';
import { act, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { SavedRotation } from './SavedRotation';

vi.mock('@sim/state/subscriptions', async () => (await import('@testing/sim/subscriptions')).mockSubscriptions());

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const trackEvent = vi.hoisted(() => vi.fn());
vi.mock('../../../../tracking/utils', () => ({ trackEvent }));

const STORAGE_KEY = 'warrior-fury__savedRotation';

const storedJson = (rotation: APLRotation) => SavedRotationProto.toJson(SavedRotationProto.create({ rotation }));

let resolved: APLRotation;
let presets: Array<any>;
let waitForInit: () => Promise<void>;
let player: { getResolvedAplRotation: () => APLRotation; setAplRotation: ReturnType<typeof vi.fn> };

const hostParts = () => ({
	player,
	sim: { store: createSimStore(), waitForInit: () => waitForInit() },
	individualConfig: { presets: { rotations: presets } },
	getSavedRotationStorageKey: () => STORAGE_KEY,
});

const renderPanel = async () => {
	const result = renderWithHost(<SavedRotation />, hostParts());
	await act(async () => {});
	return result;
};

const chipNamed = (name: string) => [...document.querySelectorAll('[data-testid="saved-data-set-chip"]')].find(chip => chip.textContent?.startsWith(name))!;

beforeEach(() => {
	trackEvent.mockClear();
	window.localStorage.clear();
	resolved = makeAplRotation();
	presets = [];
	waitForInit = () => Promise.resolve();
	player = { getResolvedAplRotation: () => resolved, setAplRotation: vi.fn() };
});

describe('SavedRotation', () => {
	it('marks a saved APL rotation active against the same list resolved as Auto, which serialises differently', async () => {
		resolved = makeAplRotation({ type: APLRotation_Type.TypeAuto });
		window.localStorage.setItem(STORAGE_KEY, JSON.stringify({ Mine: storedJson(makeAplRotation()) }));
		await renderPanel();

		expect(chipNamed('Mine').hasAttribute('data-active')).toBe(true);
	});

	it('restores a stored rotation through setAplRotation and reports it', async () => {
		const stored = makeAplRotation({ priorityList: makeAplRotation().priorityList.slice(0, 1) });
		window.localStorage.setItem(STORAGE_KEY, JSON.stringify({ Mine: storedJson(stored) }));
		await renderPanel();

		fireEvent.click(chipNamed('Mine').querySelector('[data-testid="saved-data-set-name"]')!);

		expect(player.setAplRotation).toHaveBeenCalledWith(stored);
		expect(trackEvent).toHaveBeenCalledWith({ action: 'settings', category: 'load', label: 'rotation_tab.saved_rotations.label' });
	});

	it("loads a preset with no rotation as an empty one and runs the preset's onLoad", async () => {
		const onLoad = vi.fn();
		presets = [{ name: 'Default', rotation: SavedRotationProto.create(), onLoad }];
		await renderPanel();

		fireEvent.click(chipNamed('Default').querySelector('[data-testid="saved-data-set-name"]')!);

		expect(player.setAplRotation).toHaveBeenCalledWith(APLRotation.create());
		expect(onLoad).toHaveBeenCalledWith(player);
	});
});

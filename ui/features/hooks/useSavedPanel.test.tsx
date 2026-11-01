import { SimHostProvider } from '@sim/context/SimHostContext';
import type { IndividualSimHost } from '@sim/sim_host';
import { fakeHost } from '@testing/sim/fake_host';
import { act, renderHook } from '@testing-library/react';
import type { ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { type SavedPanelOptions, useSavedPanel } from './useSavedPanel';

const trackEvent = vi.hoisted(() => vi.fn());
vi.mock('../../tracking/utils', () => ({ trackEvent }));

interface Data {
	value: number;
}

const SPEC_PRESETS = { epWeights: [] };
const ENTRY = { name: 'Mine', data: { value: 2 }, json: '{"value":2}' };

let host: IndividualSimHost<any>;
let resolveInit: () => void;

beforeEach(() => {
	trackEvent.mockClear();
	// One host per test: the ready status restarts whenever the sim it reads changes.
	host = fakeHost({
		sim: { waitForInit: () => new Promise<void>(resolve => (resolveInit = resolve)) },
		individualConfig: { presets: SPEC_PRESETS },
	});
});

const options = (overrides: Partial<SavedPanelOptions<Data>> = {}): SavedPanelOptions<Data> => ({
	label: 'Things',
	storage: {
		entries: [ENTRY],
		storedNames: [ENTRY.name],
		save: vi.fn(),
		saveMany: vi.fn(),
		remove: vi.fn(),
		codec: { toJson: data => ({ ...data }), fromJson: json => json },
	},
	current: { value: 1 },
	serialize: data => JSON.stringify(data),
	load: vi.fn(),
	...overrides,
});

const renderPanel = (opts: SavedPanelOptions<Data>) =>
	renderHook(() => useSavedPanel(opts), {
		wrapper: ({ children }: { children: ReactNode }) => <SimHostProvider host={host}>{children}</SimHostProvider>,
	}).result;

describe('useSavedPanel', () => {
	it('holds the presets back until the sim is ready, and passes the saved entries through at once', async () => {
		const preset = { name: 'Default', data: { value: 3 }, json: '{"value":3}', isPreset: true };
		const presets = vi.fn(() => [preset]);
		const opts = options({ presets });
		const result = renderPanel(opts);

		expect(result.current.presets).toEqual([]);
		expect(presets).not.toHaveBeenCalled();
		expect(result.current.userData).toBe(opts.storage.entries);

		await act(async () => resolveInit());

		expect(presets).toHaveBeenCalledWith(SPEC_PRESETS, host);
		expect(result.current.presets).toEqual([preset]);
	});

	it.each([
		[{ value: 1 }, '{"value":1}'],
		[null, ''],
	])('compares the entries against %j serialised as %j', (current, json) => {
		expect(renderPanel(options({ current })).current.currentJson).toBe(json);
	});

	it('saves the current value under the given name, and reports it', () => {
		const opts = options();
		renderPanel(opts).current.onSave('Raid');

		expect(opts.storage.save).toHaveBeenCalledWith('Raid', { value: 1 });
		expect(trackEvent).toHaveBeenCalledWith({ action: 'settings', category: 'save', label: 'Things' });
	});

	it('saves nothing, and reports nothing, while there is no current value', () => {
		const opts = options({ current: null });
		renderPanel(opts).current.onSave('Raid');

		expect(opts.storage.save).not.toHaveBeenCalled();
		expect(trackEvent).not.toHaveBeenCalled();
	});

	it('loads an entry against the host, and reports it', () => {
		const opts = options();
		renderPanel(opts).current.onLoad(ENTRY);

		expect(opts.load).toHaveBeenCalledWith(ENTRY, host);
		expect(trackEvent).toHaveBeenCalledWith({ action: 'settings', category: 'load', label: 'Things' });
	});

	it('deletes an entry by its name, and reports it', () => {
		const opts = options();
		renderPanel(opts).current.onDelete(ENTRY);

		expect(opts.storage.remove).toHaveBeenCalledWith('Mine');
		expect(trackEvent).toHaveBeenCalledWith({ action: 'settings', category: 'delete', label: 'Things' });
	});
});

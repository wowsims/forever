import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { useNativeSimUpdate } from './useNativeSimUpdate';

vi.mock('@i18n/config', () => ({ default: { t: (key: string, options?: { error?: string }) => `${key}${options?.error ? `:${options.error}` : ''}` } }));

const info = { version: 'v1.0.0', latest: 'v1.1.0', canUpdate: true };
const json = (body: unknown, status = 200) => new Response(JSON.stringify(body), { status });

describe('useNativeSimUpdate', () => {
	const reload = vi.fn();

	beforeEach(() => {
		vi.useFakeTimers();
		vi.stubGlobal('location', { reload });
	});

	afterEach(() => {
		vi.useRealTimers();
		vi.unstubAllGlobals();
		reload.mockReset();
	});

	it('polls through the restart and reloads once the version changes', async () => {
		const versions = [new Error('down'), json({ version: 'v1.0.0' }), json({ version: 'v1.1.0' })];
		vi.stubGlobal(
			'fetch',
			vi.fn((url: string) => {
				if (url === '/update') return Promise.resolve(json({ version: 'v1.1.0' }, 202));
				const next = versions.shift();
				return next instanceof Error ? Promise.reject(next) : Promise.resolve(next);
			}),
		);
		const { result } = renderHook(() => useNativeSimUpdate(info));

		await act(async () => {
			void result.current.update();
			await vi.advanceTimersByTimeAsync(5000);
		});

		expect(reload).toHaveBeenCalledTimes(1);
	});

	it('shows the server error and does not poll', async () => {
		const fetchMock = vi.fn(() => Promise.resolve(json({ error: 'no newer release' }, 409)));
		vi.stubGlobal('fetch', fetchMock);
		const { result } = renderHook(() => useNativeSimUpdate(info));

		await act(async () => {
			await result.current.update();
		});

		expect(result.current.status).toEqual({ state: 'error', error: 'info.native_sim_update_failed:no newer release' });
		expect(fetchMock).toHaveBeenCalledTimes(1);
	});

	it('gives up after a minute', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn((url: string) => (url === '/update' ? Promise.resolve(json({}, 202)) : Promise.reject(new Error('down')))),
		);
		const { result } = renderHook(() => useNativeSimUpdate(info));

		await act(async () => {
			void result.current.update();
			await vi.advanceTimersByTimeAsync(61_000);
		});

		expect(result.current.status).toEqual({ state: 'error', error: 'info.native_sim_update_timeout' });
		expect(reload).not.toHaveBeenCalled();
	});
});

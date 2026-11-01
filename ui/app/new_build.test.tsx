import { withRelease } from '@testing/build_info';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { watchForNewBuild } from './new_build';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const toasts = vi.hoisted(() => [] as Array<{ title?: string }>);
vi.mock('@ui-kit/Toast', async importOriginal => ({
	...(await importOriginal<typeof import('@ui-kit/Toast')>()),
	toastManager: {
		add: (options: { title?: string }) => {
			toasts.push(options);
			return '';
		},
		close: () => {},
	},
}));

const OWN = 'a'.repeat(40);
const NEWER = 'b'.repeat(40);
const OPENED_AT = 'https://www.wowsims.com/forever/hunter/dps/#link';
const TEN_MINUTES = 10 * 60 * 1000;

let live = OWN;
const fetchVersion = vi.fn(async () => ({ ok: true, json: async () => ({ commit: live }) }));
const reload = vi.fn();
const replaceState = vi.fn();
// Each call adds listeners to the one document every test shares; they come off after each test.
const removeListeners: Array<() => void> = [];
const tracked = (target: EventTarget) => {
	const add = target.addEventListener.bind(target);
	vi.spyOn(target, 'addEventListener').mockImplementation((type, listener, options) => {
		add(type, listener, options);
		removeListeners.push(() => target.removeEventListener(type, listener, options));
	});
};
const settle = async () => {
	for (let i = 0; i < 10; i++) await Promise.resolve();
};
let restoreRelease: () => void;

beforeEach(() => {
	vi.useFakeTimers();
	vi.stubEnv('DEV', false);
	vi.stubGlobal('fetch', fetchVersion);
	vi.stubGlobal('location', { href: OPENED_AT, reload });
	vi.spyOn(history, 'replaceState').mockImplementation(replaceState);
	tracked(window);
	tracked(document);
	restoreRelease = withRelease({ date: '2026-11-01', commit: OWN, gameVersion: '', specs: {} });
	live = OWN;
});

afterEach(() => {
	removeListeners.splice(0).forEach(remove => remove());
	restoreRelease();
	sessionStorage.clear();
	toasts.length = 0;
	vi.clearAllMocks();
	vi.restoreAllMocks();
	vi.unstubAllGlobals();
	vi.unstubAllEnvs();
	vi.useRealTimers();
});

describe('watchForNewBuild', () => {
	it('reloads a page that opens on an older build at the URL it opened at, once per build', async () => {
		live = NEWER;
		watchForNewBuild();
		await settle();

		expect(fetchVersion).toHaveBeenCalledWith('/forever/version.json', { cache: 'no-store' });
		expect(replaceState).toHaveBeenCalledWith(history.state, '', OPENED_AT);
		expect(reload).toHaveBeenCalledTimes(1);

		// The server still answers with the old page after the reload.
		watchForNewBuild();
		await settle();
		expect(reload).toHaveBeenCalledTimes(1);
	});

	it('offers a reload once when a newer build is deployed while the page is open', async () => {
		watchForNewBuild();
		await settle();
		expect(reload).not.toHaveBeenCalled();

		await vi.advanceTimersByTimeAsync(TEN_MINUTES);
		await settle();
		expect(toasts).toEqual([]);

		live = NEWER;
		await vi.advanceTimersByTimeAsync(TEN_MINUTES);
		await settle();
		expect(toasts).toEqual([expect.objectContaining({ title: 'sim.new_build.title' })]);

		await vi.advanceTimersByTimeAsync(TEN_MINUTES);
		document.dispatchEvent(new Event('visibilitychange'));
		await settle();
		expect(toasts).toHaveLength(1);
		expect(reload).not.toHaveBeenCalled();
	});

	it('asks again at once when Back brings the page back from memory or a chunk fails to load', async () => {
		watchForNewBuild();
		await settle();
		fetchVersion.mockClear();

		window.dispatchEvent(Object.assign(new Event('pageshow'), { persisted: false }));
		await settle();
		expect(fetchVersion).not.toHaveBeenCalled();

		live = NEWER;
		window.dispatchEvent(Object.assign(new Event('pageshow'), { persisted: true }));
		await settle();
		expect(toasts).toHaveLength(1);

		window.dispatchEvent(new Event('vite:preloadError'));
		await settle();
		expect(toasts).toHaveLength(1);
	});

	it('leaves the dev server and a build that names no commit alone', async () => {
		vi.stubEnv('DEV', true);
		watchForNewBuild();
		vi.stubEnv('DEV', false);
		restoreRelease();
		restoreRelease = withRelease({ date: '', commit: '', gameVersion: '', specs: {} });
		watchForNewBuild();
		await settle();

		expect(fetchVersion).not.toHaveBeenCalled();
	});
});

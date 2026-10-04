import { renderWithHost } from '@testing/react';
import type { FakeHostParts } from '@testing/sim/fake_host';
import { act } from '@testing-library/react';
import type { ReactElement } from 'react';

/**
 * One subscription source for a Saved* panel test. `vi.mock` is hoisted above the imports, so the
 * factory reaches it lazily:
 *
 * ```ts
 * vi.mock('@sim/state/subscriptions', async () =>
 * 	(await import('@testing/sim/subscriptions')).mockSubscriptions((await import('@testing/saved_panel')).source.subscribe),
 * );
 * ```
 */
const listeners = new Set<() => void>();
export const source = {
	listeners,
	subscribe: (onChange: () => void) => {
		listeners.add(onChange);
		return () => listeners.delete(onChange);
	},
	notify: () => listeners.forEach(listener => listener()),
};

// The presets and the stored sets both land on `waitForInit`, so every assertion about a chip has to
// come after that microtask.
export const renderReady = async (ui: ReactElement, host: FakeHostParts) => {
	const result = renderWithHost(ui, host);
	await act(async () => {});
	return result;
};

export const nameInput = () => document.querySelector<HTMLInputElement>('[data-testid="saved-data-save-input"]')!;
export const saveButton = () => document.querySelector<HTMLButtonElement>('[data-testid="saved-data-save-button"]')!;
export const chips = (section: 'presets' | 'custom') => [
	...document.querySelectorAll(`[data-testid="saved-data-${section}"] [data-testid="saved-data-set-chip"]`),
];
export const chipNamed = (name: string) =>
	[...document.querySelectorAll('[data-testid="saved-data-set-chip"]')].find(chip => chip.textContent?.startsWith(name))!;
export const readStored = (key: string) => JSON.parse(window.localStorage.getItem(key) ?? 'null');

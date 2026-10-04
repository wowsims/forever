import { Race } from '@generated/proto/common';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { fakeHost } from '@testing/sim/fake_host';
import { renderHook } from '@testing-library/react';
import type { ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useSavedSettings } from './useSavedSettings';

vi.mock('@sim/state/subscriptions', async () => (await import('@testing/sim/subscriptions')).mockSubscriptions());

let key = '';

const store = (entries: Record<string, unknown>) => window.localStorage.setItem(key, JSON.stringify(entries));

const load = () =>
	renderHook(() => useSavedSettings(), {
		wrapper: ({ children }: { children: ReactNode }) => (
			<SimHostProvider host={fakeHost({ getSavedSettingsStorageKey: () => key })}>{children}</SimHostProvider>
		),
	}).result.current;

beforeEach(() => {
	// A fresh key per test: `useTypedLocalStorage` caches the raw string per key across a file.
	key = `tbc-test-savedSettings-${Math.random()}`;
});

describe('useSavedSettings', () => {
	it('reads a current entry', () => {
		store({ Raid: { race: 'RaceOrc' } });

		const { entries } = load();
		expect(entries.map(entry => entry.name)).toEqual(['Raid']);
		expect(entries[0].data.race).toBe(Race.RaceOrc);
	});

	it('reads every entry of the slot', () => {
		store({ Raid: { debuffs: { improvedSealOfTheCrusader: true } }, Current: { race: 'RaceOrc' } });

		expect(load().entries.map(entry => entry.name)).toEqual(['Raid', 'Current']);
	});
});

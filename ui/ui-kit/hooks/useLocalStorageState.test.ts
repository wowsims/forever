import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';

import { parseStringArray, useLocalStorageState } from './useLocalStorageState';

const KEY = 'useLocalStorageState.test';

afterEach(() => window.localStorage.clear());

describe('useLocalStorageState', () => {
	it.each([
		['nothing is stored', null, ['initial']],
		['the stored value fails to parse', JSON.stringify([1, 2]), ['initial']],
		['a value is stored', JSON.stringify(['stored']), ['stored']],
	])('starts from the initial value when %s, otherwise from storage', (_case, stored, expected) => {
		if (stored !== null) window.localStorage.setItem(KEY, stored);
		const { result } = renderHook(() => useLocalStorageState(KEY, parseStringArray, () => ['initial']));

		expect(result.current[0]).toEqual(expected);
	});

	it('moves the state and writes it through to storage', () => {
		const { result } = renderHook(() => useLocalStorageState(KEY, parseStringArray, () => ['initial']));

		act(() => result.current[1](['next']));

		expect(result.current[0]).toEqual(['next']);
		expect(JSON.parse(window.localStorage.getItem(KEY)!)).toEqual(['next']);
	});
});

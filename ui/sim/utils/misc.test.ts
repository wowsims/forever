import { describe, expect, it, vi } from 'vitest';

import { freeName, memoLast } from './misc';

describe('freeName', () => {
	it('keeps a free name and otherwise takes the first free number', () => {
		expect(freeName('Cooldowns', ['Opener'])).toBe('Cooldowns');
		expect(freeName('Cooldowns', ['Cooldowns', 'Cooldowns (2)'])).toBe('Cooldowns (3)');
		expect(freeName('Cooldowns', ['Cooldowns', 'Cooldowns (3)'])).toBe('Cooldowns (2)');
	});

	it('numbers a numbered name from its base, never nesting the suffix', () => {
		expect(freeName('Berserk (2)', ['Berserk', 'Berserk (2)'])).toBe('Berserk (3)');
	});

	it('takes the names as a set too', () => {
		expect(freeName('Cooldowns', new Set(['Cooldowns']))).toBe('Cooldowns (2)');
	});
});

describe('memoLast', () => {
	it('computes once while the deps keep their identities, and again when one changes', () => {
		const memo = memoLast<object>();
		const compute = vi.fn(() => ({}));
		const dep = {};
		const first = memo([dep, 1], compute);
		expect(memo([dep, 1], compute)).toBe(first);
		expect(compute).toHaveBeenCalledTimes(1);

		const second = memo([{}, 1], compute);
		expect(second).not.toBe(first);
		expect(memo([dep, 2], compute)).not.toBe(second);
		expect(compute).toHaveBeenCalledTimes(3);
	});

	it('holds one entry only', () => {
		const memo = memoLast<number>();
		const compute = vi.fn(() => compute.mock.calls.length);
		memo(['a'], compute);
		memo(['b'], compute);
		memo(['a'], compute);
		expect(compute).toHaveBeenCalledTimes(3);
	});
});

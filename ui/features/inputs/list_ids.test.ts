import { describe, expect, it } from 'vitest';

import { canDeleteAt, carryId, hasId, keyFor, moveItem, reconcileIds } from './list_ids';

describe('canDeleteAt', () => {
	it('allows every row when there is no minimum', () => {
		expect(canDeleteAt(0, undefined)).toBe(true);
		expect(canDeleteAt(0, 0)).toBe(true);
	});

	it('withholds delete from exactly the first minimumItems rows', () => {
		expect([0, 1, 2].map(index => canDeleteAt(index, 1))).toEqual([false, true, true]);
		expect([0, 1, 2].map(index => canDeleteAt(index, 2))).toEqual([false, false, true]);
	});
});

describe('moveItem', () => {
	const list = ['a', 'b', 'c', 'd'];

	it('moves forwards to exactly where the drop cue says', () => {
		expect(moveItem(list, 0, 2)).toEqual(['b', 'a', 'c', 'd']);
		expect(moveItem(list, 0, 3)).toEqual(['b', 'c', 'a', 'd']);
		expect(moveItem(list, 2, 0)).toEqual(['c', 'a', 'b', 'd']);
	});

	it('leaves the order unchanged for a destination on either side of the item itself', () => {
		expect(moveItem(list, 1, 1)).toEqual(list);
		expect(moveItem(list, 1, 2)).toEqual(list);
	});

	it('moves backwards to exactly the destination index', () => {
		expect(moveItem(list, 3, 1)).toEqual(['a', 'd', 'b', 'c']);
	});

	it('appends when the destination is past the end', () => {
		expect(moveItem(list, 0, 4)).toEqual(['b', 'c', 'd', 'a']);
	});

	it('leaves the source array alone', () => {
		moveItem(list, 0, 3);
		expect(list).toEqual(['a', 'b', 'c', 'd']);
	});
});

describe('keyFor', () => {
	it('mints one id per object and keeps returning it', () => {
		const item = { name: 'a' };
		const id = keyFor(item, 0);
		expect(keyFor(item, 5)).toBe(id);
	});

	it('mints different ids for different objects, including ones at the same index', () => {
		expect(keyFor({ name: 'a' }, 0)).not.toBe(keyFor({ name: 'b' }, 0));
	});

	it('falls back to a per-index id for a nullish item, never throwing', () => {
		expect(() => keyFor(undefined, 3)).not.toThrow();
		expect(keyFor(undefined, 3)).toBe(keyFor(undefined, 3));
		expect(keyFor(undefined, 3)).not.toBe(keyFor(undefined, 4));
	});

	it('never collides a nullish fallback id with a real object id', () => {
		const item = { name: 'a' };
		expect(keyFor(item, 0)).not.toBe(keyFor(undefined, 0));
	});
});

describe('carryId', () => {
	it('moves the old id onto the new object, so it resolves to the same key', () => {
		const oldItem = { name: 'a' };
		const newItem = { name: 'a2' };
		const id = keyFor(oldItem, 0);

		carryId(oldItem, newItem);

		expect(keyFor(newItem, 0)).toBe(id);
	});

	it('does nothing for a nullish old or new item', () => {
		expect(() => carryId(undefined, { name: 'a' })).not.toThrow();
		expect(() => carryId({ name: 'a' }, undefined)).not.toThrow();
	});
});

describe('hasId', () => {
	it('reports an object once keyFor has minted its id, and never a primitive', () => {
		const item = { name: 'a' };
		expect(hasId(item)).toBe(false);
		keyFor(item, 0);
		expect(hasId(item)).toBe(true);
		expect(hasId('a')).toBe(false);
		expect(hasId(null)).toBe(false);
	});
});

describe('reconcileIds', () => {
	it('carries each id onto the clone at the same index', () => {
		const prev = [{ name: 'a' }, { name: 'b' }];
		const ids = prev.map(keyFor);
		const clones = prev.map(item => ({ ...item }));

		reconcileIds(prev, clones);

		expect(clones.map(keyFor)).toEqual(ids);
	});

	it('leaves an object that is still in the list with its own id', () => {
		const a = { name: 'a' };
		const b = { name: 'b' };
		const idA = keyFor(a, 0);
		const idB = keyFor(b, 1);

		reconcileIds([a, b], [b, a]);

		expect([keyFor(b, 0), keyFor(a, 1)]).toEqual([idB, idA]);
	});

	it('mints a fresh id for an item past the end of the previous list', () => {
		const prev = [{ name: 'a' }];
		const idA = keyFor(prev[0], 0);
		const next = [{ name: 'a' }, { name: 'b' }];

		reconcileIds(prev, next);

		expect(keyFor(next[0], 0)).toBe(idA);
		expect(keyFor(next[1], 1)).not.toBe(idA);
	});

	it('does nothing without a previous list, and gives the same answer when run twice', () => {
		const prev = [{ name: 'a' }];
		const idA = keyFor(prev[0], 0);
		const clone = [{ name: 'a' }];

		reconcileIds(null, clone);
		reconcileIds(prev, clone);
		reconcileIds(prev, clone);

		expect(keyFor(clone[0], 0)).toBe(idA);
	});
});

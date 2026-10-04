import { renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { useOwnListSortable } from './useOwnListSortable';

const ref = (group: string, index: number) => ({ group, id: `${group}-${index}`, index, parentId: null });

describe('useOwnListSortable', () => {
	it('applies a move inside its own group and ignores the moves its nested lists report', () => {
		const move = vi.fn();
		const { result } = renderHook(() => useOwnListSortable(move, 'outer'));

		result.current.onMove({ from: ref('inner', 1), to: { group: 'inner', parentId: null, index: 0 } });
		result.current.onMove({ from: ref('outer', 2), to: { group: 'inner', parentId: null, index: 0 } });
		expect(move).not.toHaveBeenCalled();

		result.current.onMove({ from: ref('outer', 0), to: { group: 'outer', parentId: null, index: 2 } });
		expect(move).toHaveBeenCalledExactlyOnceWith(0, 2);
	});

	it('keeps one sortable object across renders and reads the latest move and group', () => {
		const first = vi.fn();
		const second = vi.fn();
		const { result, rerender } = renderHook(({ move, group }) => useOwnListSortable(move, group), { initialProps: { move: first, group: 'a' } });
		const sortable = result.current;

		rerender({ move: second, group: 'b' });
		expect(result.current).toBe(sortable);
		result.current.onMove({ from: ref('b', 0), to: { group: 'b', parentId: null, index: 2 } });
		expect(first).not.toHaveBeenCalled();
		expect(second).toHaveBeenCalledExactlyOnceWith(0, 2);
	});
});

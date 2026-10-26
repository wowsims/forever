import { act, renderHook } from '@testing-library/react';
import { toastManager } from '@ui-kit/Toast';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { useCopyToClipboard } from './useCopyToClipboard';

let copied: string[];

beforeEach(() => {
	vi.useFakeTimers();
	copied = [];
	vi.stubGlobal('navigator', {
		clipboard: {
			writeText: vi.fn((text: string) => {
				copied.push(text);
				return Promise.resolve();
			}),
		},
	});
});

afterEach(() => {
	vi.useRealTimers();
	vi.unstubAllGlobals();
});

describe('useCopyToClipboard', () => {
	it('reads the content getter at click time, not at mount', async () => {
		let payload = 'before';
		const getContent = vi.fn(() => payload);
		const { result } = renderHook(() => useCopyToClipboard(getContent));

		expect(getContent).not.toHaveBeenCalled();

		payload = 'after';
		await act(async () => result.current.copy());

		expect(getContent).toHaveBeenCalledTimes(1);
		expect(copied).toEqual(['after']);
	});

	it('runs the latest getter a re-render handed it', async () => {
		const { result, rerender } = renderHook(({ getContent }) => useCopyToClipboard(getContent), { initialProps: { getContent: () => 'first' } });

		rerender({ getContent: () => 'second' });
		await act(async () => result.current.copy());

		expect(copied).toEqual(['second']);
	});

	it('reports copied once the write lands, and reverts after 1500ms', async () => {
		const { result } = renderHook(() => useCopyToClipboard(() => 'payload'));
		expect(result.current.copied).toBe(false);

		await act(async () => result.current.copy());
		expect(result.current.copied).toBe(true);

		act(() => void vi.advanceTimersByTime(1499));
		expect(result.current.copied).toBe(true);

		act(() => void vi.advanceTimersByTime(1));
		expect(result.current.copied).toBe(false);
	});

	it('ignores clicks until the copied window ends, two in one task included', async () => {
		const getContent = vi.fn(() => 'payload');
		const { result } = renderHook(() => useCopyToClipboard(getContent));

		await act(async () => {
			result.current.copy();
			result.current.copy();
		});
		expect(getContent).toHaveBeenCalledTimes(1);
		expect(copied).toHaveLength(1);

		await act(async () => result.current.copy());
		expect(copied).toHaveLength(1);

		act(() => void vi.advanceTimersByTime(1500));
		await act(async () => result.current.copy());
		expect(copied).toHaveLength(2);
	});

	it('stays out of the copied window when the write is refused, so the next click tries again', async () => {
		const add = vi.spyOn(toastManager, 'add').mockReturnValue('toast');
		vi.stubGlobal('navigator', { clipboard: { writeText: vi.fn(() => Promise.reject(new Error('denied'))) } });
		const { result } = renderHook(() => useCopyToClipboard(() => 'payload'));

		await act(async () => result.current.copy());
		expect(result.current.copied).toBe(false);
		expect(add).toHaveBeenCalledTimes(1);

		await act(async () => result.current.copy());
		expect(add).toHaveBeenCalledTimes(2);
	});

	it('drops its revert timer on unmount', async () => {
		const { result, unmount } = renderHook(() => useCopyToClipboard(() => 'payload'));

		await act(async () => result.current.copy());
		expect(vi.getTimerCount()).toBe(1);

		unmount();
		expect(vi.getTimerCount()).toBe(0);
	});
});

import { act, fireEvent, render } from '@testing-library/react';
import { toastManager } from '@ui-kit/Toast';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { CopyButton } from './CopyButton';

let copied: string[];

const withClipboard = () =>
	vi.stubGlobal('navigator', {
		clipboard: {
			writeText: vi.fn((text: string) => {
				copied.push(text);
				return Promise.resolve();
			}),
		},
	});

beforeEach(() => {
	vi.useFakeTimers();
	copied = [];
	withClipboard();
});

afterEach(() => {
	vi.useRealTimers();
	vi.unstubAllGlobals();
	vi.restoreAllMocks();
});

const button = (container: HTMLElement) => container.querySelector('button')!;

describe('CopyButton', () => {
	it('carries the copy-button testid and is a plain button', () => {
		const { container } = render(<CopyButton getContent={() => 'x'} />);

		expect(button(container).getAttribute('data-testid')).toBe('copy-button');
		expect(button(container).getAttribute('type')).toBe('button');
	});

	it('copies the content and swaps to the copied label and icon', () => {
		const { container } = render(<CopyButton getContent={() => 'payload'} text="Copy me" />);

		expect(button(container).textContent).toContain('Copy me');
		expect(container.querySelector('svg')?.getAttribute('data-icon')).toBe('copy');

		act(() => void fireEvent.click(button(container)));

		expect(copied).toEqual(['payload']);
		expect(container.querySelector('svg')?.getAttribute('data-icon')).toBe('check');
		expect(button(container).textContent).not.toContain('Copy me');
	});

	it('shows the payload in a sticky error toast when there is no clipboard, and never enters the copied window', () => {
		const add = vi.spyOn(toastManager, 'add').mockReturnValue('toast');
		vi.stubGlobal('navigator', {});
		const { container } = render(<CopyButton getContent={() => 'payload'} text="Copy me" />);

		act(() => void fireEvent.click(button(container)));
		act(() => void fireEvent.click(button(container)));

		expect(add).toHaveBeenCalledTimes(2);
		const options = add.mock.calls[0][0];
		expect(options).toMatchObject({ variant: 'error', autohide: false });
		expect(render(<>{options.body}</>).container.textContent).toBe('payload');
		expect(button(container).textContent).toContain('Copy me');
		expect(button(container).querySelector('svg')?.getAttribute('data-icon')).toBe('copy');
	});

	it('copies nothing while disabled', () => {
		const { container } = render(<CopyButton getContent={() => 'payload'} disabled />);

		expect(button(container).disabled).toBe(true);
		act(() => void fireEvent.click(button(container)));
		expect(copied).toEqual([]);
	});

	it('anchors a tooltip only when it is given one', () => {
		const { container } = render(<CopyButton getContent={() => 'x'} />);
		expect(button(container).getAttribute('data-tooltip-id')).toBeNull();

		const withTooltip = render(<CopyButton getContent={() => 'x'} tooltip="Copies the whole sim" />);
		expect(button(withTooltip.container).getAttribute('data-tooltip-id')).toBeTruthy();
	});
});

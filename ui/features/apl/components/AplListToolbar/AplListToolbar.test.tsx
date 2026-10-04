import { FakeIntersectionObserver } from '@testing/apl/dom';
import { renderWithHost } from '@testing/react';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { AplListToolbar, AplListToolbarView, type AplListToolbarViewProps } from './AplListToolbar';

vi.mock('@i18n/config', () => ({ default: { t: (key: string, opts?: Record<string, unknown>) => (opts?.itemName ? `${key}:${opts.itemName}` : key) } }));

const newButton = () => screen.getByRole('button', { name: 'rotation_tab.apl.floatingActionBar.new:Action' });
const resetButton = () => screen.getByRole('button', { name: 'rotation_tab.apl.floatingActionBar.reset' });

describe('AplListToolbar', () => {
	it('resets through host.applyEmptyAplRotation() and creates through onCreate', () => {
		const applyEmptyAplRotation = vi.fn();
		const onCreate = vi.fn();
		renderWithHost(<AplListToolbar itemName="Action" onCreate={onCreate} />, { applyEmptyAplRotation });

		fireEvent.click(resetButton());
		expect(applyEmptyAplRotation).toHaveBeenCalledTimes(1);

		fireEvent.click(newButton());
		expect(onCreate).toHaveBeenCalledWith();
	});
});

const mount = (props: Partial<AplListToolbarViewProps> = {}) => {
	const scrollRoot = document.createElement('div');
	document.body.appendChild(scrollRoot);
	return render(<AplListToolbarView itemName="Action" onCreate={() => {}} onReset={() => {}} {...props} />, { container: scrollRoot });
};

const createButton = () => screen.getByRole('button', { name: 'rotation_tab.apl.nameModal.create' }) as HTMLButtonElement;

describe('AplListToolbarView', () => {
	beforeEach(() => {
		FakeIntersectionObserver.instances = [];
		vi.stubGlobal('IntersectionObserver', FakeIntersectionObserver);
	});

	afterEach(() => {
		vi.unstubAllGlobals();
	});

	it('labels the new button with the item name', () => {
		mount({ itemName: 'Variable' });

		expect(screen.getByRole('button', { name: 'rotation_tab.apl.floatingActionBar.new:Variable' })).toBeTruthy();
	});

	it('calls onCreate with no argument immediately when there is no nameDialog', () => {
		const onCreate = vi.fn();
		mount({ onCreate });

		fireEvent.click(newButton());
		expect(onCreate).toHaveBeenCalledWith();
		expect(screen.queryByRole('textbox')).toBeNull();
	});

	it('opens the dialog and calls onCreate with the submitted name, reading existingNames at open time', () => {
		const onCreate = vi.fn();
		let names = ['First'];
		mount({ onCreate, nameDialog: { inputLabel: 'Name', existingNames: () => names } });

		names = ['First', 'Second'];
		act(() => {
			fireEvent.click(newButton());
		});

		const input = screen.getByRole('textbox') as HTMLInputElement;
		act(() => {
			fireEvent.change(input, { target: { value: 'Second' } });
		});
		expect(createButton().disabled).toBe(true);

		act(() => {
			fireEvent.change(input, { target: { value: 'Third' } });
		});
		act(() => {
			fireEvent.click(createButton());
		});
		expect(onCreate).toHaveBeenCalledWith('Third');
	});

	it('calls onReset from the reset button', () => {
		const onReset = vi.fn();
		mount({ onReset });

		fireEvent.click(resetButton());
		expect(onReset).toHaveBeenCalledTimes(1);
	});

	it('sets the stuck state from the entry, not by toggling', () => {
		const { container } = mount();
		const root = container.querySelector('[data-testid="apl-floating-action-bar-root"]') as HTMLElement;
		expect(root.hasAttribute('data-stuck')).toBe(false);

		const observer = FakeIntersectionObserver.instances[0];
		act(() => observer.fire({ isIntersecting: false }));
		expect(root.hasAttribute('data-stuck')).toBe(true);

		act(() => observer.fire({ isIntersecting: false }));
		expect(root.hasAttribute('data-stuck')).toBe(true);

		act(() => observer.fire({ isIntersecting: true }));
		expect(root.hasAttribute('data-stuck')).toBe(false);
	});
});

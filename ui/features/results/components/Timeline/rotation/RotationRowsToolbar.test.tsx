import { castRow, rotationModel } from '@testing/timeline_rotation';
import { fireEvent, render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { RotationRowsToolbar } from './RotationRowsToolbar';

const MODEL = rotationModel(['Alpha', 'Beta', 'Gamma', 'Delta'].map((label, index) => castRow(`cast:${index}`, label, [])));

const mount = (hidden: Array<string> = []) => {
	const onToggle = vi.fn();
	const onShowAll = vi.fn();
	const view = render(<RotationRowsToolbar model={MODEL} hidden={new Set(hidden)} onToggle={onToggle} onShowAll={onShowAll} />);
	return { ...view, onToggle, onShowAll, toggle: view.container.querySelector<HTMLButtonElement>('[data-testid="rotation-floating-action-bar-toggle"]')! };
};

const chips = (container: HTMLElement) => [...container.querySelectorAll<HTMLButtonElement>('[data-testid="rotation-floating-action-bar-chip"]')];

describe('RotationRowsToolbar', () => {
	it('builds no chips until the drawer is opened', () => {
		const { container, toggle } = mount();
		expect(chips(container)).toHaveLength(0);

		fireEvent.click(toggle);
		expect(chips(document.body).map(chip => chip.textContent)).toEqual(['Alpha', 'Beta', 'Gamma', 'Delta']);
		expect(document.querySelector('[data-testid="rotation-floating-action-bar-group-title"]')!.textContent).toBe(
			'results_tab.details.timeline.sections.player',
		);
	});

	it('keeps the drawer open across a new result and rebuilds its chips', () => {
		const { toggle, rerender, onToggle, onShowAll } = mount();
		fireEvent.click(toggle);

		const next = rotationModel(['Epsilon', 'Zeta'].map((label, index) => castRow(`cast:${index}`, label, [])));
		rerender(<RotationRowsToolbar model={next} hidden={new Set()} onToggle={onToggle} onShowAll={onShowAll} />);

		expect(toggle.getAttribute('aria-expanded')).toBe('true');
		expect(chips(document.body).map(chip => chip.textContent)).toEqual(['Epsilon', 'Zeta']);
	});

	it('marks a hidden row’s chip off, and reports a click on it', () => {
		const { toggle, onToggle } = mount(['cast:1']);
		fireEvent.click(toggle);
		const [alpha, beta] = chips(document.body);
		expect(alpha.getAttribute('aria-checked')).toBe('true');
		expect(beta.getAttribute('aria-checked')).toBe('false');
		expect(beta.hasAttribute('data-active')).toBe(false);

		fireEvent.click(beta);
		expect(onToggle).toHaveBeenCalledWith('cast:1');
	});

	it('previews up to three hidden rows and marks the rest with an ellipsis', () => {
		expect(mount(['cast:0', 'cast:1']).container.querySelector('[data-testid="rotation-floating-action-bar-preview"]')!.textContent).toBe('Alpha, Beta');
		expect(
			mount(['cast:0', 'cast:1', 'cast:2', 'cast:3']).container.querySelector('[data-testid="rotation-floating-action-bar-preview"]')!.textContent,
		).toBe('Alpha, Beta, Gamma, …');
	});

	it('ignores hidden keys the current model does not have', () => {
		const { container } = mount(['cast:0', 'from-another-result']);
		expect(container.querySelector('[data-testid="rotation-floating-action-bar-preview"]')!.textContent).toBe('Alpha');
	});

	it('offers show-all only while something is hidden', () => {
		expect(mount().container.querySelector<HTMLElement>('[data-testid="rotation-floating-action-bar-show-all"]')!.hidden).toBe(true);
		const { container, onShowAll } = mount(['cast:0']);
		const showAll = container.querySelector<HTMLElement>('[data-testid="rotation-floating-action-bar-show-all"]')!;
		expect(showAll.hidden).toBe(false);
		fireEvent.click(showAll);
		expect(onShowAll).toHaveBeenCalled();
	});

	it('keeps one tab stop per group and walks it with the arrow keys, wrapping at both ends', () => {
		const { toggle } = mount();
		fireEvent.click(toggle);
		expect(chips(document.body).map(chip => chip.tabIndex)).toEqual([0, -1, -1, -1]);

		fireEvent.keyDown(chips(document.body)[0], { key: 'ArrowRight' });
		expect(document.activeElement).toBe(chips(document.body)[1]);
		expect(chips(document.body).map(chip => chip.tabIndex)).toEqual([-1, 0, -1, -1]);

		fireEvent.keyDown(chips(document.body)[1], { key: 'ArrowLeft' });
		fireEvent.keyDown(chips(document.body)[0], { key: 'ArrowLeft' });
		expect(document.activeElement).toBe(chips(document.body)[3]);

		fireEvent.keyDown(chips(document.body)[3], { key: 'ArrowRight' });
		expect(document.activeElement).toBe(chips(document.body)[0]);
	});
});

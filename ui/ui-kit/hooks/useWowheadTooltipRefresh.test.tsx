import { render, screen } from '@testing-library/react';
import { IconPicker } from '@ui-kit/IconPicker';
import { WowheadAnchor } from '@ui-kit/WowheadAnchor';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { useWowheadTooltipRefresh } from './useWowheadTooltipRefresh';

const Probe = ({ token }: { token: unknown }) => {
	useWowheadTooltipRefresh(token);
	return null;
};

const withWowhead = () => {
	const refreshLinks = vi.fn();
	const triggerTooltip = vi.fn();
	(window as unknown as { WH: unknown }).WH = { Tooltips: { refreshLinks, triggerTooltip } };
	return { refreshLinks, triggerTooltip };
};

afterEach(() => {
	delete (window as unknown as { WH?: unknown }).WH;
	vi.restoreAllMocks();
});

describe('useWowheadTooltipRefresh', () => {
	it('refreshes once per change of the token, not on every render', () => {
		const { refreshLinks } = withWowhead();
		const { rerender } = render(<Probe token="2552-" />);
		expect(refreshLinks).toHaveBeenCalledTimes(1);

		rerender(<Probe token="2552-" />);
		expect(refreshLinks).toHaveBeenCalledTimes(1);

		rerender(<Probe token="2553-" />);
		expect(refreshLinks).toHaveBeenCalledTimes(2);
	});

	it.each([
		['the wowhead script is absent', undefined],
		['the script loaded without the tooltip api', {}],
	])('does nothing when %s', (_, wh) => {
		if (wh) (window as unknown as { WH: unknown }).WH = wh;
		expect(() => render(<Probe token="2552-" />)).not.toThrow();
	});

	it('re-opens the tooltip of the link under the pointer, so it does not go stale mid-hover', () => {
		const { triggerTooltip } = withWowhead();
		const anchor = document.createElement('a');
		document.body.append(anchor);
		vi.spyOn(document, 'elementFromPoint').mockReturnValue(anchor);

		const { rerender } = render(<Probe token="2552-" />);
		document.dispatchEvent(new MouseEvent('mousemove', { clientX: 12, clientY: 34 }));
		rerender(<Probe token="2553-" />);

		expect(triggerTooltip).toHaveBeenCalledWith(anchor);
		anchor.remove();
	});

	it('finds the link when the pointer is over something inside it', () => {
		const { triggerTooltip } = withWowhead();
		const anchor = document.createElement('a');
		const inner = document.createElement('span');
		anchor.append(inner);
		document.body.append(anchor);
		vi.spyOn(document, 'elementFromPoint').mockReturnValue(inner);

		const { rerender } = render(<Probe token="2552-" />);
		document.dispatchEvent(new MouseEvent('mousemove', { clientX: 1, clientY: 2 }));
		rerender(<Probe token="2553-" />);

		expect(triggerTooltip).toHaveBeenCalledWith(anchor);
		anchor.remove();
	});

	it('re-opens the tooltip of an icon swatch whose value changed under the pointer, from its counter', () => {
		const { triggerTooltip } = withWowhead();
		const icon = { iconUrl: 'totem.jpg', name: 'Windfury Totem', href: 'https://www.wowhead.com/spell=10614' };
		const Picker = ({ value }: { value: number }) => {
			useWowheadTooltipRefresh(value);
			return <IconPicker states={0} icon={icon} value={value} onChange={() => {}} />;
		};
		const { rerender } = render(<Picker value={1} />);
		const swatch = screen.getByTestId('icon-picker-root');
		const counter = swatch.firstElementChild!;
		vi.spyOn(document, 'elementFromPoint').mockReturnValue(counter);

		document.dispatchEvent(new MouseEvent('mousemove', { clientX: 5, clientY: 6 }));
		rerender(<Picker value={2} />);

		expect(swatch.textContent).toBe('2');
		expect(swatch.getAttribute('href')).toBe(icon.href);
		expect(swatch.dataset.whtticon).toBe('false');
		expect(triggerTooltip).toHaveBeenLastCalledWith(swatch);
	});

	it('re-opens the tooltip of a wowhead anchor that acts, from the text inside it', () => {
		const { triggerTooltip } = withWowhead();
		const Cell = ({ name }: { name: string }) => {
			useWowheadTooltipRefresh(name);
			return (
				<WowheadAnchor href="https://www.wowhead.com/item=19019" testId="cell" onActivate={() => {}}>
					<span data-testid="cell-name">{name}</span>
				</WowheadAnchor>
			);
		};
		const { rerender } = render(<Cell name="Thunderfury" />);
		vi.spyOn(document, 'elementFromPoint').mockReturnValue(screen.getByTestId('cell-name'));

		document.dispatchEvent(new MouseEvent('mousemove', { clientX: 7, clientY: 8 }));
		rerender(<Cell name="Sulfuras" />);

		expect(triggerTooltip).toHaveBeenLastCalledWith(screen.getByTestId('cell'));
	});

	it('does not trigger a tooltip when nothing is hovered', () => {
		const { triggerTooltip, refreshLinks } = withWowhead();
		render(<Probe token="2552-" />);

		expect(refreshLinks).toHaveBeenCalledTimes(1);
		expect(triggerTooltip).not.toHaveBeenCalled();
	});
});

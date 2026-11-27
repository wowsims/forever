import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { Tooltip } from './Tooltip';

const pointer = vi.hoisted(() => ({ hover: true }));
vi.mock('../../shared/pointer', () => ({ hasHover: () => pointer.hover }));

const Anchored = ({ content = 'Reforge to hit cap' }: { content?: string }) => (
	<>
		<button data-tooltip-id="t1">anchor</button>
		<Tooltip id="t1" content={content} />
	</>
);

describe('Tooltip', () => {
	it('shows its content when the anchor is hovered', async () => {
		render(<Anchored />);
		fireEvent.mouseEnter(screen.getByRole('button'));
		expect(await screen.findByText('Reforge to hit cap')).toBeTruthy();
	});

	// react-tooltip builds content lazily for free: children are not rendered until the tooltip first
	// opens, so a picker inside one costs nothing until it is shown.
	it('does not render its content until it is first opened', () => {
		let renders = 0;
		const Content = () => {
			renders++;
			return <span>Bonus Strength</span>;
		};
		render(
			<>
				<button data-tooltip-id="bonus">Strength</button>
				<Tooltip id="bonus" content={<Content />} />
			</>,
		);
		expect(renders).toBe(0);

		fireEvent.mouseEnter(screen.getByRole('button'));
		expect(renders).toBe(1);
	});

	it('opens on click, not on hover, when openOnClick is set', () => {
		render(
			<>
				<button data-tooltip-id="bonus">Strength</button>
				<Tooltip id="bonus" openOnClick clickable content={<span>Bonus Strength</span>} />
			</>,
		);
		fireEvent.mouseEnter(screen.getByRole('button'));
		expect(screen.queryByText('Bonus Strength')).toBeNull();

		fireEvent.click(screen.getByRole('button'));
		expect(screen.getByText('Bonus Strength')).toBeTruthy();
	});

	// The library injects its own stylesheet unless disableStyleInjection is "core" — bare `true`
	// stops only the base styles. An injected tag lands after the bundle's <link> and outranks the
	// component's theme, which is how the tooltip kept the library's 0.9 opacity in the browser.
	it('injects no stylesheet of its own', async () => {
		render(<Anchored />);
		fireEvent.mouseEnter(screen.getByRole('button'));
		await screen.findByText('Reforge to hit cap');
		expect(document.querySelectorAll('style[id^="react-tooltip-"]')).toHaveLength(0);
	});

	describe('carriesLink', () => {
		afterEach(() => {
			pointer.hover = true;
		});

		// Two link anchors and a tooltip that names its anchor and links to its page, as the game tooltips do.
		const Linked = ({ carriesLink = true }: { carriesLink?: boolean }) => (
			<>
				<a href="https://example.com/1">first</a>
				<a href="https://example.com/2">second</a>
				<Tooltip
					id="linked"
					anchorSelect='a[href^="https://example.com/"]'
					carriesLink={carriesLink}
					render={({ activeAnchor }) => (
						<span>
							Tooltip of {activeAnchor?.textContent}
							<a href={activeAnchor?.getAttribute('href') ?? ''}>page</a>
						</span>
					)}
				/>
			</>
		);
		const pause = (ms: number) => act(() => new Promise(resolve => setTimeout(resolve, ms)));

		it('stays open while the pointer moves from the anchor onto its link, which anchors no tooltip', async () => {
			render(<Linked />);
			fireEvent.mouseOver(screen.getByText('first'));
			const page = await screen.findByRole('link', { name: 'page' });

			fireEvent.mouseOut(screen.getByText('first'), { relatedTarget: page });
			fireEvent.mouseOver(page);
			await pause(250);
			expect(screen.getByRole('tooltip').textContent).toContain('Tooltip of first');
			expect(screen.getByRole('link', { name: 'page' }).getAttribute('href')).toBe('https://example.com/1');
		});

		it('still hides when the pointer leaves an anchor after one closed by Escape with the pointer on its link', async () => {
			render(<Linked />);
			fireEvent.mouseOver(screen.getByText('first'));
			const page = await screen.findByRole('link', { name: 'page' });
			fireEvent.mouseOut(screen.getByText('first'), { relatedTarget: page });
			fireEvent.mouseOver(page);
			fireEvent.keyDown(window, { key: 'Escape' });
			await waitFor(() => expect(screen.queryByRole('tooltip')).toBeNull());

			fireEvent.mouseOver(screen.getByText('second'));
			await screen.findByText(/Tooltip of second/);
			fireEvent.mouseOut(screen.getByText('second'), { relatedTarget: document.body });
			await waitFor(() => expect(screen.queryByRole('tooltip')).toBeNull());
		});

		it('switches to the next anchor the pointer moves onto', async () => {
			render(<Linked />);
			fireEvent.mouseOver(screen.getByText('first'));
			await screen.findByText(/Tooltip of first/);

			fireEvent.mouseOut(screen.getByText('first'), { relatedTarget: screen.getByText('second') });
			fireEvent.mouseOver(screen.getByText('second'), { relatedTarget: screen.getByText('first') });
			expect(await screen.findByText(/Tooltip of second/)).toBeTruthy();
		});

		it('without hover, opens on a tap of a link anchor instead of following it, and follows its own link', async () => {
			pointer.hover = false;
			render(<Linked />);
			expect(fireEvent.click(screen.getByText('first'))).toBe(false);
			const page = await screen.findByRole('link', { name: 'page' });
			expect(fireEvent.click(page)).toBe(true);
		});

		it('leaves a tapped link anchor to follow its link on a tooltip that carries none', () => {
			pointer.hover = false;
			render(<Linked carriesLink={false} />);
			expect(fireEvent.click(screen.getByText('first'))).toBe(true);
		});

		it('leaves a clicked link anchor to follow its link where there is hover', () => {
			render(<Linked />);
			expect(fireEvent.click(screen.getByText('first'))).toBe(true);
		});
	});
});

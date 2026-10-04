import i18n from '@i18n/config';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ContentBlock, type ContentBlockHeaderProps } from './ContentBlock';

describe('ContentBlock', () => {
	it.each([
		['absent', {}],
		['an empty object', { header: {} as ContentBlockHeaderProps }],
	])('renders no header when config.header is %s', (_, config) => {
		render(<ContentBlock className="mt-2" config={config} />);
		expect(screen.queryByTestId('content-block-header')).toBeNull();
	});

	it('renders the header with the title text when config.header has a title', () => {
		render(<ContentBlock className="mt-2" config={{ header: { title: 'My Title' } }} />);
		const header = screen.queryByTestId('content-block-header');
		expect(header).not.toBeNull();
		const title = within(header!).getByTestId('content-block-title');
		expect(title.tagName).toBe('H6');
		expect(title.textContent).toBe('My Title');
	});

	it('uses titleTag for the title element, defaulting to h6', () => {
		render(<ContentBlock className="mt-2" config={{ header: { title: 'My Title', titleTag: 'h3' } }} />);
		expect(screen.getByTestId('content-block-title').tagName).toBe('H3');
	});

	// Five of the eight shipped header tooltips are translation strings carrying <strong> or <br>.
	it('renders a header tooltip as HTML, not as escaped text', () => {
		render(<ContentBlock className="mt-2" config={{ header: { title: 'Raid Buffs', tooltip: 'Buffs by <strong>other</strong> members' } }} />);
		fireEvent.mouseEnter(screen.getByRole('button'));
		expect(screen.getByText('other').tagName).toBe('STRONG');
	});

	it('renders headerChildren after the title, inside the header', () => {
		render(<ContentBlock className="mt-2" config={{ header: { title: 'Raid Buffs' } }} headerChildren={<p className="text-sm">Describes it</p>} />);
		const header = screen.getByTestId('content-block-header');
		expect(Array.from(header.children).map(child => child.getAttribute('data-testid') ?? child.tagName)).toEqual(['content-block-title', 'P']);
		expect(header.lastElementChild!.textContent).toBe('Describes it');
	});

	it('puts the tooltip button inside the title element, not the header', () => {
		render(<ContentBlock className="mt-2" config={{ header: { title: 'My Title', tooltip: 'explains it' } }} />);
		const title = screen.getByTestId('content-block-title');
		expect(within(title).queryByRole('button')).not.toBeNull();
		// Not a direct child of the header (only of the title).
		const header = screen.getByTestId('content-block-header');
		expect(Array.from(header.children)).toEqual([title]);
	});

	it('names the tooltip button Help, not after the title it sits in', () => {
		render(<ContentBlock config={{ header: { title: 'My Title', tooltip: 'explains it' } }} />);
		const title = screen.getByTestId('content-block-title');
		expect(within(title).getByRole('button', { name: i18n.t('common.help') })).toBeTruthy();
		expect(within(title).queryByRole('button', { name: 'My Title' })).toBeNull();
	});

	it('renders no tooltip button when config.header.tooltip is absent', () => {
		render(<ContentBlock className="mt-2" config={{ header: { title: 'My Title' } }} />);
		expect(screen.queryByRole('button')).toBeNull();
	});

	it('renders children inside the body', () => {
		render(
			<ContentBlock className="mt-2" config={{}}>
				<span>child content</span>
			</ContentBlock>,
		);
		const body = screen.getByTestId('content-block-body');
		expect(body.textContent).toBe('child content');
	});
});

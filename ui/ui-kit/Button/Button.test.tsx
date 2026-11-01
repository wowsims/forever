import { render, screen, within } from '@testing-library/react';
import { Field } from '@ui-kit/Field';
import { Icon } from '@ui-kit/Icon';
import { describe, expect, it } from 'vitest';

import { Button } from './Button';

describe('Button', () => {
	it('is a non-submitting button by default', () => {
		render(<Button>Simulate</Button>);
		const button = screen.getByRole('button', { name: 'Simulate' });
		// The tree has several <button>s inside forms with no type, which submit on click.
		expect(button.getAttribute('type')).toBe('button');
	});

	it('renders an anchor for an href', () => {
		render(
			<Button href="https://wowhead.com/forever/item=1" variant="link">
				Item
			</Button>,
		);
		const link = screen.getByRole('link', { name: 'Item' });
		expect(link.tagName).toBe('A');
		expect(link.getAttribute('href')).toBe('https://wowhead.com/forever/item=1');
		expect(link.getAttribute('type')).toBeNull();
		expect(screen.queryByRole('button')).toBeNull();
	});

	it('accepts an anchor only through href, and an icon-only button only with an aria-label', () => {
		const byAs = (
			// @ts-expect-error `as="a"` is not a Button prop, even with an href: `href` alone renders the anchor.
			<Button as="a" href="/x">
				Item
			</Button>
		);
		// @ts-expect-error An icon-only button has no text to announce, so `aria-label` is required.
		const unnamed = <Button iconOnly>×</Button>;
		const named = (
			<Button iconOnly aria-label="Close">
				×
			</Button>
		);
		expect([byAs, unnamed, named]).toHaveLength(3);
	});

	it('renders a label when asked, for a peer-checked radio', () => {
		render(
			<Button as="label" htmlFor="chart-view-rotation" variant="outline-primary">
				Rotation
			</Button>,
		);
		const label = screen.getByText('Rotation');
		expect(label.tagName).toBe('LABEL');
		expect(label.getAttribute('for')).toBe('chart-view-rotation');
	});

	it('adds rel to a cross-origin link without being asked', () => {
		render(
			<>
				<Button href="https://www.wowhead.com/forever/item=1" target="_blank">
					Item
				</Button>
				<Button href="//discord.gg/p3DgvmnDCS">Discord</Button>
			</>,
		);
		screen.getAllByRole('link').forEach(link => expect(link.getAttribute('rel')).toBe('noopener noreferrer'));
	});

	it('merges an explicit rel rather than replacing it, and does not duplicate', () => {
		render(
			<Button href="https://www.wowhead.com/forever/item=1" rel="nofollow noopener">
				Item
			</Button>,
		);
		expect(screen.getByRole('link').getAttribute('rel')).toBe('nofollow noopener noreferrer');
	});

	// Relative links stay in this document, and the non-http schemes open no browsing context.
	it('leaves relative and non-http hrefs alone', () => {
		render(
			<>
				<Button href="/forever/warrior/arms/">Arms</Button>
				<Button href="#gear-tab">Gear</Button>
				<Button href="mailto:someone@example.com">Mail</Button>
				<Button href="javascript:void(0)">Inert</Button>
			</>,
		);
		screen.getAllByRole('link').forEach(link => expect(link.getAttribute('rel')).toBeNull());
	});

	it.each([
		['a framed button', false],
		['an iconOnly button', true],
	])('lays a small spinner over %s, keeping its content in place but hidden, and marks itself busy', (_kind, iconOnly) => {
		const icon = <Icon name="gear" testId="gear" />;
		render(
			iconOnly ? (
				<Button iconOnly loading aria-label="Settings">
					{icon}
				</Button>
			) : (
				<Button variant="outline-primary" loading aria-label="Settings">
					{icon}
				</Button>
			),
		);
		const button = screen.getByRole('button', { name: 'Settings' });
		expect(button.getAttribute('aria-busy')).toBe('true');

		const rendered = within(button).getByTestId('gear');
		expect(rendered.parentElement!.parentElement).toBe(button);
		expect(rendered.parentElement!.hasAttribute('data-hidden')).toBe(true);

		const spinner = within(button).getByTestId('loader');
		expect(spinner.previousElementSibling).toBe(rendered.parentElement);
		expect(spinner.getAttribute('data-size')).toBe('xs');
	});

	it('emits no size for an iconOnly, unstyled or size none button, nor for a link in a sized field', () => {
		render(
			<Field size="xs">
				<Button iconOnly aria-label="Icon">
					×
				</Button>
				<Button variant="unstyled">Unstyled</Button>
				<Button size="none">None</Button>
				<Button variant="link">Link</Button>
			</Field>,
		);
		expect(['Icon', 'Unstyled', 'None', 'Link'].map(name => screen.getByRole('button', { name }).hasAttribute('data-size'))).toEqual([
			false,
			false,
			false,
			false,
		]);
	});

	it('renders no spinner and no aria-busy when not loading', () => {
		render(<Button>Simulate</Button>);
		const button = screen.getByRole('button', { name: 'Simulate' });
		expect(button.hasAttribute('aria-busy')).toBe(false);
		expect(within(button).queryByTestId('loader')).toBeNull();
	});
});

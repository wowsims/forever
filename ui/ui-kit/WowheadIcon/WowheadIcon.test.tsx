import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { WowheadIcon } from './WowheadIcon';

describe('WowheadIcon', () => {
	it('renders an anchor forwarding href and rel by default, and a div when asked for one', () => {
		const anchor = render(<WowheadIcon iconUrl="icon.jpg" href="https://wowhead/spell=1" />).container.firstElementChild!;
		expect(anchor.tagName).toBe('A');
		expect(anchor.getAttribute('href')).toBe('https://wowhead/spell=1');
		expect(anchor.getAttribute('rel')).toBe('noopener noreferrer');

		const div = render(<WowheadIcon as="div" iconUrl="icon.jpg" />).container.firstElementChild!;
		expect(div.tagName).toBe('DIV');
	});

	it.each(['a', 'div'] as const)('paints the background image on the %s', as => {
		const icon = render(<WowheadIcon as={as} iconUrl="icon.jpg" />).container.firstElementChild as HTMLElement;
		expect(icon.style.backgroundImage).toBe('url("icon.jpg")');
	});

	it('spreads the wowhead dataset onto the element, as an anchor or a div', () => {
		const anchor = render(<WowheadIcon iconUrl="icon.jpg" wowhead={{ 'data-wowhead': 'spell=1&buff=1' }} />).container.querySelector('a')!;
		expect(anchor.getAttribute('data-wowhead')).toBe('spell=1&buff=1');

		const div = render(<WowheadIcon as="div" iconUrl="icon.jpg" wowhead={{ 'data-wowhead': 'item=2' }} />).container.querySelector('div')!;
		expect(div.getAttribute('data-wowhead')).toBe('item=2');
	});

	it('writes no dataset without one', () => {
		const icon = render(<WowheadIcon iconUrl="icon.jpg" />).container.querySelector('a')!;
		expect(icon.hasAttribute('data-wowhead')).toBe(false);
	});
});

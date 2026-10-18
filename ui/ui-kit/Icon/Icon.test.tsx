import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { DEFINITIONS, iconDefinition } from './definitions';
import { Icon } from './Icon';
import type { IconName } from './types';

const el = (c: HTMLElement) => c.querySelector('svg')!;

describe('Icon', () => {
	it('draws the glyph of the style it is given', () => {
		const path = (c: HTMLElement) => el(c).querySelector('path')!.getAttribute('d');
		const solid = path(render(<Icon name="circle-question" />).container);
		const regular = path(render(<Icon name="circle-question" style="regular" />).container);
		expect(regular).not.toBe(solid);
		expect(path(render(<Icon name="circle-question" style="base" />).container)).toBe(solid);
		expect(el(render(<Icon name="github" style="brands" />).container).dataset.prefix).toBe('fab');
	});

	it('falls back to solid for a name FontAwesome Free has no regular glyph for', () => {
		expect(iconDefinition('ban', 'regular')).toBe(iconDefinition('ban', 'solid'));
		expect(iconDefinition('star', 'regular')).not.toBe(iconDefinition('star', 'solid'));
	});

	it('keeps rotate-left and arrow-rotate-left as distinct glyphs', () => {
		expect(iconDefinition('rotate-left', 'solid').icon[3]).toBe('f2ea');
		expect(iconDefinition('arrow-rotate-left', 'solid').icon[3]).toBe('f0e2');
	});

	it('draws the World of Warcraft logo as one path in a 20 by 20 box', () => {
		const svg = el(render(<Icon name="wow" style="brands" />).container);
		expect(svg.getAttribute('viewBox')).toBe('0 0 20 20');
		expect(svg.querySelectorAll('path')).toHaveLength(1);
	});

	it.each(Object.keys(DEFINITIONS) as IconName[])('renders %s from a definition, canonically named', name => {
		const definition = iconDefinition(name, 'solid');
		expect(definition).toBeDefined();
		expect(definition.iconName).toBe(name);
		expect(el(render(<Icon name={name} />).container).dataset.icon).toBe(name);
	});

	it('marks the size as data-size', () => {
		expect(el(render(<Icon name="spinner" size="2xl" />).container).dataset.size).toBe('2xl');
		expect(el(render(<Icon name="spinner" />).container).hasAttribute('data-size')).toBe(false);
	});

	it('takes its test id from testId', () => {
		const { getByTestId } = render(<Icon name="gear" testId="settings-icon" />);
		expect(getByTestId('settings-icon').tagName.toLowerCase()).toBe('svg');
	});

	it('keeps a data-testid passed as an attribute when it has no testId', () => {
		const { getByTestId } = render(<Icon name="gear" data-testid="gear-marker" />);
		expect(getByTestId('gear-marker').tagName.toLowerCase()).toBe('svg');
	});

	it('is hidden from assistive tech unless it carries a title', () => {
		expect(el(render(<Icon name="xmark" />).container).getAttribute('aria-hidden')).toBe('true');
		const titled = el(render(<Icon name="xmark" title="Close" />).container);
		expect(titled.getAttribute('aria-hidden')).toBe('false');
		expect(titled.getAttribute('aria-label')).toBe('Close');
	});

	// character_stats.tsx anchors a tooltip on the icon itself; a dropped attribute fails silently.
	it('passes an unknown attribute through, so it can anchor its own tooltip', () => {
		const { container } = render(<Icon name="circle-question" data-tooltip-id="bonus-str" />);
		expect(el(container).dataset.tooltipId).toBe('bonus-str');
	});
});

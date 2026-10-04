import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { Notice } from './Notice';

const icon = (el: HTMLElement) => el.querySelector('svg')!;

describe('Notice', () => {
	it.each([
		['info', 'circle-info'],
		['warning', 'triangle-exclamation'],
		['error', 'triangle-exclamation'],
	] as const)('renders the %s icon before the children', (level, glyph) => {
		render(
			<Notice level={level} testId="notice">
				Some text
			</Notice>,
		);
		const notice = screen.getByTestId('notice');
		expect(notice.tagName).toBe('DIV');
		expect(notice.getAttribute('data-level')).toBe(level);
		expect(notice.textContent).toBe('Some text');
		expect(notice.firstElementChild).toBe(icon(notice));
		expect(icon(notice).getAttribute('data-icon')).toBe(glyph);
	});
});

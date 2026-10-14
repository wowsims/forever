import { Chart } from 'chart.js';
import { afterEach, describe, expect, it } from 'vitest';

import { applyThemeFontDefaults } from './chart_fonts';

const root = document.documentElement.style;

const theme = (fontSize: string, textXs: string, textSm: string) => {
	root.fontSize = fontSize;
	root.setProperty('--text-xs', textXs);
	root.setProperty('--text-sm', textSm);
};

afterEach(() => {
	root.removeProperty('font-size');
	root.removeProperty('--text-xs');
	root.removeProperty('--text-sm');
});

describe('applyThemeFontDefaults', () => {
	it('sets the body size and every title size from the theme, scaling rem by the root font size', () => {
		theme('14px', '0.75rem', '0.875rem');

		applyThemeFontDefaults();

		expect(Chart.defaults.font.size).toBe(10.5);
		expect(Chart.defaults.plugins.title.font).toMatchObject({ size: 12.25 });
		expect(Chart.defaults.plugins.legend.title.font).toMatchObject({ size: 12.25 });
	});

	it('takes a px size as it is', () => {
		theme('16px', '13px', '15px');

		applyThemeFontDefaults();

		expect(Chart.defaults.font.size).toBe(13);
		expect(Chart.defaults.plugins.title.font).toMatchObject({ size: 15 });
	});
});

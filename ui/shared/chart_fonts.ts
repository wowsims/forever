import { Chart } from 'chart.js';

const cssLengthToPx = (value: string, rootPx: number): number => {
	const length = parseFloat(value);
	return value.trim().endsWith('rem') ? length * rootPx : length;
};

// Chart.js sizes text in pixels, so the theme's rem sizes are resolved against the current root font size (14 px or 16 px by viewport).
export const applyThemeFontDefaults = () => {
	const root = getComputedStyle(document.documentElement);
	const rootPx = parseFloat(root.fontSize);
	const family = root.getPropertyValue('--font-sans').trim();
	const body = cssLengthToPx(root.getPropertyValue('--text-xs'), rootPx);
	const heading = cssLengthToPx(root.getPropertyValue('--text-sm'), rootPx);

	if (family) Chart.defaults.font.family = family;
	if (body) Chart.defaults.font.size = body;
	if (!heading) return;
	// Font sizes inherit from the global default only, so each titled element names the larger size itself.
	Chart.defaults.set('plugins.title', { font: { size: heading } });
	Chart.defaults.set('plugins.legend', { title: { font: { size: heading } } });
	Chart.defaults.set('scale', { title: { font: { size: heading } } });
};

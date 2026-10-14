// One per colour slot, in slot order; the order keeps neighbouring slots apart under colour-vision deficiency.
export const SERIES_SWATCH_CLASSES = [
	'bg-chart-series-1',
	'bg-chart-series-2',
	'bg-chart-series-3',
	'bg-chart-series-4',
	'bg-chart-series-5',
	'bg-chart-series-6',
	'bg-chart-series-7',
	'bg-chart-series-8',
] as const;

export const SERIES_SLOT_COUNT = SERIES_SWATCH_CLASSES.length;

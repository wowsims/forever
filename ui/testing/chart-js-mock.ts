export const fakeCharts: Array<unknown> = [];

/**
 * A `chart.js` stand-in that builds nothing and records each chart it is asked for:
 *
 * ```ts
 * vi.mock('chart.js', async () => (await import('@testing/chart-js-mock')).chartJsMock);
 * ```
 */
export const chartJsMock = {
	BarController: class {},
	BubbleController: class {},
	DoughnutController: class {},
	LineController: class {},
	PieController: class {},
	PolarAreaController: class {},
	RadarController: class {},
	ScatterController: class {},
	Chart: class {
		static register() {}
		data: unknown;
		updates = 0;
		destroyed = false;
		constructor(
			public canvas: HTMLCanvasElement,
			public config: { data: unknown; options: unknown },
		) {
			this.data = config.data;
			fakeCharts.push(this);
		}
		// chart.js resolves `options` through a proxy whose set trap writes back to `config.options`.
		get options() {
			return this.config.options;
		}
		update() {
			this.updates++;
		}
		destroy() {
			this.destroyed = true;
		}
	},
};

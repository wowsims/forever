// How a sim's numbers read, in the editor's status bar and decorations and in the panel alike.

import type { SimResult, SimRun } from './protocol';

// A row's runs per fight: whole numbers from 10, one decimal from 1, two below.
export const perFight = (n: number) => (n >= 10 ? n.toFixed(0) : n.toFixed(n >= 1 ? 1 : 2));

// The standard error of a run's mean DPS.
export const standardError = (run: SimRun) => run.dps.stdev / Math.sqrt(run.iterations);

// A compare's difference: "+12.3 (0.45%)".
export const differenceText = ({ delta, percent }: NonNullable<SimResult['difference']>) =>
	`${delta >= 0 ? '+' : ''}${delta.toFixed(1)} (${percent.toFixed(2)}%)`;

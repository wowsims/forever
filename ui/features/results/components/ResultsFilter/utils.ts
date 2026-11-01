import i18n from '@i18n/config';
import type { SimResult, SimResultFilter } from '@sim/proto/sim_result';
import type { ChoiceOption } from '@ui-kit/ChoiceField';
import { CLASS_TEXT } from '@ui-kit/utils/colors';

/** No target picked: every target's damage is counted. */
export const ALL_UNITS = -1;

export const simResultFilter = (target: number): SimResultFilter => ({ target: target === ALL_UNITS ? null : target });

/** False once a run has fewer targets than the one the filter is pointing at. */
export const hasTarget = (result: SimResult, target: number): boolean => target === ALL_UNITS || result.getTargets().some(unit => unit.index === target);

/** All targets, then one option per target of the run, each valued by its encounter index; an add of the design's waves goes by its wave and add. */
export const targetOptions = (result: SimResult): Array<ChoiceOption<number>> => [
	{ value: ALL_UNITS, label: i18n.t('results_tab.details.all_targets') },
	...result.getTargets().map((unit): ChoiceOption<number> => ({
		value: unit.index,
		label: unit.designAddName ?? i18n.t('results_tab.details.target_number', { number: unit.index + 1 }),
		icon: unit.iconUrl ? { iconUrl: unit.iconUrl, name: unit.name } : undefined,
		className: unit.classColor && CLASS_TEXT[unit.classColor],
	})),
];

import i18n from '@i18n/config';
import type { SimResult, SimResultFilter } from '@sim/proto/sim_result';
import type { ChoiceOption } from '@ui-kit/ChoiceField';
import { CLASS_TEXT } from '@ui-kit/utils/colors';

/** No target picked: every target's damage is counted. */
export const ALL_UNITS = -1;

export const simResultFilter = (target: number): SimResultFilter => ({ target: target === ALL_UNITS ? null : target });

const healingTargetIndexes = (result: SimResult): Array<number> => {
	const healingPlayers = result.getRaidIndexedPlayers().filter(player => player.spec?.isHealingSpec);
	const targetIndexes = new Set<number>();
	for (const player of healingPlayers) {
		for (const action of player.getHealingActions()) {
			for (const target of action.targets) {
				if (target.healing > 0 || target.shielding > 0) targetIndexes.add(target.data.unitIndex);
			}
		}
	}
	return [...targetIndexes].sort((a, b) => a - b);
};

/** False once a run has fewer targets than the one the filter is pointing at. */
export const hasTarget = (result: SimResult, target: number): boolean =>
	target === ALL_UNITS || result.getTargets().some(unit => unit.index === target) || healingTargetIndexes(result).includes(target);

/** Shows only healed targets in a healing sim; otherwise, all targets in encounter order. */
export const targetOptions = (result: SimResult): Array<ChoiceOption<number>> => {
	const healers = result.getRaidIndexedPlayers().filter(player => player.spec?.isHealingSpec);
	const healingSimulation = healers.length > 0;
	let healTargetNumber = 0;
	const healerIndexes = new Set(healers.map(healer => healer.unitIndex));
	const targets = healingSimulation
		? healingTargetIndexes(result).map((index): ChoiceOption<number> => ({
				value: index,
				label: healerIndexes.has(index)
					? i18n.t('results_tab.details.healer_target')
					: i18n.t('results_tab.details.heal_target_number', { number: ++healTargetNumber }),
			}))
		: result.getTargets().map((unit): ChoiceOption<number> => ({
				value: unit.index,
				label: i18n.t('results_tab.details.target_number', { number: unit.index + 1 }),
				icon: unit.iconUrl ? { iconUrl: unit.iconUrl, name: unit.name } : undefined,
				className: unit.classColor && CLASS_TEXT[unit.classColor],
			}));
	return [{ value: ALL_UNITS, label: i18n.t('results_tab.details.all_targets') }, ...targets];
};

// The topline metric list, as data. One model, two renderers: the sidebar panel draws it as a
// column of `.results-metric` divs, the three detailed-results panes draw it as a one-row table.
// Neither layout lives here — both are `components/ResultMetricList`, picked by its `layout` prop.
import { Spec } from '@generated/proto/common';
import i18n from '@i18n/config';
import { Hunter } from '@sim/player/classes/hunter';
import { Rogue } from '@sim/player/classes/rogue';
import { Warrior } from '@sim/player/classes/warrior';
import { PlayerSpecs } from '@sim/player/specs/index';
import { ActionMetrics, type SimResult, type SimResultFilter } from '@sim/proto/sim_result';
import { DANGER_TEXT } from '@ui-kit/utils/colors';

import { resultMetricCategories, type ResultMetrics } from './sim_results';

export type ResultMetricUnit = 'percentage' | 'number' | 'seconds' | undefined;

export type ResultMetric = {
	metric: keyof ResultMetrics;
	name: string;
	average: number;
	stdev?: number;
	extraClass?: string;
	unit?: ResultMetricUnit;
};

export type ToplineDisplayMetrics = { damage: boolean; threat: boolean; healing: boolean };

export type ToplineMetricsOptions = {
	showOutOfMana?: boolean;
	displayMetrics?: ToplineDisplayMetrics;
};

const NO_MANA_CLASSES = [Rogue, Warrior, Hunter];

// The two specs whose topline drops Time-to-OOM and HPS. The list is narrower than the set of
// tank specs: a protection warrior and a feral bear both tank, but only the bear's sheet drops
// the mana rows.
const TANK_SPECS = [Spec.SpecFeralBearDruid, Spec.SpecProtectionPaladin];

const column = (
	metric: keyof ResultMetrics,
	average: number,
	options: { stdev?: number; unit?: ResultMetricUnit; extraClass?: string } = {},
): ResultMetric => ({
	metric,
	name: i18n.t(`sidebar.results.metrics.${metric}.label`),
	average,
	stdev: options.stdev,
	extraClass: options.extraClass,
	unit: options.unit,
});

/** The out-of-mana column is a single-player, mana-using-class call the detailed-results rows make and the sidebar does not. */
export const showsOutOfMana = (simResult: SimResult, filter?: SimResultFilter): boolean => {
	const players = simResult.getRaidIndexedPlayers(filter);

	return players.length === 1 && !!players[0].spec && !NO_MANA_CLASSES.some(klass => PlayerSpecs.getPlayerClass(players[0].spec!) === klass);
};

export const toplineResultMetrics = (simResult: SimResult, filter?: SimResultFilter, options: ToplineMetricsOptions = {}): ResultMetric[] => {
	const { showOutOfMana = false, displayMetrics } = options;

	const players = simResult.getRaidIndexedPlayers(filter);

	const resultColumns: ResultMetric[] = [];

	const playerMetrics = players[0];
	const spec = playerMetrics.spec?.specID;
	const isTank = TANK_SPECS.includes(spec);
	const isHealer = playerMetrics.spec?.isHealingSpec ?? false;

	const targetIndex = playerMetrics.getTargetIndex(filter);
	const filteredActions = targetIndex === null ? [] : simResult.getRaidIndexedActionMetrics(filter);
	const filteredMetric = filteredActions.length ? ActionMetrics.merge(filteredActions) : null;

	const hpsColumn = isTank
		? null
		: targetIndex === null
			? column('hps', playerMetrics.hps.avg, { stdev: playerMetrics.hps.stdev })
			: column('hps', filteredMetric?.hps ?? 0);
	if (isHealer && hpsColumn) resultColumns.push(hpsColumn);

	if (targetIndex === null) {
		const { chanceOfDeath, dps: dpsMetrics, tps: tpsMetrics, dtps: dtpsMetrics, tmi: tmiMetrics } = playerMetrics;

		resultColumns.push(column('dps', dpsMetrics.avg, { stdev: dpsMetrics.stdev }));
		resultColumns.push(column('tps', tpsMetrics.avg, { stdev: tpsMetrics.stdev }));
		resultColumns.push(column('apm', playerMetrics.actionsPerMinute(filter)));
		resultColumns.push(column('dtps', dtpsMetrics.avg, { stdev: dtpsMetrics.stdev }));

		resultColumns.push(column('tmi', tmiMetrics.avg, { stdev: tmiMetrics.stdev, unit: 'percentage' }));

		// A protection paladin's chance of death is the one number the sheet suppresses.
		if (spec !== Spec.SpecProtectionPaladin) {
			resultColumns.push(column('cod', chanceOfDeath.avg, { stdev: chanceOfDeath.stdev, unit: 'percentage' }));
		}
	} else {
		if (filteredMetric) {
			const { dps, tps } = filteredMetric;
			resultColumns.push(column('dps', dps));
			resultColumns.push(column('tps', tps));
			resultColumns.push(column('apm', playerMetrics.actionsPerMinute(filter)));
		}

		const targetActions = simResult
			.getTargets(filter)
			.map(target => target.actions)
			.flat()
			.map(action => action.forTarget({ player: playerMetrics.unitIndex }));
		if (!!targetActions.length) {
			const { dps: dtps } = ActionMetrics.merge(targetActions);
			resultColumns.push(column('dtps', dtps));
		}
	}

	if (!isTank) {
		if (!isHealer) {
			resultColumns.push(column('tto', playerMetrics.tto.avg, { stdev: playerMetrics.tto.stdev, unit: 'seconds' }));
		}
		if (!isHealer && hpsColumn) resultColumns.push(hpsColumn);
	}

	if (simResult.request.encounter?.useHealth) {
		resultColumns.push(column('dur', simResult.result.avgIterationDuration, { unit: 'seconds' }));
	}

	if (showOutOfMana) {
		const secondsOOM = players[0].secondsOomAvg;
		const percentOOM = secondsOOM / simResult.encounterMetrics.durationSeconds;
		const dangerLevel = percentOOM < 0.01 ? 'safe' : percentOOM < 0.05 ? 'warning' : 'danger';

		resultColumns.push(column('oom', secondsOOM, { unit: 'seconds', extraClass: DANGER_TEXT[dangerLevel] }));
	}

	return displayMetrics
		? resultColumns.filter(({ metric }) => {
				// Threat per second and actions per minute are the metrics every spec shows; the rest follow the setting.
				if (metric === 'tps' || metric === 'apm') return true;
				const category = resultMetricCategories[metric];
				return category !== 'damage' && category !== 'threat' && category !== 'healing' ? true : displayMetrics[category];
			})
		: resultColumns;
};

import i18n from '@i18n/config';
import type { ActionMetrics } from '@sim/proto/sim_result';
import { percentOf } from '@sim/utils/math';

import { MetricsCombinedTooltip, type MetricsCombinedTooltipGroup } from '../MetricsCombinedTooltip';

/** The Count header replaced by "Amount", which every breakdown of a damage or threat total uses. */
export const amountHeader = (): Array<string | undefined> => [undefined, i18n.t('results_tab.details.tooltip_table.amount')];

/**
 * Crushing blows land only on the player, so the damage-taken table is the one breakdown that shows
 * them. Every other such outcome — the `resisted*` family, blocked crits — applies to both
 * directions and needs no flag.
 */
export interface AttackBreakdownOptions {
	crush?: boolean;
}

const damageBreakdownGroup = (metric: ActionMetrics, { crush }: AttackBreakdownOptions = {}): MetricsCombinedTooltipGroup => {
	const done = metric.damageDone;
	return {
		spellSchool: metric.spellSchool,
		totalPercentage: 100,
		data: [
			{ name: i18n.t('results_tab.details.attack_types.hit'), ...done.hit },
			{ name: i18n.t('results_tab.details.attack_types.resisted_hit'), ...done.resistedHit },
			{ name: i18n.t('results_tab.details.attack_types.critical_hit'), ...done.critHit },
			{ name: i18n.t('results_tab.details.attack_types.resisted_critical_hit'), ...done.resistedCritHit },
			{ name: i18n.t('results_tab.details.attack_types.tick'), ...done.tick },
			{ name: i18n.t('results_tab.details.attack_types.resisted_tick'), ...done.resistedTick },
			{ name: i18n.t('results_tab.details.attack_types.critical_tick'), ...done.critTick },
			{ name: i18n.t('results_tab.details.attack_types.resisted_critical_tick'), ...done.resistedCritTick },
			{ name: i18n.t('results_tab.details.attack_types.glancing_blow'), ...done.glance },
			{ name: i18n.t('results_tab.details.attack_types.blocked_hit'), ...done.block },
			{ name: i18n.t('results_tab.details.attack_types.blocked_critical_hit'), ...done.blockedCrit },
			...(crush ? [{ name: i18n.t('results_tab.details.attack_types.crushing_blow'), ...done.crush }] : []),
		],
	};
};

/** The values a class or spec records on the action, one row per key, or no group when it records none. */
export const customMetricsGroups = (metric: ActionMetrics): Array<MetricsCombinedTooltipGroup> => {
	const customMetrics = metric.customMetrics;
	if (!customMetrics.length) return [];
	return [
		{
			name: i18n.t('results_tab.details.custom_metrics.group'),
			spellSchool: metric.spellSchool,
			totalPercentage: 100,
			data: customMetrics.map(({ key, ...entry }) => ({
				name: i18n.t(`results_tab.details.custom_metrics.${key}`, { defaultValue: key }),
				...entry,
			})),
		},
	];
};

const castsGroup = (metric: ActionMetrics, { crush }: AttackBreakdownOptions = {}): MetricsCombinedTooltipGroup => {
	const landed = metric.landedHits || metric.casts;
	const blocked = metric.blocks + metric.blockedCrits;
	return {
		spellSchool: metric.spellSchool,
		totalPercentage: 100,
		data: [
			{
				name: i18n.t('results_tab.details.attack_types.hit'),
				value: (metric.landedHits || metric.casts - metric.totalMisses) - blocked - (crush ? metric.crushes : 0),
				percentage: (landed / (landed + metric.totalMisses)) * 100,
			},
			{
				name: i18n.t('results_tab.details.attack_types.blocked_hit'),
				value: blocked,
				percentage: metric.blockPercent + metric.blockedCritPercent,
			},
			{ name: i18n.t('results_tab.details.attack_types.miss'), value: metric.misses, percentage: metric.missPercent },
			{ name: i18n.t('results_tab.details.attack_types.parry'), value: metric.parries, percentage: metric.parryPercent },
			{ name: i18n.t('results_tab.details.attack_types.dodge'), value: metric.dodges, percentage: metric.dodgePercent },
			...(crush ? [{ name: i18n.t('results_tab.details.attack_types.crushing_blow'), value: metric.crushes, percentage: metric.crushPercent }] : []),
		],
	};
};

const hitGroups = (metric: ActionMetrics, { crush }: AttackBreakdownOptions = {}): Array<MetricsCombinedTooltipGroup> => {
	const ofHits = (value: number) => percentOf(value, metric.landedHits);
	const ofTicks = (value: number) => percentOf(value, metric.landedTicks);

	return [
		{
			spellSchool: metric.spellSchool,
			totalPercentage: 100,
			data: [
				{
					name: i18n.t('results_tab.details.attack_types.hit'),
					value: metric.hits - metric.resistedHits,
					percentage: ofHits(metric.hits - metric.resistedHits),
				},
				{ name: i18n.t('results_tab.details.attack_types.resisted_hit'), value: metric.resistedHits, percentage: ofHits(metric.resistedHits) },
				{
					name: i18n.t('results_tab.details.attack_types.critical_hit'),
					value: metric.crits - metric.resistedCrits,
					percentage: ofHits(metric.crits - metric.resistedCrits),
				},
				{ name: i18n.t('results_tab.details.attack_types.blocked_critical_hit'), value: metric.blockedCrits, percentage: ofHits(metric.blockedCrits) },
				{
					name: i18n.t('results_tab.details.attack_types.resisted_critical_hit'),
					value: metric.resistedCrits,
					percentage: ofHits(metric.resistedCrits),
				},
				{ name: i18n.t('results_tab.details.attack_types.glancing_blow'), value: metric.glances, percentage: ofHits(metric.glances) },
				{ name: i18n.t('results_tab.details.attack_types.blocked_hit'), value: metric.blocks, percentage: ofHits(metric.blocks) },
				...(crush
					? [{ name: i18n.t('results_tab.details.attack_types.crushing_blow'), value: metric.crushes, percentage: ofHits(metric.crushes) }]
					: []),
			],
		},
		{
			spellSchool: metric.spellSchool,
			totalPercentage: 100,
			data: [
				{
					name: i18n.t('results_tab.details.attack_types.tick'),
					value: metric.ticks - metric.resistedTicks,
					percentage: ofTicks(metric.ticks - metric.resistedTicks),
				},
				{ name: i18n.t('results_tab.details.attack_types.resisted_tick'), value: metric.resistedTicks, percentage: ofTicks(metric.resistedTicks) },
				{
					name: i18n.t('results_tab.details.attack_types.critical_tick'),
					value: metric.critTicks - metric.resistedCritTicks,
					percentage: ofTicks(metric.critTicks - metric.resistedCritTicks),
				},
				{
					name: i18n.t('results_tab.details.attack_types.resisted_critical_tick'),
					value: metric.resistedCritTicks,
					percentage: ofTicks(metric.resistedCritTicks),
				},
			],
		},
	];
};

const missGroup = (metric: ActionMetrics): MetricsCombinedTooltipGroup => ({
	spellSchool: metric.spellSchool,
	totalPercentage: metric.totalMissesPercent,
	data: [
		{ name: i18n.t('results_tab.details.attack_types.miss'), value: metric.misses, percentage: metric.missPercent },
		{ name: i18n.t('results_tab.details.attack_types.parry'), value: metric.parries, percentage: metric.parryPercent },
		{ name: i18n.t('results_tab.details.attack_types.dodge'), value: metric.dodges, percentage: metric.dodgePercent },
	],
});

const threatGroup = (metric: ActionMetrics, value: number): MetricsCombinedTooltipGroup => ({
	spellSchool: metric.spellSchool,
	totalPercentage: 100,
	data: [{ name: i18n.t('results_tab.details.attack_types.threat'), value, percentage: 100 }],
});

/** The threat veto, as a body that is simply not built: `showThreatMetrics` comes from the store, and a `render` that returns nothing draws no tooltip at all. */
export const threatTooltip = (metric: ActionMetrics, value: number, showThreatMetrics: boolean) =>
	showThreatMetrics && value ? <MetricsCombinedTooltip headerValues={amountHeader()} groups={[threatGroup(metric, value)]} /> : null;

export const damageBreakdownTooltip = (metric: ActionMetrics, options?: AttackBreakdownOptions) => (
	<MetricsCombinedTooltip headerValues={amountHeader()} groups={[damageBreakdownGroup(metric, options), ...customMetricsGroups(metric)]} />
);

export const castsTooltip = (metric: ActionMetrics, options?: AttackBreakdownOptions) =>
	(!metric.landedHits && !metric.totalMisses) || metric.isPassiveAction ? null : <MetricsCombinedTooltip groups={[castsGroup(metric, options)]} />;

export const hitsTooltip = (metric: ActionMetrics, options?: AttackBreakdownOptions) =>
	!metric.landedHits && !metric.landedTicks ? null : <MetricsCombinedTooltip groups={hitGroups(metric, options)} />;

export const missTooltip = (metric: ActionMetrics) => (metric.totalMissesPercent ? <MetricsCombinedTooltip groups={[missGroup(metric)]} /> : null);

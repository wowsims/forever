import { Stat } from '@generated/proto/common';
import { ResourceType } from '@generated/proto/spell';
import i18n from '@i18n/config';
import { translateProtoStatName } from '@i18n/localization';
import type {
	AuraLog,
	AuraStacksLog,
	AutoDelayLog,
	CastBeganLog,
	CastCancelledLog,
	CastCompletedLog,
	CastFailedLog,
	CastLog,
	CastPushbackLog,
	CombatLog,
	DamageLog,
	MajorCooldownLog,
	ResourceGroupLog,
	ResourceLog,
	SpellQueuedLog,
	StatChangeLog,
} from '@sim/proto/combat_log';
import { matchTimestampPrefix } from '@sim/proto/combat_log';
import { resourceNames } from '@sim/proto/names';
import { UnitStat } from '@sim/proto/stats';
import { resourceTextClass } from '@ui-kit/utils/colors';
import type { ComponentType } from 'react';
import { Trans } from 'react-i18next';

import { DamageResult } from '../DamageResult';
import { EntityLabel } from '../EntityLabel';
import { ActionLink } from './ActionLink';

// One component per log kind. They stay in this file rather than taking one each: they are private
// one-expression renderers for a single discriminated union, and the file is their unit.
const SourcePrefix = ({ log }: { log: CombatLog }) => <>{log.source && <EntityLabel entity={log.source} />}</>;

// The sim's Stats.FlatString(), as in `{"Strength": 20.000,"CritRating": 14.000,}`, and one entry of it.
const FLAT_STATS = /\{(?:"\w+": -?\d+\.\d+,)+\}/g;
const FLAT_STAT_ENTRY = /"(\w+)": (-?\d+\.\d+),/g;
// The sim's StatDependency.String() on a dynamic dep line: `CritRating *= 1.10` or `CritRating += Intellect * 0.50`.
const DYNAMIC_DEP = /(Dynamic dep \w+ \()(\w+) (?:\*= (-?\d+\.\d+)|\+= (\w+) \* (-?\d+\.\d+))\)/;

const unitStatOf = (name: string): UnitStat | null => {
	const stat = Stat[`Stat${name}` as keyof typeof Stat] as Stat | undefined;
	return stat === undefined ? null : UnitStat.fromStat(stat);
};

// A rating with a percent reads as that percent and defense rating as the skill it buys. Names ending in Percent have no
// Stat of their own and already hold a percent.
const amountAsShown = (name: string, amount: number, decimals: number): string => {
	const unitStat = unitStatOf(name);
	const shown = unitStat?.convertRatingToPercent(amount) ?? amount;
	return `${Number(shown.toFixed(decimals))}${unitStat?.isPercentRating() || name.endsWith('Percent') ? '%' : ''}`;
};

// A += amount is per point of the source and in the destination's units, so a rating source reads per 1% of it. Three
// decimals, as a two-decimal rating amount can be under 0.01%. A *= amount is a bare multiplier.
const dynamicDepAsShown = (_match: string, prefix: string, stat: string, multiplier?: string, source?: string, amount?: string): string => {
	const name = translateProtoStatName(stat);
	if (source === undefined) return `${prefix}${name} *= ${multiplier})`;
	const perShownSource = parseFloat(amount!) * (unitStatOf(source)?.convertPercentToRating(1) ?? 1);
	return `${prefix}${name} += ${translateProtoStatName(source)} * ${amountAsShown(stat, perShownSource, 3)})`;
};

// A rating conversion is a stat dependency, so one block can hold a rating and its percent. Where the two share a name
// (Dodge, Parry, Block, Expertise) only the percent is listed; crit and hit keep their distinct names.
const flatStatsAsShown = (flat: string): string => {
	const entries = Array.from(flat.matchAll(FLAT_STAT_ENTRY), ([, name, amount]) => ({
		name,
		amount: parseFloat(amount),
		label: translateProtoStatName(name),
	}));
	const percentLabels = new Set(entries.filter(({ name }) => name.endsWith('Percent')).map(({ label }) => label));
	return entries
		.filter(({ name, label }) => !(name.endsWith('Rating') && percentLabels.has(label)))
		.map(({ name, amount, label }) => `${amountAsShown(name, amount, 2)} ${label}`)
		.join(', ');
};

// Render only: the exporter and search read the line as the sim wrote it.
const withStatsAsShown = (text: string): string => text.replace(FLAT_STATS, flatStatsAsShown).replace(DYNAMIC_DEP, dynamicDepAsShown);

const DefaultLine = ({ log }: { log: CombatLog }) => {
	// Group 3 only: combat_log's own rawWithoutTimestamp keeps the entity bracket for plain-text
	// export, while this path drops it because EntityLabel supplies a styled replacement.
	const captureArr = matchTimestampPrefix(log.raw);
	const body = <>{withStatsAsShown(captureArr ? captureArr[3] : log.raw)}</>;
	if (!log.source) return body;
	return (
		<>
			<EntityLabel entity={log.source} /> {body}
		</>
	);
};

const DamageLine = ({ log }: { log: DamageLog }) => {
	const threatPostfix = log.source?.isTarget ? '' : ` ${i18n.t('results_tab.details.logs.lines.threat', { threat: log.threat.toFixed(2) })}`;
	return (
		<>
			<SourcePrefix log={log} /> <ActionLink actionId={log.actionId!} /> <DamageResult log={log} />
			{threatPostfix}
		</>
	);
};

const RESOURCE_LINE_KEYS = {
	healthLost: 'results_tab.details.logs.lines.health_lost',
	healthRecovered: 'results_tab.details.logs.lines.health_recovered',
	spent: 'results_tab.details.logs.lines.resource_spent',
	gained: 'results_tab.details.logs.lines.resource_gained',
} as const;

const ResourceLine = ({ log }: { log: ResourceLog }) => {
	const signedDiff = (log.valueAfter - log.valueBefore) * (log.isSpend ? -1 : 1);
	const isHealth = log.resourceType == ResourceType.ResourceTypeHealth;
	const key = isHealth
		? log.isSpend
			? RESOURCE_LINE_KEYS.healthLost
			: RESOURCE_LINE_KEYS.healthRecovered
		: log.isSpend
			? RESOURCE_LINE_KEYS.spent
			: RESOURCE_LINE_KEYS.gained;
	const resourceName = resourceNames.get(log.resourceType)!;
	const resourceClass = resourceTextClass(log.resourceType);

	return (
		<>
			<SourcePrefix log={log} />{' '}
			<Trans
				i18n={i18n}
				i18nKey={key}
				values={{ amount: signedDiff.toFixed(1), resource: resourceName, before: log.valueBefore.toFixed(1), after: log.valueAfter.toFixed(1) }}
				components={{ amount: <strong className={resourceClass} />, action: <ActionLink actionId={log.actionId!} /> }}
			/>
		</>
	);
};

const AURA_LINE_KEYS = {
	gained: 'results_tab.details.logs.lines.aura_gained',
	faded: 'results_tab.details.logs.lines.aura_faded',
	refreshed: 'results_tab.details.logs.lines.aura_refreshed',
} as const;

const AuraLine = ({ log }: { log: AuraLog }) => (
	<>
		<SourcePrefix log={log} />
		{'  '}
		<Trans
			i18n={i18n}
			i18nKey={log.isGained ? AURA_LINE_KEYS.gained : log.isFaded ? AURA_LINE_KEYS.faded : AURA_LINE_KEYS.refreshed}
			components={{ action: <ActionLink actionId={log.actionId!} isAura /> }}
		/>
	</>
);

const AuraStacksLine = ({ log }: { log: AuraStacksLog }) => (
	<>
		<SourcePrefix log={log} />{' '}
		<Trans
			i18n={i18n}
			i18nKey="results_tab.details.logs.lines.aura_stacks"
			values={{ from: log.oldStacks, to: log.newStacks }}
			components={{ action: <ActionLink actionId={log.actionId!} isAura /> }}
		/>
	</>
);

const MajorCooldownLine = ({ log }: { log: MajorCooldownLog }) => (
	<>
		<SourcePrefix log={log} />{' '}
		<Trans i18n={i18n} i18nKey="results_tab.details.logs.lines.major_cooldown" components={{ action: <ActionLink actionId={log.actionId!} /> }} />
	</>
);

const CastBeganLine = ({ log }: { log: CastBeganLog }) => (
	<>
		<SourcePrefix log={log} />{' '}
		<Trans
			i18n={i18n}
			i18nKey="results_tab.details.logs.lines.cast_began"
			values={{ castTime: log.castTime.toFixed(2), cost: log.manaCost.toFixed(1) }}
			components={{ action: <ActionLink actionId={log.actionId!} /> }}
		/>
	</>
);

const CastCancelledLine = ({ log }: { log: CastCancelledLog }) => (
	<>
		<SourcePrefix log={log} />{' '}
		<Trans
			i18n={i18n}
			i18nKey="results_tab.details.logs.lines.cast_cancelled"
			values={{ time: log.cancelTime.toFixed(2) }}
			components={{ action: <ActionLink actionId={log.actionId!} /> }}
		/>
	</>
);

const CastCompletedLine = ({ log }: { log: CastCompletedLog }) => (
	<>
		<SourcePrefix log={log} /> {i18n.t('results_tab.details.logs.lines.cast_completed', { spell: log.actionId!.name })}
	</>
);

const AutoDelayLine = ({ log }: { log: AutoDelayLog }) => (
	<>
		<SourcePrefix log={log} />{' '}
		<Trans
			i18n={i18n}
			i18nKey="results_tab.details.logs.lines.auto_delay"
			values={{ delay: log.delayText, readyAt: log.readyAtLogText }}
			components={{ action: <ActionLink actionId={log.actionId!} /> }}
		/>
	</>
);

const SpellQueuedLine = ({ log }: { log: SpellQueuedLog }) => (
	<>
		<SourcePrefix log={log} />{' '}
		<Trans
			i18n={i18n}
			i18nKey="results_tab.details.logs.lines.spell_queued"
			values={{ fireAt: log.fireAtText }}
			components={{ action: <ActionLink actionId={log.actionId!} /> }}
		/>
	</>
);

const CastFailedLine = ({ log }: { log: CastFailedLog }) => (
	<>
		<SourcePrefix log={log} />{' '}
		<Trans
			i18n={i18n}
			i18nKey="results_tab.details.logs.lines.cast_failed"
			values={{ reason: log.reason }}
			components={{ action: <ActionLink actionId={log.actionId!} /> }}
		/>
	</>
);

const CastPushbackLine = ({ log }: { log: CastPushbackLog }) => (
	<>
		<SourcePrefix log={log} />{' '}
		<Trans
			i18n={i18n}
			i18nKey={log.isChanneling ? 'results_tab.details.logs.lines.pushback_channeling' : 'results_tab.details.logs.lines.pushback_casting'}
			values={{ pushback: log.pushbackText }}
			components={{ action: <ActionLink actionId={log.actionId!} /> }}
		/>
	</>
);

const StatChangeLine = ({ log }: { log: StatChangeLog }) => (
	<>
		<SourcePrefix log={log} />{' '}
		<Trans
			i18n={i18n}
			i18nKey={log.isGain ? 'results_tab.details.logs.lines.stats_gained' : 'results_tab.details.logs.lines.stats_lost'}
			values={{ stats: withStatsAsShown(log.stats) }}
			components={{ action: <ActionLink actionId={log.actionId!} /> }}
		/>
	</>
);

const ResourceGroupLine = ({ log }: { log: ResourceGroupLog }) => (
	<>
		<SourcePrefix log={log} /> {resourceNames.get(log.resourceType)}: {log.valueBefore.toFixed(1)} &rarr; {log.valueAfter.toFixed(1)}
	</>
);

const CastLine = ({ log }: { log: CastLog }) => (
	<>
		<SourcePrefix log={log} /> {i18n.t('results_tab.details.logs.lines.cast', { spell: log.actionId!.name, castTime: log.castTime.toFixed(2) })}
	</>
);

export interface LogLineProps {
	log: CombatLog;
}

type LogOfKind<K extends CombatLog['kind']> = Extract<CombatLog, { kind: K }>;

// A mapped type rather than a switch: the compiler then requires an entry for every kind, where a
// switch only warns about a missing `return`.
const LINE_BY_KIND: { [K in CombatLog['kind']]: ComponentType<{ log: LogOfKind<K> }> } = {
	damage: DamageLine,
	resource: ResourceLine,
	aura: AuraLine,
	'aura-stacks': AuraStacksLine,
	'major-cooldown': MajorCooldownLine,
	'cast-began': CastBeganLine,
	'cast-cancelled': CastCancelledLine,
	'cast-completed': CastCompletedLine,
	'auto-delay': AutoDelayLine,
	'spell-queued': SpellQueuedLine,
	'cast-failed': CastFailedLine,
	'cast-pushback': CastPushbackLine,
	'stat-change': StatChangeLine,
	movement: DefaultLine,
	'resource-group': ResourceGroupLine,
	cast: CastLine,
	plain: DefaultLine,
	dps: DefaultLine,
	'threat-group': DefaultLine,
	'aura-uptime': DefaultLine,
};

export const LogLine = ({ log }: LogLineProps) => {
	// TypeScript cannot correlate the looked-up component with the narrowed `log` — the correlated
	// union limitation — so the cast is here, once, and every component above keeps its narrow props.
	const Content = LINE_BY_KIND[log.kind] as ComponentType<LogLineProps>;
	return <Content log={log} />;
};

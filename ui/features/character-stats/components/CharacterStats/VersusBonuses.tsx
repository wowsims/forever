import {
	areaState,
	familyState,
	mobTypeState,
	type VersusChip,
	versusEffectFactor,
	versusEffectValues,
	type VersusMatch,
	type VersusState,
} from '@features/gear/model/versus_state';
import type { MobTypeBonusSource, PlayerStats, VersusEffect } from '@generated/proto/api';
import type { AreaType, Stat, Target } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateAreaType, translateCreatureFamily, translateMobType, translateStat } from '@i18n/localization';
import { useSim } from '@sim/context/SimHostContext';
import { useActionId } from '@sim/hooks/useActionId';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { ActionId } from '@sim/proto/action_id';
import { signedStatAmount } from '@sim/proto/stats';
import { AccordionItem } from '@ui-kit/Accordion';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { Pill, type PillTone } from '@ui-kit/Pill';
import { Table, TableCell, TableHeaderCell, TableRow } from '@ui-kit/Table';
import { WowheadIcon } from '@ui-kit/WowheadIcon';
import clsx from 'clsx';
import { useState } from 'react';
import { useStore } from 'zustand';

export type VersusBonusSet = Pick<PlayerStats, 'mobTypeBonuses' | 'areaStats' | 'versusEffects'>;

export interface VersusBonusesViewProps {
	bonuses: VersusBonusSet;
	targets: ReadonlyArray<Target>;
	areaTypes: ReadonlyArray<AreaType>;
}

interface VersusLine {
	key: string;
	/** The source a line names by its icon and name; an area's stat line sums its items and has none. */
	actionId?: ActionId;
	/** The name shown in place of the source's own, or the stat an area line sums. */
	label?: string;
	/** What the line gives, one entry per stat or factor. */
	values: Array<string>;
}

interface VersusGroup {
	key: string;
	label: string;
	match: VersusMatch;
	area: boolean;
	lines: Array<VersusLine>;
}

// The gear chips' tones, so a group reads the same as its items' chips.
const TONES: Record<VersusState, PillTone> = { on: 'positive', partial: 'warning', off: 'neutral' };

const stateText = ({ match, area }: VersusGroup) =>
	match.state === 'partial'
		? i18n.t('sidebar.character_stats.versus.state_partial', { matched: match.matched, total: match.total })
		: i18n.t(`sidebar.character_stats.versus.${area ? 'state_area_' : 'state_'}${match.state}`);

const statEntries = (stats: { [key: number]: number }): Array<[Stat, number]> => Object.entries(stats).map(([key, value]) => [Number(key) as Stat, value]);

const areaStatLines = (stats: { [key: number]: number }): Array<VersusLine> =>
	statEntries(stats).map(([stat, value]) => ({ key: `stat-${stat}`, label: translateStat(stat), values: [signedStatAmount(stat, value)] }));

// Worded and rounded as versusEffectFactor words an item's factor.
const factorValues = (key: string, multiplier: number): Array<string> =>
	Math.abs(multiplier - 1) < 1e-9 ? [] : [i18n.t(`sidebar.character_stats.versus.${key}`, { factor: Number(multiplier.toFixed(2)) })];

const sourceLines = (sources: Array<MobTypeBonusSource>): Array<VersusLine> =>
	sources.map(source => {
		const actionId = ActionId.fromProto(source.actionId!);
		return {
			key: `source-${actionId.equalityKey()}`,
			actionId,
			values: [
				...statEntries(source.stats).map(([stat, value]) =>
					i18n.t('sidebar.character_stats.versus.stat', { amount: signedStatAmount(stat, value), stat: translateStat(stat) }),
				),
				...factorValues('effect_damage', source.damageMultiplier),
				...factorValues('effect_crit_damage', source.critDamageMultiplier),
			],
		};
	});

// One line per source, an item or the spell of a set bonus or an enchant, with each kind of factor it has.
const effectLines = (effects: Array<VersusEffect>): Array<VersusLine> => {
	const lines = new Map<string, VersusLine>();
	for (const effect of effects) {
		const actionId = ActionId.fromProto(effect.source!);
		const key = `effect-${actionId.equalityKey()}`;
		const line = lines.get(key) ?? { key, actionId, label: effect.itemName || undefined, values: [] };
		line.values.push(versusEffectFactor(effect));
		lines.set(key, line);
	}
	return [...lines.values()];
};

const uniqueSorted = (values: ReadonlyArray<number>): Array<number> => [...new Set(values)].sort((a, b) => a - b);

const versusGroups = ({ bonuses, targets, areaTypes }: VersusBonusesViewProps): Array<VersusGroup> => {
	const { versusEffects } = bonuses;
	const valuesOf = (kind: VersusChip['kind']) => versusEffects.flatMap(effect => versusEffectValues(effect, kind));
	const effectsOn = (kind: VersusChip['kind'], value: number) =>
		effectLines(versusEffects.filter(effect => versusEffectValues(effect, kind).includes(value)));
	const mobTypeGroups = uniqueSorted([...bonuses.mobTypeBonuses.map(entry => entry.mobType), ...valuesOf('mobType')]).map(mobType => ({
		key: `mob-type-${mobType}`,
		label: translateMobType(mobType),
		match: mobTypeState(mobType, targets),
		area: false,
		lines: [...sourceLines(bonuses.mobTypeBonuses.find(entry => entry.mobType === mobType)?.sources ?? []), ...effectsOn('mobType', mobType)],
	}));
	const familyGroups = uniqueSorted(valuesOf('family')).map(family => ({
		key: `creature-family-${family}`,
		label: translateCreatureFamily(family),
		match: familyState(family, targets),
		area: false,
		lines: effectsOn('family', family),
	}));
	const areaGroups = uniqueSorted([...bonuses.areaStats.map(entry => entry.areaType), ...valuesOf('area')]).map(areaType => ({
		key: `area-type-${areaType}`,
		label: translateAreaType(areaType),
		match: areaState(areaType, areaTypes),
		area: true,
		lines: [...areaStatLines(bonuses.areaStats.find(entry => entry.areaType === areaType)?.stats ?? {}), ...effectsOn('area', areaType)],
	}));
	return [...mobTypeGroups, ...familyGroups, ...areaGroups];
};

const SourceLabel = ({ actionId, label }: { actionId: ActionId; label?: string }) => {
	const { iconUrl, name } = useActionId(actionId);
	return (
		<>
			<WowheadIcon as="div" className="size-4 flex-none border border-border" testId="character-stats-versus-source-icon" iconUrl={iconUrl} />
			{label ?? name}
		</>
	);
};

const GroupRows = ({ group }: { group: VersusGroup }) => (
	<tbody className="ui-character-stats-versus-group" data-state={group.match.state} data-testid="character-stats-versus-group">
		<TableRow className="ui-character-stats-row">
			<TableHeaderCell scope="rowgroup" colSpan={2} wrap className="ui-character-stats-label text-left font-semibold">
				<span className="flex items-center justify-between gap-2">
					{group.label}
					<Pill tone={TONES[group.match.state]} size="xs" rounded="xs" className="shrink-0 whitespace-nowrap" testId="character-stats-versus-state">
						{stateText(group)}
					</Pill>
				</span>
			</TableHeaderCell>
		</TableRow>
		{group.lines.map(line => (
			// One cell per line, so a long value widens its own line and not the whole table's value column.
			<TableRow key={line.key} className="ui-character-stats-row ui-character-stats-versus-line">
				<TableCell colSpan={2}>
					<span className="flex items-start justify-between gap-2">
						<span className="ui-character-stats-label flex min-w-0 items-center gap-2">
							{line.actionId ? <SourceLabel actionId={line.actionId} label={line.label} /> : line.label}
						</span>
						<span className="ui-character-stats-value shrink-0 text-right">
							{line.values.map(value => (
								<span key={value} className="block whitespace-nowrap">
									{value}
								</span>
							))}
						</span>
					</span>
				</TableCell>
			</TableRow>
		))}
	</tbody>
);

export const VersusBonusesView = (props: VersusBonusesViewProps) => {
	const [showInactive, setShowInactive] = useState(false);
	const groups = versusGroups(props);
	const inactive = groups.filter(group => group.match.state === 'off');

	return (
		<Table variant="plain" flush>
			{groups
				.filter(group => group.match.state !== 'off')
				.map(group => (
					<GroupRows key={group.key} group={group} />
				))}
			{inactive.length > 0 && (
				<tbody>
					<TableRow className="ui-character-stats-row">
						<TableCell colSpan={2} className="ui-character-stats-label">
							<Button
								variant="unstyled"
								className="flex w-full cursor-pointer items-center justify-between gap-1 text-muted hover:text-link-hover focus-visible:focus-ring"
								aria-expanded={showInactive}
								onClick={() => setShowInactive(!showInactive)}>
								{i18n.t('sidebar.character_stats.versus.inactive', { count: inactive.length })}
								<Icon name="chevron-down" fixedWidth className={clsx('transition-transform duration-150', !showInactive && '-rotate-90')} />
							</Button>
						</TableCell>
					</TableRow>
				</tbody>
			)}
			{showInactive && inactive.map(group => <GroupRows key={group.key} group={group} />)}
		</Table>
	);
};

export const VersusBonuses = () => {
	const sim = useSim();
	const targets = useStore(sim.store, state => state.encounter.targets);
	const areaTypes = useStore(sim.store, state => state.encounter.areaTypes);
	const bonuses = usePlayerStore('currentStats');
	if (!bonuses.mobTypeBonuses.length && !bonuses.areaStats.length && !bonuses.versusEffects.length) return null;

	return (
		<AccordionItem value="versus" title={i18n.t('sidebar.character_stats.versus.title')} panelClassName="pb-1" testId="character-stats-group-versus">
			<VersusBonusesView bonuses={bonuses} targets={targets} areaTypes={areaTypes} />
		</AccordionItem>
	);
};

import { type PlayerStats, type VersusEffect, VersusEffectKind } from '@generated/proto/api';
import type { AreaType, Stat, Target } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateAreaType, translateCreatureFamily, translateMobType, translateStat } from '@i18n/localization';
import { useSim } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { signedStatAmount } from '@sim/proto/stats';
import { AccordionItem } from '@ui-kit/Accordion';
import { Table, TableCell, TableHeaderCell, TableRow } from '@ui-kit/Table';
import { useStore } from 'zustand';

import { areaState, familyState, mobTypeState, type VersusMatch } from '../../../gear/model/versus_state';

export type VersusBonusSet = Pick<PlayerStats, 'mobTypeBonuses' | 'areaStats' | 'versusEffects'>;

export interface VersusBonusesViewProps {
	bonuses: VersusBonusSet;
	targets: ReadonlyArray<Target>;
	areaTypes: ReadonlyArray<AreaType>;
}

interface VersusLine {
	key: string;
	label: string;
	value: string;
}

interface VersusGroup {
	key: string;
	label: string;
	match: VersusMatch;
	/** The suffix when the group is `off`; a group of an area says it is not in this area. */
	offKey: 'no_target' | 'not_in_area';
	lines: Array<VersusLine>;
}

const suffix = ({ match, offKey }: VersusGroup): string =>
	match.state === 'on'
		? ''
		: i18n.t(match.state === 'partial' ? 'sidebar.character_stats.versus.partial' : `sidebar.character_stats.versus.${offKey}`, {
				matched: match.matched,
				total: match.total,
			});

const signed = (value: number, suffix = '') => `${value > 0 ? '+' : ''}${Number(value.toFixed(2))}${suffix}`;

const statLines = (stats: { [key: number]: number }): Array<VersusLine> =>
	Object.entries(stats).map(([key, value]) => {
		const stat = Number(key) as Stat;
		return { key: `stat-${stat}`, label: translateStat(stat), value: signedStatAmount(stat, value) };
	});

const multiplierLines = (key: string, multiplier: number): Array<VersusLine> =>
	Math.abs(multiplier - 1) < 1e-9 ? [] : [{ key, label: i18n.t(`sidebar.character_stats.versus.${key}`), value: signed((multiplier - 1) * 100, '%') }];

const effectKeys: Record<VersusEffectKind, string> = {
	[VersusEffectKind.VersusEffectKindDamage]: 'effect_damage',
	[VersusEffectKind.VersusEffectKindRestore]: 'effect_restore',
	[VersusEffectKind.VersusEffectKindProcChance]: 'effect_proc_chance',
};

const effectLines = (effects: Array<VersusEffect>): Array<VersusLine> =>
	effects.map(effect => ({
		key: `item-${effect.itemId}`,
		label: effect.itemName,
		value: i18n.t(`sidebar.character_stats.versus.${effectKeys[effect.kind]}`, {
			factor: Number(effect.multiplier.toFixed(2)),
		}),
	}));

const versusGroups = ({ bonuses, targets, areaTypes }: VersusBonusesViewProps): Array<VersusGroup> => {
	const mobTypes = [...new Set([...bonuses.mobTypeBonuses.map(entry => entry.mobType), ...bonuses.versusEffects.flatMap(effect => effect.mobTypes)])];
	const mobTypeGroups = mobTypes
		.sort((a, b) => a - b)
		.map(mobType => {
			const bonus = bonuses.mobTypeBonuses.find(entry => entry.mobType === mobType);
			return {
				key: `mob-type-${mobType}`,
				label: i18n.t('sidebar.character_stats.versus.against', { target: translateMobType(mobType) }),
				match: mobTypeState(mobType, targets),
				offKey: 'no_target' as const,
				lines: [
					...(bonus
						? [
								...statLines(bonus.stats),
								...multiplierLines('damage', bonus.damageMultiplier),
								...multiplierLines('crit_damage', bonus.critDamageMultiplier),
							]
						: []),
					...effectLines(bonuses.versusEffects.filter(effect => effect.mobTypes.includes(mobType))),
				],
			};
		});
	const familyGroups = [...new Set(bonuses.versusEffects.flatMap(effect => effect.creatureFamilies))]
		.sort((a, b) => a - b)
		.map(family => ({
			key: `creature-family-${family}`,
			label: i18n.t('sidebar.character_stats.versus.against', { target: translateCreatureFamily(family) }),
			match: familyState(family, targets),
			offKey: 'no_target' as const,
			lines: effectLines(bonuses.versusEffects.filter(effect => effect.creatureFamilies.includes(family))),
		}));
	const areaGroups = bonuses.areaStats.map(entry => ({
		key: `area-type-${entry.areaType}`,
		label: i18n.t('sidebar.character_stats.versus.in_area', { area: translateAreaType(entry.areaType) }),
		match: areaState(entry.areaType, areaTypes),
		offKey: 'not_in_area' as const,
		lines: statLines(entry.stats),
	}));
	return [...mobTypeGroups, ...familyGroups, ...areaGroups];
};

export const VersusBonusesView = (props: VersusBonusesViewProps) => (
	<Table variant="plain" flush>
		{versusGroups(props).map(group => (
			<tbody key={group.key} className="ui-character-stats-versus-group" data-state={group.match.state} data-testid="character-stats-versus-group">
				<TableRow className="ui-character-stats-row">
					<TableHeaderCell scope="rowgroup" colSpan={2} wrap className="ui-character-stats-label text-left font-semibold">
						<span className="ui-character-stats-versus-dot" aria-hidden="true" />
						{group.label}
						{group.match.state !== 'on' && <span className="ui-character-stats-versus-suffix"> {suffix(group)}</span>}
					</TableHeaderCell>
				</TableRow>
				{group.lines.map(line => (
					<TableRow key={line.key} className="ui-character-stats-row ui-character-stats-versus-line">
						<TableCell className="ui-character-stats-label">{line.label}</TableCell>
						<TableCell className="ui-character-stats-value">{line.value}</TableCell>
					</TableRow>
				))}
			</tbody>
		))}
	</Table>
);

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

import { areaState, familyState, mobTypeState, versusEffectFactor, type VersusMatch, type VersusState } from '@features/gear/model/versus_state';
import type { PlayerStats, VersusEffect } from '@generated/proto/api';
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
	label: string;
	value: string;
	/** The item a line of an item's factor comes from; a stat line sums several sources and has none. */
	itemId?: number;
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

const signed = (value: number, suffix = '') => `${value > 0 ? '+' : ''}${Number(value.toFixed(2))}${suffix}`;

const statLines = (stats: { [key: number]: number }): Array<VersusLine> =>
	Object.entries(stats).map(([key, value]) => {
		const stat = Number(key) as Stat;
		return { key: `stat-${stat}`, label: translateStat(stat), value: signedStatAmount(stat, value) };
	});

const multiplierLines = (key: string, multiplier: number): Array<VersusLine> =>
	Math.abs(multiplier - 1) < 1e-9 ? [] : [{ key, label: i18n.t(`sidebar.character_stats.versus.${key}`), value: signed((multiplier - 1) * 100, '%') }];

const effectLines = (effects: Array<VersusEffect>): Array<VersusLine> =>
	effects.map(effect => ({
		key: `item-${effect.itemId}`,
		label: effect.itemName,
		value: versusEffectFactor(effect),
		itemId: effect.itemId,
	}));

const versusGroups = ({ bonuses, targets, areaTypes }: VersusBonusesViewProps): Array<VersusGroup> => {
	const mobTypes = [...new Set([...bonuses.mobTypeBonuses.map(entry => entry.mobType), ...bonuses.versusEffects.flatMap(effect => effect.mobTypes)])];
	const mobTypeGroups = mobTypes
		.sort((a, b) => a - b)
		.map(mobType => {
			const bonus = bonuses.mobTypeBonuses.find(entry => entry.mobType === mobType);
			return {
				key: `mob-type-${mobType}`,
				label: translateMobType(mobType),
				match: mobTypeState(mobType, targets),
				area: false,
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
			label: translateCreatureFamily(family),
			match: familyState(family, targets),
			area: false,
			lines: effectLines(bonuses.versusEffects.filter(effect => effect.creatureFamilies.includes(family))),
		}));
	const areaGroups = bonuses.areaStats.map(entry => ({
		key: `area-type-${entry.areaType}`,
		label: translateAreaType(entry.areaType),
		match: areaState(entry.areaType, areaTypes),
		area: true,
		lines: statLines(entry.stats),
	}));
	return [...mobTypeGroups, ...familyGroups, ...areaGroups];
};

const ItemIcon = ({ itemId }: { itemId: number }) => {
	const { iconUrl } = useActionId(ActionId.fromItemId(itemId));
	return <WowheadIcon as="div" className="size-4 flex-none border border-border" testId="character-stats-versus-item-icon" iconUrl={iconUrl} />;
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
			<TableRow key={line.key} className="ui-character-stats-row ui-character-stats-versus-line">
				<TableCell className="ui-character-stats-label">
					<span className="flex items-center gap-2">
						{line.itemId !== undefined && <ItemIcon itemId={line.itemId} />}
						{line.label}
					</span>
				</TableCell>
				<TableCell className="ui-character-stats-value">{line.value}</TableCell>
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

import { STAT_GROUPS, StatFilterPopover, useStatFilter } from '@features/stat-filters';
import { ItemSlot, PseudoStat, Spec } from '@generated/proto/common';
import i18n from '@i18n/config';
import { useSimHost, useSpecConfig } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { computeStatAttribution, type Stats, type UnitStat } from '@sim/proto/stats';
import { Accordion, AccordionItem } from '@ui-kit/Accordion';
import { parseStringArray, useLocalStorageState } from '@ui-kit/hooks/useLocalStorageState';
import { Table } from '@ui-kit/Table';
import clsx, { type ClassValue } from 'clsx';
import type { ReactNode } from 'react';
import { useMemo } from 'react';
import { useStore } from 'zustand';

import { AvoidanceRow } from './AvoidanceRow';
import { CritCapRow } from './CritCapRow';
import { CritImmunityRow } from './CritImmunityRow';
import { MissRow } from './MissRow';
import { StatRow } from './StatRow';
import { buildRows, defaultOpenStatGroups, type Row } from './utils/rows';
import {
	critImmunityCapDisplayString,
	meleeCritCapDisplayString,
	shouldShowCritImmunity,
	shouldShowMeleeCritCap,
	statDisplayString,
} from './utils/stat_display';
import { WeaponSkills } from './WeaponSkills';

// Every stat the sheet has a row for: buildRows drops the rest.
const STAT_FILTER_UNIVERSE = STAT_GROUPS.flatMap(group => group.stats);

const GROUPS_STORAGE_SUFFIX = '__characterStatsGroups';

export interface CharacterStatsHeaderViewProps {
	title: string;
	filter?: ReactNode;
	className?: ClassValue;
}

export const CharacterStatsHeaderView = ({ title, filter }: CharacterStatsHeaderViewProps) => (
	<div className="flex items-center justify-between gap-2">
		<h3 data-testid="character-stats-label" className="m-0 text-base leading-tight font-bold">
			{title}
		</h3>
		{filter}
	</div>
);

export interface CharacterStatsPanelViewProps extends CharacterStatsHeaderViewProps {
	children?: ReactNode;
}

export const CharacterStatsPanelView = ({ title, filter, className, children }: CharacterStatsPanelViewProps) => (
	<section className={clsx('ui-character-stats-panel', className)} data-testid="character-stats-root">
		<CharacterStatsHeaderView title={title} filter={filter} />
		{children}
	</section>
);

interface CharacterStatsProps extends Pick<CharacterStatsPanelViewProps, 'className'> {}

export const CharacterStats = ({ className }: CharacterStatsProps) => {
	const host = useSimHost();
	const player = host.player;
	const { displayStats, modifyDisplayStats, overwriteDisplayStats } = useSpecConfig();
	const scope = useMemo(() => ({ defaults: displayStats, universe: STAT_FILTER_UNIVERSE }), [displayStats]);
	const statFilter = useStatFilter('characterStats', scope);
	const [openGroups, setOpenGroups] = useLocalStorageState<string[]>(host.getStorageKey(GROUPS_STORAGE_SUFFIX), parseStringArray, () =>
		defaultOpenStatGroups(player.getPlayerSpec().isTankSpec, displayStats),
	);
	const currentStats = usePlayerStore('currentStats');
	const bonusStats = usePlayerStore('bonusStats');
	const gear = usePlayerStore('gear');
	const dualWielding = player.isDualWielding();
	const { groups, shownStats } = useMemo(() => buildRows(player, statFilter.shown), [player, statFilter.shown]);
	const race = usePlayerStore('race');
	const consumables = usePlayerStore('consumables');
	// Not player state, but read by the facade getters below: getDebuffStats and getMissChanceInfo
	// go through raid.debuffs.
	const debuffs = useStore(host.sim.store, s => s.raid.debuffs);

	const snapshot = useMemo(() => {
		const attribution = computeStatAttribution(
			player.getCurrentStats(),
			bonusStats,
			player.getDebuffStats(),
			modifyDisplayStats ? modifyDisplayStats(player) : {},
			overwriteDisplayStats ? overwriteDisplayStats(player) : undefined,
		);
		const isTank = shouldShowCritImmunity(player);
		const critCapInfo = shouldShowMeleeCritCap(player) ? player.getMeleeCritCapInfo() : undefined;
		return {
			pending: !currentStats.finalStats,
			attribution,
			rangedWeapon: player.getEquippedItem(ItemSlot.ItemSlotRanged),
			critCap: critCapInfo ? { info: critCapInfo, text: meleeCritCapDisplayString(critCapInfo.delta) } : null,
			miss: isTank ? player.getMissChanceInfo() : null,
			avoidance: isTank ? player.getAvoidanceInfo() : null,
			critImmunity: isTank ? { info: player.getCritImmunityInfo(), text: critImmunityCapDisplayString(player) } : null,
		};
		// eslint-disable-next-line react-hooks/exhaustive-deps -- every store value below reaches the body through the player/sim facades rather than by name; they are the invalidation keys, and dropping one stales the snapshot.
	}, [player, currentStats, bonusStats, gear, race, consumables, debuffs, modifyDisplayStats, overwriteDisplayStats]);

	const { pending, attribution, rangedWeapon, critCap, miss, avoidance, critImmunity } = snapshot;
	const show = (deltaStats: Stats, unitStat: UnitStat, includeBase?: boolean) => statDisplayString(player, deltaStats, unitStat, includeBase);

	const hasParry = shownStats.some(stat => stat.equalsPseudoStat(PseudoStat.PseudoStatParryPercent));
	const hasBlock = shownStats.some(stat => stat.equalsPseudoStat(PseudoStat.PseudoStatBlockPercent));

	const renderRow = (row: Row) => {
		switch (row.kind) {
			case 'stat':
				return (
					<StatRow
						key={row.id}
						displayStat={row.displayStat}
						bonusStats={bonusStats}
						attribution={attribution}
						show={show}
						rangedWeapon={rangedWeapon}
						dualWielding={dualWielding}
						pending={pending}
					/>
				);
			case 'miss':
				return miss && <MissRow key={row.id} info={miss} pending={pending} />;
			case 'avoidance':
				return (
					avoidance && (
						<AvoidanceRow
							key={row.id}
							info={avoidance}
							withHolyShield={player.isSpec(Spec.SpecProtectionPaladin)}
							hasParry={hasParry}
							hasBlock={hasBlock}
							pending={pending}
						/>
					)
				);
			case 'crit-immunity':
				return critImmunity && <CritImmunityRow key={row.id} info={critImmunity.info} text={critImmunity.text} pending={pending} />;
			case 'melee-crit-cap':
				return critCap && <CritCapRow key={row.id} info={critCap.info} text={critCap.text} pending={pending} />;
		}
	};

	const table = (rows: Row[]) => (
		<Table variant="plain" flush aria-busy={pending || undefined}>
			<tbody>{rows.map(renderRow)}</tbody>
		</Table>
	);

	const primary = groups.find(group => group.key === 'primary');
	const collapsible = groups.filter(group => group !== primary);

	return (
		<CharacterStatsPanelView
			className={className}
			title={i18n.t('sidebar.character_stats.title')}
			filter={<StatFilterPopover {...statFilter} ariaLabel={i18n.t('stat_filters.triggers.character_stats')} testId="character-stats-filter" />}>
			{primary && (
				<div className="border-b border-border pb-1" data-testid="character-stats-primary">
					{table(primary.rows)}
				</div>
			)}
			<Accordion value={openGroups} onValueChange={setOpenGroups} variant="divided" density="compact" testId="character-stats-groups">
				{collapsible.map(group => (
					<AccordionItem
						key={group.key}
						value={group.key}
						title={i18n.t(`sidebar.character_stats.groups.${group.key}`)}
						panelClassName="pb-1"
						testId={`character-stats-group-${group.key}`}>
						{table(group.rows)}
					</AccordionItem>
				))}
				<WeaponSkills />
			</Accordion>
		</CharacterStatsPanelView>
	);
};

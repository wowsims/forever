import type { PlayerStats } from '@generated/proto/api';
import type { AreaType, Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateAreaType, translateMobType, translateStat } from '@i18n/localization';
import { useSimHost } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { UnitStat } from '@sim/proto/stats';
import { AccordionItem } from '@ui-kit/Accordion';
import { Table, TableCell, TableRow } from '@ui-kit/Table';
import { useStore } from 'zustand';

export type VersusBonuses = Pick<PlayerStats, 'mobTypeStats' | 'mobTypeDamage' | 'areaStats'>;

export interface VersusBonusesViewProps {
	bonuses: VersusBonuses;
	// The encounter's area types, whose stats the other groups already count.
	areaTypes: ReadonlyArray<AreaType>;
}

interface VersusLine {
	label: string;
	value: string;
}

interface VersusGroup {
	key: string;
	label: string;
	lines: Array<VersusLine>;
	counted: boolean;
}

const signed = (value: number, suffix = '') => `${value > 0 ? '+' : ''}${Number(value.toFixed(2))}${suffix}`;

// A rating the game shows as a percent reads as that percent, as everywhere on the sheet.
const statLines = (stats: { [key: number]: number }): Array<VersusLine> =>
	Object.entries(stats).map(([key, value]) => {
		const stat = Number(key) as Stat;
		const unitStat = UnitStat.fromStat(stat);
		return { label: translateStat(stat), value: unitStat.isPercentRating() ? signed(unitStat.convertRatingToPercent(value)!, '%') : signed(value) };
	});

const multiplierLines = (key: string, multiplier: number): Array<VersusLine> =>
	Math.abs(multiplier - 1) < 1e-9 ? [] : [{ label: i18n.t(`sidebar.character_stats.versus.${key}`), value: signed((multiplier - 1) * 100, '%') }];

const versusGroups = ({ bonuses, areaTypes }: VersusBonusesViewProps): Array<VersusGroup> => {
	const mobTypes = [...new Set([...bonuses.mobTypeStats.flatMap(entry => entry.mobTypes), ...bonuses.mobTypeDamage.map(entry => entry.mobType)])];
	const mobTypeGroups = mobTypes
		.sort((a, b) => a - b)
		.map(mobType => {
			const damage = bonuses.mobTypeDamage.find(entry => entry.mobType === mobType);
			return {
				key: `mob-type-${mobType}`,
				label: translateMobType(mobType),
				lines: [
					...bonuses.mobTypeStats.filter(entry => entry.mobTypes.includes(mobType)).flatMap(entry => statLines(entry.stats)),
					...(damage ? [...multiplierLines('damage', damage.damageMultiplier), ...multiplierLines('crit_damage', damage.critDamageMultiplier)] : []),
				],
				counted: false,
			};
		});
	const areaGroups = bonuses.areaStats.map(entry => ({
		key: `area-type-${entry.areaType}`,
		label: i18n.t('sidebar.character_stats.versus.in_area', { area: translateAreaType(entry.areaType) }),
		lines: statLines(entry.stats),
		counted: areaTypes.includes(entry.areaType),
	}));
	return [...mobTypeGroups, ...areaGroups];
};

export const VersusBonusesView = (props: VersusBonusesViewProps) => (
	<Table variant="plain" flush>
		{versusGroups(props).map(group => (
			<tbody key={group.key} className="data-current:text-success" data-current={group.counted || undefined} data-testid="character-stats-versus-group">
				<TableRow className="ui-character-stats-row">
					<TableCell className="ui-character-stats-label font-bold" colSpan={2}>
						{group.label}
					</TableCell>
				</TableRow>
				{group.counted && (
					<TableRow className="ui-character-stats-row">
						<TableCell className="ui-character-stats-label text-xs text-muted" colSpan={2}>
							{i18n.t('sidebar.character_stats.versus.already_counted')}
						</TableCell>
					</TableRow>
				)}
				{group.lines.map(line => (
					<TableRow key={line.label} className="ui-character-stats-row">
						<TableCell className="ui-character-stats-label pl-2">{line.label}</TableCell>
						<TableCell className="ui-character-stats-value">{line.value}</TableCell>
					</TableRow>
				))}
			</tbody>
		))}
	</Table>
);

export const VersusBonuses = () => {
	const host = useSimHost();
	const bonuses = usePlayerStore('currentStats');
	const areaTypes = useStore(host.sim.store, s => s.encounter.areaTypes);
	if (!bonuses.mobTypeStats.length && !bonuses.mobTypeDamage.length && !bonuses.areaStats.length) return null;

	return (
		<AccordionItem value="versus" title={i18n.t('sidebar.character_stats.versus.title')} panelClassName="pb-1" testId="character-stats-group-versus">
			<VersusBonusesView bonuses={bonuses} areaTypes={areaTypes} />
		</AccordionItem>
	);
};

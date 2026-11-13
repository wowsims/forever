import { type PlayerStats, type VersusEffect, VersusEffectKind } from '@generated/proto/api';
import type { AreaType, Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateAreaType, translateCreatureFamily, translateMobType, translateStat } from '@i18n/localization';
import { useSimHost } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { UnitStat } from '@sim/proto/stats';
import { AccordionItem } from '@ui-kit/Accordion';
import { Table, TableCell, TableHeaderCell, TableRow } from '@ui-kit/Table';
import { useStore } from 'zustand';

export type VersusBonusSet = Pick<PlayerStats, 'mobTypeBonuses' | 'areaStats' | 'versusEffects'>;

export interface VersusBonusesViewProps {
	bonuses: VersusBonusSet;
	// The encounter's area types, whose stats the other groups already count.
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
	lines: Array<VersusLine>;
	counted: boolean;
}

const signed = (value: number, suffix = '') => `${value > 0 ? '+' : ''}${Number(value.toFixed(2))}${suffix}`;

// A rating the game shows as a percent reads as that percent, as everywhere on the sheet.
const statLines = (stats: { [key: number]: number }): Array<VersusLine> =>
	Object.entries(stats).map(([key, value]) => {
		const stat = Number(key) as Stat;
		const unitStat = UnitStat.fromStat(stat);
		return {
			key: `stat-${stat}`,
			label: translateStat(stat),
			value: unitStat.isPercentRating() ? signed(unitStat.convertRatingToPercent(value)!, '%') : signed(value),
		};
	});

const multiplierLines = (key: string, multiplier: number): Array<VersusLine> =>
	Math.abs(multiplier - 1) < 1e-9 ? [] : [{ key, label: i18n.t(`sidebar.character_stats.versus.${key}`), value: signed((multiplier - 1) * 100, '%') }];

const effectLines = (effects: Array<VersusEffect>): Array<VersusLine> =>
	effects.map(effect => ({
		key: `item-${effect.itemId}`,
		label: effect.itemName,
		value: i18n.t(`sidebar.character_stats.versus.${effect.kind === VersusEffectKind.VersusEffectKindRestore ? 'effect_restore' : 'effect_damage'}`, {
			factor: Number(effect.multiplier.toFixed(2)),
		}),
	}));

const versusGroups = ({ bonuses, areaTypes }: VersusBonusesViewProps): Array<VersusGroup> => {
	const mobTypes = [...new Set([...bonuses.mobTypeBonuses.map(entry => entry.mobType), ...bonuses.versusEffects.flatMap(effect => effect.mobTypes)])];
	const mobTypeGroups = mobTypes
		.sort((a, b) => a - b)
		.map(mobType => {
			const bonus = bonuses.mobTypeBonuses.find(entry => entry.mobType === mobType);
			return {
				key: `mob-type-${mobType}`,
				label: translateMobType(mobType),
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
				counted: false,
			};
		});
	const familyGroups = [...new Set(bonuses.versusEffects.flatMap(effect => effect.creatureFamilies))]
		.sort((a, b) => a - b)
		.map(family => ({
			key: `creature-family-${family}`,
			label: translateCreatureFamily(family),
			lines: effectLines(bonuses.versusEffects.filter(effect => effect.creatureFamilies.includes(family))),
			counted: false,
		}));
	const areaGroups = bonuses.areaStats.map(entry => ({
		key: `area-type-${entry.areaType}`,
		label: i18n.t('sidebar.character_stats.versus.in_area', { area: translateAreaType(entry.areaType) }),
		lines: statLines(entry.stats),
		counted: areaTypes.includes(entry.areaType),
	}));
	return [...mobTypeGroups, ...familyGroups, ...areaGroups];
};

export const VersusBonusesView = (props: VersusBonusesViewProps) => (
	<Table variant="plain" flush>
		{versusGroups(props).map(group => (
			<tbody key={group.key} className="data-current:text-success" data-current={group.counted || undefined} data-testid="character-stats-versus-group">
				<TableRow className="ui-character-stats-row">
					<TableHeaderCell scope="rowgroup" colSpan={2} wrap className="ui-character-stats-label text-left font-bold">
						{group.label}
					</TableHeaderCell>
				</TableRow>
				{group.counted && (
					<TableRow className="ui-character-stats-row">
						<TableCell className="ui-character-stats-label text-xs text-muted" colSpan={2}>
							{i18n.t('sidebar.character_stats.versus.already_counted')}
						</TableCell>
					</TableRow>
				)}
				{group.lines.map(line => (
					<TableRow key={line.key} className="ui-character-stats-row">
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
	if (!bonuses.mobTypeBonuses.length && !bonuses.areaStats.length && !bonuses.versusEffects.length) return null;

	return (
		<AccordionItem value="versus" title={i18n.t('sidebar.character_stats.versus.title')} panelClassName="pb-1" testId="character-stats-group-versus">
			<VersusBonusesView bonuses={bonuses} areaTypes={areaTypes} />
		</AccordionItem>
	);
};

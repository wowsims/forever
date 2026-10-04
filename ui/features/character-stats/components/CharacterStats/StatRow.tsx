import { Class, PseudoStat, RangedWeaponType } from '@generated/proto/common';
import type { UIItem } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import type { EquippedItem } from '@sim/proto/equipped_item';
import type { StatAttribution, Stats, UnitStat } from '@sim/proto/stats';
import { Skeleton } from '@ui-kit/Skeleton';
import { TableCell, TableRow } from '@ui-kit/Table';
import { Tooltip } from '@ui-kit/Tooltip';
import { TooltipAnchor } from '@ui-kit/TooltipAnchor';
import { type ReactNode, useId } from 'react';

import { BonusStatsLink } from './BonusStatsLink';
import { type ShowStat, StatTooltipBody } from './StatTooltipBody';
import type { DisplayStat } from './utils/rows';
import { bonusStatClass, bonusTone } from './utils/stat_display';

/** The weapons whose speed the effective-weapon-speed tooltip row is meaningful for. */
const HUNTER_RANGED_TYPES = [RangedWeaponType.RangedWeaponTypeBow, RangedWeaponType.RangedWeaponTypeCrossbow, RangedWeaponType.RangedWeaponTypeGun];

export interface StatRowViewProps {
	displayStat: DisplayStat;
	playerClass: Class;
	bonusStats: Stats;
	attribution: StatAttribution;
	show: ShowStat;
	rangedWeapon: UIItem | null;
	dualWielding: boolean;
	pending?: boolean;
	/** The bonus-stat picker, left out on a row whose stat takes no bonus. */
	bonusStatsLink?: ReactNode;
}

export const StatRowView = ({
	displayStat,
	playerClass,
	bonusStats,
	attribution,
	show,
	rangedWeapon,
	dualWielding,
	pending,
	bonusStatsLink,
}: StatRowViewProps) => {
	const mainHandId = useId();
	const offHandId = useId();
	const { stat: unitStat, offHandStat } = displayStat;
	const bonusStatValue = unitStat.hasRootStat()
		? bonusStats.getStat(unitStat.getRootStat())
		: unitStat.isPseudoStat()
			? bonusStats.getPseudoStat(unitStat.getPseudoStat())
			: 0;
	const contextualClass = bonusStatClass(bonusStatValue);
	const tone = bonusTone(bonusStatValue);

	const showEffectiveWeaponSpeed =
		playerClass === Class.ClassHunter &&
		unitStat.equalsPseudoStat(PseudoStat.PseudoStatRangedHastePercent) &&
		!!rangedWeapon &&
		HUNTER_RANGED_TYPES.includes(rangedWeapon.rangedWeaponType);

	// A row with an off-hand half shows each hand on its own line, prefixed MH and OH, while dual
	// wielding with hands that differ, and the main-hand value alone otherwise.
	// Each hand's final display string, computed once and shared by its anchor and its tooltip's Total row.
	const mainHandTotal = show(attribution.final, unitStat, true);
	const offHandTotal = dualWielding && offHandStat ? show(attribution.final, offHandStat, true) : mainHandTotal;
	const hands: { stat: UnitStat; id: string; total: string; label?: string; prefix?: string }[] =
		offHandStat && offHandTotal !== mainHandTotal
			? [
					{
						stat: unitStat,
						id: mainHandId,
						total: mainHandTotal,
						label: i18n.t('slots.main_hand', { ns: 'character' }),
						prefix: i18n.t('sidebar.character_stats.hands.main'),
					},
					{
						stat: offHandStat,
						id: offHandId,
						total: offHandTotal,
						label: i18n.t('slots.off_hand', { ns: 'character' }),
						prefix: i18n.t('sidebar.character_stats.hands.off'),
					},
				]
			: [{ stat: unitStat, id: mainHandId, total: mainHandTotal }];

	return (
		<TableRow testId="character-stats-table-row" className="ui-character-stats-row">
			<TableCell className="ui-character-stats-label">{unitStat.getShortName(playerClass)}</TableCell>
			<TableCell className="ui-character-stats-value">
				{pending ? (
					<Skeleton />
				) : (
					<>
						<div className="ui-stat-value-link-container" data-tone={tone}>
							{hands.map(hand => (
								<span key={hand.stat.getKey()}>
									{hand.prefix && <span className="font-normal">{hand.prefix} </span>}
									<TooltipAnchor tooltipId={hand.id} testId="stat-value-link" className={contextualClass}>
										{hand.total}
									</TooltipAnchor>
								</span>
							))}
						</div>
						{bonusStatsLink}
						{hands.map(hand => (
							<Tooltip
								key={hand.stat.getKey()}
								id={hand.id}
								content={
									<StatTooltipBody
										hand={hand.stat}
										handLabel={hand.label}
										attribution={attribution}
										show={show}
										bonusStats={bonusStats}
										bonusStatValue={bonusStatValue}
										total={hand.total}
										rangedWeaponSpeed={showEffectiveWeaponSpeed ? rangedWeapon!.weaponSpeed : undefined}
									/>
								}
							/>
						))}
					</>
				)}
			</TableCell>
		</TableRow>
	);
};

export interface StatRowProps extends Omit<StatRowViewProps, 'playerClass' | 'rangedWeapon' | 'bonusStatsLink'> {
	rangedWeapon: EquippedItem | null;
}

export const StatRow = ({ rangedWeapon, ...props }: StatRowProps) => {
	const player = usePlayer();
	const { stat, notEditable } = props.displayStat;
	return (
		<StatRowView
			{...props}
			playerClass={player.getClass()}
			rangedWeapon={rangedWeapon?.item ?? null}
			bonusStatsLink={notEditable ? undefined : <BonusStatsLink unitStat={stat} />}
		/>
	);
};

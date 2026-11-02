import { ConnectedNumberField } from '@features/inputs/components/ConnectedTextField';
import type { NumberFieldConfig } from '@features/inputs/configs';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import type { UnitStat } from '@sim/proto/stats';
import { useMemo } from 'react';

export interface BonusStatsPickerProps {
	unitStat: UnitStat;
	label: string;
	onCommit: () => void;
}

/**
 * Bonus stats include pseudo stats — the speed multipliers have no root stat at all — so
 * the picker keys on the UnitStat rather than on a Stat.
 */
export const BonusStatsPicker = ({ unitStat, label, onCommit }: BonusStatsPickerProps) => {
	const player = usePlayer();
	const rootStat = unitStat.hasRootStat() ? unitStat.getRootStat() : undefined;
	// A rating stat is stored as the rating its percent costs; a school hit row is stored as the percent itself.
	const ratingStoredAsPercent = unitStat.isPercentRating();
	const percentInput = ratingStoredAsPercent || unitStat.isRatingBackedPercent();

	const config = useMemo(
		(): NumberFieldConfig<Player<any>> => ({
			id: `character-bonus-stat-${unitStat.getKey()}`,
			label: percentInput ? `${label} (${i18n.t('sidebar.character_stats.percent_suffix')})` : label,
			inputClassName: 'm-0 w-32 flex-1',
			float: percentInput,
			maxDecimalDigits: percentInput ? 3 : undefined,
			storeField: 'bonusStats',
			getValue: subject => {
				const bonusStats = subject.getBonusStats();
				const value = rootStat !== undefined ? bonusStats.getStat(rootStat) : bonusStats.getUnitStat(unitStat);
				return ratingStoredAsPercent ? unitStat.convertRatingToPercent(value)! : value;
			},
			setValue: (subject, newValue) => {
				const bonusStats = subject.getBonusStats();
				const value = ratingStoredAsPercent ? unitStat.convertPercentToRating(newValue)! : newValue;
				subject.setBonusStats(rootStat !== undefined ? bonusStats.withStat(rootStat, value) : bonusStats.withUnitStat(unitStat, value));
				onCommit();
			},
		}),
		[unitStat, rootStat, label, percentInput, ratingStoredAsPercent, onCommit],
	);

	return <ConnectedNumberField modObject={player} config={config} />;
};

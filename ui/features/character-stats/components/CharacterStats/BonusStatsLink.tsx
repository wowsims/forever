import i18n from '@i18n/config';
import { translatePseudoStat, translateStat } from '@i18n/localization';
import type { UnitStat } from '@sim/proto/stats';

import { BonusLink } from './BonusLink';
import { BonusStatsPicker } from './BonusStatsPicker';

export interface BonusStatsLinkProps {
	unitStat: UnitStat;
}

/** The name the bonus-stat picker is labelled with: the root stat where there is one, the pseudo stat otherwise. */
const bonusStatName = (unitStat: UnitStat): string =>
	unitStat.hasRootStat()
		? translateStat(unitStat.getRootStat())
		: unitStat.isPseudoStat()
			? translatePseudoStat(unitStat.getPseudoStat())
			: translateStat(unitStat.getStat());

export const BonusStatsLink = ({ unitStat }: BonusStatsLinkProps) => {
	const label = `${i18n.t('sidebar.character_stats.bonus_prefix')} ${bonusStatName(unitStat)}`;
	return <BonusLink label={label} picker={onCommit => <BonusStatsPicker unitStat={unitStat} label={label} onCommit={onCommit} />} />;
};

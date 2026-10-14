import i18n from '@i18n/config';
import type { DamageLog } from '@sim/proto/combat_log';
import { isAvoidedOutcome } from '@sim/proto/combat_log/types';
import { spellSchoolNames } from '@sim/proto/names';
import { SPELL_SCHOOL_TEXT } from '@ui-kit/utils/colors';
import { Trans } from 'react-i18next';

import { EntityLabel } from '../EntityLabel';
import { outcomeLabel } from './utils';

export interface DamageResultProps {
	log: DamageLog;
}

/** The outcome-and-amount tail of a damage, healing or shielding line. */
export const DamageResult = ({ log }: DamageResultProps) => {
	const spellSchoolString = typeof log.spellSchool === 'number' ? spellSchoolNames.get(log.spellSchool) : undefined;
	const schoolClass = spellSchoolString ? SPELL_SCHOOL_TEXT[spellSchoolString.toLowerCase()] : undefined;
	const isHealing = log.effect === 'healing';
	const isShielding = log.effect === 'shielding';
	const isMissLike = isAvoidedOutcome(log.outcome);
	const outcome = log.outcome === 'hit' && log.tick ? i18n.t('results_tab.details.attack_types.tick') : outcomeLabel(log.outcome);
	const amount = log.amount.toFixed(2);
	const resist = <>{log.resist > 0 && <> {i18n.t('results_tab.details.damage_result.resist', { percent: log.resist })}</>}</>;

	return (
		<>
			{isHealing ? `${i18n.t('results_tab.details.damage_result.healed')} ` : ''}
			{isShielding ? `${i18n.t('results_tab.details.damage_result.shielded')} ` : ''}
			{!(isHealing || isShielding) && <>{outcome}</>}
			{` `}
			{log.target ? <EntityLabel entity={log.target} /> : ''}
			{!isMissLike ? (
				<>
					{' '}
					{isHealing || isShielding ? (
						<Trans
							i18n={i18n}
							i18nKey="results_tab.details.damage_result.health_amount"
							values={{ amount }}
							components={{ amount: <strong className="text-resource-health" />, resist }}
						/>
					) : (
						<Trans
							i18n={i18n}
							i18nKey="results_tab.details.damage_result.damage_amount"
							values={{ amount }}
							components={{
								amount: <strong className={schoolClass ?? 'text-danger'} />,
								school: <>{spellSchoolString && <> ({spellSchoolString})</>}</>,
								resist,
							}}
						/>
					)}
				</>
			) : (
				''
			)}
		</>
	);
};

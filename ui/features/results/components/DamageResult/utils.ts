import i18n from '@i18n/config';
import type { Outcome } from '@sim/proto/combat_log';

export const outcomeLabel = (outcome: Outcome): string => {
	switch (outcome) {
		case 'miss':
			return i18n.t('results_tab.details.attack_types.miss');
		case 'dodge':
			return i18n.t('results_tab.details.attack_types.dodge');
		case 'parry':
			return i18n.t('results_tab.details.attack_types.parry');
		case 'critical-block':
			return i18n.t('results_tab.details.attack_types.critical_block');
		case 'block':
			return i18n.t('results_tab.details.attack_types.block');
		case 'glance':
			return i18n.t('results_tab.details.attack_types.glance');
		case 'crit':
			return i18n.t('results_tab.details.attack_types.crit');
		case 'crush':
			return i18n.t('results_tab.details.attack_types.crush');
		case 'hit':
			return i18n.t('results_tab.details.attack_types.hit');
	}
};

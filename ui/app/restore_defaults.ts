import i18n from '@i18n/config';
import { toastManager } from '@ui-kit/Toast';

import { trackEvent } from '../tracking/utils';
import type { SimHostObject } from './individual_sim_ui';

/** The Restore Defaults action: gear, consumes, buffs, talents and EP weights go back to the spec's defaults; saved settings stay. */
export const restoreDefaults = (host: SimHostObject<any>) => {
	trackEvent({ action: 'settings', category: 'restore-defaults', label: 'restore' });
	host.applyDefaults();
	toastManager.add({ variant: 'success', body: i18n.t('info.options.restore_defaults.success_message') });
};

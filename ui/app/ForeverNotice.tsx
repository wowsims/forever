import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Notice } from '@ui-kit/Notice';

import type { SimHostObject } from './individual_sim_ui';
import { restoreDefaults } from './restore_defaults';

export interface ForeverNoticeViewProps {
	/** Undefined disables the button, until the host exists. */
	onRestoreDefaults?: () => void;
}

export const ForeverNoticeView = ({ onRestoreDefaults }: ForeverNoticeViewProps) => (
	<div className="flex w-full flex-wrap items-center justify-center gap-x-4 gap-y-2 bg-overlay p-4 text-center" data-testid="forever-notice">
		<Notice level="warning">{i18n.t('info.forever_notice')}</Notice>
		<Button variant="outline-primary" size="sm" disabled={!onRestoreDefaults} onClick={onRestoreDefaults}>
			{i18n.t('info.options.restore_defaults.button')}
		</Button>
	</div>
);

export const ForeverNotice = ({ host }: { host: SimHostObject<any> | null }) => (
	<ForeverNoticeView onRestoreDefaults={host ? () => restoreDefaults(host) : undefined} />
);

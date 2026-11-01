import i18n from '@i18n/config';
import { Icon, type IconName } from '@ui-kit/Icon';

export type BulkSetupKind = 'gear' | 'talents' | 'rotation' | 'settings';

const ICONS: Record<BulkSetupKind, IconName> = { gear: 'shirt', talents: 'book', rotation: 'list-check', settings: 'gear' };

export interface BulkSetupLabelProps {
	kind: BulkSetupKind;
	/** The setup's name, or "Current ..." for the running one. */
	label: string;
}

// The kind shows as an icon; screen readers get it as hidden text in front of the name.
export const BulkSetupLabel = ({ kind, label }: BulkSetupLabelProps) => (
	<>
		<Icon name={ICONS[kind]} fixedWidth />
		<span className="sr-only">{i18n.t(`bulk_tab.results.${kind}`)}</span>
		{label}
	</>
);

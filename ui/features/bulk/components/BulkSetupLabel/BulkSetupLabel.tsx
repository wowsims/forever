import i18n from '@i18n/config';
import { Icon, type IconName } from '@ui-kit/Icon';

export type BulkSetupKind = 'gear' | 'talents' | 'rotation' | 'settings';

const ICONS: Record<BulkSetupKind, IconName> = { gear: 'shirt', talents: 'book', rotation: 'list-check', settings: 'gear' };

export const BULK_SETUP_KINDS: readonly BulkSetupKind[] = ['gear', 'talents', 'rotation', 'settings'];

/** A build's four parts by kind, each its name or "Current ..." for the running one. */
export type BulkSetupLabels = Readonly<Record<BulkSetupKind, string>>;

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

// One line per part, for a named build's tooltip.
export const BulkSetupLines = ({ labels }: { labels: BulkSetupLabels }) =>
	BULK_SETUP_KINDS.map(kind => (
		<span key={kind} className="flex items-center gap-1">
			<BulkSetupLabel kind={kind} label={labels[kind]} />
		</span>
	));

/** The parts as one line of text, which describes a named build for screen readers. */
export const bulkSetupText = (labels: BulkSetupLabels): string =>
	BULK_SETUP_KINDS.map(kind => `${i18n.t(`bulk_tab.results.${kind}`)} ${labels[kind]}`).join(', ');

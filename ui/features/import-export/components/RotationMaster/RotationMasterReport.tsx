import i18n from '@i18n/config';
import { Accordion, AccordionItem } from '@ui-kit/Accordion';
import { useState } from 'react';

import type { Fidelity, ReportRow } from '../../rotation_master';

// The labels are spelled out so the locale test finds every key.
const FIDELITY: Record<Fidelity, { className: string; label: () => string }> = {
	exact: { className: 'text-success', label: () => i18n.t('import.rotation_master.fidelity.exact') },
	approximate: { className: 'text-warning', label: () => i18n.t('import.rotation_master.fidelity.approximate') },
	lossy: { className: 'text-danger', label: () => i18n.t('import.rotation_master.fidelity.lossy') },
	unsupported: { className: 'text-danger', label: () => i18n.t('import.rotation_master.fidelity.unsupported') },
};

const Row = ({ row }: { row: ReportRow }) => (
	<li className="py-1" data-testid="rotation-master-row" data-fidelity={row.fidelity}>
		<div>
			<span className={FIDELITY[row.fidelity].className}>{FIDELITY[row.fidelity].label()}</span> <span>{row.title}</span>{' '}
			<span className="text-muted">
				{row.path} - {row.outcome}
			</span>
		</div>
		{row.detail && <div className="break-all text-muted">{row.detail}</div>}
		{row.sentences.length > 0 && (
			<ul className="list-disc pl-6">
				{row.sentences.map(sentence => (
					<li key={sentence}>{sentence}</li>
				))}
			</ul>
		)}
	</li>
);

export interface RotationMasterReportProps {
	rows: ReadonlyArray<ReportRow>;
}

/** The items that need a look are listed; the exact ones are folded away. */
export const RotationMasterReport = ({ rows }: RotationMasterReportProps) => {
	const [open, setOpen] = useState<string[]>(['attention']);
	const attention = rows.filter(row => row.fidelity !== 'exact');
	const exact = rows.filter(row => row.fidelity === 'exact');
	return (
		<Accordion value={open} onValueChange={setOpen}>
			{attention.length > 0 && (
				<AccordionItem
					value="attention"
					title={i18n.t('import.rotation_master.report.attention', { count: attention.length })}
					testId="rotation-master-attention">
					<ul>
						{attention.map(row => (
							<Row key={row.key} row={row} />
						))}
					</ul>
				</AccordionItem>
			)}
			{exact.length > 0 && (
				<AccordionItem value="exact" title={i18n.t('import.rotation_master.report.exact', { count: exact.length })} testId="rotation-master-exact">
					<ul>
						{exact.map(row => (
							<Row key={row.key} row={row} />
						))}
					</ul>
				</AccordionItem>
			)}
		</Accordion>
	);
};

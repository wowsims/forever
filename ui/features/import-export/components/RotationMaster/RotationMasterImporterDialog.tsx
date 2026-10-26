import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { Button } from '@ui-kit/Button';
import { Dialog } from '@ui-kit/Dialog';
import { TextArea } from '@ui-kit/FormControl';
import { Icon } from '@ui-kit/Icon';
import { toastManager } from '@ui-kit/Toast';
import { useMemo, useRef, useState } from 'react';

import { trackPageView } from '../../../../tracking/utils';
import { applyRotationMasterImport, checkRotationMasterImport } from '../../importers/rotation_master';
import { describeError, describeWarning, importRows, summaryLine } from '../../rotation_master';
import type { ImporterDialogProps } from '../Importer';
import { RotationMasterReport } from './RotationMasterReport';

type Checked = ReturnType<typeof checkRotationMasterImport>;

/**
 * Paste a Rotation Master export, see what it becomes, then replace the APL with it. Nothing but the APL is touched
 * (the JSON importer, by contrast, resets every setting the file does not carry).
 */
export const RotationMasterImporterDialog = ({ open, onOpenChange }: ImporterDialogProps) => {
	const host = useSimHost();
	const textRef = useRef<HTMLTextAreaElement>(null);
	const [checked, setChecked] = useState<Checked | null>(null);

	const rows = useMemo(() => (checked?.ok || checked?.items ? importRows(checked.items ?? []) : []), [checked]);

	const check = (hideLossy: boolean): Checked => {
		const result = checkRotationMasterImport(host, textRef.current?.value ?? '', { hideLossy });
		setChecked(result);
		return result;
	};
	const replace = (hideLossy: boolean) => {
		const result = check(hideLossy);
		if (!result.ok) return;
		try {
			applyRotationMasterImport(host, result.apl);
		} catch (error: any) {
			toastManager.add({ variant: 'error', body: `Import error: ${error?.message || error}` });
			return;
		}
		toastManager.add({
			variant: 'success',
			body: i18n.t('import.rotation_master.success', { actions: result.summary.actions, prepull: result.summary.prepull }),
		});
		onOpenChange(false);
	};

	return (
		<Dialog
			open={open}
			onOpenChange={next => {
				if (next) trackPageView(i18n.t('import.rotation_master.title'), '/import/rotation-master');
				onOpenChange(next);
			}}
			testId="importer"
			title={i18n.t('import.rotation_master.title')}
			scrollContents
			footer={
				<div className="flex gap-2">
					<Button variant="secondary" data-testid="check-button" className="w-48" onClick={() => check(false)}>
						<Icon name="check" style="base" className="mr-1" />
						{i18n.t('import.rotation_master.check_button')}
					</Button>
					<Button data-testid="import-button" className="w-48" onClick={() => replace(false)}>
						<Icon name="download" style="base" className="mr-1" />
						{i18n.t('import.rotation_master.replace_button')}
					</Button>
					<Button variant="secondary" data-testid="import-hidden-button" onClick={() => replace(true)}>
						{i18n.t('import.rotation_master.replace_hidden_button')}
					</Button>
				</div>
			}>
			<div>
				<p>
					{i18n.t('import.rotation_master.description')}{' '}
					<Button href="https://rotationmaster.neuromancy.net" target="_blank" variant="unstyled">
						{i18n.t('import.rotation_master.site_link')}
					</Button>
				</p>
				<p>{i18n.t('import.rotation_master.feature_description')}</p>
				<p>{i18n.t('import.rotation_master.instructions')}</p>
			</div>
			<TextArea
				spellCheck={false}
				data-testid="importer-textarea"
				className="h-[20vh] w-full resize-none"
				ref={textRef}
				onChange={() => setChecked(null)}
			/>
			{checked && !checked.ok && (
				<p className="mt-2 text-danger" data-testid="rotation-master-error">
					{describeError(checked.error)}
				</p>
			)}
			{checked?.ok && (
				<div className="mt-2" data-testid="rotation-master-result">
					<p>
						{i18n.t('import.rotation_master.result', {
							actions: checked.summary.actions,
							prepull: checked.summary.prepull,
							summary: summaryLine(checked.summary),
						})}
					</p>
					{checked.warnings.map(warning => (
						<p key={warning.code} className="text-warning">
							{describeWarning(warning)}
						</p>
					))}
				</div>
			)}
			{rows.length > 0 && <RotationMasterReport rows={rows} />}
		</Dialog>
	);
};

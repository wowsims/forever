import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { Button } from '@ui-kit/Button';
import { CopyButton } from '@ui-kit/CopyButton';
import { Dialog } from '@ui-kit/Dialog';
import { TextArea } from '@ui-kit/FormControl';
import { Icon } from '@ui-kit/Icon';
import { downloadString } from '@ui-kit/utils/dom';
import { useEffect, useMemo, useRef } from 'react';

import { trackPageView } from '../../../../tracking/utils';
import { exportRotationMasterText } from '../../importers/rotation_master';
import { describeError, describeWarning, exportRows, summaryLine } from '../../rotation_master';
import type { ExporterDialogProps } from '../Exporter';
import { RotationMasterReport } from './RotationMasterReport';

/** The player's rotation as a Rotation Master export, with what was lost on the way. */
export const RotationMasterExporterDialog = ({ open, onOpenChange }: ExporterDialogProps) => {
	const host = useSimHost();
	const textRef = useRef<HTMLTextAreaElement>(null);
	const made = useMemo(() => (open ? exportRotationMasterText(host) : null), [open, host]);
	const text = made?.result.ok ? made.result.text : '';
	const textValue = useRef(text);
	textValue.current = text;

	useEffect(() => {
		if (textRef.current) textRef.current.value = text;
	}, [text]);
	useEffect(() => {
		if (open) trackPageView(i18n.t('export.rotation_master.title'), '/export/rotation-master');
	}, [open]);

	const rows = useMemo(() => (made?.result.items ? exportRows(made.result.items) : []), [made]);

	return (
		<Dialog
			open={open}
			onOpenChange={onOpenChange}
			testId="exporter"
			title={i18n.t('export.rotation_master.title')}
			scrollContents
			footer={
				<>
					<CopyButton
						getContent={() => textValue.current}
						variant="primary"
						className="w-48"
						text={i18n.t('export.json.copy_button')}
						tooltip={i18n.t('export.json.copy_tooltip')}
					/>
					<Button className="ml-2 w-48" onClick={() => downloadString(textValue.current, 'rotation-master-export.txt', 'text/plain')}>
						<Icon name="download" style="base" />
						{i18n.t('export.json.download_button')}
					</Button>
				</>
			}>
			<p>
				{i18n.t('export.rotation_master.description')}{' '}
				<Button href="https://rotationmaster.neuromancy.net" target="_blank" variant="unstyled">
					{i18n.t('import.rotation_master.site_link')}
				</Button>
			</p>
			<p className="text-warning" data-testid="rotation-master-warning">
				<strong>{i18n.t('export.rotation_master.warning_title')}</strong> {i18n.t('export.rotation_master.warning_message')}
			</p>
			{made?.generatedFrom && <p className="text-warning">{i18n.t('export.rotation_master.generated', { type: made.generatedFrom })}</p>}
			{made && !made.result.ok && (
				<p className="text-danger" data-testid="rotation-master-error">
					{describeError(made.result.error)}
				</p>
			)}
			{made?.result.ok && (
				<>
					<p data-testid="rotation-master-result">
						{i18n.t('export.rotation_master.result', {
							steps: made.result.summary.steps,
							disabled: made.result.summary.disabled,
							summary: summaryLine(made.result.summary),
						})}
					</p>
					{made.result.warnings.map(warning => (
						<p key={warning.code} className="text-warning">
							{describeWarning(warning)}
						</p>
					))}
				</>
			)}
			<TextArea spellCheck={false} className="h-[20vh] w-full resize-none" data-testid="exporter-textarea" ref={textRef} defaultValue={text} />
			{rows.length > 0 && <RotationMasterReport rows={rows} />}
		</Dialog>
	);
};

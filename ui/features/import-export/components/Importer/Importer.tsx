import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { kebabCase } from '@sim/utils/format';
import { Button } from '@ui-kit/Button';
import { Dialog } from '@ui-kit/Dialog';
import { TextArea } from '@ui-kit/FormControl';
import { Icon } from '@ui-kit/Icon';
import { toastManager } from '@ui-kit/Toast';
import type { ReactNode } from 'react';
import { useEffect, useRef } from 'react';

import { trackPageView } from '../../../../tracking/utils';
import type { ImporterDefinition } from '../../importers';

export interface ImporterProps extends ImporterDefinition {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	children?: ReactNode;
}

export const Importer = ({ open, onOpenChange, title, allowFileUpload = false, onImport, children }: ImporterProps) => {
	const host = useSimHost();
	const textRef = useRef<HTMLTextAreaElement>(null);
	const uploadInputId = `upload-input-${kebabCase(title)}`;

	useEffect(() => {
		if (!open) return;
		trackPageView(title, `/import/${kebabCase(title)}`);
	}, [open, title]);

	const runImport = async () => {
		try {
			await onImport(host, textRef.current?.value || '');
			onOpenChange(false);
		} catch (error: any) {
			toastManager.add({ variant: 'error', body: `${i18n.t('import.error')} ${error?.message || error}` });
		}
	};

	return (
		<Dialog
			open={open}
			onOpenChange={onOpenChange}
			testId="importer"
			title={title}
			footer={
				<div className="flex gap-2">
					{allowFileUpload && (
						<Button as="label" htmlFor={uploadInputId} data-testid="upload-button" className="w-48">
							<Icon name="file-arrow-up" />
							{i18n.t('import.json.upload_button')}
						</Button>
					)}
					<input
						type="file"
						id={uploadInputId}
						data-testid="importer-upload-input"
						className="hidden"
						hidden
						onChange={async event => {
							const file = event.target.files?.[0];
							if (!file || !textRef.current) return;
							textRef.current.value = await file.text();
						}}
					/>
					<Button data-testid="import-button" className="w-48" onClick={runImport}>
						<Icon name="download" style="base" />
						{i18n.t('import.json.import_button')}
					</Button>
				</div>
			}>
			{/* The dialog body is a flex column with a gap; this stays a single flex item so that gap doesn't land between the description and the textarea. */}
			<div className="flex flex-col">
				{children}
				<TextArea spellCheck={false} data-testid="importer-textarea" className="h-[40vh] w-full resize-none" ref={textRef} />
			</div>
		</Dialog>
	);
};

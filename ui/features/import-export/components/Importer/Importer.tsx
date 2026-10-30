import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { kebabCase } from '@sim/utils/format';
import { Button } from '@ui-kit/Button';
import { Dialog } from '@ui-kit/Dialog';
import { TextArea } from '@ui-kit/FormControl';
import { Icon } from '@ui-kit/Icon';
import { Notice } from '@ui-kit/Notice';
import { toastManager } from '@ui-kit/Toast';
import type { ReactNode } from 'react';
import { useEffect, useRef, useState } from 'react';

import { trackPageView } from '../../../../tracking/utils';
import { type ImporterDefinition, ImportProblems } from '../../importers/types';

export interface ImporterProps extends ImporterDefinition {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	children?: ReactNode;
}

export const Importer = ({ open, onOpenChange, title, allowFileUpload = false, onImport, children }: ImporterProps) => {
	const host = useSimHost();
	const textRef = useRef<HTMLTextAreaElement>(null);
	// The name of the file the text was uploaded from. It stays while the text is edited, and goes with the text:
	// cleared, or pasted over whole.
	const fileName = useRef<string>(undefined);
	const [problems, setProblems] = useState<ImportProblems>();
	// The import that has not ended, which the Import button waits on and closing the dialog drops.
	const running = useRef<AbortController>(undefined);
	const [pending, setPending] = useState(false);
	const uploadInputId = `upload-input-${kebabCase(title)}`;

	const dropRunning = () => {
		running.current?.abort();
		running.current = undefined;
		setPending(false);
	};

	// A closed dialog drops its text, and with it the file the text came from, what was wrong with it and the import
	// that has not ended.
	const changeOpen = (next: boolean) => {
		if (!next) {
			dropRunning();
			fileName.current = undefined;
			setProblems(undefined);
		}
		onOpenChange(next);
	};

	useEffect(() => {
		if (!open) return;
		trackPageView(title, `/import/${kebabCase(title)}`);
	}, [open, title]);

	useEffect(() => () => running.current?.abort(), []);

	const runImport = async () => {
		const run = new AbortController();
		running.current = run;
		setPending(true);
		setProblems(undefined);
		try {
			await onImport(host, textRef.current?.value || '', fileName.current, run.signal);
			if (!run.signal.aborted) changeOpen(false);
		} catch (error: any) {
			// A closed dialog cancelled the import, and dropped the text its problems point into.
			if (run.signal.aborted) return;
			if (error instanceof ImportProblems) setProblems(error);
			else toastManager.add({ variant: 'error', body: `${i18n.t('import.error')} ${error?.message || error}` });
		} finally {
			if (running.current === run) dropRunning();
		}
	};

	return (
		<Dialog
			open={open}
			onOpenChange={changeOpen}
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
							fileName.current = file.name;
						}}
					/>
					<Button data-testid="import-button" className="w-48" disabled={pending} loading={pending} onClick={runImport}>
						<Icon name="download" style="base" />
						{i18n.t('import.json.import_button')}
					</Button>
				</div>
			}>
			{/* The dialog body is a flex column with a gap; this stays a single flex item so that gap doesn't land between the description and the textarea. */}
			<div className="flex flex-col">
				{children}
				<TextArea
					spellCheck={false}
					data-testid="importer-textarea"
					className="h-[40vh] w-full resize-none"
					ref={textRef}
					onInput={event => {
						if (!event.currentTarget.value) fileName.current = undefined;
					}}
					onPaste={event => {
						const { selectionStart, selectionEnd, value } = event.currentTarget;
						if (selectionStart === 0 && selectionEnd === value.length) fileName.current = undefined;
					}}
				/>
				{problems && (
					<Notice level="error" className="mt-2" testId="importer-problems">
						<p>{problems.message}</p>
						<ul>
							{problems.problems.map((problem, i) => (
								<li key={i}>{problem}</li>
							))}
						</ul>
					</Notice>
				)}
			</div>
		</Dialog>
	);
};

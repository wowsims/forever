import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { kebabCase } from '@sim/utils/format';
import { Button } from '@ui-kit/Button';
import { CodeTextArea } from '@ui-kit/CodeTextArea';
import { CopyButton } from '@ui-kit/CopyButton';
import { Dialog } from '@ui-kit/Dialog';
import { TextArea } from '@ui-kit/FormControl';
import { Icon } from '@ui-kit/Icon';
import { toastManager } from '@ui-kit/Toast';
import { downloadString } from '@ui-kit/utils/dom';
import { useCallback, useEffect, useRef, useState } from 'react';

import { trackPageView } from '../../../../tracking/utils';
import { defaultExportCategories, type ExporterDefinition } from '../../exporters';
import { ExporterCategoryPickers } from './ExporterCategoryPickers';

export interface ExporterProps extends ExporterDefinition {
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

export const Exporter = ({
	open,
	onOpenChange,
	title,
	allowDownload = false,
	downloadFileName = 'wowsims.json',
	downloadMimeType,
	selectCategories = false,
	holdsLink = false,
	getData,
}: ExporterProps) => {
	const host = useSimHost();
	const categories = useRef(defaultExportCategories());
	const [categoryRevision, setCategoryRevision] = useState(0);
	const onCategoryChange = useCallback(() => setCategoryRevision(revision => revision + 1), []);

	const [data, setData] = useState('');
	useEffect(
		() => {
			const result = open ? getData(host, categories.current) : '';
			if (typeof result === 'string') {
				setData(result);
				return;
			}
			let live = true;
			setData('');
			result.then(
				text => {
					if (live) setData(text);
				},
				(error: unknown) => {
					if (!live) return;
					const message = error instanceof Error ? error.message : String(error);
					toastManager.add({ variant: 'error', body: i18n.t('export.error', { message }) });
				},
			);
			return () => {
				live = false;
			};
		},
		// `categories` is a ref, so the revision counter is what says its contents moved.
		// oxlint-disable-next-line react-hooks/exhaustive-deps
		[open, categoryRevision, host, getData],
	);
	const dataRef = useRef(data);
	dataRef.current = data;

	useEffect(() => {
		if (!open) return;
		trackPageView(title, `/export/${kebabCase(title)}`);
	}, [open, title]);

	return (
		<Dialog
			open={open}
			onOpenChange={onOpenChange}
			testId="exporter"
			title={title}
			footer={
				<>
					{/* An exporter whose text comes later has none until it arrives, and none at all when it fails. */}
					<CopyButton
						getContent={() => dataRef.current}
						disabled={!data}
						variant="primary"
						className="w-48"
						text={i18n.t('export.json.copy_button')}
						tooltip={i18n.t('export.json.copy_tooltip')}
					/>
					{allowDownload && (
						<Button className="ml-2 w-48" disabled={!data} onClick={() => downloadString(dataRef.current, downloadFileName, downloadMimeType)}>
							<Icon name="download" style="base" />
							{i18n.t('export.json.download_button')}
						</Button>
					)}
				</>
			}>
			{selectCategories && <ExporterCategoryPickers categories={categories.current} onChange={onCategoryChange} />}
			{holdsLink ? (
				<TextArea spellCheck={false} className="h-[60vh] w-full resize-none" data-testid="exporter-textarea" value={data} readOnly />
			) : (
				<CodeTextArea className="h-[60vh]" testId="exporter-textarea" value={data} readOnly />
			)}
		</Dialog>
	);
};
